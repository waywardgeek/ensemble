// Package auth owns protected registration state, OAuth and token renewal.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"syscall"
	"time"

	"ensemble/internal/common"
)

// APIKey is an immutable credential value, not a service with its own lifetime.
type APIKey string

// GetBearerToken supplies the configured key without interpreting its contents.
func (k APIKey) GetBearerToken(context.Context) (string, error) {
	return string(k), nil
}

// Kind names the metered funding source without revealing the key.
func (APIKey) Kind() common.CredentialKind { return common.CredentialAPIKey }

// Each registration stays separate by issued client ID. The UX currently exposes
// one selected account, but email collisions never merge credential records.
type profile struct {
	// ClientID is the issued registration, never the bootstrap identifier.
	ClientID string `json:"client_id"`
	// Subject and Email come only from a verified ID token.
	Subject string `json:"subject"`
	// Email is display-only and does not key the profile.
	Email string `json:"email"`
	// AccessToken is sent only in the inference authorization header.
	AccessToken string `json:"access_token,omitempty"`
	// RefreshToken rotates and must be persisted with the access token.
	RefreshToken string `json:"refresh_token,omitempty"`
	// IDToken stays protected even after its display claims have been extracted.
	IDToken string `json:"id_token,omitempty"`
	// ExpiresAt controls proactive renewal rather than assuming a token lifetime.
	ExpiresAt time.Time `json:"expires_at"`
	// Scopes are actual grants from the token endpoint.
	Scopes []string `json:"scopes,omitempty"`
}
type saved struct {
	// HostID survives all registrations and sign-outs.
	HostID string `json:"host_id"`
	// Issuer prevents reusing a fake or foreign issuer's tokens on another server.
	Issuer string `json:"issuer"`
	// Active selects the account without conflating accounts sharing an email.
	Active string `json:"active"`
	// Pending retains an issued ID before exchange without replacing the active account.
	Pending string `json:"pending,omitempty"`
	// Billing records explicit user intent; credential expiry never selects API billing.
	Billing common.CredentialKind `json:"billing"`
	// Profiles groups protected account registrations by their issued ID.
	Profiles map[string]profile `json:"profiles"`
}
type store struct {
	parent  common.Ensemble
	options common.AuthOptions
	client  *http.Client
	mu      sync.Mutex
}

// New constructs an Ensemble-owned service without reading anyone else's store.
func New(parent common.Ensemble, options common.AuthOptions) common.AuthStore {
	if parent == nil {
		panic("AuthStore requires Ensemble")
	}
	if options.Issuer == "" {
		options.Issuer = "https://auth.openai.com"
	}
	return &store{parent: parent, options: options, client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

// Ensemble exposes the service's required owner.
func (s *store) Ensemble() common.Ensemble { return s.parent }

// Kind distinguishes subscription credentials from metered API keys.
func (s *store) Kind() common.CredentialKind { return common.CredentialChatGPT }

// A mutex serializes goroutines; the advisory file lock also serializes this
// app's CLI and GUI processes. Always reload under that lock: a cached rotating
// token becomes invalid as soon as another process renews it.
func (s *store) locked(ctx context.Context, fn func(*saved) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.options.Directory == "" {
		return errors.New("auth directory is not set")
	}
	if err := os.MkdirAll(s.options.Directory, 0700); err != nil {
		return errors.New("cannot create private auth directory")
	}
	f, err := os.OpenFile(filepath.Join(s.options.Directory, "auth.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return errors.New("cannot open auth lock")
	}
	defer f.Close()
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK {
			return errors.New("cannot lock credentials")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	value := saved{Issuer: s.options.Issuer, Profiles: map[string]profile{}}
	data, err := os.ReadFile(filepath.Join(s.options.Directory, "auth.json"))
	if err == nil {
		if json.Unmarshal(data, &value) != nil {
			return errors.New("invalid auth store")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.New("cannot read auth store")
	}
	if value.Issuer != s.options.Issuer {
		return errors.New("auth store issuer differs from configured issuer")
	}
	if value.Profiles == nil {
		value.Profiles = map[string]profile{}
	}
	return fn(&value)
}
func (s *store) save(value *saved) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return errors.New("cannot encode credentials")
	}
	f, err := os.CreateTemp(s.options.Directory, ".auth-")
	if err != nil {
		return errors.New("cannot create credential replacement")
	}
	path := f.Name()
	defer os.Remove(path)
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		return errors.New("cannot write credential replacement")
	}
	if os.Rename(path, filepath.Join(s.options.Directory, "auth.json")) != nil {
		return errors.New("cannot install credential replacement")
	}
	return nil
}
func status(value *saved) common.AuthStatus {
	p := value.Profiles[value.Active]
	return common.AuthStatus{ClientID: p.ClientID, Subject: p.Subject, Email: p.Email, HostID: value.HostID, ExpiresAt: p.ExpiresAt, Scopes: append([]string(nil), p.Scopes...), SignedIn: p.AccessToken != "", PlanEnabled: p.AccessToken != "" && slices.Contains(p.Scopes, "chatgpt.tokens.use.direct"), Billing: value.Billing}
}

// Status exposes only verified display metadata and grant state.
func (s *store) Status(ctx context.Context) (out common.AuthStatus, err error) {
	err = s.locked(ctx, func(value *saved) error { out = status(value); return nil })
	return
}
func random(size int) ([]byte, error) { b := make([]byte, size); _, err := rand.Read(b); return b, err }
func hostID() (string, error) {
	b, err := random(16)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
