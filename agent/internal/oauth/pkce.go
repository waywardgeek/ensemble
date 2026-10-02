package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// Scope names requested at authorization.
//
// The first four are standard OIDC. The last two are not advertised in
// auth.openai.com's discovery document under scopes_supported — that omission
// is expected for plan-gated surfaces and must NOT be "fixed" by dropping
// them, because dropping them is precisely how a client ends up signed in but
// unable to infer.
const (
	ScopeOpenID         = "openid"
	ScopeProfile        = "profile"
	ScopeEmail          = "email"
	ScopeOfflineAccess  = "offline_access"
	ScopeResourceInvoke = "resource.invoke"

	// ScopeDirectTokens is the entitlement that actually authorizes billing
	// inference against the user's ChatGPT plan. Everything else in the list
	// identifies the user; only this one pays for anything.
	ScopeDirectTokens = "chatgpt.tokens.use.direct"
)

// DefaultScopes returns the full scope set for Sign in with ChatGPT.
//
// It is a function returning a fresh slice rather than a package-level var
// because a package-level slice is mutable by any importer, and a caller that
// appended to it — or worse, sorted it in place — would silently change what
// every subsequent sign-in requests.
//
// offline_access is what makes the server issue a refresh token at all;
// without it the grant dies in an hour with no way back except another
// browser round trip.
func DefaultScopes() []string {
	return []string{
		ScopeOpenID,
		ScopeProfile,
		ScopeEmail,
		ScopeOfflineAccess,
		ScopeResourceInvoke,
		ScopeDirectTokens,
	}
}

// pkceParams carries a PKCE verifier and the challenge derived from it.
//
// PKCE exists because this is a public client on a loopback redirect: there
// is no client secret, so possession of the authorization code would
// otherwise be sufficient to redeem it. Any local process that can observe
// the redirect — a malicious app watching the loopback port, a shell history,
// a log — could steal the code and exchange it. Binding the code to a secret
// that never leaves this process closes that window: the thief has the code
// but not the verifier, and the token endpoint refuses the exchange.
type pkceParams struct {
	// Verifier is the high-entropy secret. It is sent ONLY to the token
	// endpoint, never to the authorization endpoint.
	Verifier string
	// Challenge is SHA-256(Verifier), base64url-encoded. This is what travels
	// through the browser, where it may be logged or observed.
	Challenge string
}

// codeChallengeMethodS256 is the only challenge method this package will ever
// send.
//
// auth.openai.com's discovery document advertises
// code_challenge_methods_supported: ["S256"] and nothing else. But the real
// reason is not that the server requires it: "plain" sends the verifier itself
// through the browser as the challenge, which defeats the entire mechanism —
// anyone who can see the challenge then holds the secret. A client that
// negotiates down to "plain" has PKCE in name only, so this package offers no
// way to select it.
const codeChallengeMethodS256 = "S256"

// newPKCE mints a fresh verifier and its S256 challenge.
//
// Fresh per attempt, always. Reusing a verifier across attempts would let a
// code stolen from one attempt be redeemed with a verifier captured from
// another.
func newPKCE() (pkceParams, error) {
	// 32 bytes of entropy encodes to 43 base64url characters, which sits at
	// the top of RFC 7636's 43–128 character range for a verifier.
	verifier, err := randomURLSafe(32)
	if err != nil {
		return pkceParams{}, fmt.Errorf("oauth: generating PKCE verifier: %w", err)
	}
	sum := sha256.Sum256([]byte(verifier))
	return pkceParams{
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
	}, nil
}

// randomURLSafe returns n bytes from the cryptographic RNG, base64url-encoded
// without padding.
//
// crypto/rand, never math/rand. These values are the state parameter, the
// nonce and the PKCE verifier; each one is a secret whose only job is to be
// unguessable. math/rand is seeded predictably and its output is recoverable
// from a handful of samples, which would make every one of those defences
// decorative. An error here is a broken system RNG and is propagated rather
// than papered over.
func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// hasScope reports whether granted contains want.
//
// Scope comparison is exact and case-sensitive: RFC 6749 defines scope tokens
// as case-sensitive strings, and a fuzzy match here would be a privilege
// check that can be fooled by a lookalike.
func hasScope(granted []string, want string) bool {
	for _, s := range granted {
		if s == want {
			return true
		}
	}
	return false
}

// splitScopes parses the space-delimited scope string used in OAuth requests
// and responses into a slice, discarding empties so that irregular spacing
// cannot produce a phantom "" scope.
func splitScopes(s string) []string {
	return strings.Fields(s)
}

// joinScopes renders a scope slice for the wire.
func joinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}
