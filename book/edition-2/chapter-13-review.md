# Chapter 13 independent contract and prose review

Review date: October 8, 2026. Reviewed the complete frozen draft at
`e4ef9a9299b570fff615c1567fcd42d3951dd70b`. Its chapter SHA-256 is
`2316a55ea0027d38be02fddcf5260eb68440e2ff283510cf909c1c2ef6eecc0f`.
This is pre-implementation review. Chapter 12 implementation, Chapter 13 student
work, runtime checks, native audio, model use and final manuscript acceptance
remain pending under the coordinator's validation record.

## Reviewer independence and reads

The reviewer is `/root/coder_ch08`, reassigned explicitly by the coordinator.
Earlier roles were Chapter 8 student coder, initial Chapter 9 checker engineer
and Chapter 11 checker engineer. The reviewer did not author Chapter 13 or
implement a Chapter 13 solution. Those earlier roles are disclosed rather than
presented as a fresh student context.

Read all of the current `book/voice.md`, `book/chapter-writing-procedure.md`,
architecture, Chapter 13 outline/evidence/validation and the frozen manuscript.
Read the entire first-edition `book/chapter-14.md` for story comparison. Refreshed
new Chapter 7 §§7.1–7.2/7.6–7.7; Chapter 8 §§8.6–8.7; Chapter 11 ownership,
transport, correlation/cancellation and frozen-binding/resume passages; and
Chapter 12 §§12.1–12.8, including its exact five-tool endpoint, attachment,
bounded routing, human draft and request-observation rules. Consulted Chapter 6
§6.6 for failed provisional output. No historical implementation, current
student runtime or future chapter was opened for this review. The author evidence
ledger's source-repair investigations remain attributed to that author; this
review did not rerun historical code or recover its missing sessions.

## Consolidated findings

Both findings need resolution before a student receives this contract. No
additional material finding arose from the full read. Locations below refer to
the frozen draft, not subsequent revisions.

### R1: define incomplete journal reads after recording fails

**§13.4, lines 294–301 and 329–337; dependent §§13.5/13.7/13.9.** The draft stops
new recording/admission when a required counter exhausts, but also requires
exactly one recorded terminal phase for every queued utterance. A queued record
at the last sequence leaves no sequence for its terminal record. Earlier
unsettled utterances can exhaust the remaining capacity too. Ring eviction
cannot explain this absence: no terminal record was written to evict.

The success-shaped read/export currently has no persistent status distinguishing
that damaged journal from an intact suffix. Specify what the reader sees and
which evidence claims fail, including at a legitimate maximum-counter seam.
Also distinguish journal-owned counters from the inherited identities required
to fence actual native requests and cancellation.

The author agreed. The coordinator accepted an out-of-band sticky recording
fault, with a diagnostic boundary: **failure to record must not stop native
speech admission/output, cleanup or another Page's speech**. Stop journal
recording without wrapping or reusing identifiers. Every scoped read exposes
the fault; exports and listener evaluations explicitly refuse completeness.
Missing terminal records are then explainable only under that fault. Recorder
delivery that lacks its required completion evidence explicitly fails/settles
its owned request, releases pause and does not invent successful delivery.
Recovery uses a new service lifetime. Actual playback ownership identities
remain valid independently of diagnostic recording; journal fault cannot permit
their reuse.

The author's initial alternative also stopped native admission; the coordinator
superseded it with the diagnostic-only boundary above before revision. Publish
exact status/error fields, nullability, read-result accounting, manifest failure
value and recovery together. Keep file-export failure separate from speech-service
recording failure. Review remains open until those bytes are frozen and checked.

### R2: separate the launcher run from the standalone example

**§13.8, lines 650–664.** The instructions say to use the public example to
construct the listener “on that endpoint” after launching the GUI. The next
paragraph says the example creates its own application and fresh target/listener
Agents. A reader cannot follow both meanings without an unpublished attach mode.

Use two explicitly separate demonstrations: the launcher/PTY/browser run, then
the standalone public example's own URL and fresh application. Alternatively,
publish an actual attach mode if desired, but no new feature is necessary to
fix this prose. The author selected separate demonstrations. Retain both source
identities and avoid claiming that the second observer heard the first target.
Also use the actual adapter name `recorder` for §13.7's “Record mode” sentence.

## Contract and teaching assessment

The owner chain is teachable: Page owns parser/queue and pause causes;
BrowserApplication owns SpeechService; that service creates the adapter and
owns diagnostic records. The recorder has no independent conversation reducer.
The native adapter retains the same-origin lease, waiter cancellation and
application-close order. Stream/final identities, canceled provisional text,
autoplay-off consumed positions and captured rate/revision remain explicit.

Manual full-card speech is preserved. The automatic 8,192-byte pending limit
does not clip a manual utterance; the journal's prefix, byte counts and digest
describe any omitted diagnostic text honestly. The mandatory oversize-manual
positive checks full adapter input separately from journal completeness.

The distinct listener attachment extends the outer GUI protocol without adding
tools to Chapter 12's exact-five endpoint. Target and listener identities can
coexist; shared transport/operation budgets count both. Creation-time three-tool
rights, fixed completion-only profile, dispatch denial and no automatic target
observations supply enforceable isolation. Programmatic input preserves human
draft refusals and does not claim trusted typing. Closed-Page diagnostic leases
and replacement mounts have explicit separate lifetimes.

