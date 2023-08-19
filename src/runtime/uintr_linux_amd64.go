//go:build linux && amd64

package runtime

import (
	"internal/abi"
)

type __uintr_frame struct {
	rip uint64
	rflags uint64
	rsp uint64
}

func minitUserInterrupts() {
	mp := getg().m

	fn := abi.FuncPCABI0(uintrtramp)
	_ = uintr_register_handler(fn, 0)

	mp.uintrfd = uintr_create_fd(0, 0)

	stui()
}

//go:nosplit
//go:nowritebarrierrec
func uintrtrampgo(frame *__uintr_frame, vector int32) {
	// empty handler for now
}
