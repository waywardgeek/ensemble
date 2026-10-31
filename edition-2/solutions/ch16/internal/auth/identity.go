package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type identityClaims struct {
	// Issuer must match the explicitly trusted discovery authority.
	Issuer string `json:"iss"`
	// Audience can be one client ID or an OIDC array of client IDs.
	Audience json.RawMessage `json:"aud"`
	// AuthorizedParty disambiguates an ID token with several audiences.
	AuthorizedParty string `json:"azp"`
	// Subject is stable identity; email cannot substitute for it.
	Subject string `json:"sub"`
	// Email supplies optional human-readable display metadata.
	Email string `json:"email"`
	// ExpiresAt rejects old or replayed identity assertions.
	ExpiresAt int64 `json:"exp"`
	// Nonce binds the ID token to this specific login attempt.
	Nonce string `json:"nonce"`
}

// Only RS256 is accepted, matching the issuer's signed identity-token contract.
// JWT-provided URLs and algorithms never decide where keys come from or how a
// signature is checked. Claims become identity only after signature validation.
func (s *store) identity(ctx context.Context, config discovery, token, clientID, nonce string) (identityClaims, error) {
	fail := errors.New("OAuth ID token validation failed")
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return identityClaims{}, fail
	}
	decode := base64.RawURLEncoding.DecodeString
	header, err := decode(parts[0])
	if err != nil {
		return identityClaims{}, fail
	}
	var h struct {
		// Algorithm must match the fixed signature algorithm, never JWT-directed fallback.
		Algorithm string `json:"alg"`
		// KeyID selects a published RSA verification key.
		KeyID string `json:"kid"`
	}
	if json.Unmarshal(header, &h) != nil || h.Algorithm != "RS256" || h.KeyID == "" {
		return identityClaims{}, fail
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.JWKS, nil)
	if err != nil {
		return identityClaims{}, fail
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return identityClaims{}, errors.New("OAuth verification keys unavailable")
	}
	defer resp.Body.Close()
	var keys struct {
		// Keys comes only from the trusted issuer's discovery metadata.
		Keys []struct {
			// ID identifies the signing key without trusting a URL from the JWT.
			ID string `json:"kid"`
			// Type restricts the material to RSA public keys.
			Type string `json:"kty"`
			// Algorithm, when supplied, must agree with RS256.
			Algorithm string `json:"alg"`
			// Use, when supplied, must permit signature verification.
			Use string `json:"use"`
			// N encodes the unsigned RSA modulus.
			N string `json:"n"`
			// E encodes the unsigned public exponent.
			E string `json:"e"`
		} `json:"keys"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(resp.Body).Decode(&keys) != nil {
		return identityClaims{}, fail
	}
	signature, err := decode(parts[2])
	if err != nil {
		return identityClaims{}, fail
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	valid := false
	for _, key := range keys.Keys {
		if key.ID != h.KeyID || key.Type != "RSA" || (key.Algorithm != "" && key.Algorithm != "RS256") || (key.Use != "" && key.Use != "sig") {
			continue
		}
		n, errN := decode(key.N)
		exponent, errE := decode(key.E)
		if errN != nil || errE != nil || len(exponent) == 0 || len(exponent) > 4 {
			continue
		}
		e := 0
		for _, b := range exponent {
			e = e*256 + int(b)
		}
		if e < 3 {
			continue
		}
		if rsa.VerifyPKCS1v15(&rsa.PublicKey{N: new(big.Int).SetBytes(n), E: e}, crypto.SHA256, digest[:], signature) == nil {
			valid = true
			break
		}
	}
	if !valid {
		return identityClaims{}, fail
	}
	raw, err := decode(parts[1])
	if err != nil {
		return identityClaims{}, fail
	}
	var claims identityClaims
	if json.Unmarshal(raw, &claims) != nil || claims.Issuer != config.Issuer || claims.Subject == "" || claims.ExpiresAt <= time.Now().Unix() || (nonce != "" && claims.Nonce != nonce) {
		return identityClaims{}, fail
	}
	var audiences []string
	var audience string
	if json.Unmarshal(claims.Audience, &audience) == nil {
		audiences = []string{audience}
	} else if json.Unmarshal(claims.Audience, &audiences) != nil {
		return identityClaims{}, fail
	}
	matched := false
	for _, aud := range audiences {
		matched = matched || aud == clientID
	}
	if !matched || (len(audiences) > 1 && claims.AuthorizedParty != clientID) || (claims.AuthorizedParty != "" && claims.AuthorizedParty != clientID) {
		return identityClaims{}, fail
	}
	return claims, nil
}
