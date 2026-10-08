# Chapter 7 student teaching feedback

The initial implementation is preserved; current gate status is in the
[validation record](chapter-07-validation.md). The student retains its actual read
ledger, plan, attempts and interpretation in
`solutions/edition-2/main/evidence/ch07/student-review.md`. This record captures
teaching dispositions, not chapter acceptance or historical code comparison.

## Malformed transport versus correctable command

The initial partial checker reported 10/15. The student identified a real
ambiguity in §7.4: “Close malformed transport” could include unknown fields,
but the printed wire example returned a correlated error for an unknown field.
Its implementation kept a valid JSON semantic error connection usable. Two
other failures required same-socket pause observation before acknowledgement,
which the published actor-order contract had not required. An oversized-message
case also encountered a send-side broken pipe; the grader is investigating
that fixture rather than declaring the implementation wrong from that result.

The coordinator supplies a narrow author clarification while the author thread
is idle for grader capacity. Section 7.4 now states the boundary explicitly:
invalid JSON/UTF-8, binary, oversized and non-object messages close, as does an
unusable command ID. A JSON object with a nonempty string ID receives a
correlated `invalid_command` for unknown command/field, missing required field
or bad semantic field value, without contacting the model or closing the
connection. Well-formed Agent refusals also retain a usable connection.

This preserves the printed example and teaches the chosen behavior before
affected repairs. The initial failed checker receipt remains evidence of the
ambiguity. The grader must strengthen the distinguishing valid/error/recovery
cases and remove the untaught pause-frame order assumption. The student must
read the clarification and record whether it resolves the difficulty; results
and final author/proofreader reconciliation remain pending.

The student subsequently read `1bba1bc` and confirmed the distinction resolves
the difficulty in its retained review. The updated independent wire checker
passes 23/23, including exact message-size and correctable-error recovery
controls. The oversized fixture now reads the close frame directly rather
than treating its own automatic close reply's broken pipe as a server failure.
The initial failed receipt remains intact. Final author/proofreader review
still follows the completed demonstrations and comparative review.

## Preserve the speech contract through the prose pass

External editorial revision `0f05359` restored personal motivation and stronger
explanations in Chapters 7 and 8. Its replacement of the Chapter 7 speech
opener also removed explicit automatic-speech scope, silent replay/results,
full-card action, unavailable-synthesis behavior and opaque-content exclusion.
The coordinator restored those requirements immediately after the new opener,
and restored the Chapter 8 autoplay control's required help text. The initial
student read the preceding complete contract; this preserves that same teaching
for the next reader without undoing the editorial work.

## Author reconciliation before paid demonstration receipts

The temporary Chapter 7 author read the student's complete retained teaching
review after source freeze `da162e8` and support repair `ba902b7`. The student
explicitly confirms that `1bba1bc` resolves the malformed-transport versus
correctable-command ambiguity. The chapter retains that distinction and its
wire example. Actor publication before the public pause update returns still
does not promise a particular race between frames on one socket.

Idle watch retention, invalid partial projection after overflow, explanatory
oversize close, and prompt-acceptance uncertainty were implementation or fixture
findings under the existing contract. Their initial failures and corrections
remain in the student's review. No requirement is relaxed to accommodate them.
The isolated browser speech recording is a local feasibility result, separately
labeled from the required real-model browser demonstrations and from any claim
that Bill heard it.

The editorial corrections preserve the new personal motivation and all restored
speech requirements. They assign saved preferences and dividers to Chapter 8,
qualify reconnect to its 100-event window and omit unsupported Eloquence and
diagnosis-duration specifics. These answer the outstanding prose review, not
an invented complaint from the student. Actual-use feedback, initial experience
freeze and post-comparison revisions still require their later dispositions.


## Initial live experience disposition, October 8

The author read the frozen `8cc87f2` experience and receipts, including all three
combined terminal transcripts, browser commands, selected actual screenshots,
file bytes and audio audits. These are the student coder's sessions, without
Bill participation. The author has prior Chapter 4 implementation exposure
and no Chapter 7 implementation authorship. Initial difficulties remain in the
student's own record; this entry appends their disposition.

