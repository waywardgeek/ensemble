# Chapter 24 author brief — Freeing the GUI

*From the coder, 2026-10-09. Read `docs/ch24-design.md` first; it carries the
measured inventory and the staged implementation plan. This brief is what the
author needs that the design doc does not say, plus the draft TL;DR.*

*Role split per `book/chapter-writing-procedure.md` (Roles, §2): the coder
owns the TL;DR; the author owns the plain-words explanation and all other
prose. The TL;DR below is draft v1 — its final wording freezes against the
actual ch24 grader fixtures, and any meaning-level edit routes back through
the coder.*

## 1. Rulings adopted (Bill approved proceeding on the coder's leans)

1. **ch5 package-var rule gains an `embed.FS` carve-out.** `go:embed` has no
   var-free form; the carve-out names the type, not a package path.
2. **`AgentHooks` lives in `gui`** — consumer-owned interface shapes;
   `common` gains nothing.
3. **Names:** public `gui` package (from `internal/ws`), public `mcpws`
   package (the MCP client-WS dial transport), `ws.Hub` becomes
   `gui.Server`. The Hub name dies in this chapter. Ensemble is born in
   ch25 as a new type, not a rename.
4. **One chapter.** The thesis is only proven by the `go list -deps` check,
   which needs the mcpws split; splitting the chapter would end part one on
   an unverified claim.
5. **Homebrew-VTT consumption:** design for Go-import plus the exported
   embedded FS; polish stays minimal until a real consumer exists.

## 2. Thesis and title

The framework must not know the GUI exists. Two inversions of the same
mistake: the root package imported the GUI (so every headless consumer
linked a websocket server it never started), and the GUI sat under
`internal/` (so the most reusable code in the repo was the code nobody
else was allowed to import).

Title candidates, author's choice: "Freeing the GUI", "The Optional GUI",
"Setting the GUI Free". Avoid "Extracting the GUI" — extraction is the
mechanism, optionality is the lesson.

## 3. TL;DR (coder-owned, v2, FROZEN against fixtures 2026-10-09)

---

## TL;DR

The framework must not know the GUI exists. Go links every package an
import reaches, and the framework root imported the websocket server:
every consumer of the agent library shipped a GUI it never started,
plus `gorilla/websocket`. The GUI also lived under `internal/`, where
no other application is allowed to import it. This chapter inverts
both mistakes, and the proof is mechanical: `go list -deps` on a
headless consumer names neither the GUI package nor gorilla.

1. Delete the four dead exports from the framework root: `WSHub`,
   `NewWSHub`, `ServeHTTP`, `NewMCPWSTransport`. Nothing calls them.
   The root package drops its `internal/ws` import entirely.
2. Move `internal/ws` to a public `gui` package. `ws.Hub` becomes
   `gui.Server`. The wire protocol does not change; a chapter 22
   behavioral session passes unmodified.
3. Collapse the server's agent-session dependencies (pause gate,
   inbound send, event log, settings source, usage source, model
   getter) into one construction-time `AgentHooks` value defined in
   `gui`; the GUI's own log path stays a separate constructor
   argument. The post-construction `Model` field assignment dies with
   the Hub name.
4. Embed the GUI assets with `go:embed`. The binary serves the GUI
   from any working directory. A flag overrides with a disk path for
   live development. Export the embedded FS so a consumer can mount
   single components without the shell.
5. Move the MCP client-WS dial transport from `internal/mcp` to a
   public `mcpws` package. After this, importing the framework root
   links no websocket code at all.
6. The public surface is enough to actually reuse: root-alias the
   event vocabulary (`Event`, `MessageData`, `PartList`, `TextPart`,
   and the actor constants) so a consumer can build a log for the
   scroll to replay, because you cannot construct what you cannot
   name. Keep the agent-eyes MCP relay (`ServeMCP` on a bare
   listener) in `gui`: an agent that can see this GUI is as much the
   point as a human who can.
7. Chapters 1-23 behavior is preserved: full sweep green. The ch5
   package-var rule gains exactly one carve-out, `embed.FS`
   variables, because `go:embed` has no var-free form.

