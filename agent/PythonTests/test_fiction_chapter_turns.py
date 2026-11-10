"""Exercise chapter sequencing with actual files and a cooperative writing peer."""
import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills/fiction-workflow"))
from chapter_turns import advance_chapter
from run_state import RunStore, StateError


def reply(value):
    text = value if type(value) is str else json.dumps(value)
    return {"content": [{"type": "text", "text": text}], "isError": False}


class WritingPeer:
    """Writes drafts, requests one revision, then approves; retains saved results."""
    def __init__(self, root):
        self.root = root
        self.saved = {}
        self.admissions = []
        self.lose_reply = False
        self.mutate_draft = False
        self.approve_on_turn = 4
        self.invalid_artifact = None

    async def call_tool(self, name, arguments):
        if name == "agents_status":
            return reply({"agents": [{"agent_id": key, "activity": "idle",
                                      "paused": False, "failure": None}
                                     for key in self.saved]})
        if name == "get_submission":
            return reply({"agent_id": arguments["agent_id"], "historical": True,
                          "payload": self.saved[arguments["agent_id"]]})
        if name not in ("spawn_sub_agent", "ask_agent"):
            raise AssertionError(name)
        pending = json.loads((self.root / "state.json").read_text())["pending"]
        identity = arguments["agent_id"]
        self.admissions.append((identity, name))
        payload = {"assignment_id": pending["id"]}
        if identity == "author":
            (self.root / "chapter.md").write_text("Draft " + str(pending["turn"]))
            (self.root / "carryover.md").write_text("Continuity " + str(pending["turn"]))
        else:
            payload["decision"] = ("approve" if pending["turn"] >= self.approve_on_turn
                                   else "revise")
            (self.root / "feedback.md").write_text(payload["decision"])
            if self.mutate_draft:
                (self.root / "chapter.md").write_text("Unreviewed replacement")
        if self.invalid_artifact is not None:
            artifact_name, data = self.invalid_artifact
            (self.root / artifact_name).write_bytes(data)
        self.saved[identity] = payload
        if self.lose_reply:
            self.lose_reply = False
            raise ConnectionError("Reply lost after the files and submission were written")
        if name == "spawn_sub_agent":
            return reply(json.dumps({"agent_id": identity, "task_admitted": True}) + "\n"
                         + json.dumps({"agent_id": identity, "task_completed": True,
                                       "submitted": True, "payload": payload}))
        return reply({"agent_id": identity, "response": "Submitted"})


