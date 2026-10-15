package modbus

import (
	"context"
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

func (c *ModbusClient) readBits(ctx context.Context, slaveId byte, code byte, address uint16, count uint16) ([]bool, error) {
	if count < 1 || count > 2000 {
		return nil, fmt.Errorf("bit count %d out of range (1..2000)", count)
	}
	resp, err := c.request(ctx, slaveId, code, encodeUint16s(address, count))
	if err != nil {
		return nil, err
	}
	return decodeBools(resp, count)
}

func (c *ModbusClient) readRegisters(ctx context.Context, slaveId byte, code byte, address uint16, count uint16) ([]uint16, error) {
	if count < 1 || count > 125 {
		return nil, fmt.Errorf("register count %d out of range (1..125)", count)
	}
	resp, err := c.request(ctx, slaveId, code, encodeUint16s(address, count))
	if err != nil {
		return nil, err
	}
	return decodeUint16s(resp, count)
}

func (c *ModbusClient) ReadCoils(ctx context.Context, slaveId byte, address uint16, count uint16) ([]bool, error) {
	return c.readBits(ctx, slaveId, FnReadCoils, address, count)
}

func (c *ModbusClient) ReadDiscreteInputs(ctx context.Context, slaveId byte, address uint16, count uint16) ([]bool, error) {
	return c.readBits(ctx, slaveId, FnReadDiscreteInputs, address, count)
}

func (c *ModbusClient) WriteCoil(ctx context.Context, slaveId byte, address uint16, value bool) error {
	_, err := c.request(ctx, slaveId, FnWriteCoil, encodeUint16s(address, encodeBool(value)))
	return err
}

func (c *ModbusClient) WriteCoils(ctx context.Context, slaveId byte, address uint16, values []bool) error {
	if len(values) < 1 || len(values) > 1968 {
		return fmt.Errorf("coil count %d out of range (1..1968)", len(values))
	}
	data := encodeWithPayload(encodeBools(values), address, uint16(len(values)))
	_, err := c.request(ctx, slaveId, FnWriteCoils, data)
	return err
}

func (c *ModbusClient) ReadHoldingRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	return c.readRegisters(ctx, slaveId, FnReadHoldingRegisters, address, count)
}

func (c *ModbusClient) ReadInputRegisters(ctx context.Context, slaveId byte, address uint16, count uint16) ([]uint16, error) {
	return c.readRegisters(ctx, slaveId, FnReadInputRegisters, address, count)
}

func (c *ModbusClient) WriteRegister(ctx context.Context, slaveId byte, address uint16, value uint16) error {
	_, err := c.request(ctx, slaveId, FnWriteRegister, encodeUint16s(address, value))
	return err
}

func (c *ModbusClient) WriteRegisters(ctx context.Context, slaveId byte, address uint16, values []uint16) error {
	if len(values) < 1 || len(values) > 123 {
		return fmt.Errorf("register count %d out of range (1..123)", len(values))
	}
	data := encodeWithPayload(encodeUint16s(values...), address, uint16(len(values)))
	_, err := c.request(ctx, slaveId, FnWriteRegisters, data)
	return err
}

func (c *ModbusClient) ReadWriteRegisters(ctx context.Context, slaveId byte, readAddress uint16, readCount uint16, writeAddress uint16, writeValues []uint16) ([]uint16, error) {
	if readCount < 1 || readCount > 125 {
		return nil, fmt.Errorf("read register count %d out of range (1..125)", readCount)
	}
	if len(writeValues) < 1 || len(writeValues) > 121 {
		return nil, fmt.Errorf("write register count %d out of range (1..121)", len(writeValues))
	}
	data := encodeWithPayload(encodeUint16s(writeValues...), readAddress, readCount, writeAddress, uint16(len(writeValues)))
	resp, err := c.request(ctx, slaveId, FnReadWriteRegisters, data)
	if err != nil {
		return nil, err
	}
	return decodeUint16s(resp, readCount)
}
