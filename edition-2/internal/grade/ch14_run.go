package grade

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// --- the five scenarios -------------------------------------------------
//
// Each entry here is one launch of the student's system. They are grouped by
// MODE rather than one per check: a launch costs seconds, and the streamed
// prose run feeds four checks at once for the price of one.

// ch14Result carries what each scenario observed.
type Ch14Result struct {
	HarnessOK  bool
	HarnessErr string

	DiscriminationOK  bool
	DiscriminationErr string

	FiltersOK  bool
	FiltersErr string

	BuffersOK  bool
	BuffersErr string

	UnstreamedOK  bool
	UnstreamedErr string

	GateOK  bool
	GateErr string

	ErrorsOK  bool
	ErrorsErr string

	BoundariesOK  bool
	BoundariesErr string

	IdentifiersOK  bool
	IdentifiersErr string

	Ch13Parity    bool
	Ch13ParityErr string
}

// Ch14Run grades chapter 14 by running the student's speech harness and
// reading the log it produces.
func Ch14Run(dir string) Ch14Result {
	var r Ch14Result

	script, err := ch14Discover(dir)
	if err != nil {
		r.HarnessErr = err.Error()
		return r
	}
	man := ch14Describe(script)

	// Scenario 1: streamed prose. One reply carrying markdown emphasis, a
	// fenced code block, a camelCase identifier, an all-caps word, and two
	// sentences. Feeds filters, buffers, boundaries and identifiers.
	prose := "Reading **now**. The HTTPServer handles RPGLit in ALL CAPS.\n\n" +
		"```go\nx := 1\n```\n\nThat is the whole file."
	entries, stderr, err := ch14Run(script, ch14Scenario{
		name:    "prose",
		prompt:  "describe the file",
		replies: []fakevendor.Reply{{Text: prose}},
	})
	if err != nil {
		r.HarnessErr = fmt.Sprintf("%v; harness output: %s", err, ch14Tail(stderr))
		return r
	}
	if len(ch14Spoken(entries)) == 0 {
		r.HarnessErr = "the harness ran but nothing was spoken; harness output: " + ch14Tail(stderr)
		return r
	}
	r.HarnessOK = true
	ch14GradeProse(&r, entries)

	// Scenario 2: the discrimination test. The same string arrives twice,
	// once as a tool result and once as the model's own words. One must be
	// silent and the other must be spoken.
	const secret = "ZEBRAFISH"
	ch14GradeDiscrimination(&r, script, man, secret)

	// Scenario 3: the reply arrives whole rather than in pieces.
	ch14GradeUnstreamed(&r, script, man)

	// Scenario 4: the model's endpoint fails. Silence here is the defect.
	ch14GradeErrors(&r, script)

	// Scenario 5: someone types while the agent is talking.
	ch14GradeGate(&r, script, man)

	return r
}

// ch14GradeProse reads the four properties visible in one streamed reply.
func ch14GradeProse(r *Ch14Result, entries []ch14Entry) {
	all := ch14AllSpeech(entries)
	spoken := ch14Spoken(entries)

	// Markup must not be read aloud, and a fenced block must not be
	// shattered into its punctuation.
	switch {
	case strings.Contains(all, "**"):
		r.FiltersErr = fmt.Sprintf("asterisks were spoken aloud: %q", all)
	case strings.Contains(all, "```"):
		r.FiltersErr = fmt.Sprintf("code fence punctuation was spoken aloud: %q", all)
	case strings.Contains(all, "x := 1"):
		r.FiltersErr = fmt.Sprintf("the body of a code block was read out: %q", all)
	default:
		r.FiltersOK = true
	}

	// Fragments must be assembled. A system that speaks every delta as it
	// arrives produces many short utterances and splits words.
	if len(spoken) == 0 {
		r.BuffersErr = "nothing was spoken"
	} else if len(spoken) > 12 {
		r.BuffersErr = fmt.Sprintf("speech arrived in %d pieces, which is delta-by-delta rather than assembled: %q", len(spoken), spoken)
	} else if !strings.Contains(all, "Reading") {
		r.BuffersErr = fmt.Sprintf("the opening words never arrived intact: %q", all)
	} else {
		r.BuffersOK = true
	}

	// Sentences are the unit a listener can follow, so two sentences must
	// not be welded into one utterance.
	if len(spoken) >= 2 {
		r.BoundariesOK = true
	} else {
		r.BoundariesErr = fmt.Sprintf("a multi-sentence reply was spoken as %d utterance(s); sentences are the unit a listener can interrupt: %q", len(spoken), spoken)
	}

	// Identifiers are read as words. An all-caps word is left alone rather
	// than spelled out letter by letter.
	switch {
	case strings.Contains(all, "HTTPServer"):
		r.IdentifiersErr = fmt.Sprintf("HTTPServer was never split into words: %q", all)
	case strings.Contains(all, "A L L") || strings.Contains(all, "C A P S"):
		r.IdentifiersErr = fmt.Sprintf("an all-caps word was spelled out letter by letter: %q", all)
	case !strings.Contains(all, "ALL CAPS"):
		r.IdentifiersErr = fmt.Sprintf("the all-caps words did not survive intact: %q", all)
	case !strings.Contains(all, "HTTP Server"):
		r.IdentifiersErr = fmt.Sprintf("HTTPServer was not spoken as separate words: %q", all)
	default:
		r.IdentifiersOK = true
	}
}

