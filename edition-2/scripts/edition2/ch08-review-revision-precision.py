#!/usr/bin/env python3
"""Reproduce the Ch8 numeric-revision boundary through the actual Chrome UI.

Diagnostic evidence, not a universal acceptance test. The current contract has
no interoperable upper bound; do not silently prescribe one through this probe.
Usage: python3 ch08-review-revision-precision.py GUI_BINARY
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
from accept_ch08 import Backend, Program, defaults

binary = Path(sys.argv[1]).resolve()
# This reproduction intentionally targets the preserved failing executable;
# a repair acceptance check must have its own source identity and expectations.
expected_binary = '6cae6399675914545efc897cef9c1203e8e913e3be269b81d46b89c419d6b617'
assert hashlib.sha256(binary.read_bytes()).hexdigest() == expected_binary, 'frozen diagnostic binary identity differs' 
rows = []
backend = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Backend)
backend.requests = []
thread = threading.Thread(target=backend.serve_forever, daemon=True)
thread.start()
try:
    for revision in (7, 9007199254740992, 9007199254740993):
        with tempfile.TemporaryDirectory(prefix='ch08-revision-') as temporary:
            directory = Path(temporary)
            for domain, value in [('preferences', defaults()), ('policy', {'max_model_requests': 0})]:
                (directory / (domain + '.json')).write_text(json.dumps(dict(version=1, revision=revision, **{domain: value})))
            with contextlib.closing(Program(binary, directory, f'http://127.0.0.1:{backend.server_port}', revision)) as program:
                result = subprocess.run(['node', str(Path(__file__).with_suffix('.cjs')), program.listen()], capture_output=True, text=True, timeout=30)
                assert result.returncode == 0, result.stderr
                rows.append(dict(seed_revision=revision, browser=json.loads(result.stdout), files={domain: json.loads((directory / (domain + '.json')).read_text()) for domain in ('preferences', 'policy')}))
    assert not backend.requests, 'settings controls made an HTTP model request'
    for row in rows:
        messages = row['browser']['outbound']
        commands = [m for m in messages if m['type'] in ('preferences_update', 'policy_update')]
        assert len(commands) == 2
        if row['seed_revision'] == 7:
            assert all(m['base_revision'] == 7 for m in commands)
            assert all(v['revision'] == 8 for v in row['files'].values()), 'safe revision positive failed'
        elif row['seed_revision'] == 9007199254740992:
            assert all(v['revision'] == 9007199254740993 for v in row['files'].values()), 'transition not exercised'
        else:
            assert all(m['base_revision'] == 9007199254740992 for m in commands), 'rounding reproduction changed'
            errors = [m for m in row['browser']['inbound'] if m.get('id') in [c['id'] for c in commands]]
            assert len(errors) == 2 and all(m.get('code') == 'revision_conflict' for m in errors), 'intended conflict absent'
            assert row['files']['preferences']['preferences']['theme'] == 'dark'
            assert row['files']['policy']['policy']['max_model_requests'] == 0
    print(json.dumps(dict(scope='Diagnostic reproduction: actual production Chrome UI and server; safe positive plus unsafe numeric revisions; no model traffic', source_revision='cd9de3e4ec6be4a560144bf46caf9ddd54302dc2', binary=str(binary), binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(), checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__), Path(__file__).with_suffix('.cjs')]}, requests=backend.requests, reproduced=True, rows=rows), indent=2))
finally:
    backend.shutdown()
    backend.server_close()
    thread.join()
