// ch13-speech launches the real GUI, or explicitly attaches to an existing page. CDP is private driver
// plumbing: no listener tool accepts JavaScript, selectors, URLs or model config.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"ensemble"
	"github.com/gorilla/websocket"
)

const contract = `{
 "read_file_tool":"read_file","read_file_arg":"path","supports_type_during_turn":true,"supports_streaming_off":true,"log_format":"jsonl",
 "invocation":"tts-harness.sh PROMPT [--type-during-turn TEXT]",
 "environment":{"LLM_BASE_URL":"Required local external provider URL","WORKSPACE":"Required prepared directory; fixture paths are absolute, GUI runs in its owned runtime directory","TTS_LOG":"Required destination FILE for observed JSONL","STREAMING":"off requests plain delivery; otherwise stream","LLM_VENDOR":"Optional current provider vendor (default openai)","LLM_MODEL":"Optional current model (default chapter13-local)","CHROME_BIN":"Optional installed Chromium executable; never downloaded"},
 "attach":{"invocation":"tts-harness.sh --attach [--jsonl|--listener] PROMPT","environment":{"ENSEMBLE_GUI_URL":"Existing GUI page URL","ENSEMBLE_CDP_URL":"Existing page-target CDP URL","TTS_LOG":"Transcript FILE","ENSEMBLE_SPEECH_MODE":"native or transcript","ENSEMBLE_LISTENER_MODEL":"Separate listener model","ENSEMBLE_LISTENER_VENDOR":"Separate listener vendor","ENSEMBLE_LISTENER_BASE_URL":"Separate listener endpoint","ENSEMBLE_LISTENER_API_KEY":"Optional listener credential, never logged","ENSEMBLE_LISTENER_GOAL":"Conversational goal"}},
 "jsonl":{"input":"One {id,op,text?,ms?,timeout_ms?} per line after initial prompt submission","operations":{"hear":"Incremental admitted page utterances and status, not pending text","type":"Trusted composer input; whitespace counts","submit":"Submit only driver-owned input to idle selected target","wait":"Observe the owned target turn; distinguish success, error, interruption and timeout","sleep":"Wait 0..30000 ms"},"output":"{id,ok,result?,error?}; EOF settles or reports incomplete owned work"},
 "listener":{"tools":["hear","type","submit","wait","sleep"],"input":"Only goal and tool results; no DOM, GUI-debug, target history, Engine or filesystem capability"},
 "evidence":"Normal entry runs the real GUI with an asynchronous synthesizer fake. Speech-channel records and fake callbacks are not native/hearing proof. --describe starts no GUI/browser/model."
}`

type page struct {
	conn *websocket.Conn
	mu   sync.Mutex
	next int
}

func (p *page) call(ctx context.Context, method string, params any, out any) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("CDP %s: %w", method, err)
		}
	}()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	// Cancel a blocked page operation when its peer dies. The connection is
	// intentionally not reusable after cancellation; cleanup reports its loss.
	cancelled := make(chan struct{})
	stop := context.AfterFunc(ctx, func() { p.conn.Close(); close(cancelled) })
	defer func() {
		if !stop() {
			<-cancelled
		}
	}()
	deadline := time.Now().Add(15 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	p.conn.SetWriteDeadline(deadline)
	p.conn.SetReadDeadline(deadline)
	p.next++
	id := p.next
	if err := p.conn.WriteJSON(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return fmt.Errorf("page command write failed: %w", err)
	}
	for {
		var response struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  json.RawMessage `json:"error"`
		}
		if err := p.conn.ReadJSON(&response); err != nil {
			return fmt.Errorf("page command read failed: %w", err)
		}
		if response.ID != id {
			continue
		}
		if len(response.Error) > 0 {
			return errors.New("page command rejected")
		}
		if out != nil {
			return json.Unmarshal(response.Result, out)
		}
		return nil
	}
}
func (p *page) evaluate(ctx context.Context, expression string, out any) error {
	var result struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
		Exception json.RawMessage `json:"exceptionDetails"`
	}
	if err := p.call(ctx, "Runtime.evaluate", map[string]any{"expression": expression, "returnByValue": true, "awaitPromise": true}, &result); err != nil {
		return err
	}
	if len(result.Exception) > 0 {
		return errors.New("page operation failed")
	}
	if out != nil {
		return json.Unmarshal(result.Result.Value, out)
	}
	return nil
}
func literal(value any) string { raw, _ := json.Marshal(value); return string(raw) }

