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
from evidence import HERE, compare_neutral_events, preflight

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
        (run/'budget.json').write_text(json.dumps({'requests':23}))
        with urllib.request.urlopen(request,timeout=5) as response:response.read()
        try:urllib.request.urlopen(request,timeout=5);raise RuntimeError('HTTP request ceiling not enforced')
        except urllib.error.HTTPError as error:assert error.code==429
        assert json.loads((run/'budget.json').read_text())['requests']==24
        results.append({'check':'revision-provider-http-ceiling-forward24-refuse25','passed':True})
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
    for name,reason in [('source-hash','historical source mismatch'),('source-set','incomplete historical source set'),('binary','executable mismatch'),('launch-source','launch source mismatch'),('launch-binary','launch executable mismatch'),('support','historical support mismatch'),('launch-support','launch support mismatch'),('empty-source','empty source set'),('browser-tools','browser dependency mismatch'),('browser-tools-set','incomplete browser dependency set'),('launch-browser-tools','launch browser tools mismatch')]:
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
        if name=='browser-tools-set':candidate['browser_tools'].pop(next(iter(candidate['browser_tools'])))
        if name=='launch-browser-tools':newlaunch['browser_tools']={**newlaunch['browser_tools'],next(iter(newlaunch['browser_tools'])):'0'*64}
        (fixture/'launch.json').write_text(json.dumps(newlaunch));output=root/name
        try:verifier.verify(candidate,[fixture],paths,output);raise RuntimeError('negative control accepted')
        except AssertionError as error:
            assert reason in str(error),str(error);assert not output.exists();results.append({'check':name,'refusal':str(error),'passed':True})
        (fixture/'launch.json').write_text(json.dumps(launch))
    # Exercise the actual Node launch adapter, starting from a successful path.
    browser_run=root/'browser';browser_run.mkdir()
    browser_launch={**launch,'launched_executable':'gui'}
    (browser_run/'launch.json').write_text(json.dumps(browser_launch))
    browser_env=env.copy();browser_env['CH02_LOG']=str(browser_run/'session.log')
    gui=subprocess.Popen([paths['gui'],'--port','0'],cwd=browser_run,env=browser_env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    try:
        url=gui.stdout.readline().strip();assert url.startswith('http://127.0.0.1:')
        command=[paths['node'],str(HERE/'browser-live.mjs'),str(browser_run),url]
        actions=[{'action':'settings'},{'action':'open'},{'action':'settings','page':1},{'action':'drag-divider','field':'sidebar_width','delta':20},{'action':'wait','page':1,'text':'widths 280/380'},{'action':'inspect'},{'action':'quit'}]
        valid=subprocess.run(command,input=''.join(json.dumps(a)+'\n' for a in actions),capture_output=True,text=True,timeout=30)
        assert valid.returncode==0,(valid.stdout,valid.stderr)
        rows=[json.loads(line) for line in (browser_run/'browser-original.jsonl').read_text().splitlines()]
        assert not [row for row in rows if row['kind']=='action_failed'],valid.stdout
        sent=[json.loads(row['payload']) for row in rows if row['kind']=='browser_sent']
        assert any(row.get('type')=='preferences_update' and row.get('patch',{}).get('sidebar_width')==280 for row in sent),sent
        results.append({'check':'actual-browser-launch-valid-path','passed':True})
        for name,reason in [('browser-launch-source','launch source mismatch'),('browser-launch-dependencies','launch browser tools mismatch')]:
            bad=root/name;bad.mkdir();badlaunch=copy.deepcopy(browser_launch)
            if name=='browser-launch-source':badlaunch['source_revision']='wrong'
            else:badlaunch['browser_tools'].pop(next(iter(badlaunch['browser_tools'])))
            (bad/'launch.json').write_text(json.dumps(badlaunch))
            invalid=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(bad),url],input='',capture_output=True,text=True,timeout=30)
            assert invalid.returncode!=0 and reason in invalid.stderr,invalid.stderr
            assert not (bad/'browser-original.jsonl').exists()
            results.append({'check':name,'refusal':reason,'passed':True})
        # Two deliberate browser prompts are admitted; a third action is refused
        # locally before WebSocket submission. No external provider is involved.
        local=http.server.ThreadingHTTPServer(('127.0.0.1',0),Plain);local_thread=threading.Thread(target=local.serve_forever);local_thread.start()
        consumer_run=root/'consumer';consumer_run.mkdir()
        (consumer_run/'launch.json').write_text(json.dumps({**launch,'launched_executable':'consumer','vendor':'fixture-control'}))
        consumer_env=env.copy();consumer_env['LLM_BASE_URL']=f'http://127.0.0.1:{local.server_port}'
        consumer=subprocess.Popen([paths['consumer']],cwd=consumer_run,env=consumer_env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
        prompt_budget=HERE/'prompt-budget-fixture-control.json'
        assert not prompt_budget.exists(),'local fixture prompt budget already exists'
        prompt_budget.write_text(json.dumps({'prompts':8}))
        try:
            lines=[consumer.stdout.readline().strip() for _ in range(3)];url=lines[-1];assert url.startswith('http://127.0.0.1:')
            actions=[{'action':'prompt','panel':0,'text':'local one'},{'action':'wait','panel':0,'text':'fixture','exact':True},{'action':'close-view','panel':0},{'action':'open-view','panel':0},{'action':'prompt','panel':1,'text':'local two'},{'action':'wait','panel':1,'text':'fixture','exact':True},{'action':'prompt','panel':0,'text':'refused third'},{'action':'quit'}]
            controlled=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(consumer_run),url],input=''.join(json.dumps(a)+'\n' for a in actions),capture_output=True,text=True,timeout=40)
            assert controlled.returncode==0,(controlled.stdout,controlled.stderr)
            rows=[json.loads(line) for line in (consumer_run/'browser-original.jsonl').read_text().splitlines()]
            sent=[json.loads(row['payload']) for row in rows if row['kind']=='browser_sent']
            assert sum(row.get('type')=='prompt' for row in sent)==2
            failures=[row for row in rows if row['kind']=='action_failed']
            assert len(failures)==1 and 'ten-prompt limit' in failures[0]['error'],failures
            results.append({'check':'actual-browser-two-prompts-refuse-third-and-remount','passed':True})
        finally:
            consumer.terminate();consumer.wait(timeout=10);local.shutdown();local.server_close();local_thread.join();prompt_budget.unlink(missing_ok=True)
    finally:
        gui.terminate();gui.wait(timeout=10)
    replay_run=root/'replay';replay_run.mkdir()
    (replay_run/'launch.json').write_text(json.dumps({**launch,'launched_executable':'replay','vendor':'fixture-control'}))
    replay=subprocess.Popen([paths['replay'],'anthropic',str(fixture/'session.log')],cwd=replay_run,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    try:
        url=replay.stdout.readline().strip();assert url.startswith('http://127.0.0.1:')
        actions=[{'action':'wait','text':'fixture','exact':True},{'action':'settings'},{'action':'open'},{'action':'settings','page':1},{'action':'race-preferences','font':19,'width':420},{'action':'inspect','page':1},{'action':'quit'}]
        controlled=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(replay_run),url],input=''.join(json.dumps(a)+'\n' for a in actions),capture_output=True,text=True,timeout=40)
        assert controlled.returncode==0,(controlled.stdout,controlled.stderr)
        rows=[json.loads(line) for line in (replay_run/'browser-original.jsonl').read_text().splitlines()]
        assert not [row for row in rows if row['kind']=='action_failed'],controlled.stdout
        sent=[json.loads(row['payload']) for row in rows if row['kind']=='browser_sent']
        assert sum(row.get('type')=='preferences_update' for row in sent)==2,sent
        assert not any(row.get('type')=='prompt' for row in sent)
        results.append({'check':'public-retained-replay-and-two-page-settings-race','passed':True})
    finally:replay.terminate();replay.wait(timeout=10)
    original=(fixture/'session.log').read_bytes();admitted=(replay_run/'readmitted.jsonl').read_bytes()
    comparison=compare_neutral_events(original,admitted)
    assert comparison['rewritten_admission_times']==comparison['event_count']>0
    for name,reason in [('count','event count changed'),('order','non-time facts changed'),('seq','non-time facts changed'),('payload','non-time facts changed'),('time','admission time missing')]:
        changed=[json.loads(x) for x in admitted.splitlines()]
        if name=='count':changed.pop()
        if name=='order':changed[1],changed[2]=changed[2],changed[1]
        if name=='seq':changed[1]['seq']+=1
        if name=='payload':changed[1]['type']='modified'
        if name=='time':del changed[1]['time']
        try:compare_neutral_events(original,'\n'.join(json.dumps(x) for x in changed).encode());raise RuntimeError('replay mutation accepted')
        except AssertionError as error:assert reason in str(error)
    results.append({'check':'replay-admission-time-only-positive-and-five-exact-refusals','passed':True})
    stream_release=threading.Event()
    class LiveStream(http.server.BaseHTTPRequestHandler):
        def log_message(self,*args):pass
        def do_POST(self):
            self.rfile.read(int(self.headers['content-length']))
            self.send_response(200);self.send_header('content-type','text/event-stream');self.end_headers()
            events=[{'type':'message_start','message':{'model':'fixture','content':[],'usage':{'input_tokens':1}}},{'type':'content_block_start','index':0,'content_block':{'type':'text','text':''}},{'type':'content_block_delta','index':0,'delta':{'type':'text_delta','text':'Nonempty provisional fixture text.'}}]
            for event in events:self.wfile.write(('event: '+event['type']+'\ndata: '+json.dumps(event)+'\n\n').encode())
            self.wfile.flush();stream_release.wait(20)
    local=http.server.ThreadingHTTPServer(('127.0.0.1',0),LiveStream);local_thread=threading.Thread(target=local.serve_forever);local_thread.start()
    stream_run=root/'streaming';stream_run.mkdir()
    (stream_run/'launch.json').write_text(json.dumps({**launch,'launched_executable':'consumer','vendor':'fixture-control'}))
    stream_env=env.copy();stream_env.update(EN_DISABLE_STREAMING='0',LLM_BASE_URL=f'http://127.0.0.1:{local.server_port}')
    consumer=subprocess.Popen([paths['consumer']],cwd=stream_run,env=stream_env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    try:
        lines=[consumer.stdout.readline().strip() for _ in range(3)];url=lines[-1];assert url.startswith('http://127.0.0.1:')
        actions=[{'action':'prompt','panel':1,'text':'local streaming positive'},{'action':'wait-stream','panel':1},{'action':'interrupt','panel':1},{'action':'wait','panel':1,'text':'Outcome: interrupted'},{'action':'quit'}]
        controlled=subprocess.run([paths['node'],str(HERE/'browser-live.mjs'),str(stream_run),url],input=''.join(json.dumps(a)+'\n' for a in actions),capture_output=True,text=True,timeout=40)
        assert controlled.returncode==0,(controlled.stdout,controlled.stderr)
        rows=[json.loads(line) for line in (stream_run/'browser-original.jsonl').read_text().splitlines()]
        assert not [row for row in rows if row['kind']=='action_failed'],controlled.stdout
        assert any(row['kind']=='page_receipt' and 'Nonempty provisional fixture text.' in json.dumps(row) and 'Provisional' in json.dumps(row) for row in rows)
        results.append({'check':'actual-nonempty-provisional-stream-barrier-and-interrupt','passed':True})
    finally:
        stream_release.set();consumer.terminate();consumer.wait(timeout=10);local.shutdown();local.server_close();local_thread.join();(HERE/'prompt-budget-fixture-control.json').unlink(missing_ok=True)
(HERE/'local-evidence-controls.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
