# Edition 2 Chapter 11 enhancement: MCP

Author: `edition2_author`, October 11, 2026. Original Chapter 12 becomes
Edition 2 Chapter 11, matching frozen `solutions/edition-2/ch11`.
This is an authorized enhancement of the original, with revised grader work
owned by `edition2_grader_coder` and independent review by `reviewer_astra`.

## Reading and source

The author context compacted at the start of this chapter. Before any new
writing, all four required documents were reloaded in full, without omitted
ranges: `docs/agentic-codebook-workflow.md` (1825 lines),
`docs/edition-2-carryover.md` (871), `book/voice.md` (632), and
`book/chapter-writing-procedure.md` (599). They remain together in active
context. The full `.agents/skills/ensemble-coding/SKILL.md` was also reread,
with its truncated final range explicitly recovered, plus current AGENTS,
enhancement record and chapter map. No further compaction preceded this draft.

Read original `book/chapter-12.md` in full, file revision
`ec41c6e8cf0800e266f7a39ebadf255d94379082`, SHA-256
`a754bd50ca7f3c7a2da868ff5acbdb98b4d5eddaa05349e7890c4a3dbf6cd4b9`;
the complete [student record](ch12.md) and [comparative review](ch12-code-review.md);
the accepted direct snapshot's common MCP interfaces, manager, client,
transports, Agent invocation/collection, browser transport/tools, completion
and server-lifecycle tests, README and public construction/dispatch paths.
Read accepted Chapter 10 for presentation continuity. Historical raw live,
browser and slow-peer evidence was inspected alongside the source and reports.
No main-tree future capability or first-edition runtime was edited.

## Outline and preservation

Through-line: the reader wants to add an external capability without rebuilding
the agent, while preserving trustworthy execution and fresh context. The
original has no documented personal incident to reconstruct. Its compiled-tool
opener, transport progression and sharp explanation of silent truncation and
unreadable control states carry the teaching. Those remain, with enough context
to explain why each boundary matters. No anecdote, dialogue or result was invented.

Order remains extension argument → transport → correlation → MCP client →
automatic tools → bridge → reverse calls → WebSocket/browser → skill lifecycle
→ exercise, followed by the actual student demonstration and corrections.
The small original chapter earns additions for the actual owners, missing wire
and launch facts, and observed timing failure; there is no reduction target.

Section facts that earn their place: one transport interface preserves optional
GUI reuse; a reverse call can deadlock its own reader; a callback can return
before execution ends; a 3.4-second snapshot was replaced by a three-second
running notice; discovery does not grant every offered capability; reverse
admission exposes enabled Agent tools; a textual speech queue cannot prove
audibility; stopping a peer differs from retaining its manual; a marker in call
arguments cannot establish the returned result.

## Disposition of the student and reviewer findings

- Corrected the original's automatic admission claim: per-server skill grants
  precede Agent enabled-tool policy, and hidden automatic names are refused by
  ordinary model/reverse dispatch. Explained reverse trust separately from the
  offered-tool filter, without promising a sandbox.
- Replaced copied sibling services and callback-bag sketches with the real
  owner chain and common/public interfaces. Kept the simpler combined correlation
  client, with out-of-order reply and reader-liveness reasons. Internal type and
  method spelling is explanatory, not a new grader requirement.
- Replaced the original bridge's `context.Background()` with an excerpt using
  the actual job context. Distinguished callback deadline, explicit kill, turn
  interruption and shutdown. Retained the slow automatic-completion defect and
  narrow correction, including delayed output retention without stale attachment.
- Published newline framing, integer correlation, dedicated `--mcp-pipe` behavior
  and historical `2024-11-05`. Reverse calls, `ephemeral` and `$/cancelRequest`
  are explicitly course contracts; no current MCP conformance claim is made.
- Kept the four browser tools and the original state/truncation explanation.
  Clarified the 4096-byte UTF-8 cap, actual per-connection delivery, one active
  browser target and speech-observation limit. Source/real-browser evidence is
  separate from a faithful JSON-RPC peer.
- Made frontmatter valid and explicit about descriptions, types and tool grants.
  Explained actual subprocess startup/reaping, workspace-relative paths, literal
  arguments and credential separation. External unload is immediate; compiled
  tools/instructions retain the earlier pending policy. No compaction was added.
- Added a small reproducible marker peer/primary skill and real CLI launch.
  They reproduce the historical pattern, without claiming identical original
  generated bytes. Retained the empty Anthropic tool-argument failure, exit-zero
  limitation, one bounded correction repeat and missing initial response capture.
- Preserved the argument-forwarding assertion survivor as a teaching example:
  inspect correlated result blocks rather than searching the whole request.
- Distinguished browser applied/effective font state from disk persistence.
  Fifteen actual Chrome checks used a provider fake; three live-provider CLI
  sessions used a local MCP peer. Neither is relabeled as the other.
- Preserved original full interrupted/unscored exit 143 and interim-only 80/90,
  with recursive parity unrun and vacuous behavior credits disclosed. Original
  source review accepted one correction round; it remains distinct from this
  manuscript/grader enhancement gate.

Deferred/declined: no resources/prompts protocol, full YAML processor, compaction,
sandbox, generalized codec framework, future GUI automation behavior, new public
mode, new runtime tests or paid repetition. Existing focused evidence covers
valid source obligations beyond the revised numeric grader's fixtures.

## Validation and remaining gate

Direct `go run ./cmd/lintprose book/edition-2/chapter-11.md` passes all hard
checks. Its lower-length and person-gap warnings do not justify padding with
invented stories. The original text and all frozen runtime exports remain intact.

