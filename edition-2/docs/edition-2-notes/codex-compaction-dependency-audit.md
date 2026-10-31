# Background compaction: chapter dependency audit

October 14, 2026. Independent dependency reviewer, GPT-6 Astra/high.
Read-only audit of baseline `56ec71cb` in `/Users/bill/projects/ensemble-compaction`.
This note is a recommendation for Bill's placement decision, not acceptance of
new implementation, manuscript, graders or snapshots. No graders or live/runtime
operations were run by this reviewer; no source files were edited.

## Recommendation and decision

Adopt the proposed move of current Chapters 15 and 16 together after current
Chapter 23, provided Bill approves two consequences: stage accounting's
memory/recall-specific obligations at their new introduction points, and produce
new reviewed cumulative solution exports rather than rename the old ones.
The move puts the sandbox, child lifecycle and fleet inspection before the
new background compressor that uses them. Auto-Recall still follows memory.
There is no unavoidable earlier exercise dependency that requires memory or
recall before the fleet chapters.

| Current chapter | Proposed chapter | Material |
| --- | --- | --- |
| 0–14 | unchanged | Through recorded context management |
| 17 | 15 | The Invisible Invoice |
| 18 | 16 | Leaving Anthropic for OpenAI |
| 19 | 17 | The Crossover; still ungraded |
| 20 | 18 | Giving the Agent the Web, and a Status Line |
| 21 | 19 | The Sandbox |
| 22 | 20 | The Ensemble; public reuse and child lifecycle |
| 23 | 21 | Managing the Swarm |
| 15 | 22 | Forgetting on Purpose; revised background implementation |
| 16 | 23 | Auto-Recall |
| 24–26 | unchanged | Optional route, reserved crossover, Gateway |

The narrow approval question is whether Bill wants this mapping with staged
accounting and newly validated cumulative exports replacing the current
book-facing `solutions/ch15`–`ch23` references, while old source/manifests/results
remain reachable through immutable checkpoints. Carry the revised cumulative
implementation into Chapters 24 and 26 and validate their affected behavior too.
This is reviewed maintenance/reconstruction, not a fresh blind student experiment.
No new blind build is implied by the current maintenance assignment.

## Material dependencies

1. **Accounting has real forward obligations that must move, not disappear.**
   Current `book/chapter-17.md:49` includes compressor/judge usage in Engine's
   accumulator, with separate latest-main counts; `:70` excludes those requests
   from main-request cache comparisons. `:88` requires the surviving-memory
   fallback after a handoff is compressed. `:425`, `:433` and `:447` explain recall
   clearing, folded boundaries and expected prefix rewrites. Preserve main-request
   accounting, stable-prefix inspection and Chapter 14 handoffs in new15. Complete
   the memory-boundary and compressor attribution contracts in new22; complete
   judge attribution and recall integration in new23. This changes their teaching
   stage, not their eventual requirement. Current `book/chapter-20.md:116` and
   `:935` similarly need staged wording in new18: session versus latest-main usage
   works before auxiliary requests exist and must remain correct after they do.

   The numerical invoice grader does not currently prove all those stronger
   claims: `internal/gradeedition2/ch17.go:257` explicitly states that its marker
   fixture has no ephemeral tail, handoff or memory. Separate implementation
   tests cover them, including `agent/ensemble/cache_test.go:193` (surviving
   memory), `:275` (auxiliary model identity), and `:342` (unpriced auxiliary
   model). Preserve or adapt these later-stage obligations; do not claim an
   unchanged invoice score proves the new child accounting path.

2. **Authentication has bounded explanatory references, not a required memory
   API.** `book/chapter-18.md:258` describes the cuts/memory/recall of current
   14–16 as existing context-mutation controls. In new16 retain the already
   implemented Chapter 14 controls and point forward to new22/23 for memory and
   recall. The authentication, billing-route and Responses exercises do not
   need those features to precede them. Preserve dated historical claims and
   demonstrations as history, not fresh runs of the reordered solution.

3. **Sandbox and fleet must precede the new compressor.** Current21 supplies
   child confinement; current22 supplies actual child construction, submission,
   bottom-up completion, messaging and bounded idle correction; current23
   supplies visible inspection, activity, result retrieval and lifetime controls.
   Their order is retained as new19–21. They do not require memory first.
   `book/chapter-22.md:84` and `:93` make factory-enabled submission behavior
   explicit while ordinary embedders retain ordinary final replies. The new
   default internal factory must not accidentally turn every existing embedder
   into a submit-required fleet. Teach the distinction between construction
   availability and public delegation/submission enablement in new20.

4. **The optional submission schema belongs with fleet creation, before memory
   consumes it.** Current `book/chapter-22.md:487` specifies `submit(result)`;
   `book/chapter-23.md:297` specifies result retrieval and durable `submit.md`.
   The current API is explicitly string-shaped:
   `agent/internal/common/fleet.go:38`, `:54`, `:60`, and `:79` expose status
   result, retrieval, spawn and submission. `agent/internal/fleet/management.go:66`
   also decodes a string fallback. Teach and verify optional launch schemas,
   validated completion and structured result preservation in new20/21; new22
   then uses the memory schema. Default text callers and historical saved text
   must retain their documented compatibility. A string convenience method must
   not bypass an explicitly custom schema. This is the authorized handoff's
   scope, not a proposal to redesign all messaging or build a generic workflow.

5. **Recall is a genuine successor of memory.** Current `book/chapter-16.md:5`,
   `:90`, `:100` and `:106` require the memory bands, archive corpus and exclusion
   of content already actually present in context. Moving the pair preserves
   that prerequisite. The judge is expressly one stateless call (`:119`, `:393`),
   so a compressor implemented as a child does not authorize converting the
   recall judge into another Agent. Review its actual interaction with late
   memory commits, context deduplication and turn cancellation when new23 is
   validated. Preserve Bill's quota/all-memory decisions.

6. **Historical discussions are not forward implementation requirements.**
   Current20's constructor table mentions `NewJudge` and `recall.New`
   (`book/chapter-20.md:730`), but `:737` explicitly labels the numbering as
   original. Retain that documented diagnosis; do not mechanically replace
   original chapter numbers or delete the story because its repaired features
   now arrive later. Current14's forward sentence at `book/chapter-14.md:864`
   should point to new22. The earlier historical crossover remains a historical
   demonstration, with its own source identity, when it becomes new17.

## Grader chains, routing and fixtures

There is no recursive memory→recall→invoice parity chain. Current memory calls
`Chapter14` directly at `internal/gradeedition2/ch15.go:20`; retain that context
parity when memory becomes new22. Replacing it with new21 merely because that
chapter precedes it would silently change the check's meaning. Current16,17,18,
20,21 and22 use feature rubrics rather than recursively running the preceding
chapter. Current23 reuses `fleetMessaging` and `fleetIdle` from22 at
`internal/gradeedition2/ch23.go:27`, then checks its own GUI/management behavior.
Current22 composes original24 and25 weights and reuses the actual model-switch
session from20 (`internal/gradeedition2/ch22.go:11`, `:30`); this is preserved
as new20 depending on new18 support, not on memory.

Helpers such as `recallWrite`, `recallReadEvents` and `memoryConversation` are
shared grader utilities, not student prerequisites. Their use in invoice,
authentication, sandbox or fleet drivers does not imply those features must
exist in the submission. The fleet peer explicitly disables recall with
`EN_RECALL_DISABLED=1` (`internal/gradeedition2/ch22_fleet.go:49`). Preserve
external-provider fakes and the real earlier implementation beneath them.

The memory driver assumes serial completion: `internal/gradeedition2/ch15.go:86`
constructs fixed request expectations, and its checks inspect completed summary
results. Adapt these to observable child/candidate completion barriers without
discarding the eleven existing memory weights. Add honest coverage for the
authorized background/schema behavior. A total-request ordinal is unsafe once
parent and compressor requests overlap. Separate the three identities:
ordinary parent requests, submit-only memory workers, and tool-free judges.
Do not identify every tool-free final/terminal request as recall.

The command's numeric dispatch cannot be changed alone:

- `cmd/grade-edition2/main.go:31`–`:38`: help text and chapter-specific options.
- `:41`–`:47`: valid/ungraded chapters and implicit `solutions/chNN` selection.
  New17 is ungraded; new19 is graded. Optional24 and reserved25 remain ungraded.
- `:76`–`:84`: executable routing. Current14,15,16,17,18,20 use the GUI entry;
  after mapping that is14,22,23,15,16,18. New19,20,21 use core CLI and separately
  built GUI probes where needed. Gateway26 retains its own command.
- `:93`–`:155`: feature dispatch and report identities. Keep feature identity
  separate from predecessor arithmetic. Shared `internal/grade` names such as
  `Ch16Checks`, `Ch17Checks`, `Ch24Checks` and `Ch25Checks` identify original
  rubrics and are not wholesale rename targets.

Root reports baseline current14/15/16 grading on explicit `./agent` at100/100.
Root also reports that old22/23 CLI grading stops before checks because
`grade.Build` chooses the present `agent/cmd/` directory without a Go program;
a temporary runner using a CLI built from `agent` obtained200/200 and100/100
from the unchanged feature graders. Those are coordinator-supplied observations,
not this reviewer's executed checks. Preserve the launch failure separately and
repair routing before claiming the final advertised commands work.

## Cumulative exports and historical access

Old17 already contains memory and recall, for example
`solutions/ch17/internal/common/types.go:247`, `:282`, `:559`. Old23 constructs
both services (`solutions/ch23/ensemble/ensemble.go:253`, `:305`). Renaming old17
to new15 would place future features in the reference and would not test that a
reader can follow the new sequence. Its old grade cannot become new-order
acceptance. The same problem carries into old24/26 even though their numbers do
not move.

