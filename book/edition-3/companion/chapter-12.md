# Chapter 12: The Agent Sees Its Working Interface

An Agent can know its own environment only to the extent that the environment makes facts available. This chapter gives that idea practical machinery: recorded manuals, truthful startup gates, and a browser bridge that can inspect the page through the same MCP transport already carrying controls. There is no additional faculty of introspection. There are instructions, capabilities, and observations with identifiable sources.

That distinction is especially useful when the Agent tests its own interface. Source code says what the page was intended to do. A browser observation says something about what the running page currently does. Neither can substitute for the other.

## A manual is part of the application

Every built-in tool has a contract beyond its schema. The Agent needs to understand the difference between accepted work and completed work, observation timeouts and cancellation, output limits and full results, and the handles needed to continue. A skill manual teaches those habits while granting the matching tools.

The shipped `ensemble` primary establishes ordinary operation. Other manuals add the browser bridge and GUI debugging behavior. The host must actually ship and resolve the directory containing them; a README mentioning a skill is not availability. CLI and GUI receive an explicit skill directory and primary selection, with compiled-in defaults resolving the shipped tree. Core receives values rather than discovering the environment for itself.

Manual placement is equally substantive. The primary body renders once at creation, with loadable manuals later entering provider instruction context exactly once. A friendly load-result message alone would leave the manual in ordinary tool prose, where it has the wrong role and can disappear with result retention. The recorded rendered text explains what was taught even after a source file changes.

Templates need to advertise usable capabilities. A parser that cannot resolve the forms in the shipped manuals can accept a skill file while quietly breaking its teaching function. The current contract supports both bare and braced references, including the shipped `${TOOLS}` form, and leaves unknown names literal. This is about making the actual manual work, not adding a general templating language.

The embedding host can also supply an ordered initial skill set. Primary and initial skills are validated as one transaction before load events or grants are published. The CLI defaults to the primary and programming skill; an explicit CLI list replaces its default additions. The library does not silently inherit those CLI defaults. One invalid entry must not leave the earlier entries active.

Rendering uses the prospective capability set, including the tools being registered by the same skill. Otherwise a successfully loaded manual can describe yesterday's tool surface. The model-facing loader accepts a skill name, not new directories, grants, or remote registrations disguised as extra arguments.

## Readiness is observable behavior

A wholly local initial skill closure retains synchronous initialization. Any remote dependency in the primary plus ordered initial skills requires asynchronous startup after persistence restoration and diagnostic setup. Until its transaction commits, the manager is unready. User input and direct tool calls are refused while controls remain responsive; the application must not advertise an Agent that has only half its required environment.

Any browser requirement in that complete initial closure requires the GUI host to start listening before the connection can be claimed. The complete closure determines that wait, not the primary alone. A set whose remote dependencies are all stdio-based needs no browser, and a CLI that cannot provide a required browser transport fails instead of waiting indefinitely. Startup cancellation or failure closes owned work without overwriting an earlier good save.

GUI text arriving during startup remains in a bounded host queue and enters the Agent in order only after readiness. New arrivals cannot overtake the backlog during drain. MCP replies and control traffic continue to flow while that queue waits. Otherwise the very browser messages needed to finish initialization could get trapped behind the readiness gate.

Later disconnect is a capability loss, not permission to operate on stale state. A running ordinary-primary Agent can lose the browser’s Tools while retaining a manual pending removal. A newly connected browser supplies a new registration; the old session is not revived by changing a socket reference. Failure during initial-set activation is stricter: the required startup has not succeeded.

## Looking through the bridge

The browser implements an MCP peer over the existing `/ws` connection. A single message path separates GUI envelopes from MCP envelopes, and replies remain associated with the connection that received each request. The request’s Agent identity must match the server-provided identity. A second tab can observe without impersonating the active tool-serving tab.

