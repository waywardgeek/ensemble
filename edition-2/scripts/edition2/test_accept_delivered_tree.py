"""Actual Go discovery controls, including a retained standalone support file."""
import subprocess
import tempfile
import unittest
from pathlib import Path
from accept_delivered_tree import discover


class DeliveredTree(unittest.TestCase):
    def test_evidence_is_not_excluded_and_explicit_helper_still_builds(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            (root/'go.mod').write_text('module example.com/core\n\ngo 1.22\n')
            (root/'core.go').write_text('package core\n')
            optional = root/'optional'
            optional.mkdir()
            (optional/'go.mod').write_text('module example.com/optional\n\ngo 1.22\n')
            (optional/'optional.go').write_text('package optional\n')
            positive = discover(root)
            self.assertTrue(positive['passed'], positive)
            self.assertEqual(len(positive['checks']), 2)
            support = root/'evidence'/'kept'
            support.mkdir(parents=True)
            helper = support/'replay.go'
            body = 'package main\nimport _ "example.com/optional"\nfunc main() {}\n'
            helper.write_text(body)
            negative = discover(root)
            self.assertFalse(negative['passed'])
            failed = [r for r in negative['checks'] if r['exit']]
            self.assertEqual([r['module'] for r in failed], ['.'])
            self.assertIn('no required module provides package example.com/optional', failed[0]['stderr'])
            helper.write_text('//go:build ignore\n\n'+body)
            restored = discover(root)
            self.assertTrue(restored['passed'], restored)
            invocation = subprocess.run(['go', 'build', '-o', str(root/'replay'), str(helper)], cwd=optional, capture_output=True, text=True)
            self.assertEqual(invocation.returncode, 0, invocation.stderr)

    def test_missing_module_tree_cannot_pass(self):
        with tempfile.TemporaryDirectory() as temporary:
            self.assertFalse(discover(Path(temporary))['passed'])


if __name__ == '__main__':
    unittest.main()
