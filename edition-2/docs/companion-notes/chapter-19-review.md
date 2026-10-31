# Chapter 19 independent review

## Result

No material findings in the stable-release manuscript. The H1 exactly matches the source: `# Chapter 19: The Sandbox`. The chapter is a concise human explanation of what the boundary does and why it needs these mechanisms, rather than an implementation assignment or acceptance checklist.

## Review basis

Read the complete current `book/companion/voice.md`, `book/edition-3/chapter-19.md` including all exercises and ending, accepted `book/companion/chapter-18.md`, and `docs/companion-notes/chapter-18-author.md`. Recovered the clipped Chapter 18 read with a bounded overlapping range. After the editor's explicit stable-release confirmation, fully read `book/companion/chapter-19.md` and `docs/companion-notes/chapter-19-author.md`. Judgments use the actual source chapter, not the author's summary.

## Material checks

- The opening carries Chapter 18's untrusted-content and granted-capability distinction forward. The runtime remains trusted host code, not a hostile process supposedly contained by its own helpers. Command network denial and the named web-skill gate are not presented as universal Agent egress controls.
- The explanation preserves both enforcement paths: canonical containment and traversal-resistant file operations, plus inherited OS confinement for command descendants. It retains state/media/job-output access, refusal-path privacy, platform limits, real system exceptions, the nested-cache escape, and the distinction between necessary metadata and broad subtree access.
- Credential handling covers transport ownership, pre-retention redaction and reflected tokens, allowlisted command environments, safe Git configuration, and the limits of known-string scrubbing. Safe-mode declaration removal and dispatch refusal remain distinct. Child configuration enforces narrowing independently of its diagnostic without pretending child spawning is implemented here.
- Exercise-derived evidence requirements survive: existing outside fixtures, useful positive controls, a reachable network peer, individual compound-command outcomes, absent write effects, and guessed-dispatch refusal. Historical runs and later direct regressions are not conflated; the restart account does not certify universal isolation or the deferred OAuth/browser flow.
- The credential incident, removal bug, profile failures, cache correction, and full-access device response retain their source-supported causes and qualifications. No invented causal step or material source conflict found.

## Limits

Document review only. No implementation inspection, build, lint, grader, runtime verification, Git operation, student pipeline, or child agent. Only this review file was written. No correction or follow-up recheck is currently requested.
