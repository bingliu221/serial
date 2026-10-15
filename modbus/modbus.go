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
	ExceptionCodeIllegalFunction                    byte = 1
	ExceptionCodeIllegalDataAddress                 byte = 2
	ExceptionCodeIllegalDataValue                   byte = 3
	ExceptionCodeServerDeviceFailure                byte = 4
	ExceptionCodeAcknowledge                        byte = 5
	ExceptionCodeServerDeviceBusy                   byte = 6
	ExceptionCodeNegativeAcknowledge                byte = 7
	ExceptionCodeMemoryParityError                  byte = 8
	ExceptionCodeGatewayPathUnavailable             byte = 10
	ExceptionCodeGatewayTargetDeviceFailedToRespond byte = 11
)

var (
	ErrNotImplemented = errors.New("not implemented")
	// ErrUnsupportedFunction is returned when a response carries a function
	// code whose frame length cannot be determined.
	ErrUnsupportedFunction = errors.New("unsupported function code")
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
	case ExceptionCodeServerDeviceFailure:
		return "server device failure"
	case ExceptionCodeAcknowledge:
		return "acknowledge"
	case ExceptionCodeServerDeviceBusy:
		return "server device busy"
	case ExceptionCodeNegativeAcknowledge:
		return "negative acknowledge"
	case ExceptionCodeMemoryParityError:
		return "memory parity error"
	case ExceptionCodeGatewayPathUnavailable:
		return "gateway path unavailable"
	case ExceptionCodeGatewayTargetDeviceFailedToRespond:
		return "gateway target device failed to respond"
	default:
		return fmt.Sprintf("unknown exception code %d", e)
	}
}

type message struct {
	slaveId byte
	code    byte
	data    []byte
}

type Handler interface {
	Send(ctx context.Context, msg *message) ([]byte, error)
	Close() error
}

type ModbusClient struct {
	handler Handler
}

type options struct {
	broadcastDelay  time.Duration
	responseTimeout time.Duration
}

type Option func(*options)

// WithBroadcastDelay sets the turnaround delay after a broadcast request.
// Default 100ms.
func WithBroadcastDelay(d time.Duration) Option {
	return func(o *options) {
		o.broadcastDelay = d
	}
}

// WithResponseTimeout sets the response timeout applied when the request
// context has no deadline. Default 1s; 0 waits indefinitely.
func WithResponseTimeout(d time.Duration) Option {
	return func(o *options) {
		o.responseTimeout = d
	}
}
