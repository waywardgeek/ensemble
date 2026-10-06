package oauth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// browserSim stands in for the user and their browser. It receives the
// authorization URL the flow wants opened, records it for assertions, and
// then performs the loopback callback itself.
//
// It can do this synchronously because the flow starts serving the loopback
// listener BEFORE it calls the Opener, and the callback handler drops its
// result into a buffered channel rather than blocking on a reader. The
// callback has therefore already landed by the time the Opener returns.
type browserSim struct {
	t      *testing.T
	issuer *fakeIssuer

	// observations
	authQuery url.Values
	called    int

	// knobs
	issuedClientID   string // id handed back on a NEW registration
	stateOverride    string // send a different state on the callback
	clientIDOverride string // send a different client_id on the callback
	sendClientID     bool   // force sending client_id even on a reauth
	omitClientID     bool   // omit client_id even on a registration
	nonceOverride    string // bind the code to a different nonce
	errorCode        string // return error=... instead of a code
	omitCode         bool
}

func (b *browserSim) opener() Opener {
	return func(_ context.Context, raw string) error {
		b.called++
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}
		q := u.Query()
		b.authQuery = q

		cb := url.Values{}
		state := q.Get("state")
		if b.stateOverride != "" {
			state = b.stateOverride
		}
		cb.Set("state", state)

		if b.errorCode != "" {
			cb.Set("error", b.errorCode)
			cb.Set("error_description", "simulated by the test")
		} else {
			requested := q.Get("client_id")
			isRegistration := requested == BootstrapClientID

			issued := requested
			if isRegistration {
				issued = b.issuedClientID
				if issued == "" {
					issued = testIssuedClient
				}
			}

			nonce := q.Get("nonce")
			if b.nonceOverride != "" {
				nonce = b.nonceOverride
			}
			if !b.omitCode {
				cb.Set("code", b.issuer.newCode(q.Get("code_challenge"), nonce, issued))
			}

			switch {
			case b.clientIDOverride != "":
				cb.Set("client_id", b.clientIDOverride)
			case b.omitClientID:
				// deliberately absent
			case isRegistration || b.sendClientID:
				// A successful new registration returns the issued client_id.
				cb.Set("client_id", issued)
			default:
				// Reauthorization: the real issuer may omit it, and the flow
				// must retain the id already on record.
			}
		}

		resp, err := http.Get(q.Get("redirect_uri") + "?" + cb.Encode())
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}
}

// testEnv bundles a fake issuer, a simulated browser and a Config pointing at
// a temp directory.
type testEnv struct {
	issuer  *fakeIssuer
	browser *browserSim
	cfg     Config
	out     *bytes.Buffer
	dir     string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	f := newFakeIssuer(t)
	b := &browserSim{t: t, issuer: f}
	dir := t.TempDir()
	out := &bytes.Buffer{}
	return &testEnv{
		issuer:  f,
		browser: b,
		out:     out,
		dir:     dir,
		cfg: Config{
			Issuer:     f.issuerURL(),
			AgentName:  "ensemble-test",
			Store:      NewStore(filepath.Join(dir, "creds.json")),
			HostIDPath: filepath.Join(dir, "host_id"),
			HTTPClient: f.srv.Client(),
			Opener:     b.opener(),
			Output:     out,
			Now:        time.Now,
		},
	}
}

func (e *testEnv) authorize(t *testing.T) (*Credentials, error) {
	t.Helper()
	flow, err := NewFlow(e.cfg)
	if err != nil {
		return nil, err
	}
	return flow.Authorize(context.Background())
}

// mustAuthorize runs a sign-in that is expected to succeed.
func (e *testEnv) mustAuthorize(t *testing.T) *Credentials {
	t.Helper()
	creds, err := e.authorize(t)
	if err != nil {
		t.Fatalf("Authorize() failed: %v", err)
	}
	return creds
}

