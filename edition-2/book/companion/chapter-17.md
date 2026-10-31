# Chapter 17: The Crossover

Bill’s first crossover was supposed to take a morning: replace CodeRhapsody, the agent used to write the book, with Ensemble, the agent built from it. The first-edition account records twelve days, thirty-seven changes, and a provider caching bug. Passing the existing checks had not prepared the new agent for an ordinary workday.

The failures were not all spectacular. Some answered every request correctly while using the wrong model or credential. Others saved the entire conversation and displayed none of it after restart. The cache diagnostic was particularly industrious: it reported a problem on every request, making the real problem indistinguishable from normal operation.

These are historical incidents from the [first-edition crossover changelog](../../edition-1/book/crossover-changelog.md), begun September 27, 2026, not a defect list for the current runtime. Chapter 16 supplies the implementation predecessor for this edition. This chapter adds no implementation and reports no fresh crossover or private-memory migration. Its later October 11 coding run is a separate, second-edition observation.

## The visible half can work alone

The second assistant response appeared inside the first response’s bubble. Each vendor parser started its part counter again for each API response, so every turn could contain a `part_id` of 1. The GUI retained its artifact map across turns. New text therefore found an old DOM node and appended there.

Clearing the map at each turn would have hidden the collision, but made the client responsible for uniqueness and dependent on receiving a turn-boundary signal before the first delta. The historical repair put a single monotonic counter on the actor, shared by parsers and both delta and final emitters. A content address then remained unique throughout the session.

The settings dialog presented a subtler illusion. Every control called `sendPatch`, which had never been defined. Theme and font size still changed locally, so the operator saw the effect while the persistence call failed. Four wrong hypotheses preceded the harness log, whose first line named the cause. Temperature, maximum-token, and system-prompt controls did nothing at all; they were removed rather than left to advertise nonexistent behavior.

Even the working local font adjustment had limits. Absolute pixel sizes ignored a change to the root font size. Converting the stylesheet to `rem` made the controls scale together. The replacement multiline input also needed its height cap in relative units; a fixed pixel cap would hold fewer lines as the text grew. A zero stored font size meant “use the default,” not a meaningful size to apply.

Shutdown failed for a different reason. The CLI saved after its blocking stdin loop returned, but Ctrl-C terminated the process before that path could run. There was no race to win. A signal handler that saved and called `os.Exit` would still skip load-bearing deferred cleanup. Moving stdin reading into a goroutine let the main goroutine wait for either input completion or a signal, then leave through the existing cleanup path.

Daily use exposed these gaps quickly. Once understood, they were ordinary regression targets: another turn, a settings reload, a signal, larger text. The missing evidence was not whether the application could display a response, but whether its connected behaviors survived the next action.

## A readout must be able to disagree

The model dropdown wrote one configuration field while startup read another. Requests continued to use the startup default. Worse, `effectiveModel` preferred the settings store because that reflected what the operator had just selected. The readout showed the intended model, not the model making requests.

A readout fed from intent always agrees with the operator. A readout fed from execution can reveal that the selection failed.

Changing providers exposed the same gap in funding. The historical runtime switch already moved model, vendor, API surface, base URL, and API key together. The credential provider was introduced later and attached only at startup, behind a vendor check. Once models became changeable during a session, that startup check no longer established which credential belonged on the next request.

Switching from Claude to GPT changed the dialect and destination without bringing the plan credential along. A metered key could answer while bypassing the intended subscription. The changelog inferred that fallback from code and a 401 response; it did not observe a charge on an invoice. The risk was real without the invoice claim: an apparently successful answer could conceal a funding mismatch.

The repair added the credential provider to the state already moved by the switch. It also made the effective credential observable. `DescribeCredential` could name the credential kind but had no path to emit a token. That placed secrecy in one owner rather than requiring every logging call to remember redaction. An absent provider was described as a static key, not “none,” because the former named the funding behavior that mattered.

Startup ordering caused another wrong-model session. The resolver selected a vendor before resolving the stored model, so an Astra selection could produce a healthy-looking session running Sonnet. The diagnosis concerned stored choices outside the environment model override. The browser happened to send its selection after connecting, correcting the mistake; the CLI never did and stayed wrong throughout.

The repair resolved the model first and derived the vendor from its table entry. An explicit vendor override still took precedence, allowing a local OpenAI-compatible endpoint to serve an identifier unknown to the table. This is not a reason to eliminate overrides. It is a reason to make the default dependency run in the right direction.

## One conversation, two delivery paths

Live streaming and replay shared a WebSocket protocol, but not enough of their construction logic. Restarting exposed the difference.

The hub’s renderable-history count advanced only when it observed live events. Restoring a full log left that count at zero, so replay iterated over nothing. Initializing the count from the supplied log exposed the next failure: sends silently dropped messages when a buffer filled. A long conversation exceeded the historical 256-slot queue.

A newer progress snapshot may replace an older one. Text deltas, tool results, and replayed artifacts cannot be discarded on that assumption. The repair waited for capacity, with a timeout to avoid retaining a dead connection indefinitely. The durable requirement is preserving content or terminating the failed connection honestly, not preserving that particular queue size.

