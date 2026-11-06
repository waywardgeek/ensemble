# Chapter 27: Running Local Models

> **In development.** The learning brief and exercises below define the current
> implementation target, not completed acceptance. A Windows-native Ollama trial
> has demonstrated buffered text, structured tool calls and result consumption.
> Ensemble integration, streaming and lifecycle verification remain pending.

## TL;DR: Specification Brief

Add an explicitly configured local inference route to the existing Agent and
Engine. Offer its models in the ordinary model selector without changing the
current selection automatically. Local inference is another transport destination,
not another agent architecture.

- Select the model, endpoint, wire surface and funding policy together. A local
  model name must never travel to the cloud endpoint inherited from its parent.
- Use the existing OpenAI-compatible Chat Completions path where the server
  actually supports it. Do not build an Ollama-specific adapter merely because
  the server has a native API too.
- Keep local requests credential-free. Do not borrow an API key, subscription
  bearer or account header. Failure never authorizes a cloud fallback.
- Make configured local models available through normal model selection and
  child creation. Do not impose a child-only restriction. Selecting a small model
  does not make a large existing conversation fit it.
- Distinguish the model's advertised context maximum, the server's loaded context,
  and the Agent's working-context policy. Last-request usage is not a preflight
  token count or a guarantee that the next request will fit.
- Preserve canonical calls, call IDs and correlated results. JSON printed in
  ordinary text is not a tool call and grants no execution authority.
- Keep buffered and streaming behavior consistent across rendering, endpoint
  selection and response parsing. Preserve ordinary cancellation and ownership.
- Report measured token usage when present; missing usage remains unknown. Local
  inference is not proof of zero operating cost or a known cloud-equivalent price.
- Separate deterministic protocol tests from actual-model usefulness and human
  acceptance. A model listing's tool-support label is not an end-to-end test.

## In plain words

Moving inference to another computer does not move the Agent there. Ensemble still
owns the conversation, tools, Jobs, permissions and saved history. The model server
receives a request and returns an answer or a structured request to use a tool.
Ensemble decides whether that tool is granted and records its result before asking
for the next answer.

A model dropdown therefore chooses more than a spelling. If a parent uses a cloud
subscription and its child selects a local model, retaining the parent's endpoint
or bearer is wrong. The local route must be explicitly available, and selecting it
must select the correct destination without carrying cloud credentials across.
The same rule applies when a human changes an idle Agent's model in the GUI.

Small models are useful for bounded work, but a successful arithmetic example does
not show that they can manage a large system prompt, extensive memory and dozens of
tool declarations. Start with a small task and a modest tool set. Measure the loaded
context and GPU placement rather than assuming the advertised model maximum fits
in video memory. Raising a working-context setting does not enlarge the server's
allocation, and a handoff cannot remove all fixed prompt overhead.

The first live trial made a useful distinction. One coding model printed an object
that looked like a function call inside ordinary answer text. A newer package
returned a real `tool_calls` array, accepted a result with the matching call ID,
and answered the question. Only the second exchange demonstrated the structured
protocol we need. Neither exchange alone demonstrated reliable agentic coding.

## 27.1 Start with a verified server

The initial target is Windows-native Ollama on a separate laptop, not an assumed
WSL installation. Inspect the installed server version, GPU, model package and
loaded model before recommending changes. A package can require a particular
server version or parser. Parameter count, quantization, projector size and runtime
context all affect memory use.

Verify in layers:

1. A local version request establishes that the API server is running.
2. A small local inference request establishes that the model can answer.
3. The loaded-model report establishes GPU placement and allocated context for
   that configuration, not every possible request.
4. A structured tool call and correlated result establish a basic protocol round
   trip. No arbitrary model-supplied program needs to run for this test.
5. A version request from the Ensemble host establishes network reachability,
   not inference quality or authorization to expose the service publicly.

