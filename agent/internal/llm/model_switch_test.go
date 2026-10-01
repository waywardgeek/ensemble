package llm

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// A model switch is not one field. Rendering and redaction follow Cfg.Model
// on their own because they look features up per call, but Vendor and Surface
// are separate config fields. Move the model without them and the next
// request is rendered in one vendor's dialect and posted to another vendor's
// endpoint, which fails at the API with an error that points nowhere near the
// cause.
func TestSetModelMovesVendorAndSurfaceTogether(t *testing.T) {
	eng := NewEngine(common.Config{
		Model:   "claude-sonnet-5",
		Vendor:  common.VendorAnthropic,
		Surface: common.DefaultSurface(common.VendorAnthropic),
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	a := NewActor(eng, nil)

	a.handleSetModel(common.SetModel{Model: "gpt-5.6-sol"})

	if got := eng.Cfg.Model; got != "gpt-5.6-sol" {
		t.Errorf("Model = %q, want gpt-5.6-sol", got)
	}
	if got, want := eng.Cfg.Vendor, common.VendorOpenAI; got != want {
		t.Errorf("Vendor = %v, want %v: the model moved but the endpoint did not", got, want)
	}
	if got, want := eng.Cfg.Surface, common.DefaultSurface(common.VendorOpenAI); got != want {
		t.Errorf("Surface = %v, want %v", got, want)
	}
}

// An unknown model must be refused, not guessed at. A prefix table would
// happily resolve this one to Anthropic and dial the wrong API.
func TestSetModelRefusesUnknownModel(t *testing.T) {
	eng := NewEngine(common.Config{
		Model:   "claude-sonnet-5",
		Vendor:  common.VendorAnthropic,
		Surface: common.DefaultSurface(common.VendorAnthropic),
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	a := NewActor(eng, nil)

	a.handleSetModel(common.SetModel{Model: "claude-not-a-real-model"})

	if got := eng.Cfg.Model; got != "claude-sonnet-5" {
		t.Errorf("Model = %q, want the switch refused and claude-sonnet-5 kept", got)
	}
	if got := eng.Cfg.Vendor; got != common.VendorAnthropic {
		t.Errorf("Vendor = %v, want it left alone on a refused switch", got)
	}
}
