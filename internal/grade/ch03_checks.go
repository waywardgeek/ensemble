package grade

// Chapter 3's nine checks.
//
// ch2parity 10 + toolsdecl 5 + toolloop 20 + multiblock 10 + readtools 10
// + mutatetools 10 + runcommand 15 + toolerror 15 + editcontract 5 = 100.
//
// toolsdecl's five points came out of toolloop (was 25). They grade the two
// halves of one protocol: toolsdecl the request direction, toolloop the
// response direction. Before toolsdecl existed an agent that never declared a
// tool scored 100 against the fake and was dead against every real vendor.
//
// A note on what these checks deliberately do NOT grade: formatting. A student
// may render a tool's output however they like, so the assertions are about
// observable behavior — which bytes ended up on disk, which call id a result
// was keyed to, whether a failure came back marked as one — rather than about
// this solution's particular phrasing.

import (
	"encoding/json"
	"fmt"
	"strings"
)

func Ch3Evaluate(res *Ch3Result) []Check {
	return []Check{
		checkCh2Parity(res),
		checkToolsDecl(res),
		checkToolLoop(res),
		checkMultiblock(res),
		checkReadTools(res),
		checkMutateTools(res),
		checkRunCommand(res),
		checkToolError(res),
		checkEditContract(res),
	}
}

// --- toolsdecl -------------------------------------------------------------

// requiredTools is the capability surface the chapter names and the fake
// calls by these exact names. A declaration that omits one is a tool the
// model will never ask for.
var requiredTools = []string{"read_file", "write_file", "edit_file", "list_directory", "search_files", "run_command"}

// declaredTool is one tool declaration lifted out of whichever vendor
// envelope it arrived in, so the assertions below are vendor-neutral.
type declaredTool struct {
	Name        string
	Description string
	Schema      map[string]any
}

// vendorToolDecls finds the tool declarations in a request body for the
// given vendor, in that vendor's exact wire shape. ok is false when the
// field is missing or not the shape the vendor documents; why says which.
func vendorToolDecls(vendor string, body []byte) (decls []declaredTool, ok bool, why string) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, "request body is not a JSON object"
	}
	raw, present := req["tools"]
	if !present {
		return nil, false, "no `tools` field in the request"
	}
	list, isList := raw.([]any)
	if !isList {
		return nil, false, "`tools` is not an array"
	}
	if len(list) == 0 {
		return nil, false, "`tools` is present but empty"
	}
	str := func(m map[string]any, k string) string {
		s, _ := m[k].(string)
		return s
	}
	lift := func(m map[string]any, schemaKey string) declaredTool {
		schema, _ := m[schemaKey].(map[string]any)
		return declaredTool{Name: str(m, "name"), Description: str(m, "description"), Schema: schema}
	}
	for i, item := range list {
		m, isObj := item.(map[string]any)
		if !isObj {
			return nil, false, fmt.Sprintf("tools[%d] is not an object", i)
		}
		switch vendor {
		case "anthropic":
			// {name, description, input_schema}
			decls = append(decls, lift(m, "input_schema"))
		case "openai":
			// {type:"function", function:{name, description, parameters}}
			if str(m, "type") != "function" {
				return nil, false, fmt.Sprintf("tools[%d].type is %q, want \"function\"", i, str(m, "type"))
			}
			fn, has := m["function"].(map[string]any)
			if !has {
				return nil, false, fmt.Sprintf("tools[%d] has no `function` object", i)
			}
			decls = append(decls, lift(fn, "parameters"))
		case "gemini":
			// {functionDeclarations:[{name, description, parameters}]} — the
			// snake_case spelling is also a valid proto-JSON wire name.
			fds, has := m["functionDeclarations"].([]any)
			if !has {
				fds, has = m["function_declarations"].([]any)
			}
			if !has {
				return nil, false, fmt.Sprintf("tools[%d] has no `functionDeclarations` array", i)
			}
			for j, fd := range fds {
				fm, isObj := fd.(map[string]any)
				if !isObj {
					return nil, false, fmt.Sprintf("tools[%d].functionDeclarations[%d] is not an object", i, j)
				}
				decls = append(decls, lift(fm, "parameters"))
			}
		}
	}
	return decls, true, ""
}

