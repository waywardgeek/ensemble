# Chapter 19: The crossover

The first edition opens its crossover chapter with two agents on the screen:
CodeRhapsody, which helped write the book, and Ensemble, which the book describes.
Bill has to choose which one gets the next piece of work. An agent can pass its
exercises and still be the wrong tool to leave open all afternoon.

That account earns its place here because the failures were ordinary enough to
be convincing. A model picker displayed the new selection while requests still
went to the old model. A GUI-debug skill promised access to the browser, but the
shipping application never connected it. The replacement stdio fixture answered
perfectly. Nobody at that endpoint had seen a browser.

The second edition teaches those boundaries earlier. This chapter asks them to
hold together while the reader does something useful. Choose a real small job,
give it to Ensemble, inspect its work, close the application and return to the
conversation. The answer to “would you use it again?” belongs after that sequence.

*Contract draft for independent review. The accepted scope is a bounded real-work
capstone. The preceding Chapter 18 source and all required checks must be accepted
before this exercise begins. No crossover result is claimed yet; the
[validation record](chapter-19-validation.md) tracks release and evidence.*

## TL;DR

Read the complete [coding skill](skills/ensemble-coding/SKILL.md) and
[architecture](architecture.md). This chapter adds no mandatory runtime service
or wire format. Its deliverable is a useful reviewed patch and an honest record
of obtaining it through Ensemble.

1. Freeze a real task before seeing the model's answer: accepted source, allowed
   files, observable expected behavior, meaningful checks and a permitted resumed
   next step. Use an isolated second-edition checkout. Preserve first-edition
   artifacts and unrelated human work.
2. Deliver the complete coding skill, Chapter 1 architecture teaching and relevant
   new contracts before coding. Use §19.2's existing dependency-manual setup,
   verify exact rendered coverage and retain the permitted-source manifest.
   A successful truncated file read does not establish instruction delivery.
3. Start the principal session with the optional GUI launcher, `--gui-debug`,
   `--gui-view main` and `--terminal`. Its CLI and browser share one Agent with
   fixed aliases, scope and session identity from construction. A later browser
   stage cannot widen a CLI-only session.
4. Have Ensemble inspect the task, edit permitted files and run the selected
   checks through its ordinary tools. Record corrections and effects, including
   a model's explanation when the file or test contradicts it.
5. Checkpoint at a settled boundary, close normally, resume the same compatible
   store and perform the frozen next step. Use actual GUI-MCP observation/action
   and public two-Agent embedding. Keep those interface claims tied to the
   principal route; other routes use human CLI parity exercises.
6. Keep the complete allocation within 40 model starts and the separate finite
   preflight budget. Every continuation, admitted failed attempt and unexpected helper
   counts. No restart, route change or nicer prompt replenishes it.
7. Freeze the initial answer and experience before independent patch review.
   Correct defects at their responsible earlier teaching/source, preserve the
   initial record and revalidate affected paths. End with a scoped adoption
   decision, including an unsuccessful one when the evidence requires it.

Use the accepted predecessor's documented build commands to produce the core CLI,
optional GUI command and public consumer, with source/binary associations. Run
relevant module tests, vet and race checks, retaining inherited acceptance coverage.
The independent Chapter 19 integration/evidence invocation must be published
before exercise release; §19.7 defines its required coverage. There is no working
historical `CH=20` grader to substitute. A full earlier-chapter score is a useful
prerequisite, not the result of this exercise.

## 19.1 Give the agent a job worth finishing

“Improve the project” is an invitation to discover what the model feels like
editing. It supplies no end condition and gives the reviewer little ground for
disagreement. A crossover needs a job whose result someone can reject.

Before the first model start, the coordinator and reviewer select one real issue
in the accepted second-edition source. Prefer a small reproducible defect with
a concrete user consequence. If no suitable defect exists, choose a useful
bounded improvement and say so. Do not plant a fault and narrate its discovery
as a production failure. The following is a task-brief template, not a report of
an issue already found:

```text
Task ID and date:
Accepted source revision and tree:
Observed problem or requested improvement:
User-visible expected behavior:
Allowed edit paths:
Forbidden changes:
Baseline reproduction and actual result:
Exact candidate check commands and expected outcomes:
Independent distinguishing check:
Permitted resumed next step:
Principal route, selected model and funding:
Frozen teaching manifest and catalog identities:
```