func TestAuthorizeHappyPath(t *testing.T) {
	env := newTestEnv(t)
	creds := env.mustAuthorize(t)

	q := env.browser.authQuery

	t.Run("first-time registration uses the bootstrap client id", func(t *testing.T) {
		if got := q.Get("client_id"); got != BootstrapClientID {
			t.Errorf("authorize client_id = %q, want %q", got, BootstrapClientID)
		}
	})

	t.Run("registration sends the required host id and name hint", func(t *testing.T) {
		hostID := q.Get("ext_agent_host_id")
		if hostID == "" {
			t.Fatal("authorize request omitted the required ext_agent_host_id")
		}
		if !strings.HasPrefix(hostID, "urn:uuid:") {
			t.Errorf("ext_agent_host_id = %q, want a urn:uuid: form", hostID)
		}
		// Opaque: it must not be derived from the user's identity.
		if strings.Contains(hostID, testEmail) || strings.Contains(hostID, testSubject) {
			t.Errorf("ext_agent_host_id %q leaks the user's identity", hostID)
		}
		if got := q.Get("agent_name_hint"); got != "ensemble-test" {
			t.Errorf("agent_name_hint = %q, want %q", got, "ensemble-test")
		}
	})

	t.Run("requests the full scope set", func(t *testing.T) {
		got := splitScopes(q.Get("scope"))
		for _, want := range DefaultScopes() {
			if !hasScope(got, want) {
				t.Errorf("authorize scope %v is missing %q", got, want)
			}
		}
	})

	t.Run("uses S256 PKCE and never sends the verifier to the browser", func(t *testing.T) {
		if got := q.Get("code_challenge_method"); got != "S256" {
			t.Errorf("code_challenge_method = %q, want S256", got)
		}
		challenge := q.Get("code_challenge")
		verifier := env.issuer.tokenForm().Get("code_verifier")
		if verifier == "" {
			t.Fatal("token exchange sent no code_verifier")
		}
		if challenge == verifier {
			t.Fatal("code_challenge equals the verifier: this is 'plain', not S256")
		}
		sum := sha256.Sum256([]byte(verifier))
		if want := base64.RawURLEncoding.EncodeToString(sum[:]); challenge != want {
			t.Errorf("code_challenge = %q, want S256(verifier) = %q", challenge, want)
		}
	})

	t.Run("redirects to a loopback literal with an ephemeral port", func(t *testing.T) {
		redirect := q.Get("redirect_uri")
		u, err := url.Parse(redirect)
		if err != nil {
			t.Fatalf("parsing redirect_uri %q: %v", redirect, err)
		}
		if u.Hostname() != "127.0.0.1" {
			t.Errorf("redirect host = %q, want the literal 127.0.0.1 (never 'localhost')", u.Hostname())
		}
		if u.Port() == "" || u.Port() == "0" {
			t.Errorf("redirect port = %q, want a bound ephemeral port", u.Port())
		}
	})

	t.Run("exchanges and persists the ISSUED client id", func(t *testing.T) {
		if creds.ClientID != testIssuedClient {
			t.Errorf("stored client_id = %q, want the issued %q", creds.ClientID, testIssuedClient)
		}
		if creds.ClientID == BootstrapClientID {
			t.Fatal("persisted the bootstrap client id as the identity")
		}
		if got := env.issuer.tokenForm().Get("client_id"); got != testIssuedClient {
			t.Errorf("token exchange client_id = %q, want the issued %q", got, testIssuedClient)
		}
	})

	t.Run("is a public client", func(t *testing.T) {
		if got := env.issuer.tokenForm().Get("client_secret"); got != "" {
			t.Errorf("token exchange sent a client_secret (%q); this is a public client", got)
		}
	})

	t.Run("records the granted scopes and the identity", func(t *testing.T) {
		if !creds.HasDirectTokenScope() {
			t.Errorf("granted scopes %v lack %s", creds.GrantedScopes, ScopeDirectTokens)
		}
		if creds.Subject != testSubject {
			t.Errorf("subject = %q, want %q", creds.Subject, testSubject)
		}
		if creds.Email != testEmail {
			t.Errorf("email = %q, want %q", creds.Email, testEmail)
		}
	})

	t.Run("persists to disk", func(t *testing.T) {
		loaded, err := env.cfg.Store.Load()
		if err != nil {
			t.Fatalf("Load() after Authorize: %v", err)
		}
		if loaded.AccessToken != creds.AccessToken || loaded.RefreshToken != creds.RefreshToken {
			t.Error("stored credentials do not match the returned ones")
		}
	})
}

