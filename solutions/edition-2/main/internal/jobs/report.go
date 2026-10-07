package jobs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"example.com/ensemble/internal/common"
)

func (s *Service) Report(owned common.Job, call common.Part, limits common.Limits, note string) error {
	return s.report(owned.(*job), call, limits, note, -1)
}
func (s *Service) report(j *job, call common.Part, limits common.Limits, note string, matchStart int64) error {
	timer := time.NewTimer(limits.Delay)
	defer timer.Stop()
	s.mu.Lock()
	defer s.mu.Unlock()
	if matchStart < 0 {
		matchStart = j.cursor
	}
	for {
		if s.fault != nil {
			return s.fault
		}
		matched := false
		if limits.Pattern != nil && j.snapshot.Bytes > matchStart {
			file, err := s.open(j)
			if err != nil {
				return err
			}
			matched = limits.Pattern.MatchReader(bufio.NewReader(io.NewSectionReader(file, matchStart, j.snapshot.Bytes-matchStart)))
			file.Close()
		}
		if j.snapshot.Status != "running" || matched {
			break
		}
		changed := j.changed
		s.mu.Unlock()
		expired := false
		select {
		case <-timer.C:
			expired = true
		case <-changed:
		}
		s.mu.Lock()
		if expired {
			break
		}
	}
	if s.fault != nil {
		return s.fault
	}
	text, capped, err := s.content(j, limits.MaxBytes)
	if err != nil {
		return err
	}
	snapshot := j.snapshot
	// Holding Jobs' mutex across report append orders its snapshot and cursor with
	// terminal publication; the long wait above never holds this mutex.
	if snapshot.Status != "done" || j.reports > 0 || capped || note != "" || snapshot.Cwd != "" || snapshot.IsError {
		metadata := fmt.Sprintf("job %d status: %s; output: %s; total bytes: %d", snapshot.Handle, snapshot.Status, snapshot.Output.Locator, snapshot.Bytes)
		if snapshot.Cwd != "" {
			metadata += "; cwd: " + snapshot.Cwd
		}
		if snapshot.IsError {
			metadata += "; execution failed"
		}
		if snapshot.Status == "running" {
			metadata += "; use wait_for_job, send_input, or kill_job"
		}
		if snapshot.Status == "killed" && j.pid == 0 {
			metadata += "; local function may continue and still have side effects"
		}
		text = metadata + "\n" + text
	}
	text = note + text
	result := common.ToolEvent{CallID: call.CallID, Parts: []common.Part{{Type: "text", Text: &text}}}
	_, supervision := s.parent.Registry().Kind(call.Name)
	if !supervision {
		result.Job = &snapshot
		result.IsError = snapshot.IsError
	}
	if err = s.parent.RecordJob(common.Event{Type: "tool_returned", Tool: &result}); err != nil {
		s.fail(err)
		return err
	}
	j.cursor = snapshot.Bytes
	j.reports++
	return nil
}
func (s *Service) open(j *job) (*os.File, error) {
	file, err := os.Open(filepath.Join(s.parent.Workspace(), j.snapshot.Output.Locator))
	if err != nil {
		s.fail(fmt.Errorf("read job output: %w", err))
	}
	return file, err
}
func (s *Service) content(j *job, budget int) (string, bool, error) {
	count := j.snapshot.Bytes - j.cursor
	if count == 0 {
		return "", false, nil
	}
	file, err := s.open(j)
	if err != nil {
		return "", false, err
	}
	defer file.Close()
	read := func(offset int64, n int) ([]byte, error) {
		data := make([]byte, n)
		_, err := file.ReadAt(data, offset)
		return data, err
	}
	if count <= int64(budget) {
		data, err := read(j.cursor, int(count))
		return string(data), false, err
	}
	head, err := read(j.cursor, budget/2+budget%2)
	if err != nil {
		return "", false, err
	}
	tail, err := read(j.snapshot.Bytes-int64(budget/2), budget/2)
	if err != nil {
		return "", false, err
	}
	// Preserve complete UTF-8 code points at both cuts; unused budget is harmless.
	for len(head) > 0 && !utf8.Valid(head) {
		head = head[:len(head)-1]
	}
	for len(tail) > 0 && !utf8.RuneStart(tail[0]) {
		tail = tail[1:]
	}
	omitted := count - int64(len(head)+len(tail))
	notice := fmt.Sprintf("\n[truncated: %d bytes omitted; total bytes: %d; output: %s]\n", omitted, j.snapshot.Bytes, j.snapshot.Output.Locator)
	return string(head) + notice + string(tail), true, nil
}
