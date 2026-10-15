package modbus

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bingliu221/serial/uart"
)

type ModbusRtuClientHandler struct {
	port            uart.Port
	broadcastDelay  time.Duration
	responseTimeout time.Duration
	frameGap        time.Duration

	sendMutex sync.Mutex
	lastIO    time.Time
}

func NewModbusRtuClientHandler(path string, cfg uart.Config, opts ...Option) (*ModbusRtuClientHandler, error) {
	o := options{
		broadcastDelay:  100 * time.Millisecond,
		responseTimeout: time.Second,
	}
	for _, opt := range opts {
		opt(&o)
	}

	port, err := uart.Open(path, cfg)
	if err != nil {
		return nil, err
	}

	return &ModbusRtuClientHandler{
		port:            port,
		broadcastDelay:  o.broadcastDelay,
		responseTimeout: o.responseTimeout,
		frameGap:        frameGap(cfg.Baudrate),
	}, nil
}

// frameGap returns the t3.5 inter-frame silence. Above 19200 baud the spec
// fixes it at 1.75ms.
func frameGap(baudrate int) time.Duration {
	if baudrate > 19200 {
		return 1750 * time.Microsecond
	}
	// 11 bits per character, 3.5 characters.
	return time.Duration(38_500_000/baudrate) * time.Microsecond
}

func (h *ModbusRtuClientHandler) Close() error {
	return h.port.Close()
}

func (h *ModbusRtuClientHandler) readFrame() (*message, error) {
	var buf [260]byte
	if _, err := io.ReadFull(h.port, buf[:3]); err != nil {
		return nil, err
	}

	code := buf[1]
	remain := 2
	if code&0x80 == 0 {
		switch code {
		case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
			remain += int(buf[2])
		case FnWriteCoil, FnWriteCoils, FnWriteRegister, FnWriteRegisters:
			remain += 3
		case FnMaskWriteRegister:
			remain += 5
		default:
			return nil, fmt.Errorf("%w: %#02x", ErrUnsupportedFunction, code)
		}
	}

	total := 3 + remain
	if _, err := io.ReadFull(h.port, buf[3:total]); err != nil {
		return nil, err
	}

	adu := buf[:total]
	if err := h.verifyCRC(adu); err != nil {
		return nil, err
	}

	return &message{
		slaveId: adu[0],
		code:    adu[1],
		data:    adu[2 : total-2],
	}, nil
}

func (h *ModbusRtuClientHandler) readWithContext(ctx context.Context) (*message, error) {
	deadline, _ := ctx.Deadline()
	if err := h.port.SetReadDeadline(deadline); err != nil {
		return nil, err
	}

	fired := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(fired)
		_ = h.port.SetReadDeadline(time.Now())
	})
	defer func() {
		// If the callback already started, wait for it so it cannot touch
		// the deadline of a later request.
		if !stop() {
			<-fired
		}
		_ = h.port.SetReadDeadline(time.Time{})
	}()

	resp, err := h.readFrame()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}
	return resp, nil
}

func (h *ModbusRtuClientHandler) encodeFrame(msg *message) ([]byte, error) {
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

func (h *ModbusRtuClientHandler) verifyCRC(adu []byte) error {
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

func parseResponse(req, resp *message) ([]byte, error) {
	if resp.slaveId != req.slaveId || resp.code&0x7F != req.code {
		return nil, errors.New("mismatched response")
	}

	if resp.code&0x80 != 0 {
		if len(resp.data) == 0 {
			return nil, errors.New("invalid exception response")
		}
		return nil, ExceptionError(resp.data[0])
	}

	switch resp.code {
	case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
		if len(resp.data) < 1 {
			return nil, errors.New("invalid response data")
		}
		return resp.data[1:], nil
	case FnWriteCoil, FnWriteRegister, FnWriteCoils, FnWriteRegisters:
		// Echo of address + value/quantity.
		if len(req.data) < 4 || !bytes.Equal(resp.data, req.data[:4]) {
			return nil, errors.New("write response echo mismatch")
		}
		return nil, nil
	case FnMaskWriteRegister:
		if !bytes.Equal(resp.data, req.data) {
			return nil, errors.New("write response echo mismatch")
		}
		return nil, nil
	default:
		return nil, nil
	}
}

// waitFrameGap keeps at least t3.5 of bus silence since the last frame.
func (h *ModbusRtuClientHandler) waitFrameGap(ctx context.Context) error {
	wait := time.Until(h.lastIO.Add(h.frameGap))
	if wait <= 0 {
		return nil
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (h *ModbusRtuClientHandler) write(ctx context.Context, adu []byte) error {
	if err := h.waitFrameGap(ctx); err != nil {
		return err
	}
	if err := h.port.FlushInput(); err != nil {
		return err
	}
	if _, err := h.port.Write(adu); err != nil {
		return err
	}
	// Wait until the frame is physically on the wire so timing that follows
	// (turnaround delay, response timeout) starts from the right moment.
	if err := h.port.Drain(); err != nil {
		return err
	}
	h.lastIO = time.Now()
	return nil
}

func (h *ModbusRtuClientHandler) Send(ctx context.Context, req *message) ([]byte, error) {
	h.sendMutex.Lock()
	defer h.sendMutex.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	adu, err := h.encodeFrame(req)
	if err != nil {
		return nil, err
	}

	if err = h.write(ctx, adu); err != nil {
		return nil, err
	}

	if req.slaveId == 0 {
		if h.broadcastDelay > 0 {
			t := time.NewTimer(h.broadcastDelay)
			defer t.Stop()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-t.C:
			}
		}
		return nil, nil
	}

	if _, ok := ctx.Deadline(); !ok && h.responseTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.responseTimeout)
		defer cancel()
	}

	resp, err := h.readWithContext(ctx)
	h.lastIO = time.Now()
	if err != nil {
		return nil, err
	}

	return parseResponse(req, resp)
}
