#!/usr/bin/env python3
"""Partial Chapter 8 strict file/patch validation from the published contract.

No Chapter 8 student/reference implementation was used to derive these fixtures.
This is not the public policy-effect, disk-fault, concurrency or browser gate.
"""
import argparse
import contextlib
import copy
import hashlib
import http.server
import json
from pathlib import Path
import queue
import subprocess
import tempfile
import threading
import time
from accept_ch07 import Socket
from accept_ch08 import Backend, Program, defaults, subscribe, update

SENTINEL = 'UNTRUSTED-CH08-CONTENT-MUST-NOT-BE-ECHOED'


def seed(domain):
    value = defaults() | dict(theme='light', autoplay=True, font_size=22, speech_rate=1.5) if domain == 'preferences' else dict(max_model_requests=17)
    return dict(version=1, revision=7 if domain == 'preferences' else 9, **{domain: value})


def encoded(value):
    return json.dumps(value, separators=(',', ':'), ensure_ascii=False).encode()


def files(directory):
    return {d: directory/(d+'.json') for d in ('preferences', 'policy')}


def initialize(directory):
    for domain, path in files(directory).items():
        path.write_bytes(encoded(seed(domain)))


def file_cases(domain):
    valid = seed(domain)
    raw = encoded(valid)
    yield 'exact-65536', raw + b' '*(65536-len(raw)), True
    yield 'oversize-65537', raw + b' '*(65537-len(raw)), False
    for name, body in [('empty', b''), ('null-root', b'null'), ('array-root', b'[]'), ('invalid-utf8', raw+b'\xff'), ('trailing-object', raw+b'{}'), ('trailing-text', raw+b' '+SENTINEL.encode()), ('duplicate-version', raw[:-1]+b',"version":1}'), ('duplicate-domain', raw[:-1]+b','+encoded(domain)+b':'+encoded(valid[domain])+b'}')]:
        yield name, body, False
    for key, value in [('version', 2), ('version', True), ('version', None), ('revision', -1), ('revision', True), ('revision', '7'), ('revision', 0.5), ('revision', None), (domain, None), (domain, [])]:
        broken = copy.deepcopy(valid)
        broken[key] = value
        yield 'invalid-'+key+'-'+repr(value), encoded(broken), False
    for key in valid:
        broken = copy.deepcopy(valid)
        del broken[key]
        yield 'missing-'+key, encoded(broken), False
    broken = copy.deepcopy(valid)
    broken['unrecognized'] = SENTINEL
    yield 'unknown-envelope-field', encoded(broken), False
    for key in valid[domain]:
        broken = copy.deepcopy(valid)
        del broken[domain][key]
        yield 'missing-value-'+key, encoded(broken), False
    for name, patch in patch_cases(domain):
        # Complete-file validation gets the same invalid member values while
        # retaining every unrelated required member. Duplicate members are
        # appended as raw JSON, never normalized by a dictionary decoder.
        if name == 'duplicate-field':
            pairs = list(valid[domain].items())
            key = 'autoplay' if domain == 'preferences' else 'max_model_requests'
            pairs = [(k, v) for k, v in pairs if k != key]
            prefix = b','.join(encoded(k)+b':'+encoded(v) for k, v in pairs)
            merged = b'{' + prefix + (b',' if prefix else b'') + patch[1:]
        else:
            merged = encoded(valid[domain] | json.loads(patch))
        body = b'{"version":1,"revision":7,'+encoded(domain)+b':'+merged+b'}'
        yield 'invalid-members-'+name, body, False


def patch_cases(domain):
    if domain == 'preferences':
        values = [('theme', 'sepia'), ('font_size', 11), ('font_size', 29), ('font_size', 12.5), ('font_size', '16'), ('font_size', True), ('sidebar_width', 179), ('sidebar_width', 481), ('actions_width', 199), ('actions_width', 641), ('autoplay', 0), ('autoplay', 'false'), ('speech_rate', 0.49), ('speech_rate', 2.01), ('speech_rate', '1'), ('speech_rate', True)]
    else:
        values = [('max_model_requests', -1), ('max_model_requests', 257), ('max_model_requests', 1.5), ('max_model_requests', '0'), ('max_model_requests', True)]
    values += [(key, None) for key in seed(domain)[domain]]
    for index, (key, value) in enumerate(values):
        yield f'{index}-{key}', encoded({key: value})
    yield 'unknown-field', encoded({'unrecognized': SENTINEL})
    key = 'autoplay' if domain == 'preferences' else 'max_model_requests'
    yield 'duplicate-field', b'{'+encoded(key)+b':0,'+encoded(key)+b':0}' if domain == 'policy' else b'{"autoplay":true,"autoplay":false}'


def check_refusal(reply, command_id, domain):
    assert reply.get('type') == 'error' and reply.get('id') == command_id, 'missing correlated settings refusal'
    assert reply.get('code') == 'invalid_'+domain, 'wrong intended settings refusal'
    assert isinstance(reply.get('message'), str) and reply['message'].strip(), 'missing safe reason'
    assert SENTINEL not in json.dumps(reply), 'untrusted patch echoed'


