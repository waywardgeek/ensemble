# Chapter 25: Crossing Over to Ensemble

“I gave Astra a book. Astra gave me an AI coding agent.”

That is how I described the crossover. The book included graders. Astra built
Ensemble while working in Codex, and then I began working with Astra inside the
agent we had built. The result was no longer just something to inspect. It was
the place where we were doing the next piece of work.

That distinction matters. I had supplied the book, the tests, architectural
requirements, and continuing engineering direction. This was not software
appearing without human expertise. Nor does moving the conversation into a new
application establish that the model carried its entire earlier context with
it. What crossed over was a working collaboration, supported by code, records,
and further correction.

This chapter is a historical interlude about that earlier, Go-based crossover.
The [crossover record](../../docs/edition-2-notes/ch25-crossover.md) is its source.
It preserves tests, failures, installations, and observations made at different
times. Those are not fresh results for this C# edition. There is no new
implementation exercise, grader, or source checkpoint in this chapter.

## Using the thing we had built

My first requests came from using the GUI, not from reading its source. I listen
quickly, and the speech-speed slider stopped at 2×. Increasing a maximum sounds
like a small change until the underlying voice rate turns out not to be linear.
For my Eloquence setup, I reported that native rates of 2.0, 2.6, and 2.9 sounded
approximately like 2×, 3.5×, and 5×. A native rate of 3.0 was already around 6×.

The correction used those observations to interpolate the higher settings and
changed both the control and saved-setting validation. The label said
*approximate*. That qualification was part of correctness: the mapping was
calibrated for this setup, not a universal timing law for speech engines.

Tests checked the actual control, saved settings, and native utterance rates.
A deliberately incorrect direct mapping made the new test fail. Those checks
established the wiring and mapping, not how quickly a human heard the speech.
Other speech-suite failures remained separate; fixing the slider did not make
the entire speech system verified.

There was another trap. Building a new executable did not change the executable
used by the launcher. Installing it did not restart the running process. Later,
restarting the server still left an open browser tab running old JavaScript.

These are separate operations:

1. Edit the source.
2. Build the replacement.
3. Install it at the intended launch target.
4. Restart the process deliberately.
5. Reload the browser assets when necessary.
6. Observe the changed behavior.

A successful earlier step is not evidence that a later one happened. It is also
not permission to interrupt somebody's working session.

## Observation should serve the work

I kept hearing “GUI snapshot” and “tts queue” while the agent worked. The browser
had registered both tools for automatic collection before every model request.
The results entered ephemeral context, and their ordinary tool events reached
the speech announcement path.

The diagnosis was not that snapshots accumulated forever. Replacement was
working: the model received the current snapshot and queue, not an ever-growing
stack. The unwanted behavior was repeated automatic collection and its noise.

I chose on-demand observation. We removed automatic registration for those two
tools while retaining their implementations and the general ephemeral-context
mechanism. The agent could inspect the browser when needed, act on visible
controls, and take another snapshot to confirm the result.

The distinction saved an unnecessary redesign. Silencing announcements would
have hidden the activity rather than stopped it. Removing all ephemeral context
would have discarded a useful mechanism to solve a narrower problem.

Here the historical record includes a live confirmation: after the updated
server was running, a hard browser refresh removed the old registrations.
Automatic attachments ceased, and explicit snapshot and queue calls still
worked. That confirmation belongs to this particular fix. It does not establish
the later configuration migration's still-pending live verification.

## Explain before acting

I require the agent to explain what it intends to do and why before meaningful
actions. Silence followed by a long retrospective does not provide the same
oversight. The conversation explains intent; the Actions pane shows execution.
Both need to remain available.

Private reasoning is not that explanation. Neither is an encrypted replay item,
a spinner, or a provider's brief thinking title. A request for more detailed
provider summaries also cannot guarantee that the provider will return them.
The ordinary visible explanation has its own purpose and must stand on its own.

The earlier implementation added a model capability for enforcing explanation
presence and order. An unexplained batch was refused before any tool effects,
with paired corrective results so the model could try again. No minimum length
was imposed. A mechanical check can establish that text preceded tools; it
cannot certify that the text explained anything useful.

Streaming made the rule more subtle. A parser could receive a tool fragment
first and explanatory text later, then normalize the final result into
text-first order. Inspecting only that final representation would falsely
credit an explanation supplied after the proposed action. The correction
checked actual stream arrival order too.

The lesson is broader than this one policy: test the boundary at which a
requirement matters. A tidy final representation can erase the evidence needed
to enforce it.

## A check-in is not a deadline

A job returning control after three seconds had caused confusion. That interval
meant “look again,” not “terminate the work.” Yet a different timeout really was
ending model requests: the HTTP client imposed a sixty-second deadline on the
whole response.

