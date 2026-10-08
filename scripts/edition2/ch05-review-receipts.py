#!/usr/bin/env python3
"""Independent frozen-receipt audit and unmasked identity controls; no model calls."""
from copy import deepcopy
import argparse
import hashlib
import importlib.util
import json
import pathlib
import shutil
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[2]
HERE = ROOT / 'solutions/edition-2/main/evidence/ch05'
FREEZE = '469730f7217e09620d471f8a71825a92d760113f'

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--executable', action='append', default=[], metavar='BINDING.NAME=PATH',
                    help='explicit archived executable path; its bound hash must still match')
args = parser.parse_args()
overrides = {}
for value in args.executable:
    identity, path = value.split('=', 1)
    binding, name = identity.split('.', 1)
    overrides.setdefault(binding, {})[name] = str(pathlib.Path(path).resolve())


def digest(data):
    return hashlib.sha256(data).hexdigest()


def historical(path, revision=FREEZE):
    return subprocess.check_output(['git', 'show', revision + ':' + str(path.relative_to(ROOT))], cwd=ROOT)


def read(name):
    return json.loads((HERE / name).read_text())


plan = read('gemini38-addendum-plan.json')
manifest = read('revised-receipt-manifest.json')
masked = read('local-controls/live-addendum-identity-results.json')
raw_hashes = masked['raw_receipts_before_and_after']
assert len(raw_hashes) == 247
for name, identity in {**manifest['files'], **raw_hashes}.items():
    path = HERE / name
    assert digest(path.read_bytes()) == identity, name
    assert digest(historical(path)) == identity, 'frozen receipt differs: ' + name

bindings = {}
source_counts = {}
for name, entry in plan['bindings'].items():
    path = HERE / entry['path']
    assert digest(path.read_bytes()) == entry['sha256']
    assert historical(path) == path.read_bytes()
    binding = json.loads(path.read_text())
    revision = binding['source_revision']
    names = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', revision, '--', 'solutions/edition-2/main'], cwd=ROOT, text=True).splitlines()
    expected = {p for p in names if p.endswith('.go') or pathlib.Path(p).name in ('go.mod', 'go.sum')}
    assert expected and expected == set(binding['sources'])
    for source, identity in binding['sources'].items():
        assert digest(historical(ROOT / source, revision)) == identity
    for executable, identity in binding['executables'].items():
        path = overrides.get(name, {}).get(executable, identity['path'])
        assert digest(pathlib.Path(path).read_bytes()) == identity['sha256']
    for support, identity in {**binding['support'], **binding.get('addendum_support', {})}.items():
        path = HERE / support
        assert digest(path.read_bytes()) == identity == digest(historical(path, revision))
    source_counts[name] = len(expected)
    bindings[name] = binding

