# Original24 / Edition2ch22 opening stage: reviewer preparation

Reviewer `reviewer_astra`, GPT-6 Astra/high, October 11, 2026.
**Accepted with three narrow grader-fixture exceptions at initial review;
zero formal correction rounds. Intermediate original24 checkpoint only.**
The final disposition below supersedes historical pending-review statements.
Mandatory `.agents/skills/ensemble-coding/SKILL.md` applies. Full chapter/read
coverage and independent merge agreement are recorded in [ch22 review](ch22-code-review.md).
No original answer, provider or grader execution was used for this preflight.

Inspected launch/public-consumer contracts in `internal/grade/ch24_harness.go`,
revision `6896df7faca8b049f80293c19a944436e1656007`, SHA-256
`5631a794d043a79ac43b22178666fcd5d9e6b797ddb195afe4741c3adee10de5`.
Wire compatibility reuses original22's GUI/stdin/model-switch harness. Embedded
deployment launches `--port PORT` from a fresh unrelated cwd, with the same
LLM/skill environment inputs as22 and no disk asset override. The core-only
entry and legacy terminal-envelope limitations therefore remain relevant.

External probes assume one module with `<base>/gui`, a `gui.AgentHooks`
callback bundle and `gui.New`; the temporary module supplies only one root-module
replace. The probes spell `Assets.ReadFile("web/artifact-scroll.js")` and
`web/renderers.js`, `ServeWS`, `ServeMCP(net.Listener)`, and a specific root
event/log vocabulary. The supplied outside consumer mounts only the component.
Its WebSocket sends `type:subscribe`, then `type:prompt` with text, and observes
`type:message` with text. Its bare TCP MCP relay sends a JSONL `tools/list`
request and expects a JSON-RPC result/error with the same id before any browser
connects. These are observations of the harness, not required implementation
names or authority to introduce callback bags.

Both external programs build together: an incompatible GUI probe can prevent
the independent headless dependency checks from running. The checks also assume
`./cmd/virtual-user`, `<base>/mcpws`, and literal root `func NewAgent(` as a
positive control. Fixed package/module/constructor assumptions cannot override
Bill's separate optional GUI module and actual-parent architecture. Review any
actual failed result independently rather than pre-awarding exceptions here.

Preserve all valid obligations: outside consumers construct public inputs and
exercise replay/prompt/relay paths; components serve without the application
shell; embedded deployment and explicit disk override work; core/MCP transitive
dependencies exclude GUI/WebSocket implementation with positive controls; the
actual model switch survives unchanged. No additional feature is authorized by
the grader's reference-shaped fixture. Original24 remains separately reviewed
before25, with only the accepted combined source exported as Edition2ch22.


## First-green comparative preflight

Inspected frozen student `35df549210130f824fdc2a4eb632ea878b2c529b`, main tree
`5b36f28cbc8bac320febe1a9afa810efd27452e3`, after reading the student's stage
review. Only then read matching first-edition implementation, own-predecessor
changes and relevant tests. Final student review/live/mutation assessment remains
pending; no acceptance or formal correction round is claimed.

Actual original grade is **0/100 FAIL**,20.89s, no external timeout. Four checks
stop while assembling fixed root/same-module GUI probes; two launch the headless
CLI with GUI flags and never get a GUI port; the root scan's literal free-function
NewAgent positive control misses the actual public composition-root method.
These match the earlier sourced preflight. Valid obligations remain required;
no exception is awarded merely from these fixture diagnoses.

Independent exact-archive `TestExternalReuse` passes under race detection,
including the separately compiled outside-module tests: public event construction,
replay and prompt through real Agent/Engine with an external HTTP fake;
embedded/disk asset selection; actual Chrome component rendering with only two
modules served and no shell; same-ID concurrent relay clients with reverse-order
replies alongside the core handshake; shutdown; actual headless/GUI executable
transitive dependency negative and positive controls. See
[output](../../solutions/evidence/ch24/reviewer-first-green-external.txt).
The archive only links the already installed node_modules as test dependencies;
its implementation bytes come from the frozen commit, not the moving main tree.

An additional small real-network probe passes under `-race`: after a relay
request reaches the browser, TCP EOF removes the parked request; its late browser
reply does not block the next request; browser EOF returns an error preserving
the second caller's original ID. It uses actual Ensemble/Agent/GUI and external
socket peers, with no model call or credential access. Retained
[probe](../../solutions/evidence/ch24/reviewer-relay-disconnects-test.go.txt)
and [output](../../solutions/evidence/ch24/reviewer-relay-disconnects.txt)
check the new advertised lifetimes without editing student source.