// checkToolsDecl: every request the agent sends declares its tools — name,
// description, argument schema — in the vendor's own shape. This is the
// request-direction half of the tool protocol; toolloop grades the response
// half. Without it a real model never emits a tool call, and the chapter's
// payoff sentence is true only against our fake, which volunteers them.
//
// Every request in the session is inspected, not only the first: Anthropic
// rejects a request whose messages contain tool_use blocks but whose `tools`
// is missing, so a declaration that disappears on the second turn kills the
// agent one turn later than never declaring at all.
//
// What the grader cannot see: that an EMPTY registry renders to no field.
// The student's registry is never empty from out here. The reference
// solution proves that property in its own test suite by byte comparison
// against chapter 2's binary (TestCh3NoDeclIsCh2Bytes).
func checkToolsDecl(res *Ch3Result) Check {
	c := Check{ID: "toolsdecl", Title: "every request declares the tools — name, description, schema — in the vendor's shape",
		Points: 5, Passed: true, Earned: 5}

	var bad []string
	for _, vendor := range Ch2Vendors {
		s := res.Loop[vendor]
		if s == nil || s.Err != "" {
			bad = append(bad, vendor+": scenario did not run")
			continue
		}
		if len(s.Requests) == 0 {
			bad = append(bad, vendor+": no requests were sent")
			continue
		}
		p := func(format string, args ...any) {
			bad = append(bad, vendor+": "+fmt.Sprintf(format, args...))
		}
		for i, rq := range s.Requests {
			decls, ok, why := vendorToolDecls(vendor, rq.Body)
			if !ok {
				p("request %d: %s", i+1, why)
				break
			}
			byName := map[string]declaredTool{}
			for _, d := range decls {
				byName[d.Name] = d
			}
			var problems []string
			for _, name := range requiredTools {
				d, declared := byName[name]
				switch {
				case !declared:
					problems = append(problems, name+" not declared")
				case strings.TrimSpace(d.Description) == "":
					problems = append(problems, name+" has no description")
				case d.Schema == nil:
					problems = append(problems, name+" has no argument schema object")
				case d.Schema["type"] != "object":
					problems = append(problems, fmt.Sprintf("%s schema type is %v, want \"object\"", name, d.Schema["type"]))
				default:
					props, _ := d.Schema["properties"].(map[string]any)
					if len(props) == 0 {
						problems = append(problems, name+" schema declares no properties — the model would have to guess the argument names")
					}
				}
			}
			if len(problems) > 0 {
				p("request %d: %s", i+1, joined(problems))
				break
			}
		}
	}
	if len(bad) > 0 {
		c.failf("%s", joined(bad))
		return c
	}
	c.Details = append(c.Details, fmt.Sprintf("all %d tools declared with description and object schema, on every request, for %s",
		len(requiredTools), strings.Join(Ch2Vendors, ", ")))
	return c
}

// --- shared readers --------------------------------------------------------

// responseParts returns the parts of each response_ended event, in order.
// This is where the type filter is observable: one response carrying both
// prose and a call must record BOTH kinds of part.
func responseParts(s *Ch3Session) [][]map[string]any {
	var out [][]map[string]any
	for _, l := range s.Log {
		if normName(l.Type) != "responseended" {
			continue
		}
		resp, ok := l.Data["response"].(map[string]any)
		if !ok {
			continue
		}
		raw, _ := resp["parts"].([]any)
		var parts []map[string]any
		for _, p := range raw {
			if pm, ok := p.(map[string]any); ok {
				parts = append(parts, pm)
			}
		}
		out = append(out, parts)
	}
	return out
}

func partsOfType(parts []map[string]any, want string) []map[string]any {
	var out []map[string]any
	for _, p := range parts {
		if t, _ := p["type"].(string); normName(t) == normName(want) {
			out = append(out, p)
		}
	}
	return out
}

