package grade

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Chapter 18 launch machinery.
//
// This chapter grades a defect that produces correct answers. Every check
// here therefore measures something other than the agent's behaviour: the
// bytes it puts on the wire, the counts it writes to its event log, and the
// numbers it reports to its own GUI. An agent can pass every other chapter's
// grader while wasting ninety percent of its budget, so nothing in this file
// asks the agent whether its cache is working.
//
// Two rules shaped the design.
//
// First, the grader measures the prefix itself rather than reading the lens's
// verdict. The lens is one of the artifacts under test. If the grader trusted
// its log line, a lens that printed "identical" unconditionally would score
// full marks for prefix stability, which is precisely the failure mode the
// chapter warns about: an instrument that agrees with you is worse than none.
//
// Second, turns are counted from turn_ended frames on the WebSocket, not from
// request counts at the vendor. Chapter 17 added recall, so a turn can issue
// a judge request before its real one, and chapter 16 added a compressor that
// issues another. Waiting for "two requests" can be satisfied by two judge
// calls while no turn has happened at all.

// ch18Opts describes one launch.
type ch18Opts struct {
	dir     string
	model   string
	vendor  string
	prompts []string
	replies []fakevendor.Reply
	route   func(body []byte) *fakevendor.Reply
}

// ch18UsageFrame is the meter exactly as the GUI receives it.
type ch18UsageFrame struct {
	Input      int     `json:"input"`
	CacheWrite int     `json:"cache_write"`
	CacheRead  int     `json:"cache_read"`
	Output     int     `json:"output"`
	HitRate    float64 `json:"hit_rate"`
	CostUSD    float64 `json:"cost_usd"`
	Priced     bool    `json:"priced"`
}

// ch18Event is a line of the agent's own event log. The usage numbers are read
// from here rather than from the meter when the question is what the agent
// RECORDED, because the log is what a later run replays.
type ch18Event struct {
	Type     string `json:"type"`
	Response *struct {
		Usage *struct {
			Input      int `json:"input"`
			CacheWrite int `json:"cache_write"`
			CacheRead  int `json:"cache_read"`
			Output     int `json:"output"`
		} `json:"usage"`
	} `json:"response,omitempty"`
}

type ch18Out struct {
	reqs []fakevendor.Recorded
	// usage holds every meter frame seen, in order. The last one is the
	// session total at the end of the run.
	usage  []ch18UsageFrame
	events []ch18Event
	// rawUsage keeps the undecoded meter frames so a check can ask whether a
	// key exists at all, which a typed struct cannot answer.
	rawUsage []map[string]any
	// rawRespUsage holds the usage object of each recorded response, undecoded.
	// A typed struct silently discards an unexpected field, and an unexpected
	// field is exactly what "cost was stored in the log" looks like.
	rawRespUsage []map[string]any
	// lensFiles maps a filename the lens rotates to its contents.
	lensFiles map[string][]byte
	fatal     string
}

