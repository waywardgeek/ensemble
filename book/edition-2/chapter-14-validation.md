# Chapter 14 validation

Context maintenance preparation, October 8, 2026. Outline `9758e89` has advisory
review `3a6ae52`; there is no accepted full contract, checker or implementation.
The coordinator has supplied working design choices for the author's full draft.
These are implementation proposals within Bill's authorized scope, not new
personal rulings from Bill. Chapter 13 must be validated before student release.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract and ownership | Author, coordinator, independent reviewer | Outline `9758e89`; advisory `3a6ae52` identifies six boundaries requiring exact teaching | Publish full contract, literal shapes and reader path; resolve A1–A6 before acceptance |
| Independent checks | Future grader phase | Not started | Derive checks from accepted teaching, with genuine positives before runtime credit |
| Fresh student and owner plan | Future new-only student | Not started | Release accepted Chapter 13 source and new teaching |
| Implementation and local gates | Future student, reviewer | Not started | Validate recorded cuts, ownership, compatibility, bounds and inherited behavior |
| Actual use | Future student, reviewer | Not started | Review bounded CLI/browser/public all-provider feature matrix before paid use |
| Historical comparison and revisions | Independent code reviewer | Not started | Preserve initial source, experience and actual runs first |
| Manuscript and feedback | Author, student, proofreader | Outline only | Reconcile actual evidence and resolve student/proofreader feedback |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable export/tag |

The working direction is an explicitly selected fresh v4 context-capable session,
initially disabled maintenance, opt-in tool registration, and a strict v2 extension
of the existing Agent policy. Old session formats, catalog identity and handler
ceilings remain exact; no implicit migration or configuration filtering is planned.
The ordinary CLI user needs a concrete compatible-old-store or fresh-v4 path.

Actor retains the sole durable mutation path. A handoff stages its first valid
intent and later records a committed, refused or canceled disposition; interruption
cancels an uncommitted intent after normal pairing. Logical hint/manual placement
survives paired removal. Automatic compatibility deferral must proceed unchanged
without spinning; explicit incompatible handoff refuses. Original log evidence,
current projection and snapshot-represented history keep distinct promises.

Proposed policy bounds are T=400,000 by default, minimum 20,000 and maximum
16,000,000 neutral bytes, with integer-floor fractions and a 64 KiB UTF-8 note.
Inherited complete encoded-record and semantic-state limits still apply. Protecting
the newest eligible batch can leave a truthful target overshoot. The full contract
must settle exact measurement, selection, keep lifetime, facts and version shapes;
this gate does not supply private assertions for a grader.

MCP remains transport-independent. The GUI WebSocket tunnel stays in the optional
module; context policy must work through headless public interfaces. See the
[advisory review](chapter-14-review.md) for all unresolved draft obligations and
[research record](chapter-14-evidence.md) for attributed historical sources.
