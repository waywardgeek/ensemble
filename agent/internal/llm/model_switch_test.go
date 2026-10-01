package llm

import (
	"path/filepath"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// endpoints is a stand in for what the composition root resolves, with values
// distinctive enough that a stale endpoint is obvious in a failure message.
func endpoints() map[common.Vendor]common.Endpoint {
	return map[common.Vendor]common.Endpoint{
		common.VendorAnthropic: {BaseURL: "https://anthropic.test", APIKey: "key-anthropic"},
		common.VendorOpenAI:    {BaseURL: "https://openai.test", APIKey: "key-openai"},
		common.VendorGemini:    {BaseURL: "https://gemini.test", APIKey: "key-gemini"},
	}
}

// A model switch is not one field. Rendering and redaction follow Cfg.Model on
// their own because they look features up per call, but Vendor, Surface and
// the endpoint are separate config fields. Move the model without them and the
// next request is rendered in one vendor's dialect and posted to another
// vendor's host, carrying that vendor's key, which fails at the API with an
// error that points nowhere near the cause.
//
// This asserts the endpoint itself rather than Vendor alone. Vendor is only a
// proxy for where the request goes, and an earlier version of this test
// checked the proxy while the real request still went to the previous host.
func TestSetModelMovesEndpointWithModel(t *testing.T) {
	eng := NewEngine(common.Config{
		Model:     "claude-sonnet-5",
		Vendor:    common.VendorAnthropic,
		Surface:   common.DefaultSurface(common.VendorAnthropic),
		BaseURL:   "https://anthropic.test",
		APIKey:    "key-anthropic",
		Endpoints: endpoints(),
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	a := NewActor(eng, nil)

	a.handleSetModel(common.SetModel{Model: "gpt-5.6-sol"})

	if got := eng.Cfg.Model; got != "gpt-5.6-sol" {
		t.Errorf("Model = %q, want gpt-5.6-sol", got)
	}
	if got, want := eng.Cfg.Vendor, common.VendorOpenAI; got != want {
		t.Errorf("Vendor = %v, want %v", got, want)
	}
	if got, want := eng.Cfg.Surface, common.DefaultSurface(common.VendorOpenAI); got != want {
		t.Errorf("Surface = %v, want %v", got, want)
	}
	if got, want := eng.Cfg.BaseURL, "https://openai.test"; got != want {
		t.Errorf("BaseURL = %q, want %q: the model moved but the request still goes to the old host", got, want)
	}
	if got, want := eng.Cfg.APIKey, "key-openai"; got != want {
		t.Errorf("APIKey = %q, want %q: the old vendor's key would be sent to the new vendor", got, want)
	}
}

// An unknown model must be refused, not guessed at. A prefix table would
// happily resolve this one to Anthropic and dial the wrong API.
func TestSetModelRefusesUnknownModel(t *testing.T) {
	eng := NewEngine(common.Config{
		Model:     "claude-sonnet-5",
		Vendor:    common.VendorAnthropic,
		Surface:   common.DefaultSurface(common.VendorAnthropic),
		BaseURL:   "https://anthropic.test",
		APIKey:    "key-anthropic",
		Endpoints: endpoints(),
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

// A vendor with no configured endpoint is refused outright. The tempting
// alternative is to move the model and leave the endpoint alone, which is
// precisely the bug: the agent would report the new model while every request
// went to the old vendor. Nothing may change, so the operator sees a failed
// switch rather than a silently misdirected one.
func TestSetModelRefusesVendorWithNoEndpoint(t *testing.T) {
	eng := NewEngine(common.Config{
		Model:   "claude-sonnet-5",
		Vendor:  common.VendorAnthropic,
		Surface: common.DefaultSurface(common.VendorAnthropic),
		BaseURL: "https://anthropic.test",
		APIKey:  "key-anthropic",
		Endpoints: map[common.Vendor]common.Endpoint{
			common.VendorAnthropic: {BaseURL: "https://anthropic.test", APIKey: "key-anthropic"},
		},
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	a := NewActor(eng, nil)

	a.handleSetModel(common.SetModel{Model: "gpt-5.6-sol"})

	if got := eng.Cfg.Model; got != "claude-sonnet-5" {
		t.Errorf("Model = %q, want the switch refused and claude-sonnet-5 kept", got)
	}
	if got, want := eng.Cfg.BaseURL, "https://anthropic.test"; got != want {
		t.Errorf("BaseURL = %q, want %q", got, want)
	}
	if got, want := eng.Cfg.APIKey, "key-anthropic"; got != want {
		t.Errorf("APIKey = %q, want %q", got, want)
	}
}
