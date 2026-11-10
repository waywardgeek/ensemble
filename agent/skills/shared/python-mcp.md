# Python MCP skill tools

`ensemble_mcp.py` is a standard-library-only Unix stdio helper, not a skill,
installer, grant, or alternate Agent dispatcher. Copy it alongside executable
skill code **inside the Agent's sandbox**. The host loads the skill normally;
its stdio MCP declaration launches the script. Nothing imports this module into
the Ensemble process. Python 3.8 or later is required.

## Declare a tool

```python
from ensemble_mcp import Server

server = Server("example")

@server.tool("echo", "Return supplied text", {
    "type": "object",
    "properties": {"text": {"type": "string"}},
    "required": ["text"],
    "additionalProperties": False,
})
async def echo(connection, arguments):
    return {"content": [{"type": "text", "text": arguments["text"]}]}

server.run()
```

Handlers must be asynchronous. Declare every accepted argument: Ensemble validates
calls against this declaration before invoking the server. The helper does not
provide a second JSON Schema validator. It implements the existing historical
MCP `2024-11-05` stdio exchange, not every MCP version or transport.

Keep stdout exclusively for protocol messages. Ordinary diagnostics belong on
stderr and must not contain secrets. Arbitrary handler exceptions return a fixed
error rather than disclosing exception details. Handler results are full MCP
result dictionaries; preserve `isError` and content rather than treating every
reply as success.

## A workflow is the same kind of tool

Inside the handler:

```python
from ensemble_mcp import check_contracts

# expected_contracts is a reviewed, checked-in mapping of tool names to exact
# input schemas. Do not build it from today's catalog in production: that would
# compare the catalog with itself and conceal drift.
check_contracts(expected_contracts, await connection.list_tools())
result = await connection.call_tool("some_tool", arguments)
```

`list_tools` returns the calling Agent's currently enabled definitions.
`call_tool` returns the full result of its ordinary dispatcher. The Python caller
cannot select another Agent through this interface. Listing and passing preflight
confer no authority; every later call still undergoes current policy enforcement.

The same `check_contracts(expected, definitions)` function is available to offline
contract tests. It checks only expected names, refuses missing tools or duplicate
catalog names, and compares complete parsed input schemas. Object-key order and
JSON formatting do not matter; array order, fields, values and parsed scalar types
do. Thus integer and floating-point representations are distinguished, as are
booleans and numbers. Optional additions still require deliberate expectation
updates. A mismatch reports the tool and structural location, without schema
values. This is drift detection, not proof of behavioral/result compatibility.

## Lifetime and limits

One reader continues receiving replies while async handlers await reverse calls.
The server correlates its own reverse request IDs, serializes writes, owns each
handler task, and removes completed tasks. EOF closes admission, releases pending
waits, cancels handlers, and joins them. A callback's cancellation or lost reply
is **not rollback** of an operation already admitted by Ensemble. No automatic
retry is performed. MCP cancellation notifications do not retract admitted work;
server lifetime termination cancels its Python handlers, not independent Agent
Jobs.

Handlers must yield to the event loop. Blocking synchronous I/O or CPU loops can
prevent progress and cooperative shutdown; this helper is not a watchdog. The
host retains ownership of process termination and joining. The stdio reader has
a 16 MiB message limit; oversized frames fail instead of truncating. This limit
is not a model context window or a promise that a whole novel fits in a request.

## Current implementation checks

From the repository root, with Python available:

```sh
python3 -m unittest discover -s agent/PythonTests -v
```

These tests use real subprocess pipes and exercise callback correlation, remote
refusal, disconnect, exact matching, and refusing drift before issuing a tool
call. They are implementation tests, not the independent chapter grader.

The C# test runner's `--mcp-process-tests` includes a macOS sandbox integration
that copies the helper into a disposable workspace, loads a real skill, dispatches
its outer workflow tool, performs an allowed reverse write, refuses outside-file
and ungranted-tool calls, and unloads/joins the server. It makes no provider call.
Unsupported operating systems are not certified by that integration test.
