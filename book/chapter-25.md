# Chapter 25 — Goals

I would give Ensemble a job that ought to keep it occupied for most of a day. An hour later, it could be sitting there doing nothing. Codex had done the same thing. Neither program was necessarily broken. The model had finished a response, but the job was not finished. Work resumed when I noticed and poked it.

I did not want to spend the day being the agent's timer.

There was a second problem. A long job can pass through many context compressions. The model may retain the latest debugging detail and lose the reason it was debugging in the first place. Another poke is not enough if the agent no longer has a reliable statement of what it is supposed to accomplish.

Goals addresses these two problems separately:

1. Keep the agreed objective in a file and read its current contents into every ordinary model request.
2. Optionally give an idle Agent another turn in which to inspect its state and decide what to do next.

The first mechanism preserves an objective across compression. The second supplies an opportunity to continue. Neither guarantees progress, correct judgment, or permission to act.

> **C# edition boundary.** This implementation was originally built on the preserved `solutions/ch23` endpoint. The optional Codex path and Gateway have since been completed locally; their historical endpoints retain their original names. Goals does not depend on either integration. The final crossover exercise in Chapter 26 still requires the human's replacement acceptance.

## §25.0 TL;DR — What to Build

Build an opt-in, session-local Goals feature on the existing Agent actor. Do not add a scheduler service, a goal database, a second model conversation, or a browser-owned timer.

### The control

Install the `goals` skill as a loadable skill. It grants one tool, `goal_control`:

| Field | Meaning |
|---|---|
| `file` | Goal-file path. Omitted means unchanged; an empty string clears configuration. |
| `enabled` | Whether idle reminders are enabled. Omitted means unchanged. |
| `idle_seconds` | Positive finite idle delay, default 480 seconds. |

Calling the tool without fields reports status. Loading the skill grants the control but creates no file and enables no reminders. Discovery is not activation. The ordinary legacy library catalog must not accidentally grant this new tool just because it permits older tools without a Skills manager.

The public C# operations are:

```csharp
await agent.GoalControlAsync(new GoalUpdate(
    File: goalPath,
    Enabled: true,
    IdleSeconds: 480));

GoalStatus observed = await agent.GoalStatusAsync();

// Stop reminders without removing instructions or claiming completion.
await agent.GoalControlAsync(new GoalUpdate(Enabled: false));

// Remove the goal from this session. Do not delete the user's file.
await agent.GoalControlAsync(new GoalUpdate(File: ""));
```

Configuration requires an active Goals grant. `GoalStatusAsync` is read-only and can report that the skill is unloaded. Neither form of status query restarts the idle deadline.

### The file and request

Use the Agent's existing Sandbox for file access. Validate a newly selected file before publishing the configuration. Read it again on each ordinary request, including another request in the same tool loop. Do not cache its text in Skills, History, or the GUI.

The C# implementation accepts nonblank, strictly decoded UTF-8 text up to one MiB. It rejects directories, nonregular streams, invalid bytes, and paths outside the Agent's confinement. Selecting an invalid file leaves the previous configuration unchanged.

A configured file supplies request instructions even when reminders are off. Clearing the file or unloading Goals removes that injection. A fresh Agent starts without the previous session's configuration, and a child does not inherit its parent's goal.

If an existing configured file becomes unreadable, disable reminders and expose a stable error. Ordinary human work can proceed with an explanation of the file problem. An automatic wake must not send inference without its goal. Off and Clear must remain usable when the file is missing.

### The idle reminder

Default to eight minutes. The timer callback only posts actor mail carrying a generation token. It does not call the provider, append History, or execute tools.

The actor accepts the wake only if the token is current, the configuration is enabled, the Goals grant remains active, and the Agent is still idle. Active foreground work, admitted turns, controls, startup work, and direct Engine sends prevent an automatic wake. Independent Jobs can remain alive; the reminder instructs the model to inspect them rather than cancel or duplicate them.

Pause suspends the timer. Resume starts a fresh interval. Interrupt disables reminders, even when there is no active turn. Shutdown, actor failure, successful delegated submission, unload, and destruction must not leave a later automatic turn behind.

### The GUI

Show the selected Agent's status, not the host actor's status by accident. Distinguish unloaded, unconfigured, working, waiting for a reminder, paused, reminders off, and file error. Show disconnected or switching state honestly rather than presenting an old observation as current.

Provide labelled Apply, Reminders off, Clear, and Reload controls. Acknowledgment means the actor applied the update. The browser must not optimistically claim success when a checkbox changes. These controls are session-local; they do not write the durable memory/recall policy or appearance preferences.

