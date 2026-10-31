# Edition 2 Chapter 17 grader validation

Coder: GPT-6 Astra/high. Original source: Chapter 18, *The Invisible Invoice*,
revision `ca9d807d9e344375037d0c75bb29185192920aab`, SHA-256
`2224c498d28cb1ba2aaec8b08eb3e21153f80f816842b4898929522e3f173d11`.
Mandatory `.agents/skills/ensemble-coding/SKILL.md`, full carryover/workflow and
current enhancement/map were reread after compaction. Student and comparative
reviews and the original grader were read completely. Final source and evidence are frozen for independent review. Original source,
runtime and frozen exports are unchanged; acceptance is the reviewer’s decision.

The separately versioned grader preserves all seven original weights, totaling
100, without recursive parity. It uses the actual optional GUI, a typed primary
with five declared real tools, public completed-turn observations and fresh
subscription meter snapshots. Subscribers remain alive until normal EOF; the
process must save and exit successfully. The GUI's actual default save is
cwd/save.json, despite the accepted Agent's data directory being cr. No original,
runtime or frozen-source file is changed.

Run `go run ./cmd/grade-edition2 -ch 17 -json`. Grader-only artifact options
`-cache-current` and `-cache-prior` default to `cr/cache-current.txt` and
`cr/cache-prior.txt`. `-cache-prices 2,2.5,0.2,10` supplies the accepted October 11
Sonnet-5 rates per million fresh input/cache write/cache read/output tokens.
These are dated fixture rates, not a current pricing recommendation or values
read back from the implementation under test. Legacy uses:

```sh
go run ./cmd/grade-edition2 -ch 17 -json -gui solutions/ch18 \
  -cache-current request.json -cache-prior prior_request.json \
  -cache-prices 3,3.75,0.3,15 solutions/ch18
```

The independently agreed legacy tariff tests its declared historical arithmetic,
without treating different prices as a calculation defect. No provider call is
paid and no active fixture selects Anthropic 5.5.

The stage reports under `solutions/edition-2/evidence/ch18/enhancement17/` record
actual accepted **100/100**, exit 0, 0.756 seconds, and legacy **85/100**, exit 1,
0.625 seconds. Earlier actual full 100/85 measurements remain separately bound.
Legacy omits the tool-definition breakpoint: actual captured requests have zero
tool markers, one system marker and advancing history markers. Original Chapter
18's TL;DR and vendor-asymmetry section explicitly require that tool breakpoint;
this is missing valid coverage in the original grader, not a new feature.

The initial local test decoded `response_ended` as the payload key instead of
actual `response`; that setup failure is retained in `initial-fixture.json`.
A follow-up capture run crashed legacy with `send on closed channel` after a
short-lived meter subscriber disconnected. Its actual 0/100 panic report is
`legacy-wire-report.json`, separate from the successful 85/100 baseline. Keeping
observing subscribers open avoids making that unrelated known GUI lifetime
failure a prerequisite to cache grading; normal EOF is still checked. The first
line-based dump parser also rejected valid pretty JSON; delimiter-based receipt
parsing corrected only that observation. `legacy-markers.json` covers only the
one request captured before the panic, while `legacy-held-wire.txt` contains the
successful full run's two-turn marker evidence.

The new checks require actual rotating request content, nonempty stable tools
and primary instructions, direct markers on the final tool/system/stable-history
blocks, exact durable and live disjoint usage buckets, cumulative price and hit
rate arithmetic, unknown-price reporting, and retained dialogue alongside zero
initial restart usage. Gemini's separate thought/output counters are exercised.
Artifact JSON is decoded for content comparison, allowing report indentation and
object-key order; this is not byte-offset diagnostic proof. The stable wire
sections themselves are compared without reserializing their content.

The original **0/100 on 7436fcb** remains historical: omitted skill type blocked
startup and masked all seven checks. Twelve successful real browser turns and
their actual grants, dated model identities and earlier-source provenance remain
in ch18 notes/review. No new cache grant, native DOM observation or invoice is
claimed. Main-only lens routing, detailed byte diagnostics, surviving-memory and
ephemera boundaries, OpenAI caching variants, per-model/tier attribution and
interrupted-worker prediction ownership retain their accepted source, mutation
and live evidence separately; this small numeric fixture does not certify them.

The first green stage was committed at `b1b7c9f6`. Reviewer preflight strengthened
legal marker placement and retained original spend/charge monetary-key checks
before that checkpoint. Stage race/vet receipts remain separate from final ones.

## Final measurements

Final actual full commands score **accepted 100/100**, exit 0, **0.984 seconds**,
and **legacy 85/100**, exit 1, **0.689 seconds**. These are complete runs, not
composed totals; `accepted-final-run.json` and `legacy-final-run.json` bind their
exact commands, output and source hashes. Only legacy's missing legal last-tool
marker fails. Final race passes in 1.529 seconds and vet exits 0. No broader
unchanged suite or paid model run was repeated for freshness.

Final source is bound in `final-source.json`, including command, driver, check
and test hashes. The accepted frozen tree remains
`2ee06bb6189898809d80de97df2b73c695cfb0d5`; legacy remains
`31df8d9c254722401c5cf0378e7c532eb015bba6`. Git reports no changes in either
frozen export, original Chapter 18 or its grader. Concurrent main/crossover and
first-edition work is outside this commit's exact owned paths.

