"""Real ledger plus a small child protocol fake; no inference or native admission."""
import json
from pathlib import Path
import sys
import tempfile
import unittest

DIRECTORY = Path(__file__).resolve().parents[1] / "skills/fiction-workflow"
sys.path.insert(0, str(DIRECTORY))
from assignments import dispatch
from child_calls import ObservationError
from run_state import RunStore, StateError


def reply(value):
    text = value if type(value) is str else json.dumps(value)
    return {"content": [{"type": "text", "text": text}], "isError": False}


class ChildPeer:
    """Retains one submission and admits tasks only after persisted intent."""
    def __init__(self, directory, *, lose_reply=False, stale=False):
        self.directory = directory
        self.lose_reply = lose_reply
        self.stale = stale
        self.calls = []
        self.payload = None

    async def call_tool(self, name, arguments):
        self.calls.append(name)
        pending = json.loads((self.directory / "state.json").read_text())["pending"]
        if name in ("spawn_sub_agent", "ask_agent"):
            instruction = arguments.get("task", arguments.get("content"))
            assert pending["id"] in instruction
            assert pending["child_id"] == arguments["agent_id"]
            self.payload = {"assignment_id": "older" if self.stale else pending["id"]}
            if self.lose_reply:
                raise ConnectionError("Admission reply lost")
            if name == "spawn_sub_agent":
                return reply(json.dumps({"agent_id": arguments["agent_id"],
                                         "task_admitted": True}) + "\n" + json.dumps({
                    "agent_id": arguments["agent_id"], "task_completed": True,
                    "submitted": True, "payload": self.payload}))
            return reply({"agent_id": arguments["agent_id"], "response": "Submitted"})
        if name == "get_submission":
            return reply({"agent_id": arguments["agent_id"], "historical": True,
                          "payload": self.payload})
        raise AssertionError("Unexpected operation: " + name)


class AssignmentTests(unittest.IsolatedAsyncioTestCase):
    def options(self, new=True):
        value = dict(child_id="author-1", phase="write", round_number=0,
                     chapter=1, turn=1, task="Write the supplied chapter.", new_child=new)
        if new:
            value["submit_schema"] = {"type": "object"}
        return value

    async def test_new_child_result_remains_pending_until_validated(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = ChildPeer(Path(directory))
            payload = await dispatch(peer, store, **self.options())
            self.assertEqual(peer.calls, ["spawn_sub_agent"])
            self.assertEqual(payload["assignment_id"], store.snapshot()["pending"]["id"])
            self.assertEqual(store.snapshot()["accepted"], [])

    async def test_existing_role_gets_new_turn_and_matching_saved_result(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = ChildPeer(Path(directory))
            await dispatch(peer, store, **self.options(False))
            self.assertEqual(peer.calls, ["ask_agent", "get_submission"])

    async def test_lost_reply_persists_intent_and_blocks_redispatch(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = ChildPeer(Path(directory), lose_reply=True)
            with RunStore(directory) as store:
                with self.assertRaises(ConnectionError):
                    await dispatch(peer, store, **self.options())
                identity = store.snapshot()["pending"]["id"]
            with RunStore(directory) as store:
                self.assertEqual(store.snapshot()["pending"]["id"], identity)
                with self.assertRaises(StateError):
                    await dispatch(peer, store, **self.options())
            self.assertEqual(peer.calls, ["spawn_sub_agent"])

    async def test_stale_submission_never_accepted(self):
        for new in (True, False):
            with self.subTest(new=new), tempfile.TemporaryDirectory() as directory:
                with RunStore(directory) as store:
                    with self.assertRaises(ObservationError):
                        await dispatch(ChildPeer(Path(directory), stale=True), store,
                                       **self.options(new))
                    self.assertIsNotNone(store.snapshot()["pending"])
                    self.assertEqual(store.snapshot()["accepted"], [])

    async def test_invalid_context_policy_has_no_dispatch_effect(self):
        for value in (0, -1, True, 1.5, "700000"):
            with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
                peer = ChildPeer(Path(directory))
                with self.assertRaises(StateError):
                    await dispatch(peer, store, **self.options(), working_context_tokens=value)
                self.assertIsNone(store.snapshot()["pending"])
                self.assertEqual(peer.calls, [])

    async def test_missing_new_child_schema_has_no_effect(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = ChildPeer(Path(directory))
            options = self.options()
            del options["submit_schema"]
            with self.assertRaises(StateError):
                await dispatch(peer, store, **options)
            self.assertIsNone(store.snapshot()["pending"])
            self.assertEqual(peer.calls, [])


if __name__ == "__main__":
    unittest.main()
