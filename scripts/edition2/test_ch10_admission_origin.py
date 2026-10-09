#!/usr/bin/env python3
"""Receipt-only controls; no Go or Ensemble command is executed."""
import unittest
from unittest.mock import patch

import accept_ch10_admission_origin
import test_ch10_remaining

class Receipts(test_ch10_remaining.Receipts):
    def args(self, directory):
        args = super().args(directory)
        store = args.source_directory / 'internal/persistence/store.go'
        store.parent.mkdir(parents=True)
        store.write_text('f, err := os.OpenFile(filepath.Join(s.path, "origin.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)\n')
        return args

    def setUp(self):
        self.runner_patch=patch.object(test_ch10_remaining,'runner',accept_ch10_admission_origin)
        self.runner_patch.start()
        self.addCleanup(self.runner_patch.stop)

if __name__=='__main__':unittest.main()