Another restart showed user messages and tool calls but no assistant text or thinking blocks. All 392 events were in the save. The replay conversion omitted `part_id`, so the GUI placed responses under the same undefined identity. The live conversion supplied it. Adding the missing field repaired the immediate symptom; sharing content identity across both paths addressed the underlying design mistake.

Reset failed in the opposite direction. The saved conversation cleared correctly, but the screen kept its old text until refresh. Recording an event appended it to the log and applied the reducer; it did not notify the observer. A live `ConversationCleared` observation made the reset visible without relying on replay to repair the screen later. Inspecting the saved record separately from the live view located the failure. Refresh alone would have concealed it.

The browser also tried to decide whether input was a new prompt or a hint by consulting its copy of turn state. After a mid-turn kill left stale in-flight state, it sent hints to a turn that no longer existed. The connection stayed open while messages disappeared.

The repair removed that decision from the client. It always sent a prompt; the actor classified the input by where it received it. At the main loop the actor was idle. At the mid-turn drain it was running a turn, so the input was a hint. The control flow supplied the fact that the browser’s copied state had failed to preserve.

The CLI initially bypassed this actor and called a bare engine loop. It could not reproduce the GUI’s failures because it was effectively exercising another agent. Rebuilding it as a text observer over the same actor made the comparison useful. A failure in both renderers points toward the shared path; a failure in only one points toward its input, timing, or rendering. Neither observation alone proves the core innocent.

## The diagnostic that cried divergence

The crossover record reports a morning consuming 46% of a weekly plan allowance: 175 requests, 6.9 million uncached input tokens, and a 22.5% hit rate. Those are recorded usage and allowance observations, not an independently reconstructed invoice. The answers still arrived normally.

The cache lens was meant to explain such costs. It split request bodies into named sections in the provider’s actual cache order and compared them with prior requests. Stable local bytes, a changed prefix, and a provider cache miss despite stable bytes are different findings. The last leaves routing, eligibility, expiry, or eviction to investigate; local comparison cannot identify which occurred.

The lens first had to learn what growth looked like. Appending to a JSON array moves its closing bracket, so a naive string comparison calls every append a rewrite. Comparing sections after trimming the prior container’s trailing framing distinguished an unchanged prefix from ordinary growth.

A second mistake defeated that distinction on Responses. The canonicalizer recognized Chat Completions’ `messages` and Gemini’s `contents`, but not Responses’ `input`. It treated the conversation as a frozen section. Every normal appended turn triggered the same divergence warning. A genuine rewrite that reduced the surviving prefix to 34% looked exactly like the harmless appends around it.

Reading the conversation key from the body repaired that classification. Reporting growth separately from rewriting, with the offset where matching stopped, made the warning useful. The problem was not absent instrumentation. The earlier cache chapter had already supplied a lens and meter. The instrument was observing the wrong thing with great consistency.

Breakpoint placement needed the same scrutiny. One marker at the previous exchange repeated a boundary already offered by the earlier rolling marker. In the tested explicit-cache layout, its slot was more useful at the end of the tool array. Tools came before the system prompt in cache order, so that boundary could preserve an earlier eligible prefix when the prompt changed.

The resulting possible boundaries were tools, system prompt, retained checkpoint, and the last eligible retained content before ephemera. If a handoff had been removed, surviving memory could represent the checkpoint. That memory remained data, not a mutable system prompt. These markers offered reuse opportunities on a supporting route; only provider usage could show that a cache write or read was granted.

## Stable bytes do not establish a cache policy

After repairing the lens, the historical investigation varied request size. A locally stable 99% prefix produced no cache writes at 8,000, 10,000, or 18,000 tokens. The first request above roughly 22,000 tokens was cached, and the next read back 22,144 tokens. The investigator compared that with the then-documented 1,024-token minimum and described an apparent floor.

Seven requests did not establish a universal threshold. Request size changed alongside session history, and the lens could not see which provider-side entries existed. Padding future prompts to reach 22,000 tokens would have turned an unexplained observation into a spending policy. Trimming can trade fewer input tokens against less reuse, but deciding which is cheaper requires actual usage and applicable prices.

Nor did size explain the whole record. The expensive morning’s requests contained 55,000–73,000 tokens. A first-party client using the same OAuth token on a different endpoint recorded 61–85% hit rates, including 12,160 cached tokens on the first request of a fresh session. The record ended with a provider bug report and a route still too expensive for that workload. An expected fix within weeks was a hope, not a delivery commitment.

The live plan route also exposed assumptions shared by a parser and its fake: rejected explicit-cache controls, absent Content-Type, empty terminal output despite completed item events, and missing tool-call provenance. Each repair exposed the next mismatch. Chapter 16’s separate second-edition run reproduced its own cache-options rejection and empty-terminal-output failure. Its live header was not retained; captured-body and controlled-header evidence supported the missing-header parser condition instead. Those records cannot be combined into a wire capture that never existed.

