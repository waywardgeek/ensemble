# Original 22 / Edition 2 ch20: independent review

Reviewer `reviewer_astra` (GPT-6 Astra/high), October 11, 2026.
**Accepted with three narrow grader exceptions at initial formal review.
Zero correction rounds used.**
Read the governing documents and mandatory
`.agents/skills/ensemble-coding/SKILL.md`. The preparation below preceded implementation and old-answer access. After the
student froze its implementation, the reviewer read the matching first-edition
solution and complete student review. No reviewer grader or provider call was made.

## Placement and source coverage

Independently agree with the coordinator's mapping, authorized by Bill and
committed in `6ec35e2`: original21+22 becomes Edition2ch20, original23 becomes
ch21, and original24+25 becomes ch22. Status/accounting completes introspection
before the sandbox; public frontend reuse belongs beside application/fleet
composition. Review original24 before implementing25, retaining every capability
and original-numbered result. Fleet GUI selection remains deferred to26; inspect
its completion at the25 review boundary. This is placement, not authorship or
acceptance of any new implementation.

Read every line, including examples, of these current chapters:

| Original | Lines | File revision | SHA-256 |
| --- | ---: | --- | --- |
| 22 | 586 | `7855ab179562b94c353efda07b53302528ca7c92` | `2589a196ad87504ab7193475a3ffba0e5961d57dd5a35ed4cc4e9a2888208118` |
| 23 | 808 | `e509c2ddc237f799844a6efe6ebdb744289ce295` | `3b19960c74cf4d01e225b9c60bc175b199cc95ec01cad754b97d9a3dcf4bd061` |
| 24 | 565 | `2ca63636c8ba592ea0ed190fb0a3c4731e2b0848` | `c6ac58046ed8e5496c44931e20d126d36081325fc06bfc27f4276c73424976b1` |
| 25 | 582 | `009d74c70b108681b2107b4c3a08685508567cee` | `eb09ec25fe8a681479310ffe22f4cf19d0b4eba34bf507dbbfd8181f7302616f` |

Read the existing24/25 notes and subsequently created student-safe22 notes.
The25 notes still correctly identify unresolved `tool_limits` wording and the
timer/message-wake contradiction. Those are later contract questions, not22
requirements. Do not recreate flat roots, callback bags or coupled GUI modules
to imitate the historical repair narratives.

## Original22 grader preflight

Source inspection confirms the narrow launch/report erratum recorded in
[the student-safe notes](ch22.md). The behavioral harness
`internal/grade/ch22_harness.go` is at
`6f6279367b4e75f47e8eb24568c928fa80dcdc5f`, SHA-256
`df7884a797d8056098473b2de137d7ec295dc58fefd6a5d945834269a3f68dd0`.
It launches `--port PORT` in a fresh cwd with the five documented LLM/skill
environment fields; it sends JSONL `kind:prompt`, awaits nonempty JSONL
`assistant`, and switches models through `/ws` subscribe/update_settings.
No GUI asset override is supplied. Its prompt scanner has no local deadline:
the coordinator must bound the unchanged original run externally. The known
human-terminal envelope mismatch therefore remains relevant, without any
automatic waiver of22 behavior or prediction of a score.

The shared builder `internal/grade/harness.go`, revision
`c24b8e84da112ef66f729c0cdc60413ee6fb547d`, SHA-256
`a54ae4976b4a58dd7a2cd0c4d2fe86f8abda9ee990874669759192bd5d144eff`,
prefers `./cmd/` when that directory exists, otherwise the module root; it has
no optional-GUI entry override. Source checks still require a source tree.

`internal/grade/ch22_status.go`, at the same revision as the22 harness, SHA-256
`c3b4ebc4a78ad58cfe75ef350b0586ff7f95f34cc292ad2e3f2bfcc25c4f12af`,
requires the text labels/order published in the erratum, including lowercase
`model:` and `session` for discovery. It observes the actual returned tool text
in a subsequent provider request. Expected cost uses the models observed on
the wire, avoiding an assumed timing boundary for asynchronous model changes.

The run invokes original21 parity before the new status session, so inherited
failures/timeouts must remain distinguishable from22 findings. Valid status,
per-model accounting, ownership, composition and observer-lifetime obligations
remain binding. Formal disposition awaits frozen code, the student review and
actual original/live/mutation evidence; no new exception is accepted here.


## Initial formal disposition

