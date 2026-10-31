package gui_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestExternalReuse builds outside both modules. The only replacements are the
// two public module locations; private imports therefore cannot hide in-tree.
func TestExternalReuse(t *testing.T) {
	guiDir, _ := os.Getwd()
	coreDir := filepath.Dir(guiDir)
	dir := t.TempDir()
	module := fmt.Sprintf("module outside.example/reuse\n\ngo 1.25.0\nrequire (\n ensemble v0.0.0\n ensemble-gui-stub v0.0.0\n)\nreplace ensemble => %s\nreplace ensemble-gui-stub => %s\n", coreDir, guiDir)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("testdata/reuse/reuse_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "reuse_test.go"), data, 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOPROXY=off", "GOWORK=off", "GUI_SOURCE="+guiDir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("external %v: %v\n%s", args, err, out)
		}
		return string(out)
	}
	run("mod", "tidy")
	t.Log(run("test", "-race", "-v", "-count=1", "-timeout=45s", "."))
	// The headless consumer actually reaches MCP through its owning Agent.
	if err := os.Mkdir(filepath.Join(dir, "headless"), 0700); err != nil {
		t.Fatal(err)
	}
	headless := `package main
import ("io"; "ensemble/ensemble")
func main() { root := ensemble.New(io.Discard); a, err := root.NewAgent(ensemble.Config{Vendor:ensemble.OpenAI, Model:"gpt-4o-mini"}); if err != nil { panic(err) }; defer a.Shutdown(); _ = a.MCP().Tools() }
`
	if err := os.WriteFile(filepath.Join(dir, "headless/main.go"), []byte(headless), 0600); err != nil {
		t.Fatal(err)
	}
	run("build", "-o", filepath.Join(dir, "headless-probe"), "./headless")
	cmd := exec.Command(filepath.Join(dir, "headless-probe"))
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("headless runtime: %v %s", err, out)
	}
	for _, target := range []string{"./headless", "ensemble", "ensemble-gui-stub", "ensemble-gui-stub/cmd"} {
		deps := run("list", "-mod=mod", "-deps", target)
		gui := strings.Contains(deps, "ensemble-gui-stub\n")
		ws := strings.Contains(deps, "github.com/gorilla/websocket\n")
		want := strings.HasPrefix(target, "ensemble-gui-stub")
		if gui != want || ws != want {
			t.Fatalf("dependency control %s: GUI=%v WS=%v want=%v", target, gui, ws, want)
		}
	}
}
