# Seed: the GUI chapter

Status: seed, awaiting Bill's ruling on placement and stack. Proposed by Bill:
a GUI chapter after ch4 and before the sub-agent chapter, likely TypeScript.

---

## Bill's argument

> "Otherwise, students will start creating their own slash commands to do things
> like select models when using the command line. This book is about building
> **advanced** AI coding agents. IMO, they work better with a GUI."

## The stronger form of it

Slash commands are a user interface accreting inside a text box without anyone
admitting it is a user interface. Unversioned, undiscoverable, untested, and
they multiply. Model selection is always the first one, because it is the first
piece of *state* the user needs to see and change that is not part of the
conversation.

And there is a second symptom of the same missing surface, which is worse
because it looks like a feature: **status tools.** CodeRhapsody has
`agent_status`, a tool whose entire job is to tell the user the token count, the
context percentage, the session cost, and the time. That is a status bar. It is
implemented as a tool call because the conversation was the only channel
available.

That is the thesis in one observation:

> Without a GUI, the agent's metadata competes with the user's content for the
> context window.

Every status readout, every model listing, every slash-command help text is
tokens spent inside the very context you are trying to protect, to report on
that context. A GUI moves the metadata out of band. This chapter is not
cosmetics. It is context engineering, which is the spine of the whole book.

Supporting receipt, already in the notes: the ch2/ch3 reorder "dissolved the
`agent_status` kludge." When a design forces an awkward artifact, suspect the
ordering. An awkward artifact that *survives* a reorder is pointing at a missing
component instead.

## The real risk, and the answer

**Risk: a GUI is not gradable the way this book grades.** Our entire apparatus is
deterministic graders plus P9 deletion audits. Pixels defeat that, and a chapter
that cannot be graded breaks the model the student has relied on for five
chapters.

**Answer: do not grade the GUI. Grade the seam it rides on.**

The chapter's actual content is not "build a React app." It is: *your agent needs
an observation and control surface, so design that seam.* An event stream the
front end subscribes to, and a control channel it sends on. That is a wire
format, and a wire format is gradable exactly the way chapter 2 is gradable —
byte comparisons against a fixture, no browser required.

The TypeScript client is then the **reference consumer**, shipped and runnable,
but not the graded artifact. Benefits:

- It rhymes with ch2 and the student will hear it: *one log, three vendors*
  becomes *one event stream, N front ends.* Same lesson, second application.
- It ages well. Framework churn is the fastest-rotting content in any
  programming book. The seam outlives the framework; the graded artifact never
  mentions React.
- A terminal client and a web client can both ride it, which is itself the proof
  the seam is real.

## Placement

**RULED 13 Sep 2026 (Bill): the GUI is Chapter 5.** That places it immediately
after jobs and *before* actors, which inverts the proposal below. The detailed
conversation is deferred; this section records the ruling and one consequence
that turns out to favor it.

**The consequence, and why I now think Bill's order is the better one.** My
argument was that actors should come first so the hint channel exists before the
surface that uses it. Reversing them produces something stronger. A GUI built on
a single-threaded blocking agent can *watch* but not *interject*: the moment the
agent starts a long tool call, the window sits there, inert, and the user's
typing goes nowhere. In a terminal, blocking feels normal, because terminals
block all the time. In a GUI, a frozen window is self-evidently broken.

So Chapter 5 ends with the student staring at a UI that is visibly, annoyingly
wrong, having built it themselves. Chapter 6 then introduces the mailbox to fix a
pain they have actually felt rather than one the author asserts. That is the
better teaching order, and it is the same move the book already makes in Chapter
1, where the naive format is built specifically so Chapter 2 can demolish it.

Consequence to carry into the ch6 outline: the actors chapter now inherits a
concrete opening scene and does not have to manufacture motivation.

### Original proposal, superseded (kept for the record)

Proposed: **immediately after the actors chapter, not before it.**

Actors introduces the mailbox and makes a hint an event. The GUI is what lets a
human *produce* that event at human speed. Ordering it before actors would mean
building a surface for a channel that does not exist yet.

This gives the actors chapter a genuinely good closing line: you now have a hint
channel and no way to type into it.

It also sets up the sub-agent chapter, which currently has an open question it
cannot answer — surfacing a child's stream so a human can supervise it. That is
a GUI question. With the GUI chapter in place, the sub-agent chapter can propose
a real mechanism instead of an aspiration. See `subagents-seed.md`.

Working order under this proposal:

1. naive loop
2. the LLM seam
3. six tools
4. jobs
5. **GUI (observation and control surface)**  ← RULED
6. actors and hints
7. capability seam and MCP
8. skills
9. security
10. sub-agent management

## Stack: argue for framework-free TypeScript

Bill said TypeScript, which is right — the browser is the portable surface and
the student already has one.

Recommend **no framework**: plain TypeScript, no React, no build pipeline beyond
`tsc`, served by the Go server the student already wrote. Reasons:

- The book's cold open is "the money says cathedral, the source says a loop and
  a log." A chapter that requires React, a bundler, and a CSS framework to
  render a list of events contradicts the thesis in its own exercise.
- It keeps the diff readable, which is how every other chapter is taught.
- CodeRhapsody's GUI is React, and that is fine for a product. A book chapter has
  different constraints: it must still compile in three years.

The thing worth spending complexity on is not the widget layer. It is rendering
a *streaming* response without flicker, and accepting input while the agent is
mid-turn.

## Candidate declined decision (P6)

What the surface does when the user types while the agent is mid-turn and the
mailbox is not empty: coalesce the messages, deliver them in order, or drop all
but the newest. All three are defensible, all three are visible to the user, and
the chapter should teach enough to decide without deciding.

(Confirm this does not collide with the actors chapter's own declined decision
before committing to it.)

## Open questions for Bill

- Chapter number: is this 6, or does it go earlier/later?
- Is the graded artifact the event-stream seam, as argued above, or do you want
  something rendered actually graded?
- Transport: SSE, WebSocket, or long-poll? CodeRhapsody's `gui_api.log` is a
  receipt we can draw on either way.
- How much of CodeRhapsody's real GUI design is fair game as an exhibit?
