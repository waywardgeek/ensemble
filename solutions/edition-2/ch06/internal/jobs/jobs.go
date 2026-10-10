// Package jobs owns supervised work, retained output and caller-controlled waits.
package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"ensemble/internal/common"
)

// The actor dispatches calls in order. Workers can supervise existing jobs and
// set the next-call limit, so collection admission and settings share this lock.
type jobs struct {
	mu      sync.Mutex
	stopped bool
	parent  common.Agent
	entries []common.Job
	pending *common.Limits
}

// New creates an Agent-owned collection; handles and retained files live beyond
// individual waits.
func New(parent common.Agent) common.Jobs {
	if parent == nil {
		panic("Jobs requires Agent")
	}
	return &jobs{parent: parent}
}

// Agent returns the required owning Agent rather than a copied service bundle.
func (j *jobs) Agent() common.Agent { return j.parent }

// Start creates a recoverable output file before publishing a new running job.
func (j *jobs) Start() (common.Job, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.stopped {
		return nil, common.ErrStopped
	}
	// Handles belong to Ensemble so two Agents do not silently reuse one.
	// Create the output before publishing the child: a failed open must not
	// launch work whose result has no recovery address.
	handle := j.parent.Ensemble().NextJobHandle()
	path := filepath.Join(j.parent.Config().DataDir, "io", fmt.Sprint(handle))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	child := &job{parent: j, file: file, changed: make(chan struct{}), data: common.JobData{Handle: handle, Status: "running", Output: common.Ref{Kind: common.RefHandle, Locator: path}}}
	j.entries = append(j.entries, child)
	return child, nil
}

// Find resolves a handle even after execution ended, allowing later output recovery.
func (j *jobs) Find(handle int) (common.Job, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	// Finished entries stay in the collection. Waiting after completion is a
	// normal observation, not a second receive from a consumed result channel.
	for _, child := range j.entries {
		if child.Data().Handle == handle {
			return child, nil
		}
	}
	return nil, fmt.Errorf("unknown job %d", handle)
}

// Shutdown closes process admission before visiting jobs. Go handlers retain
// their output and are joined by the actor; explicit Kill still discards late work.
func (j *jobs) Shutdown() error {
	j.mu.Lock()
	j.stopped = true
	children := append([]common.Job(nil), j.entries...)
	j.mu.Unlock()
	var failures []error
	for _, child := range children {
		failures = append(failures, child.Shutdown())
	}
	return errors.Join(failures...)
}

