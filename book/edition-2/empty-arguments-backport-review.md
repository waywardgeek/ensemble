# Review of the zero-byte Messages argument correction

Chapter 6 scoped maintenance accepted by the coordinator on October 8, 2026.
The maintenance coder `/root/grader_ch05` had seen later second-edition work;
this is explicitly not a new cold-student attempt. The coordinator is independent
of that maintenance implementation. Chapter 7/8 propagation remains pending.

The isolated Chapter 6 branch starts at accepted source
`c3fa7583c5e4c3dcf026b3d5a0a93da000dd8307`. Runtime and regression commit
`788c5e9a6928fb125c9e606336ad3a67632b6cc6` changes five production lines and
adds a separately derived regression and original provider fixture. Evidence is
frozen at `5e48b3818147cbfd5d2e064794146f211802d5e1`. No later implementation
file was copied backward. The original chapter tag and live receipts remain
unchanged.

The coordinator read the changed production code, complete regression, relevant
Chapter 6 contract, maintenance README, source-bound summary and mutation
adapter. A zero-length argument fragment provides no replacement bytes. The
guard therefore leaves the start object intact; nonempty fragments still
replace it rather than merge. Returning the parser's sticky error preserves
wrong-type refusal. Whitespace remains nonempty and fails JSON-object
validation. The guard does not change block completion, message termination,
actor acceptance or tool dispatch.

The 2,668-byte fixture is the actual retained Chapter 9 Messages response,
SHA-256 `e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef`.
Its final blank line belongs to SSE framing and is intentionally preserved.
The 13-case regression checks start retention, nonempty replacement, malformed
and non-object inputs, wrong types, missing termination, exact usage, and one
complete display of arguments. The original source fails exactly the recorded,
one-empty and two-empty controls; ten other controls pass. All 13 pass after
the repair. These outcomes distinguish the defect from unrelated setup failure.

The retained gate passes 22 of 23 command groups, including all seven module
vet/tests. The remaining group fails because its globally unique source anchor
now appears twice. The maintenance-only adapter qualifies the same
thinking-suppression mutation with its containing method; it changes neither
the positive test nor the intended refusal. That separate positive/mutant pair
passes, as do the other 18 original deletions. The failed full-gate receipt
stays failed rather than being rewritten as a clean rerun.

The archived repaired CLI passes all 14 existing independent local HTTP
controls, including the exact original response and a held message-stop barrier
that permits no accepted response or tool effect early. Core race checks pass.
The coordinator independently verified all 78 bound source hashes, 12 retained
receipt hashes, 14 checker hashes and the executable hash
`f4eea3fcde69a4faf288a37559e9f1b7271a48493261232249268a99c5ae9b69`.
Only the original `stream.go` blob changes; 1,470 earlier evidence files remain
unchanged. These are inspected, source-bound test results; no additional broad
coordinator rerun is claimed.

Chapter 9 separately demonstrates the repaired shared parser on a real Messages
response with empty arguments, a successful actual tool call and final answer.
That receipt retains its Chapter 9 runtime identity. The Chapter 6 correction
makes no new all-provider or native-GUI claim and does not relabel old live runs.
Its earlier unaffected live evidence remains applicable at its original source.

The explicit zero-byte rule is taught at `305b1b0`. Accepted Chapter 7 and 8
sources contain the same original parser bytes; propagation must begin from
those exact earlier sources, retain their own evidence and validate affected
behavior before a new revision tag. Current Chapter 9 main already contains its
independently implemented repair. Earlier source trees must never replace it.
