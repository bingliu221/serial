package uart

import (
	"io"
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

func Open(path string, cfg Config) (Port, error) {
	return open(path, cfg)
}
