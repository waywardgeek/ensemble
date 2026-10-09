//go:build !darwin && !linux

package mcpstdio

import "os/exec"

func supported() bool                  { return false }
func processGroup(c *exec.Cmd)         {}
func terminate(c *exec.Cmd, kill bool) {}
