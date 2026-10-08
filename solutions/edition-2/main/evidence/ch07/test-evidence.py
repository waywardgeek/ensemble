"""Offline streaming relay and valid-path-first identity controls; no credentials."""
import copy
import http.server
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile
import threading
import urllib.request
from evidence import HERE, preflight

def module(name,path):
    spec=importlib.util.spec_from_file_location(name,path);value=importlib.util.module_from_spec(spec);spec.loader.exec_module(value);return value
launcher=module('launcher',HERE/'terminal-run.py');verifier=module('verifier',HERE/'verify-receipts.py')
binding=json.loads((HERE/'initial-binding.json').read_text());paths={k:v['path'] for k,v in binding['executables'].items()}
results=[]
with tempfile.TemporaryDirectory() as temp:
    root=Path(temp);release=threading.Event()
    class Stream(http.server.BaseHTTPRequestHandler):
        def log_message(self,*args):pass
        def do_POST(self):
            self.rfile.read(int(self.headers['content-length']))
            self.send_response(200);self.send_header('content-type','text/event-stream');self.end_headers()
            self.wfile.write(b'data: FIRST\n\n');self.wfile.flush()
            assert release.wait(5),'relay buffered first fragment'
            self.wfile.write(b'data: LAST\n\n');self.wfile.flush()
    upstream=http.server.ThreadingHTTPServer(('127.0.0.1',0),Stream);thread=threading.Thread(target=upstream.serve_forever);thread.start()
    run=root/'relay';run.mkdir();(run/'requests').mkdir();(run/'responses').mkdir()
    launcher.ORIGINS['anthropic']=f'http://127.0.0.1:{upstream.server_port}'
    relay=launcher.Relay('anthropic','fixture-key',run);worker=threading.Thread(target=relay.serve_forever);worker.start()
    try:
        request=urllib.request.Request(f'http://127.0.0.1:{relay.server_port}/v1/messages',data=b'{}')
        with urllib.request.urlopen(request,timeout=5) as response:
            first=response.read1(4096);assert b'FIRST' in first and b'LAST' not in first
            release.set();assert b'LAST' in response.read()
        results.append({'check':'relay-first-fragment-before-upstream-release','passed':True})
        (run/'budget.json').write_text(json.dumps({'requests':31}))
        with urllib.request.urlopen(request,timeout=5) as response:response.read()
        try:urllib.request.urlopen(request,timeout=5);raise RuntimeError('HTTP request ceiling not enforced')
        except urllib.error.HTTPError as error:assert error.code==429
        assert json.loads((run/'budget.json').read_text())['requests']==32
        results.append({'check':'provider-global-http-ceiling-forward32-refuse33','passed':True})
    finally:
        release.set();relay.shutdown();relay.server_close();worker.join();upstream.shutdown();upstream.server_close();thread.join()
    fixture=root/'fixture';fixture.mkdir();(fixture/'requests').mkdir()
    class Plain(http.server.BaseHTTPRequestHandler):
        def log_message(self,*args):pass
        def do_POST(self):
            body=self.rfile.read(int(self.headers['content-length']));(fixture/'requests/001.json').write_bytes(body)
            self.send_response(200);self.send_header('content-type','application/json');self.end_headers()
            self.wfile.write(b'{"content":[{"type":"text","text":"fixture"}],"usage":{"input_tokens":1,"output_tokens":1}}')
    server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Plain);thread=threading.Thread(target=server.serve_forever);thread.start()
    env=os.environ.copy();env.update(EN_DISABLE_STREAMING='1',LLM_VENDOR='anthropic',LLM_MODEL='fixture',LLM_API_KEY='fixture',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',CH02_LOG=str(fixture/'session.log'))
    try:
        subprocess.run([paths['cli'],'protocol'],input=b'{"user":"fixture"}\n',env=env,cwd=fixture,check=True,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    finally:server.shutdown();server.server_close();thread.join()
    launch={'source_revision':binding['source_revision'],'executables':{k:v['sha256'] for k,v in binding['executables'].items()},'support':binding['support'],'launched_executable':'cli','delivery':'plain','browser_tools':binding['browser_tools']}
    (fixture/'launch.json').write_text(json.dumps(launch))
    verifier.verify(binding,[fixture],paths,root/'positive')
    results.append({'check':'valid-path-complete-source-executables-launch-and-replay','passed':True})
    for name,reason in [('source-hash','historical source mismatch'),('source-set','incomplete historical source set'),('binary','executable mismatch'),('launch-source','launch source mismatch'),('launch-binary','launch executable mismatch'),('support','historical support mismatch'),('launch-support','launch support mismatch'),('empty-source','empty source set'),('browser-tools','browser dependency mismatch')]:
        candidate=copy.deepcopy(binding);newlaunch=copy.deepcopy(launch);chosen=next(iter(candidate['sources']))
        if name=='source-hash':candidate['sources'][chosen]='0'*64
        if name=='source-set':candidate['sources'].pop(chosen)
        if name=='binary':candidate['executables']['cli']['sha256']='0'*64
        if name=='launch-source':newlaunch['source_revision']='wrong'
        if name=='launch-binary':newlaunch['executables']['cli']='0'*64
        if name=='support':candidate['support']['browser-live.mjs']='0'*64
        if name=='launch-support':newlaunch['support']={**newlaunch['support'],'browser-live.mjs':'0'*64}
        if name=='empty-source':candidate['sources']={}
        if name=='browser-tools':candidate['browser_tools'][next(iter(candidate['browser_tools']))]='0'*64
        (fixture/'launch.json').write_text(json.dumps(newlaunch));output=root/name
        try:verifier.verify(candidate,[fixture],paths,output);raise RuntimeError('negative control accepted')
        except AssertionError as error:
            assert reason in str(error),str(error);assert not output.exists();results.append({'check':name,'refusal':str(error),'passed':True})
        (fixture/'launch.json').write_text(json.dumps(launch))
(HERE/'local-evidence-controls.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
