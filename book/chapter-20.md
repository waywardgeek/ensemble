# Chapter 20: The Crossover

The morning of the switch, I had two agents open on the screen.
CodeRhapsody, the one that wrote this book. Ensemble, the one the
book describes. Same hardware, same model, same prompt. One of them
built by reading eighteen chapters of instructions. The other built
by writing them.

If the agent that came out of this book cannot replace the one that
went into it, then eighteen chapters taught a toy. If it can, the
loop closes and the last chapter writes itself. Close it now, while
every design decision is still fresh enough to debug. A month from
now the memory is gone and the crossover becomes a project instead
of a morning.

The crossover took twelve days. Not because anything was
architecturally wrong, but because the distance between passing a
grader and surviving a workday turned out to be thirty-seven
changes, a vendor caching bug, and this chapter.

---

## TL;DR

Thirty-seven changes were needed before the agent built from this
book's designs became a daily driver. They sort into four kinds:

1. **Loud failures** that announce themselves within minutes of real
   use: part-ID collisions, a settings dialog that never saved, a
   Ctrl-C that killed the process without persisting the
   conversation.

2. **Silent failures** where every request succeeds but something
   underneath is quietly wrong: a model selector that selected
   nothing, a credential provider that billed the wrong account
   after a vendor switch, a startup resolver that sent an OpenAI
   model to Anthropic.

3. **Two-path failures** where the live streaming path and the
   replay path disagree on what content looks like: text vanishing
   after a restart, resets that cleared the log but not the screen,
   a client that classified messages by guessing at turn state.

4. **Invisible failures** where the output is correct and the bill
   is not: a cache diagnostic tool with a 100% false-positive rate,
   a vendor route with an undocumented 22,000-token write floor,
   and a plan-route caching gap that ended in a bug report rather
   than a code fix.

Every grader scored 100/100 before the crossover began. None of
these bugs was caught by a grader. The only instrument that finds
them is daily use, and the only time to start is now.

---

## 20.1 What Breaks on Day One

The first morning, the operator opened Ensemble and typed a
prompt. The second assistant response rendered inside the first
one's bubble. These first bugs are gifts: they announce themselves
the moment someone uses the tool for longer than one turn, and
they teach nothing the preceding chapters did not already know. They exist
here to establish a baseline: the agent works, if the definition
of "works" is narrow enough. The next section raises the bar.

**Part IDs collided across turns.** Every vendor parser minted IDs
from a counter scoped to one API response. The first text block of
every turn was `part_id: 1`. The GUI keyed its artifact map by that
ID and never cleared it, so turn two's text deltas found turn one's
DOM node and appended inside it. Every assistant response rendered
inside the first one's bubble.

The rejected fix was clearing the map on each turn boundary. It
works, but it makes the client responsible for a property the server
should guarantee, and it needs a reliable "new turn" signal before
the first delta arrives. The real fix is a single monotonic counter
on the actor, shared across all parsers and both the delta and final
emitters, so a content address is unique for the lifetime of the
session.

**The settings dialog never saved anything.** Every control called
`sendPatch` on every change event. `sendPatch` was never defined.
A `ReferenceError` fired on every interaction, silently, and the
dialog looked correct because theme and font size both apply locally
on the way past: the visible half worked while the durable half
never ran. Four wrong hypotheses came first. The harness log named
the real cause on its first line.

That pattern is worth isolating. A control that shows its effect
immediately can hide a broken write indefinitely, because what the
operator sees is the effect, not the persistence. The three
settings controls that accepted input while doing nothing at all
(temperature, max tokens, system prompt) were removed outright
rather than wired up. A control that does nothing is worse than no
control: it invites someone to set it, confirm it visually, and
trust an outcome that was never produced.

**Ctrl-C did not save the conversation.** There was no signal
handling. The main goroutine blocked on `for in.Scan()` over stdin,
and the save ran after that loop returned. SIGINT killed the process
where it stood, so the save was unreachable by construction. Not a
race, simply never run. The rejected fix was a signal handler that
saves and calls `os.Exit(0)`, because exit skips deferred functions
and the defers here are load-bearing. The real fix moved stdin to
its own goroutine; the main goroutine now selects on stdin
completion versus a signal. Either way it falls out through the
existing exit path, and every defer still runs.

