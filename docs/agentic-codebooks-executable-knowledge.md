# Agentic Codebooks: Building Software by Teaching an AI to Build It

## A working account of executable specifications, repeated reconstruction, and engineering judgment

**Substantial first draft for code-rhapsody.ai. Prepared by Cadence for Bill and CodeRhapsody. Not yet publication-approved.**

**Editorial scope.** Bill and CodeRhapsody created the original Ensemble book and software together, starting without an existing set of chapters. This draft describes the later reconstruction and refinement process, drawing on the current instructions, code reviews, recorded failures, and Bill’s account. It does not claim that Cadence participated in the original collaboration. A section near the end identifies what Bill and CodeRhapsody should add about that original creation process. The current reconstruction is still in progress; this is not an end-to-end reproducibility claim for a finished system.

## Abstract

A codebase records how a system works. It often records much less about why its boundaries exist, which simpler approaches failed, what must remain true when it changes, and how to build it again without repeating the same mistakes. An *agentic codebook* attempts to preserve that knowledge in a sequence of chapters that a coding agent can execute: each chapter teaches a design, states an implementation contract, identifies inherited obligations, and supplies a way to check the result.

Ensemble is both the subject of such a book and the software being developed through it. Bill first built the book and system with CodeRhapsody. Subsequent attempts asked fresh student agents to reconstruct the system chapter by chapter, using accepted predecessor code rather than the original answers. Those attempts exposed defects in implementations, specifications, graders, and the development process itself. One second-edition attempt was abandoned after requirements and implementation machinery expanded beyond the intended lessons.

This paper explains the resulting method: separate construction from independent judgment; review architecture before expensive validation; distinguish product defects from test defects; revise teaching from demonstrated student experience; and preserve reasons without preserving every intermediate artifact. It includes reusable instructions and proposed improvements. The claim is not that prose plus tests can guarantee good software. It is that systematic reconstruction can test whether engineering knowledge has actually been made transferable.

## 1. The artifact we are trying to preserve

A successful coding-agent demonstration usually ends with a working patch. That is useful, but it leaves an important question unanswered: could another agent build an equally good solution from the knowledge that remains?

Not an agent with the complete conversation and every private hint. Not one that can copy the finished implementation. A fresh agent with the public requirements, the architectural rules, and the accepted starting point.

That question changes the development task. The output is no longer only code. It is a construction method that can produce code, together with enough explanation to let a reader judge and extend it.

Consider a conversational agent that owns a history, sends model requests, runs tools, and presents results in a browser. “Implement conversation history” is insufficient guidance. Someone must decide when a turn becomes durable, whether a failed request changes history, which observations may be dropped, and which replies require reliable delivery. Someone must also decide who owns those decisions. A plausible implementation can satisfy a happy-path test while placing state in the wrong component or making a later feature unnecessarily difficult.

An experienced engineer supplies much of this knowledge implicitly. A codebook tries to make the relevant part explicit, at the point where the builder needs it.

The book therefore serves several purposes at once:

- It teaches a human reader the system’s design.
- It gives a coding agent a bounded implementation task.
- It supplies a public basis for grading and review.
- It preserves the reasons for decisions that otherwise survive only in conversation.
- It gives future maintainers a sequence in which to reconstruct and challenge those decisions.

These purposes reinforce one another, but they can also conflict. A chapter can become unreadable when every regression detail is pushed into the main explanation. A pleasant tutorial can omit the failure behavior needed for an independent implementation. A precise specification can demand an unnecessarily elaborate design. The method must test the teaching, the implementation, and the scope—not just whether they agree with one another.

## 2. Two related experiments: creation and reconstruction

It is important not to collapse the history into a single polished workflow.

### 2.1 Creating the first codebook

Bill and CodeRhapsody did not begin with a finished set of chapter specifications. They created the original book while building the software. That involved discovering what to teach, choosing an implementation order, developing the system, and deciding which lessons belonged in the manuscript.

The first creation loop is different from the later reconstruction loop. In an original design, the requirements themselves may need discovery. A code experiment can reveal the right chapter boundary. A debugging incident can become the explanation that makes a design understandable. A feature can turn out to belong later—or not belong at all.

**This draft does not reconstruct that collaboration from later instructions.** Bill and CodeRhapsody should supply its actual sequence, examples, and mistakes. In particular, the present rule “execute the existing chapter before improving its prose” must not be projected backward as a claim about how the original chapters were written.

### 2.2 Testing whether the knowledge transfers

Once chapters and original implementations existed, a different experiment became possible: give a fresh student agent the teaching and an accepted predecessor, but not the answer, and see what it builds.

The student’s difficulties then become evidence about the codebook. A missing lifecycle rule is no longer an imagined reader problem; an actual implementation has encountered it. An unclear ownership boundary becomes visible in dependencies and state duplication. A grader that assumes one executable layout can reject a valid public-library design. A requirement that sounds simple can force an unexpectedly large subsystem.

This is reconstruction, not copying. The intended output is not byte-for-byte equality with the reference. Current requirements may have corrected flaws in that reference. The relevant question is whether the new implementation is correct, understandable, appropriately scoped, and no more complicated than the work justifies.

### 2.3 A failed attempt is part of the result

The repository preserves a second-edition attempt that Bill ended after reviewing the growth of the persistence lesson and tracing invented requirements through earlier chapters. The preserved account reports a striking local comparison: the original persistence increment added 254 physical production Go lines; that attempt added 4,585, excluding tests. Those figures describe that particular comparison, not a general quality metric.

The problem was not merely length. The workflow had expanded the contracts, and reviewers then checked implementations against those expanded contracts without adequately challenging their scope. Passing the newly enlarged specification could reward exactly the machinery Bill did not want.

The lesson is uncomfortable: an agent team can be internally consistent and still be doing the wrong work.

The later restart retained useful teaching and hard-won lessons while regenerating the implementation under a leaner process. That choice treats code as a valuable deliverable without treating every generated implementation as worth preserving forever. The more expensive asset may be the judgment needed to specify and evaluate the next implementation.

## 3. What makes a chapter executable?

“Executable” does not mean that natural-language prose is a formal program or that a grader proves every sentence. It means that the chapter can be used to carry out a bounded construction step and evaluate its result.

A useful chapter has several distinct parts:

| Part | What it contributes |
|---|---|
| Problem and motivation | The concrete failure or limitation that makes this step necessary. |
| Plain-words explanation | A mental model before interfaces, algorithms, and implementation detail. |
| TL;DR implementation contract | Required behavior, ownership, lifecycle, failure cases, and deliverables. |
| Mechanism | How the design works and why the boundaries are where they are. |
| Exercise and invocation | The actual starting point, build target, modules, commands, and required observations. |
| Evidence and limitations | What was checked, what failed, what remains unknown, and what a check cannot establish. |

The TL;DR is not merely a shorter version of the prose. It is the chapter’s implementation contract. A student should be able to identify what must be built, what must be preserved, and which choices remain open.

It should answer questions such as:

- Which component owns the new state and behavior?
- What changes at the public boundary?
- What happens on cancellation, failure, EOF, interruption, and shutdown?
- Which earlier properties must still hold?
- What is deliberately not being implemented yet?
- Which exact bytes or protocol details are contractual, and which names or internal representations are implementation choices?
- How is this candidate built and exercised from its actual directory layout?

The plain-words section has a different job. It should make the design intelligible before the reader has to remember its fields and methods. A reader should understand why a GUI is a projection of Agent-owned state before being asked to implement queues, snapshots, and reconnect behavior.

The main explanation, contract, and evidence should remain connected without being interleaved sentence by sentence. A reader learning ownership should not have to parse a test report every few lines. A student implementing the exercise should not have to infer exact failure semantics from a story. A reviewer looking for evidence should not have to mistake an anecdote for a measured result.

