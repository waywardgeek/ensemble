# Idle-transition goal continuation

## Contract

Bill requested immediate continuation when an enabled Agent returns to idle,
replacing the normal idle timer. Failed submitted model turns wait 1, 10, 100,
then 1,000 seconds repeatedly. Success resets the failure streak. Direct tool
controls do not reset it; a normally delivered tool error is model-handled.

The existing actor owns a pending wake, generation guard, failure count and
failure-only timer. Timer callbacks only enqueue mail. Normal idle continuation
queues that mail directly. Admission still checks goal authority, fresh readable
instructions, pause, foreground work and current generation. Independent Jobs
remain owned. A retry continues from recorded state, not a replay of tool effects.
Broken History publication remains an actor fault rather than a retry loop.

Pause invalidates pending work; resume reevaluates eligibility and starts the
current failure delay again when applicable. Interrupt, Off, submission, unload
and shutdown prevent automatic continuation. Explicit enable resets backoff.
Retry now bypasses one wait without enabling, unpausing or resetting the streak.
Clear also clears transient failure status. Registered child restart retains its
own configuration and pause, but starts with a fresh failure streak.

Legacy idle_seconds is still validated and retained for old callers/restart
records, but no longer controls scheduling. The GUI removes that editable field,
shows failure count and next retry time, and provides Retry now. The tool schema,
source skill, original Chapter24 and its historical Goals companion are updated.
No frozen endpoint or numbered historical grader is changed.

## Verification

Evidence directory: `/private/tmp/cadence-idle-goals.9LJj2t`.
All providers are owned loopback fakes. Builds use the existing explicit SDK,
cleared package sources and existing local packages; no installation or download.

- Initial causal test1852 passes against the real actor and HTTP boundary.
- Existing goal tests1856 initially report4 pass/3 fail against old timer
  assumptions. Submission's shared listener could see the now-immediate parent
  request. Revised fixtures pause explicitly before configuration/publication,
  and stale-timer coverage uses an actual failed turn and unload during backoff.
- Revised lifecycle groups1859 pass7/0. No assertion accepts both old and new
  normal scheduling behavior.
- Shared core1866 passes352/0. This precedes moving the unchanged delay helper
  from Common into Agent/Goals and clearing transient failure state on Clear.
- Final1880 clean core/GUI builds, causal backoff1/0, lifecycle7/0, restart2/0,
  real two-tab GUI socket test pass. The final causal test exercises five failures:
  it waits for the real first one-second retry, verifies10/100/1000/1000 deadlines,
  checks queries preserve deadlines, uses retry-now for longer waits, and verifies
  success resets backoff. It does not wait out a full1,000-second interval.
- Actual-page1877 passes no timer field, failure/deadline display, typing pause,
  disabled retry while paused, resume/retry-now, Off and reload. The fixture
  reports exactly3 requests and joined shutdown. This page build precedes only
  the helper relocation/Clear-state change, covered by final focused tests.
- Static header1881 passes12 layouts with updated labels. It extracts renderers
  for layout assertions and is not lifecycle evidence. The actual-page test uses
  normal DOM controls and real sockets, not injected application frames.
- Active Edition3 Go grader sources contain no Goals timer assertions. Affected
  executable checks are the C# tests and browser scripts, not a new public score.
- Chapter24 and skill hard prose checks pass with advisories. The companion's
  two em dashes cause lint failure and match HEAD exactly. Initial chat wording
  incorrectly attributed lint failures to both texts; exact output corrected it.

No installed-application, live-provider, native accessibility or universal
failure-recovery claim follows. Goal instructions and reminders remain explicit
opt-in authority; this does not grant deployment, tools or external access.
Source and test fixtures are reviewed; the final source receipt is labelled
post-execution, not represented as a pre-run snapshot.
