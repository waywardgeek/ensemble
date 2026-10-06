package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// jsonWebKey is one RSA public key from a JWKS document.
//
// Only RSA is modelled because the issuer advertises
// id_token_signing_alg_values_supported: ["RS256"]. A key of another type is
// skipped rather than guessed at.
type jsonWebKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"` // modulus, base64url big-endian
	E   string `json:"e"` // public exponent, base64url big-endian
}

// rsaPublicKey converts the JWK's base64url big-endian integers into a usable
// key.
func (k jsonWebKey) rsaPublicKey() (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("oauth: JWK %q has key type %q, want RSA", k.Kid, k.Kty)
	}
	// JWK integers are base64url WITHOUT padding (RFC 7517 §3).
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("oauth: JWK %q has an undecodable modulus: %w", k.Kid, err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("oauth: JWK %q has an undecodable exponent: %w", k.Kid, err)
	}
	if len(nBytes) == 0 || len(eBytes) == 0 {
		return nil, fmt.Errorf("oauth: JWK %q has an empty modulus or exponent", k.Kid)
	}
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() || e.Int64() < 3 {
		// A public exponent of 0 or 1 makes the signature check a no-op:
		// m^1 mod n == m, so every "signature" verifies. Refusing it here
		// means a hostile JWKS cannot turn verification off by shipping a
		// degenerate key.
		return nil, fmt.Errorf("oauth: JWK %q has an unusable public exponent", k.Kid)
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(e.Int64())}, nil
}

// keySet caches the issuer's signing keys and refetches them when a token
// arrives signed by a key id we have not seen.
//
// The cache matters because verification happens on every sign-in and every
// refresh that returns a new ID token, and a network round trip per
// verification would make the issuer's availability a hard dependency of
// every token operation. The refetch-on-unknown-kid matters because issuers
// rotate signing keys without notice; a client that caches forever starts
// rejecting perfectly good tokens the day the key rolls.
//
// The minimum interval between refetches stops an attacker from using a
// stream of tokens bearing random kids to make us hammer the JWKS endpoint.
type keySet struct {
	client  *http.Client
	uri     string
	minWait time.Duration
	now     func() time.Time

	mu          sync.Mutex
	keys        map[string]*rsa.PublicKey
	lastFetched time.Time
}

func newKeySet(client *http.Client, uri string, now func() time.Time) *keySet {
	return &keySet{
		client:  client,
		uri:     uri,
		minWait: time.Minute,
		now:     now,
		keys:    make(map[string]*rsa.PublicKey),
	}
}

// key returns the public key for kid, fetching the JWKS if necessary.
func (ks *keySet) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	ks.mu.Lock()
	if k, ok := ks.keys[kid]; ok {
		ks.mu.Unlock()
		return k, nil
	}
	fetchedAt := ks.lastFetched
	ks.mu.Unlock()

	if !fetchedAt.IsZero() && ks.now().Sub(fetchedAt) < ks.minWait {
		return nil, fmt.Errorf("oauth: no signing key with kid %q (JWKS refetch rate-limited)", kid)
	}
	if err := ks.refresh(ctx); err != nil {
		return nil, err
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	k, ok := ks.keys[kid]
	if !ok {
		return nil, fmt.Errorf("oauth: issuer published no signing key with kid %q", kid)
	}
	return k, nil
}

// refresh fetches the JWKS document and replaces the cache.
func (ks *keySet) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ks.uri, nil)
	if err != nil {
		return fmt.Errorf("oauth: building JWKS request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := ks.client.Do(req)
	if err != nil {
		return fmt.Errorf("oauth: fetching JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oauth: JWKS endpoint returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("oauth: reading JWKS: %w", err)
	}

	var doc struct {
		Keys []jsonWebKey `json:"keys"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("oauth: parsing JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, jwk := range doc.Keys {
		// Skip keys that are explicitly for encryption, and keys that declare
		// an algorithm other than RS256. A signing key is the only thing we
		// are looking for here.
		if jwk.Use != "" && jwk.Use != "sig" {
			continue
		}
		if jwk.Alg != "" && jwk.Alg != algRS256 {
			continue
		}
		pub, err := jwk.rsaPublicKey()
		if err != nil {
			// One malformed key must not poison the whole set — the issuer
			// may be mid-rotation and publishing a type we do not handle.
			continue
		}
		keys[jwk.Kid] = pub
	}
	if len(keys) == 0 {
		return fmt.Errorf("oauth: JWKS at %s contained no usable RS256 signing keys", ks.uri)
	}

	ks.mu.Lock()
	ks.keys = keys
	ks.lastFetched = ks.now()
	ks.mu.Unlock()
	return nil
}