Fill every field and freeze the brief before launch. Limit the candidate to at
most six explicitly named tracked source/test/documentation files in at most two
existing modules. If the real issue needs more, select a smaller issue before
starting; discovering that the chosen job exceeds its scope is a legitimate stop.
The six-file limit excludes generated build/evidence artifacts, which have their
own designated directories. It never authorizes changes to unrelated files.

For a defect, preserve the original failure using the same public behavior the
repair is supposed to change. For an improvement, preserve the original behavior
and a distinguishing check of the added requirement. The model can write its own
tests, but independent verification must not rely only on those tests or its
summary of their output. A compiler exit and an application's observed behavior
answer different questions.

Create a clean isolated checkout from the accepted source and use separate data
directories for sessions and jobs. Record initial file hashes and the Git tree.
Use this checkout as the working directory for Ensemble's ordinary file/process
tools. Those tools are not a filesystem sandbox; keep the permitted teaching
copy and scratch data deliberately accessible and forbid outside edits in the
task. Do not copy credential files into the checkout. Do not use reset or clean
to erase work whose ownership has not been established.

The running executable remains bound to the source that built it. An agent editing
its own source does not thereby replace its current process. Preserve the candidate
patch separately, then review and build any revised executable as a new artifact.
The first live run cannot be retrospectively attributed to that later binary.

## 19.2 Put the instructions where the model can actually read them

The mandatory coding skill in the scope review was 19,879 bytes. The ordinary
Job report retained 16,384 bytes by default. Reading the whole file into a job
artifact while sending a truncated report would leave the missing instructions
on disk, where the model cannot use them.

Use Chapter 9's existing dependency manuals for this exercise. This is explicit
instruction setup, not a new model request or another implementation of Skills.
Before constructing the principal Agent, prepare a frozen caller-supplied catalog
with a crossover primary based on the accepted ordinary coding primary. Preserve
the selected local tools and GUI-debug integration, and add dependency manuals
containing the permitted teaching below. The complete rendered manuals must reach
the first actual request before any code-changing call is accepted.

The permitted teaching set is the complete mandatory skill, architecture.md,
Chapter 1, this chapter, and the complete current second-edition chapter contracts
that govern the selected task. Freeze those exact paths and revisions in the
brief. Do not follow historical/research links inside them. Grader implementations,
reviewer findings, first-edition answers and unreleased future source remain outside
the supplied context. Give the model only the accepted predecessor implementation
and the selected task's public requirements.

Make an exact local copy of each permitted file under a read-only teaching root,
preserving its repository-relative name. Record source revision, path, byte count
and SHA-256. These are ordinary teaching files, not another checkout's credentials
or executable hooks. The accepted main-source export does not itself contain the
book's skill; the task manifest must identify this additional permitted root.

To turn a file into dependency bodies, split its UTF-8 bytes at complete-character
boundaries into contiguous nonoverlapping chunks of at most 30,000 bytes, without
dropping line endings or adding text inside a chunk. Escape every literal dollar
as `$$` in the catalog source so Chapter 9's renderer produces the original bytes.
The chunk may include Markdown or the original skill's frontmatter as literal
body text. Give each generated dependency a fixed valid ID and description, no
tool grants and no further dependencies. Its Chapter 9 source, including the
small generated header and escaped body, must fit the existing 65,536-byte limit.
Use distinct generated IDs that do not replace existing definitions. Validate
the complete catalog before exposing an Agent: at most 256 definitions and
8,388,608 source bytes, at most 8,388,608 newly rendered body bytes, and the
existing 67,108,864-byte complete transition-record bound. The primary and its
ordinary dependencies count too.

The primary's depends list selects every chunk in documented file/chunk order.
Use sequence IDs `crossover-0001`, `crossover-0002` and onward so Chapter 9's
deterministic ordering agrees with that manifest. The primary instructs the
model to read these manuals before code,
stay within the task and report surprises. Its source and rendering must satisfy
the same existing variable rules; do not silently replace unknown variables.
No change to the shipped skill file, parser or catalog limits is required.

Preflight verifies that each rendered body equals its exact original byte slice
and concatenating the slices reproduces the entire file. The first captured real
request must include every rendered manual through the taught Skills projection,
with no redaction, clipping, missing prefix or substituted summary. Compare the
decoded request text, accounting for its normal JSON escaping. Record that
coverage separately from the ordinary tool transcript. Full delivery proves what
was supplied; it does not prove that the model understood or obeyed it.

Keep automatic maintenance disabled during this bounded work. If required text
does not fit an existing local/provider bound, stop and report instruction setup
incomplete. Do not shrink the skill or secretly create a second preparation
conversation. Any subsequent model-driven read, recovery request or continuation
uses the allocated starts. After a task change or context loss, the skill's own
reload rule still applies; insufficient remaining allowance means the task stops.

