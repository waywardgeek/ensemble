package ensemble_test

import (
	"ensemble"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestCh19R7StartupFilePaths(t *testing.T) {
	caller := t.TempDir()
	base := filepath.Join(caller, "work")
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	t.Chdir(caller)
	raw := []byte(`{"safe_mode":true}`)
	for name, data := range map[string][]byte{"settings.json": raw, "malformed.json": []byte("{]")} {
		if err := os.WriteFile(filepath.Join(base, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("settings.json", filepath.Join(base, "inside.json")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte("outside fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(base, "outside.json")); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	cfg := ensemble.Config{Model: "offline", BaseURL: "http://127.0.0.1:1", BaseDir: base, SandboxRoot: caller, DataDir: "data"}
	for _, name := range []string{"settings.json", "inside.json", "malformed.json", "missing.json"} {
		file, err := app.ReadStartupFile(cfg, name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if file.Path != filepath.Join(base, name) || file.DataDir != filepath.Join(base, "data") || file.SavePath != "" {
			t.Errorf("%s: selected paths %+v", name, file)
		}
		if file.Exists != (name != "missing.json") {
			t.Errorf("%s: existence %t", name, file.Exists)
		}
		want := string(raw)
		if name == "malformed.json" {
			want = "{]"
		}
		if name == "missing.json" {
			want = ""
		}
		if string(file.Data) != want {
			t.Errorf("%s: bytes %q want %q", name, file.Data, want)
		}
	}
	for _, path := range []string{"outside.json", outside, filepath.Join(filepath.Dir(outside), "missing.json")} {
		file, err := app.ReadStartupFile(cfg, path)
		if err == nil || errors.Is(err, os.ErrNotExist) || len(file.Data) != 0 {
			t.Errorf("refused path became data/default: %+v %v", file, err)
		}
	}
	explicit := cfg
	explicit.SavePath = "caller.save"
	file, err := app.ReadStartupFile(explicit, "settings.json")
	if err != nil {
		t.Fatal(err)
	}
	if file.SavePath != filepath.Join(caller, "caller.save") {
		t.Fatalf("explicit SavePath moved: %s", file.SavePath)
	}
	empty := cfg
	empty.DataDir = ""
	file, err = app.ReadStartupFile(empty, "settings.json")
	if err != nil {
		t.Fatal(err)
	}
	if file.DataDir != filepath.Join(caller, ".ensemble") || file.SavePath != filepath.Join(file.DataDir, "save.json") {
		t.Fatalf("empty-DataDir defaults changed: %+v", file)
	}
	if len(app.Agents()) != 0 {
		t.Fatal("startup reads installed an Agent")
	}
	file, err = app.ReadStartupFile(cfg, "settings.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg.DataDir, cfg.SavePath = file.DataDir, file.SavePath
	a, err := app.NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID() != "agent-1" || a.Config().SavePath != filepath.Join(base, "data", "agent-1", "save.json") {
		t.Fatalf("ordinary construction defaults changed: id=%s save=%s", a.ID(), a.Config().SavePath)
	}
}

func TestCh19R7StartupFileClosesSandbox(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("local descriptor-name check uses Darwin /dev/fd")
	}
	dir := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file"), []byte("local"), 0600); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	cfg := ensemble.Config{Model: "offline", BaseURL: "http://127.0.0.1:1", BaseDir: dir, SandboxRoot: dir, DataDir: filepath.Join(dir, "data")}
	if _, err := app.ReadStartupFile(cfg, "file"); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	previous := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previous)
	count := func() int {
		t.Helper()
		f, err := os.Open("/dev/fd")
		if err != nil {
			t.Fatal(err)
		}
		names, err := f.Readdirnames(-1)
		closeErr := f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		return len(names)
	}
	before := count()
	for i := 0; i < 10; i++ {
		for _, name := range []string{"file", "missing", outside} {
			_, err := app.ReadStartupFile(cfg, name)
			if (name == outside) != (err != nil) {
				t.Fatalf("read %s: %v", name, err)
			}
		}
		invalid := cfg
		invalid.DataDir = outside
		if _, err := app.ReadStartupFile(invalid, "file"); err == nil {
			t.Fatal("invalid construction accepted")
		}
	}
	if after := count(); after != before {
		t.Fatalf("startup reads retained descriptors: before=%d after=%d", before, after)
	}
	if len(app.Agents()) != 0 {
		t.Fatal("startup reads installed an Agent")
	}
}
