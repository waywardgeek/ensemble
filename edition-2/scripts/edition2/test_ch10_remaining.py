#!/usr/bin/env python3
"""Orchestration tests only. No Go/compiler/Ensemble command is executed."""
import json
from pathlib import Path
import subprocess
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

import accept_ch10_remaining as runner


class Receipts(unittest.TestCase):
    def args(self, directory):
        root = Path(directory)
        return SimpleNamespace(source_directory=root, source_commit="pinned", receipt=root / "receipt.json",
                               audit=False, run="^TestCh10Remaining")

    def test_identity_refusal_does_not_rewrite_receipt(self):
        with tempfile.TemporaryDirectory() as d:
            args = self.args(d)
            args.receipt.write_text("prior evidence\n")
            with patch.object(runner, "bind", side_effect=ValueError("identity")), \
                    patch.object(runner.subprocess, "run") as command:
                with self.assertRaisesRegex(ValueError, "identity"):
                    runner.evaluate(args)
                command.assert_not_called()
            self.assertEqual(args.receipt.read_text(), "prior evidence\n")

    def test_interruption_retains_started_and_completed_commands(self):
        with tempfile.TemporaryDirectory() as d:
            args = self.args(d)
            good = subprocess.CompletedProcess([], 0, "", "")
            with patch.object(runner, "bind", return_value=("pinned", {"source.go": "hash"})), \
                    patch.object(runner.subprocess, "run", side_effect=[good, KeyboardInterrupt()]):
                with self.assertRaises(KeyboardInterrupt):
                    runner.evaluate(args)
            receipt = json.loads(args.receipt.read_text())
            self.assertFalse(receipt["complete"])
            self.assertEqual(len(receipt["checks"]), 2)
            self.assertTrue(receipt["checks"][0]["complete"])
            self.assertFalse(receipt["checks"][1]["complete"])

    def test_failed_positive_stops_before_mutations(self):
        with tempfile.TemporaryDirectory() as d:
            args = self.args(d)
            args.audit = True
            bad = subprocess.CompletedProcess([], 1, "fixture failure", "")
            with patch.object(runner, "bind", return_value=("pinned", {"source.go": "hash"})), \
                    patch.object(runner.subprocess, "run", return_value=bad) as command:
                self.assertFalse(runner.evaluate(args))
                self.assertEqual(command.call_count, 1)
            receipt = json.loads(args.receipt.read_text())
            self.assertTrue(receipt["complete"])
            self.assertFalse(receipt["passed"])
            self.assertNotIn("mutations", receipt)


if __name__ == "__main__":
    unittest.main()
