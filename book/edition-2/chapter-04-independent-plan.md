# Chapter 4 independent review preparation

Prepared 2026-10-07 before inspecting the new Chapter 4 implementation.
This reviewer authored the preceding Chapter 3 human-client integration and
its evidence verifier. The reviewer is independent of the fresh Chapter 4
builder; this is not a claim of complete source blindness. The accepted
Chapter 3 public declarations were consulted at `edition-2-ch03-r1`.

The full current coding skill, architecture, new Chapter 4 contract, voice,
writing procedure, and global forward-lesson map were read. No new Chapter 4
implementation or first-edition standard was inspected during preparation.
Historical comparison follows the initial student implementation/run freeze.

The existing independent CLI checker contains 35 cases and four pure checker
controls. The controls passed before extension. No case has yet run against
the new student's Chapter 4 binary. Tests derive from §§4.2–4.9, not private
implementation names; fault injection may need an explicitly scoped adapter
after the implementation freezes.

## Expectations fixed before source inspection

- Retain all existing declaration, result pairing, cumulative usage, lifecycle,
  replay, terminal setup, cursor, process-group cleanup and one-shot tests.
- Park a provider response; release a gated process; observe its durable
  `job_ended` while HTTP remains parked. Then return an omitted-ID Gemini call
  and a supplied-ID call. The synthesized ID uses the actual later response
  sequence and part index; the supplied ID and parser response bytes survive.
- Create Agents in one application with shared and distinct workspaces.
  Handles must be globally distinct, artifacts cannot collide, and a foreign
  handle cannot be waited on, written to, or killed. Pending limits stay with
  their Agent. Exercise public requests and Agent-attributed ordered observers.
  Choose adapter spelling from the accepted public API after freeze.
- Use barriers for pre-call persistence failure (zero side effects), terminal
  persistence failure (fault, waiter wakeup, cleanup), and failed report commit
  (no duplicated or silently consumed report). Inspect actual output and file
  effects. Do not invent rollback or demand an unpublished injection method.
- Check completed, killed and repeat reports against recorded state. Stress
  kill/exit races and deterministically test done-before-kill and kill-before-
  release. Inspect terminal events and artifacts rather than scanning all child
  output for words such as done or exit_code.
- Test cursor progression after a capped interval and later appended bytes;
  preserve complete UTF-8 heads/tails and exact omission counts. Gate later
  output externally so absence/presence checks have known boundaries.
- Drive actual Delve through human CLI chat using a local scripted provider,
  alongside the student's separate real-model PTY evidence. Compile the target
  before timing; require breakpoint output and the inspected known value from
  the debugger artifact. A missing Delve prerequisite fails explicitly.
- After the freeze, compare the new code with the first-edition standard at
  Chapter 4 scope. Inspect all ownership/import paths and helper logger access,
  data/race boundaries, useful invariants and unnecessary complexity. Give
  rationale to the student without copying old code.

The author owns teaching changes; the student owns `solutions/edition-2/main`;
the coordinator owns frozen exports and tags. This reviewer owns independent
checks and reviewer notes. No shared Go grader changes are planned; if needed,
first capture its legacy baseline and preserve all legacy regression coverage.

Preparation update, October 7: the CLI/replay checker now contains 50 cases;
the separate public-consumer runner contains three provider cases. Six pure
checker controls pass. No Chapter 4 student run is claimed. The external Go
consumer was formatted, vetted, tested and built against an isolated export
of the accepted Chapter 3 library solely to validate its public adapter.
Its runtime job checks await the Chapter 4 freeze.

The author's allocation clarification is included: preserve occupied regular
files, directories, symlinks and targets; consume candidates monotonically;
reuse a workspace across two fresh CLI processes without granting old-job
supervision rights. A separate non-directory artifact parent must refuse
execution rather than skipping indefinitely or modifying the parent.
Current public cases check shared-workspace handles, foreign wait/input/kill,
per-Agent pending limits, omitted tool selection, ordered public completion
observations while HTTP is parked, and isolated/idempotent shutdown. Durable
storage faults, injected races and first-edition comparison remain pending.
