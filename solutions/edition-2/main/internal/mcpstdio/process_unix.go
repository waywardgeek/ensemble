//go:build darwin || linux

package mcpstdio

import (
	"os/exec"
	"syscall"
)

func supported() bool          { return true }
func processGroup(c *exec.Cmd) { c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }
func terminate(c *exec.Cmd, kill bool) {
	signal := syscall.SIGTERM
	if kill {
		signal = syscall.SIGKILL
	}
	_ = syscall.Kill(-c.Process.Pid, signal)
}
