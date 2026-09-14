package main

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
	"os"
	"path/filepath"
	"regexp"
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

// JobStatus is the whole life of a job: running, and then exactly one of done
// or killed. There is no "abandoned" and no "timed out", because nothing here
// times out. iota+1 so that the zero value is not a valid status.
type JobStatus int

const (
	StatusRunning JobStatus = iota + 1
	StatusDone
	StatusKilled
)

func (s JobStatus) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusDone:
		return "done"
	case StatusKilled:
		return "killed"
	}
	return fmt.Sprintf("JobStatus(%d)", int(s))
}

func (s JobStatus) MarshalJSON() ([]byte, error) { return json.Marshal(s.String()) }

func (s *JobStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "running":
		*s = StatusRunning
	case "done":
		*s = StatusDone
	case "killed":
		*s = StatusKilled
	default:
		return fmt.Errorf("unknown job status %q", str)
	}
	return nil
}

// Limits is what the model may set about the NEXT wait: how long to block
// before being woken with whatever there is, a pattern in the output that
// wakes it early, and how much of the result may enter the context inline.
//
// Sources, later wins: the defaults; a pending `tool_limits` call (one-shot,
// consumed by the next tool call whatever it is); the call's own arguments.
type Limits struct {
	Delay     time.Duration
	Pattern   *regexp.Regexp
	MaxOutput int
}

const (
	DefaultDelay     = 3 * time.Second
	DefaultMaxOutput = 16 * 1024
)

func DefaultLimits() Limits { return Limits{Delay: DefaultDelay, MaxOutput: DefaultMaxOutput} }

// limitArgs is the wire spelling of Limits. Every tool that waits on a job
// declares these three in its schema; `tool_limits` accepts them for the
// tools that do not.
type limitArgs struct {
	Delay     *float64 `json:"ai_callback_delay"`
	Pattern   *string  `json:"ai_callback_pattern"`
	MaxOutput *int     `json:"max_output_bytes"`
}

// overlay applies whichever of the three the arguments set.
func (l Limits) overlay(a limitArgs) (Limits, error) {
	if a.Delay != nil {
		if *a.Delay < 0 {
			return l, fmt.Errorf("ai_callback_delay must be >= 0, got %v", *a.Delay)
		}
		l.Delay = time.Duration(*a.Delay * float64(time.Second))
	}
	if a.Pattern != nil {
		if *a.Pattern == "" {
			l.Pattern = nil
		} else {
			re, err := regexp.Compile(*a.Pattern)
			if err != nil {
				return l, fmt.Errorf("ai_callback_pattern: bad pattern %q: %v", *a.Pattern, err)
			}
			l.Pattern = re
		}
	}
	if a.MaxOutput != nil {
		if *a.MaxOutput <= 0 {
			return l, fmt.Errorf("max_output_bytes must be > 0, got %d", *a.MaxOutput)
		}
		l.MaxOutput = *a.MaxOutput
	}
	return l, nil
}

// limitsInArgs peeks at a call's arguments for the three limit keys. Tools
// whose schema does not declare them will never be sent them by a model that
// reads schemas, and `tool_limits` exists for exactly that case.
func limitsInArgs(args json.RawMessage) (limitArgs, error) {
	var a limitArgs
	if len(args) == 0 {
		return a, nil
	}
	if err := json.Unmarshal(args, &a); err != nil {
		// Not this function's problem: the tool's own decoder will say so.
		return limitArgs{}, nil
	}
	return a, nil
}

// --- the table -------------------------------------------------------------

type Jobs struct {
	mu   sync.Mutex
	next int
	all  map[int]*Job
	// pending is what `tool_limits` set for the next call. One-shot: Take
	// clears it, and it is taken by the next call no matter which tool.
	pending *Limits
}

func NewJobs() *Jobs { return &Jobs{all: map[int]*Job{}} }

// Start allocates a handle and its output file. It is called by the
// dispatcher for every job-creating tool before the tool runs.
func (js *Jobs) Start(tool, callID string) (*Job, error) {
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
		file: f, status: StatusRunning, changed: make(chan struct{}),
	}
	js.mu.Lock()
	js.all[h] = j
	js.mu.Unlock()
	return j, nil
}

