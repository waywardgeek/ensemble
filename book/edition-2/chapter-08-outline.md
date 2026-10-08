# Chapter 8 outline: Preferences that do something

Status: full draft manuscript/contract exists in `chapter-08.md`, awaiting
coordinator review before a student handoff.
Maps to first-edition Chapter 9. Chapter 7 must be accepted before implementation.
No Chapter 8 code, passing check, browser session or live result is claimed.

## Stake and story preservation

A reader turns speech off, changes the layout and restarts the application;
the screen must tell the truth about what was saved and what the program will
actually do. Resolve that stake with two real browser clients, a restart and
an observed behavioral change, while preserving Chapter 7's steering controls.

Retain the old chapter's invitation to start using the agent for real work,
without promising that using it automatically makes it better. Retain the
three-pane reuse test: two ArtifactScroll instances should not require two
renderers. The historical settings failure is the sharper opener: validation
corrected a value to zero, `omitempty` omitted it, and the browser continued
to show the rejected value. False speech enablement was equally unrepresentable.
Sources: old Chapter 9 TL;DR, old Chapter 13's “The bug that only appears after
you fix the bug”, and commit `3d7b7d1`. Use the incident as an earlier recorded
failure, not a freshly reproduced measurement or an invented Bill session.

Omit old “a human can watch” as a complete safety claim, unmeasured thirty-second
costs, and the assertion that Go graders cannot test browser behavior. Existing
browser tooling makes DOM and interaction checks available now. Omit promises
that the next chapter introduces sub-agents: the current map's next topic is
skills, and source chronology is not authority over dependencies.

## Draft teaching order

1. A saved setting can still be dead: distinguish accepted value, durable value,
   rendered control state and observed effect. Each needs a check.
2. Separate display preferences from Agent execution configuration. Display
   concerns remain in the optional GUI module; behavior settings retain their
   actual core owner and public control path. No settings-to-engine closures.
3. Teach a sparse patch with presence separate from value and a complete owned
   snapshot with explicit false/zero. Unknown fields, nulls and wrong types have
   an explicit policy; no partially applied malformed patch.
4. Validate both incoming changes and persisted input. State ranges/defaults
   once, decide rejection versus normalization, and send the applied values.
5. Serialize persistence and publication. Define failures before acknowledging
   success, concurrent updates, reconnect snapshot order and revision scope.
   Avoid inventing a transaction across separate owners or files.
6. Reuse ArtifactScroll for chat and actions, retaining full identities,
   partial/final replacement, safe text rendering and accessible expansion.
7. Provide keyboard-usable pane resizing, theme choice, font controls, speech
   controls and a one-Agent sidebar. Show authoritative state in semantic DOM
   attributes. Do not invent sub-agent execution for a decorative tree.
8. Shared autoplay governs future enqueue at the page's applied revision; queued
   utterances keep captured rate/revision. A local cancel reconciles only that
   page's pause. A settings patch cannot forge another client's transition.
9. Demonstrate restart and two-tab updates, actual speech behavior and a bounded
   coding task through the real browser. Retain human CLI/public composition.
10. Compare the initial student's answer only after its first implementation and
    real runs. Preserve the student teaching review and explicit revisions.

## Coordinator-accepted ownership and speech choices

The coordinator accepts GUI Server owning a persistent display-preference service in the optional
module. Its children retain owner interfaces back to Server/public client and
Ensemble for logging. The one application preference snapshot is shared by its
browser clients; input contents, pending speech and pause causes remain local
client state. Agent execution configuration stays Agent-owned. Separate message
kinds identify the preference domain instead of calling every owner “settings”.

The old exercise eventually included a saved execution limit that two loops
ignored. If this chapter exposes an editable limit, retain the new actor's
existing response-count semantics and explicit turn-start capture; do not import
old default-200/tool-batch semantics silently. Proposed core setting changes
travel through actor-ordered public requests, separately from GUI persistence.
The coordinator requires the existing default of 16 and Chapter 5 counting to
remain if the limit is exposed. The exact persisted execution subset and restart
initialization path must be specified before a student is asked to build them. No credential, log identity,
endpoint or hidden config object belongs in a browser snapshot.

