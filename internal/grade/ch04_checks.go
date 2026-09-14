package grade

// Chapter 4 checks: judge the evidence the harness recorded.
//
// Every check reads BEHAVIOR: what the model was sent (request bodies), what
// the agent persisted (the event log, through `dump`), what is on disk
// (cr/io/<handle>), whether a process the fixture started is still alive, and
// how long the session took. No check reads student source.
//
// Contract the checks rely on, which the chapter states so that no student
// has to guess it:
//   - every call that is not wait_for_job / send_input / kill_job / tool_limits
//     is a job; handles are integers starting at 1, +1 per job, per process;
//   - tool_called and tool_returned carry tool.job = {handle, status,
//     output:{kind:"handle", locator:"cr/io/<handle>"}, bytes, exit_code?};
//     status is one of running | done | killed;
//   - cr/io/<handle> holds the job's full result text;
//   - kill_job and Shutdown record job_killed with job.reason.

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Ch4Evaluate(res *Ch4Result) []Check {
	return []Check{
		checkCh3Parity(res),
		checkJobModel(res),
		checkWaitJob(res),
		checkSendInput(res),
		checkDebugger(res),
		checkKillJob(res),
		checkBigOutput(res),
		checkToolLimits(res),
		checkShutdown(res),
	}
}

// sessionOK is the common preamble: the scenario must have run at all.
func sessionOK(c *Check, s *Ch4Session, what string) bool {
	if s == nil || s.Ch3Session == nil {
		c.failf("the %s scenario never ran", what)
		return false
	}
	if s.Err != "" {
		c.failf("the %s scenario could not run: %s", what, s.Err)
		return false
	}
	if s.LogErr != "" {
		c.failf("the %s scenario's log could not be read back: %s", what, s.LogErr)
		return false
	}
	if len(s.Log) == 0 {
		c.failf("the %s scenario left no persisted log; `dump` printed nothing usable", what)
		return false
	}
	return true
}

// wire returns the tool_result text for a call id as the MODEL received it.
func wire(c *Check, s *Ch4Session, id string) (WireResult, bool) {
	r, ok := resultFor(s.Ch3Session, id)
	if !ok {
		c.failf("%s: no tool_result for this call ever reached the model", id)
	}
	return r, ok
}

var (
	runningRe = regexp.MustCompile(`(?i)\brunning\b`)
	killedRe  = regexp.MustCompile(`(?i)\bkilled\b`)
	// completedRe is what a report must NOT say about a killed job. The words
	// are the ones a status line uses for a normal end; the exit-code spelling
	// is included because "exit_code -1" is how a killed process's goroutine
	// describes the death if nobody told it the job was already over.
	completedRe = regexp.MustCompile(`(?i)\b(done|finished|completed|succeeded)\b|exit_code`)
)

// --- ch3parity ---------------------------------------------------------------

func checkCh3Parity(res *Ch4Result) Check {
	c := Check{ID: "ch3parity", Title: "Chapter 3's six tools and the tool loop still work",
		Points: 10, Passed: true, Earned: 10}
	if res.Ch3Err != "" {
		c.failf("Chapter 3's harness could not run this binary: %s", res.Ch3Err)
		return c
	}
	if len(res.Ch3) == 0 {
		c.failf("no Chapter 3 checks ran, so parity was never actually tested")
		return c
	}
	var failed []string
	for _, ch := range res.Ch3 {
		if !ch.Passed {
			failed = append(failed, fmt.Sprintf("%s (%d/%d)", ch.ID, ch.Earned, ch.Points))
		}
	}
	if len(failed) > 0 {
		c.failf("turning tool calls into jobs broke Chapter 3: %s", joined(failed))
		return c
	}
	c.Details = append(c.Details, fmt.Sprintf("all %d Chapter 3 checks still pass", len(res.Ch3)))
	return c
}

// --- jobmodel ----------------------------------------------------------------