// TestCallbackRejections covers every way a callback can be untrustworthy.
func TestCallbackRejections(t *testing.T) {
	tests := []struct {
		name     string
		tweak    func(b *browserSim)
		wantErr  error
		wantText string
	}{
		{
			name:    "state mismatch is rejected",
			tweak:   func(b *browserSim) { b.stateOverride = "attacker-chosen-state" },
			wantErr: ErrStateMismatch,
		},
		{
			name:    "user denial stops the flow",
			tweak:   func(b *browserSim) { b.errorCode = "access_denied" },
			wantErr: ErrAccessDenied,
		},
		{
			name:     "a registration callback without a client_id is rejected",
			tweak:    func(b *browserSim) { b.omitClientID = true },
			wantText: "did not return a client_id",
		},
		{
			name:     "the bootstrap id is never accepted as an identity",
			tweak:    func(b *browserSim) { b.clientIDOverride = BootstrapClientID },
			wantText: "bootstrap client id",
		},
		{
			name:     "a callback with neither code nor error is rejected",
			tweak:    func(b *browserSim) { b.omitCode = true; b.omitClientID = false },
			wantText: "neither an authorization code nor an error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t)
			tc.tweak(env.browser)

			_, err := env.authorize(t)
			if err == nil {
				t.Fatal("Authorize() succeeded; want an error")
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantText != "" && !strings.Contains(err.Error(), tc.wantText) {
				t.Fatalf("error = %v, want it to mention %q", err, tc.wantText)
			}

			// Nothing may be persisted from a rejected callback.
			if _, loadErr := env.cfg.Store.Load(); !errors.Is(loadErr, ErrNoCredentials) {
				t.Errorf("a rejected callback left credentials on disk (load err = %v)", loadErr)
			}
		})
	}
}

// TestStateMismatchHappensBeforeAnyExchange proves the CSRF check is a real
// gate and not a post-hoc assertion: no token call may occur at all.
func TestStateMismatchHappensBeforeAnyExchange(t *testing.T) {
	env := newTestEnv(t)
	env.browser.stateOverride = "wrong"

	if _, err := env.authorize(t); !errors.Is(err, ErrStateMismatch) {
		t.Fatalf("error = %v, want ErrStateMismatch", err)
	}
	if env.issuer.calls() != 0 {
		t.Errorf("token endpoint was called %d times after a state mismatch; want 0", env.issuer.calls())
	}
}

// TestReauthorizationRetainsIssuedClientID covers the second sign-in: the
// bootstrap id must not reappear, and a callback that omits client_id must
// not wipe the registration.
func TestReauthorizationRetainsIssuedClientID(t *testing.T) {
	env := newTestEnv(t)
	first := env.mustAuthorize(t)

	// Second sign-in. The browser omits client_id, as a real reauthorization
	// callback may.
	env.browser.omitClientID = true
	second := env.mustAuthorize(t)

	if got := env.browser.authQuery.Get("client_id"); got != first.ClientID {
		t.Errorf("reauthorization sent client_id %q, want the issued %q", got, first.ClientID)
	}
	if got := env.browser.authQuery.Get("client_id"); got == BootstrapClientID {
		t.Error("reauthorization re-registered with the bootstrap client id")
	}
	if second.ClientID != first.ClientID {
		t.Errorf("client id changed across reauthorization: %q -> %q", first.ClientID, second.ClientID)
	}
	// The retained ID token should be offered as a hint so the issuer
	// re-confirms the same account.
	if env.browser.authQuery.Get("id_token_hint") == "" {
		t.Error("reauthorization sent no id_token_hint")
	}
}

// TestCallbackClientIDSubstitutionRejected is the account-takeover defence:
// a callback must never be able to repoint an existing registration.
func TestCallbackClientIDSubstitutionRejected(t *testing.T) {
	env := newTestEnv(t)
	first := env.mustAuthorize(t)

	// A second sign-in in which the callback claims a DIFFERENT client id.
	env.browser.clientIDOverride = "client_attacker_9999"
	_, err := env.authorize(t)
	if !errors.Is(err, ErrClientIDMismatch) {
		t.Fatalf("error = %v, want ErrClientIDMismatch", err)
	}

	// The registration on disk must be untouched.
	stored, loadErr := env.cfg.Store.Load()
	if loadErr != nil {
		t.Fatalf("Load() after rejected substitution: %v", loadErr)
	}
	if stored.ClientID != first.ClientID {
		t.Errorf("stored client id was replaced: %q, want %q", stored.ClientID, first.ClientID)
	}
}