// Presence matters: zero delay means look now. Pending limits are consumed
// even by a malformed call or another setter, so they cannot become sticky.
func limits(raw json.RawMessage, base common.Limits) (common.Limits, error) {
	var args struct {
		// Delay optionally overrides this wait; zero means inspect immediately.
		Delay *float64 `json:"ai_callback_delay"`
		// Pattern wakes the wait on a regex match in output not yet reported.
		Pattern *string `json:"ai_callback_pattern"`
		// Max bounds the visible output without changing retained file contents.
		Max *int `json:"max_output_bytes"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return base, err
	}
	if args.Delay != nil {
		base.Delay = time.Duration(*args.Delay * float64(time.Second))
	}
	if args.Pattern != nil {
		base.Pattern = *args.Pattern
	}
	if args.Max != nil {
		base.MaxBytes = *args.Max
	}
	if base.Delay < 0 || base.MaxBytes < 1 {
		return base, errors.New("delay must be nonnegative and max_output_bytes positive")
	}
	if _, err := regexp.Compile(base.Pattern); err != nil {
		return base, err
	}
	return base, nil
}
func defaults() common.Limits { return common.Limits{Delay: 3 * time.Second, MaxBytes: 16 * 1024} }

// Consume burns one-shot limits and applies explicit overrides for this call.
func (j *jobs) Consume(raw json.RawMessage) (common.Limits, string, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	base := defaults()
	notice := ""
	if j.pending != nil {
		base = *j.pending
		notice = fmt.Sprintf("ai_callback_delay %s, ai_callback_pattern %q, max_output_bytes %d", base.Delay, base.Pattern, base.MaxBytes)
		j.pending = nil
	}
	value, err := limits(raw, base)
	if err != nil {
		return defaults(), notice, err
	}
	return value, notice, nil
}

// SetLimits validates the next-call override without extending any job lifetime.
func (j *jobs) SetLimits(raw json.RawMessage) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	// This setter starts from defaults, not from the setting it just consumed.
	// Otherwise omitted fields on consecutive setters would become sticky.
	value, err := limits(raw, defaults())
	if err == nil {
		j.pending = &value
	}
	return err
}

// One mutex covers state, cursor, writes and notification. A waiter takes its
// notification channel while holding the same lock that tests its condition;
// completion between that test and select therefore cannot be missed.
type job struct {
	parent   common.Jobs
	mu       sync.Mutex
	data     common.JobData
	file     *os.File
	changed  chan struct{}
	cursor   int64
	reports  int
	err      error
	process  *process
	stopping bool
}

// Jobs returns the owner of this execution's retained output and lookup lifetime.
func (j *job) Jobs() common.Jobs { return j.parent }

// Data returns a synchronized snapshot that cannot mutate the underlying job.
func (j *job) Data() common.JobData {
	j.mu.Lock()
	defer j.mu.Unlock()
	data := j.data
	// Events receive independent snapshots, including the optional exit value.
	// Returning the stored pointer would let a consumer mutate the job's fact.
	if data.ExitCode != nil {
		code := *data.ExitCode
		data.ExitCode = &code
	}
	return data
}

// The helpers below run with mu held. Replacing the channel after closing it
// broadcasts a change to current waiters without consuming the next change.
func (j *job) notify() { close(j.changed); j.changed = make(chan struct{}) }
func (j *job) write(text string) {
	if j.data.Status != "running" {
		return
	}
	n, err := j.file.WriteString(text)
	j.data.Bytes += int64(n)
	if err != nil {
		j.err = err
	}
	j.notify()
}
func (j *job) complete(text string, err error) {
	// A killed goroutine may eventually return. Its late result cannot reopen
	// the output file or change the recorded kill into successful completion.
	if j.data.Status != "running" {
		return
	}
	if err != nil {
		text = err.Error()
		j.err = err
	}
	j.write(text)
	if closeErr := j.file.Close(); closeErr != nil {
		j.err = closeErr
	}
	j.data.Status = "done"
	j.notify()
}

// Finish records Go handler output; a launched process retains completion ownership.
func (j *job) Finish(text string, err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	// A started process has transferred completion to its reader. Returning
	// from run_command means it started, not that the output file can close.
	if j.process == nil {
		j.complete(text, err)
	}
}

// Read only the requested byte span. Reports normally read at most their cap;
// regex matching needs the unseen span because patterns may cross write chunks.
func (j *job) read(offset, count int64) (string, error) {
	// Reopen by address so a completed job remains readable after its writer
	// closes. The cursor is a job fact, not a file descriptor position.
	file, err := os.Open(j.data.Output.Locator)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data := make([]byte, count)
	if count == 0 {
		return "", nil
	}
	_, err = file.ReadAt(data, offset)
	return string(data), err
}

// Wait returns on completion, a fresh output match or the callback delay; it does
// not cancel execution.
func (j *job) Wait(limits common.Limits) (string, error) {
	pattern, err := regexp.Compile(limits.Pattern)
	if err != nil {
		return "", err
	}
	timer := time.NewTimer(limits.Delay)
	defer timer.Stop()
	// A wake observes the same job on either branch. It neither closes output
	// nor signals the process, so a later call can choose a different delay.
	for {
		j.mu.Lock()
		ready := j.data.Status != "running"
		// An old prompt cannot certify readiness for another input line.
		// Match only the bytes written after the last report cursor.
		if !ready && limits.Pattern != "" {
			unseen, readErr := j.read(j.cursor, j.data.Bytes-j.cursor)
			if readErr != nil {
				j.mu.Unlock()
				return "", readErr
			}
			ready = pattern.MatchString(unseen)
		}
		if ready {
			result, err := j.report(limits)
			j.mu.Unlock()
			return result, err
		}
		changed := j.changed
		j.mu.Unlock()
		select {
		case <-changed:
		case <-timer.C:
			j.mu.Lock()
			result, err := j.report(limits)
			j.mu.Unlock()
			return result, err
		}
	}
}
func (j *job) report(limits common.Limits) (string, error) {
	if j.data.Status == "killed" {
		// Name the recovery file without appending old text that could look
		// like a successful exit. The stored bytes remain intact.
		note := ""
		if j.process == nil {
			note = "; Go cannot stop the tool goroutine"
		}
		return fmt.Sprintf("job %d killed%s; output at %s", j.data.Handle, note, j.data.Output.Locator), nil
	}
	count := j.data.Bytes - j.cursor
	size := min(count, int64(limits.MaxBytes))
	text, err := j.read(j.cursor, size)
	if err != nil {
		return "", err
	}
	if count > size {
		// Bound the payload before it reaches context. The omitted middle stays
		// on disk; the marker names both its extent and the recovery address.
		head := size / 2
		tail, err := j.read(j.data.Bytes-(size-head), size-head)
		if err != nil {
			return "", err
		}
		text = text[:head] + fmt.Sprintf("\n[%d bytes omitted; %d bytes total at %s]\n", count-size, j.data.Bytes, j.data.Output.Locator) + tail
	}
	// Preserve small first results verbatim. Later reports need a status even
	// when there are no new bytes, otherwise an empty result looks ambiguous.
	bare := j.reports == 0 && j.data.Status == "done" && count <= size && j.data.Cwd == ""
	// Omitted bytes count as reported too: repeating the same flood on the
	// next wait would defeat truncation. read_file remains the recovery route.
	j.cursor = j.data.Bytes
	j.reports++
	if bare {
		return text, j.err
	}
	cwd := ""
	if j.data.Cwd != "" {
		cwd = "; cwd " + j.data.Cwd
	}
	header := fmt.Sprintf("job %d %s%s; %d bytes total at %s\n", j.data.Handle, j.data.Status, cwd, j.data.Bytes, j.data.Output.Locator)
	if text == "" {
		text = "no new output\n"
	}
	if j.data.Status == "running" {
		text += fmt.Sprintf("\nUse wait_for_job(%d), send_input or kill_job to supervise this job.", j.data.Handle)
	}
	return header + text, j.err
}
