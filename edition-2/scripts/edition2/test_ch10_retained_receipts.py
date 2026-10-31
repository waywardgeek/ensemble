"""Exercise retained-gate orchestration; no Go/runtime acceptance is asserted."""
import ast
import hashlib
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
import ch10_retained_adapters as adapters


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
    def test_only_named_direct_checkers_use_prepared_bundle(self):
        selected = ('ch09-review-record-bounds.py', 'ch09-review-boundaries.py', 'accept_ch09_management.py')
        calls = []

        def command(args, cwd, timeout):
            calls.append(tuple(map(str, args)))
            return dict(exit=0, stdout='', stderr='')

        def capture(selected=()):
            calls.clear()
            with patch.object(gate.subprocess, 'check_output',
                              side_effect=['revision\n', archive()]), \
                 patch.object(gate, 'prepare', return_value={}), \
                 patch.object(gate, 'command', side_effect=command):
                result = gate.evaluate('revision', prepared_checkers=selected)
            return {row['id']: args for row, args in zip(result['checks'], calls)}

        original = capture()
        adapted = capture(selected)
        self.assertEqual(set(original), set(adapted))
        for check, args in original.items():
            for index, arg in enumerate(args):
                path = Path(arg)
                if path.parent != gate.HERE:
                    continue
                changed = Path(adapted[check][index])
                if path.name in selected:
                    self.assertEqual(changed.parent.name, 'checkers')
                    self.assertEqual(changed.name, path.name)
                else:
                    self.assertEqual(changed, path)
        for check in ('accept_ch09', 'accept_ch09_catalog', 'accept_ch09_graph',
                      'accept_ch09_configuration',
                      'accept_ch09_replay'):
            self.assertEqual(Path(adapted[check][1]), gate.HERE / (check + '.py'))
        self.assertEqual(Path(adapted['accept_ch09_management'][1]).parent.name, 'checkers')
        self.assertEqual(Path(adapted['complete-delivered-package-discovery'][1]),
                         gate.HERE / 'accept_delivered_tree.py')
        # Existing retained scripts still intentionally execute from the bundle.
        self.assertEqual(Path(adapted['retained-ch06-deletions'][1]).parent.name, 'checkers')

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
            self.assertEqual(kwargs['prepared_checkers'],
                             ('ch09-review-record-bounds.py', 'ch09-review-boundaries.py', 'accept_ch09_management.py'))
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


