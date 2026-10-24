package ensemble_test

import (
	"bytes"
	"context"
	"encoding/json"
	"ensemble"
	"ensemble/internal/common"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type ch19Auth struct {
	common.Auth
	token string
}

func (a ch19Auth) Resolve(context.Context, common.Config) (common.Credential, error) {
	return common.Credential{Mode: common.AuthAPI, Token: a.token, Profile: "fixture-profile"}, nil
}

type ch19Trace chan []byte

func (w ch19Trace) Write(p []byte) (int, error) { w <- append([]byte(nil), p...); return len(p), nil }
func TestCh19ResolvedCredentialDiagnostics(t *testing.T) {
	secret := "profile-\"<&\\token"
	encoded, _ := json.Marshal(secret)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+secret {
			t.Error("missing resolved credential")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1},"echo":%s}`, encoded)
	}))
	defer server.Close()
	app := ensemble.New(nil)
	defer app.Shutdown()
	if err := app.SetAuth(ch19Auth{token: secret}); err != nil {
		t.Fatal(err)
	}
	trace := make(ch19Trace, 32)
	app.SetTrace(trace)
	a, err := app.NewAgent(ensemble.Config{Vendor: ensemble.OpenAI, Surface: ensemble.ChatCompletions, AuthMode: ensemble.AuthAPI, Model: "fixture", BaseURL: server.URL, BaseDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	var diagnostic bytes.Buffer
	_, err = a.Engine().ExchangeRequest(context.Background(), []byte(`{"model":"fixture","messages":[]}`), common.ParseCallbacks{OnFrame: func(f common.Frame) { diagnostic.WriteString(f.Data) }})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(diagnostic.String(), string(encoded[1:len(encoded)-1])) || strings.Contains(diagnostic.String(), secret) {
		t.Fatal("resolved credential leaked in frame diagnostic")
	}
	deadline := time.After(time.Second)
	for {
		select {
		case line := <-trace:
			var record common.TraceRecord
			if err = json.Unmarshal(line, &record); err != nil {
				t.Fatal(err)
			}
			if record.Stage == "response_frame" {
				if strings.Contains(record.Payload, string(encoded[1:len(encoded)-1])) || strings.Contains(record.Payload, secret) {
					t.Fatal("resolved credential leaked in trace")
				}
				if !strings.Contains(record.Payload, "REDACTED") {
					t.Fatal("redaction marker missing")
				}
				return
			}
		case <-deadline:
			t.Fatal("response trace absent")
		}
	}
}
func TestCh19WebPolicyTransition(t *testing.T) {
	var initialized atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&m)
		if len(m.ID) == 0 {
			w.WriteHeader(202)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		result := `{}`
		switch m.Method {
		case "initialize":
			initialized.Add(1)
			result = `{"protocolVersion":"2025-03-26","capabilities":{"tools":{}}}`
		case "tools/list":
			result = `{"tools":[{"name":"lookup","description":"fixture lookup","inputSchema":{"type":"object"}}]}`
		case "tools/call":
			result = `{"content":[{"type":"text","text":"fixture"}]}`
		}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":%s}`, m.ID, result)
	}))
	defer server.Close()
	dir := t.TempDir()
	skills := filepath.Join(dir, "skills")
	for name, doc := range map[string]string{"primary": "---\nname: primary\ntype: primary\ndescription: fixture\nloadable-skills: web-search\n---\nFixture", "web": "---\nname: web-search\ntype: loadable\ndescription: fixture web\ntools: lookup\nmcp_servers:\n  - name: web-fixture\n    transport: url\n    url: " + server.URL + "\n---\nFixture"} {
		path := filepath.Join(skills, name, "SKILL.md")
		os.MkdirAll(filepath.Dir(path), 0700)
		if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	a, err := app.NewAgent(ensemble.Config{Model: "offline", BaseDir: dir, SandboxRoot: dir, SafeMode: true, EnableWebSearch: false, Skills: &ensemble.SkillSpec{Directory: skills, Primary: "primary"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Skills().LoadSkill("web-search"); err == nil {
		t.Fatal("disabled web loaded")
	}
	a.SetSecurityPolicy(true, true)
	if err = a.Skills().LoadSkill("web-search"); err != nil {
		t.Fatal(err)
	}
	if initialized.Load() != 1 {
		t.Fatal("web not connected")
	}
	if text, failed := c3Call(t, a, "lookup", map[string]any{}); failed {
		t.Fatal(text)
	}
	a.SetSecurityPolicy(true, false)
	for _, tool := range a.Skills().Snapshot().Definitions {
		if tool.Name == "lookup" {
			t.Fatal("disabled web declaration retained")
		}
	}
	if _, failed := c3Call(t, a, "lookup", map[string]any{}); !failed {
		t.Fatal("disabled web callable")
	}
	a.SetSecurityPolicy(true, true)
	if err = a.Skills().LoadSkill("web-search"); err != nil {
		t.Fatal(err)
	}
	if initialized.Load() != 2 {
		t.Fatal("web connection was not replaced after disable/reload")
	}
	if !a.Config().SafeMode {
		t.Fatal("safe policy lost")
	}
}
