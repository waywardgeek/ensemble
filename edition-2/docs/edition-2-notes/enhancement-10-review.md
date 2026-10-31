# Edition 2 Chapter 10 independent enhancement review

Reviewer: `reviewer_astra`, October 11, 2026. Mandatory skill:
`.agents/skills/ensemble-coding/SKILL.md`.

**Accepted at initial combined prose and grader review, zero formal correction
rounds.** This covers Edition 2 Chapter 10, original Chapter 11, Persistence.
The coordinator owns checkpointing and release; this review does not advance
Chapter 11 or change historical implementation results.

## Reading and teaching

After compaction, reloaded all four complete documents into active context:
workflow (1,804 lines), carryover (871), voice (632) and chapter procedure (599).
Also read the entire skill, AGENTS, enhancement plan and map. Full original
Chapter 11, its complete student/coordinator notes and comparative review,
accepted persistence/client source and original grader were read for this gate.
Reviewed the complete enhanced chapter and author record, including final
reconciliation `f245a228`/`373ce62b`; manuscript SHA-256 is
`32723da4cf2531b5786330502708d58c62db613fd71a7a73acd855c2e83b6b22`.

The original goldfish opening, Bill's default-load request, anchor dilemma,
seven-section progression and independent snapshot/log demonstration survive.
The returning reader's afternoon of work remains at stake through the refusal
section. The additions explain real failures rather than replacing the chapter
with private workflow: correct recall can coexist with false prompt metadata;
a normal save cannot expose a missing tail replay; an error-returning process
can still overwrite its rejected input. The prose gives those mechanisms enough
room without introducing an invented incident or repetitive personality filler.

The chapter distinguishes command-selected cwd/save.json from the library's
Agent-owned path, effective instructions from an empty raw field, and a complete
log from a trimmed log's descendants. Snapshot and anchor share the existing
owner lock; restoration stages state; shutdown joins jobs before saving. Shared
vocabulary stays in common and the fold excerpt belongs in history. Current
configuration wins without reinstalling handlers or loaded skill state. The
four-field exercise remains small: no canonical disk representation, journal,
integrity framework or future memory feature has been added.

Direct hard prose lint passes at 2,945 words. Soft lower-length, six negation
forms and the long person gap remain reasonable human-review exceptions to
warnings, not failed gates. The concrete lost-conversation problem carries the
mechanism; extra biography or padding would not improve it. Every local chapter
link resolves. The original premature self-driving-GUI claim is accurately
changed to streaming into its GUI.

The retained live receipts were independently inspected: eight successful
response processes across Claude/OpenAI/Gemini, including Claude's independent
snapshot-only and log-only variants, then a ninth damaged-save refusal with no
request and unchanged bytes. These are historical chat-through-pipes runs, not
a new paid demonstration or a personal Bill session. The later effective-prompt
fix has local HTTP-boundary regression/mutation evidence; the chapter correctly
says no new **paid** run was needed. Original full grading remains interrupted
and unscored; focused70/85 retains its undefined-think early return, which did
not establish the masked anchor/sequence predicates.

## Grader, fixtures and compatibility

Read the full coder submission at `91b96b1a`, all new grader/test code, command
delta, preserved original predicates and matching legacy persistence/model
source. All four final source hashes match `final-validation.json`. The exact
two Go fence hashes also match the manuscript and published-sample receipt.
Those excerpts compile against real vocabulary and the actual validator/reducer;
the restore excerpt's surrounding validation is explicitly outside its scope.
Four actual CLI chat processes verify the published default and override commands
against a small HTTP fake. This is deterministic sample validation, not live
model evidence; GUI input/EOF has separate actual-command coverage.

The six original persistence weights remain 15/20/20/15/10/5, plus15 for corrected
Chapter9 parity. Both real core and GUI applications execute startup, actor,
reducer, filesystem and shutdown. Private saved context/event layouts remain
opaque except for the published sequence identity. The checks retain actual
request-byte comparison, old-snapshot/new-tail and snapshot-only cases, and
strengthen exact next identity, loaded-log retention, wire-effective metadata,
path symmetry, configuration precedence and refusal before HTTP without an
overwrite. Undefined think is removed rather than implemented for the fixture.
No source-shape detector, callback framework or new runtime protocol appears.

