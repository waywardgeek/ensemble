# Chapter 9 initial actual demonstrations

All planned N/F/G/P interfaces were exercised with the three real providers on October 8, 2026. This is the initial student checkpoint before independent historical comparison, **not chapter acceptance**. Original failures and incomplete model tasks remain below. The operator was the Codex student coder, not Bill. Human CLI commands used actual PTYs; browser actions used headful Chrome and the public consumer used two independently configured Agents.

[Machine-readable results](live-results.json) retain each launch identity, actual calls/results, final states, material digests, file bytes/hashes, reported model identities, accepted usage and audio measurements. Each `live-*` directory retains original terminal, request/response, event-log, catalog and scratch evidence. GUI directories also contain original browser frames/actions, screenshots/text, native WAVs and final browser manifests. Browsers closed before the GUI launch was sealed. No original receipt was rewritten by verification.

## Models and limits

Fresh discovery selected Anthropic `claude-sonnet-4-6`, OpenAI `gpt-4.1` (reported `gpt-4.1-2025-04-14`) and **exact** `models/gemini-3.8-flash` (explicit wire mapping/report `gemini-3.8-flash`). No fallback occurred. Discovery receipts are `discovery-{anthropic,openai,gemini}.json`.

| Provider | N | F | G | P | Generation HTTP | Discovery HTTP | All HTTP |
|---|---:|---:|---:|---:|---:|---:|---:|
| Anthropic | 14 | 5 | 9 | 4 | 32 | 1 | 33 |
| OpenAI | 11 | 5 | 8 | 5 | 29 | 1 | 30 |
| Gemini | 16 | 5 | 7 | 4 | 32 | 1 | 33 |
| Total | 41 | 15 | 24 | 13 | **93** | **3** | **96** |

All failures and corrections are included, below both36/provider and108overall even including discovery. Coordinator message `bounded-public-recovery` moved only OpenAI's unused sixth F call to P: N16/F5/G10/P5. Other caps stayed N16/F6/G10/P4. Ledgers were never reset. There are no further paid calls planned.

## Feature/provider results

N/F/G/P below link to the original launch/terminal folders. “Observed” means the named real action/effect occurred; it does not convert every attempted task into a success.

| Feature/action from reviewed matrix | Anthropic | OpenAI | Gemini |
|---|---|---|---|
| Source/primary, `/skills` before first prompt | N revision0, base, narrow declarations, zero prior generation calls | Same N observation | Same N observation |
| Load edit and actual write | N revision1; read dependency activation2, edit3; `notes-created.txt`13 bytes | Same | Same before later transport failure |
| Fixed primary and once-expanded material | All11 N request bodies contain exact primary once; captured System empty; material digests match | All11 N bodies match | All8 initial N bodies match, including failed attempt |
| Dependency offer, explicit roots, unload | N review/search roots at revision4; read2 survives; edit3 retired; write absent | Same | Same in initial N; recovery also independently preserves search after edit unload |
| Post-revocation write request | Model refuses without reloading; exact file unchanged | Model refuses without reloading; exact file unchanged | Request captured with write absent; relay transport failed, **no model refusal observed**; file unchanged. Local forced-call tests establish admission enforcement |
| No-op and absent skill | Actual search unchanged + absent refusal, revision4 stable | Same | Final N management run loads review, then actual unchanged review + absent refusal at revision1. Search-name no-op was not completed; equivalent operation on review observed |
| Full ensemble primary/all12 declarations | F actual notes→summary,62 bytes | F actual notes→summary,55 bytes | F actual notes→summary,59 bytes |
| Literal-next-call limits | F limits(delay0, bytes1)→inactive unload consumes→full untruncated read. No management job | Same | Same |
| Manual chronological placement and immutable lifetime | Original requests exact; retired manual survives separate load-result redaction | Same | Same, using G's retired edit manual |
| Shared browser/terminal Agent | G browser load/write, PTY inspection then review/search/unload; same agent-1, revision4 | Same | Same |
| Current sidebar and retained cards | Both browser tabs reconnect; revision4 roots/grants, retained edit marked retired, actions carry ordinary management records | Same | Same |
| Safe long display/full expansion | Actual click expansion and literal `<example>` retained as text; screenshots/text | Same | Same, plus separately bound focused **Enter** action on live retired edit card (`keyboard-gemini-live`) |
| Explicit manual speech/native evidence | Two original8s Chrome-only WAV captures; full2455-character manual passed to speech; nonzero native waveform after silent control | One8s capture and matching full text | One8s capture and matching full text |
| Autoplay/retirement/reconnect | Subsequent ordinary answer/tool speech observed; no automatic manual start on transition or reload; pause released | Same | Same; immediate cancel screenshot precedes asynchronous release, reconnect shows speaking0 |
| Two-Agent ceilings/configuration/getters/material/contributors | P typed alpha edit, beta edit refusal, independent literal bindings and actual reads bothAgents. **Exit1:** beta redundant load/unload exhausted policy; final beta report absent. Coordinator accepted required Skills coverage independently | Original P transport failed before actual read. Approved P recovery completes both reads, exit0; alpha reports alpha-$TOOLS, beta reports beta rather than full configured literal. Captured beta material remains beta-$TOOLS | P both actual reads and reports, exit0; alpha-$TOOLS and beta-$TOOLS retained |
| Owned getters/creation-only replacement | Local public controls, not a fabricated live mutation | Same local public controls | Same local public controls |
| Exact offline reconstruction/usage | All32 attempted request bodies accounted for; public failed4 covered by independent audit7dc9cc1 | All29 exact, including failed transport attempt | All32 exact, including both failed N attempts |
| No catalog/network dependency offline | OS-denied catalog read control fails; exact original N request replay succeeds with all network denied | Same | Same |
| Retained compatibility | F ordinary tools, G streaming/settings/speech; full67-row initial gate and source-specific repairs accepted | Same common implementation/local gates | Same |
| Malformed catalogs/graphs/limits; forced denied calls; concurrency/persistence; counter/precision/card-window boundaries | Required deterministic controls retained at their source identities; these were never claimed as reliably elicited live corner cases | Same | Same |