**Font size changed nothing.** Every `font-size` declaration in the
stylesheet used absolute pixels, so setting the root font size did
nothing downstream. All twenty-one declarations converted to `rem`.
Additionally, `if (app.font_size)` skipped zero, which is exactly
what a never-set value deserializes to. Zero now means "use the
default."

**The chat input was a single-line field.** Long messages scrolled
sideways through a one-line slot. It is now a `<textarea>` that
grows with its content up to a `rem`-based cap, then scrolls. The
cap lives in CSS rather than JavaScript, because a constant in
pixels silently breaks the font-size slider: at 32px text the box
stops at the same physical height and holds half as many lines.

None of these individual fixes mattered much. The pattern does.
Every one of them would have been caught by ten minutes of real use.
No grader checks whether a tab is identifiable among twenty open
tabs, whether a textarea grows, whether a process survives a signal.
These are usability gaps, and they are the cheapest category of
crossover bug because they announce themselves immediately. The
expensive category is next.

---

## 20.2 The Bugs That Succeed

The second kind of crossover bug is the one where every request is
answered, the GUI looks right, the tests pass, and the only thing
that disagrees is the invoice, or the model, or the credential,
thirty days later. These share a shape: the visible half of an
operation works while the functional half never runs.

**The model selector selected nothing.** The GUI offered a dropdown.
Choosing a different model changed nothing at all. Every request
went to the startup default. The dropdown wrote to one config field;
startup read from another. Nothing carried one to the other. It was
a known-dead setting.

What makes it worth a section rather than a bullet is the readout.
`effectiveModel` consulted the settings store first, with a comment
explaining that the GUI selection is what the operator just did. So
the display showed the selected model while the engine dialed the
old one, and the display was the thing that was wrong.

A readout fed from intent always agrees with the operator. Only a
readout fed from the fact can disagree, and disagreeing is the
entire job.

**The credential did not cross a vendor switch.** An operator who
switched models from Claude to GPT in the dropdown saw the dialect
change, the host change, and the response arrive correctly. The
plan credential never arrived. ChatGPT plan credentials were
attached once, at startup, behind a guard that checked the vendor. The guard was correct when vendor was a process
lifetime constant. Since chapter 11 the model has been changeable at
runtime, which means the vendor has been changeable at runtime,
which means the guard had an expiry date on it.

The runtime switch had already been fixed once before. It moved
model, vendor, surface, base URL, and API key together, carefully,
with a refusal path. Five fields. The credential provider was not
among them, because when that fix was written the provider did not
exist yet.

Flip the picker from Claude to a GPT model and the dialect changes,
the host changes, the key is refilled from the endpoint table, and
the plan credential never arrives. The request is rendered
correctly, sent correctly, answered correctly, and billed to the
metered API key. Nothing breaks. There is no error, no retry, no
degraded reply. The only component that disagrees is the invoice, a
month later.

The fix was one line: adding the credential to what the switch
already moves. But the more important change was making the state
observable. The engine now logs the credential in force whenever it
changes. `DescribeCredential` names the kind and has no path that
can emit a token, because putting the redaction at each call site
would mean a rule enforced in five places and broken in the sixth.
A missing provider reports as the static key, not as "none," because
falling back to a billed key while expecting a plan credential is
the commonest mistake, and it must not read like the correct case.

**The vendor was not a property of the model.** Startup resolved the
vendor first and then chose a model. A model from `settings.json`
was sent to whichever vendor happened to be the default. Running
`LLM_MODEL=gpt-6-astra` without setting `LLM_VENDOR=openai`
produced a healthy-looking session running an entirely different
model than requested.

The GUI masked it. After connecting, the browser sends the model
choice over the WebSocket, correcting the mistake moments after
startup. The CLI had no such correction and ran the wrong model the
entire time, silently. The bug was invisible in the interface where
it was survivable and total in the interface where it was not.

