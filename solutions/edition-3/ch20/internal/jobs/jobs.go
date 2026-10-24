// Package jobs owns execution, output cursors, and supervised lifetimes.
package jobs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"ensemble/internal/common"
)

type collection struct {
	parent common.Agent
	mu     sync.Mutex
	closed bool
}

func New(parent common.Agent) common.Jobs {
	return &collection{parent: parent}
}
func (c *collection) Parent() common.Agent { return c.parent }
func (c *collection) Get(handle int) (common.Job, error) {
	j, ok := c.parent.Parent().LookupJob(handle)
	if !ok || j.Parent() != c {
		return nil, fmt.Errorf("unknown job handle %d for this Agent", handle)
	}
	return j, nil
}
func (c *collection) Shutdown() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	var owned []common.Job
	for _, j := range c.parent.Parent().JobTable() {
		if j.Parent() == c {
			owned = append(owned, j)
		}
	}
	errs := make([]error, len(owned))
	var joins sync.WaitGroup
	for i, j := range owned {
		joins.Add(1)
		go func() { defer joins.Done(); errs[i] = j.Shutdown() }()
	}
	joins.Wait()
	return errors.Join(errs...)
}

type job struct {
	parent       *collection
	mu           sync.Mutex
	data         common.JobData
	file         *os.File
	cursor       int64
	changed      chan struct{}
	settled      chan struct{}
	killRecorded chan struct{}
	cancel       context.CancelFunc
	input        io.Writer
	stop         func() error
	failed       bool
	setupError   string
	shutting     bool
	pending      func()
}

func (j *job) Parent() common.Jobs { return j.parent }
func (j *job) snapshot() common.JobData {
	d := j.data
	d.Args = append([]byte(nil), d.Args...)
	if d.ExitCode != nil {
		n := *d.ExitCode
		d.ExitCode = &n
	}
	return d
}
func (j *job) Data() common.JobData { j.mu.Lock(); defer j.mu.Unlock(); return j.snapshot() }
func (j *job) wake()                { close(j.changed); j.changed = make(chan struct{}) }

// Start retains the immediate-start API for direct job users.
func (c *collection) Start(ctx context.Context, handler common.ToolHandler, args []byte, cwd string, origin common.JobData) (common.Job, error) {
	j, err := c.Reserve(ctx, handler, args, cwd, origin)
	if err == nil {
		j.Start()
	}
	return j, err
}

// Reserve assigns the stable handle and output before the Agent records the call.
// No handler is released until that record (and configured save) succeeds.
func (c *collection) Reserve(ctx context.Context, handler common.ToolHandler, args []byte, cwd string, origin common.JobData) (common.Job, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, errors.New("Agent jobs are shut down")
	}
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	j := &job{parent: c, changed: make(chan struct{}), settled: make(chan struct{}), cancel: cancel, data: origin}
	j.mu.Lock()
	handle, err := c.parent.Parent().RegisterJob(j)
	if err != nil {
		j.mu.Unlock()
		cancel()
		return nil, err
	}
	j.data.Handle = handle
	j.data.Status = common.JobRunning
	j.data.Cwd = cwd
	config := c.parent.Config()
	locator := filepath.Join(config.DataDir, "io", fmt.Sprint(handle))
	path := locator
	if !filepath.IsAbs(path) {
		path = filepath.Join(config.BaseDir, path)
	}
	j.data.Output = common.Ref{Kind: common.RefHandle, Locator: filepath.ToSlash(locator)}
	if err = c.parent.Sandbox().MkdirAll(filepath.Dir(path), 0700); err == nil {
		j.file, err = c.parent.Sandbox().OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
	}
	if err != nil {
		j.setupError = "preparing job output: " + err.Error()
		j.data.Output = common.Ref{}
	}

	if err != nil {
		j.failed = true
		j.data.Status = common.JobDone
		if j.file != nil {
			_ = j.file.Close()
			j.file = nil
		}
		j.wake()
		close(j.settled)
		j.mu.Unlock()
		cancel()
		return j, err
	}
	input := append([]byte(nil), args...)
	j.pending = func() {
		if err := runCtx.Err(); err != nil {
			j.finish(err)
			return
		}
		err := handler(runCtx, &call{parent: c.parent, job: j}, input)
		j.finish(err)
	}
	j.mu.Unlock()
	return j, nil
}
func (j *job) Start() bool { started, _ := j.StartAfter(nil); return started }

// The callback may save/queue facts but must not call back into this Job or
// await job/actor work. Holding reservation admission excludes Kill/Abort races;
// no worker or terminal-fact acknowledgement runs until the boundary succeeds.
func (j *job) StartAfter(before func() error) (bool, error) {
	j.mu.Lock()
	if j.data.Status != common.JobRunning || j.pending == nil {
		j.mu.Unlock()
		return false, nil
	}
	if before != nil {
		if err := before(); err != nil {
			j.abortLocked(err)
			j.mu.Unlock()
			return false, err
		}
	}
	run := j.pending
	j.pending = nil
	j.mu.Unlock()
	go run()
	return true, nil
}
func (j *job) abortLocked(err error) {
	if j.pending == nil {
		return
	}
	j.pending = nil
	j.failed = true
	j.data.Status = common.JobDone
	j.data.Reason = "not started: " + err.Error()
	j.setupError = j.data.Reason
	j.cancel()
	j.closeWriter()
	j.wake()
	close(j.settled)
}
func (j *job) Abort(err error) { j.mu.Lock(); defer j.mu.Unlock(); j.abortLocked(err) }

