# Chapter 7 student teaching feedback

Initial implementation is in progress. The student retains its actual read
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