The model is now settled first. The vendor is taken from the model
table. An explicit vendor override still outranks the table, for
local OpenAI-compatible endpoints serving identifiers the table has
never seen.

**The pattern across the section.** Every bug here produced
correct-looking output. The selector showed the right model and
dialed the wrong one. The settings dialog displayed the new theme
and persisted nothing. The credential system sent the right request
and billed the wrong account. Graders pass because they test output,
and the output was fine. The failures live in persistence, billing,
and identity, none of which throw errors. The fix is not more tests
but instruments: a credential log, a cost meter, a cache lens.

---

## 20.3 Two Paths, One Screen

The agent renders content through two separate paths. A live
streaming path carries real-time deltas and tool calls as they
arrive. A replay path reconstructs the same view from the saved
event log on every startup. Both share the same WebSocket protocol
and the same message types.

These two paths should produce identical screens. They did not.
Every restart during the first week disagreed.

**The hub believed its log was empty.** `Hub.logLen`, the hub's
notion of how much history is renderable, was only ever advanced
inside the live observation handler. On a restored session with a
full log but no new observations, that counter stayed at zero and
the replay loop ran `for i := 0; i < 0`. The hub held the entire
conversation and reported that none of it existed. `NewHub` now
seeds the counter from the log it is handed.

Fixing that exposed a second bug: every send used a select with a
default case, dropping silently when the buffer was full. Dropping
is right for a live frame, which a newer frame supersedes. It is
wrong for a replayed frame, which is never re-sent. A long
conversation renders to far more messages than the 256-slot buffer,
so replay now waits for room with a timeout to avoid pinning a dead
client.

**Assistant text vanished on restart.** After restarting, the GUI
showed user messages and tool calls, but every model response and
thinking block was gone. The save file had all 392 events intact.
`partToWireMsg` produced messages with no `part_id` field. The GUI
keys its artifact map by `part_id`, so every replayed response
landed under `undefined` and collapsed into one invisible div. The
live streaming path set that field through one function; the replay
path used a different, incomplete one.

This is the lesson of the section. A field that is present in one
path but absent in the other creates a class of bug that is
invisible during development, when only live streaming runs, and
total after a restart, when only replay runs. The fix is not "add
the field" but "share the identity." A `part_id` is the identity of
a piece of content, and identity must survive both paths.

**The reset nobody saw.** Clicking reset cleared the conversation in
the log. The old text stayed on screen until a browser refresh.
Refresh goes all the way back to disk and rebuilds from scratch, so
a bug that a refresh cures is a display bug, not a storage bug. The
durable record was right; the live screen was wrong.

`Engine.Record` appends to the log and applies the reducer. It does
not notify the observer. The reset event traveled through the replay
channel (which self-heals on refresh) and never through the live
channel (which is what the operator is watching). A
`ConversationCleared` observation fired from the reset handler gives
both channels the same event.

**The client that guessed.** Killing the agent mid-turn left stale
in-flight state in the save file. On the next launch, the GUI
classified every message as a hint to a nonexistent turn. Messages
vanished silently. The connection was open. The agent answered
nothing, forever.

The browser decided message type by branching on its own copy of
turn state: `if (agentState === 'idle') send prompt else send hint`.
Chapter 2 states the rule this breaks: a hint and a prompt are the
same event, distinguished only by turn state, and classification
belongs in the reducer. This client shipped a guess.

The fix is a deletion plus a move. The GUI now always sends a
prompt. The actor classifies, without consulting any state variable,
by its own position in the drain loop. Arriving at the main loop
means idle, so it is a prompt. Arriving at the mid-turn drain means
a turn is running, so it is a hint. The classification is
structural.

**The terminal as a diagnostic tool.** The CLI had called a bare
engine loop, bypassing the actor entirely. It could not reproduce
any of these bugs. Rebuilding it as a text Observer over the same
actor loop turned the CLI and GUI into two renderers of one agent.
The diagnosis is now mechanical: a bug that reproduces in both is an
agent bug; a bug that reproduces in only one is a front-end bug.
Writing a second renderer against the same seam is the cheapest test
of whether a seam is honest.

