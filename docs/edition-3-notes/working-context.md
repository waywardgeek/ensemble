# Per-model working context

## Contract

The Agent owns durable limits keyed by vendor and exact routing model. Values are
whole provider-reported input tokens, not bytes or implicit thousands. The default
working limit is 250,000, reduced to a known smaller capacity. Unknown capacity
does not disable reminders. A supplied positive capacity constrains the working
limit; it is not a tokenizer or a guarantee that the next request fits. Astra's
approximately one-million-token capacity is Bill's configurable report, not newly
verified catalog metadata. The existing byte stubbing setting is unchanged.

At 80% the next request gets one system-generated reminder per checkpoint cycle.
At 90% every model request gets a reminder. At 100% the Agent uses the existing
safe handoff boundary after companion results are recorded. The fallback explicitly
says no model-authored continuation was supplied; no model tool call is fabricated.
Independent Jobs and ordinary tool authority remain intact. Completed handoff
invalidates pre-cut usage and resets the reminder cycle. Missing usage is unknown.
Provider normalization already counts cached input once. Direct Engine sends,
compressor usage and session totals are not this measurement.

The status percentage appears immediately after Agent status, including Paused.
Its detail identifies last reported input and the working limit. Model changes
clear usage, while per-model settings persist. Settings save before publication;
limit and ratio publish together. Form rollback does not replace live usage, and
form identity stays attached to its displayed values across model changes.

Original teaching is in Chapter 21 section 21.6 and its companion; Chapter 14
identifies the separate byte stubbing control. No legacy value such as 250 is
converted to 250,000 tokens.

## Verification and limits

Evidence is retained at `/private/tmp/cadence-working-context.qgHvfV`.

- Two broader core runs passed 351 groups with zero failures. The later run covers
  the reminder-admission correction, but predates the final atomic-publication
  lock. Final focused tests after that lock pass values/strict codec, durable
  restart/sibling settings and real request threshold behavior.
- Real provider-fake tests cover just below 80%, exact 80%, once-only delivery,
  repeated 90%, cached input, two finalized effect results before automatic
  checkpoint, no fabricated tool call, and no stale measurement retrigger.
  Additional pressure tests cover missing usage, direct/auxiliary isolation,
  model switching, unchanged authority and independent Job survival.
- Real GUI WebSocket settings tests pass durable acknowledgment, sibling-policy
  preservation, selected-Agent isolation, unknown values and refused saves.
- The final embedded-page test passes exact token fields, last-input percentage,
  Paused placement, limit changes, draft preservation, reload, automatic-handoff
  invalidation and two real filesystem save-failure cases. No injected application
  frames or private browser state are used. Fixture/provider/Agents shut down and
  the host exits zero.
- Both current Go grading packages pass. No public artifact points were changed.
- Chapter 14 and 21 lint reports retain only their previously existing em-dash
  failures, compared against HEAD; companion hard checks pass with advisories.

Failures remain evidence rather than being rewritten as success. An initial
canonical-result assertion used a tool name instead of CallID and was corrected.
Earlier page runs selected a model absent from recommendations, then a model whose
catalog row could not accept the fixture image. A dedicated mode of the existing
fixture uses its supported model without changing production capability metadata.
An attempted direct shell invocation of the nonexecutable runner failed before
startup; invoking it through sh ran the intended fixture.

Two actual UI defects were reproduced before correction. Failed-save rollback
replaced a newer 31% measurement with unknown; `page-rollback-before.log` records
that assertion. After a model switch, rollback restored the previous model's form;
`page-model-before.log` records Haiku where Opus was required. Corrected runs pass
in `page-rollback-after`, `page-atomic-final`, and finally `page-model-after`.
The final core/GUI-settings logs are `atomic-core-tests.log` and
`atomic-gui-tests.log`; final GUI build is `model-form-build.log`.

This is source/build and owned-fake/browser evidence, not installed-binary,
live-provider, native phone, or universal concurrency certification. No installed
settings, application restart, package installation, push or publication occurred.
Idle-transition goal reminders are a separate next upgrade. Bill temporarily
requested the existing idle timer at one second; that control change does not
implement the queued lifecycle policy.
