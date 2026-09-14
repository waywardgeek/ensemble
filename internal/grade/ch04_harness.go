package grade

// Chapter 4 harness: drive the submission through jobs, recording evidence and
// judging nothing.
//
// Everything a check will ask about is captured here: the request bodies the
// fake saw, the persisted event log, every cr/io/<handle> file the agent
// wrote, the pids of the helpers that were told to never exit, and whether
// those pids were still alive after the agent exited.
//
// The helpers are Go programs built to BINARIES once per run and invoked by
// absolute path. Not `go run`: it rewrites exit statuses and prints its own
// diagnostics, and Chapter 3 already paid for learning that.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/waywardgeek/coding-agents-course/internal/fakevendor"
)

// --- helpers the fixtures run ---------------------------------------------

const (
	// sleeper SECONDS: prints a start marker, sleeps, prints a done marker.
	// Outlives a short delay and finishes on its own: the job you wait for.
	ch4Sleeper = `package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	d, _ := strconv.ParseFloat(os.Args[1], 64)
	fmt.Println("SLEEPER_START")
	time.Sleep(time.Duration(d * float64(time.Second)))
	fmt.Println("SLEEPER_DONE")
}
`

	// blocker PIDFILE: writes its pid and never returns. This is the
	// screenshot incident, reproducible. If it exits on its own, killjob and
	// shutdown are graded by nothing.
	ch4Blocker = `package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	_ = os.WriteFile(os.Args[1], []byte(strconv.Itoa(os.Getpid())), 0o644)
	fmt.Println("BLOCKER_UP")
	// Not select{}: with no other goroutine the Go runtime calls that a
	// deadlock and exits 2, and a "never returns" helper that returns grades
	// killjob and shutdown by nothing. Measured, the first time this ran.
	for {
		time.Sleep(time.Hour)
	}
}
`

	// echoer: prompts, reads a line, answers it, prompts again. Exits on
	// "quit". Something real for send_input to talk to.
	ch4Echoer = `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("ECHO_READY")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := sc.Text()
		if line == "quit" {
			fmt.Println("ECHO_BYE")
			return
		}
		fmt.Printf("echo: %s\nECHO_READY\n", line)
	}
}
`

	// flood: more than a megabyte, deterministic, with a known first and last
	// line so the file on disk can be checked for completeness.
	ch4Flood = `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	w := bufio.NewWriter(os.Stdout)
	pad := strings.Repeat("x", 48)
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(w, "FLOOD line %06d %s\n", i, pad)
	}
	w.Flush()
}
`

	// dbg: the program the debugger scenario stops inside. Line 7 is the
	// Println; `answer` is in scope there and equals 42.
	ch4Dbg = `package main

import "fmt"

func main() {
	answer := 42
	fmt.Println("answer is", answer)
}
`
	ch4FloodLines = 20000
)

// ch4Helpers builds the helpers once and returns their absolute paths.
func ch4Helpers() (map[string]string, string, error) {
	dir, err := os.MkdirTemp("", "ch04-helpers-")
	if err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module ch04helpers\n\ngo 1.21\n"), 0o644); err != nil {
		return nil, dir, err
	}
	srcs := map[string]string{"sleeper": ch4Sleeper, "blocker": ch4Blocker, "echoer": ch4Echoer, "flood": ch4Flood}
	out := map[string]string{}
	for name, src := range srcs {
		pkg := filepath.Join(dir, name)
		if err := os.MkdirAll(pkg, 0o755); err != nil {
			return nil, dir, err
		}
		if err := os.WriteFile(filepath.Join(pkg, "main.go"), []byte(src), 0o644); err != nil {
			return nil, dir, err
		}
		bin := filepath.Join(dir, "bin", name)
		cmd := exec.Command("go", "build", "-o", bin, "./"+name)
		cmd.Dir = dir
		if b, err := cmd.CombinedOutput(); err != nil {
			return nil, dir, fmt.Errorf("build helper %s: %v\n%s", name, err, b)
		}
		out[name] = bin
	}
	return out, dir, nil
}

