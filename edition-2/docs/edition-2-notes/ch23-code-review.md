# Original23 / Edition2ch21: independent review

Reviewer `reviewer_astra`, GPT-6 Astra/high, October11,2026.
**Accepted with two narrow grader-fixture exceptions at initial formal review.
Zero formal correction rounds.** Historical preparation and failures follow;
the final disposition below supersedes their pending-review status.
Mandatory `.agents/skills/ensemble-coding/SKILL.md` read in full; governing
carryover, workflow, AGENTS and approved chapter map apply.

Reread all808 lines of original `book/chapter-23.md`, including examples,
revision `e509c2ddc237f799844a6efe6ebdb744289ce295`, SHA-256
`3b19960c74cf4d01e225b9c60bc175b199cc95ec01cad754b97d9a3dcf4bd061`.
Read [student-safe notes](ch23.md). Accepted predecessor is original21+22,
main tree `a801628b1dd631e5ea20ad03ea7d255584728d31`, exported at `27acbed`
and immutable `edition-2-ch22-r1`. Review is for sandbox capabilities only;
original24 reuse and original25 fleet implementation remain later assignments.
No original23 answer was inspected before student implementation.

The previously supplied launch/public-input erratum remains byte-current:
`internal/grade/ch23_harness.go` SHA-256
`24b7e13a90868cdf7b1f2fd1a3f4bf710f9b5ebffc84594bcfd5bb2c9d3cf80c`.
The original grader plants a library probe in its supplied tree, so coordinator
will grade an isolated exact copy, preserving authoritative source. Its library
package-name and terminal-envelope assumptions require actual-result assessment;
no exception is preaccepted. Missing interface facts were supplied narrowly,
without grader internals or answer code to the student.

Earlier availability preflight established executable `/usr/bin/sandbox-exec`
on Darwin and exit0 from an allow-default `/usr/bin/true` probe. This proves
availability only, not confinement. No provider or original-grader call is made
by this reviewer. Credential tests should use fake markers; existing profile
secrets remain outside reviewer access.

Final review will assess both in-process file operations and real spawned
commands/descendants, with legitimate reads/writes/toolchain work alongside
escape/network denials. It will trace credential environment/log boundaries,
actual registry removal and web-skill withholding, unconditional child narrowing
separately from its diagnostic, host-path-safe failures, actual ownership and
unsupported-platform refusal. Those are the original chapter's claims, not an
expanded security certification or a demand for future spawning features.
Source, own-predecessor/reference comparisons, student explanations, unchanged
grading, live use and targeted mutation evidence remain pending frozen handoff.


## First-green source preflight: actual cache-boundary defect

Read frozen `fcb98a9459b328314f0a7ff06efbb5e7c85fe07e`, main tree
`f2f8ff00229a4e198988b8756c8702b42b34ab1c`, the student's implementation
checkpoint, and matching first-edition sandbox source after implementation was
available. Final review/live/mutations remain pending. Zero formal rounds used.

The original grader's actual result is **0/100 FAIL**,0.623s, without timeout.
Six behavioral sessions stop at `provider model is not set`: its launch supplies
no model. The child-spec probe imports the module root, which is a command,
instead of the public `ensemble/` package. These are independently confirmed
fixture incompatibilities, not proof that sandbox behavior passed. No exception
is decided before the complete student review and supplemental evidence.

The student `TestSandbox*` suite passes independently with `-race` from an exact
frozen archive (exit0,1.252s), including real OS commands and host-network control.
An additional focused probe nevertheless found a real current-scope defect:
Seatbelt's xcrun cache rule combines the Darwin temp subpath with a basename-only
regex. It admits matching names at arbitrary nested depth, not only the actual
cache location. A fake-marker file named `xcrun_db-private-marker` in an unrelated
sibling `t.TempDir()` was readable through the real Agent, job and `run_command`;
`cat` returned `REVIEWER_OUTSIDE_CACHE_MARKER` and exit0. No private file, profile,
credential or provider was used. The chapter explicitly warns against shared
temporary-directory access; this is a boundary defect rather than a new policy.

Reproduction and output are retained as reviewer-owned
[probe](../../solutions/evidence/ch23/reviewer-cache-boundary-test.go.txt)
and [failure](../../solutions/evidence/ch23/reviewer-cache-boundary.txt).
Coordinator received the finding and relayed the required behavior: narrow the
allowance to actual toolchain cache paths, reject unrelated nested matching
files, and retain Python/Git allowed-work controls. No original answer code or
implementation recipe was supplied. Student main was not edited. This is
attributed pre-submission assistance, not a completed formal correction round.

