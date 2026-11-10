"""One-step fiction orchestration using ordinary child tools and durable receipts.

The MCP entry point holds RunStore, validates contracts, and repeatedly calls
advance_book while progress is ordinary. Supervision and observation are returned
explicitly. No task is retried here and no new scheduler or Agent registry exists.
"""
import json
import os
from pathlib import Path
import tempfile

from book_plan import next_step
from chapter_turns import advance_chapter
from child_calls import ObservationError, call_json
from review_inputs import freeze_final, freeze_round, _read, _write
from review_turn import advance_review, source_info
from role_homes import MARKER, prepare_home
from run_state import StateError, _object, _reject_constant
from story_changes import author_artifacts, chapter_files, digest, verify_unchanged


def _publish(path, value):
    """Publish a small local value under the caller's exclusive run lock."""
    temporary = None
    try:
        with tempfile.NamedTemporaryFile(dir=path.parent, prefix=".workflow-",
                                         delete=False) as output:
            temporary = Path(output.name)
            output.write((json.dumps(value, sort_keys=True, indent=2) + "\n").encode())
            output.flush()
            os.fsync(output.fileno())
        os.replace(temporary, path)
    finally:
        if temporary is not None:
            temporary.unlink(missing_ok=True)


def _load(path):
    if path.is_symlink() or not path.is_file():
        raise StateError("Expected ordinary workflow record")
    with path.open("rb") as source:
        data = source.read(1048577)
    if len(data) > 1048576:
        raise StateError("Workflow record exceeds one MiB")
    return json.loads(data, object_pairs_hook=_object, parse_constant=_reject_constant)


def bind_plan(store, *, chapter_count, outline, voice, baseline_root,
              children_directory, source_limit, result_limit):
    """Pin run inputs before admission; repeated invocations must match exactly.

    children_directory is the actual initiating Agent's children directory, supplied
    by trusted host setup. It is not a way to choose another Agent for reverse calls.
    Source and result limits are transport bytes, never a token capacity assertion.
    """
    if type(chapter_count) is not int or chapter_count < 1:
        raise StateError("Chapter count must be a positive integer")
    working_outline = "working-outline.md"
    if working_outline in (outline, voice) or outline == voice:
        raise StateError("Original outline, working outline and voice require distinct paths")
    inputs = {}
    for name in (outline, voice):
        inputs[name] = source_info(store, name, source_limit, result_limit)["sha256"]
    plan = {"version": 2, "chapter_count": chapter_count, "outline": outline,
            "working_outline": working_outline,
            "voice": voice, "input_hashes": inputs,
            "baseline_root": str(Path(baseline_root).resolve(strict=True)),
            "children_directory": str(Path(children_directory).resolve(strict=True)),
            "source_limit": source_limit, "result_limit": result_limit}
    path = store.directory / "plan.json"
    state = store.snapshot()
    if path.exists() or path.is_symlink():
        if _load(path) != plan:
            raise StateError("Run inputs changed; reconcile rather than retargeting work")
    else:
        if state["accepted"] or state["pending"] is not None:
            raise StateError("Existing work has no pinned run plan")
        working = store.directory / working_outline
        if working.exists() or working.is_symlink():
            if digest(store, working_outline) != inputs[outline]:
                raise StateError("Unbound working outline differs from the initial outline")
        else:
            _write(working, _read(store.directory / outline, source_limit))
        _publish(path, plan)
    return plan


def _child(store, round_number, chapter, role):
    return "fiction-%s-r%d-c%d-%s" % (
        store.snapshot()["run_id"], round_number, chapter, role)


def _paths(chapter):
    stem = "chapters/%04d" % chapter
    return {"manuscript": stem + ".md", "carryover": stem + "-carryover.md",
            "feedback": stem + "-editor.md"}


async def _cleanup(connection, store, keep):
    """Join completed, obsolete children before memory transfer or snapshotting.

    A saved joined receipt avoids repeating successful shutdown. If shutdown took
    effect but its reply was lost, absence is deliberately not synthesized into a
    receipt: ask the supervisor to reconcile that exceptional case.
    """
    identities = dict.fromkeys(r["assignment"]["child_id"]
                               for r in store.snapshot()["accepted"])
    directory = store.directory / "joined"
    directory.mkdir(exist_ok=True)
    for identity in identities:
        if identity in keep:
            continue
        path = directory / (identity + ".json")
        expected = {"agent_id": identity, "stopped": True}
        if path.exists():
            if _load(path) != expected:
                raise StateError("Invalid joined-child receipt")
            continue
        try:
            reply = await call_json(connection, "shutdown_agent", {"agent_id": identity})
            if reply != expected:
                raise ObservationError("Invalid joined shutdown result")
            _publish(path, expected)
        except (ObservationError, ConnectionError):
            return {"action": "supervise", "child_id": identity,
                    "reason": "Child cleanup uncertain; verify joined teardown before continuing"}
    return None


