# Chapter 17: The Crossover

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

The first crossover took twelve days. The distance between passing a
grader and surviving a workday turned out to be thirty-seven
changes, a vendor caching bug, and this chapter.

---

## TL;DR

The first-edition account counted thirty-seven changes on the way
to a daily driver. They sort into four kinds:

1. **Loud failures** that announce themselves within minutes of real
   use: part-ID collisions, a settings dialog that never saved, a
   Ctrl-C that killed the process without persisting the
   conversation.

2. **Silent failures** where every request succeeds but something
   underneath is quietly wrong: a model selector that selected
   nothing, a credential provider that selected the wrong credential
   after a vendor switch, a startup resolver that sent an OpenAI
   model to Anthropic.

3. **Two-path failures** where the live streaming path and the
   replay path disagree on what content looks like: text vanishing
   after a restart, resets that cleared the log but not the screen,
   a client that classified messages by guessing at turn state.

4. **Invisible failures** where the output is correct and the bill
   is not: a cache diagnostic tool with a 100% false-positive rate,
   a route where cache reads first appeared around 22,000 tokens,
   and a plan-route caching gap that ended in a bug report rather
   than a code fix.

The existing checks had missed these paths. Daily use exposed them;
focused regressions could then keep them from returning. A passing
check establishes what it observed, which may be a smaller claim
than its name suggests.

This chapter is a field report, with no new graded implementation.
The historical repairs below explain failures to avoid, not defects
to recreate. The final section follows the second-edition agent
through a real Astra/high coding task: reading source, reproducing
failures, editing a file and passing the unchanged tests. That is a
concrete step toward replacing a daily driver; it is not the same
as having written the next edition with it.

---

## 17.1 What Breaks on Day One

The following incidents come from the first-edition crossover and
its [running changelog](../edition-1/book/crossover-changelog.md), begun September 27,
2026. Names such as `Hub`, `view_gui` and the narration gate describe
that implementation. The new implementation has already learned
several of these lessons in earlier chapters.

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
DOM node and appended inside it. The second response appeared
inside the first one's bubble.

The rejected fix was clearing the map on each turn boundary. It
works, but it makes the client responsible for a property the server
should guarantee, and it needs a reliable "new turn" signal before
the first delta arrives. The historical fix was a single monotonic counter
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

The pattern matters beyond any one fix. Real use quickly exposes
whether a tab is identifiable among twenty open tabs, whether a
textarea grows, or whether a process saves on a signal. Those last
two behaviors can also be tested directly. The missing step was to
exercise the interface as someone would actually use it, then keep
that observation in a regression. The expensive category is next.

---

## 17.2 The Bugs That Succeed

The second kind of crossover bug is the one where every request is
answered, the GUI looks right, the tests pass, and the model or
credential underneath is wrong. These share a shape: the visible half of an
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
lifetime constant. In that implementation the model had become changeable at
runtime, which means the vendor had become changeable at runtime,
which means the guard had an expiry date on it.

The runtime switch had already been fixed once before. It moved
model, vendor, surface, base URL, and API key together, carefully,
with a refusal path. Five fields. The credential provider was not
among them, because when that fix was written the provider did not
exist yet.

Flip the picker from Claude to a GPT model and the dialect changes,
the host changes, the key is refilled from the endpoint table, and
the plan credential never arrives. A metered key can answer the
request while bypassing the intended subscription. The changelog
inferred that fallback from the code and a 401 response; it did not
observe the charge on an invoice. A correct answer would not reveal
which credential paid for it.

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
was sent to whichever vendor happened to be the default. The
changelog records a stored Astra selection producing a
healthy-looking session running Sonnet instead. Its diagnosis
specifically concerns model choices outside the `LLM_MODEL`
environment override.

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
with the wrong credential selected. Output-only checks can pass because
the output looks fine. Persistence, billing and identity need their
own observations: a credential log, a cost meter, a cache lens.
Tests can compare those observations with actual requests, saved
files and provider usage. The instrument makes the hidden state
inspectable; it also needs a test that can tell when it lies.

---

