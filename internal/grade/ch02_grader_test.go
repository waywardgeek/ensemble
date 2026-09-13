package grade

// Mutation tests: proof that the grader is SENSITIVE, not merely green.
//
// Each mutation breaks exactly one thing in the reference solution and asserts
// the EXACT SET of check ids that fail. Asserting the exact set, rather than
// "something failed", is what catches a check that fires for the wrong reason
// — which is the failure mode that makes a grader worse than useless, because
// it is confidently wrong in the student's favour.
//
// A mutation that does not apply is a test that silently passes. Every edit
// below asserts that it actually changed the source.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type edit struct {
	file    string
	pattern string // regexp, tolerant of gofmt's alignment padding
	repl    string
}

type mutation struct {
	name string
	// why records what real student mistake this models.
	why      string
	edits    []edit
	extra    string // optional extra .go file contents
	wantFail []string
}

var mutations = []mutation{
	{
		name:     "provenance-model-dropped",
		why:      "parser forgets to record which model answered",
		edits:    []edit{{"anthropic.go", `Model:\s*resp\.Model,\s*Surface:\s*SurfaceMessages\}`, `Model: "", Surface: SurfaceMessages}`}},
		wantFail: []string{"seam-parse"},
	},
	{
		name:     "provenance-hardcoded",
		why:      "gemini parser copy-pasted from the anthropic one and kept its provenance",
		edits:    []edit{{"gemini.go", `from := Provenance\{Vendor: VendorGemini, Model: resp\.ModelVersion, Surface: SurfaceGenerateContent\}`, `from := Provenance{Vendor: VendorAnthropic, Model: "claude-sonnet-5-fake", Surface: SurfaceMessages}`}},
		wantFail: []string{"seam-parse"},
	},
	{
		name:     "openai-usage-summed-naively",
		why:      "treating OpenAI's prompt_tokens as if cached tokens were disjoint from it",
		edits:    []edit{{"openai.go", `uncached := u\.PromptTokens - u\.Details\.CachedTokens - u\.Details\.CacheWriteTokens`, `uncached := u.PromptTokens`}},
		wantFail: []string{"usage"},
	},
	{
		name:     "gemini-thoughts-assumed-included",
		why:      "assuming thoughtsTokenCount is inside candidatesTokenCount; undercounts billed output",
		edits:    []edit{{"gemini.go", `Output:\s*u\.CandidatesTokenCount \+ u\.ThoughtsTokenCount,`, `Output: u.CandidatesTokenCount,`}},
		wantFail: []string{"usage"},
	},
	{
		name:     "anthropic-cache-read-ignored",
		why:      "reading input_tokens alone on Anthropic, which undercounts on a warm cache",
		edits:    []edit{{"anthropic.go", `CacheRead:\s*resp\.Usage\.CacheReadTokens,`, `CacheRead: 0,`}},
		wantFail: []string{"usage"},
	},
	{
		name: "clock-in-the-renderer",
		why:  "the first of the four ways non-determinism gets into a renderer",
		edits: []edit{
			{"anthropic.go", `MaxTokens: cfg\.MaxTokens,`, `MaxTokens: cfg.MaxTokens + mutantNonce(),`},
		},
		extra:    "package main\n\nimport \"time\"\n\n// The clock, reaching into a renderer. It must vary ACROSS PROCESSES:\n// `replay` compares two separate runs, so a per-process counter starts at the\n// same value both times and is not a mutation at all.\nfunc mutantNonce() int { return int(time.Now().UnixNano() % 997) }\n",
		wantFail: []string{"replay"},
	},
	{
		name:     "anthropic-no-merge",
		why:      "rendering each entry as its own message, so the tool result and the following instruction do not merge",
		edits:    []edit{{"anthropic.go", `msgs\[n-1\]\.Role == role \{`, `msgs[n-1].Role == "\x00no-merge" {`}},
		wantFail: []string{"seam-render"},
	},
	{
		name:     "ephemera-never-cleared",
		why:      "forgetting that RequestSent consumes pending ephemera; a stale timestamp is a lie",
		edits:    []edit{{"context.go", `c\.Ephemera = nil`, `_ = 0`}},
		wantFail: []string{"ephemera"},
	},
	{
		name:     "ephemera-classified-as-dialogue",
		why:      "classifying at the capture site instead of in the reducer",
		edits:    []edit{{"context.go", `if e\.Message\.Actor == ActorSystem \{`, `if false {`}},
		wantFail: []string{"ephemera"},
	},
	{
		name:     "redaction-ignored",
		why:      "a reducer that is total by ignoring an event it should have handled",
		edits:    []edit{{"context.go", `c\.applyRedaction\(\*e\.Redact\)`, `_ = e.Redact`}},
		wantFail: []string{"redaction"},
	},
	{
		name:     "opaque-replayed-to-wrong-model",
		why:      "handing a thinking signature to a vendor that never issued it",
		edits:    []edit{{"gemini.go", `if op\.From\.SameModel\(target\) \{`, `if true {`}},
		wantFail: []string{"seam-render"},
	},
	{
		name:     "gemini-uses-messages-key",
		why:      "renaming Anthropic's shape instead of learning Gemini's",
		edits:    []edit{{"gemini.go", `json:"contents"`, `json:"messages"`}},
		wantFail: []string{"seam-render"},
	},
	{
		name: "openai-arguments-as-object",
		why:  "sending tool arguments as an object, when OpenAI wants a JSON-encoded string",
		edits: []edit{
			{"openai.go", `Arguments string ` + "`" + `json:"arguments"` + "`", "Arguments json.RawMessage `json:\"arguments\"`"},
			{"openai.go", `Arguments: string\(jsonObject\(call\.Args\)\),`, `Arguments: jsonObject(call.Args),`},
		},
		wantFail: []string{"seam-render"},
	},
	{
		name:     "system-prompt-omitted",
		why:      "the system prompt is renderer output; drop it and all three placements vanish — and Chapter 1 wire conformance, which requires a system prompt, breaks too",
		edits:    []edit{{"main.go", `SystemPrompt: systemPrompt,`, `SystemPrompt: "",`}},
		wantFail: []string{"ch1parity", "seam-render"},
	},
	{
		name:     "dump-prints-nothing",
		why:      "a log that cannot be dumped cannot be replayed, and the parse check reads the dump",
		edits:    []edit{{"main.go", `if err := log\.Write\(os\.Stdout\); err != nil \{`, `if _ = log; false {`}},
		wantFail: []string{"logdump", "seam-parse"},
	},
	{
		name:     "ch1-protocol-broken",
		why:      "renaming the reply key; ch1parity catches the regression and session names the cause",
		edits:    []edit{{"main.go", `emit\(out, map\[string\]string\{"assistant": reply\}\)`, `emit(out, map[string]string{"reply": reply})`}},
		wantFail: []string{"ch1parity", "session"},
	},
	{
		// The mutant for the field §2.4a grew. Before GeminiReplayLog existed
		// this mutation scored 100/100 — the field was graded by nothing in
		// either direction, so a student could omit it and be told Chapter 2
		// was perfect, then meet the 400 in Chapter 3.
		name:     "gemini-call-opaque-dropped",
		why:      "keeping opaque material only as a standalone part, so material bound to ONE call is lost; replaying a functionCall without its signature is a Gemini 3.x 400",
		edits:    []edit{{"gemini.go", `p\.ThoughtSignature = call\.Opaque`, `_ = call.Opaque`}},
		wantFail: []string{"seam-render"},
	},
}

