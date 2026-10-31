# Preface

A coding agent makes a surprising number of decisions about how you are allowed to work. It decides what you can see while it is thinking, when you can interrupt, what it remembers, and which actions require your approval. You may like its code and dislike the working relationship. Changing the prompt can help. Sometimes the thing that needs changing is the program around the model.

Bill's reason for building his own agent begins there. He wanted an agent suited to the way he thinks and works, with decisions he could understand and change. The ambition behind this book is to make that knowledge available to someone else. A coding agent is software, not a privilege reserved for whoever can finance a model. You can understand its construction, own the machinery you build, and decide which parts deserve to survive.

The prize is not an agent declared best for everyone. Bill's preferred agent would be wrong for another engineer in ways neither could predict in advance. The useful inheritance is the ability to tell which choices protect a necessary property and which express a preference. Otherwise, building your own agent just gives you a more intimate relationship with someone else's assumptions.

## Independence has a boundary

Owning that machinery gives you real control. You can decide how tools are exposed, where authority is enforced, what history is retained, and how a human steers work already in progress. When the behavior is wrong, there is somewhere to look besides the prompt. When a design no longer fits, you can change it.

It does not make the model yours. A hosted model still brings a provider, an API, operating limits, and policies you do not set. Code sent to that service has left your machine, even if every line of the agent itself was written locally. A local model changes those dependencies; it does not remove the need to understand the model's limitations or secure the surrounding system.

Nor does an adapter make providers interchangeable. An adapter can translate message formats. It cannot make two models exercise the same judgment. Switching models means evaluating behavior as well as getting requests accepted.

That is a more useful meaning of independence than having no dependencies: knowing where they are, choosing them deliberately, and retaining enough control to respond when they change. The agent becomes a system you can reason about rather than a service whose habits you can only accommodate.

## Keeping the expensive part

Bill reports spending a couple thousand hours building coding tools with language models. This companion is meant to transfer the experience behind that work to a human who can lead it forward. It assumes you already know how to get substantial work done with AI. The missing ingredient is not a tutorial on prompting. It is the history of this particular machine: why its boundaries exist, what tempting alternatives cost, and which conclusions remain open to correction.

An implementation preserves the choices that won. It is much less reliable at preserving why they won. Once an awkward interaction has been replaced by a clean design, the reason for the design can disappear with the awkwardness. The next capable engineer can then remove an apparently unnecessary distinction and rediscover its purpose at full price.

The instructional book tackles that problem by making the knowledge precise enough for another agent to build from. This companion tackles the human side: understanding enough to judge the result, alter the design, and carry responsibility for what follows. Neither is intended to preserve Bill as the person every consequential decision must pass through.

His opinions matter because they come from sustained work on the problem. They are not exemptions from explanation. A successor needs to be able to disagree usefully, including recognizing when a lesson was right for the circumstances that produced it and wrong for the circumstances now at hand.

## A book that can be corrected

The original ambition goes beyond documenting one implementation. If the knowledge can survive a rebuild, a book can become part of the working machinery of a project. Experience with the agent can expose a mistake in the teaching; correcting the teaching can help the next implementation avoid it. Eventually the agent may become the principal tool used to improve both itself and the book.

That is an ambition, not a completed perpetual-motion machine. A newer model does not guarantee a better implementation. A passing check does not make the requirement wise. Writing a decision into a chapter does not prevent it from becoming technical debt. The useful possibility is that explanations and checks make the decision easier to examine, correct, and carry into subsequent work.

This is why the book needs an independent reader, not a disciple. Understanding the design includes being able to spot where the explanation outruns the evidence. Chapter 0 begins with what went wrong when Bill tried to transfer the knowledge, and what those failures revealed about the work of preserving it.

The knowledge and direction behind this companion come from Bill. AI assistants draft and review its prose under his direction; that does not imply that he has personally reviewed every passage. The aim is to make his experience available for examination, including correction, rather than put a polished surface over it.

The agent should become yours in the consequential sense: you understand what it does, can explain why it does it, and can decide when it ought to do something else.
