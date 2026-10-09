"""Offline phase6 observations; bound originals remain separate and unchanged."""
import collections
import json
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE / 'support'))
from identity import preflight, file_hash

repo = HERE.parents[4]
binding = json.loads((HERE / 'provider-prep-binding-final.json').read_text())
preflight(binding, repo)
base = HERE / 'live-20261008'

def read(p):
    return json.loads(p.read_text())

def rows(p):
    return [json.loads(s) for s in p.read_text().splitlines()]

def completion(p):
    v = read(p)
    return v.get('completion', v.get('Completion'))

result = {'source_revision': binding['source_revision'],
          'support_revision': binding['support_revision'],
          'binding_sha256': file_hash(HERE / 'provider-prep-binding-final.json'),
          'provider_results': {}, 'chapter_accepted': False}
for vendor in ('anthropic', 'openai', 'gemini'):
    root = base / vendor
    budget = rows(root / 'attempts.jsonl')
    generation = [r for r in budget if r['kind'] == 'generation']
    counts = dict(collections.Counter(r['row'] for r in generation))
    assert counts == {'A': 3, 'B': 1, 'C': 3, 'D': 4}
    assert sum(r['kind'] == 'discovery' for r in budget) == 1
    assert all(r['at'] <= r['row_deadline'] for r in generation)
    transport = [r for p in base.glob(vendor + '-*/transport.jsonl') for r in rows(p)]
    finished = [r for r in transport if r['phase'] == 'finished']
    assert len(finished) == 11
    assert all(r['status'] == 200 and r['outcome'] == 'completed' and
               not r['credential_redacted'] and r['elapsed'] < 120 for r in finished)
    a1 = read(base / f'{vendor}-A1/closed-store/checkpoint.json')
    a2 = read(base / f'{vendor}-A2/closed-store/checkpoint.json')
    assert a1['session_id'] == a2['session_id'] and a1['as_of'] == 11 and a2['as_of'] == 16
    history = rows(base / f'{vendor}-A2/closed-store/events.log')
    assert not any(e.get('type') == 'tool_called' and e['seq'] > 11 for e in history)
    marker = (root / 'workspace/marker.txt').read_bytes()
    assert marker == f'CH10-A-{vendor}\n'.encode() == (base / f'{vendor}-A1/marker.txt').read_bytes()
    c = base / f'{vendor}-C/public'
    old, new = read(c/'C-old.json'), read(c/'C-new.json')
    assert old['as_of'] < new['as_of']
    c1, c2, c3 = [completion(c/n) for n in ('C1-completion.json','C2-completion.json','C3-completion.json')]
    assert all(x['outcome'] == 'success' for x in (c1,c2,c3))
    assert f'CH10-C-{vendor}' in c1['text'] and c2['text'] == f'FRESH-{vendor}' and f'CH10-C-{vendor}' in c3['text']
    origin = read(c/'origin-result.json')
    assert origin['unchanged'] and origin['usage_restored_once'] and origin['pre_origin_error'] == 'history_unavailable'
    offline = read(base/f'{vendor}-offline/public/offline-result.json')
    assert offline['equal_render_context_history_usage_skills_watch_requests'] and offline['provider_calls'] == 0
    # Compare public reconstruction to the actual captured bytes, not just to
    # another reconstruction. C2/C3 live originals are retained independently.
    exact = []
    for seq, step, request in ((4,'A1',2),(9,'A1',3),(14,'A2',4),(19,'B',5),
                              (28,'C',next(r['number'] for r in generation if r['step']=='C1'))):
        captured = base/f'{vendor}-{step}/{request:03d}-request.json'
        replay = base/f'{vendor}-offline/public/latest-request-{seq}.json'
        assert captured.read_bytes() == replay.read_bytes()
        exact.append({'seq':seq,'captured':str(captured.relative_to(base)),
                      'reconstructed':str(replay.relative_to(base)), 'sha256':file_hash(captured)})
    policy = read(root/'current-policy.json')
    prefs = read(root/'preferences.json')
    assert policy['policy']['max_model_requests'] == 1 and prefs['preferences']['font_size'] == 18
    turns = [e['turn']['policy'] for e in rows(c/'C-events.log') if e.get('type')=='turn_started']
    assert [(t['max_model_requests'],t['effective_max_model_requests']) for t in turns] == [(0,16),(0,16),(1,1),(2,2)]
    speech = []
    for run in (vendor+'-B',vendor+'-B-restart'):
        for event in rows(base/run/'browser-original.jsonl'):
            if event['kind'].endswith('_speech'):
                # Lifecycle records deliberately retain before/after separately;
                # inspect the restored/closed side, never erase the original.
                state=event.get('state',event.get('after',event.get('closed')))
                assert state is not None
                assert state['nativeAdmissions']==0 and state['serviceAdmissions']==0
                assert state['servicePending']==0 and not state['serviceActive']
                assert all(p['queue']==0 and not p['current'] and not p['causes']['speaking'] and p['partials']==0 for p in state['pages'])
                speech.append({'run':run,'kind':event['kind']})
    d = read(root/'D/checkpoint.json')
    before = read(base/f'{vendor}-D2/public/D2-before.json')
    prepared = read(base/f'{vendor}-D-prepare/public/D-prepared.json')
    assert before['state']['skills'] == prepared['state']['skills'] == d['state']['skills']['state']
    skill = d['state']['skills']
    assert [m['event_seq'] for m in skill['material']] == [2,3,5]
    assert [m['retired'] for m in skill['material']] == [False,True,False]
    consumed = read(base/f'{vendor}-D2/public/D2-consumption.json')
    assert consumed['observed'] and consumed['count']==1
    limits = [e for e in rows(root/'D/events.log') if e.get('type') in ('tool_limits_set','tool_limits_consumed')]
    assert len(limits)==2 and limits[0]['limits']['overrides']==limits[1]['limits']['overrides']=={'ai_callback_pattern':'','max_output_bytes':17}
    assert d['state']['limits'] is None
    assert (root/'workspace/after-resume.txt').read_bytes() == f'CH10-D-{vendor}\n'.encode()
    seed = read(base/f'{vendor}-D-seed/public/seed-result.json')
    assert seed['requests']==2 and not seed['error']
    result['provider_results'][vendor] = dict(
        selected_model=read(base/f'{vendor}-A1/launch.json')['model'],
        generation_attempts=11, discovery_attempts=1, rows=counts,
        all_transport_200=True, session_id=a1['session_id'],
        original_checkpoint=old['as_of'], tail_checkpoint=new['as_of'],
        public_outcomes=[x['outcome'] for x in (c1,c2,c3)],
        exact_request_comparisons=exact, origin=origin, offline=offline,
        speech_boundaries=speech, policies=turns, current_policy=policy,
        preferences=prefs, skills=skill['state'], limits_consumption=consumed,
        synthetic_seed={'requests':2,'provider_requests':0,'input_tokens':2,'output_tokens':2},
        d_cumulative_usage_including_labeled_seed=d['state']['usage'])
with (base/'observations.json').open('x') as f:
    json.dump(result,f,indent=2);f.write('\n')
print('Three vendors: 33 generation / 3 discovery; 15 byte-exact captured/replayed request pairs; original observations verified.')
