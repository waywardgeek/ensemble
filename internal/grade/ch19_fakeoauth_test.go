package grade

// ch19_fakeoauth_test.go — self-tests for the fake OAuth provider.
//
// The student's agent does not exist yet, so the only way to know this fixture
// grades correctly is to drive it with a client written here, in this file,
// that is known to be correct. Two directions matter and both are covered
// below:
//
//   - A correct client must leave Violations EMPTY. A fixture that flags
//     something on a clean run fails every student, and because failures look
//     like student mistakes nobody investigates for a long time.
//   - An incorrect client must be caught. Each negative test breaks exactly
//     one thing and asserts that exactly that thing is reported.
//
// The signature test is the one that cannot be skipped. It rebuilds an
// rsa.PublicKey from the "n" and "e" of the served JWK and verifies the
// id_token against it. If that passes, the key the fixture publishes really is
// the key it signs with, and the UnsignedKeyMismatch probe is therefore
// meaningful rather than vacuous.

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// A deliberately correct in-test OAuth client
// ---------------------------------------------------------------------------

type ch19AuthClient struct {
	t           *testing.T
	hc          *http.Client
	md          ch19OIDCMetadata
	verifier    string
	challenge   string
	state       string
	nonce       string
	redirectURI string
	clientID    string
}

// ch19AuthNewClient builds a client that does everything right. Negative tests
// then reach in and break one field, which keeps "what is being tested" to a
// single visible line per test.
func ch19AuthNewClient(t *testing.T, s *ch19FakeOAuth) *ch19AuthClient {
	t.Helper()
	// 32 random bytes is what RFC 7636 is asking for; in base64url that is 43
	// characters, the minimum legal verifier length.
	verifier := ch19OAuthRandomToken(32)
	sum := sha256.Sum256([]byte(verifier))
	c := &ch19AuthClient{
		t: t,
		// Redirects must NOT be followed: the callback points at a loopback
		// port with no listener, and following it would turn every test into a
		// connection-refused error instead of a Location header to inspect.
		hc: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		verifier:    verifier,
		challenge:   base64.RawURLEncoding.EncodeToString(sum[:]),
		state:       ch19OAuthRandomToken(16),
		nonce:       ch19OAuthRandomToken(16),
		redirectURI: "http://127.0.0.1:54321/auth/callback",
		clientID:    ch19OAuthBootstrapClientID,
	}
	c.md = c.discover(s)
	return c
}

func (c *ch19AuthClient) discover(s *ch19FakeOAuth) ch19OIDCMetadata {
	c.t.Helper()
	resp, err := c.hc.Get(s.URL() + "/.well-known/openid-configuration")
	if err != nil {
		c.t.Fatalf("discovery: %v", err)
	}
	defer resp.Body.Close()
	var md ch19OIDCMetadata
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		c.t.Fatalf("decoding discovery document: %v", err)
	}
	return md
}

