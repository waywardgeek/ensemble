package tools

import (
	"bytes"
	"errors"
	"example.com/ensemble/internal/common"
	"fmt"
	"os/exec"
)

type capture struct {
	parent    common.Registry
	data      bytes.Buffer
	limit     int
	truncated bool
}

// Always report consuming the whole write, even once retained output is full.
// os/exec must keep draining both streams until the child actually exits.
func (c *capture) Write(p []byte) (int, error) {
	n := len(p)
	available := c.limit - c.data.Len()
	if len(p) > available {
		p = p[:available]
		c.truncated = true
	}
	c.data.Write(p)
	return n, nil
}
func (c *capture) text() string {
	raw := c.data.String()
	text := textPrefix(c.parent, raw, c.limit)
	if c.truncated || len(text) < len(raw) {
		text += "\n[truncated: max_output_bytes limit reached]\n"
	}
	return text
}
func runCommand(r *Registry, a arguments) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", a["command"].(string))
	cmd.Dir = r.parent.Workspace()
	stdout := &capture{parent: r, limit: a["max_output_bytes"].(int)}
	stderr := &capture{parent: r, limit: a["max_output_bytes"].(int)}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		return "", r.failure("run_command command: shell could not run: %v", err)
	}
	return fmt.Sprintf("stdout:\n%s\nstderr:\n%s\nexit_status: %d", stdout.text(), stderr.text(), cmd.ProcessState.ExitCode()), nil
}
