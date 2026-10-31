# Chapter 03 independent review

## Verdict: approve

No material manuscript findings or blocking revisions. No optional nits warrant an author pass.

Read the complete approved set: companion voice guidance; all 947 lines of the Edition 3 source chapter, including exercises and live-run account; the accepted Chapter 02 companion and its author note; and the current Chapter 03 companion and author note. Recovered the clipped Chapter 02 output before completing the review. This was an editorial review, not a new execution of the implementation or reproduction of the recorded runs.

## Substantive review correction

The first version of this note incorrectly described explanation-policy refusal as transactional rejection with unchanged logs, conversation, usage, and private material, and no tool error results. That was a review error, not a manuscript defect. On recheck, both the current manuscript and source explicitly preserve the assistant response, return a matching error result for every refused call, and permit an explained retry within the existing round budget. The effect-free requirement concerns refusing the entire batch before any handler runs; it does not prohibit recording the response and refusal results. Several purported quotations in the original review were also inaccurate. All quotations retained below have been checked against the saved companion chapter; other descriptions are unquoted paraphrases. This recheck finds no actual manuscript problem.

## Findings

The chapter advances Chapter 02 rather than repeating it: the event-history and rendering seam now carries actions, correlated results, and recovery. The progression from corpus findings through the loop, explanation policy, tool affordances, edit ambiguity, authority boundaries, and evidence is readable as a chapter for a technical lead. The stories do explanatory work without becoming invented reader experiences. It avoids exercise instructions, code dumps, contract tables, and a summary-first substitute for the argument.

The source’s essential exercise knowledge survives causally. This includes per-Agent registration and empty-registry omission; vendor declarations and the narrow Gemini schema projection; mixed text/call replies, multiple calls, missing-ID normalization, and results-first ordering; recoverable argument and execution failures; the sixteen-round limit’s inability to stop a hanging command; read/search/write details; overlapping edit sites and actionable refusals; shell-key filtering; and the inherited fake’s terminal zero-usage reply. Fixture and grader mechanics are translated into reasons an observation can or cannot establish the promised behavior.

The especially consequential boundaries are stated correctly:

- **Historical corpus and causation:** “It does not, by itself, isolate the tool’s causal effect from every other feature of the workload.” The tool counts and ratios are supported by the source and framed as observations of Bill’s corpus, not universal measurements or causal proof. The original ambiguous-anchor defect is not represented as a verified deployed repair.
- **Source-bound signatures:** “Its signature stays bound to that exact text part and to its source vendor, API surface, and model, and is replayed only to that same source.” Visible ordinary text remains portable while genuine hidden thinking stays hidden. The later missing launch identity is explicitly a setup problem, not evidence of parser loss.
- **Explanation refusal before handler effects:** “It preserves the response and returns an error for every call, explicitly saying that nothing ran and requesting an explained retry.” The draft requires refusal of the whole unexplained batch before any handler, not partial execution. It preserves same-response text-before-call ordering, an explained retry within the existing round budget, and the selected-model scope of the policy. It does not require transactional rollback of the response or conversation.
- **Command and tool errors:** “That is ordinary command data, not a malformed tool invocation.” The silent `exit 7` versus `go run` wrapper distinction makes the evidentiary trap clear rather than merely naming exit-code reporting.
- **Authority:** The draft explicitly distinguishes a small Agent-local registry, launch directory, and named-variable filtering from filesystem confinement, shell restrictions, and comprehensive credential protection. It does not present them as a sandbox.
- **Live-run chronology:** “Those fixes had deterministic coverage; the earlier live run is not evidence that the later binary exercised them.” The repair, recall, actual missing-file errors, and additional tool coverage remain source-reported observations; overlapping-anchor and empty-registry corrections retain their separate deterministic evidence.

The ending follows from the actual limitation of synchronous execution and gives Chapter 04 a concrete problem to solve. No manuscript revision or author-fix recheck is required on this draft.
