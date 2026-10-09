"""Small local controls. Synthetic identities never claim real Go builds."""
import copy
import http.client
import http.server
import json
from pathlib import Path
import socket
import sys
import tempfile
import threading
import time
import unittest
from unittest.mock import patch

import identity as ident
from driver import launch_record
from relay import Budget, Relay


class IdentityControls(unittest.TestCase):
    def setUp(self):
        self.temp=tempfile.TemporaryDirectory();self.addCleanup(self.temp.cleanup);self.root=Path(self.temp.name)
        self.support=self.root/'support';self.support.mkdir();(self.support/'catalog').mkdir()
        # A complete synthetic parent set, with real file hashes. Git lookup is
        # substituted here; no nested repository or unbuilt binary claim.
        self.blobs={ident.PREFIX+'main.go':b'package main\n',ident.PREFIX+'go.mod':b'module example.test\n',ident.PREFIX+'session-api.md':b'api',ident.SUPPORT_PREFIX+'schedule.json':b'{}',ident.SUPPORT_PREFIX+'driver.py':b'# fixture',ident.SUPPORT_PREFIX+'catalog/SKILL.md':b'manual',ident.SUPPORT_PREFIX+'bindings.json':b'{}'}
        for p,b in self.blobs.items():
            f=self.root/p;f.parent.mkdir(parents=True,exist_ok=True);f.write_bytes(b)
            if p.startswith(ident.SUPPORT_PREFIX):(self.support/p.removeprefix(ident.SUPPORT_PREFIX)).write_bytes(b)
        self.sources={p:ident.digest(b) for p,b in self.blobs.items()}
        self.modules={p:h for p,h in self.sources.items() if p.endswith('go.mod')}
        self.dep=self.root/'deps';self.dep.mkdir();(self.dep/'package-lock.json').write_text('{}');(self.dep/'engine.js').write_text('fixture')
        self.bin=self.root/'binary';self.bin.write_bytes(b'not executable: local identity fixture')
        binaries={k:{'path':str(self.bin),'sha256':ident.file_hash(self.bin)} for k in ident.ROLES}
        binaries['python']={'path':sys.executable,'sha256':ident.file_hash(sys.executable)}
        support={p:h for p,h in self.sources.items() if p.startswith(ident.SUPPORT_PREFIX)}
        self.b={'source_revision':'1'*40,'sources':self.sources,'support':support,'modules':self.modules,'catalog':{p:h for p,h in support.items() if '/catalog/' in p or p.endswith('/bindings.json')},'binaries':binaries,'builds':{role:{'source_revision':'1'*40,'binary_sha256':ident.file_hash(self.bin),'module':ident.PREFIX+d,'target':t,'exit':0,'modules':self.modules,'go_version':'SYNTHETIC CONTROL','module_graph':'SYNTHETIC CONTROL','binary_build_info':'SYNTHETIC CONTROL'} for role,(d,t) in ident.BUILD_ROLES.items()},'browser_dependencies':{'root':str(self.dep),'files':{p:ident.file_hash(self.dep/p) for p in ident.dependency_paths(self.dep)}},'schedule_sha256':support[ident.SUPPORT_PREFIX+'schedule.json']}
        self.patches=[patch.object(ident,'source_paths',return_value=sorted(self.blobs)),patch.object(ident,'historical',side_effect=lambda repo,rev,path:self.blobs[path])]
        for p in self.patches:p.start();self.addCleanup(p.stop)
        self.launch=launch_record(self.b,'anthropic','fixture','cli',[str(self.bin),'chat'],{'LLM_VENDOR':'anthropic','LLM_MODEL':'fixture'},self.root,self.root)

    def check(self,b=None):return ident.preflight(b or self.b,self.root,support_root=self.support)
    def test_valid_parent(self):self.check();ident.check_launch(self.b,self.launch)
    def test_intended_mismatches_before_write(self):
        # Each mutation begins with a passing parent and isolates one property.
        cases=[('source','source identity',lambda b:b['sources'].__setitem__(ident.PREFIX+'main.go','0'*64)),('missing_source','incomplete source',lambda b:b['sources'].pop(ident.PREFIX+'main.go')),('support_map','incomplete support',lambda b:b['support'].pop(ident.SUPPORT_PREFIX+'driver.py')),('module','module dependency',lambda b:b['modules'].__setitem__(ident.PREFIX+'go.mod','0'*64)),('catalog','catalog/binding',lambda b:b['catalog'].__setitem__(ident.SUPPORT_PREFIX+'bindings.json','0'*64)),('build','build association',lambda b:b['builds']['consumer'].__setitem__('source_revision','2'*40)),('dependency','browser dependency identity',lambda b:b['browser_dependencies']['files'].__setitem__('engine.js','0'*64))]
        for role in ('cli','gui','consumer'):cases.append((role,role+' binary identity',lambda b,r=role:b['binaries'][r].__setitem__('sha256','0'*64)))
        for name,message,mutate in cases:
            with self.subTest(name=name):
                self.check();b=copy.deepcopy(self.b);mutate(b)
                sentinel=self.root/'derived.json'
                with self.assertRaisesRegex(ValueError,message):
                    self.check(b);sentinel.write_text('should not write')
                self.assertFalse(sentinel.exists())
        self.check();(self.support/'driver.py').write_text('changed')
        with self.assertRaisesRegex(ValueError,'support identity'):self.check()
    def test_launch_mutations(self):
        for field,value in [('command',['wrong']),('source_revision','2'*40),('model','wrong'),('workspace','elsewhere')]:
            with self.subTest(field=field):
                ident.check_launch(self.b,self.launch);v=copy.deepcopy(self.launch);v[field]=value
                with self.assertRaises((ValueError,IndexError)):ident.check_launch(self.b,v)
    def test_derived_write_guard(self):
        import argparse
        import verify
        run=self.root/'run';run.mkdir()
        (run/'launch.json').write_text(json.dumps(self.launch));(run/'exit.json').write_text('{"exit":0}')
        (run/'receipt.txt').write_text('original')
        binding=self.root/'binding.json';binding.write_text(json.dumps(self.b))
        args=argparse.Namespace(binding=binding,repo=self.root,run=run,seal=True)
        with patch.object(verify,'preflight',side_effect=lambda b,r:self.check(b)):
            verify.run(args);args.seal=False
            valid=copy.deepcopy(self.b)
            valid['sources'][ident.PREFIX+'main.go']='0'*64;binding.write_text(json.dumps(valid))
            before={str(p):ident.file_hash(p) for p in run.rglob('*') if p.is_file()}
            with self.assertRaisesRegex(ValueError,'source identity'):verify.run(args)
            self.assertEqual(before,{str(p):ident.file_hash(p) for p in run.rglob('*') if p.is_file()})
            self.assertFalse((run/'derived').exists())
            binding.write_text(json.dumps(self.b));verify.run(args)
            self.assertTrue((run/'derived/verified.json').exists())

    def test_untracked_source_refuses(self):
        self.check();(self.root/ident.PREFIX/'extra.go').write_text('package main')
        with self.assertRaisesRegex(ValueError,'working source set'):self.check()

    def test_dependency_addition_refuses(self):
        self.check();(self.dep/'new.js').write_text('new')
        with self.assertRaisesRegex(ValueError,'incomplete browser dependencies'):self.check()
    def test_empty_source_refuses(self):
        b=copy.deepcopy(self.b);b['sources']={}
        with self.assertRaisesRegex(ValueError,'incomplete source'):self.check(b)