// call carries its real Agent parent, not a copied sibling-service bag.
type call struct {
	parent common.Agent
	job    *job
}

func (c *call) Parent() common.Agent        { return c.parent }
func (c *call) Write(p []byte) (int, error) { return c.job.write(p) }
func (c *call) SetExitCode(code int) {
	c.job.mu.Lock()
	defer c.job.mu.Unlock()
	if c.job.data.Status == common.JobRunning {
		c.job.data.ExitCode = &code
	}
}
func (c *call) StartProcess(start func() (io.Writer, func() error, error)) error {
	j := c.job
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.data.Status != common.JobRunning || j.shutting {
		return errors.New("job no longer running")
	}
	input, stop, err := start()
	if err != nil {
		return err
	}
	j.input, j.stop = input, stop
	return nil
}
func (j *job) write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.data.Status != common.JobRunning {
		return len(p), nil
	}
	n, err := j.file.Write(p)
	j.data.Bytes += int64(n)
	if err != nil {
		j.failed = true
	}
	j.wake()
	return n, err
}
func (j *job) finish(err error) {
	j.mu.Lock()
	completed := j.data.Status == common.JobRunning
	defer func() {
		report := common.JobReport{Job: j.snapshot(), IsError: j.failed}
		j.mu.Unlock()
		// Kill already recorded its terminal transition. A late Go return
		// cannot change that outcome or publish another killed/done fact.
		if completed {
			j.parent.parent.ReportJob(report)
		}
		close(j.settled)
	}()
	if j.data.Status != common.JobRunning {
		return
	}
	if err != nil {
		j.failed = true
		message := err.Error()
		if j.data.Bytes > 0 {
			message = "\n" + message
		}
		n, _ := j.file.WriteString(message)
		j.data.Bytes += int64(n)
	}
	j.data.Status = common.JobDone
	j.cancel()
	// Observations open bounded read views; completion closes the writer.
	j.closeWriter()
	j.wake()
}
func (j *job) closeWriter() {
	if j.file != nil {
		if err := j.file.Close(); err != nil {
			j.failed = true
		}
		j.file = nil
	}
}
func (j *job) Kill(reason string) error {
	if reason != "kill_job" && reason != "shutdown" {
		return errors.New("invalid kill reason")
	}
	j.mu.Lock()
	if j.data.Status != common.JobRunning {
		process := j.stop != nil
		j.mu.Unlock()
		if process {
			<-j.settled
		}
		return nil
	}
	// Reserve terminal-fact delivery under the same lock as the transition.
	// Handler settlement alone does not mean an external Kill has reported it.
	j.killRecorded = make(chan struct{})
	j.data.Status = common.JobKilledStatus
	j.data.ExitCode = nil
	j.data.Reason = reason
	if j.pending != nil {
		j.pending = nil
		close(j.settled)
	}
	j.cancel()
	stop := j.stop
	j.closeWriter()
	j.wake()
	report := common.JobReport{Job: j.snapshot()}
	j.mu.Unlock()
	var err error
	if stop != nil {
		err = stop()
		<-j.settled // Managed processes must really stop and be reaped.
	}
	// Go-only work is contained, not joined, by explicit kill. Shutdown is
	// responsible for joining the handler after it actually returns.
	j.parent.parent.ReportJob(report)
	close(j.killRecorded)
	return err
}

// Shutdown cancels cooperative Go work, but joins it and preserves its real
// result. Only managed processes receive destructive kill semantics.
func (j *job) Shutdown() error {
	j.mu.Lock()
	j.shutting = true
	j.abortLocked(errors.New("Agent shutdown before call admission"))
	j.cancel()
	process := j.stop != nil
	j.mu.Unlock()
	var err error
	if process {
		err = j.Kill("shutdown")
	}
	<-j.settled
	j.mu.Lock()
	recorded := j.killRecorded
	j.mu.Unlock()
	// Keep the actor alive to acknowledge any admitted kill, even when its
	// external caller is still between process reap and terminal reporting.
	if recorded != nil {
		<-recorded
	}
	return err
}

func (j *job) outputPath() string {
	path := filepath.FromSlash(j.data.Output.Locator)
	if !filepath.IsAbs(path) {
		path = filepath.Join(j.parent.parent.Config().BaseDir, path)
	}
	return path
}

