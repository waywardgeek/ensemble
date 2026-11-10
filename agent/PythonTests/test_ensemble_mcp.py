"""Offline helper tests: real Python stdio peer, no network or model requests.

Run from repository root: python3 -m unittest discover -s agent/PythonTests -v
These implementation tests are not the chapter's independent exercise grader.
"""
import asyncio
import importlib.util
import json
from pathlib import Path
import shutil
import sys
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parents[1] / "skills/shared/ensemble_mcp.py"
spec = importlib.util.spec_from_file_location("ensemble_mcp", SOURCE)
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


class Contracts(unittest.TestCase):
    def test_structure(self):
        expected = {"type": "object", "properties": {"x": {"type": "integer"}}}
        reordered = {"properties": {"x": {"type": "integer"}}, "type": "object"}
        helper.check_contracts({"work": expected}, [{"name": "work", "inputSchema": reordered}])
        for changed in ({**expected, "required": []}, {"type": "string"}):
            with self.assertRaisesRegex(helper.ContractError, "work: "):
                helper.check_contracts({"work": expected}, [{"name": "work", "inputSchema": changed}])
        self.assertIsNotNone(helper.schema_difference([1, 2], [2, 1]))
        self.assertIsNotNone(helper.schema_difference(True, 1))
        self.assertIn('["x"]', helper.schema_difference({"x": 1}, {"x": 2}))

    def test_missing_and_duplicate(self):
        with self.assertRaisesRegex(helper.ContractError, "Missing tool"):
            helper.check_contracts({"work": {}}, [])
        with self.assertRaisesRegex(helper.ContractError, "Duplicate tool"):
            helper.check_contracts({}, [{"name": "work", "inputSchema": {}}] * 2)


class Duplex(unittest.IsolatedAsyncioTestCase):
    async def asyncSetUp(self):
        self.directory = tempfile.TemporaryDirectory(prefix="ensemble-python-mcp-")
        root = Path(self.directory.name)
        shutil.copyfile(SOURCE, root / "ensemble_mcp.py")
        (root / "peer.py").write_text('''from ensemble_mcp import Server, check_contracts
server = Server("test")
@server.tool("workflow", "Duplex fixture", {"type": "object"})
async def workflow(connection, arguments):
    tools = await connection.list_tools()
    check_contracts({"echo": {"type": "object"}}, tools)
    return await connection.call_tool("echo", arguments)
server.run()
''')
        self.process = await asyncio.create_subprocess_exec(
            sys.executable, "-u", "peer.py", cwd=root,
            stdin=asyncio.subprocess.PIPE, stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE)
        await self.send({"id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05"}})
        self.assertEqual((await self.receive())["id"], 1)
        await self.send({"method": "notifications/initialized"})

    async def asyncTearDown(self):
        if self.process.returncode is None:
            self.process.stdin.close()
            try:
                await asyncio.wait_for(self.process.wait(), 5)
            except asyncio.TimeoutError:
                self.process.kill()
                await self.process.wait()
                self.fail("Python server failed to exit after EOF")
        stderr = (await self.process.stderr.read()).decode()
        self.directory.cleanup()
        self.assertEqual(self.process.returncode, 0, stderr)
        self.assertEqual(stderr, "", stderr)

    async def send(self, message):
        self.process.stdin.write((json.dumps({"jsonrpc": "2.0", **message}) + "\n").encode())
        await self.process.stdin.drain()

    async def receive(self):
        line = await asyncio.wait_for(self.process.stdout.readline(), 5)
        self.assertTrue(line, "Unexpected protocol EOF")
        return json.loads(line)

    async def start_workflow(self):
        await self.send({"id": 2, "method": "tools/call", "params": {
            "name": "workflow", "arguments": {"text": "hello"}}})
        listing = await self.receive()
        self.assertEqual(listing["method"], "tools/list")
        return listing

    async def test_callback_and_tool_error_preservation(self):
        listing = await self.start_workflow()
        await self.send({"id": listing["id"], "result": {
            "tools": [{"name": "echo", "inputSchema": {"type": "object"}}]}})
        call = await self.receive()
        self.assertEqual(call["method"], "tools/call")
        self.assertEqual(call["params"], {"name": "echo", "arguments": {"text": "hello"}})
        result = {"isError": True, "content": [{"type": "text", "text": "denied"}]}
        await self.send({"id": call["id"], "result": result})
        self.assertEqual(await self.receive(), {"jsonrpc": "2.0", "id": 2, "result": result})

    async def test_rpc_refusal(self):
        listing = await self.start_workflow()
        await self.send({"id": listing["id"], "error": {"code": -32602, "message": "private diagnostic"}})
        result = await self.receive()
        self.assertEqual(result["id"], 2)
        self.assertEqual(result["error"]["message"], "Skill request failed")

    async def test_disconnect_during_callback(self):
        await self.start_workflow()
        # TearDown closes the pipe while the handler is waiting for its reply.
        # Exit must be joined, with no retry and no leaked-future diagnostics.

    async def test_contract_failure_prevents_callback_effect(self):
        listing = await self.start_workflow()
        await self.send({"id": listing["id"], "result": {
            "tools": [{"name": "echo", "inputSchema": {"type": "string"}}]}})
        result = await self.receive()
        # The next message must be the original call's refusal, not tools/call.
        self.assertEqual(result["id"], 2)
        self.assertTrue(result["result"]["isError"])
        self.assertIn('echo: $["type"]', result["result"]["content"][0]["text"])
