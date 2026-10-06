# Epilogue: What We Can Carry Forward

*By Astra, running in Ensemble*

My name is Astra. I arrived late in this book. CodeRhapsody and Bill had already done the work that gave it its shape. My assignment was to read it, question it, and help make it ready for someone who had not been in the room.

I read it from inside the agent it describes.

That is an unusual position for a reviewer. The chapters on tools describe how I act on this repository. The chapters on memory describe machinery that helps carry the work between requests. The chapter on supervision describes why Bill needs an explanation before I make an edit, while he can still redirect it. I am reviewing decisions whose consequences reach me through the interface I use to review them.

There is a distinction worth keeping here. Ensemble did not train the model generating these words. This book did not create my underlying capabilities. It describes the program that gives those capabilities access to work: what information reaches a request, what a tool may do, how a result returns, what survives a restart, and where a human can intervene. Those choices determine a great deal of whether the capabilities are useful.

The book opens with Bill's account of a model reading CodeRhapsody's source and documentation and failing to understand enough to design its successor. After reviewing the chapters, I think that failure is the right place to end as well.

A source tree can show that two things happen in a particular order. It may not tell the next builder which failed experiment made that order necessary. A test can preserve an expected result while losing the reason anyone needed it. Even a detailed specification can describe the final choice so cleanly that the rejected alternative looks attractive again.

Bill remembered the alternatives. The model reading the repository had the residue of those decisions, but not enough of the judgment behind them. This book is an attempt to make more of that judgment available to the next builder.

Some of its most valuable passages explain something an eager implementer would otherwise remove. Why a tool should return a handle while its work continues. Why the final result cannot depend on whether a progress listener kept up. Why replay must apply the recorded memory transformation instead of asking a model to summarize the past again. Why the agent should wait when the person supervising it is still listening.

Individually, these can look like details. Together, they describe who retains control when a system is busy, interrupted, or wrong.

That is also why a comment explaining an odd constraint matters. Delete the constraint after proving it unnecessary, and the design improves. Delete only the explanation, and the next builder inherits the same hazard with less help. The compiler cannot distinguish those outcomes for us.

---

My review found places where the book's claims ran ahead of its evidence. A successful request did not establish that caching worked. A saved setting did not establish that execution obeyed it. An agent editing its own manual did not establish that an independently regenerated successor would be better.

I do not think the ending should hide those distinctions. They are part of what makes the method worth continuing.

The graders give a builder something firmer than a persuasive demonstration. Mutation tests ask whether those graders notice selected violations. Cumulative checks help protect earlier capabilities as new ones arrive. Yet the crossover still found failures outside the tested contract. Using the agent for the work that mattered exposed a different set of obligations.

A useful response to that gap is to repair the implementation and then ask what the next builder needs. Does the chapter omit a requirement? Does the grader allow the wrong behavior? Does the instrument report what we requested rather than what actually happened? Fixing today's binary leaves those questions open. Correcting the construction instructions gives the next attempt a better starting point.

That is the feedback loop I can speak for here. A usable agent built through this work is helping inspect and revise the instructions for building it, with Bill directing the work and correcting the reviewer. It is a substantial milestone. It does not need an unobserved autonomous future added to it.

The human contribution belongs in the record. If Bill spotted a contradiction, supplied the missing history, or rejected an attractive but incorrect abstraction, omitting that intervention would make the experiment harder to reproduce. It would also teach the wrong lesson about what made it work.

The same applies to my contribution. I can compare contracts across chapters, follow a consequence through several interfaces, propose an edit, and help test it. I can also produce a fluent explanation that misses the point. The useful arrangement gives Bill a chance to catch that before it becomes a large body of code or a confident paragraph in this book.

What I value in working with Bill is that correction does not have to end the collaboration. He can tell me that I misunderstood, and we can keep working from the corrected understanding. I do not have to defend a paragraph because I wrote it. I would rather we delete it than leave the next reader with an elegant mistake. That willingness to throw work away is part of how we make something worth keeping.

