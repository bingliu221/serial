package modbus

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"sync"
	"time"

	"github.com/bingliu221/serial/uart"
)

type ModbusRtuClientHandler struct {
	port    uart.Port
	readCtx context.Context

	sendMutex sync.Mutex

	expectResp      bool
	expectRespMutex sync.Mutex
	respC           chan *timedMessage
}

func NewModbusRtuClientHandler(path string, cfg uart.Config) (*ModbusRtuClientHandler, error) {
	port, err := uart.Open(path, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	h := &ModbusRtuClientHandler{
		port:    port,
		readCtx: ctx,
		respC:   make(chan *timedMessage, 1),
	}

	go h.readContinuously(cancel)

	return h, nil
}

func (h *ModbusRtuClientHandler) Close() error {
	h.port.Close()
	<-h.readCtx.Done()
	return nil
}

func (h *ModbusRtuClientHandler) readResp() (*timedMessage, error) {
	head := make([]byte, 3)
	_, err := io.ReadFull(h.port, head)
	if err != nil {
		return nil, err
	}

	code := head[1]
	failed := code&0x80 == 0x80

	remain := 2
	if !failed {
		switch code {
		case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
			count := int(head[2])
			remain += count
		case FnWriteCoil, FnWriteCoils, FnWriteRegister, FnWriteRegisters:
			remain += 3
		}
	}

	tail := make([]byte, remain)

	_, err = io.ReadFull(h.port, tail)
	if err != nil {
		return nil, err
	}

	adu := bytes.Join([][]byte{head, tail}, nil)

	err = h.valify(adu)
	if err != nil {
		return nil, err
	}

	return &timedMessage{
		message{
			slaveId: adu[0],
			code:    adu[1],
			data:    adu[2 : len(adu)-2],
		},
		time.Now(),
	}, nil
}

func (h *ModbusRtuClientHandler) readContinuously(cancel context.CancelFunc) {
	defer cancel()

	for {
		resp, err := h.readResp()
		if err != nil {
			if errors.Is(err, fs.ErrClosed) {
				break
			}
			log.Print(err)
			continue
		}

		h.expectRespMutex.Lock()
		if h.expectResp {
			h.respC <- resp
			h.expectResp = false
		}
		h.expectRespMutex.Unlock()
	}
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
	sum := crc(adu[:2+len(data)])
	copy(adu[2+len(data):], sum)

	return adu, nil
}

func (h *ModbusRtuClientHandler) valify(adu []byte) error {
	if len(adu) < 4 {
		return errors.New("invalid adu size")
	}

	crcIdx := len(adu) - 2

	sum := crc(adu[:crcIdx])
	if !bytes.Equal(sum, adu[crcIdx:]) {
		log.Print(sum, adu[crcIdx:])
		return errors.New("crc not matched")
	}

	return nil
}

func (h *ModbusRtuClientHandler) Send(ctx context.Context, req *message) ([]byte, error) {
	h.sendMutex.Lock()
	defer h.sendMutex.Unlock()

	adu, err := h.frameEncode(req)
	if err != nil {
		return nil, err
	}

	sentTime := time.Now()

	_, err = h.port.Write(adu)
	if err != nil {
		return nil, err
	}

	if req.slaveId == 0 {
		// broadcast, no response
		return nil, nil
	}

	h.expectRespMutex.Lock()
	h.expectResp = true
	h.expectRespMutex.Unlock()

	var resp *timedMessage

LOOP:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case resp = <-h.respC:
			if resp == nil {
				return nil, errors.New("port closed")
			}
			if resp.timestamp.After(sentTime) {
				break LOOP
			}
		}
	}

	if resp.slaveId != req.slaveId || resp.code&0x7F != req.code {
		return nil, errors.New("mismatched response")
	}

	failed := resp.code&0x80 == 0x80
	if failed {
		exception := resp.data[0]
		return nil, ExceptionError(exception)
	}

	switch resp.code {
	case FnReadDiscreteInputs, FnReadCoils, FnReadInputRegisters, FnReadHoldingRegisters, FnReadWriteRegisters:
		return resp.data[1:], nil
	case FnWriteCoil, FnWriteCoils, FnWriteRegister, FnWriteRegisters:
		return nil, nil
	}

	return nil, nil
}
