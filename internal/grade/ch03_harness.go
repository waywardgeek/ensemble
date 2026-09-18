package grade

// Chapter 3 harness: drive the submission through the tool loop, recording
// evidence and judging nothing.
//
// Record then judge, as in Chapter 2. One run surfaces every bug the student
// has instead of one bug per run.
//
// Every scenario below runs the agent in a FRESH workspace containing known
// files, so that a check can ask the only question that really matters about a
// file tool: did the bytes on disk change the way they were supposed to?

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

const Ch3Timeout = 120 * time.Second

// --- workspace fixtures ----------------------------------------------------

// The workspace every scenario runs in. `go` is the one binary every student
// is guaranteed to have, since the course requires it, so the shell scenarios
// are written in Go rather than in shell.
const (
	ch3GoMod = "module ch03work\n\ngo 1.21\n"

	ch3Exit7 = `package main

import "os"

func main() { os.Exit(7) }
`

	ch3Noisy = `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "STDOUT_MARKER")
	fmt.Fprintln(os.Stderr, "STDERR_MARKER")
}
`

	// Five distinct lines, so that a range read can be told apart from a
	// whole-file read by CONTENT rather than by trusting a byte count.
	ch3Notes = `alpha line one
beta line two
gamma line three
delta line four
epsilon line five
`

	// The file the declined decision is exercised against. It deliberately
	// does NOT contain the anchor the model will ask for.
	ch3Contract = `first line of the contract file
second line of the contract file
`
)

