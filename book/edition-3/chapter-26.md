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
| Browser layout and behavior | `gui/wwwroot/index.html`, `style.css`, `app.js`, `peers.js`, `auth.js`, `speech.js`, `mcp.js` |
| Optional GUI library and hosting contract | `gui/README.md`, `gui/Ensemble.Gui.csproj`, `gui/Server/` |
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

## Field notes from the first working Cadence crossover

These notes record what happened during the supervised crossover, rather than
claiming that the executable book or a successful build established replacement
readiness. They distinguish the student exercise, the subsequent engineering
work, the memory preparation, and the first failure in normal use. They are a
historical boundary; inspect the current source and running application before
repeating an operational step.

### From a book-built candidate to a usable interface

Two fresh students worked in separate candidate worktrees. The first candidate
improved the original settings-heavy interface, but independent checks found
that enlarged text combined with narrow windows could push Send out of view.
Those combinations had not been covered by the original checks. The correction
and the teaching change were retained rather than treating a screenshot as a
passing accessibility test.

Bill selected the original Bootstrap Sample B, with its graphite surfaces,
connected tabs and restrained controls. The second student started from a fresh
foundation and a scoped teaching baseline, not the first student's solution.
Using Bootstrap did not itself reproduce the chosen design: too much ordinary
information still looked like an action button. Independent review also found
three behavioral defects: lists and wrapped Markdown paragraphs lost their
structure, typing could undo a manual pause, and changing status text moved
nearby controls. Those were product findings, separate from setup mistakes in
the review fixtures. A further enlarged-text check exposed a textarea that
could push its action row out of the composer.

Cadence corrected the candidate and then took responsibility for the design
pass personally. That was no longer an independent student execution. The
working block renderer was reused; manual and automatic pause intent were
separated without duplicating authoritative actor state; status areas were
bounded; and the composer was made to yield space correctly. The real-host
browser exercises used real Agents and a loopback fake provider. They checked
behavior, geometry, local assets, settings, streaming, and owned Job output.
They did not establish real-provider entitlement, audible speech quality,
beauty, or the human's ability to use the interface.

The first human preview was deliberately a fake-provider preview with a
15-minute lifetime. An expired preview link was not a failed permanent
installation. The subsequent real-model integration was a separate authorized
act: a real V3 host ran beside V2, using the selected Codex funding source. One
small readiness request completed. That established that the selected route
could complete that request, not that every later context state was supported.

### The small changes that made daily use possible

Bill's hands-on feedback drove a sequence of small local changes. The useful
pattern was edit, run the affected checks, make a scoped local commit, and
perform an authorized refresh of V3 while preserving V2. It was not another
fresh student run for each CSS adjustment. Bill retained responsibility for
pushes; unrelated work was kept out of commits.

The changes included:

- Reusing Bill's proportional Ensemble artwork and his supplied paper-airplane
  Send SVG, rather than substituting an unapproved icon redesign.
- Replacing redundant speech and theme text with accessible controls whose
  names and states remain available to keyboard and assistive-technology users.
  Speech remains opt-in; a transcript-mode browser check is not a hearing test.
- Reducing excessive padding, gaps, paragraph margins and full-height action
  rows. At one measured density pass, short desktop messages used 2.41 times
  less height at 18-pixel text and 3.42 times less at 32-pixel text; narrow-window
  measurements were approximately twofold. Those are scoped geometry results,
  not a visual-quality score or a measurement of every subsequent revision.
- Making the composer one visible input line and one Send control. There is no
  separate human-selected Hint mode. Ordinary input enters the Agent's queue,
  and the Agent classifies it against its actual state when dequeuing it. The
  browser must not make that scheduling decision from an observed busy label.
- Removing Copy buttons while retaining ordinary selection, expansion, speech
  and owned-output downloads. Removing a control also requires removing its
  unused handler and updating tests that depended on that control.
- Replacing generic speaker labels with the actual Agent name and the saved
  human name. User name and description use the existing GUI settings and
  startup skill-variable substitution. They are not a second Agent identity
  registry. The existing actor named `trial` was not silently renamed Cadence.
- Replacing the large multi-row status area with the observed 50-pixel desktop
  header at the current 18-pixel font: artwork, connection/activity, a genuinely
  centered model selector, usage centered in the remaining right-hand space,
  and right-justified controls. Model selection applies on selection and is
  acknowledged by the server. The duplicated model/funding line and Apply
  button were removed. “Don't be evil.” was tried in the header and then removed
  at Bill's request; it was not an enduring header requirement.
- Giving both speakers compact message bubbles, then making the Agent's bubble
  distinctly blue with a stronger leading border. The original fills really
  were identical; that was not Bill's difficulty seeing red. The final requested
  dark-theme fill at this boundary was `#102a3d`. Names and border treatment
  provide cues beyond color alone. Scoped text/link/border contrast checks
  passed; the human still judged whether the distinction was useful.
