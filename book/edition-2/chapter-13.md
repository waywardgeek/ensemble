# Chapter 13: Test the speech channel

This chapter is for me. I have 20/180 vision from macular dystrophy, close
enough to legally blind that most of my screen is a guess. I work by
listening. The agent narrates its thinking, announces its tools, explains its
reasoning aloud, and that spoken channel is the interface I actually use.

Accessibility is an afterthought everywhere, never in the critical path of
shipping a product. Nobody discovers a broken speech pipeline until somebody
who depends on it sits down and hears silence. Since I am in control of this
book, this agent gets decent accessibility from the start.

A correct answer can conceal a broken interface. An observer that reads the
conversation can identify a failed command even when the person listening to
the application hears nothing. The answer passed. The channel was never tested.

The first test given to a restricted listener was designed to fail. Tool
results never reach the speech channel; only dispatch announcements do. The
task was to run a command against a path that does not exist and report the
exit code. The prediction, written down before the run, was that the listener
would come back empty.

It came back with the correct exit code.

Six utterances entered the channel. The first was the dispatch announcement.
The rest were the model's own prose, describing what happened. No tool result
appeared in the transcript. The listener learned the exit code because the
model chose to mention it, not because the system delivered it. From inside the
channel, narration sounds the same as delivered data. Accessibility that rests
on a model's prose habits is discoverability by luck.

The transcript also held two utterances that were nothing but a single
backtick: markdown fence markers, read aloud as punctuation. A visual card and
a spoken explanation are different channels, and each can break while the other
works. Bill listens to this agent for hours a day and steers it by ear. He
cannot audit the speech channel by looking at the screen. If the pipeline
swallows a sentence or reads punctuation aloud, he is the one who hears it,
and he is the last person anyone builds a regression test for.

This chapter records what the speech service accepts, tests the text that
reaches its output adapter, and gives a separate listener access only to that
channel. Actual browser audio remains a separate check. A transcript cannot
establish whether a voice was audible or a sentence intelligible.

*Student contract draft for review. The [validation
record](chapter-13-validation.md) tracks the unfulfilled gates. The accepted
Chapter 12 design has no validated implementation yet; see its [gate
record](chapter-12-validation.md). The [preparation
evidence](chapter-13-evidence.md) preserves the historical sources and
limitations. No Chapter 13 implementation or successful spin is claimed here.*

## TL;DR

Read the complete [coding skill](skills/ensemble-coding/SKILL.md) and
[architecture](architecture.md). Extend the accepted Chapter 12 source when
available. Keep the core library usable without the optional GUI module.

1. Add the bounded incremental automatic-speech normalizer in §13.2. Its output
   is invariant under source chunk boundaries. Preserve Chapter 7 part identity,
   final-only delivery and silent replay/results. Overflow visibly abandons only
   that part's unqueued remainder; previously admitted speech retains its owner.
2. Keep Page-owned buffers and queues, BrowserApplication-owned SpeechService,
   native arbitration and Chapter 8 preference semantics. Provide native and
   recorder output adapters created by that actual service. Recorder mode is an
   explicit test-application choice, never a silent native fallback.
3. Add a bounded service-owned speech journal with strict records, scoped
   cursors and explicit eviction gaps. Record queue acceptance, native submission,
   start and terminal status distinctly. Recorder completion is its own phase.
   A native end callback is never called “heard.”
4. Add an opt-in, separately prepared listener MCP endpoint through the existing
   GUI WebSocket tunnel. Keep Chapter 12's five-tool endpoint unchanged. Freeze
   the listener Agent's three aliases before construction; grant no DOM, file,
   artifact, general GUI or automatic-context access to the target's answers.
5. Speak a static safe notification for a new owned turn/transport failure once,
   only while autoplay is enabled. Silent tool results, intentional cancel,
   reconnect/replay and synthesis failures do not become announcements.
6. Ship a public listener consumer and a bounded harness. The external harness
   exports journal pages and owns its file, browser and child-process lifetimes.
   A failed export invalidates the evidence run without disabling native speech.
7. Test whole-browser wiring as well as pure text fixtures. Prove listener
   isolation, phase/provenance selection, cursor gaps, overflow, shutdown and
   inherited native/pause ownership. Exercise real CLI/browser/public paths on
   all three providers; retain native audio separately from recorder receipts.

Public method/class names remain student choices. Exact observable values,
limits, wire fields and behavior below are required. From
`solutions/edition-2/main/`:

```sh
go build -o /tmp/ensemble-ch13-cli ./cmd
go vet ./...
go test ./... -count=1
```

From `solutions/edition-2/main/gui/`, build with `go build -o
/tmp/ensemble-ch13-gui ./cmd/ensemble-gui`. Run formatting, vet and tests in
every affected module. From the repository root, retain `make grade-dir CH=14
DIR=solutions/edition-2/main` as a historical diagnostic. An independent new
checker command must be published before student handoff; §13.9 defines the
required checks. The historical score alone is insufficient.

## 13.1 Follow the text to its destination

The original direct-module tests could call a speech helper and observe good
output while the running application never called it. A later repair connected
a failed turn's existing error field to an existing error helper.
Both ends had been built. A listener still received silence.

Another repair removed a tool result emitted a second time as assistant text.
The browser's filter trusted that provenance and read the file's contents
aloud. Fixing the speech filter would have treated the symptom. The source
notification had to tell the truth about who supplied the text.

