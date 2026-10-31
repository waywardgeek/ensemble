# Chapter 15: The Invisible Invoice

Bill reported $2,475 in model spend on the first edition through its recall chapter. He attributed a few hundred dollars to a caching bug. The answers were correct and arrived on time; the repeated input simply stopped receiving its discount. Nothing in the agent's behavior invited investigation.

Chapter 14 reduces the working set while trying to preserve stable prefixes. That does not establish that a provider caches them. Each request carries tools, instructions, and retained conversation; much of it repeats, but reusable bytes, correctly placed markers, and an actual cache grant are different facts. The [source chapter](../edition-3/chapter-15.md) connects them through three components: a lens that locates prefix changes, markers that identify reusable boundaries, and a meter that makes the resulting cost visible while someone can still act on it.

## Observe the Request That Actually Leaves

The cache lens belongs to Engine and observes actual main sends, not every call to `Render`. Rendering can happen without a request being sent. Later compressor and judge conversations need their own accounting, but including them between main requests would make unrelated conversations look like prefix damage.

The lens rotates current and prior request files under the owning Agent's directory. The supplied clients use `cr/cache-current.txt` and `cr/cache-prior.txt`; startup resets these diagnostic files even when conversation history is restored. Named sections appear in the provider's cache concatenation order, not whichever order the JSON struct happens to emit. For Anthropic that means tools, system, then messages. A prefix is contiguous: comparing those sections in the wrong order gives the wrong boundary.

Each section is identical, appended, or changed. An array that gains elements at its tail is appended even though its closing bracket moves. Treating that as an edit would sound the alarm on every ordinary turn. Unexpected changes to stable sections deserve attention; deliberate skill, checkpoint, or model changes need interpretation rather than automatic diagnosis as defects. The files keep the evidence inspectable with ordinary `diff`, while structured verdicts make it usable by automation.

The conversation section is recognized by shape: an array of objects with `role` fields. Anthropic and OpenAI call it `messages`; Gemini calls it `contents`. An early lens exempted only `messages` from its stable-section alarm. Its Gemini report could therefore say `contents:appended` and simultaneously announce a broken prefix with a precise byte offset. The section comparison was right; the alarm interpreting it was wrong.

Markers must also disappear from the comparison. The history marker advances each turn, so yesterday's marked block is unmarked today. Those bytes are provider instructions, not changed conversation. Leaving them in the diff truncates the predicted reusable prefix at the old marker position.

Stripping them through a JSON unmarshal/remarshal cycle is tempting but changes key ordering and byte counts. Equality might remain consistent between the two transformed requests, yet the measured bytes would no longer describe the request being diagnosed. The source instead strips the marker's shape in place while preserving every other byte.

The first-edition implementation got that shape wrong. It matched a constant containing exactly the compact marker emitted by the renderer, justified by the comment “we control the emitter.” But the lens saw the canonicalizer's `MarshalIndent` output, where spaces changed the marker's spelling. The strip matched nothing and failed without an error. Every compact, hand-built test passed. A replacement test using the actual indented shape exposed the defect: the retained review's two-turn comparison changed from an edited messages section to an appended one. The instrument needed a sample of the artifact, not another copy of its author's assumption.

## A Late Reply Keeps Its Own Prediction

Engine's serialized turn decisions do not imply serialized HTTP completion. An interrupted worker can still finish after a new turn has begun. The lens therefore returns its predicted reusable fraction to the individual request, whose response callback carries that prediction back alongside the provider's usage.

Independent review demonstrated why a shared “latest prediction” is insufficient. It held an old cold response, interrupted the turn, and completed a new request. The predictions were 0.0000 for the old request and 0.8554 for the new one. Releasing the old response made it report 0.8554 too. Locking access to that shared value would remove a data race while preserving the attribution bug.

The corrected probe associated each response with its own prediction and counted both actual usages. A short lock protects comparisons and file rotation; none spans the HTTP wait. Interrupting the conversation does not erase a charge already incurred.

This prediction is a byte fraction, whereas the provider's grant is a token fraction. Comparing the ratios can reveal a large discrepancy; adding or subtracting the raw quantities cannot. Tokenization is not linear in bytes, so close ratios are evidence consistent with reuse, not an identity or a billing audit. The original Anthropic report paired a 96.8 percent byte prediction with a 99.4 percent token grant.

## Spend the Marker Budget on Different Boundaries

The source's explicit-cache layout spends a four-marker budget on distinct positions: the end of tools, the end of SYSTEM, the surviving compaction boundary, and the stable end of retained history. Each protects a different portion of the request. Missing content needs no placeholder just to spend a slot.

A marker on the last tool declaration banks the earliest prefix. The system marker normally includes those tools anyway, but cannot preserve them independently when the prompt changes. A separate tool marker lets unchanged schemas remain cached through a SYSTEM edit.

