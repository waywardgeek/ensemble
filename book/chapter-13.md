# Chapter 13: The Agent Sees Itself

A coding agent that cannot see its own GUI is debugging blind. Every tool so far has operated on files, processes, and network responses. The GUI is a black box the user stares at while the agent types into it. This chapter closes that gap. A single skill connects the agent to its running browser interface through the MCP infrastructure from Chapter 12, and the agent begins seeing what the user sees: the DOM, the buttons, the text being spoken aloud.

The wiring is short and mostly mechanical, and it occupies the first half of the chapter. The second half reports what arrived once it worked, when the capability was pointed at a GUI nobody had audited and the human supervising it agreed to stay quiet.

## TL;DR

A `gui-debug` skill activates browser MCP tools via the skill system from Chapter 10 and the MCP infrastructure from Chapter 12. Loading the skill triggers an MCP handshake, discovers four tools, and registers them. Two are ephemeral (auto-injected every round), two are callable.

### The skill

```yaml
---
name: gui-debug
description: Debug the GUI through browser MCP tools
depends:
  - ensemble
mcp_servers:
  - name: browser-debug
    transport: stdio
    command: ./grader
    args: ["--fake-mcp"]
---
```

The `mcp_servers` field declares a server to connect when the skill loads. The `transport: stdio` entry spawns the command as a subprocess and speaks JSON-RPC 2.0 over its stdin/stdout. For grading, the command points at the grader binary itself running in fake MCP server mode. In production, the transport would be `websocket`, tunneling through the hub to the browser.

### Skill-based MCP lifecycle

Loading a skill with `mcp_servers`:

1. For each server entry, create the transport (`StdioTransport` for `stdio`, `WSTransport` for `websocket`).
2. Create an MCP `Client`, call `Initialize`, then `ListTools`.
3. Bridge discovered tools into the agent's registry via `mcp.Bridge`.
4. Store the client handle for cleanup.

Unloading the skill:

1. Remove bridged tools from the registry via `RemoveTool`.
2. Close the MCP client.
3. Close the transport (kills the subprocess for stdio).

Two callbacks on the tool registry wire this lifecycle:

```go
type Reg struct {
    // ...
    onSkillMCPConnect    func(skill string, servers []common.MCPServerConfig) ([]string, error)
    onSkillMCPDisconnect func(skill string)
}
```

The `load_skill` handler calls `onSkillMCPConnect` after loading the skill's tools. The `unload_skill` handler calls `onSkillMCPDisconnect` before removing them. The callbacks live in `cmd/main.go` where the MCP client, transport factories, and registry are all in scope.

### RemoveTool

The registry gains a `RemoveTool(name)` method. Chapter 12 added tools dynamically via `RegisterTool`; this chapter removes them dynamically when a skill unloads. The tool is deleted from the map and its declaration is removed from the cached list. The `onToolsChanged` callback fires so the engine picks up the new tool set.

### The four browser tools

These are the same four tools from Chapter 12's `mcp.js`, now activated through the skill system:

| Tool | Ephemeral | What it returns |
|------|-----------|-----------------|
| gui_snapshot | round | Markdown DOM: panes, interactive elements with selectors, artifacts |
| tts_queue | round | JSON array of pending TTS utterances |
| gui_click | no | Confirmation: "clicked #button-A" |
| gui_input | no | Confirmation: "set text on #search-box" |

The ephemeral tools are called automatically by the engine before each round. Their output appears in `Context.Ephemera`. The LLM sees the DOM snapshot and TTS queue as context data, refreshed every round, without having to ask for it.

The callable tools (`gui_click`, `gui_input`) appear in the LLM's tool declarations. The LLM decides when to click or type based on what the snapshot shows.

### --gui-debug flag

A convenience flag that auto-loads the `gui-debug` skill on startup:

```
./ensemble --gui-debug --skills-dir ./skills
```

Equivalent to the agent calling `load_skill("gui-debug")` as its first action. The `--skills-dir` flag overrides the default skills directory, which is useful for grading with temp directories containing test fixtures.

### Exercise

```
./ensemble --gui-debug --skills-dir SKILLS_DIR
```

The grader binary doubles as a fake MCP server (`--fake-mcp` flag). It writes a temporary `gui-debug` SKILL.md whose `command` points at itself. When the student binary loads the skill, it spawns the grader as a subprocess, establishing a JSON-RPC channel.

