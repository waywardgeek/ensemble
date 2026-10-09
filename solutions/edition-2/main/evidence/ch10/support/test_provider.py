"""Synthetic keys/localhost only. No ambient settings, auth, discovery or provider."""
import contextlib
import argparse
import copy
import http.client
import http.server
import io
import json
import os
import socket
from pathlib import Path
import tempfile
import threading
import time
import unittest
from unittest.mock import patch

from provider import Provider, Redactor, load_key, require_discovery, FIELDS, ORIGINS, GEMINI_TARGET
from relay import Budget, Relay
from test_support import IdentityControls
import identity as ident
from driver import launch_record, launch

CANARY = 'SYNTHETIC-ONLY-credential-DO-NOT-SEND-123'


class ProviderControls(unittest.TestCase):
    def setUp(self):
        self.env=patch.dict(os.environ, {}, clear=True);self.env.start();self.addCleanup(self.env.stop)
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup);self.root=Path(self.tmp.name)
        self.seen=[];self.status=200;self.response=b'{}';self.credentials=[]
        owner=self
        class Backend(http.server.BaseHTTPRequestHandler):
            def log_message(self,*a): pass
            def respond(self):
                owner.seen.append(self.path)
                # Prove accounting is on disk before the upstream sees any request.
                owner.assertTrue((owner.root/'budget').read_text().strip())
                owner.credentials.append(dict(self.headers))
                self.rfile.read(int(self.headers.get('content-length',0)))
                self.send_response(owner.status)
                self.send_header('Location','http://127.0.0.1:1/'+CANARY)
                self.send_header('X-Leaked-Credential',CANARY)
                self.send_header('Content-Type','text/plain; '+CANARY)
                self.send_header('Content-Length',str(len(owner.response)));self.end_headers()
                try:
                    for i in range(0,len(owner.response),7):self.wfile.write(owner.response[i:i+7]);self.wfile.flush()
                except OSError: pass
            do_POST=respond;do_GET=respond
        self.up=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend)
        self.t=threading.Thread(target=self.up.serve_forever);self.t.start()
        self.addCleanup(self.stop)
        self.origin=f'http://127.0.0.1:{self.up.server_port}'
    def stop(self): self.up.shutdown();self.up.server_close();self.t.join()
    def adapter(self,vendor='openai'):
        return Provider(vendor,CANARY,GEMINI_TARGET if vendor=='gemini' else 'fixture',fixture_origin=self.origin)
    def request(self, adapter, path=None, body=None, exception=False):
        vendor=adapter.vendor
        route={'anthropic':'/v1/messages','openai':'/v1/chat/completions','gemini':'/v1beta/'+GEMINI_TARGET+':streamGenerateContent?alt=sse'}[vendor]
        payload={'model':'fixture','max_tokens':4096} if vendor!='gemini' else {'generationConfig':{'maxOutputTokens':4096}}
        r=Relay(adapter.origin,adapter.headers,Budget(self.root/'budget',vendor),self.root,['A1'],provider=adapter)
        t=threading.Thread(target=r.serve_forever);t.start();output=io.StringIO()
        try:
            with contextlib.redirect_stderr(output),contextlib.redirect_stdout(output):
                # Patch only outbound request, preserving the real local client socket.
                original=http.client.HTTPConnection.request
                def call(conn,*args,**kw):
                    if conn.port==self.up.server_port and exception: raise OSError('secret URL https://bad/'+CANARY)
                    return original(conn,*args,**kw)
                with patch.object(http.client.HTTPConnection,'request',call):
                    c=http.client.HTTPConnection('127.0.0.1',r.server_port,timeout=3)
                    c.request('POST',path or '/A1'+route,body or json.dumps(payload).encode(),{'Authorization':'Bearer INBOUND-SYNTHETIC-ONLY'})
                    response=c.getresponse();result=(response.status,repr(response.getheaders()).encode()+response.read());c.close()
        finally:r.shutdown();r.server_close();t.join()
        combined=result[1]+output.getvalue().encode()+b''.join(p.read_bytes() for p in self.root.iterdir() if p.is_file())
        self.assertNotIn(CANARY.encode(),combined)
        self.assertNotIn(b'INBOUND-SYNTHETIC-ONLY',combined)
        return result
    def test_success_refusal_redirect_headers_all_routes(self):
        for vendor in FIELDS:
            for status in (200,401,307):
                with self.subTest(vendor=vendor,status=status):
                    # Each isolated run is a different temporary fixture, never a reset live budget.
                    with tempfile.TemporaryDirectory() as td:
                        old=self.root;self.root=Path(td);self.status=status
                        encoded=''.join('%%%02X'%x for x in CANARY.encode())
                        self.response=('prefix '+CANARY+' '+encoded+' suffix').encode()
                        self.assertEqual(self.request(self.adapter(vendor))[0],status)
                        self.assertEqual(len((self.root/'budget').read_text().splitlines()),1)
                        row=json.loads((self.root/'transport.jsonl').read_text().splitlines()[-1]);self.assertTrue(row['credential_redacted'])
                        self.assertEqual(row['outcome'],'completed' if status==200 else 'http_failure')
                        self.root=old
        self.assertEqual(len(self.seen),9) # no redirect follow or retries
        for h in self.credentials:self.assertNotIn('INBOUND-SYNTHETIC-ONLY',repr(h))
    def test_transport_exception_has_no_secret_text_and_spends_one(self):
        self.assertEqual(self.request(self.adapter(),exception=True)[0],502)
        self.assertEqual(self.seen,[]);self.assertEqual(len((self.root/'budget').read_text().splitlines()),1)
    def test_url_and_body_secret_refuse_before_record_or_send(self):
        self.assertEqual(self.request(self.adapter(),path='/A1/v1/chat/completions?key='+CANARY)[0],400)
        self.assertFalse((self.root/'budget').exists())
        self.assertEqual(self.request(self.adapter(),body=json.dumps({'model':'fixture','max_tokens':4096,'secret':CANARY}).encode())[0],429)
        self.assertFalse((self.root/'budget').exists());self.assertEqual(self.seen,[])
    def test_cross_provider_and_model_paths_refuse(self):
        self.assertEqual(self.request(self.adapter(),path='/A1/v1/messages')[0],429)
        self.assertFalse((self.root/'budget').exists())
        self.assertEqual(self.request(self.adapter(),body=b'{"model":"different","max_tokens":4096}')[0],429)
        self.assertFalse((self.root/'budget').exists())
    def test_discovery_one_page_then_refusal(self):
        self.response=b'{"data":[{"id":"fixture"}],"has_more":true}'
        b=Budget(self.root/'budget','openai');out=self.root/'discovery.json'
        d=self.adapter().discover(b,out,'binding')
        self.assertEqual(d['models'],['fixture']);self.assertTrue(d['has_more']);self.assertEqual(len(self.seen),1)
        with self.assertRaisesRegex(ValueError,'discovery ceiling'):self.adapter().discover(b,self.root/'second.json','binding')
        self.assertFalse((self.root/'second.json').exists())
        with self.assertRaisesRegex(ValueError,'bound live discovery'):require_discovery(d,'openai','fixture','binding')
    def test_discovery_failure_redacts_and_never_follows_redirect(self):
        self.status=307;self.response=('https://bad/'+CANARY).encode()
        d=self.adapter().discover(Budget(self.root/'budget','openai'),self.root/'discovery.json','binding')
        self.assertEqual(len(self.seen),1);self.assertEqual(d['models'],[])
        self.assertNotIn(CANARY.encode(),b''.join(p.read_bytes() for p in self.root.iterdir() if p.is_file()))
    def test_key_schema_only_synthetic_file(self):
        path=self.root/'synthetic-settings.json'
        path.write_text(json.dumps({**{field:CANARY for field in FIELDS.values()},'unused':'DO-NOT-RETURN'}))
        for vendor in FIELDS:self.assertEqual(load_key(vendor,settings=path),CANARY)
        for content in ('{}','not JSON','{"directOpenAIAPIKey":123}','{"directOpenAIAPIKey":"line\\nkey"}'):
            path.write_text(content)
            with self.assertRaisesRegex(ValueError,'required API-key field unavailable or invalid'):load_key('openai',settings=path)
    def test_fixed_origins_and_exact_gemini_target(self):
        for vendor in FIELDS:self.assertEqual(Provider(vendor,CANARY).origin,ORIGINS[vendor])
        for origin in ('https://api.openai.com','http://localhost:123','http://127.0.0.1:123/?key='+CANARY):
            with self.assertRaisesRegex(ValueError,'literal loopback'):Provider('openai',CANARY,fixture_origin=origin)
        with self.assertRaisesRegex(ValueError,'Gemini target'):Provider('gemini',CANARY,'models/gemini-2.0-flash')
    def test_absolute_deadline_interrupts_header_drip(self):
        owner=self
        class Drip(http.server.BaseHTTPRequestHandler):
            def log_message(self,*args): pass
            def do_GET(self):
                owner.seen.append(self.path)
                try:
                    self.wfile.write(b'HTTP/1.0 200 OK\r\n')
                    for _ in range(100):
                        self.wfile.write(b'X-Progress: alive\r\n');self.wfile.flush();time.sleep(.01)
                    self.wfile.write(b'\r\n{}')
                except OSError: pass
        server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Drip)
        thread=threading.Thread(target=server.serve_forever);thread.start()
        try:
            adapter=Provider('openai',CANARY,fixture_origin=f'http://127.0.0.1:{server.server_port}')
            start=time.monotonic()
            receipt=adapter.discover(Budget(self.root/'budget','openai'),self.root/'drip.json','binding',timeout=.06)
            self.assertLess(time.monotonic()-start,.4)
            self.assertNotEqual(receipt['outcome'],'completed');self.assertEqual(len(self.seen),1)
        finally:server.shutdown();server.server_close();thread.join()
    def test_resolution_timeout_cannot_send_late_request(self):
        gate=threading.Event();original=socket.getaddrinfo
        def delayed(*args,**kwargs): gate.wait(1);return original(*args,**kwargs)
        with patch('provider.socket.getaddrinfo',delayed):
            start=time.monotonic()
            d=self.adapter().discover(Budget(self.root/'budget','openai'),self.root/'dns.json','binding',timeout=.03)
            self.assertLess(time.monotonic()-start,.3);self.assertNotEqual(d['outcome'],'completed')
            gate.set();time.sleep(.06)
        self.assertEqual(self.seen,[]);self.assertEqual(len((self.root/'budget').read_text().splitlines()),1)

    def test_tls_receives_remaining_connect_budget(self):
        from provider import connection
        timeouts=[]
        class Socket:
            def settimeout(self,seconds): timeouts.append(seconds)
            def connect(self,endpoint): time.sleep(.02)
            def close(self): pass
        with patch('provider.socket.getaddrinfo',return_value=[(socket.AF_INET,socket.SOCK_STREAM,0,'',('127.0.0.1',443))]),patch('provider.socket.socket',return_value=Socket()):
            conn=connection(http.client.HTTPSConnection,'fixture.invalid',443,.1)
            conn._create_connection(('fixture.invalid',443),.1)
        self.assertEqual(len(timeouts),2);self.assertLess(timeouts[1],timeouts[0]-.015)

    def test_bound_discovery_model_and_budget_controls(self):
        budget=Budget(self.root/'budget','openai');item=budget.reserve('models','discovery')
        # Synthetic selection control, never reported as real discovery.
        d=dict(mode='live',vendor='openai',origin=ORIGINS['openai'],path='/v1/models',
               binding_sha256='binding',outcome='completed',status=200,models=['fixture'],
               budget_path=str(budget.path.resolve()),attempt=item)
        require_discovery(d,'openai','fixture','binding',budget.path)
        for key,value in (('mode','local'),('vendor','gemini'),('binding_sha256','other'),('models',[]),('status',401),('budget_path','wrong')):
            altered=copy.deepcopy(d);altered[key]=value
            with self.subTest(key=key),self.assertRaises(ValueError):require_discovery(altered,'openai','fixture','binding',budget.path)
        with self.assertRaisesRegex(ValueError,'Gemini target'):require_discovery(d,'gemini','models/gemini-2.0-flash','binding')

    def test_split_redactor_every_boundary_and_unchanged_clean(self):
        variants=[CANARY.encode(),''.join('%%%02X'%x for x in CANARY.encode()).encode(),''.join('\\u%04x'%ord(x) for x in CANARY).encode()]
        for token in variants:
            for split in range(len(token)+1):
                r=Redactor([CANARY]);actual=r.feed(b'prefix'+token[:split])+r.feed(token[split:]+b'suffix',True)
                self.assertEqual(actual,b'prefix[redacted credential]suffix')
        r=Redactor([CANARY]);self.assertEqual(r.feed(b'{"normal": 1.0}\n',True),b'{"normal": 1.0}\n')