Recommended concrete approach, conditional on Bill's decision:

1. Keep all existing immutable tags, original-numbered notes, raw evidence,
   manifests and first-edition sources unchanged. Record a dated old↔new table
   and identify which checkpoint supplies each historical reference.
2. Reconstruct cumulative current15–21 in isolated source stages from accepted14,
   bringing in only each stage's capabilities. Then add revised memory as22 and
   recall as23. These may reuse reviewed source; explicitly label the work
   maintenance reconstruction with prior-answer exposure. Validate each stage's
   public exercise and affected inherited behavior, plus independent scope,
   architecture and comparative review. A final full-feature `agent` pass cannot
   establish those intermediate boundaries.
3. Produce exact exports/manifests from committed, accepted staged source and new
   immutable revision tags. Replace book-facing directories only after approval
   and acceptance. Extend/revalidate the changed cumulative implementation for
   optional24 and Gateway26; retain25 as a placeholder without an accepted export.
   No symlinks, fabricated student runs or new history directory are required.
4. Make historical grading explicit through the preserved checkpoint checkout
   and its command, with an explicit submission path. Current grading selects
   the newly accepted mapping. Until replacement exports exist, require an
   explicit submission for affected current chapters or fail clearly rather than
   silently grade the old default path. A separate historical selector in the
   current binary is optional; immutable checkout commands already preserve the
   exact old grader and fixture identity without maintaining duplicate engines.

The top-level preservation rule forbids editing frozen snapshots during normal
development (`.agents/skills/ensemble-coding/SKILL.md:10` and
`book/chapter-writing-procedure.md:38`). The workflow permits reviewed
implementation corrections but rejects silent frozen edits
(`docs/agentic-codebook-workflow.md:2103`). Its explicit earlier replacement
precedent is the combined ch20 export (`:849`), replaced only after independent
acceptance while its old tag stayed immutable. The procedure's historical
revision-export mechanism (`book/chapter-writing-procedure.md:568`) likewise
does not turn renamed source into validation. Bill's approval should explicitly
cover the proposed current reference replacement and maintenance review status.

## Change map for the authorized later stages

- Teaching: `book/chapter-14.md` forward pointer; current
  `book/chapter-{15,16,17,18,19,20,21,22,23}.md` filenames, headings, section
  references, examples, commands and staged contracts; inspect24/26 references.
  Preserve Bill's `book/preface.md` verbatim.
- Navigation and build declarations: `book/chapters.md`, `book/README.md`,
  root `README.md`, `Makefile`, and current mapping statements in
  `docs/edition-2-chapter-map.md`, `docs/edition-2-enhancement.md`,
  `docs/edition-2-publication.md`, workflow/carryover/current instructions.
  Add current clarification above dated history; do not globally rewrite it.
- Grading: `cmd/grade-edition2/main.go`, affected feature families under
  `internal/gradeedition2/ch15*`, `ch16*`, `ch17*`, `ch18*`, `ch20*`, `ch21*`,
  `ch22*`, `ch23*`, plus affected26 and build-selection support/tests.
  File renaming is a maintenance choice; dispatch and published identities must
  agree. Preserve original `edition-1` graders and shared original rubric meaning.
- Public/runtime contracts: `agent/internal/common/{types,fleet}.go`,
  `agent/internal/fleet/`, `agent/internal/memory/`, Agent/Engine request and
  completion paths, `agent/ensemble/`, public GUI result/inspection paths and
  source tests affected by `Engine.Compress` or string submissions. This is a
  dependency list, not authority for speculative APIs.
- References: `solutions/README.md`, approved replacement `solutions/chNN/`
  exports and sibling manifests; preserve old tags and historical evidence URLs.
  `scripts/edition2/export_snapshot.py` exports committed `agent` by default and
  accepts historical prefixes; it creates neither acceptance nor replacement
  authorization (`:5`, `:73`).
- Publication inputs: inspect `scripts/build-ebook.sh`, `book/build-print.py`
  and the generated assembly's chapter order when renumbering is accepted.
  Existing EPUB/PDF/assembled Markdown remain their earlier publication artifact;
  no publication build or upload occurred in this audit.

Reading coverage: full handoff; AGENTS; mandatory coding skill; full workflow,
carryover, voice and chapter procedure; full current Chapters14–23; chapter map
and solution README; command dispatch; relevant feature grader, driver, fixture,
public-interface and frozen-source sections cited above. After one compaction,
the full mandatory skill and all four governing/craft documents were reloaded
before completing this review. This is not a claim to have reread every runtime
file or every grader line, nor a formal implementation acceptance.

## Follow-up: approved mapping and build-selection preparation

October 14, resumed at coordinator baseline `7a362ec2`. The coordinator reports
Bill approved the proposed mapping, staged accounting and validated cumulative
export replacement. The decision paragraphs above remain the original audit;
they no longer represent an outstanding permission gate. New exports still need
their implementation, validation and independent review before acceptance.

The narrow build repair is justified by the existing advertised selection
contract, without changing any chapter obligation. `internal/grade/harness.go:86`
says to find a main package first at `cmd/`, then root; `:104` actually selects
any directory named `cmd`. Independent metadata inspection with
`go list -e -json ./cmd/` from `agent/` returns no package name and the explicit
error `no Go files in .../agent/cmd`. This is a command-container directory,
not the CLI package.

Recommended implementation: ask Go for the candidate's package metadata, select
`./cmd/` when it identifies a main package, otherwise build the root. Keep the
existing single selected build; a selected main package's compilation error
must fail instead of triggering a build of a healthy root. Go's error-tolerant
metadata mode can still identify `main` when dependencies are invalid. Avoid
recursive command discovery or choosing among arbitrary child executables;
Gateway already passes its explicit command directory.

Focused real-module tests are warranted because this selector serves earlier
graders as well as current22/23. Check root-only, a `cmd` container with a nested
command, preferred `cmd` main beside a healthy root, a non-main `cmd` package,
and a selected `cmd` main that fails compilation despite a healthy root. Execute
successful binaries and compare distinct output so a successful build of the
wrong target cannot pass. Preserve the direct-file input behavior. These tests
need neither mocks nor model inference. No existing `internal/grade/*_test.go`
files were found during preparation. This is a strategy review, not approval of
an implementation that has not yet been submitted.

For later reconstruction, use the existing export tool and short source-stage
commits instead of a new generator. Start from14, then validate new15 accounting,
16 authentication,17 historical crossover's bounded integration,18 web/status,
19 sandbox,20 fleet/schema,21 management,22 background memory, and23 recall.
At each stage verify that the new capability is present and later memory/recall
services have not accidentally arrived in an earlier cumulative tree. Carry
the resulting23 into optional24 and Gateway26, preserving their existing route
and Gateway exercises and adding only affected integration checks. Do not make
unrelated live repetitions or an imagined chapter25 acceptance part of this
work. Detailed source selection waits until the behavioral stages and prose
contracts are accepted; the sequence itself is now approved.

## Independent build-selection review

October 14, reviewer disposition: **accepted for the narrow build repair**, with
zero correction rounds and no blocking findings. This does not accept later
submission-schema, compaction or chapter work. Reviewed against `7a362ec2`:

- `internal/grade/harness.go`, SHA-256
  `43f93da9b67504fc9cb69a88572fdc3ebd382463a2d5e31c6703fd422c01fb7c`.
- `internal/grade/harness_test.go`, SHA-256
  `05877258c04632ae5ba4b14eecb1f3d1307b1f37b27c091b85d20851ddcb9335`.

The implementation queries Go package metadata only when `cmd` is a directory,
selects it only for `main`, and keeps exactly one subsequent build. Metadata
failure returns an error and cleans the temporary directory; selected command
compilation failure is never replaced with a successful root build. Direct-file
input and root-only behavior are unchanged. Go, rather than a hand-written source
scanner, determines the package under the effective build constraints. No
recursive command search, new dependencies, chapter assertions or runtime changes
were introduced.

For the `Build` function, counted from its declaration to the following function
comment, physical lines grow31→43 and active lines27→36; active means nonblank,
non-`//` lines. The nine active lines implement the necessary metadata query,
error cleanup and selection. The explanatory comment states why directory
existence is insufficient and why a broken main must remain selected. The five
real-module scenarios execute successful binaries and distinguish their output;
the broken-main case requires the actual undefined-symbol compiler diagnostic.
This is small, behavior-based coverage of the reported failure and retained
selection semantics, without mocks or a test framework.

Independent verification copied the exact submitted `Build` function and
unchanged test file into an isolated temporary Go module, adding only its
standard-library imports. It did not modify shared source or invoke a grader:

| Isolated source | Result | Intended observation |
| --- | --- | --- |
| Submitted implementation | PASS,0.319s | All five scenarios |
| Always select root | FAIL,0.328s | Preferred command executes root; broken preferred command wrongly succeeds |
| Original directory-existence predicate | FAIL,0.186s | Command container has no Go files; library archive cannot execute |

The failures occur at the intended runtime/assertion boundaries, not because the
mutated harness cannot compile. Temporary probe modules were removed. These
mutations establish the selector's tested distinctions, not exhaustive coverage
of every possible Go toolchain failure.

The reviewer also tried a temporary main importing an unavailable module, with
network lookup disabled. Both `go list -f '{{.Name}}' ./cmd/` and its `-e` form
returned `main` and exit0 in this environment. Thus a missing-import test would
broaden compilation-error examples but would not establish sensitivity to `-e`
here. The existing undefined-symbol failure already protects against hiding a
selected command's compile failure; no additional blocker is raised.

