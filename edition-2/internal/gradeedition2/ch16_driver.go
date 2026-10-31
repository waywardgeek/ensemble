package gradeedition2

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"syscall"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

type subscriptionOpts struct {
	prompts                   []string
	oauth                     grade.Edition2OAuthOptions
	responses                 grade.Edition2ResponsesOptions
	skipLogin, gate           bool
	executableOnly, signedOut bool
}
type subscriptionOut struct {
	oauth                          grade.Edition2OAuthObservations
	responses                      grade.Edition2ResponsesObservations
	messages                       []map[string]any
	errors                         []string
	loginOutput, output, artifacts string
	loginErr, exitErr, fatal       string
	gated                          bool
}

// The proxy forwards the original peer's real SSE frames. It withholds the
// remainder after the first summary delta until the client displays that delta.
// Mere summary-before-text ordering after a buffered response cannot open it.
type subscriptionGate struct {
	mu             sync.Mutex
	active, opened bool
	release        chan struct{}
	once           sync.Once
	delivered      string
}

func (g *subscriptionGate) observe(m map[string]any) {
	if m["type"] != "part_delta" || m["kind"] != "reasoning_summary" {
		return
	}
	g.mu.Lock()
	if g.active {
		chunk, _ := m["chunk"].(string)
		g.delivered += chunk
		// Delivery may split a provider delta across several observations.
		if g.delivered == "Reading " {
			g.opened = true
			g.once.Do(func() { close(g.release) })
		}
	}
	g.mu.Unlock()
}
func (g *subscriptionGate) proxy(target string, gated bool) *httptest.Server {
	var first sync.Once
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request body", 400)
			return
		}
		var request map[string]any
		json.Unmarshal(body, &request)
		model, _ := request["model"].(string)
		req, err := http.NewRequestWithContext(r.Context(), r.Method, target+r.URL.RequestURI(), bytes.NewReader(body))
		if err != nil {
			http.Error(w, "request construction", 500)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "peer request failed", 502)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		// Retain the original absent-Content-Type case instead of letting net/http
		// invent a content type when the stream's first bytes are written.
		if resp.Header.Get("Content-Type") == "" {
			w.Header()["Content-Type"] = nil
		}
		w.WriteHeader(resp.StatusCode)
		scan := bufio.NewScanner(resp.Body)
		var frame strings.Builder
		for scan.Scan() {
			line := scan.Text()
			// Captured real terminal responses identify their model. The original
			// fake omitted it, which made valid provenance-aware replay impossible
			// for clients that trust response metadata. Fill only that absent fact.
			if strings.HasPrefix(line, "data: ") {
				var event map[string]any
				if json.Unmarshal([]byte(line[6:]), &event) == nil && event["type"] == "response.completed" {
					response, _ := event["response"].(map[string]any)
					if _, present := response["model"]; response != nil && !present {
						response["model"] = model
						data, _ := json.Marshal(event)
						line = "data: " + string(data)
					}
				}
			}
			frame.WriteString(line + "\n")
			if line != "" {
				continue
			}
			data := frame.String()
			frame.Reset()
			hold := false
			if gated && strings.Contains(data, "event: response.reasoning_summary_text.delta\n") {
				first.Do(func() { hold = true })
			}
			if hold {
				g.mu.Lock()
				g.active = true
				g.mu.Unlock()
			}
			io.WriteString(w, data)
			w.(http.Flusher).Flush()
			if hold {
				select {
				case <-g.release:
				case <-r.Context().Done():
				case <-time.After(3 * time.Second):
				}
				g.mu.Lock()
				g.active = false
				g.mu.Unlock()
			}
		}
		// Ordinary JSON error bodies have no blank SSE delimiter.
		io.WriteString(w, frame.String())
	}))
}

func subscriptionLogin(bin, dir string, env []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "auth", "login", "--print-url")
	cmd.Dir = dir
	cmd.Env = env
	var output subscriptionBuffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return "", err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	visited := false
	for {
		select {
		case err := <-done:
			return output.String(), err
		case <-time.After(10 * time.Millisecond):
			if visited {
				continue
			}
			for _, field := range strings.Fields(output.String()) {
				raw := strings.Trim(field, "\"'<>()[],")
				if !strings.HasPrefix(raw, "http") || !strings.Contains(raw, "/authorize") {
					continue
				}
				visited = true
				client := http.Client{Timeout: 5 * time.Second}
				resp, err := client.Get(raw)
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				// The callback may close its listener after receiving the code,
				// before the browser's response finishes. The bounded login process
				// and observed token exchange determine authentication success.
				break
			}
		}
	}
}

