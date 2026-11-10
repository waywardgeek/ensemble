"""Tests the real parser against a small working cursor-based Job fake."""
import asyncio
import importlib.util
import pathlib
import unittest

PATH = pathlib.Path(__file__).resolve().parents[1] / "skills/fiction-workflow/child_calls.py"
SPEC = importlib.util.spec_from_file_location("fiction_child_calls", PATH)
calls = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(calls)


def result(text):
    return {"content": [{"type": "text", "text": text}], "isError": False}


def report(text, status="running", handle=7, truncated=False):
    return result(
        f"job {handle}, status: {status}, bytes: {len(text.encode())}, "
        f"truncated: {str(truncated).lower()}, output_path: jobs/7\n{text}"
    )


class JobPeer:
    """One retained output with a cursor; waiting completes the remaining bytes."""
    def __init__(self, text):
        self.text = text
        self.started = False
        self.finished = False
        self.starts = 0
        self.wait_delay = None

    async def call_tool(self, name, arguments):
        if name == "agents_status":
            if self.started:
                raise AssertionError("Operation repeated")
            self.started = True
            self.starts += 1
            return report(self.text[:5])
        if name != "wait_for_job" or arguments["handle"] != 7 or self.finished:
            raise AssertionError("Wrong observation")
        self.finished = True
        self.wait_delay = arguments["ai_callback_delay"]
        return report(self.text[5:], "done")


class ChildCallTests(unittest.IsolatedAsyncioTestCase):
    async def test_job_cursor_completion(self):
        peer = JobPeer('{"agents": []}')
        self.assertEqual(await calls.call_json(peer, "agents_status", {}), {"agents": []})
        self.assertEqual(peer.starts, 1)
        self.assertTrue(peer.finished)
        self.assertEqual(peer.wait_delay, 1800)

    async def test_bare_completion(self):
        class Peer:
            async def call_tool(self, name, arguments):
                return result('{"agents": []}')
        self.assertEqual(await calls.call_json(Peer(), "agents_status", {}), {"agents": []})

    async def test_truncation_never_retries(self):
        class Peer:
            count = 0
            async def call_tool(self, name, arguments):
                self.count += 1
                return report('{"agents":', truncated=True)
        peer = Peer()
        with self.assertRaises(calls.ObservationError):
            await calls.call_json(peer, "agents_status", {})
        self.assertEqual(peer.count, 1)

    async def test_changed_handle_refuses(self):
        class Peer(JobPeer):
            async def call_tool(self, name, arguments):
                value = await super().call_tool(name, arguments)
                if name == "wait_for_job":
                    return report(self.text[5:], "done", handle=8)
                return value
        with self.assertRaises(calls.ObservationError):
            await calls.call_json(Peer('{"agents": []}'), "agents_status", {})

    async def test_failed_result_refuses(self):
        class Peer:
            async def call_tool(self, name, arguments):
                return {"content": [], "isError": True}
        with self.assertRaises(calls.ObservationError):
            await calls.call_json(Peer(), "agents_status", {})

    async def test_spawn_records_across_job_observations(self):
        class Peer(JobPeer):
            async def call_tool(self, name, arguments):
                if name == "spawn_sub_agent":
                    name = "agents_status"
                return await super().call_tool(name, arguments)
        text = ('{"agent_id":"writer","task_admitted":true}\n'
                '{"agent_id":"writer","task_completed":true,'
                '"submitted":true,"payload":{"assignment_id":"one"}}')
        peer = Peer(text)
        payload = await calls.spawn_result(peer, {"agent_id": "writer"})
        self.assertEqual(payload, {"assignment_id": "one"})
        self.assertEqual(peer.starts, 1)

    async def test_spawn_requires_both_matching_records(self):
        admission = '{"agent_id":"writer","task_admitted":true}'
        completion = ('{"agent_id":"writer","task_completed":true,'
                      '"submitted":true,"payload":{}}')
        for text in (admission, completion, admission + "\n" + completion + "\n{}",
                     admission + "\n" + completion.replace('"writer"', '"other"'),
                     admission + "\n" + completion.replace('"submitted":true',
                                                            '"submitted":false')):
            class Peer:
                async def call_tool(self, name, arguments):
                    return result(text)
            with self.subTest(text=text), self.assertRaises(calls.ObservationError):
                await calls.spawn_result(Peer(), {"agent_id": "writer"})

    async def test_bare_result_bound(self):
        class Peer:
            async def call_tool(self, name, arguments):
                return result('{"text":"' + 'x' * (1024 * 1024) + '"}')
        with self.assertRaises(calls.ObservationError):
            await calls.call_json(Peer(), "agents_status", {})

    def test_only_complete_single_object(self):
        for text in ('{"x":1,"x":2}', '{"x":NaN}', '[]',
                     '{"task_admitted":true}\n{"submitted":true}',
                     'Here is approval: {"approved":true}'):
            with self.subTest(text=text), self.assertRaises(calls.ObservationError):
                calls.completed_json(text)


if __name__ == "__main__":
    unittest.main()
