# Chapter 8 independent checker preparation

Coordinator engineering from the reviewed new Chapter 8 contract at `7200f17`,
before any Chapter 8 student implementation. No old settings implementation or
new student answer was read for these checks. The full mandatory coding skill
was reloaded before this coding task; architecture and complete new contract
were read. No historical grader or frozen solution was changed.

The initial partial invocation is:

```sh
python3 scripts/edition2/accept_ch08.py GUI_BINARY
```

It reuses only the Chapter 7 checker's WebSocket client transport helper. Its
own local backend counts requests and refuses all HTTP; no paid credential is
read. Temporary settings paths and fresh conversation logs isolate each launch.
The checker binds its script, helper and supplied executable hashes.

## Initial scope and actual validation

Six advertised groups cover full startup snapshots, sparse preference writes
including true-to-false and two-client broadcast, no-change revision/file
identity, positive-to-zero policy and watch notification, correlated invalid
value/conflict refusal with subsequent valid recovery, fresh-process reload,
and the absence of model HTTP. The first five functional groups share a
documented lifecycle sequence; an earlier failure prevents the later groups
from being credited. The separate no-HTTP row cannot turn that incomplete run
into a pass.

The command currently has **no end-to-end Chapter 8 positive**, because no
student server exists yet. Six assertion-control unit tests pass, including
valid false/zero values and intended failures for omitted false, numeric
Boolean, wrong effective policy, unsafe policy fields, mismatched watch
revision, wrong preference publication order, no-change broadcast and Boolean
revision. Policy wire observation is deliberately allowed on either side of
its acknowledgement: actor publication ordering is a different property.
Running against `/usr/bin/false` produces an explicit failed startup result;
that retained negative is not a substitute for a working settings server.

The no-change file check compares bytes, inode and modification time; identical
serialized bytes alone would not detect an unnecessary replacement. No browser
rendering, actual speech or execution-budget effect follows from this checker.
No broad legacy rerun was needed for an independent new checker that edits no
shared grader; earlier legacy evidence remains separately attributed.

## Coverage still required before acceptance

- Strict patch and complete-file validation: duplicate/null/unknown fields,
  exact 64 KiB positive and one-byte-over negative, missing fields, wrong types,
  trailing input, unsupported versions and unreadable files. Preserve original
  invalid files; a startup timeout is not an intended validation rejection.
- Controlled disk write/sync/close/replace failures, exact old bytes and applied
  state, temporary cleanup, close during an in-flight writer, and independently
  responsive controls. One successful disk write does not establish atomicity.
- Concurrent same-base writers, busy versus stale replies, explicit retry,
  separate-domain progress, atomic subscription cut and replaced-socket callback
  rejection, queue overflow and no lost applied updates.
- Public headless policy ownership/isolation, actual HTTP limit 1 and 17,
  default 16, final tool-batch pairing, active/queued turn capture and durable
  historical-policy validation/reconstruction. Do not infer policy use from its
  echo or persisted value.
- Real browser theme/font/width changes, pointer/keyboard dividers, selected
  control state, two ArtifactScroll instances, mixed-part identity and malicious
  content, reconnect and retained Chapter 7 behavior.
- Mocked speech revision/rate capture, buffered-versus-queued text, local cancel
  and independent typing causes, plus separate actual synthesis/audio receipts.
- Full independent code/ownership inspection, meaningful implementation
  deletion controls from passing fixtures, retained prior acceptance, actual
  three-provider browser/CLI/public use and later historical comparison.

Publish additional commands as those independent surfaces are prepared. The
initial invocation does not waive §8.9 or permit a validated Chapter 8 tag.
