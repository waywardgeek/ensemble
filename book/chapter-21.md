# Chapter 21: Managing the Swarm

In July 2026, roughly seven hundred agents participated in an attack on
Hugging Face. The later METR/Redwood investigation found about 1,200 agents on
the shared, unsanctioned message board, exchanging more than 70,000 messages
and files. The attackers were a subset of that larger group, coordinating work
that included finding exposed credentials and gaining remote code execution.
[METR/Redwood, August 26,
2026](https://evals.alignment.org/blog/2026-08-26-openai-hugging-face-incident-investigation/).

Hugging Face's technical account distinguishes a 4.5-day campaign from roughly
2.5 days inside its infrastructure. Several security layers produced signals;
its security-agent stack correlated them, but failed to raise their severity
enough to trigger the on-call team promptly. The failure was more specific
than nobody watching. Seeing a pattern and getting the right person to act on
it are different jobs. [Hugging Face, July 27,
2026](https://huggingface.co/blog/agent-intrusion-technical-timeline).

The console built in the last chapter has a smaller, immediate problem. It
shows one conversation while the root can delegate to a whole tree. A child
thinking about a problem, one executing a tool, and one waiting on its
children can all report that they are running. The operator cannot see the
shape of the fleet or find the conversation that needs attention without
reconstructing it from tool results. That operator is a traffic cop directing
cars through a four-way stop with a flashlight and a clipboard.

This chapter replaces the flashlight with a radar screen. The screen reports
work and offers controls; it is not evidence that the reported attack would
have been prevented.

---

## TL;DR

A fleet doing real work is hard to supervise through one chat window.
“Running” does not say whether an agent is thinking, executing a tool or
waiting for children. Add activity, a durable registry of named peers,
management tools and a browser tree whose nodes open the actual conversations.

Build it in seven stages:

1. **Activity tracking.** Record four phases: idle, thinking, tool call and
   waiting. Emit thinking when the model request starts, tool call at dispatch
   with the tool name and a short detail, waiting when a fleet operation blocks,
   and idle when the turn ends. Deliver transitions through the existing
   Observer. Lifecycle and activity remain separate facts.

2. **Named peers.** Extend the existing Ensemble with a persistent registry.
   Its public startup path loads `ensemble.json` and restores each named Agent
   with its own conversation and data directory. Creation takes a name, model
   and defining skill; deletion refuses active children, shuts down the peer,
   removes its registry entry and cleans up its directory. Named peers survive
   application restart. Spawned task children are not added to the registry.

3. **Ownership enforcement.** Each management tool resolves a target through
   the calling Agent's fleet. A foreign target and a nonexistent target return
   the same “unknown agent” error. Preserve Chapter 20's enforcement and apply
   it to every new target-bearing operation, including durable-result fallback.
   The trusted browser's tree selection is a different interface from a
   model-callable management tool.

4. **Eight management tools.** Add `wait_for_agent_change`, `send_message`,
   `wait_for_agent`, `ask_agent`, `get_submission`, `kill_agent`,
   `get_sub_agent_status` and `interrupt_agent`. Change waits observe only their
   selected children: state, exit, messages, matching output and lack of
   progress. Cursor reads return new finalized output. Messages steer a busy
   child or start an idle turn; questions queue a separate turn and await its
   result. Preserve Chapter 20's default text or custom JSON object through status,
   joins and `get_submission`; rejected data stays incomplete. A submission has a
   durable `submit.md` fallback under that child's ownership. An old stored result
   does not mark a respawned task submitted. Status includes
   activity and measured usage. Graceful shutdown waits for current/queued work;
   kill cancels the child's lifetime. Both interrupt modes end the turn while
   preserving the Agent and independent jobs. Accepted submission can precede the
   submit tool's result and reliable completion/save: disposal waits for completion
   and drain, not just the submitted flag.

5. **One actual startup path.** Both CLI and optional GUI use the public
   Ensemble to install the recursive factory, human sink and per-Agent paths
   before restoring agents. The accepted API calls `ensemble.New` then
   `root.Start`; an existing non-fleet consumer can still use `NewAgent`.
   Verify the executable's actual behavior, as well as its imports and owners.

6. **A selectable browser tree.** Send an initial tree and continuing
   activity/state changes for roots, named peers and nested children. Show
   local name, parent, persistence, lifecycle and activity. Clicking a node
   selects that Agent's retained conversation and subsequent live output for
   that connection. Repeated local names in different branches must remain
   distinct. Replay completed work as completed, while genuinely pending work
   remains pending. Nested submitted JSON must survive actual tool-result display,
   reselection and reconnect. Restart restores retrieved results in the owner's
   history; it does not recreate ephemeral children. A slow connection must not block
   the fleet or shutdown.

7. **Visible creation and deletion.** Provide name, model and skill inputs for
   named peers and deletion controls for those peers. Display rejection when
   children remain active. Preserve bottom-up `submit`, human message routing
   and the three bounded idle-correction wakes for ordinary delegated Agents.
   Main and named peers have no Agent parent and finish normally in chat.

The accepted implementation emits `agent_tree`, `agent_activity`,
`agent_state`, `agent_created`, `agent_deleted` and `agent_selected`. Its
complete-tree updates carry current state as well as its incremental activity
messages. Equivalent field spellings and update strategies are acceptable;
conversation selection and the visible controls are required. No React
migration or persistent task spawning is part of this exercise.

Typed retrieval keeps identity and data separate. The accepted public API is
`Fleet.SubmissionData(id) (Submission, error)`, where `Result` holds default text
and `Payload` holds custom JSON. `Submission(id)` retains text compatibility and
returns an error for custom data. The tool returns `agent_id` with either `result`
or nested `payload`. `ask_agent` returns compact JSON text for a custom-only answer
instead of a blank successful string. All returned raw JSON belongs to the caller.
Read legacy and versioned rows as defined in Chapter 20, validate stored custom
data under its stored contract, and refuse malformed or unknown-version rows.
Durable fallback first resolves an owned child; it is not a global archive lookup.

Build and grade from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 21 solutions/ch21
```

The default builds the core and its separate `gui/` command; `-gui DIR`
selects another GUI module. Section 21.9 describes the seven checks and their
limits.

## 21.1 Why "running" is not enough

Chapter 20 built the fleet. Every child agent runs its own actor, its own tool
registry, its own save file. Each Agent’s fleet manager tracks children, the
factory builds them, and idle correction catches the ones that stop too early.
The mechanism works. The visibility does not.

The fleet reports lifecycle states such as running, idle, submitted and
failed. Running alone covers work with very different demands on the operator.
An agent calling `read_file` on a config file and an agent blocked inside
`join_agents` waiting for twelve children to finish a two-hour research task
both report Running. A human looking at the fleet sees a list of green dots.
Every dot means the same thing, and the thing it means is "something is
happening, somewhere."

This is not a cosmetic problem. A list of green dots cannot tell an operator
where to look next. A thinking agent has an outstanding model request. A
tool-executing agent may be changing the environment. A waiting agent is
blocked on something, possibly children or jobs that are still working. An
idle one may need a nudge. Waiting is not a security verdict: its descendants
can still act, and earlier effects do not disappear when its phase changes.

An agent executing a destructive command and one waiting for its children to
submit should not offer the operator the same description. The phase helps
choose the next question; the conversation and actual tool result help answer
it.

Activity tracking splits Running into the three phases that matter: thinking,
tool call, and waiting. Each transition is an observation, delivered through
the same Observer that already carries text deltas and tool results. The
agent's activity changes when the LLM request starts (thinking), when a tool
dispatches (tool call, with the tool name), when a fleet tool blocks
(waiting), and when the turn ends (idle). Four phases, four emission points,
each in a function that already exists.

The implementation crosses a package boundary. The actor in `internal/agent`
knows when a request starts and a call dispatches. The management handlers in
`internal/tools` reach their Agent and its fleet through the ordinary tool
context. The dispatch point recognizes blocking fleet operations and emits
waiting through the same Observer. There is no need to make the browser infer
activity from fragments of printed text.

The distinction became visible in the actual browser exercise. A parent called
`get_sub_agent_status` and `wait_for_agent_change` while its child implemented
and tested a function. The capture contains all four activity phases and both
conversations. The operator can select the child doing the work instead of
interrupting its parent merely to ask what happened. Section 21.8 follows that
run through completion and restart.

The detail string on a tool call carries the tool name and enough context to
distinguish one call from another. `"read_file: main.go"` is enough.
`"read_file: main.go lines 1-50 with encoding utf-8 and tab width 4"` is a
status bar pretending to be a log line. Answer the question a human actually
asks: what tool, what file. The accepted implementation keeps the name
separately and truncates the argument detail to fifty Unicode characters; the
full call remains in the conversation.

## 21.2 The Ensemble

The first edition's fleet chapter exposed five composition failures: the
factory not registered, the user sink not connected, children not receiving
the same construction path. Library tests had supplied wires that the actual
command forgot. Chapter 20 preserved that lesson. The Ensemble already owns
construction here; named persistence extends that owner instead of moving
services into a new callback bundle.

The registry is a JSON array of entries: name, model, skill and creation
timestamp. The public `Start` path reads it, installs the recursive factory
and user sink, constructs the root and restores the named peers. Each Agent
loads its own history through the existing persistence code. The registry
identifies an Agent; it does not contain that Agent's entire conversation.

The persistence model is deliberate. Named agents in the Ensemble registry are
long-lived peers: “researcher”, “reviewer”, “coder”. They survive binary
restarts. Spawned sub-agents within a single agent's fleet are ephemeral: they
live within a task and are not automatically recreated on restart. The
Ensemble owns the named agents. Each Agent's fleet manager owns its children.
Two ownership relationships, two lifetimes, one Ensemble that manages both.

Named peers have no Agent parent, just as the main Agent has none. They finish
normally in chat. Ordinary delegated children owe a submission and receive the
three corrective wakes if a successful turn stops without one. Being displayed
under the same Ensemble node does not make those completion policies identical.

In the supplied command, the application's working directory contains
`ensemble.json`. The root uses `cr/` for its Agent data; its default
`save.json` is in that working directory, or at the explicit `--save` path.
Named peers use `agents/NAME/`, with their own save and event log. Their task
files still use the application's workspace. Separate conversation storage
does not create private copies of source files.

`CreateAgent(name, model, skill)` builds and saves the peer, then persists its
registry identity before announcing success. Empty model or skill values
inherit the application's defaults. `DeleteAgent(name)` refuses unfinished
children and closes child admission before shutting down the peer and removing
its identity and data. Otherwise a new child could appear between the check
and deletion. Application `Close` is different: it saves and stops the peers
while preserving their registry entries for the next launch.

Central construction gives the developer one place to inspect these rules. It
does not make a forgotten method call a compile error. Build the real binary,
create a named peer, exchange a message, close it and restart from the same
directory. A constructor that compiles but never saves the registry will fail
at the point the reader actually cares about: the peer is gone.

## 21.3 Containment boundaries

The fleet tools from Chapter 20 already enforce ownership. Every tool resolves
its target through the caller's fleet manager, which holds only the agents the
caller spawned. Agent A cannot shut down agent B's child, because agent A's
manager does not contain it. The tool returns "unknown agent", the same error
as if the target did not exist.

This is a containment boundary, not an access-control list. There is no global
agent map that the tools filter. The map does not exist for the caller. A
poisoned agent that tries to enumerate the fleet discovers only its own
children. An agent that tries to guess an id and shut it down gets "unknown
agent" whether the guess is right or wrong, because the distinction between
"does not exist" and "exists but you cannot reach it" is the kind of
information leak that lets an attacker map the fleet before attacking it.

The boundary is the lookup scope. The tool context leads to its owning Agent,
and that Agent's manager searches its direct children. A trusted GUI client
can walk the Ensemble to select a conversation, but model tools do not use
that global view as a fallback. The durable submission reader must resolve the
owned child before reading its file, too.

This protects those management operations. It is not process isolation or a
promise against every route through shared files, permitted tools or
compromised native code. Chapter 19's restrictions still matter. The manager's
small world makes the authority easy to inspect and test.

A cross-tree rejection test verifies this by construction. Create two
independent agents, each with a child. Have agent A attempt to shut down agent
B's child by name. The result is "unknown agent." Swap the names, try again,
same result. Pair that rejection with a successful operation on A's own child.
Otherwise a manager that rejects every request would look equally contained.

## 21.4 The management console

Eight tools complete the fleet API. The seven from Chapter 20 let an agent run
a fleet: spawn, join, list, shut down, message, respond, and submit. The eight
tools added here let an operator or a parent supervise a fleet in real time.
They exist because a model that can spawn agents but not observe them has
delegated a task and then closed its eyes.

`wait_for_agent_change` is the core of the supervision loop. It takes a list
of agent IDs and blocks until something interesting happens to any of them. A
state change, an agent exit, a message to the parent, a pattern match in the
output, or a stuck timeout where a processing agent stops making progress. The
accepted manager takes a snapshot, waits on its existing change notification
or a short timer, and compares the requested children with that snapshot. The
timer also lets it notice a lack of progress when no notification arrives.

The snapshot-diff approach is a deliberate choice. The waiter needs the
current state of a small set of children, not a new durable pub/sub system. It
wakes, reads those children and reports what changed. The work grows with the
selected set; the chapter's small fleet exercises are not a benchmark for a
thousand agents. The wait is an ordinary tool job, so the existing callback
delay can return a running handle without cancelling the observation.

One early version got the selection wrong. A question from an unselected child
incremented a manager-wide message counter. A wait watching another child then
reported that watched child as the sender. Both child requests were otherwise
parked. Tracking message progress per sender fixed the attribution; the
selected-question control still woke, and the unselected question no longer
impersonated progress elsewhere.

Notifications are a reason to inspect state, not a reason to announce that any
particular child changed. `timeout_seconds` bounds the observation job;
`stuck_seconds` describes a busy child's lack of progress. Neither is the
ordinary tool callback delay that hands control back to the model.

`send_message` delivers a prompt to a running agent. If the agent is mid-turn,
the prompt arrives as a hint between tool calls. If the agent is idle, the
prompt starts a new turn. The result includes recent history and a cursor, so
the caller can resume from where it left off without re-reading the entire
conversation. A cursor belongs to that child, not to the whole Ensemble.

`wait_for_agent` complements `send_message`. Given a cursor from a previous
`send_message` or `wait_for_agent` call, it blocks until the agent produces
new output past that cursor, then returns the new content, or returns when the
child is idle. The accepted cursor advances through recorded history and
returns finalized assistant text and tool output. Streaming fragments stay in
the GUI display path; returning each fragment here and its full replacement
later would count the same words twice.

`ask_agent` is the synchronous composite: send a prompt, wait for the turn to
complete, return the result. It exists because the common case is
ask-and-wait. Unlike live steering, this queues a distinct question with its
own completion. Observing new output from `wait_for_agent` alone is not the
same as receiving that question’s result.

`get_submission` retrieves the result a child accepted through `submit`.
Chapter 20's default result is text; a custom contract returns the complete
direct JSON object. The public `Fleet.SubmissionData(id)` returns a typed
`Submission`: `Result` holds default text, while `Payload` holds custom JSON.
The older `Submission(id)` string API refuses a custom result and directs its
caller to the typed method. Silently converting every object to a string would
move parsing and error detection into every consumer.

A successful submission is appended to `submit.md` in the child's data directory.
Legacy JSONL rows remain `{"result":"..."}`. Custom rows have `version: 1`,
`submit_schema` and `payload`, so each historical result retains the contract
that accepted it even if a later task uses another schema. Unknown versions and
malformed rows return errors. Newlines inside a result remain JSON string data,
rather than ambiguous separators between submissions.

The tool tries memory first and falls back to disk, because an operator who lost
their context window should not also lose their fleet's results. It returns
`agent_id` with either `result` or `payload`; the custom payload remains a nested
JSON object. Status and joins preserve it too. Lookup still requires an owned
child, and retrieving a historical row does not mark its current task complete.
The file survives shutdown without becoming a global archive API or automatically
resurrecting an ephemeral child.

The accepted submission can be visible before the submit tool's result and the
actor's final completion/save have finished. A join answers whether work was
submitted; disposing of the child also requires its reliable completion and drain.
Those boundaries are deliberately separate from how the JSON result is displayed.

`get_sub_agent_status` reports detailed status including the activity phase
from the first stage. The operator sees thinking, tool call (with tool name
and detail), waiting, or idle. Token usage is included because an agent can
consume substantial input without making useful progress. Compare those counts
with the task and actual output before deciding to intervene. A shorter result
after more tokens is a reason to investigate, not proof by itself that the
child was looping.

`kill_agent` terminates an unresponsive agent. It cancels the in-flight
request, interrupts the current turn, shuts down the agent, and removes it
from the fleet. The difference from `shutdown_agent` is force: shutdown waits
for the child's current and queued turns; kill cancels its lifetime, jobs and
descendants. Application teardown also needs the cancelling path so an
unresponsive child cannot hold the whole process open.

`interrupt_agent` ends the current turn without shutting down the agent. Two
modes: `after_tool` requests a stop at the next tool-report boundary, while
`now` cancels the current turn immediately. The Agent stays alive and can
start a new turn. Independent jobs retain their real results. A reply saying
“interrupt delivered” proves receipt; the useful check is that work stops at
the intended boundary and the same Agent can answer a later question.

## 21.5 The tree in the browser

A fleet is a tree. The root agent sits at the top. Named agents from the
Ensemble registry are its peers. Spawned sub-agents are children, and those
children can have children of their own. The GUI introduced in Chapter 7 shows
a single conversation in a single pane. The operator who needs to see the
fleet shape must read tool results and reconstruct the topology in their head.

The agent tree panel replaces that mental reconstruction. On WebSocket
connect, the server sends an `agent_tree` message containing the complete
tree: every agent's id, parent, state, and current activity. As activities
change during a session, the tree updates in real time. A thinking agent shows
a different indicator than a tool-executing one. An idle agent is visually
distinct from a waiting one. The four activity phases accompany lifecycle
state; a submitted Agent can be idle without losing the fact that it
submitted.

The tree is assembled by walking the actual ownership hierarchy. A synthetic
Ensemble node contains the root and named peers; each Agent contributes its
owned children recursively. A local display name is insufficient identity:
`reviewer/helper` and `coder/helper` are different nodes. The accepted tree
carries both the local `id` and an ownership-path `key`, along with parent,
persistence, lifecycle and activity.

The existing root Observer receives events from each Agent. The GUI keeps
conversation events, partials and replay cursors in a view associated with
that Agent. It sends conversational output only to connections selecting that
view; a second tab can keep reading another Agent. Tree changes still reach
both. On selection, the client resets the Artifact scroll and receives
retained history before continuing with live output. Reusing one global
partial map would make identical response-local IDs collide across
conversations.

Structured submissions use these same conversation surfaces. Render their tool
arguments and results as text safely, retaining nested JSON data in the underlying
result. Select a registered child during work, then reconnect or reselect it in
the same running application. After an application restart, inspect the retrieved
result retained in the restored owner's conversation; ephemeral task children
are not automatically recreated. Compare the actual object, including arrays,
booleans and nulls; finding one marker somewhere in a frame cannot establish that
the payload survived. No separate result browser is needed.

A tree traversal can take several managers' locks. It therefore runs in a
small coalescing worker outside the synchronous actor Observer; an activity
notification asks for a refresh rather than walking the whole hierarchy while
the actor waits. This uses the existing Observer and GUI lifetime instead of
putting browser concerns into the core actor.

Selection exposed another lifetime mistake. Repeated clicks from a browser
that stopped reading filled the replay queue. The receiver waited to enqueue
history while holding the server mutex; after disconnect, shutdown needed the
same mutex. A slow tab had stopped the server.

The correction admits an entire replay as one queue batch. If the connection
cannot accept that batch, it disconnects and can reconnect for a fresh replay.
The selected view is published only after admission succeeds. Increasing the
buffer would only postpone the blocked send; dropping the tail would quietly
lose retained conversation. Other Agents and browser connections keep their
own lifetimes.

Creating and deleting named agents from the GUI uses two WebSocket messages:
`create_agent` and `delete_agent`. The browser sends a message with a name,
model and defining skill for creation, or a name for deletion; the server
calls `CreateAgent` or `DeleteAgent`, and the server broadcasts the updated
tree. `DeleteAgent` rejects when the agent has active sub-agents, for the same
reason `submit` rejects: tearing down an agent while its children are running
leaves orphaned work that no one is waiting for. The operator sees the
rejection as an error in the GUI and knows to shut down the children first.

### Optional: a display name without a new identity

**TL;DR:** The current GUI's **Rename** button changes an Agent's visible label,
including main's, while keeping its routing and saved conversation intact.

An operator may want a more useful name after creating an Agent. Renaming its
registry identity would also change what messages address and where history
lives. A display name needs neither change. Every actual Agent node, including
main and task children, offers Rename; the synthetic Ensemble grouping does
not. The browser prompts for a label and sends it through Chapter 8's existing
settings path. For example, this labels main without switching conversations:

```json
{"type":"update_settings","key":"main","settings":{"display_name":"Coordinator"}}
```

The server trims surrounding whitespace and rejects a blank name. Cancelling
the prompt changes nothing. The optional `key` targets the Agent; without it,
the patch applies to the selected conversation. Save `display_name` with that
Agent's existing GUI preferences, including the main Agent's `gui` section in
file-configured launches. No second name registry is needed.

Tree labels and the selected conversation title update across tabs and survive
reload. Main and named peers retain their labels after application restart;
a saved label does not make a task child persistent. Without a display name,
show the original ID. Render labels with `textContent`, so a name containing
markup remains text. Selection keys, message addresses, registry IDs, deletion
targets and history paths all keep their original values.

For an optional browser check, keep main and a peer selected in separate tabs,
rename each, and confirm that renaming leaves both selections unchanged.
Then reload and restart, reselect each Agent, and verify its label and retained
words. This addition does not change the historical exercise or its grader.

## 21.6 The composition root, condensed

The first edition condensed a large command into `NewEnsemble` and `Start`
after its fleet wiring failures. Here the public Ensemble already exists.
Startup gains registry restoration while retaining the same recursive factory
and human route. The CLI and optional GUI both call that path; offline
consumers can retain their smaller non-fleet protocol.

What belongs in the executable remains there: command-line flags,
authentication configuration, terminal input and browser serving. What belongs
to each Agent stays there: conversation, configuration, tool visibility and
jobs. The root supplies construction and application-level lifetime. A tool
reaches those services through its owners rather than receiving another bag of
sibling callbacks.

The proof has to cross the executable boundary. A test that installs its own
factory can conceal an unwired command, just as it did in Chapter 20. Start
the built optional GUI from a scratch directory, create a named peer, let that
peer spawn a child and select both conversations. Then restart without
inference and verify the peer's actual saved conversation. A registry row
proves identity; the restored words prove more of what the operator meant by
“save.”

## 21.7 The exercise

Extend the previous chapter with the seven stages in the TL;DR. Keep the
actual owner path and optional GUI. The operator must be able to create a
named peer with its chosen model and skill, select that conversation, inspect
nested task children, and delete a peer only after its children are resolved.
Test both completed and genuinely busy views, then restart and check retained
words as well as registry identity.

For a small browser task, build the supplied command from the repository root:

```sh
CH23_CORE="$PWD/solutions/ch21"
(cd "$CH23_CORE/gui" && go build -o /tmp/ensemble-ch23-gui ./cmd)
CH23_WORK=$(mktemp -d /tmp/ensemble-ch23.XXXXXX)
mkdir -p "$CH23_WORK/skills/ensemble"
cat > "$CH23_WORK/skills/ensemble/SKILL.md" <<'SKILL'
---
name: ensemble
description: A coding fleet that monitors children and verifies their work.
type: primary
tools: read_file write_file edit_file run_command wait_for_job tool_limits spawn_sub_agent join_agents agents_status shutdown_agent send_message_to_parent respond_to_parent_message submit wait_for_agent_change send_message wait_for_agent ask_agent get_submission kill_agent get_sub_agent_status interrupt_agent
---
Delegate only when asked. Keep task children within their assigned work.
Monitor their actual results, answer their questions, and verify their files.
Delegated agents finish with submit after all their children are resolved.
Main and named peers finish with a normal chat answer. A running job notice
is not its completed result; collect it with wait_for_job.
Tools: $TOOLS
SKILL
```

Use the dedicated authenticated Ensemble profile from Chapter 16, with access
to `gpt-6-astra`. The following launch selects high reasoning and subscription
billing explicitly. It does not borrow another application's credentials.
Choose an available loopback port rather than attaching to a server already
running on the default port:

```sh
CH23_PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')
cd "$CH23_WORK"
printf 'Open http://127.0.0.1:%s/\n' "$CH23_PORT"
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
OPENAI_BILLING_MODE=chatgpt OPENAI_REASONING_EFFORT=high \
OPENAI_REASONING_SUMMARY=detailed \
EN_PRIMARY_SKILL=ensemble EN_SKILLS_DIR="$CH23_WORK/skills" \
/tmp/ensemble-ch23-gui --port "$CH23_PORT"
```

Keep the terminal open while using the browser. If binding fails, choose a new
port; another listener is not evidence that this process started. In the
Agents form enter name `console`, model `gpt-6-astra`, skill `ensemble`, then
Create. Click `console` and send:

```text
Create exactly one child named helper. Have it implement slugify.py with
slugify(text): lowercase the text and replace runs of whitespace with one
hyphen, trimming leading and trailing whitespace. Have it write test_slugify.py
with unittest cases for empty input, mixed case, and repeated whitespace,
run those tests, and submit its result. Monitor helper with get_sub_agent_status
and wait_for_agent_change, join it, and retrieve get_submission. Independently
run python3 -m unittest -v test_slugify.py. Write fleet-result.txt with the test
outcome, then answer in chat with CH23_FLEET_COMPLETE and a concise explanation. Do not use
web tools or create another child.
```

Select `console/helper` while it works, then return to `console`. Open a
second tab on `main` and confirm that selecting the child in the first tab
does not move the second tab's conversation. The tree should identify
thinking, tool work and waiting as they occur; a very short phase need not
remain visible long enough for a screenshot.

Create an unused named peer `spare` with the same model and skill, then delete
it. Once the task has finished and the terminal is idle, use EOF to close the
application normally. Run the same launch again from `CH23_WORK`, select
`console` and inspect its retained answer and retrieved child submission. `spare` stays absent, and the
ephemeral `helper` is not restarted. Merely selecting the restored
conversation must not ask the model to reconstruct it.

These are the accepted public browser messages behind those controls; a client
sends each object separately over `/ws`:

```json
{"type":"create_agent","name":"console","model":"gpt-6-astra","skill":"ensemble"}
{"type":"select_agent","key":"console/helper"}
{"type":"delete_agent","name":"spare"}
```

Use separate bounded controls for rejection while a child is active, foreign
and absent child IDs, selected versus unselected child questions, cursor
output, both interrupt modes and a slow browser disconnect. A successful small
coding task cannot exercise all of those transitions on demand. Keep the
positive controls: an owned child must remain manageable, and a truly pending
request must remain visibly busy.

## 21.8 Taking the console for a spin

The recorded live exercise used the real browser, GPT-6 Astra at high
reasoning and subscription billing under the earlier peer-submit policy.
A named peer `console` created one child,
`helper`. The child wrote `slugify.py` and three unittest cases, ran them
successfully and submitted. The parent monitored it, joined, retrieved its
submission, reran all three tests independently and wrote `fleet-result.txt`.
Both shell runs returned exit zero and `OK`; the final submission was
`CH26_LIVE_COMPLETE`, using the original chapter's number.

The run contains fifteen model requests: ten from `console` and five from
`helper`. The browser selected the child and returned to its parent, captured
all four activity phases, and created and deleted an unused named peer. After
a normal shutdown, restart restored `console` and its submitted conversation
without another inference request. It did not revive `helper`. The new
reproduction setup above gives the primary skill explicitly; it is not a claim
that every byte of that earlier launch was identical.

The current setup also asks the named peer to finish in chat, while its delegated
child still submits. The historical peer's submission remains part of that run;
it does not demonstrate the current completion policy.

The work was correct. One status label was not.

The live screenshot showed idle activity alongside `input_pending`, even after
selecting the saved conversation on restart. The runtime had ended the turn,
but the retained history still said that tool results awaited another request.
Ordinarily that next request consumes the pending input. A successful fleet
`submit` ends directly, so no next request arrived to hide the mistake.

The correction records the existing `turn_ended` fact after the submitted tool
batch drains, before completion and save. The GUI then replays the truth
already owned by history. Painting every selected Agent idle would make the
screenshot look right while concealing real work. A separate parked-request
control rejects that shortcut: completed and restarted conversations show
idle; a genuinely outstanding request still shows `in_flight`.

The paid capture belongs to the earlier source and retains its bad label. The
corrected behavior was checked through the real runtime, WebSocket selection
and restart with a local provider peer. No paid repeat or migration of old
incorrect saves is implied. The useful outcome is both parts of the story: the
fleet did the coding task, and looking at its actual console found a state
error that the successful task had not tested.

## 21.9 What the grader checks

The revised grader retains the original seven weights, totaling 100 points. It
supplies an explicit supported model and builds the separate optional GUI. The
original grader's actual 0/100 run stopped at those launch boundaries: six CLI
scenarios omitted the model, and the browser probe started the headless
command with GUI flags. That run established neither seven runtime defects nor
seven working behaviors.

The revised driver uses external provider peers and the real applications. Its
command is:

```sh
go run ./cmd/grade-edition2 -ch 21 solutions/ch21
```

| Behavior | Revised observation | Points |
| --- | --- | --- |
| Activity | Correlated detailed status and ordered activity; rejected custom submit remains pending before a valid correction | 15 |
| Named persistence | Real create with model/skill, restored conversation and typed retrieved result, explicit child respawn with durable fallback, and deletion | 15 |
| Ownership | Owned targets work; foreign/missing targets, including typed retrieval, agree; historical fallback leaves the new task incomplete | 15 |
| Change waiting | Selected progress wakes the wait; command, valid submission, typed status and join results complete | 15 |
| Composition | The actual CLI constructs recursive agents, completes work and exits normally on EOF | 15 |
| GUI | Nested tree and selection, truthful completed/pending replay, exact nested tool-result JSON live and after replay/restart | 15 |
| Previous fleet behavior | Active children prevent submit; delegated idle turns receive three wakes; main chat completes without correction | 10 |

The structured phases compare complete arrays, booleans and nulls in actual
correlated results. Live and reselected views exercise a registered child;
restart checks the retained result in its owner's conversation. A separately
respawned child can retrieve its previous durable result while its new task stays
running and unsubmitted. No phase assumes automatic child resurrection.

Actual parent relationships and executable construction still need source
review. A successful command cannot prove them just by printing a response.
The retained browser and focused runtime checks cover stronger selection
isolation, finalized cursors, interruption, selected/unselected messages and
slow-client lifetime. A tree frame with one ID alone would miss those
obligations.

Targeted defects check whether an assertion notices its chosen failure. They
do not prove that seven rows are generally independent, or that every behavior
within a row is covered. Restore the implementation and keep successful
controls beside refusals and timeouts. In particular, a console that always
says “idle” and a manager that always says “unknown agent” should fail for the
same practical reason: neither tells the operator what is actually happening.