A fake and its client can agree because they share the same mistake. Captured provider behavior can correct the fake, after which a focused regression can catch the defect without another paid request. The meter supplies usage categories, the lens explains local changes, and provider counters establish granted reuse. None substitutes for the other two.

## Reaching the screen, and reaching the supervisor

The historical agent had a GUI and a documented MCP server but no active path to inspect its own browser. A successful stdio client check had not established that connection. The repair supplied two directions: a loopback TCP relay let external MCP clients drive the GUI, while a local `view_gui` builtin requested a snapshot over the existing WebSocket. A recorded live check then had the agent correctly name the browser’s buttons and tabs.

Before that worked, the tool had been registered and unit-tested without ever reaching a model request. The registry normalized `view_gui` to `viewgui`; the skill filter compared that key with the declared name `view_gui` and rejected it. Builtins survived accidentally because their constructor stored raw names while registration stored normalized ones.

Checking the declared name repaired filtering, then broke the MCP connection path, which recorded enabled tools using the normalized key. A cross-chapter check caught the regression, and the second repair aligned that path too. A normalized lookup key is not a public identity. Fixing one end of an inconsistent convention can reveal the other end’s dependence on it.

The supervisor needed a different kind of visibility. Bill listens to narration through a screen reader at high speed; it lets him catch a mistaken approach before it becomes a sequence of edits. The changelog records a silent session followed by a hard reset that also discarded his uncommitted work. Missing narration made review harder. Destructive recovery added a separate failure. Interrupting the turn and inspecting the diff preserves more options, though the Agent and other jobs remain alive and still need supervision.

That implementation added a model-specific visible-text gate. If a round contained no visible text, the actor rejected every tool call before execution and returned a result for every call, preserving call-result pairing. Rejecting after execution would have been too late. The errors explained the rule and a concrete way forward, and asked the model to report surprises that might expose a mismatch.

This was a historical enforcement mechanism, not a new requirement here or a demand for private reasoning. Useful progress explains the intended change and its reason. As Chapter 16 establishes, public reasoning summaries and ordinary assistant explanation are separate streams; one cannot be assumed to supply the other.

Other observed mistakes produced explicit engineering standards. Deleting a comment that recall was hooked in two places removed the only explanation of an invariant while leaving both dependent call sites intact. Unformatted edits and stray backup files caused different harms. The resulting standards included checking output rather than trusting exit status, anchored edits, preserving explanations until their conditions are gone, and never weakening tests. Each rule carried a reason, so an exception had to answer an argument rather than merely ignore a slogan.

## A file that actually changed

The first-edition account reports that Ensemble edited chapters describing itself, including a sub-agent voice pass. That is evidence of use, not a guarantee that each regenerated edition improves, or proof that the second-edition agent had completed the same crossover.

The second edition began with a smaller test: could Ensemble run Bill’s selected model, repair a real program through tools, and leave a result another process could verify?

For the October 11, 2026 run, the selection was `gpt-6-astra` at high effort. Its existing table entry still chose Chat Completions. The documentation cited in the source required Responses for Astra function calls, so moving that row to the existing renderer was the entire runtime change. Effort settings, credentials, and stream handling already existed. The provider seam made the change small, but a model name alone had not selected a usable route.

The task was a CSV spending report that split lines on commas and summed amounts as binary floats. A merchant named `Books, Inc.` broke the field split; a quoted newline broke the line split; decimal currency and refunds exposed the arithmetic. Six tests fixed expectations before the agent began, including ordinary CSV quoting, escaped quotes, blank rows, exact merchant totals, and the existing sorted, tab-separated, two-decimal CLI output.

The CLI session read the source and tests, then ran them: three failures and three errors. Before editing, Astra explained the CSV and precision defects. It replaced manual parsing with the standard library’s `csv.reader`, opened the file with `newline=""`, used `Decimal` arithmetic, and skipped blank rows while retaining the output format. The same six tests then passed. An independent rerun of preserved before-and-after directories reproduced both outcomes, and the test file was byte-identical. The [source diff](../../solutions/evidence/ch20/task.patch) made the actual repair inspectable rather than leaving it inside the model’s report.

Five requests completed in 27.46 seconds. Requests and terminal responses consistently named Astra/high. Subscription funding was selected explicitly, with no API key configured. The CLI exited normally after EOF and saved its conversation. The endpoint was established through configuration and transport source, not an independent capture of the HTTPS URL or authorization headers.

The run produced 161 ordinary text deltas and no public reasoning-summary deltas despite requesting detailed summaries. Every response reported zero cache grants and zero reasoning tokens. Those fields do not measure private mental activity, and the successful repair does not establish cache savings. The ordinary explanation arrived before the edit, which supplied the supervisory visibility this task needed.

This was one bounded coding task, without a steering or GUI demonstration. It established a real edit, unchanged tests, independently reproducible results, and normal shutdown. It did not establish general project reliability or authorship of the next edition. The six cases narrowed what had to be trusted; the preserved diff let someone other than the agent judge the change.
