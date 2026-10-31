package gradeedition2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Memory uses the same public GUI completion/settings boundaries as Chapter14.
// Auxiliary compressor requests are routed separately from scripted main replies.
type memoryOpts struct {
	dir, model string
	prompts    []string
	expect     []int
	replies    []fakevendor.Reply
	settings   map[string]any
	route      func([]byte) *fakevendor.Reply
	// pending is signaled after the provider has handled an actual graduation.
	// Its response is held outside the capture lock until release or cancellation.
	pending chan struct{}
	release <-chan struct{}
	unblock func()
	// lengthStop marks an otherwise complete worker reply as provider-truncated.
	lengthStop bool
	// accounting finishes the worker after the main reply, then samples the GUI
	// meter before another main response could conceal a corrupted Last value.
	accounting bool
}

// Turn-ended acknowledges the main turn only. Correlated durable outcomes below
// establish worker completion before verification turns or EOF shuts its owner.
func memoryLaunch(bin, skillsDir, guiDir string, o memoryOpts) (out contextOut) {
	srv := fakevendor.NewWithOptions(o.replies, fakevendor.Options{Route: o.route})
	defer srv.Close()
	upstream, _ := url.Parse(srv.URL())
	forward := httputil.NewSingleHostReverseProxy(upstream)
	var held sync.Once
	mainFinished := make(chan struct{})
	var releaseMain sync.Once
	defer releaseMain.Do(func() { close(mainFinished) })
	boundary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		req.Body = io.NopCloser(bytes.NewReader(raw))
		proxy := *forward
		if o.pending != nil && memoryCompressor(raw) && strings.Count(memoryConversation(raw), "MEMO-") >= 2 {
			// The fake has recorded and answered before this hook runs. Holding its
			// HTTP response here leaves other requests and capture inspection free.
			proxy.ModifyResponse = func(*http.Response) error {
				wait := false
				held.Do(func() { close(o.pending); wait = true })
				if wait {
					select {
					case <-o.release:
					case <-req.Context().Done():
						return req.Context().Err()
					}
				}
				return nil
			}
		}
		if o.accounting && memoryCompressor(raw) {
			proxy.ModifyResponse = func(*http.Response) error {
				select {
				case <-mainFinished:
					return nil
				case <-req.Context().Done():
					return req.Context().Err()
				}
			}
		}
		if o.lengthStop && memoryCompressor(raw) {
			// Preserve valid tool arguments and measured usage. Truncation is the
			// provider's terminal fact, even when its complete JSON happens to parse.
			proxy.ModifyResponse = func(response *http.Response) error {
				data, err := io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil {
					return err
				}
				var body map[string]any
				if err = json.Unmarshal(data, &body); err != nil {
					return err
				}
				body["stop_reason"] = "max_tokens"
				data, err = json.Marshal(body)
				if err != nil {
					return err
				}
				response.Body = io.NopCloser(bytes.NewReader(data))
				response.ContentLength = int64(len(data))
				response.Header.Set("Content-Length", fmt.Sprint(len(data)))
				return nil
			}
		}
		proxy.ServeHTTP(w, req)
	}))
	defer boundary.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+o.model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+boundary.URL, "EN_DISABLE_STREAMING=1", "EN_SKILLS_DIR="+skillsDir, "EN_PRIMARY_SKILL=base", "CH02_LOG="+filepath.Join(o.dir, "events.jsonl"))
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.Discard, &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		return contextOut{fatal: err.Error()}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	// Join the real child before inspecting capture state. A deliberately held
	// provider reply is then released so fake-server cleanup cannot hang.
	defer func() {
		stdin.Close()
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		if o.unblock != nil {
			o.unblock()
		}
		out.reqs = srv.Requests()
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
			return contextOut{fatal: fmt.Sprintf("GUI exited before ready: %v", e)}
		default:
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		return contextOut{fatal: "GUI socket did not start: " + err.Error()}
	}
	defer peer.close()
	if err = peer.send("subscribe", ""); err != nil {
		return contextOut{fatal: err.Error()}
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		return contextOut{fatal: err.Error()}
	}
	if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": o.settings}); err != nil {
		return contextOut{fatal: err.Error()}
	}
	settings, err := readSettings(peer, "settings_changed")
	if err != nil {
		return contextOut{fatal: err.Error()}
	}
	if e := memorySettingsMatch(settings, o.settings); e != "" {
		return contextOut{fatal: "effective settings: " + e}
	}
	for i, prompt := range o.prompts {
		if err = json.NewEncoder(stdin).Encode(map[string]string{"kind": "prompt", "text": prompt}); err != nil {
			return contextOut{fatal: err.Error()}
		}
		// Kill only after the external peer has received an actual graduation.
		// The peer holds that response until cleanup, so no completion can race
		// ahead of this crash observation.
		if o.pending != nil && i == len(o.prompts)-1 {
			select {
			case <-o.pending:
				if err = cmd.Process.Kill(); err != nil {
					return contextOut{fatal: err.Error()}
				}
				<-done
				finished = true
				return out
			case e := <-done:
				finished = true
				return contextOut{fatal: fmt.Sprintf("exited before pending fold: %v", e)}
			case <-time.After(10 * time.Second):
				return contextOut{fatal: "no pending graduation reached external peer"}
			}
		}
		messages, e := peer.until(20*time.Second, messageKind("turn_ended"))
		if e != nil {
			return contextOut{fatal: fmt.Sprintf("prompt %d completion: %v", i+1, e)}
		}
		ending := messages[len(messages)-1]
		if problem, ok := ending["error"].(string); ok && problem != "" {
			return contextOut{fatal: "turn failed: " + problem}
		}
		want := i + 1
		if i < len(o.expect) {
			want = o.expect[i]
		}
		if n := len(memoryTurns(srv.Requests())); n != want {
			return contextOut{fatal: fmt.Sprintf("completed prompt %d made %d requests, want %d", i+1, n, want)}
		}
		releaseMain.Do(func() { close(mainFinished) })
		if err = memoryDrain(peer, o.dir); err != nil {
			return contextOut{fatal: fmt.Sprintf("prompt %d memory completion: %v", i+1, err)}
		}
		if o.accounting {
			// Reselecting asks the real owner for its current meter without inference.
			// The preceding drain/settings barrier consumed older replay trailers.
			if err = peer.send("select_agent", ""); err != nil {
				return contextOut{fatal: err.Error()}
			}
			frames, err := peer.until(5*time.Second, func(m map[string]any) bool {
				return m["type"] == "usage" && m["replay"] == true
			})
			if err != nil {
				return contextOut{fatal: "fresh accounting meter: " + err.Error()}
			}
			out.meters = append(out.meters, frames[len(frames)-1])
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

// Read the journal before its snapshot: a concurrent checkpoint installs the
// snapshot before trimming the journal. Union by task identity tolerates overlap.
// Only complete JSON-lines count; an in-progress tail cannot acknowledge work.
// These fixtures retain the complete public log; this is not a log-retention reader.
func memoryRecords(dir string) ([]json.RawMessage, error) {
	path := filepath.Join(dir, "save.json")
	tail, err := os.ReadFile(path + ".journal")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var save struct {
		// Log is the public durable event sequence, not the disposable projection.
		Log []json.RawMessage `json:"log"`
	}
	// A fresh ordinary turn need not have checkpointed a save yet. Its empty
	// launch set is valid; a real managed launch must publish its durable record.
	if len(data) > 0 {
		if err = json.Unmarshal(data, &save); err != nil {
			return nil, err
		}
	}
	if end := bytes.LastIndexByte(tail, '\n'); end >= 0 {
		for _, line := range bytes.Split(tail[:end], []byte{'\n'}) {
			if len(bytes.TrimSpace(line)) != 0 {
				save.Log = append(save.Log, line)
			}
		}
	}
	return save.Log, nil
}

func memoryPending(dir string) ([]string, error) {
	records, err := memoryRecords(dir)
	if err != nil {
		return nil, err
	}
	launched, finished := map[string]bool{}, map[string]bool{}
	for _, raw := range records {
		var event map[string]any
		if err = json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		switch event["type"] {
		case "compactor_launched":
			id, _ := object(event["compactor"])["task_id"].(string)
			if id == "" {
				return nil, fmt.Errorf("compactor launch lacks background task correlation")
			}
			launched[id] = true
		case "compaction_outcome":
			outcome := object(event["compaction_outcome"])
			id, _ := outcome["task_id"].(string)
			state, _ := outcome["state"].(string)
			if id == "" || (state != "committed" && state != "failed" && state != "canceled" && state != "invalidated" && state != "abandoned") {
				return nil, fmt.Errorf("invalid terminal compaction outcome: %s", raw)
			}
			reason, _ := outcome["reason"].(string)
			if state != "committed" && strings.TrimSpace(reason) == "" {
				return nil, fmt.Errorf("failed compaction outcome lacks reason: %s", raw)
			}
			finished[id] = true
		}
	}
	var pending []string
	for id := range launched {
		if !finished[id] {
			pending = append(pending, id)
		}
	}
	return pending, nil
}

// A terminal outcome may immediately launch the next waterfall step. An empty
// settings patch crosses the existing actor settings boundary, then a second
// durable read checks that successor. Neither a sleep nor turn-ended is success.
func memoryDrain(peer *guiPeer, dir string) error {
	deadline := time.Now().Add(10 * time.Second)
	for {
		pending, err := memoryPending(dir)
		if err != nil {
			return err
		}
		if len(pending) == 0 {
			if err = peer.conn.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]any{}}); err != nil {
				return err
			}
			if _, err = readSettings(peer, "settings_changed"); err != nil {
				return err
			}
			pending, err = memoryPending(dir)
			if err != nil || len(pending) == 0 {
				return err
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("no terminal compaction outcome for %v", pending)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A full snapshot may include defaults absent from the patch. Every supplied
// leaf must match; seeing the request echo alone would not prove application.
func memorySettingsMatch(got, want map[string]any) string {
	for k, v := range want {
		if nested, ok := v.(map[string]any); ok {
			g, ok := got[k].(map[string]any)
			if !ok {
				return k + " missing"
			}
			if e := memorySettingsMatch(g, nested); e != "" {
				return k + "." + e
			}
			continue
		}
		a, _ := json.Marshal(got[k])
		b, _ := json.Marshal(v)
		if !bytes.Equal(a, b) {
			return fmt.Sprintf("%s=%s, want %s", k, a, b)
		}
	}
	return ""
}

// Provider content can be a string or text blocks. Tool input/result echoes are
// deliberately excluded: a submitted memory is not evidence it was populated.
func memoryText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var blocks []struct{ Type, Text string }
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var b strings.Builder
	for _, p := range blocks {
		if p.Type == "text" {
			b.WriteString(p.Text)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
func memoryConversation(body []byte) string {
	var r struct {
		Messages []struct{ Content json.RawMessage }
	}
	if json.Unmarshal(body, &r) != nil {
		return ""
	}
	var b strings.Builder
	for _, m := range r.Messages {
		b.WriteString(memoryText(m.Content))
		b.WriteByte('\n')
	}
	return b.String()
}
func memorySystem(body []byte) string {
	var r struct {
		System   json.RawMessage
		Messages []struct {
			Role    string
			Content json.RawMessage
		}
	}
	json.Unmarshal(body, &r)
	text := memoryText(r.System)
	// Instruction placement includes inline system messages, not only the top
	// field. Otherwise moving memory into that channel could satisfy both the
	// presence assertion and a misleading "not in system" absence assertion.
	for _, message := range r.Messages {
		if message.Role == "system" {
			text += "\n" + memoryText(message.Content)
		}
	}
	return text
}
func memoryTools(body []byte) []string {
	var r struct{ Tools []struct{ Name string } }
	json.Unmarshal(body, &r)
	var names []string
	for _, t := range r.Tools {
		names = append(names, t.Name)
	}
	return names
}
func memoryCompressor(body []byte) bool {
	var request struct {
		// Tools is the declaration list actually sent to the provider.
		Tools []struct {
			// Name distinguishes the completion tool from arbitrary capabilities.
			Name string
			// Schema distinguishes memory completion from an ordinary text result.
			Schema struct {
				// Properties contains the direct submit argument definitions.
				Properties map[string]struct {
					// Type requires memory text, not merely a similarly named property.
					Type string
				}
			} `json:"input_schema"`
		}
	}
	if json.Unmarshal(body, &request) != nil || len(request.Tools) != 1 {
		return false
	}
	t := request.Tools[0]
	return t.Name == "submit" && t.Schema.Properties["memory"].Type == "string"
}
func memoryTurns(reqs []fakevendor.Recorded) []fakevendor.Recorded {
	var out []fakevendor.Recorded
	for _, q := range reqs {
		if !memoryCompressor(q.Body) {
			out = append(out, q)
		}
	}
	return out
}
func memoryLast(out contextOut) []byte {
	q := memoryTurns(out.reqs)
	if len(q) == 0 {
		return nil
	}
	return q[len(q)-1].Body
}
func memoryJSON(text string) string {
	b, _ := json.Marshal(map[string]string{"memory": text})
	return string(b)
}

const memoryPlant = "ZORBLAX-7741"

// Each source memory is large enough to cross the chosen watermarks, but has
// only one identifier. A fold must leave newer files to distinguish direction.
func memoryMemo(n int) string {
	return fmt.Sprintf("MEMO-%02d I finished this task. %s", n, strings.Repeat("I retained the verified detail. ", 6))
}
func memoryFold(n int) string { return fmt.Sprintf("FOLD-%02d I retained the earlier decisions.", n) }

type memoryRouter struct {
	mu                 sync.Mutex
	n, folds           int
	bodies, foldBodies [][]byte
	firstFoldN         int
	failFold           bool
	pending            chan struct{}
	release            chan struct{}
	once               sync.Once
	usage              fakevendor.Canonical
}

func (r *memoryRouter) unblock() {
	r.once.Do(func() {
		if r.release != nil {
			close(r.release)
		}
	})
}

// Only the public submit-only request is auxiliary. Main tool declarations and
// requests remain real. Unique source markers distinguish session compression
// from graduation without recognizing implementation-specific prompt wording.
func (r *memoryRouter) route(body []byte) *fakevendor.Reply {
	if !memoryCompressor(body) {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bodies = append(r.bodies, append([]byte(nil), body...))
	source := memoryConversation(body)
	var text, id string
	if strings.Count(source, "MEMO-") >= 2 {
		r.folds++
		r.foldBodies = append(r.foldBodies, append([]byte(nil), body...))
		if r.folds == 1 {
			r.firstFoldN = r.n
		}
		if r.failFold {
			return &fakevendor.Reply{Status: 500, ErrBody: `{"error":"graduation unavailable"}`}
		}
		text = memoryFold(r.folds)
		id = fmt.Sprintf("fold-%d", r.folds)
	} else {
		r.n++
		text = memoryMemo(r.n)
		id = fmt.Sprintf("memo-%d", r.n)
	}
	if strings.Contains(source, memoryPlant) {
		text += " I remember " + memoryPlant + "."
	}
	usage := r.usage
	if usage == (fakevendor.Canonical{}) {
		usage = fakevendor.Canonical{Input: 100, Output: 20}
	}
	return &fakevendor.Reply{ToolName: "submit", ToolID: id, ToolArgs: memoryJSON(text), Usage: usage}
}
func memoryTalk(n int) []fakevendor.Reply {
	var out []fakevendor.Reply
	for i := 0; i < n; i++ {
		out = append(out, fakevendor.Reply{ToolName: "micro_handoff", ToolID: fmt.Sprintf("checkpoint-%d", i), ToolArgs: fmt.Sprintf(`{"text":"Checkpoint %d. I finished this work and will continue."}`, i), Usage: fakevendor.Canonical{Input: 200, Output: 30}}, fakevendor.Reply{Text: fmt.Sprintf("Step %d recorded.", i), Usage: fakevendor.Canonical{Input: 220, Output: 20}})
	}
	return out
}
func memoryPrompts(n int) []string {
	var out []string
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("Work item %d. %s END-WORK-%d.", i, strings.Repeat("Keep the exact decision and its reason. ", 12), i))
	}
	return out
}

// Explicit false matters on a persisted restart: omission would retain the
// preceding disabled value and falsely turn a re-enable experiment into no-op.
func memorySettings(budget int, off ...string) map[string]any {
	m := map[string]any{"conversation": map[string]any{"budget": 300}, "session": map[string]any{"budget": budget}, "8x": map[string]any{"budget": 5000}, "64x": map[string]any{"budget": 5000}}
	for _, v := range m {
		v.(map[string]any)["disabled"] = false
	}
	for _, band := range off {
		m[band].(map[string]any)["disabled"] = true
	}
	return map[string]any{"context_target": 200000, "memory": m}
}
