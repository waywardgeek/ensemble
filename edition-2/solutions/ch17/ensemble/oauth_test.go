package ensemble_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"ensemble/ensemble"
)

// This external OAuth peer signs real JWTs, checks PKCE and consumes rotating
// refresh tokens. It replaces only the vendor, not the credential implementation.
type oauthPeer struct {
	t             *testing.T
	server        *httptest.Server
	key           *rsa.PrivateKey
	mu            sync.Mutex
	pending       url.Values
	refresh       string
	rotations     int
	failure       string
	invalid       string
	noScope       bool
	revoked       bool
	initialExpiry int
	tokenCalls    int
	revokeFail    bool
}

func newOAuthPeer(t *testing.T) *oauthPeer {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	peer := &oauthPeer{t: t, key: key, refresh: "refresh-one", initialExpiry: 1}
	peer.mu.Lock()
	peer.server = httptest.NewServer(http.HandlerFunc(peer.serve))
	peer.mu.Unlock()
	t.Cleanup(peer.server.Close)
	return peer
}
func (p *oauthPeer) serve(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	base := p.server.URL
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		json.NewEncoder(w).Encode(map[string]any{"issuer": base, "authorization_endpoint": base + "/authorize", "token_endpoint": base + "/token", "jwks_uri": base + "/keys", "revocation_endpoint": base + "/revoke"})
	case "/keys":
		json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]any{"kid": "key", "kty": "RSA", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(p.key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(p.key.E)).Bytes())}}})
	case "/authorize":
		p.pending = r.URL.Query()
		q := p.pending
		callback, _ := url.Parse(q.Get("redirect_uri"))
		if callback.Hostname() != "127.0.0.1" || callback.Path != "/auth/callback" || q.Get("code_challenge_method") != "S256" || q.Get("nonce") == "" || q.Get("state") == "" || !strings.HasPrefix(q.Get("ext_agent_host_id"), "urn:uuid:") || !strings.Contains(q.Get("scope"), "resource.invoke") {
			p.t.Error("invalid authorization contract")
		}
		if q.Get("client_id") == "dynamic_agent_client" && q.Get("agent_name_hint") != "Ensemble" {
			p.t.Error("missing actual app name")
		}
		if q.Get("client_id") != "dynamic_agent_client" && q.Get("agent_name_hint") != "" {
			p.t.Error("returning login repeated registration metadata")
		}
		state := q.Get("state")
		if p.invalid == "state" {
			state = "wrong"
		}
		callback.RawQuery = url.Values{"state": {state}, "code": {"code-one"}, "client_id": {"issued-client"}}.Encode()
		if p.invalid == "client" {
			callback.RawQuery = url.Values{"state": {state}, "code": {"code-one"}, "client_id": {"dynamic_agent_client"}}.Encode()
		}
		http.Redirect(w, r, callback.String(), http.StatusFound)
	case "/token":
		p.tokenCalls++
		r.ParseForm()
		if r.Form.Get("client_id") != "issued-client" || r.Form.Get("resource") != "https://api.openai.com/v1" {
			p.t.Error("token exchange lost issued registration/resource")
		}
		expires := p.initialExpiry
		if r.Form.Get("grant_type") == "refresh_token" {
			if p.failure != "" {
				status := 400
				if p.failure == "temporary" {
					status = 503
				}
				w.WriteHeader(status)
				json.NewEncoder(w).Encode(map[string]string{"error": p.failure})
				return
			}
			if r.Form.Get("refresh_token") != p.refresh {
				p.t.Error("stale rotating refresh token")
			}
			if r.Form.Get("scope") != "" {
				p.t.Error("refresh scope should retain grant")
			}
			p.rotations++
			p.refresh = fmt.Sprintf("refresh-%d", p.rotations+1)
			expires = 3600
		} else {
			hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if base64.RawURLEncoding.EncodeToString(hash[:]) != p.pending.Get("code_challenge") || r.Form.Get("redirect_uri") != p.pending.Get("redirect_uri") || r.Form.Get("code") != "code-one" {
				p.t.Error("PKCE/code/redirect not bound to attempt")
			}
		}
		claims := map[string]any{"iss": base, "aud": "issued-client", "sub": "subject-one", "email": "person@example.test", "nonce": p.pending.Get("nonce"), "exp": time.Now().Add(time.Hour).Unix()}
		switch p.invalid {
		case "issuer":
			claims["iss"] = "wrong"
		case "audience":
			claims["aud"] = "wrong"
		case "nonce":
			claims["nonce"] = "wrong"
		case "expiry":
			claims["exp"] = 1
		case "subject":
			claims["sub"] = ""
		}
		token := p.sign(claims)
		if p.invalid == "signature" {
			parts := strings.Split(token, ".")
			signature, _ := base64.RawURLEncoding.DecodeString(parts[2])
			signature[0] ^= 1
			parts[2] = base64.RawURLEncoding.EncodeToString(signature)
			token = strings.Join(parts, ".")
		}
		scope := "openid chatgpt.tokens.use.direct"
		if p.noScope {
			scope = "openid"
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": fmt.Sprintf("access-secret-%d", p.rotations), "refresh_token": p.refresh, "id_token": token, "token_type": "Bearer", "expires_in": expires, "scope": scope})
	case "/revoke":
		if p.revokeFail {
			w.WriteHeader(503)
			return
		}
		r.ParseForm()
		p.revoked = r.Form.Get("token") == p.refresh && r.Form.Get("client_id") == "issued-client"
		w.WriteHeader(200)
	case "/v1/models":
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer access-secret-%d", p.rotations) {
			p.t.Error("catalog used wrong credential")
		}
		json.NewEncoder(w).Encode(map[string]any{"models": []any{map[string]string{"visibility": "list", "slug": "gpt-ch19-course", "display_name": "Course"}, map[string]string{"visibility": "hidden", "slug": "internal", "display_name": "Hidden"}}})
	default:
		http.NotFound(w, r)
	}
}
func (p *oauthPeer) sign(claims any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"key"}`))
	raw, _ := json.Marshal(claims)
	body := header + "." + base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(body))
	signature, err := rsa.SignPKCS1v15(rand.Reader, p.key, crypto.SHA256, hash[:])
	if err != nil {
		p.t.Fatal(err)
	}
	return body + "." + base64.RawURLEncoding.EncodeToString(signature)
}
func authRoot(t *testing.T, p *oauthPeer, dir string) *ensemble.Ensemble {
	root := ensemble.New(io.Discard)
	root.ConfigureAuth(ensemble.AuthOptions{Directory: dir, Issuer: p.server.URL})
	return root
}
func loginFake(root *ensemble.Ensemble) (ensemble.AuthStatus, error) {
	return root.Auth().Login(context.Background(), func(endpoint string) error {
		resp, err := http.Get(endpoint)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	})
}

// Two independently constructed stores model CLI/GUI processes sharing one file.
// Both reload under the file lock, so only one spends the rotating refresh token.
func TestOAuthLoginRefreshAndLogout(t *testing.T) {
	p := newOAuthPeer(t)
	dir := t.TempDir()
	root := authRoot(t, p, dir)
	status, err := loginFake(root)
	if err != nil {
		t.Fatal(err)
	}
	if !status.PlanEnabled || status.Subject != "subject-one" || status.Billing != ensemble.CredentialChatGPT {
		t.Fatalf("identity/grant status: %+v", status)
	}
	info, err := os.Stat(filepath.Join(dir, "auth.json"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("credentials are not owner-only")
	}
	other := authRoot(t, p, dir)
	var group sync.WaitGroup
	failures := make(chan error, 8)
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			store := root.Auth()
			if i%2 == 1 {
				store = other.Auth()
			}
			token, err := store.GetBearerToken(context.Background())
			if err == nil && token != "access-secret-1" {
				err = fmt.Errorf("wrong bearer")
			}
			failures <- err
		}(i)
	}
	group.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	if p.rotations != 1 {
		t.Fatalf("rotated %d times", p.rotations)
	}
	catalog, err := root.Auth().Models(context.Background(), p.server.URL)
	if err != nil || len(catalog) != 1 || catalog[0].Slug != "gpt-ch19-course" {
		t.Fatalf("catalog: %v %v", catalog, err)
	}
	if err := root.Auth().Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	after, err := root.Auth().Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !p.revoked || after.SignedIn || after.HostID != status.HostID || after.ClientID != status.ClientID || after.Billing != ensemble.CredentialChatGPT {
		t.Fatalf("logout changed identity or billing: %+v", after)
	}
	if _, err := root.Auth().GetBearerToken(context.Background()); err == nil {
		t.Fatal("signed-out token exposed")
	}
	again, err := loginFake(root)
	if err != nil || again.HostID != status.HostID || again.ClientID != status.ClientID {
		t.Fatalf("returning registration: %+v %v", again, err)
	}
}

// Each identity predicate is independently rejected; a valid signature alone
// cannot establish that the token belongs to this app or authorization attempt.
func TestOAuthRejectsInvalidIdentity(t *testing.T) {
	p := newOAuthPeer(t)
	for _, invalid := range []string{"state", "client", "signature", "issuer", "audience", "nonce", "expiry", "subject"} {
		t.Run(invalid, func(t *testing.T) {
			p.invalid = invalid
			root := authRoot(t, p, t.TempDir())
			if _, err := loginFake(root); err == nil {
				t.Fatal("invalid identity accepted")
			}
			status, err := root.Auth().Status(context.Background())
			if err != nil || status.SignedIn {
				t.Fatalf("invalid login installed credentials: %+v %v", status, err)
			}
		})
	}
}

// Authentication without the inference grant remains a visible signed-in state,
// but cannot return a bearer for plan use or silently select the present API key.
func TestOAuthMissingScopeAndTerminalRefresh(t *testing.T) {
	p := newOAuthPeer(t)
	p.noScope = true
	root := authRoot(t, p, t.TempDir())
	status, err := loginFake(root)
	if err != nil || !status.SignedIn || status.PlanEnabled {
		t.Fatalf("identity-only status: %+v %v", status, err)
	}
	if _, err := root.Auth().GetBearerToken(context.Background()); err == nil {
		t.Fatal("missing grant permitted inference")
	}
	p.noScope = false
	for _, code := range []string{"invalid_grant", "invalid_refresh_token", "token_expired", "refresh_token_expired", "refresh_token_invalidated", "refresh_token_reused"} {
		t.Run(code, func(t *testing.T) {
			p.failure = ""
			root := authRoot(t, p, t.TempDir())
			before, err := loginFake(root)
			if err != nil {
				t.Fatal(err)
			}
			p.failure = code
			if _, err := root.Auth().GetBearerToken(context.Background()); err == nil {
				t.Fatal("terminal refresh succeeded")
			}
			after, err := root.Auth().Status(context.Background())
			if err != nil || after.SignedIn || after.ClientID != before.ClientID || after.HostID != before.HostID || after.Billing != ensemble.CredentialChatGPT {
				t.Fatalf("terminal cleanup: %+v %v", after, err)
			}
		})
	}
}

// Successful login is an explicit subscription selection. Even with an API key
// configured, requests and a 401 renewal must stay on that selected credential.
func TestOAuthInferenceBillingAndIsolation(t *testing.T) {
	p := newOAuthPeer(t)
	p.mu.Lock()
	p.initialExpiry = 3600
	p.mu.Unlock()
	root := authRoot(t, p, t.TempDir())
	if _, err := loginFake(root); err != nil {
		t.Fatal(err)
	}
	var log strings.Builder
	root.SetAPILog(&log)
	requests := 0
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		want := fmt.Sprintf("Bearer access-secret-%d", requests-1)
		if r.Header.Get("Authorization") != want {
			t.Errorf("request used incorrect selected bearer")
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		// Production plan admission rejects the explicit-mode option. Preserve
		// the paired omission through 401 retry without weakening API-key caching.
		wire, _ := json.Marshal(body)
		if body["prompt_cache_options"] != nil || strings.Contains(string(wire), `"prompt_cache_breakpoint":`) {
			t.Error("plan explicit cache mechanism was not omitted")
		}
		for _, field := range []string{"max_output_tokens", "temperature", "previous_response_id", "user"} {
			if _, ok := body[field]; ok {
				t.Errorf("plan route included %s", field)
			}
		}
		if body["tools"].([]any)[0].(map[string]any)["type"] != "namespace" {
			t.Error("plan functions missing namespace")
		}
		if requests == 1 {
			w.WriteHeader(401)
			io.WriteString(w, `{"error":{"message":"access-secret-0"}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		frame(w, wireFrame{"response.output_text.delta", decoded(`{"output_index":0,"delta":"access-secret-1"}`)})
		frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"access-secret-1"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
	}))
	defer peer.Close()
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", APIKey: "metered-secret", BaseURL: peer.URL, DataDir: t.TempDir(), BuiltinTools: true})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	answer, err := a.Ask(context.Background(), "hello")
	if err != nil || answer != "[REDACTED]" {
		t.Fatalf("secret echo reached answer: redacted=%t err=%v", answer == "[REDACTED]", err)
	}
	if requests != 2 || p.rotations != 1 {
		t.Fatalf("401 recovery requests=%d rotations=%d", requests, p.rotations)
	}
	// Renewal retries one logical request: only its successful response spends,
	// and there is no preceding main request to rotate into the prior artifact.
	meter := a.Engine().Meter()
	wantUsage := ensemble.Usage{Input: 1, Output: 1}
	if meter.Usage != wantUsage || meter.Total != wantUsage {
		t.Fatalf("401 retry accounting: %+v", meter)
	}
	if _, err := os.Stat(filepath.Join(a.Config().DataDir, "cache-prior.txt")); !os.IsNotExist(err) {
		t.Fatalf("401 retry rotated the main request twice: %v", err)
	}
	var saved strings.Builder
	a.History().Save(&saved)
	for _, value := range []string{log.String(), saved.String()} {
		if strings.Contains(value, "access-secret") || strings.Contains(value, "metered-secret") || strings.Contains(value, "refresh-") {
			t.Fatal("credential entered trace or conversation")
		}
	}
}

