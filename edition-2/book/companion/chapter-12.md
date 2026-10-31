# Chapter 12: The Agent Sees Itself

Chapter 11 gives an Agent a connection to browser tools. That does not yet give it a reliable account of its own interface. A click can succeed while the screen stays wrong. A snapshot can describe every button and omit which one is selected. An observer can even stop the system it came to inspect.

Bill's experiment with a GUI-only driver made those distinctions unusually concrete. The driver could not read source, edit files, or run commands. It had to use the interface, much as a human would. Its first attempt to type a prompt froze the coding agent before the task began. Later it reported a Settings panel that did not exist. Between those failures, it helped expose defects that source development and passing checks had missed.

The [source chapter](../edition-3/chapter-12.md) records that history alongside a smaller startup design: select one skill and reuse the browser connection already built. The plumbing deserves to stay small. The judgment required to use its results does not.

## Selecting the View at Startup

Both the headless command and the optional GUI accept `--gui-debug` and `--skills-dir PATH`. The first selects the reachable `gui-debug` skill during Agent construction; the second overrides `EN_SKILLS_DIR`. Debugging without a selected directory uses `./skills`. When a directory is selected and `EN_PRIMARY_SKILL` is empty, these applications choose `ensemble` as the primary. These are application defaults, not reasons for the core to acquire GUI dependencies.

The primary and optional skill make different promises. The `ensemble` primary grants ordinary coding tools and lists `gui-debug` as loadable. The latter is explicitly a loadable skill, depends on `ensemble`, and grants `gui_snapshot`, `tts_queue`, `gui_click`, and `gui_input`. Its `browser-debug` MCP server uses the WebSocket tunnel. Kind determines how a skill participates, reachability determines whether it can be selected, dependency supplies the base, and the tool names determine admission. None can substitute for another. Discovering an impressive extra tool does not grant it, and reverse calls still pass through the Agent's enabled-tool dispatch.

The application expresses the startup choice through `Config.StartupSkills`; construction loads the primary and activates `gui-debug` through the same admission path as a later load. Both manuals enter initial SYSTEM. There is no model-authored `load_skill` call to answer, so a fabricated tool result would be fabricated history. The implementation described in the source records application-origin activation at sequence zero instead.

This stays inside the existing ownership tree. The Agent owns its registry and MCP manager; the registry reaches the manager through its actual parent interface. The composition root attaches those children before loading skills. The GUI supplies its public transport and owns browser framing, while its executable consumes the public core. A new callback bundle, second dispatcher, or permanent driver application would mostly create new places to get these relationships wrong.

Startup has one genuine timing wrinkle: the bundled browser connects after the Agent is constructed. Known metadata for its four tools permits admission before attachment, not execution without a browser. An already-connected peer supplies its actual catalog instead; cold-start metadata must not replace discovery. Configured endpoint and transport overrides remain meaningful. When the real peer attaches, its schemas and ordinary-versus-automatic identities are validated, and the owning skill's explicit names still govern permission. An unknown required name is an admission failure, not an invitation to guess a capability into existence.

Later loading adds instructions through Chapter 9's provider-aware SYSTEM path, with the paired tool result reporting activation rather than becoming instruction authority. No cache-hit promise follows from that arrangement. Unloading has the split lifetime established in Chapter 11: external grants disappear immediately, peers with no remaining grants close, and a stdio peer is terminated and reaped while the Agent can continue. Guessed calls must fail too. Withdrawn external declarations disappear even from Anthropic's deferred catalog; ordinary dormant compiled tools retain Chapter 9's disclosure behavior. Compiled grants and retained manuals remain subject to pending removal. Compaction is still future work, not a side effect of closing a browser connection.

The operational distinction matters even before a test starts. A headless process can use a stdio peer, but the actual browser needs the separate GUI application and a connected tab. Its web assets come from disk, so the asset directory must match the launch location. The debug flag supplies capabilities, not an implicit prompt. Headless protocol input remains newline-delimited, EOF drains submitted work, and the interactive `chat` command follows its flags. Chapter 10's default conversation loading also still applies; an apparently fresh browser session can be an old conversation unless its saved state was deliberately separated.

## Observation Is a Call, Not a Background Assumption

All four browser tools are ordinary on-demand tools. `gui_snapshot` returns the DOM account: panes, selectors, semantic control state, and artifact previews. `tts_queue` returns pending utterances with their text, state, and timing. `gui_click` acts on a visible selector; `gui_input` sets the named control and dispatches input and change events. Chapter 11's automatic-context mechanism remains useful elsewhere, but browser snapshots and queue inspections are not attached to every request.

The ordinary sequence is an observation, an action using what was observed, and a new observation. In the supplied browser, `ok` from a click or input establishes that the handler ran. It does not establish that Settings opened, that a value survived validation, or that the effective page font changed. The next snapshot has to reflect the interaction. A running-job notice is not a delayed DOM rendered in an unusual format; the observer needs the actual result.

