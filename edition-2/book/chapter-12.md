# Chapter 12: The Agent Sees Itself

A coding agent that cannot see its own GUI is debugging blind. Reading the source tells it what a button should do. Watching the running interface tells it what the button did. Chapter 11 supplied the browser tools and MCP transport; this chapter makes that view available at startup through one skill, then points it at the interface the agent is using.

The wiring is short and mostly mechanical. What follows it is less tidy: an observer that froze what it watched, a debugger whose escape hatch was unreachable, and a settings repair that made the server correct while leaving the screen wrong. Giving an agent eyes is useful partly because it exposes how much the programmer had been assuming.

## TL;DR

Extend Chapter 11 with `--gui-debug` and `--skills-dir PATH` in both the headless command and the optional GUI application. The first selects the reachable `gui-debug` skill during Agent construction; the second overrides `EN_SKILLS_DIR`. With debug enabled and no directory selected, use `./skills`. When a directory is selected and `EN_PRIMARY_SKILL` is empty, these applications select `ensemble` as primary. Keep explicit skill kinds and tool grants.

Startup uses the existing skill/MCP lifecycle. Assemble the primary and startup manuals before freezing the system prompt. Ordinary later `load_skill` calls still return instructions through history. Discover the peer's tools, admit only the skill's grants, hide automatic tools from model declarations, and refresh completed DOM and speech-queue content before every round. A click or input must affect the next snapshot.

Unloading `gui-debug` immediately removes its external tools and automatic context and closes its peers. A guessed tool name cannot bypass removal. Compiled grants and instruction text remain pending removal under Chapter 9's existing rules; this chapter adds no compaction. Preserve human typing and speech pacing without treating programmatic input as a human pause cause. Explicit unpause in any tab still unpauses the whole Agent.

Build on the existing owners and common interfaces. Agent owns its skill registry and MCP manager; the optional GUI owns browser connections and framing. No new callback bundle, dispatcher or driver application is needed.

### The skill

Here are the supplied files, including the primary that makes the optional skill reachable. Put this in `skills/ensemble/SKILL.md`:

```yaml
---
name: ensemble
type: primary
description: A concise coding assistant with optional browser inspection
tools: read_file write_file edit_file list_directory search_files run_command
loadable-skills: gui-debug
---
You are a coding assistant. Use the available tools to inspect facts before
claiming results. Available skills: $SKILLS.
```

And this in `skills/gui-debug/SKILL.md`:

```yaml
---
name: gui-debug
type: loadable
description: Inspect and operate the connected browser
depends: ensemble
tools: gui_snapshot tts_queue gui_click gui_input
mcp_servers:
  - name: browser-debug
    transport: websocket
---
The current browser snapshot and speech queue arrive automatically before each
round. Use gui_click with a visible CSS selector and gui_input with selector and
text. Observe the next snapshot to confirm the effect. Hidden or omitted state
is not evidence of success: report what you can actually observe.
```

The kind, dependency and grants do different jobs. `type: loadable` makes this an optional skill; the primary's `loadable-skills` makes it reachable; `depends` ensures the base is present; `tools` admits the named capabilities. Discovery alone grants nothing. An MCP peer advertising another tool does not authorize the model to use it. Reverse calls still pass through the Agent's enabled-tool dispatch.

The browser entry uses the WebSocket tunnel from Chapter 11. For a deterministic check, replace only that server entry with a stdio peer's command and arguments, retaining the skill kind and grants. The existing Chapter 11 Python peer illustrates that external boundary. A subprocess and a browser use the same discovery and dispatch path; only transport creation differs. The historical wire version and course extensions retain Chapter 11's scope.

### Skill-based MCP lifecycle

Loading a skill with `mcp_servers` asks the Agent's MCP manager to start its peers. It opens the transport, initializes the client, discovers the tools, and retains the connections for cleanup. The manager exposes only admitted tools through the Agent's existing declarations and dispatcher. Jobs still own running calls and their output; a callback delay is no new server lifetime.