Reviewed the coordinator's complete raw reports and invocation metadata at
`/tmp/ensemble-compaction-baseline/grader-22-buildfix{,-run}.json` and
`grader-23-buildfix{,-run}.json`: advertised explicit-`./agent` commands exit0,
Chapter22 scores200/200 in2.489s and Chapter23 scores100/100 in1.125s, all rows
passed. Those are coordinator-executed checks, independently inspected here.
The preceding command-selection failures remain historical baseline results.
Broader package regression checks were still running when this disposition was
written; their eventual outcome belongs in the coordinator's progress record.
No model inference or live application operation was performed for this review.

## Reordered accounting source-plan review

October 14, after build/scope checkpoint `a8eec84a`. The coordinator reports
broader grader-package regression PASS in30.429s, vet PASS, and the existing
long-response baseline PASS in63.11s. These are coordinator-reported follow-up
results, separate from this reviewer's independently executed Build probes above.

Reviewed the complete pre-edit plan at
`/Users/bill/projects/ensemble-reordered/docs/edition-2-notes/codex-reordered-foundations.md`
from `reordered_foundations_coder`. **Plan accepted**, with no blocking findings.
This is permission to perform the bounded maintenance stage, not acceptance of
its implementation, an export or a new manuscript order. The independent source
and validation review remains after implementation.

Verified the proposed baseline source trees against `a8eec84a`:

| Source | Tree |
| --- | --- |
| Accepted14 | `0c751dbd3985dee1a80284461e31346471ccc984` |
| Accepted16 | `83e2f4cc85eb91ef9f62a54f638baa4d538cacf8` |
| Accepted17 | `2ee06bb6189898809d80de97df2b73c695cfb0d5` |

The plan uses14 as the actual cumulative baseline and16→17 only as a guide to
the accounting change. It retains the real actor/history/journal and separate
GUI, adding the Engine-owned lens, model/tier counts, markers and process meter.
It excludes memory/recall construction, band settings, compressor/judge APIs,
forcing, fleet and authentication/Responses. Handoff markers remain here;
surviving-memory fallback and auxiliary attribution have explicit new22/23
obligations. Mixed historical test helpers will be replaced by small external
provider fakes while the actual runtime remains exercised.

Two selected maintenance corrections have appropriate source provenance:
`215ae6ac` excludes unsupported Anthropic5.5 models from active catalog/examples
and tests; `09111e0b` removes the60-second whole-response timeout while retaining
caller-context cancellation and default connection safeguards. The reviewer
inspected the latter's actual Engine diff: it is a small transport correction,
not a reason to import later billing, auth or fleet machinery. Historical
snapshots and evidence retain their original identities.

Validation clarification sent to coder and coordinator: if the renderer changes
require adapting separate `ch05`/`ch06` example modules or their tests, run those
affected modules as well. Core/GUI tests do not cover separate modules. Keep
grader execution coordinated and use explicit staged `agent` paths with old14
and17 feature IDs. The partial `agent` tree belongs only to the isolated
reconstruction branch; it must never replace the full live development tree by
an ordinary whole-branch merge.

## Independent accounting implementation review

October 14. **Accepted as the bounded reconstructed new15 maintenance stage**,
at initial independent review with zero correction rounds. Source is
`288fc6e8f866bc19d65c17909f270f294c90d1de`, `agent` tree
`2b671b45dc5f7b760f2faaff2a77c83e6c7c09c8`, on
`codex/reordered-foundations`. The worktree was clean when its identity was
verified. This is not a blind student result, an accepted frozen export,
manuscript acceptance, renumbering or permission to replace the complete runtime.

Read the complete coder's consolidated source/validation review at
`docs/edition-2-notes/codex-reordered-foundations.md` in that branch before
deciding. Compared the staged source against accepted14 and the selected
accepted16→17 accounting delta, including all31 changed paths, actual
construction, renderer and transport changes, vocabulary, GUI integration,
adapted tests and new lifetime regressions. The mandatory coding skill was
reloaded for this review; the four complete governing/craft documents remain
the previously reloaded active context, with no intervening compaction.

The root constructs the lens with the actual Engine; attachment checks that
owner and establishes its retained reference. Common remains vocabulary and
interfaces. Send observes sanitized request bytes and retains the prediction
locally through the corresponding response; account retains its request-local
model/tier identity. GUI reads Engine's disposable meter rather than restored
History usage. The implementation preserves Chapter14's actual actor, reducers,
journaling, job and skill machinery unchanged. The new marker helper uses only
surviving handoffs. No memory/recall construction, band kinds, forcing-window
field, auxiliary send API, fleet, auth or Responses machinery was found.

The supported-model correction is confined to catalog/active fixtures; the
historical provider capture is unchanged. The transport correction retains the
normal HTTP client and caller contexts while removing the total response
deadline. Its tests use this stage's Chat Completions transport rather than
importing later Responses code. These match the two approved carry-forwards.

Independent line counting reproduced the coder's Go measurements, counting
physical lines, active nonblank/non-comment-only lines and `//`-only comments:

| Change against its own predecessor | Physical | Active | Comment-only |
| --- | ---: | ---: | ---: |
| Staged production Go versus14 | +502 | +388 | +85 |
| Old17 production Go versus16 | +516 | +405 | +84 |
| Staged primary Go tests versus14 | +685 | +631 | +32 |
| Old17 primary Go tests versus16 | +598 | +547 | +35 |
| Staged separate Go examples/tests versus14 | +8 | +6 | +2 |

The slightly smaller production delta excludes future memory fallback and
auxiliary plumbing. Test growth is explained by the151-line real response
lifetime/cancellation coverage and scoped boundary adaptations. The new
exported types, members and interface methods have useful contract comments;
lens locking, actual-send placement, price identity and temporary meter meaning
are explained beside their behavior. No generic framework or unnecessary
abstraction was introduced. JavaScript/HTML integration remains the small
historical meter addition; current application behavior, not line reduction,
supports this acceptance.

Independently executed, from the frozen stage with provider/application variable
prefixes stripped:
`go test -race ./ensemble -run 'TestCache|TestBlockedModelResponseCancellation' -count=1`
in `agent`: PASS,1.410s package time, exit0. No grader or paid call was run by this
reviewer. Inspected the coder's complete raw reports at
`/tmp/ensemble-reordered-checks`: old14 and17 each100/100, all rows passing;
core/GUI and affected separate-module checks, actual browser meter, race checks
and the two real61-second response cases. The latter's raw package time is
62.019s; the coder's70.046s entry includes build/launcher time.

The five submitted mutation outputs fail their intended assertions: omitted
handoff marker in both explicit renderers, lens rotation from offline rendering
for all three vendors, wrong Sonnet price, absent real external-tool log despite
surrounding lens diagnostics, and the restored60-second deadline on both
streaming and buffered61-second responses. The failed initial log-mutation
launcher remains a setup error. Unchanged historical overlap/rotation mutations
retain their old identity; the current overlap tests were rerun under race.
The stronger memory and auxiliary-accounting obligations remain deferred to
new22/23 and receive no present acceptance credit.

The ch05 test correction is accepted: it selects the actual tool's log record
and validates that record's timestamp instead of requiring the tool record to
remain the final diagnostic. Its omission mutant fails at that exact check.
The prior disk-full runs establish infrastructure failures, not runtime
regressions; subsequent clean checks are separate results. The peer idle-EOF
timeout during concurrent browser checks remains an unexplained intermittent
failure. The peer production source is unchanged, and both the exact14 control
and isolated staged repeat pass; this supports retaining the limitation without
inventing a reproduced baseline defect or waiving the shutdown behavior. No
threshold or runtime workaround was added to hide it.

No blocking scope, architecture, implementation or evidence finding remains for
this stage. The coordinator may release the next bounded authentication plan.
Keep this partial source on its reconstruction branch; later export and
publication gates remain separate.

## Authentication reconstruction source-plan review

October 14. **Plan accepted**, no blocking findings, at plan-only commit
`185a1e31969eb94c2cbec54a9437455988ce1fba` on the reconstruction branch. Read the
complete new "Accounting acceptance and authentication source plan" section in
the coder's consolidated note. Verified that `agent` remains the independently
accepted accounting tree `2b671b45dc5f7b760f2faaff2a77c83e6c7c09c8`; no
authentication implementation is accepted by this disposition.

The proposed guide is accepted17 tree `2ee06bb6189898809d80de97df2b73c695cfb0d5`
to accepted18 tree `c72d29294763f23f51071d7484b0d4283ee70c33`. Independently
verified that recorded source `534112a2d6f9222577db94159c8bd3dbb8368376` has that
accepted18 tree under its historical source path. Inspected the32-path delta,
transport/lens integration, public auth API and shared command client, route
capabilities/error mapping, Responses rendering, original retry fixture and
newly selected model rows. The selected feature boundary is coherent.

The four-line Responses branch in the historical judge belongs at new23 and
stays deferred. The four `RecallSettings{Disabled:true}` literals in new tests
can be omitted because no recall service exists; the OAuth/Responses assertions
remain. No memory, forcing, fleet, sandbox, web, Codex backend or later model
switching feature is needed. Keeping current14 handoff boundaries in the new
Responses renderer preserves the already accepted accounting stage without
restoring memory fallback. Only the reviewed Sol/course routing rows are added
now; later crossover integration retains its own stage.

The single existing Send path can receive credential acquisition, same-source
401 renewal and successful response accounting without recreating the old
auxiliary send seam. The plan explicitly checks successful usage exactly once,
latest-main counts and one logical-request lens observation across the retry.
Reviewer clarification sent to coder: perform that observation through the
resolved-credential redaction path, preserving sanitized artifacts while
retaining the logical-request prediction. Temporary `EN_HOME` matches the
actual public command client's configuration boundary. The auth service's
bounded exchange timeout is distinct from the already corrected model-response
lifetime; carrying it does not undo `09111e0b`.