type utterance struct {
	Kind           string `json:"kind"`
	MS             int64  `json:"ms"`
	CausalSequence int    `json:"causal_sequence,omitempty"`
	Audio          bool   `json:"audio,omitempty"`
	Utterance      string `json:"utterance,omitempty"`
	Cause          string `json:"cause,omitempty"`
	Status         string `json:"status,omitempty"`
	AgentID        string `json:"agent_id"`
	TurnID         uint64 `json:"turn_id"`
	PartID         uint64 `json:"part_id"`
	Seq            int    `json:"seq"`
	Text           string `json:"text"`
	Source         string `json:"source"`
	Mode           string `json:"mode"`
	Event          string `json:"event"`
	Evidence       string `json:"evidence"`
}
type transcript struct {
	Cursor int         `json:"cursor"`
	Mode   string      `json:"mode"`
	Events []utterance `json:"events"`
}
type pageStatus struct {
	AgentID         string `json:"agent_id"`
	SelectedAgentID string `json:"selected_agent_id"`
	TurnID          uint64 `json:"turn_id"`
	Outcome         string `json:"outcome"`
	Prompt          string `json:"prompt"`
	Draft           string `json:"draft"`
	InputRevision   int    `json:"input_revision"`
	HumanTyping     bool   `json:"human_typing"`
	Ready           bool   `json:"ready"`
	Turn            string `json:"turn"`
	Paused          bool   `json:"paused"`
	Busy            bool   `json:"busy"`
	Revision        int    `json:"revision"`
	Ended           int    `json:"ended"`
}
type operation struct {
	ID        json.RawMessage `json:"id,omitempty"`
	Op        string          `json:"op"`
	Text      string          `json:"text,omitempty"`
	MS        int             `json:"ms,omitempty"`
	TimeoutMS int             `json:"timeout_ms,omitempty"`
}
type driver struct {
	actionsMu                     sync.Mutex
	page                          *page
	cursor                        int
	agentID                       string
	baseline, turnID              uint64
	submittedText, draft, outcome string
	draftRevision                 int
	ownsDraft                     bool
	submitted                     bool
	speech, actions               *json.Encoder
	mu                            sync.Mutex
	operationMu                   sync.Mutex
}

