"""Historical source and executable preflight shared by launch and replay tools."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]
PREFIX = 'solutions/edition-2/main/'
SUPPORT = ['evidence.py', 'terminal-run.py', 'verify-receipts.py']

def digest(data):
    return hashlib.sha256(data).hexdigest()

def source_set(revision):
    paths = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', revision, '--', PREFIX], cwd=ROOT, text=True).splitlines()
    return sorted(p for p in paths if p.endswith('.go') or Path(p).name in ('go.mod', 'go.sum'))

def preflight(binding, paths):
    expected = source_set(binding['source_revision'])
    assert expected and binding['sources'], 'empty source set; no evidence modified'
    assert sorted(binding['sources']) == expected, 'incomplete historical source set; no evidence modified'
    for path in expected:
        data = subprocess.check_output(['git', 'show', binding['source_revision'] + ':' + path], cwd=ROOT)
        assert digest(data) == binding['sources'][path], 'historical source mismatch; no evidence modified'
    assert set(binding['executables']) == {'cli', 'consumer', 'interpreter', 'recorder'}, 'incomplete executable identities; no evidence modified'
    assert digest(Path(sys.executable).read_bytes()) == binding['executables']['interpreter']['sha256'], 'interpreter mismatch; no evidence modified'
    for name, identity in binding['executables'].items():
        assert digest(Path(paths.get(name, identity['path'])).read_bytes()) == identity['sha256'], 'executable mismatch; no evidence modified'
    assert set(binding['support']) == set(SUPPORT), 'incomplete support identities; no evidence modified'
    for name in SUPPORT:
        assert digest((HERE / name).read_bytes()) == binding['support'][name], 'support identity mismatch; no evidence modified'

def make_binding(revision, paths):
    paths = {**paths, 'interpreter': sys.executable, 'recorder': '/usr/bin/script'}
    files = source_set(revision)
    assert files
    binding = {'source_revision': revision, 'sources': {p: digest(subprocess.check_output(['git', 'show', revision + ':' + p], cwd=ROOT)) for p in files},
               'executables': {name: {'path': str(Path(path).resolve()), 'sha256': digest(Path(path).read_bytes())} for name, path in paths.items()},
               'support': {name: digest((HERE / name).read_bytes()) for name in SUPPORT}}
    preflight(binding, paths)
    return binding