func TestCh2ReferenceSolutionScores100(t *testing.T) {
	bin := buildMutant(t, nil, "")
	res, err := Ch2Run(bin)
	if err != nil {
		t.Fatalf("Ch2Run: %v", err)
	}
	checks := Ch2Evaluate(res)
	total, max := 0, 0
	for _, c := range checks {
		total += c.Earned
		max += c.Points
		if !c.Passed {
			t.Errorf("reference solution fails %q: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	if max != 100 {
		t.Errorf("checks sum to %d points, want exactly 100", max)
	}
	if total != 100 {
		t.Errorf("reference solution scored %d/100", total)
	}
}

func TestCh2MutationsAreDetected(t *testing.T) {
	if testing.Short() {
		t.Skip("mutation suite builds a binary per mutation")
	}
	for _, m := range mutations {
		m := m
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			bin := buildMutant(t, m.edits, m.extra)
			res, err := Ch2Run(bin)
			if err != nil {
				t.Fatalf("Ch2Run: %v", err)
			}
			var failed []string
			for _, c := range Ch2Evaluate(res) {
				if !c.Passed {
					failed = append(failed, c.ID)
				}
			}
			sort.Strings(failed)
			want := append([]string{}, m.wantFail...)
			sort.Strings(want)
			if strings.Join(failed, ",") != strings.Join(want, ",") {
				t.Errorf("mutation %q (%s)\n  failed: %v\n  want:   %v\n"+
					"When a mutation expectation misses, ask FIRST whether the grader is right.",
					m.name, m.why, failed, want)
			}
		})
	}
}

// buildMutant copies the reference solution into a throwaway module, applies
// the edits, and builds it. A throwaway module rather than a directory inside
// this one: the go tool would otherwise try to build the mutants as part of
// ./... and the copies would collide.
func buildMutant(t *testing.T, edits []edit, extra string) string {
	t.Helper()
	src := referenceDir(t)
	dir := t.TempDir()

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read reference solution: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module mutant\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if extra != "" {
		if err := os.WriteFile(filepath.Join(dir, "zz_mutant.go"), []byte(extra), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	for _, ed := range edits {
		path := filepath.Join(dir, ed.file)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("mutation targets missing file %s: %v", ed.file, err)
		}
		re, err := regexp.Compile(ed.pattern)
		if err != nil {
			t.Fatalf("bad mutation pattern %q: %v", ed.pattern, err)
		}
		if n := len(re.FindAll(b, -1)); n != 1 {
			// A mutation that does not apply is a test that silently passes:
			// the submission stays correct, the grader says 100/100, and the
			// suite records that as "the grader caught nothing" when nothing
			// was ever broken. A mutation that applies more than once is the
			// mirror image — the mutant then fails for reasons beyond the one
			// defect under test, so the failing-check set stops being
			// evidence about that defect. Demand exactly one site.
			t.Fatalf("mutation pattern %q matched %d sites in %s, want exactly 1",
				ed.pattern, n, ed.file)
		}
		out := re.ReplaceAll(b, []byte(ed.repl))
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	bin := filepath.Join(dir, "ch02bin")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building mutant failed: %v\n%s", err, out)
	}
	return bin
}

func referenceDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "solutions", "ch02")
}