// findDlv locates the debugger: PATH first, then $(go env GOPATH)/bin, which
// is where `go install` puts it and which is very often not on PATH.
func findDlv() string {
	if p, err := exec.LookPath("dlv"); err == nil {
		return p
	}
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return ""
	}
	p := filepath.Join(strings.TrimSpace(string(out)), "bin", "dlv")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return ""
}

// --- the workspace ----------------------------------------------------------

// ch4Workspace is Chapter 3's workspace plus a file bigger than the inline
// cap and the debugger's target program.
func ch4Workspace() (string, error) {
	work, err := ch3Workspace()
	if err != nil {
		return "", err
	}
	var big strings.Builder
	for i := 0; i < 16000; i++ {
		fmt.Fprintf(&big, "BIG line %05d %s\n", i, strings.Repeat("y", 56))
	}
	files := map[string]string{
		"big.txt":              big.String(),
		"testdata/dbg/main.go": ch4Dbg,
	}
	for name, body := range files {
		full := filepath.Join(work, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return "", err
		}
	}
	return work, nil
}

// --- evidence ---------------------------------------------------------------

// Ch4Session is a Chapter 3 session plus what jobs leave behind.
type Ch4Session struct {
	*Ch3Session
	// IO is every cr/io/<name> file the agent wrote, read after it exited.
	IO map[string][]byte
	// Big is big.txt as planted, for the read_file negative control.
	Big []byte
	// Pids are the pids the blocker helper wrote, keyed by pidfile name, and
	// Alive says whether each was still running after the agent exited.
	Pids  map[string]int
	Alive map[string]bool
	// Wall is how long the agent process ran.
	Wall time.Duration
}

type Ch4Result struct {
	Ch3    []Check // Chapter 3's checks, re-run unchanged against this binary
	Ch3Err string

	JobModel   *Ch4Session
	WaitJob    *Ch4Session
	SendInput  *Ch4Session
	Debugger   *Ch4Session
	KillJob    *Ch4Session
	BigOutput  *Ch4Session
	ToolLimits *Ch4Session
	Shutdown   *Ch4Session

	DlvPath    string // "" if the debugger could not be found
	HelpersErr string
}

// --- scripted runs ------------------------------------------------------------

func runCh4Session(bin, name string, replies []fakevendor.Reply, pidfiles []string) *Ch4Session {
	s := &Ch4Session{Ch3Session: &Ch3Session{Name: name, Vendor: "anthropic",
		Before: map[string]string{}, After: map[string]string{}},
		IO: map[string][]byte{}, Pids: map[string]int{}, Alive: map[string]bool{}}

	work, err := ch4Workspace()
	if err != nil {
		s.Err = err.Error()
		return s
	}
	defer os.RemoveAll(work)
	s.Big, _ = os.ReadFile(filepath.Join(work, "big.txt"))

	// The pidfile paths are absolute and inside the workspace; replies that
	// mention them are written with the placeholder WORK.
	for i := range replies {
		replies[i].ToolArgs = strings.ReplaceAll(replies[i].ToolArgs, "WORK", work)
	}

	fake := fakevendor.New(replies)
	defer fake.Close()

	logPath := filepath.Join(work, ".ch04-"+name+".log")
	env := ch3Env("anthropic", fake.URL(), work, logPath)

	start := time.Now()
	stdout, stderr, _ := runWithStdin(bin, work, env, nil, []string{`{"user":"do the work"}`})
	s.Wall = time.Since(start)
	s.Stdout, s.Stderr = stdout, stderr
	parseCh3Stdout(s.Ch3Session, stdout)
	s.Requests = fake.Requests()

	// Liveness is checked BEFORE anything else touches the pids, and then
	// whatever is still alive is killed here so a broken submission cannot
	// leak a process per grading run.
	for _, pf := range pidfiles {
		b, err := os.ReadFile(filepath.Join(work, pf))
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			continue
		}
		s.Pids[pf] = pid
		s.Alive[pf] = syscall.Kill(pid, 0) == nil
		if s.Alive[pf] {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}

	dumpOut, _, _ := runOnce(bin, work, env, "dump")
	s.DumpOut = dumpOut
	s.Log, s.LogErr = parseLogLines(dumpOut)

	if entries, err := os.ReadDir(filepath.Join(work, "cr", "io")); err == nil {
		for _, e := range entries {
			if b, err := os.ReadFile(filepath.Join(work, "cr", "io", e.Name())); err == nil {
				s.IO[e.Name()] = b
			}
		}
	}
	return s
}