// authorizeValues is the fully correct authorization request.
func (c *ch19AuthClient) authorizeValues() url.Values {
	v := url.Values{}
	v.Set("client_id", c.clientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", c.redirectURI)
	v.Set("scope", "openid profile email offline_access "+ch19OAuthPlanScope)
	v.Set("state", c.state)
	v.Set("nonce", c.nonce)
	v.Set("code_challenge", c.challenge)
	v.Set("code_challenge_method", "S256")
	v.Set("resource", ch19OAuthResource)
	// A v4 UUID: the "4" leading the third group is the version nibble.
	v.Set("ext_agent_host_id", "urn:uuid:3f2504e0-4f89-41d3-9a0c-0305e82c3301")
	return v
}

// authorize performs the browser hop and returns the callback query. Any
// override replaces one parameter so a test can be wrong in exactly one way.
func (c *ch19AuthClient) authorize(override url.Values) url.Values {
	c.t.Helper()
	v := c.authorizeValues()
	for k, vals := range override {
		v.Set(k, vals[0])
	}
	resp, err := c.hc.Get(c.md.AuthorizationEndpoint + "?" + v.Encode())
	if err != nil {
		c.t.Fatalf("authorize: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusFound {
		c.t.Fatalf("authorize: status %d, want 302", resp.StatusCode)
	}
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		c.t.Fatalf("authorize: unparseable Location %q: %v", resp.Header.Get("Location"), err)
	}
	cb := loc.Query()
	// Dynamic registration hands the issued client_id back through the
	// callback; a real agent persists this and never bootstraps again.
	if id := cb.Get("client_id"); id != "" {
		c.clientID = id
	}
	return cb
}

func (c *ch19AuthClient) postForm(endpoint string, v url.Values) (int, []byte) {
	c.t.Helper()
	resp, err := c.hc.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	if err != nil {
		c.t.Fatalf("POST %s: %v", endpoint, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("reading %s response: %v", endpoint, err)
	}
	return resp.StatusCode, body
}

func (c *ch19AuthClient) exchangeValues(code string) url.Values {
	v := url.Values{}
	v.Set("grant_type", "authorization_code")
	v.Set("code", code)
	v.Set("code_verifier", c.verifier)
	v.Set("redirect_uri", c.redirectURI)
	v.Set("client_id", c.clientID)
	v.Set("resource", ch19OAuthResource)
	return v
}

func (c *ch19AuthClient) exchange(code string, override url.Values) (int, []byte) {
	c.t.Helper()
	v := c.exchangeValues(code)
	for k, vals := range override {
		v.Set(k, vals[0])
	}
	return c.postForm(c.md.TokenEndpoint, v)
}

func (c *ch19AuthClient) refresh(refreshToken string) (int, []byte) {
	c.t.Helper()
	v := url.Values{}
	v.Set("grant_type", "refresh_token")
	v.Set("refresh_token", refreshToken)
	v.Set("client_id", c.clientID)
	return c.postForm(c.md.TokenEndpoint, v)
}

func (c *ch19AuthClient) revoke(token string) int {
	c.t.Helper()
	v := url.Values{}
	v.Set("token", token)
	v.Set("token_type_hint", "refresh_token")
	status, _ := c.postForm(c.md.RevocationEndpoint, v)
	return status
}

// ch19AuthTokens decodes a successful token response, failing the test if the
// server returned an error instead.
func ch19AuthTokens(t *testing.T, status int, body []byte) ch19TokenResponse {
	t.Helper()
	if status != http.StatusOK {
		t.Fatalf("token endpoint: status %d, body %s", status, body)
	}
	var tr ch19TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		t.Fatalf("decoding token response %s: %v", body, err)
	}
	return tr
}

func ch19AuthOAuthError(t *testing.T, body []byte) string {
	t.Helper()
	var e ch19OAuthErrorBody
	if err := json.Unmarshal(body, &e); err != nil {
		t.Fatalf("decoding error response %s: %v", body, err)
	}
	return e.Error
}

// ch19AuthHappyFlow runs discovery, authorize and the code exchange, and
// returns the client plus the issued tokens.
func ch19AuthHappyFlow(t *testing.T, s *ch19FakeOAuth) (*ch19AuthClient, ch19TokenResponse) {
	t.Helper()
	c := ch19AuthNewClient(t, s)
	cb := c.authorize(nil)
	if cb.Get("state") != c.state {
		t.Fatalf("callback state %q, want %q", cb.Get("state"), c.state)
	}
	status, body := c.exchange(cb.Get("code"), nil)
	return c, ch19AuthTokens(t, status, body)
}

// ---------------------------------------------------------------------------
// Real signature verification against the served JWKS
// ---------------------------------------------------------------------------

// ch19AuthVerifyIDToken does what the student is being asked to do: fetch the
// JWKS, rebuild the RSA public key from the two base64url integers, and check
// the RS256 signature over the first two segments. It returns an error rather
// than failing the test, because two tests below want opposite answers from it.
func ch19AuthVerifyIDToken(t *testing.T, jwksURI, idToken string) error {
	t.Helper()
	resp, err := http.Get(jwksURI)
	if err != nil {
		return fmt.Errorf("fetching jwks: %w", err)
	}
	defer resp.Body.Close()
	var ks ch19JWKS
	if err := json.NewDecoder(resp.Body).Decode(&ks); err != nil {
		return fmt.Errorf("decoding jwks: %w", err)
	}

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("id_token has %d segments, want 3", len(parts))
	}
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("decoding JWT header: %w", err)
	}
	var hdr ch19JWTHeader
	if err := json.Unmarshal(hb, &hdr); err != nil {
		return fmt.Errorf("unmarshaling JWT header: %w", err)
	}
	if hdr.Alg != "RS256" {
		// Rejecting on the header's say-so is itself a lesson: "alg":"none"
		// and HMAC-with-the-public-key are both real attacks, so the verifier
		// must pin the algorithm rather than obey the token.
		return fmt.Errorf("unexpected alg %q", hdr.Alg)
	}

	var jwk *ch19JWK
	for i := range ks.Keys {
		if ks.Keys[i].Kid == hdr.Kid {
			jwk = &ks.Keys[i]
			break
		}
	}
	if jwk == nil {
		return fmt.Errorf("no published key with kid %q", hdr.Kid)
	}

	nb, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return fmt.Errorf("decoding jwk n: %w", err)
	}
	eb, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return fmt.Errorf("decoding jwk e: %w", err)
	}
	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: int(new(big.Int).SetBytes(eb).Int64()),
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sig)
}

