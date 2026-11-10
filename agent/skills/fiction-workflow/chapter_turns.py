"""Bounded chapter turn sequencing, derived from durable accepted receipts.

This is not the complete book driver. The caller prepares role homes, preflights
contracts, supplies complete-input task instructions and owns child cleanup. The
later detail/global review rounds are deliberately outside this module.
"""
import hashlib
from pathlib import Path

from assignments import dispatch
from child_calls import ObservationError
from recovery import reconcile
from run_state import StateError
from story_changes import author_artifacts, validate_submission, verify_unchanged


AUTHOR_SCHEMA = {
    "type": "object",
    "properties": {"assignment_id": {"type": "string"},
                   "prior_chapter_requests": {"type": "array", "items": {"type": "integer", "minimum": 1}, "uniqueItems": True}},
    "required": ["assignment_id"],
    "additionalProperties": False,
}
EDITOR_SCHEMA = {
    "type": "object",
    "properties": {
        "assignment_id": {"type": "string"},
        "decision": {"type": "string", "enum": ["revise", "approve"]},
        "prior_chapter_permissions": {"type": "array", "items": {"type": "integer", "minimum": 1}, "uniqueItems": True},
    },
    "required": ["assignment_id", "decision"],
    "additionalProperties": False,
}


def _validate_text_artifact(store, name):
    """Require usable UTF-8 before accepting a writing receipt, not good prose.

    Read bounded chunks rather than allocating a novel-sized output. Carryover
    must also contain text; an explicit statement that no threads remain is valid.
    Cooperative roles have finished writing before this check and acceptance.
    """
    path = (store.directory / name).resolve(strict=True)
    if store.directory not in path.parents or not path.is_file():
        raise StateError("Writing artifact must be a run-local regular file")
    has_text = False
    try:
        with path.open("r", encoding="utf-8", errors="strict") as source:
            while True:
                text = source.read(65536)
                if not text:
                    break
                if text.strip():
                    has_text = True
    except UnicodeDecodeError as error:
        raise StateError("Writing artifact is not valid UTF-8") from error
    if not has_text:
        raise StateError("Writing artifact is empty or whitespace-only")


