# Chapter 8: Build Your Dream GUI

The settings server saved a speech rate of 1.75. After a restart, the label still said 1.75. The native speech API received approximately 1.8.

In the [source chapter’s account](../edition-3/chapter-08.md#the-saved-number-that-speech-never-used), an independent reviewer followed the value past the successful WebSocket echo. The browser’s range input had a step of 0.1, so Chrome rounded the saved value to a legal slider position. The label read the settings object. The speech queue read the slider. Each had a plausible source of truth, and the person using the application got the difference.

Chapter 7 established a view from which a human can watch and steer one Agent. This chapter makes that view somewhere a human might choose to work: separate chat and action panes, a sidebar, adjustable widths, coherent themes, and preferences that survive a process restart. The additional server machinery is small. The judgment lies in deciding which state belongs where, when a change takes effect, and what evidence shows that it reached its destination.

For Bill, speech is central to using the agent, not an ornamental accessibility checkbox. The broader goal is to start doing useful coding inside the new system early enough that daily use exposes its missing features.

## Two Views of the Same Work

The workbench has a left sidebar, chat in the center, and actions on the right. Two dividers control the widths. This is a change in presentation, not a new execution architecture.

Chapter 7’s `ArtifactScroll` now has two instances. User messages, readable thinking, and reply text go to Chat. Tool-call argument streams and their final parts go to Actions, followed by dispatch and finish observations and tool output. Status messages update workbench controls rather than becoming transcript content.

The routing must recognize both a part’s kind and the messages that later update it. A tool’s arguments can arrive before its dispatch observation. Preserving artifact and call identity lets dispatch update the card already showing those arguments; it must not create a second, apparently unrelated execution. The same rule applies to replay. Moving content to a different pane does not change what event it represents.

Both panes retain the same streaming protocol, renderer lifecycle, and styling vocabulary. Routing is a filter over one component, not permission to build a second tool renderer.

Each pane scrolls independently. Following a long tool result should not drag the human away from a reply being read in Chat, and the follow control needs its own space rather than covering transcript text.

The dividers are similarly modest. A drag changes a width, with a minimum preventing a pane from collapsing into uselessness. The accepted page described by the source uses pointer capture so a gesture continues after the pointer leaves the narrow divider. Releasing the gesture ends its tracking. Focused dividers also accept arrow keys and expose their current value through separator attributes.

There is no prescribed visual style. There is a requirement for judgment: strong contrast, readable text, clear grouping, and tabs whose selected state is unmistakable. A row of generic buttons may switch panels correctly while failing to communicate that the panels belong to one workspace. Browser measurements can establish geometry and control behavior. They cannot decide whether this is a place Bill wants to spend the afternoon.

## The Settings File Is Not the Agent

The sidebar switches between Agents and Settings. Four visible controls cover theme, automatic speech, speech speed, and font size. They send changes through the existing WebSocket. No model call is necessary to remember a font size.

The source records an ownership correction here. Browser settings and their file handling were initially placed in the core Agent, then moved back into the optional GUI module. The Agent owns its configuration, but that does not make every configurable thing an Agent concern. A theme does not become execution policy by sharing a JSON object with it.

The GUI server owns the browser preferences, their mutex, and their persistence path. The application chooses that path; the reference command uses `cr/settings.json` beneath its working directory. Provider credentials remain in the intended process environment, not in this preferences file, logs, commands, or browser controls. There is no API-key or model-selection panel in this chapter.

One field crosses the boundary deliberately: `max_tool_rounds`. It already belongs to the Agent’s execution policy and travels in the saved settings envelope so a restart can restore it through the public Agent interface. It does not need a fifth visible control. Settings snapshots obtain that policy from its Agent owner rather than allowing the GUI to maintain a competing runtime value.

The distinction matters during failure diagnosis. The optional server can remember and display the requested budget without the actor ever using it. Conversely, an execution policy can work perfectly while the browser falsely reports a different value. Storage, display, and execution are related responsibilities, not interchangeable proof.

Restart has a narrow meaning here, too. Reusing the working directory restores preferences and the next-turn budget. It does not restore a saved conversation. Chapter 7’s browser reconnect recovers a view of a still-running Agent; this chapter’s settings load survives a new process. Neither supplies general Agent save/load.

## A Patch Comes In; a Snapshot Goes Out

Consider a tab turning Read aloud off. Its `update_settings` message contains `tts_enabled: false`; it need not resend the theme, font, speed, or budget. The server interprets this as a patch over the accepted settings. After a successful update, `settings_changed` carries a complete snapshot to every connected client, including the sender. A new subscriber receives `current_settings` alongside the existing history and active-part handshake.

Absence means different things in those two directions. In a patch, an absent field means “leave it alone.” An explicit false means “turn it off.” An explicit zero can select a policy default. In a snapshot, those values are part of the answer and must remain present.

An ordinary Go boolean with `omitempty` cannot represent both an absent patch field and a request to set false. The small adequate design decodes a sparse object over a copy of the current accepted value, or tracks presence explicitly. The full snapshot omits neither false nor zero. Otherwise, the server can accept a change and then send a reply incapable of describing it. A client may keep showing the old value, or even a number the server has just clamped away.

The copy is also an error boundary. JSON decoding can encounter a type error after reading earlier fields. Decoding directly into live state would let a rejected message leave half its changes behind. Decoding into a candidate keeps the accepted state intact until the whole candidate is valid. Input must be a JSON object, not merely something the JSON decoder can consume.

Validation occurs on entry, for both wire updates and disk loads. The same path checks the theme vocabulary and clamps the bounded preferences. In the reference design, an unrecognized theme falls back to dark. The particular font-size endpoints are implementation choices, not a new core policy. A missing settings file starts from defaults; malformed saved JSON produces a startup error instead of silently discarding the person’s preferences.

The source’s first-edition audit found a temperature of −5 reaching the engine because both update and load paths trusted their input. That is a historical lesson about validation boundaries, not a reason to add a temperature control here. Checking only at the point of use would leave invalid accepted state available to every other consumer. Checking only network input would let disk restoration bypass the repair.

The budget has two boundaries with different duties. The public Agent interface rejects negative budgets. The GUI’s shared validator may normalize a negative saved or incoming value to zero before calling that interface. The GUI accommodates its input; core protects its invariant. Teaching the Engine to repair browser preferences would muddle both responsibilities.

## Save Before Announcing Success

A valid candidate is not yet an accepted update. It still has to be saved.

The original settings sketch ignored decoding and write errors while mutating live state. That creates an attractive failure mode: the current tab works, other tabs agree, and a restart discovers that none of them was describing durable preferences.

The accepted source design writes a small replacement file and renames it over the destination before publishing the candidate. If writing or renaming fails, the previous accepted value stays active and no success snapshot goes out. The server reports the failure, and the page displays the existing `settings_error` reply. After a successful save, the server restores the budget through the public Agent setter and publishes the complete accepted settings.

The existing server mutex orders application, persistence, publication, and new subscriptions. A client subscribing during an update should see a coherent accepted value, not a mixture of the old snapshot and the new policy. This is coordination around one small file, not a transaction framework. Replacement by rename is not a claim of power-loss durability.

Using the existing WebSocket keeps settings on the same ordered connection as observations and subscription state. Adding a REST settings path would require reconciling an independently timed fetch or update with that stream, especially after a tab sleeps. The source avoids that extra reconciliation rather than declaring that two transports can never work.

This does not upgrade the socket into a reliable-message service. Chapter 7’s slow-client progress limits remain. Broadcasting a successful change to connected clients and supplying the accepted snapshot on subscription are the intended synchronization paths; they are not a promise that every overloaded tab sees every intermediate update.

## A Budget Belongs to a Turn

The source records that `max_tool_rounds` once persisted correctly while two orchestration loops ignored it and used a hard-coded limit of 16. The GUI could faithfully save 200 while execution stopped at sixteen. This defect survived longer than the early settings audit because all the storage behavior looked healthy.

Chapter 5’s design now gives the policy one owner. Zero selects the named default of 200; positive N permits N tool batches per human turn. The unit is a batch, not an individual tool call, a provider request, or the total number of processes still alive. The Agent captures the setting at turn start.

Capturing it matters because a settings update is not an interruption command. Suppose a turn starts with a two-batch budget and another tab raises the saved budget while that turn is running. The settings snapshot can immediately report the newly accepted setting, but the active turn still has its original allowance. The next turn receives the new one. Reading live configuration at every dispatch would let a different tab quietly rewrite the bounds of work already under way.

At the captured limit, the actor retains Chapter 5’s final-response allowance. A stop-tools instruction goes into request-local prompt context so the model can finish its answer. It does not alter History, modify signed content, or leak into later turns. Remembering a budget must not corrupt the conversation used to enforce it.

The source describes an entry-point test that connects these pieces rather than inspecting the saved JSON alone. It saved a two-batch limit and restarted the actual two-view application. While the first provider response was held, a settings update raised the saved limit to eighteen. The active turn stopped after two batches; the next terminal turn completed eighteen. A further restart retained eighteen, and zero allowed execution beyond the old sixteen-batch ceiling by selecting the default of 200.

Only the external provider was fake in that test. The actor, tools, terminal input, WebSocket, and settings file remained real. Restoring the live-read defect made the active turn execute eighteen batches instead of two. That is evidence about when the policy takes effect, not evidence of a live provider’s capabilities.

## Controls Must Describe Their State

Theming is mostly a useful exercise in keeping presentation centralized. Semantic CSS custom properties provide colors for surfaces, text, borders, controls, code, errors, and tool cards. Changing the active theme changes those values without reconstructing the transcript. Dark is the source’s default because this workbench was built around Bill’s speech-oriented use; that is a design choice, not a universal accessibility rule.

A saved `system` preference is different from a saved dark or light choice. The page resolves `prefers-color-scheme` to a concrete theme and listens for changes while system mode remains selected. A one-time lookup follows the operating system only until it changes. The saved preference remains system even though the displayed colors are currently dark or light.

Control state also needs a semantic representation. A CSS class can make a toggle look selected while telling assistive technology and automated observers nothing about its state. The relevant control attribute must say what is selected, enabled, or pressed; CSS then styles that state. The same distinction applies to sidebar tabs and dividers.

The agent tree supplies a small but memorable example of false agreement. In the source’s implementation history, both the CSS and an initial browser check used state names copied from an example: `thinking` and `tool_use`. The actual reducer emitted `in_flight` and `tools_pending`. The stylesheet and the test agreed with each other. The active indicator did not pulse.

The corrected check inspected the active node’s computed animation using actual Agent state. Restoring the old selectors broke it. An attribute that a test wrote itself was not evidence that production events reached the rendering rule.

The sidebar’s data model is a tree, and its renderer walks children recursively, but there is currently one node. Its indicator becomes active for working states and dims when idle. An empty children array leaves room for later display growth; it creates no child Agents, assigns no ownership, and establishes no sub-agent lifetime or protocol.

## Speech Has Several Consumers

Settings application handles both the initial `current_settings` and later `settings_changed` snapshots. It updates visible controls and the behavior behind them. Updating only the change path would make a newly opened tab display defaults until someone touched a setting.

Read aloud controls automatic speech. Turning it off cancels that speech, but explicit Speak requests still play. A changed rate applies to the next utterance, not retroactively to sound already under way. These distinctions preserve the difference Chapter 7 established between automatic notification and a human asking to hear a specific artifact.

The interaction with typing is more subtle. Changing Read aloud releases the old typing pause without deleting the draft. Fresh typing must still veto automatic continuation, even if an acknowledgement of an earlier change arrives afterward. Explicit Continue in any tab releases that veto. An implementation that merely clears pause whenever a settings echo arrives can erase a newer human action with an older reply. The text survives, yet the boundary the human expected to hold has opened.

The fractional-rate failure at the start of this chapter lives at a different boundary: the saved preference entering a native control. The correction was to let the slider preserve arbitrary fractions with `step="any"`, with a comment explaining why tenths were insufficient. Changing the label would not have helped; the label was already right.

The source’s focused check saved 1.75, restarted the real application, inspected the snapshot, label, and slider, then submitted browser input and recorded the native utterance’s rate. It repeated the sequence with 1.2345 so a smaller fixed step could not merely relocate the defect. The [historical baseline](../../solutions/evidence/ch09/speed-r1.json) and [restored-defect record](../../solutions/evidence/ch09/speed-r1-mutation.json) are linked by the source. With the old step restored, the two requested values became approximately 1.8 and 1.2 at the native API. Both correct and defective utterances still produced start and end events. Completion was never proof of the requested rate.

A separate failed check confused floating-point representation with a product defect. A saved 0.6 arrived at the native API as `0.6000000238418579`, so exact equality timed out. The focused regression used a tolerance of 0.00001: enough for that representation noise, far too small to hide the rounding from 1.75 to 1.8. The source also retains an earlier timeout without enough location detail to identify its cause. It does not recruit that failure as additional evidence for a diagnosis it cannot support.

There is one further distinction between the historical demonstration and the current source design. The fractional-rate check used the earlier control range, which topped out at 2. The later crossover extended the saved preference and control to 5, while every native utterance construction clamps the rate to 0.5–3. Thus a saved 2.5 reaches native speech as 2.5; a saved 5 reaches it as 3. The old test establishes fractional preservation within its tested range, not verification of that later extension.

Preserving fractions in the control and clamping at the native consumer solve different problems. The first prevents accidental quantization of accepted input. The second deliberately maps the application’s approximate preference into its chosen native range. That mapping must cover both automatic speech and explicit Speak. Native voice rates are nonlinear, so a value labelled 5 is not a measured fivefold speedup.

Snapshots establish what was accepted. Slider values establish what the browser control retained. Utterance rates establish what the native API received. Native callbacks establish API activity. None establishes that a person listened and found the result intelligible. For a speech-centered workbench, that final judgment is not an optional embellishment.

## Usable Enough to Expose the Next Problem

The source retains browser work covering two-tab settings changes, system-theme changes, pointer and keyboard dividers, pane routing, the active Agent indicator, and visible restoration after restart. Those historical checks used the earlier speech control range. The separate fractional-rate investigation matters precisely because a tenths-aligned value in the broader browser run did not expose rounding.

It also records a historical live OpenAI turn: saved controls restored after restart, real shell output appeared in Actions, the answer remained in Chat, and a terminal hint reached the same Agent. The answer’s speaker produced native callbacks. Unchanged provider coverage was reused from Chapter 7 instead of paying for a fresh matrix merely because settings had changed. These are results from a specific historical run, not present provider-availability claims.

Bill’s stated destination in the source is to run GPT-6 Astra at high reasoning in the new Ensemble and use that agent, rather than Codex, to build Edition 3. That is a target, not a report that access, reasoning configuration, or the coding crossover has been verified. A usable workbench brings the target closer by making ordinary work possible; it does not certify that the switch has happened.
