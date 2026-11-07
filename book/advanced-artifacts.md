# Advanced Artifacts

> **Initial design packet.** TL;DR, plain-language explanation and exercises for
> discussion and implementation with Bill. This is an unnumbered working chapter,
> outside the release assembly. Directory and file rendering, shared helpers,
> and renderer trial/recovery have bounded implementation evidence; remaining
> work is tracked in the design notes. The detailed chapter body will follow
> implemented, observed behavior.

> **Body-writing plan: figures.** Bill requested images of the Artifacts we
> create. Use actual implemented views with synthetic example data, after capture
> approval, not mockups or private conversations. Candidate figures are directory
> and grouped search output, an applied replacement, one continuing command
> session, inline dice disclosure, and Try live/Keep/Revert controls. Select only
> figures that teach a distinct point. Give each meaningful alt text and a caption;
> retain a complete prose explanation and avoid color-only distinctions. Figures
> are planned, not captured or visually approved.

## TL;DR

Let a user describe how tool output should look, then let the Agent find and adapt
its renderer. Start with `list_directory`. Extend the same mechanism to the other
five original tools and to tools supplied through skill MCP connections.

### Package presentation with the tool

Move JavaScript Artifact renderers into the skills directory structure. A tool's
renderer belongs alongside the skill that supplies it. Give chat, readable
thinking and the fallback renderer appropriate skill-package homes too. Renderers
are shared across Agents for now; they are associated with tool or content types.

Keep shared libraries in or adjacent to that structure. The layout must support
reusable JavaScript, Python, Go and other tool implementations without requiring
a universal cross-language runtime. Ship readable source, concise API guidance
and examples so an installed Ensemble can be customized without a source checkout.

Expose a small discovery index through the skill system. Load relevant helper
instructions into model context when needed; read implementation source when
changing or debugging it. Browser library loading and model-context loading are
separate operations. Reuse the safe Markdown renderer rather than duplicate its
parser in every skill.

### Preserve facts while changing their presentation

An Artifact has stable identity, original content and lifecycle. Its renderer
chooses its visual form; the host determines where that form appears. A common
card border, heading or Tools pane is optional. An inline dice result in a
single-pane chat application is a valid Artifact.

Retain call/result relationships, Agent ownership and recorded Job observations.
Distinguish partially generated arguments, admitted execution, returned
observations and completed work. Presentation changes must preserve the recorded
content and must not silently alter model-facing results. Use a readable default
when a renderer is absent, unapproved or unable to render its input. The default
is itself inspectable and customizable.

### Give each tool a useful display

- **`list_directory`:** a compact listing that distinguishes files and directories,
  preserves returned names, and represents only the entries actually listed.
- **`search_files`:** matches grouped by file, actual line numbers, highlighted
  matches and returned surrounding context. Identify omitted or truncated output.
- **`read_file`:** readable source with actual file line numbers and suffix-based
  syntax highlighting where supported. Preserve the starting line of a range.
- **`write_file`:** a formatted view of the supplied content, with resulting file
  line numbers after success. Append numbering accounts for the existing file,
  including an unterminated last line. Label proposed content separately from a
  confirmed write; a streamed argument is not evidence of a filesystem change.
- **`edit_file`:** a visual comparison of the actual applied replacement, with
  truthful positions and extent. An exact replacement is not automatically a
  whole-file diff. Share code-highlighting helpers where practical.
- **`run_command`:** ANSI-formatted output resembling a dark-theme terminal.
  Select an existing library appropriate to output display. Define the supported
  control sequences; output must not gain clipboard or application authority
  merely by containing an escape sequence.

Unknown file suffixes remain readable. Line numbers and change positions must
come from the actual operation and its content, rather than a later reread of a
possibly changed file. Preserve existing file-access and exact-edit semantics.

### Continue one Job's presentation

A Job retains one continuing Artifact. `wait_for_job` and `send_input` foreground
that Artifact and move it to the bottom of the Tools pane. Ordinary background
output does not continually reorder the pane. Preserve focus, selection and
internal scroll state; a reader inspecting earlier output can stop following
the tail.

These controls retain the originating tool's output format for both the model
and the GUI. Correlate by actual Agent/Job identity. Keep individual calls and
observations recorded while combining their presentation. Reconnect reconstructs
the continuing view without rerunning a command or duplicating consumed output.
An observation can finish while its Job remains running.

