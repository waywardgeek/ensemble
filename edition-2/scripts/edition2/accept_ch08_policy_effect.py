#!/usr/bin/env python3
"""Chapter 8 policy consumption through actual GUI/Agent/tools and local model HTTP.

Tests captured limits and complete final batches, not merely settings echoes.
No paid model, public-headless API, persistence fault or browser claim is made.
"""
import argparse
import contextlib
import hashlib
import http.server
import json
import os
from pathlib import Path
import subprocess
import tempfile
import threading
from accept_ch07 import Socket
from accept_ch08 import Program, defaults, subscribe, update


class Model(http.server.BaseHTTPRequestHandler):
    def log_message(self, *_):
        pass

    def do_POST(self):
        request = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
        messages = request['messages']
        users = [m['content'] for m in messages if m.get('role') == 'user']
        label = users[-1]
        if isinstance(label, list):
            label = ''.join(p.get('text', '') for p in label)
        assert label in self.server.counts, 'unexpected human fixture identity'
        with self.server.lock:
            self.server.counts[label] += 1
            number = self.server.counts[label]
            self.server.requests.append(request)
        if label == 'ACTIVE' and number == 1:
            self.server.entered.set()
            if not self.server.release.wait(10):
                self.send_error(500, 'fixture barrier not released')
                return
        calls = [dict(id=f'{label}-{number}-{part}', type='function', function=dict(name='write_file', arguments=json.dumps(dict(path=f'{label}-{number}-{part}.txt', content=f'{label}:{number}:{part}\n')))) for part in (1, 2)]
        terminal = label == 'NORMAL'
        message = dict(role='assistant', content='NORMAL-DONE') if terminal else dict(role='assistant', content=None, tool_calls=calls)
        finish = 'stop' if terminal else 'tool_calls'
        usage = dict(prompt_tokens=1, completion_tokens=1, total_tokens=2)
        self.send_response(200)
        self.send_header('Content-Type', 'text/event-stream' if request.get('stream') else 'application/json')
        self.end_headers()
        if request.get('stream'):
            delta = dict(role='assistant', content='NORMAL-DONE') if terminal else dict(role='assistant', tool_calls=[dict(index=i, **call) for i, call in enumerate(calls)])
            chunks = [dict(id='fixture', model='fixture-ch08-independent', choices=[dict(index=0, delta=delta, finish_reason=None)]), dict(id='fixture', model='fixture-ch08-independent', choices=[dict(index=0, delta={}, finish_reason=finish)]), dict(id='fixture', model='fixture-ch08-independent', choices=[], usage=usage)]
            self.wfile.write(b''.join(b'data: '+json.dumps(c).encode()+b'\n\n' for c in chunks)+b'data: [DONE]\n\n')
        else:
            self.wfile.write(json.dumps(dict(id='fixture', model='fixture-ch08-independent', choices=[dict(index=0, message=message, finish_reason=finish)], usage=usage)).encode())
        self.wfile.flush()


def check_effect(directory, label, maximum, requests):
    assert requests == maximum, f'{label}: HTTP count {requests}, expected {maximum}'
    for number in range(1, maximum+1):
        for part in (1, 2):
            path = directory/f'{label}-{number}-{part}.txt'
            assert path.read_text() == f'{label}:{number}:{part}\n', 'last accepted batch did not finish both actual effects'
    assert not (directory/f'{label}-{maximum+1}-1.txt').exists(), 'effect after captured request limit'


def completion(client, request_id, outcome):
    match = lambda x: x.get('type') == 'completion' and x.get('request_id') == request_id
    # Independent completion-delivery workers may expose the later turn's
    # completion first even though actor activation is serial. Correlate by ID.
    record = next((x for x in client.records if match(x)), None)
    if record is None:
        record = client.until(match)
    assert record.get('outcome') == outcome, 'wrong reliable completion outcome'
    return record


