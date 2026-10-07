# Second edition: Chapter 3 outline

Status: prior tool contract validated at `7cbbd8e2`, with initial checkpoint
`590c4f4` retained. Independent comparison and final prose review accepted;
39 acceptance cases, eight checker controls, control plus five mutants, and
full legacy regression passed. Human-client review and evidence reconciliation
are now accepted under Bill's ordinary-text chat requirement. The initial integration and all-three-API
human PTY runs are retained at outer commit `a347ce31511c4b124e486bb41ef98c07bd17cec5`;
45 independent interface cases and a passing control plus seven exact mutants
pass. Evidence-only repair `339a2a6` passed four identity controls; production
code and original paid receipts are unchanged. Chapter 3 demonstrates tools
and history-selected redaction through that human interface. Coordinator
checkpoint/export uses `edition-2-ch03-r1` and its validation manifest;
Bill's editorial approval remains separate.

## Voice plan

Stake: the reader is letting model output change real files and must know
what ran, which result answered which call, and what remains true after a
failure. Resolve with an inspectable multi-round tool session and on-disk
evidence, not the model's claim that the edit succeeded.

Open with the change from answering questions to changing files. Retain the
historical tool-count exhibit only as a dated, private-corpus account; do not
turn it into a fresh general benchmark. Explain dedicated tools through
bounded reads, unambiguous edits, and capability selection before schemas.

## Story-preservation pass, voice v5

Sources: first-edition `book/chapter-03.md` §§3.0–3.1 and its tool-use exhibit;
actual October 7 second-edition Chat Completions terminal/call/disk receipts.
Keep the historical concentration of ordinary file/shell work as the reason
for a small tool set, with its existing exhibit link and private-corpus
qualification. Do not repeat the old agent-narrator percentages as universal
or current measurements.

Foreground the new observed overwrite omission and correction in the opener;
retain the exact follow-up and resulting bytes in §3.10. It shows a person
using the logs to distinguish model narration from effects. The coder drove
the interaction; no Bill participation is invented. Compress the closing
acceptance/mutation ledger into the existing validation link while retaining
the Unicode and verifier lessons. Exact contracts, tables and code blocks
remain unchanged; original tags and runtime receipts remain historical.

## Thesis and order

A tool call is part of an unfinished turn. Persist the call, execute it under
the requesting Agent's capabilities, return an ID-matched result even for
ordinary errors, and let the model continue. The event model and ownership
architecture from Chapters 1–2 remain intact.

1. Why the file operations deserve tools besides a shell.
2. Explicit owner choice: Agent-owned Registry in `internal/tools`, Agent
   parent interface in common, Engine reaches it through Agent.
3. TL;DR and complete six-tool argument/result contract.
4. Durable call/result execution order and distinct failure classes.
5. Provider declarations, result/error rendering, signatures and ordering.
6. Bounded synchronous filesystem/shell behavior and honest scope.
7. Independent fixtures, inherited points, new acceptance/deletion targets.
8. Actual CLI/public-consumer demonstrations on all three introduced APIs.

## Working choices

- Registry storage is per Agent for this chapter, a coordinator/author
  choice within Bill's allowed alternatives, not a new claimed user ruling.
  The same visible set controls declarations and dispatch. No global map or
  closure-wired sibling dependencies.
- Agent owns its captured absolute workspace. Relative paths resolve there;
  no process-global `os.Chdir`. This is not containment or a sandbox.
- Calls run sequentially. Sixteen model requests per human turn, counting
  the initial request. Complete the last accepted batch, then fail before
  request 17 if continuation is still needed. No whole-turn rollback.
- CLI emits only the final assistant response for each user input; intermediate
  text, calls, and results remain in history and public observer events.
- Exact unique-anchor edits; missing/ambiguous/empty anchors refuse without
  changing the file. This strengthens the deliberately open first-edition
  choice and will be taught/checked explicitly.
- Guarded replacement, supported append, bounded read/search/list output,
  and bounded captured command streams. Truncation is visible. Synchronous
  execution remains; jobs and process containment belong to Chapter 4.

## Source issues already resolved

The inherited ordering fixture contains a deferred human before an outstanding
tool result. Root approved a narrow Chapter 2 clarification now published:
one deferred human is representable through the same append/replay path;
Submit refuses it synchronously, and rendering waits for all results then
projects results before the deferred input. This introduces no runtime mailbox.

Do not copy the original global Registry or ToolFunc wiring, stale defaults,
unguarded shell-output size, claims that cwd confines access, or a blanket
claim that malformed argument JSON cannot appear inside a valid response.

## Gates

Use a fresh student context without coordinator history. Its curriculum is
the new chapters/skill/architecture and preceding new solution; first-edition
chapters, solutions, graders, and author/reviewer research are excluded,
including historical links in the skill. Record actual reads.

Independent contract review before the student builds; retain inherited
Chapter 3 checks and earlier regressions, add explicitly published properties,
audit deletion controls, then actual live feature runs. Preserve initial
student answer for independent first-edition comparison and revise code and
teaching before final validation. Collect evidence for the newly requested
final cross-edition comparison without pre-claiming an improvement.
