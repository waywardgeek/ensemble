package ensemble

import (
	"ensemble/internal/agent"
	"errors"
	"os"
	"path/filepath"
)

// StartupFile is a synchronous pre-construction read, not a live Agent or
// Sandbox. Missing files have Exists=false. Path is absolute and relative input
// follows BaseDir. DataDir and SavePath are the path choices to carry into the
// original Config before NewAgent; other Config fields remain caller-owned.
type StartupFile struct {
	Path     string
	Data     []byte
	Exists   bool
	DataDir  string
	SavePath string
}

// ReadStartupFile uses the ordinary construction path owner and closes its
// Sandbox before returning. It installs no Agent, History, capabilities or
// transport, and does not parse the bytes. Missing files are distinguished from
// read/confinement/close errors so applications can apply their own defaults.
func (e *Ensemble) ReadStartupFile(config Config, path string) (file StartupFile, err error) {
	if e == nil {
		return file, errors.New("startup read requires an Ensemble parent")
	}
	e.mu.Lock()
	closed := e.closed
	e.mu.Unlock()
	if closed {
		return file, errors.New("Ensemble is shut down")
	}
	bootstrap, err := agent.New(e, config)
	if err != nil {
		return file, err
	}
	defer func() { err = errors.Join(err, bootstrap.Sandbox().Close()) }()
	selected := bootstrap.Config()
	file.DataDir = selected.DataDir
	// An originally empty DataDir uses save.json, not the explicit-DataDir
	// per-Agent convention. Otherwise keep an absent SavePath absent for NewAgent.
	if config.DataDir == "" || config.SavePath != "" {
		file.SavePath = selected.SavePath
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(selected.BaseDir, path)
	}
	file.Path, err = filepath.Abs(path)
	if err != nil {
		return file, err
	}
	if _, err = bootstrap.Sandbox().ValidatePath(file.Path); err != nil {
		return file, err
	}
	file.Data, err = bootstrap.Sandbox().ReadFile(file.Path)
	if errors.Is(err, os.ErrNotExist) {
		return file, nil
	}
	file.Exists = err == nil
	return file, err
}
