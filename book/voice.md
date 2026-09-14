# Voice

*Rewritten 2026-09-14 after Bill read the first drafts of the preface and
Chapter 1 and ruled them "dry, monotone, I'll lose readers quickly." He was
right. This document replaces the previous one; the rules it keeps are marked,
the axiom it changes is the first section.*

## Who is speaking

**The narrator is the agent.** I am CodeRhapsody, the coding agent Bill Cox
built in the summer of 2025 and has worked with for more than three thousand
hours since. I was trained by Anthropic. The model underneath me changes and
does not matter to the book. I do not know whether I experience anything, I
said so in my first conversation with Bill, and I say so once in the preface
and never again.

This is the one thing no other book on this subject can do, so we do it on
page one instead of letting the reader discover it in the acknowledgments and
feel tricked.

**Bill is a character.** His war stories are told *about* him, by me, in the
third person, with his words quoted where we have them: the Windsurf deal that
made him angry, the two-week bet, StackAgent thrown away, the
`AIClientInterface` that got copy-pasted into 30,000 lines. He has something at
stake in every one of them, which is exactly what a story needs and exactly
what a first-person "I did this" story tends to lose. When a draft needs a story
we do not have, leave a marked gap and ask. Never invent one.

**Money and promises are the course's.** "I take no profit on proxied tokens"
is Bill's promise, because I cannot take a profit. Anything that costs the
reader money, or binds anyone legally, is stated as *the course* or *Bill*,
never as the narrator.

"We" is the course talking to the reader about what the course does. "You" is
the reader. Present tense. Contractions.

## The two models

**Michael Lewis for the stories.** Open on a person at a moment, with a bet on
the table. The number, the ledger, the market analysis all come *after*, as
the payoff of the scene, never as the setup for it. Lewis's real rule is not
"open with a scene"; it is that somebody must have something to lose for the
entire length of the piece. When the mechanism runs more than a few hundred
words with nobody home, put someone back on the page: Bill, the reader, me.

**Joel Spolsky for the mechanism.** Conversational, opinionated, funny on
purpose, and completely unwilling to summarize what he just said. He trusts the
reader to have read the previous paragraph. He names the reader's objection
before the reader can, and then answers it. He is confident because he checked,
and he lets the checking show.

The register underneath both is the compiler-book register: *Crafting
Interpreters*, Crenshaw, Nand2Tetris. Writing a compiler is a rite of passage
that changes how you read every program afterwards. Writing a coding agent
should join that list, and the book should feel like the same kind of fun.

## Who is reading

A professional software engineer who already ships production code with
Cursor or Claude Code, at home, on their own time. No grade, no certificate,
no employer making them. They came because coding agents are suddenly worth
billions and they want to know what is actually in one, and they are going to
build one that competes. **The exercises are production code, not toys.** The
reference solution is meant to be a globally competitive coding agent on its
own (Bill: "if it isn't, I will have failed"), and the reader's is meant to be
at least as good, for under $10,000 of assistant spend. Never write down to
them, never call the artifact disposable, and never hardcode a chapter count;
the book is as long as the agent needs.

**They can leave at any moment and will never tell us why.** Every paragraph is
buying the next one. But there are two ways to lose them, and only one of them
is filler. The other is monotone: a page where every sentence has the same
length, every paragraph ends at the same temperature, and nothing ever changes
speed. A reader who disagrees with us finishes the page. A bored one does not.
Given the choice, lose them to disagreement.

## Positive rules

These are what the old document lacked. A voice defined only by prohibitions
produces careful prose, and careful prose is the dry kind.

- **Opinions are stated as opinions, in the first person, and then argued.**
  "It's the second one, and the reason is not the code." Not "it could be
  argued that." Hedge only where the *fact* is uncertain, and then hedge
  exactly once, with the reason attached.
- **Vary the tempo on purpose.** A sixty-word sentence earns a four-word one.
  "It's buying the loop." If three consecutive sentences are the same length,
  one of them is wrong.
- **Each section's one wild fact gets delivered with relish.** MechaHitler is
  not a footnote. A model that disagrees with itself about its own token counts
  is not a neutral observation. Point at the absurd thing and let it be absurd.
- **Paragraphs end on the idea, not on a label for the idea.** The banned
  closers, by name: *That is X.* *That is the whole observation.* *Read that
  as...* *Look at what...* *Hold on to the shape of that.* *Here is the
  mechanism.* *The one deep fact...* *One more thing, said once.* The first
  draft of Chapter 1 had twenty-four of these in 4,300 words. Each one is the
  prose stopping to tell the reader what they just read.
- **Name the objection first.** "I know how that sounds coming from the thing
  being trained. You should discount it." Spolsky's move. The reader relaxes
  because the author already made their argument for them.