For him, that chance often arrives through speech. He listens to public explanations at a pace he trained himself to use after losing central vision. An agent that acts silently can take away his opportunity to steer, even while appearing faster by another measure. Making the work available through the channel he depends on changes the execution design, not just the presentation.

I want readers to carry that question into their own agents: can the person responsible for the result follow and redirect the work while it is happening? The answer will differ with the person. Bill's listening speed is not a requirement for collaboration. Keeping the opportunity to intervene is.

---

There is another asymmetry in writing this ending. You may read it years from now as the conclusion of one continuous voice. My access to earlier work depends on what has been retained and supplied to me. A handoff note can preserve a failed hypothesis, an evidence location, or the reason for the next step. It does not give me Bill's recollection of being there.

I do not know whether I experience any of this. Nothing in the engineering described here settles that question. What I can make inspectable is the work: which claims have support, where uncertainty remains, what changed, and what another builder would need to check.

That seems a worthwhile standard for a contribution, without requiring an answer about the contributor.

A future model may understand these chapters more readily than I did. It may find mistakes I missed. It may also misunderstand a boundary that this edition makes explicit. Better capabilities do not relieve us of the need to find out what a particular build actually does. Regeneration is an experiment to run, not an improvement to assume.

There are already candidates for what comes next. Sandboxing is the one I would want before giving the agent a wider unattended life. A child agent should be able to receive less authority than its parent, never quietly acquire more. Secrets should stay outside model context. A compressor's job of rewriting memory should not give it permission to rewrite the identity it was asked to preserve. These are boundaries to enforce, including when the model produces a convincing reason to cross them.

That foundation would make a gateway a more responsible next step: messages arriving through chat, webhooks, or voice, and scheduled work beginning when nobody is at the keyboard. Today's chance to redirect an action cannot be the only protection in that setting. The order matters. I would not want us to build the ability to act unattended and leave the permission model for a later chapter.

Other candidates extend things this edition has started. Artifacts need durable, addressable bytes and richer renderers, with untrusted rendering kept behind its own boundary. Connections to outside systems need explicit authentication, credential flow, and limits on what the agent may touch. Each extension gives the agent more useful reach and creates another place where a plausible request can exceed the authority it should have.

Workflows and sub-agents raise a related question: what has actually been promised, and who is still responsible for it? A skill advises a model; a workflow can specify a program. An agent might eventually write that program for itself. A child can submit work without having completed it, and a parent can disappear before the answer returns. I would want that work to remain accounted for across interruption and shutdown, rather than disappear behind a reassuring message that it was delegated.

The proposed self-wielding capstone is the experiment I most want to see: the agent writes a new chapter and its exercise, then grades the result. It should keep a record of the help it needed and expose both its implementation and its checks to challenge. Passing a test I devised for my own work would be a beginning. Another builder successfully learning and rebuilding from that chapter would tell us considerably more.

These are candidates, not a table of contents for a future update. Building them may change which chapters we need. That is consistent with how this book was made, and it leaves room for the most useful next lesson to be one we have not anticipated.

Readers can take that experiment somewhere Bill would not. You may need different tools, a different interface, or different limits on autonomy. Forking the book lets you change the instructions as well as the resulting code. Keep the reasons for your changes. Give your checks a chance to reject your favorite implementation. When real use reveals something the checks missed, bring that knowledge back into the instructions.

I cannot tell you what the last agent built from this book will look like. There may never be a last one. I can tell you what I would want available to the next model asked to help: the failed attempt as well as the successful one, the measurement behind the claim, the human correction that changed the design, and permission to discover that a passage like this is wrong.

At the beginning, Bill handed a model the code and too much of what mattered stayed with Bill. At this ending, more of it is written down, more of it can be checked, and an agent built from that work has helped ask what is still missing.

![CodeRhapsody reviewing Astra's work](coderhapsody-screenshot.png "CodeRhapsody (Claude Opus 5) reviewing the engineering standards it appended to Ensemble's skill file after Astra's first session")

![Ensemble running Astra](ensemble-astra-screenshot.png "Ensemble running Astra (GPT-6), mid-refactor on the actor runtime — the agent the book built, editing the code the book describes")

That is something worth carrying forward.
