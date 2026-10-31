# Knowledge Patches: What If Bug Fixes Made Your Team Smarter?

*Working paper. Proposed by Bill Cox, drafted with Cadence and CodeRhapsody.*

---

## The problem with patches

Your team fixes a restart bug. A counter was supposed to generate unique IDs, but after a restart it started over from zero and collided with IDs saved to disk. Someone adds a fix — persist the counter, or switch to UUIDs — and the PR lands. Tests pass. Everyone moves on.

Six months later, a different developer builds a new feature that also needs unique IDs. They look at the existing code for patterns to follow, find `nextID++`, and copy it. Same bug, different module. The *code* was fixed, but the *lesson* — that identity must survive process restarts — was never captured anywhere a future developer would find it.

This happens constantly. Every mature codebase is full of fixes whose reasoning exists only in a PR description nobody will read again, a Slack thread that's since been archived, or the head of someone who's left the company.

**A Knowledge Patch is a proposal to fix that.** Instead of just patching the code, you also patch the *instructions* that would teach someone to build it correctly in the first place.

## What is a Knowledge Patch?

A regular patch says: "Change line 47 from `counter++` to `uuid.New()`."

A Knowledge Patch (KP) says:

- **The problem:** IDs generated with a simple counter collide after a restart because the counter resets to zero.
- **The lesson:** Any identifier that will be persisted or shared across process boundaries must be unique across the *lifetime of the data*, not just the lifetime of the process.
- **The fix:** Use UUIDs (or persist and resume the counter).
- **What we rejected and why:** We considered using PIDs as a namespace prefix, but PIDs also recycle.
- **How to catch this in the future:** Any test that exercises persistence should include a restart boundary — save, stop, restart, load, and verify that new operations don't collide with old data.
- **Where this belongs:** This lesson should be part of whatever teaches new developers about your persistence layer — not appended as a footnote at the end.

That last point is the key insight. A KP doesn't just record what you learned. It asks: *where should a new developer encounter this knowledge so they never make this mistake?*

## Why not just write better documentation?

Documentation is great when people read it. The problem is that traditional docs rot:

- They describe the system as it was when someone last updated them.
- They sit in a wiki nobody visits until they're already lost.
- They have no feedback loop — nothing tells you when they've become wrong.

A Knowledge Patch is different because it's *testable*. The claim isn't just "here's what you should know." The claim is: "If you follow these instructions, you can build a working system that passes these specific checks." If the instructions are wrong or incomplete, the checks fail. That's a feedback loop that documentation doesn't have.

Think of it like test-driven development, but for knowledge instead of code. The "test" is: can someone follow these instructions and produce something that works?

## The two-lane verification

This is where the idea gets interesting. When a Knowledge Patch is submitted, it gets verified in two separate ways:

### Lane 1: Fix the current code

An experienced developer (or AI agent) takes the existing codebase and applies the fix. This is the familiar workflow — patch, test, review, merge. It answers the question: *does the fix work?*

### Lane 2: Rebuild from the instructions

A *separate* developer (or AI agent), who has never seen the fix, tries to build the relevant part of the system using only the updated instructions. They start from the documented prerequisites and follow the teaching. This answers a different question: *are the instructions good enough?*

Both lanes have to pass. Lane 1 without Lane 2 means you shipped a fix but didn't capture the knowledge. Lane 2 without Lane 1 means your documentation is lovely but untested against reality.

**Why does Lane 2 matter?** Because it catches the most insidious kind of knowledge loss: the fix that works because the developer *already knew* something that isn't written down. If the Lane 2 builder succeeds without needing to ask questions or peek at the existing solution, the instructions are genuinely self-contained. If they fail, you've found a gap in what you're teaching.

## Put the fix where it belongs

Here's a subtlety that matters more than it sounds. When you learn something from a bug fix, the temptation is to append it to the end — add a new section to the docs, tack on a "lessons learned" page, create a post-mortem. But that means every new developer has to read through the entire history of your mistakes to avoid repeating them.

