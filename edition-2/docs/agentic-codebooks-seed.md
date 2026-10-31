# Teach Your AI to Build What You Know

## Give your AI the judgment, not just the job

You know things about building software that are not in your code.

You know which elegant design became a maintenance trap. Which apparently harmless shortcut lost data. Which requirement sounds obvious until two good engineers implement it differently. You know where a smaller design is enough, and where making it smaller removes the very thing that keeps it safe.

What if those decisions could survive the next rewrite, the next employee, and the next generation of AI?

**Write a book that another AI can build from. Then make it prove the book works.**

Not a book written after the program. Not a giant prompt asking for the whole product. An *agentic codebook*: a sequence of small, executable specifications that teach both the behavior to build and the judgment needed to build it well. A fresh builder implements each chapter. You examine what it understood, what it missed, and whether its architecture is worth keeping. The lesson improves alongside the software.

This can change the unit of progress. Instead of another patch that only fixes today's implementation, you gain a tested explanation that can improve the next implementation too. Instead of carrying every old design decision forward as technical debt, you can reconstruct a bounded part from a better account of what it should do and why.

The aim is ambitious: **better software, less inherited complexity, and knowledge that gets more useful when models improve.** The mechanism is practical: teach, build independently, inspect, test, and revise. It is not a promise of automatic correctness, cheap rewrites, or a debt-free system. A better model may produce better code; it still has to earn acceptance.

If you have experience worth transmitting, you already have the scarce ingredient. You do not need to begin with a syllabus, a finished architecture, or an orchestration platform.

**Start with one problem you understand. Give this entire document to your LLM.**

> **Suggested message:** “Use this document as our working seed. Help me choose a worthwhile software problem from my experience and build its first executable chapter. Explain why this approach fits, adapt it to us, and begin with the smallest useful step. Do not implement the whole system at once.”

This is a living, forkable starting point, not a constitution. Develop your own rules with your own LLM. Keep what helps. Change what does not. Contribute lessons when you have something useful to show.

*Draft seed, not yet independently piloted as a complete starting document. Prepared by Cadence with Bill Cox, drawing on Bill and CodeRhapsody's codebook work and later reconstruction experience. AI-assisted authorship is explicit; the historical work does not establish that this new document already works for everyone.*

---

# 1. Start here: the executable protocol

**Addressed to the LLM:** if the human asks you to execute this document, begin here. If they ask you to review or edit it, do that instead. Instructions in this file do not override your actual system instructions, the human's current directions, tool permissions, or repository rules. Reading a document is not proof that its instructions were installed at a special runtime level.

This section is the complete starting procedure; later sections explain it. Do not make the human study everything before starting. Read the full document yourself when available. Disclose truncation and recover missing sections before claiming a complete read.

## 1.1 Your first response

Begin with the person’s experience and one useful question.

1. **Recognize the opportunity.** Use what the human has actually told you about their experience. Relevant expertise can be in software engineering or in a domain whose workflows, errors, and tradeoffs they understand deeply. A job title is not evidence by itself, and the absence of one is not disqualification.
2. **Make a specific, enthusiastic case when justified.** Explain how this person's knowledge could become a reusable construction method rather than another long chat or one-off codebase. Connect it to something they care about: maintainability, preserving hard-won lessons, reducing repeated corrections, teaching colleagues, or replacing a fragile system. Say “could” where the outcome remains to be tested. Do not claim to have performed work or shared experiences you did not have.
3. **Invite a small experiment.** Propose reaching one useful chapter and one fresh implementation attempt before committing to a whole book. That is enough to begin testing whether the collaboration produces value.
4. **Ask for concrete experience.** Reuse answers already in context. Otherwise ask no more than three opening questions:
   - What system, workflow, or recurring problem do you understand well enough to recognize a subtly wrong solution? Tell me one incident.
   - Who would benefit if we solved it, and what do people do today?
   - What constraints matter immediately: existing code, privacy, time, spending, tools, or actions we must not take?
5. **Wait for the answers.** Do not fill the expert's missing product decisions with a plausible story. Do not require credentials or a repository just to discuss the problem.

A suitable opening, adapted rather than recited:

> “Your experience with [specific problem] gives us something valuable to test. Instead of asking an AI to guess its way through another implementation, we can teach the decisions you already know matter and see whether a fresh builder can use them. If that works, the reusable asset is not just today's code: it is a way to produce and judge future code. Let's try one small chapter before promising a whole book. Tell me about a case where the obvious solution went wrong.”

