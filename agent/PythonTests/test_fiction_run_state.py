"""Real-file tests for the fiction assignment ledger; no inference or network."""
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parents[1] / "skills/fiction-workflow/run_state.py"
spec = importlib.util.spec_from_file_location("fiction_run_state", SOURCE)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
RunStore = module.RunStore
StateError = module.StateError


class Ledger(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="fiction-ledger-")
        self.root = Path(self.temporary.name)

    def tearDown(self):
        self.temporary.cleanup()

    def prepare(self, store):
        return store.prepare("author-ch1", "write", 0, 1, 1, "Write the first chapter")

    def test_intent_survives_restart_and_cannot_be_repeated(self):
        with RunStore(self.root) as store:
            assignment = self.prepare(store)
        # The same state is retained whether dispatch never happened or its
        # callback was lost. This module cannot distinguish those remote facts.
        with RunStore(self.root) as restored:
            self.assertEqual(restored.snapshot()["pending"], assignment)
            with self.assertRaisesRegex(StateError, "Reconcile"):
                self.prepare(restored)
            self.assertEqual(restored.snapshot()["accepted"], [])

    def test_result_before_advancement_is_reconciled_once(self):
        with RunStore(self.root) as store:
            assignment = self.prepare(store)
        (self.root / "chapter.md").write_text("A complete test chapter.\n")
        submission = {"assignment_id": assignment["id"], "result": "written"}
        # Simulate a durable child result arriving before workflow advancement.
        (self.root / "child-result.json").write_text(json.dumps(submission))
        with RunStore(self.root) as restored:
            saved = json.loads((self.root / "child-result.json").read_text())
            restored.accept(saved, ["chapter.md"])
        with RunStore(self.root) as restored:
            state = restored.snapshot()
            self.assertIsNone(state["pending"])
            self.assertEqual(len(state["accepted"]), 1)
            self.assertEqual(state["accepted"][0]["submission"], submission)
            with self.assertRaisesRegex(StateError, "No pending"):
                restored.accept(submission, ["chapter.md"])
            next_assignment = restored.prepare("editor-ch1", "write", 0, 1, 2, "Review chapter")
            self.assertNotEqual(next_assignment["id"], assignment["id"])
            with self.assertRaisesRegex(StateError, "does not match"):
                restored.accept(submission, ["chapter.md"])

    def test_missing_outputs_and_wrong_identity_leave_state_unchanged(self):
        with RunStore(self.root) as store:
            assignment = self.prepare(store)
            before = (self.root / "state.json").read_bytes()
            with self.assertRaisesRegex(StateError, "does not match"):
                store.accept({"assignment_id": "stale"}, ["chapter.md"])
            with self.assertRaises(FileNotFoundError):
                store.accept({"assignment_id": assignment["id"]}, ["chapter.md"])
            self.assertEqual((self.root / "state.json").read_bytes(), before)
            self.assertEqual(store.snapshot()["pending"], assignment)

    def test_snapshot_is_not_mutable_authority(self):
        with RunStore(self.root) as store:
            assignment = self.prepare(store)
            assignment["task"] = "changed"
            snapshot = store.snapshot()
            snapshot["pending"]["task"] = "also changed"
            self.assertEqual(store.snapshot()["pending"]["task"], "Write the first chapter")

    def test_second_process_refuses_lock_then_can_reopen(self):
        program = '''import importlib.util, sys
spec = importlib.util.spec_from_file_location("state", sys.argv[1])
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)
try:
    with m.RunStore(sys.argv[2]):
        print("opened")
except BlockingIOError:
    print("locked")
'''
        def probe():
            result = subprocess.run([sys.executable, "-c", program, str(SOURCE), str(self.root)],
                                    capture_output=True, text=True, timeout=10, check=True)
            self.assertEqual(result.stderr, "")
            return result.stdout.strip()
        with RunStore(self.root):
            self.assertEqual(probe(), "locked")
        self.assertEqual(probe(), "opened")

    def test_malformed_state_is_preserved_and_lock_released(self):
        path = self.root / "state.json"
        for contents in ('{"version":1,"version":1}', '{"version":99}',
                         '{"version":NaN}', 'not JSON'):
            path.write_text(contents)
            with self.assertRaises((StateError, json.JSONDecodeError)):
                with RunStore(self.root):
                    self.fail("Malformed state was accepted")
            self.assertEqual(path.read_text(), contents)
        path.unlink()
        with RunStore(self.root) as store:
            self.assertIsNone(store.snapshot()["pending"])

    def test_real_save_failure_closes_admission_and_cleans_temporary(self):
        with RunStore(self.root) as store:
            # A directory at the destination obstructs atomic replacement.
            # No mock of os.replace manufactures the failure.
            state = self.root / "state.json"
            state.rename(self.root / "original-state.json")
            state.mkdir()
            with self.assertRaises(OSError):
                self.prepare(store)
            with self.assertRaisesRegex(StateError, "closed or"):
                store.snapshot()
            self.assertEqual(list(self.root.glob(".state-*")), [])
            state.rmdir()
            (self.root / "original-state.json").rename(state)
        with RunStore(self.root) as restored:
            self.assertIsNone(restored.snapshot()["pending"])

    def test_escape_and_boolean_position_refuse(self):
        with RunStore(self.root) as store:
            with self.assertRaisesRegex(StateError, "position"):
                store.prepare("author", "write", False, 1, 1, "task")
            assignment = self.prepare(store)
            with self.assertRaisesRegex(StateError, "relative run path"):
                store.accept({"assignment_id": assignment["id"]}, ["../outside.md"])


if __name__ == "__main__":
    unittest.main()
