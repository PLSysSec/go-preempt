//go:build linux && amd64

package runtime

import (
	"internal/abi"
	"internal/goarch"
	"unsafe"
)

type __uintr_frame struct {
	rip    uintptr
	rflags uint64
	rsp    uintptr
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
	ret = uintr_alt_stack(s.hi, uint64(s.hi-s.lo), 0)
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
	gp := getg()
	mp := gp.m
	if gp == nil || (mp != nil && mp.isExtraInC) {
		print("warning: unhandled case in uintrtrampgo\n")
	}

	// switch to the signal g
	setg(mp.gsignal)

	// check that we are on the alternate stack
	sp := uintptr(unsafe.Pointer(&vector))
	if sp < mp.gsignal.stack.lo || sp >= mp.gsignal.stack.hi {
		print("error: not on alternate stack\n")
	}

	uintrhandler(gp, frame)

	setg(gp)
}

// uintrhandler is invoked when a UIPI occurs. The global g will be
// set to a gsignal goroutine and we will be running on the alternate
// signal stack. The parameter gp will be the value of the global g
// when the UIPI was delivered. The frame is the frame that was pushed
// onto the stack during delivery of the UIPI.
//
// The garbage collector may have stopped the world, so write barriers
// are not allowed.
//
//go:nowritebarrierrec
func uintrhandler(gp *g, frame *__uintr_frame) {
	if wantAsyncPreempt(gp) {
		ok, newpc := isAsyncSafePoint(gp, frame.rip, frame.rsp, 0)
		if ok {
			// Adjust the PC and inject a call to asyncPreempt
			pushCall(abi.FuncPCABI0(asyncPreempt), newpc, frame)
		}
	}

	// Acknowledge the preemption
	gp.m.preemptGen.Add(1)
	gp.m.signalPending.Store(0)
}

func pushCall(targetPC uintptr, resumePC uintptr, frame *__uintr_frame) {
	// Make it look like we called target at resumePC
	sp := frame.rsp
	sp -= goarch.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = resumePC
	frame.rsp = sp
	frame.rip = targetPC
}
