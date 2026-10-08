# Chapter 12 contract review

Independent coordinator review, October 8, 2026. Initial full draft: `46fd7d8`.
The reviewer read the complete new chapter, outline and research record, the
complete first-edition Chapter 13, current architecture, voice and chapter
procedure, plus the relevant published predecessor contracts. Root did not
author this chapter or implement a Chapter 12 solution. This review is of the
teaching contract; no runtime, live demonstration or final acceptance is claimed.

## Direction retained

Bill's required browser capability is explicit: actual MCP traffic crosses the
WebSocket through the optional GUI module and the Chapter 11 public transport
boundary. The view-before-Agent bootstrap respects creation-time tool binding.
Agent leases preserve shared resources; typed integration policy selects
automatic observation without making peer annotations authoritative. Capturing
owned observations directly in the attempted request avoids a second durable
queue with ambiguous cleanup.

The new draft distinguishes applied state from a sent command, retained text
from omitted text, and speech telemetry from heard audio. It also retains the
historical limitation that a fake stdio peer could pass while the shipped GUI
integration remained disconnected. Those distinctions are useful improvements
over the old exercise contract, not evidence that the new implementation works.

## Consolidated first-round findings

| Finding | Required revision | Status |
|---|---|---|
| GUI build path | Use the accepted `gui/cmd/ensemble-gui` command path; the draft's `./cmd` invocation from the GUI module is wrong | Sent to author |
| Human skill control | Distinguish read-only `/skills`, ordinary model-issued skill tools and the existing typed public API; do not imply an inherited load/unload slash command | Sent to author |
| Human draft demonstration | Preserve typing pause. Use an explicitly scoped independent observer/public action or an already-admitted race to test the browser's human-draft refusal | Sent to author |
| Browser request lifecycle | Define duplicate request admission separately from the client's reply classifier, with bounded per-channel/generation state and cancellation semantics | Author proposal reviewed; precise revision pending |
| Omitted byte count | Define the measured bytes so independent checks need not guess between omitted display text and encoded JSON | Sent to author |
| Phantom-panel story | Restore the concrete historical task, false report and missing selected-state mechanism; attribute the account and retain the missing-transcript limitation | Sent to author |
| Frame terminology | Count a complete WebSocket text message, including wire fragmentation, against the bound | Sent to author |

For the server lifecycle, the coordinator accepts a contiguous seen-ID prefix
and at most 64 disjoint ranges above it, independently of the 64 pending-operation
bound. Mark a fresh canonical request before effects; discard seen duplicates;
merge ranges and fault only that route before a request would exceed the range
budget. Pending cancellation fences that work; settled/unseen cancellation is
ignored without a tombstone. The tunnel preserves each request before its own
cancellation while allowing independent requests and completions out of order.
State belongs to channel and generation; old callbacks cannot act on a rebind.
Revision `2d4ea47` teaches this distinction before independent checker work.

## Revision review

Root reviewed the complete revision at `2d4ea47` against all seven findings.
Each is resolved: the build path names the actual command, skill control uses
the existing public API or model tools, human-draft testing preserves pause,
server admission has bounded duplicate tracking, omitted bytes have an exact
text definition, the phantom-panel account retains its evidence limitation,
and the frame bound covers a complete WebSocket message.

Independent parsing passed for all seven JSON fixtures. The three provider
examples carry identical canonical observation text. The maximum compact
envelope for an 8 MiB decoded message measures 11,184,968 bytes, within the
12 MiB contract bound. These are fixture and contract checks, not runtime
measurements. The revised contract is accepted for checker preparation;
student release still requires the accepted Chapter 11 implementation and
published independent checks.

## Source checks and limits

The reviewer opened the primary sources on October 8:

- [RFC 4648, sections 3–4](https://www.rfc-editor.org/rfc/rfc4648.html) supports
  the selected base64 alphabet, padding and canonical pad-bit requirements.
- [WHATWG WebSockets](https://websockets.spec.whatwg.org/#dom-websocket-send)
  supports the distinction between sending into a browser buffer and obtaining
  peer/application acknowledgment, and the need to account buffered bytes.
- [WHATWG DOM event trust](https://dom.spec.whatwg.org/#dom-event-istrusted)
  supports the distinction between programmatic dispatch and trusted events.

These checks support protocol explanations; they do not measure the proposed
queue limits or establish a working tunnel. Historical narration remains
attributed where raw run transcripts were not recovered. Unsupported debugger
timings and blanket claims about unsupervised success are not restored merely
to make the story livelier.

The author's fixture checks were independently confirmed during revision
review. The eventual student must
receive only accepted new teaching and predecessor source, not this review or
the author's historical research.
