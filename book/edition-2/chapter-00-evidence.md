# Chapter 0 evidence and reconciliation

Status: complete execution-guide draft, independently reviewed and ready
for use. Historical external claims are omitted; verify before any future reuse.

## Authority and reads

Bill explicitly requested Chapter 0 enhancement after the earlier instruction
to leave it alone. Read the existing `book/chapter-00.md` in full, current
full voice guide, full chapter-writing procedure, and full mandatory coding
skill. Current architecture/workflow and the new Chapter 1/2 contracts inform
the practical directions. Root is updating procedural scope to reflect the
new instruction; it is not treated as an unresolved permission gate.

Created only the separate second-edition chapter and its outline/evidence.
The first-edition chapter, preface, legacy code, and solutions are unchanged.

## Preserved and corrected material

The first-person opener is an excerpt of the existing Chapter 0 account:
the prompt to Astra, the author's assessment of its result, and the reason
for recording experience in the book. No new quote, private story, timing,
or benchmark is introduced. Removed the current-best-model assertion and
unverified external counts from the running argument. Verification of those
historical claims is required before any future reuse. No fresh measurements
or universal claims about models' capabilities are substituted.

Reconciled old “build code first, write prose later” and “score 100, move on”
directions with the current procedure: teach first, cold student build,
checks/deletion audit, actual live use, independent first-edition comparison,
coder/author revision, proofreading, then validated checkpoint. The comparison
happens after initial implementation/run and preserves the initial attempt.

Added the mandatory full coding-skill read and reload instructions with the
actual repository path, architecture rationale, separate optional GUI module,
public client boundary, a complete sample handoff, roles/phase scheduling,
question routing, credential-agnostic live guidance, separate new Git histories,
artifact map, and durable restart state. No assumption that all workers remain
resumable or that an entire textbook stays pinned in context.

The ordinary reader does not need a publishing team. Independent roles are
explained for cold evaluation and authoring/extension, with the coordinator
able to perform grading and the reviewer scheduled in phases. The chapter
does not promise every model will score 100 or every new model will produce
a better design. Passing checks establish only the exercised properties;
comparison and real use provide additional evidence.

## Verification

Scoped `go run ./cmd/lintprose book/edition-2/chapter-00.md
book/edition-2/chapter-01.md book/edition-2/chapter-02.md` exits 0 with no
hard failures. Chapter 0 has soft length, negation-density, and person-gap
warnings; the execution guide is not padded to meet a chapter word target.
The reviewer read the complete new chapter, original chapter, outline/evidence,
and current procedure, then accepted the execution guide without material
revision. It verified the source excerpt, reader/team distinction, complete
handoff, question routing, review gate, and limits on persistence/enforcement
claims. The coordinator subsequently requested removal of the visible
publication marker because none of those old figures is reused. That marker
is removed from the manuscript; verify the historical external claims before
any future reintroduction. There is no new program in this
chapter to claim tested or live-run. Other chapters retain their implementation
and evidence gates.
