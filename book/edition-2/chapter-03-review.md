# Chapter 3 independent contract and prose review

Date: 2026-10-07. Status: final manuscript and live receipts accepted; see
the post-run review below. The earlier contract review is retained to show
what the fresh student received. Code comparison and revisions are accepted
separately, and the coordinator's full legacy regression subsequently passed.
Bill's later human-chat requirement reopens the client teaching and live gate;
this review records the earlier scope, not acceptance of that new interface.

## Reading and scope

Freshly read the full current voice guide, chapter-writing procedure, coding
skill, architecture decisions, and Chapter 3 manuscript. Read its evidence
note and checked the predecessor's amended deferred-input and provenance
rules against the new loop. Earlier whole-book findings remain in the global
map; this review does not claim the entire textbook is still resident.

The opening gives the reader a concrete reason to build tools: an answer
does not establish that a file changed. The later sections carry that stake
through independent disk observations, preserved results, and explicit
side-effect boundaries. No invented new run appears in the pending spin.
The historical tool-use exhibit is identified as a dated private corpus,
without extending its measurement to current users or models.

## Contract findings and resolution

1. Search originally said that reaching a limit produces a truncation notice,
   conflicting with the earlier omitted-content rule. The revised text
   requires an omitted match or byte; a complete result exactly filling its
   limit is not reported truncated.
2. “Regular text files” left binary classification unspecified. The contract
   now states a NUL-byte test over the first 8192 bytes, making a future
   independent check reproducible instead of grading an unstated heuristic.
3. Automatic continuation still obeys Chapter 2's exact provenance rule.
   The live instructions now explain selecting a known resolved identity
   before a tool turn when needed, rather than guessing aliases after a
   side effect. This clarifies the existing requirement without weakening it.

The reviewer rechecked all three amendments. No remaining material
architecture or synchronous lifecycle contradiction was identified.

## Properties to retain during implementation

- Registry storage is an explicit Agent-owned working choice allowed by
  Bill's broader ruling. The tools service has an actual Agent interface
  parent; Engine reaches it through Agent, with no llm/tools sibling import.
- Declarations and dispatch share one capability set. The empty public
  selection, CLI's explicit six-tool selection, and offline-only arbitrary
  declaration snapshots are distinguished.
- Dispatch facts persist before effects; result facts persist afterward.
  A result-write failure cannot roll back an executed command. Ordinary
  tool errors complete their batch, while persistence and transport failures
  end the operation. Request sixteen's entire batch completes before refusal
  of request seventeen.
- Recorded replay never executes a tool. Deferred human input retains its
  source sequence while projecting after completed results. Bound signatures
  and usage remain tied to the producing response.
- Read/search fixtures are independent of writing. Silent `exit 7` cannot
  pass by extracting a wrapper's diagnostic. Cap controls need both exact
  complete outputs and outputs with actual omissions.
- Current schemas and all three result/error wire representations are
  taught explicitly. GUI remains a separate optional stub/public client.

The Chapter 2 JSON Schema mapping was independently checked against the
official [FunctionDeclaration documentation](https://googleapis.github.io/js-genai/release_docs/interfaces/types.FunctionDeclaration.html)
on October 7, 2026: `parametersJsonSchema` accepts an object JSON Schema and
is mutually exclusive with `parameters`. This supports the teaching carried
into Chapter 3; it does not establish a new live run.

Future student evaluation must use a fresh context and the new-material-only
source boundary in the current procedure. Author/reviewer historical research
is not a student input. Bill's editorial approval remains separate.

## Post-run full-manuscript and receipt review

The reviewer reread the full latest Chapter 3 after reconciliation, including
the clarified empty-file and UTF-8 rules and the actual spin. Checked all three
initial CLI stdout streams and neutral logs, recomputing each usage total from
accepted response records: Messages 56911/0/0/1976, Chat Completions
4004/0/20608/775, and generateContent 50081/0/0/1469 (input/write/read/output).
The selected/returned identities agree with the manuscript. Every initial
path records all six tools, actual error results for overwrite and missing/
ambiguous anchors, and subsequent recovery. The receipts' independently
observed final file contents match the printed examples.

Checked the three consumer stdout/report artifacts: NORTH-314 and SOUTH-927
are read from independent workspaces; the empty-selection Agent returns READY.
The consumer implementation checks observation attribution/order and performs
an actual post-unsubscribe event before confirming absence of delivery.
Checked offline receipt flags and compared the two retained render files
directly on each surface: byte equality holds, and the receipts identify
keyless commands, unchanged log, and unchanged scratch-file hashes.

The guided-run stdout/logs separately show Silent Seven, SilentEcho, and
Cobalt Falcon invented then recalled, and all three empty-read modes plus
silent exit 7. Their binary hash differs from the initial run as the prose
states. These are real guided runs, not initial-attempt evidence.

No material prose or receipt blocker remains. The author should replace the
final pending UTF-8/comparison sentence once the coordinator binds the accepted
revision and remaining full-regression gate. UTF-8 sensitivity and the accepted
code changes are recorded in `chapter-03-code-review.md`; no new paid run is
claimed for that deterministic boundary correction.
