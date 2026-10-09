# Chapter 24 design — Extracting the GUI from the framework

*Analysis date: 2026-10-09. All line counts and import claims below were measured
against the live tree at that date, not recalled.*

## 1. The flaw

The GUI lives at `agent/internal/ws` (Go) and `agent/web/gui` (JS/CSS/HTML).
Three distinct problems compound:

1. **`internal/` makes the GUI unreachable.** No application outside this
   module — Homebrew-VTT is the motivating example — can import the websocket
   hub, the replay machinery, the GUI-RPC correlation layer, or serve the
   Artifact scroll component. The most reusable code in the repo is in the one
   directory Go forbids anyone from importing.

2. **The framework root links the GUI.** `agent/agent.go` imports
   `internal/ws` to export four symbols. Any consumer who imports the `agent`
   package — even a headless batch tool — links the websocket server, the
   GUI-RPC layer, and `gorilla/websocket`. The framework cannot be used
   without the GUI riding along.

3. **The GUI assets are served from a CWD-relative disk path.** There is no
   `go:embed` anywhere in the module. The binary only finds `web/gui/` when
   run from the right directory. (The ch22 grader harness works around this by
   setting `cmd.Dir`.)

The irony worth naming in the chapter: ch22 was about architectural decay, and
this flaw was *partially* addressed there (agents went headless, the Hub left
the library conceptually) — but the physical move never happened. The GUI
stayed `internal/`, and the root package kept its staples. Decay isn't an
event you fix once; it's a gradient you must keep checking.

## 2. Ground truth inventory (measured)

### Go, `agent/internal/ws` — 2,603 lines, 5 files

| File | Lines | Role |
|---|---|---|
| `handler.go` | 1,041 | Hub struct, ws client lifecycle, replay, inbound routing, settings/usage broadcast |
| `gui_call.go` | 187 | `CallGUI` browser-RPC with inflight correlation, `view_gui` tool |
| `mcp_port.go` | 136 | MCP-over-websocket tunnel (`ServeMCP`, receiver registry) |
| `tts_log.go` | 112 | TTS log endpoint |
| `gui_log.go` | 51 | GUI log file |

Dependencies: `internal/common` only (plus `gorilla/websocket`). It does NOT
use `common.Agent` — it takes six narrow injected deps:

```go
func NewHub(gate *common.PauseGate, send func(common.Inbound),
    guiLogPath string, eventLog *common.Log,
    settings common.SettingsSource, usage common.UsageSource) *Hub
```

plus a seventh dependency set *after* construction (the closure staple):

```go
// handler.go:39
Model func() string
// cmd/main.go:582
hub.Model = func() string { return eng.Cfg.Model }
```

### JS/CSS/HTML, `agent/web/gui` — 2,762 lines

| File | Lines | Reusable? |
|---|---|---|
| `gui.js` | 669 | app shell — monolith |
| `style.css` | 674 | mixed shell + component styles |
| `mcp.js` | 632 | MCP tunnel client — monolith |
| `tts.js` | 281 | TTS client |
| **`artifact-scroll.js`** | **264** | **yes — the component Homebrew-VTT wants** |
| `index.html` | 200 | shell |
| `renderers.js` | 38 | yes — pluggable part renderers |

### The staples in the framework root (`agent/agent.go`, 944 lines)

| Symbol | Line | Consumers found |
|---|---|---|
| `NewMCPWSTransport(hub *WSHub)` | 67 | **zero** (dead) |
| `type WSHub = ws.Hub` | 919 | **zero** (dead) |
| `NewWSHub(...)` | ~925 | **zero** (dead; `cmd/main.go` calls `ws.NewHub` directly) |
| `ServeHTTP(...)` | ~930 | **zero** (dead; cmd wires `http.FileServer` itself) |
| `NewMCPClientWSTransport(url, source)` | 84 | `cmd/virtual-user/main.go:146` — **live** |

Four of five are pure dead weight. The fifth (the *client* dial transport) is
real but is the wrong symbol for the root package to own, because:

