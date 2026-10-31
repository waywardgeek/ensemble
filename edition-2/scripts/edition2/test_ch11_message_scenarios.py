#!/usr/bin/env python3
"""Fixture-only controls; no Ensemble client is compiled or exercised."""
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

from accept_ch11 import Refusal, cancellation, compact, parse
from ch11_message_scenarios import Script, bundle, scenarios


class Controls(unittest.TestCase):
    def test_direct_and_stdio_same_complete_messages(self):
        peer = Path(__file__).with_name("ch11_message_scenarios.py")
        for name, spec in scenarios().items():
            with self.subTest(name=name), tempfile.TemporaryDirectory(prefix="ch11-script-control-") as directory:
                script = Script(spec)
                expected = []
                inputs = []
                for item in spec["steps"]:
                    raw = item["request"].encode()
                    inputs.append(raw)
                    expected.extend(script.accept(raw))
                script.finish()
                receipt = Path(directory) / "peer.jsonl"
                child = subprocess.run([sys.executable, str(peer), "--peer", name, "--receipt", str(receipt)],
                                       input=b"\n".join(inputs) + b"\n", capture_output=True, timeout=5)
                self.assertEqual(child.returncode, 0, child.stderr)
                self.assertEqual(child.stdout, b"".join(value + b"\n" for value in expected))
                records = [json.loads(line) for line in receipt.read_text().splitlines()]
                self.assertEqual([bytes.fromhex(r["incoming_hex"]) for r in records if "incoming_hex" in r], inputs)
                self.assertEqual(records[-1], {"fixture_complete": True, "runtime_acceptance": False})

    def test_intended_script_refusals_from_valid_parent(self):
        spec = scenarios()["ordinary"]
        parent = Script(spec)
        for item in spec["steps"]:
            parent.accept(item["request"].encode())
        parent.finish()
        wrong = spec["steps"][0]["request"].replace("rpc-1", "rpc-2").encode()
        with self.assertRaisesRegex(Refusal, "script-request-mismatch"):
            Script(spec).accept(wrong)
        with self.assertRaisesRegex(Refusal, "script-incomplete"):
            Script(spec).finish()
        with self.assertRaisesRegex(Refusal, "unexpected-extra-message"):
            parent.accept(spec["steps"][-1]["request"].encode())

    def test_cancel_bytes_and_fault_cleanup(self):
        spec = scenarios()["cancel-late-sibling"]
        script = Script(spec)
        for item in spec["steps"][:-1]:
            script.accept(item["request"].encode())
        notice = compact(cancellation("rpc-3"))
        with self.assertRaisesRegex(Refusal, "cancellation-literal-bytes"):
            script.accept(notice + b" ")
        self.assertEqual(len(script.accept(notice)), 2)
        script.finish()
        for name in ("future-id", "reverse-request"):
            script = Script(scenarios()[name])
            for item in script.steps:
                script.accept(item["request"].encode())
            script.finish()
            self.assertEqual(script.accept(notice), [])
            script.finish()

    def test_vector_json_preserves_precision_and_pairing(self):
        corpus = bundle()
        self.assertFalse(corpus["runtime_acceptance"])
        for vector in corpus["argument_vectors"]:
            parse(vector["schema"].encode())
            self.assertTrue(vector["accept"] and vector["reject"])
            for raw in vector["accept"] + vector["reject"]:
                parse(raw.encode())
        precision = next(v for v in corpus["argument_vectors"] if v["id"] == "precision")
        self.assertIn("9007199254740993", precision["schema"])
        self.assertIn("90071992547409930e-1", precision["accept"][1])
        for vector in corpus["preparation_vectors"]:
            parse(vector["positive"].encode())
            parse(vector["reject"].encode())


if __name__ == "__main__":
    unittest.main(verbosity=2)