## 17.3 Two Paths, One Screen

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
default case, dropping silently when the buffer was full. A newer
progress snapshot can supersede an older one. A text delta, tool
result or replayed artifact cannot be discarded on that assumption.
A long conversation rendered to far more messages than the historical
256-slot buffer, so that repair waited for room with a timeout to
avoid pinning a dead client. Preserve content and terminate a failed
connection honestly; the old queue is not a new design requirement.

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
Refresh requested replay and rebuilt the screen. That gave the
investigation a useful split: inspect the saved record separately
from live notification. In this case the durable record was right;
the live screen was wrong. A refresh alone would not establish both.

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
The comparison narrows the diagnosis: a bug in both points toward
the shared path; a bug in one points toward its inputs, timing or
rendering. Neither result proves the core innocent. Writing a second
renderer against the same seam makes those differences observable
without building a second agent.

---

## 17.4 What You Cannot See Cannot Be Fixed

At large context sizes, the gap between a 91% cache hit rate
and a 22% hit rate can dominate input cost. The crossover record
reports nearly half a weekly allowance consumed in one morning. The correct
answer still arrives. Nothing errors. Nothing retries. The session
looks normal. The only artifact that disagrees is the bill.

Chapter 15 teaches the mechanism. This section tells the story of
what happens when that mechanism meets a production session and the
bill arrives.

**The meter comes first.** A cache breakpoint placed on an unstable
prefix buys nothing while looking like it should have worked, so
proving stability is a prerequisite, not a follow-up. The cache lens
intercepts each request body, splits it into named sections in the
provider's actual cache concatenation order, and diffs against the
prior request. The historical display distinguished MATCH, PREFIX
DIVERGED (with the byte offset), and CACHE MISSED ANYWAY. The last
verdict leaves a cause to investigate: routing, expiry, eviction or
eligibility can differ while local bytes stay stable. Collapsing these into a single
hit-rate figure would make the instrument unactionable.

An early run of the lens caught a bug in itself. Appending a
message to a JSON array shifts the closing bracket, so naive
string comparison flags every single turn as an edit. The alarm
would have fired constantly and been muted inside a day. A false
alarm that always fires is worse than no alarm: it trains the
operator to ignore the real one. The fix was per-section comparison
with grown-container detection: trim the prior section's trailing
framing and check whether what remains is a prefix of the current.

**The alarm that always fired.** The changelog reports a morning
consuming 46% of a weekly plan allowance, alongside 175 requests
carrying 6.9 million uncached input tokens and a 22.5% hit rate.
These are recorded allowance and usage observations, not an
independently reconstructed invoice. The
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

An alarm that labels every harmless append a divergence loses its
ability to distinguish the event the operator needs to see. The
operator stared at it every morning and learned nothing. The
genuine cache break singled out in the record, a rewrite
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
cached write from the prior round could be read back. The earlier
rolling marker already offered that prefix for caching. In the
tested explicit-cache layout, the extra marker repeated a boundary
instead of protecting a different stable section. Marker placement
creates an opportunity; provider usage determines whether a write
or read was granted.

The freed slot went to the tool array, at the very front of cache
order. A tool-boundary marker leaves an earlier eligible prefix
when the system prompt changes. Without it, the next marked boundary
includes that changed prompt, and tool schemas are not small.

The four possible positions are the end of the tool array, the
system prompt, the retained checkpoint boundary, and the last
eligible retained content block before ephemera. A checkpoint may
be represented by surviving memory after a handoff is removed.
Memory remains data; it does not become a mutable system prompt.

**The plan route differed from the fixture.** The first-edition
changelog records a fake and a parser sharing assumptions about
OpenAI's Responses stream. The live plan route exposed rejected
explicit cache controls, absent Content-Type, an empty terminal
output array despite completed item events, and missing provenance
on the assembled tool calls. Fixing one exposed the next.

Chapter 16's second-edition run encountered its own cache-options
rejection and empty-terminal stream failure. Its actual live header
was not retained, so the absent-header parser condition has separate
captured-body and controlled-header proof. Keeping those observations
separate avoids turning a plausible diagnosis into a fabricated wire
capture.

