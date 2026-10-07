# Chapter 1: One Conversation, Every Request

Bill Cox's first coding agent was a two-week bet. The first edition records
the July 2025 demonstration of StackAgent, followed by a decision to start
again with what building it had taught him. The useful part of that story
fits on a workbench: build something small enough to understand, run it,
and make its failures visible before adding machinery.

This exercise starts with a conversation you can inspect one request at a
time. Your program sends the messages, keeps the answers, and reports the
tokens consumed. Its package boundaries and ownership rules apply from the
first implementation, while the program is still small enough to inspect.

## 1.1 The idea in plain words

An agent needs both a place to keep facts and a way to reach them. A model
name belongs to an agent's configuration. The code sending a request needs
that name. A tool reporting status needs it too. Copying the name into
every interested object creates several answers to a question that should
have one answer, especially when the user changes the model.

**Give each fact an owner.** Keep configuration on the Agent. Keep usage
accounting with the Engine that receives the model's responses. A child
holds an interface to its parent, and each parent exposes a route to its
own parent. Code follows that chain to the owner of the information it
needs. Adding a consumer does not require another copy of the information.

**Keep the interfaces together and the work apart.** Shared declarations
tell the parts of the program what they can ask of one another. Separate
packages do the work: model requests, job supervision, tool execution.
Each package imports the shared declarations. It calls across a boundary
through an interface, without importing the other implementation.

**Make logging reachable before something breaks.** A parser can have no
stored state and still need to explain which response it rejected. Give
code likely to need debug logging access to an owned object from which it
can reach the logger. Calling a function stateless does not remove that
need. Logging follows the same ownership route as other shared facilities.

These rules prevent a common form of plumbing. A new feature needs a value
that another object already owns; its constructor gains another parameter,
several callers gain the same parameter, and a closure carries the value
across the final gap. Repeat that for settings, usage, and logging, and a
small feature changes construction throughout the program. A parent
interface gives the child a continuing route to its owner's facilities.

## 1.2 Packages follow responsibilities

The dependency pattern is a star. The shared package, `internal/common`,
is its center. It holds the core data structures, constants, and
interfaces. Implementation packages form the spokes. Each spoke imports
`common` and the standard library; it does not import a sibling spoke.
The library's composition root assembles the implementations, and the
command-line program enters through the library's public API.

Directories describe responsibilities. `internal/llm` contains model
request construction, response parsing, and the engine's work.
`internal/jobs` contains job supervision when jobs are introduced.
`internal/tools` contains tool behavior when tools are introduced. Shared
conversation data and the interfaces connecting these responsibilities
belong in `internal/common`. Create the packages needed for actual work;
an empty directory does not establish an architectural boundary.