The bounded grammar teaches its order and limitations before exact arrays.
The huge discarded fence, accepted end, exact/+1 token budget, earlier admitted
utterance and next-part recovery distinguish streaming from whole-buffer
substitution. The proposed independent checks require real event wiring and
application ownership, with pure normalization tests as complements. No private
JavaScript spelling or historical checker score becomes an unpublished API.

## Story, voice and evidence assessment

The old chapter's useful story survives: the listener obtained a correct exit
code through narration, so task success did not establish result delivery.
The single-chunk fence control then defeats the tempting stream-boundary
diagnosis and explains why grammar recognition must precede newline splitting.
The new telling retains the reader's dependence on speech while removing old
claims that a transcript persona simulates blindness or establishes native
hearing. The first-edition unrestricted/default logging, post-construction tool
removal and different Escape behavior are not reproduced as new requirements.

The opening identifies a person with something to lose, and the service,
missing-sentence, failure and final-use passages return to that stake. The
long journal/listener middle is necessarily detailed, but it develops a concrete
testable surface rather than repeating the motivation or inserting unsupported
anecdotes. Paragraph endings and explanations vary enough to retain the engineer
voice. No additional story-padding request is warranted by the linter's soft
person-gap warning. The draft carefully labels historical accounts, recorder
output, callbacks and captured audio as different evidence.

The planned human CLI/browser/public/listener/native paths cover the right
surfaces and retain all-provider obligations. They remain a plan. Before spending,
the student must publish actual prompt/HTTP caps and a feature/action/result
matrix for coordinator review; this prose review approves no paid run. The final
spin still needs source-bound actual actions, screenshots, transcript and audio
receipts with their limits. No new successful demonstration can be inferred from
this draft or the old chapter.

## Local prose checks and disposition

Independently parsed both literal JSON blocks and all seven source/output JSON
fixture pairs from the frozen chapter, and verified the UTF-8 `Ready.` SHA-256
and byte count. These checks prove literal validity only, not normalization or
speech behavior. The first review helper accidentally retained a Markdown table
delimiter and raised JSONDecodeError; correcting that extraction made the checks
pass. This was a reviewer helper error, not a manuscript defect.

The author reports a scoped hard-lint pass and 5,953 linter words at the frozen
draft, with a soft person-gap warning. The reviewer performed the complete manual
voice/story read. No retained linter executable was available; the coordinator
directed against a redundant Go compile under disk pressure, so an independent
lint rerun is not claimed. No build, cache growth, provider call or audio test was
performed. Only this review file is reviewer-owned.

Disposition: promising full contract and voice pass, with R1/R2 awaiting the
author's grouped revision and independent verification. This is neither student
release clearance nor chapter validation.

## Independent R1/R2 closure

Reviewed frozen author revision `f3a50d35d274a19ea0c55605cadd5d1b51d233b2` on
October 8, 2026. Chapter SHA-256:
`1e1ea9557466cc3bd79b6d31733be8045425e566ef5e6bff8dd1096ed62cf0ac`.
Reread the complete current voice guide and chapter-writing procedure, the full
three-file grouped diff, and the affected journal, listener, export, checks and
demonstration passages in context. The full first-round review above remains
the basis for unchanged portions; no runtime source was read.

**R1 resolved.** Section 13.4 now separates diagnostic seq/time/utterance
counters from actual playback ownership. Recording failure latches bounded
out-of-band `journal_status` and `journal_error`, retains earlier records and
the last written sequence, and cannot wrap, reset on remount or silently clear.
Both scoped reader profiles expose that status within the response byte cap,
including empty reads at the last cursor. Faulted recording explicitly permits
missing terminal records while requiring actual native cleanup and settlement.
Native admission, FIFO, lease and cancellation ownership remain operational;
diagnostic failure cannot authorize reuse of a live ownership identity.

Recorder deliveries lacking a recordable queued/completion fact fail and settle
promptly, releasing their owned speech work without fabricated completion or
native fallback. The listener consumer refuses complete evaluation even if
earlier completion records remain. Export reports `complete:false` with
`journal_fault`; file-export failure alone does not fault the speech journal.
The new maximum-counter fixture, recovery rule and §13.9 checks agree with these
boundaries. Existing gap, scope-filtering, cursor and text-truncation rules remain
distinct; no contradiction requiring another revision was found.

**R2 resolved.** Section 13.8 now closes the launcher's demonstration under its
own source/service/launch identity, then starts the standalone example's own
application, URL and fresh Agents. It explicitly provides no attach-to-another-
launcher mode and makes no cross-run delivery claim. The adapter name in §13.7
is consistently `recorder`. The public example's construction contract and
Chapter 12 attachment behavior remain unchanged.

Independent local checks parsed all three JSON fixtures, verified the new read
field set and exact maximum uint64 cursor, and confirmed the two earlier JSON
blocks and seven normalization pairs are byte-identical to the initial draft.
The `Ready.` digest still matches. Scoped diff whitespace checks pass. These are
prose/literal checks, not an executed parser or journal. The revised prose passes
the manual voice/cut review without losing the retained historical story or
adding unsupported outcomes. No new hard-lint run is claimed: the author's
initial lint remains bound to `e4ef9a9`, as documented above.

Disposition: **R1 and R2 closed; no remaining material finding in this contract
review.** The draft is cleared for independent checker preparation and the later
student handoff once its accepted predecessor and runnable checker exist. This
does not validate an implementation, native audio, listener performance, live
demonstration, final post-build manuscript or chapter checkpoint. No model call,
build/cache growth or non-review-file edit occurred during closure.