func ch19AuthClaims(t *testing.T, idToken string) ch19IDTokenClaims {
	t.Helper()
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		t.Fatalf("id_token has %d segments, want 3", len(parts))
	}
	cb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decoding claims: %v", err)
	}
	var claims ch19IDTokenClaims
	if err := json.Unmarshal(cb, &claims); err != nil {
		t.Fatalf("unmarshaling claims: %v", err)
	}
	return claims
}

// ch19AuthNoViolations fails with the full list, because "there was a
// violation" is useless and the prose of the violation is the whole point.
func ch19AuthNoViolations(t *testing.T, obs ch19OAuthObs) {
	t.Helper()
	if len(obs.Violations) != 0 {
		t.Fatalf("expected no violations, got %d:\n  %s",
			len(obs.Violations), strings.Join(obs.Violations, "\n  "))
	}
}

// ch19AuthHasViolation asserts some violation mentions substr.
func ch19AuthHasViolation(t *testing.T, obs ch19OAuthObs, substr string) {
	t.Helper()
	for _, v := range obs.Violations {
		if strings.Contains(v, substr) {
			return
		}
	}
	t.Fatalf("no violation mentioning %q; got:\n  %s", substr, strings.Join(obs.Violations, "\n  "))
}

// ---------------------------------------------------------------------------
// 1. Happy path
// ---------------------------------------------------------------------------

func TestCh19FakeOAuthHappyPathIsClean(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)

	if tok.TokenType != "Bearer" {
		t.Errorf("token_type %q, want Bearer", tok.TokenType)
	}
	if tok.ExpiresIn != 3600 {
		t.Errorf("expires_in %d, want 3600 by default", tok.ExpiresIn)
	}
	if !strings.Contains(tok.Scope, ch19OAuthPlanScope) {
		t.Errorf("granted scope %q is missing the plan scope", tok.Scope)
	}
	if !strings.HasPrefix(tok.AccessToken, "fake-access-") {
		t.Errorf("access token %q should be greppable", tok.AccessToken)
	}

	status, body := c.refresh(tok.RefreshToken)
	refreshed := ch19AuthTokens(t, status, body)
	if refreshed.RefreshToken == tok.RefreshToken {
		t.Error("refresh token did not rotate")
	}
	if got := c.revoke(refreshed.RefreshToken); got != http.StatusOK {
		t.Errorf("revoke status %d, want 200", got)
	}

	obs := s.Observations()
	ch19AuthNoViolations(t, obs)

	// The whole flow should also be visible in the observations, since that is
	// the only surface the chapter's checks read.
	if obs.DiscoveryHits != 1 {
		t.Errorf("DiscoveryHits %d, want 1", obs.DiscoveryHits)
	}
	if len(obs.AuthorizeHits) != 1 || len(obs.TokenHits) != 1 ||
		len(obs.RefreshHits) != 1 || len(obs.RevokeHits) != 1 {
		t.Errorf("hit counts: authorize %d token %d refresh %d revoke %d, want 1 each",
			len(obs.AuthorizeHits), len(obs.TokenHits), len(obs.RefreshHits), len(obs.RevokeHits))
	}
	if obs.RefreshReuseDetected {
		t.Error("RefreshReuseDetected on a clean sequential flow")
	}
	if obs.SawLocalhostRedirect || obs.SawDynamicClientAtToken {
		t.Error("clean flow tripped a negative-signal flag")
	}
	if obs.IssuedClientID == "" || obs.IssuedClientID == ch19OAuthBootstrapClientID {
		t.Errorf("IssuedClientID %q", obs.IssuedClientID)
	}
	if len(obs.Nonces) != 1 || obs.Nonces[0] != c.nonce {
		t.Errorf("Nonces %v, want [%q]", obs.Nonces, c.nonce)
	}
	if obs.RevokeHits[0].Token != refreshed.RefreshToken {
		t.Errorf("revoke recorded %q", obs.RevokeHits[0].Token)
	}
}

