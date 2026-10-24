package ensemble_test

import (
	"ensemble"
	"os"
	"path/filepath"
	"testing"
)

func TestCh19R5ExplicitRelativeSavePath(t *testing.T) {
	for _, confined := range []bool{false, true} {
		t.Run(map[bool]string{false: "unconfined", true: "confined"}[confined], func(t *testing.T) {
			caller := t.TempDir()
			t.Chdir(caller)
			base := filepath.Join(caller, "worktree")
			if err := os.Mkdir(base, 0700); err != nil {
				t.Fatal(err)
			}
			cfg := ensemble.Config{Model: "fixture", BaseDir: base, DataDir: "cr", SavePath: "selected.save", Tools: []string{}}
			if confined {
				cfg.SandboxRoot = caller
			}
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(cfg)
			if err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(caller, "selected.save")
			if a.Config().SavePath != want {
				t.Fatalf("explicit path relocated: %q want %q", a.Config().SavePath, want)
			}
			if a.Config().DataDir != filepath.Join(base, "cr") {
				t.Fatalf("derived data path: %q", a.Config().DataDir)
			}
			if err = a.Save(); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(want); err != nil {
				t.Fatal(err)
			}
			if _, err = os.Stat(filepath.Join(base, "selected.save")); !os.IsNotExist(err) {
				t.Fatalf("BaseDir unexpectedly owns explicit save: %v", err)
			}
			if _, err = a.Log().LoadFile(want); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("confined-outside-refused", func(t *testing.T) {
		caller := t.TempDir()
		root := t.TempDir()
		t.Chdir(caller)
		app := ensemble.New(nil)
		defer app.Shutdown()
		if _, err := app.NewAgent(ensemble.Config{Model: "fixture", BaseDir: root, SandboxRoot: root, DataDir: "cr", SavePath: "outside.save", Tools: []string{}}); err == nil {
			t.Fatal("outside caller-selected save silently moved into sandbox")
		}
		if _, err := os.Stat(filepath.Join(root, "outside.save")); !os.IsNotExist(err) {
			t.Fatalf("relocated save created: %v", err)
		}
	})
}