// Temporary failures preserve a usable registration and billing intent, whereas
// explicit sign-out clears locally even when remote revocation is unavailable.
func TestOAuthTransientRecovery(t *testing.T) {
	p := newOAuthPeer(t)
	root := authRoot(t, p, t.TempDir())
	before, err := loginFake(root)
	if err != nil {
		t.Fatal(err)
	}
	p.failure = "temporary"
	calls := p.tokenCalls
	if _, err := root.Auth().GetBearerToken(context.Background()); err == nil {
		t.Fatal("temporary rejection accepted")
	}
	after, err := root.Auth().Status(context.Background())
	if err != nil || !after.SignedIn || after.ClientID != before.ClientID || p.tokenCalls-calls != 3 {
		t.Fatalf("transient state/retry: %+v attempts=%d err=%v", after, p.tokenCalls-calls, err)
	}
	p.failure = ""
	if _, err := root.Auth().GetBearerToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	// A subscription choice is scoped to OpenAI, not a veto on other providers.
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.Anthropic, Model: "fake", APIKey: "key", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	if a.Config().Credentials.Kind() != ensemble.CredentialAPIKey {
		t.Fatal("vendor switch retained OAuth credentials")
	}
	p.revokeFail = true
	if err := root.Auth().Logout(context.Background()); err == nil {
		t.Fatal("failed revocation was not reported")
	}
	after, err = root.Auth().Status(context.Background())
	if err != nil || after.SignedIn || after.ClientID != before.ClientID {
		t.Fatalf("local sign-out did not clear tokens: %+v %v", after, err)
	}
}

