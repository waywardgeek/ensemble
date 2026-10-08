"""Orchestration-only controls; no Ensemble runtime acceptance is asserted."""
import importlib.util
import json
import tempfile
import types
import unittest
from pathlib import Path
from unittest.mock import patch

SPEC = importlib.util.spec_from_file_location('ch10_lifecycle', Path(__file__).with_name('accept_ch10_lifecycle.py'))
RUNNER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(RUNNER)


class ReceiptControls(unittest.TestCase):
    def test_identity_refusal_precedes_any_receipt_write(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'receipt.json'
            path.write_bytes(b'ORIGINAL')
            with patch.object(RUNNER, 'bind', side_effect=ValueError('frozen source identity')), \
                 patch.object(RUNNER.subprocess, 'run') as command:
                with self.assertRaisesRegex(ValueError, 'frozen source identity'):
                    RUNNER.evaluate(Path(directory), 'revision', '^fixture$', path)
                command.assert_not_called()
            self.assertEqual(path.read_bytes(), b'ORIGINAL')

    def test_completed_command_survives_interrupted_next_stage(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'receipt.json'
            ok = types.SimpleNamespace(returncode=0, stdout='', stderr='')
            with patch.object(RUNNER, 'bind', return_value=('revision', {'source.go': 'sha'})), \
                 patch.object(RUNNER.subprocess, 'run', side_effect=[ok, KeyboardInterrupt()]):
                with self.assertRaises(KeyboardInterrupt):
                    RUNNER.evaluate(Path(directory), 'revision', '^fixture$', path)
            saved = json.loads(path.read_text())
            self.assertFalse(saved['passed'])
            self.assertFalse(saved['complete_run'])
            self.assertEqual(len(saved['checks']), 1)
            self.assertTrue(saved['checks'][0]['passed'])
            self.assertEqual(saved['active_command'][:2], ['go', 'vet'])

    def test_completed_orchestration_control(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'receipt.json'
            source = {'source.go': 'sha'}
            ok = types.SimpleNamespace(returncode=0, stdout='', stderr='')
            with patch.object(RUNNER, 'bind', return_value=('revision', source)), \
                 patch.object(RUNNER, 'files', return_value=source), \
                 patch.object(RUNNER.subprocess, 'run', return_value=ok):
                result = RUNNER.evaluate(Path(directory), 'revision', '^fixture$', path)
            self.assertTrue(result['passed'])
            self.assertTrue(result['complete_run'])
            self.assertEqual(len(result['checks']), 3)
            self.assertEqual(result, json.loads(path.read_text()))


if __name__ == '__main__':
    unittest.main()
