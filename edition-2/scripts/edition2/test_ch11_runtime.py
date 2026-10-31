#!/usr/bin/env python3
"""Checker controls only; canned captures are never student runtime positives."""
import copy
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from accept_ch11 import compact, request
from accept_ch11_runtime import EXPECTED, PEER, TEXT, evaluate


def positive():
    return {"case": "valid", "exit": 0, "timeout": False, "peer_cleanup": [], "cwd": "/fixture",
            "stdout": "", "stderr": "", "notebook_hex": TEXT.encode().hex(),
            "peer": [{"startup": True, "cwd": "/fixture/config", "allowed": "allowlisted-value", "credential_present": False, "unlisted_present": False}] +
                    [{"response": "CH11_IGNORED_METADATA"}] + [{"request": json.dumps(r)} for r in [request(1, "server/discover"), request(2, "tools/list"), request(3, "tools/call", name="notes.append", arguments={"text": TEXT})]],
            "events": [{"type": "tool_returned", "tool": {"call_id": "remote-1", "job": {"status": "done", "output": {"locator": "cr/io/1"}}}}],
            "artifacts_hex": {"cr/io/1": EXPECTED.hex()},
            "requests": [{"tools": [{"name": "append_note"}]}, {"messages": [{"content": [{"type": "tool_result", "tool_use_id": "remote-1", "content": EXPECTED.decode()}]}]}]}


class Controls(unittest.TestCase):
    def test_canned_positive_and_predicate_negatives(self):
        parent = positive()
        self.assertTrue(all(evaluate(parent).values()))
        mutations = [
            ("actual-notebook-effect-once", lambda x: x.update(notebook_hex=None)),
            ("literal-discover-list-call", lambda x: x["peer"].pop()),
            ("canonical-artifact-and-job", lambda x: x["artifacts_hex"].clear()),
            ("declaration-and-recorded-continuation", lambda x: x["requests"].pop()),
            ("relative-path-unused-entry-allowlist", lambda x: x["peer"][0].update(credential_present=True)),
            ("metadata-excluded", lambda x: x.update(stdout="CH11_IGNORED_METADATA")),
        ]
        for name, mutate in mutations:
            with self.subTest(name=name):
                changed = copy.deepcopy(parent)
                mutate(changed)
                self.assertFalse(evaluate(changed)[name])

    def test_no_send_requires_controlled_error(self):
        parent = positive()
        parent.update(case="invalid-arguments", notebook_hex=None, artifacts_hex={})
        parent["peer"].pop()
        parent["events"][0]["tool"].update(is_error=True)
        self.assertTrue(all(evaluate(parent).values()))
        parent["events"].clear()
        self.assertFalse(all(evaluate(parent).values()))

    def test_whole_refusal_and_prelaunch_oracles(self):
        parent = positive()
        parent["case"] = "mixed-result"
        parent["events"][0]["tool"]["is_error"] = True
        error = b'{"error":"mcp_unsupported_content","message":"Remote result contains unsupported content."}\n'
        parent["artifacts_hex"] = {"cr/io/1": error.hex()}
        self.assertTrue(all(evaluate(parent).values()))
        parent["artifacts_hex"]["cr/io/1"] = (TEXT.encode() + error).hex()
        self.assertFalse(evaluate(parent)["mixed-result-whole-refusal"])
        parent = positive()
        parent.update(case="missing-reference", exit=1, peer=[], requests=[], artifacts_hex={})
        self.assertTrue(all(evaluate(parent).values()))
        parent["peer"].append({"startup": True})
        self.assertFalse(all(evaluate(parent).values()))

    def test_actual_peer_effect_and_mixed_response(self):
        for mixed in (False, True):
            with self.subTest(mixed=mixed), tempfile.TemporaryDirectory(prefix="ch11-peer-control-") as directory:
                path = Path(directory)
                messages = [request(1, "server/discover"), request(2, "tools/list"), request(3, "tools/call", name="notes.append", arguments={"text": TEXT})]
                proc = subprocess.run([sys.executable, str(PEER), directory] + (["--mixed"] if mixed else []),
                                      input=b"\n".join(compact(m) for m in messages) + b"\n", capture_output=True, timeout=5,
                                      env={"CH11_ALLOWED": "allowlisted-value"}, cwd=path)
                self.assertEqual(proc.returncode, 0, proc.stderr)
                replies = [json.loads(line) for line in proc.stdout.splitlines()]
                self.assertEqual(len(replies), 3)
                self.assertEqual((path / "notebook.txt").read_bytes(), TEXT.encode())
                self.assertEqual(len(replies[-1]["result"]["content"]), 2 if mixed else 1)
                records = [json.loads(line) for line in (path / "peer.jsonl").read_text().splitlines()]
                self.assertEqual([json.loads(r["request"]) for r in records if "request" in r], messages)
                self.assertEqual(records[-1], {"eof": True})


if __name__ == "__main__":
    unittest.main(verbosity=2)
