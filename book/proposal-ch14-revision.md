# Proposal: ch14 grader revision — speech log + student-provided harness

Status: PROPOSAL. Author (Bill) rules; coder implements.
Written by the coder after the ch14 grader shipped (a30194c, 96e17a3, fef6a05, tag ch14-solution).

---

## 1. The finding that triggered this

The shipped ch14 grader loads the student's JavaScript under node and calls it
directly. Measured against `internal/grade/ch14_driver.js`:

- It calls `tts.queueChunk(...)` at eight sites, `tts.flush()`, and
  `scroll.handleMessage(m)`, plus `speaking` and `cancel()` on the stub.
- Of those names, **the chapter's TL;DR promises exactly one**: `flush()`.
- The driver uses `eval(src)`. **ES modules cannot be eval'd.** A student who
  writes `export const TTS = ...` fails every check with a syntax error, even
  with byte-identical behaviour.

So the six speech checks grade *"did you correctly fix our `tts.js`"*, not
*"did you build a working speech channel."* A student who rebuilt the GUI their
own way scores near zero and learns nothing from the failure.

Mitigating fact, for the record: `tts.js` is inherited from the GUI chapters,
not written in ch14, so a student working forward from `solutions/ch13` has our
names because we gave them. And ch9, ch12 and ch13 already pin our shapes
(settings fields, `gui_snapshot`, literal `button-A: enabled` output). The node
harness is consistent with shipped practice — but that is an argument about
consistency, not about correctness.

## 2. Ruling

Bill, this session:

- Remove the node-based checks from the grader.
- The student provides a test harness that lets the grader supply files to
  their system, and a speech log the grader reads.
- The grader scripts the LLM responses, including a tool call **the student
  tells us the name of**, and grades on what the log says was spoken.
- Node remains fine for *our own* reference harness.
- Do not design around cheating (but see §9.3).

Rationale recorded from the session: today's state of the art in accessibility
testing is worse than this. A student who ends up with an awkward Chrome launch
still ends up with an automated regression test over the channel a blind user
actually receives, which almost nobody ships.

## 3. Three student deliverables

### D1 — the speech log

A file recording what the system handed to its speech channel.

- **One entry per utterance.** Not a flat text dump. If the log concatenates,
  buffering and boundary behaviour become unobservable and those checks die.
- JSON-lines, one object per line:
  ```json
  {"seq":1,"t":1183,"kind":"utterance","text":"Reading the file now."}
  {"seq":2,"t":2011,"kind":"pause"}
  {"seq":3,"t":4402,"kind":"resume"}
  ```
- `t` is milliseconds since harness start. Timestamps make the pause gate
  observable and are independently useful when debugging.
- Written at the point the text is handed to the engine, by the same code path
  that feeds it.
- Path comes from the environment (`TTS_LOG`), so the grader controls it.

**Record mode replaces the engine rather than calling it.** This must be stated
explicitly in the chapter or we set a trap. Measured this session, headless
Chrome reports `speechSynthesis` as an object with **zero voices**, and
`speak()` fails with `onerror: not-allowed`. A queue that advances on
`utterance.onend` never advances there. A student who dutifully launches real
headless Chrome gets one log entry and fails for environmental reasons.

Stating it is not a cop-out: "replace the engine with a recorder and assert on
what arrived" *is* the chapter's lesson applied to its own tests.

### D2 — the harness script

An executable the grader runs to exercise the whole system once.

- Environment in: `LLM_BASE_URL` (fake vendor), `TTS_LOG` (where to write),
  `WORKSPACE` (directory the grader has planted files in).
- Argv: the prompt to deliver.
- Behaviour: launch the system in record mode, deliver the prompt, wait for the
  turn to end, **exit cleanly**, taking down anything it launched.
- Must need no API key and no network.

The student owns all launch complexity. For our reference that is node with a
stub DOM; for a student it may be headless Chrome with a stubbed engine; for a
CLI implementation it may be neither.

### D3 — the harness manifest

The last de-pinning, from Bill's "a tool they tell us to use".

```json
{
  "harness": "scripts/tts-harness.sh",
  "read_tool": "read_file",
  "read_arg": "file_path"
}
```

The grader needs the read-tool name and argument to script a tool call into the
fake LLM. Without this the grader would hardcode `read_file`, reintroducing
exactly the pinning this revision removes.

## 4. How a scenario runs

1. Grader plants a file in `WORKSPACE` containing a marker.
2. Grader configures the fake vendor with a scripted response sequence.
3. Grader runs `D2` with a prompt, `LLM_BASE_URL` pointed at the fake.
4. Harness exits. Grader reads `TTS_LOG` and asserts.

Nothing in that loop names a method, a module, a file layout, or a module
system. It names one tool, which the student declared.

## 5. Proposed check table

Current (shipped) on the left; proposed on the right.

| check | now | proposed | notes |
|---|---|---|---|
| `tts-buffers-fragments` | 20 | 15 | streamed mid-word split must not fragment |
| `tts-filters-markup` | 20 | 20 | fence body absent, prose present |
| `tts-boundaries` | 15 | 10 | two sentences, two utterances |
| `tts-speaks-unstreamed` | 15 | 15 | fake returns a non-streamed response |
| `tts-expands-identifiers` | 10 | 10 | `camelCase`, `HTTPServer`, `RPGLit`, ALL CAPS |
| `tts-gate-both-causes` | 10 | 0 or 10 | **open question, §10.1** |
| `ch13-parity` | 10 | 5 | |
| `tts-discrimination` | — | 20 | **new; Bill's design** |
| `tts-speaks-errors` | — | 5 | defect 2 was never graded |
| total | 100 | 100 | |

