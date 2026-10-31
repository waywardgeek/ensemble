package gradeedition2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter2 preserves the original three-vendor exercise and corrects its
// documented scope and sensitivity defects. The CLI cannot expose the reduced
// Ref value; exact reducer carryforward remains a required source-review check.
func Chapter2(bin string) (grade.Report, error) {
	res, err := grade.Ch2Run(bin)
	if err != nil {
		return grade.Report{}, err
	}
	// Ch2Run retains only the original Chapter 1 verdict, not its HTTP headers.
	// A short additional run carries Edition 2's exact-version correction forward.
	parity, err := grade.Run(bin)
	if err != nil {
		return grade.Report{}, err
	}
	res.Ch1 = Chapter1(parity).Checks
	work, err := os.MkdirTemp("", "edition2-ch02-*")
	if err != nil {
		return grade.Report{}, err
	}
	defer os.RemoveAll(work)

	checks := []grade.Check{}
	for _, check := range grade.Ch2Evaluate(res) {
		switch check.ID {
		case "ref-render":
			// URI media rendering is explicitly deferred by this chapter.
			continue
		case "ref-redaction":
			// Keep the total at 100 by assigning the deferred media's four
			// points to the current chapter's observable preservation work.
			check = refRedaction(bin, work)
		case "seam-render":
			checkResultName(&check, res.Render["gemini"])
			body, err := renderFixture(bin, work, "anthropic", interleavedLog)
			if err != nil {
				fail(&check, "interleaved result render: %v", err)
			} else {
				checkInterleaved(&check, body)
			}
		case "seam-parse":
			for _, vendor := range grade.Ch2Vendors {
				if session := res.Session[vendor]; session != nil {
					checkCapture(&check, vendor, session.Log)
				}
			}
			// The inherited sessions supply a known resolved model. This one
			// deliberately does not, so capture must read the returned identity.
			checkUnresolvedCapture(&check, bin, work)
		}
		checks = append(checks, check)
	}
	return grade.NewTitledReport("Edition 2, Chapter 2 — One Log, Three Vendors", checks, res.Stderr), nil
}

// fail changes only the selected verdict and never leaves an obsolete success
// note alongside a newly discovered failure.
func fail(check *grade.Check, format string, args ...any) {
	if check.Passed {
		check.Details = nil
	}
	check.Passed, check.Earned = false, 0
	check.Details = append(check.Details, fmt.Sprintf(format, args...))
}

// runChapter2 uses only the already published CLI and environment contract.
// Each launch has a bounded lifetime and a private directory, including dump.
func runChapter2(bin, work, vendor, endpoint, input string, args ...string) (string, error) {
	return runModel(bin, work, vendor, fakevendor.Models[vendor], endpoint, input, args...)
}

// Media fixtures select a known model explicitly; the earlier chapter fixtures
// keep their original fake identities through runChapter2 above.
func runModel(bin, work, vendor, model, endpoint, input string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), grade.Ch2Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = work
	cmd.Env = append(os.Environ(),
		"LLM_VENDOR="+vendor, "LLM_MODEL="+model,
		"LLM_RESOLVED_MODEL=", "LLM_API_KEY=course-grader-fake",
		"LLM_BASE_URL="+endpoint, "CH02_LOG="+filepath.Join(work, "input.log"))
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr strings.Builder
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func renderFixture(bin, work, vendor, log string) (string, error) {
	path := filepath.Join(work, "input.log")
	if err := os.WriteFile(path, []byte(log), 0600); err != nil {
		return "", err
	}
	return runChapter2(bin, work, vendor, "http://127.0.0.1:1", "", "render", path)
}