def _prepare_role(store, plan, step, role, identity):
    home = Path(plan["children_directory"]) / identity
    if home.exists():
        if _load(home / MARKER) != {"version": 1, "role": role}:
            raise StateError("Existing child home has the wrong role")
        return
    state = store.snapshot()
    if ((state["pending"] is not None and state["pending"]["child_id"] == identity)
            or any(r["assignment"]["child_id"] == identity for r in state["accepted"])):
        raise StateError("Admitted writing child home is missing; requires supervision")
    previous = None
    # The schedule orders these receipts. The latest prior role owns its memory;
    # cleanup above has joined it before this copy can happen.
    for receipt in state["accepted"]:
        old = receipt["assignment"]
        if old["phase"] in ("write", "revise"):
            old_role = "author" if old["turn"] % 2 else "editor"
            if old_role == role and old["child_id"] != identity:
                previous = Path(plan["children_directory"]) / old["child_id"]
    prepare_home(home, role, plan["baseline_root"], previous)


def _reading(store, plan, names):
    instructions = []
    for name in dict.fromkeys(names):
        info = source_info(store, name, plan["source_limit"], plan["result_limit"])
        instructions.append("Read the complete input (%d source bytes): %s" % (
            info["bytes"], json.dumps(info["read"], sort_keys=True)))
    return "\n".join(instructions)


def _writing_receipts(state, step):
    return [r for r in state["accepted"]
            if all(r["assignment"][key] == step[key] for key in ("phase", "round", "chapter"))]


def _initial_hashes(plan):
    return dict(plan["input_hashes"], **{
        plan["working_outline"]: plan["input_hashes"][plan["outline"]]})