Preserve Chapter 7's automatic sources: new visible answer text, explicitly
exposed thinking and concise tool intent summaries. Tool results, durable skill
manuals, automatic observations, replay and opaque provider material stay
silent. Chapter 13 adds only the narrow failure announcements in §13.6.
On-demand card speech remains an explicit user action and retains its full
accessible-text behavior. It is identified as manual output in the journal.

Each automatic source has a full owned identity: Page/mount, runtime Agent,
request, model operation, part and source kind. Reused provider-local part IDs
cannot merge two responses. Queueing advances that source's consumed position.
An accepted final contributes only text that its preceding deltas have not
already supplied; a final-only response uses the same normalizer. A provisional
source canceled by interrupt does not become accepted final text merely because
the reader saw it.

The recorder must receive the same normalized text the native adapter would
receive. Logging a second independently rendered conversation proves only that
both renderers can read history. Route both adapters through one SpeechService
admission boundary and retain the selected adapter mode in every journal
record.

## 13.2 Normalize incrementally without losing the fence

The fenced-code failure looked like a streaming problem at first. Deltas
arrive on arbitrary boundaries, a fenced code block split across two chunks
gets filtered as two halves, and neither half contains a complete fence, so the
markers survive. That explanation was clean, mechanical, and wrong. The same
input failed in a single chunk. Newline splitting had separated the fence
markers from the body before the filter examined them. Testing more stream
splits could not repair the wrong order of operations.

Use the following deliberately small grammar for automatic speech. It is not a
complete Markdown parser or a language-specific pronunciation engine. Apply it
to Unicode scalar values in source order. Network chunk boundaries have no
meaning; hold an incomplete UTF-8 scalar at the byte decoder and an incomplete
surrogate pair at the browser string boundary. Invalid input follows the
inherited text-validation refusal rather than silently inventing replacement
words. Treat CRLF as one LF and lone CR as a space.

1. At column zero, a line consisting of three backticks, an optional language
   token `[A-Za-z0-9_-]{1,32}`, optional spaces/tabs and LF opens a fenced block.
   The same valid opening line at accepted EOF also opens an empty block. A
   closing line consists of three backticks followed only by spaces/tabs and
   LF or EOF. Emit `Code block.` once when recognizing the opening line; discard
   the body incrementally until the closing line. An unclosed block stays
   discarded at end. Fence bodies need no accumulating text buffer. A line
   that fails this exact grammar follows the ordinary rules below.
2. Outside fences, remove a line-leading heading marker of one through six `#`
   followed by one ASCII space, or a line-leading `- `, `+ ` or `* ` bullet.
   There is no indentation rule or ordered-list rule in this subset. Recognize
   this structure before joining lines. Other `#`, `-` and `+` remain text.
3. A single-backtick inline span ends at its next backtick before a paragraph
   boundary. Remove its delimiters and normalize its content as ordinary text.
   It cannot open a fenced block. At paragraph/accepted EOF, an unmatched single
   backtick is removed and its retained content follows the same rule. Runs of
   multiple backticks outside recognized fences are removed, without treating
   their following text as a block body.
4. Recognize `[label](destination)` and `![label](destination)` with no nested
   brackets/parentheses or newline in either field. Retain the label and discard
   the destination. Normalize the label as ordinary inline text. An incomplete
   or nonmatching candidate is literal text at the paragraph/end boundary.
   Never fetch a link or parse it as an instruction.
5. Remove `*` emphasis markers outside literal link candidates. Replace `_`
   between ASCII letters/digits with a space; remove other underscores. For
   ASCII letter runs, split lowercase-to-uppercase and the boundary before an
   uppercase letter followed by lowercase when the preceding letter is uppercase.
   Thus HTTPServer becomes `HTTP Server`, RPGLit becomes `RPG Lit`, camelCase
   becomes `camel Case`, and ALLCAPS stays unchanged. Preserve case and digits.
6. Collapse spaces, tabs and a single LF to one space. Two or more successive
   LF, with only spaces/tabs between them, terminate a paragraph. Trim utterance
   edges. A period, question mark or exclamation mark followed by whitespace or
   accepted EOF terminates an utterance, retaining that punctuation. A decimal
   point between digits is ordinary text. Other punctuation is retained.

Hold an ambiguous suffix until its classification is known. For example, a
period at the end of a delta may be followed by a letter rather than
whitespace. An inline candidate may contain sentence punctuation before its
closing marker; do not commit that punctuation until the candidate resolves.
After recognizing a link/inline span, process its retained text through the
same sentence rules. At a tool boundary or accepted part end, flush any
nonempty ordinary remainder. A fence marker's `Code block.` is its own
utterance, flushing preceding ordinary text first. Blank paragraphs and empty
filtered bodies enqueue nothing.

At most 64 automatic part parsers may be pending on one Page. A 65th receives
the same scoped overflow disposition below before allocating a parser, with
visible text `Automatic speech skipped: too many pending parts.` Existing parts
remain admitted. Retire parser state at its terminal boundary rather than
retaining it for every historical card. The pending budget is 8,192 UTF-8 bytes
per automatic part: unresolved source bytes plus normalized text not yet
admitted to the queue. Count each retained representation actually stored; do
not retain a second unlimited copy in the normalizer. Already discarded fence
content contributes zero. Small finite parser state, counters and at most three
incomplete UTF-8 bytes are separate. Process input incrementally; receiving a
large delta cannot allocate its whole size again before checking this budget.
The existing artifact may retain its ordinary bounded text independently.

