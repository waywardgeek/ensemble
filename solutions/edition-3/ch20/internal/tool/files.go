package tool

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Errors from an opened os.File carry its host name. Reattribute only these
// handle-operation errors to the requested file, not unrelated output errors.
func fileError(path string, err error) error {
	var e *fs.PathError
	if errors.As(err, &e) {
		return &fs.PathError{Op: e.Op, Path: path, Err: e.Err}
	}
	return err
}

func (r *registry) writeBytes(path string, data []byte) error {
	f, err := r.parent.Sandbox().OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err != nil {
		return fileError(path, err)
	}
	return fileError(path, closeErr)
}

// Walk through the rooted owner at every directory, never via filepath.WalkDir's
// host opens. Symlinks encountered in a search remain un-followed.
func (r *registry) walk(path string, visit fs.WalkDirFunc) error {
	info, err := r.parent.Sandbox().Stat(path)
	if err != nil {
		return visit(path, nil, err)
	}
	var walk func(string, fs.DirEntry) error
	walk = func(path string, entry fs.DirEntry) error {
		if err := visit(path, entry, nil); err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		entries, err := r.parent.Sandbox().ReadDir(path)
		if err != nil {
			return err
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, child := range entries {
			if err := walk(filepath.Join(path, child.Name()), child); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(path, fs.FileInfoToDirEntry(info))
}
