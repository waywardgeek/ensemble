# Chapter 8 contract review

Coordinator review of the complete draft at `cfb0d87`, October 7, 2026.
The ownership, persistence and execution-policy choices are accepted as working
design decisions. Two wire/lifetime clarifications below must be reconciled
before the student handoff. Chapter 7 acceptance and publication of the initial
independent Chapter 8 checker remain prerequisites; no implementation or live
result is claimed here.

Read the complete new chapter and its outline/evidence, current architecture,
full voice/procedure and the preceding Chapter 7 contract. Historical incidents
are grounded in the author's retained source research; this pass does not claim
an independent replay of first-edition settings failures or their measurements.

The contract gives a saved setting four observable boundaries: validation,
persistence, application and display. Dedicated patch presence and complete
snapshots preserve explicit false and zero. Separate Server and Agent services
keep GUI preferences out of Engine; actor-ordered execution policy has a real
consumer at turn activation. Default 16, the final accepted tool batch and
historical absent policy fields retain their prior meanings. Distinct owner
revisions and separate files avoid claiming a cross-owner transaction.

Checked replacement before acknowledgement, unchanged published state on an
earlier write failure, serialized candidates and owned workers give persistence
failures a teachable meaning. The close path resolves a started writer's actual
result instead of pretending a committed file was rolled back. The shared
autoplay rule is explicit about delayed delivery: each page captures its locally
applied revision/rate on enqueue, owns its existing queue and sends only its own
pause transitions. Disabling future speech and canceling present speech remain
separate user actions.

## Clarifications before handoff

1. Section 8.4 says all malformed command shapes receive correlated errors,
   while Chapter 7 closes malformed transport/command input. Preserve that
   earlier boundary explicitly. Invalid JSON, invalid UTF-8, binary/oversized
   messages and the earlier strict envelope failures keep their Chapter 7
   behavior. A valid settings command whose domain patch has a rejected value,
   null/duplicate/unknown member, stale revision or persistence failure gets
   its attributable settings error and leaves the connection usable. Specify
   envelope errors separately so a grader cannot silently choose a conflicting
   interpretation of the two chapters.
2. Preference snapshots/changes have their own revision and arrive before the
   Agent snapshot generation is known. State that they remain bound to the
   current socket/Connector connection lifetime. Late callbacks from a replaced
   connection cannot overwrite preferences on the new connection, and the new
   initial snapshot resets that connection's applied preference state. This
   preserves Chapter 7's stale-generation rule without inventing an Agent watch
   revision for Server-owned preferences. A controlled old-socket callback must
   distinguish this behavior.

These are precision fixes to the published wire/lifetime boundary, not requests
for a new owner or additional settings feature. Method names, CSS breakpoint,
storage representation and reusable component implementation remain student
choices. After author reconciliation, update this review with exact revision
and disposition before releasing a fresh student.

Final story/voice proofreading follows actual implementation and receipts.
The draft already keeps the omitted-zero failure and dead execution-limit
lesson attached to concrete user consequences, and explicitly labels its spin
as pending. No invented successful browser or speech transcript is present.