Go restricts where methods can be declared: the receiver's base type must
be defined in the same package. This is the language's
[method-declaration rule](https://go.dev/ref/spec#Method_declarations).
If a shared conversation type lives in
`common`, an ordinary method on that type must also live there. Following
that pressure would gradually move request construction and parsing into
the shared package just to preserve method-call syntax.

Use a free function instead. An operation that sends a conversation to a
model belongs in `internal/llm`, taking the shared data as an argument.
Its home follows its responsibility. The language permits this directly;
there is no need to put the operation beside the data declaration.

There is a narrow reason to retain methods on shared types: satisfying
standard-library interfaces. `MarshalJSON`, `UnmarshalJSON`, `String`, and
`Error` need method receivers for the relevant interface dispatch. Keep
those methods with their types. They do not justify moving unrelated
engine behavior into `common`.

When several spokes need the same behavior, declare its interface in
`common` and implement it in the package responsible for that service.
Expose the service through the ownership chain. Do not make every caller
import its implementation, or fill `common` with shared function bodies.

## 1.3 Follow the owner

A constructor receives an interface to the object that creates and owns
it. That interface is declared in `common`, along with the interfaces for
the facilities reachable through it. Store the back-pointer. When a child
needs its owner's configuration or another owned service, it asks through
that interface. Reaching a higher ancestor means following the parent
chain, rather than storing an additional shortcut to the root.

An interface limits what the child can do while preserving that route.
It exposes operations and accessors, without exposing a concrete parent's
fields. A constructor's inputs can describe the object being created;
they must not become a bag of logger, configuration, registry, and usage
references copied from objects that already own those facilities.
Closures added at the wiring site to recover those missing routes are
the same mistake in another form.

Mutable state belongs to objects. Do not use package-level variables for
the current agent, configuration, conversation, counters, logger, or tool
registry. Even a single-agent program must have explicit ownership.
One application-level owner, called Ensemble here, can own many Agents.
That single runtime owner is constructed explicitly; it is not a global
variable that any package can read.

For this implementation, Ensemble owns the logger, Agent owns its
configuration and conversation, and Engine owns its transport and usage
counters. Engine reaches configuration through Agent and logging through
Agent's parent. The CLI constructs Ensemble through the public library
API and asks it to create an Agent; it does not assemble a parallel set
of engine, logger, and configuration objects itself.

The runtime owner and the shared package have different jobs. The word
*hub* in the star diagram means the package containing shared declarations.
The runtime Hub, or Ensemble, is an object that coordinates agents and
their communication. An Agent holds an interface back-pointer to that
owner. Naming a package `common` creates neither an owner nor a route to it.

Tool visibility belongs to each agent. Registry storage may belong to
each Agent or to Ensemble with access filtered by agent. Either arrangement
must prevent a tool visible to one agent from silently becoming available
to every other agent. The storage choice does not change the visibility
requirement or permit a mutable global registry.

## 1.4 Events, requests, and the optional GUI

Ensemble receives streaming observations through the Observer pattern.
An Agent publishes events such as arriving text or a changed status; its
observer receives them as they happen. The application owner mediates
delivery to the GUI or an external gateway. Publishing an observation
does not require the Agent to know which screen or transport displays it.

A request has different semantics. Asking the owner to create a sub-agent,
send a message, or provide access to an owned facility can use an explicit
method on the parent interface. Observer is the route for live events;
it is not the only permitted communication with the parent. Keep concrete
coordinator details behind those interfaces in both cases.

The GUI belongs in a separate, optional Go module. Its WebSocket server
is a GUI implementation detail, outside `agent/internal`. The agent
library must work in a headless application without importing that GUI
module. A user's application can select both modules, or use the agent
library alone.

Starting in Chapter 2, Ensemble serves two client interfaces: a CLI and
a browser GUI connected through WebSocket. Both address the same
application owner. A GUI can
begin as a stub while the CLI exercises the real agent; a second client
does not create a second place to construct and coordinate agents.

Reusable GUI components, including WebSocket, Artifact, and Connector,
need public interfaces so another application can compose its own UI.
The GUI consumes the agent library's public, transport-independent
interfaces. It must not reach into private agent packages. Moving the
GUI to another directory inside the same required module does not
establish this boundary.

The conversation exercise does not implement a GUI, tools, or sub-agents.
It does establish the ownership and dependency rules under which those
features can be added. Its first request belongs in the responsible
package, with access to its owned data and logger through the interfaces
just described.

## TL;DR

Build a Go program using the standard library and raw HTTP. No SDK or agent
framework. Apply the architecture in §§1.1–1.4 from the first implementation:
shared structures and interfaces in `internal/common`, model behavior in
`internal/llm`, and access to owned data and logging through interface
back-pointers. Ensemble owns the logger, Agent owns configuration and
history, and Engine owns HTTP transport and usage. Engine's parent is
Agent; Agent's parent is Ensemble. Constructor and accessor names are
yours. Put the public library at the module root and the executable in
`cmd/`. This text-only conversation uses the small representation below;
Chapter 2 introduces the provider-neutral event structures without
discarding the ownership and package architecture.

```go
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Conversation []Message
```

1. With no arguments, read one `{"user":"..."}` object per stdin line.
   Input questions are nonempty strings. For each question, emit exactly
   one `{"assistant":"..."}` object on stdout. After clean EOF, emit
   `{"usage":{"input":i,"output":o}}` with cumulative token totals and
   exit 0. An empty input stream reports zero totals. Nothing else goes
   on stdout. Diagnostics go on stderr. Ignore blank input lines;
   malformed JSON or an empty question terminates nonzero.
2. Read `ANTHROPIC_API_KEY` and `ANTHROPIC_MODEL` from the environment.
   Either missing or empty means exit nonzero before any HTTP request,
   with empty stdout and a diagnostic naming the missing variable. The
   model diagnostic must point to `GET /v1/models`. No fallback model ID.
   Read `ANTHROPIC_BASE_URL`; when empty, use
   `https://api.anthropic.com`. Remove trailing slashes before appending
   `/v1/messages`.
3. Make exactly one `POST` per question. Set `x-api-key` to the configured
   key, `anthropic-version` to `2023-06-01`, and `content-type` to
   `application/json`. Send JSON with the configured `model`, positive
   integer `max_tokens`, nonempty fixed `system` string, and `messages`.
   Omit `stream` or set it to false.
4. Each request contains all completed user/assistant pairs followed by
   the new user question. Roles alternate, beginning and ending with
   `user`; content is nonempty. Preserve every earlier role and text
   exactly. The next request adds exactly the preceding assistant reply
   and the new question. These are exercise rules, not universal limits
   of the Messages API.
5. Decode a successful response's `content` array. Concatenate its text
   blocks in order, inserting nothing between them. The fake splits each
   reply across multiple blocks. Read `usage.input_tokens` and
   `usage.output_tokens` and add them to running totals from the first
   response. Return the assembled answer and retain it in history.
   A valid response has a nonempty assembled answer and a `usage` object
   with nonnegative integer input and output counts; reject one that
   lacks these fields instead of manufacturing an answer or accounting.
6. On a transport error, non-200 response, or malformed response JSON,
   report an error on stderr and exit nonzero. Do not retry, invent an
   assistant answer, or emit a success usage record. Apply this rule in
   both modes if you add interactive chat. A rejected response terminates
   the same way. Commit a user/assistant pair only after a valid response.
7. Give each piece of mutable state one authoritative owner. Configuration
   belongs on Agent; usage belongs with Engine. Declare parent interfaces
   in `internal/common`; constructors establish the ownership chain.
   Keep model behavior in `internal/llm`, using free functions on shared
   data where necessary. Both front ends enter through the same library
   API. The optional `chat` mode changes input and presentation only.
   The CLI imports the public library, never an implementation package.
8. Use a finite HTTP timeout. Let an external consumer supply Ensemble's
   logger output writer. Request failures must reach that writer through
   the parent chain; diagnostics must not include credentials.

**Yours.** Internal names and function boundaries, system-prompt wording,
positive output-token limit, and optional human interface, within the
architecture above. No tools, streaming, persistence, retries, model switching,
or dollar-price calculation belong in this exercise.

**Exercise.** Create a self-contained Go module and Git repository with
the library at its root. The course's new reference snapshot lives at
`solutions/edition-2/ch01`; your own directory works equally well. Build
`./cmd`, then run the free local grader from the course repository root:

```sh
make grade-dir CH=1 DIR=path/to/your/solution
```

The existing first-edition grader awards 100 points across seven checks:
`protocol` 15, `wire` 15, `calls` 10, `replies` 10, `memory` 25, `growth`
15, and `usage` 10. It does not establish compliance with the architecture
above. The additional required properties and their acceptance procedure
are described in §1.8. A passing protocol score alone is insufficient.

## 1.5 The request carries the conversation

The Messages API accepts a list of messages and generates a response.
A conversation uses that same endpoint repeatedly, with earlier turns
included in each request. The program owns the history it sends.
Stateless requests do not imply anything about a provider's data-retention
policy; they describe how the client supplies conversational context.
See the [Messages API reference](https://platform.claude.com/docs/en/api/messages/create).

The first request contains one user message. The second contains that
question, the assistant's answer, and the next question. The answer is
part of the evidence supplied to the next request. Keeping only the
questions loses facts the model introduced itself.

The request body also contains the configured model, the output-token
limit, and the fixed system instruction. The system instruction sits
outside the message array. Choose its wording, but send it consistently.
In this exercise a request begins and ends with a user message, with
strictly alternating roles and nonempty text. Those restrictions define
the small program being built; they are not a claim about every message
shape the live API accepts.

Model selection belongs in configuration. Ask the models endpoint for
IDs available to the account rather than selecting a familiar-looking
string from a tutorial. The following is a discovery command, not a
model invocation:

```sh
python3 - <<'PY'
import json
import os
import urllib.request

request = urllib.request.Request(
    "https://api.anthropic.com/v1/models",
    headers={
        "x-api-key": os.environ["ANTHROPIC_API_KEY"],
        "anthropic-version": "2023-06-01",
    },
)
with urllib.request.urlopen(request, timeout=30) as response:
    page = json.load(response)
for model in page["data"]:
    print(model["id"])
PY
```

This discovery helper reads the key from its environment and prints only
model IDs. The credential never appears in a command-line argument.

The [Models API reference](https://platform.claude.com/docs/en/api/models/list)
describes the response and pagination. Select an available model suitable
for text generation, then set `ANTHROPIC_MODEL` to its ID. The executable
does no discovery of its own: automatically calling that endpoint would
add work the user did not request. An unset model produces an error that
names the variable and the discovery endpoint.

Read the base URL separately from the model and key. For grading it
points at a local server. For a live run it points at the service that
will answer the request. Append `/v1/messages` once, after removing
trailing slashes from the configured base. Log request failures without
logging authorization headers or the API key.

## 1.6 A complete exchange has two messages

Agent owns the conversation; Engine performs the HTTP exchange. The
engine receives access to Agent through the shared parent interface.
The code building and decoding requests belongs in `internal/llm`.
The public library makes that behavior available to the CLI without
requiring the CLI to know how the engine is constructed.

For each question, prepare a request from the completed history and
the new user message. After the response has been decoded and validated,
retain the user message and assembled assistant answer together. On
failure, return an error and end the CLI session. A rejected question
must not become half of a completed exchange.

This boundary matters even without retries. The earlier reference added
the user message before the request and let interactive chat continue
after an error. Another question could then leave two user messages
adjacent in history. Terminating on errors keeps this exercise's behavior
small; committing only successful pairs also keeps the owned conversation
honest for a library caller.

The request permits plain text as message content. Response content is
an array of typed blocks. Walk the array, take the text blocks in order,
and concatenate their text without adding spaces or newlines. A block
boundary can occur in the middle of a word. Treating it as a formatting
instruction changes the answer.

The grader deliberately returns multiple text blocks. That fixture came
from a real audit failure: commit `835946f` records that replacing the
reference's walk with first-block indexing still scored 100 because the
fake returned only one block. Changing the fixture made the shortcut
observable. A check named `replies` could not protect a case the server
never produced.

Unknown block types are not text. Ignore them when assembling the text
answer, but require a nonempty answer for this exercise. Do not invent a
non-text block with a made-up `text` field merely to score the filter.
The fixture must preserve the protocol it is meant to teach.

An HTTP 200 is only the start of parsing. Require decodable JSON, the
content needed to assemble an answer, and valid usage counts. Missing
accounting is not zero accounting. On an invalid response, emit a
diagnostic and exit nonzero; do not print a plausible answer or a success
usage record. Successful answers already printed before a later failure
remain valid answers, but the session did not end successfully.

Use a finite timeout on the owned HTTP client so a stalled server cannot
wait forever. The precise duration is a deployment choice; the grader
also has its own deadline. Neither mechanism authorizes a retry. Each
accepted input question makes one request in this exercise.

## 1.7 Read the token counts

Engine owns usage because it receives the response that reports it.
Add input and output counts from every valid response, starting with
the first one. Expose the totals through the library. The CLI reads
them when stdin closes and emits the final protocol record.

The totals describe this engine's conversation, not a process-wide
counter. Two Agents running under one Ensemble must not borrow each
other's history or usage. A shared logger does not make every piece
of state shared.

Repeated history creates repeated input. With similarly sized exchanges,
the first request sends roughly one unit of conversation, the second
sends more, and the total grows as a sum of increasingly long prefixes.
That is a reason to observe usage from the beginning. It is not a promise
that every live request costs more dollars than the preceding one:
answers have different lengths, providers account for cached tokens,
and model prices differ.

The fake's token counter is a deterministic grading device. Its counts
are not predictions for a live tokenizer. Your program reports the
numbers it receives from whichever server it is using; estimating from
character counts would substitute an invented measurement for an
available one.

No dollar figure belongs in this exercise. Model-specific prices and
cache accounting require more information than the two counters here.
Likewise, keep the cost of having an assistant write the exercise
separate from the cost of running the program you built. Local fake
grading does not spend model tokens at all.

## 1.8 Exercise, graded

The exercise has a behavioral contract and an architectural contract.
Both apply. Passing the old seven-check grader proves only the behavior
that grader actually observes.

| Existing check | Points | Required behavior |
|---|---:|---|
| `protocol` | 15 | One JSON answer per round, then usage and clean EOF exit; no stdout noise |
| `wire` | 15 | Correct request fields, headers, supplied key/model, and exercise message rules |
| `calls` | 10 | Exactly one API call per input question |
| `replies` | 10 | Output comes from concatenating the server's text blocks |
| `memory` | 25 | Later requests retain facts supplied in earlier assistant replies |
| `growth` | 15 | Earlier message roles and text remain unchanged |
| `usage` | 10 | Final counters equal the sum of response usage |

The server supplies a fact in an early assistant answer and then checks
whether a later request still carries it. That distinguishes retaining
a conversation from sending each question in isolation. The fake also
checks the actual key and model values supplied by the harness; accepting
any nonempty value would fail to detect hardcoded configuration.

The second edition additionally requires these acceptance checks. They
are pass/fail requirements, outside the inherited 100-point score:

| Property | What the acceptance run must establish |
|---|---|
| Missing configuration | Missing key or model exits nonzero, names the variable, prints no stdout, and makes no HTTP call; missing model points to discovery |
| Empty input | Clean EOF without questions emits zero usage and exits 0 |
| Invalid input | Malformed JSON or empty question emits a diagnostic and exits nonzero without issuing a request for that line |
| Request failure | Transport failure, non-200 status, malformed response JSON, empty answer, or missing/invalid usage ends the session without retry, fabricated answer, or success usage record |
| Bounded wait | The owned HTTP client has a finite timeout; a stalled request ends with a diagnostic |
| Exact conversation | Each new request adds exactly the prior answer and new question to the previous role/text prefix |
| Public library | An external Go consumer can construct and use the library; the CLI imports that public API |
| Ownership | Ensemble owns the logger; Agent owns configuration/history; Engine owns transport/usage; multiple Agents keep independent state |
| Parent access | Engine reaches Agent, then Ensemble, through common interfaces; request and parsing code can reach logging through that chain |
| Package boundaries | Common declares shared data/interfaces; behavior remains in its responsible spoke; no sibling implementation imports or CLI implementation imports |
| State and wiring | No mutable application globals, duplicated owned state, dependency bags, closure bridges, or second composition root |

Architecture review must inspect the actual imports, constructors, and
paths to owned data. Finding an identifier named `Host` proves nothing
about what an engine stores or calls. A rename must preserve the result;
breaking a required relationship must change it. Apply the same standard
to behavioral checks: remove the promised behavior in a copy and verify
that the relevant check fails.

Use the public library from a separate consumer to exercise ownership.
Create two Agents under one Ensemble, send distinct conversations, and
verify that neither includes the other's messages or token totals.
Direct logger output to a capture writer and provoke a request failure.
The diagnostic must reach that writer through the owner chain, without
the consumer handing the engine a separate logging closure.

Build and check your CLI from the repository root:

```sh
make grade-dir CH=1 DIR=solutions/edition-2/ch01
```

Substitute your own directory if you are building elsewhere. The reference
is a legal starting point; the course does not grade authorship or require
you to reproduce its names. It does require the relationships and behavior
described on this page.

Each second-edition snapshot is a complete repository. Record validated
work in its own Git history; the succeeding chapter starts from that
history and extends it. The snapshot is both runnable code and a record
of how the code arrived there. Preserve the first-edition solutions.

## 1.9 Taking it for a spin

The coder must exercise every feature with a real model before this
section is complete, initially through the CLI. The local fake verifies
deterministic cases; a live run verifies that a real service accepts the
requests and that the program is usable through its public interface.

The run must cover model discovery, configured transport, a multi-turn
conversation retaining information from an earlier answer, final usage,
and failure diagnostics. Run a separate, executable public-library
consumer against the real backend to exercise independent Agents and
logger access; an internal unit test does not demonstrate this user path.
Local negative probes accompany
the live run for failures that should not require intentionally malformed
traffic to a paid service. If optional chat mode is implemented, run it
as well. Record the commands, date, returned model identity, observations,
and reported usage without recording secrets.

[LIVE RECEIPT PENDING: insert the coder's actual commands, observed
responses, usage, and feature-by-feature evidence. No live run or
second-edition acceptance pass is claimed by this draft.]
