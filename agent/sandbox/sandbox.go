// Package sandbox confines file and command operations to a directory tree.
//
// It is public API, not an internal detail, because confinement is the
// wrong thing to make every application reimplement. An application built
// on the Ensemble framework gets the same boundary the agent's own tools
// use, with the same canonicalization and the same refusals, rather than a
// fifth hand-rolled copy that gets the symlink case wrong.
//
// # What this package does and does not protect
//
// The asset being protected is the agent's CONTEXT WINDOW, not the host
// binary. The binary is trusted: it holds API credentials in memory and
// puts them on outgoing requests, and nothing is gained by pretending
// otherwise. The model is what must never see a credential or a file it
// was not given, because context is the only surface prompt injection can
// reach, and anything in context can be re-encoded and exfiltrated later.
//
// There are two boundaries, because there are two kinds of operation, and
// one mechanism cannot cover both:
//
//   - IN-PROCESS file operations (read_file, write_file, ...) run inside
//     the unsandboxed host binary. The kernel cannot help: these are Go
//     function calls, not subprocesses. They are confined HERE, in
//     userspace, by resolving every path against the root before opening
//     anything.
//
//   - SUBPROCESSES (run_command) are confined by the KERNEL, via the
//     operating system's own sandbox. A userspace path check is worthless
//     against a shell, because the shell never calls our code. See
//     seatbelt.go.
//
// Neither half is redundant and neither half is sufficient.
package sandbox

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Agent is the owner of a Sandbox: the object that knows where the
// boundary is.
//
// A Sandbox holds a BACK-POINTER to its parent rather than a copy of the
// path, so there is exactly one source of truth. Copying the root into the
// child is the staple pattern: it works until the owner's value changes,
// and then the child is confidently enforcing a boundary that moved.
//
// The interface is declared here, and narrowly, for two reasons. It cannot
// be the agent's own concrete type, because this package is imported BY
// the agent package and the reverse edge would be an import cycle. And it
// must not be the framework's wide Agent interface, because an application
// on this framework has its own agent type and should not have to satisfy
// a dozen unrelated methods to get confined file access.
type Agent interface {
	// SandboxRoot reports the directory operations are confined to, or ""
	// for no confinement.
	SandboxRoot() string
}

// ErrNoSandbox is returned by every operation on a nil Sandbox or one with
// a nil parent.
//
// This is the fail-closed rule, and it covers WIRING mistakes: a Sandbox
// that was never given a parent must not behave like an unconfined one,
// because then forgetting to wire it produces an unconfined agent that
// looks correct in every test. A parent that deliberately reports "" is a
// different thing, and is honored as the explicit opt-out.
var ErrNoSandbox = errors.New("sandbox: no parent agent (use New or Unconfined)")

// Sandbox confines operations to the directory tree its parent names.
//
// The zero value is unusable on purpose; construct one with New,
// Rooted, or Unconfined.
type Sandbox struct {
	// agent is the back-pointer. Every resolution walks up to it, so the
	// boundary is whatever the parent currently says it is.
	agent Agent
}

// New returns a Sandbox that asks parent where the boundary is.
//
// It cannot fail. A nil parent is not rejected here but at use time,
// where every operation returns ErrNoSandbox: failing closed at the
// operation is strictly safer than failing at construction, because a
// construction error can be ignored by a caller that has nothing useful
// to do with it, and an ignored error on a security boundary produces an
// agent that looks wired and is not.
func New(parent Agent) *Sandbox {
	return &Sandbox{agent: parent}
}

// fixedRoot is an Agent whose boundary never moves. It backs Rooted and
// Unconfined, so those constructors produce a Sandbox of exactly the same
// shape as a real one rather than a special case threaded through every
// method.
type fixedRoot string

func (f fixedRoot) SandboxRoot() string { return string(f) }

// Rooted returns a Sandbox confined to a fixed directory, for tests and
// for standalone use where there is no agent to point back at.
func Rooted(root string) (*Sandbox, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("sandbox: root must not be empty (use Unconfined for no confinement)")
	}
	return &Sandbox{agent: fixedRoot(root)}, nil
}

// Unconfined returns a Sandbox that confines nothing.
//
// It exists so that "no confinement" is a decision with a name, visible at
// the call site and greppable across a codebase, rather than the
// accidental consequence of an empty string. It is the library-level
// equivalent of typing --yolo.
func Unconfined() *Sandbox {
	return &Sandbox{agent: fixedRoot("")}
}