The system marker banks the instructions actually sent. Skill manuals belong in SYSTEM under Chapters 9 and 14, even when their presence changes the prefix. Moving a manual into a tool result to produce a better cache statistic would break instruction delivery. Likewise, editable memory is not promoted into frozen system instructions merely because stable bytes are cheaper.

History needs a rolling boundary. Marking the first user message adds only that message beyond the tools and instructions already covered by the system marker. Everything learned afterward remains cold. Instead, the stable-end marker sits on the last eligible retained block and advances as history grows. Moving the marker does not invalidate the content it previously banked.

For this layout, a separate start-of-turn marker duplicates the rolling boundary from the previous turn. The providers can look backward through a bounded set of positions for a prior write; the client does not need to reproduce their exact lookback algorithm. Spending another slot on an anchor can obscure an actual prefix rewrite rather than cure it.

The stable endpoint is chosen after final tool-result adjacency has been established in the disposable wire view, without altering canonical history. It must stop before ephemera. Those delivered-once bytes are appended last and then cleared, sometimes sharing the final wire message with retained content. A marker after them banks a boundary guaranteed to disappear. An earlier prefix may still hit, which can make the placement mistake less obvious; the disappearing tail still cannot match. Capturing the retained boundary before appending ephemera avoids confusing message boundaries with content lifetime.

### Compaction Changes What Can Be Banked

A `micro_handoff` removes tool traffic, opaque reasoning, and private replay signatures from the affected history. The old rolling marker may have banked bytes that no longer exist. The cache must rebuild from the retained projection, not from a reconstruction invented to preserve a discount.

The newest surviving handoff provides a useful compaction boundary. Its cuts were already recorded and applied by History; rendering does not recompute them. The inherited renderer projects the surviving note before transient ephemera rather than preserving a handoff dialogue message. Marker placement must use a legal boundary in that actual projection. If none is available, the slot is omitted. Neither resurrecting old dialogue nor relocating the note into SYSTEM is an acceptable substitute.

A uniform handoff cut makes this boundary simpler than a policy that retains recent demonstration pairs. If “the three most recent pairs” is recomputed on every render, a newer handoff can displace pairs kept before an older one. The first surviving pair then becomes a moving boundary of its own. Ensemble's uniform strip avoids that extra dependency.

Chapter 22 adds a fallback to the last surviving memory entry when a memory fold removes the handoff. Rewriting memory near the front invalidates the suffix from that point, while earlier tools and instructions may remain reusable. That is a real cost of replacing memory, not a reason to keep obsolete content.

### Similar Intent, Different Wire Contracts

Anthropic accepts `cache_control` on content blocks and tool definitions, allowing all four positions. SYSTEM must be an array of content blocks rather than a string.

The source's supported GPT-5.6 Sol Chat Completions route uses `prompt_cache_breakpoint` on message content blocks, with request-level `prompt_cache_options` selecting explicit mode. The tool array is not a message content block, so only three of the four positions are expressible. The system marker still covers tools during ordinary operation, but cannot preserve them independently through a prompt change. Support is selected from the model table; older rows retain implicit caching. These are model- and route-specific contracts, not interchangeable fields across OpenAI APIs.

Explicit mode has a particularly expensive empty case: it disables implicit caching. Selecting it without successfully placing any marker caches nothing. Anthropic's missing marker forfeits an explicitly requested cache; OpenAI's empty explicit request also gives up the implicit cache it might otherwise have received. The opt-in follows evidence that at least one marker landed, not an intention to place one.

Gemini remains on `generateContent` with implicit caching. There are no markers to budget or separately managed cache objects in this design. Its eligibility threshold still matters. The source's October 11, 2026 documentation lists minima of 1,024 tokens for Sonnet-5 and GPT-5.6 Sol, and 4,096 for Gemini 3.8 Flash. Those dated values require the selected model and route; request-file bytes are not a substitute for tokens. A small prompt with no cache reads has not demonstrated a cache defect.

## The Provider Must Actually Grant the Read

Historical live runs show why plausible requests and correct answers are insufficient evidence. They also need their experimental conditions attached.

The first-edition Anthropic run reported cache reads of 0, 3,627, and 3,648 tokens. The 21-token increase matched the previous exchange, showing that the advancing marker banked more history in that run. It does not establish a fixed increment for other responses or providers, and the measurement predates the final marker arrangement.

An OpenAI experiment initially removed a marker from a primed position and got zero reads. It also changed the message from an array of content parts to a plain string. The apparent explanation, that removing the marker had broken the prefix, had not been isolated. A second probe changed only the encoding and received 7,813 cached tokens out of 7,816. That tested prefix survived the encoding change; the missing lookup breakpoint explained the earlier zero. A probe changing two variables had offered a much more confident diagnosis than it earned.

