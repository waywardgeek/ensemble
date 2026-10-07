"""Independent checker controls; no student source or serializer imports."""

import copy
import json
import pathlib
import tempfile
import unittest

import accept_ch03 as check


class ProjectionControls(unittest.TestCase):
    def test_literal_valid_surfaces(self):
        # These are written independently of response(), which generates the
        # opposite direction (provider responses). Pretty/escaped JSON is legal.
        fixtures = {
            "anthropic": {"messages": [
                {"role": "assistant", "content": [{"type": "tool_use", "id": "a", "name": "read_file", "input": {"path": "x"}}]},
                {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "a", "content": [{"type": "text", "text": "caf\u00e9"}], "is_error": True}]}]},
            "openai": {"messages": [
                {"role": "assistant", "tool_calls": [{"id": "a", "type": "function", "function": {"name": "read_file", "arguments": '{ "path" : "x" }'}}]},
                {"role": "tool", "tool_call_id": "a", "content": [{"type": "text", "text": "caf"}, {"type": "text", "text": "\u00e9"}]}]},
            "gemini": {"contents": [
                {"role": "model", "parts": [{"functionCall": {"id": "a", "name": "read_file", "args": {"path": "x"}}, "thoughtSignature": "fixture-signature-a"}]},
                {"role": "user", "parts": [{"functionResponse": {"id": "a", "name": "read_file", "response": {"error": "caf\u00e9"}}}]}]},
        }
        for vendor, fixture in fixtures.items():
            with self.subTest(vendor=vendor):
                body = json.loads(json.dumps(fixture, indent=2, ensure_ascii=True))
                calls, results, signatures = check.wire_facts(body, vendor)
                self.assertEqual(calls, [("a", "read_file", {"path": "x"})])
                self.assertEqual(results, [("a", "caf\u00e9", None if vendor == "openai" else True)])
                if vendor == "gemini":
                    self.assertEqual(signatures, {"a": "fixture-signature-a"})

    def test_gemini_rejects_wrong_name_id_and_role(self):
        body = {"contents": [
            {"role": "model", "parts": [{"functionCall": {"id": "a", "name": "read_file", "args": {}}}]},
            {"role": "user", "parts": [{"functionResponse": {"id": "a", "name": "read_file", "response": {"result": "ok"}}}]}]}
        for field in ("id", "name", "role"):
            with self.subTest(field=field):
                bad = copy.deepcopy(body)
                if field == "role":
                    bad["contents"][1]["role"] = "model"
                else:
                    bad["contents"][1]["parts"][0]["functionResponse"][field] = "wrong"
                with self.assertRaises(ValueError):
                    check.wire_facts(bad, "gemini")

    def test_text_arrays_do_not_invent_separators(self):
        self.assertEqual(check.text_content([{"type": "text", "text": "a\n"}, {"type": "text", "text": "b"}]), "a\nb")
        for invalid in (["not-a-part"], [{"type": "image", "text": "marker"}], [{"type": "text", "text": 4}]):
            with self.assertRaises(ValueError):
                check.text_content(invalid)


class OutcomeControls(unittest.TestCase):
    @staticmethod
    def result(text, failed=False):
        return {"parts": [{"type": "text", "text": text}], "is_error": failed}

    def test_read_caps_have_positive_and_negative_controls(self):
        good = {
            "range": self.result("gamma line three\ndelta line four\n"),
            "clamp": self.result("epsilon line five\n"), "empty": self.result(""),
            "empty-zero": self.result(""), "empty-range": self.result("range refused", True),
            "beyond": self.result("range refused", True), "exact": self.result("~" * 32),
            "omit": self.result("~" * 16 + "\n16 bytes omitted due to max_bytes"),
        }
        errors = []
        check.check_reads(None, good, errors)
        self.assertEqual(errors, [])
        for key, replacement in (("exact", "~" * 32 + " truncated"), ("omit", "~" * 32), ("range", "entire unrelated file")):
            bad = copy.deepcopy(good)
            bad[key] = self.result(replacement)
            errors = []
            check.check_reads(None, bad, errors)
            self.assertTrue(errors)

    def test_search_binary_fixture_is_selected_by_glob(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            calls = check.search_case(root)
            self.assertEqual(next(c for c in calls if c["id"] == "tree")["args"]["file_pattern"], "*.txt")
            self.assertEqual((root / "tree/b.txt").read_bytes()[:1], b"\x00")
            self.assertTrue((root / "tree/link").is_symlink())
            self.assertTrue((root / "tree/.git/hidden.txt").exists())

    def test_unicode_control_detects_replacement_and_false_truncation(self):
        good = {f"{kind}-{n}": self.result(("", "é", "éX")[n - 1] + (" [omitted]" if n < 3 else ""))
                for kind in ("read", "command") for n in (1, 2, 3)}
        errors = []
        check.check_unicode(None, good, errors)
        self.assertEqual(errors, [])
        for key, wrong in (("read-1", "\ufffd [truncated]"), ("command-3", "éX [truncated]"),
                           ("command-2", "éX [truncated]")):
            bad = copy.deepcopy(good)
            bad[key] = self.result(wrong)
            errors = []
            check.check_unicode(None, bad, errors)
            self.assertTrue(errors)

    def test_ordinary_errors_cannot_skip_final_call(self):
        results = {identifier: self.result("read_file invalid " + field, True)
                   for identifier, field in (("unknown", "unknown"), ("type", "path"),
                                             ("unknown-field", "invented"), ("zero", "max_bytes"),
                                             ("fraction", "start_line"), ("negative", "end_line"))}
        results["after"] = self.result("AFTER-ERROR-READ")
        with tempfile.TemporaryDirectory() as tmp:
            errors = []
            check.check_invalid(pathlib.Path(tmp), results, errors)
            self.assertEqual(errors, [])
            del results["after"]
            errors = []
            check.check_invalid(pathlib.Path(tmp), results, errors)
            self.assertIn("missing result after", errors)

    def test_disk_observation_catches_lying_success(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            check.edit_case(root)
            results = {name: self.result("ok") for name in ("unique", "delete", "replace", "append", "new", "empty", "append-new")}
            results.update(ambiguous=self.result("2 matches", True), missing=self.result("0 matches", True),
                           **{"empty-anchor": self.result("empty anchor", True), "guard": self.result("guard 13 bytes", True),
                              "noop": self.result("unchanged")})
            errors = []
            check.check_edits(root, results, errors)
            self.assertIn("incorrect actual disk bytes for unique", errors)
            self.assertIn("incorrect actual disk bytes for parents/new", errors)


if __name__ == "__main__":
    unittest.main()