class ChapterTurnsTests(unittest.IsolatedAsyncioTestCase):
    async def advance(self, peer, store, review_round=0):
        return await advance_chapter(
            peer, store, chapter=1, author_id="author", editor_id="editor",
            author_task="Read the supplied inputs; write chapter.md and carryover.md.",
            editor_task="Read the draft and write feedback.md; revise or approve.",
            manuscript="chapter.md", carryover="carryover.md", feedback="feedback.md",
            review_round=review_round)

    async def test_revision_approval_and_repeated_resume(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = WritingPeer(Path(directory))
            # Reopen after every turn to prove progression comes from receipts.
            for expected in ("advanced", "advanced", "advanced", "approved"):
                with RunStore(directory) as store:
                    self.assertEqual((await self.advance(peer, store))["action"], expected)
            self.assertEqual(peer.admissions, [("author", "spawn_sub_agent"),
                ("editor", "spawn_sub_agent"), ("author", "ask_agent"), ("editor", "ask_agent")])
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"], "approved")
                self.assertEqual(len(store.snapshot()["accepted"]), 4)
            self.assertEqual(len(peer.admissions), 4)

    async def test_three_negative_reviews_stop_across_restart(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = WritingPeer(Path(directory))
            peer.approve_on_turn = 100
            for turn in range(1, 7):
                with RunStore(directory) as store:
                    result = await self.advance(peer, store)
                    expected = "revision_limit" if turn == 6 else "advanced"
                    self.assertEqual(result["action"], expected)
            with RunStore(directory) as store:
                before = store.snapshot()
                self.assertEqual((await self.advance(peer, store))["action"],
                                 "revision_limit")
                self.assertEqual(store.snapshot(), before)
                self.assertEqual(before["accepted"][-1]["submission"]["decision"],
                                 "revise")
            self.assertEqual(len(peer.admissions), 6)
            self.assertEqual((Path(directory) / "feedback.md").read_text(), "revise")

    async def test_third_attempt_can_be_approved(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = WritingPeer(Path(directory))
            peer.approve_on_turn = 6
            for _ in range(6):
                result = await self.advance(peer, store)
            self.assertEqual(result["action"], "approved")
            self.assertEqual((await self.advance(peer, store))["action"], "approved")
            self.assertEqual(len(peer.admissions), 6)

    async def test_lost_final_review_reconciles_to_limit(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = WritingPeer(Path(directory))
            peer.approve_on_turn = 100
            with RunStore(directory) as store:
                for _ in range(5):
                    await self.advance(peer, store)
                peer.lose_reply = True
                self.assertEqual((await self.advance(peer, store))["action"], "supervise")
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"],
                                 "revision_limit")
                self.assertIsNone(store.snapshot()["pending"])
            self.assertEqual(len(peer.admissions), 6)

    async def test_review_rounds_have_independent_attempt_counts(self):
        # This peer models result delivery only. Fresh native child homes remain
        # the caller's responsibility, tested separately by role preparation.
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = WritingPeer(Path(directory))
            peer.approve_on_turn = 100
            for review_round in range(4):
                for _ in range(6):
                    result = await self.advance(peer, store, review_round)
                self.assertEqual(result["action"], "revision_limit")
            self.assertEqual(len(store.snapshot()["accepted"]), 24)
            before = len(peer.admissions)
            with self.assertRaisesRegex(StateError, "round"):
                await self.advance(peer, store, 4)
            self.assertEqual(len(peer.admissions), before)

    async def test_lost_reply_recovers_result_without_new_turn(self):
        with tempfile.TemporaryDirectory() as directory:
            peer = WritingPeer(Path(directory))
            peer.lose_reply = True
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"], "supervise")
                self.assertIsNotNone(store.snapshot()["pending"])
            with RunStore(directory) as store:
                self.assertEqual((await self.advance(peer, store))["action"], "advanced")
                self.assertIsNone(store.snapshot()["pending"])
            self.assertEqual(len(peer.admissions), 1)

    async def test_editor_cannot_approve_or_request_revision_of_changed_draft(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = WritingPeer(Path(directory))
            await self.advance(peer, store)
            peer.mutate_draft = True
            with self.assertRaisesRegex(StateError, "draft changed"):
                await self.advance(peer, store)
            self.assertEqual(len(store.snapshot()["accepted"]), 1)
            self.assertIsNotNone(store.snapshot()["pending"])

    async def test_invalid_writing_text_never_commits_or_redispatches(self):
        for name in ("chapter.md", "carryover.md", "feedback.md"):
            for data in (b"", b" \n\t", b"\xff"):
                for lost_reply in (False, True):
                    with self.subTest(name=name, data=data, recovered=lost_reply):
                        with tempfile.TemporaryDirectory() as directory:
                            peer = WritingPeer(Path(directory))
                            peer.approve_on_turn = 100
                            with RunStore(directory) as store:
                                # Exercise the final negative Editor decision too:
                                # it must not finish the chapter with unusable criticism.
                                if name == "feedback.md":
                                    for _ in range(5):
                                        await self.advance(peer, store)
                                accepted = store.snapshot()["accepted"]
                                peer.invalid_artifact = (name, data)
                                peer.lose_reply = lost_reply
                                if lost_reply:
                                    self.assertEqual((await self.advance(peer, store))["action"],
                                                     "supervise")
                                else:
                                    with self.assertRaises(StateError):
                                        await self.advance(peer, store)
                                pending = store.snapshot()["pending"]
                                self.assertIsNotNone(pending)
                                self.assertEqual(store.snapshot()["accepted"], accepted)
                                admissions = len(peer.admissions)
                            # Reopen and validate the retained result again. Invalid
                            # output must not erase intent or dispatch replacement work.
                            with RunStore(directory) as store:
                                before = store.snapshot()
                                with self.assertRaises(StateError):
                                    await self.advance(peer, store)
                                self.assertEqual(store.snapshot(), before)
                                self.assertEqual(len(peer.admissions), admissions)

    async def test_missing_required_file_keeps_pending(self):
        with tempfile.TemporaryDirectory() as directory, RunStore(directory) as store:
            peer = WritingPeer(Path(directory))
            peer.lose_reply = True
            await self.advance(peer, store)
            (Path(directory) / "carryover.md").unlink()
            with self.assertRaises(FileNotFoundError):
                await self.advance(peer, store)
            self.assertEqual(store.snapshot()["accepted"], [])
            self.assertEqual(len(peer.admissions), 1)


if __name__ == "__main__":
    unittest.main()
