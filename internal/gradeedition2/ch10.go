package gradeedition2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter10 preserves the six original persistence weights for both real clients
// and uses corrected Chapter9 parity. Saved context/event internals stay opaque;
// only the public four fields and sequence identity are inspected.
func Chapter10(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	checks := persistenceChecks(bin, gui)
	parityBin, cleanup, err := persistenceParityBinary(bin)
	if err != nil {
		return grade.Report{}, err
	}
	defer cleanup()
	previous, err := Chapter9(parityBin, core, gui, guiModel, consumer, workflow, exercise, streamModel, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch9-parity", Title: "All corrected Chapter 9 checks pass", Points: 15, Earned: 15, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 10 — Persistence", append(checks, parity), ""), nil
}

// Chapter10Executable preserves the six existing CLI persistence checks.
// The C# GUI has a WebSocket/host lifetime rather than stdin/EOF; its restart
// exercise is separate evidence, not silently counted as another CLI pass.
func Chapter10Executable(bin string) grade.Report {
	return grade.NewTitledReport("Edition 3 C#, Chapter 10 — CLI persistence (excludes GUI and inherited Go checks)", persistenceClient(bin, nil), "")
}

// Earlier CLI probes are independent conversations, even when their harness
// reuses a working directory. Persistence makes that old fixture assumption
// false. Give each no-argument parity process its own explicit save, retaining
// its cwd/files and passing offline dump/render commands through unchanged.
// The enclosing grader owns and removes every file; persistenceChecks separately
// exercises the unwrapped application's default and overridden save paths.
func persistenceParityBinary(bin string) (string, func(), error) {
	dir, err := os.MkdirTemp("", "edition2-persistence-parity-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	path := filepath.Join(dir, "launch")
	script := "#!/bin/sh\nif [ \"$#\" -eq 0 ]; then\n exec " + quote(bin) + " --save " + quote(dir) + "/save-$$.json\nfi\nexec " + quote(bin) + " \"$@\"\n"
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		cleanup()
		return "", nil, err
	}
	return path, cleanup, nil
}

// Both clients own startup and shutdown. A passing core run cannot establish
// that GUI EOF flushes work or writes the selected file before exiting.
func persistenceChecks(bin, gui string) []grade.Check {
	checks := persistenceClient(bin, nil)
	guiBin, cleanup, err := grade.Build(gui)
	if err != nil {
		for i := range checks {
			fail(&checks[i], "GUI build: %v", err)
		}
		return checks
	}
	defer cleanup()
	static, err := filepath.Abs(filepath.Join(gui, "web", "gui"))
	if err != nil {
		for i := range checks {
			fail(&checks[i], "GUI assets: %v", err)
		}
		return checks
	}
	// Port zero asks the OS for an unused port. These checks use the existing
	// stdin/EOF persistence contract; WebSocket behavior has its own parity suite.
	optional := persistenceClient(guiBin, []string{"--port", "0", "--gui-dir", static})
	for i := range checks {
		for j := range checks[i].Details {
			checks[i].Details[j] = "core: " + checks[i].Details[j]
		}
		if !optional[i].Passed {
			fail(&checks[i], "GUI: %s", strings.Join(optional[i].Details, "; "))
		}
	}
	return checks
}

const persistenceKey = "fake-persistence-private-key"

// A run retains observed HTTP and the save from its own process. Keeping the
// endpoint permits a concrete leak check without reading any real credentials.
type persistenceRun struct {
	requests []fakevendor.Recorded
	saved    []byte
	stderr   string
	endpoint string
	err      error
}

