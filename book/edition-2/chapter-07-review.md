# Chapter 7 contract review

Coordinator review of the complete proposed chapter and the clarification at
`5890de8`, October 7, 2026. The contract is accepted for preparation of the
student handoff. Implementation still requires accepted Chapter 6 and a
published independent acceptance command; neither gate is waived here.

The reviewer read the current voice and chapter procedure, the full new
chapter, outline and evidence ledger, and the full first-edition Chapter 8.
The historical teardown account was checked against commit `352b590`.
This is a prose/contract review before implementation, not the later comparison
of completed student code with a historical answer.

The rewrite retains the radio/receiver explanation, Artifact lifecycle and
human need to finish a correction before the next action. It gives that need
an observable boundary: an applied pause acknowledgement before a later tool
admission. The old shared Boolean could not identify a second client's cause;
the new Agent-owned registrations do. The actor remains available for interrupt,
close and job facts while admissions wait. A tool already admitted continues,
so the text makes no false promise to suspend an effect halfway through.

The contract places GUI transport and components in the separate optional
module and reaches the Agent through public interfaces. The atomic watch
supplies the missing snapshot/tail boundary, owned values, complete identities
and explicit overflow recovery. It retains a declared window rather than
claiming the transient observer queue is a durable recording. CLI reuse on the
same Ensemble and the fixed last-100-event window are accepted coordinator
working choices; these do not claim additional rulings by Bill.

The initial full read found two concrete gaps. Other clients needed an exact
live pause-change record, including count changes that leave the aggregate
true. Browser-safe projection also needed a rule for nested typed part lists.
Revision `5890de8` resolves both with payloads, ordering and valid fixtures.
It includes pause changes in the watch watermark and represents an active
operation before its first fragment, so reconnect does not invent a second
begin. Projection preserves order and empty text, removes bound replay material
recursively, and retains an ordinary tool argument named `opaque`. The result
fixture respects Chapter 2's permitted result-child kinds.

The teardown story names a concrete send/close race and its consequence;
the requirement uses a controlled interleaving rather than an unsupported
promise about repeated random tests. Speech teaching likewise distinguishes
queued work, actual synthesis, cancellation generations and an audible result.
The old universal browser workaround is excluded until a versioned failure
justifies it. User text remains data in the browser; safe projection is separate
from HTML rendering and from model prompt handling.

The proposed spin is plainly labeled pending and covers actual browser,
terminal and public consumers on all three APIs. It requires real speech
evidence or an open blocker and preserves model noncompliance as an outcome.
Final prose reconciliation, actual live evidence, student feedback, grader
coverage and independent comparative code review remain required after the
initial implementation. No Chapter 7 code, passing chapter grade, browser
session, publication or editorial approval is claimed by this review.

## Browser tooling preparation

While Chapter 6 repairs ran, the coordinator installed pinned Playwright
1.64.0 outside the repository at
`/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/`; its local
package lock records the dependency. The installed Chrome 154.0.8037.93 launched
in an isolated headless session, rendered a local test button and accepted a
real locator click. Its speech API enumerated 180 local voices. No speech was
requested, no audio was verified and no Ensemble browser implementation exists
yet. This only establishes an available browser driver for the later student;
it cannot satisfy the chapter's real synthesis or GUI gates.

The driver uses the documented [Playwright library lifecycle](https://playwright.dev/docs/library)
and [Chrome channel launch option](https://playwright.dev/docs/api/class-browsertype#browser-type-launch-option-channel).
No dependency or frontend framework was added to the student source. Keep the
browser choice separate from the student's GUI implementation choices.

Chapter 6's comparative review found quadratic copying in the initial stream
assemblers and repaired it before acceptance. The coordinator carried that
lesson into §7.3 before a Chapter 7 student exists: the recoverable presentation
projection also accumulates changes incrementally and materializes a snapshot
at the watch boundary. Its owner, semantic bounds and wire contract are
unchanged; reconnect support must not reintroduce that cost on every delta.
