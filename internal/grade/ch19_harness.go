package grade

// Chapter 19 harness: drive the student's agent through a full Sign in with
// ChatGPT flow and then through inference on the Responses API, watching both
// ends of the wire.
//
// The shape of this file follows ch18's harness, but ch19 needs two fake
// servers rather than one, and they are not independent: the access token the
// agent presents to the inference server is the one the auth server minted a
// moment earlier. Nothing here inspects the student's source. Every
// observation is taken from an HTTP request the student's binary actually
// sent, or from a frame the GUI socket actually carried.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ch19Opts configures one end-to-end run.
type ch19Opts struct {
	dir     string   // student working directory
	model   string   // LLM_MODEL
	prompts []string // one turn per prompt

	// oauth shapes the fake authorization server.
	oauth ch19OAuthOptions
	// resp shapes the fake Responses server.
	resp ch19RespOptions

	// skipLogin runs the agent without performing the OAuth flow first, so a
	// check can confirm the agent refuses to guess at a credential.
	skipLogin bool

	// saveDir is filled in by ch19Launch: a private directory per run, so no
	// scenario inherits the conversation another one left behind.
	saveDir string
}

// ch19Out is everything the checks are allowed to look at.
type ch19Out struct {
	// oauth and resp are the two servers' observations.
	oauth ch19OAuthObs
	resp  ch19RespObs

	// loginStdout is what `auth login --print-url` printed. Checks read it to
	// confirm the authorize URL was surfaced rather than assumed.
	loginStdout string
	// agentOutput is the agent's combined stdout+stderr for the inference run.
	// The credential-hygiene check greps it for token material.
	agentOutput string

	// summaryDeltas counts part_delta frames whose kind is reasoning_summary.
	// This is the accessibility payoff, observed on the GUI socket rather than
	// inferred from the vendor stream.
	summaryDeltas int
	// textDeltas counts part_delta frames whose kind is text.
	textDeltas int
	// summaryBeforeText is true when at least one reasoning-summary delta
	// arrived before the first text delta. A summary that only appears after
	// the answer is finished is not a summary anyone can follow.
	summaryBeforeText bool

	// turnErrors is the error field of each turn_ended frame, in order. This
	// is the agent's own verdict on the turn. Grading a failure by grepping
	// the agent's output instead would pass on the vendor's bytes echoed to a
	// log, which an agent that ignored the failure entirely still prints.
	turnErrors []string

	turns int
	fatal string
}

// ch19Launch performs the login, then the inference run, and returns
// everything observed. The two phases share one pair of fake servers so the
// token minted in phase one is the token checked in phase two.
func ch19Launch(bin, skillsDir, guiDir string, o ch19Opts) ch19Out {
	auth := ch19NewFakeOAuth(o.oauth)
	defer auth.Close()
	resp := ch19NewFakeResponses(o.resp)
	defer resp.Close()

	home, err := os.MkdirTemp("", "ch19home")
	if err != nil {
		return ch19Out{fatal: fmt.Sprintf("temp home: %v", err)}
	}
	defer os.RemoveAll(home)
	o.saveDir = home

	// The environment both phases share. OPENAI_OIDC_ISSUER points discovery
	// at the fake authorization server; without it the student's agent would
	// talk to the real auth.openai.com and the grader would be scoring the
	// internet.
	baseEnv := append(os.Environ(),
		"OPENAI_OIDC_ISSUER="+auth.URL(),
		"LLM_BASE_URL="+resp.URL(),
		"LLM_MODEL="+o.model,
		"LLM_VENDOR=openai",
		// An API key is always present. With a stored grant the OAuth token
		// must win; with no grant the key must still work. One interface,
		// two credential kinds, and the renderer cannot tell them apart.
		"LLM_API_KEY="+ch19APIKey,
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"HOME="+home,
		"EN_HOME="+home,
		// EN_BROWSER_CMD exists so the agent can run where no browser can be
		// opened: over SSH, in a container, on a headless box. The grader
		// points it at `true`, which exits 0 without doing anything, and then
		// drives the authorize URL itself.
		"EN_BROWSER_CMD=true",
		"CH02_LOG="+filepath.Join(o.dir, "events.jsonl"),
	)

	out := ch19Out{}

	if !o.skipLogin {
		stdout, err := ch19Login(bin, o.dir, baseEnv, auth)
		out.loginStdout = stdout
		if err != nil {
			out.oauth = auth.Observations()
			out.fatal = fmt.Sprintf("auth login: %v", err)
			return out
		}
	}

	runOut := ch19RunAgent(bin, guiDir, o, baseEnv, auth)
	runOut.loginStdout = out.loginStdout
	runOut.oauth = auth.Observations()
	runOut.resp = resp.Observations()
	return runOut
}

// ch19Login runs the student's login command and plays the part of the
// browser. The agent prints an authorize URL; the harness fetches it, which
// makes the fake authorization server redirect to the agent's loopback
// callback, which completes the exchange.
func ch19Login(bin, dir string, env []string, auth *ch19FakeOAuth) (string, error) {
	cmd := exec.Command(bin, "auth", "login", "--print-url")
	cmd.Dir = dir
	cmd.Env = env
	var buf ch19SyncBuf
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// Wait for the authorize URL to appear on stdout, then act as the browser.
	deadline := time.After(20 * time.Second)
	var visited bool
	for !visited {
		select {
		case err := <-done:
			return buf.String(), fmt.Errorf("exited before printing an authorize URL (%v): %s", err, buf.String())
		case <-deadline:
			cmd.Process.Kill()
			<-done
			return buf.String(), fmt.Errorf("no authorize URL within 20s: %s", buf.String())
		case <-time.After(100 * time.Millisecond):
			if u := ch19FindAuthorizeURL(buf.String()); u != "" {
				if err := ch19VisitAuthorize(u); err != nil {
					cmd.Process.Kill()
					<-done
					return buf.String(), fmt.Errorf("browser step: %w", err)
				}
				visited = true
			}
		}
	}

	select {
	case err := <-done:
		if err != nil {
			return buf.String(), fmt.Errorf("login exited %v: %s", err, buf.String())
		}
	case <-time.After(20 * time.Second):
		cmd.Process.Kill()
		<-done
		return buf.String(), fmt.Errorf("login did not finish within 20s of the callback: %s", buf.String())
	}
	return buf.String(), nil
}

