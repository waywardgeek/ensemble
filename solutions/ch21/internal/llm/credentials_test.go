package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// The credential seam, tested through the thing that proves it works: the
// Authorization header on a rendered request.
//
// It would be easier to assert that requestCfg copied a string into a struct
// field. That test would pass while the header was built from something else
// entirely, because the field being set is a proxy for the behaviour and not
// the behaviour. The question the chapter actually cares about is whether a
// static key and a rotating OAuth token arrive at the vendor looking
// identical, and only the rendered request can answer it.

// fakeCredentialProvider stands in for the OAuth provider. It counts calls so
// a test can ask when the credential was resolved, which is the difference
// between resolving once at startup and resolving per request.
type fakeCredentialProvider struct {
	token string
	err   error
	calls int
}

func (p *fakeCredentialProvider) GetBearerToken(context.Context) (string, error) {
	p.calls++
	if p.err != nil {
		return "", p.err
	}
	return p.token, nil
}

func (p *fakeCredentialProvider) Kind() common.CredentialKind {
	return common.CredentialChatGPTOAuth
}

func credEngine(creds common.CredentialProvider, configured string) *Engine {
	return &Engine{
		Cfg: common.Config{
			BaseURL:      "https://example.invalid",
			Model:        "gpt-5.6-sol",
			Vendor:       common.VendorOpenAI,
			APIKey:       configured,
			SystemPrompt: "you are a helpful agent",
		},
		Creds: creds,
	}
}

// TestBothCredentialKindsProduceTheSameHeader is the chapter's central claim:
// the code that builds the request cannot tell the two apart.
func TestBothCredentialKindsProduceTheSameHeader(t *testing.T) {
	cases := []struct {
		name  string
		creds common.CredentialProvider
		want  string
	}{
		{"static api key", common.NewAPIKeyProvider("sk-static-key"), "Bearer sk-static-key"},
		{"oauth access token", &fakeCredentialProvider{token: "oauth-access-token"}, "Bearer oauth-access-token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The configured key is deliberately wrong. If the renderer reads
			// configuration instead of the resolved credential, the header
			// shows it.
			e := credEngine(tc.creds, "sk-configured-and-stale")
			cfg, err := e.requestCfg()
			if err != nil {
				t.Fatalf("requestCfg: %v", err)
			}
			req, err := (openAISeam{}).Render(ctxWith(workEntries(1, "alpha")), cfg)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if got := req.Header.Get("Authorization"); got != tc.want {
				t.Errorf("Authorization = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestNilProviderUsesConfiguredKey locks in the behaviour of every chapter
// before this one. The cross-chapter sweep depends on it: eighteen graders
// build an Engine as a struct literal with no provider at all, and all of them
// must keep working.
func TestNilProviderUsesConfiguredKey(t *testing.T) {
	e := credEngine(nil, "sk-from-config")
	cfg, err := e.requestCfg()
	if err != nil {
		t.Fatalf("requestCfg: %v", err)
	}
	req, err := (openAISeam{}).Render(ctxWith(workEntries(1, "alpha")), cfg)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got, want := req.Header.Get("Authorization"), "Bearer sk-from-config"; got != want {
		t.Errorf("Authorization = %q, want %q", got, want)
	}
}

// TestRequestCfgDoesNotMutateEngineConfig guards the copy.
//
// Writing the resolved token back onto e.Cfg would work, right up to the point
// where someone reads e.Cfg.APIKey expecting the operator's configuration and
// finds an access token that expires in forty minutes. The bug is invisible
// until something persists or displays that field.
func TestRequestCfgDoesNotMutateEngineConfig(t *testing.T) {
	e := credEngine(&fakeCredentialProvider{token: "ephemeral-token"}, "sk-configured")
	if _, err := e.requestCfg(); err != nil {
		t.Fatalf("requestCfg: %v", err)
	}
	if got, want := e.Cfg.APIKey, "sk-configured"; got != want {
		t.Errorf("e.Cfg.APIKey = %q after resolution, want it untouched at %q", got, want)
	}
}

// TestCredentialResolvedPerRequest proves the resolution is not hoisted to
// startup. An OAuth token that is fetched once and reused forever is a 401
// waiting for the hour to elapse.
func TestCredentialResolvedPerRequest(t *testing.T) {
	p := &fakeCredentialProvider{token: "tok"}
	e := credEngine(p, "")
	for i := 0; i < 3; i++ {
		if _, err := e.requestCfg(); err != nil {
			t.Fatalf("requestCfg %d: %v", i, err)
		}
	}
	if p.calls != 3 {
		t.Errorf("provider consulted %d times for 3 requests, want 3", p.calls)
	}
}

// TestCredentialFailureIsReported checks that a provider error stops the
// request rather than sending an empty bearer token and letting the vendor
// explain the problem as a 401.
func TestCredentialFailureIsReported(t *testing.T) {
	want := errors.New("refresh failed: invalid_grant")
	e := credEngine(&fakeCredentialProvider{err: want}, "sk-configured")
	if _, err := e.requestCfg(); err == nil {
		t.Fatal("requestCfg returned nil error when the provider failed")
	} else if !errors.Is(err, want) {
		t.Errorf("error %v does not wrap the provider's error %v", err, want)
	}
}