def submit(client, label):
    client.send(dict(type='prompt', id='prompt-'+label, text=label))
    reply = client.until(lambda x: x.get('id') == 'prompt-'+label)
    assert reply.get('type') == 'accepted', 'fixture prompt not accepted'
    return reply['request_id']


def model_server():
    server = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Model)
    server.requests = []
    server.counts = {label: 0 for label in ('ONE', 'SEVENTEEN', 'DEFAULT', 'NORMAL', 'ACTIVE', 'QUEUED')}
    server.lock = threading.Lock()
    server.entered = threading.Event()
    server.release = threading.Event()
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    return server, thread


def fixture_preflight(binary):
    """Prove the new local model/tool fixture against an accepted predecessor.

    This exercises no Chapter 8 settings and cannot credit their runtime gate.
    """
    rows = []
    for plain in (False, True):
        for label in ('NORMAL', 'DEFAULT'):
            server, thread = model_server()
            try:
                with tempfile.TemporaryDirectory(prefix='ch08-model-positive-') as temporary:
                    directory = Path(temporary)
                    env = {k:v for k,v in os.environ.items() if not k.startswith(('LLM_', 'CH02_', 'COURSE_', 'OPENAI_', 'ANTHROPIC_', 'GEMINI_', 'EN_DISABLE_STREAMING'))}
                    env.update(LLM_VENDOR='openai', LLM_MODEL='fixture-ch08-independent',
                               LLM_RESOLVED_MODEL='fixture-ch08-independent', LLM_API_KEY='LOCAL-ONLY',
                               LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',
                               CH02_LOG=str(directory/'events.log'), EN_DISABLE_STREAMING='1' if plain else '0')
                    result = subprocess.run([str(binary), 'protocol'],
                                            input=json.dumps(dict(kind='prompt', text=label))+'\n',
                                            cwd=directory, env=env, text=True, capture_output=True, timeout=30)
                    output = [json.loads(line) for line in result.stdout.splitlines()]
                    completed = [r['completion'] for r in output if 'completion' in r]
                    assert len(completed) == 1, 'predecessor did not complete fixture turn'
                    expected = 'success' if label == 'NORMAL' else 'round_limit'
                    assert completed[0]['outcome'] == expected, completed[0]
                    assert result.returncode == (0 if label == 'NORMAL' else 1), result.stderr
                    assert all(bool(r.get('stream')) != plain for r in server.requests)
                    if label == 'NORMAL':
                        assert server.counts[label] == 1 and completed[0]['text'] == 'NORMAL-DONE'
                    else:
                        check_effect(directory, label, 16, server.counts[label])
                    rows.append(dict(id=f'{label}-'+('plain' if plain else 'stream'), passed=True,
                                     http=server.counts[label], completion=completed[0]['outcome']))
            except Exception as error:
                rows.append(dict(id=f'{label}-'+('plain' if plain else 'stream'), passed=False, details=str(error)))
            finally:
                server.release.set()
                server.shutdown()
                server.server_close()
                thread.join(timeout=2)
    return dict(scope=fixture_preflight.__doc__, passed=all(r['passed'] for r in rows), checks=rows,
                binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                checker_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest())


