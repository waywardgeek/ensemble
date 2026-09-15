package grade

// Chapter 4's grader audit (course-policy P9).
//
// Every mutant below deletes ONE behavior from the reference solution and
// asserts the EXACT set of check ids that fails. A mutation that does not
// apply is a test failure, not a silent pass: an unapplied mutation scores
// 100 and manufactures a fake finding, and this project has been bitten by
// that three times.
//
// Two mutants here exist to prove NEGATIVE CONTROLS are load-bearing:
//   - handle-only-for-run-command: the common wrong answer. Without the
//     read_file and list_directory legs of jobmodel it scores full marks.
//   - truncation-in-run-command-only: caps the shell's output and nothing
//     else. Without the read_file leg of bigoutput it scores full marks.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type ch4edit struct {
	file    string
	find    string // regexp, must match EXACTLY once
	replace string
}

type ch4mutation struct {
	name     string
	wantFail []string
	edits    []ch4edit
	why      string
}

func ch4ReferenceDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate this test file")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "solutions", "ch04")
}

// buildCh4Mutant copies the reference, applies the edits, and builds it.
//
// Unlike Chapter 3's rig the mutant is not a dependency-free module: the
// reference imports creack/pty, so the mutant's go.mod carries the same
// require line the course module does, and the course go.sum is copied in
// so the build is offline and pinned.
func buildCh4Mutant(t *testing.T, m ch4mutation) string {
	t.Helper()
	ref := ch4ReferenceDir(t)
	root := filepath.Join(ref, "..", "..")
	dir := t.TempDir()

	entries, err := os.ReadDir(ref)
	if err != nil {
		t.Fatalf("read reference: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(ref, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatalf("write %s: %v", e.Name(), err)
		}
	}
	rootMod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read course go.mod: %v", err)
	}
	req := regexp.MustCompile(`(?m)^require github\.com/creack/pty v[^\s]+$`).Find(rootMod)
	if req == nil {
		t.Fatal("course go.mod no longer requires creack/pty; the mutant module needs its version")
	}
	mod := "module mutant\n\ngo 1.21\n\n" + string(req) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	sum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatalf("read course go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0o644); err != nil {
		t.Fatalf("write go.sum: %v", err)
	}

	for _, ed := range m.edits {
		path := filepath.Join(dir, ed.file)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s: mutation targets %s, which does not exist", m.name, ed.file)
		}
		re, err := regexp.Compile(ed.find)
		if err != nil {
			t.Fatalf("%s: bad pattern %q: %v", m.name, ed.find, err)
		}
		hits := re.FindAllIndex(b, -1)
		if len(hits) != 1 {
			t.Fatalf("MUTATION DID NOT LAND: %s: pattern %q matched %d times in %s, want exactly 1. "+
				"An unapplied mutation scores 100 and manufactures a fake finding.",
				m.name, ed.find, len(hits), ed.file)
		}
		mutated := re.ReplaceAll(b, []byte(ed.replace))
		if string(mutated) == string(b) {
			t.Fatalf("MUTATION DID NOT LAND: %s: replacing %q in %s changed no bytes", m.name, ed.find, ed.file)
		}
		if err := os.WriteFile(path, mutated, 0o644); err != nil {
			t.Fatalf("%s: write %s: %v", m.name, ed.file, err)
		}
	}

	bin := filepath.Join(dir, "mutant-bin")
	cmd := exec.Command("go", "build", "-mod=mod", "-o", bin, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: mutant does not compile: %v\n%s", m.name, err, out)
	}
	return bin
}

func scoreCh4(t *testing.T, bin string) (total int, failed []string) {
	t.Helper()
	res, err := Ch4Run(bin)
	if err != nil {
		t.Fatalf("Ch4Run: %v", err)
	}
	for _, c := range Ch4Evaluate(res) {
		total += c.Earned
		if !c.Passed {
			failed = append(failed, c.ID)
		}
	}
	sort.Strings(failed)
	return total, failed
}

// --- the mutants -----------------------------------------------------------

