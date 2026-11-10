"""Creative changes use the existing turn budget and receipt-based permissions."""
import json
from pathlib import Path
import tempfile
import unittest

import test_fiction_book_driver as fixtures
from test_fiction_book_driver import BookPeer, reply
from book_driver import advance_book, bind_plan
from run_state import RunStore, StateError
from story_changes import validate_submission
from review_inputs import freeze_round


class CreativePeer(BookPeer):
    mode = "normal"
    target = 2

    async def call_tool(self, name, args):
        result = await super().call_tool(name, args)
        if name not in ("spawn_sub_agent", "ask_agent"):
            return result
        pending = self.admissions[-1]
        if pending["phase"] not in ("write", "revise"):
            return result
        chapter, turn = pending["chapter"], pending["turn"]
        payload = self.saved[args["agent_id"]]
        if turn % 2:
            (self.root / "working-outline.md").write_text("Improved outline " + pending["id"])
            if chapter == self.target and turn == 1:
                payload["prior_chapter_requests"] = [1]
                if self.mode == "unauthorized":
                    (self.root / "chapters/0001.md").write_text("Unapproved change")
            if chapter == self.target and turn == 3:
                (self.root / "chapters/0001.md").write_text("Authorized repair")
                (self.root / "chapters/0001-carryover.md").write_text("Repaired continuity")
        elif chapter == self.target and turn == 2:
            payload.update(decision="revise", prior_chapter_permissions=[1])
            if self.mode == "denied":
                payload["prior_chapter_permissions"] = []
            if self.mode == "unrequested":
                payload["prior_chapter_permissions"] = [2]
        elif chapter == self.target and turn == 4 and self.mode in ("expired", "review_again"):
            payload["decision"] = "revise"
        elif chapter == self.target and turn == 4 and self.mode == "editor_mutation":
            (self.root / "chapters/0001.md").write_text("Edited during review")
        if self.mode == "expired" and chapter == self.target and turn == 5:
            (self.root / "chapters/0001.md").write_text("Permission was for an earlier turn")
        if self.mode == "lost" and chapter == self.target and turn == 3:
            self.mode = "normal"
            raise ConnectionError("Lost reply after authorized edits")
        if name == "spawn_sub_agent":
            return reply(json.dumps({"agent_id": args["agent_id"], "task_admitted": True}) + "\n" +
                         json.dumps({"agent_id": args["agent_id"], "task_completed": True,
                                     "submitted": True, "payload": payload}))
        return result


