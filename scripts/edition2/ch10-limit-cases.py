#!/usr/bin/env python3
"""Ordering mutations of a genuine accepted post-consumption session export.

CalledAt/ReturnedAt/RequestSeqs come from published 81aa8cf/b926929 grammar.
The public caller must inspect its original and verify old calls left the watch
window before invoking this helper. Generation alone is not runtime evidence.
"""
import base64
import copy
import json
import sys
from accept_ch10 import canonical, digest, encoded_json, envelope


def cases(raw):
    original = envelope(raw)
    context = original['state']['context']
    calls = context['Calls']
    assert calls['setter']['Returned'] and calls['next']['Returned'], 'actual completed call positive required'
    window = original['state']['window']['events']
    assert window and all(int(x['event']['seq']) > int(calls['next']['ReturnedAt']) for x in window), 'old calls remain in window or positive window absent; cannot distinguish retained witnesses'
    facts = context['LimitFacts']
    set_index = next(i for i,x in enumerate(facts) if x['kind']=='set' and x['call_id']=='setter')
    consumed_index = next(i for i,x in enumerate(facts) if x['kind']=='consumed' and x['call_id']=='next')
    assert original['state']['limits'] is None
    rows=[]
    def emit(name, edit):
        value=copy.deepcopy(original);edit(value['state'])
        value['state_sha256']=digest(canonical(value['state']).encode())
        data=encoded_json(value).encode();envelope(data)
        rows.append(dict(name=name,bytes=base64.b64encode(data).decode(),expected_code='session_corrupt',state_hash_repaired=True))
    emit('set-at-dispatch-not-after',lambda s:s['context']['LimitFacts'][set_index].__setitem__('seq',calls['setter']['CalledAt']))
    emit('set-at-return-not-before',lambda s:s['context']['LimitFacts'][set_index].__setitem__('seq',calls['setter']['ReturnedAt']))
    emit('consumption-after-dispatch',lambda s:s['context']['LimitFacts'][consumed_index].__setitem__('seq',int(calls['next']['CalledAt'])+1))
    emit('dispatch-after-result',lambda s:s['context']['Calls']['next'].__setitem__('CalledAt',int(calls['next']['ReturnedAt'])+1))
    emit('result-at-dispatch',lambda s:s['context']['Calls']['next'].__setitem__('ReturnedAt',calls['next']['CalledAt']))
    emit('set-overrides-differ-from-arguments',lambda s:s['context']['LimitFacts'][set_index]['overrides'].__setitem__('max_output_bytes',81))
    emit('consumed-overrides-differ-from-prior',lambda s:s['context']['LimitFacts'][consumed_index]['overrides'].__setitem__('max_output_bytes',81))
    emit('consumed-name-differs-from-call',lambda s:s['context']['LimitFacts'][consumed_index].__setitem__('name','read_file'))
    emit('consumed-setting-resurrected',lambda s:s.__setitem__('limits',{'max_output_bytes':80,'ai_callback_pattern':''}))
    guidance=next(i for i,x in enumerate(context['Guidance']) if int(x['consumed_at'])>0)
    emit('guidance-consumed-at-non-send',lambda s:s['context']['Guidance'][guidance].__setitem__('consumed_at',calls['setter']['CalledAt']))
    emit('accepted-send-witness-duplicated',lambda s:s['context']['RequestSeqs'].insert(0,s['context']['RequestSeqs'][0]))
    return {'scope':__doc__,'cases':rows}


if __name__=='__main__':
    print(json.dumps(cases(sys.stdin.buffer.read())))
