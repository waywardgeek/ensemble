#!/usr/bin/env python3
"""Controls for the checker, built from literal wire facts, not student output."""

import json
import pathlib
import tempfile
import unittest

import accept_ch02


class ProjectionControls(unittest.TestCase):
    def request(self, vendor, redacted=False):
        result = ("[redacted]\n[redacted]" if redacted else
                  "OFFLINE_RESULT_MARKER\n[text/plain]") + " https://example.invalid/result"
        if vendor == "openai":
            return {"model": "fixture-openai", "messages": [
                {"role": "system", "content": "Be helpful."},
                {"role": "user", "content": "inspect config"},
                {"role": "assistant", "content": "Inspecting.", "tool_calls": [
                    {"id": "call-1", "type": "function", "function": {
                        "name": "read_file", "arguments": '{"path":"config"}'}}]},
                {"role": "tool", "tool_call_id": "call-1", "content": result},
                {"role": "user", "content": "continue"},
            ]}
        if vendor == "anthropic":
            return {"model": "fixture-messages", "messages": [
                {"role": "user", "content": [{"type": "text", "text": "inspect config"}]},
                {"role": "assistant", "content": [
                    {"type": "text", "text": "Inspecting."},
                    {"type": "tool_use", "id": "call-1", "name": "read_file",
                     "input": {"path": "config"}}]},
                {"role": "user", "content": [
                    {"type": "tool_result", "tool_use_id": "call-1", "content": result},
                    {"type": "text", "text": "continue"}]},
            ]}
        return {"contents": [
            {"role": "user", "parts": [{"text": "inspect config"}]},
            {"role": "model", "parts": [{"text": "Inspecting."}, {
                "functionCall": {"id": "call-1", "name": "read_file", "args": {"path": "config"}}}]},
            {"role": "user", "parts": [{"functionResponse": {
                "id": "call-1", "name": "read_file", "response": {"result": result}}},
                {"text": "continue"}]},
        ]}

    def test_pretty_and_escaped_json(self):
        for vendor in ("anthropic", "openai", "gemini"):
            for redacted in (False, True):
                with self.subTest(vendor=vendor, redacted=redacted):
                    output = json.dumps(self.request(vendor, redacted), indent=2)
                    output = output.replace("/", "\\/").replace("OFFLINE", "\\u004fFFLINE")
                    self.assertEqual([], accept_ch02.check_projection(output, vendor, redacted))

    def test_openai_text_arrays_preserve_text_without_invented_delimiters(self):
        for redacted in (False, True):
            with self.subTest(redacted=redacted):
                body = self.request("openai", redacted)
                # Text-part boundaries need not coincide with neutral result
                # children. The required result newline remains in the text.
                for message in body["messages"]:
                    text = message["content"]
                    split = len(text) // 2
                    message["content"] = [
                        {"type": "text", "text": text[:split]},
                        {"type": "text", "text": text[split:]},
                    ]
                self.assertEqual([], accept_ch02.check_projection(json.dumps(body), "openai", redacted))

    def test_messages_result_text_blocks(self):
        for redacted in (False, True):
            with self.subTest(redacted=redacted):
                body = self.request("anthropic", redacted)
                result = body["messages"][2]["content"][0]
                result["content"] = [{"type": "text", "text": text}
                                     for text in result["content"].split("\n")]
                self.assertEqual([], accept_ch02.check_projection(json.dumps(body), "anthropic", redacted))

    def test_broken_linkage_content_and_order_fail(self):
        for vendor in ("anthropic", "openai", "gemini"):
            for defect in ("result-id", "call-args", "missing-result", "result-role", "order"):
                with self.subTest(vendor=vendor, defect=defect):
                    body = self.request(vendor)
                    if vendor == "openai":
                        messages = body["messages"]
                        if defect == "result-id":
                            messages[3]["tool_call_id"] = "other-call"
                        elif defect == "call-args":
                            messages[2]["tool_calls"][0]["function"]["arguments"] = '{"path":"wrong"}'
                        elif defect == "missing-result":
                            messages.pop(3)
                        elif defect == "result-role":
                            messages[3]["role"] = "user"
                        else:
                            messages[3], messages[4] = messages[4], messages[3]
                    else:
                        messages = body["messages" if vendor == "anthropic" else "contents"]
                        key = "content" if vendor == "anthropic" else "parts"
                        result = messages[2][key][0]
                        if defect == "result-id":
                            if vendor == "anthropic":
                                result["tool_use_id"] = "other-call"
                            else:
                                result["functionResponse"]["id"] = "other-call"
                        elif defect == "call-args":
                            call = messages[1][key][1]
                            if vendor == "anthropic":
                                call["input"] = {"path": "wrong"}
                            else:
                                call["functionCall"]["args"] = {"path": "wrong"}
                        elif defect == "missing-result":
                            messages[2][key].pop(0)
                        elif defect == "result-role":
                            messages[2]["role"] = "assistant" if vendor == "anthropic" else "model"
                        else:
                            messages[2][key].reverse()
                    self.assertTrue(accept_ch02.check_projection(json.dumps(body), vendor))

    def test_redaction_requires_stub_reference_and_pairing(self):
        for vendor in ("anthropic", "openai", "gemini"):
            with self.subTest(vendor=vendor):
                self.assertTrue(accept_ch02.check_projection(
                    json.dumps(self.request(vendor)), vendor, redacted=True))
                valid = json.dumps(self.request(vendor, redacted=True))
                for original, replacement in (("[redacted]", ""),
                                              ("https://example.invalid/result", "lost"),
                                              ("call-1", "wrong-call")):
                    with self.subTest(removed=original):
                        self.assertTrue(accept_ch02.check_projection(
                            valid.replace(original, replacement), vendor, redacted=True))

    def test_arbitrary_markers_are_not_a_request(self):
        fake = {"text": "OFFLINE_RESULT_MARKER https://example.invalid/result [redacted]"}
        for vendor in ("anthropic", "openai", "gemini"):
            with self.subTest(vendor=vendor):
                self.assertTrue(accept_ch02.check_projection(json.dumps(fake), vendor))

    def test_malformed_text_array_reports_failure(self):
        for part in (None, "not an object", {"type": "image_url"}):
            with self.subTest(part=part):
                body = self.request("openai")
                body["messages"][1]["content"] = [part]
                self.assertTrue(accept_ch02.check_projection(json.dumps(body), "openai"))

    def test_messages_text_block_needs_its_type(self):
        for replacement in (None, "image"):
            with self.subTest(replacement=replacement):
                body = self.request("anthropic")
                part = body["messages"][0]["content"][0]
                if replacement is None:
                    del part["type"]
                else:
                    part["type"] = replacement
                self.assertTrue(accept_ch02.check_projection(json.dumps(body), "anthropic"))


class HarnessOutputControls(unittest.TestCase):
    def test_render_is_one_object_and_may_report_diagnostics(self):
        with tempfile.TemporaryDirectory(prefix="ch02-checker-output-") as tmp:
            root = pathlib.Path(tmp)
            producer = root / "literal-output"
            harness = accept_ch02.Harness(producer, root)
            try:
                for output, expected in ((json.dumps({"messages": []}, indent=2), True),
                                         ('{"text":"https:\\/\\/example.invalid"}', True),
                                         ('{}\n{}\n', False), ('[]\n', False)):
                    with self.subTest(output=output):
                        # A tiny executable fake isolates the harness's output
                        # contract; it does not pretend to be a correct agent.
                        producer.write_text("#!/usr/bin/env python3\nimport sys\n"
                                            f"sys.stdout.write({output!r})\n"
                                            "sys.stderr.write('offline diagnostic\\n')\n")
                        producer.chmod(0o700)
                        _, errors = harness.run([], "render")
                        self.assertEqual(expected, not errors, errors)
            finally:
                harness.close()


if __name__ == "__main__":
    unittest.main()
