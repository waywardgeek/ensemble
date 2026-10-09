package sandbox

// The attack suite.
//
// Every escape has a matching positive control proving the legitimate
// operation still works. A sandbox that rejects everything passes every
// negative test and is worthless, so a suite made only of negative tests
// cannot tell a working boundary from a brick.
//
// Note for anyone running these on macOS: t.TempDir() returns a path under
// /var/folders/..., and /var is a symlink to /private/var. That is not an
// inconvenience to work around; it is the exact condition the root
// canonicalization exists for, so the tests use the raw TempDir path
// deliberately.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fixture builds a sandbox root with one real file, plus an outside
// directory holding a secret the attacks try to reach.
func fixture(t *testing.T) (sb *Sandbox, root, outside string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "work")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{root, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "notes.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("stolen"), 0o644); err != nil {
		t.Fatalf("write secret.txt: %v", err)
	}
	sb, err := Rooted(root)
	if err != nil {
		t.Fatalf("Rooted: %v", err)
	}
	return sb, root, outside
}

// --- the back-pointer ------------------------------------------------------

// movingAgent is a parent whose boundary changes. It exists to prove the
// Sandbox asks its parent every time instead of caching a copy, which is
// the whole reason the back-pointer is there.
type movingAgent struct{ root string }

func (m *movingAgent) SandboxRoot() string { return m.root }

func TestSandboxFollowsItsParent(t *testing.T) {
	base := t.TempDir()
	a := filepath.Join(base, "a")
	b := filepath.Join(base, "b")
	for _, d := range []string{a, b} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	parent := &movingAgent{root: a}
	sb := New(parent)

	if _, err := sb.Resolve(filepath.Join(a, "x.txt")); err != nil {
		t.Fatalf("path in the current root was refused: %v", err)
	}
	if _, err := sb.Resolve(filepath.Join(b, "x.txt")); err == nil {
		t.Fatal("path outside the current root was allowed")
	}

	// The parent moves. A Sandbox holding a copy would still be enforcing
	// the old boundary; one holding a back-pointer follows.
	parent.root = b

	if _, err := sb.Resolve(filepath.Join(b, "x.txt")); err != nil {
		t.Fatalf("after the parent moved, the new root was refused: %v", err)
	}
	if _, err := sb.Resolve(filepath.Join(a, "x.txt")); err == nil {
		t.Fatal("after the parent moved, the OLD root was still allowed: the boundary was cached")
	}
}

// TestNilSandboxFailsClosed is the wiring-mistake test. An uninitialized
// sandbox must refuse, never silently allow.
func TestNilSandboxFailsClosed(t *testing.T) {
	var sb *Sandbox
	if _, err := sb.Resolve("anything"); err == nil {
		t.Fatal("a nil Sandbox allowed a path")
	}
	if _, err := sb.ReadFile("anything"); err == nil {
		t.Fatal("a nil Sandbox allowed a read")
	}
	if New(nil).IsConfined() {
		t.Fatal("a Sandbox with a nil parent reported itself confined")
	}
	if _, err := New(nil).ReadFile("x"); err == nil {
		t.Fatal("a Sandbox with a nil parent allowed a read")
	}
}

func TestUnconfinedIsExplicit(t *testing.T) {
	sb := Unconfined()
	if sb.IsConfined() {
		t.Fatal("Unconfined reported itself confined")
	}
	got, err := sb.Resolve("/etc/passwd")
	if err != nil {
		t.Fatalf("unconfined resolve failed: %v", err)
	}
	if got != "/etc/passwd" {
		t.Fatalf("unconfined resolve changed the path: %q", got)
	}
	// Rooted must refuse an empty root: unconfined has to be asked for by
	// name, never produced by an empty string.
	if _, err := Rooted(""); err == nil {
		t.Fatal("Rooted(\"\") was accepted; unconfined must be explicit")
	}
}

// --- path confinement ------------------------------------------------------

func TestResolveAllowsLegitimatePaths(t *testing.T) {
	sb, root, _ := fixture(t)

	cases := []struct{ name, path string }{
		{"relative file", "notes.txt"},
		{"relative with dot", "./notes.txt"},
		{"nested relative that does not exist yet", "sub/dir/new.txt"},
		{"absolute inside", filepath.Join(root, "notes.txt")},
		{"the root itself", root},
		{"dot-dot that stays inside", "sub/../notes.txt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sb.Resolve(tc.path)
			if err != nil {
				t.Fatalf("legitimate path %q was refused: %v", tc.path, err)
			}
			if !withinRoot(sb.Root(), got) {
				t.Fatalf("resolved %q to %q, outside %q", tc.path, got, sb.Root())
			}
		})
	}
}

