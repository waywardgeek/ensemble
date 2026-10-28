package gradeedition2

// This scope exercises the real C# GUI, not its internal test assembly. Grants
// are deliberately seeded: these checks cannot establish login, signed identity,
// refresh, or provider acceptance. The original Go OAuth route remains separate.
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

func Chapter16Executable(bin string) (grade.Report, error) {
	r, dir := subscriptionExecutableChecks(bin)
	checks := subscriptionExecutableChecklist(r)
	return grade.NewTitledReport("Edition 3 C#, Chapter 16 — local executable transport scope (70 points; OAuth/refresh excluded)", checks, "Synthetic-grant fixtures retained at "+dir+". Signed login/refresh require separate evidence; live OAuth remains incomplete."), nil
}

func subscriptionExecutableChecklist(r contextResult) []grade.Check {
	titles := map[string]string{
		"credential-provider": "Selected synthetic bearer without secret leakage into model bodies, GUI output or conversation files",
		"responses-format":    "Current plan Responses fields, completed tool result and encrypted replay",
		"cache-breakpoints":   "Documented API cache-routing key and default plan omission, without invented marker fields",
		"reasoning-summaries": "Separate summaries with tool-adjacent suffixes and causal incremental delivery",
		"billing-mode":        "Admission, quota and signed-out refusals without API-key fallback",
	}
	var checks []grade.Check
	for _, c := range grade.Ch19Checks(grade.Ch19Result{Errs: r.Errs, Fatal: r.Fatal}) {
		title, included := titles[c.ID]
		if !included {
			continue
		}
		c.Title = title
		checks = append(checks, c)
	}
	return checks
}

func subscriptionExecutableChecks(bin string) (contextResult, string) {
	r := contextResult{Errs: map[string]string{}}
	for _, c := range subscriptionExecutableChecklist(r) {
		r.set(c.ID, "")
	}
	dir, err := os.MkdirTemp("", "subscription-csharp-grade-")
	if err != nil {
		r.Fatal = err.Error()
		return r, dir
	}
	skills := filepath.Join(dir, "skills")
	if err := recallWrite(skills, map[string]string{"base/SKILL.md": "---\nname: base\ndescription: Local Responses fixture\ntype: primary\ntools: read_file\n---\nAnswer briefly. Preserve subscription-constitution-741.\n"}); err != nil {
		r.Fatal = err.Error()
		return r, dir
	}
	run := func(name string, o subscriptionOpts) subscriptionOut {
		cwd := filepath.Join(dir, name)
		if err := os.MkdirAll(cwd, 0700); err != nil {
			return subscriptionOut{fatal: err.Error()}
		}
		o.executableOnly = true
		out := subscriptionLaunch(bin, skills, "", cwd, o)
		// Retain private fixture evidence even when an assertion fails. These
		// requests contain synthetic grants and never enter the repository.
		raw, err := json.MarshalIndent(map[string]any{
			"responses": out.responses, "messages": out.messages, "errors": out.errors,
			"fatal": out.fatal, "exit_error": out.exitErr, "output": out.output, "gated": out.gated,
		}, "", "  ")
		if err == nil {
			err = os.WriteFile(filepath.Join(cwd, "observations.json"), raw, 0600)
		}
		if err != nil {
			out.fatal = "Could not preserve executable fixture observations: " + err.Error()
		}
		return out
	}
	a := run("plan", subscriptionOpts{prompts: []string{"Explain the migration.", "Now read a file."}, gate: true, responses: grade.Edition2ResponsesOptions{Scenarios: []string{"summary_then_text", "explained_summary_toolcall", "text"}}})
	if subscriptionOK(&r, a, "credential-provider", "responses-format", "cache-breakpoints", "reasoning-summaries") {
		subscriptionBearers(&r, "credential-provider", a, "Bearer "+subscriptionFixtureAccess)
		subscriptionExecutableSecrets(&r, a)
		subscriptionExecutableFormat(&r, a)
		subscriptionExecutableSummaries(&r, a)
		subscriptionExecutableCache(&r, a, false)
	}
	api := run("api", subscriptionOpts{skipLogin: true, prompts: []string{"Use the explicitly selected API route."}, responses: grade.Edition2ResponsesOptions{Scenarios: []string{"text"}}})
	if subscriptionOK(&r, api, "credential-provider", "cache-breakpoints") {
		subscriptionBearers(&r, "credential-provider", api, "Bearer "+grade.Edition2APIKey)
		subscriptionExecutableSecrets(&r, api)
		if len(api.responses.Requests) != 1 {
			r.fail("cache-breakpoints", "API request count %d, want one", len(api.responses.Requests))
		}
		subscriptionExecutableCache(&r, api, true)
	}
	for _, code := range []string{"subscription_sharing_usage_limit_exceeded", "subscription_sharing_usage_unavailable"} {
		for _, mid := range []bool{false, true} {
			opts := grade.Edition2ResponsesOptions{FailWith: code}
			suffix := "-admission"
			if mid {
				opts.Scenarios = []string{"fail_midstream"}
				suffix = "-stream"
			}
			b := run(code+suffix, subscriptionOpts{prompts: []string{"This request must fail without payment fallback."}, responses: opts})
			if subscriptionExecutableRefusal(&r, b) {
				if len(b.responses.Requests) != 1 {
					r.fail("billing-mode", "%s%s made %d requests", code, suffix, len(b.responses.Requests))
				}
				subscriptionBearers(&r, "billing-mode", b, "Bearer "+subscriptionFixtureAccess)

				subscriptionExecutableSecrets(&r, b)
			}
		}
	}
	signedOut := run("signed-out", subscriptionOpts{signedOut: true, prompts: []string{"Do not spend the ambient API key."}})
	if subscriptionExecutableRefusal(&r, signedOut) {
		if len(signedOut.responses.Requests) != 0 || !strings.Contains(strings.Join(signedOut.errors, "\n"), "No API-key fallback") {
			r.fail("billing-mode", "signed-out plan did not refuse before HTTP without fallback")
		}
		subscriptionExecutableSecrets(&r, signedOut)
	}
	return r, dir
}

