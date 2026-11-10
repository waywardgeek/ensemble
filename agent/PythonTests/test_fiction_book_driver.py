"""Whole-book local peer tests: real files, receipts, role transfer and recovery.

The peer supplies deterministic writing, not literary judgment or native Agent
behavior. Native reverse-MCP and role-context evidence remain separate checks.
"""
import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/fiction-workflow"))
from book_driver import advance_book, bind_plan
from run_state import RunStore, StateError


def reply(value):
    return {"content": [{"type": "text", "text": value if isinstance(value, str)
                         else json.dumps(value)}], "isError": False}


class BookPeer:
    def __init__(self, root, homes):
        self.root = root
        self.homes = homes
        self.saved = {}
        self.live = set()
        self.admissions = []
        self.stops = []
        self.lose_reply = False
        self.lose_shutdown = False
        self.approve = True

    async def call_tool(self, name, args):
        identity = args.get("agent_id")
        if name == "shutdown_agent":
            if identity not in self.live:
                return {"isError": True, "content": [{"type": "text", "text": "Missing child"}]}
            self.live.remove(identity)
            self.stops.append(identity)
            if self.lose_shutdown:
                self.lose_shutdown = False
                raise ConnectionError("Lost joined shutdown reply")
            return reply({"agent_id": identity, "stopped": True})
        if name == "agents_status":
            return reply({"agents": [{"agent_id": child, "activity": "idle",
                                      "paused": False, "failure": None}
                                     for child in self.live]})
        if name == "get_submission":
            return reply({"agent_id": identity, "historical": True,
                          "payload": self.saved[identity]})
        if name not in ("spawn_sub_agent", "ask_agent"):
            raise AssertionError(name)
        if name == "spawn_sub_agent":
            if identity in self.live:
                raise AssertionError("Duplicate child creation")
            self.live.add(identity)
        elif identity not in self.live:
            raise AssertionError("Continued turn lost its child")
        pending = json.loads((self.root / "state.json").read_text())["pending"]
        assert pending["child_id"] == identity
        self.admissions.append(pending)
        payload = {"assignment_id": pending["id"]}
        chapter = pending["chapter"]
        if pending["phase"] in ("write", "revise"):
            stem = self.root / "chapters" / ("%04d" % chapter)
            if pending["turn"] % 2:
                stem.with_suffix(".md").write_text("Draft " + pending["id"])
                Path(str(stem) + "-carryover.md").write_text("Continuity")
            else:
                payload["decision"] = "approve" if self.approve else "revise"
                Path(str(stem) + "-editor.md").write_text("Specific feedback")
            with (self.homes / identity / "MEMORY.md").open("a") as memory:
                memory.write("Learned from " + pending["id"] + "\n")
        else:
            output = self.root / "reviews" / ("r%d-c%d.md" % (pending["round"], chapter))
            output.write_text("Independent review with passage references")
        self.saved[identity] = payload
        if self.lose_reply:
            self.lose_reply = False
            raise ConnectionError("Lost completed assignment reply")
        if name == "spawn_sub_agent":
            return reply(json.dumps({"agent_id": identity, "task_admitted": True}) + "\n"
                         + json.dumps({"agent_id": identity, "task_completed": True,
                                       "submitted": True, "payload": payload}))
        return reply({"agent_id": identity, "response": "Submitted"})


class BookDriverTests(unittest.IsolatedAsyncioTestCase):
    def setup_run(self, root, count=2):
        homes = root / "children"
        homes.mkdir()
        baseline = root / "baseline"
        for role in ("author", "editor"):
            directory = baseline / role
            directory.mkdir(parents=True)
            (directory / "SOUL.md").write_text("Role " + role)
            (directory / "MEMORY.md").write_bytes(b"")
        (root / "outline.md").write_text("Supplied outline")
        (root / "voice.md").write_text("Supplied voice")
        with RunStore(root) as store:
            plan = bind_plan(store, chapter_count=count, outline="outline.md", voice="voice.md",
                             baseline_root=baseline, children_directory=homes,
                             source_limit=65536, result_limit=131072)
        return plan, BookPeer(root, homes)

    async def test_whole_book_reopen_every_step_and_join_all(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            for _ in range(26):
                with RunStore(root) as store:
                    result = await advance_book(peer, store, plan)
            self.assertEqual(result, {"action": "complete"})
            self.assertEqual(len(peer.admissions), 25)
            final = (root / "final-book/manuscript.md").read_text()
            for chapter in (1, 2):
                self.assertIn((root / ("chapters/%04d.md" % chapter)).read_text(), final)
            self.assertNotEqual(final, (root / "review-inputs-3/manuscript.md").read_text())
            self.assertEqual(peer.live, set())
            self.assertEqual(len(peer.stops), len(set(peer.stops)))
            authors = [a for a in peer.admissions if a["child_id"].endswith("author")]
            first = peer.homes / authors[0]["child_id"] / "MEMORY.md"
            last = peer.homes / authors[-1]["child_id"] / "MEMORY.md"
            self.assertIn(first.read_text(), last.read_text())
            self.assertNotEqual(first.parent, last.parent)
            for number in (1, 2, 3):
                self.assertTrue((root / ("review-inputs-%d/manuscript.md" % number)).is_file())
            with RunStore(root) as store:
                self.assertEqual(await advance_book(peer, store, plan), {"action": "complete"})
            self.assertEqual(len(peer.admissions), 25)

    async def test_changed_revision_draft_after_lost_reply_uses_saved_task(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root, 1)
            with RunStore(root) as store:
                for _ in range(4):
                    await advance_book(peer, store, plan)
                # Initial pair and both reviews are accepted. The revision changes
                # a file which was itself an input to this pending assignment.
                peer.lose_reply = True
                self.assertEqual((await advance_book(peer, store, plan))["action"], "supervise")
                pending = store.snapshot()["pending"]
                self.assertEqual(pending["phase"], "revise")
            with RunStore(root) as store:
                self.assertEqual((await advance_book(peer, store, plan))["action"], "advanced")
                self.assertIsNone(store.snapshot()["pending"])
            self.assertEqual(len(peer.admissions), 5)

    async def test_uncertain_shutdown_stops_without_memory_transfer(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            with RunStore(root) as store:
                await advance_book(peer, store, plan)
                await advance_book(peer, store, plan)
                peer.lose_shutdown = True
                before = store.snapshot()
                self.assertEqual((await advance_book(peer, store, plan))["action"], "supervise")
                self.assertEqual(store.snapshot(), before)
                self.assertEqual((await advance_book(peer, store, plan))["action"], "supervise")
            self.assertEqual(len(peer.admissions), 2)
            # The hexadecimal run ID may itself contain "c2". Match the
            # delimited chapter component, not an incidental UUID substring.
            self.assertFalse(any("-c2-" in path.name for path in peer.homes.iterdir()))

    async def test_changed_outline_and_missing_input_refuse_before_dispatch(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            (root / "outline.md").write_text("Different outline")
            with RunStore(root) as store:
                with self.assertRaisesRegex(StateError, "changed"):
                    await advance_book(peer, store, plan)
            self.assertEqual(peer.admissions, [])
            (root / "outline.md").unlink()
            with RunStore(root) as store:
                with self.assertRaises(FileNotFoundError):
                    await advance_book(peer, store, plan)
            self.assertEqual(peer.admissions, [])


if __name__ == "__main__":
    unittest.main()