func (d *driver) hear(ctx context.Context) (transcript, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out transcript
	if err := d.page.evaluate(ctx, fmt.Sprintf("window.speechChannel.read(%d)", d.cursor), &out); err != nil {
		return out, err
	}
	filtered := make([]utterance, 0, len(out.Events))
	for _, entry := range out.Events {
		if err := d.speech.Encode(entry); err != nil {
			return out, err
		}
		if entry.Event == "utterance" || entry.Event == "complete" {
			filtered = append(filtered, entry)
		}
	}
	d.cursor = out.Cursor
	out.Events = filtered
	return out, nil
}
func (d *driver) status(ctx context.Context) (pageStatus, error) {
	var out pageStatus
	arg := ""
	if d.agentID != "" {
		arg = literal(d.agentID)
	}
	err := d.page.evaluate(ctx, "window.speechChannel.status("+arg+")", &out)
	return out, err
}
func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func (d *driver) act(ctx context.Context, op operation) (result any, err error) {
	d.operationMu.Lock()
	defer d.operationMu.Unlock()
	d.actionsMu.Lock()
	defer d.actionsMu.Unlock()
	defer func() {
		d.mu.Lock()
		defer d.mu.Unlock()
		err = errors.Join(err, d.actions.Encode(map[string]any{"at": time.Now().UTC(), "operation": op, "ok": err == nil}))
	}()
	switch op.Op {
	case "hear":
		return d.hear(ctx)
	case "type":
		if op.Text == "" {
			return nil, errors.New("type requires text (spaces are allowed)")
		}
		status, e := d.status(ctx)
		if e != nil {
			return nil, e
		}
		if d.agentID == "" {
			d.agentID = status.AgentID
		}
		if !status.Ready || status.SelectedAgentID != d.agentID {
			return nil, errors.New("target is not the selected ready Agent")
		}
		if (status.Draft != "" || status.HumanTyping) && (!d.ownsDraft || status.InputRevision != d.draftRevision || status.Draft != d.draft) {
			return nil, errors.New("composer belongs to human input")
		}
		if err = d.page.evaluate(ctx, "document.getElementById('input').focus()", nil); err != nil {
			return nil, err
		}
		if err = d.page.call(ctx, "Input.insertText", map[string]any{"text": op.Text}, nil); err != nil {
			return nil, err
		}
		after, e := d.status(ctx)
		if e != nil {
			return nil, e
		}
		if after.Draft != status.Draft+op.Text {
			return nil, errors.New("composer changed during driver input; ownership not acquired")
		}
		d.ownsDraft = true
		d.draft = after.Draft
		d.draftRevision = after.InputRevision
		return map[string]any{"typed": true}, nil
	case "submit":
		if d.submitted {
			return nil, errors.New("an owned target turn is still unsettled")
		}
		status, e := d.status(ctx)
		if e != nil {
			return nil, e
		}
		if status.SelectedAgentID != d.agentID || !d.ownsDraft || status.InputRevision != d.draftRevision || status.Draft != d.draft {
			return nil, errors.New("composer is not driver-owned")
		}
		if status.Turn != "idle" || status.Busy {
			return nil, errors.New("target must be idle before submitting")
		}
		var accepted bool
		err = d.page.evaluate(ctx, "(()=>{const input=document.getElementById('input');if(!input.value.trim())return false;document.getElementById('composer').requestSubmit();return input.value==='';})()", &accepted)
		if err != nil {
			return nil, err
		}
		if !accepted {
			return nil, errors.New("composer did not submit")
		}
		d.baseline = status.TurnID
		d.turnID = 0
		d.submittedText = d.draft
		d.submitted = true
		d.outcome = "pending"
		d.ownsDraft = false
		d.draft = ""
		return map[string]any{"submitted": true}, nil
	case "sleep":
		if op.MS < 0 || op.MS > 30000 {
			return nil, errors.New("ms must be 0..30000")
		}
		err = sleep(ctx, time.Duration(op.MS)*time.Millisecond)
		return map[string]any{"slept_ms": op.MS}, err
	case "wait":
		observed, e := d.waitTarget(ctx, op.TimeoutMS)
		if e != nil {
			return observed, e
		}
		if observed["status"] != "success" {
			return observed, fmt.Errorf("target turn %s", observed["status"])
		}
		return observed, nil
	default:
		return nil, errors.New("operation must be hear, type, submit, wait or sleep")
	}
}

// waitTarget keeps observation/capture success separate from target success.
// Interactive wait still requires success; only normal one-shot accepts error.
func (d *driver) waitTarget(ctx context.Context, timeoutMS int) (map[string]any, error) {
	ms := timeoutMS
	if ms == 0 {
		ms = 120000
	}
	if ms < 1 || ms > 300000 {
		return nil, errors.New("timeout_ms must be 1..300000")
	}
	if !d.submitted {
		if d.outcome != "" && d.outcome != "pending" {
			return map[string]any{"completed": d.outcome == "success", "status": d.outcome}, nil
		}
		return nil, errors.New("no observed owned target turn")
	}
	until := time.Now().Add(time.Duration(ms) * time.Millisecond)
	for time.Now().Before(until) {
		status, e := d.status(ctx)
		if e != nil {
			return nil, e
		}
		if status.AgentID != d.agentID {
			return nil, errors.New("target identity changed")
		}
		if status.TurnID > d.baseline {
			if d.turnID == 0 {
				if status.Prompt != d.submittedText {
					return nil, errors.New("observed turn does not match owned submission")
				}
				d.turnID = status.TurnID
			}
			if status.TurnID != d.turnID {
				return nil, errors.New("owned target turn was overtaken")
			}
			if status.Outcome != "pending" && status.Outcome != "" && !status.Busy {
				d.mu.Lock()
				captureErr := d.speech.Encode(map[string]any{"kind": "target_outcome", "agent_id": status.AgentID, "turn_id": status.TurnID, "status": status.Outcome})
				d.mu.Unlock()
				if captureErr != nil {
					return nil, captureErr
				}
				d.submitted = false
				d.outcome = status.Outcome
				return map[string]any{"completed": status.Outcome == "success", "status": status.Outcome}, nil
			}
		}
		if e = sleep(ctx, 20*time.Millisecond); e != nil {
			return nil, e
		}
	}
	return map[string]any{"completed": false, "status": "timeout"}, errors.New("target timeout; owned turn is incomplete")
}
func (d *driver) observeOneShot(ctx context.Context) error {
	if d.ownsDraft {
		return errors.New("incomplete owned input: typed but not submitted")
	}
	observed, err := d.waitTarget(ctx, 0)
	if err != nil {
		return err
	}
	if observed["status"] != "success" && observed["status"] != "error" {
		return fmt.Errorf("target turn %s", observed["status"])
	}
	return nil
}

