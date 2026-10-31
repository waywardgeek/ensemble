# Enhancement notes: Chapter 2

Author: `enhancement_author`, resumed by `enhancement_author_resumed`, October 10,
2026. Initial prose review found no blockers; coder validation is complete and
final independent review is pending. Handoff skill:
`.agents/skills/ensemble-coding/SKILL.md`.

## Sources and story preservation

Read the entire 1,189-line original `book/chapter-02.md`, student/coordinator
notes and comparative review. Original file revision
`e1605916114230570b348ab5961c193fa1cebf80`, SHA-256
`b2ebe8bd541a9be0a025280ab594b34b5ea1378a7eecfa8ec9a314e02c8615fd`.
Read accepted Chapter 2 common types/enums, History loader, Agent/Engine flow,
CLI, and live record. Accepted source remains frozen at `edition-2-ch02-r3`;
no implementation or first-edition file edited. Governing skill, carryover,
workflow, voice and procedure were read in full for this author pass.

The through-line remains the cost of a Claude-shaped interface, and the reader's
opportunity to discover a missing seam before growing three divergent clients.
Preserved the complete opening account, measured first-edition migration,
vendor-migration frustration, mid-turn classifier bug, Gemini compaction/handoff
story, redaction-summary multiplication, usage-convention absurdity, empty versus
absent live failure, opaque Waywardgeek material and one-field seam bet. The
opening is explicitly identified as the original CodeRhapsody account; its
first-person voice is retained under Bill's enhancement instruction rather than
rewritten to match later voice guidance. Historical observations remain dated
claims of that edition, not fresh product verification.

## Findings incorporated

| Source/finding | Change and reason |
| --- | --- |
| Chapter1 architecture/carryover | Usage owner corrected to Engine; new History owner and optional GUI public-API stub placed within existing boundaries. Reducer is behavior in History rather than a method forced into common. |
| Original permanent-addition promise | Replaced absolute no-code-ever-deleted wording with preserving owners and extending/correcting representations. No future-feature requirements. |
| Original permission to copy answers | Own accepted predecessor first; comparative reviewer reads reference after the attempt. |
| Coordinator missing-contract stop and erratum | Explicit commands, CH02_LOG save/load meaning, full environment names, exact ephemeral string and ack, EOF usage shape, no invented tool/redact stdin directives. |
| Erratum model/surface ambiguity | Requested routing model distinct from resolved target and returned producer; canonical Surface strings and the one historical Gemini input alias supplied. |
| Public log absent | Added compact complete 9-event JSONL example with two calls, interleaved human text, both results and a redaction. Explicit envelope/part names and numeric Ref kinds remove need to inspect grader internals. Example is labelled illustrative, not a fabricated live receipt. |
| Student Ref scope conflict/reviewer exception | Media loading/rendering deferred; dump preserves original Ref and content, reducer preserves exact Kind/Locator in stub. Unsupported unredacted media fails explicitly. Original 92/100 remains historical FAIL. |
| R1 Anthropic ordering defect | Exhibit B now includes two results around human text; group results in the completed wire message while preserving log/Context arrival order. Public example demonstrates the case. |
| Parse mutation survivors | Added capture-versus-prewritten-replay distinction, returned identity and Gemini result-name gap. Exact numeric arguments retain meaning (9007199254740993 regression). |
| Ref mutation/source-test gap | Explain that secret-free render alone cannot prove exact internal Ref carryforward. Preserve source tests/review alongside black-box checks. |
| Student's zero Context changes for both new vendors | Clearly separated from historical one-field bet; informed starting shape explains difference. |
| Live script/excerpt VERIFY markers | Replaced first-edition solution launchers with direct build/chat/render instructions for reader's own implementation. Used actual student 324/23 Anthropic totals, nine-way offline replay and 718/650 byte counts from retained evidence. Old illustrative event excerpt labelled honestly. |
| Ephemera self-correction failure | Distinguish pending instruction removal from the model echoing it in retained dialogue; preserve observed purple-kettle denial and initial OpenAI404 without invented diagnosis. |
| Stale score weights | Coder-agreed revised table totals100; ch1parity10, explicit Ref rows, old four media-render points assigned to stronger ref-redaction8. Historical results not relabelled. |
| Numbering | Current-scope actor/interrupt points to new Chapter5; skills referred to by feature name beyond bounded pass. Evidence IDs stay original. |

## Checks and limits

No new runtime requirement or live API call. Changed code blocks are a complete
public log plus invocation/examples and actual recorded usage; the original Go
type/interface examples remain, with explicit permission for the accepted simpler
tagged Part design. All original conceptual tables remain except corrected grade
weights/additional existing Ref checks.

The [coder's completed validation](enhancement-02-code.md) confirms the published
command and 100-point table: revised grading passes both accepted and legacy
Chapter 2 at 100/100. Original grading remains accepted 92/100 FAIL and legacy
100/100 PASS. The author reconciled these distinct results without changing the
chapter's historical account.

The coder extracted and executed the exact public nine-event fixture, SHA-256
`06fc7b51af868ce7386008bdee4faf7daf2987fac11901b720b26c5848f9bff9`.
Dump preserved the original secret and all events; all three offline renders
removed the secret, kept a nonempty stub, second result and human text, and
repeated byte-identically. Anthropic placed both results before the text.
These new sample checks are separate from the historical live-log byte counts
in the chapter. No prose mismatch or additional live run was needed.

`go run ./cmd/lintprose book/chapter-02.md book/edition-2/chapter-02.md`
comparison: original and corrected enhanced draft have no hard failures. Original
prose count8,473; enhanced9,793. Both exceed the soft length/person-gap budgets.
Most added text supplies the contract that stopped the original student; the
rest explains demonstrated architecture/corrections. No cuts to sound original
stories or padding with human mentions were made to satisfy a counter. Reviewer
must assess whether the additions earn their place and retain enjoyment.

Initial diff against original: approximately250 added/85 removed lines; over
1,100 original lines remain. This is a source-relative comparison, not a diff
against the discarded Edition2 attempt. First-edition text and all historical
graders, exports and actual scores are preserved.