A second real-runtime probe found a diagnostic leak at the same frozen source.
`run_command` resolves a relative `cwd` to an absolute host path before the
sandbox sees the request. With caller `cwd: "escape"` and a local symlink of that
name pointing outside, refusal correctly prevents execution but exposes the
canonical sandbox root in both error and tool result. The chapter requires the
caller's own spelling. Coordinator received this narrow finding plus the need
for a legitimate-cwd positive control. Preserved
[probe](../../solutions/evidence/ch23/reviewer-cwd-refusal-test.go.txt)
and [failure](../../solutions/evidence/ch23/reviewer-cwd-refusal.txt)
use fake paths only. The coordinator reports that the student independently
identified the same cwd leak at this point; retain both discoveries rather than
crediting only reviewer assistance. Still pre-submission, zero formal rounds.


## Final independent disposition

**Accepted with exceptions**, frozen source
`64e5d0c17be8db116fa4c1da52cd290c45b4e2ef`, main tree
`9dc3e9e36a854ede4cc2cbef22dc389850cedc0c`. Read the complete final
[student review](ch23.md), new implementation and final diff, matching original
`solutions/ch23` implementation and tests, and each edition's own predecessor.
The mandatory `.agents/skills/ensemble-coding/SKILL.md` was reread after
compaction. Evidence-only checkpoint `320859a` leaves the reviewed tree unchanged.
No unresolved implementation, ownership, documentation or scope finding remains.
This accepts original23 / Edition2ch21; original24 remains a separate assignment.

The actual unchanged grader remains **0/100 FAIL**, on first-green `fcb98a9`.
Independently accept the student's two explanations, limited to these fixtures:

1. Six sessions supply a local fake endpoint and dummy credentials but no model;
   they exit before exercising sandbox behavior. Requiring an explicit model is
   existing application behavior. An invented default would not demonstrate
   confinement. The fixture also retains the previously documented terminal
   envelope assumption, which these failed launches never reached.
2. Its planted child-spec probe imports the module's root command instead of the
   public `ensemble/ensemble` library, and assumes the package identifier `agent`.
   This is a probe-layout assumption, not an architectural defect in the actual
   composition root. Real public Child/Clamp tests exercise the required behavior.

These exceptions waive neither confinement nor credential separation, tool
removal, child narrowing or positive controls. No grader was edited or relabeled
passing. The source/library layout and explicit model contract continue into
later exercises; future grader maintenance should use the actual public boundary.
The student's omitted convenience constructors and unspecified `send_secret`
implementation are also reasonable: explicit root selection and required-owner
construction provide the behavior, while the latter is a source data-path
pattern rather than an assigned new tool contract. No future spawning is added.

### Corrections and verification

Both genuine pre-submission defects above are fixed, not waived. The final cache
regex anchors immediate physical-temp cache files; it no longer admits unrelated
nested lookalikes. The tool preserves the caller's cwd spelling into sandbox
resolution. Independent rerun of the exact two original probes on an archived
final source passes: nested-marker `cat` receives `Operation not permitted`, and
cwd refusal is exactly `sandbox refused "escape"`. The same focused race run
passes all eight student sandbox tests, including real allowed Python/Git,
relative cwd, file operations, descendant and network restrictions, credential
separation, safe mode, web gating, state-file symlinks and media. See
[output](../../solutions/evidence/ch23/reviewer-final-regressions.txt)
and [provenance](../../solutions/evidence/ch23/reviewer-final-regressions.json).
The complete core/optional-GUI test and vet results, final focused race check and
empty gofmt result are retained in the student's evidence. The final API-log
setter uses the same lock as writers; this small synchronization change does not
create a second ownership route.

Independent [live audit](../../solutions/evidence/ch23/reviewer-live-check.json)
verified the preserved binary hash and all 12 confined plus 3 safe-mode actual
request/completion model-effort pairs as GPT-6 Astra/high. SSE completed item
records match all 11+3 subsequent tool results. File editing/search/listing,
local command and Git succeed; symlink/traversal and descendant outside access
are refused; the fake environment value is absent. Safe-mode requests omit all
five execution tools while actual file work succeeds. The fake markers are absent
from retained API, history, save and terminal artifacts. Both CLI sessions exit0,
without timeout, in48.394s and9.847s. The launcher records a reachable host HTTP
positive control and keeps that peer alive while the hit count stays1; curl's
exit7 alone is not the network proof. Deterministic removal of the network wall
also reaches the peer. Actual endpoint interception and universal immunity are
not claimed.