At exactly 8,192 pending bytes the part remains valid. If the next scalar or
normalization expansion would exceed it, discard that part's unqueued
remainder, mark the source exhausted and ignore later deltas/final text for
automatic speech. Show `Automatic speech skipped: text exceeded the speech
buffer.` on that Page with the affected source identity. Record one overflow
fact. Clear that parser's pending state and reconcile pause immediately; retain
speaking while already queued/current utterances remain. Do not cancel those
utterances, other parts or another Page. Never split an overlong token or dump
suppressed code to make the record fit. A subsequent part starts with a fresh
budget and can speak normally.

The same scalar-by-scalar decision applies to a whole final response and any
partition of its deltas. Turning autoplay off discards unqueued automatic state
and advances consumed positions as Chapter 8 requires; enabling it later does
not reread that skipped text. Cancel/interrupt/disconnect fences the
appropriate source generation and clears pending state. A canceled source is
never flushed by a late final callback.

These exact arrays are fixtures, with no trailing spaces inside utterances:

| Source text (JSON string) | Ordered automatic utterances (JSON array) |
| --- | --- |
| `"Reading **now**. HTTPServer RPGLit ALLCAPS."` | `["Reading now.","HTTP Server RPG Lit ALLCAPS."]` |
| `"# Plan\n- first\n- second\n\nDone."` | `["Plan first second","Done."]` |
| `"Use \u0060read_file\u0060 and [the guide](https://example.invalid/a)."` | `["Use read file and the guide."]` |
| `"Before.\n\u0060\u0060\u0060go\nsecret()\n\u0060\u0060\u0060\nAfter."` | `["Before.","Code block.","After."]` |
| `"\u0060\u0060\u0060\nunfinished secret"` | `["Code block."]` |
| `"Version 1.25 works. Tail:"` | `["Version 1.25 works.","Tail:"]` |
| `"é🙂\nnext."` | `["é🙂 next."]` |

Run each as a single source and split at every scalar boundary, plus byte
splits through the multibyte fixture at its decoder. Hold the final-only
fixture and its streamed equivalent to the same outputs. For overflow, 8,192
lowercase `a` bytes followed by accepted EOF produce one utterance; 8,193
produce one visible overflow and none. Prepend `Ready. ` to that latter source:
`Ready.` remains admitted, and only the following oversized remainder is lost.
A huge discarded fence body remains bounded without producing its body or an
overflow merely because those discarded bytes are numerous.

## 13.3 Put the recorder on the actual service

Page owns parsing and its logical queue. BrowserApplication owns SpeechService,
which owns its output adapter and journal. An adapter receives the actual
service parent; helpers reach diagnostics through that route. Preserve the
existing optional-module boundary and public component construction. The Agent
library has no dependency on browser classes, speech APIs or WebSocket framing.

The native adapter uses Chapter 8's same-origin/storage-bucket lease. Record
native submission only after obtaining that lease and immediately before the
actual speak call. Start the inherited no-start timer at that point. A waiting
Page keeps its speaking cause; a canceled waiter aborts its wait without
calling native cancel. Only the active owner can cancel native work, and
terminal cleanup precedes release. Closing the application stops admission
before closing Pages.

A reader canceling one panel should still hear the other panel finish. The
recorder must exercise that ownership boundary too, even though it settles much
faster than a voice can speak.

The recorder adapter receives the identical service request, records
recorder_complete and settles it once without invoking native synthesis.
Schedule that completion through the service's ordinary serialized path, so a
canceled request cannot complete a replacement or clear another Page's pause.
Recorder mode does not acquire the native lease or emit simulated native
callbacks. Both modes preserve FIFO, ownership, queue identity and captured
preference values.

Select the adapter when constructing BrowserApplication. The optional launcher
adds `--speech-output native|recorder`, default native. Recorder requires the
explicit listener/debug test configuration in §13.5 and displays a persistent
`Speech output: recorder; no native audio` status. Ordinary startup retains
native behavior. A native error never switches modes. Switching an occupied
application requires closing it and constructing another owner.

Automatic normalization does not replace the manual card speaker's full-text
contract. Journal text is bounded as described next, so a large manual
utterance may have a visibly incomplete diagnostic prefix while its playback
remains unchanged. The listener acceptance task uses complete automatic
records.

## 13.4 A journal that can admit what it lost

SpeechService creates a fresh 32-lowercase-hex service identity from 16 random
bytes. It owns one ring of at most 4,096 records and 8,388,608 canonical JSON
bytes, counting records themselves without array separators or framing LF.
Evict oldest whole records until both limits hold. Closing a Page leaves its
records until ordinary eviction; closing the application disposes the journal.
This is a diagnostic store, outside the Agent's durable conversation and
session snapshot. Restart creates a new service identity and no recovered
speech work.

For the reader investigating a missing sentence, a gap is more useful than a
plausible but incomplete transcript. Make eviction visible before any consumer
can mistake the remaining suffix for the whole conversation.

Each strict version-1 record has exactly these fields:

```json
{"version":1,"seq":1,"elapsed_ms":0,"view":"main","mount":"0123456789abcdef0123456789abcdef","agent_id":"agent-1","request_id":"request-1","operation_id":"model-1","part_id":1,"connection_generation":"g1","utterance":1,"generation":1,"source":"answer","mode":"recorder","phase":"queued","text":"Ready.","text_bytes":6,"omitted_bytes":0,"truncated":false,"sha256":"e20846d20bcd60f0f57708fb7e0afdc584cfac4711ea4bb636dfce053ad50136","rate":1,"preferences_revision":0,"code":""}
```

