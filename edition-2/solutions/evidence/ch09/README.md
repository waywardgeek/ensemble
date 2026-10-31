# Chapter 9 evidence

The student review is `docs/edition-2-notes/ch09.md` in the outer repository.
`original-grade.txt` is the sole running original grader's actual output;
`grader-source.json` records its input. No final score is claimed until it ends.
`review-source.json` describes the later review source (CSS correction, comments,
README and one additional public boundary test).

`browser-check.mjs` drives installed headed Chrome with real native speech and
WebSocket APIs, the built GUI application and a faithful external HTTP peer.
`live-check.mjs` discovers an available OpenAI model and drives one bounded live
turn, terminal hint and speaker. It reads only authorized credentials in memory;
do not print the private settings file. JSON files retain selected sanitized
observations, not raw private configuration or auth headers. Native events are
not a subjective listening result. Initial browser failures are retained with
explanations in the student review.

To reproduce, install `playwright@1.64.0` under `/tmp/ensemble-ch09-browser`, build
`solutions/edition-2/main/gui/cmd` into that directory as `ensemble-gui`, copy a
script there, and run it with Node. The scripts use installed Google Chrome,
ports 19391/19392 and disposable workspaces. `CH09_WEB` and `CH09_EVIDENCE`
redirect the deterministic browser check for an isolated static-source mutation.
The wrappers record speech API calls/events and invoke the actual native API.
They do not replace speech. No Node dependency belongs to the shipped GUI.

Mutation logs preserve each defect and its intended behavioral failure. Scratch
source/static copies are removed after testing. Unchanged Chapter 8 evidence
continues to cover earlier behavior; no repeated paid provider matrix was needed.

Correction round1 adds `speed-check.mjs` (install `ws@8.18.3` alongside Playwright).
It uses port19393, saves/restores speed1.75 and checks native speech at that value
and1.2345. The isolated restored defect is exactly HTML `step="any"` changed to
`step="0.1"`. Baseline and actual wrong native rates are retained in
`speed-r1.json` and `speed-r1-mutation.json`. `browser-r1.json` retains the broader
passing follow-up; original browser results remain untouched. The current generic
browser script uses End=2; its original0.6 version is retained in progress commit
89d17e0. `review-source-r1.json` records the source frozen for reassessment.

Final original result collected October10: **80/100 FAIL**, all four settings
checks passing; only inherited parity fails. `grader-result.json` gives the exact
raw-output digest, collection date and approximate duration. Earlier pending
entries above are checkpoint history. The final student review explains each
nested failure and requests independent acceptance of inherited exceptions;
failing scores remain failing. No post-R1 grader repeat is claimed.
