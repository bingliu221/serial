//go:build darwin

package uart

import (
	"syscall"
	"unsafe"
)

var baudrateOptions = map[int]uint32{
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
}

func setTermio(fd uintptr, c *config) error {
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  syscall.CREAD | syscall.CLOCAL | uint64(c.FlagDataBits|c.FlagBaudrate|c.FlagStopBits|c.FlagParity),
		Ispeed: uint64(c.FlagBaudrate),
		Ospeed: uint64(c.FlagBaudrate),
	}
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = deciSecondInUint8(c.Timeout)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return errno
	}

	return nil
}

func drain(fd uintptr) error {
	var flushValue int = 0
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCFLUSH), uintptr(unsafe.Pointer(&flushValue)))
	if errno != 0 {
		return errno
	}
	return nil
}
