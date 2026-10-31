# Edition 2 solutions

Accepted cumulative solutions live directly in `ch01/`, `ch02/`, and so on.
[`agent/`](../agent/) is the authoritative development tree; snapshots are reader
references. Chapters 1–24 and 26 are accepted. Chapter 17 is an ungraded field
report; Chapter 24 is optional and ungraded. Chapter 25 remains the reserved
Codex-to-Ensemble crossover, with no accepted export.

The October 14 dependency order teaches accounting, authentication, confinement
and fleet management before background memory. Chapters 15–21 replace the former
17–23; memory and recall are now 22–23. The cumulative sources were reconstructed
and independently reviewed at each dependency boundary, then exported from the
exact accepted commits. Renaming the former cumulative answers would have kept
memory in chapters that precede it. Chapters 1–13 remain unchanged; Chapter 14
also receives the accepted handoff correction. See the [chapter map](../docs/edition-2-chapter-map.md)
and [review record](../docs/edition-2-notes/codex-compaction-runtime-audit.md).

Each manifest identifies the source commit, subtree and file hashes. For rebuilt
Chapters 14–24 and 26, `edition-2-background-chNN-r1` identifies that exact source commit.
These are new checkpoints; earlier tags have not moved. The table retains the
original-numbered checkpoints for unchanged Chapters 1–13.

| Current chapter | Accepted solution and manifest | First-edition source chapter | Source checkpoint |
| --- | --- | --- | --- |
| 1 | [ch01](ch01/) | 1 | `edition-2-ch01-r1` |
| 2 | [ch02](ch02/) | 2 | `edition-2-ch02-r3` |
| 3 | [ch03](ch03/) | 3 | `edition-2-ch03-r3` |
| 4: jobs and external consumer | [ch04](ch04/) | 5, including 4 | `edition-2-ch05-r2` |
| 5: actor and steering | [ch05](ch05/) | 6 | `edition-2-ch06-r3` |
| 6: streaming | [ch06](ch06/) | 7 | `edition-2-ch07-r4` |
| 7: GUI and replay | [ch07](ch07/) | 8 | `edition-2-ch08-r3` |
| 8: workbench and settings | [ch08](ch08/) | 9 | `edition-2-ch09-r2` |
| 9: Skills | [ch09](ch09/), [manifest](ch09-manifest.json) | 10 | `edition-2-ch10-r2` |
| 10: Persistence | [ch10](ch10/), [manifest](ch10-manifest.json) | 11 | `edition-2-ch11-r1` |
| 11: MCP | [ch11](ch11/), [manifest](ch11-manifest.json) | 12 | `edition-2-ch12-r1` |
| 12: The Agent Sees Itself | [ch12](ch12/), [manifest](ch12-manifest.json) | 13 | `edition-2-ch13-r1` |
| 13: Speech and blind persona | [ch13](ch13/), [manifest](ch13-manifest.json) | 14 | `edition-2-ch14-r1` |
| 14: Context management | [ch14](ch14/), [manifest](ch14-manifest.json) | 15 | `82d3952a`, `edition-2-background-ch14-r1` |
| 15: The Invisible Invoice | [ch15](ch15/), [manifest](ch15-manifest.json) | 18 | `6edde040`, `edition-2-background-ch15-r1` |
| 16: Subscription authentication and reasoning | [ch16](ch16/), [manifest](ch16-manifest.json) | 19 | `ec853c18`, `edition-2-background-ch16-r1` |
| 17: The Crossover | [ch17](ch17/), [manifest](ch17-manifest.json) | 20 | `a4b9a1d3`, `edition-2-background-ch17-r1` |
| 18: Web access, status and accounting | [ch18](ch18/), [manifest](ch18-manifest.json) | 21 + 22 | `6677cb9c`, `edition-2-background-ch18-r1` |
| 19: The Sandbox | [ch19](ch19/), [manifest](ch19-manifest.json) | 23 | `dee87fbc`, `edition-2-background-ch19-r1` |
| 20: The Ensemble, including reusable GUI | [ch20](ch20/), [manifest](ch20-manifest.json) | 24 + 25 | `9f4fdca2`, `edition-2-background-ch20-r1` |
| 21: Fleet management and GUI | [ch21](ch21/), [manifest](ch21-manifest.json) | 26 | `3a5673b6`, `edition-2-background-ch21-r1` |
| 22: Background memory | [ch22](ch22/), [manifest](ch22-manifest.json) | 16 | `c67173c1`, `edition-2-background-ch22-r1` |
| 23: Auto-Recall | [ch23](ch23/), [manifest](ch23-manifest.json) | 17 | `8faca265`, `edition-2-background-ch23-r1` |
| 24: Optional Codex workaround | [ch24](ch24/), [manifest](ch24-manifest.json) | New in Edition 2 | `f18737f6`, `edition-2-background-ch24-r1` |
| 26: Gateway | [ch26](ch26/), [manifest](ch26-manifest.json) | 27 | `59e25a25`, `edition-2-background-ch26-r1` |

