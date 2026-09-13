package grade

// Chapter 2 log fixtures and normalization.
//
// NEVER GRADE ON GO IDENTIFIERS. Event type names and field names are compared
// lowercased with punctuation stripped, so ToolCalled, tool_called and
// TOOL-CALLED are the same name, and CallID, call_id and callid are the same
// field. Failing someone for Kind instead of Type is not a lesson.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExhibitLog is the log the seam is graded on. One conversation containing
// every shape the chapter's exhibits need:
//
//   - a human message
//   - an assistant turn with opaque replay material, text, and a tool call
//   - a tool result (Exhibit A: three authorships)
//   - a following human message (Exhibit B: merged by Anthropic, separate
//     everywhere else)
//
// The opaque block is tagged with the SAME model the grader renders for, so a
// correct Anthropic render replays it and a correct OpenAI or Gemini render
// omits it — material goes back only to the exact model that issued it.
const ExhibitLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"read config.json"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"opaque","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"data":{"type":"thinking","thinking":"The user wants the config file.","signature":"sig-exhibit-1"}},{"type":"text","text":"I'll read it."},{"type":"tool_call","call_id":"toolu_exhibit_1","from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"},"name":"read_file","args":{"path":"config.json","limit":40}}],"usage":{"input":100,"cache_write":0,"cache_read":50,"output":20},"from":{"vendor":"anthropic","model":"claude-sonnet-5-course","surface":"messages"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"toolu_exhibit_1","name":"read_file","args":{"path":"config.json","limit":40}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"toolu_exhibit_1","parts":[{"type":"text","text":"port=8080\nhost=localhost"}]}}
{"seq":5,"type":"message_received","time":"2026-01-01T00:00:04Z","message":{"actor":"human","parts":[{"type":"text","text":"now check the logs instead"}]}}
`

// RedactionLog is ExhibitLog plus a Redacted event superseding the tool
// result. Chapter 2 exercises only RedactResult: the RESULT becomes a stub and
// the CALL survives, so the model can still see what it asked for and why.
const RedactionLog = ExhibitLog + `{"seq":6,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}
`

// RedactedSecret is the content that must NOT survive a redaction.
const RedactedSecret = "port=8080"

// --- Ref fixtures ----------------------------------------------------------
//
// A blob no longer carries a path. It carries a Ref: a Kind saying WHAT the
// locator is, and the locator itself. These fixtures pin the three kinds, the
// two ways a Ref can be malformed, and the one kind a vendor can actually
// fetch.

// RefLocatorPath is a file on the machine running the agent.
const RefLocatorPath = "/var/agent/out/build-1.txt"

// RefLocatorURI is a Gemini File API uri — one of the three remote forms a
// local path could never express.
const RefLocatorURI = "https://generativelanguage.googleapis.com/v1beta/files/ch2exhibit"

// RefLocatorHandle is framework-managed output. It is deliberately not a
// filesystem path: the jobs chapter allows an in-memory buffer.
const RefLocatorHandle = "handle:job-7/stdout"

// RefRoundTripLog carries one blob of EACH kind. It is never rendered, only
// loaded and re-emitted, so that the three kinds are graded on surviving the
// log rather than on any vendor's opinion of them.
const RefRoundTripLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"here are three attachments"},{"type":"blob","mime":"text/plain","ref":{"kind":1,"locator":"` + RefLocatorPath + `"}},{"type":"blob","mime":"image/png","ref":{"kind":2,"locator":"` + RefLocatorURI + `"}},{"type":"blob","mime":"application/json","ref":{"kind":3,"locator":"` + RefLocatorHandle + `"}}]}}
`

// RefURILog carries the one kind a vendor can fetch for itself.
const RefURILog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"describe this image"},{"type":"blob","mime":"image/png","ref":{"kind":2,"locator":"` + RefLocatorURI + `"}}]}}
`

// RefOldFormatLog is a log written BEFORE the Ref type: the blob carries the
// old "path" spelling and no ref at all. It must be refused, loudly. Coercing
// it to RefPath would silently downgrade every remote reference in a real file
// to a local filename that never existed.
const RefOldFormatLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"describe this image"},{"type":"blob","mime":"image/png","path":"` + RefLocatorPath + `"}]}}
`