The planned real external-boundary fakes, full affected module checks,
OAuth/Responses/cache/cancellation races, actual browser summary/meter checks,
repeated long-response regression and serialized old18/17/14 feature grading
cover the adapted boundaries. New mutations focus on retry accounting and
Responses input observation; unchanged historical evidence remains historical.
The known browser callback EOF limitation is preserved honestly and must not
become a fabricated authentication failure or new silent repair.

No fresh provider-support claim, account access, source edit or test execution
was performed by this plan review. The coordinator may release this bounded
implementation, with independent implementation review before advancement.

## Authentication reconstruction implementation review

October 14. **Accepted at initial independent review, zero correction rounds.**
Reviewed frozen commit `970676abeebce8f156d5cdc7ee0dd47a40708136`, Agent tree
`21b15bc6cd1c24a2ee3e03873a87a9c080744402`, against accepted accounting
`288fc6e8` / tree `2b671b45dc5f7b760f2faaff2a77c83e6c7c09c8` and the selected
historical17→18 trees recorded above. The reconstruction worktree remains clean.
This is authorized maintenance with disclosed historical-answer access, not a
new blind student run. Reloaded the entire mandatory
`.agents/skills/ensemble-coding/SKILL.md` and all four governing/craft documents
after compaction before this formal review. Read the coder's complete current
plan and implementation report, all31 changed paths, and the raw auth results.

The31-path selection follows the accepted plan. Exact byte comparison confirms
the auth service, public auth API/shared command client, route-capability and
safe-error handling, Responses renderer/parser and captured plan fixture match
accepted18. Differences from that cumulative source are the deliberately absent
memory/recall features, previously accepted supported-model/deadline corrections,
single-Send integration and44 added test lines. No judge, forcing, memory,
recall, web, sandbox, fleet or Codex-backend implementation entered this stage.
The historical judge's Responses branch remains deferred to new23.

Source review verifies the actual ownership and public-client paths:
`agent/ensemble/auth.go:32` constructs the Ensemble-owned AuthStore;
`agent/ensemble/ensemble.go:127` selects explicit/saved billing before constructing
the Agent; both commands use the public auth client. Shared interfaces retain
useful contract comments, and implementation spokes depend on common rather
than each other. The auth service preserves private atomic replacement,
cross-process refresh serialization, verified identity and scope checks, safe
errors, and retained billing intent on expiry/sign-out. Its30-second exchange
timeout does not reinstate a deadline on model generation.

The adapted transport is coherent: `agent/internal/engine/engine.go:110` resolves
request-local credentials before sanitizing diagnostics; line174 sends the same
sanitized bytes to the log and main-request lens; line186 permits one renewal
through the same selected AuthStore. The retry reuses rendered bytes without
another lens rotation. Line254 accounts the successful response once, updates
latest-main usage and applies that request's own prediction. Failed Responses
streams cannot reach this accounting callback because terminal completion and
present usage are validated first. Existing handoff-based marker placement and
model/surface-bound opaque replay remain intact. No unused auxiliary transport
seam or new framework is needed.

Independently executed on the frozen source with provider/application variables
stripped and a fresh temporary `EN_HOME`:
`go test -race ./ensemble -run 'TestOAuthInferenceBillingAndIsolation|TestOAuthLensRedactsResolvedBearer|TestResponses' -count=1`
from `agent`: **PASS,1.236s package time, exit0**. Raw output is
`/tmp/ensemble-reordered-checks/auth-independent-review.txt`. This exercises the
real Engine/AuthStore/actor against external HTTP peers, including the two
adapted credential/retry tests and captured-wire Responses tool execution.

Inspected complete coder logs in `/tmp/ensemble-reordered-checks/auth-*`:
old18,17 and14 each score100/100 with every row passing; core/GUI tests and vet,
separate ch05/ch06 tests and vet, ch07 build, focused and GUI race checks,
actual-browser Responses summary and accounting checks pass. The real61-second
response cases pass at62.018s package time. Distinguish package/browser durations
from the report's launcher wall times. Earlier full/grader/lifetime results
retain their pre-cleanup provenance; the final focused race run and this
independent run cover the final equivalent cleanup. No broad grader or paid
repeat was needed.

The four new mutation outputs fail at their intended behavioral assertions:
double successful usage, duplicate retry lens rotation, a resolved bearer left
in the actual readable lens artifact, and omitted Responses input observation.
They are ordinary assertion failures, with no compile/setup failure credited.
The test additions exercise real transport and state; no internal mock or
parallel test framework was introduced. Historical39-mutation evidence remains
historical, not a claim of39 new runs. Callback acknowledgment EOF and the
earlier intermittent peer EOF remain recorded limitations; this stage does not
claim either fixed. There is no new grader exception.

Independently reproduced the report's source counts from tracked Git blobs:
physical/nonblank-non-`//`/comment-only production-Go delta is
1631/1378/189 versus historical1635/1383/188; test delta799/741/40 versus
755/702/36. Examples are unchanged; production JavaScript grows5/5/0 and
JavaScript tests53/50/2 in both comparisons; HTML is unchanged. No generated or
dependency files enter these counts. Production totals are10760/8546/1641.
The auth delta's comment share is below the rough target, but the preserved
explanations teach ownership, identity validation, token rotation, terminal
authority and public-summary boundaries. Exported declarations retain useful
documentation. Growth is the selected existing capability plus focused boundary
coverage; line counts do not substitute for the source and behavior review.

No blocking scope, architecture, implementation, comment or evidence finding
remains. The coordinator may prepare/release the next bounded reconstruction
stage; this acceptance does not merge this partial Agent tree into the primary
runtime or authorize frozen exports, renumbering or publication. Carry-forward
reminder from coordination: when the later stages combine, explicitly place
`df38abe0`'s required actual visible reasoning and main-chat/child-submit behavior,
and preserve `09111e0b`'s response lifetime correction. The old cumulative
snapshots predate those maintenance changes; this reminder adds no auth-stage
implementation requirement.

## Historical crossover reconstruction source-plan review

October 14. **Plan accepted**, no blocking findings, at reconstruction plan
`484c7234796492d7a0dbb470a7f24d58578d7fb8`. The clean reconstruction worktree
still has accepted auth Agent tree `21b15bc6cd1c24a2ee3e03873a87a9c080744402`.
Read the complete new source-plan section and exact accepted18→19 diff; reloaded
the full mandatory `.agents/skills/ensemble-coding/SKILL.md`. The four complete
governing/craft documents remain in active context from the preceding review;
no intervening compaction occurred.

Independently verified historical source `bade50cef5cdc53490d4263fa57f5e37abbb93ec`
has tree `bd2998e68f976dbaf437b642b131d350e086242f`, and live source
`fab031529cda55a76d530478f09de0cd08e7db36` has tree
`59010a4c2b333c2f74777b6d76079e8f2b3af90d`. Only README changes between those
two historical sources. The selected accepted18→19 delta is exactly three paths:
the Astra capability row in `internal/engine/media.go`, actual-rendered
model/high-effort assertions in `ensemble/responses_test.go`, and the README's
explicit Astra/high/subscription configuration paragraph. Moving that row to
the existing Responses renderer needs no new public API, dependency, credential
mechanism or auxiliary service. Retaining the earlier Chat controls checks that
the migration stays model-specific.

The planned affected core/GUI tests and vet, focused boundary race check, and
two isolated wrong-route/wrong-effort mutations are proportionate. The separate
examples contain no Astra selection or reasoning-effort configuration; there is
no reason to repeat their unchanged checks, the browser/lifetime tests or paid
tasks solely for this row change. Preserve no-total-response-deadline and
supported-model corrections. Current19/new17 remains ungraded; do not invent a
score or copy the historical repair checklist into this small source stage.

Historical receipts are correctly separated: ch20-code-review records the actual
27.46-second five-request Astra/high task, six unchanged passing tests, normal
exit/save and zero observed summaries/cache grants; enhancement-19-review records
the separate1.003-second scripted replay and its additional recall-judge request.
That extra fixture exchange establishes no early recall dependency. The plan
keeps both original identities and makes no fresh live/current-provider claim.

The explicit later integration ledger correctly preserves `df38abe0`'s visible
explanation gate and the main-chat/child-submit distinction without importing
fleet or later maintenance machinery here. Those outstanding obligations must
be placed and validated before complete reconstruction/export acceptance; this
plan earns no credit for their implementation. The coordinator may release the
three-path stage, followed by independent implementation review. No Agent edits,
test execution, exports, renumbering or primary-runtime replacement occurred in
this plan review.

## Historical crossover reconstruction implementation review

October 14. **Accepted at initial independent review, zero correction rounds.**
Frozen submission `078cccbd2e6b999b320289aa1d096f7abe394350` has Agent tree
`f94f414deb977d9f106b8d749a57faa934987049`; its worktree is clean. Reloaded the
entire mandatory `.agents/skills/ensemble-coding/SKILL.md` before review. The four
complete governing/craft documents remain in active context with no intervening
compaction. Read the coder's complete new implementation report and source diff,
then inspected every retained `crossover-*` result.

The source changes exactly the three approved paths and matches the selected
historical18→19 delta: `agent/internal/engine/media.go:31` routes Astra through
the existing Responses implementation; `agent/ensemble/responses_test.go:140`
checks actual rendered model/high effort while keeping earlier Chat controls;
`agent/README.md:126` documents the explicit configuration. The route comment
explains reuse of existing Agent-owned settings. No new symbol, abstraction,
dependency or future feature appears; existing authentication, accounting,
supported-model and model-response lifetime source is untouched.

Independently reproduced tracked-source counts against each stage's own
predecessor: production Go +2 physical/0 active/+2 comment-only; Go tests
+6/+6/0; README +3 physical lines. Reconstruction and historical deltas match.
Active lines exclude blanks and `//`-only comments. No example, other language,
capture or generated file changes. Preserved comments explain the sole behavior
change; the small test growth protects observable output from the real renderer
without internal mocks or new test machinery.

