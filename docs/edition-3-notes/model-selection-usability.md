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

## Durable explicit Agent selection

A persistent Agent now saves explicit configured selection in its own settings.json
as selected_offering, a strict vendor/endpoint/model reference. The actor saves
before publishing the new Model membership/configuration. No credential or endpoint
copy is stored. Memory, recall and working-context writes preserve the selection;
selection writes preserve those sibling policies. Transient Agents keep existing
in-memory behavior. Legacy name-only selection retains its existing compatibility
path; no endpoint or funding is invented from a wire name.

The shared bounded reader uses an actual Sandbox before Model attachment and Engine
construction. Sandbox initialization reads filesystem security policy independently
of inference admission; normal Configuration/render/send validation remains strict.
A saved explicit selection precedes old child/peer launch routes. All immediate
restoration candidates are prevalidated before startup, then resolved again during
construction. Invalid/missing offerings refuse without legacy fallback or file
repair. Launch records are not a second commit point for configured model changes.

Evidence remains /private/tmp/cadence-model-preferences.ERLGRh. Final focused checks
cover25 groups: selection6, child3, peer6, domain7, working-context3. The new real-file
scenario restores top-level/child/peer selections with the old launch offering absent,
preserves sibling policy, and records no inference. An unavailable saved choice with
a valid old route refuses; a valid-first/bad-later peer set leaves the fresh first
home uncreated. Null/duplicate/unavailable top-level preferences roll back membership.
Blocked settings replacement and corrupted live settings leave selection unchanged;
repair permits a later successful selection on the same Agent.

Initial tests failed because they still injected failure into child.json or expected
legacy wire-name validation despite a new explicit preference. The save-failure
oracle now targets settings.json. Legacy refusal cases explicitly omit the preference;
separate positive/negative tests establish precedence rather than weakening those
legacy contracts. The shared core suite passed371/0 on final runtime bytes, before
only the additional negative test assertions. A fresh clean build and all25 focused
groups passed afterward. The GUI rebuilt cleanly and its six existing regression
groups passed, including real peer HTTP/lifecycle paths and joined teardown; this
is not a fresh visual browser acceptance test. This is internal evidence, not
public-grader or live-provider acceptance. No installed settings, credentials, service or release artifact changed.

## Remaining integration

The current host registers only its launch credential binding for a model document.
Bill authorized ~/.en/settings.json as the global inference credential/default store;
its three existing API-key field names/types were inspected without printing values.
Host import and application-wide remembered funding are still pending. Keep the
existing transcription payer separate. Global defaults guide future explicit choices,
never silently retarget existing Agents. Host-approved Ollama configuration remains
explicit. Extend public Local Models checks and ordinary GUI evidence for these
contracts before claiming the full usability effort complete.
