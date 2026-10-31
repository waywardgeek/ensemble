# Chapter 19: The Ensemble

Bill could listen to one Agent's reasoning at five times speaking rate and redirect it between tool calls. Several Agents are different. Their work needs named owners, private durable state, and a result the parent can actually collect. A fleet is not several independent chat windows with a shared optimistic ending.

This chapter opens two boundaries: an existing application can host the GUI around an Agent it owns, and an Agent can delegate work to children it owns. Both expose the same question. Does the public boundary carry a complete lifetime, or did a test quietly supply the missing pieces?

## Borrow the view, keep the Agent

The core assembly remains headless. CLI references it directly; the GUI adds ASP.NET and browser dependencies in a separate assembly. Common declarations still live with the core, and functional modules implement their matching partial members in that assembly. A public GUI package does not move implementation fragments of core partial types into another DLL.

A host passes an existing Agent to the public GUI entry point. It does not receive a hidden replacement with fresh History and unrelated tools. Browser prompts, approvals, settings, speech, and replies use the supplied Agent's actor and existing tool paths. The view can therefore mount under an application's route prefix without pretending to own the whole application.

The package embeds its real assets. A deployed program started in an unrelated directory does not need the repository's JavaScript tree nearby. An explicit disk override supports development without guessing from the current directory. An asset-only mount exposes the existing frontend modules; this chapter does not invent a new component API. Relative imports must work under the chosen mount instead of escaping to the application root.

The GUI borrows its Agent and, when supplied, its web host. Cancelling a view stops admission, resolves pending browser work, joins relays and subscriptions, and releases resources the view actually owns. It does not destroy the Agent or stop the caller's host. A standalone entry point owns more infrastructure, but borrowing an Agent still does not transfer its lifetime.

Replay reconstructs the view without starting inference or repeating speech. Serving a JavaScript file demonstrates deployment; rendering a usable interface is another observation. The current source includes outside consumers and isolated browser work, while leaving audible speech and human accessibility unclaimed by those checks.

## A child has two parents

Ensemble strongly owns every live Agent. A delegated Agent also has the Agent that spawned it as a parent. The first relationship covers global resources and identity; the second covers work, questions, and recursive cleanup. A constructor establishes both through the parents' relationship operations. Destruction removes both directions and recursively closes descendants, Jobs, pending questions, subscriptions, and MCP clients without destroying any multiply owned object twice.

C# garbage collection cannot perform that logical destruction. Nor does a callback that can manufacture an Agent replace the actual parent relationship. The current implementation constructs children directly from their owning Agents, using Common's declarations and the functional partial bodies.

Each child receives a named directory below its parent's state, including its own conversation and `submit.md`. Names are path components, not paths. Descendants cannot escape through a separator, and two live children cannot claim the same address. Shutting a child down removes its live identity but preserves the work on disk.

Isolation matters even when children share a task workspace for deliberate collaboration. Their conversation and submission files must not overwrite each other. The historical ownership audit changed a child's data directory to its parent's and broke several fleet behaviors at once. That was one serious shared-state fault with several symptoms, not several independent defenses.

A child inherits the parent's construction choices and enabled primary capability. Its task text grants nothing. A reviewer launched with a custom submission shape does not silently impose that shape on its ordinary grandchild. Persistent choices have owners; ad hoc delegation arguments are not a new ambient policy.

## “Done” is a durable event

A delegated Agent finishes by calling `submit`, not by producing a plausible final paragraph. The main human-facing Agent and named peers still finish in ordinary chat: they have no Agent parent waiting for a submitted result.

Default submission contains text. A parent can instead supply a bounded structural contract for a typed object. The same compiled contract generates the tool declaration and validates the actual arguments, including nested values. Unsupported schema features refuse at launch rather than claiming checks the implementation ignores. References do not fetch URLs or local files.