This also determines where evidence is read. Searching an entire provider request for the desired value can find the value in the earlier call's arguments. That proves what was requested. The correlated result and refreshed state show what happened. Requested speech-queue contents likewise establish pending text and status, not that a listener heard intelligible audio.

A stateful stdio peer is useful here because it substitutes only the external transport boundary. It can retain the same skill kinds and grants, exercise discovery and dispatch, change state after an interaction, and reveal unsolicited observations or failed cleanup. It cannot prove semantic state in a real DOM, native speech behavior, or delivery through the browser tunnel. The historical wire version and course extensions retain Chapter 11's scope; changing transport does not make them a universal MCP guarantee.

## A Driver That Cannot Peek

The source attributes the first-edition experiment to a retained experience report. Its driver was a separate small binary built on `NewBareAgent`, with no builtin filesystem, shell, or editing tools. It connected to the browser's hub, discovered interaction tools through MCP, and supplied `file_report` as a place to put its conclusions. That historical driver also gained submission and waiting helpers. Those helpers are not additions to this chapter's four-tool startup design.

Removing source access changed the question the experiment could answer. A coding agent can explain a button by finding its handler, which is useful when repairing the button and inadequate when establishing what the user experienced. The restricted observer had to obtain its answer through the interface. A prompt telling a fully equipped agent not to read files is a weaker boundary than having no file tool to call.

Bill supervised while reading the streamed reasoning. He let the driver inspect before supplying a diagnosis, because an observer that walks past a defect is revealing something about its view. Pointing at the defect first would erase that evidence. He still supplied operational direction: what to exercise next, which binary to rebuild, and, on one occasion, that a stale server held port 8084 while the new binary was serving nobody. There is little experimental value in letting a driver spend its effort on the wrong process.

Findings need a separate attribution. The report credits Bill with a screenshot observation that led to the truncation investigation, and credits later source inspection with finding unsafe rendering. Neither becomes an unaided GUI discovery because an agent performed the repair. Keeping assistance attached to the result lets a successor judge both the tool's reach and the human work still needed.

## Typing Without Becoming a Human

The first driver runs stopped at the prompt field. Typing caused the coding agent to pause permanently.

The pause feature existed for a good reason. Bill reads reasoning as it streams and sends corrections mid-turn. An engine that keeps deciding while he is composing a hint can make the decision before the correction arrives. The original browser listener therefore paused when nonblank input appeared and unpaused when the field became empty.

The driver followed a different event sequence. `gui_input` set the value and dispatched an input event, so the pause fired. The historical `gui_submit` then sent the prompt and cleared the field by assigning to its `value`. JavaScript assignment does not dispatch another input event. The unpause branch depended on an event that never came. The GUI had observed a user start typing and, as far as its listener knew, continue forever.

The first repair marked programmatic input so the listener could ignore it and made submission explicitly unpause instead of relying on an implied event. That fixed the immediate path. It did not yet preserve the reason for pausing throughout the browser.

In the second-edition work, a real-browser check showed that the model could fill the field and click Send with speech disabled. Independent review enabled native speech. When the utterance finished, programmatically supplied text was still in the field and the Agent stayed paused. The speech callback called `pace()`, which inspected the field's contents anew. The input listener knew the text was programmatic; the later callback silently reclassified it as human typing.

The durable correction keeps the cause rather than reconstructing it from the value. MCP input dispatches a bubbling input event marked with `detail.mcp`, followed by change. The listener updates an explicit `humanTyping` state: unmarked input with nonempty contents represents human typing; marked input does not. Speech and typing combine as `speaking || humanTyping`. Successful submission clears `humanTyping` explicitly after clearing the field. Human whitespace counts as typing. Programmatic text does not become human merely by remaining visible until speech ends.

That distinction has a second boundary: pause belongs to the whole Agent, not to a tab. A browser sends automatic pause and unpause commands as its local causes change. If another tab explicitly unpauses, an unchanged local cause must not immediately reassert pause and veto the human's decision. Local pacing cannot quietly promote itself to permanent authority over the Agent.

The corrected native-browser check covered speech ending while programmatic text remained, genuine whitespace holding pause, explicit unpause from another tab, and a later actual MCP submission. It waited for the concluding speech to finish. Those checks used the real browser and Agent with a local provider fake. Native start/end events establish the measured transitions, not listening quality. The second tab closed after the global-unpause control, and an earlier two-tab speech stall has no established cause. This is evidence for the corrected paths, not a blanket guarantee about concurrent speech.

The failure generalizes beyond typing. Human-facing pacing, debounce, idle detection, and animations all encode assumptions about the cause and timing of events. An automated user can violate those assumptions while producing perfectly ordinary-looking field values. Ignoring one synthetic event is insufficient when another callback later invents its origin again.

