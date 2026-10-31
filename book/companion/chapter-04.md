# Chapter 4: Jobs, or Why a Tool Call Is a Process You Supervise

The failure that motivated this chapter was a screenshot. The tool stopped returning, the agent remained apparently busy, and there was no way to recover the conversation. Bill killed the process, taking unrelated work in the agent tree with it.

That detail matters. A repair limited to shell commands would not have fixed the screenshot. The problem was the shape of dispatch: the application handed control to a tool and waited for the tool to give it back.

A timeout seems like the obvious answer until the tool is a healthy build, a debugger, or a server intended to stay alive. Silence cannot distinguish those from a hang. The more useful question is how long the caller wants to wait before looking again.

## Give the operation a place to live

An asynchronous handler already continues independently of its caller. Abandoning the wait does not necessarily stop it; it may simply discard the only convenient route to its eventual result. The job system keeps that route.

Every ordinary tool call passes through one dispatch path. The dispatcher allocates a Job, gives it a handle, records the call, and only then releases its worker. A logging failure before that release can refuse the operation without an unrecorded effect escaping into the world. The Job owns the lifecycle and retained output; the caller owns an observation of it.

Fast work still returns a normal completed result. Slow work returns a running report and a handle. The caller can inspect the same operation later rather than launch another copy and risk duplicated effects. A running report is an answer to the observation, not an answer from the unfinished tool.

The distinction removes a whole class of alarming error messages. Without a handle, an abandoned call needs a warning that it may still complete and should not be blindly repeated. With a handle, the application can say where the work is and how to return to it.

Handles are allocated across an Ensemble so its Agents do not accidentally reuse them. Output belongs in each Agent's data directory. A convenient CLI directory is not a global location that every embedded application must share.

## Put supervision above the tools

A file read, a directory listing, a shell command, and a tool supplied by an application all need the same observation rules. Otherwise responsiveness depends on knowing in advance which tools might block. External tools make that prediction especially unreliable.

The job boundary therefore lives at dispatch, not inside six individually improved handlers. A new ordinary tool inherits supervision by entering through that boundary. Its author does not get to opt the whole conversation back into an indefinite wait by choosing an unfamiliar name.

The controls that supervise jobs are the deliberate exception. Waiting for a job must not recursively create another ordinary job that then needs supervision itself. Input, waiting, and termination operate on the existing handle. A one-shot observation setting changes how a subsequent call is watched; it is not a new execution subsystem.

This separation also makes ownership intelligible. The Agent owns Jobs, Jobs owns each Job, and the work's resources remain reachable through those relationships until cleanup has finished. The observation may be brief. The ownership cannot be.

## Keep the full result, bound the conversation

A megabyte of test output is a useful artifact and usually a terrible model message. The framework needs both facts to be true.

The full tool output is retained on disk. Each observation returns a bounded preview, the job's status, and a locator for the rest. Later waits return newly available output rather than replaying the entire transcript on every poll. The model can request a larger excerpt or inspect the artifact when the next decision actually requires it.

Truncation must be visible. A shortened preview that looks complete can make a passing build appear to have no final status, or hide the one diagnostic that matters. Status and exit information therefore remain explicit instead of being inferred from whichever text happened to fit.

Durable output also improves conversation reduction. When a large result no longer belongs in working context, a compact reference can still lead back to the actual evidence. Without retained output, the only recovery instruction would be to rerun the effect. That is wasteful for a read and potentially dangerous for a mutation.

These jobs are not a durable distributed queue. Their output can survive for inspection, but their execution lifetime belongs to the running application. Application exit requires shutdown, not a fiction that every in-process task will continue elsewhere.

## Observation is a call policy

Early watchdog designs tried to infer how long each tool could legitimately block. A table of tool names missed unfamiliar tools. Scanning descriptions for timeout hints added another unreliable source of truth. Requiring every tool to declare a blocking limit repaired the table while preserving the wrong question.

A caller may want to check a build after a fraction of a second, wait longer once it is clearly running, or wake when a debugger prints a recognizable prompt. Those are choices about an observation. The tool's identity does not determine them.

Ensemble lets the call select its delay, output cap, and optional pattern. Expiration wakes the observer; it does not kill the operation. Pattern matching is another reason to look again, not evidence that the process has finished.

A one-shot override applies to the next call and reports where it was consumed. That last part matters. A model can prepare a generous output allowance for a file read, glance at a directory first, and accidentally spend the allowance there. Silently retaining or guessing the intended target would create more hidden state. Naming the consuming call makes the mistake local and repairable.

A completed job remains observable too. Completion does not erase its handle just because the first caller was absent when it happened.

## A terminal is a conversation with a process

An interactive debugger needs more than a command string and a future exit code. It prints a prompt, waits for input, and produces another prompt. A pseudo-terminal supplies that interaction, and the Job supplies its identity.

The shell tool therefore runs with terminal semantics. Standard output and error are merged in the terminal stream, while exit status remains a separate fact. The process can receive additional input through its handle. The observer can wait for a prompt without treating the prompt as termination.

A fresh shell invocation starts fresh. Changing directory or exporting a variable inside one call does not silently alter later calls. State persists when an explicitly retained interactive process owns it, not because the application happens to reuse a global shell. A working-directory override belongs to the invocation and is reported as such.

That model makes a debugger, a test runner, and a development server variations of the same supervised work rather than three unrelated tool frameworks.

## Cancellation must tell the truth

Ending a wait leaves the operation alive. Explicit termination asks the owned work to stop and records why. Application shutdown ends all remaining owned work before the Agent's relationships disappear.

The mechanism depends on what is running. A pseudo-terminal process group can be terminated and drained. Arbitrary managed code cannot safely be forcibly aborted by .NET. Its handler receives cancellation and must cooperate. A handler that ignores cancellation can delay shutdown; changing its recorded status does not make its thread or effects vanish.

This limitation is not a reason to obscure the boundary. It is a reason to document the contract of application-supplied handlers and keep their resources attached until actual cleanup is complete. The parent paths used for logging and final reports must still exist while workers are finishing.

Supervision gives the agent a way back to unfinished work, a bounded view of large results, and an explicit way to end what it owns. It does not yet guarantee that the agent can hear a new instruction while watching a job. If the conversation loop itself is waiting, the user can still be talking to an unread inbox.

That is the next ownership problem.

*Implementation details: [codebook Chapter 4](../chapter-04.md).*
