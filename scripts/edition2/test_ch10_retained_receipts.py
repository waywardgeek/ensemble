"""Exercise retained-gate orchestration; no Go/runtime acceptance is asserted."""
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

import accept_ch09_gate as gate
import accept_ch10_retained as retained


def archive():
    data = io.BytesIO()
    with tarfile.open(fileobj=data, mode='w') as stream:
        for name in ('go.mod', 'gui/go.mod', 'main.go'):
            body = b'fixture\n'
            entry = tarfile.TarInfo(gate.PREFIX + name)
            entry.size = len(body)
            stream.addfile(entry, io.BytesIO(body))
    return data.getvalue()


class PriorGateControls(unittest.TestCase):
    def test_progress_does_not_change_commands_or_failure_results(self):
        calls = []

        def command(args, cwd, timeout):
            # Absolute temporary directories differ between extractions.
            args = [str(a) for a in args]
            calls.append((Path(args[0]).name,
                          [Path(a).name for a in args[1:]], timeout))
            code = int(any(a.endswith('/accept_ch09_public.py') for a in args))
            return dict(exit=code, stdout='', stderr='fixture refusal' if code else '')

        def run(observer=None):
            with patch.object(gate.subprocess, 'check_output',
                              side_effect=['revision\n', archive()]), \
                 patch.object(gate, 'prepare', return_value={}), \
                 patch.object(gate, 'command', side_effect=command):
                return gate.evaluate('revision', progress=observer)

        original = run()
        original_calls = list(calls)
        calls.clear()
        events = []
        observed = run(events.append)
        self.assertEqual(original, observed)
        self.assertEqual(original_calls, calls)
        self.assertFalse(observed['passed'])
        self.assertTrue(observed['complete_run'])
        self.assertEqual(events[0]['phase'], 'source')
        self.assertEqual([e['id'] for e in events if e['phase'] == 'started'],
                         [r['id'] for r in original['checks']])
        self.assertEqual([e['check'] for e in events if e['phase'] == 'completed'],
                         original['checks'])


class DurableControls(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.path = Path(self.directory.name) / 'receipt.json'

    def saved(self):
        return json.loads(self.path.read_text())

    @staticmethod
    def stages(revision, only, progress, **unused):
        progress(dict(phase='source', source_revision=revision, source_files={'main.go': 'sha'}))
        progress(dict(phase='started', id='one', args=['fixture-one'], cwd='/fixture'))
        progress(dict(phase='completed', check=dict(id='one', passed=True)))
        progress(dict(phase='started', id='two', args=['fixture-two'], cwd='/fixture'))

    def run_gate(self, body, **kwargs):
        with patch.object(retained.subprocess, 'check_output', return_value='revision\n'), \
             patch.object(retained.gate, 'evaluate', side_effect=body), \
             patch.object(retained.shutil, 'disk_usage', return_value=SimpleNamespace(free=2**40)):
            return retained.evaluate('revision', self.path, **kwargs)

    def test_interrupt_retains_completed_result_and_active_command(self):
        def interrupted(*args, **kwargs):
            self.stages(*args, **kwargs)
            raise KeyboardInterrupt()
        with self.assertRaises(KeyboardInterrupt):
            self.run_gate(interrupted)
        saved = self.saved()
        self.assertEqual(saved['checks'], [dict(id='one', passed=True)])
        self.assertEqual(saved['active_check']['id'], 'two')
        self.assertFalse(saved['passed'])
        self.assertFalse(saved['complete_run'])

    def test_late_failure_retains_prior_result_and_error(self):
        def failed(*args, **kwargs):
            self.stages(*args, **kwargs)
            raise OSError('fixture disk exhaustion')
        result = self.run_gate(failed)
        self.assertEqual(result, self.saved())
        self.assertEqual(result['checks'], [dict(id='one', passed=True)])
        self.assertEqual(result['active_check']['id'], 'two')
        self.assertIn('fixture disk exhaustion', result['error'])
        self.assertFalse(result['complete_run'])

    def test_source_resolution_failure_preserves_existing_receipt(self):
        self.path.write_bytes(b'ORIGINAL')
        with patch.object(retained.subprocess, 'check_output',
                          side_effect=subprocess.CalledProcessError(128, ['git'])), \
             patch.object(retained.gate, 'evaluate') as evaluate:
            with self.assertRaises(subprocess.CalledProcessError):
                retained.evaluate('missing', self.path)
            evaluate.assert_not_called()
        self.assertEqual(self.path.read_bytes(), b'ORIGINAL')

    def test_successful_subset_is_not_full_retained_acceptance(self):
        def finished(*args, **kwargs):
            self.stages(*args, **kwargs)
            kwargs['progress'](dict(phase='completed', check=dict(id='two', passed=True)))
            return dict(passed=True, complete_run=False)
        result = self.run_gate(finished, only=['two'])
        self.assertTrue(result['passed'])
        self.assertTrue(result['complete_run'])
        self.assertFalse(result['full_retained_run'])
        self.assertFalse(result['full_chapter_acceptance'])
        self.assertEqual(result['selected_checks'], ['two'])
        self.assertIsNone(result['active_check'])

    def test_low_space_without_opt_in_never_clears_cache(self):
        observer = retained.Progress(self.path)
        with patch.object(retained.shutil, 'disk_usage', return_value=SimpleNamespace(free=0)), \
             patch.object(retained.subprocess, 'run') as command:
            with self.assertRaisesRegex(RuntimeError, 'no cache changed'):
                observer(dict(phase='started', id='next', args=['fixture'], cwd='/fixture'))
            command.assert_not_called()
        self.assertEqual(self.saved()['active_check']['id'], 'next')

    def test_active_compiler_blocks_opted_in_cache_cleanup(self):
        observer = retained.Progress(self.path, maintain_cache=True)
        with patch.object(retained.shutil, 'disk_usage', return_value=SimpleNamespace(free=0)), \
             patch.object(retained.subprocess, 'check_output', return_value='/bin/zsh\n/go/pkg/tool/compile\n'), \
             patch.object(retained.subprocess, 'run') as command:
            with self.assertRaisesRegex(RuntimeError, 'Another Go build'):
                observer.ensure_space()
            command.assert_not_called()

    def test_explicit_idle_cleanup_is_recorded_before_continuing(self):
        observer = retained.Progress(self.path, maintain_cache=True)
        with patch.object(retained.shutil, 'disk_usage',
                          side_effect=[SimpleNamespace(free=0), SimpleNamespace(free=2**40)]), \
             patch.object(retained.subprocess, 'check_output', return_value='/bin/zsh\n'), \
             patch.object(retained.subprocess, 'run',
                          return_value=SimpleNamespace(returncode=0, stdout='', stderr='')) as command:
            observer.ensure_space()
        self.assertEqual(command.call_args.args[0], ['go', 'clean', '-cache'])
        self.assertEqual(self.saved()['maintenance'][0]['after_bytes'], 2**40)

    def test_callback_failure_is_a_durable_incomplete_result(self):
        def started(revision, only, progress, **unused):
            progress(dict(phase='started', id='blocked', args=['fixture'], cwd='/fixture'))
        with patch.object(retained.Progress, 'ensure_space', side_effect=RuntimeError('fixture reserve')):
            result = self.run_gate(started)
        self.assertEqual(result, self.saved())
        self.assertEqual(result['active_check']['id'], 'blocked')
        self.assertIn('fixture reserve', result['error'])
        self.assertFalse(result['complete_run'])


if __name__ == '__main__':
    unittest.main()
