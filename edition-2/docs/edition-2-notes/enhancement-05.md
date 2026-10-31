# Enhancement notes: Chapter 5 (original Chapter 6)

> Historical paths: the `edition-2/ch04`–`ch09` paths recorded below use
> original student numbering. Those exports are now preserved under
> `solutions/edition-2/history/`; see the [solution mapping](../../solutions/README.md).

Author: `enhancement_author_resumed`, October 10, 2026. Final prose submitted for
independent combined review; grader/sample evidence reconciled. Mandatory handoff:
`.agents/skills/ensemble-coding/SKILL.md`.

## Source and preservation

Read all 939 lines of original `book/chapter-06.md`, file revision
`e1605916114230570b348ab5961c193fa1cebf80`, SHA-256
`fc82efd7fc0caaa9805a7bfa7f2b8b457cfb096b341f0c8f92d60b672dff2495`.
Read all student/coordinator notes and comparative review, accepted actor,
public root, shared vocabulary, CLI, media renderer and external workflow source,
and retained live artifacts. End reference is `solutions/edition-2/ch06`,
accepted at `edition-2-ch06-r3`. Historical paths and tags remain unchanged.
Governing skill/carryover/voice/procedure and enhancement plan were read; this
bounded enhancement supersedes the retired author-first procedure.

Preserve the original sequence, title and argument: the user trying to speak
while a tool runs needs an owner that can hear, not another synchronous loop.
The GUI-first dependency trap, twenty-three/thirteen tests and five commits,
load-bearing observer fix, mailbox explanation, 3 AM concurrency warning,
five bugs in 120 runs, and 307,984/3,447-byte history-versus-observation lesson
remain. The chapter marks these first-edition CodeRhapsody measurements as
historical. It preserves the deaf/listening closing line. The stale promise
that the GUI comes immediately next is corrected to streaming, then GUI.
The human CLI already exists, so the old nowhere-to-type cliffhanger now
acknowledges the terminal rather than erasing an available feature.

No standalone refactor, second actor root, generic workflow language, grouped
completion API, spawning tool or premature streaming implementation is assigned.

## Recorded findings and dispositions

| Finding | Enhancement |
| --- | --- |
| Bill's interrupt ruling versus stopped-actor TL;DR/sample | Remove the actor return/discard semantics. Preserve current-turn interruption, new input, continuing jobs and actual late output; distinguish job kill and shutdown. |
| Inconsistent ownership/service sketches | Agent owns its lazy actor/mailbox; Engine performs one exchange, History owns reduction/persistence. Worker outcomes enter the actor; actual Ensemble retains peer Agents and tags observations. No service bag or second Framework root. |
| Text-only PartFinal versus full tool content | Show full neutral Part, including call/result identity. Progress is not the durable event log or a replay substitute. |
| Blocking versus live input | Ask/Submit owns a buffered reliable reply and queues a new turn; Post text uses current Agent state and may become a hint. Observe/Wait/WaitAny remain bounded progress. |
| Save ordering, shutdown and round limits | Preserve original save-before-completion and worker-join lessons, clarify caller-owned persistence when LogPath is empty, keep existing default 200/N batches. No new setting is introduced. |
| Hint during HTTP and replay cutoff | Record request boundary before transport leaves the owner; hint belongs to next request. Compare captured vendor bytes with rendering the prefix before request_sent, since applying the marker consumes pending ephemera. |
| Late report after interrupted call pairing | Explain real ongoing-job report and no duplicate tool result; accepted job_report is identified as implementation vocabulary. |
| Bill withdrew grouped wake | Replace TL;DR rule 5, §5.6 and prospective supervision promises with distinguishable peer Agents and sequential reliable completion. Original wake-once label remains historical; new label multi-agent-observations. The initial reviewer/student agreement and Bill's correction remain in the lesson. |
| Missing exercise protocol | Publish sourced think seconds/thought, targeted hint, observation discriminator, stage turn_ended and pipeline markers. Extra records belong to application, not core observation types. Separate-module public boundary is explicit. |
| Manuscript tool schemas | Explain accepted example's string text, empty count/read arguments and decision/notes as application choices; actual stored text and distinct role tools remain required. |
| Unknown text model refusal defect | Media table does not ban unknown text models. Local file bytes must become real attachment fields; unsupported/unknown media fails with model/category. URI/handle attachments remain deferred. |
| Media mutation survivor | Preserve the wrong mimeType survivor and field-level test repair, distinguishing sensitivity from shared fixture assumptions. No new paid matrix. |
| Source preview and actual evidence | Replace the false streaming-chunk live demo with explicitly illustrative complete-part observations plus actual October 9 CLI/workflow/media receipts. Own-module build commands, safe credential reference and scratch instructions replace historical reference-binary/key commands. |
| Small inherited accuracy defects | Base64 overhead is approximately one third, not double; PartID correlates parts rather than eliminating buffers; late observer attachment is not replay; Post's short mutex is not a lock-free guarantee. |

## Grader coordination and sample review

