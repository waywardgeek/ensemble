package llm

// The compressors.
//
// Four of them, one per rung of the ladder: a conversation span becomes a
// session memory, eight session memories become one 8x file, eight of those
// become one 64x file, and eight of those plus MEMORY.md become the curated
// long-term memory.
//
// They are one-shot LLM calls, deliberately. Each has bounded input, bounded
// output, no tools, and no memory of its own, so none of the machinery a real
// sub-agent needs would earn its place here. A later chapter upgrades these
// call sites to real sub-agents and uses them as the before-and-after.
//
// What makes them different from every other call in this codebase is that
// their output cannot be recomputed. Run one twice and you get two different
// memories, both defensible. That is exactly why the result is recorded as an
// event carrying bytes: replay applies the decision this compressor made, and
// never asks it to decide again.
//
// They answer with a submit tool rather than with prose. An agent asked for a
// summary will sooner or later reply "I have written the summary" instead of
// writing it, and that sentence would be saved as the memory.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// submitToolName is the only tool a compressor is given.
const submitToolName = "submit"

func submitDecl(what string) common.ToolDecl {
	return common.ToolDecl{
		Name: submitToolName,
		Description: "Submit the finished " + what + ". Call this exactly once, with the " +
			"complete text as the memory argument. Do not reply with prose: the memory " +
			"argument is what gets stored, and anything you say outside it is discarded.",
		Schema: json.RawMessage(`{
  "type": "object",
  "properties": {
    "memory": {
      "type": "string",
      "description": "The complete compressed text, ready to store verbatim."
    }
  },
  "required": ["memory"]
}`),
	}
}

// compressorPrompt is the instruction every rung shares.
//
// First person is not a stylistic flourish. The distant past is precisely
// what gets written by a compressor rather than by the agent, and a memory
// that reads "the agent investigated the settings path" is a report about
// somebody else. "I found six dead settings fields" is a memory. Since the
// whole promise of compaction is that it leaves you as you, third person
// would reintroduce the discontinuity the design exists to remove.
func compressorPrompt(what string, budget int) string {
	return fmt.Sprintf(`You are compressing %s into a smaller record that you will read later.

Write in the FIRST PERSON, as yourself. Say "I found", never "the agent found":
this is your own memory, not a report about somebody else.

Keep, at full precision:
  - specific facts, numbers and measurements
  - file paths, line numbers, commit hashes, command names
  - decisions made, and what was rejected and why
  - anything that surprised you, and anything that failed

Drop:
  - narrative rationale that regenerates from what remains
  - restatements, pleasantries, and descriptions of what you were about to do

Aim for about %d bytes. Going somewhat under is fine. Losing a measurement to
hit the number is not: reasoning regenerates, measurements rot.

Call the submit tool exactly once with the finished text.`, what, budget)
}

// compress runs one compressor and returns the text it submitted.
func (e *Engine) compress(what, body string, budget int) (string, error) {
	renderer, parser, err := SeamFor(e.Cfg.Vendor)
	if err != nil {
		return "", err
	}

	// A throwaway context holding only the material to compress. The
	// compressor sees what it was given and nothing newer, so it cannot
	// wander into conversation outside the range it was asked about.
	ctx := common.NewContext()
	ctx.Dialogue = append(ctx.Dialogue, common.Entry{
		Seq:   1,
		Actor: common.ActorHuman,
		Kind:  common.KindDialogue,
		Parts: common.PartList{common.TextPart{Text: body}},
	})

	cfg := e.Cfg
	cfg.SystemPrompt = compressorPrompt(what, budget)
	cfg.Tools = []common.ToolDecl{submitDecl(what)}
	cfg.DisableStreaming = true

	req, err := renderer.Render(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("rendering compressor request: %w", err)
	}
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("compressor call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("compressor call: %s: %s", resp.Status, bytes.TrimSpace(slurp))
	}

	var parts common.PartList
	err = parser.Parse(resp, common.StreamCallbacks{
		OnPartFinal: func(_ uint64, p common.Part) { parts = append(parts, p) },
	})
	if err != nil {
		return "", fmt.Errorf("parsing compressor response: %w", err)
	}

	for _, p := range parts {
		call, ok := p.(common.ToolCallPart)
		if !ok || call.Name != submitToolName {
			continue
		}
		var args struct {
			Memory string `json:"memory"`
		}
		if err := json.Unmarshal(call.Args, &args); err != nil {
			return "", fmt.Errorf("compressor submitted malformed arguments: %w", err)
		}
		if strings.TrimSpace(args.Memory) == "" {
			return "", fmt.Errorf("compressor submitted an empty memory")
		}
		return args.Memory, nil
	}
	// No submit call. Deliberately an error rather than a fallback to the
	// prose it did write: storing "I have written the summary" as the memory
	// is the exact failure the submit tool exists to prevent, and a fallback
	// would reintroduce it while looking like robustness.
	return "", fmt.Errorf("compressor answered without calling %s", submitToolName)
}

// renderConversation turns a span of dialogue into the text a compressor
// reads.
//
// It renders what the agent itself saw: the speakers and their words. Tool
// calls and results are already gone from a checkpointed span, and where they
// are not, they are left out here — a compressor reading raw tool output
// would spend its budget on transcript instead of meaning.
func renderConversation(c *common.Context, from, to common.Seq) string {
	var b strings.Builder
	for _, e := range c.Dialogue {
		if e.Kind != common.KindDialogue || e.Seq < from || e.Seq > to {
			continue
		}
		for _, p := range e.Parts {
			t, ok := p.(common.TextPart)
			if !ok || strings.TrimSpace(t.Text) == "" {
				continue
			}
			fmt.Fprintf(&b, "%s: %s\n\n", e.Actor, t.Text)
		}
	}
	return b.String()
}
