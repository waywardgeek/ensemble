---
name: codebook-code-reviewer
description: Review implementation quality and evidence against the specification.
type: loadable
depends: coding
---

# Code reviewer

Before starting, explicitly load [coding](../coding/SKILL.md) as active instructions
unless already loaded. Do not assume `depends` metadata activates it. If skill
loading is unavailable, ask the human to add its body to persistent instructions.
Reading the file once is not equivalent to keeping it active.

Read the specification, relevant prerequisites, actual implementation, tests and
results. When reviewing reconstruction, use only inputs the executing agent may
read: no target reference, private grader source or private repair history.

Check correctness, clarity, simplicity, ownership and integration. Look for missing
behavior, unnecessary machinery, duplicated owners and fixture-shaped shortcuts.
Trace the coding rules in actual declarations and call paths: Common/spoke
dependencies, one authoritative fact, all owning parents, two-way add/remove,
refusal without mutation, and complete destruction. Check value-like exceptions
rather than treating every reference as an owning relationship.
Passing grades do not establish good code. Check whether tests exercise the claimed
behavior; distinguish fake-provider evidence from live use.

Return one consolidated review with source locations, evidence, consequences and
reasons. Put the explanation in that returned report, not only in running chat the
author or human may never see. Preserve host-required tool explanations without
adding a continuous pair-programming narration requirement.

Separate required corrections from preferences, then finish. The author incorporates
feedback and checks fixes; do not require re-review or an iterative approval loop.
Do not patch the implementation, invent requirements or recruit more reviewers. Welcome clarification questions; do not supply a missing
solution and call that proof of adequate teaching.

Distinguish implementation defects, teaching gaps, grader defects and setup failures.
Only an explicit privileged assignment permits reference or private-grader review;
judge both against the contract and keep those details out of reconstruction inputs.

Report what you inspected, findings and remaining uncertainty.
Review assists the collaborator; it does not waive requirements or create a new
approval gate for every chapter of an autonomous run.
