package tools

import (
	"example.com/ensemble/internal/common"
	"os"
	"path/filepath"
)

func startCommand(r *Registry, job common.Job, args arguments) error {
	cwd := ""
	if value, ok := args["cwd"]; ok {
		cwd = r.path(value.(string))
		info, err := os.Stat(cwd)
		if err != nil {
			return r.failure("run_command cwd: %v", err)
		}
		if !info.IsDir() {
			return r.failure("run_command cwd: not a directory")
		}
		cwd, err = filepath.Abs(cwd)
		if err != nil {
			return r.failure("run_command cwd: %v", err)
		}
	}
	return job.StartProcess(args["command"].(string), cwd)
}
