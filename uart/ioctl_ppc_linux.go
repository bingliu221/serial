//go:build linux && (ppc64 || ppc64le)

package uart

// arch/powerpc/include/uapi/asm/ioctls.h
const (
	ioctlTCSBRK = 0x2000741d
	ioctlTCFLSH = 0x2000741f
)
