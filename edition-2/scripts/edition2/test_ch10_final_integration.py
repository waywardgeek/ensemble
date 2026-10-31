#!/usr/bin/env python3
"""Final integration identity controls; subprocess execution is mocked."""
import json
from pathlib import Path
import subprocess
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import Mock, patch

import accept_ch10_final_integration as runner


class Identity(unittest.TestCase):
    def parent(self, directory):
        root = Path(directory)
        cli = root / 'cli'
        cli.write_bytes(b'controlled binary fixture')
        sha = runner.digest(cli)
        build = dict(source_revision='pinned', sources={'solutions/edition-2/main/file.go': 'source-hash'},
                     binaries={'cli': dict(path=str(cli), sha256=sha)},
                     builds={'cli': dict(exit=0, binary_sha256=sha)})
        binding = root / 'binding.json'
        binding.write_text(json.dumps(build))
        args = SimpleNamespace(source_directory=root, source_commit='pinned',
                               build_binding=binding, receipt=root / 'receipt.json')
        return args, build

    def run_fixture(self, args):
        command = Mock(return_value=subprocess.CompletedProcess([], 0, '', ''))
        with patch.object(runner, 'bind', return_value=('pinned', {'file.go': 'source-hash'})), \
                patch.object(runner, 'subprocess', SimpleNamespace(run=command)):
            passed = runner.evaluate(args)
        return passed, command

    def test_valid_parent_records_all_commands(self):
        with tempfile.TemporaryDirectory() as directory:
            args, _ = self.parent(directory)
            passed, command = self.run_fixture(args)
            self.assertTrue(passed)
            receipt = json.loads(args.receipt.read_text())
            self.assertTrue(receipt['complete'])
            self.assertEqual(command.call_count, len(receipt['checks']))
            self.assertGreater(command.call_count, 0)

    def refused_mutation(self, mutate, expected):
        with tempfile.TemporaryDirectory() as directory:
            args, build = self.parent(directory)
            self.assertTrue(self.run_fixture(args)[0])
            before = args.receipt.read_bytes()
            mutate(build)
            args.build_binding.write_text(json.dumps(build))
            command = Mock()
            with patch.object(runner, 'bind', return_value=('pinned', {'file.go': 'source-hash'})), \
                    patch.object(runner, 'subprocess', SimpleNamespace(run=command)):
                with self.assertRaisesRegex(ValueError, expected):
                    runner.evaluate(args)
                command.assert_not_called()
            self.assertEqual(before, args.receipt.read_bytes())

    def test_incomplete_source_map_refuses_before_write(self):
        self.refused_mutation(lambda x: x['sources'].clear(), 'incomplete final executable source map')

    def test_changed_source_hash_refuses_before_write(self):
        self.refused_mutation(lambda x: x['sources'].update({'solutions/edition-2/main/file.go': 'different'}),
                              'incomplete final executable source map')

    def test_changed_binary_hash_refuses_before_write(self):
        self.refused_mutation(lambda x: x['binaries']['cli'].update(sha256='different'), 'final CLI build or hash mismatch')


if __name__ == '__main__':
    unittest.main()
