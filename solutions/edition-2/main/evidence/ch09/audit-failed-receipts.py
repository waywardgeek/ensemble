"""Explicit diagnostic audit of named failed attempts; never a success verifier."""
import argparse,json,subprocess
from pathlib import Path
from evidence import HERE,ROOT,PREFIX,preflight,check_launch,file_digest,digest,historical,scenario_catalog,scenario_inputs

FAILURES={
 'live-openai-p':(['responses/001.body'],'relay transport exception; no upstream response body'),
 'live-gemini-n':(['responses/008.body'],'relay transport exception after seven retained responses'),
 'live-gemini-n-recovery':([], 'response004 omits required candidatesTokenCount; complete stream refused'),
}

def prepare(binding):
 preflight(binding)
 prepared=[]
 for name,(missing,cause) in FAILURES.items():
  run=HERE/name;launch=json.loads((run/'launch.json').read_text());check_launch(binding,launch)
  assert launch['exit_code']==1,'expected preserved failed attempt'
  assert 0<launch['requests']<=launch['http_cap']
  req=[f'requests/{i:03}.json' for i in range(1,launch['requests']+1)]
  res=[f'responses/{i:03}.body' for i in range(1,launch['requests']+1)]
  assert sorted(str(p.relative_to(run)) for p in (run/'requests').glob('*.json'))==req
  assert sorted(str(p.relative_to(run)) for p in (run/'responses').glob('*.body'))==sorted(set(res)-set(missing))
  logs=['workspace/alpha/events.jsonl','workspace/beta/events.jsonl'] if launch['scenario']=='P' else ['session.log']
  assert launch['logs']==logs and launch['workspace']==str(run/'workspace')
  catalog={p:digest(b) for p,b in scenario_catalog(binding,launch['scenario']).items()}
  inputs=scenario_inputs(launch['scenario']);retained={'inputs/'+p:digest(b) for p,b in inputs.items()}
  assert launch['catalog_files']==catalog
  assert launch['scratch_inputs']=={'workspace/'+p:digest(b) for p,b in inputs.items()}
  outputs={str(p.relative_to(run)) for p in (run/'workspace').rglob('*') if p.is_file()}-set(logs)
  assert outputs==set(launch['scratch_outputs']) and outputs
  required=set(req+res+logs+['terminal.txt'])-set(missing)|set(catalog)|set(retained)|outputs
  assert required<=set(launch['originals'])
  assert not set(missing)&set(launch['originals'])
  for path,sha in {**catalog,**retained,**launch['scratch_outputs']}.items():assert launch['originals'][path]==sha
  for path,sha in launch['originals'].items():
   target=(run/path).resolve();assert target.is_relative_to(run)
   assert file_digest(target)==sha,'original receipt mismatch'
  events=[]
  for log in logs:
   values=[json.loads(line) for line in (run/log).read_text().splitlines()][1:]
   assert values
   events.extend((log,e) for e in values if e['type']=='request_sent')
  assert len(events)==launch['requests']
  prepared.append((run,launch,events,missing,cause))
 return prepared

if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--source',required=True);p.add_argument('--output',type=Path,required=True);args=p.parse_args()
 assert historical(args.source,PREFIX+'evidence/ch09/audit-failed-receipts.py')==Path(__file__).read_bytes(),'audit helper identity mismatch'
 binding=json.loads((HERE/'stream-binding.json').read_text());prepared=prepare(binding)
 # All original sets and historical/executable identities above pass before any replay or derived write.
 rows=[];derived=[]
 for run,launch,events,missing,cause in prepared:
  rendered=[]
  for log,event in events:
   body=subprocess.check_output([binding['executables']['cli']['path'],'replay',str(run/log),str(event['seq'])],cwd=run).removesuffix(b'\n')
   rendered.append(body);derived.append((f'{run.name}-{Path(log).parent.name}-{event["seq"]}.json',body+b'\n'))
  assert sorted(rendered)==sorted(p.read_bytes() for p in (run/'requests').glob('*.json')),'failed-attempt request byte mismatch'
  assert all(file_digest(run/path)==sha for path,sha in launch['originals'].items()),'original changed during replay'
  rows.append(dict(run=run.name,exit_code=1,status='failed attempt preserved; request reconstruction only',requests=launch['requests'],reconstructed=len(rendered),missing_response_bodies=missing,cause=cause,launch_sha256=file_digest(run/'launch.json')))
 assert not args.output.exists();args.output.mkdir()
 for name,body in derived:(args.output/name).write_bytes(body)
 report=dict(helper_revision=args.source,helper_sha256=file_digest(__file__),binding_sha256=file_digest(HERE/'stream-binding.json'),cli=binding['executables']['cli'],runs=rows,new_model_requests=0,comparison='exact bytes, excluding CLI final output LF; no success relabeling or response fabrication')
 (args.output/'audit.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))