class FixtureControls(unittest.TestCase):
    def test_three_wire_shapes_and_stream_framing(self):
        import importlib.util
        spec=importlib.util.spec_from_file_location('local_fixture',Path(__file__).with_name('local-fixture.py'))
        module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
        for path in ('/v1/messages','/v1/chat/completions','/v1beta/models/fixture:generateContent'):
            with self.subTest(path=path):
                payload={'model':'fixture','messages':[{'role':'user','content':'Use write_file once to create marker.txt with CH10-A-anthropic'}]}
                kind,raw=module.response(path,payload,1)
                self.assertEqual(kind,'application/json');parsed=json.loads(raw);self.assertIn('write_file',json.dumps(parsed))
                payload['stream']=True;kind,raw=module.response(path,payload,1);self.assertEqual(kind,'text/event-stream')
                for frame in raw.decode().split('\n\n'):
                    if frame and frame!='data: [DONE]':json.loads(frame.removeprefix('data: '))


class BudgetControls(unittest.TestCase):
    def setUp(self):
        self.temp=tempfile.TemporaryDirectory();self.addCleanup(self.temp.cleanup);self.path=Path(self.temp.name)/'attempts.jsonl'
    def test_concurrent_ceiling_and_restart(self):
        errors=[];threads=[]
        def reserve():
            try:Budget(self.path,'anthropic').reserve('A1')
            except ValueError:errors.append('refused')
        for _ in range(10):threads.append(threading.Thread(target=reserve))
        for t in threads:t.start()
        for t in threads:t.join()
        self.assertEqual(len(self.path.read_text().splitlines()),3);self.assertEqual(len(errors),7)
        with self.assertRaisesRegex(ValueError,'step ceiling'):Budget(self.path,'anthropic').reserve('A1')
    def test_row_deadline_and_discovery(self):
        Budget(self.path,'openai',lambda:0).reserve('A1')
        with self.assertRaisesRegex(ValueError,'row deadline'):Budget(self.path,'openai',lambda:601).reserve('A2')
        Budget(self.path,'openai').reserve('models','discovery')
        with self.assertRaisesRegex(ValueError,'discovery ceiling'):Budget(self.path,'openai').reserve('next-page','discovery')
    def test_full_22_and_owner(self):
        b=Budget(self.path,'gemini')
        for step,cap in [('A1',3),('A2',3),('B',4),('C1',2),('C2',2),('C3',2),('D1',3),('D2',3)]:
            for _ in range(cap):b.reserve(step)
        with self.assertRaisesRegex(ValueError,'row/provider ceiling'):b.reserve('B')
        with self.assertRaisesRegex(ValueError,'owner mismatch'):Budget(self.path,'openai').reserve('C1')


