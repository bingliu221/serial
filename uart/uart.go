package uart

import (
	"fmt"
	"io"
	"syscall"
	"time"
)

type Port interface {
	io.Reader
	io.Writer
	io.Closer
}

const (
	ParityNone string = "None"
	ParityOdd  string = "Odd"
	ParityEven string = "Even"
)

type Config struct {
	DataBits    int
	Baudrate    int
	Parity      string
	StopBits    float64
	ReadTimeout time.Duration
}

func (cfg Config) config() (*config, error) {
	baudrate, ok := map[int]uint32{
		50:     syscall.B50,
		75:     syscall.B75,
		110:    syscall.B110,
		134:    syscall.B134,
		150:    syscall.B150,
		200:    syscall.B200,
		300:    syscall.B300,
		600:    syscall.B600,
		1200:   syscall.B1200,
		1800:   syscall.B1800,
		2400:   syscall.B2400,
		4800:   syscall.B4800,
		9600:   syscall.B9600,
		19200:  syscall.B19200,
		38400:  syscall.B38400,
		57600:  syscall.B57600,
		115200: syscall.B115200,
		230400: syscall.B230400,
	}[cfg.Baudrate]
	if !ok {
		return nil, fmt.Errorf("baudrate %d not supported", cfg.Baudrate)
	}

	databits, ok := map[int]uint32{
		5: syscall.CS5,
		6: syscall.CS6,
		7: syscall.CS7,
		8: syscall.CS8,
	}[cfg.DataBits]
	if !ok {
		return nil, fmt.Errorf("databits %d not supported", cfg.DataBits)
	}

	stopbits, ok := map[float64]uint32{
		1: 0,
		2: syscall.CSTOPB,
	}[cfg.StopBits]
	if !ok {
		return nil, fmt.Errorf("stopbits %f not supported", cfg.StopBits)
	}

	parity, ok := map[string]uint32{
		ParityNone: 0,
		ParityOdd:  syscall.PARENB | syscall.PARODD,
		ParityEven: syscall.PARENB,
	}[cfg.Parity]
	if !ok {
		return nil, fmt.Errorf("invalid parity %s", cfg.Parity)
	}

	return &config{
		FlagDataBits: databits,
		FlagBaudrate: baudrate,
		FlagParity:   parity,
		FlagStopBits: stopbits,
		ReadTimeout:  cfg.ReadTimeout,
	}, nil
}

type config struct {
	FlagDataBits uint32
	FlagBaudrate uint32
	FlagParity   uint32
	FlagStopBits uint32
	ReadTimeout  time.Duration
}

func Open(path string, cfg Config) (Port, error) {
	c, err := cfg.config()
	if err != nil {
		return nil, err
	}

	return open(path, c)
}
