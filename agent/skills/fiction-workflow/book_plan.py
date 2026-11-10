"""Derive the fiction book schedule from one authoritative assignment ledger.

No second cursor is saved. This module performs no I/O or dispatch and owns no
Agents. The driver must establish the run's fixed chapter count before using it.
"""
from run_state import StateError
from story_changes import validate_submission


def next_step(state, chapter_count):
    """Validate the accepted prefix and return its next phase/round/chapter/turn.

    A writing chapter ends on explicit approval or its third editorial decision.
    Each review consumes one accepted assignment. Reviews for an entire round
    precede any revisions. Pending intent must match the next schedule position;
    this does not establish whether that assignment was admitted remotely.
    """
    if type(chapter_count) is not int or chapter_count < 1:
        raise StateError("A book requires a positive chapter count")
    accepted = state["accepted"]
    offset = 0

    def position(phase, round_number, chapter, turn):
        return {"phase": phase, "round": round_number,
                "chapter": chapter, "turn": turn}

    def match(assignment, expected):
        if any(assignment[key] != value for key, value in expected.items()):
            raise StateError("Assignment is outside the book schedule")

    def missing(expected):
        if state["pending"] is not None:
            match(state["pending"], expected)
        return expected

    def writing(phase, round_number, chapter):
        nonlocal offset
        for turn in range(1, 7):
            expected = position(phase, round_number, chapter, turn)
            if offset == len(accepted):
                return missing(expected)
            receipt = accepted[offset]
            match(receipt["assignment"], expected)
            payload = receipt["submission"]
            previous = accepted[offset - 1]["submission"] if turn > 1 else None
            validate_submission(payload, turn % 2 == 0, chapter, turn, previous)
            offset += 1
            if turn % 2 == 0:
                if payload["decision"] not in ("approve", "revise"):
                    raise StateError("Invalid editorial decision in book schedule")
                if payload["decision"] == "approve":
                    break
        return None

    for chapter in range(1, chapter_count + 1):
        step = writing("write", 0, chapter)
        if step is not None:
            return step

    for round_number in range(1, 4):
        # Global review uses chapter zero: it assesses the whole assembled book.
        reviews = [("detail_review", chapter)
                   for chapter in range(1, chapter_count + 1)]
        reviews.append(("global_review", 0))
        for phase, chapter in reviews:
            expected = position(phase, round_number, chapter, 1)
            if offset == len(accepted):
                return missing(expected)
            match(accepted[offset]["assignment"], expected)
            offset += 1
        for chapter in range(1, chapter_count + 1):
            step = writing("revise", round_number, chapter)
            if step is not None:
                return step

    if offset != len(accepted) or state["pending"] is not None:
        raise StateError("Work follows the completed third book revision")
    return {"phase": "complete"}
