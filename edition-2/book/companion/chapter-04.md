# Chapter 4: Jobs, or Why a Tool Call Is a Process You Supervise

The screenshot tool was supposed to grab the framebuffer and return.

In the first-edition incident preserved in the [source chapter](../edition-3/chapter-04.md), the narrator asked for a screenshot. The tool call never returned. The agent still looked busy, but its turn never completed and no observer fired. Bill had to kill the process. That killed every agent in the tree, including sub-agents doing unrelated work without trouble.

That it was a screenshot matters: this was neither a shell command nor a network fetch. A repair confined to shell execution would have missed the actual hang. Substitute a test suite, a development server, or a tool supplied by somebody else, and the underlying defect remains: the caller has handed over control until the callee chooses to return. Chapter 3’s tool loop can survive a failed command. It cannot survive one that never finishes answering.

The tempting repair is a timeout. Give everything thirty seconds, then cancel. That certainly returns control. It also kills the healthy thirty-one-second build and makes the development server impossible by definition. A hung screenshot and a running server can both be silent; elapsed time alone does not tell you which one you have.

Bill’s more useful question is not “How long may this operation live?” It is “How long should I wait before hearing from it again?”

Those are different clocks. Keeping them different is the chapter.

## A Handle Instead of a Hostage

The design introduces a job between calling a tool and receiving its eventual result. The job owns an identity, accumulated output, lifecycle state, and enough machinery to observe or control the work. The first call may still return a completed result immediately. If the work takes longer, the call returns a running report and a handle instead. Work continues under supervision.

Three seconds is the chapter’s default wake interval. It is a policy for returning control, not evidence that a process has misbehaved. A report can truthfully say that job 3 is running, has produced no new output, and has stored zero bytes so far. The model can wait again, do something else, send input if the tool supports it, or explicitly kill the job.

The useful change is that silence is now observable. It no longer confiscates the conversation.

This also avoids splitting the tool vocabulary into foreground and background variants. The model need not predict whether a command will finish quickly before launching it. A fast read looks much like the old read. A long command acquires a handle when the wait ends. The lifetime belongs to the job in either case; the difference is how much the caller observes during its first visit.

That separation must hold below the user interface. A context deadline attached to the work would recreate the timeout design, however politely the report said “still running.” Likewise, cancelling a waiter must not quietly cancel the operation it was observing. Killing work is a distinct decision with its own recorded reason.

Jobs here do not survive application exit. They outlive an individual wait, not the process that owns them. Ensemble still has to shut them down. This is supervised work inside one application, not a durable distributed queue smuggled into a coding agent.

## The Shell Is Not the Boundary

A shell command is the obvious place to notice a hang. It is the wrong place to put the whole remedy.

The source uses shell-only supervision as a negative control. Its two file-tool checks are load-bearing: remove either the `read_file` job-model check or the large-output check, and the wrong implementation can score 100. That is a measured weakness in what the reduced test suite can distinguish, not the original implementation’s score. Without the common job path, a large file result can flood the model instead of receiving the bounded preview, and a blocked read can still freeze the loop.

The design belongs at dispatch, where *every ordinary tool* enters execution. Reading a file, listing a directory, launching a command, and invoking an application’s custom tool all pass through the same job lifecycle. Otherwise the framework can only promise responsiveness for tools whose internals it happens to know. The screenshot might live behind a foreign tool. A registry entry should not be able to opt the whole agent back into captivity merely by having a different name.

This makes ordering important. A job’s handle must exist before its handler can have effects. The `tool_called` event carries that handle before execution begins. If a tool writes a file and then fails, the record already identifies the job responsible. Allocating a handle only when the first wait expires would leave fast calls, early failures, and effects before the timeout outside the supervised account.

In the chapter’s model, Ensemble allocates integer handles so its Agents cannot accidentally reuse them. Each Agent owns its own jobs collection and output directory; a job retains that collection as its parent. The specimen `cr/io/3` belongs to the CLI’s data-directory convention, not a global storage location for every embedded Agent. The handle names supervised work, not a process ID and not a promise that the underlying operation is a shell.

Admission and execution also have to agree during shutdown. If shutdown closes admission while a call is arriving, the application cannot accept new work after deciding that all outstanding work has been cleaned up. Nor can cancellation remove a job from view while its delayed startup quietly launches a process afterward. Cancellation before startup must prevent execution from beginning; otherwise the application has lost track of work it still allowed to start.

A goroutine is not a lifecycle. “We start it asynchronously” explains how a caller escapes blocking; it does not explain who owns the work, whether it may still start, or who will clean it up.