## The Escape Hatch Behind the Wait

Once input worked, the historical driver asked the coding agent to write Tower of Hanoi and debug it. The agent produced fifteen moves, started `dlv`, and hung. The book had already warned about precisely this trap.

The command tool synchronously read the PTY until output closed. A short `go build` made that lifetime look harmless. A debugger waiting at its prompt kept the PTY open, so the tool function never returned. The callback pattern could have recognized the prompt, except that pattern matching happened inside `Wait`, and `Wait` was reached only after the tool function returned. The escape hatch was on the far side of the blockage.

The repair separated starting the work from owning its completion. The tool returned once the process was running; a reader goroutine retained the PTY, copied output into the job, reaped the process, and finished the job at exit. The historical dispatcher used a `DeferFinish` field for that handoff. It is an explanation of the old repair, not a new API to introduce now. Chapter 4 already establishes the independent job lifetime that browser access must preserve.

Then the debugger worked and the older Chapter 4 checks failed. Returned tool output was empty, and the capture file contained zero bytes. The apparent output defect was another ownership defect: that codebase had two dispatchers. `Actor.dispatchTool` knew about deferred completion; `Engine.Execute`, used by the older `Ask` path, still finished the job immediately when the tool returned. Finishing closed the capture file while the reader goroutine was writing to it.

Debug prints in the actor dispatcher never appeared. The first suspicion was a stale build cache. The failing path did not call that dispatcher at all. Before rebuilding the world, it is worth finding out whether execution visits the line being instrumented.

With both dispatchers corrected, the report records a complete rerun: a three-disk program produced seven moves, then the agent set a breakpoint on `move`, continued, printed `disk`, and quit the debugger. The report gives a zero debugger exit and ninety-eight seconds elapsed. The earlier hung run used four disks and fifteen moves; they are not one increasingly successful measurement. This is a reported historical outcome, not a new benchmark or evidence of the present checkout.

## Correct on the Server, Wrong on the Screen

The settings test first found the obvious defect. The driver entered a temperature of negative five and the server accepted it. Clamping incoming patches and settings loaded from disk closed both entry paths, including a hand-edited settings file.

Repeating the same interaction exposed the more interesting defect. The server now clamped correctly, but the screen still displayed negative five.

One `Settings` struct was doing two jobs with incompatible meanings for omission. In a sparse patch, an absent field means leave the existing value alone. In a complete broadcast, every field must be represented, including zero and false. The struct used `omitempty`, so the temperature clamped to zero disappeared from the broadcast. The browser updated only fields present in the message and therefore retained the rejected value. The same encoding omitted `TTSEnabled: false`: turning speech off could not be represented in that full-state message.

A comment in `SettingsStore.Apply` acknowledged the false-value problem and routed around it. The method had no callers outside tests. It was evidence that somebody had encountered the defect, not that the state model had been repaired.

There were two reasonable directions depending on compatibility. Separate patch and broadcast types could preserve a wire format with external consumers. In this case the maintainer established that there were none, so preserving that format would retain complexity without protecting anyone. The historical repair removed `omitempty` from the broadcast fields and deleted the unused `Apply`, leaving one merge path with defined semantics. The lesson is not that two types are inherently bad; it is that compatibility should protect an actual dependency rather than an imagined one.

A complete state broadcast must carry `temperature: 0` and `tts_enabled: false` explicitly. Even then, the stored value, the server's applied value, and the client's displayed value are separate observations. A correct file on disk does not prove the screen. Chapter 8's corrected account follows all three. The source also reports a historical check with clamping disabled, but that report is distinct from this edition's settings evidence.

## The Panel That Wasn't There

Asked to exercise sidebar tabs, the driver confidently reported that clicking Artifacts opened Settings. There was no Settings tab in that version of the markup. The tab-to-panel mapping was correct.

The interface was hiding the fact the driver needed. Selected state lived in CSS classes, while `gui_snapshot` built labels from text, `aria-label`, and placeholders without reporting state attributes. The observer could see that tabs existed but could not see which was selected. It also used the ambiguous selector `button:nth-of-type(2)`, because the controls offered little better to aim at. Faced with a request to say what changed, it supplied a plausible panel instead of an observation.

The repair made state explicit. The hamburger exposed `aria-expanded`; tabs exposed `aria-selected`; click handlers kept those values synchronized. Initial state came from the live CSS class rather than a hardcoded default, because markup and rendered defaults had already drifted. The snapshot included the state beside each control.

The historical report says that repeating the same test with the same model and prompt then produced correct before-and-after state and used “not observable” where information was still absent. That is a useful result, not a law that missing information forces every model to invent an answer. In this run, hiding state helped produce a false report. Exposing state supplied the evidence for a correct one.

