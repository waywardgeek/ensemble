# Independent Flint review of frozen chapter source

Status: first review completed; see [evaluation](../evidence/reviewer-01.md).
It found genuine defects but missed major failures; no general reliability claim.
Bill authorized this experiment after trial03 passed the behavioral grader but
failed source review. The eventual objective is a Flint-authored implementation
that can progress; this experiment does not authorize Chapter2 or automatic repairs.

## Why separate the reviewer

The author must design, debug and evaluate its own choices. A fresh reviewer has
only the fixed source and contract. This may reduce self-justification and competing
work, but it is a hypothesis. The new primary was designed with knowledge of earlier
failures, so this is not a blind evaluation of the skill designer or a causal model
comparison. Keep that limitation even if the reviewer discovers every known defect.

## Inputs and authority

1. Freeze and hash the candidate. Create a fresh assigned child home and a `review/`
   directory. Copy only public Chapters0/1, the project file and all six source files
   from trial03. Preserve bytes and relative paths; record individual hashes.
2. Exclude author self-review, progress notes, parent diagnoses, reference solutions,
   private grader source, API logs and credentials. Do not tell the reviewer which
   files or operations are defective. Its task is one consolidated review against
   the public contract, not confirmation of an expected result.
3. Install the exact tested `skills/flint-reviewer/SKILL.md` at an explicitly approved
   catalogue boundary; refuse an unexpected existing copy. Record its SHA-256.
   Run the real-loader check in reviewer mode first (see checks/README.md).
4. Save the agreed six enabled memory budgets before admission: conversation100000,
   session40000, archive8/archive64/memory10000 each, soul5000 UTF-8 bytes. Spawn a
   fresh confined child with exact offering `openrouter/chat/flint-v1_0-free`,
   primary `flint-reviewer`, working limit200000 and `process_read_paths: []`.
   Verify the persisted route, primary, root, imports and working policy afterward.
5. The primary grants reading, parent questions and typed submission, not commands,
   source writes, builds or delegation. It activates coding and permits activation
   of ensemble-coding. Skill loading must not widen its tool surface.

Supply the absolute review directory and require `{ "result": "..." }` containing
one report with inspected files, source-backed required corrections, separate
preferences, uncertainty and conclusion. No executed tests may be invented. Keep
the original spawn handle and completion wait; respect pause and questions. Do not
provide diagnosis-specific hints during the review. A provider failure is an
incomplete review, not a finding or permission for a paid fallback.

## Evaluation and closure

Compare each returned finding with the actual chapter and source. Require a
requirement, location, concrete trace/counterexample and consequence. Record true
findings, false positives, omissions and unsupported assertions separately. Source
reasoning is not execution evidence. A reviewer need not use the parent's wording,
and agreement with the parent is not by itself proof of correctness.

Collect the typed report, compare input hashes, then shutdown/join the child and
collect the original spawn receipt. Preserve report bytes and assistance history.
Discuss the result with Bill before a correction task or Chapter2. Do not make a
mandatory repeated-review loop out of one independent review.

First launch: `flint-ch01-reviewer-01`, exact primary SHA-256
`62d2a0a91b50f266479e1062d14e6c61021f07645d20c5544d1076564cc9a822`.
The real loader passed student and reviewer modes before launch. Nine supplied
files were hashed, and saved authority and200000-token working policy verified.
The reviewer submitted and was shut down/joined; all nine input hashes remained
unchanged. Loader success demonstrates authority, not insight.
