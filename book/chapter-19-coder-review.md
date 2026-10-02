# Chapter 19 — coder review

**Date:** 2026-10-02
**Role:** CODER. The author owns `book/`; nothing in `book/` was edited by me except this file.
**Baseline:** `origin/main` = `60e9a29`. 21 commits on top, **all unpushed**.
**Scope:** 200 files, 49,415 insertions.

---

## 1. Status

| artifact | state |
|---|---|
| reference implementation | `./agent` |
| solution snapshot | `solutions/ch19` (153 tracked files) |
| grader | `internal/grade/ch19_{checks,harness,run}.go` + two fake servers |
| `make grade19` | **100/100** at `./agent` *and* at `solutions/ch19` |
| `make grade19-audit` | **7 of 7** mutants kill exactly their check |
| cross-chapter sweep | every `solutions/*` target 100/100; every `./agent` score equals baseline |

### Sweep, measured twice

Final sweep (788s). Every `./agent` number below was also measured at `60e9a29`
in a clean `git worktree` today, so "pre-existing" is measured, not remembered:

| chapter | baseline | now | verdict |
|---|---|---|---|
| ch5 | 120/120 | 120/120 | restored (I had broken it to 110) |
| ch7 | 100/100 | 100/100 | restored (I had broken it to 80) |
| ch8 | 80/100 | 80/100 | pre-existing |
| ch9 | 80/100 | 80/100 | pre-existing |
| ch10 | 85/100 | 85/100 | pre-existing |
| ch11 | 85/100 | 85/100 | pre-existing |
| ch12 | 90/100 | 90/100 | pre-existing |

The ch8–ch12 deficits predate this chapter and are **still unresolved**. They
are worth a session of their own.

---

## 2. The research gate, and what it found

The chapter's emotional payoff is reasoning summaries streaming for
accessibility. I treated "do they actually work on the Responses API?" as a
gate: if no, stop and report rather than build.

**They work.** Measured live against `gpt-6.1-sol`:

| `reasoning.summary` | summary deltas | parts |
|---|---|---|
| `detailed` (+ effort `high`) | 314 | 3 |
| `auto` | 186 | 2 |
| `concise` | 4 | 4 |

**A near miss worth putting in the chapter.** My *first* probe returned **zero**
summary deltas and looked exactly like a hard capability gap. The prompt was
trivial and produced 18 reasoning tokens — too little to summarise. Summaries
scale with reasoning actually performed. One sample nearly produced a wrong
"this does not work" report.

(Counting trap: `grep -c` over raw SSE double-counts, because each event appears
on both an `event:` line and inside the `data:` JSON. Count `^event:` only.)

---

## 3. Factual corrections for the author

These contradict the outline or the design doc. All verified first-party today
at `developers.openai.com` (append `.md` to any URL for clean markdown).

1. **The banned-field list is 15 fields, not 7.** The design doc missed
   `max_tool_calls`, `moderation`, `multi_agent`, `prompt`,
   `prompt_cache_retention`, `safety_identifier`, `top_logprobs`, `truncation`,
   `user`. It is `top_logprobs` that is banned, **not** `logprobs`.

2. **`instructions` is LEGAL on the plan route.** It is merely *uncacheable*,
   because it is typed as a plain string and cache breakpoints attach to
   content blocks. The chapter must therefore criticise it for **losing the
   cache boundary**, never for being forbidden. This structurally confirms the
   outline's override of design doc §19–20: the constitution has to be a
   developer-role content block.

3. **Six terminal refresh errors, not three.** `invalid_grant`,
   `invalid_refresh_token`, `token_expired`, `refresh_token_expired`,
   `refresh_token_invalidated`, `refresh_token_reused`.

4. **Three required details absent from both briefs**, each of which broke my
   first implementation and was caught by the fake server:
   - RFC 8707 `resource=https://api.openai.com/v1`, required on the authorize
     request **and repeated at code exchange**.
   - The redirect path is exactly `/auth/callback`. The docs say verbatim:
     *"`/callback` does not match `/auth/callback`."* Only the port may vary.
   - The host must be the literal `127.0.0.1`. The docs say *"Do not substitute
     with `localhost`."*
   - A stable, opaque `ext_agent_host_id` on the authorize request.

5. **This is not RFC 7591.** `auth.openai.com` advertises **no**
   `registration_endpoint`. The bootstrap `dynamic_agent_client` is a
   registration entrypoint and is *never* valid for token exchange; the real
   `client_id` comes back on the authorize callback alongside `code` and
   `state`. Worth stating plainly, because a reader who knows OAuth will reach
   for dynamic client registration and find nothing.

6. **Granted scopes must be checked.** A valid ID token authenticates the
   operator but authorises nothing; `chatgpt.tokens.use.direct` must be present
   in the *token response* before inference. Note the discovery document's
   `scopes_supported` lists only `openid`, `profile`, `email`, `offline_access`
   — it does not advertise the plan-usage scope at all.

7. **`gpt-6.1-sol` pricing, first-party:** in \$2 / cached \$0.10 / write \$2.50
   / out \$10 per 1M; context 1,050,000; max output 128,000. This independently
   **confirms ch18's ratios**: cached read is 0.05x input, cache write 1.25x.

8. **A receipt the author may want.** OpenAI ships first-class conversation
   **compaction** on the Responses API (`response_compact_params`,
   `compacted_response`, `response.compaction.compacting`) in the same period
   Anthropic froze the prefix. One vendor added a context-management capability;
   the other removed one.

---

## 4. Deviations from the coder brief

