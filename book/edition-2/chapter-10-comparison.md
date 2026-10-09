# Chapter 10 historical quality comparison: round 1

**Round 1 ready for coordinator handoff.** Initial live experience is now frozen
at `44d7627`; the coordinator authorized comparative feedback and accepted the
Q1–Q3 rationale. No coder repair has started. Q4 below adds the independently
diagnosed empty-card presentation issue from those original live receipts.
Findings remain unresolved until the responsible coder/author replies and the
reviewer checks the revisions. This is not chapter acceptance.

The original preparation at `1a173a0`, dated 2026-10-08, explicitly left the
initial real-provider spin pending and withheld feedback from student and author.
The coordinator authorized that source review after deterministic clearance.
That earlier exposure boundary and draft remain preserved in Git; this update
does not relabel the preparation as a post-live review.

The reviewer previously owned independent Chapter 10 deterministic grading and
has read its checker internals. The reviewer did not implement the student
runtime. This comparison used source/evidence reads only: no Go build, provider
request, credential access, runtime edit or historical code copying.

## Sources and scope

New Chapter 10 corresponds to first-edition Chapter 11. Compare persistence at
that feature boundary, using later repairs as evidence of problems worth
preventing. Later features and old recovery policy are not new requirements.

| Source | Immutable identity / scope |
|---|---|
| New runtime | `57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`; `solutions/edition-2/main` tree `b3f950ced3d4c7068cfeacdc05c0c6201e1e34fc` |
| New reviewed extraction | All 159 runtime/API/asset files freshly matched to that commit and the [final required map](checkpoint-evidence/ch10-final-required-map-57d4aac.json); support/evidence files are outside this runtime map |
| Old standard | `c24b8e84da112ef66f729c0cdc60413ee6fb547d`, `solutions/ch11` tree `b2f27d35cacd241d1882d7146f2456d514463692`; unchanged at review start `78f4f6b14f9a5a89497470c1075262f39db93ce2` |
| Old teaching | `book/chapter-11.md`, last changed at `994bf218b31033b7dbf2740dd0187d509d2927dc`; SHA-256 `ed6482dd1e5247880122382a3b4199276d8fbda13ea8863676eda5c6e49e9df0` |
| Current contract | `book/edition-2/chapter-10.md`, SHA-256 `1fa6c5448fdd07d0685b634254015bdce62815141803c1ec69ee65e5f1d28ef1` |
| Current outline | SHA-256 `efdd99dc2fd74afaa31c0833fa1dedc3ef318856dfd32c9b8eaba4517f6bfb88` |

The entire coding skill, architecture ledger, voice specification, writing
procedure and new Chapter 10 contract were freshly read for this task. Locations
below refer to the bound versions, not future working-tree revisions. No source
from Bill's concurrent first-edition sandbox work was modified.

An initial overbroad binding attempt enumerated every extraction file and stopped
at generated `test_ch06.log`, which is not source in 57d4aac (`git show` exit 128).
No source mismatch was accepted. The corrected read used the existing complete
runtime selection and verified all 159 files in both directions against Git and
the accepted evidence map. No historical test result was rerun or relabeled.

Relevant later historical corrections were inspected at their actual revisions:

- `8b711a3c19c025fba39c6a6cbf1f6bce9312b9f7` moves Context/log/save behavior
  out of common into llm using free functions. The old standard's convenient
  receiver methods had put implementation in the vocabulary package.
- `72196ddc2bab1219a17730814186784aa5bc5cf3` adds signal selection around
  stdin completion so normal shutdown can reach the save. Old Chapter 20's
  account is consistent with that source change.
- `a1eebb4ed23162b76f4b5d25265ef17238b264c9` restores missing replay part IDs
  and thinking display in `agent/internal/ws/handler.go`; its added regression
  tests distinguish multiple responses and thinking on reconnect.
- `ace699ef7d890d30fb2fb9ce292a99bf2c5adec2` diagnoses unanswered calls at
  restore while retaining the later first edition's permissive recovery policy.
  That policy is evidence of a real lifecycle problem, not authority to synthesize
  missing outcomes under the new chapter's explicit unfinished-session refusal.

## What improves, and what it costs

**Ownership and extension.** Old `internal/common/save.go` owns file I/O, Restore
and sequence allocation. Its CLI constructs the Engine and installs `Ctx` and
`Log.Events` directly (`cmd/main.go:282`), while the public Agent exposes its log
and exports concrete GUI construction. Persistence is therefore hard to reuse
without repeating executable wiring. New `internal/common/session.go` contains
values and owner interfaces; persistence owns codec/files/lock/worker, llm owns
semantic validation and actor decisions, and Skills validates its own historical
transitions. Public OpenSession, InspectSession, ExportCheckpoint and ImportSession
serve the CLI, separate optional GUI and embedding consumers. This incorporates
the later common-to-llm lesson when persistence first appears.

