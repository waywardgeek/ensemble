#!/usr/bin/env python3
"""Fixture-serialization controls only; sample() is not a runtime snapshot."""
import unittest
from accept_ch10 import Number, canonical, digest, encoded_json, envelope, faults, parse, sample


class Serialization(unittest.TestCase):
    def test_hash_encoding_and_mutation_encoding_have_different_purposes(self):
        value = {'counter': Number('10'), 'schema': Number('1.00'), 'raw': '{"n": 1.0}'}
        self.assertEqual(canonical(value), '{"counter":1e1,"raw":"{\\"n\\": 1.0}","schema":1}')
        self.assertEqual(parse(encoded_json(value)), value)

    def test_corruption_changes_one_field_without_retyping_state_integers(self):
        value = sample()
        value['as_of'] = value['high_watermarks']['event'] = 10
        value['state']['counter'] = Number('10')
        value['state']['raw'] = '{"n": 1.0}'
        value['state_sha256'] = digest(canonical(value['state']).encode())
        original = encoded_json(value).encode()
        envelope(original)
        bad, _ = faults(original, b'{"log_version":1}\n{"seq":1,"type":"fixture"}\n')['unknown-outer-field']
        changed = parse(bad)
        self.assertEqual(changed.pop('unexpected'), True)
        self.assertEqual(changed, parse(original))
        self.assertEqual(changed['state']['counter'], Number('10'))
        self.assertEqual(changed['as_of'], Number('10'))
        self.assertEqual(changed['state_sha256'], value['state_sha256'])


if __name__ == '__main__':
    unittest.main()