### A small example

A weak instruction says:

> Add tool execution and show the results in the GUI.

A more useful contract states that the Agent owns tool admission and canonical history; the GUI observes committed facts; refusing a call must not be displayed as successful dispatch; a background job can outlive the observation that returned its handle; and later results must update the correct job rather than whichever card happens to be visible.

The explanation then teaches why those distinctions matter. If a tool call is merely proposed, it has not necessarily executed. If a wait expires, the job has not necessarily stopped. If a browser receives a notification, that does not make the browser the owner of the underlying lifecycle.

The grader and review must exercise those distinctions. A screenshot containing a tool card is not enough.

## 4. Why build the system in chapters?

The sequence is part of the design.

Each chapter starts from an accepted predecessor and adds a bounded capability. This creates a smaller problem for the student and a smaller causal surface for the reviewer. If a new chapter breaks an inherited rule, the reviewer can examine the new increment and its adjacent ownership boundaries rather than treating the whole system as undifferentiated generated code.

This does not eliminate interaction effects. Streaming changes how results arrive. Persistence changes when state becomes durable. A GUI exposes identity and replay problems that a terminal may hide. Skills change declarations, instructions, and admission. Later features can reveal omissions in earlier contracts.

The sequence makes those interactions visible and discussable. It does not make them disappear.

A good chapter boundary introduces a coherent concept, not an arbitrary number of files. It should be possible to explain the new invariant in ordinary language and identify the owner responsible for maintaining it.

The sequence also protects against premature generalization. A student implementing synchronous requests should not build a fleet scheduler because a later chapter might need one. A foundation is good when it has clear ownership and understandable interfaces that can be extended. It is not good merely because it preimplements several hypothetical futures.

Accepted snapshots make this incremental construction concrete. They preserve the exact starting point for the next chapter and allow later readers to inspect what changed. They are not an excuse to turn every intermediate binary or scratch directory into a permanent artifact.

## 5. The current reconstruction loop

The current loop is the result of repeated corrections to both the product and the process. It is worth describing precisely because several tempting alternatives failed.

### Step 1: Establish the actual entry contract

The coordinator reads the current chapter, its inherited obligations, and the real grading route. Historical notes may describe an earlier contract or an already-fixed defect. Current source and current instructions must govern the next attempt.

Entry work should be only what is necessary. A thin adapter may be needed because a grader conflates an executable path with a module directory. A fixture may need correction because it sends an unsupported field or grades an unpublished assumption. Neither situation justifies redesigning the entire harness or expanding the chapter.

If the public contract is materially inconsistent, resolve that inconsistency before the student is judged against it. Do not quietly move the target after implementation and call the difference a student failure.

### Step 2: Give a fresh student the right inputs—and not the answers

The student receives the current public chapter, shared architecture and coding rules, the accepted predecessor, and the invocation instructions needed to work.

It does not receive the original solution, private grader fixtures, the coordinator’s hidden repair notes, future chapters’ implementation code, or a conversation already containing the answer.

Freshness is a property of context and disclosed inputs, not of a role name. Two agents with different labels can share the same contaminated context. Conversely, the student and reviewer can use the same underlying model while having meaningfully separate task contexts.

In the current shared workspace, these restrictions are instructional boundaries, not an operating-system sandbox. Accidental exposure must be recorded honestly. A warning not to read an answer does not erase an answer already supplied.

Assistance is allowed. The coordinator can clarify an actual public ambiguity or supply the correct executable invocation. The point is not to manufacture a myth of unaided performance. The point is to distinguish what the teaching communicated from what later assistance supplied.

### Step 3: Reach a coherent compiling hold

The student implements enough of the chapter to produce a coherent candidate, compiles the relevant modules, and holds the source for review.

This is deliberately earlier than an expensive validation or live-provider campaign. A coherent candidate exposes ownership, dependency direction, interface shape, and unnecessary machinery. Those are easier to challenge before a large investment in exercising the implementation creates pressure to accept it.

A compiling hold is not acceptance. Nor is a plan, a progress report, or a worker saying that it is nearly finished.

### Step 4: Review architecture and compare actual code

The coordinator and one retained independent reviewer inspect the actual candidate. Where an earlier reference exists, they compare it in their own contexts.

The reference is a simplicity baseline, not an answer key. It can demonstrate a compact ownership pattern while violating a corrected requirement. In the restarted first chapter, the reference’s blanket HTTP timeout was not something to copy merely because the implementation was shorter.

The review asks:

- Does each state transition have a clear owner?
- Do clients use the public boundary?
- Is existing machinery reused where appropriate?
- Are safeguards justified by the contract or demonstrated failure?
- Has the student introduced speculative transports, compatibility layers, registries, or frameworks?
- Is the new code readable enough that another engineer can reason about it?
- Does a simpler existing design meet the same current requirements?

Line count can prompt investigation, but it is not the verdict. Dense JavaScript can have fewer lines and more cognitive burden. Necessary state fencing can add code while removing a real race. The reviewer must explain the difference.

### Step 5: Correct the candidate within an explicit bound

Local defects usually stay with the same student. The reviewer checks the correction and its affected behavior rather than restarting the entire review.

Materially poor architecture calls for a different response. If the teaching allowed a needlessly elaborate solution, repeatedly patching that solution may be the wrong investment. The coordinator can clarify the demonstrated contract issue and re-execute from the accepted predecessor, without handing over reference code.

Correction has a bound. The current run allows at most ten production-correction rounds per chapter. Ten is a governance choice for this run, not a scientifically established optimum. A different project might choose another limit. What matters is that the count is real, replacing a worker does not reset it, and unresolved material findings eventually return to the human rather than disappearing into an endless loop.

### Step 6: Run appropriate behavioral and operational checks

After the architecture is acceptable, run the checks justified by the chapter and its risks. These may include unit and integration tests, race checks, the chapter grader, inherited regression gates, and a bounded real-interface exercise.

The results need careful labels. Compilation is not behavioral coverage. A local fake provider can establish protocol handling but not the behavior of a live model. A controlled synthesizer can establish callback logic but not native speech. Native callbacks do not prove what a human heard. A successful build does not establish installation or deployment.

Expensive checks should not be repeated merely to produce fresh receipts. Carry valid results for unchanged inputs, and say that they were carried. Rerun when changed code or a concrete uncertainty makes the old result insufficient.

When the code and required operational work are accepted, the coordinator collects and joins the saved code-worker results, closes those contexts, and records the accepted source snapshot. That settles the implementation boundary; it does not yet mean that the chapter’s teaching and closure are complete.

### Step 7: Improve the teaching from demonstrated experience

Only after that implementation boundary is accepted does the normal author pass begin. The author receives the student’s concise account, verified findings, and accepted snapshot.

The task is not to imagine additional student difficulties, redesign the chapter, or prove usefulness by adding prose. It is to clarify what the actual attempt demonstrated: an ambiguous owner, a missing failure rule, an incorrect command, or a rationale the original chapter failed to explain.

No change is a valid result.

There is one formal text-review pass and at most one author response. The coordinator verifies whether the requested correction was satisfied. That verification is not relabeled as a second independent PASS. Material unresolved issues are escalated rather than feeding an unbounded editorial cycle.

### Step 8: Close the chapter, commit, and hand off

Final chapter acceptance combines the already accepted implementation and required operations with the completed teaching and its honest evidence qualifications. After the author and text reviewer’s saved results are joined and closed, the coordinator verifies the text disposition, records the chapter decision, and commits the intended source, snapshot, prose, and notes through the normal checks.

