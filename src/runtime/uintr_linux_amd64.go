//go:build linux && amd64

package runtime

func minitUserInterrupts() {
	mp := getg().m

	mp.uintr_fd = uintr_create_fd(0, 0)
}