// ch18Launch runs the student's agent against a fake vendor and collects
// everything the checks need: the wire, the event log, the meter, and whatever
// the cache lens left on disk.
func ch18Launch(bin, skillsDir, guiDir string, o ch18Opts) ch18Out {
	srv := fakevendor.NewWithOptions(o.replies, fakevendor.Options{Route: o.route})
	defer srv.Close()

	vendor := o.vendor
	if vendor == "" {
		vendor = "anthropic"
	}

	port := freePort()
	cmd := exec.Command(bin, "--port", port, "--gui-dir", guiDir)
	cmd.Dir = o.dir
	cmd.Env = append(os.Environ(),
		"LLM_BASE_URL="+srv.URL(),
		"LLM_MODEL="+o.model,
		"LLM_VENDOR="+vendor,
		"LLM_API_KEY=test-key",
		"EN_SKILLS_DIR="+skillsDir,
		"EN_PRIMARY_SKILL=base",
		"CH02_LOG="+filepath.Join(o.dir, "events.jsonl"),
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ch18Out{fatal: fmt.Sprintf("stdin pipe: %v", err)}
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return ch18Out{fatal: fmt.Sprintf("start: %v", err)}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stop := func() {
		cmd.Process.Kill()
		<-done
	}

	if !waitForHTTP("http://localhost:"+port+"/", 10*time.Second) {
		stop()
		return ch18Out{fatal: "GUI server did not start"}
	}

	meter := newCh18Meter()
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:"+port+"/ws", nil)
	if err != nil {
		stop()
		return ch18Out{fatal: fmt.Sprintf("dial /ws: %v", err)}
	}
	go meter.read(conn)
	// The server streams nothing until a client asks. This is the client-pull
	// protocol a reconnecting GUI uses to rebuild itself, and a grader that
	// skips it sits on a silent socket watching turns it cannot see.
	if err := conn.WriteJSON(map[string]any{"type": "subscribe"}); err != nil {
		conn.Close()
		stop()
		return ch18Out{fatal: fmt.Sprintf("subscribe: %v", err)}
	}

	for i, p := range o.prompts {
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": p})
		fmt.Fprintln(stdin, string(line))
		if !meter.waitTurns(i+1, 40*time.Second) {
			conn.Close()
			stop()
			return ch18Out{
				reqs: srv.Requests(), events: ch18ReadLog(o.dir),
				usage: meter.frames(), rawUsage: meter.raw(),
				fatal: fmt.Sprintf("after prompt %d the GUI saw %d turn_ended frames, expected %d (%d vendor requests)",
					i+1, meter.turns(), i+1, len(srv.Requests())),
			}
		}
	}
	// The meter frame for the final turn can trail its turn_ended.
	time.Sleep(500 * time.Millisecond)
	conn.Close()

	stdin.Close()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		<-done
		return ch18Out{
			reqs: srv.Requests(), events: ch18ReadLog(o.dir),
			usage: meter.frames(), rawUsage: meter.raw(),
			fatal: "agent did not exit within 15s of stdin closing",
		}
	}

	return ch18Out{
		reqs:         srv.Requests(),
		events:       ch18ReadLog(o.dir),
		usage:        meter.frames(),
		rawUsage:     meter.raw(),
		rawRespUsage: ch18ReadRawUsage(o.dir),
		lensFiles:    ch18ReadLensFiles(o.dir),
	}
}

// ---------------------------------------------------------------------------
// The meter, read from the GUI socket
// ---------------------------------------------------------------------------

type ch18Meter struct {
	mu        sync.Mutex
	turnEnded int
	decoded   []ch18UsageFrame
	rawFrames []map[string]any
}

func newCh18Meter() *ch18Meter { return &ch18Meter{} }

func (m *ch18Meter) read(conn *websocket.Conn) {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var head struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(data, &head) != nil {
			continue
		}
		m.mu.Lock()
		switch head.Type {
		case "turn_ended":
			m.turnEnded++
		case "usage":
			var f ch18UsageFrame
			if json.Unmarshal(data, &f) == nil {
				m.decoded = append(m.decoded, f)
			}
			var raw map[string]any
			if json.Unmarshal(data, &raw) == nil {
				m.rawFrames = append(m.rawFrames, raw)
			}
		}
		m.mu.Unlock()
	}
}

func (m *ch18Meter) turns() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.turnEnded
}

func (m *ch18Meter) frames() []ch18UsageFrame {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ch18UsageFrame{}, m.decoded...)
}

func (m *ch18Meter) raw() []map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]map[string]any{}, m.rawFrames...)
}

func (m *ch18Meter) waitTurns(n int, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if m.turns() >= n {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return m.turns() >= n
}

// ---------------------------------------------------------------------------
// Reading what the run left behind
// ---------------------------------------------------------------------------

func ch18ReadLog(dir string) []ch18Event {
	data, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		return nil
	}
	var out []ch18Event
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e ch18Event
		if json.Unmarshal([]byte(line), &e) == nil {
			out = append(out, e)
		}
	}
	return out
}

