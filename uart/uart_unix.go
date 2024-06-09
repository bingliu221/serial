//go:build unix

package uart

import (
	"fmt"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
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

func open(filename string, cfg Config) (Port, error) {
	baudrate, ok := map[int]uint32{
		50:      unix.B50,
		75:      unix.B75,
		110:     unix.B110,
		134:     unix.B134,
		150:     unix.B150,
		200:     unix.B200,
		300:     unix.B300,
		600:     unix.B600,
		1200:    unix.B1200,
		1800:    unix.B1800,
		2400:    unix.B2400,
		4800:    unix.B4800,
		9600:    unix.B9600,
		19200:   unix.B19200,
		38400:   unix.B38400,
		57600:   unix.B57600,
		115200:  unix.B115200,
		230400:  unix.B230400,
		460800:  unix.B460800,
		500000:  unix.B500000,
		576000:  unix.B576000,
		921600:  unix.B921600,
		1000000: unix.B1000000,
		1152000: unix.B1152000,
		1500000: unix.B1500000,
		2000000: unix.B2000000,
		2500000: unix.B2500000,
		3000000: unix.B3000000,
		3500000: unix.B3500000,
		4000000: unix.B4000000,
	}[cfg.Baudrate]
	if !ok {
		return nil, fmt.Errorf("baudrate %d not supported", cfg.Baudrate)
	}

	databits, ok := map[int]uint32{
		5: unix.CS5,
		6: unix.CS6,
		7: unix.CS7,
		8: unix.CS8,
	}[cfg.DataBits]
	if !ok {
		return nil, fmt.Errorf("databits %d not supported", cfg.DataBits)
	}

	stopbits, ok := map[float64]uint32{
		1: 0,
		2: unix.CSTOPB,
	}[cfg.StopBits]
	if !ok {
		return nil, fmt.Errorf("stopbits %f not supported", cfg.StopBits)
	}

	parity, ok := map[string]uint32{
		ParityNone: 0,
		ParityOdd:  unix.PARENB | unix.PARODD,
		ParityEven: unix.PARENB,
	}[cfg.Parity]
	if !ok {
		return nil, fmt.Errorf("invalid parity %s", cfg.Parity)
	}

	f, err := os.OpenFile(filename, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	t := unix.Termios{
		Iflag:  unix.IGNPAR,
		Cflag:  unix.CREAD | unix.CLOCAL | baudrate | databits | stopbits | parity,
		Ispeed: baudrate,
		Ospeed: baudrate,
	}
	vmin, vtime := timeoutValues(cfg.ReadTimeout)
	t.Cc[unix.VMIN] = vmin
	t.Cc[unix.VTIME] = vtime

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uintptr(unix.TCSETS), uintptr(unsafe.Pointer(&t))); errno != 0 {
		f.Close()
		return nil, errno
	}

	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), uintptr(unix.TCFLSH), uintptr(unix.TCIOFLUSH))
	if errno != 0 {
		f.Close()
		return nil, errno
	}

	return f, nil
}
