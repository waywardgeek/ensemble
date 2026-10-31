"""Harness controls, not a substitute implementation or chapter acceptance.

The executable fixture speaks only the literal example. Mutants isolate the
barrier and identity assertions without depending on an unfinished student.
"""
import os
import pathlib
import tempfile
import unittest

from accept_ch06_clients import VENDORS, check


EXECUTABLE = r'''#!/usr/bin/env python3
import json, os, sys, urllib.request
human = sys.argv[1] == 'chat'
vendor = os.environ['LLM_VENDOR']
mutant = __MUTANT__
def emit(value):
    print(value if isinstance(value, str) else json.dumps(value), flush=True)
if human:
    emit('You>')
sys.stdin.readline()
identity = dict(agent_id='agent-fixture', request_id='request-fixture', operation_id='operation-fixture')
def observation(kind, **kw):
    emit(dict(observation=dict(kind=kind, **identity, **kw)))
if not human:
    emit(dict(accepted='prompt', request_id=identity['request_id']))
    observation('model_begin', delivery='stream')
body = dict(stream=True, stream_options=dict(include_usage=True))
path = '/models/fixture-ch06:streamGenerateContent?alt=sse' if vendor == 'gemini' else '/messages'
request = urllib.request.Request(os.environ['LLM_BASE_URL'] + path, data=json.dumps(body).encode(), headers={'Content-Type':'application/json'})
with urllib.request.urlopen(request) as response:
    first = False
    for line in response:
        if b'Hel' in line and not first:
            first = True
            if mutant != 'buffer':
                if human: emit('Hel')
                else: observation('part_delta', part_id=1, channel='text', text='Hel')
if mutant == 'buffer':
    if human: emit('Hel')
    else: observation('part_delta', part_id=1, channel='text', text='Hel')
if human:
    emit('lo.')
    emit('Request request-fixture (success)')
    sys.stdin.readline()
else:
    if mutant == 'identity': identity['operation_id'] = 'wrong-operation'
    observation('part_delta', part_id=1, channel='text', text='lo.')
    observation('part_final', part_id=1, response_seq=8, part_index=0, part=dict(type='text', text='Hello.'))
    if vendor == 'gemini':
        observation('part_final', part_id=2, response_seq=8, part_index=1, part=dict(type='opaque', data={'thoughtSignature':'fixture-signature', 'text':''}))
    observation('model_end', accepted=True, response_seq=8)
    emit(dict(completion=dict(request_id=identity['request_id'], outcome='success', text='Hello.')))
    sys.stdin.readline()
'''


class HarnessControls(unittest.TestCase):
    def run_fixture(self, vendor, human, mutation='none'):
        with tempfile.TemporaryDirectory(prefix='ch06-harness-control-') as tmp:
            path = pathlib.Path(tmp) / 'fixture-client'
            path.write_text(EXECUTABLE.replace('__MUTANT__', repr(mutation)))
            path.chmod(0o700)
            check(path, vendor, human, timeout=1)

    def test_positive_all_vendors_both_clients(self):
        for vendor in VENDORS:
            for human in (False, True):
                with self.subTest(vendor=vendor, human=human):
                    self.run_fixture(vendor, human)

    def test_buffer_until_terminal_rejected_at_barrier(self):
        for human in (False, True):
            with self.subTest(human=human):
                with self.assertRaisesRegex(AssertionError, '^first fragment absent while terminal frames withheld$'):
                    self.run_fixture('openai', human, 'buffer')

    def test_identity_mutant_rejected_after_passing_barrier(self):
        with self.assertRaisesRegex(AssertionError, '^operation identity changed$'):
            self.run_fixture('openai', False, 'identity')


if __name__ == '__main__':
    unittest.main()
