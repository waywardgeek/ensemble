"""Checker controls independent of every student implementation."""

import pathlib
import copy
import json
import tempfile
import types
import unittest

import accept_chat as check


class ChatControls(unittest.TestCase):
    def test_tool_projection_has_positive_and_redaction_negative_controls(self):
        names = sorted(("read_file", "list_directory", "search_files", "write_file", "edit_file", "run_command"))
        schema = {"type": "object", "additionalProperties": False}
        calls = [{"id": "read-visible", "name": "read_file", "args": {"path": "notes.txt"}}]
        fixtures = {
            "anthropic": {"tools": [{"name": n, "input_schema": schema} for n in names], "messages": [
                {"role": "assistant", "content": [{"type": "tool_use", "id": "read-visible", "name": "read_file", "input": {"path": "notes.txt"}}]},
                {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "read-visible", "content": check.PRIVATE_RESULT}]}]},
            "openai": {"tools": [{"type": "function", "function": {"name": n, "parameters": schema}} for n in names], "messages": [
                {"role": "assistant", "tool_calls": [{"id": "read-visible", "type": "function", "function": {"name": "read_file", "arguments": '{"path":"notes.txt"}'}}]},
                {"role": "tool", "tool_call_id": "read-visible", "content": check.PRIVATE_RESULT}]},
            "gemini": {"tools": [{"functionDeclarations": [{"name": n, "parametersJsonSchema": schema} for n in names]}], "contents": [
                {"role": "model", "parts": [{"functionCall": {"id": "read-visible", "name": "read_file", "args": {"path": "notes.txt"}}, "thoughtSignature": "fixture-signature-read-visible"}]},
                {"role": "user", "parts": [{"functionResponse": {"id": "read-visible", "name": "read_file", "response": {"result": check.PRIVATE_RESULT}}}]}]},
        }
        for vendor, body in fixtures.items():
            with self.subTest(vendor=vendor):
                self.assertEqual(check.tool_projection_errors(body, vendor, calls, False), [])
                self.assertTrue(check.tool_projection_errors(body, vendor, calls, True))
                redacted = json.loads(json.dumps(body).replace(check.PRIVATE_RESULT, "[redacted]"))
                self.assertEqual(check.tool_projection_errors(redacted, vendor, calls, True), [])
                self.assertTrue(check.tool_projection_errors(redacted, vendor, calls, False))
                broken = copy.deepcopy(redacted)
                if vendor == "openai":
                    broken["messages"][1]["tool_call_id"] = "wrong"
                elif vendor == "anthropic":
                    broken["messages"][1]["content"][0]["tool_use_id"] = "wrong"
                else:
                    broken["contents"][0]["parts"][0]["thoughtSignature"] = "wrong"
                self.assertTrue(check.tool_projection_errors(broken, vendor, calls, True))

    def test_tool_fixture_is_independent_and_covers_all_six(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            calls = check.tool_plan(root)
            self.assertEqual(len(set(c["name"] for c in calls)), 6)
            self.assertEqual((root / "notes.txt").read_text(), check.PRIVATE_RESULT + "\n")
            self.assertFalse((root / "absent.txt").exists())
            self.assertFalse((root / "made.txt").exists())
            self.assertEqual(calls[0]["id"], "missing")

    def test_usage_format_variants_and_wrong_counts(self):
        for text in ("Input: 15; cache write: 0; cache read: 0; output: 6",
                     "INPUT TOKENS = 15\nCACHE_WRITE=0\nCACHE-READ: 0\nOUTPUT 6"):
            self.assertEqual(check.usage_errors(text, 3), [])
        for text in ("input: 150 cache_write: 0 cache_read: 0 output: 6",
                     "input: 15 output: 6", "elsewhere 15 0 0 6"):
            self.assertTrue(check.usage_errors(text, 3))

    def test_user_projection_supports_text_arrays(self):
        fixtures = {
            "openai": {"messages": [{"role": "system", "content": "private"},
                        {"role": "user", "content": [{"type": "text", "text": "  café"},
                                                       {"type": "text", "text": "  "}]}]},
            "anthropic": {"messages": [{"role": "user", "content": [{"type": "text", "text": "  café  "}]}]},
            "gemini": {"contents": [{"role": "user", "parts": [{"text": "  café  "}]}]},
        }
        for vendor, body in fixtures.items():
            self.assertEqual(check.user_texts(body, vendor), ["  café  "])

    def test_tool_only_fixtures_really_have_no_text_parts(self):
        for vendor in check.VENDORS:
            body = check.fixture(vendor, "tool-only")[0]
            if vendor == "openai":
                self.assertIsNone(body["choices"][0]["message"]["content"])
            elif vendor == "anthropic":
                self.assertEqual([p["type"] for p in body["content"]], ["tool_use"])
            else:
                self.assertFalse(any("text" in p for p in body["candidates"][0]["content"]["parts"]))

    def test_terminal_driver_uses_a_tty_and_preserves_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            binary = root / "fixture"
            binary.write_text('#!/usr/bin/env python3\nimport os, sys\n'
                              'assert os.isatty(0) and os.isatty(1)\n'
                              'print("READY\\nYou> ", end="", flush=True)\n'
                              'for line in sys.stdin:\n'
                              ' if line == "/quit\\n": break\n'
                              ' print("Assistant: first\\nsecond\\nYou> ", end="", flush=True)\n'
                              'print("FINISHED", flush=True)\n')
            binary.chmod(0o700)
            session = types.SimpleNamespace(binary=binary, root=root, vendor="anthropic",
                                            log=root / "log", server=types.SimpleNamespace(server_port=1))
            terminal = check.Terminal(session, ["chat"])
            try:
                self.assertEqual(terminal.prompt(), "READY\nYou> ")
                out = terminal.prompt(b"NOT-ECHOED\n")
                self.assertEqual(out, "Assistant: first\nsecond\nYou> ")
                self.assertNotIn("NOT-ECHOED", out)
                code, out, err = terminal.finish(b"/quit\n")
                self.assertEqual((code, out, err), (0, "FINISHED\n", ""))
            finally:
                terminal.close()

    def test_terminal_driver_rejects_missing_prompt(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            binary = root / "fixture"
            binary.write_text('#!/usr/bin/env python3\nprint("no prompt")\n')
            binary.chmod(0o700)
            session = types.SimpleNamespace(binary=binary, root=root, vendor="anthropic",
                                            log=root / "log", server=types.SimpleNamespace(server_port=1))
            terminal = check.Terminal(session, [])
            try:
                with self.assertRaises(AssertionError):
                    terminal.prompt()
            finally:
                terminal.close()


if __name__ == "__main__":
    unittest.main()
