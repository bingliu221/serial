//go:build linux && !(mips || mipsle || mips64 || mips64le || ppc64 || ppc64le)

package uart

// asm-generic/ioctls.h; not exported by package syscall on every arch.
const (
	ioctlTCSBRK = 0x5409
	ioctlTCFLSH = 0x540B
)
