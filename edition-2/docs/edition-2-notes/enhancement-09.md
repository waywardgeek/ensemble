# Enhancement notes: Chapter 9 (original Chapter 10)

Author: `edition2_author`, October 11, 2026. Prose and exact samples complete;
initial independent prose review has no blockers, combined review pending.
Mandatory technical handoff:
`.agents/skills/ensemble-coding/SKILL.md`. No later chapter drafted.

## Reading and source

Loaded all four required documents in full into active context before authorship:
workflow (then 1,768 lines), carryover (862), voice (632), chapter procedure (575).
Read the subsequently added authorization/context notices in full. Also read all
AGENTS, mandatory skill, enhancement plan and chapter map, original Chapter 10,
the complete ch10 student/coordinator notes and comparative review, and accepted
Edition 2 Chapter 8 for continuity. No compaction has occurred during this work;
after any compaction all four must be reloaded in full before further authorship.

Original chapter revision `e1605916114230570b348ab5961c193fa1cebf80`, 240 lines,
SHA-256 `b04d6facb19ae61b78a8246028ac505955bb8c139a710b9aaaaf29f5c4fb3476`.
Accepted source is direct `solutions/edition-2/ch09` (original10), preserved by
`edition-2-ch10-r2`; original evidence remains under `evidence/ch10`. Read its
skills/parser/common/tool implementation, real-loop and CLI tests, constructor
assembly and public API/README. All originals and frozen solutions remain intact.

## Preservation and outline

Through-line: a person extending the agent needs to grow its capabilities without
continually rewriting its identity or advertising tools it cannot execute.
Resolve that problem with the original constitution distinction: fixed initial
instructions, dynamic manuals delivered through the existing tool result.

Original10 has no documented personal anecdote. Preserve its concrete argument,
the progressively unfolding skill tree and the constitution explanation instead
of inventing one. Retain original section order: TL;DR, Format, Registry,
Progressive Disclosure, Variable Substitution, Tool Provenance, Constitution.
Add Taking it for a spin from retained actual sessions, including the reviewer's
correction of the unload inference. Useful section facts: horizontal rules must
survive parsing; a rejected dependency graph must grant nothing; a guessed tool
name tests dispatch; dynamic `$SKILLS` cannot rewrite the initial prompt; `$TOOLS`
can reenter the registry; shared initial tools keep initial provenance; a read
after unload proves nothing about a writer enabled by a different skill.

There is no trimming or padding target. The draft retains 172 of 240 original
lines by exact SequenceMatcher alignment before final small edits. Replacement
lines chiefly correct tool/launch contracts, architecture and forward numbering;
additions explain demonstrated behavior and provide a runnable small skill graph.

## Findings and dispositions

| Finding | Disposition |
| --- | --- |
| Stale eight-tool count and undefined `think` | Avoid the opener's count. Full primary lists the twelve implemented tools, replacing `think` with inherited `tool_limits`. Unknown names create no handler. |
| Missing load/unload argument and launch contract | Publish both `name` JSON arguments, CLI prompt/chat boundary, separate GUI launch and revised grader entry point. Environment belongs at application construction. |
| Optional per-Agent identity | Explain registry ownership/attachment, shared directory versus independent state, and retained prior public registration without a skill spec. |
| Original behavior in `common` sample | Replace `common.NewVarRegistry` with accepted public `AgentSpec.Variables`/`RegisterVar` excerpt; behavior stays in skills. This is an example API, not a new required spelling. |
| Discovery and activation conflated | Explain no capability at discovery, graph validation before activation, cycle/missing errors and instruction contribution from dependencies. |
| Visibility versus dispatch | Require same enabled set for declarations and dispatch; real hidden-write refusal paired with enabled write. |
| Progressive disclosure ambiguity | Clarify initial offered list versus later rendered bodies, restricted to reachable, discovered, unloaded loadables. |
| Unload deferred | Mark pending only, retain instructions and enabled tools, reject a never-loaded name. No premature compaction/removal. |
| Rendering time/lock lifetime | Initial variables installed before primary render; later registration affects later bodies; explain unknown placeholders and callback reentry without prescribing a larger framework. |
| Provenance and cache overclaim | Preserve first enabling source and actual call sequence. Explain later cache motivation without claiming metadata or a growing tool array proves cache grants. |
| Memory sequencing | Correct forward reference to Edition2 Chapter15, remove implied present `save_memory` tool. |
| Live unload inference | Preserve successful sessions and exact model/request identities; explain why initial read after unload did not prove dynamic retention. Distinguish real declaration test/mutation and revised grader's planned executable write. |
| Historical grading | Retain interrupted/unscored full run and focused65/85 with only undefined-name failures; no completed65/100 claim. New grader results remain separate. |

## Grader agreement and validation

Coder `edition2_grader_coder` agreed original Skills weights10/10/10/15/15/10/10/5,
plus corrected Edition2 Chapter8 parity15. All skills phases use actual GUI
WebSocket input; full-primary stdin-to-GUI assumption is corrected. Strengthen
body freezing, real dynamic dispatch/pending retention and instructions/variables
without creating a new product protocol. Graph atomicity/cycles, provenance and
Agent isolation retain focused source evidence rather than structural name checks.
Command: `go run ./cmd/grade-edition2 -ch 9 path/to/core`; existing GUI and inherited
fixture flags remain. Exact snippets and TL;DR are with the coder for validation.

Author checked raw retained live receipts: Claude5 requests, OpenAI4, Gemini5;
all exit0/no timeout, expected actual marker files, dynamic load instructions and
unload call. Initial extraction missed Gemini's camel-case `systemInstruction`;
explicit follow-up checked all five actual field values equal. No empty-field
comparison is relied on. No new paid sessions or current model-availability
claims are made. The live launcher uses human chat input through pipes; no claim
is made that Bill personally drove a PTY or listened.

The published small skill files include description metadata for compatibility
with the original parser; the historical live fixtures omitted it. The manuscript
identifies the added descriptions/body wrapping rather than claiming identical
historical bytes. The revised grader's first legacy attempt also omitted that
metadata and failed at setup; the coder records its correction and later actual
behavior results separately. This is a fixture defect, not missing legacy Skills.

Final direct chapter lint passes all hard checks at2,275 prose words, with only
the soft lower-length warning. Original-to-enhanced exact-line alignment retains
172/240 lines; the submitted chapter has458lines, SHA-256
`804d37e6449f38421c0ad26b09835ee3301fc9caed0bff0b307f08ccc34d1393`.
No padding is warranted. Initial independent prose review of `e533aa35` found no
blocking issue and separately verified all three raw live receipts and source
claims. The grader report link awaits the coder-owned file; other links resolve.

Coder validated both exact Go fences in an external module: construction and
`RegisterVar` run through the real Agent HTTP loop, delivering load-time text
while preserving system text. All three exact published skill files drive the
actual CLI prompt through load/write/read/unload against a small external fake,
five requests and actual `note.txt=SKILLS_OK`. Full-primary YAML declares all
twelve tools; the first frontmatter example loads with small explicit dependency
and offered-skill companion files. All processes exit0. Author checked the
[Go result](../../solutions/evidence/ch10/enhancement09/published-go.txt),
[source binding](../../solutions/evidence/ch10/enhancement09/published-samples.json)
and [CLI results](../../solutions/evidence/ch10/enhancement09/published-cli.json).
These are deterministic published-example checks, not new paid or human-chat
demonstrations. No prose correction was needed. Combined independent grader
review and chapter acceptance remain pending.
