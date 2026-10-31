#!/usr/bin/env python3
"""Independent Chapter 7 public watch/pause groups; broader chapter gate remains pending.

Uses only the published public API and held local HTTP fixtures. No paid calls.
Source is copied outside the repository; receipt binds the exact copied files.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command

HERE = Path(__file__).resolve().parent
GROUPS = {
    'pause-ownership': '^TestC7PublicPauseOwnership$',
    'atomic-partial-watch': '^TestC7PublicAtomicPartialWatch$',
    'window-owned-copies': '^TestC7PublicWindowAndCopies$',
    'watch-count-overflow': '^TestC7PublicWatchOverflowAndClose$',
    'pause-tool-admission': '^TestC7PublicPauseAdmission$',
    'idle-watch-release': '^TestC7PublicIdleWatchRelease$',
    'watch-agent-isolation': '^TestC7PublicWatchAgentIsolation$',
}


def install(root):
    for chapter in ('06', '07'):
        (root / f'independent_ch{chapter}_public_test.go').write_bytes(
            (HERE / f'ch{chapter}_public_test.go.txt').read_bytes())


def evaluate(source, groups=None):
    with tempfile.TemporaryDirectory(prefix='ch07-independent-public-') as tmp:
        root = Path(tmp) / 'source'
        shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence', '__pycache__'))
        hashes = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(root.rglob('*')) if p.is_file() and
                  (p.suffix == '.go' or p.name in ('go.mod', 'go.sum'))}
        install(root)
        rows = []
        for name in groups or GROUPS:
            result = command(['go', 'test', '-json', '-race', '-count=1', '-timeout=90s',
                              '-run', GROUPS[name], '.'], root, timeout=120)
            ran = any(json.loads(line).get('Action') == 'run'
                      for line in result['stdout'].splitlines() if line.startswith('{'))
            rows.append(dict(id=name, passed=result['exit'] == 0 and ran, command=result))
        return dict(scope=__doc__, source=str(source), source_files=hashes,
                    checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in
                                   (Path(__file__), HERE / 'ch07_public_test.go.txt', HERE / 'ch06_public_test.go.txt')},
                    passed=all(r['passed'] for r in rows), checks=rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    parser.add_argument('--only', action='append', choices=GROUPS)
    args = parser.parse_args()
    result = evaluate(args.source.resolve(strict=True), args.only)
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
