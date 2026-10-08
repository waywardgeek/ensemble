"""Negative identity controls using real receipts without mutating their files."""
from copy import deepcopy
import importlib.util
import json
from pathlib import Path
import tempfile

from evidence import HERE, digest

spec = importlib.util.spec_from_file_location('addendum', HERE / 'verify-addendum.py')
adapter = importlib.util.module_from_spec(spec)
spec.loader.exec_module(adapter)
plan = json.loads((HERE / 'gemini38-addendum-plan.json').read_text())


def raw_hashes():
    return {str(p.relative_to(HERE)): digest(p.read_bytes())
            for run in plan['runs'].values()
            for p in sorted((HERE / run['directory']).rglob('*')) if p.is_file()}


before = raw_hashes()
adapter.validate_plan(plan, {})
rows = [{'control': 'all actual bindings and launches', 'result': 'passed'}]
with tempfile.TemporaryDirectory() as temporary:
    root = Path(temporary).resolve()
    adapter.HERE = root
    for path in HERE.iterdir():
        (root / path.name).symlink_to(path, target_is_directory=path.is_dir())
    original = json.loads((HERE / 'revised-binding.json').read_text())
    binding_file = root / 'revised-binding.json'
    binding_file.unlink()

    def install(binding):
        data = json.dumps(binding).encode()
        binding_file.write_bytes(data)
        plan['bindings']['revised']['sha256'] = digest(data)

    def refused(name):
        try:
            adapter.verify(plan, {}, root / 'derived')
        except AssertionError as error:
            assert not (root / 'derived').exists()
            assert raw_hashes() == before
            rows.append({'control': name, 'refusal': str(error), 'raw_unchanged': True,
                         'no_derived_write': True})
        else:
            raise AssertionError('invalid identity passed: ' + name)

    mutations = [('empty source set', lambda b: b.update(sources={})),
                 ('missing source', lambda b: b['sources'].pop(next(iter(b['sources'])))),
                 ('extra source', lambda b: b['sources'].update({'fake.go': '0' * 64})),
                 ('bad source hash', lambda b: b['sources'].update({next(iter(b['sources'])): '0' * 64}))]
    for name in ('cli', 'workflow', 'interpreter', 'recorder'):
        mutations.append(('wrong ' + name, lambda b, name=name: b['executables'][name].update(sha256='0' * 64)))
    for name, mutate in mutations:
        bad = deepcopy(original)
        mutate(bad)
        install(bad)
        refused(name)
    install(original)
    last = root / plan['runs']['input-openai']['directory']
    last.unlink()
    last.mkdir()
    actual = HERE / last.name
    for path in actual.iterdir():
        if path.name != 'launch.json':
            (last / path.name).symlink_to(path, target_is_directory=path.is_dir())
    launch = json.loads((actual / 'launch.json').read_text())
    for name in ('cli', 'workflow', 'interpreter', 'recorder'):
        bad = deepcopy(launch)
        bad['executables'][name] = '0' * 64
        (last / 'launch.json').write_text(json.dumps(bad))
        refused('last launch ' + name)
    bad = deepcopy(launch)
    bad['command'] = ['/wrong', 'chat']
    (last / 'launch.json').write_text(json.dumps(bad))
    refused('last launch command')
assert raw_hashes() == before
result = {'controls': rows, 'raw_receipts_before_and_after': before}
(HERE / 'local-controls/live-addendum-identity-results.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps({'controls': len(rows), 'raw_files_preserved': len(before)}, indent=2))
