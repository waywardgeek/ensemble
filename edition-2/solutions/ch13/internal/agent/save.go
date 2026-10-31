package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

// The application chooses identity and storage; persistence follows the owning
// Agent instead of consulting the process working directory itself.
func (a *agent) savePath() string {
	cfg := a.Config()
	if cfg.SavePath != "" {
		return cfg.SavePath
	}
	return filepath.Join(cfg.DataDir, "save.json")
}

// Load restores before the actor receives work. A missing file is the only
// failure treated as a fresh start; callers must not save after any other error.
func (a *agent) Load() error {
	path := a.savePath()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err == nil {
		defer f.Close()
		err = a.history.Restore(f)
	}
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	return nil
}

// Save replaces the destination only after encoding and closing a complete
// temporary file. Call after Shutdown so final managed-job facts are included.
func (a *agent) Save() error {
	path := a.savePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".save-*")
	if err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	defer os.Remove(f.Name())
	err = a.history.Save(f)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}