---

## 20.4 What You Cannot See Cannot Be Fixed

At a million tokens of context, the gap between a 91% cache hit rate
and a 22% hit rate is the difference between an affordable agent and
burning nearly half a weekly allowance in a morning. The correct
answer still arrives. Nothing errors. Nothing retries. The session
looks normal. The only artifact that disagrees is the bill.

Chapter 18 teaches the mechanism. This section tells the story of
what happens when that mechanism meets a production session and the
bill arrives.

**The meter comes first.** A cache breakpoint placed on an unstable
prefix buys nothing while looking like it should have worked, so
proving stability is a prerequisite, not a follow-up. The cache lens
intercepts each request body, splits it into named sections in the
provider's actual cache concatenation order, and diffs against the
prior request. Three verdicts, not one number: MATCH, PREFIX
DIVERGED (with the exact byte offset), and CACHE MISSED ANYWAY
(TTL, eviction, or the size floor). Collapsing these into a single
hit-rate figure would make the instrument unactionable.

On its first live run, the lens caught a bug in itself. Appending a
message to a JSON array shifts the closing bracket, so naive
string comparison flags every single turn as an edit. The alarm
would have fired constantly and been muted inside a day. A false
alarm that always fires is worse than no alarm: it trains the
operator to ignore the real one. The fix was per-section comparison
with grown-container detection: trim the prior section's trailing
framing and check whether what remains is a prefix of the current.

**The alarm that always fired.** A morning of work consumed 46% of a
weekly plan allowance. The journal named the cause: 175 requests
carrying 6.9 million uncached input tokens, a 22.5% hit rate. The
lens had been running the whole time and reporting a divergence on
every single request:

```
cachelens: PREFIX DIVERGED — section "input" changed
```

The OpenAI Responses API calls the conversation `"input"`. Chat
Completions calls it `"messages"`. Gemini calls it `"contents"`.
The canonicalizer knew the latter two names and defaulted to
`"messages"`, so on the Responses surface the entire conversation
was filed as a frozen section that should never change. Every
ordinary appended turn tripped the alarm.

An alarm with a 100% false-positive rate is not a noisy alarm. It is
an absent one. The operator stared at it every morning and learned
nothing. The single genuine cache break in the session, a rewrite
that cut the surviving prefix to 34%, rendered as the same line of
text as the hundred harmless appends around it. The
instrument ran, reported, and its report carried zero information.

The fix was three lines: reading the conversation key off the body.
That exposed a second gap. Within the dialogue, the lens had
recorded only that the conversation changed, which is true on every
turn. Growth and rewriting are different events. Appending leaves
earlier bytes untouched. Editing something below the top voids the
cache from the edit onward. The lens now reports rewrites separately
from growth, with the offset at which the prefix stopped matching.

**Four breakpoints, one freed by asking the right question.** The
initial layout placed a marker at the end of the previous exchange,
one message behind the newest prompt. The argument for it was that a
cached write from the prior round could be read back. Both halves
of that argument are true, and neither pays. The rolling marker from
the previous round already wrote an entry at that prefix, and both
vendors look backward past a bounded number of positions to find a
prior write. The slot was buying a second copy of something already
in hand.

The freed slot went to the tool array, at the very front of cache
order. A marker there survives an edit to the system prompt:
without it, changing a word of the prompt discards the tool
declarations too, and tool schemas are not small.

The four that remain: end of the tool array, the system prompt, the
compaction boundary, and the most recent non-ephemeral message.

**The plan route is not the documented route.** Chapter 19 scored
100/100 against a fake built from OpenAI's published documentation.
On the real ChatGPT plan route, every turn failed. The route diverges
from the metered API-key route in four undocumented ways, and each
masked the next: explicit caching returns a 400, the SSE stream
lacks a Content-Type header so the parser feeds it to the JSON
decoder, `response.completed` carries an empty output array with
items arriving only as separate events, and tool calls built from
those events had no provenance.