There is a small, necessary exception to the ordinary-job path. The four supervision tools—`wait_for_job`, `send_input`, `kill_job`, and `tool_limits`—operate the machinery directly. Waiting for job 1 must not allocate job 2 to wait for job 1. Killing a job must not create another job that now needs killing. Their calls remain part of conversation history, but their own call records do not acquire job handles.

This is a semantic bypass, not a second execution engine. Ordinary work enters the common lifecycle. Control operations inspect or alter that lifecycle without becoming recursively supervised work themselves.

## Observation Has Policy Too

Once a call can return while work continues, “wait” needs a more precise meaning than “sleep for a bit.”

The source’s watchdog history shows why this policy belongs to the caller. An early repair tried to infer how long each tool might block from a name-to-timeout map and its description. A missing tool got the default; a healthy tool whose own wait exceeded that default could be abandoned mid-wait. The next repair made tools declare their blocking contracts. That repaired the missing information but kept the question that required the table.

Bill changed the question. “How long may this tool block?” demands knowledge about the tool. “How long before I want to look again?” is a choice on this call. A caller can look shortly after launch, wait longer for a build, then choose differently on its next observation. The missing-row defect has nowhere to live in that design. This is more valuable than maintaining the timeout map more conscientiously.

The chapter has three observation controls: a delay, an optional output pattern, and a maximum inline byte count. A call may ask to wake after a particular interval, when unseen output matches a regular expression, or when the job ends. The byte cap limits how much comes back in the report. None of these authorizes the work to run longer or grants it additional capabilities. They govern the caller’s view.

The precedence is deliberately short: defaults, then any pending `tool_limits`, then the call’s own arguments. The waiting tools expose observation arguments directly. Other local tools, and tools whose schemas the framework does not own, may have nowhere to put them. `tool_limits` supplies that missing route by staging limits for the very next call. It is not a persistent Agent or Ensemble policy setter.

“Next” means the next attempted tool call, not the next convenient target. A `kill_job` consumes the setting. Another `tool_limits` consumes it. An unknown tool or malformed call spends it too: dispatch applies and consumes pending limits before validation or execution. Calls consume them in model-response order, even when one response contains both setter and target; this is about dispatch order, not which job finishes first. An explicit argument wins for its call without becoming another pending setting.

A friendlier rule might save limits until the next job-creating or waiting call. It would also leave the model remembering which categories count and wondering whether a setting from several calls ago was still armed. The source rejects that hidden persistence. One setting, one next call, then gone.

Two consecutive setters expose a subtle consequence. The first setting governs observation of the second call and is spent. The second creates a fresh pending setting from defaults plus only its own explicit fields. If the first raised the byte cap and the second sets only a delay, the elevated cap does not carry forward. Otherwise replacement would quietly become accumulation, and an omitted field would retain a history the model could no longer see.

Literal rules can still misfire. The source records a live model calling out the danger: “it's easy to burn it on the wrong thing.” Set a generous cap for a file read, glance at a directory first, and the directory listing gets the allowance. The subsequent truncated read looks inexplicable unless the consumption is visible.

The repair is a diagnostic, not a target-matching language. Every consuming call’s result begins with a line naming the tool that consumed the limits and the delay, pattern, and byte cap applied. This includes the four control tools that are not jobs. The setter’s reply also warns that whichever call comes next will report consumption. A mistaken burn now explains itself in the result that caused it, close enough to correct without reconstructing several rounds of hidden state.

The argument details preserve that local control. Fractional delays are useful for a quick look, and zero means look now; a negative delay, nonpositive byte cap, or invalid regex is an ordinary tool failure. A pattern selects the output worth waking for while the delay still bounds the observation. None of these choices should turn into an undocumented persistent setting.

The distinction from lifetime limits remains absolute. “Wait up to thirty seconds for this report” does not mean “permit thirty seconds of execution.” Changing the observation policy cannot accidentally become permission to kill.

## A Wait Is a Read With Consequences

A job report is not a subscription to all future output, nor is it a fresh copy of the complete transcript each time. Each job has a cursor. A wait reads output after that cursor and advances it.

This is destructive observation of the incremental stream, even though the retained output file is not destroyed. Two callers waiting on the same job share consumption: one can drain bytes the other would otherwise receive. If a second observer needs the complete history, the saved file is the appropriate surface. A wait handle is not secretly a broadcast channel.

