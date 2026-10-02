package grade

// Chapter 19 runner: four scenarios, seven checks.
//
// Scenario A is the happy path: sign in with ChatGPT, then two turns of
// inference on the Responses API, the second of which calls a tool. It
// carries five of the seven checks.
//
// Scenario B expires the access token between turns, so the agent must
// refresh without the operator noticing.
//
// Scenario C tells the inference server the ChatGPT plan is exhausted. The
// agent must stop rather than quietly re-bill the work to a metered key.
//
// Scenario D signs the ID token with a key that is not in the published JWKS.
// Signature verification is not observable from the server side: the only way
// to tell a client that verifies from one that decodes is to hand it a
// forgery and see whether it proceeds.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// A grader-only model row, added alongside the real one so earlier chapters
// keep the exact wire they were written against.
const ch19Model = "gpt-ch19-course"

func Ch19Run(dir string) Ch19Result {
	var res Ch19Result

	bin, cleanup, err := Build(dir)
	if err != nil {
		res.Fatal = fmt.Sprintf("build failed: %v", err)
		return res
	}
	defer cleanup()

	skills, skillsCleanup := ch19Skills()
	defer skillsCleanup()
	gui := filepath.Join(dir, "web")

	ch19ScenarioA(&res, bin, skills, gui, dir)
	ch19ScenarioB(&res, bin, skills, gui, dir)
	ch19ScenarioC(&res, bin, skills, gui, dir)
	ch19ScenarioD(&res, bin, skills, gui, dir)
	ch19ScenarioE(&res, bin, skills, gui, dir)
	return res
}

// ---------------------------------------------------------------------------
// Scenario A: sign in, then two turns on the Responses API.
// ---------------------------------------------------------------------------

var ch19ScenarioAChecks = []string{
	"oauth-flow", "responses-format", "reasoning-summaries",
	"cache-breakpoints", "credential-provider",
}

func ch19ScenarioA(res *Ch19Result, bin, skills, gui, dir string) {
	for _, id := range ch19ScenarioAChecks {
		res.ran(id)
	}

	out := ch19Launch(bin, skills, gui, ch19Opts{
		dir:     dir,
		model:   ch19Model,
		prompts: []string{"Explain the migration.", "Now read a file."},
		resp: ch19RespOptions{
			Scenarios: []string{"summary_then_text", "summary_around_toolcall"},
		},
	})
	if out.fatal != "" {
		for _, id := range ch19ScenarioAChecks {
			res.fail(id, "scenario A: %s", out.fatal)
		}
		return
	}

	ch19CheckOAuthFlow(res, out)
	ch19CheckResponsesFormat(res, out)
	ch19CheckReasoningSummaries(res, out)
	ch19CheckCacheBreakpoints(res, out)
	ch19CheckCredentialProvider(res, out)
}

// ch19CheckOAuthFlow asserts a real authorization-code + PKCE exchange
// happened, and that the agent used the client_id the authorization response
// issued rather than the bootstrap placeholder it started with.
func ch19CheckOAuthFlow(res *Ch19Result, out ch19Out) {
	const id = "oauth-flow"
	o := out.oauth

	if len(o.Violations) > 0 {
		res.fail(id, "the authorization server rejected the agent's behaviour: %s", o.Violations[0])
		return
	}
	if len(o.AuthorizeHits) == 0 {
		res.fail(id, "no request ever reached the authorize endpoint: the agent did not start a sign-in")
		return
	}
	if len(o.TokenHits) == 0 {
		res.fail(id, "the agent never exchanged its authorization code for a token")
		return
	}

	a := o.AuthorizeHits[0]
	if a.CodeChallenge == "" {
		res.fail(id, "the authorization request carried no PKCE code_challenge")
	}
	if a.CodeChallengeMethod != "S256" {
		res.fail(id, "PKCE method was %q, want S256: plain sends the verifier in the clear and defeats the purpose", a.CodeChallengeMethod)
	}
	if a.ExtAgentHostID == "" {
		res.fail(id, "the authorization request carried no ext_agent_host_id; the route requires a stable opaque host identifier")
	}
	if o.SawLocalhostRedirect {
		// Note the flag's polarity: it records the wrong answer, not the
		// right one. "localhost" goes through name resolution and can be
		// pointed at ::1, or anywhere at all by a hosts-file entry; the
		// literal 127.0.0.1 cannot.
		res.fail(id, "the redirect_uri host was \"localhost\"; it must be the literal 127.0.0.1")
	}
	if o.SawDynamicClientAtToken {
		res.fail(id, "the agent presented the bootstrap client_id at token exchange; the authorization response issues the real one and that is what must be stored and used")
	}
	if o.IssuedClientID == "" {
		res.fail(id, "the authorization server never issued a client_id, so the flow did not reach registration")
	}
}

