"""Offline tests only: synthetic credentials and a loopback HTTP fixture."""
import contextlib
import importlib.util
import io
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from unittest.mock import patch

SOURCE = Path(__file__).with_name('moltbook_server.py')
spec = importlib.util.spec_from_file_location('moltbook', SOURCE)
m = importlib.util.module_from_spec(spec)
spec.loader.exec_module(m)


class Tests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.creds = Path(self.temp.name) / 'credentials.json'
        self.creds.write_text(json.dumps({'api_key': 'synthetic-secret-only'}))
        self.calls = []
        self.response = {'your_account': {'karma': 7}, 'activity_on_your_posts': []}
        self.code = 200
        owner = self
        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                owner.calls.append((self.command, self.path, self.headers.get('Authorization')))
                self.send_response(owner.code)
                if owner.code == 302:
                    self.send_header('Location', '/must-not-follow')
                self.end_headers()
                data = owner.response
                try:
                    self.wfile.write(data if isinstance(data, bytes) else json.dumps(data).encode())
                except (BrokenPipeError, ConnectionResetError):
                    pass  # Bounded client intentionally stops oversized/redirect bodies.
            def do_POST(self):
                self.rfile.read(int(self.headers.get('Content-Length', 0)))
                self.do_GET()
            def log_message(self, *args):
                pass
        self.server = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        self.thread = threading.Thread(target=self.server.serve_forever)
        self.thread.start()
        self.patch = patch.multiple(m, CREDS_PATH=str(self.creds),
            BASE_URL='http://127.0.0.1:' + str(self.server.server_port))
        self.patch.start()

    def tearDown(self):
        self.patch.stop()
        self.server.shutdown()
        self.thread.join()
        self.server.server_close()
        self.temp.cleanup()

    def call(self, name, args=None):
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            m.handle({'jsonrpc':'2.0', 'id':1, 'method':'tools/call',
                      'params':{'name':name, 'arguments':args or {}}})
        return json.loads(output.getvalue())

    def test_status_real_http(self):
        result = self.call('moltbook_status')['result']
        self.assertEqual(json.loads(result['content'][0]['text'])['karma'], 7)
        self.assertEqual(self.calls, [('GET', '/home', 'Bearer synthetic-secret-only')])

    def test_errors_never_become_empty_success_or_echo_secret(self):
        for name in ['moltbook_status','moltbook_feed','moltbook_read_post',
                     'moltbook_comments','moltbook_search']:
            self.code = 401
            self.response = {'error':'synthetic-secret-only reflected body'}
            args = {'post_id':'id'} if name in ('moltbook_read_post','moltbook_comments') else {}
            if name == 'moltbook_search': args = {'query':'query'}
            result = self.call(name, args)
            self.assertTrue(result['result']['isError'])
            self.assertNotIn('synthetic-secret-only', json.dumps(result))
            self.assertIn('401', json.dumps(result))

    def test_json_failure_bounds_and_redirect(self):
        for response in [b'not json', [], {'success':False}, {'error':'private'}, b'x'*(m.MAX_RESPONSE_BYTES+1)]:
            self.response = response
            self.assertTrue(self.call('moltbook_feed')['result']['isError'])
        self.code = 302
        before = len(self.calls)
        self.assertTrue(self.call('moltbook_status')['result']['isError'])
        self.assertEqual(len(self.calls), before+1)
        self.assertNotIn('/must-not-follow', [c[1] for c in self.calls])

    def test_validation_and_url_components(self):
        for args in [{'limit':0},{'limit':True},{'limit':101},{'extra':'x'}]:
            self.assertTrue(self.call('moltbook_feed',args)['result']['isError'])
        self.assertEqual(self.calls, [])
        self.response = {'post':{'content':'synthetic'}}
        self.call('moltbook_read_post', {'post_id':'a/b?secret#fragment'})
        self.assertEqual(self.calls[-1][1], '/posts/a%2Fb%3Fsecret%23fragment')

    def test_credentials_and_exception_masking(self):
        self.creds.write_text('synthetic-secret-only invalid json')
        result = self.call('moltbook_status')
        self.assertTrue(result['result']['isError'])
        self.assertNotIn('synthetic-secret-only', json.dumps(result))
        self.assertEqual(self.calls, [])
        with patch.object(m, 'api_request', side_effect=RuntimeError('synthetic-secret-only')):
            self.assertNotIn('synthetic-secret-only', json.dumps(self.call('moltbook_status')))

    def test_writes_pending_and_no_retry(self):
        self.response = {'comment':{'id':'new','verification_status':'pending',
            'verification':{'challenge_text':'fixture only','verification_code':'code'}}}
        result = self.call('moltbook_comment', {'post_id':'id','content':'Synthetic'})
        text = json.loads(result['result']['content'][0]['text'])
        self.assertEqual(text['status'], 'pending')
        self.assertEqual(len(self.calls), 1)
        self.code = 500
        self.assertTrue(self.call('moltbook_comment', {'post_id':'id','content':'Synthetic'})['result']['isError'])
        self.assertEqual(len(self.calls), 2)
        self.code = 200
        self.response = {'comment': {}}
        uncertain = self.call('moltbook_comment', {'post_id':'id','content':'Synthetic'})
        self.assertTrue(uncertain['result']['isError'])
        self.assertIn('outcome unknown', json.dumps(uncertain))
        self.assertEqual(len(self.calls), 3)

    def test_discovery_subprocess_without_credentials(self):
        requests = [
            {'jsonrpc':'2.0','id':1,'method':'initialize'},
            {'jsonrpc':'2.0','method':'notifications/initialized'},
            {'jsonrpc':'2.0','id':2,'method':'tools/list'},
            {'jsonrpc':'2.0','method':'$/cancelRequest','params':{'id':99}}]
        result = subprocess.run([sys.executable,str(SOURCE),'--credentials',str(self.creds)+'.missing'],
            input=''.join(json.dumps(r)+'\n' for r in requests), text=True,
            capture_output=True, timeout=10, check=True)
        replies = [json.loads(line) for line in result.stdout.splitlines()]
        self.assertEqual(len(replies), 2)
        self.assertEqual(len(replies[1]['result']['tools']), 12)
        self.assertEqual(self.calls, [])

    def test_process_stop_interrupts_blocked_call(self):
        # Inject a blocking API at the module boundary; no network or real key.
        wrapper = Path(self.temp.name)/'blocked.py'
        entered = Path(self.temp.name)/'entered'
        wrapper.write_text("import sys, time\nsys.path.insert(0,"+repr(str(SOURCE.parent))+ ")\nimport moltbook_server as m\ndef blocked(*a,**k):\n open("+repr(str(entered))+",'w').close()\n time.sleep(60)\nm.api_request=blocked\nm.main()\n")
        process = subprocess.Popen([sys.executable,str(wrapper)], stdin=subprocess.PIPE,
            stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        try:
            process.stdin.write(json.dumps({'id':1,'method':'tools/call','params':{'name':'moltbook_status'}})+'\n')
            process.stdin.flush()
            deadline = time.monotonic() + 5
            while not entered.exists() and time.monotonic() < deadline:
                time.sleep(.01)
            self.assertTrue(entered.exists(), "call never entered")
            process.stdin.write(json.dumps({'method':'$/cancelRequest','params':{'id':1}})+'\n')
            process.stdin.flush()
            with self.assertRaises(subprocess.TimeoutExpired):
                process.wait(timeout=.1)
            process.terminate()
            process.communicate(timeout=5)
            self.assertIsNotNone(process.returncode)
        finally:
            if process.poll() is None:
                process.kill();process.communicate(timeout=5)

if __name__ == '__main__':
    unittest.main()