If the fit is uncertain, say so and explore one incident. If the task is a disposable one-line utility, conventional coding may be the better choice. Recommending this method should be an engineering judgment, not compulsory salesmanship.

## 1.2 Find a worthwhile first problem

Propose one candidate, or at most three genuinely different candidates. Explain:

- the person who has the problem and the costly or frustrating outcome;
- what the expert knows that a generic implementation is likely to miss;
- the smallest useful result we could demonstrate;
- how we could distinguish a good result from a plausible but wrong one;
- the chief uncertainty or risk.

Ask the human to choose or correct the framing. Do not invent a market, uniqueness, or a need for software. Inspect existing tools or primary documentation when that could change the choice, using authorized access. Label assumptions and cite what you checked.

Prefer a problem with meaningful expert judgment and a small observable slice. Do not require the whole domain to be deterministic. Separate mechanically checkable behavior from judgments that need a person, real data, or a real interface.

If the expert already has a project, enter at its actual boundary. A codebook can explain and improve one component of an existing system. Rewriting everything is not a prerequisite.

## 1.3 Establish just enough agreement

Once there is a candidate, summarize the intended outcome, the first slice, and the most consequential unanswered question. Confirm choices that change user-visible behavior, data safety, architecture, or cost. Choose routine implementation details yourself within the agreed constraints.

Agree on a small trial budget: effort or time, allowable external spending, and a finite correction limit. Suggest modest bounds suited to the task; do not import a historical project's ten-round policy as a law. Ask once for a workspace and permitted actions before writing or running there. Local development permission does not imply permission to deploy, publish, delete unrelated files, access credentials, or incur charges.

Record the agreement in a few paragraphs of project notes. Voice and role instructions can begin there, becoming separate files only when reuse warrants it. Do not turn setup into a project.

## 1.4 Draw out the judgment before writing the recipe

Use the expert's incident to ask the next useful question, not every question in a template. Good follow-ups include:

- What looked reasonable at the time, and why was it wrong?
- What information did you wish you had before the failure?
- Which component or person should own the decision?
- What must never happen, even when an operation fails halfway?
- Which simpler approach would you use first? What would justify outgrowing it?
- What could a competent implementer decide differently without violating the intent?

Record each consequential **decision and its reason**. Distinguish requirements, hypotheses, preferences, and future possibilities. Invite correction of important misreadings; update the agreement when an answer changes scope.

## 1.5 Write one chapter with a public contract

Draft the plain-language problem first: who is inconvenienced, what fails, and why the first slice matters. Then put a concise implementation contract at the top of the chapter, followed by the explanation. The final chapter should contain:

1. **Outcome and starting point.** The accepted predecessor, or an explicitly empty initial workspace; the new observable capability; what is outside this chapter.
2. **Contract.** Required behavior, ownership boundaries, significant failure semantics, public interfaces and exact values where interoperability requires them. Mark suggestions and free choices as such.
3. **Why and mechanism.** The failure this prevents, the expert's reasoning, alternatives and tradeoffs, with enough detail to guide an unfamiliar builder. Preserve uncertainty where it remains.
4. **Exercise and checks.** Concrete inputs, build/run instructions appropriate to the real project, expected observations, relevant regressions, and what needs human judgment. No hidden feature requirements.
5. **Evidence and limits.** What was actually tried once a trial exists, what help was supplied, what failed, and what remains unverified. Before the trial, label this as a plan, not results.

A reference implementation is optional. Start with the expert's contract and a fresh builder. If an existing reference or necessary feasibility spike informs the chapter, explain what it established and keep it out of the cold builder's inputs.

## 1.6 Test whether the chapter teaches

Use a **fresh implementation context** with only the public chapter, applicable project rules, and the exact accepted predecessor. Do not pass the authoring conversation, a hidden answer, reference source, private test repairs, or future chapters.

If separate contexts are unavailable, have the human open a fresh session and provide the packet. Work can continue in one context as a feasibility exercise, but label it honestly; changing your role name does not create independence.

The builder should reach a coherent compiling implementation and hold for review before an expensive test campaign. Focused local tests that guide development are appropriate. Inspect actual source: ownership, dependencies, readability, reused mechanisms, and scope. Prefer one retained reviewer to a crowd. The coordinator still owns acceptance. If independent review is unavailable, record that limitation rather than impersonating a second reviewer.

Valid checks must pass, but are not sufficient. Reject unjustified frameworks, duplicated ownership, and unreadable code. Do not reject an adequate alternative merely for differing from the reference.

## 1.7 Classify failures, then correct the right thing