// Finishing the listener's own Ask does not prove target completion.
func (d *driver) settle(ctx context.Context) error {
	if d.ownsDraft {
		return errors.New("incomplete owned input: typed but not submitted")
	}
	if d.submitted {
		_, err := d.act(ctx, operation{Op: "wait"})
		return err
	}
	if d.outcome != "success" {
		return errors.New("target did not complete successfully")
	}
	return nil
}
func (d *driver) releaseDraft(ctx context.Context) error {
	if !d.ownsDraft {
		return nil
	}
	var released bool
	err := d.page.evaluate(ctx, "window.speechChannel.releaseDraft("+literal(d.agentID)+","+fmt.Sprint(d.draftRevision)+","+literal(d.draft)+")", &released)
	d.ownsDraft = false
	if err != nil {
		return err
	}
	if !released {
		return errors.New("driver draft changed; human input left untouched")
	}
	return nil
}
func (d *driver) cleanup(ctx context.Context) error {
	err := d.releaseDraft(ctx)
	if !d.submitted {
		return err
	}
	status, e := d.status(ctx)
	if e != nil {
		return errors.Join(err, e)
	}
	if status.AgentID == d.agentID && status.TurnID > d.baseline && status.Prompt == d.submittedText && (d.turnID == 0 || d.turnID == status.TurnID) && status.Outcome == "pending" {
		var interrupted bool
		e = d.page.evaluate(ctx, "window.speechChannel.interruptOwned("+literal(d.agentID)+","+fmt.Sprint(status.TurnID)+")", &interrupted)
		if e == nil && !interrupted {
			e = errors.New("owned target changed before cleanup")
		}
	}
	return errors.Join(err, e)
}
func required(name string) (string, error) {
	value := os.Getenv(name)
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}
func endpoint(value string, schemes ...string) error {
	u, err := url.Parse(value)
	if err != nil || u.Host == "" || u.User != nil {
		return errors.New("invalid endpoint")
	}
	for _, scheme := range schemes {
		if u.Scheme == scheme {
			return nil
		}
	}
	return errors.New("invalid endpoint scheme")
}
func listener(ctx context.Context, d *driver, dir string) error {
	model, err := required("ENSEMBLE_LISTENER_MODEL")
	if err != nil {
		return err
	}
	base, err := required("ENSEMBLE_LISTENER_BASE_URL")
	if err != nil {
		return err
	}
	if err = endpoint(base, "http", "https"); err != nil {
		return err
	}
	goal, err := required("ENSEMBLE_LISTENER_GOAL")
	if err != nil {
		return err
	}
	var vendor ensemble.Vendor
	switch os.Getenv("ENSEMBLE_LISTENER_VENDOR") {
	case "openai":
		vendor = ensemble.OpenAI
	case "anthropic":
		vendor = ensemble.Anthropic
	case "gemini":
		vendor = ensemble.Gemini
	default:
		return errors.New("ENSEMBLE_LISTENER_VENDOR must be openai, anthropic or gemini")
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	agent, err := app.NewAgent(ensemble.Config{Vendor: vendor, Model: model, BaseURL: base, APIKey: os.Getenv("ENSEMBLE_LISTENER_API_KEY"), Tools: []string{}, BaseDir: dir, DataDir: filepath.Join(dir, "listener"), DisableStreaming: true, MaxToolRounds: 40, SystemPrompt: "You are a blind speech-channel listener conversing with another agent through a real GUI. Your only capabilities are hear, type, submit, wait, sleep. You cannot see its DOM, screenshots, state, tools, engine, or history. The initial target prompt has already been submitted. Use wait and hear to understand its response, then carry out your goal with conversational follow-ups. Transcript mode is speech-channel evidence, not hearing. Type and submit use the real user composer. Report what the spoken channel did and did not make understandable; do not invent observations."})
	if err != nil {
		return errors.New("listener setup failed")
	}
	for _, name := range []string{"hear", "type", "submit", "wait", "sleep"} {
		name := name
		properties := map[string]any{}
		switch name {
		case "type":
			properties["text"] = map[string]any{"type": "string"}
		case "wait":
			properties["timeout_ms"] = map[string]any{"type": "integer", "minimum": 1, "maximum": 300000}
		case "sleep":
			properties["ms"] = map[string]any{"type": "integer", "minimum": 0, "maximum": 30000}
		}
		schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
		if name == "type" {
			schema["required"] = []string{"text"}
		}
		if name == "sleep" {
			schema["required"] = []string{"ms"}
		}
		raw, _ := json.Marshal(schema)
		err = agent.RegisterExternalTool(ensemble.ToolDefinition{Name: name, Description: "Speech listener operation: " + name, Parameters: raw}, func(ctx context.Context, call ensemble.ToolCall, args json.RawMessage) error {
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(args, &fields); err != nil {
				return err
			}
			for key := range fields {
				if _, ok := properties[key]; !ok {
					return errors.New("unsupported listener argument")
				}
			}
			var op operation
			if err := json.Unmarshal(args, &op); err != nil {
				return err
			}
			op.Op = name
			result, err := d.act(ctx, op)
			if err != nil {
				return err
			}
			return json.NewEncoder(call).Encode(result)
		})
		if err != nil {
			return err
		}
	}
	// Fail closed if future defaults accidentally widen listener capabilities.
	definitions := agent.Tools().Definitions()
	if len(definitions) != 5 {
		return errors.New("listener capability set is not closed")
	}
	allowed := map[string]bool{"hear": true, "type": true, "submit": true, "wait": true, "sleep": true}
	for _, definition := range definitions {
		if !allowed[definition.Name] {
			return errors.New("unexpected listener capability")
		}
	}
	answer, err := agent.Ask(ctx, goal)
	if err != nil {
		return errors.New("listener turn failed; inspect local listener diagnostics")
	}
	return os.WriteFile(filepath.Join(dir, "listener.txt"), []byte(answer+"\n"), 0600)
}
func runSession(ctx context.Context, p *page, prompt, logPath, mode, typeDuring string, jsonl, listen bool, runDir string, peer *peerInput) (err error) {
	if peer != nil {
		defer func() {
			if cause := context.Cause(peer.ctx); cause != nil {
				err = errors.Join(err, cause)
			}
		}()
	}
	if mode != "native" && mode != "transcript" {
		return errors.New("speech mode must be native or transcript")
	}
	if err = os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return err
	}
	output, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	d := &driver{page: p, speech: json.NewEncoder(output), actions: json.NewEncoder(io.Discard)}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err = errors.Join(err, d.cleanup(cleanup))
		_, captureErr := d.hear(cleanup)
		err = errors.Join(err, captureErr, output.Close())
	}()
	readyBy := time.Now().Add(15 * time.Second)
	for {
		status, e := d.status(ctx)
		if e == nil && status.Ready {
			break
		}
		if time.Now().After(readyBy) {
			return errors.New("GUI did not become ready")
		}
		if e = sleep(ctx, 30*time.Millisecond); e != nil {
			return e
		}
	}
	if err = p.evaluate(ctx, "window.speechChannel.setMode("+literal(mode)+")", nil); err != nil {
		return err
	}
	if _, err = d.act(ctx, operation{Op: "type", Text: prompt}); err != nil {
		return err
	}
	if _, err = d.act(ctx, operation{Op: "submit"}); err != nil {
		return err
	}
	if typeDuring != "" {
		for {
			status, e := d.status(ctx)
			if e != nil {
				return e
			}
			if status.TurnID > d.baseline && status.Prompt == d.submittedText && (status.Busy || status.Turn != "idle") {
				break
			}
			if status.TurnID > d.baseline && status.Outcome != "pending" {
				return errors.New("target completed before typing opportunity")
			}
			if e = sleep(ctx, 10*time.Millisecond); e != nil {
				return e
			}
		}
		if _, err = d.act(ctx, operation{Op: "type", Text: typeDuring}); err != nil {
			return err
		}
		if err = sleep(ctx, 150*time.Millisecond); err != nil {
			return err
		}
		if err = d.releaseDraft(ctx); err != nil {
			return err
		}
	}
	if listen {
		if err = listener(ctx, d, runDir); err != nil {
			return err
		}
		return d.settle(ctx)
	}
	if jsonl {
		if peer == nil {
			return errors.New("JSONL requires an input peer")
		}
		var operationErrors error
		for {
			line, e := peer.next()
			if e != nil {
				return errors.Join(operationErrors, e)
			}
			var op operation
			decoder := json.NewDecoder(strings.NewReader(line))
			decoder.DisallowUnknownFields()
			if e = decoder.Decode(&op); e != nil {
				return e
			}
			result, e := d.act(ctx, op)
			reply := map[string]any{"id": op.ID, "ok": e == nil, "result": result}
			if e != nil {
				reply["error"] = e.Error()
				operationErrors = errors.Join(operationErrors, e)
			}
			if e = json.NewEncoder(os.Stdout).Encode(reply); e != nil {
				return e
			}
		}
	}
	return d.observeOneShot(ctx)
}