Provisional own-predecessor counts are in
[receipt](../../solutions/evidence/ch24/reviewer-counts.json).
Student net production adds216 physical Go lines (187 active,20 comment-only,9
blank) and8 JS (4 active,3 comments,1 blank); tests add351 Go and38 MJS.
Original net production adds24 Go (13 active,11 comments), with no JS growth;
tests add144 Go. Complete before/after totals and lexical method are recorded.
Moves are counted once: the original moved its existing GUI/WebSocket/assets;
the student extracted existing renderers/component from app.js. The original
already had its TCP relay and standalone component, explaining much of the
student's additional Go growth. No core production change or fleet capability
appears in the student's delta. Final KISS/comments/tests disposition waits for
completed evidence; no blocking source finding has emerged at this checkpoint.

Display transcripts normalize trailing whitespace only.


## Final independent disposition

**Accepted with exceptions**, original24 / opening stage of Edition2ch22:
`0430df4340fffac5234a94aea843f9a7c9d02282`, main tree
`8555e3cb9ab5200f0a1c52691007a7fc4a1c6797`. Read the complete final
[student review](ch24.md), final delta and mutation records alongside the source
and comparative inspection above. Mandatory `.agents/skills/ensemble-coding/SKILL.md`
read in full. Independently verified all164 final manifest hashes. There are no
unresolved source, documentation, ownership, scope or bloat findings. Zero formal
correction rounds; the student's test repairs occurred before formal submission.
Do not export an accepted combined Edition2ch22 yet: original25 must first be
implemented and independently accepted.

### Accepted explanations, actual grade retained

The unchanged original grader remains **0/100 FAIL** on `35df549`,20.89s,
make2/grader1, without external timeout. Independently accept these three fixture
exceptions, agreeing with the complete student review for these specific reasons:

1. Four checks never reach their assertions because one assembled external
   fixture imports the root executable and a same-module GUI. The actual public
   core is `ensemble/ensemble` and GUI is the separate optional module required
   by Bill. Neither a callback bundle nor a same-module compatibility facade is
   an appropriate repair. The actual outside-module consumer constructs public
   events, drives real replay/prompt/relay behavior and checks both dependency
   closures. The old `mcpws` pathname is not a capability: this implementation's
   existing virtual user uses a browser peer over the public raw MCP transport,
   with its browser implementation outside core. That actual consumer still runs.
2. Two checks launch the headless executable with GUI flags. The actual optional
   GUI executable serves embedded assets from an unrelated cwd, including through
   the live human browser. Model selection still reaches the same actor at the
   turn boundary; unchanged source, current deterministic tests and accepted
   original22 live/model-switch evidence support reuse of that result. A fresh
   paid switch solely for asset extraction would add no relevant observation.
3. The root scanner's free-function `NewAgent` positive control cannot recognize
   the public Ensemble method. Both shipped commands still use the actual public
   root, and the external headless MCP consumer plus actual command dependency
   checks establish the intended GUI/WebSocket exclusion, with GUI positive
   controls. No special free-function wrapper or root relocation is needed.

These waive fixture layout and shape assumptions only. Public reuse, component
independence, relay correlation, deployment/override, event vocabulary, headless
imports and actual model-switch behavior are retained obligations. No original
source or grader was edited, no failing check is recast as passing, and no
future fleet feature is included. Later grader maintenance needs both module
replacements and the real optional-GUI entry point, while preserving historical
layout support.

### Final verification and evidence limits

The final source differs from live `35df549` only in a clearer public method
comment, multiline constructor formatting and tests. Independent final
[external run](../../solutions/evidence/ch24/reviewer-final-external.txt)
passes, including its own nested race instrumentation and actual Chrome;
[receipt](../../solutions/evidence/ch24/reviewer-final-check.json)
records exact source,164-file manifest match and command. It exercises the repaired
embedded-shell assertion and browser-disconnect/shutdown cases. The separate
reviewer client-EOF/late-reply probe above remains valid for unchanged behavior.
Core and GUI tests/vet, complete GUI race suite, renderer/speech browser checks,
virtual-user/peer lifetime checks and formatting all have retained passing results.
The initially empty race selection and missing-checksum/copied-memory setup
failures are correctly excluded from evidence of working behavior.

