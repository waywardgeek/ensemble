package persistence

// This review-only constructor wraps the actual exclusive origin descriptor.
// Mode is scoped to a deliberately named test directory; no global hook exists.
import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ch10OriginFile struct {
	parent *Store
	real   *os.File
	mode   string
}

func (f *ch10OriginFile) note(operation string, n int) {
	path := filepath.Join(filepath.Dir(f.parent.path), filepath.Base(f.parent.path)+".operations")
	out, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = fmt.Fprintf(out, "%s %d\n", operation, n); err != nil {
		panic(err)
	}
	if err = out.Close(); err != nil {
		panic(err)
	}
}
func ch10ReviewOriginOpen(s *Store, path string, flag int, mode os.FileMode) (*ch10OriginFile, error) {
	real, err := os.OpenFile(path, flag, mode)
	if err != nil {
		return nil, err
	}
	f := &ch10OriginFile{parent: s, real: real, mode: filepath.Base(s.path)}
	f.note("open", 0)
	return f, nil
}
func (f *ch10OriginFile) Write(raw []byte) (int, error) {
	if f.mode == "origin-short" || f.mode == "origin-write" {
		n, err := f.real.Write(raw[:len(raw)/2])
		f.note("write", n)
		if err == nil && f.mode == "origin-write" {
			err = errors.New("review origin write failure")
		}
		return n, err
	}
	n, err := f.real.Write(raw)
	f.note("write", n)
	return n, err
}
func (f *ch10OriginFile) Sync() error {
	f.note("sync", 0)
	err := f.real.Sync()
	if err == nil && f.mode == "origin-sync" {
		return errors.New("review origin sync failure")
	}
	return err
}
func (f *ch10OriginFile) Close() error {
	f.note("close", 0)
	err := f.real.Close()
	if err == nil && f.mode == "origin-close" {
		return errors.New("review origin close failure")
	}
	return err
}
