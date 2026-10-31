# Chapter 12: The Agent Sees Its Working Interface

Chapter 11 connected external tools to the Agent through MCP. This chapter makes
one of those capabilities convenient at startup: an optional `gui-debug` skill
lets the Agent inspect and operate the Ensemble page in which the conversation is
happening. It does not give the Agent pixel vision, continuous observation, or
access to every browser tab. It supplies a particular connected page's tools and
a manual that explains how to use them honestly.

The distinction between *having a tool* and *having an observation* matters. A
loaded browser skill does not mean the Agent has seen the page. The Agent calls
`gui_snapshot` when it needs a structured view and `tts_queue` when it needs the
speech queue. It calls again after an action to check the result. A successful
click response reports dispatch; it does not prove the intended visible change.

The implementation builds on Chapter 11 rather than adding another GUI engine,
transcript, browser controller, or scheduler. The new library vocabulary is an
optional ordered startup selection. The GUI flag is a convenience for choosing
that selection. The actual browser still supplies its catalog through MCP.

## Startup Selection Is Configuration, Not a Conversation

The value-like `SkillSpec` record now has a final optional member:

```csharp
public sealed record SkillSpec(
    string Directory,
    string Primary,
    IReadOnlyDictionary<string, string>? Variables = null,
    IReadOnlyList<string>? StartupSkills = null);
```

Existing callers remain source-compatible. The library reads no environment
variables. `Skills` copies the selection into an immutable array at construction,
just as it snapshots initial variables. Mutating a caller's list later cannot
change a startup already in progress.

For an already constructed Agent with its ordinary local tools installed, the
library setup is:

```csharp
var skills = new Skills(agent, new SkillSpec(
    skillDirectory,
    "ensemble",
    StartupSkills: ["gui-debug"]));

// Application hosts load persistence and open diagnostics before this step.
// A GUI host also starts listening and supplies the browser registration first.
await skills.InitializeAsync();
```

Do not manufacture a model `load_skill` call to implement this. Startup is host
configuration. There is no provider response or tool-call ID to pair with an
invented result. The existing startup mailbox operation returns reliable completion
without appending assistant dialogue or a host tool observation.

`Common/Skills.cs` declares the shared vocabulary and frozen selection.
`Skills/Startup.cs` resolves it; `Skills/Connections.cs` prepares and commits remote
connections; `Agent/Skills.cs` carries completion through the existing actor. All
parts are ordinary partial types in the same assembly. There is no separate
startup service or copied list of active capabilities.

## Validate the Whole Initial Batch

Resolution begins with the primary and its dependencies. It then visits startup
names in the host-supplied order. Each must name a loadable skill currently offered
by the prospective active set. Thus an earlier selection may offer a later one.
A repeated name, or a selected loadable skill already included as a dependency,
is activated only once. Missing names, wrong skill types, unavailable selections,
and dependency cycles fail before publication.

The prospective set is temporary algorithm state, not an authoritative registry.
History's immutable `SkillPart` values remain the source of active instructions
and grants. The same resolver is used during initial validation and again at
commit, so the prepared closure cannot silently differ from the published one.

For an entirely local closure, construction publishes one dependency-first
activation. If *any* initial skill needs MCP, construction publishes no partial
primary and grants no provisional tools. `InitializeAsync` uses the existing
asynchronous preparation path, validates catalogs and collisions, renders all
manuals, and commits one activation on the actor. Reverse calls remain responsive
but cannot borrow unpublished authority. A server failure destroys prepared
clients and leaves the batch unpublished.

`Ready()` derives from the primary and selected names in History. It does not read
a second mutable ready flag. `RequiresBrowserStartup()` examines the complete
initial closure, not only the primary. A stdio-only optional skill must not wait
for a browser. A WebSocket-backed optional skill must not be declared ready before
its actual page has initialized.

Concurrent initialization callers join the existing task. Cancelling one observer
detaches that observer; Agent shutdown cancels and joins startup work. A failed
startup does not arm shutdown saving over an earlier good session. Restoring a
conversation does not restore permission to execute old tools: the current
startup configuration rebuilds instructions, grants, and connections as described
in Chapters 10 and 11.

## Instructions Must Reach the Provider as Instructions

