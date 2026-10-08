"""Local, credential-free controls for capture and the complete preflight boundary."""
import copy
import http.server
import importlib.util
import json
from pathlib import Path
import tempfile
import threading
import urllib.request
from evidence import HERE, digest, make_binding, preflight

def module(name, filename):
    spec=importlib.util.spec_from_file_location(name,HERE/filename)
    result=importlib.util.module_from_spec(spec);spec.loader.exec_module(result);return result

def hashes(path):
    return {str(p.relative_to(path)):digest(p.read_bytes()) for p in path.rglob('*') if p.is_file()}

launcher=module('launcher','terminal-run.py');verifier=module('verifier','verify-receipts.py')
receipts=HERE/'local-controls';receipts.mkdir(exist_ok=True)
raw=b'{"fixture":"raw bytes preserved"}'
class Upstream(http.server.BaseHTTPRequestHandler):
    def log_message(self,*args):pass
    def do_POST(self):
        assert self.rfile.read(int(self.headers['Content-Length']))==raw
        assert self.headers['x-api-key']=='fixture-key'
        self.send_response(200);self.end_headers();self.wfile.write(b'{"fixture_response":true}')

upstream=http.server.ThreadingHTTPServer(('127.0.0.1',0),Upstream)
thread=threading.Thread(target=upstream.serve_forever);thread.start()
launcher.ORIGINS['anthropic']=f'http://127.0.0.1:{upstream.server_port}'
for name in ('requests','responses'):(receipts/name).mkdir(exist_ok=True)
relay=launcher.Relay('anthropic','fixture-key',receipts);relay_thread=threading.Thread(target=relay.serve_forever);relay_thread.start()
try:
    request=urllib.request.Request(f'http://127.0.0.1:{relay.server_port}/v1/messages',data=raw,method='POST')
    with urllib.request.urlopen(request) as response:assert response.read()==b'{"fixture_response":true}'
finally:
    relay.shutdown();relay.server_close();relay_thread.join();upstream.shutdown();upstream.server_close();thread.join()
assert (receipts/'requests/001.json').read_bytes()==raw
assert b'fixture-key' not in b''.join(p.read_bytes() for p in receipts.rglob('*.json'))

# /usr/bin/true is deliberately a fixture, not a claimed Ensemble build.
binding=make_binding('55e64115ee5243cac6b4958fd12e6811862ae033',{'cli':'/usr/bin/true','workflow':'/usr/bin/true'})
(receipts/'positive-binding-fixture.json').write_text(json.dumps(binding,indent=2)+'\n')
preflight(binding,{})
results=[]
mutations=[('empty source map',lambda b:b.update(sources={})),('missing source',lambda b:b['sources'].pop(next(iter(b['sources'])))),('extra source',lambda b:b['sources'].update({'invented.go':'bad'})),('wrong source hash',lambda b:b['sources'].update({next(iter(b['sources'])):'bad'}))]
for name in ('cli','workflow','interpreter','recorder'):
    mutations.append(('wrong '+name,lambda b,n=name:b['executables'][n].update(sha256='bad')))
for name,mutation in mutations:
    changed=copy.deepcopy(binding);mutation(changed);before=hashes(receipts)
    try:preflight(changed,{})
    except AssertionError as error:results.append({'control':name,'refusal':str(error),'raw_unchanged':hashes(receipts)==before})
    else:raise AssertionError('negative control accepted: '+name)
    assert hashes(receipts)==before

# The last launch is invalid. Every earlier launch must be checked without any
# replay or derived write, and none of their raw receipts may be changed.
with tempfile.TemporaryDirectory() as temporary:
    root=Path(temporary);verifier.HERE=root
    binding['runs']=sorted(prefix+'-'+vendor for prefix in ('controls','eof','workflow','collection') for vendor in ('anthropic','openai','gemini'))
    for name in binding['runs']:
        run=root/name;run.mkdir();mode='workflow' if name.startswith('workflow') else 'collection' if name.startswith('collection') else 'chat';executable='cli' if mode=='chat' else 'workflow'
        paths={n:e['path'] for n,e in binding['executables'].items()}
        launch={'source_revision':binding['source_revision'],'executables':{n:e['sha256'] for n,e in binding['executables'].items()},'support':binding['support'],'exit_code':0,'mode':mode,'launched_executable':executable,'executable_paths':paths,'command':[paths[executable],mode]}
        (run/'launch.json').write_text(json.dumps(launch));(run/'terminal.txt').write_text('original raw terminal fixture')
    last=root/binding['runs'][-1]/'launch.json';launch=json.loads(last.read_text());launch['executables']['workflow']='wrong last launch';last.write_text(json.dumps(launch))
    before=hashes(root)
    try:verifier.verify(binding,{'cli':Path('/usr/bin/true'),'workflow':Path('/usr/bin/true')})
    except AssertionError as error:results.append({'control':'last launch identity','refusal':str(error),'raw_unchanged':hashes(root)==before})
    else:raise AssertionError('invalid last launch passed')
    assert hashes(root)==before
    (receipts/'all-launch-negative-original-hashes.json').write_text(json.dumps(before,indent=2)+'\n')
(receipts/'results.json').write_text(json.dumps({'backend':'local fake only; no paid requests','positive_capture':{'request':digest(raw),'response':digest((receipts/'responses/001.json').read_bytes())},'controls':results},indent=2)+'\n')
print(json.dumps({'capture':'passed','preflight_controls':results},indent=2))