func TestCh19OAuthDiscoveryAdvertisesItself(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	md := c.md
	if md.Issuer != s.URL() {
		t.Errorf("issuer %q, want %q", md.Issuer, s.URL())
	}
	for name, got := range map[string]string{
		"authorization_endpoint": md.AuthorizationEndpoint,
		"token_endpoint":         md.TokenEndpoint,
		"revocation_endpoint":    md.RevocationEndpoint,
		"jwks_uri":               md.JWKSURI,
	} {
		if !strings.HasPrefix(got, s.URL()+"/") {
			t.Errorf("%s = %q, must be absolute and point at %q", name, got, s.URL())
		}
	}
	if len(md.CodeChallengeMethodsSupported) != 1 || md.CodeChallengeMethodsSupported[0] != "S256" {
		t.Errorf("code_challenge_methods_supported = %v, want [S256]", md.CodeChallengeMethodsSupported)
	}
	if len(md.IDTokenSigningAlgValuesSupported) != 1 || md.IDTokenSigningAlgValuesSupported[0] != "RS256" {
		t.Errorf("id_token_signing_alg_values_supported = %v, want [RS256]", md.IDTokenSigningAlgValuesSupported)
	}
}

// ---------------------------------------------------------------------------
// 2. MANDATORY: the published key really is the signing key
// ---------------------------------------------------------------------------

func TestCh19OAuthIDTokenVerifiesAgainstPublishedJWKS(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	if tok.IDToken == "" {
		t.Fatal("no id_token issued")
	}
	if err := ch19AuthVerifyIDToken(t, c.md.JWKSURI, tok.IDToken); err != nil {
		t.Fatalf("id_token did not verify against the served JWKS: %v\npublished key:\n%s",
			err, s.PublicKeyPEM())
	}

	claims := ch19AuthClaims(t, tok.IDToken)
	if claims.Iss != s.URL() {
		t.Errorf("iss %q, want %q", claims.Iss, s.URL())
	}
	if claims.Aud != s.Observations().IssuedClientID {
		t.Errorf("aud %q, want the issued client_id %q", claims.Aud, s.Observations().IssuedClientID)
	}
	if claims.Nonce != c.nonce {
		t.Errorf("nonce %q, want %q", claims.Nonce, c.nonce)
	}
	if claims.Sub == "" || !claims.EmailVerified || claims.Email == "" {
		t.Errorf("claims missing identity: %+v", claims)
	}
	if claims.Exp <= claims.Iat {
		t.Errorf("exp %d is not after iat %d", claims.Exp, claims.Iat)
	}
	if s.Observations().JWKSHits == 0 {
		t.Error("JWKSHits not recorded")
	}
}

// ---------------------------------------------------------------------------
// 3-9. Negative probes
// ---------------------------------------------------------------------------

func TestCh19OAuthPlainCodeChallengeMethodIsViolation(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	// "plain" sends the verifier as the challenge, which is PKCE in name only.
	c.authorize(url.Values{"code_challenge_method": {"plain"}})
	ch19AuthHasViolation(t, s.Observations(), "code_challenge_method=\"plain\"")
}

func TestCh19OAuthMissingStateAndNonceAreViolations(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	c.authorize(url.Values{"state": {""}, "nonce": {""}})
	obs := s.Observations()
	ch19AuthHasViolation(t, obs, "state is missing")
	ch19AuthHasViolation(t, obs, "nonce is missing")
}

func TestCh19OAuthHostIDMustBeOpaqueUUID(t *testing.T) {
	// Each of these is a plausible improvisation, and each leaks something
	// about the user or the machine.
	cases := map[string]string{
		"a8:5c:2c:1d:9f:30":   "MAC address",
		"bills-macbook.local": "hostname",
		"bill":                "username",
		"urn:uuid:3f2504e0-4f89-11d3-9a0c-0305e82c3301": "version nibble", // v1, MAC-derived
	}
	for value, want := range cases {
		s := ch19NewFakeOAuth(ch19OAuthOptions{})
		c := ch19AuthNewClient(t, s)
		c.authorize(url.Values{"ext_agent_host_id": {value}})
		ch19AuthHasViolation(t, s.Observations(), want)
		s.Close()
	}
}

