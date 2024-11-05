//go:build darwin

package uart

import (
	"syscall"
	"unsafe"
)

func setTermio(fd uintptr, c *config) error {
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  syscall.CREAD | syscall.CLOCAL | uint64(c.FlagDataBits|c.FlagBaudrate|c.FlagStopBits|c.FlagParity),
		Ispeed: uint64(c.FlagBaudrate),
		Ospeed: uint64(c.FlagBaudrate),
	}
	vmin, vtime := timeoutValues(c.ReadTimeout)
	t.Cc[syscall.VMIN] = vmin
	t.Cc[syscall.VTIME] = vtime

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return errno
	}

	return nil
}

func drain(fd uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCIOFLUSH), uintptr(syscall.TCIOFLUSH))
	if errno != 0 {
		return errno
	}

	return nil
}