// Do not inherit the old fixture's ban on instructions: the current plan
// documentation permits instructions or developer messages, but not system items.
func subscriptionExecutableFormat(r *contextResult, a subscriptionOut) {
	if len(a.responses.Requests) != 3 {
		r.fail("responses-format", "got %d requests, want two turns plus tool follow-up", len(a.responses.Requests))
	}
	for _, req := range a.responses.Requests {
		var body map[string]any
		if json.Unmarshal(req.Body, &body) != nil {
			r.fail("responses-format", "invalid request JSON")
			continue
		}
		reasoning, _ := body["reasoning"].(map[string]any)
		if body["model"] != "gpt-6-astra" || req.StreamValue != true || req.StoreValue != false || !req.IncludesEncryptedReasoning || reasoning["effort"] != "medium" || reasoning["summary"] != "auto" {
			r.fail("responses-format", "missing current stream/store/reasoning/encrypted fields")
		}
		for _, field := range []string{"background", "conversation", "max_output_tokens", "max_tool_calls", "metadata", "moderation", "multi_agent", "prompt", "prompt_cache_retention", "safety_identifier", "temperature", "top_logprobs", "top_p", "truncation", "user", "previous_response_id"} {
			if _, present := body[field]; present {
				r.fail("responses-format", "plan included unsupported %s", field)
			}
		}
		if !strings.Contains(subscriptionString(body["instructions"]), "subscription-constitution-741") {
			r.fail("responses-format", "primary instruction missing")
		}
		tools, _ := body["tools"].([]any)
		foundRead := false
		for _, item := range tools {
			namespace, _ := item.(map[string]any)
			if namespace["type"] != "namespace" || namespace["name"] != "ensemble" {
				r.fail("responses-format", "tool outside Ensemble namespace")
			}
			nested, _ := namespace["tools"].([]any)
			for _, t := range nested {
				tool, _ := t.(map[string]any)
				if tool["name"] == "read_file" {
					foundRead = true
				}
			}
		}
		if !foundRead {
			r.fail("responses-format", "read_file declaration missing")
		}
		input, ok := body["input"].([]any)
		if !ok {
			r.fail("responses-format", "input is not an array")
		}
		for _, item := range input {
			m, _ := item.(map[string]any)
			if m["role"] == "system" {
				r.fail("responses-format", "explicit system input item")
			}
		}
	}
	if len(a.responses.Requests) < 3 {
		return
	}
	req := a.responses.Requests[2]
	var body map[string]any
	_ = json.Unmarshal(req.Body, &body)
	input, _ := body["input"].([]any)
	hosts, err := os.ReadFile("/etc/hosts")
	if err != nil {
		r.fail("responses-format", "read expected file: %v", err)
		return
	}
	call, output, opaque := false, false, false
	for _, item := range input {
		m, _ := item.(map[string]any)
		if m["type"] == "function_call" && m["call_id"] == "call_ch19_c1" && m["name"] == "read_file" && m["namespace"] == "ensemble" {
			call = true
		}
		if m["type"] == "function_call_output" && m["call_id"] == "call_ch19_c1" {
			output = strings.Contains(subscriptionString(m["output"]), strings.TrimSpace(string(hosts)))
		}
		if m["type"] == "reasoning" && m["encrypted_content"] == "ch19-encrypted-reasoning-blob" {
			opaque = true
		}
	}
	if !call || !output || !opaque {
		r.fail("responses-format", "tool/replay missing: call=%v actual_file=%v encrypted=%v", call, output, opaque)
	}
}

