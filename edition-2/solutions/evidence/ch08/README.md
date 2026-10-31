# Chapter 8 evidence

Student source is `../../main/`; review and exceptions are in
`../../../../docs/edition-2-notes/ch08.md`. `checks.txt` summarizes commands actually
run. The original GUI submission's raw grading output is `grader-gui.txt`; it is
not the headless core's parity score.

`browser-check.mjs` and `speech-check.mjs` are the small launchers used for the
headed Chrome checks. They use faithful external HTTP responses while exercising
the actual built Go server, Agent, native WebSocket and native speech service.
The speech wrappers only record calls/events and then invoke the real API.
`browser-initial.json` preserves the first cancelled auto-speech observation;
`browser.json` is the corrected final two-tab check. `speech.json` records the
separate real error-continuation exercise. These are API observations, not a
claim that a human listened to the audio.

To reproduce on Bill's machine, install Playwright1.64.0 under
`/tmp/ensemble-ch08-browser`, copy the scripts there as `browser.mjs` and
`speech.mjs`, and build `main/gui/cmd` as
`/tmp/ensemble-ch08-browser/ensemble-gui`. Run each script with Node. The scripts
use the installed Google Chrome application, temporary fake-provider workspaces,
ports19388/19390, and repository paths visible in their headers. They require no
live credential and remove their child servers and scratch directories.

Provider live transcripts separately retain actual model content, real shell
results, browser observations and the simultaneous terminal hint. Credentials
were read only into memory and the intended server environment. The launchers,
large temporary installation and binaries were removed after evidence capture.

Initial mutation findings, including the cleanup timeout and survivor, remain in
`mutations.txt`. `mutations-corrected.txt` contains the four corrected checks;
`window-mutation.txt` tests fixed replay bounds. `browser-mutations.txt` contains
safe-rendering, accessible-full-value and speech-error defects. These were applied
one at a time to isolated temporary source/static copies, never this repository's
grader or production tree. The correct baseline was restored and checked.

Correction round 1 adds `projection-check.mjs`: install `ws@8.18.3` alongside
Playwright and run the copied script under the same temporary Node directory.
This focused browser test uses a faithful external WebSocket peer to replay the
retained Gemini empty-text artifact and opaque examples into production browser
code. Native speech remains real. `projection.json` is the passing baseline;
`projection-mutation.json` restores the original projection fallback and checks
display, automatic speech input, and speaker input separately. The initial attempt
is retained separately because its later positive control timed out following
invalid-value cancellations. `replay-status-mutation.txt` contains the targeted
missing-`from` failure; the actual module's race suite passes.
