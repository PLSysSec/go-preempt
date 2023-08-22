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
	ret := uintr_register_handler(fn, 0)
	if ret < 0 {
		print("error registering UIPI handler: ", ret, "\n")
	}

	// TODO: is it ok that we reuse gsignal's stack here or do we need a
	// separate one for UIPI?
	// Note: it seems that we have to specify a pointer to the top of the stack
	// rather than the base of the stack as with sigaltstack.
	s := &mp.gsignal.stack
	ret = uintr_alt_stack(s.hi, uint64(s.hi - s.lo), 0)
	if ret < 0 {
		print("error registering alt stack for UIPI handler: ", ret, "\n")
	}

	mp.uintrfd = uintr_create_fd(0, 0)
	if mp.uintrfd < 0 {
		print("error creating UIPI fd: ", mp.uintrfd, "\n")
	}

	stui()
}

//go:nosplit
//go:nowritebarrierrec
func uintrtrampgo(frame *__uintr_frame, vector int32) {
	// Acknowledge the preemption
	gp := getg()
	mp := gp.m
	mp.preemptGen.Add(1)
	mp.signalPending.Store(0)
}
