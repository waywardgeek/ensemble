# Independent story-pass review: Chapters 1–3 and 5–6

October 7, 2026. Reviewer: the thread that previously implemented Chapter 3's
human-client integration and subsequently independently reviewed Chapter 4.
This is independent of the manuscript author, not a claim of source blindness
or a new cold-student evaluation. Chapter 4's completed review remains in
`chapter-04-proofreading.md`; it is outside this report's new reading scope.

**Outcome:** the five story revisions are supported and preserve the exercise
artifacts. Two stale Chapter 2 manuscript status statements and related summary
statuses still need author reconciliation before this editorial pass closes.
No implementation change, new live run, or renewed runtime validation is
claimed. Chapters 5–6 retain their explicit pending implementation/live gates.

## Actual reading and scope

Read the complete current Chapters 1, 2, 3, 5 and 6 and their complete outlines.
Read the complete voice v5 and chapter-writing procedure, reloading them after
compaction; read the author's complete `story-pass-2026-10-07.md`. Compared the
full five-file diff against `/tmp/ensemble-story-pass-before/`, not just the
new openings. The existing architecture and global-review map were previously
read during this thread's Chapter 4 review; no fresh whole-book reread is claimed.

Historical source reading was selective and matched the outline decisions:

- First-edition Chapter 1 lines 1–165: bet, requested demonstration, StackAgent,
  deletion/restart, and surrounding original motivation.
- First-edition Chapter 2 lines 1–135: provider-shaped interface and duplicated
  clients, including the numeric/autobiographical claims omitted here.
- First-edition Chapter 3 lines 1–110: ordinary tool-use concentration and its
  private-corpus qualification.
- First-edition Chapter 6 lines 1–120, corresponding to new Chapter 5: correction
  during work, actor ownership, and the two-loops/two-owners explanation.
- First-edition Chapter 7 lines 1–110 and 490–565, corresponding to new Chapter
  6: blank-terminal problem and buffered-writer explanation.

These passages were read in this thread's preceding story-source comparison
phase. They are not described as complete historical-chapter reads. No old
implementation was copied or edited for this prose review.

For evidence, reread Chapter 1's complete `evidence/live.json`; complete
Chapter 1/2 validation records; complete Chapter 2 evidence reconciliation;
all three complete Chapter 2 human `*-main/terminal.txt` files; and the complete
Chapter 3 `human-chat/openai-main/terminal.txt`. The Chapter 3 overwrite episode
also belongs to this reviewer's preserved prior implementation experience.
The other unchanged Chapter 2 machine/public-consumer demonstrations retain
their existing accepted evidence review; this pass did not re-audit every raw
machine receipt. Historical model identifiers remain dated receipts, not
current recommendations. No provider requests were made.

## Story and teaching assessment

| Chapter | Retained stake and resolution | Assessment |
|---|---|---|
| 1 | A two-week demonstration produces a working prototype, then a deliberate restart; the reader receives ownership lessons before constructing the first conversation. | Restores the consequential choice lost by the biographical summary. The wording attributes the account to the first edition and avoids asserting a newly measured superiority result. Omitting acquisition prices and exact dates is justified. The architecture still precedes the first code example. |
| 2 | A nominal interface still asks every provider to speak the first provider's types; the reader later needs the real prior request after a bad answer. | The sharp interface observation explains neutrality; `port=8080` and redaction sustain the debugging purpose through the mechanism. Historical line counts and migration duration are appropriately omitted. Status contradictions below need correction. |
| 3 | The model claims recovery after omitting overwrite permission; the human checks the log, corrects the instruction, and obtains changed bytes. | The actual transcript supports the incident and its distinction between narration and effects. The model's parallel-execution explanation remains visibly a model claim. Moving repeated acceptance counts behind the validation link improves the ending without removing the Unicode or verifier lessons. |
| 5 | A correction admitted only after completion becomes a review comment; one actor preserves both blocking and asynchronous public entry points. | The two-loops/two-owners explanation earns its place and connects the familiar API to one mutable-state owner. The prose correctly distinguishes asynchronous jobs from responsive turn admission, and avoids the old implication that synchronous APIs must serialize independent Agents. |
| 6 | The reader sees a proposed bad write before acceptance, and needs a terminal that actually displays fragments. | The buffered writer's technically correct but useless behavior gives the timing test a human purpose. Held-open observer and PTY checks resolve that problem. Plain/stream equivalence remains a controlled-fixture claim, never a promise of identical paid generations. |

Paragraph endings and pace vary; the openings lead into specific mechanisms
rather than generic encouragement. The person-gap warnings remain substantial,
but the long sections continue to discuss concrete reader or caller failures.
No invented anecdote is needed merely to satisfy that counter. The soft warnings
are recorded below, not represented as absent.