This is a literal record fixture with the actual SHA-256 of UTF-8 `Ready.`. A
record's text is the first at most 8,192 UTF-8 bytes of the actual normalized
service text, ending on a scalar boundary. Text_bytes is its full UTF-8 byte
count, omitted_bytes is exactly that count minus the prefix length, truncated
is exactly omitted_bytes > 0, and sha256 hashes the exact full text offered to
the adapter. The journal performs no trimming or further normalization,
including on manual text. Hashing must not allocate another full copy. The
maximum canonical record is 65,536 bytes; reject an invalid oversized producer
value before journal insertion. The text prefix and bounded identity fields
below fit even with JSON escaping.

Seq is positive uint64, local to the journal. Utterance is a positive uint64
diagnostic identity assigned to a Page-owned speech request; elapsed_ms is
nonnegative uint64 milliseconds from the journal's monotonic start, ordered
nondecreasing with seq. Equal times are legal. These recording counters do not
own actual playback. Generation records the Page's positive cancellation
generation, whose live fencing remains independently required. Preserve exact
integer tokens through the browser/public seam.

The journal also owns bounded out-of-band status: journal_status is recording
or faulted, and journal_error is null while recording or counter_exhausted or
invalid_record while faulted. A required record that would overflow a
diagnostic counter faults recording before allocation or wrap; an invalid or
oversized producer record faults it before insertion. Latch the first error,
retain all previous records and the last written seq, and show `Speech
recording is incomplete.` A fault requires no additional journal record or
sequence number. Later recording attempts cannot clear the error, evict
retained records or advance the written sequence. A newly constructed application obtains
a fresh service identity; clearing, remounting or reconnecting a Page does not
reset this state.

A recording fault must not stop native speech admission, output or cleanup.
Existing and later native work continue under actual owned request identities,
FIFO, cancellation fencing and the native lease, without requiring a new
journal ID. Another Page's playback remains independent. Never wrap or reuse a
live ownership identity because its diagnostic representation failed; an
ownership identity's own exhaustion still follows its lifecycle refusal rule.
This distinction lets a failed diagnostic store remain outside speech
authority.

Recorder mode requires a recorded completion. If the queued or
recorder_complete fact cannot be recorded, fail that recorder delivery
explicitly, settle its owned work once and release its speaking cause.
Pending/later recorder requests also fail promptly while faulted; do not invent
recorder_complete, wait forever or switch to native output. This output failure
remains visible outside the failed journal and invalidates the listener
evaluation.

View/mount follow Chapter 12. Agent, request, operation and
connection_generation use existing public string identities, at most 256 UTF-8
bytes each. Part_id is the inherited positive uint64. Request/operation/part
are null only when inapplicable, such as manual output; connection_generation
still names the Page connection. These fields preserve Chapter 6's full part
identity, including its model operation. Do not rewrite a larger identity by
truncation; report a safe diagnostic refusal. Source is answer, thinking,
tool_summary, manual, failure or overflow. Mode is native or recorder. Rate and
preferences_revision are the values captured when queued; overflow has null
rate/revision/utterance because no utterance was admitted. Code is empty on
success or one of the safe codes below. No raw exception, provider response,
credential or hidden reasoning appears here.

Phases have precise meanings:

| Phase | Fact recorded |
| --- | --- |
| queued | The service accepted this Page's complete utterance and immutable text/rate. |
| submitted | The native lease is held and the service is invoking native speak. |
| start | The matching live native start callback arrived. |
| end | The matching native end callback settled this utterance. This does not establish hearing. |
| recorder_complete | The recorder received this text and settled its request. No native output occurred. |
| canceled | The owner canceled an unsettled request; code is canceled. |
| error | Output failed; code is unavailable, start_timeout or synthesis_error. |
| overflow | This automatic part exceeded its pending budget; source is overflow, code is speech_overflow and text is empty. |

Overflow's text count/hash describe empty text; it retains the affected request
and part. Other phases repeat the immutable utterance text metadata. Queued is
first; a native request can fail/cancel while waiting without submitted/start.
End requires submitted, but start may be missing if the platform omits that
callback. Recorder_complete requires queued and recorder mode. Exactly one
terminal phase end/recorder_complete/canceled/error settles a queued utterance
while recording is healthy. If recording faults before that phase, actual
playback still settles, but its terminal record may be absent. Every later read
must expose the sticky fault; absence then cannot be presented as a complete
trace. Ignore stale callbacks after settlement and record no invented success.
Ring eviction is a different loss: it may leave a suffix without an earlier
queued record, and its gap must be reported even when recording remains
healthy.

The public journal read and the listener's speech_read tool accept exactly
service, mount, after and limit. Public reader construction fixes its
authorized scope and profile: diagnostics can select all phases; the restricted
listener profile includes only recorder_complete from answer, thinking,
tool_summary or failure. The model has no profile selector. Native-mode
speech_read refuses with listener_unavailable; native diagnostics use the
explicit public reader. Service/mount select the explicit endpoint's scope;
after is a nonnegative uint64 cursor, limit an integer 1–64. An endpoint is
bound to one view, mount and Agent at creation; the caller cannot choose a
different Agent. A stale service or mount refuses. Version is integer 1; all
sequence/cursor/count fields use exact uint64 tokens. Canonical byte accounting
uses Chapter 10's canonical JSON representation. New subscriptions begin with
after 0; this means start of this service's retained history, with any loss
reported, not “ignore whatever happened before now.”

