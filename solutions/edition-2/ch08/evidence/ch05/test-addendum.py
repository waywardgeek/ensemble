"""Credential-free preflight controls for the mixed-revision receipt adapter."""
from copy import deepcopy
import importlib.util
import json
from pathlib import Path
import tempfile

from evidence import HERE, digest

spec = importlib.util.spec_from_file_location('addendum', HERE / 'verify-addendum.py')
adapter = importlib.util.module_from_spec(spec)
spec.loader.exec_module(adapter)
binding = json.loads((HERE / 'local-controls/positive-binding-fixture.json').read_text())
rows = []
with tempfile.TemporaryDirectory() as temporary:
    root = Path(temporary).resolve()
    adapter.HERE = root
    plan = {'verifier_sha256': digest((HERE / 'verify-addendum.py').read_bytes()),
            'bindings': {'fixture': {'path': 'binding.json', 'sha256': ''}}, 'runs': {}}
    for logical in sorted(adapter.EXPECTED):
        mode, vendor = logical.split('-')
        mode = 'chat' if mode in ('controls', 'eof', 'input') else mode
        exe = 'cli' if mode == 'chat' else 'workflow'
        run = root / logical
        run.mkdir()
        paths = {n: d['path'] for n, d in binding['executables'].items()}
        launch = {'source_revision': binding['source_revision'], 'executables': {n: d['sha256'] for n, d in binding['executables'].items()},
                  'support': binding['support'], 'executable_paths': paths, 'launched_executable': exe,
                  'command': [paths[exe], mode], 'mode': mode, 'vendor': vendor, 'requested_model': 'models/gemini-3.8-flash', 'exit_code': 0}
        (run / 'launch.json').write_text(json.dumps(launch))
        (run / 'raw-fixture.txt').write_text('Fake adapter control; no model or application executed.\n')
        plan['runs'][logical] = {'directory': logical, 'binding': 'fixture'}
    def install(value):
        data = json.dumps(value).encode()
        (root / 'binding.json').write_bytes(data)
        plan['bindings']['fixture']['sha256'] = digest(data)
    def raw_hashes():
        return {str(p.relative_to(root)): digest(p.read_bytes()) for p in root.rglob('*') if p.is_file()}
    install(binding)
    adapter.validate_plan(plan, {})
    rows.append({'control': 'complete fake launch map', 'result': 'preflight passed', 'fake_only': True})
    mutations = [('empty source set', lambda b: b.update(sources={})),
                 ('missing source', lambda b: b['sources'].pop(next(iter(b['sources'])))),
                 ('extra source', lambda b: b['sources'].update({'fake.go': '0' * 64})),
                 ('bad source hash', lambda b: b['sources'].update({next(iter(b['sources'])): '0' * 64}))]
    for name in ('cli', 'workflow', 'interpreter', 'recorder'):
        mutations.append(('wrong ' + name, lambda b, name=name: b['executables'][name].update(sha256='0' * 64)))
    for name, mutation in mutations:
        bad = deepcopy(binding)
        mutation(bad)
        install(bad)
        before = raw_hashes()
        try:
            adapter.verify(plan, {}, root / 'derived')
        except AssertionError as error:
            assert raw_hashes() == before and not (root / 'derived').exists()
            rows.append({'control': name, 'refusal': str(error), 'raw_unchanged': True})
        else:
            raise AssertionError('invalid identity passed: ' + name)
    install(binding)
    last = root / sorted(adapter.EXPECTED)[-1] / 'launch.json'
    original = last.read_bytes()
    for name, mutation in [('last launch executable', lambda l: l['executables'].update(cli='0' * 64)),
                           ('last launch command', lambda l: l.update(command=['/wrong', 'workflow']))]:
        value = json.loads(original)
        mutation(value)
        last.write_text(json.dumps(value))
        before = raw_hashes()
        try:
            adapter.verify(plan, {}, root / 'derived')
        except AssertionError as error:
            assert raw_hashes() == before and not (root / 'derived').exists()
            rows.append({'control': name, 'refusal': str(error), 'raw_unchanged': True})
        else:
            raise AssertionError('invalid final launch passed')
    last.write_bytes(original)
    before = raw_hashes()
    try:
        adapter.verify(plan, {}, root / 'derived')
    except AssertionError as error:
        assert 'raw request' in str(error) and raw_hashes() == before
        rows.append({'control': 'preflight alone is not live evidence', 'refusal': str(error), 'raw_unchanged': True})
    else:
        raise AssertionError('fake metadata accepted as live evidence')
(HERE / 'local-controls/addendum-results.json').write_text(json.dumps(rows, indent=2) + '\n')
print(json.dumps(rows, indent=2))