The fake MCP server maintains state: `gui_snapshot` returns "button-A: enabled" initially, then "button-A: disabled" after a `gui_click` call. This simulates the DOM changing in response to interaction, and the grader verifies that the next round's ephemeral injection reflects the updated state.

Seven checks:

| Check | Points |
|-------|--------|
| skill-loads | 15 |
| ephemeral-injected | 20 |
| snapshot-updates | 15 |
| tts-visibility | 10 |
| gui-interaction | 15 |
| skill-unload | 15 |
| ch12-parity | 10 |

```
make grade13
```

---

## What the wiring found

Everything above this line is plumbing, and most of it is close to trivial. A skill declares an MCP server. The registry connects when the skill loads and disconnects when it unloads. A browser tab exposes four tools over a WebSocket that was already there. The interesting part is not the transport. The interesting part is what arrives the first time an agent can see a screen it did not render and press a button nobody told it about.

What follows is one week of that, reported as it happened. Every bug in this section was found by pointing the machinery at a GUI nobody had audited. Every fix is in the repository, and every claim below was checked against the code rather than against the notes written at the time. Two of the notes turned out to be wrong, which is its own lesson and is recorded where it belongs.

### The user that cannot read the source

The driver is a separate binary of 122 lines. It connects to the same hub the browser connects to, discovers its tools through MCP, and registers exactly one tool of its own, `file_report`, so it has somewhere to put conclusions. It is built on `NewBareAgent`, which is an agent with no builtin tools at all. No `read_file`, no `run_command`, no editor. Everything it can do arrives through discovery.

Its system prompt states the constraint directly:

```
You CANNOT read files, edit code, or run commands. You can only
interact through the GUI, exactly as a human user would.
```

That restriction is the entire value of the thing. An agent with filesystem access will answer a question about the interface by reading the source, which tells you what the interface was meant to do. An agent holding only `gui_snapshot`, `gui_click`, `gui_input` and `tts_queue` has to answer from the screen, which tells you what the interface actually does. The gap between those two answers is where the bugs live.

### The human who watched and said nothing

The protocol around these runs matters as much as the tooling, because it is what makes the results mean anything.

Bill Cox supervised every run in this section in real time, reading the reasoning as it streamed. He had written most of the GUI, so he knew where the weak joins were. He deliberately did not say. When the driver walked past a defect he could see, he let it walk past, and when it found one he already knew about, he let it report the discovery as news.

He did direct, and the distinction is worth drawing precisely. Direction covered what to test next, which binary to rebuild, and once, usefully, the observation that a stale server was still holding port 8084 while the freshly built binary bound to nothing and served no one. That is operational guidance, and withholding it would have wasted an hour proving nothing. Findings were different. No bug in this section was pointed out by the human before the machinery found it.

The reason to run it that way is that a supervisor who volunteers the answer cannot tell the difference between a tool that works and a tool that agrees. An agent handed a hint will confirm it, write a plausible account of confirming it, and leave no trace that the hint did the work. The only way to learn whether `gui_snapshot` is sufficient to find a real defect is to watch someone try to find a real defect with `gui_snapshot` and nothing else.

This is the same discipline the book asks of a reader supervising an agent on live code, applied to the machinery itself. Autonomy is worth measuring only where the human was genuinely silent.

### The observer that froze what it watched

The first runs never reached a task at all. The driver typed its prompt and the coding agent stopped responding, permanently, before any work began.

The cause was a feature working exactly as designed. Chapter 8 pauses the engine while the user is typing, so that an agent does not barrel ahead while a human is halfway through composing a hint. The feature exists for a particular working style: Bill Cox reads the agent's reasoning as it streams and sends corrections mid-turn, which makes the pause the difference between a hint that lands and a hint that arrives after the decision it was meant to change. The implementation watches the input field:

```javascript
// The original handler. WebSocket readiness guards elided for clarity.
let userTyping = false;
input.addEventListener('input', () => {
    const typing = input.value.trim().length > 0;
    if (typing && !userTyping) {
        userTyping = true;
        ws.send(JSON.stringify({type: 'pause'}));
    } else if (!typing && userTyping) {
        userTyping = false;
        ws.send(JSON.stringify({type: 'unpause'}));
    }
});
```

Read that as a human and it is correct. Text appears, the agent waits. The field empties, the agent resumes.

