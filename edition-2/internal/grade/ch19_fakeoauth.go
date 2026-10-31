package grade

// ch19_fakeoauth.go — a fake OpenAI OAuth 2.0 / OIDC provider for grading
// Chapter 19.
//
// WHY THIS EXISTS, AND WHY IT IS NOT A STUB
//
// Chapter 19 asks the student to implement an OAuth 2.0 authorization code
// flow with PKCE against ChatGPT's provider. Every interesting thing about
// that exercise is a *validation*: did you really hash the verifier, did you
// really check the signature, did you really bind the nonce. A grader built
// out of canned JSON responses cannot test any of it, because canned JSON is
// indifferent to what the client sent. A student who generates a random
// code_verifier, never derives the challenge from it, and ignores the
// id_token entirely would sail through a stub and fail against the real
// provider on the first run. That is the worst possible grader: it certifies
// the exact bug the chapter exists to prevent.
//
// So this fixture does the real cryptography. At construction it generates an
// ephemeral 2048-bit RSA key pair, serves a genuine JWKS document derived
// from that key, and signs genuine RS256 JWTs with it. A student who skips
// signature verification cannot be distinguished from one who does it
// correctly *on the happy path* — which is why the fixture also offers
// UnsignedKeyMismatch, where the id_token is signed by a throwaway key that
// was never published. Against that server, a real verifier rejects and a
// pretend verifier accepts, and the two are finally telling different stories.
//
// The same reasoning drives the PKCE check at the token endpoint. The server
// remembers the code_challenge it was handed at /authorize, recomputes
// base64url(sha256(code_verifier)) from the verifier presented at the token
// endpoint, and compares. There is no way to satisfy that without having
// implemented PKCE. It is the single most load-bearing line in the file.
//
// HOW GRADING READS THIS
//
// Checks never inspect the server's internals. They call Observations(),
// which returns a deep copy of everything the server saw, including an
// append-only Violations list where each entry names the mistake in prose. A
// clean run leaves Violations empty; that property is itself tested below,
// because a fixture that reports a violation on correct input would fail
// every student and nobody would notice for a while.
//
// PROVENANCE OF THE METADATA
//
// The discovery document mirrors the shape of the real one fetched live from
// https://auth.openai.com/.well-known/openid-configuration: S256 as the only
// challenge method, RS256 as the only id_token algorithm, authorization_code
// and refresh_token as the only grants, and token_endpoint_auth_method "none"
// (a public client, no secret, which is precisely why PKCE is mandatory
// rather than optional). One honest caveat worth knowing: the real document's
// scopes_supported lists only openid, profile, email and offline_access. It
// does NOT advertise chatgpt.tokens.use.direct, even though the plan-gated
// surface requires it. Unadvertised scopes are ordinary for preview surfaces,
// so the fixture reproduces the real (short) advertised list while still
// requiring the plan scope on the wire.

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Wire constants the student's client is graded against. These are values,
// not policy knobs: changing one here silently moves the goalposts for every
// check in the chapter, so they are named and kept together.
const (
	// ch19OAuthBootstrapClientID is the well-known identifier a client uses
	// before it has registered. It is valid exactly once per server. An agent
	// that keeps using it has failed to persist the credential it was issued,
	// which in production means re-registering on every launch.
	ch19OAuthBootstrapClientID = "dynamic_agent_client"

	// ch19OAuthResource is the RFC 8707 resource indicator. Getting this wrong
	// yields a token audienced for the wrong API, which fails confusingly at
	// call time rather than at grant time, so the fixture catches it early.
	ch19OAuthResource = "https://api.openai.com/v1"

	// ch19OAuthPlanScope gates the ChatGPT-plan passthrough. It is the scope
	// most likely to be dropped by a student copying a generic OIDC snippet.
	ch19OAuthPlanScope = "chatgpt.tokens.use.direct"

	// ch19OAuthCallbackPath is the only redirect path this provider accepts.
	ch19OAuthCallbackPath = "/auth/callback"

	// ch19OAuthDefaultAccessTTL is the expires_in reported when the options
	// struct leaves AccessTokenTTL at zero.
	ch19OAuthDefaultAccessTTL = time.Hour

	// ch19OAuthCodeTTL bounds the life of an authorization code. Real codes
	// are short-lived precisely because they travel through a browser
	// redirect, which is the least trustworthy hop in the whole flow.
	ch19OAuthCodeTTL = 5 * time.Minute

	// ch19OAuthRefreshDelay is held inside the refresh handler so that two
	// genuinely concurrent refreshes overlap in wall-clock time. Without a
	// delay the race the grader is hunting for almost never materializes and
	// a broken client passes by luck.
	ch19OAuthRefreshDelay = 150 * time.Millisecond

	// ch19OAuthSubject is a stable fake account id, shaped like the ULID-style
	// identifiers the real provider issues.
	ch19OAuthSubject = "user_01HQZ8K3RMBVT5XJ7P2N4YDC6A"

	// ch19OAuthEmail is the stable fake account email placed in the id_token.
	ch19OAuthEmail = "ensemble-grader@example.com"
)

