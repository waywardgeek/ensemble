# Chapter 5 comparative code review

Reviewer: `review_restart_ch01`, October 9, 2026. **Accepted with documented
exceptions at initial review; zero correction rounds used.** Actual original
grading is **110/120 FAIL**, solely `logger-accessible`. The inherited media
exception also remains visible in nested Chapter 4's **90/100** result.

Read the entire mandatory `.agents/skills/ensemble-coding/SKILL.md`, AGENTS,
carryover, original Chapter 5 and [completed student review](ch05.md) before
deciding. Chapter SHA-256 is
`452a3edc46346bcd48a63483fc7c2ffbbb4056f29446dfc7fb6c292d3a0e4520`.
Verified the submitted 32-file fingerprint
`5d64b57085a9358bd309193591664b5c36d08cbf2ef467cd9c4e288754f9520a`.
Compared the student with frozen `solutions/edition-2/ch04`, and original
`solutions/ch05` with its own Chapter 4 predecessor, following moved files.

## Design, scope and teaching

The chapter makes the existing architecture usable by an external application
without rebuilding it. Registration extends the Agent's ordered registry; both
declarations and dispatch use that same storage. Duplicate names cannot silently
replace builtins, rejected registration leaves visibility unchanged, and the
registry copies the caller's schema buffer. Registration is explicitly synchronous
setup before Ask, without a new concurrency mechanism. Custom tools use the
existing ordinary-job path, including stored output and call/result identity.

Ownership is concrete: call context retains its actual Agent, Engine retains
that Agent, and Agent stores its required Ensemble interface. Constructors reject
missing required parents and attachment checks both sides. Logging follows
`call.Engine().Agent().Ensemble().Logf`; Engine and Job diagnostics also walk
existing parents. One standard logger belongs to Ensemble, synchronizes writes
and timestamps messages. Narrowing its public interface avoids exposing mutable
logger configuration. No logger callback bag, global registry, extra forwarding
layer or future actor feature was added.

Import inspection confirms the common hub has no implementation imports and
spokes import no siblings. CLI, GUI and external consumer use the public root.
This improves on the original's separate CLI wiring, Engine service bundle and
behavior-heavy common package; it preserves the restart's already established
architecture. The original also copied registration declarations into Engine
configuration and bypassed ordinary job supervision for custom handlers. The
student needs neither duplication nor a custom dispatch path.

`ConfigFromEnv` moves the existing parser into the public library; CLI tool/data
choices stay at the application boundary. Explicit discovered model selection
preserves accepted behavior rather than adding historical model defaults. The
separate `ch05` module registers its own `shout`, calls real Ask, saves the event
log and shuts down. Its bounded one-shot prompt interface satisfies this chapter's
external-consumer exercise. The extra example lines make input validation,
cleanup, usage and inspectable evidence clear; they are not core framework growth.

New comments explain setup ordering, ownership of schema bytes, logging access,
application configuration choices and evidence limitations. Core comments are
14.2% of nonblank production versus the original's 23.1%, below Bill's approximate
20% target. Reasons remain beside the affected mechanisms; no padding or compressed
code was used to improve the totals. The new code is proportionate to this exercise.

## Comparative counts

Each physical Go line is classified as blank, leading-`//` comment-only, or active
(including imports, braces and literal schemas); no block comments complicate
this method. Tests and the external example are separate. Original package moves
are counted once, not as newly implemented features.

| Go category | Active | Comments | Physical | Chapter delta: active / comments / physical |
| --- | ---: | ---: | ---: | ---: |
| Original production | 3,324 | 1,000 | 4,760 | +228 / −83 / +195 |
| Student core production | 2,396 | 397 | 2,926 | +23 / +9 / +36 |
| Student GUI production | 8 | 3 | 13 | 0 / 0 / 0 |
| Student external example | 74 | 12 | 91 | +74 / +12 / +91 |
| Original tests | 348 | 103 | 487 | +5 / 0 / +5 |
| Student core tests | 1,051 | 36 | 1,122 | +51 / +1 / +55 |
| Student GUI tests | 29 | 2 | 34 | 0 / 0 / 0 |
| Student external example tests | 70 | 2 | 75 | +70 / +2 / +75 |

The public parser is 32 active lines, largely moved from the CLI; it is not 32
lines of new behavior. Combined student production including GUI/example totals
2,478 active / 412 comments / 3,030 physical; combined tests are 1,150 / 40 / 1,231.
The frozen original tree contains no separate exercise module to count. Its
compiled predecessor binary and generated logs are excluded. Module files/sums
are support metadata: original Chapter 5 has seven physical lines; student core
and GUI retain seven and eleven, with eleven new example-module lines and a
27-line example README. No generated implementation or other language was added.

The main improvement is avoiding duplicate composition/configuration and retaining
real ownership, not the smaller delta. The student's additional tests exercise
the newly public capability rather than merely adapting existing package imports.

## Verification and exceptions

Independent formatting checks were empty; vet and uncached race tests pass in
all three modules. Built the consumer separately. Reviewed all 15 recorded
mutations and their intended failures; the aborted reference-count setup made
no mutation and is correctly excluded. Independently reproduced disconnected
root logging and no-op registration: the external HTTP-fake test fails for the
missing diagnostic and missing declaration/result respectively, without race
or build failures. The fake exercises actual Engine, Agent, History, job and
handler code; it does not mock their behavior. Prior unchanged coverage remains
applicable. Scratch copies and the reviewer build were removed.

Audited all three retained live logs and the transcript. Each contains the
requested lowercase argument, one correlated successful custom call/result,
done job with 15 stored bytes, and the final response. Usage sums match
932/73, 363/25 and 222/65; captured root diagnostics include timestamps. Actual
disk-byte inspection remains the student's recorded observation, not a reviewer
paid rerun. The first Anthropic capture loss and unnecessary repeat are disclosed;
the retained repeat supports the claim. No paid calls were repeated for review.

Independently reran the unchanged original grader: **110/120 FAIL**, matching the
student. Coordinator's frozen original reference baseline is **120/120 PASS**.
Accept the student's logger explanation after inspecting the checker:
`storedBackPointerUse` requires a direct `x.field.Logf(...)` selector, and
`dispatchStructFor` additionally requires a struct in common carrying that same
logging interface. These particular shapes do not recognize the student's
multi-hop accessors and private dispatch implementation behind a common interface.
The actual required logging capability is demonstrated by source, the external
test, the disconnected-logger mutation and live diagnostics. Waive only these
structural predicates; do not require redundant forwarding or a false parent to
appease them. Bill's architecture is fully binding and is not waived.

Two passing grader checks also need honest interpretation. The tool-call check
accepts an unknown `calculate` error because it looks for a tool-result marker;
it does not prove custom execution. The external consumer supplies that positive
proof. The parity check awards 40/40 at an 80% threshold, so its passing label
does not erase nested Chapter 4's **90/100** inherited media failure. Chapter 5
adds no attachment capability: retain that earlier scoped exception and reassess
when an actual chapter introduces attachments. Keep both exceptions visible in
later original results; neither excuses a new runtime failure. Leave grader and
prose corrections for the later author phase. No current correction is requested.