// ch14GradeDiscrimination is the check the rest of the chapter exists for.
//
// The same token arrives twice. Once as the content of a file the agent
// read, which the listener did not ask to hear, and once as the model's own
// sentence, which they did. Speak everything and the first half fails;
// speak nothing and the second half fails. Only telling them apart passes.
func ch14GradeDiscrimination(r *Ch14Result, script string, man ch14Manifest, secret string) {
	const chatToken = "PELICAN"
	target := "notes.txt"

	args, _ := json.Marshal(map[string]string{man.ReadFileArg: target})
	entries, stderr, err := ch14Run(script, ch14Scenario{
		name:   "discrimination",
		prompt: "read " + target + " and tell me what it says",
		plant:  map[string]string{target: "the file contains " + secret + " and nothing else\n"},
		replies: []fakevendor.Reply{
			{ToolName: man.ReadFileTool, ToolArgs: string(args), ToolID: "t1"},
			{Text: "The file mentions " + chatToken + " in its only line."},
		},
	})
	if err != nil {
		r.DiscriminationErr = fmt.Sprintf("%v; harness output: %s", err, ch14Tail(stderr))
		return
	}
	all := ch14AllSpeech(entries)

	if strings.Contains(all, secret) {
		r.DiscriminationErr = fmt.Sprintf("the contents of a file the agent read were spoken aloud; a listener hears every byte of every tool result: %q", all)
		return
	}
	if !strings.Contains(all, chatToken) {
		r.DiscriminationErr = fmt.Sprintf("the model's own sentence was never spoken, so the listener learned nothing: %q", all)
		return
	}
	r.DiscriminationOK = true
}

// ch14GradeUnstreamed covers the reply that arrives whole.
func ch14GradeUnstreamed(r *Ch14Result, script string, man ch14Manifest) {
	if !man.SupportsOff {
		r.UnstreamedErr = "the harness reports it cannot run with streaming off, so the case that arrives whole cannot be checked"
		return
	}
	const token = "MARMOSET"
	entries, stderr, err := ch14Run(script, ch14Scenario{
		name:      "unstreamed",
		prompt:    "answer briefly",
		streaming: "off",
		replies:   []fakevendor.Reply{{Text: "The answer is " + token + "."}},
	})
	if err != nil {
		r.UnstreamedErr = fmt.Sprintf("%v; harness output: %s", err, ch14Tail(stderr))
		return
	}
	all := ch14AllSpeech(entries)
	if !strings.Contains(all, token) {
		r.UnstreamedErr = fmt.Sprintf("with streaming off the reply was never spoken; a listener hears silence and assumes a crash: %q", all)
		return
	}
	r.UnstreamedOK = true
}

// ch14GradeErrors covers the failure a listener would otherwise never learn
// about.
func ch14GradeErrors(r *Ch14Result, script string) {
	entries, stderr, err := ch14Run(script, ch14Scenario{
		name:      "error",
		prompt:    "this will fail",
		brokenLLM: true,
		replies:   nil,
	})
	if err != nil {
		r.ErrorsErr = fmt.Sprintf("%v; harness output: %s", err, ch14Tail(stderr))
		return
	}
	if len(ch14Spoken(entries)) == 0 {
		r.ErrorsErr = "the model's endpoint failed and nothing was spoken; the listener is left waiting for a reply that will never come"
		return
	}
	r.ErrorsOK = true
}

// ch14GradeGate covers the half-built pause gate.
func ch14GradeGate(r *Ch14Result, script string, man ch14Manifest) {
	if !man.SupportsType {
		r.GateErr = "the harness reports it cannot type during a turn, so the pause gate cannot be checked"
		return
	}
	entries, stderr, err := ch14Run(script, ch14Scenario{
		name:     "gate",
		prompt:   "say something long",
		typeText: "wait",
		replies:  []fakevendor.Reply{{Text: "One. Two. Three. Four. Five. Six."}},
	})
	if err != nil {
		r.GateErr = fmt.Sprintf("%v; harness output: %s", err, ch14Tail(stderr))
		return
	}

	var paused, resumed, typingCause bool
	for _, e := range entries {
		switch e.Kind {
		case "pause":
			paused = true
			if strings.Contains(e.Cause, "typing") {
				typingCause = true
			}
		case "resume":
			if paused {
				resumed = true
			}
		}
	}
	switch {
	case !paused:
		r.GateErr = "typing never stopped the speech; the gate is not wired to anything"
	case !typingCause:
		r.GateErr = "speech paused but the log never says typing caused it, so the two reasons a gate closes cannot be told apart"
	case !resumed:
		r.GateErr = "speech paused and never resumed; a gate that only closes is worse than no gate"
	default:
		r.GateOK = true
	}
}

// ch14Tail keeps failure messages readable when a harness is noisy.
func ch14Tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 400 {
		return s
	}
	return "..." + s[len(s)-400:]
}
