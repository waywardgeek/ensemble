package oauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// BootstrapClientID is the well-known entrypoint used for a first-time
// registration.
//
// First-party rule, verbatim: "dynamic_agent_client is the first-time
// registration entrypoint, not the client ID to save or use for token
// exchange." It is sent on the authorize request of a brand-new client and
// never anywhere else. Persisting it would mean every sign-in looked like a
// first one, re-registering endlessly and orphaning the registration the user
// already approved.
const BootstrapClientID = "dynamic_agent_client"

// DefaultAgentName is the agent_name_hint sent at registration.
//
// The docs require "the app's actual name, used consistently" — consistently
// because the user sees it on the consent screen and in their account's list
// of authorized apps. A name that changes between versions turns one
// recognisable entry into a pile of unrecognisable ones that the user cannot
// safely revoke.
const DefaultAgentName = "ensemble-agent"

// defaultCallbackTimeout bounds how long the loopback listener waits.
//
// Long enough for a real human to find a password manager and complete MFA,
// short enough that an abandoned sign-in does not leave a socket and a
// goroutine alive forever.
const defaultCallbackTimeout = 5 * time.Minute

// defaultRefreshSkew is how far ahead of expiry a token is refreshed.
const defaultRefreshSkew = 2 * time.Minute

// Config describes how to run the Sign in with ChatGPT flow.
//
// Every field has a working default, so the zero value plus a Store is a
// usable configuration. The fields exist to be overridden by deployments
// (issuer, browser behaviour) and by tests (clock, HTTP client) through the
// same seams, which is how we know the seams are real rather than test-only
// scaffolding.
type Config struct {
	// Issuer overrides the OIDC issuer. Empty means consult
	// OPENAI_OIDC_ISSUER, then fall back to DefaultIssuer.
	Issuer string

	// AgentName is the agent_name_hint. Empty means DefaultAgentName.
	AgentName string

	// Store persists credentials. Required.
	Store *Store

	// HostIDPath is where the ext_agent_host_id lives. Empty means
	// DefaultHostIDPath.
	HostIDPath string

	// HostID overrides the persisted host id outright. Normally empty.
	HostID string

	// HTTPClient is used for discovery, JWKS and token calls. Empty means a
	// client with a sane timeout.
	HTTPClient *http.Client

	// Opener decides how the user reaches the authorization URL. Empty means
	// CommandOpener, unless NoBrowser (or the environment) asks for the
	// print-only path.
	Opener Opener

	// NoBrowser forces the print-the-URL path for headless and SSH use.
	NoBrowser bool

	// Output receives human-facing progress text. Empty means os.Stderr —
	// stderr rather than stdout so that a tool piping its real output
	// somewhere does not get a sign-in prompt mixed into it.
	Output io.Writer

	// Now supplies the current time. Empty means time.Now.
	Now func() time.Time

	// RefreshSkew is how far before expiry to refresh. Zero means
	// defaultRefreshSkew.
	RefreshSkew time.Duration

	// CallbackTimeout bounds the wait for the browser redirect. Zero means
	// defaultCallbackTimeout.
	CallbackTimeout time.Duration
}

