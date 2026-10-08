# Chapter 10 retained-check integration review

Reviewer: replacement `/root/reviewer_ch17`, 2026-10-08. This is a coordinator
review, not a fresh-student report. The preceding Ch17 advisory is separate.
I read the complete Ensemble coding skill, architecture and Chapter 10 contract;
reloaded the skill after context compaction; inspected the retained orchestration,
its fixtures, and the immutable Chapter 10 runtime at the two mutation boundaries.
I read the relevant Chapter 2 and Chapter 9 contracts for the newly reachable
failures. That source exposure disqualifies any claim of cold student discovery.
No student runtime, historical solution or original historical checker was edited.
No credentials, external providers or paid calls were used.

## Result and evidence boundary

The joined result is **66/68 retained checks passing**, on immutable source
`8882a18cf98e9a4b70afccfbe980f6344630aae6`. This is not a single clean full run,
Chapter 10 acceptance, a live demonstration, or approval of a later source.
Two actual runtime assertions remain failing after integration repairs.
[The joined receipt](checkpoint-evidence/ch10-retained-joined.json) names every
check's selected receipt, hashes all evidence, verifies identical 4,092-file
source maps, and enumerates checker identity deltas.

| Receipt | Completed / passing | Meaning |
|---|---:|---|
| [Initial full run](checkpoint-evidence/ch10-retained-full-first.json) | 68 / 59 | Original failures preserved byte-for-byte |
| [Repaired nine](checkpoint-evidence/ch10-retained-repaired-nine.json) | 11 / 9 | Nine affected checks plus two required builds; reveals two runtime failures |
| [Dependency run, interrupted](checkpoint-evidence/ch10-retained-routing-dependencies.json) | 4 / 4 | Two builds and two checks; reserve refusal before command-shapes |
| [Remaining dependencies](checkpoint-evidence/ch10-retained-routing-dependencies-rest.json) | 7 / 7 | Five remaining checks plus two required builds, reduced reserve |

Fifty initial passing rows are reused. The completed receipts bind unchanged
passing checker/fixture identities; only the owned orchestration, the two adapted
mutation scripts, and unrelated concurrent remaining-grader inventory differ.
The last category is listed explicitly in the join: those files are neither
invoked nor imported by the retained checker closure. An identical whole bundle
inventory is not claimed.

Seven previously passing checks were rerun because changing their execution
location changes imported helper identities. Settings-wire, strict-files-and-patches,
command-shapes, actual-policy-effects, exact-settings-revisions and
projection-integer-deletion import `accept_ch07` directly or through `accept_ch08`;
snapshot-handoff-deletions imports `accept_ch05`. Their prepared and original
helper hashes differ, so unchanged main-script hashes alone were insufficient.
Other moved direct scripts use identical original/prepared sibling fixtures and
imports. Existing deliberately bundled retained scripts keep their locations and
identities, apart from the two repaired mutation fixtures which were rerun.

The interrupted receipt has no final whole-checker inventory: gate return never
occurred. Its two completed checks independently include their main/helper hashes
in captured stdout. Those hashes equal the original-file identities in complete
receipts before and after the interruption. The join validates these nested hashes
and records the missing-inventory limitation; it does not manufacture a completed
run. The initial 768 MiB reserve stopped execution after controlled idle cache
maintenance. The coordinator inspected remaining temporary storage (31.6 MiB)
and authorized the existing `--reserve-mib 512` override for the five remaining
checks. No broad cleanup or runner feature was added. The reduced-reserve run
completed; its original receipt and the interruption both remain intact.

## Integration repairs and validation

Only explicitly adapted direct checkers now execute from the disposable bundle:
`ch09-review-record-bounds.py` and `ch09-review-boundaries.py`. The preceding gate's
default execution paths remain unchanged. The former boolean redirected all
HERE-based scripts, causing six `accept_ch09*` programs and delivered-package
discovery to derive the repository root from a temporary location.

The skill boundary deletion now targets the outer persistence branch shared by
ordinary and prepared appends. Its rollback mutation executes after the selected
append attempt. The original standalone positive cases and exact assertions remain
unchanged: `failed transition changed authority` and
`durable transition was rolled back after result failure`. This restores the
existing semantic deletion test; it does not claim new session fault coverage.

The Chapter 6 thinking-display mutation's empty-text guard became ambiguous after
an unrelated partial-argument guard was added. Its adapter qualifies the anchor by
the original stream emit function. It still suppresses thinking display and must
produce the original `thinking fragments missing or entered wrong part` refusal.
All 19 Chapter 6 mutations and the skill boundary controls passed after adaptation.
No failing assertion was removed, renamed or weakened.