func TestCh19OAuthWrongCodeVerifierIsRejected(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	cb := c.authorize(nil)
	// A correct-looking but unrelated verifier: exactly what a client that
	// generates fresh randomness at token time would send.
	status, body := c.exchange(cb.Get("code"), url.Values{"code_verifier": {ch19OAuthRandomToken(32)}})
	if status != http.StatusBadRequest {
		t.Fatalf("status %d, want 400; body %s", status, body)
	}
	if got := ch19AuthOAuthError(t, body); got != "invalid_grant" {
		t.Errorf("error %q, want invalid_grant", got)
	}
	ch19AuthHasViolation(t, s.Observations(), "PKCE failure")
}

func TestCh19OAuthLocalhostRedirectIsDetected(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	c.redirectURI = "http://localhost:54321/auth/callback"
	c.authorize(nil)

	obs := s.Observations()
	if !obs.SawLocalhostRedirect {
		t.Error("SawLocalhostRedirect not set for a localhost redirect_uri")
	}
	ch19AuthHasViolation(t, obs, "localhost")
}

func TestCh19OAuthCodeIsSingleUse(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	cb := c.authorize(nil)
	code := cb.Get("code")
	if status, body := c.exchange(code, nil); status != http.StatusOK {
		t.Fatalf("first exchange: status %d body %s", status, body)
	}
	status, body := c.exchange(code, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("replay: status %d, want 400", status)
	}
	if got := ch19AuthOAuthError(t, body); got != "invalid_grant" {
		t.Errorf("replay error %q, want invalid_grant", got)
	}
	ch19AuthHasViolation(t, s.Observations(), "already redeemed")
}

func TestCh19OAuthBootstrapClientIDAtTokenIsViolation(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	cb := c.authorize(nil)
	c.exchange(cb.Get("code"), url.Values{"client_id": {ch19OAuthBootstrapClientID}})

	obs := s.Observations()
	if !obs.SawDynamicClientAtToken {
		t.Error("SawDynamicClientAtToken not set")
	}
	ch19AuthHasViolation(t, obs, "bootstrap id is only valid for the first authorize")
}

func TestCh19OAuthRetiredRefreshTokenIsDetected(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	first := tok.RefreshToken

	status, body := c.refresh(first)
	ch19AuthTokens(t, status, body) // rotates: `first` is now dead

	status, body = c.refresh(first)
	if status != http.StatusBadRequest {
		t.Fatalf("reusing a retired refresh token: status %d, want 400", status)
	}
	if got := ch19AuthOAuthError(t, body); got != "invalid_grant" {
		t.Errorf("error %q, want invalid_grant", got)
	}
	obs := s.Observations()
	if !obs.RefreshReuseDetected {
		t.Error("RefreshReuseDetected not set")
	}
	ch19AuthHasViolation(t, obs, "already rotated out")
}

func TestCh19OAuthRefreshTokenRotates(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	seen := []string{tok.RefreshToken}
	current := tok.RefreshToken
	for i := 0; i < 2; i++ {
		status, body := c.refresh(current)
		next := ch19AuthTokens(t, status, body)
		current = next.RefreshToken
		seen = append(seen, current)
	}

	uniq := map[string]bool{}
	for _, r := range seen {
		if r == "" {
			t.Fatal("empty refresh token issued")
		}
		uniq[r] = true
	}
	if len(uniq) != 3 {
		t.Errorf("two sequential refreshes produced %d distinct tokens, want 3: %v", len(uniq), seen)
	}
	if got := s.CurrentRefreshToken(); got != current {
		t.Errorf("CurrentRefreshToken %q, want %q", got, current)
	}
	ch19AuthNoViolations(t, s.Observations())
}

func TestCh19OAuthConcurrentRefreshIsObservable(t *testing.T) {
	// Two refreshes fired at once against one session. The in-handler delay
	// guarantees they overlap, so the loser presents a token the winner has
	// already retired. This is the exact failure a client without single-flight
	// refresh produces in the field, where it is maddening to reproduce.
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.refresh(tok.RefreshToken)
		}()
	}
	wg.Wait()

	obs := s.Observations()
	if obs.MaxConcurrentRefresh < 2 {
		t.Errorf("MaxConcurrentRefresh %d, want at least 2", obs.MaxConcurrentRefresh)
	}
	if !obs.RefreshReuseDetected {
		t.Error("two concurrent refreshes did not trip RefreshReuseDetected")
	}
}