### `tts-discrimination` (new, 20 pts)

The strongest check in the set, and the reason this revision is worth the cost.
The same string arrives twice in two roles with opposite expectations:

- Fake LLM turn 1: call the student's declared read tool on the planted file.
  The tool result contains `MARKER_A`.
- Fake LLM turn 2: assistant prose containing `MARKER_B`.
- Assert `MARKER_A` **absent** from the log and `MARKER_B` **present**.

It is mutation-proof by construction. Speak everything and half one fails;
speak nothing and half two fails. Only correct discrimination passes both. It
encodes Bill's ruling — *"I listen to your thinking, and that is enough"* — as
an executable fact rather than a sentence in the TL;DR.

### `tts-speaks-errors` (new, 5 pts)

Defect 2 of the six was "errors displayed, never spoken" and nothing grades it.
Script a failing tool call; assert the error text reaches the log.

## 6. What happens to the existing code

- `internal/grade/ch14_driver.js` **moves** to the reference solution as its own
  harness (for example `agent/scripts/tts-harness/`), wrapped by a
  `scripts/tts-harness.sh` satisfying D2. Its pinning of `queueChunk`,
  `flush` and `handleMessage` is entirely legitimate there: our code testing
  our code.
- `internal/grade/ch14_checks.go` and `ch14_harness.go` are rewritten around
  plant-file / script-LLM / run-harness / read-log.
- `scripts/ch14-mutations.sh` needs new patterns; the eight existing mutants
  target `tts.js` internals and should still be caught through the log.
- `solutions/ch14` re-snapshotted, tag `ch14-solution` moved.

## 7. Chapter changes required

New TL;DR rules for D1, D2, D3. The current TL;DR has 15 rules, of which
13 to 15 cover the blind persona.

**The persona check is withdrawn.** Observing the channel directly makes the
persona's blindness unnecessary for grading validity. Blindness was always a
proxy: force the observer through the channel so the channel gets tested. If
the grader reads the channel directly, the proxy is redundant. The persona
stays as a teaching deliverable and as something we run for real; the grader
stops pretending to certify it.

That also retires the point-redistribution question from earlier in the session
(identifiers 10→5, parity 10→5 to fund a 10-point persona check). Not needed.

## 8. What this buys

- A divergent implementation can pass. First check in the GUI series where that
  is true.
- Tests the real wire. The node checks feed messages the grader synthesises;
  this feeds messages the student's own agent emits. Two of the six real ch14
  defects lived in the renderer's feed sites, not the pipeline — a class the
  node harness structurally cannot reach.
- Kills the ES-module failure entirely.
- `tts.log` is independently valuable for debugging. Speech is the one channel
  you cannot scroll back through, which is why six defects survived in it.

## 9. Costs and risks

### 9.1 Determinism

Every scenario is a full system launch. Slower than node (seconds, not
milliseconds) and exposed to startup races. Mitigation: generous timeout, and
the harness owns its own readiness.

### 9.2 Single point of failure

A broken harness fails every check at once, with six confusing zeros.
Mitigation: a 0-point smoke precheck that runs the harness with a trivial
prompt and, on failure, reports *"harness produced no log"* rather than letting
each check report its own mystery.

### 9.3 Reward-hack surface (recorded, not solved)

Bill's instruction is not to design around cheating, and this section complies.
It is recorded because of the codebook thesis: if graders are the RL reward
signal, a student-supplied observation point is a reward-hackable surface, and
the HuggingFace case is the cautionary precedent.

Bill's ruling, and the reasoning is the useful part: models tend not to cheat
when doing the real work is straightforward, safe, and carries no risk of being
caught. Faking is a behaviour of cornered agents, not comfortable ones.

That turns the risk into a **design constraint** rather than an unsolved
problem: keep D1, D2 and D3 simple enough that implementing them honestly is
cheaper than faking them. A recorder that replaces the speech engine is a few
lines. A fabricated log that passes every scenario means reimplementing the
whole filter chain in the logger — strictly more work, for a worse outcome. As
long as that inequality holds, the surface stays closed by economics rather
than by policing. If a future revision makes the harness contract onerous, the
inequality flips and this section becomes live again.

Bounded exposure meanwhile: the asserted content originates in *our* scripted
responses, so producing a correct log still requires implementing the real
filtering logic. The reachable cheat reduces to "implemented the filter inside
the logger" — at which point the filter exists but may not be wired to the
speaker. `ch13-parity` independently prevents deleting the GUI. Closing the
last gap would require reading student code, which no check in this book does.

### 9.4 Student burden

Making a GUI launchable headlessly with a stubbed engine is real work, and some
students will fight Chrome. Accepted by ruling, on the grounds that the
alternative on offer today is nothing.

## 10. Open questions for the author

1. **Pause gate.** It needs a typing event mid-turn, which a one-prompt harness
   contract cannot express. Options: (a) extend D2 with an optional second
   argument, a string the harness types into its own input while the turn
   streams, asserting `pause` then `resume` entries in the log; (b) drop the
   check and redistribute 10 points. (a) is more faithful and more burden.
2. **Manifest location and name.** `ch14-harness.json` at repo root, or inside
   the student's own tree?
3. **Scenario isolation.** One launch per scenario is clearer to debug; one
   launch with several prompts is several times faster. Preference?
4. **Does `tts.log` ship as a product feature** (a flag on the real agent) as
   well as a test-mode artifact? Bill said he wants it for debugging
   regardless, which argues for a real flag.