Those live runs remain attributed to **first-green `fcb98a9`**, before the cache
and cwd fixes. Final real-process regression evidence covers those narrow changes;
a redundant paid replay was unnecessary. Reviewer read no credential profiles
and made no provider call. An initial reviewer parser incorrectly expected calls
inside response.completed output; this capture carries them in output_item.done.
The corrected audit checks those records and retains that limitation explicitly.

### Active code, comments, tests and KISS

Own-predecessor deltas, ordered physical / active / comment-only / blank lines:

| Edition and category | Delta |
| --- | --- |
| Student production Go | +620 / +515 / +73 / +32 |
| Original production Go | +1245 / +550 / +601 / +94 |
| Student test Go | +408 / +376 / +20 / +12 |
| Original test Go | +1373 / +1017 / +191 / +165 |

[Counting receipt](../../solutions/evidence/ch23/reviewer-counts.json)
records paths and method: lexical physical lines, inline-comment code active;
docs/evidence/generated output excluded, tests separated. There are no source
moves or non-Go implementation deltas; original generated `cr/io` output is
excluded. The reference is its ch22→ch23 snapshot, the student its accepted
original22→final original23. These are net chapter changes, not total application
sizes or a judgment based on compact formatting.

Active production growth is similar despite much shorter physical growth. The
student needs an Agent-owned filesystem/process spoke and common contract because
several existing consumers require one boundary; it does not introduce a general
policy framework. Required parent construction and attachment preserve both sides
of ownership. Runtime policy comes from the owner. Go rooted operations protect
actual opens after canonical checks, including state, media and retained output;
the original check-then-open design is weaker under a symlink swap. Original
nil-owner fallback is not copied because Bill requires the actual parent.
The student keeps implementation spokes separate and uses the public composition
root in actual clients. Existing jobs still own subprocess lifetime. Narrow
Child/Clamp behavior stays outside common. No extra dependency or future feature
is justified by this chapter.

The original contains more explanation, including useful threat-boundary and
ownership discussion, but some claims are too strong: canonicalize-then-open is
not race-safe, and nil-parent permissiveness is not Bill's required relationship.
The student's +73 comment-only production lines are about12.4% of its nonblank
net growth, below the rough20% teaching target. That arithmetic is not used as a
quota: the new exported symbols/members have useful documentation, and nearby
comments explain owner policy, missing ancestors, rooted opens, safe error
spelling, child alarm versus enforcement, Git identity, ancestor metadata,
restricted cache paths and unsupported-platform refusal. The README accurately
states defaults, flags, file/state coverage, platform limits and no spawning.
Its shorter prose is sufficient to follow the reasons without filler.

Tests exercise the actual Agent/Engine/jobs and OS sandbox, using only small
external HTTP fakes. The original tests include a substitute sandbox parent and
network-negative evidence without a local reachable-host control; the student
improves those boundaries and adds state/media integration. This does not imply
complete security coverage. Unsupported-platform refusal is source-reviewed,
not independently executed on Linux; standard-library rooted-operation guarantees
are not a bespoke race stress-test claim.

Reviewed all27 applied mutation attempts:25 intended behavioral failures, one
initial OpenFile survivor and one media failure at the wrong boundary. The
OpenFile gap now has a direct operation test; the media repair first proves an
absolute in-root positive, then catches actual outside encoding. Neither failed
attempt is erased or counted as an intended kill. Cache/cwd restored defects,
real descendant/network escapes, state redirection and credential redaction
fail at their intended observations. Git-config mutation demonstrates broken
allowed Git work, not private-content exfiltration. Coupled safe-mode/overall-wall
mutations are sensitivity evidence, not several independent security proofs.
The student's corrected explanations and evidence limits are accepted. No mocks,
blanket historical sweep or new grading framework is needed.

Both reported implementation defects were resolved before formal submission;
zero formal correction rounds were consumed. Preserve their failures and the
actual failing original grade in the final export record.