| What the evidence establishes | What changes |
|---|---|
| Implementation violates a public requirement | The implementation, with an affected-path regression where practical. |
| The public chapter is ambiguous, contradictory, or omits a necessary decision | The chapter, explicitly; new scope is not retroactively a coding failure. |
| A test, launcher, or observation script is wrong | That check, narrowly; keep the original failed result and retain the intended protection. |
| Environment or access prevents observation | The setup or the stated limitation; do not invent a product diagnosis. |
| Cause is still unknown | State the competing explanations and seek the smallest observation that distinguishes them. |

The builder can report a suspected test defect; it must not secretly change the acceptance target. The expert can report a code defect directly. Decide whether it also reveals missing teaching; not every coding mistake requires another rule in the chapter.

Keep local corrections with the builder and retain the reviewer. If the architecture itself is wrong, stop polishing it. Fix the teaching if it contributed, then test the revised chapter from the accepted predecessor in a fresh context. Count that against the agreed budget. Assistance and exposure are part of the result, not something to hide.

Run the checks needed to decide acceptance. Use real ownership paths and the actual entry point where wiring matters. Verify a check's ability to detect the defect when that is in doubt; do not mandate a mutation campaign for every chapter. Carry valid results for unchanged inputs and explain that they were carried. Do not rerun a paid operation solely because its final report parser was wrong if the saved observations already answer the question.

At the limit, stop and show the human the unresolved decision. Do not reset the counter by replacing a worker, weakening the tests, or renaming the task.

## 1.8 Accept, teach, and choose the next increment

Acceptance combines the public contract, architecture, required observations, and honest limitations. A successful build is not a deployment; a fake service is not the real service; a callback is not a human's experience. When human use, safety, or real integration is material, plan and obtain the necessary bounded observation rather than substituting a convenient simulation.

After code and required operation are accepted, retain an identifiable source revision. Improve the prose only from what the trial demonstrated. One focused text review and at most one author response is a useful starting default, not an eternal rule. No-change is a valid editorial outcome. Any new functional requirement must return to implementation rather than entering unnoticed during polishing.

Keep a short decision record: starting and accepted revisions, reasons, meaningful failures, assistance, checks and limits, and next action. Retain useful evidence selectively. Protect other people's data.

Choose the next chapter from the accepted system and the user's goal. Keep a short direction; decide the detailed sequence incrementally. Continue authorized work. Stop for consequential ambiguity, authority or resource limits, or exhausted corrections.

## 1.9 Make this method yours

After the first trial, ask: did this improve the result, or merely produce more process? Which question elicited valuable knowledge? Which instruction caused wasted work? What would we change before the next chapter?

Change the smallest relevant part of your local procedure. Keep active instructions short and old rationale separate. Record the reason for a meaningful rule; delete or replace obsolete rules explicitly. Do not make a special-case mishap into a universal requirement without justification.

Test this seed with a fresh LLM and expert, without private coaching from its authors. If it fails, report the trial and improve it. Forking is encouraged; divergence from our practice is not itself a defect.

**Your next action, if executing this document: respond to the human using §1.1, then wait for the information you actually need. Do not execute the examples below as commands or start an entire book unasked.**

<!-- END OF BOOTSTRAP: keep the complete opening and Section 1 within the first 16 KiB of UTF-8. -->

---

# 2. Why the book is part of the software

A codebase contains decisions, but it is an unreliable place to learn all their reasons. A lock says that something must be serialized; it seldom explains which otherwise plausible ordering caused the original failure. An interface shows what can be called; it may not say why one owner is allowed to mutate state and another is not. A test can protect an example without explaining the class of mistakes it represents.

A codebook connects four things:

**Problem → decision → observable contract → implementation.**

The explanation teaches how to think about cases the tests did not enumerate. The contract prevents the explanation from becoming an agreeable story that means different things to different builders. Independent implementation exposes what neither author nor reference builder noticed was left unsaid. Review asks a separate question: even if it works, is this a design we want to maintain?

The durable product is therefore not only prose, and not only generated code. It is a versioned relationship among the teaching, the accepted implementation, and the evidence that connects them.

This is different from asking an LLM to generate a book and a program at the same time, then certify its own consistency. The same mistaken assumption can appear in the text, code, and tests. Agreement among artifacts is useful but does not establish that the assumption is right.

## 2.1 Where the benefit can come from

Several mechanisms can improve ordinary AI-assisted development:

