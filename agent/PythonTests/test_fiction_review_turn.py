"""Review transport bounds and durable result handling, not literary grading."""
import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/fiction-workflow"))
from review_turn import advance_review, source_info
from run_state import RunStore, StateError


def reply(value):
    return {"content": [{"type": "text", "text": value if isinstance(value, str)
                          else json.dumps(value)}], "isError": False}


class ReviewPeer:
    """Small child protocol peer: real report file, retained typed submission."""
    def __init__(self, root):
        self.root = root
        self.payload = None
        self.admissions = 0
        self.spawn_arguments = []
        self.lose_reply = False
        self.change_source = False
        self.report = "Specific criticism of the fixture passage."

    async def call_tool(self, name, arguments):
        if name == "agents_status":
            return reply({"agents": []})
        if name == "get_submission":
            return reply({"agent_id": "reviewer", "historical": True,
                          "payload": self.payload})
        if name != "spawn_sub_agent":
            raise AssertionError("Unexpected child operation: " + name)
        self.admissions += 1
        self.spawn_arguments.append(arguments)
        pending = json.loads((self.root / "state.json").read_text())["pending"]
        self.payload = {"assignment_id": pending["id"]}
        (self.root / "review.md").write_text(self.report, encoding="utf-8")
        if self.change_source:
            (self.root / "chapter.md").write_text("Different source")
        if self.lose_reply:
            raise ConnectionError("Completed result; reply lost")
        return reply(json.dumps({"agent_id": "reviewer", "task_admitted": True})
                     + "\n" + json.dumps({"agent_id": "reviewer", "task_completed": True,
                                          "submitted": True, "payload": self.payload}))


class ReviewTurnTests(unittest.IsolatedAsyncioTestCase):
    async def advance(self, peer, store, **overrides):
        options = dict(phase="detail_review", review_round=1, chapter=1,
                       child_id="reviewer", source="chapter.md", output="review.md",
                       guidance="Assess concrete prose and continuity.",
                       source_limit=10000, result_limit=20000)
        options.update(overrides)
        return await advance_review(peer, store, **options)

    def seed(self, directory):
        root = Path(directory)
        (root / "chapter.md").write_bytes("é\r\nline two\n".encode())
        return ReviewPeer(root)

    async def test_exact_source_and_numbered_result_bounds(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            self.seed(directory)
            info = source_info(store, "chapter.md", 100, 10000)
            size = len((Path(directory) / "chapter.md").read_bytes())
            self.assertEqual(info["bytes"], size)
            self.assertEqual(info["read"]["max_bytes"], size)
            self.assertEqual(info["read"]["max_output_bytes"], size + 6 + 4096)
            self.assertEqual(info["read"]["start_line"], 1)
            self.assertEqual(info["read"]["end_line"], 0)
            with self.assertRaisesRegex(StateError, "source byte"):
                source_info(store, "chapter.md", size - 1, 10000)
            with self.assertRaisesRegex(StateError, "result byte"):
                source_info(store, "chapter.md", size, size + 6 + 4095)

    async def test_review_and_repeat_do_not_dispatch_twice(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = self.seed(directory)
            self.assertEqual((await self.advance(peer, store))["action"], "reviewed")
            self.assertEqual((await self.advance(peer, store))["action"], "reviewed")
            self.assertEqual(peer.admissions, 1)
            task = store.snapshot()["accepted"][0]["assignment"]["task"]
            self.assertIn("read_file exactly once", task)
            self.assertIn("model capacity", task)
            self.assertIn("do not submit a completed review", task)

    async def test_lost_reply_reopens_without_redispatch(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = self.seed(directory)
            peer.lose_reply = True
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"], "supervise")
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"], "reviewed")
            self.assertEqual(peer.admissions, 1)

    async def test_only_global_review_requests_large_working_context(self):
        for phase, chapter in (("detail_review", 1), ("global_review", 0)):
            with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
                peer = self.seed(directory)
                peer.lose_reply = True
                self.assertEqual((await self.advance(peer, store, phase=phase,
                                                    chapter=chapter))["action"], "supervise")
                arguments = peer.spawn_arguments[0]
                if phase == "global_review":
                    self.assertEqual(arguments["working_context_tokens"], 700_000)
                else:
                    self.assertNotIn("working_context_tokens", arguments)
                self.assertEqual((await self.advance(peer, store, phase=phase,
                                                    chapter=chapter))["action"], "reviewed")
                self.assertEqual(peer.admissions, 1)

    async def test_changed_source_keeps_pending(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = self.seed(directory)
            peer.change_source = True
            with self.assertRaisesRegex(StateError, "changed during"):
                await self.advance(peer, store)
            self.assertIsNotNone(store.snapshot()["pending"])
            with self.assertRaisesRegex(StateError, "inputs changed"):
                await self.advance(peer, store)
            self.assertEqual(peer.admissions, 1)

    async def test_empty_report_and_oversize_input_refuse(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = self.seed(directory)
            with self.assertRaisesRegex(StateError, "source byte"):
                await self.advance(peer, store, source_limit=1)
            self.assertEqual(peer.admissions, 0)
            self.assertIsNone(store.snapshot()["pending"])
            peer.report = " \n"
            with self.assertRaisesRegex(StateError, "empty"):
                await self.advance(peer, store)
            self.assertIsNotNone(store.snapshot()["pending"])
            self.assertEqual(store.snapshot()["accepted"], [])


if __name__ == "__main__":
    unittest.main()