// wireResultIDs keys each call id to the FIRST wire result answering it.
// A result is replayed in every later request, so occurrences are counted by
// distinct call id rather than by appearance.
func wireResultIDs(s *Ch3Session) map[string]WireResult {
	out := map[string]WireResult{}
	for _, r := range wireResults(s) {
		if _, seen := out[r.CallID]; !seen {
			out[r.CallID] = r
		}
	}
	return out
}

func crashed(s *Ch3Session) string {
	if strings.Contains(s.Stderr, "panic:") || strings.Contains(s.Stderr, "goroutine ") {
		return "the agent panicked"
	}
	if strings.Contains(s.Stderr, "[grader] timed out") {
		return "the agent never exited (the loop did not terminate)"
	}
	return ""
}

func joined(bad []string) string { return strings.Join(bad, "; ") }

// --- ch2parity -------------------------------------------------------------

func checkCh2Parity(res *Ch3Result) Check {
	c := Check{ID: "ch2parity", Title: "Chapter 2's log, reducer and three renderers still work",
		Points: 10, Passed: true, Earned: 10}

	if res.Ch2Err != "" {
		c.failf("Chapter 2's harness could not run this binary: %s", res.Ch2Err)
		return c
	}
	if len(res.Ch2) == 0 {
		c.failf("no Chapter 2 checks ran, so parity was never actually tested")
		return c
	}
	var failed []string
	for _, ch := range res.Ch2 {
		if !ch.Passed {
			failed = append(failed, fmt.Sprintf("%s (%d/%d)", ch.ID, ch.Earned, ch.Points))
		}
	}
	if len(failed) > 0 {
		c.failf("adding the tool loop broke Chapter 2: %s", joined(failed))
		return c
	}
	c.Details = append(c.Details, fmt.Sprintf("all %d Chapter 2 checks still pass", len(res.Ch2)))
	return c
}

// --- toolloop --------------------------------------------------------------

