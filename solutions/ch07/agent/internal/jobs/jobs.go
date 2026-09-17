package jobs

// Jobs: a tool call is a job you start and supervise, for EVERY tool.
//
// The handle is allocated at the dispatch site, before the dispatcher knows
// or cares which tool it is. That is the whole design: `read_file` on an NFS
// mount that has gone away hangs exactly as well as a shell command does, and
// the freeze that motivated this chapter was a screenshot, not a subprocess.
//
// Nothing here dies on its own. A job that outlives the caller's patience is
// still running, with a handle, and the model decides what to do about it:
// wait some more, talk to it, or kill it. The default wake-up is three
// seconds, and that is short on purpose — it costs nothing, because nothing is
// cancelled; it only hands control back to the model so it can look.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/waywardgeek/coding-agents-course/agent/internal/common"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// IODir is where every job's full output lives, as cr/io/<handle>. The path
// is not decoration: it is the recovery route, and the model already has a
// tool that can read a range of it.
const IODir = "cr/io"

// --- the table -------------------------------------------------------------

type Jobs struct {
	host common.Host
	mu   sync.Mutex
	next int
	all  map[int]*Job
	// pending is what `tool_limits` set for the next call. One-shot: Take
	// clears it, and it is taken by the next call no matter which tool.
	pending *common.Limits
}

func NewJobs(host common.Host) *Jobs { return &Jobs{host: host, all: map[int]*Job{}} }

// Logf satisfies common.Host via the parent chain.
func (js *Jobs) Logf(format string, args ...any)    { js.host.Logf(format, args...) }
func (js *Jobs) APILogf(format string, args ...any) { js.host.APILogf(format, args...) }
func (js *Jobs) Debugf(format string, args ...any)  { js.host.Debugf(format, args...) }

// Start allocates a handle and its output file. It is called by the
// dispatcher for every job-creating tool before the tool runs.
func (js *Jobs) Start(tool, callID string) (common.JobHandle, error) {
	if err := os.MkdirAll(IODir, 0o755); err != nil {
		return nil, fmt.Errorf("job output dir: %w", err)
	}
	js.mu.Lock()
	js.next++
	h := js.next
	js.mu.Unlock()

	path := filepath.Join(IODir, strconv.Itoa(h))
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("job output file: %w", err)
	}
	j := &Job{
		Handle: h, Tool: tool, CallID: callID, Started: time.Now(), Path: path,
		file: f, status: common.StatusRunning, changed: make(chan struct{}),
	}
	js.mu.Lock()
	js.all[h] = j
	js.mu.Unlock()
	return j, nil
}

func (js *Jobs) Get(h int) (common.JobHandle, bool) {
	js.mu.Lock()
	defer js.mu.Unlock()
	j, ok := js.all[h]
	return j, ok
}

// Handles lists every handle issued so far, so an error about a bad handle
// can say what the good ones are.
func (js *Jobs) Handles() []int {
	js.mu.Lock()
	defer js.mu.Unlock()
	out := make([]int, 0, len(js.all))
	for h := range js.all {
		out = append(out, h)
	}
	sort.Ints(out)
	return out
}

// Running returns the jobs still running, lowest handle first.
func (js *Jobs) Running() []common.JobHandle {
	var out []common.JobHandle
	for _, h := range js.Handles() {
		if j, _ := js.Get(h); j != nil && j.Status() == common.StatusRunning {
			out = append(out, j)
		}
	}
	return out
}

func (js *Jobs) SetNext(l common.Limits) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.pending = &l
}

// Take returns the limits for the next call: defaults, then the pending
// tool_limits if any (consumed), then the call's own arguments. The bool
// says whether a pending tool_limits was consumed; the caller puts that in
// the report, so limits that land on the wrong call are seen, not suffered.
func (js *Jobs) Take(args json.RawMessage) (common.Limits, bool, error) {
	l := common.DefaultLimits()
	consumed := false
	js.mu.Lock()
	if js.pending != nil {
		l = *js.pending
		js.pending = nil
		consumed = true
	}
	js.mu.Unlock()
	a, err := common.LimitsInArgs(args)
	if err != nil {
		return l, consumed, err
	}
	l, err = l.Overlay(a)
	return l, consumed, err
}

// --- one job ---------------------------------------------------------------

