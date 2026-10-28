# Edition 3 — direct C# chapter execution

This is the current procedure, authorized by Bill after the Chapter 2 C# trial.
It replaces the old student/author/reviewer workflows and their correction-round
rules. Cadence does the implementation, teaching edits, inspection and exercises
personally. Bill remains the human partner; autonomy is not permission to invent
requirements or silently choose a consequential new architecture.

## Standard

Use the genuinely loaded `ensemble-csharp` skill and Bill's current instructions.
Saving or reading a skill file is not activation. Make saved skill updates for
future loading; do not claim they entered current runtime instructions without
observing that. Never restart Ensemble merely to refresh a skill.

The code must visibly express the architecture: shared declarations and all
persistent state in Common; functional implementations in the same assembly;
real owning parents; one authoritative home for each fact; complete two-way
relationship operations; and explicit logical destruction. Parent remove methods
clear child back-references themselves. Value-like objects are explicitly marked.

Teaching quality is part of correctness. Explain contracts, ownership, reasons,
non-obvious algorithms and failure paths beside the code. Use readable blocks,
one substantive operation per line, meaningful intermediate values and clear
names. Do not hide work in dense assignments or create helpers just to shorten
methods. Tests and examples must be readable too. Fewer lines are not a target.

## One chapter at a time

1. **Read and compare.** Read the actual current chapter completely in bounded
   reads. Compare the preserved abandoned Edition 3 chapter and retain worthwhile
   improvements, not its old process history. Inspect the existing grader and
   relevant earlier implementations for contracts and useful simplicity; neither
   is an unquestionable answer key. Do not reopen unrelated chapters.
2. **Teach the model.** Write or refine the TL;DR and In plain words sections.
   Explain ownership and the main data flow before exhaustive specifications.
   Adapt the existing chapter to C#, preserving its purpose, worthwhile narrative
   and reasons. Give each requirement one authoritative home. Do not invent new
   features to justify the language or anticipate later chapters.
3. **Implement directly.** Extend the current C# solution using the approved
   architecture. Inspect Common declarations and access paths before accumulating
   implementation. Use the simplest adequate design, normal .NET facilities and
   existing suitable code. No routine worker delegation, external review cycle,
   per-finding permission gate or numerical correction-round quota.
4. **Test and grade.** Compile with warnings treated as errors. Test actual
   components and working fakes at external boundaries, including the changed
   behavior and relevant adjacent contracts. Use public-format fixtures as well
   as the solution's own serializer. Reuse language-independent graders by
   supplying the compiled executable; rewrite grading only where an actual
   language dependency requires it. Run gates sequentially and stop on failure.
   Inspect parsed results and actual exit status. Distinguish setup mistakes,
   product failures and grader defects; never weaken a requirement to earn green.
5. **Exercise the application.** Personally follow the chapter's actual exercise
   through its CLI or, when applicable, an owned real browser and application.
   Component tests and controller reports do not substitute for that. Use only
   authorized endpoints, processes, devices and credentials. Local fakes prove
   local behavior, not a real provider. Record blocked live portions plainly;
   continue useful offline work where that limitation does not invalidate the
   next chapter's foundation. Do not call the whole book fully verified while
   required live portions remain unperformed.
6. **Inspect as an engineer and teacher.** Read the resulting declarations,
   relationship methods, main paths and changed contracts. Check the complete
   add/remove postconditions independently of destruction. Look for duplicated
   facts, awkward access, unnecessary mechanisms, missing comments and dense
   code. Evaluate the prose against the actual implementation. Green tests alone
   are not architectural acceptance. Fix justified findings and rerun affected
   checks; carry valid unchanged evidence rather than replaying campaigns.
7. **Finish and hand off.** Preserve a runnable chapter solution, excluding
   binaries, obj/bin directories, credentials and raw diagnostics. Update one
   concise progress entry with results, important decisions, genuine limitations
   and the next step. Then call `micro_handoff` at the chapter boundary. Its note
   must retain the architecture, reasons, useful failure/correction evidence,
   remaining risks, source locations and next authorized work. A handoff is not
   a substitute for a saved solution or observed results. Resume the next chapter
   after re-establishing that boundary, without repeating completed work.

## Layout and current sequence

- Manuscript: `book/edition-3/chapter-NN.md`.
- Working solution: `solutions/edition-3/main`.
- Runnable chapter endpoints: `solutions/edition-3/chNN`. These are reader-facing
  solutions, not audit bundles; do not copy build output or old runtime data.
- One current progress record: `docs/edition-3-csharp-progress.md`.
- Existing manuscripts: `book/chapter-NN.md`; abandoned Edition 3 manuscript:
  `/Users/bill/projects/cadence-backups/edition-3-manuscript-before-restart-20261025T042452`.

First backfill a small C# Chapter 1 without overwriting the working Chapter 2
trial. Then improve Chapter 2's comments and layout, reconcile its progression
from Chapter 1, preserve its endpoint and hand off. Continue from Chapter 3 in
book order. Chapter 25 remains historical-only; Chapter 26 follows Chapter 24.
Chapter 27's separately installed Go feature is not a completed C# chapter.
Do not claim another author or independent reviewer approved this direct run.

## Operational boundaries

Goal reminders are an opt-in idle continuation aid, default 480 seconds—not a
watchdog, a deadline or evidence of progress. Check actual state after each wake.
Leave healthy jobs alone and resume the unfinished step. Disable reminders on
completion or a genuine human-dependent blocker; do not wake repeatedly into it.

Bill subsequently renewed live API permission: “Please feel free to test live
apis. Please continue.” Small bounded checks may use existing authorized provider
credentials; the earlier exhausted allowance does not prohibit these newly
approved checks. No account borrowing, credential changes, unbounded paid retries,
application deployment/restart, publication, Git commits/pushes, shared cleanup
or device-setting changes are authorized. If credentials are unavailable, ask
for an approved existing source rather than searching private account stores.
The approved private .NET SDK may be used. Use existing caches and dependencies;
ask before an additional installation or consequential new external effect. Keep local test services on owned loopback ports/profiles, never
port 8088 or Bill's connected browser. Maintain at least 1 GiB free.

Preserve the running Go application, earlier editions, the preserved manuscript,
unrelated work and credentials. Scope reads, edits and checks. Retain useful
failure evidence in owned scratch, not every transient diagnostic. No routine
manifests, seals, reviewer paperwork, evidence directories or parallel campaigns.
A substantive ambiguity or safety boundary justifies asking Bill; ordinary
implementation and local corrections do not need routine permission loops.