A fake and code built from the same mistaken assumption can agree.
Captured provider behavior can correct the fixture, and a targeted
regression can then catch the parser defect without another paid
request. Live use supplies evidence the shared assumption lacked;
it does not make the repaired test useless.

**The apparent floor.** With the lens repaired, the historical
investigation varied request size in a live session. Despite identical
tools, an identical system prompt, and a stable 99% cacheable
prefix, the provider wrote nothing at 8,000 tokens. Nothing at
10,000. Nothing at 18,000. The first request to cross roughly 22,000
tokens was cached, and the request after it read back 22,144 of
them.

The investigation compared that result with the then-documented
1,024-token minimum and called the difference a floor. Seven
requests establish an observation worth investigating; they do not
establish a universal 22,000-token threshold. Request size changed
alongside session history, and a local prefix estimate cannot show
which provider-side cache entry was available.

Trimming can trade fewer input tokens against less reuse, but a
longer request is not automatically cheaper. Compare actual counts
and applicable prices. Padding a prompt to reproduce one historical
threshold would turn an unexplained result into a cost policy.

Size did not explain all the misses in the record. Its morning session
consumed 46% of a weekly allowance on requests between 55,000 and 73,000
tokens, well above that apparent boundary, while a first-party client using the
same OAuth token but a different endpoint sustained 61-85% hit
rates, including 12,160 cached tokens on the first request of a
fresh session. Same credentials, different route, different caching.

The record says a bug was filed with the provider and the route
remained too expensive for that daily workload. An expected fix
within weeks was a hope in the original account, not a delivery date.
The lesson worth keeping is that an instrument never
checked against an independent measurement is a belief dressed up as
data. The lens ran all morning, reported constantly, and carried zero
signal, because it was only ever compared to its own logic.

---

## 17.5 The Agent Looks in the Mirror

The historical agent had a GUI and a built-in MCP server but still
could not inspect its own screen through the active configuration.
The server had been built and documented in original Chapter 13:
`gui_snapshot`, `gui_click`, `gui_input`. Nothing connected to it.
A stdio client check did not establish the missing path from agent
to browser. The real browser/skill path in Chapter 12 addresses that
boundary in this edition.

**Two doors, one per direction.** The historical `--mcp-port` relayed the GUI's MCP
server onto loopback TCP, so any external agent that speaks MCP can
connect and drive the browser. Its `view_gui` builtin
asked the connected GUI for a snapshot over the existing WebSocket.
One exposed the browser to an external client; the other made it
reachable from the local agent.

The changelog records a live check: prompted to list its buttons,
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

The one-line fix broke the original Chapter 15 check. The MCP connect path recorded
enabled tools under the normalized key, under a comment explaining
why. The old convention was internally consistent; the fix moved one
end of it. The cross-chapter sweep caught the regression. Bisecting
in a worktree localized the commit, and a second fix recorded the
declared name instead.

**The model that worked in silence.** Some models reason at length
internally and then act without narrating. For a supervisor who
reads narration live through a screen reader at high speed, the
narration is the review. It happens as the agent works, catching a
wrong assumption before it becomes ten wrong edits. The changelog
records a silent session followed by a hard reset that also
discarded the supervisor's uncommitted work. The missing narration
made review harder; the destructive recovery added a second failure.
Interrupting the current turn and inspecting the actual diff
preserves more options. Interrupt leaves the Agent and other jobs
alive, so those jobs still need deliberate supervision.

That implementation added `RequiresVisibleReasoning` on
`ModelFeatures`, with enforcement in its actor loop: if no visible
text was produced in a round, every tool call in
the batch is rejected before execution. Refusing before execution is
the point, because a tool that has already run cannot be un-run by a
complaint. Every call in the batch is answered, because a vendor
requires a result for each call it made. The error text is long on
purpose. The message explains the rule and a concrete next step.
It also asks the model to report what surprised it, giving the
supervisor a chance to investigate a mismatch. This historical gate
is not a new requirement of the prose-only chapter, nor a request
for private chain of thought. Useful progress says what is about to
change and why; its truth still has to agree with the actual work.

