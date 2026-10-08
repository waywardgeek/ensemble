"""Bound human PTY launcher and streaming capture relay; no automatic retry."""
import argparse
import datetime
import fcntl
import http.server
import json
import os
from pathlib import Path
import subprocess
import threading
import urllib.error
import urllib.request
from evidence import HERE, CAPS, preflight, launch_identity, file_digest, digest, scenario_catalog, scenario_inputs

FIELDS={'anthropic':'directClaudeAPIKey','openai':'directOpenAIAPIKey','gemini':'directGeminiAPIKey'}
ORIGINS={'anthropic':'https://api.anthropic.com','openai':'https://api.openai.com','gemini':'https://generativelanguage.googleapis.com'}

def credential(vendor):
    return json.loads((Path.home()/'.cr/settings.json').read_text())[FIELDS[vendor]]

def headers(vendor,key):
    return {'anthropic':{'x-api-key':key,'anthropic-version':'2023-06-01'},'openai':{'Authorization':'Bearer '+key},'gemini':{'x-goog-api-key':key}}[vendor]

def discover(vendor,output):
    path={'anthropic':'/v1/models?limit=100','openai':'/v1/models','gemini':'/v1beta/models?pageSize=1000'}[vendor]
    with urllib.request.urlopen(urllib.request.Request(ORIGINS[vendor]+path,headers=headers(vendor,credential(vendor))),timeout=30) as response:page=json.load(response)
    names=[m.get('id',m.get('name')) for m in page.get('data',page.get('models',[])) if vendor!='gemini' or 'generateContent' in m.get('supportedGenerationMethods',[])]
    result=dict(at=datetime.datetime.now(datetime.timezone.utc).isoformat(),vendor=vendor,models=names)
    output.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))

class Relay(http.server.ThreadingHTTPServer):
    def __init__(self,vendor,key,run,scenario,budget,origin=None):
        super().__init__(('127.0.0.1',0),Handler)
        self.vendor,self.key,self.run,self.scenario,self.budget=vendor,key,run,scenario,budget
        self.origin=origin or ORIGINS[vendor]
        self.lock=threading.Lock();self.number=0

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self,*args):pass
    def do_POST(self):
        owner=self.server
        with owner.budget.open('a+') as counter:
            fcntl.flock(counter,fcntl.LOCK_EX);counter.seek(0);raw=counter.read();used=json.loads(raw) if raw else {name:0 for name in CAPS}
            allowed=used[owner.scenario]<CAPS[owner.scenario] and sum(used.values())<36
            if allowed:
                used[owner.scenario]+=1;counter.seek(0);counter.truncate();counter.write(json.dumps(used));counter.flush()
            fcntl.flock(counter,fcntl.LOCK_UN)
        if not allowed:self.send_error(429,'bounded demonstration limit');return
        with owner.lock:owner.number+=1;number=owner.number
        body=self.rfile.read(int(self.headers['content-length']))
        (owner.run/'requests'/f'{number:03}.json').write_bytes(body)
        request=urllib.request.Request(owner.origin+self.path,data=body,headers={**headers(owner.vendor,owner.key),'content-type':'application/json'},method='POST')
        try:response=urllib.request.urlopen(request,timeout=60)
        except urllib.error.HTTPError as error:response=error
        except Exception:self.send_error(502,'relay transport failed');return
        retained=bytearray()
        try:
            self.send_response(response.status);self.send_header('content-type',response.headers.get('content-type','application/json'));self.send_header('connection','close');self.end_headers()
            with response:
                while True:
                    chunk=response.read1(4096)
                    if not chunk:break
                    retained.extend(chunk);self.wfile.write(chunk);self.wfile.flush()
        except (BrokenPipeError,ConnectionResetError):pass
        finally:
            (owner.run/'responses'/f'{number:03}.body').write_bytes(bytes(retained).replace(owner.key.encode(),b'[redacted credential]'))
            self.close_connection=True

