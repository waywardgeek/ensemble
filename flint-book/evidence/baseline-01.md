# Baseline 01: general primary, isolated Chapter 1

## Configuration and boundary

The child used `ensemble` as primary, activated `code-tools` and
`ensemble-coding`, and selected exactly `openrouter/chat/flint-v1_0-free`
(wire model `recursive/flint-v1.0:free`). Working-context limit: 100,000 tokens.
It was confined to its assigned child home, with read-only Go and .NET SDK imports.
The parent supplied Chapters 0 and 1, an absolute project path, an environment
script and offline NuGet configuration. No reference implementation was supplied.

Harness source milestone: `8f5cb81b`; grading was parent-mediated using the trusted
host launcher, not a student-callable tool. The first trial preceded the new file
error diagnostic (`1e3ae9ef`) and isolated-exercise guidance (`827bbcf1`). Those
changes cannot be credited with its recovery. This record is a historical baseline,
not an exactly replayable claim about a provider's future behavior.

## Results

The unchanged offline Chapter 1 grader returned exit 0, `passed:true`, 100/100,
including the two 62-second delayed-header/body checks. No parent source repair
was supplied. Setup/recovery interventions and grading were parent-assisted.
The separate live exercise was not performed. The child was shut down after its
post-grade interview; no further test was performed on this passed baseline.

Personal source review found Common algorithm bodies instead of partial-member
declarations, shared state outside Common, separate construction and registration,
and incomplete relationship/resource teardown. It also found `/v1/messages`
appended twice. The ordinary fake accepts a path ending in `/messages`, so this
escaped the grade. Preserve the grade's actual result without treating it as
complete protocol or architecture certification. Historical graders remain intact.

Counting reported usage before rejecting an empty reply is *correct*: Chapter 1
requires counting consumed tokens even without usable text. An initial review
criticism of that ordering was withdrawn after checking the chapter.

## Observed friction and proposed responses

The journal contains 102 tool calls: 32 reads, 13 directory listings, 3 searches,
25 commands, 14 writes, 7 edits, 3 waits, 2 skill loads, 1 parent question and
2 submissions. Counts describe this trial, not an efficiency benchmark.

- Successful skill activation was followed by nine outside-root refusals while
  searching for skill/procedure files. The general skill unconditionally referred
  to repository documents unavailable in the isolated exercise. Separate checkout
  guidance from student instructions; loading is not source-file discovery.
- Reads confused source-byte limits with report-byte limits. Increasing only
  `max_bytes` left one report clipped. An out-of-range line request returned empty
  success. Improve recovery guidance and EOF diagnostics; do not infer full reading
  from a successful tool status.
- An automatic context checkpoint occurred early. The student disputed its origin
  and repeated setup. Parent clarification directed resumption from saved files.
  Teach resumption explicitly; do not adopt the student's recommendation to
  distrust real host checkpoints.
- Four replacement writes to Owners.cs omitted `overwrite:true`; another replacement
  of Program.cs did the same. Generic errno 17 looked like authority refusal.
  The student recovered with shell heredocs. The general diagnostic fix belongs
  in the main book. Missing paths also produced generic errno 2, still a tool
  improvement candidate, not a reason to weaken confinement.
- Two failed builds piped through `tail` returned shell exit 0. The student read
  and repaired actual compiler errors. Preserve exit status in the command recipe.
- Twice the student executed a publish directory rather than its apphost. Negative
  tests under `set -e` skipped later checks. Use an explicit executable path and
  separate expected-failure calls.
- A request to an external API hostname with a fake key failed DNS under confinement.
  This was outside the offline assignment, not a successful live-provider test.
- The student submitted a pre-grade closeout after a parent acknowledgement. The
  actual grade and interview came later. Clarify that acknowledgement is not the
  grade and that submission closes an assignment.

The interview was not an authoritative diagnosis: it incorrectly suggested that
`edit_file` requires overwrite and disputed the verified checkpoint. Use interviews
as hypotheses, then inspect actual actions and contracts.

## Lessons retained

Keep security, chapter requirements and honest evidence fixed. Change the small
skill and task surface first. Self-review must point to actual source relationships,
not merely repeat architecture vocabulary. Preserve failed experiments and the
parent's interventions; do not describe this baseline as autonomous acceptance.
