package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

type recallOpts struct {
	dir, model            string
	prompts               []string
	expect                []int
	replies               []fakevendor.Reply
	judge                 string
	judgeError, judgeHang bool
	after                 func(int) error
}
type recallOut struct {
	reqs   []fakevendor.Recorded
	events []map[string]json.RawMessage
	fatal  string
}

// A blocked judge cannot hold the capture mutex or the main response hostage.
// This small external Anthropic peer leaves all recall/actor/rendering code real.
type recallPeer struct {
	*httptest.Server
	mu       sync.Mutex
	requests []fakevendor.Recorded
	main     int
	unblock  chan struct{}
	once     sync.Once
}

func newRecallPeer(o recallOpts) *recallPeer {
	p := &recallPeer{unblock: make(chan struct{})}
	p.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		judge := recallJudge(body)
		p.mu.Lock()
		p.requests = append(p.requests, fakevendor.Recorded{Vendor: "anthropic", Path: r.URL.Path, Body: body})
		n := p.main
		if !judge {
			p.main++
		}
		p.mu.Unlock()
		reply := fakevendor.Reply{Text: "Acknowledged."}
		if judge {
			if o.judgeHang {
				select {
				case <-p.unblock:
				case <-r.Context().Done():
					return
				}
			}
			if o.judgeError {
				http.Error(w, "judge unavailable", 500)
				return
			}
			reply.Text = o.judge
		} else if n < len(o.replies) {
			reply = o.replies[n]
		}
		parts := []any{}
		if reply.Text != "" {
			parts = append(parts, map[string]any{"type": "text", "text": reply.Text})
		}
		stop := "end_turn"
		if reply.ToolName != "" {
			var args any
			if json.Unmarshal([]byte(reply.ToolArgs), &args) != nil {
				http.Error(w, "invalid fixture args", 500)
				return
			}
			parts = append(parts, map[string]any{"type": "tool_use", "id": reply.ToolID, "name": reply.ToolName, "input": args})
			stop = "tool_use"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "recall-response", "type": "message", "role": "assistant", "model": o.model, "content": parts, "stop_reason": stop, "usage": map[string]int{"input_tokens": 50, "output_tokens": 5}})
	}))
	return p
}
func (p *recallPeer) release() { p.once.Do(func() { close(p.unblock) }) }

// Requests returns a stable copy of captured exchanges after brief locking.
// Callers own the slice; the peer never mutates a recorded request body.
func (p *recallPeer) Requests() []fakevendor.Recorded {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]fakevendor.Recorded(nil), p.requests...)
}
func recallJudge(body []byte) bool {
	var r struct{ Tools []json.RawMessage }
	return json.Unmarshal(body, &r) == nil && len(r.Tools) == 0
}
func recallTurns(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, r := range reqs {
		if !recallJudge(r.Body) {
			out = append(out, r)
		}
	}
	return out
}

