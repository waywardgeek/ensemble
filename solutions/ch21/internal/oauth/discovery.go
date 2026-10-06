package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DefaultIssuer is the production Sign in with ChatGPT issuer.
const DefaultIssuer = "https://auth.openai.com"

// IssuerEnvVar names the environment variable that overrides the issuer.
//
// This is a real feature, not a test hook. Enterprise deployments front the
// consumer issuer with their own tenant host, and an agent that hardcodes
// auth.openai.com simply cannot sign in there. That it also makes the flow
// testable against an httptest server is a consequence of the seam being in
// the right place, not its purpose.
const IssuerEnvVar = "OPENAI_OIDC_ISSUER"

// Issuer returns the OIDC issuer to use: the OPENAI_OIDC_ISSUER override if
// set and non-empty, otherwise the production default.
//
// The trailing slash is trimmed because the issuer is compared verbatim
// against the "iss" claim of the ID token and against the discovery
// document's own issuer field. "https://auth.openai.com/" and
// "https://auth.openai.com" are the same server but different strings, and an
// equality check that disagrees with the user's typing is a confusing failure.
func Issuer() string {
	if v := strings.TrimSpace(os.Getenv(IssuerEnvVar)); v != "" {
		return strings.TrimSuffix(v, "/")
	}
	return DefaultIssuer
}

// Discovery is the subset of the OIDC discovery document this package uses.
//
// Only the fields we act on are modelled. Unmarshalling into a narrow struct
// rather than a map keeps the compiler honest about what the flow depends on:
// if a field is not here, no code path can quietly start relying on it.
type Discovery struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	RevocationEndpoint               string   `json:"revocation_endpoint"`
	JWKSURI                          string   `json:"jwks_uri"`
	CodeChallengeMethodsSupported    []string `json:"code_challenge_methods_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// FetchDiscovery retrieves and validates {issuer}/.well-known/openid-configuration.
//
// The validation is the point. A discovery document tells us which URL to send
// the user's browser to and which URL to post their authorization code to; if
// we fetch it from a host we were pointed at by configuration and then trust
// whatever it says about itself, a bad OPENAI_OIDC_ISSUER value becomes a
// credential-exfiltration channel. So:
//
//   - The document's own "issuer" must equal the issuer we asked for. This is
//     required by OIDC Discovery §4.3 and it is what stops a document served
//     at host A from claiming to speak for host B — which matters because the
//     "iss" claim of every ID token is checked against this same value.
//   - The endpoints must be present and absolute.
func FetchDiscovery(ctx context.Context, client *http.Client, issuer string) (*Discovery, error) {
	issuer = strings.TrimSuffix(issuer, "/")
	endpoint := issuer + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("oauth: building discovery request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth: fetching discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth: discovery document at %s returned HTTP %d", endpoint, resp.StatusCode)
	}

	// Bounded read: the discovery document is a few kilobytes, and an
	// unbounded io.ReadAll against a hostile or broken server is an
	// out-of-memory waiting to happen.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth: reading discovery document: %w", err)
	}

	var d Discovery
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("oauth: parsing discovery document: %w", err)
	}

	if d.Issuer != issuer {
		return nil, fmt.Errorf("oauth: discovery document at %s claims issuer %q, want %q", endpoint, d.Issuer, issuer)
	}
	for name, raw := range map[string]string{
		"authorization_endpoint": d.AuthorizationEndpoint,
		"token_endpoint":         d.TokenEndpoint,
		"jwks_uri":               d.JWKSURI,
	} {
		if raw == "" {
			return nil, fmt.Errorf("oauth: discovery document is missing %s", name)
		}
		u, err := url.Parse(raw)
		if err != nil || !u.IsAbs() {
			return nil, fmt.Errorf("oauth: discovery document has a non-absolute %s: %q", name, raw)
		}
	}

	// Refuse to proceed if the server does not support S256. We would rather
	// fail loudly here than fall back to "plain", which is PKCE with the
	// security removed. An empty list is tolerated because the field is
	// optional in the spec and some tenant fronts omit it; a list that is
	// present and lacks S256 is a real incompatibility.
	if len(d.CodeChallengeMethodsSupported) > 0 && !hasScope(d.CodeChallengeMethodsSupported, codeChallengeMethodS256) {
		return nil, fmt.Errorf("oauth: issuer %s does not support the S256 code challenge method (advertises %v)", issuer, d.CodeChallengeMethodsSupported)
	}

	return &d, nil
}
