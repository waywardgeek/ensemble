package gradeedition2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

const streamThought = "The caller wants the configuration file. I should call the tool before answering."
const streamAnswer = "The configuration sets the listen address to port 8080."

// Retain only content the existing observation envelopes expose in both
// implementations. Complete opaque provider replay remains a source-test duty;
// inventing a new wire field here would change the exercise being graded.
type streamPart struct {
	kind, text, name string
	args             any
}
type streamRun struct {
	exec            grade.Ch7Exec
	finals          map[uint64]streamPart
	content         []streamPart
	usage           map[string]float64
	requests        []fakevendor.Recorded
	streamRequested bool
}

func streamReplies() []fakevendor.Reply {
	return []fakevendor.Reply{
		{Thinking: streamThought, Text: "Let me take a look at that configuration for you.", ToolName: "run_command", ToolArgs: `{"command":"printf STREAM_TOOL_RESULT"}`, ToolID: "call_ch7_1", Usage: fakevendor.Canonical{Input: 120, Output: 64}},
		// Repeat the part layout: text uses the same provider ID in both
		// responses. Whole-turn ID maps would conflate two valid text parts.
		{Thinking: streamThought, Text: streamAnswer, Usage: fakevendor.Canonical{Input: 190, Output: 24}},
	}
}

func driveStream(bin, model string, plain, external bool) streamRun {
	run := streamRun{exec: grade.Ch7Exec{FinalText: map[uint64]string{}, FinalTool: map[uint64]string{}}, finals: map[uint64]streamPart{}, usage: map[string]float64{}}
	work, err := os.MkdirTemp("", "edition2-stream-*")
	if err != nil {
		run.exec.ExitErr = err.Error()
		return run
	}
	defer os.RemoveAll(work)
	peer := fakevendor.New(streamReplies())
	defer peer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	if external {
		cmd.Args = append(cmd.Args, "what does the configuration say?")
	}
	cmd.Dir = work
	disable, exerciseDisable := "0", "0"
	if plain {
		if external {
			exerciseDisable = "1"
		} else {
			disable = "1"
		}
	}
	log := filepath.Join(work, "agent.jsonl")
	cmd.Env = append(os.Environ(), "LLM_VENDOR=anthropic", "LLM_MODEL="+model, "LLM_RESOLVED_MODEL=", "LLM_API_KEY=fake-key", "LLM_BASE_URL="+peer.URL(), "EN_DISABLE_STREAMING="+disable, "CH07_NO_STREAM="+exerciseDisable, "CH02_LOG="+log, "CH07_AGENT_LOG="+log, "CH07_LOG="+filepath.Join(work, "observations.jsonl"))
	cmd.Stdin = strings.NewReader("{\"kind\":\"prompt\",\"text\":\"what does the configuration say?\"}\n")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	run.exec.Ran = true
	if err := cmd.Run(); err != nil {
		run.exec.ExitErr = fmt.Sprintf("%v: %s", err, strings.TrimSpace(stderr.String()))
		return run
	}
	run.exec.Stdout, run.exec.Stderr = out.String(), stderr.String()
	run.requests = peer.Requests()
	for _, request := range run.requests {
		var body map[string]any
		json.Unmarshal(request.Body, &body)
		run.streamRequested = run.streamRequested || body["stream"] == true
	}
	records, err := jsonLines(out.Bytes())
	if err != nil {
		run.exec.ExitErr = err.Error()
		return run
	}
	// Job-result finals are not provider parts. The neutral envelope labels them;
	// the legacy envelope flattens them to text, so consult their existing durable
	// event identity instead of guessing from content or a repeated part ID.
	toolSeq := map[float64]bool{}
	if data, err := os.ReadFile(log); err == nil {
		if events, err := jsonLines(data); err == nil {
			for _, e := range events {
				if normalize(fmt.Sprint(e["type"])) == "toolreturned" {
					toolSeq[number(e["seq"])] = true
				}
			}
		}
	}
	response := -1
	type key struct {
		response int
		id       uint64
	}
	ids := map[key]uint64{}
	// Dense IDs are comparison aliases only. They let the original evaluator's
	// ID-keyed maps represent (response, part) without changing production IDs
	// or requiring the later GUI's persistent response-identity protocol.
	mapped := func(id uint64) uint64 {
		k := key{response, id}
		if n, ok := ids[k]; ok {
			return n
		}
		n := uint64(len(ids) + 1)
		ids[k] = n
		return n
	}
	for _, m := range records {
		run.exec.Lines = append(run.exec.Lines, m)
		if text, ok := m["assistant"].(string); ok {
			run.exec.Assistant = text
		}
		if u, ok := m["usage"].(map[string]any); ok {
			for k, v := range u {
				run.usage[k] = number(v)
			}
		}
		if m["observation"] == "stream_stats" {
			run.exec.Stats = append(run.exec.Stats, m)
		}
		if external {
			continue
		} // Aggregate-only applications need no delta protocol.
		switch m["observation"] {
		case "state_changed":
			// Both historical actors emit this before starting the next HTTP response.
			// Do not infer a new response from an accidental repeated or wrong part ID.
			if normalize(fmt.Sprint(m["to"])) == "inflight" {
				response++
			}
		case "part_delta":
			if response < 0 {
				run.exec.ExitErr = "delta precedes request state boundary"
				return run
			}
			// Zero is a valid provider block index; an absent ID is not zero.
			rawID, valid := m["part_id"].(float64)
			if !valid || rawID < 0 || rawID != float64(uint64(rawID)) {
				run.exec.ExitErr = "observation needs a nonnegative integer part_id"
				return run
			}
			id := mapped(uint64(rawID))
			run.exec.Deltas = append(run.exec.Deltas, grade.Ch7Delta{PartID: id, Kind: fmt.Sprint(m["kind"]), Chunk: fmt.Sprint(m["chunk"]), Resp: response})
		case "part_final":
			part := object(m["part"])
			if part["type"] == "tool_result" || toolSeq[number(m["seq"])] {
				continue
			}
			if response < 0 {
				run.exec.ExitErr = "final precedes request state boundary"
				return run
			}
			rawID, valid := m["part_id"].(float64)
			if !valid || rawID < 0 || rawID != float64(uint64(rawID)) {
				run.exec.ExitErr = "observation needs a nonnegative integer part_id"
				return run
			}
			id := mapped(uint64(rawID))
			if _, exists := run.finals[id]; exists {
				run.exec.ExitErr = "provider part finalized twice in one response"
				return run
			}
			final := streamPart{kind: "thinking"}
			// Typed parts take precedence over generic observation labels.
			var textValue any
			if part["type"] == "text" {
				textValue = part["text"]
			} else if len(part) == 0 {
				textValue = m["text"]
			}
			if text, ok := textValue.(string); ok {
				final.kind, final.text = "text", text
				run.exec.FinalText[id] = text
			}
			var nameValue any
			if part["type"] == "tool_call" {
				nameValue = part["name"]
			} else if len(part) == 0 {
				nameValue = m["tool"]
			}
			if name, ok := nameValue.(string); ok {
				final.kind, final.name = "tool_call", name
				run.exec.FinalTool[id] = name
				final.args = part["args"]
				if final.args == nil {
					if raw, ok := m["args"].(string); ok {
						json.Unmarshal([]byte(raw), &final.args)
					} else {
						final.args = m["args"]
					}
				}
			}
			run.finals[id] = final
			run.content = append(run.content, final)
			run.exec.Finals = append(run.exec.Finals, grade.Ch7PartFinal{PartID: id, Kind: final.kind})
		}
	}
	return run
}