Accessibility semantics are consequently part of the debugging interface, not decorative metadata. A selector gives an observer somewhere to act. A state attribute gives it something to learn afterward. Both have to describe the live interface rather than the intended default.

Two nearby findings sharpened the distinction between a useful view and a trustworthy one. Source inspection found the tool-card header interpolating tool names and serialized arguments into `innerHTML`, even though the result path nearby used safe `textContent`. Arguments can contain filenames, URLs, and file contents influenced by an attacker. Their arrival through a tool call does not make them trusted markup. The unsafe path needed a rendering repair; the browser driver's other findings did not prove it could discover that path independently.

Bill's screenshot observation prompted the truncation investigation. The initial notes claimed the driver saw the full DOM while the human saw a trimmed view. Measurement found the opposite imbalance. The historical human view capped tool input at 500 characters and results at 1000; the snapshot capped each artifact at 120 and the whole document at 4000, then silently omitted every artifact after the tenth. It could not distinguish ten artifacts from fifty.

Those are historical caps, not a replacement for Chapter 11's revised 4096-UTF-8-byte snapshot boundary. The enduring repair kept bounded views but made their omissions legible. The historical interface retained truncated text through a `title` attribute, and the snapshot reported withheld characters and artifacts. Limits protect attention and context; silent limits manufacture confidence about material the observer has never received.

## What the Later Runs Actually Established

The source records later real-browser sessions with the historically exercised `claude-haiku-5-5`, `gpt-4o-mini`, and `gemini-3.8-flash`. These names identify the recorded sessions, not today's model catalog.

Claude Haiku and Gemini changed the effective font to 20, received a refreshed snapshot, and unloaded the skill. Claude then said the MCP tools remained available until compaction. The final captured request already showed them gone. An old generic unload message had encouraged the wrong explanation, so the message was clarified to distinguish immediate external removal from pending instruction removal. A wording repair did not require another paid run to establish the already-observed removal.

Both GPT-4o mini attempts failed the full task. The first guessed `input[type='number']`, received a missing-selector error, and unloaded without recovering. In a second, guided attempt, it was given the observed `#font-size` selector and told the control was a select. It changed the setting and unloaded, then attempted to read `path/to/the/snapshot` despite an instruction forbidding file access. The nonexistent path failed, and the final response claimed the snapshot could not be inspected, although the captured interaction already contained its refreshed state.

That second attempt passed the script's GUI assertions. It still failed the task. The controls worked; the model mishandled the evidence and violated the requested boundary. The supplied selector was assistance, and a prompt-level prohibition did not provide the confinement of the earlier bare driver.

Those Edition 2 sessions preceded the final native-speech correction and small manual and message corrections. They establish their recorded interactions on their captured source, not a paid rerun of the final source. Speech has its separate native-browser evidence. Effective font size is also not restart persistence.

The source separately records an accepted Edition 3 run that changed font size from 16 to 20 using selectors observed in snapshots. A fresh snapshot and computed styles confirmed the effect, not pixels. After unload, the next provider request omitted all four browser declarations while retaining the manual. The GUI displayed the final response and saved on normal EOF exit. That run did not repeat native speech testing. None of these records is new implementation verification performed for this companion.

Earlier checks have narrower meanings too. The original Chapter 12 grader returned zero after an untyped fixture blocked startup, while its parity detector depended on particular method names. It observed neither successful browser behavior nor the later speech/input failure. The corrected obligations instead follow actual state through both clients, correlated snapshot and queue results, absence of unsolicited observations, interaction, unload refusal, and peer cleanup. Existing command lifetimes remain relevant: the inherited two 62-second header/body waits must finish against the same executable before a compatibility claim can be made. A short timeout is missing evidence, not an abbreviated pass.

## The Book Gets a Bug Report

The historical repairs traveled back into the chapters that should have made their obligations clear. Explaining an interactive job model had not been enough to prevent a tool from waiting for process exit. Showing a tool card had not identified the attacker-influenced text it rendered. Specifying a snapshot cap had not said that the cap must announce its omissions.

The old settings material was especially revealing. It treated the graded surface as a handful of server checks and left client behavior as something visible on screen. But seeing a value accepted does not establish validation, seeing the server's corrected value does not establish the browser's, and a human-readable tab does not establish machine-readable selected state. Several defects fit inside those unstated obligations.

That does not make every implementation mistake the book's fault. It does make the instructions worth repairing when an experienced implementer could have inferred the missing duty and did not. The experiment produced changes to both the software and the teaching that would shape the next implementation. It is a concrete feedback loop, not evidence of unrestricted autonomous self-improvement.

Even the account of that loop needed correction. The first-edition chapter said the sentence “You know the client works because you can see it.” had been invented and was absent from the original settings chapter. The source chapter reports that the exact words were there, split across a newline, including in the revision before the narrative was appended. A one-line search had missed a wrapped sentence.
