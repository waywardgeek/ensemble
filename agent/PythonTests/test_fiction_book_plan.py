"""Book ordering through real durable receipts, without provider or child calls."""
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/fiction-workflow"))
from book_plan import next_step
from run_state import RunStore, StateError


class BookPlanTests(unittest.TestCase):
    def accept_step(self, store, step, decision="approve"):
        assignment = store.prepare("fixture", step["phase"], step["round"],
                                   step["chapter"], step["turn"], "Fixture assignment")
        result = {"assignment_id": assignment["id"]}
        if step["phase"] in ("write", "revise") and step["turn"] % 2 == 0:
            result["decision"] = decision
        (store.directory / "result.md").write_text(str(step), encoding="utf-8")
        store.accept(result, ["result.md"])

    def test_all_rounds_reopen_after_every_assignment(self):
        with tempfile.TemporaryDirectory() as directory:
            observed = []
            while True:
                with RunStore(directory) as store:
                    step = next_step(store.snapshot(), 2)
                    if step["phase"] == "complete":
                        break
                    observed.append(step)
                    self.accept_step(store, step)
            expected = []
            for chapter in (1, 2):
                for turn in (1, 2):
                    expected.append(("write", 0, chapter, turn))
            for round_number in (1, 2, 3):
                expected.extend([("detail_review", round_number, 1, 1),
                                 ("detail_review", round_number, 2, 1),
                                 ("global_review", round_number, 0, 1)])
                for chapter in (1, 2):
                    for turn in (1, 2):
                        expected.append(("revise", round_number, chapter, turn))
            self.assertEqual([(s["phase"], s["round"], s["chapter"], s["turn"])
                              for s in observed], expected)

    def test_negative_reviews_finish_without_approval(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            for _ in range(30):
                step = next_step(store.snapshot(), 1)
                self.assertNotEqual(step["phase"], "complete")
                self.accept_step(store, step, "revise")
            self.assertEqual(next_step(store.snapshot(), 1), {"phase": "complete"})
            editorial = [r for r in store.snapshot()["accepted"]
                         if "decision" in r["submission"]]
            self.assertEqual(len(editorial), 12)
            self.assertTrue(all(r["submission"]["decision"] == "revise" for r in editorial))

    def test_pending_position_is_not_skipped(self):
        with tempfile.TemporaryDirectory() as directory:
            with RunStore(directory) as store:
                step = next_step(store.snapshot(), 1)
                store.prepare("author", "write", 0, 1, 1, "Pending")
            with RunStore(directory) as store:
                self.assertEqual(next_step(store.snapshot(), 1), step)
                bad = store.snapshot()
                bad["pending"]["chapter"] = 2
                with self.assertRaisesRegex(StateError, "schedule"):
                    next_step(bad, 1)

    def test_skipping_a_review_refuses(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            for _ in range(2):
                self.accept_step(store, next_step(store.snapshot(), 1))
            # Ledger validation permits phases; book ordering must reject this
            # global result because the chapter detail review has not happened.
            self.accept_step(store, {"phase": "global_review", "round": 1,
                                     "chapter": 0, "turn": 1})
            with self.assertRaisesRegex(StateError, "schedule"):
                next_step(store.snapshot(), 1)

    def test_no_fourth_round_and_invalid_chapter_count(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            for value in (0, -1, True, 1.5):
                with self.assertRaises(StateError):
                    next_step(store.snapshot(), value)
            while True:
                step = next_step(store.snapshot(), 1)
                if step["phase"] == "complete":
                    break
                self.accept_step(store, step)
            store.prepare("extra", "detail_review", 4, 1, 1, "Not authorized")
            with self.assertRaisesRegex(StateError, "completed third"):
                next_step(store.snapshot(), 1)


if __name__ == "__main__":
    unittest.main()