The structured read result has exactly version, service, view, mount, agent_id,
after, first_available, through, next, more, gap, journal_status, journal_error
and records. Transport success means a diagnostic read succeeded; it does not
assert that recording remained complete. Both reader profiles expose the sticky
status on every read, even when records is empty or the cursor is already at
the last written sequence. The status fields count within the existing response
byte cap. A faulted result remains readable for diagnosis; the public listener
consumer must stop with an explicit incomplete evaluation, and export must fail
with complete false and error journal_fault. Retained recorder_complete entries
cannot override that fault. File-export failure alone does not fault recording.

Through is the service's current last seq at the read's coherent boundary;
first_available is the first retained seq, or 1 before any record has been
written. The ring is never cleared during a service lifetime; disposed services
refuse reads. Gap is null or exactly from/to, the lost inclusive service
sequence interval after the requested cursor and before first_available.
Records contain only the selected scope and reader profile and are ordered by
seq. Service-wide sequences can skip other Pages; that alone is not a loss. A
reported eviction gap may include other Pages' records; it makes no claim about
how many selected utterances were lost. It never exposes those Pages' IDs or
text.

Advance next over every scanned service record, including records outside the
selected scope/profile. Stop before a selected record that would exceed limit
or the 524,288-byte complete canonical result cap; next must not consume that
record. Set more exactly when next is less than through. Scan at most the
retained 4,096 records; a snapshot read performs no wait and no browser/model
I/O. After beyond through is invalid. When an eviction gap is reported, begin
scanning at first_available and advance next past the lost interval. A single
record fits the response budget. Empty results can still advance a cursor past
other Pages.

For example, with retained service sequences 4–9, request after 1 reports gap
2–3. If only 5 and 8 belong to this endpoint and limit is 1, return record 5,
next 7, more true, through 9. The next call after 7 returns record 8 and
advances to 9. An exporter must not silently call either page a complete trace
starting at sequence 1.

If a queued record takes the last available sequence, the next required phase
faults recording. Through freezes at that last written value; never manufacture
a missing sequence or advance next beyond it. Reads keep their usual filtering,
limits and gap semantics, with journal_status faulted even after the caller has
consumed every retained record. A listener-profile read at that cursor can be:

```json
{"version":1,"service":"11111111111111111111111111111111","view":"main","mount":"0123456789abcdef0123456789abcdef","agent_id":"agent-1","after":18446744073709551615,"first_available":18446744073709551615,"through":18446744073709551615,"next":18446744073709551615,"more":false,"gap":null,"journal_status":"faulted","journal_error":"counter_exhausted","records":[]}
```

This is a constructed maximum-counter seam, not a claim that a short complete
history reached that value. A new service rejects the old service ID; recovery
never makes an old cursor refer to newly recorded speech.

## 13.5 Give the listener a smaller endpoint

Chapter 12's target endpoint still advertises exactly gui_snapshot, tts_queue,
gui_click, gui_input and gui_submit. The optional launcher adds `--listener`,
valid only with `--gui-debug` and an explicitly selected view. It prepares a
second logical endpoint after the target Page exists. Public embedding exposes
the same operation; it does not require private launcher functions.

This explicitly extends the GUI protocol with listener_attach and
listener_attached; it does not reinterpret Chapter 12's mcp_attach. The request
has exactly type, version 1, id, view and mount. Success has exactly type
listener_attached, version 1, the same id, view, mount, channel and generation.
Use the existing ordinary 65,536-byte command limit and correlation budget.
Reject missing, duplicate, unknown or wrong-type fields before attachment. The
application must have authorized the listener kind for this existing live
view/mount; otherwise return the correlated gui_unavailable refusal.

Attachment identity is endpoint kind plus view plus mount within the owning GUI
application. At most one live attachment of each kind owns that key. The target
and listener can coexist because their kinds differ. A listener attachment
borrows the selected Page scope; it neither remounts nor reclaims the target
view. Duplicate listener attachment refuses without replacing the original. An
explicit detach/close releases that endpoint's key after its work is fenced. A
new mount requires a new authorized attachment; old callbacks cannot write into
it.

Use a distinct application-selected channel, such as listener_main. The actual
MCP Connection supplies its generation before listener_attached and discovery,
through the same pending-endpoint construction sequence as Chapter 12.
Thereafter reuse unchanged mcp_frame and mcp_detach. Root's 32
prepared-connection ceiling, per-channel operation/byte budgets and
per-physical-socket limits count BOTH endpoint kinds together. No second
physical socket is required. Closing the listener route leaves target route,
watch and native speech intact; physical socket failure still fails its child
routes honestly.

This endpoint advertises exactly three remote tools: speech_read,
listener_input and listener_submit. Prepare/discover it before constructing the
separate listener Agent. Its installed ceiling contains only those aliases, its
system instructions describe the bounded transcript exercise, and it has no
skills or automatic integration sources. Use a fresh listener Agent for each
acceptance run. Ordinary public model/request configuration and credentials
still follow the inherited rules. The endpoint is application-selected
authority; an arbitrary model argument cannot register a new view or find
another Agent.

Return the service/mount identities to the public consumer at preparation, so
its listener instructions can name them without a DOM snapshot. No speech text
is included in that preparation value. Closed-Page records remain readable
through a retained scoped read lease until journal eviction or explicit
endpoint close; action calls after Page close return listener_unavailable. A
replacement mount requires a newly bound logical endpoint and explicit
compatible preparation. Old callbacks, cursors and input calls cannot target
its new owner.

