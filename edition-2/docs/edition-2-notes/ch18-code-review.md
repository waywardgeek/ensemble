# Original Chapter18 / Edition2 Chapter17 — independent review

## October11 public-contract preparation — no implementation verdict

Reviewer `enhancement_reviewer_resumed` read all665 lines of original
`book/chapter-18.md`, revision `ca9d807d9e344375037d0c75bb29185192920aab`,
SHA-256 `2224c498d28cb1ba2aaec8b08eb3e21153f80f816842b4898929522e3f173d11`,
and the complete current [student-safe preparation](ch18.md). Read the mandatory
`.agents/skills/ensemble-coding/SKILL.md` and current governing documents; the
student extends accepted original17 tree
`83e2f4cc85eb91ef9f62a54f638baa4d538cacf8`. No original answer, grader or changing
implementation was inspected, and no test/provider call was made for this prep.

Scope is the actual-request cache lens, supported vendor breakpoints and visible
session usage/cost meter. Independent review will distinguish bytes from tokens,
stable payload from moving marker metadata, and fake grants from actual provider
cache reads. Main conversation requests belong in the lens; stateless judge and
compression calls must not create false conversation-prefix alarms. Genuine JSON
array growth and Gemini's `contents` name need the same correct classification as
`messages`. The byte-measurement warning prohibits silently substituting remarshal
sizes for the measured artifact; it does not prohibit parsing to classify shape.

Existing carryover resolves the source's obsolete memory-in-system assumption:
do not move mutable data into the prompt. Identify the actual surviving compaction
boundary in the accepted implementation. The stable end excludes ephemeral content
even when rendering merges it into a message with retained content. Marker limits
are budgets, not reasons to invent missing tool/system/handoff content. OpenAI's
explicit opt-in depends on a marker actually landing. These are chapter obligations,
not a prescribed implementation or additional cache protocol.

The meter must retain the distinction between session totals, latest response,
restored history and the effective model. Counts are durable facts, money is
derived, unknown pricing is visibly unknown. The coordinator's dated provider
references already cover model support, current field names, thresholds and pricing
tiers; this preparation does not repeat that research or independently claim live
grant. Historical numeric rates, warm-up patterns and precise provider lookback
counts are source observations, not guaranteed acceptance targets. No new product
ambiguity presently needs Bill's decision.

The public chapter supplies the conceptual lens methods and visible meter fields
but not a concrete external grading launch, artifact-path/schema or GUI message
envelope. If those facts are needed, the coordinator may recover only the missing
invocation/observation contract under the standing authorization and record exactly
what was supplied. It must not impose the original implementation's method names,
single-module layout or private source paths. A mismatch remains a review question,
not a reason to invent compatibility machinery before seeing an actual result.

Await frozen implementation, original grade and completed student evidence for
formal comparative review. Zero review rounds or acceptance claims begin here.
No future chapter or authorship is released; original22/24 stops remain. Eventual
accepted export is direct Edition2 `ch17`, owned by the coordinator.

## October11 original grader preflight — first green, not acceptance

Independently read complete unchanged `ch18_harness.go`, `ch18_run.go` and
`ch18_checks.go`, plus the relevant frozen launch/skill/lens/usage/save paths at
`7436fcb35ef54e2996114a2d3fe68fae4cfac82e`, main tree
`05e0d2e92a9f70f5d2433af76c1a3cbf5208b493`. The actual completed original result
is **0/100 FAIL**, grader exit1/make exit2, 13:34:32–13:34:54 UTC. See
[output](../../solutions/evidence/ch18/original-grade.txt) and
[provenance](../../solutions/evidence/ch18/original-grade-provenance.json),
output SHA-256 `82053233f0bfdee03295053fdbb7b400df3f1406bc056db4de4e26e1e3a3dbe1`.
No reviewer grader or behavioral probe ran. Final student evidence is pending.

The primary fixture at run.go:80–97 omits `type`; inherited parse.go:96–99
requires primary/loadable/dependency, as original10 documents. `root.NewAgent`
fails before GUI startup (gui/cmd/main.go:102–104). The prefix launch assigns
that failure to six checks, then the vendor scenario fails on Anthropic and
returns before OpenAI/Gemini. There were two attempted launches, not seven
independent failing implementations. There is no recursive parity in Chapter18.
The missing-type fixture exception is independently justified; all seven valid
behavioral obligations remain unmeasured by this result.

