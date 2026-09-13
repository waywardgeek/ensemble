# Review — Chapter 2 amendment: the `Ref` type

Coder's report to the author. Repo state: four commits on `main`,
`6a253cf` → `f5f9a29`. `go build ./...` clean, `go test ./internal/grade/...`
green, `-ch 2 solutions/ch02` = 100/100, `-ch 3 solutions/ch03` = 100/100.

---

## 1. What changed

**`solutions/ch02`** (`6a253cf`) — the type, exactly as the brief specified.
`RefKind` (constants at `iota+1`), `Ref{Kind, Locator}`, `BlobPart{MIME, Ref}`,
`RedactedPart{Stub, Ref}`. No inline-bytes kind. No compatibility shim. An old
`"path"` log is refused by name.

Three mechanism choices worth naming, none of which the brief dictated:

- **`partJSON.Ref` is a POINTER.** A `RedactedPart` that superseded content with
  no locator has a legitimately zero `Ref`, and writing `{"kind":0,"locator":""}`
  would put an invalid kind on the wire for the loader to reject on the way back
  in. `omitempty` needs a pointer to express "legitimately absent".
- **`partJSON.Path` is RETAINED** — not to read, but so the loader can recognize
  a pre-Ref log and refuse it *by name*. Deleting the field entirely would make
  an old log fail with the vaguer "blob has no ref".
- **`Ref.validate()` refuses an empty locator as well as an invalid kind**, so a
  half-filled Ref cannot reach a renderer.

**`solutions/ch03`** (`a4626a1`) — mirrored, because `ch2parity` re-runs
Chapter 2's real checks against the Chapter 3 binary. Without this ch3 scored
90/100. `part.go` and `context.go` were byte-identical to their pre-amendment
ch02 counterparts and were copied wholesale; `seam.go` and the three renderers
had already diverged for tool execution, so the amendment was applied by hand.

**`internal/grade`** (`2918d9f`, `f5f9a29`) — six fixtures, a harness phase,
five checks, five mutations.

---

## 2. What is graded by what

| Brief property | Check ID | Fixture | Driven through |
|---|---|---|---|
| 1. three kinds round-trip | `ref-roundtrip` | `RefRoundTripLog` | `dump` |
| 3. old-format log refused | `ref-oldformat` | `RefOldFormatLog` | `dump` |
| 4. zero kind refused | `ref-zerokind` | `RefZeroKindLog` | `dump` |
| 2. `RefURI` → remote form | `ref-render` | `RefURILog` | `render` (gemini) |
| 5. stub carries the Ref | `ref-redaction` | `RefRedactionLog` + plain control | `render` (gemini) |
| 6. 100/100, ch3 unaffected | — | — | both graders |

Three decisions here are load-bearing and would be easy to undo by accident:

**The serialization fixtures go through `dump`, not `render`.** Properties 3
and 4 are about the *loader*. Rendering them lets the renderer's own kind switch
refuse a malformed Ref that the loader happily accepted — the check passes, and
the loader bug stays invisible behind a renderer that caught it later. I found
this by predicting the mutation outcome and getting it wrong on paper first.

**`ref-redaction` has a negative control** (`RefRedactionPlainLog`, the same log
with no `Redacted` event). Without it, a submission that simply never renders
tool results passes the "secret is absent" half for the wrong reason. The
control asserts both that the secret *is* present un-redacted and that the
locator *is* present, so the check can show that redaction removed one and kept
the other.

**Every fixture carrying a live blob renders under Gemini only** — see §4.

### Point weighting (settled)

`ch1parity` 25 → **10**; the five `ref-*` checks take the freed **15**. Total
remains exactly 100, as `TestCh2ReferenceSolutionScores100` requires.

The rationale is a symmetry worth stating in the text: **Chapter 3 already
weights its own regression gate, `ch2parity`, at 10.** A regression gate is
therefore worth 10 in both chapters, and the points go to the chapter's new
material rather than to a re-run of material the student already passed.

Within the 15, the split is `ref-roundtrip` 3, `ref-oldformat` 2,
`ref-zerokind` 2, `ref-render` 4, `ref-redaction` 4 — the three serialization
IDs sharing one budget of 7. See §5(a) for why that budget is *split* rather
than parked on one ID.

---

## 3. The mutant table (P9)

Measured, not asserted by hand: the test now computes the score and logs the
row. All 22 mutations pass; every one drops the score.

| Mutation | Score | Checks fired |
|---|---|---|
| `ref-not-serialized` | 97/100 | `ref-roundtrip` |
| `old-path-coerced-into-refpath` | 98/100 | `ref-oldformat` |
| `zero-kind-accepted` | 98/100 | `ref-zerokind` |
| `refuri-not-rendered` | 96/100 | `ref-render` |
| `stub-drops-superseded-ref` | 96/100 | `ref-redaction` |

Three pre-existing mutations now fail an additional check. I checked each
against "is the grader right?" before widening, and in all three the new failure
is a true consequence, not a false positive:

| Mutation | Score | Checks fired | Why the extra one is real |
|---|---|---|---|
| `gemini-uses-messages-key` | 81/100 | `seam-render`, `ref-render` | the fileData part lives *inside* `contents`; rename the array and a remote reference cannot be delivered either |
| `redaction-ignored` | 86/100 | `redaction`, `ref-redaction` | no redaction ⇒ no stub ⇒ no superseded Ref to carry forward |
| `dump-prints-nothing` | 77/100 | `logdump`, `seam-parse`, `ref-roundtrip` | `ref-roundtrip` is graded on `dump` |

**Two mutations needed care to be real rather than decorative.**

`old-path-coerced-into-refpath` does **not** delete the refusal — it replaces it
with the coercion the loader's comment warns about. Simply deleting the error
leaves the "blob has no ref" guard to refuse the log anyway, so the check would
have passed while the behavior it names was gone. This is the same trap the
brief's prior art describes, and it caught me: my first draft of this mutation
was the useless one.

`refuri-not-rendered` deliberately mutates the **Blobs** loop and not the
carried-forward **Refs** loop below it, which renders a nearly identical line.
`buildMutant`'s exactly-one-match-site assertion is what forced the distinction.

**A change to the audit itself:** `TestCh2MutationsAreDetected` now asserts
`earned < max`. Asserting the failing-ID set alone cannot catch a check worth
zero points — which is precisely how a green dashboard with a schema around it
gets built. This matters immediately; see the open question in §5.

---

## 4. Wire facts, with sources

Verified against live docs. Nothing below was guessed.

- **Gemini has four file input methods**, three of which are not local paths —
  File API upload, registered `gs://` object, external URL.
  <https://ai.google.dev/gemini-api/docs/file-input-methods>
  This is the brief's premise, and it is correct.
- **Gemini's remote reference in a Part is `fileData`**, and `FileData` is
  `{mimeType, fileUri}` — `fileUri` required, `mimeType` optional.
  <https://ai.google.dev/api/generate-content>
  This is what `RefURI` renders to, and what `ref-render` grades exactly.
- **Gemini's File resource carries a `uri`**; `files.register` registers GCS
  URIs. <https://ai.google.dev/api/files>
- **Gemini inline bytes are `inline_data: {mime_type, data}`, base64.**
  <https://ai.google.dev/api/caching> — the contrast case. `ref-render` asserts
  this form is *absent*, since inlining a URI means something read a URI as a
  filename.
- **Anthropic references an uploaded file as**
  `{"type":"document","source":{"type":"file","file_id":"..."}}`.
  <https://docs.claude.com/en/docs/build-with-claude/files>

**Not verified, therefore not implemented:** Anthropic's `source:{"type":"url"}`
form, and OpenAI's file-reference shape. Per "do not invent wire facts", blob
rendering exists for **Gemini only**. Anthropic and OpenAI raise a loud,
explicit error on a `BlobPart` rather than guess a field name — and crucially
rather than *drop* the blob, which is how a multimodal request silently loses
its attachment and comes back with a confident answer about a file the model
never saw. This is why every live-blob fixture is graded under Gemini: under the
other two the checks would be grading the refusal, not the Ref.

---

## 5. Open questions for the author

**(a) I did not park the shared budget on one check, and I think your rule
needs this amendment.** You asked for five IDs covering three skills, with
points split only where skills separate. Parked literally — `ref-roundtrip` 7,
the other two 0 — the two zero-point checks become unfalsifiable: deleting the
path-refusal fires `ref-oldformat` and the score stays 100. So the serialization
budget is *split across* its three IDs (3/2/2) instead. Losing the skill
entirely still costs 7, not 21, which I believe preserves your intent; but
"itemize for diagnosis, split points only where skills separate" and P9 are in
direct tension whenever an itemized check is worth zero, and the rule as stated
resolves it the wrong way. The `earned < max` assertion added in §3 now makes
that failure mode impossible to reintroduce silently.

**(b) Mutation mechanism deviates from your instruction, deliberately.** You
asked for mutants under `testdata/students/mutant` with `COURSE_MUTATION`. That
is Chapter 1's single-file pattern. Chapter 2 already has a better one: the
`mutations` table in `ch02_grader_test.go`, where `buildMutant` copies the
**real reference solution**, applies regexp edits, asserts exactly one match
site, and builds. It audits the actual reference rather than a hand-maintained
copy that can silently drift out of sync with it — which is what the brief's
P9 actually asks for. Say the word and I will convert it.

**(c) Chapter 5's printed seam.** The amendment makes the type honest, but
Anthropic and OpenAI blob rendering is still unimplemented. If Chapter 5 prints
a seam that claims to render blobs for all three vendors, that promise is not
yet backed by code. Worth deciding before it goes to print whether Chapter 2
should carry the Anthropic `file_id` path (the wire fact is verified and in §4)
or whether the honest `not implemented in this chapter` error is the lesson.

**(d) Nothing in the brief turned out to be wrong.** Every ruling held up under
implementation. Ruling 2 in particular — carrying the Ref forward into the stub
— is the nicest thing in the amendment: it makes redaction recoverable by
construction rather than by a side table, and `ref-redaction` grades exactly
that.