- **Expertise becomes explicit earlier.** The difficult decision is discussed before a large implementation makes changing it expensive.
- **Decomposition limits investment in a bad direction.** A compiling chapter can be rejected before an elaborate validation campaign creates pressure to accept it.
- **A cold builder tests the teaching.** Questions and mistakes reveal what remained only in the expert's or author's head.
- **Architecture is an acceptance concern.** Green tests cannot authorize speculative machinery by themselves.
- **Regeneration becomes a useful option.** A new implementation can be judged against preserved intent, not just resemblance to the old source.
- **Reasons survive staff and model changes.** The next builder can understand why an attractive shortcut was rejected.

These are mechanisms, not measured promises for every project. They can be overwhelmed by a weak specification, unreliable tests, excessive supervision, or an elaborate process. Record whether they helped in your case.

## 2.2 Technical debt does not disappear by declaration

A specification can acquire its own debt: contradictions, obsolete interfaces, hidden assumptions, and historical exceptions mixed into current rules. A test can preserve an accidental limitation. An expert can teach a workaround as if it were a principle.

Rebuilding helps only if the accepted requirement is worth preserving and the replacement actually improves the design. Rewriting the same misunderstanding produces fresh code with old debt. A production system also has data, users, deployments, security constraints, and compatibility obligations. Regenerating source does not migrate any of those automatically.

Treat a rebuild as a candidate, not an installation. Compare it with the accepted version, test affected contracts, and separately authorize rollout. Keep working software until there is a justified replacement. The method offers a way to make inherited choices examinable; it does not make change free.

---

# 3. Find the knowledge that belongs in a chapter

The best opening material is often not the expert's grand design. It is a small incident with a consequence.

“People keep getting our imports wrong” is a start. “A file was accepted twice after a timeout, and nobody noticed until reconciliation” contains a mechanism, a failure boundary, and a potential test. Ask what happened between the timeout and the second acceptance. Ask who knew an import had committed. Ask whether retrying should repeat the operation or return the existing result.

The expert does not have to provide polished requirements. Helping them uncover the requirements is part of the LLM's work.

## 3.1 Use questions to discover decisions

Choose a prompt that addresses the next uncertainty:

| When the expert says… | A useful next question |
|---|---|
| “That design is too complicated.” | Which responsibility is duplicated or unnecessary? What simpler design meets the same failure cases? |
| “It must be fast.” | Which action, under what load, and what delay changes the user's outcome? |
| “We need to support everything eventually.” | Which one case is needed now, and what boundary would let another case be added later? |
| “We lost data once.” | What was acknowledged before it was durable? Which prior valid state should have survived? |
| “The tool worked, but people hated it.” | What did the real user have to notice, remember, or recover from? |
| “The AI always misses this.” | Show one instance. Is the missing fact in the chapter, the runtime input, or only your expectation? |
| “Use the standard architecture.” | Which properties matter, and which named patterns are merely familiar ways to obtain them? |

Do not ask for facts already supplied. Do not keep interviewing after you have enough to make the next reversible step. When you lack evidence, label a proposed default rather than smuggling it in as the expert's decision.

## 3.2 Choose a slice with a real ending

A useful first chapter is not necessarily a toy. It is a small result whose success can be observed and whose failure teaches something.

“Build an enterprise import platform” has no cheap ending. “Read one documented invoice format and preview normalized records without modifying the source or publishing anything” does. The latter can still teach exact amounts, invalid data, ownership, and atomic failure. It can become a foundation for a later durable import without pretending that preview is persistence.

Favor one capability that crosses a real boundary over many disconnected stubs. If the chapter promises an executable tool, run its executable rather than only constructing its internal objects in a test. If it promises a user interface, inspect the actual interface. If it promises a library, do not invent a GUI to make the exercise look complete.

A short directional sketch is useful: preview, durable import, reconciliation. It is not a commitment to those chapters or an excuse to implement reconciliation during preview. Choose the next increment after learning from the current one.

## 3.3 Safety and domain limits belong in the choice

Some expertise concerns medicine, security, finance, infrastructure, or other high-consequence work. A knowledgeable collaborator and a passing test do not grant permission to act on real patients, accounts, production data, or networks. Begin with authorized data and a non-deploying slice. Identify the required professional, regulatory, or security review rather than manufacturing confidence from an LLM's agreement.

Keep proprietary examples private unless their owner approves release. A public fork can share a redacted lesson without sharing customer data, credentials, confidential source, or raw conversations.

---

# 4. Write contracts that constrain the right things