This wiring already exists. The registry reaches `Agent.MCP()` through its actual parent interface. The composition root constructs and attaches the Agent's children before loading skills. The command selects configuration and calls that root; it does not become a collection of closures holding sibling services together. The optional GUI supplies its public transport without importing GUI code into the core.

At startup the application sets `AgentSpec.StartupSkills` to `[]string{"gui-debug"}`. The constructor loads the primary, activates the optional startup skill through the same admission path, and assembles both manuals before the prompt freezes. There is no model-authored call to answer yet, so inventing a `load_skill` result would invent conversation history. The accepted implementation records application-origin activation at sequence zero instead.

Once construction finishes, the distinction matters: a later `load_skill` returns its manual as a tool result and leaves the system prompt alone. Likewise, unloading a startup skill cannot rewrite the frozen prompt. Describe the immediate effect accurately: its MCP tools disappear, automatic browser context stops, and its clients close. For stdio that includes terminating and reaping the peer while the Agent can continue. Instruction removal remains pending; an old manual is not evidence of a live connection.

### The four browser tools

These are the tools from Chapter 11's browser peer:

| Tool | Automatic timing | Observable result |
| --- | --- | --- |
| `gui_snapshot` | Each round | Markdown DOM: panes, selectors, semantic control state and artifact previews |
| `tts_queue` | Each round | JSON array of pending speech utterances with state and timing |
| `gui_click` | Model call | Clicks the visible element named by `selector` |
| `gui_input` | Model call | Sets the element named by `selector` to `text`, then sends input/change events |

The two automatic tools remain absent from model declarations. Their completed output appears in ephemeral context, replaced on the next round. The model sees what changed without repeatedly asking for another screenshot. Preserve Chapter 11's wait for completed automatic content; a three-second running notice cannot describe a DOM.

The callable tools return `ok` on success in the supplied browser. That confirms the handler ran. The next snapshot establishes whether Settings opened or a field acquired its new value. A test that searches the whole request for the desired text may find only the earlier call arguments. Read the result associated with that call and the actual state in the refreshed snapshot.

### Exercise

Add the startup selection and directory override to both clients, reusing the existing skill lifecycle. Supply the two complete manuals above. Preserve fresh automatic context, actual click/input behavior, immediate external-tool removal, peer cleanup, and the human-input cause distinction described below.

From your core solution directory, build the headless executable:

```sh
go build -o /tmp/ensemble-ch12 .
```

With a stdio peer declared in the selected skill directory and the existing `LLM_VENDOR`, `LLM_MODEL`, `LLM_BASE_URL` and `LLM_API_KEY` configured, its protocol invocation is:

```sh
/tmp/ensemble-ch12 --gui-debug --skills-dir ./skills
```

Stdin uses the existing newline-delimited input; the flag supplies no implicit prompt:

```json
{"kind":"prompt","text":"Inspect the browser state, click a visible control, and report the change from the next snapshot."}
```

EOF drains the submitted work and ends the session. For human terminal conversation, append `chat`. A headless process can talk to an external stdio peer; to reach the actual browser, build and run the separate GUI as shown below.

The revised grader runs both actual clients with valid skills and a stateful external stdio peer. It checks startup instructions, directory override, fresh snapshot and TTS data, forwarded click/input arguments and results, then unloads the skill and attempts another browser call. That call must fail, the automatic context must be gone, and the peer must have stopped while the Agent remains alive. The independent browser evidence covers native speech, semantic DOM state and the real tunnel; a stdio fake cannot prove those browser properties.

| Check | Points |
| --- | ---: |
| Skill loads at startup | 15 |
| Hidden automatic context is injected | 20 |
| Snapshot reflects interaction | 15 |
| Speech queue is visible | 10 |
| GUI interaction has actual effects | 15 |
| Skill unload removes external capability | 15 |
| Corrected Chapter 11 parity | 10 |

From the repository root:

