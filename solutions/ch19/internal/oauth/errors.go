// Package oauth implements "Sign in with ChatGPT" (SIWC): the OAuth 2.0 +
// OIDC authorization-code flow, with PKCE, that lets this agent borrow a
// user's consumer ChatGPT plan instead of billing a metered API key.
//
// The flow is an ordinary public-client authorization-code flow in its bones,
// but it has three peculiarities that are not negotiable and that the rest of
// this package exists to get right:
//
//  1. Registration rides the authorize round trip. There is no RFC 7591
//     registration endpoint — the discovery document does not advertise one.
//     A first-time client authorizes as the well-known bootstrap id
//     dynamic_agent_client and is handed its real, issued client_id back on
//     the callback. See flow.go.
//
//  2. A valid ID token does not authorize plan usage. Sign-in and entitlement
//     are separate facts. The gate is the scope chatgpt.tokens.use.direct as
//     reported by the TOKEN RESPONSE, not by the ID token and not by what we
//     asked for. See provider.go.
//
//  3. Refresh tokens rotate, and a rotated-away token is gone for good. Six
//     distinct error codes mean the grant is dead and only a fresh
//     authorization can revive it. They are enumerated below.
package oauth

import (
	"errors"
	"fmt"
)

// Sentinel errors. Callers are expected to branch on these with errors.Is
// rather than on strings, because the strings are for humans and will change.
var (
	// ErrNoCredentials means nothing has been stored yet: the user has never
	// signed in on this host. It is not a failure, it is a prompt to run the
	// authorization flow.
	ErrNoCredentials = errors.New("oauth: no stored ChatGPT credentials; sign-in required")

	// ErrReauthRequired means the stored grant is dead and cannot be revived
	// by any refresh. The access and refresh tokens have been cleared. The
	// issued client_id and the host id are deliberately RETAINED, because
	// re-authorization must reuse them (a new registration would orphan the
	// account's existing one).
	ErrReauthRequired = errors.New("oauth: authorization is no longer valid; full re-authorization required")

	// ErrPlanScopeNotGranted means sign-in succeeded but the user's account
	// was not granted chatgpt.tokens.use.direct. Per the first-party error
	// guidance the correct response is to RETAIN the sign-in and disable plan
	// usage, not to discard the credential and loop the OAuth flow — looping
	// would spin forever against an account that is simply not eligible.
	ErrPlanScopeNotGranted = errors.New("oauth: scope " + ScopeDirectTokens + " was not granted; ChatGPT plan usage is disabled")

	// ErrStateMismatch means the callback's state did not match the pending
	// request. This is the CSRF defence: without it an attacker can feed us a
	// code minted for their own account and silently sign us in as them.
	ErrStateMismatch = errors.New("oauth: callback state did not match the pending authorization request")

	// ErrClientIDMismatch means the callback tried to hand us a client_id
	// different from the one already associated with the pending request.
	//
	// First-party rule, verbatim: "If the callback supplies a different client
	// ID, reject the result rather than replacing the selected account's
	// registration." Accepting it would let whoever controls the redirect
	// repoint a working registration at a client they control, which is an
	// account-takeover primitive, not a configuration update.
	ErrClientIDMismatch = errors.New("oauth: callback supplied a different client_id than the one on record")

	// ErrAccessDenied means the user declined at the consent screen. The state
	// is still validated first, but no code is exchanged because none was
	// issued.
	ErrAccessDenied = errors.New("oauth: the user denied the authorization request")
)

// TokenError is the RFC 6749 §5.2 error object returned by the token endpoint.
//
// It is a distinct type rather than a formatted string because the Code is
// load-bearing: it decides whether a refresh failure is a transient blip worth
// retrying or a terminal condition that must interrupt the user for a fresh
// sign-in. Flattening it to text would force that decision to be made by
// substring matching.
type TokenError struct {
	// Code is the machine-readable "error" field. It may be empty if the
	// server returned a non-conforming body — see the note on pre-stream
	// errors in Terminal.
	Code string
	// Description is the optional human-readable "error_description".
	Description string
	// Status is the HTTP status that carried the error.
	Status int
}

func (e *TokenError) Error() string {
	switch {
	case e.Code != "" && e.Description != "":
		return fmt.Sprintf("oauth: token endpoint returned %s (HTTP %d): %s", e.Code, e.Status, e.Description)
	case e.Code != "":
		return fmt.Sprintf("oauth: token endpoint returned %s (HTTP %d)", e.Code, e.Status)
	default:
		return fmt.Sprintf("oauth: token endpoint failed with HTTP %d", e.Status)
	}
}

// terminalRefreshCodes are the six error codes that mean the refresh token is
// dead and no amount of retrying will help. The user must authorize again.
//
// There are SIX, not the three that a reading of the generic OAuth spec would
// suggest. invalid_grant is the standard one; the other five are specific to
// rotating refresh tokens, and refresh_token_reused in particular is a
// security signal rather than an expiry — it means the token we just presented
// had already been redeemed, which is how a rotating-token server detects
// theft. It responds by invalidating the whole chain, so there is nothing left
// to retry with.
//
// Treating any of these as retryable produces the worst possible behaviour:
// a client that hammers the token endpoint with a credential that is
// permanently dead, while never telling the user the one thing that would fix
// it.
var terminalRefreshCodes = map[string]bool{
	"invalid_grant":             true,
	"invalid_refresh_token":     true,
	"token_expired":             true,
	"refresh_token_expired":     true,
	"refresh_token_invalidated": true,
	"refresh_token_reused":      true,
}

// Terminal reports whether this error ends the grant's life.
//
// Note that it keys off Code alone and never off Status. A 400 is terminal
// when it carries invalid_grant and transient when it carries something else,
// so the HTTP status is not a usable proxy. An empty Code is treated as NOT
// terminal: the first-party docs warn that some failures arrive as
// {"detail": "..."} with no machine-readable code at all, and destroying a
// user's working refresh token on the strength of an unparseable body would
// turn a transient outage into a forced re-login.
func (e *TokenError) Terminal() bool {
	return terminalRefreshCodes[e.Code]
}

// IsTerminalRefreshError reports whether err is a token-endpoint error that
// requires full re-authorization. It is the exported form of the rule above,
// for callers that hold an error they did not construct.
func IsTerminalRefreshError(err error) bool {
	var te *TokenError
	if errors.As(err, &te) {
		return te.Terminal()
	}
	return errors.Is(err, ErrReauthRequired)
}
