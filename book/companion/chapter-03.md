# Chapter 3: Six Tools, Ninety-Two Percent of an AI Coding Agent

A model that can describe a repair is useful. A model that can read the relevant file, make the repair, and run the tests becomes a different kind of application. The new capability comes from effects, not from a more elaborate conversation window.

The obvious temptation is to expose every operation the application can perform. The project's earlier tool-use records suggest a more selective starting point. Across 511 recorded sessions, the five most-used tools accounted for 91 percent of calls; thirteen accounted for 98 percent. Many carefully imagined tools barely appeared. Targeted editing was used far more often than whole-file writing.

Those figures describe that workload, not a universal distribution of software work. They are still a good reason to begin with a small, useful tool surface rather than a catalogue of possible features.

## Declarations are part of the request

The model cannot call a tool merely because a method exists in the application. The request must advertise a name, a description, and an argument schema. Different providers spell that declaration differently, just as they spell conversational content differently. The tool's meaning belongs to Ensemble; the provider projection belongs at the protocol boundary.

A library Agent may have no tools. Its ordinary requests should then remain ordinary requests. The CLI installs the six initial operations as one application choice, rather than making every consumer inherit a shell and filesystem access.

In C#, a Tool is an owned object with a model-facing definition and a concrete operation. Its constructor registers it with the Agent. The implementation does not need an interface wrapper merely to connect that definition to its body. A small explicit dispatch over known operations is easier to follow than an extensibility framework built before an extension requires it.

There was an instructive testing trap here. A fake provider can volunteer a tool call whether or not the client advertised any tools. The resulting loop looks correct locally and does nothing against a real provider. Exercising the loop therefore does not establish that the request contains the declaration. A useful fake reproduces enough of the external contract to expose that distinction, and tests inspect the outgoing request as well as the returned result.

## The assistant proposes; the application executes

A response can contain text followed by one or more tool calls. The application records that response before performing the effects. Each call has an identity, a name, and input. Execution produces a result tied to the call's identity, and that result becomes input to another model request.

The call identity is essential when several operations are outstanding. A result named `read_file` cannot distinguish two reads of different files. Position is no better once another component observes or serializes the events. If a provider omits an identity, Ensemble assigns a deterministic distinct one while capturing the call, before logging or dispatch. Inventing it later during serialization would leave the execution path and the wire using different facts.

The loop continues until the model returns without asking for tools, or reaches a bounded round limit. A failed command is usually information for the next model round, not a reason to pretend the whole conversation succeeded or to crash the application. Unknown names and invalid arguments likewise become explicit tool failures without performing the requested effect.

Protocol details still matter. Anthropic's result-bearing messages require tool results before accompanying material. That ordering becomes especially important once hints can arrive during a tool call. The canonical result and its identity stay the same; the provider renderer owns the legal arrangement on the wire.

A short explanation before a tool batch gives the human a chance to understand the proposed action. Ensemble can require that explanation through explicit configuration. It does not infer the policy from a vendor's name. When enforcement is enabled, a calls-only response receives explanatory failures rather than execution. The next round can explain and try again. This is an observable application rule, not a claim that a model's private reasoning has been exposed.

## Six operations, useful edges

The initial tools cover reading, writing, targeted editing, listing, searching, and running a command. Their ordinary behavior is familiar. Their edges determine whether the assistant can use them safely and economically.

A read should expose enough location information to support a later edit. A search should return paths and matching lines rather than a decorative summary that requires another discovery pass. A directory listing should give a predictable view of what is present. The model pays for ambiguity with additional calls, and the user pays for those calls.

Writing and editing deserve different operations. Creating a file, appending to one, and replacing its contents have different consequences; they should not collapse into an accidental overwrite. Targeted editing takes an exact old anchor and a replacement. Zero matches means the model's view is stale or wrong. Multiple matches mean the request is ambiguous. Neither is permission to choose a location and hope.

This explains the measured preference for editing over writing. Replacing a small unique passage preserves unrelated work and avoids regenerating an entire file to change one line. Refusal is useful information: reread the relevant region and form a better anchor. A guessed line number would merely disguise uncertainty as precision.

The command runner introduces a broader effect boundary. Output and exit status both matter. A nonzero process exit is a normal tool result: the command ran, and its answer may be that the tests failed. Marking that as a tool-call error would invite the model to repair a call that was valid. The result preserves diagnostics and exit status, so successful execution of the tool cannot be mistaken for successful execution of the tests.

Working-directory changes apply to the particular invocation. A command must not quietly change the application's global directory and redirect later file tools. Fresh calls do not inherit an interactive shell's incidental state. A long-lived interactive process will eventually need its own explicit identity; it should not emerge accidentally from a global shell session.

The early filesystem restrictions provide a bounded workspace. The fuller sandbox chapter will examine canonical paths, symlinks, and the difference between filesystem policy and an operating-system security boundary. It would be misleading to promise all of those properties merely because these six tools accept a root directory.

## Teach the errors

The model's next action depends on what the tool says happened. A useful refusal identifies the violated condition without making up an effect. A missing file is different from a refused overwrite. A nonunique edit anchor is different from a malformed argument. An unknown tool is different from a command that ran and failed.

The same applies to malformed model output. The application should not execute a fragment just because it resembles a useful operation. Tool names, argument shapes, and call identities cross a trust boundary. Parsing them into a plausible object is the beginning of validation, not the end.

There is no benefit in adding elaborate recovery machinery before these simple distinctions work. A direct loop that reports a failure honestly gives the next model request a chance to repair it. A supposedly helpful wrapper that converts every exception into a vague success makes the next request less informed than the last.

## A new kind of silence

The six tools make the assistant capable of useful work. They also expose a limitation in the loop: control returns when the tool returns.

That is harmless for a short file read. It is much less harmless for a test suite, an interactive debugger, or an external tool that stops answering. The model can be perfectly capable of deciding what to do next and still have no opportunity to decide anything.

The next chapter gives work a lifetime of its own, so the conversation does not become its hostage.

*Implementation details: [codebook Chapter 3](../chapter-03.md).*
