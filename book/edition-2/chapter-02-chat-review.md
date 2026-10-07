# Chapter 2 human client review

## Contract review, October 7, 2026

The independent reviewer read the complete current voice guide and chapter
procedure, architecture decisions, mandatory coding skill, and the Chapter 2
human-client additions in the TL;DR, §2.7, and §2.10. The bounded contract
review finds no material implementation ambiguity. This is acceptance of the
published client contract, not acceptance of an implementation or live run.

The motivation identifies the actual usability failure: JSON-lines input is
a machine interface, even when it sends requests to a real model. The new
contract teaches explicit chat/protocol modes, terminal-based default selection,
flushed prompts, readable answers, local commands, discoverable history IDs,
exact input preservation and byte limits, and distinct local-command versus
operation-failure behavior. Both clients use the public Agent operations.
The GUI remains an explicitly optional stub. Successful result redaction in
Chapter 2 remains a controlled public-consumer exercise; the prose does not
pretend an ordinary no-tool chat session contains tool results.

Earlier machine-interface receipts remain labeled as such. The manuscript
does not claim that Bill personally ran a terminal demonstration. Its new
terminal receipts and revised source checkpoint remain pending.

## Independent acceptance prepared before student source review

`scripts/edition2/accept_chat.py` exercises actual local PTYs, waits for each
prompt, supplies individual inputs, and reads each displayed answer. It also
uses explicit piped chat for the 1 MiB input boundary, where the operating
system's canonical terminal line limit would otherwise confound the program's
own limit. Three literal fake provider surfaces are used; no student serializer
or implementation is imported or inspected.

The planned 42 cases cover explicit/default terminal chat; local inspection and
malformed commands; spaces, literal JSON, and slash escaping; one-request
ephemera; real multiline display; usage, quit, and EOF; exact UTF-8 byte limits;
invalid encoding and oversize refusal; configuration, redaction, and request
failures; empty/tool-only notices; and explicit/default machine compatibility.

Five independent checker controls passed. They include a minimal executable
that asserts both streams are terminals, verifies incremental prompt/answer
handling without terminal echo, and a negative executable with no prompt.
The other controls protect usage interpretation, alternate valid text wire
forms, and the absence of text parts in the tool-only fixtures.

The first frozen student client checkpoint is
`56dacfad01f71f2a1b20d39bca15846edef41ddd`, derived from the reviewed Chapter 2
history. Before opening its implementation, the checker ran against its retained
executable, SHA256
`f62927fa1bc2449d16f3f10fc3eebacbed12fbd0d20b6508dcf8ffb5e7fa36db`:
**42/42 passed**. The initial receipt is
`/tmp/ensemble-ed2-chat-independent-initial.json`, pending durable preservation.
This is deterministic client evidence, separate from paid usability.

## Post-run code comparison and requested revisions

After that first black-box run, the reviewer read the complete new chat client,
mode selection, and platform terminal helpers, and compared the first-edition
`solutions/ch02/main.go` (last changed at
`ef155b1cccecf90a1f6fa791e61cc800489d6f1c`). The old standard already had an
explicit human chat option. The new edition initially omitted that user path;
passing its machine tests did not establish equivalent usability.

The new client keeps both presentations over public submission rather than
constructing internal engines in the executable. It preserves nonblank input
bytes, gives actual terminals a useful default, implements local history and
directives, and bounds UTF-8 input without confusing CRLF with prompt bytes.
The buffer flush before reading has a useful explanation. Output failures reach
the caller through checked flushes. Separate input/presentation functions add
some code but do not duplicate conversation orchestration. Platform automatic
terminal detection currently covers Darwin and Linux; other builds retain
explicit chat. Linux has cross-build evidence, not a claimed terminal run.

Two concrete revision requests were sent as rationale only, without historical
answer code:

- Redaction validation still collapses invalid span/level/reason and absence of
  an eligible result into the generic conversation-transition error. Introduce
  static actionable reasons and retain private-payload exclusion and failed-event
  nonmutation. This implements the existing diagnostic teaching.
- The platform `terminalFD` helpers drop the owner interface passed to their
  caller. Preserve that context across the syscall boundary so diagnostics can
  reach the application logger when needed.

The reviewed revision resolves both findings. Its new public regression rejects
invalid spans, levels, reasons and spans without results with distinct static
reasons, checks returned/root-log privacy, and proves rejected directives leave
log bytes, context, events and usage unchanged. A valid target still works after
the refusals. Every platform helper now retains the public owner interface.

The receipt verifier also needed a historical binding: it originally compared
initial source hashes to the current mutable worktree. The revision reads source
bytes from the preserved initial Git commit, accepts an explicit executable path,
and checks its hash against launch receipts. Legitimate revisions no longer make
historical receipts unverifiable. Original launch records remain unchanged.

The frozen reviewed executable SHA256 is
`6e4399307e20fedf1c982677e39c3dadc70beb94e933cb237323468949bbfce9`.
The accepted revision is checkpoint
`ad0d80e33a3a2857e8e0887117d9b099f1a4786d`, preserving initial `56dacfad`.
It passes **42/42** independent checks. A serial disposable audit has one passing
control and five compiled defects, each producing exactly its expected failure
set across the three local provider surfaces:

| Defect | Expected failing cases per provider |
|---|---|
| Disable automatic terminal detection | default PTY chat |
| Keep the doubled slash | explicit/default PTY chat |
| Accept invalid UTF-8 | invalid UTF-8 input |
| JSON-escape displayed answers | explicit/default PTY chat |
| Omit prompt flush | explicit/default PTY chat |

All six audit rows passed, and source hashes remained unchanged. This proves
sensitivity for these selected properties, not exhaustive mutation coverage of
every client promise. Initial, reviewed and mutation receipts are preserved
byte-for-byte under `solutions/edition-2/ch02/evidence/ch02/human-chat`;
`review-binding.json` binds their hashes and the reviewed source/executable.
No additional paid requests were needed for the
local diagnostic/detection changes.

## Human terminal receipt inspection

The reviewer read the complete primary transcripts for all three providers and
the feature ledger under `evidence/ch02/human-chat`. Each shows plain questions,
readable recall and reversal on separate lines, local commands, syntax recovery,
literal slash submission, and clean quit. Final disjoint counts in the actual
terminal output are:

| API | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| Messages | 418 | 0 | 0 | 168 |
| Chat Completions | 263 | 0 | 0 | 160 |
| generateContent | 212 | 0 | 0 | 894 |

These are the coder's terminal interactions, not Bill's participation. Historical
machine/public-consumer demonstrations retain their original source identities.
The reviewer also inspected the controlled-result public test: public Append
establishes a result, human history exposes its sequence/call ID, human redaction
removes its content from rendered context. This is correctly labeled local
evidence. The updated §2.10 abridged transcript, table, local-control distinctions
and macOS runtime/Linux cross-build qualification agree with the retained files.
Final coordinator validation/checkpoint binding remains separate from this
accepted code and prose review.
