# Chapter 04 independent review

**Verdict: accept.** No unresolved material findings in the final manuscript. The factual corrections are verified against the approved source, and the chapter delivers technically substantial, readable companion prose rather than an exercise sheet or compressed recap.

## Corrections verified

- **Opening, companion lines 3–7:** now uses the source narrator's framebuffer screenshot hang and Bill's process-tree recovery. The unsupported Playwright/localhost details are gone. Matches source lines 10–29.
- **Lifecycle history, companion lines 41 and 165–179:** cancellation-before-start is explained as lifecycle reasoning, not an attributed observed defect. The unsupported short-timeout teardown anecdote is gone. The retained shutdown demonstration and waiter/late-finish ordering failures match the source's actual evidence (lines 926–934 and 1003–1009).
- **Debugger, companion lines 125–129:** now records the successful Gemini sequence, result 42 and exit code 0, then the mistaken escaping/timing diagnosis and unnecessary repeat. The sub-three-tenths-of-a-second summary agrees with the four measured intervals in source lines 647–659. No premature-quit history remains.
- **Coverage score, companion lines 33 and 187:** the shell-only perfect score is now the negative-control counterfactual, not an observed original implementation score. This matches source lines 884–889. The separate killed-as-done audit story at companion lines 155–159 remains correctly distinguished.
- **Consumer evidence, companion lines 201–203:** the successful consumer call losing capture through log-header parsing and premature scratch deletion is supported by source lines 1038–1046. It was not a finding in this review and should remain.

## Final contract and voice checks

- Companion lines 35–45 place ordinary work at common dispatch, allocate handles before effects, and keep the four controls out of recursive job creation. **Agent owns its jobs collection and output directory; Ensemble allocates handles** (line 39; source lines 238–245).
- Lines 59–71 accurately explain one-shot limits for the very next attempted call, including controls and failures, explicit-field precedence, replacement setters starting from defaults, and mandatory visible consumption diagnostics.
- Lines 77–101 distinguish shared unseen-output cursors, patterns, finished-job reads, initial full-result compatibility and bounded previews versus full retention. **Later no-pattern waits return promptly on unseen output**; nonmatching output does not satisfy a pattern wait (lines 83–89; source lines 322–328).
- Lines 113–129 explain the shell-handler/reader lifetime handoff, premature output closure, PTY versus pipes, echo versus reply, and measured versus assumed platform behavior. Lines 133–143 preserve explicit shell persistence, per-call cwd, missing-directory failure, optional override serialization, and workspace versus log/output storage.
- Lines 149–187 distinguish killed state, physical cancellation limits, process groups versus confinement, notification, shutdown evidence and genuine ordering tests. Lines 191–219 explain the separate consumer, correlated successful execution and actual-parent logging without treating a printed diagnostic as proof of architecture.

The causal stories and failure modes carry the essential exercise knowledge without assignments, recipes, code dumps or contract tables. The prose remains conversational and lightly playful while respecting a technically fluent reader. The ending correctly separates supervised work from interruptible human input. No arbitrary length target was applied.

## Reading scope and limitations

Fully read the approved voice guide, the complete current source Chapter 04 through EOF—including all exercise, grading and retained-evidence text—the accepted companion Chapter 03 (recovering clipped output), and its author note. After the final-save notice, fully reread the actual final companion Chapter 04, lines 1–233, and author note, lines 1–36. Checked the corrections against the source rather than adopting the editor's verdict.

The interim review incorrectly summarized job ownership and later no-pattern waits; those were review errors, corrected above, not manuscript defects. No preserved comparison establishes an intervening source change, and none is claimed.

This was a manuscript/source/voice review only. No linked historical logs, implementation files, legacy process documents, builds, graders, student pipeline or Git operations were used. Historical observations were checked against the approved source, not independently rerun. Only this review note was written. There are no unresolved material findings; independent runtime verification remains outside this review's scope.