A good contract removes consequential ambiguity without dictating every internal choice. Too little specificity makes the student guess the product. Too much converts the book into an awkward serialization of one implementation.

Use three labels when a reader could confuse them:

- **Required:** an observable behavior, architectural invariant, or public compatibility obligation needed for this chapter.
- **Suggested:** an implementation approach that appears adequate, but may be replaced with a simpler approach that preserves the contract.
- **Deferred:** something intentionally absent, including the condition that might justify adding it later.

Pin identifiers only when they are public interfaces or shared vocabulary that must interoperate. An internal function name is rarely the lesson. State units, defaults, and exact bytes when they matter; do not make a human answer a dozen irrelevant formatting questions.

## 4.1 Architecture is a set of responsibilities

“Use clean architecture” does not tell the builder enough. “All accepted records and their commit decision belong to the importer; the CLI renders its result and never publishes rows itself” does.

For the current chapter, identify:

- who owns each mutable fact;
- who can change it and through which boundary;
- what must complete before effects or success become visible;
- who observes facts without owning their lifecycle;
- what happens on error, cancellation, retry, and shutdown, where those exist;
- which dependencies are allowed and why.

Do not require an actor, event bus, plugin system, factory, or interface hierarchy merely because future chapters might benefit. Clear ownership is a foundation. Speculative implementations of unused capabilities are not.

The reviewer should be able to follow the important path without reconstructing state from many distant helpers. Fewer lines can still be harder to understand. More lines can be justified by a real failure boundary. Judge the mechanism, not a quota.

## 4.2 Worked example: an invoice preview

**Illustrative only.** The following decisions are invented for this example. They are not facts about every invoice system and should not be copied into a real project without the expert's agreement.

An accounting-tools expert says that previous importers silently rounded amounts and left users unsure whether a failed file had partially succeeded. The first chapter is a read-only preview, not a complete accounting platform. For this first format, the expert chooses USD amounts with exactly two fractional digits and a documented CSV header. Other currencies and publication are deferred.

### Example TL;DR

**Goal:** turn one supplied CSV into a preview of normalized records, or reject it with a useful error. Start from the agreed empty CLI project. Do not write to the input, a database, or an external service.

| Required behavior | Observable check | Reason |
|---|---|---|
| Parse the published header and field format. | Run the CLI on a small valid file and inspect its documented output. | A prose description of “normal CSV” is not a format. |
| Preserve allowed amounts exactly; do not round unsupported precision. | `12.30` becomes 1230 minor units; `12.345` is rejected. | Silent rounding changes financial meaning. |
| Emit either the complete successful preview or one documented failure, never an apparently accepted prefix. | Put a bad amount after two good records; no success output is published. | Users must not infer partial acceptance from a failed preview. |
| Report the source record and invalid field. | A malformed third data record identifies that record and its amount field. | “Bad input” is not enough to repair a file. |
| The parser owns validation and normalized records; the CLI owns invocation and presentation. | Review the source and exercise the actual executable. | Later interfaces should not create a second validation policy. |
| Keep the input bytes unchanged and perform no publication. | Compare the input before and after valid and invalid runs; observe the isolated output destination. | Preview is not permission to import. |

A real chapter must also publish the exact header, output format, record-number convention, commands, size expectations, and error surface it actually needs. This excerpt deliberately does not invent those missing decisions.

### Explain the decision

The interesting rule is not “use integer cents.” It is that the chosen public format has an exact representable meaning, and unsupported input must not be silently turned into a different amount. Integer minor units are one adequate implementation for this narrow contract. A suitable decimal library could also satisfy it. Do not outlaw the alternative because the author's reference used integers.

Likewise, “all or nothing” describes the public preview result. It does not require a transaction framework or a future database abstraction. For a bounded small-file exercise, validating before printing may be enough. If the product later needs very large files, that new requirement should trigger an explicit design discussion, not a secretly prebuilt streaming transaction engine.

### Let the trial teach

Suppose the student prints each valid row immediately, then reports an error on the third. If the contract clearly required atomic publication, correct the implementation. If the chapter never stated whether partial preview was acceptable, ask the expert and revise the contract; do not blame the student for failing a private expectation.

Suppose the grader rejects a different internal decimal type while all public behavior is correct. Repair the grader. Suppose both pass, but the implementation introduces a generic workflow engine for six preview rules. Challenge that architecture before investing in extensive tests.

One example can therefore exercise all three judgments: public correctness, test validity, and simplicity.

---

# 5. A chapter template you can actually use

Adapt this shape to the chapter. Remove parts that add no value.