Three unreached assumptions need separate treatment:

- The collector only reads cwd `request.json`/`prior_request.json`
  (harness.go, ch18ReadLensFiles), whereas the real Engine-owned lens writes
  `cache-current.txt`/`cache-prior.txt` below its Agent DataDir
  (cachelens/lens.go:139–148). GUI selects DataDir `cr`. The chapter explicitly
  permits artifact framing choices and requires inspectable consecutive content,
  not these filenames; ownership is binding. This is a latent naming/path mismatch.
- Session check run.go:357 requires cwd `save.json`. The Agent's inherited save
  path correctly uses DataDir/save.json, so a real saved `cr/save.json` would be
  missed. Correct session reset must still be proven against actual restored
  history; a path mismatch does not establish the meter's behavior.
- The cost check requires `claude-sonnet-5` to report priced/positive cost.
  Frozen student pricing currently covers the three independently verified live
  rows and displays other prices as unknown. This is not an observed failure or
  automatically accepted exception: final review must assess the student's scope
  explanation and actual pricing evidence. Inventing a price to satisfy a fake is
  inappropriate; a truthful unknown display alone is not proof of correct pricing.

The fixture's undefined `think` grant is **not** another observed startup failure.
No Chapter18 fake calls it, and declarations iterate implemented tools and their
visibility, rather than demanding that every granted name have an implementation
(agent.go:74–81). Do not invent a think tool or imply its missing contract caused
the seven failures. Valid nonempty tool-prefix checks can use real existing tools.

| Original check | Still-required independent evidence |
| --- | --- |
| cache-lens-exists,15 | Actual main-request rotation, readable sections and truthful byte measurement; marker stripping, append versus edit, and auxiliary-call exclusion. Mere file presence does not prove these. |
| prefix-is-stable,20 | Independently compared rendered tool/system bytes across real turns and correct change detection. Original predicate skips missing sections; do not infer coverage from an empty prefix. |
| request-has-breakpoints,15 | Actual supported vendor markers, advancing stable end, surviving compaction boundary, tool/system placement, budget and ephemera exclusion; OpenAI opt-in only when a marker lands, implicit routes preserved. Original check tests only Anthropic system presence/history advancement. |
| cost-is-computed-not-stored,15 | Durable integer counts without money; GUI derives correct nonzero price, distinguishes unknown, uses effective model and appropriate request tier. Original greater-than-zero check alone cannot prove arithmetic. |
| usage-is-session-scoped,10 | Real restart retains history but clears process spend/totals; latest-response values and session tooltip/rate remain distinct. Original check primarily compares latest input and does not establish cost/rate reset. |
| cache-read-rate,10 | Exact normalized provider reads, correct denominator and visible reporting, plus bounded live cache grant. Fake cold/warm values cannot establish provider eligibility or successful caching. |
| all-vendors-report-usage,15 | Actual parser paths preserve disjoint input/write/read/output counts, including streaming usage; absent provider write fields are not invented. Only the first failed Anthropic launch occurred in this original run. |

Harness/run hashes match the coordinator's public erratum. Checks revision
`9bd6b14a033db347212bc7f1aae8a837816bcced`, SHA-256
`00ffe9ddfb7745c71325efd0230d14ebb9232b49fb80a3a678a682bc689b9bb7`.
The original stripper itself recognizes only a few exact Anthropic spellings;
its presence is not broad artifact-format coverage. Preserve valid checks and
source expectations for later grader revision, without changing this grader now.
Await frozen final student review, actual live/browser/mutation evidence and
comparative source assessment. Zero formal correction rounds; no acceptance yet.
Mandatory handoff: `.agents/skills/ensemble-coding/SKILL.md`.

## October11 live-artifact inspection — final submission still pending

After the student moved the unchanged receipts to authoritative
`solutions/edition-2/evidence/ch18`, independently decoded all12 real request/
response exchanges from `*-api.log`, their recorded `*-event.log` usages,
WebSocket `*-gui.log` meter frames, `*-browser.json` DOM readings and saved
context totals. Source is the first green7436fcb, not the pending final source.
The inspected Gemini screenshot also visibly matches its recorded final meter
and answer. No browser or provider process was launched by this inspection.

