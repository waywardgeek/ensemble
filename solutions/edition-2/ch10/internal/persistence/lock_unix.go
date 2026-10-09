//go:build darwin || linux

package persistence

import (
	"errors"
	"os"
	"syscall"
)

func supported() bool { return true }
func lockFile(f *os.File, shared bool) error {
	syscall.CloseOnExec(int(f.Fd()))
	op := syscall.LOCK_EX
	if shared {
		op = syscall.LOCK_SH
	}
	return syscall.Flock(int(f.Fd()), op|syscall.LOCK_NB)
}
func inUse(err error) bool {
	return errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN)
}