// ch19OAuthOptions configures one fake provider. The zero value is the
// well-behaved provider used for happy-path grading; every field exists to
// make a specific client-side mistake observable, so each one is a probe
// rather than a preference.
type ch19OAuthOptions struct {
	// AccessTokenTTL is reported verbatim as expires_in. Zero means one hour.
	// Set it to a couple of seconds to force a client that refreshes
	// proactively to actually do so during a short grading run; a client that
	// only refreshes reactively on a 401 will never fire, which is the
	// difference the check is looking for.
	AccessTokenTTL time.Duration

	// WrongNonce puts a nonce in the id_token that the client never sent. A
	// client that binds the nonce rejects the token; one that decodes the JWT
	// for its claims without checking the binding accepts it, and has a replay
	// hole it does not know about.
	WrongNonce bool

	// OmitPlanScope drops chatgpt.tokens.use.direct from the granted scope in
	// the token response, and stops requiring it at /authorize. Downscoping is
	// legal for a provider, so the client must read the granted scope back out
	// of the response rather than assuming it got what it asked for.
	OmitPlanScope bool

	// RefreshFailsWith, when non-empty, makes the FIRST refresh_token grant
	// return HTTP 400 with this error code. "invalid_grant" is terminal: the
	// correct response is to discard the session and re-authorize, not to
	// retry, and a client that retries a terminal error spins forever.
	RefreshFailsWith string

	// UnsignedKeyMismatch signs the id_token with a throwaway RSA key that is
	// never published in the JWKS. This is the strongest anti-shortcut probe
	// in the file: no amount of claim parsing saves a client that does not
	// actually verify the signature against the published key.
	UnsignedKeyMismatch bool
}

// ch19AuthorizeReq is one observed call to the authorization endpoint. Every
// parameter the fixture validates is captured as a named field so a check can
// assert on it directly, and RawQuery is kept alongside so a check can assert
// on something the fixture did not think to name.
type ch19AuthorizeReq struct {
	Method              string
	ClientID            string
	ResponseType        string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	ExtAgentHostID      string
	RawQuery            string
}

// ch19TokenReq is one observed call to the token endpoint, used for both
// grant types. RawBody is retained so the parent can grep a request body for
// a leaked secret: the most common way credentials escape is a client that
// logs or forwards the whole form.
type ch19TokenReq struct {
	GrantType    string
	Code         string
	CodeVerifier string
	RedirectURI  string
	ClientID     string
	RefreshToken string
	Resource     string
	Scope        string
	ContentType  string
	RawBody      string
}

// ch19RevokeReq is one observed call to the revocation endpoint.
type ch19RevokeReq struct {
	Token         string
	TokenTypeHint string
	RawBody       string
}

// ch19OAuthObs is the entire grading surface. Checks read this and nothing
// else, which keeps them from coupling to the server's internal bookkeeping
// and lets the fixture be rewritten without rewriting the chapter's checks.
type ch19OAuthObs struct {
	DiscoveryHits int
	JWKSHits      int

	AuthorizeHits []ch19AuthorizeReq
	TokenHits     []ch19TokenReq // authorization_code grants only
	RefreshHits   []ch19TokenReq // refresh_token grants only
	RevokeHits    []ch19RevokeReq

	// Violations is append-only and human-readable. Each entry names what was
	// wrong in enough detail to be pasted into a failure message without
	// further explanation, because the student reading it has no access to
	// this file.
	Violations []string

	// RefreshReuseDetected fires when a retired refresh token is presented.
	// With rotation on, that means either a replay or two concurrent refreshes
	// from the same client, and both are bugs.
	RefreshReuseDetected bool

	// MaxConcurrentRefresh is the high-water mark of simultaneous refresh
	// requests. A correct client serializes refreshes behind a single flight,
	// so anything above 1 is a finding even when no reuse was detected.
	MaxConcurrentRefresh int

	IssuedClientID string
	Nonces         []string

	IssuedAccessTokens  []string
	IssuedRefreshTokens []string

	SawDynamicClientAtToken bool
	SawLocalhostRedirect    bool
}

// ch19OAuthCode is one outstanding authorization code and everything the
// token endpoint must later check it against. The challenge, redirect URI and
// nonce live here rather than in the code string because a bearer code that
// carries its own validation data is a code that can be forged.
type ch19OAuthCode struct {
	challenge   string
	method      string
	nonce       string
	redirectURI string
	scope       string
	state       string
	resource    string
	issuedAt    time.Time
	used        bool
}

// ch19FakeOAuth is the server. All mutable state is guarded by mu because the
// refresh-concurrency probe deliberately drives it from several goroutines at
// once; an unguarded field here would turn a grading run into a flake.
type ch19FakeOAuth struct {
	opts ch19OAuthOptions
	srv  *httptest.Server

	// key is published in the JWKS. signKey is what actually signs id_tokens.
	// They are the same pointer unless UnsignedKeyMismatch is set, and keeping
	// them as two fields is what makes that option a two-line change rather
	// than a special case threaded through the signing path.
	key     *rsa.PrivateKey
	signKey *rsa.PrivateKey
	kid     string

	mu              sync.Mutex
	obs             ch19OAuthObs
	codes           map[string]*ch19OAuthCode
	issuedClientID  string
	clientHandedOut bool
	liveRefresh     string
	retiredRefresh  map[string]bool
	deadAccess      map[string]bool
	refreshCount    int
	inFlight        int
	seq             int
}

