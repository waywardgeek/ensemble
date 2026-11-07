# Model selection usability follow-up

Bill requested dropdowns for Create Agent's model and primary skill, enabled
cross-vendor selection using his available credentials, and a separate Ollama
vendor group. The agreed policy is application-wide last-selected funding per
vendor, with each Agent retaining its own settings and exact model offering.
Changing defaults must not retarget existing Agents. Unavailable funding refuses
rather than falling back. Agent settings store references, not credential values.

## Implemented first slice: qualified peer creation and dropdowns

CreatePeerAsync accepts an optional qualified offering, mutually exclusive with a
legacy model-name override. Resolution and request-policy validation occur before
peer construction and registry commit. Existing inheritance, startup, cancellation
and restoration behavior remain. A read-only primary-template catalog exposes
names/descriptions without granting tools or exposing variables/manuals/secrets.

The guarded, no-store /peers/choices route reuses the existing model projection.
Create Agent groups offerings by vendor, shows funding, disables unavailable
choices, and lists only primary skills. Inherit choices identify the mounted Agent
as the source. Opening the dialog refreshes choices; creation revalidates them.
Loading/failure text appears inside the dialog; a failed fetch disables creation.
No provider request is made by enumeration.

Evidence: /private/tmp/cadence-model-preferences.ERLGRh. Clean core build and six
peer groups pass, including explicit cross-vendor creation and missing/conflicting
offering refusal without registry or directory effects. Clean GUI build/default
suite passes with real guarded HTTP catalog/creation checks and joined teardown.
Shared core suite passes370/0. The shipped dialog's DOM/HTTP-boundary browser check
passes desktop/mobile, exact offering payload, primary choice, inheritance and
failed-load/retry cases. That browser fixture substitutes the HTTP boundary;
actual server behavior is covered separately, not claimed as one integrated live
browser/provider experiment. Final browser-only refinements followed the shared
run and were rechecked with the browser script. No installation or restart.

## Remaining integration

The current host registers only its launch credential binding for a model document.
It does not automatically import other inference keys; the existing private-file
reader is explicitly for transcription, not general inference. Choose an explicit
inference key-store input before extending host bindings; never silently reuse the
speech payer. Application-wide funding preferences and per-Agent selected-model
persistence are not yet implemented by this slice. Existing Agent settings currently
cover memory, recall and working-context policy; child restart records separately
retain model identity. Reconcile that storage rather than writing competing live
selection authorities. Host-approved Ollama configuration must remain explicit.
