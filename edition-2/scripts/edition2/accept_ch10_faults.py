#!/usr/bin/env python3
"""Chapter 10 source-bound real-I/O fault and checkpoint lifetime overlay.

The added methods are review-only adapters, not required public API. No student
file is written. Requires a coherent frozen source and proves injected barriers
were reached. This subset does not establish the complete chapter matrix.
"""
import argparse
import hashlib
import json
import subprocess
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
FILES = {
    'ch10_independent_faults_test.go': 'ch10-session-faults_test.go',
    'internal/persistence/ch10_independent_seam.go': 'ch10-fault-store-seam.go',
    'internal/eventlog/ch10_independent_seam.go': 'ch10-fault-log-seam.go',
}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def files(source):
    return {str(p.relative_to(source)): digest(p) for p in sorted(source.rglob('*'))
            if p.is_file() and 'evidence' not in p.relative_to(source).parts
            and (p.suffix in ('.go', '.js', '.html', '.css') or
                 p.name in ('go.mod', 'go.sum', 'session-api.md', 'persistence-format.md'))}


def bind(source, revision):
    root = Path(subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], cwd=source, text=True).strip())
    prefix = source.relative_to(root).as_posix() + '/'
    commit = subprocess.check_output(['git', 'rev-parse', revision + '^{commit}'], cwd=root, text=True).strip()
    current = files(source)
    if not current:
        raise ValueError('empty runtime source map')
    for name, sha in current.items():
        raw = subprocess.check_output(['git', 'show', commit + ':' + prefix + name], cwd=root)
        if hashlib.sha256(raw).hexdigest() != sha:
            raise ValueError('runtime source differs from frozen revision: ' + name)
    # A complete current map alone could hide a removed historical source file.
    tracked = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', commit, '--', prefix], cwd=root, text=True)
    for path in tracked.splitlines():
        name = path[len(prefix):]
        p = Path(name)
        if 'evidence' not in p.parts and (p.suffix in ('.go', '.js', '.html', '.css') or
                    p.name in ('go.mod', 'go.sum', 'session-api.md', 'persistence-format.md')):
            if name not in current:
                raise ValueError('missing frozen runtime source: ' + name)
    return commit, current


def evaluate(source, revision, pattern):
    commit, before = bind(source, revision)
    checker_paths = [Path(__file__).resolve(), *(HERE / name for name in FILES.values())]
    checker_before = {p.name: digest(p) for p in checker_paths}
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch10-fault-review-') as directory:
        work = Path(directory)
        replacements = {}
        for target, fixture in FILES.items():
            virtual = source / target
            if virtual.exists():
                raise ValueError('review-only overlay collides with student file: ' + target)
            local = work / fixture
            local.write_bytes((HERE / fixture).read_bytes())
            replacements[str(virtual)] = str(local)
        overlay = work / 'overlay.json'
        overlay.write_text(json.dumps({'Replace': replacements}))
        commands = [
            ['gofmt', '-l', *replacements.values()],
            ['go', 'vet', '-overlay', str(overlay), '.'],
            ['go', 'test', '-overlay', str(overlay), '-race', '-count=1', '-timeout=60s', '-run', pattern, '-v', '.'],
        ]
        for command in commands:
            try:
                p = subprocess.run(command, cwd=source, text=True, capture_output=True, timeout=180)
                passed = p.returncode == 0 and (command[0] != 'gofmt' or not p.stdout)
                row = dict(command=command, exit=p.returncode, stdout=p.stdout, stderr=p.stderr, passed=passed)
            except subprocess.TimeoutExpired as e:
                def text(value):
                    return value.decode(errors='replace') if isinstance(value, bytes) else (value or '')
                row = dict(command=command, exit=None, stdout=text(e.stdout), stderr=text(e.stderr), passed=False, timeout_seconds=180)
            rows.append(row)
            if not row['passed']:
                break
    unchanged = before == files(source)
    checker_unchanged = checker_before == {p.name: digest(p) for p in checker_paths}
    return dict(passed=len(rows) == 3 and all(r['passed'] for r in rows) and unchanged and checker_unchanged,
                source_commit=commit, source_directory=str(source), source_files=before,
                checker_files=checker_before, source_unchanged=unchanged, checker_unchanged=checker_unchanged,
                checks=rows, scope=__doc__, test_pattern=pattern, full_chapter_acceptance=False)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory', type=Path)
    parser.add_argument('--source-commit', required=True)
    parser.add_argument('--run', default='^TestCh10Review')
    parser.add_argument('--receipt', type=Path)
    args = parser.parse_args()
    result = evaluate(args.source_directory.resolve(strict=True), args.source_commit, args.run)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