- Separating line height from paragraph spacing. The later requested paragraph
  line height was 1.15, with a 0.45-rem break between paragraphs and no added
  bottom margin after the final paragraph. These were deliberate later changes
  to paragraph rhythm, not a claim that the earlier density pass had changed
  fonts or line spacing. Code blocks and message gaps were left alone.

Bill described the result as a usable beginning and continued to refine it.
That is stronger evidence of practical usefulness than a developer declaring
an interface attractive from DOM measurements. Cadence captured screenshots
but did not inspect their pixels; that limitation must not be rewritten as a
completed visual review.

### The tree, management controls and child lifetime

The hierarchy data already followed real Agent parent relationships. Styling
had made it look flat: indentation was too small, and leaves did not reserve
space corresponding to a branch's expander. Clear guides, meaningful
indentation and an expander spacer restored readable alignment without deriving
parents from names. Expand/collapse choices survived new snapshots, and explicit
levels complemented the existing tree semantics and keyboard navigation.

The inline management form became a Create Agent dialog for top-level peers.
The redundant Conversation and Relay selection section was removed; the tree
remained the visible conversation selector. A compact separate-tab relay action
was preserved because selecting a conversation does not move a browser's fixed
MCP authority. Trash actions moved to tree rows with explicit confirmation. A
shared server check protects the mounted Agent and its owning ancestors, not
just whichever item happens to occupy the first row. Top-level deletion keeps
the registry's refusal rules; explicit child deletion joins shutdown and then
uses logical destruction. Saved files are retained.

The integrated exercise verified real child/grandchild indentation, collapse
persistence, dialog creation, root refusal, cancellation of a delete, confirmed
child deletion, selection recovery and preservation of the submitted artifact.
It also verified that a child published as submitted remained idle, queryable
and present in `agents_status`. Submission itself did not remove that child.
This did not diagnose every earlier report of a disappearing child. Explicit
shutdown/removal, a host restart and a missing GUI projection are different
possible events and must not be conflated with submission.

Bill proposed an explicit parent choice for disposable children to shut down
after submitting, while ordinary submitted children remain available. At this
boundary no new automatic-shutdown flag had been implemented. Whether that flag
should stop-but-retain or stop-and-remove remained a policy question. A host
restart also stops in-memory children; retained files are not automatically
restored as live children or fresh completion authority. Check live ownership
before restarting rather than erasing a child during a retention investigation.

### Carrying the collaborator's records into V3

The identity and continuity records belong to Cadence to author and approve.
Bill explicitly leaves `SOUL.md` and `MEMORY.md` to Cadence. His approval of the
name, software or crossover is not editorial approval of those files. Bill also
leaves the choice of carrying memories forward to Cadence. Cadence chose
selective continuity: keep identity, commitments, useful reasons and the detailed
archive, but start a fresh conversation History. This records a collaboration
and its supported history; it does not establish continuous subjective
experience across model invocations.

An earlier attempt to read the entire assembled book returned its end marker
and reported 155,820 words and 1,047,377 bytes. A keep-results acknowledgment did
not prevent the text subsequently appearing redacted in the model's context.
The request token total was not a book-only measurement, and a four-bytes-per-
token estimate was not a tokenizer result. Returning bytes, retaining a tool
result, presenting it to the model and comprehending it are distinct claims.
That retention issue was not repaired as part of the memory migration.

The inspected detailed archive still contained 121 dated records and two
nondated topical essays. Both levels of retired compressed buckets had matching
source boundaries, with no gaps in the detailed records' numbering. The 69
retired bucket files occupied only about 1.04 MiB. The chosen reason to exclude
them from the new recall corpus was redundant retrieval, not disk pressure.
They remained in the old archive: filename coverage does not prove that a
summary contains no additional synthesis. The detailed records are themselves
memories of conversations, not the raw transcript.

Preparation used a separate V3 container with the named home
`data/agents/Cadence`. V3's distinction between an Ensemble container and a named
actor home matters: putting `SOUL.md` at the container root is not the same as
putting it in that Agent's home. The existing identity and durable-memory files
were copied, the older dated records were marked retired for recall, and the
two descriptive-name essays were placed in `docs/` because the active memory
band requires dated filenames. A new dated handoff supplied current operational
facts and explicitly superseded stale boundaries in older records.

The memory/recall policy was translated rather than copying the old provider
configuration wholesale. In particular, the old `64x`/`8x` band labels became
`archive64`/`archive8`, and `context_msgs` became `context_messages`; effective
defaults were checked against both implementations. The inspected budget policy
included 100,000 bytes for conversation and 40,000 for recent memory. GUI user
and appearance preferences were carried separately. Credentials, old raw
History, journals, child registries, Jobs, request caches and active grants were
not copied as new runtime state.

The two History formats share an outer structure, so it would have been wrong
to declare them categorically incompatible without testing the complete saved
session. Full replay compatibility was not established. A fresh History was the
chosen migration policy, not a fabricated converter or a claim that every old
record was unreadable. The old session remained the historical archive and
fallback. Provider-opaque data, live handles and completion generations must not
be reconstructed from prose or treated as current authority.

