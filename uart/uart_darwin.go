//go:build darwin

package uart

import (
	"syscall"
	"unsafe"
)

const IOSSIOSPEED = 0x80085402

const fread = 0x1 // FREAD from <sys/fcntl.h>

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
	460800:  460800,
	500000:  500000,
	576000:  576000,
	921600:  921600,
	1000000: 1000000,
	1152000: 1152000,
	1500000: 1500000,
	2000000: 2000000,
	2500000: 2500000,
	3000000: 3000000,
	3500000: 3500000,
	4000000: 4000000,
}

func setTermio(fd uintptr, c *config) error {
	// On darwin the B* constants are the literal baud values and overlap with
	// Cflag bits, so the speed must only go into Ispeed/Ospeed.
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR,
		Cflag:  syscall.CREAD | syscall.CLOCAL | uint64(c.FlagDataBits|c.FlagStopBits|c.FlagParity),
		Ispeed: uint64(c.FlagBaudrate),
		Ospeed: uint64(c.FlagBaudrate),
	}
	if err := ioctlPtr(fd, syscall.TIOCSETA, unsafe.Pointer(&t)); err != nil {
		return err
	}

	if c.FlagBaudrate > syscall.B230400 {
		speed := uint64(c.FlagBaudrate) // speed_t is 8 bytes
		if err := ioctlPtr(fd, IOSSIOSPEED, unsafe.Pointer(&speed)); err != nil {
			return err
		}
	}

	return nil
}

func flushInput(fd uintptr) error {
	which := int32(fread)
	return ioctlPtr(fd, syscall.TIOCFLUSH, unsafe.Pointer(&which))
}

func drainOutput(fd uintptr) error {
	return ioctl(fd, syscall.TIOCDRAIN, 0)
}
