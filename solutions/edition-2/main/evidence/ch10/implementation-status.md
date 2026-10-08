# Chapter 10 implementation status

2026-10-08, milestone 1: Q1–Q3 acknowledged; complete nested semantic grammar
published in ../../../../persistence-format.md (main/persistence-format.md) and
public signatures in main/session-api.md. This documents the implementation
contract for independent checker preparation; runtime implementation and every
validation gate remain pending. Grammar updates, if needed, will be explicit.

Source predecessor ac55f641e69220a612debbcd6f75e77fa9259b36; accepted plan 7cb8429;
clarifications af5a762. Exact teaching/read ledger is in student-review.md.
No provider/credential access, tests/builds, new workers or unrelated edits.
Free disk at phase opening: 636 MiB. Preserve retained artifacts and caches.

Next: common codec/session declarations, strict byte-preserving codec and owner
validation, store lifecycle, actor controls, clients; then local/independent
checks and a bounded live matrix for separate release. The initial independent
checker remains an initial subset; additional commands await coordinator inbox.
