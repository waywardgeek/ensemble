# Edition 2 Chapter 22 author record

Author: `edition2_author`. Assigned October 12, 2026 after Chapter21 acceptance
`d62ab24c`, checkpoint `7e6dc5bc` / `edition-2-enhanced-ch21-r1`.
Status: final author submission; exact examples and revised grader evidence
reconciled below, including the independent attribution correction. Combined
acceptance remains the independent reviewer’s gate.
No Chapter23 authorship is released by this record.
Mandatory technical skill: `.agents/skills/ensemble-coding/SKILL.md`.

## Reading and context

The author read both complete originals and all four complete original-numbered
student/comparative records before drafting: original24 1–300/301–565;
original25 1–300/301–582; ch24 notes315, review242; ch25 notes349, review243.
After compaction and before any Chapter22 writing, the author reloaded all four
required documents fully, without truncated output:

- Workflow2233: 1–350,351–700,701–1050,1051–1400,1401–1750,1751–2100,2101–2233.
- Carryover981: 1–340,341–680,681–981.
- Voice632: 1–340,341–632.
- Chapter procedure599: 1–320,321–599.

The full373-line coding skill was reread, alongside current AGENTS authority,
enhancement opening and the chapter map. Both original chapters were then reread
in full (24:1–290/291–565;25:1–300/301–582). All four full documents remained
active together during this initial draft. Any subsequent compaction requires
another full reload before writing or reconciliation.

Source bindings:

- Original24 revision `2ca63636c8ba592ea0ed190fb0a3c4731e2b0848`, SHA-256
  `c6ac58046ed8e5496c44931e20d126d36081325fc06bfc27f4276c73424976b1`.
- Original25 revision `009d74c70b108681b2107b4c3a08685508567cee`, SHA-256
  `eb09ec25fe8a681479310ffe22f4cf19d0b4eba34bf507dbbfd8181f7302616f`.
- Accepted combined source `e00d5bcaf1412c50787428ef1fa08bed59b99e11`, main tree
  `30fb05d154b9808eb88622aaff9c1424cd7e8af1`, direct `solutions/edition-2/ch22`.
- Intermediate original24 accepted source `0430df4340fffac5234a94aea843f9a7c9d02282`,
  tree `8555e3cb9ab5200f0a1c52691007a7fc4a1c6797` remains separate.

Read actual public Ensemble/fleet construction, internal fleet and shared/tool
interfaces, terminal construction/reply handling, actor completion correction,
GUI construction/observation/assets/relay and external reuse/browser fixture.
The existing root supplies construction and observations; the actual Agent owns
state. No callback-bag design or new runtime is proposed.

## Through-line, outline and preservation

Stake: Bill wants reusable views and more hands without losing the ability to
steer; the deployed application must connect the public boundaries its library
already exposes. Resolution: an actual outside GUI consumer and a terminal
fleet that can answer questions, preserve a running join and submit tested work,
with the observed failure to obey a delivered hint retained.

The merged outline follows reusable GUI → actual root/Agent construction →
embedded/components/relay proof → two kinds of fleet silence → spawning and
identity → asking versus submission → message wake → bounded correction →
historical product wiring failures → runnable combined exercise → actual runs.
Both assignments appear explicitly in TL;DR and exercise. The fourteen original
rubric rows retain GUI100 + fleet100, pending final predicate reconciliation.

Preserved original material:

- Homebrew-VTT's desired Artifact component and the inaccessible-public-boundary
  irony; the historical two paths to Gorilla and born-unused exports.
- Current-model getter motivation, binary-without-its-interface deployment,
  positive/negative import controls, outside event vocabulary and public relay.
- Build-tool wrapper holding inherited pipes, and affected harness paths after
  a package move. These are historical reports, not new observed failures.
- Bill's listening/steering motivation; delegating mistaken for completion;
  combined idle/wait hazard; full Agent children and filesystem ownership tree.