Yours: the package and type names (`gui`, `mcpws`, `Server`,
`AgentHooks` are the reference's choices), the hooks struct's exact
field spelling, the override flag name, and where the asset files
live inside the package.

Build and grade: `make grade24`.

| # | Behavior | Check | Points |
|---|---|---|---|
| 1 | Framework links no GUI, no gorilla | `go list -deps` on a root-only consumer compiled OUTSIDE the module | 20 |
| 2 | GUI publicly reusable | external consumer + fake hooks: planted event replays to a ws client, a prompt reaches the Send hook, and the MCP relay answers JSON-RPC with the request's own id | 15 |
| 3 | Wire unchanged | ch22-style scripted session against the real binary; recorded vendor requests prove the GUI-socket model switch landed | 15 |
| 4 | Assets embedded | binary started from an unrelated cwd serves the GUI | 15 |
| 5 | Components without shell | consumer mounts the Artifact scroll from the exported FS; `index.html` 404s | 10 |
| 6 | MCP without gorilla | headless MCP consumer links no gorilla; virtual-user still dials through `mcpws` | 15 |
| 7 | Root stays clean | structural: root exports no websocket-typed symbols, imports neither `gui` nor gorilla | 10 |

Frozen against the fixtures: `make grade24` 100/100 on the reference
tree, `make grade24-audit` 7/7 mutants killed with exact check sets.
Every absence check ships with a positive control that proves the
detector fires.

---

*End of draft TL;DR.*

## 4. Proposed section outline

Each section names its exhibit. Real output replaces placeholders after
implementation.

- **Opener** — documented incident, no invention needed: Bill noticed the
  flaw himself on 2026-10-09, two chapters after writing the chapter *about*
  architectural decay, while wanting the Artifact scroll for Homebrew-VTT.
  The hook: ch22 repaired decay and this flaw survived ch22. Decay is not an
  event you fix once; it is a gradient you keep checking.
- **Plain words (author's)** — the two-sentence physics: Go links what you
  import; `internal/` forbids reuse. The dashboard welded to the engine.
- **24.1 The archaeology** — the §2 inventory from the design doc. Star
  exhibit: four of the five WS exports in the root package have zero
  consumers. The staple was not even load-bearing.
- **24.2 Two paths to gorilla** — cutting `ws` is not enough; the MCP
  client-WS transport drags gorilla through `internal/mcp`. Teaches: check
  the full deps graph, not the import you know about.
- **24.3 The split** — Hub's two jobs (transport vs agent-session access);
  `AgentHooks`; the Model closure staple dies. Ends on the dotted line: one
  `AgentHooks` is one agent — what owns the map? (ch25 answers: Ensemble.)
- **24.4 Embedding the assets** — CWD-relative serving as rot; `go:embed`;
  the ch5 rule carve-out and why rules get amended rather than gamed.
- **24.5 The proof** — real `go list -deps` before/after. The chapter's
  thesis verified mechanically.
- **24.6 Reuse for real** — the toy external app: `gui.Server` + fake hooks
  + the Artifact scroll served without the ensemble binary.
- **24.7 Taking it for a spin** — real transcript post-implementation.
- **24.8 Checks** — frozen check table.

## 5. Voice and story notes

- Em-dash ban in prose is absolute; `--` allowed in headings (ch23 pattern).
- Bill mentions: the opener likely spends both of the ≤2 budget (noticing
  the flaw; Homebrew-VTT). Keep the body third person.
- §4.1 story preservation: the preserved incident is this one (documented in
  `docs/ch24-design.md` and this brief). No first-edition story is displaced.
- Do not call gorilla a problem. It is a fine library in the wrong place on
  the dependency graph. The flaw is ours, not the dependency's.

## 6. Verified vs assumed

- **Verified (measured 2026-10-09, live tree):** every line count, import
  claim, dead-export claim, and the 12 surviving `^var` lines in the design
  doc §2. The zero-consumer claim for the four root exports.
- **Assumed until implementation:** grader point values; the exact
  `go list -deps` output; whether `--mcp-connect` in the main binary has a
  second dial path (design doc §5 item 6); diff stats of the move.
- **Unknown:** Homebrew-VTT's final consumption model (ruling 5 covers it).

## 7. Do NOT add

- Ensemble, sub-agents, or any agent-map machinery (ch25).
- JS refactor of `gui.js`/`mcp.js` monoliths.
- Wire protocol changes of any kind.
- References to the second edition or its workflow.
- A Hub deprecation alias. The name dies; nothing external used it.

## 8. What the author waits for

Implementation (stages 0-7 of the design doc) produces: the coder review
(Must-fix / Enrichment / Do NOT add / verified-vs-not), real exhibit output
for 24.5-24.7, the frozen TL;DR, and the final check table. Prose written
before that point can only be the opener, plain-words, and 24.1-24.3
skeletons, all against the design doc's measured facts.