`SessionStore -> SessionAgent -> Ensemble` gives storage a real parent and logger.
Checkpoint I/O and its file wrapper retain that chain, while permitting focused
failure controls. `sessionReader` streams accepted records back to its owning
Agent instead of importing a sibling reducer. The root's 806-line `session.go`
is substantial: it coordinates cross-owner identity, material references and
publication as well as construction. There is no evidence that replacing it
with a new framework would improve this chapter. Keep responsibilities visible
when later features extend it.

**State and authority.** The old four-field SaveFile and its single replay loop
are admirably easy to inspect. They also trust a decoded Context and do not prove
that its anchor, calls, skills and available log describe the same conversation.
New semantic state carries the necessary provenance, identity and reduced facts,
with one material dictionary for Skills bodies. Full-origin loading independently
reduces the available prefix before installing the snapshot; explicit import
keeps immutable origin bytes and admits only the actual anchor/tail as raw history.
Current handlers/catalog, policy and credentials have separate authorities.
Historical jobs remain useful evidence without gaining process control.

The extra state is justified by those promises. The new code does not claim
faster startup, unbounded retention or automatic crash-work recovery. An
unfinished conversation is inspectable but cannot quietly acquire an invented
tool result. That is an intentional difference from the later first edition.

**Failure and lifetime.** Old Save uses a same-directory temporary and rename,
which is a useful design to retain, but its Write checks only the error, omits
Sync, and the CLI prints a save error without making it determine the returned
failure status (`cmd/main.go:498`). The new Store checks short/write/sync/close
failures, names replacement as the commit, and propagates close/save failures.
The actor captures owned state at a settled boundary; one worker writes it while
subsequent events become the tail. A writer lock lives through joined shutdown,
and both terminal clients reach that close path. The deterministic evidence
covers the actual signal, fault, job-tail and surviving-child lock cases; this
draft does not substitute those results for the pending real-provider spin.

**Comments and clarity.** Retain the old explanation of `as_of`, the deliberately
nonempty tail and the reason corrupt saves must not become fresh sessions.
The new chapter carries those explanations forward. New comments explaining
close-on-exec, prepared record ownership, immutable material, replay Raw spelling
and result-list presence explain actual invariants rather than narrating syntax.
The old comment that a save is ordinary data rather than secret because it omits
the API key is misleading: conversation content can itself be sensitive. The
new teaching explicitly makes that distinction and creates private store files.

The reflection codec avoids a large duplicate DTO hierarchy and publishes its
closed grammar. Its cost is coupling the disk schema to common struct fields,
with deliberate exceptions for raw JSON, handler schemas, argument strings,
optional limits and tool-result presence. The two exactness surfaces are not
interchangeable: replay preserves accepted bytes, while identity/hashes use
canonical values. The comments and published codec mostly make this choice
reviewable. Do not replace it merely because reflection is present. The retained
nil/empty unused Part.Parts comparison was a diagnosed test representation issue,
not a reason to demand a new public invariant or another runtime repair.

For scale, `git ls-tree` plus physical-line counts of tracked `.go` files,
excluding `evidence/`, gives:

| Delivered Go source | Old Chapter 11 | New Chapter 10 |
|---|---:|---:|
| Production files / lines, including blanks and comments | 33 / 9,729 | 83 / 16,780 |
| Test files / lines | 5 / 1,012 | 42 / 7,773 |

These whole-tree counts include different preceding contracts and module layouts;
they are not a controlled persistence-size comparison. More locally, old save.go
has 140 lines; the new persistence spoke has 1,250 production lines, alongside
the 806-line root coordinator, semantic validation and actor/Skills extensions.
Calling the new implementation smaller would be false. Its improvement is
explicit ownership and protected behavior; redundant work still deserves removal.

**Tests and teaching.** The old seven-check grader's request equality and older
snapshot/newer-log construction remain strong ideas. The new evidence extends
them to three routes, prefix-free import, authority, exact payloads, write faults,
public clients and real file limits. The [final deterministic review](chapter-10-final-deterministic-review.md)
records the actual joined 70/70 retained result and the original 68/70 attempt,
the narrow compatibility corrections, prior-source deletion scopes and the
inapplicable old harness's 0/7 diagnostic. No aggregate score is used as proof
of quality. Student tests add focused store-fault, capture-tail, import and
append-failure regressions close to their owners.

