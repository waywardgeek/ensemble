"""Exercise the published scratch setup with real tools and captured external SSE."""
import hashlib
import http.server
import json
import os
import pathlib
import re
import shutil
import subprocess
import tempfile
import threading
import time
import urllib.request

root = pathlib.Path(os.environ['CH19_ROOT'])
evidence = root / 'solutions/edition-2/evidence/ch20'
out = evidence / 'enhancement19'
chapter = root / 'book/edition-2/chapter-19.md'
sha = hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha == '1058dc2c2e68e73c2ae0c14c7205a5bf88717bd8f1c4fec614cca3182a442b09'
blocks = re.findall(r'```([^\n]*)\n(.*?)```', chapter.read_text(), re.S)
# These are actual provider events retained from the accepted paid run. Replaying
# them supplies external decisions; the submitted CLI still performs every tool.
streams = []
for line in (evidence / 'live-api.log').read_text().splitlines():
    record = json.loads(line)
    if record['event'] == 'request':
        streams.append([])
    else:
        streams[-1].append(record)
requests, headers, receipts, errors, judges = [], [], [], [], []

class Provider(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def do_GET(self):
        # The identifier/display pair was actually listed in the retained signed
        # account catalog. The external fixture uses the existing models shape.
        assert self.path == '/v1/models'
        headers.append(self.headers.get('Authorization', ''))
        catalog = json.loads((root / 'solutions/edition-2/evidence/ch19/live-plan-models.json').read_text())
        assert 'gpt-6-astra\tGPT-6-Astra' in catalog['stdout']
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps({'models': [{'slug': 'gpt-6-astra', 'display_name': 'GPT-6-Astra', 'visibility': 'list'}]}).encode())

    def do_POST(self):
        try:
            q = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
            if not q.get('tools'):
                # The newly published primary is searchable. Its recall judge
                # may run before the main turn; it receives a valid empty pick.
                judges.append(q)
                self.send_response(200)
                self.send_header('Content-Type', 'text/event-stream')
                self.end_headers()
                response = dict(status='completed', model=q['model'], usage=dict(input_tokens=100, output_tokens=1, input_tokens_details=dict(cached_tokens=0)), output=[dict(type='message', role='assistant', content=[dict(type='output_text', text='[]')])])
                self.wfile.write(('event: response.completed\ndata: '+json.dumps(dict(type='response.completed', response=response))+'\n\n').encode())
                self.wfile.flush()
                return
            requests.append(q)
            headers.append(self.headers.get('Authorization', ''))
            assert self.path == '/v1/responses' and len(requests) <= 5
            assert q['model'] == 'gpt-6-astra'
            assert q['reasoning'] == {'effort': 'high', 'summary': 'detailed'}
            assert q['stream'] is True and q['store'] is False
            assert 'prompt_cache_options' not in q and 'prompt_cache_breakpoint' not in json.dumps(q)
            self.send_response(200)
            self.send_header('Content-Type', 'text/event-stream')
            self.end_headers()
            for record in streams[len(requests)-1]:
                self.wfile.write(('event: '+record['event']+'\ndata: '+record['data']+'\n\n').encode())
                self.wfile.flush()
        except Exception as error:
            errors.append(repr(error))
            raise

server = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Provider)
threading.Thread(target=server.serve_forever, daemon=True).start()
work = None

def run(command, cwd, env=None, input=None, expected_exit=0):
    result = subprocess.run(['bash', '-c', command], cwd=cwd, env=env,
                            input=input, capture_output=True, text=True, timeout=30)
    receipts.append(dict(command=command, exit=result.returncode,
                         stdout=result.stdout, stderr=result.stderr))
    assert result.returncode == expected_exit, receipts[-1]
    return result.stdout