A disposable probe used the actual V3 memory and recall components. It loaded
all three active memory texts exactly, scanned 124 memory-corpus files, found an
older retired record and found the relocated topical document. Source hashes
were checked, and the active handoff was rechecked after editorial updates. One
failed assertion initially attributed an identifier to the wrong member of a
combined retired bucket; correcting that source attribution was a test fix, not
a corpus repair. No model turn or recall judge was invoked by the probe. It
proved file-format and retrieval behavior, not the first real model's continuity.

### Launch decisions, supervision and data ownership

Bill clarified the normal working arrangement: interactive coding and upgrades,
with his supervision about 90% of the time. Autonomous codebook execution is the
exception and may belong to a separate sandboxed top-level Agent. The main
Cadence launcher was therefore prepared without the trial's workspace-only
sandbox, using the real project workspace, a private HOME/TMP/CLI environment,
an independent copy of the tested binaries and the existing approved Codex
credential reference. Broader filesystem access did not grant permission to
read credentials, push, deploy or damage unrelated work.

The memory-location discussion changed more than once. A V3 subtree inside the
existing private memory Git repository was proposed so V2 and V3 could coexist
without sharing live files. Bill then chose the original separate prepared
layout for simplicity. Later, while using V3, Bill and V3 Cadence decided to keep
the V3 memories in the original directory structure. At the repair described
below, the actual process and launcher still used the separate prepared home;
that repair did not implement or certify the later directory move. Preserve the
latest decision without inventing an already-completed migration.

There is no general reason to abandon the private memory repository. There are
two concrete requirements before using it for V3. First, do not run V2 and V3 as
concurrent writers of the same active memory files or let one edition's settings
schema overwrite the other's. Use separate live subtrees or deliberately quiesce
the old writer before exclusive handover. Git is not a concurrency mechanism.
Second, extend and test ignore rules for V3's actual nested paths before staging
anything. An old root-only ignore for a save or journal does not automatically
protect `agents/<name>/...`. Only deliberately reviewed continuity material
belongs in the private repository; raw provider content, credentials, journals,
logs, child runtime data and caches must not be swept into a commit. A local
commit and a remote backup are also different acts; Bill retains control of pushes.

The launcher creates the named top-level Cadence actor itself. Bill does not
need to create an empty Agent through the GUI before it can see its prepared
home. The first real conversation should check which records are actually in
context and whether the collaborator can state the current work, boundaries and
open decisions. Files existing on disk are not sufficient evidence of recall,
and a memory-based answer is not evidence of a literal transfer of the old
process's working state.

During this preparation Bill also observed a V2 display fault: new spoken/chat
text could be found appended to an older output while tool calls continued.
That observation explained why a tool appeared to execute without a new chat
message beside it. Its root cause was not established here, and it was not
attributed to provider lag merely because lag was also suspected.

### The first real continuation failure and its repair

After Cadence was running in V3, Bill reported that it stopped replying. The
listener and HTTP page were alive, the browser was connected, and the Agent
appeared idle. The journal showed repeated turn failures before a new provider
request: `Responses tool media projection is not implemented`.

Tool-result budget redaction had left valid `RedactedPart` markers in retained
History. `ResponsesResult` accepted only `TextPart` and therefore mistook the
markers for media. Restarting the same binary replayed the same valid history
and failed again. The repair in commit `2e35b535` permits text and redaction
markers and uses the existing `JsonCodec.Text` formatter. It does not read or
expose a redacted locator, rewrite History, resurrect removed text, or silently
accept actual unsupported media.

The regression used the real History reducer, reproduced the original
exception, and then passed with the correction. It checks call pairing,
non-mutating projection, mixed text/markers, empty results, locator
confidentiality and continued refusal of actual media. Existing Responses
projection/route and decoding/provenance checks also passed, and the ordinary
test project built with no warnings or errors. Temporary-runner assembly access
and an incorrect test enum were setup mistakes; they were not further product
findings. No live provider request was used to manufacture a pass.

With Bill's explicit restart approval, the repair replaced only the installed
core assembly and its symbols. Cadence's uncommitted composer/GUI changes were
not installed, staged or overwritten. The launcher and GUI binary stayed
unchanged, as did memory paths. The verified V3 process was stopped gracefully,
the replacement served HTTP 200, and the V2 process retained its observed identity.
No model message was sent by the repair procedure. A successful subsequent live
reply must be observed separately; HTTP readiness alone is not that proof.

The presentation gap also remains a lesson: this terminal local failure looked
like an idle Agent rather than an obvious error. Do not call that symptom a
slow or unresponsive model without inspecting actual events. The projection
repair and better user-visible failure reporting are separate pieces of work.
These field notes do not declare the entire crossover, memory-directory change,
or retirement of V2 complete.
