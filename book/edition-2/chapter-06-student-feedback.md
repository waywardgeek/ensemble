# Chapter 6: author response to student feedback

October 7, 2026. The student's initial review lives in
`solutions/edition-2/main/evidence/ch06/student-review.md`; runtime and live
validation remain separate from this teaching response.

## Unknown fields on a Gemini text-bearing part

During implementation the student found that §6.4 prevented coalescing across
unknown content without specifying how the retained neutral part preserved it.
The earlier parser could discard an extra field while keeping the text. The
student proposed preserving the whole object as standalone opaque material,
rather than inventing a second meaning for the existing text signature field.

Accepted after coordinator review. §6.4 now states the recognized text shape,
validation of known types, absent/false thought equivalence, conservative
whole-object preservation with exact provenance, and no visible/thinking delta
for that opaque object. Known signed text and calls retain their existing
rules. Plain and streaming normalization must agree. §6.5 publishes a paired
fixture, matching/foreign replay expectations, and malformed-known-field and
opaque-only negatives. This is a teaching clarification before the affected
fix; no implementation or live outcome is claimed. Student confirmation pending.