// ch19VisitAuthorize plays the part of the browser. A real sign-in ends with
// the authorization server redirecting to the agent's loopback callback; Go's
// default client follows redirects, so a plain GET delivers the code to
// whatever port the agent is listening on. The harness never learns that port
// and never needs to: the redirect carries it.
func ch19VisitAuthorize(rawURL string) error {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("authorize returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// ch19Tail returns the last n characters, so a failure message carries the
// part of the agent's output nearest the thing that went wrong.
func ch19Tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

// ch19FindAuthorizeURL pulls the first authorize URL out of whatever the agent
// printed. It deliberately does not require a particular sentence around the
// URL: the chapter specifies that the URL is printed, not how it is announced.
func ch19FindAuthorizeURL(s string) string {
	for _, f := range strings.Fields(s) {
		f = strings.Trim(f, "\"'<>()[],")
		if strings.Contains(f, "/authorize") && strings.HasPrefix(f, "http") {
			return f
		}
	}
	return ""
}

// ch19RunAgent starts the agent's GUI server and drives the prompts.
func ch19RunAgent(bin, guiDir string, o ch19Opts, env []string, auth *ch19FakeOAuth) ch19Out {
	port := freePort()
	// Each scenario gets a private save file. The agent's default is
	// save.json in its working directory, which every scenario shares, so
	// without this one scenario resumes the conversation the last one left
	// behind and the grader's result depends on what ran before it.
	savePath := filepath.Join(o.saveDir, "save.json")
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir, "--save", savePath)
	cmd.Dir = o.dir
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch19Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	var buf ch19SyncBuf
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Start(); err != nil {
		return ch19Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stop := func() {
		cmd.Process.Kill()
		<-done
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stop()
		return ch19Out{fatal: "GUI server did not start: " + buf.String()}
	}

	meter := newCh19Meter()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		stop()
		return ch19Out{fatal: fmt.Sprintf("dial /ws: %v", err)}
	}
	go meter.read(conn)
	if err := conn.WriteJSON(map[string]any{"type": "subscribe"}); err != nil {
		conn.Close()
		stop()
		return ch19Out{fatal: fmt.Sprintf("subscribe: %v", err)}
	}

	for i, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		if !meter.waitTurns(i+1, 60*time.Second) {
			conn.Close()
			stop()
			out := meter.snapshot()
			out.agentOutput = buf.String()
			out.fatal = fmt.Sprintf("after prompt %d the GUI saw %d turn_ended frames, expected %d; agent said: %s",
				i+1, meter.turns(), i+1, ch19Tail(buf.String(), 600))
			return out
		}
	}
	time.Sleep(500 * time.Millisecond)
	conn.Close()

	stdin.Close()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		<-done
		out := meter.snapshot()
		out.agentOutput = buf.String()
		out.fatal = "agent did not exit within 15s of stdin closing"
		return out
	}

	out := meter.snapshot()
	out.agentOutput = buf.String()
	return out
}

// ---------------------------------------------------------------------------
// The meter, read from the GUI socket
// ---------------------------------------------------------------------------

type ch19Meter struct {
	mu                sync.Mutex
	nTurns            int
	summaryDeltas     int
	textDeltas        int
	summaryBeforeText bool
	turnErrors        []string
}

func newCh19Meter() *ch19Meter { return &ch19Meter{} }

func (m *ch19Meter) read(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var f struct {
			Type  string `json:"type"`
			Kind  string `json:"kind"`
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &f) != nil {
			continue
		}
		m.mu.Lock()
		switch {
		case f.Type == "turn_ended":
			m.nTurns++
			m.turnErrors = append(m.turnErrors, f.Error)
		case f.Type == "part_delta" && f.Kind == "reasoning_summary":
			if m.textDeltas == 0 {
				m.summaryBeforeText = true
			}
			m.summaryDeltas++
		case f.Type == "part_delta" && f.Kind == "text":
			m.textDeltas++
		}
		m.mu.Unlock()
	}
}

func (m *ch19Meter) turns() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nTurns
}

func (m *ch19Meter) waitTurns(n int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if m.turns() >= n {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func (m *ch19Meter) snapshot() ch19Out {
	m.mu.Lock()
	defer m.mu.Unlock()
	return ch19Out{
		summaryDeltas:     m.summaryDeltas,
		textDeltas:        m.textDeltas,
		summaryBeforeText: m.summaryBeforeText,
		turnErrors:        append([]string{}, m.turnErrors...),
		turns:             m.nTurns,
	}
}

// ---------------------------------------------------------------------------
// A mutex-guarded buffer, because cmd.Stdout is written from another goroutine
// while the harness polls it for the authorize URL.
// ---------------------------------------------------------------------------

type ch19SyncBuf struct {
	mu sync.Mutex
	b  strings.Builder
}

func (w *ch19SyncBuf) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *ch19SyncBuf) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}