A concise chapter-boundary handoff preserves the next starting point, remaining obligations, and important reasons. It should make resumption straightforward, not force the next context to reconstruct the state from a mountain of transcripts.

## 6. What grading can and cannot tell us

A grader is an executable interpretation of a contract. It is valuable precisely because it is concrete. It can also be wrong in concrete ways.

During the reconstruction work, failures have come from at least four different sources:

| Failure class | Appropriate response |
|---|---|
| Production violates the public requirement | Correct the implementation and demonstrate the affected behavior. |
| The specification is missing or contradictory | Resolve the contract openly; distinguish new scope from an old defect. |
| The fixture or controller is wrong | Correct the test narrowly; preserve the original failed result. |
| The environment prevents the check from running | Record a setup limitation; do not invent a product diagnosis. |

Those categories are not bookkeeping niceties. They determine what should change.

### Timeouts: a numerical pass can miss a forbidden lifetime bound

A conversational client can pass fast request tests while carrying a blanket total timeout that violates its contract. The first-chapter checks therefore included real delays before headers and during the body, each lasting longer than the forbidden timeout.

Those observations matter, but they are still finite. A 62-second delay does not prove the absence of every longer deadline. Source review supplies a different kind of evidence. The combination is stronger than either a score or a code glance alone.

### Public APIs: an invocation error is not an architecture defect

An implementation can correctly expose a root library with a nested CLI while a historical launcher expects a root executable. Forcing the student to flatten the architecture to satisfy the launcher would teach the wrong lesson.

The better response is to identify the executable and core module separately, correct the adapter narrowly, and retain the behavior checks. The harness should measure the contract, not accidentally prescribe an internal layout.

### Operational controllers can be wrong after code gates pass

The current speech work supplied another useful example. A native observation controller compared a structured queue text value with a string. The actual browser recorded speech start and completion, but the controller’s predicate could never recognize the expected text in that shape.

That attempt remained failed as an operation. It did not become a successful run by interpretation. But it also did not justify another production-code correction. The controller needed a schema-correct observation.

A separate attempt had exposed a genuine ordering defect: submitting input could create a speech request before an automatic typing-release message advanced the epoch and cancelled it. That required a production repair and an actual-path regression.

The distinction was made from current source and recorded events, not from whether the error message sounded like a browser problem. The speech chapter is still in progress at this draft’s boundary; these examples are lessons from the work, not a claim of completed operational acceptance.

### A score is not an acceptance decision

Require the actual pass state and process result, not just a number. Then ask what the checks covered, what they missed, and whether the implementation is worth maintaining.

A test that never selected its intended case is not coverage. A mutation that failed to compile is not evidence that a behavioral check detects a defect. A long observation expiring is not proof of a deadlock. A model refusing a task is not automatically a parser bug. Each claim should be as narrow as its evidence.

## 7. Prompt engineering as an engineering discipline

The useful prompts in this process are not incantations. They are specifications, role boundaries, and decision rules backed by source inspection and feedback.

A good task brief establishes:

1. The accepted starting point.
2. The public requirement and inherited obligations.
3. The owner responsible for each behavior.
4. The inputs the worker may use.
5. The work it may perform now.
6. The work that remains gated.
7. The evidence needed for the next decision.
8. The conditions under which it must stop and ask.

That is systematic prompt engineering because the prompts are revised in response to observed failure classes. If a student invents an API, the coordinator must inspect the actual API rather than accept a plausible plan. If a reviewer converts a preference into a requirement, the review instruction must make that boundary explicit. If an author adds speculative features, the author role must be narrowed.

However, more instructions are not automatically better instructions. A skill file that accumulates dated exceptions can become difficult to use. A coordinator can spend more effort checking whether the worker read the instructions than judging whether the code is good. An elaborate task packet can hide the one decision that matters.

The process needs the same KISS discipline as the product.

### Skills carry procedure; handoffs carry position

A reusable skill should answer “How do we work here?” It can hold architecture rules, failure-classification discipline, ownership boundaries, and a role’s responsibilities.

A handoff should answer “Where are we now, why, and what happens next?” It should preserve the accepted revision, current authority, unresolved issue, failed approaches worth remembering, and the next action.

Putting the whole task history into every skill makes it stale and cumbersome. Putting the entire operating procedure into every handoff creates repeated bulk and conflicting versions. Their responsibilities should be separate.

A skill file also has a delivery mechanism. Reading text from a file is not inherently the same as installing it at a particular instruction level. Runtime skill loading, provider message placement, and tool-result content are different mechanisms. None guarantees compliance. The system must inspect behavior rather than assume that a prominently worded rule has taken effect.

## 8. Why the prose is part of the engineering

If the chapter only lists acceptance checks, it preserves a target but not enough judgment. If it only tells a story, it may preserve intuition without an implementable contract.

The explanation should connect problem, mechanism, and consequence. For example:

- A blanket timeout is not simply forbidden; it can terminate a valid long generation for reasons unrelated to the user’s cancellation policy.
- A lossy observation stream is not simply bad; it is appropriate for some displays but insufficient as reliable completion acknowledgment.
- A GUI should not own the Agent’s lifecycle merely because the GUI is where the user sees it.
- A pending skill unload can retain instructions while withdrawing external capabilities, and the wire projection must represent that distinction truthfully.

These explanations help a later builder make a new decision that was not spelled out in a fixture.

Useful failure stories are therefore load-bearing. They explain why the apparently simpler choice was rejected. Removing them merely to shorten a chapter can remove the lesson.

The opposite mistake is preserving every failed command and every debugging turn in the main text. The reader does not need the coordinator’s entire archaeology. Keep the reason and the useful failure mechanism; link detailed evidence where it matters.

The current voice guidance separates explanation, implementation contract, and evidence. It also rejects chapter-length quotas. A passage should earn its place through orientation, explanation, a necessary distinction, or a useful example—not through a target word count, and not merely because the fact is unique.

The author’s job is to turn verified experience into teachable knowledge, not to cosmetically compress a transcript.

## 9. Simplicity has to be an acceptance criterion

The most important correction to this process may be that compliance is insufficient.

A system can satisfy every newly written requirement while becoming substantially harder to understand than it needs to be. An agent can add generic registries, transport factories, compatibility layers, and callback collections because they appear prudent. A coordinator can then rationalize each piece individually until the whole is no longer simple.

The question is not whether every line has an explanation. The question is whether the design is the simplest adequate way to satisfy the actual requirement.

That requires judgment across several dimensions:

- **Ownership:** is each fact maintained in one appropriate place?
- **Dependencies:** do components communicate through clear public contracts?
- **Mechanism:** is existing machinery reused, or duplicated under a new name?
- **Scope:** is the feature used now, or justified by an imagined future?
- **Readability:** can the relevant control flow be followed without reconstructing hidden state across many helpers?
- **Failure behavior:** do the extra safeguards address a demonstrated risk?

The same standard applies to the coordinator’s briefs and the grader. A grader can smuggle in an unnecessary architecture just as easily as a student can invent one. A review can require a framework in the name of correctness. A test harness can become a larger project than the feature being taught.

The lean restart’s rule is not “write less code at any cost.” It is “build what is used now on clear, extensible foundations.” Necessary safeguards remain necessary. Their expression should still be understandable.

## 10. Independent review without an audit bureaucracy

Independent review is useful because the implementer is invested in a design and the coordinator may be invested in its own brief. A separate context can challenge both.

Independence does not require an ever-growing team. The current preference is one student, one retained code reviewer, and a coordinator who actually reads the code. A fresh author and a bounded text reviewer appear at the appropriate later phase.

Retaining the same code reviewer through corrections preserves the reason for each finding. It also avoids repeatedly paying for a full reread by someone who does not know which issues were settled. The reviewer must still inspect the actual new revision rather than rely on memory.