Raw evidence confirms core short tests (ensemble13.356s), GUI tests1.118s,
core/GUI vet, focused model-boundary race1.020s and final restored focused
test0.008s pass. Both scratch mutations compile and fail the intended checks:
Astra's former Chat route fails its transport assertion, and forced low effort
fails the model/effort assertion on all three Responses rows. No setup failure
is credited. The initial nonexistent-filename read remains a reporting/setup
observation, not a failed runtime check. This tiny, exact historical source
selection needs no duplicate independent test run after inspecting those checks;
this review ran no grader or live/model task.

No blocking implementation, scope, architecture, comment or evidence finding
remains. Current19/new17 remains ungraded, and the historical27.46-second live
and1.003-second scripted results keep their original source identities and
limits. Later visible-explanation, main-chat/child-submit and auxiliary/memory
integration obligations remain outstanding. Acceptance permits the coordinator
to release the current20/new18 source-plan stage; it does not accept those later
features, merge the partial source into the primary runtime, export snapshots,
renumber chapters or publish. Checkpointing waited for the coordinator's shared
Git operation to finish at `6b8f4a2e`; only this owned review note is included.

## Combined web/status reconstruction source-plan review

October 14. **Plan accepted**, no blocking findings, at
`68793cad6c3fd8cb88a27e250e4bc8b9c087e4c8`. The clean isolated worktree still
has accepted crossover Agent tree `f94f414deb977d9f106b8d749a57faa934987049`.
Reloaded the full mandatory `.agents/skills/ensemble-coding/SKILL.md`; all four
complete governing/craft documents remain in active context with no intervening
compaction. Read the complete new source plan, all23 historical delta paths
(recovering the truncated test range), both selected correction diffs, existing
GUI/settings/meter code and relevant combined-grader source/review.

Independently verified source `f35b6ae8ce43c91a52b9972361093db6258e3e20` yields
web tree `f351d66a7bc456650e05ac6656a7db59bbbd725a` and source
`1da0891a7ea0f121b05225c62be3b2190e4f3918` yields combined20 tree
`a801628b1dd631e5ea20ad03ea7d255584728d31`. Selection from those two accepted
steps preserves MCP-owned HTTP, connection/job cancellation, safe diagnostics,
skill declaration/dispatch enforcement, shipped web supervision, owner-derived
status and actor-serialized model selection. The existing Engine already retains
model/tier usage and unknown-price state (`agent/internal/engine/usage.go:20` and
`:37`); status needs a formatter, not another accounting owner or copied facts.

The narrow GUI prerequisite is justified independently of memory. Current
`agent/gui/gui.go:283` holds `Server.mu` across `updateSettings`; root observation
calls observers synchronously (`agent/ensemble/ensemble.go:217`). The proposed
SetModel waits for the actor, so that actor can need the mutex held by its
waiting settings caller. Historical `05b43a0` separates patch/subscription order
from observation locking and takes only short snapshot/publication locks. Carry
those sections into gui.go/settings.go, omitting all memory fields, validation,
actor messages and tests. This is a necessary current feature dependency, not
permission to copy the cumulative old15 or combined20 tree.

The proposed causal real-GUI regression is appropriate: establish streamed
observer delivery and a setter actually waiting on the actor before release;
check busy refusal, progress/shutdown and an idle apply/broadcast control. A
sleep or merely starting a goroutine cannot establish that ordering. Its scratch
mutation must restore the actual lock cycle and fail the bounded liveness
assertion; outer process timeout alone would not establish the claimed cause.
This plan review establishes the source-level dependency, not a newly reproduced
runtime deadlock or a preaccepted regression implementation.

The supported fixtures correctly select the existing `215ae6ac` correction.
Sonnet-5 charges are independently consistent with the declared rates:
100 input,30 write,20 read,10 output yield $0.000379;70/10/20/3 yield $0.000199.
Subsequent unpriced Opus-5 work must keep the total unknown. Current20's separate
grader uses Sonnet-5→Opus4.6→Sonnet-5. Neither fixture establishes summation of two
priced Anthropic models after the supported-model change; the plan preserves
that limit and adds no5.5 row or fabricated price to evade it.

Planned core/GUI and separate-module checks, focused HTTP/status/model/settings
races, actual built CLI and browser model-field observation, and serialized
current20 grading cover affected public paths. The advertised current grader
total is155, retaining historical original21/22 scores separately. Its two
manager-only surviving mutants remain limitations; HTTP cancellation, parked
Recv wakeup and manager isolation retain their separate tests/source evidence.
The three adapted-boundary mutations focus on the new lock dependency and the
supported pricing fixtures, while unchanged historical mutations keep their
identities. Temporary EN_HOME and stripped FIRECRAWL/provider variables preserve
the no-live/no-private-profile boundary.

No memory, recall, judge, fleet, sandbox or Codex backend enters this stage.
Auxiliary usage/status, latest-main forcing, lens exclusion and credential/route
inheritance remain obligations at new22/23. Later visible-explanation and
main-chat/child-submit maintenance also remain outstanding; historical sample
judge calls do not pull them forward. No runtime edits or tests occurred here.
The coordinator may release this bounded implementation, including the stated
lock prerequisite, followed by independent implementation review. Export,
renumbering, manuscript and complete-runtime integration remain separate gates.


## Combined web/status reconstruction implementation review

October 14. **Accepted at initial independent review, zero correction rounds:**
source `9af3864363c1d574f66b44585b35e65c488dfec1`, Agent tree
`6cb2eb95e7a7ad5d5fe9afc8ca2d4f5a7ad3e99d`, extending accepted crossover
`078cccbd` / `f94f414deb977d9f106b8d749a57faa934987049`. The isolated
reconstruction worktree is clean. This is the approved maintenance reconstruction
with historical-answer exposure, not a blind student run. After compaction,
reloaded the entire `.agents/skills/ensemble-coding/SKILL.md` and all four full
governing/craft documents before resuming formal review. Read the consolidated
submission, selected historical deltas, final changes and raw local evidence.

The 24 Agent paths implement accepted plan `68793cad` / review `076048cd`.
HTTP transport, MCP client/manager, skill parsing, model selection and status
implementation match their selected historical combined20 source. The extra
production adaptation is the reviewed `05b43a0` GUI ordering prerequisite; it
adds no memory service or future actor message. `agent/gui/gui.go:286` orders
patches with settingsMu and takes the observation mutex only for publication;
`agent/gui/settings.go:91` snapshots under the short lock and waits for SetModel
without holding it. Subscription follows the same order. The model remains a
separate session-only patch, with mixed persisted patches refused.

The actual ownership path remains dispatched status → ToolContext.Engine →
Engine.Agent, with no copied meter or model facts (`agent/internal/tools/status.go:17`).
SetModel changes only model/resolved identity between turns, preserving provider,
billing, credentials, history and accumulated usage (`agent/internal/agent/model.go:15`).
The CLI and separate GUI call the public Agent; core and GUI still construct
through the public Ensemble root. New exported interface/field/method contracts
are documented. HTTP lifetime, redaction, optional request cancellation and
transport ownership retain the accepted implementations and useful why comments.
The existing single-Send credential/accounting path and deadline correction are
unchanged. No production, architecture, scope or KISS blocker remains.

Independently reproduced counts from tracked Git blobs, using physical / active
(nonblank, not `//`-only) / comment-only lines. Historical web grows production
230/192/25 and tests405/369/27; historical status grows production118/95/15 and
tests314/283/23. Combined historical growth is348/287/40 production and
719/652/50 tests. Reconstruction grows361/298/42 production and834/757/57 tests.
The differences are exactly13/11/2 production for GUI ordering and115/105/7 tests
for its causal regression. Both sequences add3 JavaScript and1 HTML line;
JavaScript tests and separate ch05/ch06/ch07 examples are unchanged. Final
production Go is11,123/8,844/1,685, examples711/625/64, tests7,079/6,462/421.
The added production comment share is12.4%, below the rough teaching target;
comments explain the actual lifetime/ordering/ownership decisions and preserve
useful inherited teaching rather than padding the ratio. This is selected source
reuse, not a claimed new size improvement over the historical solution.

Raw evidence verifies core short tests, GUI tests, focused core/GUI race checks,
separate ch05/ch06 tests, ch07 compilation and reported vet checks. The unchanged
current20 grader returns155/155 across all11 rows; no grader or command path
changed in this stage. The actual Chrome probe edits the model field, observes
both tabs, captures Opus-5 then Sonnet-5 requests and checks retained prior
conversation; its GUI exits0. This establishes effective interface behavior
against a local external-provider fake, not live model/service availability.
The status test is byte-identical to correction `215ae6ac`; no unsupported5.5
fixture or invented pricing row is introduced.

The new regression uses real Agent/GUI/WebSockets and a small external HTTP fake.
Its observer gate establishes streamed progress, then observes the public
SetModel waiting before releasing that progress. The reviewer independently ran
`go test -race -count=1 -timeout=30s -run TestModelSelectionDuringStream -v .`
in the GUI module with a fresh EN_HOME and stripped provider/application
variables: PASS, package1.025s, including busy refusal, continued progress,
later idle selection, two-tab acknowledgment and shutdown. Raw result:
`/tmp/ensemble-reordered-checks/web-independent-review.txt`. No internal mock,
production testing interface or generic test framework was added.

