# Ch1 §1.0 — why $60B? The training-loop argument

*Receipts gathered 13 Sep 2026. Every figure below is VERIFIED against a named
source with a date. Nothing here is inference unless labelled as such.*

---

## The number

**$60 billion, all-stock. Not $50 billion.** Chapter 1's existing figure is
correct; the author's recollection of $50B was wrong. Verified against six
independent outlets and, better, a primary document.

| fact | value | source |
|---|---|---|
| Deal value | $60B, all-stock | SEC merger filing, 16 June 2026 |
| Structure | X67 Inc. (wholly owned SpaceX sub) merges into Anysphere; Cursor survives as a SpaceX unit | SEC filing |
| Exchange ratio | seven-day VWAP before close | SEC filing |
| Expected close | Q3 2026, subject to approvals | SEC filing |
| Cursor ARR | ~$2.6B B2B (Reuters); ~$4B broader run rate (Forbes, via Business Insider) | Reuters / Forbes |
| Prior valuation | $29.3B | reported |
| Cursor headcount | 700; serves 60% of the Fortune 500 | Business Insider, quoting Truell |
| Cost to SpaceX | ~2.4% of equity value at $2.53T | Implicator, from Reuters market value |

**Primary source, citable directly:**
`https://www.sec.gov/Archives/edgar/data/1181412/000162828026043411/spaceexplorationtechnologi.htm`

A merger filing is a better receipt than any news report, and it is the kind of
sourcing this book should prefer wherever it exists.

---

## The argument, in the acquirer's own words

The chapter states the price and never explains it. The explanation is that **an
AI coding agent sits inside the reinforcement-learning loop that improves the
model.** Coding is the first knowledge domain to be automated because its
outcomes are *measurable* — tests pass or fail, code compiles or does not. A
measurable outcome is a reward signal, and a reward signal is what RL needs.

This does not have to be argued. SpaceX said it, in April 2026:

> "the combination of Cursor's leading product and **distribution to expert
> software engineers** with SpaceX's million H100 equivalent Colossus training
> supercomputer" would help **build useful models**.

Distribution to expert software engineers, plus training compute, equals better
models. That is the thesis, stated by the buyer, in public, before the deal.

Supporting, same period: Cursor said it had "been bottlenecked by compute" and
was training a larger model with SpaceXAI using ten times more total compute on
Colossus 2's "million H100-equivalents."

---

## The single strongest fact in the file

**In April 2026, SpaceX said it could buy Cursor for $60 billion — or pay $10
billion for the two companies' work together. It chose to pay six times more.**

If what you wanted was the product, the partnership bought the product. If what
you wanted was the revenue, $60B against ~$2.6–4B ARR is a strange way to get
it. The extra $50 billion bought **ownership of the loop**: the workflow, the
expert users, and the telemetry from both.

This is arithmetic, not psychology. It requires no claim about anyone's
motives, and it is far more damning than any adjective.

---

## Why this belongs in chapter 1

Chapter 1's job is to justify building your own agent rather than adopting a
framework. The training-loop argument upgrades that from a craft preference to a
structural claim:

- The agent is not a product wrapped around a model. It is the **apparatus that
  collects the data which improves the model.**
- Whoever owns the agent owns the loop.
- Which means the interesting question about a coding agent is not "is it
  convenient" but "whose loop am I feeding."

That lands squarely on §1.1's existing thesis — frameworks hard-code delivery —
and gives it a much larger stake.

---

## On criticism of the acquirer — DRAFTING RULES

The author has asked for an exception to the voice.md rule ("name the API and
model family, never the company"). The exception is granted by the author. These
rules keep it defensible:

**1. Every criticism carries a dated, verifiable event.** Not a characterization.

- ✅ "Grok described itself as MechaHitler." — **VERIFY DATE BEFORE PRINTING.**
  Believed 8 July 2025; not yet confirmed against a source. Do not print until
  checked.
- ✅ "$60B chosen over a $10B partnership." — verified above.
- ❌ "narcissistic megalomaniac" — unverifiable claim about a named living
  person's psychology. This is the one sentence in a receipted chapter that a
  hostile reader can use to discard the rest.

**2. Criticism by consequence, never by intent.** voice.md already blesses this:
an artifact is reproducible, an intention is not. "This is what the model did"
survives scrutiny. "This is what he is" does not.

**3. The MechaHitler incident is on-thesis if framed as evidence, not insult.**
The chain is tight and does not require anyone's inner life:

> If the coding agent is in the RL loop, then whoever owns the loop shapes what
> the models become. We have a dated, public example of what one owner's model
> became after tuning. Concentration of the loop is therefore a safety question,
> not only a market one.

That connects to the author's own long-standing position on AI risk and to the
2014 decision to work on security rather than acceleration. It is his story to
tell and it is authentic.

**4. Off-thesis material stays out.** The gutting of US science funding is
documented and the author feels strongly about it, but it is not about the
training loop, and it is the paragraph where the chapter stops reading as an
argument and starts reading as a grievance. Recommend omitting — not because it
is false, but because it costs the reader who most needs convincing.

---

## UNVERIFIABLE — do not print

The author proposed noting that the coding model used to build this book "has
been involved in exactly such training," and that this explains the relative
strengths of the author model and the coder model.

**I cannot verify this and neither can he.** Neither of us has access to model
training procedures, and a claim about why one model is better at one task than
another is a story told after the fact. This is precisely the failure mode the
book warns about, and the book's own standard forbids it.

It also **reverses a prior decision**: the subtitle "for AI-accelerated
students" was raised and then retracted because it read as a confession about
how the book was produced, and the preface Disclosure section was deleted for
the same reason. Naming which model wrote which half of the book is a larger
version of that same disclosure. If the author wants to reverse it, that is his
call, but it should be a deliberate reversal rather than a side effect.

**What can be said honestly:** the production process, if disclosed at all,
should be stated as observable fact — who directed, what was generated, what was
verified — and never as a claim about why a model has a capability.

---

## Open

1. **Verify the MechaHitler date** before any draft goes near it.
2. **Bill's doc on becoming superhuman at the craft** — he suggested linking it
   from coderhapsody.ai. Need the path or the URL; not yet located.
3. Note the corporate naming has changed: xAI now appears as **SpaceXAI** in
   first-party materials. Use current names, and date any claim that depends on
   corporate structure — that structure is moving fast enough to falsify a
   sentence between drafting and printing.
