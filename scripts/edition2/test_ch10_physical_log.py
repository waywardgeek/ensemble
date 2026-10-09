#!/usr/bin/env python3
"""Receipt-only controls; no Go or Ensemble command is executed."""
import unittest
from unittest.mock import patch

import accept_ch10_physical_log
import test_ch10_remaining


class Receipts(test_ch10_remaining.Receipts):
    def setUp(self):
        self.runner_patch = patch.object(test_ch10_remaining, 'runner', accept_ch10_physical_log)
        self.runner_patch.start()
        self.addCleanup(self.runner_patch.stop)


if __name__ == '__main__':
    unittest.main()
