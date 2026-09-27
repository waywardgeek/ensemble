package llm

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Judge is the model call behind snippet selection. It implements
// common.SnippetJudge.
//
// The single most important property of this type is that it is STATELESS.
// Every Pick builds a throwaway context containing exactly one user message —
// the prompt it was handed — and throws that context away when the call
// returns. The judge holds no Dialogue, accumulates nothing across calls, and
// is never told that it is part of an agent.
//
// That is not an optimisation, it is what makes recall safe to run on every
// single turn. A judge that kept its own history would grow without bound in
// the one place nobody is watching, and would eventually need recall of its
// own to manage the context it had accumulated while deciding what to recall.
// Statelessness is what stops that regress before it starts: there is no
// second conversation here, only a function that happens to be evaluated by a
// language model.
//
// It follows that the judge sees the conversation ONLY as data inside its
// prompt — as quoted text it is asked to reason about, never as messages it
// participated in. The caller decides how much of the conversation to quote.
type Judge struct {
	engine *Engine
}

// NewJudge returns a judge that makes its calls through the given engine's
// HTTP client and vendor configuration.
func NewJudge(e *Engine) *Judge {
	return &Judge{engine: e}
}

// judgeMaxReply caps how much of the judge's answer is read. The expected
// reply is a handful of numbers; anything beyond this is a model that has
// started explaining itself, and the parser only wants the digits.
const judgeMaxReply = 4096

// Pick sends one stateless request and returns the raw reply text.
//
// Parsing is deliberately NOT done here. This method's only job is to get
// bytes back from a model; deciding what those bytes mean — including
// deciding that they are garbage — belongs to the caller, which is the
// component that knows what it asked for and what it should do when the
// answer makes no sense.
func (j *Judge) Pick(prompt string) (string, error) {
	e := j.engine
	if e == nil {
		return "", fmt.Errorf("judge has no engine")
	}
	renderer, parser, err := SeamFor(e.Cfg.Vendor)
	if err != nil {
		return "", err
	}

	// The throwaway context. One entry, one part, no history. Built fresh on
	// every call and dropped on return.
	ctx := common.NewContext()
	ctx.Dialogue = append(ctx.Dialogue, common.Entry{
		Seq:   1,
		Actor: common.ActorHuman,
		Kind:  common.KindDialogue,
		Parts: common.PartList{common.TextPart{Text: prompt}},
	})

	cfg := e.Cfg
	cfg.SystemPrompt = judgeSystemPrompt
	// No tools. The judge is being asked a question, not being given a job,
	// and a judge holding the agent's toolbelt could act on the conversation
	// it was only supposed to read. Zero tools is also what distinguishes a
	// judge call from a turn on the wire.
	cfg.Tools = nil
	cfg.DisableStreaming = true

	req, err := renderer.Render(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("rendering judge request: %w", err)
	}
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("judge call: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("judge call: %s: %s", resp.Status, bytes.TrimSpace(slurp))
	}

	var sb strings.Builder
	err = parser.Parse(resp, common.StreamCallbacks{
		OnPartFinal: func(_ uint64, p common.Part) {
			if t, ok := p.(common.TextPart); ok && sb.Len() < judgeMaxReply {
				sb.WriteString(t.Text)
			}
		},
	})
	if err != nil {
		return "", fmt.Errorf("parsing judge response: %w", err)
	}
	return sb.String(), nil
}

// judgeSystemPrompt frames the judge as a filter rather than an assistant.
//
// The instruction to answer with numbers alone is a request, not a guarantee.
// Models add preamble, apologise, wrap answers in prose, and occasionally
// refuse. The parser on the other side assumes none of this worked.
const judgeSystemPrompt = `You are a relevance filter. You are shown a question and a numbered list of snippets retrieved from an archive.

Your only job is to decide which snippets would genuinely help answer the question.

Reply with the numbers of the useful snippets, separated by commas, in the order you consider most useful. Reply with the single word NONE if no snippet is useful.

Do not explain. Do not summarise the snippets. Do not answer the question itself. Numbers only.

Most retrieved snippets are not useful. Discarding all of them is a correct and common answer.`