## Attempts and failures retained

- `live-anthropic-g`: first turn refused Messages `list_directory` start input{} plus empty `partial_json`; no early effect. Old runtime c0e3171. Explicit-path bounded recovery completed the actual GUI write. Runtime repair75a7255 preserved valid start input when zero bytes arrive; all11 module checks and narrow inherited gate passed. `live-anthropic-n-repair` elicited the same real empty delta and completed in3 calls on the repaired binary. Independent14-control review7da5052 is separate evidence.
- `live-anthropic-p`: actual reads for bothAgents but beta final prose missing at round limit; exit1 retained. Independent audit7dc9cc1 verifies4 byte-exact requests and chapter-required public coverage. No paid repeat for greener prose.
- `live-openai-p`: first attempted request hit the relay's local transport-exception502 path; there is no response body and no evidence that OpenAI itself returned502. `live-openai-p-recovery` completed the released four-call retry with unchanged public consumer. Missing original response remains missing.
- `live-gemini-n`: request008 hit the same relay transport path after successful write and revision4 unload. `live-gemini-n-recovery` subsequently completed three management calls, then received an empty STOP response without required `candidatesTokenCount`; compliant refusal retained. Final `live-gemini-n-management` completed no-op/unavailable operations within N16. No inferred zero usage or parser relaxation.
- Browser driver timeouts remain in originals: Anthropic first failed turn; OpenAI tried waiting on composer-local status for a terminal-originated turn. OpenAI terminal and browser outcome card show success; the wrong selector caused no repeated model prompt.

## Replay, redaction and identities

The strict success verifier remains unchanged for original launches. Successful receipts reconstruct76 request bodies across `derived-anthropic`, `derived-stream-success`, `derived-gemini-final`, and `derived-openai-public-recovery`. `derived-failed-attempts/audit.json` separately validates/reconstructs13 requests from the three named OpenAI/Gemini failures, preserving exit1 and explicit missing response lists. Its immutable helper1939c42 validates all selected original sets before replay; passing preparation and a changed last-run response identity refusal are retained in `failed-audit-controls.json`. The remaining4 Anthropic public requests have independent audit7dc9cc1. Thus all93 attempted generation bodies have exact reconstruction evidence, **not93 successful responses**. Accepted durable usage agrees with normalized raw usage; missing/refused responses are not asserted to have zero billed cost.