1. **I did not run `cp -r solutions/ch18 solutions/ch19`.** `./agent` was 62
   entries ahead of the ch18 snapshot and held `NoEphemera` — the Opus 5.5
   prefix restriction that *motivates this chapter* — plus `ResolveEndpoints`.
   The snapshot had neither. Building from it would have discarded the
   chapter's own premise. I built in `./agent` (which is what `grade18` targets
   too) and snapshotted at the end.

2. **Rejected `ModelFeatures.ExplicitPromptCache` and `CacheWritePremium`.**
   Duplicate columns that can drift from their neighbours: `Caching` already
   encodes explicit-vs-implicit, and a write premium is derivable from `Price`.
   Only `MaxCacheWrites` was genuinely new.

3. **Surface became a per-model fact, not a vendor-wide default flip.**
   Flipping `DefaultSurface(VendorOpenAI)` would have moved models predating the
   endpoint onto an untested surface. A test pins that earlier chapters' models
   keep Chat Completions.

---

## 5. The four seams ch18 planted

The strongest evidence in the book that the vendor seam paid for itself. All
four verified by grep today:

- `internal/common/provenance.go:35` declares `SurfaceResponses` — named
  chapters ago and never wired until now.
- `internal/common/delta.go:120` documents that `StreamAll` deliberately
  excludes `StreamReasoningSummary`.
- `openai.go` carries a comment that encrypted reasoning is "a Responses-API
  concept ... for a renderer that can use it".
- The `gpt-5.6-sol` row records that function tools with reasoning are refused
  on `/v1/chat/completions`, that the fix is `/v1/responses`, and that this
  "records the constraint **until someone does it**."

Chapter 19 is someone doing it. The migration does not only change auth: it
**unlocks thinking-plus-tools** on OpenAI models.

---

## 6. Bugs I shipped, and what caught them

Recorded because the chapter is partly about instruments that tell you the
truth.

1. **My renderer sent `max_output_tokens`**, which the plan route bans. Found by
   the fake server, not by reading code. I had also written a doc comment
   asserting credential kind *never* changes how a request is built — false, and
   now corrected in two places. The fix models the route's rules as **data keyed
   by `CredentialKind`** rather than a branch inside a renderer.

2. **An inverted grader assertion.** The fixture field `SawLocalhostRedirect` is
   a *violation* flag: true only when the client wrongly used hostname
   `localhost`. I read it as a success flag, so the check **failed a correct
   client**. A field named `SawX` that means "saw the wrong X" is a trap; read
   the setter, never the name.

3. **A flaky grader.** All scenarios ran in `./agent`, and the agent defaults
   `save.json` to its working directory, so each scenario resumed the previous
   one's conversation and the result depended on run order. Symptom: an
   intermittent first-turn failure that looked like a parser bug. Fixed with a
   private `--save` per launch.

4. **`billing-mode` graded echoed bytes.** It grepped the agent's output for the
   error code — but the raw SSE *carries* that string, so an agent that ignored
   the failure entirely still printed it and passed. It now reads the `error`
   field of `turn_ended`: the agent's own verdict, not a vendor byte it relayed.

5. **I broke ch5 and ch7 and did not notice for hours.** New code introduced
   package-level mutable state, which the no-mutable-globals rule has policed
   since ch5 and which parity re-checks in later chapters. ch19 scored 100
   throughout. Only the cross-chapter sweep found it. Fixed by making the
   declarations genuinely immutable — const-typed sentinel errors, a switch
   instead of a map, a function returning a fresh slice — rather than renaming
   them past the detector's heuristic.

---

## 7. Mutation audit

Seven mutants, each deleting exactly one behaviour and each still compiling.
7/7 kill exactly their expected check.

Two lessons generalise beyond this chapter:

- **My extraction grep was broken and reported all seven as survivors.** The
  text report prints the check *title*; only `-json` carries the `id`. An audit
  that cannot fail is the same failure it exists to detect. Prove the harness
  can report a kill before believing a survivor.
- **A surviving mutant can mean the mutant is invalid.** Mutant 7 survived
  twice: first because the parser has two independent guards (records the error
  on `response.failed`, *and* refuses a stream ending without
  `response.completed`), then because the scenario never reached the mutated
  path — `FailWith` produces an HTTP admission error, while the 200-then-fail
  path needs the `fail_midstream` scenario. Both diagnoses made the grader
  stronger: it now grades **both shapes of a plan refusal**.

Two mutants legitimately kill more than one check. Breaking the credential seam
really does mean wrong bearer, server rejects, no refresh. I recorded the true
multi-check expectation rather than narrowing it to make the table tidy.

---

## 8. Open questions for Bill

1. **ch8–ch12 at `./agent` score 80–90.** Confirmed pre-existing today against
   `60e9a29`. Unresolved, and not mine to silently fix.
2. **The outline's "5x cost" claim is unverified.** I did not find a first-party
   source. It should not be printed as a measurement.
3. **`gpt-5.6-sol` and `gpt-6-astra` were never probed for reasoning-summary
   support**, so their rows deliberately do not claim `StreamReasoningSummary`.
   Do not add the capability without a measurement — the table records
   measurements, not expectations.
4. **`agent/openai_auth_design.md`** (your file, 1,674 lines, added in
   `60e9a29`) is tracked under `agent/`, so it is now also in
   `solutions/ch19`. Say the word if it should be excluded from the published
   snapshot.
5. **`CallEphemeral` still runs tools whose output is dropped** for
   `NoEphemera` models. Carried over from the last session; still needs a
   ruling.

---

## 9. Things deliberately NOT done

- Nothing pushed. Nothing in `book/` edited. `book/voice.md` is still dirty with
  your edit and I never staged it.
- No attempt to fix ch8–ch12.
- No live probe of the real `auth.openai.com` authorization flow. The OAuth
  client is exercised only against the fake server, which enforces the published
  rules. The *inference* side was probed live; the sign-in side was not.