// root asks the parent for the boundary and canonicalizes it.
//
// Canonicalizing on every call is deliberate. Caching would reintroduce
// precisely the staleness the back-pointer exists to eliminate, and
// because tools run on job goroutines a memo field would need a mutex to
// avoid a data race. One extra EvalSymlinks against an operation that is
// already doing disk I/O is not worth a cache and a lock.
//
// The canonicalization itself is not a nicety: on macOS /tmp is a symlink
// to /private/tmp and /var to /private/var, so a sandbox created from
// mktemp -d has a raw root that never matches the resolved form of any
// path inside it. Canonicalize both sides or compare neither.
func (s *Sandbox) root() (string, bool, error) {
	if s == nil || s.agent == nil {
		return "", false, ErrNoSandbox
	}
	raw := s.agent.SandboxRoot()
	if strings.TrimSpace(raw) == "" {
		return "", false, nil // explicitly unconfined
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", false, fmt.Errorf("sandbox: resolving root: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// A root that does not exist yet is a usable boundary: a caller
		// may name a workspace the agent is about to create.
		resolved = filepath.Clean(abs)
	}
	return resolved, true, nil
}

// Root reports the canonical boundary, or "" when unconfined.
func (s *Sandbox) Root() string {
	r, confined, err := s.root()
	if err != nil || !confined {
		return ""
	}
	return r
}

// IsConfined reports whether this Sandbox actually restricts anything.
func (s *Sandbox) IsConfined() bool {
	_, confined, err := s.root()
	return err == nil && confined
}

// Resolve returns an absolute, symlink-resolved path guaranteed to be
// inside the boundary, or an error naming the violation.
//
// Prefer the operation methods (ReadFile, Open, ...) to calling Resolve
// directly. The dangerous pattern this package exists to prevent is
// resolving a path and then passing the ORIGINAL string to the os package,
// which compiles, passes casual testing, and is completely unconfined.
// Resolve is exported for the cases the methods do not cover, such as
// choosing a working directory for a subprocess.
//
// The order of operations is the security property, and each step defeats
// a different escape:
//
//  1. Join a relative path to the root, so "notes.txt" means the sandbox.
//  2. Clean, collapsing ".." and defeating "../../etc/passwd".
//  3. EvalSymlinks, defeating a symlink planted by an earlier command.
//  4. Prefix-check the RESOLVED path against the CANONICAL root, which is
//     what makes steps 2 and 3 mean anything.
func (s *Sandbox) Resolve(path string) (string, error) {
	root, confined, err := s.root()
	if err != nil {
		return "", err
	}
	if !confined {
		return path, nil
	}
	if path == "" {
		return "", errors.New("sandbox: path is required")
	}

	// The caller's own string, kept for the error message. The resolved
	// path is an absolute host path, and handing that to a language model
	// describes the filesystem it is being confined away from.
	asked := path

	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path = filepath.Clean(path)

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		// The path may not exist yet: creating a file is legitimate, and
		// EvalSymlinks fails outright on a missing leaf. Resolve the
		// longest existing ancestor and check where the new file WOULD
		// land, which also catches an ancestor that is a symlink out.
		resolved, err = resolveExistingPrefix(path)
		if err != nil {
			return "", &EscapeError{Path: asked}
		}
	}
	if !withinRoot(root, resolved) {
		return "", &EscapeError{Path: asked}
	}
	return resolved, nil
}

// EscapeError reports an attempt to reach outside the sandbox.
//
// It deliberately carries only the path the caller ASKED for, never the
// resolved host path or the sandbox root. A refusal that echoes the host
// layout back to a language model teaches it the shape of the filesystem
// it is confined away from, which is a slow information leak paid out one
// error message at a time.
type EscapeError struct {
	Path string
}

func (e *EscapeError) Error() string {
	return "path outside sandbox: " + e.Path
}

// withinRoot reports whether path is root itself or lives underneath it.
//
// The separator is the point. A plain strings.HasPrefix(path, root)
// reports that "/tmp/work-evil" is inside "/tmp/work", because one string
// really is a prefix of the other. Comparing against root+separator is
// what makes this a path test rather than a string test.
func withinRoot(root, path string) bool {
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(filepath.Separator))
}

