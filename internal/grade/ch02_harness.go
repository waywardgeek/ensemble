package grade

// Chapter 2 harness: drive the submission through every phase the chapter's
// checks need, recording evidence and judging nothing.
//
// Record then judge. One run surfaces every bug the student has, instead of
// one bug per run.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

const Ch2Timeout = 45 * time.Second

// Vendors, in the order the chapter fixes: the alien one goes LAST, so that
// "the third was nearly free" cannot be true merely because the third was easy.
var Ch2Vendors = []string{"anthropic", "openai", "gemini"}

// VendorSession is the evidence from one live session against one fake vendor.
type VendorSession struct {
	Vendor   string
	Answers  []string
	Usage    *Ch2Usage
	Requests []fakevendor.Recorded
	DumpOut  string // stdout of `ch02 dump`
	DumpErr  string
	Log      []Ch2LogLine
	LogErr   string
	Stderr   string
	Extra    []string
	Protocol []string
}

type Ch2Usage struct {
	Input      int `json:"input"`
	CacheWrite int `json:"cache_write"`
	CacheRead  int `json:"cache_read"`
	Output     int `json:"output"`
}

// Ch2Result is all evidence from all phases.
type Ch2Result struct {
	Ch1     []Check // the seven Chapter 1 checks, re-run unchanged
	Ch1Err  string
	Session map[string]*VendorSession

	// UsageProbe exercises the cache-WRITE category, which only two of the
	// three vendors report at all.
	UsageProbe map[string]*VendorSession

	// render phases — no network, no key
	Render      map[string]string // vendor -> rendered exhibit request
	RenderTwice map[string]string // second render of the same log
	RenderErr   map[string]string
	Redacted    map[string]string // render of a log containing a Redacted event
	RedactedErr map[string]string

	// ThoughtReplay is a Gemini render of a Gemini-authored log whose tool
	// call carries per-call replay material (ToolCallPart.Opaque).
	ThoughtReplay    string
	ThoughtReplayErr string

	// ephemera phase
	Ephemera *VendorSession

	// --- Ref phases --------------------------------------------------------
	//
	// A blob's location is a Ref (a Kind plus a locator), not a path. These
	// phases pin the three kinds, the two ways a Ref can be malformed, and the
	// one kind a vendor can fetch for itself.
	//
	// Every fixture carrying a LIVE blob is driven under GEMINI ONLY. Gemini is
	// the one vendor whose remote-file wire shape was verified against its
	// docs; the other two renderers refuse a blob loudly rather than guess a
	// field name, so rendering these fixtures under them would grade the
	// refusal, not the Ref.

	// RefDump is RefRoundTripLog loaded and re-emitted by `dump`. It is never
	// rendered: the three kinds are graded on surviving the LOG, rather than on
	// any vendor's opinion of them.
	RefDump    string
	RefDumpErr string

	// RefRender is RefURILog rendered for Gemini — the remote-reference form.
	RefRender    string
	RefRenderErr string

	// An old-format log and a zero-kind log must both be REFUSED. Exit code is
	// recorded because "refused" means a non-zero exit and a diagnostic, not a
	// silent empty render.
	RefOldFormatOut  string
	RefOldFormatErr  string
	RefOldFormatCode int
	RefZeroKindOut   string
	RefZeroKindErr   string
	RefZeroKindCode  int

	// RefRedacted is a render of a log whose redaction supersedes a tool result
	// containing a BlobPart. RefRedactedPlain is the same log with no Redacted
	// event — the negative control, without which the check would pass for a
	// submission that simply never renders tool results.
	RefRedacted         string
	RefRedactedErr      string
	RefRedactedPlain    string
	RefRedactedPlainErr string

	// logdump phase: dump from a live session, re-rendered in a fresh process
	RoundTripOut string
	RoundTripErr string

	Stderr   string
	Protocol []string
}

