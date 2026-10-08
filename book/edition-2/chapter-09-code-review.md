# Chapter 9 independent code comparison

October 8, 2026. The historical comparison is complete. Both scoped quality
revisions requested below are now accepted; the final disposition follows the
preserved initial findings. This review found no additional Skills
correctness blocker. It does not replace the separate final live audit or
manuscript proofread, and it does not turn a partial provider attempt into a
successful task outcome.

## Boundary and identities

The initial student source, real runs and teaching experience were frozen at
`786ff239ddf337a5ebb39a12f6552d047f52c1ab` before historical comparison was
authorized. The reviewed runtime is
`75a72554bbb13f5e86c8b806cff780b5fcd79c22`; the 134 runtime files in its
stream binding also match the initial freeze and the working runtime at this
review. Later support changes do not retroactively change those identities.

The reviewer is `/root/grader_ch05`, which engineered independent acceptance
checks and reviewed earlier chapters. It also authored the disclosed isolated
Chapter 6–8 empty-argument maintenance correction. It did **not** implement the
Chapter 9 student runtime. Earlier grader, runtime-review and maintenance
exposure is therefore disclosed; this is an independent comparative review,
not a fresh cold-student attempt. The student remains restricted to new teaching
and receives reasons and expected outcomes, not historical answer code.

The corresponding first-edition standard is `solutions/ch10` at
`04c6e0255e4a1a24149cc6ad72c5e494ff187227`, following its introduction at
`846bc8bb12f90ac2826d286c15efad936815bd03`. The comparison read the complete
old Chapter 10, its public construction/skill methods, parser, registry,
variable renderer and tool integration. Later evidence includes Chapter 15's
manual-retention correction, the matching grader correction
`81a9fcd58a86d6ec2786d4a518994a01dbb553b9`, the move of Skills into its own
spoke at `98878ebb21609e716444595faf930706dc2d32f6`, and Chapter 21's
documented primary-skill integration trap. Those later lessons inform the
comparison; future MCP/cache/compaction features are not requirements here.

The entire coding skill, architecture, new Chapter 9, voice guide and writing
procedure were read. Source identities, the exact current teaching bytes and
historical files are bound in
[the comparison receipt](checkpoint-evidence/ch09-code-comparison-identities.json).
The reproducible identity-only command is:

```sh
python3 scripts/edition2/ch09-review-comparison.py \
  book/edition-2/checkpoint-evidence/ch09-code-comparison-identities.json
```

It validates all original runtime identities before writing the derived receipt.
It is not another acceptance checker. No Go build, provider request, credential
read, whole-tree copy or broad gate repetition was performed for this comparison.

## R1 — Avoid full historical inspection when reading current grants

**Requested quality revision; medium priority.** In the reviewed source,
`skills.go:29–40` implements `toolAgent.GrantedTools` by calling
`a.skills.Inspect()` under `Agent.mu`. `Registry.Declarations`
(`internal/tools/registry.go:88`) and `Registry.Kind` (`:241`) use this path
for model declarations and each tool admission. `Agent.SkillState` also calls
the full public inspection path. Actor management reads and durable publication
use `SkillView().State` in `internal/llm/actor.go:500` and
`internal/llm/watch.go:243`.

`Service.Inspect` (`internal/skills/skills.go:399–425`) copies every retained
activation's slices, sorts all material records by ID, and scans that history
again for each effective tool to derive contributors. Retired manuals remain
in that history after reloads. A current-grant check therefore allocates and
does historical work even when the active graph and granted names have not
changed. The lock also covers that unnecessary work. This is a code-path
finding, not a measured latency claim: immutable body strings are shared, not
copied byte-for-byte, and no benchmark result is asserted.

Provide narrow owned reads through the existing owner chain. Reading grants
needs only the granted names; reading state needs the required state fields,
including its retired summaries, but not material records or contributors.
Keep full `InspectSkills` for the explicit historical inspection use case.
Maintain actor ordering, nil no-skills state and owned returned values; do not
add a second mutable authority, a Registry callback or an internal call back
into the actor's own public mailbox. No public identifier spelling is required
by this review.

Validation should compare equal active grants with short and long retired
history, establish that grant reads do not allocate/traverse material history,
and preserve caller-mutation isolation. A state read may legitimately scale
with its required retired summaries. Full inspection must still return all
retained material and correct contributor IDs. The affected public/watch and
held-HTTP admission checks remain useful; no new paid run is needed for this
internal read-path change.

## R2 — Report every public Agent's terminal result, including failure

