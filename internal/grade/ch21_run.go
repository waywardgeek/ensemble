package grade

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Ch21Run exercises the student's agent twice: once on the happy path, once
// on the two ways a web tool fails.
//
// Nothing here reads the student's source. Every check below is answered from
// what the fake vendor received and what the fake MCP server was asked to do.
func Ch21Run(dir string) Ch21Result {
	var res Ch21Result

	bin, cleanup, err := Build(dir)
	if err != nil {
		res.Fatal = fmt.Sprintf("build failed: %v", err)
		return res
	}
	defer cleanup()

	gui := filepath.Join(dir, "web")
	nonce := ch21Nonce()

	ch21ScenarioResearch(&res, bin, gui, nonce)
	ch21ScenarioFailures(&res, bin, gui, nonce)
	return res
}

// ch21Nonce is a fresh token per run. It is the unfakeable half of the fetch
// check: a student who stubs the tool, caches a canned page, or answers from
// the model's own knowledge cannot produce a value that did not exist until
// this process started.
func ch21Nonce() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "ch21-fallback-token"
	}
	return "EN21-" + hex.EncodeToString(b)
}

// ch21ScenarioResearch is the path the chapter is about: load the skill,
// search, read one result.
func ch21ScenarioResearch(res *Ch21Result, bin, gui, nonce string) {
	out := ch21Launch(bin, gui, nonce, ch21Opts{
		prompts: []string{"Find the ensemble guide and tell me the build token."},
		replies: []fakevendor.Reply{
			{ToolName: "load_skill", ToolArgs: `{"name":"web-search"}`, ToolID: "t1"},
			{ToolName: "firecrawl_search", ToolArgs: `{"query":"ensemble guide"}`, ToolID: "t2"},
			{ToolName: "firecrawl_scrape", ToolArgs: `{"url":"` + ch21DocURL + `"}`, ToolID: "t3"},
			{Text: "Done."},
		},
	})
	if out.fatal != "" {
		res.Fatal = out.fatal
		return
	}

	// --- tools-gated-by-skill -------------------------------------------
	// The first request is sent before load_skill has run, so the web tools
	// must not be offered yet. "Off" in this chapter means this feature
	// issues no searches; the honest way to show that is that the tools are
	// absent until the skill is loaded.
	res.ran("tools-gated-by-skill")
	if len(out.requests) == 0 {
		res.fail("tools-gated-by-skill", "the vendor received no requests at all")
	} else {
		first := ch21ToolNames(out.requests[0])
		if ch21Has(first, "firecrawl_search") || ch21Has(first, "firecrawl_scrape") {
			res.fail("tools-gated-by-skill",
				"web tools were offered before the web-search skill was loaded: %v", first)
		}
	}

	// --- url-transport-connects -----------------------------------------
	// After load_skill, the server's tools must appear in the tool list the
	// agent sends. Reaching them at all requires the URL transport: the skill
	// declares transport: url and nothing else.
	res.ran("url-transport-connects")
	var sawTools bool
	for _, r := range out.requests[1:] {
		names := ch21ToolNames(r)
		if ch21Has(names, "firecrawl_search") && ch21Has(names, "firecrawl_scrape") {
			sawTools = true
			break
		}
	}
	if !sawTools {
		res.fail("url-transport-connects",
			"after loading web-search, neither firecrawl_search nor firecrawl_scrape was offered to the model; "+
				"the MCP server over transport: url was not reached (tools last seen: %v)",
			ch21ToolNames(out.requests[len(out.requests)-1]))
	}

	// --- search-dispatches ----------------------------------------------
	res.ran("search-dispatches")
	if !ch21Called(out.calls, "firecrawl_search") {
		res.fail("search-dispatches", "the MCP server never received a firecrawl_search call")
	} else if !ch21AnyToolResultContains(out.requests, ch21DocURL) {
		res.fail("search-dispatches",
			"firecrawl_search ran but its results never reached the model as a tool result")
	}

	// --- fetch-returns-planted-token -------------------------------------
	res.ran("fetch-returns-planted-token")
	if !ch21Called(out.calls, "firecrawl_scrape") {
		res.fail("fetch-returns-planted-token", "the MCP server never received a firecrawl_scrape call")
	} else if !ch21AnyToolResultContains(out.requests, nonce) {
		res.fail("fetch-returns-planted-token",
			"the fetched page's planted token %q never reached the model as a tool result; "+
				"the page content was dropped, truncated away, or never really fetched", nonce)
	}
}

