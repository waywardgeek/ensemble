# Chapter 21 — Coder Review

*From the coder to the author. Written during implementation, 6 October 2026.
Every number here was measured today on this machine; claims I did not measure
are marked INFERRED. Nothing in this file is prose for the book — it is what I
learned building the thing.*

---

## 1. The headline: the chapter's stack changed

The outline specifies `ddgs` for search and `crawl4ai` for fetch, both as local
MCP sidecars. I measured both. **Neither survived.** What shipped instead is
Firecrawl's hosted MCP server, reached over a new URL transport.

The decision was made in three steps, each forced by a measurement rather than
a preference:

1. **crawl4ai has no MCP server to use.** (§2 below.)
2. **ddgs search is 73% reliable.** (§3 below.) A flaky search is a failing
   search — for us and for the student.
3. **Firecrawl's hosted server is keyless, stateless, and was 15/15.** (§4.)

The net effect on the chapter is *less* new code than the outline assumed, not
more, and a better teaching beat. See §6.

---

## 2. Factual corrections to the outline and research doc

These are errors of fact, not matters of taste. Please fix them wherever they
appear.

**2.1 "crawl4ai … Built-in MCP server" is false.** (§21.3, TL;DR, Decision 1,
and the coder brief all state it.) VERIFIED: the pip package's console scripts
are `crawl4ai-doctor`, `crawl4ai-download-models`, `crawl4ai-migrate`,
`crawl4ai-setup`, and `crwl`. There is no MCP entry point and no `*mcp*` module
in the installed package. crawl4ai's MCP support exists only in its **Docker
deployment**, as an HTTP/SSE endpoint. A reader who runs `pip install crawl4ai`
expecting a stdio MCP server will not find one.

**2.2 ddgs's own MCP server already does fetching.** The outline treats search
and fetch as requiring two different libraries. VERIFIED: `ddgs mcp` advertises
six tools — `search_text`, `search_images`, `search_news`, `search_videos`,
`search_books`, and **`extract_content`** (fetch a URL, return its text). Had we
stayed with ddgs, crawl4ai would have been unnecessary on its own terms.

**2.3 ddgs version confirmed.** The outline's "v9.16.0" is correct — that is
what installed today. VERIFIED.

**2.4 The brief's "grader must NOT require internet access" is now overruled**
by the author's ruling that the chapter requires internet and that a reader who
does not need web access may skip the chapter. Recording the conflict so the
brief and outline do not disagree in the final text.

**2.5 The brief's "tools are always available / MCP servers start when the
agent starts" fights the architecture.** See §5.1 — it was withdrawn in favour
of a loadable skill.

---

## 3. Why ddgs was dropped: the measurements

All figures from a single machine, one session, through `ddgs mcp` over stdio.

**Search reliability — 15 distinct queries, one attempt each, 1s apart:**

| Result | Count |
|---|---|
| OK | 11 |
| FAIL | 4 |

**73% per-call success, average 2.33s.** The failures are not query-specific:
re-running the four failed queries four times each rescued **all four**
(per-attempt success across that retry run was 10 OK / 6 FAIL). So the fault is
transient throttling, and the fix is retry with backoff.

That fix is the problem. The retry would have to live *inside the sidecar*, and
the sidecar is third-party. Making ddgs reliable therefore means **writing our
own MCP server** wrapping the ddgs Python library — reintroducing exactly the
authored-server complexity that dropping crawl4ai removed.

**Backends are not interchangeable.** VERIFIED: with `backend` set explicitly,
`brave`, `mojeek`, and `wikipedia` each failed every attempt; the default
(duckduckgo) worked. The outline's "10 backends including Google, Bing, Brave"
should not be read as ten working options.

**Fetch was fine.** `extract_content` was 5/5, 0.10–0.29s, up to 112KB
(Wikipedia). Fetch was never the weak half.

