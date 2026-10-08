# Chapter 10 Q1–Q3 clarification review

October 8, 2026. Independent narrow proofreading of author freeze
`af5a7625282f1c1fffe4a18ba508e69252b647c1`, answering the initial student plan
`7cb84290af3e8bac5359339f8ab5d4e7e941bf74`. This is teaching review before
affected implementation, not approval of the complete owner/API plan or runtime.

The reviewer previously coded Chapter 8, prepared early Chapter 9/11 checks,
reviewed Chapter 13/14 contracts and audited/proofread Chapter 9. The reviewer
did not author this clarification or implement Chapter 10. Reads for this task:
full current voice and writing procedure, the complete student plan, new Chapter
10 and the complete three-file author delta/direct response. Truncated combined
output was followed by focused reads. Some chapter text was first read while
the author was finishing it; the final working bytes were then confirmed equal
to the published freeze. No old answers, historical implementation or checker
internals were opened. No implementation, checker, provider or build work occurred.

| Question | Independent disposition |
| --- | --- |
| Q1: physical record bounds and refusal precedence | Resolved in §10.8. Every session event has the 64 MiB physical bound, including actual LF/whitespace/escaping; standalone per-kind limits and all header limits remain inherited. Explicit readers choose mode first; generic Load has a bounded first-record probe and cannot switch on an invalid/interior initializer. Incremental reading prevents an unbounded allocation before the check. Skill candidate preflight retains controlled skill_too_large; separate session storage admission retains terminal session_limit. Earlier decoded/client/source bounds remain intact. |
| Q2: imported history | Resolved in §10.4. Raw Events/Dump describe actual anchor/tail records; inspection and /history label the origin and incomplete history. Genuine saved watch facts may precede the origin without becoming fabricated raw records. A pre-origin send returns history_unavailable; available-tail reconstruction seeds from validated origin and uses that send's captured inputs. Full-origin reconstruction remains complete. |
| Q3: original bytes versus semantic equality | Resolved in §10.3. Replay-bearing raw JSON and exact text retain recorded bytes and number lexemes. Schema/identity/argument-value comparisons use canonical semantic equality only where specified. Snapshot/rebuilt comparisons must still detect different replay-bearing bytes. Validated raw-JSON strings are a permitted private encoding, not prescribed field names or an exemption from nested validation. |

No newly contradictory instruction was found in the affected contract. These
distinctions preserve one durable writer, strict origin validation and the
difference between a reduced conversation and raw event history. The direct
response accurately classifies the initial questions as teaching omissions or
clarifications and leaves student confirmation and plan approval separate.
Updating the stale checker sentence to point to the existing initial command
does not claim that its subset covers the full chapter matrix.

The retained prose executable independently passes all hard rules at 7,289
words. Its existing dense-contract person-gap warning remains; this narrow
clarification makes the three mechanisms more explicit without inventing a
story or successful restart. The complete diff leaves the sole existing JSON
fixture unchanged. Scoped review introduces no fixture or runtime test claim.

Frozen author hashes:

- chapter-10.md: `46249bc8fadd59f55c5dc12aebfe2c3ca22ce9c031d4a7972aa97e7140b22d60`.
- chapter-10-evidence.md: `9ad1942ad47a510a07638fa22d9e9b12ace7ad919b7824e31fef9cdb7620c98d`.
- chapter-10-student-feedback.md: `63bea996ae2e31aa96e9f609287d1d9c6f6c834289fef263ab4013ca73beb46c`.

Disposition: Q1–Q3 clarification proofreading accepted, no remaining finding in
this scoped pass. This does not replace full final manuscript review after
implementation, actual runs and author reconciliation.

## Separate Unicode clarification closure

Reviewed the complete three-file author correction
`cf73a646ba665d2070bb51a77bc0b1c5c4912a2b` and its placement in §§10.3/10.8.
Current voice/procedure remained loaded with no intervening compaction. Working
author bytes match the freeze. This addendum is a coordinator/grader-discovered
lexical-validation gap; neither the chapter evidence nor direct feedback
retroactively attributes it to the student's Q1–Q3.

The new boundary is coherent: session/canonicalization strings and keys require
Unicode scalar values, so lone high/low surrogate escapes refuse before lossy
decoder replacement. Valid adjacent pairs and genuine literal/escaped U+FFFD
remain valid. The escaped-backslash example correctly represents ordinary text,
not a surrogate escape. Replay-bearing raw JSON, including its dedicated string
wrappers, receives the same nested validation while retaining exact valid original
bytes. This does not turn arbitrary text strings into JSON documents or replace
byte-preserving replay with canonical equality. Legacy standalone decoding is
expressly unchanged; the strict session admission scope is named rather than
silently rewriting predecessor behavior.

No new contradiction or unresolved wording issue was found. The author's local
Go documentation read remains its attributed evidence; this proofreading pass
does not claim a runtime reproduction or completed implementation. The retained
prose executable independently passes hard rules at 7,422 words, with soft
negation/person-gap warnings. The correction's necessary lexical distinctions
introduce no new story or success claim. Existing fenced JSON is unchanged.

Reviewed hashes:

- chapter-10.md: `3faa154f60e875497d96594a1afdc219f398856b0a9bd46b39cf72f36ca126df`.
- chapter-10-evidence.md: `de9e470e5a0d8cef173b61a1ce8f90700198e2537889a789c3fcf41c7dabf0b0`.
- chapter-10-student-feedback.md: `cfb93f93cf7b53440a92c7304c8cb690a65e69f887d6c758cdfa1370d5b7f264`.

Disposition: Unicode clarification proofreading accepted. No implementation,
grader, legacy, provider or build work occurred; future affected checks and
student confirmation remain separate.
