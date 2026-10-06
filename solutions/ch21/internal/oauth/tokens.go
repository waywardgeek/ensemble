package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tokenResponse is the token endpoint's success body (RFC 6749 §5.1).
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`

	// Scope is the GRANTED scope set. This is the authoritative answer to
	// "what may we do", and it can legitimately differ from what we asked
	// for.
	Scope string `json:"scope"`
}

// postToken performs a form POST to the token endpoint and decodes the result.
//
// There is no client secret and no Authorization header: this is a public
// client, and the discovery document lists "none" among the supported
// endpoint auth methods. The proof that we are the legitimate redeemer of
// this code is the PKCE verifier, not a secret we would have had to ship
// inside a binary the user can read.
func postToken(ctx context.Context, client *http.Client, endpoint string, form url.Values) (*tokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("oauth: building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth: calling token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth: reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseTokenError(resp.StatusCode, body)
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, fmt.Errorf("oauth: parsing token response: %w", err)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("oauth: token response contained no access_token")
	}
	// RFC 6750 defines the bearer type case-insensitively. An empty type is
	// tolerated because some issuers omit it; a type we do not understand is
	// not, because we would be about to put it in an Authorization header
	// under the wrong scheme.
	if tr.TokenType != "" && !strings.EqualFold(tr.TokenType, "bearer") {
		return nil, fmt.Errorf("oauth: token response has unsupported token_type %q", tr.TokenType)
	}
	return &tr, nil
}

// parseTokenError turns a non-200 token response into a *TokenError.
//
// Two shapes are handled. The RFC 6749 §5.2 object — {"error": "...",
// "error_description": "..."} — is the normal case and the one whose Code
// drives the terminal/transient decision.
//
// The other shape is {"detail": "..."}, which the first-party error docs warn
// is returned by direct-admission failures before a request starts. It has no
// machine-readable code. The important consequence, and the reason this is
// not a one-line json.Unmarshal: a parser that assumes error.code always
// exists crashes on exactly the failure it most needs to report, and a
// classifier that defaults to "terminal" when it cannot parse would discard a
// working refresh token because of a transient 503.
func parseTokenError(status int, body []byte) *TokenError {
	te := &TokenError{Status: status}

	var rfc struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Detail           string `json:"detail"`
	}
	if err := json.Unmarshal(body, &rfc); err == nil {
		te.Code = rfc.Error
		te.Description = rfc.ErrorDescription
		if te.Description == "" {
			te.Description = rfc.Detail
		}
	}
	if te.Code == "" && te.Description == "" {
		// Keep a bounded excerpt of whatever did come back, so the operator
		// has something to go on, without pasting a megabyte of HTML error
		// page into a log.
		excerpt := strings.TrimSpace(string(body))
		if len(excerpt) > 200 {
			excerpt = excerpt[:200] + "..."
		}
		te.Description = excerpt
	}
	return te
}

// exchangeCode redeems an authorization code for tokens.
//
// client_id here is the ISSUED client id — the one the callback handed us on
// a new registration, or the one already on record for a reauthorization.
// Sending dynamic_agent_client at this point would be wrong: the docs are
// explicit that it is "the first-time registration entrypoint, not the client
// ID to save or use for token exchange".
//
// redirect_uri must be byte-identical to the one in the authorization
// request. The server compares them to ensure the code is being redeemed by
// the client it was issued to, so a cosmetic difference — a trailing slash,
// localhost instead of 127.0.0.1 — fails the exchange with invalid_grant.
func exchangeCode(ctx context.Context, client *http.Client, endpoint, clientID, code, redirectURI, verifier string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {verifier},
		// The same resource indicator sent at authorization. The server
		// compares them, so a token exchange that omits it is refused.
		"resource": {resourceIndicator},
	}
	return postToken(ctx, client, endpoint, form)
}

// refreshGrant exchanges a refresh token for a new access token.
//
// The refresh token is rotating: a successful call usually returns a NEW
// refresh token and invalidates the one just presented. That is why the
// caller must persist the result atomically and must never retry a refresh
// with the old token after a failure of unknown outcome — presenting a
// already-redeemed token is what refresh_token_reused detects, and the
// issuer's response to suspected theft is to kill the entire chain.
func refreshGrant(ctx context.Context, client *http.Client, endpoint, clientID, refreshToken string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}
	return postToken(ctx, client, endpoint, form)
}

// expiryFrom converts the response's relative lifetime into an absolute
// instant.
//
// Absolute, because a relative lifetime is only meaningful at the moment it
// was received, and the credential file outlives that moment. A default is
// applied when the server omits expires_in so that a missing field does not
// produce a zero time that NeedsRefresh would read as "refresh on every
// single request".
func expiryFrom(tr *tokenResponse, now time.Time) time.Time {
	seconds := tr.ExpiresIn
	if seconds <= 0 {
		seconds = int64(defaultTokenLifetime / time.Second)
	}
	return now.Add(time.Duration(seconds) * time.Second)
}

// defaultTokenLifetime is assumed when the server omits expires_in. It is
// deliberately short: guessing low costs an extra refresh, guessing high
// costs failed requests.
const defaultTokenLifetime = 15 * time.Minute
