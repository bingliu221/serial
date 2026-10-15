package modbus

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/bingliu221/serial/uart"
)

type ModbusRtuClientHandler struct {
	port           uart.Port
	broadcastDelay time.Duration
	sendMutex      sync.Mutex
}

func NewModbusRtuClientHandler(path string, cfg uart.Config, opts ...Option) (*ModbusRtuClientHandler, error) {
	o := options{
		broadcastDelay: 100 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(&o)
	}

	port, err := uart.Open(path, cfg)
	if err != nil {
		return nil, err
	}

	return &ModbusRtuClientHandler{
		port:           port,
		broadcastDelay: o.broadcastDelay,
	}, nil
}

func (h *ModbusRtuClientHandler) Close() error {
	return h.port.Close()
}

func (h *ModbusRtuClientHandler) readResp() (*message, error) {
	var buf [260]byte
	_, err := io.ReadFull(h.port, buf[:3])
	if err != nil {
		return nil, err
	}

	code := buf[1]
	failed := code&0x80 == 0x80

	remain := 2
	if !failed {
		switch code {
		case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
			remain += int(buf[2])
		case FnWriteCoil, FnWriteCoils, FnWriteRegister, FnWriteRegisters:
			remain += 3
		case FnMaskWriteRegister:
			remain += 5
		}
	}

	total := 3 + remain
	_, err = io.ReadFull(h.port, buf[3:total])
	if err != nil {
		return nil, err
	}

	adu := buf[:total]

	err = h.valify(adu)
	if err != nil {
		return nil, err
	}

	return &message{
		slaveId: adu[0],
		code:    adu[1],
		data:    adu[2 : total-2],
	}, nil
}

func (h *ModbusRtuClientHandler) frameEncode(msg *message) ([]byte, error) {
	data := msg.data
	if len(data) > 252 {
		return nil, errors.New("pdu length too large")
	}

	adu := make([]byte, len(data)+4)
	adu[0] = msg.slaveId
	adu[1] = msg.code
	copy(adu[2:], data)
	adu[2+len(data)], adu[3+len(data)] = crc(adu[:2+len(data)])

	return adu, nil
}

func (h *ModbusRtuClientHandler) valify(adu []byte) error {
	if len(adu) < 4 {
		return errors.New("invalid adu size")
	}

	crcIdx := len(adu) - 2

	low, high := crc(adu[:crcIdx])
	if adu[crcIdx] != low || adu[crcIdx+1] != high {
		return errors.New("crc not matched")
	}

	return nil
}

func (h *ModbusRtuClientHandler) Send(ctx context.Context, req *message) ([]byte, error) {
	h.sendMutex.Lock()
	defer h.sendMutex.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	adu, err := h.frameEncode(req)
	if err != nil {
		return nil, err
	}

	if err = h.port.Drain(); err != nil {
		return nil, err
	}

	_, err = h.port.Write(adu)
	if err != nil {
		return nil, err
	}

	if req.slaveId == 0 {
		// broadcast, no response
		if h.broadcastDelay > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(h.broadcastDelay):
			}
		}
		return nil, nil
	}

	deadline, _ := ctx.Deadline()
	if err = h.port.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	stop := context.AfterFunc(ctx, func() {
		_ = h.port.SetReadDeadline(time.Now())
	})
	defer func() {
		stop()
		_ = h.port.SetReadDeadline(time.Time{})
	}()

	resp, err := h.readResp()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}

	if resp.slaveId != req.slaveId || resp.code&0x7F != req.code {
		return nil, errors.New("mismatched response")
	}

	failed := resp.code&0x80 == 0x80
	if failed {
		if len(resp.data) == 0 {
			return nil, errors.New("invalid exception response")
		}
		exception := resp.data[0]
		return nil, ExceptionError(exception)
	}

	switch resp.code {
	case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
		if len(resp.data) < 1 {
			return nil, errors.New("invalid response data")
		}
		return resp.data[1:], nil
	case FnWriteCoil, FnWriteCoils, FnWriteRegister, FnWriteRegisters, FnMaskWriteRegister:
		return nil, nil
	}

	return nil, nil
}