Original Chapter 5 is absorbed into Chapter 4, original 21+22 into current 18,
and original 24+25 into current 20. The remaining `ch05/` is the actor chapter.
The first edition's 26 original chapters still map to 23 second-edition chapters;
Gateway follows as 26, with optional 24 and reserved 25 between them.

## Checks and historical results

Current revised graders cover Chapters 1–16, 18–23 and 26:

```sh
go run ./cmd/grade-edition2 -ch 14
go run ./cmd/grade-edition2 -ch 22
```

The command defaults to the matching `solutions/chNN` directory; an explicit
submission path selects another implementation. Graders exercise local fake
services without paid inference. Structural/source review and live behavior have
their own evidence; a numeric score cannot establish every chapter obligation.
The [carry-forward record](../docs/edition-2-notes/codex-handoff-carryforward.md)
records corrected 14–21 checks, and the [reconstruction note](../docs/edition-2-notes/codex-reordered-foundations.md)
records memory, recall and optional-route checks. The [final default-path receipt](evidence/background-compaction/reordered-final.json)
records passing grades on the installed exports, package checks and exact source
hashes. No new live use is claimed by copying an accepted implementation into a different cumulative chapter.

The complete former snapshot directories and manifests are preserved verbatim at
`edition-2-before-background-compaction-20261014` (`5d15954e`). Existing original
checkpoint tags remain immutable too, including the old-order
`edition-2-ch16-r1` memory, `edition-2-ch17-r1` recall,
`edition-2-ch25-r1` fleet and `edition-2-ch26-r1` management snapshots.
For example, `git show edition-2-before-background-compaction-20261014:solutions/README.md`
shows the former mapping. Its directory numbers and scores describe that order.

Original graders remain under `edition-1/`, with original numbering; run them
from that archived module against the intended source. [Chapter notes](../docs/edition-2-notes/)
and [`evidence/`](evidence/) retain their original numbers and actual outcomes.
Gateway's historical original 0/100 field-spelling exception and web's original
0/100 launch/protocol exceptions remain failures, alongside their independent
acceptance and separate revised results. The historical crossover's bounded
Astra/high coding demonstration is not a new run of the reconstructed Chapter 17.
See the [enhancement record](../docs/edition-2-enhancement.md) for earlier revised
scores and their source identities.

Use older supported Anthropic models such as Sonnet-5; Bill excludes 5.5 models
from supported use because of their ephemeral-context behavior. Historical
snapshots/results retain their actual model identities. Subsequent authorized
coding uses GPT-6 Astra/high. Auto-Recall preserves Bill's final-selection quotas
and searches the corpus while skipping content already in actual context.

[`history/ch05/`](history/ch05/) remains the sole retired history directory.
There are no new history folders or symlinks. Keep snapshot changes tied to an
explicitly accepted replacement and fresh checkpoint; develop in `agent/`.
Nested example names and historical evidence paths are intentionally retained.
The manuscript order and these source exports do not rebuild or publish the
assembled Markdown, EPUB or PDFs. Bill's preface and those reading-copy artifacts
remain unchanged.
