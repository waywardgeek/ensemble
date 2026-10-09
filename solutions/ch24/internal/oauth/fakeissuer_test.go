package oauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// This file builds a fake authorization server good enough that the
// production code cannot tell it from the real one. That standard matters:
// a fake that accepts anything proves only that the client can talk to a
// fake that accepts anything. In particular this one generates a real RSA
// key pair, publishes a real JWKS, signs real RS256 ID tokens, recomputes
// the S256 challenge from the verifier the client sends, and rotates refresh
// tokens with reuse detection.

// testKey is generated once for the whole package. RSA key generation is
// slow enough (tens of milliseconds) that doing it per-test-case would
// dominate the suite's runtime for no added coverage.
var (
	testKeyOnce sync.Once
	testKey     *rsa.PrivateKey
	// testOtherKey is a DIFFERENT valid key, never published in the JWKS. It
	// is how the forged-signature tests produce a token that is structurally
	// perfect and cryptographically wrong.
	testOtherKey *rsa.PrivateKey
)

func keys(t *testing.T) (*rsa.PrivateKey, *rsa.PrivateKey) {
	t.Helper()
	testKeyOnce.Do(func() {
		var err error
		testKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		testOtherKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
	})
	return testKey, testOtherKey
}

const (
	testKID          = "test-signing-key-1"
	testIssuedClient = "client_issued_7f3a"
	testSubject      = "user_abc123"
	testEmail        = "user@example.com"
	allScopes        = "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct"
)

// codeGrant is an authorization code the fake issuer has handed out.
type codeGrant struct {
	challenge string
	nonce     string
	clientID  string
	redeemed  bool
}

// refreshState tracks a rotating refresh token.
type refreshState struct {
	clientID string
	scopes   string
	used     bool
}

// fakeIssuer is an httptest-backed OIDC issuer.
type fakeIssuer struct {
	t   *testing.T
	srv *httptest.Server

	signKey *rsa.PrivateKey // key published in the JWKS

	mu        sync.Mutex
	codes     map[string]*codeGrant
	refreshes map[string]*refreshState
	seq       int

	// ---- knobs the tests turn ----

	// grantedScopes is what the token endpoint reports it granted. Defaults
	// to allScopes.
	grantedScopes string
	// refreshScopes, if set, is reported on refresh instead of grantedScopes.
	refreshScopes string
	// expiresIn is the access token lifetime in seconds.
	expiresIn int64
	// idTokenSigner overrides the key used to SIGN id tokens, without
	// changing the key published in the JWKS. Setting it to testOtherKey
	// yields a forged token.
	idTokenSigner *rsa.PrivateKey
	// idTokenClaims mutates the claim set just before signing.
	idTokenClaims func(m map[string]any)
	// idTokenHeader mutates the JOSE header just before signing.
	idTokenHeader func(m map[string]string)
	// omitIDToken drops the id_token from the token response.
	omitIDToken bool
	// refreshErrorCode, if set, makes every refresh fail with this code.
	refreshErrorCode string
	// refreshErrorStatus is the HTTP status paired with refreshErrorCode.
	refreshErrorStatus int
	// omitRefreshScope drops the scope field from refresh responses.
	omitRefreshScope bool
	// omitRefreshToken drops the rotated refresh token from the response.
	omitRefreshToken bool

	// ---- observations the tests assert on ----
	lastTokenForm url.Values
	tokenCalls    int
}

// calls returns the number of token-endpoint calls, under the lock.
//
// The HTTP handler runs on the server's goroutine while the test asserts from
// its own, so these reads need the same mutex the writes use. Without it the
// race detector flags the suite — and it would be right to.
func (f *fakeIssuer) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokenCalls
}

// tokenForm returns the most recent token-endpoint form, under the lock.
func (f *fakeIssuer) tokenForm() url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastTokenForm
}

func newFakeIssuer(t *testing.T) *fakeIssuer {
	t.Helper()
	k, _ := keys(t)
	f := &fakeIssuer{
		t:             t,
		signKey:       k,
		codes:         make(map[string]*codeGrant),
		refreshes:     make(map[string]*refreshState),
		grantedScopes: allScopes,
		expiresIn:     3600,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", f.handleDiscovery)
	mux.HandleFunc("/.well-known/jwks.json", f.handleJWKS)
	mux.HandleFunc("/api/accounts/oauth/token", f.handleToken)
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

// issuerURL is both the httptest base URL and the value the discovery
// document must claim, since FetchDiscovery insists the two agree.
func (f *fakeIssuer) issuerURL() string { return f.srv.URL }

func (f *fakeIssuer) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	doc := map[string]any{
		"issuer":                                f.srv.URL,
		"authorization_endpoint":                f.srv.URL + "/api/accounts/authorize",
		"token_endpoint":                        f.srv.URL + "/api/accounts/oauth/token",
		"revocation_endpoint":                   f.srv.URL + "/api/accounts/oauth/revoke",
		"jwks_uri":                              f.srv.URL + "/.well-known/jwks.json",
		"code_challenge_methods_supported":      []string{"S256"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      []string{"openid", "profile", "email", "offline_access"},
	}
	writeJSON(w, http.StatusOK, doc)
}

func (f *fakeIssuer) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	pub := f.signKey.Public().(*rsa.PublicKey)
	writeJSON(w, http.StatusOK, map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"alg": "RS256",
			"kid": testKID,
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}},
	})
}

// newCode registers an authorization code bound to a PKCE challenge, a nonce
// and a client id — exactly the binding a real issuer keeps.
func (f *fakeIssuer) newCode(challenge, nonce, clientID string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seq++
	code := fmt.Sprintf("authcode_%d", f.seq)
	f.codes[code] = &codeGrant{challenge: challenge, nonce: nonce, clientID: clientID}
	return code
}

