# Chapter 8 source research

Author/reviewer-only material, October 7, 2026. Do not give these historical
sources to the fresh student. This is research for new Chapter 8 (old Chapter 9),
not a ready contract or a claim of implementation/live validation.

## Reads

Read the full current voice and chapter procedure in this author phase; read
all of old `book/chapter-09.md`, the full old CH9 checks and harness, the current
architecture ledger, current Chapter 7 handoff/owner sections and the global
map's settings/UI, GUI-testing and later integration lessons. Read the exact
old Chapter 13 settings-validation/zero-broadcast account and old Chapter 22's
settings closure/dead-value passage. Relevant old CH9 source history was read
with `git log` and exact diffs; no student source was edited.

## Historical leads and corrections

- `ca775b6` identifies the old chapter's self-use motivational addition. Keep
  the reader's practical reason to customize an interface; do not import an
  unmeasured productivity comparison or claim the reader already used it.
- `3d7b7d19ea14c774b0f2480868b3d43af1ac523e` adds explicit obligations after
  defects survived the original teaching: input/load validation, patch versus
  snapshot, and observable control state. The old phrasing that sparse patches
  need `omitempty` because zero is “unset” is itself insufficient. New teaching
  must track field presence separately so explicit zero/false remains writable.
- Old Chapter 13 records temperature -5 accepted, then a corrected zero missing
  from the broadcast while the browser still displayed -5. It records speech
  false omitted for the same reason. Its later measurement paragraph says zero
  was absent after a repair described as removing omission; that internal
  contradiction is a reason not to reproduce the measurement as proof. Use the
  mechanism and source-backed incident, with new independent fixtures later.
- `2b2fedbc8b50bc688c31ab823712553e589f8a3b` adds the fourth obligation: prove a
  saved setting affects execution. The old GUI saved max_tool_rounds=200 while
  two loops still used16. Preserve the missing-consumer lesson, not old loop
  architecture or default/tool-batch semantics. New Chapter 5 already owns turn
  decisions and snapshots configuration at turn start.
- The global map and old Chapter 22 identify settings closures bridging an
  unreachable owner and settings values with no consumer. GUI persistence must
  not become a second Agent configuration authority.

## Old checker coverage observed from source

CH9 checks compare sent theme/tts_speed values in the acknowledgement, a
nonempty current-settings object on subscribe, dark theme in a second client,
and dark theme after restart. They also invoke old CH8 parity. The harness does
not demonstrate explicit false/zero, wrong-type atomicity, failed writes,
versioned disk load, stale revisions, semantic browser control state, actual
speech behavior or an execution-setting consumer. No new deletion run was made
by the author, and source inspection is not a measured checker result.

Its executable layout, `--gui-dir`, handshake and old fake identity differ from
new Chapter 7. Retain its historical diagnostic identity; the independent new
checker must derive from published new teaching rather than force the student
to restore old wiring.

## Decisions to publish next

The outline proposes Server-owned GUI preferences and Agent-owned execution
settings as distinct domains, with actual owner interfaces and public commands.
The coordinator accepted that scope before the complete student contract:
persisted display defaults are shared, actual page speech/input/pause is locally
owned, and Agent changes remain actor-ordered. The full draft must define when
a preference update applies in each tab and preserve active-turn snapshots. Shared complete
snapshots must be safe public projections, never serialized credentials or
creation-only log identity. Missing values, explicit values, invalid values and
failed writes need different documented outcomes.

No new external browser/API claim has been made in this research pass. Current
platform behavior needed by the eventual browser contract will be checked
against official sources when drafting. Existing browser tooling availability
is coordinator evidence; it does not prove the new interface or audible speech.