[Eleven orchestration tests passed](checkpoint-evidence/ch10-retained-orchestration-tests.txt):
`python3 -m unittest discover -s scripts/edition2 -p test_ch10_retained_receipts.py -v`.
New tests compare actual scheduled paths, preserve prior default behavior, check
adapter hashes and unchanged originals, compare original/adapted assertion ASTs,
and prove the narrowed mutation selects its owner when both empty guards exist.
The focused retained runs exercise the real compiling mutation controls. No Go
source or module changed in this repair; retained module vet/test rows remain
bound to the identical source. Scoped whitespace verification passed. The compiler
slot was released to the remaining-check grader; no further build is authorized
for this reviewer without coordinator release.

## Newly reachable failures and contract questions

The diagnostic executable was copied from the last immutable-source build before
its temporary directory was removed. [Its binding](checkpoint-evidence/ch10-retained-diagnostic-binding.json)
records the complete source map, build command and SHA-256
`fea84d00250d88cc0ffbe7451f9f00c298d798bb6909deac787410636d3c7b3c`.
It is at `/Users/bill/projects/ensemble-edition-2-revisions/ch10-retained-diagnostic-cli`.
This is a separate build of the same source, not a claim that distinct temporary
builds produced identical executable hashes. The unchanged original checkers were
rerun against that bound binary with explicit receipt output.

### Duplicate management arguments terminate a standalone run

[Management diagnostics](checkpoint-evidence/ch10-retained-management-diagnostic.json)
pass 72/138 assertions. All three local provider routes exit unsuccessfully at the
duplicate-argument case with `session validation: duplicate JSON member`, followed
by terminal persistence/transition errors. The log reaches the preceding
`setter-duplicate` call/result, but does not produce the expected controlled pair
for `load_skill` with `{"name":"edit","name":"edit"}`. Many later assertions are
consequentially unexecuted; 66 failed assertions are not 66 independent defects.
These launches select explicit standalone `CH02_LOG` in machine protocol mode,
with no session selector.

Applicable published teaching:

- Chapter 9 §9.6, lines 376–386, requires rejecting duplicate management arguments
  before transition, paired call/result facts, consumption of pending limits even
  on invalid calls, and defaults for the next ordinary call. Lines 430–440 specify
  `invalid_skill_arguments`, no transition, and retained call/result/limit facts.
- Chapter 10 §10.2, selection table at lines 132–142, preserves standalone
  `CH02_LOG` and unselected machine-protocol behavior. Section 10.3, lines 239–250
  and 304–306, leaves legacy standalone validation/behavior under its prior contract.
- Chapter 10 §10.8, lines 779–784, requires strict outer/identity/state parsing and
  duplicate event-member detection. Section 10.3, lines 327–331, distinguishes
  required structural state from arbitrary tool-argument data and its inherited schema.

The observed standalone terminal failure conflicts with the explicit Chapter 9
controlled-error behavior retained by Chapter 10. The teaching question is how to
state the boundary between duplicate structural event members and duplicate keys
inside untrusted tool-argument data, especially for new sessions. Root should
resolve that wording before a student correction; do not silently weaken strict
structural validation or replace the inherited controlled error with acceptance.

Safe student reproducer, using only preceding/current contracts: start standalone
machine protocol with an explicit fresh `CH02_LOG`, a base skill that offers `edit`,
and a localhost provider fixture. Issue a pending `tool_limits` setting (for example
`max_output_bytes:1`), then duplicate-member `load_skill` arguments, then an ordinary
`read_file` call. Observe process status and public events/results. Expect a consumed
limit note and paired `invalid_skill_arguments` error with unchanged skill state;
the subsequent ordinary call must still run with default limits. Do not send the
student grader source or a proposed runtime implementation.

### OpenAI continuation replay changes the argument string

[CLI diagnostics](checkpoint-evidence/ch10-retained-cli-diagnostic.json) pass 50/51:
only OpenAI offline replay fails. The live local run finishes successfully with
five input and five output tokens. [Focused reconstruction evidence](checkpoint-evidence/ch10-retained-replay-diagnostic.json)
retains raw original logs, captured continuation requests, replay outputs and
explicit mismatch paths. Request sequences 8, 16, 23 and 29 differ inside
`$.messages[...].tool_calls[...].function.arguments`. For example,
`{"name": "edit"}` becomes `{"name":"edit"}`. This is a difference in a wire JSON
string, not merely insignificant whitespace between outer request tokens.

Applicable teaching:

- Chapter 2 §2.5, lines 511–515, defines Chat Completions arguments as a JSON string;
  lines 561–562 require byte-identical rendering for fixed context/configuration/adapter.
- Chapter 10 §10.3, lines 282–298, permits one-time formatting normalization at the
  boundary for **new session event writes**, with identical accepted fragments
  used for append, application and observation. Lines 304–306 preserve existing
  records and legacy standalone behavior.
- Chapter 10 §10.3, lines 308–324, preserves exact accepted replay-bearing raw bytes;
  historical reconstruction restores recorded spelling, and semantic equality is
  limited to the specified comparisons. Chapter 9's retained exact replay evidence
  at lines 904–909 is supporting history, not the sole normative basis.