- User as an address rather than a managed child; why asking and submitting must
  differ; fan-in yielding a child question; background job lifetime; three wakes.
- All five historical executable wiring failures, plus the package-counter and
  misleading relative-path bisect stories, scoped to the reported tests.

There is no length target. Replaced material consists chiefly of the superseded
refactor instructions/AgentHooks design, unsupported code-size inventories,
universal success/security claims and duplicated original check tables. Useful
explanations remain; a new complete public consumer and terminal setup make both
halves reproducible. The original source-relative diff, not the word count,
governs the preservation review.

## Teaching corrections and their evidence

1. `go list -deps` measures import reachability, not every linked machine-code
   byte. Internal-package restrictions really apply to an outside consumer;
   importing internal/llm is no escape hatch. Keyed struct literals can omit
   fields. Official Go pages checked October12:
   https://pkg.go.dev/cmd/go#hdr-Internal_Directories and
   https://go.dev/ref/spec#Composite_literals. Existing carryover records all
   three corrections. Actual public root/Agent construction replaces AgentHooks.
2. Embedded contents are read-only, but an exported variable binding is not
   immutable. Accepted gui/assets.go uses a private anchor and Assets() fs.FS;
   https://pkg.go.dev/embed checked October12. Explicit relative overrides still
   depend on cwd; the chapter recommends an absolute override for relocation.
3. An asset HTTP200 can be a directory listing. The original24 student sensitivity
   correction required actual shell identity; this positive-result lesson stays
   in prose. Component-only serving is distinct from actual Chrome rendering.
4. Public relay uses real listeners, original caller IDs and actual browser
   forwarding. No-browser error, connected-silent wait and client/browser/server
   closure have different conditions. Accepted gui/relay.go and its outside
   reuse tests supply the mechanism, without imposing a new timeout.
5. Routing IDs do not replace owner interfaces. Actual factory receives parent
   Agent and spec, descendants reach the shared root, and state directories are
   separate while task Workspace remains shared. Chapter21 narrowing remains.
6. Question IDs belong to the recipient pending map. Same-name parent/child
   overwrote a human question before the accepted correction; omission of a reply
   ID is allowed only when selection is unique. Accepted ch25 review/probes bind it.
7. Message wake is independent of callback delay. tool_limits remains the prior
   one-shot next-call mechanism, not the original25's unsupported per-name map.
   Preserve running jobs and later collection under the same handle.
8. Factory installation enables library correction for every fleet root/child.
   Earlier single-Agent embedders without a factory keep normal response completion.
   Exactly three framework wakes precede escalation; successful submit resets.
   Exhausted child state must be failed, with no invented result. Failed turns
   do not enter the correction loop. Actual actor finish and fleet Idle bind this.
9. Remove universals that every chapter/mutant passed, every failure mode is
   prevented, or a broad six-row failure proves no narrower mutation exists.
   Preserve the original real wiring lessons without adding their old repair
   sequence to the new assignment.

## Historical results and live provenance

Original24 grade remains0/100 on its recorded launch/import assumptions;
original25 grade remains0/100 before inference with no supplied model. Accepted
exceptions did not turn those into passes. Neither score describes the revised
combined grader. The original24 source review accepted three narrow fixture
exceptions; original25 accepted missing-model and factory-enabled compatibility
exceptions, both at initial review, zero formal rounds.

Directly read coordinator live receipts under evidence/ch24 and ch25. Original24
live source `35df549210130f824fdc2a4eb632ea878b2c529b`, tree
`5b36f28cbc8bac320febe1a9afa810efd27452e3`:9.649seconds, exit0,
three Astra/high requests/completions, seven assets, actual42-byte file and
matching browser DOM. A large combined live-results read truncated; only its
fully read coordinator receipt supplies the published aggregate claims. Component,
relay and switching proof retain their separate actual test/live sources.

