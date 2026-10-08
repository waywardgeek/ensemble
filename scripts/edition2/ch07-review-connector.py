#!/usr/bin/env python3
"""Independent actual Connector queue, socket deadline and worker lifetime probes."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command

HERE = Path(__file__).resolve().parent
GROUPS = {
    'count-and-peer': ('^TestC7ConnectorQueueCountAndPeer$', True),
    'drain-capacity': ('^TestC7ConnectorQueueDrain$', True),
    'encoded-bytes': ('^TestC7ConnectorQueueBytes$', False),
    'selected-close': ('^TestC7ConnectorSelectedClose$', True),
    'write-deadline-join': ('^TestC7ConnectorWriteDeadlineAndJoin$', True),
    'dead-peer-deadline': ('^TestC7ConnectorDeadPeerDeadline$', True),
}

def evaluate(source, groups=None):
    with tempfile.TemporaryDirectory(prefix='ch07-connector-review-') as tmp:
        root = Path(tmp) / 'source'
        shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence', '__pycache__', 'node_modules'))
        hashes = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(root.rglob('*')) if p.is_file()}
        fixture = HERE / 'ch07-review-connector_test.go.txt'
        (root / 'gui/independent_connector_test.go').write_bytes(fixture.read_bytes())
        rows = []
        for name in groups or GROUPS:
            pattern, race = GROUPS[name]
            cmd = ['go', 'test', '-json', '-count=1', '-timeout=60s', '-run', pattern]
            if race:
                cmd.append('-race')
            cmd.append('.')
            result = command(cmd, root / 'gui', timeout=90)
            ran = any(json.loads(line).get('Action') == 'run'
                      for line in result['stdout'].splitlines() if line.startswith('{'))
            rows.append(dict(id=name, passed=result['exit'] == 0 and ran, command=result))
        return dict(scope=__doc__, source=str(source), source_files=hashes,
                    checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                                   for p in (Path(__file__), fixture, HERE / 'accept_ch05.py')},
                    passed=bool(rows) and all(r['passed'] for r in rows), checks=rows)

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
