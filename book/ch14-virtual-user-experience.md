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

**Status:** FIXED (ce32c27). Validation lives in one `clamp()` called from
both `ApplyRaw` and the load from disk. Guarding only the WebSocket patch
would have left `settings.json` open, and files get hand-edited.

**The second bug underneath it.** Re-running the virtual user against the
fixed server produced a more interesting result: the server clamped the
value correctly, and the GUI *still displayed -5*. The persisted file
proved the clamp had worked (`max_tokens` sat at exactly the new ceiling,
`temperature` was gone entirely). The client never learned about the
correction.

The cause was `omitempty` on every field of `Settings`. Clamping -5
produced 0, `omitempty` dropped the zero from the broadcast, and the
client's `if (s.temperature !== undefined)` guard skipped the field. The
server said "I set this to zero" and the wire format turned it into "I
said nothing about this."

The root cause was one struct serving two roles with opposite needs: a
sparse *patch*, where a missing key means "leave it alone", and a full
*state snapshot*, where zero is a real value. `omitempty` is right for the
first and silently corrupting for the second. The fix drops `omitempty`
entirely, so a snapshot always carries all ten fields, and deletes the
unused `Apply` whose zero-means-unset merge was the same mistake in method
form. Patches are already expressed by key presence in `ApplyRaw`.

The deleted code had confessed in its own comment: "Since omitempty skips
false, we handle this via the raw patch. For simplicity, always apply."
Someone met this bug for the `tts_enabled` bool, worked around that one
field, and left the general case in place.

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

The first symptom of the missing `Engine.Execute` fix was misleading: every
ch4 behavioral check reported `Output: ""` and `0 bytes` on disk. The job
was being finished by the dispatcher the instant the tool function returned,
which closed the output file before the reader goroutine had written a
single byte. The empty output looked like a capture bug; it was a lifetime
bug.

### Verified End to End

The virtual user re-ran the scenario against the live GUI and the interactive
dlv session completed. Sequence observed:

1. Virtual user typed the task into the prompt box and submitted it
2. Coding agent wrote `/tmp/hanoi/hanoi.go` and ran it (7 moves, 3 disks)
3. Coding agent started `dlv debug` with `ai_callback_pattern` on the `(dlv)` prompt
4. `run_command` returned a handle immediately instead of blocking
5. `send_input` drove `break move`, `continue`, `print disk`, `quit`
6. dlv exited with code 0

Total run time 98 seconds. The virtual user filed a report confirming the
session worked, and noted that `wait_for_idle` plus `tts_queue` gave it a
reliable way to tell when the coding agent had finished a turn.

This is the payoff of the job model from chapter 4. The tool did not need a
new interface, a new tool, or a special case for debuggers. It needed to stop
conflating "the tool function returned" with "the process finished." Once
those two events were separated, `send_input`, `wait_for_job`, and
`kill_job` all worked on an interactive process exactly as they already
worked on a batch one.

## Cost

- Virtual user run 1 (exploratory, no gui_submit): ~$5
- Virtual user run 2 (Tower of Hanoi attempt): ~$5
- Coding agent API calls for the successful Tower of Hanoi: ~$3
- Coder sub-agent for ch14 infrastructure: ~$24
- Total ch14 cost so far: ~$37

## What Remains

1. ~~Fix interactive process support (PTY reader in goroutine)~~ DONE (b139325)
2. ~~Re-run virtual user — dlv session should now complete~~ DONE, verified end to end
3. ~~Input validation on settings~~ DONE (ce32c27), plus the `omitempty` bug it exposed
4. Remaining GUI bugs: tool card truncation, hamburger toggle not in DOM
5. Write chapter prose
6. Commit all changes

## Lessons for the Author: TL;DR Updates

Every bug below was found by the virtual user driving the real GUI. Each
one is also a teaching failure: the chapter that introduced the mechanism
did not say the thing that would have prevented it. A student LLM building
from these chapters will write the same bug unless the TL;DR says
otherwise, because the TL;DR is the part it reads most carefully.

These are notes for the author, not prose. Each entry names the target
chapter, the property to state, and why the obvious phrasing is not enough.

### Ch4 (jobs): a tool function returning is not the process finishing

**The bug.** `run_command` read the PTY to EOF inside the tool function.
For `dlv`, which never exits on its own, the function blocked forever. The
agent could start a debugger and then never speak to it again.

**Why the chapter invited it.** Ch4 teaches that a tool returns its output.
That is true for `ls` and false for every interactive process, and the
chapter never marks the boundary. Worse, `ai_callback_pattern` is matched
inside `Wait()`, which the dispatcher calls *after* the tool function
returns. So the blocking tool function makes the chapter's own
pattern-matching feature unreachable. The feature and the bug cannot
coexist, and nothing says so.

