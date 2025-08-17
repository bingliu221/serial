//go:build linux

package uart

import (
	"syscall"
	"unsafe"
)

var baudrateOptions = map[int]uint32{
	50:      syscall.B50,
	75:      syscall.B75,
	110:     syscall.B110,
	134:     syscall.B134,
	150:     syscall.B150,
	200:     syscall.B200,
	300:     syscall.B300,
	600:     syscall.B600,
	1200:    syscall.B1200,
	1800:    syscall.B1800,
	2400:    syscall.B2400,
	4800:    syscall.B4800,
	9600:    syscall.B9600,
	19200:   syscall.B19200,
	38400:   syscall.B38400,
	57600:   syscall.B57600,
	115200:  syscall.B115200,
	230400:  syscall.B230400,
	460800:  syscall.B460800,
	921600:  syscall.B921600,
	1500000: syscall.B1500000,
	2000000: syscall.B2000000,
	3000000: syscall.B3000000,
	4000000: syscall.B4000000,
}

func setTermio(fd uintptr, c *config) error {
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  syscall.CREAD | syscall.CLOCAL | c.FlagDataBits | c.FlagBaudrate | c.FlagStopBits | c.FlagParity,
		Ispeed: c.FlagBaudrate,
		Ospeed: c.FlagBaudrate,
	}
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = deciSecondInUint8(c.Timeout)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return errno
	}

	return nil
}

func drain(fd uintptr) error {
	TCFLSH := 0x540b
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(TCFLSH), uintptr(syscall.TCIOFLUSH))
	if errno != 0 {
		return errno
	}

	return nil
}