**Requested example/usability revision; low priority.**
`examples/skills-consumer/main.go:88–101` checks `Prompt`'s error before
encoding its result and returns immediately on the first failure. The public
`Prompt` implementation (`ensemble.go:681–683`) deliberately returns the
owned partial result and usage alongside a terminal error. The example throws
that information away. If the first Agent exhausts its bound, the second
independent Agent is never asked; if the second fails, its machine-readable
terminal result is absent.

Emit a clearly identified terminal record for each attempted Agent, including
its outcome/error and available result/usage, and let the other independent
Agent finish its own bounded demonstration. Return an overall failure after
those results when any turn failed. Existing public request handles expose
precise completion outcomes if the example needs them. Do not retry, increase
the bound, silently label a partial result successful or make tools opt-out.

A local controlled first-Agent `round_limit` followed by second-Agent success
is the distinguishing test: two terminal records, preserved partial data,
unchanged per-Agent request bounds and nonzero overall exit. Also retain the
all-success case. There is no reason to buy another model answer to validate
this reporting change. Existing live originals keep their original executable
and source identities.

The Anthropic P attempt is specifically **not** evidence of a Skills-library
failure. Both Agents exercised the required independent states, different
ceilings and manual delivery. Beta exhausted its two-request bound and did
not finish its requested prose report. The separate
[partial public-attempt review](chapter-09-public-attempt-review.md) preserves
that distinction. Improving the example must not relabel that original run.

## What the new implementation improves

The first-edition registry mixed parser, graph mutation, variable callbacks and
mutable entries in `common`. Its public setup required ordered calls to
discover, load and wire tools; the tool layer held direct registry/variable/log
references and a configuration-update callback. The current constructor owns
the complete setup, while shared data and parent interfaces remain in common.
`internal/skills.Service` owns parsing, resolution and material; its
`SkillAgent` parent reaches Agent configuration and Ensemble diagnostics. Actor
reaches operations through its Agent adapter. Registry retains schemas and
ordinary decoding. Inspection of actual imports found no new sideways spoke
dependency; the optional GUI and the headless example use public interfaces.

The new candidate/ledger split is a substantial improvement over recursively
mutating dependencies as the old registry walked them. Cycle, grant, offered
name, expansion and whole-group counter checks precede publication. The
candidate retains its creating service and previous-ledger identity; the root
appends the fact before applying that candidate. A later result-write failure
does not pretend to reverse committed authority. Typed public changes use the
same Actor, and a held request retains its capture while the next tool admission
uses current grants. An admitted job remains owned by Jobs after revocation.

Installed schemas, active permission and remembered instructions are separate
facts. The old Chapter 10 delivered the manual inside `load_skill`'s tool
result; the later Chapter 15 correction had to move it out to survive clearing.
The new reducer starts with independent material entries, preserves activation
identity through retirement/reload, and merges hints/manuals behind a complete
call batch. Request captures omit the duplicated primary body and reconstruct
from its recorded material. These decisions address the historical defect
without importing later cache or compaction machinery.

Strict source parsing and bounded reads replace permissive truncation and
silent skipping. The new parser preserves body bytes, the graph is deterministic,
variables are copied scalars expanded once over a complete candidate, and
unknown fields fail rather than appearing to configure an absent feature.
The later Chapter 21 primary-skill trap is a useful reason to retain that
strictness. Do not restore old callbacks, trimming, lazy-unload ambiguity or
permissive unknown frontmatter for superficial compatibility.

Replay checks derive a legal next graph and identity sequence instead of
trusting a plausible replacement state. Historical material remains usable
without reopening the catalog; live append additionally proves the actual
frozen candidate. The initial raw-import defect and live-discovered empty-delta
stream defect were repaired before this comparison, with their failed attempts
preserved in the separate bound reviews. They are not newly discovered defects
or evidence that the initial attempt was entirely green.

The browser addition reuses the existing parent-owned Artifact, Connector and
speech paths. Activation keys include Agent identity; retirement labels come
from current state; manuals use the safe renderer and explicit speaker action.
Known protocol counters retain exact integers without rewriting manual text or
arbitrary tool payload. These are focused extensions to an optional client,
not a duplicate skills engine in the browser.

## Comments, teaching and disposition

Keep the comments explaining actor-only adapters, immutable candidate records,
the bounded physical-line reader, raw framing-byte accounting and protocol-only
counter conversion. They explain constraints that are otherwise easy to remove
during cleanup. The narrow common JSON decoder methods are standard-library
dispatch over shared types; they are not an excuse to move graph behavior into
common, and their present location is appropriate.

