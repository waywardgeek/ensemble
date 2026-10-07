//go:build !darwin && !linux

package main

import "example.com/ensemble"

// Other systems support explicit chat; automatic terminal detection currently
// supports macOS and Linux only. Never mistake a character device for a TTY.
func terminalFD(owner ensemble.ClientOwner, fd uintptr) bool { return false }
