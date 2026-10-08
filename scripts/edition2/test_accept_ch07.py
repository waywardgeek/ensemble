"""Assertion controls only; these do not claim a Chapter 7 runtime exists."""
import copy
import unittest
from accept_ch07 import check_empty_snapshot, check_pause_ack


def snapshot():
    return [dict(type='snapshot_begin',id='c1',generation='g1',agent_id='a1',watermark=0,
                 first_seq=None,last_seq=None,log_seq=0,omitted=0,
                 state=dict(lifecycle='idle',active_request_id=None,active_operation=None,queued_request_ids=[],
                            paused=False,typing_clients=0,speaking_clients=0,model='fixture-ch07-independent',usage=[])),
            dict(type='snapshot_end',generation='g1',watermark=0)]


def pause():
    return [dict(type='observation',generation='g1',revision=42,
                 observation=dict(kind='pause_changed',agent_id='a1',paused=True,typing_clients=1,speaking_clients=1)),
            dict(type='ack',id='c2',revision=42,paused=True,typing_clients=1,speaking_clients=1)]


class Controls(unittest.TestCase):
    def test_snapshot_positive(self):
        check_empty_snapshot(snapshot(),'c1')

    def test_generation_and_watermark_refusal(self):
        for key,value in [('generation','old'),('watermark',1)]:
            rows=snapshot();rows[1][key]=value
            with self.assertRaisesRegex(AssertionError,'snapshot end differs'):check_empty_snapshot(rows,'c1')

    def test_empty_range_refusal(self):
        rows=snapshot();rows[0]['first_seq']=0
        with self.assertRaisesRegex(AssertionError,'empty history range'):check_empty_snapshot(rows,'c1')

    def test_credential_refusal(self):
        rows=snapshot();rows[0]['state']['config']={'api_key':'LOCAL-FIXTURE-NOT-A-SECRET'}
        with self.assertRaisesRegex(AssertionError,'unsafe state configuration'):check_empty_snapshot(rows,'c1')

    def test_pause_positive_and_cause_refusal(self):
        check_pause_ack(pause(),'c2',1,1)
        rows=pause();rows[1]['typing_clients']=0
        with self.assertRaisesRegex(AssertionError,'independent pause causes lost'):check_pause_ack(rows,'c2',1,1)

    def test_ack_before_observation_refusal(self):
        with self.assertRaisesRegex(AssertionError,'ack preceded'):check_pause_ack(list(reversed(pause())),'c2',1,1)


if __name__=='__main__':unittest.main()