## §25.1 Write an Objective, Not a Repeated Command

A useful goal file tells the agent what remains, how to recognize completion, and where its authority ends:

```markdown
# Goal — finish the parser correction

The parser accepts a malformed trailing record. Reproduce that defect with
an owned local fixture, correct it, and review the complete changed call path.

Done means the failing case passes, the relevant existing tests still pass,
and the explanation describes the actual behavior and remaining limits.

Do not install packages, modify credentials, restart the running application,
or commit changes. Preserve unrelated work. If a healthy test job is already
running, observe that job rather than launching another copy.

Disable reminders when finished or genuinely blocked on a human decision.
```

“Run the tests again” is not a good long-running objective. It does not tell the agent whether the first test process is still running, whether a failure is already understood, or whether another run would answer a useful question.

The ordinary file tools write and edit the document. `goal_control` only names it and controls the reminder. Keeping these responsibilities separate avoids turning a small control tool into another document store.

The path is not the objective. The file's current bytes are. A handoff can replace working conversation without erasing those bytes. On the next request, the renderer reads the file again.

That does not make the file magical. It can be stale, badly written, or missing. Keep it short enough to understand, update it when the work changes, and do not put secrets in a document that will be sent to the model.

## §25.2 Keep the State on the Agent

`Common/Goals.cs` declares the vocabulary and operations. `Agent/Goals.cs` implements them as another partial section of the same Agent type.

The persistent runtime state is small:

- One immutable `GoalConfiguration`, synchronized by the Agent's ingress lock.
- One actor-owned timer, deadline, and generation number.
- Immutable update, status, and read-failure values passed across the existing boundaries.

The configuration contains the path, enabled flag, interval, and last safe diagnostic. It does not contain goal text. `GoalStatus` is an observation of those facts and the Agent's actual activity, not another mutable configuration owner.

This feature does not need a relationship-bearing `GoalManager`. It has no independent lifetime: it belongs to the Agent's session. The existing Agent already owns admission, pause, interruption, shutdown, Skills, History, and Sandbox. Reach those services through that Agent rather than copying them into a callback bag.

Configuration runs through the actor mailbox. Validate the complete proposed update before publishing it. A bad interval or unreadable replacement file must not partly change an otherwise working goal.

The timer accepts a finite interval of at least one millisecond and no more than 4,294,967.294 seconds, matching the chosen .NET timer's millisecond range. Those are implementation bounds, not useful recommendations. Eight minutes is the default; a two-second interval belongs in an owned test, not an unattended production job.

Unloading a skill has two distinct effects in this system: its authority is revoked immediately, while its old manual can remain in canonical History pending later retirement. Goals uses the current grant, not the presence of an old manual. Unload also clears its session configuration and timer. Removing the Skills relationship does the same without deleting the goal file.

## §25.3 Preserve the Goal Across Requests, Not as Transcript Copies

The renderer already projects primary instructions, tool policy, retained manuals, and request-only pressure guidance. Goal text joins that request-only instruction projection.

The four supported surfaces place it in their real instruction channels:

| Surface | Placement |
|---|---|
| OpenAI Chat Completions | System message |
| OpenAI Responses | `instructions` |
| Anthropic | `system` |
| Gemini | `systemInstruction` |

This placement is part of the feature. A tool result somewhere in an old transcript is not equivalent to current system instructions. Nor is a file path whose contents the model has not read.

Read the file for each render. Suppose a turn reads the old objective, edits the goal file, and calls `micro_handoff`. When that same turn makes its next model request, the new objective must appear. Requiring another human message would defeat the continuation mechanism precisely when it is needed.

Do not append the file body to canonical History each time. The runtime retains its configuration; History retains the real dialogue and effects. Repeated file copies would create stale competing objectives and unnecessary context growth. As always, a user or tool can deliberately put text into dialogue; the Goals feature itself does not create those transcript copies.

Auxiliary recall judges and memory compressors do not receive the parent's goal instructions. They have narrower tasks and existing request builders. A compressor that inherits the main agent's day-long objective is liable to work on that objective instead of compressing its input.

File errors require a careful split. A human turn should be able to say, “The configured goal file is unavailable; repair it and explicitly re-enable reminders.” An automatic wake should not send a request with its objective silently missing. The wake preflights the read, and rendering checks again before send. A revision-bearing failure cannot disable a newer configuration selected while the old read was in progress.

## §25.4 Let the Actor Decide Whether to Wake

