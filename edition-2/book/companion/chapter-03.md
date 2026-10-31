# Chapter 3: Six Tools, Ninety-Two Percent of an AI Coding Agent

Bill counted the tools after months of using them.

The CodeRhapsody corpus described in the [original chapter](../edition-3/chapter-03.md) contained 511 sessions and 70,401 tool calls. Five tools accounted for roughly 91 percent: running commands, reading files, editing files, searching, and writing. Adding directory listings brought the share to 92 percent. Fifty-one available tools, and most of the work fit in six.

That is evidence about Bill’s recorded workload, not a minimum viable agent established by nature. The logs were private, and several newer tools, including much of the sub-agent suite, postdated the measurement. Their absence could not establish their uselessness. Even the directory-listing tool earned its place partly through judgment: orientation is cheap; guessing paths is not.

The more revealing result was what happened *within* those ordinary operations. Targeted edits outnumbered whole-file writes about seven to one. Yet the frequently used edit tool had a defect Bill discovered while writing the chapter: when an anchor appeared more than once, it changed the first occurrence and reported success. The zero-match case refused correctly. The many-match case guessed quietly.

Chapter 2 gave the agent a durable account of conversation, a reducer, and vendor-specific request boundaries. It could preserve a request to run a tool without running anything. This chapter connects that recorded intention to effects: files read, bytes changed, commands executed. The difficult part is less the filesystem API than making “requested,” “ran,” “failed,” and “finished” mean different things all the way around the loop.

## A Tool Call Is the Middle of a Turn

Consider an ordinary repair. The human asks about a failing test. The model explains that it will inspect the failure and requests a command. The agent records the response, runs the command, records its output, and sends the updated context back. The model may then read a file, edit it, and run the test again. Only when a response contains no more tool calls does the turn finish.

The human has supplied one request. The model and tools may have exchanged several rounds underneath it. Returning after the first tool result would stop the repair just as the model acquired the information needed to continue.

The conversation remains Chapter 2’s conversation. `ResponseEnded.Parts` preserves ordinary text and calls together, in their original order. A call part records what the model asked for; dispatch and return events record what the runtime did about it. Agent coordinates those facts through History. Engine continues to own the vendor exchange rather than acquiring a second history or a private tool-loop policy for each provider.

The caller receives only the final response’s text. Intermediate narration still belongs in history, but it is not part of the final answer returned through the existing interface. The source’s student implementation initially concatenated “I’ll check” with the eventual answer. That was not lost data; it was data returned through the wrong interface. Preserving the conversation and selecting the answer are separate responsibilities.

Every result refers to its call’s ID, not merely its name or its position. Two reads of the same file can ask for different ranges. Two calls in one response can fail differently. When a provider omits an ID, the parser assigns a deterministic, distinct identity while capturing the call, using stable event identity and part position. Waiting until rendering is too late: dispatch already needs the identity that the log and eventual result will share.

Calls execute one at a time, in the order requested. This is an intentional constraint, not an unfinished concurrency feature. A command can change the tree that the next read observes. Running the same calls concurrently changes the meaning of the response unless their independence has been established elsewhere.

The loop also has a round limit. The source uses sixteen rounds; the important property is a finite bound on repeated model exchanges, not that particular number. A model can keep asking, and a scripted fake can keep repeating its last call. The bound prevents an endless conversation with the tools. It does not interrupt a command, impose a deadline, or make a blocked shell responsive. A single round can still last forever.

## One Registry, Two Views

Before a model can request a tool, the request must declare it. A declaration supplies a name, a description, and an argument schema. The implementation adds the function that performs the operation. Those are two views of the same registered tool, not two catalogs maintained by different people.

The registry belongs to an Agent. The composition root installs its definitions; both rendering and dispatch consult that same set. Otherwise a model can be offered a tool that cannot run, or dispatch can expose a capability the Agent was not meant to advertise. A mutable global registry would make one Agent’s configuration someone else’s surprise.

