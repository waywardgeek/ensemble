# Chapter 24 Coder Review — Freeing the GUI

*Written by the CODER (grader engineer) for the AUTHOR. This is the primary
source for the chapter prose. Design doc: `docs/ch24-design.md`. Author
brief with draft TL;DR: `book/brief-ch24-author.md`. Everything below is
measured on the live tree unless marked otherwise.*

## 1. What shipped, in commit order

| Commit | Stage | What |
|---|---|---|
| `247fb64` | 1 | Deleted the four dead root exports (`WSHub`, `NewWSHub`, `ServeHTTP`, `NewMCPWSTransport`) and the root's `internal/ws` import. 43 deletions, zero consumers anywhere. |
| `588e865` | 2 | `git mv internal/ws → gui`, `ws.Hub → gui.Server`, `NewHub → New`. Wire protocol byte-identical; ch22 behavioral harness was the net. |
| `f3ff34b` | 3 | `AgentHooks`: six injected deps plus the post-construction `Model` closure collapse into one construction-time bundle. The staple apologia comment dies with the staple. |
| `b24ec98` | 4 | `web/gui → gui/web` under `go:embed`; exported `Assets`; `StaticHandler(dir)` with `""` = embedded; `--gui-dir` becomes the live-dev override. ch5 grader gets the `embed.FS` carve-out. |
| `b467faf` | 5 | `internal/mcp/client_ws_transport.go → agent/mcpws` (public). The framework closure is now gorilla-free: the headless proof passes. |
| `a0ee97d` | 6 | External reuse smoke test in `package gui_test` + root aliases `SettingsSource`, `UsageSource`. |
| `fa28658` | 6b | Root-aliased the event vocabulary (`Event`, `EventType`, `MessageData`, `PartList`, `Seq`, plus `MessageReceived`/`ActorHuman`/`ActorAgent`). Reuse test now replays a consumer-planted sentinel to a real ws client. |
| `2089a14` | 7a | Grader: seven checks, probes compiled OUTSIDE the module, agent-eyes (`ServeMCP`) leg included. |
| `6896df7` | 7b | Mutation audit script; wire-compat ground-truth assertion. |

Grades at review time: `make grade24` **100/100**, `make grade24-audit`
**7/7 killed, 0 misbehaved**. Cross-chapter sweep: running; §8 records the
result before the TL;DR freezes.

## 2. The exhibits, in the order the chapter should hit them

### 2.1 The staple was not even load-bearing

The framework root re-exported four websocket symbols. `grep -rn` across
the entire repo found **zero consumers**: `cmd/main.go` called `ws.NewHub`
directly, and nothing anywhere called the root wrappers. The chapter's
decay story opens here: the staple survived ch22 — the chapter ABOUT
decay — because nothing exercised it. Dead exports do not fail tests.
`NewWSHub` was ch8's published API; its retirement is the first time the
book deletes a former chapter's surface, and the honest telling is that
ch8's API had already died in practice the moment cmd stopped calling it.

### 2.2 The GUI never needed the framework