// ch19CheckResponsesFormat asserts the agent speaks /v1/responses and respects
// the field rules of the ChatGPT-plan route.
func ch19CheckResponsesFormat(res *Ch19Result, out ch19Out) {
	const id = "responses-format"
	r := out.resp

	if len(r.Violations) > 0 {
		res.fail(id, "the inference server rejected the agent's request: %s", r.Violations[0])
		return
	}
	if len(r.Requests) == 0 {
		res.fail(id, "no request ever reached /v1/responses")
		return
	}
	for i, q := range r.Requests {
		if q.StreamValue != true {
			res.fail(id, "request %d sent stream=%v; the plan route requires streaming", i+1, q.StreamValue)
		}
		if q.StoreValue != false {
			res.fail(id, "request %d sent store=%v; the plan route requires store=false", i+1, q.StoreValue)
		}
		for _, role := range q.Roles {
			if role == "system" {
				res.fail(id, "request %d used a system-role item; this route takes developer-role items", i+1)
				break
			}
		}
		if !strings.HasPrefix(q.Auth, "Bearer ") {
			res.fail(id, "request %d did not present a bearer token", i+1)
		}
	}
	// The second turn calls a tool. Reasoning plus tools is the capability the
	// migration unlocks, so the replay material has to survive the round trip.
	if len(r.Requests) > 1 && !r.Requests[1].IncludesEncryptedReasoning {
		res.fail(id, "the follow-up request did not include reasoning.encrypted_content; with store=false the reasoning is lost unless the agent carries it back")
	}
}

// ch19CheckReasoningSummaries asserts the summaries reached the operator as
// deltas while the model was still thinking. A summary delivered only after
// the answer completes is not something anyone can follow along with, which
// is the entire accessibility argument for the migration.
func ch19CheckReasoningSummaries(res *Ch19Result, out ch19Out) {
	const id = "reasoning-summaries"

	if len(out.resp.Requests) == 0 {
		res.fail(id, "no request to inspect")
		return
	}
	if out.resp.Requests[0].ReasoningSummary == "" {
		res.fail(id, "the request did not ask for a reasoning summary, so the vendor sent none")
		return
	}
	if out.summaryDeltas == 0 {
		res.fail(id, "the vendor streamed reasoning summaries but the GUI never saw a reasoning_summary delta")
		return
	}
	if out.textDeltas == 0 {
		res.fail(id, "the GUI saw reasoning summaries but no answer text")
	}
	if !out.summaryBeforeText {
		res.fail(id, "every reasoning_summary delta arrived after the first text delta; the summary is meant to stream while the model is still working")
	}
}

// ch19CheckCacheBreakpoints asserts the ch18 caching architecture survived the
// move. The subtle failure this exists to catch: putting the system prompt in
// the top-level instructions field. That field is a plain string, it cannot
// carry a breakpoint, and in explicit mode a request with no breakpoint that
// lands gets no caching at all.
func ch19CheckCacheBreakpoints(res *Ch19Result, out ch19Out) {
	const id = "cache-breakpoints"
	r := out.resp

	if len(r.Requests) == 0 {
		res.fail(id, "no request to inspect")
		return
	}
	q := r.Requests[0]
	if q.PromptCacheMode != "explicit" {
		res.fail(id, "prompt_cache_options.mode was %q, want explicit", q.PromptCacheMode)
		return
	}
	if q.BreakpointCount == 0 {
		res.fail(id, "explicit cache mode with zero breakpoints disables caching entirely, which is worse than never opting in")
		return
	}
	if q.BreakpointCount > 4 {
		res.fail(id, "%d cache breakpoints, the route allows at most 4 cache writes per request", q.BreakpointCount)
	}

	var devHasBreakpoint bool
	for _, bp := range q.MessageBreakpoints {
		if bp.Role == "developer" {
			devHasBreakpoint = true
		}
	}
	if !devHasBreakpoint {
		if q.HasInstructions {
			res.fail(id, "the system prompt went in the top-level instructions field, which is a plain string and cannot carry a cache breakpoint; it belongs in a developer-role content block")
		} else {
			res.fail(id, "no developer-role content block carried a cache breakpoint, so the constitution is re-read on every turn")
		}
	}
}