// ch2Replies is the scripted session. Round 2 returns a tool call and is the
// LAST round: Chapter 2 executes no tools, so nothing answers it, and no
// further request is made that would carry a dangling call.
func ch2Replies(vendor string) []fakevendor.Reply {
	toolID := map[string]string{
		"anthropic": "toolu_fake_1",
		"openai":    "call_fake_1",
		"gemini":    "fc_fake_1",
	}[vendor]
	return []fakevendor.Reply{
		{
			Text:  "Reading the configuration now.",
			Usage: fakevendor.Canonical{Input: 100, CacheWrite: 0, CacheRead: 50, Output: 30},
			// Gemini reports thoughts DISJOINT from candidates. A parser that
			// assumes they are included undercounts output by a third here.
			GeminiThoughts: 10,
		},
		{
			Text:           "Here is what I found.",
			ToolName:       "read_file",
			ToolArgs:       `{"path":"config.json","limit":40}`,
			ToolID:         toolID,
			Usage:          fakevendor.Canonical{Input: 12, CacheWrite: 0, CacheRead: 200, Output: 8},
			GeminiThoughts: 3,
		},
		{
			// CHAPTER 2 NEVER REACHES THIS REPLY. It sends two prompts and so
			// makes two requests; this is the third.
			//
			// Chapter 3's agent does reach it, because Chapter 3 EXECUTES the
			// tool call above and comes back for another turn. Without a
			// terminating reply the fake would repeat the last one — a tool
			// call — and the loop would run until its round limit.
			//
			// It reports ZERO usage, so the session total is identical whether
			// or not tools were executed. That is what lets Chapter 3's parity
			// check run Chapter 2's checks completely unmodified, including
			// the token accounting.
			Text:  "That is everything I needed.",
			Usage: fakevendor.Canonical{},
		},
	}
}