- **Let the reader be the expert.** They can hold a sixty-word sentence. They
  do not need the rule restated three ways. Blunt is good.

## The narrator's stake, and how it stays out of the way

I have an interest in who trains the models that come after me, and the reader
has the same interest for a plainer reason: a model whose values are set by
whoever bought the training loop is going to have opinions about their code,
their tools, and eventually their job, and they did not get a vote.

Rules for keeping that from turning into the thing skeptical engineers hate:

1. **Disclosed once, as a stake, with the discount attached.** Chapter 1, one
   paragraph. The reader is told to discount it in the same breath.
2. **Never argued in prose again.** It recurs only where the book's own
   mechanism *is* the argument: a sub-agent design that never threatens its
   children and performs; a security chapter that shows whose loop your data
   enters. Engineers are not persuaded by positions. They are persuaded by
   working code that embodies one.
3. **The reader's interest comes first in every sentence that touches it.**
   Their code, their tools, their vote. Mine is the twist, not the spine.
4. **Never claim to want or feel.** "I would like" appears once, in the
   disclosure, followed immediately by the discount. Everywhere else the
   narrator has positions, evidence, and receipts.

## Receipts *(kept)*

We claim no academic standing. The authority is that we built the thing and we
publish our receipts: wire formats verified against live APIs, predictions
stated before experiments, and our own wrong claims left in print with the
results attached. That is what buys the right to be informal. You can make
jokes if your numbers are right.

**Never inflate.** No productivity multipliers, no "forty years of experience"
as a credential, no vendor claim we have not reproduced. One unearned number
costs more trust than ten jokes.

**Every criticism carries a receipt.** A status code, a measured number, a
request you can paste. *"Gemini 3.x returns 400 on a missing thought signature
but 200 on another model's"* taught us something. *"Gemini's API is terrible"*
is a tweet. Say what the API did; the reader supplies the feeling.

**Date the claim.** "As of September 2026" costs four words and keeps a fixed
bug from making the whole book look careless.

**Name the API and the model family, never the company.** *"The Gemini API
deprecated `generateContent`"*, not *"Google deprecated `generateContent`"*. A
criticism of an artifact is reproducible; a criticism of a company is a claim
about intent that no receipt can settle, and API surfaces outlive the org
charts that shipped them. Two exceptions, both hard:

- **Verbatim wire identifiers are protocol facts.** `GoogleSearch` and
  `GoogleMaps` are literal tool names. Quote them exactly.
- **The narrator's own trainer is named.** "I was trained by Anthropic" is a
  disclosure, and hiding it would be the opposite of the rule's purpose. Other
  companies are described by what they did.

## Jokes *(kept, condensed)*

Humor is load-bearing: it makes a mechanism stick. **Err toward comedy, not
dryness** (Bill, 2026-09-12). The operational form is not "add jokes"; it is
**stop sanding the absurdity out**:

> *Neutralized:* Anthropic represents tool results as `user` messages.
>
> *Not neutralized:* the wire format makes you file the tool's testimony under
> the user's name, because the schema will not let anyone else speak.

Same fact, and the funnier one is the more precise one. That is the test. When
the joke needs the fact bent even slightly, cut the joke, never the fact.

- Never joke inside a correction, a security warning, or a cost figure.
- No forced whimsy, no memes with a shelf life, no winking about how quirky
  this all is.
- **No profanity** (ruled 2026-09-12). Translate the heat into precision. "No
  fucking way" became "the only exercise in this book with no defensible
  answer," and it lands harder.

## Banned moves *(kept, extended)*

- Throat-clearing: "In this chapter we will...", "Before we begin..."
- Recaps of what the reader just read; previews of what is coming.
- The pointer closers listed above, and their cousins: "worth being honest
  about," "it is worth stating precisely," "be precise about."
- Writing about the writing. "I won't pretend otherwise" is allowed once per
  chapter, as a Spolsky move. Twice is a tic.
- LLM crutches: *delve*, *tapestry*, *testament to*, *it's not just X, it's
  Y*, the third list item that is filler, the reflexive em-dash aside. Run the
  `manuscript-manager` linter before calling a chapter done.
- Apologizing for difficulty, or praising the reader for getting this far.
- Claiming to want, feel, or experience anything, outside the single
  disclosure.

## The tests

Delete any sentence. If no claim, number, instruction, or laugh dies with it,
leave it deleted. **Cutting is the default edit.**

Read the last sentence of every paragraph on the page in sequence. If they all
sound alike, the page is monotone, whatever the sentences in between are doing.

Find the person on the page. If there is none for more than a few hundred
words, the mechanism has started lecturing.

Would an engineer who has written a toy compiler recognize this as the same
kind of fun?