Accept final submission `1da0891a7ea0f121b05225c62be3b2190e4f3918`, authoritative
main tree `a801628b1dd631e5ea20ad03ea7d255584728d31` (152 tracked files), with the
three exceptions below. Read the complete [student review](ch22.md), all source
changes against accepted original21 `f35b6ae`, the actual owner/composition and
meter paths, original22's relevant implementation/tests, and its changes against
original21. No outstanding implementation, scope or KISS finding remains.
This completes original21+22 / Edition2ch20; export may now refresh combined
ch20, preserving the old21-only tag. Original23 is next under Bill's new map.

The unchanged original grader remains **20/100 FAIL**, make exit2/grader exit1,
51.3 seconds on `18c2d48`, without external timeout. Its build and import checks
awarded20 points. This is not a passing grade or successful behavioral session.
Production bytes in the final submission equal the graded/live source; the only
later source edits strengthen the status-result assertion and add the real CLI
model-control regression. No redundant paid run or original-grade rerun was needed.

### Accepted explanations and exact exceptions

1. **Back-pointer detector false negative.** The scanner chooses a concrete
   struct containing interfaces declared in its package. Actual dispatch uses
   the common `ToolContext` interface and private `agent.callContext`, so the
   detector instead selects `common.Config` and treats `CredentialProvider` as
   a parent. That is a transport service, not Config's parent. The actual call
   context reaches Engine through its required Agent; Engine exposes Agent,
   Agent exposes Ensemble, and the root constructs/retains those relationships.
   The auth store itself has its Ensemble parent/accessor. I accept the student's
   explanation and waive this detector shape, not actual-parent architecture.
2. **Composition detector false negative.** The AST helper skips method calls
   on values and its function lookup excludes methods. Thus it misses
   `app.NewAgent`/`root.NewAgent`. I traced the delivered headless main, GUI main
   and virtual-user main into public `Ensemble.NewAgent`, which actually calls
   `agent.New(e, cfg)`, `engine.New(a)` and the other owned constructors. These
   executables do not assemble an independent flat runtime. Accept the student's
   explanation; waive the call-graph heuristic, not the requirement that every
   delivered binary use the composition root.
3. **Headless/GUI launch mismatch.** The supplied authoritative tree has no
   `cmd/` directory, so the original builder builds its headless root executable.
   It rejects the grader's `--port`; the three visible usage errors and unopened
   port explain status, cost and inherited21 behavioral failures. The session
   never reached those checks. This is distinct from21's terminal-envelope
   timeout. Bill requires a separate optional GUI module, and submitting only
   that module would hide core from this structural grader. Accept this narrow
   launch exception, backed by actual GUI/CLI and tool-loop evidence below.
   No wrapper, core GUI dependency or modified original grader was introduced.

All valid masked obligations remain required. Source confirms the tool package
imports common rather than Agent/Engine spokes; its awarded15 points alone would
not prove that, since the detector picked Config. The status tool uses the real
Engine meter with no copied counters or status callback. Original21 HTTP MCP,
selected tool gating, error and cancellation code are unchanged, retain their
accepted live/mutation evidence, and remain in the passing broad test suite.
No new web run was needed merely because this harness never launched its session.

Accept same-provider, between-turn model selection as the minimal missing
control for the exercise. The actor rejects an active turn, clears stale resolved
identity, and changes the Agent's own configuration without resetting usage or
conversation. Request accounting retains the model captured for each request.
Both human clients use this operation; the GUI broadcasts effective state and
rejects mixed model/preference patches before mutation. Session-only selection
preserves launch configuration on restart. This does not promise arbitrary model
availability, provider migration, new billing policy or automatic discovery.

The student's observer-lifetime explanation is accurate: the data queue stays
open, a separate completion signal stops the writer, teardown joins it, and
client-map changes/broadcast use the same mutex. The old send-to-closed-channel
pattern is absent. Inherited source/tests suffice here; no new deterministically
scheduled teardown mutation or exhaustive race proof is claimed.

### Independent runtime and evidence checks

Read sanitized raw API records, not only model narration or summaries. Verified
four request/completed-response pairs: Astra, Astra, GPT-6.1 Sol, GPT-6.1 Sol,
all high effort. The first status result matches its preceding response's
740 input/14 output; the second matches977 input/14 output and session2553
input/122 output, including earlier Astra requests. Both actual tool results
appear in subsequent provider requests. All cache buckets were zero. Final
answer usage arrives after each tool snapshot; the final3628/251 total is
consistent. The visible browser trace shows prompt, Settings selection and second
prompt. Browser and GUI exited0 in16.34 seconds. Verified the GUI binary hash.
The preserved first browser attempt has an empty API log and failed before
inference because the bundled browser was absent; using installed Chrome fixed
launch. See [reviewer receipt](../../solutions/evidence/ch22/reviewer-live-check.json).

