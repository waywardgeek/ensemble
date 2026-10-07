# Chapter 3 evidence and reconciliation

Status: prior tool contract validated at `7cbbd8e2`; comparative code review
and complete prose/receipt proofreading accepted. Coordinator reports 39
acceptance cases, eight checker controls, passing control plus five exact
defects, and full legacy regression exit 0 (grade package 508.374 seconds).
Bill's subsequent human-client requirement is now satisfied by accepted
ordinary-text chat runs in actual PTYs on all three APIs. The initial human integration and actual
runs are retained at outer commit `a347ce31511c4b124e486bb41ef98c07bd17cec5`;
independent client/code/manuscript review accepted the result. All 45 interface
cases and passing control plus seven exact mutants pass; evidence-only revision
`339a2a6` passed four identity controls. Earlier JSON-lines runs remain separately
attributed machine-interface/tool evidence. The coordinator's dedicated export/
checkpoint identifier is `edition-2-ch03-r1`; its manifest records reviewed
source and file identities. Bill's editorial approval remains separate.

## Initial student and grader correction

The coordinator reports that first unchanged student checkpoint `590c4f4`
scored 65 with inherited grader assumptions: Gemini declarations recognized
only the narrower `parameters` field, and the fake lacked the explicit resolved
identity required by the new provenance rule. The same student binary scored
100 after those harness corrections. Legacy Chapter 2/3 reference and deletion
checks passed; the later full root regression passed as recorded above. Preserve both runs
and do not misclassify these fixture failures as missing student behavior.

The fresh student's empty-file question exposed a contract ambiguity now
resolved in §3.5: omitted start with omitted end or `end_line:0` is a full-file
read and succeeds with empty content. An explicit start, even 1, or positive
end requests a line target and fails on an empty file. This is a published
clarification before the affected student revision, not a private grader hint.

## Read ledger

Read the first-edition Chapter 3, the derivation's data/loop/grader/staleness
sections, the complete tool-use exhibit, current grader scenario definitions
and wire-result readers, relevant policy passages, history below, current
architecture, and global reviewer's Chapter 3 recommendations. Reloaded the
voice guide and chapter-writing procedure. First-edition implementation is
author/reviewer evidence and is not supplied to the student.

## Historical receipts

- `6f4b4c1`: explicit tool declarations on all three request surfaces, stable
  declaration order, no field for an empty registry, and separate read/mutate
  grading. Its initial read checks depended on newly written files and blamed
  a write failure on reading; planted independent fixtures corrected that.
- `554d1cb`: overwrite refusal became a separate five-point check with new-file
  and explicit-overwrite controls. The prior successful write tests did not
  establish the guard.
- `c9f51e6`: deletion audit corrected exact failure sets and detected mutations
  whose text anchors had stopped matching. Historical timing flakes remain
  historical observations, not claimed current failures.
- `29276c6`: source prose added search context and overwrite explanations.
  `derive-ch03.md` records that search context was built in the live tree but
  absent from the frozen Chapter 3 answer. The new exercise will teach and
  test the behavior it actually promises.
- `derive-ch03.md` records the `go run` exit-code fixture flaw: wrapper stderr
  contains `exit status 7` even when the tool never reports the actual process
  status. Use a silent shell `exit 7` plus a successful-command control.
- The inherited `Ch3OrderingLog` puts a deferred human before the tool return,
  making the result-first rule falsifiable. This exposed a real predecessor
  contract dependency, corrected explicitly in Chapter 2 after coordinator
  approval and reviewer consultation; no legacy fixture was weakened.

## Source status

The current grader has ten checks totaling 100, despite older nine/seven-check
comments: ch2parity10, toolsdecl5, toolloop20, multiblock10, readtools10,
mutatetools5, writeguard5, runcommand15, toolerror15, editcontract5. Preserve
actual current behavior rather than stale header arithmetic.

The tool distribution is a September 13, 2026 exhibit of a private archived
corpus. Its printed rows can be checked arithmetically; the underlying sessions
have not been independently recounted here. The later search-friction analysis
used a different corpus. Do not mix their denominators or present either as
current popularity data. No new model IDs, costs, request sizes, or live
transcripts are claimed.

## Later lessons brought forward

Per-agent visibility constrains execution as well as declarations. Agent-owned
Registry is a working design choice within Bill's allowed ownership options.
The registry is an `internal/tools` service with actual Agent back-pointer;
the Engine reaches it through Agent/common interfaces. Common contains shared
declarations and values, not tool execution. The GUI remains a separate optional
module/public observer client. Logging is reachable through actual owners.

Persist dispatch before a side effect and result afterward. Ordinary call
errors answer the model; persistence/transport failures end the operation and
must not be converted into tool content. Failed result persistence cannot roll
back an executed shell command. Output caps and visible truncation prevent the
old read/shell asymmetry; process lifecycle/containment still belongs to jobs.

## Current primary verification

