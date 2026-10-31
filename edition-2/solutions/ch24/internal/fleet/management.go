package fleet

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"ensemble/internal/common"
)

// Find has no global fallback: guessing another tree's ID reveals nothing.
func (m *manager) Find(id string) (common.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c := m.children[id]; c != nil {
		return c.agent, nil
	}
	return nil, fmt.Errorf("unknown agent %q", id)
}

// Children exposes copies of owned references for the application's tree view.
func (m *manager) Children() []common.Agent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]common.Agent, 0, len(m.children))
	for _, c := range m.children {
		out = append(out, c.agent)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Spec().ID < out[j].Spec().ID })
	return out
}

// GracefulShutdown leaves the current turn running. Root teardown and kill use
// ShutdownChild directly so an unresponsive provider cannot hold application exit.
func (m *manager) GracefulShutdown(ctx context.Context, id string) error {
	a, err := m.Find(id)
	if err != nil {
		return err
	}
	if err = a.Flush(ctx); err != nil {
		return err
	}
	return m.ShutdownChild(id)
}

// Output uses durable history sequence numbers. A caller never confuses a
// streamed fragment with its finalized replacement or replays old user input.
func (m *manager) Output(ctx context.Context, id string, cursor uint64, wait bool) (common.ChildOutput, error) {
	for {
		// Capture notification before the read; progress between read and sleep
		// closes this channel and cannot strand the waiter.
		m.mu.Lock()
		changed := m.changed
		m.mu.Unlock()
		a, err := m.Find(id)
		if err != nil {
			return common.ChildOutput{}, err
		}
		out := common.ChildOutput{ID: id, Cursor: cursor, Status: a.Fleet().Status()}
		var text strings.Builder
		for _, e := range a.History().EventsAfter(cursor) {
			out.Cursor = e.Seq
			if e.Response != nil {
				for _, p := range e.Response.Parts {
					if p.Type == "text" {
						text.WriteString(p.Text)
						text.WriteByte('\n')
					}
				}
			}
			if e.Tool != nil && (e.Type == "tool_returned" || e.Type == "job_report") {
				for _, p := range e.Tool.Parts {
					text.WriteString(p.Text)
					text.WriteByte('\n')
				}
			}
		}
		out.Content = text.String()
		if !wait || out.Content != "" || out.Status.Activity.Phase == common.ActivityIdle {
			return out, nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return out, ctx.Err()
		}
	}
}

// WaitChange snapshots only the requested owned children. Notifications prompt
// a fresh read; the timer supplies stuck detection even if no event arrives.
func (m *manager) WaitChange(ctx context.Context, ids []string, pattern string, stuck time.Duration) (string, error) {
	if len(ids) == 0 {
		for _, s := range m.List() {
			ids = append(ids, s.ID)
		}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	previous := map[string]string{}
	lastProgress := map[string]time.Time{}
	m.mu.Lock()
	messageSeq := make(map[string]uint64, len(ids))
	for _, id := range ids {
		messageSeq[id] = m.messages[id]
	}
	m.mu.Unlock()
	for _, id := range ids {
		a, err := m.Find(id)
		if err != nil {
			return "", err
		}
		previous[id] = encoded(a.Fleet().Status()) + fmt.Sprint(a.History().LastSeq())
		lastProgress[id] = time.Now()
	}
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()
	for {
		m.mu.Lock()
		changed := m.changed
		messages := make(map[string]uint64, len(ids))
		for _, id := range ids {
			messages[id] = m.messages[id]
		}
		m.mu.Unlock()
		var events []map[string]any
		for _, id := range ids {
			a, err := m.Find(id)
			if err != nil {
				events = append(events, map[string]any{"agent_id": id, "reason": "exit"})
				continue
			}
			status := a.Fleet().Status()
			now := encoded(status) + fmt.Sprint(a.History().LastSeq())
			reason := ""
			if status.State == "submitted" || status.State == "failed" {
				reason = "complete"
			}
			if now != previous[id] {
				reason = "change"
				lastProgress[id] = time.Now()
			}
			if messages[id] != messageSeq[id] {
				reason = "message"
			}
			if pattern != "" {
				out, _ := m.Output(ctx, id, 0, false)
				if re.MatchString(out.Content) {
					reason = "pattern"
				}
			}
			if progress := a.ProgressTime(); progress.After(lastProgress[id]) {
				lastProgress[id] = progress
			}
			if stuck > 0 && status.Activity.Phase != common.ActivityIdle && time.Since(lastProgress[id]) >= stuck {
				reason = "stuck"
			}
			if reason != "" {
				events = append(events, map[string]any{"agent_id": id, "reason": reason, "status": status})
			}
		}
		if len(events) > 0 {
			return encoded(events), nil
		}
		select {
		case <-changed:
		case <-timer.C:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
