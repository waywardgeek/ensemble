# Chapter 8 contract review

Coordinator review of the complete draft at `cfb0d87`, October 7, 2026.
The ownership, persistence and execution-policy choices are accepted as working
design decisions. The coordinator reconciled the two wire/lifetime clarifications
below before any Chapter 8 student handoff. Chapter 7 acceptance and publication of the initial
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

## Clarifications found during review

1. Section 8.4 says malformed command shapes receive correlated errors,
   while Chapter 7's original “malformed transport” wording could be read more
   broadly. The printed Chapter 7 unknown-field error also supports the
   nonfatal interpretation. Specify transport, usable correlation and domain
   validation separately so a grader cannot silently choose a conflicting
   interpretation of the two chapters. The student finding and settled
   behavior below resolve this ambiguity in both contracts.
2. Preference snapshots/changes have their own revision and arrive before the
   Agent snapshot generation is known. State that they remain bound to the
   current socket/Connector connection lifetime. Late callbacks from a replaced
   connection cannot overwrite preferences on the new connection, and the new
   initial snapshot resets that connection's applied preference state. This
   preserves Chapter 7's stale-generation rule without inventing an Agent watch
   revision for Server-owned preferences. A controlled old-socket callback must
   distinguish this behavior.

Chapter 7's fresh student subsequently found the same transport ambiguity in
its printed unknown-field error example. The coordinator clarified §7.4 at
`1bba1bc`: correlatable semantic command errors leave a usable connection,
while malformed transport and unusable IDs close. Chapter 8 now explicitly
inherits that settled boundary. Section 8.4 also binds preference callbacks to
the current Connector lifetime, with an old-socket distinguishing control in
§8.9. Root made these narrow prose changes while the author thread was idle for
Chapter 7 grader capacity; final independent proofreading remains required.

These are precision fixes to the published wire/lifetime boundary, not requests
for a new owner or additional settings feature. Method names, CSS breakpoint,
storage representation and reusable component implementation remain student
choices. The full contract is accepted for checker preparation; its eventual
student still requires accepted Chapter 7 and a published acceptance command.

Final story/voice proofreading follows actual implementation and receipts.
The draft already keeps the omitted-zero failure and dead execution-limit
lesson attached to concrete user consequences, and explicitly labels its spin
as pending. No invented successful browser or speech transcript is present.
