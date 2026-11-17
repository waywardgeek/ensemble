# Trial03: behavioral pass, architectural corrections still required

Fresh student `flint-ch01-student-03`, exact free Flint Chat offering, confined
workspace, only approved Go/.NET read-only imports, working limit200000. The six
larger memory budgets were saved before admission. Bill rebuilt/restarted the host;
live read diagnostics and five local wake tests against the installed DLL passed.
The prepared primary was installed with approval and its previous copy preserved.
This combines setup, diagnostic, example and policy changes; it is not a
single-variable comparison with trial02.

## Observed outcome

The student initially repeated line-one reads despite saying it would use
`start_line`. Received tool arguments omitted the field; the outgoing tool schema
included it. That establishes neither a missing schema field nor the exact cause
inside generation/provider processing. Cadence recommended pausing for diagnosis,
but Bill asked for fresh evidence. The student had recovered independently: it
compiled a partial-constructor experiment, then wrote the actual implementation.
No proposed pause or corrective hint was sent. Lesson: a loop at one checkpoint is
not evidence of a permanent stall.

The student froze its candidate and requested grading. The parent kept the question
open until the actual result. Attempt1 used the unchanged offline grader through
the trusted confined launcher:100/100, passed:true, process exit0, including both
62-second probes. All13 frozen source/review/build hashes remained unchanged. The
grader noted message growth larger than expected but allowed it. No parent source
repair or second grade occurred.

## Independent parent source review

Corrections remain required despite the score:

- Common/Values.cs contains executable constructor bodies.
- Engine construction checks null but relies on Agent's subsequent AddEngine call
  for attachment, violating constructor-complete relationships.
- Engine.Request and Application.RunAsync both append the same exchange; this
  explains the grader's warning about duplicate conversation growth.
- Application error returns bypass cleanup; normal completion destroys the Agent
  but not the root. Successful cleanup of one path is not complete lifetime handling.
- The HTTP client keeps its default timeout; the62-second probes do not establish
  the chapter's unrestricted generation lifetime.

There are improvements: direct Engine destruction now detaches and disposes, and
cross-parent additions reject before mutation. The self-review still asserted
constructor and Common guarantees contradicted by source. An operation checklist
is not evidence that its postconditions were actually traced.

A bounded follow-up interview supplied these diagnoses but failed with a provider
error before submission. It provides no evidence of recognition or understanding.
No retry or configuration change was made. Hashes remained unchanged; the student
was shut down and joined, with the original spawn receipt collected. Chapter1 is
not architecturally accepted and Chapter2 was not started. Bill next authorized a
fresh independent Flint reviewer, excluding this diagnosis record from its inputs.
