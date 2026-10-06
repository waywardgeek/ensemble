package oauth

import (
	"context"
	"crypto/rsa"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestIDTokenValidation drives every rejection through the real Authorize
// flow. Testing verifyIDToken in isolation would prove the function works;
// testing it here proves it is actually WIRED IN — that a forged token fails
// the sign-in rather than merely failing a helper nobody calls.
func TestIDTokenValidation(t *testing.T) {
	_, otherKey := keys(t)

	tests := []struct {
		name         string
		issuerTweak  func(f *fakeIssuer)
		browserTweak func(b *browserSim)
		wantErrText  string
		wantOK       bool
	}{
		{
			// The control. Without it, a table in which every row expects an
			// error would still pass if Authorize were broken for an
			// unrelated reason.
			name:   "a correctly signed token is accepted",
			wantOK: true,
		},
		{
			name: "a token signed by the wrong key is rejected",
			issuerTweak: func(f *fakeIssuer) {
				// Structurally perfect, correct claims, real RS256
				// signature — but by a key that is not in the JWKS.
				f.idTokenSigner = otherKey
			},
			wantErrText: "signature verification failed",
		},
		{
			name: "alg:none is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenHeader = func(m map[string]string) { m["alg"] = "none" }
			},
			wantErrText: `signed with "none"`,
		},
		{
			name: "an HS256 header is rejected without consulting a key",
			issuerTweak: func(f *fakeIssuer) {
				// The algorithm-confusion attack: claim HMAC so a naive
				// verifier uses the issuer's PUBLIC key as a shared secret.
				f.idTokenHeader = func(m map[string]string) { m["alg"] = "HS256" }
			},
			wantErrText: `signed with "HS256"`,
		},
		{
			name: "an unknown signing key id is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenHeader = func(m map[string]string) { m["kid"] = "no-such-key" }
			},
			wantErrText: "no signing key",
		},
		{
			name: "a token from a different issuer is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) { m["iss"] = "https://auth.evil.example" }
			},
			wantErrText: "issuer",
		},
		{
			name: "a token minted for a different client is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) { m["aud"] = "some_other_client" }
			},
			wantErrText: "audience",
		},
		{
			name: "an expired token is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) {
					m["exp"] = time.Now().Add(-2 * time.Hour).Unix()
				}
			},
			wantErrText: "expired",
		},
		{
			name: "a token with no exp is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) { delete(m, "exp") }
			},
			wantErrText: "no exp claim",
		},
		{
			name: "a token issued in the future is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) {
					m["iat"] = time.Now().Add(2 * time.Hour).Unix()
				}
			},
			wantErrText: "issued in the future",
		},
		{
			name: "a nonce mismatch is rejected",
			browserTweak: func(b *browserSim) {
				// The issuer mints the token with the nonce the code was
				// bound to, so overriding it here produces a validly signed
				// token carrying the wrong nonce — exactly what a replayed
				// token from another session looks like.
				b.nonceOverride = "a-nonce-from-another-session"
			},
			wantErrText: "nonce",
		},
		{
			name: "a token with no subject is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.idTokenClaims = func(m map[string]any) { delete(m, "sub") }
			},
			wantErrText: "no sub claim",
		},
		{
			name: "a token response with no id_token at all is rejected",
			issuerTweak: func(f *fakeIssuer) {
				f.omitIDToken = true
			},
			wantErrText: "no id_token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			if tc.issuerTweak != nil {
				tc.issuerTweak(env.issuer)
			}
			if tc.browserTweak != nil {
				tc.browserTweak(env.browser)
			}

			creds, err := env.authorize(t)

			if tc.wantOK {
				if err != nil {
					t.Fatalf("Authorize() = %v, want success", err)
				}
				if creds.Subject != testSubject {
					t.Errorf("subject = %q, want %q", creds.Subject, testSubject)
				}
				return
			}

			if err == nil {
				t.Fatalf("Authorize() succeeded; want rejection mentioning %q", tc.wantErrText)
			}
			if !strings.Contains(err.Error(), tc.wantErrText) {
				t.Fatalf("error = %v, want it to mention %q", err, tc.wantErrText)
			}
			// A rejected ID token must leave nothing behind.
			if _, loadErr := env.cfg.Store.Load(); !errors.Is(loadErr, ErrNoCredentials) {
				t.Errorf("a rejected ID token still persisted credentials (load err = %v)", loadErr)
			}
		})
	}
}

// TestAudienceAcceptsBothForms: OIDC allows aud to be a string or an array,
// and a verifier that handles only one shape rejects conforming tokens.
func TestAudienceAcceptsBothForms(t *testing.T) {
	tests := []struct {
		name string
		aud  any
		ok   bool
	}{
		{name: "string form", aud: testIssuedClient, ok: true},
		{name: "array containing us", aud: []string{"other", testIssuedClient}, ok: true},
		{name: "array without us", aud: []string{"other", "another"}, ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			env.issuer.idTokenClaims = func(m map[string]any) { m["aud"] = tc.aud }

			_, err := env.authorize(t)
			if tc.ok && err != nil {
				t.Fatalf("Authorize() = %v, want success", err)
			}
			if !tc.ok && err == nil {
				t.Fatal("Authorize() accepted a token minted for someone else")
			}
		})
	}
}

// TestDegenerateJWKSKeyRejected: a public exponent of 1 makes RSA
// verification a no-op, so a hostile JWKS could switch signature checking off
// by publishing one.
func TestDegenerateJWKSKeyRejected(t *testing.T) {
	for _, e := range []string{"AQ" /* 1 */, "AA" /* 0 */} {
		jwk := jsonWebKey{Kty: "RSA", Kid: "degenerate", N: "AQAB", E: e}
		if _, err := jwk.rsaPublicKey(); err == nil {
			t.Errorf("rsaPublicKey() accepted a degenerate exponent %q", e)
		}
	}
}

// TestMalformedIDTokens checks the parser fails safely on junk rather than
// panicking, which matters because this input is attacker-controlled.
func TestMalformedIDTokens(t *testing.T) {
	ks := &keySet{keys: map[string]*rsa.PublicKey{}, now: time.Now}
	inputs := []string{
		"",
		"not-a-jwt",
		"only.two",
		"a.b.c.d",
		"!!!.???.***",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			_, err := verifyIDToken(context.Background(), ks, in, verifyParams{
				Issuer: "https://issuer.example", ClientID: "c", Now: time.Now(),
			})
			if err == nil {
				t.Fatalf("verifyIDToken(%q) succeeded; want an error", in)
			}
		})
	}
}
