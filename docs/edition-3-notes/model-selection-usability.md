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

## Global settings persistence foundation

GlobalSettings is a strongly owned application object with explicit path input;
read-only operations never create files. It binds each configured synthetic API
key through the existing ModelAuthentication owner and exposes only nonsecret
vendor-to-binding references in model_funding_defaults. Funding resolution uses
the real vendor/endpoint/model graph. Remembering a choice changes no Agent.
Missing/ambiguous/unavailable saved funding refuses without selecting another key;
an unrelated endpoint with a removed binding does not hide a valid saved choice.

Global writes reread current settings under a fail-fast exclusive lock file,
preserve unrelated JSON fields and credential values, flush a private temporary
file, then replace. This is cooperative-writer serialization, not confinement,
arbitrary external-writer protection, directory-fsync crash recovery, or a
transaction with an Agent settings file. The application destroys the settings
owner after Agents; destroying settings alone does not destroy credential bindings.

The clean final build passes two new global-settings groups and all six model
selection groups, plus seven existing model-domain groups, in the existing evidence
directory (global-settings-final-* and global-settings-domain.log).
Synthetic-only tests cover ownership/teardown, distinct vendor keys, preserved
external sibling updates, contention and directory-shaped-path refusal, fresh-owner
persistence, defaults not retargeting an Agent, unavailable-binding refusal/recovery,
and absent/malformed/duplicate/oversized/invalid-UTF8 input. Internal key inspection
confirms synthetic binding contents; this is not public request-header evidence.
No installed global settings were read or written during these tests.

Initial build failed because a new directory was outside explicit compile globs;
implementation moved into Application. Initial refusal helper omitted the distinct
InvalidDataException type. Review found the same omission in recoverable Agent
selection errors; it is now handled, with an actual damaged-document/refusal/repair
check on the same live Agent. A source-review lookup correction excludes unrelated
missing bindings without making the remembered missing binding fall back. No
pre-fix behavioral credit is claimed for that review-only correction. No new shared
suite, host/browser, public-grader or live-provider acceptance is claimed here.

## Host and GUI funding integration

The host now loads the root-owned global store before Agent construction and
fleet restoration. Explicit --global-settings wins over EN_GLOBAL_SETTINGS,
then ~/.en/settings.json. Curated OpenAI, Anthropic and Gemini API offerings use
stable graph identities and exact capability lookup; absent keys add no offerings.
Startup does not probe providers, rewrite settings or retarget the launch Agent.
Configured Ollama remains explicit and separately grouped. Sixteen isolated
startup/socket cases passed before the context investigation, including override,
malformed global settings and reserved-vendor refusal. Provider-name dots cannot
serve directly as graph IDs; stable underscore IDs preserve exact wire names.

Both GUI controls distinguish an explicit offering from a vendor/model preference.
The preference resolves through GlobalSettings at action time, never from cached
browser routing. Missing or ambiguous saved funding refuses without fallback.
Explicit selection/creation remembers funding only after Agent success. A failed
global write reports partial success without rolling back or misreporting the
Agent operation. The files are not one atomic transaction.

Final parent job38 built cleanly and passed the configured socket fixture and six
GUI regression groups. Actual HTTP creation checks cover preferred creation,
missing-funding refusal, conflicting forms and locked-global-file warnings with
durable peer state and unchanged global bytes. Socket checks cover independent
Agent/default state, default selection, refusal and explicit recovery. The new
funding-save assertion failed before its handler change, then passed afterward.
No inference was submitted.

The actual model-picker page passed ambiguity, explicit/default commands and
unavailable reload/recovery; browser and host exited zero without forced cleanup.
Evidence is in /private/tmp/cadence-model-preferences.ERLGRh/funding-page-s99l4q1q.
The Create Agent dialog passed desktop/phone preferred payload, disabled default,
inheritance, failed-load retry and partial-success display against an owned HTTP
substitute. Real peer-route behavior is the separate C# check above, not a claimed
end-to-end browser creation proof. Its first added run failed because a broad
Unavailable locator matched two different options; scoping it to explicit options
fixed the fixture without weakening the separate missing-default assertion.

The initial source receipt funding-gui-source-receipt.json is post-execution,
not a predeclared audit. Review corrected a disabled preference's legacy-route
tooltip. A clean rebuild and actual-page check of the final embedded assets passed
with an explicit tooltip assertion; browser and host both exited zero without
forced cleanup in funding-page-final-k5c6yaw_. The final post-execution receipt
funding-integration-final-receipt.json records the reviewed source hashes and the
separate test scopes. Hard chapter/companion/note lint passes with advisories.
At this host/GUI checkpoint, public preference checks were the next slice; the
following section records their completion. Installed settings, provider entitlement and human usability remain separate gates. No
installation, restart, private credential output or live provider call was performed.

## Public preference and process-restart checks

The existing Local Models runner now embeds PreferenceChecks.cs and requires 126
main checks plus 4 writer and 6 reader checks: 136 total, still ungraded/exit3.
Fifteen public-API checks use actual isolated files for independent Agent/default
state, ambiguity, remembered binding, unrelated key/settings preservation, writer
lock refusal, sibling settings, failed Agent save, and fresh-root restoration.
A usable global choice cannot replace an unavailable saved Agent preference.

The process-separated writer saves a child on a new offering, remembers another
global default, and preserves a deliberately stale launch record. The reader
omits the old launch endpoint, yet restores both Agents' saved selection and the
paused child conversation without replayed inference. This is clean restart, not
crash recovery, full fleet coverage or browser restoration certification.

Two fixture assumptions were corrected, not production behavior: the original
child-negative test first refused the parent's newly saved offering, so it now
explicitly makes only that parent legacy before testing child refusal. The first
new writer expected child.json to stay stale through later pause/shutdown saves;
it now captures the original paused launch and deliberately restores those bytes
after shutdown, leaving the real Agent settings untouched. Initial failed receipts
remain in parent jobs67/72; final packages and real integration pass in job74.

The actual CLI good report is public-preferences-good.json under
/private/tmp/cadence-model-preferences.ERLGRh, with actual exit3. A separate
successfully built source copy changes only remembered binding equality to
inequality; public-preferences-wrong-funding.json reports exit1 at exactly
preference-resolves-remembered-binding, after its earlier guards pass. Mutant
source/build is /private/tmp/cadence-funding-mutant.tiw6c_w9. The good runtime was
not overwritten. No every-assertion mutation coverage or new chapter points are
claimed. Original chapter and consumer instructions describe the expanded scope.
