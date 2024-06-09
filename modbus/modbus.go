package modbus

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	FnReadCoils          byte = 1
	FnReadDiscreteInputs byte = 2
	FnWriteCoil          byte = 5
	FnWriteCoils         byte = 15

	FnReadHoldingRegisters byte = 3
	FnReadInputRegisters   byte = 4
	FnWriteRegister        byte = 6
	FnWriteRegisters       byte = 16
	FnMaskWriteRegister    byte = 22
	FnReadWriteRegisters   byte = 23
)

const (
	ExceptionCodeIllegalFunction              byte = 1
	ExceptionCodeIllegalDataAddress           byte = 2
	ExceptionCodeIllegalDataValue             byte = 3
	ExceptionCodeServerFailure                byte = 4
	ExceptionCodeUnknownAddress               byte = 5
	ExceptionCodeSlaveDeviceBusy              byte = 6
	ExceptionCodeMemoryParityError            byte = 7
	ExceptionCodeExceptionResponseLengthError byte = 8
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type ExceptionError byte

func (e ExceptionError) Error() string {
	switch byte(e) {
	case ExceptionCodeIllegalFunction:
		return "illegal function"
	case ExceptionCodeIllegalDataAddress:
		return "illegal data address"
	case ExceptionCodeIllegalDataValue:
		return "illegal data value"
	case ExceptionCodeServerFailure:
		return "server failure"
	case ExceptionCodeUnknownAddress:
		return "unknown address"
	case ExceptionCodeSlaveDeviceBusy:
		return "slave device busy"
	case ExceptionCodeMemoryParityError:
		return "memory parity error"
	case ExceptionCodeExceptionResponseLengthError:
		return "exception response length error"
	default:
		return fmt.Sprintf("unknown exception code %d", e)
	}
}

type message struct {
	slaveId byte
	code    byte
	data    []byte
}

type timedMessage struct {
	message
	timestamp time.Time
}

type Handler interface {
	Send(ctx context.Context, msg *message) ([]byte, error)
	Close() error
}

type ModbusClient struct {
	handler Handler
}