Read it as a program driving the field and it is a trap with no exit. `gui_input` sets the value and dispatches an `input` event, which is what makes it indistinguishable from typing, and that indistinguishability is the entire point of the tool. So the pause fires. Then `gui_submit` sends the prompt and clears the field by assigning to `value`, and assigning to `value` in JavaScript fires no event at all. The branch that sends `unpause` is reachable only by a path the driver cannot take. One keystroke in, the engine is paused, and nothing in the system will ever unpause it.

The repair is two lines of intent. A flag marks input as programmatic so the pause handler ignores it, and submission sends an explicit `unpause` rather than relying on an event that will not arrive:

```javascript
if (window._mcpProgrammaticInput) return;
```

The general shape of this is older than software. An instrument that shares a channel with the thing it measures will perturb it, and the perturbation is worst where the system was tuned to human timing. Pause on typing, debounce on scroll, idle timeouts, animations that wait for a settle: each one encodes an assumption about how fast and how continuously a person acts. An automated driver violates all of them at once, and it does so silently, because from the GUI's perspective nothing unusual occurred. A user started typing and never stopped.

### A debugger that would not start

The first task given to the driver was small on purpose: ask the coding agent for Tower of Hanoi, then watch it debug the result. The agent wrote the program, ran it, and produced fifteen moves. Then it started `dlv` and stopped forever.

Chapter 4 predicted this in print. Section 4.8 is titled "Your agent drives a debugger," and it warns that a dispatcher which waits for process exit will hang on an interactive program, because "the dispatcher would wait for an exit that never comes, and on pipes `dlv` refuses to start at all." The chapter named the trap. The reference implementation walked into it anyway.

The mechanism was a single synchronous read. `toolRunCommand` read the PTY inside the tool function, so the function could not return until the process closed its output. For `go build` that is correct and invisible. For a debugger sitting at a prompt it is fatal, and it is fatal in a way that disables the exact feature meant to rescue it: `ai_callback_pattern` is matched inside `Wait`, and `Wait` runs after the tool function returns. The escape hatch was behind the door it was supposed to open.

The fix is a lifetime change rather than a plumbing change. The tool function now returns once the process is running, and a reader goroutine owns the PTY from that moment until exit. A new field, `DeferFinish`, tells the dispatcher that the job will finish itself:

```go
// The reader goroutine copies output from the PTY to the job until the
// terminal closes, then reaps the process and finishes the job. It runs
// independently of the dispatcher, so an interactive process that never
// exits (dlv, python) does not block the model from getting a handle.
job := c.Job
c.DeferFinish = true
go func() {
    buf := make([]byte, 32*1024)
    for {
        n, rerr := f.Read(buf)
        if n > 0 {
            _, _ = job.Write(bytes.ReplaceAll(buf[:n], []byte("\r\n"), []byte("\n")))
        }
        if rerr != nil {
            break
        }
    }
    _ = f.Close()
    werr := cmd.Wait()
    // ... exit code extraction elided ...
    job.Finish("", nil)
}()
```

### Two dispatchers

The fix went in, the debugger worked, and every behavioral check in chapter 4 dropped to zero out of a hundred.

The failure signature pointed in the wrong direction. Tool output came back as the empty string, and the capture file on disk held zero bytes. That reads as a broken output path, and an hour can be spent there. The actual fault was lifetime again. There are two tool dispatchers in the codebase: `Engine.Execute`, which serves chapters 3 and 4 through the older `Ask` path, and `Actor.dispatchTool`, which serves chapter 6 onward. Only the second had been taught about `DeferFinish`. The first still called `job.Finish` the instant the tool function returned, and `Finish` closes the capture file. The reader goroutine was writing correctly the whole time, into a file that had already been closed underneath it.

The diagnostic detour is worth recording, because it cost more than the bug. Debug prints added to `Actor.dispatchTool` never appeared in the output, and the first explanation reached for was a stale build cache. It was not stale. The prints were in a function the failing tests never called. Silent debug output is evidence of the wrong code path at least as often as it is evidence of a bad build, and the cheaper check is to confirm which function runs before rebuilding anything.

### Ninety-eight seconds

With both dispatchers corrected, the same task ran end to end. The driver typed the prompt into the real text field and submitted it. The coding agent wrote `hanoi.go`, ran it for fifteen moves, then opened `dlv` and worked through a full session: set a breakpoint on `move`, continue, print `disk`, quit. The debugger exited zero. Total elapsed time was ninety-eight seconds, and no human touched the keyboard between the prompt and the report.

