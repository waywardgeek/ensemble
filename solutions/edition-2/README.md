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