// Within reports whether path lies inside root, canonicalizing both sides
// first. It exists for callers deciding whether one boundary nests inside
// another -- a sub-agent's sandbox against its parent's, say -- rather than
// deciding whether one file may be opened.
//
// Both sides must be canonicalized, not just the path. On macOS /tmp and
// /var are symlinks, so a root taken from mktemp or t.TempDir() is already
// a symlink and a path resolved under it will never share its prefix. A
// check that resolves only the path reports "outside" for every legitimate
// file in the sandbox.
//
// An empty root means unconfined, which contains everything. An empty path
// means an unconfined child, which is not inside a confined parent.
func Within(root, path string) (bool, error) {
	if root == "" {
		return true, nil
	}
	if path == "" {
		return false, nil
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	canonRoot, err := resolveExistingPrefix(absRoot)
	if err != nil {
		return false, err
	}
	canonPath, err := resolveExistingPrefix(absPath)
	if err != nil {
		return false, err
	}
	return withinRoot(canonRoot, canonPath), nil
}

// resolveExistingPrefix resolves the longest existing ancestor of path and
// rejoins the components that do not exist yet.
//
// Without this, creating a file is either unchecked or impossible: the
// target does not exist, so EvalSymlinks fails, and an implementation must
// choose between accepting the path blind (the escape) and refusing every
// write (a sandbox nobody can use). Walking up to the nearest real
// directory and resolving THAT catches a symlinked ancestor pointing out
// of the sandbox.
func resolveExistingPrefix(path string) (string, error) {
	var missing []string
	cur := path
	for {
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("no existing ancestor for %s", path)
		}
		missing = append([]string{filepath.Base(cur)}, missing...)
		if resolved, err := filepath.EvalSymlinks(parent); err == nil {
			return filepath.Join(append([]string{resolved}, missing...)...), nil
		}
		cur = parent
	}
}

// --- confined file operations ---------------------------------------------
//
// These mirror the os package. Each resolves its path internally and never
// returns a raw path for the caller to misuse. Together they are the whole
// reason this package is public: an application on the framework should
// reach for sb.ReadFile, not for os.ReadFile plus a path check it wrote
// itself.

// ReadFile reads a file inside the sandbox.
func (s *Sandbox) ReadFile(name string) ([]byte, error) {
	p, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

// WriteFile writes a file inside the sandbox, creating parent directories.
func (s *Sandbox) WriteFile(name string, data []byte, perm os.FileMode) error {
	p, err := s.Resolve(name)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(p); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(p, data, perm)
}

// Append appends to a file inside the sandbox, creating it if needed.
func (s *Sandbox) Append(name string, data []byte, perm os.FileMode) error {
	p, err := s.Resolve(name)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// Open opens a file inside the sandbox for reading.
func (s *Sandbox) Open(name string) (*os.File, error) {
	p, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Open(p)
}

// Create creates or truncates a file inside the sandbox.
func (s *Sandbox) Create(name string) (*os.File, error) {
	p, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Create(p)
}

// Stat reports file metadata inside the sandbox.
func (s *Sandbox) Stat(name string) (os.FileInfo, error) {
	p, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(p)
}

// ReadDir lists a directory inside the sandbox.
func (s *Sandbox) ReadDir(name string) ([]os.DirEntry, error) {
	p, err := s.Resolve(name)
	if err != nil {
		return nil, err
	}
	return os.ReadDir(p)
}

// MkdirAll creates a directory tree inside the sandbox.
func (s *Sandbox) MkdirAll(name string, perm os.FileMode) error {
	p, err := s.Resolve(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(p, perm)
}

// Remove deletes a file inside the sandbox.
func (s *Sandbox) Remove(name string) error {
	p, err := s.Resolve(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

// WalkDir walks a tree inside the sandbox.
//
// Entries that resolve outside the sandbox are skipped rather than
// reported. A walk that followed a symlink out would hand the caller
// filenames from the host filesystem even if it never opened them, and a
// directory listing is itself information.
func (s *Sandbox) WalkDir(name string, fn fs.WalkDirFunc) error {
	start, err := s.Resolve(name)
	if err != nil {
		return err
	}
	root, confined, err := s.root()
	if err != nil {
		return err
	}
	return filepath.WalkDir(start, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fn(path, d, err)
		}
		if confined {
			real, rerr := filepath.EvalSymlinks(path)
			if rerr != nil || !withinRoot(root, real) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		return fn(path, d, nil)
	})
}
