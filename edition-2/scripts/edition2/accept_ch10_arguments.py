#!/usr/bin/env python3
"""Narrow Ch10 public argument/replay controls; local HTTP, no paid providers.

Binds every tracked new-only file before an external consumer builds. Never edits
student source. These checks complement retained/remaining gates, not acceptance.
"""
import argparse
import ast
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile

from accept_ch10_lifecycle import digest, save_receipt

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
PREFIX = 'solutions/edition-2/main/'
FIXTURES = ('ch10-session-public_test.go', 'ch10-session-roundtrip_test.go',
            'ch10-session-limits_test.go', 'ch10-session-skills_test.go',
            'ch10-arguments-public_test.go')


def bind_full(source, revision):
    commit = subprocess.check_output(['git', 'rev-parse', revision + '^{commit}'],
                                     cwd=REPO, text=True).strip()
    actual = {}
    process = subprocess.Popen(['git', 'archive', commit, PREFIX], cwd=REPO,
                               stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    try:
        with tarfile.open(fileobj=process.stdout, mode='r|') as archive:
            for member in archive:
                if member.isdir():
                    continue
                if not member.isfile() or not member.name.startswith(PREFIX):
                    raise ValueError('unsupported new-only archive member')
                name = member.name[len(PREFIX):]
                relative = Path(name)
                if relative.is_absolute() or '..' in relative.parts:
                    raise ValueError('unsafe source path')
                local = source / relative
                if local.is_symlink() or not local.is_file() or not local.resolve().is_relative_to(source):
                    raise ValueError('missing/escaping frozen file: ' + name)
                expected = hashlib.sha256()
                with archive.extractfile(member) as stream:
                    for block in iter(lambda: stream.read(1024 * 1024), b''):
                        expected.update(block)
                actual[name] = digest(local)
                if actual[name] != expected.hexdigest():
                    raise ValueError('frozen source mismatch: ' + name)
        if process.wait() != 0:
            raise ValueError('git archive failed: ' + process.stderr.read().decode())
    finally:
        process.stdout.close()
        if process.poll() is None:
            process.kill()
            process.wait()
        process.stderr.close()
    if not actual or 'go.mod' not in actual:
        raise ValueError('empty/incomplete new-only source')
    # Untracked build inputs cannot silently augment the frozen tracked tree.
    for path in source.rglob('*'):
        if not path.is_file() or '.git' in path.relative_to(source).parts:
            continue
        if path.suffix in ('.go', '.c', '.h', '.s', '.syso') or path.name in ('go.mod', 'go.sum', 'go.work'):
            if str(path.relative_to(source)) not in actual:
                raise ValueError('unbound build input: ' + str(path.relative_to(source)))
    return commit, actual


def checker_paths():
    pending = [Path(__file__).resolve(), HERE / 'ch10-arguments-cases.py']
    found = set(HERE / name for name in FIXTURES)
    while pending:
        path = pending.pop()
        if path in found:
            continue
        found.add(path)
        for node in ast.walk(ast.parse(path.read_text())):
            names = ([node.module] if isinstance(node, ast.ImportFrom) and node.module else
                     [a.name for a in node.names] if isinstance(node, ast.Import) else [])
            for name in names:
                local = HERE / (name.split('.')[0] + '.py')
                if local.is_file() and local not in found:
                    pending.append(local)
    return sorted(found)


def evaluate(source, revision, receipt):
    # All binding failures happen before the first receipt write or compiler.
    source = source.resolve(strict=True)
    commit, before = bind_full(source, revision)
    paths = checker_paths()
    checkers = {p.name: digest(p) for p in paths}
    result = dict(scope=__doc__, source_revision=commit, source_directory=str(source),
                  source_files=before, checker_files=checkers, checks=[],
                  complete=False, passed=False, full_chapter_acceptance=False)
    save_receipt(receipt, result)
    try:
        with tempfile.TemporaryDirectory(prefix='ch10-arguments-') as directory:
            work = Path(directory)
            (work / 'go.mod').write_text('module example.com/ch10-argument-consumer\n\ngo 1.25\n\n'
                'require example.com/ensemble v0.0.0\nreplace example.com/ensemble => '
                + json.dumps(str(source)) + '\n')
            for name in FIXTURES:
                (work / name).write_bytes((HERE / name).read_bytes())
            env = dict(os.environ, GOWORK='off', CH10_ARGUMENT_CASES=str(HERE / 'ch10-arguments-cases.py'))
            commands = [['gofmt', '-l', *FIXTURES], ['go', 'mod', 'tidy'],
                        ['go', 'vet', './...'],
                        ['go', 'test', '-count=1', '-timeout=120s', '-run', '^TestCh10Arguments', '-v', './...']]
            for command in commands:
                result['active_command'] = command
                save_receipt(receipt, result)
                process = subprocess.run(command, cwd=work, env=env, capture_output=True, text=True, timeout=240)
                passed = process.returncode == 0 and (command[0] != 'gofmt' or not process.stdout)
                result['checks'].append(dict(command=command, exit=process.returncode,
                                             stdout=process.stdout, stderr=process.stderr, passed=passed))
                result['active_command'] = None
                save_receipt(receipt, result)
                if not passed:
                    break
        unchanged = bind_full(source, commit)[1] == before
        checker_unchanged = checkers == {p.name: digest(p) for p in paths}
        result.update(complete=True, source_unchanged=unchanged, checker_unchanged=checker_unchanged,
                      passed=unchanged and checker_unchanged and len(result['checks']) == 4
                      and all(row['passed'] for row in result['checks']))
    except BaseException as error:
        result['error'] = type(error).__name__ + ': ' + str(error)
        save_receipt(receipt, result)
        raise
    save_receipt(receipt, result)
    return result


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory', type=Path)
    parser.add_argument('--source-commit', required=True)
    parser.add_argument('--receipt', type=Path, required=True)
    args = parser.parse_args()
    outcome = evaluate(args.source_directory, args.source_commit, args.receipt)
    print(json.dumps(dict(passed=outcome['passed'], complete=outcome['complete'], receipt=str(args.receipt))))
    raise SystemExit(0 if outcome['passed'] else 1)
