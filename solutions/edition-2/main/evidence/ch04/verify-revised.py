"""Check revised raw receipts against immutable source and both executables."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]
p = argparse.ArgumentParser()
p.add_argument('--cli-binary', type=Path, required=True)
p.add_argument('--public-binary', type=Path, required=True)
args = p.parse_args()
binding = json.loads((HERE/'boundary-binding.json').read_text())
for kind, binary in [('cli', args.cli_binary), ('public', args.public_binary)]:
    assert hashlib.sha256(binary.read_bytes()).hexdigest() == binding['binaries'][kind]
for path, want in binding['sources'].items():
    data = subprocess.check_output(['git','show',binding['source_revision']+':'+path],cwd=ROOT)
    assert hashlib.sha256(data).hexdigest() == want
# All launch identity checks precede any output file mutation.
launches = {}
for kind in ['live','public']:
    for vendor in ['anthropic','openai','gemini']:
        run = HERE/('boundary-'+kind+'-'+vendor)
        launch = json.loads((run/'launch.json').read_text())
        assert launch['source_revision'] == binding['source_revision']
        assert launch['binary_sha256'] == binding['binaries']['cli' if kind=='live' else 'public']
        assert launch['exit_code'] == 0
        launches[kind,vendor] = launch

def facts(path):
    rows = [json.loads(x) for x in path.read_text().splitlines()]
    assert rows[0] == {'log_version':1}
    events = rows[1:]
    assert all(e['seq']==i+1 for i,e in enumerate(events))
    usage = {key:sum(e['response']['usage'][key] for e in events if e['type']=='response_ended') for key in ['input','cache_write','cache_read','output']}
    identities = [json.loads(s) for s in sorted({json.dumps({key:e['response'].get(key) for key in ['from','requested','model_reported']},sort_keys=True) for e in events if e['type']=='response_ended'})]
    return events, usage, identities

def artifacts(run, events):
    out = {}
    for e in events:
        if e['type'] not in ['job_ended','job_killed']: continue
        j = e['job']; data=(run/'workspace'/j['output']['locator']).read_bytes()
        assert len(data)==j['bytes']
        assert str(j['handle']) not in out
        out[str(j['handle'])] = dict(snapshot=j, sha256=hashlib.sha256(data).hexdigest())
    return out

receipts = {'binding':binding['source_revision'],'runs':{}}
for vendor in ['anthropic','openai','gemini']:
    run=HERE/('boundary-live-'+vendor)
    events,usage,identities=facts(run/'session.log')
    calls={e['tool']['call_id']:e for e in events if e['type']=='tool_called'}
    results={e['tool']['call_id']:e for e in events if e['type']=='tool_returned'}
    assert calls.keys()==results.keys()
    stored=artifacts(run,events)
    assert len(stored)==sum('job' in e['tool'] for e in calls.values())
    source_caps=[]; notes=[]
    for ident,e in calls.items():
        call=e['tool'];result=results[ident]['tool'];text=result['parts'][0].get('text','')
        if call['name']=='read_file' and call['args'].get('path')=='notes.txt' and call['args'].get('max_bytes')==4 and not result.get('is_error'):
            handle=call['job']['handle'];data=(run/'workspace'/f'cr/io/{handle}').read_bytes()
            assert data==b'alph' and stored[str(handle)]['snapshot']['bytes']==4
            assert 'max_bytes limit reached' in text
            source_caps.append({'call_seq':e['seq'],'handle':handle,'report':text})
        if text.startswith('tool_limits consumed'): notes.append(text)
    if vendor=='gemini':
        for e in events:
            if e['type']!='response_ended': continue
            for i,part in enumerate(e['response']['parts']):
                if part['type']=='tool_call' and part['call_id'].startswith('call-'): assert part['call_id']==f"call-{e['seq']}-{i}"
    assert len(source_caps)>=2
    assert any('2 bytes omitted; total bytes: 4' in x['report'] for x in source_caps)
    assert any('consumed by kill_job' in x for x in notes)
    assert any('consumed by run_command' in x and 'max_output_bytes=40' in x and 'ai_callback_delay=1' in x for x in notes)
    exact=[e for ident,e in calls.items() if e['tool']['name']=='send_input' and e['tool']['args'].get('append_newline') is False and e['tool']['args']['input']=='beta\n' and 'REPLY:beta READY>' in results[ident]['tool']['parts'][0]['text']]
    assert exact
    debugger=next(e for e in calls.values() if e['tool']['name']=='run_command' and 'dlv exec' in e['tool']['args']['command'])
    handle=debugger['tool']['job']['handle'];output=(run/'workspace'/f'cr/io/{handle}').read_text()
    assert 'Breakpoint 1 set' in output and '[Breakpoint 1]' in output and re.search(r'print value\n42\n',output)
    assert stored[str(handle)]['snapshot']['status']=='done' and stored[str(handle)]['snapshot']['exit_code']==0
    assert any('LINE-050\nLINE-051\nLINE-052' in e['tool']['parts'][0].get('text','') for e in results.values())
    shutdown=[e for e in events if e['type']=='job_killed' and e['job'].get('reason')=='shutdown'];assert len(shutdown)==1
    before=json.loads((run/'process-before-eof.json').read_text());after=json.loads((run/'process-after-eof.json').read_text())
    assert before['shutdown']['alive'] and not after['shutdown']['alive'] and before['shutdown']['pid']==after['shutdown']['pid']
    transcript=(run/'terminal.txt').read_text()
    assert f"Final usage: input={usage['input']}, cache write={usage['cache_write']}, cache read={usage['cache_read']}, output={usage['output']}" in transcript
    receipts['runs'][vendor]={'launch':launches['live',vendor],'usage':usage,'identities':identities,'events':len(events),'tool_calls':len(calls),'source_caps':source_caps,'consumption_notes':notes,'debugger_handle':handle,'artifacts':stored,'shutdown':shutdown[0]}
    public=HERE/('boundary-public-'+vendor);summary=json.loads((public/'stdout.json').read_text())
    assert summary['foreign_handle']=='refused' and summary['close']=='idempotent'
    agents=[]
    for i in [1,2]:
        ev,u,ids=facts(public/f'agent-{i}.log'); agent=summary['agents'][i-1]
        assert u==agent['usage'] and agent['observed_sequences']==list(range(1,len(ev)+1))
        agents.append({'usage':u,'identities':ids,'artifacts':artifacts(public,ev),'events':len(ev)})
    assert summary['agents'][0]['handle']!=summary['agents'][1]['handle']
    receipts['runs'][vendor]['public']={'launch':launches['public',vendor],'summary':summary,'agents':agents}
(HERE/'boundary-receipts.json').write_text(json.dumps(receipts,indent=2)+'\n')
print(json.dumps({v:{'events':r['events'],'tool_calls':r['tool_calls'],'usage':r['usage'],'public':'passed'} for v,r in receipts['runs'].items()},indent=2))
