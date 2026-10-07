// Package jobs owns retained execution, wait timing, and transactional reports.
package jobs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"example.com/ensemble/internal/common"
)

type Service struct {
	parent  common.JobAgent
	mu      sync.Mutex
	jobs    map[uint64]*job
	pending *common.LimitOverrides
	closed  bool
	fault   error
}
type job struct {
	parent     common.Jobs
	snapshot   common.JobSnapshot
	file       *os.File
	cursor     int64
	reports    int
	reportNote string
	changed    chan struct{}
	reaped     chan struct{}
	terminal   *os.File
	pid        int
	killing    bool
	killReason string
}

func New(parent common.JobAgent) *Service { return &Service{parent: parent, jobs: map[uint64]*job{}} }
func (s *Service) Agent() common.JobAgent { return s.parent }
func (j *job) Jobs() common.Jobs          { return j.parent }
func (j *job) service() *Service          { return j.parent.(*Service) }
func (j *job) Snapshot() common.JobSnapshot {
	s := j.service()
	s.mu.Lock()
	defer s.mu.Unlock()
	snapshot := j.snapshot
	if snapshot.ExitCode != nil {
		code := *snapshot.ExitCode
		snapshot.ExitCode = &code
	}
	return snapshot
}
func (s *Service) changed(j *job) { close(j.changed); j.changed = make(chan struct{}) }
func (s *Service) fail(err error) {
	if s.fault == nil {
		s.fault = err
		s.parent.Fault(err)
	}
	for _, j := range s.jobs {
		s.changed(j)
	}
}
func (s *Service) Create() (common.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.fault != nil {
		return nil, fmt.Errorf("Jobs is closed or faulted")
	}
	// Exclusive creation is also the collision test: a prior application's
	// artifacts consume candidates but never become this Agent's live jobs.
	var handle uint64
	var locator string
	var file *os.File
	for {
		handle = s.parent.Ensemble().AllocateHandle()
		locator = fmt.Sprintf("cr/io/%d", handle)
		path := filepath.Join(s.parent.Workspace(), locator)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, fmt.Errorf("create job output: %w", err)
		}
		var err error
		file, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("create job output exclusively: %w", err)
		}
		break
	}
	j := &job{parent: s, snapshot: common.JobSnapshot{Handle: handle, Status: "running", Output: common.Ref{Kind: 3, Locator: locator}}, file: file, changed: make(chan struct{}), reaped: make(chan struct{})}
	s.jobs[handle] = j
	return j, nil
}
func (s *Service) Start(owned common.Job, call common.Part) {
	j := owned.(*job)
	go func() {
		s.mu.Lock()
		stopped := j.snapshot.Status != "running" || j.killing || s.closed
		s.mu.Unlock()
		if stopped {
			return
		}
		result := s.parent.Registry().ExecuteJob(call, j)
		if result == nil {
			return
		} // The process worker now owns completion.
		s.mu.Lock()
		defer s.mu.Unlock()
		if j.snapshot.Status != "running" || j.killing {
			return
		}
		s.write(j, []byte(result.Text))
		j.reportNote = result.Note
		j.snapshot.IsError = result.IsError
		s.finish(j, "done", "")
	}()
}
func (s *Service) write(j *job, data []byte) {
	if s.fault != nil || len(data) == 0 {
		return
	}
	n, err := j.file.Write(data)
	j.snapshot.Bytes += int64(n)
	if err != nil || n != len(data) {
		j.snapshot.IsError = true
		s.fail(fmt.Errorf("job %d output write failed; complete output unavailable", j.snapshot.Handle))
		if j.pid > 0 {
			_ = syscall.Kill(-j.pid, syscall.SIGKILL)
		}
	}
	s.changed(j)
}
func (s *Service) finish(j *job, status, reason string) {
	if j.snapshot.Status != "running" {
		return
	}
	j.snapshot.Status = status
	j.snapshot.Reason = reason
	if status == "killed" {
		j.snapshot.ExitCode = nil
	}
	if err := j.file.Close(); err != nil {
		s.fail(fmt.Errorf("job output close: %w", err))
	}
	event := "job_ended"
	if status == "killed" {
		event = "job_killed"
	}
	snapshot := j.snapshot
	if err := s.parent.RecordJob(common.Event{Type: event, Job: &snapshot}); err != nil {
		s.fail(err)
		s.parent.Ensemble().Logf("job %d cleanup/terminal record could not be retained: %v", snapshot.Handle, err)
	}
	s.changed(j)
}

// RequestKill starts the owned effect without waiting for pipe readers/reaping.
// The actor can continue handling controls while a report worker awaits Ready.
func (s *Service) RequestKill(owned common.Job, reason string) (<-chan struct{}, error) {
	j := owned.(*job)
	s.mu.Lock()
	defer s.mu.Unlock()
	ready := func() <-chan struct{} { done := make(chan struct{}); close(done); return done }
	if j.snapshot.Status != "running" {
		return ready(), nil
	}
	if j.killing {
		return j.reaped, nil
	}
	if j.pid == 0 {
		s.finish(j, "killed", reason)
		return ready(), nil
	}
	if err := syscall.Kill(-j.pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return nil, fmt.Errorf("kill process group: %w", err)
	}
	j.killing = true
	j.killReason = reason
	s.changed(j)
	return j.reaped, nil
}
func (s *Service) kill(j *job, reason string) error {
	done, err := s.RequestKill(j, reason)
	if err == nil {
		<-done
	}
	return err
}
func (s *Service) Close() error {
	s.mu.Lock()
	if s.closed {
		err := s.fault
		s.mu.Unlock()
		return err
	}
	s.closed = true
	list := make([]*job, 0, len(s.jobs))
	for _, j := range s.jobs {
		list = append(list, j)
	}
	s.mu.Unlock()
	var first error
	for _, j := range list {
		if err := s.kill(j, "shutdown"); first == nil {
			first = err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if first != nil {
		return first
	}
	return s.fault
}

// Lookup enforces Agent scope even though Ensemble allocates application-wide IDs.
func (s *Service) Lookup(handle uint64) (common.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := s.jobs[handle]
	if j == nil {
		return nil, fmt.Errorf("handle %d is unavailable to this Agent", handle)
	}
	return j, nil
}
func (s *Service) Kill(owned common.Job, reason string) error { return s.kill(owned.(*job), reason) }
func (s *Service) Send(owned common.Job, input string) (int64, error) {
	j := owned.(*job)
	s.mu.Lock()
	if j.snapshot.Status != "running" || j.killing || j.terminal == nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("job is finished or has no input-capable process")
	}
	matchStart := j.snapshot.Bytes
	terminal := j.terminal
	s.mu.Unlock()
	if _, err := terminal.Write([]byte(input)); err != nil {
		return 0, fmt.Errorf("send input: %w", err)
	}
	return matchStart, nil
}

// Abort releases an allocation whose pre-dispatch event failed. It never
// invents a terminal event for a handle absent from the durable conversation.
func (s *Service) Abort(owned common.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j := owned.(*job)
	_ = j.file.Close()
	delete(s.jobs, j.snapshot.Handle)
}