Live cost is truthfully unknown. Known-price tests independently calculate
per-model token-price estimates, including cache buckets, and prove that selecting
another model does not reprice historical work. These are table-derived usage
estimates, not evidence of actual ChatGPT subscription charges. The live run
provides no numeric-dollar proof, and no extra paid probe was warranted merely
to manufacture one. Endpoint is source/configuration-derived, not an independently
captured URL. The second model performed only the requested status demonstration;
it does not replace Astra/high for coding. No reviewer credential/profile access
or provider call occurred.

Student core/GUI test+vet results pass; final core checks were repeated after
test changes. Independently ran focused Agent/status/model tests with `-race`
and GUI model selection with `-race` from an immutable18c2d48 archive. After the
final freeze, reran focused core race tests (exit0,1.096s) and the built CLI
model-selection test (exit0,0.277s) from an immutable1da0891 archive. The latter
observes the actual executable's provider request. No mocks replace the runtime;
fakes are at external HTTP boundaries, and GUI tests use real WebSockets.

Reviewed all14 targeted mutation outputs and applied substitutions. They fail
intended behavioral assertions, not compilation. The GUI mixed-patch case times
out waiting for its required error event with a working baseline; this is a
specific missing-response check, not general timing proof. The historically named
`status-dispatch` mutation actually corrupts the report's leading model label;
its claim is report-label sensitivity, while declaration and real tool-loop tests
establish exposure/dispatch. Its initial substring-assertion survivor and the
first malformed repair's failed unmodified baseline remain preserved. The final
text-block decoder checks the prefix, and the same mutant now fails correctly.
No survivor was silently relabeled. Pre-submission test strengthening uses zero
formal correction rounds.

### Comparative KISS review

Counts below compare each edition with its own predecessor, using physical,
active, comment-only and blank lines. Inline-comment code remains active;
classification is lexical, not a complexity metric. Go, JS and HTML are separate;
examples, generated data, documentation and evidence/support files are excluded
from production and tests. No moved source files occur in these deltas. Full
per-file counts are in [reviewer counts](../../solutions/evidence/ch22/reviewer-counts.json).

| Delta | Physical | Active | Comment-only | Blank |
| --- | ---: | ---: | ---: | ---: |
| New edition production Go | +118 | +95 | +15 | +8 |
| New edition production JS | +3 | +3 | 0 | 0 |
| New edition production HTML | +1 | +1 | 0 | 0 |
| New edition Go tests | +314 | +283 | +23 | +8 |
| First edition production Go, ch21→ch22 | +432 | +174 | +222 | +36 |
| First edition Go tests, ch21→ch22 | +544 | +375 | +113 | +56 |

The smaller production delta is explained by ownership and per-model accounting
already existing, not by omitting the current requirements. First edition moves
usage from Agent into Engine, repairs command wiring and WebSocket teardown;
its status formatter also prints per-model rows. The student adds a28-line view
of the existing meter and46-line actor selection, plus small controls, instead
of repeating that historical repair. Existing public composition and optional
GUI separation remain simpler than the reference's residual sibling wiring.
There is no new status framework, speculative provider switching or future
sandbox/fleet feature. Tests add more lines than production for meaningful
external-boundary observations, not a new testing framework.

First-edition explanations of historical ownership and channel-lifetime mistakes
are useful; several describe an architecture different from Bill's required
Agent-owned configuration/Ensemble root, and its Call comment promises one root
route while Call still embeds Agent beside Engine. The student's comments
accurately explain view-versus-owner, turn-boundary selection, stale identity and
why model preference is not persisted. New exported methods/interface members
and Settings.Model have useful contract comments; new test fixture fields are
also explained. Net production-Go additions are about13.6% comment-only versus
nonblank lines, below the rough20% teaching target, but the small wrappers reuse
well-documented ownership rather than deleting explanations or compressing code.
No padding or arbitrary quota is warranted.

The reference status tests replace Agent/Engine with handwritten stand-ins and
exercise formatting; the student runs real composition, Engine, tools, history,
HTTP rendering and usage through external fakes, plus real GUI/CLI controls.
Its expected cache ratio includes cache writes in total input; the reference's
status formatter omits them despite its separate usage helper including them.
Exact priced totals and historical repricing mutations improve confidence over
reusing pricing helpers as the expected result. The reference's broad teardown
stress test protects a real old bug, but does not require importing that bug or
copying a timing-heavy stress suite into an unchanged safe queue lifecycle.

The student explanations are accepted with these explicit limits. Original
source/grader results remain history; future author/grader work should expose
launch contracts and inspect actual owner relationships and method composition.
No further correction or feature work is required for original22 acceptance.
Mandatory `.agents/skills/ensemble-coding/SKILL.md` applies to the next handoff.
