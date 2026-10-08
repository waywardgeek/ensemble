"""Preflight every identity, then compare captured raw requests to exact replay."""
import argparse
from collections import Counter
import json
import os
from pathlib import Path
import subprocess

from evidence import HERE, digest, preflight

def verify(binding,paths):
    preflight(binding,paths)
    expected={'controls-'+v for v in ('anthropic','openai','gemini')} | {'eof-'+v for v in ('anthropic','openai','gemini')} | {'workflow-'+v for v in ('anthropic','openai','gemini')} | {'collection-'+v for v in ('anthropic','openai','gemini')}
    assert set(binding['runs'])==expected, 'incomplete planned run set; no evidence modified'
    launches={}
    # Entire launch set is validated before constructing any derived destination.
    for name in binding['runs']:
        launch=json.loads((HERE/name/'launch.json').read_text())
        assert launch['source_revision']==binding['source_revision'], 'launch source mismatch; no evidence modified'
        assert launch['executables']=={n:v['sha256'] for n,v in binding['executables'].items()}, 'launch executable mismatch; no evidence modified'
        assert launch['support']==binding['support'], 'launch support mismatch; no evidence modified'
        assert launch['exit_code']==0, 'unsuccessful launch remains unvalidated'
        executable='cli' if launch['mode']=='chat' else 'workflow'
        assert launch['launched_executable']==executable and launch['command']==[launch['executable_paths'][executable],launch['mode']], 'launch command identity mismatch'
        assert set(launch['executable_paths'])==set(binding['executables']), 'incomplete launch executable paths'
        launches[name]=launch
    env=os.environ.copy()
    for key in list(env):
        if key.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_')):del env[key]
    rows={};derivatives={}
    for name,launch in launches.items():
        run=HERE/name;raw={p.name:p.read_bytes() for p in sorted((run/'requests').glob('*.json'))}
        assert raw and len(raw)==launch['requests'], 'raw request set incomplete'
        remaining=Counter(raw.values());requests=[];terminal={};calls=results=0;usage={k:0 for k in ('input','cache_write','cache_read','output')}
        for log in sorted(run.glob('*.log')):
            records=[json.loads(line) for line in log.read_text().splitlines()];assert records[0]=={'log_version':1}
            events=records[1:];assert all(e['seq']==i+1 for i,e in enumerate(events))
            called={e['tool']['call_id'] for e in events if e['type']=='tool_called'};returned=[e['tool']['call_id'] for e in events if e['type']=='tool_returned'];assert called==set(returned) and len(returned)==len(set(returned));calls+=len(called);results+=len(returned)
            ended=[e['job']['handle'] for e in events if e['type'] in ('job_ended','job_killed')];assert len(ended)==len(set(ended))
            for e in events:
                if e['type']=='response_ended':
                    for k in usage:usage[k]+=e['response']['usage'][k]
                if e['type'] in ('job_ended','job_killed'):
                    job=e['job'];data=(run/'workspace'/job['output']['locator']).read_bytes();assert len(data)==job['bytes'];terminal[str(job['handle'])]={'status':job['status'],'reason':job.get('reason'),'sha256':digest(data),'bytes':len(data)}
                if e['type']!='request_sent':continue
                output=subprocess.check_output([str(paths['cli']),'replay',str(log),str(e['seq'])],cwd=run/'workspace',env=env).removesuffix(b'\n')
                assert remaining[output]>0, f'{name}/{log.name}:{e["seq"]} replay differs from captured request'
                remaining[output]-=1;requests.append({'log':log.name,'seq':e['seq'],'sha256':digest(output),'hints':e['request']['hints']});derivatives[run/f'replay-{log.stem}-{e["seq"]}.json']=output
        assert not any(remaining.values()), 'captured request has no exact recorded prefix'
        transcript=(run/'terminal.txt').read_bytes();assert transcript
        if launch['mode']=='chat':assert b'Final usage:' in transcript
        if name.startswith('controls-'):
            for marker in (b'Hint received',b'interrupted=true',b'Command refused: Agent busy',b'/history',b'/usage'):assert marker in transcript,marker
            assert any(r['hints'] for r in requests);assert any(j['reason']=='shutdown' for j in terminal.values());assert any(j['reason']=='kill_job' for j in terminal.values())
        if name.startswith('workflow-'):
            assert all((run/(role+'-completion.json')).is_file() for role in ('author','editor','reviewer'));draft=(run/'workspace/draft.txt').read_text().lower();assert all(x in draft for x in ('saturday','10','bring a broken lamp','repairs are free'))
        if name.startswith('collection-'):
            c=json.loads((run/'collection-completion.json').read_text());assert len(c['results'])==len(c['independent_results'])==2 and c['canceled']['outcome']=='canceled' and c['exhausted']
        rows[name]={'requests':requests,'raw_requests':{n:digest(b) for n,b in raw.items()},'terminal_sha256':digest(transcript),'calls':calls,'results':results,'jobs':terminal,'usage':usage}
    # All identities and positive comparisons have passed. Only now write derivatives.
    for path,data in derivatives.items():path.write_bytes(data)
    (HERE/'receipts.json').write_text(json.dumps({'source_revision':binding['source_revision'],'executables':binding['executables'],'runs':rows},indent=2)+'\n')
    return rows

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--cli',type=Path,required=True);parser.add_argument('--workflow',type=Path,required=True);args=parser.parse_args()
    binding=json.loads((HERE/'initial-binding.json').read_text());rows=verify(binding,{'cli':args.cli.resolve(),'workflow':args.workflow.resolve()});print(json.dumps({name:{'requests':len(row['requests']),'usage':row['usage']} for name,row in rows.items()},indent=2))
