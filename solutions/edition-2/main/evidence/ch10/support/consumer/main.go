// Chapter 10 public evidence consumer. No private/runtime imports or credentials
// reader. Requests only target a caller-supplied literal loopback relay.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"example.com/ensemble"
	"example.com/ensemble/cli"
)

type owner struct {
	app                                    *ensemble.Ensemble
	root, output, relay, provider, catalog string
	base                                   ensemble.Config
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() (err error) {
	root := flag.String("root", "", "provider scratch root")
	output := flag.String("output", "", "new result directory")
	step := flag.String("step", "", "C, D-prepare, D-seed, D2, or offline")
	relay := flag.String("relay", "", "loopback enforcing relay")
	catalog := flag.String("catalog", "", "exact catalog directory")
	flag.Parse()
	if *root == "" || *output == "" {
		return fmt.Errorf("root and output required")
	}
	parsed, e := url.Parse(*relay)
	if e != nil || parsed.Scheme != "http" || parsed.Hostname() != "127.0.0.1" || parsed.User != nil || parsed.RawQuery != "" {
		return fmt.Errorf("literal loopback relay required")
	}
	a := ensemble.New(io.Discard)
	defer func() { err = errors.Join(err, a.Close()) }()
	o := &owner{app: a, root: *root, output: *output, relay: *relay, catalog: *catalog, base: cli.Configuration(a)}
	o.provider = o.base.Vendor
	if o.provider == "" {
		o.provider = "anthropic"
	}
	o.base.System = ""
	o.base.LogPath = ""
	o.base.APIKey = "local-relay-placeholder"
	o.base.MaxTokens = 4096
	o.base.Builtins = []string{"read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command", "wait_for_job", "send_input", "kill_job", "tool_limits"}
	if err = os.Mkdir(*output, 0700); err != nil {
		return err
	}
	switch *step {
	case "C":
		return o.multi()
	case "D-prepare":
		return o.prepare()
	case "D-seed":
		return o.seed()
	case "D2":
		return o.resumeSkill()
	case "offline":
		return o.offline()
	default:
		return fmt.Errorf("unknown step")
	}
}
func (o *owner) save(name string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return o.raw(name, append(b, '\n'))
}
func (o *owner) raw(name string, b []byte) error {
	f, e := os.OpenFile(filepath.Join(o.output, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = f.Write(b)
	return errors.Join(e, f.Close())
}
func (o *owner) config(store, step string, skills bool) ensemble.Config {
	c := o.base
	c.DataDir = filepath.Join(o.root, store)
	c.Workspace = filepath.Join(o.root, "workspace")
	c.BaseURL = o.relay + "/" + step
	if skills {
		c.Builtins = append(append([]string{}, c.Builtins...), "load_skill", "unload_skill")
		c.Skills = &ensemble.SkillConfig{Directory: o.catalog, Primary: "base", Variables: map[string]string{}}
	} else {
		c.Skills = nil
	}
	return c
}
func (o *owner) open(store, step string, skills bool) (*ensemble.Agent, error) {
	return o.app.OpenSession(ensemble.SessionOptions{Config: o.config(store, step, skills)})
}
func policy(a *ensemble.Agent, n int) error {
	_, e := a.UpdatePolicy(a.ExecutionPolicy().Revision, json.RawMessage(fmt.Sprintf(`{"max_model_requests":%d}`, n)))
	return e
}
func (o *owner) view(name string, a *ensemble.Agent) error {
	s, w, e := a.Watch()
	if e != nil {
		return e
	}
	defer w.Close()
	return o.save(name, s)
}
func (o *owner) prompt(a *ensemble.Agent, text string) (ensemble.Completion, error) {
	h, e := a.Submit(strings.ReplaceAll(text, "-P", "-"+o.provider))
	if e != nil {
		return ensemble.Completion{}, e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	v, e := h.Wait(ctx)
	if e != nil {
		cancelErr := h.Cancel()
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()
		v, err := h.Wait(ctx2)
		return v, errors.Join(e, cancelErr, err)
	}
	if v.Outcome != "success" {
		return v, fmt.Errorf("request outcome %s", v.Outcome)
	}
	return v, nil
}
func (o *owner) multi() error {
	a, e := o.open("A", "C1", false)
	if e != nil {
		return e
	}
	defer a.Close()
	b, e := o.open("C-fresh", "C2", false)
	if e != nil {
		return e
	}
	defer b.Close()
	for _, x := range []*ensemble.Agent{a, b} {
		if e = policy(x, 2); e != nil {
			return e
		}
	}
	old, e := a.ExportCheckpoint()
	if e != nil {
		return e
	}
	if e = o.raw("C-old.json", old.Bytes); e != nil {
		return e
	}
	type result struct {
		Role       string
		Completion ensemble.Completion
		Error      string
	}
	results := make(chan result, 2)
	for _, work := range []struct {
		role, text string
		a          *ensemble.Agent
	}{{"C1", "Remember CH10-C-P in this conversation and repeat it once. Do not call tools.", a}, {"C2", "Reply ONLY FRESH-P. Do not call tools.", b}} {
		go func(role, text string, x *ensemble.Agent) {
			c, e := o.prompt(x, text)
			r := result{Role: role, Completion: c}
			if e != nil {
				r.Error = e.Error()
			}
			results <- r
		}(work.role, work.text, work.a)
	}
	var failures []error
	for i := 0; i < 2; i++ {
		r := <-results
		if e = o.save(r.Role+"-completion.json", r); e != nil {
			return e
		}
		if r.Error != "" {
			failures = append(failures, fmt.Errorf("%s: %s", r.Role, r.Error))
		}
	}
	if e = o.view("C1-watch.json", a); e != nil {
		return e
	}
	if e = o.view("C2-watch.json", b); e != nil {
		return e
	}
	fresh, e := a.ExportCheckpoint()
	if e != nil {
		return e
	}
	if e = o.raw("C-new.json", fresh.Bytes); e != nil {
		return e
	}
	if e = errors.Join(a.Close(), b.Close()); e != nil {
		return e
	}
	log, e := boundedFile(filepath.Join(o.root, "A", "events.log"))
	if e != nil {
		return e
	}
	if e = o.raw("C-events.log", log); e != nil {
		return e
	}
	// Closed source owners precede importing the same SessionID.
	imported, e := o.app.ImportSession(fresh.Bytes, ensemble.SessionOptions{Config: o.config("C-import", "C3", false)})
	if e != nil {
		return e
	}
	defer imported.Close()
	origin, e := boundedFile(filepath.Join(o.root, "C-import", "origin.json"))
	if e != nil {
		return e
	}
	if e = o.view("C3-before.json", imported); e != nil {
		return e
	}
	var preOriginRequest uint64
	for _, event := range a.Events() {
		if event.Type == "request_sent" {
			preOriginRequest = event.Seq
			break
		}
	}
	if preOriginRequest == 0 {
		return fmt.Errorf("genuine pre-origin request missing")
	}
	_, historyErr := imported.ReconstructRequest(preOriginRequest)
	var se *ensemble.SessionError
	if !errors.As(historyErr, &se) || se.Code != "history_unavailable" {
		return fmt.Errorf("pre-origin refusal missing: %v", historyErr)
	}
	dump, e := imported.Dump()
	if e != nil {
		return e
	}
	if e = o.raw("C3-original-history.jsonl", dump); e != nil {
		return e
	}
	if len(failures) == 0 {
		if e = policy(imported, 2); e != nil {
			return e
		}
		v, turnErr := o.prompt(imported, "What marker beginning CH10-C is in this conversation? Answer without tools.")
		if e = o.save("C3-completion.json", map[string]any{"completion": v, "error": safeError(turnErr)}); e != nil {
			return e
		}
		if turnErr != nil {
			failures = append(failures, turnErr)
		}
	}
	usageBeforeReopen := imported.UsageByModel()
	if _, e = imported.Checkpoint(); e != nil {
		return e
	}
	if e = imported.Close(); e != nil {
		return e
	}
	reopened, e := o.open("C-import", "C3", false)
	if e != nil {
		return e
	}
	defer reopened.Close()
	current, e := boundedFile(filepath.Join(o.root, "C-import", "origin.json"))
	if e != nil {
		return e
	}
	if !reflect.DeepEqual(usageBeforeReopen, reopened.UsageByModel()) {
		return fmt.Errorf("usage changed during replay")
	}
	if !bytes.Equal(origin, current) {
		return fmt.Errorf("origin changed")
	}
	if e = o.view("C3-reopened.json", reopened); e != nil {
		return e
	}
	if e = o.save("origin-result.json", map[string]any{"unchanged": true, "sha256": fmt.Sprintf("%x", sha256.Sum256(origin)), "pre_origin_error": se.Code, "pre_origin_request_seq": preOriginRequest, "usage_restored_once": true}); e != nil {
		return e
	}
	return errors.Join(failures...)
}
func safeError(e error) string {
	if e != nil {
		return e.Error()
	}
	return ""
}
func boundedFile(path string) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, 2*1024*1024+1))
	if len(b) > 2*1024*1024 {
		return nil, fmt.Errorf("support small-input ceiling")
	}
	return b, e
}
func (o *owner) prepare() error {
	a, e := o.open("D", "D1", true)
	if e != nil {
		return e
	}
	defer a.Close()
	if _, e = a.LoadSkill("scratch"); e != nil {
		return e
	}
	if _, e = a.UnloadSkill("scratch"); e != nil {
		return e
	}
	if _, e = a.LoadSkill("scratch"); e != nil {
		return e
	}
	return o.view("D-prepared.json", a)
}
func (o *owner) seed() error {
	a, e := o.open("D", "D2", true)
	if e != nil {
		return e
	}
	defer a.Close()
	before := a.Config()
	var n atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := n.Add(1)
		if i > 2 {
			http.Error(w, "fixture ceiling", 429)
			return
		}
		body, e := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024+1))
		if e != nil || len(body) > 2*1024*1024 {
			http.Error(w, "fixture body", 400)
			return
		}
		if e = o.raw(fmt.Sprintf("seed-%d-request.json", i), body); e != nil {
			http.Error(w, "fixture recording", 500)
			return
		}
		response := seedResponse(o.provider, i == 1)
		if e = o.raw(fmt.Sprintf("seed-%d-response.json", i), []byte(response)); e != nil {
			http.Error(w, "fixture recording", 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, response)
	}))
	defer server.Close()
	c := before
	c.BaseURL = server.URL
	c.APIKey = "local-seed-placeholder"
	c.DisableStreaming = true
	if e = a.SetConfig(c); e != nil {
		return e
	}
	v, turnErr := o.prompt(a, "LOCAL FIXTURE: seed one-shot tool_limits; this is not a provider demonstration.")
	if e = a.SetConfig(before); e != nil {
		return e
	}
	if e = o.save("seed-result.json", map[string]any{"source": "loopback fixture, not provider output", "requests": n.Load(), "completion": v, "error": safeError(turnErr)}); e != nil {
		return e
	}
	if turnErr != nil {
		return turnErr
	}
	if n.Load() != 2 {
		return fmt.Errorf("seed expected two local requests")
	}
	return o.view("D-seeded.json", a)
}
func seedResponse(vendor string, first bool) string {
	switch vendor {
	case "openai":
		if first {
			return `{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"ch10-local-seed","type":"function","function":{"name":"tool_limits","arguments":"{\"ai_callback_pattern\":\"\",\"max_output_bytes\":17}"}}]},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
		}
		return `{"choices":[{"message":{"role":"assistant","content":"LOCAL FIXTURE limits retained"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
	case "gemini":
		if first {
			return `{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"id":"ch10-local-seed","name":"tool_limits","args":{"ai_callback_pattern":"","max_output_bytes":17}}}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
		}
		return `{"candidates":[{"content":{"role":"model","parts":[{"text":"LOCAL FIXTURE limits retained"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
	default:
		if first {
			return `{"content":[{"type":"tool_use","id":"ch10-local-seed","name":"tool_limits","input":{"ai_callback_pattern":"","max_output_bytes":17}}],"usage":{"input_tokens":1,"output_tokens":1}}`
		}
		return `{"content":[{"type":"text","text":"LOCAL FIXTURE limits retained"}],"usage":{"input_tokens":1,"output_tokens":1}}`
	}
}
func (o *owner) resumeSkill() error {
	a, e := o.open("D", "D2", true)
	if e != nil {
		return e
	}
	defer a.Close()
	if e = policy(a, 3); e != nil {
		return e
	}
	if e = o.view("D2-before.json", a); e != nil {
		return e
	}
	boundary := a.Snapshot().LastSeq
	v, turnErr := o.prompt(a, "Use write_file once to create after-resume.txt containing CH10-D-P and one newline, then confirm completion.")
	if e = o.save("D2-completion.json", map[string]any{"completion": v, "error": safeError(turnErr)}); e != nil {
		return e
	}
	if e = o.view("D2-after.json", a); e != nil {
		return e
	}

	var consumed, called uint64
	count := 0
	for _, event := range a.Events() {
		if event.Seq <= boundary {
			continue
		}
		if event.Type == "tool_limits_consumed" {
			count++
			consumed = event.Seq
		}
		if event.Type == "tool_called" && called == 0 {
			called = event.Seq
		}
	}
	if e = o.save("D2-consumption.json", map[string]any{"count": count, "consumed_seq": consumed, "first_call_seq": called, "observed": count == 1 && consumed < called}); e != nil {
		return e
	}
	if count != 1 || consumed == 0 || called <= consumed {
		return errors.Join(turnErr, fmt.Errorf("next attempted call/once-only limits consumption missing"))
	}
	return turnErr
}
func (o *owner) offline() error {
	// Inputs are originals from C, supplied as root; no HTTP and no live owner.
	old, e := boundedFile(filepath.Join(o.root, "C-old.json"))
	if e != nil {
		return e
	}
	latest, e := boundedFile(filepath.Join(o.root, "C-new.json"))
	if e != nil {
		return e
	}
	log, e := boundedFile(filepath.Join(o.root, "C-events.log"))
	if e != nil {
		return e
	}
	var raw map[string]json.RawMessage
	if e = json.Unmarshal(old, &raw); e != nil {
		return e
	}
	raw["state"] = json.RawMessage("null")
	raw["state_sha256"] = json.RawMessage("null")
	null, e := json.Marshal(raw)
	if e != nil {
		return e
	}
	var baseline []byte
	var events []ensemble.Event
	var snapshot ensemble.Context
	var usage map[ensemble.Provenance]ensemble.Usage
	var skills ensemble.SkillInspection
	requests := map[uint64][]byte{}
	var watchBytes []byte
	for _, branch := range []struct {
		name string
		cp   []byte
	}{{"latest", latest}, {"older", old}, {"null", null}} {
		dir := filepath.Join(o.output, branch.name)
		if e = os.Mkdir(dir, 0700); e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(dir, "events.log"), log, 0600); e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(dir, "checkpoint.json"), branch.cp, 0600); e != nil {
			return e
		}
		inspect, e := o.app.InspectSession(dir)
		if e != nil {
			return e
		}
		c := o.base
		c.APIKey = ""
		c.BaseURL = "http://127.0.0.1:1"
		c.Skills = nil
		rendered, e := inspect.Render(c)
		if e != nil {
			return e
		}
		if baseline == nil {
			baseline = rendered
			events = inspect.Events()
			snapshot = inspect.Snapshot()
			usage = inspect.UsageByModel()
			skills = inspect.InspectSkills()
		} else if !bytes.Equal(baseline, rendered) || !reflect.DeepEqual(events, inspect.Events()) || !sameContext(snapshot, inspect.Snapshot()) || !reflect.DeepEqual(usage, inspect.UsageByModel()) || !reflect.DeepEqual(skills, inspect.InspectSkills()) {
			current := inspect.Snapshot()
			differences := map[string]any{}
			left, right := reflect.ValueOf(snapshot), reflect.ValueOf(current)
			for i := 0; i < left.NumField(); i++ {
				if !reflect.DeepEqual(left.Field(i).Interface(), right.Field(i).Interface()) {
					differences[left.Type().Field(i).Name] = map[string]any{"latest": left.Field(i).Interface(), "branch": right.Field(i).Interface(), "latest_go": fmt.Sprintf("%#v", left.Field(i).Interface()), "branch_go": fmt.Sprintf("%#v", right.Field(i).Interface())}
				}
			}
			if e = o.save(branch.name+"-mismatch.json", map[string]any{"render_equal": bytes.Equal(baseline, rendered), "events_equal": reflect.DeepEqual(events, inspect.Events()), "usage_equal": reflect.DeepEqual(usage, inspect.UsageByModel()), "skills_equal": reflect.DeepEqual(skills, inspect.InspectSkills()), "context_differences": differences}); e != nil {
				return e
			}
			return fmt.Errorf("offline branch mismatch")
		}
		if e = o.raw(branch.name+"-render.json", rendered); e != nil {
			return e
		}
		for _, event := range inspect.Events() {
			if event.Type == "request_sent" {
				b, e := inspect.ReconstructRequest(event.Seq)
				if e != nil {
					return e
				}
				if branch.name == "latest" {
					requests[event.Seq] = b
				} else if !bytes.Equal(requests[event.Seq], b) {
					return fmt.Errorf("reconstructed request mismatch")
				}
				if e = o.raw(fmt.Sprintf("%s-request-%d.json", branch.name, event.Seq), b); e != nil {
					return e
				}
			}
		}

		// Public mounted watch on disposable branches only. No Submit/Ask is
		// reachable here; endpoint remains disabled. Close each before the next.
		c.DataDir = dir
		c.LogPath = ""
		c.System = ""
		c.APIKey = "disabled-offline"
		live, e := o.app.OpenSession(ensemble.SessionOptions{Config: c})
		if e != nil {
			return e
		}
		view, watch, e := live.Watch()
		if e != nil {
			live.Close()
			return e
		}
		watch.Close()
		if e = o.save(branch.name+"-watch.json", view); e != nil {
			live.Close()
			return e
		}
		// Epoch/revision/session checkpoint anchor vary by branch; retained
		// coordinates, window, usage, Skills and historical ownership may not.
		normalized := map[string]any{"first": view.FirstSeq, "last": view.LastSeq, "log": view.LogSeq, "omitted": view.Omitted, "events": view.Events, "partials": view.Partials, "usage": view.State.Usage, "skills": view.State.Skills, "jobs": view.State.JobAccess}
		encoded, e := json.Marshal(normalized)
		if e != nil {
			live.Close()
			return e
		}
		if watchBytes == nil {
			watchBytes = encoded
		} else if !bytes.Equal(watchBytes, encoded) {
			live.Close()
			return fmt.Errorf("normalized restored watch mismatch")
		}
		if e = live.Close(); e != nil {
			return e
		}
	}
	return o.save("offline-result.json", map[string]any{"equal_render_context_history_usage_skills_watch_requests": true, "provider_calls": 0, "inputs": "identical current Config; endpoints disabled; no new prompt admitted"})
}

// Only identity schemas use semantic JSON comparison. Replay Raw fields in the
// rest of Context retain exact byte comparison. UseNumber avoids binary64 loss;
// this bounded CLI catalog uses identical numeric lexemes in its fixed schemas.
func sameContext(left, right ensemble.Context) bool {
	if left.Session == nil || right.Session == nil {
		return reflect.DeepEqual(left, right)
	}
	a, b := *left.Session, *right.Session
	encode := func(v any) any {
		raw, _ := json.Marshal(v)
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var value any
		if d.Decode(&value) != nil {
			return nil
		}
		return value
	}
	if !reflect.DeepEqual(encode(a.Identity), encode(b.Identity)) {
		return false
	}
	a.Identity = nil
	b.Identity = nil
	left.Session = &a
	right.Session = &b
	return reflect.DeepEqual(left, right)
}
