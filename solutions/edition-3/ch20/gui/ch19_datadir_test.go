package gui

import (
	"ensemble"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCh19R3BootstrapDataDirReuse(t *testing.T) {
	temp := t.TempDir()
	t.Setenv("TMPDIR", temp)
	base := t.TempDir()
	settings := filepath.Join(base, "settings.json")
	app := ensemble.New(nil)
	defer app.Shutdown()
	server, a, err := NewServerFromConfig(app, ensemble.Config{Model: "offline", BaseDir: base, SafeMode: true}, "web/gui", settings)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	entries, err := os.ReadDir(temp)
	if err != nil {
		t.Fatal(err)
	}
	var allocated []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "ensemble-agent-") {
			allocated = append(allocated, filepath.Join(temp, entry.Name()))
		}
	}
	if len(allocated) != 1 {
		t.Errorf("bootstrap/real Agent allocated %d data roots: %v", len(allocated), allocated)
	}
	cfg := a.Config()
	if len(allocated) == 1 && cfg.DataDir != allocated[0] {
		t.Errorf("real Agent did not retain selected normalized DataDir: %s", cfg.DataDir)
	}
	if cfg.SavePath != filepath.Join(cfg.DataDir, "save.json") {
		t.Errorf("empty-DataDir SavePath default changed: %s", cfg.SavePath)
	}
}