Gemini's observations depended on conversation shape. The original small-prompt script produced three zeroes, initially diagnosed as unimplemented caching. An eligible run corrected that interpretation. Later probes used a fixed large user message, a large stable system reference with small growing history, and history growing by roughly 8,400 tokens per turn. The first two warmed on turn two; the growing-history run began with two zeroes. An early review tried to challenge the history result using the fixed-message probe. Both measurements were real, but they answered different questions.

A later browser run with a large stable reference also warmed Gemini on the second request. Its changing read counts establish neither a universal 4,096-token block rule nor a rule that partial blocks remain uncached forever. Stable content belongs early, but “stable for this demonstration” does not authorize placing arbitrary retrieved text in instructions. Warm-up and the amount granted must be observed on the shape the application actually sends.

Fake providers remain valuable for request structure, parser behavior, and accounting. They grant whatever counts the fixture supplies, including below a real provider's minimum. A larger byte fixture does not fix that evidence boundary. The source's numeric marker fixture also has no handoff or ephemeral tail, so it cannot establish those placement rules. Focused observations cover browser rendering, interrupted-worker attribution, and compaction boundaries separately.

Real use can reveal a contract the fake never modeled. The historical GPT-5.6 Sol Chat Completions run refused function-tool requests with reasoning effort. Omitting `reasoning_effort` did not repair it: omission selected the provider's reasoning default and produced the same refusal. That route needed the explicit value `"none"`. A faithful regression can preserve the discovered restriction, but a fake cannot discover an absent contract by itself.

## Count Once, Price at Display Time

Usage records four disjoint categories: ordinary input, cache writes, cache reads, and output. Prices live on `ModelFeatures` as full dollar rates per million tokens. A cache-write rate is not a surcharge to add on top of ordinary input. Each token belongs to its applicable bucket and is charged once.

Vendor responses do not arrive in that shared form. Anthropic reports disjoint ordinary, write, and read input. OpenAI's prompt total already includes cached reads and writes, so those buckets must be subtracted to derive ordinary input. Gemini's prompt total includes cache reads; its charged output includes reported thinking. Streams retain final usage rather than summing cumulative updates.

Within a valid usage object, an omitted optional cache bucket means zero. Missing usage, missing required input or output counts, and invalid supplied counts remain unknown. Treating any of those as free usage makes an accounting failure look like a bargain. An unknown price likewise displays a dash, while available token counts remain visible.

Each request's counts retain the model and applicable price tier that incurred them. Engine derives cost from those identities and rates when rendering the meter, rather than storing dollars in the event log. A price correction can then change the derived cost without rewriting measured history; switching models cannot reprice earlier work as if the new model had performed it.

The first implementation's meter looked up the model named in settings. When the environment selected the model, that setting was empty. The lookup failed and a paid Sonnet run reported `priced:false, cost:0`. The instrument intended to expose waste was concealing it. The correction takes the model actually in force from Engine configuration at the composition root. The retained review attributes this finding to the grader, not to the live run: ordinary application ownership was enough to explain and reproduce it.

## The Meter Has Two Time Scales

The GUI header displays session cost, last-response input tokens, last-response cache-read tokens, and session cache hit rate. “Input” here means all input to the latest main response: ordinary input plus cache writes plus cache reads, each once. The public `usage` frame exposes those disjoint buckets separately; the header combines them. A tooltip holds session totals so the current prompt size is not buried under a number that can only grow.

Engine owns process-lifetime accounting and separately retains the latest main-response counts. Neither is restored from `save.json`. Conversation history can return after restart while the meter starts at zero, because no response has yet been charged to the new process. Restored historical usage answers a different question from the cost since the developer sat down.

Chapters 22 and 23 add compressor and judge calls to session totals and cost, priced by the model that performed them. They do not replace the latest main-response display and do not enter the main-request lens comparison. The distinction keeps a small auxiliary response from masquerading as the size of the actor's current prompt while still charging it to the session.

Earlier October 11 browser sessions on Haiku 5.5, GPT-5.6 Sol, and Gemini 3.8 Flash each returned four correct answers, with provider usage independently compared against visible meters. Their session hit rates included the cold first request, unlike a last-response rate. Haiku 5.5 remains a historical identity, not a current supported choice. Those captures had no restart or handoff and preceded the lens-overlap correction, so they could not establish those behaviors.

The unchanged original grader, meanwhile, returned 0/100 because its primary skill fixture omitted the required type and the GUI never started. That was a setup failure, not evidence of seven broken cache features.

The source's accepted October 22 Sonnet 4.6 run returned four correct answers. Provider usage yielded $0.0755091; the header displayed $0.0755 and the tooltip $0.075509. Restart replayed the conversation with a zero process meter.