### gorilla/websocket links in two ways, not one

`go.mod` deps: `creack/pty`, `gorilla/websocket`. Gorilla is imported by:
1. `agent/internal/ws/*` — the GUI server (fixed by the extraction), and
2. `agent/internal/mcp/client_ws_transport.go` — the MCP client-WS dial
   transport. The root package imports `internal/mcp` for its MCP aliases, so
   **even after cutting `ws`, every framework consumer still links gorilla**
   through `mcp`.

A truly headless framework needs the client-WS transport out of `internal/mcp`
too.

## 3. Target architecture

```
                      (public)                    (public)
   Homebrew-VTT ───► agent/gui ───┐        ┌──── agent/mcpws
                                  │        │     (ws dial transport,
   ensemble cmd ──────────────────┤        │      used by virtual-user)
                                  ▼        ▼
                            agent (root, headless)
                                  │
                        internal/{common,llm,mcp,skills,...}
```

- **`agent/gui`** — new PUBLIC package: the moved `internal/ws` plus the
  embedded `web/gui` assets. Exposes the server type, `ServeWS`, `ServeMCP`,
  `CallGUI`, the `view_gui` tool, and an exported `embed.FS` of assets so
  consumers can mount individual files (Artifact scroll without the shell).
- **`agent/mcpws`** — tiny PUBLIC package holding the client-WS dial
  transport (one file moves from `internal/mcp`). `virtual-user` imports it
  directly; the dead root aliases are deleted.
- **root `agent` package** — loses its `ws` import entirely. Framework
  consumers link neither the GUI nor gorilla. Go only links what you import;
  after this cut, `go list -deps` of a headless consumer is the proof.
- **Assets** — `go:embed` inside `gui`, with a `-gui-dir` override for live
  development (edit JS without rebuilding). Binary works from any CWD.

### Public-surface gap to close

The root package already aliases almost everything `gui`'s signature needs
(`PauseGate`, `Inbound`, `Log`, `SettingsStore`, `Observation`, ...) — but
**`common.SettingsSource` and `common.UsageSource` have no aliases**. External
consumers couldn't name two of the constructor's parameter types. Either add
the two aliases or have `gui` declare its own narrow interfaces. (Declaring
them in `gui` is self-contained and reads better in godoc; the alias route is
one line each. Lean: declare in `gui` — the GUI is the only consumer of these
shapes, so the interface belongs with its consumer.)

## 4. The Hub split, and the Ensemble question

The Hub conflates two jobs:

**Transport/presentation** (stays in `gui`): ws client set, replay-on-connect
from the event log, JSON-RPC broadcast, GUI-RPC inflight correlation, MCP
tunnel, gui/tts logs, static assets.

**Agent-session access** (the seam): pause gate, inbound send, event log,
settings, usage, current model. Today these arrive as six loose constructor
params plus the post-construction `Model` closure. This bundle is *exactly the
per-agent surface a multi-agent manager must present* — it is proto-Ensemble.

Bill's framing: "Hub may be renamed Ensemble, and will be the class that
manages multiple agents when we add sub-agents." My read, for discussion:

- The **transport half should not carry the Ensemble name** — it's a GUI
  server, and Homebrew-VTT wants it *without* any multi-agent machinery.
- The **session-bundle half** is what grows into Ensemble: the thing that owns
  N agents, routes `Inbound` by agent ID, and fans source-tagged observations
  out to frontends (GUI server, chat observer, virtual user are all already
  `common.Observer`s — the attach seam exists and supports multiple
  observers).

### Recommendation for ch24: carve the seam, don't build Ensemble

Group the constructor deps into one struct and kill the closure staple:

```go
type AgentHooks struct {        // in gui (or common — see Q2)
    Gate     *common.PauseGate
    Send     func(common.Inbound)
    EventLog *common.Log
    Settings SettingsSource
    Usage    UsageSource
    Model    func() string      // construction-time, no longer set-after
}
func New(hooks AgentHooks, guiLogPath string) *Server
```

