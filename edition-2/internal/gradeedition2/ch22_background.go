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
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// memoryBackground exercises the submitted executable, two continuing owner
// turns and real GUI selection while the external worker response is held. The
// provider then guesses a forbidden tool before correcting to a valid submit.
func memoryBackground(bin, dir string) error {
	skill := "---\nname: base\ndescription: Background memory fixture\ntype: primary\ntools: micro_handoff read_file write_file\n---\nMAIN CODING ROLE. Finish main work in chat; delegated tasks submit.\n"
	if err := recallWrite(filepath.Join(dir, "skills"), map[string]string{"base/SKILL.md": skill}); err != nil {
		return err
	}
	const memory = "I kept the selected decision because its reason avoids repeating the failure."
	// A real file is the forbidden effect. A refusal label without enforcement
	// would still leave this evidence, even if the next submission succeeds.
	forbidden := filepath.Join(dir, "forbidden-worker-write")
	args, _ := json.Marshal(map[string]string{"path": forbidden, "content": "must not exist"})
	mainReplies := append(memoryTalk(2), fakevendor.Reply{Text: "Verified after completion.", Usage: fakevendor.Canonical{Input: 100}})
	provider := fakevendor.NewWithOptions(mainReplies, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		if !memoryCompressor(body) {
			return nil
		}
		if !fleetHas(body, "forged-memory-write") {
			return &fakevendor.Reply{Text: "I will write the memory.", ToolName: "write_file", ToolID: "forged-memory-write", ToolArgs: string(args), Usage: fakevendor.Canonical{Input: 80, Output: 10}}
		}
		return &fakevendor.Reply{Text: "I will submit the memory instead.", ToolName: "submit", ToolID: "memory-finish", ToolArgs: memoryJSON(memory), Usage: fakevendor.Canonical{Input: 90, Output: 20}}
	}})
	defer provider.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	var enterOnce, releaseOnce sync.Once
	unpark := func() { releaseOnce.Do(func() { close(release) }) }
	upstream, _ := url.Parse(provider.URL())
	forward := httputil.NewSingleHostReverseProxy(upstream)
	boundary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		raw, err := io.ReadAll(request.Body)
		request.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		request.Body = io.NopCloser(bytes.NewReader(raw))
		proxy := *forward
		if memoryCompressor(raw) {
			// Capture and response generation finish before this hold. Keeping the
			// wait outside fakevendor is what lets the real main Agent progress.
			proxy.ModifyResponse = func(*http.Response) error {
				wait := false
				enterOnce.Do(func() { close(entered); wait = true })
				if wait {
					select {
					case <-release:
					case <-request.Context().Done():
						return request.Context().Err()
					}
				}
				return nil
			}
		}
		proxy.ServeHTTP(w, request)
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
	// Both turns must finish while the first worker HTTP response is still held.
	// The second checkpoint must not create another attempt for the same owner.
	for i, prompt := range memoryPrompts(2) {
		if err = peer.send("prompt", prompt); err != nil {
			return err
		}
		frames, e := peer.until(3*time.Second, messageKind("turn_ended"))
		if e != nil {
			return fmt.Errorf("main turn %d blocked behind held worker: %w", i+1, e)
		}
		if problem, _ := frames[len(frames)-1]["error"].(string); problem != "" {
			return fmt.Errorf("main turn %d failed: %s", i+1, problem)
		}
		if i == 0 {
			select {
			case <-entered:
			case <-time.After(3 * time.Second):
				return fmt.Errorf("no memory request reached held external peer")
			}
		}
	}
	requests := provider.Requests()
	if len(requests)-len(memoryTurns(requests)) != 1 {
		return fmt.Errorf("repeated checkpoint did not retain exactly one worker request")
	}
	frames, err := managementSelect(peer, "main")
	if err != nil {
		return err
	}
	var owner map[string]any
	// Selection's public snapshot supplies ownership and a selectable identity;
	// guessing a generated worker name would test an implementation detail.
	for _, frame := range frames {
		if frame["type"] == "agent_tree" {
			owner = managementNode(frame["root"], "main")
		}
	}
	children := list(owner["children"])
	if len(children) != 1 {
		return fmt.Errorf("held owner has %d visible children, want one", len(children))
	}
	child := object(children[0])
	key, _ := child["key"].(string)
	if key == "" || child["parent"] != "main" || child["persistent"] != false {
		return fmt.Errorf("worker is not an owned task child: %v", child)
	}
	if _, err = managementSelect(peer, key); err != nil {
		return fmt.Errorf("select held worker: %w", err)
	}
	unpark()
	// Removing a child from a tree alone is insufficient. The selected browser
	// must return to the owner and receive its settings before issuing more work.
	if _, err = peer.until(3*time.Second, func(frame map[string]any) bool {
		return frame["type"] == "agent_selected" && frame["key"] == "main"
	}); err != nil {
		return fmt.Errorf("completed worker selection did not fall back to owner: %w", err)
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return fmt.Errorf("owner fallback lacks settings: %w", err)
	}
	if err = memoryDrain(peer, dir); err != nil {
		return err
	}
	frames, err = managementSelect(peer, "main")
	if err != nil {
		return err
	}
	for _, frame := range frames {
		if frame["type"] == "agent_tree" && len(list(managementNode(frame["root"], "main")["children"])) != 0 {
			return fmt.Errorf("completed worker retained in owner tree")
		}
	}
	if _, err = os.Stat(forbidden); !os.IsNotExist(err) {
		return fmt.Errorf("forbidden worker write was not prevented: %v", err)
	}
	if err = peer.send("prompt", "Verify accepted memory and retained unrelated work without another checkpoint."); err != nil {
		return err
	}
	if _, err = peer.until(3*time.Second, messageKind("turn_ended")); err != nil {
		return err
	}
	if err = managementEOF(p); err != nil {
		return err
	}
	requests = provider.Requests()
	// Routing needs only the memory property; grading checks the entire published
	// schema. A permissive extra-property contract must not pass by route alone.
	var schema any
	json.Unmarshal([]byte(`{"type":"object","properties":{"memory":{"type":"string"}},"required":["memory"],"additionalProperties":false}`), &schema)
	workers := 0
	refused := false
	for _, request := range requests {
		if !memoryCompressor(request.Body) {
			continue
		}
		workers++
		decl := object(list(request.JSON()["tools"])[0])
		if !reflect.DeepEqual(decl["input_schema"], schema) {
			return fmt.Errorf("worker declaration differs from the exact memory contract")
		}
		if strings.Contains(memorySystem(request.Body), "MAIN CODING ROLE") {
			return fmt.Errorf("worker inherited the main primary role")
		}
		if _, isError, e := anthropicToolText(request.Body, "forged-memory-write"); e == nil && isError {
			refused = true
		}
	}
	if workers != 2 || !refused {
		return fmt.Errorf("forged tool refusal/correction missing: requests%d error-result%v", workers, refused)
	}
	last := memoryConversation(memoryLast(contextOut{reqs: requests}))
	// The fixed response proves delivery, not semantic summarization. Its selected
	// source disappears while work added during the held response stays visible.
	if !strings.Contains(last, memory) || strings.Contains(last, "END-WORK-0.") || !strings.Contains(last, "END-WORK-1.") {
		return fmt.Errorf("accepted selected replacement or retained newer dialogue is wrong")
	}
	return nil
}
