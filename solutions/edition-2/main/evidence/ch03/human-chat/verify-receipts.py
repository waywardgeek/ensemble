"""Offline evidence checks; never supplies a model response or a human prompt."""
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

HERE = Path(__file__).resolve().parent
BINARY = '/tmp/ensemble-ed2-ch03-human-chat'


def objects(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from objects(child)
    elif isinstance(value, list):
        for child in value:
            yield from objects(child)


def check(run):
    launch = json.loads((run / 'launch.json').read_text())
    records = [json.loads(line) for line in (run / 'session.log').read_text().splitlines()]
    events = records[1:]
    calls = {e['tool']['call_id']: e for e in events if e['type'] == 'tool_called'}
    results = [e for e in events if e['type'] == 'tool_returned']
    assert set(calls) == {e['tool']['call_id'] for e in results}
    assert {e['tool']['name'] for e in calls.values()} == {'list_directory','read_file','search_files','write_file','edit_file','run_command'}
    target = next(e for e in results if e['seq'] == 5)
    target_id = target['tool']['call_id']
    terminal = (run / 'terminal.txt').read_text()
    assert f'5 tool_returned call_id={target_id}' in terminal
    assert '/redact 5 5 demonstration' in terminal
    assert launch['exit_code'] == 0
    assert (run / 'workspace/created.txt').read_text() == 'replacement\ntail\n'
    assert (run / 'workspace/anchors.txt').read_text() == 'first repaired\nsecond anchor\n'
    before = {str(p.relative_to(run)): hashlib.sha256(p.read_bytes()).hexdigest() for p in [run/'session.log', *sorted((run/'workspace').iterdir())] if p.is_file()}
    env = {k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
    env.update(LLM_VENDOR=launch['vendor'], LLM_MODEL=launch['selected_model'], LLM_RESOLVED_MODEL=launch['selected_model'].removeprefix('models/'), CH02_LOG=str(run/'session.log'))
    redaction_seq = next(e['seq'] for e in events if e['type']=='redacted')
    request_checks=[]
    ephemeral=[]
    for event in events:
        if event['type'] != 'request_sent':
            continue
        seq=event['seq']
        prefix=run/f'prefix-before-{seq}.log'
        prefix.write_text('\n'.join(json.dumps(e,separators=(',',':')) for e in records if 'seq' not in e or e['seq']<seq)+'\n')
        a=subprocess.run([BINARY,'render',str(prefix)],env=env,capture_output=True,check=True)
        b=subprocess.run([BINARY,'render',str(prefix)],env=env,capture_output=True,check=True)
        assert a.stdout==b.stdout
        (run/f'request-before-{seq}.json').write_bytes(a.stdout)
        body=json.loads(a.stdout)
        ephemeral.append({'request_seq':seq,'consumed':event['request']['ephemera'],'marker_in_projection':b'LILAC-614' in a.stdout})
        if seq>redaction_seq:
            target_objects=[o for o in objects(body) if (o.get('type')=='tool_result' and o.get('tool_use_id')==target_id) or (o.get('role')=='tool' and o.get('tool_call_id')==target_id) or ('response' in o and o.get('id')==target_id)]
            assert len(target_objects)==1, (seq,target_objects)
            assert '[redacted]' in json.dumps(target_objects[0]), (seq,target_objects)
            assert 'notes.md' not in json.dumps(target_objects[0])
            request_checks.append(seq)
    assert request_checks
    assert sum(x['marker_in_projection'] for x in ephemeral)==1
    assert sum(bool(x['consumed']) for x in ephemeral)==1
    dump=subprocess.run([BINARY,'dump'],env=env,capture_output=True,check=True)
    (run/'offline-dump.jsonl').write_bytes(dump.stdout)
    assert [json.loads(x) for x in dump.stdout.splitlines()]==records
    after={p:hashlib.sha256((run/p).read_bytes()).hexdigest() for p in before}
    assert before==after
    usage={k:sum(e['response']['usage'][k] for e in events if e['type']=='response_ended') for k in ['input','cache_write','cache_read','output']}
    final=f"Final usage: input={usage['input']}, cache write={usage['cache_write']}, cache read={usage['cache_read']}, output={usage['output']}"
    assert final in terminal
    errors=[{'seq':e['seq'],'name':calls[e['tool']['call_id']]['tool']['name'],'args':calls[e['tool']['call_id']]['tool']['args'],'parts':e['tool']['parts']} for e in results if e['tool'].get('is_error')]
    assert {e['name'] for e in errors}>={'write_file','edit_file','read_file','search_files'}
    return {'vendor':launch['vendor'],'model':launch['selected_model'],'returned_models':sorted({e['response']['from']['model'] for e in events if e['type']=='response_ended'}),'human_turns':sum(e['type']=='message_received' and e['message']['actor']=='human' for e in events),'requests':len(ephemeral),'tool_calls':len(calls),'usage':usage,'redaction_target':5,'redacted_requests_verified':request_checks,'ephemeral':ephemeral,'errors':errors,'files_unchanged_by_replay':before,'exit_code':launch['exit_code'],'terminal_sha256':hashlib.sha256((run/'terminal.txt').read_bytes()).hexdigest()}

receipts=[check(HERE/(v+'-main')) for v in ['anthropic','openai','gemini']]
(HERE/'receipts.json').write_text(json.dumps(receipts,indent=2)+'\n')
print(json.dumps([{k:v for k,v in x.items() if k in ['vendor','model','usage','human_turns','requests','tool_calls','exit_code']} for x in receipts],indent=2))