func subscriptionExecutableCache(r *contextResult, a subscriptionOut, api bool) {
	for _, req := range a.responses.Requests {
		var body map[string]any
		if json.Unmarshal(req.Body, &body) != nil {
			r.fail("cache-breakpoints", "invalid JSON")
			continue
		}
		key, present := body["prompt_cache_key"]
		if api && key != "chapter16-cache-routing" || !api && present {
			r.fail("cache-breakpoints", "wrong explicit API/default plan cache routing")
		}
		if body["prompt_cache_options"] != nil || req.BreakpointCount != 0 {
			r.fail("cache-breakpoints", "historical explicit cache fields emitted")
		}
	}
}

// This is fixture construction, not an authentication flow. Long-lived synthetic
// tokens deliberately avoid the fixed production issuer; none are real secrets.
func subscriptionSeed(home string, signedOut bool) error {
	dir := filepath.Join(home, "credentials")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	account := map[string]any{"label": "grade", "issuer": "https://auth.openai.com", "subject": "fixture-subject", "client_id": "fixture-client", "email": "fixture@example.invalid", "access_token": subscriptionFixtureAccess, "refresh_token": subscriptionFixtureRefresh, "id_token": subscriptionFixtureID, "expires_at": "2099-01-01T00:00:00.0000000+00:00", "scope": "openid resource.invoke chatgpt.tokens.use.direct"}
	if signedOut {
		for _, key := range []string{"access_token", "refresh_token", "id_token", "expires_at"} {
			account[key] = nil
		}
		account["scope"] = nil
	}
	raw, err := json.Marshal(map[string]any{"version": 2, "host_id": "urn:uuid:00000000-0000-4000-8000-000000000001", "selected_plan_account": "grade", "accounts": []any{account}})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "credentials.json"), raw, 0600)
}

