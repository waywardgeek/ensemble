# Edition 2 Chapter 20: web access, status and accounting

Coder: GPT-6 Astra/high. Submitted for independent review; no later chapter is
released here. This combines originals 21 and 22 against frozen
`solutions/edition-2/ch20`. The entire mandatory
`.agents/skills/ensemble-coding/SKILL.md`, carryover and workflow were reloaded
after compaction; both originals, full student/comparative records and original
graders were read. Runtime, frozen exports and original graders are unchanged.

## Scope and results

The agreed rubric has **155 points**: original web seven checks retain 100;
original build/status/accounting/import checks add 5/20/15/15. Direct web checks
replace redundant ten-point parity. Parent relationships and executable
composition remain mandatory independent source-review obligations, without the
old 20/15 points for identifier shapes. Neither a passing graph nor 155 proves
those relationships.

Evidence is under `solutions/edition-2/evidence/ch21/enhancement20/`; original
22 evidence stays in its original directory. `final-source.json` binds the
command and four grader/test files, accepted tree
`a801628b1dd631e5ea20ad03ea7d255584728d31`, legacy tree
`0d0c7787b00c2b8d1e811d3137228115cc9302ec`, and manuscript. Final actual full runs:

| Source | Score | Process | Wall time |
| --- | --- | --- | --- |
| Accepted combined `ch20` | **155/155** | exit 0 | 1.490 s |
| Legacy original `solutions/ch22` | **135/155** | grade exit 1 | 1.260 s |

These are actual full runs, not composed totals. `accepted-final.json` and
`legacy-final.json` have adjacent `-run.json` command/exit/time receipts.
Legacy fails only status's cache-hit denominator: 16.7/29.4/29.7 percent rather
than 13.333/23.810/24.176, because its formatter omits cache-write input. Actual
web operations and separately priced legacy accounting pass. This is a real
formatter defect, not rejection of different historical prices.

```
go run ./cmd/grade-edition2 -ch 20 -json solutions/edition-2/ch20
go run ./cmd/grade-edition2 -ch 20 -gui solutions/ch22 -status-prices 3,3.75,0.3,15 -status-second-prices 5,6.25,0.5,25 -json solutions/ch22
go test -race ./internal/gradeedition2 -run '^TestChapter20Combined$' -count=1
go vet ./internal/gradeedition2 ./cmd/grade-edition2
```

Final focused race passes (package 1.899 s; command 3.416 s), vet passes,
formatting is clean. No inherited full-book sweep or new paid session was run.
Fresh cwd/save/profile/ephemeral ports isolate each launched client; no user
listener on 8088 was accessed.

Original web **0/100** and architecture **20/100** remain historical results.
The web grader waited for a legacy completion envelope after actual MCP work;
the architecture grader launched the headless command as GUI and selected the
wrong source shapes. Its awarded import points likewise were not proof of the
intended owner path. Original records and exceptions are not relabeled passing.

## Corrected fixtures and observations

The initial revised **35/155** is retained in `accepted-initial.json`. The skill
omitted `transport: url`, so load returned `invalid MCP server declaration`,
masking downstream web checks. A history assertion also used a recall-only text
extractor. Those were grader errors, not runtime diagnoses. Inspection replaced
a flat-key meter comparison, whose absent keys could agree vacuously, with actual
nested totals. First green `7c592004` scored 155/135, with race 2.054 s; its
`stage-source.json` and reports remain separate from final receipts.

The revised web scenario uses the actual CLI, a typed primary and loadable skill,
real HTTP JSON and multi-frame SSE, session echo, environment bearer, extra
advertised capability, correlated calls/results, named missing-env degradation,
MCP `isError` and HTTP body errors. The second provider request is a causal
pre-load observation: load has not been issued and the peer has received no HTTP.
Initialize/initialized/list cadence is observed. Guessed pre-load, ungranted and
post-unload calls must produce unknown-tool refusals without remote execution.
The final unique terminal answer precedes EOF, and partial artifacts cannot
excuse a failed process exit.

Reviewer preflight found that marker-only delivery accepted missing planted
words, and any error accepted inventory-revealing disabled-tool text. The initial
`partial-web-body` and `disabled-inventory-error` mutants both passed on
`7c592004`; their receipts remain. Corrected checks require both complete planted
text blocks, with joining separators unconstrained, and ordinary unknown-tool
category wording. Both defects now fail their intended rows. The page marker is
checked inside the correlated result, then excluded from system/operator content;
this does not prove a model will ignore malicious instructions.

Status uses the shipped primary and separate actual GUI client. Completed turns
and effective selection acknowledgments precede actual captured requests. The
next request must use the selected model; old conversation and measured totals
must remain. Last-response and session values include all four disjoint buckets;
cache hit rate divides by fresh + read + write input. No mandatory printed
per-model breakdown is invented.

Current accepted fixtures use supported Sonnet-5 → Opus 4.6 → Sonnet-5.
Independent dated Sonnet rates are 2/2.5/0.2/10 per million, ordered
input/write/read/output. Opus is unpriced in this frozen table: first-phase cost
is numeric; later cost must remain honestly incomplete even after returning to
Sonnet. Legacy's explicit rates are 3/3.75/0.3/15 and 5/6.25/0.5/25, so its
combined arithmetic is checked numerically. An explicitly qualified lower bound
is permitted. These are fixture identities, not current provider-price claims.
No 5.5 model was newly selected and no table was changed. Accepted two-priced
arithmetic retains its historical source/test evidence and original identities;
this new accepted scenario does not demonstrate two known prices.

## Sensitivity and retained limits