The timer callback is deliberately uninteresting:

```text
post GoalWake(generation)
```

The actor does the interesting work:

```text
Is this still the current timer generation?
Are reminders still enabled?
Is Goals still granted?
Is the Agent still idle?
Can the configured goal be read now?
```

Only after those checks does it admit a normal turn containing a clearly labelled automatic reminder. The canonical input is ordinary data identifying its runtime origin; the fresh goal is separately projected as instructions. There is no invented human reply and no special completion event.

The reminder asks the model to inspect actual state, continue authorized unfinished work, avoid repeating commands blindly, and leave healthy jobs alone. Synthetic reminders skip user-query recall: they are not a new user's search question.

Why the generation token? Disposing a timer does not prove that its callback was never queued. If the user changes the interval, pauses, unloads the skill, or interrupts near the deadline, old mail may still arrive. A stale token must be harmless. A callback is an opportunity to check, not authority to run.

Why use the existing admission path? A reminder is real work. It must respect queued input, accepted-work bookkeeping, tool limits, pause, cancellation, and the same provider lifecycle as an ordinary turn. Bypassing admission to call Engine directly would create a second scheduler inside the first.

Direct library sends count as busy through the existing Engine-send counter. Starting or ending one posts a refresh opportunity to the actor. The Goals code does not maintain another count of provider activity.

Independent Jobs are different. They may legitimately run while the foreground Agent is idle. A reminder can help the Agent observe one, but it is not a timeout or a license to restart it. The model still has to distinguish waiting for healthy work from having forgotten the task.

The interval starts again after work becomes idle. This is not “run every eight minutes,” and it is not a watchdog for a hung active request.

## §25.5 Stop Has More Than One Meaning

The interface must not collapse these states:

| Action or event | Effect |
|---|---|
| Load Goals | Grants control only; no file or timer |
| Set a file | Configures instructions; does not implicitly enable reminders |
| Enable | Validates the file and arms when idle |
| Off | Stops reminders, retaining goal instructions |
| Clear | Removes session configuration/injection, retaining the file |
| Pause | Suspends scheduling, preserving configuration |
| Resume | Starts a fresh idle interval if enabled |
| Interrupt | Disables reminders and applies normal turn interruption |
| Successful delegated submission | Disables reminders without changing the submission or parent's goal |
| Unload/remove Skills | Revokes control and clears the session goal |
| Shutdown/fault/destruction | Invalidates timer activity and completes normal joined cleanup |

**Off is not completed.** A task can be blocked on credentials, permission, or a human decision. Stopping reminders is correct in that state, but the result is still incomplete.

**Submitted is not idle.** A delegated Agent's accepted submission is an authoritative result. A later automatic turn would admit new work and invalidate that result. Disable the goal at successful submission publication, then let the actor dispose its timer. A refused submission must leave the goal alone.

**Restart is not resume.** The file remains on disk, but configuration is session-local. A new Agent needs an explicit setup. Children have their own sessions; the parent's file and enabled flag are not copied into them.

In earlier runtime work, a missing visible explanation made a real failure look like an unexplained pause. That is a useful lesson here, but it is not a claim that the historical provider failure was reproduced in C#. If a file becomes unreadable, status should say so. If the socket disconnects, the GUI should say its current knowledge ended there.

## §25.6 One Status, Several Views

`GoalStatusAsync` lets a trusted host observe the Agent even when the Goals skill is unavailable. It grants nothing. Configuration and model dispatch still require the skill.

The GUI uses this query in its existing Agent-tree refresh. It does not start a second status scheduler. Each row carries that Agent's status, and the selected-Agent panel displays the corresponding row.

The socket controls are deliberately small:

- `get_goal` reads the selected Agent.
- `update_goal` submits a strict patch to that Agent.
- An acknowledged result carries the Agent identity.

Decode the whole patch first. Reject duplicate names, unknown names, explicit nulls, wrong types, and invalid delays. Catch file-selection refusal at this control boundary without exposing operating-system paths or closing an otherwise usable socket.

Await actor application before acknowledging. Broadcast updates only to tabs viewing that Agent. A queued frame for the previously selected Agent must not overwrite a new view. An in-progress switch shows that status is pending; a lost socket shows that current state is unknown.

The form has a draft, not a second configuration. Periodic status frames must not erase what the user is typing. Apply sends the draft; the acknowledgment repopulates the controls with accepted values. Reload discards the draft in favor of an observation. Off and Clear remain distinct actions.

