# Trial 02 history review — friction beyond the interviews

Bill requested this review after the architecture follow-up interview. No student
was reopened, rerun or modified. Review used tool arguments/results, visible
assistant prose and lifecycle metadata; opaque reasoning and API payloads were
excluded.

## Evidence boundary

The retained child journal contains 620 events, sequences 480–1099. Earlier
implementation events are not present in this journal; this is a detailed review
of the later source-review/checkpoint/grading period, not a complete first-to-last
audit. It includes 120 tool calls: 51 reads, 24 edits, 16 commands, eight directory
listings, nine waits and other controls. Seven pause/resume pairs are recorded;
paused time is not attributed to model inefficiency. Event numbers below refer
to this journal. A private compact extraction remains with Cadence's trial files.

## 1. Reading instructions did not translate into bounded calls

There were 25 chapter reads in the retained period. Several calls described as
bounded omitted `end_line`, reading toward EOF until a byte cap. Later calls
advanced `end_line` but omitted `start_line`, restarting at line 1 (883,889,895,
901). This created overlapping reads while the model said it was continuing.
Only five of all 51 read calls supplied both bounds; most source files were small,
so that count alone is not a defect. The chapter-specific repetitions are.

Events987/990 attempted Chapter0 recovery with end lines220/330, no start line,
and the same7,000-byte source cap. Saved results for jobs368/369 were exactly
identical, 7,320 bytes including line prefixes and truncation marker. Changing the
end bound could not help because the byte cap stopped both reads at the same point.

Lesson: teach a concrete read/continue example with both bounds and use the last
actually delivered line, not an intended range, as the continuation point.

## 2. Two different clipping signals are easy to confuse

Calls with an8,000-byte source cap produced a literal `[truncated]` in file text
while the enclosing job metadata said `truncated:false` (603–612). That metadata
correctly reports that the tool report itself was not clipped; it does not say the
file read reached its requested end. Other calls hit report clipping as well.
The experimental skill explained two caps, but that explanation did not prevent
repeated trouble. A clearer diagnostic naming source truncation separately from
report truncation, with an explicit continuation location, is a plausible general
tool improvement. It is not implemented by this review.

## 3. Partial-file placement confused concrete edits and reads

After listing actual subdirectories (802–814), the student tried nonexistent
root-level `Agent.cs`, `Engine.cs`, `Ensemble.cs` (832–838), then corrected them.
The errors said `Sandbox filesystem operation refused ... (errno 2)`, although
these were missing paths inside its workspace, not an attempt to escape it.
It then reread the correct implementation files in consecutive batches without
intervening edits (847–868).

A failed comment edit at716 searched `Agent/Agent.cs` for persistent fields that
were actually declared in Common. This is concrete evidence of confusing partial
type file roles, not proof that it could not compile partial constructors.
Lesson: a small file-role example and actionable missing-path diagnostics may
help more than restating an architectural slogan.

## 4. Checkpoint recovery was mixed, not simply successful

Event480 contains a useful model-authored milestone note with paths, build result
and remaining work. Event785 is an automatic checkpoint with no model-authored
note, immediately followed by compression launch786. After it the student
re-enumerated directories, reread setup and source, tried wrong paths and restarted
chapter reading. Temporal order supports a recovery-friction observation, not a
claim that compression alone caused every repetition.

The first observed creation of `PROGRESS.md` is event999, just before the grading
request, despite the primary asking for a short running progress note. The earlier
directory listing also had no such file. Thus the suggested persistent recovery
aid was not in place for these checkpoints. The interview's broad claim that
checkpoint recovery preserved the next action missed the cost of re-establishing
state. A milestone note should be maintained before a cut, not written as a final
closeout artifact. No change can retroactively recover omitted detail.

## 5. Review found a symptom but did not complete the contract

At622 the student correctly noticed direct parent assignments instead of parent
registration. Edits623/626 removed those assignments. The Engine constructor still
did not call its parent's add method; the caller's separate attachment remained.
That is an observed partial repair, stronger evidence than a retrospective guess
about why the original implementation was wrong.

Events704–770 contain19 edit attempts whose non-comment text was unchanged; two
failed anchors. Teaching comments were required, so this is not inherently wasted
work. However, the lengthy comment pass left the structural violations intact and
included comments describing intended rather than complete actual guarantees.
Operation-level checks should precede declaring those guarantees in prose.

## 6. Parent-message protocol and short waits added friction

The child requested grading at1005 and received question job372. It waited eight
times on that handle, usually with the default short observation delay (one10-second
wait), then submitted at1063 while the grade was still pending. The question job
was not the actual grader job376. Its eventual reply was the parent's acknowledgement
that grading would start and the result would follow in a later message.

The host did not wake an idle parent model turn on a child question; Bill's message
brought the issue to attention. That is supervision/runtime friction, not a student
compiler failure. The acknowledgement-then-later-result protocol also required the
student to distinguish reply completion from grade completion. Its early submission
violated the primary's wait-for-actual-grade instruction, but it honestly reported
that the grade was pending. Teach the two stages explicitly, use longer observations
when appropriate, and address the already-queued general message-wake behavior.
Do not credit repeated short polling as productive implementation work.

## Implication for the next experiment

Do not explain this trial only as failure to understand abstract architecture.
There were also concrete file-reading, source-location, progress-record and
supervision problems. Keep the general coding standards. A next authorized trial
could fix the configuration up front and test one small operational teaching change,
rather than stack all proposed improvements and attribute the result to one skill.
No code, skill, grader, runtime or candidate change was made by this history review.