// TestHostIDIsStableAcrossSignIns checks the id is persisted before first use
// and reused afterwards.
func TestHostIDIsStableAcrossSignIns(t *testing.T) {
	env := newTestEnv(t)

	env.mustAuthorize(t)
	first := env.browser.authQuery.Get("ext_agent_host_id")

	// The file must exist on disk, written before the sign-in that used it.
	onDisk, err := os.ReadFile(env.cfg.HostIDPath)
	if err != nil {
		t.Fatalf("host id file was not persisted: %v", err)
	}
	if strings.TrimSpace(string(onDisk)) != first {
		t.Errorf("persisted host id %q != the one sent %q", strings.TrimSpace(string(onDisk)), first)
	}

	env.mustAuthorize(t)
	if second := env.browser.authQuery.Get("ext_agent_host_id"); second != first {
		t.Errorf("host id changed between sign-ins: %q -> %q", first, second)
	}
}

// TestHostIDSurvivesCredentialLoss is the reason the two files are separate.
func TestHostIDSurvivesCredentialLoss(t *testing.T) {
	env := newTestEnv(t)
	env.mustAuthorize(t)
	first := env.browser.authQuery.Get("ext_agent_host_id")

	if err := env.cfg.Store.Clear(); err != nil {
		t.Fatalf("Clear(): %v", err)
	}
	env.mustAuthorize(t)

	if got := env.browser.authQuery.Get("ext_agent_host_id"); got != first {
		t.Errorf("host id regenerated after credential loss: %q -> %q", first, got)
	}
}

// TestNoBrowserPrintsURL covers the headless / SSH seam.
func TestNoBrowserPrintsURL(t *testing.T) {
	env := newTestEnv(t)

	// Drop the test opener so normalize() picks one based on NoBrowser, which
	// is what a real headless caller does.
	var printed bytes.Buffer
	cfg := env.cfg
	cfg.Opener = nil
	cfg.NoBrowser = true
	cfg.Output = &printed
	cfg.CallbackTimeout = 200 * time.Millisecond

	flow, err := NewFlow(cfg)
	if err != nil {
		t.Fatalf("NewFlow(): %v", err)
	}
	// No browser will answer, so the flow times out. What matters is that the
	// URL reached the writer instead of a subprocess.
	_, err = flow.Authorize(context.Background())
	if err == nil {
		t.Fatal("Authorize() succeeded without any callback")
	}

	out := printed.String()
	if !strings.Contains(out, env.issuer.issuerURL()+"/api/accounts/authorize") {
		t.Errorf("headless mode did not print the authorization URL; got:\n%s", out)
	}
	if !strings.Contains(out, "code_challenge_method=S256") {
		t.Errorf("printed URL lacks the S256 challenge method; got:\n%s", out)
	}
}

// TestIssuerOverride covers the enterprise / testing seam.
func TestIssuerOverride(t *testing.T) {
	t.Run("env var wins", func(t *testing.T) {
		t.Setenv(IssuerEnvVar, "https://auth.enterprise.example/")
		if got := Issuer(); got != "https://auth.enterprise.example" {
			t.Errorf("Issuer() = %q, want the override with the trailing slash trimmed", got)
		}
	})
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv(IssuerEnvVar, "")
		if got := Issuer(); got != DefaultIssuer {
			t.Errorf("Issuer() = %q, want %q", got, DefaultIssuer)
		}
	})
	t.Run("flow honours the env var", func(t *testing.T) {
		env := newTestEnv(t)
		t.Setenv(IssuerEnvVar, env.issuer.issuerURL())
		cfg := env.cfg
		cfg.Issuer = "" // force the env lookup
		flow, err := NewFlow(cfg)
		if err != nil {
			t.Fatalf("NewFlow(): %v", err)
		}
		if _, err := flow.Authorize(context.Background()); err != nil {
			t.Fatalf("Authorize() against the overridden issuer: %v", err)
		}
	})
}