Prefer a narrow, explicit schema over a generic map with arbitrary keys. The
complete draft must name defaults, valid ranges, file version, absent/null/wrong
kind behavior, persistence failure semantics, revision ordering, acknowledgement
shape and the initial snapshot/watch handoff. These are open author decisions,
not permission for a grader to assume its favorite implementation.

## Checks and evidence plan

The old grader is CH=9 and checks matching updates, arrival on subscription,
second-client broadcast, theme persistence and old GUI parity. Preserve it as
a historical diagnostic. A new checker must cover the published domains and
Chapter 7 wire protocol; no invocation has been invented yet.

Required distinguishing cases include explicit false and zero, omitted field,
wrong type, invalid persisted data, two nonoverlapping concurrent patches,
failed persistence with unchanged published state, disconnect/reconnect during
an update, stale display rejection, and safe handling of an unknown version.
Seed defaults and nondefaults independently so setting a value back to its
default still proves persistence. Observe actual theme/font/speech changes in
both tabs; a JSON echo is insufficient. Test keyboard resizing and control
attributes, malicious text in both panes and partial-to-final replacement.

If an execution setting is included, prove it changes the running policy on the
next turn and survives restart as taught, including a value beyond the previous
hard-coded limit. Preserve active-turn snapshots during concurrent changes.
Deleting the settings consumer must fail even when serialization still passes.

Live plan: all three APIs through real browser task/stream/tool/control paths,
settings/restart and reusable public embedding. Actual audible speech must be
observed where advertised; deterministic speech-error/cancel fixtures remain
separate. Keep scratch-workspace task artifacts, original terminal/browser
receipts, source/binary identities and actual failure/correction history.
Detailed demonstrations await actual student evidence.


Coordinator response: accepts this ownership split as a working choice, not a
new Bill ruling. Persisted display preferences/defaults belong to Server;
each page owns its actual input, speech queue and pause registration. The full
contract must explain multi-tab preference effects without allowing one tab's
preference edit to cancel another tab's speech or release its pause accidentally.
Agent settings stay actor-ordered and validated at load/update. No transaction
across the two owners is implied. A new user-facing tradeoff should be surfaced
before implementation rather than hidden in a store helper.


## Complete draft ready for contract review

The manuscript fixes the remaining routine choices: strict rejection rather
than silent clamp, explicit presence in patches, complete snapshots, separate
nonnegative domain revisions, stale-base conflict and one writer per domain.
Checked temporary-file replacement precedes application/acknowledgement;
filesystem work never holds the actor or socket reader. Startup validates
complete version-1 files and refuses corruption instead of resetting it.
An optional Agent policy path supports persistent or explicit memory-only use.

The editable policy is max_model_requests, stored 0 for default 16 or 1–256;
counting remains model requests and the last accepted tool batch stays paired.
Activation captures policy in the new optional turn_started.turn.policy value;
queued versus active turns, old logs and vendor-body reconstruction are explicit.
Preference snapshots have their own subscribe boundary; Agent policy changes
use the existing actor/watch boundary. The chapter promises no cross-domain
transaction. These detailed mechanics await root review, not a new Bill ruling.

The coordinator also accepted the speech semantics: current applied preference
revision/rate are captured at enqueue, existing utterances survive a shared
disable, and only the owning page cancels its queue. The draft adds a two-tab
fixture and keeps pending automatic buffers/backlog separate from queued work.

Scoped lint passes all hard checks; soft density/person-gap warnings were read.
There is no student code or live evidence. Next action: coordinator full-contract
review, then independent checker publication and handoff only after Chapter 7
acceptance. The complete demonstration remains visibly planned, not performed.
