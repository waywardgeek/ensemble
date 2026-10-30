# Chapter 26: Crossing Over to Ensemble

“I gave Astra a book. Astra gave me an AI coding agent.”

That is how I described the earlier crossover. The book included graders. Astra
built Ensemble while working in Codex, and then I began working with Astra inside
the agent we had built. The next edition was built inside Ensemble. The loop
closed, but closing it once did not make every later replacement ready to use.

When I opened the C# edition, I saw settings above a conversation, arranged in one
long column. There was no useful agent tree in the interface I was looking at.
The controls and the tests existed; the application did not yet feel like the
place where I could do my work. I asked for a real GUI.

That is not cosmetic work to do after the engineering. It is engineering. I can
hardly see, and I listen quickly. If I cannot find the selected agent, hear the
right answer, follow the explanation beside the tool results, or interrupt work
without guessing which control will move next, the application is not ready to
replace the one I use now.

This is the final implementation chapter. Its result is a replacement candidate
and a safe way for the human to evaluate it. The crossover is complete when the
human accepts that candidate for ordinary work—not when the student declares it
finished, and not when an author adds up earlier test counts.

## TL;DR — finish the replacement, not another demonstration

Start from the completed C# application in `solutions/edition-3/main`. Do not
rebuild it from nothing or transplant the old Go GUI. Implement the missing
crossover work, using the ownership and behavior established in the preceding
chapters:

1. Make the GUI a coherent, accessible three-pane application using the approved
   Bootstrap control system and graphite/blue design direction below, with a
   stable status bar, agent tree, settings, chat, and tool artifacts.
2. Provide an ordinary foreground `start` launcher and actionable startup errors.
   Make it possible to run beside the current application without stopping it.
3. Verify the real browser, actor, persistence, and shutdown paths with owned
   local resources. A screenshot alone does not verify a control; a socket test
   alone does not verify a usable interface.
4. Prepare a safe, explicitly armed crossover and rollback procedure. The
   current application and its data remain recoverable throughout the trial.
5. Give the human a runnable candidate, clear instructions, the tests you actually
   performed, and the limitations that still need human or live-provider work.

This is the only crossover chapter in Edition 3. The earlier historical
crossover chapters have been removed and the manuscript renumbered sequentially.
Preserved source endpoints and grader commands retain their original names;
see the numbering map in [the edition guide](README.md). Earlier editions remain
historical records, not files to edit here.

## In plain words — one system, a better place to use it

The GUI is a projection and control surface over the existing application.
Ensemble owns Agents; Agents own their work, History, Engine, and related
services. The browser does not become a scheduler, credential owner, transcript,
model registry, memory database, or usage meter because it needs a nicer display.

Begin by reading the current implementation. The existing GUI has peer creation
and deletion, an agent selection protocol, model changes, settings acknowledgments,
usage, memory and recall controls, Goals, speech, and tool-card rendering. Most of
the work is to arrange and connect those capabilities well. Replace an awkward
presentation, not a correct ownership model.

C# changes still follow the Common/implementation structure. Common declares
persistent state, relationships, and externally callable operations; functional
modules implement them. Real parents provide access to services. Destruction
completes both ends of relationships. Value-like display data need not acquire
invented lifecycle machinery. Explain any necessary vocabulary change before
implementing it; ordinary work within the agreed model needs no per-function
permission.

Teaching code must remain readable. Comment relationships, contracts, and
non-obvious browser lifecycle decisions. Use the approved component library
below; do not build another framework, configuration language, second event bus,
or collection of copied state around it. Keep the existing HTML/JavaScript
integration and C# host rather than adding a React or Vue rewrite merely to use
styled controls.

## Required GUI reading

