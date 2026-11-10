"""Supervised recovery tests: real durable ledger, small read-only child peer."""
import json
import pathlib
import sys
import tempfile
import unittest

DIRECTORY = pathlib.Path(__file__).resolve().parents[1] / "skills/fiction-workflow"
sys.path.insert(0, str(DIRECTORY))
from recovery import reconcile
from run_state import RunStore


class ChildrenPeer:
    def __init__(self, activity="idle", payload=None, present=True):
        self.activity = activity
        self.payload = payload
        self.present = present
        self.calls = []

    async def call_tool(self, name, arguments):
        self.calls.append(name)
        if name == "agents_status":
            agents = []
            if self.present:
                agents.append({"agent_id": "author-1", "activity": self.activity,
                               "paused": False, "failure": None, "submitted": False})
            value = {"agents": agents}
        elif name == "get_submission":
            if self.payload is None:
                return {"content": [{"type": "text", "text": "Missing artifact"}],
                        "isError": True}
            value = {"agent_id": "author-1", "historical": True, "generation": 1,
                     "payload": self.payload}
        else:
            raise AssertionError("Recovery attempted a non-read-only operation")
        return {"content": [{"type": "text", "text": json.dumps(value)}]}


class RecoveryTests(unittest.IsolatedAsyncioTestCase):
    async def test_pending_recovery_preserves_bytes(self):
        cases = [("busy", None, True, "observe"),
                 ("idle", None, True, "supervise"),
                 ("idle", {"assignment_id": "old"}, True, "supervise"),
                 ("stopping", None, True, "supervise"),
                 ("idle", None, False, "supervise")]
        for activity, payload, present, expected in cases:
            with self.subTest(activity=activity, payload=payload, present=present):
                with tempfile.TemporaryDirectory() as directory:
                    with RunStore(directory) as store:
                        store.prepare("author-1", "write", 0, 1, 0, "Write chapter one")
                    path = pathlib.Path(directory) / "state.json"
                    before = path.read_bytes()
                    peer = ChildrenPeer(activity, payload, present)
                    with RunStore(directory) as store:
                        result = await reconcile(peer, store)
                    self.assertEqual(result["action"], expected)
                    self.assertEqual(path.read_bytes(), before)
                    self.assertTrue(set(peer.calls) <= {"agents_status", "get_submission"})

    async def test_matching_result_is_candidate_not_automatic_acceptance(self):
        for present in (True, False):
            with self.subTest(present=present), tempfile.TemporaryDirectory() as directory:
                with RunStore(directory) as store:
                    assignment = store.prepare("author-1", "write", 0, 1, 0, "Write chapter one")
                payload = {"assignment_id": assignment["id"], "draft": "chapter.md"}
                with RunStore(directory) as store:
                    outcome = await reconcile(ChildrenPeer(payload=payload, present=present), store)
                    self.assertEqual(outcome["action"], "validate_result")
                    self.assertEqual(outcome["submission"], payload)
                    self.assertEqual(store.snapshot()["pending"], assignment)
                    self.assertEqual(store.snapshot()["accepted"], [])

    async def test_no_pending_assignment_needs_no_remote_calls(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = ChildrenPeer()
            self.assertEqual(await reconcile(peer, store), {"action": "ready"})
            self.assertEqual(peer.calls, [])


if __name__ == "__main__":
    unittest.main()