Fleet live source `891d513bf142013d6888760cfae4fcaab1edd680`, tree
`6cbf18933f82ee450d91750442161b176e469f8e`:48.198seconds, exit0,
20Astra/high requests/completions, children add/multiply/subtract with saved
conversations and successful independent combined assertions. Join job12 starts
18:30:55.751093Z and returns running18:30:55.887642Z after steering; the same job's
completed result18:31:10.340664Z contains all three submissions. Request15 includes
the steering text; final submit omits STEER_COBALT. Delivery is demonstrated,
compliance fails. The malformed initial child shell command and corrected success
are retained; the concurrent wire trace is not claimed to pair requests by ordinal.
Final accepted production matches that live implementation; later README/test
changes do not relabel the paid run as a final-source rerun.

## Coherent-stage validation

Direct hard prose lint passed on the coherent draft (6527 prose words);
soft person-gap warning is considered without inventing a story. The chapter
will be reread and checked after final sample/rubric reconciliation. No runtime,
frozen export, first-edition or concurrent main/crossover file has been edited.
Exact new public consumer, embedded/disk/component/relay guidance and typed fleet
launch are being handed to the grader coder for technical validation against
external peers. No new paid run or autonomous-model quality result is claimed.
Independent whole-chapter review and the final combined grader freeze remain.


## Final author reconciliation, October 12

Read complete coder report at `a02dc4ce`, final-source.json, samples.json,
samples-run.json, samples-go-bindings.json, both actual final full run receipts,
and complete sensitivity.json/sensitivity-corrected.json. Then read coder
follow-up `693cece4`, its updated report, post-review-source.json, corrected
row-sensitivity receipt and focused baseline/restored outputs. The author separately
recomputed the current manuscript and all sixteen code-fence hashes. They match
the measured sample exactly:
`d2f514380905245e14720fbf3d9d9a1a6555e813077487da3385ab5505bfdf3e`.
No sample change or new execution was needed. The self-review refinements after
`0790bb70` change prose only: child state is not process-crash isolation, restart
loads saved conversation before the new task, inherited historical absolutes are
narrowed, and the numeric question row explicitly describes three generations.

All four full required documents remain active from the complete reload recorded
above; there has been no subsequent compaction. The full mandatory
`.agents/skills/ensemble-coding/SKILL.md` remains active too. Independent reviewer
reports the whole original-relative prose clear; acceptance remains the reviewer's
decision after this final submission; the reviewer has independently resolved
the attribution correction.

### Exact public examples

The actual outside application compiled from the published Go fence and the two
module replacements. Real Chrome rendered the component-only `/demo`, while
`/index.html` returned404. Full mode replayed the public human event, delivered
actual prompts to the real Agent, switched captured requests Astra→Sol, retained
relay caller/result correlation and served the distinctive disk override. The
headless and GUI dependency commands provide negative and positive controls.
Every launch used its newly printed listener addresses; no port8088 or live-user
process was touched.

The remaining Go snippets are accurately classified: the constructor signature,
asset accessor, factory type and callback bind to actual owning source tokens;
the partial AgentSpec excerpt is vocabulary, not a standalone definition. The
complete public consumer is the executable compilation proof. No conceptual type
fence is described as a runtime behavioral test.

The published fleet setup and both prompts ran through the actual CLI against a
faithful scripted Responses peer, with printed prose line wraps folded into one
terminal input. Actual add/multiply/subtract children wrote files, ran delayed
assertions, asked/answered the negative-integer question and submitted. The
ordinary hint woke join job13 before final collection under that same handle.
The parent combined check and independent Python assertions passed; normal EOF
retained each child's save. Receipts contain22 main requests (2GUI+20fleet) and
four successful auxiliary empty-judge responses. The fake API-key environment is
one of the stated launch choices; no production profile or paid provider was used.
Scripted decisions establish plumbing and actual tool effects, not a new autonomous
coding result or evidence of model obedience. The older real run still omitted
STEER_COBALT, and its9.649/48.198-second provenance is unchanged.