// ch18ReadRawUsage returns each response's usage object exactly as written,
// with no schema imposed. The check that matters here asks whether a field is
// present that should not be, and decoding into a struct would throw that
// evidence away before the question could be asked.
func ch18ReadRawUsage(dir string) []map[string]any {
	data, err := os.ReadFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		return nil
	}
	var out []map[string]any
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e struct {
			Type     string `json:"type"`
			Response *struct {
				Usage map[string]any `json:"usage"`
			} `json:"response,omitempty"`
		}
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		if e.Type == "response_ended" && e.Response != nil && e.Response.Usage != nil {
			out = append(out, e.Response.Usage)
		}
	}
	return out
}

// ch18LensNames are the filenames the lens rotates. The grader looks for the
// pair rather than for a particular implementation, because the deliverable is
// "a plain diff lands on the first broken byte", not a file layout.
var ch18LensNames = []string{"request.json", "prior_request.json"}

func ch18ReadLensFiles(dir string) map[string][]byte {
	out := map[string][]byte{}
	for _, name := range ch18LensNames {
		if b, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			out[name] = b
		}
	}
	return out
}

// ch18Responses returns the usage block of every response event, in order.
func ch18Responses(events []ch18Event) []ch18UsageFrame {
	var out []ch18UsageFrame
	for _, e := range events {
		if e.Type != "response_ended" || e.Response == nil || e.Response.Usage == nil {
			continue
		}
		u := e.Response.Usage
		out = append(out, ch18UsageFrame{
			Input: u.Input, CacheWrite: u.CacheWrite, CacheRead: u.CacheRead, Output: u.Output,
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Reading the wire
// ---------------------------------------------------------------------------

// ch18Sections splits a request body into the sections that lay out in cache
// order. Comparing whole bodies would be useless: the tail changes every turn
// by design, so a whole-body diff always differs and says nothing about which
// part broke.
//
// The grader does its own splitting rather than calling the student's
// canonicaliser, because that canonicaliser is under test.
func ch18Sections(body []byte) map[string]string {
	var top map[string]json.RawMessage
	if json.Unmarshal(body, &top) != nil {
		return nil
	}
	out := map[string]string{}
	for _, k := range []string{"tools", "system", "messages", "systemInstruction", "contents", "input"} {
		if v, ok := top[k]; ok {
			out[k] = ch18StripMarkers(string(v))
		}
	}
	return out
}

// ch18StripMarkers removes cache breakpoints before comparing sections.
//
// A breakpoint is an instruction to the provider, not conversation content,
// and the rolling one MOVES: a block carrying a marker in one request carries
// none in the next. Comparing raw bytes would report the tool declarations as
// unstable the moment a student marks them, failing a correct implementation.
func ch18StripMarkers(s string) string {
	for _, m := range []string{
		`,"cache_control":{"type":"ephemeral"}`,
		`"cache_control":{"type":"ephemeral"},`,
		`,"cache_control": {"type": "ephemeral"}`,
	} {
		s = strings.ReplaceAll(s, m, "")
	}
	return s
}

// ch18HasSystemBreakpoint reports whether the system section carries a marker.
// Read from the wire: the failure this guards against is a marker that exists
// in the struct and never reaches the JSON.
func ch18HasSystemBreakpoint(body []byte) bool {
	var top struct {
		System json.RawMessage `json:"system"`
	}
	if json.Unmarshal(body, &top) != nil || len(top.System) == 0 {
		return false
	}
	var blocks []map[string]any
	if json.Unmarshal(top.System, &blocks) != nil {
		// A plain string cannot carry a breakpoint, which is the whole reason
		// the field had to become an array of content blocks.
		return false
	}
	for _, b := range blocks {
		if cc, ok := b["cache_control"].(map[string]any); ok {
			if cc["type"] == "ephemeral" {
				return true
			}
		}
	}
	return false
}

// ch18CountMarkers counts breakpoints anywhere in a request body.
func ch18CountMarkers(body []byte) int {
	s := string(body)
	return strings.Count(s, `"cache_control"`)
}