Speech_read has §13.4's exact arguments/result with its mechanically enforced
completion-only profile. It never returns queued/submitted text, manual records
or native phases. The transcript-only evaluation requires omitted_bytes 0 in
every relevant returned entry. Public diagnostic records cannot be supplied to
this listener as another source of answers. A recording fault, gap, unexpected
mode or truncated relevant entry is an explicit incomplete evaluation, rather
than a successful complete-transcript assertion. There is no native “heard
transcript” mode. Public diagnostics can inspect native phases separately.

Listener_input takes exactly mount and text, with the inherited 4,096-scalar
limit, and writes only the selected Page's prompt through its Chapter 12
programmatic-input action. Listener_submit takes exactly mount and mode, where
mode is prompt or hint, and submits that draft through the same Page action.
Both retain human-draft conflict checks at the point of effect. They neither
simulate trusted typing nor acquire user-gesture privileges. They return
exactly view, mount, outcome and request_id, using edited/admitted and
null/request ID as Chapter 12 defines. They expose no control ID, arbitrary
selector, artifact, settings value, hidden draft or model answer.

All three input schemas forbid extra fields and require the exact argument
objects taught here, using Chapter 11's schema profile. Return content as an
empty array, structuredContent as the exact success/failure shape, and isError
false/true respectively. Descriptions expose scope and limits without target
answer text. Schema/refusal errors retain Chapter 11 MCP behavior. Local tool
failures contain exactly error and message with static safe text. Codes are
listener_unavailable, listener_stale_scope, listener_invalid_cursor,
listener_human_draft and listener_action_failed. No failed action returns
target text. Closing the listener Agent cancels only its own calls; root
retains the shared connection and socket. Closing its dedicated endpoint
releases its own scope lease, preserving the ordinary GUI watch and target
endpoint. Repeated/canceled RPCs use Chapter 12's bounded duplicate-effect
protection; there is no second action queue or retry.

For a meaningful listener test, put an independently generated nonce only in
the target's assistant output. Do not place it in the listener's initial task,
its preparation metadata, a shared file, tool description, settings or
automatic observations. Ask the listener to report that nonce using the allowed
speech channel. Inspect its real offered tools and request bodies. In a
denied-delivery control, remove recorder completion delivery while preserving
queued diagnostics and the target's visible answer; the listener must not
receive the nonce through another route. A guessed privileged alias must fail
at actual dispatch.

A separate provenance control puts one marker only in a tool result and another
in assistant prose. Only the assistant marker reaches automatic output. Do not
ask the listener to recover the silent marker and then call its failure an
accessibility defect. An explicit manual speaker action is labeled manual and
is excluded from this automatic-channel acceptance task.

## 13.6 A failure needs a short explanation

The reader who hears a turn stop needs to distinguish completion from failure.
Add an automatic announcement for a new completion with outcome error for an
owned live request, including model HTTP/stream transport failure. The exact
text is `The request failed; check the displayed error for details.` A
transport failure is the cause of that same failed request, not a second
announcement. Never read a raw provider body, URL, exception, stderr or
credential-bearing diagnostic.

Use the Page's request identity and connection generation to deduplicate this
selection. Only a request already tracked by that Page's accepted live-request
lifecycle can select an announcement. Settled/replayed requests cannot be
reintroduced by an unsolicited completion, and no forever-growing error-ID set
is required. One failed completion queues at most one failure utterance;
duplicate completion/observation deliveries cannot repeat it. Capture the
current applied autoplay preference/rate at that decision, with source failure
and the failed request ID. If autoplay is off, mark it consumed without
queueing so enabling speech later does not announce an old failure. The
failure's ordinary visual status and existing CLI error remain independently
available.

For this chapter, transport means the model/request transport failure reported
through that owned completion. A GUI socket disconnect retains Chapter 7's
cancel-and-release behavior and visible connection status; reconnect does not
speak its recovered errors. Intentional interruption, application close,
round_limit, a denied individual tool and a nonzero command exit are not failed
turn announcements. A failed tool result stays silent; later assistant
narration of it is ordinary answer text. Output-adapter failure produces an
error journal record and a visible local status, never another speech request.

Flush only already accepted ordinary text before queueing the failure notice.
Discard unresolved provisional output from a failed model operation under the
inherited streaming contract. Neither a late model final nor a failed native
callback can revive it. The announcement uses the same service path as other
speech, so tests must reach the real completion handler rather than call a
speech helper directly.

## 13.7 Export a bounded observation, then close its owner

A log file is useful when the defect happened several seconds ago. It does not
need another continuous browser-to-server writer. The public journal cursor and
listener endpoint already provide bounded, coherent pages. An external harness
owns exporting them to its selected output file and can report failure without
changing the target's speech mode.

Ship `examples/listener-consumer` as a public example in its own module. It
constructs the GUI application, prepares the scoped listener endpoint, builds
the restricted Agent and exposes native/recorder runs. It must use public
constructors and complete-message transports. It cannot import another module's
internal packages, substitute a fake DOM for its actual-browser demonstration
or access the target's private conversation to help the listener answer. A
headless consumer still builds and uses Ensemble with no GUI dependency.

