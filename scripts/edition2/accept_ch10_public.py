#!/usr/bin/env python3
"""Prepared Chapter 10 public session controls; genuine runtime required.

Uses a small external Go module and the documented public API. Does not copy
student source, invoke providers, or establish full Chapter 10 acceptance.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def identities(source):
    return {str(p.relative_to(source)): digest(p) for p in sorted(source.rglob('*'))
            if p.is_file() and '.git' not in p.parts and
            (p.suffix == '.go' or p.name in ('go.mod', 'go.sum', 'session-api.md', 'persistence-format.md'))}


def evaluate(source, pattern):
    before = identities(source)
    fixtures = [HERE / name for name in ('ch10-session-public_test.go', 'ch10-session-roundtrip_test.go',
                                       'ch10-session-limits_test.go', 'ch10-session-skills_test.go')]
    checker_paths = [Path(__file__), *fixtures, HERE / 'ch10-semantic-cases.py', HERE / 'ch10-limit-cases.py',
                     HERE / 'accept_ch10.py', HERE / 'accept_ch09.py']
    checker_before = {p.name: digest(p) for p in checker_paths}
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch10-public-') as directory:
        work = Path(directory)
        (work / 'go.mod').write_text('module example.com/ch10-independent-consumer\n\ngo 1.25\n\n'
                                   'require example.com/ensemble v0.0.0\nreplace example.com/ensemble => '
                                   + json.dumps(str(source)) + '\n')
        for fixture in fixtures:
            shutil.copyfile(fixture, work / fixture.name)
        for command in [
            ['go', 'mod', 'tidy'], ['gofmt', '-l'] + [p.name for p in fixtures], ['go', 'vet', './...'],
            ['go', 'test', '-race', '-count=1', '-timeout=90s', '-run', pattern, '-v', './...'],
        ]:
            env = dict(os.environ, CH10_SEMANTIC_CASES=str(HERE / 'ch10-semantic-cases.py'),
                       CH10_LIMIT_CASES=str(HERE / 'ch10-limit-cases.py'))
            result = subprocess.run(command, cwd=work, env=env, text=True, capture_output=True, timeout=180)
            passed = result.returncode == 0 and (command[0] != 'gofmt' or not result.stdout)
            rows.append(dict(command=command, exit=result.returncode, stdout=result.stdout,
                             stderr=result.stderr, passed=passed))
            if not passed:
                break
    assert before == identities(source), 'source changed during public session checks'
    assert checker_before == {p.name: digest(p) for p in checker_paths}, 'checker changed during public session checks'
    return dict(passed=len(rows) == 4 and all(r['passed'] for r in rows), checks=rows,
                source_directory=str(source), source_files=before, test_pattern=pattern,
                checker_files=checker_before,
                scope=__doc__, full_chapter_acceptance=False)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory', type=Path)
    parser.add_argument('--run', default='^TestCh10Public')
    parser.add_argument('--receipt', type=Path)
    args = parser.parse_args()
    result = evaluate(args.source_directory.resolve(strict=True), args.run)
    if args.receipt:
        args.receipt.write_text(json.dumps(result, indent=2) + '\n')
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