The historical diagnosis followed the failure to that client. A request begun
at one timestamp failed a minute later with the client-timeout error. The fix
removed the total response deadline while retaining connection protections and
explicit interruption and shutdown. It did not add automatic retries.

The useful test was deliberately longer than the old limit. With a local
provider boundary and the real actor and HTTP parser, both a streaming response
and a response withheld before headers completed after sixty-one seconds.
Restoring the old client made the new streaming test fail at about sixty
seconds. Separate tests showed that interruption and shutdown still canceled
requests.

That is stronger than making a short test green. The corrected implementation
survived the condition that had killed legitimate work, and the negative control
showed the test could detect the original defect. It was still local evidence,
not a claim about a live provider's response duration.

## Roles must agree with their result contracts

The human conversation and a delegated worker do not have the same completion
protocol. A main agent can finish with an ordinary reply. A worker may owe its
parent an explicit structured submission. Confusing those roles produced more
than awkward wording during the crossover.

In one later live experiment, an auxiliary compressor was required to call
`submit(memory)`, but its input also contained the main-agent instruction not
to submit. The provider opened a stream, produced substantial ordinary prose,
and completed without a function call. The runtime rejected the result and
retained the source instead of committing a memory replacement.

This was not a silent provider stall. Nor was it successful compression merely
because useful-looking text had arrived. The output contract failed. The
conflicting instructions were observable; the model's internal reason for
choosing prose was not.

An earlier interrupted wait remained a different incident. A later completed
failure did not prove that both had the same cause. Similarly, disabling all
memory bands and successfully recording a handoff isolated the checkpoint path;
it did not repair the compressor.

Moving compression into the background would not, by itself, solve contradictory
instructions. A restricted worker needs an unambiguous role, a bounded result
contract, observable completion, and parent-controlled acceptance and cleanup.
Chapter 22 teaches those responsibilities in C#. Its local verification is
recorded there, not borrowed from this historical experiment.

## Configuration should be inspectable

A question about reasoning settings exposed launch state that was difficult to
inspect: the running request asked for high effort and automatic summaries, not
the detailed summaries I expected. That led to a larger request for explicit
file-based application configuration.

The earlier Go implementation made the selected file authoritative, kept
credentials external, and distinguished current settings from saved changes
that required restart. It also preserved the legacy startup path used by tests
and graders. A save had to succeed before the application published the new
settings as accepted.

Review found a serious detail that a successful launch would not have caught.
Loading a credential before registering flags could place that credential in
the default value printed by help or error output. The fix kept the override
flag's default empty. A negative control restored the unsafe behavior and made
the regression test fail for the leak.

The source record says the replacement was installed and its configuration
validated. It also says the final restart, browser refresh, and live
configuration verification remained pending. This chapter does not silently
finish those steps.

The C# edition has its own ownership and persistence design. In particular,
Agent-home memory and recall policies are separate from shared GUI appearance
preferences. Do not import the old Go application's single-file schema merely
because the historical lesson is sound. Carry forward the requirement—clear
authority, secret-free configuration, save-before-publication, and honest
active-versus-next-launch state—not an unrelated implementation shape.

## Identity outlives a process

Another crossover report concerned new streaming text appearing in an old card.
The earlier runtime allocated artifact IDs uniquely within one process, but
restarted its counter at zero after restoring history. Saved cards still had
the old IDs.

An isolated two-runtime reproduction made the collision visible. The first
response, `FIRST`, used ID 1. After save and reload, `SECOND` also used ID 1.
Both remained separate in durable history, while the renderer appended the new
delta to the old card and then replaced its text with the final response.

The archive had not lost a message. The projection had confused two identities.
Clearing the display every turn would hide the symptom, not satisfy the identity
contract. Starting future IDs above the old maximum would also leave existing
collisions in older history unresolved.

The record labels this investigation **reproduced, not fixed or deployed** at
that boundary. Its lesson is an obligation to test identity across restoration,
not evidence of a repair that the record never demonstrates.

## What the crossover contributes to the book

These incidents are why I do not treat generated code or a green grader as the
end of engineering. A real user exposes assumptions about speech, observation,
latency, roles, settings, and identity that a construction exercise may miss.
The response must be diagnosis and correction, not a claim that the user used
the program incorrectly.

The transfer of knowledge goes in both directions. An expert helps build and
check the system; demonstrated work then supplies better explanations and
requirements for the codebook. A fresh agent can execute that improved book
again. The code is a rebuildable output, but its quality still matters: it is
where the teaching meets real behavior.

For this edition, Chapter 25 adds no hidden repair pass. Chapter 24's live OAuth
and cache decision remains open, and its optional workaround has not been
implemented. Chapter 26 continues from the locally verified Chapter 23 C#
foundation. The next feature must earn its own implementation and evidence.
