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
this value; selecting and restoring the live Model relationship remain pending.
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
Default working-context limits use the configured serving allocation when known.
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

**Integration boundary:** synthetic HTTP checks exercise the internal projection,
including the exact headers and captured provenance. Configured child registration
and restoration have focused coverage, but live Agent selection membership, peer
restoration, per-offering working-context persistence and GUI configuration remain
open. Configured OAuth/Codex renewal and streaming conversations still need their
own real-path coverage. Do not infer a working local dropdown or MiMo acceptance
from these internal tests.

The construction-point changes belong alongside configuration and transport,
model selection, child ownership, persistence and accounting. This capstone joins
those contracts. The user-facing configuration interface and executable grader
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