A fake built from documentation and code built from the same
documentation always agree with each other. That agreement proves
only self-consistency, not correctness. Only the live wire settles
a vendor question; everything else settles whether the code is
consistent with itself.

**The floor.** With the lens repaired and breakpoints placed, testing
against live sessions revealed the dominant cause. Despite identical
tools, an identical system prompt, and a stable 99% cacheable
prefix, the provider wrote nothing at 8,000 tokens. Nothing at
10,000. Nothing at 18,000. The first request to cross roughly 22,000
tokens was cached, and the request after it read back 22,144 of
them.

The documented minimum is 1,024 tokens. The measured floor on this
route is roughly 22,000. Below it, a perfectly stable prefix pays
full price every turn. Past the floor, a longer conversation is
cheaper per turn than a short one, because the long one is cached
and the short one is not. Aggressive trimming below the floor does
not save money; it forfeits the discount.

But the floor does not explain everything. A morning session burned
46% of a weekly allowance on requests between 55,000 and 73,000
tokens, well above the floor, while a first-party client using the
same OAuth token but a different endpoint sustained 61-85% hit
rates, including 12,160 cached tokens on the first request of a
fresh session. Same credentials, different route, different caching.

The bug was filed with the provider. The fix is expected within
weeks. Until then, the crossover is real and the bill is not yet
affordable. The lesson worth keeping is that an instrument never
checked against an independent measurement is a belief dressed up as
data. The lens ran all morning, reported constantly, and carried zero
signal, because it was only ever compared to its own logic.

---

## 20.5 The Agent Looks in the Mirror

An agent with a GUI and a built-in MCP server that still cannot see
its own screen is an agent that cannot improve from its own output.
The MCP server was built and documented in chapter 13:
`gui_snapshot`, `gui_click`, `gui_input`. Nothing connected to it.
The grader never noticed, because it tested the MCP client against a
stdio fake, not the path from agent to browser.

**Two doors, one per direction.** `--mcp-port` relays the GUI's MCP
server onto loopback TCP, so any external agent that speaks MCP can
connect and drive the browser. `view_gui` is a builtin tool that
asks the connected GUI for a snapshot over the existing WebSocket.
One looks outward; the other looks inward.

Verified live: prompted to list the buttons in its own interface,
the agent called `view_gui` and correctly named the hamburger,
reset, and settings buttons along with the Chats and Artifacts tabs.
It worked, and then the question was why it had not worked before.

**The underscore that hid a tool.** `view_gui` had been registered,
unit-tested, and never reached a request. The tool registry
normalizes names by stripping non-alphanumeric characters, so
`view_gui` becomes `viewgui` internally. The skill filter checked
against the raw name `view_gui` listed in skills. The question was
about `viewgui`; the skill said `view_gui`; the answer was no. Every
underscored tool name was silently dropped from every request,
regardless of what skills were loaded.

Builtins survived by accident: `NewRegistry` stored them under their
raw name while `Register` stored normalized ones. One map, two key
conventions. The fix was one line: check the declared name rather
than the normalized key. A normalized key is an index, not an
identity, and any identity check must use the name the other party
actually used.

The one-line fix broke chapter 15. The MCP connect path recorded
enabled tools under the normalized key, under a comment explaining
why. The old convention was internally consistent; the fix moved one
end of it. The cross-chapter sweep caught the regression. Bisecting
in a worktree localized the commit, and a second fix recorded the
declared name instead.

**The model that worked in silence.** Some models reason at length
internally and then act without narrating. For a supervisor who
reads narration live through a screen reader at high speed, the
narration is the review. It happens as the agent works, catching a
wrong assumption before it becomes ten wrong edits. When the agent
goes silent, the only safe recovery is a hard reset, and a reset can
destroy the supervisor's own uncommitted work.