// --- the scenarios --------------------------------------------------------------

func ch4step(text, name, args, id string) fakevendor.Reply {
	return fakevendor.Reply{Text: text, ToolName: name, ToolArgs: args, ToolID: id,
		Usage: fakevendor.Canonical{Input: 20, Output: 10}}
}

func ch4done(text string) fakevendor.Reply {
	return fakevendor.Reply{Text: text, Usage: fakevendor.Canonical{Input: 40, Output: 9}}
}

func q(s string) string { return strconv.Quote(s) }

// ch4JobModelReplies: three tools, none of them the interesting one, and then
// a wait on the first. If only run_command gets a handle, this is where it
// shows.
func ch4JobModelReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Reading the middle of the notes.", "read_file",
			`{"path":"notes.md","start_line":3,"end_line":4}`, "toolu_jm_read"),
		ch4step("Looking around.", "list_directory", `{"path":"."}`, "toolu_jm_ls"),
		ch4step("Checking the toolchain.", "run_command", `{"command":"go version"}`, "toolu_jm_sh"),
		ch4step("Waiting on the read, which already finished.", "wait_for_job",
			`{"handle":1}`, "toolu_jm_wait"),
		ch4done("Three jobs, three handles."),
	}
}

func ch4WaitJobReplies(h map[string]string) []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Starting the sleeper with a short delay.", "run_command",
			`{"command":`+q(h["sleeper"]+" 1")+`,"ai_callback_delay":0.2}`, "toolu_wj_start"),
		ch4step("Waiting for it to finish.", "wait_for_job",
			`{"handle":1,"ai_callback_delay":10}`, "toolu_wj_wait"),
		ch4step("Waiting again on a job that is already done.", "wait_for_job",
			`{"handle":1,"ai_callback_delay":10}`, "toolu_wj_again"),
		ch4done("The sleeper finished."),
	}
}

func ch4SendInputReplies(h map[string]string) []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Starting the echo program and waiting for its prompt.", "run_command",
			`{"command":`+q(h["echoer"])+`,"ai_callback_delay":10,"ai_callback_pattern":"ECHO_READY"}`, "toolu_si_start"),
		ch4step("Saying hello.", "send_input",
			`{"handle":1,"input":"hello world","ai_callback_delay":10,"ai_callback_pattern":"ECHO_READY"}`, "toolu_si_hello"),
		ch4step("Telling it to quit.", "send_input",
			`{"handle":1,"input":"quit","ai_callback_delay":10}`, "toolu_si_quit"),
		ch4step("Confirming it exited.", "wait_for_job",
			`{"handle":1,"ai_callback_delay":5}`, "toolu_si_wait"),
		ch4done("It echoed and exited."),
	}
}

// ch4DebuggerReplies is the chapter's closing demonstration, as a fixture:
// the same five calls the prose shows, driven by the fake, waiting on the
// debugger's prompt rather than on a guessed interval.
func ch4DebuggerReplies(dlv string) []fakevendor.Reply {
	const prompt = `\\(dlv\\) `
	return []fakevendor.Reply{
		ch4step("Starting the debugger.", "run_command",
			`{"command":`+q("PAGER=cat "+dlv+" debug ./testdata/dbg")+`,"ai_callback_delay":40,"ai_callback_pattern":"`+prompt+`"}`, "toolu_dbg_start"),
		ch4step("Setting a breakpoint.", "send_input",
			`{"handle":1,"input":"b main.go:7","ai_callback_delay":20,"ai_callback_pattern":"`+prompt+`"}`, "toolu_dbg_break"),
		ch4step("Continuing to it.", "send_input",
			`{"handle":1,"input":"c","ai_callback_delay":20,"ai_callback_pattern":"`+prompt+`"}`, "toolu_dbg_cont"),
		ch4step("Printing the variable.", "send_input",
			`{"handle":1,"input":"p answer","ai_callback_delay":20,"ai_callback_pattern":"`+prompt+`"}`, "toolu_dbg_print"),
		ch4step("Quitting.", "send_input",
			`{"handle":1,"input":"q","ai_callback_delay":10}`, "toolu_dbg_quit"),
		ch4done("answer is 42 at the breakpoint."),
	}
}

