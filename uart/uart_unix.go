//go:build unix

package uart

import (
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
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

type port struct {
	file    *os.File
	rawConn syscall.RawConn
	timeout time.Duration

	deadlineMu  sync.Mutex
	hasDeadline bool
}

func (p *port) Write(b []byte) (int, error) {
	return p.file.Write(b)
}

func (p *port) Close() error {
	return p.file.Close()
}

func (p *port) SetReadDeadline(t time.Time) error {
	p.deadlineMu.Lock()
	defer p.deadlineMu.Unlock()
	p.hasDeadline = !t.IsZero()
	return p.file.SetReadDeadline(t)
}

func (p *port) Read(b []byte) (int, error) {
	if p.timeout > 0 {
		p.deadlineMu.Lock()
		var err error
		if !p.hasDeadline {
			err = p.file.SetReadDeadline(time.Now().Add(p.timeout))
		}
		p.deadlineMu.Unlock()
		if err != nil {
			return 0, err
		}
	}
	return p.file.Read(b)
}

func (p *port) FlushInput() error {
	return p.control(flushInput)
}

func (p *port) Drain() error {
	return p.control(drainOutput)
}

func (p *port) control(fn func(fd uintptr) error) error {
	var sysErr error
	if err := p.rawConn.Control(func(fd uintptr) { sysErr = fn(fd) }); err != nil {
		return err
	}
	return sysErr
}

func ioctl(fd, req, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, arg); errno != 0 {
		return errno
	}
	return nil
}

func ioctlPtr(fd, req uintptr, arg unsafe.Pointer) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(arg)); errno != 0 {
		return errno
	}
	return nil
}

func open(filename string, c *config) (Port, error) {
	f, err := os.OpenFile(filename, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	rawConn, err := f.SyscallConn()
	if err != nil {
		f.Close()
		return nil, err
	}

	p := &port{file: f, rawConn: rawConn, timeout: c.Timeout}
	err = p.control(func(fd uintptr) error {
		if err := setTermio(fd, c); err != nil {
			return err
		}
		return flushInput(fd)
	})
	if err != nil {
		f.Close()
		return nil, err
	}

	return p, nil
}