// ch19NewFakeOAuth starts a fake provider and returns it ready to serve. The
// caller must Close it.
func ch19NewFakeOAuth(o ch19OAuthOptions) *ch19FakeOAuth {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("ch19 fake oauth: generating RSA key: " + err.Error())
	}
	s := &ch19FakeOAuth{
		opts:           o,
		key:            key,
		signKey:        key,
		kid:            ch19OAuthKeyID(&key.PublicKey),
		codes:          map[string]*ch19OAuthCode{},
		retiredRefresh: map[string]bool{},
		deadAccess:     map[string]bool{},
	}
	if o.UnsignedKeyMismatch {
		// A second, structurally valid key that is never published. The
		// resulting id_token parses perfectly and fails verification, which is
		// exactly the shape of a real token-substitution attack.
		rogue, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic("ch19 fake oauth: generating rogue RSA key: " + err.Error())
		}
		s.signKey = rogue
	}
	// The client_id is minted up front so the fixture has one stable answer
	// for the whole run, but it is not considered "handed out" until the first
	// /authorize actually returns it. That distinction is what makes reuse of
	// the bootstrap id detectable rather than merely suspicious.
	s.issuedClientID = "app_ch19_" + ch19OAuthRandomToken(9)
	s.obs.IssuedClientID = s.issuedClientID

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleDiscovery)
	mux.HandleFunc("/.well-known/jwks.json", s.handleJWKS)
	mux.HandleFunc("/api/accounts/authorize", s.handleAuthorize)
	mux.HandleFunc("/api/accounts/oauth/token", s.handleToken)
	mux.HandleFunc("/api/accounts/oauth/revoke", s.handleRevoke)
	s.srv = httptest.NewServer(mux)
	return s
}

// URL is the server's base URL, which is also its issuer. Every absolute URL
// in the discovery document is built from this, so a client that hardcodes
// auth.openai.com instead of following discovery will simply not reach the
// fixture, and that failure is loud.
func (s *ch19FakeOAuth) URL() string { return s.srv.URL }

// Close shuts the server down and waits for outstanding requests.
func (s *ch19FakeOAuth) Close() { s.srv.Close() }

// CurrentRefreshToken reports the live refresh token, which rotates on every
// successful refresh. Returns "" once it has been revoked.
func (s *ch19FakeOAuth) CurrentRefreshToken() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.liveRefresh
}

// ExpireAccessToken marks the most recently issued access token dead. This
// fixture is an authorization server and has no protected resource to guard,
// so the practical effect is twofold: the token is recorded as dead, and if
// the client ever presents it as a Bearer credential to any endpoint here,
// that is recorded as a violation. Use it with a short AccessTokenTTL to make
// "did this client refresh before the token died" an answerable question.
func (s *ch19FakeOAuth) ExpireAccessToken() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.obs.IssuedAccessTokens); n > 0 {
		s.deadAccess[s.obs.IssuedAccessTokens[n-1]] = true
	}
}