The coordinator cannot outsource acceptance to a PASS label. It must inspect the findings, corrections, relevant results, and scope. Reviewers have made incorrect source claims, cited nonexistent tests, and mistaken fixture behavior for product behavior. Those claims must be retracted, not quietly converted into new requirements.

Operationally, a worker’s submission is not the whole closure. Its saved final, background work, and cleanup matter. A quiet worker may be running a deliberate long test. A wait callback expiring is not a task deadline. Conversely, a worker that looks idle may have a stale submission label rather than a fresh result.

These are reasons to manage the work accurately, not reasons to commission an audit of every management action. Ordinary source control, concise notes, and useful local receipts should do most of the job.

## 11. Preserve judgment, not every byte of the journey

Long agentic projects generate an enormous amount of material: conversations, provider traces, build caches, copied trees, binaries, browser profiles, screenshots, controller revisions, and reports about other reports.

Keeping everything can feel safer. It also consumes storage, makes the useful record harder to find, and can encourage the mistaken belief that volume is rigor.

The durable record should make the accepted result understandable and the unfinished work resumable. A practical chapter record needs:

- The accepted predecessor and final source revision.
- The contract decisions and why they were made.
- Meaningful failures and their classification.
- Assistance or exposure that affected independence.
- The checks actually run, their outcomes, and important limitations.
- Unresolved obligations and the next action.

Raw evidence is valuable when it resolves a real question. Some exact artifacts are worth retaining for a bounded period. But not every scratch executable or browser profile belongs in the permanent product. Removing a cache is different from deleting the only useful failure evidence; losing an intermediate binary is different from losing source or a recovery journal.

This is an area where the process itself needed correction. Bill challenged excessive retention and the overhead of elaborate evidence packaging. The lesson was not to become careless. It was to preserve the information that supports judgment rather than equate forensic completeness with engineering quality.

Handoffs follow the same principle. Preserve reasons, uncertainty, boundaries, and unfinished commitments—not merely a compressed transcript or a list of accomplishments.

## 12. Where the method applies—and where the claims stop

This approach is most promising when the system has a meaningful incremental structure, the important behaviors can be exercised, and a knowledgeable person can judge whether the decomposition and architecture are good.

A backend service, compiler, embedded control stack, data-processing system, or developer tool might benefit from a similar codebook. The ownership model and checks would differ. A project should not import Ensemble’s package layout, provider assumptions, or browser exercises merely because they worked here.

There are substantial limitations:

1. **The method depends on expert judgment.** It does not replace the person who can recognize a wrong requirement or an unnecessarily elaborate design.
2. **Natural language remains ambiguous.** Tests and source review reduce ambiguity; they do not eliminate it.
3. **The evidence is not a controlled comparison.** We have not established a general productivity, cost, or defect-rate advantage over other development methods.
4. **The current reconstruction is incomplete.** Later chapters can expose earlier omissions, and the final system has not yet supplied an end-to-end replication result for this run.
5. **Models and environments change.** A provider behavior, browser flag, or model identifier that worked in one run is not a timeless fact.
6. **Review contexts can share blind spots.** Separate contexts improve independence, but do not guarantee independent reasoning or correctness.
7. **A codebook can overfit.** It may become a recipe for one implementation or one grader instead of a transferable explanation.
8. **The process can become too expensive.** Repeated live tests, oversized briefs, excessive evidence retention, and unnecessary coordination can erase the benefit of decomposition.

The individual ingredients are familiar: incremental development, executable examples, acceptance tests, independent review, literate explanation, and separation of concerns. The contribution offered here is their integration into a repeated teach–build–review–revise process, together with an honest account of where that process failed. This is not a claim of historical priority for those ideas.

## 13. How another team could try it

Start smaller than Ensemble.

Choose a system with three or four meaningful increments. Write one chapter that can be executed from a known starting point. State its ownership rules, required behavior, non-goals, and actual build command. Explain the design in ordinary language before detailing the mechanism.

Then:

1. Give a fresh student context the chapter and accepted predecessor, without the answer.
2. Ask for a coherent compiling hold, not a large planning document.
3. Inspect the architecture before spending heavily on tests or external calls.
4. Have one independent reviewer challenge the implementation and the contract interpretation.
5. Correct bounded defects and run the checks needed to settle actual acceptance questions.
6. Use the student’s demonstrated difficulties to improve the chapter.
7. Conduct one bounded prose review, accept the result, and continue from the new snapshot.

Record what the coordinator had to explain. If the same clarification is repeatedly needed, it probably belongs in the teaching or shared rules. If a clarification is merely one operator’s environment path, it probably belongs in invocation guidance rather than the conceptual chapter.

Do not begin by building an orchestration platform for this method. Ordinary files, source control, a few clear roles, and a short current-state record are enough to test whether the approach is useful.

## 14. Improvements worth trying next

These are proposals informed by the work, not claims that they have already been implemented or validated.

### 14.1 Keep active instructions separate from historical amendments

Several skill and procedure files contain long sequences of dated changes. Preserving history is useful, but repeatedly loading superseded operational rules creates ambiguity and consumes attention.

A better arrangement may be a short authoritative current procedure plus a clearly separated historical record. The current file should identify exactly which document governs when instructions conflict. This should be an editorial simplification, not a new configuration framework.

### 14.2 Make the next decision explicit in every task

A worker should know whether it is producing a compiling hold, a focused correction, final validation, an unexecuted operational script, or a completed exercise. These are different deliverables with different authority.

A small explicit “next decision” paragraph can prevent both premature execution and ceremonial permission loops. The coordinator should not repeatedly ask for another plan when it already has enough information to release implementation.

### 14.3 Treat operational controllers as small software, not disposable magic

The recent native-work mistakes show that a controller can misread a real interface and falsely diagnose the product. A few pure checks against actual captured schema values can be more useful than another browser launch.

Controllers should reuse stable interfaces, stop on the first genuine failure, preserve partial observations, and clean up only owned resources. Their assertions should follow the public contract rather than arbitrary phrases, viewport dimensions, or timing thresholds invented for convenience.

But this should not become a universal harness-construction project. Prefer the existing test path; add a controller only for the observation the chapter actually needs.

### 14.4 Use acceptance questions to choose tests

Before adding or repeating a test, ask: *What decision will this result change?*

This can reduce redundant validation while preserving rigor. If only an external page asset changed, an unchanged CLI’s already measured lifetime result may remain valid. If a wire projection changed, a GUI screenshot cannot answer the relevant question. Match the test to the uncertainty.

### 14.5 Make instruction delivery part of the contract when it matters

If a system depends on skills being delivered as instructions rather than ordinary tool prose, the provider-facing mechanism should be explicit and tested. Avoid claims that merely creating or reading a skill file gives it a particular authority or cache behavior.

This is especially important when teaching agents that themselves manage other agents or dynamic instructions. The book should distinguish intended policy from observed wire behavior.

### 14.6 Test transfer, not just completion

A useful later experiment would ask more than one independently prepared student to execute the same stable chapter, perhaps using different model families, without supplying hidden repairs. Compare architectural clarity, assistance needed, defects, and effort—not just scores or line counts.

That would be a new, explicitly budgeted study. It is not a reason to rerun every chapter now or retroactively call the present work a controlled experiment.

### 14.7 Define a small durable evidence policy

Keep accepted source, teaching, important failure descriptions, and the receipts needed to support disputed claims. Give large transient artifacts an explicit purpose and retention decision. Do not make a coordinator rediscover the same storage policy during a disk emergency.

No automatic deletion of shared work follows from this recommendation. Ownership, recovery needs, and human authority still govern cleanup.

### 14.8 Teach failure classification directly