The distinction matters most with patterns. Suppose a debugger has already printed `(dlv) ` and the caller has consumed that prompt. After sending the next command, a wait for `(dlv) ` must match a *new* prompt. Matching the entire accumulated file would let the old prompt satisfy every future wait immediately. The debugger would appear miraculously responsive while the model raced ahead of it.

Patterns therefore apply to unseen bytes. They also change what counts as a wake. On a later `wait_for_job` without a pattern, unseen output must return control promptly rather than notify the waiter only for it to go back to sleep. With a pattern, arbitrary new bytes are not enough; the requested match is the interesting event. Completion still ends the wait even if the pattern never appears, and the delay bounds how long the caller waits when the job remains live. A missing prompt should not conceal the fact that the process has already exited.

These distinctions give the predicates their practical precedence: terminal work can be reported as terminal; a requested pattern replaces “any output” as the output condition; the timer returns control without changing the job’s state. They are not three independent reasons to abandon whichever check happens to be inconvenient. In particular, a fragment arriving before a prompt should not defeat a pattern wait, and a timer firing should not turn a live process into a failed one.

Waiting on an already-finished job is valid. It may return retained unread output, or no new output if the cursor has caught up, together with the terminal state. Requiring a job to be live would turn a harmless race—completion just before the wait—into an API error. It would also make the natural sequence “launch, inspect something else, collect the result” unreliable whenever the work was fast.

The initial no-pattern observation deliberately differs from that later wait. It preserves the fast full-result path instead of returning at the first output fragment. When a job completes within the wake delay and under the cap, with no directory override or consumed-limit diagnostic to disclose, its first report preserves Chapter 3’s bare result. The PTY’s merged shell output is an explicit formatting exception, while exit-status information remains. This makes ordinary short work familiar without hiding new facts. It is not permission to block indefinitely: when the wake delay ends, unfinished work yields a running report. Later observations can then return as new output arrives.

This is where the interface stops being just “async.” The hard part is deciding what the observer has consumed, what event it is waiting for, and which facts remain true after the observation ends.

## Keep the Bytes; Spend the Context Carefully

Output storage and output presentation have different budgets.

The chapter keeps a job’s full output under a locator such as `cr/io/3`. The inline report has a default cap of 16 KiB, adjustable per call. When output exceeds the cap, the report gives a bounded preview, the locator, and accurate byte accounting. A model can decide whether to read more instead of receiving a megabyte of compiler complaints as an unsolicited contribution to its context window.

The cap is not a retention limit. Truncating the saved file would destroy precisely the evidence needed to investigate an abbreviated report. Conversely, keeping the full file does not justify sending all of it to the model. Disk and conversational context serve different readers at different costs.

The cursor and cap must not be confused either. A report consumes an interval of new output; its preview may omit some of that interval. The omitted bytes remain available through the file. A caller should not assume repeated waits will eventually reproduce every omitted byte inline. That would make a presentation cap behave like an undocumented pagination protocol.

The source’s flood fixture is useful because it distinguishes these surfaces: the complete retained output must include its beginning and end, while the inline report stays bounded and names what it left out. The large-file read is equally important. A shell-only flood test can pass while every other tool retains the old unbounded return path.

The saved record also gives completion and killing a durable representation. Ordinary command completion has an exit status; a deliberate kill has its own terminal meaning. Those facts belong in job metadata and events as well as in a readable report. Model-facing prose is a convenient account, not the sole source of state.

This continues Chapter 3’s distinction between effect and narration. “The tests passed” is not a substitute for the captured result. “The job is done” is not a substitute for a lifecycle that can explain how it ended.

## A Terminal Is More Than Two Pipes

Returning control solves only half the problem for an interactive program. The caller also needs a way to talk back.

`send_input` writes to an existing job and then observes its response using the same wait machinery. It does not launch the command again. That makes a prompt and answer possible across separate tool calls: start the program, see its prompt, send input, receive the reply.

The shell handler and the shell job have different finish lines. The handler returns once the process is running; from there the job owns the process, while a reader goroutine collects output and the eventual exit status. Waiting for exit inside the handler would leave a blocking call underneath the new handle.

Dispatch must respect the other half of that handoff: handler return is not job completion. Finishing the job closes its output file. If dispatch does that as soon as launch returns, it cuts off the reader still collecting the process’s output. The result can look like successful execution with an empty report and zero saved bytes. That is not a formatting problem in the capture code. It is the wrong owner declaring the lifetime over.

