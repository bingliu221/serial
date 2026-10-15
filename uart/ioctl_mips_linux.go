//go:build linux && (mips || mipsle || mips64 || mips64le)

package uart

// arch/mips/include/uapi/asm/ioctls.h
const (
	ioctlTCSBRK = 0x5405
	ioctlTCFLSH = 0x5407
)