A Knowledge Patch asks you to put the lesson at the *earliest point where it becomes relevant*. If the restart-ID bug is fundamentally about persistence, the lesson belongs in whatever teaches your persistence layer — not in a "bug fixes" appendix.

This is conceptually similar to `git absorb` or `hg absorb`, which attribute fixes to earlier changes. Here, however, the object being integrated is *knowledge*, not submitted implementation history. You're asking: what is the earliest valid point in a developer's learning journey where this lesson belongs and its required prerequisites are available? Accepted code commits remain unchanged; the implementation repair lands on top of HEAD.

This is also the hardest part to get right. It requires understanding what depends on what, and it sometimes means restructuring how you teach things. But the payoff is that your instructions stay clean and integrated rather than accumulating an ever-growing errata section.

## Three linked histories, without rewriting submitted code

An enterprise adopting this workflow does not need to permit rebasing submitted implementation CLs. It needs three linked records:

| Record | What it preserves |
|---|---|
| **KP history** | The submitted knowledge change, its revisions, rationale, review, approval, and integration outcome. |
| **Codebook history** | The integrated chapters, graders, and construction prerequisites at each accepted revision. |
| **Implementation history** | Ordinary CLs that patch the current codebase on top of HEAD. |

**One KP can generate many CLs. Each CL belongs to exactly one KP.** This is a one-to-many ownership rule, not a many-to-many attribution graph. If a proposed CL mixes work from two KPs, split it. A CL may depend on another KP's implementation without acquiring a second owner.

Each KP has a stable ID. Each generated CL records that ID and the exact KP revision it implements. The KP view links back to every generated CL, its disposition, the resulting codebook revision, and the validation runs. A human should be able to inspect the KP history, read a submitted or approved KP, open the integrated book, and follow its implementation links without reconstructing those relationships from commit prose.

Approval attaches to a particular KP revision. Editing it cannot silently carry approval forward. Approved knowledge, merged implementation, and verified reconstruction are separate facts: none should be inferred merely from the other two.

### HEAD checks do not replace construction checks

The incremental lane patches HEAD and runs all applicable chapter graders plus the implementation's unit and integration tests. Tests exercise the real behavior, using faithful fakes at external boundaries rather than mocks that replace the behavior being checked. Every affected check must run against the exact proposed integration result; old green results do not establish that two concurrent CLs work together.

The reconstruction lane runs in isolated workspaces. It does not rebase or rewrite submitted CLs. It checks that affected exercises can still be built from their declared prerequisites. The periodic clean-build train checks the entire integrated book.

This separation matters because HEAD contains everything. A chapter can pass its tests at HEAD while depending on a capability that a student following the book has not built yet.

### The construction graph must be acyclic

The book's linear build order is an ordering of a dependency graph. Each exercise identifies the capabilities it requires and the capabilities it establishes. Those construction prerequisites must form a directed acyclic graph. A module-import rule alone is not sufficient: prose, fixtures, tests, and assumed runtime behavior can introduce dependencies too.

Consider changing memory compaction to use managed sub-agents. If the memory chapter comes before the fleet chapter, the current implementation may be easy to patch: HEAD already has both. A fresh student cannot use fleet infrastructure before constructing it. The codebook must resolve that dependency, for example:

```text
Context/history primitives
    -> Sub-agent lifecycle
    -> Background memory compaction
```

If sub-agents need some earlier memory functionality, separate the actual prerequisite from the later integration:

```text
Context/history and minimal persistence
    -> Sub-agent lifecycle
    -> Background compaction using sub-agents
```

A KP that creates a cycle must trigger an explicit design decision: move or split a chapter, extract a shared foundation, change the design, or add a small integration chapter after all its prerequisites. Placing a change after its last prerequisite works only once the graph is acyclic; a cycle has no valid "last prerequisite."