// The fake replaces only HTTP. Stdin closure is the application lifecycle
// boundary; neither a fabricated completion event nor a sleep ends the turn.
func runPersistence(bin, work string, args []string, vendor, model, skills, primary string, prompts ...string) persistenceRun {
	replies := make([]fakevendor.Reply, len(prompts))
	var input strings.Builder
	for i, prompt := range prompts {
		replies[i] = fakevendor.Reply{Text: "PERSISTED_REPLY_" + prompt, Usage: fakevendor.Canonical{Input: 100, Output: 30}}
		line, _ := json.Marshal(map[string]string{"kind": "prompt", "text": prompt})
		input.Write(line)
		input.WriteByte('\n')
	}
	peer := fakevendor.New(replies)
	defer peer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "LLM_VENDOR="+vendor, "LLM_MODEL="+model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY="+persistenceKey, "LLM_BASE_URL="+peer.URL(), "EN_DISABLE_STREAMING=1", "EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL="+primary, "CH02_LOG="+filepath.Join(work, "events.jsonl"))
	cmd.Stdin = strings.NewReader(input.String())
	cmd.Stdout = io.Discard
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		err = fmt.Errorf("EOF did not finish within 10 seconds: %w", ctx.Err())
	}
	return persistenceRun{requests: peer.Requests(), stderr: stderr.String(), endpoint: peer.URL(), err: err}
}

// Fields remain raw JSON so splicing a snapshot never assumes its representation.
// Sequence numbers are the sole event field named by the persistence contract.
type persistenceSave struct {
	fields map[string]json.RawMessage
	anchor uint64
	log    []json.RawMessage
}

// Inspect public vocabulary without deserializing into either solution's types.
// That preserves compatibility with different private context representations.
func readPersistenceSave(data []byte) (persistenceSave, error) {
	var s persistenceSave
	if err := json.Unmarshal(data, &s.fields); err != nil {
		return s, err
	}
	for key := range s.fields {
		switch key {
		case "config", "as_of", "context", "log", "log_mode", "dialogue_log":
		default:
			return s, fmt.Errorf("unknown save field %s", key)
		}
	}
	for _, key := range []string{"config", "as_of", "context", "log"} {
		if s.fields[key] == nil {
			return s, fmt.Errorf("missing %s", key)
		}
	}
	if err := json.Unmarshal(s.fields["as_of"], &s.anchor); err != nil {
		return s, err
	}
	if err := json.Unmarshal(s.fields["log"], &s.log); err != nil {
		return s, err
	}
	return s, nil
}

// After a normal save, the latest retained fact and snapshot anchor agree.
// The log may start above one after a snapshot-only restore; array length is
// deliberately not used as sequence identity.
func persistenceSeqs(s persistenceSave) ([]uint64, error) {
	var seqs []uint64
	for _, raw := range s.log {
		var event map[string]json.RawMessage
		if err := json.Unmarshal(raw, &event); err != nil {
			return nil, err
		}
		var seq uint64
		if err := json.Unmarshal(event["seq"], &seq); err != nil {
			return nil, err
		}
		if seq == 0 || (len(seqs) > 0 && seq <= seqs[len(seqs)-1]) {
			return nil, fmt.Errorf("log sequences are not positive and strictly increasing: %v then %d", seqs, seq)
		}
		seqs = append(seqs, seq)
	}
	if len(seqs) == 0 || seqs[len(seqs)-1] != s.anchor {
		return nil, fmt.Errorf("saved anchor %d does not match nonempty log", s.anchor)
	}
	return seqs, nil
}