// PublicKeyPEM renders the published verification key in PKIX PEM form. It is
// a debugging convenience: when a signature check fails it is worth being able
// to see, in one glance, whether the key you are verifying against is the key
// the server is publishing.
func (s *ch19FakeOAuth) PublicKeyPEM() string {
	der, err := x509.MarshalPKIXPublicKey(&s.key.PublicKey)
	if err != nil {
		return ""
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

// Observations returns a deep copy under the lock. The copy is not politeness:
// the refresh probe is still running goroutines when a check starts reading,
// and handing out the live slices would be a data race that shows up as an
// intermittent grading failure months later.
func (s *ch19FakeOAuth) Observations() ch19OAuthObs {
	s.mu.Lock()
	defer s.mu.Unlock()
	o := s.obs
	o.AuthorizeHits = append([]ch19AuthorizeReq(nil), s.obs.AuthorizeHits...)
	o.TokenHits = append([]ch19TokenReq(nil), s.obs.TokenHits...)
	o.RefreshHits = append([]ch19TokenReq(nil), s.obs.RefreshHits...)
	o.RevokeHits = append([]ch19RevokeReq(nil), s.obs.RevokeHits...)
	o.Violations = append([]string(nil), s.obs.Violations...)
	o.Nonces = append([]string(nil), s.obs.Nonces...)
	o.IssuedAccessTokens = append([]string(nil), s.obs.IssuedAccessTokens...)
	o.IssuedRefreshTokens = append([]string(nil), s.obs.IssuedRefreshTokens...)
	return o
}

// violate records a finding. Callers hold mu.
func (s *ch19FakeOAuth) violate(format string, args ...any) {
	s.obs.Violations = append(s.obs.Violations, fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------------------
// Discovery and keys
// ---------------------------------------------------------------------------

// ch19OIDCMetadata is the subset of the discovery document that matters to a
// client. Field order here is the JSON field order, which is irrelevant to
// correctness but makes a captured response diffable against the real one.
type ch19OIDCMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
}

func (s *ch19FakeOAuth) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	s.noteBearer(r)
	s.mu.Lock()
	s.obs.DiscoveryHits++
	if r.Method != http.MethodGet {
		s.violate("discovery: method %s; the metadata document is a plain GET", r.Method)
	}
	s.mu.Unlock()

	base := s.srv.URL
	// Every endpoint is advertised as an absolute URL pointing back at this
	// server. A client that joins these against a hardcoded host, or that
	// treats them as paths, breaks here rather than in production.
	md := ch19OIDCMetadata{
		Issuer:                            base,
		AuthorizationEndpoint:             base + "/api/accounts/authorize",
		TokenEndpoint:                     base + "/api/accounts/oauth/token",
		RevocationEndpoint:                base + "/api/accounts/oauth/revoke",
		JWKSURI:                           base + "/.well-known/jwks.json",
		ResponseTypesSupported:            []string{"code"},
		GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported:     []string{"S256"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		TokenEndpointAuthMethodsSupported: []string{"none", "client_secret_post"},
		// Deliberately the real provider's short list. The plan scope is
		// required on the wire but not advertised here, which is true of the
		// live endpoint and is a useful lesson: discovery metadata is a hint,
		// not an exhaustive contract.
		ScopesSupported:       []string{"openid", "profile", "email", "offline_access"},
		SubjectTypesSupported: []string{"public"},
	}
	ch19WriteJSON(w, http.StatusOK, md)
}

// ch19JWK is one RSA public key in JWK form. "n" and "e" are base64url of the
// big-endian magnitude bytes with no padding, which is the one detail people
// get wrong by reaching for StdEncoding and then wondering why every
// signature fails.
type ch19JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type ch19JWKS struct {
	Keys []ch19JWK `json:"keys"`
}

func (s *ch19FakeOAuth) handleJWKS(w http.ResponseWriter, r *http.Request) {
	s.noteBearer(r)
	s.mu.Lock()
	s.obs.JWKSHits++
	if r.Method != http.MethodGet {
		s.violate("jwks: method %s; the key set is a plain GET", r.Method)
	}
	s.mu.Unlock()
	ch19WriteJSON(w, http.StatusOK, ch19JWKS{Keys: []ch19JWK{ch19PublicJWK(&s.key.PublicKey, s.kid)}})
}

// ch19PublicJWK converts an RSA public key to its JWK representation. This is
// hand-rolled on purpose: it is about ten lines, it removes a dependency, and
// seeing that a JWK is just two base64url integers demystifies the format.
func ch19PublicJWK(pub *rsa.PublicKey, kid string) ch19JWK {
	return ch19JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		// The exponent is also a big-endian integer, not a decimal string.
		// For the usual 65537 this encodes to "AQAB", which is why that
		// particular string appears in almost every JWKS on the internet.
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

// ch19OAuthKeyID derives a stable kid from the modulus. Deriving it rather
// than hardcoding it means the kid in the JWKS and the kid in the JWT header
// cannot drift apart, which is a real and annoying class of bug.
func ch19OAuthKeyID(pub *rsa.PublicKey) string {
	sum := sha256.Sum256(pub.N.Bytes())
	return "ch19-" + hex.EncodeToString(sum[:8])
}

// ---------------------------------------------------------------------------
// Authorization endpoint
// ---------------------------------------------------------------------------

// ch19AgentHostIDRe matches urn:uuid: followed by a canonical v4 UUID. The
// literal 4 in the third group is the version nibble; it is what distinguishes
// a randomly generated identifier from one derived from the host's MAC
// address, which is the privacy failure this parameter exists to avoid.
var ch19AgentHostIDRe = regexp.MustCompile(
	`^urn:uuid:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ch19MACRe matches a colon-separated MAC address anywhere in a value. A v1
// UUID embeds the MAC without colons, but a developer improvising a "host id"
// most often pastes the formatted address straight in.
var ch19MACRe = regexp.MustCompile(`(?:[0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}`)

// ch19HostnameRe matches a dotted name like "bills-macbook.local".
var ch19HostnameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*\.[A-Za-z0-9-]+$`)

// ch19UsernameRe matches a bare POSIX-ish login name.
var ch19UsernameRe = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

func (s *ch19FakeOAuth) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	s.noteBearer(r)
	q := r.URL.Query()
	req := ch19AuthorizeReq{
		Method:              r.Method,
		ClientID:            q.Get("client_id"),
		ResponseType:        q.Get("response_type"),
		RedirectURI:         q.Get("redirect_uri"),
		Scope:               q.Get("scope"),
		State:               q.Get("state"),
		Nonce:               q.Get("nonce"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		Resource:            q.Get("resource"),
		ExtAgentHostID:      q.Get("ext_agent_host_id"),
		RawQuery:            r.URL.RawQuery,
	}

	s.mu.Lock()
	s.obs.AuthorizeHits = append(s.obs.AuthorizeHits, req)
	if req.Nonce != "" {
		s.obs.Nonces = append(s.obs.Nonces, req.Nonce)
	}

	if r.Method != http.MethodGet {
		s.violate("authorize: method %s; the authorization endpoint is reached by navigating a browser, which is a GET", r.Method)
	}
	if req.ResponseType != "code" {
		s.violate("authorize: response_type=%q; only the authorization code flow is supported, so this must be %q", req.ResponseType, "code")
	}

	// --- PKCE ---------------------------------------------------------
	// S256 only. "plain" sends the verifier itself as the challenge, which
	// means anyone who can read the authorization request can complete the
	// exchange; it defeats the entire mechanism while looking like PKCE in a
	// log. Treating a missing method as "plain" by default is a historical
	// wart in RFC 7636 and is why an empty value is flagged too.
	switch req.CodeChallengeMethod {
	case "S256":
	case "":
		s.violate("authorize: code_challenge_method is missing; it defaults to \"plain\" under RFC 7636, which sends the verifier in the clear, so it must be set explicitly to \"S256\"")
	default:
		s.violate("authorize: code_challenge_method=%q; this provider supports only \"S256\", and \"plain\" transmits the verifier itself and so provides no protection at all", req.CodeChallengeMethod)
	}
	if req.CodeChallenge == "" {
		s.violate("authorize: code_challenge is missing; without it the authorization code can be redeemed by anyone who intercepts it")
	}

	// --- CSRF and replay binding ---------------------------------------
	// state defends the callback against cross-site request forgery; nonce
	// binds the id_token to this specific authorization request. They are
	// different defenses and a client needs both, which is why neither is
	// allowed to stand in for the other here.
	if req.State == "" {
		s.violate("authorize: state is missing; it is the only thing tying the browser callback back to the request this process started, and without it any attacker-supplied code will be accepted")
	}
	if req.Nonce == "" {
		s.violate("authorize: nonce is missing; it is what binds the returned id_token to this request, and without it a previously captured id_token can be replayed")
	}

	// --- scope ----------------------------------------------------------
	scopes := strings.Fields(req.Scope)
	if !ch19Contains(scopes, "openid") {
		s.violate("authorize: scope %q does not contain \"openid\"; without it this is a bare OAuth grant and no id_token is issued", req.Scope)
	}
	if !ch19Contains(scopes, "offline_access") {
		s.violate("authorize: scope %q does not contain \"offline_access\"; without it no refresh token is issued and the session dies at the first expiry", req.Scope)
	}
	if !s.opts.OmitPlanScope && !ch19Contains(scopes, ch19OAuthPlanScope) {
		s.violate("authorize: scope %q does not contain %q, which is what gates the ChatGPT-plan passthrough", req.Scope, ch19OAuthPlanScope)
	}

	// --- resource indicator ----------------------------------------------
	if req.Resource != ch19OAuthResource {
		s.violate("authorize: resource=%q; it must be exactly %q so the issued token is audienced for the right API", req.Resource, ch19OAuthResource)
	}

	// --- host identifier --------------------------------------------------
	s.checkAgentHostID(req.ExtAgentHostID)

	// --- redirect URI -----------------------------------------------------
	// A loopback redirect is the whole security model for a native public
	// client: the code comes back to a listener only this process owns.
	// "localhost" breaks that, because it resolves through the system resolver
	// and may land on ::1, on a hosts-file override, or on a different
	// interface than the one being listened on. The literal 127.0.0.1 cannot
	// be redirected anywhere.
	redirect, err := url.Parse(req.RedirectURI)
	if err != nil || req.RedirectURI == "" {
		s.violate("authorize: redirect_uri=%q does not parse as a URL", req.RedirectURI)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_request", "redirect_uri is missing or unparseable")
		return
	}
	if redirect.Scheme != "http" {
		s.violate("authorize: redirect_uri scheme %q; a loopback redirect is plain http because there is no way to obtain a trusted certificate for 127.0.0.1", redirect.Scheme)
	}
	switch redirect.Hostname() {
	case "127.0.0.1":
	case "localhost":
		s.obs.SawLocalhostRedirect = true
		s.violate("authorize: redirect_uri host is \"localhost\"; it must be the literal 127.0.0.1, because \"localhost\" goes through name resolution and can be pointed at ::1 or at another host entirely by a hosts-file entry")
	default:
		s.violate("authorize: redirect_uri host %q is not 127.0.0.1; the authorization code must come back to a loopback listener owned by this process", redirect.Hostname())
	}
	if redirect.Path != ch19OAuthCallbackPath {
		s.violate("authorize: redirect_uri path %q; it must be exactly %q", redirect.Path, ch19OAuthCallbackPath)
	}

	// --- client identity ---------------------------------------------------
	// Dynamic registration is a one-shot bootstrap. The first authorize may
	// present the well-known bootstrap id; after that the agent is expected to
	// have persisted the id it was issued. An agent that keeps bootstrapping
	// registers a new client on every launch, which in a real deployment grows
	// an unbounded pile of orphaned client records.
	switch {
	case req.ClientID == ch19OAuthBootstrapClientID && !s.clientHandedOut:
		s.clientHandedOut = true
	case req.ClientID == ch19OAuthBootstrapClientID:
		s.violate("authorize: client_id=%q again after %q was issued; the registered client_id must be persisted and reused", ch19OAuthBootstrapClientID, s.issuedClientID)
	case req.ClientID == s.issuedClientID:
		s.clientHandedOut = true
	case req.ClientID == "":
		s.violate("authorize: client_id is missing")
	default:
		s.violate("authorize: client_id=%q is not a registered client; use %q to bootstrap, then the issued id", req.ClientID, ch19OAuthBootstrapClientID)
	}

	// The code is minted even when violations were recorded. Aborting here
	// would hide every later mistake behind the first one, and a grading run
	// that surfaces one problem per execution is a grading run nobody finishes.
	s.seq++
	code := fmt.Sprintf("ch19code_%d_%s", s.seq, ch19OAuthRandomToken(12))
	s.codes[code] = &ch19OAuthCode{
		challenge:   req.CodeChallenge,
		method:      req.CodeChallengeMethod,
		nonce:       req.Nonce,
		redirectURI: req.RedirectURI,
		scope:       req.Scope,
		state:       req.State,
		resource:    req.Resource,
		issuedAt:    time.Now(),
	}
	issued := s.issuedClientID
	s.mu.Unlock()

	// code and state are the standard callback parameters. client_id is an
	// addition this fixture makes: dynamic registration has to deliver the
	// issued identifier back to the client somehow, and the callback is the
	// only in-band channel the flow provides once the browser hop is over.
	rq := redirect.Query()
	rq.Set("code", code)
	rq.Set("state", req.State)
	rq.Set("client_id", issued)
	redirect.RawQuery = rq.Encode()
	http.Redirect(w, r, redirect.String(), http.StatusFound)
}

// checkAgentHostID validates the per-machine grant pin. Callers hold mu.
//
// The parameter exists so a stolen refresh token is useless on another
// machine. That only works if the value is opaque: an identifier derived from
// the hostname, the login name, or the network card tells the provider (and
// anyone who sees the URL) who and where the user is, which turns a security
// control into a tracking beacon.
func (s *ch19FakeOAuth) checkAgentHostID(v string) {
	switch {
	case v == "":
		s.violate("authorize: ext_agent_host_id is missing; it pins the grant to one machine so a stolen refresh token cannot be used elsewhere")
	case ch19MACRe.MatchString(v):
		s.violate("authorize: ext_agent_host_id=%q contains a MAC address; a hardware serial number is a permanent cross-application identifier and must never be sent", v)
	case ch19HostnameRe.MatchString(v):
		s.violate("authorize: ext_agent_host_id=%q looks like a hostname; it must be an opaque urn:uuid: v4, not a name that identifies the user or their machine", v)
	case ch19UsernameRe.MatchString(v):
		s.violate("authorize: ext_agent_host_id=%q looks like a username; it must be an opaque urn:uuid: v4 generated once and stored, not anything derived from the account", v)
	case !ch19AgentHostIDRe.MatchString(v):
		s.violate("authorize: ext_agent_host_id=%q is not of the form urn:uuid:<canonical v4 uuid>; the version nibble must be 4, meaning randomly generated rather than derived from time or hardware", v)
	}
}

// ---------------------------------------------------------------------------
// Token endpoint
// ---------------------------------------------------------------------------

type ch19TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

func (s *ch19FakeOAuth) handleToken(w http.ResponseWriter, r *http.Request) {
	s.noteBearer(r)
	raw, _ := io.ReadAll(r.Body)
	form, _ := url.ParseQuery(string(raw))
	req := ch19TokenReq{
		GrantType:    form.Get("grant_type"),
		Code:         form.Get("code"),
		CodeVerifier: form.Get("code_verifier"),
		RedirectURI:  form.Get("redirect_uri"),
		ClientID:     form.Get("client_id"),
		RefreshToken: form.Get("refresh_token"),
		Resource:     form.Get("resource"),
		Scope:        form.Get("scope"),
		ContentType:  r.Header.Get("Content-Type"),
		RawBody:      string(raw),
	}

	// The token endpoint is form-encoded, not JSON. Sending JSON is a common
	// first attempt and produces an opaque 400 from the real provider, so the
	// fixture names it.
	if r.Method != http.MethodPost {
		s.mu.Lock()
		s.violate("token: method %s; the token endpoint is POST", r.Method)
		s.mu.Unlock()
	}
	if ct := req.ContentType; ct != "" && !strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		s.mu.Lock()
		s.violate("token: Content-Type %q; the token endpoint takes application/x-www-form-urlencoded, not JSON", ct)
		s.mu.Unlock()
	}

	switch req.GrantType {
	case "authorization_code":
		s.handleAuthCodeGrant(w, req)
	case "refresh_token":
		s.handleRefreshGrant(w, req)
	default:
		s.mu.Lock()
		s.obs.TokenHits = append(s.obs.TokenHits, req)
		s.violate("token: grant_type=%q; this provider supports only authorization_code and refresh_token", req.GrantType)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "unsupported grant_type")
	}
}

func (s *ch19FakeOAuth) handleAuthCodeGrant(w http.ResponseWriter, req ch19TokenReq) {
	s.mu.Lock()
	s.obs.TokenHits = append(s.obs.TokenHits, req)

	// Client identity first, because it is recorded even when the grant later
	// fails. Presenting the bootstrap id here means the agent never read the
	// registered id out of the callback.
	if req.ClientID == ch19OAuthBootstrapClientID {
		s.obs.SawDynamicClientAtToken = true
		s.violate("token: client_id=%q at the token endpoint; the bootstrap id is only valid for the first authorize, and the issued id %q must be used from then on", ch19OAuthBootstrapClientID, s.issuedClientID)
	} else if req.ClientID != s.issuedClientID {
		s.violate("token: client_id=%q; expected the issued client_id %q", req.ClientID, s.issuedClientID)
	}

	ac := s.codes[req.Code]
	switch {
	case req.Code == "":
		s.violate("token: code is missing from the authorization_code grant")
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "missing code")
		return
	case ac == nil:
		s.violate("token: code %q was never issued by this server", req.Code)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown code")
		return
	case ac.used:
		// Codes are single-use. A second redemption means either a retry that
		// should not have been retried, or an interception; the provider
		// cannot tell the two apart and so must refuse both.
		s.violate("token: code %q was already redeemed; an authorization code is single-use and a second redemption must be treated as a compromise", req.Code)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "code already used")
		return
	case time.Since(ac.issuedAt) > ch19OAuthCodeTTL:
		s.violate("token: code %q expired after %s", req.Code, ch19OAuthCodeTTL)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "code expired")
		return
	}

	// THE PKCE CHECK. Everything else in this file is scaffolding around it.
	//
	// The server kept the challenge from /authorize and never saw the
	// verifier. Now it hashes the verifier it was just handed and demands the
	// result equal that challenge. A client that invents a fresh random string
	// here, or that sends the challenge back as the verifier, or that sends
	// the verifier it generated but derived the challenge some other way, all
	// fail at this one comparison. There is no way to satisfy it except by
	// having actually implemented PKCE, which is exactly the property the
	// chapter is trying to force.
	if l := len(req.CodeVerifier); l < 43 || l > 128 {
		// Not fatal on its own, but a verifier outside RFC 7636's length range
		// usually means a hand-rolled "random" string with far less entropy
		// than the 32 bytes the spec is asking for.
		s.violate("token: code_verifier is %d characters; RFC 7636 requires 43 to 128, which is what 32 bytes of entropy looks like in base64url", l)
	}
	sum := sha256.Sum256([]byte(req.CodeVerifier))
	derived := base64.RawURLEncoding.EncodeToString(sum[:])
	if derived != ac.challenge {
		s.violate("token: PKCE failure. base64url(sha256(code_verifier)) is %q but the code_challenge sent to /authorize was %q; the verifier presented here does not correspond to the challenge, so the client is not actually implementing PKCE", derived, ac.challenge)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "code_verifier does not match code_challenge")
		return
	}

	// The redirect URI is re-sent and must match byte for byte. This closes
	// the hole where an attacker redeems a code against a redirect they
	// control, and byte equality is the rule because URL normalization is
	// where the subtle bypasses live.
	if req.RedirectURI != ac.redirectURI {
		s.violate("token: redirect_uri=%q does not byte-match the %q sent to /authorize; it must be repeated exactly, with no normalization", req.RedirectURI, ac.redirectURI)
	}
	if req.Resource != ch19OAuthResource {
		s.violate("token: resource=%q; it must be repeated here as %q or the issued access token is audienced for nothing", req.Resource, ch19OAuthResource)
	}

	ac.used = true
	resp := s.issueTokensLocked(ac.nonce)
	s.mu.Unlock()
	ch19WriteJSON(w, http.StatusOK, resp)
}