Normalized raw final usage independently: Anthropic retains disjoint input/write/
read counts and final output; OpenAI subtracts both cache details from total prompt
tokens; Gemini uses its final revised prompt total minus cache reads, and adds
visible candidates plus thinking for output. Summing intermediate stream usage
would be wrong; Gemini's final frame changes prompt totals as well as revealing
cached tokens. Each of the12 normalized records exactly matches its durable event.

Using the coordinator's already sourced below-threshold rates, independently
recomputed each cumulative USD amount, session hit fraction, latest total input
and cache read, and tooltip totals. Exact WebSocket values and rounded DOM text
match all12 cases. Final values are:

| Provider/model | Cache reads by turn | Exact final USD | Final session hit rate |
| --- | --- | ---: | ---: |
| Anthropic / claude-haiku-5-5 | 0,21924,21968,22058 | 0.003481475 | 74.8879% |
| OpenAI / gpt-5.6-sol | 0,16349,16349,16349 | 0.1040918 | 74.6439% |
| Gemini / gemini-3.8-flash | 0,15725,15801,15910 | 0.0226167 | 66.2209% |

These session rates include the cold first turn; they are not the last request's
grant ratio. Prices are small because these are bounded runs, not free requests.
Raw serialized tools and system values stay byte-identical across each session;
for OpenAI the embedded first system message is byte-identical. Comparison used
JSON value offsets into original request strings, without remarshal sizes.
Anthropic's three markers occupy final tool, system, and advancing message0/2/4/6;
OpenAI's two occupy system and message1/3/5/7, with explicit mode actually present.
Gemini carries no markers and its raw responses identify gemini-3.8-flash. These
no-handoff runs do not exercise compaction-bound placement or ephemera exclusion.

All captured answers report the archive code and requested facts consistently with
the supplied archive text. Each provider grants nonzero reads from turn2 here;
the source's generalized Gemini two-cold-turn account is not a universal property.
Save totals equal the sum of the four observed responses, and initial meters
start at zero. There is **no restarted process in these captures**, so they do not
establish reset after loading a prior save; the separate deterministic restart
control must supply that evidence. Likewise no price-tier transition, model change,
unknown-price behavior or broad retrieval-quality result is inferred from these
live sessions. Final source, mutations, student review and acceptance remain pending.

## Initial formal review — correction required

Independent reviewer, October11, final submission `6194918c2f9dac312263ce35ed644d3bea5d7a99`,
main tree `0ba40790660cddb0044e376fedaea3d9e8d3f483`. Read the complete final
student response above the gate, including all35 mutation attempts and their
three explicitly invalid attempts. One confirmed current-scope blocker requires
**correction round1**; no additional paid or original-grader repetition is needed.

**Cache instrumentation assumes serialization which the real runtime does not
provide.** `internal/agent/actor.go:375–384` starts Send in workers; interrupt
completes the old turn and permits a new one before its old worker has joined.
`internal/cachelens/lens.go:18` incorrectly claims actor serialization. Its
request observation reads/writes `previous`, prediction and rotating artifacts
without protecting this real overlap. An isolated exact-source probe gates the
first actual diagnostic write, interrupts through the public Agent, and starts
another actual request. Both requests complete as expected, but race detection
reports concurrent writes at lens.go:151 and reads at:85. Twelve bounded cases
finish in0.309s. The20ms scheduling allowance is not a deterministic scheduling
proof; the recorded runtime race stacks are the observed defect. A serial gate
control and the student's complete focused TestCache suite pass under race
(1.147s).

The same state assumption also misattributes provider grants even without a
data race. A separate channel-gated actual HTTP-frame observation holds the old
cold response; interrupt returns, a second real Ask completes, then the old frame
is released. Request predictions are0.0000 then0.8554. The second20-token grant
correctly logs0.8554, but the old10-token grant incorrectly logs that same0.8554,
not its own cold prediction. This deterministic check fails in0.016s. Merely
protecting shared state from races does not establish request attribution.
Preserve interrupt/new-turn behavior and the distinction between auxiliary
accounting and main-request diagnostics; fix the instrument's ownership of each
request's observations, with meaningful overlap and restored-defect checks.
No particular API shape, global lock or new feature is prescribed.

Exact probe sources, stdout/stack, source provenance and commands are preserved
in [reviewer-overlap.json](../../solutions/evidence/ch18/reviewer-overlap.json)
and its linked sibling `reviewer-*-probe.go.txt`/`reviewer-*.txt` artifacts. Only
external HTTP/log/observation boundaries are gated; the real Agent, Engine,
parser, interruption and cache lens execute. Student production is untouched.

