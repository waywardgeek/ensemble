package oauth

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// newSignedInProvider runs a real sign-in and returns a Provider over the
// resulting credential file.
func newSignedInProvider(t *testing.T, env *testEnv) *Provider {
	t.Helper()
	env.mustAuthorize(t)
	p, err := NewProvider(env.cfg)
	if err != nil {
		t.Fatalf("NewProvider(): %v", err)
	}
	return p
}

func TestProviderKind(t *testing.T) {
	env := newTestEnv(t)
	p, err := NewProvider(env.cfg)
	if err != nil {
		t.Fatalf("NewProvider(): %v", err)
	}
	if p.Kind() != common.CredentialChatGPTOAuth {
		t.Errorf("Kind() = %v, want CredentialChatGPTOAuth", p.Kind())
	}
}

func TestProviderNoCredentials(t *testing.T) {
	env := newTestEnv(t)
	p, err := NewProvider(env.cfg)
	if err != nil {
		t.Fatalf("NewProvider(): %v", err)
	}
	if _, err := p.GetBearerToken(context.Background()); !errors.Is(err, ErrNoCredentials) {
		t.Fatalf("GetBearerToken() = %v, want ErrNoCredentials", err)
	}
}

func TestProviderReturnsValidToken(t *testing.T) {
	env := newTestEnv(t)
	p := newSignedInProvider(t, env)

	tok, err := p.GetBearerToken(context.Background())
	if err != nil {
		t.Fatalf("GetBearerToken(): %v", err)
	}
	if tok == "" {
		t.Fatal("GetBearerToken() returned an empty token")
	}
	// The access token is long-lived here, so no refresh should have
	// happened: exactly one token call, the original code exchange.
	if env.issuer.calls() != 1 {
		t.Errorf("token endpoint called %d times, want 1 (no needless refresh)", env.issuer.calls())
	}
}

// TestScopeGateBlocksInference is the entitlement check. A valid ID token
// alone must not authorize plan usage.
func TestScopeGateBlocksInference(t *testing.T) {
	env := newTestEnv(t)
	// Everything except the one scope that authorizes billing.
	env.issuer.grantedScopes = "openid profile email offline_access resource.invoke"

	// The sign-in itself SUCCEEDS and is retained: the first-party guidance
	// is to disable plan usage, not to discard the credential and loop.
	creds := env.mustAuthorize(t)
	if creds.HasDirectTokenScope() {
		t.Fatal("HasDirectTokenScope() is true despite the scope being withheld")
	}
	if creds.Subject != testSubject {
		t.Error("the sign-in should still identify the user")
	}
	if !strings.Contains(env.out.String(), ScopeDirectTokens) {
		t.Errorf("the user was not told why plan usage is disabled; output:\n%s", env.out.String())
	}

	// But inference is blocked.
	p, err := NewProvider(env.cfg)
	if err != nil {
		t.Fatalf("NewProvider(): %v", err)
	}
	tok, err := p.GetBearerToken(context.Background())
	if !errors.Is(err, ErrPlanScopeNotGranted) {
		t.Fatalf("GetBearerToken() = (%q, %v), want ErrPlanScopeNotGranted", tok, err)
	}
	if tok != "" {
		t.Errorf("GetBearerToken() returned a token (%q) despite the missing scope", tok)
	}
}

// TestScopeGateUsesTokenResponseNotRequest: we request the scope every time,
// so a gate that consulted the REQUEST would always pass.
func TestScopeGateUsesTokenResponseNotRequest(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.grantedScopes = "openid profile email"

	creds := env.mustAuthorize(t)

	// Sanity: the request did ask for the scope.
	if !hasScope(splitScopes(env.browser.authQuery.Get("scope")), ScopeDirectTokens) {
		t.Fatal("precondition failed: the authorize request did not ask for the scope")
	}
	// The stored scopes must reflect the RESPONSE, not the request.
	if hasScope(creds.GrantedScopes, ScopeDirectTokens) {
		t.Errorf("stored granted scopes %v came from the request, not the token response", creds.GrantedScopes)
	}
}

