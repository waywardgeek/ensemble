package llm

// Wiring tests for the surface seam.
//
// The migration's risk is not that the new surface fails; it is that the new
// surface quietly captures models that were never tested on it. These tests
// pin both directions.

import (
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

func TestCh19SeamForDispatchesOnSurface(t *testing.T) {
	cases := []struct {
		name    string
		vendor  common.Vendor
		surface common.Surface
		want    common.Renderer
		wantErr bool
	}{
		{"openai responses", common.VendorOpenAI, common.SurfaceResponses, responsesSeam{}, false},
		{"openai chat", common.VendorOpenAI, common.SurfaceChatCompletions, openAISeam{}, false},
		{"anthropic", common.VendorAnthropic, common.SurfaceMessages, anthropicSeam{}, false},
		{"gemini", common.VendorGemini, common.SurfaceGenerateContent, geminiSeam{}, false},

		// Refuse rather than guess. An unset surface on a vendor that serves
		// two dialects is a configuration bug, and picking one sends a
		// well-formed request to the wrong endpoint.
		{"openai unset surface", common.VendorOpenAI, 0, nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, p, err := SeamFor(tc.vendor, tc.surface)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("got seam %T, want an error", r)
				}
				return
			}
			if err != nil {
				t.Fatalf("SeamFor: %v", err)
			}
			if r != tc.want {
				t.Errorf("renderer = %T, want %T", r, tc.want)
			}
			if p == nil {
				t.Error("nil parser")
			}
		})
	}
}

// The migration must not drag models onto an endpoint they were never tested
// against. A row earns the new surface by being measured on it, never by
// sharing a vendor with something that was; an unmeasured model keeps what it
// had, and a model with no row at all falls back to the vendor default.
//
// Chapter 18's grader model and the Chat Completions thinking fixture are now
// course rows rather than production ones. That is the mechanism by which an
// earlier chapter's grader keeps seeing exactly the wire it was graded on
// while a production row is free to record a later measurement. Pinning a
// grader to a production row makes its premise expire the day the vendor ships
// a better endpoint for that model.
func TestCh19SurfaceIsPerModelNotPerVendor(t *testing.T) {
	cases := []struct {
		model string
		want  common.Surface
		why   string
	}{
		{"gpt-6.1-sol", common.SurfaceResponses, "the new default, measured against the live endpoint"},
		{"gpt-ch19-course", common.SurfaceResponses, "the grader's stand-in for it"},
		{"gpt-6-astra", common.SurfaceResponses, "measured on the live plan route: thinking, tools and streamed summaries"},
		{"gpt-5.6-sol", common.SurfaceResponses, "measured likewise, and Chat Completions refuses it tools alongside a reasoning budget"},

		{"gpt-5-course", common.SurfaceChatCompletions, "the grader model for chapters 2 and 18, present in their frozen snapshots"},
		{"gpt-5.6-course", common.SurfaceChatCompletions, "the fixture carrying the tools-plus-reasoning restriction"},

		// A model with no row at all still resolves, via the vendor default.
		{"gpt-unknown-future", common.SurfaceChatCompletions, "unknown model falls back to the vendor default"},
	}

	for _, tc := range cases {
		got := common.SurfaceForModel(tc.model, common.VendorOpenAI)
		if got != tc.want {
			t.Errorf("SurfaceForModel(%q) = %v, want %v (%s)", tc.model, got, tc.want, tc.why)
		}
	}

	// Other vendors are untouched by any of this.
	if got := common.SurfaceForModel("claude-sonnet-5", common.VendorAnthropic); got != common.SurfaceMessages {
		t.Errorf("anthropic surface = %v, want SurfaceMessages", got)
	}
}
