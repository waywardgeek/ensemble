# Edition 2 student implementation

`main/` is the authoritative working source, built fresh by the student from
the original book and accepted carryover. This outer repository tracks it;
there is no nested Git repository.

`chNN/` is an exact source export at an accepted chapter checkpoint. Do not edit
exports. Carry fixes forward in `main/`, revalidate, and create a fresh revision
tag rather than moving an existing tag. The first edition and retired attempt
remain historical evidence, not source for the student to copy.

Chapter reviews live in `../../docs/edition-2-notes/`; small sanitized run evidence
lives in `evidence/chNN/`. Graders remain the original course graders during the
student run. For Chapter 1, run from the repository root:

```
make grade-dir CH=1 DIR=solutions/edition-2/ch01
```

Chapter 1 was accepted on initial review, with no exception. Its checkpoint is
the immutable annotated tag `edition-2-ch01-r1`. The sorted relative-path and
SHA-256 source manifest has digest
`c2dd9901cc615844e717901de89e3b091046b62b85303393b3bf68b27416c41d`.
Source and export matched byte for byte before checkpointing. See the student
and comparative reviews for actual grading, live use and mutation results.

Chapter 2 was accepted with a scoped exception after one correction round.
Its immutable annotated checkpoint is `edition-2-ch02-r3`; the existing older
`edition-2-ch02-r2` tag is preserved. Source/export manifest SHA-256:
`3a042fc7f3308471777f4a74178df2e246df5e7b87ea5fa31f7cc4a21b93ef90`.
All 17 exported source files matched byte for byte before checkpointing.

The original Chapter 2 grader remains **92/100 FAIL**: its added URI-media
rendering and unredacted media control exceed the chapter's stated scope.
The reviewer verified the explicitly required Ref/redaction behavior separately.
Keep these actual failures visible through later parity checks. Chapter 4's
independent scope reassessment confirms that output references are not URI
attachments: §4.5 explicitly excludes rendering them as parts. The exception
therefore continues through that chapter. Do not add early features
merely to conceal the inherited score. See `ch02.md` and `ch02-code-review.md`
in the review directory for the decision, original results and revision history.

Chapter 3 was accepted on initial review, with the inherited media exception
and no correction rounds. Checkpoint: `edition-2-ch03-r3`; older r1/r2 tags are
preserved. Its 19 source/export files match; manifest SHA-256:
`b92d8f8c7fdd79e2b39226ed88067f0f9fdd2a8fcae3c702e6138bf49e6ca459`.
Original grade is **90/100 FAIL**: all nine new checks pass, while `ch2parity`
reports the same two waived media controls. The inherited exception was
subsequently reassessed against Chapter 4 as described above. Student/reviewer notes
and live three-provider tool exercises are recorded with the checkpoint.

Chapter 4 was accepted with the inherited media exception after two test-only
correction rounds. Checkpoint: `edition-2-ch04-r2`; the older r1 tag remains.
All 25 source/export files match; manifest SHA-256:
`36605e815176cf0ed244c4ba98ef88df97e59f51efb59cf4e8927b2e9858bf3d`.
Original grade remains **90/100 FAIL**, with all eight new checks passing.
Live three-provider supervision, actual debugger prompt matching and shutdown
are recorded alongside the unsuccessful attempts and corrected evidence claims.
Independent review verified the two lifecycle regression tests, including their
failure paths, before acceptance. See `ch04.md` and `ch04-code-review.md`.

Chapter 5 was accepted at initial review with documented grader exceptions and
zero correction rounds. Checkpoint: `edition-2-ch05-r2`; the older r1 remains.
All 32 source/export files match; manifest SHA-256:
`5d64b57085a9358bd309193591664b5c36d08cbf2ef467cd9c4e288754f9520a`.
Original grade is **110/120 FAIL**: its logger-shape check rejects the demonstrated
multi-hop ownership chain. Nested Chapter 4 remains 90/100 despite the enclosing
parity check's passing label. The separately built consumer proves real custom
registration, execution and root logging on all three providers; the grader's
generic tool-return check does not establish that capability. See `ch05.md` and
`ch05-code-review.md` for the precise exceptions and comparative results.
