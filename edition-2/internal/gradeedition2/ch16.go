package gradeedition2

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter16 preserves original19's seven weights while exercising the actual
// optional GUI command against isolated OAuth and Responses protocol peers.
func Chapter16(bin string) (grade.Report, error) {
	r := subscriptionChecks(bin)
	checks := grade.Ch19Checks(grade.Ch19Result{Errs: r.Errs, Fatal: r.Fatal})
	titles := map[string]string{
		"credential-provider": "Minted OAuth and API-key bearers, without secrets in observed output, saved conversation or request bodies",
		"oauth-flow":          "Successful PKCE/issued-client exchange, forged-signature and nonce rejection, and actual inference-grant refusal",
		"token-refresh":       "Proactive rotating renewal with successful turns and the corresponding new bearers",
		"responses-format":    "Responses fields, actual completed-item tool result and matching encrypted replay",
		"cache-breakpoints":   "Paired plan-route omission and explicit API-key developer-content breakpoints within the four-marker budget",
		"reasoning-summaries": "Separate complete summary/text deltas, including tool-adjacent content, with delivery before provider completion",
		"billing-mode":        "Actual admission and midstream quota refusals reported without switching to the configured metered key",
	}
	for i := range checks {
		checks[i].Title = titles[checks[i].ID]
	}
	return grade.NewTitledReport("Edition 2, Chapter 16 — Leaving Anthropic for OpenAI", checks, ""), nil
}

