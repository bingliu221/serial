//go:build unix

package uart

import (
	"os"
	"syscall"
	"time"
)

func timeoutValues(timeout time.Duration) (uint8, uint8) {
	if timeout <= 0 {
		return 1, 0
	}
	deciSecond := timeout.Milliseconds() / 100
	if deciSecond < 1 {
		deciSecond = 1
	}
	if deciSecond > 255 {
		deciSecond = 255
	}
	return 0, uint8(deciSecond)
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
