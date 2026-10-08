# Chapter 3 log destination correction

This is a coordinator-authored scoped correction to the accepted second-edition
source at `1a61e1f2487cdc94ee65bd0e1593065cc1e95f45`. It is not a new cold-student attempt. Chapter 2 now
teaches that an Agent's log destination stays fixed throughout its lifetime.
The correction was already present from Chapter 4 onward; this branch applies
that invariant directly to the earlier accepted source without copying later
implementation.

`SetConfig` now rejects a different or empty destination while holding the
configuration lock, before changing any configuration. A same-path model update
remains valid. The public regression checks both rejection cases, complete
configuration preservation, absence of a second log, actual model selection,
and continued persistence to the original file.

`original-control.json` reruns the new test on the original production file:
exactly the two rejection subtests fail, while the same-path positive passes.
`local-checks.json` records all modules' vet/tests, main race and formatting.
`acceptance-checks.json` links inherited grading and independent offline and
human-interface checks. These are deterministic local checks, not paid API runs.
`repair.json` binds the two changed source files.

All earlier live receipts retain their original identities. This refusal occurs
locally before a model request and changes no successful wire request, tool or
CLI path, so the unchanged paid demonstrations are retained without relabeling.
Independent review and a new immutable revision checkpoint remain pending.
