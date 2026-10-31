package gradeedition2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChapter16ExecutableScope(t *testing.T) {
	r := contextResult{Errs: map[string]string{}}
	ids := []string{"credential-provider", "responses-format", "cache-breakpoints", "reasoning-summaries", "billing-mode"}
	for _, id := range ids {
		r.set(id, "")
	}
	checks := subscriptionExecutableChecklist(r)
	points := 0
	for _, c := range checks {
		if !c.Passed || c.ID == "oauth-flow" || c.ID == "token-refresh" {
			t.Fatalf("wrong inclusion: %+v", c)
		}
		points += c.Points
	}
	if len(checks) != 5 || points != 70 {
		t.Fatalf("scope: %d checks, %d points", len(checks), points)
	}
	r.fail("responses-format", "intentional invalid request")
	for _, c := range subscriptionExecutableChecklist(r) {
		if c.ID == "responses-format" && (c.Passed || c.Earned != 0) {
			t.Fatal("wire failure credited")
		}
	}
}

func TestChapter16ExecutableSeed(t *testing.T) {
	for _, signedOut := range []bool{false, true} {
		home := t.TempDir()
		if err := subscriptionSeed(home, signedOut); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(home, "credentials", "credentials.json")
		stat, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if stat.Mode().Perm() != 0600 {
			t.Fatalf("mode %v", stat.Mode())
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var store map[string]any
		if err = json.Unmarshal(raw, &store); err != nil {
			t.Fatal(err)
		}
		accounts := store["accounts"].([]any)
		account := accounts[0].(map[string]any)
		if store["version"] != float64(2) || store["selected_plan_account"] != "grade" || len(accounts) != 1 {
			t.Fatal("wrong saved selection")
		}
		if signedOut {
			if account["access_token"] != nil || account["refresh_token"] != nil || account["expires_at"] != nil {
				t.Fatal("signed-out grant retained")
			}
		} else if account["access_token"] != subscriptionFixtureAccess || account["refresh_token"] != subscriptionFixtureRefresh {
			t.Fatal("wrong grant")
		}
	}
}

func TestChapter16ExecutableEnvironment(t *testing.T) {
	t.Setenv("EN_BILLING", "foreign")
	t.Setenv("OPENAI_OIDC_ISSUER", "https://foreign.invalid")
	t.Setenv("OPENAI_API_KEY", "foreign-secret")
	t.Setenv("CH06_API_LOG", "/foreign")
	for _, api := range []bool{false, true} {
		env := subscriptionExecutableEnv("http://127.0.0.1:1", "/owned/home", "/owned/skills", "/owned/work", api)
		got := map[string]string{}
		for _, item := range env {
			key, value, _ := strings.Cut(item, "=")
			if _, exists := got[key]; exists {
				t.Fatalf("duplicate %s", key)
			}
			got[key] = value
			if strings.Contains(value, "foreign") {
				t.Fatalf("inherited foreign setting %s", key)
			}
		}
		want := "chatgpt"
		if api {
			want = "api"
		}
		if got["EN_BILLING"] != want || got["EN_SURFACE"] != "responses" || got["LLM_MODEL"] != "gpt-6-astra" || got["OPENAI_REASONING_SUMMARY"] != "auto" {
			t.Fatal("wrong launch selection")
		}
		if _, exists := got["OPENAI_OIDC_ISSUER"]; exists {
			t.Fatal("issuer override passed to C#")
		}
		if (got["EN_PROMPT_CACHE_KEY"] != "") == !api {
			t.Fatal("wrong cache routing option")
		}
	}
}

func TestChapter16ExecutableRefusalAndLeakage(t *testing.T) {
	r := contextResult{Errs: map[string]string{}}
	good := subscriptionOut{errors: []string{"safe refusal"}}
	if !subscriptionExecutableRefusal(&r, good) {
		t.Fatal("clean reported refusal rejected")
	}
	bad := good
	bad.exitErr = "signal"
	if subscriptionExecutableRefusal(&r, bad) {
		t.Fatal("failed exit credited")
	}
	events := []map[string]any{{"type": "lifecycle", "kind": "error_occurred", "message": "safe refusal"}, {"type": "turn_ended", "error": "generic"}}
	if subscriptionExecutableError(events) != "safe refusal" {
		t.Fatal("lost durable error")
	}
	r = contextResult{Errs: map[string]string{}}
	subscriptionExecutableSecrets(&r, subscriptionOut{messages: []map[string]any{{"text": subscriptionFixtureID}}})
	if r.Errs["credential-provider"] == "" {
		t.Fatal("identity token leak missed")
	}
}