func persistenceClient(bin string, clientArgs []string) []grade.Check {
	r := grade.Ch11Result{SaveShapeOK: true, DefaultLoadOK: true, ReplayEqualsSnapshotOK: true, TailAppliedOnceOK: true, LogNotNeededOK: true, BadSaveRefusedOK: true}
	checks := grade.Ch11Checks(r)[:6]
	work, err := os.MkdirTemp("", "edition2-persistence-")
	if err != nil {
		for i := range checks {
			fail(&checks[i], "scratch: %v", err)
		}
		return checks
	}
	defer os.RemoveAll(work)
	// Start without a primary: a supplied body hid the student's original bug
	// where the save recorded empty config instead of the effective wire prompt.
	launch := func(dir string, args []string, vendor, model, skills, primary string, prompts ...string) persistenceRun {
		run := runPersistence(bin, dir, args, vendor, model, skills, primary, prompts...)
		if run.err == nil && len(run.requests) != len(prompts) {
			run.err = fmt.Errorf("want %d requests, got %d", len(prompts), len(run.requests))
		}
		return run
	}
	// One prompt per process avoids treating a second already-buffered input as
	// a live hint. EOF drains that admitted turn; restart itself is the barrier.
	seed := launch(work, clientArgs, "anthropic", "fake-model", "", "", "FIRST_WORD")
	if seed.err != nil {
		for i := range checks {
			fail(&checks[i], "seed process: %v", seed.err)
		}
		return checks
	}
	first := launch(work, clientArgs, "anthropic", "fake-model", "", "", "SECOND_WORD")
	first.saved, err = os.ReadFile(filepath.Join(work, "save.json"))
	if first.err != nil || err != nil {
		for i := range checks {
			fail(&checks[i], "first save: process=%v read=%v", first.err, err)
		}
		return checks
	}
	s2, err := readPersistenceSave(first.saved)
	if err != nil {
		for i := range checks {
			fail(&checks[i], "first save: %v", err)
		}
		return checks
	}
	checkPersistenceShape(&checks[0], first, "anthropic")
	second := launch(work, clientArgs, "anthropic", "fake-model", "", "", "THIRD_WORD")
	second.saved, err = os.ReadFile(filepath.Join(work, "save.json"))
	if second.err != nil || err != nil {
		for i := range checks {
			fail(&checks[i], "second save: process=%v read=%v", second.err, err)
		}
		return checks
	}
	s3, err := readPersistenceSave(second.saved)
	if err != nil {
		for i := range checks {
			fail(&checks[i], "second save: %v", err)
		}
		return checks
	}
	checkPersistenceShape(&checks[0], second, "anthropic")
	checkPersistenceWords(&checks[1], second.requests[0], "FIRST_WORD", "SECOND_WORD", "THIRD_WORD")
	// Saving after restart must retain prior facts, not just emit an equivalent
	// context. Compare JSON values because whitespace is not part of the format.
	if len(s3.log) <= len(s2.log) {
		fail(&checks[0], "second save did not append to loaded log")
	} else {
		for i, raw := range s2.log {
			if !sameJSON(raw, s3.log[i]) {
				fail(&checks[0], "second save changed loaded event %d", i)
				break
			}
		}
		seqs, e := persistenceSeqs(s3)
		if e == nil && seqs[len(s2.log)] != s2.anchor+1 {
			fail(&checks[0], "first new sequence did not follow prior anchor exactly")
		}
	}
	// Reuse one baseline next request for all variants; each receives the same
	// new prompt/reply. These splices modify only the public top-level fields.
	variants := []string{"snapshot", "log-only", "tail", "snapshot-only"}
	var reference []byte
	for _, variant := range variants {
		fields := map[string]json.RawMessage{}
		for k, v := range s3.fields {
			fields[k] = v
		}
		index := 2
		switch variant {
		case "log-only":
			fields["context"] = json.RawMessage("null")
		case "tail":
			fields["context"], fields["as_of"] = s2.fields["context"], s2.fields["as_of"]
			index = 3
		case "snapshot-only":
			fields["log"] = json.RawMessage("[]")
			index = 4
		}
		if variant == "tail" && s3.anchor <= s2.anchor {
			fail(&checks[index], "fixture has no tail above old anchor")
			continue
		}
		dir := filepath.Join(work, variant)
		data, e := json.Marshal(fields)
		if e == nil {
			e = os.MkdirAll(dir, 0700)
		}
		if e == nil {
			e = os.WriteFile(filepath.Join(dir, "save.json"), data, 0600)
		}
		if e != nil {
			fail(&checks[index], "%s fixture: %v", variant, e)
			continue
		}
		run := launch(dir, clientArgs, "anthropic", "fake-model", "", "", "NEXT_WORD")
		if run.err != nil {
			fail(&checks[index], "%s: %v", variant, run.err)
			continue
		}
		body := run.requests[0].Body
		checkPersistenceWords(&checks[index], run.requests[0], "FIRST_WORD", "SECOND_WORD", "THIRD_WORD", "NEXT_WORD")
		// These are the actual next HTTP bodies under identical configuration.
		// Successful recall alone would hide duplicated or missing history.
		if variant == "snapshot" {
			reference = body
		} else if len(reference) == 0 || !bytes.Equal(reference, body) {
			fail(&checks[index], "%s next request differs from full snapshot", variant)
		}
		saved, e := os.ReadFile(filepath.Join(dir, "save.json"))
		post, decode := readPersistenceSave(saved)
		if e != nil || decode != nil {
			fail(&checks[index], "%s exit save: %v %v", variant, e, decode)
			continue
		}
		seqs, e := persistenceSeqs(post)
		if e != nil {
			fail(&checks[index], "%s next numbering: %v", variant, e)
			continue
		}
		// A trimmed input cannot recreate discarded log entries when resaved.
		// Its first new event still follows the retained snapshot anchor.
		offset := len(s3.log)
		if variant == "snapshot-only" {
			offset = 0
		}
		if len(seqs) <= offset || seqs[offset] != s3.anchor+1 {
			fail(&checks[index], "%s first new seq must equal %d", variant, s3.anchor+1)
		}
	}
	checkPersistenceOverride(&checks[1], bin, work, clientArgs, first.saved)
	checkPersistenceCurrentConfig(&checks[1], bin, work, clientArgs, second.saved)
	backup, backupErr := os.ReadFile(filepath.Join(work, "save.json.bak"))
	if backupErr != nil || !sameJSON(backup, first.saved) {
		fail(&checks[5], "previous validated primary not retained as backup: %v", backupErr)
	}
	checkPersistenceBackup(&checks[5], bin, work, clientArgs, first.saved)
	checkPersistenceBad(&checks[5], bin, work, clientArgs)
	return checks
}