type Job struct {
	Handle  int
	Tool    string
	CallID  string
	Started time.Time
	Path    string

	mu     sync.Mutex
	out    bytes.Buffer // everything the job has produced, in order
	file   *os.File     // the same bytes, on disk, as they arrive
	status common.JobStatus
	err    error // a tool error, when the tool returned one
	exit   *int  // process exit code, for jobs that are processes
	// cursor is how many bytes the model has already been shown. Every
	// report advances it, so no report repeats bytes and the file is the
	// only place the whole output exists.
	cursor int
	// changed is closed and replaced on every change; a waiter blocks on it.
	changed chan struct{}

	proc  *os.Process // set by tools that start a process; nil otherwise
	stdin interface {
		Write([]byte) (int, error)
	}
	// cwd is where a run_command job ran, when that was not the working
	// directory. Set once by the tool, before the process starts.
	cwd string
}

// Write appends output. It is the io.Writer the tool's process writes into,
// and it is safe to call from the tool's goroutine while the dispatcher waits.
func (j *Job) Write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.out.Write(p)
	if j.file != nil {
		_, _ = j.file.Write(p)
	}
	j.broadcast()
	return len(p), nil
}

// broadcast wakes every waiter. Caller holds j.mu.
func (j *Job) broadcast() {
	close(j.changed)
	j.changed = make(chan struct{})
}

// Attach registers the job's process so send_input and kill_job can reach
// it. stdin may be nil for a process that takes no input.
func (j *Job) Attach(proc *os.Process, stdin interface{ Write([]byte) (int, error) }) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.proc, j.stdin = proc, stdin
}

// SetExit records how a process ended. It does not end the job; Finish does.
func (j *Job) SetExit(code int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.exit = &code
}

// Finish is called exactly once, by the dispatcher, when the tool returns.
// A non-streaming tool hands back its whole result here; a streaming one has
// already written everything and hands back "".
//
// A job that was killed while the tool was still running stays killed: the
// tool's late answer is not appended, because the model was already told the
// job ended and a result that arrives after that is a second, contradictory
// truth.
func (j *Job) Finish(result string, err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status == common.StatusRunning {
		text := result
		if err != nil {
			j.err = err
			text = err.Error()
		}
		if text != "" {
			j.out.WriteString(text)
			if j.file != nil {
				_, _ = j.file.WriteString(text)
			}
		}
		j.status = common.StatusDone
	}
	if j.file != nil {
		_ = j.file.Close()
		j.file = nil
	}
	j.broadcast()
}

func (j *Job) Status() common.JobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.status
}

func (j *Job) Err() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.err
}

func (j *Job) Bytes() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.out.Len()
}

// Wait blocks until the job is no longer running, or l.Delay passes, or the
// output the model has not yet seen matches l.Pattern — whichever first.
//
// It works on a job that has already finished, and returns at once. A student
// who implements this as "block on the channel" hangs forever on a completed
// job and diagnoses it as a deadlock in their own code.
func (j *Job) Wait(l common.Limits) common.WakeReason {
	timer := time.NewTimer(l.Delay)
	defer timer.Stop()
	for {
		j.mu.Lock()
		if j.status != common.StatusRunning {
			j.mu.Unlock()
			return common.WokeDone
		}
		if l.Pattern != nil && l.Pattern.Match(j.out.Bytes()[j.cursor:]) {
			j.mu.Unlock()
			return common.WokePattern
		}
		ch := j.changed
		j.mu.Unlock()
		select {
		case <-ch:
		case <-timer.C:
			return common.WokeDelay
		}
	}
}

// Kill ends a running job. For a process it kills the whole process group,
// so a `sh -c` wrapper cannot leave its child behind. For a tool with no
// process there is nothing to kill: Go cannot stop a goroutine, so the job
// is marked killed, the dispatcher stops listening, and the report says so.
//
// Returns false if the job had already ended.
func (j *Job) Kill(reason string) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status != common.StatusRunning {
		return false
	}
	if j.proc != nil {
		_ = syscall.Kill(-j.proc.Pid, syscall.SIGKILL)
	}
	j.status = common.StatusKilled
	note := fmt.Sprintf("\n[job %d killed: %s]\n", j.Handle, reason)
	j.out.WriteString(note)
	if j.file != nil {
		_, _ = j.file.WriteString(note)
	}
	j.broadcast()
	return true
}