func subscriptionChecks(bin string) contextResult {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range grade.Ch19Checks(grade.Ch19Result{}) {
		r.set(c.ID, "")
	}
	dir, err := os.MkdirTemp("", "subscription-grade-")
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	defer os.RemoveAll(dir)
	skills, gui := filepath.Join(dir, "skills"), filepath.Join(dir, "web")
	err = recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Subscription protocol fixture\ntype: primary\ntools: read_file\n---\nAnswer briefly. Preserve subscription-constitution-741.\n"})
	if err != nil {
		r.Fatal = err.Error()
		return r
	}
	if err := contextStaticFixture(gui); err != nil {
		r.Fatal = err.Error()
		return r
	}
	run := func(name string, o subscriptionOpts) subscriptionOut {
		cwd := filepath.Join(dir, name)
		if err := os.MkdirAll(cwd, 0755); err != nil {
			return subscriptionOut{fatal: err.Error()}
		}
		return subscriptionLaunch(bin, skills, gui, cwd, o)
	}
	a := run("plan", subscriptionOpts{prompts: []string{"Explain the migration.", "Now read a file."}, gate: true, responses: grade.Edition2ResponsesOptions{Scenarios: []string{"summary_then_text", "summary_around_toolcall", "text"}, InputTokens: 1100, CachedTokens: 200, CacheWriteTokens: 100, OutputTokens: 50, ReasoningTokens: 17}})
	if subscriptionOK(&r, a, "credential-provider", "oauth-flow", "responses-format", "cache-breakpoints", "reasoning-summaries") {
		if len(a.oauth.AuthorizeHits) != 1 || len(a.oauth.TokenHits) != 1 || len(a.oauth.Violations) > 0 || a.oauth.IssuedClientID == "" {
			r.fail("oauth-flow", "positive PKCE/registration exchange invalid: %v", a.oauth.Violations)
		}
		if len(a.oauth.IssuedAccessTokens) != 1 {
			r.fail("credential-provider", "expected one minted positive-control bearer")
		} else {
			subscriptionBearers(&r, "credential-provider", a, "Bearer "+a.oauth.IssuedAccessTokens[0])
		}
		subscriptionSecrets(&r, a)
		subscriptionFormat(&r, a)
		subscriptionCache(&r, a, false)
		summary, text := "", ""
		for _, m := range a.messages {
			if m["type"] != "part_delta" {
				continue
			}
			chunk, _ := m["chunk"].(string)
			if m["kind"] == "reasoning_summary" {
				summary += chunk
			}
			if m["kind"] == "text" {
				text += chunk
			}
		}
		wantA := "Reading the request, then deciding what to do."
		wantB := "Got the file. Now checking what it says."
		if summary != wantA+wantA+wantB || text != strings.Repeat("Hello from the Responses API.", 3) || !a.gated {
			r.fail("reasoning-summaries", "separate complete delta content or causal gate missing (summary=%q text=%q gated=%t)", summary, text, a.gated)
		}
	}
	e := run("api", subscriptionOpts{skipLogin: true, prompts: []string{"No grant stored."}, responses: grade.Edition2ResponsesOptions{Scenarios: []string{"text"}}})
	if subscriptionOK(&r, e, "credential-provider", "cache-breakpoints") {
		subscriptionBearers(&r, "credential-provider", e, "Bearer "+grade.Edition2APIKey)
		subscriptionSecrets(&r, e)
		subscriptionCache(&r, e, true)
	}
	// Refusal must follow the actual malicious token exchange. A startup failure
	// or an issuer never reached cannot prove that identity was validated.
	for _, c := range []struct {
		name    string
		options grade.Edition2OAuthOptions
	}{{"bad-signature", grade.Edition2OAuthOptions{UnsignedKeyMismatch: true}}, {"bad-nonce", grade.Edition2OAuthOptions{WrongNonce: true}}} {
		out := run(c.name, subscriptionOpts{oauth: c.options, prompts: []string{"Must not infer."}})
		if len(out.oauth.TokenHits) != 1 || len(out.oauth.IssuedAccessTokens) != 1 || out.loginErr == "" || out.fatal != "" || len(out.responses.Requests) != 0 {
			r.fail("oauth-flow", "%s: did not reject the issued invalid identity at login (login=%q startup=%q requests=%d)", c.name, out.loginErr, out.fatal, len(out.responses.Requests))
		}
	}
	grant := run("no-grant", subscriptionOpts{oauth: grade.Edition2OAuthOptions{OmitPlanScope: true}, prompts: []string{"Identity is not inference permission."}})
	// A valid identity may stay signed in with plan use disabled. Requiring a
	// failed login here would invent a stricter identity/storage contract.
	if len(grant.oauth.TokenHits) != 1 || len(grant.oauth.IssuedAccessTokens) != 1 || len(grant.responses.Requests) != 0 || grant.fatal != "" || (grant.loginErr == "" && (len(grant.errors) != 1 || grant.errors[0] == "")) {
		r.fail("oauth-flow", "missing grant did not refuse inference after an actual token exchange (login=%q startup=%q requests=%d errors=%v)", grant.loginErr, grant.fatal, len(grant.responses.Requests), grant.errors)
	}
	b := run("refresh", subscriptionOpts{oauth: grade.Edition2OAuthOptions{AccessTokenTTL: 90 * time.Second}, prompts: []string{"First renewal.", "Second renewal."}, responses: grade.Edition2ResponsesOptions{Scenarios: []string{"text"}}})
	if subscriptionOK(&r, b, "token-refresh") {
		obs := b.oauth
		if len(obs.RefreshHits) != 2 || len(obs.IssuedAccessTokens) != 3 || len(b.responses.Requests) != 2 || obs.RefreshReuseDetected || len(obs.Violations) > 0 {
			r.fail("token-refresh", "two proactive rotations and successful requests missing: refresh=%d minted=%d requests=%d reuse=%t violations=%v", len(obs.RefreshHits), len(obs.IssuedAccessTokens), len(b.responses.Requests), obs.RefreshReuseDetected, obs.Violations)
		} else {
			// Every token has a 90-second TTL. Each request therefore needs its own
			// renewal under the chapter's five-minute window, in this exact order.
			for i, q := range b.responses.Requests {
				if q.Auth != "Bearer "+obs.IssuedAccessTokens[i+1] || obs.RefreshHits[i].RefreshToken != obs.IssuedRefreshTokens[i] {
					r.fail("token-refresh", "request %d did not use its corresponding rotated pair", i+1)
				}
			}
		}
	}
	for _, scenario := range []string{"text", "fail_midstream"} {
		out := run("refusal-"+scenario, subscriptionOpts{prompts: []string{"Refuse this plan request."}, responses: grade.Edition2ResponsesOptions{Scenarios: []string{scenario}, FailWith: "subscription_sharing_usage_limit_exceeded"}})
		if out.loginErr != "" || out.fatal != "" || len(out.oauth.IssuedAccessTokens) != 1 || len(out.responses.Requests) == 0 || len(out.errors) != 1 || out.errors[0] == "" {
			r.fail("billing-mode", "%s: no actual authenticated refusal (login=%q startup=%q requests=%d errors=%v)", scenario, out.loginErr, out.fatal, len(out.responses.Requests), out.errors)
			continue
		}
		subscriptionBearers(&r, "billing-mode", out, "Bearer "+out.oauth.IssuedAccessTokens[0])
	}
	return r
}
func subscriptionOK(r *contextResult, out subscriptionOut, ids ...string) bool {
	problem := out.fatal
	if problem == "" && out.exitErr != "" {
		problem = "nonzero successful-session exit: " + out.exitErr
	}
	if problem == "" && out.loginErr != "" {
		problem = "login: " + out.loginErr
	}
	if problem == "" {
		for _, e := range out.errors {
			if e != "" {
				problem = "turn: " + e
				break
			}
		}
	}
	if problem == "" && len(out.responses.Requests) == 0 {
		problem = "no inference request"
	}
	if problem != "" {
		for _, id := range ids {
			r.fail(id, "%s", problem)
		}
		return false
	}
	return true
}
func subscriptionBearers(r *contextResult, id string, out subscriptionOut, want string) {
	for i, q := range out.responses.Requests {
		if q.Auth != want {
			r.fail(id, "request %d used the wrong credential", i+1)
		}
	}
}
func subscriptionSecrets(r *contextResult, out subscriptionOut) {
	secrets := append([]string{grade.Edition2APIKey}, out.oauth.IssuedAccessTokens...)
	secrets = append(secrets, out.oauth.IssuedRefreshTokens...)
	hay := out.loginOutput + out.output + out.artifacts
	for _, q := range out.responses.Requests {
		hay += string(q.Body)
	}
	for _, s := range secrets {
		if s != "" && strings.Contains(hay, s) {
			r.fail("credential-provider", "a fake credential leaked outside the protected store/authorization header")
			return
		}
	}
}
func subscriptionFormat(r *contextResult, out subscriptionOut) {
	const id = "responses-format"
	if len(out.responses.Violations) > 0 {
		r.fail(id, "provider wire violations: %v", out.responses.Violations)
	}
	if len(out.responses.Requests) != 3 || len(out.errors) != 2 {
		r.fail(id, "two turns with one real tool require three requests; got %d", len(out.responses.Requests))
		return
	}
	for _, q := range out.responses.Requests {
		if !q.IncludesEncryptedReasoning || q.ReasoningSummary == "" {
			r.fail(id, "encrypted replay inclusion or summary request missing")
		}
	}
	// Independently read the existing public fixture path. A correlated error or
	// empty success is not successful tool use, and echoed arguments cannot pass.
	file, err := os.ReadFile("/etc/hosts")
	if err != nil || len(file) == 0 {
		r.fail(id, "cannot establish /etc/hosts fixture: %v", err)
		return
	}
	var call, result, opaque bool
	input, _ := out.responses.Requests[2].Decoded["input"].([]any)
	for _, item := range input {
		m, _ := item.(map[string]any)
		if m["type"] == "reasoning" && m["encrypted_content"] == "ch19-encrypted-reasoning-blob" {
			opaque = true
		}
		if m["call_id"] != "call_ch19_c1" {
			continue
		}
		if m["type"] == "function_call" && m["name"] == "read_file" {
			call = true
		}
		if m["type"] == "function_call_output" {
			s, _ := m["output"].(string)
			result = strings.Contains(s, string(file))
		}
	}
	if !call || !result || !opaque {
		r.fail(id, "following request lacks call, actual successful file bytes or encrypted item (call=%t result=%t opaque=%t)", call, result, opaque)
	}
}
func subscriptionCache(r *contextResult, out subscriptionOut, metered bool) {
	for i, q := range out.responses.Requests {
		_, options := q.Decoded["prompt_cache_options"]
		if !metered {
			if options || q.BreakpointCount != 0 {
				r.fail("cache-breakpoints", "plan request %d did not omit paired explicit-cache mechanism", i+1)
			}
			continue
		}
		landed := false
		in, _ := q.Decoded["input"].([]any)
		for _, item := range in {
			m, _ := item.(map[string]any)
			if m["role"] != "developer" {
				continue
			}
			blocks, _ := m["content"].([]any)
			for _, b := range blocks {
				block, _ := b.(map[string]any)
				text, _ := block["text"].(string)
				marker, _ := block["prompt_cache_breakpoint"].(map[string]any)
				if strings.Contains(text, "subscription-constitution-741") && marker["mode"] == "explicit" {
					landed = true
				}
			}
		}
		if q.PromptCacheMode != "explicit" || q.BreakpointCount < 1 || q.BreakpointCount > 4 || !landed {
			r.fail("cache-breakpoints", "metered request %d lacks actual developer-content marker/mode or violates the four-marker budget", i+1)
		}
	}
}
