package sandbox

import (
	"ensemble/internal/common"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type ch19Parent struct {
	common.Agent
	cfg common.Config
}

func (p *ch19Parent) Config() common.Config        { return p.cfg }
func (p *ch19Parent) Skills() common.SkillRegistry { return nil }
func TestCh19PreparationPolicy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", "/untrusted/bin")
	t.Setenv("HOME", "/untrusted/home")
	t.Setenv("TMPDIR", "/untrusted/tmp")
	t.Setenv("GIT_CONFIG_GLOBAL", "/untrusted/git")
	t.Setenv("GIT_CONFIG_SYSTEM", "/untrusted/system")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "0")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "credential.helper")
	t.Setenv("GIT_CONFIG_VALUE_0", "should-not-be-inherited")
	t.Setenv("XDG_CONFIG_HOME", "/untrusted/xdg")
	p := &ch19Parent{cfg: common.Config{BaseDir: dir, SandboxRoot: dir, SafeMode: true}}
	files, err := New(p)
	if err != nil {
		t.Fatal(err)
	}
	defer files.Close()
	gitPath := filepath.Join(dir, ".ensemble-runtime", "gitconfig")
	if err := os.MkdirAll(filepath.Dir(gitPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gitPath, []byte("[include]\n path = /untrusted/config\n[credential]\n helper = should-not-survive\n[url \"file:///untrusted/\"]\n insteadOf = fixture:\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if files.ValidateCommand("printf forbidden") == nil {
		t.Fatal("safe arbitrary command admitted")
	}
	cmd := exec.Command("/bin/echo", "authorized MCP fixture")
	cmd.Dir = dir
	cmd.Env = []string{"CH19_SECRET=not-inherited"}
	if err = files.PrepareCommand(cmd); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(cmd.Env, "\n"), "CH19_SECRET") {
		t.Fatal("inherited environment")
	}
	env := map[string]string{}
	for _, item := range cmd.Env {
		key, value, _ := strings.Cut(item, "=")
		env[key] = value
	}
	for key, want := range map[string]string{"PATH": "/usr/bin:/bin:/usr/sbin:/sbin", "HOME": filepath.Join(dir, ".ensemble-runtime", "home"), "TMPDIR": filepath.Join(dir, ".ensemble-runtime", "tmp"), "GIT_CONFIG_GLOBAL": filepath.Join(dir, ".ensemble-runtime", "gitconfig"), "GIT_CONFIG_SYSTEM": "/dev/null", "GIT_CONFIG_NOSYSTEM": "1"} {
		if env[key] != want {
			t.Errorf("confined %s=%q want %q", key, env[key], want)
		}
	}
	for _, key := range []string{"GIT_CONFIG_COUNT", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0", "XDG_CONFIG_HOME"} {
		if _, ok := env[key]; ok {
			t.Errorf("inherited Git injection: %s", key)
		}
	}
	raw, err := files.ReadFile(gitPath)
	if err != nil || string(raw) != "[user]\n\tname = Ensemble\n\temail = ensemble@localhost\n" {
		t.Fatalf("generated identity-only Git config: %q %v", raw, err)
	}
	if runtime.GOOS == "darwin" && cmd.Path != "/usr/bin/sandbox-exec" {
		t.Fatal("MCP not confined")
	}
	profile := Profile(dir, "/private/var/folders/test/T")
	if strings.Contains(profile, "(allow file-read-metadata)") {
		t.Fatal("global metadata grant")
	}
	if !strings.Contains(profile, `(literal "/")`) || !strings.Contains(profile, `xcrun_db[^/]*$`) {
		t.Fatal("missing ancestor or bounded cache grant")
	}
	if _, err = files.ValidatePath(filepath.Join(dir, "..", "outside")); err == nil {
		t.Fatal("traversal admitted")
	}
}