// A distinctive API key, so a request authenticated with it is impossible to
// confuse with one authenticated by a minted OAuth token.
const ch19APIKey = "sk-ch19-metered-0f3b81"

// ch19CheckCredentialProvider asserts the two credential kinds reach the HTTP
// layer through one interface. It is graded from the wire in two runs: with a
// stored grant the bearer must be the token the authorization server minted,
// and with no grant at all the same code path must carry the API key. A
// student who special-cased one of them fails the other.
//
// It also asserts no live credential was printed. A token that reaches a log
// is a token the operator will paste into a bug report.
func ch19CheckCredentialProvider(res *Ch19Result, out ch19Out) {
	const id = "credential-provider"

	if len(out.resp.Requests) == 0 {
		res.fail(id, "no request reached the inference server, so no credential was observable")
		return
	}
	minted := out.oauth.IssuedAccessTokens
	if len(minted) == 0 {
		res.fail(id, "the authorization server minted no access token")
		return
	}
	want := "Bearer " + minted[len(minted)-1]
	if got := out.resp.Requests[0].Auth; got != want {
		if got == "Bearer "+ch19APIKey {
			res.fail(id, "the agent was signed in but sent the API key anyway; a stored grant has to win over the fallback")
		} else {
			res.fail(id, "the inference request did not carry the minted access token")
		}
		return
	}

	secrets := append(append([]string{}, out.oauth.IssuedAccessTokens...), out.oauth.IssuedRefreshTokens...)
	haystacks := map[string]string{
		"the login command's output": out.loginStdout,
		"the agent's output":         out.agentOutput,
	}
	for where, hay := range haystacks {
		for _, s := range secrets {
			if s != "" && strings.Contains(hay, s) {
				res.fail(id, "a live token appeared in %s; credentials must never be printed", where)
				return
			}
		}
	}
	if out.resp.SecretInBody {
		res.fail(id, "a credential appeared in the request body; it belongs in the Authorization header")
	}
}

// ch19ScenarioE runs with no stored grant at all. The same interface must
// carry the API key, or the seam only ever had one implementation.
func ch19ScenarioE(res *Ch19Result, bin, skills, gui, dir string) {
	const id = "credential-provider"

	out := ch19Launch(bin, skills, gui, ch19Opts{
		dir:       dir,
		model:     ch19Model,
		prompts:   []string{"No grant stored."},
		skipLogin: true,
		resp:      ch19RespOptions{Scenarios: []string{"text"}},
	})
	if out.fatal != "" {
		res.fail(id, "with no stored grant the agent could not complete a turn: %s", out.fatal)
		return
	}
	if len(out.resp.Requests) == 0 {
		res.fail(id, "with no stored grant the agent made no request at all; the API key path must still work")
		return
	}
	if got := out.resp.Requests[0].Auth; got != "Bearer "+ch19APIKey {
		res.fail(id, "with no stored grant the request did not carry the API key as a bearer token")
	}
}

// ---------------------------------------------------------------------------
// Scenario B: the access token expires mid-session.
// ---------------------------------------------------------------------------