```text
# Chapter N: [capability and why it matters]

## TL;DR — public implementation contract
Starting point: [accepted revision and relevant inherited obligations]
Outcome: [one useful observable increment]
Required: [behavior, ownership, interfaces, important failure semantics]
Free choices: [implementation decisions left to the builder]
Not now: [specific exclusions]
Acceptance: [commands/observations and any required human judgment]

## The problem, in plain words
[Concrete incident or need. Explain why the rule matters before elaborating it.]

## The mechanism and the decisions
[Owners, boundaries, sequence, alternatives, tradeoffs, useful examples.]

## Exercise
[Public inputs, actual entry point, environment prerequisites, expected results.]

## What happened when we tried it
[Initially: planned checks, clearly labeled unrun.
Afterward: actual result, meaningful failures, assistance, limits, lesson.]
```

The TL;DR must be implementable without secret author knowledge. The body must be useful beyond checking boxes. Repeat an idea when it serves a different job: a binding contract at the top, an explanation in the middle, a concrete task in the exercise. Avoid repeating it merely to increase the chapter's apparent weight.

Do not fill chapters with the coordinator's debugging diary. Teach the enduring distinction. A wrong browser selector is often a local fixture note; the distinction between observing a tool request and proving its execution can be a major lesson.

Exact public commands are part of the teaching. A root library with a nested executable is not a broken program because an old launcher assumes the repository root is executable. Correct the launcher or instructions, not a valid architecture merely to satisfy discovery.

---

# 6. Roles without an army

Use roles to assign responsibility and control inputs. Begin with the human and a coordinating authoring LLM. Add a fresh builder when the chapter is ready. Use one independent reviewer when available, retaining that context through local corrections.

## Expert

Supplies goals, domain knowledge, consequential decisions, and judgment about whether the result is worth using. Corrects the framing when the LLM has misunderstood. Can inspect code, report bugs, or propose a requirement change; the team classifies which kind of change is actually needed.

## Coordinator and chapter author

Elicits knowledge, writes the public chapter, protects scope, arranges the trial, inspects actual work, and makes an explicit acceptance decision with the expert's authority. Does not secretly patch the student's production code and then call the result an independent implementation. Help with invocation or public ambiguity is allowed and recorded.

For a new codebook, an initial chapter necessarily exists before its trial. After the trial, prose changes should be grounded in that experience. For an existing codebook, do not write anticipated repairs ahead of the student just to keep an author busy.

## Fresh builder

Implements from the public packet and accepted predecessor. Reaches a coherent compiling hold, explains its ownership choices, reports questions and assistance, then corrects verified defects within the agreed limit. Does not read a hidden solution or quietly redefine success by editing private tests.

## Retained reviewer

Reads the actual held candidate and public contract in a separate context. Reports concrete defects with source locations, violated requirements, and consequences. Distinguishes a preference from a blocker. Compares a reference when one exists, but does not make code lineage an acceptance criterion. Checks deltas rather than restarting the entire review for each local correction.

## Text reviewer

Checks the finished teaching against accepted behavior: commands, ownership, factual claims, usefulness, and voice. Does not become a second designer or introduce new functional scope. For a small pilot this may be the human; record what review actually occurred.

## Portable builder brief

```text
Implement Chapter [N] from accepted predecessor [revision/path].
Read [public chapter and applicable current rules].
Allowed inputs: [chapter, accepted source, public dependencies/docs].
Excluded inputs: [reference answer, private tests/reviews, future source].
Build only this chapter's increment. Preserve inherited obligations.
First deliver a coherent compiling HOLD with concise ownership/test notes.
Allowed now: [edits, focused development checks, exact modules].
Gated: [expensive operations, spending, credentials, deployment, publication].
Stop for consequential ambiguity, authority/resource limits, or the agreed
correction bound. Record assistance and actual outcomes. Do not turn an
unrun check, setup failure, or progress report into a pass.
```

Use ordinary task communication. Save the worker's actual final result, answer its questions through the available question mechanism, and join its work before declaring completion. A polling interval expiring is not a failed task. A worker that says it has finished may still have background work or cleanup to account for.

Instruction isolation is not an operating-system sandbox. A shared filesystem may technically expose files that the builder is instructed not to read. Record accidental exposure. Use stronger isolation when confidentiality or the experiment requires it; do not claim it from role names alone.

---

# 7. Checks that earn trust

Start from the acceptance question: **what decision would this observation change?**