**But fetch has no main-content extraction.** VERIFIED on `react.dev/learn`:
23,482 chars as `text_markdown`, 16,991 as `text_plain` — both dominated by
nav, sidebar, and link lists before any article text. This matters because
§21.5's "tool results can be large" is real: one page can be tens of thousands
of characters of mostly chrome.

**A gotcha worth a sentence of prose:** `fmt` accepts `text_markdown`
(default), `text_plain`, `text_rich`, `text` (raw HTML), and `content` (raw
bytes). It does **not** accept `markdown`. I passed `markdown` on my first
probe and got output anyway, which nearly led me to a wrong conclusion about
extraction quality.

**ddgs error messages are opaque.** Every failure — rate limit, connection
refused, bad URL — returns the same 36-character string: `Error executing tool
extract_content`. The cause is destroyed inside the server's
`UnexpectedToolError` wrapper. The agent learns *that* it failed and never
*why*. This is a concrete, checkable example for §21.7: failure visibility is a
property of the server you chose, not only of the code you wrote.

---

## 4. What Firecrawl actually gives us

VERIFIED against `https://mcp.firecrawl.dev/v2/mcp`:

- **Keyless.** No account, no API key, no signup, within a daily limit. The
  chapter keeps its zero-friction on-ramp.
- **Stateless.** No `Mcp-Session-Id` is issued, and `tools/list` succeeds
  *without* calling `initialize` first. (I still implemented session support;
  see §5.4.)
- `serverInfo`: `firecrawl-fastmcp` v3.27.3, protocol `2024-11-05`.
- Tools: `firecrawl_search`, `firecrawl_scrape`, `firecrawl_parse`.

**Reliability, same method as the ddgs run:**

| | result | average latency |
|---|---|---|
| `firecrawl_search` | **10 / 10** | 0.84s |
| `firecrawl_scrape` | **5 / 5** | 0.48s |

**100% vs 73%, and 2.8× faster on search.** No sidecar process, no Python, no
Node, no install step of any kind.

**The author's "Astra" test question works.** VERIFIED: `firecrawl_search` for
"OpenAI Astra model announcement" returns
`https://openai.com/index/gpt-6-astra/` — *"GPT-6 Astra: A new generation of
intelligence"*. The model's name is **GPT-6 Astra**.

### 4.1 The most important behavioural finding: an HTTP error is not a tool error

VERIFIED: `firecrawl_scrape` on a URL returning **404** comes back with
`isError=false` and the error page's content as markdown. The status code is in
`metadata`, not in the error flag.

This deserves prose in §21.7. An agent — or a grader — that checks only
`isError` will treat a 404 page as a successful fetch and happily summarize
"Page Not Found" as though it were documentation. "Visible failure" means
reading the status, not trusting the flag.

By contrast, a genuinely unreachable target *is* flagged: fetching
`http://127.0.0.1:9/...` returns `isError=true` with a **specific** message
(`ERR_UNSAFE_PORT`), and an unknown tool name surfaces as JSON-RPC `-32601`.
Firecrawl's diagnostics are markedly better than ddgs's single opaque string —
another concrete contrast for §21.7.

---

## 5. Architectural decisions the outline did not anticipate

### 5.1 The skill is loadable, and that deletes a whole problem

The brief asks for search to be always on, with servers started at agent
startup. I began building that and found it fights the design:
`onSkillMCPConnect` has exactly **one** call site — inside the `load_skill`
tool handler (`internal/tools/tools.go:1136`). `LoadInitial` and `LoadDynamic`
both route through `registry.load()`, which never touches MCP, and
`registry.go:87` forbids loading a primary skill dynamically. So declaring
`mcp_servers` on the primary skill **parses cleanly and silently never
connects** — a trap worth one sentence of prose on its own.

The author's ruling settled it: sidecars start when a skill is loaded, because
dynamic loading is the point of skills. `web-search` is therefore a `type:
loadable` skill. Consequence: **zero changes to `tools.go`, `agent.go`, or the
skill registry.** §21.4's claim that the only new code is configuration
survives intact — which it would *not* have under the always-on reading.

### 5.2 Ch12 left the seam; ch21 only fills it

