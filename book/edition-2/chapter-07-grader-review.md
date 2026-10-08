# Chapter 7 independent acceptance plan

Contract-first milestone, 2026-10-07. This reviewer reloaded the full coding
skill, architecture ledger and Chapter 7 before writing the checker. The current
contract includes the incremental watch-projection paragraph at `2f105ce`.
No Chapter 7 student code exists or has been inspected. The role previously
reviewed Chapters 5 and 6, including their historical comparisons; it has not
opened the historical GUI answer for this new acceptance plan. Student coders
receive public requirements and results, never checker implementation.

The complete contract remains required. These are separate distinguishing groups:

| Group | Independent evidence and controls |
|---|---|
| Atomic watch | Hold before/after the actor snapshot cut at partial/final/tool transitions; snapshot plus strictly later revisions has neither a hole nor duplicate; include joining after begin before first delta |
| Window and ownership | More than 100 renderable events, omitted prefix and earlier-call result; mutate returned buffers; no HTTP or credential/config leakage; generation reset discards partial snapshots |
| Incremental projection | Fixed-size fragment doubling shows append-oriented cost; materialize owned snapshot at watch cut; preserve full identity, channel and bound accounting |
| Pause/admission | Two public registrations and two sockets with independent typing/speaking causes; pause observation precedes applied ack; exact barrier excludes later tool admissions; existing work and controls continue |
| Capacity/lifetime | Saturate watch and outgoing socket queues separately by count and encoded bytes; valid exact positives; close generation without dropping finals; other client and reliable handle work; deterministic selected-recipient/teardown race |
| Wire/projection | Exact Host/Origin/JSON/UTF-8/message/command-ID limits; correlated refusals; one serialized writer; recursive typed signature stripping with argument keys preserved; trace stages match actual queue/write/failure boundaries |
| Browser cards | Actual DOM for partial/final identity mapping, window reset, running versus terminal jobs, long-text expansion, accessible controls, focus and scroll preservation; malicious payload must actually reach every tested renderer |
| Speech | Mock queue A/B, cancel, then C and stale callbacks; typing survives speech transitions; replay silent and finals not repeated; separate real synthesis/audio evidence with browser/platform identity |
| Public composition | Two Agents with independent watches/registrations; reuse public Go and browser components from different layout; core builds with GUI module absent; inspect actual owners/imports/logger chains |
| Prior behavior | Retain Chapter 6 deterministic gate and CLI/protocol/plain/replay coverage; new pause observations do not alter old protocol output; initial live scope retains original source identities |

Every deletion or negative starts with a passing fixture and must reach its
intended assertion. Malformed setup or a prior rejection is not evidence for a
later guard. Examples include valid oversized JSON for the byte limit, a sender
known to have selected the departing recipient, and an XSS payload confirmed
present in the rendered card before asserting inert behavior. Public Go adapter
names will follow the student's reviewed ownership plan; the checker will not
invent method spellings or architectural requirements.

## Initial executable command

```sh
python3 scripts/edition2/accept_ch07.py /absolute/path/to/ensemble-gui
python3 -m unittest discover -s scripts/edition2 -p test_accept_ch07.py -v
```

Python `websocket-client` is required by this local checker. It launches the
published `--port 0` command in an isolated temporary workspace with explicit
fake model configuration. All model requests are counted by a local rejecting
endpoint; this milestone expects none. It never reads real credentials.

The initial executable contains 15 checks: empty snapshot wire shape;
two-client pause/count/ack/disconnect; a correlated idle hint refusal followed
by usable pause controls; four Host/Origin refusals; an exactly 65,536-byte valid
subscribe command; six malformed/over-limit cases; and absence of model HTTP.
(The listed totals are verified from the generator before publication.) It
uses ordinary WebSocket client framing under [RFC 6455](https://www.rfc-editor.org/rfc/rfc6455.html);
it neither specifies the student's server library nor adds one to the core.

Six assertion-control tests pass. They include positive empty snapshots and
pause changes, then targeted generation/watermark, null-range, credential,
cause-loss and acknowledgement-order failures. A deliberately absent server
(`/usr/bin/true`) fails at startup as expected, with a retained receipt. That
negative proves only the launch guard. There is no claim of a passing GUI
runtime or end-to-end positive transport fixture at this milestone. Initial
checker integration can reveal adapter/test defects; preserve those findings
separately from student defects.

This is a partial command, not a complete Chapter 7 score. Atomic public watch,
window seeding, admission barriers, both capacity queues, deterministic teardown,
projection, actual browser and speech tests remain to implement against the
published seams. Required real all-provider browser/CLI/public demonstrations
and independent historical comparison follow the initial source and student
experience freeze. Historical graders and earlier independent checks are
unchanged.
