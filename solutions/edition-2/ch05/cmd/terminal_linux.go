package main

import (
	"example.com/ensemble"
	"syscall"
	"unsafe"
)

func terminalFD(owner ensemble.ClientOwner, fd uintptr) bool {
	var state syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&state)))
	return errno == 0
}