func run() (err error) {
	describe := flag.Bool("describe", false, "print the public contract without starting processes")
	sourceRoot := flag.String("source-root", "../..", "source root (set by tts-harness.sh)")
	attach := flag.Bool("attach", false, "attach to an explicitly supplied existing GUI page")
	jsonl := flag.Bool("jsonl", false, "read restricted JSONL driver operations after initial submission")
	listen := flag.Bool("listener", false, "run the restricted listener with separate model configuration")
	flag.Parse()
	if *describe {
		_, err = fmt.Fprintln(os.Stdout, contract)
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	limit := 80 * time.Second
	if *attach {
		limit = 10 * time.Minute
	}
	ctx, timeout := context.WithTimeout(ctx, limit)
	defer timeout()
	var peer *peerInput
	if *jsonl || *listen {
		peer, err = watchPeer(ctx, os.Stdin, *jsonl)
		if err != nil {
			return err
		}
		ctx = peer.ctx
		defer func() { err = errors.Join(err, peer.close()) }()
	}
	var p *page
	var stop func() error
	var prompt, logPath, typeDuring, mode string
	mode = "native"
	runDir, e := os.MkdirTemp("", "ch13-speech-")
	if e != nil {
		return e
	}
	if *attach {
		if flag.NArg() != 1 {
			return errors.New("attach requires a prompt argument")
		}
		prompt = flag.Arg(0)
		expected, e := required("ENSEMBLE_GUI_URL")
		if e != nil {
			return e
		}
		target, e := required("ENSEMBLE_CDP_URL")
		if e != nil {
			return e
		}
		logPath, e = required("TTS_LOG")
		if e != nil {
			return e
		}
		if e = endpoint(expected, "http", "https"); e != nil {
			return e
		}
		if e = endpoint(target, "ws", "wss"); e != nil {
			return e
		}
		if value := os.Getenv("ENSEMBLE_SPEECH_MODE"); value != "" {
			mode = value
		}
		conn, _, e := websocket.DefaultDialer.DialContext(ctx, target, nil)
		if e != nil {
			return e
		}
		p = &page{conn: conn}
		defer conn.Close()
		var actual string
		if e = p.evaluate(ctx, "location.href", &actual); e != nil {
			return e
		}
		want, e1 := url.Parse(expected)
		got, e2 := url.Parse(actual)
		if e1 != nil || e2 != nil || want.Scheme != got.Scheme || want.Host != got.Host || want.Path != got.Path {
			return errors.New("CDP target is not the expected GUI page")
		}
	} else {
		config, text, e := normalConfig(*sourceRoot, flag.Args())
		if e != nil {
			return e
		}
		prompt = text
		logPath = config.logPath
		typeDuring = config.typeDuring
		p, stop, e = launchPage(ctx, config, runDir)
		if e != nil {
			return e
		}
		defer func() { err = errors.Join(err, p.conn.Close(), stop()) }()
	}
	return runSession(ctx, p, prompt, logPath, mode, typeDuring, *jsonl, *listen, runDir, peer)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