// On-disk formatting is free; request byte equality above is a separate claim.
func sameJSON(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}

func checkPersistenceWords(c *grade.Check, r fakevendor.Recorded, words ...string) {
	// Reply markers contain the prompt too; inspect user messages, not substrings
	// of the whole request that an assistant answer could accidentally satisfy.
	messages, _ := r.JSON()["messages"].([]any)
	var text strings.Builder
	for _, raw := range messages {
		m := object(raw)
		if m["role"] == "user" {
			value, _ := json.Marshal(m["content"])
			text.Write(value)
		}
	}
	for _, word := range words {
		if strings.Count(text.String(), word) != 1 {
			fail(c, "want exactly one user prompt %s in next request", word)
		}
	}
}

func checkPersistenceShape(c *grade.Check, run persistenceRun, vendor string) {
	saved, err := readPersistenceSave(run.saved)
	if err != nil {
		fail(c, "save shape: %v", err)
		return
	}
	if _, err := persistenceSeqs(saved); err != nil {
		fail(c, "%v", err)
	}
	if bytes.Equal(bytes.TrimSpace(saved.fields["context"]), []byte("null")) {
		fail(c, "normal exit omitted context snapshot")
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(saved.fields["config"], &config); err != nil {
		fail(c, "config: %v", err)
		return
	}
	for _, field := range []string{"model", "vendor", "system_prompt", "tools"} {
		if config[field] == nil {
			fail(c, "config lacks %s", field)
		}
	}
	// Observe effective configuration at the transport boundary. Comparing with
	// supplied environment strings would repeat the raw/default prompt mistake.
	wire := run.requests[len(run.requests)-1]
	expected := map[string]any{"model": wire.JSON()["model"], "vendor": vendor, "system_prompt": persistenceSystem(wire, vendor), "tools": persistenceTools(wire, vendor)}
	for key, value := range expected {
		raw, _ := json.Marshal(value)
		if !sameJSON(config[key], raw) {
			fail(c, "saved config.%s differs from actual wire", key)
		}
	}
	if bytes.Contains(run.saved, []byte(persistenceKey)) || bytes.Contains(run.saved, []byte(run.endpoint)) {
		fail(c, "save leaked provider credential or endpoint")
	}
}

// The no-skill phase uses Anthropic; the runtime-switch phase uses OpenAI.
// Rich three-vendor neutral replay remains covered by the accepted source tests.
func persistenceSystem(r fakevendor.Recorded, vendor string) string {
	if vendor == "anthropic" {
		return skillSystem(r)
	}
	for _, raw := range list(r.JSON()["messages"]) {
		m := object(raw)
		if m["role"] == "system" {
			s, _ := m["content"].(string)
			return s
		}
	}
	return ""
}

// Normalize only vendor declaration wrappers, not names, descriptions or
// schemas. The save's public declaration fields must preserve all three.
func persistenceTools(r fakevendor.Recorded, vendor string) []map[string]any {
	out := []map[string]any{}
	for _, raw := range list(r.JSON()["tools"]) {
		t := object(raw)
		schema := "input_schema"
		if vendor == "openai" {
			t = object(t["function"])
			schema = "parameters"
		}
		out = append(out, map[string]any{"name": t["name"], "description": t["description"], "schema": t[schema]})
	}
	return out
}

// Seed only the chosen file. Recall proves load used it; changed bytes and
// absence of the default file establish the other half of the flag contract.
func checkPersistenceOverride(c *grade.Check, bin, work string, args []string, save []byte) {
	dir := filepath.Join(work, "override")
	if err := os.MkdirAll(dir, 0700); err != nil {
		fail(c, "%v", err)
		return
	}
	path := filepath.Join(dir, "chosen.json")
	if err := os.WriteFile(path, save, 0600); err != nil {
		fail(c, "%v", err)
		return
	}
	run := runPersistence(bin, dir, append(append([]string{}, args...), "--save", path), "anthropic", "fake-model", "", "", "OVERRIDE_WORD")
	if run.err != nil || len(run.requests) != 1 {
		fail(c, "override launch: %v (%d requests)", run.err, len(run.requests))
		return
	}
	checkPersistenceWords(c, run.requests[0], "FIRST_WORD", "SECOND_WORD", "OVERRIDE_WORD")
	after, err := os.ReadFile(path)
	if err != nil || bytes.Equal(after, save) {
		fail(c, "override path not saved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "save.json")); !os.IsNotExist(err) {
		fail(c, "override also created default save")
	}
}

// Change model, vendor, primary instruction and visible tools together after
// saving. The old conversation must survive without reinstalling old settings.
func checkPersistenceCurrentConfig(c *grade.Check, bin, work string, args []string, save []byte) {
	dir := filepath.Join(work, "current-config")
	skills := filepath.Join(dir, "skills", "current")
	if err := os.MkdirAll(skills, 0700); err != nil {
		fail(c, "%v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(skills, "SKILL.md"), []byte("---\nname: current\ndescription: Current runtime identity\ntype: primary\ntools: read_file\n---\nCURRENT_RUNTIME_PROMPT"), 0600); err != nil {
		fail(c, "%v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "save.json"), save, 0600); err != nil {
		fail(c, "%v", err)
		return
	}
	// This row exists in the legacy table too; an unknown name would test model
	// refusal before reaching the restoration behavior being graded.
	run := runPersistence(bin, dir, args, "openai", "gpt-5-course", filepath.Dir(skills), "current", "CURRENT_WORD")
	if run.err != nil || len(run.requests) != 1 {
		fail(c, "current config launch: %v (%d requests): %s", run.err, len(run.requests), strings.TrimSpace(run.stderr))
		return
	}
	request := run.requests[0]
	if request.JSON()["model"] != "gpt-5-course" || strings.TrimSpace(persistenceSystem(request, "openai")) != "CURRENT_RUNTIME_PROMPT" {
		fail(c, "historical model/vendor/prompt replaced runtime configuration")
	}
	var names []string
	for _, tool := range persistenceTools(request, "openai") {
		name, _ := tool["name"].(string)
		names = append(names, name)
	}
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"load_skill", "read_file", "unload_skill"}) {
		fail(c, "historical declarations replaced current tools: %v", names)
	}
	checkPersistenceWords(c, request, "FIRST_WORD", "SECOND_WORD", "THIRD_WORD", "CURRENT_WORD")
	run.saved, _ = os.ReadFile(filepath.Join(dir, "save.json"))
	checkPersistenceShape(c, run, "openai")
}

// Syntax failure and JSON null are distinct invalid inputs. Nonzero status
// alone misses a deferred save that overwrites the rejected file; preserve its
// bytes and observe the absence of HTTP as independent refusal effects.
func checkPersistenceBad(c *grade.Check, bin, work string, args []string) {
	for i, bad := range []string{`{"config":`, "null"} {
		dir := filepath.Join(work, fmt.Sprintf("bad-%d", i))
		if err := os.MkdirAll(dir, 0700); err != nil {
			fail(c, "%v", err)
			return
		}
		path := filepath.Join(dir, "save.json")
		if err := os.WriteFile(path, []byte(bad), 0600); err != nil {
			fail(c, "%v", err)
			return
		}
		backupPath := path + ".bak"
		if err := os.WriteFile(backupPath, []byte("null"), 0600); err != nil {
			fail(c, "%v", err)
			return
		}
		run := runPersistence(bin, dir, args, "anthropic", "fake-model", "", "", "MUST_NOT_SEND")
		backupAfter, backupErr := os.ReadFile(backupPath)
		if backupErr != nil || string(backupAfter) != "null" {
			fail(c, "invalid backup was overwritten: %v", backupErr)
		}
		after, err := os.ReadFile(path)
		var exit *exec.ExitError
		// A deadline or spawn failure is not a correct nonzero refusal. Require
		// a diagnostic, but the contract prescribes no literal filename or format.
		if run.err != nil {
			exit, _ = run.err.(*exec.ExitError)
		}
		diagnosed := strings.TrimSpace(run.stderr) != ""
		if exit == nil || !diagnosed || len(run.requests) != 0 || err != nil || string(after) != bad {
			fail(c, "bad save %d: exit=%v diagnosed=%t requests=%d unchanged=%t", i, run.err, diagnosed, len(run.requests), err == nil && string(after) == bad)
		}
	}
}

// Recovery chooses one independently valid snapshot, not a fabricated splice
// from a broken primary. Successful exit must retain that recovery copy.
func checkPersistenceBackup(c *grade.Check, bin, work string, args []string, backup []byte) {
	for _, missing := range []bool{false, true} {
		dir := filepath.Join(work, fmt.Sprintf("backup-missing-%t", missing))
		if err := os.MkdirAll(dir, 0700); err != nil {
			fail(c, "backup fixture: %v", err)
			return
		}
		path := filepath.Join(dir, "save.json")
		if err := os.WriteFile(path+".bak", backup, 0600); err != nil {
			fail(c, "backup fixture: %v", err)
			return
		}
		if !missing {
			if err := os.WriteFile(path, []byte(`{"config":`), 0600); err != nil {
				fail(c, "primary fixture: %v", err)
				return
			}
		}
		run := runPersistence(bin, dir, args, "anthropic", "fake-model", "", "", "RECOVERED_WORD")
		if run.err != nil || len(run.requests) != 1 {
			fail(c, "backup fallback missing=%t: %v (%d requests)", missing, run.err, len(run.requests))
			continue
		}
		checkPersistenceWords(c, run.requests[0], "FIRST_WORD", "SECOND_WORD", "RECOVERED_WORD")
		// Recovery notices have no prescribed format or output channel. Source
		// and manual review check that they identify backup recovery and its
		// saved boundary; this finite probe checks state and file preservation.
		after, err := os.ReadFile(path + ".bak")
		if err != nil || !bytes.Equal(after, backup) {
			fail(c, "backup fallback missing=%t replaced recovery copy: %v", missing, err)
		}
		primary, err := os.ReadFile(path)
		if err != nil {
			fail(c, "recovered primary: %v", err)
		} else if _, err := readPersistenceSave(primary); err != nil {
			fail(c, "recovered primary: %v", err)
		}
	}
}