func subscriptionLaunch(bin, skills, gui, dir string, o subscriptionOpts) (out subscriptionOut) {
	auth := grade.NewEdition2OAuthPeer(o.oauth)
	defer auth.Close()
	server := grade.NewEdition2ResponsesPeer(o.responses)
	defer server.Close()
	gate := &subscriptionGate{release: make(chan struct{})}
	proxy := gate.proxy(server.URL(), o.gate)
	defer proxy.Close()
	endpoint := proxy.URL
	defer func() {
		out.oauth = auth.Observations()
		out.responses = server.Observations()
		gate.mu.Lock()
		out.gated = gate.opened
		gate.mu.Unlock()
	}()
	home := filepath.Join(dir, "profile")
	if err := os.MkdirAll(home, 0700); err != nil {
		out.fatal = err.Error()
		return
	}
	env := subscriptionEnv(os.Environ(), home, dir, skills, auth.URL(), endpoint)
	if o.skipLogin {
		env = append(env, "OPENAI_BILLING_MODE=api")
	} else {
		env = append(env, "OPENAI_BILLING_MODE=chatgpt")
	}
	if o.executableOnly {
		env = subscriptionExecutableEnv(endpoint, home, skills, dir, o.skipLogin)
		if err := subscriptionSeed(home, o.signedOut); err != nil {
			out.fatal = err.Error()
			return
		}
	}
	if !o.skipLogin && !o.executableOnly {
		var err error
		out.loginOutput, err = subscriptionLogin(bin, dir, env)
		if err != nil {
			out.loginErr = err.Error()
			if subscriptionAbnormalExit(err, out.loginOutput) {
				out.fatal = "login did not complete with an ordinary refusal: " + err.Error()
			}
			return
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		out.fatal = err.Error()
		return
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	args := []string{"--port", port, "--gui-dir", gui, "--save", filepath.Join(dir, "save.json")}
	if o.executableOnly {
		args = []string{"--port", port, "--save", filepath.Join(dir, "save.json"), "--settings", filepath.Join(dir, "settings.json"), "--credentials-dir", filepath.Join(home, "credentials")}
	}
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = env
	var output subscriptionBuffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	stdin, err := cmd.StdinPipe()
	if err != nil {
		out.fatal = err.Error()
		return
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		out.fatal = err.Error()
		return
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	finished := false
	defer func() {
		stdin.Close()
		if !finished {
			cmd.Process.Kill()
			<-done
		}
		out.output = output.String()
		// The private credential directory is deliberately excluded. A protected
		// token store is permitted; a save, event log or request capture is not.
		filepath.WalkDir(dir, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return nil
			}
			if path == home {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			data, e := os.ReadFile(path)
			if e == nil {
				out.artifacts += string(data)
			}
			return nil
		})
		if out.fatal != "" {
			out.fatal += ": " + out.output
		}
	}()
	var peer *guiPeer
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case e := <-done:
			finished = true
			out.fatal = fmt.Sprintf("GUI exited before ready: %v", e)
			return
		default:
		}
		peer, err = openGUIPeer("ws://127.0.0.1:" + port + "/ws")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		out.fatal = "GUI did not start: " + err.Error()
		return
	}
	defer peer.close()
	if o.executableOnly {
		err = peer.conn.WriteJSON(map[string]any{"type": "subscribe", "agent_id": "root"})
	} else {
		err = peer.send("subscribe", "")
	}
	if err != nil {
		out.fatal = err.Error()
		return
	}
	if _, err = readSettings(peer, "current_settings"); err != nil {
		out.fatal = err.Error()
		return
	}
	for i, prompt := range o.prompts {
		if o.executableOnly {
			err = peer.conn.WriteJSON(map[string]any{"type": "prompt", "agent_id": "root", "text": prompt})
		} else {
			_, err = fmt.Fprintln(stdin, prompt)
		}
		if err != nil {
			out.fatal = err.Error()
			return
		}
		messages, e := peer.until(8*time.Second, func(m map[string]any) bool {
			if o.gate {
				gate.observe(m)
			}
			return m["type"] == "turn_ended" || (o.executableOnly && m["type"] == "lifecycle" && m["kind"] == "turn_ended")
		})
		out.messages = append(out.messages, messages...)
		if e != nil {
			out.fatal = fmt.Sprintf("turn %d: %v", i+1, e)
			return
		}
		if o.executableOnly {
			out.errors = append(out.errors, subscriptionExecutableError(messages))
		} else {
			out.errors = append(out.errors, subscriptionTurnError(messages))
		}
	}
	// Keep the subscriber connected until the child exits: unrelated legacy
	// disconnect races must not determine these credential/transport observations.
	stdin.Close()
	if o.executableOnly {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}
	select {
	case err = <-done:
		finished = true
		if err != nil {
			// A reported authentication/billing refusal may end a client with a
			// normal nonzero status. Positive sessions still require zero; signals
			// and unexplained startup/lifetime failures never establish refusal.
			if !subscriptionAbnormalExit(err, output.String()) {
				out.exitErr = err.Error()
			} else {
				out.fatal = "GUI exit after EOF: " + err.Error()
			}
		}
	case <-time.After(5 * time.Second):
		out.fatal = "GUI did not exit after EOF"
	}
	return
}

