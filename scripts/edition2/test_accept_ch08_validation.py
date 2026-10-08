"""Checker fixture/assertion controls only; no Chapter 8 runtime pass is claimed."""
import json
import unittest
from accept_ch08_validation import SENTINEL, check_refusal, check_startup_refusal, encoded, file_cases, patch_cases, seed


class ValidationControls(unittest.TestCase):
    def test_exact_size_positive_keeps_nondefault_semantics(self):
        for domain in ('preferences', 'policy'):
            cases = {name: (body, valid) for name, body, valid in file_cases(domain)}
            raw, valid = cases['exact-65536']
            self.assertTrue(valid)
            self.assertEqual(len(raw), 65536)
            self.assertEqual(json.loads(raw), seed(domain))
            large, valid = cases['oversize-65537']
            self.assertFalse(valid)
            self.assertEqual(len(large), 65537)
            self.assertEqual(json.loads(large), seed(domain))

    def test_single_field_cases_do_not_accidentally_duplicate_keys(self):
        def reject_duplicates(pairs):
            result = {}
            for key, value in pairs:
                if key in result:
                    raise ValueError('duplicate')
                result[key] = value
            return result
        for domain in ('preferences', 'policy'):
            for name, raw, _ in file_cases(domain):
                if name.startswith('invalid-members-'):
                    if name.endswith('duplicate-field'):
                        with self.assertRaisesRegex(ValueError, 'duplicate'):
                            json.loads(raw, object_pairs_hook=reject_duplicates)
                    else:
                        value = json.loads(raw, object_pairs_hook=reject_duplicates)
                        self.assertTrue(set(seed(domain)[domain]) <= set(value[domain]))

    def test_refusal_requires_correlation_reason_and_correct_code(self):
        good = dict(type='error', id='p', code='invalid_policy', message='invalid policy field')
        check_refusal(good, 'p', 'policy')
        for patch, reason in [(dict(id='other'), 'correlated'), (dict(code='revision_conflict'), 'intended'), (dict(message=''), 'reason'), (dict(message=SENTINEL), 'echoed')]:
            with self.assertRaisesRegex(AssertionError, reason):
                check_refusal(good | patch, 'p', 'policy')

    def test_startup_failure_must_be_refusal_not_timeout_or_bad_fixture(self):
        raw = encoded(seed('policy'))
        check_startup_refusal(1, ['invalid policy value\n'], raw, raw)
        for code, lines, after, reason in [(None, [], raw, 'did not fail'), (0, ['done'], raw, 'did not fail'), (1, ['http://127.0.0.1:8088'], raw, 'listening'), (1, [], raw, 'lacks reason'), (1, [SENTINEL], raw, 'echoed'), (1, ['invalid file'], b'{}', 'overwritten')]:
            with self.assertRaisesRegex(AssertionError, reason):
                check_startup_refusal(code, lines, raw, after)

    def test_fixture_names_unique_and_raw_duplicates_preserved(self):
        for domain in ('preferences', 'policy'):
            fixtures = list(file_cases(domain))
            self.assertEqual(len(fixtures), len({name for name, _, _ in fixtures}))
            patches = dict(patch_cases(domain))
            key = b'"autoplay"' if domain == 'preferences' else b'"max_model_requests"'
            self.assertEqual(patches['duplicate-field'].count(key), 2)


if __name__ == '__main__':
    unittest.main()