// normalize fills in defaults and validates what cannot be defaulted.
func (c Config) normalize() (Config, error) {
	if c.Store == nil {
		return c, errors.New("oauth: Config.Store is required")
	}
	if c.Issuer == "" {
		c.Issuer = Issuer()
	}
	if c.AgentName == "" {
		c.AgentName = DefaultAgentName
	}
	if c.HTTPClient == nil {
		// A credential provider that can hang forever can hang a turn
		// forever, so the timeout is applied here rather than trusting every
		// caller to have supplied a bounded context.
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if c.Output == nil {
		c.Output = os.Stderr
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.RefreshSkew == 0 {
		c.RefreshSkew = defaultRefreshSkew
	}
	if c.CallbackTimeout == 0 {
		c.CallbackTimeout = defaultCallbackTimeout
	}
	if c.HostIDPath == "" && c.HostID == "" {
		p, err := DefaultHostIDPath()
		if err != nil {
			return c, err
		}
		c.HostIDPath = p
	}
	if c.Opener == nil {
		if c.NoBrowser || headlessRequested() {
			c.Opener = PrintURLOpener(c.Output)
		} else {
			c.Opener = CommandOpener(c.Output)
		}
	}
	return c, nil
}

// Flow runs the interactive authorization.
type Flow struct {
	cfg Config
}

// NewFlow validates cfg and returns a Flow.
func NewFlow(cfg Config) (*Flow, error) {
	norm, err := cfg.normalize()
	if err != nil {
		return nil, err
	}
	return &Flow{cfg: norm}, nil
}

// callbackResult is what the loopback listener captured from the redirect.
type callbackResult struct {
	code     string
	state    string
	clientID string
	errCode  string
	errDesc  string
}

// Authorize runs the full interactive sign-in and persists the result.
//
// A note on what "success" means here. If the user signs in but the issuer
// does not grant chatgpt.tokens.use.direct, Authorize SUCCEEDS and stores the
// credential. That is the first-party guidance: "retain the sign-in, mark
// plan usage disabled, offer another billing path, do not infer." Discarding
// the credential and looping the OAuth flow against an ineligible account
// would spin forever. The entitlement is enforced where it matters — at
// Provider.GetBearerToken, immediately before inference — so callers should
// check Credentials.HasDirectTokenScope before promising the user anything.
func (f *Flow) Authorize(ctx context.Context) (*Credentials, error) {
	// The host id is resolved and persisted FIRST, before any network call,
	// because the docs require it to exist before the first sign-in.
	hostID := f.cfg.HostID
	if hostID == "" {
		var err error
		hostID, err = LoadOrCreateHostID(f.cfg.HostIDPath)
		if err != nil {
			return nil, err
		}
	}

	// Load any existing registration. A re-authorization must reuse the
	// issued client id; only a host that has never registered sends the
	// bootstrap id.
	var pendingClientID string
	existing, err := f.cfg.Store.Load()
	switch {
	case err == nil:
		pendingClientID = existing.ClientID
	case errors.Is(err, ErrNoCredentials):
		// First run. Nothing to reuse.
	default:
		return nil, err
	}

	disco, err := FetchDiscovery(ctx, f.cfg.HTTPClient, f.cfg.Issuer)
	if err != nil {
		return nil, err
	}

	// Fresh per attempt, all three. Reusing any of them across attempts would
	// let a value captured from one attempt be used to complete another.
	pkce, err := newPKCE()
	if err != nil {
		return nil, err
	}
	state, err := randomURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("oauth: generating state: %w", err)
	}
	nonce, err := randomURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("oauth: generating nonce: %w", err)
	}

	// Loopback listener on an EPHEMERAL port.
	//
	// 127.0.0.1 and never "localhost": localhost resolves through the
	// resolver and may land on ::1, or on something else entirely if
	// /etc/hosts has been tampered with. The literal address cannot be
	// redirected. Binding to the loopback interface specifically — rather
	// than 0.0.0.0 — means no other machine on the network can reach the
	// listener and race us to the authorization code.
	//
	// Port 0 asks the kernel for a free port. A fixed port would collide with
	// a second agent, and worse, would let an unrelated local process squat
	// the port in advance and receive the code.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("oauth: opening loopback listener: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/auth/callback", port)

	authClientID := pendingClientID
	isRegistration := authClientID == ""
	if isRegistration {
		authClientID = BootstrapClientID
	}

	authURL := f.authorizationURL(disco, authClientID, redirectURI, state, nonce, pkce, hostID, isRegistration, existing)

	results := make(chan callbackResult, 1)
	srv := &http.Server{
		Handler:           f.callbackHandler(results),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = srv.Serve(listener) }()
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := f.cfg.Opener(ctx, authURL); err != nil {
		return nil, fmt.Errorf("oauth: directing the user to the authorization URL: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, f.cfg.CallbackTimeout)
	defer cancel()

	var cb callbackResult
	select {
	case cb = <-results:
	case <-waitCtx.Done():
		return nil, fmt.Errorf("oauth: timed out waiting for the authorization callback: %w", waitCtx.Err())
	}

	// SECURITY, and it comes first: validate state before looking at anything
	// else in the callback, including the error parameter.
	//
	// state is the CSRF defence. Without it, an attacker can send the user's
	// browser to our loopback redirect carrying a code minted for the
	// ATTACKER's account; we would exchange it, store it, and then spend the
	// attacker's plan while believing we were signed in as the user — or, in
	// the other direction, hand the user's agent to an account the attacker
	// controls. Constant-time comparison avoids leaking the expected value
	// through timing.
	if subtle.ConstantTimeCompare([]byte(cb.state), []byte(state)) != 1 {
		return nil, ErrStateMismatch
	}

	// Only now is the error parameter trustworthy enough to act on. An
	// explicit denial stops the flow without exchanging anything — there is
	// no code, and attempting an exchange would just produce a confusing
	// second error.
	if cb.errCode != "" {
		if cb.errCode == "access_denied" {
			return nil, ErrAccessDenied
		}
		return nil, fmt.Errorf("oauth: authorization failed: %s: %s", cb.errCode, cb.errDesc)
	}
	if cb.code == "" {
		return nil, errors.New("oauth: callback carried neither an authorization code nor an error")
	}

	issuedClientID, err := resolveClientID(pendingClientID, cb.clientID)
	if err != nil {
		return nil, err
	}

	tr, err := exchangeCode(ctx, f.cfg.HTTPClient, disco.TokenEndpoint, issuedClientID, cb.code, redirectURI, pkce.Verifier)
	if err != nil {
		return nil, err
	}

	now := f.cfg.Now()
	ks := newKeySet(f.cfg.HTTPClient, disco.JWKSURI, f.cfg.Now)

	// The ID token is validated against the ISSUED client id, because that is
	// the client the tokens were minted for.
	var claims *IDTokenClaims
	if tr.IDToken != "" {
		claims, err = verifyIDToken(ctx, ks, tr.IDToken, verifyParams{
			Issuer:   disco.Issuer,
			ClientID: issuedClientID,
			Nonce:    nonce,
			Now:      now,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// openid was requested, so an ID token is required. Proceeding
		// without one would mean we had authenticated nobody.
		return nil, errors.New("oauth: token response contained no id_token")
	}

	creds := &Credentials{
		Issuer:        disco.Issuer,
		ClientID:      issuedClientID,
		HostID:        hostID,
		AccessToken:   tr.AccessToken,
		RefreshToken:  tr.RefreshToken,
		IDToken:       tr.IDToken,
		GrantedScopes: splitScopes(tr.Scope),
		ExpiresAt:     expiryFrom(tr, now),
		Subject:       claims.Subject,
		Email:         claims.Email,
	}

	if err := f.cfg.Store.Save(creds); err != nil {
		return nil, err
	}

	if !creds.HasDirectTokenScope() {
		// Tell the user plainly. This is not an error — the sign-in is real
		// and is being kept — but silently storing a credential that cannot
		// bill anything would leave them wondering why inference fails later.
		fmt.Fprintf(f.cfg.Output, "\nSigned in as %s, but the %s scope was not granted.\nChatGPT plan usage is disabled; configure another billing path.\n",
			displayName(creds), ScopeDirectTokens)
	}

	return creds.clone(), nil
}

// authorizationURL builds the authorize request.
func (f *Flow) authorizationURL(disco *Discovery, clientID, redirectURI, state, nonce string, pkce pkceParams, hostID string, isRegistration bool, existing *Credentials) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {joinScopes(DefaultScopes())},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {codeChallengeMethodS256},

		// RFC 8707 resource indicator. The route requires it, and the same
		// value has to be repeated at code exchange: it names the API the
		// resulting token is allowed to call, so a token minted for one
		// resource cannot be replayed against another.
		"resource": {resourceIndicator},

		// ext_agent_host_id and agent_name_hint are REQUIRED for a first-time
		// registration. They are sent on every authorization, not just the
		// first, because a single client_id may span many hosts for the same
		// user and workspace: the issuer needs the host id to tell which of
		// those hosts is asking, and omitting it on a re-auth would make this
		// host indistinguishable from the others.
		"ext_agent_host_id": {hostID},
		"agent_name_hint":   {f.cfg.AgentName},
	}

	// id_token_hint on a re-authorization lets the issuer skip the account
	// chooser and re-confirm the SAME account, which is why the previous ID
	// token is retained after validation rather than discarded. Getting
	// silently re-authorized as a different account is a real hazard for a
	// user with both a personal and a work ChatGPT login.
	if !isRegistration && existing != nil && existing.IDToken != "" {
		q.Set("id_token_hint", existing.IDToken)
	}

	return disco.AuthorizationEndpoint + "?" + q.Encode()
}

// callbackHandler serves the loopback redirect.
func (f *Flow) callbackHandler(results chan<- callbackResult) http.Handler {
	mux := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		res := callbackResult{
			code:     q.Get("code"),
			state:    q.Get("state"),
			clientID: q.Get("client_id"),
			errCode:  q.Get("error"),
			errDesc:  q.Get("error_description"),
		}

		// Non-blocking send into a buffered channel of one. Only the FIRST
		// callback is acted on: a second request — a browser prefetch, a
		// refresh, or an attacker spraying the port — must not be able to
		// overwrite a result the flow is already processing, and must not
		// block this handler forever once nobody is receiving.
		select {
		case results <- res:
		default:
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Deliberately static HTML. Nothing from the query string is echoed
		// back into the page: the values here are attacker-influenceable, and
		// reflecting them would be a cross-site scripting bug in a page
		// served from the user's own loopback interface.
		if res.errCode != "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `<!doctype html><title>Sign-in failed</title>
<h1>Sign-in failed</h1><p>You can close this window and return to the terminal.</p>`)
			return
		}
		io.WriteString(w, `<!doctype html><title>Signed in</title>
<h1>Signed in</h1><p>You can close this window and return to the terminal.</p>`)
	}
	mux.HandleFunc("/auth/callback", handler)
	// Accept the redirect on any path. Issuers and enterprise fronts
	// occasionally normalise the redirect path, and failing the sign-in over
	// a trailing slash helps nobody — the state parameter, not the path, is
	// what makes the callback trustworthy.
	mux.HandleFunc("/", handler)
	return mux
}

// resolveClientID decides which client id to exchange and persist, and
// refuses the substitution attack.
//
// The rules come straight from the first-party docs:
//
//   - A successful NEW registration returns code, state and the issued
//     client_id. That issued id is what we exchange and save.
//   - On REAUTHORIZATION the callback may omit client_id, in which case we
//     "retain the exact client ID already associated with the pending
//     request".
//   - "If the callback supplies a different client ID, reject the result
//     rather than replacing the selected account's registration."
//
// That last rule is the security-critical one. The callback arrives over
// plain HTTP on loopback and its contents are not authenticated; if we let it
// overwrite the stored client id, anyone able to reach the listener during a
// sign-in could repoint the registration at a client they control, and every
// token this host subsequently obtains would be minted for them.
func resolveClientID(pending, fromCallback string) (string, error) {
	if fromCallback == "" {
		if pending == "" {
			// A new registration is REQUIRED to return the issued id. Without
			// one there is nothing to exchange with except the bootstrap id,
			// which the docs forbid using for token exchange.
			return "", errors.New("oauth: new registration callback did not return a client_id")
		}
		return pending, nil
	}

	// Never accept the bootstrap entrypoint as an identity, whatever the
	// callback claims. Persisting it would make every future sign-in look
	// like a first-time registration.
	if fromCallback == BootstrapClientID {
		return "", fmt.Errorf("oauth: callback returned the bootstrap client id %q as the issued client id", BootstrapClientID)
	}

	if pending != "" && subtle.ConstantTimeCompare([]byte(pending), []byte(fromCallback)) != 1 {
		return "", ErrClientIDMismatch
	}
	return fromCallback, nil
}

// displayName renders the signed-in account for human-facing messages,
// preferring the email and falling back to the stable subject.
func displayName(c *Credentials) string {
	if c.Email != "" {
		return c.Email
	}
	return c.Subject
}