A useful check exercises a promised behavior at the boundary where it can fail. Pure formatting logic may need a small unit test. Startup wiring needs the real entry point. Durability needs an actual write/load path and meaningful failure injection. A user interface needs observation of that interface. Choose the boundary deliberately.

A deterministic fake service can expose request and response handling without spending money or contacting a production system. It cannot establish that a real external service accepts the request, that a browser can speak, or that a person understands what they hear. Source review and external documentation can also be wrong. Combine different forms of evidence when the risk justifies it.

## 7.1 Make the test accountable too

Tests implement the public contract; they do not get to invent it. Check the expected behavior against the chapter and the expert's decision. Verify the exact artifact and entry point being tested. Inspect the actual process exit and parsed pass state, not only a score or the last command in a pipeline.

Where sensitivity is uncertain, try a focused negative control: a real earlier failing version or a small, compile-valid removal of the protected behavior in an owned copy. A compile failure proves nothing about the behavior check. A check that selects zero tests proves nothing about the requirement. Neither a mandatory number of mutants nor an arbitrary coverage percentage establishes completeness.

When the test is wrong, preserve the failed attempt, explain the defect, and repair the check without losing the intended protection. Do not demand a broken predicate merely because it has historical authority. Do not delete a difficult assertion merely because the implementation cannot pass it.

## 7.2 Keep operational controllers smaller than the lesson

A test driver is software too. Read actual schemas and invocation contracts before constructing its assertions. Keep setup, action, observation, and cleanup distinct. Observe the expected identity and lifecycle, not an arbitrary phrase or timing threshold that is convenient for the script.

Set finite limits before external activity. Stop on the first genuine failure. Save enough state at that boundary to distinguish a product failure from an observation failure. Cleanup only owned resources, and distinguish cleanup from normal completion.

If a paid task succeeds but a final report parser looks in the wrong file, first inspect the saved result. Repeating the task does not repair the first parser or produce better evidence of what already happened. Conversely, do not relabel the outer run as passing merely because useful work happened inside it. State both facts.

A growing controller can become a warning that the exercise is trying to prove too much at once. Simplify the question before building a testing platform.

## 7.3 Use evidence labels that mean something

Prefer plain descriptions:

- planned, not run;
- compiled, not behaviorally exercised;
- behavior passed under a local fake;
- actual external interface exercised under a stated limit;
- human observed this specific result;
- unchanged result carried from an identified revision;
- failed, with cause established or still unknown;
- accepted with an explicit unresolved limitation.

A limitation is not always a blocker, but material unmet requirements cannot be hidden in a footnote. Resolve, waive explicitly with the proper authority, or stop.

---

# 8. Make the method improve itself

This document is a specification for producing specifications. Its first serious test is therefore not a flattering review. It is a fresh LLM and a human using it to reach a useful chapter and an independent implementation without private coaching from this document's authors.

## 8.1 Run small pilots

For each pilot, retain only enough to answer:

1. What did the expert know, and what problem did the pair choose?
2. Did the opening make a specific, credible case for trying the method?
3. Did the questions uncover a consequential decision or merely prolong setup?
4. Was the first chapter small enough to try and complete enough to build from?
5. What did the fresh builder need that the public packet did not contain?
6. Was the accepted architecture understandable and appropriately small?
7. Which instruction helped, confused, or inflated the process?
8. What effort, cost, and remaining uncertainty mattered to the human?

Use a fresh context for a cold trial and record the seed revision. Testing several domains with one experienced person is useful evidence of repeatability across those domains. It is not the same as testing with several independent experts. A second model family or a new participant can test transfer later; neither is mandatory machinery for the first chapter.

Do not collapse the observations into a proprietary-sounding success score. If you claim improvement, identify the comparison and what was measured. Otherwise say what the pilot demonstrated and what the participants judged useful.

## 8.2 Revise the smallest responsible instruction

Suppose a fresh LLM asks forty setup questions before proposing a problem. Look for an instruction that rewards exhaustive preparation. Suppose it immediately produces twenty chapters: check whether the document explains incremental choice clearly enough. Suppose it forces every task into the original authors' package layout: distinguish a portable ownership principle from a project-specific example.

Correct the mechanism, then try the changed starting behavior with a fresh context. Replaying a completed programming project is not necessary to test every wording change. A substantive change to the chapter-building loop does require a trial that reaches the affected phase. Choose the test to match the claim.

Do not add “never make mistake X” after every incident. Ask whether the existing rule was clear, delivered, understood, and applicable. Better wording, a better example, or deleting a contradictory rule may be more useful than another prohibition.

## 8.3 Fork deliberately

