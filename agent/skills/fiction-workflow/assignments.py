"""Dispatch one durable assignment through ordinary owned-child tools.

The chapter driver owns role preparation, preflight schema checks, and turn order.
This module owns no child lifetime and never accepts a result on the ledger's
behalf. After a failure or cancellation, use recovery.reconcile, not dispatch.
"""
from child_calls import ObservationError, call_json, spawn_result
from run_state import StateError


async def dispatch(connection, store, *, child_id, phase, round_number,
                   chapter, turn, task, new_child, submit_schema=None):
    """Save intent, dispatch once, and return a matching validation candidate.

    A new child must have its role home prepared before this call. Existing
    children must be quiescent from the preceding accepted turn. The caller is
    the sole dispatcher and holds RunStore throughout. A schema is required for
    new children; existing children retain their original submission contract.

    No exception rolls back intent: the admitted operation may survive a lost
    reply. Returning a payload does not establish artifact validity or approval.
    """
    if type(new_child) is not bool:
        raise StateError("Specify whether this assignment creates a child")
    if new_child and type(submit_schema) is not dict:
        raise StateError("New writing children require a submission schema")
    if not new_child and submit_schema is not None:
        raise StateError("An existing child's submission schema cannot be replaced")

    assignment = store.prepare(child_id, phase, round_number, chapter, turn, task)
    # The saved task is the supplied instruction; this deterministic envelope
    # binds the actual turn to its durable assignment without another state copy.
    instruction = (
        "Workflow assignment_id: " + assignment["id"] + "\n"
        "Include this exact assignment_id in your typed submission.\n"
        "Ask your parent if you need clarification; do not invent approval.\n\n"
        + task
    )
    if new_child:
        payload = await spawn_result(connection, {
            "agent_id": child_id,
            "task": instruction,
            "submit_schema": submit_schema,
        })
    else:
        response = await call_json(connection, "ask_agent", {
            "agent_id": child_id,
            "content": instruction,
        })
        if (response.get("agent_id") != child_id
                or type(response.get("response")) is not str):
            raise ObservationError("Invalid child turn completion")
        saved = await call_json(connection, "get_submission", {"agent_id": child_id})
        if saved.get("agent_id") != child_id or saved.get("historical") is not True:
            raise ObservationError("Invalid saved submission identity")
        payload = saved.get("payload")

    if type(payload) is not dict or payload.get("assignment_id") != assignment["id"]:
        raise ObservationError("Completed submission belongs to a different assignment")
    return payload
