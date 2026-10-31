package gradeedition2

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter6 preserves the original streaming weights with declared-model fixtures
// and corrected parity. Stream delivery is measured through the existing CLI;
// the separate public example remains an aggregate observer, not another CLI.
func Chapter6(bin, core, consumer, workflow, exercise, model string, example ConsumerCase, media map[string]string) (grade.Report, error) {
	previous, err := Chapter5(bin, core, consumer, workflow, example, media)
	if err != nil {
		return grade.Report{}, err
	}
	checks := streamingChecks(bin, core, exercise, model)
	parity := grade.Check{ID: "ch5-parity", Title: "All corrected Chapter 5 checks pass", Points: 20, Earned: 20, Passed: true}
	for _, c := range previous.Checks {
		if !c.Passed {
			fail(&parity, "%s: %s", c.ID, strings.Join(c.Details, "; "))
		}
	}
	checks = append(checks, parity)
	return grade.NewTitledReport("Edition 2, Chapter 6 — Streaming", checks, ""), nil
}

// Chapter6Executable grades compiled C# applications through their public wire
// protocols. Plain C# responses publish finals, not synthetic transient deltas.
// It deliberately excludes inherited Go consumer/topology and Chapter 5 checks.
func Chapter6Executable(bin, observer, model string) grade.Report {
	checks := streamingChecksFor(bin, "", observer, model, true)
	return grade.NewTitledReport("Chapter 6 — executable streaming only (80 points; excludes Go checks and inherited Chapter 5)", checks, "")
}

func streamingChecks(bin, core, exercise, model string) []grade.Check {
	return streamingChecksFor(bin, core, exercise, model, false)
}

func streamingChecksFor(bin, core, exercise, model string, finalOnly bool) []grade.Check {
	stream, plain := driveStream(bin, model, false, false), driveStream(bin, model, true, false)
	// Reuse the original five numerical/content checks. Their private launcher
	// hard-codes the contradictory unknown-model fixture, so only driving and
	// response scoping are replaced; the inherited minima retain their meaning.
	original := grade.Ch7Evaluate(&grade.Ch7Result{BuildOK: true, Stream: stream.exec, Plain: plain.exec})
	checks := original[:5] // Chapter6 replaces the original inherited parity.
	if stream.exec.ExitErr == "" {
		correlateStream(&checks[1], stream, "text")
		correlateStream(&checks[2], stream, "thinking")
		correlateStream(&checks[3], stream, "tool_call")
	}
	if finalOnly {
		// Replace only the legacy plain-delta contract, not its streaming checks.
		checks[4] = grade.Check{ID: "delivery-not-content", Title: "Streaming and final-only delivery retain content", Points: 10, Earned: 10, Passed: true}
		c := &checks[4]
		if stream.exec.ExitErr != "" || plain.exec.ExitErr != "" {
			fail(c, "stream error=%s plain error=%s", stream.exec.ExitErr, plain.exec.ExitErr)
		}
		if stream.exec.Assistant != streamAnswer || plain.exec.Assistant != streamAnswer {
			fail(c, "both modes must deliver the fixture's completed answer")
		}
		if len(plain.exec.Deltas) != 0 || len(stream.exec.Deltas) == 0 {
			fail(c, "expected real stream chunks and no synthetic plain chunks")
		}
	}
	delivery := &checks[4]
	if stream.exec.ExitErr == "" && plain.exec.ExitErr == "" {
		if !reflect.DeepEqual(stream.content, plain.content) {
			fail(delivery, "observable finalized content differs between streaming and plain")
		}
		for label, run := range map[string]streamRun{"stream": stream, "plain": plain} {
			// A correct first request cannot excuse a later plain request. Checking
			// each body also separates transport selection from reported counters.
			for _, request := range run.requests {
				var body map[string]any
				json.Unmarshal(request.Body, &body)
				if (body["stream"] == true) != (label == "stream") {
					fail(delivery, "%s request streaming flag disagrees with selected delivery", label)
				}
			}
			// These are the two fake replies' measured totals, not a comparison
			// that could bless equal invented counts in both delivery modes.
			if run.usage["input"] != 310 || run.usage["output"] != 88 || run.usage["cache_read"] != 0 || run.usage["cache_write"] != 0 {
				fail(delivery, "%s usage must retain measured input310/output88 and zero cache counts, got %v", label, run.usage)
			}
			if len(run.requests) != 2 {
				fail(delivery, "%s expected two provider requests, got %d", label, len(run.requests))
			} else if text, isError, err := anthropicToolText(run.requests[1].Body, "call_ch7_1"); err != nil || isError || !strings.Contains(text, "STREAM_TOOL_RESULT") {
				fail(delivery, "%s did not execute the complete tool call: %v error=%t text=%q", label, err, isError, text)
			}
		}
	}
	// The absent unknown fixture must remain usable without acquiring a hidden
	// stream-capability row merely to satisfy a test. It is a distinct control.
	expectedPlainText := 2
	if finalOnly {
		expectedPlainText = 0
	}
	unknown := driveStream(bin, "unknown-stream-model-edition2", false, false)
	if unknown.exec.ExitErr != "" || unknown.exec.Assistant != streamAnswer || unknown.streamRequested || unknown.exec.KindCount("text") != expectedPlainText {
		fail(delivery, "unknown model must complete in plain mode: error=%s stream=%t text-deltas=%d", unknown.exec.ExitErr, unknown.streamRequested, unknown.exec.KindCount("text"))
	}
	if finalOnly {
		checkShapeExecutable(delivery, exercise, model, true)
	} else {
		shapeExercise(delivery, core, exercise, model)
	}
	delivery.Details = append(delivery.Details, "compares observable finals and measured usage, unknown-model fallback and separate aggregate exercise; gated flush, SSE framing and complete opaque replay remain required source/integration proof")
	return checks
}

