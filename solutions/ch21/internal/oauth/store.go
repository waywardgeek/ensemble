package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Credentials is everything this host needs to keep using a ChatGPT plan
// without sending the user back through a browser.
//
// The four fields that rotate together — AccessToken, ExpiresAt,
// GrantedScopes and RefreshToken — are deliberately in one struct written by
// one atomic Save. A refresh replaces all four at once, and persisting them
// separately would create windows in which the file describes a state that
// never existed: a new access token with an old refresh token, say, or new
// tokens with stale scopes. Recovering from that is worse than it sounds,
// because the stale refresh token has already been invalidated by the
// rotation and the user is locked out with no error to explain why.
type Credentials struct {
	// Issuer is the OIDC issuer these tokens came from. It is stored so that
	// changing OPENAI_OIDC_ISSUER does not silently present a tenant's token
	// to a different tenant.
	Issuer string `json:"issuer"`

	// ClientID is the ISSUED client id, handed to us on the callback of a
	// successful new registration. It is never dynamic_agent_client: that is
	// the bootstrap entrypoint, not an identity, and persisting it would mean
	// re-registering on every sign-in and orphaning the previous registration.
	ClientID string `json:"client_id"`

	// HostID is the ext_agent_host_id presented at authorization. Stored
	// alongside so a re-auth from this host is recognisably the same host.
	HostID string `json:"ext_agent_host_id"`

	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`

	// IDToken is retained, not discarded after validation, because it is the
	// id_token_hint for a later re-authorization: it lets the issuer skip the
	// account chooser and re-confirm the same account.
	IDToken string `json:"id_token,omitempty"`

	// GrantedScopes is what the TOKEN RESPONSE said we were given, which is
	// not necessarily what we asked for. Storing the requested scopes here
	// instead would turn the entitlement gate into a check that we intended
	// to be entitled.
	GrantedScopes []string `json:"granted_scopes"`

	ExpiresAt time.Time `json:"expires_at"`

	// Subject and Email identify the signed-in account for display. Subject
	// is the stable identifier; Email can change.
	Subject string `json:"subject,omitempty"`
	Email   string `json:"email,omitempty"`
}

// HasDirectTokenScope reports whether the issuer granted the entitlement that
// authorizes inference against the user's ChatGPT plan.
//
// First-party rule, verbatim: "Check the token response's granted scopes for
// chatgpt.tokens.use.direct before proceeding to inference. A valid ID token
// alone does not authorize ChatGPT plan usage."
func (c *Credentials) HasDirectTokenScope() bool {
	return c != nil && hasScope(c.GrantedScopes, ScopeDirectTokens)
}

// NeedsRefresh reports whether the access token should be replaced before use.
//
// The skew is why this is not simply "expired". An access token that is valid
// for another two seconds will very likely be invalid by the time a streaming
// inference request finishes its handshake, so refreshing only on observed
// expiry guarantees a crop of spurious 401s under load. Refreshing early
// costs one cheap round trip; refreshing late costs a failed turn.
func (c *Credentials) NeedsRefresh(now time.Time, skew time.Duration) bool {
	if c == nil || c.AccessToken == "" {
		return true
	}
	if c.ExpiresAt.IsZero() {
		// No expiry information: treat as needing refresh rather than
		// assuming immortality.
		return true
	}
	return !now.Add(skew).Before(c.ExpiresAt)
}

// String renders the credential with every secret redacted.
//
// This exists so that the obvious mistake — dropping a Credentials into a log
// line, an error, or a debug print during a 2am incident — cannot leak a
// bearer token. Go reaches for String() automatically in %v and %s, so the
// safe rendering is the default one and leaking requires deliberately
// reaching past it.
func (c *Credentials) String() string {
	if c == nil {
		return "Credentials(nil)"
	}
	return fmt.Sprintf("Credentials{client_id:%s subject:%s scopes:%v expires_at:%s access_token:REDACTED refresh_token:REDACTED id_token:REDACTED}",
		c.ClientID, c.Subject, c.GrantedScopes, c.ExpiresAt.UTC().Format(time.RFC3339))
}

// clone returns a deep copy.
//
// Credentials are handed out to callers and mutated by refresh under a lock.
// Returning the live pointer would let a caller read a half-updated struct, or
// retain a token slice that the next rotation overwrites underneath them.
func (c *Credentials) clone() *Credentials {
	if c == nil {
		return nil
	}
	cp := *c
	cp.GrantedScopes = append([]string(nil), c.GrantedScopes...)
	return &cp
}

// Store persists Credentials to a single JSON file.
type Store struct {
	path string
}

// NewStore returns a Store backed by the file at path.
func NewStore(path string) *Store { return &Store{path: path} }

// Path reports the file this Store reads and writes.
func (s *Store) Path() string { return s.path }

// DefaultCredentialPath is the app-defined location for the credential file,
// following the first-party guidance to use an app-owned config directory.
func DefaultCredentialPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("oauth: locating user config directory: %w", err)
	}
	return filepath.Join(dir, "ensemble", "chatgpt_credentials.json"), nil
}

// Load reads the stored credentials.
//
// It returns ErrNoCredentials when the file does not exist, so that "never
// signed in" is distinguishable from "signed in but the file is corrupt" —
// the first is a normal first-run state and the second is a real fault that
// should not be silently papered over by starting a new sign-in.
func (s *Store) Load() (*Credentials, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNoCredentials
		}
		return nil, fmt.Errorf("oauth: opening credential file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("oauth: stat credential file: %w", err)
	}
	// Refuse to read a credential file that others can read. A long-lived
	// refresh token in a world-readable file is a credential leak that no
	// amount of care elsewhere in this package can compensate for, and
	// failing loudly is the only way the user ever finds out.
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("oauth: credential file %s has permissions %#o; want 0600 (owner-only)", s.path, info.Mode().Perm())
	}

	var c Credentials
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, fmt.Errorf("oauth: parsing credential file %s: %w", s.path, err)
	}
	return &c, nil
}

// Save writes the credentials atomically with owner-only permissions.
//
// Atomic means: write a temp file in the SAME directory, fsync it, then
// rename over the target. Rename within a directory is atomic on POSIX, so a
// reader at any instant sees either the whole old file or the whole new one —
// never a truncated one. The temp file must share the directory because
// rename across filesystems is not atomic and will fail outright.
//
// This is not theoretical tidiness. The moment of maximum danger is exactly
// the refresh: the old refresh token has already been spent and invalidated
// by the server, so a crash that leaves a truncated file destroys the only
// copy of the replacement and locks the user out.
func (s *Store) Save(c *Credentials) error {
	if c == nil {
		return errors.New("oauth: refusing to save nil credentials")
	}
	dir := filepath.Dir(s.path)
	// 0700: the directory itself should not be listable by others, or the
	// existence and name of the credential file leaks even though its
	// contents do not.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("oauth: creating credential directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("oauth: encoding credentials: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".credentials-*.tmp")
	if err != nil {
		return fmt.Errorf("oauth: creating temp credential file: %w", err)
	}
	tmpName := tmp.Name()
	// Clean up the temp file on any failure path. After a successful Rename
	// the name no longer exists and the Remove is a harmless no-op.
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	// CreateTemp already makes the file 0600, but set it explicitly: the
	// guarantee is important enough to state in code rather than inherit, and
	// a umask-sensitive future change to how the file is created would
	// otherwise silently widen it.
	if err := tmp.Chmod(0o600); err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("oauth: setting credential file permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("oauth: writing credentials: %w", err)
	}
	// fsync before rename. Without it the rename can land while the data
	// blocks are still in the page cache, and a power loss leaves a correctly
	// named, empty credential file.
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("oauth: syncing credentials: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("oauth: closing temp credential file: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		return fmt.Errorf("oauth: replacing credential file: %w", err)
	}
	return nil
}

// Clear removes the stored credentials. A missing file is success, because
// the caller's intent — "there must be no credential here" — is satisfied.
func (s *Store) Clear() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("oauth: removing credential file: %w", err)
	}
	return nil
}