**Engineering standards written from observation.** After reviewing
a new model's first day on the crossover tree, a pattern emerged:
deleted load-bearing comments, unformatted Go committed through
Python string replacements, stray `.bak` files in a public
repository. Each mistake had a specific mechanism of harm.

A deleted comment that read "recall is hooked in TWO places" removed
the only record of an invariant that both call sites still depend
on. The condition it warned about survived; the warning did not. A
compiler can accept the deletion while both call sites still depend
on the invariant. The diff does show the missing explanation; a
reviewer has to ask whether its reason survived.

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

## 17.6 The Book That Rewrites Itself

Thirty-seven changes and a vendor caching bug were the first
account's tally. The useful part of that tally is the gap it names:
building a feature and trusting it with a workday are different tests.

**What the book got right.** The vendor seam from Chapter 16, built
because of billing and context-control constraints, turned out
to be the mechanism that made the model crossover possible. The
agent already knew how to switch providers; the switch was a table
change, not a rewrite.

The observer pattern taught in Chapter 6 held throughout those
repairs. The CLI and the GUI are two renderers over the same actor
loop. The rebuilt CLI made the comparison in §17.3 possible without
opening a browser.

The graders caught real regressions across twenty chapters of
cross-chapter sweeps: the underscore filter, a mutable global, a
frozen-snapshot model row that scored 100/100 against the live tree
and 85/100 against the frozen one. Each would have shipped silently
without the sweep.

**What the crossover changed.** The exercises teach individual
mechanisms. Daily work combines them under conditions a short
fixture may never reach: another turn, a restart, a model switch,
a long conversation. Those conditions belong in focused checks once
the failure is understood. Ten minutes can expose a broken control;
it cannot certify a reliable workday.

The first account overreached when it said instrumentation had been
missing: the original cache chapter already taught the lens and
meter. Chapters 17 and 18 carry those instruments and the credential
report forward. They ask whether a
correct-looking answer came from the selected model and credential,
and what usage accompanied it. An instrument with a wrong input
field can fail as quietly as the feature it measures. Comparing it
with independent requests and usage closes that loop.

Inference economics changes the engineering choices. Prompt caching
can reduce input cost as well as latency, but a rate difference is
not a measured weekly bill. Use the meter to see usage in each
category, the lens to understand local changes, and the actual
provider counters to determine whether reuse occurred.

**The recursive part.** The first-edition account reports that the
agent edited the chapters describing it, including a sub-agent voice
pass. That is a use of the tool worth preserving, not proof that
sub-agents already belong to this chapter's implementation or that
each regenerated edition must improve. The second edition needs its
own demonstration.

A builder's crossover will differ in its specifics. These four
categories give the investigation a place to start: loud failures that
announce themselves in ten minutes, silent failures that announce
themselves on the invoice, two-path failures that announce
themselves after the first restart, and invisible failures that
announce themselves only when someone builds the instrument to see
them.

The next test is a piece of work that has to be right outside the
agent's own explanation. Start there, while the implementation is
still familiar enough to debug.

---

## 17.7 A File That Actually Changed

The second-edition crossover began with a smaller claim. Could the
new Ensemble run the model Bill selected, repair a real program
through its tools, and leave a result another process could verify?