class CreativeTests(unittest.IsolatedAsyncioTestCase):
    def setup_run(self, root):
        plan, peer = fixtures.BookDriverTests().setup_run(root)
        return plan, CreativePeer(root, peer.homes)

    async def test_outline_and_editor_authorized_prior_repair_survive_reopen(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            for _ in range(6):
                with RunStore(root) as store:
                    await advance_book(peer, store, plan)
            self.assertEqual((root / "outline.md").read_text(), "Supplied outline")
            self.assertEqual((root / "chapters/0001.md").read_text(), "Authorized repair")
            with RunStore(root) as store:
                receipts = store.snapshot()["accepted"]
                self.assertIn("working-outline.md", receipts[0]["artifacts"])
                self.assertIn("chapters/0001.md", receipts[4]["artifacts"])
                self.assertIn("chapters/0001-carryover.md", receipts[4]["artifacts"])
                # Rebinding still pins original inputs, not the changed working outline.
                self.assertEqual(bind_plan(store, chapter_count=2, outline="outline.md",
                    voice="voice.md", baseline_root=root / "baseline",
                    children_directory=root / "children", source_limit=65536,
                    result_limit=131072), plan)
                await advance_book(peer, store, plan)
            editor_task = peer.admissions[5]["task"]
            self.assertIn(str(root / "chapters/0001.md"), editor_task)
            self.assertIn(str(root / "chapters/0001-carryover.md"), editor_task)
            frozen = root / "review-inputs-1/context-0001.md"
            self.assertEqual(frozen.read_text(), (root / "working-outline.md").read_text())

    async def test_unauthorized_prior_edit_preserves_pending(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            peer.mode = "unauthorized"
            with RunStore(root) as store:
                await advance_book(peer, store, plan)
                await advance_book(peer, store, plan)
                with self.assertRaisesRegex(StateError, "outside the authorized turn"):
                    await advance_book(peer, store, plan)
                self.assertEqual(len(store.snapshot()["accepted"]), 2)
                self.assertIsNotNone(store.snapshot()["pending"])

    async def test_authorized_edit_lost_reply_not_repeated(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            peer.mode = "lost"
            with RunStore(root) as store:
                for _ in range(4):
                    await advance_book(peer, store, plan)
                self.assertEqual((await advance_book(peer, store, plan))["action"], "supervise")
            with RunStore(root) as store:
                self.assertEqual((await advance_book(peer, store, plan))["action"], "advanced")
                self.assertIsNone(store.snapshot()["pending"])
            self.assertEqual(len(peer.admissions), 5)

    async def test_editor_cannot_change_authorized_prior_chapter(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            peer.mode = "editor_mutation"
            with RunStore(root) as store:
                for _ in range(5):
                    await advance_book(peer, store, plan)
                with self.assertRaises(StateError):
                    await advance_book(peer, store, plan)
                self.assertEqual(len(store.snapshot()["accepted"]), 5)

    async def test_invalid_permission_refuses(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            peer.mode = "unrequested"
            with RunStore(root) as store:
                for _ in range(3):
                    await advance_book(peer, store, plan)
                with self.assertRaises(StateError):
                    await advance_book(peer, store, plan)
                self.assertEqual(len(store.snapshot()["accepted"]), 3)

    def test_permission_validation_and_last_turn_deferral(self):
        request = {"assignment_id": "a", "prior_chapter_requests": [1]}
        validate_submission(request, False, 3, 5)
        for numbers, decision, turn, message in (
                ([1], "approve", 2, "remaining"),
                ([1], "revise", 6, "remaining"),
                ([2], "revise", 2, "explicit requests"),
                ([True], "revise", 2, "distinct"),
                ([1, 1], "revise", 2, "distinct"),
                ([3], "revise", 2, "distinct")):
            with self.subTest(numbers=numbers, decision=decision, turn=turn):
                with self.assertRaisesRegex(StateError, message):
                    validate_submission({"assignment_id": "e", "decision": decision,
                        "prior_chapter_permissions": numbers}, True, 3, turn, request)
        # Last-attempt requests may remain recorded, but no unused permission.
        validate_submission({"assignment_id": "e", "decision": "revise"},
                            True, 3, 6, request)

    async def test_denied_and_expired_permission_cannot_authorize_edits(self):
        for mode, accepted_count in (("denied", 4), ("expired", 6)):
            with self.subTest(mode=mode), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                plan, peer = self.setup_run(root)
                peer.mode = mode
                with RunStore(root) as store:
                    for _ in range(accepted_count):
                        await advance_book(peer, store, plan)
                    with self.assertRaisesRegex(StateError, "outside the authorized turn"):
                        await advance_book(peer, store, plan)
                    self.assertEqual(len(store.snapshot()["accepted"]), accepted_count)
                    self.assertIsNotNone(store.snapshot()["pending"])

    async def test_working_outline_cannot_change_between_assignments(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            (root / "working-outline.md").write_text("Unaccepted change")
            with RunStore(root) as store:
                with self.assertRaisesRegex(StateError, "outside the authorized turn"):
                    await advance_book(peer, store, plan)
            self.assertEqual(peer.admissions, [])

    async def test_frozen_outline_not_replaced_from_working_copy(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, peer = self.setup_run(root)
            with RunStore(root) as store:
                for _ in range(6):
                    await advance_book(peer, store, plan)
                names = ["chapters/0001.md", "chapters/0002.md"]
                contexts = ["working-outline.md", "voice.md"]
                frozen = freeze_round(store, 1, names, 65536, contexts)
                before = (root / frozen["contexts"][0]).read_bytes()
                (root / "working-outline.md").write_text("Later living plan")
                self.assertEqual(freeze_round(store, 1, names, 65536, contexts), frozen)
                self.assertEqual((root / frozen["contexts"][0]).read_bytes(), before)
                (root / frozen["contexts"][0]).write_text("Tampered snapshot")
                with self.assertRaisesRegex(StateError, "Frozen review input changed"):
                    freeze_round(store, 1, names, 65536, contexts)

    async def test_final_editor_reads_repairs_from_previous_attempt(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            plan, base = fixtures.BookDriverTests().setup_run(root, 3)
            peer = CreativePeer(root, base.homes)
            peer.target = 3
            peer.mode = "review_again"
            with RunStore(root) as store:
                for _ in range(10):
                    await advance_book(peer, store, plan)
            final_editor = peer.admissions[-1]
            self.assertEqual((final_editor["chapter"], final_editor["turn"]), (3, 6))
            self.assertIn(str(root / "chapters/0001.md"), final_editor["task"])
            self.assertIn(str(root / "chapters/0001-carryover.md"), final_editor["task"])
