#!/usr/bin/env python3
"""Messages zero-byte argument deltas: local CLI effects and terminal barrier.

The recorded SSE positive is replayed from retained bytes; all variants and the
continuation are local fixtures. No model backend or Go compilation is involved.
"""
import argparse
import hashlib
import http.server
import json
from pathlib import Path
import subprocess
import tempfile
import threading
from accept_ch06_clients import Client, fixture, frame

ROOT = Path(__file__).resolve().parents[2]
RECORDED = ROOT / 'solutions/edition-2/main/evidence/ch09/live-anthropic-g/responses/002.body'
RECORDED_SHA = 'e6f7b7446b3dab2fe5ef6466111a6790c31ac52d45eccc8296a416ea9a4971ef'
MARKER = 'AFTER-ARGUMENT-BLOCK'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def synthetic(initial, deltas, terminal=True, error=False):
    values = [dict(type='message_start', message=dict(model='fixture-ch06', usage=dict(input_tokens=2))),
              dict(type='content_block_start', index=0,
                   content_block=dict(type='tool_use', id='write-one', name='write_file', input=initial))]
    values += [dict(type='content_block_delta', index=0,
                    delta=dict(type='input_json_delta', partial_json=value)) for value in deltas]
    values += [dict(type='content_block_stop', index=0),
               dict(type='content_block_start', index=1, content_block=dict(type='text', text=MARKER)),
               dict(type='content_block_stop', index=1),
               dict(type='message_delta', delta=dict(stop_reason='tool_use'), usage=dict(output_tokens=3))]
    prefix = b''.join(frame(value) for value in values)
    suffix = frame(dict(type='error', error=dict(type='fixture_error', message='local refusal'))) if error else frame(dict(type='message_stop')) if terminal else b''
    return prefix, suffix


class Exchange:
    def __init__(self, prefix, suffix, hold):
        self.release = threading.Event()
        self.requests, self.errors = [], []
        if not hold:
            self.release.set()
        owner = self
        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass
            def do_POST(self):
                try:
                    owner.requests.append(json.loads(self.rfile.read(int(self.headers['Content-Length']))))
                    index = len(owner.requests)
                    if index > 2:
                        self.send_error(429, 'local request bound')
                        return
                    self.send_response(200)
                    self.send_header('Content-Type', 'text/event-stream')
                    self.end_headers()
                    if index == 1:
                        self.wfile.write(prefix)
                        self.wfile.flush()
                        if not owner.release.wait(10):
                            raise RuntimeError('test did not release terminal')
                        self.wfile.write(suffix)
                    else:
                        first, last = fixture('anthropic')
                        self.wfile.write(first + last)
                    self.wfile.flush()
                except (BrokenPipeError, ConnectionResetError):
                    pass
                except Exception as error:
                    owner.errors.append(str(error))
        self.http = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        self.http.daemon_threads = True
        self.thread = threading.Thread(target=lambda: self.http.serve_forever(poll_interval=.01), daemon=True)
        self.thread.start()

    def close(self):
        self.release.set()
        self.http.shutdown()
        self.http.server_close()
        self.thread.join()