Sample validator mistakes remain in their original artifacts: a public human
replay was mistaken for a final Artifact, the expected settings-envelope name
was wrong, and auxiliary judges were incorrectly required to carry main-request
reasoning settings. That last task-completing run forced judges into fallback;
the corrected peer produces four genuine successful[] judgments. No runtime or
manuscript change was used to accommodate those fixture assumptions.

### Revised grader and scope

The actual full runs bound by `a02dc4ce` each score200/200: accepted2.286seconds,
legacy1.615seconds, both exit0. Focused race passes2.219seconds and vet is clean.
These are complete measured runs at that source, not sums or retellings of the
original0/100 scores. The final attribution-only reviewer follow-up is recorded
separately below; it does not relabel these earlier source bytes.

Fourteen original weights remain: GUI100 and fleet100, no parity. The grader
uses real outside clients/CLI/GUI and external boundary peers. Its historical
`-reuse-legacy` option adapts the old constructor to actual Agent behavior rather
than fabricating a Send callback. The current book retains actual root/Agent
ownership and the optional-module boundary. Human replies, same-name questions,
recursive lifetime, restoration and failed-turn nonretry retain their stronger
accepted built-CLI/source proof; numeric three-generation messaging alone is not
that proof. Component rendering has the actual Chrome receipt, separate from
numeric HTTP asset delivery. The code report's twelve inherited file bindings
retain their original accepted revisions and corresponding evidence limits.

The first accepted fleet failures were save-text/handle-wording/exhaustion
fixture assumptions; the initial transcript is explicitly labelled a transcription.
The accepted130/200 full result used an absent ServeWS method. Legacy60/200
combined a missing directory with misclassification of the post-submit response;
legacy185/200 then rejected valid JSONL child journals. Corrected actual full
results are200/200. None of these setup failures is called a repaired product
bug or evidence of every masked downstream failure.

### Sensitivity limits

Twelve distinct source defects were attempted. Nine initially failed their
intended observations. Reply-sender removal and final-Artifact omission initially
passed: a correlation ID contained the parent's name, and turn-ended text stood
in for part_final. Corrected predicates exclude the ID and inspect actual final
Artifacts; both same defects now fail their intended rows. Hidden live-child
status initially failed only through a coupled idle timeout; the focused repair
also reads the already captured status and directly reports the missing child.

Shared child saves and the live-child case still have coupled failures. The
message-wake omission is a causal liveness check with the command held and a
long independent callback timer. The nine unchanged cases retain their earlier
pre-correction source attribution, including the lack of a contemporaneous full
hash manifest; they are not represented as fresh final-source runs. Focused
restored controls pass. Twelve detected defects with overlapping assertions do
not establish twelve independent guarantees or complete correctness.

The independent reviewer requests one final score-attribution correction: the
observed transitive headless WebSocket import must also debit the row promising
headless MCP without Gorilla. Coder `693cece4` resolves this without changing the
exercise, runtime or samples. The same defect now explicitly fails
mcp-without-gorilla as well as the earlier related rows. Focused accepted GUI
baseline and restored controls score100/100, with Go package times1.029/0.665s.
Together with unchanged measured fleet100/100, the final accepted total is
explicitly composed200/200. No later full or legacy run is claimed. This remains
the same twelve-defect record, not a thirteenth mutation.

The author verified all six current grader source hashes against
post-review-source.json and both frozen Git tree identities against the accepted
and legacy bindings. The old final-source.json continues to bind the actual full
runs. Exact manuscript and sixteen fences remain the measured d2f51438 bytes.
Direct hard lint passes6527 prose words; source-relative self-review and the
independent whole-prose review identify no remaining correction. Only owned
chapter22 and this note are included in the final author commit. No Chapter23,
first-edition, frozen solution or live crossover work is changed.