For a shell job, stdin and stdout pipes seem like enough. The source’s debugger experiment says otherwise. Delve refused an ordinary nonterminal invocation, emitting “Stdin is not a terminal.” A fake prompt printed by a tiny helper could never establish that a real interactive debugger would accept the same setup.

The design therefore runs shell work through a pseudoterminal. A PTY gives the child terminal behavior rather than merely a pair of byte streams. Programs can detect the terminal, use their usual interactive modes, and issue prompts the model can answer. Output is a merged terminal stream, with the buffering and echo behavior that implies; it is not two pristine independent stdout and stderr logs.

The source’s deterministic input check makes a subtle distinction here. Waking on the command’s echo is not the same as waking on the program’s reply. An echoed input line proves that bytes reached the terminal machinery. It does not prove that the program processed them. The test waits for the meaningful prompt or response, not just any comforting repetition of what was sent.

The historical debugger demonstration then reaches beyond that helper. The source records a second-edition Gemini session driving `dlv` through the tool: launching it, waiting for its prompt, setting a breakpoint, continuing, reading a variable as `42`, and quitting with exit code 0.

The mistake came in reading that evidence. The student saw escaped text in a serialized display and concluded that the pattern contained two backslashes before each parenthesis. It also concluded that the first four calls had waited their full ten seconds, and paid for a repeat. The original log disproved both claims: the decoded pattern was `\(dlv\) `, and each of those four calls had returned in under three-tenths of a second. Quit correctly waited for exit without a pattern. The repeat worked, but it answered a question the retained evidence had already settled. Read the decoded argument and measured intervals before diagnosing behavior from its printed representation.

That is a good reason for a live demonstration: it answers a concrete question a toy cannot. It does not establish every future debugger version, terminal mode, or operating system. The chapter records PTY measurements on macOS and treats Linux behavior as assumed, not measured. Portability should not acquire a checkmark merely because the chosen library is portable in principle.

## A Directory Is Not a Shell Session

A PTY does not mean every command shares one persistent shell. Each `run_command` starts a fresh shell job. That isolation matters when two commands overlap: one bash cannot serve as two independently executing jobs with unrelated working directories and state.

Persistence is still available, explicitly. A shell can itself be a long-running job, with later `send_input` calls changing its directory and running commands inside it. Its handle identifies the session, and the ordered input events preserve what happened there. The source’s distinction is not state versus no state. It is attributable state versus an ambient shell whose history silently affects the next call. You can compose a persistent session from isolated jobs; you cannot recover isolation merely by wishing a shared shell had forgotten something.

The source reports that roughly seven in ten archived command calls began with `cd`, mostly to the same few directories. That looked like a demand for persistence until the measurements separated directory changes from environment setup, which was rare. The agent was repeatedly correcting where it had been launched, not rebuilding a rich shell session. The remedy was a sensible launch workspace plus a `cwd` argument for an individual call.

A relative `cwd` resolves against that workspace. A missing directory is an error, not an invitation to fall back silently and run the command somewhere else. The effective directory belongs in the job’s account and appears in the report when it differs from the default. The serialized `tool.job.cwd` field records the optional override, however, rather than repeating the default absolute path on every job. Absence of that override does not mean execution has no working directory.

The workspace and the directory holding logs and saved output are separate configuration. Commands may need to run in a project while their event log lives elsewhere. Deriving one location casually from the other makes moving evidence storage change program behavior—a remarkably expensive interpretation of “keep the logs over here.”

There is deliberately no sticky `set_cwd`. A directory change made dozens of calls ago can disappear from compacted context while continuing to govern destructive commands. Making that safe would require keeping the setting permanently visible to the reducer. A one-call override has no such afterlife. The wrong wake costs attention or time; the wrong directory can remove the wrong build tree. Both knobs stay local to the call rather than acquiring invisible persistence.

## Killing Has to Mean Something

Once work can persist across calls, stopping it becomes an ordinary operation rather than an emergency trip to Activity Monitor.

But a shell job is rarely one process. The shell can start children, which can start children. Killing only the shell PID may leave the test runner or browser happily alive. The chapter uses process-group control so cancellation reaches the shell’s group rather than just its first process.

That is a practical cleanup boundary, not a sandbox. Process groups do not constrain filesystem access, network access, or credentials. Nor do they prove that no descendant can escape the group by creating a new session or group. Chapter 3’s shell still has the authority of the launched process; Chapter 4 makes a defined set of that work more manageable. Supervision is not confinement with a friendlier name.

There are two truths to preserve when a kill happens. The external work must be stopped as far as the control mechanism promises, and the job must become terminal as *killed*. A late completion must not overwrite that state with an ordinary success or failure.