Before implementing this crossover, read the complete original
[Chapter 8, “Build Your Dream GUI”](../chapter-08.md), as well as the current
[C# Chapter 8](chapter-08.md). The original chapter explains the design intent
and why using the interface matters; this chapter is not a substitute for it.
Check your candidate against that intent and its still-applicable behavior.

Earlier Go names, grader routes, and superseded implementation details are
historical, not instructions to replace the current C# architecture. Current
contracts and this chapter's safety boundaries govern. If the readings expose a
genuine conflict, explain it rather than silently choosing a convenient version.
Do not edit the earlier edition. Report the relevant checks in your handoff.

## The selected control system and visual direction

The first crossover candidate improved the layout, but I still found it ugly.
“Good enough for now” was not an endorsement of its design or acceptance as my
replacement. We compared two interactive component-library samples. I chose
**Sample B: Bootstrap**, with a graphite background and blue accents, rather than
the navy/teal Web Awesome sample. Do not reopen that choice as an unspecified
framework-selection task or continue polishing the first candidate's homemade
control system.

Use **Bootstrap 5.3.8**, the stable version used in that sample. Its CSS and
JavaScript supply the common control vocabulary: genuine tabs, buttons, forms,
menus where needed, and consistent focus/disabled states. Read the
[official Bootstrap 5.3 documentation](https://getbootstrap.com/docs/5.3/getting-started/introduction/)
and [distribution instructions](https://getbootstrap.com/docs/5.3/getting-started/download/).
Pin the dependency and retain its MIT license and provenance. Bundle the required
assets locally with the optional GUI; do not depend on a runtime CDN, remote
fonts, or an unrequested package installation. A Node build system is not needed
just to use the published browser assets. Use one control system, not Bootstrap
alongside a second component library.

The temporary localhost sample is not a required service or an answer key. These
are its durable design requirements:

- **Graphite surfaces, light text, blue accents.** The selected sample used dark
  page `#15191f`, pane surface `#1f252e`, raised surface `#29323f`, main text
  `#f1f4f9`, secondary text `#bbc6d5`, border `#455365`, accent `#a8c8ff`, and
  selected background `#263e64`. Use these as the initial application palette;
  verify actual foreground/background combinations rather than assuming the
  library guarantees contrast. The light counterpart used `#f5f6f9`, white,
  `#e9edf5`, text `#17253b`, secondary text `#4a5c75`, and accent `#2359a6`.
- **A quiet shell and a clear hierarchy.** Compact status controls, aligned pane
  headings, consistent spacing, and breathing room around messages. Give the
  conversation the main reading space. Avoid enclosing every label or sentence
  in another bordered box. Reserve tool cards for distinct activities/artifacts.
- **Tabs that select a shared panel.** Agents/Settings are a connected tab strip,
  not pill-shaped action buttons. The active tab joins or clearly identifies its
  panel; inactive tabs recede. Keyboard focus is distinct from selected state.
  Use the library's tab behavior without adding competing click/key handlers.
- **Consistent controls, not the sample's fake application.** Keep the button,
  form, spacing and theme treatment, but bind it to the real acknowledged Agent
  operations described below. The sample's conversation, timings, artifacts,
  creation/deletion, pause and speech behavior were illustrative. They are not
  production data, a scheduler, persistence, or accepted test results.

Bootstrap does **not** supply Ensemble's agent relationships, tree semantics,
pane splitters, transcript or tool renderer. Keep those responsibilities in the
existing application/browser layers and style them consistently. A nested list
of sample buttons is not a keyboard-accessible real agent tree. Use the existing
ownership and operations, implementing missing presentation behavior without
rebuilding the core. Do not add another tree or layout framework without a
concrete need and architectural discussion.

Integrate the existing acknowledged appearance preferences with Bootstrap's
light/dark presentation. Do not create a second localStorage authority, lose
saved explicit choices, or apply a rejected setting optimistically. New users
start dark. Library event handling must not duplicate sends, resets or peer
operations; attaching it again after replay/reconnection must not accumulate
handlers. Preserve target identity while updating the presentation.

Verify locally bundled CSS/JavaScript in the ordinary embedded build **and** a
mounted GUI base path. Check for failed asset requests and unintended external
requests, real tab/keyboard behavior, both themes and enlarged text. A library
makes controls consistent; it does not certify contrast, beauty, accessibility,
or correct application wiring. Carry these checks into the integrated exercise,
not only a standalone component demo.

## The interface I require

### A stable status bar

Put a status bar across the top. Give its contents reserved space so changing
status and numbers do not push the controls around. Keep proportional artwork on
the left and one compact row to its right. Put each speaker’s name above their
message bubble, with compact spacing for both human and Agent text. Distinguish
Agent messages with a blue fill and stronger leading border, not red-dependent
color differences alone. Preserve text and link contrast in both themes; keep
the speaker names as an explicit identity cue.
Do not repeat the Agent name under the artwork or in a conversation-title row.
Show connection/activity,
then the model selector, centered in the full header rather than merely its
remaining left-hand space. Center usage in the remaining space before the action
controls, without claiming unknown amounts are zero. Group Cost, Input, Cache,
and Hit closely: each numeric output gets a fixed-width, left-aligned slot beside
its label. Tabular digits help changing numbers remain still. Long values may
truncate with the full value available accessibly; they must not shove the
controls off screen. Do not repeat the selected model or funding in another row.

Reserve the geometry, not just the labels. Wrapping flex rows and tabular digits
alone do not stop a longer state or usage value from changing neighboring widths
or row heights. Bound identity and numeric slots; give model and action controls
stable tracks. At enlarged text or narrow widths, use an explicit responsive
arrangement with reachable overflow rather than shrinking text. Pause/Resume
state changes must not make adjacent controls jump either.

The status bar contains:

- A model dropdown, sized to its current name within responsive limits,
  using the existing same-provider, between-turn model-change operation. Choosing
  a model sends the change immediately; there is no separate Apply button. The
  server acknowledgment, not the click, establishes the displayed selection. A
  refused change must restore that selection. Keep all groups on one vertically
  centered row without shrinking the fonts.
  Compatible model names are not proof of account entitlement. Preserve the
  distinction between running selection and persisted launch configuration.
- A speaker-icon Speech control with an accessible name and native checked
  semantics. Speech remains opt-in. Its state is the acknowledged appearance/
  speech setting, not an independent second toggle. Per-message Speak and Stop
  speech also use named speaker icons. Retain keyboard access and visible focus.
- Compact Pause/Resume and Interrupt controls with useful labels, keyboard focus,
  and adequate click targets. Do not confuse pausing admission with interrupting a
  turn, or interrupting a turn with killing independent Jobs.
- A theme button showing a moon for dark and a sun for light. Its accessible
  name must state the switch action, not leave the meaning of the icon ambiguous.
  Its setting must persist through the existing appearance store and agree with any equivalent
  settings control. Use dark for a new user unless a saved explicit choice says
  otherwise. Both themes need high contrast.

Use the human's approved artwork proportionally for branding and the browser
icon. Do not silently crop or redesign it, or leave a sample's placeholder in the
working application. Bundle the assets locally with appropriate content types.

An unavailable connection is not an idle Agent. Pending selection and settings
changes are not acknowledged state. Disable controls when their target is
uncertain and show why rather than silently applying them to a different Agent.

### Named collaborators

Show each speaker's name above the message, including in accessible Speak labels,
not the generic labels “You” and “Assistant.” Agent names derive from the real
Agent tree; do not invent a second identity or silently rename an Agent's files.
Settings includes a User name and short description in the existing settings file,
with save-before-acknowledgment and a neutral fallback when no name is configured.
The GUI reads those fields directly. At startup its host passes them into the
existing skill variables as `$HUMAN_NAME` and `$HUMAN_DESCRIPTION`; `$AGENT_NAME`
derives from the actual Agent. The primary skill describes the named engineering
collaboration. Preserve one-pass substitution, immutable activated manuals, and
historical messages: changing a display name does not rewrite earlier instructions
or provider protocol roles. Test save/reload, save refusal, literal user text, and
name placement with real browser geometry.

### Three panes below it

On a desktop, all three panes are visible at once.

**Left: Agents and Settings.** Use two genuine tabs. The Agents tab contains an
expandable hierarchy and agent creation/deletion controls. Represent the real
parent relationships: named top-level peers are siblings, not children of the
first displayed Agent. Show the selected row and meaningful status. Support
keyboard traversal, expansion, selection, and visible focus without requiring a
mouse or color perception. Compact styling must not erase the hierarchy: give
children clearly visible indentation and branch guides, and reserve the expander
slot on leaf rows so their names align with branch names at the same level.
Preserve explicit expand/collapse choices across new snapshots. Test real parent,
child, and grandchild geometry as well as keyboard behavior.

The tree is the sole visible conversation selector. Use a **Create Agent** button
that opens a dialog, not a permanently expanded creation form. Put confirmed trash
actions on eligible tree rows, not a second name-entry deletion form. Protect the
mounted Agent and its owning ancestors in both presentation and server enforcement.
Creation and deletion are acknowledged operations, not optimistic edits to a local
list. Reuse registry-aware operations for top-level peers, including refusal when
they have unresolved children. Explicit child deletion cancels and joins its owned
subtree before logical destruction; it does not edit the top-level registry or
delete saved files. Test confirmation cancellation, server-side protection, view
recovery after deletion, and actual retention of saved submission files.

Submission is result publication, not a request to stop or remove a child. Keep
submitted children selectable and show their submitted status alongside activity.
A status-tool query and a GUI snapshot must not hide them merely because their
work is complete. Keep explicit lifecycle operations separate from observation.

A selection changes the view; it does not transfer that tab's browser relay or
MCP authority. Remove redundant selection/relay explanations from the normal
sidebar, but keep a compact separate-relay action discoverable for Agents that
require it. A pending creation must leave the tree usable for that startup step.

The Settings tab has clearly grouped **AI**, **Memory**, and **Accessibility**
sections. AI includes the existing model/context/tool-policy and account/funding
information where applicable. Memory includes the existing band and recall
settings. Accessibility includes theme, font size, voice, speech speed and pause
behavior. Keep Goals and existing account controls reachable too; reorganizing
controls is not permission to drop them. Memory is a settings group, not another
persistent memory authority or a new recall corpus viewer. A supplemental dialog
must not replace or bury this agreed Settings tab, duplicate its state, or make
its grouped controls harder to reach.

Appearance settings belong to their existing store. Memory/recall settings
belong to the selected Agent. Preserve durable-save-before-acknowledgment,
sibling-policy preservation, draft protection, and rejection of stale selection
frames. Funding stays fixed for a running Agent. Codex renewal remains Codex's
responsibility; public OAuth operations must not turn into Codex operations.

**Middle: Chat.** Give the conversation the main reading space, with a composer
anchored at the bottom. The composer has exactly two visible widgets: a one-line
input and the supplied paper-airplane Send icon beside it. Longer or multiline
input scrolls within that one-line area rather than taking height from the
conversation. Name the selected Agent in the placeholder, such as “Ask Cadence
anything…”, and keep an accessible input name. Do not add a visible label, help
footer, or separate Hint button. Preserve the supplied artwork and its proportions.

The browser sends ordinary messages, including while the Agent is working. The
Agent classifies input against its active turn when the event queue receives it;
stale browser state must not select a different “hint” mode. Preserve Enter to
send and Shift-Enter to enter a newline, without a special forced-hint shortcut.
Keep ordinary explanations visible while tools run. Preserve selection/replay,
streaming and final replacement, near-bottom-only autoscroll, and scroll position
when the human is reading older material. A new answer must not silently replace
an earlier one.

Density is part of usability, not an excuse to shrink fonts. Reduce unnecessary
widget padding/gaps and paragraph margins while preserving the selected font and
line spacing, clear grouping, focus indicators and usable click targets. A short
message should not require a separate full-height action row above its text.
Measure identical content before and after at the same fonts and viewport; retain
text and semantic markup rather than hiding content to manufacture a smaller box.
Let prose use the available pane width. Pane splitters should allow Actions or
conversation to take most of the workspace, with small reachable minimum panes,
not arbitrary 18–35% side-pane caps. Check mouse/keyboard movement in both
directions and re-clamping after window/font changes.
Private reasoning and opaque replay payloads are not chat text.

**Right: Tools and artifacts.** Show tool calls, results, running work, failures,
and available artifacts in coherent expandable cards. A pending call must not
look like a successful result. Preserve exact call/job identity and existing
controls; use the existing renderer and artifact capabilities rather than
inventing an unrelated file browser. Long output should be inspectable without
forcing the conversation to scroll sideways. Render untrusted text safely, not
as executable provider-supplied HTML.

Exercise automatic tool-result redaction on the selected provider route, not
only unredacted sample results. A retained `RedactedPart` is a readable omission
notice, not media. Project it through the common text formatter without exposing
its private locator or mutating History; keep genuine unsupported media refusal.
A valid saved redaction must not make every subsequent request fail locally.

Inline truncation is not permission to make the retained result unreachable.
Provide a usable way to inspect or download the complete output when a current
owned Job has it. Resolve an Agent and its Job through the existing ownership
operations, including canonical qualified addresses for nested Agents. A named
top-level lookup is not a substitute for GUI tree/address resolution. Never
accept a browser-supplied filesystem path. Open the owned output
with the existing `Sandbox.OpenRegularRead` operation: an ordinary unconfined
`OpenFile` can wait for a pipe peer if a file has been substituted, whereas the
regular-file reader refuses nonseekable input without that wait. Dispose the
returned stream and preserve the same confinement authority.

Nonblocking regular-file access is not a blanket symlink ban. An empty Sandbox
root deliberately retains full access; a rooted Agent enforces its existing
confinement and permits admissible inward links. Do not invent a second Sandbox
or a GUI metadata/ancestor-checking scheme: checking then opening does not prevent
concurrent substitution. Test that a rooted output redirected outside its root
refuses without returning escaped bytes, separately from nonseekable-file refusal.
Preserve the existing unconfined policy and report its limits. Job ownership and
completion do not authenticate original file bytes against hostile local writers;
stronger immutable-artifact provenance would require a separately agreed design.

Preserve the GUI's origin/header guards, mounted base path, request cancellation, and private-response
headers. A convenient ordinary download link is not a reason to replace the
required custom-header guard with optional browser fetch metadata. Mount shutdown
must cancel and join the active response, not just stop future admission. Reuse
the available core operations; if a genuine gap requires a new public operation,
explain the need and obtain the shared-vocabulary agreement before adding it.
Reading a download must not consume the Job observer's output cursor.
Refuse running, missing, wrong-owner, or historical-only handles explicitly rather
than returning partial output as complete. Bound browser memory use and explain
any size refusal; an unbounded `fetch().blob()` is not streaming to disk. Test
exact bytes, ownership refusals, and a later ordinary Job observation.

Make pane widths adjustable with pointer and keyboard controls. Panes should
scroll independently, with useful minimum widths and without hiding the chat
composer. At narrow widths, provide an explicit way to reach each pane rather
than squeezing three unreadable columns or dropping functions. This is not a
request for a separate mobile application.

### Deliberate appearance, not a color swatch on the old page

I prefer fairly dark backgrounds and light text. Choose restrained, attractive
accent colors, consistent spacing, readable typography, clear section headings,
and controls that look related. Slightly different dark surfaces can distinguish
panes; low-contrast gray labels cannot. Tabs must look and behave like tabs.
Selection, focus, warning, and error states must not rely on color alone.

A component library does not decide the visual hierarchy. When the human selects
a reference, study the actual page: its connected tabs, navigation rows, prose,
spacing, and card boundaries, not merely which library it imports. Ordinary
status information is not an action; navigation should not resemble a row of
primary buttons. Keep the reference's visual discipline while replacing its
illustrative interactions with the real application's ownership and behavior.
Use clear conversation labels without inventing identities, times, or activity.
Compact action icons need accessible names and generous targets. A restrained
appearance must not mean faint text or indistinct input boundaries.

Check both themes, enlarged text, keyboard use, and long content in an actual
browser. Combine the conditions: test the largest supported text size in narrow
windows as well as on a desktop, not large text only on a wide screen and narrow
screens only at the default size. Include long status values and an expanded or
resized composer. Verify that the prompt, Send, model, theme, and pause/interrupt
controls remain reachable, with a useful conversation reading area and no
page-wide horizontal overflow. Hiding overflowing controls or silently shrinking
the chosen text size is not a correction. Bounded status summaries, shorter
navigation labels, and accessible secondary help can preserve space without
dropping operations. Keep screenshots, but also assert geometry and keyboard
reachability; a successful settings save does not establish a usable layout.
A bounded composer with its own scrollbar can still leave Send outside the
visible area after textarea resizing. Reserve space for its label and action row,
let both the form and textarea yield space, and test short phone-height windows
as well as tall ones. Include nonempty notices and settings acknowledgments:
those rows also consume space. A flexing textarea inside a fixed outer form can
still overflow after other rows grow. Do not call an off-screen button anchored
merely because its outer form is.
Check computed contrast as well as the layout. Capturing a screenshot is not
viewing its pixels: say whether the acting model actually received and inspected
the image, or only examined DOM/style/geometry. If visual input is unavailable,
retain the screenshot for the human and report that limitation; do not claim a
visual review or invent a hidden vision capability. Do not invent a numerical
beauty grade. The human decides whether the application is pleasant and usable;
the student must still bring a considered design.

## Make startup ordinary and failure useful

Provide a simple executable `start` script in the working solution, documented
for this C# edition. It runs the GUI in the foreground, forwards supported launch
options, exposes errors, and stops cleanly with Ctrl+C. Explain its build
prerequisite and how it locates the SDK, assets, skills, workspace, and data.
Do not silently install dependencies, fetch credentials, kill a port owner, or
rebuild an unrelated running installation. Keep machine-specific account paths
out of the repository. Use an explicitly supplied credential path when the human
selects Codex; a path alone must not silently select the payer.

The GUI library normally embeds its assets. Verify the built application's
ordinary launch, not only an explicit `--gui-dir` source override: a CSS edit
cannot change an already built assembly or an already loaded tab. Keep the
existing borrowed-host and base-path behavior intact, with asset and socket URLs
relative to the mounted page. Do not make the optional GUI a core dependency.

The current host accepts `--port`, `--workspace`, `--sandbox`, `--data-dir`,
`--agent-name`, `--gui-dir`, and `--skills-dir`. Codex selection uses
`--billing codex --codex-auth-file PATH`; its endpoint is pinned. Inspect the
saved parser rather than assume other flags exist. Provide a parallel-run
example on an unused port, not instructions to evict the working server. A human
may run something like the following after building:

```sh
# Choose a genuinely unused loopback port and a private trial home first.
# This is a human-authorized live launch, not an offline test command.
./start --port "$TRIAL_PORT" --data-dir "$TRIAL_HOME" \
  --billing codex --codex-auth-file "$APPROVED_CODEX_AUTH_FILE"
```

The first manual trial used the default port while the old Ensemble still owned
8088. The GUI printed only `GUI failed: AggregateException`. Starting on 8089
worked. That is evidence consistent with a bind conflict; the type-only message
did not diagnose the underlying exception. Improve startup reporting at the
proper ownership boundary: show an actionable, safely masked cause while the
relationships needed for diagnostic masking still exist. Do not dump nested
exceptions, request headers, tokens, or private paths merely to improve debugging.
Test bind failure, malformed setup, and cleanup faults without hiding the primary
error. Retain defensive type-only reporting where safe redaction is unavailable.

A separate remote trial restored the old server after thirty minutes. It left no
provider request in the V3 API trace. A browser log showed a request to select
`main` while the V3 actor was named `trial`; that observation did not establish
why the remote trial failed. Inspect stale-browser and invalid-selection behavior
without turning this clue into a claimed diagnosis. Give the human clear reload
instructions and show a selection error without leaving controls falsely ready.

## A safe trial and an independent way back

Preserve the existing executable, launcher, configuration, history, and actor
homes. Begin with separate trial data; do not copy private stores or let a new
application overwrite the current session. The initial student run is not a
migration of Cadence's memories. A later migration needs an explicit inventory,
format compatibility checks, a copy rather than an in-place experiment, and the
human's approval.

First run the new GUI beside the current application. Only after the human is
ready should a deliberately armed switch occupy the familiar port. The rollback
controller must outlive both GUI processes and their tool Jobs. A shell command
backgrounded under the application is not enough: application shutdown may kill
its process group or owned jobs.

On macOS, the earlier trial used a separate per-user launchd job, with a private
temporary definition, automatic restart disabled, and an explicit start. A probe
showed that bare `launchctl submit` could remain waiting; acceptance by launchctl
was not proof the controller ran. Require an observed armed state before stopping
the original process. Do not install a login-startup service or a global daemon
merely to run this trial.

Prepare a one-shot, approximately thirty-minute rollback with an early-restore
command. It must check the original process identity and port ownership before
acting; never terminate an unexpected owner. Allow graceful old-server shutdown,
and do not force-kill it to make a trial proceed. Restore early if the candidate
exits or cannot be launched at all. A missing or unexecutable candidate may throw
before there is a process to monitor; that exception must still reach the restore
path after the original has stopped. Bound the candidate's shutdown, and restrict any forced termination to
processes the trial actually owns. Start the restored application independently
of the controller that is finishing. Preserve status and private logs, reporting
uncertain or failed restoration honestly. A slow original shutdown must not
cause the controller to disappear and leave an empty port later.

Test those lifecycle rules using real disposable local servers before offering a
live switch. Cover timer expiry, early restore, candidate launch failure and crash,
a candidate ignoring termination, changed-owner refusal, and cancellation before the switch.
An armed timer is not proof of recovery. An observed restored listener is not
proof of remote connectivity or successful inference. Power, network, and OS
failures remain outside a software controller's guarantee.

The earlier private trial controller is experience, not a hidden student
solution or a permanent service to borrow. Deliver and test the small launcher
and recovery support in the working solution. Do not read the human's private
trial directory as an answer key.

## Exercise the behavior that daily use exposes

Use the real C# host and an isolated browser with a fresh profile, plus a small
working provider fake where deterministic replies or failures are needed. The
fake is a boundary peer, not a mock of Agent, History, or the GUI itself. It must
speak the selected transport faithfully. Merely assigning a dummy key does not
make a public endpoint an offline test.

A useful session proceeds through these paths:

1. Launch on an owned spare port. See the three panes, the selected agent, and a
   truthful connection state. Change status/usage text and show that the status
   bar controls do not move around. Compare bounding rectangles before and
   after realistic long agent names, status values, and multi-digit usage/cost
   values at a fixed viewport. Checking only that controls are inside the window
   does not establish reserved space. Label a display-value geometry probe as
   such; it does not verify backend accounting. Inspect dark and light themes, large text,
   keyboard navigation and pane resizing. Keep screenshots for the human.
2. Send a message through the real browser. Observe explanation, streaming,
   a tool call and result, and a final reply in the correct panes. Use enough
   output to test scrolling. Preserve the existing safe Markdown behavior:
   unordered and ordered lists must remain semantic lists, and source-wrapped
   prose must remain one paragraph rather than a stack of separate paragraphs.
   Check headings, inline code, links, and fenced code without executing supplied
   HTML. A styling change is not permission to replace a working renderer with a
   less capable one. Reuse the working block parser when adding inline emphasis
   or expandable code; test semantic elements and displayed whitespace, not only
   a screenshot or text-content match. Inspect a real artifact supported by the renderer.
   Disconnect/reconnect and verify repair without duplicate cards or lost text.
3. Create a peer, see its actual place in the tree, select it, return to the
   original, and delete the eligible peer through the existing operation.
   Demonstrate a refusal without pretending deletion succeeded. A displayed
   child must not grant its parent-tab browser authority.
4. Change a model between turns and exercise the existing busy/refusal path.
   Change appearance and selected-Agent memory/recall settings; observe the
   acknowledgment and restart persistence. Simulate a write failure and show
   that the old accepted state remains. Selection changes must not apply a
   delayed result or a draft to the wrong Agent.
5. Exercise Pause/Resume, Interrupt, opt-in speech and theme controls. In
   particular, type while pause-on-input is active, press Pause explicitly, and
   send: releasing the automatic typing pause must not undo the manual pause.
   Also test the reverse order: press Pause, wait for its acknowledged state,
   type a draft, then clear it or submit it. Observe outgoing controls as well as
   the real actor's state: the browser must not send an automatic unpause that
   releases the deliberate pause. A simple Pause/Resume click pair misses this
   interaction. Keep automatic-pause ownership distinct from a manual request,
   without creating a browser-owned copy of authoritative actor state.
   A tab may remember that it requested an automatic pause; that is local control
   intent, not authority over an already paused Agent. Typing after another tab's
   Pause must not claim that pause. Explicit Resume must work with an unchanged
   draft, and explicit Pause must be reachable while typing has automatically
   paused the Agent. Exercise selection, reconnect, and interruption: stale local
   intent must neither release someone else's pause nor travel to another Agent.
   Wait for actual ready phases: a seeded conversation can be `inputpending`,
   while a completed turn can be `idle`. A guessed state predicate is a fixture
   failure, not evidence that Resume failed.
   Keep model changes disabled during admitted work even when later usage/meter
   frames arrive; a status refresh must not accidentally re-enable the control.
   Queued speech counts as busy according to the existing policy; selection cancels
   stale callbacks and replay is silent. A fake speech boundary verifies wiring,
   not hearing. Native voice quality, useful speed, and accessibility need the
   human's own browser and ears before acceptance.
6. Stop and restart the owned host with the same saved trial state. Continue the
   conversation and perform a new tool operation. New content must not overwrite
   an old card; identity must survive replay and restoration. Confirm graceful
   shutdown joins work and saves state, including when the browser disconnects.
7. Exercise failure and recovery: occupied port, refused settings, interrupted
   provider work, stale browser selection, and the independent rollback cases.
   Keep demonstrated failures and corrections; do not count a timeout as proof
   that a refusal worked.

These are integrated exercises, not seven new frameworks or a fabricated point
score. Reuse the existing test programs for affected contracts. Run broader
regression when shared behavior changes. Read saved source and complete relevant
spans; build success cannot excuse broken ownership or a barely usable page.

Do not borrow live credentials for the student run. The author and human arrange
any later bounded live check explicitly. Public OAuth and Codex are separate
funding/credential paths with no silent API-key fallback. The prior C# Codex cache
sample had two hits and one miss; it does not promise savings for the candidate.
No new cache campaign is required simply because the GUI changed.

## Lessons that remain obligations

The earlier Go crossover record is history, not a list of patches to transplant.
Several of its lessons belong in the replacement exercise:

- **Explain before effects.** Ordinary visible assistant text explains intent;
  private reasoning, a spinner, and a tool result do not replace it. Keep that
  explanation visible beside Actions. The existing enforcement checks arrival
  order, not just a normalized final message; do not weaken it for the new UI.
- **Observation serves work.** Use explicit GUI/speech observation when needed.
  Do not add automatic snapshot polling before every model request as a layout
  shortcut or turn diagnostic noise into spoken dialogue.
- **A check-in is not a deadline.** A Job observation delay does not end the Job.
  Do not impose a short whole-response HTTP timeout or automatic paid retries to
  make a slow provider look responsive. Controls and shutdown must still work.
- **Roles have different result contracts.** A main conversation ends with a
  normal reply; a delegated worker or compressor may owe a structured result.
  Retain source through failed compression. Pleasant prose is not a committed
  replacement, and a background worker is not permission to publish History.
- **Configuration needs honest authority.** Show current versus next-launch
  state, save before success, and keep credentials out of settings and History.
  Do not import the earlier Go configuration schema over the C# ownership model.
- **Build, install, restart, refresh, observe are different acts.** A new binary
  on disk does not change a running process or an already loaded browser tab.
  Record which act actually happened. Neither building nor a timer grants restart
  authority over the human's working application.

These principles explain why green grades alone did not finish the crossover.
They also keep a redesign from regressing less visible work that already matters.

## Where to work

Read the actual declarations and callers, not just this source map:

| Concern | Starting points under `solutions/edition-3/main` |
|---|---|
| Browser layout and behavior | `Gui/wwwroot/index.html`, `style.css`, `app.js`, `peers.js`, `auth.js`, `speech.js`, `mcp.js` |
| Optional GUI library and hosting contract | `Gui/README.md`, `Gui/Ensemble.Gui.csproj`, `Gui/Server/` |
| Actual GUI composition and startup errors | `Examples/Gui/Program.cs`, `Examples/Gui/Gui.csproj` |
| Ownership and public operations | `Common/`, with corresponding implementation directories |
| Actor/replay/settings tests | `Tests/Program.cs` and its referenced real-component test groups |
| External GUI consumer | `GuiTests/GuiTests.csproj` |
| Build/run orientation | `README.md` |

Preserved numbered endpoints are read-only historical boundaries. The working
solution is the candidate. Do not edit an earlier endpoint to make a regression
disappear. Do not change the public protocol or persistence schema merely to
simplify styling; a genuinely needed change requires architectural agreement.

## The fresh-student loop

The author first commits this chapter and its agreed starting implementation.
A fresh student receives that revision, the current architectural rules, and a
clean isolated working tree. The student owns implementation, tests, and the
local exercise through completion, and may ask about genuine ambiguity. Do not
interrupt it for routine early compiling checkpoints or supply an unrecorded
solution through hints. The author performs the other roles: engineering review,
behavioral verification, and teaching revision. The human evaluates the candidate.

The submission states changed files, how to build and start, the actual local
results, known limitations, and any clarification received. Keep it concise;
use the repository's normal records rather than inventing an audit bureaucracy.
Do not call the whole system accepted because the submission was accepted by a
tool. A fresh context, different model name, or isolated directory alone is not
proof that the teaching is sufficient; the work and recorded assistance matter.

The Bootstrap decision and the first run's substantive behavior/layout lessons
justify a new fresh-student run. Retain the first candidate and its failures for
comparison, but do not merge its solution into the next starting implementation
and call that a rebuild from the chapter. The next student receives the completed
C# foundation plus this revised assignment. Small corrections can still be made
within a run; another reset for every CSS adjustment would not teach us more.

When enough new requirements or failure lessons accumulate, preserve what was
learned, revise the chapter, commit the new teaching baseline, and run a fresh
student again. Discard only the owned student run after its work is no longer
needed. Never `reset --hard` or `clean` the human's shared working repository to
simulate a fresh student. Minor corrections may be inspected directly without
pretending each one has earned a new independent run.

## The human acceptance and promotion gate

The student hands over a candidate that can be run beside the old Ensemble.
The human then checks actual readability, speech, keyboard use, normal coding
work, interruption, settings, agent management, restart continuity, and recovery.
Diagnose problems kindly and concretely; do not blame the human for exposing a
path the tests missed. The chapter improves when those observations become clear
teaching and meaningful regression checks.

When the human says the new Ensemble is ready to replace the current one, the
crossover is complete. Only then archive the current Edition 2 implementation
and book and promote Edition 3 to the top level. Inventory the actual repository
layout first, preserve history and the known-good launcher/data, stage the move,
fix current links and launch paths, and verify the promoted build and rollback
before retiring the old launch route. Do not overwrite the Edition 1 archive,
move private state into Git, delete the old installation, or interpret acceptance
as permission to publish or push. Execute the actual switch with the human.

The code is rebuildable, not disposable. The durable artifact includes the
judgment needed to rebuild it—and the judgment to know when it is ready to use.