All three scratch defects fail their intended claims. The lock defect first
passes idle selection, then records a genuinely blocked SetModel and five-second
wire-read failure; the retained stacks show the actor blocked in ObserveAgent
on the GUI mutex held by the waiting setter. The12-second alarm also terminates
blocked teardown. This is bounded causal liveness evidence with passing baseline
and restored controls, not a clean ordinary assertion exit or an unrelated
timeout credited as a kill. The other defects fail historical-count/cost
preservation on selection and honest unknown cost after unpriced work. No
compile/setup error is credited. Reviewed raw browser script/output, mutation
JSON/logs, grader JSON and restored checks under `/tmp/ensemble-reordered-checks/`.
Unchanged historical HTTP/control mutation evidence retains its original source
and the previously recorded manager-only survivor limits; no broad repeat was
needed.

Supported pricing checks prove priced Sonnet-5 work, preservation of incurred
cost on selection and unknown cost after unpriced Opus-5 work. They do not newly
prove summing two priced Anthropic models. Historical23.35-second web and
16.34-second browser status demonstrations, old failed grades and auxiliary
judge sample exchanges keep their original identities. This review made no
paid/live call or private-profile read, repeated no unchanged61-second request
control, and makes no new EOF/callback claim.

New22/23 auxiliary usage, latest-main forcing/lens separation, child credential
inheritance and judge Responses remain deferred. Required visible-explanation
and main-chat/child-submission maintenance retain their later integration gate.
Acceptance permits the coordinator's next bounded reconstruction handoff; it
does not merge this partial Agent into primary runtime, export snapshots,
renumber, revise manuscript or publish. Stage D plan review is a separate task.


## Sandbox reconstruction source-plan review

October 14. **Plan accepted, no blocking findings:**
`7ab4b60a357be5cb5d797181bebb6f3abc06a9e4`, current21 → proposed19.
Verified clean reconstruction worktree and unchanged accepted Agent tree
`6cb2eb95e7a7ad5d5fe9afc8ca2d4f5a7ad3e99d`. Reread the complete mandatory
`.agents/skills/ensemble-coding/SKILL.md`; all four complete governing/craft
documents remain active from the preceding full reload, with no intervening
compaction. Read the complete current Chapter21, source plan, full historical
22-path delta, all408 lines of sandbox tests, final correction and historical
independent review. No implementation, grader or live test ran for this plan review.

Independently bound accepted combined20 tree
`a801628b1dd631e5ea20ad03ea7d255584728d31` to final sandbox source
`64e5d0c17be8db116fa4c1da52cd290c45b4e2ef` / tree
`9dc3e9e36a854ede4cc2cbef22dc389850cedc0c`. Evidence checkpoint320859a0,
export574fbab5 and frozen solutions/ch21 retain that same tree. Both sandbox
implementation blobs also match complete baselinea8eec84a exactly. First-green
fcb98a94 and its live evidence precede the two fixes and are correctly excluded
as reconstruction source. The22-path historical delta is1,124 insertions and
76 removals; selecting20 paths and deferring only memory.go/recall.go is coherent.

The selected boundary follows the current chapter's two distinct enforcement
paths: required-Agent Sandbox owns rooted file operations, while commands and
descendants use the macOS profile. Canonicalization includes existing ancestors
of missing files, and operations use os.Root rather than reopening a checked
name outside confinement (`internal/sandbox/files.go`, accepted21). State, media,
cache artifacts and retained job output therefore need the selected hooks now;
they are current consumers, not new services. Composition installs Sandbox before
those consumers and preserves omitted-spec full access versus explicitly disabled
web access. Spec policy and its implementation stay with their existing owners.

The public defined AgentSpec supports Child/Clamp outside common without adding
a fleet. Unconditional narrowing and its diagnostic remain independent, and
canonical child-root checks reject existing symlink escapes. Safe mode enforces
the named execution/supervision removals at registration, declaration and
dispatch while file tools remain useful. Web policy withholds the named skill
and its startup graph; it does not promise to police arbitrary custom MCP.
There is no new policy framework, credential-store tool, alternate-platform
implementation or invented future API in this source selection.

Final64e5d0c's immediate physical-temp xcrun rule and caller-spelled cwd path
are required corrections (`internal/sandbox/command.go:49` and `:100`,
`internal/tools/tools.go:281` in the historical tree). Retain their real allowed
Python/Git/relative-cwd controls beside the escape/disclosure assertions. The
synchronized SetAPILog setter shares the root writer lock; configured-key log
sanitization must be layered onto the accepted request-side configured/resolved/
renewed-credential redaction, without replacing that single-Send code. The plan
explicitly preserves this and the prior supported-model/deadline corrections.
The direct OpenFile, actual readable-log, snapshot/job-output and absolute-in-root
media tests address known historical observation gaps without extra capabilities.

Planned core/GUI/separate-module tests and vet, focused sandbox race and affected
HTTP/accounting/OAuth checks are proportionate to these public and filesystem
changes. Serialized unchanged current21 grading has seven rows totaling100;
its actual CLI/public consumer, reachable-host network control, successful file/
Python/Git work and readable trace distinguish real enforcement from setup
failure or blanket refusal. Confirm macOS/sandbox-exec availability first. The
two new scratch restorations should expose the real nested-cache escape and cwd
disclosure at their intended assertions, retaining successful controls and
restored baseline. Reusing unchanged historical mutation evidence is appropriate
with the OpenFile survivor and earlier media wrong-boundary attempt preserved.
No new browser flow or paid live repeat is warranted by this bounded plan.

The two omitted files are explicit later obligations: new22 memory Scan/write/
temporary-install/archive operations use Sandbox; new23 confines memory/handoff
walk/read, skips nonregular entries and disabled web-skill recall, and roots any
explicitly selected external trusted skill corpus separately. Deferred auxiliary
accounting, forcing/lens separation, judge Responses/credentials and later visible
explanation/main-chat-child-submit maintenance remain unchanged. Actual child
construction must enforce derived specs at the later fleet stage; Child/Clamp
alone proves no spawned-child boundary.

This accepts the concrete source plan for coordinator release into bounded
implementation and subsequent independent review. It claims neither new runtime
acceptance nor universal security: system/device/cache allowances, macOS scope
and rooted-filesystem limitations remain explicit. Historical0/100 and the
48.394s/9.847s live sessions retain their own source identities. No export, tag,
renumbering, manuscript edit, whole-tree merge or publication follows from this
plan disposition.


## Sandbox reconstruction implementation review

October 14. **Accepted at initial independent review, zero correction rounds:**
`922caeb4be404c335d209348a9d36aca4f88645e`, Agent tree
`06c7edc9489eff9cf1bf071a8b6cccba303fcc8b`, extending accepted web/status
`9af38643` / `6cb2eb95e7a7ad5d5fe9afc8ca2d4f5a7ad3e99d`. Clean isolated
worktree verified. Read the complete submission and reread the entire mandatory
`.agents/skills/ensemble-coding/SKILL.md`; all four full governing/craft documents
remain in active context from the last full reload, without intervening
compaction. This remains approved maintenance with historical-source exposure.

Exactly the20 planned Agent paths change. Compared every selected path against
accepted historical20→21: all added/removed source lines match, except gofmt
alignment of the new private Agent fields in the smaller reconstructed struct.
The five new files, including all408 sandbox-test lines, are byte-identical to
accepted21. Shared files retain their earlier staged source and receive only the
sandbox delta. No grader or executable-dispatch change accompanies this stage.
The only non-Agent path in the implementation checkpoint is its consolidated
reconstruction note. The source adaptation conflicts are resolved without
importing memory, recall or their forced-handoff branch.

Required-parent ownership, construction before consumers and rooted operations
remain intact: Agent holds the specification and Sandbox, Sandbox retains that
Agent, and tools/state/media/jobs reach its actual owner. The public root handles
policy validation and configured-key log sanitization; the existing transport
continues handling resolved/renewed credentials before trace or context. The
selected SetAPILog lock matches writers. The current single-Send accounting and
request-lifetime code is unchanged. Public Child/Clamp behavior stays outside
common; no child runtime is introduced. Safe-mode registration/declaration/
dispatch refusal, named web-skill withholding and physical confinement preserve
their distinct meanings. No new architecture, scope or KISS finding remains.

Independent tracked-blob counting reproduces historical growth620 physical /
515 active /73 comment-only production Go lines versus reconstructed587/486/69.
Both add408/376/20 test lines. Active excludes blanks and `//`-only lines;
examples, tests and languages are separate. The33/29/4 difference is precisely
deferred recall integration; omitted memory hooks have zero net line growth.
Examples and JavaScript/HTML are unchanged. Final reconstructed production is
11,710/9,330/1,754; tests7,487/6,838/441; separate examples711/625/64. The roughly
12.4% comment share of added nonblank production remains below the teaching
target but preserves useful explanation of ownership, rooted access, missing
ancestors, caller-path diagnostics, descendant profile allowances and narrowing.
New exported contracts are documented. This is smaller current scope, not a
claimed redesign improvement inferred from counts.

Raw local evidence under `/tmp/ensemble-reordered-checks/sandbox-*` confirms
core tests (integration14.007s), GUI tests1.044s, focused sandbox/HTTP/web/status/
OAuth race3.460s, separate ch05/ch06 tests0.008s/0.246s and restored command
race1.250s. Vet/build reports retain empty successful outputs. The actual
unchanged current21 grader reports100/100 across all seven rows. Its real
CLI/public-consumer and OS checks remain separate from the source tests'
additional web/state/media coverage. The misspelled timeout flag failed before
tests; the corrected result is distinct and no confinement result is inferred
from that setup failure. An initial reviewer read used the wrong cwd-mutant
filename; the correctly named artifact was subsequently read in full.