// Exercise the shipped headless command's browser handoff and restart, not just
// the public library. The original grader separately launches the GUI command.
func TestOAuthCLI(t *testing.T) {
	p := newOAuthPeer(t)
	p.mu.Lock()
	p.initialExpiry = 3600
	p.mu.Unlock()
	binary := filepath.Join(t.TempDir(), "ensemble")
	if output, err := exec.Command("go", "build", "-o", binary, "..").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	env := append(os.Environ(), "EN_HOME="+filepath.Join(dir, "auth"), "OPENAI_OIDC_ISSUER="+p.server.URL, "EN_BROWSER_CMD=true")
	cmd := exec.CommandContext(ctx, binary, "auth", "login", "--print-url")
	cmd.Dir = dir
	cmd.Env = env
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	scan := bufio.NewScanner(pipe)
	if !scan.Scan() {
		t.Fatal("login did not provide browser URL")
	}
	parsed, err := url.Parse(scan.Text())
	if err != nil || parsed.Path != "/authorize" || parsed.Query().Get("id_token_hint") != "" {
		t.Fatal("unsafe or invalid login URL")
	}
	resp, err := http.Get(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var output strings.Builder
	for scan.Scan() {
		output.WriteString(scan.Text())
		output.WriteByte('\n')
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("login command: %v %s", err, diagnostic.String())
	}
	if !strings.Contains(output.String(), "Billing: chatgpt") || strings.Contains(output.String(), "access-secret") {
		t.Fatal("login status omitted billing or leaked token")
	}
	for _, action := range []string{"status", "logout"} {
		cmd := exec.CommandContext(ctx, binary, "auth", action)
		cmd.Dir = dir
		cmd.Env = env
		data, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("auth %s: %v %s", action, err, data)
		}
		if strings.Contains(string(data), "access-secret") || strings.Contains(string(data), "refresh-one") {
			t.Fatal("auth command printed credential")
		}
	}
}