try:
    assert len(blocks) == 5 and [x[0] for x in blocks] == ['', 'sh', 'sh', 'text', 'sh']
    env = {k: v for k, v in os.environ.items()
           if not k.endswith('API_KEY') and not k.startswith(('LLM_', 'OPENAI_', 'EN_', 'ENSEMBLE_', 'CH02_', 'CH07_'))}
    work = pathlib.Path(run(blocks[1][1]+'\npwd', root, env).strip())
    env.update(EN_HOME=str(work/'profile'), OPENAI_OIDC_ISSUER=os.environ['CH19_ISSUER'],
               LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}', chapter19_repo=str(root),
               CH02_LOG=str(work/'events.jsonl'))
    browser = work/'browser.py'
    browser.write_text('import sys,urllib.request\nwith urllib.request.urlopen(sys.argv[1],timeout=5) as r:r.read()\n')
    env['EN_BROWSER_CMD'] = 'python3 '+str(browser)
    run('/tmp/ensemble-ch19 auth login', work, env)
    # Keep stdin open through the final answer as the published recipe says.
    with (work/'stdout.txt').open('w') as stdout, (work/'stderr.txt').open('w') as stderr:
        child = subprocess.Popen(['bash','-c',blocks[2][1]], cwd=work, env=env,
                                 stdin=subprocess.PIPE, stdout=stdout, stderr=stderr, text=True)
        child.stdin.write(blocks[3][1]); child.stdin.flush()
        start = time.monotonic()
        completed = False
        while child.poll() is None and time.monotonic()-start < 25:
            path = work/'events.jsonl'
            if path.exists():
                events = [json.loads(line) for line in path.read_text().splitlines() if line]
                for event in events:
                    parts = event.get('response', {}).get('parts', [])
                    if event.get('type') == 'response_ended' and any(p['type'] == 'text' for p in parts) and not any(p['type'] == 'tool_call' for p in parts):
                        completed = True
            if completed:
                child.stdin.close()
                break
            time.sleep(.02)
        if not completed:
            child.kill()
        exit_code = child.wait(timeout=10)
    receipts.append(dict(command=blocks[2][1], exit=exit_code,
                         stdout=(work/'stdout.txt').read_text(), stderr=(work/'stderr.txt').read_text()))
    assert completed and exit_code == 0 and not errors, receipts[-1]
    assert 'Signed in: true' in receipts[-1]['stdout'] and 'gpt-6-astra\tGPT-6-Astra' in receipts[-1]['stdout']
    assert len(requests) == 5 and len(headers) == 6 and all(x.startswith('Bearer ') for x in headers)
    # Inspect actual correlated outputs, not the model's scripted final report.
    calls = [json.loads(record['data'])['item'] for stream in streams for record in stream
             if record['event'] == 'response.output_item.done' and json.loads(record['data'])['item']['type'] == 'function_call']
    actual = {item['call_id']: item['output'] for q in requests for item in q['input'] if item.get('type') == 'function_call_output'}
    texts = [actual[call['call_id']] for call in calls]
    assert texts[0] == (evidence/'task-before/aggregate.py').read_text()
    assert texts[1] == (evidence/'task-before/test_aggregate.py').read_text()
    assert 'FAILED (failures=3, errors=3)' in texts[2] and 'exit_code: 1' in texts[2]
    assert 'replaced one exact match' in texts[3]
    assert 'Ran 6 tests' in texts[4] and '\nOK\n' in texts[4] and 'exit_code: 0' in texts[4]
    assert (work/'aggregate.py').read_bytes() == (evidence/'task-after/aggregate.py').read_bytes()
    assert (work/'save.json').is_file()
    run(blocks[4][1], work, env)
    (out/'samples-judge-requests.json').write_text(json.dumps(judges, indent=2)+'\n')
    (out/'samples-requests.json').write_text(json.dumps(requests, indent=2)+'\n')
    (out/'samples-run.json').write_text(json.dumps(receipts, indent=2)+'\n')
    (out/'samples.json').write_text(json.dumps(dict(result='PASS', manuscript_sha256=sha,
        fences=[dict(index=i, language=k, sha256=hashlib.sha256(v.encode()).hexdigest()) for i,(k,v) in enumerate(blocks)],
        requests=5, auxiliary_judge_requests=len(judges), completed_before_eof=True, cli_exit=exit_code, save=True,
        source_matches_historical_repair=True, unchanged_six_tests=True,
        launch_substitutions=['Local issuer and provider base URL', 'Isolated fake-auth EN_HOME and browser visitor', 'CH02_LOG capture path'],
        limits='Replayed real external SSE supplies known tool decisions; real frozen CLI performs all reads/edit/commands. This proves sample mechanics, not fresh autonomous coding or current account eligibility. No paid inference, real sign-in, private profile or GUI.'), indent=2)+'\n')
    print('PASS exact sample', sha, 'five real CLI requests; tests unchanged; exit 0')
finally:
    server.shutdown()
    server.server_close()
    if work:
        shutil.rmtree(work)
