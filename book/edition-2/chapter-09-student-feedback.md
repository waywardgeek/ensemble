# Chapter 9 student teaching feedback

## Initial plan questions, October 8, 2026

The fresh student's initial review is
`solutions/edition-2/main/evidence/ch09/student-review.md`, based on teaching
`8736c951393320cf8beef849806bde3b90bc445f` and accepted Chapter 8 source
`446d7f27fcef052d6bb780901c786842dc5286c1`. The student reported these questions
before implementation, not after a failing grader. The original questions and
source-read account remain unchanged. The coordinator reviewed the owner/API plan
and supplied the working directions below; these are not new Bill rulings.

The responding author previously implemented Chapter 4 and authored later
teaching. The author has not implemented Chapter 9. Source inspection for this
response read the complete initial student review, affected Chapter 5/9/10
contracts, the Chapter 6 opener, and accepted predecessor CLI configuration and
renderer portions to confirm the reported seams. No runtime or grader was edited.

| Question | Teaching disposition |
|---|---|
| Q1, numbered question 1: system override environment name | §9.8 explicitly adds LLM_SYSTEM through the shared reader with LookupEnv and no provider fallback. Skill mode rejects a supplied nonempty value before defaults; no-skills uses existing Config normalization. Public configuration semantics remain intact. |
| Q2, numbered question 2: primary recorded once versus request captures | §9.7 keeps the existing captured system field as the empty string in skill mode and resolves the primary from its immutable recorded activation for rendering/reconstruction. No-skills captures are unchanged. §9.8 distinguishes Config's effective primary view and accepted round trip from captured request configuration. |
| Q3, numbered question 3: H/S/P placement after a held batch | §9.7 defines skill-mode-only stable post-batch anchors: pending unanchored hints join when a batch becomes unresolved, and later hints join while it remains unresolved. Existing anchors remain fixed. It merges anchored hints/manuals by durable sequence and gives literal suffixes for all three renderers. After the turn ends and prompt P arrives, the suffix is results→H→S→P; consumption removes H while leaving S before P. Non-anchored hints retain request-tail placement. No global sorting or moving manuals is allowed. |
| Minor closing-reference finding | Verified Chapter 6 teaches streaming. Chapter 5's closing promise now leads into visible proposed work and the complete-response execution boundary; its old skills-next promise is removed. |

Chapter 10's required semantic state and settled-boundary passages also now
explicitly retain completed-batch hint anchors. An unconsumed hint is not an
unfinished tool batch; it can survive checkpoint/resume without moving a manual.
No new event format or snapshot authority is introduced by that clarification.

Scoped prose lint passes every hard rule; JSON fixture parsing and scoped
git diff --check pass. The author read the changed passages and cut repeated
wording. Remaining soft warnings concern existing chapter density/person gaps
and the short feedback record; no runtime validation is inferred. Student
confirmation is still pending publication.
This response does not claim an implementation, paid run, grader acceptance or
resolved student confirmation before the student reads it.

## Student confirmation after publication

The author read the appended `Implementation phase: release-8f24360` section in
the student's working-tree review on October 8, 2026. The coordinator preserved
the original plan/questions at 5ac45e4; this later confirmation was not yet a
separate frozen checkpoint when read. The student explicitly confirms that all
three published responses resolve the questions and that the three-provider
fixtures clarify the distinction between request capture, public configuration
and chronological dialogue. The student reports proceeding with local integration,
with the pre-provider review gate intact. This confirmation is teaching feedback,
not an acceptance claim for the implementation or later demonstrations.
