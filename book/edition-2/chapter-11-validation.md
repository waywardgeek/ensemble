# Chapter 11 validation

MCP contract preparation, October 8, 2026. No Chapter 11 student implementation
or live demonstration exists. Accepted Chapter 10 permits a fresh plan-only
handoff. The actual client-runtime checker invocation must be published after
public seams are documented and before implementation integration or acceptance.
Bill's editorial approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator, reviewer | Draft `a5fc9ea`, clarifications `302d77e`/`39bc526`; plan/runtime distinction `5d6255f` accepted at `d13738e` | Resolve actual student questions through teaching |
| Independent checks | Author `/root/coder_ch08`, reviewer `/root/grader_ch05` | Foundation `ee89051`, repaired at `c80af0a`, independently reviewed at `d13346c`: 90 fixture/oracle controls and nine intended predicate deletions; [coverage](chapter-11-grader-review.md), [review](chapter-11-foundation-review.md) | Add actual student client, custom-transport, lifecycle and persistent-identity controls after public seams exist; preparation is not runtime acceptance |
| Fresh student and owner plan | Fresh CLI student `01a11ec6-f18e-7912-9822-01b975ebd218`, coordinator | Phase1 plan running from accepted `edition-2-ch10-r1` at `d918612`;13 pinned new-only teaching files | Review owner/public API/codec plan before affected implementation |
| Implementation and local gates | Future student, grader | Not started | Implement and check full contract plus inherited behavior |
| Actual use | Future student, independent reviewer | Planned only | Review bounded matrix, then actual CLI/browser/public all-provider runs |
| Historical comparison and revisions | Independent reviewer | Not started | Preserve initial implementation, runs and teaching experience first |
| Manuscript and feedback | Author, student, proofreader | Draft honestly labels spin pending | Reconcile receipts, resolve feedback and proofread final chapter |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable export/tag |

## Foundation command and remaining runtime gate

The manuscript prints `python3 scripts/edition2/accept_ch11.py --self-test --receipt PATH`
as oracle preparation only. Its 90 fixture/oracle controls and predicate deletions
are evidence about the foundation, not a grade for student code. The actual
client-runtime invocation remains unpublished. It must cover the full §11.10
contract, including public/custom transports, concurrency/lifetime and persistent
identity, before implementation integration or acceptance. This staged handoff
does not reduce the required runtime, live-use or independent-review coverage.

## Fresh plan-only handoff

The independent new student started from accepted Chapter10 at `d918612`, source
`265fe34`, tree `5bb92f0`. External ch11-student-inputs retains exact prompt,
new-only manifest, JSONL events and one coordinator inbox. The 13 inputs are
Chapters1–11, architecture and the mandatory coding skill, pinned from the accepted
tag. Memories and inherited conversations are excluded. No earlier/future answer,
author research or grader implementation is allowed. Current work is a compact
owner/public transport/session-codec plan and teaching questions, not runtime code.
The grader prepares the actual client command separately; all required coverage,
real-interface use and later historical comparison remain unchanged.