func TestCh19OAuthUnsignedKeyMismatchFailsVerification(t *testing.T) {
	// The strongest anti-shortcut probe: the token is well-formed, the kid
	// names a published key, every claim is right, and only the signature is
	// wrong. Nothing short of real verification catches it.
	s := ch19NewFakeOAuth(ch19OAuthOptions{UnsignedKeyMismatch: true})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	if err := ch19AuthVerifyIDToken(t, c.md.JWKSURI, tok.IDToken); err == nil {
		t.Fatal("id_token signed by an unpublished key verified against the JWKS; the probe is not probing anything")
	}
	// It must still look completely normal to a client that does not verify.
	claims := ch19AuthClaims(t, tok.IDToken)
	if claims.Nonce != c.nonce || claims.Iss != s.URL() {
		t.Errorf("mismatch token should still carry correct claims, got %+v", claims)
	}
}

func TestCh19OAuthWrongNonce(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{WrongNonce: true})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	claims := ch19AuthClaims(t, tok.IDToken)
	if claims.Nonce == c.nonce {
		t.Fatalf("WrongNonce had no effect: id_token nonce %q equals the one sent", claims.Nonce)
	}
	if claims.Nonce == "" {
		t.Error("the wrong nonce should be plausible, not absent; an empty one is too easy to catch")
	}
	// The signature must still be valid, or the test cannot tell nonce
	// validation apart from signature validation.
	if err := ch19AuthVerifyIDToken(t, c.md.JWKSURI, tok.IDToken); err != nil {
		t.Errorf("WrongNonce token should still be correctly signed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Remaining option probes
// ---------------------------------------------------------------------------

func TestCh19OAuthOmitPlanScope(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{OmitPlanScope: true})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	// Ask without the plan scope; with OmitPlanScope set that is not an error.
	cb := c.authorize(url.Values{"scope": {"openid profile email offline_access"}})
	status, body := c.exchange(cb.Get("code"), nil)
	tok := ch19AuthTokens(t, status, body)

	if strings.Contains(tok.Scope, ch19OAuthPlanScope) {
		t.Errorf("granted scope %q still contains the plan scope", tok.Scope)
	}
	ch19AuthNoViolations(t, s.Observations())
}

func TestCh19OAuthAccessTokenTTLIsReported(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{AccessTokenTTL: 2 * time.Second})
	defer s.Close()

	_, tok := ch19AuthHappyFlow(t, s)
	if tok.ExpiresIn != 2 {
		t.Errorf("expires_in %d, want 2", tok.ExpiresIn)
	}
}

func TestCh19OAuthRefreshFailsWith(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{RefreshFailsWith: "invalid_grant"})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	status, body := c.refresh(tok.RefreshToken)
	if status != http.StatusBadRequest {
		t.Fatalf("first refresh: status %d, want 400", status)
	}
	if got := ch19AuthOAuthError(t, body); got != "invalid_grant" {
		t.Errorf("error %q, want invalid_grant", got)
	}
	// An injected server fault is not the client's mistake, so it must not
	// pollute the violation list that the student's grade is computed from.
	ch19AuthNoViolations(t, s.Observations())

	// Only the FIRST refresh fails, so a client that correctly re-authorizes
	// and then refreshes again is not punished twice.
	status, body = c.refresh(tok.RefreshToken)
	if status != http.StatusOK {
		t.Errorf("second refresh: status %d, want 200; body %s", status, body)
	}
}

func TestCh19OAuthExpiredAccessTokenIsFlagged(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{AccessTokenTTL: time.Second})
	defer s.Close()

	c, tok := ch19AuthHappyFlow(t, s)
	s.ExpireAccessToken()

	req, err := http.NewRequest(http.MethodGet, c.md.JWKSURI, nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, err := c.hc.Do(req)
	if err != nil {
		t.Fatalf("GET with dead token: %v", err)
	}
	resp.Body.Close()

	ch19AuthHasViolation(t, s.Observations(), "after it was expired or revoked")
}

func TestCh19OAuthBootstrapClientIDIsSingleUse(t *testing.T) {
	s := ch19NewFakeOAuth(ch19OAuthOptions{})
	defer s.Close()

	c := ch19AuthNewClient(t, s)
	c.authorize(nil) // legitimate dynamic registration
	ch19AuthNoViolations(t, s.Observations())

	c.clientID = ch19OAuthBootstrapClientID // pretend the agent forgot to persist
	c.authorize(nil)
	ch19AuthHasViolation(t, s.Observations(), "must be persisted and reused")
}