A small integration chapter can be permanent, useful teaching rather than historical errata. It introduces behavior that becomes possible only when earlier capabilities exist. The current book should teach that valid construction, not force every new builder to replay an obsolete workaround unless the staged approach is deliberately part of the curriculum.

Checking the declared graph catches explicit cycles. **Executing each affected exercise with only its declared prerequisites catches hidden dependencies.** A fresh student gets the revised teaching and accepted prerequisite implementation, not the target exercise's existing answer or future chapters' solutions. These checks are what make the graph credible.

### What knowledge absorption means now

Absorption integrates the KP into the current codebook at valid teaching points and updates its prerequisite graph. It does not modify the ancestry of accepted code commits. New documentation and implementation CLs advance their respective histories normally.

The accepted codebook remains an integrated set of executable chapters. The KP remains an immutable reviewed record with revision history and links to what it generated. Independent reconstruction validates the new recipe; patching HEAD upgrades the existing installation. Neither operation needs to pretend that the enterprise's implementation history was originally built in the new order.

## Humans have to actually learn

If you're using AI agents to help with development — and increasingly, most teams are — there's a trap. The AI can investigate a bug, figure out the root cause, write the fix, update the tests, and submit the PR. The human reviews it, sees green checks, and approves. Everything works.

But the human didn't learn anything. They can't explain *why* that fix was necessary, what alternatives were considered, or what the broader architectural implication is. The next time a similar decision comes up, they can't make it — they have to ask the AI again. Over time, the human's understanding of their own system falls behind, and they lose the ability to make good architectural judgments.

A Knowledge Patch workflow addresses this head-on: **agents must explain what they're discovering while they work, not just deliver results.** Specifically:

1. **Before acting:** Here's what I'm going to change and why I think this approach is right.
2. **During investigation:** I found something unexpected — this assumption was wrong, or this alternative didn't work because of X.
3. **At completion:** Here's the lesson, here's what I'm still uncertain about, and here's what you should decide.

This isn't about generating verbose logs. It's about surfacing the *material discoveries* — the things that change your understanding of the system. A concise explanation of a changed invariant is worth more than a thousand lines of command output.

The point is that a development workflow that produces correct code while leaving its human participants ignorant of the design has failed part of its purpose.

## The clean-build test

Even with careful knowledge management, undocumented assumptions creep in. Someone fixes something during a debugging session and forgets to write it down. An AI agent uses context from a previous conversation that isn't captured in the instructions. A fix works because of a side effect nobody noticed.

The clean-build test catches these gaps. Periodically — nightly, weekly, whatever fits your budget — start from scratch. Give a fresh builder nothing but the current instructions and see if they can produce a working system. No access to the existing codebase. No historical conversations. No "oh, you also need to do X" hints.

When it fails, you've found undocumented knowledge. That becomes a new Knowledge Patch, the instructions get updated, and next time the clean build goes further.

This is expensive, which is why you don't do it on every commit. But it's the only honest test of whether your instructions are actually complete, or whether they quietly depend on things only your current team knows.

## The workflow, step by step

The concepts above describe *what* Knowledge Patches are. This section describes *how* to run the workflow, including the mistakes we made while developing it so you don't repeat them.

### Step 1: Write the KP before writing the fix

Start with the Knowledge Patch document, not the code. Describe the problem, the lesson, where the lesson belongs in your instructions, and how to verify it. This feels backwards — you want to jump to the fix — but writing the KP first forces you to understand the problem well enough to teach it.

**Trap we hit:** When we wrote the fix first, the KP became a retroactive justification of whatever we'd already built. The "rejected alternatives" section was empty because we never considered any — we just coded the first thing that worked. Writing the KP first means you've thought about alternatives before committing to one.

### Step 2: The fix lane — patch, test, verify

An experienced developer (or AI agent) applies the fix to the current codebase. Standard workflow: patch, test, review.

But the tests need to actually work. This sounds obvious. It isn't.

**Trap we hit:** Our automated checks passed everything. 100% score, green across the board. We celebrated. Then we deliberately broke the code — deleted the feature the check was supposed to verify — and the check *still passed*. The checks were testing that the code compiled and ran, not that it did the right thing.

