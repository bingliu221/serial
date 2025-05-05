package uart

import (
	"fmt"
	"io"
	"time"
)

type Port interface {
	io.Reader
	io.Writer
	io.Closer
}

type Parity string

const (
	ParityNone Parity = "None"
	ParityOdd  Parity = "Odd"
	ParityEven Parity = "Even"
)

type Config struct {
	DataBits int
	Baudrate int
	Parity   Parity
	StopBits float64
	Timeout  time.Duration
}

func (cfg Config) config() (*config, error) {
	baudrate, ok := baudrateOptions[cfg.Baudrate]
	if !ok {
		return nil, fmt.Errorf("baudrate %d not supported", cfg.Baudrate)
	}

	databits, ok := databitsOptions[cfg.DataBits]
	if !ok {
		return nil, fmt.Errorf("databits %d not supported", cfg.DataBits)
	}

	stopbits, ok := stopbitsOptions[cfg.StopBits]
	if !ok {
		return nil, fmt.Errorf("stopbits %f not supported", cfg.StopBits)
	}

	parity, ok := parityOptions[cfg.Parity]
	if !ok {
		return nil, fmt.Errorf("invalid parity %s", cfg.Parity)
	}

	return &config{
		FlagDataBits: databits,
		FlagBaudrate: baudrate,
		FlagParity:   parity,
		FlagStopBits: stopbits,
		Timeout:      cfg.Timeout,
	}, nil
}

type config struct {
	FlagDataBits uint32
	FlagBaudrate uint32
	FlagParity   uint32
	FlagStopBits uint32
	Timeout      time.Duration
}

func Open(path string, cfg Config) (Port, error) {
	c, err := cfg.config()
	if err != nil {
		return nil, err
	}

	return open(path, c)
}