**State in the TL;DR.** A tool that starts a process returns as soon as the
process is *started*, not when it finishes. Output is not a return value;
it is a stream the job accumulates. A reader goroutine owns the PTY, writes
into the job, and calls `Finish` when the process exits. Whoever owns the
finish must be the only one who calls it, so a tool that defers the finish
must say so (`DeferFinish`) and the dispatcher must honor it.

**The trap worth naming.** If the dispatcher calls `Finish` when the tool
function returns, and the tool has handed the work to a goroutine, `Finish`
closes the output file while the reader is still writing. The symptom is an
empty result and a zero-byte file, which reads like a capture bug and is
really a lifetime bug. That misdiagnosis cost real time here.

### Ch6 (actors): a second dispatcher is a second contract

**The bug.** Fixing the actor's dispatch path left every ch4 behavioral
check at 0/100, because ch3/ch4 run through `Engine.Execute` and ch6+ runs
through `Actor.dispatchTool`. Two goroutines, the same job protocol,
independently written.

**State in the TL;DR.** When the actor takes over dispatch, it inherits the
job contract rather than restating it. If both paths must exist for
backward compatibility, the job lifecycle rules apply identically to both,
and a change to one is incomplete until the other matches.

**Debugging lesson worth printing.** Debug output that never appears does
not always mean a stale build. It can mean the code is not on the path
being exercised. Check for a second implementation before blaming the build
cache.

### Ch9 (settings): validate at every mutation path, not just the network one

**The bug.** `temperature: -5` was accepted and forwarded to the model API.

**State in the TL;DR.** Settings arrive from more than one direction: the
WebSocket patch and the file on disk. Validation belongs in one function
called from both. A check that lives only in the message handler is a check
a hand-edited `settings.json` walks straight past.

**Design detail worth stating.** Out-of-range is not a preference, it is a
bug or an attack, so the server corrects rather than rejects. Negative
values collapse to zero, meaning unset, so the default applies; a client
sending -5 has said nothing about what it wants. Small positive values rise
to the minimum instead, because a font size of 3 is a real request that
happens to render nothing.

### Ch9 (settings): a state snapshot and a patch want opposite JSON

**The bug, and the best one the virtual user found.** After validation
worked, the server clamped -5 to 0 and the GUI went on displaying -5.
`omitempty` dropped the zero from the broadcast, and the client's
`!== undefined` guard skipped the field. The server said "I set this to
zero"; the wire format said "I mentioned nothing."

**State in the TL;DR.** One struct cannot be both a sparse patch and a full
state snapshot. A patch means "change the keys I sent", so absence is
meaningful and `omitempty` is correct. A snapshot means "this is
everything", so zero is a real value and `omitempty` is silent corruption.
Express patches by key presence. Never use zero as a sentinel for "unset"
in a type that is also broadcast as authoritative state.

**Why this generalizes.** The same shape breaks any boolean: `omitempty`
skips `false`, so "TTS is off" is unsendable. The code being replaced had
already hit this and patched the single bool with the comment "for
simplicity, always apply", leaving the general case. A workaround on one
field is evidence the model is wrong, not evidence the field is special.

### Ch13 (MCP in the browser): the page must actually load the server

**The bug.** `mcp.js` was never referenced from `index.html`, so the MCP
handshake hung with no server on the other end.

**State in the TL;DR.** A browser-side MCP server only exists once the page
loads it. Say that the script tag is part of the deliverable, and give the
handshake a visible failure rather than an indefinite wait, since "hangs
forever" is the least diagnosable symptom available.

### Ch8 (pause gate): programmatic input is not a human typing

**The bug.** `gui_input` dispatched the events a human keystroke produces,
which tripped the typing pause gate and paused the engine permanently.

**State in the TL;DR.** The pause gate exists so the agent does not talk
over a human mid-sentence. Synthetic input from an automated driver must be
distinguishable from human input, or the GUI deadlocks precisely when
something is testing it. Any "is the human busy" heuristic needs an
explicit way to say "this was not the human."

### Ch9 (GUI): the DOM is the accessibility surface and the test surface

**The bugs.** Tool card content is truncated, so the human sees less than
the virtual user reads. The hamburger toggle's state is invisible in a DOM
snapshot.

**State in the TL;DR.** State expressed only in CSS or in a JavaScript
variable is invisible to a screen reader and to any automated observer.
Toggles carry `aria-expanded`; panels carry their state in attributes. This
is one requirement, not two: what makes the GUI usable by speech is exactly
what makes it observable by a virtual user. A GUI that cannot be read
cannot be tested.

### The meta-lesson for the chapter

The virtual user found bugs that graders could not, because graders check
the contract the author thought to write down. Every bug here lived in the
gap between components that were each individually correct: a tool function
and a dispatcher, a clamp and a wire format, a pause gate and an automated
driver. Integration bugs are the residue of correct parts, and driving the
real interface is how they surface.