def evaluate(binary):
    rows = []
    server, thread = model_server()
    def run(name, action):
        try:
            action()
            rows.append(dict(id=name, passed=True))
            return True
        except Exception as error:
            rows.append(dict(id=name, passed=False, details=type(error).__name__+': '+str(error)))
            return False
    try:
        with tempfile.TemporaryDirectory(prefix='ch08-policy-effect-') as temporary:
            directory = Path(temporary)
            endpoint = f'http://127.0.0.1:{server.server_port}'
            stage = 'startup'
            try:
                with contextlib.closing(Program(binary, directory, endpoint, 1)) as program:
                    url = program.listen()
                    with contextlib.closing(Socket(url)) as client:
                        subscribe(client, defaults(), 0, 0, 0)
                        revision = 0
                        for label, stored, effective in [('ONE', 1, 1), ('SEVENTEEN', 17, 17), ('DEFAULT', 0, 16), ('NORMAL', 1, 1)]:
                            stage = 'captured-limit-'+label.lower()
                            def limit_case(label=label, stored=stored, effective=effective):
                                update(client, 'policy', 'limit-'+label, revision, dict(max_model_requests=stored), revision+1, stored)
                                request_id = submit(client, label)
                                completion(client, request_id, 'success' if label == 'NORMAL' else 'round_limit')
                                if label == 'NORMAL':
                                    assert server.counts[label] == 1, 'normal answer issued extra HTTP'
                                else:
                                    check_effect(directory, label, effective, server.counts[label])
                            if not run(stage, limit_case):
                                break
                            revision += 1
                        else:
                            stage = 'active-and-queued-policy-capture'
                            def capture():
                                active = submit(client, 'ACTIVE')
                                assert server.entered.wait(4), 'positive held HTTP barrier never entered'
                                queued = submit(client, 'QUEUED')
                                update(client, 'policy', 'during-http', revision, dict(max_model_requests=2), revision+1, 2)
                                # Read through a fresh public watch while HTTP is
                                # held, without manufacturing a second begin.
                                with contextlib.closing(Socket(url)) as watch:
                                    watch.send(dict(type='subscribe', id='held'))
                                    watch.next()  # preferences domain precedes Agent
                                    begin = watch.next()
                                    assert begin['state']['active_max_model_requests'] == 1, 'active capture changed with next-turn policy'
                                    assert begin['state']['execution_policy']['max_model_requests'] == 2
                                    assert queued in begin['state']['queued_request_ids']
                                server.release.set()
                                completion(client, active, 'round_limit')
                                completion(client, queued, 'round_limit')
                                check_effect(directory, 'ACTIVE', 1, server.counts['ACTIVE'])
                                check_effect(directory, 'QUEUED', 2, server.counts['QUEUED'])
                            run(stage, capture)
            except Exception as error:
                rows.append(dict(id=stage, passed=False, details=type(error).__name__+': '+str(error)))
            finally:
                server.release.set()
            if len(rows) == 5 and all(r['passed'] for r in rows):
                def records():
                    events = [json.loads(line) for line in (directory/'events-1.log').read_text().splitlines()]
                    started = [e for e in events if e.get('type') == 'turn_started']
                    assert len(started) == 6, 'missing recorded policy captures'
                    policies = [e['turn']['policy'] for e in started]
                    assert [(p['revision'], p['max_model_requests'], p['effective_max_model_requests']) for p in policies] == [(1,1,1),(2,17,17),(3,0,16),(4,1,1),(4,1,1),(5,2,2)], 'durable policy differs from activation capture'
                    called = [e['tool']['call_id'] for e in events if e.get('type') == 'tool_called']
                    returned = [e['tool']['call_id'] for e in events if e.get('type') == 'tool_returned']
                    assert called == returned and len(called) == 2*(1+17+16+1+2), 'accepted final batch pairing differs'
                run('durable-capture-and-final-batch-pairing', records)
    finally:
        server.release.set()
        server.shutdown()
        server.server_close()
        thread.join(timeout=2)
    return dict(scope=__doc__, passed=len(rows) == 6 and all(r['passed'] for r in rows), checks=rows, http_by_prompt=server.counts,
                binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(), checker_files={name: hashlib.sha256(Path(__file__).with_name(name).read_bytes()).hexdigest() for name in ['accept_ch08_policy_effect.py', 'accept_ch08.py', 'accept_ch07.py']})


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', type=Path)
    parser.add_argument('--fixture-cli', action='store_true', help='Validate only the local model/tool fixture against an accepted CLI')
    args = parser.parse_args()
    result = (fixture_preflight if args.fixture_cli else evaluate)(args.binary.resolve(strict=True))
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
