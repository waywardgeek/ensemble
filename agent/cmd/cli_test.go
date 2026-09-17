package main

// Tests for the command-line surface: the part a HUMAN touches.
//
// Every defect these cover was found by someone running the binary by hand,
// and not one of them was visible to the grader. The grader always sets
// CH02_LOG, never passes --help, and never feeds a line that is not JSON, so
// all three behaviors below can rot without moving the score by a point. That
// is precisely why they are tested here instead.
//
// The tests build a real binary and run it, because two of the three
// properties — the default log NAME and the exit status — exist only in a
// process. Neither can be observed by calling a function.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildAs compiles this package to a binary with a chosen NAME, which is the
// whole point: the default log name is derived from argv[0], so the name has
// to be a variable of the test rather than a constant of the build.
func buildAs(t *testing.T, name string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), name)
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

// runBin runs the binary in workdir with stdin, and returns stdout, stderr and
// the exit code separately. Keeping the two streams apart is not fussiness: the
// entire design of the bad-input fix is that the machine protocol stays on
// stdout and the human explanation goes to stderr.
func runBin(t *testing.T, bin, workdir, stdin string, env ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Dir = workdir
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	// A deliberately bare environment. Inheriting the caller's CH02_LOG would
	// make the default-name test pass or fail according to the shell it was
	// run from, which is no test at all.
	cmd.Env = append([]string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}, env...)
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s: %v", bin, err)
	}
	return stdout.String(), stderr.String(), code
}

func runArgs(t *testing.T, bin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = t.TempDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME")}
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run: %v", err)
	}
	return stdout.String(), stderr.String(), code
}

// --help, -h and help all print the commands table on STDOUT and exit 0.
// Before this, `./ch03 --help` printed `unknown command "--help"` and exited 2.
func TestHelpPrintsCommandsTable(t *testing.T) {
	bin := buildAs(t, "ch03")
	for _, flag := range []string{"--help", "-h", "help"} {
		stdout, _, code := runArgs(t, bin, flag)
		if code != 0 {
			t.Errorf("%s: exit = %d, want 0 (help was asked for, so it is not an error)", flag, code)
		}
		// Every command the binary accepts must appear in the table it prints.
		// A usage message that omits a mode is how modes become folklore.
		for _, want := range []string{"usage:", "chat", "render LOG", "dump", "--help", "CH02_LOG"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("%s: stdout missing %q\ngot:\n%s", flag, want, stdout)
			}
		}
	}
}

// An unknown command still exits 2, but now says what the known ones are.
func TestUnknownCommandPrintsUsage(t *testing.T) {
	bin := buildAs(t, "ch03")
	_, stderr, code := runArgs(t, bin, "--bogus")
	if code != 2 {
		t.Errorf("exit = %d, want 2", code)
	}
	if !strings.Contains(stderr, `unknown command "--bogus"`) {
		t.Errorf("stderr should name the offending token, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("stderr should show the commands table, got:\n%s", stderr)
	}
}

// The default log name is derived from the executable, so a binary built from
// solutions/ch03 no longer writes ch02.log. Under the old code every binary
// built from either chapter wrote to the same ch02.log in whatever directory it
// happened to be standing in.
func TestDefaultLogNameDerivesFromExecutable(t *testing.T) {
	// An arbitrary name, not a chapter name: if the default were still
	// hardcoded to any fixed string, this could not pass.
	bin := buildAs(t, "zzagent")
	work := t.TempDir()

	// One malformed line is enough to drive a session to its end without
	// needing an API key: the log is written on the way out regardless.
	runBin(t, bin, work, "not json\n")

	if _, err := os.Stat(filepath.Join(work, "zzagent.log")); err != nil {
		got, _ := filepath.Glob(filepath.Join(work, "*"))
		t.Fatalf("want zzagent.log, found %v", got)
	}
	if _, err := os.Stat(filepath.Join(work, "ch02.log")); err == nil {
		t.Error("wrote ch02.log: the default is still hardcoded to the chapter 2 name")
	}
}

// CH02_LOG keeps its name and keeps winning. Students who passed Chapter 2 read
// this variable; only the DEFAULT was allowed to change.
func TestCH02LogStillOverridesTheDefault(t *testing.T) {
	bin := buildAs(t, "zzagent")
	work := t.TempDir()

	runBin(t, bin, work, "not json\n", "CH02_LOG=chosen.log")

	if _, err := os.Stat(filepath.Join(work, "chosen.log")); err != nil {
		t.Error("CH02_LOG no longer selects the log path")
	}
	if _, err := os.Stat(filepath.Join(work, "zzagent.log")); err == nil {
		t.Error("derived name overrode an explicit CH02_LOG")
	}
}

// Typing prose at the bare command used to produce
//
//	{"error":"bad input: invalid character 'H' looking for beginning of value"}
//
// and nothing else: accurate about the symptom, silent about the mistake.
//
// The fix adds an explanation on stderr and leaves stdout ALONE. The two
// assertions below are equally important — the second is what keeps this a
// usability fix rather than a protocol change.
func TestBadInputExplainsItselfOnStderrOnly(t *testing.T) {
	bin := buildAs(t, "ch03")
	stdout, stderr, _ := runBin(t, bin, t.TempDir(), "Hi.\n")

	// stdout: unchanged machine protocol.
	first := strings.SplitN(stdout, "\n", 2)[0]
	if !strings.HasPrefix(first, `{"error":"bad input:`) {
		t.Errorf("stdout protocol line changed shape: %q", first)
	}
	for _, leak := range []string{"JSON-lines", "chat`", "--help"} {
		if strings.Contains(stdout, leak) {
			t.Errorf("human guidance leaked onto stdout (%q); stdout is the machine protocol", leak)
		}
	}

	// stderr: names the cause and the way out.
	for _, want := range []string{"not JSON", "JSON-lines log on stdin", "chat", "--help"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr missing %q\ngot:\n%s", want, stderr)
		}
	}
}

// The explanation is printed once per session, not once per bad line. A
// malformed 10,000-line file should not produce 10,000 identical paragraphs.
func TestBadInputHintIsPrintedOnce(t *testing.T) {
	bin := buildAs(t, "ch03")
	stdout, stderr, _ := runBin(t, bin, t.TempDir(), "Hi.\nthere.\nagain.\n")

	if n := strings.Count(stderr, "that line is not JSON"); n != 1 {
		t.Errorf("hint printed %d times, want exactly 1\n%s", n, stderr)
	}
	// but every bad line is still reported on the protocol stream
	if n := strings.Count(stdout, `"error"`); n != 3 {
		t.Errorf("stdout reported %d errors, want 3 (one per bad line)", n)
	}
}
