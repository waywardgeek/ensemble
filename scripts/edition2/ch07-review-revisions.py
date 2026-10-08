#!/usr/bin/env python3
"""Immutable Chapter 7 comparative repair controls; no paid model or audible-speech claim."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile
from accept_ch05 import command

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
spec = importlib.util.spec_from_file_location('retained', HERE/'ch07-review-retained.py')
helper = importlib.util.module_from_spec(spec)
spec.loader.exec_module(helper)


def evaluate(revision):
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch07-comparative-repairs-') as directory:
        temporary = Path(directory)
        root = temporary/'source'
        revision, hashes = helper.extract(revision, root)
        logpath = 'solutions/edition-2/main/evidence/ch07/browser-anthropic-r1/session.log'
        original = subprocess.check_output(['git', 'show', '8cc87f2:'+logpath], cwd=REPO)
        assert (REPO/logpath).read_bytes() == original, 'original live log changed'
        log = temporary/'original-session.log'
        log.write_bytes(original)
        os.environ['C7_REAL_LOG'] = str(log)
        os.environ['C7_SNAPSHOT'] = str(temporary/'snapshot.json')
        (root/'independent_review_killed_test.go').write_bytes((HERE/'ch07-review-killed-test.go.txt').read_bytes())
        target = root/'gui/cmd/review/main.go'
        target.parent.mkdir(parents=True)
        target.write_bytes((HERE/'ch07-review-killed-server.go.txt').read_bytes())

        def run(name, args, cwd=root, timeout=180):
            result = command(args, cwd, timeout)
            rows.append(dict(id=name, passed=result['exit'] == 0, command=result))
            return result

        run('public-killed-window', ['go', 'test', '-race', '-count=1', '-run', '^TestC7ReviewKilledWindow$', '.'])
        if rows[-1]['passed']:
            target = root/'internal/llm/watch.go'
            text = target.read_text()
            anchor = '"job_ended", "job_killed", "turn_started"'
            assert text.count(anchor) == 1, 'killed filter mutation anchor absent or ambiguous'
            target.write_text(text.replace(anchor, '"job_ended", "turn_started"'))
            try:
                result = command(['go', 'test', '-race', '-count=1', '-run', '^TestC7ReviewKilledWindow$', '.'], root, 180)
            finally:
                target.write_text(text)
            rows.append(dict(id='killed-filter-deletion', passed=result['exit'] != 0 and 'killed outcome omitted from counted renderable window' in result['stdout'], command=result))
        run('actual-browser-killed-reconnect', ['python3', HERE/'ch07-review-killed-transport.py', revision], REPO)
        run('page-and-shared-speech-deletions', ['python3', HERE/'ch07-review-browser-mutations.py', root], REPO)
        gui = temporary/'gui'
        run('build-gui', ['go', 'build', '-o', gui, './cmd/ensemble-gui'], root/'gui')
        if rows[-1]['passed']:
            run('message-size-boundaries', ['python3', HERE/'ch07-review-message-size.py', gui], REPO)
            if rows[-1]['passed']:
                target = root/'gui/connector.go'
                text = target.read_text()
                anchor = 'c.stop("message exceeds 65536 bytes")'
                assert text.count(anchor) == 1, 'explanatory close mutation anchor absent or ambiguous'
                target.write_text(text.replace(anchor, 'c.stop("")'))
                try:
                    build = command(['go', 'build', '-o', gui, './cmd/ensemble-gui'], root/'gui', 120)
                    result = command(['python3', HERE/'ch07-review-message-size.py', gui], REPO, 180) if not build['exit'] else build
                finally:
                    target.write_text(text)
                observed = json.loads(result['stdout']) if not build['exit'] else {}
                failed = {r['id'] for r in observed.get('checks', []) if not r['passed']}
                expected = {'malformed-oversize', 'single-frame-131072-byte-close', 'single-frame-1048576-byte-close', 'fragmented-131072-byte-close'}
                rows.append(dict(id='explanatory-close-deletion', passed=build['exit'] == 0 and result['exit'] != 0 and failed == expected, expected_failures=sorted(expected), actual_failures=sorted(failed), build=build, command=result))
        formatted = run('fixture-format', ['gofmt', '-l', root/'independent_review_killed_test.go', root/'gui/cmd/review/main.go'])
        rows[-1]['passed'] &= not formatted['stdout'].strip()
        for module in (root, root/'gui'):
            label = str(module.relative_to(root))
            run(label+'/vet-with-fixtures', ['go', 'vet', './...'], module)
            run(label+'/tests-with-fixtures', ['go', 'test', './...', '-count=1', '-timeout=120s'], module)
    checker_paths = [p for p in HERE.glob('ch07-review-*') if p.is_file() and any(s in p.name for s in ('revision', 'browser', 'speech', 'killed', 'message-size', 'retained'))]
    return dict(scope=__doc__, source_revision=revision, source_files=hashes, original_log_revision='8cc87f2', original_log_sha256=hashlib.sha256(original).hexdigest(),
                checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in checker_paths}, passed=all(r['passed'] for r in rows), checks=rows)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('revision')
    args = parser.parse_args()
    result = evaluate(args.revision)
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