`internal/ws` depended only on `internal/common` — never on the agent,
the engine, or the actor. Six narrow injected deps plus one
post-construction closure. The flaw was pure placement: the most
reusable code in the repo sat in the one directory Go forbids anyone to
import. The move itself was mechanical (perl word-boundary renames; BSD
sed's `\b` silently no-ops — worth a parenthetical in prose).

### 2.3 AgentHooks and the death of the Model staple

The old wiring set `hub.Model` AFTER construction, with an apology
comment at the assignment site (`cmd/main.go:582` in the pre-move tree,
visible at `git show d209a65:agent/internal/ws/handler.go`). The repair
is a construction-time bundle, and the closure survives INSIDE the
bundle with its reason documented: the operator can switch models
mid-session, and a string captured at construction would price every
later turn at the old rate while looking entirely correct. `AgentHooks`
is one agent's control surface handed to a frontend — the chapter should
end on the dotted line this draws: one AgentHooks = one agent; what owns
the map? (ch25's question. Do not answer it.)

### 2.4 The second gorilla path nobody briefed

Cutting `internal/ws` did NOT make the framework websocket-free:
`internal/mcp/client_ws_transport.go` also imported gorilla, and the
root imports `internal/mcp`. Only `cmd/virtual-user` ever dialed through
it. The fix is a public `agent/mcpws` package, and the lesson is the
chapter's sharpest: **check the full dependency closure, not the package
you just moved.** `go list -deps` is the tool, and the grader runs it as
check 1.

### 2.5 You cannot construct what you cannot name (stage 6b)

The reuse fixture forced an API decision the design doc missed: a
consumer with fake hooks could serve the GUI, but the event vocabulary
was internal, so it could not BUILD a log for the scroll to replay — a
GUI with an empty log renders an empty scroll. Exporting
`Event`/`MessageData`/`TextPart` aliases is what makes reuse real
(Homebrew-VTT feeds the scroll its own campaign content). One collision:
`Actor` was already taken by the llm actor runtime at root, so the event
actor TYPE stays unaliased — the constants carry the type, which is all
a consumer needs. Good two-sentence exhibit on alias collisions in a
flat namespace.

### 2.6 Agent eyes are part of the reuse surface (Bill's mid-build call)

`ServeMCP` — the `--mcp-port` relay that gives an agent
`gui_snapshot`/`gui_click` eyes on the GUI — moved with the package and
was already public: a method on `gui.Server` taking a bare
`net.Listener`, zero hidden deps. Bill's ruling mid-implementation: the
agent seeing the GUI matters as much as the human seeing it, so lock it
into the graded surface. The fixture leg has one subtlety worth prose:
with no browser attached the relay answers `noGUIError` IMMEDIATELY,
but a connected ws client that ignores MCP frames parks the request
forever waiting on a browser reply — so the probe dials the relay
BEFORE any ws client connects. Read the dispatch condition, not just
the error constant.

## 3. The grader (for §24.8 and the TL;DR freeze)

Seven checks, 100 points, every absence check paired with a positive
control:

1. **headless-linkage (20)** — probe module OUTSIDE the submission
   (replace directive), imports only the framework root; `go list -deps`
   must name neither `gui`, `mcpws`, nor gorilla. The gui-importing
   probe is the control proving the detector sees gorilla at all.
2. **public-reuse (15)** — the external guiapp probe: consumer-planted
   sentinel replays to a ws client; a browser prompt reaches the one
   Send hook the consumer supplied; the MCP relay answers JSON-RPC with
   the request's own id. Compiling AT ALL from outside the module is
   half the check — an in-tree test cannot prove `internal/` stopped
   blocking anyone.
3. **wire-compat (15)** — ch22's scripted session rerun whole: stdin
   prompts, GUI-socket model switch, three replies. The recorded vendor
   requests are the ground truth; the assertion is endpoint-shaped
   (starts dear, ends cheap) because the switch lands at a turn
   boundary, not instantly.
4. **assets-embedded (15)** — the real binary started from an empty,
   unrelated cwd serves `/` and `/renderers.js` over HTTP.
5. **components-without-shell (10)** — the probe mounts the Artifact
   scroll from the exported FS; `/index.html` must 404 from the
   component-only consumer.
6. **mcp-without-gorilla (15)** — `internal/mcp` stays IN the headless
   closure while gorilla stays out; virtual-user still links both
   `mcpws` and gorilla (the positive control that the transport
   survived the split).
7. **root-clean (10)** — structural scan: no gui/gorilla imports in root
   files, websocket-era symbols gone; the `func NewAgent(` sighting is
   the control that the scanner reads the right files.

Mutation audit: 7 mutants, each proved to land and compile, each killing
exactly its expected set. Mutants 1 and 6 kill multiple checks through
real, distinct dependencies (documented in the script header).

## 4. Harness war stories (author's pick, at most one)

- **The orphan held the pipe.** First grader run "hung" for ten minutes:
  the report was finished and buffered. Killing a `go run` wrapper
  orphans the grandchild server, which inherits stderr and holds the
  output pipe open — tail never sees EOF. Fix: build the probe to a real
  binary and exec it directly. (Same family as the ch22 Setpgid lesson.)
- **Mutant 4 found a weak check before it shipped.** The reply-count
  assertion passed a server that dropped `update_settings` frames on the
  floor; the recorded request models are the observable that catches it.
  The mutation audit improving the GRADER, not just validating it, is
  the page-contract system working as designed.

## 5. Must-fix (author)

1. Design doc §2 line counts were measured BEFORE the move; if the
   chapter quotes sizes, re-measure or date them.
2. The §1 irony (flaw survived the decay chapter) is documented reality —
   keep it; it is the opener.
3. The brief's §3 is now the FROZEN TL;DR v2 (reconciled with the
   agent-eyes leg by the coder on 2026-10-09); treat it as verbatim
   contract text, not as a draft.

## 6. Do NOT add

- Ensemble, sub-agents, or any agent-map machinery (ch25 owns these).
- A JS refactor of `gui.js`/`mcp.js` monoliths.
- Wire protocol changes of any kind, or a Hub deprecation alias (the
  name dies; nothing external used it).
- References to the second edition or its workflow.
- Do not call gorilla a problem. Fine library, wrong place on the graph.

## 7. Verified vs. not

- **Verified (this session, live tree):** all grades and audit results in
  §1; the zero-consumer claim; the `[opus opus sonnet sonnet]` model
  sequence; the noGUIError zero-browser semantics; probe compile from
  outside the module.
- **Pre-existing, NOT ch24 regressions (baseline-worktree proven):**
  ch22 at 85 on the live tree (star-topology check: `internal/tools`
  imports `agent/sandbox` directly — ch23's doing; open architecture
  question for Bill, ch25 candidate). ch19 at 85 (wants
  `agent/events.jsonl`). ch3 70 / ch4 90.
- **Pending at writing time:** cross-chapter sweep (§8), solutions/ch24
  snapshot.

## 8. Status (updated as the tail stages land)

- [x] Stages 0–6b implemented and committed
- [x] Grader 100/100; mutation audit 7/7
- [ ] Cross-chapter sweep green
- [ ] `solutions/ch24` snapshot + tree-to-tree diff clean
- [ ] TL;DR frozen against fixtures (then the author starts)