On October 7, 2026, searched current official documentation for
[Messages tool-call handling](https://platform.claude.com/docs/en/agents-and-tools/tool-use/handle-tool-calls),
which specifies result blocks before other content in the answering user
message, and [generateContent response schema](https://ai.google.dev/api/generate-content),
which provides a structured function-response object with an error member.
These sources inform the contract; they do not establish a successful new run.

## Contract review before implementation

The tool schema/defaults/limits, lifecycle and CLI contract, independent
fixtures, and acceptance map are published. Scoped prose lint passed all
hard checks; soft length/person warnings were read without adding filler.
Independent review preceded the student's implementation from validated
Chapter 2 history. The student supplied real-use evidence before the live
section was written.

Chapter 2's real Gemini consumer exposed the narrower `parameters` schema
field. The verified `parametersJsonSchema` mapping is carried into §3.6;
the shared schema is preserved rather than weakening it to satisfy a wire
representation. Registry declarations remain the sole live authority; the
old public declaration input cannot silently create a second capability set.

Independent full-contract review found two specification gaps: search wording
could label an exactly complete result truncated, and binary classification
was unstated. Corrected the first to require omitted matches/bytes; specified
a NUL in the first 8192 bytes as the exercise's binary test. The live procedure
also calls out known resolved identity before a tool turn: automatic
continuations must obey the existing provenance rule after side effects.
Reviewer found no architecture or lifecycle contradiction. Resolution check
accepted all three amendments; `chapter-03-review.md` records that earlier
readiness for a fresh student handoff. Those contract findings themselves
establish no implementation or live result.

## Live reconciliation and comparative finding

The live manuscript now uses `540fb4a`'s retained all-three-provider CLI and
public-consumer receipts, following initial implementation `590c4f4`. Author
read the feature ledger, actual calls/results, final files, stdout usage,
selected/returned model facts, public consumer source, and all three offline
replay receipts. Each source run is dated October 7, 2026; no paid calls were
made by the author. The guided empty-file/owner-interface revision is separately
bound by receipt source hashes under `revision-{vendor}` and is not attributed
to the initial binary. Its model-invented first answer/recall is stronger than
the earlier supplied-marker check.

Independent comparison then exposed a shared first/new-edition UTF-8 cap
defect: slicing valid `éX` at one byte became a replacement character when
JSON encoded, corrupting text and exceeding the retained-content budget.
Published the complete-prefix rule and cap1/cap2/cap3 positive/negative cases
before the student's affected edit. No binary transcoding format is added.
The correction subsequently passed deterministic Unicode controls and the
independent reviewed acceptance noted above. Successful earlier ASCII live
runs still do not prove the multibyte case. The later human-chat runs below
also exercised the corrected output boundary through actual model calls.

Chapter 3's comparative-review instructions now explicitly assign old-source
reading to the independent reviewer. The fresh student receives rationale
and revised new teaching, not an instruction to inspect the old answer.
Scoped prose lint passes all hard checks after live reconciliation.

## Human-client reconciliation

Read the complete three main terminal transcripts and EOF transcripts, launch
metadata, FEATURES.txt, initial-binding.json, receipts.json and the complete
tool-feature-map.json under `solutions/edition-2/main/evidence/ch03/human-chat/`.
After broad output truncation, reread the feature map in compact form and the
terminal files individually. Recomputed usage, prompt/request/tool counts and
terminal hashes directly from the saved logs/files; inspected final workspace
bytes and both the omitted-overwrite and corrective-turn disk receipts. No
new model calls, verifier execution or derived evidence writes by the author.

All three actual macOS PTY sessions used binary SHA256
`3d49d9037077929b7380fbbae020e968c79075ac94eb95c939535aa606523990`;
the launch and initial binding record source hashes. Main sessions began
2026-10-07 around 19:57:57 UTC and ended around 20:02:21–24 UTC. Messages used
default terminal selection; the other two used explicit chat. Actual actor is
the fresh student coder, never a claimed Bill session. Runtime Linux behavior
is not established by these receipts.

Measured human turns/requests/tool calls were 5/18/27, 6/22/29 and 5/32/27 for
Messages, Chat Completions and generateContent. Recomputed normalized usage
I/W/R/O was 78489/0/0/3313, 10777/0/33664/1381 and 120234/0/0/4452. Selected
and returned IDs are distinguished in the manuscript. Every path used all six
tools, refused unflagged replacement and ambiguous/missing edits, recovered,
verified empty-range and Unicode-cap cases, and reported a silent exit7.

Chat Completions omitted overwrite:true during the requested recovery. The
recorded calls and intermediate file contents expose the omission; an observed
corrective user-path follow-up produced the intended bytes. Its claim about
parallel execution is not adopted as a fact: the dispatcher remains serial.
Its description of retained é as an incomplete sequence is also contradicted
by actual valid UTF-8 tool bytes. No code change is inferred from these model
narration/planning mistakes.

The human /history output made result5 discoverable; /redact5 succeeded on all
three. Saved request-prefix reconstructions preserve that result's call ID and
stub while removing the original text. These derived files are offline
reconstructions, not captured HTTP bodies. The one-request directive is consumed
once, and saved replay controls report unchanged file/log hashes. Raw session
receipts remain bound to the initial integration even if review changes code.

Independent review accepted the integration after 45/45 interface cases and a
passing control plus seven deliberate defects with exact expected failures.
The coder's evidence-only revision `339a2a6` binds the verifier to immutable
source and executable identities before execution or derived writes. The four
saved controls refuse a wrong executable, wrong source and changed launch
record without changing evidence, and reproduce the original derived bytes
with the correct identities. The reviewer independently confirmed those
controls. No production code changed or paid calls were repeated.

The manuscript now points to frozen `ch03/evidence/ch03/` receipt paths for the
coordinator's dedicated `edition-2-ch03-r1` checkpoint. The source-read path
above records where the author actually inspected them before export. Preserve
`a347ce3` as the original production/live source identity and `339a2a6` as the
evidence-only revision. The export manifest supplies the exact reviewed source
and file hashes without a self-referential final commit hash in the manuscript.
The coordinator exported reviewed source `8494fdb0` into 491 exact tracked
files, verified all four modules at that location, and bound the checkpoint
in `solutions/edition-2/manifests/ch03-r1.json`. No push or Bill editorial
approval is inferred.
