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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// memoryLifetime keeps the owner idle with an actual pending child, then applies
// an explicit kill, owner shutdown, or a real filesystem permission failure.
func memoryLifetime(bin, dir, mode string) error {
	skill := "---\nname: base\ndescription: Memory lifetime fixture\ntype: primary\ntools: micro_handoff kill_agent\n---\nFinish main work in chat.\n"
	if err := recallWrite(filepath.Join(dir, "skills"), map[string]string{"base/SKILL.md": skill}); err != nil {
		return err
	}
	const candidate = "I retained LIFETIME_MEMORY_15 and its decision reason."
	var mu sync.Mutex
	childID := ""
	replies := append(memoryTalk(1), fakevendor.Reply{Text: "Retained source verified.", Usage: fakevendor.Canonical{Input: 100}})
	provider := fakevendor.NewWithOptions(replies, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		if memoryCompressor(body) {
			return &fakevendor.Reply{ToolName: "submit", ToolID: "lifetime-memory", ToolArgs: memoryJSON(candidate), Usage: fakevendor.Canonical{Input: 100, Output: 20}}
		}
		if mode == "kill" && strings.Contains(memoryConversation(body), "KILL_PENDING_MEMORY_15") && !fleetHas(body, "kill-pending-memory") {
			mu.Lock()
			args, _ := json.Marshal(map[string]string{"agent_id": childID})
			mu.Unlock()
			return &fakevendor.Reply{ToolName: "kill_agent", ToolID: "kill-pending-memory", ToolArgs: string(args), Usage: fakevendor.Canonical{Input: 100}}
		}
		return nil
	}})
	defer provider.Close()
	entered, release, canceled := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var enteredOnce, releaseOnce, canceledOnce sync.Once
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
		if memoryCompressor(raw) {
			// No blocking route under the shared fake's capture mutex. Canceling
			// this real transport is observable even though its reply was prepared.
			proxy.ModifyResponse = func(*http.Response) error {
				enteredOnce.Do(func() { close(entered) })
				select {
				case <-release:
					return nil
				case <-req.Context().Done():
					canceledOnce.Do(func() { close(canceled) })
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
	if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": memorySettings(5000)}); err != nil {
		return err
	}
	if _, err = readSettings(peer, "settings_changed"); err != nil {
		return err
	}
	if err = peer.send("prompt", memoryPrompts(1)[0]); err != nil {
		return err
	}
	if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
		return fmt.Errorf("owner did not finish while worker pending: %w", err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		return fmt.Errorf("lifetime worker never reached held external response")
	}
	pending, err := memoryPending(dir)
	if err != nil || len(pending) != 1 {
		return fmt.Errorf("lifetime pending attempt: %v (%v)", pending, err)
	}
	frames, err := managementSelect(peer, "main")
	if err != nil {
		return err
	}
	for _, frame := range frames {
		if frame["type"] == "agent_tree" {
			children := list(managementNode(frame["root"], "main")["children"])
			if len(children) == 1 {
				mu.Lock()
				childID, _ = object(children[0])["id"].(string)
				mu.Unlock()
			}
		}
	}
	if childID == "" {
		return fmt.Errorf("pending worker is absent from its owner's tree")
	}
	wantState := "canceled"
	switch mode {
	case "kill":
		if err = peer.send("prompt", "KILL_PENDING_MEMORY_15: explicitly kill the pending child."); err != nil {
			return err
		}
		if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
			return fmt.Errorf("explicit child kill did not finish: %w", err)
		}
	case "shutdown":
		// EOF shuts the real owner down while its transport is still held.
		// The final save must include the drained attempt's terminal outcome.
		if err = managementEOF(p); err != nil {
			return err
		}
	case "write":
		wantState = "failed"
		corpus := filepath.Join(dir, "cr")
		info, e := os.Stat(corpus)
		if e != nil {
			return e
		}
		if err = os.Chmod(corpus, 0500); err != nil {
			return err
		}
		defer os.Chmod(corpus, info.Mode().Perm())
		// The existing child directory remains writable; creating a new corpus
		// candidate cannot succeed. Prove permissions actually apply on this host.
		probe := filepath.Join(corpus, "permission-probe")
		if e = os.WriteFile(probe, []byte("must fail"), 0600); !os.IsPermission(e) {
			os.Remove(probe)
			return fmt.Errorf("destination permission fixture is ineffective: %v", e)
		}
		unpark()
		// Settings persistence also needs this directory. Observe the actual
		// terminal record before restoring permission and using that GUI barrier.
		deadline := time.Now().Add(3 * time.Second)
		for {
			remaining, e := memoryPending(dir)
			if e != nil {
				return e
			}
			if len(remaining) == 0 {
				break
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("destination failure did not reach a terminal outcome")
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err = os.Chmod(corpus, info.Mode().Perm()); err != nil {
			return err
		}
	}
	if mode != "write" {
		select {
		case <-canceled:
		case <-time.After(3 * time.Second):
			return fmt.Errorf("%s did not cancel the actual worker transport", mode)
		}
	}
	if mode != "shutdown" {
		if err = memoryDrain(peer, dir); err != nil {
			return err
		}
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
			reason, _ := outcome["reason"].(string)
			if outcome["state"] != wantState || strings.TrimSpace(reason) == "" || mode == "write" && !strings.Contains(strings.ToLower(reason), "permission") {
				return fmt.Errorf("%s lacks reasoned %s outcome: %v", mode, wantState, outcome)
			}
			found = true
		}
	}
	if !found {
		return fmt.Errorf("%s lost the owned terminal outcome", mode)
	}
	if mode == "shutdown" {
		p, peer, err = managementStart(bin, dir, boundary.URL)
		if err != nil {
			return err
		}
		defer p.close()
		defer peer.close()
	}
	frames, err = managementSelect(peer, "main")
	if err != nil {
		return err
	}
	for _, frame := range frames {
		if frame["type"] == "agent_tree" && len(list(managementNode(frame["root"], "main")["children"])) != 0 {
			return fmt.Errorf("%s retained a dismissed worker in the tree", mode)
		}
	}
	if mode != "kill" {
		if err = peer.send("prompt", "Verify retained source after the failed attempt without a checkpoint."); err != nil {
			return err
		}
		if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
			return err
		}
	}
	requests := provider.Requests()
	mainCount := 3
	if mode == "kill" {
		mainCount = 4
		check := grade.Check{Passed: true}
		fleetResult(&check, requests, "kill-pending-memory", false)
		if !check.Passed {
			return fmt.Errorf("actual kill result: %s", strings.Join(check.Details, "; "))
		}
	}
	if len(memoryTurns(requests)) != mainCount || len(requests) != mainCount+1 {
		return fmt.Errorf("%s retried or lost work: %d requests, %d main", mode, len(requests), len(memoryTurns(requests)))
	}
	last := memoryConversation(memoryLast(contextOut{reqs: requests}))
	if strings.Contains(last, candidate) || !strings.Contains(last, "Work item 0.") || !strings.Contains(last, "END-WORK-0.") {
		return fmt.Errorf("%s changed the retained source in the next request", mode)
	}
	active, err := filepath.Glob(filepath.Join(dir, "cr", "memory", "*.md"))
	if err != nil || len(active) != 0 {
		return fmt.Errorf("%s left an uncommitted active memory: %v (%v)", mode, active, err)
	}
	return managementEOF(p)
}
