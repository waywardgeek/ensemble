package ensemble_test

import (
	"ensemble"
	"ensemble/internal/common"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"
)

func TestCh19RootedStateAndMedia(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside")
	const original = "outside fixture bytes"
	if err := os.WriteFile(outside, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "gpt-5.4-mini", Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, APIKey: "dummy", BaseDir: root, SandboxRoot: root, SafeMode: true, DataDir: filepath.Join(root, "data"), SavePath: filepath.Join(root, "save.json")})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Save(); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(root, "image.png")
	if err = os.WriteFile(image, []byte("local image fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx := common.Context{Dialogue: []common.Entry{{Actor: common.Human, Parts: common.PartList{{Type: "image", MIME: "image/png", Ref: &common.Ref{Kind: common.RefPath, Locator: image}}}}}}
	if _, err = a.Engine().Render(ctx); err != nil {
		t.Fatalf("rooted media render: %v", err)
	}
	replace := func(path string) {
		t.Helper()
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, path); err != nil {
			t.Fatal(err)
		}
	}
	replace(image)
	if _, err = a.Engine().Render(ctx); err == nil {
		t.Fatal("media followed outside symlink")
	}
	replace(a.Config().SavePath + ".journal")
	if err = a.RecordFact(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: "local journal fixture"}}); err == nil {
		t.Fatal("journal accepted outside symlink")
	}
	if err = os.Remove(a.Config().SavePath + ".journal"); err != nil {
		t.Fatal(err)
	}
	replace(a.Config().SavePath)
	if err = a.Save(); err == nil {
		t.Fatal("save accepted outside primary")
	}
	replace(filepath.Join(a.Config().DataDir, "cache-current.txt.tmp"))
	a.Engine().Cache().Observe(a.Config(), []byte(`{"model":"gpt-5.4-mini","messages":[]}`))
	if a.Engine().Cache().Status().DiagnosticError == "" {
		t.Fatal("cache did not report refused outside temporary")
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != original {
		t.Fatalf("outside fixture changed: %q, %v", data, err)
	}
	if !a.Config().SafeMode || a.Config().EnableWebSearch {
		t.Fatal("state operations widened policy")
	}
}
func TestCh19FailedConstructionClosesSandbox(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("local descriptor check uses Darwin /dev/fd")
	}
	runtime.GC()
	previous := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(previous)
	count := func() int {
		t.Helper()
		directory, err := os.Open("/dev/fd")
		if err != nil {
			t.Fatal(err)
		}
		entries, err := directory.Readdirnames(-1)
		closeErr := directory.Close()
		if err != nil {
			t.Fatal(err)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		return len(entries)
	}
	root := t.TempDir()
	outside := t.TempDir()
	app := ensemble.New(nil)
	defer app.Shutdown()
	before := count()
	for i := 0; i < 12; i++ {
		for _, cfg := range []ensemble.Config{{Model: "offline", BaseDir: root, SandboxRoot: root, DataDir: outside}, {Model: "offline", BaseDir: root, SandboxRoot: root, Tools: []string{"undeclared-fixture-tool"}}} {
			if a, err := app.NewAgent(cfg); err == nil {
				t.Fatalf("invalid construction succeeded: %v", a)
			}
		}
	}
	if after := count(); after > before {
		t.Fatalf("failed constructors retained descriptors: before=%d after=%d", before, after)
	}
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: root, SandboxRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	if err = app.Shutdown(); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Sandbox().ReadDir(root); err == nil {
		t.Fatal("successful ownership did not close on shutdown")
	}
}