// RefZeroKindLog has a well-formed ref whose kind is the ZERO VALUE. The
// constants start at iota+1 precisely so this cannot be mistaken for RefPath.
const RefZeroKindLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"describe this image"},{"type":"blob","mime":"image/png","ref":{"kind":0,"locator":"` + RefLocatorPath + `"}}]}}
`

// RefRedactionSecret is the tool output that must NOT survive the redaction in
// RefRedactionLog.
const RefRedactionSecret = "BUILD_SECRET_TOKEN=swordfish"

// RefRedactionLocator is where the superseded content still is. It must
// survive, because the Ref is carried forward into the stub.
const RefRedactionLocator = "https://generativelanguage.googleapis.com/v1beta/files/ch2buildlog"

const refRedactionPrefix = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"build the project"}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"text","text":"Building."},{"type":"tool_call","call_id":"fc_ref_1","from":{"vendor":"gemini","model":"gemini-3.5-flash-course","surface":"generatecontent"},"name":"run_build","args":{}}],"usage":{"input":10,"cache_write":0,"cache_read":0,"output":5},"from":{"vendor":"gemini","model":"gemini-3.5-flash-course","surface":"generatecontent"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"fc_ref_1","name":"run_build","args":{}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"fc_ref_1","parts":[{"type":"text","text":"` + RefRedactionSecret + `"},{"type":"blob","mime":"text/plain","ref":{"kind":2,"locator":"` + RefRedactionLocator + `"}}]}}
`

// RefRedactionPlainLog is the NEGATIVE CONTROL: the same log with no Redacted
// event. The secret must be present here, or the redaction check would pass
// for the wrong reason on a submission that simply never renders tool results.
const RefRedactionPlainLog = refRedactionPrefix

// RefRedactionLog supersedes the tool result. The stub replaces the bytes; the
// Ref carried forward from the superseded BlobPart says where they still are.
const RefRedactionLog = refRedactionPrefix + `{"seq":5,"type":"redacted","time":"2026-01-01T00:00:05Z","redact":{"from":4,"to":4,"level":"redact_result","reason":"compaction"}}
`

// Ch2LogLine is one dumped event, with keys normalized.
type Ch2LogLine struct {
	Seq  int
	Type string
	Data map[string]any
}

func parseLogLines(dump string) ([]Ch2LogLine, string) {
	var out []Ch2LogLine
	for i, line := range strings.Split(dump, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			return out, fmt.Sprintf("line %d of dump is not JSON: %v", i+1, err)
		}
		m := normKeys(raw).(map[string]any)
		// A header line carrying only a format version is not an event.
		if m["type"] == nil && m["seq"] == nil {
			continue
		}
		l := Ch2LogLine{Data: m}
		if s, ok := m["type"].(string); ok {
			l.Type = normName(s)
		}
		if f, ok := m["seq"].(float64); ok {
			l.Seq = int(f)
		}
		out = append(out, l)
	}
	return out, ""
}

// normName lowercases and strips everything that is not a letter or digit.
func normName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normKeys rewrites every map key in a decoded JSON tree to its normalized
// form, recursively, so lookups never depend on a student's spelling.
func normKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[normName(k)] = normKeys(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = normKeys(val)
		}
		return out
	default:
		return v
	}
}

// get looks up the first present of several normalized field names.
func get(m map[string]any, names ...string) any {
	for _, n := range names {
		if v, ok := m[normName(n)]; ok {
			return v
		}
	}
	return nil
}

func getMap(m map[string]any, names ...string) map[string]any {
	if v, ok := get(m, names...).(map[string]any); ok {
		return v
	}
	return nil
}

func getSlice(m map[string]any, names ...string) []any {
	if v, ok := get(m, names...).([]any); ok {
		return v
	}
	return nil
}

func getStr(m map[string]any, names ...string) string {
	if v, ok := get(m, names...).(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]any, names ...string) (int, bool) {
	if v, ok := get(m, names...).(float64); ok {
		return int(v), true
	}
	return 0, false
}