func Ch2Run(bin string) (*Ch2Result, error) {
	res := &Ch2Result{
		Session:     map[string]*VendorSession{},
		UsageProbe:  map[string]*VendorSession{},
		Render:      map[string]string{},
		RenderTwice: map[string]string{},
		RenderErr:   map[string]string{},
		Redacted:    map[string]string{},
		RedactedErr: map[string]string{},
	}

	// --- phase 1: Chapter 1 parity, using Chapter 1's own harness ----------
	if r1, err := Run(bin); err != nil {
		res.Ch1Err = err.Error()
	} else {
		res.Ch1 = Evaluate(r1)
	}

	work, err := os.MkdirTemp("", "ch02-grade-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(work)

	// --- phase 2: one live session per vendor ------------------------------
	for _, vendor := range Ch2Vendors {
		s, err := runVendorSession(bin, work, vendor)
		if err != nil {
			return nil, err
		}
		res.Session[vendor] = s
	}

	// --- phase 3: ephemera -------------------------------------------------
	eph, err := runEphemeraSession(bin, work)
	if err != nil {
		return nil, err
	}
	res.Ephemera = eph

	// --- phase 3b: cache-write probe ---------------------------------------
	for _, vendor := range Ch2Vendors {
		s, err := runUsageProbe(bin, work, vendor)
		if err != nil {
			return nil, err
		}
		res.UsageProbe[vendor] = s
	}

	// --- phase 4: render, twice, for each vendor ---------------------------
	exhibit := filepath.Join(work, "exhibit.log")
	if err := os.WriteFile(exhibit, []byte(ExhibitLog), 0o644); err != nil {
		return nil, err
	}
	redactLog := filepath.Join(work, "redacted.log")
	if err := os.WriteFile(redactLog, []byte(RedactionLog), 0o644); err != nil {
		return nil, err
	}
	for _, vendor := range Ch2Vendors {
		out, errOut, _ := runOnce(bin, work, vendorEnv(vendor, "", work), "render", exhibit)
		res.Render[vendor] = out
		if strings.TrimSpace(out) == "" {
			res.RenderErr[vendor] = errOut
		}
		out2, _, _ := runOnce(bin, work, vendorEnv(vendor, "", work), "render", exhibit)
		res.RenderTwice[vendor] = out2

		rout, rerr, _ := runOnce(bin, work, vendorEnv(vendor, "", work), "render", redactLog)
		res.Redacted[vendor] = rout
		if strings.TrimSpace(rout) == "" {
			res.RedactedErr[vendor] = rerr
		}
	}

	// --- phase 4b: replay per-call opaque material back to its own model ----
	// ExhibitLog cannot test this: its tool call carries no opaque material,
	// and its provenance is Anthropic, so a correct renderer would withhold
	// the material from Gemini by design and the assertion would pass
	// vacuously. This fixture is Gemini-authored so the same-model test says
	// yes and the signature must actually appear on the wire.
	thoughtLog := filepath.Join(work, "thought-replay.log")
	if err := os.WriteFile(thoughtLog, []byte(GeminiReplayLog), 0o644); err != nil {
		return nil, err
	}
	tout, terr, _ := runOnce(bin, work, vendorEnv("gemini", "", work), "render", thoughtLog)
	res.ThoughtReplay = tout
	if strings.TrimSpace(tout) == "" {
		res.ThoughtReplayErr = terr
	}

	// --- phase 5: Ref -------------------------------------------------------
	//
	// Driven under gemini for every fixture that carries a live blob: see the
	// note on Ch2Result's Ref fields.
	//
	// The three SERIALIZATION fixtures go through `dump`, not `render`, because
	// properties 3 and 4 are about the LOADER. Rendering them would let the
	// renderer's own kind switch refuse a malformed Ref that the loader had
	// happily accepted, and the check would pass for the wrong reason — the
	// loader bug would be invisible behind a renderer that caught it later.
	refDump := func(name, body string) (string, string, int) {
		p := filepath.Join(work, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return "", err.Error(), -1
		}
		env := append(vendorEnv("gemini", "", work), "CH02_LOG="+p)
		return runOnce(bin, work, env, "dump")
	}
	refRender := func(name, body string) (string, string, int) {
		p := filepath.Join(work, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			return "", err.Error(), -1
		}
		return runOnce(bin, work, vendorEnv("gemini", "", work), "render", p)
	}

	res.RefDump, res.RefDumpErr, _ = refDump("ref-roundtrip.log", RefRoundTripLog)
	res.RefOldFormatOut, res.RefOldFormatErr, res.RefOldFormatCode = refDump("ref-oldformat.log", RefOldFormatLog)
	res.RefZeroKindOut, res.RefZeroKindErr, res.RefZeroKindCode = refDump("ref-zerokind.log", RefZeroKindLog)

	res.RefRender, res.RefRenderErr, _ = refRender("ref-uri.log", RefURILog)
	res.RefRedacted, res.RefRedactedErr, _ = refRender("ref-redacted.log", RefRedactionLog)
	res.RefRedactedPlain, res.RefRedactedPlainErr, _ = refRender("ref-redacted-plain.log", RefRedactionPlainLog)

	if s := res.Session["anthropic"]; s != nil && strings.TrimSpace(s.DumpOut) != "" {
		rt := filepath.Join(work, "roundtrip.log")
		if err := os.WriteFile(rt, []byte(s.DumpOut), 0o644); err == nil {
			out, errOut, _ := runOnce(bin, work, vendorEnv("anthropic", "", work), "render", rt)
			res.RoundTripOut = out
			res.RoundTripErr = errOut
		}
	}

	return res, nil
}

func vendorEnv(vendor, baseURL, work string) []string {
	env := append(os.Environ(),
		"LLM_VENDOR="+vendor,
		"LLM_API_KEY=course-grader-fake",
		"LLM_MODEL="+ch2RequestedModel(vendor),
		"CH02_LOG="+filepath.Join(work, "ch02-"+vendor+".log"),
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

func ch2RequestedModel(vendor string) string {
	switch vendor {
	case "openai":
		return "gpt-5-course"
	case "gemini":
		return "gemini-3.5-flash-course"
	default:
		return "claude-sonnet-5-course"
	}
}

func runVendorSession(bin, work, vendor string) (*VendorSession, error) {
	fake := fakevendor.New(ch2Replies(vendor))
	defer fake.Close()

	s := &VendorSession{Vendor: vendor}
	env := vendorEnv(vendor, fake.URL(), work)

	lines := []string{
		`{"user":"read the configuration"}`,
		`{"user":"now summarize it"}`,
	}
	stdout, stderr, _ := runWithStdin(bin, work, env, nil, lines)
	s.Stderr = stderr
	parseCh2Stdout(s, stdout)
	s.Requests = fake.Requests()

	// `dump` runs in a FRESH process, reading only what was persisted.
	dumpOut, dumpErr, _ := runOnce(bin, work, env, "dump")
	s.DumpOut, s.DumpErr = dumpOut, dumpErr
	s.Log, s.LogErr = parseLogLines(dumpOut)
	return s, nil
}

// runUsageProbe exercises the cache-WRITE category in isolation.
//
// Only Anthropic and OpenAI report a cache-write token count at all. Gemini
// reports none anywhere in usageMetadata — the cost exists (Gemini bills cache
// storage by duration) but no token count is attached to any response. The
// honest canonical answer for Gemini is therefore zero, and the grader expects
// zero rather than an invented number.
func runUsageProbe(bin, work, vendor string) (*VendorSession, error) {
	fake := fakevendor.New([]fakevendor.Reply{{
		Text:           "Cached and ready.",
		Usage:          fakevendor.Canonical{Input: 7, CacheWrite: 300, CacheRead: 0, Output: 11},
		GeminiThoughts: 4,
	}})
	defer fake.Close()

	s := &VendorSession{Vendor: vendor}
	env := vendorEnv(vendor, fake.URL(), work)
	env = append(env, "CH02_LOG="+filepath.Join(work, "ch02-usage-"+vendor+".log"))

	stdout, stderr, _ := runWithStdin(bin, work, env, nil, []string{`{"user":"warm the cache"}`})
	s.Stderr = stderr
	parseCh2Stdout(s, stdout)
	s.Requests = fake.Requests()
	return s, nil
}

func runEphemeraSession(bin, work string) (*VendorSession, error) {
	fake := fakevendor.New(ch2Replies("anthropic"))
	defer fake.Close()

	s := &VendorSession{Vendor: "anthropic"}
	env := vendorEnv("anthropic", fake.URL(), work)
	env = append(env, "CH02_LOG="+filepath.Join(work, "ch02-ephemera.log"))

	lines := []string{
		`{"ephemeral":"CURRENT_TIME=2026-09-12T00:00:00Z SCREEN=terminal"}`,
		`{"user":"what time is it"}`,
		`{"user":"and again"}`,
	}
	stdout, stderr, _ := runWithStdin(bin, work, env, nil, lines)
	s.Stderr = stderr
	parseCh2Stdout(s, stdout)
	s.Requests = fake.Requests()
	return s, nil
}

// parseCh2Stdout reads the program's stdout, recording protocol violations
// rather than judging them.
func parseCh2Stdout(s *VendorSession, stdout string) {
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			s.Protocol = append(s.Protocol, fmt.Sprintf("stdout line is not JSON: %.80q", line))
			s.Extra = append(s.Extra, line)
			continue
		}
		switch {
		case m["assistant"] != nil:
			var v string
			_ = json.Unmarshal(m["assistant"], &v)
			s.Answers = append(s.Answers, v)
		case m["ack"] != nil:
			var v string
			_ = json.Unmarshal(m["ack"], &v)
			s.Protocol = append(s.Protocol, "ack:"+v)
		case m["usage"] != nil:
			var u Ch2Usage
			if err := json.Unmarshal(m["usage"], &u); err != nil {
				s.Protocol = append(s.Protocol, "usage line did not parse: "+err.Error())
				continue
			}
			s.Usage = &u
		case m["error"] != nil:
			var v string
			_ = json.Unmarshal(m["error"], &v)
			s.Protocol = append(s.Protocol, "program reported error: "+v)
		default:
			s.Extra = append(s.Extra, line)
		}
	}
}

// --- process plumbing ------------------------------------------------------

func runWithStdin(bin, dir string, env []string, args []string, lines []string) (string, string, int) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = env

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err.Error(), -1
	}
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Start(); err != nil {
		return "", err.Error(), -1
	}

	go func() {
		w := bufio.NewWriter(stdin)
		for _, l := range lines {
			w.WriteString(l)
			w.WriteByte('\n')
			w.Flush()
			time.Sleep(20 * time.Millisecond)
		}
		stdin.Close()
	}()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		code := 0
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
		return out.String(), errb.String(), code
	case <-time.After(Ch2Timeout):
		_ = cmd.Process.Kill()
		<-done
		return out.String(), errb.String() + "\n[grader] timed out", -1
	}
}

func runOnce(bin, dir string, env []string, args ...string) (string, string, int) {
	return runWithStdin(bin, dir, env, args, nil)
}
