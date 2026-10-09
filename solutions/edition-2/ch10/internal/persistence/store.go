package persistence

import (
	"errors"
	"example.com/ensemble/internal/common"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	parent     common.SessionAgent
	path       string
	lock       *os.File
	mu         sync.Mutex
	busy       bool
	closed     bool
	checkpoint *uint64
	io         common.CheckpointIO
}
type WriteResult = common.CheckpointResult
type checkpointWorker struct {
	parent   common.SessionStore
	captured common.Checkpoint
	save     bool
	done     chan WriteResult
}

func (s *Store) Agent() common.SessionAgent { return s.parent }
func (s *Store) failure(code, detail string) error {
	s.parent.Ensemble().Logf("session store: %s", detail)
	return &common.SessionError{Code: code, Detail: detail}
}

// Resolve follows the existing directory prefix only. Missing suffixes are
// appended after canonicalization; callers reserve the result under a root lock.
func Resolve(parent common.SessionAgent, directory string) (string, error) {
	if directory == "" {
		return "", &common.SessionError{Code: "session_conflict", Detail: "DataDir is required"}
	}
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(parent.Workspace(), directory)
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return "", &common.SessionError{Code: "session_io", Detail: "cannot resolve directory"}
	}
	suffix := []string{}
	prefix := absolute
	for {
		_, err = os.Lstat(prefix)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", &common.SessionError{Code: "session_io", Detail: "cannot inspect directory"}
		}
		next := filepath.Dir(prefix)
		if next == prefix {
			return "", &common.SessionError{Code: "session_io", Detail: "no directory ancestor"}
		}
		suffix = append(suffix, filepath.Base(prefix))
		prefix = next
	}
	prefix, err = filepath.EvalSymlinks(prefix)
	if err != nil {
		return "", &common.SessionError{Code: "session_io", Detail: "cannot canonicalize directory"}
	}
	for i := len(suffix) - 1; i >= 0; i-- {
		prefix = filepath.Join(prefix, suffix[i])
	}
	return prefix, nil
}
func Open(parent common.SessionAgent, path string, shared bool) (*Store, error) {
	s := &Store{parent: parent, path: path}
	s.io = &diskIO{parent: s}
	if !supported() {
		return nil, s.failure("session_unsupported", "local session locks require macOS or Linux")
	}
	if !shared {
		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, s.failure("session_io", "cannot create session directory")
		}
	}
	for _, name := range []string{"owner.lock", "events.log", "checkpoint.json", "origin.json"} {
		if err := s.leaf(name); err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(filepath.Join(path, "owner.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, s.failure("session_io", "cannot open owner lock")
	}
	// Go opens descriptors close-on-exec; explicitly retain that property in the
	// platform adapter. The lock inode is never unlinked, including on failure.
	if err = lockFile(f, shared); err != nil {
		f.Close()
		if inUse(err) {
			return nil, s.failure("session_in_use", "session has an active owner")
		}
		return nil, s.failure("session_io", "cannot lock session")
	}
	s.lock = f
	return s, nil
}
func (s *Store) leaf(name string) error {
	st, err := os.Lstat(filepath.Join(s.path, name))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return s.failure("session_io", "cannot inspect store leaf")
	}
	if !st.Mode().IsRegular() {
		return s.failure("session_corrupt", "store leaf must be a regular non-symlink file")
	}
	return nil
}
func (s *Store) Path() string { return s.path }
func (s *Store) Entries() ([]string, error) {
	entries, err := os.ReadDir(s.path)
	if err != nil {
		return nil, s.failure("session_io", "cannot list store")
	}
	out := []string{}
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out, nil
}
func (s *Store) Read(name string) ([]byte, bool, error) {
	if err := s.leaf(name); err != nil {
		return nil, false, err
	}
	f, err := os.Open(filepath.Join(s.path, name))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, s.failure("session_io", "cannot read store file")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, FileLimit+1))
	if err != nil {
		return nil, true, s.failure("session_io", "cannot read complete store file")
	}
	if len(raw) > FileLimit {
		return nil, true, s.failure("session_corrupt", "store file exceeds 512 MiB")
	}
	return raw, true, nil
}
func (s *Store) WriteOrigin(raw []byte) error {
	f, err := os.OpenFile(filepath.Join(s.path, "origin.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return s.failure("session_io", "cannot create immutable origin")
	}
	n, err := f.Write(raw)
	if err == nil && n != len(raw) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return s.failure("session_io", "cannot write immutable origin")
	}
	return nil
}
func (s *Store) Begin(cp common.Checkpoint, save bool) (<-chan WriteResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, s.failure("session_conflict", "store is closed")
	}
	if s.busy {
		return nil, s.failure("session_busy", "checkpoint worker is active")
	}
	s.busy = true
	w := &checkpointWorker{parent: s, captured: cp, save: save, done: make(chan WriteResult, 1)}
	go w.run()
	return w.done, nil
}
func (w *checkpointWorker) run() {
	s := w.parent
	raw, err := s.Agent().Codec().Encode(w.captured)
	saved := false
	if err == nil && w.save {
		err = s.ReplaceCheckpoint(raw)
		saved = err == nil
	}
	w.done <- WriteResult{Export: common.CheckpointExport{AsOf: w.captured.AsOf, Bytes: raw}, Saved: saved, Error: err}
	close(w.done)
}

// Applied is called at Actor's ordered boundary, after the worker has finished.
func (s *Store) Applied(result WriteResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if result.Saved {
		seq := result.Export.AsOf
		s.checkpoint = &seq
	}
	s.busy = false
}
func (s *Store) CheckpointSequence() *uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.checkpoint == nil {
		return nil
	}
	n := *s.checkpoint
	return &n
}
func (s *Store) LoadedCheckpoint(seq uint64) { s.mu.Lock(); defer s.mu.Unlock(); s.checkpoint = &seq }
func (s *Store) ReplaceCheckpoint(raw []byte) error {
	if err := s.leaf("checkpoint.json"); err != nil {
		return err
	}
	f, err := s.io.CreateTemp(s.path)
	if err != nil {
		return s.failure("session_io", "cannot create checkpoint temporary")
	}
	name := f.Name()
	defer s.io.Remove(name)
	n, err := f.Write(raw)
	if err == nil && n != len(raw) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = s.io.Rename(name, filepath.Join(s.path, "checkpoint.json"))
	}
	if err != nil {
		return s.failure("session_io", "checkpoint replacement did not commit")
	}
	return nil
}
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	if s.busy {
		return s.failure("session_busy", "checkpoint worker must be joined before store close")
	}
	s.closed = true
	if s.lock != nil {
		err := s.lock.Close()
		s.lock = nil
		if err != nil && !errors.Is(err, os.ErrClosed) {
			return s.failure("session_io", "cannot close session lock")
		}
	}
	return nil
}
