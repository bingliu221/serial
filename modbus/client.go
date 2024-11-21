package modbus

import (
	"context"
	"fmt"

	"github.com/bingliu221/serial/uart"
)

func NewModbusClient(mode string, path string, cfg uart.Config) (*ModbusClient, error) {
	var handler Handler
	var err error

	switch mode {
	case "rtu":
		handler, err = NewModbusRtuClientHandler(path, cfg)
		if err != nil {
			return nil, err
		}

	case "ascii":
		return nil, ErrNotImplemented

	default:
		return nil, fmt.Errorf("unsupported mode %s", mode)
	}

	return &ModbusClient{handler: handler}, nil
}

func (c *ModbusClient) Release() error {
	return c.handler.Close()
}

func (c *ModbusClient) RawRequest(ctx context.Context, slaveId byte, code byte, data []byte) ([]byte, error) {
	return c.request(ctx, slaveId, code, data)
}

func (c *ModbusClient) request(ctx context.Context, slaveId byte, code byte, data []byte) ([]byte, error) {
	return c.handler.Send(ctx, &message{slaveId, code, data})
}

func (c *ModbusClient) ReadCoils(ctx context.Context, slaveId byte, address uint16, count uint16) ([]bool, error) {
	return nil, ErrNotImplemented
}

func (c *ModbusClient) WriteCoil(ctx context.Context, slaveId byte, address uint16, value bool) error {
	return ErrNotImplemented
}

func (c *ModbusClient) WriteCoils(ctx context.Context, slaveId byte, address uint16, values []bool) error {
	return ErrNotImplemented
}

func (c *ModbusClient) ReadDiscreteInputs(ctx context.Context, slaveId byte, ddress uint16, count uint16) ([]bool, error) {
	return nil, ErrNotImplemented
}

func (c *ModbusClient) ReadHoldingRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := bytesJoin(u16be(address), u16be(count))
	data, err := c.request(ctx, slaveId, FnReadHoldingRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadInputRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := bytesJoin(u16be(address), u16be(count))
	data, err := c.request(ctx, slaveId, FnReadInputRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadWriteRegisters(ctx context.Context, slaveId byte, readAddress uint16, readCount uint16, writeAddress uint16, writeValues []uint16) ([]uint16, error) {
	data := bytesJoin(u16be(readAddress), u16be(readCount), u16be(writeAddress), u16be(uint16(len(writeValues))), u16listbe(writeValues))
	data, err := c.request(ctx, slaveId, FnReadWriteRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) WriteRegister(ctx context.Context, slaveId byte, address uint16, value uint16) error {
	data := bytesJoin(u16be(address), u16be(value))
	_, err := c.request(ctx, slaveId, FnWriteRegister, data)
	return err
}

func (c *ModbusClient) WriteRegisters(ctx context.Context, slaveId byte, address uint16, values []uint16) error {
	data := bytesJoin(u16be(address), u16be(uint16(len(values))), u16listbe(values))
	_, err := c.request(ctx, slaveId, FnWriteRegisters, data)
	return err
}
