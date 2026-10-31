package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// memoryHeadroom holds a real worker at the external response boundary. At the
// known 95% threshold, the owner must wait without making another main request.
func memoryHeadroom(bin, dir, mode string) error {
	skill := "---\nname: base\ndescription: Memory headroom fixture\ntype: primary\ntools: micro_handoff\n---\nFinish main work in chat.\n"
	if err := recallWrite(filepath.Join(dir, "skills"), map[string]string{"base/SKILL.md": skill}); err != nil {
		return err
	}
	const candidate = "I retained HEADROOM_MEMORY_15 and the reason for the selected decision."
	replies := memoryTalk(1)
	// Disjoint cache buckets together exceed 95% of the course model's 4096.
	replies[0].Usage = fakevendor.Canonical{Input: 3000, CacheRead: 900, CacheWrite: 100, Output: 20}
	provider := fakevendor.NewWithOptions(replies, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		if !memoryCompressor(body) {
			return nil
		}
		text := candidate
		if mode == "failure" {
			text = " "
		}
		return &fakevendor.Reply{ToolName: "submit", ToolID: "headroom-memory", ToolArgs: memoryJSON(text), Usage: fakevendor.Canonical{Input: 100, Output: 20}}
	}})
	defer provider.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var enteredOnce, releaseOnce sync.Once
	unpark := func() { releaseOnce.Do(func() { close(release) }) }
	upstream, _ := url.Parse(provider.URL())
	forward := httputil.NewSingleHostReverseProxy(upstream)
	boundary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Body = io.NopCloser(bytes.NewReader(raw))
		proxy := *forward
		// Hold after capture so provider.Requests and the owner remain usable.
		if memoryCompressor(raw) {
			proxy.ModifyResponse = func(*http.Response) error {
				enteredOnce.Do(func() { close(entered) })
				select {
				case <-release:
					return nil
				case <-req.Context().Done():
					return req.Context().Err()
				}
			}
		}
		proxy.ServeHTTP(w, req)
	}))
	defer boundary.Close()
	defer unpark()
	p, peer, err := managementStart(bin, dir, boundary.URL)
	if err != nil {
		return err
	}
	defer p.close()
	defer peer.close()
	if err = peer.send("subscribe", ""); err != nil {
		return err
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return err
	}
	// Session model selection and saved preferences are separate public patches.
	for _, patch := range []map[string]any{{"model": "claude-ch16-course"}, memorySettings(5000)} {
		if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": patch}); err != nil {
			return err
		}
		effective, e := readSettings(peer, "settings_changed")
		if e != nil {
			return fmt.Errorf("headroom settings: %w", e)
		}
		if effective["model"] != "claude-ch16-course" {
			return fmt.Errorf("course model was not selected: %v", effective["model"])
		}
	}
	if err = peer.send("prompt", memoryPrompts(1)[0]); err != nil {
		return err
	}
	wait := func() error {
		_, err := peer.until(3*time.Second, func(frame map[string]any) bool {
			return frame["type"] == "agent_activity" && frame["key"] == "main" && managementPhase(frame["activity"]) == "waiting"
		})
		if err != nil {
			return fmt.Errorf("owner did not report near-window waiting: %w", err)
		}
		// Acknowledgment crosses the actor while it is waiting. Neither elapsed
		// time nor an unsynchronized HTTP count establishes the held boundary.
		if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]any{}}); err != nil {
			return err
		}
		if _, err = readSettings(peer, "settings_changed"); err != nil {
			return fmt.Errorf("settings blocked behind near-window worker: %w", err)
		}
		if n := len(memoryTurns(provider.Requests())); n != 1 {
			return fmt.Errorf("near-window owner rendered before worker finished: %d main requests", n)
		}
		return nil
	}
	if err = wait(); err != nil {
		return err
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		return fmt.Errorf("near-window worker never reached external hold")
	}
	// Retain this attempt identity across interruption and its replacement turn.
	pending, err := memoryPending(dir)
	if err != nil || len(pending) != 1 {
		return fmt.Errorf("near-window owned launch: %v (%v)", pending, err)
	}
	if mode == "interrupt" {
		if err = peer.send("interrupt", ""); err != nil {
			return err
		}
		frames, e := peer.until(3*time.Second, messageKind("turn_ended"))
		if e != nil || frames[len(frames)-1]["error"] == nil || frames[len(frames)-1]["error"] == "" {
			return fmt.Errorf("waiting turn did not end explicitly on interrupt: %v", e)
		}
		// A new owner turn can wait on the same attempt. Interrupt must detach
		// the old waiter, not kill useful work or finish the replacement turn.
		if err = peer.send("prompt", "Continue after interrupt, preserving the pending memory."); err != nil {
			return err
		}
		if err = wait(); err != nil {
			return err
		}
	}
	unpark()
	frames, err := peer.until(3*time.Second, messageKind("turn_ended"))
	if err != nil {
		return fmt.Errorf("near-window turn did not resolve after worker: %w", err)
	}
	problem, _ := frames[len(frames)-1]["error"].(string)
	if mode == "failure" {
		if !strings.Contains(strings.ToLower(problem), "compaction") || len(memoryTurns(provider.Requests())) != 1 {
			return fmt.Errorf("failed held worker did not fail turn before another render: %q", problem)
		}
	} else if problem != "" {
		return fmt.Errorf("successful held worker failed owner: %s", problem)
	}
	if err = memoryDrain(peer, dir); err != nil {
		return err
	}
	wantState := "committed"
	if mode == "failure" {
		wantState = "failed"
	}
	records, err := memoryRecords(dir)
	if err != nil {
		return err
	}
	found := false
	for _, raw := range records {
		var event map[string]any
		if err = json.Unmarshal(raw, &event); err != nil {
			return err
		}
		outcome := object(event["compaction_outcome"])
		if event["type"] == "compaction_outcome" && outcome["task_id"] == pending[0] {
			if outcome["state"] != wantState {
				return fmt.Errorf("near-window attempt outcome: %v, want %s", outcome, wantState)
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("near-window attempt lacks correlated outcome")
	}
	// Failure must end the held turn first; only an explicit new prompt verifies
	// retained source. It must not automatically render the failed continuation.
	if mode == "failure" {
		if err = peer.send("prompt", "Verify retained source after the failed attempt, without a checkpoint."); err != nil {
			return err
		}
		if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
			return err
		}
	}
	requests := provider.Requests()
	if len(memoryTurns(requests)) != 2 || len(requests) != 3 {
		return fmt.Errorf("near-window completion lost bounded request counts: %d total, %d main", len(requests), len(memoryTurns(requests)))
	}
	last := memoryConversation(memoryLast(contextOut{reqs: requests}))
	failed := mode == "failure"
	if strings.Contains(last, candidate) == failed || strings.Contains(last, "END-WORK-0.") != failed {
		return fmt.Errorf("near-window continuation used the wrong replacement/source bytes")
	}
	return managementEOF(p)
}
