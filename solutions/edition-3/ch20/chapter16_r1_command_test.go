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
	"ensemble"
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
)

// Binary paths are supplied by the focused command; tests never build or launch
// an installed application, and the only identity issuer is this local peer.
func TestCh16R1HeadlessCommands(t *testing.T) {
	cli, gui := os.Getenv("CH16_TEST_CLI"), os.Getenv("CH16_TEST_GUI")
	if cli == "" || gui == "" {
		t.Skip("set CH16_TEST_CLI and CH16_TEST_GUI to owned test binaries")
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	var mu sync.Mutex
	var nonce, challenge, redirect, host string
	registered := false
	var peer *httptest.Server
	peer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{"issuer": peer.URL, "authorization_endpoint": peer.URL + "/authorize", "token_endpoint": peer.URL + "/token", "jwks_uri": peer.URL + "/keys"})
		case "/keys":
			json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]string{"kty": "RSA", "kid": "test", "alg": "RS256", "n": enc(key.PublicKey.N.Bytes()), "e": "AQAB"}}})
		case "/authorize":
			q := r.URL.Query()
			want := "dynamic_agent_client"
			if registered {
				want = "issued-client"
			}
			if q.Get("client_id") != want || q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid profile email offline_access resource.invoke chatgpt.tokens.use.direct" || !strings.HasPrefix(q.Get("ext_agent_host_id"), "urn:uuid:") {
				t.Error("wrong issued-client/PKCE request")
				http.Error(w, "invalid request", 400)
				return
			}
			if host != "" && host != q.Get("ext_agent_host_id") {
				t.Error("host identity changed")
			}
			host = q.Get("ext_agent_host_id")
			nonce = q.Get("nonce")
			challenge = q.Get("code_challenge")
			redirect = q.Get("redirect_uri")
			target, _ := url.Parse(redirect)
			v := target.Query()
			v.Set("state", q.Get("state"))
			v.Set("code", "dummy-code")
			v.Set("client_id", "issued-client")
			v.Set("iss", peer.URL)
			target.RawQuery = v.Encode()
			http.Redirect(w, r, target.String(), http.StatusFound)
		case "/token":
			r.ParseForm()
			sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if enc(sum[:]) != challenge || r.Form.Get("client_id") != "issued-client" || r.Form.Get("redirect_uri") != redirect {
				t.Error("invalid PKCE/token identity")
				http.Error(w, "invalid", 400)
				return
			}
			claims, _ := json.Marshal(map[string]any{"iss": peer.URL, "sub": "headless-subject", "aud": "issued-client", "nonce": nonce, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()})
			unsigned := enc([]byte(`{"alg":"RS256","kid":"test"}`)) + "." + enc(claims)
			digest := sha256.Sum256([]byte(unsigned))
			sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
			registered = true
			json.NewEncoder(w).Encode(map[string]any{"access_token": "dummy-headless-access", "refresh_token": "dummy-headless-refresh", "token_type": "Bearer", "expires_in": 3600, "scope": "resource.invoke chatgpt.tokens.use.direct", "id_token": unsigned + "." + enc(sig)})
		default:
			http.NotFound(w, r)
		}
	}))
	defer peer.Close()
	home := t.TempDir()
	if err := os.Chmod(home, 0700); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	for _, name := range []string{"open", "xdg-open"} {
		if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\necho attempted > \"$CH16_BROWSER_ATTEMPT\"\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name, path string
		flags      []string
	}{{"CLI print URL", cli, []string{"--print-url"}}, {"GUI print URL", gui, []string{"--print-url"}}, {"CLI suppressed launch", cli, nil}, {"GUI suppressed launch", gui, nil}} {
		t.Run(tc.name, func(t *testing.T) {
			marker := filepath.Join(t.TempDir(), "browser-attempt")
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			args := append([]string{"auth", "login"}, tc.flags...)
			cmd := exec.CommandContext(ctx, tc.path, args...)
			cmd.Env = append(os.Environ(), "EN_HOME="+home, "OPENAI_OIDC_ISSUER="+peer.URL, "EN_BROWSER_CMD=true", "PATH="+bin+":/usr/bin:/bin", "CH16_BROWSER_ATTEMPT="+marker)
			var diagnostics bytes.Buffer
			cmd.Stderr = &diagnostics
			pipe, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err = cmd.Start(); err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(pipe)
			var output strings.Builder
			visited := false
			for scanner.Scan() {
				line := scanner.Text()
				output.WriteString(line + "\n")
				if strings.HasPrefix(line, peer.URL+"/authorize?") {
					res, e := http.Get(line)
					if e != nil {
						t.Error(e)
						cancel()
						break
					}
					res.Body.Close()
					if res.StatusCode != 200 {
						t.Errorf("loopback callback status %d", res.StatusCode)
						cancel()
					}
					visited = true
				}
			}
			err = cmd.Wait()
			if err != nil || !visited {
				t.Fatalf("headless command failed: %v; %s; %s", err, output.String(), diagnostics.String())
			}
			if _, err = os.Stat(marker); err == nil {
				t.Fatal("EN_BROWSER_CMD=true did not suppress browser launch")
			}
			if strings.Contains(output.String()+diagnostics.String(), "dummy-headless-") {
				t.Fatal("token leaked to command output")
			}
			status, err := ensemble.NewAuth(ensemble.AuthOptions{Directory: home, Issuer: peer.URL}).Status(context.Background())
			if err != nil || status.Mode != ensemble.AuthChatGPT || !status.Eligible || status.Account != "headless-subject" {
				t.Fatal("issued identity/grants not retained", status, err)
			}
		})
	}
}
