#!/usr/bin/env python3
"""Affected maintenance checks on a source-bound sparse worktree; no paid calls."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('source', type=Path)
p.add_argument('revision')
p.add_argument('checkers', type=Path)
p.add_argument('binary', type=Path)
p.add_argument('receipt', type=Path)
p.add_argument('--stage', type=int, choices=(7, 8), required=True)
a = p.parse_args()
source, checkers = a.source.resolve(strict=True), a.checkers.resolve(strict=True)
repo = subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], cwd=source, text=True).strip()
revision = subprocess.check_output(['git', 'rev-parse', a.revision+'^{commit}'], cwd=source, text=True).strip()
prefix = 'solutions/edition-2/main/'

def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def bind():
    hashes = {}
    names = subprocess.check_output(['git', '-C', repo, 'ls-tree', '-r', '--name-only', revision, prefix], text=True).splitlines()
    for name in names:
        relative = Path(name).relative_to(prefix)
        if 'evidence' in relative.parts:
            continue
        blob = subprocess.check_output(['git', '-C', repo, 'show', revision+':'+name])
        expected = hashlib.sha256(blob).hexdigest()
        assert digest(source/relative) == expected, 'source mismatch: '+str(relative)
        hashes[str(relative)] = expected
    assert hashes
    return hashes

hashes = bind()
checker_hashes = {str(path.relative_to(checkers)): digest(path) for path in checkers.rglob('*')
                  if path.is_file() and '__pycache__' not in path.parts}
rows = []
with tempfile.TemporaryDirectory(prefix='maintenance-retained-adapters-') as temporary:
    bundle = checkers
    adapters = None
    if a.stage == 8:
        sys.path.insert(0, str(checkers))
        from ch08_retained_adapters import prepare
        bundle = Path(temporary)/'checkers'
        adapters = prepare(bundle, source)

    def run(name, command, cwd=source, timeout=600):
        result = subprocess.run([str(x) for x in command], cwd=cwd, text=True, capture_output=True, timeout=timeout)
        row = dict(id=name, command=[str(x) for x in command], cwd=str(cwd), exit=result.returncode,
                   stdout=result.stdout, stderr=result.stderr, passed=result.returncode == 0)
        rows.append(row)
        print(name, 'PASS' if row['passed'] else 'FAIL', flush=True)
        a.receipt.write_text(json.dumps(dict(source_revision=revision, source_files=hashes, complete=False, checks=rows), indent=2)+'\n')
        return row['passed']

    built = run('cli-build', ['go', 'build', '-o', a.binary, './cmd'])
    binary_hash = digest(a.binary) if built else None
    if built:
        run('empty-arguments-cli', ['python3', checkers/'ch09-review-empty-stream.py', a.binary,
                                  '--receipt', a.receipt.with_name('cli-boundaries.json')])
        run('early-public-and-pty-delivery', ['python3', bundle/'accept_ch06_clients.py', a.binary])
        run('retained-wire-61', ['python3', bundle/'accept_ch06_wire.py', a.binary])
    run('retained-stream-contract', ['python3', bundle/'accept_ch06.py', source])
    run('retained-watch-pause', ['python3', bundle/'accept_ch07_public.py', source])
    run('retained-prior-assertions', ['python3', bundle/'accept_ch07_prior.py', source])
    run('retained-cli-overflow-deletions', ['python3', bundle/'ch07-review-prior-cli.py', revision])
    run('complete-delivered-package-discovery', ['python3', checkers/'accept_delivered_tree.py', revision])
    run('core-race', ['go', 'test', '-race', './...', '-count=1'], timeout=180)
    for module in sorted(source.rglob('go.mod')):
        if 'evidence' in module.relative_to(source).parts or module.parent == source:
            continue
        label = str(module.parent.relative_to(source))
        run(label+'/vet', ['go', 'vet', './...'], module.parent)
        run(label+'/tests', ['go', 'test', './...', '-count=1'], module.parent)
    assert bind() == hashes, 'source changed during checks'
    assert all(digest(checkers/path) == value for path, value in checker_hashes.items()), 'checker changed during checks'
    if built:
        assert digest(a.binary) == binary_hash, 'executable changed during checks'
    result = dict(scope=__doc__, stage=a.stage, source_revision=revision, source_files=hashes,
                  checker_files=checker_hashes, retained_adapters=adapters, binary=str(a.binary),
                  binary_sha256=binary_hash, complete=True, passed=all(row['passed'] for row in rows),
                  checks=rows, runner_sha256=digest(Path(__file__)))
    a.receipt.write_text(json.dumps(result, indent=2)+'\n')
    raise SystemExit(not result['passed'])
