package grade

// Chapter 8 checks — 100 points.
//
// The chapter's claim is that the agent does not know the GUI exists.
// These checks verify the WebSocket transport, replay, gui.log, pause gate,
// and backward compatibility with ch7.

import (
	"fmt"
	"strings"
)

// Ch8Evaluate scores a ch8 submission.
func Ch8Evaluate(r *Ch8Result) []Check {
	return []Check{
		ch8WebsocketStreams(r),
		ch8EventReplay(r),
		ch8GuiLog(r),
		ch8PauseHoldsTools(r),
		ch8Parity(r),
	}
}

func ch8Ready(c *Check, r *Ch8Result) bool {
	if !r.BuildOK {
		c.failf("build failed: %s", r.BuildErr)
		return false
	}
	return true
}

// websocket-streams: subscribe, prompt via WS; receive all observation types.
func ch8WebsocketStreams(r *Ch8Result) Check {
	c := Check{ID: "websocket-streams", Title: "WebSocket delivers all observation types", Points: 25, Passed: true, Earned: 25}
	if !ch8Ready(&c, r) {
		return c
	}
	if r.HelpersErr != "" {
		c.failf("harness error: %s", r.HelpersErr)
		return c
	}
	if len(r.Messages) == 0 {
		c.failf("received zero WebSocket messages")
		return c
	}

	// Check for required message types.
	typeSeen := map[string]bool{}
	for _, m := range r.Messages {
		typeSeen[m.Type] = true
	}

	required := []string{"part_delta", "part_final", "state_changed", "turn_ended", "tool_dispatched", "tool_finished"}
	var missing []string
	for _, t := range required {
		if !typeSeen[t] {
			missing = append(missing, t)
		}
	}
	if len(missing) > 0 {
		c.failf("missing message types: %v", missing)
		c.notef("saw types: %v", sortedStringKeys(typeSeen))
		return c
	}

	// Verify field integrity on key types.
	for _, m := range r.Messages {
		switch m.Type {
		case "tool_dispatched":
			if m.CallID == "" || m.Name == "" {
				c.failf("tool_dispatched missing call_id or name: %+v", m)
				return c
			}
		case "tool_finished":
			if m.CallID == "" {
				c.failf("tool_finished missing call_id: %+v", m)
				return c
			}
		case "state_changed":
			if m.From == "" || m.To == "" {
				c.failf("state_changed missing from or to: %+v", m)
				return c
			}
		}
	}

	c.notef("received %d messages with all %d required types", len(r.Messages), len(required))
	return c
}

// event-replay: disconnect, reconnect; event log delivers completed turn.
func ch8EventReplay(r *Ch8Result) Check {
	c := Check{ID: "event-replay", Title: "reconnection delivers event-log window", Points: 20, Passed: true, Earned: 20}
	if !ch8Ready(&c, r) {
		return c
	}
	if r.ReplayErr != "" {
		c.failf("replay test error: %s", r.ReplayErr)
		return c
	}
	if !r.ReplayOK {
		c.failf("replay did not deliver event-log content")
		c.notef("expected at least one part_final and one tool event from the event log")
		if len(r.ReplayMessages) > 0 {
			var types []string
			for _, m := range r.ReplayMessages {
				types = append(types, m.Type)
			}
			c.notef("replay returned %d messages with types: %v", len(r.ReplayMessages), types)
		}
		return c
	}

	c.notef("replay returned %d messages with event-log content", len(r.ReplayMessages))
	return c
}

// gui-log: gui.log exists with correct format.
func ch8GuiLog(r *Ch8Result) Check {
	c := Check{ID: "gui-log", Title: "gui.log records WebSocket traffic", Points: 15, Passed: true, Earned: 15}
	if !ch8Ready(&c, r) {
		return c
	}
	if r.HelpersErr != "" {
		c.failf("harness error: %s", r.HelpersErr)
		return c
	}
	if r.GuiLogContent == "" {
		c.failf("gui.log is empty or missing at %s", r.GuiLogPath)
		return c
	}

	lines := strings.Split(strings.TrimSpace(r.GuiLogContent), "\n")
	if len(lines) < 2 {
		c.failf("gui.log has only %d line(s), expected at least 2 (client + server)", len(lines))
		return c
	}

	// Check for both directions.
	hasInbound := false
	hasOutbound := false
	hasTimestamp := false
	hasJSON := false

	for _, line := range lines {
		if strings.Contains(line, " > ") {
			hasInbound = true
		}
		if strings.Contains(line, " < ") {
			hasOutbound = true
		}
		// Timestamp check: ISO 8601 starts with a year.
		if len(line) > 20 && line[4] == '-' {
			hasTimestamp = true
		}
		if strings.Contains(line, "{") {
			hasJSON = true
		}
	}

	if !hasInbound {
		c.failf("gui.log has no client-to-server (>) entries")
		return c
	}
	if !hasOutbound {
		c.failf("gui.log has no server-to-client (<) entries")
		return c
	}
	if !hasTimestamp {
		c.failf("gui.log entries lack timestamps")
		return c
	}
	if !hasJSON {
		c.failf("gui.log entries lack JSON content")
		return c
	}

	c.notef("gui.log has %d lines with both directions, timestamps, and JSON", len(lines))
	return c
}

// pause-holds-tools: pause prevents tool dispatch; unpause resumes.
func ch8PauseHoldsTools(r *Ch8Result) Check {
	c := Check{ID: "pause-holds-tools", Title: "pause gate holds tool dispatch", Points: 20, Passed: true, Earned: 20}
	if !ch8Ready(&c, r) {
		return c
	}
	if r.PauseErr != "" {
		c.failf("pause test error: %s", r.PauseErr)
		return c
	}
	if !r.PauseOK {
		c.failf("pause test did not complete")
		return c
	}

	// Count total tool_dispatched and tool_finished.
	dispatched := 0
	finished := 0
	turnEnded := false
	for _, m := range r.PauseMessages {
		switch m.Type {
		case "tool_dispatched":
			dispatched++
		case "tool_finished":
			finished++
		case "turn_ended":
			turnEnded = true
		}
	}

	if dispatched < 3 {
		c.failf("expected 3 tool_dispatched, got %d", dispatched)
		return c
	}
	if finished < 3 {
		c.failf("expected 3 tool_finished, got %d", finished)
		return c
	}
	if !turnEnded {
		c.failf("turn never ended after unpause")
		return c
	}

	c.notef("all 3 tools dispatched and finished; turn ended after unpause")
	return c
}

// ch7-parity: all ch7 checks still pass.
func ch8Parity(r *Ch8Result) Check {
	c := Check{ID: "ch7-parity", Title: "chapter 7 behavior is unchanged", Points: 20, Passed: true, Earned: 20}
	if r.Ch7Err != "" {
		c.failf("ch7 harness did not run: %s", r.Ch7Err)
		return c
	}
	if r.Ch7Result == nil {
		c.failf("ch7 harness produced no result")
		return c
	}
	var failed []string
	for _, sub := range Ch7Evaluate(r.Ch7Result) {
		if !sub.Passed {
			failed = append(failed, fmt.Sprintf("%s (%s)", sub.ID, firstOr(sub.Details, "no detail")))
		}
	}
	if len(failed) > 0 {
		c.failf("%d ch7 check(s) now fail: %v", len(failed), failed)
		return c
	}
	c.notef("all ch7 checks still pass")
	return c
}

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

func sortedStringKeys(m map[string]bool) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