def check(binary, prefix, suffix, valid, expected_args, hold=False, recorded=False):
    with tempfile.TemporaryDirectory(prefix='ch09-empty-stream-') as directory:
        work = Path(directory)
        server = Exchange(prefix, suffix, hold)
        client = Client(binary, work, server, 'anthropic', False)
        def history():
            return [json.loads(line) for line in (work / 'events.jsonl').read_text().splitlines()]
        try:
            client.send(json.dumps(dict(kind='prompt', text='Exercise this local stream.')) + '\n')
            if hold:
                client.wait(lambda: any(r.get('observation', {}).get('text') == MARKER for r in client.records()),
                            'post-argument marker absent before terminal', 5)
                before = history()
                assert not any(r.get('type') in ('response_ended', 'tool_called') for r in before), 'accepted content before message_stop'
                assert not any('completion' in r or r.get('observation', {}).get('kind') == 'part_final' for r in client.records()), 'final/completion before message_stop'
                assert not (work / 'accepted.txt').exists() and not (work / 'wrong.txt').exists(), 'tool effect before message_stop'
                server.release.set()
            client.wait(lambda: any('completion' in row for row in client.records()), 'request completion absent', 8)
            records, events = client.records(), history()
            outcomes = [row['completion']['outcome'] for row in records if 'completion' in row]
            calls = [e['tool'] for e in events if e.get('type') == 'tool_called']
            responses = [e for e in events if e.get('type') == 'response_ended']
            ends = [r['observation'] for r in records if r.get('observation', {}).get('kind') == 'model_end']
            assert not server.errors, server.errors
            assert not (work / 'wrong.txt').exists(), 'placeholder arguments executed'
            if valid:
                assert outcomes == ['success'], 'valid operation rejected: ' + str(outcomes)
                assert len(server.requests) == len(responses) == len(ends) == 2 and all(e['accepted'] for e in ends)
                assert len(calls) == 1 and calls[0]['args'] == expected_args, 'complete arguments changed'
                returns = [e['tool'] for e in events if e.get('type') == 'tool_returned']
                assert len(returns) == 1 and not returns[0].get('is_error', False), 'valid call did not execute successfully'
                assert returns[0]['call_id'] == calls[0]['call_id']
                args_text = ''.join(r['observation']['text'] for r in records
                                    if r.get('observation', {}).get('kind') == 'part_delta'
                                    and r['observation'].get('channel') == 'tool_args')
                assert json.loads(args_text) == expected_args, 'argument display duplicated or altered'
                if recorded:
                    assert calls[0]['name'] == 'list_directory'
                    assert responses[0]['response']['usage'] == dict(input=2016, output=51, cache_write=0, cache_read=0)
                else:
                    assert calls[0]['name'] == 'write_file' and (work / 'accepted.txt').read_text() == 'review effect'
            else:
                assert outcomes == ['error'], 'invalid operation did not fail'
                assert len(server.requests) == 1 and not calls and not responses, 'failed operation accepted or dispatched'
                assert len(ends) == 1 and ends[0]['accepted'] is False
                assert not any(r.get('observation', {}).get('kind') == 'part_final' for r in records)
                assert not (work / 'accepted.txt').exists(), 'failed stream changed file'
            return dict(outcomes=outcomes, requests=len(server.requests), calls=calls,
                        responses=len(responses), terminal_barrier=hold)
        finally:
            # This protocol child needs no process-group teardown: no shell is
            # launched by the fixture. Closing input lets error exits settle
            # without racing killpg against an already exiting process.
            server.release.set()
            try:
                client.process.stdin.close()
            except BrokenPipeError:
                pass
            try:
                client.process.wait(timeout=2)
            except subprocess.TimeoutExpired:
                client.process.kill()
                client.process.wait(timeout=2)
            client.process.stdout.close()
            client.err.close()
            server.close()


def audit(binary):
    initial = digest(binary)
    assert digest(RECORDED) == RECORDED_SHA, 'recorded response identity changed'
    effect = dict(path='accepted.txt', content='review effect')
    placeholder = dict(path='wrong.txt', content='must not execute')
    encoded = json.dumps(effect)
    cases = [('recorded-empty-object', (RECORDED.read_bytes(), b''), True, {}, False, True)]
    variants = [('no-delta', effect, [], True, True, False, False),
                ('empty', effect, [''], True, True, False, False),
                ('multiple-empty', effect, ['', ''], True, True, False, False),
                ('nonempty-replaces', placeholder, ['', encoded[:10], '', encoded[10:], ''], True, True, False, False),
                ('malformed-not-fallback', effect, ['', '{'], False, True, False, False),
                ('whitespace-not-empty', effect, [' '], False, True, False, False),
                ('null-replacement-not-fallback', effect, ['null'], False, True, False, False),
                ('array-replacement-not-fallback', effect, ['[]'], False, True, False, False),
                ('null-start', None, [''], False, True, False, False),
                ('wrong-delta-type', effect, [None], False, True, False, False),
                ('terminal-held-positive', effect, [''], True, True, False, True),
                ('missing-terminal', effect, [''], False, False, False, True),
                ('provider-error', effect, [''], False, True, True, True)]
    cases += [(name, synthetic(start, deltas, terminal, error), valid, effect, hold, False)
              for name, start, deltas, valid, terminal, error, hold in variants]
    rows = []
    for name, (prefix, suffix), valid, args, hold, recorded in cases:
        row = dict(id=name, expected_acceptance=valid, wire_sha256=hashlib.sha256(prefix + suffix).hexdigest())
        try:
            row.update(passed=True, result=check(binary, prefix, suffix, valid, args, hold, recorded))
        except Exception as error:
            row.update(passed=False, error=type(error).__name__ + ': ' + str(error))
        rows.append(row)
    assert digest(binary) == initial and digest(RECORDED) == RECORDED_SHA
    return dict(scope=__doc__, binary=str(binary), binary_sha256=initial, unchanged=True,
                recorded_response_sha256=RECORDED_SHA, recorded_source='c0e3171fdc22834348f976f81fcaa7372bd13ed4',
                checker_files={str(p.relative_to(ROOT)):digest(p) for p in [Path(__file__), Path(__file__).with_name('accept_ch06_clients.py')]},
                passed=all(row['passed'] for row in rows), checks=rows)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', type=Path)
    parser.add_argument('--receipt', type=Path, required=True)
    args = parser.parse_args()
    result = audit(args.binary.resolve(strict=True))
    args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    raise SystemExit(not result['passed'])
