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
)

// A real graduation selects two of three public corpus files. Changing their
// bytes or turning the destination off/on while the response is held must make
// this particular candidate stale, even when identities and final settings match.
func memoryStale(bin, dir string, edit bool) error {
	const candidate = "STALE_CANDIDATE_15 must not replace the source."
	// Each file is 202 bytes: total606 exceeds high450, and retiring two
	// leaves202 below low225. This makes oldest-prefix selection observable.
	files := map[string]string{}
	for i := 1; i <= 3; i++ {
		files[fmt.Sprintf("2026-10-01-%d.md", i)] = fmt.Sprintf("SOURCE_%d. %s", i, strings.Repeat("I retained this reason. ", 8))
	}
	if err := recallWrite(filepath.Join(dir, "cr", "memory"), files); err != nil {
		return err
	}
	if err := recallWrite(filepath.Join(dir, "skills"), map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Stale memory fixture\ntype: primary\ntools: micro_handoff\n---\nFinish main work in chat.\n"}); err != nil {
		return err
	}
	// Fixed candidate text measures admission and delivery, not compression
	// quality. The actual runtime must accept or refuse its otherwise valid shape.
	provider := fakevendor.NewWithOptions(append(memoryTalk(1), fakevendor.Reply{Text: "Verified retained source."}), fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		if memoryCompressor(body) {
			return &fakevendor.Reply{ToolName: "submit", ToolID: "stale-memory", ToolArgs: memoryJSON(candidate), Usage: fakevendor.Canonical{Input: 100, Output: 20}}
		}
		return nil
	}})
	defer provider.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var enterOnce, releaseOnce sync.Once
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
			proxy.ModifyResponse = func(*http.Response) error {
				enterOnce.Do(func() { close(entered) })
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
	peer.send("subscribe", "")
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return err
	}
	settings := memorySettings(225)
	object(settings["memory"])["conversation"] = map[string]any{"budget": 100000}
	if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": settings}); err != nil {
		return err
	}
	if _, err = readSettings(peer, "settings_changed"); err != nil {
		return err
	}
	peer.send("prompt", "Checkpoint existing memory files without adding a long conversation.")
	if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
		return fmt.Errorf("main blocked by held graduation: %w", err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		return fmt.Errorf("no actual graduation reached the held external boundary")
	}
	pending, err := memoryPending(dir)
	if err != nil || len(pending) != 1 {
		return fmt.Errorf("held graduation lacks exactly one durable launch: %v (%v)", pending, err)
	}
	// Establish the selected bytes before mutating anything. The third file is
	// a retained control, not another selected source hidden by the fixture.
	workers := 0
	for _, r := range provider.Requests() {
		if memoryCompressor(r.Body) {
			workers++
			text := memoryConversation(r.Body)
			if !strings.Contains(text, files["2026-10-01-1.md"]) || !strings.Contains(text, files["2026-10-01-2.md"]) || strings.Contains(text, "SOURCE_3.") {
				return fmt.Errorf("graduation did not select the distinguishing oldest prefix")
			}
		}
	}
	if workers != 1 {
		return fmt.Errorf("held graduation made %d worker requests", workers)
	}
	if edit {
		files["2026-10-01-1.md"] += "\nHUMAN_CORRECTION_15: retain these exact edited bytes."
		if err = os.WriteFile(filepath.Join(dir, "cr", "memory", "2026-10-01-1.md"), []byte(files["2026-10-01-1.md"]), 0600); err != nil {
			return err
		}
	} else {
		// Return to the original settings. Equality of current values must not
		// revive a candidate from before the relevant policy transition.
		for _, disabled := range []bool{true, false} {
			patch := map[string]any{"memory": map[string]any{"8x": map[string]any{"disabled": disabled}}}
			if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": patch}); err != nil {
				return err
			}
			effective, e := readSettings(peer, "settings_changed")
			if e != nil || memorySettingsMatch(effective, patch) != "" {
				return fmt.Errorf("off/on effective settings: %s (%v)", memorySettingsMatch(effective, patch), e)
			}
		}
	}
	// Policy changes may already cancel the transport. Releasing a prepared
	// response is not a claim that the canceled worker consumed it afterward.
	unpark()
	if err = memoryDrain(peer, dir); err != nil {
		return err
	}
	records, err := memoryRecords(dir)
	if err != nil {
		return err
	}
	// Snapshot and journal can overlap; inspect the captured task identity in
	// every complete copy rather than treating the latest unrelated event as proof.
	outcome := false
	for _, raw := range records {
		var event map[string]any
		if err = json.Unmarshal(raw, &event); err != nil {
			return err
		}
		value := object(event["compaction_outcome"])
		if event["type"] == "compaction_outcome" && value["task_id"] == pending[0] {
			reason, _ := value["reason"].(string)
			if value["state"] != "invalidated" || strings.TrimSpace(reason) == "" {
				return fmt.Errorf("stale candidate did not record reasoned invalidation: %v", value)
			}
			outcome = true
		}
	}
	if !outcome {
		return fmt.Errorf("no correlated terminal invalidation")
	}
	for name, want := range files {
		data, e := os.ReadFile(filepath.Join(dir, "cr", "memory", name))
		if e != nil || string(data) != want {
			return fmt.Errorf("stale candidate changed or retired source %s: %v", name, e)
		}
	}
	// The destination began empty. A rejected candidate must not leave an active
	// file that a later startup could mistake for independently committed memory.
	active, err := filepath.Glob(filepath.Join(dir, "cr", "bucket-0", "*.md"))
	if err != nil || len(active) != 0 {
		return fmt.Errorf("invalidated candidate left active destination files: %v (%v)", active, err)
	}
	peer.send("prompt", "Verify retained source without another checkpoint.")
	if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
		return err
	}
	// A same-name file edit does not implicitly rescan live context. The old
	// selected source stays visible here, while the human's disk bytes survive.
	last := memoryConversation(memoryLast(contextOut{reqs: provider.Requests()}))
	if !strings.Contains(last, "Verify retained source without another checkpoint.") || strings.Contains(last, candidate) || !strings.Contains(last, "SOURCE_1.") || !strings.Contains(last, "SOURCE_2.") || !strings.Contains(last, "SOURCE_3.") {
		return fmt.Errorf("stale candidate replaced the source in the next main request")
	}
	workers = 0
	for _, request := range provider.Requests() {
		if memoryCompressor(request.Body) {
			workers++
		}
	}
	if workers != 1 {
		return fmt.Errorf("invalidated attempt retried without a new checkpoint: %d requests", workers)
	}
	return managementEOF(p)
}