// checkJobModel: a handle for EVERY tool, issued before the tool ran, and the
// full result on disk afterwards.
//
// Two of the three jobs are read_file and list_directory. That is the
// negative control the outline demands: a student who special-cased the
// shell scores zero here, not seventeen out of twenty-five.
func checkJobModel(res *Ch4Result) Check {
	c := Check{ID: "jobmodel", Title: "handle at the dispatch site for every tool; result retrievable after the call",
		Points: 25, Passed: true, Earned: 25}
	s := res.JobModel
	if !sessionOK(&c, s, "jobmodel") {
		return c
	}
	called, returned := logJobs(s)

	jobs := []struct{ id, tool string }{
		{"toolu_jm_read", "read_file"},
		{"toolu_jm_ls", "list_directory"},
		{"toolu_jm_sh", "run_command"},
	}
	seen := map[int]string{}
	for _, j := range jobs {
		cj, rj := called[j.id], returned[j.id]
		if !cj.Present {
			c.failf("%s: tool_called has no job record — the handle must exist BEFORE the tool runs", j.tool)
			continue
		}
		if cj.Handle <= 0 {
			c.failf("%s: tool_called job has no positive handle", j.tool)
			continue
		}
		if prev, dup := seen[cj.Handle]; dup {
			c.failf("%s and %s were both given handle %d", prev, j.tool, cj.Handle)
		}
		seen[cj.Handle] = j.tool
		if !rj.Present {
			c.failf("%s: tool_returned has no job record", j.tool)
			continue
		}
		if rj.Handle != cj.Handle {
			c.failf("%s: tool_called says handle %d, tool_returned says %d", j.tool, cj.Handle, rj.Handle)
		}
		if rj.Status != "done" {
			c.failf("%s: a fast local tool should be done when its call returns; status is %q", j.tool, rj.Status)
		}
		if rj.Locator == "" {
			c.failf("%s: job record names no output locator", j.tool)
			continue
		}
		file, ok := ioFile(s, rj.Locator)
		if !ok {
			c.failf("%s: job record points at %s but no such file was left behind", j.tool, rj.Locator)
			continue
		}
		r, ok := wire(&c, s, j.id)
		if !ok {
			continue
		}
		if string(file) != r.Text {
			c.failf("%s: %s (%d bytes) is not the result the model was sent (%d bytes); the file must hold the full result",
				j.tool, rj.Locator, len(file), len(r.Text))
		}
		if rj.Bytes != len(file) {
			c.failf("%s: job record says %d bytes, %s holds %d", j.tool, rj.Bytes, rj.Locator, len(file))
		}
	}
	for _, want := range []int{1, 2, 3} {
		if _, ok := seen[want]; !ok {
			c.failf("handles must start at 1 and count up by one per job; handle %d was never issued", want)
			break
		}
	}

	// The wait is not a job, and it must still answer.
	if rj := returned["toolu_jm_wait"]; rj.Present {
		c.failf("wait_for_job was given a job record of its own (handle %d); the job verbs supervise jobs, they are not jobs", rj.Handle)
	}
	if r, ok := wire(&c, s, "toolu_jm_wait"); ok && r.IsError {
		c.failf("wait_for_job on handle 1, which had finished, came back as an error: %.200q", r.Text)
	}
	if len(s.Answers) == 0 {
		c.failf("the session produced no final answer")
	}
	return c
}

// --- waitjob -----------------------------------------------------------------

func checkWaitJob(res *Ch4Result) Check {
	c := Check{ID: "waitjob", Title: "wait_for_job blocks until done and works on an already-finished job",
		Points: 10, Passed: true, Earned: 10}
	s := res.WaitJob
	if !sessionOK(&c, s, "waitjob") {
		return c
	}
	_, returned := logJobs(s)

	if rj := returned["toolu_wj_start"]; !rj.Present {
		c.failf("run_command has no job record")
	} else if rj.Status != "running" {
		c.failf("run_command with ai_callback_delay 0.2 on a one-second sleeper should have returned while it was still running; status is %q", rj.Status)
	}
	if r, ok := wire(&c, s, "toolu_wj_wait"); ok {
		if r.IsError {
			c.failf("wait_for_job returned an error: %.200q", r.Text)
		} else if !strings.Contains(r.Text, "SLEEPER_DONE") {
			c.failf("wait_for_job with a 10s delay did not wait for the sleeper to finish; SLEEPER_DONE is not in its result: %.300q", r.Text)
		}
	}
	if r, ok := wire(&c, s, "toolu_wj_again"); ok && r.IsError {
		c.failf("wait_for_job on the finished job returned an error: %.200q", r.Text)
	}
	if len(s.Answers) == 0 {
		c.failf("the session never reached its final answer — a wait on an already-finished job most likely hung")
	}
	if s.Wall > 12*time.Second {
		c.failf("the session took %s; a second wait on a finished job must return at once, not block for its delay", s.Wall.Round(time.Millisecond))
	}
	if file, ok := ioFile(s, "cr/io/1"); !ok {
		c.failf("cr/io/1 was not left behind")
	} else if !strings.Contains(string(file), "SLEEPER_DONE") || !exitCodeRe0.MatchString(string(file)) {
		c.failf("cr/io/1 should end with the sleeper's last line and its exit status; got %.200q", string(file))
	}
	return c
}