Six tools need little machinery. An ordered collection with declaration and dispatch operations is enough. The source places the shared definitions in `internal/common` and the implementations in `internal/tools`. A handler receives its actual owner through a narrow `ToolContext`, which exposes Engine and the existing ownership chain. That gives a diagnostic access to the services already owned by its Agent without globals, copied sibling services, or a fictional parent object. A synchronous function call does not need a new tool runtime to feel architectural.

An ordinary library client can still construct an Agent with no tools. In that case the renderer omits the declaration field, preserving conversation-only behavior. The command-line application installs the six tools, but no new command or alternate transcript convention is required. The existing chat, stdio, dump, offline-rendering, and log configuration remain the public interface. Capability grows beneath it.

Declaration syntax belongs behind the same vendor seam as messages. The shared schema describes the local argument contract; Engine projects it into the provider’s accepted dialect. The Edition 3 live exercise exposed a concrete difference: Gemini’s `parameters` rejected top-level `additionalProperties`. The repair omitted that keyword from the Gemini projection for these built-in object schemas. It did not loosen local argument validation, alter the other providers’ declarations, or grow a general recursive JSON Schema translator.

That narrow fix matters. A provider’s inability to express a constraint does not make the constraint disappear at execution time. Required fields still have to exist with the right types. Empty `content` or `new_text` is a valid string; a missing field or `null` is not. A structurally valid API response can carry the wrong argument types. Local validation has work to do even when JSON parsing succeeded.

An empty registry does not excuse an unanswered call either. The source’s student initially returned early when there were no tools. An unexpected request then remained in history with no result. The correct response is a correlated error explaining that the tool is unavailable. A declaration governs what should be requested; dispatch must still handle what actually arrives.

## Failure Must Come Back

A missing file is a tool result. An unknown tool is a tool result. Arguments of the wrong type produce a tool result. Each carries the original call ID and an error indication, so the model can correct its next action.

Dropping the result on an error leaves a hole in the conversation. The model requested something, the runtime knows why it did not happen, and the next exchange receives none of that knowledge. Crashing the whole agent is often little better: a recoverable misunderstanding becomes a lost turn.

A command’s nonzero exit status is different. The shell successfully ran the test suite; the test suite failed. That is ordinary command data, not a malformed tool invocation. Chapter 2’s distinction between an exchange failure and an unsuccessful operation now needs this finer resolution: a failed test does not set the tool-error flag merely because the command returned nonzero. The model needs to repair the program, not its request to run the tests.

The blocking command tool therefore reports stdout, stderr, and the process exit code separately. Parsing an error-looking sentence out of stderr is not an exit-code implementation. The source’s original check accidentally demonstrated why: `go run` of a program that exits 7 itself exits 1 and prints `exit status 7`. A check looking for the number in captured text passed even without an explicitly reported exit code. A silent shell `exit 7` exposed the missing observation. The wrapper’s status and its child’s status are different facts.

Results then travel through the normal history and rendering path. Anthropic’s result-first ordering remains a renderer obligation, not a reason to reorder historical events. Its rule needs a request containing both tool results and other content to become observable. This synchronous chapter cannot naturally receive a new human message while blocked in a handler, so an ordinary session can pass without ever testing the awkward ordering. A targeted mixed-content fixture reaches the property that the happy path cannot.

The distinction also prevents accidental reliance on Chapter 2’s recovery placeholder. A wire-only “outcome unknown” result is useful for replaying an interrupted record. It is not a substitute for recording a failure that the live dispatcher actually observed.

## Explanation Before Effects

Some model policies require an explanation before tools. Stating that requirement in the prompt is useful; enforcing it immediately before dispatch gives it operational meaning.

For a selected model, the same response must contain nonblank ordinary assistant text before its first tool call, explaining what will run and why. One explanation can cover a coherent batch. An explanation from an earlier turn does not describe this batch. Tool output, hidden thinking, and a readable reasoning summary are not substitutes. Text after the call is too late.

The runtime can check presence and order. It cannot establish from those checks that the explanation is adequate or that the proposed work is authorized. This is an explanation policy, not a safety proof.