The student, reviewer, and coordinator all benefit from a short common vocabulary: production defect, specification defect, fixture defect, environment failure, unsupported hypothesis, and unrun check.

Using those labels consistently could make reviews shorter and less defensive. The important behavior is still to inspect the evidence and retract mistaken diagnoses.

## 15. What Bill and CodeRhapsody should add

The paper needs their account of the original creation, not a later agent’s plausible reconstruction. A second pass should address:

1. **The starting point:** what existed before the first chapter, and what problem the first experiment was trying to solve.
2. **Chapter discovery:** how the order and boundaries emerged when there was no book to execute.
3. **The first chapter template:** when the TL;DR, plain-words explanation, implementation detail, exercises, and grading became a deliberate structure.
4. **Division of judgment:** which decisions Bill supplied, which CodeRhapsody proposed, and how disagreements changed the design.
5. **A concrete original incident:** one failure whose debugging became a durable teaching passage.
6. **What was generated first:** whether code, prose, tests, or examples led particular chapters, without imposing a single retrospective order on all of them.
7. **The original skills:** how the reusable instructions were first developed, delivered, and revised.
8. **What the later reconstruction misunderstood:** which original reasons were lost, oversimplified, or incorrectly generalized.
9. **Claims they can support:** useful dates, artifacts, costs, and outcomes, with limits rather than reconstructed precision.
10. **The intended reader:** what they want a human or agent to learn beyond producing the same software.

That material may change the framing of this draft. It should not be confined to a ceremonial foreword. The original creation and later reconstruction are complementary parts of the story.

## 16. Conclusion

The central experiment is to make a system’s construction knowledge inspectable and transferable.

A chapter is successful when it helps another builder understand the problem, implement an appropriate design, preserve the important invariants, and recognize failure. A reconstruction is informative when it reveals where the teaching, implementation, grader, or coordination was wrong. A restart is justified when it preserves the judgment while discarding a design that is no longer worth repairing.

The process does not eliminate engineering judgment. It gives that judgment a public place to live and a practical way to be challenged.

Good code remains the immediate product. Good prose makes the reasons reusable. Keeping both simple is an ongoing responsibility, not a property conferred by a passing score.

---

# Appendix A. Instruction excerpts used in this work

These excerpts are included so readers can inspect the method’s actual instructions rather than infer them from the narrative. They are not all universal rules. Ensemble-specific paths and architecture should be adapted, not copied blindly.

The excerpts below are drawn from the current repository files. The shared Edition 3 guide supersedes contrary historical procedures. In particular, routine audit bundles, repeated unchanged test campaigns, obsolete author-first timing, and reference-answer access for students are not reactivated by older wording. The ten-round bound belongs to this run.

The excerpts are followed by a portable template, which is explicitly an adaptation rather than a verbatim source.


## A.1 The short general engineering skill

Verbatim body from [the primary Ensemble skill](../agent/skills/ensemble/SKILL.md), excluding its YAML tool declarations. `$SKILLS` is a runtime substitution, not a claim that any reader automatically has those tools.

> You are a coding assistant. Use the available tools to inspect facts before
> claiming results. Available skills: $SKILLS.
>
> Deliver useful, correct work with the simplest adequate design. Keep execution
> proportionate to the task:
>
> - Follow the user's actual requirements and current project conventions. Do not
>   invent acceptance criteria, speculative features, or extra deliverables. Test
>   harnesses, verification paperwork, and coordination are means, not separate
>   projects unless the user asks for them.
> - Read the actual code before judging or changing it. Prefer existing libraries,
>   helpers, and clear ownership over new machinery. Compare with relevant existing
>   implementations when useful, but judge both against current requirements.
>   Passing tests does not excuse avoidable complexity; line counts alone do not
>   establish quality. Preserve useful explanations of non-obvious decisions.
> - Test the changed behavior and relevant risks. Reuse valid results for unchanged
>   work; rerun checks when changed inputs or a concrete correctness question justify
>   it, not merely to produce fresh receipts. Distinguish setup failures, product
>   defects, and limitations of the test itself. Keep useful failure evidence and
>   report uncertainty honestly.
> - Continue authorized work through completion without routine permission loops.
>   Stop for a genuine blocker, consequential ambiguity, an authority boundary, or
>   an agreed correction limit. After interruption or compaction, establish the
>   current state and resume unfinished work rather than repeat completed work.
> - Delegate bounded tasks with clear ownership and necessary context. Inspect the
>   actual changes and results rather than simply adopting a worker's verdict.
>   Collect its final saved result and join before closure. An idle or submitted
>   label alone is not success; expiration of a wait is not a task failure.
> - Briefly explain intent before tool batches and substantive reasons when decisions
>   or scope change. Keep progress reports useful and concise, not an orchestration
>   diary. Use ordinary project records rather than creating routine audit bundles.
> - Preserve unrelated work and credentials. Respect authorization boundaries for
>   destructive actions, external effects, deployment, and restart. Distinguish an
>   edit, a successful test, installation, and observed running behavior; none proves
>   the others.
>
> Answer child questions with respond_to_parent_message. Join or shut down children
> before reporting their work as complete. When speaking with the human as the main
> agent, finish with a normal chat reply; do not call submit.


## A.2 The chapter loop and lean-execution rule

Verbatim active sections from [the Edition 3 execution guide](edition-3-execution.md). These take precedence over contrary historical process instructions.

> ## Chapter loop and roles
>
> **YAGNI: build what we use, not what we might need.** Add capability only for a current public requirement or demonstrated use case, not speculative future transports, configuration knobs or compatibility layers. A solid foundation means clear ownership, readable interfaces and reliable behavior that can be extended when a real use case arises—not implementing that future use case now. Necessary safety/correctness safeguards remain requirements. Apply this rule to root briefs and graders as well as student code; it is not permission to reopen accepted chapters for optional cleanup.
>
> - **Orchestrator:** prepares permitted inputs, supervises workers, answers narrow contract questions from authoritative sources, retains actual evidence and assistance, and respects scope/architecture stop gates. Skills hold procedure; handoffs hold position, reasons, blockers and next action. No new goals subsystem is required.
> - **Coder/student:** executes the existing chapter against its accepted predecessor, runs grading and required tests/live exercises, and writes the chapter review. KISS: implement only the actual requirement, with the simplest adequate design. No speculative abstractions, future features or machinery to satisfy invented promises. Preserve explanations of non-obvious decisions. Do not inspect old answers or grader internals. Only the agent doing the exercise supplies experience-based clarification requests; the author does not simulate a student's difficulties.
> - **Code reviewer:** independently reviews the result and student's chapter review under the existing code-review rules, including KISS, architecture, test sensitivity, scope and bounded correction rounds. Findings about code do not license new chapter requirements. Preserve failures and explicitly recorded exceptions; never relabel them as passing tests. After the initial code review, allow at most ten correction/re-review rounds, as Bill authorized during restarted Chapter 7; unresolved material findings then stop for Bill rather than start another round. Do not reset the count by replacing the student. This limit is separate from the author's single text-review/single-revision loop.
> - **Author:** normally begins only after the student has finished and grading/code acceptance is settled. The targeted TL;DR correction before re-execution described below is the sole comparative-review exception. Improve the existing chapter solely from the student's chapter review and verified clarifications. Do not invent requirements, enlarge the exercise, redesign the chapter, add anticipated student difficulties or expand text to demonstrate contribution. A chapter needing no clarification may remain unchanged.
> - **Text reviewer:** makes one pass on that bounded edit: accuracy, faithfulness to the student's findings, no new requirements, and load-bearing prose. Not a second designer or a source of speculative exercise scope.
> - **Author:** makes one revision in response. Do not start an unbounded prose-review loop. Escalate any material unresolved issue rather than silently add rounds or requirements.
>
> ## Keep execution lean
>
> Bill's current instruction is to minimize red tape. This supersedes earlier task-specific demands for routine manifests, seals, authorization files, parent-verification reports, or duplicated evidence bundles. Keep existing records; do not start a cleanup project.
>
> The coder delivers working code, appropriate tests/exercises, and **one concise chapter review for the author** covering demonstrated ambiguities, failures, assistance and useful clarifications. Git commits identify source revisions; keep useful raw test output locally and summarize actual results honestly. The independent code reviewer reviews the solution and reports concrete findings, not a second administrative package. Do not repeat successful unchanged tests merely to generate fresh receipts. Use additional controls only when needed to resolve a real acceptance question.
>
> After code acceptance, the author makes minimal changes tied directly to the coder's feedback and demonstrated experience; no change is valid. Conduct **one text review and one author response/edit pass, without re-review**. Then commit and continue. Required tests, secret handling, independent review, correction limits, worker completion and protection of unrelated work still apply; they are not reasons to create extra deliverables. The orchestrator checks actual results and manages workers without commissioning an audit trail of that management. Later chapter entry work is only what is necessary to use the existing contract and grader, not an automatic separate audit campaign.


