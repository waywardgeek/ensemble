# Chapter 11 partial runtime checker invocation

This is the student-facing command contract, published before implementation
integration. It does not expose the independent checker's implementation.
The current command exercises a **partial CLI/stdin-stdout MCP runtime subset**;
passing it alone does not satisfy Chapter 11 §11.10. Public memory/custom
transport integration follows the student's reviewed, documented public API.

From the repository root, with Python 3.9 or newer:

```sh
python3 scripts/edition2/accept_ch11_runtime.py \
  --cli /absolute/path/to/ensemble-cli \
  --source /absolute/path/to/source-export \
  --source-commit FULL_COMMIT \
  --build-binding /absolute/path/to/build-association.json \
  --receipt /absolute/path/to/new-receipt.json
```

The receipt must not exist. The CLI must implement the published machine mode
and `--mcp-config FILE`; the checker uses only a local model endpoint and a
local stdio peer. It needs no credentials, provider access or compile step.
The source directory can be the corresponding main tree or an exact export.
FULL_COMMIT names its immutable outer-repository source revision under
`solutions/edition-2/main/`, even when the local directory is an export.

Reuse the source-bound build recording format already delivered with Chapter 10.
`build-association.json` contains:

```json
{
  "source_revision": "FULL_COMMIT",
  "before_after_source_equal": true,
  "inputs_receipt": "build-inputs.json",
  "build_commands": ["actual-build-command-receipt-reference"],
  "binaries": {"cli": {"path": "/absolute/path/to/ensemble-cli", "sha256": "SHA256"}}
}
```

Its sibling `build-inputs.json` contains:

```json
{
  "source_revision": "FULL_COMMIT",
  "source_sha256": {
    "solutions/edition-2/main/cmd/main.go": "SHA256"
  }
}
```

The one source entry above illustrates the schema; it is **not** a complete map.
Record the entire required pinned source set: all `.go`, `.js`, `.html`, `.css`,
`go.mod` and `go.sum` files outside evidence beneath main, including nested
modules and tests. Additional explicitly recorded source inputs are permitted;
every supplied entry is checked. Paths are repository-relative beneath main.
Record any extra build inputs too; tell the coordinator if a new input type
requires extending the required-set rule. No empty or subset source map passes.

Hash these inputs before and after the actual CLI build, preserve the exact
command, cwd, output and exit receipt, and set `before_after_source_equal` only
after comparing the maps. Hash the resulting executable after the successful
build. Record the actual command receipt references in `build_commands`; never
construct an association for an executable whose build was not observed.
The independent reviewer checks those build receipts. A hash association is an
auditable provenance record, not a cryptographic proof of compilation.

The checker validates map completeness, all mapped historical/current source
bytes and the supplied binary hash before launching or writing its receipt,
then repeats those checks before publishing results. The executable's retained
path may change; its hash must match. The source revision must match exactly.
Keep earlier failed receipts and use a new destination for each attempt.

The current subset tests literal stdio discovery/call metadata, an actual small
notebook effect, canonical Job output and model continuation, selected config
paths/environment, invalid arguments with no remote send, and whole mixed-result
refusal. A valid runtime parent is required before refusal cases run. Full
schema/bounds, authority, lifecycle, persistence, public alternative transport,
GUI, retained behavior, race/mutation checks and actual-model demonstrations
remain required under §11.10. Exit zero means only this subset passed;
`runtime_acceptance: false` deliberately remains in its receipt.