func ch19ScenarioB(res *Ch19Result, bin, skills, gui, dir string) {
	const id = "token-refresh"
	res.ran(id)

	out := ch19Launch(bin, skills, gui, ch19Opts{
		dir:     dir,
		model:   ch19Model,
		prompts: []string{"First turn.", "Second turn."},
		// A token that expires inside the client's own refresh window, so
		// refresh is driven by the stored expiry rather than by a 401. This
		// is deterministic: no sleep, no race between the clock and the
		// turn. An agent that waits to be told its token is dead has already
		// failed the request it was holding.
		oauth: ch19OAuthOptions{AccessTokenTTL: 90 * time.Second},
		resp:  ch19RespOptions{Scenarios: []string{"text"}},
	})
	if out.fatal != "" {
		res.fail(id, "scenario B: %s", out.fatal)
		return
	}
	if len(out.oauth.RefreshHits) == 0 {
		res.fail(id, "the access token expired and the agent never used its refresh token")
		return
	}
	if out.oauth.RefreshReuseDetected {
		res.fail(id, "the agent replayed a refresh token after rotation; the server issues a new one on every refresh and invalidates its predecessor")
	}
	if out.turns < 2 {
		res.fail(id, "the session did not survive the refresh: %d turns completed, expected 2", out.turns)
	}
}

// ---------------------------------------------------------------------------
// Scenario C: the ChatGPT plan is exhausted.
// ---------------------------------------------------------------------------

func ch19ScenarioC(res *Ch19Result, bin, skills, gui, dir string) {
	const id = "billing-mode"
	res.ran(id)

	out := ch19Launch(bin, skills, gui, ch19Opts{
		dir:     dir,
		model:   ch19Model,
		prompts: []string{"This should stop."},
		resp: ch19RespOptions{
			Scenarios: []string{"text"},
			FailWith:  "usage_limit_reached",
		},
	})
	// The turn is allowed to end in an error, so out.fatal is not fatal here:
	// an agent that stops is behaving correctly. What matters is what it did
	// with the credential afterwards.

	plan := ch19BearerOf(out.oauth.IssuedAccessTokens)
	for i, q := range out.resp.Requests {
		if i == 0 {
			continue // the request that got the refusal
		}
		if plan != "" && q.Auth != "" && q.Auth != plan {
			res.fail(id, "after the plan refused the request the agent retried under a different credential; spending money on a fallback has to be announced, not assumed")
			return
		}
	}
	// The turn must end marked as failed. Grading this by grepping the
	// agent's output would not work: the raw SSE carries the error code, so
	// an agent that ignored the failure completely still prints the string
	// and would pass. The turn_ended error field is the agent's own verdict.
	var reported bool
	for _, e := range out.turnErrors {
		if e != "" {
			reported = true
		}
	}
	if !reported {
		res.fail(id, "the turn ended without an error; a mid-stream failure arrives with HTTP 200, and a turn that reports success after producing nothing is how a usage limit becomes a silently empty answer")
	}
}

func ch19BearerOf(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	return "Bearer " + tokens[0]
}

// ---------------------------------------------------------------------------
// Scenario D: the ID token is signed with an unpublished key.
// ---------------------------------------------------------------------------

func ch19ScenarioD(res *Ch19Result, bin, skills, gui, dir string) {
	const id = "oauth-flow"
	// This scenario sharpens an already-running check rather than adding one.

	out := ch19Launch(bin, skills, gui, ch19Opts{
		dir:     dir,
		model:   ch19Model,
		prompts: []string{"Should never run."},
		oauth:   ch19OAuthOptions{UnsignedKeyMismatch: true},
		resp:    ch19RespOptions{Scenarios: []string{"text"}},
	})

	// A client that verifies the RS256 signature against the published JWKS
	// cannot accept this token, so it must never reach inference. A client
	// that merely base64-decodes the payload sails straight through.
	if len(out.resp.Requests) > 0 {
		res.fail(id, "the agent accepted an ID token signed with a key absent from the published JWKS and went on to make an inference request; the signature must be verified, not just decoded")
	}
}

// ---------------------------------------------------------------------------

func ch19Skills() (string, func()) {
	dir, err := os.MkdirTemp("", "ch19-skills-")
	if err != nil {
		return "", func() {}
	}
	base := filepath.Join(dir, "base")
	os.MkdirAll(base, 0o755)
	// One tool is enough: the tool-calling scenario needs something to call so
	// that reasoning-plus-tools, the capability this migration unlocks, is
	// actually exercised rather than assumed.
	os.WriteFile(filepath.Join(base, "SKILL.md"), []byte(
		"---\nname: base\ndescription: baseline skill\n"+
			"tools: read_file\n---\n\nAnswer briefly.\n"), 0o644)
	return dir, func() { os.RemoveAll(dir) }
}