Working and reminders enabled are separate facts. An Agent can continue its current request after reminders are turned off. Show “Working — reminders off,” not “Off” as if the request had disappeared. Likewise, the displayed deadline is the core's observed deadline, not a browser countdown that promises a future callback.

Use text labels and ordinary keyboard-accessible controls. The local browser exercise checks their DOM behavior; it does not establish that a particular person can see or hear them comfortably. Human accessibility and audible behavior still require human use.

## §25.7 Verify the Boundaries

Use real Agents and the existing actor, renderer, Sandbox, skill catalog, and GUI. Replace the external provider with a small loopback HTTP peer only where a request is required.

The important checks are causal:

1. A plain legacy Agent neither declares nor dispatches `goal_control`. Loading Goals grants it without configuring a timer.
2. Refused configuration preserves the old path, enabled flag, interval, and deadline. Off/Clear work after the file disappears.
3. All four provider renderers receive fresh instructions. The file body is not appended by Goals to History.
4. A same-turn handoff followed by another request reads the edited file.
5. A real idle wake sends one labelled reminder, while pause, foreground activity, and held HTTP work prevent overlap.
6. A stale timer after reconfiguration, unload, or shutdown cannot start a later request. An unreadable automatic wake makes no HTTP request.
7. A successful child submission survives without a later reminder; a refused submission does not disable Goals; parent and child settings remain separate.
8. Real sockets exercise availability, selected-Agent isolation, malformed patches, reconnect, file errors, Off and Clear. Status polling must not reset the deadline.
9. An isolated browser configures a goal, makes a human request while reminders are off, edits the file, observes one idle reminder, turns it off, reconnects, clears it, and observes disconnect honestly.

The local provider can verify actual request bytes and lifecycle. It cannot establish that a real model follows the goal well, knows when it is blocked, or avoids an unnecessary retry. Those remain application and human-judgment questions.

The numbered Chapter 25 graders in earlier editions are **Gateway** graders. Do not rename their result as Goals evidence. This chapter has focused C# component, socket, and actual-host checks, not an invented inherited score.

### Source map

Paths are relative to `agent`:

| Responsibility | Source |
|---|---|
| Goal values, state and callable declarations | `Common/Goals.cs`, `Common/Actor.cs` |
| Configuration, reading, status, timer and wake validation | `Agent/Goals.cs` |
| Admission, pause, interrupt and joined actor cleanup | `Agent/Actor.cs` |
| Direct-send activity | `Agent/Models.cs` |
| Successful submission publication | `Agent/ChildSubmission.cs` |
| Request-only projection and explicit Goals declaration gate | `Engine/Render.cs` |
| Tool schema/control dispatch | `Tools/Tool.cs`, `Jobs/Jobs.cs` |
| Opt-in manual | `Skills/goals/SKILL.md` |
| GUI adapter and controls | `Gui/Server/Goals.cs`, `Gui/Server/Host.cs`, `Gui/wwwroot/app.js` |
| Core and real-socket tests | `Tests/GoalTests.cs`, `GuiTests/GoalControlsTests.cs` |

## §25.8 Taking It for a Spin

Start with the local checks. From the repository root, using your approved .NET 10 SDK:

```sh
dotnet build agent/Tests/Tests.csproj
dotnet agent/Tests/bin/Debug/net10.0/Tests.dll --goal-tests

dotnet build agent/GuiTests/GuiTests.csproj
dotnet agent/GuiTests/bin/Debug/net10.0/GuiTests.dll --goal-tests
```

These tests use owned temporary files, local sockets, and boundary fakes. They do not need provider credentials. The GUI test expects the repository root as its working directory so it can copy the shipped Goals manual.

For a personal trial, use a new workspace, an owned port other than the port of your running application, and a provider/account configuration you have explicitly approved. Do not borrow another running Agent's credential store. Start with a tiny objective and a short, supervised interval; reminders can cause inference and spend money.

This setup prepares an isolated primary skill that permits Goals. Run from the repository root:

```sh
main="$PWD/agent"
work=$(mktemp -d "${TMPDIR:-/tmp}/ensemble-goals.XXXXXX")
mkdir -p "$work/skills/base" "$work/skills/goals"
cp "$main/Skills/goals/SKILL.md" "$work/skills/goals/SKILL.md"
cat > "$work/skills/base/SKILL.md" <<'SKILL'
---
name: base
type: primary
description: Supervised Goals trial
tools: read_file write_file
loadable-skills: goals
---
Explain what you intend to do before tools. Use only the authorized workspace.
$TOOLS
$SKILLS
SKILL
cat > "$work/GOAL.md" <<'GOAL'
# Supervised Goals trial

Read the current objective and report the next safe step. Do not install,
restart, deploy, commit, search for credentials, or make external changes.
This is a demonstration, not permission to invent additional work.
After one automatic reminder, disable reminders and report the result honestly.
GOAL
printf 'Workspace: %s\nGoal file: %s\n' "$work" "$work/GOAL.md"
```