A live/reconstructed continuation mismatch is real. The teaching question is the
accepted spelling for standalone writes versus the explicitly new-session
preparation boundary. Root should settle that before implementation feedback;
semantic equality of decoded tool arguments cannot by itself satisfy this exact
request replay assertion. No claim is made here that raw provider formatting must
survive a permitted new-session preparation step.

Safe student reproducer: use standalone machine protocol and a localhost Chat
Completions fixture returning a tool call whose `function.arguments` string includes
spaces. Allow its result and a subsequent continuation request. Invoke the public
`replay LOG REQUEST_SEQUENCE` command and compare the captured/reconstructed
argument strings exactly. The focused receipt's `mismatches` array gives the
observed paths; `replays` and `raw_logs_by_sha256` preserve the source evidence.

## Handoff

Root owns the contract decisions and any student correction. After this diagnosis,
the author confirmed that both standalone failures violate existing text. Root
accepted a narrow clarification for designated raw argument data in session mode;
its publication and independent review are a separate pending step, not evidence
that the tested source is repaired. Passage line numbers above refer to the
pre-clarification contract inspected for this review. The review's
integration/fixture internals are reviewer material: hand the student only the safe
reproducer paragraphs, applicable new/preceding contract passages and appropriate
sanitized observations. After correction, revalidate affected checks against the
new immutable source; the 8882a18 receipts do not automatically validate a revision.
No chapter tag, acceptance claim, push, provider demonstration or student source
change is part of this scoped repair.

## Clarification review closure at dd1111e

October 8, 2026. Independently reviewed the complete two-file grouped change at
`dd1111e0b09513ea59fd3b524d5446e8c5ec773f`, including the author feedback addendum,
against both retained failures and the coordinator's narrow session ruling.
Reloaded the complete voice guide and chapter-writing procedure. Read the complete
Chapter 10 contract during the retained task and reread surrounding §10.3/§10.8,
Chapter 9 §9.6, and the relevant standalone/replay requirements for this closure.
The reviewer retains the grader/runtime exposure disclosed above and is not acting
as the student. This is a prose/contract review, not runtime repair validation.

**No blocking finding remains in this clarification.** It resolves the two
standalone findings under existing teaching and defines the previously ambiguous
session case without weakening structural validation:

- Chapter 10 lines 328–336 confine duplicate tolerance to designated argument
  data. The object must remain syntactically valid and bounded; applicable scalar
  rules still apply. Malformed envelopes, broken JSON, non-object arguments and
  duplicate event/payload/call/snapshot structural members still refuse. Existing
  scalar rules at lines 241–252 and storage/client bounds in §10.8 remain intact.
- Lines 338–346 keep the containing semantic codec strict. Hashing uses the
  wrapper's decoded original text without collapsing duplicate members into a
  map. Import validates the wrapped data. If either argument has duplicates,
  correspondence requires exact accepted text, tool name and call identity;
  first/last-key selection and a single-member replacement are explicitly invalid.
- Lines 348–376 provide literal data and acknowledgement. The duplicate-name case
  returns `invalid_skill_arguments`, `name:""`, the unchanged current revision and
  the error flag. Pending limits are consumed once with their normal note; paired
  facts remain, no Skills transition or Job occurs, and the next valid call and
  continuation work in both modes. The text also requires preserved outcomes
  through checkpoint, full-log rebuild and snapshot-plus-tail. This agrees with
  Chapter 9's controlled-error, counter and next-call-default rules.
- Lines 378–383 explicitly preserve whitespace inside the decoded valid Chat
  Completions argument string. Outer layout normalization is still permitted;
  semantic equality cannot excuse changing the replay string. The independent
  spaced-valid-argument fixture prevents the duplicate-error case from being the
  only replay check.

The author addendum accurately distinguishes observations reported by the
coordinator/reviewer from the author's own source reads. It identifies the new
session clarification as a coordinator decision, preserves the failed runtime
status, and leaves student confirmation pending. The inserted explanation leads
with the bad-arguments/corrupt-structure distinction and uses literal examples
where bytes matter; no unsupported incident or successful result is invented.

Scoped whitespace and manual voice/fixture review passed. An existing executable
was found at
`/Users/bill/projects/ensemble-edition-2-revisions/executables/edition2-lintprose`;
no compiler was invoked. Running it against a temporary exact export of
`dd1111e:book/edition-2/chapter-10.md` exited 0, with soft warnings: 8,241 words,
17 negation forms versus 16, and a 4,329-word person gap after line 8. These remain
whole-chapter editorial warnings, not a claim of final manuscript acceptance.
Chapter SHA-256:
`51e32b2d336238255c5253f9f92fe3354e28ff778fdb7443e16498a0722bfd32`.
Existing lint executable SHA-256:
`1d5bc762ce42dcb15ec9c8d885d31c0d7ab848dde1c60c19f9573ed29cea5831`.

The coordinator may pin this contract for the grouped student repair. The student
must still confirm the clarification and implement/revalidate affected behavior
on a new source identity. The retained 66/68 result and all failed receipts remain
unchanged; this closure grants no runtime acceptance, live-run approval or chapter tag.