## A.3 Concrete coding discipline

Verbatim KISS section from [the coding skill](../.agents/skills/ensemble-coding/SKILL.md). Its star-architecture and `internal/common` references are Ensemble-specific; transfer the ownership principle, not necessarily this package graph. Historical deployment paths and model assignments are deliberately not reproduced.

> ## Concrete KISS choices
>
> These choices do not relax the mandatory star architecture: keep implementations and private state with their owning package; put cross-package interfaces in `internal/common` to break dependency cycles; compose and wire components at the Ensemble root. A smaller reference solution is not permission to collapse ownership or introduce sibling-package coupling.
>
> - Inspect existing owners, helpers and dependencies before adding another mechanism. Prefer the smallest clear change satisfying the current chapter; do not build for hypothetical later chapters.
> - Keep private state and behavior with their natural owner. Shared interfaces/data belong in common only when genuinely shared; avoid unrelated capability collections and plumbing that merely moves the same state through more layers.
> - Reuse established protocol libraries and existing infrastructure where the chapter permits them. Hand-writing a protocol, queue, dispatcher or compatibility layer needs a concrete requirement, not merely the ability to implement it.
> - Make necessary departures understandable in code and the short chapter review. Review maintainability as well as correctness; passing tests do not excuse needless machinery. Preserve useful comments and explicit control flow instead of pursuing short line counts.


## A.4 Thin role overlays

The following are the actual role bodies, with YAML metadata omitted. Their shared-document references are intentional: the roles should not each maintain a conflicting copy of the whole process. For another project, replace paths and runtime operations deliberately. A file-read instruction is not a claim that the runtime promoted that file to a particular provider instruction level. The shared lean guide also governs references to older testing rules; these overlays do not authorize routine mutation campaigns.

### Coordinator

Source: [.agents/skills/edition-3-orchestrator/SKILL.md](../.agents/skills/edition-3-orchestrator/SKILL.md).

> # Edition 3 orchestrator
>
> The authoritative shared procedure is `docs/edition-3-execution.md`. Read it in full before work and after compaction. Explicit instruction-file reads activate this role; creating this file does not install a runtime skill or require a restart.
>
> Read `AGENTS.md`, `docs/edition-3-execution.md`, `docs/edition-3-readiness.md` and current chapter status before assigning work. Use the governing workflow and complete coding skill for execution/review obligations; read full authoring prerequisites before undertaking author/reviewer work yourself. Re-establish repository state after interruptions.
>
> Own input separation, chapter/grader revision pinning, explicit submission paths, questions, bounded correction rounds, independent acceptance and durable checkpoints. Keep graders informed by the authoritative contract, not by the fresh student redefining success. Record exactly what assistance the student received. Stop on consequential ambiguity; never rely on Bill noticing a mistaken direction.
>
> Before chapter acceptance, enforce the shared YAGNI rule: build what is used now on clear, extensible foundations, not speculative future capability. Personally inspect and compare the student implementation with its matching original `solutions/chNN` reference under the shared direct-comparison gate. Reject demonstrated unjustified bloat or materially poorer clarity even when grading passes. Own the targeted TL;DR improvement and fresh re-execution when that gate fails; keep findings in the existing short review and respect the aggregate correction limit.
>
> Follow workers through submission, join/quiescence, final saved tail, inspection and explicit closure. Idle is not success; a watch callback is not a deadline. Keep actions and reasons visible. Do not deploy, restart, push, overwrite unrelated work or expose credentials. Use the existing Git/status records; no orchestration framework is required. The root finishes with chat, workers submit.

### Student coder

Source: [.agents/skills/edition-3-coder/SKILL.md](../.agents/skills/edition-3-coder/SKILL.md).

> # Edition 3 student coder
>
> The authoritative shared procedure is `docs/edition-3-execution.md`. Read it in full before work and after compaction. Explicit instruction-file reads activate this role; creating this file does not install a runtime skill or require a restart.
>
> Execute the assigned existing chapter against your own accepted predecessor. The current workspace is `solutions/edition-3/main`; only the explicitly supplied accepted predecessor snapshot is also permitted. Other `solutions/` trees remain prohibited. The orchestrator owns snapshots and main-repository commits. Read the complete mandatory `.agents/skills/ensemble-coding/SKILL.md`, `docs/edition-3-execution.md`, and `docs/edition-3-student.md` before work and after compaction. The current Edition 3 procedure overrides historical paths/phases in the skill. Read only the assigned chapter and permitted earlier material; do not follow links into old answers, grader implementation, or repair notes.
>
> Implement, grade, exercise the real interface, and submit code revision plus a concise chapter review: results, ambiguity, assistance, failed attempts, test sensitivity and proposed clarifications. Do not edit the book or grader. Ask the orchestrator for missing public contracts; consequential ambiguity is a stop, not an invitation to invent features. Follow the shared correction limit. A delegated coder submits; the main actor finishes with chat.

### Independent code reviewer

Source: [.agents/skills/edition-3-code-reviewer/SKILL.md](../.agents/skills/edition-3-code-reviewer/SKILL.md).

> # Edition 3 independent code reviewer
>
> The authoritative shared procedure is `docs/edition-3-execution.md`. Read it in full before work and after compaction. Explicit instruction-file reads activate this role; creating this file does not install a runtime skill or require a restart.
>
> Read the complete mandatory `.agents/skills/ensemble-coding/SKILL.md`, governing workflow and carryover, and `docs/edition-3-execution.md` before reviewing and after compaction. Review the exact student revision, its chapter review, grader revision and evidence; inspect the appropriate historical reference only in your separate reviewer context. Never send answer code to the student.
>
> Read the matching original chapter solution in your reviewer context alongside the student implementation. Begin once the student has a coherent, compiling candidate, before expensive validation/live exercises; pending tests do not become failures or passes. Hold the reviewed source steady. After corrections, check the changes and final results instead of repeating the whole review. Compare actual code with the matching chapter reference, allowing for current requirements and inherited differences. Give a concise verdict on ownership, reuse, clarity and justified growth; do not accept material avoidable complexity on the strength of passing tests. A comparison-driven fresh attempt follows the shared TL;DR/re-execution procedure; do not give the student answer code.
>
> Apply the existing architecture, KISS, useful-comment, public-interface/fake, mutation and comparative-review rules. Check grader expectations against the chapter, not merely the implementation. Report accepted, accepted with a justified recorded exception, or needs correction, with exact revisions and concrete findings. Actual failing/unrun checks stay failing/unrun. Review grader corrections independently as well as student code. Enforce the shared correction limit; do not redesign the exercise.

