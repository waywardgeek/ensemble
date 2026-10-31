# ch27 "Gateway" — CODER working notes

Role: I am CODER. Bill is author. Contract: `docs/ensemble-ch27-tldr.md`
(7 checks / 100 points / `make grade27`). Design: `docs/ch27-design.md`.

## Bill's rulings this session (supersede the design doc)

1. **Noise-KK is copied, not linked.** Source: `~/projects/gateway/pkg/noise/noise_kk.go`
   (166 lines) — byte-identical to `/Users/bill/projects/coderhapsody/gateway/pkg/noise/noise_kk.go`.
   The design doc said "adapt from OpenADP"; that is SUPERSEDED. Do not vendor
   either repo as a module dependency; copy the file into `agent/gateway/noise.go`.
2. **"Boilerplate, not novel, but easy to get wrong."** Reproduce the proven
   wire format exactly rather than inventing one.

## Proven wire format (from ~/projects/gateway/pkg/agent/manager.go:177-240)

- Handshake msg1 (agent->gateway), raw WS **binary**, NOT encrypted:
  `[1-byte agent-id length][agent-id bytes][noise-kk msg1]`
  The length prefix exists because the responder must look up the initiator's
  static public key BEFORE it can build the KK session.
- Handshake msg2 (gateway->agent): bare noise msg2, raw WS binary.
- Handshake complete after exactly 2 messages; assert `IsHandshakeComplete()`.
- Prologue: `[]byte("gateway-v1")`. Both sides must match or handshake fails.
- All later frames: `Encrypt(json.Marshal(Frame{Type,Payload}))` as WS binary.
- Keys are **X25519** (`noise.DH25519`), despite the config label `ed25519:`.

## Repo conventions (verified)

- Graders: `internal/grade/chNN_checks.go` + `chNN_harness.go` (repo root module).
- Dispatch: `cmd/grade/main.go` `case NN:`; Makefile target `gradeNN`.
- Mutants: `scripts/chNN-mutants.sh`. Sweep: `./scripts/gradesweep.sh`
  (NOT `make gradesweep`) — ch27 must be ADDED to its list.
- Frozen snapshot: `solutions/ch27`.
- `agent/` is its own Go module; `gorilla/websocket` already a dep;
  **`flynn/noise` must be added**.

## TERRITORY WARNING

Astra works this same repo concurrently on Edition 2. Uncommitted files of his
were present at session start (`book/chapters.md`, `cmd/grade-edition2/main.go`,
`internal/gradeedition2/ch07_driver.go`). **Never `git add -A`.** Stage exact
paths only. `book/edition-2/` and `solutions/edition-2/` are his.

## Stage plan (commit each green stage)

| # | Stage | Deliverable |
|---|-------|-------------|
| 0 | deps: add flynn/noise; package skeleton | — |
| 1 | `agent/gateway/noise.go` (copied) + keygen + handshake test | D1 (15) |
| 2 | `agent/gateway/envelope.go`, `protocol.go`: frames, trust, provenance | D2 (10) |
| 3 | `agent/internal/gateway/config.go`, `groups.go` | — |
| 4 | `server.go` + `connections.go`: WS accept, handshake, conn state | D3 (15) |
| 5 | `router.go`: group routing + contactability, reject unauthorized | D4 (15) |
| 6 | `queue.go` + `persistence.go`: ack, offline queue, resume | D5 (15) |
| 7 | `agent/gateway/client.go` + `handler.go`: Handler iface, recv goroutine | D6 (10) |
| 8 | MCP admin tools; RegisterChild + narrow-only grant | D7 (20) |
| 9 | `cmd/gateway/main.go` composition root | D3 |
| 10 | `cmd/interviewer/main.go` | exercise |
| 11 | grader, mutation audit, snapshot, sweep, coder review | — |

## Hard-won rules to not relearn

- A library test proves the library; **only the binary proves the product**
  (ch25: 5 bugs lived behind passing unit tests because each test wired by
  hand the thing the composition root forgot). Drive real binaries.
- **No package-level mutable vars** — cost ch5 ten points twice. Counters go
  in a struct under the lock that already protects it.
- Every absence check needs a paired positive control.
- Never edit the tree while a sweep runs.
- `git worktree prune` before reusing a worktree path; never discard stderr.

## Status

- [x] stage 1 `4e517504` Noise-KK copied + keys (13 tests)
- [x] stage 2 `4cc3cc62` envelope, trust, protocol, UUIDv7 (18 tests)
- [x] stage 3 `132db75f` config, registry, group admin, persistence (13 tests)
- [x] stages 4-7 `2fe18c9f` queue, connections, server, MCP, client library
      (12 integration tests, all green under -race)
- [x] stages 9-10 `e3371922` cmd/gateway + cmd/interviewer
- [ ] stage 11 grader, mutation audit, snapshot, sweep, coder review

## VERIFIED facts (measured, not assumed)

- `go list -deps ./gateway/` shows ZERO `ensemble/agent/internal` imports and
  no ensemble deps at all besides itself. D6's "structurally separable" claim
  is real, not aspirational.
- A program in a SEPARATE Go module (`/tmp/ch27live`, replace directive)
  compiles and runs against the public client library. That is the honest
  proof for D6, and is the shape the grader should use.
- Live three-binary run: gateway + subject (external module) + interviewer.
  4 questions routed and answered, `interviews/subject.md` written.
- SIGTERM persisted `registry.json` (both contactability declarations) and
  `queue.json` (`next_seq` 4 per agent, zero pending because all acked).

## Live rig (reproducible)

    /tmp/ch27live/gateway.json      config, listen 127.0.0.1:18093
    /tmp/ch27bin/{gateway,interviewer,subject}
    GWKEY=$(/tmp/ch27bin/gateway -config gateway.json -print-key)

`-print-key` on any binary loads the existing key file rather than making a
new one, so never hand-copy a key between commands.

## Design notes worth keeping

- Handshake msg1 carries the agent id IN THE CLEAR because KK cannot build a
  session without knowing which static key to expect. The id is a claim; the
  handshake is the proof. Said plainly in the code comment.
- Replies are authorized by origin, not by contactability. The sender chose
  to start the exchange. Reply trust is still computed by Authorize and falls
  back to external, so a reply is always delivered but never over-labeled.
- Provenance is added at DELIVERY, not at acceptance, so a redelivered
  message cannot accumulate two headers.
- Queue before deliver, never the reverse, or a message written to a socket
  the recipient never read is lost with nothing left to redeliver.
- Runtime groups are a separate namespace from config groups; config groups
  are immutable at runtime because a runtime edit would revert on restart.