Rejected for ch24: defining a `Session` interface now. ch22's precedent
applies — a sibling interface is a second path to the agent (the `hub.Model`
mistake generalized), and designing Ensemble's per-agent interface before its
real consumer exists would be speculation. ch25 will force the right shape
when the GUI needs to *address* agents by ID; that's when `AgentHooks` becomes
`Ensemble.Hooks(agentID)` or similar, and the chapter can show the struct
growing into the interface honestly.

This also gives ch25 a clean narrative: ch24 leaves a visible dotted line
("one `AgentHooks` = one agent; what owns the map?") that the sub-agent
chapter answers with Ensemble.

## 5. Other rot observed (documented per the ch22 habit)

1. **Dead exports** — the four zero-consumer symbols in §2. Deleted, not
   deprecated; nothing external can be using them (they were born this
   session's archaeology).
2. **The `Model` closure staple** — `handler.go:39` + `cmd/main.go:582`, with
   an apologia comment. Fixed by `AgentHooks` (construction-time, documented).
3. **CWD-relative asset serving** — fixed by embed.
4. **Root package re-accretion watch** — `agent.go` is 944 lines (it was cut
   to 390 in ch7). Much of the growth is legitimate (AgentSpec/clamp from
   ch23, ~45 type aliases), but by the func-body-percentage lens it deserves
   a look during ch24's edit anyway. Not a ch24 goal.
5. **JS monoliths** — `gui.js` (669) and `mcp.js` (632) are flat files, while
   `artifact-scroll.js` shows the componentized pattern. A JS refactor is NOT
   proposed for ch24 (no behavior change, big diff) — but exporting the
   embed.FS lets Homebrew-VTT take `artifact-scroll.js` + `renderers.js`
   without taking the shell, which is the actual reuse ask.
6. **`--mcp-connect` in the main binary** does not use the root
   `NewMCPClientWSTransport` (only virtual-user does) — verify during stage 5
   what main.go's path actually uses, and unify if there are two.

## 6. Staged plan (every stage compiles, tests green, wire unchanged)

- **Stage 0 — baseline.** Record live-tree grades (known pre-existing: ch3 70,
  ch4 90, ch19 85), `go test ./...`, full sweep reference.
- **Stage 1 — delete dead exports.** `WSHub`, `NewWSHub`, `ServeHTTP`,
  `NewMCPWSTransport` out of `agent.go`; root loses the `ws` import.
  Framework-headless property now true *except* for gorilla-via-mcp.
- **Stage 2 — move the package.** `git mv agent/internal/ws agent/gui`,
  package `ws`→`gui`, `Hub`→`gui.Server`. Mechanical; cmd rewires. Wire
  protocol byte-identical (ch22's behavioral harness is the regression net).
- **Stage 3 — the seam.** Introduce `AgentHooks`; `Model` moves into it;
  delete the field-set-after-construction path.
- **Stage 4 — embed assets.** `agent/web/gui` moves to `agent/gui/web/` (or
  stays, with embed pattern reaching it — embed cannot cross `..`, so the move
  is forced); exported `Assets embed.FS`; `-gui-dir` dev override.
  ⚠ Requires a package-level `var` — see Open Question 1.
- **Stage 5 — gorilla-free framework.** `client_ws_transport.go` moves from
  `internal/mcp` to public `agent/mcpws`; root alias deleted; virtual-user
  imports it directly. Proof: `go list -deps` of a headless consumer contains
  no `gorilla/websocket` and no `agent/gui`.
- **Stage 6 — reuse smoke test.** A tiny example (grader-side, or
  `examples/`) that builds a *non-ensemble* app: fake agent hooks + 
  `gui.Server` + one embedded asset served. This is Homebrew-VTT's import
  path, proven in-tree.
- **Stage 7 — coder review doc, grader, mutants, sweep, snapshot
  `solutions/ch24`.** Snapshot is a copy, not a reference — tree-to-tree diff
  after any late fix.

## 7. Grader sketch (behavioral, per the ch23 lessons)

1. **Headless-linkage check (the chapter's core property).** Build a minimal
   consumer importing the framework root; assert `go list -deps` excludes
   `gorilla/websocket` and the gui package. Mutant: re-add a `ws` import to
   the root → must fail this check and only this check.
2. **Reusability check.** Compile-and-run a toy app that stands up
   `gui.Server` with fake hooks, connects a ws client, and sees replay — no
   ensemble binary involved. Sentinel in the event log must reach the client.
3. **Wire-compat check.** ch22-style session against the real binary —
   subscribe, drive a turn, assert frames. Guards the "mechanical move" claim.
4. **Any-CWD check.** Run the binary with cwd=/tmp; fetch `index.html` and
   `artifact-scroll.js` over HTTP. Mutant: drop the embed → fails.
5. Positive controls paired with every absence assertion (ch23 lesson: every
   test of an absence needs proof the detector can fire).

## 8. Open questions for Bill

**RESOLVED 2026-10-09: Bill approved proceeding on the coder's leans — (1)
`embed.FS` carve-out in the ch5 rule; (2) `AgentHooks` in `gui`; (3) names
`gui` / `mcpws` / `gui.Server`, Hub name dies, Ensemble born in ch25 as a new
type; (4) one chapter; (5) design for Go-import + exported FS, minimal polish.
Draft TL;DR and author brief: `book/brief-ch24-author.md`.**

1. **The embed `var` vs ch5's package-var rule.** `go:embed` requires
   `var assets embed.FS` — there is no var-free form. The ch5 grader greps
   `^var ` across non-test files under `agent/` (12 current survivors, all
   `*Names` tables exempted by the "names" carve-out). Options: (a) exempt
   `embed.FS` in the ch5 rule — defensible, the compiler makes it effectively
   immutable; (b) exempt the `gui` package path; (c) something cleverer I'm
   not seeing. This needs your call before stage 4.
2. **Where does `AgentHooks` live** — in `gui` (consumer-owned, my lean) or in
   `common` (if you expect ch25's Ensemble to hand the same bundle to multiple
   frontend kinds)?
3. **Package names.** `agent/gui` and `agent/mcpws` are my proposals. If the
   Hub's *name* should survive anywhere, I'd argue it dies entirely in ch24:
   transport half becomes `gui.Server`, session half becomes `AgentHooks`,
   and **Ensemble is born in ch25 as a new type**, not a rename — renames
   carry old connotations, and this one would carry four.
4. **One chapter or two?** The extraction (stages 1–4) and the
   gorilla-free/mcpws + reuse-proof work (stages 5–6) could split. My lean is
   one chapter: the thesis "the framework must not know the GUI exists" isn't
   proven until the linkage check passes, and that needs stage 5.
5. **Homebrew-VTT's actual consumption model** — Go-import of `gui` (needs
   this repo as a module dependency or a `replace` directive), or
   copy-the-files? Affects how much API polish `gui` needs now vs later.

## 9. Non-goals for ch24

- Building Ensemble / sub-agents (ch25).
- JS componentization of `gui.js`/`mcp.js`.
- Touching `solutions/edition-2/` or `book/edition-2/` (Astra's).
- Wire protocol changes of any kind.

## 10. Role split for the chapter prose (Bill's ruling, 2026-10-09)

New division of labor, starting with ch24:

- **The coder writes the TL;DR.** The TL;DR is the page contract and the
  grader fixture is its ground truth; the role that builds the checks is the
  only one that can guarantee the published words and the graded behavior
  never drift apart. The TL;DR's final wording is frozen against the actual
  ch24 fixtures, never against this design doc alone.
- **The author writes the plain-words section and the narrative.** The why
  belongs to the author, the what belongs to the coder.
- The author may propose voice edits to the TL;DR, but any change in meaning
  goes back through the coder and the grader before publication.

The authoritative statement of this split is in
`book/chapter-writing-procedure.md` (Roles and §2). For ch24 the roles
collapse: the coder is both reference implementer and grader engineer, so
the coder writes the TL;DR; the author works from the coder review.