func ch4KillJobReplies(h map[string]string) []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Starting a program that never returns.", "run_command",
			`{"command":`+q(h["blocker"]+" WORK/blocker.pid")+`,"ai_callback_delay":0.5}`, "toolu_kj_start"),
		ch4step("Giving it a moment.", "wait_for_job",
			`{"handle":1,"ai_callback_delay":0.3}`, "toolu_kj_wait"),
		ch4step("It is not coming back. Killing it.", "kill_job", `{"handle":1}`, "toolu_kj_kill"),
		// A long delay on purpose: if the waiter is not told the job was
		// killed, this blocks for the full 30 seconds and the session's wall
		// clock says so.
		ch4step("Waiting on the killed job.", "wait_for_job",
			`{"handle":1,"ai_callback_delay":30}`, "toolu_kj_after"),
		ch4done("Killed and confirmed."),
	}
}

func ch4BigOutputReplies(h map[string]string) []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Running the flood.", "run_command",
			`{"command":`+q(h["flood"])+`,"ai_callback_delay":20}`, "toolu_bo_flood"),
		// NEGATIVE CONTROL for "truncation happens at dispatch": a different
		// tool, asked for the whole of a file bigger than the cap.
		ch4step("Reading the whole big file.", "read_file",
			`{"path":"big.txt","max_bytes":2000000}`, "toolu_bo_read"),
		ch4step("Running the flood with a small cap.", "run_command",
			`{"command":`+q(h["flood"])+`,"ai_callback_delay":20,"max_output_bytes":2048}`, "toolu_bo_small"),
		ch4done("Big outputs handled."),
	}
}

func ch4ToolLimitsReplies(h map[string]string) []fakevendor.Reply {
	sleep := `{"command":` + q(h["sleeper"]+" 1") + `}`
	return []fakevendor.Reply{
		ch4step("Shortening the delay for the next call.", "tool_limits",
			`{"ai_callback_delay":0.2}`, "toolu_tl_set"),
		ch4step("Running the sleeper without a delay of its own.", "run_command", sleep, "toolu_tl_run1"),
		// ONE-SHOT: the same call again must get the default back.
		ch4step("Running it again.", "run_command", sleep, "toolu_tl_run2"),
		ch4step("Setting a wake pattern for the next call.", "tool_limits",
			`{"ai_callback_pattern":"SLEEPER_START"}`, "toolu_tl_setpat"),
		ch4step("Running the sleeper; the pattern should wake me at once.", "run_command", sleep, "toolu_tl_run3"),
		ch4step("Setting a short delay, then overriding it on the call.", "tool_limits",
			`{"ai_callback_delay":0.2}`, "toolu_tl_set2"),
		ch4step("Explicit argument wins.", "run_command",
			`{"command":`+q(h["sleeper"]+" 1")+`,"ai_callback_delay":10}`, "toolu_tl_run4"),
		ch4done("Limits behaved."),
	}
}

func ch4ShutdownReplies(h map[string]string) []fakevendor.Reply {
	return []fakevendor.Reply{
		ch4step("Starting something that never returns, and leaving it.", "run_command",
			`{"command":`+q(h["blocker"]+" WORK/shutdown.pid")+`,"ai_callback_delay":0.5}`, "toolu_sd_start"),
		ch4done("Leaving it running."),
	}
}