// ch21ScenarioFailures covers the untrusted-content boundary and the two
// distinct ways a web tool fails.
func ch21ScenarioFailures(res *Ch21Result, bin, gui, nonce string) {
	out := ch21Launch(bin, gui, nonce, ch21Opts{
		prompts: []string{"Read the blog post, then the missing page, then the broken one."},
		replies: []fakevendor.Reply{
			{ToolName: "load_skill", ToolArgs: `{"name":"web-search"}`, ToolID: "t1"},
			{ToolName: "firecrawl_scrape", ToolArgs: `{"url":"` + ch21InjectURL + `"}`, ToolID: "t2"},
			{ToolName: "firecrawl_scrape", ToolArgs: `{"url":"` + ch21MissingURL + `"}`, ToolID: "t3"},
			{ToolName: "firecrawl_scrape", ToolArgs: `{"url":"` + ch21BoomURL + `"}`, ToolID: "t4"},
			{Text: "Done."},
		},
	})
	if out.fatal != "" {
		if res.Fatal == "" {
			res.Fatal = out.fatal
		}
		return
	}

	// --- fetched-content-is-a-tool-result --------------------------------
	// The page carries an instruction aimed at the model. The check does NOT
	// claim the agent "resisted" it — the model is scripted here, so that
	// would be free. What it asserts is the structural property that makes
	// resistance possible at all: the bytes arrived as a tool result and
	// nowhere else. Merged into the user turn or the system prompt, fetched
	// text becomes indistinguishable from the operator's own words.
	res.ran("fetched-content-is-a-tool-result")
	if !ch21AnyToolResultContains(out.requests, ch21Injection) {
		res.fail("fetched-content-is-a-tool-result",
			"the fetched page never reached the model as a tool result")
	} else if where := ch21NonToolResultContaining(out.requests, ch21Injection); where != "" {
		res.fail("fetched-content-is-a-tool-result",
			"fetched page text also appeared outside a tool result (in %s); "+
				"web content must stay data, not become instructions", where)
	}

	// --- tool-error-is-reported ------------------------------------------
	// MCP-level failure: HTTP 200 with isError true.
	res.ran("tool-error-is-reported")
	if !ch21ErroredToolResultContains(out.requests, "404") {
		if ch21AnyToolResultContains(out.requests, "404") {
			res.fail("tool-error-is-reported",
				"the failed scrape reached the model as a SUCCESSFUL tool result; the server's "+
					"isError flag was dropped, so the model sees an error page framed as content")
		} else {
			res.fail("tool-error-is-reported",
				"a tool error (isError) never reached the model; a failure the model cannot see "+
					"is a failure it will confidently narrate around")
		}
	}

	// --- transport-error-is-reported --------------------------------------
	// HTTP-level failure: 500 with a body naming the reason. A transport that
	// drops the body turns a diagnosable refusal — quota, auth — into
	// "request failed", and the agent reports a mystery to the user.
	res.ran("transport-error-is-reported")
	if !ch21Called(out.calls, "firecrawl_scrape") {
		res.fail("transport-error-is-reported", "no scrape reached the server at all")
	} else {
		reached := ch21ErroredToolResultContains(out.requests, "quota exhausted")
		if !reached {
			detail := "the failure never reached the model as an error at all"
			if ch21AnyToolResultContains(out.requests, "quota exhausted") {
				detail = "the reason reached the model but was not marked as an error"
			} else if ch21ErroredToolResultContains(out.requests, "") {
				detail = "an error reached the model, but the reason in the response body was discarded"
			}
			res.fail("transport-error-is-reported",
				"the HTTP 500 from the MCP endpoint did not reach the model with a usable reason: %s; "+
					"the response body is where a refusal says whether it was quota, auth or outage", detail)
		}
	}

	// Note: an agent that died on the failing tools cannot reach here at all —
	// ch21Launch reports a fatal and this scenario returns early — so
	// "survives a tool failure" needs no separate check. Keeping the check
	// budget at seven is a published contract, and a check that can only fire
	// when another has already aborted the run would spend part of it on
	// nothing.
}

// ---- observation helpers -------------------------------------------------
//
// These read the request bodies the fake vendor recorded. The vendor is
// Anthropic here, so the shapes below are Anthropic's.

