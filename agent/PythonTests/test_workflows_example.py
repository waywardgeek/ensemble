"""Exercise the shipped teaching example through real stdio, without a provider."""
import asyncio
import json
from pathlib import Path
import shutil
import sys
import tempfile
import unittest

SKILLS = Path(__file__).resolve().parents[1] / "skills"


class WorkflowExample(unittest.IsolatedAsyncioTestCase):
    async def asyncSetUp(self):
        self.directory = tempfile.TemporaryDirectory(prefix="workflow-example-")
        self.root = Path(self.directory.name)
        for name in ("workflows", "shared"):
            shutil.copytree(SKILLS / name, self.root / "skills" / name,
                            ignore=shutil.ignore_patterns("__pycache__"))
        (self.root / "brief.md").write_text("A quiet journey.\n")
        self.contracts = json.loads((self.root / "skills/workflows/contracts.json").read_text())
        self.calls = []
        self.process = await asyncio.create_subprocess_exec(
            sys.executable, "-u", "skills/workflows/example.py", cwd=self.root,
            stdin=asyncio.subprocess.PIPE, stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE)
        await self.send({"id": 1, "method": "initialize", "params": {
            "protocolVersion": "2024-11-05"}})
        self.assertEqual((await self.receive())["id"], 1)
        await self.send({"method": "notifications/initialized"})
        await self.send({"id": 2, "method": "tools/list"})
        listing = await self.receive()
        self.assertEqual([tool["name"] for tool in listing["result"]["tools"]], ["inspect_brief"])

    async def asyncTearDown(self):
        try:
            if self.process.returncode is None:
                self.process.stdin.close()
                try:
                    await asyncio.wait_for(self.process.wait(), 5)
                except asyncio.TimeoutError:
                    self.process.kill()
                    await self.process.wait()
                    self.fail("Example failed to join after EOF")
            stderr = (await self.process.stderr.read()).decode()
            self.assertEqual(self.process.returncode, 0, stderr)
            self.assertEqual(stderr, "")
        finally:
            self.directory.cleanup()

    async def send(self, message):
        self.process.stdin.write((json.dumps({"jsonrpc": "2.0", **message}) + "\n").encode())
        await self.process.stdin.drain()

    async def receive(self):
        line = await asyncio.wait_for(self.process.stdout.readline(), 5)
        self.assertTrue(line, "Unexpected protocol EOF")
        return json.loads(line)

    async def start(self, drift=False):
        await self.send({"id": 3, "method": "tools/call", "params": {
            "name": "inspect_brief", "arguments": {}}})
        request = await self.receive()
        self.assertEqual(request["method"], "tools/list")
        schema = self.contracts["read_file"]
        if drift:
            schema = {**schema, "required": []}
        await self.send({"id": request["id"], "result": {"tools": [
            {"name": "read_file", "inputSchema": schema}]}})
        result = await self.receive()
        if not drift:
            self.assertEqual(result["method"], "tools/call")
            self.assertEqual(result["params"], {
                "name": "read_file", "arguments": {"path": "brief.md"}})
            self.calls.append(result)
        return result

    async def test_normal_read(self):
        call = await self.start()
        # A working boundary fake reads the disposable input; no scripted
        # internal expectations replace the shipped example's implementation.
        text = (self.root / call["params"]["arguments"]["path"]).read_text()
        result = {"content": [{"type": "text", "text": "1: " + text}]}
        await self.send({"id": call["id"], "result": result})
        self.assertEqual(await self.receive(), {"jsonrpc": "2.0", "id": 3, "result": result})
        self.assertEqual(len(self.calls), 1)

    async def test_refused_call_is_not_success(self):
        call = await self.start()
        result = {"isError": True, "content": [{"type": "text", "text": "Read permission revoked"}]}
        await self.send({"id": call["id"], "result": result})
        self.assertEqual(await self.receive(), {"jsonrpc": "2.0", "id": 3, "result": result})

    async def test_contract_drift_prevents_read(self):
        result = await self.start(drift=True)
        self.assertEqual(result["id"], 3)
        self.assertTrue(result["result"]["isError"])
        self.assertIn('read_file: $["required"]', result["result"]["content"][0]["text"])
        self.assertEqual(self.calls, [])

    async def test_lost_completed_read_reply_does_not_retry(self):
        call = await self.start()
        # The peer has completed the read, but the server never receives the
        # reply. Closing the input is uncertainty, not rollback or permission
        # to resend. Drain stdout after joined exit to detect any extra call.
        retained = (self.root / call["params"]["arguments"]["path"]).read_text()
        self.process.stdin.close()
        await asyncio.wait_for(self.process.wait(), 5)
        remaining = (await self.process.stdout.read()).decode()
        for line in remaining.splitlines():
            self.assertNotEqual(json.loads(line).get("method"), "tools/call")
        self.assertEqual(retained, "A quiet journey.\n")
        self.assertEqual(len(self.calls), 1)