func subscriptionExecutableEnv(url, home, skills, cwd string, api bool) []string {
	var env []string
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(key, "LLM_") || strings.HasPrefix(key, "EN_") || strings.HasPrefix(key, "ENSEMBLE_") || strings.HasPrefix(key, "OPENAI_") || strings.HasPrefix(key, "ANTHROPIC_") || strings.HasPrefix(key, "GEMINI_") || strings.HasPrefix(key, "GOOGLE_") || key == "CH02_LOG" || key == "CH06_API_LOG" || key == "HOME" || key == "XDG_CONFIG_HOME" {
			continue
		}
		env = append(env, item)
	}
	billing := "chatgpt"
	if api {
		billing = "api"
	}
	env = append(env, "HOME="+home, "XDG_CONFIG_HOME="+filepath.Join(home, "config"), "LLM_VENDOR=openai", "LLM_MODEL=gpt-6-astra", "EN_SURFACE=responses", "EN_BILLING="+billing, "LLM_BASE_URL="+url, "LLM_API_KEY="+grade.Edition2APIKey, "EN_SKILLS_DIR="+skills, "EN_PRIMARY_SKILL=base", "EN_SAVE_PATH="+filepath.Join(cwd, "save.json"), "LLM_LOG_PATH="+filepath.Join(cwd, "events.jsonl"), "OPENAI_REASONING_EFFORT=medium", "OPENAI_REASONING_SUMMARY=auto")
	if api {
		env = append(env, "EN_PROMPT_CACHE_KEY=chapter16-cache-routing")
	}
	return env
}

const subscriptionFixtureAccess = "local-csharp-access-canary-741"
const subscriptionFixtureRefresh = "local-csharp-refresh-canary-741"
const subscriptionFixtureID = "local-csharp-identity-canary-741"

func subscriptionExecutableSecrets(r *contextResult, out subscriptionOut) {
	subscriptionSecrets(r, out)
	socket, err := json.Marshal(out.messages)
	if err != nil {
		r.fail("credential-provider", "could not inspect GUI socket messages: %v", err)
	}
	if strings.Contains(string(socket), "ch19-encrypted-reasoning-blob") {
		r.fail("credential-provider", "encrypted replay material leaked into GUI projection")
	}
	hay := out.loginOutput + out.output + out.artifacts + string(socket)
	for _, q := range out.responses.Requests {
		hay += string(q.Body)
	}
	for _, secret := range []string{subscriptionFixtureAccess, subscriptionFixtureRefresh, subscriptionFixtureID} {
		if strings.Contains(hay, secret) {
			r.fail("credential-provider", "synthetic grant leaked outside its protected store/header")
		}
	}
}

func subscriptionExecutableRefusal(r *contextResult, out subscriptionOut) bool {
	if out.fatal != "" || out.exitErr != "" || out.loginErr != "" || len(out.errors) != 1 || out.errors[0] == "" {
		r.fail("billing-mode", "expected one reported refusal and clean host exit: fatal=%q exit=%q login=%q errors=%v", out.fatal, out.exitErr, out.loginErr, out.errors)
		return false
	}
	return true
}

func subscriptionExecutableError(messages []map[string]any) string {
	for _, m := range messages {
		if m["type"] == "lifecycle" && m["kind"] == "error_occurred" {
			return subscriptionString(m["message"])
		}
	}
	return subscriptionTurnError(messages)
}

func subscriptionExecutableSummaries(r *contextResult, out subscriptionOut) {
	summary, text := "", ""
	for _, m := range out.messages {
		if m["type"] != "part_delta" {
			continue
		}
		if m["kind"] == "reasoning_summary" {
			summary += subscriptionString(m["chunk"])
		}
		if m["kind"] == "text" {
			text += subscriptionString(m["chunk"])
		}
	}
	a := "Reading the request, then deciding what to do."
	b := "Got the file. Now checking what it says."
	if summary != a+a+b || text != strings.Repeat("Hello from the Responses API.", 3) || !out.gated {
		r.fail("reasoning-summaries", "separate full summaries/text or causal gate missing: summary=%q text=%q gated=%t", summary, text, out.gated)
	}
}

func subscriptionString(value any) string { text, _ := value.(string); return text }