A saved listener setting is not a running listener. On Windows, changing a user
variable does not update an existing server process. Closing its window may leave
its tray process running. Inspect the listening address, stop the owning server
normally, relaunch it with the intended setting and verify again.

Ollama's local API should not be treated as an authenticated public service. For a
trusted-LAN experiment, restrict inbound access to the Ensemble host and the
required port. Inspect existing allowances: adding a narrow firewall rule does not
cancel a broader rule. Do not forward the port through the router or weaken the
Ensemble GUI's independent origin/authentication boundary. DHCP address changes
can invalidate address-based rules; a reservation or an authenticated tunnel is a
separate operational choice.

See the [Ollama FAQ](https://docs.ollama.com/faq) for current listener instructions
and its [OpenAI compatibility documentation](https://docs.ollama.com/openai) for
the supported protocol. These sources can change; verify the installed combination.

## 27.2 Integrate the existing owners

The Agent remains the owner of scheduling, configuration and History. Engine
consumes a frozen request and returns immutable results. The GUI projects the
selected route and available choices; it does not own credentials or a second
transcript. Child creation and restoration must resolve the same explicit route
as ordinary model selection.

Common now declares the configured graph: Ensemble strongly owns `ModelVendor`
objects and `ModelAuthentication` bindings; a vendor owns `ModelEndpoint` objects,
and each endpoint owns its `Model` offerings. The existing `Vendor` enum remains
wire/serialization compatibility vocabulary during integration. A configured
Ollama vendor is not the OpenAI vendor merely because its endpoint uses the
`ChatCompletions` surface.

An offering has a local configuration ID and a separate provider-facing model
name. The same wire model name through two endpoints denotes two offerings.
`ModelIdentity` qualifies that local ID with its vendor and endpoint IDs. It is a
credential-free value, not another owner or a provider-facing routing name.
Provenance can retain this identity alongside the exact wire model and surface:
opaque material from a different offering must not match merely because both
servers use the same model name. Legacy provenance remains readable but cannot
match an explicitly configured offering. Invalid present identities refuse rather
than silently becoming legacy data. Request projection and response capture carry
this value; configured construction and child restoration attach the live Model
relationship. The public Model-object selector changes that relationship through
the existing Agent mailbox; the ordinary dropdown submits the same qualified
identity.
Endpoint addresses exclude embedded credentials, query strings and fragments.
New credential-bearing endpoints require HTTPS except literal loopback fixtures;
public OAuth and Codex retain their additional existing endpoint restrictions.

Inference transport does not automatically follow redirects or retain response
cookies. Otherwise the first funding mode could leave hidden routing or cookie
authority on an Engine later used for a different model. A redirect is an error
at the configured endpoint, not permission to send the request elsewhere. Public
OAuth's existing bounded401 renewal remains on the same frozen endpoint, account
and request body; it is not a general inference retry or funding fallback.

An authentication binding is an explicit choice: no credential, an owned API key,
a reference to the public OAuth credential owner and account label, or a reference
to the independent Codex credential owner. It does not copy a renewable token into
a Model. Endpoints and bindings maintain both ends of their weak relationship.
Destroying a credential owner detaches its bindings; replacing that owner does not
silently reconnect them. Destroying a binding invalidates dependent endpoints
without destroying their model objects or changing their funding kind. Ordinary
root destruction finishes Agent work before removing models and credential owners.

Constructors validate before publishing relationships. Duplicate IDs within an
owning collection refuse without replacing an existing endpoint or model. A
membership snapshot is not an editable collection. Removing a relationship clears
the back-reference without destroying the detached object; destruction recursively
removes strong children. These operations are tested separately from transport.

The internal request projection freezes the offering's wire name, surface,
capabilities and serving context while preserving Agent instructions and security
policy. It references the selected authentication binding instead of copying its
API key. Explicit unauthenticated funding emits no credential headers; it is not
an empty-key attempt or permission to discover cloud credentials. Streaming and
media checks use the frozen capabilities rather than guessing from a model name.
A Model also owns an immutable, credential-free `ModelRequestPolicy`. Reasoning,
streaming suppression, explanation enforcement, deferred tools, result stubbing
and cache options come from that selected offering, not the previously selected
route. Instructions, tool-round limits and security boundaries stay with the Agent.
Model construction validates options against its surface and capabilities before
adding endpoint membership. Feature-required explanations cannot be disabled by
policy. Request projection turns an explicit reasoning summary `off` into an
omitted provider field; returning to the offering restores its declared options.

Default working-context limits use the configured serving allocation when known.
An explicit policy also includes the qualified offering identity. Two endpoints
serving the same wire model may have different allocations, so their policies
must not overwrite each other. Legacy policies keep their original wire-name
identity and do not silently apply to newly configured offerings. Settings
round trips preserve that distinction; malformed identities refuse rather than
becoming legacy rows.
Recall and compression preserve the binding; an unresolved judge-model override
refuses rather than silently changing the route.

Ensemble resolves a saved `ModelIdentity` by following its current vendor,
endpoint and model ownership. There is no second lookup registry and no fallback
by wire name. A detached or missing offering is unavailable. Ordinary child
registrations retain the legacy version1 shape for legacy configurations;
configured children use version2 with a required qualified identity. Restore
validates identities before constructing children, and checks the saved wire name
against the resolved offering. It preserves conversation and pause state, not
credentials or interrupted effects. A legacy child record under a configured
parent refuses because that record cannot identify the original route safely.

Peer launch records follow the same identity rule. A configured peer saves its
qualified offering alongside the launch model name; a legacy peer keeps the
original four-field record. These are launch choices, not a second store of the
running Agent's mutable selection. Restoration resolves the entire launch set
before starting any peer, checks each wire name, and refuses missing offerings
without substituting the host's route or funding. Invalid present identities do
not become legacy records.

A configured Agent holds a weak selection link to its actual Model, and the
Model maintains the reverse membership. Agent construction attaches both ends
before creating Engine; partial construction rolls them back. Destroying a Model
clears these weak links without destroying Agents or their conversations. A lost
selection refuses new rendering and inference, never reverting to a legacy route.
Engine admission validates the graph before incrementing its active-send guard.

`SetModelAsync(Model)` selects a configured offering only between turns, including
absence of direct Engine sends. It rejects a foreign root before acquiring its
graph lock. Under ingress and the owning graph gate, it validates the candidate,
saves the complete child restart selection when applicable, then updates both
membership ends and the frozen configuration. Failed saving leaves the old
selection intact. A successful switch retains Engine and conversation, and resets
the last-request context measurement. An explicit valid choice can recover an
Agent whose old Model was removed; no automatic substitution occurs. The legacy
string overload still preserves its route and cannot invent a configured offering.

The GUI reads a credential-free `ModelSelectionStatus`, not the inference-validating
configuration getter. Under the same ingress-to-graph lock order, this snapshot
lists configured offerings and their local policy compatibility, funding description
and serving context. Availability is not a network health check or proof of account
entitlement. Unknown serving context remains unknown. Legacy name-only choices
retain their fixed-route semantics; configured choices carry a separate strict
identity object, never a renamed provider model.

An unavailable selection remains visible after reconnect. Even recreating an
offering with the same IDs does not reconnect a lost Agent link automatically.
The picker shows a disabled current-selection placeholder and permits an explicit
valid replacement. The server resolves the chosen identity again and calls the
same Model selector; displayed status is never a send or selection grant.

Documentary persistence has a different purpose from inference admission. It can
save the last frozen configuration and retained conversation after a Model is
removed, without acquiring credentials or granting another request. During root
shutdown, closing admission must not block the Agents' final History saves.

**Integration boundary:** synthetic HTTP checks exercise the internal projection,
including the exact headers and captured provenance. Configured child registration
and restoration, weak selection membership and saving after Model loss have
focused coverage. Per-offering working-context persistence and effective-policy
isolation and qualified peer launch/restoration have focused coverage too.
Public Model switching has focused loopback routing and failed-save coverage.
The ordinary dropdown has real-socket and actual-page offering-switch/recovery
coverage. Host file loading and explicit launch-return routing have process/socket
coverage. Configured OAuth/Codex renewal and streaming conversations still need
their own real-path coverage. Do not infer an
installed MiMo offering or live-model acceptance from these owned fixtures.

### Host-owned model configuration

The host, not the model picker, controls endpoint and authentication configuration.
Load its strict document before constructing Agents or restoring the fleet. GUI
selection uses only the resulting approved graph. Do not put credential values or
token-file paths in this document: each endpoint names an authentication binding
already constructed by the host under the same Ensemble root.

`Ensemble.LoadModelConfiguration` accepts at most 1 MiB of strict UTF-8 JSON. Version 1
has `version` and `vendors`; each vendor has `id`, `label` and `endpoints`. An endpoint
has `id`, `address`, `surface`, `authentication` and `models`. Surface spellings are
`messages`, `chat_completions`, `responses` and `generate_content`. Each model has
`id`, `name`, `label` and `features`, with optional `serving_context_tokens`
and `request_policy`.
Features require an explicit `stream` boolean. Optional media and reasoning lists,
context capacity and policy flags default conservatively, not from name guesses.
Advertised `context_window` and serving capacity remain separate token values.

The optional `request_policy` object accepts boolean `disable_streaming`,
`require_tool_explanation`, `deferred_tools`, `stub_tool_results` and
`cache_breakpoints`, plus string `prompt_cache_key`, `reasoning_effort` and
`reasoning_summary`. Omitted booleans are false. Empty reasoning options choose
validated `medium` and `auto` defaults only for a Responses offering declaring
reasoning support; otherwise they remain absent. Summary `off` explicitly omits
that field for a supported offering. Unsupported options refuse instead of being
silently ignored. Deferred tools and block-cache options require Messages; prompt
cache keys require an OpenAI-compatible surface. These are explicit host assertions,
not provider discovery or guarantees that caching will occur.

Unknown or duplicate fields, invalid types and unresolved bindings refuse. The
whole document is bounded to 32 vendors, 64 endpoints and 256 models. Parsing precedes
graph publication; construction and rollback hold the existing graph gate. A late
failure removes every newly created subtree and detaches its binding memberships,
without destroying the borrowed authentication owners. Existing vendors and Agent
selections cannot be overwritten by a load. This is additive startup configuration,
not a live graph editor or an implicit model switch.

The GUI host accepts `--model-config PATH` or `EN_MODEL_CONFIG`; the command-line
option wins. Relative paths resolve after the host's explicit workspace option.
It opens one file handle, bounds both initial size and subsequent reads to 1 MiB,
and loads the graph before constructing the initial Agent or restoring the fleet.
Missing, malformed or unresolved configuration refuses startup; there is no empty
configuration fallback. Omitting the option preserves the legacy launch path.

The host supplies two binding names when loading a document:

- `none` is explicitly unauthenticated, suitable for a restricted local endpoint.
- `launch` uses the payer selected by the existing launch options: the supplied
  API key, the actual public OAuth owner and account, or the actual Codex owner.
  An API launch without a key does not create this binding. `--billing none`
  makes it explicitly unauthenticated instead; an ambient API key is ignored.

The document cannot name another token file or discover another account. Loading
also does not select an offering or reconstruct the initial Configuration: launch
reasoning, caching, funding and other policy remain unchanged. Choosing a configured
offering subsequently uses the normal configured-model projection.

When the launch binding exists, the host reserves vendor `host-launch`, endpoint
`launch` and model `launch` for the explicit **Return to launch** choice. It uses
the actual launch endpoint, binding, features and immutable request policy. A file
cannot replace that reserved vendor. A keyless API launch has no return offering;
it does not acquire an unauthenticated substitute. Initial selection is still
legacy. Explicit return uses the new qualified identity, so old opaque legacy
provenance does not automatically become eligible for replay. Conversation remains,
but identical routing alone is not permission to migrate opaque provenance.

The library boundary has strict-decoding and rollback tests. The actual host has
process/socket checks for loaded choices, retained initial selection and funding,
flag precedence, explicit no-auth, missing funding and invalid-file refusals before
Agent construction. Run the scoped check against an already-built GUI executable:

```sh
python3 agent/GuiTests/model-host.py ABS_DOTNET ABS_GUI_DLL NEW_EVIDENCE_DIRECTORY
python3 agent/GuiTests/model-launch-wire.py ABS_DOTNET ABS_GUI_DLL ANOTHER_NEW_DIRECTORY
```

These POSIX checks use isolated homes and synthetic keys and join their hosts.
The startup check makes no inference request. The wire check makes exactly three
requests to its owned loopback provider and deliberately returns buffered JSON.
Neither certifies subscription renewal, SSE replies, restored fleet execution,
live Ollama behavior or installation.

The construction-point changes belong alongside configuration and transport,
model selection, child ownership, persistence and accounting. This capstone joins
those contracts. The independent scored grader
commands will be documented here after implementation and review; no unimplemented
command or public grader is promised by this draft.

## Exercises (draft acceptance contracts)

These exercises are not yet certified against the implementation. Use isolated
homes and owned loopback fake providers for deterministic checks. A real local
model trial is a separate exercise, not a substitute for those checks.

### Exercise 1: Select a route without borrowing funding

Configure a cloud route and one explicit local route. Select the local model in
the normal GUI and for a child Agent. Verify the exact destination, model and
Chat Completions surface at an owned fake server. Assert that no cloud credential
or account header reaches that server and that the parent's selection is unchanged
by its child's choice. Refuse selection during admitted work without changing
configuration. Verify restoration resolves the intended route; an unavailable
route must not silently become a cloud request.

### Exercise 2: Complete the tool conversation

Return a structured function call with a unique ID. Execute only an explicitly
granted harmless tool, record its result, and send that result with the matching
ID. Require the final answer to consume the result. Repeat with a tool-shaped JSON
string inside ordinary content and prove that no tool executes. Exercise buffered
and streaming responses with fragmented argument text; fragments cannot authorize
execution before the call is complete.

### Exercise 3: Fail and cancel honestly

Use controlled server barriers to interrupt an in-flight request and destroy its
owning Agent. Verify cancellation and joined cleanup without replaying tool effects.
Exercise connection refusal, malformed responses and server shutdown. Preserve
recorded outcomes and report failure without contacting a fallback provider. A
quiet network wait is not evidence of completed work.

### Exercise 4: Measure limits and usefulness separately

Check reported input/output usage, missing usage and unknown pricing with fixtures.
Then, with explicit permission, run one bounded real coding task on the local
model. Record package/server versions, loaded context, GPU placement, correctness
and practical latency. Include the prompt and tool overhead in the experiment;
do not infer capacity from parameter count or extrapolate an arithmetic success
into general coding competence. Report deterministic checks, real-model results
and human acceptance separately.

Junien's [PR 3](https://github.com/waywardgeek/ensemble/pull/3) supplies the
sleeping-inference-machine use case, and
[PR 6](https://github.com/waywardgeek/ensemble/pull/6) highlights explicit streaming
policy. Wake-on-LAN remains optional and separately authorized.

The [closeout plan](../docs/edition-3-notes/first-book-closeout-plan.md) separates
synthetic verification, real-target trials and human acceptance. This chapter does
not authorize model downloads, listener changes or an inference campaign merely
by being loaded as task material.

Next: [Crossing Over to Ensemble](chapter-28.md).