func correlateStream(c *grade.Check, run streamRun, kind string) {
	grouped := map[uint64]string{}
	for _, d := range run.exec.Deltas {
		if d.Kind == kind {
			grouped[d.PartID] += d.Chunk
		}
	}
	// Check both directions: a plausible first response must not hide a final
	// in the second response for which every delta was silently dropped.
	for id, final := range run.finals {
		if _, ok := grouped[id]; final.kind == kind && !ok {
			fail(c, "%s final part %d has no matching delta group", kind, id)
		}
	}
	for id, text := range grouped {
		final, ok := run.finals[id]
		if !ok || final.kind != kind {
			fail(c, "%s part %d has no corresponding final of that kind", kind, id)
			continue
		}
		switch kind {
		case "text":
			if text != final.text {
				fail(c, "text part %d differs from its final", id)
			}
		case "thinking":
			// Some original envelopes expose only the opaque final's identity. The
			// known provider content can still be checked without inventing a field.
			if text != streamThought {
				fail(c, "thinking part %d lost or changed provider content", id)
			}
		case "tool_call":
			// Names arrive whole, JSON arrives in fragments. Compare decoded arguments
			// so canonical key order/whitespace is not confused with lost content.
			brace := strings.Index(text, "{")
			var args any
			if brace < 0 || strings.TrimSpace(text[:brace]) != final.name || json.Unmarshal([]byte(text[brace:]), &args) != nil || !reflect.DeepEqual(args, final.args) {
				fail(c, "tool part %d name/arguments do not match its finalized call", id)
			}
		}
	}
}

func shapeExercise(c *grade.Check, core, dir, model string) {
	if err := publicDependency(core, dir); err != nil {
		fail(c, "stream exercise dependency: %v", err)
		return
	}
	work, err := os.MkdirTemp("", "edition2-shape-*")
	if err != nil {
		fail(c, "%v", err)
		return
	}
	defer os.RemoveAll(work)
	bin := filepath.Join(work, "shape")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		fail(c, "stream exercise build: %v: %s", err, out)
		return
	}
	checkShapeExecutable(c, bin, model, false)
}

func checkShapeExecutable(c *grade.Check, bin, model string, finalOnly bool) {
	// Supply the public positional prompt to the external application; retain
	// the stdin envelope for consumers that use it. Aggregate reporting does
	// not require the external application to implement the core CLI protocol.
	on, off := driveStream(bin, model, false, true), driveStream(bin, model, true, true)
	for label, run := range map[string]streamRun{"stream": on, "plain": off} {
		if run.exec.ExitErr != "" {
			fail(c, "%s external exercise: %s", label, run.exec.ExitErr)
			continue
		}
		if run.exec.Assistant != streamAnswer || len(run.exec.Stats) == 0 {
			fail(c, "%s external exercise lacks actual final answer/shape report", label)
			continue
		}
		stats := run.exec.Stats[len(run.exec.Stats)-1]
		if finalOnly {
			parts := object(stats["parts"])
			if number(parts["text"]) != 2 || number(parts["thinking"]) != 2 || number(parts["tool_call"]) != 1 {
				fail(c, "%s external finalized parts differ from two-response fixture: %v", label, parts)
			}
			if run.usage["input"] != 310 || run.usage["output"] != 88 || number(stats["steps"]) != 2 {
				fail(c, "%s external must report measured usage and two model steps", label)
			}
		}
		// Presence and nonnegative units are observable here. Accurate timing and
		// final-part count reports still need source review, not a wall-clock race.
		if n, ok := stats["first_delta_ms"].(float64); !ok || n < 0 {
			fail(c, "%s external exercise needs nonnegative first_delta_ms", label)
		}
	}
	if len(on.exec.Stats) > 0 && len(off.exec.Stats) > 0 {
		a := object(on.exec.Stats[len(on.exec.Stats)-1]["deltas"])
		b := object(off.exec.Stats[len(off.exec.Stats)-1]["deltas"])
		for kind, counts := range map[string][2]float64{"text": {8, 2}, "thinking": {4, 2}, "tool_call": {3, 1}} {
			plainCount := counts[1]
			if finalOnly {
				plainCount = 0
			}
			if number(a[kind]) < counts[0] || number(b[kind]) != plainCount {
				fail(c, "external exercise %s delta shape: enabled=%v disabled=%v", kind, a[kind], b[kind])
			}
		}
	}
}

func number(v any) float64 { n, _ := v.(float64); return n }
