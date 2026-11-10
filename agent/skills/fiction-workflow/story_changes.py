"""Story-change policy derived from accepted writing receipts, not a second ledger."""
import hashlib
from run_state import StateError


def chapter_files(number):
    stem = "chapters/%04d" % number
    return [stem + ".md", stem + "-carryover.md"]


def validate_submission(payload, editor, chapter=None, turn=None, previous=None):
    required = {"assignment_id", "decision"} if editor else {"assignment_id"}
    field = "prior_chapter_permissions" if editor else "prior_chapter_requests"
    if type(payload) is not dict or not required <= set(payload) or set(payload) - required - {field}:
        raise StateError("Unexpected writing-turn submission fields")
    if type(payload["assignment_id"]) is not str or not payload["assignment_id"]:
        raise StateError("Missing writing-turn assignment identity")
    if editor and payload["decision"] not in ("revise", "approve"):
        raise StateError("Editor must explicitly request revision or approve")
    numbers = payload.get(field, [])
    if (type(numbers) is not list or any(type(n) is not int or n < 1 for n in numbers)
            or len(numbers) != len(set(numbers))
            or (chapter is not None and any(n >= chapter for n in numbers))):
        raise StateError("Prior-chapter changes require distinct earlier chapter numbers")
    if editor and numbers:
        if payload["decision"] != "revise" or turn == 6:
            raise StateError("Prior-chapter permission requires a remaining Author revision turn")
        if previous is not None and not set(numbers) <= set(previous.get("prior_chapter_requests", [])):
            raise StateError("Editor permission must answer the Author's explicit requests")


def author_artifacts(manuscript, carryover, working_outline, previous):
    names = [manuscript, carryover]
    if working_outline is not None:
        names.append(working_outline)
    if previous is not None:
        for number in previous.get("prior_chapter_permissions", []):
            names.extend(chapter_files(number))
    return names


def digest(store, name):
    store._relative(name)
    original = store.directory / name
    path = original.resolve(strict=True)
    if original.is_symlink() or store.directory not in path.parents or not path.is_file():
        raise StateError("Story artifact must be an ordinary run-local file")
    value = hashlib.sha256()
    with path.open("rb") as source:
        for block in iter(lambda: source.read(65536), b""):
            value.update(block)
    return value.hexdigest()


def verify_unchanged(store, initial_hashes, allowed=()):
    """Latest accepted version wins, including an authorized earlier-chapter edit.

    Pending Author outputs may already have changed after a lost reply. Only its
    receipt-derived allowed paths are exempt; all other accepted artifacts remain
    protected. This is cooperative validation, not an OS write-access boundary.
    """
    expected = dict(initial_hashes)
    for receipt in store.snapshot()["accepted"]:
        expected.update(receipt["artifacts"])
    for name, expected_digest in expected.items():
        if name not in allowed and digest(store, name) != expected_digest:
            raise StateError("Story artifact changed outside the authorized turn")
