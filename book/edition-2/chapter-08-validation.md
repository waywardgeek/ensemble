# Chapter 8 validation

Preferences and execution policy preparation, October 8, 2026. The required
predecessor is accepted checkpoint `edition-2-ch07-r2` at
`5b82971c3e667e08acfb2eca132307b9cdebbc51`. Fresh student `/root/coder_ch08`
started with `fork_turns="none"`; its ownership plan is accepted and
implementation is underway. Initial local passes are student-reported;
independent acceptance and real-provider demonstrations remain pending.
Bill's editorial approval is separate.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Contract | Author, coordinator | Complete draft `cfb0d87`; reviewed clarification `7200f17`; `chapter-08-review.md` | Preserve corrected Chapter 7 browser ownership in the handoff |
| Independent checks | Coordinator, grader engineer | Initial checker `b3f3137`, strict fixtures `f6ccbf5`, additional checks/coverage record `10b17c8`; student reports local wire, validation and policy-effect passes | Finish public/concurrency coverage and independently verify a frozen student source |
| Student and ownership plan | `/root/coder_ch08`, coordinator | Accepted in `main/evidence/ch08/student-review.md`: actual parents, actor-ordered policy changes, owned persistence workers and shutdown joins | Check implementation against those owners and lifetimes |
| Initial implementation and live use | `/root/coder_ch08` | Core policy and GUI preference backend implemented; student reports inherited tests and initial local checks pass; browser controls underway; no paid calls | Finish local gates, freeze source/support identities, review preflight, then run all three providers |
| Historical comparison and revisions | Independent reviewer | Not started | Preserve initial source, live receipts and teaching review first |
| Manuscript and feedback | Author, student, proofreader | Draft explicitly labels actual spin pending | Reconcile actual receipts and resolve student findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before immutable chapter export/tag |

Initial invocation: `python3 scripts/edition2/accept_ch08.py GUI_BINARY`.
Strict validation: `python3 scripts/edition2/accept_ch08_validation.py GUI_BINARY`.
Its scope and remaining independent controls are recorded in
`chapter-08-grader-review.md`. Wire/persistence success alone does not establish
browser usability, audible speech, concurrency or actual policy enforcement.

The student receives committed-only new Chapters 1–8 under
`/Users/bill/projects/ensemble-edition-2-revisions/ch08-student-inputs/`, with
hashes and predecessor identity in its manifest. It reads the mandatory coding
skill and architecture directly. Its corrected preceding main source is
`9ec94ef551b08de6fbd714bf95efee901db49400`, tree
`944b187319224bc2d17b833ae1c72bfd93643c2e`, identical to the Chapter 7 r2 export.
The original teaching manifest records the r1 handoff; `predecessor-r2.json`
records the correction without relabeling that history. Initial package
discovery exposed an evidence helper importing the optional GUI module. The
student held implementation until the isolated helper repair and independent
full-tree checks established r2. The original failure and r1 tag remain intact.
Old chapters/answers, future teaching, author research and grader implementation
remain outside the student context. This is an instruction boundary in the
shared workspace, not a filesystem sandbox.

Independent grader `/root/grader_ch05` develops remaining contract checks in
parallel without consulting the historical Chapter 9 implementation before the
student's initial freeze. Temporary author `/root/coder_ch04` is preparing the
Chapter 10 persistence outline and remains available for teaching questions.
Root owns orchestration, plan review, validation ledgers and checkpointing.

The proposed live matrix covers human CLI, actual browser/audio and public
two-Agent use. Coordinator-reviewed limits are at most ten human prompts and
24 model HTTP requests per provider, plus three read-only discovery requests,
without automatic retries. The matrix must explicitly demonstrate a policy
change during an active turn retaining that turn's capture, followed by a new
turn using the new policy. These limits do not authorize launch before local
gates and complete source/binary/support/dependency preflight are satisfied.