A custom result is the object itself, not JSON quoted inside a string or tucked under an extra default wrapper. Arrays, booleans, nulls, and nested fields survive validation, durable storage, joins, and readback. Invalid arguments leave the child unsubmitted so it can correct them. Structural validity says nothing about whether the described work is correct; the parent must inspect that work.

Persistence comes before publication. The child validates, writes the complete object to a restrictive same-directory temporary file, flushes it, and atomically replaces `submit.md`. Only then can a join observe submission. A failed write leaves the child unresolved. The file is one JSON object, not an append-only stream that needs a repair parser after a partial final row.

A child can receive replacement work. Accepted admission creates a new work generation and invalidates the old visible submission. An older turn cannot submit for the newer work, so yesterday's accepted answer cannot satisfy today's task. Old durable files may remain as evidence; they cannot resurrect current success after a restart or shutdown.

Parent submission is bottom-up. Any retained child without a current submission blocks it, including a failed child. The parent must resolve or shut down that child. Failure cannot become a fabricated empty success merely to unblock fan-in.

Even a published result is not quiescence. The submit tool may still need to append its result, and the child actor may still need to finish and save. A join can learn the answer before disposal is safe. Cleanup waits for reliable completion and drain rather than interrupting the child inside its own submission callback.

## A question must break the join

A child sometimes needs an answer before it can finish. `send_message_to_parent` records a question and blocks that tool until the parent replies. `submit` means completion; overloading it with questions once created a circular definition of done.

If a parent waiting in `join_agents` stayed asleep until submission, the child could wait forever for an answer from a parent waiting forever for the child. Fan-in therefore returns early with pending questions. The parent replies, then resumes the join. A sole pending question needs no copied opaque ID; several questions require an explicit selection rather than a guess.

Questions belong to the actual sender and recipient. Child notices enter as labeled tool-origin information, not system or developer instructions. The root's parent is the human for messaging purposes, so root questions go to the human inbox rather than waking the root to answer itself. Displaying the inbox does not consume a question.

A running wait is not the running work. When a human hint arrives, the actor records it and wakes the observer promptly, regardless of its callback timer. The underlying Job keeps the same handle and eventual result. A later wait can collect it after the Agent has dealt with the interruption.

Each Job owns a replaceable observation pulse. A message completes the pulse an active observer captured; later observers see a fresh one, not an old interruption lingering in a queue. This is separate from execution cancellation and shutdown. An interrupt with no current observer does nothing.

The terminal must keep reading while the actor works. The first-edition command loop blocked input on the same execution path that was waiting for a tool. Its comment promised mid-turn steering; its structure made that impossible. Library tests could inject messages directly and never notice the human had no working wire to the mechanism.

## Correct the protocol, then stop correcting

A successful delegated turn that ends without submission receives a framework correction. If unresolved children remain, it is told to join or shut them down; otherwise it is told to submit. The correction occurs at turn completion, not through a polling timer.

After exactly three consecutive framework wakes, the runtime stops prompting and escalates failure to the parent. Only durable submission resets the counter. New work can clear an old failure state without granting an endless fresh correction budget. An actual failed provider turn follows the error path immediately rather than spending three more requests asking a broken exchange to behave.

The historical exhausted child initially remained labeled running. That made a parent wait for work the framework had stopped trying to finish. Failed, with an error and no submission, is a much more useful state. Activity, failure, and submission are independent observations, not interchangeable colored badges.

Other first-edition failures came from the same seam: tests supplied the factory, human sink, and recursive child construction that the executable forgot. The first spawn failed, human questions went nowhere, or the fleet stopped at one generation. Current direct ownership removes the factory machinery, but only exercising the actual host establishes that its users can reach the library's capabilities.

The source records current local C# runs through root, child, and grandchild, typed submissions, human replies, busy hints, and joined subtree cleanup. A historical live fleet adds a useful narrower observation: a hint reached the later provider request, but its requested marker never appeared in the final result. Delivery worked; compliance did not. The parent still has to read what came back.
