//go:build linux

package uart

import (
	"syscall"
	"unsafe"
)

func setTermio(fd uintptr, c *config) error {
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  syscall.CREAD | syscall.CLOCAL | c.FlagDataBits | c.FlagBaudrate | c.FlagStopBits | c.FlagParity,
		Ispeed: c.FlagBaudrate,
		Ospeed: c.FlagBaudrate,
	}
	vmin, vtime := timeoutValues(c.Timeout)
	t.Cc[syscall.VMIN] = vmin
	t.Cc[syscall.VTIME] = vtime

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