class AdapterControls(unittest.TestCase):
    def test_final_part_adapter_changes_only_unused_collection_comparison(self):
        original = (gate.HERE / 'ch06_public_test.go.txt').read_bytes()
        with tempfile.TemporaryDirectory() as directory:
            bundle = Path(directory) / 'bundle'
            adapters.prepare(bundle)
            text = (bundle / 'ch06_public_test.go.txt').read_text()
            text = text.replace(adapters.FINAL_PART_HELPER, '')
            text = text.replace('ch10EqualFinalPart(event.Response.Parts[facts[2].PartIndex], *facts[2].Part)',
                                'reflect.DeepEqual(event.Response.Parts[facts[2].PartIndex], *facts[2].Part)')
            text = text.replace('ch10EqualFinalPart(*x.Part, response.Parts[i])',
                                'reflect.DeepEqual(*x.Part, response.Parts[i])')
            self.assertEqual(text.encode(), original)
            self.assertEqual((gate.HERE / 'ch06_public_test.go.txt').read_bytes(), original)
            self.assertIn('a.Type != "tool_result" && len(a.Parts) == 0', adapters.FINAL_PART_HELPER)
            self.assertIn('b.Type != "tool_result" && len(b.Parts) == 0', adapters.FINAL_PART_HELPER)
            self.assertIn('return reflect.DeepEqual(a, b)', adapters.FINAL_PART_HELPER)

    def test_duplicate_name_adaptation_preserves_every_other_expectation_and_assertion(self):
        path = gate.HERE / 'accept_ch09_management.py'
        original = path.read_bytes()
        with tempfile.TemporaryDirectory() as directory:
            bundle = Path(directory) / 'bundle'
            adapters.prepare(bundle)
            adapted = (bundle / path.name).read_bytes()
            self.assertEqual(path.read_bytes(), original)
            before, after = ast.parse(original), ast.parse(adapted)
            def invalid(tree):
                return ast.literal_eval(next(node.value for node in tree.body
                    if isinstance(node, ast.Assign) and node.targets[0].id == 'INVALID'))
            expected = invalid(before)
            old = next(row for row in expected if row[0] == 'duplicate')
            self.assertEqual(old[2], 'edit')
            self.assertEqual(invalid(after), [row[:2] + ('',) if row[0] == 'duplicate' else row for row in expected])
            def evaluate(tree):
                return ast.dump(next(node for node in tree.body
                    if isinstance(node, ast.FunctionDef) and node.name == 'evaluate'))
            self.assertEqual(evaluate(before), evaluate(after))
            # Execute imports only; original helper ROOT and file identity must
            # survive even though the adapted direct script is in the bundle.
            import importlib.util
            spec = importlib.util.spec_from_file_location('adapted_management_control', bundle / path.name)
            module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(module)
            self.assertEqual(module.ROOT, gate.REPO)

    def test_mutations_keep_assertions_and_target_only_the_owned_boundary(self):
        originals = {name: (gate.HERE / name).read_bytes() for name in
                     ('ch09-review-boundaries.py', 'audit_ch06_mutations.py')}
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'source'
            speech = source / 'gui/web/gui/speech-service.js'
            speech.parent.mkdir(parents=True)
            speech.write_text('  finish(request, error) {\n')
            bundle = root / 'bundle'
            receipt = adapters.prepare(bundle, source)
            for name, original in originals.items():
                self.assertEqual((gate.HERE / name).read_bytes(), original)
                self.assertEqual(receipt['adapted_files'][name],
                                 hashlib.sha256((bundle / name).read_bytes()).hexdigest())

            original_tree = ast.parse(originals['ch09-review-boundaries.py'])
            adapted_tree = ast.parse((bundle / 'ch09-review-boundaries.py').read_text())
            def run_assertions(tree):
                return [ast.dump(node) for node in ast.walk(tree)
                        if isinstance(node, ast.Call) and isinstance(node.func, ast.Name)
                        and node.func.id == 'run']
            self.assertEqual(run_assertions(original_tree), run_assertions(adapted_tree))

            # Exercise the adapted anchors on a small representation of both
            # append branches; actual compiling deletions remain a separate gate.
            branch = ('\tif err == nil && persist {\n\t\tif encoded != nil {\n'
                      '\t\t\terr = a.log.AppendPrepared(encoded)\n\t\t} else {\n'
                      '\t\t\terr = a.log.Append(owned)\n\t\t}\n'
                      '\t\tif err != nil {\n\t\t\ta.faulted = true\n\t\t}\n\t}\n')
            assignments = [node for node in ast.walk(adapted_tree)
                           if isinstance(node, ast.Assign) and isinstance(node.value, ast.Call)
                           and isinstance(node.value.func, ast.Name) and node.value.func.id == 'replace']
            early = next(node.value for node in assignments
                         if isinstance(node.targets[0], ast.Name) and node.targets[0].id == 'early'
                         and isinstance(node.value.args[0], ast.Name) and node.value.args[0].id == 'early')
            boundary = ast.literal_eval(early.args[1])
            self.assertEqual(branch.count(boundary), 1)
            self.assertEqual(branch.index(boundary), 0)
            rollback = next(node.value for node in assignments
                            if isinstance(node.targets[0], ast.Name) and node.targets[0].id == 'rollback')
            before, after = map(ast.literal_eval, rollback.args[1:])
            self.assertEqual(branch.count(before), 1)
            changed = branch.replace(before, after)
            self.assertGreater(changed.index('a.skills = nil'), changed.index('a.log.Append(owned)'))

            def mutations(text):
                tree = ast.parse(text)
                return ast.literal_eval(next(node.value for node in tree.body
                    if isinstance(node, ast.Assign) and node.targets[0].id == 'MUTATIONS'))
            before = mutations(originals['audit_ch06_mutations.py'])
            after = mutations((bundle / 'audit_ch06_mutations.py').read_text())
            self.assertEqual(len(before), len(after))
            for old, new in zip(before, after):
                if old[0] != 'suppress-thinking-display':
                    self.assertEqual(old, new)
                    continue
                self.assertEqual(old[:2] + old[4:], new[:2] + new[4:])
                ambiguous = (new[2] + '\n\t\treturn nil\n\t}\n}\n'
                             'if field == "partial_json" {\n\tif text == "" {\n\t}\n}')
                self.assertEqual(ambiguous.count(old[2]), 2)
                self.assertEqual(ambiguous.count(new[2]), 1)
                self.assertEqual(ambiguous.replace(new[2], new[3]).count('channel == "thinking"'), 1)


if __name__ == '__main__':
    unittest.main()