```sh
go run ./cmd/grade-edition2 -ch 12
```

The default is `solutions/ch12`; supply your core solution directory as the positional argument and use the existing `-gui` option if its separate GUI module is elsewhere. The original grader remains historical: it returned 0/100 after an untyped fixture prevented startup, while its parity detector required particular method names. That run observed neither successful browser behavior nor the later real speech/input defect. Valid runtime obligations need their own evidence.

---

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

## Taking it for a spin

Use the supplied WebSocket skill for the real browser. From the core solution directory, build its separate module:

```sh
(cd gui && go build -o /tmp/ensemble-ch12-gui ./cmd)
```

With your provider environment configured, run from that same core directory:

```sh
/tmp/ensemble-ch12-gui --gui-dir gui/web/gui --gui-debug --skills-dir ./skills --port 8088
```

Leave the terminal open and connect to `http://localhost:8088` before submitting a prompt. These assets still come from disk, so the explicit `--gui-dir` matters when launching outside the GUI module. The browser supplies the peer that the startup skill has selected. Use a scratch copy or a fresh selected `--save` path when you want an independent conversation; Chapter 10's default load still applies.

Ask the agent to open Settings, change the font size to 20, inspect the next automatic snapshot, then unload `gui-debug`. Tell it to use browser tools and avoid file or shell access. Observe the actual control and effective page font as well as its report. After unload, inspect the next provider request: browser declarations and automatic context should be absent even though the startup manual remains in the frozen system text.

The retained October 10, 2026 browser sessions used `claude-haiku-5-5`, `gpt-4o-mini` and `gemini-3.8-flash`. These are the models actually exercised, not a claim about today's catalog. The [live record](../docs/edition-2-notes/ch13.md#student-implementation-and-initial-review-submission) links the captured requests and real Chrome observations.

Claude Haiku and Gemini changed the effective font to 20, received the refreshed snapshot, and unloaded the skill. Claude then incorrectly said the MCP tools remained available until compaction. The old generic unload message encouraged that explanation; the final request already showed the tools gone. The implementation's message was clarified to separate immediate server removal from pending instruction removal. No paid repeat was needed merely to demonstrate revised wording.

Both GPT-4o mini attempts failed the full task. The first guessed `input[type='number']`, received a missing-selector error, then unloaded without repairing it. A guided second attempt supplied the observed `#font-size` selector and explained that it was a select. It changed the setting and unloaded, then tried `read_file` on `path/to/the/snapshot` despite the instruction forbidding file access. That nonexistent path failed, and the final response claimed the snapshot could not be inspected. The capture already contained the refreshed state.

The second attempt passed the script's GUI assertions. It remains a task failure. That distinction is especially useful here: working controls do not guarantee that a model will use them correctly or report their result faithfully. A guided selector is assistance, and a prompt restriction is weaker than the historical blind driver's actual absence of file tools.

These live runs preceded the final native-speech correction and small manual/message corrections. They establish the recorded interaction and unload behavior on their captured source. The corrected speech path has its separate native-browser evidence; no final-source paid rerun is claimed. Likewise, a visibly effective font change is not a restart-persistence test. Chapter 8 supplies that separate check.

### Checking the correction itself

The first-edition chapter ended with a quotation-verification story. It said that “You know the client works because you can see it.” had been invented and appeared nowhere in the original settings chapter. The retained source contains those exact words, split across a newline. They are present even in the revision before this narrative was appended.

A literal one-line search misses a wrapped sentence. That miss establishes a search result, not the sentence's absence. Read the surrounding paragraph or normalize whitespace before comparing the quotation. The supposed correction needed the same scrutiny as the original claim.

The comparison with the phantom Settings panel remains useful, with a narrower conclusion: check the observable that supports the claim. For a tab, read selected state. For a quotation, compare the actual source text. For a successful tool, inspect its correlated result. A confident diagnosis can survive an inadequate check just as easily as the bug it diagnoses.