The scrolling command/Job display retains the **latest 64 KiB** of output. Clearly
indicate when older output has rolled out; retained output remains on disk under
its existing ownership and retrieval rules. This display budget is independent
of model-facing observation limits. Feed the GUI without consuming the model's
output cursor or raising its context budget. Preserve valid text and ANSI parser
state across chunk and window boundaries.

### State the trust boundary

This chapter first supports trusted, approved renderers running in the page.
Bill has chosen that model for supervised development. Such code has the page's
privileges; an Agent's process sandbox does not also contain its browser code.
Keep unapproved downloaded renderers inactive and display the safe fallback.
Trusted renderers still treat filenames, fetched documents and tool results as
untrusted data. Preserve the shared Markdown helper's safe handling of markup,
links and remote images.

Iframe isolation for untrusted renderers, an installer, workflow orchestration,
and Python-computed skill variables are outside this increment. The shared
library layout should accommodate later workflow executables without building
that chapter here. A workflow will use reverse MCP to orchestrate sub-agents of
its calling Agent.

### Decisions to settle in the first experiment

The behavior above is agreed direction. The following interfaces remain open:
renderer declarations and invocation/lifecycle signatures; shared-library paths
and imports; approval and renderer-update handling; the typed metadata needed
for file positions; and the bounded Job-output feed and replay protocol. Select
these with Bill, record them here before they become grading requirements, and
use existing owners rather than add a second conversation store. ANSI, diff and
highlighting libraries have not yet been selected.

## In plain words

A tool result is part of the conversation between a person and an Agent. Its
presentation determines how much work the person must do to understand what
happened. A directory listing should look like a directory listing. A change to
a source file should make the changed lines easy to find. A debugger session
should remain a session while the Agent asks it another question.

Consider a user asking for directories to stand apart from files. Today, changing
that display can require finding a special case inside the GUI. With a
skill-owned renderer, the Agent starts from the skill that provides the listing,
reads the renderer and its interface, and changes that presentation. A second
Agent using the same tool gets the same improved display. The original directory
result remains available even if the custom renderer fails.

The small example establishes the extension mechanism. It should also establish
how the human can understand the change: which renderer was selected, which data
it received, and what the new code does with it. The Agent can explain those facts
while the user still has a chance to steer. A natural-language formatting request
is a programming request with a human reviewing the result.

### Reuse a helper without reading a library

A renderer that needs Markdown should call the existing safe helper. The Agent
needs to know the helper's inputs, outputs and important restrictions. It usually
has no reason to read the parser's implementation. A short library index points
to that information; detailed source is available when the task needs it.

The browser can already have the Markdown parser loaded while the model knows
only where to find its API documentation. That separation keeps useful code
available without spending context on every library in the installation. The
same packaging idea serves a Python helper used by several tools or a Go package
used by workflow executables. Each language keeps its ordinary execution model.

### A dice total is enough of an Artifact

Bill's Homebrew-VTT example makes the visual requirement concrete. The application
has a chat pane, and a dice tool's result belongs in that conversation. For
example, a result could appear as “Attack: 23.” Clicking or focusing it reveals
“d20: 17 + proficiency: 3 + ability: 3.” Hover can offer the same explanation,
with keyboard and touch access available too.

That display needs the recorded calculation and a small interaction. It does not
need to resemble a source-code card. Revealing the explanation also must not roll
again. The Artifact's identity and content let the renderer explain the event;
the chosen shape makes it useful in that application. Shared UI helpers should
make ordinary displays easy without imposing a universal card template.

### Foreground the work that already exists

Suppose the Agent starts a debugger, receives a prompt, then turns to a source
file. A later `send_input` addresses that same running process. The human should
see the debugger Artifact return to the bottom of the Tools pane, accept the
input and continue displaying output. Starting a new visual snapshot for every
interaction forces the reader to assemble the session mentally.

History still records the individual operations. Sending input, waiting for
output and receiving a result are different facts with their own identities.
The continuing Artifact is a view assembled from those facts and the Job's
retained output. It does not own the process or turn an observation into proof
that the process exited. Reopening the browser must preserve that distinction.

A busy process can produce more output than either the reader or the model needs
at once. The human gets a scrolling window of the latest 64 KiB, with a visible
indication of older retained output. The model gets the bounded observation it
requested. Reading more in the browser should not consume the next model
observation or make every model request carry the display buffer.

File operations need a similar respect for what happened. Appending text to an
existing file does not make its first displayed line line 1. If the previous
last line had no newline, the append may continue that line. The renderer needs
facts captured by the operation; reading the file later could describe another
edit entirely. Attractive line numbers are useful only when they identify the
right lines.

