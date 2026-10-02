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
// against. Every model used by an earlier chapter's grader keeps the surface
// it had, and only rows that actually name the new one get it.
func TestCh19SurfaceIsPerModelNotPerVendor(t *testing.T) {
	cases := []struct {
		model string
		want  common.Surface
		why   string
	}{
		{"gpt-6.1-sol", common.SurfaceResponses, "the new default, measured against the live endpoint"},
		{"gpt-ch19-course", common.SurfaceResponses, "the grader's stand-in for it"},

		{"gpt-5-course", common.SurfaceChatCompletions, "chapter 2's grader model"},
		{"gpt-6-astra", common.SurfaceChatCompletions, "chapter 18's grader model"},
		{"gpt-5.6-sol", common.SurfaceChatCompletions, "not exercised on the new surface"},

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
