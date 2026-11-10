"""Consume ordinary Ensemble child-tool results without retrying their effects.

The MCP boundary carries text, not ToolResultPart.Job. This adapter accepts the
current bare-completion or Job-report text contract. Unexpected/truncated output
requires supervision rather than guessing a result from a fragment.
"""
import json
import re


class ObservationError(ValueError):
    """The caller must reconcile retained work instead of repeating its call."""


_HEADER = re.compile(
    r"job ([1-9][0-9]*), status: (running|done|failed|killed), bytes: ([0-9]+)"
    r"(?:, exit_code: -?[0-9]+)?, truncated: (true|false), output_path: ([^\n]+)\n"
)


def _text(result):
    if type(result) is not dict or result.get("isError", False) is not False:
        raise ObservationError("Tool refused or failed; reconcile before retrying")
    content = result.get("content")
    if type(content) is not list or not content:
        raise ObservationError("Missing tool result content")
    pieces = []
    for item in content:
        if (type(item) is not dict or item.get("type") != "text"
                or type(item.get("text")) is not str):
            raise ObservationError("Expected textual child-tool result")
        pieces.append(item["text"])
    return "".join(pieces)


def _object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ObservationError("Duplicate completed-result field")
        result[key] = value
    return result


def _constant(value):
    raise ObservationError("Invalid JSON constant in completed result")


def completed_json(text):
    """Decode a single complete object; never extract JSON from child prose."""
    try:
        value = json.loads(text, object_pairs_hook=_object, parse_constant=_constant)
    except (ValueError, TypeError) as error:
        raise ObservationError("Expected one complete JSON tool result") from error
    if type(value) is not dict:
        raise ObservationError("Expected a completed-result object")
    return value


async def call_text(connection, name, arguments):
    """Call once and observe the same Job until complete.

    No retry, kill, or child dispatch occurs here. Cancellation only abandons this
    observer. The caller must have saved assignment intent before effectful calls.
    This consumes one Job's observation cursor, so the supervisor should inspect
    child status, not concurrently consume this Job's output. Twenty-minute
    supervisory checks are independent of these transport observations. Each wait can last thirty minutes,
    but output, completion and attention wake it sooner.
    """
    reply = await connection.call_tool(name, arguments)
    handle = None
    chunks = []
    size = 0
    while True:
        text = _text(reply)
        header = _HEADER.match(text)
        if header is None:
            if handle is not None:
                raise ObservationError("Lost Job report identity")
            if len(text.encode("utf-8")) > 1024 * 1024:
                raise ObservationError("Workflow result exceeds one MiB")
            return text
        observed = int(header.group(1))
        if handle is not None and observed != handle:
            raise ObservationError("Job handle changed during observation")
        handle = observed
        if header.group(4) != "false":
            raise ObservationError("Truncated Job output requires reconciliation")
        status = header.group(2)
        if status not in ("running", "done"):
            raise ObservationError("Job did not complete successfully")
        chunk = text[header.end():]
        size += len(chunk.encode("utf-8"))
        if size > 1024 * 1024:
            raise ObservationError("Workflow result exceeds one MiB")
        chunks.append(chunk)
        if status == "done":
            return "".join(chunks)
        reply = await connection.call_tool("wait_for_job", {
            "handle": handle,
            "ai_callback_delay": 1800,
            "max_output_bytes": 1024 * 1024,
        })


async def call_json(connection, name, arguments):
    """Observe one operation, then require exactly one JSON result object."""
    return completed_json(await call_text(connection, name, arguments))


async def spawn_result(connection, arguments):
    """Validate the two native spawn records, not arbitrary child output.

    Admission is not completion. Both records must identify the requested child;
    any missing, extra, or malformed record leaves reconciliation to supervision.
    """
    text = await call_text(connection, "spawn_sub_agent", arguments)
    records = text.strip().splitlines()
    if len(records) != 2:
        raise ObservationError("Expected spawn admission and completion records")
    admission = completed_json(records[0])
    completion = completed_json(records[1])
    identity = arguments["agent_id"]
    if (set(admission) != {"agent_id", "task_admitted"}
            or admission.get("agent_id") != identity
            or admission.get("task_admitted") is not True):
        raise ObservationError("Invalid spawn admission record")
    if (set(completion) != {"agent_id", "task_completed", "submitted", "payload"}
            or completion.get("agent_id") != identity
            or completion.get("task_completed") is not True
            or completion.get("submitted") is not True
            or type(completion.get("payload")) is not dict):
        raise ObservationError("Spawn did not return a matching submission")
    return completion["payload"]
