"""Bound public event replay for native speech; never a new provider session."""
import argparse
import datetime
import json
import os
from pathlib import Path
import subprocess
from evidence import HERE,ROOT,digest,historical,preflight
p=argparse.ArgumentParser();p.add_argument('vendor',choices=['anthropic','openai','gemini']);p.add_argument('name');p.add_argument('log');p.add_argument('retained_revision');a=p.parse_args()
b=json.loads((HERE/'initial-binding.json').read_text());paths={k:v['path'] for k,v in b['executables'].items()};preflight(b,paths)
source=(HERE/a.log).resolve();relative=str(source.relative_to(ROOT));data=source.read_bytes()
assert digest(data)==digest(historical(a.retained_revision,relative)),'retained log differs from immutable source; no evidence modified'
run=HERE/a.name;run.mkdir();workspace=run/'workspace';workspace.mkdir()
command=[paths['replay'],a.vendor,str(source)]
receipt={'start':datetime.datetime.now(datetime.timezone.utc).isoformat(),'actor':'Codex student coder, not Bill','vendor':a.vendor,'mode':'retained-event-replay','command':command,'workspace':str(workspace),'source_revision':b['source_revision'],'executables':{k:v['sha256'] for k,v in b['executables'].items()},'executable_paths':paths,'support':b['support'],'browser_tools':b['browser_tools'],'launched_executable':'replay','delivery':'retained facts; no provider request','retained':{'source_revision':a.retained_revision,'path':relative,'sha256':digest(data)},'provider_endpoint':'http://127.0.0.1:1'}
(run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
code=subprocess.call([paths['recorder'],'-q',str(run/'terminal.txt'),*command],cwd=workspace,env=env)
receipt.update(exit_code=code,end=datetime.datetime.now(datetime.timezone.utc).isoformat());(run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
if code==0:
    admitted=(workspace/'readmitted.jsonl').read_bytes()
    assert [json.loads(x) for x in data.splitlines()]==[json.loads(x) for x in admitted.splitlines()],'readmitted facts changed'
    (run/'replay-match.json').write_text(json.dumps({'retained_sha256':digest(data),'readmitted_sha256':digest(admitted),'identical_events':True,'new_model_requests':0},indent=2)+'\n')
raise SystemExit(code)