The source’s audit caught a wonderfully dishonest-looking report produced without any need for a dishonest model. The process was killed, but the finishing goroutine marked the job `done` with `exit_code -1`. The report included both the ordinary completion claim and a note saying the job had been killed. The grader searched for the word “killed,” found it, and awarded the point.

The failure explained itself in the very text that satisfied the check.

The corrected observation checks structured state and rejects a post-kill report that calls the job done or presents an ordinary exit code. A killed process may of course produce a low-level wait result in the operating system. That does not make the framework’s public outcome “completed command, exit -1.” The reason the operation ended is part of its result.

The late-result rule matters for non-shell work too. A file read can remain blocked below the job layer and return after cancellation. Marking the job killed cannot necessarily interrupt every underlying system call or foreign handler. It can, and must, prevent the eventual return from resurrecting the job as successfully finished. Logical lifecycle control and physically stopping an operation are related promises, not identical ones.

Killing must also wake observers. A caller waiting thirty seconds should not have to wait out the remaining interval after another caller kills the job. Terminal state is already available. Notification is part of the state transition, not an optional optimization for impatient callers.

Shutdown exercises the same machinery with a different reason. Outstanding jobs are killed because their owner is exiting, admission is closed, and cleanup is waited for rather than merely announced. “Cancellation requested” and “shutdown complete” are different claims; returning from shutdown should not leave its caller guessing which one the application means.

A live test in the historical account missed the relevant observation: its sleep had expired before EOF, so it proved nothing about shutdown killing. The bounded repeat first established that a longer-running process was still alive, then sent EOF, then observed it gone and inspected the shutdown reason. To show that an action stopped something, the something has to be running when the action occurs. Even sleep can trick a demonstration if nobody checks the clock.

## Concurrency Tests Need Events, Not Hope

The races in this chapter are small enough to fit in a few lines and large enough to invalidate the architecture.

Consider a test that starts a waiter goroutine and immediately kills the job. If the waiter has not actually begun waiting, it can simply observe the already-terminal state. The test passes even if killing fails to notify existing waiters. In the second-edition review recorded by the source, removing notification still passed the initial test twenty times.

Twenty passes did not make the ordering more real.

The corrected test establishes that the waiter has reached the relevant point before killing. Similarly, releasing a blocked file read does not prove its late result has reached the job’s finish path. An assertion made too early can miss the overwrite it was meant to catch. The test needs evidence of the transition, not a plausible amount of sleep after releasing it.

Even the repair had a failure path: its first version raced during cleanup when the intended assertion failed. A buffered result channel let both correct and defective executions finish cleanly. The intended assertion appearing in the output did not cancel the race warning beside it.

This is the difference between designing a test around a schedule and designing one around an ordering. A sleep suggests that the scheduler has probably done something by now. A synchronization point establishes that the thing relevant to the claim has happened. Timing remains useful for testing a wait interval, but it cannot stand in for evidence that a goroutine entered the state whose notification is under test.

The source applies the same skepticism to its larger checks. A mutation must actually change the intended behavior before a test failure counts as detection. A mutated program that cannot build because its PTY dependency was omitted demonstrates broken test setup, not a protected lifecycle. A missing debugger that causes a skipped check is an untested mechanism, not a passing interactive path.

Even a legitimate compatibility failure can have a surprising cause. The inherited Chapter 3 fixtures assumed commands returned synchronously. Under the new design, a cold `go run` that exceeds the default wake can correctly produce a running report where the old fixture expected an exit code. That is not evidence that the process was cancelled or the wake is broken. The fixture’s timing assumption has become visible.

The account keeps historical scores and failures separate from later corrected checks. A perfect score from checks unable to distinguish shell-only supervision would not prove a dispatch-wide job model. Nor do accepted student checkpoints with failing results become clean passes because the chapter moved on. A score describes what a particular observer noticed. The underlying claim still needs an observer capable of disagreeing.

## Take the Library Outside Its Own House

The chapter has a second architectural test: a small application in a separate Go module imports the public library and registers a custom tool.

This is not a second loop. The consumer creates an Ensemble and Agent through the public API, adds a declaration and handler, asks a question, and shuts down through the public lifecycle. Its tool enters the same dispatch and job machinery as the builtins. If the consumer has to import internal packages, assemble a private scheduler, or copy the conversation loop, the public boundary has not actually carried the application.