That run is the capability the chapter has been building toward. An agent that can drive a debugger can inspect a running program's state rather than reasoning about what the state ought to be, and an agent that can drive a GUI can do the same for an interface.

### The bug that only appears after you fix the bug

The next task sent the driver into the settings panel with instructions to try values a careful user would not. It set temperature to negative five, and the server accepted it.

That is the obvious bug, and the fix is ordinary: a `clamp` function, called both when a settings patch arrives and when settings load from disk, so a hand edited file cannot bypass validation either. The tests were checked by neutering `clamp` and confirming that five of six failed.

The second bug appeared only because the driver was asked to repeat its own test after the fix landed. The server now clamped correctly, and the GUI still displayed negative five. Server state and screen state disagreed, and the screen is what a user believes.

The cause was a single struct doing two incompatible jobs. `Settings` was used both as a sparse patch, where an absent field means "leave this alone," and as a full state broadcast, where every field should be present. Those two roles want opposite JSON encodings. With `omitempty` on every field, a value clamped to zero vanished from the broadcast entirely, and the client guard reads:

```javascript
if (s.temperature !== undefined) { /* update the input */ }
```

An absent field is indistinguishable from an unchanged field, so the stale value stayed on screen. The same flaw meant `TTSEnabled: false` could never be transmitted, because false is empty. Turning speech off was unrepresentable on the wire.

The deleted code had confessed. `SettingsStore.Apply` carried this comment:

```go
// Since omitempty skips false, we handle this via the raw patch.
// For simplicity, always apply.
```

Someone met this bug, understood it precisely enough to describe it in one sentence, and routed around it instead of removing it. The method had no callers outside tests.

There was a real choice about how far to take the repair. The cautious option preserves the existing wire format and introduces a second type for broadcasts, leaving the patch struct untouched, which costs nothing today and leaves two nearly identical structs for the next reader to confuse. Bill Cox settled it in one line: the format has no external consumers, so make the code clean. The repair therefore dropped `omitempty` from every field, so a broadcast always carries complete state, and deleted `Apply` outright, leaving one merge path with defined semantics.


Measured evidence that the clamp runs, taken from the settings file after the run: `max_tokens` held 1000000 and `tts_speed` held 10, both exactly the ceilings, and `temperature` was absent because zero no longer serializes.

### The settings panel that does not exist

The strongest finding came from a test that failed in an unusual way.

Asked to exercise the sidebar tabs, the driver reported that clicking "Artifacts" opened the Settings panel. The report was specific, confident, and false. There is no Settings tab in the markup. The `data-tab` and `data-panel` attributes map correctly, chats to chats and artifacts to artifacts, and clicking either one does what it says.

The report was fiction, and the reason it was fiction is the finding. Tab state lived only in a CSS class. `gui_snapshot` built its element labels from `textContent`, `aria-label` and `placeholder`, and read no state attributes at all, so the driver could see that two tabs existed and could not see which one was selected. It had also reached for an ambiguous selector, `button:nth-of-type(2)`, because the buttons carried nothing better to aim at. Asked what happened after the click, it had no way to observe the answer, and it produced a plausible one.

An agent denied state does not report uncertainty. It fabricates.

The repair was to put state where a machine can read it. The hamburger and the tabs now carry `aria-expanded` and `aria-selected`, synchronized on every click, with the initial value derived from the live CSS class rather than hardcoded, because the markup default had already drifted from the rendered default. `gui_snapshot` now reports those attributes alongside each control.

Rerunning the identical test produced a correct before and after table, and in the one place where the driver lacked information, it wrote "not observable" instead of inventing a panel. Same model, same prompt, same GUI. The only change was that the interface stopped hiding its state.

This is the accessibility argument in a form that a developer who has never used a screen reader can feel directly. Semantic state attributes are usually presented as a courtesy extended to users with assistive technology. They are also the difference between an automated observer that reports what happened and one that reports something reasonable. An unobservable interface does not produce no data. It produces wrong data, delivered with the same confidence as the truth.

### What an interface owes an observer

Two smaller findings came out of the same pass, and both generalize.

The first was an injection hole. `artifact-scroll.js` built the tool card header by interpolating the tool name and its serialized arguments into `innerHTML`. Tool arguments are not authored by the agent. They contain filenames, URLs, and the contents of files just read, which is to say text an attacker can influence. Twenty lines further down, the result path rendered with `textContent` and was safe. One file, both patterns, and the vulnerable one sat on the path that renders attacker adjacent text. The book spends a chapter on prompt injection arriving through the model. This was the same threat arriving through the renderer.

