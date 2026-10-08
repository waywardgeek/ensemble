"""Generator controls only. The skeletal state is NOT an Ensemble snapshot."""
import base64
import importlib.util
from pathlib import Path
import unittest

PATH = Path(__file__).with_name('ch10-semantic-cases.py')
SPEC = importlib.util.spec_from_file_location('semantic_cases', PATH)
SEM = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(SEM)


class GeneratorTests(unittest.TestCase):
    def setUp(self):
        identity = {'mode': 'plain', 'system': 'oracle-only 😀 � literal \\ud800', 'skills': None,
                    'handlers': [{'name': 'read_file', 'description': 'fixture', 'schema': {'minimum': 1}}]}
        water = {'event': 10, 'request': 1, 'activation': 0, 'job': 0}
        state = {
            'session': {'id': 'a' * 32, 'identity': identity, 'as_of': 10, 'high_watermarks': water},
            'context': {'LastSeq': 10, 'SkillMode': False,
                        'Responses': [{'seq': 4, 'raw_usage': '{"raw_marker":"�"}'}],
                        'Turns': {'r1': {'index': 1}}},
            'usage': [{'from': {'vendor': 'openai'}, 'usage': {'input': 3}}],
            'skills': None, 'limits': None,
            'window': {'events': [{'event': {'seq': 4}, 'activations': []}],
                       'event_count': 10, 'renderable_count': 1},
        }
        self.original = {'version': 1, 'session_id': 'a' * 32, 'identity': identity,
                         'as_of': 10, 'high_watermarks': water, 'state_version': 1,
                         'state': state, 'state_sha256': SEM.digest(SEM.canonical(state).encode())}
        self.generated = SEM.cases(SEM.wire(self.original).encode())

    def test_positive_envelopes_preserve_identity(self):
        self.assertEqual(len(self.generated['positives']), 3)
        for case in self.generated['positives']:
            v = SEM.envelope(base64.b64decode(case['bytes']))
            self.assertEqual(SEM.canonical(v['identity']), SEM.canonical(self.original['identity']))
            self.assertEqual(str(v['as_of']), '10')  # Never exponent-normalize an identity field.

    def test_semantic_mutations_repair_hash_not_structural_integer_tokens(self):
        self.assertEqual(len(self.generated['cases']), 47)
        for case in self.generated['cases']:
            if '/duplicate-' in case['name'] or case['name'].startswith('unicode-lone-'):
                continue
            v = SEM.envelope(base64.b64decode(case['bytes']))
            self.assertEqual(str(v['as_of']), '10')

    def test_duplicate_controls_fail_for_duplicate(self):
        for case in self.generated['cases']:
            if '/duplicate-' in case['name']:
                with self.assertRaisesRegex(ValueError, 'duplicate member'):
                    SEM.envelope(base64.b64decode(case['bytes']))


if __name__ == '__main__':
    unittest.main()