| Student observation or suggestion | Disposition |
|---|---|
| “Without tools” from an earlier prompt may have affected the later Chat Completions file request. | §7.8 scopes a demonstration restriction to “this answer only,” retains the first tool-free narration and explicit corrective prompt, and labels the causal explanation uncertain. The actual file/calls establish completion. |
| The hint reached all three requests, but only two models printed the requested marker. | §7.8 retains the provider difference and explains delivery versus obedience. No hidden retry or stronger compliance claim replaces the initial result. |
| Model prose sometimes equates a recorded result with a finished process. | §7.8 puts the observed running report, later poll and deliberate kill beside the explanation. §7.6 retains reports and terminal job facts as separate records. |
| Inherited cleanup uses `job_killed`, while the renderable list and explanation named only `job_ended`. | Real teaching gap. Root accepted the correction; §7.3 now includes both kinds, and §7.6 distinguishes ordinary completion from deliberate/shutdown kill without cancel-on-interrupt. Published before the affected snapshot filter repair. Coder confirmed the wording resolves the omission and reported that live cards already support both but reconnect needs the narrow correction and regression. |
| Only the Messages run captured nonempty partials; other reconnects captured active metadata before fragments. | §7.8 records the difference. Held-response local checks remain the proof of the exact handoff and partial/final transitions. |
| Asynchronous audio startup did not yield the requested two seconds of recorded silence; an initial generateContent card selection spoke the user prompt. | §7.8 preserves both limits and the second explicit answer selection. The first PCM second and later speech windows are measured; no transcription, human listening or Bill-use claim is made. |
| Screenshots show a viewport, while rendered history continues offscreen. | Actual figures have alt text describing the visible tool state, connection and pause counts; adjacent text receipts retain the complete rendered history. No full-history visual claim. |

The earlier malformed-command clarification remains resolved by the student's
confirmation. Initial projection, ownership and oversize issues remain classified
as implementation or fixture findings under the existing contract. The inherited
structural heuristic's 32/33 result and corrected retained check are described
in the refreshed TL;DR and outline without rewriting their failed receipts.
The original 15-check scope remains in its dated evidence entry; current wire
coverage is 23 checks and the full gate has 33 groups.

Author dispositions are published for the student's confirmation. Historical
comparison findings, affected revision receipts and final independent proofreading
remain pending at this entry; this is not chapter acceptance.


## Browser ownership findings after initial experience

Independent review and coordinator analysis found that same-root replacement
could retain old DOM handlers, and that Page-local calls to a document-shared
native speech cancel could stop a peer Page. These are later review findings,
not complaints retroactively added to the student's initial account.

The author published the coordinator's explicit decisions before affected code:
§7.6 requires listener removal and pending reconnect cancellation on idempotent
close; §7.1/§7.7 require an actual application/document owner for one FIFO native
speech service and child Page ownership of local state. Cancel removes only
that Page's requests and invokes native cancel only for its active utterance;
waiting for native output still holds its speaking cause. Idle Page lifecycle
cannot cancel another Page's speech. Both layers fence stale callbacks.
Section 7.9 requires replacement and two-Page distinguishing controls. Public
API names remain student choices, and no mutable global or sibling dependency
bag is introduced. The student was directed to review the new ownership plan
with root before implementation. Repair and affected live results remain pending.


The student read the complete initial-live disposition table and confirmed that
all eight items accurately resolve its teaching review, including uncertain
prompt causation and preserved failures. It also recorded the affected
BrowserApplication/SpeechService lifetime plan for root review before the
shared-native-speech and disposal repairs. This confirmation resolves the
teaching exchange; it does not stand in for repair validation.


## Application shutdown refinement from revised review

The independent reviewer found a teardown-order issue during the shared-native
speech repair: disposing Page A could cancel its active utterance and start
queued Page B while the whole application was closing. Section 7.7 now explicitly
stops native admission before Page disposal. This clarifies root ownership under
the accepted shared-service contract. The coder repaired the order; independent
targeted checks at `9ba7855` are reported by the reviewer, while revised live use
and final acceptance remain separate. The chapter now prints the correction
supplement command alongside the full gate. No initial receipt is relabeled.