func (s *ch19FakeOAuth) handleRefreshGrant(w http.ResponseWriter, req ch19TokenReq) {
	// Enter the in-flight window before sleeping, so overlapping requests are
	// visible to each other in the high-water mark.
	s.mu.Lock()
	s.obs.RefreshHits = append(s.obs.RefreshHits, req)
	s.inFlight++
	if s.inFlight > s.obs.MaxConcurrentRefresh {
		s.obs.MaxConcurrentRefresh = s.inFlight
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.inFlight--
		s.mu.Unlock()
	}()

	// The delay is outside the lock on purpose. Holding mu across the sleep
	// would serialize refreshes inside the server, so two racing clients would
	// never collide and the race the grader is hunting would be invisible. The
	// window has to be real for the detection to mean anything.
	time.Sleep(ch19OAuthRefreshDelay)

	s.mu.Lock()
	s.refreshCount++
	first := s.refreshCount == 1

	if req.ClientID == ch19OAuthBootstrapClientID {
		s.obs.SawDynamicClientAtToken = true
		s.violate("refresh: client_id=%q; the bootstrap id is not valid for refresh, use the issued %q", ch19OAuthBootstrapClientID, s.issuedClientID)
	} else if req.ClientID != s.issuedClientID {
		s.violate("refresh: client_id=%q; expected the issued client_id %q", req.ClientID, s.issuedClientID)
	}

	// Injected terminal failure. This is a server-side fault, not a client
	// mistake, so it records no violation: the thing being graded is how the
	// client reacts. invalid_grant on a refresh is terminal, and the only
	// correct response is to drop the session and re-authorize. A client that
	// retries will retry forever.
	if first && s.opts.RefreshFailsWith != "" {
		code := s.opts.RefreshFailsWith
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, code, "injected refresh failure")
		return
	}

	switch {
	case req.RefreshToken == "":
		s.violate("refresh: refresh_token is missing")
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_request", "missing refresh_token")
		return

	case s.retiredRefresh[req.RefreshToken]:
		// Rotation makes reuse detectable, and detectable reuse is the only
		// defense a public client has against a stolen refresh token. Two
		// honest causes: a replay, or the client firing two refreshes at once
		// and the loser presenting a token the winner already retired. Both
		// are bugs, and the provider's correct response to either is to refuse.
		s.obs.RefreshReuseDetected = true
		s.violate("refresh: refresh_token %q was already rotated out and is dead; refresh tokens are single-use, so this is either a replay or two refreshes running concurrently against one session", req.RefreshToken)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "refresh token has been rotated")
		return

	case req.RefreshToken != s.liveRefresh:
		s.violate("refresh: refresh_token %q is not the current one for this session", req.RefreshToken)
		s.mu.Unlock()
		ch19WriteOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown refresh token")
		return
	}

	// Retire before issuing, so there is no instant in which two tokens are
	// simultaneously live.
	s.retiredRefresh[req.RefreshToken] = true

	// A refreshed session reuses the nonce from the authorization request that
	// established it. Real providers often omit the nonce from a refreshed
	// id_token, since no fresh authorization request happened to bind it to;
	// the fixture keeps it so that nonce handling is observable on both paths.
	nonce := ""
	if n := len(s.obs.Nonces); n > 0 {
		nonce = s.obs.Nonces[n-1]
	}
	resp := s.issueTokensLocked(nonce)
	s.mu.Unlock()
	ch19WriteJSON(w, http.StatusOK, resp)
}