sys.path.insert(0, str(HERE))
spec = importlib.util.spec_from_file_location('reviewed_addendum', HERE / 'verify-addendum.py')
adapter = importlib.util.module_from_spec(spec)
spec.loader.exec_module(adapter)
rows = []
with tempfile.TemporaryDirectory(prefix='ensemble-ch05-receipt-review-') as directory:
    copy = pathlib.Path(directory).resolve()
    # The adapter's containment root and every fixture path agree. Original
    # evidence stays read-only; a mutation changes one contained copy only.
    adapter.HERE = copy
    for name in set(plan['runs'][r]['directory'] for r in plan['runs']):
        shutil.copytree(HERE / name, copy / name)
    for entry in plan['bindings'].values():
        shutil.copy2(HERE / entry['path'], copy / entry['path'])
    for binding in bindings.values():
        for name in {**binding['support'], **binding.get('addendum_support', {})}:
            shutil.copy2(HERE / name, copy / name)
    adapter.validate_plan(plan, overrides)
    actual = adapter.verify(plan, overrides, copy / 'positive-derived')
    recorded = read('verified-gemini38/receipts.json')['runs']
    assert actual == recorded, 'independent replay disagrees with recorded results'
    assert sum(len(row['requests']) for row in actual.values()) == 78
    rows.append({'control': 'contained copies of all actual bindings and runs',
                 'result': 'passed', 'runs': len(actual), 'exact_requests': 78})

    original_check_output = subprocess.check_output
    replay_calls = []

    def guard(command, *args, **kwargs):
        if str(command[0]) != 'git':
            replay_calls.append([str(s) for s in command])
            raise AssertionError('identity failure reached replay')
        return original_check_output(command, *args, **kwargs)

    def hashes():
        return {str(p.relative_to(copy)): digest(p.read_bytes()) for p in copy.rglob('*') if p.is_file()}

    def refuse(label, changed_plan, expected):
        before = hashes()
        replay_calls.clear()
        subprocess.check_output = guard
        try:
            try:
                adapter.verify(changed_plan, overrides, copy / 'negative-derived')
            except AssertionError as error:
                reason = str(error)
                assert reason.startswith(expected), (label, reason, expected)
            else:
                raise AssertionError('invalid identity passed: ' + label)
        finally:
            subprocess.check_output = original_check_output
        assert not replay_calls and hashes() == before and not (copy / 'negative-derived').exists()
        rows.append({'control': label, 'refusal': reason, 'replay_calls': 0,
                     'copied_files_unchanged': True, 'no_derived_write': True})

    mutations = [
        ('empty source set', lambda b: b.update(sources={}), 'empty source set'),
        ('missing source', lambda b: b['sources'].pop(next(iter(b['sources']))), 'incomplete historical source set'),
        ('extra source', lambda b: b['sources'].update({'fake.go': '0' * 64}), 'incomplete historical source set'),
        ('bad source hash', lambda b: b['sources'].update({next(iter(b['sources'])): '0' * 64}), 'historical source mismatch'),
    ]
    for name in ('cli', 'workflow', 'interpreter', 'recorder'):
        expected = 'interpreter mismatch' if name == 'interpreter' else 'executable mismatch'
        mutations.append(('wrong ' + name, lambda b, n=name: b['executables'][n].update(sha256='0' * 64), expected))
    binding_path = copy / plan['bindings']['revised']['path']
    original_binding = binding_path.read_bytes()
    for label, mutate, expected in mutations:
        value = deepcopy(bindings['revised'])
        mutate(value)
        binding_path.write_text(json.dumps(value))
        changed = deepcopy(plan)
        changed['bindings']['revised']['sha256'] = digest(binding_path.read_bytes())
        refuse(label, changed, expected)
        binding_path.write_bytes(original_binding)

    last = copy / plan['runs'][list(plan['runs'])[-1]]['directory'] / 'launch.json'
    original_launch = last.read_bytes()
    for name in ('cli', 'workflow', 'interpreter', 'recorder', 'command'):
        value = json.loads(original_launch)
        if name == 'command':
            value['command'] = ['/wrong', 'chat']
        else:
            value['executables'][name] = '0' * 64
        last.write_text(json.dumps(value))
        refuse('last launch ' + name, plan, 'launch command mismatch' if name == 'command' else 'launch executable mismatch')
        last.write_bytes(original_launch)
    adapter.validate_plan(plan, overrides)

assert all(digest((HERE / name).read_bytes()) == identity for name, identity in raw_hashes.items())
print(json.dumps({'freeze': FREEZE, 'script_sha256': digest(pathlib.Path(__file__).read_bytes()),
                  'executable_overrides': overrides,
                  'manifest_files_verified': len(manifest['files']), 'raw_files_preserved': len(raw_hashes),
                  'source_counts': source_counts, 'original_controls_masked_by_path_check': True,
                  'controls': rows, 'runs': {n: {'requests': len(r['requests']), 'source_revision': r['source_revision'],
                    'model': r['requested_model'], 'calls': r['calls'], 'results': r['results'], 'usage': r['usage']} for n, r in actual.items()}}, indent=2))