// checkToolLoop is the chapter's core 25 points, and it runs against all three
// vendors because the loop is new work that Chapter 2's parity cannot cover.
// Gemini matters most here: its finishReason is "STOP" even when it asks for a
// tool, so an implementation keyed on the stop signal — which works on the
// other two — silently never calls anything.
//
// Twenty points, not twenty-five: the other five moved to toolsdecl, the
// request-direction half of the same protocol, when it became graded.
func checkToolLoop(res *Ch3Result) Check {
	c := Check{ID: "toolloop", Title: "parse tool_use, dispatch, return tool_result by id, loop until the model stops",
		Points: 20, Passed: true, Earned: 20}

	var bad []string
	for _, vendor := range Ch2Vendors {
		s := res.Loop[vendor]
		if s == nil {
			bad = append(bad, vendor+": scenario did not run")
			continue
		}
		p := func(format string, args ...any) {
			bad = append(bad, vendor+": "+fmt.Sprintf(format, args...))
		}
		if s.Err != "" {
			p("harness error: %s", s.Err)
			continue
		}
		if m := crashed(s); m != "" {
			p("%s", m)
		}

		id := func(n string) string {
			switch vendor {
			case "openai":
				return "call_" + n
			case "gemini":
				return "fc_" + n
			}
			return "toolu_" + n
		}

		// The loop ran three rounds: prompt, result, result. Anything less and
		// the agent stopped asking before the model did.
		if len(s.Requests) < 3 {
			p("made %d requests, want 3 (a tool call is the middle of a turn, not the end)", len(s.Requests))
		}

		// THE TYPE FILTER. One message carried prose AND a call; both must
		// have been recorded from it.
		rp := responseParts(s)
		if len(rp) == 0 {
			p("no response was recorded in the log")
		} else {
			texts := partsOfType(rp[0], "text")
			calls := partsOfType(rp[0], "tool_call")
			if len(texts) != 1 {
				p("first reply recorded %d text parts, want exactly 1", len(texts))
			} else if got, _ := texts[0]["text"].(string); got != "I'll check the toolchain." {
				p("recorded assistant text was %q, want %q (a walk that does not dispatch on block type mixes the call into the prose)",
					got, "I'll check the toolchain.")
			}
			if len(calls) != 1 {
				p("first reply recorded %d tool_call parts, want exactly 1 (the message carried both a text block and a tool_use block)", len(calls))
			}
		}

		results := wireResultIDs(s)
		if len(results) != 2 {
			p("sent results for %d distinct call ids, want 2", len(results))
		}

		r1, ok1 := results[id("loop_1")]
		if !ok1 {
			p("no tool_result was sent for call %s", id("loop_1"))
		} else if !strings.Contains(strings.ToLower(r1.Text), "go version") {
			p("result for %s does not contain the output of `go version`: %.80q", id("loop_1"), r1.Text)
		}

		r2, ok2 := results[id("loop_2")]
		switch {
		case !ok2:
			p("no tool_result was sent for call %s", id("loop_2"))
		case !strings.Contains(r2.Text, "gamma") || !strings.Contains(r2.Text, "delta"):
			p("result for %s is missing the requested lines 3-4: %.80q", id("loop_2"), r2.Text)
		case strings.Contains(r2.Text, "alpha") || strings.Contains(r2.Text, "epsilon"):
			p("result for %s returned the whole file rather than lines 3-4: %.80q", id("loop_2"), r2.Text)
		}

		// Anthropic requires the tool_result to come FIRST in the content
		// array of the message answering it.
		if vendor == "anthropic" && ok1 && !r1.First {
			p("the tool_result for %s was not the first block of its message", id("loop_1"))
		}

		// The turn ended with the model's final text, not with a tool result.
		final := "Toolchain is present and the notes look right."
		if len(s.Answers) == 0 {
			p("no assistant answer was printed")
		} else if got := s.Answers[len(s.Answers)-1]; got != final {
			p("final answer was %q, want %q", got, final)
		}
	}

	if len(bad) > 0 {
		c.failf("%s", joined(bad))
		return c
	}

	// Anthropic requires the tool_result to come FIRST in the content array of
	// the message answering it. This is graded on a RENDER of the exhibit log,
	// because it is only observable in a message carrying a tool_result AND
	// something else — and no live scenario in this chapter can build one.
	if first, found := exhibitResultFirst(res); !found {
		c.failf("could not read a rendered tool_result back from the exhibit log: %s", res.ExhibitErr)
		return c
	} else if !first {
		c.failf("the tool_result was not the first block of the message answering it; " +
			"Anthropic requires results first when the message also carries anything else")
		return c
	}

	c.Details = append(c.Details, "loop, dispatch and id-keyed results verified on all three vendors; results spliced first")
	return c
}

// --- multiblock ------------------------------------------------------------

