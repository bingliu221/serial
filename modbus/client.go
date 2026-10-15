package modbus

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/bingliu221/serial/uart"
)

func NewModbusClient(mode string, path string, cfg uart.Config, opts ...Option) (*ModbusClient, error) {
	var handler Handler
	var err error

	switch mode {
	case "rtu":
		handler, err = NewModbusRtuClientHandler(path, cfg, opts...)
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
	data := u16bePair(address, count)
	data, err := c.request(ctx, slaveId, FnReadCoils, data)
	if err != nil {
		return nil, err
	}
	values := beboollist(data)
	if len(values) < int(count) {
		return nil, fmt.Errorf("expected %d values, got %d", count, len(values))
	}
	return values[:count], nil
}

func (c *ModbusClient) WriteCoil(ctx context.Context, slaveId byte, address uint16, value bool) error {
	data := u16bePair(address, boolu16(value))
	_, err := c.request(ctx, slaveId, FnWriteCoil, data)
	return err
}

func (c *ModbusClient) WriteCoils(ctx context.Context, slaveId byte, address uint16, values []bool) error {
	byteCount := (len(values) + 7) / 8
	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], uint16(len(values)))
	data[4] = byte(byteCount)
	putBoolList(data[5:], values)
	_, err := c.request(ctx, slaveId, FnWriteCoils, data)
	return err
}

func (c *ModbusClient) ReadDiscreteInputs(ctx context.Context, slaveId byte, address uint16, count uint16) ([]bool, error) {
	data := u16bePair(address, count)
	data, err := c.request(ctx, slaveId, FnReadDiscreteInputs, data)
	if err != nil {
		return nil, err
	}
	values := beboollist(data)
	if len(values) < int(count) {
		return nil, fmt.Errorf("expected %d values, got %d", count, len(values))
	}
	return values[:count], nil
}

func (c *ModbusClient) ReadHoldingRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := u16bePair(address, count)
	data, err := c.request(ctx, slaveId, FnReadHoldingRegisters, data)
	if err != nil {
		return nil, err
	}
	if len(data) != int(count)*2 {
		return nil, fmt.Errorf("expected %d bytes, got %d", int(count)*2, len(data))
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadInputRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := u16bePair(address, count)
	data, err := c.request(ctx, slaveId, FnReadInputRegisters, data)
	if err != nil {
		return nil, err
	}
	if len(data) != int(count)*2 {
		return nil, fmt.Errorf("expected %d bytes, got %d", int(count)*2, len(data))
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadWriteRegisters(ctx context.Context, slaveId byte, readAddress uint16, readCount uint16, writeAddress uint16, writeValues []uint16) ([]uint16, error) {
	byteCount := len(writeValues) * 2
	data := make([]byte, 9+byteCount)
	binary.BigEndian.PutUint16(data[0:2], readAddress)
	binary.BigEndian.PutUint16(data[2:4], readCount)
	binary.BigEndian.PutUint16(data[4:6], writeAddress)
	binary.BigEndian.PutUint16(data[6:8], uint16(len(writeValues)))
	data[8] = byte(byteCount)
	putU16ListBE(data[9:], writeValues)
	data, err := c.request(ctx, slaveId, FnReadWriteRegisters, data)
	if err != nil {
		return nil, err
	}
	if len(data) != int(readCount)*2 {
		return nil, fmt.Errorf("expected %d bytes, got %d", int(readCount)*2, len(data))
	}
	return beu16list(data), nil
}

func (c *ModbusClient) WriteRegister(ctx context.Context, slaveId byte, address uint16, value uint16) error {
	data := u16bePair(address, value)
	_, err := c.request(ctx, slaveId, FnWriteRegister, data)
	return err
}

func (c *ModbusClient) WriteRegisters(ctx context.Context, slaveId byte, address uint16, values []uint16) error {
	byteCount := len(values) * 2
	data := make([]byte, 5+byteCount)
	binary.BigEndian.PutUint16(data[0:2], address)
	binary.BigEndian.PutUint16(data[2:4], uint16(len(values)))
	data[4] = byte(byteCount)
	putU16ListBE(data[5:], values)
	_, err := c.request(ctx, slaveId, FnWriteRegisters, data)
	return err
}