If the batch lacks the required preceding text, the dispatcher refuses the whole batch before invoking any handler. It preserves the response and returns an error for every call, explicitly saying that nothing ran and requesting an explained retry. Refusing only after the first operation would turn a precondition into a commentary feature. The retry consumes the existing round budget, so a model that repeatedly omits its explanation cannot create a second, unbounded repair loop. Models outside the selected policy retain their ordinary behavior.

The Edition 3 Gemini work found a parser bug at exactly this boundary. The inherited parser treated signed ordinary text as hidden thinking. That produced blank answers and false explanation refusals: the model had supplied visible prose, but the application had hidden it.

A signature does not turn ordinary text into reasoning. The text stays visible and portable. Its signature stays bound to that exact text part and to its source vendor, API surface, and model, and is replayed only to that same source. Genuine hidden thinking remains hidden. Chapter 2’s provenance rule therefore protects two different things here: private replay material from being sent to the wrong issuer, and ordinary language from being swallowed by a parser that classified it by the wrong clue.

The decisive observation for refusal is an effect that did *not* occur. A handler that would leave a visible change makes this test meaningful: the unexplained batch leaves nothing behind, and an explained retry executes. Counting returned errors alone cannot establish that the runtime refused before acting. Streaming will later make arrival order another part of the problem; this chapter has complete responses to inspect.

## File Tools Spend Attention as Well as Bytes

A shell can implement all six operations. It can read, search, list, replace, and write. Bill asked whether anything beyond `run_command` was necessary, and in the capability sense the source’s answer was no.

Capability is not the most useful comparison. Dedicated tools change the cost of expressing an intention, the precision of the result, and what happens when the intention is ambiguous.

`read_file` accepts a one-based inclusive line range and an output-size cap. The range is selected before the cap is applied, and truncation is reported rather than hidden. The source’s accepted implementation uses the whole file by default and a 32,768-byte cap; that value is a documented choice, not a universal optimum. A bounded read makes it possible to ask a narrow question without importing a large file into every subsequent request. Chapter 2’s full-history exchange makes that cost cumulative until context management removes the material.

`search_files` makes those narrow reads discoverable. Its regular-expression matches use familiar grep notation: `path:line:text` for matching lines and `path-line-text` for surrounding context. Overlapping or adjacent context windows merge, and separate groups receive `--`. The output format is already legible to the model and to the human inspecting its work; a proprietary search dialect would buy little here.

Context defaults to zero lines. The historical corpus showed that the model frequently supplied an explicit context request, often asking for more than the then-default of two. That supports a cheap default for this workload, not a general claim that surrounding code is unnecessary. A model can request more. Extra text already admitted to history has already spent attention and tokens.

`list_directory` lists immediate entries rather than recursively filling the response. Sorted names and a trailing slash on directories make orientation stable and cheap. `write_file` can create parent directories and new files, but refuses replacement of an existing file unless `overwrite` is explicitly true. The refusal identifies the existing target and its size. Append preserves existing bytes and needs no overwrite flag; an empty content string remains a meaningful input.

That guard protects against one easy mistake: presenting replacement as though it were creation. It is not backup, and the explicit flag is not evidence that the model read or understood the old content. A shell can overwrite the same file without using the guard. The design makes the dedicated operation more specific; it does not make arbitrary shell access safe.

The original account gives less abstract reasons too. During work on the book, the authoring tool refused edits when required heading context was absent or a wrapped word no longer matched an anchor. The refusal created a chance to inspect instead of guessing. Shell editing also carried platform differences and quoting hazards, including backticks being interpreted while preparing text meant to be written literally. File tools remove those distractions from the model’s task. They do not make the intended change correct, but they reduce the number of unrelated ways to express it incorrectly.

The seven-to-one preference for targeted edits is consistent with that argument. It shows what agents used when both operations were available in this corpus. It does not, by itself, isolate the tool’s causal effect from every other feature of the workload. The useful design hypothesis is narrower: affordances shape behavior, so removing one because the shell is theoretically sufficient can change the work an agent actually performs.