The selected model was `gpt-6-astra` at high reasoning effort. Its
name already existed in the model table, but the table sent it to
Chat Completions. Naming the right model did not select the right
transport. The October 11, 2026 [model documentation](https://developers.openai.com/api/docs/models/gpt-6-astra)
supports high effort, and the [reasoning guide](https://developers.openai.com/api/docs/guides/reasoning)
requires Responses for Astra function calls. Moving that row to the
existing Responses renderer was the entire runtime change. Agent
settings already carried the effort; Chapter 16 already supplied
the credential and stream handling.

The task was a small CSV spending report. Its implementation split
the file into lines, then split each line on commas and added the
amount as a binary float. A merchant called `Books, Inc.` defeats
the parser; a quoted newline defeats the line splitter; an ordinary
refund can expose the arithmetic. Six tests fixed the expected
behavior before the agent started.

The actual CLI session read `aggregate.py` and `test_aggregate.py`,
then ran `python3 -m unittest -v`. The result was three failures and
three errors. Before editing, Astra explained the two causes:
manual splitting broke CSV structure, and floats lost decimal
precision. It then used `edit_file` to replace that code with the
standard library's `csv.reader`, a file opened with `newline=""`,
and `Decimal` arithmetic. Blank rows were skipped. The sorted,
tab-separated, two-decimal CLI output stayed intact.

The next command ran the same six tests. All passed. An independent
rerun of the preserved before and after directories reproduced both
outcomes; the test file was byte-identical. The useful evidence is
the [actual source diff](../solutions/evidence/ch20/task.patch)
beside the failing and passing results, rather than the model's
claim that it repaired something.

Five requests completed in the recorded 27.46-second run, and every
request and terminal response named Astra/high. Subscription billing
was selected explicitly with no API key configured. The CLI exited
normally after EOF and wrote its save. The endpoint follows from
the configured base URL and transport source; the payload log did
not independently capture the HTTPS URL or authorization headers.

There were 161 ordinary text deltas and no public reasoning-summary
deltas, despite requesting detailed summaries. Every response
reported zero cache grants and zero reasoning tokens. Those fields
do not measure private mental activity, and successful repair does
not establish cache savings. Ordinary progress arrived before the
edit, so the useful explanation was present even without a public
summary. This was one bounded coding task, without a steering or
GUI demonstration. It establishes more than a configuration change
and less than general project reliability or authorship of Edition 3.

### Try the same task

There is no grader for this chapter. To try the recorded task,
start in the repository root and build the accepted Chapter 17 CLI:

```sh
chapter19_repo=$PWD
(cd solutions/ch17 && go build -o /tmp/ensemble-ch19 .)
chapter19_work=$(mktemp -d)
cp solutions/evidence/ch20/task-before/aggregate.py "$chapter19_work/"
cp solutions/evidence/ch20/task-before/test_aggregate.py "$chapter19_work/"
cd "$chapter19_work"
mkdir -p skills/base
cat > skills/base/SKILL.md <<'EOF'
---
name: base
type: primary
description: Repair the local CSV program using its unchanged tests.
tools: read_file edit_file run_command
---
Read the source and tests, reproduce the failure, explain the intended edit,
then change the source and run the unchanged tests. Work in this directory.
EOF
```

This reproduction uses an explicit primary skill; the recorded live
run used the default registered tools. Keep `EN_HOME` set to your
own dedicated Ensemble profile from Chapter 16, and use `auth status`
and `auth models` to check sign-in and the account's available model.
If needed, sign in through the same Chapter 16 flow. Do not borrow
another application's credential files.

```sh
/tmp/ensemble-ch19 auth status
/tmp/ensemble-ch19 auth models
LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
OPENAI_BILLING_MODE=chatgpt OPENAI_REASONING_EFFORT=high \
OPENAI_REASONING_SUMMARY=detailed EN_PRIMARY_SKILL=base \
/tmp/ensemble-ch19 --skills-dir skills chat
```

Use the recorded prompt:

```text
Repair aggregate.py in this scratch directory. Read aggregate.py and test_aggregate.py using read_file, then run python3 -m unittest -v with run_command before editing to observe the failures. Make the smallest clear fix: totals(path) must return exact decimal currency totals grouped by merchant, support normal CSV quoting including embedded commas, escaped quotes and newlines, skip blank rows, and retain the CLI's sorted tab-separated two-decimal output. Use only the Python standard library. Do not change test_aggregate.py or read files outside this workspace. Edit the actual source using a file tool, run the unchanged tests again, and give a concise report of the bug, fix and real results. Do not add unrelated features or just describe a patch. Briefly report progress before editing.
```

After the final answer, send EOF. Check the files yourself:

```sh
python3 -m unittest -v
cmp "$chapter19_repo/solutions/evidence/ch20/task-before/test_aggregate.py" test_aggregate.py
```

Read the source diff, too. Passing these six cases says something
specific about this repair. It does not make the agent its own
reviewer. That distinction is what keeps the next piece of work
small enough to trust and concrete enough to improve.