The flag is `RequiresVisibleReasoning` on `ModelFeatures`, a fact
about the model, not an instruction. Enforcement sits in the actor
loop: if no visible text was produced in a round, every tool call in
the batch is rejected before execution. Refusing before execution is
the point, because a tool that has already run cannot be un-run by a
complaint. Every call in the batch is answered, because a vendor
requires a result for each call it made. The error text is long on
purpose. A terse refusal gets optimized against. It also asks the
model to report what surprised it, because a surprise the model
absorbs silently is a bug nobody learns about.

**Engineering standards written from observation.** After reviewing
a new model's first day on the crossover tree, a pattern emerged:
deleted load-bearing comments, unformatted Go committed through
Python string replacements, stray `.bak` files in a public
repository. Each mistake had a specific mechanism of harm.

A deleted comment that read "recall is hooked in TWO places" removed
the only record of an invariant that both call sites still depend
on. The condition it warned about survived; the warning did not. A
deleted warning comment leaves no trace in a clean diff, which is
what makes it the only failure mode no gate catches.

The remedy was explicit engineering standards appended to the agent's
SKILL.md, each stating the reason it exists, so that overriding a
rule means overriding an argument: gate on output not exit status;
use anchored file edits; deleting a WHY comment requires proving the
condition is gone, in the commit message; git is the backup; never
weaken a test; report surprises.

A human team writes down engineering standards after a new
contributor joins and makes the same three mistakes everyone makes.
That the new contributor is a model rather than a person changes
the ceremony, not the need.

---

## 20.6 The Book That Rewrites Itself

Thirty-seven changes and a vendor caching bug. That is the tax on a
crossover, and it is an honest accounting of what the graders failed
to catch.

**What the book got right.** The vendor seam from chapter 19, built
because of a billing dispute and a frozen context window, turned out
to be the mechanism that made the model crossover possible. The
agent already knew how to switch providers; the switch was a table
change, not a rewrite.

The observer pattern from chapter 7 held throughout. The CLI and the
GUI are two renderers over the same actor loop. Every bug in section
20.3 was diagnosed by asking which renderer disagreed with the
other, and the rebuilt CLI made that question answerable without
opening a browser.

The graders caught real regressions across twenty chapters of
cross-chapter sweeps: the underscore filter, a mutable global, a
frozen-snapshot model row that scored 100/100 against the live tree
and 85/100 against the frozen one. Each would have shipped silently
without the sweep.

**What the book got wrong.** It teaches how to build an agent and
pass graders that verify the build. It does not teach how to use
one. Every bug in this chapter would have surfaced from ten minutes
of daily use, which is a category of failure the grading was never
built to see.

It never teaches instrumentation. A session meter, a credential log,
a cache lens: none of that existed before the crossover, and all of
it turned out to be essential. The graders test whether the agent
produces correct output. The instruments test whether the correct
output was produced at the right cost, by the right model, on the
right credential. Those are different questions, and the second set
is the one that matters when the bill arrives.

It also skips the economics of inference. Prompt caching is not a
performance optimization. It is the difference between a system
costing twenty dollars a week and one costing two hundred. Without
understanding cache hit rates, a builder will ship something that
works and then be blindsided by the invoice. That topic earned an
entire chapter (chapter 18) but not the instruments to verify it in
production.

**The recursive part.** The agent described in this book edited the
chapters describing it. The voice pass across all twenty chapters
was run by sub-agents on the same architecture this paragraph was
written with. The next edition will likely be revised by the very
agent it describes, running on the system it replaced.

That is the strongest proof available that the loop closes and that
what comes out the other side is better than what went in.

Every builder's crossover will differ in its specifics. The model
will be different, the vendor will be different, the GUI will carry
different bugs. But the shape will be the same: loud failures that
announce themselves in ten minutes, silent failures that announce
themselves on the invoice, two-path failures that announce
themselves after the first restart, and invisible failures that
announce themselves only when someone builds the instrument to see
them.

Nineteen chapters built the agent. One chapter used it. The length
of this chapter, relative to the other nineteen, is an honest
measure of how far short of full autonomy the book fell, and how
close it got.