The independent [live receipt](../../solutions/evidence/ch24/reviewer-live-check.json)
checks binary SHA, all3 actual request/completion pairs as GPT-6 Astra/high,
matching provider and raw GUI results for write_file/read_file, exact42-byte disk
artifact, and the actual assistant final in the DOM. The full embedded GUI ran
from fresh cwd with no source web directory, all7 selected asset requests returned
200, no page errors occurred, and exit0 followed9.649s. A DOM match to the human
prompt alone would not have established success; the final and disk checks do.
This demonstrates full-GUI deployment; component-only and new TCP relay behavior
have separate real browser/external-peer tests. No second model client using the
new relay, new paid model-switch run or network-endpoint interception is claimed.
Reviewer read no credential profile and made no provider call.

Reviewed the18 initial mutations and2 corrected repeats, including actual failure
locations and setup. The initial embedded-default mutation **survived**: a cwd
listing returned200 and defeated the status-only control. Requiring the actual
shell title now kills that same mutation. The initial relay-client-close mutation
reached its intended assertion but then timed out during cleanup at45s; it is
not a clean original kill. Explicit client defers fix cleanup ordering, and the
repeat exits in4.28s with the intended `relay shutdown stuck` assertion and no
cleanup timeout. Preserve both corrections rather than counting20 independent
proofs. Speech-owner removal times out at the actual pre-final speech observation;
prompt/relay timeout failures likewise occur at the specific withheld result,
with successful baselines. Import mutants compile with the cached dependency
supplied, then fail the intended transitive check, not module setup.

Those checks protect the current mechanisms; they do not establish exhaustive
protocol correctness. Client EOF/late reply is additionally checked by the
independent real-network probe; the student's shutdown mutant alone is not
claimed to prove every cancellation interleaving. A silent connected browser
has the documented lifetime bound of browser/client/server closure, not an
invented per-request timeout. Public export removal is a compile boundary, not
a claimed behavioral mutation kill. No mocks or generic verification framework
were introduced.

### Final comparative assessment

Final own-predecessor deltas, physical / active / comment-only / blank:

| Category | Student | Original |
| --- | --- | --- |
| Production Go | +220 / +191 / +20 / +9 | +24 / +13 / +11 / +0 |
| Production JavaScript | +8 / +4 / +3 / +1 | 0 / 0 / 0 / 0 |
| Test Go | +376 / +356 / +14 / +6 | +144 / +108 / +24 / +12 |
| Test MJS | +38 / +35 / +2 / +1 | 0 / 0 / 0 / 0 |

The updated [count receipt](../../solutions/evidence/ch24/reviewer-counts.json)
includes whole before/after totals and separates tests, languages and support.
Lexical comment-only lines are excluded from active counts; inline code remains
active. Generated IO, docs and evidence are excluded. Moved code is counted once,
not as a new subsystem. Four final active-line additions merely expand a struct
literal for readability; they are not four new behaviors.

This chapter starts from different predecessors. The original moved an existing
GUI, standalone component and TCP relay; the student already had the required
optional-module boundary but lacked the public relay and independent scroll.
Its production growth supplies those assigned capabilities. Extracted renderer
logic stays one implementation used by both shell and component. The relay uses
its existing transport owner, writer and server lifetime, with a per-browser
request map to keep caller collisions and core MCP IDs separate. It selects one
browser, so it needs no broadcast duplicate-response router. No new core facade,
callback wiring bundle, generic event layer or speculative fleet API is present.

The original has richer explanatory blocks, particularly about imports and
request routing, but its hook bundle retains optional dependencies and exported
asset binding. The student retains actual root/Agent access and a private embed
anchor with a read-only accessor. New public Go symbols have useful contracts;
comments explain ownership, independent MCP lifetimes, private IDs, late replies,
EOF cancellation, serialized writes, embedded roots and optional component speech.
New net Go comments are about9.5% of nonblank growth, below the rough20% target;
net delta understates retained/moved explanations and is not a quota. These small
mechanisms remain understandable without padding. The corrected ServeMCP comment
now clearly distinguishes absent and silent browsers. README covers actual
imports, constructor, assets, override, component API and listener lifetime.

Student tests are larger because they build a genuinely outside module and run
the real Agent/Engine plus external peers and Chrome. The original newly added
reuse test is `gui_test` within its module with hooks, which cannot itself prove
the same external-module/runtime boundary; its original grader supplies outside
probes but assumes its own layout. Student assertions now distinguish intended
HTML from a successful directory listing and verify the relocated real command,
not only a standalone handler. Existing rendering/model-switch behavior is reused
with affected checks, not duplicated into a new test system. This is justified
scope and useful failure coverage, not line-count optimization.