Loading a manual is not enough if the next request carries it only as tool-result
text. The C# renderer already places active skill bodies in provider instruction
context: Anthropic's `system`, OpenAI Chat Completions' first system message, and
Gemini's `systemInstruction`. Startup uses that same representation. The primary
comes first; each active manual appears once, not again as ordinary dialogue.

This endpoint implements Chat Completions, not OpenAI Responses. A future Responses
adapter must explicitly map these instructions to its instruction field. Evidence
for one adapter cannot establish another adapter's behavior.

`$TOOLS` expands against the prospective allowed catalog. Automatic `per_turn_`
and `per_round_` tools remain excluded from all model-facing declarations,
discovery references, and expansions. Browser debugging does not use those
automatic extensions at all. An unrelated model round receives neither an
unsolicited DOM snapshot nor a new speech-queue dump.

## The Application Boundary

Both hosts accept a comma-separated `--startup-skills` value, with
`EN_STARTUP_SKILLS` as the application-level default. An explicit flag overrides
the environment value. Empty comma-separated fields are ignored. Both hosts also
accept `--gui-debug`, which appends `gui-debug`; duplicate selections activate once.
Without either setting, the default primary and optional-skill behavior are unchanged.

For the actual GUI:

```sh
dotnet run --project Examples/Gui/Gui.csproj -- --port 8097 --gui-debug
```

Use an available owned loopback port and an isolated profile for exercises, not a
user's working browser or server. Provider settings are the same application
settings used in earlier chapters. This flag grants no provider budget or device
permission.

The GUI constructs the Agent, loads persistence, starts logging and the HTTP
listener, and waits for a browser only when the initial closure requires one.
The page registers through the existing `/ws` MCP envelope channel. Readiness
still gates user work until activation commits. There is no separate `/mcp` route,
static fallback catalog, or guessed description for an unavailable remote tool.

The CLI can use the same startup API with a supplied stdio-backed `gui-debug`
manual, as the executable tests do. The shipped manual uses the GUI's WebSocket
registration; the plain CLI does not magically create a browser or GUI listener
for it. Without that transport, startup fails explicitly rather than claiming
browser capability.

## One Skill Owns Its Browser Capability

The shipped `gui-debug` manual owns its own WebSocket server declaration and the
five explicit grants: `gui_snapshot`, `gui_click`, `gui_input`, `tts_queue`, and
`tts_say`. It is an alternative to the general `browser` skill, not a wrapper that
leaves a separately owned dependency alive on unload. Loading both against the
same one-shot browser registration is not a supported way to duplicate ownership.

Chapter 11's relationships still apply. The MCP manager owns the browser
registration; the claimed client belongs to the manager, its skill, and that
registration. The client owns its remote Tools. Unloading `gui-debug` closes its
client and removes executable browser tools immediately, while retaining the
pending-unload manual state until reduction. It does not close unrelated GUI
traffic. Reconnection and reactivation are not silently invented when a consumed
registration is no longer usable.

The browser responder lives in `Examples/Gui/wwwroot/mcp.js`. It provides bounded
visible DOM text and selectors, a restricted click/input surface, and the existing
speech queue. It does not offer JavaScript evaluation, arbitrary navigation,
other-profile access, or pixel inspection. Password values are omitted from
snapshots; visible text can still be sensitive and must be handled accordingly.

## Preserve Input Origin, Not Just Its Value

`gui_input` identifies one visible editable text field, replaces its value with
literal text, and dispatches input and change events. Quotes, spaces, Unicode,
and strings such as `{Enter}` remain text, not a keyboard macro language. Empty
text clears the field. Ambiguous selectors, hidden or read-only fields, and
non-text inputs are refused. This is DOM input, not physical keyboard provenance.

The important change is the origin of that input. The MCP handler dispatches a
marked event:

```javascript
element.dispatchEvent(new CustomEvent("input", {
    bubbles: true,
    detail: { mcp: true }
}));
element.dispatchEvent(new Event("change", { bubbles: true }));
```

The prompt listener maintains `humanTyping` from actual input events. MCP input
does not create that cause. A later speech callback combines current speech with
`humanTyping`, rather than reconstructing human intent from a nonempty field.
Successful submission clears both the field and the human cause explicitly.
Human whitespace still counts as typing.