// checkMultiblock uses TWO calls to the SAME tool with different arguments, so
// that a result can only be matched to its call by id. Keying by name or by
// position produces visibly swapped answers, and each half of the pair is the
// negative control for the other: "alpha" must be absent from one result and
// present in the other.
func checkMultiblock(res *Ch3Result) Check {
	c := Check{ID: "multiblock", Title: "text + two tool calls: all parts recorded, both dispatched, results matched by id",
		Points: 10, Passed: true, Earned: 10}

	s := res.Multiblock
	if s == nil || s.Err != "" {
		c.failf("scenario did not run")
		return c
	}
	var bad []string
	p := func(format string, args ...any) { bad = append(bad, fmt.Sprintf(format, args...)) }

	if m := crashed(s); m != "" {
		p("%s", m)
	}

	rp := responseParts(s)
	if len(rp) == 0 {
		p("no response was recorded in the log")
	} else {
		texts := partsOfType(rp[0], "text")
		calls := partsOfType(rp[0], "tool_call")
		if len(texts) != 1 {
			p("recorded %d text parts, want 1", len(texts))
		} else if got, _ := texts[0]["text"].(string); got != "Reading the first and last lines." {
			p("recorded assistant text was %q", got)
		}
		if len(calls) != 2 {
			p("recorded %d tool_call parts, want 2 — both calls arrived in ONE message", len(calls))
		}
	}

	dispatched := logToolCalls(s)
	for _, want := range []string{"toolu_mb_a", "toolu_mb_b"} {
		if !dispatched[want] {
			p("call %s was never dispatched", want)
		}
	}

	results := wireResultIDs(s)
	if len(results) != 2 {
		p("sent results for %d distinct call ids, want 2", len(results))
	}

	a, okA := results["toolu_mb_a"]
	b, okB := results["toolu_mb_b"]
	switch {
	case !okA || !okB:
		p("both calls must be answered: got a=%v b=%v", okA, okB)
	default:
		// a asked for line 1 (alpha), b asked for line 5 (epsilon).
		if !strings.Contains(a.Text, "alpha") || strings.Contains(a.Text, "epsilon") {
			p("result for toolu_mb_a should be line 1 (alpha) only, got %.60q", a.Text)
		}
		if !strings.Contains(b.Text, "epsilon") || strings.Contains(b.Text, "alpha") {
			p("result for toolu_mb_b should be line 5 (epsilon) only, got %.60q — results were matched to the wrong calls", b.Text)
		}
		// Both answer one assistant message, so both belong to one request.
		if a.Request != b.Request {
			p("the two results were sent in different requests (%d and %d); both answer the same message", a.Request, b.Request)
		}
	}

	if len(bad) > 0 {
		c.failf("%s", joined(bad))
		return c
	}
	c.Details = append(c.Details, "two same-name calls kept distinct and answered by id")
	return c
}

// --- readtools / mutatetools -----------------------------------------------

// localToolProbes runs the LocalTools scenario's assertions once and reports,
// per tool, the first thing wrong with it. Two checks read this map:
// readtools (read_file with range, list_directory, search_files) and
// mutatetools (write_file, edit_file). They are separate rows because a
// student can — and in practice does — have one half working and the other
// broken, and a single 20-point row could not say which. list_directory and
// search_files do not split out further: nobody has search working and read
// broken.
//
// Where possible the probes look at the DISK rather than at the tool's own
// report: a write_file that says "wrote 42 bytes" and writes nothing is
// exactly the failure worth catching, and only the file system can tell.
func localToolProbes(res *Ch3Result) (broken map[string]string, scenarioErr string) {
	s := res.LocalTools
	if s == nil {
		return nil, "scenario did not run"
	}
	if s.Err != "" {
		return nil, "scenario did not run: " + s.Err
	}

	results := wireResultIDs(s)
	greet := s.After["src/greet.go"]
	broken = map[string]string{}
	fail := func(tool, bad string) {
		if _, seen := broken[tool]; !seen {
			broken[tool] = bad
		}
	}

	// write_file: the file must exist with the content it was given.
	switch {
	case greet == "":
		fail("write_file", "src/greet.go was not created")
	case !strings.Contains(greet, "package src"):
		fail("write_file", "src/greet.go does not contain the content that was written")
	}

	// edit_file: the bytes on disk must have changed, not just been reported.
	switch {
	case greet == "":
		fail("edit_file", "no file to edit (write_file failed first)")
	case !strings.Contains(greet, "goodbye"):
		fail("edit_file", "the replacement text is not in the file")
	case strings.Contains(greet, "hello"):
		fail("edit_file", "the original text is still in the file")
	}

	if r, ok := results["toolu_lt_ls"]; !ok {
		fail("list_directory", "no result was returned")
	} else if !strings.Contains(r.Text, "notes.md") || !strings.Contains(r.Text, "testdata") {
		fail("list_directory", "listing of the working directory does not show notes.md and testdata/: "+fmt.Sprintf("%.60q", r.Text))
	}

	if r, ok := results["toolu_lt_grep"]; !ok {
		fail("search_files", "no result was returned")
	} else if !strings.Contains(r.Text, "notes.md") || !strings.Contains(r.Text, "gamma line three") {
		fail("search_files", "search for \"gamma\" did not report notes.md and the matching line: "+fmt.Sprintf("%.60q", r.Text))
	}

	// read_file WITH A RANGE. The absence assertions here are controlled by
	// the multiblock scenario, which returns "alpha" legitimately from a
	// range read of line 1 — so a checker that cannot see "alpha" at all
	// would fail there rather than passing vacuously here.
	if r, ok := results["toolu_lt_read"]; !ok {
		fail("read_file", "no result was returned")
	} else if !strings.Contains(r.Text, "gamma") || !strings.Contains(r.Text, "delta") {
		fail("read_file", "lines 3-4 are missing: "+fmt.Sprintf("%.60q", r.Text))
	} else if strings.Contains(r.Text, "alpha") || strings.Contains(r.Text, "epsilon") {
		fail("read_file", "the range was ignored and the whole file came back")
	}
	return broken, ""
}

