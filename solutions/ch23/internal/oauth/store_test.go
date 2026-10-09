package oauth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sampleCredentials() *Credentials {
	return &Credentials{
		Issuer:        "https://auth.openai.com",
		ClientID:      testIssuedClient,
		HostID:        "urn:uuid:11111111-2222-4333-8444-555555555555",
		AccessToken:   "super-secret-access-token",
		RefreshToken:  "super-secret-refresh-token",
		IDToken:       "super-secret-id-token",
		GrantedScopes: splitScopes(allScopes),
		ExpiresAt:     time.Now().Add(time.Hour).Truncate(time.Second),
		Subject:       testSubject,
		Email:         testEmail,
	}
}

// TestCredentialFileIsOwnerOnly: a long-lived refresh token in a
// world-readable file is a credential leak no other care can compensate for.
func TestCredentialFileIsOwnerOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "creds.json")
	store := NewStore(path)

	if err := store.Save(sampleCredentials()); err != nil {
		t.Fatalf("Save(): %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(): %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("credential file mode = %#o, want 0600", got)
	}

	// The directory must not be listable by others either, or the file's
	// existence leaks even though its contents do not.
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat(dir): %v", err)
	}
	if got := dirInfo.Mode().Perm(); got&0o077 != 0 {
		t.Errorf("credential directory mode = %#o, want no group/other access", got)
	}
}

// TestAuthorizeWritesOwnerOnlyFile checks the mode on the file the real flow
// produces, not just one written directly by a unit test.
func TestAuthorizeWritesOwnerOnlyFile(t *testing.T) {
	env := newTestEnv(t)
	env.mustAuthorize(t)

	info, err := os.Stat(env.cfg.Store.Path())
	if err != nil {
		t.Fatalf("Stat(): %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("credential file mode after Authorize = %#o, want 0600", got)
	}
}

// TestStoreRejectsPermissiveFile: failing loudly is the only way the user
// ever finds out their tokens are exposed.
func TestStoreRejectsPermissiveFile(t *testing.T) {
	for _, mode := range []os.FileMode{0o644, 0o640, 0o604, 0o666} {
		t.Run(mode.String(), func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "creds.json")
			store := NewStore(path)
			if err := store.Save(sampleCredentials()); err != nil {
				t.Fatalf("Save(): %v", err)
			}
			if err := os.Chmod(path, mode); err != nil {
				t.Fatalf("Chmod(): %v", err)
			}

			_, err := store.Load()
			if err == nil {
				t.Fatalf("Load() accepted a credential file with mode %#o", mode)
			}
			if !strings.Contains(err.Error(), "owner-only") {
				t.Errorf("error = %v, want it to explain the permission requirement", err)
			}
		})
	}
}

// TestStoreWriteIsAtomic: the rename must leave no temp file behind, because
// a stray temp file is a second, possibly readable copy of the tokens.
func TestStoreWriteIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")
	store := NewStore(path)

	for i := 0; i < 3; i++ {
		if err := store.Save(sampleCredentials()); err != nil {
			t.Fatalf("Save(): %v", err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(): %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("Save() left a temp file behind: %s", e.Name())
		}
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory contains %v, want only the credential file", names)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "creds.json"))
	want := sampleCredentials()

	if err := store.Save(want); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken ||
		got.ClientID != want.ClientID || got.HostID != want.HostID ||
		got.Subject != want.Subject || got.Email != want.Email {
		t.Errorf("round trip lost data:\n got %s\nwant %s", got, want)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Errorf("ExpiresAt = %s, want %s", got.ExpiresAt, want.ExpiresAt)
	}
	if !got.HasDirectTokenScope() {
		t.Errorf("granted scopes lost in round trip: %v", got.GrantedScopes)
	}
}

func TestStoreLoadMissingIsDistinct(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "absent.json"))
	_, err := store.Load()
	if !errors.Is(err, ErrNoCredentials) {
		t.Fatalf("Load() = %v, want ErrNoCredentials", err)
	}
}

func TestStoreLoadCorruptIsNotMistakenForMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}
	_, err := NewStore(path).Load()
	if err == nil {
		t.Fatal("Load() accepted a corrupt file")
	}
	if errors.Is(err, ErrNoCredentials) {
		t.Error("a corrupt credential file was reported as 'never signed in', hiding a real fault")
	}
}

func TestStoreClear(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(filepath.Join(dir, "creds.json"))
	if err := store.Save(sampleCredentials()); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear(): %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrNoCredentials) {
		t.Errorf("Load() after Clear() = %v, want ErrNoCredentials", err)
	}
	// Clearing an absent file is success: the caller's intent is satisfied.
	if err := store.Clear(); err != nil {
		t.Errorf("Clear() on a missing file = %v, want nil", err)
	}
}

// TestCredentialsStringRedactsSecrets. Go reaches for String() automatically
// in %v and %s, so the safe rendering must be the DEFAULT one — otherwise
// one careless log line leaks a bearer token.
func TestCredentialsStringRedactsSecrets(t *testing.T) {
	c := sampleCredentials()
	rendered := c.String()

	for _, secret := range []string{c.AccessToken, c.RefreshToken, c.IDToken} {
		if strings.Contains(rendered, secret) {
			t.Errorf("String() leaked a secret: %s", rendered)
		}
	}
	if !strings.Contains(rendered, "REDACTED") {
		t.Errorf("String() = %q, want it to mark the secrets as redacted", rendered)
	}
	// Non-secret context should survive, or the rendering is useless.
	if !strings.Contains(rendered, c.ClientID) {
		t.Errorf("String() = %q, want it to include the client id", rendered)
	}

	// The same must hold when formatted through fmt verbs, which is how a
	// credential would actually reach a log line.
	if strings.Contains(fmt.Sprintf("%v", c), c.AccessToken) {
		t.Error("default fmt formatting leaked the access token")
	}
	if strings.Contains(fmt.Sprintf("%s", c), c.RefreshToken) {
		t.Error("string fmt formatting leaked the refresh token")
	}
}