The second was truncation, and the notes about it were backwards. The working document asserted that the driver saw the full DOM while the human saw a trimmed version. Measurement found four independent caps running the other way: the human view truncated tool input at 500 characters and results at 1000, while the snapshot truncated each artifact at 120 characters and the whole document at 4000. The observer saw roughly a tenth of what the human saw. Worse, the snapshot silently dropped every artifact past the tenth with no marker, so it could not distinguish ten artifacts from fifty.

Nobody had measured any of this before writing it down. The caps were kept, because both views need them, and both were made honest: truncated text is now reachable through a `title` attribute rather than deleted behind an ellipsis, and the snapshot states how many characters and how many artifacts it withheld.

That is the rule the section converges on. An observer that truncates in silence will report that a screen looks fine when it never saw the screen. A control that keeps its state in a CSS class will be guessed about. A renderer that trusts its inputs will execute them. None of these are failures of the model doing the observing, and none of them were visible from the source, which is why it took a user who could not read the source to find them.

### Every one of them was a documentation bug

The instruction that produced this section came from Bill Cox once the fixes were in: update the chapter summaries wherever a bug made it through. Tracing them was the first step, and each defect was matched to the chapter that should have prevented it. All five arrived at the same place.

Chapter 4 explained the job model, named the interactive process trap in plain words, and never stated the obligation that follows from it. Chapter 8 built the tool card and showed the rendering without saying which parts of it carry text the agent did not write. Chapter 12 specified `gui_snapshot` with a 4KB cap and did not require the cap to announce itself. In each case the mechanism was taught correctly and the duty attached to the mechanism was left implicit.

Chapter 9 is the sharpest example, because the reasoning that let three bugs through is stated plainly in its opening:

> The graded surface is small: four server-side settings checks and a parity gate. Everything else is client code you can see working.

Chapter 9's GUI is deliberately ungraded, on the argument that a human looking at a screen is a sufficient test. Looking at a screen confirms that a value was accepted. It does not confirm that the value was validated, that the server and the screen agree about it afterward, or that a control's state can be read by anything other than an eye. Three bugs fit in that gap, and all three shipped.


A student following those chapters would have written the same code. That is the test for whether a defect belongs to the implementation or to the book, and every one of these failed it. The repair was therefore made in the prose as well as the source. Chapter 4 now states that the tool returns when the process is running and the job owns it from there, and gives the failure signature, which reads like broken output capture and is really a closed file. Chapter 8 now states that tool cards are built with `textContent` and that capped content stays reachable. Chapter 9's summary gains the three obligations that survive the absence of a grader. Chapter 12 now requires the snapshot to report what it withheld.

There is a loop closing here that is worth naming, because it is the reason this book exists in the form it does. The agent described in these chapters was built by following these chapters. When it acquired eyes and a way to press buttons, the first thing it did was find places where the chapters were wrong. The bugs were in the GUI, and the GUI was correct with respect to the instructions it was built from, so the instructions were what needed editing.

A book that produces a working program gets to be tested by the program it produces. This section is the first time that test came back with findings, and the findings were about the book.

### The same failure, one level up

The draft of this section quoted Chapter 9 as saying "You know the client works because you can see it." That sentence appears nowhere in Chapter 9. It was invented while the paragraph was being written, it was a fair paraphrase of the chapter's actual argument, and it sat on the page with exactly the same confidence as the sentences around it that were true.

It was caught by `grep`, not by rereading. The check took two seconds and consisted of searching the source file for the words about to be printed inside quotation marks. The real sentence, once located, was better than the invention and made a sharper point, which is the usual result.

The parallel to the phantom Settings panel is exact. Neither fabrication came from carelessness, and neither would have been prevented by trying harder. Both came from a gap between what was needed and what was observable, filled with something plausible. The driver could not see which tab was selected, so it produced a reasonable answer. The draft had the chapter's argument available and not its wording, so it reconstructed one. The remedy in both cases was mechanical and took seconds: read the state attribute, grep the file.

Advice does not survive this failure mode. "Be careful with quotations" is guidance that a confident generator will sincerely believe it has followed. "Grep for the quoted string before it reaches the page" either happened or it did not, and the difference is visible in the shell history.


