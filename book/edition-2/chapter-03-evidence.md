# Chapter 3 evidence and reconciliation

Status: complete contract/prose draft; independent review accepted.
No new implementation tested.

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

## Next action

The tool schema/defaults/limits, lifecycle and CLI contract, independent
fixtures, and acceptance map are published. Scoped prose lint passed all
hard checks; soft length/person warnings were read without adding filler.
Full independent review is requested. Resolve its findings before the
student derives Chapter 3 from validated Chapter 2 history. The student
supplies real use evidence before the live section is written.

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
accepted all three amendments; `chapter-03-review.md` records readiness for
a fresh student handoff. No live or implementation result is claimed.
