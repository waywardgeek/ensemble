"""Preparation-only oracle/binding tests; no Go compilation or runtime claim."""
import base64
import copy
import importlib.util
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import accept_ch10_arguments as runner
from accept_ch10 import canonical, digest, encoded_json, envelope, parse, sample

SPEC = importlib.util.spec_from_file_location('argument_cases', Path(__file__).with_name('ch10-arguments-cases.py'))
CASES = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CASES)


class GeneratorControls(unittest.TestCase):
    def setUp(self):
        # This is only a generator parent, never claimed to be a valid runtime
        # snapshot. The Go consumer supplies and proves that separate positive.
        self.parent = sample()
        part = {'args': CASES.DUPLICATE, 'name': 'load_skill', 'call_id': 'argument-duplicate'}
        self.parent['state'] = {'context': {'Calls': {'argument-duplicate': {'Part': part}},
                                          'Entries': [copy.deepcopy(part)]},
                                'window': {'events': [{'event': {'seq': 1, 'part': copy.deepcopy(part)}}]}}
        self.parent['state_sha256'] = digest(canonical(self.parent['state']).encode())
        self.raw = encoded_json(self.parent).encode()
        self.rows = CASES.cases(self.raw)

    def test_true_generator_parents_precede_negatives(self):
        self.assertEqual(len(self.rows), 12)
        self.assertEqual([r['name'] for r in self.rows if r['positive']],
                         ['valid-export-reencoded', 'valid-alternate-duplicate-text'])
        for row in self.rows[:2]:
            value = envelope(base64.b64decode(row['bytes']))
            args = value['state']['context']['Calls']['argument-duplicate']['Part']['args']
            self.assertEqual(args, CASES.DUPLICATE if row == self.rows[0] else CASES.ALTERNATE)
        self.assertEqual(encoded_json(self.parent).encode(), self.raw)

    def test_correspondence_negatives_change_only_one_argument_witness(self):
        for row in self.rows[3:5]:
            value = envelope(base64.b64decode(row['bytes']))
            changed = value['state']['context']['Calls']['argument-duplicate']['Part']['args']
            self.assertNotEqual(changed, CASES.DUPLICATE)
            self.assertEqual(value['state']['context']['Entries'][0]['args'], CASES.DUPLICATE)
            self.assertEqual(value['state']['window'], parse(self.raw)['state']['window'])
            value['state']['context']['Calls']['argument-duplicate']['Part']['args'] = CASES.DUPLICATE
            self.assertEqual(canonical(value['state']), canonical(self.parent['state']))

    def test_grammar_negatives_keep_correspondence_equal_and_repair_hash(self):
        for row in self.rows[5:9]:
            state = envelope(base64.b64decode(row['bytes']))['state']
            argument = state['context']['Calls']['argument-duplicate']['Part']['args']
            self.assertNotEqual(argument, CASES.DUPLICATE)
            self.assertEqual(state['context']['Entries'][0]['args'], argument)
            self.assertEqual(state['window']['events'][0]['event']['part']['args'], argument)

    def test_stale_hash_case_has_a_separately_valid_same_state_parent(self):
        good = parse(base64.b64decode(self.rows[1]['bytes']))
        stale = parse(base64.b64decode(self.rows[2]['bytes']))
        self.assertEqual(good['state'], stale['state'])
        self.assertNotEqual(good['state_sha256'], stale['state_sha256'])
        with self.assertRaisesRegex(ValueError, 'state hash'):
            envelope(base64.b64decode(self.rows[2]['bytes']))
        self.assertNotEqual(canonical(CASES.DUPLICATE), canonical('{"name":"edit"}'))

    def test_structural_duplicates_are_real_not_escaped_argument_text(self):
        for row in self.rows[9:]:
            with self.assertRaisesRegex(ValueError, 'duplicate member'):
                envelope(base64.b64decode(row['bytes']))
        # The unchanged wrapped duplicate itself does not trip outer validation.
        envelope(self.raw)

    def test_wrong_parent_refuses_generation(self):
        self.parent['state']['context']['Calls']['argument-duplicate']['Part']['args'] = '{}'
        self.parent['state_sha256'] = digest(canonical(self.parent['state']).encode())
        with self.assertRaisesRegex(ValueError, 'genuine duplicate-argument parent absent'):
            CASES.cases(encoded_json(self.parent).encode())


class BindingControls(unittest.TestCase):
    def test_complete_new_only_map_checks_removed_and_changed_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            repo = Path(directory)
            source = repo / runner.PREFIX
            source.mkdir(parents=True)
            payloads = {'go.mod': 'module fixture\n', 'main.go': 'package fixture\n',
                        'evidence/receipt.json': '{"original":true}\n'}
            for name, text in payloads.items():
                p = source / name
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_text(text)
            subprocess.run(['git', 'init', '-q'], cwd=repo, check=True)
            paths = [runner.PREFIX + name for name in payloads]
            subprocess.run(['git', 'add', '--', *paths], cwd=repo, check=True)
            subprocess.run(['git', '-c', 'user.name=Fixture', '-c', 'user.email=fixture@example.invalid',
                            'commit', '-qm', 'binding positive', '--', *paths], cwd=repo, check=True)
            with patch.object(runner, 'REPO', repo):
                _, identities = runner.bind_full(source.resolve(), 'HEAD')
                self.assertEqual(set(identities), set(payloads))
                evidence = source / 'evidence/receipt.json'
                evidence.write_text('changed')
                with self.assertRaisesRegex(ValueError, 'frozen source mismatch: evidence/receipt.json'):
                    runner.bind_full(source.resolve(), 'HEAD')
                evidence.unlink()
                with self.assertRaisesRegex(ValueError, 'missing/escaping frozen file: evidence/receipt.json'):
                    runner.bind_full(source.resolve(), 'HEAD')
                evidence.write_text(payloads['evidence/receipt.json'])
                (source / 'extra.go').write_text('package fixture\n')
                with self.assertRaisesRegex(ValueError, 'unbound build input: extra.go'):
                    runner.bind_full(source.resolve(), 'HEAD')

    def test_binding_failure_preserves_receipt_and_never_launches_compiler(self):
        with tempfile.TemporaryDirectory() as directory:
            receipt = Path(directory) / 'prior.json'
            receipt.write_text('original receipt\n')
            with patch.object(runner, 'bind_full', side_effect=ValueError('source mismatch')), \
                 patch.object(runner.subprocess, 'run') as launch:
                with self.assertRaisesRegex(ValueError, 'source mismatch'):
                    runner.evaluate(Path(directory), 'frozen', receipt)
                launch.assert_not_called()
            self.assertEqual(receipt.read_text(), 'original receipt\n')

    def test_checker_identity_closure_includes_oracle_and_local_imports(self):
        names = {p.name for p in runner.checker_paths()}
        self.assertTrue({'ch10-arguments-cases.py', 'ch10-arguments-public_test.go',
                         'accept_ch10.py', 'accept_ch09.py', 'accept_ch10_lifecycle.py'} <= names)


if __name__ == '__main__':
    unittest.main()
