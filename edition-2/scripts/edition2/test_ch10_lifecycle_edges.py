#!/usr/bin/env python3
"""Reuse receipt-only orchestration controls; never invoke Go or Ensemble."""
import json
from pathlib import Path
import signal
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import accept_ch10_lifecycle_edges
import test_ch10_remaining

class Receipts(test_ch10_remaining.Receipts):
    def setUp(self):
        self.runner_patch = patch.object(test_ch10_remaining, "runner", accept_ch10_lifecycle_edges)
        self.runner_patch.start()
        self.addCleanup(self.runner_patch.stop)


class OwnedCleanup(unittest.TestCase):
    def fixture(self, root):
        root = Path(root)
        (root / "edge-child.pid").write_text("100002")
        ledger = root / "ledger.json"
        ledger.write_text(json.dumps(dict(workspace=str(root), marker="ch10-edge-owned-123",
                                         holder_pid=100001, binary="/review/ensemble.test")))
        return ledger

    def test_unrelated_reused_pid_is_never_signaled(self):
        with tempfile.TemporaryDirectory() as d:
            ledger = self.fixture(d)
            response = subprocess.CompletedProcess([], 0, "unrelated-user-process", "")
            with patch.object(accept_ch10_lifecycle_edges.subprocess, "run", return_value=response), \
                    patch.object(accept_ch10_lifecycle_edges.os, "kill") as kill:
                rows = accept_ch10_lifecycle_edges.cleanup_owned(ledger)
                kill.assert_not_called()
            self.assertTrue(all(not r["identity_matched"] for r in rows))

    def test_only_matching_owned_processes_are_signaled(self):
        with tempfile.TemporaryDirectory() as d:
            ledger = self.fixture(d)
            responses = [subprocess.CompletedProcess([], 0, "sh -c : ch10-edge-owned-123; heartbeat", ""),
                         subprocess.CompletedProcess([], 0, "/review/ensemble.test -test.run=^TestCh10EdgesLockHolder$ -test.v", "")]
            with patch.object(accept_ch10_lifecycle_edges.subprocess, "run", side_effect=responses), \
                    patch.object(accept_ch10_lifecycle_edges.os, "kill") as kill:
                rows = accept_ch10_lifecycle_edges.cleanup_owned(ledger)
                self.assertEqual(kill.call_args_list, [unittest.mock.call(100002, signal.SIGKILL),
                                                       unittest.mock.call(100001, signal.SIGKILL)])
            self.assertTrue(all(r["killed"] for r in rows))
            self.assertFalse(ledger.exists())


class MutationCredit(unittest.TestCase):
    def test_only_intended_runtime_assertion_receives_credit(self):
        name, group = "lock-inherited-by-exec", "SurvivingChildLock"
        failed = "--- FAIL: TestCh10EdgesSurvivingChildLock (1s)\n"
        intended = "surviving tool child retained store lock: session_in_use\n"
        for output, expected in [(failed + intended, True),
                                 (failed + "independent holder did not finish\n", False),
                                 (failed + intended + "panic: test timed out\n", False),
                                 ("FAIL [build failed]\n", False),
                                 (intended, False)]:
            with self.subTest(output=output):
                process = subprocess.CompletedProcess([], 1, output, "")
                self.assertEqual(accept_ch10_lifecycle_edges.mutation_detected(process, name, group), expected)

if __name__ == "__main__":
    unittest.main()
