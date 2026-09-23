# derive-ch11 — fact sheet for finishing Chapter 11

Author working notes (2026-09-22). Sources: internal/grade/ch11_checks.go,
internal/grade/ch11_harness.go (558 lines), book/chapter-11.md (105 lines).

## B. Grader map (ch11_checks.go, 7 checks = 100)

| id | pts | what the harness actually tests |
|---|---|---|
| deterministic-context | 20 | save file has `context.dialogue[]` with >=3 `actor:"human"` and >=3 `actor:"agent"` entries, `log` array >=6 events; `BIN verify --save PATH` exits 0 and output contains `MATCH` |
| save-config | 15 | `config.model`, `config.vendor`, `config.system_prompt` non-empty; `config.tools[].name` includes `read_file` and `think` |
| save-roundtrip | 10 | parse -> marshal -> parse -> compare. **Any valid JSON passes. Decorative.** |
| resume-continues | 20 | `--load PATH`, one prompt "Do you remember me?"; last vendor request body contains ("Hello" or "meaning of life") AND "Do you remember me" |
| log-not-needed | 15 | save file with `"log": []`, `--load`; last request contains "Hello" |
| partial-replay | 5 | title says "monotonic sequence ordering": `log[].seq` strictly increasing, >=4 events. **Does NOT test checkpoint + partial replay.** |
| ch10-parity | 15 | ch10 checks still pass |

## Harness contract (the fixture)

- Launch: `BIN --port P --gui-dir DIR --save PATH` (phase 1) / `--load PATH` (resume).
- Env: LLM_BASE_URL, LLM_MODEL=fake-model, LLM_VENDOR=anthropic, LLM_API_KEY,
  EN_SKILLS_DIR, EN_PRIMARY_SKILL=base, CH02_LOG.
- Waits for HTTP on `/` (GUI server must start).
- Prompts on stdin as `{"kind":"prompt","text":...}` lines, 500ms apart;
  **save happens when stdin closes** (EOF), process must exit within 10s.
- Save file JSON: top-level `context` (object with `turn`, `dialogue[]{seq, actor}`),
  `log` (**a JSON ARRAY of events**, each with `seq`), `config` object.
- Verify: `BIN verify --save PATH`, success = exit 0 + stdout/stderr contains `MATCH`.
- Skill fixture: primary skill `base`, tools `read_file think`, body
  "You are a helpful assistant."

## C. Stale text in chapter-11.md (fixture wins)

1. §11.6 prints `./agent verify SAVE_FILE` and "prints OK". Grader runs
   `verify --save PATH` and requires `MATCH`.
2. §11.4 says "The grader tests this by saving mid-conversation, continuing,
   then comparing all three values." FALSE: partial-replay only checks seq
   monotonicity.
3. §11.5 "verify the vendor receives a well-formed request with the full
   conversation history" — overstated; grader checks substring "Hello".
4. §11.2 `Log *Log` field: need to confirm it marshals as a bare array
   (grader reads `log` as `[]json.RawMessage`).
5. No TL;DR, no plain-words section, no motivational opener, no
   "Taking it for a spin". §11.7 "Looking Ahead" is a forward preview
   (voice rule: 0 previews) — delete.

## F. Grader weaknesses (coder list; P9 audit)

- save-roundtrip is vacuous (10 pts for "is valid JSON").
- deterministic-context trusts the student's own `verify` output; a
  `verify` that prints MATCH unconditionally scores 20. Grader never
  rebuilds anything itself.
- partial-replay (the chapter's invariant 2) is ungraded.
- The chapter promises all three: strengthening is permitted by P9.

## Ruling 2026-09-22 (Bill): load by default, no flag

TL;DR rewritten on the anchor design (book/chapter-11.md). The TL;DR's
check table is the contract the grader coder must implement:

| check | pts | construction |
|---|---|---|
| save-shape | 15 | config fields; `as_of` == last log seq; log seq strictly increasing |
| default-load | 20 | two runs, same temp cwd, no flags; run 2's first request contains run-1 prompts |
| replay-equals-snapshot | 20 | load save as-is vs same save with `context:null`; next request bodies byte-identical |
| tail-applied-once | 15 | splice {context, as_of} of the turn-2 save onto the log of the turn-3 save; next request == one from loading turn-3 save. Grader never parses events: splice only. |
| log-not-needed | 10 | `log:[]`; request carries prior prompts; saved new events have seq > as_of |
| bad-save-refused | 5 | garbage save.json: non-zero exit, file bytes unchanged |
| ch10-parity | 15 | unchanged |

Replaces: save-roundtrip (vacuous), deterministic-context (trusted the
student's own verify), partial-replay (monotonic only), resume-continues
(folded into default-load). `verify` leaves the contract ("Yours").

Coder cautions:
- Default load writes/reads ./save.json: confirm EVERY earlier chapter's
  harness runs in a per-run temp cwd (ch10/ch11 checked: yes), or parity
  runs will load each other's conversations.
- Rule 7 (config is a record, running config wins): confirm the
  reference obeys it; if it restores config from the file, fix the code.
- Two full-replay functions exist (Log.Replay event_log.go:62 and Rebuild
  save.go); keep one.
- Byte-identical requests assume rendering carries no time/random field;
  verify on the reference before trusting the check.