func ch4Mutants() []ch4mutation {
	return []ch4mutation{
		// ---- the job model ------------------------------------------------
		{
			name: "handle-only-for-run-command",
			why: "THE COMMON WRONG ANSWER, and jobmodel's negative control. Only the shell becomes a job; " +
				"the other five run inline as in Chapter 3. Without the read_file and list_directory legs this scores 100.",
			wantFail: []string{"bigoutput", "jobmodel"},
			edits: []ch4edit{{
				file: "engine.go", find: `if err != nil \|\| tool\.NoJob \{`,
				replace: `if err != nil || tool.NoJob || call.Name != "run_command" {`,
			}},
		},
		{
			name:     "handle-issued-after-the-tool-ran",
			why:      "The tool_called record carries no job: the handle exists only once the tool is back, which is too late to be useful.",
			wantFail: []string{"jobmodel"},
			edits: []ch4edit{{
				file: "engine.go", find: `CallID: call\.CallID, Name: call\.Name, Args: call\.Args, Job: job\.Data\(\),`,
				replace: `CallID: call.CallID, Name: call.Name, Args: call.Args,`,
			}},
		},
		{
			name:     "result-not-written-to-disk",
			why:      "A non-streaming tool's result reaches the model and never the file. The path in the log points at nothing.",
			wantFail: []string{"bigoutput", "jobmodel"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\t\t\tif j\.file != nil \{\n\t\t\t\t_, _ = j\.file\.WriteString\(text\)\n\t\t\t\}\n`,
				replace: "",
			}},
		},
		{
			name:     "job-verbs-are-jobs-too",
			why:      "wait_for_job gets a handle of its own. Supervising a job is not a job, and a log that says otherwise misleads the reader about what ran.",
			wantFail: []string{"jobmodel"},
			edits: []ch4edit{{
				file: "tools.go", find: `Run:   toolWaitForJob,\n\t\tNoJob: true,`, replace: `Run: toolWaitForJob,`,
			}},
		},

		// ---- wait_for_job -------------------------------------------------
		{
			name: "wait-blocks-on-a-finished-job",
			why: "Works the first time (the dispatcher's own wait) and then blocks for the full delay on any job that has " +
				"already finished. The student sees a hang and calls it a deadlock in their own code.",
			wantFail: []string{"killjob", "waitjob"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\t\tif j\.status != StatusRunning \{\n\t\t\tj\.mu\.Unlock\(\)\n\t\t\treturn WokeDone`,
				replace: "\t\tif j.status != StatusRunning && j.cursor == 0 {\n\t\t\tj.mu.Unlock()\n\t\t\treturn WokeDone",
			}},
		},

		// ---- send_input and the pattern -----------------------------------
		{
			name: "pattern-matches-output-already-seen",
			why: "The wake pattern is matched against the whole output instead of the part the model has not seen, so a " +
				"prompt that was already shown wakes the very next call before the process has answered.",
			wantFail: []string{"debugger", "sendinput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `l\.Pattern\.Match\(j\.out\.Bytes\(\)\[j\.cursor:\]\)`, replace: `l.Pattern.Match(j.out.Bytes())`,
			}},
		},
		{
			name:     "cursor-never-advances",
			why:      "Every report repeats everything, and the pattern re-matches the old prompt for the same reason as above.",
			wantFail: []string{"debugger", "sendinput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\tj\.cursor = len\(all\)\n`, replace: "",
			}},
		},
		{
			name:     "send-input-writes-nothing",
			why:      "The input never reaches the process; the tool waits politely for an answer to a question it did not ask.",
			wantFail: []string{"debugger", "sendinput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `_, err := w\.Write\(\[\]byte\(text\)\)\n\treturn err`, replace: "_ = w\n\treturn nil",
			}},
		},

		// ---- kill_job -----------------------------------------------------
		{
			name:     "kill-marks-but-does-not-kill",
			why:      "Status says killed, the process is still running. The pid file says so after the agent has exited.",
			wantFail: []string{"killjob", "shutdown"},
			edits: []ch4edit{{
				file: "jobs.go", find: `_ = syscall\.Kill\(-j\.proc\.Pid, syscall\.SIGKILL\)`, replace: `_ = syscall.SIGKILL`,
			}},
		},
		{
			name:     "kill-not-recorded",
			why:      "The process dies and the log never says so. A transcript that shows a job that simply stops is not a transcript.",
			wantFail: []string{"killjob"},
			edits: []ch4edit{{
				file: "jobtools.go", find: `\tc\.Events = append\(c\.Events, Event\{Type: JobKilled, Job: data\}\)\n`, replace: "",
			}},
		},
		{
			name: "killed-job-reported-as-done",
			why: "The process is killed but the job's status is left to the exiting process to set, so it ends as `done` with " +
				"an exit code. Whoever waits on it is told it finished, which is not what happened.",
			wantFail: []string{"killjob", "shutdown"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\tj\.status = StatusKilled\n`, replace: "",
			}},
		},

		// ---- big output ---------------------------------------------------
		{
			name:     "no-truncation-at-all",
			why:      "A megabyte enters the context window, and is paid for again on every subsequent turn.",
			wantFail: []string{"bigoutput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\tif len\(s\) <= max \{\n\t\treturn s\n\t\}\n\thalf`, replace: "\tif true {\n\t\treturn s\n\t}\n\thalf",
			}},
		},
		{
			name: "truncation-in-run-command-only",
			why: "bigoutput's NEGATIVE CONTROL. The shell's output is capped and nothing else is, which is what a student " +
				"who thinks of big output as a shell problem builds. Without the read_file leg this scores 100.",
			wantFail: []string{"bigoutput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\tbody := capText\(unseen, l\.MaxOutput, len\(all\), j\.Path\)\n`,
				replace: "\tbody := unseen\n\tif j.Tool == \"run_command\" {\n\t\tbody = capText(unseen, l.MaxOutput, len(all), j.Path)\n\t}\n",
			}},
		},
		{
			name:     "stub-does-not-name-the-path",
			why:      "The stub says how much is missing and not where it is. The path is the recovery route, not decoration.",
			wantFail: []string{"bigoutput"},
			edits: []ch4edit{{
				file: "jobs.go", find: `"\[\.\.\. %d bytes omitted; full output \(%d bytes\) at %s \.\.\.\]\\n", omitted, total, path\)`,
				replace: `"[... %d bytes omitted; full output (%d bytes) ...]\n", omitted, total)`,
			}},
		},

		// ---- tool_limits --------------------------------------------------
		{
			name:     "tool-limits-ignored",
			why:      "tool_limits accepts the setting and the next call never consumes it. The model set a delay and nothing happened.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\tif js\.pending != nil \{\n\t\tl = \*js\.pending\n\t\tjs\.pending = nil\n\t\tconsumed = true\n\t\}\n`, replace: "",
			}},
		},
		{
			name:     "tool-limits-sticky",
			why:      "The limits persist past the next call. A 600-second delay set once for a build applies to a read_file twenty calls later.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\t\tl = \*js\.pending\n\t\tjs\.pending = nil\n`, replace: "\t\tl = *js.pending\n",
			}},
		},
		{
			name:     "pending-limits-beat-explicit-arguments",
			why:      "When tool_limits is pending, the call's own ai_callback_delay is ignored. The nearer instruction must win.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "jobs.go", find: `\t\tconsumed = true\n\t\}\n\tjs\.mu\.Unlock\(\)\n`,
				replace: "\t\tconsumed = true\n\t\tjs.mu.Unlock()\n\t\treturn l, true, nil\n\t}\n\tjs.mu.Unlock()\n",
			}},
		},
		{
			name:     "tool-limits-is-a-job",
			why:      "Setting limits for the next call becomes a job itself, with a handle and a file, which is one more thing for the model to supervise for no reason.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "tools.go", find: `Run:   toolLimits,\n\t\tNoJob: true,`, replace: `Run: toolLimits,`,
			}},
		},

		// ---- tool_limits consumed-by note -----------------------------------
		{
			name:     "no-note-jobpath",
			why:      "A job that consumed the pending tool_limits says nothing about it. The model set a delay, the next call ate it, and nothing in the result explains why the wait was short.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "engine.go", find: `\n\tif fromPending \{\n\t\tout = pendingNote\(call\.Name, limits\) \+ out\n\t\}\n`, replace: "\n",
			}},
		},
		{
			name:     "no-note-nojobpath",
			why:      "The inline verbs consume a pending tool_limits silently. tool_limits followed by tool_limits is the easy case to forget, and a grader with only the job leg scores this at 100.",
			wantFail: []string{"toollimits"},
			edits: []ch4edit{{
				file: "engine.go", find: `\t\tif fromPending \{\n\t\t\tout = pendingNote\(call\.Name, limits\) \+ out\n\t\t\}\n`, replace: "",
			}},
		},
		{
			name: "note-always",
			why: "Every result carries the consumed-by note whether or not anything was pending. A note that is always there says nothing. " +
				"Measured: it also drops jobmodel, because the note is dispatcher decoration prepended to the wire text while cr/io/<handle> holds the tool's output, and the two no longer byte-match.",
			wantFail: []string{"jobmodel", "toollimits"},
			edits: []ch4edit{{
				file: "engine.go", find: `\n\tif fromPending \{\n\t\tout = pendingNote\(call\.Name, limits\) \+ out\n\t\}\n`,
				replace: "\n\tout = pendingNote(call.Name, limits) + out\n",
			}},
		},

		// ---- cwd ------------------------------------------------------------
		{
			name:     "cwd-ignored",
			why:      "run_command accepts cwd, records it, and runs in the working directory anyway. The record and the process disagree; the grader must believe the process.",
			wantFail: []string{"jobmodel"},
			edits: []ch4edit{{
				file: "tools.go", find: `\t\tcmd\.Dir = dir\n`, replace: "",
			}},
		},
		{
			name:     "cwd-sticky",
			why:      "cwd persists to the next call: a directory chosen once for a build is where every later command runs, and nothing after compaction says so.",
			wantFail: []string{"jobmodel"},
			edits: []ch4edit{
				{file: "tools.go", find: `\tcmd := exec\.Command\("sh", "-c", a\.Command\)\n\tcmd\.Env = append\(os\.Environ\(\), "TERM=dumb"\)\n`,
					replace: "\tcmd := exec.Command(\"sh\", \"-c\", a.Command)\n\tcmd.Env = append(os.Environ(), \"TERM=dumb\")\n\tcmd.Dir = stickyCwd\n"},
				{file: "tools.go", find: `\t\tcmd\.Dir = dir\n`, replace: "\t\tcmd.Dir = dir\n\t\tstickyCwd = dir\n"},
				{file: "tools.go", find: `\n// --- read_file ---`, replace: "\nvar stickyCwd string\n\n// --- read_file ---"},
			},
		},
		{
			name:     "cwd-fallback",
			why:      "A cwd that does not exist silently runs in the working directory. The output looks like an answer to a question nobody asked.",
			wantFail: []string{"jobmodel"},
			edits: []ch4edit{{
				file: "tools.go", find: `\t\tst, err := os\.Stat\(dir\)\n\t\tif err != nil \{\n\t\t\treturn "", fmt\.Errorf\("run_command: cwd %q: %v", a\.Cwd, err\)\n\t\t\}\n\t\tif !st\.IsDir\(\) \{\n\t\t\treturn "", fmt\.Errorf\("run_command: cwd %q is not a directory", a\.Cwd\)\n\t\t\}\n`,
				replace: "\t\tif st, err := os.Stat(dir); err != nil || !st.IsDir() {\n\t\t\tdir = \"\"\n\t\t}\n",
			}},
		},

		// ---- shutdown -----------------------------------------------------
		{
			name:     "no-shutdown",
			why:      "The agent exits and leaves its jobs running. The human is left with a process no transcript accounts for.",
			wantFail: []string{"shutdown"},
			edits: []ch4edit{{
				file: "main.go", find: `\t_ = eng\.Shutdown\(\)\n`, replace: "\t_ = eng.Save()\n",
			}},
		},
		{
			name:     "shutdown-kills-silently",
			why:      "The jobs are killed at exit and the log does not say so.",
			wantFail: []string{"shutdown"},
			edits: []ch4edit{{
				file: "engine.go", find: `\t\tif err := e\.record\(Event\{Type: JobKilled, Job: data\}\); err != nil \{\n\t\t\treturn err\n\t\t\}\n`,
				replace: "\t\t_ = data\n",
			}},
		},

		// ---- Chapter 3 still has to work ----------------------------------
		{
			name:     "ch3-read-range-ignored",
			why:      "ch3parity must be able to fail, or it is a decoration rather than a regression guard.",
			wantFail: []string{"ch3parity"},
			edits: []ch4edit{{
				file: "tools.go", find: `if a\.StartLine > 0 \|\| a\.EndLine > 0 \{`, replace: `if false {`,
			}},
		},
		{
			name:     "exit-code-not-reported",
			why:      "The exit status never reaches the model or the file. Chapter 3's shell checks and this chapter's file checks both notice.",
			wantFail: []string{"ch3parity", "sendinput", "waitjob"},
			edits: []ch4edit{{
				file: "tools.go", find: `\tfmt\.Fprintf\(c\.Job, "exit_code: %d\\n", exitCode\)\n`, replace: "",
			}},
		},
	}
}