The original0/100 remains an actual failure on7436fcb. Independently accept the
missing-type fixture explanation and the unreached artifact/save path mismatch
as narrow compatibility exceptions, not evidence that the masked behaviors
passed. Undefined `think` is unused here and is not an additional observed
failure. The final verified Sonnet5 row resolves the earlier named-model price
coverage concern. Actual live and deterministic/mutation evidence independently
cover the masked obligations except the newly found overlap: request bytes and
markers, vendor normalization, derived per-model/tier cost, restart/reset, and
GUI values. All twelve live turns retain7436fcb provenance; no restart or final
source live run is implied. Unknown models remain honestly unpriced. No further
feature or grader accommodation is requested.

### Comparative source, comments and tests

Compared final student6194918 with its own accepted predecessor
`edition-2-ch17-r1`, and matching first-edition `solutions/ch18` with `solutions/ch17`.
Counts are changed tracked source-file totals, excluding blanks/comment-only lines
from active counts; embedded prompt strings remain active. Package moves cancel
in net totals. Logs, generated output, binaries, manifests, module metadata and
external review/live scripts are excluded. The restored `cmd/virtual-user`
example is separate. [Measured counts](../../solutions/evidence/ch18/reviewer-comparison-counts.json)
preserve physical totals and method; these are not measures of compiled size.

| Own-predecessor net delta | Active | Comment-only | Physical |
| --- | ---: | ---: | ---: |
| Student production Go | +400 | +79 | +506 |
| Student production JavaScript / HTML | +8 / +1 | 0 /0 | +8 /+1 |
| Student Go tests | +402 | +24 | +437 |
| Student browser test JavaScript | +42 | +2 | +45 |
| First-edition production Go | +791 | +716 | +1653 |
| First-edition restored virtual-user Go example | +128 | +29 | +179 |
| First-edition production JavaScript / HTML / CSS | +169 /+68 /+128 | +53 /+5 /+15 | +252 /+76 /+164 |
| First-edition Go tests | +834 | +241 | +1206 |

The original additionally adds a four-line SVG. Its chapter delta includes moving
context reduction, event logging, journal/save/segment behavior out of common,
settings storage into its package, and skill registry behavior into skills;
it also changes display identities, replay/reset and GUI structure. These are
not all cache implementation costs or new requirements for this student.
The student retains its already established separation and identity behavior.
Its189-line lens,79-line marker helpers and65-line accounting file replace the
original's separate158/278/238-line canonical/diff/lens components and common
usage behavior with a smaller current-purpose design. Smaller alone is not the
verdict: the overlap finding shows exactly where a simplifying assumption failed.

The new composition uses the actual Engine→Agent→Ensemble parent chain; root
constructs/attaches the lens, common holds contracts/data only, Engine owns
session/model/tier counters, and the separate GUI reaches its public Agent's
Engine. No mutable global pricing/usage registry, callback service bag, SDK,
mock, future model-switch UI or pricing service was introduced. The original
lens takes a directory/report callback; its GUI meter takes a usage source and
model callback, and its common package contains usage behavior. The student
avoids those architecture compromises. Retaining old prices in a single current
model accumulator would be wrong; the small request-model/tier bucket preserves
attribution without adding future model-switch machinery.

New Go explanatory comments are about16.5% of the net nonblank delta (79/479),
close to the teaching target without padding. They explain byte preservation,
ephemeral exclusion, owner paths, full cache-write rates, tier choice, unknown
prices and process lifetime. New exported aliases, interfaces/methods, lens
methods and pricing/meter/model fields have useful comments. The one inaccurate
serialization claim is part of the required correction above. The original has
richer narrative diagnostics and tests, but that verbosity and its broader
refactors need not be reproduced. The student’s compact browser support test is
not counted as a production simplicity improvement.