Both fresh scratch defects fail their intended behavioral observations. The
nested-cache defect passes useful-command controls then returns the planted
OUTSIDE_COMMAND_MARKER with exit0, causing the kernel-escape assertion. The
cwd-prefix defect produces two explicit canonical-root disclosure assertions;
the test continues through its remaining controls. These are actual defects
observed through real Agent/jobs/process behavior, not compilation failures or
timeouts. Restored source passes its race control. Unchanged historical
mutations retain the earlier OpenFile survivor and wrong-boundary media attempt
as limitations, with direct-open and absolute-in-root media controls now present.
No internal mock or new test framework is added. Given exact source reuse and
these fresh checks, no duplicate independent execution was needed; the other
coder's exclusive grader slot was undisturbed.

The implementation adds no fresh live-model claim. Original0/100 and the
48.394s/9.847s historical sessions keep their original source identities.
macOS scope and system/device/cache allowances remain explicit; these tests
provide no general mount, hardlink or hostile-native-code certification. New22/23
memory/recall filesystem hooks and all recorded auxiliary-accounting, forcing,
lens, judge-credential/Responses and later explanation/submission obligations
remain deferred. Actual fleet construction must later enforce the derived policy.
No unresolved implementation or evidence finding blocks this stage. The
coordinator may release the next bounded source plan; this acceptance authorizes
no whole-tree runtime replacement, export, renumbering, manuscript or publication.


## Combined GUI reuse/fleet reconstruction source-plan review

October 14. **Plan accepted at initial independent review, zero correction
rounds:** `53958bcc6f3171cbcd1a0d169c3cfec3a67e65a5`, consolidated
reconstruction note, section “Combined GUI reuse/fleet source plan — current22
→ new20” (line1128). The clean isolated tree still has accepted sandbox Agent
`06c7edc9489eff9cf1bf071a8b6cccba303fcc8b` from `922caeb4`. This is source
selection for authorized maintenance with historical-answer exposure; no new
implementation, grade, live use or export is accepted by this disposition.
Reloaded the entire `.agents/skills/ensemble-coding/SKILL.md` and all four full
governing/craft documents after compaction before this formal disposition.

The selected historical `64e5d0c`→`0430df4` GUI and `0430df4`→`e00d5bc` fleet
deltas preserve the combined exercise. Their14 and23 paths share README, giving
36 unique paths before the explicitly selected schema and maintenance work.
The55-path estimate and necessary separate-example dependency metadata are
reasonable scope estimates, not permission to copy an entire later answer.
Source inspection covers the GUI asset/component/relay implementation, terminal
construction and human-delivery paths, owner interfaces, actor/jobs/history
integration, and the complete historical Fleet implementation. The original
separate GUI module remains separate; public reuse adds no headless GUI import.
Private embedded assets, explicit disk override, component-only rendering and
correlated/cancellable relay are retained. Existing settings locking, supported
models, credential sanitization and response-lifetime corrections must survive.

The construction boundary now makes the previously taught sandbox policy real
for spawned Agents. Historical `ensemble/fleet.go` NewChildAgent checks the root
relationship and Clamp before construction; Spawn uses the actual spawning
Agent, and the CLI factory retains parent configuration. The plan appropriately
tests actual confined child work as well as widening/escape refusal. Questions,
submissions and reliable completion remain separate; no factory, root-interface
or parent-pointer shortcut is justified by routing IDs. Existing actor/jobs
semantics preserve incoming-message wakeups without killing the waited job.

Accepted Stage B `a0acfb5e` is a separate, explicit addition to this earlier
chapter. Reviewed its schema admission, durable submission and affected public/
GUI tests against the plan. Its Fleet-owned compiled contract, exact-number
validation, independent bytes and validation-before-write/notification order
remain applicable. The bounded validator and pinned dependency are already
reviewed Stage B scope; their limits do not promise wall-clock completion.
The only runtime lookup dependency is `internal/fleet/submission.go`
SubmissionData calling later `management.go` Find. That function simply resolves
the owned child map under its mutex, so a private lookup preserves the needed
semantics without importing public traversal or management. Retaining factory
references with synchronization adapts tests without exposing a new public API.
Explicit factory/sink setup replaces EnableFleet; the existing single-Agent
GUI construction and replay are sufficient. Defer get_submission and its tool
assertion to new21, preserving typed SubmissionData and durable-read coverage
here. Inert payload keys do not introduce a memory service.

The four-line `450f7a9` dispatch correction is a justified prerequisite: once
the submit batch drains, record the existing turn_ended fact before finishing.
Its source fixes a state fact missing at the earlier submit introduction, not a
new selection feature. Completed and freshly loaded state checks, paired with
a genuinely held-running control, test this exact claim. This fact alone cannot
establish worker quiescence, final save or automatic dismissal; those Stage C
boundaries remain new22. No Terminal/Quiesce/outcome API is needed here.

The selected five fleet/role paths from `3aad2b7e` preserve main-chat completion
and delegated-child submission with three corrective wakes. The historical
CLI root-escalation expectation is obsolete under that accepted policy; replacing
only that branch while preserving human question/reply and child escalation is
appropriate. `df38abe0` supplies the actual-stream and whole-response explanation
gate now, including all six supported OpenAI rows, paired refusal and no job
side effect. BaseSystemPrompt keeps authored child roles separate from the
current-model policy. Synthetic fixtures remain unflagged; historical captured
bytes keep their original identity. No Codex, speech-calibration, later GUI
observations, management or memory mechanisms are prerequisites of these changes.

The proposed affected-module tests/vet, focused races, outside-module GUI/browser
controls and six adaptation mutants are proportionate. They exercise real
Agents, histories, jobs, validators and transport with external fakes. Reuse of
unchanged historical/Stage B mutation evidence retains its original source and
limits; the fresh mutants must fail their intended assertions after passing
baselines. No grader or runtime execution occurred during this plan review.
Run the unchanged current22 14-row/200-point baseline only in the coordinator's
serialized grader slot. Because Stage D's revised contracts/checks are still
changing, that baseline is not final export acceptance: Stage E must separately
run the finished revised grader against the exact reconstructed export. Keep
original24/25 failed scores and historical live receipts distinct throughout.

No blocking dependency, architecture or scope issue remains in this concrete
plan. Coordinator release may proceed to bounded implementation, followed by
independent comparative review of actual paths, active code, useful exported
comments, tests and evidence. New21 management/GUI selection, new22 memory-worker
lifetimes and auxiliary accounting, new23 recall/judge obligations, and optional
24/crossover integration remain in the explicit ledger. No whole-tree merge,
export, tag, renumbering, manuscript change or publication is released here.


### New20 plan correction: defer the later after-tool interrupt integration

October 14, during implementation, before submission review. The coder found
a forward dependency missed in the accepted plan and my initial source review.
Independent inspection concurs with the coordinator: accepted new19 `922caeb4`
has `Interrupt struct{}` in `agent/internal/common/actor.go`; its actor dispatch
at lines488–489 invokes immediate interruption and its turn has no afterTool
field. Commit `55cbef230e675ed0d42547b7ad21e53c5b8a554f` adds the public
AfterTool field, private turn flag, inbound handling and report-boundary handling
with original26 management, which belongs to reconstructed new21.

**Accept this narrow staging correction:** defer only `df38abe0`
`agent/internal/agent/visible_reasoning.go` lines52–55 and
`TestVisibleReasoningRejectionHonorsAfterToolInterrupt` in
`agent/ensemble/visible_reasoning_test.go` (its line349 posts AfterTool:true).
Restore both when new21 introduces the actual AfterTool capability, alongside
its accepted actor handling and focused test evidence. Keep the other explanation
checks, actual stream ordering, whole-batch paired rejection, no execution/job
side effects, round limits and existing immediate-interrupt behavior in new20.
There is no absent earlier API to emulate, and no reason to import management
early merely to compile this later integration branch.

This corrects the original plan's promise to honor after-tool interruption and
retain every focused test at new20. The full primary runtime keeps that behavior
throughout; only reconstruction staging changes. The coder discovered the
omission before formal implementation review, so it is an attributed source-plan
correction, not a new formal correction round or waiver of final coverage. No
runtime, grader or live execution was performed for this assessment. The full
mandatory skill and four governing/craft documents remain in active context
from the preceding reload, with no intervening compaction.


### New20 plan correction: successful completion validation and fixtures

October 14, first core validation before submission review. The coder found
that the dispatch-only selection from `450f7a9` was incomplete: successful
submit appends a payload-free turn_ended, but this earlier history validator
requires one payload. Independent inspection confirms the same accepted commit
`450f7a9599295dafb260c1650b9b5553010e2a54` changes
`solutions/edition-2/main/internal/history/history.go` validate to admit exactly
`e.Type == "turn_ended" && count == 0`, with the accompanying explanation.
The earlier reducer already maps turn_ended to idle; no vocabulary or state
architecture is being added. **Accept that guard/comment as a prerequisite**
of the reviewed dispatch event. Retain all other payload-count and matching-tag
refusals, and do not import later memory/recall union members. The existing
completed/reloaded-state and held-running controls remain the relevant tests.

Raw `/tmp/ensemble-reordered-checks/fleet-core.txt` retains the actual initial
“event requires one payload” failures in lifecycle and submission tests. This
is a source-plan omission missed by both the independent and coordinator plan
reviews, discovered through implementation validation; it must not be reported
as a passing first run. No formal implementation correction round has begun.

Two adjacent fixture corrections are also within scope. The capability assertion
requires claude-ch16-course to be a known model, but that synthetic row first
enters `ccd63509c3b2e08ff5a231b0409423b236953022` with memory ContextWindow
forcing. Defer that assertion to new22 alongside its row; preserve all current
six OpenAI opt-ins and unknown/default-disabled controls. The newly written
spawned-sandbox test instead looked for Tool.Name on tool_returned. Current
`agent/internal/agent/dispatch.go` puts Name on tool_called (line55), while
the returned record (line106) carries CallID, parts and IsError. Repair the
test's actual outside-read call/result correlation; neither add a runtime field
to accommodate the assertion nor let an unrelated error satisfy confinement.
The raw initial known=false and outside-read assertion failures stay recorded.