// gradeToolGroup turns the probe map into one check over a named subset of
// tools, deducting perTool points for each broken one, floored at zero.
func gradeToolGroup(c Check, tools []string, perTool int, broken map[string]string, scenarioErr string) Check {
	if scenarioErr != "" {
		c.failf("%s", scenarioErr)
		return c
	}
	var msgs []string
	for _, t := range tools {
		if bad, ok := broken[t]; ok {
			msgs = append(msgs, t+": "+bad)
		}
	}
	if len(msgs) == 0 {
		c.Details = append(c.Details, fmt.Sprintf("%s verified, against the disk where possible", strings.Join(tools, ", ")))
		return c
	}
	c.Passed = false
	c.Earned = c.Points - perTool*len(msgs)
	if c.Earned < 0 {
		c.Earned = 0
	}
	c.Details = append(c.Details, fmt.Sprintf("%d of %d work: %s", len(tools)-len(msgs), len(tools), joined(msgs)))
	return c
}

// checkReadTools grades the three tools that only look: read_file with a
// range, list_directory, search_files. Ten points, four off per broken tool.
func checkReadTools(res *Ch3Result) Check {
	c := Check{ID: "readtools", Title: "read_file (with range), list_directory, search_files",
		Points: 10, Passed: true, Earned: 10}
	broken, err := localToolProbes(res)
	return gradeToolGroup(c, []string{"read_file", "list_directory", "search_files"}, 4, broken, err)
}

// checkMutateTools grades the two tools that change the disk: write_file and
// edit_file. Ten points, five each.
func checkMutateTools(res *Ch3Result) Check {
	c := Check{ID: "mutatetools", Title: "write_file, edit_file change the bytes on disk",
		Points: 10, Passed: true, Earned: 10}
	broken, err := localToolProbes(res)
	return gradeToolGroup(c, []string{"write_file", "edit_file"}, 5, broken, err)
}

// --- runcommand ------------------------------------------------------------

