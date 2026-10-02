package oauth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// algRS256 is the only JWS algorithm this package accepts for an ID token.
const algRS256 = "RS256"

// clockSkew is the tolerance applied to time-based claims.
//
// Some tolerance is necessary because the client's clock and the issuer's
// clock are independent, and a laptop that is forty seconds fast would
// otherwise reject every token it was handed. It is kept small: the leeway is
// also the window in which a genuinely expired token is still accepted.
const clockSkew = 60 * time.Second

// idTokenHeader is the JOSE header of the ID token.
type idTokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// audience handles the "aud" claim, which OIDC allows to be either a single
// string or an array of strings. Modelling it as a plain string would panic
// the decoder on a conforming token from an issuer that happens to use the
// array form.
type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var single string
	if err := json.Unmarshal(b, &single); err == nil {
		*a = audience{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return fmt.Errorf("oauth: aud claim is neither a string nor an array of strings")
	}
	*a = audience(many)
	return nil
}

// IDTokenClaims is the subset of ID token claims this package reads.
//
// The ID token answers "who signed in". It deliberately does NOT answer "may
// this agent bill the user's plan" — that question is settled by the granted
// scopes on the token response, and conflating the two is the single most
// likely way to build a client that infers without entitlement.
type IDTokenClaims struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  audience `json:"aud"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	Nonce     string   `json:"nonce"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
}

// verifyParams are the facts the caller must supply to check a token. Every
// one of them comes from state the caller created before the browser was
// opened — which is what makes them trustworthy.
type verifyParams struct {
	Issuer   string
	ClientID string
	Nonce    string
	Now      time.Time
}

// verifyIDToken parses, cryptographically verifies and validates an OIDC ID
// token, returning its claims.
//
// Order matters: the signature is checked BEFORE any claim is trusted. Reading
// claims from an unverified token and acting on them — even to decide which
// key to use — is how JWT libraries get written into CVEs.
func verifyIDToken(ctx context.Context, ks *keySet, raw string, p verifyParams) (*IDTokenClaims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("oauth: ID token is not a three-part JWS")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("oauth: ID token header is not valid base64url: %w", err)
	}
	var hdr idTokenHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		return nil, fmt.Errorf("oauth: ID token header is not valid JSON: %w", err)
	}

	// SECURITY: pin the algorithm. This single check closes two classic
	// attacks at once.
	//
	//   alg:"none" — the attacker strips the signature entirely and asserts
	//   whatever claims they like. A verifier that dispatches on the header's
	//   alg will cheerfully "verify" it.
	//
	//   alg:"HS256" — the attacker re-signs the token with HMAC, using the
	//   issuer's PUBLIC key as the shared secret. A verifier that dispatches
	//   on alg hands the public key to an HMAC check, which succeeds, because
	//   the public key is not secret.
	//
	// The defence is not to handle those algorithms carefully; it is to
	// refuse to let the attacker choose the algorithm at all. We know from
	// discovery that this issuer signs RS256, so anything else is rejected
	// before a key is even looked up.
	if hdr.Alg != algRS256 {
		return nil, fmt.Errorf("oauth: ID token is signed with %q, want %s", hdr.Alg, algRS256)
	}
	if hdr.Typ != "" && !strings.EqualFold(hdr.Typ, "JWT") {
		return nil, fmt.Errorf("oauth: ID token has unexpected typ %q", hdr.Typ)
	}

	pub, err := ks.key(ctx, hdr.Kid)
	if err != nil {
		return nil, err
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("oauth: ID token signature is not valid base64url: %w", err)
	}

	// The signed content is the exact bytes "header.payload" as they appeared
	// on the wire. Re-encoding the decoded parts would change them (base64
	// is not canonical) and break verification on valid tokens.
	signed := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signed))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sig); err != nil {
		return nil, fmt.Errorf("oauth: ID token signature verification failed: %w", err)
	}

	// Only now are the claims worth reading.
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oauth: ID token payload is not valid base64url: %w", err)
	}
	var claims IDTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("oauth: ID token payload is not valid JSON: %w", err)
	}

	// iss: the token must come from the issuer we performed discovery
	// against. Without this, a token minted by any issuer whose key we happen
	// to have cached would pass.
	if claims.Issuer != p.Issuer {
		return nil, fmt.Errorf("oauth: ID token issuer is %q, want %q", claims.Issuer, p.Issuer)
	}

	// aud: the token must have been minted FOR US. An ID token issued for a
	// different client is a valid token that says nothing about our sign-in;
	// accepting it is the token-substitution attack.
	if !hasScope(claims.Audience, p.ClientID) {
		return nil, fmt.Errorf("oauth: ID token audience %v does not include client %q", []string(claims.Audience), p.ClientID)
	}

	// exp: an expired token proves nothing about the present. Replaying a
	// year-old ID token must not sign anyone in.
	if claims.ExpiresAt == 0 {
		return nil, fmt.Errorf("oauth: ID token has no exp claim")
	}
	if !p.Now.Add(-clockSkew).Before(time.Unix(claims.ExpiresAt, 0)) {
		return nil, fmt.Errorf("oauth: ID token expired at %s", time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339))
	}

	// iat: a token issued in our future indicates a clock problem or a forged
	// token; either way it is not safe to treat as current.
	if claims.IssuedAt != 0 && time.Unix(claims.IssuedAt, 0).After(p.Now.Add(clockSkew)) {
		return nil, fmt.Errorf("oauth: ID token was issued in the future (iat %s)", time.Unix(claims.IssuedAt, 0).UTC().Format(time.RFC3339))
	}

	// nonce: this binds the token to THIS authorization attempt. The nonce was
	// generated locally, sent through the browser, and must come back inside
	// the signed token. It is what stops an attacker replaying a legitimately
	// signed ID token — captured from another session, or from their own
	// account — into our flow. Constant-time comparison avoids leaking the
	// expected value a byte at a time through response timing.
	if p.Nonce != "" {
		if subtle.ConstantTimeCompare([]byte(claims.Nonce), []byte(p.Nonce)) != 1 {
			return nil, fmt.Errorf("oauth: ID token nonce does not match the authorization request")
		}
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("oauth: ID token has no sub claim")
	}

	return &claims, nil
}