The chapter retains the documented default-load incident and the consequence
that made it matter. It teaches why restore cannot restore authority or workers,
why current policy wins, and why SIGINT must reach the same close path. The
strict contract is much larger than the old chapter, but most added detail
resolves observable ambiguity rather than inventing features. The live section
correctly remains a reproduction plan. Its eventual narrative and the student's
initial usability report cannot be judged from fake-server receipts.

## Consolidated findings, round 1

**Q1. Avoid constructing an entire Skills snapshot to read one watermark.**
New `session.go:111` calls `a.skills.Snapshot()` inside watermarks and keeps only
LastID. `captureSession` calls watermarks at line 138, then calls Snapshot again
at line 144 while holding the Agent lock on the actor capture path. Snapshot
calls Inspect, sorts material, copies activation collections, derives contributor
lists and copies every retained transition's state. This repeats real work for
every Skills checkpoint/export; immutable body strings themselves remain shared,
so this is not a claim of two full body-byte copies. Installation and final
allocator-floor calculation also use the expensive scalar query.

Use the already captured owned Skills snapshot when computing checkpoint
watermarks, or expose the scalar through its actual Skills owner without building
an inspection. Preserve one actor boundary, complete transition/material data,
exact represented maxima and returned-buffer ownership. A short invariant comment
would help keep a future scalar query inexpensive. This is a localized quality
improvement, not a newly discovered correctness failure or a demand for a new
large benchmark. Reuse the affected capture/import/Skills tests after any change.

**Q2. Make the published watermark definition say equality in its first use.**
New `persistence-format.md:39` says activation/job are “at least represented
durable maxima”; its final clarification and Chapter 10 §§10.3/10.8 require exact
equality, and `session.go:338` enforces equality. Correct the earlier sentence,
including zero when absent, and leave the request cursor's distinct lower-bound
rule clear. This is a documentation contradiction with a known correct runtime
rule. It needs no changed expectation, runtime fix or repeated provider run.

**Q3. Narrow the goldfish opener to the actual missing capability.**
New Chapter 10 lines 3–6 say the process “forgets everything” when it exits.
The preceding edition-2 implementation already persists events and supports
offline rendering. The outline explicitly identified this mismatch, and the
chapter explains the distinction later. Preserve the returning-reader stake and
the memorable goldfish image, but make the opening about inability to resume
the recorded conversation safely. This is a small teaching correction, not a
request to discard the old story or add a new one.

**Q4. Present empty accepted response text honestly without a blank Answer card.**
This is a real presentation-quality finding from the initial Gemini live use,
not missing model text or a persistence failure. The original
[B screenshot](../../solutions/edition-2/main/evidence/ch10/live-20261008/gemini-B/browser-1.png)
visibly contains an Answer heading, Accepted status and Speak button with no
answer body. The [restart DOM receipt](../../solutions/edition-2/main/evidence/ch10/live-20261008/gemini-B-restart/browser-1.txt)
contains four such empty cards alongside the two actual answers, tool work and
request outcomes. The restart screenshot is scrolled to later cards; the full
DOM and socket receipts establish the earlier cards rather than inventing what
is visible in that viewport.

The closed B store and original socket records identify the same four empty
text parts at zero-based position 1 of responses 5, 10, 15 and 20. Responses 10
and 15 contain signed empty text; responses 5 and 20 contain ordinary empty text
after tool calls. The actual answers at position 0 of responses 10 and 15 retain
147 and 171 characters respectively. It would be inaccurate to label all four
empty parts as opaque metadata or to conclude that the answer was lost.

Responsible owner: optional GUI `gui/web/gui/artifacts.js`, specifically
ArtifactScroll.part at line 91 and its Artifact rendering. Both live final
mapping (line 85) and snapshot response replay (line 99) use that path. It assigns
Answer to every text part and creates the full card unconditionally. The Go
`gui/projection.go` implementation correctly preserves present empty text and
ordered positions while stripping bound opaque metadata. Chapters 6 §6.2 and
7 §7.4 expressly require those values; removing them upstream would regress the
contract. Chapter 10 retains their historical coordinates on resume.