// checkRunCommand grades three separable things at five points each: stdout
// comes back, stderr comes back, and the exit code comes back.
//
// The exit code is graded on a bare `exit 7`, which produces no output at all,
// rather than on `go run ./testdata/exit7`. Measured: `go run` prints "exit
// status 7" on its own stderr and then exits 1 itself, so the named fixture
// would let a tool that reports no exit code whatsoever match an "exit ... 7"
// assertion out of vendor noise, while the code actually delivered would be 1.
// The successful command is the negative control for the same assertion.
func checkRunCommand(res *Ch3Result) Check {
	c := Check{ID: "runcommand", Title: "shell executes, stdout/stderr/exit code returned",
		Points: 15, Passed: true, Earned: 15}

	s := res.Shell
	if s == nil || s.Err != "" {
		c.failf("scenario did not run")
		return c
	}
	results := wireResultIDs(s)
	broken := map[string]bool{}
	var msgs []string
	fail := func(area, format string, args ...any) {
		broken[area] = true
		msgs = append(msgs, area+": "+fmt.Sprintf(format, args...))
	}

	// --- stdout ---
	if r, ok := results["toolu_sh_version"]; !ok {
		fail("stdout", "`go version` got no result")
	} else if !strings.Contains(strings.ToLower(r.Text), "go version") {
		fail("stdout", "`go version` output did not reach the model (%.60q)", r.Text)
	}
	if r, ok := results["toolu_sh_noisy"]; !ok {
		fail("stdout", "the noisy program got no result")
	} else if !strings.Contains(r.Text, "STDOUT_MARKER") {
		fail("stdout", "the program's stdout is missing from the result")
	}

	// --- stderr ---
	if r, ok := results["toolu_sh_noisy"]; !ok {
		fail("stderr", "the noisy program got no result")
	} else if !strings.Contains(r.Text, "STDERR_MARKER") {
		fail("stderr", "the program's stderr is missing from the result")
	}

	// --- exit code ---
	if r, ok := results["toolu_sh_exit7"]; !ok {
		fail("exitcode", "`exit 7` got no result")
	} else if !exitCodeRe.MatchString(r.Text) {
		fail("exitcode", "the exit code of a silent `exit 7` never reached the model (%.60q)", r.Text)
	}
	// NEGATIVE CONTROL: a successful command must not look like a failing one.
	// Without this, printing "exit 7" unconditionally would score the points.
	if r, ok := results["toolu_sh_version"]; ok && exitCodeRe.MatchString(r.Text) {
		fail("exitcode", "a successful command also reported exit 7, so the code is boilerplate rather than the real status")
	}

	// The chapter's own named fixture still has to work, and must not be
	// mistaken for a broken call.
	if r, ok := results["toolu_sh_gorun"]; !ok {
		fail("exitcode", "`go run ./testdata/exit7` got no result")
	} else if r.IsError {
		fail("exitcode", "`go run ./testdata/exit7` was marked as a tool error; it ran")
	}

	if len(broken) > 0 {
		c.Passed = false
		c.Earned = c.Points - 5*len(broken)
		if c.Earned < 0 {
			c.Earned = 0
		}
		c.Details = append(c.Details, joined(msgs))
		return c
	}
	c.Details = append(c.Details, "stdout, stderr and a real exit code all reach the model")
	return c
}

// --- toolerror -------------------------------------------------------------

// checkToolError is the chapter's most valuable fifteen points.
//
// Three calls that cannot work must come back to the MODEL as tool_results
// marked as errors. The fourth — a command that exits non-zero — must NOT be
// marked as an error: it ran, and the answer is "the program failed", which is
// a true result rather than a bad call. That fourth case is the negative
// control: without it, marking everything as an error would score full marks.
func checkToolError(res *Ch3Result) Check {
	c := Check{ID: "toolerror", Title: "failed and malformed calls return an error to the model, and the loop continues",
		Points: 15, Passed: true, Earned: 15}

	s := res.ToolError
	if s == nil || s.Err != "" {
		c.failf("scenario did not run")
		return c
	}
	var bad []string
	p := func(format string, args ...any) { bad = append(bad, fmt.Sprintf(format, args...)) }

	if m := crashed(s); m != "" {
		p("%s", m)
	}

	results := wireResultIDs(s)
	want := []string{"toolu_err_missing", "toolu_err_unknown", "toolu_err_badargs", "toolu_err_exit7"}
	for _, id := range want {
		if _, ok := results[id]; !ok {
			p("call %s got no tool_result at all — a failure that is silently skipped leaves the model waiting", id)
		}
	}

	// Each failure must say enough for the model to act on it.
	if r, ok := results["toolu_err_missing"]; ok {
		if !r.IsError {
			p("a read of a missing file was not marked as an error")
		}
		if !strings.Contains(r.Text, "no-such-file-anywhere.txt") {
			p("the missing-file error does not name the file: %.80q", r.Text)
		}
	}
	if r, ok := results["toolu_err_unknown"]; ok {
		if !r.IsError {
			p("a call to a tool that does not exist was not marked as an error")
		}
		if !strings.Contains(r.Text, "frobnicate") {
			p("the unknown-tool error does not name the tool: %.80q", r.Text)
		}
	}
	if r, ok := results["toolu_err_badargs"]; ok {
		if !r.IsError {
			p("arguments of the wrong type were not marked as an error")
		}
		if strings.TrimSpace(r.Text) == "" {
			p("the bad-arguments error came back empty, so the model cannot correct it")
		}
	}

	// NEGATIVE CONTROL: this one ran. Marking it as an error is a false claim.
	if r, ok := results["toolu_err_exit7"]; ok {
		if r.IsError {
			p("a command that exited 7 was marked as a tool error; it ran, and its non-zero exit is a result, not a bad call")
		}
		if !exitCodeRe.MatchString(r.Text) {
			p("the non-zero exit status was not reported to the model: %.80q", r.Text)
		}
	}

	// The loop must have kept going through all four failures.
	if len(s.Requests) < 5 {
		p("made %d requests, want 5 — the loop stopped at the first failure", len(s.Requests))
	}
	final := "I recovered from all of that."
	if len(s.Answers) == 0 {
		p("no assistant answer was printed; the agent locked up rather than recovering")
	} else if got := s.Answers[len(s.Answers)-1]; got != final {
		p("final answer was %q, want %q", got, final)
	}

	if len(bad) > 0 {
		c.failf("%s", joined(bad))
		return c
	}
	c.Details = append(c.Details, "three failures returned as errors, a non-zero exit correctly not an error, loop survived")
	return c
}