## An Anchor Must Identify a Place

The edit defect gives that hypothesis teeth. A tool can offer the right operation and still define success badly.

When an anchor fails to match, several policies are coherent. Strict refusal preserves the file and asks for a better anchor. Fuzzy matching can tolerate stale context but risks selecting the wrong region. Falling back to whole-file replacement can finish the change while greatly widening its scope. The choice affects recovery cost, predictability, and how much authority a short argument grants the implementation.

The same decision has to cover multiple matches. Refusing zero matches while silently choosing the first of several is not a strict policy. It is guesswork with a well-tested error message on one side.

Bill’s original `edit_file` did precisely that. A single replacement with a limit of one made the common case look correct and the ambiguous case look equally correct. The source reports the defect as a discovery during authorship, not as a repair whose later deployment has been established here.

The second-edition student chose exact, unique anchors and found a smaller trap. In `aaa`, the anchor `aa` identifies two starting positions. A routine counting non-overlapping occurrences reports one. It answers how many disjoint replacements fit, not whether the location is unique.

Under the student’s policy, both zero and multiple matches leave the file unchanged. A zero-match refusal identifies the file and suggests rereading it. A multiple-match refusal gives the actual count, including overlapping sites, and a bounded list of one-based starting line numbers. If that list is truncated, the response says so while retaining the full count. The tool does not quietly relax exact comparison to be helpful.

These details turn refusal into useful feedback. “Could not edit” avoids damage but gives the model little basis for a better request. A count and locations expose the ambiguity without returning the entire file. The model can choose a more specific anchor rather than spend a round discovering what the tool already knows.

Strict matching is not the only defensible design. The source also describes CodeRhapsody’s line-replacement tool searching near supplied line numbers. A localized repair can reasonably trade exactness for tolerance. The tradeoff should be explicit rather than smuggled into an error path.

Text anchors have a compositional advantage. An edit near the top of a file shifts later line numbers, while an anchor elsewhere can remain valid. In the reported corpus, anchor edits outnumbered line replacements roughly thirty to one. The mechanism explains why anchors can support several sequential changes without rereading every shifted location. The count records a preference, not proof that line-oriented editing is always inferior.

## A Small Tool Set Is Not a Sandbox

Per-Agent registration creates a place to control which operations are exposed. An Agent with only reading, searching, and listing tools lacks the declared file-mutation and shell operations. That is useful capability separation.

It does not establish filesystem confinement. Read tools may still reach sensitive paths. A launch directory determines where relative paths begin; absolute paths and parent paths can reach elsewhere. A shell bundles many capabilities behind one tool name. Calling it `run_command` does not narrow the authority of the process it launches.

This chapter establishes no general permissions system. Dedicated tools make later restrictions expressible, but selecting a small registry or changing directories is not equivalent to enforcing those restrictions.