A separate module makes that failure harder to hide. Code inside the repository can casually depend on internals that a real importer cannot reach. Building the consumer independently asks a narrower and more useful question than whether the core CLI still compiles: can an outsider compose this capability through the interface we claim to offer?

The tool itself need not be glamorous. The source uses `shout`, which derives an uppercase result from its argument. Its simplicity helps isolate the framework path. A declaration with a name, description, and object schema must reach the provider; the requested handler must execute; its successful result must return under the same call ID; and its output and job metadata must be saved. A final answer saying “HELLO FRAMEWORK” proves little if the model could invent it, use a shell, or recover from an unknown-tool error.

Normal registration adds a tool to the Agent’s existing registry rather than replacing its builtins. The historical accepted consumer disabled builtins specifically for its demonstration, removing the shell as an alternate way to produce the same answer. Those are different choices for different purposes: an extensible library should preserve the caller’s selected capabilities, while a focused demonstration can deliberately narrow them to make the observed path unambiguous.

The source records successful consumer runs through all three providers, with the exact derived output stored in the job and correlated back to the call. These are historical observations, not runs performed for this companion. They support the external-use path exercised then; they do not make every current public API spelling or every internal ownership relationship self-proving.

One of the first successful runs also lost its event capture. The launcher mistook a log header for an event and deleted scratch too early. The model call had succeeded; the evidence collection had not. Repairing capture required a repeat. Saving the raw events before parsing them would have made that parsing mistake cheap.

## The Logger Follows the Parents

The external tool also emits a timestamped diagnostic. That tiny act tests a larger ownership decision.

The original refactor story, carried into the source, reached thousands of lines of Go without a logger and then introduced one. A logger is wanted nearly everywhere. That makes it an excellent temptation to add a global, or to pass another little service bundle through every function until nobody remembers who owns it.

Ensemble already supplies the root of the application. It owns the logger. Agent retains its Ensemble, Engine retains its Agent, and the executing tool retains access to its actual owner. The diagnostic follows that chain to a narrow `Logf` interface. The caller does not need the concrete logger, another copy of it, or a direct import of the sibling package that happens to implement it.

“Actual” is doing work here. Constructing a convenient substitute Agent with the same logger field would make the message print while breaking the ownership account. So would a mutable global that happened to point at the right writer during the test. With multiple applications or Agents in one process, those conveniences become cross-talk.

The source keeps the star topology from the earlier architecture: shared vocabulary belongs in the common hub; behavior stays with its responsible package. An acyclic sibling import can still violate that design. Go’s compiler proves absence of import cycles, not fidelity to ownership. Where the receiver rule prevents a method from living beside the responsible behavior, a free function is preferable to moving the behavior into common merely to make method syntax possible.

The standard logger provides synchronized writes and timestamps behind the narrow interface. There is no need to invent a logging subsystem to prove this point. The interesting result is that a leaf unknown to the framework can reach the application’s existing facility through its parents.

Still, the printed line establishes only that a line reached a writer. Source inspection has to establish the path. The source records an earlier logger check rejecting a valid multi-hop path because it expected a particular field selector, and a globals check relying on variable names. Neither spelling preference could settle architectural ownership. Runtime evidence and source review answer different questions; making one impersonate the other produces very confident nonsense.

## Responsive to Work, Still Deaf to You

The chapter’s model now has a useful rhythm. Launch a long test, receive a running report, inspect another file, wait for new output, collect the eventual result. Start a server without requiring it to terminate. Talk to a debugger through a real terminal. Stop a job without destroying the conversation that explained why it was started.

The second-edition live account records delayed jobs, prompt-and-response exchanges, and explicit kills through three providers. One session also demonstrated a smaller inline cap while retaining more bytes on disk. Those observations show the mechanisms composing in real exchanges. They do not replace the targeted ordering tests, establish unmeasured platform behavior, or prove that every foreign handler can be forcibly interrupted.

Four control tools and one ordinary-job path are enough for this stage. There is no external tool server connected yet, even though a tool the framework does not own motivates the design. There is no new scheduler in the consumer. There is no mailbox or interrupted-turn state hiding inside job supervision.

That last omission is something a user can feel. Start the long test, then type “Stop, wrong directory.” The work no longer holds the agent forever, but the foreground loop can still be parked in its current wait. The input may not be noticed until the wake returns. A handle makes the work addressable; it does not make the human’s message arrive sooner.

Chapter 5 changes that foreground shape: prompts, interruptions, and job completions need a way into a loop that does not block on a tool. It is a separate problem, not a reason to bolt more meanings onto a wait timer here.