class TransportControls(unittest.TestCase):
    def test_failure_timeout_redirect_and_disconnect_are_counted(self):
        with tempfile.TemporaryDirectory() as td:
            root=Path(td);budget=Budget(root/'budget.jsonl','anthropic');reached=[]
            class Backend(http.server.BaseHTTPRequestHandler):
                def log_message(self,*args):pass
                def do_POST(self):
                    reached.append(self.path)
                    # The admission must be visible before the backend begins.
                    self.server.test.assertGreaterEqual(len(budget.path.read_text().splitlines()),len(reached))
                    self.rfile.read(int(self.headers['Content-Length']))
                    if len(reached)==1:self.send_response(307);self.send_header('Location','http://127.0.0.1:1/must-not-follow');self.end_headers();return
                    if len(reached)==2:time.sleep(.25)
                    try:self.send_response(200);self.end_headers();self.wfile.write(b'{}')
                    except OSError:pass
            upstream=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend);upstream.test=self
            t=threading.Thread(target=upstream.serve_forever);t.start()
            relay=Relay(f'http://127.0.0.1:{upstream.server_port}',{},budget,root,['A1'],timeout=.05)
            rt=threading.Thread(target=relay.serve_forever);rt.start()
            try:
                for i in range(2):
                    c=http.client.HTTPConnection('127.0.0.1',relay.server_port,timeout=2);c.request('POST','/A1/v1/messages',b'{"max_tokens":4096}',{'Content-Type':'application/json'});r=c.getresponse();r.read();self.assertEqual(r.status,307 if i==0 else 502);c.close()
                c=socket.create_connection(('127.0.0.1',relay.server_port));c.sendall(b'POST /A1/v1/messages HTTP/1.0\r\nContent-Length: 19\r\n\r\n{"max_tokens":4096} ');c.close()
                deadline=time.monotonic()+2
                while len(reached)<3 and time.monotonic()<deadline:time.sleep(.01)
                self.assertEqual(len(reached),3)
                c=http.client.HTTPConnection('127.0.0.1',relay.server_port,timeout=2);c.request('POST','/A1/v1/messages',b'{"max_tokens":4096}');r=c.getresponse();r.read();self.assertEqual(r.status,429);c.close()
                self.assertEqual(len(reached),3);self.assertEqual(len(budget.path.read_text().splitlines()),3)
            finally:relay.shutdown();relay.server_close();rt.join();upstream.shutdown();upstream.server_close();t.join()
    def test_dripping_response_deadline_and_redaction(self):
        with tempfile.TemporaryDirectory() as td:
            root=Path(td);seen=[]
            class Backend(http.server.BaseHTTPRequestHandler):
                protocol_version='HTTP/1.1'
                def log_message(self,*args):pass
                def do_POST(self):
                    self.close_connection=True
                    self.rfile.read(int(self.headers['Content-Length']));seen.append(1)
                    self.send_response(200);self.send_header('Content-Length','200');self.end_headers()
                    try:
                        self.wfile.write(b'LOCAL-CANARY');self.wfile.flush()
                        for _ in range(50):time.sleep(.01);self.wfile.write(b'.');self.wfile.flush()
                    except OSError:pass
            upstream=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend);t=threading.Thread(target=upstream.serve_forever);t.start()
            relay=Relay(f'http://127.0.0.1:{upstream.server_port}',{'x-api-key':'LOCAL-CANARY'},Budget(root/'b','openai'),root,['C1'],timeout=.08)
            rt=threading.Thread(target=relay.serve_forever);rt.start()
            try:
                c=http.client.HTTPConnection('127.0.0.1',relay.server_port,timeout=2);start=time.monotonic();c.request('POST','/C1/v1/chat/completions',b'{"max_tokens":4096}');r=c.getresponse();r.read();c.close()
                self.assertLess(time.monotonic()-start,.4)
            finally:relay.shutdown();relay.server_close();rt.join();upstream.shutdown();upstream.server_close();t.join()
            self.assertEqual(len(seen),1)
            self.assertNotIn(b'LOCAL-CANARY',(root/'001-response.body').read_bytes())
            self.assertIn(b'[redacted credential]',(root/'001-response.body').read_bytes())
            rows=[json.loads(x) for x in (root/'transport.jsonl').read_text().splitlines()]
            self.assertNotEqual(rows[-1]['outcome'],'completed');self.assertTrue(rows[-1]['credential_redacted'])

    def test_external_upstream_refused_without_network(self):
        with tempfile.TemporaryDirectory() as td:
            with self.assertRaisesRegex(ValueError,'literal loopback'):
                Relay('https://example.invalid',{},Budget(Path(td)/'b','gemini'),td,['D2'])


if __name__=='__main__':unittest.main(verbosity=2)
