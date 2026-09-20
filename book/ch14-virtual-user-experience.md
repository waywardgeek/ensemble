# Ch14 Virtual User Experience Report

*Written by CodeRhapsody, 2026-09-20. Raw material for the author.*

## What We Built

A **virtual user agent** — a separate binary (`cmd/virtual-user/main.go`) that
connects to the coding agent's GUI via MCP over WebSocket and drives it the way
a human would. No filesystem tools, no shell access. Just:

- `gui_snapshot` — see the DOM as markdown
- `gui_click(selector)` — click buttons
- `gui_input(selector, text)` — type into fields
- `gui_submit(selector)` — press Enter
- `tts_queue` — hear what's being spoken
- `wait_for_idle(timeout)` — block until the agent finishes
- `sleep(seconds)` — simple delay
- `file_report(title, body)` — write findings to a log

The virtual user connects to the hub via its own WebSocket, with MCP traffic
tagged `source:"vu"` so the hub routes responses back to the right client.
The browser's `mcp.js` echoes the source tag in every response.

## The Task

> Write a Tower of Hanoi solver in Go that solves a 4-disk puzzle and prints
> each move. Then use dlv to set a breakpoint in the recursive function, step
> 3 levels deep into the recursion, and report all visible local variables
> at that depth.

## Infrastructure Bugs Found and Fixed

### Bug 1: mcp.js Never Loaded (critical)

**Symptom:** MCP handshake hung forever. The virtual user's `initialize`
request was forwarded by the hub to the browser, but the browser never
responded.

**Root cause:** `mcp.js` was never included in `index.html`. The browser had
no MCP server running. All of ch12's MCP infrastructure was built and tested
via `--mcp-pipe` (in-process), so the browser-side integration was never
verified against a real page load.

**Fix:** Added `<script src="mcp.js"></script>` to `index.html`, before
`gui.js` so the WebSocket constructor gets monkey-patched before `gui.js`
creates the connection.

**Lesson:** End-to-end testing matters. The grader's `--mcp-pipe` mode
tested the Go MCP client against a Go fake server. The browser MCP server
was never exercised in a real browser until the virtual user tried to use it.

### Bug 2: Pause Gate Blocks Tool Dispatch (critical)

**Symptom:** The coding agent received the prompt, generated thinking +
`write_file` for `hanoi.go`, transitioned to `tools_pending` — and froze.
No `tool_dispatched` event, no `tool_finished`. The virtual user watched
`wait_for_idle` return "timeout: agent still tools_pending" forever.

**Root cause:** When the virtual user calls `gui_input("#prompt-input",
"Write a Tower of Hanoi...")`, the MCP handler dispatches a DOM `input` event
on the element. `gui.js` has a listener on the prompt input that sends
`{type: "pause"}` to the hub when the user starts typing, so tool calls
are held while the human composes their prompt. The programmatic `gui_input`
triggered this listener, pausing the coding agent. Then `gui_submit` fired
Enter, the prompt was sent, the agent started thinking — but the pause gate
was still engaged. When the engine tried to dispatch `write_file`, it blocked
on `a.gate.WaitIfPaused(a.ctx)` forever.

Why the unpause never came: the Enter keydown handler clears
`input.value = ""`, but programmatic value assignment doesn't fire an `input`
event, so `userTyping` stays `true`. The `unpause` message is only sent when
the input becomes empty via the `input` event listener, which never fires.

**Fix (two parts):**
1. `mcp.js`: Set `window._mcpProgrammaticInput = true` before dispatching the
   `input` event. `gui.js` checks this flag and skips the pause handler.
2. `gui.js`: After the Enter keydown handler sends a prompt, always send
   `{type: "unpause"}` to unblock the gate.

**Lesson:** GUI event handlers designed for human interaction have implicit
assumptions (typing → pause → Enter → resume) that break when a machine
drives the same events. The programmatic input flag is the fix, but the
deeper lesson is that the pause gate needs to be unpause-complete: every
code path that ends the typing interaction must send unpause.

### Bug 3: No gui_submit Tool

**Symptom:** The virtual user could type text into the prompt input via
`gui_input`, but couldn't submit it. There was no send button in the GUI
(prompts are submitted by pressing Enter). The virtual user tried clicking
the input, appending `\n` to the text, clicking nearby buttons — nothing
triggered submission.

**Fix:** Added `gui_submit(selector)` to `mcp.js` — dispatches a
`KeyboardEvent("keydown", { key: "Enter", ... })` on the target element.
The keydown handler in `gui.js` picks this up exactly as it would a real
Enter press.

### Bug 4: Hub Source-Tag Routing

**Symptom:** (Caught during development, not at runtime.) The hub's
`handleClientMessage` needed to route MCP responses by source tag so the
virtual user's responses don't leak to the coding agent.

**Fix:** `SetMCPReceiver(source, fn)` / `RemoveMCPReceiver(source)` /
`BroadcastJSONRPC(data, source)` in the hub handler. When a browser
response arrives with `source:"vu"`, it routes to the virtual user's
receiver. When the virtual user sends a request, the hub forwards it to
all browser clients.

## GUI/UX Bugs Found by the Virtual User

### Bug 5: No Input Validation on Settings

**Discovery:** The virtual user (on its first exploratory run before the
pause-gate fix) opened the settings panel and entered: temperature=-5,
max_tokens=999999999, tts_speed=100. All values were accepted without
validation or clamping. The garbage values persisted across sessions.

**Status:** Not yet fixed. Needs input validation + clamping in the
settings panel.

### Bug 6: GUI Truncates Content