// TestBrowserCommandOverride covers EN_BROWSER_CMD.
func TestBrowserCommandOverride(t *testing.T) {
	t.Setenv(BrowserCmdEnvVar, "/usr/bin/my-browser --private")
	name, args, ok := browserCommand()
	if !ok {
		t.Fatal("browserCommand() reported no command")
	}
	if name != "/usr/bin/my-browser" {
		t.Errorf("command = %q, want /usr/bin/my-browser", name)
	}
	if len(args) != 1 || args[0] != "--private" {
		t.Errorf("args = %v, want [--private]", args)
	}
}

// TestDiscoveryIssuerMismatch: a document must not be able to claim an
// issuer other than the host it was served from.
func TestDiscoveryIssuerMismatch(t *testing.T) {
	env := newTestEnv(t)
	_, err := FetchDiscovery(context.Background(), env.issuer.srv.Client(),
		strings.TrimSuffix(env.issuer.issuerURL(), "/")+"")
	if err != nil {
		t.Fatalf("baseline FetchDiscovery failed: %v", err)
	}

	// Point at a server whose document claims someone else's issuer.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"issuer":                 "https://auth.openai.com",
			"authorization_endpoint": "https://evil.example/authorize",
			"token_endpoint":         "https://evil.example/token",
			"jwks_uri":               "https://evil.example/jwks",
		})
	}))
	defer srv.Close()

	_, err = FetchDiscovery(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Fatal("FetchDiscovery accepted a document claiming a different issuer")
	}
	if !strings.Contains(err.Error(), "claims issuer") {
		t.Errorf("error = %v, want it to name the issuer mismatch", err)
	}
}

// TestPKCEVerifierIsFreshAndStrong.
func TestPKCEVerifier(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		p, err := newPKCE()
		if err != nil {
			t.Fatalf("newPKCE(): %v", err)
		}
		if len(p.Verifier) < 43 {
			t.Errorf("verifier %q is shorter than RFC 7636's 43-character minimum", p.Verifier)
		}
		if seen[p.Verifier] {
			t.Fatalf("newPKCE() repeated a verifier after %d calls", i)
		}
		seen[p.Verifier] = true

		sum := sha256.Sum256([]byte(p.Verifier))
		if want := base64.RawURLEncoding.EncodeToString(sum[:]); p.Challenge != want {
			t.Errorf("challenge = %q, want %q", p.Challenge, want)
		}
	}
}

// TestResolveClientID exercises the registration-identity rules directly.
func TestResolveClientID(t *testing.T) {
	tests := []struct {
		name         string
		pending      string
		fromCallback string
		want         string
		wantErr      error
		wantErrText  string
	}{
		{
			name:         "new registration adopts the issued id",
			pending:      "",
			fromCallback: "client_new",
			want:         "client_new",
		},
		{
			name:         "reauthorization retains the id when the callback omits it",
			pending:      "client_known",
			fromCallback: "",
			want:         "client_known",
		},
		{
			name:         "reauthorization accepts a matching id",
			pending:      "client_known",
			fromCallback: "client_known",
			want:         "client_known",
		},
		{
			name:         "substitution is rejected",
			pending:      "client_known",
			fromCallback: "client_other",
			wantErr:      ErrClientIDMismatch,
		},
		{
			name:         "a new registration must return an id",
			pending:      "",
			fromCallback: "",
			wantErrText:  "did not return a client_id",
		},
		{
			name:         "the bootstrap id is never an identity",
			pending:      "",
			fromCallback: BootstrapClientID,
			wantErrText:  "bootstrap client id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveClientID(tc.pending, tc.fromCallback)
			switch {
			case tc.wantErr != nil:
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}
			case tc.wantErrText != "":
				if err == nil || !strings.Contains(err.Error(), tc.wantErrText) {
					t.Fatalf("error = %v, want it to mention %q", err, tc.wantErrText)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Errorf("client id = %q, want %q", got, tc.want)
				}
			}
		})
	}
}