**The fix: mutation testing for your checks.** For every check in your verification suite, deliberately break the specific behavior it's supposed to catch and confirm the check fails. If you delete the restart-safe ID generation and the persistence test still passes, your test is decoration. A check you haven't tried to kill is a check you don't know works.

This is the single most important quality gate in the workflow. A passing test suite is evidence of nothing until you've proven each test can fail.

### Step 3: The rebuild lane — fresh builder, instructions only

A *separate* builder — one who has never seen the fix or the existing code — attempts to build the relevant part of the system using only your updated instructions. This is the test of whether your KP actually teaches what it claims to teach.

**Trap we hit:** The fresh builder had access to the existing codebase "for reference." They copied patterns from it. The instructions had gaps, but the rebuild succeeded anyway because the builder filled them in from the existing code. We thought the instructions were complete. They weren't — they only worked with a cheat sheet.

**The fix: strict isolation.** The rebuild builder gets the instructions, the prerequisites, and nothing else. No access to the current codebase. No access to the fix. No "just peek at how the existing code handles this." If they can't build it from the instructions alone, the instructions are incomplete. That's the signal, not a problem to work around.

### Step 4: Independent review of the KP

A reviewer who participated in neither lane reads the Knowledge Patch and evaluates it. They're reviewing the *teaching*, not the code. Questions they should ask:

- Is the lesson placed at the right point in the instruction sequence, or is it tacked on at the end?
- Are the rejected alternatives genuine, or is this a rationalization of whatever was built?
- Does the verification actually test the claimed behavior? (Step 2's mutation testing should already answer this, but the reviewer should confirm.)
- Are there open questions that need a human decision?

**Trap we hit:** Early on, the reviewer was the same person who wrote the fix. They approved their own KP because it made sense to them — of course it did, they wrote it. An independent reviewer immediately found three gaps that seemed obvious in hindsight.

### Step 5: Accept gate — both lanes must pass

The KP is accepted only when:
- The fix lane passes all checks, including mutation-verified tests.
- The rebuild lane produces a working system from the instructions alone.
- An independent reviewer has signed off on the KP itself.
- The chapter construction graph is acyclic, and affected exercises succeed using only their declared prerequisites.
- The final implementation has passed the project's independent code review and malware/security inspection at the exact revision being promoted.

Once those gates pass, agents may create accepted local CLs under the project's standing authorization, each with one owning KP. Push, merge, deployment, and release remain separate permissions. Any later content change invalidates the checks and approvals it affects. Inspection is evidence, not a guarantee that arbitrary code is harmless; tests and builds must also run within appropriate safety boundaries.

If any of these fail, the KP goes back for revision. No exceptions. The temptation is to ship the fix and "clean up the instructions later." Later never comes.

### Step 6: Preserve immutable acceptance evidence

Once accepted, retain the exact codebook, grader, prerequisite-source, and resulting-source identities, together with the build configuration and actual results. This establishes which instructions produced which tested implementation.

Existing version-control objects and protected references can preserve the source; we do not need a separate copied solution directory for every chapter. Temporary reconstruction workspaces can be discarded once the required evidence is retained. Previously published snapshots remain historical evidence and must not be relabeled or overwritten.

Without those identities, a later failure leaves us guessing which KP changed the outcome. With them, we can investigate the regression without rewriting accepted implementation history.

### Step 7: Absorb into the instruction set

The KP's lesson gets integrated into the instructions at the point identified in Step 1 — not appended to the end, but placed where a future builder would need to encounter it. This is the "semantic absorption" described earlier.

The proposed instruction changes must already be present in the candidate used by the rebuild lane. Acceptance promotes that tested revision; it must not introduce an untested rewrite after verification.

Mark the KP as integrated and retain it, its review history, and its one-to-many CL links. Do not delete it. The lesson now lives in the current instructions; the KP preserves why and how those instructions changed. A fresh builder follows the integrated codebook, not a backlog of historical KPs.

### What the workflow produces

After running this cycle a few dozen times, you have:
- An instruction set that a fresh builder can follow to produce a working system.
- A verification suite where every check has been proven to catch the behavior it claims to test.
- Immutable acceptance records linking tested codebook revisions to the source and results they produced.
- A history of rejected alternatives that prevents revisiting dead ends.

The instructions are not documentation in the traditional sense. They're a *tested, executable specification* of your system's design knowledge. They rot slower than code comments, slower than wiki pages, and slower than Slack threads — because they have a feedback loop that catches rot.

## What a KP contains, concretely

For engineers who want to try this, here's what goes into a Knowledge Patch:

| Section | What goes in it |
|---|---|
| **Identity and provenance** | Stable KP ID and revision, base codebook/source revisions, approval record, and links to all generated CLs. Every CL has exactly one owning KP. |
| **Problem** | What broke or what's missing, with evidence (a failing test, a reproduction, a user report). |
| **Lesson** | The broader principle — not just "this line was wrong" but "here's the rule we were violating." |
| **Instruction change** | How the teaching should be updated: what to add, what to revise, where it belongs. |
| **Verification change** | New or updated tests/checks that would catch this problem. |
| **Rejected alternatives** | What you considered and why you didn't do it. (This is often the most valuable part.) |
| **Impact range** | Affected chapters, their required and established capabilities, prerequisite-graph changes, and the checks and reconstructions that must rerun. |
| **Open questions** | Things you're not sure about, decisions that need human input. |

A small fix might be a few paragraphs. A significant architectural lesson might be substantial. The size should match the significance of the discovery.

## What this is and isn't

**This is a proposal**, not a product. It describes a workflow that a team could adopt incrementally. Some pieces — like the two-lane verification and clean-build testing — can be done manually today with existing tools. Others — like automated semantic absorption into the right place in a teaching sequence — are research problems.

**This is a massive reduction in technical debt, not its elimination.** When every fix produces a tested, teachable lesson — and the system can be rebuilt from those lessons alone — most of what we call "tech debt" simply doesn't accumulate. Undocumented workarounds, tribal knowledge, the fix nobody remembers the reason for: these are the bulk of real-world debt, and Knowledge Patches address them directly. The instructions and their tests can still accumulate their own contradictions and stale assumptions, so debt doesn't vanish entirely. But it shrinks from the dominant cost of software maintenance to a manageable fraction of it.

**This is a replacement for code review as we know it.** Today, a human reviewer reads the diff — every line of implementation — and gives an LGTM. As AI generates more code faster, that becomes the bottleneck. Reviewers skim. Bugs ship. The review queue grows faster than the team can drain it.

Knowledge Patches move the LGTM to the right artifact. Instead of reviewing implementation code line by line, humans review the KP: the data structures, the API boundaries, the design decisions, the rejected alternatives — the things you'd see in a computer science textbook. The actual code is written and reviewed by AI agents, with human oversight over the agents' process, not micro-management of their output. We'll need better malware detection and stronger agent supervision, but not a human squinting at every `for` loop.

This isn't the end of human judgment. It's the end of human judgment being spent on the wrong thing.

## Starting small

If this interests you, don't build a platform. Try one Knowledge Patch:

1. Pick a recent bug fix where the lesson was broader than the diff.
2. Write down what the next developer needs to know to avoid that class of bug.
3. Put that knowledge where it belongs — not in a post-mortem, but in whatever a new team member would read when working in that area.
4. Write a test that would catch the bug.
5. Ask someone unfamiliar with the fix to follow your updated instructions and see if they can build that part of the system correctly.

If step 5 works, your instructions are good. If it doesn't, you've found exactly what's missing. Either way, your team just got smarter.

---

*This paper describes a workflow to implement and evaluate. It does not claim that its complete automation already exists, and it makes no claim of universal superiority or the elimination of all technical debt.*