The original protects append-versus-edit, marker stripping of indented payloads,
Gemini dialogue naming, breakpoint wire survival, pricing arithmetic and implicit
versus explicit eligibility diagnosis. Student tests use actual rendered requests
and real Agent/Engine/HTTP paths, cover current supported vendor contracts,
compaction/ephemeral boundaries, price tiers and mixed-model accounting, actual
save/load session reset, external GUI WebSocket snapshots and actual Chrome DOM.
Auxiliary exclusion exercises the real judge and compressor request path (the
fake compressor's failed submit contract limits compression-success claims).
Existing provider/replay tests remain; source-reading confirms no mock replaces
prior runtime behavior. Browser DOM assertions alone do not prove provider cache
grants; the separately verified live artifacts do.

Inspected all mutation result records:32 intended failures out of35 attempts;
the auxiliary mutant's earlier main-prerequisite failure and two `go test node`
setup errors remain invalid, with later correctly targeted failures. They supply
useful sensitivity for their listed cases, not proof against the newly reproduced
interrupted-request overlap. Independent final-source `TestCache` and serial
control passed; no broad unchanged suite or paid run was repeated. The only
formal correction requested is the one lens lifetime/attribution finding.

## Correction round1 — accepted with recorded exceptions

Independent reviewer, October11. **Accepted after one correction round** at
`afc7dd740b47543cfa0a664ddbe0adba841c25bf`, main tree
`2ee06bb6189898809d80de97df2b73c695cfb0d5`.
Read the complete updated student response and actual correction/mutation outputs.
No remaining behavioral, architecture, scope or documentation finding.

The correction locks only lens comparison and artifact rotation, not the HTTP
exchange or Agent lifetime. Each Send retains its returned prediction in its
own response callback; ObserveUsage no longer consults mutable latest-request
state. This resolves both the data race and attribution error with five net
active production lines, five comment lines and ten physical lines. Ownership,
auxiliary exclusion/accounting and interrupt semantics remain intact. No new
framework, request registry or protocol was added. The original misleading
serialization comment is corrected.

Re-ran all three **unchanged** independent probe sources in an isolated archive
of `ab78b7585cbbb6577e8aee9f71eae94d45788193`, main tree
`4d5fc7941d8644511ee2e778c64f3219e97ecd51`: race-enabled overlap, serial control and
attribution PASS in1.399s. Old cold grant now reports0.0000 while the later grant
retains0.8554. Independently ran the student's two focused overlap regressions
under race: PASS1.331s, including actual file rotation and counted late usage.
A first combined launch failed compilation because both unchanged reviewer and
student-adapted tests used `cacheGateWriter`; retain this setup error, not a
behavioral failure. For the independent run only the colliding student test file
was excluded in scratch; then it was restored and reviewer files excluded for
the separate student run. Product source was unchanged throughout. See
[round1 independent receipt](../../solutions/evidence/ch18/reviewer-round1.json)
and its output files. Scratch source is removed and no reviewer process remains.

The student's three restored-defect results are meaningful: removing locking
reproduces the actual race, a locked shared-latest prediction fails attribution
at0.8554, and removing rotation fails the prior-file check. Their positive control
passes. Inspected core/GUI tests/vet and focused cache race receipts; no broad
unchanged suite, original grader or paid run was repeated by this reviewer.
The new regression requires the old grant to appear and both usages to be
counted, so dropping a late observation cannot satisfy its attribution assertion.

During round1 reassessment I requested attached documentation for the new
exported test Writer and two tests. Final afc7dd7 adds exactly five useful comment
lines to that test file; inspected the exact diff without redundant behavioral
execution. This is a follow-up within round1. Final own-predecessor production
Go delta is405 active/84 comment-only/516 physical lines; Go tests are547/35/598.
Other language and first-edition comparison counts above are unchanged. Useful
production explanation is about17.2% of the net nonblank delta; no padding or
comment removal was needed for acceptance.

Acceptance includes the independently justified original-grader fixture/path
exceptions above. **Actual original score remains0/100 FAIL on7436fcb**; all seven
checks were startup-masked, not passed. No runtime obligation is waived: retained
real-runtime, mutation and independently inspected three-provider evidence cover
cache bytes, placement, accounting, session reset and GUI display. Twelve successful
live turns also remain on7436fcb; they do not claim a live restart, the later
Sonnet price-row exercise or the concurrency correction. Earlier invalid mutation
attempts and initial independent failures remain visible.

Repository continuity observation: initial review commit0c20fc1 is not an ancestor
of current afc7dd7, while its review/evidence bytes remained on disk. The current
history includes CodeRhapsody's dbe4970. No cause is inferred. Re-record only these
owned review/evidence paths on current history using an isolated index; do not
reset, restore unrelated paths, cherry-pick, rewrite history or move tags.
Coordinator owns the accepted direct Edition2 ch17 export and next release;
original22/24 gates and stopped authorship remain. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.
