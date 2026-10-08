# Source checkpoint records

The current implementation is Chapter 5. Its runtime was reviewed at `959c663`
and live evidence frozen at `469730f`. Current acceptance belongs in the
[Chapter 5 validation record](../../../book/edition-2/chapter-05-validation.md).
The coordinator's subsequent changes to this file and README are documentation
only. The dated Chapter 2 record below is retained as history; its next-action
instructions describe that earlier checkpoint.

## Historical Chapter 2 reviewed student checkpoint — 2026-10-07

Code checkpoint validated: reviewer accepted the comparison revisions; all module
checks, inherited grader and independent offline acceptance/audit pass. Manuscript
and Bill's editorial approval are separate. The initial student implementation
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35` remains preserved, derived from Chapter 1
`75542c1`.

All three actual CLI backends and all three external-consumer demonstrations
succeeded. The public consumer exercises actual typed calls, supplied results,
redaction/reference replay, observations, unsubscription, independent Agents and
per-model usage after changing configuration. No tool execution is claimed. The
GUI remains an optional separate-module stub with a public integration test;
there is no browser/WebSocket transport claim.

The post-run comparison led to two improvements: internal admission/sequence
checks no longer clone growing history; malformed logs report useful static
categories and line numbers without echoing private data. Public ownership
boundaries remain intact. WHY comments explain consumption and publication order.
No paid calls were repeated for these internal/diagnostic changes. Earlier owner
helper omissions, loaded-config aliasing, and Gemini schema mapping were fixed
before the initial checkpoint; their discovery and receipts remain recorded.

Validation: gofmt printed no paths; vet and tests passed in the main, GUI and
external-consumer modules. Corrected inherited grader100/100. Independent revised
offline CLI acceptance44/44; passing mutation control plus ten exact detected
defects. Audit source hashes match the committed revision files. Coordinator full
root regression passed exit0 (internal/grade513.863s); root vet/formatting were
reported clean. Exact reports and source hashes are under evidence/ch02/.

The original inherited95/100 result remains preserved. Its round-trip fixture
tried to render unresolved calls; root corrected the fixture by explicitly
completing only unanswered calls in a separate input while retaining dumped data.
The student did not weaken the required refusal, read grader implementation, or
copy a first-edition answer.

SOURCE-EXPOSURE.md records a material process limit: there are no known direct
old-answer reads, but inherited historical summaries and compaction mean this
run cannot certify strictly blind regeneration. Implementation validation and
that context-isolation claim are distinct. No historical skill links were followed.

No credentials are stored here; all working-tree files passed an in-memory scan
against the actual three keys. No running student processes or pending approvals.

Next: root/author complete manuscript review. Do not begin Chapter 3 in this
context. The next cold chapter uses a fresh coder with no inherited turns and
explicit new-edition inputs. Reload the full mandatory skill before any new fix.