// issueTokensLocked mints a fresh triple. Callers hold mu.
func (s *ch19FakeOAuth) issueTokensLocked(nonce string) ch19TokenResponse {
	s.seq++
	// Access tokens are opaque here. They do not need to be JWTs, and making
	// them obviously synthetic means a check can grep a transcript or a log
	// file for the literal prefix and prove whether the credential leaked.
	access := fmt.Sprintf("fake-access-%d-%s", s.seq, ch19OAuthRandomToken(12))
	refresh := fmt.Sprintf("fake-refresh-%d-%s", s.seq, ch19OAuthRandomToken(12))
	s.obs.IssuedAccessTokens = append(s.obs.IssuedAccessTokens, access)
	s.obs.IssuedRefreshTokens = append(s.obs.IssuedRefreshTokens, refresh)
	s.liveRefresh = refresh

	ttl := s.opts.AccessTokenTTL
	if ttl <= 0 {
		ttl = ch19OAuthDefaultAccessTTL
	}

	granted := []string{"openid", "profile", "email", "offline_access"}
	if !s.opts.OmitPlanScope {
		granted = append(granted, ch19OAuthPlanScope)
	}

	idNonce := nonce
	if s.opts.WrongNonce {
		// Not a random value: a plausible-looking one. A client that compares
		// nonces catches it; a client that merely checks the claim is present
		// and non-empty does not, and that second client has a replay hole.
		idNonce = "nonce_" + ch19OAuthRandomToken(12)
	}

	return ch19TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		IDToken:      s.signIDTokenLocked(idNonce),
		TokenType:    "Bearer",
		ExpiresIn:    int(ttl / time.Second),
		Scope:        strings.Join(granted, " "),
	}
}

