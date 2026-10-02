package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// Provider is a common.CredentialProvider backed by a ChatGPT OAuth grant.
//
// It is the piece that makes an expiring, rotating credential usable through
// an interface designed for a string. The engine asks for a bearer token; it
// may get one straight from memory, or it may get one that this provider
// minted half a millisecond ago by spending a refresh token and rewriting a
// file on disk. The caller cannot tell, and should not have to.
//
// Safe for concurrent use.
type Provider struct {
	cfg Config

	// mu serialises EVERYTHING, including the refresh network call.
	//
	// Holding a mutex across a network round trip is usually a smell. Here it
	// is the correctness requirement. Refresh tokens rotate: if two
	// goroutines notice an expiring token at the same moment and both call
	// the token endpoint, the first rotates the shared refresh token away and
	// the second presents a token that has already been redeemed. The issuer
	// treats that as evidence of theft — refresh_token_reused — and kills the
	// entire grant chain. So a concurrency bug here does not cause a retry;
	// it logs the user out and makes them sign in again.
	//
	// Serialising means the second goroutine waits, re-checks under the lock,
	// finds a freshly valid token and never calls the endpoint at all.
	mu    sync.Mutex
	creds *Credentials
	disco *Discovery
	keys  *keySet
}

// Compile-time proof that Provider satisfies the interface. If the interface
// changes, this breaks here rather than at some distant call site.
var _ common.CredentialProvider = (*Provider)(nil)

// NewProvider builds a Provider from cfg.
//
// It does NOT load credentials or touch the network: construction happens at
// startup, where a missing credential file is a normal first-run state rather
// than a fatal error, and where blocking on a token endpoint would make the
// agent slow to start for everyone. The credential is loaded lazily on first
// use.
func NewProvider(cfg Config) (*Provider, error) {
	norm, err := cfg.normalize()
	if err != nil {
		return nil, err
	}
	return &Provider{cfg: norm}, nil
}

// Kind reports that this credential is a ChatGPT OAuth grant.
//
// For display and policy only. Nothing in request construction may branch on
// it — if it did, the abstraction that lets a static key and a rotating token
// travel the same path would have failed.
func (p *Provider) Kind() common.CredentialKind { return common.CredentialChatGPTOAuth }

// GetBearerToken returns an access token that is valid at the moment it
// returns, refreshing first if necessary.
func (p *Provider) GetBearerToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.loadLocked(); err != nil {
		return "", err
	}

	// Refresh BEFORE the entitlement gate, deliberately.
	//
	// The granted scopes can change on a refresh — an entitlement can be
	// revoked server-side, and the token response is where we find out.
	// Gating on the stale stored scopes first would let a revoked plan keep
	// billing until the token happened to expire.
	if p.creds.NeedsRefresh(p.cfg.Now(), p.cfg.RefreshSkew) {
		if err := p.refreshLocked(ctx); err != nil {
			return "", err
		}
	}

	// THE GATE. This is the last thing between here and an inference request.
	//
	// First-party rule, verbatim: "Check the token response's granted scopes
	// for chatgpt.tokens.use.direct before proceeding to inference. A valid
	// ID token alone does not authorize ChatGPT plan usage."
	//
	// The scopes consulted are the ones the TOKEN RESPONSE reported, not the
	// ones we requested and not anything asserted by the ID token. Those
	// three are different facts, and only the first is the issuer telling us
	// what it actually granted. A client that checks its own request has
	// verified its own intentions.
	if !p.creds.HasDirectTokenScope() {
		return "", fmt.Errorf("%w (granted: %v)", ErrPlanScopeNotGranted, p.creds.GrantedScopes)
	}

	return p.creds.AccessToken, nil
}

// Credentials returns a snapshot of the stored credential, loading it if
// necessary. The returned value is a copy: callers cannot mutate provider
// state, and a rotation happening concurrently cannot change a struct the
// caller is in the middle of reading.
func (p *Provider) Credentials() (*Credentials, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.loadLocked(); err != nil {
		return nil, err
	}
	return p.creds.clone(), nil
}

// SignIn runs the interactive authorization flow and adopts the result.
func (p *Provider) SignIn(ctx context.Context) (*Credentials, error) {
	flow, err := NewFlow(p.cfg)
	if err != nil {
		return nil, err
	}
	creds, err := flow.Authorize(ctx)
	if err != nil {
		return nil, err
	}
	p.mu.Lock()
	p.creds = creds.clone()
	p.mu.Unlock()
	return creds, nil
}

// Refresh forces a token refresh regardless of expiry. Exposed for a
// "re-check my entitlement now" command; ordinary callers should just use
// GetBearerToken and let it decide.
func (p *Provider) Refresh(ctx context.Context) (*Credentials, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.loadLocked(); err != nil {
		return nil, err
	}
	if err := p.refreshLocked(ctx); err != nil {
		return nil, err
	}
	return p.creds.clone(), nil
}

// loadLocked populates p.creds from the store on first use.
func (p *Provider) loadLocked() error {
	if p.creds != nil {
		return nil
	}
	creds, err := p.cfg.Store.Load()
	if err != nil {
		return err
	}
	p.creds = creds
	return nil
}

// ensureDiscoveryLocked fetches and caches the discovery document and key set.
//
// Cached for the life of the provider because the endpoints do not move, and
// a discovery fetch on every refresh would double the latency of the one
// operation that sits in the hot path of a user's turn.
func (p *Provider) ensureDiscoveryLocked(ctx context.Context) error {
	if p.disco != nil {
		return nil
	}
	disco, err := FetchDiscovery(ctx, p.cfg.HTTPClient, p.cfg.Issuer)
	if err != nil {
		return err
	}
	p.disco = disco
	p.keys = newKeySet(p.cfg.HTTPClient, disco.JWKSURI, p.cfg.Now)
	return nil
}

