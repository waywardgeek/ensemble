#!/usr/bin/env python3
"""Exact settings revisions through actual Chrome UI and the public Connector.

Usage: python3 ch08-review-revisions.py GUI_BINARY
No paid model traffic. This supplements, not replaces, the complete Ch8 gate.
"""
import contextlib
import hashlib
import http.server
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import threading
import urllib.request
from accept_ch08 import Backend, Program, defaults

MAX = (1 << 64)-1
binary = Path(sys.argv[1]).resolve()
backend = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Backend)
backend.requests = []
thread = threading.Thread(target=backend.serve_forever, daemon=True)
thread.start()
checks = []

def evaluate(revision, mutation=False):
    with tempfile.TemporaryDirectory(prefix='ch08-exact-revisions-') as temporary:
        directory = Path(temporary)
        for domain, value in [('preferences', defaults()), ('policy', {'max_model_requests': 0})]:
            (directory/(domain+'.json')).write_text(json.dumps(dict(version=1, revision=revision, **{domain: value})))
        with contextlib.closing(Program(binary, directory, f'http://127.0.0.1:{backend.server_port}', revision)) as program:
            url = program.listen()
            command = ['node', str(Path(__file__).with_suffix('.cjs')), url]
            source = urllib.request.urlopen(url+'/connector.js').read()
            if mutation:
                anchor = b'this.receive(this.decode(event.data))'
                assert source.count(anchor) == 1, 'lossless decoder deletion anchor changed'
                replacement = directory/'connector.js'
                replacement.write_bytes(source.replace(anchor, b'this.receive(JSON.parse(event.data))'))
                command.append(str(replacement))
            process = subprocess.run(command, capture_output=True, text=True, timeout=40)
            assert process.returncode == 0, 'browser execution failed: '+process.stderr
            result = json.loads(process.stdout)
            if mutation:
                assert 'base revision must be exact' in result.get('mutation_refusal', '').lower(), 'decoder deletion refusal differs'
                raise AssertionError('preferences-no-change: exact base unavailable after decoder deletion')
            expected = revision
            states = {d: revision for d in ('preferences', 'policy')}
            for step in result['steps']:
                command, reply = json.loads(step['command']), json.loads(step['reply'])
                domain = command['type'].removesuffix('_update')
                assert command['base_revision'] == states[domain], step['label']+': exact numeric base lost'
                assert type(command['base_revision']) is int, step['label']+': base is not numeric JSON'
                changing = not step['label'].endswith('no-change')
                if changing and states[domain] == MAX:
                    assert reply.get('code') == 'settings_persist_failed', step['label']+': exhaustion not refused'
                else:
                    if changing:
                        states[domain] += 1
                    assert reply['type'] == domain+'_ack' and reply['revision'] == states[domain], step['label']+': exact acknowledgement absent'
                assert str(states[domain]) in step['applied'], step['label']+': live applied revision rounded'
            for retry in result['retries']:
                expected = str(states[retry['domain']])
                assert retry['before'] == retry['conflict'] == retry['ack'] == expected, 'conflict snapshot or deliberate retry lost precision'
            for domain, expected in states.items():
                assert json.loads((directory/(domain+'.json')).read_text())['revision'] == expected, 'persisted exact revision differs'
                assert str(expected) in result['reloaded'][domain if domain == 'preferences' else 'policy'], 'reconnect displays rounded revision'
            return dict(revision=str(revision), connector_sha256=hashlib.sha256(source).hexdigest(), result=result)
try:
    for revision in (7, (1 << 53)-1, (1 << 53)+1, MAX-1, MAX):
        try:
            result = evaluate(revision)
            checks.append(dict(id='revision-'+str(revision), passed=True, result=result))
        except Exception as error:
            checks.append(dict(id='revision-'+str(revision), passed=False, error=str(error)))
    if all(row['passed'] for row in checks):
        try:
            evaluate((1 << 53)+1, mutation=True)
            checks.append(dict(id='delete-lossless-decode', passed=False, error='mutation accepted'))
        except Exception as error:
            message = str(error)
            checks.append(dict(id='delete-lossless-decode', passed=message == 'preferences-no-change: exact base unavailable after decoder deletion', intended_refusal='preferences-no-change: exact base unavailable after decoder deletion', actual_refusal=message))
    assert not backend.requests, 'settings made a model request'
    passed = len(checks) == 6 and all(row['passed'] for row in checks)
    print(json.dumps(dict(scope='Actual Chrome production UI and public Connector, exact persisted revisions, reconnect, conflicts, exhaustion, targeted decoder deletion; no model calls', binary=str(binary), binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(), checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__),Path(__file__).with_suffix('.cjs')]}, passed=passed, checks=checks, requests=backend.requests),indent=2))
    sys.exit(0 if passed else 1)
finally:
    backend.shutdown()
    backend.server_close()
    thread.join()
