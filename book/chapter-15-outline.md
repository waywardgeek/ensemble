# Chapter 15 outline: context management (Chapter A of the context arc)

Source of truth: `book/ch15-context-engineering-design.md`, Parts I and II. This
outline turns that design into chapter structure. Where they disagree, the design
doc wins until Bill rules; flag the conflict, never resolve it silently.

Numbering RULED: A = ch15, B = ch16 (memory), C = ch17 (goal stack). Evening
rulings of 2026-09-22 are at the end of the design doc and override it.

Workflow for this chapter (Bill's, supersedes the outline-first order in
`chapter-writing-procedure.md`): design doc → author writes TL;DR + §15.1 plain
words → coder builds `solutions/ch15` and writes `book/review-ch15.md` → author
revises TL;DR + plain words against the review → author writes the body.

---

## Title candidates

1. **Keep the Words** (short; the tagline's first half; names the practice)
2. **Why Is This Byte Here?** (the chapter's one question, verbatim)
3. **The Working Set** (names the deliverable; weaker per the ch5 rule "name the
   activity, not the deliverable")

Recommendation: 1. Tagline in the chapter: "Keep the words. Let the bytes go."

## Through-line stake (voice.md §4)

A long session ends one of two ways today: the window fills, or a note is handed
to a stranger. Every byte in the window must answer three questions, *why is it
here, who put it here, and what removes it*, and an agent whose bytes cannot
answer them forgets what it is doing before it forgets anything it could afford
to lose.

## Per-section wild fact (voice.md §7.3)

Every fact below is on file; cite as shown. Nothing here may be replaced by a
figure reconstructed from memory.

| § | Wild fact | Source |
|---|---|---|
| 15.2 | Tool results ≈42% and tool-call arguments ≈30% of history: ~72% of the window is bytes the actor could regain | `chapter-02.md:603-611` (measured) |
| 15.2 | Chapter 2 built four redaction levels, and no production code has emitted one since | design doc §A.12a (VERIFIED) |
| 15.3 | Per-round-trip redaction was predicted to wreck the cache; measured 77% cumulative hit rate after a >100K-token cold miss | design doc §A.11.1 |
| 15.5 | The skill's instructions ride inside the `load_skill` tool result, so the first tool clear deletes the manual and keeps the tools | `agent/internal/tools/tools.go:1016` |
| 15.9 | `compress_context` under one model deleted ~80% of the conversation, starting at message 1, where the goal was stated | Bill, 2026-09-13, quoted in design doc §I.8 |
| 15.10 | The vendor's server-side context editing tells clients "You do not need to sync your client state with the edited version" | design doc §A.8 (quote verified from `cr/docs/provider-context-management.md`) |
| 15.10 | Chapter 11's `Rebuild` refuses the whole past on the first bad event | `agent/internal/common/save.go:64-68` |
| 15.11 | A stub citing `cr/io/26` names a different command's output after a restart | design doc §A.11.2 |
| 15.13 | (moved) The 0.5→3.0 recall-threshold fact belongs to the auto-recall chapter (ch16 late ruling 1); §15.13 needs a replacement fact or none | design doc §A.10 |

## Voice plan

- **Register:** third person throughout the body (voice.md v4). Short
  first-person motivational opener allowed (Bill's ruling: openers, preface,
  LinkedIn only).
- **Bill moments (≤2):** (1) the `compress_context` war story, quoted verbatim
  from §I.8; (2) the opener, see story slot.
- **[STORY SLOT — opener] (RULED substance, wording from Bill):** how most
  users manage context today. They use a chat until it degrades from poor
  context management, then paste the context into a new chat and carry on. Bill
  still does this, several times on 2026-09-22, with Opus 5.5. That manual reset
  is the baseline every mechanism in this chapter competes with. The author gets
  Bill's own wording before printing; nothing here is a quote.
- **Risk to name, not hide (UNMEASURED):** Bill's hypothesis for why Opus 5.5
  degraded is that per-round-trip auto-redaction interferes with its thinking,
  and this chapter builds that exact mechanism. Print it as an open risk with
  the measurement that would settle it (design doc Part VI), not as a finding.
- **Confession (on file):** the prediction that per-round-trip redaction would
  wreck the prompt cache was wrong (77%, §A.11.1). Print it in §15.3 as a failed
  prediction, with the number.
- **Motivational opener:** the stake, then the opener story, then TL;DR. No
  forward preview of Chapter B in the prose (procedure rule); the memory region
  appears in the figure, empty.

---

## TL;DR spec (draft; this is the grader contract)

### What you are building

- Every context entry gets a **kind**. Ordinary conversation is `Dialogue`;
  anything that must outlive tool clearing is its own kind and never carries a
  tool part.
- A **`micro_handoff`** tool whose event removes every tool call and result from
  the context and appends one `Handoff` entry.
- A **tool-bytes ladder**: a policy that emits Chapter 2's `RedactData` events
  below two watermarks, each event recording the exact Seq it cuts at.
- **Per-round-trip auto-redaction** with a **`keep_tool_results`** tool, gated
  per model: the actor-curated band of the layout (Variant B).
- **Skills that survive**: the reducer turns `SkillLoaded` into a `Skill` entry;
  the `load_skill` result becomes an acknowledgement.
- A **frozen prefix**: system prompt and fixed tool declarations are
  byte-identical on every request; mid-session tools arrive as a `ToolsChanged`
  entry.
- A **total reducer** and **crash-safe persistence**: append-on-write log,
  one backup generation, bounded truncation, on top of ch11's anchored
  snapshot.
- A **Context Management** settings tab.

### Contracts (additive only; shapes DERIVED, coder confirms)

```go
// EntryKind says why an entry is in the context and which verb removes it.
// The channel (dialogue, instruction, data) is derived from it, never stored.
type EntryKind int

const (
	KindDialogue EntryKind = iota + 1 // prompts, hints, model output, tool use, tool results
	KindHandoff                       // appended by MicroHandoff; removed by save_memory (next chapter)
	KindSkill                         // appended from SkillLoaded; removed by unload_skill
)

type Entry struct {
	Seq   Seq       `json:"seq"`
	Actor Actor     `json:"actor"`
	Kind  EntryKind `json:"kind"`
	Parts PartList  `json:"parts"`
}

// Appended to the existing EventType block after SkillLoaded (event.go:40).
// A number once assigned is never reused.
const (
	MicroHandoff EventType = iota + SkillLoaded + 1
	ToolsChanged
)

type MicroHandoffData struct {
	Text string `json:"text"`
}
```

Existing and reused unchanged: `RedactData{From, To, Level, Replacement, Reason}`
(`event.go:209-215`) with levels `RedactResult` and `RedactTool`;
`SkillData{Name, Body}` (`event.go:201-204`), which **already carries the body**.

Coder question: `ToolsChangedData` carries a declaration delta (added
declarations, removed names); its field types follow whatever Chapter 10's tool
declaration type is. The TL;DR prints the final shape after the review.

### Rules (numbered; each maps to a check)

1. **Frozen prefix.** The system prompt and fixed tool declarations are
   byte-identical across every request of a session. Only a full refresh changes
   them.
2. **Tools through the dialog.** A mid-session skill load or MCP connect emits
   `ToolsChanged`; on the Anthropic API its declarations ride in a dialog entry.
   Vendors without dialog-carried tools re-declare and pay the cache miss.
3. **Survivors carry no tool parts.** A `Handoff` or `Skill` entry never contains
   a `ToolCallPart` or `ToolResultPart`, so tool clearing cannot touch it.
4. **Skills are entries.** The reducer turns `SkillLoaded` into a `Skill` entry
   holding the body. The `load_skill` tool result is an acknowledgement only.
5. **`micro_handoff` is three records:** the tool call, an ordinary result, then
   a `MicroHandoff` event. Its reducer removes every tool call and result part
   (dropping entries left empty), then appends one `Handoff` entry. The text
   appears in the next request exactly once, and pairing is never broken.
6. **The ladder records a Seq.** Below the dialogue watermark: `RedactTool`.
   Between watermarks: `RedactResult`. Each event records `To` as a number, never
   "the watermark". The watermark moves in steps: when a band passes about twice
   its budget, one event cuts it back to one budget.
7. **The reducer is total.** A malformed event, or a redaction naming an entry
   already removed, is skipped with an observable diagnostic. `Rebuild` never
   returns an error for log content.
8. **Persistence.** Events are appended to disk as they happen, so a crash
   leaves a tail after ch11's `as_of` anchor. Recovery is ch11's load (snapshot
   plus tail), now fed by the on-disk tail. Snapshot before truncate; never
   truncate past the anchor; keep one backup of the previous snapshot. Normal
   shutdown snapshots.
9. **Settings.** The watermarks, the hysteresis factor and log retention N live
   in a Context Management tab.

### Exercise

| Check | Pts | Rule |
|---|---|---|
| skill survives the ladder | 20 | 3, 4, 6 |
| micro_handoff shape | 15 | 5 |
| ladder is recorded, not recomputed | 15 | 6 |
| frozen prefix unchanged | 10 | 1, 2 |
| replay equals snapshot | 15 | 7, 8 |
| crash recovery | 15 | 8 |
| total reducer | 10 | 7 |

Behavior only: the requests the fake vendor receives and the files on disk.
Rule 9 is ungraded (the client is not graded; ch9 precedent).

---

## Body sections

Each section has one thesis sentence. Word budget is a symptom detector only
(4,000–7,500 soft); never cut to hit it.

### §15.0 Opener (first person allowed, short)
Thesis: today's context management is a human pasting a degraded chat into a
fresh one, and this chapter makes the agent do that job continuously and
deliberately instead. Story slot above. Ends on the stake.

### TL;DR (above)

### §15.1 The idea in plain words
Thesis: the context window is not storage; it is a working set rendered from an
append-only log, and each byte in it must say why it is there, who put it there,
and what removes it. Plain-English "why" section (voice.md §9). No types.

### §15.2 Why is this byte here?
Thesis: most of the machinery is already built, and what is missing is the policy
that decides where it applies. Bill's definition (managing all the bytes, as well
as today's LLMs allow). The already-built table from §I.3, cut to what the reader
remembers: four redaction levels (ch2), compaction-as-event (ch2), constitution
prompt and lazy unload (ch10), Save/Rebuild (ch11). Wild facts: ~72%, and levels
nothing has emitted since.

### §15.3 Two laws
Thesis: every byte belongs to the frozen prefix or the append-only tail, and the
cost of changing one is proportional to its distance from the tail. Law 1 with
ch10 as precedent; the full refresh as the one explicit, accepted cache miss.
Law 2 as a cost model, with the confession: the cache prediction that failed
(77%).

### §15.4 Three channels
Thesis: every entry is dialogue, instruction or data, set at write time and
never inferred, and the channel is what makes data safe to carry. Dialogue /
instruction / data table. Placement is a renderer decision per vendor (ch2's
thesis again). Tool declarations follow Law 1: they change only through a dialog
entry. State plainly that "data" is a boundary in the data structure and a
labeling convention on the wire; the wire has no data role.

### §15.5 What survives is never a tool call
Thesis: a tool call is the door and the entry is the payload, so anything that
must survive tool clearing is its own kind. The defect first (wild fact:
`tools.go:1016`), then the kinds table (Dialogue, Handoff, Skill; the memory
kinds shown greyed as "later"), the naming rule (a survivor kind is named after
the tool that creates it), and the acknowledgement rule. The reader should see
that survivors are safe by construction, not by care.

### §15.6 The layout
Thesis: order the tail from least volatile to most volatile, and the order chosen
for meaning turns out to be the order the cache wants. The figure from §I.6 with
the memory region empty. Between reorders, new entries append at the tail in Seq
order. Two variants by model capability (A: two positional bands; B: one band the
actor curates), selected by a column in the model features table (ch6/ch7
precedent, no default row).

### §15.7 Visible reasoning is the storage format of the self
Thesis: identity survives in what the actor said out loud, not in the tool
results, so narration before action is not a UI nicety. Two payoffs of one habit:
the human's real-time steering channel (ch6 hints) and the residue that survives
`RedactTool`. Chapter 2 mentions this once (`chapter-02.md:615`); this chapter
teaches it as practice.

### §15.8 The ladder
Thesis: the conversation region is Chapter 2's redaction levels with watermarks,
and each event records a number rather than a policy. Band table. The two events
are existing `RedactData`. Why a Seq and not "the watermark" (determinism under
replay). Why steps, not a trickle (Law 2 applied to the watermark). The
unmeasured 80% yield stays out of the prose; the measured 72% is the figure.

### §15.9 From compress_context to micro_handoff
Thesis: each successor gave the model less discretion over what cannot be
recovered. The succession from §I.8: `compress_context` (Bill quote, by position,
from message 1), `handoff_task` (a note to a stranger; deleted, with the
two-write-paths bug as the reason), `micro_handoff`. The three-record sequence and
why pairing never breaks. Why capability-gated curation is not `compress_context`
again: discretion only over recoverable bytes, and gated per model.

### §15.10 Compaction is described, never performed
Thesis: the context is whatever replaying the log produces, so compaction is an
event, and the reducer must accept any log. Replacement mapping; LLM output lives
in the event payload, so replay never reruns a model. Totality: a bad event costs
one entry, never the whole past (wild fact: `save.go:64-68`). Sidebar-sized: the
vendor's context editing, compatible but declined as primary on scope, policy
shape and determinism; kept as an observable backstop.

### §15.11 Keep the words, let the bytes go
Thesis: discarding tool bytes is safe because the actor can always regain
context, not because the bytes are kept. The workspace, the tools, the user and
the actor's memory are the backing store. A conclusion drawn from bytes that will
not come back survives only if it was said out loud (ties to §15.7). One rule
survives the loss: a stub may say the bytes are gone, but never cite an address
that could later name different bytes (wild fact: `cr/io/26`; ch14's rule:
wrong data, not no data).

### §15.12 Crash-safe persistence
Thesis: a log is only ground truth from the moment it reaches the disk. Ch11
already anchors the snapshot and loads snapshot plus tail; its gap is that the
tail lives only in memory until a clean exit. Append-on-write, save procedure
with one backup generation. The two
invariants: snapshot before truncate (not commutative), never truncate past the
anchor (a display preference never overrules a correctness boundary). Truncation
loses old tool results, and that is accepted.

### §15.13 The Context Management tab
Thesis: numbers nobody can see can only be guessed at. Tabs claim a subsystem
with one owner. Contents: watermarks, hysteresis, retention N, per-model
curation capability (display). Last-request cache hit rate next to cumulative,
because cumulative is dominated by the cold start. Wild fact: 0.5 → 3.0 by feel.

### §15.14 Exercise, graded
Checks table (above), each check's property in one sentence and why it cannot
pass by accident (from design doc §A.13).

---

## Coder list (for Opus 5)

Brief goes to `book/brief-ch15-coder.md`, written by the author after Bill's
LGTM on this outline. Starts from `solutions/ch14`.

1. `Entry.Kind` with `KindDialogue`, `KindHandoff`, `KindSkill`; every existing
   entry-producing path sets `KindDialogue`; zero value rejected.
2. Reducer: `SkillLoaded` → `Skill` entry from `SkillData.Body` (the body is
   already in the event); `load_skill` result shrinks to an acknowledgement.
3. `micro_handoff` tool + `MicroHandoff` event + reducer (remove all tool parts,
   drop emptied entries, append `Handoff`). Renderer drops an assistant message
   with no content.
4. Ladder policy: two watermarks + hysteresis from settings; emits `Redacted`
   events with `RedactData{Level, To}` (existing type, existing levels).
5. `ToolsChanged` event: skill load / MCP connect / MCP disconnect; Anthropic
   renderer carries it in the dialog; other renderers re-declare.
6. Total reducer: `Rebuild` skip-and-diagnose; diagnostic observable (log line or
   event the grader can read).
7. Persistence: append-on-write log, backup generation, truncation to
   max(N, anchor), snapshot on shutdown; recovery reuses ch11's load path.
8. Context Management tab (settings over WebSocket, fixed struct fields, ch9
   rules).
9. Grader `internal/grade/ch15_*` with the seven checks (eight once `keep` is
   added); P9 deletion audit.
10. Per-round-trip auto-redaction: after each round trip, stub every tool result
    the actor did not keep (Chapter 2's `RedactResult` on that round's results,
    recorded as an event). Gated per model in the features table (Variant B).
11. `keep_tool_results` tool: no arguments; keeps every result of the batch just
    received (CodeRhapsody's semantics). Its effect must be a recorded event so
    replay reproduces it.

Coder must verify before building:
- Does the fake vendor reject an unpaired tool call or result? If not, add it, or
  the micro_handoff check is decorative.
- Do the OpenAI and Gemini renderers accept a text entry after tool results in
  the same turn? (ASSUMED yes, unchecked.)
- `ToolsChangedData` field types (Chapter 10's declaration type).

---

## Open questions for Bill

1. ~~Chapter numbers~~ RULED: A = 15, B = 16, C = 17.
2. **Title:** "Keep the Words" recommended.
3. ~~Q6~~ RULED: build per-round-trip auto-redaction and `keep_tool_results` in
   ch15 (coder list items 10-11). The grader needs a check for it; weights must be
   rebalanced to stay at 100 (author's call; proposal: `keep` 10, taken 5 from
   `frozen prefix` and 5 from `total reducer`).
4. ~~Opener story~~ RULED correction: it is a *reset*, not a resend; see the
   story slot.
5. **`handoff_task`**: RULED that ensemble gets it (reversing the design doc's
   §A.6 deletion). Its role is open: design doc Q25. §15.9 cannot be finalized
   until that is ruled.
6. **Watermark thresholds and a forced `save_memory`**: RULED direction (design
   doc evening rulings 5-6), mostly ch16 material. What ch15 needs: the
   thresholds on the tool ladder (already rule 6) and, if Bill wants it in ch15,
   the upgrade-to-compact model switch.