func TestNeedsRefresh(t *testing.T) {
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	skew := 2 * time.Minute

	tests := []struct {
		name  string
		creds *Credentials
		want  bool
	}{
		{
			name:  "nil needs refresh",
			creds: nil,
			want:  true,
		},
		{
			name:  "empty access token needs refresh",
			creds: &Credentials{ExpiresAt: now.Add(time.Hour)},
			want:  true,
		},
		{
			name:  "zero expiry needs refresh rather than being treated as immortal",
			creds: &Credentials{AccessToken: "t"},
			want:  true,
		},
		{
			name:  "comfortably valid",
			creds: &Credentials{AccessToken: "t", ExpiresAt: now.Add(time.Hour)},
			want:  false,
		},
		{
			name:  "inside the skew window refreshes proactively",
			creds: &Credentials{AccessToken: "t", ExpiresAt: now.Add(30 * time.Second)},
			want:  true,
		},
		{
			name:  "exactly at the skew boundary refreshes",
			creds: &Credentials{AccessToken: "t", ExpiresAt: now.Add(skew)},
			want:  true,
		},
		{
			name:  "already expired",
			creds: &Credentials{AccessToken: "t", ExpiresAt: now.Add(-time.Second)},
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.creds.NeedsRefresh(now, skew); got != tc.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHostIDFileIsOwnerOnlyAndStable.
func TestHostIDFileIsOwnerOnlyAndStable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "host_id")

	first, err := LoadOrCreateHostID(path)
	if err != nil {
		t.Fatalf("LoadOrCreateHostID(): %v", err)
	}
	if !strings.HasPrefix(first, "urn:uuid:") {
		t.Errorf("host id = %q, want a urn:uuid: form", first)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(): %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("host id file mode = %#o, want 0600", got)
	}

	second, err := LoadOrCreateHostID(path)
	if err != nil {
		t.Fatalf("second LoadOrCreateHostID(): %v", err)
	}
	if second != first {
		t.Errorf("host id changed on reload: %q -> %q", first, second)
	}
}

// TestHostIDsAreUniquePerHost: "distinct per host" is the requirement, so two
// fresh hosts must not collide.
func TestHostIDsAreUniquePerHost(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		id, err := LoadOrCreateHostID(filepath.Join(t.TempDir(), "host_id"))
		if err != nil {
			t.Fatalf("LoadOrCreateHostID(): %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate host id %q after %d hosts", id, i)
		}
		seen[id] = true
	}
}

// TestHostIDRecoversFromAnEmptyFile: a truncated earlier write must not
// produce an empty required parameter.
func TestHostIDRecoversFromAnEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "host_id")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}
	id, err := LoadOrCreateHostID(path)
	if err != nil {
		t.Fatalf("LoadOrCreateHostID(): %v", err)
	}
	if id == "" {
		t.Fatal("LoadOrCreateHostID() returned an empty host id")
	}
}

// TestTokenErrorClassification pins the six terminal codes at the unit level,
// including the rule that an unparseable body is NOT terminal.
func TestTokenErrorClassification(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"invalid_grant", true},
		{"invalid_refresh_token", true},
		{"token_expired", true},
		{"refresh_token_expired", true},
		{"refresh_token_invalidated", true},
		{"refresh_token_reused", true},

		{"invalid_client", false}, // fix client config, do not re-auth
		{"temporarily_unavailable", false},
		{"server_error", false},
		{"slow_down", false},
		{"", false}, // a {"detail": ...} body with no machine-readable code
	}
	for _, tc := range tests {
		name := tc.code
		if name == "" {
			name = "no code"
		}
		t.Run(name, func(t *testing.T) {
			te := &TokenError{Code: tc.code, Status: 400}
			if got := te.Terminal(); got != tc.want {
				t.Errorf("Terminal() for %q = %v, want %v", tc.code, got, tc.want)
			}
			if got := IsTerminalRefreshError(te); got != tc.want {
				t.Errorf("IsTerminalRefreshError() for %q = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

// TestParseTokenErrorHandlesNonConformingBodies: the first-party docs warn
// that direct-admission failures return {"detail": "..."} with no error code.
// A parser that assumes error.code exists crashes on exactly the failure it
// most needs to report.
func TestParseTokenError(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode string
		wantText string
	}{
		{
			name:     "RFC 6749 error object",
			body:     `{"error":"invalid_grant","error_description":"code expired"}`,
			wantCode: "invalid_grant",
			wantText: "code expired",
		},
		{
			name:     "detail-only body",
			body:     `{"detail":"not admitted"}`,
			wantCode: "",
			wantText: "not admitted",
		},
		{
			name:     "HTML error page",
			body:     `<html><body>502 Bad Gateway</body></html>`,
			wantCode: "",
			wantText: "502 Bad Gateway",
		},
		{
			name:     "empty body",
			body:     ``,
			wantCode: "",
			wantText: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			te := parseTokenError(400, []byte(tc.body))
			if te.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", te.Code, tc.wantCode)
			}
			if tc.wantText != "" && !strings.Contains(te.Description, tc.wantText) {
				t.Errorf("Description = %q, want it to contain %q", te.Description, tc.wantText)
			}
			// Must never panic and must always render.
			if te.Error() == "" {
				t.Error("Error() returned an empty string")
			}
		})
	}
}
