"""Complete-file review tasks and single-assignment acceptance.

The caller owns fresh reviewer homes, memory policy, model capacity and cleanup.
Byte limits below cover file/tool transport, not token capacity. Instructions are
not a security enforcement mechanism or proof the model followed the read plan.
"""
import hashlib
import json

from assignments import dispatch
from child_calls import ObservationError
from recovery import reconcile
from run_state import RunStore, StateError


REVIEW_SCHEMA = {
    "type": "object",
    "properties": {"assignment_id": {"type": "string"}},
    "required": ["assignment_id"],
    "additionalProperties": False,
}


def source_info(store, name, source_limit, result_limit):
    """Inspect strict UTF-8 and compute a full numbered-read transport budget.

    Use a bounded read to refuse oversized inputs without loading the whole file.
    Paths must remain in the run. Limits are caller-selected positive byte bounds;
    they must never be presented as model token counts or context guarantees.
    """
    for limit in (source_limit, result_limit):
        if type(limit) is not int or limit < 1:
            raise StateError("Read byte limits must be positive integers")
    relative = RunStore._relative(name)
    path = (store.directory / relative).resolve(strict=True)
    if store.directory not in path.parents or not path.is_file():
        raise StateError("Review source must be a run-local file")
    with path.open("rb") as stream:
        data = stream.read(source_limit + 1)
    if len(data) > source_limit:
        raise StateError("Complete review source exceeds the configured source byte limit")
    text = data.decode("utf-8", errors="strict")
    if not text.strip():
        raise StateError("Review source is empty")
    # Native read_file numbers LF-delimited lines, retaining CRLF and not adding
    # a fictitious last line after a trailing LF. Include observation headroom.
    lines = text.count("\n") + (0 if text.endswith("\n") else 1)
    prefix_bytes = sum(len(str(line)) + 2 for line in range(1, lines + 1))
    result_bytes = len(data) + prefix_bytes + 4096
    if result_bytes > result_limit:
        raise StateError("Complete numbered review exceeds the configured result byte limit")
    return {"bytes": len(data), "sha256": hashlib.sha256(data).hexdigest(),
            "read": {"path": str(path), "start_line": 1, "end_line": 0,
                     "max_bytes": len(data), "max_output_bytes": result_bytes}}


async def advance_review(connection, store, *, phase, review_round, chapter,
                         child_id, source, output, guidance,
                         source_limit, result_limit):
    """Dispatch or reconcile one fresh reviewer and accept its complete report.

    The schedule owner calls this only at the correct phase. A source fingerprint
    is part of saved task intent, so changed input refuses reconciliation rather
    than applying an old result to new text. Source is checked again after review.
    Matching submission alone is not proof of one-read behavior; integration tests
    must inspect actual tool calls. No grade is required for workflow progress.
    """
    if (phase not in ("detail_review", "global_review")
            or type(review_round) is not int or not 1 <= review_round <= 3
            or type(chapter) is not int
            or (chapter != 0 if phase == "global_review" else chapter < 1)):
        raise StateError("Invalid review position")
    RunStore._relative(output)
    if source == output:
        raise StateError("Review must not overwrite its source")
    info = source_info(store, source, source_limit, result_limit)
    task = (
        guidance + "\nReview the entire supplied text independently. Source bytes: "
        + str(info["bytes"]) + ". Source SHA256: " + info["sha256"]
        + ".\nCall read_file exactly once for this text, with these arguments: "
        + json.dumps(info["read"], sort_keys=True)
        + ".\nDo not use keep_tool_results, chunking, summaries or partial reading. "
        "If tool limits or model capacity prevent this reading, ask your parent "
        "and do not submit a completed review. The byte size is not a token count. "
        "Write specific, constructive criticism with passage references to "
        + str(store.directory / output)
        + ". Finish the write and all owned work, then submit the assignment_id. "
        "Aim for reading, writing and submission in about three model round trips. "
        "A grade is optional, not a completion gate."
    )
    state = store.snapshot()
    position = {"phase": phase, "round": review_round, "chapter": chapter, "turn": 1}
    for receipt in state["accepted"]:
        assignment = receipt["assignment"]
        if all(assignment[key] == value for key, value in position.items()):
            if (assignment["task"] != task or assignment["child_id"] != child_id
                    or set(receipt["artifacts"]) != {output}):
                raise StateError("Accepted review inputs changed")
            return {"action": "reviewed", **position}
    pending = state["pending"]
    if pending is not None:
        if (any(pending[key] != value for key, value in position.items())
                or pending["child_id"] != child_id or pending["task"] != task):
            raise StateError("Pending review inputs changed")
        result = await reconcile(connection, store)
        if result["action"] != "validate_result":
            return result
        payload = result["submission"]
    else:
        try:
            payload = await dispatch(connection, store, child_id=child_id,
                phase=phase, round_number=review_round, chapter=chapter, turn=1,
                task=task, new_child=True, submit_schema=REVIEW_SCHEMA)
        except (ObservationError, ConnectionError):
            return {"action": "supervise", "child_id": child_id,
                    "reason": "Review outcome uncertain; reconcile existing work"}
    if type(payload) is not dict or set(payload) != {"assignment_id"}:
        raise StateError("Unexpected review submission fields")
    after = source_info(store, source, source_limit, result_limit)
    if after != info:
        raise StateError("Review source changed during review")
    # A missing, empty or malformed report is not a completed review. This
    # validates text presence, not the quality or truth of its literary judgment.
    source_info(store, output, source_limit, result_limit)
    store.accept(payload, [output])
    return {"action": "reviewed", **position}
