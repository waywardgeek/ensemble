# Chapter 20 restart — independent text review

**Verdict: REQUIRED_REVISION — two local prose corrections only.**

This is the ONE formal independent text review of the restarted Chapter 20 author pass at stable HOLD, before its bounded response. The current lean restart rules supersede the historical author-first/multiple-review process. After one author response, root should verify the two corrections directly; no second formal review or new validation campaign is requested.

## Scope and reading

Read `AGENTS.md` and the complete required `docs/agentic-codebook-workflow.md`, `docs/edition-2-carryover.md`, `book/voice.md`, and `book/chapter-writing-procedure.md` in bounded reads, recovering clipped passages. Read the current lean restart instructions in `docs/edition-3-execution.md` and `docs/edition-3-student.md`, the complete current 1,204-line `book/edition-3/chapter-20.md`, the author note, the actual chapter diff, the student note, and the relevant code-review acceptance/evidence record through its final closure recommendation. Read `.agents/skills/ensemble-coding/SKILL.md` completely before checking source.

The frozen 282-file `solutions/edition-3/ch20` snapshot and R6 code/operations acceptance are inputs to this review, not reopened gates. Targeted source reads checked prose against the public construction API, fleet tool schema and question handling, durable submission rows, GUI constructor/assets/handler/relay, and the independent ArtifactScroll consumer. Relevant files include `ensemble.go`, `fleet.go`, `internal/tool/fleet.go`, `internal/common/submission_file.go`, `gui/go.mod`, `gui/assets.go`, `gui/server.go`, `gui/relay.go`, and `gui/web/gui/artifact-scroll.js`.

No chapter, source, or test was edited. No build, test, provider/browser campaign, install, commit, or delegation was performed. No private raw provider capture or credential was needed. This note is the only authored file.

## Required corrections

### 1. Remove the obsolete stub-name explanation (§20.4, lines 307–308)

The preceding line now correctly names `ensemble/gui`, matching `solutions/edition-3/ch20/gui/go.mod`. The next sentence is residue from the former `ensemble-gui-stub` spelling and no longer explains the text the reader sees.

Delete exactly:

> The latter is the optional module's historical name, not an indication that its
> browser interface is still a stub.

Keep the corrected `ensemble` / `ensemble/gui` module-replacement sentence. No replacement explanation or module change is needed.

### 2. Describe the disconnected peer accurately (§20.14, lines 1094–1096)

The accepted browser record says that **peerone**, one of the two existing TCP peers, disconnected with observation88 pending; peertwo then succeeded on id89. “A third pending caller” incorrectly suggests another caller/peer rather than another pending request on an existing peer.

Replace:

> peers reused one request ID without crossing replies. Disconnecting a third
> pending caller released its local wait while another peer, the browser and the
> Agents survived; that does not prove every browser-internal operation was aborted.

with:

> peers reused one request ID without crossing replies. Disconnecting one peer
> with another request pending released its local wait while the other peer, the
> browser and the Agents survived; that does not prove every browser-internal
> operation was aborted.

Keep the cancellation limitation. This is a correction to the description of accepted evidence, not a request for another experiment.

## Otherwise accepted text

- **GOOD PROSE / KISS:** The fresh experience paragraphs explain what failed, why it failed, what the narrower successful evidence establishes, and what remains unproved. The existing construction, ownership, task-admission and terminal-submission rationale gives the reader reasons for the mechanisms. No architectural rewrite, checklist expansion, stylistic sweep, or new requirement is warranted.
- **Bill's voice:** Bill's opener and preface are unchanged in the actual diff. Preserve them, along with the useful first-person historical experience rather than relabeling those older runs as this restart.
- **API accuracy:** `gui.New(root, agent, log)` takes an `io.Writer` and returns `(*Server, error)`; the optional GUI module is `ensemble/gui`; `Handler(directory)` mounts rather than validates an override. The factory/construction versus `Spawn` admission distinction and versioned default/custom durable-row explanation match the checked source.
- **Withdrawn brief error stays withdrawn:** The public tool is `respond_to_parent_message(agent_id, message_id?, response)`, with `agent_id` and `response` required. Omitting `message_id` is valid only when that child has exactly one pending matching question. The chapter's table and rationale correctly retain this contract; do not substitute the mistaken `request_id`/`message` brief.

## Evidence limits to preserve

- R5 supplies the full accepted gates, grade20 **200/200**, and grade1 **100/100** with genuine approximately **62-second** waits. R6 is the HTML classic-to-module entry correction plus affected regression and actual browser evidence; do not imply fresh R6 full-gate results.
- The browser exercise used **two synthetic requests**. It supports the observed public reuse, replay/turns, relay correlation and caller isolation, selected-model request, and independent two-file component behavior. It supplies no human hearing, visual-quality verdict, or reasoning-UI-change evidence. The chapter correctly distinguishes these limits, apart from correction 2 above.
- The live arithmetic follow-up used **19 POSTs**, actual **three children**, addressed question/reply, successful terminal fan-in, normal EOF, and separate source-inspected **18-assertion** success. Its first proposed join was skipped before the human hint. Active running-join hint wake and same-job survival were established **separately** by the controlled actual-Agent loopback test, not causally by that live trace. The chapter preserves the distinction and appropriately limits the single race-enabled ordering result.
- Preserve the earlier **32-POST** live failure as root's missing-tool-grants setup error: a present deferred catalogue did not grant execution availability. Preserve the key-free usage-prefix parser failure as root controller setup failure. Neither is student blame or evidence against the provider.

No other required revision identified. The two exact edits above are sufficient for root's direct post-response verification; accepted source and operational evidence should remain frozen.