`mutations.json`, `mutations-extra.json` and `corrected-survivors.json` describe
**18 distinct source changes: 16 detected, two surviving this numeric fixture**.
The two pre-correction survivors add earlier attempts, not two extra defects.
All post-correction tests use behavior-identical final grader code; the later
source changes are comments and import grouping only. Each patch, complete
failure array and output is retained; `focused-overlay.go.txt` drives real
isolated submitted binaries. Final full runs are restored controls.

Detected changes cover complete content versus markers, leaking refusal,
auth/session forwarding, absent-env diagnostic, second result block, MCP/HTTP
errors, actual whole-unload omission, successful work followed by exit 17,
result-kind corruption, cache-write denominator, changed cost, forgotten
unpriced work, ineffective model selection, and a real tools→engine import.
Several fail together: result-kind corruption loses all correlated results;
exit 17 fails the whole web scenario; whole-unload omission affects declarations,
refusal and remote-call count. These are not independent proofs for every row.

Removing only the manager's grant filter survives because Agent-level grants
still protect this fixture. Removing only manager `StopSkill` selection survives
because skill unload separately revokes the visible/executable grants. The
numeric result therefore does **not** prove per-server isolation or connection
teardown. Original `ch21/skill-grants-dispatch.txt` records actual earlier-source
unauthorized declaration/dispatch failures, and `unload-noop-dispatch.txt`
records the client-close timeout. Current manager source and retained real
integration tests provide the stronger review evidence; those historical runs
are not rebranded as new final-source executions. The actual current
`TestMCPHTTPSkillAndResults` explicitly observes `MCP().Done` on unload;
`TestMCPHTTPJobCancellationAndClose` observes blocked POST cancellation.
`internal/mcp/http_test.go` separately checks the parked receiver in
`TestHTTPCloseWakesReceiver`.
Their earlier mutation, redirect, redaction and ordinary-timeout proofs remain
separate. No broader cancellation or observer stress sweep was repeated.

Original 22's 14 targeted cases and actual model-switch/live receipts retain
attributed counts, unknown-price honesty, prior-history and observer-lifetime
proof. Original 21's incomplete first live run and corrected 23.35-second
search/scrape/wait task remain separate: selected/truncated output is not full
remote-page delivery. No new fake session establishes autonomous research,
provider grants, service availability, an invoice or browser usability.

## Actual ownership and executable source review

`reviewed-source.json` binds the inspected sources and retained evidence.
Accepted `internal/tools/tools.go` registers the real status handler;
`internal/agent/dispatch.go` constructs `callContext` with its actual Agent.
The context's `Engine()` reaches `parent.Engine()`; status calls that Engine's
`Meter()` and its owning Agent's configuration. Engine stores its required Agent,
Agent stores its required Ensemble, and attachment checks enforce matching
parents. `ensemble/ensemble.go` creates and retains these objects. The actual
CLI (`main.go:120`) and GUI (`gui/cmd/main.go:110`) call that public root's
`NewAgent`; this is not an unused auxiliary constructor. Engine accounting keys
counts by request vendor/model/tier and computes display cost from those counts.

The numeric graph positively resolves the actual tools/common packages and
import edge before checking forbidden spoke imports. It cannot establish
constructor use or behavior ownership, hence the separate source obligation.
Legacy reaches its Engine through `common.Call.Engine` and uses public
`agent.NewAgent` in its command, preserving its original architecture; it does
not acquire the accepted Ensemble ownership design merely by scoring import
points. No source-name detector awards either tree ownership credit.

The new three production grader files contain 683 active, 34 comment-only and
737 physical Go lines; the focused test has 22 active, three comment-only and
28 physical lines. Counting excludes blank and `//`-only lines from active code.
This remains a small chapter-specific peer plus two actual-client scenarios,
reusing existing build, request, GUI and result helpers. Comments explain causal
barriers, lifetime, source-proof limits and price ownership. The comment ratio
is below the teaching target; no filler was added. The exported Chapter20 API
and test are documented; independent review assesses explanatory sufficiency.

## Exact published examples

`samples.json` binds all **17 fences** to manuscript
`b975d7aaa77d767ca7e6c7f3976367048cab89b15a6491d307b9182ee89cb1f9`.
The two current Go fences compile against the actual common/tools vocabulary;
interface names are renamed only to avoid redeclaration, with bidirectional
assignments checking compatibility. The five historical diagnostic Go excerpts
are intentionally not current runnable setup. Initial overlay import/redeclaration
errors are preserved in `samples-go-initial.txt`, corrected receipt in
`samples-go.txt`.

The exact build, complete skills, launch, three prompts and `/model` command run
through the frozen CLI with an isolated signed fake issuer, local Responses peer
and synthetic Firecrawl credential. Only the web endpoint is substituted in
its published skill. Final sample test passes in **1.152 s**, with nine main
requests, three valid empty-pick judge responses, full correlated local search/
fetch content, unload, both status calls, save and exit 0. Astra/high/detailed
is captured for the first seven main requests, Sol/high/detailed for the last
two. The status totals are respectively 400/55/100/0 and 640/88/160/0 in
fresh-input/output/read/write order; both costs honestly remain unknown.

The initial sample reached the model change but the validator watched stdout
for an acknowledgment actually written to stderr. Its failed output is retained;
correcting that observation changed no manuscript/runtime behavior. The local
scrape completes immediately, so the prompt's conditional wait branch is not
exercised anew; historical real wait evidence remains. Scripted decisions prove
CLI/tool mechanics, not another autonomous web-research success, real catalog
eligibility or paid caching. No paid inference or production auth profile was used.
