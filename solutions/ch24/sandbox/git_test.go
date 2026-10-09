package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Git is the test case that decides whether read confinement is usable.
// It is the tool an agent reaches for most, and the one that most reliably
// reads a credential store, so "sandbox the agent" and "let it use git" look
// like opposing requirements. They are not: git needs an identity, not the
// host's identity.
//
// These tests run real git under the real kernel sandbox. There is no fake
// here, because the thing being tested is whether the kernel lets git do its
// job, and a fake kernel would answer a different question.

func gitSandbox(t *testing.T) (*Sandbox, string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Fatalf("git is not installed, so this test cannot check anything: %v", err)
	}
	root := t.TempDir()
	return New(&movingAgent{root: root}), root
}

// POSITIVE CONTROL, and the one that matters most: a confined agent can
// still do ordinary version control. Without this, denying git access to
// everything would look like a pass.
func TestSandboxedGitCanCommit(t *testing.T) {
	s, root := gitSandbox(t)

	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("seeding the repository: %v", err)
	}

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "add", "file.txt"},
		{"git", "commit", "-m", "initial"},
		{"git", "log", "--oneline"},
	} {
		cmd, err := s.Command(args[0], args[1:]...)
		if err != nil {
			t.Fatalf("building %v: %v", args, err)
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed inside the sandbox: %v\n%s", args, err, out)
		}
	}
}

// The identity git uses must come from inside the sandbox. If it came from
// the host config, that config was readable, and a readable gitconfig is a
// readable credential helper.
func TestSandboxedGitUsesTheSandboxIdentity(t *testing.T) {
	s, _ := gitSandbox(t)

	cmd, err := s.Command("git", "config", "--global", "--get", "user.email")
	if err != nil {
		t.Fatalf("building command: %v", err)
	}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("reading git identity inside the sandbox: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if got != "agent@ensemble.invalid" {
		t.Errorf("git identity inside the sandbox is %q, want the generated one; "+
			"a different value means git read a config outside the sandbox", got)
	}
}

// The host's own config must be unreachable as a file, independently of
// where git happens to look. This is the kernel's claim, not git's.
func TestSandboxedGitCannotReadHostConfig(t *testing.T) {
	s, _ := gitSandbox(t)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("locating the host home directory: %v", err)
	}
	hostConfig := filepath.Join(home, ".gitconfig")
	if _, err := os.Stat(hostConfig); err != nil {
		t.Fatalf("no host ~/.gitconfig exists, so this test would pass vacuously: %v", err)
	}

	cmd, err := s.Command("cat", hostConfig)
	if err != nil {
		t.Fatalf("building command: %v", err)
	}
	out, _ := cmd.CombinedOutput()

	if !strings.Contains(string(out), "not permitted") &&
		!strings.Contains(string(out), "Operation not permitted") {
		t.Errorf("the host gitconfig was readable from inside the sandbox; output was:\n%s", out)
	}
}

// An inherited GIT_CONFIG_GLOBAL must not survive into the sandbox, or the
// redirection is bypassable by whoever launched the agent.
func TestSandboxedGitIgnoresInheritedConfigVars(t *testing.T) {
	root := t.TempDir()
	env := SanitizedEnv([]string{
		"GIT_CONFIG_GLOBAL=/etc/evil.gitconfig",
		"GIT_CONFIG_SYSTEM=/etc/evil.system",
		"GIT_CONFIG=/etc/evil.plain",
		"PATH=/usr/bin",
	}, root)

	for _, kv := range env {
		if strings.HasPrefix(kv, "GIT_CONFIG_GLOBAL=") &&
			kv != "GIT_CONFIG_GLOBAL="+filepath.Join(root, ".gitconfig") {
			t.Errorf("inherited GIT_CONFIG_GLOBAL survived: %q", kv)
		}
		if strings.HasPrefix(kv, "GIT_CONFIG=") {
			t.Errorf("inherited GIT_CONFIG survived: %q", kv)
		}
		if kv == "GIT_CONFIG_SYSTEM=/etc/evil.system" {
			t.Errorf("inherited GIT_CONFIG_SYSTEM survived: %q", kv)
		}
	}
}

func TestWriteGitConfigLeavesAnExistingFileAlone(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gitconfig")
	if err := os.WriteFile(path, []byte("[user]\n\tname = Chosen\n"), 0o600); err != nil {
		t.Fatalf("seeding: %v", err)
	}

	if err := WriteGitConfig(root); err != nil {
		t.Fatalf("WriteGitConfig: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}
	if !strings.Contains(string(got), "Chosen") {
		t.Errorf("an existing config was overwritten: %q", got)
	}
}