class SplitIdentityControls(IdentityControls):
    # Reuse the valid complete local parent controls, never weakening build maps.
    def setUp(self):
        super().setUp()
        self.old=copy.deepcopy(self.blobs)
        self.blobs[ident.SUPPORT_PREFIX+'driver.py']=b'# updated interpreted support'
        self.blobs[ident.SUPPORT_PREFIX+'provider.py']=b'# new interpreted adapter'
        for path in (ident.SUPPORT_PREFIX+'driver.py',ident.SUPPORT_PREFIX+'provider.py'):
            (self.root/path).write_bytes(self.blobs[path]);(self.support/path.removeprefix(ident.SUPPORT_PREFIX)).write_bytes(self.blobs[path])
        # Replace the two synthetic historical lookup seams with revision-aware ones.
        self.patches[0].stop();self.patches[1].stop()
        self.lookup=patch.object(ident,'source_paths',side_effect=lambda repo,rev:sorted(self.old if rev=='1'*40 else self.blobs));self.lookup.start();self.addCleanup(self.lookup.stop)
        self.history=patch.object(ident,'historical',side_effect=lambda repo,rev,path:(self.old if rev=='1'*40 else self.blobs)[path]);self.history.start();self.addCleanup(self.history.stop)
        self.b['build_sources']=copy.deepcopy(self.sources);self.b['support_revision']='2'*40
        self.sources={p:ident.digest(v) for p,v in self.blobs.items()};self.b['sources']=self.sources
        self.b['support']={p:h for p,h in self.sources.items() if p.startswith(ident.SUPPORT_PREFIX)}
        self.launch=launch_record(self.b,'anthropic','fixture','cli',[str(self.bin),'chat'],{'LLM_VENDOR':'anthropic','LLM_MODEL':'fixture'},self.root,self.root)
    def test_live_launch_keeps_key_out_of_child_and_receipt(self):
        root=self.root/'provider';root.mkdir()
        item=Budget(root/'attempts.jsonl','openai').reserve('models','discovery')
        receipt=dict(mode='live',vendor='openai',origin=ORIGINS['openai'],path='/v1/models',
                     binding_sha256=ident.digest(ident.canonical(self.b)),outcome='completed',status=200,models=['fixture'],
                     budget_path=str((root/'attempts.jsonl').resolve()),attempt=item)
        discovery=self.root/'selection.json';discovery.write_text(json.dumps(receipt))
        args=argparse.Namespace(vendor='openai',model='fixture',step='A1',root=root,run=self.root/'run-live-control',mode='live',origin=None,discovery=discovery,catalog=self.support/'catalog')
        captured=[]
        class IdleRelay:
            server_port=12345
            def __init__(self,*args,**kwargs): self.provider=kwargs['provider']
            def serve_forever(self): pass
            def shutdown(self): pass
            def server_close(self): pass
        class Process:
            pid=1
            def wait(self,**kw): return 0
            def poll(self): return 0
        def spawn(command,**kw): captured.append((command,kw));return Process()
        with patch('driver.load_key',return_value=CANARY) as key,patch('driver.Relay',IdleRelay),patch('driver.subprocess.Popen',spawn),contextlib.redirect_stdout(io.StringIO()):
            self.assertEqual(launch(args,self.b),0);key.assert_called_once_with('openai')
        self.assertEqual(captured[0][1]['env']['LLM_API_KEY'],'local-relay-placeholder')
        self.assertNotIn(CANARY,repr(captured))
        self.assertNotIn(CANARY.encode(),b''.join(f.read_bytes() for f in args.run.iterdir() if f.is_file()))
        recorded=json.loads((args.run/'launch.json').read_text());ident.check_launch(self.b,recorded)
        self.assertEqual(recorded['support_revision'],'2'*40)
        # A local/fabricated-selection marker cannot open credentials or create a run.
        receipt['mode']='local';discovery.write_text(json.dumps(receipt));args.run=self.root/'refused'
        with patch('driver.load_key',side_effect=AssertionError('must not read settings')):
            with self.assertRaisesRegex(ValueError,'bound live discovery'):launch(args,self.b)
        self.assertFalse(args.run.exists())

    def test_split_specific_refusals(self):
        self.check()
        for field,mutation,message in (
            ('support revision',lambda b:b.__setitem__('support_revision','1'*40),'incomplete source'),
            ('build source',lambda b:b['build_sources'].__setitem__(ident.PREFIX+'main.go','0'*64),'historical build source'),
            ('missing build source',lambda b:b['build_sources'].pop(ident.PREFIX+'main.go'),'incomplete historical build'),
        ):
            with self.subTest(field=field):
                b=copy.deepcopy(self.b);mutation(b)
                with self.assertRaisesRegex(ValueError,message):self.check(b)
        launch=copy.deepcopy(self.launch);launch.pop('support_revision')
        with self.assertRaisesRegex(ValueError,'launch support revision'):ident.check_launch(self.b,launch)
        path=ident.SUPPORT_PREFIX+'consumer/new.go';self.blobs[path]=b'package main'
        f=self.root/path;f.parent.mkdir(parents=True);f.write_bytes(self.blobs[path])
        self.b['sources'][path]=ident.digest(self.blobs[path])
        with self.assertRaisesRegex(ValueError,'compiled source requires a new build'):self.check()


if __name__=='__main__':unittest.main(verbosity=2)