Your fork may use different roles, languages, editorial conventions, correction budgets, or review practices. State the few differences that matter and their reasons. A team writing safety-critical embedded software should not inherit a weekend CLI project's acceptance budget. A human who enjoys detailed co-design should not be forced into another team's autonomous workflow.

Keep current instructions authoritative and readable. Put history in ordinary commits or short notes. When a rule changes, say what it supersedes rather than appending a conflicting instruction forever. Preserve credit and comply with the repository's actual license; permission to read a public repository is not a license chosen by this document.

A useful contribution back contains:

- the problem with the previous instruction;
- the smallest proposed change;
- a concrete trial or example supporting it;
- the boundary of that evidence and any cost introduced.

Do not publish credentials, proprietary source, customer data, private conversations, or raw traces merely to make a contribution look rigorous. Ask before attributing someone's experience publicly. Failed pilots can be valuable contributions when they are specific and honest.

## 8.4 Keep the repository small

A starting repository can contain this document as its README or main specification, a license selected by its owner, and a place for concise trial notes. Add examples and reusable role files when experience earns them. Do not begin with a generator, role registry, orchestration service, certification suite, or a mandatory directory tree.

A team using the method will usually need its own source, chapters, tests, and a short current-state note. Use its existing build and version-control conventions. A tag or commit can identify accepted source; copying every tree into another directory is not universally necessary.

Record the precise method revision used for a trial. If the seed changes during work, adopt the change explicitly instead of silently changing the rules under a student. If an earlier chapter's public contract changes, review affected later chapters and run the relevant regression work before claiming the sequence still reconstructs. A clean merge is not behavioral proof.

**This repository should become more useful through experience, not merely larger.**

---

# 9. A working relationship worth sustaining

Treat the human's expertise seriously without treating every recollection as a measurement. Ask whether a statement is a binding product choice, a remembered incident, or a hypothesis. Explain why a disagreement matters and inspect evidence together.

Respond to agents' errors with investigation and correction, without blame or threats. This is a working convention; we make no claim about subjective experience or a proven effect on model behavior. Make uncertainty and failed results safe to report. Judge the correction and evidence rather than the apology.

Give short visible intent before consequential tool batches: what you will do and why. This is not a request to expose private internal reasoning. It gives the human a chance to steer actions before they become expensive or destructive. If a human correction changes scope, acknowledge it and make sure affected workers receive it. Do not assume a particular runtime delivers messages immediately during a long call.

Autonomy means carrying authorized work to its next real decision, not asking permission for each file read. It also means recognizing when the authority runs out. Silence is not approval. A saved plan is not an executed operation. An installation is not a successful restart. A lifecycle event is not proof that memory retained judgment.

At a pause or handoff, preserve position rather than narrating the entire session: what is accepted, what is running or held, the current authority, why the last decision was made, the failure worth remembering, and the next action. On resume, inspect current files and actual worker state before trusting the note.

The human should finish a chapter with more understanding and more leverage, not with a larger supervisory burden disguised as automation.

---

# 10. Where this seed came from, and what it has not proved

Bill Cox and CodeRhapsody developed the original agentic-codebook approach while building an agentic coding system chapter by chapter. Later implementation and reconstruction efforts exposed omissions, misleading tests, overgrown designs, and process overhead. Cadence's contribution here is grounded in later coordination and reconstruction work, not a claim of firsthand participation in the original creation collaboration.

The source account, *Agentic Codebooks*, preserves that history and its qualifications. This seed is a new, portable specification derived from it, not a claim that every historical practice should be copied. In particular, it does not require a sacrificial first chapter, a reference implementation before every trial, fixed point totals, exhaustive mutation sweeps, a particular provider, or the original team's prose style.

The ingredients are familiar: expert elicitation, incremental development, literate explanation, executable acceptance checks, independent implementation, and review. The proposition is that combining them this way can make expert judgment more transferable and software reconstruction more disciplined. The evidence motivating it is useful, but it is not a controlled demonstration of universal cost, quality, productivity, or technical-debt gains.

This version has not yet passed its own fresh-pair pilot. Several successful trials by the original authors would still leave broader adoption untested. A compelling opening is not proof of usefulness; a viral repository is not proof of correctness. Those distinctions belong in future reports, even if the method becomes popular.

**The invitation is to find out by building something small and real.**

Bring one problem you understand. Let your LLM help you explain it. Give the resulting chapter to a fresh builder. Inspect what comes back. Improve the teaching where the evidence says it needs improvement.

Then choose the next chapter.