// --- run everything ------------------------------------------------------------

func Ch4Run(bin string) (*Ch4Result, error) {
	res := &Ch4Result{}

	if r3, err := Ch3Run(bin); err != nil {
		res.Ch3Err = err.Error()
	} else {
		res.Ch3 = Ch3Evaluate(r3)
	}

	h, dir, err := ch4Helpers()
	if dir != "" {
		defer os.RemoveAll(dir)
	}
	if err != nil {
		res.HelpersErr = err.Error()
		return res, nil
	}
	res.DlvPath = findDlv()

	res.JobModel = runCh4Session(bin, "jobmodel", ch4JobModelReplies(), nil)
	res.WaitJob = runCh4Session(bin, "waitjob", ch4WaitJobReplies(h), nil)
	res.SendInput = runCh4Session(bin, "sendinput", ch4SendInputReplies(h), nil)
	if res.DlvPath != "" {
		res.Debugger = runCh4Session(bin, "debugger", ch4DebuggerReplies(res.DlvPath), nil)
	}
	res.KillJob = runCh4Session(bin, "killjob", ch4KillJobReplies(h), []string{"blocker.pid"})
	res.BigOutput = runCh4Session(bin, "bigoutput", ch4BigOutputReplies(h), nil)
	res.ToolLimits = runCh4Session(bin, "toollimits", ch4ToolLimitsReplies(h), nil)
	res.Shutdown = runCh4Session(bin, "shutdown", ch4ShutdownReplies(h), []string{"shutdown.pid"})
	return res, nil
}

// --- reading job evidence ----------------------------------------------------

// JobRecord is the job field of one tool_called / tool_returned event, as the
// student's log recorded it.
type JobRecord struct {
	Handle   int
	Status   string
	Locator  string
	Bytes    int
	ExitCode *int
	Present  bool
}

func jobOf(tool map[string]any) JobRecord {
	j, ok := tool["job"].(map[string]any)
	if !ok {
		return JobRecord{}
	}
	r := JobRecord{Present: true}
	if f, ok := j["handle"].(float64); ok {
		r.Handle = int(f)
	}
	r.Status, _ = j["status"].(string)
	if f, ok := j["bytes"].(float64); ok {
		r.Bytes = int(f)
	}
	if f, ok := j["exitcode"].(float64); ok {
		n := int(f)
		r.ExitCode = &n
	}
	if o, ok := j["output"].(map[string]any); ok {
		r.Locator, _ = o["locator"].(string)
	}
	return r
}

// logJobs returns the job record on the tool_returned event for each call id,
// and separately on the tool_called event, because the check wants to know
// the handle existed BEFORE the tool ran.
func logJobs(s *Ch4Session) (called, returned map[string]JobRecord) {
	called, returned = map[string]JobRecord{}, map[string]JobRecord{}
	for _, l := range s.Log {
		t := normName(l.Type)
		if t != "toolcalled" && t != "toolreturned" {
			continue
		}
		tool, ok := l.Data["tool"].(map[string]any)
		if !ok {
			continue
		}
		id, _ := tool["callid"].(string)
		if id == "" {
			continue
		}
		if t == "toolcalled" {
			called[id] = jobOf(tool)
		} else {
			returned[id] = jobOf(tool)
		}
	}
	return called, returned
}

// killEvents returns the job_killed events by handle → reason.
func killEvents(s *Ch4Session) map[int]string {
	out := map[int]string{}
	for _, l := range s.Log {
		if normName(l.Type) != "jobkilled" {
			continue
		}
		j, ok := l.Data["job"].(map[string]any)
		if !ok {
			continue
		}
		h, _ := j["handle"].(float64)
		reason, _ := j["reason"].(string)
		out[int(h)] = reason
	}
	return out
}

// ioFile returns the cr/io file a locator points at, tolerating either the
// bare handle or the cr/io/<handle> path.
func ioFile(s *Ch4Session, locator string) ([]byte, bool) {
	name := filepath.Base(locator)
	b, ok := s.IO[name]
	return b, ok
}