### Choose trust deliberately

Bill supervises the code being written for his installation and has chosen to
run approved renderers directly in the page. This makes inline interaction and
ordinary DOM integration available without an iframe boundary. The choice also
means the renderer is trusted application code. Watching its development helps
review it; it does not technically prevent that code from accessing the page.

A downloaded skill needs an explicit trust decision before its JavaScript runs.
Until then, the host can show the ordinary readable result. Even approved code
must handle result text as data: a malicious filename should remain a filename.
An isolated renderer mode may be useful later, but that is a separate security
and integration feature. The first directory display should not require building
an untrusted-extension platform.

## Exercises

These exercises define the intended observable outcomes. They are a design draft,
not a claim of existing graders or completed implementation. Establish the open
interfaces with Bill during Exercise 1 before treating the packet as sufficient
for independent reconstruction. Use local fixtures and isolated application
instances; publication, live-service changes and paid inference are not required.

### 1. Make a directory listing useful

Implement the first skill-owned renderer for `list_directory`. Use a small
fixture containing files, directories and a name containing HTML-like characters.
Keep the displayed entries faithful to the tool result and distinguish directories
without relying on color alone. Change the format in response to a short human
request, then explain the files changed and show the result.

Demonstrate the readable fallback when no custom renderer is supplied and when
custom rendering fails. Record the chosen renderer declaration and lifecycle
interface in the TL;DR. Keep the existing tool's execution and access rules.

### 2. Package the renderers and discover their helpers

Move the existing Artifact rendering code into the agreed skill structure,
including chat/readable thinking and the default presentation. Factor common
Markdown and code-display helpers into the shared library area. Keep host
transport, authoritative identity and tool authority with their existing owners.

From a disposable distribution directory with no repository checkout, locate the
renderer source, discover a helper through its concise documentation, and use it
from two renderers. Verify that ordinary model requests do not acquire every
library's source code. Include a small inline dice-result fixture with an
explanation accessible by click, keyboard and touch. No Homebrew-VTT deployment
or React migration is required.

### 3. Show file contents and changes at their real positions

Add suffix-based code highlighting and line numbers for reads and writes, plus
a visual diff for successful exact edits. Use a file read starting partway down
the file, an ordinary write, an append after a trailing newline, and an append
to an unterminated line. Show the resulting line numbers and the actual edited
region. Keep a readable presentation for an unknown suffix.

Make a refused edit and a failed write visibly unsuccessful. Partially generated
write arguments may be displayed as proposed content, but must not be labeled as
a completed write. Show that the presentation preserves the original result and
that the diff's extent is labeled accurately. Explain the metadata captured by
the operation rather than infer historical positions from a later file read.

### 4. Group search results without losing their meaning

Render matches from several files with file grouping, line numbers, highlighted
matches and the returned context. Exercise adjacent matches and a bounded result.
Preserve returned content and make truncation explicit. Include markup-like text
in a match to check that display remains data. The renderer must not silently
rerun the search or open extra files merely to decorate the result.

### 5. Keep one command session in view

Use a controlled local process that emits colored output, accepts input and stays
alive across multiple observations. Render its supported ANSI output in one
continuing Artifact. Call `wait_for_job`, then `send_input`, and verify that the
same presentation is foregrounded at the bottom of the Tools pane. Background
output alone must not continually move it there. Avoid duplicating terminal echo.

Generate more than 64 KiB. Check the bounded tail display, valid text and styling
across chunk/window boundaries, an older-output indication, and retained output
on disk. Keep the model's smaller observation limit unchanged. While reading
older visible output, test selection, focus and scroll preservation. Reconnect
and verify the continuing view without restarting the process or duplicating
output. History must retain the distinct calls and observations, with final Job
state separate from each observation's completion.

### 6. Exercise trust and fallback on the actual page

Present an unapproved renderer and verify that it remains unexecuted while the
safe result is readable. Then deliberately approve a controlled fixture and
exercise its renderer. Treat filenames, search text, Markdown and command escape
sequences as data even in this trusted mode. Check that content alone cannot
request a tool action, remote image fetch or executable markup through the
provided safe helpers.

Record the trust boundary honestly: in-page approved JavaScript has application
privileges. These checks establish the selected helper and approval behavior,
not isolation from malicious approved code. Review the useful rendered examples
with the human before closing the increment. Do not expand this exercise into
iframe implementation or a universal extension-security certification.