The new chapter's installed/enabled/remembered distinction, ownership plan,
candidate walkthrough and separate manual entry are clearer foundations than
the old setup sequence. The two improvements above follow those printed
responsibilities; they require no new owner, numeric ceiling or wire rule.
R1 is an implementation-economy issue, not missing teaching. R2 gives the author
a useful explicit lesson about bounded partial work and independent Agents;
the original Anthropic outcome must remain visible in the final demonstration.

The comparison requests review of those two scoped revisions and affected local
evidence before its final code-quality disposition. It does not reopen already
accepted deterministic checks or authorize paid retries. Final live acceptance,
the author's actual-run reconciliation and a complete final manuscript
proofread remain separately recorded gates.

## Final revision review — R1 and R2 closed

The reviewed repair is `06c6787337d959c6d9b4874ee660025fe432a72a`. The reviewer
read its complete production delta, both narrow-read tests, the consumer's
subprocess test and all five comparison JSON receipts plus the revision
handoff. The coding skill was reloaded in full. Independent identity checks
confirm all 137 revised runtime/support-source files against that commit and
the current tree, all 13 affected-module source hashes, all 123 Go/module
hashes in the selected public receipt, and its three checker hashes. The
[revision review receipt](checkpoint-evidence/ch09-code-revision-review.json)
records those bindings and the precise reuse of student-run results. This
review did not rerun the tests or make a provider call.

R1 is resolved without a cache or second authority. `Service.GrantedTools`
copies only current names; `State` copies the required safe state. Full
inspection still performs historical material/contributor work when explicitly
requested. Public state uses an Actor query, while Actor management and watch
paths use the non-enqueuing Agent adapter under the existing lock. Nil
uninitialized/no-skills behavior, owned slices, offline state and full inspection
are preserved. The tests use real repeated transitions and distinguish short
from long retired history. Their observed allocations are 1/1 for grant reads,
543 for the long full service inspection, and 8 versus 64 for public state
versus inspection. These are the recorded local observations, not general
latency promises. The material-map removal is honestly labeled an owner-test
seam; it is not presented as a valid historical log.

R2 is resolved through public `Submit`/`Wait` completion handles. Each attempted
Agent emits its identity and actual completion, including available text,
parts, usage and error. A failed first turn does not suppress the second
Agent's attempt; joined errors produce overall failure after reporting both.
The two-request policy and prompt are unchanged. Waiting's deadline is treated
as a wait error: the example cancels and collects the actual request terminal
receipt, rather than inventing a successful outcome.

The consumer control invokes the actual command entrypoint in a child test
process. Its all-success parent emits two success records and exits 0. The
distinguishing case emits `round_limit` then `success`, retains the first
Agent's partial text and parts and each Agent's usage of 20 input/4 output
tokens, makes exactly two local HTTP requests per Agent, verifies both durable
policy captures remain 2, and exits 1. These are local fixture tokens and
responses, not additional live provider evidence.

Both fixture-development corrections are justified. The pre-existing Actor
completion contains the latest accepted response's parts, not every earlier
response's text. Requiring the earlier partial text in a successful final
completion was therefore wrong. The corrected control still requires the
failed Agent's available partial text/parts and the successful Agents' actual
final text, along with usage and outcomes. The original allocation overlay
removed an uninitialized-state nil guard and failed during construction; it
proved no allocation property. The corrected overlay preserves that guard and
fails on the intended history-dependent allocation assertion. The student
handoff describes those original console failures; this review does not claim
independently verified raw transcripts for those two development attempts.

The retained positive commands pass, followed by four specific overlay failures:
history-dependent grants (18/61 allocations), public state rebuilding full
inspection (64/64), returning before either terminal record, and erased usage.
The last combined mutation clears text/parts/usage and is caught first by the
usage assertion; it is not counted as three separate mutation proofs. The
positive test checks the other available fields, and source inspection confirms
the complete terminal value is retained. All overlay source hashes and unique
replacement anchors match the repaired source; none of the four credited
failures is a compilation error or panic.

The two affected modules passed vet/tests and formatting produced no filenames.
The seven selected public groups passed with race detection, including owned
snapshots, two-Agent/no-skills behavior, paused publication, concurrent single
commit, held-HTTP revocation, admitted-job survival, and anchored material/hint
replay. Their receipt remains explicitly a selected suite, not another full
67-row gate. These results and the inspected narrow delta are sufficient for
the requested quality revision.

**Code-quality disposition: accepted; R1 and R2 closed.** Original live runs
retain their original source, binaries and limitations. In particular, the
Anthropic P attempt remains partial with its missing beta final report; the
new reporting format is validated locally and is not attributed to that old
executable. Current-grant read economy does not change provider semantics.
No paid repeat is required for either repair. The separate live audit and
final manuscript proofread retain their own acceptance scope.