Build the actual GUI host:

```sh
dotnet build "$main/Examples/Gui/Gui.csproj"
```

After explicitly configuring your chosen provider and selecting an unused local port, launch from the temporary workspace. Supply `GOALS_PORT` yourself rather than copying a port already used by another instance:

```sh
: "${GOALS_PORT:?Choose an unused local port, not the existing application port}"
(cd "$work" && dotnet "$main/Examples/Gui/bin/Debug/net10.0/Gui.dll" \
    --port "$GOALS_PORT" --data-dir "$work/data" --agent-name trial \
    --skills-dir "$work/skills" --primary-skill base --startup-skills goals)
```

The omitted provider configuration is intentional: a chapter example must not silently choose an account, credential source, endpoint, or billable model for you. This demonstration is separate from the completed, bounded OAuth/cache exercise recorded in Chapter 23.

Open the owned GUI URL and the Goals controls:

1. Confirm **No goal configured**. Startup loading the skill has not enabled anything.
2. Enter the printed goal-file path, leave reminders unchecked, and Apply. Observe the acknowledgment and **Reminders off — goal retained**.
3. Send a short human request. Inspect an authorized request trace if you need evidence of instruction placement; the model's answer alone does not prove what was sent.
4. Edit the goal file. Send another request or enable a short supervised reminder and observe the fresh objective.
5. Pause before a deadline and verify that inference does not start. Resume and observe a fresh interval rather than an immediate accumulated alarm.
6. Turn reminders off while work is active. The current work remains work; only future idle wakes stop.
7. Reload the page. The same running Agent still owns its configuration. This is not process-restart persistence.
8. Clear the goal. Confirm the file still exists. Close the owned host through its normal shutdown path.

If a test or provider request is already running, inspect it instead of starting another copy just to obtain a newer timestamp. When finished or genuinely blocked, turn reminders off and say which state applies.

## §25.9 Exercises

### Exercise 1 — Freshness in the Same Turn

Use a boundary provider that returns a tool batch editing the goal and calling `micro_handoff`. Capture the next request. Show that it contains the new objective without requiring a new human message. Also show that Goals did not append the file body to canonical History.

### Exercise 2 — The Stale Callback

Arrange for a timer callback to become eligible, then reconfigure or unload before the actor accepts it. Show why disposing the timer alone is insufficient. Assert that the old generation produces no request.

### Exercise 3 — Submission Is a Boundary

Give a delegated Agent an independent goal. Refuse one malformed submission, then accept a valid one. Verify the goal stays enabled after refusal and stops after successful publication, while the parent's configuration is unchanged. Do not equate ordinary final prose with submission.

### Exercise 4 — Status Without Authority

Query an Agent with Goals unloaded. The GUI should show that state, not grant the control to obtain status. Then load the skill and verify that no configuration appeared as a side effect.

### Exercise 5 — Two Tabs, Two Agents

View different Agents in two tabs. Apply a goal in one, switch the other, and introduce a delayed old status frame. Demonstrate both identity filtering and that an unsaved form draft is not periodically replaced by a status refresh.

### Exercise 6 — Missing File, Honest Outcome

Remove a configured file before a wake. Verify zero inference from that wake, an explicit file-error status, and usable Off/Clear controls. Restore the file and explicitly re-enable. Do not report the interrupted task as completed.

## §25.10 What Goals Does—and Does Not—Buy Us

The goal file carries a durable objective. The runtime brings its current contents into ordinary requests. An opt-in timer gives an idle Agent another chance to examine the work.

That is useful, but it is not self-certifying autonomy. A reminder can encounter a real blocker. The model can misunderstand the objective. A healthy independent job may still need time. The process can be stopped, the file can disappear, or an active request can remain stuck without becoming idle.

I require the agent to inspect, distinguish those states, and explain its decision. Continue when there is useful authorized work. Wait when waiting is correct. Ask when a human decision is genuinely needed. Stop reminders when the goal is complete or human-dependent—and preserve which of those happened.

Goals should reduce the number of times I have to poke the agent. It should not reduce the amount of judgment in the work.