Also ship an executable `scripts/speech-harness` in the student tree. This is a
public acceptance entry point; its implementation language and private names
are free. `scripts/speech-harness --describe` writes exactly one JSON object
with version 1, read_tool, read_argument and modes. The first two strings
describe the installed target read-file tool and its path argument; modes is
exactly `["recorder","native"]`. No network or child launch occurs for
describe.

A run accepts `--plan ABSOLUTE_JSON_PATH --output ABSOLUTE_NEW_DIRECTORY`. The
plan is strict version 1 with exactly mode, prompt, workspace and operations in
addition to version. Mode is recorder or native; prompt is at most 16,384 UTF-8
bytes, workspace an existing absolute directory. The target uses the ordinary
shared LLM configuration reader; the independent checker supplies its local
endpoint and fixture model through that reader's documented environment. Never
require real credentials for a local fake endpoint. Paid runs use the same
interfaces with separately authorized credentials and source bindings.

Operations is an array of at most 64 exact objects, in order. Each has action
and value. Supported actions are human_type (string of at most 4,096 scalars),
human_clear (null), cancel_speech (null), wait_turn_end (null), and wait_ms
(integer 0–5,000). These are test-driver actions on the selected Page. Human
input uses the actual browser input path; assigning a property and dispatching
an untrusted event cannot prove a browser gesture-dependent action. Submit the
plan's prompt once through the Page's ordinary prompt action before executing
operations. First enable autoplay through its ordinary control and wait for its
applied acknowledgement; do not change the product's default-off behavior. The
driver waits for that submission's admission acknowledgement; it does not wait
for completion before subsequent operations.

The minimal local plan is:

```json
{"version":1,"mode":"recorder","prompt":"Read the supplied file and summarize it.","workspace":"/absolute/fixture","operations":[{"action":"wait_turn_end","value":null}]}
```

The harness waits at most 120 seconds overall from launch to owned cleanup,
including readiness and all waits. Expiry fails the run; it does not fabricate
turn completion. Recorder mode drains admitted speech before final export;
native mode drains admitted speech too, or ends with an explicitly recorded
cancel when the plan requests it. Produce a final scoped journal read before
disposing the service when possible. No slow export may park the Actor or grow
an unbounded memory queue. The bounded journal is the source of exported pages;
polling does not create a second unbounded list in the browser or harness.

Create the output directory exclusively; refuse an existing path. Write a
strict manifest with exactly version, mode, service, view, mount, agent_id,
profile, sources, binaries, plan_sha256, artifacts, started_at, ended_at,
complete and error. Version is 1; profile is diagnostics or recorder_complete.
Identity strings follow the journal; times are UTC RFC3339Nano strings.
Sources, binaries and artifacts are arrays of unique objects with exactly path
and sha256, using documented base directories and lowercase SHA-256 values.
Require nonempty sources/binaries, at most 1,024 source entries, 64 binaries
and 4,096 artifacts, within the byte cap below. If startup fails before
attachment, unavailable runtime identities may be null only in a complete:false
manifest.
Bind actual source and executable bytes before launch; never hash credential
files. Plan_sha256 binds the exact input file. Complete is Boolean. Error is
null on success or one of launch_failed, export_limit, export_io,
scope_changed, journal_gap, journal_fault, timeout or incomplete_listener on
failure; report a more specific safe explanation outside the machine value.
Write journal reads as `journal.jsonl`, one complete §13.4 result per LF-framed
line, in cursor order, at most 16,777,216 bytes including LF. A version-1
manifest is at most 65,536 UTF-8 bytes and may not contain credentials or raw
environment values. Persist a page before requesting another; retain the last
persisted cursor and never mark a skipped page complete. Each line has the same
524,288-byte result cap plus its LF. On export cap, filesystem failure, scope
mismatch, eviction gap or recording fault, visibly fail with complete false. A
truncated record remains an honest export; complete here means complete cursor
export, not full text or native hearing. The transcript-only listener
separately refuses a required truncated record, while the oversized-manual
control expects it. Preserve partial original evidence. Do not truncate the
file to turn a failed run into a passing one.

Record actual browser actions, received public events and screenshots
separately, with bounded original artifacts and their hashes in the manifest.
Native runs also bind actual audio and capture metadata if that gate is
claimed. An absent or empty audio artifact cannot be replaced by a synthesized
waveform or a start callback. These support artifact spellings may be
documented by the student; the declared journal, manifest fields and limits are
the common contract.

Every normal/failing exit closes admitted calls, listener/target Agents,
endpoint leases, browser contexts, application and owned child processes in
their proper order. Terminate and join process groups the harness created on
timeout; do not kill unrelated processes by name. Export failure fails the
evidence run; during the remaining owned cleanup interval the journal stays
bounded and native speech continues under its ordinary cancellation rules. The
harness may then close its own application normally. A diagnostic failure is
never an excuse to silently switch the output adapter or stall another
application.

## 13.8 Taking it for a spin

The demonstration is pending. The student must use the built application before
this section can contain observed answers, screenshots, usage or audio results.
The following sequence is a reproduction plan, not a transcript.

Build the CLI and optional GUI as shown in the TL;DR. After selecting a real
provider through the inherited shared configuration, run human chat in a PTY:

```sh
/tmp/ensemble-ch13-cli chat
```

Ask for a bounded file read and summary, observe the answer, then follow up on
its result. Keep that CLI parity receipt separate from the speech experiment.
Start the optional GUI with an explicitly selected view:

```sh
/tmp/ensemble-ch13-gui --gui-debug --gui-view main --listener --speech-output recorder --terminal
```

