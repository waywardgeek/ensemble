# Epilogue: What we can rebuild

Bill describes the origin of this work in a sentence: “I gave Astra a book. Astra gave me an AI coding agent.” The book carried a construction sequence, explanations, and checks. Executing it produced an application he could use. Work on a later implementation then took place inside the environment that earlier work had produced.

That is a more interesting result than a large pile of generated source. It also leaves a harder question. Which of the judgments that made the application useful can the next execution recover?

The C# work did not answer that by removing the expert. Bill corrected the architecture, challenged unnecessary mechanisms, resolved requirements, and insisted that the code remain teachable. Compilation could establish that partial declarations had matching implementations. It could not establish that the relationships were sensible, that the assistant had preserved one authoritative fact, or that the next reader could follow the cleanup.

Two late repairs show what that involvement was protecting.

## The input that was already there

The memory compressor's captured input already contained the existing long-term memory. A later step appended that memory again while preparing the worker request.

Taken separately, each function looked reasonable. Capture included the information needed to understand the conversation. Request construction made sure the summarizer had the current memory. Taken together, they sent the same material twice.

A local test of either function could have reinforced the mistake. The useful test used the real capture formatter and traced the complete request. Once the duplicated input was visible, the repair was to keep one authoritative formatting path, not introduce a policy for reconciling two copies.

The same mistake can occur without an LLM anywhere nearby. One layer adds a default, another layer helpfully adds it again, and neither knows which one owns the decision. Model requests make the cost easy to observe: the duplicate consumes context and money, and may change the emphasis of the input. The structural problem is older than the API.

The chapter needs to preserve the reason for the repair. “Include memory exactly once” is a useful assertion. Knowing that capture already owns its inclusion tells the next engineer where a new feature belongs.

## The task that was already done

Goals initially had local checks for file refresh, reminders, pause, interruption, and GUI controls. A delegated Agent could nevertheless publish its result and leave reminders enabled. Later, a timer could admit more work for a task whose parent had already received completion.

The missing boundary was successful submission. It was not enough to disable reminders whenever the model tried to submit: a malformed result or failed durable write leaves the task unfinished. It was not enough to wait for generic idle status: an idle Agent may still owe the parent an answer.

The correction stopped reminders only after durable publication succeeded. The actor retained responsibility for timer disposal. The goal file remained available, and the parent's configuration stayed unchanged. A follow-up check then exposed a stale status snapshot still showing the old deadline; the visible account of the task needed to agree with its new state too.

Each individual feature had a plausible implementation. Their composition lacked one important transition. The failure became understandable only by following the whole operation from the parent's request through the child's accepted result and into the later timer wake.

This is why ownership and lifecycle occupy so much of a book nominally about talking to language models. The network call is only one participant. The application also has to know what has ended, which obligations remain, and which object is allowed to make that decision.

## Preserve reasons, not a legend

A passing test is evidence about the path it exercised. It is not an architectural review, a provider agreement, or a prediction that the next model will cooperate. A local fake can show that cancellation reaches the transport. It cannot establish that a commercial endpoint accepts today's credentials. A speech transcript can expose missing words without showing that a listener can comfortably understand them.

The useful response is to keep the evidence attached to the claim. Historical observations stay historical. A current experiment reports the request that was admitted, the result that was measured, and the attempts that never reached measurement. A defect found by inspection is not promoted into a demonstrated runtime failure. An experiment that required expert correction is not presented as unattended success.

That discipline keeps the book open to improvement. It also prevents ordinary setup mistakes from acquiring the prestige of discoveries. Guessing a filename or a method signature and receiving an error consumes time; it need not reveal anything about the product. Reading the source before making the change is still cheaper.

The agentic codebook attempts to preserve enough knowledge to rebuild the system. This companion attempts to make that knowledge available to an engineer without requiring the engineer to execute every specification. Neither artifact makes judgment obsolete. They make the judgments visible enough to question and specific enough to test.

Calling generated source a secondary artifact does not lower its quality requirements. Regenerating tangled ownership or incomplete cleanup is not a useful kind of reproducibility. The result still has to be worth maintaining, and its explanations must remain accurate as it changes.

The next execution may use a better model, encounter a changed provider, or reveal an ambiguity nobody noticed. The durable contribution is a route back to the decisions: what was authoritative, what the alternative broke, what observation supported the repair, and what remained uncertain.

That is knowledge an engineer can carry into a different agent, or a different kind of program altogether.

*Related account: [codebook epilogue](../epilogue.md).*

Source: [Codebook epilogue](../epilogue.md).