def launch(args):
    binding=json.loads(args.binding.read_text());paths={n:v['path'] for n,v in binding['executables'].items()}
    if args.cli:paths['cli']=args.cli
    preflight(binding,paths)
    discovery=json.loads(args.discovery.read_text());assert discovery['vendor']==args.vendor and args.model in discovery['models'],'model is not in recorded discovery'
    if args.vendor=='gemini':assert args.model=='models/gemini-3.8-flash','unexpected Gemini demonstration target'
    run=args.run.resolve();run.mkdir();workspace=run/'workspace';workspace.mkdir();(run/'requests').mkdir();(run/'responses').mkdir();catalog=run/'catalog';catalog.mkdir()
    for path,data in scenario_catalog(binding,args.scenario).items():
        target=run/path;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
    for path,data in scenario_inputs(args.scenario).items():
        for base in (workspace,run/'inputs'):
            target=base/path;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
    key=credential(args.vendor)
    relay=Relay(args.vendor,key,run,args.scenario,HERE/('budget-'+args.vendor+'.json'));thread=threading.Thread(target=relay.serve_forever);thread.start()
    env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
    env.update(LLM_VENDOR=args.vendor,LLM_MODEL=args.model,LLM_RESOLVED_MODEL=args.model.removeprefix('models/'),LLM_API_KEY=key,LLM_BASE_URL=f'http://127.0.0.1:{relay.server_port}',LLM_SKILLS_DIR=str(catalog),LLM_PRIMARY_SKILL='ensemble' if args.scenario=='F' else 'base',LLM_SYSTEM='',CH02_LOG=str(run/'session.log'),EN_DISABLE_STREAMING='0')
    executable={'N':'cli','F':'cli','G':'gui','P':'consumer'}[args.scenario];command=[paths[executable]]
    if executable=='cli':command+=['chat']
    if executable=='gui':command+=['--port','0','--terminal','--gui-log',str(run/'gui.jsonl')]
    if executable=='consumer':command+=['--ask']
    receipt={**launch_identity(binding,executable,args.scenario),'start':datetime.datetime.now(datetime.timezone.utc).isoformat(),'actor':'Codex student coder, not Bill','vendor':args.vendor,'requested_model':args.model,'resolved_model':args.model.removeprefix('models/'),'discovery_sha256':file_digest(args.discovery),'command':command,'workspace':str(workspace),'catalog_files':{str(p.relative_to(run)):file_digest(p) for p in catalog.rglob('SKILL.md')},'scratch_inputs':{str(p.relative_to(run)):file_digest(p) for p in workspace.rglob('notes.txt')},'delivery':'stream','executable_paths':paths,'exit_code':None,'requests':0,'logs':[],'originals':{}}
    (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
    code=-1
    try:code=subprocess.call([paths['recorder'],'-q',str(run/'terminal.txt'),*command],cwd=workspace,env=env)
    finally:
        relay.shutdown();relay.server_close();thread.join()
        logs=[run/'session.log'] if executable!='consumer' else [workspace/n/'events.jsonl' for n in ('alpha','beta')]
        receipt.update(exit_code=code,end=datetime.datetime.now(datetime.timezone.utc).isoformat(),requests=relay.number,logs=[str(p.relative_to(run)) for p in logs if p.exists()])
        receipt['scratch_outputs']={str(p.relative_to(run)):file_digest(p) for p in workspace.rglob('*') if p.is_file() and p not in logs}
        originals=[*run.glob('requests/*.json'),*run.glob('responses/*.body'),*logs,*catalog.rglob('SKILL.md'),run/'terminal.txt',*(run/'inputs').rglob('notes.txt'),*(run/p for p in receipt['scratch_outputs'])]
        if executable=='gui':
            originals += [run/'gui.jsonl',run/'browser-binding.json']
            browser_binding=run/'browser-binding.json'
            if browser_binding.exists():originals += [run/p for p in json.loads(browser_binding.read_text())['originals']]
        receipt['originals']={str(p.relative_to(run)):file_digest(p) for p in originals if p.exists()}
        (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
    raise SystemExit(code)

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('vendor',choices=FIELDS);p.add_argument('--discover',type=Path);p.add_argument('--model');p.add_argument('--scenario',choices=CAPS);p.add_argument('--run',type=Path);p.add_argument('--binding',type=Path,default=HERE/'initial-binding.json');p.add_argument('--discovery',type=Path);p.add_argument('--cli');args=p.parse_args()
    if args.discover:discover(args.vendor,args.discover)
    else:
        assert args.model and args.scenario and args.run and args.discovery,'model, scenario, run and discovery are required'
        launch(args)