// --- the tests -------------------------------------------------------------

func TestCh4ReferenceScores100(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs the reference solution")
	}
	ref := ch4ReferenceDir(t)
	bin := filepath.Join(t.TempDir(), "ch04")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = ref
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build reference: %v\n%s", err, out)
	}
	total, failed := scoreCh4(t, bin)
	if total != 100 || len(failed) > 0 {
		t.Fatalf("reference scored %d/100, failing %v; want 100 and nothing failing", total, failed)
	}
}

func TestCh4PointsSumTo100(t *testing.T) {
	res := &Ch4Result{}
	sum := 0
	seen := map[string]bool{}
	for _, c := range Ch4Evaluate(res) {
		if seen[c.ID] {
			t.Errorf("duplicate check id %q", c.ID)
		}
		seen[c.ID] = true
		sum += c.Points
		if c.Points <= 0 {
			t.Errorf("check %q carries %d points; a zero-point check is invisible to the deletion audit", c.ID, c.Points)
		}
	}
	if sum != 100 {
		t.Fatalf("checks sum to %d, want exactly 100", sum)
	}
	if len(seen) != 9 {
		t.Fatalf("got %d checks, want 9", len(seen))
	}
}

// TestCh4DeletionAudit: delete each protected behavior from the REFERENCE
// and confirm the score drops by exactly the checks that should notice.
func TestCh4DeletionAudit(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and grades one mutant per case")
	}
	names := map[string]bool{}
	for _, m := range ch4Mutants() {
		if names[m.name] {
			t.Fatalf("duplicate mutant name %q", m.name)
		}
		names[m.name] = true
	}
	for _, m := range ch4Mutants() {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			bin := buildCh4Mutant(t, m)
			total, failed := scoreCh4(t, bin)
			t.Logf("AUDIT %-40s score=%3d/100 failing=%v", m.name, total, failed)

			want := append([]string{}, m.wantFail...)
			sort.Strings(want)
			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Fatalf("%s\nwhy: %s\nfailing checks = %v\nwant            = %v\nscore %d/100",
					m.name, m.why, failed, want, total)
			}
			if len(want) == 0 {
				t.Fatalf("%s: every Chapter 4 mutant deletes a graded behavior; an empty wantFail is a mistake here", m.name)
			}
			if total >= 100 {
				t.Fatalf("%s: score did not drop (%d/100). A check that cannot fail is not a check.", m.name, total)
			}
		})
	}
}
