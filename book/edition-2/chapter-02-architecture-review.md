# Chapter 2 owner and public-data review

Date: 2026-10-07. Status: targeted owner/public-data corrections reviewed and
accepted, preserved in initial checkpoint
`39a92ca27a418712832ac0dcbbfbe4e32b3bca35` and retained by the accepted
post-comparison revision `cc1bec45c3327c87728a4040f762155d8e860a0b`.
This is an early architecture review, not the required comparison with the
first-edition standard after the student's initial implementation and live run.

The reviewer freshly read the complete mandatory coding skill and inspected
the new Chapter 2 working tree: common owner interfaces, public construction
and event application, Engine accounting/exchange, event-log I/O, parser and
renderer owner paths, and the separate GUI client and its public test. No
first-edition solution was read for this phase, no code was edited, and no
paid calls or duplicate test runs were made.

The coordinator found logger-reachability drift despite an inherited 100
grade and a passing scoped package check. The reviewer independently confirmed:

- Standalone event validation and reduction helpers lacked an owner context,
  including `validPart`, `validProvenance`, and `Apply`. The same review applies
  to JSON-object validation, transition inspection, route/configuration
  mapping, usage accumulation, and error-reporting helpers.
- `eventlog.Dump` could fail while encoding but had no parent context. The
  root's configuration-normalization helper also had no owner context.
- The optional GUI client stored a public `ClientOwner` interface that had
  request/subscription methods but no route to Ensemble's logger.
- `NewAgent` copied the configuration's tool/schema buffers, but `Load` used
  the same constructor without that copy. Caller mutation could therefore
  alter a loaded Agent's stored configuration.

These are violations of existing Chapter 1/2 ownership teaching, not missing
new requirements. Corrections must restore actual parent access and owned
data; adding a global logger, separate logger parameter, or sibling-service
closure would repeat the original mistake. Pure text boxing needs no invented
diagnostic dependency.

Existing paths that inspected correctly: Agent retains `common.Ensemble`;
Engine and Log retain `common.Agent`; parser and renderer objects retain
`common.Engine`. Public event append clones incoming data before validation
and persistence, then uses separate copies for durable history and context.
Observer delivery clones per recipient. These observations do not establish
the completeness of behavioral acceptance or the later code-quality gate.

The source checker only claims same-module import and common-placement
properties. This incident demonstrates why that limited pass cannot be
reported as proof of owner chains or logger reachability. Preserve the
initial finding and corrected source/evidence for the final comparison;
do not defer the correction to a repair chapter.

## Revision reviewed

The reviewer inspected the corrected declarations and actual call sites,
including recursive validation: diagnostic-relevant llm free functions now
receive the current `common.Engine`; usage accumulation has an Engine
receiver; `eventlog.Dump` receives Agent; configuration normalization receives
its owning Ensemble; CLI configuration has the public owner context. No
extra global logger or sibling-service reference was introduced.

The GUI's public `ClientOwner` now exposes the root logger. Its regression
closes the client, submits a request, and requires the rejection diagnostic
in the Ensemble's captured output. This exercises a real path rather than
checking whether an interface contains a particular identifier.

Agent construction now deep-copies configuration centrally before either
`NewAgent` or `Load` publishes the object. The added public regression changes
both a caller-owned declaration name and its original schema bytes, then
requires the loaded Agent's configuration to remain unchanged. Existing
setters and snapshots retain their copy boundaries.

The coder reports passing main/GUI module tests. The reviewer read the revised
source and targeted regression assertions without repeating the suites. No
remaining material finding was found in this review's scope. Full behavioral
acceptance, actual live feature evidence, and the subsequent independent
first-edition comparison are separate gates. Their later resolution is recorded
in `chapter-02-code-review.md` and `chapter-02-validation.md`; this note preserves
the earlier finding and does not retroactively claim it covered those gates.