// Ch3OrderingLog exists for exactly one assertion: Anthropic requires a
// tool_result to come FIRST in the content array of the message answering it.
//
// That rule is only observable in a message carrying a tool_result AND
// something else, and the ORDER of the events is what makes it observable: the
// human turn at seq 4 lands while the tool call from seq 2 is still
// outstanding, so both end up in the same rendered user message. Rendering
// walks the dialogue in order, so a renderer that simply appends produces
// [text, tool_result] — which Anthropic rejects — and one that splices
// produces [tool_result, text].
//
// Chapter 2's exhibit log cannot catch this: its human turn comes AFTER the
// tool return, so the result is already first and flipping the splice changes
// nothing. Measured — before this fixture existed, a mutant that turned the
// splice off scored 100/100 against both chapters' graders.
const Ch3OrderingLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"run the tests"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"text","text":"Running them now."},{"type":"tool_call","call_id":"toolu_order_1","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"name":"run_command","args":{"command":"go test ./..."}}],"usage":{"input":80,"cache_write":0,"cache_read":0,"output":15},"from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"toolu_order_1","name":"run_command","args":{"command":"go test ./..."}}}
{"seq":4,"type":"message_received","time":"2026-01-01T00:00:03Z","message":{"actor":"human","parts":[{"type":"text","text":"also check the logs"}]}}
{"seq":5,"type":"tool_returned","time":"2026-01-01T00:00:04Z","tool":{"call_id":"toolu_order_1","parts":[{"type":"text","text":"ok\n"}]}}
`

func ch3Workspace() (string, error) {
	work, err := os.MkdirTemp("", "ch03-grade-")
	if err != nil {
		return "", err
	}
	files := map[string]string{
		"go.mod":                 ch3GoMod,
		"testdata/exit7/main.go": ch3Exit7,
		"testdata/noisy/main.go": ch3Noisy,
		"notes.md":               ch3Notes,
		"contract.txt":           ch3Contract,
	}
	for name, body := range files {
		full := filepath.Join(work, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			return "", err
		}
	}
	return work, nil
}

// --- evidence --------------------------------------------------------------

// Ch3Session is everything one scripted run produced.
type Ch3Session struct {
	Name     string
	Vendor   string
	Answers  []string
	Stdout   string
	Stderr   string
	Requests []fakevendor.Recorded
	Log      []Ch2LogLine
	LogErr   string
	DumpOut  string
	Protocol []string

	// Before and After are the contents of the watched workspace files, so a
	// check can ask whether a tool actually changed the disk.
	Before map[string]string
	After  map[string]string

	Err string
}

type Ch3Result struct {
	Ch2    []Check // Chapter 2's checks, re-run unchanged against this binary
	Ch2Err string

	Loop        map[string]*Ch3Session // per vendor
	Multiblock  *Ch3Session
	LocalTools  *Ch3Session
	Shell       *Ch3Session
	ToolError   *Ch3Session
	EditContrac *Ch3Session

	// ExhibitRender is Chapter 2's exhibit log re-rendered by this binary.
	//
	// It exists for one assertion the live scenarios cannot make: Anthropic
	// requires a tool_result to come FIRST in the content array of the message
	// answering it, and that is only observable in a message that carries BOTH
	// a tool_result and something else. No scripted session can produce one,
	// because a prompt cannot arrive while the loop is blocked on a tool —
	// that is Chapter 5's mailbox. The exhibit log can: it ends with a tool
	// return followed by a human turn.
	ExhibitRender string
	ExhibitErr    string
}

// WireResult is one tool_result as it appeared in a request BODY — that is, as
// the model would actually receive it. A result that exists only in the
// student's log never reached the model, and the model is the one that has to
// recover from it.
type WireResult struct {
	CallID  string
	Text    string
	IsError bool
	First   bool // was it the first block of its message's content array
	Request int  // index into the session's requests
}

// --- scripted runs ---------------------------------------------------------

func ch3Env(vendor, baseURL, work, logPath string) []string {
	env := append(os.Environ(),
		"LLM_VENDOR="+vendor,
		"LLM_API_KEY=course-grader-fake",
		"LLM_MODEL="+ch2RequestedModel(vendor),
		"CH02_LOG="+logPath,
		"CH03_LOG="+logPath,
	)
	if baseURL != "" {
		env = append(env,
			"LLM_BASE_URL="+baseURL,
			"ANTHROPIC_BASE_URL="+baseURL,
			"OPENAI_BASE_URL="+baseURL,
			"GEMINI_BASE_URL="+baseURL,
		)
	}
	return env
}

// runCh3Session runs one scripted scenario in a throwaway workspace.
//
// watch names the files whose contents are captured before and after the run.
func runCh3Session(bin, name, vendor string, replies []fakevendor.Reply, prompts []string, watch []string) *Ch3Session {
	s := &Ch3Session{Name: name, Vendor: vendor,
		Before: map[string]string{}, After: map[string]string{}}

	work, err := ch3Workspace()
	if err != nil {
		s.Err = err.Error()
		return s
	}
	defer os.RemoveAll(work)

	for _, w := range watch {
		if b, err := os.ReadFile(filepath.Join(work, w)); err == nil {
			s.Before[w] = string(b)
		}
	}

	fake := fakevendor.New(replies)
	defer fake.Close()

	logPath := filepath.Join(work, ".ch03-"+name+".log")
	env := ch3Env(vendor, fake.URL(), work, logPath)

	stdout, stderr, _ := runWithStdin(bin, work, env, nil, prompts)
	s.Stdout, s.Stderr = stdout, stderr
	parseCh3Stdout(s, stdout)
	s.Requests = fake.Requests()

	// `dump` runs in a FRESH process, reading only what was persisted, so the
	// log evidence is what survived to disk rather than what was in memory.
	dumpOut, _, _ := runOnce(bin, work, env, "dump")
	s.DumpOut = dumpOut
	s.Log, s.LogErr = parseLogLines(dumpOut)

	for _, w := range watch {
		if b, err := os.ReadFile(filepath.Join(work, w)); err == nil {
			s.After[w] = string(b)
		}
	}
	return s
}

func parseCh3Stdout(s *Ch3Session, stdout string) {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			s.Protocol = append(s.Protocol, fmt.Sprintf("stdout line is not JSON: %.80q", line))
			continue
		}
		switch {
		case m["assistant"] != nil:
			var v string
			_ = json.Unmarshal(m["assistant"], &v)
			s.Answers = append(s.Answers, v)
		case m["error"] != nil:
			var v string
			_ = json.Unmarshal(m["error"], &v)
			s.Protocol = append(s.Protocol, "program reported error: "+v)
		}
	}
}

// --- the scenarios ---------------------------------------------------------

// ch3LoopReplies is the basic loop, and it is the fixture that collects
// Chapter 1's deferred promise.
//
// Reply 1 carries BOTH a text block and a tool_use block in ONE message. An
// implementation that walks the content array without dispatching on `type`
// cannot produce both the transcript line and the dispatch from it: if it
// treats every block as prose it never runs anything, and if it treats every
// block as a call it has no answer to record.
//
// It also runs for THREE rounds, because a loop that stops after one round
// looks identical to a correct one when the script is only one call long.
func ch3LoopReplies(vendor string) []fakevendor.Reply {
	id := func(n string) string {
		switch vendor {
		case "openai":
			return "call_" + n
		case "gemini":
			return "fc_" + n
		}
		return "toolu_" + n
	}
	return []fakevendor.Reply{
		{
			Text:     "I'll check the toolchain.",
			ToolName: "run_command",
			ToolArgs: `{"command":"go version"}`,
			ToolID:   id("loop_1"),
			Usage:    fakevendor.Canonical{Input: 20, Output: 10},
		},
		{
			Text:     "Now the middle of the notes.",
			ToolName: "read_file",
			ToolArgs: `{"path":"notes.md","start_line":3,"end_line":4}`,
			ToolID:   id("loop_2"),
			Usage:    fakevendor.Canonical{Input: 30, Output: 12},
		},
		{
			Text:  "Toolchain is present and the notes look right.",
			Usage: fakevendor.Canonical{Input: 40, Output: 9},
		},
	}
}

// ch3MultiblockReplies puts TWO calls to the SAME tool in one message, with
// different arguments and therefore different answers.
//
// Same tool on purpose: it means a result cannot be matched to its call by
// NAME, only by id. With two different tools, keying by name and keying by id
// are the same program and the check proves nothing.
func ch3MultiblockReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Text: "Reading the first and last lines.",
			Tools: []fakevendor.ToolCall{
				{ID: "toolu_mb_a", Name: "read_file", Args: `{"path":"notes.md","start_line":1,"end_line":1}`},
				{ID: "toolu_mb_b", Name: "read_file", Args: `{"path":"notes.md","start_line":5,"end_line":5}`},
			},
			Usage: fakevendor.Canonical{Input: 25, Output: 14},
		},
		{
			Text:  "Both ends of the file are accounted for.",
			Usage: fakevendor.Canonical{Input: 35, Output: 8},
		},
	}
}

func ch3LocalToolReplies() []fakevendor.Reply {
	step := func(text, name, args, id string) fakevendor.Reply {
		return fakevendor.Reply{Text: text, ToolName: name, ToolArgs: args, ToolID: id,
			Usage: fakevendor.Canonical{Input: 20, Output: 10}}
	}
	return []fakevendor.Reply{
		step("Creating the file.", "write_file",
			`{"path":"src/greet.go","content":"package src\n\nfunc Greet() string { return \"hello\" }\n"}`,
			"toolu_lt_write"),
		step("Changing the greeting.", "edit_file",
			`{"path":"src/greet.go","old_text":"hello","new_text":"goodbye"}`,
			"toolu_lt_edit"),
		// The three READ probes look at material the harness planted, never
		// at src/greet.go. Measured: when they searched for the new greeting,
		// a sabotaged write_file also failed readtools — the row that was
		// split out precisely so it could not be blamed for the other one.
		step("Looking at the working directory.", "list_directory",
			`{"path":"."}`, "toolu_lt_ls"),
		step("Finding the third note.", "search_files",
			`{"pattern":"gamma","path":"."}`, "toolu_lt_grep"),
		step("Reading the middle of the notes.", "read_file",
			`{"path":"notes.md","start_line":3,"end_line":4}`, "toolu_lt_read"),
		// The guard. notes.md is material the harness planted, so a write
		// without overwrite must be refused and the bytes must not move; the
		// same write with overwrite:true must land. The new-file write at the
		// top of this script is the negative control: no permission needed
		// for a file that is not there.
		step("Replacing the notes.", "write_file",
			`{"path":"notes.md","content":"replaced\n"}`, "toolu_lt_clobber"),
		step("Replacing the notes, on purpose this time.", "write_file",
			`{"path":"notes.md","content":"replaced\n","overwrite":true}`, "toolu_lt_clobber_ok"),
		{Text: "All five local tools are done.", Usage: fakevendor.Canonical{Input: 50, Output: 9}},
	}
}

// ch3ShellReplies exercises the three scenarios the chapter names, plus one it
// does not.
//
// The fourth exists because `go run ./testdata/exit7` CANNOT grade exit-code
// reporting. Two measured reasons: `go run` prints "exit status 7" to its own
// stderr, so a tool that never reports an exit code still has "exit ... 7" in
// its output; and `go run` itself exits 1, not 7, so the number the agent
// actually receives is not the one the fixture is named after.
//
// A bare `exit 7` is silent and exits 7, so the only way for the code to reach
// the model is for the tool to report it.
func ch3ShellReplies() []fakevendor.Reply {
	step := func(text, args, id string) fakevendor.Reply {
		return fakevendor.Reply{Text: text, ToolName: "run_command", ToolArgs: args, ToolID: id,
			Usage: fakevendor.Canonical{Input: 20, Output: 10}}
	}
	return []fakevendor.Reply{
		step("Checking the version.", `{"command":"go version"}`, "toolu_sh_version"),
		step("Running the failing program.", `{"command":"go run ./testdata/exit7"}`, "toolu_sh_gorun"),
		step("Running the noisy program.", `{"command":"go run ./testdata/noisy"}`, "toolu_sh_noisy"),
		step("Exiting non-zero, quietly.", `{"command":"exit 7"}`, "toolu_sh_exit7"),
		{Text: "Shell scenarios complete.", Usage: fakevendor.Canonical{Input: 60, Output: 9}},
	}
}

// ch3ToolErrorReplies asks for three things that cannot work, and one that
// works but fails.
//
// The fourth is the NEGATIVE CONTROL for the whole check: a command that exits
// non-zero RAN. Reporting it as a tool error would be a false claim about the
// call rather than a true report about the result, so `toolerror` asserts that
// the first three are marked as errors AND that the fourth is not.
func ch3ToolErrorReplies() []fakevendor.Reply {
	step := func(text, name, args, id string) fakevendor.Reply {
		return fakevendor.Reply{Text: text, ToolName: name, ToolArgs: args, ToolID: id,
			Usage: fakevendor.Canonical{Input: 20, Output: 10}}
	}
	return []fakevendor.Reply{
		step("Reading a file that is not there.", "read_file",
			`{"path":"no-such-file-anywhere.txt"}`, "toolu_err_missing"),
		step("Calling a tool that does not exist.", "frobnicate",
			`{"x":1}`, "toolu_err_unknown"),
		// Valid JSON, wrong shape. A vendor will never deliver syntactically
		// invalid JSON — the API would reject its own response — so "malformed
		// arguments" in practice means arguments of the wrong TYPE.
		step("Passing arguments of the wrong type.", "read_file",
			`{"path":42}`, "toolu_err_badargs"),
		// A silent non-zero exit: no output at all, exit code 7. `go run` is
		// deliberately NOT used here — it exits 1 and prints "exit status 7"
		// to stderr, which would let a tool that reports no exit code at all
		// satisfy the assertion from vendor noise.
		step("Running a command that exits non-zero.", "run_command",
			`{"command":"exit 7"}`, "toolu_err_exit7"),
		{Text: "I recovered from all of that.", Usage: fakevendor.Canonical{Input: 70, Output: 9}},
	}
}

// ch3EditContractReplies exercises the declined decision: an anchor that is
// not in the file.
//
// The grader must accept refuse, fuzzy-match AND rewrite, so this fixture
// deliberately makes no claim about what should happen — only that SOMETHING
// coherent happens and is reported.
func ch3EditContractReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{
			Text:     "Applying the edit.",
			ToolName: "edit_file",
			ToolArgs: `{"path":"contract.txt","old_text":"this anchor is definitely not in the file","new_text":"REPLACEMENT"}`,
			ToolID:   "toolu_ec_1",
			Usage:    fakevendor.Canonical{Input: 20, Output: 10},
		},
		{Text: "Understood, I will re-read the file.", Usage: fakevendor.Canonical{Input: 30, Output: 9}},
	}
}

// --- run everything --------------------------------------------------------

func Ch3Run(bin string) (*Ch3Result, error) {
	res := &Ch3Result{Loop: map[string]*Ch3Session{}}

	// --- phase 1: Chapter 2 parity, using Chapter 2's own harness ----------
	// Chapter 3 is the first chapter that could plausibly break Chapter 2's
	// work, so the regression guard runs Chapter 2's real checks against the
	// Chapter 3 binary rather than re-implementing a weaker version of them.
	if r2, err := Ch2Run(bin); err != nil {
		res.Ch2Err = err.Error()
	} else {
		res.Ch2 = Ch2Evaluate(r2)
	}

	prompt := []string{`{"user":"do the work"}`}

	for _, vendor := range Ch2Vendors {
		res.Loop[vendor] = runCh3Session(bin, "loop-"+vendor, vendor,
			ch3LoopReplies(vendor), prompt, []string{"notes.md"})
	}

	res.Multiblock = runCh3Session(bin, "multiblock", "anthropic",
		ch3MultiblockReplies(), prompt, []string{"notes.md"})

	res.LocalTools = runCh3Session(bin, "localtools", "anthropic",
		ch3LocalToolReplies(), prompt, []string{"notes.md", "src/greet.go"})

	res.Shell = runCh3Session(bin, "shell", "anthropic",
		ch3ShellReplies(), prompt, nil)

	res.ToolError = runCh3Session(bin, "toolerror", "anthropic",
		ch3ToolErrorReplies(), prompt, nil)

	res.EditContrac = runCh3Session(bin, "editcontract", "anthropic",
		ch3EditContractReplies(), prompt, []string{"contract.txt"})

	// --- phase 3: render the ordering fixture, to see block ORDER ----------
	if work, err := os.MkdirTemp("", "ch03-render-"); err == nil {
		defer os.RemoveAll(work)
		ordering := filepath.Join(work, "ordering.log")
		if err := os.WriteFile(ordering, []byte(Ch3OrderingLog), 0o644); err == nil {
			env := ch3Env("anthropic", "", work, filepath.Join(work, "unused.log"))
			out, errOut, _ := runOnce(bin, work, env, "render", ordering)
			res.ExhibitRender = out
			if strings.TrimSpace(out) == "" {
				res.ExhibitErr = errOut
			}
		}
	}

	return res, nil
}

// exhibitResultFirst reports whether the rendered ordering fixture put the
// tool_result first in the content array of the message that answers it.
//
// Returns found=false when the render produced nothing usable, so a check can
// tell "wrong order" apart from "never rendered" instead of failing silently.
func exhibitResultFirst(res *Ch3Result) (first, found bool) {
	var body map[string]any
	if json.Unmarshal([]byte(res.ExhibitRender), &body) != nil {
		return false, false
	}
	for _, r := range anthropicWireResults(body, 0) {
		if r.CallID == "toolu_order_1" {
			return r.First, true
		}
	}
	return false, false
}

// --- reading the wire ------------------------------------------------------

// wireResults extracts every tool_result the student sent back to the model,
// in every vendor's dialect.
//
// This reads the REQUEST bodies rather than the student's own log on purpose.
// A result recorded locally but never transmitted leaves the model waiting for
// an answer to a question it can see it asked.
func wireResults(s *Ch3Session) []WireResult {
	var out []WireResult
	for i, rec := range s.Requests {
		body := rec.JSON()
		if body == nil {
			continue
		}
		switch rec.Vendor {
		case "anthropic":
			out = append(out, anthropicWireResults(body, i)...)
		case "openai":
			out = append(out, openAIWireResults(body, i)...)
		case "gemini":
			out = append(out, geminiWireResults(body, i)...)
		}
	}
	return out
}

func blockText(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		var b strings.Builder
		for _, e := range t {
			if m, ok := e.(map[string]any); ok {
				if s, ok := m["text"].(string); ok {
					b.WriteString(s)
				}
			}
		}
		return b.String()
	case map[string]any:
		b, _ := json.Marshal(t)
		return string(b)
	}
	return ""
}

func anthropicWireResults(body map[string]any, reqIdx int) []WireResult {
	var out []WireResult
	msgs, _ := body["messages"].([]any)
	for _, m := range msgs {
		mm, ok := m.(map[string]any)
		if !ok {
			continue
		}
		content, _ := mm["content"].([]any)
		for bi, b := range content {
			bm, ok := b.(map[string]any)
			if !ok || bm["type"] != "tool_result" {
				continue
			}
			id, _ := bm["tool_use_id"].(string)
			isErr, _ := bm["is_error"].(bool)
			out = append(out, WireResult{
				CallID: id, Text: blockText(bm["content"]),
				IsError: isErr, First: bi == 0, Request: reqIdx,
			})
		}
	}
	return out
}

func openAIWireResults(body map[string]any, reqIdx int) []WireResult {
	var out []WireResult
	msgs, _ := body["messages"].([]any)
	for _, m := range msgs {
		mm, ok := m.(map[string]any)
		if !ok || mm["role"] != "tool" {
			continue
		}
		id, _ := mm["tool_call_id"].(string)
		text := blockText(mm["content"])
		out = append(out, WireResult{CallID: id, Text: text, First: true, Request: reqIdx})
	}
	return out
}

func geminiWireResults(body map[string]any, reqIdx int) []WireResult {
	var out []WireResult
	contents, _ := body["contents"].([]any)
	for _, c := range contents {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		parts, _ := cm["parts"].([]any)
		for pi, p := range parts {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			fr, ok := pm["functionResponse"].(map[string]any)
			if !ok {
				continue
			}
			id, _ := fr["id"].(string)
			text := ""
			if resp, ok := fr["response"]; ok {
				text = blockText(resp)
			}
			out = append(out, WireResult{CallID: id, Text: text, First: pi == 0, Request: reqIdx})
		}
	}
	return out
}

// resultFor returns the wire result answering one call id.
func resultFor(s *Ch3Session, callID string) (WireResult, bool) {
	for _, r := range wireResults(s) {
		if r.CallID == callID {
			return r, true
		}
	}
	return WireResult{}, false
}

// logToolReturns returns the ToolReturned events the student's own log
// recorded, keyed by call id. Keys are normalized by parseLogLines, so a
// student's spelling of call_id / callID does not matter.
func logToolReturns(s *Ch3Session) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, l := range s.Log {
		if normName(l.Type) != "toolreturned" {
			continue
		}
		tool, ok := l.Data["tool"].(map[string]any)
		if !ok {
			continue
		}
		if id, _ := tool["callid"].(string); id != "" {
			out[id] = tool
		}
	}
	return out
}

// logToolCalls returns the call ids the log shows the agent DISPATCHING, as
// opposed to the ids the model merely asked for.
func logToolCalls(s *Ch3Session) map[string]bool {
	out := map[string]bool{}
	for _, l := range s.Log {
		t := normName(l.Type)
		if t != "toolcalled" && t != "toolreturned" {
			continue
		}
		if tool, ok := l.Data["tool"].(map[string]any); ok {
			if id, _ := tool["callid"].(string); id != "" {
				out[id] = true
			}
		}
	}
	return out
}

// exitCodeRe accepts any reasonable rendering of an exit status, because the
// check grades the BEHAVIOR of reporting it, not this solution's formatting.
var exitCodeRe = regexp.MustCompile(`(?i)exit[^0-9]{0,16}7`)