// Exit status alone cannot certify an expected refusal. Signals, deadlines and
// Go runtime failures remain failures even if an earlier error was displayed.
// Ordinary refusal still needs the scenario's actual issuer/request/error proof.
func subscriptionAbnormalExit(err error, output string) bool {
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() < 0 {
		return true
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "panic:") || strings.HasPrefix(line, "fatal error:") {
			return true
		}
	}
	return false
}

// Login prints while its child is still running; readers need a stable copy.
type subscriptionBuffer struct {
	mu   sync.Mutex
	text strings.Builder
}

func (b *subscriptionBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.Write(p)
}
func (b *subscriptionBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.text.String()
}

// Use only the fixture's selected issuer, transport and process-local state.
// Host provider/config overrides and credentials must not change an offline run.
func subscriptionEnv(base []string, home, dir, skills, issuer, endpoint string) []string {
	out := make([]string, 0, len(base)+24)
	for _, value := range base {
		key, _, _ := strings.Cut(value, "=")
		if key == "HOME" || key == "CH02_LOG" || strings.HasPrefix(key, "LLM_") || strings.HasPrefix(key, "EN_") || strings.HasPrefix(key, "ENSEMBLE_") || strings.HasPrefix(key, "ANTHROPIC_") || strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "GEMINI_") || strings.HasPrefix(key, "GOOGLE_") {
			continue
		}
		out = append(out, value)
	}
	return append(out, "HOME="+home, "EN_HOME="+home,
		"OPENAI_OIDC_ISSUER="+issuer, "LLM_BASE_URL="+endpoint, "LLM_VENDOR=openai",
		"LLM_MODEL=gpt-ch19-course", "LLM_RESOLVED_MODEL=gpt-ch19-course", "LLM_API_KEY="+grade.Edition2APIKey,
		"LLM_STREAM_CAPABILITY=sse", "OPENAI_REASONING_EFFORT=medium", "OPENAI_REASONING_SUMMARY=detailed",
		"EN_BROWSER_CMD=true", "EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL=base", "EN_DISABLE_STREAMING=0",
		"LLM_DATA_DIR="+filepath.Join(dir, "cr"), "LLM_SAVE_PATH="+filepath.Join(dir, "save.json"),
		"LLM_LOG_PATH="+filepath.Join(dir, "events.jsonl"), "LLM_LOG_MODE=full", "CH02_LOG="+filepath.Join(dir, "events.jsonl"),
		"LLM_TOOLS=default", "LLM_PAUSE_ON_START=0", "LLM_SPEAKING=0")
}

// ErrorOccurred is published before TurnEnded by the accepted GUI. Retain
// that actual error card, rather than requiring an extra terminal error field.
func subscriptionTurnError(messages []map[string]any) string {
	for _, m := range messages {
		if m["type"] == "error" {
			if text, ok := m["data"].(string); ok && text != "" {
				return text
			}
		}
		if m["type"] == "turn_ended" {
			if text, ok := m["error"].(string); ok {
				return text
			}
		}
	}
	return ""
}
