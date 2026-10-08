"""Human-driven PTY launcher; reuse the predecessor credential-safe script path.
Raw request bodies are captured by a local relay, independently of event replay.
No authorization header, settings object, or secret is printed or retained.
"""
import argparse
import datetime
import http.server
import json
import os
from pathlib import Path
import subprocess
import threading
import urllib.error
import urllib.request

from evidence import HERE, digest, preflight
FIELDS = {'anthropic':'directClaudeAPIKey','openai':'directOpenAIAPIKey','gemini':'directGeminiAPIKey'}
ORIGINS = {'anthropic':'https://api.anthropic.com','openai':'https://api.openai.com','gemini':'https://generativelanguage.googleapis.com'}

def credential(vendor):
    return json.loads((Path.home()/'.cr/settings.json').read_text())[FIELDS[vendor]]

def headers(vendor,key):
    return {'anthropic':{'x-api-key':key,'anthropic-version':'2023-06-01'},'openai':{'Authorization':'Bearer '+key},'gemini':{'x-goog-api-key':key}}[vendor]

def discover():
    result={'at':datetime.datetime.now(datetime.timezone.utc).isoformat(),'providers':{}}
    for vendor in FIELDS:
        path={'anthropic':'/v1/models?limit=100','openai':'/v1/models','gemini':'/v1beta/models?pageSize=1000'}[vendor]
        with urllib.request.urlopen(urllib.request.Request(ORIGINS[vendor]+path,headers=headers(vendor,credential(vendor))),timeout=30) as response:
            page=json.load(response)
        result['providers'][vendor]=[m.get('id',m.get('name')) for m in page.get('data',page.get('models',[])) if vendor!='gemini' or 'generateContent' in m.get('supportedGenerationMethods',[])]
    (HERE/'discovery.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2))

class Relay(http.server.ThreadingHTTPServer):
    def __init__(self,vendor,key,run):
        super().__init__(('127.0.0.1',0),Handler)
        self.vendor,self.key,self.run=vendor,key,run
        self.lock=threading.Lock();self.number=0

class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self,*args): pass
    def do_POST(self):
        owner=self.server
        with owner.lock:
            owner.number+=1;number=owner.number
        if number>60:
            self.send_error(429,'bounded demonstration limit');return
        body=self.rfile.read(int(self.headers['content-length']))
        (owner.run/'requests'/f'{number:03}.json').write_bytes(body)
        request=urllib.request.Request(ORIGINS[owner.vendor]+self.path,data=body,headers={**headers(owner.vendor,owner.key),'content-type':'application/json'},method='POST')
        try:
            with urllib.request.urlopen(request,timeout=60) as response:
                status=response.status;data=response.read()
        except urllib.error.HTTPError as error:
            status=error.code;data=error.read()
        except Exception:
            status=502;data=b'{"error":"relay transport failed"}'
        safe=data.replace(owner.key.encode(),b'[redacted credential]')
        (owner.run/'responses'/f'{number:03}.json').write_bytes(safe)
        try:
            self.send_response(status);self.send_header('content-type','application/json');self.send_header('content-length',str(len(data)));self.end_headers();self.wfile.write(data)
        except (BrokenPipeError,ConnectionResetError): pass

def launch(args):
    binding=json.loads((HERE/'initial-binding.json').read_text());paths={name:entry['path'] for name,entry in binding['executables'].items()}
    if args.cli:paths['cli']=args.cli
    if args.workflow:paths['workflow']=args.workflow
    preflight(binding,paths)
    run=HERE/args.name;run.mkdir();workspace=run/'workspace';workspace.mkdir();(run/'requests').mkdir();(run/'responses').mkdir()
    key=credential(args.vendor);relay=Relay(args.vendor,key,run);thread=threading.Thread(target=relay.serve_forever);thread.start()
    env=os.environ.copy()
    for name in list(env):
        if name.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_')):del env[name]
    env.update(LLM_VENDOR=args.vendor,LLM_MODEL=args.model,LLM_RESOLVED_MODEL=args.model.removeprefix('models/'),LLM_API_KEY=key,LLM_BASE_URL=f'http://127.0.0.1:{relay.server_port}',CH02_LOG=str(run/'session.log'),ENSEMBLE_RUN_DIRECTORY=str(run))
    executable='cli' if args.mode=='chat' else 'workflow';command=[paths[executable],args.mode]
    receipt={'start':datetime.datetime.now(datetime.timezone.utc).isoformat(),'actor':'Codex student coder, not Bill','vendor':args.vendor,'requested_model':args.model,'command':command,'workspace':str(workspace),'source_revision':binding['source_revision'],'executables':{n:binding['executables'][n]['sha256'] for n in paths},'support':binding['support'],'mode':args.mode,'launched_executable':executable,'executable_paths':paths,'transport':'actual execution-tool PTY and macOS script terminal; raw body relay'}
    (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n')
    try:
        code=subprocess.call([paths['recorder'],'-q',str(run/'terminal.txt'),*command],cwd=workspace,env=env)
    finally:
        relay.shutdown();relay.server_close();thread.join()
    receipt.update(exit_code=code,end=datetime.datetime.now(datetime.timezone.utc).isoformat(),requests=relay.number)
    (run/'launch.json').write_text(json.dumps(receipt,indent=2)+'\n');raise SystemExit(code)

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('vendor');parser.add_argument('model',nargs='?');parser.add_argument('mode',nargs='?',choices=['chat','workflow','collection']);parser.add_argument('name',nargs='?');parser.add_argument('--cli');parser.add_argument('--workflow');args=parser.parse_args()
    if args.vendor=='discover':discover()
    else:launch(args)
