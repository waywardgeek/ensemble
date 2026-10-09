# Chapter10 initial live handback — independent review pending

The reviewed schedule ran on October8 PDT / October9 UTC2026. All scheduled real
turns completed on all three vendors with **33/66 generation attempts and3/3
discovery requests**. No retry, repeated prompt, pagination, fallback, capability
probe or unplanned paid correction occurred. This freezes initial experience;
it is not chapter acceptance or historical comparison.

Runtime/build: `57d4aac3d26edcb78dc8d9fd27d7d0e1cce0ebf4`.
Interpreted reviewed support: `9822b2b63257724001d7af195b377483baad3531`.
Binding: [provider-prep-binding-final.json](provider-prep-binding-final.json),
unchanged and checked before discovery, each launch, sealing and derived work.
No runtime/support edit, Go command, new binary, compiler reservation, tag or push.
New helpers below only inspect/freeze evidence and have this handback's commit
identity; they do not relabel the runtime/support or old receipts.

| Actual binary | SHA256 |
|---|---|
| CLI | d43e5b309facf02416a14a1ab390dc44a48eac1749eddc565f5f0196619e62e5 |
| GUI | 57257823725494b87f19ba1ad3de824c671bcea73a2da37dc527af9d1eb369be |
| public consumer | 7487d1e857310a968658b3da49c4a0c1ee3cab26b1c71b8c5268c9d88ab701a2 |

| Vendor / discovered selection | A /6 | B /4 | C /6 | D /6 | Total /22 |
|---|---:|---:|---:|---:|---:|
| Anthropic / claude-sonnet-4-6 | 3 | 1 | 3 | 4 | 11 |
| OpenAI / gpt-4.1 | 3 | 1 | 3 | 4 | 11 |
| Gemini / models/gemini-3.8-flash | 3 | 1 | 3 | 4 | 11 |

Discovery was one successful200 per vendor, no next pages; Gemini advertised
generateContent. Actual OpenAI response provenance was gpt-4.1-2025-04-14.
All33 generation transports completed200 within120s; each row stayed inside its
600s durable deadline. max output4096 was unchanged. Count journals spend before
transport. The selected credential fields were loaded only by the reviewed adapter;
children received a dummy loopback key. No credential headers were retained.

| Observable | Anthropic | OpenAI | Gemini |
|---|---|---|---|
| Actual human PTY create, local session/checkpoint/history commands, restart recall | observed | observed | observed |
| Exact marker, stable SessionID, no startup HTTP or tool reexecution on recall | observed | observed | observed |
| Headed GUI: applied checkpoint before saved ack; historical tools lack live owner | observed | observed | observed; extra empty cards noted |
| Current policy1 vs old raw0/effective16, real final tool result and round_limit | observed | observed | observed |
| Font16→18 and policy persistence across actual server restart | observed | observed | observed |
| Socket reconnect, Page close/reopen, second tab close, terminal EOF then checkpoint | observed | observed | observed |
| No restored native/service speech admission, queue/current work, pause or provisional content | measured | measured | measured |
| Public two-Agent isolation, independent completions and snapshot-only continuation | observed | observed | observed |
| Immutable origin, pre-origin history_unavailable, usage restored once | observed | observed | observed |
| Real older checkpoint plus newer tail / latest / null-state offline equivalence | equal | equal | equal |
| Public replay vs actual original request bodies (five sends/API) | byte-exact | byte-exact | byte-exact |
| Retired/reloaded Skills, exact manuals/coordinates, unchanged relocated catalog | observed | observed | observed |
| Real next tool attempt consumes pending empty-pattern/17-byte limits once before dispatch | observed | observed | observed |

The coder drove actual PTYs and headed browser controls, observing responses
before follow-up inputs. These are not Bill's actions. Speech claims are measured
admission/ownership, not human hearing. C uses the compiled separate public module,
not internal function calls. Branch replay is explicitly offline with endpoints
disabled and is not additional real-provider traffic.

D's setting preparation is **six local synthetic exchanges total**, two/vendor,
kept in `*-D-seed/public/seed-*-{request,response}.json`. They contribute2 input and2
output tokens/vendor to the cumulative nominal model ledger but zero real-provider
usage or request spending. [observations.json](live-20261008/observations.json)
retains that distinction and the per-model cumulative ledger. D2's actual file
contents and once-only consumption were checked independently of model prose.

Evidence is under [live-20261008](live-20261008): discovery receipts and unmodified
bodies; selection.json; each run's launch/exit, original terminal or public records,
transport journal and exact request/response bytes; browser socket/actions/DOM/
screenshots and native/service probes; closed-store copies; immutable original
manifests and separate derived findings. Provider-root attempts.jsonl is durable.
Local offline branches and reconstructed bodies remain in separate `*-offline`
runs. The 30 launches comprise A1/A2/B/B-restart/C/D-prepare/D1/D-seed/D2/offline
for each vendor. Original launch records name both source revisions. Discovery and
other subordinate files inherit those identities through selection/binding and
sealed manifests, rather than rewriting original receipts.

Commands and failures are retained in commands.jsonl and named .out files. New
`live-command.py` records command start/end; `live-observations.py` verifies actual
observations after preflight; `live-credential-audit.py` scans only phase6 files,
named owned prose/helpers and scoped staged Git bytes, emitting aggregate results.
The first observation-summary attempt failed with KeyError(state); its corrected
reader handles the original lifecycle before/after/closed shapes. No receipt was
changed to pass. A read-only exploratory events-log read also hit its version
header's absent type field before any write. No runtime failure follows from either.

Original sealing also refused a reused `gemini-B-restart-seal.out` label from
GUI attachment setup. `freeze-label-collision.json` records that observed naming
error; only unfinished seals resumed with distinct labels. The exclusive-write
safeguard preserved the earlier receipt.

All30 original launch seals and30 source-bound verifier runs completed successfully.
The first safe credential file audit passed; the final staged audit is retained
as credential-audit-staged.json. Owned client/fixture processes exited; compact
original evidence and unchanged binaries remain, with no cache/source cleanup.

Engine-reported real usage, excluding synthetic seeds and avoiding repeated
snapshot totals (not a dollar-cost estimate):

| Vendor | Input | Cache write | Cache read | Output |
|---|---:|---:|---:|---:|
| anthropic | 26593 | 0 | 0 | 1002 |
| openai | 5452 | 0 | 6528 | 311 |
| gemini | 18182 | 0 | 0 | 895 |

Concrete findings for coordinator/author review:

- Gemini rendered extra empty “Answer / Accepted” cards next to the real retained
  answer. Full DOM and public completion prove text remains present; empty Parts
  carrying opaque provider metadata suggest the display cause. Review a focused
  presentation correction while preserving replay metadata. No correction or
  paid rerun was performed in this pass.
- Anthropic D1 additionally proposed two invalid activation-like skill names.
  Paired controlled management refusals preserved Skills and continuation; the
  requested write succeeded. D2's prose conflated report size with file size;
  actual bytes, rather than its explanation, establish the effect.
- Combined GUI shutdown is clearer when terminal EOF precedes server SIGTERM.
  The first Anthropic restart recorder waited for EOF after server termination;
  this sequence and eventual exit0 remain preserved.

No new unresolved ownership or persistence-format teaching conflict was found.
The durable [student review](student-review.md) records the actual initial
experience, help from the authorized inbox, reads and these limitations before
comparative feedback. Independent live review, any scoped corrections and their
revalidation, historical comparison, author prose confirmation and immutable
chapter release remain later gates. The compiler was never acquired in this pass.
