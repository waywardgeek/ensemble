// Package sandbox owns Agent file and process confinement, not conversation state.
package sandbox

import (
	"crypto/rand"
	"encoding/hex"
	"ensemble/internal/common"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type sandbox struct {
	parent   common.Agent
	root     string
	files    *os.Root
	once     sync.Once
	closeErr error
}

func New(parent common.Agent) (common.Sandbox, error) {
	s := &sandbox{parent: parent}
	if root := parent.Config().SandboxRoot; root != "" {
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		// Opening and canonicalizing happen once, before the Agent is published.
		abs, err = filepath.EvalSymlinks(abs)
		if err != nil {
			return nil, fmt.Errorf("sandbox root: %w", err)
		}
		s.files, err = os.OpenRoot(abs)
		if err != nil {
			return nil, fmt.Errorf("sandbox root: %w", err)
		}
		s.root = abs
	}
	return s, nil
}
func (s *sandbox) Parent() common.Agent { return s.parent }
func (s *sandbox) Root() string         { return s.root }
func (s *sandbox) Close() error {
	s.once.Do(func() {
		if s.files != nil {
			s.closeErr = s.files.Close()
		}
	})
	return s.closeErr
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// Keep filesystem causes useful without returning the canonical/host pathname
// embedded in an os.PathError. This is parameter attribution, not text scrubbing.
func pathError(op, supplied string, err error) error {
	if err == nil {
		return nil
	}
	var nested *fs.PathError
	for errors.As(err, &nested) {
		err = nested.Err
	}
	return &fs.PathError{Op: op, Path: supplied, Err: err}
}

func (s *sandbox) ValidatePath(path string) (resolved string, err error) {
	defer func() { err = pathError("resolve", path, err) }()
	candidate := path
	if candidate == "" {
		candidate = "."
	}
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(s.parent.Config().BaseDir, candidate)
	}
	p, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if s.root == "" {
		return filepath.Clean(p), nil
	}
	p, err = common.CanonicalPath(p)
	if err != nil {
		return "", err
	}
	if !within(s.root, p) {
		return "", fmt.Errorf("path resolves outside sandbox root")
	}
	return p, nil
}
func (s *sandbox) name(path string) (string, error) {
	p, err := s.ValidatePath(path)
	if err != nil {
		return "", err
	}
	if s.files == nil {
		return p, nil
	}
	return filepath.Rel(s.root, p)
}
func (s *sandbox) ReadFile(path string) (data []byte, err error) {
	defer func() { err = pathError("read", path, err) }()
	n, err := s.name(path)
	if err != nil {
		return nil, err
	}
	if s.files != nil {
		return s.files.ReadFile(n)
	}
	return os.ReadFile(n)
}
func (s *sandbox) ReadDir(path string) (entries []os.DirEntry, err error) {
	defer func() { err = pathError("readdir", path, err) }()
	f, err := s.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entries, err = f.ReadDir(-1)
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, err
}
func (s *sandbox) Stat(path string) (info fs.FileInfo, err error) {
	defer func() { err = pathError("stat", path, err) }()
	n, err := s.name(path)
	if err != nil {
		return nil, err
	}
	if s.files != nil {
		return s.files.Stat(n)
	}
	return os.Stat(n)
}
func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (file *os.File, err error) {
	defer func() { err = pathError("open", path, err) }()
	n, err := s.name(path)
	if err != nil {
		return nil, err
	}
	if s.files != nil {
		return s.files.OpenFile(n, flags, mode)
	}
	return os.OpenFile(n, flags, mode)
}
func (s *sandbox) MkdirAll(path string, mode fs.FileMode) (err error) {
	defer func() { err = pathError("mkdir", path, err) }()
	n, err := s.name(path)
	if err != nil {
		return err
	}
	if s.files != nil {
		return s.files.MkdirAll(n, mode)
	}
	return os.MkdirAll(n, mode)
}
func (s *sandbox) CreateTemp(dir, prefix string) (*os.File, error) {
	// Root has no CreateTemp. Use an exclusive rooted open, not os.CreateTemp
	// after a pathname check; intermediate directories may be replaced meanwhile.
	for i := 0; i < 10; i++ {
		var b [12]byte
		if _, err := rand.Read(b[:]); err != nil {
			return nil, err
		}
		f, err := s.OpenFile(filepath.Join(dir, prefix+hex.EncodeToString(b[:])), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if !os.IsExist(err) {
			return f, err
		}
	}
	return nil, fmt.Errorf("create temporary file: name collisions")
}
func (s *sandbox) Rename(old, new string) error {
	a, err := s.name(old)
	if err != nil {
		return err
	}
	b, err := s.name(new)
	if err != nil {
		return err
	}
	if s.files != nil {
		err = s.files.Rename(a, b)
	} else {
		err = os.Rename(a, b)
	}
	if err == nil {
		return nil
	}
	var link *os.LinkError
	if errors.As(err, &link) {
		err = link.Err
	}
	return &os.LinkError{Op: "rename", Old: old, New: new, Err: err}
}
func (s *sandbox) Remove(path string) (err error) {
	defer func() { err = pathError("remove", path, err) }()
	n, err := s.name(path)
	if err != nil {
		return err
	}
	if s.files != nil {
		return s.files.Remove(n)
	}
	return os.Remove(n)
}
func (s *sandbox) ValidateCommand(command string) error {
	if s.parent.Config().SafeMode {
		return fmt.Errorf("commands are disabled in safe mode")
	}
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}

// PrepareCommand is used before every shell/PTY start. The fixed executable
// and profile are owned by us, not shell text. Env is always explicit.
func (s *sandbox) PrepareCommand(cmd *exec.Cmd) error {
	// Authorization belongs to the caller: command tools call ValidateCommand,
	// while Skills authorize MCP servers independently. Both use this confinement.
	if cmd == nil || len(cmd.Args) == 0 {
		return fmt.Errorf("empty process command")
	}
	cmd.Env = common.ProcessEnvironment(s.parent)
	if s.root == "" {
		return nil
	}
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("command confinement is unavailable on %s", runtime.GOOS)
	}
	if _, err := os.Stat("/usr/bin/sandbox-exec"); err != nil {
		return fmt.Errorf("sandbox-exec unavailable: %w", err)
	}
	// The command owner retains supplied spelling until this final rooted check.
	dir, err := s.ValidatePath(cmd.Dir)
	if err != nil {
		return err
	}
	cmd.Dir = dir
	private := filepath.Join(s.root, ".ensemble-runtime")
	home, tmp := filepath.Join(private, "home"), filepath.Join(private, "tmp")
	for _, p := range []string{home, tmp} {
		if err := s.MkdirAll(p, 0700); err != nil {
			return err
		}
	}
	// Never read host Git configuration or reuse mutable includes/helpers.
	// Regenerate an identity-only file inside the same rooted filesystem owner.
	gitConfig := filepath.Join(private, "gitconfig")
	// Atomic replacement also keeps concurrent preparations from exposing a
	// truncated configuration to an already running confined Git process.
	configFile, err := s.CreateTemp(private, ".gitconfig-")
	if err != nil {
		return err
	}
	defer s.Remove(configFile.Name())
	_, writeErr := configFile.WriteString("[user]\n\tname = Ensemble\n\temail = ensemble@localhost\n")
	closeErr := configFile.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err = s.Rename(configFile.Name(), gitConfig); err != nil {
		return err
	}
	cmd.Env = append(cmd.Env, "HOME="+home, "TMPDIR="+tmp, "TMP="+tmp, "TEMP="+tmp,
		"GIT_CONFIG_GLOBAL="+gitConfig, "GIT_CONFIG_SYSTEM=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	cacheCmd := exec.Command("/usr/bin/getconf", "DARWIN_USER_TEMP_DIR")
	cacheCmd.Env = common.ProcessEnvironment(s.parent)
	cacheCmd.Dir = s.root
	cacheBytes, err := cacheCmd.Output()
	if err != nil {
		return fmt.Errorf("toolchain cache directory unavailable")
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(cacheBytes)))
	if err != nil {
		return err
	}
	profile := Profile(s.root, cache)
	cmd.Args = append([]string{"/usr/bin/sandbox-exec", "-p", profile, cmd.Path}, cmd.Args[1:]...)
	cmd.Path = "/usr/bin/sandbox-exec"
	return nil
}

// Profile grants system startup reads but never the entire per-user temp tree.
// SBPL sees physical paths; only immediate xcrun_db* cache files are writable.
func Profile(root string, caches ...string) string {
	q := strconv.Quote(root)
	// Traversal requires metadata on these literal ancestors, not host-wide access.
	var metadata strings.Builder
	seen := map[string]bool{}
	paths := []string{root, "/private/etc/ssl"}
	if len(caches) > 0 && caches[0] != "" {
		paths = append(paths, caches[0])
	}
	for _, path := range paths {
		for ancestor := filepath.Clean(path); ; ancestor = filepath.Dir(ancestor) {
			if !seen[ancestor] {
				fmt.Fprintf(&metadata, "(allow file-read-metadata (literal %s))\n", strconv.Quote(ancestor))
				seen[ancestor] = true
			}
			if ancestor == filepath.Dir(ancestor) {
				break
			}
		}
	}

	// Apple tool shims traverse the lexical /var alias and inspect the physical
	// developer_dir link into the already-readable /Library runtime tree.
	// Permit only that metadata lookup, not selection-directory contents.
	for _, path := range []string{"/var", "/private/var/select", "/private/var/select/developer_dir"} {
		if !seen[path] {
			fmt.Fprintf(&metadata, "(allow file-read-metadata (literal %s))\n", strconv.Quote(path))
			seen[path] = true
		}
	}

	cache := ""
	if len(caches) > 0 && caches[0] != "" {
		pattern := "^" + regexp.QuoteMeta(filepath.Join(caches[0], "xcrun_db")) + "[^/]*$"
		cache = "\n(allow file-read* file-write* (regex #" + strconv.Quote(pattern) + "))"
	}
	return `(version 1)
(allow default)
(deny file-read* file-write*)
(allow file-read* (literal "/"))
` + metadata.String() + `(allow file-read* (subpath "/usr") (subpath "/bin") (subpath "/System") (subpath "/Library") (subpath "/dev") (subpath "/private/etc/ssl"))
(allow file-read* file-write* (subpath ` + q + `))
(allow file-write* (literal "/dev/null") (literal "/dev/tty") (literal "/dev/stdout") (literal "/dev/stderr") (literal "/dev/ptmx") (regex #"^/dev/ttys[0-9]+$"))
(deny network*)` + cache
}
