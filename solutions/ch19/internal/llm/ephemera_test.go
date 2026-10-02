package llm

import (
	"io"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// renderBody renders one request through a vendor seam and returns the whole
// body as text. Asserting against the entire body, rather than against the
// field each vendor happens to put ephemera in, is deliberate: the question is
// whether the content reaches the wire at all, and a test that checks a
// particular field would pass if a renderer smuggled it somewhere else.
func renderBody(t *testing.T, seam common.Renderer, model, ephemeral string) string {
	t.Helper()
	c := &common.Context{
		Dialogue: []common.Entry{
			{Seq: 1, Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: "hello"}}},
		},
		Ephemera: common.PartList{common.TextPart{Text: ephemeral}},
	}
	cfg := common.Config{Model: model, BaseURL: "https://vendor.test", APIKey: "key"}
	req, err := seam.Render(c, cfg)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(raw)
}

// Anthropic's Opus 5.5 tolerates no change to a prefix it has already seen,
// and ephemera are a change by construction. Such a model must never be sent
// them.
//
// Every seam is checked against the same model even though that model is
// Anthropic's. The pairing is artificial on purpose: the drop is a property of
// the model rather than of a vendor, and rendering all three this way proves
// no renderer reached past the gate and read Ephemera directly. That is the
// drift this test exists to catch, since each vendor renders ephemera in its
// own shape and the obvious way to add a fourth is to copy one of them.
func TestEphemeraDroppedForModelThatCannotTakeThem(t *testing.T) {
	const needle = "EPHEMERAL-NEEDLE-42"

	seams := []struct {
		name string
		seam common.Renderer
	}{
		{"anthropic", anthropicSeam{}},
		{"openai", openAISeam{}},
		{"gemini", geminiSeam{}},
	}

	for _, s := range seams {
		t.Run(s.name+"/dropped for claude-opus-5-5", func(t *testing.T) {
			body := renderBody(t, s.seam, "claude-opus-5-5", needle)
			if strings.Contains(body, needle) {
				t.Errorf("ephemeral content reached the wire for a model that cannot take it:\n%s", body)
			}
		})

		t.Run(s.name+"/kept for a model that can take them", func(t *testing.T) {
			body := renderBody(t, s.seam, "claude-opus-5", needle)
			if !strings.Contains(body, needle) {
				t.Errorf("ephemeral content was dropped for a model that accepts it:\n%s", body)
			}
		})
	}
}

// The flag is negative so that the zero value, which every other row takes,
// means the model is fine. A positive flag would have made every model in the
// table broken by default and every new row broken until someone remembered.
func TestOnlyMeasuredModelsDeclareNoEphemera(t *testing.T) {
	if f, ok := common.LookupModel("claude-opus-5-5"); !ok || !f.NoEphemera {
		t.Errorf("claude-opus-5-5 must declare NoEphemera: looked up %v, found %v", ok, f.NoEphemera)
	}
	if f, ok := common.LookupModel("claude-opus-5"); !ok || f.NoEphemera {
		t.Errorf("claude-opus-5 must not declare NoEphemera: looked up %v, found %v", ok, f.NoEphemera)
	}
}