Reviewer found that matching each expected request independently could accept
two identical files, each containing both requests. The original distinct-file
guard was valid coverage and had been omitted. `duplicate-pair-initial-survivor.txt`
preserves that failure on the stage grader; `duplicate-pair-corrected.txt` records
rejection after restoring the guard. These use captured real request bodies and
actual scratch files. Independent reviewer controls also allow distinct raw and
pretty captures. The numeric lens check still makes no claim about arbitrary
verdict wording or byte-offset diagnostics.

## Targeted sensitivity and its limits

`mutations.json` preserves 25 source-defect runs between unmodified/restored
controls; `followup-mutations.json` adds two, and `final-followup-mutations.json`
adds two. All **29 source defects** are detected, with overlapping observations,
plus the separate duplicate-artifact control above. They are not 29 independent
proofs. Each patch was applied to an isolated copy of the accepted core and GUI,
compiled, run through the actual client against external HTTP, then restored.
The scratch copy was removed. No mocks or runtime interfaces were introduced.
`probe-source.go.txt` preserves the transient Go overlay used for these focused
checks; it invokes the actual revised grader functions.

- Captures: missing rotation, empty captures, and distinct nonempty captures
  containing wrong request content fail. The empty-current defect makes both
  files empty and is caught by equality; the distinct-wrong-capture defect
  separately establishes content sensitivity.
- Prefix: absent tools/system and changing tool/system content fail. Absence
  causes several coupled failures; these are not independent claims for each row.
- Markers: first-tool placement, schema-nested metadata, nested system/history
  metadata, stale history and excess markers fail. A separate final defect drops
  old history while marking the current message: it fails actual advancement
  and, separately, restored-dialogue retention.
- Accounting: wrong units, omitted cache-write price, last/total confusion,
  output in the hit-rate denominator, last-only rate, and unknown-priced-free
  reporting fail. `last-shows-total` initially named the session row, but that
  row stays passing; the actual failures are the cost and cache-read rows.
  Its incorrect target label is retained rather than reported as a session-row kill.
- Persistence: discarding restored dialogue fails despite a fresh meter;
  seeding the initial live total from restored history fails despite retained
  dialogue. A persisted `spend` field fails the counts-only obligation.
- Vendor buckets: OpenAI and Gemini inclusive-input double counting, omitted
  Gemini thought output and omitted Anthropic cache-write counts fail. The broad
  Anthropic omission has coupled meter/cost failures as well as parser failure.
- Successful completion: a GUI that saves normally and then exits 17 fails
  the launcher, despite complete response and save artifacts.

The 25-case batch and first two follow-ups ran after the duplicate-pair fix but
before the final explicit message-index advancement guard and documentation line.
Their source history is retained, not relabeled as 29 runs on final bytes. The
last two cases and final full/race controls use the final hash. The earlier
restored control passes; final full/race checks also verify the restored source.
A temporary follow-up script edit initially failed Python parsing before any
source mutation or child launch; it is a setup error, not a defect result.

Unchanged richer behavior is reused from `ch18.md` and `ch18-code-review.md`:
actual main/auxiliary ownership, rendered byte comparisons and marker stripping,
compaction/ephemera positions, price tiers and request identity, native GUI
readings, and interrupted-observation/prediction tests. Those records preserve
35 historical attempts (32 intended failures, three invalid setups), the later
three focused concurrency/rotation defects, and the accepted `afc7dd7` correction.
No active test newly selects the frozen historical Haiku 5.5 row. The current
numeric arithmetic fixtures cover Sonnet-5's ordinary tariff and three vendors'
plain response normalization, not every tier, stream or route.

## Published examples

`samples.json` binds all nine fences to manuscript SHA-256
`c00a640ca14d7b4bbde424314230857841abd01249cc3f9c9fe976290642ff48`.
Both Go vocabulary fences compile in an isolated core module with the existing
`common.Engine` context. The initial harness wrongly assumed a public
`ensemble.Engine` alias; that failed setup and its correction are preserved.
The manuscript described common vocabulary and needed no edit.

The exact shell setup builds the separate GUI and creates the full 700-record
typed primary. All four published prompt lines pass through the real WebSocket
and Agent. The final meter reports 1300 fresh input, 500 cache-write, 4200
cache-read and 70 output tokens, $0.00539 and a 70% session hit rate. The actual
restart retains four completed responses while beginning with zero live usage
and cost. Both processes exit 0. Captured requests, meters and rotated files
are retained alongside the receipt.

Port 8088 was already occupied by unrelated Ensemble PID 4689. The first attempt
subscribed and read its meter, then stopped at the nonzero-initial assertion
**before sending any prompt**. That is a harness collision, not a product defect;
the existing process was preserved. The successful run substitutes only free
port 51108 in the published launch and checks that the listener PID is the newly
launched child before sending prompts. Existing provider environment settings
point to a local HTTP fake and select plain delivery. The receipt includes both
published and executed commands. This is exact sample plumbing with that stated
port substitution, not a fresh provider grant, native DOM or model-quality claim.

## Size and review boundary

Counting physical lines, nonblank comment-only `//` lines and remaining active
lines, the new production grader has 526/27/480 respectively; its focused test
has 25/2/21. The original three grader files total 998/172/751. Shared existing
helpers and command dispatch are separate from these chapter-file counts.
Comments explain subscription lifetime, completion barriers, representation
choices, metadata-only comparison and tariff provenance. The approximately 5%
comment share is below Bill's teaching target; this is recorded for substantive
review rather than padded with syntax narration. All new exported symbols and
the local decoded Usage member have purpose comments. The smaller line count
alone is not a quality or completeness claim.

Independent combined acceptance is pending. No Chapter 18 enhancement or Gateway
implementation is started by this submission.
