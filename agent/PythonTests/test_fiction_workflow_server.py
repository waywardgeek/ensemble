"""Configured entry-point tests using real state and the working book peer."""
import asyncio
import copy
import json
from pathlib import Path
import sys
import tempfile
import unittest

SKILLS = Path(__file__).resolve().parents[1] / "skills"
sys.path.insert(0, str(SKILLS / "shared"))
sys.path.insert(0, str(SKILLS / "fiction-workflow"))
from ensemble_mcp import ContractError
from server import Workflow
from run_state import RunStore, StateError
from test_fiction_book_driver import BookPeer


class CatalogPeer(BookPeer):
    def __init__(self, root, contracts):
        super().__init__(root, root / "children")
        self.catalog = [{"name": name, "inputSchema": copy.deepcopy(schema)}
                        for name, schema in contracts.items()]

    async def list_tools(self):
        return self.catalog


class WorkflowServerTests(unittest.IsolatedAsyncioTestCase):
    def prepare(self, root):
        baseline = root / "baseline"
        for role in ("author", "editor"):
            directory = baseline / role
            directory.mkdir(parents=True)
            (directory / "SOUL.md").write_text("Role " + role)
            (directory / "MEMORY.md").write_bytes(b"")
        (root / "outline.md").write_text("Supplied outline")
        (root / "voice.md").write_text("Supplied voice")
        setup = root / "setup.json"
        setup.write_text(json.dumps({"version": 1, "owner_directory": str(root),
            "run_directory": str(root), "baseline_root": str(baseline),
            "chapter_count": 1, "outline": "outline.md", "voice": "voice.md",
            "source_limit": 65536, "result_limit": 131072}))
        workflow = Workflow(setup)
        return workflow, CatalogPeer(root, workflow.contracts)

    async def test_complete_configured_book_and_reopen_without_dispatch(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            workflow, peer = self.prepare(root)
            result = await workflow.continue_book(peer, {})
            self.assertEqual(json.loads(result["content"][0]["text"]), {"action": "complete"})
            self.assertTrue((root / "final-book/manuscript.md").is_file())
            self.assertFalse(peer.live)
            count = len(peer.admissions)
            await Workflow(root / "setup.json").continue_book(peer, {})
            self.assertEqual(len(peer.admissions), count)

    async def test_contract_drift_and_owner_argument_refuse_before_state(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            workflow, peer = self.prepare(root)
            with self.assertRaises(StateError):
                await workflow.continue_book(peer, {"owner_directory": "another-agent"})
            peer.catalog[0]["inputSchema"]["extra"] = True
            with self.assertRaises(ContractError):
                await workflow.continue_book(peer, {})
            self.assertFalse((root / "state.json").exists())
            self.assertFalse((root / "children").exists())
            self.assertFalse(peer.admissions)

    async def test_actual_stdio_entrypoint_refuses_missing_reverse_contracts(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.prepare(root)
            process = await asyncio.create_subprocess_exec(
                sys.executable, "-u", str(SKILLS / "fiction-workflow/server.py"),
                str(root / "setup.json"), stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)

            async def send(value):
                process.stdin.write((json.dumps(value) + "\n").encode())
                await process.stdin.drain()

            async def receive():
                line = await asyncio.wait_for(process.stdout.readline(), 5)
                self.assertTrue(line, "Server closed before protocol response")
                return json.loads(line)

            try:
                await send({"jsonrpc": "2.0", "id": 1, "method": "initialize",
                            "params": {"protocolVersion": "2024-11-05", "capabilities": {},
                                       "clientInfo": {"name": "fixture", "version": "1"}}})
                self.assertEqual((await receive())["id"], 1)
                await send({"jsonrpc": "2.0", "method": "notifications/initialized"})
                await send({"jsonrpc": "2.0", "id": 2, "method": "tools/list", "params": {}})
                tools = (await receive())["result"]["tools"]
                self.assertEqual([tool["name"] for tool in tools], ["continue_fiction_book"])
                await send({"jsonrpc": "2.0", "id": 3, "method": "tools/call",
                            "params": {"name": "continue_fiction_book", "arguments": {}}})
                reverse = await receive()
                self.assertEqual(reverse["method"], "tools/list")
                await send({"jsonrpc": "2.0", "id": reverse["id"], "result": {"tools": []}})
                response = await receive()
                self.assertEqual(response["id"], 3)
                self.assertTrue(response["result"]["isError"])
                self.assertIn("Missing tool", response["result"]["content"][0]["text"])
                self.assertFalse((root / "state.json").exists())
            finally:
                process.stdin.close()
                try:
                    await asyncio.wait_for(process.wait(), 5)
                except asyncio.TimeoutError:
                    process.kill()
                    await process.wait()
                    self.fail("Server did not join after EOF")
            self.assertEqual(process.returncode, 0, (await process.stderr.read()).decode())

    async def public_calls(self, root, peer, count=1):
        """Drive the real server and answer reverse requests with the working peer."""
        process = await asyncio.create_subprocess_exec(
            sys.executable, "-u", str(SKILLS / "fiction-workflow/server.py"),
            str(root / "setup.json"), stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)

        async def send(value):
            process.stdin.write((json.dumps({"jsonrpc": "2.0", **value}) + "\n").encode())
            await process.stdin.drain()

        async def receive():
            line = await asyncio.wait_for(process.stdout.readline(), 5)
            self.assertTrue(line, "Unexpected protocol EOF")
            return json.loads(line)

        responses = []
        try:
            await send({"id": 1, "method": "initialize", "params": {
                "protocolVersion": "2024-11-05"}})
            self.assertEqual((await receive())["id"], 1)
            await send({"method": "notifications/initialized"})
            for index in range(count):
                identity = 10 + index
                await send({"id": identity, "method": "tools/call", "params": {
                    "name": "continue_fiction_book", "arguments": {}}})
                while True:
                    message = await receive()
                    if message["id"] == identity:
                        responses.append(message)
                        break
                    if message["method"] == "tools/list":
                        result = {"tools": await peer.list_tools()}
                    else:
                        self.assertEqual(message["method"], "tools/call")
                        call = message["params"]
                        result = await peer.call_tool(call["name"], call["arguments"])
                    await send({"id": message["id"], "result": result})
        finally:
            process.stdin.close()
            try:
                await asyncio.wait_for(process.wait(), 5)
            except asyncio.TimeoutError:
                process.kill()
                await process.wait()
                self.fail("Public workflow server did not join")
        self.assertEqual(process.returncode, 0, (await process.stderr.read()).decode())
        return responses

    async def test_public_read_constraints_preserve_pending_and_do_not_admit(self):
        for field, limit, diagnostic in (("source_limit", 2, "source byte limit"),
                                         ("result_limit", 2, "result byte limit")):
            for pending in (False, True):
                with self.subTest(field=field, pending=pending):
                    with tempfile.TemporaryDirectory() as directory:
                        root = Path(directory)
                        workflow, peer = self.prepare(root)
                        with RunStore(root) as store:
                            if pending:
                                store.prepare(child_id="existing", phase="write",
                                    round_number=0, chapter=1, turn=1, task="Retained task")
                            before = store.snapshot()
                        setup = json.loads((root / "setup.json").read_text())
                        setup[field] = limit
                        (root / "setup.json").write_text(json.dumps(setup))
                        response = (await self.public_calls(root, peer))[0]
                        self.assertIn("result", response, response)
                        self.assertTrue(response["result"]["isError"])
                        self.assertIn(diagnostic, response["result"]["content"][0]["text"])
                        self.assertEqual(json.loads((root / "state.json").read_text()), before)
                        self.assertFalse(peer.admissions)

    async def test_public_invalid_artifact_refuses_recovered_result_without_retry(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            workflow, peer = self.prepare(root)
            original = peer.call_tool

            async def corrupt_author(name, arguments):
                result = await original(name, arguments)
                if name == "spawn_sub_agent":
                    (root / "chapters/0001.md").write_bytes(b"\xff")
                return result

            peer.call_tool = corrupt_author
            responses = await self.public_calls(root, peer, count=2)
            for response in responses:
                self.assertIn("result", response, response)
                self.assertTrue(response["result"]["isError"])
                self.assertIn("UTF-8", response["result"]["content"][0]["text"])
            state = json.loads((root / "state.json").read_text())
            self.assertEqual(state["accepted"], [])
            self.assertEqual(state["pending"]["id"], peer.admissions[0]["id"])
            self.assertEqual(len(peer.admissions), 1)

    async def test_uncertain_result_returns_for_supervision_without_retry(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            workflow, peer = self.prepare(root)
            peer.lose_reply = True
            result = await workflow.continue_book(peer, {})
            self.assertEqual(json.loads(result["content"][0]["text"])["action"], "supervise")
            self.assertEqual(len(peer.admissions), 1)
            self.assertIsNotNone(json.loads((root / "state.json").read_text())["pending"])


if __name__ == "__main__":
    unittest.main()