// TestRefreshRotation covers the core rotation contract: all four rotating
// fields are replaced together and persisted.
func TestRefreshRotation(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1 // the first access token is already due a refresh
	env.mustAuthorize(t)

	before, err := env.cfg.Store.Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}
	// Make the refreshed token long-lived so the second call is a cache hit.
	env.issuer.expiresIn = 3600

	p, err := NewProvider(env.cfg)
	if err != nil {
		t.Fatalf("NewProvider(): %v", err)
	}
	tok, err := p.GetBearerToken(context.Background())
	if err != nil {
		t.Fatalf("GetBearerToken(): %v", err)
	}

	after, err := env.cfg.Store.Load()
	if err != nil {
		t.Fatalf("Load() after refresh: %v", err)
	}

	if tok == before.AccessToken {
		t.Error("GetBearerToken() returned the stale access token")
	}
	if after.AccessToken != tok {
		t.Errorf("stored access token %q != the one handed out %q", after.AccessToken, tok)
	}
	if after.RefreshToken == before.RefreshToken {
		t.Error("the refresh token was not rotated")
	}
	if after.RefreshToken == "" {
		t.Error("the refresh token was blanked out")
	}
	if !after.ExpiresAt.After(before.ExpiresAt) {
		t.Errorf("expiry did not advance: %s -> %s", before.ExpiresAt, after.ExpiresAt)
	}
	if !after.HasDirectTokenScope() {
		t.Errorf("granted scopes were lost on refresh: %v", after.GrantedScopes)
	}
	if after.ClientID != before.ClientID {
		t.Errorf("client id changed on refresh: %q -> %q", before.ClientID, after.ClientID)
	}

	// A second call must not refresh again.
	callsBefore := env.issuer.calls()
	if _, err := p.GetBearerToken(context.Background()); err != nil {
		t.Fatalf("second GetBearerToken(): %v", err)
	}
	if env.issuer.calls() != callsBefore {
		t.Errorf("a fresh token still triggered a refresh (%d -> %d calls)", callsBefore, env.issuer.calls())
	}
}

// TestRefreshIsProactive: a token that has not technically expired but will
// within the skew window must be refreshed before it is handed out.
func TestRefreshIsProactive(t *testing.T) {
	env := newTestEnv(t)
	// Valid for 60s, which is inside the default 2-minute skew.
	env.issuer.expiresIn = 60
	env.mustAuthorize(t)

	before, _ := env.cfg.Store.Load()
	if before.ExpiresAt.Before(time.Now()) {
		t.Fatal("precondition failed: the token should not be expired yet")
	}

	env.issuer.expiresIn = 3600
	p, _ := NewProvider(env.cfg)
	if _, err := p.GetBearerToken(context.Background()); err != nil {
		t.Fatalf("GetBearerToken(): %v", err)
	}
	if env.issuer.calls() < 2 {
		t.Errorf("token calls = %d; a token expiring inside the skew window was not refreshed proactively", env.issuer.calls())
	}
}

// TestTerminalRefreshErrors covers all six codes that require full
// re-authorization.
func TestTerminalRefreshErrors(t *testing.T) {
	terminal := []string{
		"invalid_grant",
		"invalid_refresh_token",
		"token_expired",
		"refresh_token_expired",
		"refresh_token_invalidated",
		"refresh_token_reused",
	}

	for _, code := range terminal {
		t.Run(code, func(t *testing.T) {
			env := newTestEnv(t)
			env.issuer.expiresIn = 1
			env.mustAuthorize(t)

			before, _ := env.cfg.Store.Load()
			env.issuer.refreshErrorCode = code

			p, _ := NewProvider(env.cfg)
			_, err := p.GetBearerToken(context.Background())

			if !errors.Is(err, ErrReauthRequired) {
				t.Fatalf("GetBearerToken() = %v, want ErrReauthRequired for %q", err, code)
			}
			if !IsTerminalRefreshError(err) {
				t.Errorf("IsTerminalRefreshError() = false for %q", code)
			}
			// The specific code must survive for diagnostics.
			var te *TokenError
			if !errors.As(err, &te) {
				t.Fatalf("error %v does not carry a *TokenError", err)
			}
			if te.Code != code {
				t.Errorf("TokenError.Code = %q, want %q", te.Code, code)
			}

			// Tokens cleared, identity retained: re-authorization must reuse
			// the issued client id rather than registering again.
			after, loadErr := env.cfg.Store.Load()
			if loadErr != nil {
				t.Fatalf("Load() after terminal error: %v", loadErr)
			}
			if after.AccessToken != "" || after.RefreshToken != "" {
				t.Error("dead tokens were left on disk after a terminal refresh error")
			}
			if after.ClientID != before.ClientID {
				t.Errorf("issued client id was discarded: %q, want %q", after.ClientID, before.ClientID)
			}
			if after.HostID != before.HostID {
				t.Errorf("host id was discarded: %q, want %q", after.HostID, before.HostID)
			}
		})
	}
}

