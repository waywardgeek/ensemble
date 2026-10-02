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

// A credential provider is not merely a startup fact. attachCredentials used
// to install the ChatGPT plan provider only when OpenAI was the vendor in
// force at startup, so an operator who began on Anthropic and then switched
// with the model picker spent the rest of the session on their metered API
// key while believing they were spending a subscription they had already paid
// for. Nothing failed. The request succeeded and only the invoice disagreed,
// which is why this went unnoticed until the picker made it reachable in one
// click.
//
// This asserts the token actually presented and the route actually taken,
// never that eng.Creds is non-nil. A provider that is attached but never
// consulted satisfies the weaker check and still bills the wrong account.
func TestSetModelMovesCredentialsWithVendor(t *testing.T) {
	plan := &fakeCredentialProvider{token: "plan-bearer-token"}
	eng := NewEngine(common.Config{
		Model:     "claude-sonnet-5",
		Vendor:    common.VendorAnthropic,
		Surface:   common.DefaultSurface(common.VendorAnthropic),
		BaseURL:   "https://anthropic.test",
		APIKey:    "key-anthropic",
		Endpoints: endpoints(),
	}, filepath.Join(t.TempDir(), "journal.jsonl"), nil, nil, nil)
	eng.CredsByVendor = map[common.Vendor]common.CredentialProvider{
		common.VendorOpenAI: plan,
	}
	a := NewActor(eng, nil)

	// Anthropic is not the provider's vendor, so the metered key stands.
	cfg, err := eng.requestCfg()
	if err != nil {
		t.Fatalf("requestCfg() on the startup vendor: %v", err)
	}
	if cfg.APIKey != "key-anthropic" {
		t.Errorf("before the switch APIKey = %q, want key-anthropic", cfg.APIKey)
	}

	a.handleSetModel(common.SetModel{Model: "gpt-6.1-sol"})

	cfg, err = eng.requestCfg()
	if err != nil {
		t.Fatalf("requestCfg() after switching to OpenAI: %v", err)
	}
	if cfg.APIKey != "plan-bearer-token" {
		t.Errorf("after switching to OpenAI APIKey = %q, want the plan bearer token; "+
			"the metered key means the switch did not carry the provider", cfg.APIKey)
	}
	if !cfg.Route.RequiresStateless {
		t.Error("after switching to OpenAI the plan route was not selected")
	}

	// Switching away must not present a ChatGPT token to Anthropic's host.
	a.handleSetModel(common.SetModel{Model: "claude-sonnet-5"})

	cfg, err = eng.requestCfg()
	if err != nil {
		t.Fatalf("requestCfg() after switching back: %v", err)
	}
	if cfg.APIKey != "key-anthropic" {
		t.Errorf("after switching back APIKey = %q, want key-anthropic; a provider left "+
			"behind presents one vendor's bearer token to another vendor's host", cfg.APIKey)
	}
	if cfg.Route.RequiresStateless {
		t.Error("the plan route survived a switch away from OpenAI")
	}
}