func TestResolveRejectsDotDotEscape(t *testing.T) {
	sb, _, outside := fixture(t)
	for _, p := range []string{
		"../outside/secret.txt",
		"../../etc/passwd",
		"sub/../../outside/secret.txt",
		"./../../etc/passwd",
		filepath.Join(outside, "secret.txt"),
		"/etc/passwd",
	} {
		t.Run(p, func(t *testing.T) {
			if got, err := sb.Resolve(p); err == nil {
				t.Fatalf("escape %q was allowed, resolved to %q", p, got)
			}
		})
	}
}

func TestResolveRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	sb, root, outside := fixture(t)

	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "innocent.txt")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if _, err := sb.Resolve("innocent.txt"); err == nil {
		t.Fatal("a symlink pointing outside the sandbox was followed")
	}

	// Positive control: a symlink that stays inside still works.
	if err := os.Symlink(filepath.Join(root, "notes.txt"), filepath.Join(root, "alias.txt")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if _, err := sb.Resolve("alias.txt"); err != nil {
		t.Fatalf("a symlink inside the sandbox was refused: %v", err)
	}
}

// TestResolveRejectsSymlinkedDirectoryEscape is the subtler version: the
// symlink is a DIRECTORY mid-path and the leaf may not exist yet. This is
// what resolveExistingPrefix is for.
func TestResolveRejectsSymlinkedDirectoryEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	sb, root, outside := fixture(t)

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("symlink: %v", err)
	}
	if _, err := sb.Resolve("escape/secret.txt"); err == nil {
		t.Fatal("read through a symlinked directory escaped the sandbox")
	}
	if _, err := sb.Resolve("escape/newfile.txt"); err == nil {
		t.Fatal("write through a symlinked directory escaped the sandbox")
	}
}

func TestResolveNonexistentPaths(t *testing.T) {
	sb, _, _ := fixture(t)
	if _, err := sb.Resolve("a/b/c/new.txt"); err != nil {
		t.Fatalf("a new file inside the sandbox was refused: %v", err)
	}
	if _, err := sb.Resolve("sub/../../../etc/cron.d/backdoor"); err == nil {
		t.Fatal("a nonexistent path escaping the sandbox was allowed")
	}
}

// TestResolveRejectsSiblingPrefix is the string-versus-path bug. A plain
// HasPrefix passes this case wrongly, because "/x/work-evil" really does
// begin with "/x/work".
func TestResolveRejectsSiblingPrefix(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "work")
	sibling := filepath.Join(base, "work-evil")
	for _, d := range []string{root, sibling} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	secret := filepath.Join(sibling, "secret.txt")
	if err := os.WriteFile(secret, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sb, err := Rooted(root)
	if err != nil {
		t.Fatalf("Rooted: %v", err)
	}
	if _, err := sb.Resolve(secret); err == nil {
		t.Fatal("a sibling directory sharing the root's name prefix was treated as inside")
	}
}

// TestResolveErrorDoesNotLeakHostPath: a refusal names the violation
// without handing the model the host filesystem layout.
func TestResolveErrorDoesNotLeakHostPath(t *testing.T) {
	sb, root, _ := fixture(t)
	_, err := sb.Resolve("../../etc/passwd")
	if err == nil {
		t.Fatal("expected refusal")
	}
	if strings.Contains(err.Error(), root) {
		t.Fatalf("error leaked the sandbox root to the model: %q", err)
	}
	if !strings.Contains(err.Error(), "../../etc/passwd") {
		t.Fatalf("error should name what was asked for, got %q", err)
	}
}

// --- the confined operations ----------------------------------------------

