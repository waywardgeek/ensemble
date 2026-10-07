"""Verify historical source/binary binding before producing replay derivatives."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]
parser = argparse.ArgumentParser()
parser.add_argument('--binary',type=Path,required=True)
args = parser.parse_args()
binding = json.loads((HERE/'initial-binding.json').read_text())
assert hashlib.sha256(args.binary.read_bytes()).hexdigest()==binding['binary_sha256'], 'executable mismatch; no evidence modified'
for path,want in binding['sources'].items():
    source = subprocess.check_output(['git','show',binding['source_revision']+':'+path],cwd=ROOT)
    assert hashlib.sha256(source).hexdigest()==want, 'historical source mismatch; no evidence modified'

# All source/executable validation above precedes writes and subprocess replay.
receipts = {}
for vendor in ['anthropic','openai','gemini']:
    run = HERE/('live-'+vendor)
    launch = json.loads((run/'launch.json').read_text())
    assert launch['binary_sha256']==binding['binary_sha256'] and launch['source_revision']==binding['source_revision']
    records = [json.loads(x) for x in (run/'session.log').read_text().splitlines()]
    assert records[0]=={'log_version':1}
    events = records[1:]
    calls = {e['tool']['call_id']:e for e in events if e['type']=='tool_called'}
    results = {e['tool']['call_id']:e for e in events if e['type']=='tool_returned'}
    assert calls.keys()==results.keys()
    assert all(e['seq']==i+1 for i,e in enumerate(events))
    ordinary = [e for e in calls.values() if e['tool']['name'] in ['read_file','run_command']]
    assert all(e['tool']['job']['status']=='running' for e in ordinary)
    terminal = {e['job']['handle']:e for e in events if e['type'] in ['job_ended','job_killed']}
    assert len(terminal)==len(ordinary)
    assert len([e for e in events if e['type'] in ['job_ended','job_killed']])==len(terminal)
    artifacts = {}
    for handle,e in terminal.items():
        job=e['job'];data=(run/'workspace'/job['output']['locator']).read_bytes()
        assert len(data)==job['bytes']
        if job['status']=='killed': assert 'exit_code' not in job
        artifacts[str(handle)]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'status':job['status'],'locator':job['output']['locator']}
    debugger = next(e for e in ordinary if e['tool']['name']=='run_command' and 'dlv exec' in e['tool']['args']['command'])
    debug_handle=debugger['tool']['job']['handle']
    debugger_output=(run/'workspace'/f'cr/io/{debug_handle}').read_text()
    assert 'Breakpoint 1 set' in debugger_output and '[Breakpoint 1]' in debugger_output
    assert re.search(r'print value\n42\n',debugger_output)
    assert terminal[debug_handle]['job']['status']=='done'
    assert any(e['tool']['name']=='send_input' and e['tool']['args'].get('append_newline') is False and e['tool']['args']['input']=='beta\n' and 'REPLY:beta READY>' in results[ident]['tool']['parts'][0]['text'] for ident,e in calls.items())
    recovery=[{'call_seq':e['seq'],'result_seq':results[ident]['seq'],'path':e['tool']['args']['path']} for ident,e in calls.items() if e['tool']['name']=='read_file' and e['tool']['args'].get('start_line')==50]
    assert recovery
    assert any('LINE-050\nLINE-051\nLINE-052' in e['tool']['parts'][0].get('text','') for e in results.values())
    consumption=[{'seq':e['seq'],'text':e['tool']['parts'][0]['text']} for e in results.values() if e['tool']['parts'][0].get('text','').startswith('tool_limits consumed')]
    assert any('consumed by kill_job' in e['text'] for e in consumption)
    assert any('consumed by read_file' in e['text'] for e in consumption)
    assert any('consumed by run_command' in e['text'] and 'ai_callback_delay=1' in e['text'] and 'max_output_bytes=100' in e['text'] for e in consumption)
    assert any(e['type']=='redacted' for e in events)
    shutdown=[e for e in terminal.values() if e['job'].get('reason')=='shutdown']
    assert len(shutdown)==1
    before=json.loads((run/'process-before-eof.json').read_text());after=json.loads((run/'process-after-eof.json').read_text())
    assert before['shutdown'].get('alive',before['shutdown'].get('alive_before_eof'))
    assert not any(v['alive'] for v in after.values())
    usage={key:sum(e['response']['usage'][key] for e in events if e['type']=='response_ended') for key in ['input','cache_write','cache_read','output']}
    transcript=(run/'terminal.txt').read_text()
    assert '/history' in transcript and 'Recorded redact.' in transcript and 'Final usage:' in transcript
    assert f"Final usage: input={usage['input']}, cache write={usage['cache_write']}, cache read={usage['cache_read']}, output={usage['output']}" in transcript
    env=os.environ.copy()
    for key in list(env):
        if key.startswith(('LLM_','OPENAI_','ANTHROPIC_','GEMINI_')): del env[key]
    env.update(LLM_VENDOR=vendor,LLM_MODEL=launch['requested_model'],LLM_RESOLVED_MODEL=launch['requested_model'].removeprefix('models/'),CH02_LOG=str(run/'session.log'))
    dump=subprocess.check_output([str(args.binary),'dump'],env=env,cwd=run/'workspace')
    render1=subprocess.check_output([str(args.binary),'render',str(run/'session.log')],env=env,cwd=run/'workspace')
    render2=subprocess.check_output([str(args.binary),'render',str(run/'session.log')],env=env,cwd=run/'workspace')
    assert render1==render2
    (run/'offline-dump.jsonl').write_bytes(dump)
    (run/'offline-render.json').write_bytes(render1)
    identities=list({json.dumps({'from':e['response']['from'],'requested':e['response'].get('requested'),'model_reported':e['response'].get('model_reported')},sort_keys=True) for e in events if e['type']=='response_ended'})
    receipts[vendor]={'launch':str(run/'launch.json'),'events':len(events),'tool_calls':len(calls),'jobs':len(ordinary),'debugger_handle':debug_handle,'debugger_artifact':f'workspace/cr/io/{debug_handle}','recovery_calls':recovery,'consumption_notes':consumption,'usage':usage,'identities':[json.loads(i) for i in identities],'shutdown_event':shutdown[0],'artifacts':artifacts,'offline_replay':'two deterministic renders, no API credentials','observed_model_misstatements':'See student review; tool records are authoritative.'}
(HERE/'receipts.json').write_text(json.dumps({'binding':binding['source_revision'],'binary_sha256':binding['binary_sha256'],'runs':receipts},indent=2)+'\n')
print(json.dumps({v:{k:r[k] for k in ['events','tool_calls','jobs','debugger_handle','usage']} for v,r in receipts.items()},indent=2))