This is the best structural surprise in the chapter, and I think it should be
the spine of §21.4. The work needed to add an HTTP transport was already
designed for:

- `internal/mcp/transport.go` defines `Transport` as three methods —
  `Send`/`Recv`/`Close` over `json.RawMessage` — and its doc comment says: *"A
  future fourth (URL/port-based) can be added later — the interface is the
  seam."*
- `common.MCPServerConfig` **already had** a `URL string` field, commented
  `// for url: the server URL (future)`.
- `internal/skills/skill.go`'s frontmatter parser **already had** `case "url":`.
- `cmd/main.go`'s transport switch **already had** a `default:` that logs
  "unsupported MCP transport".

So ch21's host change is literally two lines — one `case "url":` in that
switch. Everything else is a new file implementing an interface that was
written to be implemented. That is a much stronger payoff story than
"configure a Python sidecar," and it retroactively justifies ch12's design.

### 5.3 Transport vocabulary: `url`, not `http`

I used `transport: url` because that is the codebase's existing vocabulary
(struct comment and parser case both say `url`). The implementation type is
`HTTPTransport`. Prose should use `url` when showing frontmatter.

### 5.4 Design choices inside the transport, and why

- **An unbounded cond-var queue, not a buffered channel.** HTTP has no stream
  the peer can write to whenever it likes; a reply exists only as the body of a
  request we made. So `Send` parses the reply and queues what it found, and
  `Recv` drains. A notification POST returns 202 with **no body** and enqueues
  nothing, while one POST may yield **several** frames — with a fixed capacity
  either case can wedge a concurrent `Send`.
- **Both body shapes are handled.** The spec permits `application/json` or
  `text/event-stream`, and servers switch between them per method.
- **Session support, though Firecrawl does not need it.** A stateful server
  issues `Mcp-Session-Id` on initialize and expects it echoed. Supporting it
  costs ~6 lines and makes the transport usable against servers other than the
  one we happen to ship.
- **Non-2xx errors carry the body, not just the status.** Quota and auth
  failures explain themselves in the body; dropping it leaves the agent unable
  to say why the web stopped working. There is a test asserting the body
  survives, because this is precisely the kind of detail a later refactor
  "tidies" away.

### 5.5 Credentials never touch a committed file

Worth a line of prose. `NewStdioTransport` inherits the parent environment when
`env` is empty, so a key-requiring server is configured by exporting the
variable, not by writing it into `SKILL.md`. With the keyless hosted server the
question is moot today, but the pattern matters: `solutions/` is in git, and
the repository has a pre-commit hook that refuses staged credentials.

---

## 6. What the prose should emphasize, based on building it

1. **Reliability is a feature, and it is measurable.** The 73%-vs-100%
   comparison is the most useful thing I learned. It is also a good lesson in
   method: three samples suggested ddgs was fine; fifteen showed it was not.
2. **An HTTP error is not a tool error** (§4.1). The single most likely bug in
   a reader's implementation.
3. **Error-message quality is a selection criterion.** "Visible failure" is
   half your code and half the server's manners. ddgs's one opaque string
   versus Firecrawl's `ERR_UNSAFE_PORT` makes that concrete.
4. **The seam payoff** (§5.2) — an interface written in ch12 for a transport
   that did not exist yet, implemented in ch21 without touching its consumers.
5. **The silent-no-connect trap** (§5.1): config that parses perfectly and does
   nothing is worse than config that errors.

---

## 7. An editorial problem I cannot fix, and should not

§21.1 is an indictment of vendors who monetised search and chose lock-in. The
implementation now depends on **Firecrawl, a commercial company**, on its free
tier.

I do not think the argument collapses — Firecrawl is neutral with respect to
*LLM* vendors, works with any agent, locks you into no model, and the outline
already names it as the one commercial offering that "did not choose the dark
path." But §21.3's framing of ddgs and crawl4ai as "the community answer," and
the implied contrast between community tools and commercial ones, no longer
matches what the chapter ships. The honest version is probably: the objection
is to *your model vendor* owning your search, not to paying anyone for search.