// Even caller-supplied rendered bytes can echo a credential. The real transport
// resolves the current bearer before the lens writes its readable artifact.
func TestOAuthLensRedactsResolvedBearer(t *testing.T) {
	p := newOAuthPeer(t)
	p.mu.Lock()
	p.initialExpiry = 3600
	p.mu.Unlock()
	root := authRoot(t, p, t.TempDir())
	if _, err := loginFake(root); err != nil {
		t.Fatal(err)
	}
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access-secret-0" {
			t.Error("resolved bearer not used")
		}
		w.Header().Set("Content-Type", "text/event-stream")
		frame(w, wireFrame{"response.completed", decoded(`{"response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"done"}]}],"usage":{"input_tokens":1,"output_tokens":1}}}`)})
	}))
	defer peer.Close()
	a, err := root.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", APIKey: "metered-secret", BaseURL: peer.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	body := []byte(`{"model":"gpt-ch19-course","input":[{"role":"user","content":"access-secret-0 metered-secret"}],"stream":true,"store":false}`)
	if _, err := a.Engine().Send(context.Background(), body, ensemble.StreamCallbacks{}); err != nil {
		t.Fatal(err)
	}
	artifact, err := os.ReadFile(filepath.Join(a.Config().DataDir, "cache-current.txt"))
	if err != nil || strings.Contains(string(artifact), "access-secret") || strings.Contains(string(artifact), "metered-secret") || strings.Count(string(artifact), "[REDACTED]") != 2 {
		t.Fatalf("lens artifact did not redact both resolved and configured credentials: %v", err)
	}
}