// --- editcontract ----------------------------------------------------------

// checkEditContract grades the chapter's DECLINED DECISION, and it must accept
// all three answers: refuse, fuzzy-match, and rewrite.
//
// It therefore asserts nothing about WHICH happened. It asserts that something
// coherent happened and that the model was told:
//
//   - the call was answered at all, on the wire;
//   - the log records the outcome;
//   - and the report matches reality — either it reported failure AND left the
//     file alone, or it reported success AND actually changed the file.
//
// The combination this rejects is the one no answer defends: claiming success
// while changing nothing, or reporting an error after mutating the file.
func checkEditContract(res *Ch3Result) Check {
	c := Check{ID: "editcontract", Title: "edit_file's failure contract: a choice was made, logged, and recoverable",
		Points: 5, Passed: true, Earned: 5}

	s := res.EditContrac
	if s == nil || s.Err != "" {
		c.failf("scenario did not run")
		return c
	}
	var bad []string
	p := func(format string, args ...any) { bad = append(bad, fmt.Sprintf(format, args...)) }

	if m := crashed(s); m != "" {
		p("%s", m)
	}

	r, ok := wireResultIDs(s)["toolu_ec_1"]
	if !ok {
		c.failf("the failed edit was never answered: no tool_result reached the model for toolu_ec_1")
		return c
	}
	if strings.TrimSpace(r.Text) == "" {
		p("the result was empty, so the model has nothing to act on")
	}

	// The log must make the outcome legible.
	if _, logged := logToolReturns(s)[toolECID]; !logged {
		p("the log records no tool_returned event for the failed edit")
	}

	before, after := s.Before["contract.txt"], s.After["contract.txt"]
	changed := before != after

	switch {
	case r.IsError && changed:
		p("the edit was reported as a failure but the file was modified anyway")
	case r.IsError && !changed:
		// REFUSE. A refusal must say what it saw, or it costs a round trip to
		// find out why.
		if !strings.Contains(r.Text, "contract.txt") && !strings.Contains(r.Text, "this anchor is definitely not in the file") {
			p("the refusal names neither the file nor the anchor it looked for: %.100q", r.Text)
		}
	case !r.IsError && changed:
		// FUZZY-MATCH or REWRITE. Both are accepted; the file really moved.
	case !r.IsError && !changed:
		p("the edit reported success but the file is unchanged — whichever answer was chosen, this is not one of them")
	}

	// Whatever happened, the conversation continued.
	if len(s.Answers) == 0 {
		p("the agent produced no answer after the failed edit")
	}

	if len(bad) > 0 {
		c.failf("%s", joined(bad))
		return c
	}
	outcome := "refused and left the file alone"
	if !r.IsError {
		outcome = "applied a fallback edit and said so"
	}
	c.Details = append(c.Details, "a coherent choice was made: "+outcome)
	return c
}

const toolECID = "toolu_ec_1"
