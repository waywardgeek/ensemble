// Package auth owns issued-client identity and account grants. Nothing here is
// History, a prompt fragment, or a model-visible tool.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"ensemble/internal/common"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type clientIdentity struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"client_id"`
	HostID   string `json:"host_id"`
}
type grant struct {
	Subject    string    `json:"subject"`
	Account    string    `json:"account"`
	Issuer     string    `json:"issuer"`
	ClientID   string    `json:"client_id"`
	Access     string    `json:"access_token"`
	Refresh    string    `json:"refresh_token"`
	Expires    time.Time `json:"expires_at"`
	Eligible   bool      `json:"eligible"`
	Scopes     []string  `json:"scopes"`
	Generation uint64    `json:"generation"`
	Reason     string    `json:"reason,omitempty"`
}
type store struct {
	Client   *clientIdentity  `json:"client,omitempty"`
	Selected string           `json:"selected,omitempty"`
	Mode     common.AuthMode  `json:"billing_mode,omitempty"`
	Accounts map[string]grant `json:"accounts"`
}

func profileID(issuer, client, subject string) string {
	encoded, _ := json.Marshal([]string{issuer, client, subject})
	h := sha256.Sum256(encoded)
	return hex.EncodeToString(h[:])
}
func privateInfo(path string, dir bool) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 || fi.IsDir() != dir || (!dir && !fi.Mode().IsRegular()) {
		return errors.New("auth path must be a regular private file or directory")
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok || int(st.Uid) != os.Geteuid() || fi.Mode().Perm()&0077 != 0 {
		return errors.New("auth storage must be owned by this user and private")
	}
	return nil
}
func (s *Service) directory() (string, error) {
	dir := s.options.Directory
	if dir == "" {
		home, err := os.UserConfigDir()
		if err != nil {
			return "", errors.New("auth home unavailable")
		}
		dir = filepath.Join(home, "ensemble")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	// Refuse symlink traversal. Intermediate system-owned directories need not be
	// private; the final owner-only directory contains all sensitive files.
	for p := dir; p != filepath.Dir(p); p = filepath.Dir(p) {
		fi, err := os.Lstat(p)
		if err == nil && (fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir()) {
			return "", errors.New("auth directory traverses a non-directory or symlink")
		}
		if err != nil && !os.IsNotExist(err) {
			return "", errors.New("auth directory unavailable")
		}
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", errors.New("cannot create auth directory")
	}
	if err := privateInfo(dir, true); err != nil {
		return "", err
	}
	return dir, nil
}

// Every writer, including login and logout, uses this same OS lock. The kernel
// releases it on process death; no stale lock-file lease is guessed from time.
func (s *Service) locked(ctx context.Context, fn func(*store) error) error {
	dir, err := s.directory()
	if err != nil {
		return err
	}
	fd, err := syscall.Open(filepath.Join(dir, "writer.lock"), syscall.O_CREAT|syscall.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return errors.New("cannot open auth writer lock")
	}
	f := os.NewFile(uintptr(fd), "auth writer lock")
	defer f.Close()
	if err := privateInfo(filepath.Join(dir, "writer.lock"), false); err != nil {
		return err
	}
	for {
		err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			return errors.New("cannot lock auth storage")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	defer syscall.Flock(fd, syscall.LOCK_UN)
	if err := ctx.Err(); err != nil {
		return err
	}
	st := store{Accounts: map[string]grant{}}
	path := filepath.Join(dir, "profiles.json")
	if err := privateInfo(path, false); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return errors.New("cannot read auth profiles")
		}
		b, err := io.ReadAll(io.LimitReader(file, 2<<20))
		file.Close()
		if err != nil || json.Unmarshal(b, &st) != nil || st.Accounts == nil {
			return errors.New("invalid auth profiles")
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	before, _ := json.Marshal(st)
	actionErr := fn(&st)
	after, err := json.Marshal(st)
	if err != nil {
		return errors.New("cannot encode auth profiles")
	}
	if string(before) != string(after) {
		// Persist invalidation even when the operation returns a bounded error.
		if err = atomicPrivate(dir, path, after); err != nil {
			return err
		}
	}
	return actionErr
}
func atomicPrivate(dir, path string, b []byte) error {
	if err := privateInfo(path, false); err != nil && !os.IsNotExist(err) {
		return err
	}
	f, err := os.CreateTemp(dir, ".profiles-*")
	if err != nil {
		return errors.New("cannot stage auth profiles")
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(b)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return errors.New("cannot write auth profiles")
	}
	if err = os.Rename(name, path); err != nil {
		return errors.New("cannot replace auth profiles")
	}
	d, err := os.Open(dir)
	if err != nil {
		return errors.New("cannot sync auth directory")
	}
	defer d.Close()
	if err = d.Sync(); err != nil {
		return errors.New("cannot sync auth directory")
	}
	return nil
}

// save is used only inside locked, when host/registration must survive a failed flow.
func (s *Service) save(st *store) error {
	dir, err := s.directory()
	if err != nil {
		return err
	}
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return atomicPrivate(dir, filepath.Join(dir, "profiles.json"), b)
}
func profileKey(c *clientIdentity, sub string) string { return profileID(c.Issuer, c.ClientID, sub) }
