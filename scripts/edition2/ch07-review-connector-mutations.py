#!/usr/bin/env python3
"""Delete Connector protections in isolated copies after a complete positive run."""
import argparse
import hashlib
import importlib.util
import json
from pathlib import Path
import shutil
import tempfile

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('connector_probe', HERE / 'ch07-review-connector.py')
probe = importlib.util.module_from_spec(spec)
spec.loader.exec_module(probe)
MUTATIONS = [
    ('count-limit', 'len(c.queue) < ensemble.WatchItems', 'len(c.queue) <= ensemble.WatchItems', 1,
     'count-and-peer', '257th outgoing item accepted'),
    ('byte-limit', 'len(data) <= ensemble.WatchBytes-c.bytes', 'true', 1,
     'encoded-bytes', 'aggregate byte limit +1 accepted'),
    ('drain-accounting', 'c.bytes -= len(item.data)', '// deleted drain byte accounting', 1,
     'drain-capacity', 'drained capacity retained'),
    ('late-send', 'if c.ctx.Err() != nil {\n\t\t\tc.mu.Unlock()\n\t\t\treturn false',
     'if false {\n\t\t\tc.mu.Unlock()\n\t\t\treturn false', 1,
     'selected-close', 'selected sender enqueued after disconnect'),
    ('release-pause', 'pause.Close()', '// deleted pause release', 1,
     'count-and-peer', 'overflow retained owning pause'),
    ('write-bound', '_ = c.socket.SetWriteDeadline(time.Now().Add(5 * time.Second))',
     '_ = c.socket.SetWriteDeadline(time.Now().Add(60 * time.Second))', 1,
     'write-deadline-join', 'write deadline is not bounded to 5s'),
    ('dead-peer-bound', 'time.Now().Add(25 * time.Second)', 'time.Now().Add(60 * time.Second)', 2,
     'dead-peer-deadline', 'dead peer outlived 30-second connection bound'),
]

def evaluate(source):
    original = (source / 'gui/connector.go').read_text()
    for name, old, _, count, _, _ in MUTATIONS:
        if original.count(old) != count:
            raise RuntimeError(f'{name}: expected {count} unique anchors, found {original.count(old)}')
    baseline = probe.evaluate(source)
    rows = []
    if baseline['passed']:
        for name, old, new, count, group, expected in MUTATIONS:
            with tempfile.TemporaryDirectory(prefix='ch07-connector-mutation-') as tmp:
                root = Path(tmp) / 'source'
                shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence', '__pycache__', 'node_modules'))
                path = root / 'gui/connector.go'
                text = path.read_text()
                if text.count(old) != count:
                    raise RuntimeError(f'{name}: expected {count} unique anchors, found {text.count(old)}')
                path.write_text(text.replace(old, new))
                result = probe.evaluate(root, [group])
                command = result['checks'][0]['command']
                rows.append(dict(id=name, group=group, intended_failure=expected,
                                 passed=not result['passed'] and expected in command['stdout'],
                                 command=command))
    return dict(scope=__doc__, baseline=baseline, checks=rows,
                checker_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                passed=baseline['passed'] and len(rows)==len(MUTATIONS) and all(r['passed'] for r in rows))

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    args = parser.parse_args()
    result = evaluate(args.source.resolve(strict=True))
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