// The existing settings snapshot establishes subscription readiness. No settings
// patch is needed: these fixtures exercise the chapter's published defaults.
// Turn-ended, not a request count or post-request sleep, establishes completion.
func recallLaunch(bin, skillsDir, guiDir string, o recallOpts) (out recallOut) {
	srv := newRecallPeer(o)
	defer srv.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return recallOut{fatal: err.Error()}
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+o.model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+srv.URL, "EN_DISABLE_STREAMING=1", "EN_SKILLS_DIR="+skillsDir, "EN_PRIMARY_SKILL=base", "CH02_LOG="+filepath.Join(o.dir, "events.jsonl"))
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.Discard, &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return recallOut{fatal: err.Error()}
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		return recallOut{fatal: err.Error()}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	// Join the real child before inspecting capture state. A held judge is
	// released independently: its timeout must not block a main HTTP response.
	defer func() {
		stdin.Close()
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		srv.release()
		out.reqs = srv.Requests()
		out.events, err = recallReadEvents(o.dir)
		if err != nil && out.fatal == "" {
			out.fatal = err.Error()
		}
		if out.fatal != "" {
			out.fatal += ": " + stderr.String()
		}

	}()
	var peer *guiPeer
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		peer, err = openGUIPeer("ws://localhost:" + port + "/ws")
		if err == nil {
			break
		}
		select {
		case e := <-done:
			finished = true
			return recallOut{fatal: fmt.Sprintf("GUI exited before ready: %v", e)}
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		return recallOut{fatal: "GUI socket did not start: " + err.Error()}
	}
	defer peer.close()
	if err = peer.send("subscribe", ""); err != nil {
		return recallOut{fatal: err.Error()}
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return recallOut{fatal: err.Error()}
	}
	for i, prompt := range o.prompts {
		if err = json.NewEncoder(stdin).Encode(map[string]string{"kind": "prompt", "text": prompt}); err != nil {
			return recallOut{fatal: err.Error()}
		}
		messages, e := peer.until(25*time.Second, messageKind("turn_ended"))
		if e != nil {
			return recallOut{fatal: fmt.Sprintf("prompt %d completion: %v", i+1, e)}
		}
		ending := messages[len(messages)-1]
		if problem, ok := ending["error"].(string); ok && problem != "" {
			return recallOut{fatal: "turn failed: " + problem}
		}
		if o.after != nil {
			if err := o.after(i); err != nil {
				return recallOut{fatal: err.Error()}
			}
		}
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if n := len(recallTurns(srv.Requests())); n != want {
			return recallOut{fatal: fmt.Sprintf("completed prompt %d made %d requests, want %d", i+1, n, want)}
		}
	}
	stdin.Close()
	select {
	case err = <-done:
		finished = true
		if err != nil {
			out.fatal = "nonzero GUI exit after EOF: " + err.Error()
		}
	case <-time.After(10 * time.Second):
		out.fatal = "GUI did not save and exit after EOF"
	}
	return out
}

// Both public GUI clients save the complete recorded Log. Reading that existing
// persistence contract avoids assuming that every client honors CH02_LOG.
func recallReadEvents(dir string) ([]map[string]json.RawMessage, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "save.json"))
	if err != nil {
		return nil, err
	}
	var saved struct{ Log []map[string]json.RawMessage }
	if err := json.Unmarshal(raw, &saved); err != nil {
		return nil, err
	}
	return saved.Log, nil
}
func recallAttachments(out recallOut) []string {
	var texts []string
	for _, e := range out.events {
		var kind string
		json.Unmarshal(e["type"], &kind)
		if kind == "recall_added" {
			texts = append(texts, recallParts(e["parts"]))
			continue
		}
		if kind != "recall_attached" {
			continue
		}
		var r struct{ Parts json.RawMessage }
		json.Unmarshal(e["recall_attached"], &r)
		texts = append(texts, recallParts(r.Parts))
	}
	return texts
}

// Match recorded attachment bytes inside actual text parts, not tool arguments,
// results or a hard-coded message role. Adjacent user data may share a message.
func recallWire(body []byte) string {
	var r struct {
		Messages []struct{ Content json.RawMessage }
	}
	json.Unmarshal(body, &r)
	var texts []string
	for _, m := range r.Messages {
		var s string
		if json.Unmarshal(m.Content, &s) == nil {
			if i := strings.Index(s, recallHeader); i >= 0 {
				texts = append(texts, s[i:])
			}
			continue
		}
		var parts []struct{ Type, Text string }
		json.Unmarshal(m.Content, &parts)
		for _, p := range parts {
			if p.Type == "text" {
				if i := strings.Index(p.Text, recallHeader); i >= 0 {
					texts = append(texts, p.Text[i:])
				}
			}
		}
	}
	return strings.Join(texts, "\n")
}

func recallParts(raw json.RawMessage) string {
	var parts []struct{ Type, Text string }
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range parts {
		if p.Type == "text" || p.Type == "recall" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}