I accept the buffering and model-fixture corrections. Two already-buffered
prompts can become prompt plus hint in the accepted actor; one admitted turn
per process makes restart the actual boundary. Legacy model.go explicitly
declares `gpt-5-course`; earlier current-model and gpt-5 launch failures never
reached configuration restoration and remain setup failures. The corrected probe
now observes current vendor/model/prompt/tools and preserved conversation.

I also accept the narrow parity launcher. Original Ch2Run creates one directory
then runs independent vendor cases inside it. Its observed doubled/tripled
histories match newly correct automatic restoration. Chapter10's six direct
persistence checks run on the unwrapped core/GUI before constructing the wrapper.
Only inherited no-argument processes get fresh explicit save paths; cwd, fixture
files, executable, input and environment remain real. Offline arguments pass
through unchanged. This fixes the old fixture's session lifetime without hiding
restoration from its own checks or changing an original grader.

An independent actual-command control verifies both sides: direct restart
retains the earlier prompt and cwd save, while wrapped sessions omit that prior
prompt and create no cwd default save. Actual dump/render bytes remain identical
with or without the wrapper, and cleanup removes its owned directory. Baseline
passes in0.450s. Disabling only isolation fails the intended second-session
assertion in0.332s, after the direct control passes; restoration passes in0.348s.
The source, patch and full outputs are bound in the
[reviewer receipt](../../solutions/evidence/ch11/enhancement10/reviewer-validation.json).
Scratch copies were removed; no shared grader or frozen solution was edited.

Final actual accepted grading is **100/100 PASS**. Final legacy persistence is
**80/85**, with **80/100 FAIL explicitly composed**, retaining already observed
failing Skills parity rather than claiming a new full run. JSON null is a real
legacy failure: unmarshalling into a struct value accepts it, and both actual
clients then send one HTTP request and overwrite the file. Truncated JSON refusal
passes. A valid snapshot-only save does not justify accepting no SaveFile at all.
The earlier85/100 accepted and60/100 legacy full runs remain recorded separately;
no setup failure is counted as a final runtime defect. A reviewer-requested
receipt-only correction at `443fe6e6` names initial-tools, progressive-disclosure,
depends-autoload and var-substitution accurately; load-skill itself passed. Scores
and evidence are unchanged.

## Sensitivity, comparison and disposition

Inspected all28 patches/observations:24 distinct compiling defects and four
repeated current-config observations under the final model fixture. Every raw
result array matches its receipt and includes the intended failed check; output
outside those arrays contains no compilation error, panic, race or cleanup
failure. Broad missing-field/save defects propagate into dependent setup; those
are not independent additional kills. Bridge exit0 means result collection,
not passing mutated behavior. No survivor or invalid failure is relabeled.
The separate isolation mutation above checks the new fixture correction itself.

Affected grader/command tests, vet and formatting pass. The actual accepted
source race tests establish rich neutral replay through all three renderers,
owner-directory isolation, staged bad-save refusal and intentionally empty
primary instructions. Those obligations retain their original source mutations
and review; the numeric grader does not claim comprehensive malformed-input,
concurrency or durability proof. No unchanged broad suite or paid run was
repeated merely for freshness.

Independent counts match: new grader **467 active /56 comment-only /543 physical**
Go lines; integration test **22/2/27**. Preserved original harness/checks total
**693/121/890**. Count nonblank, non-`//`-only lines as active, leaving inline
comments active; no block comments alter these counts. The command only adds
Chapter10 dispatch/default/range support. These are separate grader additions,
not runtime deletions. The roughly10.7% comment share is below the teaching
target, but useful explanations cover session lifetime, opaque splicing,
observation boundaries and exact sequence identity. The sole new export has a
useful contract comment. A direct sequential fixture reusing existing helpers
is simpler than a second runtime or a generic persistence-testing framework.

The accepted runtime comparison remains [the original review](ch11-code-review.md):
against each own predecessor, final student production grows184 active/41
comment-only/235 physical lines, versus legacy152/88/254; student tests grow
304/21/334, while legacy adds no chapter-specific tests. The student places
behavior with its owners and verifies actual failure/replay paths; the legacy
has useful anchor/replacement explanations but persistence behavior in common.
This author phase changes neither implementation, frozen tests, original text
nor original grader. Counts alone do not establish quality.

No blocking prose, coverage, architecture, scope or simplicity finding remains.
The enhanced chapter and separate grader agree with the accepted solution and
preserve the actual failures and evidence limits. Chapter10 is accepted at its
initial combined review.