async def advance_chapter(connection, store, *, chapter, author_id, editor_id,
                          author_task, editor_task, manuscript, carryover, feedback,
                          review_round=0, working_outline=None, initial_hashes=None):
    """Perform one writing turn, or reconcile it; stop after three editorial decisions.

    Round zero is initial drafting; rounds one through three revise reviewed books.
    Each round uses fresh chapter-local children prepared by the caller.

    Caller supplies run-relative artifact paths and stable task instructions,
    including complete input readings. Writing Agents must finish all file work
    before submitting: native submission alone is not a background-Job barrier.
    The cooperative sole driver must not admit other work to these role Agents.

    Returns advanced/approved/revision_limit, observe, or supervise. Repeated calls
    after approval or the third editorial decision do nothing. Unknown admission never dispatches automatically. Changing caller
    inputs during a resumed assignment refuses rather than retargeting that work.
    """
    if type(chapter) is not int or chapter < 1:
        raise StateError("Chapter numbers start at one")
    if type(review_round) is not int or not 0 <= review_round <= 3:
        raise StateError("Review round must be zero through three")
    phase = "write" if review_round == 0 else "revise"
    if author_id == editor_id:
        raise StateError("Author and Editor must be separate children")
    paths = (manuscript, carryover, feedback) + (() if working_outline is None else (working_outline,))
    for path in paths:
        if (type(path) is not str or not path or "\\" in path
                or Path(path).is_absolute()
                or any(part in ("", ".", "..") for part in path.split("/"))):
            raise StateError("Artifacts must be run-relative paths")
    if len(set(paths)) != len(paths):
        raise StateError("Chapter, carryover and feedback require distinct paths")

    state = store.snapshot()
    receipts = [item for item in state["accepted"]
                if item["assignment"]["phase"] == phase
                and item["assignment"]["round"] == review_round
                and item["assignment"]["chapter"] == chapter]
    if len(receipts) > 6:
        raise StateError("Chapter exceeds three author/editor attempts")
    for index, receipt in enumerate(receipts):
        assignment = receipt["assignment"]
        editor = index % 2 == 1
        expected_child = editor_id if editor else author_id
        if (assignment["turn"] != index + 1 or assignment["round"] != review_round
                or assignment["child_id"] != expected_child):
            raise StateError("Accepted chapter turns do not match this role sequence")
        previous = receipts[index - 1]["submission"] if index else None
        validate_submission(receipt["submission"], editor, chapter, index + 1, previous)
        expected_paths = {feedback} if editor else set(author_artifacts(
            manuscript, carryover, working_outline, previous))
        if set(receipt["artifacts"]) != expected_paths:
            raise StateError("Accepted chapter artifact paths changed")
        if editor and receipt["submission"]["decision"] == "approve":
            if index != len(receipts) - 1 or state["pending"] is not None:
                raise StateError("Work follows an already-approved chapter")
            return {"action": "approved", "chapter": chapter}

    # The limit is derived, never reset by restarting the driver. A negative
    # final review remains negative; stopping is not editorial approval.
    if len(receipts) == 6:
        if state["pending"] is not None:
            raise StateError("Work follows a chapter revision limit")
        return {"action": "revision_limit", "chapter": chapter}

    turn = len(receipts) + 1
    editor = turn % 2 == 0
    child_id = editor_id if editor else author_id
    task = editor_task if editor else author_task
    artifacts = [feedback] if editor else author_artifacts(
        manuscript, carryover, working_outline, receipts[-1]["submission"] if receipts else None)
    pending = state["pending"]
    if task is None:
        # A driver resuming pending work must not regenerate its instruction from
        # mutable draft files. None explicitly selects the saved original task.
        if pending is None:
            raise StateError("A new writing assignment requires task instructions")
        task = pending["task"]
    else:
        task += (
            "\nFinish all file writes and owned work before submitting. "
            "Submission alone does not stop background work. "
            "Ask your parent when blocked; do not silently omit required reading."
        )
    if pending is not None:
        if (pending["phase"] != phase or pending["round"] != review_round
                or pending["chapter"] != chapter or pending["turn"] != turn
                or pending["child_id"] != child_id or pending["task"] != task):
            raise StateError("Pending assignment differs from the requested chapter turn")
        result = await reconcile(connection, store)
        if result["action"] != "validate_result":
            return result
        payload = result["submission"]
    else:
        # Each role's first appearance creates its chapter-local child. Later
        # turns reuse that conversation; uncertain creation is handled above.
        new_child = turn <= 2
        schema = EDITOR_SCHEMA if editor else AUTHOR_SCHEMA
        try:
            payload = await dispatch(
                connection, store, child_id=child_id, phase=phase, round_number=review_round,
                chapter=chapter, turn=turn, task=task, new_child=new_child,
                submit_schema=schema if new_child else None)
        except (ObservationError, ConnectionError):
            return {"action": "supervise", "child_id": child_id,
                    "assignment_id": store.snapshot()["pending"]["id"],
                    "reason": "Assignment outcome uncertain; reconcile before continuing"}

    validate_submission(payload, editor, chapter, turn,
                        receipts[-1]["submission"] if receipts else None)
    # The selected files, not child-supplied paths, define this turn's artifacts.
    # File errors leave pending intent intact. Hashes describe acceptance-time
    # bytes, not an immutable archive or an editorial quality judgment.
    if editor:
        prior = receipts[-1]["artifacts"]
        for name in prior:
            path = (store.directory / name).resolve(strict=True)
            if store.directory not in path.parents:
                raise StateError("Reviewed artifact escapes the run directory")
            digest = hashlib.sha256()
            with path.open("rb") as source:
                while True:
                    block = source.read(65536)
                    if not block:
                        break
                    digest.update(block)
            if digest.hexdigest() != prior[name]:
                raise StateError("Reviewed draft changed after the Author's accepted turn")
    if initial_hashes is not None:
        verify_unchanged(store, initial_hashes, artifacts)
    for name in artifacts:
        _validate_text_artifact(store, name)
    store.accept(payload, artifacts)
    approved = editor and payload["decision"] == "approve"
    action = "advanced"
    if approved:
        action = "approved"
    elif editor and turn == 6:
        action = "revision_limit"
    return {"action": action,
            "chapter": chapter, "turn": turn, "role": "editor" if editor else "author"}
