package ensemble_test

import (
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
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"ensemble"
)

func TestCh16PlanBillingTransport(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	enc := base64.RawURLEncoding.EncodeToString
	var nonce, challenge, redirect string
	var requests, refreshes atomic.Int32
	var host *httptest.Server
	host = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{"issuer": host.URL, "authorization_endpoint": host.URL + "/authorize", "token_endpoint": host.URL + "/token", "jwks_uri": host.URL + "/jwks", "revocation_endpoint": host.URL + "/revoke"})
		case "/revoke":
			w.WriteHeader(200)
		case "/jwks":
			json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]string{"kty": "RSA", "kid": "local", "alg": "RS256", "use": "sig", "n": enc(key.N.Bytes()), "e": enc(big.NewInt(int64(key.E)).Bytes())}}})
		case "/token":
			r.ParseForm()
			access := "plan-token"
			refresh := "refresh-one"
			if r.Form.Get("client_id") != "issued-client" || r.Form.Get("resource") != "https://api.openai.com/v1" {
				t.Error("wrong token identity/resource")
			}
			if r.Form.Get("grant_type") == "refresh_token" {
				refreshes.Add(1)
				if r.Form.Get("refresh_token") != "refresh-one" {
					t.Error("stale refresh")
				}
				access = "renewed-plan-token"
				refresh = "refresh-two"
			} else {
				sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				if enc(sum[:]) != challenge || r.Form.Get("redirect_uri") != redirect {
					t.Error("PKCE or redirect changed")
				}
			}
			header, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "local"})
			claims, _ := json.Marshal(map[string]any{"iss": host.URL, "aud": "issued-client", "sub": "synthetic-account", "nonce": nonce, "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()})
			signing := enc(header) + "." + enc(claims)
			digest := sha256.Sum256([]byte(signing))
			sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
			json.NewEncoder(w).Encode(map[string]any{"access_token": access, "refresh_token": refresh, "token_type": "Bearer", "expires_in": 3600, "scope": "openid profile offline_access resource.invoke chatgpt.tokens.use.direct", "id_token": signing + "." + enc(sig)})
		case "/models":
			if r.Header.Get("Authorization") != "Bearer plan-token" {
				t.Error("models bypassed credential owner")
			}
			fmt.Fprint(w, `{"models":[{"slug":"gpt-ch19-course","display_name":"Course model","visibility":"list"}]}`)
		case "/chosen/responses":
			requests.Add(1)
			body, _ := io.ReadAll(r.Body)
			auth := r.Header.Get("Authorization")
			if auth == "Bearer plan-token" {
				w.WriteHeader(401)
				return
			}
			var model struct {
				Model string `json:"model"`
			}
			if err := json.Unmarshal(body, &model); err != nil {
				t.Error(err)
			}
			checkCh16R2CacheEnvelope(t, body, auth == "Bearer dummy-api" && model.Model == "gpt-ch19-course")
			if auth == "Bearer renewed-plan-token" {
				if bytes.Contains(body, []byte("prompt_cache_breakpoint")) || !bytes.Contains(body, []byte(`"type":"namespace"`)) || !bytes.Contains(body, []byte(`"stream":true`)) || !bytes.Contains(body, []byte(`"store":false`)) {
					t.Errorf("plan envelope %s", body)
				}
			} else if auth == "Bearer dummy-api" {
				if (model.Model == "gpt-ch19-course" && !bytes.Contains(body, []byte("prompt_cache_breakpoint"))) || bytes.Contains(body, []byte(`"type":"namespace"`)) {
					t.Errorf("API envelope %s", body)
				}
			} else {
				t.Error("unknown credential")
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "data: {\"type\":\"response.completed\",\"response\":{\"status\":\"completed\",\"model\":\"gpt-ch19-course\",\"output\":[{\"type\":\"message\",\"role\":\"assistant\",\"content\":[{\"type\":\"output_text\",\"text\":\"Done\"}]}],\"usage\":{\"input_tokens\":10,\"output_tokens\":1}}}\n\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer host.Close()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	redirect = "http://" + l.Addr().String() + "/auth/callback"
	l.Close()
	directory := t.TempDir()
	if err := os.Chmod(directory, 0700); err != nil {
		t.Fatal(err)
	}
	options := ensemble.AuthOptions{Directory: directory, Issuer: host.URL, RedirectURI: redirect, ModelsURL: host.URL + "/models"}
	auth := ensemble.NewAuth(options)
	var command bytes.Buffer
	err = ensemble.RunAuthCommand(context.Background(), auth, []string{"login"}, &command, func(raw string) error {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}
		q := u.Query()
		nonce = q.Get("nonce")
		challenge = q.Get("code_challenge")
		if q.Get("client_id") != "dynamic_agent_client" || q.Get("redirect_uri") != redirect || !strings.HasPrefix(q.Get("ext_agent_host_id"), "urn:uuid:") {
			t.Errorf("registration fields: client=%q redirect=%q host=%q", q.Get("client_id"), q.Get("redirect_uri"), q.Get("ext_agent_host_id"))
			return fmt.Errorf("wrong client registration")
		}
		res, err := http.Get(redirect + "?" + url.Values{"state": {q.Get("state")}, "code": {"local-code"}, "iss": {host.URL}, "client_id": {"issued-client"}}.Encode())
		if err == nil {
			res.Body.Close()
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []string{"status", "models"} {
		if err = ensemble.RunAuthCommand(context.Background(), auth, []string{cmd}, &command, nil); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(command.String(), "billing=chatgpt") || !strings.Contains(command.String(), "Course model") || strings.Contains(command.String(), "plan-token") {
		t.Fatalf("unsafe/incomplete shared command output: %s", command.String())
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	app.SetAuth(ensemble.NewAuth(options))
	config := ensemble.Config{SystemPrompt: ch16R2Constitution, Vendor: ensemble.OpenAI, Model: "gpt-ch19-course", APIKey: "dummy-api", BaseURL: host.URL + "/chosen/responses", BaseDir: t.TempDir(), DataDir: t.TempDir(), Tools: []string{"read_file"}, DisableStreaming: true}
	if _, err := app.NewAgent(config); err == nil {
		t.Fatal("plan accepted contradictory stream-off")
	}
	config.DisableStreaming = false
	a, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	if a.Config().AuthMode != ensemble.AuthChatGPT {
		t.Fatal("persisted billing lost")
	}
	if _, err = a.Ask(context.Background(), "Explain the route"); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 || refreshes.Load() != 1 {
		t.Fatal("401 renewal was not bounded", requests.Load(), refreshes.Load())
	}
	config.AuthMode = ensemble.AuthAPI
	config.BaseDir = t.TempDir()
	config.DataDir = t.TempDir()
	b, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.Ask(context.Background(), "Explain the API route"); err != nil {
		t.Fatal(err)
	}
	if b.Engine().Status().Cache.MarkerCount == 0 {
		t.Fatal("API Responses markers not measured")
	}
	if a.Engine().Status().Cache.MarkerCount != 0 {
		t.Fatal("plan Responses markers measured")
	}
	config.Model = "gpt-ch19-course-unknown"
	config.Surface = ensemble.Responses
	config.BaseDir = t.TempDir()
	config.DataDir = t.TempDir()
	unsupported, err := app.NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = unsupported.Ask(context.Background(), "Keep the unsupported route conservative"); err != nil {
		t.Fatal(err)
	}
	if unsupported.Engine().Status().Cache.MarkerCount != 0 {
		t.Fatal("unsupported route acquired cache markers")
	}
	if err = ensemble.RunAuthCommand(context.Background(), app.Auth(), []string{"logout"}, &command, nil); err != nil {
		t.Fatal(err)
	}
	before := requests.Load()
	if _, err = a.Ask(context.Background(), "Do not fall back"); err == nil || requests.Load() != before {
		t.Fatal("logged out plan fell back to API")
	}
}
