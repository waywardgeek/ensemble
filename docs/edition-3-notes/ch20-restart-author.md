# Chapter 20 restarted author pass

Status: **stable HOLD for root's one formal independent text review**. One fresh post-experience manuscript pass; no reviewer delegated and no code work reopened. Only `book/edition-3/chapter-20.md` and this note were edited. Existing worktree chapter changes, Bill's opener and preface, historical experience and useful ownership rationale were preserved.

## Context and source

Read root `AGENTS.md`, required coding guidance and author skill, complete current `book/voice.md` and `book/chapter-writing-procedure.md`, and complete historical workflow/carryover (root confirmed the latter remained mandatory). Bounded reads recovered clipped ranges. Read the current chapter, Chapter 1 architecture and Chapter 19 prerequisite seams, current Chapter 20 student brief, recent execution acceptance, and student/reviewer FINAL closure. Source checks used only the accepted public `solutions/edition-3/ch20` snapshot, notably `fleet.go`, `config.go`, common config/submission definitions, `internal/tool/fleet.go`, CLI factory construction, and GUI server/assets/relay/projection/browser modules. No original solution or private grader implementation, provider/browser/native run, install, deployment or commit.

Root's lean restart overrides historical author-first/two-pass process. This note records an author pass, not independent review or a new implementation acceptance.

## Sections changed

- **20.1:** Exact `gui.New` return and `io.Writer` argument; local app-only submit versus admitted fleet completion; versioned default and custom durable-row shapes, whole-file validation, and defensive ownership.
- **20.4:** Selected-view isolation stays in the existing scoped hub, not a second conversation or fleet registry.
- **20.6:** Replaced obsolete `SubAgentManager`, `AgentSpec`, `defaultFactory` and `config.Clone` sketches with the actual `AgentFactory(context.Context, Config)`, `CreateAgent`/`NewAgent` construction versus explicit `Spawn` admission, and `Config.Child` state ownership. Kept the reason application policy belongs at construction and actors own lifecycle. No compatibility framework or alternate provider path introduced.
- **20.7:** Identity sketch now uses `Config`; a root's `ParentID` is empty, while its parent-message route reaches the human sink.
- **20.8:** Durable submission remains authoritative over optimistic status; schema/payload bytes are independently owned.
- **20.10:** Both terminal success and failure close task admission while permitted cleanup drains.
- **20.13:** Corrected stale Chapter 22 temporary names, and the standalone component frame to the actual `agent_id` plus numeric `part_id`/`role` envelope. Override instructions preserve the real module imports. Distinguished launcher validation of `index.html`, `app.js`, `style.css` from low-level `Handler(directory)`, which only mounts the caller's directory.
- **20.14:** Three short restarted-experience paragraphs: the HTML module-entry failure and root's missed integration check; successful real arithmetic fleet plus root-owned setup failures; separately established active-join hint ordering. Existing historical runs remain identified as historical, rather than relabeled as this restart.

The parent-reply signature in the initial author brief conflicted with public source. Asked root rather than silently redesigning the contract. Root confirmed the brief was mistaken: retain `respond_to_parent_message(agent_id, response, message_id?)`; omission of `message_id` is valid only for one pending matching question. Existing table/rationale therefore stay intact. Callback controls remain inside `tool_limits`.

## Evidence limits

No operational gate was rerun by this author. Root's current acceptance record is authoritative: grade20-v4 200/200, pass true, exit 0; grade1 100/100, pass true, exit 0 with genuine approximately 62-second waits. R5 full core/GUI race, vet and module results carry forward; R6 was the HTML classic-to-module correction plus served-HTML regression and rebuilt outside-consumer Chrome evidence. The root grader executable's v8 label is distinct from result filenames.

The recent browser run used synthetic requests. It established replay, two composer responses, separate override, same-ID relay correlation and caller-disconnect isolation. It did not establish cancellation of all browser-internal work, a reasoning UI change, visual beauty or native hearing; screenshots were not viewed. The independent component used the actual two-file import graph. Root owns the earlier module-only harness's failure to load real HTML.

The paid arithmetic follow-up completed real children, the negative-integer question/reply and final all-submission join; root inspected the files before independent 18-case success. The earlier 32-POST attempt failed because root omitted primary tool grants despite the deferred catalogue. The key-free usage-prefix parser failure also belonged to root's controller. The paid first proposed join was skipped before the hint, so it proves no active-join wake. The separate actual-Agent loopback test established running join before hint, non-error observation, parent continuation with child blocked, and terminal typed payload from the same handle after real submission. Its 20-second context was below the 60-second observation limit; one race pass proves that exercised order, not every interleaving. The manuscript does not merge those traces.

## Prose lint

Ran the established Chapter 19 author command against Chapter 20 only:

```sh
/usr/bin/env -i HOME=/tmp TMPDIR=/tmp PATH=/usr/local/go/bin:/usr/bin:/bin:/usr/sbin:/sbin GOCACHE=/Users/bill/Library/Caches/go-build GOMODCACHE=/Users/bill/go/pkg/mod GOMAXPROCS=2 GOFLAGS='-mod=readonly -p=1' GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off GOWORK=off CGO_ENABLED=0 /usr/local/go/bin/go run ./cmd/lintprose -v book/edition-3/chapter-20.md
```

Exit 0; all hard checks pass. Soft warnings: negation forms 17 against 16, and person gap 4106 words after line 519. Kept the existing explanatory voice rather than make unrelated prose changes to satisfy soft counters. Pager-free `git diff --check` for the chapter also passed; the first whitespace-check invocation opened a pager and was stopped, not counted as a failed manuscript check. No broader tests run.

## One bounded response to formal text review

Read `docs/edition-3-notes/ch20-restart-text-review.md` and implemented exactly its two REQUIRED_REVISION items: deleted the obsolete historical/stub module-name sentence in §20.4, preserving the `ensemble` / `ensemble/gui` replacements; changed §20.14 to describe one of the two peers disconnecting with another request pending, preserving the explicit limit on browser-internal cancellation. No other manuscript revision was made in this response.

Reran the established isolated Chapter 20 prose-lint command above: exit 0, all hard checks pass. Negation forms now pass (16 against 16); the remaining soft warning is person gap, 4109 words after line 517. The scoped `git --no-pager diff --check -- book/edition-3/chapter-20.md docs/edition-3-notes/ch20-restart-author.md` passed. Root separately reports an offline build-only compilation of the unchanged outside-GUI recipe against accepted Chapter 20 at `/private/tmp/ch20-prose-compile.5LRqlf`; this is root's result, not an author-run build or runtime observation.

Final status: **stable HOLD for root's direct verification**. This exhausts the one authorized bounded author response. No second formal review, code/source/test edits, operational work, other-chapter changes, commit or delegation.