No runtime or test execution was performed by this reviewer. Source inspection
and existing raw output establish the assessment; post-correction test results
remain the coder's pending evidence. Mandatory skill and all four full governing/
craft documents remain in active context without intervening compaction.

### New20 implementation review: initial disposition and evidence

October 14, independent dependency reviewer. Reviewed reconstruction submission
`f3009a05eb5875d029028c07c047e63adf33579f`, Agent tree
`a709f71f0fb57fcfd541738d8bb578f172d06fd5`, against accepted new19
`922caeb4` and the sources selected by plan `53958bcc` / review `ab511273`.
The full mandatory `.agents/skills/ensemble-coding/SKILL.md`, workflow, carryover,
voice and chapter procedure were read, including complete reloads after
compaction. This is authorized maintenance reconstruction with disclosed answer
exposure, not a fresh blind student run. No runtime, grader or live execution
was performed by this reviewer; the assessment inspects source and actual retained
outputs, rather than repeating the coder's unchanged checks.

**Initial disposition: one documentation correction; no remaining production or
test findings.** `agent/README.md:222–235` still tells the CLI/root to finish
with submit and applies three corrective wakes without limiting them to ordinary
delegates. That conflicts with the accepted main-chat/delegated-submit policy
already implemented and tested. Return this single grouped finding as correction
round 1 after initial review. Correct only the explanation and submission record;
there is no behavior change requiring another test or paid run. Historical
chapter numbers in this staged README are not a separate renumbering request.

The 60 changed Agent paths match the accepted scope and subsequent narrow
dependency corrections: 54 implementation/test/documentation paths and six
separate-example module metadata files. Historical GUI reuse `64e5d0c→0430df4`
and fleet `0430df4→e00d5bc` remain the foundation; selected `3aad2b7e` role
policy, `df38abe0` explanation enforcement, Stage B `a0acfb5e` contracts and
`450f7a9` completion facts are bounded additions. The validation guard and
after-tool deferral preserve the two source-plan omissions recorded above.
No memory/recall, management/selection, named peers, Stage C worker machinery,
optional Codex route, core GUI dependency or unsupported model rows enter here.

Actual ownership and construction are coherent. `ensemble/ensemble.go:109`
validates the immutable schema before Agent side effects; required parent links
and child-owned config remain explicit. Child specs copy their values and clear
inherited submission schemas for ordinary grandchildren. `internal/fleet` owns
its child map, message routing, bounded idle correction and durable submission
behavior. Its private `findChild` (`submission.go:159`) resolves only owned
children without importing the later public management lookup. Validation
precedes durable rows and completion; custom JSON survives copies/reload while
legacy rows remain supported. The successful submit batch records the existing
turn-ended fact (`internal/agent/dispatch.go:18`, history validator:173). This
establishes idle state, not Stage C worker quiescence or automatic dismissal.

The independently reviewed identity correction `e9fd0676` is appropriate here:
the composition root binds the immutable schema to its installed compiled submit
definition (`ensemble/ensemble.go:266`); declarations copy stored definitions.
An MCP or custom tool sharing the name retains its own schema/admission, including
when a factory is installed later. SafeMode and compiled-over-MCP precedence
remain enforced. Six real-runtime cases at `ensemble/submission_test.go:523`
exercise these distinctions. The full primary runtime needs its own scoped
carry-forward; this earlier cumulative tree must never replace it wholesale.

The visible-explanation gate checks actual streamed order and the whole response,
pairs refused calls without executing jobs, and retains the round limit and
immediate interruption. Six supported OpenAI rows opt in; synthetic/unknown
models remain unflagged. Restore the specifically deferred AfterTool branch/test
with its real new21 API, and the memory course-row assertion with its new22 row.
The CLI supplies recursive construction and human routing explicitly. Reusable
GUI assets, component-only Artifact scroll and public MCP relay retain their
historical behavior and separate module boundary; the external consumer constructs
real public values and tests headless absence beside a GUI dependency control.

KISS comparison independently reproduces `fleet-line-counts.json` for all four
trees. Active lines exclude blanks and comment-only `//` lines; generated files,
Markdown and module metadata are excluded, tests/examples/languages separated.
New20 Go production grows 1,579 physical / 1,270 active / 226 comment-only lines,
versus historical combined-stage growth 1,022 / 836 / 132. The extra 434 active
lines implement the approved contracts, explanation policy and maintenance;
they do not introduce a second registry or orchestration framework. Go tests
grow 2,781 / 2,597 / 123 versus 1,087 / 1,015 / 50. Their additional 1,582 active
lines protect real contract/copy/disk/provider/GUI boundaries and staged APIs.
JS production grows 7 / 4 / 3 (historical 8 / 4 / 3, trailing blank difference),
JS tests 38 / 35 / 2 in both; examples and HTML are unchanged. Cumulative Go
production is 13,289 / 10,600 / 1,980. Comment shares of 15.1% for the delta and
15.7% cumulatively fall below the approximate 20% target, but useful exported
contracts and adjacent ownership, lifetime, validation and ordering explanations
are present. No padding or arbitrary line reduction is warranted.

The tests use real earlier actors, jobs, histories, renderers and transports with
small external peers. GUI structured-submission tests (`gui/submission_test.go:23`
and :119) compare live and freshly loaded nested JSON and idle state, alongside
genuinely held inference. Factory-retained references adapt Stage B tests without
inventing later Find/EnableFleet APIs. The spawned Sandbox check correlates the
actual outside-read call and returned error and retains an allowed-read control.
CLI checks exercise human reply while work runs and main chat completion. Actual
Chrome and external-module checks cover reusable UI/relay behavior; they are not
fresh live-model demonstrations.

Raw `/tmp/ensemble-reordered-checks/fleet-*` outputs support the submitted checks:
final core passes (ensemble 15.499s), GUI 4.429s, focused restored race ensemble
3.874s/fleet 1.081s, structured GUI race 1.065s and CLI 0.266s. Separate ch05/06
tests pass; ch07 is build-only. Vet output is empty, with successful invocation
reported by the coder. The Responses browser passes, and the speech regression
has a failed candidate, passing exact922caeb4 control and passing correction.
Package/browser durations differ from the note's explicitly labeled launcher
wall times. Initial completion-validation failures, the GUI peer cleanup timeout
and fixture mistakes remain failures, with their corrections attributed.

All eight executed mutations fail their intended observations: child Clamp,
root idle correction, schema inheritance, owned lookup refusal, custom JSON
GUI output, turn-ended state, submit name gating and name-based schema replacement.
The ownership mutant reaches a deliberately malformed owned row, proving wrong
lookup/refusal behavior, not foreign-file exfiltration. The schema setup attempt
that matched two anchors did not execute a test and is not a kill. Restored
focused checks pass. Reused historical/Stage B/explanation mutation evidence
retains its original sources and limits; these eight defects are not complete
coverage proofs.

Independently matched all 166 manifest entries against the submitted Git blobs.
The final source capture has SHA-256
`eafea8fcaf7d1e2746b11a639520860fd986ac872ef6d6d1c3ee4d26f926ae97`.
Actual current22 grading is 14 rows, 200/200, on the earlier capture
`fa7009f5c5c8c2f4af3d170911fdda97118b1841b805383fddc1d85841300681`.
Seven paths differ from that capture, including the identity correction; it is
not a final-source grader pass. Final affected tests cover those changes.
Stage E still requires the finished revised grader on the exact reconstructed
export. Original24/25 0/100 scores and 9.649s/48.198s live observations remain
historical. No new paid run, export, tag, renumbering or next-stage release is
claimed by this implementation review.

During reconciliation, root also identified current22 §22.6's default-factory
wording as a forward dependency missed by narrative review. This reviewer
confirmed new20 has only `SetAgentFactory` (`ensemble/fleet.go:32–37`) and the
CLI callback (`main.go:137–143`); `EnableFleet`/default construction belongs to
later `ensemble/registry.go`. Correct the prose to this stage's explicit factory
and preserve later capabilities at their introduction; do not expand new20 code.

### New20 final disposition after documentation correction round 1

**Accept reconstructed new20** at
`b4708839e8bb1a523069be02b9188145ae50dac4`, Agent tree
`39449051049e6d7325b4c368e627739f97eab28f`. The exact two-path correction changes
only README root/delegate completion wording and the consolidated submission
note. It closes the sole initial-review finding: the root finishes in chat,
ordinary delegated children submit and receive the bounded corrective wakes.
All production and test bytes remain those reviewed at `f3009a05`; no rerun was
needed for this documentation-only round. One formal correction round was used.

The isolated reconstruction worktree is clean. Independently verified all166
Git blobs in the updated source manifest and its aggregate SHA-256
`b218df3d36173b91ea9163afdc9f02f48a8e0e57427850130e5fe7754be740e8`.
The original submission, captures, initial failures, historical receipts and
pre-identity 200/200 grade retain their original identities. Acceptance closes
this scoped reconstruction implementation review, not Stage E's revised-grader
and exact-export gate. Root controls the next source-plan release. No frozen
export, full-runtime replacement, chapter renumbering or publication occurred.

Scoped prose reconciliation also passes: the coordinator's two-paragraph
`book/chapter-22.md` correction replaces the forward default-factory assertion
with explicit terminal/application factory installation and parent-based
completion policy, and ends the demonstration after the main Agent collects
child results and answers in chat. This matches the accepted new20 construction
and `internal/fleet/fleet.go:347–352` root exemption. It removes premature named
peer/default-factory teaching without adding future APIs. This is only source
consistency review of that diff, not full Stage D prose/grader acceptance. The
chapter remains coordinator-owned and is excluded from this review checkpoint.