func (f *fakeIssuer) newRefreshToken(clientID, scopes string) string {
	f.seq++
	tok := fmt.Sprintf("refresh_%d", f.seq)
	f.refreshes[tok] = &refreshState{clientID: clientID, scopes: scopes}
	return tok
}

func (f *fakeIssuer) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeTokenError(w, http.StatusBadRequest, "invalid_request", "unparseable form")
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.tokenCalls++
	f.lastTokenForm = r.PostForm

	// A public client must not be sending a secret. Asserting it here means
	// the test fails if the production code ever starts sending one.
	if r.PostForm.Get("client_secret") != "" {
		writeTokenError(w, http.StatusBadRequest, "invalid_client", "this client is public; no secret expected")
		return
	}

	switch r.PostForm.Get("grant_type") {
	case "authorization_code":
		f.handleAuthCodeGrant(w, r.PostForm)
	case "refresh_token":
		f.handleRefreshGrant(w, r.PostForm)
	default:
		writeTokenError(w, http.StatusBadRequest, "unsupported_grant_type", "")
	}
}

func (f *fakeIssuer) handleAuthCodeGrant(w http.ResponseWriter, form url.Values) {
	grant, ok := f.codes[form.Get("code")]
	if !ok {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "unknown authorization code")
		return
	}
	if grant.redeemed {
		// Authorization codes are single-use. Enforcing it keeps the fake
		// honest about why PKCE exists.
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "authorization code already redeemed")
		return
	}
	if form.Get("client_id") != grant.clientID {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "client_id does not match the code")
		return
	}

	// Recompute the S256 challenge from the verifier the client supplied.
	// This is the whole point of PKCE and the fake must actually do it,
	// otherwise a client that sent the wrong verifier — or none — would pass.
	verifier := form.Get("code_verifier")
	if verifier == "" {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "missing code_verifier")
		return
	}
	sum := sha256.Sum256([]byte(verifier))
	if base64.RawURLEncoding.EncodeToString(sum[:]) != grant.challenge {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "PKCE verifier does not match the challenge")
		return
	}
	grant.redeemed = true

	f.writeTokens(w, grant.clientID, grant.nonce, f.grantedScopes, true, false)
}

func (f *fakeIssuer) handleRefreshGrant(w http.ResponseWriter, form url.Values) {
	if f.refreshErrorCode != "" {
		status := f.refreshErrorStatus
		if status == 0 {
			status = http.StatusBadRequest
		}
		writeTokenError(w, status, f.refreshErrorCode, "injected by the test")
		return
	}

	tok := form.Get("refresh_token")
	state, ok := f.refreshes[tok]
	if !ok {
		writeTokenError(w, http.StatusBadRequest, "invalid_refresh_token", "unknown refresh token")
		return
	}
	if state.used {
		// Reuse detection: a rotating issuer treats a second presentation of
		// the same token as evidence of theft and kills the chain.
		writeTokenError(w, http.StatusBadRequest, "refresh_token_reused", "this refresh token was already redeemed")
		return
	}
	if form.Get("client_id") != state.clientID {
		writeTokenError(w, http.StatusBadRequest, "invalid_grant", "client_id does not match the refresh token")
		return
	}
	state.used = true

	scopes := state.scopes
	if f.refreshScopes != "" {
		scopes = f.refreshScopes
	}
	f.writeTokens(w, state.clientID, "", scopes, !f.omitRefreshToken, true)
}

// writeTokens mints a token response, including a real signed ID token.
func (f *fakeIssuer) writeTokens(w http.ResponseWriter, clientID, nonce, scopes string, withRefresh, isRefresh bool) {
	now := time.Now()
	resp := map[string]any{
		"access_token": fmt.Sprintf("access_%d", f.seq+1000),
		"token_type":   "Bearer",
		"expires_in":   f.expiresIn,
	}
	f.seq++

	if withRefresh {
		resp["refresh_token"] = f.newRefreshToken(clientID, scopes)
	}
	if !(isRefresh && f.omitRefreshScope) {
		resp["scope"] = scopes
	}

	if !f.omitIDToken {
		claims := map[string]any{
			"iss":   f.srv.URL,
			"sub":   testSubject,
			"aud":   clientID,
			"exp":   now.Add(time.Hour).Unix(),
			"iat":   now.Unix(),
			"email": testEmail,
			"name":  "Test User",
		}
		if nonce != "" {
			claims["nonce"] = nonce
		}
		if f.idTokenClaims != nil {
			f.idTokenClaims(claims)
		}
		header := map[string]string{"alg": "RS256", "typ": "JWT", "kid": testKID}
		if f.idTokenHeader != nil {
			f.idTokenHeader(header)
		}
		signer := f.signKey
		if f.idTokenSigner != nil {
			signer = f.idTokenSigner
		}
		resp["id_token"] = signJWT(f.t, signer, header, claims)
	}

	writeJSON(w, http.StatusOK, resp)
}

// signJWT produces a real compact JWS. When the header says alg:none it
// emits an empty signature, which is what an attacker's forged token looks
// like.
func signJWT(t *testing.T, key *rsa.PrivateKey, header map[string]string, claims map[string]any) string {
	t.Helper()
	hb, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshalling JWT header: %v", err)
	}
	cb, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshalling JWT claims: %v", err)
	}
	signing := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(cb)

	if header["alg"] == "none" {
		return signing + "."
	}
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatalf("signing JWT: %v", err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeTokenError(w http.ResponseWriter, status int, code, desc string) {
	writeJSON(w, status, map[string]string{"error": code, "error_description": desc})
}
