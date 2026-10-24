package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble/internal/common"
)

type oauthPeer struct {
	server                   *httptest.Server
	key                      *rsa.PrivateKey
	options                  Options
	nonce, challenge         string
	registrations, refreshes atomic.Int32
	reject                   atomic.Bool
	transient                atomic.Bool
	noGrant                  bool
}

func newPeer(t *testing.T) *oauthPeer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	redirect := "http://" + listener.Addr().String() + "/auth/callback"
	listener.Close()
	p := &oauthPeer{key: key}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]string{"issuer": p.server.URL, "authorization_endpoint": p.server.URL + "/authorize", "token_endpoint": p.server.URL + "/oauth/token", "jwks_uri": p.server.URL + "/.well-known/jwks.json", "revocation_endpoint": p.server.URL + "/revoke"})
		case "/revoke":
			w.WriteHeader(200)
		case "/.well-known/jwks.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": "local-key", "alg": "RS256", "use": "sig", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": "AQAB"}}})
		case "/oauth/token":
			_ = r.ParseForm()
			if r.Form.Get("client_id") != "issued-local-client" {
				t.Error("unissued client")
			}
			token := "access-one"
			refresh := "refresh-one"
			expires := 3600
			if r.Form.Get("grant_type") == "authorization_code" {
				sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				if base64.RawURLEncoding.EncodeToString(sum[:]) != p.challenge || r.Form.Get("redirect_uri") != redirect || r.Form.Get("code") != "local-code" {
					t.Error("PKCE/callback/code mismatch")
				}
			} else {
				n := p.refreshes.Add(1)
				time.Sleep(30 * time.Millisecond)
				if p.transient.Load() {
					w.WriteHeader(503)
					return
				}
				if p.reject.Load() {
					w.WriteHeader(400)
					fmt.Fprint(w, `{"error":"invalid_grant","error_description":"refresh-secret"}`)
					return
				}
				if r.Form.Get("refresh_token") != "refresh-one" {
					t.Errorf("stale refresh attempt %q", r.Form.Get("refresh_token"))
				}
				token = fmt.Sprintf("access-rotated-%d", n)
				refresh = "refresh-two"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": token, "refresh_token": refresh, "token_type": "Bearer", "expires_in": expires, "id_token": p.sign(t, p.claims()), "scope": func() string {
				if p.noGrant {
					return "openid profile offline_access"
				}
				return loginScopes
			}()})
		default:
			http.NotFound(w, r)
		}
	}))
	p.options = Options{Directory: filepath.Join(t.TempDir(), "auth"), Issuer: p.server.URL, RedirectURI: redirect}
	t.Cleanup(p.server.Close)
	return p
}
func (p *oauthPeer) claims() map[string]any {
	return map[string]any{"iss": p.server.URL, "sub": "local-subject", "aud": "issued-local-client", "nonce": p.nonce, "iat": time.Now().Unix() - 10, "exp": time.Now().Unix() + 3600, "https://api.openai.com/auth": map[string]any{"chatgpt_account_id": "local-account", "chatgpt_plan_type": "plus"}}
}
func (p *oauthPeer) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "local-key"})
	payload, _ := json.Marshal(claims)
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	hash := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}
func (p *oauthPeer) login(t *testing.T, s *Service) common.AuthStatus {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	status, err := s.Login(ctx, func(raw string) error {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}
		q := u.Query()
		p.nonce = q.Get("nonce")
		p.challenge = q.Get("code_challenge")
		if p.nonce == "" || q.Get("state") == "" || q.Get("code_challenge_method") != "S256" || (q.Get("client_id") != "issued-local-client" && q.Get("client_id") != bootstrapClient) {
			return fmt.Errorf("invalid authorization URL")
		}
		if q.Get("client_id") == bootstrapClient {
			p.registrations.Add(1)
		}
		if !strings.HasPrefix(q.Get("ext_agent_host_id"), "urn:uuid:") || q.Get("resource") != resource || !strings.Contains(q.Get("scope"), inferenceScope) {
			return fmt.Errorf("missing registration/resource/scope")
		}
		callback := q.Get("redirect_uri")
		for _, values := range []url.Values{{"state": {"wrong"}, "code": {"local-code"}}, {"state": {q.Get("state")}, "code": {"local-code"}, "iss": {"wrong"}}} {
			r, err := http.Get(callback + "?" + values.Encode())
			if err != nil {
				return err
			}
			r.Body.Close()
			if r.StatusCode != 400 {
				return fmt.Errorf("invalid callback accepted")
			}
		}
		r, err := http.Get(callback + "?" + url.Values{"state": {q.Get("state")}, "code": {"local-code"}, "iss": {p.server.URL}, "client_id": {"issued-local-client"}}.Encode())
		if err != nil {
			return err
		}
		r.Body.Close()
		if r.StatusCode != 200 {
			return fmt.Errorf("valid callback rejected")
		}
		return nil
	})
	if err != nil || status.Eligible == p.noGrant || status.Account != "local-subject" {
		t.Fatalf("login: %+v %v", status, err)
	}
	return status
}
func expire(t *testing.T, s *Service) {
	t.Helper()
	if err := s.locked(context.Background(), func(st *store) error {
		g := st.Accounts[st.Selected]
		g.Expires = time.Now().Add(-time.Second)
		st.Accounts[st.Selected] = g
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
func TestAuthLoginRefreshAcrossProcesses(t *testing.T) {
	p := newPeer(t)
	s := New(p.options)
	p.login(t, s)
	p.login(t, New(p.options))
	if p.registrations.Load() != 1 {
		t.Fatal("host client was not reused")
	}
	cfg := common.Config{AuthMode: common.AuthChatGPT, Surface: common.Responses, APIKey: "dummy-api"}
	stale, err := s.Resolve(context.Background(), cfg)
	if err != nil || stale.Token != "access-one" {
		t.Fatal(err, stale)
	}
	expire(t, s)
	encoded, _ := json.Marshal(p.options)
	commands := []*exec.Cmd{}
	for i := 0; i < 3; i++ {
		cmd := exec.Command(os.Args[0], "-test.run=^TestAuthRefreshProcess$")
		cmd.Env = append(os.Environ(), "ENSEMBLE_AUTH_CHILD="+string(encoded))
		commands = append(commands, cmd)
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
	}
	for _, cmd := range commands {
		if err := cmd.Wait(); err != nil {
			t.Fatal("refresh subprocess", err)
		}
	}
	if p.refreshes.Load() != 1 {
		t.Fatalf("refresh count=%d; writer did not reread", p.refreshes.Load())
	}
	if err := s.Reject(context.Background(), stale); err != nil {
		t.Fatal(err)
	}
	current, err := New(p.options).Resolve(context.Background(), cfg)
	if err != nil || current.Token != "access-rotated-1" {
		t.Fatal("stale rejection cleared rotation", err)
	}
	status, err := s.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(status)
	for _, secret := range []string{"access-", "refresh-", "private_key", "issued-local-client"} {
		if strings.Contains(string(b), secret) {
			t.Fatal("status leaks credential")
		}
	}
	for _, name := range []string{"profiles.json", "writer.lock"} {
		info, err := os.Stat(filepath.Join(p.options.Directory, name))
		if err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("profile permissions", info, err)
		}
	}
	if err := s.Logout(context.Background(), "local-subject", false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(context.Background(), common.Config{AuthMode: common.AuthChatGPT, Surface: common.Responses, APIKey: "dummy-api"}); err == nil {
		t.Fatal("logout still authorized")
	}
	api, err := s.Resolve(context.Background(), common.Config{AuthMode: common.AuthAPI, APIKey: "dummy-api"})
	if err != nil || api.Token != "dummy-api" {
		t.Fatal("explicit API selection", err)
	}
}
func TestAuthRefreshProcess(t *testing.T) {
	raw := os.Getenv("ENSEMBLE_AUTH_CHILD")
	if raw == "" {
		t.Skip("subprocess helper")
	}
	var o Options
	if json.Unmarshal([]byte(raw), &o) != nil {
		t.Fatal("options")
	}
	c, err := New(o).Resolve(context.Background(), common.Config{AuthMode: common.AuthChatGPT, Surface: common.Responses})
	if err != nil || c.Token != "access-rotated-1" {
		t.Fatal("child resolve", err)
	}
}
func TestAuthIdentityRejectionsAndRevocation(t *testing.T) {
	p := newPeer(t)
	s := New(p.options)
	p.login(t, s)
	for _, name := range []string{"issuer", "audience", "azp", "subject", "expiry", "issued", "not_before", "nonce", "signature"} {
		t.Run(name, func(t *testing.T) {
			c := p.claims()
			switch name {
			case "issuer":
				c["iss"] = "wrong"
			case "audience":
				c["aud"] = "wrong"
			case "azp":
				c["aud"] = []string{"issued-local-client", "other"}
			case "subject":
				c["sub"] = ""
			case "account":
				c["https://api.openai.com/auth"] = map[string]string{"chatgpt_plan_type": "plus"}
			case "expiry":
				c["exp"] = time.Now().Unix() - 1
			case "issued":
				c["iat"] = time.Now().Unix() + 100
			case "not_before":
				c["nbf"] = time.Now().Unix() + 100
			case "nonce":
				c["nonce"] = "wrong"
			}
			token := p.sign(t, c)
			if name == "signature" {
				pieces := strings.Split(token, ".")
				pieces[1] = base64.RawURLEncoding.EncodeToString([]byte(`{}`))
				token = strings.Join(pieces, ".")
			}
			if _, err := s.verify(context.Background(), token, "issued-local-client", p.nonce); err == nil {
				t.Fatal("invalid identity accepted")
			}
		})
	}
	expire(t, s)
	p.reject.Store(true)
	cfg := common.Config{AuthMode: common.AuthChatGPT, Surface: common.Responses, APIKey: "dummy-api"}
	for i := 0; i < 2; i++ {
		if _, err := New(p.options).Resolve(context.Background(), cfg); err == nil {
			t.Fatal("revoked grant fell through to API")
		}
	}
	if p.refreshes.Load() != 1 {
		t.Fatal("revocation repeatedly refreshed")
	}
	b, err := os.ReadFile(filepath.Join(p.options.Directory, "profiles.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "access-one") || strings.Contains(string(b), "refresh-one") {
		t.Fatal("revoked credentials persisted")
	}
}
func TestAuthProtectedStorage(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "auth")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	s := New(Options{Directory: dir})
	if _, err := s.Status(context.Background()); err == nil {
		t.Fatal("world-readable directory accepted")
	}
	if err := os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "target")
	if err := os.WriteFile(target, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "profiles.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Status(context.Background()); err == nil {
		t.Fatal("symlink profile accepted")
	}
}

func TestAuthIdentityWithoutInferenceGrant(t *testing.T) {
	p := newPeer(t)
	p.noGrant = true
	s := New(p.options)
	st := p.login(t, s)
	if st.Eligible || st.Mode != common.AuthChatGPT || st.Account != "local-subject" {
		t.Fatal("identity and inference eligibility conflated", st)
	}
	if _, err := s.Resolve(context.Background(), common.Config{Vendor: common.OpenAI, Surface: common.Responses, APIKey: "dummy-api"}); err == nil {
		t.Fatal("identity-only login inferred plan permission or fell back")
	}
	if st, err := New(p.options).Status(context.Background()); err != nil || st.Eligible || st.Mode != common.AuthChatGPT {
		t.Fatal("ineligible persisted selection lost", st, err)
	}
}
func TestAuthTransientRefreshPreservesGrant(t *testing.T) {
	p := newPeer(t)
	s := New(p.options)
	p.login(t, s)
	if err := s.locked(context.Background(), func(st *store) error {
		g := st.Accounts[st.Selected]
		g.Expires = time.Now().Add(-time.Minute)
		st.Accounts[st.Selected] = g
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	p.transient.Store(true)
	c := common.Config{Vendor: common.OpenAI, Surface: common.Responses, AuthMode: common.AuthChatGPT}
	if _, err := s.Resolve(context.Background(), c); err == nil {
		t.Fatal("transient refresh unexpectedly succeeded")
	}
	p.transient.Store(false)
	if _, err := New(p.options).Resolve(context.Background(), c); err != nil {
		t.Fatal("temporary outage destroyed grant", err)
	}
}
