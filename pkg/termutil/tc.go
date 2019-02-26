// +build !windows

package termutil

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func tcget(fd uintptr, p *Termios) syscall.Errno {
	_, _, err := unix.Syscall(unix.SYS_IOCTL, fd, uintptr(getTermiosOp), uintptr(unsafe.Pointer(p)))
	return err
}

func tcset(fd uintptr, p *Termios) syscall.Errno {
	_, _, err := unix.Syscall(unix.SYS_IOCTL, fd, setTermiosOp, uintptr(unsafe.Pointer(p)))
	return err
}
