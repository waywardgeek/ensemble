# Chapter 17 validation

Preparation reviewed, October 8, 2026. Outline/research `580ce83` and independent
advisory `b59c47e` are complete. The coordinator has read both preparation files
and the full advisory. The author acknowledges Bill's continuing requirement
that MCP remain transport-independent and that the actual GUI tunnel live in the
optional module. No Chapter 17 implementation, full contract or experiment is
accepted. Chapter 16 has contract/prose acceptance only.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Research and consequential choices | Author, reviewer, coordinator, Bill for the output-cap choice | Preparation `580ce83`, advisory `b59c47e`; Bill approved explicit subscription mode without a per-response output-token cap | Specify complete supported route and per-purpose compatibility before dependent code |
| Full contract and prose | Author `/root/author`, independent reviewer | Partial draft `b9772c3`; scoped hard prose checks pass, full manuscript review pending | Resolve route-dependent material and complete v7/public schemas before full contract review |
| Independent checks | Future grader | Not started | Derive distinguishing checks from complete published teaching |
| Fresh student and owner plan | Future fresh student, reviewer | Not released | Requires accepted Chapter 16 source and complete contract |
| Local implementation | Future student, grader | Not started | Preserve preceding ownership, replay and helper guarantees |
| Actual use and OAuth caching investigation | Future student, reviewer, coordinator | Not started | Review finite request matrix, own-application authorization and comparable controls |
| Historical comparison and revisions | Independent reviewer | Not started | Freeze initial implementation, runs and student experience first |
| Feedback, final prose and checkpoint | Author, student, reviewer, coordinator | Not started | Reconcile actual results before export/tag |

## Working decisions for independent drafting

These are coordinator design choices, not new statements attributed to Bill.
They accept the advisory's D1–D5 and D7 direction and the author's response.

1. Engine owns the diagnostic child. Observe actual prepared request bytes at
   transport handoff after admission, with separate foreground, compression and
   judge streams. Retain at most two 8 MiB bodies per purpose, 48 MiB per Agent.
   Clear the relevant baseline on route/model/funding-connection generation
   change; configuration edits within that route remain comparable. Oversize
   attempts create explicit gaps rather than comparisons across a missing turn.
   Export exact immutable pairs with hashes and a manifest published last.
   The draft must separately bound parsing, temporary copies, concurrent export
   work and metadata; the retained-body number alone is insufficient.
2. Separate raw prefix bytes, structural append/edit classification and
   directive-aware comparison. Preserve original byte coordinates and lexemes.
   Derive expected results independently from literal fixtures, including the
   historical compact-versus-indented failure. Do not infer backend hits from
   local byte equality.
3. Use a small provider-specific policy, preserving existing default/off bytes.
   Select Messages manual placement within four slots, with no simultaneous
   automatic placement; select Chat Completions implicit-plus-markers within
   three explicit slots. Gemini retains implicit caching. The author must
   verify exact supported wire shapes and capabilities before publishing them.
   No explicit-only mode or remote Gemini cache-resource lifecycle is added.
   Respect the earliest transient material in a purported stable prefix;
   unchanged later blocks do not make an earlier one-shot hint stable.
4. Use an explicit v7 rendering capability and preserve exact v1–v6 routes.
   Capture sufficient policy for all affected request purposes before rendering.
   Metadata-only displays and prices do not create a second durable history.
   Retired body data stays retired, including diagnostic copies; unavailable
   reconstruction remains unavailable. The draft must print the complete
   initializer, identity, policy and record-class rules before any checker.
5. Keep durable, mounted-run, purpose and last-accepted usage views distinct.
   Use exact rational arithmetic and lossless browser counters. Price observed
   producing identities against a small dated table, with visible unknown rates,
   incomplete usage and partial estimates. A failed current attempt cannot appear
   to have the preceding accepted response's usage. No invoice or subscription
   charge is inferred from this estimate.
6. Publish one copied public diagnostic view for CLI, optional GUI and a real
   two-Agent consumer. Defer an automatically installed model-facing status tool;
   existing granted GUI inspection can observe the display. Bound wire values and
   specify reconnect/mount lifetime before implementation.

## Subscription route: Bill's decision

The coordinator independently searched and opened the official OpenAI pages on
October 8, 2026. The [preview limitations](https://developers.openai.com/siwc/token-sharing-open-source/preview-limitations)
exclude `max_output_tokens` and require streamed responses on this plan route.
The [inference guide](https://developers.openai.com/siwc/token-sharing-open-source/models-and-inference)
requires account-specific model discovery and the public Responses endpoint.
These are current documentation findings, not account-access or cache measurements.

Bill agreed to an explicitly selected ChatGPT subscription mode without the
per-response output-token cap and instructed the team to continue. Retain finite
request-count limits, timeouts and usage reporting, and state explicitly that
these do not guarantee a token or charge ceiling. Existing capped API-key
behavior stays intact. The author must teach and capture each intentional
per-purpose difference for the selected plan route, including streamed helper
delivery and absence of the remote output cap. Local result-size, validation,
deadline and request-count limits remain applicable; they cannot be described
as equivalent remote generation limits. No silent credential/funding fallback
is authorized by this design choice.

The author may now complete the route design and its compatibility contract.
Do not release dependent implementation, authenticate, or run a separate probe
and label it an Ensemble success. Bill's requested cache retest remains an open
chapter obligation. No drafting decision here postpones or waives it. Application
authorization, credential ownership/renewal and the complete Responses contract
must be explicit before actual integration or use.

Partial manuscript `b9772c3` teaches the independent diagnostic, marker and cost
mechanisms, with literal fixtures. It expressly leaves the complete capability,
initializer and public-command schemas unfinished. The coordinator's scoped lint
receipt is [prose-ch10-clarification-ch17-partial.json](checkpoint-evidence/prose-ch10-clarification-ch17-partial.json).
Hard rules pass; soft density and long mechanism stretches remain editorial
review inputs, not proof that the human-story requirement is met. No full draft
acceptance or provider experiment follows from that lint result.
