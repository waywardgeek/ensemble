#!/usr/bin/env python3
"""Oracle-only controls. These synthetic records are not Ensemble evidence."""
import copy
import unittest
from accept_ch10_clients import saved, session, subscribe


class RecordedSocket:
    def __init__(self, incoming):
        self.incoming = iter(incoming)
        self.records = []
        self.sent = []

    def send(self, value):
        self.sent.append(value)

    def until(self, predicate):
        for value in self.incoming:
            self.records.append(value)
            if predicate(value):
                return value
        raise AssertionError('missing record')


class Oracles(unittest.TestCase):
    def setUp(self):
        self.previous = dict(id='a' * 32, resumed=False, checkpoint_seq=None)
        self.change = dict(type='observation', revision=7,
                           observation=dict(kind='session_changed', agent_id='agent-2',
                                            session=dict(self.previous, checkpoint_seq=10)))
        self.ack = dict(type='command_ack', id='save', status='saved', as_of=10, watch_revision=7)

    def test_checkpoint_positive_and_distinguishing_refusals(self):
        valid = [self.change, self.ack]
        self.assertEqual(saved(valid, 'save', self.previous, 'agent-2')[0], self.ack)
        negatives = []
        negatives.append(('missing publication', [self.ack], 'matching applied observation'))
        negatives.append(('late publication', [self.ack, self.change], 'matching applied observation'))
        for label, path, value, reason in [
            ('wrong owner', ('observation', 'agent_id'), 'agent-3', 'owner/kind'),
            ('wrong kind', ('observation', 'kind'), 'state', 'owner/kind'),
            ('different file anchor', ('observation', 'session', 'checkpoint_seq'), 11, 'different anchor'),
            ('identity replaced', ('observation', 'session', 'id'), 'b' * 32, 'SessionID'),
            ('mount provenance changed', ('observation', 'session', 'resumed'), True, 'mount provenance'),
        ]:
            item = copy.deepcopy(self.change)
            target = item
            for part in path[:-1]:
                target = target[part]
            target[path[-1]] = value
            negatives.append((label, [item, self.ack], reason))
        for label, records, reason in negatives:
            with self.subTest(label=label), self.assertRaisesRegex(AssertionError, reason):
                saved(records, 'save', self.previous, 'agent-2')

    def test_session_exact_shape_and_types(self):
        self.assertEqual(session(self.previous, False), self.previous)
        for changed in (dict(self.previous, path='/private/store'), dict(self.previous, resumed=0),
                        dict(self.previous, checkpoint_seq=True), dict(self.previous, checkpoint_seq=2**64)):
            with self.subTest(changed=changed), self.assertRaises(AssertionError):
                session(changed)

    def test_inherited_snapshot_end_has_no_command_id(self):
        begin = dict(type='snapshot_begin', id='subscribe', agent_id='agent-2', generation='g', watermark=3,
                     state=dict(session=self.previous, job_access=[]))
        end = dict(type='snapshot_end', generation='g', watermark=3)
        client = RecordedSocket([dict(type='preferences_snapshot'), begin, end])
        self.assertEqual(subscribe(client), begin)
        self.assertEqual(client.sent, [dict(type='subscribe', id='subscribe')])


if __name__ == '__main__':
    unittest.main()
