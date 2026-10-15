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
	// The speed is carried by the CBAUD bits of Cflag.
	t := syscall.Termios{
		Iflag: syscall.IGNPAR,
		Cflag: syscall.CREAD | syscall.CLOCAL | c.FlagDataBits | c.FlagBaudrate | c.FlagStopBits | c.FlagParity,
	}
	return ioctlPtr(fd, syscall.TCSETS, unsafe.Pointer(&t))
}

func flushInput(fd uintptr) error {
	return ioctl(fd, ioctlTCFLSH, syscall.TCIFLUSH)
}

func drainOutput(fd uintptr) error {
	// TCSBRK with a non-zero argument is tcdrain(3).
	return ioctl(fd, ioctlTCSBRK, 1)
}
