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