def check_startup_refusal(exit_code, lines, original, after):
    assert exit_code is not None and exit_code != 0, 'invalid file did not fail startup'
    assert not any('http://127.0.0.1:' in line for line in lines), 'invalid settings file reached listening state'
    assert ''.join(lines).strip(), 'startup failure lacks reason'
    assert SENTINEL not in ''.join(lines), 'untrusted file content echoed'
    assert after == original, 'invalid file overwritten'


def rejected_startup(program, path, original):
    lines = []
    deadline = time.monotonic()+8
    while time.monotonic() < deadline:
        try:
            line = program.lines.get(timeout=max(0.01, deadline-time.monotonic()))
        except queue.Empty:
            break
        if line is None:
            break
        lines.append(line)
        if 'http://127.0.0.1:' in line:
            break
    if program.process.poll() is None:
        try:
            program.process.wait(timeout=0.2)
        except subprocess.TimeoutExpired:
            pass
    check_startup_refusal(program.process.poll(), lines, original, path.read_bytes())


def evaluate(binary):
    rows = []
    backend = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Backend)
    backend.requests = []
    thread = threading.Thread(target=backend.serve_forever, daemon=True)
    thread.start()
    endpoint = f'http://127.0.0.1:{backend.server_port}'
    number = 0
    def run(name, action):
        try:
            action()
            rows.append(dict(id=name, passed=True))
            return True
        except Exception as error:
            rows.append(dict(id=name, passed=False, details=type(error).__name__+': '+str(error)))
            return False
    try:
        with tempfile.TemporaryDirectory(prefix='ch08-validation-') as temporary:
            directory = Path(temporary)
            initialize(directory)
            def positive():
                with contextlib.closing(Program(binary, directory, endpoint, 0)) as program:
                    with contextlib.closing(Socket(program.listen())) as client:
                        subscribe(client, seed('preferences')['preferences'], 7, 17, 9)
            baseline = run('nondefault-complete-files-positive', positive)
            if baseline:
                for domain in ('preferences', 'policy'):
                    for name, raw, valid in file_cases(domain):
                        number += 1
                        initialize(directory)
                        path = files(directory)[domain]
                        path.write_bytes(raw)
                        def test(raw=raw, valid=valid, path=path):
                            with contextlib.closing(Program(binary, directory, endpoint, number)) as program:
                                if valid:
                                    with contextlib.closing(Socket(program.listen())) as client:
                                        subscribe(client, seed('preferences')['preferences'], 7, 17, 9)
                                    assert path.read_bytes() == raw, 'startup rewrote valid file'
                                else:
                                    rejected_startup(program, path, raw)
                        run(domain+'/'+name, test)
                initialize(directory)
                number += 1
                def patches():
                    with contextlib.closing(Program(binary, directory, endpoint, number)) as program:
                        url = program.listen()
                        with contextlib.closing(Socket(url)) as client:
                            subscribe(client, seed('preferences')['preferences'], 7, 17, 9)
                            saved = {domain: path.read_bytes() for domain, path in files(directory).items()}
                            for domain in ('preferences', 'policy'):
                                for index, (name, patch) in enumerate(patch_cases(domain)):
                                    command_id = domain+'-'+str(index)
                                    body = b'{"type":'+encoded(domain+'_update')+b',"id":'+encoded(command_id)+b',"base_revision":'+str(seed(domain)['revision']).encode()+b',"patch":'+patch+b'}'
                                    client.ws.send(body.decode())
                                    reply = client.until(lambda x: x.get('id') == command_id)
                                    check_refusal(reply, command_id, domain)
                                    assert {d: p.read_bytes() for d, p in files(directory).items()} == saved, 'invalid patch changed persisted bytes'
                            with contextlib.closing(Socket(url)) as fresh:
                                subscribe(fresh, seed('preferences')['preferences'], 7, 17, 9)
                            update(client, 'preferences', 'valid-after-refusals', 7, dict(autoplay=False), 8, seed('preferences')['preferences'] | dict(autoplay=False))
                            update(client, 'policy', 'zero-after-refusals', 9, dict(max_model_requests=0), 10, 0)
                run('strict-patches-no-mutation-and-recovery', patches)
    finally:
        backend.shutdown()
        backend.server_close()
        thread.join(timeout=2)
    rows.append(dict(id='no-model-http', passed=not backend.requests, details=backend.requests))
    return dict(scope=__doc__, binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                checker_files={name: hashlib.sha256(Path(__file__).with_name(name).read_bytes()).hexdigest() for name in ['accept_ch08_validation.py', 'accept_ch08.py', 'accept_ch07.py']},
                passed=baseline and all(r['passed'] for r in rows), checks=rows)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', type=Path)
    args = parser.parse_args()
    result = evaluate(args.binary.resolve(strict=True))
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