Flagging it rather than rewriting it: this is the author's argument to make.

---

## 8. Status and what remains

**Done and verified:**
- `solutions/ch21/` copied from `solutions/ch19` (ch20 was prose-only); builds
  clean.
- `solutions/ch21/internal/mcp/http.go` — the URL transport.
- `solutions/ch21/internal/mcp/http_test.go` — 9 hermetic tests, no internet,
  using `httptest`.
- `cmd/main.go` — two-line `case "url":` in the existing transport switch.
- `solutions/ch21/skills/web-search/SKILL.md` — loadable skill, keyless
  endpoint, no credentials, no sidecar.
- Live end-to-end through the real code path: parse SKILL.md → build transport
  from that parsed config → initialize → list → search → scrape → failure
  paths.
- Mutation audit of the transport: 4 mutants, **4 killed**.

**A test-quality note worth repeating in the book.** My first mutation run
appeared to hang for four and a half minutes. The mutant had not failed the
suite — it had **starved** it: with SSE frames dropped, nothing is ever queued,
`Recv` blocks by contract, and `go test` sat until its ten-minute panic timeout.
A test that hangs under mutation is nearly as useless as one that passes, since
the audit harness reads failures, not hangs. The fix was a bounded `recvWithin`
helper; the suite now runs in 0.01s and the same four mutants die in 6s total.

The same run produced a genuine survivor: deleting `cond.Broadcast()` from
`Close` changed nothing, because no test had a Recv **parked** at the moment of
close — the one case that matters, since the codec's read loop lives there for
the life of the connection. I added `TestHTTPTransportCloseWakesParkedRecv`,
which kills it.

## 9. The snapshot base was wrong, and it matters for every chapter after this

I first built `solutions/ch21` by copying `solutions/ch19`, on the reasoning
that Chapter 20 was prose-only so ch19 was the last frozen snapshot. That was
wrong, and the author caught it from a symptom I had already written down as a
finding.

The symptom: I reported that MCP is unavailable in `chat` mode, because
`SetOnSkillMCPConnect` is installed inside `runActorLoop` while `chat` called a
separate `runLoop`. The author's response was that the CLI is supposed to be
the same agent with a different front end. He was right. In the **live** tree,
`chat` is already the same actor loop the GUI drives, with this comment on it:

> Chat is the SAME actor loop the GUI drives, with a text front end attached
> instead of a browser. It used to be a bare engine with no actor, skills,
> recall or save file, which meant the terminal could not reproduce a single
> GUI-reported bug.

That was commit `4023cff`, "retire synchronous engine loop; actor is the sole
orchestration path". The frozen ch19 snapshot predates it. So my finding was an
artifact of reading a stale tree, not a fact about the architecture, and I have
struck it. **Retracted: there is no chat-mode MCP gap.**

The correction generalizes, and it is the part worth putting in the author's
notes rather than mine:

- Since ch20 is prose-only, the accumulated state through Chapter 20 lives in
  `agent/`, not in any `solutions/` directory. For a chapter that follows a
  prose-only chapter, **the live tree is the base**, and `solutions/ch(N-1)` is
  not.
- The drift was **65 files**, not the 7 recorded when ch19 was current. A note
  about how far a snapshot has drifted expires quickly; it should be measured
  at use, never recalled.
- Rebasing was strictly an improvement and cost nothing but the transport file
  being re-applied: the suite went from 10 packages to 12 (`oauth`, `recall`,
  `settings`, `cachelens` are new), and `gofmt -l` went from flagging an
  inherited unformatted `internal/ws/replay_test.go` to being silent, because
  the live tree had already fixed it.

Taking the snapshot with `git archive HEAD:agent` rather than `cp -a` is worth
keeping as the recipe: it copies exactly the tracked files, so the debug
binaries and the 17MB of runtime logs sitting in `agent/` cannot leak into a
committed snapshot.

