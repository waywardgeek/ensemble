"""Read-only reconciliation of a persisted fiction assignment.

The run ledger is authoritative for intent. Ensemble owns child lifetime and saved
submissions. This operation observes those owners; it never creates a replacement
child or treats inactivity as proof that dispatch did not occur.
"""
from child_calls import ObservationError, call_json


async def reconcile(connection, store):
    """Return observed progress or a supervisor request without changing the ledger.

    A matching submission is only a candidate: the driver must still validate its
    phase-specific fields and required files before RunStore.accept. These status
    snapshots are not a global concurrency barrier. The workflow must be the sole
    dispatcher for these children; external intervention requires reconciliation.
    """
    pending = store.snapshot()["pending"]
    if pending is None:
        return {"action": "ready"}
    identity = pending["child_id"]

    def supervision(reason):
        return {
            "action": "supervise",
            "assignment_id": pending["id"],
            "child_id": identity,
            "reason": reason,
        }

    try:
        catalog = await call_json(connection, "agents_status", {})
        agents = catalog.get("agents")
        if type(agents) is not list:
            return supervision("Invalid child status catalog")
        matches = []
        for status in agents:
            if type(status) is not dict or type(status.get("agent_id")) is not str:
                return supervision("Invalid child status entry")
            if status["agent_id"] == identity:
                matches.append(status)
        if len(matches) > 1:
            return supervision("Duplicate child status identity")
        if matches:
            status = matches[0]
            if status.get("paused") is not False or status.get("failure") is not None:
                return supervision("Child is paused, failed, or status is incomplete")
            activity = status.get("activity")
            if activity == "busy":
                return {"action": "observe", "assignment_id": pending["id"],
                        "child_id": identity}
            if activity != "idle":
                return supervision("Child activity requires inspection")

        # Saved results can outlive the child. Missing evidence is not evidence
        # of non-admission, and an older submission must never satisfy new work.
        saved = await call_json(connection, "get_submission", {"agent_id": identity})
        if saved.get("agent_id") != identity or saved.get("historical") is not True:
            return supervision("Saved submission identity is invalid")
        payload = saved.get("payload")
        if type(payload) is not dict or payload.get("assignment_id") != pending["id"]:
            return supervision("No matching saved assignment result")
        return {"action": "validate_result", "assignment_id": pending["id"],
                "child_id": identity, "submission": payload}
    except ObservationError:
        return supervision("Child evidence unavailable; inspect before any redispatch")