There is one narrow credential boundary now. Provider keys belong to the HTTP transport, not to the child shell. Before launching a command, the implementation removes the supported key variables `LLM_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and `GEMINI_API_KEY` from the shell environment while retaining ordinary build settings. A fresh shell starts on each call, so a `cd` in one invocation also does not silently alter the next.

Environment filtering prevents this specific inheritance path. It does not remove credentials from files, other variable names, or resources the process can otherwise access. Dummy key values observed through the actual launched shell can establish that the named variables were removed without exposing a real secret. A helper tested in isolation cannot prove that the launch path used it.

The source’s first book-repository demo supplied a modest warning about effects. The agent invented a project codename and saved it in the repository root. The demo worked and left a file behind. Nothing had to crash for the run to exceed its intended footprint. Disposable workspaces and recoverable source state limit the consequences of a demonstration; they do not become containment merely because they are convenient.

## A Fake Can Agree Too Easily

A deterministic fake makes the loop inspectable. It can serve text plus multiple calls, missing IDs, recoverable failures, and further rounds before the final answer. Unlike a live model, it can repeat the same awkward arrangement every time. Its value lies in stating and exercising a protocol, not in responding intelligently to the prompt.

The source’s fake was generous in a particularly damaging way: it volunteered tool calls whether or not the request declared tools. An implementation could pass the loop checks without sending declarations, then do nothing useful against the real Anthropic API. Bill found the omission by auditing the chapter. The test world had supplied a capability the implementation never requested.

Chapter 2 supplied the opposite generosity: the fake accepted an empty assistant message rendered with `null` content where the real endpoint rejected it. A fake can send too much or accept too much. Either way, agreement between implementation and fake does not establish agreement with the provider.

A fake does not automatically discover that the real system changed. Focused observations of the real API supply missing facts, and those facts must return to deterministic coverage. This is an outer loop for resolving concrete uncertainty, not a reason to make every edit depend on a slow, metered, nondeterministic model call.

Even inherited fixtures can become wrong when the product gets better. Chapter 2’s fake ended with a tool call because that chapter only recorded calls. Chapter 3 answers it and asks for another response. A fake that repeats its last reply then drives the new loop into its round limit and changes accumulated usage. Extending the script with a final, zero-usage reply preserves the earlier accounting assertions while completing the newly possible protocol. Removing the assertions would hide the change rather than explain it.

The useful review question is whether the observation can distinguish the promised behavior from a plausible broken version. A single tool call cannot expose miscorrelation between two calls. An existing-file refusal alone cannot prove new-file creation remains allowed. An error result cannot establish that a forbidden handler never ran. A canned signature cannot establish that the parser captures newly arriving signed text. Each distinction needs a situation in which getting it wrong becomes visible.

## What a Live Repair Establishes

The source records the second-edition student driving its own binary in three disposable Go projects, through Anthropic, OpenAI, and Gemini. Each project contained a greeting returning `helo` and a test expecting `hello`. The models listed the directory, ran the failing test, read the source, made a targeted edit, and reran the test. Inspection found the intended change; the command results moved from exit code 1 to exit code 0.

Those observations support more than an assistant saying “fixed.” The call records establish which operations were requested and returned; the inspected diff establishes the changed bytes; the second test result establishes that the tested behavior passed. A later recall response showed that conversation context survived the work. None of these observations substitutes for the others.

The error demonstration initially failed to demonstrate an error. One model interpreted “No more tool calls” as forbidding the requested read; another declined to read a nonexistent file. Clearer follow-ups produced actual missing-file results. A model declining to act and a tool acting unsuccessfully are different paths, even when both leave the filesystem alone.

An additional bounded session exercised search, write, and read-back, bringing all six tools into the recorded live work. The source explicitly says that the live build preceded the overlapping-anchor and empty-registry corrections. Those fixes had deterministic coverage; the earlier live run is not evidence that the later binary exercised them. This companion reports the source account, not a new reproduction of those runs.

The Edition 3 additions sharpen the diagnosis further. Tool declarations could fail at Gemini’s schema boundary. Visible signed text could disappear in parsing. Separately, a missing resolved model identity could suppress private replay material even with a correct parser. The launch and harness omitted `LLM_RESOLVED_MODEL`; correcting that setup required no source change. The resolved identity comes from an observed returned producer on the same vendor, API surface, and route, not from assuming the requested routing alias is the issuer.

These failures can look similar from the final answer and require different repairs. A model listing establishes what a credential can see, not which model supports tools on the chosen API surface. A newer or larger model name settles neither suitability nor cost. An unknown replay target correctly withholds private extras; it does not prove the parser lost them. Diagnosing from the boundary that actually failed is cheaper than rewriting a working loop.

The resulting agent can now close a coding cycle: inspect, change, execute, interpret, and continue. It remains deliberately synchronous. A long test leaves it waiting inside the handler; a server that never exits prevents the turn from returning at all. Sixteen allowed rounds offer no consolation during round one.

That is the next chapter’s problem. The small job-control share in Bill’s historical count understates its importance: the next design changes the behavior of `run_command`, the most-used tool in the corpus. The agent already has reach. What it lacks is a way to keep observing and responding while the work continues.