## Findings sent for reconciliation

1. `chapter-02.md` §2.10, immediately after the human feature-ledger link,
   says independent review of the client remains open (reviewed line 890).
   Its closing paragraph says final human review/acceptance remains pending
   (reviewed line 1018). Both contradict the accepted `ad0d80e3` opening and
   `chapter-02-validation.md`. Replace those current statuses while retaining
   initial-versus-reviewed receipt chronology and separate Bill approval.
2. `chapter-02-evidence.md` opens with human validation reopened and its human
   correction summary still says revisions await re-review. Update the current
   summary to the accepted record. Historical pending entries within the
   chronological account can remain when clearly followed by their resolution.
3. `chapter-05-outline.md` still calls Chapters 2–4 human corrections active.
   Reconcile this handoff status with the accepted Chapter 4 transition; do not
   accidentally claim Chapter 5 implementation or live validation is complete.

Findings were sent to the author and coordinator. A later attempt to resume the
author failed because the orchestration tool reported a thread limit, so the
coordinator received the outstanding locations. No author-owned prose was
edited by this reviewer. This report remains open for that small status-only
resolution; it does not manufacture an accepted author response.

Separate from the story pass, Chapter 2's new immutable-log-path paragraph is
a normative clarification. Its earlier-snapshot propagation needs the
coordinator's code checks; the runtime correction reviewed in Chapter 4 does
not by itself certify the earlier frozen exports. That issue was flagged to
the coordinator without editing code or asserting an unperformed backport.
Chapter 1's private-runtime clarification follows Bill's explicit ruling.
Neither change is mislabeled as mere editorial polishing.

## Artifact and check results

Independent extraction from the saved pre-story baseline and current files
compared complete fenced blocks and contiguous Markdown table blocks:

| Chapter | Fenced blocks | Tables | Byte equality |
|---|---:|---:|---|
| 1 | 8 | 3 | Equal |
| 2 | 12 | 10 | Equal |
| 3 | 8 | 5 | Equal |
| 5 | 4 | 3 | Equal |
| 6 | 6 | 2 | Equal |

All 38 fenced blocks and 23 tables match. The entire prose diff was read;
the separately identified normative clarifications are visible in that diff.
No example or fixture changed in the story edits.

`go run ./cmd/lintprose` with the five chapter paths passed every hard check.
Word counts were 4380, 7287, 5681, 4254 and 5062. Soft warnings: person gaps
4329, 5849, 2754, 2391 and 4625 words respectively; Chapter 2 had 20 counted
negation forms against a soft allowance of 14, and Chapter 6 had 13 against 10.
Scoped `git diff --check` passed. No runtime grader was rerun for this prose-only
review; prior validation stays attributed to its existing source and receipts.

## Reviewed identities

These hashes identify the working manuscripts read, not a new accepted runtime
snapshot. Baseline hashes identify the actual temporary saved files used for
artifact comparison; their continued availability is not assumed.

| Chapter | Current SHA-256 | Pre-story baseline SHA-256 |
|---|---|---|
| 1 | `42307cda20c9ac340805ddace287a97707f7e80b0e508777f3408683324e1881` | `2906097109361d480882cfa9c2c2eb913de2e142109505bdeb4d9955842659ef` |
| 2 | `e1f2329132844cf4a4ffc4ff8e8605cd5923ce6ef4c84d887f35d069b015a029` | `edaa0dc2bd12989d2a18118aaf8f462c8d0117d78b6d422dc19ce08c9a4ff67c` |
| 3 | `515e8bea29e0143d3f0527117de9fcb1fbb0d50b1d5888098ae2d7d6cec6b480` | `0c3ddf8a51c6a39f8db614eecf404b971a96c70bd1dd2a0bb7df9a01928a83ba` |
| 5 | `14a83e68277cbb31aa8bbca55e689e202b312ff5db5b62f9f4ddbc5c1fce14b4` | `091a50b03e6f472231ca4698411b23e3bb779eee425f81b0fba3f7c83ec836ae` |
| 6 | `2b5960dfcd8c77ee4a0194d97b3ee06a23e128f45e4d4813903c6a1613796e54` | `0fbf74bd06eb15b29ced9b30280a7a561aab5cc100dbf1e9753008f5d240bfaa` |

Voice v5 SHA-256: `17883353cf9653c0df46081c216b5b1623dfa640619e06fe89bff46fc29dfad1`.
Procedure SHA-256: `826a6138c85deb85ba08362075dc8423ee29887af2be4510bf4462c54391fbac`.