func (js *Jobs) Get(h int) (*Job, bool) {
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
func (js *Jobs) Running() []*Job {
	var out []*Job
	for _, h := range js.Handles() {
		if j, _ := js.Get(h); j != nil && j.Status() == StatusRunning {
			out = append(out, j)
		}
	}
	return out
}

func (js *Jobs) SetNext(l Limits) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.pending = &l
}

// Take returns the limits for the next call: defaults, then the pending
// tool_limits if any (consumed), then the call's own arguments.
func (js *Jobs) Take(args json.RawMessage) (Limits, error) {
	l := DefaultLimits()
	js.mu.Lock()
	if js.pending != nil {
		l = *js.pending
		js.pending = nil
	}
	js.mu.Unlock()
	a, err := limitsInArgs(args)
	if err != nil {
		return l, err
	}
	return l.overlay(a)
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
	status JobStatus
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
	if j.status == StatusRunning {
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
		j.status = StatusDone
	}
	if j.file != nil {
		_ = j.file.Close()
		j.file = nil
	}
	j.broadcast()
}

func (j *Job) Status() JobStatus {
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

// WakeReason says why a wait ended. The model is told which, because "here
// is some output" means something different after three seconds than after
// the process exited.
type WakeReason int

const (
	WokeDone WakeReason = iota + 1
	WokeDelay
	WokePattern
)

// Wait blocks until the job is no longer running, or l.Delay passes, or the
// output the model has not yet seen matches l.Pattern — whichever first.
//
// It works on a job that has already finished, and returns at once. A student
// who implements this as "block on the channel" hangs forever on a completed
// job and diagnoses it as a deadlock in their own code.
func (j *Job) Wait(l Limits) WakeReason {
	timer := time.NewTimer(l.Delay)
	defer timer.Stop()
	for {
		j.mu.Lock()
		if j.status != StatusRunning {
			j.mu.Unlock()
			return WokeDone
		}
		if l.Pattern != nil && l.Pattern.Match(j.out.Bytes()[j.cursor:]) {
			j.mu.Unlock()
			return WokePattern
		}
		ch := j.changed
		j.mu.Unlock()
		select {
		case <-ch:
		case <-timer.C:
			return WokeDelay
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
	if j.status != StatusRunning {
		return false
	}
	if j.proc != nil {
		_ = syscall.Kill(-j.proc.Pid, syscall.SIGKILL)
	}
	j.status = StatusKilled
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
	if st != StatusRunning {
		return fmt.Errorf("job %d is %s; it is not reading input", j.Handle, st)
	}
	if w == nil {
		return fmt.Errorf("job %d has no stdin: %s is not a process", j.Handle, j.Tool)
	}
	_, err := w.Write([]byte(text))
	return err
}

// Data is the job as the event log records it.
func (j *Job) Data() *JobData {
	j.mu.Lock()
	defer j.mu.Unlock()
	return &JobData{
		Handle:   j.Handle,
		Status:   j.status,
		Output:   Ref{Kind: RefHandle, Locator: j.Path},
		Bytes:    j.out.Len(),
		ExitCode: j.exit,
	}
}

// Report renders what the model sees after a wait, and advances the cursor
// past it. Truncation to l.MaxOutput happens HERE, on the way into the
// context, with the full text already on disk.
//
// A job that finished within the delay, on its first report, with a result
// that fits, is reported as the bare result — byte for byte what Chapter 3
// returned. Everything else gets a status line, because the model needs to
// know whether it is looking at all of the answer or some of it.
func (j *Job) Report(reason WakeReason, l Limits) string {
	j.mu.Lock()
	defer j.mu.Unlock()
	all := j.out.Bytes()
	unseen := string(all[j.cursor:])
	first := j.cursor == 0
	j.cursor = len(all)

	body := capText(unseen, l.MaxOutput, len(all), j.Path)
	if reason == WokeDone && j.status == StatusDone && first {
		return body
	}

	var b strings.Builder
	switch j.status {
	case StatusDone:
		if j.err != nil {
			fmt.Fprintf(&b, "job %d done with error.", j.Handle)
		} else if j.exit != nil {
			fmt.Fprintf(&b, "job %d done, exit_code %d.", j.Handle, *j.exit)
		} else {
			fmt.Fprintf(&b, "job %d done.", j.Handle)
		}
	case StatusKilled:
		fmt.Fprintf(&b, "job %d killed.", j.Handle)
	default:
		switch reason {
		case WokePattern:
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
	if j.status == StatusRunning {
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