That one retained fact prevents two different freezes: the immediate pause on
programmatic input, and the delayed pause when speech finishes while programmatic
text remains. Speech still pauses when configured to do so. Genuine human typing
still holds pause after speech ends. Another tab's explicit unpause is not
continuously vetoed by an unchanged local cause, as in the previous chapters.

A tool must not silently overwrite a user's unfinished draft without permission.
Origin tracking fixes scheduling; it is not authorization to replace text.
The manual tells the Agent to preserve human input and inspect the effect of its
actions. A follow-up snapshot, not the requested argument string, establishes the
resulting field value.

## Historical First Use

The following account is retained from the earlier manuscript.
It describes earlier encounters, including their failures and human reactions,
not a claim that this C# endpoint repeated those live-provider trials. Historical
Go snippets and tool names below describe those runs, not the current C# API. Its practical lesson is the difference between
building a test harness and using the real interface.

## What the wiring found

The plumbing above is small because the preceding chapters already own the difficult lifetimes. A skill declares an MCP server. The registry connects when the skill loads and disconnects when it unloads. A browser tab exposes four tools over a WebSocket that was already there. The interesting part is what arrives the first time an agent can see a screen it did not render and press a button nobody told it about.

The first-edition experiment ran on September 19 and 20, 2026. Its driver and transport arrived on the first day; four code repairs and three documentation commits followed the next afternoon. The retained [experience report](../docs/edition-2-notes/enhancement-12.md#historical-source-checks) supplies the account below. The ninety-eight-second outcome is that report's measurement, not a new run of this edition.

The asymmetry is useful. Once a failure was visible, the repair was often small. The experiment exposed faults that source development and passing graders had missed, but it did not find every fault without help. The report specifically attributes a truncation observation to the human supervisor and the unsafe rendering path to subsequent source inspection. Those are valuable findings too; their origins matter when assessing the observer.

### The user that cannot read the source

The historical driver was a separate small binary. It connects to the same hub the browser connects to, discovers its tools through MCP, and registers exactly one tool of its own, `file_report`, so it has somewhere to put conclusions. It is built on `NewBareAgent`, which is an agent with no builtin tools at all. No `read_file`, no `run_command`, no editor. Its interaction tools arrived through discovery. That experimental driver also gained submission and waiting helpers; those are history, not extra tools assigned by this four-tool exercise.

Its system prompt states the constraint directly:

```
You CANNOT read files, edit code, or run commands. You can only
interact through the GUI, exactly as a human user would.
```

That restriction is the entire value of the driver. An agent with filesystem access will answer a question about the interface by reading the source, which tells you what the interface was meant to do. An observer restricted to the GUI has to answer from the screen, which tells you what the interface actually does. The gap between those two answers is where the bugs live.

### The human observer

The protocol around these runs matters as much as the tooling, because it is what makes the results mean anything.

Bill Cox supervised the experiment, reading the reasoning as it streamed. The method was to let the driver inspect the interface before supplying a diagnosis. Watching it walk past a defect is useful evidence about the view it has; immediately pointing out the defect would erase that evidence.

He did direct, and the distinction is worth drawing precisely. Direction covered what to test next, which binary to rebuild, and once, usefully, the observation that a stale server was still holding port 8084 while the freshly built binary bound to nothing and served no one. That is operational guidance, and withholding it would have wasted an hour proving nothing. Findings were different. The retained report also records a screenshot observation about truncation supplied by the human. That finding cannot be counted as an unaided discovery merely because the subsequent repair was automated.

The reason to run it that way is that a supervisor who volunteers the answer cannot tell the difference between a tool that works and a tool that agrees. A supplied hint can make a successful report look like an independent discovery unless the assistance is recorded. To learn whether the snapshot is sufficient for a finding, let the observer attempt it before supplying the missing answer.

This is the same discipline the book asks of a reader supervising an agent on live code, applied to the machinery itself. Keep the human assistance with the result so the next reader can tell which capability was exercised.

### The observer that froze what it watched

The first runs never reached a task at all. The driver typed its prompt and the coding agent stopped responding, permanently, before any work began.

The cause was a feature working exactly as designed. Chapter 7 pauses the engine while the user is typing, so that an agent does not barrel ahead while a human is halfway through composing a hint. The feature exists for a particular working style: Bill Cox reads the agent's reasoning as it streams and sends corrections mid-turn, which makes the pause the difference between a hint that lands and a hint that arrives after the decision it was meant to change. The first-edition implementation watched the input field:

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

The historical repair expressed two intentions. A flag marks input as programmatic so the pause handler ignores it, and submission sends an explicit `unpause` rather than relying on an event that will not arrive:

```javascript
if (window._mcpProgrammaticInput) return;
```

The general shape of this is older than software. An instrument that shares a channel with the thing it measures will perturb it, and the perturbation is worst where the system was tuned to human timing. Pause on typing, debounce on scroll, idle timeouts, animations that wait for a settle: each one encodes an assumption about how fast and how continuously a person acts. An automated driver violates all of them at once, and it does so silently, because from the GUI's perspective nothing unusual occurred. A user started typing and never stopped.

### The second pause, after the fix

The second-edition student reproduced the self-input freeze in a real browser. Marking MCP input as programmatic fixed the immediate listener: a model could fill the field and click Send with speech disabled. Independent review then turned native speech on. The utterance finished, the programmatically filled field still contained `MCP submitted`, and the Agent remained paused.

The missed path was `pace()`. On speech transitions it looked at the field's contents again, forgetting how the text got there. The input listener knew the event was programmatic; the later speech callback silently promoted that same text to human typing. Ignoring one event could not repair a second reader of the same mistaken signal.

Keep the cause, rather than reconstructing it from the value. The browser MCP handler sends a marked event after setting the control:

```javascript
element.dispatchEvent(new CustomEvent("input", {
  bubbles: true, detail: { mcp: true }
}));
element.dispatchEvent(new Event("change", { bubbles: true }));
```

The ordinary input listener preserves the actual cause:

```javascript
input.addEventListener("input", event => {
  humanTyping = !event.detail?.mcp && input.value.length > 0;
  pace();
});
```

Speech and typing then combine as `speaking || humanTyping`. Successful submission clears `humanTyping` explicitly after clearing the field. Human whitespace still counts as typing; programmatic text does not acquire a human origin merely because a speech callback ran later.

Pause remains Agent-wide. A browser sends automatic pause/unpause commands when its local causes change. If another tab explicitly unpauses, an unchanged local cause cannot immediately veto that command. The corrected native-browser check covers speech ending with programmatic text still present, genuine whitespace typing holding pause, explicit unpause from another tab, and a subsequent actual MCP submission. It also waits for the concluding speech to finish. The earlier passing check with speech disabled proved less than its author thought.

Those checks used native speech start/end events, with a local provider fake driving the actual Agent and browser. They establish the measured pause transitions, not subjective listening quality. The second tab closed after the global-unpause control; the retained earlier two-tab speech stall has no established cause and is not a general claim about concurrent speech reliability.

### A debugger that would not start

The first historical task given to the driver was small on purpose: ask the coding agent for Tower of Hanoi, then watch it debug the result. The agent wrote the program, ran it, and produced fifteen moves. Then it started `dlv` and stopped forever.

Chapter 4 predicted this in print. Section 4.8 is titled "Your agent drives a debugger," and it warns that a dispatcher which waits for process exit will hang on an interactive program, because "the dispatcher would wait for an exit that never comes, and on pipes `dlv` refuses to start at all." The chapter named the trap. The reference implementation walked into it anyway.

The mechanism was a single synchronous read. `toolRunCommand` read the PTY inside the tool function, so the function could not return until the process closed its output. For `go build` that is correct and invisible. For a debugger sitting at a prompt it is fatal, and it is fatal in a way that disables the exact feature meant to rescue it: `ai_callback_pattern` is matched inside `Wait`, and `Wait` runs after the tool function returns. The escape hatch was behind the door it was supposed to open.

The first-edition repair changed the lifetime: the tool function returned once the process was running, and a reader goroutine owned the PTY until exit. Its `DeferFinish` field told that dispatcher the job would finish itself. The following historical excerpt shows the handoff; it is not an API to add to the current implementation:

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

The experience report records the next surprise: the debugger worked, but the older Chapter 4 checks returned zero.

The failure signature pointed in the wrong direction. Tool output came back as the empty string, and the capture file on disk held zero bytes. That reads as a broken output path, and an hour can be spent there. The actual fault was lifetime again. There were two tool dispatchers in that first-edition codebase: `Engine.Execute`, which serves chapters 3 and 4 through the older `Ask` path, and `Actor.dispatchTool`, which serves chapter 6 onward. Only the second had been taught about `DeferFinish`. The first still called `job.Finish` the instant the tool function returned, and `Finish` closes the capture file. The reader goroutine was writing correctly the whole time, into a file that had already been closed underneath it.

The diagnostic detour is worth recording, because it cost more than the bug. Debug prints added to `Actor.dispatchTool` never appeared in the output, and the first explanation reached for was a stale build cache. It was not stale. The prints were in a function the failing tests never called. Silent debug output is evidence of the wrong code path at least as often as it is evidence of a bad build, and the cheaper check is to confirm which function runs before rebuilding anything.

### Ninety-eight seconds

With both dispatchers corrected, the report records an end-to-end rerun. The driver typed the prompt into the real text field and submitted it. This time the coding agent wrote a three-disk `hanoi.go`, ran its seven moves, then opened `dlv` and worked through a full session: set a breakpoint on `move`, continue, print `disk`, quit. The reported debugger exit was zero and elapsed time ninety-eight seconds. The earlier hung run had used four disks and fifteen moves; those are different runs.

That historical run explains why the capability is worth building. Chapter 4 of this edition already establishes the independent job lifetime; keep it when adding browser access. An agent that can drive a debugger can inspect a running program's state rather than reasoning about what the state ought to be, and an agent that can drive a GUI can do the same for an interface.

### The bug that only appears after you fix the bug

The next task sent the driver into the settings panel with instructions to try values a careful user would not. It set temperature to negative five, and the server accepted it.

That is the obvious bug, and the fix is ordinary: a `clamp` function, called both when a settings patch arrives and when settings load from disk, so a hand edited file cannot bypass validation either. The original chapter also reports a targeted check with `clamp` disabled; that historical claim is separate from this edition's settings evidence.

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

There was a real choice about how far to take the repair. The cautious option preserves the existing wire format and introduces a second type for broadcasts, leaving the patch struct untouched, which costs nothing today and leaves two nearly identical structs for the next reader to confuse. The maintainer settled it in one line: the format has no external consumers, so make the code clean. The repair therefore dropped `omitempty` from every field, so a broadcast always carries complete state, and deleted `Apply` outright, leaving one merge path with defined semantics.


A zero missing from a full snapshot would still be the defect. The corrected broadcast must carry `temperature: 0` and `tts_enabled: false` explicitly. A clamped value on disk, the value applied by the server, and the value shown by the client are separate observations. Chapter 8 now follows all three instead of treating a settings-file inspection as proof of the screen.

### The settings panel that does not exist

The strongest finding came from a test that failed in an unusual way.

Asked to exercise the sidebar tabs, the driver reported that clicking "Artifacts" opened the Settings panel. The report was specific, confident, and false. There was no Settings tab in that version of the markup. The `data-tab` and `data-panel` attributes map correctly, chats to chats and artifacts to artifacts, and clicking either one does what it says.

The report was fiction, and the reason it was fiction is the finding. Tab state lived only in a CSS class. `gui_snapshot` built its element labels from `textContent`, `aria-label` and `placeholder`, and read no state attributes at all, so the driver could see that two tabs existed and could not see which one was selected. It had also reached for an ambiguous selector, `button:nth-of-type(2)`, because the buttons carried nothing better to aim at. Asked what happened after the click, it had no way to observe the answer, and it produced a plausible one.

In this run, the driver filled missing state with a confident guess. The observation supports fixing the missing state; it does not establish that every model must guess whenever information is absent.

The repair was to put state where a machine can read it. The hamburger and the tabs now carry `aria-expanded` and `aria-selected`, synchronized on every click, with the initial value derived from the live CSS class rather than hardcoded, because the markup default had already drifted from the rendered default. `gui_snapshot` now reports those attributes alongside each control.

Rerunning the identical test produced a correct before and after table, and in the one place where the driver lacked information, it wrote "not observable" instead of inventing a panel. Same model, same prompt, same GUI. The only change was that the interface stopped hiding its state.

This is the accessibility argument in a form that a developer who has never used a screen reader can feel directly. Semantic state attributes are usually presented as a courtesy extended to users with assistive technology. They are also the difference between an automated observer that reports what happened and one that reports something reasonable. This observer produced wrong data where the interface supplied no selected state, and delivered it with the same confidence as the truth.

### What an interface owes an observer

Two smaller findings came out of the same pass, and both generalize.

Source inspection found an injection hole. `artifact-scroll.js` built the tool card header by interpolating the tool name and its serialized arguments into `innerHTML`. Tool arguments can carry text from outside the agent. They contain filenames, URLs, and the contents of files just read, which is to say text an attacker can influence. Twenty lines further down, the result path rendered with `textContent` and was safe. One file, both patterns, and the vulnerable one sat on the path that renders attacker-adjacent text. The book spends a chapter on prompt injection arriving through the model. This was the same threat arriving through the renderer.

The truncation investigation followed the human screenshot observation, and the initial notes about it were backwards. The working document asserted that the driver saw the full DOM while the human saw a trimmed version. Measurement found four independent caps running the other way: the human view truncated tool input at 500 characters and results at 1000, while the snapshot truncated each artifact at 120 characters and the whole document at 4000. For an individual result preview, the snapshot could show 120 characters where the human view showed 1000. Worse, the snapshot silently dropped every artifact past the tenth with no marker, so it could not distinguish ten artifacts from fifty.

The report corrected its earlier claim after checking the actual caps. The caps were kept, because both views need them, and both were made honest: truncated text is now reachable through a `title` attribute rather than deleted behind an ellipsis, and the snapshot states how many characters and how many artifacts it withheld.

That is the rule the section converges on. Silent truncation lets an observer report that a screen looks fine without seeing the omitted content. Hidden control state invited the guess in this run. Unsafe HTML interpolation gives attacker-controlled markup an execution path. Runtime observation and source inspection answer different parts of the diagnosis. The unsafe renderer was visible in source; the observer's confident report showed what the missing state caused during use.

### The bugs travel back into the book

The instruction that produced this section came once the fixes were in: update the chapter summaries wherever a bug made it through. Tracing them was the first step, and each defect was matched to the chapter that should have prevented it. Several repairs also exposed an instruction that should have made the obligation explicit.

The original Chapter 4 explained the job model, named the interactive process trap in plain words, and never stated the obligation that follows from it. Chapter 8 built the tool card and showed the rendering without saying which parts of it carry text the agent did not write. Chapter 12 specified `gui_snapshot` with a 4KB cap and did not require the cap to announce itself. In each case the mechanism was taught correctly and the duty attached to the mechanism was left implicit.

The original Chapter 9 is the sharpest example, because the reasoning that let three bugs through is stated plainly in its opening:

> The graded surface is small: four server-side settings checks and a parity gate. Everything else is client code you can see working.

That passage left client behavior to visual inspection. Looking at a screen confirms that a value was accepted. It does not confirm that the value was validated, that the server and the screen agree about it afterward, or that a control's state can be read by anything other than an eye. Three bugs fit in that gap, and all three shipped.


A missing obligation in the teaching is worth repairing even when an experienced implementer could have inferred it. That does not make every implementation defect inevitable or transfer every mistake to the book. The repair was therefore made in the prose as well as the source. The historical repair made Chapter 4 state that the tool returns when the process is running and the job owns it from there, and gives the failure signature, which reads like broken output capture and is really a closed file. Original Chapter 8 states that tool cards are built with `textContent` and that capped content stays reachable. Original Chapter 9's summary gains the three obligations that survive the absence of a grader. Original Chapter 12 requires the snapshot to report what it withheld.

A loop closes here. The agent described in these chapters was built by following these chapters. When it acquired eyes and a way to press buttons, the first thing it did was find places where the chapters were wrong. Some repairs belonged in the GUI and some obligations needed to become explicit in the instructions. The next implementation gets both lessons.

A book that produces a working program gets to be tested by the program it produces. This section is the first time that test came back with findings, and the findings were about the book.


### Later Historical Browser Exercises

The retained October 10, 2026 browser sessions used `claude-haiku-5-5`, `gpt-4o-mini` and `gemini-3.8-flash`. These are the models actually exercised, not a claim about today's catalog. The [live record](../docs/edition-2-notes/ch13.md#student-implementation-and-initial-review-submission) links the captured requests and real Chrome observations.

Claude Haiku and Gemini changed the effective font to 20, received the refreshed snapshot, and unloaded the skill. Claude then incorrectly said the MCP tools remained available until compaction. The old generic unload message encouraged that explanation; the final request already showed the tools gone. The implementation's message was clarified to separate immediate server removal from pending instruction removal. No paid repeat was needed merely to demonstrate revised wording.

Both GPT-4o mini attempts failed the full task. The first guessed `input[type='number']`, received a missing-selector error, then unloaded without repairing it. A guided second attempt supplied the observed `#font-size` selector and explained that it was a select. It changed the setting and unloaded, then tried `read_file` on `path/to/the/snapshot` despite the instruction forbidding file access. That nonexistent path failed, and the final response claimed the snapshot could not be inspected. The capture already contained the refreshed state.

The second attempt passed the script's GUI assertions. It remains a task failure. That distinction is especially useful here: working controls do not guarantee that a model will use them correctly or report their result faithfully. A guided selector is assistance, and a prompt restriction is weaker than the historical blind driver's actual absence of file tools.

These live runs preceded the final native-speech correction and small manual/message corrections. They establish the recorded interaction and unload behavior on their captured source. The corrected speech path has its separate native-browser evidence; no final-source paid rerun is claimed. Likewise, a visibly effective font change is not a restart-persistence test. Chapter 8 supplies that separate check.

### Checking the correction itself

The first-edition chapter ended with a quotation-verification story. It said that “You know the client works because you can see it.” had been invented and appeared nowhere in the original settings chapter. The retained source contains those exact words, split across a newline. They are present even in the revision before this narrative was appended.

A literal one-line search misses a wrapped sentence. That miss establishes a search result, not the sentence's absence. Read the surrounding paragraph or normalize whitespace before comparing the quotation. The supposed correction needed the same scrutiny as the original claim.

The comparison with the phantom Settings panel remains useful, with a narrower conclusion: check the observable that supports the claim. For a tab, read selected state. For a quotation, compare the actual source text. For a successful tool, inspect its correlated result. A confident diagnosis can survive an inadequate check just as easily as the bug it diagnoses.

## Exercise This Endpoint

Start with the ordinary default path and verify that no browser tools are enabled
merely because a page connected. Then run with `--gui-debug`. Inspect the first
actual model request: the manual must be in provider instruction context exactly
once, browser declarations must be present, and there must be no fabricated
startup `load_skill` exchange or unsolicited snapshot.

Use explicit calls to request a snapshot, choose a visible selector, change a
field with literal text, and request another snapshot. Inspect the real input
events and their marked origin. Leave programmatic text present through speech
completion, then verify that genuine human whitespace still holds pause. Queue
and inspect speech, while keeping queue wiring separate from an audible or human-accessibility judgment.
Unload the skill and verify that later calls are refused and the ordinary GUI
connection remains available.

The source-only endpoint is `solutions/ch12`; it includes the same
library, CLI, optional GUI, manuals, and component tests as `main` at this boundary.
The current offline verification uses real components with working local peers:

- ordered startup, duplicate elimination, validation rollback, instruction placement
  in all three adapters, frozen inputs, and no invented tool dialogue;
- held optional-server initialization, readiness refusal, real remote dispatch,
  joined cleanup, and preservation of an earlier save on startup failure;
- a scoped six-check CLI executable grade, covering startup handshake, explicit
  observations, interaction, speech-queue content, and unload;
- an isolated actual browser with shipped `--gui-debug`, literal Unicode input,
  human/programmatic pause-cause separation, explicit snapshots, queue wiring,
  identity refusal, unload,
  and normal owned-host exits;
- a separate public-API C# consumer exercising explicit startup selection,
  instruction rendering, network-free initialization, and complete destruction.

These results do not establish real-provider integration, pixel vision, audible
speech quality, general accessibility, or permission to control another person's
browser. The real-provider allowance for this run is exhausted. Keep those
exercises explicitly incomplete rather than converting a successful fake-provider
session into evidence it cannot supply.

The useful result is modest and concrete: a configured Agent can obtain evidence
about its own working page, act through declared tools, and inspect the result.
The interface becomes part of the engineering feedback loop without becoming a
second owner of the Agent's conversation or lifecycle.
