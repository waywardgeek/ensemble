# Independent source-only review — Chapter 1 (flint-ch01-reviewer-01)

## 1. Inspected files and coverage limits

**Supplied inputs inspected (all read completely):**
- ch01.csproj (lines 1–11): net10.0, C# 14, Exe, Ch01, Nullable + ImplicitUsings, stdlib only.
- Common/Owners.cs (lines 1–68), Common/Values.cs (lines 1–91), Ensemble/Ensemble.cs (lines 1–61), Agent/Agent.cs (lines 1–102), Engine/Engine.cs (lines 1–102), Application/Program.cs (lines 1–82).

**Public contract read:** chapter-00.md (lines 1–416) and chapter-01.md (lines 1–697); multiple bounded reads were needed due to source truncation and all ranges were subsequently covered.

**Coverage limits / evidence status:** Source reasoning only. No build, test, grader, or live-provider run was performed (per reviewer instructions: reading tools only, no execution). No test files or grader artifacts were supplied in this directory, and none were used. Parent-supplied results were not present among the supplied inputs.

## 2. Public requirements distilled from the supplied chapters

- One loop: read question → POST completed conversation + question → collect every text block → retain user/assistant pair only after success → report measured usage (not estimate). No tools/retries/streaming in Chapter 1.
- Ownership: Ensemble owns Agents + diagnostics; Agent owns conversation, configuration, Engine; Engine owns HTTP + measured usage. Engine reaches settings/diagnostics via Agent → Ensemble. Add/remove maintain both ends; refusal changes nothing; Destroy cascades strong children, removes from parent.
- §1.3: Agent's public conversation operation returns an immutable snapshot, not a writable reference to the retained list.
- §1.4: Request lifetime — no fixed deadline; HttpClient.Timeout = Timeout.InfiniteTimeSpan; pass caller's CancellationToken through header and body waits; reject unsuccessful status; Engine rejects missing usage and empty text; missing accounting ≠ measured zero; count measured usage even when no usable answer.
- §1.6: Chat needs API key + model in env; base URL defaults to Anthropic's; no model fallback; documented diagnostic when model unset.
- Exercise: JSON-lines stdio; stdout protocol only; stderr diagnostics; malformed/read error → safe diagnostic + nonzero exit without usage record; ReadLineAsync; three env vars read, none hardcoded; grader mode default, chat opt-in; seven checks.

## 3. Required findings (corrections needed)

F1 — Failed Agent construction leaves a partially constructed Agent retained by the Ensemble.
- Requirement: coding rules §6 (failed construction must not leave partial parent memberships or escaped owned work); chapter §1.1.
- Source locations: Agent/Agent.cs:8, 10–15, 18; Common/Owners.cs:39.
- Trace: `new Agent(ensemble)` → line 8 completes AddAgent (both ends mutated) → lines 10–15 throw if an ANTHROPIC_* variable is missing → exception escapes the constructor → the Ensemble retains an Agent with parent set, default Configuration, no Engine, and destroyed == false.
- Consequence: partial parent membership survives a failed construction; inconsistent owned state.
- Invariant to restore: validate required configuration before the parent add, or roll back via the declared remove on failure; refused paths must leave both ends unchanged.

F2 — Request lifetime violates the chapter's infinite-timeout/cancellation contract.
- Requirement: chapter §1.4 ("Do not put a fixed deadline on a valid generation… set it to Timeout.InfiniteTimeSpan and pass the caller's CancellationToken through both header and body waits").
- Source locations: Engine/Engine.cs:60, 44; Common/Owners.cs:65.
- Trace: default HttpClient timeout (100 s) at :60; no CancellationToken at :44; Request cannot receive a caller token.
- Consequence: fixed deadline can kill a valid generation; no caller cancellation path.
- Invariant to restore: Engine.Request accepts a CancellationToken; HttpClient.Timeout = Timeout.InfiniteTimeSpan; token passed to SendAsync.

F3 — Chat base-URL default missing; configuration error handling deviates from the documented chat contract.
- Requirement: chapter §1.6 (base URL defaults to Anthropic's; model has no fallback with a specific diagnostic).
- Source locations: Agent/Agent.cs:12–13, 10–15.
- Trace: unset ANTHROPIC_BASE_URL → throws; contract expects a default to Anthropic's endpoint; unset ANTHROPIC_MODEL → implementation stops before first request, but with a different message than the documented one.
- Consequence: chat cannot operate with defaults; diagnostic drift.
- Invariant to restore: default base URL when env absent; preserve documented model diagnostic.

F4 — Conversation read returns the live retained list rather than an immutable snapshot.
- Requirement: chapter §1.3 ("returns an immutable snapshot, not a writable reference to the retained list").
- Source locations: Common/Owners.cs:45; Agent/Agent.cs:72–75 (`return conversation;`).
- Trace: Engine/Engine.cs:23 copies via `new List<Message>(agent.GetConversation())` at request time, but the public operation itself exposes the retained list.
- Consequence: retained mutable list is handed out as the conversation view.
- Invariant to restore: return a snapshot copy.

## 4. Preferences / teaching-level notes (not required corrections)

- P1: Empty-text rejection is implemented in Agent.AddExchange (Agent/Agent.cs:80) where the chapter places it in Engine; behavior still fails safely via the Program catch (Application/Program.cs:50–60).
- P2: Engine/Engine.cs contains no explanatory comments, against the teaching standard.
- P3: Diagnostic strings differ textually from the chapter's illustrative wording.

## 5. Conclusion

**Corrections are required.** The source-only review found four required deviations from the supplied public contract and coding rules: failed-construction retention (F1), request lifetime (F2), base-URL default (F3), and immutable conversation snapshot (F4). Behavioral compliance could not be established beyond source reasoning because no build/test/grader execution was permitted or supplied.

---

{"result":"your complete review"}