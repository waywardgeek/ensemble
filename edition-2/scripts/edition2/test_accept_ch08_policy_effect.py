import tempfile
import unittest
from pathlib import Path
from accept_ch08_policy_effect import check_effect, completion


class PolicyEffectAssertions(unittest.TestCase):
    def test_both_actual_effects_and_exact_http_limit(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            for part in (1, 2):
                (directory/f'ONE-1-{part}.txt').write_text(f'ONE:1:{part}\n')
            check_effect(directory, 'ONE', 1, 1)
            with self.assertRaisesRegex(AssertionError, 'HTTP count'):
                check_effect(directory, 'ONE', 1, 2)
            (directory/'ONE-1-2.txt').write_text('unapplied')
            with self.assertRaisesRegex(AssertionError, 'last accepted batch'):
                check_effect(directory, 'ONE', 1, 1)
            (directory/'ONE-1-2.txt').write_text('ONE:1:2\n')
            (directory/'ONE-2-1.txt').write_text('extra')
            with self.assertRaisesRegex(AssertionError, 'effect after'):
                check_effect(directory, 'ONE', 1, 1)

    def test_completion_uses_identity_not_delivery_order(self):
        class Socket:
            records = [dict(type='completion', request_id='queued', outcome='round_limit'),
                       dict(type='completion', request_id='active', outcome='round_limit')]
            def until(self, match):
                raise AssertionError('already received completion was ignored')
        completion(Socket(), 'active', 'round_limit')
        completion(Socket(), 'queued', 'round_limit')
        with self.assertRaisesRegex(AssertionError, 'wrong reliable completion'):
            completion(Socket(), 'queued', 'success')


if __name__ == '__main__':
    unittest.main()
