// Package sandbox enforces the boundary owned by an Agent. It imports only the
// common owner contract, never the file tools or process supervision spokes.
package sandbox

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"ensemble/internal/common"
)

type sandbox struct{ parent common.Agent }

// New requires the policy owner; an empty owner root explicitly means full access.
func New(parent common.Agent) common.Sandbox {
	if parent == nil {
		panic("Sandbox requires Agent")
	}
	return &sandbox{parent: parent}
}

// Agent returns the source of truth for the current confinement policy.
func (s *sandbox) Agent() common.Agent { return s.parent }

// canonical resolves existing ancestors too, so creation beneath a symlink is
// checked by the same rule as reading an existing file. Only ENOENT permits
// walking upward: permission errors and symlink loops are not nonexistent paths.
func canonical(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolved, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	parent := filepath.Dir(path)
	if parent == path {
		return "", err
	}
	resolved, err = canonical(parent)
	return filepath.Join(resolved, filepath.Base(path)), err
}

// Within checks physical containment, including not-yet-created descendants.
// An empty candidate never widens a confined parent into full access.
func Within(root, path string) bool {
	if root == "" || path == "" {
		return false
	}
	r, err := canonical(root)
	if err != nil {
		return false
	}
	p, err := canonical(path)
	return err == nil && inside(r, p)
}
func inside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Refusals retain only the caller's own spelling. OS errors otherwise disclose
// the canonical host root through PathError, including for ordinary ENOENT.
func refusal(path string) error { return fmt.Errorf("sandbox refused %q", path) }
func safeError(path string, err error) error {
	if err == nil {
		return nil
	}
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return &fs.PathError{Op: pe.Op, Path: path, Err: pe.Err}
	}
	return refusal(path)
}
func (s *sandbox) root() (string, error) {
	if s == nil || s.parent == nil {
		return "", errors.New("sandbox has no owner")
	}
	root := s.parent.Spec().SandboxRoot
	if root == "" {
		return "", nil
	}
	return filepath.EvalSymlinks(root)
}

// Resolve checks canonical root and candidate together, including /tmp aliases.
// Actual operations additionally use os.Root so a symlink swap after this check
// cannot turn a checked path into an unconfined open.
func (s *sandbox) Resolve(path string) (string, error) {
	root, err := s.root()
	if err != nil {
		return "", refusal(path)
	}
	if root == "" {
		return path, nil
	}
	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(root, candidate)
	}
	resolved, err := canonical(candidate)
	if err != nil || !inside(root, resolved) {
		return "", refusal(path)
	}
	return resolved, nil
}
func (s *sandbox) openRoot(path string) (*os.Root, string, error) {
	root, err := s.root()
	if err != nil {
		return nil, "", refusal(path)
	}
	if root == "" {
		return nil, path, nil
	}
	resolved, err := s.Resolve(path)
	if err != nil {
		return nil, "", err
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil {
		return nil, "", refusal(path)
	}
	r, err := os.OpenRoot(root)
	return r, rel, safeError(path, err)
}

// OpenFile binds a checked relative name to an open root for race-safe access.
func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (*os.File, error) {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return os.OpenFile(rel, flags, mode)
	}
	defer root.Close()
	f, err := root.OpenFile(rel, flags, mode)
	return f, safeError(path, err)
}

// ReadFile keeps resolution and opening together instead of trusting callers.
func (s *sandbox) ReadFile(path string) ([]byte, error) {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return os.ReadFile(rel)
	}
	defer root.Close()
	data, err := root.ReadFile(rel)
	return data, safeError(path, err)
}

// ReadDir uses the same confined filesystem as file reads.
func (s *sandbox) ReadDir(path string) ([]os.DirEntry, error) {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return os.ReadDir(rel)
	}
	defer root.Close()
	entries, err := fs.ReadDir(root.FS(), filepath.ToSlash(rel))
	return entries, safeError(path, err)
}

// MkdirAll permits ordinary nested creation without following an escaping link.
func (s *sandbox) MkdirAll(path string, mode fs.FileMode) error {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return err
	}
	if root == nil {
		return os.MkdirAll(rel, mode)
	}
	defer root.Close()
	return safeError(path, root.MkdirAll(rel, mode))
}

// Stat exposes metadata with caller-relative diagnostics.
func (s *sandbox) Stat(path string) (fs.FileInfo, error) {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return os.Stat(rel)
	}
	defer root.Close()
	info, err := root.Stat(rel)
	return info, safeError(path, err)
}

// WalkDir preserves the supplied spelling in search output while traversing the
// root-backed filesystem. Symlink directories are not followed by fs.WalkDir.
func (s *sandbox) WalkDir(path string, fn fs.WalkDirFunc) error {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return err
	}
	if root == nil {
		return filepath.WalkDir(path, fn)
	}
	defer root.Close()
	return fs.WalkDir(root.FS(), filepath.ToSlash(rel), func(name string, d fs.DirEntry, err error) error {
		suffix, _ := filepath.Rel(rel, name)
		asked := filepath.Join(path, suffix)
		return fn(asked, d, safeError(asked, err))
	})
}

// WriteFile shares OpenFile's race-safe path enforcement.
func (s *sandbox) WriteFile(path string, data []byte, mode fs.FileMode) error {
	f, err := s.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	return safeError(path, errors.Join(writeErr, f.Close()))
}

// Rename checks both names against one root before using root-relative rename.
func (s *sandbox) Rename(old, next string) error {
	root, oldRel, err := s.openRoot(old)
	if err != nil {
		return err
	}
	if root == nil {
		return os.Rename(old, next)
	}
	defer root.Close()
	resolved, err := s.Resolve(next)
	if err != nil {
		return err
	}
	nextRel, err := filepath.Rel(root.Name(), resolved)
	if err != nil {
		return refusal(next)
	}
	return safeError(next, root.Rename(oldRel, nextRel))
}

// Remove cannot traverse a symlink to delete an outside file.
func (s *sandbox) Remove(path string) error {
	root, rel, err := s.openRoot(path)
	if err != nil {
		return err
	}
	if root == nil {
		return os.Remove(path)
	}
	defer root.Close()
	return safeError(path, root.Remove(rel))
}

// CreateTemp keeps checkpoint installation within the same root filesystem.
// Random names plus exclusive creation preserve the existing snapshot protocol.
func (s *sandbox) CreateTemp(dir string) (*os.File, error) {
	return s.OpenFile(filepath.Join(dir, ".save-"+rand.Text()), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
}