`catalog-denied-replay.json` retains an actual macOS sandbox profile and executable hash: catalog reads from every retained live catalog and the shipped source catalog are denied, as is all networking. `/bin/cat` fails with Operation not permitted; one actual final N request per provider still reconstructs byte-exactly. No catalog or original was moved or modified.

Separate original-load-result redaction outputs contain `[redacted]` and exactly one unchanged edit manual. Original reconstruction and end-of-log redacted projection are distinct fields; public helper checks equal usage and reports zero new model requests. The three redaction receipts live under `derived-anthropic`, `derived-stream-success`, and `derived-gemini-final`.

Original `review-binding.json` binds c0e3171 and original executable paths. `stream-binding.json` binds75a7255/new distinct binaries; all old files remain runnable. `recovery-binding.json` binds3da764b's isolated support-recovery to the same134 runtime delivery hashes and nine executable hashes as stream-binding. Both original and copied recovery identity controls passed, including last-run refusal before replay/derived output. `initial-binding.json` remains unchanged as historical failed preparation evidence.

## Display/audio observation limits and initial teaching experience

I inspected actual screenshots. OpenAI browser12 shows revision4, review/search roots, no write_file, successful terminal-originated outcome and the unload call/result in Actions. Gemini browser4 shows revision1/edit and the actual15-byte write beside expanded manual text; keyboard-gemini-live shows revision4 and the retained manual expanded through Enter after retirement. Anthropic's earlier screenshots preserve both its failed proposal and recovered current state. Scrollable panes mean a screenshot shows a viewport; full body text receipts retain the complete card.

All native audio is original float32 stereo48kHz Chrome-process capture, with silent pre-action samples and nonzero post-action waveform. Full2455-character text delivery is independently recorded and matches material exactly. This environment did not provide listening support when audio was submitted, so **I do not claim intelligibility, a complete heard utterance, Bill participation, or model hearing**. Eight-second captures establish actual bounded native audio, not the entirety of the long narration. Click expansion occurred with every provider; live keyboard expansion is a separate zero-model Gemini-GUI receipt for the shared component, not three invented keyboard runs.

The chapter's owner plan, immutable activation/manual distinction, public API contract and provider-specific chronological fixtures were useful during implementation and actual use. The initial three teaching questions and whole-record bound omission are preserved in student-review.md; the author has responded and I have confirmed each published clarification. Import re-encoding and empty-delta bugs were my implementation defects under that teaching, not permissions to change semantics. The explicit zero-byte examples now make the subtle fallback much easier to test.

The largest live difficulty was keeping task completion separate from correct authority: a provider may consume its policy on unnecessary management calls, refuse to propose a revoked tool, return incomplete usage, or fail transport. A successful local test cannot predict those outcomes. `/skills`, short management receipts, exact captured requests and real file hashes made this distinction inspectable. I would teach one benign empty-object streamed tool fixture and require safe transport stage/class diagnostics in evidence support before live launch. These suggestions do not add runtime contract requirements or justify hiding failed attempts. No old/future chapter, historical answer, grader implementation, worker log or root conversation was consulted. Independent historical comparison remains the next coordinator-owned review gate.

### Keyboard receipt linkage limitation found before freeze

Independent message `keyboard-launch-binding` identified that keyboard.json hashes the live launch before final GUI sealing; it does not match the finalized launch.json. The exact sampled launch bytes were **not retained**. The original receipt remains unchanged. [keyboard-linkage.json](keyboard-linkage.json) records the distinct hashes and independently matching immutable source/binding, Node/Chrome identities, GUI URL, within-run timestamp and retained-manual identity. These support the actual Enter/expansion observation, but do not establish an exact historical launch-byte comparison or a minimal finalization diff. Full finalized launch binding for this supplemental keyboard action is therefore limited. Future capture should preserve the sampled launch bytes as well as their hash; no repeated paid call is warranted. The three main GUI receipts remain completely sealed and pass their strict verifier.

The author’s direct response5063f35 was read in full and its manifest verified before freeze. Its interim findings accurately distinguish support shortcomings, implementation repair, partial model outcomes and the published missing-usage refusal. The final matrix above supplies the later recovery outcomes without retroactively relabeling earlier failures.