// observeLocked reads at most the payload cap; the cursor covers omitted bytes too.
func (j *job) observeLocked(l common.ToolLimits) (report common.JobReport, err error) {
	report = common.JobReport{Job: j.snapshot(), IsError: j.failed}
	defer func() {
		if j.data.Status == common.JobKilledStatus && j.stop == nil {
			if report.Text != "" {
				report.Text += "\n"
			}
			report.Text += "Go cannot stop this handler; the job is marked killed and further output is discarded."
		}
	}()
	if j.setupError != "" {
		report.Text = j.setupError
		return report, nil
	}
	count := j.data.Bytes - j.cursor
	if count == 0 {
		return report, nil
	}
	f, err := j.parent.parent.Sandbox().OpenFile(j.outputPath(), os.O_RDONLY, 0)
	if err != nil {
		return report, err
	}
	defer f.Close()
	read := func(offset int64, n int) (string, error) {
		b := make([]byte, n)
		_, err := f.ReadAt(b, offset)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if count <= int64(l.MaxOutputBytes) {
		report.Text, err = read(j.cursor, int(count))
	} else {
		head := l.MaxOutputBytes / 2
		tail := l.MaxOutputBytes - head
		var a, b string
		a, err = read(j.cursor, head)
		if err == nil {
			b, err = read(j.data.Bytes-int64(tail), tail)
		}
		report.Text = fmt.Sprintf("%s\n[output truncated; %d bytes omitted; %d bytes total; full output: %s]\n%s", a, count-int64(l.MaxOutputBytes), j.data.Bytes, j.data.Output.Locator, b)
	}
	if err == nil {
		j.cursor = j.data.Bytes
	}
	return report, err
}

// InitialReport preserves fast completed calls' original bytes. Without a
// pattern it waits for completion or the observation delay, not the first chunk.
func (j *job) InitialReport(ctx context.Context, l common.ToolLimits) (common.JobReport, error) {
	return j.wait(ctx, l, false)
}
func (j *job) Wait(ctx context.Context, l common.ToolLimits) (common.JobReport, error) {
	return j.wait(ctx, l, true)
}
func (j *job) wait(ctx context.Context, l common.ToolLimits, outputWake bool) (common.JobReport, error) {
	if l.Delay < 0 || l.MaxOutputBytes <= 0 {
		return common.JobReport{}, errors.New("invalid observation limits")
	}
	var pattern *regexp.Regexp
	var err error
	if l.Pattern != "" {
		pattern, err = regexp.Compile(l.Pattern)
		if err != nil {
			return common.JobReport{}, fmt.Errorf("invalid callback pattern: %w", err)
		}
	}
	timer := time.NewTimer(l.Delay)
	defer timer.Stop()
	expired := l.Delay == 0
	for {
		j.mu.Lock()
		ready := expired || j.data.Status != common.JobRunning
		if !ready && j.data.Bytes > j.cursor {
			if pattern == nil {
				ready = outputWake
			} else {
				f, e := j.parent.parent.Sandbox().OpenFile(j.outputPath(), os.O_RDONLY, 0)
				if e != nil {
					j.mu.Unlock()
					return common.JobReport{}, e
				}
				ready = pattern.MatchReader(bufio.NewReader(io.NewSectionReader(f, j.cursor, j.data.Bytes-j.cursor)))
				_ = f.Close()
			}
		}
		if ready {
			result, e := j.observeLocked(l)
			j.mu.Unlock()
			return result, e
		}
		changed := j.changed
		j.mu.Unlock()
		select {
		case <-changed:
		case <-timer.C:
			expired = true
		case <-ctx.Done():
			// Wake only this observation, consuming bounded unseen bytes once.
			expired = true
		}
	}
}
func (j *job) SendInput(ctx context.Context, text string, l common.ToolLimits) (common.JobReport, error) {
	j.mu.Lock()
	input, status := j.input, j.data.Status
	j.mu.Unlock()
	if input == nil {
		return common.JobReport{}, errors.New("job does not support input")
	}
	if status != common.JobRunning {
		return common.JobReport{}, errors.New("job is not running")
	}
	// Submit one terminal line; a caller-supplied terminator is not doubled.
	if !strings.HasSuffix(text, "\n") && !strings.HasSuffix(text, "\r") {
		text += "\n"
	}
	if _, err := io.WriteString(input, text); err != nil {
		return common.JobReport{}, fmt.Errorf("sending input: %w", err)
	}
	// As with a new command, allow the response to arrive after terminal echo.
	// An explicit pattern still wakes promptly; subsequent Wait calls return
	// available unseen output immediately.
	return j.InitialReport(ctx, l)
}

// Complete is a cancellable observer of terminal output, not an execution timeout.
// It is used for automatic context; ordinary model calls retain callback reports.
func (j *job) Complete(ctx context.Context) (common.JobReport, error) {
	select {
	case <-ctx.Done():
		return common.JobReport{}, ctx.Err()
	case <-j.settled:
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	report := common.JobReport{Job: j.data, IsError: j.failed}
	if j.setupError != "" {
		report.Text = j.setupError
		return report, nil
	}
	data, err := j.parent.parent.Sandbox().ReadFile(j.outputPath())
	report.Text = string(data)
	return report, err
}
