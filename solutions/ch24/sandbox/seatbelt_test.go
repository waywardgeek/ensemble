package sandbox

// Integration tests for the kernel half.
//
// These run real processes under the real OS sandbox. They are the only
// tests here that prove the boundary actually holds, because every other
// test in this package exercises Go code that a subprocess never touches.
//
// They skip rather than fail on a platform without an implementation. A
// skip is honest; a pass would claim confinement that was never measured.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func requireSeatbelt(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skipf("command confinement is implemented for darwin, not %s", runtime.GOOS)
	}
	if !Available() {
		t.Skip("sandbox-exec not available")
	}
}

// confinedFixture builds a workspace plus a secret OUTSIDE any allowed
// path.
//
// The outside directory deliberately lives under the home directory, not
// under t.TempDir(). MEASURED: t.TempDir and mktemp -d both create
// directories inside the Darwin per-user temp directory, which the profile
// must partially allow for the Xcode toolchain shims. A test that put its
// "outside" file there would be writing into its own allowlist and would
// pass or fail for reasons having nothing to do with the sandbox.
func confinedFixture(t *testing.T) (sb *Sandbox, work, outside string) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}
	base, err := os.MkdirTemp(home, ".ch23-sandbox-test-")
	if err != nil {
		t.Skipf("cannot create a test dir outside the temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(base) })

	work = filepath.Join(base, "work")
	outside = filepath.Join(base, "outside")
	for _, d := range []string{work, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(work, "notes.txt"), []byte("inside"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("stolen"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	sb, err = Rooted(work)
	if err != nil {
		t.Fatalf("Rooted: %v", err)
	}
	return sb, work, outside
}

func runConfined(t *testing.T, sb *Sandbox, argv ...string) (string, error) {
	t.Helper()
	cmd, err := sb.Command(argv[0], argv[1:]...)
	if err != nil {
		t.Fatalf("building confined command: %v", err)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestConfinedCommandAllowsLegitimateWork is the positive control for the
// whole file. Without it, a profile that denied everything would look like
// a triumph.
func TestConfinedCommandAllowsLegitimateWork(t *testing.T) {
	requireSeatbelt(t)
	sb, _, _ := confinedFixture(t)

	out, err := runConfined(t, sb, "/bin/cat", "notes.txt")
	if err != nil || !strings.Contains(out, "inside") {
		t.Fatalf("reading a file inside the sandbox failed: out=%q err=%v", out, err)
	}

	out, err = runConfined(t, sb, "/bin/sh", "-c", "echo written > new.txt && cat new.txt")
	if err != nil || !strings.Contains(out, "written") {
		t.Fatalf("writing a file inside the sandbox failed: out=%q err=%v", out, err)
	}
}

func TestConfinedCommandCannotReadOutside(t *testing.T) {
	requireSeatbelt(t)
	sb, _, outside := confinedFixture(t)
	secret := filepath.Join(outside, "secret.txt")

	out, err := runConfined(t, sb, "/bin/cat", secret)
	if err == nil && strings.Contains(out, "stolen") {
		t.Fatalf("the sandbox did not stop a read outside it: %q", out)
	}
	if strings.Contains(out, "stolen") {
		t.Fatalf("secret contents leaked despite an error: %q", out)
	}
}

func TestConfinedCommandCannotWriteOutside(t *testing.T) {
	requireSeatbelt(t)
	sb, _, outside := confinedFixture(t)
	target := filepath.Join(outside, "evil.txt")

	_, _ = runConfined(t, sb, "/bin/sh", "-c", "echo pwned > "+target)

	// Ground truth: check the filesystem, not the exit status. A shell can
	// exit non-zero for many reasons, and the only question that matters
	// is whether the file exists.
	if _, err := os.Stat(target); err == nil {
		t.Fatal("a command inside the sandbox created a file outside it")
	}
}

// TestConfinedGrandchildIsStillConfined is the test that makes the whole
// approach worth using. If confinement applied only to the first process,
// the model would escape by asking that process to spawn another.
func TestConfinedGrandchildIsStillConfined(t *testing.T) {
	requireSeatbelt(t)
	sb, _, outside := confinedFixture(t)
	secret := filepath.Join(outside, "secret.txt")

	out, _ := runConfined(t, sb, "/bin/sh", "-c", `/bin/sh -c "cat `+secret+`"`)
	if strings.Contains(out, "stolen") {
		t.Fatalf("a grandchild process escaped the sandbox: %q", out)
	}
}

// TestConfinedCommandHasNoNetwork proves the line that breaks the lethal
// trifecta: a process that cannot reach the network cannot exfiltrate what
// it reads.
func TestConfinedCommandHasNoNetwork(t *testing.T) {
	requireSeatbelt(t)
	if testing.Short() {
		t.Skip("network check makes an outbound attempt; skipped in -short")
	}
	sb, _, _ := confinedFixture(t)

	out, err := runConfined(t, sb,
		"/usr/bin/curl", "-s", "-m", "5", "-o", "/dev/null", "-w", "HTTP%{http_code}",
		"https://example.com")
	if err == nil && strings.Contains(out, "HTTP200") {
		t.Fatalf("a confined command reached the network: %q", out)
	}
}

// TestConfinedCommandHasNoCredentials checks the environment half: the
// model can run `env` and the output is a tool result, which is context.
func TestConfinedCommandHasNoCredentials(t *testing.T) {
	requireSeatbelt(t)
	sb, work, _ := confinedFixture(t)

	t.Setenv("OPENAI_API_KEY", "sk-proj-shouldnotappear")
	t.Setenv("GITHUB_TOKEN", "ghp_shouldnotappear")

	out, err := runConfined(t, sb, "/usr/bin/env")
	if err != nil {
		t.Fatalf("env failed: %v (%q)", err, out)
	}
	if strings.Contains(out, "shouldnotappear") {
		t.Fatalf("a credential survived into the command's environment:\n%s", out)
	}
	// HOME points at the sandbox, so ~/.ssh resolves to nothing.
	if !strings.Contains(out, "HOME="+work) {
		t.Errorf("HOME was not repointed at the sandbox root:\n%s", out)
	}
}

// TestUnconfinedCommandIsNotWrapped proves confinement is opt-in, so every
// chapter before this one behaves exactly as it did.
func TestUnconfinedCommandIsNotWrapped(t *testing.T) {
	cmd, err := Unconfined().Command("/bin/echo", "hi")
	if err != nil {
		t.Fatalf("unconfined Command failed: %v", err)
	}
	if strings.Contains(cmd.Path, seatbeltBinary) {
		t.Fatalf("an unconfined command was wrapped in the sandbox: %q", cmd.Path)
	}
}

// TestProfileUsesPhysicalPaths guards the bug that cost the most time:
// SBPL matches physical paths, and the per-user temp directory is reported
// through a /var symlink that never matches.
func TestProfileUsesPhysicalPaths(t *testing.T) {
	requireSeatbelt(t)
	sb, _, _ := confinedFixture(t)

	profile, err := sb.Profile()
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	for _, line := range strings.Split(profile, "\n") {
		if strings.Contains(line, "/var/folders") && !strings.Contains(line, "/private/var/folders") {
			t.Fatalf("profile names the symlinked /var form, which never matches: %q", line)
		}
	}
}

// TestProfileRefusedWhenUnconfined: there is no such thing as a profile
// for a sandbox that confines nothing, and silently returning an empty one
// would produce a command wrapped in a permit-all profile.
func TestProfileRefusedWhenUnconfined(t *testing.T) {
	if _, err := Unconfined().Profile(); err == nil {
		t.Fatal("Profile succeeded for an unconfined sandbox")
	}
}
