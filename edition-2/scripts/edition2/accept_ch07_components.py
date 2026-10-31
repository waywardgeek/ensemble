#!/usr/bin/env python3
"""Narrow Chapter 7 queue/projection components. Browser and full acceptance remain separate.

Large exact-byte fixtures run without race instrumentation; ordering/lifetime and
GUI projection fixtures use it. Only disposable copies are edited; no paid calls.
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
    'watch-byte-capacity': ('./internal/llm', '^TestC7WatchEncodedByteCapacity$', False),
    'pause-publication-order': ('./internal/llm', '^TestC7PausePublicationBeforeAck$', True),
    'projection-bound-invalidation': ('./internal/llm', '^TestC7ProjectionOverflowInvalidatesWatch$', False),
    'watch-selected-close': ('./internal/llm', '^TestC7SelectedRecipientClose$', True),
    'typed-display-projection': ('gui', '^TestC7GUIProjection$', True),
}


def install(root):
    for chapter in ('06', '07'):
        (root / 'internal/llm' / f'independent_ch{chapter}_internal_test.go').write_bytes(
            (HERE / f'ch{chapter}_internal_test.go.txt').read_bytes())
    (root / 'gui/independent_ch07_gui_test.go').write_bytes((HERE / 'ch07_gui_test.go.txt').read_bytes())


def evaluate(source, groups=None):
    with tempfile.TemporaryDirectory(prefix='ch07-independent-components-') as tmp:
        root = Path(tmp) / 'source'
        shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence', '__pycache__', 'node_modules'))
        # Retain and bind embedded static assets as well as Go/module files.
        hashes = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(root.rglob('*')) if p.is_file()}
        install(root)
        rows = []
        for name in groups or GROUPS:
            package, pattern, race = GROUPS[name]
            cwd = root / 'gui' if package == 'gui' else root
            cmd = ['go', 'test', '-json', '-count=1', '-timeout=90s', '-run', pattern]
            if race:
                cmd.append('-race')
            cmd.append('.' if package == 'gui' else package)
            result = command(cmd, cwd, timeout=120)
            ran = any(json.loads(line).get('Action') == 'run'
                      for line in result['stdout'].splitlines() if line.startswith('{'))
            rows.append(dict(id=name, passed=result['exit'] == 0 and ran, command=result))
        return dict(scope=__doc__, source=str(source), source_files=hashes,
                    checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in
                                   (Path(__file__), HERE / 'ch07_internal_test.go.txt', HERE / 'ch07_gui_test.go.txt', HERE / 'ch06_internal_test.go.txt')},
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