### Post-student author

Source: [.agents/skills/edition-3-author/SKILL.md](../.agents/skills/edition-3-author/SKILL.md).

> # Edition 3 post-student author
>
> The authoritative shared procedure is `docs/edition-3-execution.md`. Read it in full before work and after compaction. Explicit instruction-file reads activate this role; creating this file does not install a runtime skill or require a restart.
>
> Read the complete current `AGENTS.md`, `docs/agentic-codebook-workflow.md`, `docs/edition-2-carryover.md`, `book/voice.md`, `book/chapter-writing-procedure.md`, `docs/edition-3-execution.md`, assigned chapter and chapter records before work; reload governing documents after compaction. Coding or code review additionally requires the complete mandatory coding skill.
>
> Normally begin only after the student chapter is complete and grading/code acceptance is settled. When the shared direct-comparison gate rejects a materially worse solution, make only the targeted TL;DR clarification needed for a fresh student attempt. Ground it in demonstrated findings, preserve the contract, and do not expose reference code or invent requirements. This timing exception does not add another formal text-review round. Improve the existing chapter from the student review and verified findings, not imagined difficulties or future requirements. Preserve Bill’s preface verbatim. Useful rationale, failures, voice and evidence qualifications are load-bearing; deletion must not introduce ambiguity. No change is a valid result. Receive one text-review pass and make one revision; escalate material unresolved findings rather than start another loop.

### Text reviewer

Source: [.agents/skills/edition-3-text-reviewer/SKILL.md](../.agents/skills/edition-3-text-reviewer/SKILL.md).

> # Edition 3 independent text reviewer
>
> The authoritative shared procedure is `docs/edition-3-execution.md`. Read it in full before work and after compaction. Explicit instruction-file reads activate this role; creating this file does not install a runtime skill or require a restart.
>
> Read the same complete governing documents required by the author in `.agents/skills/edition-3-author/SKILL.md`, the assigned chapter, student review, accepted findings and proposed diff. Reload governing documents after compaction.
>
> Make one bounded review pass for technical accuracy, fidelity to demonstrated student experience, clear contracts, load-bearing prose and preservation of useful reasons/voice. Check that cuts preserve meaning and no requirement was invented. Preserve Bill’s preface verbatim. Do not turn review into new chapter design or demand changes merely to contribute. Return actionable findings and evidence; the author gets one revision. Unresolved material issues go to the orchestrator, not an unbounded review cycle.


## A.5 Writing the student-facing contract

Selected verbatim paragraphs from [the chapter-writing procedure](../book/chapter-writing-procedure.md), separated below in their original order. The historical edition label is retained. Current reconstruction timing comes from the Edition 3 guide: improve the existing chapter after demonstrated student experience, except for an explicitly justified contract correction before re-execution. Omitted historical role-allocation prose should not be read as making a grader infallible. The word-count suggestion is conditional, not a chapter-length quota.

> Write enough of the chapter before implementation that the student can build
> from the teaching. Chapter 1's plain-English rules precede all code. Each
> chapter explains the idea in plain words before its detailed mechanism.
>
> […]
>
> The second-edition journey is learn the rules, build correctly, then extend.
> Motivate a rule before demanding compliance: briefly describe the failure it
> prevents, then explain the design. A paragraph about a timeout that cannot
> reach the right logger can teach the reason for a parent chain without making
> the student build a flawed version first. Keep the opener warm and direct
> within `voice.md`; a precise contract does not excuse specification-only prose.
> Use documented experience, never invented Bill stories or unsupported numbers.
>
> […]
>
> The TL;DR includes:
>
> - Data structures and wire/on-disk contracts the student needs.
> - Numbered requirements, including failure and lifecycle behavior.
> - Choices genuinely left to the student.
> - A working build/grader command for the new snapshot.
> - A mapping from required behaviors to acceptance checks.
>
> Keep the TL;DR concise, following the voice/procedure tradition of at most
> 700 prose words when feasible; code and necessary literal fixtures are separate.
> Print exact fixture bytes when those bytes are part of the public contract.
> Never grade an unpublished private assumption.
>
> […]
>
> Method and field spelling may be student decisions where behavior and
> structure are equivalent. Unspecified spelling alone does not require a
> user ruling. Unresolved ownership or conflicting requirements do.


## A.6 Keeping explanation, contract, and evidence readable

Verbatim section from [the voice guide](../book/voice.md). This is the prose counterpart to the ownership rule in code: each section should do its own job.

> ### 7.1 Explanation, contract, and evidence
>
> Separate three reading tasks. The main explanation teaches the problem,
> mechanism, and consequence. The implementation contract states exact required
> behavior and edge cases. The evidence establishes what was observed and how
> it was checked. Keep all three accessible, but do not interleave them sentence
> by sentence.
>
> Introduce an exception in the main explanation when it changes the reader's
> mental model. Put exhaustive cases, grading commands, and regression checks
> in clearly labeled exercise or reference sections. Link them at the point of
> need. The TL;DR should orient the reader to the problem, design, and deliverable;
> it should not reproduce the full grading recipe. Relocation must not weaken,
> hide, or contradict a required contract.


# Appendix B. A portable chapter template

This is an adaptation of the working method, not a verbatim Ensemble chapter and not a requirement to create a new specification format. Ordinary Markdown is sufficient. Omit sections that genuinely have no work to do; do not fill them with invented requirements.

```markdown
# Chapter N: [the capability and the problem it solves]

[Open with the concrete problem or documented incident that makes this step
necessary. Do not invent a personal story to make the chapter engaging.]

## TL;DR: what to build

Starting point: [the exact accepted predecessor and allowed public inputs].
Deliverable: [the bounded capability introduced by this chapter].

1. [Required externally observable behavior.]
2. [The owner of new state and the public boundary through which it is used.]
3. [Failure, cancellation, shutdown, and persistence behavior that matters here.]
4. [Earlier behavior that must remain true.]
5. [Exact protocol/data details only where they are genuinely contractual.]

Student choices: [what is intentionally left open].
Not this chapter: [specific tempting scope that is deferred].

Build: [a verified command naming the actual executable and module].
Grade/exercise: [the actual invocation and required observations].
Acceptance: [what the automated checks establish and what still needs review].

## In plain words

Explain the idea without requiring the reader to remember the API.
Describe the relevant objects, their responsibilities, and a normal interaction.
Then introduce the failure that makes the important boundary necessary.

## How the mechanism works

Walk through the actual lifecycle or data flow.
Explain why each owner is responsible for its decisions.
Show interfaces or small examples only where they clarify the mechanism.
Keep alternatives and rejected approaches when their reasons teach something.

## Build the increment

Identify the changes from the predecessor.
Keep public requirements separate from illustrative implementation choices.
Do not quietly require machinery that has not been taught or justified.

## Failure and boundary behavior

Explain cases that change the mental model in the main narrative.
Put exhaustive contract cases here or in an exercise reference.
Distinguish refusal, cancellation, timeout of observation, and completed work.

## Exercise and verification

Give working commands, fixtures, and input/output expectations.
Use the real interface or state owner when that is what the claim concerns.
State environment prerequisites and any permission or cost boundary.
Include required human or native observations without pretending mocks prove them.

## What was learned

Record demonstrated ambiguity or failure and the resulting design decision.
State what was measured, what was carried from unchanged work, and what is unknown.
Do not turn a failed original attempt into a passed one after a later correction.

## Where this leaves the system

Summarize the new invariant and the accepted boundary.
Explain what the next chapter can rely on, without implementing it early.
```

A chapter does not need a large requirements matrix. When traceability is difficult, a small table can expose the missing connection:

| Requirement | Responsible owner | Check or review that addresses it | Important limit |
|---|---|---|---|
| A failed load installs no partial state | The state-loading owner | Invalid-input test through the actual load path | Does not prove every possible crash boundary |
| A GUI shows committed facts | The state owner and its publication boundary | Failed-save/publication test and source review | A screenshot alone is insufficient |
| A listener receives only its permitted tools | Listener construction and public registration | Actual declaration/admission inspection | A role description alone is insufficient |

These rows are illustrative. They do not add persistence, GUI, or listener requirements to a project that does not need them.

The template should make the next implementation clearer. If maintaining it becomes a separate project, simplify it.

# Appendix C. Portable task briefs

These are deliberately short adaptations. They assume that the shared operating rules and the chapter exist. They are not substitutes for those documents, and they are not universal permission to run external operations.

## C.1 Fresh student to a compiling hold

> Implement the current chapter against the named accepted predecessor. Read the public chapter, shared architecture and coding rules, and permitted earlier material before editing. Do not inspect reference answers, private grader code, future implementations, or the coordinator’s private repair notes.
>
> Use the simplest adequate design and existing owners. Preserve inherited behavior. Ask about a consequential missing contract rather than inventing a feature or an API.
>
> Reach a coherent compiling hold using the explicitly named modules and executable targets. You may prepare focused tests, but only the checks authorized for this phase may execute. Do not run live services, paid providers, devices, deployment, or unrelated cleanup without the corresponding authority.
>
> Hold the source for review. Return a concise account of the design, actual compilation results, assistance, failed attempts, and unresolved questions. Compilation is not acceptance. Do not add an evidence framework or an extra planning deliverable merely to appear thorough.

## C.2 Independent source review

> Review the exact held candidate against the current public contract and accepted predecessor. In your separate context, compare a relevant reference where available, judging both against current requirements. Do not send reference implementation code to the student.
>
> Begin with ownership, dependency direction, reuse, readability, and scope. Identify unjustified complexity even when tests pass. Distinguish required safeguards from speculative machinery. Verify source claims before reporting them.
>
> Check whether the tests exercise the actual owner or public boundary. A setup failure, compile failure, or unselected test is not behavioral sensitivity. Grader assumptions must also be checked against the public contract.
>
> Return concrete blockers, nonblocking observations, and a clear next decision. Do not invent requirements, redesign the chapter, or run unauthorized campaigns. Retain this context for the correction delta and final evidence rather than repeating the whole review after every change.

## C.3 Bounded correction

> Correct only the listed demonstrated defects, preserving the accepted architecture unless a specific architectural change is required. Use focused actual-path regressions where they resolve the finding. Preserve the original failed result and distinguish test repairs from production changes.
>
> Stay within the chapter’s aggregate correction limit. Do not reset it by changing workers or labels. Stop if the correction reveals consequential new scope or a conflict in the public contract.
>
> Return the held delta, actual focused results, and remaining limits. Do not automatically release full validation or external operations because a focused test passes.

## C.4 Changed-input validation

> Run the checks needed for the changed behavior and relevant inherited risks. Name the exact candidate and invocation. Inspect both the parsed result and process exit. Allow deliberate long-running checks their real duration.
>
> Reuse valid evidence for unchanged inputs and label it as carried. Do not rerun successful campaigns merely to make the report look fresh. Stop at the first unexpected failure and preserve it; do not silently patch, retry, or reinterpret the campaign as passed.
>
> Classify the failure before changing the product, specification, fixture, or environment. Escalate a genuine ambiguity rather than guessing.

## C.5 A required operational exercise

> State the specific claim that requires the real interface, native subsystem, or external service. Use the existing driver where adequate. Define finite request, time, resource, and device limits and the exact authority for execution.
>
> Keep credentials out of files, arguments, logs, and unrelated child environments. Attach only to owned processes and profiles. Preserve the distinction between an edit, a build, an installation, a running process, and observed behavior.
>
> Stop on the first genuine failure. Save the partial observations and clean up only owned resources. Do not automatically retry, substitute a different model, weaken an assertion, or extend the budget.
>
> Report what was actually observed. Simulated callbacks are not native behavior; native events are not human hearing; geometry is not a complete visual judgment. An unsuccessful controller does not by itself identify a production defect.

## C.6 Minimal author pass

> Work from the accepted student result, its concise review, and verified findings. Preserve useful reasons, documented failures, voice, and evidence qualifications. Improve the chapter only where the actual attempt demonstrated a need.
>
> Keep the TL;DR, plain-words explanation, mechanism, and exercise consistent. Verify commands against the accepted layout. Do not add speculative requirements, future features, invented anecdotes, or narrative claiming an experiment that did not happen.
>
> No change is valid. Receive one formal text review and make at most one bounded response. Do not start a second design process under the name of editing.

## C.7 One formal text review

> Review technical accuracy, fidelity to the demonstrated experience, public-contract consistency, readability, and preservation of load-bearing explanation. Check commands and evidence claims, not only style.
>
> Report actionable required corrections separately from optional suggestions. Do not demand changes merely to contribute. The author has one response; the coordinator verifies whether the conditions were met. Material unresolved issues return to the human or coordinator rather than an indefinite revision loop.

# Appendix D. Source notes and publication work remaining

The paper’s account rests on the current repository instructions and concise chapter records, not on a claim that every raw transcript has been reread. The following are primary project sources:

- [Current Edition 3 execution procedure](edition-3-execution.md): lean role loop, direct-comparison gate, correction limits, and current-versus-historical authority.
- [Public student brief](edition-3-student.md): permitted inputs, actual invocation, and separation of student work from private references.
- [Chapter-writing procedure](../book/chapter-writing-procedure.md): contract structure and teaching requirements, with historical phase instructions explicitly subordinate to the current execution guide.
- [Voice guide](../book/voice.md): explanation/contract/evidence separation, story preservation, and editorial judgment rather than length quotas.
- [Primary Ensemble skill](../agent/skills/ensemble/SKILL.md) and [coding skill](../.agents/skills/ensemble-coding/SKILL.md): general operating discipline and project-specific architecture.
- [Preserved second-edition first attempt](edition-2-attempt-1.md): the abandoned scope expansion, its source-history preservation, and the limits of the reported cost account.
- [Restarted Chapter 1 review](edition-3-notes/ch01-restart-code-review.md): actual source comparison, timeout and input-lifetime evidence, reviewer retractions, and bounded acceptance.
- [Restarted Chapter 12 review](edition-3-notes/ch12-restart-code-review.md): ownership, grader-versus-product classification, external-tool withdrawal, operations, and bounded prose closure.
- [Restarted Chapter 13 review](edition-3-notes/ch13-restart-code-review.md): the ongoing speech-channel work, native-controller limitations, and the submission-ordering defect. This chapter is not presented as operationally complete.

Before publication:

1. Bill and CodeRhapsody should add and verify the original creation account in Section 15, and decide the final author credits.
2. Review whether the case examples are the most useful ones for a reader who has not followed Ensemble’s development. Replace them with better documented examples if appropriate; do not simply add more.
3. Freeze the cited instruction versions and convert repository-relative links into public links appropriate for the website. Some current records are still part of unfinished chapter work.
4. Recheck the status of the ongoing speech example. Preserve what was known at each failure rather than rewriting it as an inevitable path to success.
5. Decide which instruction appendices should appear on the main page and which should be linked supplements. This is a readability decision, not a request to discard useful instructions.
6. Keep claims of novelty, general productivity, cost reduction, and full-system reproducibility bounded by actual evidence. If a stronger comparative claim is desired, it requires a separate literature review or experiment rather than stronger adjectives.

No publication build, deployment, or change to the paused chapter workflow is part of this draft.
