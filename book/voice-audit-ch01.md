# Voice audit — chapter 1

Pass run 2026-09-13 against `book/voice.md`, after Bill's tic-apply pass and
after the grader-audit rulings in `bc54c4f`. Line numbers are as of commit
`bc54c4f` + this pass; quote the text, not the number, if you act on this later.

## Applied

1. **§1.0 profanity — RULED VIOLATION.** "a vibe-coded pile of shit I had to
   throw away" → "a working proof of concept, vibe-coded, that I threw away in
   its entirety." `voice.md` bans profanity outright (ruled 2026-09-12) and
   argues the dry version is funnier. It is: "threw away in its entirety" is a
   bigger admission than the swear, because it is specific. The heat becomes
   precision, which is the rule's stated form.

2. **§1.2 false continuity, introduced by me an hour earlier.** "the line that
   reads position zero will be four chapters old and entirely trusted" → "will
   be old code that you trust." The original promised the reader that this
   code survives four chapters. Chapter 2 throws it away, so the sentence was
   false, and it leaked the fact that the program continues — a soft breach of
   the **ambush ruling** (ch1 gives no hint that ch2 demolishes it). Deleting
   the count removes both problems and loses nothing.

3. **§1.7 undated pricing claim.** "Anthropic (last checked) wants ~$400" →
   "Anthropic's API tiers (as of September 2026) want ~$400". `voice.md`:
   *date the claim*, because a stale complaint tells the reader our
   measurements have an expiry we did not track. Also shifts the subject from
   the company to the artifact (API tiers), per the name-the-API rule.

4. **§1.7 bullet head that contradicted itself.** "You cannot cheaply get a key
   that goes fast enough — eventually." → "Cheap keys exist. Fast ones do not,
   and that only bites you later." The original hung "eventually" off the end
   of a claim it was busy qualifying, so the head said one thing and the body
   said another.

5. **§1.7 meter 3 repetition.** "...to help you write it, the assistant spend
   it takes an AI-accelerated engineer to produce a coding agent good enough to
   replace Claude Code or Codex" named the same two products twice in one
   sentence and smuggled in "AI-accelerated engineer," which is the flavour of
   self-congratulation `voice.md` bans. Now: "...to help you write it: the
   assistant spend it takes to produce a coding agent good enough to replace
   them."

Measured after: profanity 0, em-dashes 22 → 20 (I added none in the rulings
pass either; verified HEAD vs working tree, not by eye).

## NOT applied — these are Bill's calls, not the editor's

**V1. The §1.0 deal figures are still unverified, and they are the first
numbers in the book.** SpaceX/Anysphere ~$60B, the three-way Windsurf shape,
OpenAI/Ona, Anthropic/Bun. The outline carries a publication note to check them
against primary sources, which is correct and not yet done. Flagging it here
because of where they sit: `voice.md` stakes the book's entire authority on
"we publish our receipts," and *"one unearned number costs more trust than ten
jokes."* These are the numbers a reader meets before any code. If one is wrong,
it is wrong on page one.

**V2. §1.7 contains the book's only criticism of companies rather than APIs.**
"The providers' response is to own the high-value tools on top... The message
is explicit: use our agent; don't build your own." `voice.md` says a claim
about corporate intent cannot be settled by a receipt and so reads as a
grievance however carefully phrased. The mitigation already present is that it
is **receipted by consequence**: the token discount through first-party
products is observable and checkable, whoever intended what. Two honest
options, and both are the author's:

- Keep it. It is the book's motivating thesis and the consequence is real.
- Narrow it to the observable: state the price difference and let the reader
  draw the inference, which `voice.md` argues lands harder anyway ("the reader
  supplies the feeling").

I did not touch it, because rewriting an author's thesis paragraph on a voice
rule is a bigger decision than a voice pass gets to make.

**V3. Thirteen pre-existing lines run over 78 columns.** Cosmetic, untouched;
reflowing them would churn the diff against a chapter still under edit.