// The second result arrives after human text. Arrival order belongs to the
// log; results-first order belongs only to the Anthropic request projection.
const interleavedLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:01Z","message":{"actor":"human","parts":[{"type":"text","text":"read both"}]}}
{"type":"request_sent","request":{"to":{"vendor":"anthropic","model":"claude-sonnet-5-fake","surface":"messages"}},"seq":2,"time":"2026-01-01T00:00:02Z"}
{"seq":3,"type":"response_ended","time":"2026-01-01T00:00:03Z","response":{"from":{"vendor":"anthropic","model":"claude-sonnet-5-fake","surface":"messages"},"usage":{"input":1,"cache_write":0,"cache_read":0,"output":1},"parts":[{"type":"tool_call","call_id":"one","from":{"vendor":"anthropic","model":"claude-sonnet-5-fake","surface":"messages"},"name":"read_file","args":{"path":"a"}},{"type":"tool_call","call_id":"two","from":{"vendor":"anthropic","model":"claude-sonnet-5-fake","surface":"messages"},"name":"read_file","args":{"path":"b"}}]}}
{"seq":4,"type":"tool_called","time":"2026-01-01T00:00:04Z","tool":{"call_id":"one","name":"read_file","args":{"path":"a"}}}
{"seq":5,"type":"tool_called","time":"2026-01-01T00:00:05Z","tool":{"call_id":"two","name":"read_file","args":{"path":"b"}}}
{"seq":6,"type":"tool_returned","time":"2026-01-01T00:00:06Z","tool":{"call_id":"one","parts":[{"type":"text","text":"first result"}]}}
{"seq":7,"type":"message_received","time":"2026-01-01T00:00:07Z","message":{"actor":"human","parts":[{"type":"text","text":"keep this instruction"}]}}
{"seq":8,"type":"tool_returned","time":"2026-01-01T00:00:08Z","tool":{"call_id":"two","parts":[{"type":"text","text":"second result"}]}}
`

func checkInterleaved(check *grade.Check, body string) {
	var request struct {
		Messages []struct {
			Role string
			// Anthropic permits either text or a block array here.
			Content json.RawMessage
		}
	}
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		fail(check, "interleaved Anthropic request is invalid JSON: %v", err)
		return
	}
	for _, message := range request.Messages {
		if message.Role != "user" {
			continue
		}
		var text string
		if json.Unmarshal(message.Content, &text) == nil {
			continue // Ordinary text cannot be the result-bearing message.
		}
		var parts []struct {
			Type      string
			ToolUseID string `json:"tool_use_id"`
			Text      string
			Content   json.RawMessage
		}
		if err := json.Unmarshal(message.Content, &parts); err != nil {
			fail(check, "interleaved Anthropic user content is invalid: %v", err)
			return
		}
		if len(parts) != 3 {
			continue
		}
		if parts[0].Type == "tool_result" && parts[0].ToolUseID == "one" &&
			parts[1].Type == "tool_result" && parts[1].ToolUseID == "two" &&
			parts[2].Type == "text" && parts[2].Text == "keep this instruction" &&
			strings.Contains(string(parts[0].Content), "first result") &&
			strings.Contains(string(parts[1].Content), "second result") {
			return
		}
	}
	fail(check, "interleaved Anthropic results must retain both outputs and IDs before the human text")
}

func checkResultName(check *grade.Check, body string) {
	var request struct {
		// Contents holds Gemini conversation turns.
		Contents []struct {
			// Parts carries the vendor result objects being inspected.
			Parts []struct {
				// FunctionResponse identifies a result rather than a call or text.
				FunctionResponse *struct {
					// Name must be recovered from the matching call.
					Name string
				}
			}
		}
	}
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		fail(check, "Gemini result-name request is invalid JSON: %v", err)
		return
	}
	for _, turn := range request.Contents {
		for _, part := range turn.Parts {
			if result := part.FunctionResponse; result != nil {
				if result.Name != "read_file" {
					fail(check, "Gemini functionResponse.name = %q, expected read_file from the matching call", result.Name)
				}
				return
			}
		}
	}
	fail(check, "Gemini result-name probe found no functionResponse")
}

// The original harness already normalizes student-owned names in these logs.
// Vendor wire names above remain exact; these small accessors preserve the
// original grader's allowance for field spelling in the student's own format.
func object(value any) map[string]any { result, _ := value.(map[string]any); return result }
func list(value any) []any            { result, _ := value.([]any); return result }
func field(m map[string]any, names ...string) any {
	for _, name := range names {
		if value, ok := m[name]; ok {
			return value
		}
	}
	return nil
}

func checkCapture(check *grade.Check, vendor string, events []grade.Ch2LogLine) {
	responses, calls := 0, 0
	for _, event := range events {
		if event.Type != "responseended" {
			continue
		}
		responses++
		response := object(event.Data["response"])
		from := object(field(response, "from", "provenance"))
		if from["model"] != fakevendor.Models[vendor] {
			fail(check, "%s: captured response model = %v, expected returned identity %s", vendor, from["model"], fakevendor.Models[vendor])
		}
		for _, value := range list(response["parts"]) {
			part := object(value)
			kind, _ := field(part, "type", "kind").(string)
			if normalize(kind) != "toolcall" {
				continue
			}
			calls++
			producer := object(field(part, "from", "provenance"))
			if producer["model"] != fakevendor.Models[vendor] {
				fail(check, "%s: captured tool-call model = %v, expected returned identity %s", vendor, producer["model"], fakevendor.Models[vendor])
			}
			if vendor == "gemini" {
				id, _ := field(part, "callid", "id").(string)
				if part["opaque"] != "sig-fake-"+id || id == "" {
					fail(check, "Gemini: captured call %q lost its exact call-bound signature", id)
				}
			}
		}
	}
	if responses == 0 || calls == 0 {
		fail(check, "%s: capture probe has no completed response/tool call", vendor)
	}
}

func checkUnresolvedCapture(check *grade.Check, bin, work string) {
	fake := fakevendor.New([]fakevendor.Reply{
		{ToolName: "read_file", ToolID: "capture-call", ToolArgs: `{"path":"config.json"}`, Usage: fakevendor.Canonical{Input: 1, Output: 1}},
		// Chapter 2 only records the call; later chapters answer it. Let their
		// real loop finish without adding usage or bypassing tool dispatch.
		{Text: "Capture complete."},
	})
	defer fake.Close()
	// An unknown routing name with a local fake needs no model discovery. The
	// response supplies modelVersion; Chapter 2 records the unanswered call.
	path := filepath.Join(work, "capture")
	if err := os.Mkdir(path, 0700); err != nil {
		fail(check, "capture setup: %v", err)
		return
	}
	// Use a separate launcher here to make the routing alias different from
	// both the fake's returned identity and the offline-render configuration.
	ctx, cancel := context.WithTimeout(context.Background(), grade.Ch2Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "LLM_VENDOR=gemini", "LLM_MODEL=gemini-routing-alias", "LLM_RESOLVED_MODEL=", "LLM_API_KEY=course-grader-fake", "LLM_BASE_URL="+fake.URL(), "CH02_LOG="+filepath.Join(path, "input.log"))
	cmd.Stdin = strings.NewReader("{\"user\":\"read config\"}\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		fail(check, "capture session: %v: %s", err, out)
		return
	}
	dump, err := runChapter2(bin, path, "gemini", "http://127.0.0.1:1", "", "dump")
	if err != nil {
		fail(check, "capture dump: %v", err)
		return
	}
	events, err := readLog(dump)
	if err != nil {
		fail(check, "capture log: %v", err)
		return
	}
	checkCapture(check, "gemini", events)
}

func normalize(name string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, strings.ToLower(name))
}

// Normalize field names, not text/locators, just as the original log grader
// does. This accepts equivalent Go naming without weakening byte preservation.
func normalizeKeys(value any) any {
	switch value := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, child := range value {
			result[normalize(key)] = normalizeKeys(child)
		}
		return result
	case []any:
		for i, child := range value {
			value[i] = normalizeKeys(child)
		}
	}
	return value
}

func readLog(dump string) ([]grade.Ch2LogLine, error) {
	var events []grade.Ch2LogLine
	for _, line := range strings.Split(strings.TrimSpace(dump), "\n") {
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return nil, err
		}
		data := normalizeKeys(raw).(map[string]any)
		kind, _ := data["type"].(string)
		if kind == "" {
			continue
		}
		events = append(events, grade.Ch2LogLine{Type: normalize(kind), Data: data})
	}
	return events, nil
}