Expected correction: for an accepted response part whose type is text and whose
present text is exactly the empty string, use a compact, explicit empty-text
presentation, such as “Empty text part”, at the same ordered durable identity.
It must not look like another blank Answer or offer Speak/expansion controls
that have no content. Keep the tracked card identity, including provisional-to-
durable mapping, and use the same treatment after restart. This can be a compact
Artifact state; it needs no new model, storage or wire abstraction. Do not use a
general truthiness or whitespace-trimming filter: nonempty strings stay exact.
If all response text is empty, keep an honest visible empty-content indication
and the actual request outcome rather than manufacturing an answer.

The correction must retain the original parts and signatures in library/history/
checkpoint/request reconstruction, retain safe empty parts on the browser wire,
and leave explicit opaque placeholders visible. It must not hide ordinary empty
tool-result reports, their error flags or job lifecycle/status. Those reports
have a distinct event presentation path and are not empty assistant answers.
Keep empty text distinct from absent content; no fabricated metadata label is
needed to remove the misleading blank card. This is a quality refinement of the
Artifact presentation, not a retroactive claim that Chapter 7 already prohibited
its current card shape.

Finite local verification after repair, using the captured data and no new paid
request solely for this finding:

1. Feed the original B/restart snapshot events and captured live final into the
   public ArtifactScroll browser fixture. Assert the four durable response keys
   keep their positions and explicit empty-text presentation; the two real
   answers and tool cards remain unchanged. Exercise the captured live-final
   path separately from snapshot response replay, then reset twice to check
   identical keys/order and no duplicate cards. Replacing an empty presentation
   with nonempty text at the same component key must restore ordinary rendering
   and its usable controls, without leaving the compact-empty state attached.
2. Keep the existing signed-empty/opaque recursive projection fixture unchanged.
   Add only the focused presentation controls: an empty result report remains
   visible with its lifecycle, an opaque placeholder stays labeled, nonempty
   text is retained, and an empty final does not leave an old provisional Answer
   card or speech controls behind. Assert no automatic or manual speech dispatch
   for the empty presentation while nonempty speech still works. Check the DOM
   and one local screenshot.
3. Reuse affected GUI/browser component and projection regressions under the
   repaired source; retain any original failing fixture result. Original live
   screenshots, DOM, socket records and signatures remain untouched. A local
   captured-data rendering is new local evidence, never a corrected live receipt.

Source and receipt binding for Q4: runtime
`57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`, live support
`9822b2b63257724001d7af195b377483baad3531`, evidence freeze `44d7627`.
The files below were matched byte-for-byte to the frozen Git evidence; the browser
files also match their original sealed manifests. Both responsible source files
were freshly matched to 57d4aac.

| File, relative to live-20261008 unless stated | SHA-256 |
|---|---|
| gemini-B/browser-1.png | `b711f749a0c74bddf3f79c7d0d82cf6fee4ef6f2f1762d5b73f0893b9d8d1857` |
| gemini-B/browser-original.jsonl | `19d70f5cb19828fc8086a9495919ca76222d95fdd167077928a23330cbcd1037` |
| gemini-B/closed-store/events.log | `81e2904c65278199d632f1303afb3fe0f156d3f36cfa1b65aea93ebe8fed749e` |
| gemini-B-restart/browser-1.txt | `783c47b7aa23f6afde3c9d85dc3808fc12e4c7a16c8426c4fbcb9bbdc6fb4df5` |
| gemini-B-restart/browser-original.jsonl | `716c0397884c6d08f32f722b91df641d3353727515b0f87b72beb1bf55df69c9` |
| main/gui/projection.go | `f93a4ecdabc5b61791507e04f8e850bf4606b3a1ebf2e6e30278f82784a2a884` |
| main/gui/web/gui/artifacts.js | `09dec5f6016d053601c61986b9dd001fd07370c51ba4ef8373da546fb329fbba` |

The read-only diagnosis initially guessed `project.go` rather than `projection.go`
and tried a linewise JSON reader on the pretty-printed originals manifest. Those
exploratory reads failed before any write; corrected reads above did not alter
the originals. No build, runtime edit, browser replay or provider call occurred
in this diagnosis. The verification steps are a plan, not claimed test results.
The per-feature live reviewer retains acceptance ownership and received the
diagnosis; feedback to the student remains the coordinator's grouped handoff.

No new persistence correctness blocker was established in this comparison. Q1
needs a fix or reasoned quality disposition; Q2 needs the local specification
correction; Q3 belongs to manuscript reconciliation; Q4 needs the narrow
presentation repair and local captured-data check. None has been resolved by
this reviewer. Initial live experience is frozen, so the coordinator can now
release the grouped rationale, review revisions and affected validation, and
finish the comparison gate. Deterministic clearance remains scoped to 57d4aac.