// refreshLocked performs the refresh exchange and persists the result.
//
// Must be called with p.mu held.
func (p *Provider) refreshLocked(ctx context.Context) error {
	if p.creds.RefreshToken == "" {
		// offline_access was not granted, or the token has already been
		// cleared by an earlier terminal failure. Either way there is no path
		// back except a browser.
		return fmt.Errorf("%w: no refresh token is stored", ErrReauthRequired)
	}
	if err := p.ensureDiscoveryLocked(ctx); err != nil {
		return err
	}

	tr, err := refreshGrant(ctx, p.cfg.HTTPClient, p.disco.TokenEndpoint, p.creds.ClientID, p.creds.RefreshToken)
	if err != nil {
		var te *TokenError
		if errors.As(err, &te) && te.Terminal() {
			// Terminal. The grant is dead; keeping the tokens around would
			// only produce more failed refreshes.
			//
			// Note what is NOT cleared: the issued client id, the host id and
			// the ID token. The docs require re-authorization to happen "WITH
			// THE SAVED ISSUED CLIENT ID" — discarding it would start a
			// second registration and orphan the one the user already
			// approved, leaving an entry in their account they cannot match
			// to anything.
			if clearErr := p.clearTokensLocked(); clearErr != nil {
				return fmt.Errorf("%w: %w (and clearing the stored tokens failed: %w)", ErrReauthRequired, te, clearErr)
			}
			// Wrapping both sentinels keeps errors.Is(err, ErrReauthRequired)
			// true for the recovery path while errors.As(&TokenError) still
			// recovers the specific code for diagnostics.
			return fmt.Errorf("%w: %w", ErrReauthRequired, te)
		}
		// Transient: a 503, a network blip, an unparseable body. The refresh
		// token is left intact precisely because it may still be good, and
		// destroying it here would convert a momentary outage into a forced
		// re-login.
		return err
	}

	now := p.cfg.Now()

	// Validate the new ID token if one came back. A refresh that returns an
	// ID token is re-asserting who the user is, and an unverified assertion
	// is worth nothing.
	//
	// There is no nonce to check here: the nonce binds an ID token to an
	// interactive authorization request, and this exchange is not one. The
	// binding that matters for a refresh is the refresh token itself, which
	// only we hold. Passing an empty nonce says that explicitly rather than
	// leaving it to chance.
	if tr.IDToken != "" {
		claims, err := verifyIDToken(ctx, p.keys, tr.IDToken, verifyParams{
			Issuer:   p.disco.Issuer,
			ClientID: p.creds.ClientID,
			Now:      now,
		})
		if err != nil {
			return err
		}
		// A refresh must not quietly change which account we are. If the
		// subject moved, something is badly wrong and continuing would spend
		// the wrong person's plan.
		if p.creds.Subject != "" && claims.Subject != p.creds.Subject {
			return fmt.Errorf("oauth: refreshed ID token is for a different subject; refusing to switch accounts")
		}
	}

	// Build the complete replacement first, then swap it in and write it in
	// one go. The four rotating fields — access token, expiry, granted scopes
	// and refresh token — move together or not at all.
	updated := p.creds.clone()
	updated.AccessToken = tr.AccessToken
	updated.ExpiresAt = expiryFrom(tr, now)

	// A rotating server returns a NEW refresh token. Some return none on some
	// refreshes, which per RFC 6749 §6 means "keep using the one you have" —
	// so an empty field must not be allowed to blank out the only credential
	// capable of recovering this session.
	if tr.RefreshToken != "" {
		updated.RefreshToken = tr.RefreshToken
	}

	// Likewise, an omitted scope means "identical to the previously granted
	// scope" (RFC 6749 §5.1). Treating absent as empty would revoke the
	// user's entitlement on our side and block inference against a perfectly
	// good plan. A scope that IS present replaces the old one wholesale,
	// which is how a server-side revocation reaches the gate.
	if tr.Scope != "" {
		updated.GrantedScopes = splitScopes(tr.Scope)
	}

	if tr.IDToken != "" {
		updated.IDToken = tr.IDToken
	}

	// In-memory first, then disk. The rotation has ALREADY happened on the
	// server by this point: the old refresh token is dead whatever we do, so
	// the worst outcome is to forget the new one. Adopting it in memory means
	// this process stays alive even if the write fails.
	p.creds = updated
	if err := p.cfg.Store.Save(updated); err != nil {
		// Still an error, and a loud one. The new refresh token exists only
		// in this process's memory; when it exits, the user is locked out and
		// must sign in again. They need to hear about that now, while the
		// cause — a full disk, a bad permission — is still fixable.
		return fmt.Errorf("oauth: refreshed successfully but could not persist the new tokens: %w", err)
	}
	return nil
}

// clearTokensLocked wipes the rotating secrets while preserving the identity
// needed to re-authorize.
func (p *Provider) clearTokensLocked() error {
	cleared := p.creds.clone()
	cleared.AccessToken = ""
	cleared.RefreshToken = ""
	cleared.ExpiresAt = time.Time{}
	cleared.GrantedScopes = nil
	p.creds = cleared
	return p.cfg.Store.Save(cleared)
}

// HTTPClientWithTimeout is a small helper for callers assembling a Config.
func HTTPClientWithTimeout(d time.Duration) *http.Client { return &http.Client{Timeout: d} }
