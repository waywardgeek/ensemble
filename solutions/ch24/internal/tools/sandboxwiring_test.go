package tools

// Chapter 23. Proof that the file tools are actually wired to the sandbox.
//
// The package-level tests in agent/sandbox prove the BOUNDARY works. These
// prove the TOOLS go through it, which is a different claim and the one
// that would silently regress: every tool here compiled fine while calling
// the os package directly, and did so for twenty-two chapters.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// confinedAgent is a parent reporting a fixed sandbox root.
type confinedAgent struct{ root string }

func (confinedAgent) Logf(string, ...any)    {}
func (confinedAgent) APILogf(string, ...any) {}
func (confinedAgent) Debugf(string, ...any)  {}
func (c confinedAgent) SandboxRoot() string  { return c.root }

// toolFixture builds a sandboxed call plus a secret outside the sandbox.
func toolFixture(t *testing.T) (call *common.Call, root, secret string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "work")
	outside := filepath.Join(base, "outside")
	for _, d := range []string{root, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	secret = filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("STOLEN"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return &common.Call{Agent: confinedAgent{root: root}}, root, secret
}

func argsJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

// TestFileToolsRefuseEscape is the end-to-end claim: a sandboxed agent's
// tools cannot reach outside, and the refusal does not hand the model the
// host path.
func TestFileToolsRefuseEscape(t *testing.T) {
	call, root, secret := toolFixture(t)

	cases := []struct {
		name string
		run  func() (string, error)
	}{
		{"read_file absolute", func() (string, error) {
			return ToolReadFile(call, argsJSON(t, map[string]any{"path": secret}))
		}},
		{"read_file dot-dot", func() (string, error) {
			return ToolReadFile(call, argsJSON(t, map[string]any{"path": "../outside/secret.txt"}))
		}},
		{"write_file outside", func() (string, error) {
			return ToolWriteFile(call, argsJSON(t, map[string]any{
				"path": filepath.Join(filepath.Dir(root), "outside", "evil.txt"), "content": "pwned"}))
		}},
		{"edit_file outside", func() (string, error) {
			return ToolEditFile(call, argsJSON(t, map[string]any{
				"path": secret, "old_text": "STOLEN", "new_text": "x"}))
		}},
		{"list_directory outside", func() (string, error) {
			return ToolListDirectory(call, argsJSON(t, map[string]any{"path": filepath.Dir(secret)}))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.run()
			if err == nil {
				t.Fatalf("escape allowed, returned %q", out)
			}
			if strings.Contains(out, "STOLEN") {
				t.Fatalf("secret contents leaked: %q", out)
			}
			if strings.Contains(err.Error(), root) {
				t.Fatalf("refusal leaked the sandbox root to the model: %q", err)
			}
		})
	}

	// Ground truth: the refused write must not have created anything, and
	// the refused edit must not have changed the file.
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "outside", "evil.txt")); err == nil {
		t.Fatal("a refused write_file created a file outside the sandbox")
	}
	if b, _ := os.ReadFile(secret); string(b) != "STOLEN" {
		t.Fatalf("a refused edit_file modified a file outside the sandbox: %q", b)
	}
}

// TestFileToolsWorkInsideSandbox is the positive control. Without it, a
// tool that refused everything would pass every test above.
func TestFileToolsWorkInsideSandbox(t *testing.T) {
	call, _, _ := toolFixture(t)

	out, err := ToolReadFile(call, argsJSON(t, map[string]any{"path": "notes.txt"}))
	if err != nil || !strings.Contains(out, "inside") {
		t.Fatalf("read inside the sandbox failed: %q %v", out, err)
	}
	if _, err := ToolWriteFile(call, argsJSON(t, map[string]any{
		"path": "sub/new.txt", "content": "hello"})); err != nil {
		t.Fatalf("write inside the sandbox failed: %v", err)
	}
	if _, err := ToolEditFile(call, argsJSON(t, map[string]any{
		"path": "sub/new.txt", "old_text": "hello", "new_text": "goodbye"})); err != nil {
		t.Fatalf("edit inside the sandbox failed: %v", err)
	}
	if _, err := ToolListDirectory(call, argsJSON(t, map[string]any{"path": "."})); err != nil {
		t.Fatalf("list inside the sandbox failed: %v", err)
	}
	out, err = ToolSearchFiles(call, argsJSON(t, map[string]any{"pattern": "goodbye"}))
	if err != nil || !strings.Contains(out, "goodbye") {
		t.Fatalf("search inside the sandbox failed: %q %v", out, err)
	}
}

// TestSearchFilesDoesNotLeakHostPaths: the walk yields absolute resolved
// paths, so without translation every grep hit would print the host
// layout back into the model's context.
func TestSearchFilesDoesNotLeakHostPaths(t *testing.T) {
	call, root, _ := toolFixture(t)

	out, err := ToolSearchFiles(call, argsJSON(t, map[string]any{"pattern": "inside"}))
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if strings.Contains(out, root) {
		t.Fatalf("search results leaked the sandbox root:\n%s", out)
	}
	if !strings.Contains(out, "notes.txt") {
		t.Fatalf("search did not find the file it should have:\n%s", out)
	}
}

// TestUnconfinedToolsBehaveAsBefore: confinement is opt-in, so an agent
// with no sandbox root works exactly as it did in every earlier chapter.
func TestUnconfinedToolsBehaveAsBefore(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "anywhere.txt")
	if err := os.WriteFile(target, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	call := &common.Call{Agent: confinedAgent{root: ""}}

	out, err := ToolReadFile(call, argsJSON(t, map[string]any{"path": target}))
	if err != nil || !strings.Contains(out, "ok") {
		t.Fatalf("an unconfined agent could not read an absolute path: %q %v", out, err)
	}
}

// TestNilCallFailsClosed: a tool invoked without a call has no agent and
// therefore no known boundary. It must refuse, not assume freedom.
func TestNilCallFailsClosed(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "x.txt")
	if err := os.WriteFile(target, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := ToolReadFile(nil, argsJSON(t, map[string]any{"path": target})); err == nil {
		t.Fatal("a tool with no call read a file; it must fail closed")
	}
}
