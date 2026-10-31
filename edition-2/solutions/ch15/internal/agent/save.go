package agent

import "path/filepath"

// The application chooses identity; persistence follows the owning Agent.
func (a *agent) savePath() string {
	cfg := a.Config()
	if cfg.SavePath != "" {
		return cfg.SavePath
	}
	return filepath.Join(cfg.DataDir, "save.json")
}

// Load restores snapshot plus journal before admitting new work.
func (a *agent) Load() error { return a.history.Recover(a.savePath()) }

// Save checkpoints after Shutdown so final managed-job facts are included.
func (a *agent) Save() error { return a.history.Checkpoint(a.savePath()) }
