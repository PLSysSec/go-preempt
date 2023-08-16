//go:build linux && amd64

package runtime

import (
	"internal/abi"
)

func minitUserInterrupts() {
	mp := getg().m

	fn := abi.FuncPCABI0(uintrtramp)
	_ = uintr_register_handler(fn, 0)

	mp.uintrfd = uintr_create_fd(0, 0)

	stui()
}

func uintrtrampgo() {
	print("uintrtampgo!")
}