// TestNonTerminalRefreshErrorPreservesTheGrant is the other half of the
// rule: a transient failure must NOT destroy a working refresh token.
func TestNonTerminalRefreshErrorPreservesTheGrant(t *testing.T) {
	transient := []struct {
		code   string
		status int
	}{
		{"temporarily_unavailable", 503},
		{"server_error", 500},
		{"slow_down", 429},
		{"", 503}, // a {"detail": ...} body with no machine-readable code
	}

	for _, tc := range transient {
		name := tc.code
		if name == "" {
			name = "unparseable body"
		}
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			env.issuer.expiresIn = 1
			env.mustAuthorize(t)
			before, _ := env.cfg.Store.Load()

			env.issuer.refreshErrorCode = tc.code
			env.issuer.refreshErrorStatus = tc.status
			if tc.code == "" {
				env.issuer.refreshErrorCode = "detail-only"
			}

			p, _ := NewProvider(env.cfg)
			_, err := p.GetBearerToken(context.Background())
			if err == nil {
				t.Fatal("GetBearerToken() succeeded despite a failing refresh")
			}
			if errors.Is(err, ErrReauthRequired) {
				t.Fatalf("a transient failure (%v) was treated as terminal", err)
			}

			after, loadErr := env.cfg.Store.Load()
			if loadErr != nil {
				t.Fatalf("Load(): %v", loadErr)
			}
			if after.RefreshToken != before.RefreshToken {
				t.Error("a transient failure destroyed the refresh token")
			}
		})
	}
}

// TestRefreshTokenReuseIsDetected exercises the real rotation hazard against
// the fake issuer's own reuse detection, rather than an injected error.
func TestRefreshTokenReuseIsDetected(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1
	env.mustAuthorize(t)

	stale, _ := env.cfg.Store.Load()

	// First provider refreshes successfully, rotating the token away.
	p1, _ := NewProvider(env.cfg)
	if _, err := p1.GetBearerToken(context.Background()); err != nil {
		t.Fatalf("first refresh: %v", err)
	}

	// Now replay the stale refresh token, as a stale copy of the credential
	// file would.
	if err := env.cfg.Store.Save(stale); err != nil {
		t.Fatalf("Save(stale): %v", err)
	}
	p2, _ := NewProvider(env.cfg)
	_, err := p2.GetBearerToken(context.Background())

	if !errors.Is(err, ErrReauthRequired) {
		t.Fatalf("replaying a rotated-away refresh token gave %v, want ErrReauthRequired", err)
	}
	var te *TokenError
	if errors.As(err, &te) && te.Code != "refresh_token_reused" {
		t.Errorf("TokenError.Code = %q, want refresh_token_reused", te.Code)
	}
}

// TestRefreshScopeDowngradeBlocksInference: the gate must consult the scopes
// from the LATEST token response, so a server-side revocation takes effect.
func TestRefreshScopeDowngradeBlocksInference(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1
	env.mustAuthorize(t)

	// The plan entitlement is revoked between sign-in and refresh.
	env.issuer.refreshScopes = "openid profile email offline_access"
	env.issuer.expiresIn = 3600

	p, _ := NewProvider(env.cfg)
	_, err := p.GetBearerToken(context.Background())
	if !errors.Is(err, ErrPlanScopeNotGranted) {
		t.Fatalf("GetBearerToken() = %v, want ErrPlanScopeNotGranted after a scope downgrade", err)
	}

	after, _ := env.cfg.Store.Load()
	if after.HasDirectTokenScope() {
		t.Error("the revoked scope is still recorded on disk")
	}
}

