# Chapter 2: One History, Three APIs

An interface built around one provider's message format makes every other
provider translate through it. The translation then spreads into history,
tool handling, and accounting. This chapter gives the conversation its own
meaning and puts each provider's format at the edge.

## 2.1 The idea in plain words

**The log records what happened.** A user supplied a prompt, the engine
sent a request, a model returned parts and usage. Each fact receives an
ordered identity. Changing what the next request should contain does not
change the record of what previously occurred.

**Context is the current reading of that record.** A reducer applies the
events in order. If an event says to replace a tool result with a stub,
the next context contains the stub and the log still contains the original
result. Replaying the same events reconstructs that choice without asking
a model or deciding again whether redaction was a good idea.

**A request is a provider's view of the context.** A tool result might
become a user-message block on one API and a separate tool message on
another. The author of the result remains a tool. A wire convention
changes the rendering, not who produced the information.

Keep Chapter 1's ownership chain and package boundaries. Shared events,
parts, and interfaces belong in `internal/common`; reducer, rendering,
and parsing behavior belong in `internal/llm`. Ensemble still owns the
logger and Agents, Agent owns configuration and its conversation, and
Engine owns transport and accounting. Replace the narrow text slice
with events and a derived context inside that architecture.

> **Contract draft.** The command and schema details below are being
> reconciled with independent fixtures. This chapter is not yet the
> student handoff. No implementation or live run is claimed.

## TL;DR

Continue the validated Chapter 1 repository and its history in
`solutions/edition-2/ch02`. Keep its public library and CLI. Add three
non-streaming adapters: Messages, Chat Completions, and generateContent.
Read current model IDs from provider discovery for live runs.

1. Agent owns an append-only, versioned JSON-lines event log and its
   derived context. Each event has a positive sequence number, kind,
   timestamp metadata, and exactly the matching payload. Sequence is
   ordering authority. Reject malformed records, unknown kinds, and
   unsupported versions with line-numbered diagnostics.
2. Use one append/apply path for human inputs, parsed responses, and
   recorded redactions. Parsers return facts; they cannot mutate context.
   Copy slices and raw bytes when taking ownership. Replaying a fixed
   log rebuilds the same context and usage without network calls.
3. Entries have actor, purpose, sequence, and ordered typed parts. Keep
   text, calls, results, references, opaque replay material, and redacted
   stubs distinct. Purpose distinguishes dialogue, enduring instruction,
   and one-request ephemera; actor alone cannot make that distinction.
4. Every model response and opaque/call-bound replay part records vendor,
   exact model, and API surface. Replay opaque bytes only to matching
   provenance. Preserve issued call IDs; deterministic synthesized IDs
   cover routes that omit them. Do not dispatch tools in this chapter.
5. Normalize response usage into disjoint ordinary input, cache write,
   cache read, and output counts. Keep the producing provenance and raw
   usage observation. Engine aggregates by provenance; changing current
   configuration cannot change historical attribution.
6. Render without mutation or network access from context plus explicit
   request configuration. Keep credentials outside log, context, and
   rendered output. Fixed log, configuration, and adapter produce
   identical bytes. Consuming pending ephemera belongs to the recorded
   request event, never to the act of rendering.
7. Record result redaction as an event naming an inclusive sequence span.
   Replace result content in context with a deterministic stub, preserving
   call/result linkage and any external reference. The original log stays
   unchanged. Do not add an unbounded second set of redacted sequence IDs.
8. Preserve Chapter 1's successful CLI protocol and public-library use.
   Add offline `render LOG` and `dump` commands plus explicit log selection.
   Valid present empty text and tool-only responses are now representable;
   absence, malformed payload, and HTTP failure remain distinct errors.
9. Expose public, transport-neutral client requests and observations on
   Ensemble. Both CLI and a GUI stub use them. The GUI stub lives in a
   separate optional Go module; the core has no GUI or WebSocket dependency.
   Observer events carry Agent identity; synchronous requests return their
   own answers instead of waiting for a broadcast to imply completion.
10. Demonstrate every implemented feature through actual user paths.
    Run the CLI with all three real APIs, plus an external public consumer.
    Label GUI-stub integration and deterministic failure fixtures as such;
    neither is a working live browser demonstration.

**Yours.** Internal names and exact method signatures, provided the
ownership and schema contracts hold. No tool execution, jobs, actors,
streaming parser, skills, automatic compaction, retries, or functioning
browser interface is required here.

**Exercise.** The inherited grader is a useful regression baseline:

```sh
make grade-dir CH=2 DIR=solutions/edition-2/ch02
```

Its passing score is insufficient for the new ownership/client/validation
requirements. The final acceptance table and literal fixtures must accompany
this contract before the student begins.

## Contract details to complete before handoff

The following work remains on the author and grader engineer, not the
student: print the event payload and part schemas, an independent literal
tool-history fixture, exact ephemera/redaction directives and acknowledgments,
log path rules, usage normalization formulas, and the three request/response
wire subsets. Publish the GUI stub's public seam and evidence expectations.

## Taking it for a spin

[LIVE RECEIPTS PENDING: all three actual CLI API paths; recorded history
and offline replay; ephemeral delivery; result redaction; independent
Agents through an external executable; clearly labeled GUI-stub integration.
No real-model response or successful acceptance result has been invented.]
