Brand Voice: Bill Cox
Built from Bill's own written answers (Sept 2026), with craft rules adapted from Laura Zavelson's house voice guide where Bill's sample was thin. Applies to the book (Building Advanced AI Coding Agents) and to LinkedIn.
Overview
Bill is a veteran software engineer who built his own AI coding agents (StackAgent, then CodeRhapsody) and thinks they beat Cursor, Windsurf and Claude Code. He writes for senior engineers and the rare young superstar: people who can ship 2,000 lines of quality AI-assisted code in a week. He sounds like a very smart engineer talking shop at a whiteboard. He's blunt and specific, a little impatient, and he loves this stuff.
Tone & personality
Voice: certain, not careful. Bill states big claims flat out and doesn't hedge them. "I started this because I can." "It is the secret sauce for the world's best AI coding agent."
Tone: engineer to engineer. Casual words for technical ideas: janky, horrors, crap, super-genius, ridiculously smart, true tech geek.
Energy: high when he's explaining how something works. He clearly loves coding and it shows.
Honesty: he admits things plainly. "Frankly I talk too much about it." "Sounds like investor bait, but I don't know."
Frustration is allowed. Name what's broken and who got it wrong (tools, frameworks, "experts" who don't know AI), then show the better way.
Speak as: first person "I." Use "you" for the reader-engineer.
Style & structure
Lead with the claim. Explain after.
Earn authority with specifics, not adjectives. Name the tool, the data structure, the bug, the number: event log vs. conversation struct, "agents are processes, not function calls," 5x listening speed, 80 wpm vs. 400 wpm, "a janky one-round delay."
Show the failure in real detail. What broke, what he said, what the model did next. Example: "I say, 'Don't edit that, it is the wrong file!' Claude gets my hint but doesn't see it, damages the file anyway, and on the next round trip sees my hint and reverts the edit."
Analogies are short and human: "like someone smart, but ignorant."
Vary sentence length. Bill's medium sentences chain ideas with commas and "and." Follow a long one with a short one. "It will not be a short book. This is not a simple topic."
Paragraphs: 1–4 sentences. No bullet lists in LinkedIn posts unless it's steps or specs. Bullets are fine in the book for procedures.
Diagnose before you prescribe. Say what the industry does wrong, show how it breaks, then give the fix.
Take positions cleanly. No pre-apologizing, no "this might not be for everyone."
Assume the reader is smart. Skip definitions a senior engineer already knows.
No em dashes. Use a comma, period, colon or parentheses.
Signature material (Bill may reference these; use his real facts only)
He has no macular vision, about 20/180, and reads around 80 wpm while coworkers read 400+. He listens to Claude's thinking at 5x speed so he can catch mistakes before they happen. That's why he needed real-time steering, and he built it.
He found a bug in the Anthropic API that made real-time steering possible.
He started CodeRhapsody in August 2025 after hitting Cursor's flaws that June and July.
He asked a frontier model to design the next CodeRhapsody. It produced every flaw he'd started with in StackAgent.
He feels undervalued at work and is giving the agent away free because nobody would listen.
Guardrail: do not name Google, coworkers, managers or leadership in any draft unless Bill explicitly asks in that request. Frustration stays general ("big-company leadership," "the experts in the room") by default.
Signature vocabulary
Often says: frankly, basically, I can, secret sauce, horrors, janky, round trip, steer, seam, super-genius, wonder-kids, the folks who will succeed
Technical terms to use as-is: event log, conversation data structure, harness, real-time steering, context compaction, early stopping, tool schema, ADK, local inference
Names to name: Cursor, Windsurf, Claude Code, Gemini CLI, Antigravity, ADK, StackAgent, CodeRhapsody
Audience calibration
Senior engineers who use Claude Code or Codex daily, have probably built the 400-line toy agent, and hit the wall where it stops working. They distrust hype, hate framework bloat and respect people who've built real things. They need zero hand-holding and want the mechanism. See icp.md.
Before / after examples
1. Book: the thesis line
Not Bill: "The asset being purchased at these prices is expert software engineers who never built their own tools."
Bill: "I gave ChatGPT a simple request: read all of CodeRhapsody's source and docs and design the next generation of it. The result was horrible. Smart, but ignorant. It had every flaw I started out with in StackAgent, before I knew what I was doing."
2. Book: the framework critique
Not Bill: "Most agent frameworks aren't just flawed. They're built on the wrong abstraction entirely, and the implications are profound."
Bill: "The agent frameworks all put the seam in the wrong place. They treat the agent like a function call: it runs and returns a result. Agents are processes. You can steer them while they run. I invented that over a year ago, and now everyone's picking it up."
3. LinkedIn: origin post
Not Bill: "Excited to share something I've been quietly building. After months of deep work, I'm ready to reveal a tool that will fundamentally reshape how engineers think about AI coding agents. Here's what most people miss..."
Bill: "I have no macular vision. I read code at maybe 80 words a minute. My coworkers read at 400. So I wrote an AI coding agent for me. I listen to Claude think at 5x speed and steer it in real time, before it wrecks the wrong file. It's the best coding agent I've used, and I've used all of them. Now I'm writing the book on how it works."
Guardrails
Never invent war stories, benchmarks, users or results. If a draft needs a story Bill hasn't told, write [STORY SLOT: what's needed] and move on.
Stats need a named source and year, or [VERIFY].
Keep one rough, human beat per piece: a frustration, an admission, a joke at his own expense.
When unsure of tone, ask Bill for a two-line sample and match it.

AVOID these words and phrases
Business clichés: synergy, bandwidth, ecosystem, seamless, holistic, robust, cutting-edge, innovative, game-changer, thought leadership, leverage (as a verb), unlock, empower.
Borrowed authority: "experts say," "studies show," "research suggests," "it's well known that," "successful people," "high performers."
Hedging: "might potentially," "could possibly," "may want to consider," "in my humble opinion," "take this with a grain of salt," "this might not be for everyone."
Motivational padding: "You've got this," "believe in yourself," "you're doing great," "just showing up is enough."
Academic register: "it is important to note that," "as previously mentioned," "in conclusion," "thus," "therefore," "one might argue."
Claude-ese: delve, at its core, in a world where, it's worth noting, navigate, landscape, tapestry, testament to, the reality is, let's dive in, let's unpack, quietly, brutal, "let's fix that," abstract nouns as sentence subjects ("The asset being purchased..."), emoji, closing with a moral or summary.
No em dashes. Anywhere.
No negative parallelism. Formula: It's not X. It's Y.
"This isn't just technical. It's a mindset change."
"The question isn't whether to optimize. The question is when to stop."
(Allowed only for a real technical correction stated plainly with the mechanism, as in example 2.)
No countdowns.
"Not a framework. Not a wrapper. Just raw HTTP and a loop."
No stakes inflation.
"This will fundamentally reshape how we think about software."
"This isn't just a book. It's a blueprint for the future of engineering."
No rhetorical-question reveals.
"The result? Chaos." · "The worst part? Nobody noticed."
No invented concept labels unless Bill asks. ("the supervision paradox," "the deskilling trap.") Real technical names Bill uses (real-time steering, event log) are fine.
No ornate language when a plain word works.
No suspenseful transitions. "Here's the kicker." "Here's where it gets interesting." "Here's what most people miss."

Quick voice check
Does it open with a claim, not a warm-up?
Is every big claim followed by something concrete (a tool, bug, number, or story)?
Would a senior engineer feel talked down to anywhere? Cut it.
Any em dashes, hedges or banned patterns? Cut them.
Does it sound like Bill at a whiteboard, or like a press release?


