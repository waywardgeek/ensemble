package persistence

import (
	"example.com/ensemble/internal/common"
	"os"
)

type diskIO struct{ parent common.SessionStore }

func (d *diskIO) Store() common.SessionStore { return d.parent }
func (d *diskIO) CreateTemp(dir string) (common.CheckpointFile, error) {
	f, err := os.CreateTemp(dir, ".checkpoint-*")
	if err != nil {
		return nil, err
	}
	return &diskFile{parent: d, File: f}, nil
}
func (d *diskIO) Rename(from, to string) error { return os.Rename(from, to) }
func (d *diskIO) Remove(path string) error     { return os.Remove(path) }

type diskFile struct {
	parent common.CheckpointIO
	*os.File
}

func (f *diskFile) IO() common.CheckpointIO { return f.parent }
