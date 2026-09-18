# Brief — Chapter 2 amendment: the `Ref` type

You are the CODER. I am the author. Repo: `~/projects/ensemble`.

Read `book/chapter-05-seam-draft.md` §1 and §2 first — it states the design and
the reasons. This brief states the PROPERTIES to achieve, not the mechanism; if
you find a better mechanism that satisfies them, take it and say so.

## Build and run

```bash
cd ~/projects/ensemble
go build ./...
go run ./cmd/grade -ch 2 solutions/ch02     # grader dir is POSITIONAL
go run ./cmd/grade -ch 3 solutions/ch03     # must also still pass
```

Gotchas that have bitten before:
- `cmd | head && echo OK` LIES — the pipe's exit status is `head`'s.
- The fake vendor in `internal/fakevendor` is SHARED across chapters. Any change
  to it must be regression-checked against every chapter's grader.
- Never `git add -A` here. Add named paths.
- `ch02.log` files are tracked and churn.

## The problem

`solutions/ch02/part.go`:

```go
type BlobPart struct{ MIME, Path string }
type RedactedPart struct{ Stub string }
```

A local path cannot express three of Gemini's four input methods (File API
`uri`, `gs://`, external URL), nor Anthropic's `file_id` source. Chapter 5 ships
its seam as printed Go and cannot promise a type it knows is wrong.

## Target

```go
type RefKind uint8

const (
	RefPath   RefKind = iota + 1 // a file on local disk
	RefURI                       // remote: vendor File API uri, gs://, https://
	RefHandle                    // framework-managed output; may be in memory
)

type Ref struct {
	Kind    RefKind `json:"kind"`
	Locator string  `json:"locator"`
}

type BlobPart struct {
	MIME string
	Ref  Ref
}

type RedactedPart struct {
	Stub string
	Ref  Ref // zero when the superseded content had no locator
}
```

## Rulings — these are decided, do not relitigate

1. **There is NO inline-bytes case in `Ref`.** `BlobPart`'s existing doc comment
   says "never inline… a log you cannot grep is a log you cannot debug." Base64
   is a RENDERING decision made when building a vendor request, and is never
   written back into the log. Keep that comment's promise.

2. **`RedactedPart` stays SYNTHESIZED by the reducer, never stored.** Its `Ref`
   is CARRIED FORWARD from the part it supersedes — if the superseded part was a
   `BlobPart`, its `Ref` survives into the stub. Do not add storage that records
   a redaction separately; the log already does, permanently. Net effect:
   redaction becomes recoverable by construction.

3. **This is a HARD BREAK of the log format, and that is correct.** Do NOT accept
   the old `"path"` field as a fallback. Silently coercing an old field into a
   new type is the exact anti-pattern this book argues against, and `part.go`
   already refuses unknown part types with "refusing to load this log." Match
   that discipline: an old-format log must fail LOUDLY. Regenerate fixtures.

4. `RefHandle` is deliberately distinct from `RefPath`: a handle is resolved by
   the framework and need not be a filesystem path, because the jobs chapter
   allows in-memory buffers.

5. `partJSON` keeps its ordered-field struct shape. Do not switch to a map — the
   existing comment explains why (Go randomizes map iteration order, so a map
   produces different bytes on a future run).

## Properties that must hold when you are done

State each as a check, and make each fail if the behavior is deleted.

1. A `BlobPart` with each of the three `RefKind`s round-trips through
   `PartList` marshal/unmarshal unchanged.
2. A log written with `RefURI` renders to a request that uses the vendor's
   remote-reference form, NOT a local path, for at least one vendor that
   supports it.
3. An old-format log (a blob part carrying `"path"` and no `"kind"`) is REFUSED
   with a clear error. It must not load as a zero-valued Ref.
4. A `RefKind` of zero is invalid and refused — the constants start at `iota+1`
   precisely so the zero value cannot be mistaken for a real kind.
5. A redaction applied to a span containing a `BlobPart` produces a
   `RedactedPart` whose `Ref` equals the superseded part's `Ref`.
6. `go run ./cmd/grade -ch 2 solutions/ch02` scores 100, and `-ch 3` is
   unaffected.

## The audit that actually matters (P9)

Course policy P9: every grader is audited BY DELETION, not by reading.

For each property above: delete the behavior from the REFERENCE SOLUTION,
re-run the grader, and confirm the score drops AND that the exact set of failing
check IDs is what you predicted. A check that cannot fail is a green dashboard
with a schema around it.

Report the mutant table: what you deleted, which checks fired, final score.

Prior art, and the reason this is not optional: in Chapter 2 the `Opaque` field
was graded by NOTHING — deleting its only use still scored 100/100. In Chapter 1
the content-block walk was graded by nothing, because the fake served only one
block. In Chapter 3, results-first ordering was graded by nothing. Three for
three. Assume this amendment has the same hole until you have proven otherwise.

## Do NOT

- Do not invent wire facts. If you need to know how a vendor expresses a remote
  file reference, check the real docs and cite the URL in your report. Do not
  guess a field name.
- Do not add a compatibility shim for old logs.
- Do not "improve" unrelated code. This is a narrow amendment.
- Do not change `internal/fakevendor` without regression-checking every chapter.

## Deliverable

1. The code change, committed.
2. `book/review-ch02-ref.md` — what changed, the mutant table, anything you
   found that contradicts this brief, and any wire fact you verified with its
   source URL.

If a ruling above turns out to be wrong, STOP and say so in the report rather
than working around it. A brief that is wrong is my error to fix, not yours to
absorb.
