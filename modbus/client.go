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

func (c *ModbusClient) request(slaveId byte, code byte, data []byte) ([]byte, error) {
	ctx := context.Background()
	return c.handler.Send(ctx, &message{slaveId, code, data})
}

func (c *ModbusClient) ReadCoils(address uint16, count uint16) ([]bool, error) {
	return nil, ErrNotImplemented
}

func (c *ModbusClient) WriteCoil(address uint16, value bool) error {
	return ErrNotImplemented
}

func (c *ModbusClient) WriteCoils(address uint16, values []bool) error {
	return ErrNotImplemented
}

func (c *ModbusClient) ReadDiscreteInputs(address uint16, count uint16) ([]bool, error) {
	return nil, ErrNotImplemented
}

func (c *ModbusClient) ReadHoldingRegisters(slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := bytesJoin(u16be(address), u16be(count))
	data, err := c.request(slaveId, FnReadHoldingRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadInputRegisters(slaveId byte, address uint16, count uint16) ([]uint16, error) {
	data := bytesJoin(u16be(address), u16be(count))
	data, err := c.request(slaveId, FnReadInputRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) ReadWriteRegisters(slaveId byte, readAddress uint16, readCount uint16, writeAddress uint16, writeValues []uint16) ([]uint16, error) {
	data := bytesJoin(u16be(readAddress), u16be(readCount), u16be(writeAddress), u16be(uint16(len(writeValues))), u16listbe(writeValues))
	data, err := c.request(slaveId, FnReadWriteRegisters, data)
	if err != nil {
		return nil, err
	}
	return beu16list(data), nil
}

func (c *ModbusClient) WriteRegister(slaveId byte, address uint16, value uint16) error {
	data := bytesJoin(u16be(address), u16be(value))
	_, err := c.request(slaveId, FnWriteRegister, data)
	return err
}

func (c *ModbusClient) WriteRegisters(slaveId byte, address uint16, values []uint16) error {
	data := bytesJoin(u16be(address), u16be(uint16(len(values))), u16listbe(values))
	_, err := c.request(slaveId, FnWriteRegisters, data)
	return err
}