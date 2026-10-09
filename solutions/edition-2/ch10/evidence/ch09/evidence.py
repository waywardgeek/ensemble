"""Identity-first support, adapted from accepted Chapter 8 student machinery."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]
PREFIX = 'solutions/edition-2/main/'
SUPPORT = ['evidence.py', 'terminal-run.py', 'verify-receipts.py', 'browser-live.mjs', 'capture-audio.swift', 'test-evidence.py']
EXECUTABLES = {'cli', 'gui', 'consumer', 'redaction', 'interpreter', 'node', 'recorder', 'chrome', 'capture'}
CAPS = {'N': 16, 'F': 6, 'G': 10, 'P': 4}

def scenario_catalog(binding, scenario):
    result = {}
    for source in binding['catalog']:
        name = Path(source).parent.name
        data = historical(binding['source_revision'], source)
        if scenario == 'G' and name == 'edit':
            data += b'\nLiteral display example: <example> is text, not a page element.\n' + b''.join(f'Review step {i}: inspect the requested scratch file, preserve unrelated content, and report observed results.\n'.encode() for i in range(1,21))
        result[f'catalog/{name}/SKILL.md'] = data
    return result

def scenario_inputs(scenario):
    notes = b'CHAPTER-NINE-NOTE-MARKER\nport=8080\n'
    values = {'notes.txt':notes}
    if scenario == 'P':
        values.update({name+'/notes.txt':notes+name.encode()+b'\n' for name in ('alpha','beta')})
    return values

def digest(data):
    return hashlib.sha256(data).hexdigest()

def file_digest(path):
    with Path(path).open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()

def historical(revision, path):
    assert re.fullmatch('[0-9a-f]{40}', revision), 'invalid immutable revision'
    assert path.startswith(PREFIX) and '..' not in Path(path).parts, 'invalid source path'
    return subprocess.check_output(['git', 'show', revision + ':' + path], cwd=ROOT)

def source_set(revision):
    assert re.fullmatch('[0-9a-f]{40}', revision), 'invalid immutable revision'
    paths = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', revision, '--', PREFIX], cwd=ROOT, text=True).splitlines()
    return sorted(p for p in paths if '/evidence/' not in p and (p.endswith('.go') or Path(p).name in ('go.mod', 'go.sum') or '/gui/web/gui/' in p or '/skills/' in p and p.endswith('/SKILL.md')))

def browser_set(root):
    root = Path(root).resolve()
    assert (root / 'package-lock.json').is_file(), 'browser dependency lock missing'
    return sorted(str(p.resolve()) for p in root.rglob('*') if p.is_file() and p.suffix in ('.js', '.mjs', '.json'))

def preflight(binding, paths=None):
    paths = paths or {}
    revision = binding['source_revision']
    expected = source_set(revision)
    assert expected and binding['sources'], 'empty source set; no evidence modified'
    assert sorted(binding['sources']) == expected, 'incomplete historical source set; no evidence modified'
    for path in expected:
        assert digest(historical(revision, path)) == binding['sources'][path], 'historical source mismatch; no evidence modified'
    assert set(binding['executables']) == EXECUTABLES, 'incomplete executable identities; no evidence modified'
    assert file_digest(sys.executable) == binding['executables']['interpreter']['sha256'], 'interpreter mismatch; no evidence modified'
    for name, identity in binding['executables'].items():
        assert file_digest(paths.get(name, identity['path'])) == identity['sha256'], 'executable mismatch; no evidence modified'
    assert set(binding['support']) == set(SUPPORT), 'incomplete support identities; no evidence modified'
    for name in SUPPORT:
        assert digest(historical(revision, PREFIX + 'evidence/ch09/' + name)) == binding['support'][name], 'historical support mismatch; no evidence modified'
        assert file_digest(HERE / name) == binding['support'][name], 'support identity mismatch; no evidence modified'
    catalog = {p: binding['sources'][p] for p in expected if '/skills/' in p and p.endswith('/SKILL.md')}
    assert catalog and binding['catalog'] == catalog, 'catalog identity mismatch; no evidence modified'
    assert binding['browser_tools'], 'missing browser dependencies; no evidence modified'
    assert sorted(binding['browser_tools']) == browser_set(binding['browser_root']), 'incomplete browser dependency set; no evidence modified'
    for path, sha in binding['browser_tools'].items():
        assert file_digest(path) == sha, 'browser dependency mismatch; no evidence modified'

def launch_identity(binding, executable, scenario):
    return dict(source_revision=binding['source_revision'], executables={n:v['sha256'] for n,v in binding['executables'].items()}, support=binding['support'], catalog=binding['catalog'], browser_tools=binding['browser_tools'], launched_executable=executable, scenario=scenario, http_cap=CAPS[scenario])

def check_launch(binding, launch):
    expected = launch_identity(binding, launch['launched_executable'], launch['scenario'])
    assert launch['launched_executable'] == {'N':'cli', 'F':'cli', 'G':'gui', 'P':'consumer'}[launch['scenario']], 'launch executable role mismatch; no evidence modified'
    for key, value in expected.items():
        assert launch.get(key) == value, 'launch ' + key + ' mismatch; no evidence modified'

def make_binding(revision, paths, browser_root):
    paths = {**paths, 'interpreter': sys.executable, 'recorder':'/usr/bin/script'}
    sources = {p:digest(historical(revision,p)) for p in source_set(revision)}
    binding = dict(source_revision=revision, sources=sources, executables={n:dict(path=str(Path(p).resolve()),sha256=file_digest(p)) for n,p in paths.items()}, support={n:file_digest(HERE/n) for n in SUPPORT}, catalog={p:h for p,h in sources.items() if '/skills/' in p and p.endswith('/SKILL.md')}, browser_root=str(Path(browser_root).resolve()), browser_tools={p:file_digest(p) for p in browser_set(browser_root)})
    preflight(binding, paths)
    return binding
