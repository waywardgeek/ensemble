#!/usr/bin/env python3
"""Discover every Go package in the complete immutable delivered source tree.

Includes evidence/support sources and nested modules. This supplements runtime
test copies; it does not claim vet, tests, explicit helper use or live behavior.
"""
import argparse
import hashlib
import io
import json
from pathlib import Path, PurePosixPath
import subprocess
import tarfile
import tempfile


def discover(directory):
    modules = sorted(directory.rglob('go.mod'))
    rows = []
    for module in modules:
        result = subprocess.run(['go', 'list', './...'], cwd=module.parent,
                                text=True, capture_output=True, timeout=120)
        rows.append(dict(module=str(module.parent.relative_to(directory)),
                         command=['go', 'list', './...'], exit=result.returncode,
                         stdout=result.stdout, stderr=result.stderr))
    return dict(passed=bool(modules) and all(r['exit'] == 0 for r in rows), checks=rows)


def evaluate(commit, prefix='solutions/edition-2/main'):
    repository = Path(__file__).resolve().parents[2]
    source = subprocess.check_output(['git', 'rev-parse', '--verify', commit+'^{commit}'], cwd=repository, text=True).strip()
    tree = subprocess.check_output(['git', 'rev-parse', source+':'+prefix], cwd=repository, text=True).strip()
    archive = subprocess.check_output(['git', 'archive', source, prefix], cwd=repository)
    with tempfile.TemporaryDirectory(prefix='ensemble-complete-delivery-') as temporary:
        root = Path(temporary)
        with tarfile.open(fileobj=io.BytesIO(archive)) as bundle:
            for member in bundle.getmembers():
                name = PurePosixPath(member.name)
                assert not name.is_absolute() and '..' not in name.parts, 'unsafe archive path'
                # The delivered Ensemble tree has no source symlinks. Refuse to
                # follow a future one silently into files outside this snapshot.
                assert member.isfile() or member.isdir(), 'unsupported archive member'
            bundle.extractall(root, filter='data')
        directory = root/prefix
        files = sorted(p for p in directory.rglob('*') if p.is_file())
        result = discover(directory)
        result.update(source_commit=source, source_tree=tree, source_prefix=prefix,
                      delivered_file_count=len(files), scope=__doc__,
                      checker_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest())
        return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_commit')
    parser.add_argument('--source-prefix', default='solutions/edition-2/main')
    arguments = parser.parse_args()
    result = evaluate(arguments.source_commit, arguments.source_prefix)
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