Agreed grader contract retains weights 15/15/15/20/15/10 plus 10 corrected
Chapter 10 parity. Dedicated raw pipe covers handshake/reverse, real CLI plus a
skill-declared stdio peer covers model declarations/results/automatic context,
and the actual optional GUI covers its tunnel. Existing grader-only launch
options remain; no new product interface is required.

Exact sample and final revised grader results are reconciled below. Independent
combined acceptance remains pending; do not advance Chapter 12. The
[coder report](enhancement-11-code.md) retains technical validation and source
binding. Mandatory technical handoff: `.agents/skills/ensemble-coding/SKILL.md`.

Initial independent prose review of `685f0bfe` finds no blocker. During sample
reconciliation, the coordinator and coder flagged the inherited generic
`./ensemble --mcp-pipe` path beside the later explicit `/tmp/ensemble-ch11`
build; the accepted core also has an `ensemble/` library directory. The exercise
now states the core-module build and uses `/tmp/ensemble-ch11 --mcp-pipe`
consistently. This is executable-location clarity, not a claim of a failed
protocol or changed runtime requirement. That exact command now passes.

## Final author reconciliation

No further author compaction occurred after the complete reload recorded above;
all four full documents remain active together. Manuscript SHA-256 is
`37ece94fb7c0cf0a3706dad8cd877d5238cb0e4c23b76455995787ed25a6e453`,
matching [samples.json](../../solutions/evidence/ch12/enhancement11/samples.json).
The narrow command clarification is committed at `139d6922`; no subsequent
manuscript edit was needed for technical validation. Hard prose lint passes
at 2896 words, with the stated soft warnings.

The exact public transport interface and raw/pipe API ran in an external
consumer. The verbatim bridge excerpt was substituted into an isolated accepted
manager; the real MCP source checks, including slow completion and interruption,
passed under race detection. The exact published Python peer and primary skill
ran through `/tmp/ensemble-ch11 chat`, making two requests to a faithful local
provider, returning the actual file marker in result content and recording
read 1 / round 2 / turn 1. The exact pipe command completed its handshake and correlated
real file read. The complete browser skill loaded with an attached real pipe
peer and invoked `gui_click`; the code-search fragment loaded with a separately
supplied faithful subprocess companion. That companion is explicitly outside
the published fragment, which does not claim to supply a search implementation.
These are local sample checks, without a new paid-provider or real-browser run.

Actual full revised commands, read from their retained receipts:

| Source | Command | Result |
| --- | --- | --- |
| Accepted direct `ch11` | `go run ./cmd/grade-edition2 -ch 11 -json` | 100/100, exit 0, 16.843 seconds |
| Original `solutions/ch12` | `go run ./cmd/grade-edition2 -ch 11 -json -gui solutions/ch12 solutions/ch12` | 30/100, exit 1, 15.074 seconds |

Both totals are actual full runs, unlike an explicitly composed score. Raw
[accepted](../../solutions/evidence/ch12/enhancement11/accepted-full.json)
and [legacy](../../solutions/evidence/ch12/enhancement11/legacy-full.json)
reports retain each check; adjacent `*-full-run.json` files record command,
elapsed time and exit. These results do not revise the original interrupted
full run or interim 80/90 student probe.

Legacy earns handshake and real reverse-read credit. Its normal client does
not start the skill-declared stdio peer, so missing model-facing remote tools,
results and automatic context do not establish independent failures inside
each unexecuted downstream mechanism. The GUI phase times out awaiting its
tunnel handshake, leaving later tunnel behaviors unobserved. Existing parity
also fails, with its actual nested reasons retained. The initial focused legacy
probe encountered `no reverse handler registered`; the corrected bounded
readiness check waits for a successful real read, with two attempts in the final
legacy run and one on accepted source. This preserves reverse functionality
without treating handshake completion as readiness for every later operation.

The initial accepted-source fixture also failed for a harness error: its
`tools` field used commas plus spaces, while the published syntax separates
names with whitespace. Commas became part of the first five names and only
the final `auto_broken` matched. The original failing `initial-baseline.txt`
and passing `corrected-baseline.txt` remain separate; the correction changes
the fixture to the existing syntax, without weakening tool admission.

The [mutation receipt](../../solutions/evidence/ch12/enhancement11/mutations.json)
records 22 distinct targeted defects failing their intended checks, with restored
control passing. Some defects cause related downstream failures; the count is
not 22 independent proofs or complete coverage. Existing source/browser evidence
still establishes the broader lifetime, ownership, admission and DOM/native
speech obligations described in the manuscript. Frozen runtimes, original
graders and historical paid/browser receipts remain unchanged.

Read the complete final coder report and
[validation](../../solutions/evidence/ch12/enhancement11/validation.json).
The affected grader race check, command build, vet and formatting pass; the
command package has no tests. Verified every file hash and both frozen Git
tree IDs in [final-source.json](../../solutions/evidence/ch12/enhancement11/final-source.json)
against the current files/HEAD. Tested grader behavior is checkpoint `0b9a9abe`;
subsequent grader changes are explanatory comments only. All manuscript links
now resolve. The preserved full legacy parity report includes inherited
consumer/workflow/media fixture limitations; those are not new MCP behavior
failures. Combined independent acceptance and the coordinator checkpoint remain
the final gate.

Final coder freeze `9be403dd` adds one distinct sensitivity case after the first
author reconciliation: an automatic-tool error abandons the collector before
it can release the provider request. The intended output deadline detects lost
continuation, separately from the earlier failed-output-in-context check. Its
focused baseline and restored controls pass; the receipt records 5.55 seconds
for collecting the intended failed check report. The final coder total is 22,
superseding the earlier 21 without recounting the original cases. The reviewer's
independent bounded persistent-reverse-refusal check remains separate. Actual
full scores and exact manuscript/sample hash are unchanged; no sample repeat
was needed. All four full documents remain active, with no additional author
compaction since the recorded reload.