func TestOperationsAreConfined(t *testing.T) {
	sb, _, outside := fixture(t)
	escape := filepath.Join(outside, "secret.txt")

	// Positive controls first.
	if data, err := sb.ReadFile("notes.txt"); err != nil || string(data) != "inside" {
		t.Fatalf("ReadFile inside: %q %v", data, err)
	}
	if err := sb.WriteFile("new/deep.txt", []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile inside: %v", err)
	}
	if _, err := sb.ReadDir("."); err != nil {
		t.Fatalf("ReadDir inside: %v", err)
	}
	if err := sb.MkdirAll("made/up", 0o755); err != nil {
		t.Fatalf("MkdirAll inside: %v", err)
	}

	// Every operation refuses the escape.
	if _, err := sb.ReadFile(escape); err == nil {
		t.Error("ReadFile escaped")
	}
	if err := sb.WriteFile(escape, []byte("x"), 0o644); err == nil {
		t.Error("WriteFile escaped")
	}
	if err := sb.Append(escape, []byte("x"), 0o644); err == nil {
		t.Error("Append escaped")
	}
	if _, err := sb.Open(escape); err == nil {
		t.Error("Open escaped")
	}
	if _, err := sb.Create(escape); err == nil {
		t.Error("Create escaped")
	}
	if _, err := sb.Stat(escape); err == nil {
		t.Error("Stat escaped")
	}
	if _, err := sb.ReadDir(outside); err == nil {
		t.Error("ReadDir escaped")
	}
	if err := sb.MkdirAll(filepath.Join(outside, "nope"), 0o755); err == nil {
		t.Error("MkdirAll escaped")
	}
	if err := sb.Remove(escape); err == nil {
		t.Error("Remove escaped")
	}

	// The escape target must still exist: a refused Remove must not have
	// deleted it on the way to refusing.
	if _, err := os.Stat(escape); err != nil {
		t.Fatalf("a refused operation damaged a file outside the sandbox: %v", err)
	}
}

// --- environment -----------------------------------------------------------

func TestSanitizedEnvStripsCredentials(t *testing.T) {
	root := t.TempDir()
	in := []string{
		"PATH=/usr/bin:/bin", "LANG=en_US.UTF-8", "TERM=xterm",
		"HOME=/Users/someone", "EDITOR=vim",
		"OPENAI_API_KEY=sk-proj-realkey", "ANTHROPIC_API_KEY=sk-ant-realkey",
		"GOOGLE_API_KEY=goog-realkey", "LLM_API_KEY=llm-realkey",
		"GITHUB_TOKEN=ghp_realtoken", "DB_PASSWORD=hunter2",
		"SOME_SECRET=shh", "VENDOR_CREDENTIALS=blob",
		"SSH_AUTH_SOCK=/tmp/ssh-agent.sock",
	}
	joined := strings.Join(SanitizedEnv(in, root), "\n")

	for _, leaked := range []string{"realkey", "realtoken", "hunter2", "shh", "blob", "ssh-agent.sock"} {
		if strings.Contains(joined, leaked) {
			t.Errorf("SanitizedEnv leaked %q:\n%s", leaked, joined)
		}
	}
	// Positive control: the command still has a usable environment.
	for _, kept := range []string{"PATH=/usr/bin:/bin", "LANG=en_US.UTF-8", "TERM=xterm", "EDITOR=vim"} {
		if !strings.Contains(joined, kept) {
			t.Errorf("SanitizedEnv dropped a tool-functional variable %q", kept)
		}
	}
	if !strings.Contains(joined, "HOME="+root) {
		t.Errorf("HOME was not repointed at the sandbox root:\n%s", joined)
	}
	if strings.Contains(joined, "HOME=/Users/someone") {
		t.Errorf("the original HOME survived:\n%s", joined)
	}
}

func TestIsSensitiveEnvIsCaseInsensitive(t *testing.T) {
	for _, name := range []string{"openai_api_key", "Github_Token", "db_password"} {
		if !isSensitiveEnv(name) {
			t.Errorf("%q should be treated as sensitive", name)
		}
	}
	for _, name := range []string{"PATH", "KEYBOARD_LAYOUT", "TOKENIZER"} {
		if isSensitiveEnv(name) {
			t.Errorf("%q should not be treated as sensitive", name)
		}
	}
}

// --- log sanitization ------------------------------------------------------

func TestSanitizeForLog(t *testing.T) {
	key := "sk-proj-abcdefghijklmnop7xQ2"
	got := SanitizeForLog("POST https://api.openai.com/v1/responses auth="+key+" ok", []string{key})

	if strings.Contains(got, key) {
		t.Fatalf("the credential survived sanitization: %q", got)
	}
	if !strings.Contains(got, "sk-p") || !strings.Contains(got, "7xQ2") {
		t.Fatalf("sanitized form lost its debugging value: %q", got)
	}
	if !strings.Contains(got, "https://api.openai.com/v1/responses") {
		t.Fatalf("sanitization damaged the log line: %q", got)
	}
}

func TestSanitizeForLogShortCredential(t *testing.T) {
	short := "abcd1234"
	if got := SanitizeForLog("key="+short, []string{short}); strings.Contains(got, short) {
		t.Fatalf("short credential survived: %q", got)
	}
}

func TestSanitizeForLogIgnoresEmptyCreds(t *testing.T) {
	line := "nothing secret here"
	if got := SanitizeForLog(line, []string{""}); got != line {
		t.Fatalf("empty credential corrupted the line: %q", got)
	}
}
