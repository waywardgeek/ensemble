package oauth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// hostIDPrefix is the URN form the first-party docs use for
// ext_agent_host_id.
const hostIDPrefix = "urn:uuid:"

// DefaultHostIDPath is where this host's ext_agent_host_id lives.
//
// It is a separate file from the credentials, and that separation is
// deliberate: the host id must OUTLIVE the credentials. Signing out, or
// recovering from a terminal refresh error, clears tokens — and if the host
// id went with them, the next sign-in would present a brand-new host to the
// issuer and start a second registration for a host that already had one.
func DefaultHostIDPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("oauth: locating user config directory: %w", err)
	}
	return filepath.Join(dir, "ensemble", "agent_host_id"), nil
}

// newHostID mints a fresh RFC 4122 version-4 UUID in URN form.
//
// The value is random and carries no information. That is a requirement, not
// an implementation detail: the first-party docs are explicit that
// ext_agent_host_id must be opaque and must never be derived from an email
// address or a user id. Deriving it would turn an identifier that is sent in
// a browser URL — logged by proxies, visible in history, stored in the
// issuer's access logs — into a disclosure of who the user is. The docs also
// note it "is not an authentication credential", so it needs unlinkability
// rather than secrecy.
//
// It is per-host, not per-user: one client_id may span many hosts for the
// same user and workspace, and each host needs its own id so the issuer can
// tell a second laptop from a stolen token.
func newHostID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("oauth: generating host id: %w", err)
	}
	// Set the version (4) and variant (RFC 4122) bits so the value is a
	// well-formed UUID rather than 16 arbitrary bytes wearing a UUID's
	// punctuation.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%s%x-%x-%x-%x-%x", hostIDPrefix, b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// LoadOrCreateHostID returns this host's stable ext_agent_host_id, creating
// and persisting one if none exists.
//
// The write happens here, before the value is ever used, because the docs
// require the id to be "chosen/persisted BEFORE first sign-in". The failure
// this prevents is subtle: if the id were generated in memory, used to
// register, and only persisted afterwards, then a crash between the callback
// and the save would leave the issuer holding a registration for a host id
// that no longer exists anywhere on disk. The next run would invent a new one
// and register again, accumulating orphan registrations the user cannot see
// or revoke.
func LoadOrCreateHostID(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
		// An empty file is a failed earlier write; fall through and mint a
		// new id rather than returning "" and sending an empty required
		// parameter.
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("oauth: reading host id file: %w", err)
	}

	id, err := newHostID()
	if err != nil {
		return "", err
	}
	if err := writeHostID(path, id); err != nil {
		return "", err
	}
	return id, nil
}

// writeHostID persists the host id atomically with owner-only permissions.
//
// The id is not a secret, but it is a stable identifier for this machine, and
// the same atomic-rename discipline used for the credentials applies: a
// truncated host id file read on the next run would look like "no id yet" and
// trigger a duplicate registration.
func writeHostID(path, id string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("oauth: creating host id directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".host-id-*.tmp")
	if err != nil {
		return fmt.Errorf("oauth: creating temp host id file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if err := tmp.Chmod(0o600); err != nil {
		return fmt.Errorf("oauth: setting host id permissions: %w", err)
	}
	if _, err := tmp.WriteString(id + "\n"); err != nil {
		return fmt.Errorf("oauth: writing host id: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("oauth: syncing host id: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("oauth: closing temp host id file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("oauth: replacing host id file: %w", err)
	}
	return nil
}