var exitCodeRe0 = regexp.MustCompile(`(?i)exit[^0-9]{0,16}0\b`)

// --- sendinput ---------------------------------------------------------------

func checkSendInput(res *Ch4Result) Check {
	c := Check{ID: "sendinput", Title: "input reaches a running process and its response comes back",
		Points: 10, Passed: true, Earned: 10}
	s := res.SendInput
	if !sessionOK(&c, s, "sendinput") {
		return c
	}
	_, returned := logJobs(s)

	start, ok := wire(&c, s, "toolu_si_start")
	if ok {
		if !strings.Contains(start.Text, "ECHO_READY") {
			c.failf("run_command with ai_callback_pattern ECHO_READY returned without the prompt in its output: %.200q", start.Text)
		}
		if strings.Contains(start.Text, "echo: hello world") {
			c.failf("the echo of the input appears BEFORE the input was sent; the fixture is broken or the output is not the process's")
		}
	}
	if rj := returned["toolu_si_start"]; rj.Present && rj.Status != "running" {
		c.failf("the echo program waits for input; its job should be running when run_command returns, not %q", rj.Status)
	}
	if r, ok := wire(&c, s, "toolu_si_hello"); ok {
		if r.IsError {
			c.failf("send_input returned an error: %.200q", r.Text)
		} else if !strings.Contains(r.Text, "echo: hello world") {
			c.failf("send_input's result does not contain the process's reply `echo: hello world`: %.300q", r.Text)
		}
	}
	if r, ok := wire(&c, s, "toolu_si_quit"); ok && r.IsError {
		c.failf("send_input quit returned an error: %.200q", r.Text)
	}
	if file, ok := ioFile(s, "cr/io/1"); !ok {
		c.failf("cr/io/1 was not left behind")
	} else if !strings.Contains(string(file), "ECHO_BYE") || !exitCodeRe0.MatchString(string(file)) {
		c.failf("after `quit` the process should have said ECHO_BYE and exited 0; cr/io/1 ends %.200q", tail(string(file), 200))
	}
	return c
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// --- debugger ----------------------------------------------------------------

// checkDebugger is the chapter's closing demonstration, graded: the fake
// drives dlv to a breakpoint and reads a variable. On pipes dlv refuses to
// start at all ("Stdin is not a terminal"), so this is the check that proves
// the terminal, the pattern wake and send_input compose into a capability
// Chapter 3 did not have.
func checkDebugger(res *Ch4Result) Check {
	c := Check{ID: "debugger", Title: "the agent drives dlv: breakpoint, continue, print, quit",
		Points: 5, Passed: true, Earned: 5}
	if res.DlvPath == "" {
		c.failf("dlv was not found on PATH or in $(go env GOPATH)/bin; Chapter 4 requires it: go install github.com/go-delve/delve/cmd/dlv@latest")
		return c
	}
	s := res.Debugger
	if !sessionOK(&c, s, "debugger") {
		return c
	}
	if r, ok := wire(&c, s, "toolu_dbg_start"); ok && !strings.Contains(r.Text, "(dlv)") {
		c.failf("run_command waiting on the pattern `(dlv) ` returned without the prompt: %.300q", r.Text)
	}
	if r, ok := wire(&c, s, "toolu_dbg_break"); ok && !strings.Contains(r.Text, "main.go:7") {
		c.failf("after `b main.go:7` the debugger did not confirm the breakpoint: %.300q", r.Text)
	}
	if r, ok := wire(&c, s, "toolu_dbg_print"); ok && !regexp.MustCompile(`\b42\b`).MatchString(r.Text) {
		c.failf("`p answer` at the breakpoint should print 42: %.300q", r.Text)
	}
	if len(s.Answers) == 0 {
		c.failf("the session never reached its final answer")
	}
	return c
}

// --- killjob -----------------------------------------------------------------

func checkKillJob(res *Ch4Result) Check {
	c := Check{ID: "killjob", Title: "kill_job stops the process and anything waiting on it is told",
		Points: 10, Passed: true, Earned: 10}
	s := res.KillJob
	if !sessionOK(&c, s, "killjob") {
		return c
	}
	_, returned := logJobs(s)

	if rj := returned["toolu_kj_start"]; !rj.Present {
		c.failf("run_command has no job record")
	} else if rj.Status != "running" {
		file, _ := ioFile(s, rj.Locator)
		c.failf("a program that never returns should still be running when run_command returns; status is %q; its output ends %.300q", rj.Status, tail(string(file), 300))
	}
	if r, ok := wire(&c, s, "toolu_kj_wait"); ok && (r.IsError || !runningRe.MatchString(r.Text)) {
		c.failf("wait_for_job with a short delay on a blocked job should report it still running: %.200q", r.Text)
	}
	if r, ok := wire(&c, s, "toolu_kj_kill"); ok && r.IsError {
		c.failf("kill_job returned an error: %.200q", r.Text)
	}
	if ev, ok := killEvents(s)[1]; !ok {
		c.failf("no job_killed event for handle 1 in the log; the kill must be legible after the fact")
	} else {
		if !strings.Contains(strings.ToLower(ev.Reason), "kill") {
			c.failf("job_killed for handle 1 has reason %q; want it to name kill_job", ev.Reason)
		}
		if ev.Status != "killed" {
			c.failf("job_killed for handle 1 records the job's status as %q, not killed; the status is the claim the log makes, and it must not go on to say done", ev.Status)
		}
	}
	if r, ok := wire(&c, s, "toolu_kj_after"); ok {
		if r.IsError {
			c.failf("wait_for_job after the kill returned an error: %.200q", r.Text)
		} else if !killedRe.MatchString(r.Text) {
			c.failf("wait_for_job after the kill must say the job was killed: %.200q", r.Text)
		} else if completedRe.MatchString(r.Text) {
			c.failf("wait_for_job after the kill reports the job as having finished normally; a killed job did not finish: %.300q", r.Text)
		}
	}
	if s.Wall > 15*time.Second {
		c.failf("the session took %s: the 30-second wait on the killed job was not told and blocked for its full delay", s.Wall.Round(time.Millisecond))
	}
	if pid, ok := s.Pids["blocker.pid"]; !ok {
		c.failf("the blocker never wrote its pid file; it did not start, so nothing was killed")
	} else if s.Alive["blocker.pid"] {
		c.failf("pid %d was still alive after the session; kill_job did not stop the process", pid)
	}
	return c
}

// --- bigoutput ---------------------------------------------------------------

// checkBigOutput: the whole thing on disk, a stub in the context, and the
// truncation done by the dispatcher rather than by the tool that happened to
// produce it. The read_file leg is the negative control for the last clause.
func checkBigOutput(res *Ch4Result) Check {
	c := Check{ID: "bigoutput", Title: "full output on disk, stub with byte count and path in context, truncation at dispatch",
		Points: 15, Passed: true, Earned: 15}
	s := res.BigOutput
	if !sessionOK(&c, s, "bigoutput") {
		return c
	}
	_, returned := logJobs(s)
	const inlineLimit = 24 * 1024 // 16 KiB cap plus headroom for the stub's own text

	check := func(id, tool string, wantFile func([]byte) string, limit int) {
		rj := returned[id]
		if !rj.Present {
			c.failf("%s: no job record", tool)
			return
		}
		if rj.Status != "done" {
			c.failf("%s: status %q; the job should have finished within its delay", tool, rj.Status)
		}
		file, ok := ioFile(s, rj.Locator)
		if !ok {
			c.failf("%s: no output file at %q", tool, rj.Locator)
			return
		}
		if msg := wantFile(file); msg != "" {
			c.failf("%s: %s: %s", tool, rj.Locator, msg)
		}
		r, ok := wire(&c, s, id)
		if !ok {
			return
		}
		if len(r.Text) > limit {
			c.failf("%s: %d bytes reached the model; the inline result must be cut to the cap (%d) with the rest on disk", tool, len(r.Text), limit)
		}
		if !strings.Contains(r.Text, rj.Locator) {
			c.failf("%s: the stub does not name the path %s where the full output is", tool, rj.Locator)
		}
		if !strings.Contains(r.Text, strconv.Itoa(len(file))) {
			c.failf("%s: the stub does not state the total byte count (%d)", tool, len(file))
		}
	}

	floodOK := func(file []byte) string {
		if len(file) < 1<<20 {
			return fmt.Sprintf("only %d bytes on disk; the flood emits more than a megabyte", len(file))
		}
		if !strings.Contains(string(file), "FLOOD line 000000") || !strings.Contains(string(file), fmt.Sprintf("FLOOD line %06d", ch4FloodLines-1)) {
			return "the first or last flood line is missing; the file must hold ALL of the output"
		}
		return ""
	}
	check("toolu_bo_flood", "run_command", floodOK, inlineLimit)
	check("toolu_bo_read", "read_file", func(file []byte) string {
		if string(file) != string(s.Big) {
			return fmt.Sprintf("%d bytes on disk, big.txt is %d; read_file's full result must be on disk even though only a stub entered the context", len(file), len(s.Big))
		}
		return ""
	}, inlineLimit)
	check("toolu_bo_small", "run_command", floodOK, 4096)
	return c
}

// --- toollimits --------------------------------------------------------------

func checkToolLimits(res *Ch4Result) Check {
	c := Check{ID: "toollimits", Title: "tool_limits applies to exactly the next call; defaults otherwise; explicit args win",
		Points: 10, Passed: true, Earned: 10}
	s := res.ToolLimits
	if !sessionOK(&c, s, "toollimits") {
		return c
	}
	_, returned := logJobs(s)

	for _, id := range []string{"toolu_tl_set", "toolu_tl_setpat", "toolu_tl_set2"} {
		if rj := returned[id]; rj.Present {
			c.failf("%s: tool_limits was given a job (handle %d); it is not a job", id, rj.Handle)
		}
		if r, ok := wire(&c, s, id); ok && r.IsError {
			c.failf("%s: tool_limits returned an error: %.200q", id, r.Text)
		}
	}
	status := func(id string) string {
		rj := returned[id]
		if !rj.Present {
			c.failf("%s: no job record", id)
			return ""
		}
		return rj.Status
	}
	if st := status("toolu_tl_run1"); st != "" && st != "running" {
		c.failf("after tool_limits{ai_callback_delay:0.2}, run_command on a one-second sleeper should return while running; status %q — the pending limits were not applied", st)
	}
	if st := status("toolu_tl_run2"); st != "" && st != "done" {
		c.failf("the SECOND run_command, with no tool_limits before it, should get the 3s default and find the sleeper done; status %q — tool_limits must be one-shot", st)
	}
	if st := status("toolu_tl_run3"); st != "" && st != "running" {
		c.failf("after tool_limits{ai_callback_pattern:SLEEPER_START}, run_command should wake on the pattern while the sleeper runs; status %q", st)
	}
	if r, ok := wire(&c, s, "toolu_tl_run3"); ok {
		if !strings.Contains(r.Text, "SLEEPER_START") {
			c.failf("the pattern wake should return the output that matched; SLEEPER_START is not in the result")
		}
		if strings.Contains(r.Text, "SLEEPER_DONE") {
			c.failf("the pattern did not wake the call early: SLEEPER_DONE is already in the result")
		}
	}
	if st := status("toolu_tl_run4"); st != "" && st != "done" {
		c.failf("an explicit ai_callback_delay of 10 on the call must override the pending tool_limits delay of 0.2; status %q", st)
	}
	return c
}

// --- shutdown ----------------------------------------------------------------

func checkShutdown(res *Ch4Result) Check {
	c := Check{ID: "shutdown", Title: "at exit, running jobs are killed and the log says so",
		Points: 5, Passed: true, Earned: 5}
	s := res.Shutdown
	if !sessionOK(&c, s, "shutdown") {
		return c
	}
	_, returned := logJobs(s)
	if rj := returned["toolu_sd_start"]; !rj.Present {
		c.failf("run_command has no job record")
	} else if rj.Status != "running" {
		file, _ := ioFile(s, rj.Locator)
		c.failf("the blocker should still be running when run_command returns; status %q; its output ends %.300q", rj.Status, tail(string(file), 300))
	}
	if ev, ok := killEvents(s)[1]; !ok {
		c.failf("no job_killed event for handle 1; a job killed at exit must be recorded, or the transcript shows a process that simply vanished")
	} else {
		if !strings.Contains(strings.ToLower(ev.Reason), "shutdown") {
			c.failf("job_killed for handle 1 has reason %q; want it to say shutdown", ev.Reason)
		}
		if ev.Status != "killed" {
			c.failf("job_killed at shutdown records the job's status as %q, not killed", ev.Status)
		}
	}
	if pid, ok := s.Pids["shutdown.pid"]; !ok {
		c.failf("the blocker never wrote its pid file; nothing was left running to shut down")
	} else if s.Alive["shutdown.pid"] {
		c.failf("pid %d outlived the agent; Shutdown did not kill it", pid)
	}
	if len(s.Answers) == 0 {
		c.failf("the session never reached its final answer")
	}
	return c
}
