"""Real-file evidence for stable review inputs and refused partial publication."""
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/fiction-workflow"))
from review_inputs import freeze_round
from run_state import RunStore, StateError


class ReviewInputTests(unittest.TestCase):
    def test_reopen_keeps_original_bytes_and_order(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "a.md").write_text("First café", encoding="utf-8")
            (root / "b.md").write_text("Second", encoding="utf-8")
            with RunStore(directory) as store:
                paths = freeze_round(store, 1, ["a.md", "b.md"], 1024)
            (root / "a.md").write_text("Revised", encoding="utf-8")
            with RunStore(directory) as store:
                self.assertEqual(freeze_round(store, 1, ["a.md", "b.md"], 1024), paths)
            text = (root / paths["manuscript"]).read_text(encoding="utf-8")
            self.assertEqual(text, "# Chapter 1\n\nFirst café\n\n# Chapter 2\n\nSecond\n\n")

    def test_corruption_and_parameter_changes_refuse(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            (store.directory / "a.md").write_text("A")
            (store.directory / "b.md").write_text("B")
            paths = freeze_round(store, 1, ["a.md", "b.md"], 1024)
            for chapters, limit in [(["b.md", "a.md"], 1024), (["a.md", "b.md"], 2048)]:
                with self.assertRaisesRegex(StateError, "parameters"):
                    freeze_round(store, 1, chapters, limit)
            (store.directory / paths["chapters"][0]).write_text("Changed")
            with self.assertRaisesRegex(StateError, "changed"):
                freeze_round(store, 1, ["a.md", "b.md"], 1024)

    def test_failed_assembly_publishes_nothing(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            (store.directory / "a.md").write_text("A" * 20)
            with self.assertRaisesRegex(StateError, "Assembled"):
                freeze_round(store, 1, ["a.md"], 20)
            self.assertFalse((store.directory / "review-inputs-1").exists())
            self.assertEqual(list(store.directory.glob(".review-inputs-*")), [])
            self.assertIsNone(store.snapshot()["pending"])

    def test_symlink_and_invalid_text_refuse(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            (store.directory / "real.md").write_text("Real")
            (store.directory / "a.md").symlink_to("real.md")
            with self.assertRaisesRegex(StateError, "link"):
                freeze_round(store, 1, ["a.md"], 1024)
            (store.directory / "a.md").unlink()
            (store.directory / "a.md").write_bytes(b"\xff")
            with self.assertRaises(UnicodeDecodeError):
                freeze_round(store, 1, ["a.md"], 1024)
            self.assertFalse((store.directory / "review-inputs-1").exists())


if __name__ == "__main__":
    unittest.main()