// TestRefreshOmittingScopeRetainsIt: RFC 6749 §5.1 says an omitted scope
// means unchanged. Blanking it would revoke the user's entitlement on our
// side for no reason.
func TestRefreshOmittingScopeRetainsIt(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1
	env.mustAuthorize(t)

	env.issuer.omitRefreshScope = true
	env.issuer.expiresIn = 3600

	p, _ := NewProvider(env.cfg)
	if _, err := p.GetBearerToken(context.Background()); err != nil {
		t.Fatalf("GetBearerToken(): %v", err)
	}
	after, _ := env.cfg.Store.Load()
	if !after.HasDirectTokenScope() {
		t.Errorf("an omitted scope field wiped the granted scopes: %v", after.GrantedScopes)
	}
}

// TestRefreshOmittingRefreshTokenRetainsIt: likewise, an absent rotated token
// means "keep using the one you have", not "you now have none".
func TestRefreshOmittingRefreshTokenRetainsIt(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1
	env.mustAuthorize(t)
	before, _ := env.cfg.Store.Load()

	env.issuer.omitRefreshToken = true
	env.issuer.expiresIn = 3600

	p, _ := NewProvider(env.cfg)
	if _, err := p.GetBearerToken(context.Background()); err != nil {
		t.Fatalf("GetBearerToken(): %v", err)
	}
	after, _ := env.cfg.Store.Load()
	if after.RefreshToken != before.RefreshToken {
		t.Errorf("refresh token changed to %q when the server sent none", after.RefreshToken)
	}
}

// TestProviderConcurrentUse is not a smoke test. Without serialisation, two
// goroutines racing to refresh would each spend the shared rotating refresh
// token, and the loser would be told refresh_token_reused — logging the user
// out. Run with -race.
func TestProviderConcurrentUse(t *testing.T) {
	env := newTestEnv(t)
	env.issuer.expiresIn = 1
	env.mustAuthorize(t)
	env.issuer.expiresIn = 3600

	p, _ := NewProvider(env.cfg)
	callsBefore := env.issuer.calls()

	const goroutines = 24
	var wg sync.WaitGroup
	errs := make([]error, goroutines)
	toks := make([]string, goroutines)
	start := make(chan struct{})

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // maximise the overlap
			toks[i], errs[i] = p.GetBearerToken(context.Background())
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: GetBearerToken() = %v", i, err)
		}
		if toks[i] != toks[0] {
			t.Errorf("goroutine %d got a different token than goroutine 0", i)
		}
	}

	// Exactly one refresh, no matter how many callers noticed the expiry.
	if got := env.issuer.calls() - callsBefore; got != 1 {
		t.Errorf("%d refresh calls for %d concurrent readers; want exactly 1", got, goroutines)
	}
}

// TestProviderSatisfiesInterface pins the seam: the engine must be able to
// hold this behind common.CredentialProvider without knowing what it is.
func TestProviderSatisfiesInterface(t *testing.T) {
	env := newTestEnv(t)
	p := newSignedInProvider(t, env)

	var cp common.CredentialProvider = p
	tok, err := cp.GetBearerToken(context.Background())
	if err != nil {
		t.Fatalf("GetBearerToken() through the interface: %v", err)
	}
	if tok == "" {
		t.Fatal("empty token through the interface")
	}
}

// TestCredentialsSnapshotIsACopy: a caller must not be able to mutate
// provider state, and must not see a rotation tear a struct in half.
func TestCredentialsSnapshotIsACopy(t *testing.T) {
	env := newTestEnv(t)
	p := newSignedInProvider(t, env)

	snap, err := p.Credentials()
	if err != nil {
		t.Fatalf("Credentials(): %v", err)
	}
	snap.AccessToken = "tampered"
	snap.GrantedScopes[0] = "tampered"

	again, err := p.Credentials()
	if err != nil {
		t.Fatalf("Credentials(): %v", err)
	}
	if again.AccessToken == "tampered" {
		t.Error("mutating the snapshot changed the provider's access token")
	}
	if again.GrantedScopes[0] == "tampered" {
		t.Error("the granted-scopes slice is shared with callers")
	}
}
