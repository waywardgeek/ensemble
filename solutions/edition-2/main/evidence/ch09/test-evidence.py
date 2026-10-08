"""Local support positives and single-identity negatives. Never reads credentials."""
import argparse
import copy
import http.server
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import shutil
import tempfile
import threading
import urllib.error
import urllib.request
from evidence import HERE, CAPS, preflight, launch_identity, historical, file_digest, scenario_inputs, digest

def module(name,path):
    spec=importlib.util.spec_from_file_location(name,path);value=importlib.util.module_from_spec(spec);spec.loader.exec_module(value);return value

p=argparse.ArgumentParser();p.add_argument('--binding',required=True,type=Path);args=p.parse_args()
binding=json.loads(args.binding.read_text());paths={k:v['path'] for k,v in binding['executables'].items()}
preflight(binding,paths)
launcher=module('launcher',HERE/'terminal-run.py');verifier=module('verifier',HERE/'verify-receipts.py');checks=[]
with tempfile.TemporaryDirectory(prefix='ch09-evidence-controls-') as temporary:
    root=Path(temporary);relay_run=root/'relay';relay_run.mkdir();(relay_run/'requests').mkdir();(relay_run/'responses').mkdir();release=threading.Event()
    class Stream(http.server.BaseHTTPRequestHandler):
        def log_message(self,*args):pass
        def do_POST(self):
            self.rfile.read(int(self.headers['content-length']));self.send_response(200);self.send_header('content-type','text/event-stream');self.end_headers();self.wfile.write(b'data: FIRST\n\n');self.wfile.flush();assert release.wait(5);self.wfile.write(b'data: LAST\n\n');self.wfile.flush()
    upstream=http.server.ThreadingHTTPServer(('127.0.0.1',0),Stream);worker=threading.Thread(target=upstream.serve_forever);worker.start()
    budget=root/'budget.json';relay=launcher.Relay('anthropic','fixture-key',relay_run,'N',budget,f'http://127.0.0.1:{upstream.server_port}');thread=threading.Thread(target=relay.serve_forever);thread.start()
    try:
        request=urllib.request.Request(f'http://127.0.0.1:{relay.server_port}/v1/messages',data=b'{}')
        with urllib.request.urlopen(request,timeout=5) as response:
            first=response.read1(4096);assert b'FIRST' in first and b'LAST' not in first;release.set();assert b'LAST' in response.read()
        budget.write_text(json.dumps({**{n:0 for n in CAPS},'N':15}))
        with urllib.request.urlopen(request,timeout=5) as response:response.read()
        try:urllib.request.urlopen(request,timeout=5);raise RuntimeError('cap not enforced')
        except urllib.error.HTTPError as error:assert error.code==429
        assert json.loads(budget.read_text())['N']==16
        checks.append('relay preserves first fragment and refuses N request 17 before forwarding')
    finally:
        release.set();relay.shutdown();relay.server_close();thread.join();upstream.shutdown();upstream.server_close();worker.join()
    run=root/'fixture';run.mkdir();(run/'requests').mkdir();(run/'responses').mkdir();catalog=run/'catalog';catalog.mkdir()
    workspace=run/'workspace';workspace.mkdir()
    for path,data in scenario_inputs('N').items():
        for base in (workspace,run/'inputs'):
            target=base/path;target.parent.mkdir(parents=True,exist_ok=True);target.write_bytes(data)
    for source in binding['catalog']:
        target=catalog/Path(source).parent.name/'SKILL.md';target.parent.mkdir();target.write_bytes(historical(binding['source_revision'],source))
    requests=[]
    class Model(http.server.BaseHTTPRequestHandler):
        def log_message(self,*args):pass
        def do_POST(self):
            body=self.rfile.read(int(self.headers['content-length']));requests.append(body);number=len(requests);(run/'requests'/f'{number:03}.json').write_bytes(body)
            if number==1:content=[dict(type='tool_use',id='load',name='load_skill',input=dict(name='edit'))]
            elif number==2:content=[dict(type='tool_use',id='write',name='write_file',input=dict(path='proof',content='proof bytes'))]
            else:content=[dict(type='text',text='done')]
            response=json.dumps(dict(content=content,usage=dict(input_tokens=1,output_tokens=1))).encode();(run/'responses'/f'{number:03}.body').write_bytes(response)
            self.send_response(200);self.send_header('content-type','application/json');self.end_headers();self.wfile.write(response)
    server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Model);thread=threading.Thread(target=server.serve_forever);thread.start()
    env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
    env.update(EN_DISABLE_STREAMING='1',LLM_VENDOR='anthropic',LLM_MODEL='fixture',LLM_API_KEY='fixture',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',CH02_LOG=str(run/'session.log'),LLM_SKILLS_DIR=str(catalog),LLM_PRIMARY_SKILL='base')
    try:
        result=subprocess.run([paths['cli'],'protocol'],input=b'{"user":"local load/write fixture"}\n',env=env,cwd=workspace,check=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE,timeout=10)
    finally:server.shutdown();server.server_close();thread.join()
    assert (workspace/'proof').read_text()=='proof bytes' and len(requests)==3
    (run/'terminal.txt').write_bytes(b'LOCAL PROTOCOL FIXTURE, NOT A LIVE PTY\n'+result.stdout)
    launch={**launch_identity(binding,'cli','N'),'scope':'local protocol fixture','workspace':str(workspace),'exit_code':0,'requests':3,'logs':['session.log'],'catalog_files':{str(p.relative_to(run)):file_digest(p) for p in catalog.rglob('SKILL.md')},'scratch_inputs':{'workspace/'+p:digest(b) for p,b in scenario_inputs('N').items()},'scratch_outputs':{str(p.relative_to(run)):file_digest(p) for p in workspace.rglob('*') if p.is_file()}}
    originals=[run/'session.log',run/'terminal.txt',*run.glob('requests/*.json'),*run.glob('responses/*.body'),*catalog.rglob('SKILL.md'),*(run/'inputs').rglob('notes.txt'),*workspace.iterdir()]
    launch['originals']={str(p.relative_to(run)):file_digest(p) for p in originals};(run/'launch.json').write_text(json.dumps(launch))
    events=[json.loads(line) for line in (run/'session.log').read_text().splitlines()][1:]
    result_seq=next(e['seq'] for e in events if e['type']=='tool_returned' and e['tool']['call_id']=='load')
    redactions={str((run/'session.log').resolve()):result_seq}
    verifier.verify(binding,[run],paths,root/'positive',redactions)
    redaction=json.loads(next((root/'positive').glob('*-redaction.json')).read_text())
    assert '[skill edit activation' in json.dumps(redaction['end_projection_after_redaction'])
    assert '[redacted]' in json.dumps(redaction['end_projection_after_redaction'])
    assert redaction['end_projection_after_redaction']!=redaction['end_projection_before_redaction']
    assert redaction['reconstructed_original_request']==json.loads(requests[-1])
    checks.append('complete historical binding, exact captured request bytes, separate result-redaction projection')
    cases=[('source-hash','historical source mismatch'),('source-set','incomplete historical source set'),('empty-source','empty source set'),('binary','executable mismatch'),('support','historical support mismatch'),('catalog','catalog identity mismatch'),('browser-set','incomplete browser dependency set'),('browser-hash','browser dependency mismatch'),('launch-source','launch source_revision mismatch'),('launch-binary','launch executables mismatch'),('launch-support','launch support mismatch'),('launch-catalog','launch catalog mismatch'),('capture','original receipt mismatch')]
    for name,reason in cases:
        candidate=copy.deepcopy(binding);changed=copy.deepcopy(launch);chosen=next(iter(candidate['sources']))
        if name=='source-hash':candidate['sources'][chosen]='0'*64
        if name=='source-set':candidate['sources'].pop(chosen)
        if name=='empty-source':candidate['sources']={}
        if name=='binary':candidate['executables']['cli']['sha256']='0'*64
        if name=='support':candidate['support']['evidence.py']='0'*64
        if name=='catalog':candidate['catalog'][next(iter(candidate['catalog']))]='0'*64
        if name=='browser-set':candidate['browser_tools'].pop(next(iter(candidate['browser_tools'])))
        if name=='browser-hash':candidate['browser_tools'][next(iter(candidate['browser_tools']))]='0'*64
        if name=='launch-source':changed['source_revision']='0'*40
        if name=='launch-binary':changed['executables']['cli']='0'*64
        if name=='launch-support':changed['support']['evidence.py']='0'*64
        if name=='launch-catalog':changed['catalog'][next(iter(changed['catalog']))]='0'*64
        if name=='capture':changed['originals']['requests/001.json']='0'*64
        (run/'launch.json').write_text(json.dumps(changed));output=root/name
        try:verifier.verify(candidate,[run],paths,output);raise RuntimeError('negative accepted')
        except AssertionError as error:assert reason in str(error),str(error);assert not output.exists();checks.append(name+': '+reason)
        (run/'launch.json').write_text(json.dumps(launch))
    # Begin every late-run refusal from a passing two-run fixture. The seam
    # counts only replay/redaction execution; immutable Git reads still occur.
    other=root/'last-fixture';shutil.copytree(run,other)
    other_launch=copy.deepcopy(launch);other_launch['workspace']=str(other/'workspace')
    (other/'launch.json').write_text(json.dumps(other_launch))
    verifier.verify(binding,[run,other],paths,root/'two-run-positive',redactions)
    execute=verifier.execute
    invoked=[]
    def forbid(command,run):
        invoked.append(command);raise RuntimeError('replay before all-run preflight')
    verifier.execute=forbid
    try:
        late_cases=[('capture','incomplete original receipt set'),('response','incomplete original receipt set'),('log','incomplete original receipt set'),('catalog','launch catalog file mismatch'),('scratch','scratch input mismatch'),('terminal','incomplete original receipt set'),('logs-empty','missing log identities'),('requests-empty','empty request set')]
        for name,reason in late_cases:
            changed=copy.deepcopy(other_launch)
            missing={'capture':'requests/001.json','response':'responses/001.body','log':'session.log','terminal':'terminal.txt'}
            if name in missing:changed['originals'].pop(missing[name])
            if name=='catalog':changed['catalog_files'].pop(next(iter(changed['catalog_files'])))
            if name=='scratch':changed['scratch_inputs']['workspace/notes.txt']='0'*64
            if name=='logs-empty':changed['logs']=[]
            if name=='requests-empty':changed['requests']=0
            (other/'launch.json').write_text(json.dumps(changed));output=root/('late-'+name)
            try:verifier.verify(binding,[run,other],paths,output,redactions);raise RuntimeError('late negative accepted')
            except AssertionError as error:assert reason in str(error),str(error);assert not output.exists() and not invoked;checks.append('last-run '+name+': '+reason+' before replay/redaction')
        try:verifier.verify(binding,[],paths,root/'empty-runs');raise RuntimeError('empty accepted')
        except AssertionError as error:assert 'empty run set' in str(error);assert not invoked
    finally:verifier.execute=execute
    # The actual GUI/browser support path is an explicitly zero-request local
    # control, separate from verify()'s nonempty paid-demonstration contract.
    browser_run=root/'browser';browser_run.mkdir()
    browser_launch=launch_identity(binding,'gui','G')
    (browser_run/'launch.json').write_text(json.dumps(browser_launch))
    gui_env=env.copy();gui_env['CH02_LOG']=str(browser_run/'session.log')
    gui=subprocess.Popen([paths['gui'],'--port','0'],cwd=browser_run,env=gui_env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    try:
        url=gui.stdout.readline().strip();assert url.startswith('http://127.0.0.1:'),url
        actions=[{'action':'inspect'},{'action':'open'},{'action':'reload'},{'action':'quit'}]
        controlled=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(browser_run),url,str(args.binding.resolve())],input=''.join(json.dumps(a)+'\n' for a in actions),capture_output=True,text=True,timeout=40)
        assert controlled.returncode==0,(controlled.stdout,controlled.stderr)
        rows=[json.loads(line) for line in (browser_run/'browser-original.jsonl').read_text().splitlines()]
        assert not [row for row in rows if row['kind']=='action_failed'],controlled.stdout
        assert any(row['kind']=='page_receipt' and 'base' in row['text'] for row in rows)
        assert not any(row['kind']=='browser_sent' and json.loads(row['payload']).get('type')=='prompt' for row in rows)
        captured=json.loads((browser_run/'browser-binding.json').read_text())
        assert all(file_digest(browser_run/path)==sha for path,sha in captured['originals'].items())
        checks.append('bound actual GUI and two browser views/reconnect, retained receipt hashes; zero model requests')
        negative=root/'browser-negative';negative.mkdir();bad=copy.deepcopy(browser_launch);bad['executables']['gui']='0'*64;(negative/'launch.json').write_text(json.dumps(bad))
        refusal=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(negative),url,str(args.binding.resolve())],input='{"action":"quit"}\n',capture_output=True,text=True,timeout=15)
        assert refusal.returncode!=0 and 'launch executables mismatch' in refusal.stderr and not (negative/'browser-original.jsonl').exists()
        checks.append('browser launch executable identity refused before browser receipts')
    finally:gui.terminate();gui.wait(timeout=10)
print(json.dumps({'passed':True,'scope':'local fixtures and identity refusals; zero provider calls or credentials','checks':checks},indent=2))
