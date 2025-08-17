//go:build unix

package uart

import (
	"os"
	"syscall"
	"time"
)

var databitsOptions = map[int]uint32{
	5: syscall.CS5,
	6: syscall.CS6,
	7: syscall.CS7,
	8: syscall.CS8,
}

var stopbitsOptions = map[float64]uint32{
	1: 0,
	2: syscall.CSTOPB,
}

var parityOptions = map[Parity]uint32{
	ParityNone: 0,
	ParityOdd:  syscall.PARENB | syscall.PARODD,
	ParityEven: syscall.PARENB,
}

func deciSecondInUint8(duration time.Duration) uint8 {
	return uint8(min(max(duration.Milliseconds()/100, 1), 255))
}

func open(filename string, c *config) (Port, error) {
	f, err := os.OpenFile(filename, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	err = setTermio(f.Fd(), c)
	if err != nil {
		f.Close()
		return nil, err
	}

	err = drain(f.Fd())
	if err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}