async def advance_book(connection, store, plan):
    """Execute/reconcile one scheduled assignment, or report completion/blockage.

    All required input files are checked before admission. A reviewer reads frozen
    text; an Author edits working text. Role skills are explicitly requested, not
    silently granted. Actual role activation/read behavior is separately tested.
    """
    if _load(store.directory / "plan.json") != plan:
        raise StateError("Driver plan differs from the pinned run inputs")
    for name, digest in plan["input_hashes"].items():
        if source_info(store, name, plan["source_limit"], plan["result_limit"])["sha256"] != digest:
            raise StateError("Outline or voice changed during the run")
    state = store.snapshot()
    step = next_step(state, plan["chapter_count"])
    writing = step["phase"] in ("write", "revise")
    allowed = []
    if state["pending"] is not None and writing:
        paths = _paths(step["chapter"])
        receipts = _writing_receipts(state, step)
        allowed = ([paths["feedback"]] if step["turn"] % 2 == 0 else author_artifacts(
            paths["manuscript"], paths["carryover"], plan["working_outline"],
            receipts[-1]["submission"] if receipts else None))
    verify_unchanged(store, _initial_hashes(plan), allowed)
    keep = set()
    if writing:
        for role in ("author", "editor"):
            keep.add(_child(store, step["round"], step["chapter"], role))
    if state["pending"] is not None:
        keep.add(state["pending"]["child_id"])
    blocked = await _cleanup(connection, store, keep)
    if blocked is not None:
        return blocked
    if step["phase"] == "complete":
        freeze_final(store,
            [_paths(number)["manuscript"] for number in range(1, plan["chapter_count"] + 1)],
            plan["source_limit"], contexts=[plan["working_outline"], plan["voice"]])
        return {"action": "complete"}
    (store.directory / "chapters").mkdir(exist_ok=True)
    (store.directory / "reviews").mkdir(exist_ok=True)
    round_number = step["round"]
    chapter = step["chapter"]
    frozen = None
    if round_number:
        frozen = freeze_round(store, round_number,
            [_paths(number)["manuscript"] for number in range(1, plan["chapter_count"] + 1)],
            plan["source_limit"], contexts=[plan["working_outline"], plan["voice"]])
    if not writing:
        global_review = step["phase"] == "global_review"
        source = frozen["manuscript"] if global_review else frozen["chapters"][chapter - 1]
        return await advance_review(connection, store,
            phase=step["phase"], review_round=round_number, chapter=chapter,
            child_id=_child(store, round_number, chapter, "reviewer"),
            source=source, output="reviews/r%d-c%d.md" % (round_number, chapter),
            guidance="Load fiction-reviewer before reviewing.\n" + _reading(
                store, plan, frozen["contexts"]),
            source_limit=plan["source_limit"], result_limit=plan["result_limit"])

    paths = _paths(chapter)
    author = _child(store, round_number, chapter, "author")
    editor = _child(store, round_number, chapter, "editor")
    # Only prepare the role whose first turn is next. An interruption before
    # ledger preparation leaves a harmless prepared home, not an admitted task.
    role = "author" if step["turn"] % 2 else "editor"
    _prepare_role(store, plan, step, role, author if role == "author" else editor)
    if state["pending"] is not None:
        return await advance_chapter(connection, store, chapter=chapter,
            author_id=author, editor_id=editor, author_task=None, editor_task=None,
            review_round=round_number, working_outline=plan["working_outline"],
            initial_hashes=_initial_hashes(plan), **paths)
    inputs = [plan["working_outline"], plan["voice"]]
    receipts = _writing_receipts(state, step)
    authorized = []
    requested = []
    if receipts:
        if role == "author":
            authorized = receipts[-1]["submission"].get("prior_chapter_permissions", [])
        else:
            requested = receipts[-1]["submission"].get("prior_chapter_requests", [])
            # Approval must cover repairs made on earlier attempts too, even
            # when the latest Author no longer has permission to edit them.
            authorized = sorted({number for receipt in receipts
                for number in receipt["submission"].get("prior_chapter_permissions", [])})
    for number in sorted(set(authorized + requested)):
        inputs.extend(chapter_files(number))
    if chapter > 1:
        previous = _paths(chapter - 1)
        inputs.extend([previous["manuscript"], previous["carryover"]])
    if round_number:
        inputs.extend(["reviews/r%d-c%d.md" % (round_number, chapter),
                       "reviews/r%d-c0.md" % round_number])
    # Prompts must remain stable within each assignment, including reconciliation.
    # Each turn explicitly adds only inputs that already exist at that position.
    if step["turn"] > 1 or round_number:
        inputs.extend([paths["manuscript"], paths["carryover"]])
    if step["turn"] > 2:
        inputs.append(paths["feedback"])
    reading = _reading(store, plan, inputs)
    shared = ("\nChapter %d of %d; review round %d.\n" % (
        chapter, plan["chapter_count"], round_number) + reading
        + "\nIf complete reading exceeds model capacity, ask your parent; do not chunk or summarize.\n")
    author_task = ("Load fiction-author. Write or revise the assigned chapter." + shared
                   + "Write chapter: " + str(store.directory / paths["manuscript"])
                   + "\nWrite continuity carryover: " + str(store.directory / paths["carryover"])
                   + "\nWorking outline: " + str(store.directory / plan["working_outline"])
                   + "\nPreserve the brief, voice, chapter count and established story facts; explain consequential changes in carryover."
                   + "\nEditor-authorized earlier chapter numbers for THIS turn: " + json.dumps(authorized)
                   + "\nYou may edit only those earlier chapters and their carryovers, at these paths: "
                   + json.dumps([str(store.directory / name) for n in authorized for name in chapter_files(n)])
                   + "\nFor additional earlier repairs, submit prior_chapter_requests (integer chapter numbers) and explain scope/reason in carryover. Do not edit them yet."
                   + "\nUnresolved requests at the attempt limit stay in carryover for later revision or parent supervision; no extra turns.")
    editor_task = ("Load fiction-editor. Review the assigned draft without modifying it." + shared
                   + "Write feedback: " + str(store.directory / paths["feedback"])
                   + "\nReview the working outline and every supplied changed earlier chapter/carryover as well as the current draft."
                   + "\nJudge creative departures by reader benefit and continuity, not mechanical outline compliance."
                   + "\nAuthor requests for earlier chapters: " + json.dumps(requested)
                   + "\nYou may grant a subset with prior_chapter_permissions in your typed submission, decision revise, only if another Author turn remains. Explain the bounded repair in feedback."
                   + "\nApproval covers the current outline/draft and every earlier chapter repaired during this loop. Third-attempt requests must be deferred in feedback or brought to the parent, not granted unused permission."
                   + "\nSubmit decision approve or revise honestly, even on the third attempt.")
    return await advance_chapter(connection, store, chapter=chapter,
        author_id=author, editor_id=editor, author_task=author_task,
        editor_task=editor_task, review_round=round_number,
        working_outline=plan["working_outline"], initial_hashes=_initial_hashes(plan), **paths)
