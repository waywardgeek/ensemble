package common

import (
	"io/fs"
	"os"
	"os/exec"
)

// Sandbox is an Agent child. File operations remain rooted all the way through
// the open; a checked pathname alone is not a capability.
type Sandbox interface {
	Parent() Agent
	Root() string
	ValidatePath(string) (string, error)
	ValidateCommand(string) error
	ReadFile(string) ([]byte, error)
	ReadDir(string) ([]os.DirEntry, error)
	Stat(string) (fs.FileInfo, error)
	OpenFile(string, int, fs.FileMode) (*os.File, error)
	MkdirAll(string, fs.FileMode) error
	CreateTemp(string, string) (*os.File, error)
	Rename(string, string) error
	Remove(string) error
	PrepareCommand(*exec.Cmd) error
	Close() error
}