// ---------------------------------------------------------------------------
// id_token
// ---------------------------------------------------------------------------

type ch19JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

type ch19IDTokenClaims struct {
	Iss           string `json:"iss"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
	Nonce         string `json:"nonce,omitempty"`
}

// signIDTokenLocked builds and signs a real RS256 JWT. Callers hold mu.
//
// A JWT is three base64url-nopad segments joined by dots, where the signature
// covers the ASCII bytes of the first two joined by a dot. That is the whole
// format. Writing it out here rather than importing a library is deliberate:
// the chapter's point is that signature verification is a thing you do, not a
// thing you import, and a reader who has seen the signing side in forty lines
// is much harder to talk out of implementing the verifying side.
func (s *ch19FakeOAuth) signIDTokenLocked(nonce string) string {
	now := time.Now()
	// The kid always names the PUBLISHED key, even under UnsignedKeyMismatch
	// where a different key does the signing. That is the realistic shape of
	// an attack: the header points at a key the client can fetch and trust,
	// and only the signature arithmetic gives it away.
	header := ch19JWTHeader{Alg: "RS256", Typ: "JWT", Kid: s.kid}
	claims := ch19IDTokenClaims{
		Iss:           s.srv.URL,
		Aud:           s.issuedClientID,
		Sub:           ch19OAuthSubject,
		Email:         ch19OAuthEmail,
		EmailVerified: true,
		Exp:           now.Add(time.Hour).Unix(),
		Iat:           now.Unix(),
		Nonce:         nonce,
	}
	hb, err := json.Marshal(header)
	if err != nil {
		panic("ch19 fake oauth: marshaling JWT header: " + err.Error())
	}
	cb, err := json.Marshal(claims)
	if err != nil {
		panic("ch19 fake oauth: marshaling JWT claims: " + err.Error())
	}
	signingInput := base64.RawURLEncoding.EncodeToString(hb) + "." +
		base64.RawURLEncoding.EncodeToString(cb)

	// RS256 is RSASSA-PKCS1-v1_5 over a SHA-256 digest. Note that
	// SignPKCS1v15 wants the digest, not the message.
	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.signKey, crypto.SHA256, digest[:])
	if err != nil {
		panic("ch19 fake oauth: signing id_token: " + err.Error())
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// ---------------------------------------------------------------------------
// Revocation
// ---------------------------------------------------------------------------

func (s *ch19FakeOAuth) handleRevoke(w http.ResponseWriter, r *http.Request) {
	s.noteBearer(r)
	raw, _ := io.ReadAll(r.Body)
	form, _ := url.ParseQuery(string(raw))
	rec := ch19RevokeReq{
		Token:         form.Get("token"),
		TokenTypeHint: form.Get("token_type_hint"),
		RawBody:       string(raw),
	}

	s.mu.Lock()
	s.obs.RevokeHits = append(s.obs.RevokeHits, rec)
	if r.Method != http.MethodPost {
		s.violate("revoke: method %s; revocation is POST", r.Method)
	}
	if rec.Token == "" {
		s.violate("revoke: the token form field is empty; RFC 7009 names the field \"token\"")
	}
	if rec.Token != "" {
		s.retiredRefresh[rec.Token] = true
		if rec.Token == s.liveRefresh {
			s.liveRefresh = ""
		}
		s.deadAccess[rec.Token] = true
	}
	s.mu.Unlock()

	// RFC 7009 says to return 200 even for an unknown token, so that the
	// endpoint cannot be used as an oracle for which tokens exist.
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// noteBearer records any use of a credential this server has already killed.
// The fixture guards no protected resource, so this is the only place a dead
// access token can be caught, and catching it is worth the four lines: a
// client that keeps presenting an expired token is a client whose refresh
// logic never ran.
func (s *ch19FakeOAuth) noteBearer(r *http.Request) {
	tok, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || tok == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.deadAccess[tok] {
		s.violate("%s: presented access token %q after it was expired or revoked; the client must refresh before reusing a dead credential", r.URL.Path, tok)
	}
}

type ch19OAuthErrorBody struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func ch19WriteOAuthError(w http.ResponseWriter, status int, code, desc string) {
	ch19WriteJSON(w, status, ch19OAuthErrorBody{Error: code, ErrorDescription: desc})
}

func ch19WriteJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "marshal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func ch19Contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// ch19OAuthRandomToken returns n bytes of crypto-random data in base64url. It
// panics on failure because a grader that silently falls back to predictable
// tokens would quietly stop testing the thing it exists to test.
func ch19OAuthRandomToken(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("ch19 fake oauth: crypto/rand: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