The shipped browser offers visible-DOM inspection, click, text input, speech-queue inspection, and speech enqueueing. The `gui-debug` manual grants all five through its owned browser connection. Each tool still has one owner; depending on a capability does not justify constructing another copy of its handler.

Snapshots are bounded descriptions of visible DOM content with usable selectors and disclosed truncation. Click requires one visible match and cannot quietly navigate to a different origin. Input requires one visible editable text field; quotes, Unicode, spaces, and a string such as `{Enter}` stay literal text rather than becoming keyboard macros. Hidden, read-only, ambiguous, or unsuitable controls are refused.

A requested action is not proof of its result. A follow-up snapshot establishes the resulting field value or visible page state. Speech-queue state establishes what the page has queued, not audible quality or human hearing. The bridge does not attach a snapshot to every provider request merely because the browser is available. These observations are requested on demand. Inspection omits passwords, but choosing to expose the surrounding page still requires care.

The `gui-debug` skill state owns its browser connection. Its complete surface includes DOM snapshots, clicks, text input, inspection of the spoken-audio queue, and requesting speech. Unloading must release the connection as well as remove those capabilities. Browser availability alone is not a reason to activate them.

## The input field is not the cause

The retained first-use account describes a particularly instructive freeze. The GUI paused its coding Agent while a human typed. The browser driver filled the prompt programmatically, producing an input event that looked like typing. Submission then cleared the field by assigning its value, which produces no input event. The pause transition happened; the corresponding unpause path never did.

The first repair distinguished programmatic input and explicitly cleared pause on submission. A later native-browser check found another path to the same mistake. With speech enabled, an utterance ended while programmatically inserted text remained in the prompt. The speech callback looked at the nonempty field, reconstructed “human typing,” and held the Agent paused again. The input listener had remembered origin for one event, but the later reader had forgotten it.

The current bridge marks MCP-origin input. The prompt listener retains a separate human-typing fact, and speech transitions combine that fact with the current speech cause instead of inferring intent from text length. Successful submission clears both the field and the human cause. Genuine human whitespace still counts as typing; programmatic text does not acquire human provenance merely because time or a speech callback has passed.

This also retains the earlier Agent-wide control rule. Another tab can explicitly unpause; an unchanged local cause must not immediately reassert a veto. A later real transition can send another control. The cause is retained for correct scheduling, not converted into permanent ownership of the Agent.

Origin tracking does not authorize overwriting a human’s unfinished draft. The manual requires preserving that input unless replacement is permitted. Correct event plumbing solves a scheduling defect; permission remains a separate question.

## What the historical encounter established

The retained historical experiment used a separate GUI-only driver without filesystem or command tools. That restriction made the screen its evidence. Given source access, it could have answered interface questions by reading the intended implementation and missed the running failure entirely.

The driver also exposed an interactive-debugger deadlock. The command handler waited in a synchronous PTY read for process output to close. A debugger waiting at its prompt would never close it, and the observation routine that could recognize the prompt ran only after the handler returned. The escape hatch sat after the operation it needed to interrupt. The repair gave the running job and its reader a lifetime independent of the initial observation, matching the job-versus-wait distinction already central to this architecture.

These were earlier encounters, not fresh live-provider runs of the C# endpoint. Bill Cox supervised them, and the record distinguishes operational guidance from findings: a stale server holding the port needed human direction; truncation was noticed by the human; an unsafe rendering path emerged in subsequent source inspection. Crediting all of these to unaided model observation would erase the very information needed to assess the observer.

The speech-disabled input check also proved less than it first appeared to prove. Native speech start and end events exposed the second pause path. Those observations establish measured scheduling transitions, not subjective listening quality, and an earlier unexplained two-tab speech stall remains unexplained. Self-knowledge improves when these limits stay attached to the result. The valuable loop is concrete: inspect the running system, act through a declared capability, observe the consequence, and keep the cause and evidence intact.

Source: [Chapter 12](../chapter-12.md).