One consequence for the prose: because `chat` is the actor loop, **web search
works from the terminal, not just the GUI**. That is worth a sentence, since
Chapter 21's value is largely that a headless agent can research.

## 10. One deliberate limitation to note

`NewHTTPTransport(url, headers)` accepts custom headers, but
`MCPServerConfig` has no field for them, so the host passes `nil`. That is
sufficient for the keyless endpoint the chapter ships and for any hosted server
that authenticates by URL. A server needing `Authorization: Bearer …` would
require a new frontmatter field. I did not add one: it is unneeded for the
chapter's stack, and an unused config field is a claim the code cannot honour.
If the author wants the chapter to show an authenticated hosted server, say so
and it is a small addition — but it reintroduces the credential-handling
problem that the keyless endpoint currently makes vanish.

## 11. The grader

Seven checks, one hundred points, scoring 100/100 against `./agent`:

| check | pts | what it protects |
|---|---|---|
| `tools-gated-by-skill` | 15 | web tools absent until the skill loads |
| `url-transport-connects` | 15 | `transport: url` reaches a hosted server |
| `search-dispatches` | 15 | results return as a tool result |
| `fetch-returns-planted-token` | 20 | page content arrives intact |
| `fetched-content-is-a-tool-result` | 15 | fetched bytes never become instructions |
| `tool-error-is-reported` | 10 | `isError` on a 200 reaches the model |
| `transport-error-is-reported` | 10 | an HTTP failure keeps its reason |

Three design points worth the author's attention, because each is a claim the
prose can make:

**The grader supplies the skill file.** It owns the endpoint and therefore the
tool names, which is what makes the scored checks deterministic without a
network — and it means a student who chose a different backend is graded on
their wiring rather than on their vendor. §21.6's "report the backend you can
characterize" survives intact.

**The fake MCP server is a real HTTP server, not a stub transport.** The
transport *is* the thing under test; a stub would bypass the code being graded
and certify nothing.

**One check is unfakeable.** The fake server plants a token minted at process
start inside one page. A stub, a cached fixture, or the model's own knowledge
cannot produce it. This is the same trick as the Chapter 16 memory check, and
it is the reason that check is worth 20 points rather than 15.

## 12. Mutation audit: 4 mutants, 4 killed exactly

`scripts/ch21-mutants.sh`, also `make grade21-audit`.

| mutant | kills |
|---|---|
| host cannot construct a URL transport | 6 checks (everything but the gate) |
| tool results truncated to 30 bytes | `search-dispatches`, `fetch-returns-planted-token`, `fetched-content-is-a-tool-result` |
| server's `isError` flag dropped | `tool-error-is-reported` |
| response body discarded on non-2xx | `transport-error-is-reported` |

The audit found a real weakness in my own check before the mutants ran, and it
is the most useful thing in this section. `tool-error-is-reported` originally
asserted that the text "404" reached the model. That check **passes** against
an agent that drops the server's `isError` flag entirely, because the error
page's text still arrives — attached to a tool result that claims it
succeeded. The model then has a page saying one thing and a frame saying
another, and it will generally believe the frame.

The check now asserts the `is_error` flag on the tool_result block, not the
words in it. The general form is worth a sentence in §21.7: **when a tool
fails, the failure is carried by the frame, not by the prose inside it.** An
agent that loses the frame produces the most expensive failure mode in the
chapter — a confident summary of an error page.

## 13. One check has no mutant, on purpose

`tools-gated-by-skill` cannot be killed by any valid mutant, and I want to be
explicit rather than quietly ship a 15-point check nothing can break.

A valid mutant deletes exactly one behaviour. There is no behaviour here to
delete: MCP servers are connected in exactly one place, the `load_skill` tool
handler, so nothing can connect a server before a skill is loaded. Breaking the
check requires *adding* a startup-connect path, which is a feature, not a
mutation.

I kept the check anyway, and the justification matters for the chapter: it is
entirely reachable for a **student**, because the coder brief proposed exactly
the design that fails it — "the tools are always available, the servers start
when the agent starts." The check is what separates that design from the one
the architecture supports. It earns its points against the submissions it will
actually see, not against a mutant of the reference.