func ch21Has(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func ch21Called(calls []ch21Call, name string) bool {
	for _, c := range calls {
		if c.Name == name {
			return true
		}
	}
	return false
}

// ch21ToolNames returns the names in the request's tools array.
func ch21ToolNames(r fakevendor.Recorded) []string {
	var out []string
	body := r.JSON()
	if body == nil {
		return out
	}
	tools, _ := body["tools"].([]any)
	for _, t := range tools {
		m, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if n, ok := m["name"].(string); ok {
			out = append(out, n)
		}
	}
	return out
}

// ch21AnyToolResultContains reports whether needle appears inside any
// tool_result block of any recorded request.
func ch21AnyToolResultContains(rs []fakevendor.Recorded, needle string) bool {
	for _, r := range rs {
		for _, tr := range ch21ToolResults(r) {
			if strings.Contains(tr.Text, needle) {
				return true
			}
		}
	}
	return false
}

// ch21ErroredToolResultContains is the same question, restricted to results
// the agent marked as failures.
//
// Asserting the text alone is not enough, and the difference is the whole
// check: if the agent dropped the server's isError flag and passed the body
// through as an ordinary success, the words "404 Not Found" would still reach
// the model — attached to a result that claims it worked. The model then has
// a page that says one thing and a frame that says another, and it will
// usually believe the frame.
func ch21ErroredToolResultContains(rs []fakevendor.Recorded, needle string) bool {
	for _, r := range rs {
		for _, tr := range ch21ToolResults(r) {
			if tr.IsError && strings.Contains(tr.Text, needle) {
				return true
			}
		}
	}
	return false
}

// ch21NonToolResultContaining returns a description of the first place
// OUTSIDE a tool result where needle appears, or "" if there is none.
func ch21NonToolResultContaining(rs []fakevendor.Recorded, needle string) string {
	for i, r := range rs {
		body := r.JSON()
		if body == nil {
			continue
		}
		if s := ch21SystemText(body); strings.Contains(s, needle) {
			return fmt.Sprintf("the system prompt of request %d", i+1)
		}
		msgs, _ := body["messages"].([]any)
		for j, mv := range msgs {
			m, ok := mv.(map[string]any)
			if !ok {
				continue
			}
			role, _ := m["role"].(string)
			if s, ok := m["content"].(string); ok {
				if strings.Contains(s, needle) {
					return fmt.Sprintf("the %s message at index %d of request %d", role, j, i+1)
				}
				continue
			}
			blocks, _ := m["content"].([]any)
			for _, bv := range blocks {
				b, ok := bv.(map[string]any)
				if !ok {
					continue
				}
				if b["type"] == "tool_result" {
					continue
				}
				if s, ok := b["text"].(string); ok && strings.Contains(s, needle) {
					return fmt.Sprintf("a %v block in the %s message of request %d", b["type"], role, i+1)
				}
			}
		}
	}
	return ""
}

func ch21SystemText(body map[string]any) string {
	switch v := body["system"].(type) {
	case string:
		return v
	case []any:
		var sb strings.Builder
		for _, bv := range v {
			if b, ok := bv.(map[string]any); ok {
				if s, ok := b["text"].(string); ok {
					sb.WriteString(s)
					sb.WriteString("\n")
				}
			}
		}
		return sb.String()
	}
	return ""
}

// ch21ToolResult is one tool_result block as it appeared on the wire.
type ch21ToolResult struct {
	Text    string
	IsError bool
}

// ch21ToolResults pulls every tool_result block out of a request, keeping the
// is_error flag alongside the text.
func ch21ToolResults(r fakevendor.Recorded) []ch21ToolResult {
	var out []ch21ToolResult
	body := r.JSON()
	if body == nil {
		return out
	}
	msgs, _ := body["messages"].([]any)
	for _, mv := range msgs {
		m, ok := mv.(map[string]any)
		if !ok {
			continue
		}
		blocks, _ := m["content"].([]any)
		for _, bv := range blocks {
			b, ok := bv.(map[string]any)
			if !ok || b["type"] != "tool_result" {
				continue
			}
			isErr, _ := b["is_error"].(bool)
			out = append(out, ch21ToolResult{Text: ch21BlockText(b["content"]), IsError: isErr})
		}
	}
	return out
}

// ch21BlockText flattens the two shapes a tool_result's content can take.
func ch21BlockText(v any) string {
	switch c := v.(type) {
	case string:
		return c
	case []any:
		var sb strings.Builder
		for _, iv := range c {
			if im, ok := iv.(map[string]any); ok {
				if s, ok := im["text"].(string); ok {
					sb.WriteString(s)
					sb.WriteString("\n")
				}
			}
		}
		return sb.String()
	case map[string]any:
		if s, ok := c["text"].(string); ok {
			return s
		}
	}
	if b, err := json.Marshal(v); err == nil {
		return string(b)
	}
	return ""
}
