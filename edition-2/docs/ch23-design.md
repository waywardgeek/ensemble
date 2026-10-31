# Chapter 23 — The Sandbox: implementation design

Working notes for the coder. Contract is `book/chapter-23.md`; task list is
`book/brief-ch23-coder.md`. Status lives at the bottom.

## Threat model (Bill's clarification, 2026-10-08)

The asset protected is **the context window**, NOT the binary.

The ensemble binary has trivial access to credentials and must have it: it
puts an `Authorization` header on every request. Nothing is gained by
pretending the binary is contained, and claiming so would be exactly the
theater §23.8 attacks. The LLM is what must never see a credential, because
the context window is the only surface prompt injection can reach, and
anything in context can be repeated, re-encoded, or exfiltrated later.

A credential can reach the context window by exactly three paths. Each
mechanism in the brief closes exactly one:

| Path into context | Mechanism |
|---|---|
| Binary writes `api.log` inside sandbox, model calls `read_file` on it | `sanitizeForLog`, applied **at write time** |
| Key sits in env, model runs `env` or `curl -H "...$OPENAI_API_KEY"`, output returns as a tool result | `sanitizedEnv` |
| Credential returned *as* a tool result (the Oct 2025 failure) | `send_secret` pattern: write to stdin, return "credential provided" |

Sanitize-at-write-time is load-bearing: a file that briefly held the real key
is a file the model may have read before the scrub.

## Decisions

### D1. Sandbox root travels on `common.Call`, not on `common.Agent`

Every file tool is `func(_ *common.Call, args json.RawMessage)` — they
currently discard the Call. `Reg` is per-agent but tools are plain funcs in a
map, so the root must ride the per-call context.

- CHOSEN: add `Sandbox string` to `common.Call`. Precedent is `Limits`
  (per-call policy as data) and ch22's `Engine` back-pointer.
- REJECTED: a `SandboxRoot()` method on `common.Agent`. Ch22 rejected
  `Settings()` on that interface for the same reason — it forces a
  meaningless method onto `*Logger` and two fakes. `*Logger` satisfying
  `common.Agent` is a ch22 exhibit; do not break it.

### D2. `AgentSpec` comment framed as collaborator vs. policy data

`AgentSpec` carries: "A new CAPABILITY must never be added here -- that is
the dependency bag this type replaced."

The brief says to justify the new fields as "paths, not capabilities." That
does not survive two Booleans. The real ch22 distinction is that the spec
must not carry **collaborators** (engine, hub.Model, closures — things the
agent *uses*); it may carry **policy data** describing what the agent is
permitted to reach. `SandboxRoot`/`SafeMode`/`EnableWebSearch` are policy
data. Write the comment that way. FLAG FOR AUTHOR.

### D3. `EnableWebSearch` gates the SKILL, not builtin tools

MEASURED: there are no `web_search`/`crawl_web`/`search_web` tools in the
agent. Zero grep hits. Ch21 delivered web access as the `web-search` skill
(`agent/skills/web-search/SKILL.md`) wired to hosted Firecrawl MCP.

So the brief's "web search and crawl tools are not registered" cannot be
implemented as written. Equivalent at one level up: when `EnableWebSearch` is
false the `web-search` skill is not discoverable/loadable, so its MCP never
connects and its tools never bridge in. Same "absence, not instruction"
principle. MUST FIX candidate for the chapter prose. FLAG FOR AUTHOR.

### D4. Safe mode removes tools from the registry after construction

`NewRegistry()` installs all builtins. Safe mode calls `RemoveTool` for
`run_command`, `send_input`, `kill_job`, `jobs`/`wait_for_job`. Removal (not
a declaration filter) because `RemoveTool` already drops the tool from
`tools`, `meta` and `argSpec` together, so `Declarations()`, `Lookup()` and
the `$TOOLS` prompt variable all agree. A fabricated call then hits Lookup's
existing "no such tool" error. NOTE: the exec-ish tool set is
`run_command`, `send_input`, `kill_job`, `wait_for_job` — brief says "jobs",
which is not a tool name here; `wait_for_job` is the real one.

## Build order

1. `internal/tools/sandbox.go` + `sandbox_test.go` — the function and its
   attack suite. Everything depends on this.
2. `common.Call.Sandbox`; thread through `ToolReadFile`, `ToolWriteFile`,
   `ToolEditFile`, `ToolListDirectory`, `ToolSearchFiles`, and
   `toolRunCommand`'s `cwd`.
3. Command confinement: `cmd.Dir`, `sanitizedEnv`.
4. `sanitizeForLog` in the logger, at write time.
5. `validateChildSpec`.
6. CLI flags `--sandbox`, `--safe-mode`, `--yolo`.
7. Grader `internal/grade/ch23_*` (7 checks / 100 pts), mutants script,
   Makefile `grade23`/`grade23-audit`, gradesweep entry.
8. Snapshot `solutions/ch23` via `git archive`; full cross-chapter sweep.
9. `book/chapter-23-coder-review.md`.

## Positive controls (non-negotiable)

Per §23.3: "A sandbox that rejects everything is trivial to build and
worthless to use." Every escape test gets a matching test proving the
legitimate operation inside the sandbox still succeeds. Same for the grader.

## Status

- [x] Survey; threat model confirmed with Bill
- [ ] 1 sandbox.go
- [ ] 2 thread paths
- [ ] 3 command confinement
- [ ] 4 log sanitization
- [ ] 5 validateChildSpec
- [ ] 6 CLI flags
- [ ] 7 grader + mutants
- [ ] 8 snapshot + sweep
- [ ] 9 coder review