Coder validated the final seven checks/100 points:
`ch4-parity` 10, `not-deaf` 25, `replay-is-live` 15, `observer-fires` 15,
`multi-agent-observations` 15, `loud-refusal` 10, `hub-clean` 10. The chapter
publishes `-workflow` and `-media-models` file inputs alongside inherited
consumer inputs; they configure only grading. Default model identifiers are
historical fixtures, not availability claims. There is no silent core fallback
when the separate workflow does not build.

The not-deaf assertion orders the hint before its correlated think report,
which may be a running-job callback; it does not claim execution ended.
The workflow's three text responses and markers do not prove role-tool storage,
actual edited handoff or reliable request delivery under saturated progress.
Those remain explicitly required source/public-API integration evidence. The
chapter does not turn a numeric grade into an ownership proof.

Changed Go examples are explicitly schematic: full neutral observation fields,
public input separated from private worker messages, an internal FIFO, Agent
runtime dispatch without interrupt return, request-specific Ask and an existing
Ensemble observation envelope. The coder compared them with accepted source;
they remain schematic rather than independently compilable programs. All eight
JSON fences parse. The exact prompt/hint and think-argument fences ran through
the accepted separate workflow with four actual fake-provider HTTP requests:
hint observation index 5 precedes correlated tool-result index 8, followed by
ordered author/editor/reviewer completions. The exact model map passed the
three-adapter local-media probes. Module-root build and no-argument workflow
launch were exercised in scratch. No prose-contract mismatch was found.

Sample analysis initially selected the tool call sharing the result's ID,
producing a false failure. The raw transcript was already correct; the coder
fixed the analysis to require the tool-result part. No runtime or manuscript
change was made for that false alarm. Full sample evidence is in
[the coder submission](enhancement-05-code.md).

Coordinator draft clarifications: removed the unsupported exact one-second
hint timing from rule 2, and distinguished Submit returning its reply channel
from Ask waiting on it. Neither changes behavior or introduces an API.
Reviewer precision suggestions also applied: external Agent identity tagging
does not claim the event vocabulary stayed unchanged, and Post does not wait
for execution/a receiver while its enqueue still uses a short mutex.

## Retained live evidence and validation

Independently summed response usage from the five actual Agent JSONL logs:
CLI Anthropic 4,431/177; Gemini author 5,695/1,157, editor 1,310/591,
reviewer 663/324; OpenAI image client 63/4. All cache counters are zero.
The reviewer independently confirmed those totals and the exact CLI hint,
interrupt, new COMET answer and REAL_LATE_RESULT timestamps. This is one
feature-specific run per provider, not all features across all providers.

The retained CLI record includes the initial premature job_finished/running
metadata and its later guard repair. The actual output was retained, so no
paid repeat was needed. Gemini's author wrote 59 words, counted, added night,
rewrote and counted 60; the editor's stored text then reached the reviewer.
The separate OpenAI client consumed a local 96×96 red PNG and refused audio.
Other positive media formats have deterministic wire evidence, not invented
live demonstrations. Links point to actual files, not a nonexistent live.txt.

Original grading remains 80/100 FAIL, with independently accepted logger-shape
and text-only refusal exceptions. Original grouped-wake PASS did not prove the
withdrawn claim. Legacy original grading remains 100/100 PASS.

The full revised runs returned accepted 100/100 and legacy 25/100. The legacy
replay case had rejected a routing fixture model before HTTP, so that failure
did not test replay. After the coordinator approved the historical workflow's
`fake-model` for capture and rendering, both implementations passed all three
focused request comparisons. The final revised scores are **composed** accepted
100/100 and legacy 40/100: retain the other six actual checks and substitute
only those final replay results. These are not described as fresh full runs.
Legacy's actual local-media failure also persists with its own declared models;
the record distinguishes that behavior from an identifier mismatch.

Seventeen runnable targeted defects failed the intended new checks. Compile
failures from a malformed initial mutation and unrelated concurrent grader
work are preserved as setup failures, not kills. The corrected replay mutation
was rerun after the routing-fixture change. Final affected grader tests/vet
passed in the coordinator-approved stable checkout, whose overlay hashes match
the submitted files. Focused accepted-source race tests verified hint delivery,
private replies under progress overflow, observer content/identity, interrupt
and live jobs, plus actual workflow role tools and edited-draft handoff. Earlier
unchanged mutations/live evidence are reused with their historical identity.

Author alignment is complete; independent combined acceptance remains pending.

Final submitted source-relative comparison retains 672 of 939 original lines,
adding 421 and removing 267. Changed spans chiefly repair repeated ownership/interrupt
sketches and the misleading exercise/demonstration, while original incidents
and progression remain. Local links resolve and every JSON example line parses.
Prose lint hard checks pass after contract reconciliation (6,607 words versus
original 5,052). Inherited/enlarged negation and person-gap warnings remain for
independent judgment. Submitted manuscript SHA-256:
`19a847585e17f658e930cce3512c22d0cd7b1e684511b7c8ae162756801a172b`.
No original/frozen source, score or tag changed;
no commit, mock, paid test or later chapter draft was made.