## 19.3 Start with one Agent and two real clients

The principal route and model are chosen before the task. Prepare its existing
authorized connection, model discovery and funding selection under Chapter 17;
this chapter adds no login flow. Use a fresh compatible summary/request profile
and the base profile required by the fixed skills and GUI integration. Record
the complete creation identity. Those wrappers grant no missing base capability.

Start the accepted optional GUI launcher with `--gui-debug --gui-view main
--terminal`, its inherited explicit session-directory selector, frozen catalog
and selected route configuration. Use the actual published options of the
accepted source; record the sanitized full invocation. The core CLI cannot serve
this browser. Chapter 12's phased startup creates the view, prepares its endpoint
and freezes aliases before constructing or mounting the Agent. Open the printed
loopback URL and verify ready state before the first prompt.

The terminal and mounted Page now belong to the same Agent. Keep the logical view
name main and all session bindings unchanged across the principal workflow. On
restart, explicitly attach a new physical endpoint and accept its new mount under
the same compatible logical identity. Never turn a CLI-only store into a GUI-enabled
one by adding aliases later. Keep ordinary no-GUI startup available to the parity
routes and public headless consumer.

Read the effective state before trusting a friendly label. Record current selected
route/model/funding and then the actual attempt's captured configuration, returned
model identity and producing usage. The historical model picker agreed with the
operator because it read what the operator wanted; the engine had read something
else. A useful instrument has to be able to disagree.

Keep the principal route fixed for this exercise. Foreign bound replay material
cannot become portable because a dropdown offers another model. The other funding
and API routes get fresh compatible stores. Current plan access and cache behavior
come from current evidence, never the historical crossover's floor or bug report.
No missing plan connection may silently become a billed API-key request.

## 19.4 Watch the work, then come back to it

Submit the frozen task through ordinary human chat. A useful first instruction is
specific about the boundary: read the supplied manuals, reproduce the stated
problem, explain the intended small change, edit only the named files and run the
listed checks. The exact filled prompt belongs in the preflight brief. Do not
paste an expected implementation or perform the edit outside Ensemble and ask
the model to endorse it afterward.

Read the answer and actual tool result before sending a follow-up. A file tool's
report is not the file; a test explanation is not its exit/output receipt. Inspect
the candidate files and the complete retained job artifacts. If a report truncates
a decisive result, use the existing bounded retrieval/report controls and count
any resulting model continuation. Record a corrected misunderstanding instead
of rewriting the first exchange to make the agent seem more capable.

The reader can send a hint for a later request or interrupt work still awaiting
acceptance. Neither changes a body already sent. An admitted tool may finish after
the turn ends, and interruption cannot erase its side effect. Provider-exposed
summaries help explain what is arriving when the selected route supplies them;
they are not full internal reasoning or approval to act. If the model offers no
summary, preserve that outcome without inventing a narration-enforcement feature.

At a settled boundary, record the diff and check results, then use the existing
checkpoint and normal close path. With the combined launcher, detach terminal
input using EOF before orderly server shutdown, retaining the actual shutdown
receipt. Do not terminate the process to simulate a successful checkpoint.

Restart with the same source and compatible configuration. Ask about the original
problem, the change and the test result without retelling the answer. Require
the saved SessionID, a fresh live mount and unchanged hashes for completed effects.
Then request the one permitted next step frozen in the brief. Choose that next
step before launch, such as running a named regression command or inspecting a
named boundary; it cannot become an invitation to edit a second subsystem.

A resumed answer can remember the wrong explanation faithfully. Compare its account
with the saved result. Historical jobs must remain historical: no live owner may
be inferred from a printed handle, no old edit may run twice and no prior speech
may autoplay on reconnect. A transcript that looks intact is only the first of
those checks.

## 19.5 Let the agent look at the page

The first edition's missing GUI connection is a useful embarrassment. The model
had a manual promising access, the test had a server answering requests, and the
real browser was outside the conversation. New Chapter 12 already fixes the
construction boundary. This exercise follows the actual path again during work.

Use the same principal Agent and mounted Page. Keep the human draft empty while
asking the Agent to inspect the view with gui_snapshot. Require an actual scoped
MCP call and result, then compare the selected tab and one retained artifact with
the real page. A tool declaration or the model naming a plausible button is
insufficient. The screenshot and accessible description belong beside the tool
result, with the same mount and observation time.