**Discovery:** Bill noted from a screenshot that tool cards on the right
pane truncate command arguments and output. The virtual user sees the full
DOM (gui_snapshot returns everything), but the actual human user sees
truncated content.

**Status:** Not yet fixed. The GUI needs expandable/scrollable tool cards.

### Bug 7: Hamburger Menu Toggle Not Visible in Snapshot

**Discovery:** The virtual user clicked the hamburger button and took a
snapshot, but the DOM didn't visually indicate whether the sidebar was
expanded or collapsed. Bill confirmed it works visually (the sidebar
expands/collapses) but the state isn't reflected in the DOM attributes
that gui_snapshot reads.

**Status:** Not yet fixed. The sidebar toggle could expose its state via
an `aria-expanded` attribute or a class name change.

## The Successful Run (after fixes 1-3)

After applying the pause-gate fix and gui_submit:

1. **Virtual user typed the prompt** via `gui_input` + `gui_submit`. The
   prompt was sent as `{type: "prompt"}` (not hint). ✅

2. **Coding agent received it**, started thinking: "I'll start by writing
   the Go program, then use dlv to debug it." ✅

3. **Agent called `write_file`** with correct `hanoi.go`:
   ```go
   func hanoi(n int, from, to, via string) {
       if n == 0 { return }
       hanoi(n-1, from, via, to)
       fmt.Printf("Move disk %d from %s to %s\n", n, from, to)
       hanoi(n-1, via, to, from)
   }
   func main() { hanoi(4, "A", "C", "B") }
   ```
   **Tool dispatched and finished.** The pause gate fix worked. ✅

4. **Agent ran `go run hanoi.go`** — all 15 moves printed correctly. ✅

5. **Agent said:** "Program works correctly, producing all 15 moves for the
   4-disk puzzle. Now let's debug it with Delve." ✅

6. **Agent verified dlv** — `which dlv; dlv version` → Delve 1.27.2. ✅

7. **Agent started headless dlv** on port 34567, then realized it should be
   interactive. Killed the headless process and started
   `dlv debug hanoi.go` with `ai_callback_pattern: "\\(dlv\\) "`. ✅

8. **The interactive dlv session hung.** The `(dlv)` prompt appeared in the
   GUI (confirmed via screenshot — you can see it), but the tool never
   returned. The `ai_callback_pattern` didn't match, or more precisely:
   the tool function reads from the PTY synchronously in its own goroutine,
   and `Wait()` is never called concurrently. ❌

## Fixed: Interactive Process Support (commit b139325)

The ensemble agent's `run_command` tool reads from the PTY in the same
goroutine as the tool function:

```go
f, err := pty.StartWithSize(cmd, &pty.Winsize{...})
c.Job.Attach(cmd.Process, f)

for {
    n, rerr := f.Read(buf)     // blocks until output or close
    if n > 0 { c.Job.Write(...) }
    if rerr != nil { break }
}
cmd.Wait()                      // blocks until process exits
```

For a short command (go run, which dlv), the process exits and the function
returns. For an interactive process (dlv), the function blocks forever on
`f.Read()` because the process never exits.

The `ai_callback_pattern` matching happens in `j.Wait(limits)`, which the
engine calls AFTER the tool function returns. But the tool function never
returns for interactive processes. So `j.Wait()` is never called, and the
pattern `\(dlv\) ` is never checked against the accumulated output.

**The fix:** The PTY read loop must run in a SEPARATE goroutine. The tool
function should start the reader goroutine, then return immediately (or
after a brief delay). The engine then calls `j.Wait(limits)` with the
`ai_callback_pattern`, which polls `j.out` for the pattern. When matched,
the engine returns the result to the LLM with a handle for `send_input`
and `wait_for_job`.

This is the same architecture CodeRhapsody uses (learning #38: "dual PTY
for dlv, reset pattern scan per wait, normalize CRLF across read
boundaries"). The ensemble agent's simpler synchronous design worked for
chapters 1-13 but breaks on interactive processes.

### Fix Applied (commit b139325)

Three changes:
1. **`common.Call.DeferFinish`**: New bool field. When a tool sets this, the dispatcher leaves the job Running instead of calling Finish — the reader goroutine calls Finish itself when the process exits.
2. **`toolRunCommand`**: Rewritten to spawn PTY reader in a goroutine and return immediately. Sets `c.DeferFinish = true`. The goroutine reads from the PTY, writes to the job, waits for process exit, records exit code, calls `job.Finish`.
3. **Two dispatch paths updated**: Both `Engine.Execute` (ch3/ch4 backward-compat) and `Actor.dispatchTool` (ch6+) check `c.DeferFinish` before calling `job.Finish`.

Key debugging discovery: `Engine.Execute` has its own dispatch goroutine, separate from `Actor.dispatchTool`. The ch3/ch4 backward-compat `eng.Ask()` method routes through `Engine.Execute`. Both dispatch paths needed the fix. All graders pass (ch2-13 100/100, ch5 120/120).

## Cost

- Virtual user run 1 (exploratory, no gui_submit): ~$5
- Virtual user run 2 (Tower of Hanoi attempt): ~$5
- Coding agent API calls for the successful Tower of Hanoi: ~$3
- Coder sub-agent for ch14 infrastructure: ~$24
- Total ch14 cost so far: ~$37

## What Remains

1. ~~Fix interactive process support (PTY reader in goroutine)~~ DONE (b139325)
2. Re-run virtual user — dlv session should now complete
3. Fix GUI bugs found (input validation, content truncation, etc.)
4. Write chapter prose
5. Commit all changes
