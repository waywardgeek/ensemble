package gradeedition2

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter13 grades the executable speech transcript contract and corrected parity.
// Actual blind-Agent capability, final-only renderer delivery and native speech
// retain separate browser/source proof: this log cannot expose those boundaries.
func Chapter13(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	checks := speechChecks(core)
	previous, err := Chapter12(bin, core, gui, guiModel, consumer, workflow, exercise, streamModel, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	parity := grade.Check{ID: "ch12-parity", Title: "All corrected Chapter 12 checks pass", Points: 5, Earned: 5, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	return grade.NewTitledReport("Edition 2, Chapter 13 — The Channel Nobody Tested", append(checks, parity), ""), nil
}

// Chapter13Executable retains all nine speech assertions and their five real
// launches. It does not claim inherited Go checks, listener restriction or sound.
func Chapter13Executable(core string) grade.Report {
	return grade.NewTitledReport("Edition 3 C#, Chapter 13 — executable speech (excludes inherited checks and acoustic proof)", speechChecks(core), "")
}

// Keep the original executable input/log vocabulary. Source and exact local
// object names are not part of this contract; source-aware tests live separately.
type speechEntry struct {
	// Kind distinguishes actual utterances from pause/resume observations.
	Kind string `json:"kind"`
	// Text is the speech-channel output, not the raw unfiltered source.
	Text string `json:"text"`
	// Cause records the harness-observed gate cause when available.
	Cause string `json:"cause"`
}
type speechManifest struct {
	// ReadFileTool names the existing file tool used for silence discrimination.
	ReadFileTool string `json:"read_file_tool"`
	// ReadFileArg is its public path argument, with no required implementation name.
	ReadFileArg string `json:"read_file_arg"`
	// SupportsType advertises the established mid-turn typing input.
	SupportsType bool `json:"supports_type_during_turn"`
	// SupportsOff advertises provider nonstream mode, not renderer-only finals.
	SupportsOff bool `json:"supports_streaming_off"`
}

// Each phrase is independently recognizable. The fence has no intervening prose:
// something between its two sentinels must be the announcement, whatever its name.
const speechProse = "Reading carefully preserves fragments.\n\n" +
	"Wrapped words stay\ntogether.\n\nFirst paragraph\n\nSecond paragraph\n\n" +
	"# Heading\n\n**Emphasis** and `inline` with [linked words](https://example.invalid).\n\n" +
	"- Bullet words\n- Following bullet.\n\nUse HTTPServer with queueChunk and max_tool_rounds in ALL CAPS. Hello ordinary World.\n\n" +
	"Before the fence.\n\n```go\nFENCED_BODY_ONLY();\n```\n\nAfter the fence.\n\nTail without punctuation:"

func speechChecks(core string) []grade.Check {
	// The actual original rubric is 95 speech points, unlike its printed table.
	checks := grade.Ch14Checks(grade.Ch14Result{HarnessOK: true, DiscriminationOK: true, FiltersOK: true, BuffersOK: true, UnstreamedOK: true, GateOK: true, ErrorsOK: true, BoundariesOK: true, IdentifiersOK: true})[:9]
	checks[4].Title = "Provider nonstream replies are spoken once (final-only renderer proof is separate)"
	checks[5].Title = "Typing tool-dispatch pause and subsequent release appear in the transcript"
	script, err := speechHarness(core)
	if err != nil {
		for i := range checks {
			fail(&checks[i], "%v", err)
		}
		return checks
	}
	man := speechManifest{ReadFileTool: "read_file", ReadFileArg: "path", SupportsType: true, SupportsOff: true}
	// --describe is optional in the existing protocol. A missing/invalid response
	// retains the historical defaults; observed behavior must still pass below.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	data, describeErr := exec.CommandContext(ctx, script, "--describe").Output()
	cancel()
	if describeErr == nil {
		var supplied speechManifest
		if json.Unmarshal(data, &supplied) == nil {
			if supplied.ReadFileTool != "" {
				man.ReadFileTool = supplied.ReadFileTool
			}
			if supplied.ReadFileArg != "" {
				man.ReadFileArg = supplied.ReadFileArg
			}
			man.SupportsType, man.SupportsOff = supplied.SupportsType, supplied.SupportsOff
		}
	}
	// Five launches retain the original cost/shape: prose, whole prose,
	// discrimination, failed provider, and mid-turn typing. Each serves several
	// assertions without making a fresh browser for every phrase or marker.
	run := func(label, mode, typing string, replies []fakevendor.Reply, fileArg string, indices ...int) ([]speechEntry, []fakevendor.Recorded) {
		entries, requests, err := runSpeech(script, mode, typing, replies, fileArg)
		if err != nil {
			fail(&checks[0], "%s: %v", label, err)
			for _, i := range indices {
				fail(&checks[i], "%s harness: %v", label, err)
			}
		}
		return entries, requests
	}
	entries, requests := run("streamed prose", "on", "", []fakevendor.Reply{{Text: speechProse}}, "", 2, 3, 7, 8)
	gradeSpeechProse(checks, entries)
	if len(requests) != 1 {
		fail(&checks[0], "streamed prose made %d provider requests, want1", len(requests))
	} else if !speechStreaming(requests[0]) {
		fail(&checks[3], "streamed fixture did not request actual provider deltas")
	}
	if !man.SupportsOff {
		fail(&checks[4], "harness reports streaming-off unavailable")
	} else {
		whole, requests := run("whole prose", "off", "", []fakevendor.Reply{{Text: speechProse}}, "", 2, 4)
		// Provider nonstream is observable here. Some real Engines synthesize deltas
		// for it, so this is deliberately not labeled a final-without-deltas test.
		if len(requests) != 1 {
			fail(&checks[4], "whole prose made %d provider requests, want1", len(requests))
		} else if speechStreaming(requests[0]) {
			fail(&checks[4], "streaming-off fixture still requested provider deltas")
		}
		spoken := speechWords(whole)
		for _, phrase := range []string{"Reading carefully preserves fragments.", "After the fence.", "Tail without punctuation:"} {
			if strings.Count(spoken, phrase) != 1 {
				fail(&checks[4], "whole reply must speak %q exactly once: %q", phrase, spoken)
			}
		}
		checkSpeechMarkup(&checks[2], spoken)
	}
	entries, requests = run("discrimination", "on", "", []fakevendor.Reply{{ToolName: man.ReadFileTool, ToolID: "speech-read"}, {Text: "The model says PELICAN."}}, man.ReadFileArg, 1)
	spoken := speechWords(entries)
	if strings.Contains(spoken, "ZEBRAFISH") || strings.Count(spoken, "PELICAN") != 1 {
		fail(&checks[1], "tool/model discrimination: %q", spoken)
	}
	// Silence is meaningful only if the real read succeeded. Otherwise even an
	// implementation that speaks all results can pass by returning an empty error.
	if len(requests) != 2 {
		fail(&checks[1], "read conversation made %d requests, want2", len(requests))
	} else if !speechReadResult(requests[1], man.ReadFileTool) {
		fail(&checks[1], "actual file result missing")
	}
	entries, _ = run("provider error", "on", "", []fakevendor.Reply{{Status: 500, ErrBody: `{"error":{"message":"upstream is on fire"}}`}}, "", 6)
	spoken = strings.ToLower(speechWords(entries))
	if !strings.Contains(spoken, "500") && !strings.Contains(spoken, "error") && !strings.Contains(spoken, "fail") && !strings.Contains(spoken, "upstream is on fire") {
		fail(&checks[6], "no failure-specific speech: %q", spoken)
	}
	if !man.SupportsType {
		fail(&checks[5], "harness reports mid-turn typing unavailable")
	} else {
		entries, _ = run("typing gate", "on", " ", []fakevendor.Reply{{Text: "One. Two. Three. Four. Five. Six."}}, "", 5)
		paused, typing, resumed := false, false, false
		for _, e := range entries {
			switch e.Kind {
			case "pause":
				paused = true
				if strings.Contains(e.Cause, "typing") {
					typing = true
				}
			case "resume":
				if paused && typing {
					resumed = true
				}
				paused = false
			}
		}
		if !typing || !resumed || paused {
			fail(&checks[5], "typing pause/release missing or final gate remains paused: typing=%v resumed=%v paused=%v", typing, resumed, paused)
		}
		// The log does not identify field edits or both independent causes. Initial
		// prompt typing can provide its typing edge, so the space argument is not
		// by itself proof of whitespace preservation or two-cause ordering.
	}
	return checks
}

func gradeSpeechProse(checks []grade.Check, entries []speechEntry) {
	spoken := speechWords(entries)
	checkSpeechMarkup(&checks[2], spoken)
	for _, phrase := range []string{"Reading carefully preserves fragments.", "Wrapped words stay together.", "After the fence.", "Tail without punctuation:"} {
		if strings.Count(spoken, phrase) != 1 {
			fail(&checks[3], "fragmented or repeated %q: %q", phrase, spoken)
		}
	}
	// A joined transcript alone cannot tell whether a wrap became a pause.
	// Require these exact phrases as individual utterances, including paragraphs
	// with no punctuation and the tail that only explicit flush can release.
	utterances := map[string]int{}
	for _, e := range entries {
		if e.Kind == "utterance" {
			utterances[strings.Join(strings.Fields(e.Text), " ")]++
		}
	}
	for _, phrase := range []string{"Reading carefully preserves fragments.", "Wrapped words stay together.", "First paragraph", "Second paragraph", "Use HTTP Server with queue Chunk and max tool rounds in ALL CAPS.", "Hello ordinary World.", "Tail without punctuation:"} {
		if utterances[phrase] != 1 {
			fail(&checks[7], "want one complete utterance %q, got%d", phrase, utterances[phrase])
		}
	}
	for _, phrase := range []string{"HTTP Server", "queue Chunk", "max tool rounds", "ALL CAPS", "Hello ordinary World."} {
		if strings.Count(spoken, phrase) != 1 {
			fail(&checks[8], "identifier/ordinary text %q missing or repeated: %q", phrase, spoken)
		}
	}
}

func checkSpeechMarkup(check *grade.Check, spoken string) {
	for _, marker := range []string{"*", "`", "#", "[", "]", "https://example.invalid", "FENCED", "BODY", "- Bullet", "- Following"} {
		if strings.Contains(spoken, marker) {
			fail(check, "markup/code leaked %q: %q", marker, spoken)
		}
	}
	for _, words := range []string{"Heading", "Emphasis and inline with linked words.", "Bullet words", "Following bullet."} {
		if !strings.Contains(spoken, words) {
			fail(check, "filtered content lost %q: %q", words, spoken)
		}
	}
	before, after := "Before the fence.", "After the fence."
	start, end := strings.Index(spoken, before), strings.Index(spoken, after)
	if start < 0 || end < start+len(before) || strings.Trim(strings.TrimSpace(spoken[start+len(before):end]), ".!?:") == "" {
		fail(check, "fence was not named between its surrounding sentences: %q", spoken)
	}
}

func speechWords(entries []speechEntry) string {
	var words []string
	for _, e := range entries {
		if e.Kind == "utterance" {
			words = append(words, strings.Fields(e.Text)...)
		}
	}
	return strings.Join(words, " ")
}

// Discover the established executable locations first, then the original
// name-based fallback. This locates a public entrypoint, not implementation code.
func speechHarness(core string) (string, error) {
	root, err := filepath.Abs(core)
	if err != nil {
		return "", err
	}
	for _, name := range []string{"scripts/tts-harness.sh", "tts-harness.sh"} {
		p := filepath.Join(root, name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, nil
		}
	}
	var found string
	err = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found == "" && !d.IsDir() && strings.Contains(d.Name(), "tts-harness") {
			found = p
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("no executable tts-harness under %s", core)
	}
	return found, nil
}

// The harness owns its browser/application children. Bound the entire process
// group on failure; killing only a build/shell wrapper can leave children alive.
func runSpeech(script, mode, typing string, replies []fakevendor.Reply, fileArg string) ([]speechEntry, []fakevendor.Recorded, error) {
	work, err := os.MkdirTemp("", "edition2-speech-")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(work)
	if fileArg != "" {
		// The selected harness may launch its GUI from another cwd. An absolute
		// public tool path ensures this exercise tests silence after a successful
		// read, rather than accidentally passing on a missing relative file.
		file := filepath.Join(work, "notes.txt")
		if err = os.WriteFile(file, []byte("The actual file contains ZEBRAFISH.\n"), 0600); err != nil {
			return nil, nil, err
		}
		args, _ := json.Marshal(map[string]string{fileArg: file})
		replies[0].ToolArgs = string(args)
	}
	provider := fakevendor.New(replies)
	defer provider.Close()
	log := filepath.Join(work, "speech.jsonl")
	args := []string{"Exercise the speech channel."}
	if typing != "" {
		args = append(args, "--type-during-turn", typing)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, script, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// The provider and credentials are local fakes; explicit streaming mode avoids
	// inheriting an unrelated user's previous STREAMING=off setting.
	cmd.Env = append(os.Environ(), "LLM_BASE_URL="+provider.URL(), "WORKSPACE="+work, "TTS_LOG="+log, "STREAMING="+mode)
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) }
	cmd.WaitDelay = 3 * time.Second
	output, err := cmd.CombinedOutput()
	if err != nil {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		return nil, provider.Requests(), fmt.Errorf("harness: %w: %s", err, strings.TrimSpace(string(output)))
	}
	file, err := os.Open(log)
	if err != nil {
		return nil, provider.Requests(), err
	}
	defer file.Close()
	var entries []speechEntry
	scan := bufio.NewScanner(file)
	scan.Buffer(make([]byte, 4096), 4*1024*1024)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		if line == "" || !strings.HasPrefix(line, "{") {
			continue
		}
		var entry speechEntry
		if err = json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, provider.Requests(), fmt.Errorf("invalid JSON log: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, provider.Requests(), scan.Err()
}

// Gemini selects streaming in the endpoint; the other course vendors use JSON.
func speechStreaming(r fakevendor.Recorded) bool {
	return r.JSON()["stream"] == true || strings.Contains(r.Path, "streamGenerateContent") || strings.Contains(r.Path, "alt=sse")
}

// The executable harness may use any course vendor. Match only the actual
// result of our single read; its output marker never appears in call arguments.
func speechReadResult(r fakevendor.Recorded, tool string) bool {
	if r.Vendor == "anthropic" {
		text, bad, err := anthropicToolText(r.Body, "speech-read")
		return err == nil && !bad && strings.Contains(text, "ZEBRAFISH")
	}
	for _, raw := range list(r.JSON()["messages"]) {
		m := object(raw)
		if m["role"] == "tool" && m["tool_call_id"] == "speech-read" {
			data, _ := json.Marshal(m["content"])
			return strings.Contains(string(data), "ZEBRAFISH")
		}
	}
	for _, raw := range list(r.JSON()["contents"]) {
		for _, part := range list(object(raw)["parts"]) {
			result := object(object(part)["functionResponse"])
			if result["name"] == tool && (result["id"] == nil || result["id"] == "speech-read") {
				response := object(result["response"])
				data, _ := json.Marshal(response)
				return response["error"] == nil && strings.Contains(string(data), "ZEBRAFISH")
			}
		}
	}
	return false
}