For one concrete action, freeze applied theme light during setup. Ask the Agent
to use the supported settings controls to select dark, then inspect again.
Resolve current control IDs from the actual snapshot; do not teach arbitrary
selectors or JavaScript execution. If opening settings is necessary, use its
supported click action. gui_input may edit an ordinary registered select; only
the existing acknowledged setting change establishes applied state. Check the
theme and preference revision independently in the Page and safe state.

If a call times out after the setting reached the server, retain the uncertainty
and inspect current state before considering another action. Do not resend it
blindly or call it rolled back. A human-owned draft, stale mount or disabled
control keeps its taught refusal; none may be bypassed to complete the chapter.

Reconnect the Page and check the accepted artifacts, applied theme and generation
replacement. The same answer must not appear as a new answer or start speaking
again. Use the existing recorder when exact speech admissions need a local
receipt. A queue, journal or callback does not prove a human heard audio. Native
speech coverage from an accepted predecessor retains its original source and
scope; this capstone must not relabel it as a new listening session.

## 19.6 Spend the allowance once

Before any live launch, freeze the completed task brief, teaching/catalog manifests,
source/build identities, actual prompts, expected effects and exact check commands.
The coordinator and reviewer must confirm that every row below is feasible through
the published interfaces. Unfilled task-specific fields are a launch blocker, not
permission for a checker to invent them privately.

Use at most 40 model starts over five routes: Messages API-key, Chat Completions
API-key, generateContent API-key, Responses API-key and Responses ChatGPT-plan.
The principal route receives 24; each other route receives four. A start consumes
its slot at durable model admission, including failed or canceled attempts. Count
every automatic continuation and any unexpected helper. No paid listener/judge
gets an unlisted allowance. Unused slots expire, and budgets do not transfer between
rows or routes.

| Principal allocation | Action and required evidence |
|---|---|
| Initial human terminal, 8 | Complete instruction delivery, task inspection, bounded edit and meaningful test commands. Preserve actual calls/results, diff, failures and corrections. |
| Resumed terminal, 4 | Settled checkpoint/close/reopen, recall grounded in saved facts and the frozen next step. Verify no repeated old effect. |
| Same Agent's GUI, 8 | Scoped view/artifact inspection, light-to-dark applied setting, second inspection and reconnect. Actual browser/GUI-MCP receipts, no duplicated accepted content or historical autoplay. |
| Separate public embedding, 4 | Two fresh Agents with at most two starts each, distinct scratch inputs and owned captures. Close A after its task, then exercise the already-created B; confirm unaffected completion/accounting. |

The public tasks use ordinary read_file plus a short answer: give A and B distinct
files whose marker bytes are frozen independently before launch. Require actual
read calls/results and answers from those results. A public executable imports
only the public library, obtains real completions and retains observer/usage
records. A correct answer alone cannot prove which file or Agent supplied it.
These are principal-route public checks; the text must not call them five-route
embedding evidence.

For each nonprincipal route, use actual human CLI with a fresh compatible store
and a fixed ordinary tool set. Supply read_file, edit_file and run_command through
the existing creation/skill rules; retain the same set after restart. Prepare
note.txt as exactly `alpha\nbeta\n` and a separate sentinel file with a recorded
hash. The user task is to read note.txt, replace its sole beta line with checked,
then run a bounded command that prints the file and confirms its bytes. Expected
file content is exactly `alpha\nchecked\n`; the sentinel must remain unchanged.

The exact parity check is:

```sh
python3 -c 'from pathlib import Path; b = Path("note.txt").read_bytes(); print(repr(b)); assert b == b"alpha\nchecked\n"'
```

Require the actual process's printed bytes and zero exit status, then independently
inspect the file. Python 3 is a preflight prerequisite for this small fixture;
installing it is outside the model's task and allowance.

Reserve three starts for that initial task and one for its resumed follow-up.
After a settled checkpoint and restart, ask which line changed and why, explicitly
requesting no tool use. If the model nevertheless calls a tool, preserve the paired
result and actual one-response-limit outcome. Do not hide tools from the existing
session or spend a fifth start to obtain a prettier answer. A reviewer may approve
a different four-start split before launch only if both phases remain; the default
published split is three plus one.

The principal task's code-instruction preload is not needed for these data-only
parity tasks, which must not edit program source. They use the accepted ordinary
primary and exact frozen task prompt. Their evidence covers human CLI continuity
on that route, not independent repair of the principal defect or GUI operation.