Complete Chapter 12's view attachment and target preparation. Enable autoplay
through its ordinary control. Ask for two short sentences, a fenced code
example and a final sentence. Compare the actual deltas, visible answer and
journal rather than assuming the model followed the formatting request. This
launcher demonstration exercises the human terminal and its browser. Preserve
its source, service and launch identities when closing it.

Run a separate public-consumer demonstration from `examples/listener-consumer/`
with `go run . --mode recorder`. The example constructs its own application,
prints its own local URL and uses the shared provider configuration for fresh
target and listener Agents. Attach that new view and enable its autoplay, then
carry out the independently seeded hearing task on its own scoped endpoint. The
listener's answer must come from complete recorder_complete entries in this
second application's healthy journal; it cannot establish delivery by the first
application's target. Retain this run's own source/launch identities, offered
tools and provider requests. The example has no attach-to-another-launcher
mode. Its native mode uses `--mode native`. Document actual readiness and
action controls alongside the example so the reader can repeat each observed
run.

Repeat the feature matrix on each supported provider with bounded calls and
fresh source/launch bindings. Use deterministic local fixtures for exact
grammar, forced transport failures, disabled aliases, counter exhaustion,
journal gaps and timing races. A model need not reliably emit a chosen
malformed fence for that failure boundary to be testable.

For native speech, start a separate application with `--speech-output native`.
Use the same bounded content, actual user activation if required, real browser
controls and captured audio. Exercise two Pages and cooperating tabs under the
inherited native lease: cancel one Page's waiting work without canceling the
holder, then verify the holder's owned completion. Record native submitted,
start and terminal events beside audio timestamps. Preserve missing voices,
blocked playback, timing limitations and incomplete captures explicitly.

Also exercise an oversized manual card. The actual output adapter must receive
its full accessible text while the journal truthfully reports a bounded prefix.
Do not pass that record to the fast listener as complete evidence. For active
speech or human-draft conflict inspection, use an explicitly granted
independent observer/public consumer; typing or speech on the target Agent can
prevent it from admitting a new self-observation tool.

The eventual manuscript should include one real screenshot with descriptive alt
text, an abridged journal excerpt and the listener's actual answer. Link the
raw source-bound receipts. Neither a recording nor an observer's summary
establishes that Bill personally tested it or that a human found the voice
intelligible. Those claims require their own actual participation.

## 13.9 Checks that can fail for the right reason

The old direct-module grader coupled tests to private JavaScript methods. Its
later harness revision tested the real event path, but several assertions could
still pass on an unrelated utterance or a logged pause without proving the
promised effect. Keep those historical checks as diagnostics and add
independent fixtures at the published boundaries.

| Required property | Distinguishing positive and negative controls |
| --- | --- |
| Exact normalization | Literal §13.2 arrays, every scalar split, decoder byte splits, paragraph/link/fence boundaries and final-only equivalence; deleting fence-before-line handling must expose the intended failure. |
| Bounded incremental state | Exact pending limit and +1, large single delta, long token, huge discarded fence, retained earlier utterance and next-part recovery; no growth proportional to discarded source. |
| Provenance | Independently planted tool marker stays absent while assistant marker arrives; delete source discrimination and require the negative assertion to fail. |
| Whole-event wiring | Failed owned completion reaches the actual Page handler and exact safe announcement once; unrelated opening text cannot satisfy the check. |
| Silence rules | Replay, reconnect, failed tool, round limit, deliberate cancel and synthesis error produce no failure announcement or recursive output. |
| Adapter equivalence | Native and recorder receive byte-identical service text/rate on a controlled adapter seam; recorder produces no native calls or native phases. |
| Manual compatibility | Oversize manual text reaches the adapter whole; journal prefix, digest, omitted bytes and truncation agree, without applying automatic overflow. |
| Identity/lifetime | Two Pages, repeated provider-local IDs, stale callbacks, replaced mount, cancel and close; no cross-owner settlement or disclosure. |
| Native arbitration | Retain same-context multi-tab lease and unavailable-platform controls; actual audio evidence is separate from controlled callbacks. |
| Journal bounds | Record and byte eviction, cursor exact/+1, coherent reads, scope/profile filtering and gaps; counter/invalid-record faults stay visible after cursor exhaustion, preserve native work and fail recorder/export/listener completeness. Recovery requires a new service identity. |
| Listener attachment | Authorized target/listener coexistence, malformed/unauthorized/duplicate refusal, stale mount/generation, shared budgets and independent close. |
| Listener isolation | Real fixed ceiling, guessed alias refusal, no DOM/automatic observations, planted answer obtainable only through recorder_complete; denied-delivery control exposes no answer. |
| Export and cleanup | Source-bound passing export before changed identities, truncated/gapped/file-failed negatives, bounded bytes; normal and timeout teardown with a proven-running owned canary. |
| Public/module boundary | Independent public consumer in another module, headless core build, owner/logger paths and all existing parity checks. |

Run original positive controls before mutations and preserve their exact source
identities. Use valid owner-state seams for exhaustion rather than inventing a
short history that somehow reached the maximum counter. Scope each mutation so
an unrelated setup failure cannot masquerade as the intended rejection.

The exercise resolves a practical uncertainty: the reader can inspect which
text reached which output path, and the listener cannot quietly obtain it
elsewhere. Whether the native voice is pleasant at the reader's chosen speed
still belongs to actual use, beyond the transcript's claims.
