package common

import (
	"io/fs"
	"os"
	"os/exec"
)

// Sandbox owns the Agent's file and subprocess boundary. Operations enforce the
// current owner policy; callers must not resolve a name then open the old name.
type Sandbox interface {
	// Agent returns the required policy owner.
	Agent() Agent
	// Resolve canonicalizes a name, rejecting paths outside the current root.
	Resolve(string) (string, error)
	// ReadFile reads only within the current boundary.
	ReadFile(string) ([]byte, error)
	// ReadDir lists a confined directory in name order.
	ReadDir(string) ([]os.DirEntry, error)
	// OpenFile opens a confined file with the caller's creation policy.
	OpenFile(string, int, fs.FileMode) (*os.File, error)
	// WriteFile replaces confined bytes without a check/open gap.
	WriteFile(string, []byte, fs.FileMode) error
	// Rename atomically moves two names within the boundary.
	Rename(string, string) error
	// Remove deletes one confined name.
	Remove(string) error
	// CreateTemp creates a private temporary file within the boundary.
	CreateTemp(string) (*os.File, error)
	// MkdirAll creates confined parent directories.
	MkdirAll(string, fs.FileMode) error
	// Stat reports confined file metadata.
	Stat(string) (fs.FileInfo, error)
	// WalkDir visits a confined subtree, preserving caller-relative names.
	WalkDir(string, fs.WalkDirFunc) error
	// Command prepares a shell with confined descendants and a clean environment.
	Command(string, string) (*exec.Cmd, error)
}