Preflight permits at most one selected-route model-list request per route, five
total, with no automatic pagination/retry for this exercise. Reuse still-valid
account-associated discovery where the inherited contract allows it, recording
its date rather than claiming it was freshly queried. Allow at most one inherited
refresh exchange per distinct selected existing Connection during preflight,
with its sanitized outcome recorded; no new authorization/login attempt belongs
to this budget. A needed but unavailable access/refresh/discovery step stops that
route. Model calls and lease-time refresh remain counted separately in their
respective receipts; this is not a claim that five GETs can establish entitlement.

Keep the existing per-operation refresh/retry rules; no capstone support script
may retry generation or switch funding. Use a 120-second deadline per model
operation, 20 minutes per eight-start task and 10 minutes per four-start task.
Each parity phase shares its route's ten-minute ceiling across restart. API-key
requests capture a positive 4,096 output-token cap; explicitly selected plan
requests have no remote output-token cap. Local limits and elapsed time cannot
promise a token, subscription or charge ceiling.

Stop an affected task on exhausted budget/deadline, unresolved ownership, an
out-of-scope edit, corrupted history, credential exposure or missing reliable
receipts. Use ordinary interrupt/close and retain completed effects. Continue
unblocked work only within its own allowance. A required missing route or outcome
prevents a successful crossover claim, even when another route did excellent work.

## 19.7 Review the work outside its explanation

Freeze the first patch, terminal/browser record and student experience before
comparative feedback. The independent reviewer evaluates what the files and
checks establish. Record source exposure and all outside assistance; don't call
a patched answer an independent cold success.

| Promise | Required distinguishing evidence |
|---|---|
| Real useful work | Brief and original failure precede the answer; actual permitted diff and meaningful checks; the original defect or a deletion of the candidate behavior fails the intended check. |
| Instructions delivered | Exact source/chunk/render coverage in the first captured request, including the complete skill; omitted/truncated manual fails before code is credited. |
| Product performed the work | Actual PTY prompts and accepted tool calls produce the files/test receipts. Outside commands are labeled verification, never substituted for a missing product action. |
| GUI path | Actual selected WebSocket MCP endpoint, effect acknowledgment and independent page/state evidence; catalog discovery or a stdio substitute cannot satisfy it. |
| Continuity | Same saved session, fresh mount, unchanged prior effects, truthful historical jobs and no duplicate cards/speech; a replay-and-repeat implementation must fail. |
| Ownership and identity | Independent public Agents; actual requested/returned/producing identities and explicit funding; no shared state, credential disclosure or paid fallback. |
| Honest accounting | All admissions/continuations/helpers and route caps included; unknown prices remain unknown, subscription use is not silently valued as an API invoice. |
| Bound record | Source/executable/launch identity checked before replay or derived writes; raw originals remain unchanged; every failed attempt remains attributable. |
| Regressions and judgment | Relevant predecessor, module, race and mutation controls; independent patch/teaching review and a scoped adoption decision supported by the outcome. |

The integration checker must include passing parents and deliberate failures for
missing instruction chunks, wrong source/binary identity, an omitted product tool
effect, a false test-success report, missing real GUI routing, repeated historical
effects and uncounted admissions. Each refusal must reach its intended predicate;
an earlier missing file cannot prove a later hash check. Local tests establish
these failure controls without manufacturing a real-provider failure.

If use exposes a product defect, name its responsible earlier chapter and repair
the teaching before affected code. Preserve the initial reproduction, review the
change and rerun its meaningful local and live paths under a separately approved
finite repair allocation. The original 40-start attempt does not restart or gain
an invisible extension. The capstone records the work required; it does not move
known architectural obligations into another repair chapter.

## 19.8 Taking it for a spin: the outcome is still ahead

There is no second-edition crossover transcript yet. The runnable exercise begins
only after the task-specific brief, accepted predecessor, complete checker and
bounded preflight are ready. Its eventual account should show a short original
exchange, the actual patch or effect, a useful correction if one occurred, and
the resumed session. Link detailed ledgers rather than making the reader audit
every file to understand the result.

End with one of three judgments: suitable for this specified supervised workflow;
usable with named restrictions; or not yet suitable because named required
outcomes failed. Explain what the next task can rely on and what the operator
still has to watch. A discovered defect is useful evidence. It is not a reason
to announce success before its correction is checked.

The first-edition crossover helped its authors learn what a workday demanded.
This one must earn its own account. Whether the completed second edition improves
on the first under matched conditions belongs to the final comparison chapter.
Whether Bill personally adopts it needs his actual experience. Neither conclusion
can be supplied by the agent's confident last paragraph.