// SendInput writes to the process's stdin. It is an error on a job that has
// ended or that has no process, and both errors say which.
func (j *Job) SendInput(text string) error {
	j.mu.Lock()
	st, w := j.status, j.stdin
	j.mu.Unlock()
	if st != common.StatusRunning {
		return fmt.Errorf("job %d is %s; it is not reading input", j.Handle, st)
	}
	if w == nil {
		return fmt.Errorf("job %d has no stdin: %s is not a process", j.Handle, j.Tool)
	}
	_, err := w.Write([]byte(text))
	return err
}

// Data is the job as the event log records it.
// HasProcess reports whether the job has a running process.
func (j *Job) HasProcess() bool {
	return j.proc != nil
}

func (j *Job) Data() *common.JobData {
	j.mu.Lock()
	defer j.mu.Unlock()
	return &common.JobData{
		Handle:   j.Handle,
		Tool:     j.Tool,
		Status:   j.status,
		Output:   common.Ref{Kind: common.RefHandle, Locator: j.Path},
		Bytes:    j.out.Len(),
		ExitCode: j.exit,
		Cwd:      j.cwd,
	}
}

// SetCwd records the directory a job's process ran in. run_command calls it
// before the process starts and only when a cwd other than the working
// directory was asked for, so the record is the exception, not the rule.
func (j *Job) SetCwd(dir string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cwd = dir
}

// Report renders what the model sees after a wait, and advances the cursor
// past it. Truncation to l.MaxOutput happens HERE, on the way into the
// context, with the full text already on disk.
//
// A job that finished within the delay, on its first report, with a result
// that fits, is reported as the bare result — byte for byte what Chapter 3
// returned. Everything else gets a status line, because the model needs to
// know whether it is looking at all of the answer or some of it.
func (j *Job) Report(reason common.WakeReason, l common.Limits) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	all := j.out.Bytes()
	unseen := string(all[j.cursor:])
	first := j.cursor == 0
	j.cursor = len(all)

	body := capText(unseen, l.MaxOutput, len(all), j.Path)
	if reason == common.WokeDone && j.status == common.StatusDone && first {
		return body
	}

	var b strings.Builder
	switch j.status {
	case common.StatusDone:
		if j.err != nil {
			fmt.Fprintf(&b, "job %d done with error.", j.Handle)
		} else if j.exit != nil {
			fmt.Fprintf(&b, "job %d done, exit_code %d.", j.Handle, *j.exit)
		} else {
			fmt.Fprintf(&b, "job %d done.", j.Handle)
		}
	case common.StatusKilled:
		fmt.Fprintf(&b, "job %d killed.", j.Handle)
	default:
		switch reason {
		case common.WokePattern:
			fmt.Fprintf(&b, "job %d still running; ai_callback_pattern %q matched.", j.Handle, l.Pattern.String())
		default:
			fmt.Fprintf(&b, "job %d still running after %s.", j.Handle, l.Delay)
		}
	}
	if unseen == "" {
		fmt.Fprintf(&b, " no new output; %d bytes total at %s\n", len(all), j.Path)
	} else {
		fmt.Fprintf(&b, " new output (%d bytes of %d total at %s):\n%s", len(unseen), len(all), j.Path, body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteString("\n")
		}
	}
	if j.status == common.StatusRunning {
		fmt.Fprintf(&b, "[wait_for_job(%d) to keep waiting, send_input(%d, ...) to talk to it, kill_job(%d) to stop it]\n",
			j.Handle, j.Handle, j.Handle)
	}
	return b.String()
}

// capText keeps the first and last halves of max bytes, each cut back to a
// line boundary, and says how much is missing and where all of it is.
func capText(s string, max int, total int, path string) string {
	if len(s) <= max {
		return s
	}
	half := max / 2
	head := s[:half]
	if i := strings.LastIndexByte(head, '\n'); i >= 0 {
		head = head[:i+1]
	}
	tail := s[len(s)-half:]
	if i := strings.IndexByte(tail, '\n'); i >= 0 {
		tail = tail[i+1:]
	}
	omitted := len(s) - len(head) - len(tail)
	return head + fmt.Sprintf("[... %d bytes omitted; full output (%d bytes) at %s ...]\n", omitted, total, path) + tail
}