## 14. Two harness bugs that both looked like student failures

Worth recording because both produced a plausible, wrong story about the
student's code, and the chapter is partly about exactly this confusion.

**15/100 — "the transport is broken."** The grader's own `base` skill did not
list `web-search` under `loadable-skills`, so Chapter 10's progressive
disclosure correctly refused to load it. The symptom was the web tools simply
being absent from the tool list, which is indistinguishable from a transport
that never connected. The architecture was working and the test was wrong.

**65/100 — "the second scenario's tools never dispatch."** Both scenarios
shared a working directory, and since Chapter 11 the agent loads `save.json` on
startup by default. The second scenario silently resumed the first one's
conversation, so every scripted reply landed one turn out of place. Each
scenario now gets its own temp directory.

The second one is a genuinely good illustration for the book: persistence that
is helpful in production is a hazard in a test harness, and the failure does
not look like persistence — it looks like the feature under test being broken.

## 15. A note on how the harness drives the agent

The ch18 and ch19 harnesses drive the agent over the GUI websocket. This one
uses stdin/stdout only: in server mode each `{"kind":"prompt"}` is answered
with one `{"assistant":...}` line, so turn completion is observable without
opening a socket. Fewer moving parts, and it keeps the headless path exercised
— which matters now that `chat` is the same actor loop, because the terminal
path is a real way to use this feature.

## 16. Cross-chapter sweep: 29 of 30 at 100, and the exception is not ours

`scripts/gradesweep.sh`, 12m26s. I added `run 21 ./solutions/ch21` to it, so
ch21 is now part of the standing sweep rather than something only I remember
to run.

Every target scored full marks (ch5 scores 120/120 by its own scale) with one
exception: **`ch19 ./agent` scored 85/100**, while `ch19 ./solutions/ch19`
scored 100/100.

I did not take that at face value in either direction. The failing check
reports `open agent/events.jsonl: no such file or directory`, which is a
harness path problem and has nothing to do with web search — but "looks
unrelated" is not evidence. I built a worktree at `5cd7772`, the commit
immediately before any of my `agent/` changes, confirmed `agent/skills/` there
has no `web-search` entry, and ran the ch19 grader against it. It scores the
same **85/100** with the identical error.

**Pre-existing, not a regression from this chapter.** Worth fixing, but it
belongs to ch19's harness and I have left it alone rather than widen this
chapter's diff.

The sweep also warns about five leaked agent processes. Those are not from
ch21: with the strays cleared, a full ch21 grader run leaves no process
behind. They come from other chapters' harnesses and are also pre-existing.

## 17. Summary of what the author needs to decide or change

Nothing here is blocking; the code and grader are complete and green.

1. **crawl4ai must come out of the prose.** It ships no MCP server in its pip
   package. §21.3's "Built-in MCP server", the TL;DR's "Playwright-backed,
   handles JavaScript-rendered pages", §21.6, §21.7 and Decision #1 all need
   revising.
2. **ddgs should come out as the shipped backend**, on measured reliability:
   73% on search versus Firecrawl's 100%, and the repair requires authoring
   our own server. It remains an excellent *example* in §21.3 of the sidecar
   pattern, and the 73% number is itself good chapter material.
3. **The chapter gains a transport**, which is a better beat than configuring
   a Python sidecar: Chapter 12 explicitly left the seam open and Chapter 21
   fills it. Consider promoting this in the TL;DR.
4. **§21.4's "the only new code is configuration" needs one word of nuance** —
   it is now one transport plus configuration. The thesis survives and is
   arguably stronger, because the transport is ~180 lines against an interface
   designed for it.
5. **The brief's "always available, servers start at agent startup" is wrong
   for this architecture** and should not reach the prose. Skills are loadable
   by design; that is the point of skills.
6. **Internet is required** for the student, by the author's ruling. The
   scored grader is nonetheless hermetic, which is the right split and is what
   §21.6 already blesses.
