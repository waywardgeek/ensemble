# Trial 02 — compact primary, offline behavioral pass

Status: grade and source review complete; not architectural or full Chapter 1 acceptance.

## Configuration and assistance

Fresh `flint-ch01-student-02` used exact `openrouter/chat/flint-v1_0-free`, primary
`flint-student`, confinement to its assigned child home, and only Go/.NET read-only
toolchain imports. Public chapters and offline NuGet input matched baseline 01;
the environment script changed only workspace identity. No reference source,
private grader source, real credentials or parent source repairs were supplied.
The installed primary was byte-identical to the codebook candidate (SHA-256
`59982f6bafbd8fdb176db741151e4ab92a9016d2fa1c53c4b40e4878c3ae4a0f`).

This was not a single-variable or unassisted comparison. It changed the primary,
isolated-workspace guidance and installed existing-file diagnostic. During the
trial Bill raised the working limit from 100,000 to 200,000 tokens (advertised
capacity stayed 262,144); the parent sent Bill-requested milestone handoff advice.
Bill subsequently changed all enabled byte budgets to Conversation 100,000,
Session 40,000, archive8/archive64/MEMORY 10,000 each and SOUL 5,000. Conversation
initially remained 3,000 and was corrected after saved-policy verification.
These policy changes cannot reconstruct previously summarized detail.

## Behavioral evidence

The unchanged external Chapter 1 grader ran once through the previously verified
trusted confined launcher. The candidate remained frozen; SHA-256 checks of all
11 captured source/build-output files matched before and after grading.
Grade 1 returned **100/100**, `passed:true`, exit 0, including both 62-second
response-header/body delays. Parent job 376 is the completed grading receipt.
The JSON result is retained alongside this note. No second grade or live-provider
exercise was performed. This proves the grader's offline behavior, not full
architecture, every malformed-input case or live-provider acceptance.

## Independent source findings

The final frozen files show improvements over baseline: shared partial method
and constructor declarations, transactional conversation append after success,
checked usage sums, and an application `finally` cleanup path. The request uses
an absolute `/v1/messages` path rather than baseline's doubled `/v1/messages`.
These observations do not establish a causal benefit from the primary.

Architectural acceptance still fails:

- `Common/Owners.cs:7–9,20–22,40–55,81–88` contains method/accessor bodies,
  contrary to the supplied declaration/implementation split.
- `Engine/Engine.cs:14–23` validates its parent but never attaches to it.
  `Agent/Agent.cs:22–24` separately calls `AddEngine` after construction.
  Direct Engine construction therefore does not establish ownership.
- `Engine/Engine.cs:123–128` disposes its transport without asking its parent to
  detach. `Agent.Destroy` compensates by removing it first (lines108–112), so
  direct Engine destruction leaves an owned reference to a destroyed object.
- `Ensemble/Ensemble.cs:16–30` rejects membership in this Ensemble only, not an
  existing different parent. Adding an already-owned Agent to another Ensemble
  overwrites its parent while the first Ensemble retains its membership.

These are source findings, not newly executed candidate tests. The student's
`SOURCE_REVIEW.md` cited these locations as satisfying the requirements but did
not detect the contract violations. A source-location checklist is insufficient:
review each operation's postconditions when invoked directly, including another
parent and direct construction/destruction, not only the CLI's normal path.

Other source observations, not grader failures: `Application/Program.cs` does not
dispose each parsed JsonDocument; the unsuccessful-HTTP branch starts a body read
without awaiting it before disposing the response. These should not be confused
with the specific architectural acceptance findings above.

## Process and interview limits

The student corrected compiler errors itself. It used the explicit SDK path after
one bare `dotnet` command failed. Its rebuilds eventually succeeded and local EOF
and missing-credential checks exited 0 and 2 respectively. Repeated chapter reads,
clipping and final-review announcements persisted; terminal logger escape traffic
also enlarged reports. Some parent reports were clipped, so this is not a complete
chronological audit of every call.

A model-authored handoff was observed followed by resumed source review. The
student's interview claimed successful automatic-checkpoint recovery; that report
is not independent proof of every checkpoint transition. It named SDK-path and
self-review repetition as friction and suggested protocol/error-handling changes.
Those suggestions do not demonstrate that the proposed features would address
its main process or ownership failures. Its final disclosure omitted the completed
memory-budget intervention; the parent record above is authoritative for that fact.

The student submitted once while grading was pending. Original spawn job 48
therefore contains a pre-grade closeout. A separate post-grade message supplied
the real result and requested the single bounded interview; the collected typed
generation-3 submission acknowledges 100/100 and offline-only scope. No process
nudge was sent: the student independently requested grading before the parent
acted on the proposed nudge. No Chapter 2 work or candidate repair was requested.

The question exposed a separate host scheduling limitation: child notices wake
pending observers but do not start an idle parent model turn. Bill queued the
broader requirement that incoming messages wake idle Agents. That is a queued
runtime improvement, not an implemented change or a student defect.

Graceful shutdown job 389 completed with `stopped:true`; original spawn job 48
was collected. Candidate hashes still matched after closeout. No owned trial work
was intentionally left running.


## Bill-authorized follow-up interview

After the trial was closed, Bill explicitly requested a second interview about
architectural confusion. The same child was reopened on its retained free route,
primary, sandbox and compiler imports; all were reverified after admission.
This was interview-only, not a candidate rerun or repair. Original interview
job 407 returned a typed submission, and graceful shutdown completed. All 11
captured source/output hashes remained unchanged afterward.

The parent supplied the four concrete mismatches listed above and asked the
student to distinguish actual guarantees from required guarantees, evidence from
retrospective guesses, and useful teaching changes. The student agreed with all
four findings. Its useful review diagnosis was that it "checked requested source
locations but did not fully trace constructor/destroy postconditions." It said
"exact reasoning remains partly unrecoverable" and treated narrow interpretation
and compressed review as possible contributors, not established causes.

It recommended an "operation-level postcondition checklist," with a worked
example/counterexample as possible support. It correctly described registration
at constructor return and detachment on direct destruction after this feedback.
These answers demonstrate recognition after supplied diagnoses, not independent
error discovery or proof that abstract wording caused the original mistakes.
A future teaching experiment may test that recommendation; this interview alone
does not establish its effectiveness. No skill or candidate changes followed.