// canonicalJSON re-serializes a decoded value with map keys sorted, so that
// `{"a":1,"b":2}` and `{"b":2,"a":1}` compare equal, and so that OpenAI's
// JSON-encoded argument STRING compares equal to Anthropic's nested object
// once the student has decoded it.
func canonicalJSON(v any) string {
	if s, ok := v.(string); ok {
		var inner any
		if json.Unmarshal([]byte(s), &inner) == nil {
			v = inner
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// SaidProjection reduces a dumped log to "everything the model said", with the
// things that legitimately differ between vendors removed:
//
//   - provenance (checked separately, and required to survive)
//   - tool-call ids, which are issued by the vendor and cannot match
//   - opaque replay material, which is bound to one model by definition
//   - token accounting, which is graded by `usage`
//
// Usage is deliberately NOT compared here. It is parsing work, and it is real
// parsing work — but it has its own check so that a student who gets the
// message shapes right and the accounting wrong is told which half failed,
// instead of losing two large blocks to one cause.
//
// What remains must be identical across all three vendors. If it is not, the
// student has leaked vendor shape past the parser — and leaked vendor shape is
// precisely what makes the second implementation a copy-paste.
func SaidProjection(lines []Ch2LogLine) []string {
	var out []string
	for _, l := range lines {
		switch l.Type {
		case normName("response_ended"):
			resp := getMap(l.Data, "response")
			if resp == nil {
				out = append(out, "response_ended:<no payload>")
				continue
			}
			out = append(out, "response:"+partsProjection(getSlice(resp, "parts")))
		case normName("message_received"):
			msg := getMap(l.Data, "message")
			if msg == nil {
				continue
			}
			out = append(out, "message:"+getStr(msg, "actor")+":"+partsProjection(getSlice(msg, "parts")))
		}
	}
	return out
}

func partsProjection(parts []any) string {
	var bits []string
	for _, p := range parts {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		switch normName(getStr(m, "type", "kind")) {
		case normName("tool_call"):
			// Name and arguments must normalize. The ID must not: it is
			// vendor-issued. Its PRESENCE is checked elsewhere.
			bits = append(bits, "call("+getStr(m, "name")+","+canonicalJSON(get(m, "args", "arguments", "input"))+")")
		case normName("tool_result"):
			bits = append(bits, "result")
		case normName("opaque"):
			// Excluded: bound to one model by definition.
		case normName("redacted"):
			bits = append(bits, "redacted")
		default:
			if t := getStr(m, "text"); t != "" {
				bits = append(bits, "text("+t+")")
			}
		}
	}
	return strings.Join(bits, "|")
}

func usageProjection(u map[string]any) string {
	if u == nil {
		return "<none>"
	}
	f := func(names ...string) string {
		if n, ok := getInt(u, names...); ok {
			return fmt.Sprint(n)
		}
		return "?"
	}
	return strings.Join([]string{
		f("input"), f("cache_write"), f("cache_read"), f("output"),
	}, "/")
}

// GeminiReplayLog exercises ToolCallPart.Opaque, which ExhibitLog cannot.
//
// Two properties are required at once and ExhibitLog has neither:
//
//  1. the tool call must CARRY opaque material (ExhibitLog's does not), and
//  2. it must have been produced by the model we render back to, or a correct
//     renderer withholds it by design and the assertion passes vacuously.
//
// So the provenance here is Gemini, and the model matches ch2RequestedModel
// ("gemini") exactly — vendor, model AND surface, since all three participate
// in the same-model test. Replaying a functionCall to Gemini 3.x WITHOUT its
// signature is a 400, which is the entire reason §2.4a grew the field.
const GeminiReplayLog = `{"log_version":1}
{"seq":1,"type":"message_received","time":"2026-01-01T00:00:00Z","message":{"actor":"human","parts":[{"type":"text","text":"Check the deploy script."}]}}
{"seq":2,"type":"response_ended","time":"2026-01-01T00:00:01Z","response":{"parts":[{"type":"text","text":"Reading it now."},{"type":"tool_call","call_id":"gemini_call_1","from":{"vendor":"gemini","model":"gemini-3.5-flash-course","surface":"generate_content"},"name":"read_file","args":{"path":"deploy.sh"},"opaque":"sig-bound-to-this-call"}],"usage":{"input":40,"cache_write":0,"cache_read":0,"output":12},"from":{"vendor":"gemini","model":"gemini-3.5-flash-course","surface":"generate_content"}}}
{"seq":3,"type":"tool_called","time":"2026-01-01T00:00:02Z","tool":{"call_id":"gemini_call_1","name":"read_file","args":{"path":"deploy.sh"}}}
{"seq":4,"type":"tool_returned","time":"2026-01-01T00:00:03Z","tool":{"call_id":"gemini_call_1","parts":[{"type":"text","text":"#!/bin/sh\nexec ./serve"}]}}
`
