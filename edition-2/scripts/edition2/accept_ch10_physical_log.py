#!/usr/bin/env python3
"""Real regular-file 1 GiB session log boundary; distinct from generated Reader proof."""
import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

from accept_ch10_lifecycle import bind, digest, save_receipt

HERE = Path(__file__).resolve().parent
FIXTURES = ('ch10-physical-log_test.go', 'ch10-remaining-aggregate_test.go')
MIN_FREE = (1 << 30) + (512 << 20)


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, source_files = bind(source, args.source_commit)
    paths = [Path(__file__).resolve(), HERE / 'accept_ch10_lifecycle.py', *(HERE / p for p in FIXTURES)]
    checkers = {p.name: digest(p) for p in paths}
    work = Path(tempfile.mkdtemp(prefix='ch10-physical-log-'))
    payload = work / 'owned-payloads'
    payload.mkdir()
    result = dict(source_commit=commit, source_files=source_files, source_directory=str(source),
                  checker_files=checkers, work_directory=str(work), checks=[], complete=False,
                  passed=False, full_chapter_acceptance=False, scope=__doc__)

    def save():
        save_receipt(args.receipt, result)

    def run(command, execute=False, mutation=False):
        row = dict(command=command, started_utc=datetime.now(timezone.utc).isoformat(),
                   disk_free_before=shutil.disk_usage(work).free, complete=False)
        result['checks'].append(row)
        if execute and row['disk_free_before'] < MIN_FREE:
            row.update(resource_refusal=True, required_free_bytes=MIN_FREE, passed=False)
            save()
            return False
        row['environment_overrides'] = dict(GOMEMLIMIT='1GiB', GOGC='25', TMPDIR=str(payload))
        save()
        try:
            p = subprocess.run(command, cwd=source, env=dict(os.environ, **row['environment_overrides']),
                               text=True, capture_output=True, timeout=300)
            passed = p.returncode == 0 and (command[0] != 'gofmt' or not p.stdout)
            if execute:
                if mutation:
                    passed = (p.returncode != 0 and 'physical log one-over limit accepted' in p.stdout
                              and '--- FAIL: TestCh10PhysicalLogBoundary' in p.stdout
                              and 'panic: test timed out' not in p.stdout)
                else:
                    passed = passed and '--- PASS: TestCh10PhysicalLogBoundary' in p.stdout
            row.update(exit=p.returncode, stdout=p.stdout, stderr=p.stderr, complete=True, passed=passed)
        except subprocess.TimeoutExpired as exc:
            def text(value):
                return value.decode(errors='replace') if isinstance(value, bytes) else value
            row.update(timeout=True, stdout=text(exc.stdout), stderr=text(exc.stderr), passed=False)
        finally:
            row.update(finished_utc=datetime.now(timezone.utc).isoformat(), disk_free_after=shutil.disk_usage(work).free)
            save()
        return row['passed']

    save()
    good = False
    try:
        replacements = {}
        for fixture in FIXTURES:
            destination = source / fixture.replace('-', '_')
            if destination.exists():
                raise ValueError('overlay collision')
            local = work / fixture
            local.write_bytes((HERE / fixture).read_bytes())
            replacements[str(destination)] = str(local)
        overlay = work / 'positive.json'
        overlay.write_text(json.dumps({'Replace': replacements}))
        binary = work / 'physical-log.test'
        good = all(run(command) for command in [
            ['gofmt', '-l', *replacements.values()],
            ['go', 'vet', '-overlay', str(overlay), './...'],
            ['go', 'test', './...', '-count=1'],
            ['go', 'test', '-c', '-overlay', str(overlay), '-o', str(binary), '.'],
        ])
        if good:
            result['binary_sha256'] = digest(binary)
            save()
            good = run(['/usr/bin/time', '-l', str(binary), '-test.v', '-test.timeout=4m',
                        '-test.run=^TestCh10PhysicalLogBoundary$'], execute=True)
        if good and args.audit:
            target = source / 'internal/eventlog/log.go'
            original = target.read_text()
            start = original.index('func Stream(')
            end = original.index('// Resume opens', start)
            fragment = original[start:end]
            if fragment.count('1<<30') != 3:
                raise ValueError('physical log mutation anchors changed')
            changed = work / 'log-plus-one.go'
            changed.write_text(original[:start] + fragment.replace('1<<30', '((1<<30)+1)') + original[end:])
            result['mutation'] = dict(path=str(target.relative_to(source)), original_sha256=digest(target),
                                      changed_sha256=digest(changed), old='1<<30', new='((1<<30)+1)', count=3,
                                      expected_assertion='physical log one-over limit accepted')
            replacements[str(target)] = str(changed)
            overlay = work / 'mutation.json'
            overlay.write_text(json.dumps({'Replace': replacements}))
            binary = work / 'physical-log-mutated.test'
            good = run(['go', 'test', '-c', '-overlay', str(overlay), '-o', str(binary), '.'])
            if good:
                result['mutation']['binary_sha256'] = digest(binary)
                save()
                good = run(['/usr/bin/time', '-l', str(binary), '-test.v', '-test.timeout=4m',
                            '-test.run=^TestCh10PhysicalLogBoundary$'], execute=True, mutation=True)
        if bind(source, commit)[1] != source_files or checkers != {p.name: digest(p) for p in paths}:
            raise ValueError('source/checker changed during run')
        result.update(complete=True, passed=good)
    finally:
        # Only this runner-created payload directory is disposable. Retain its
        # small source adapters, binaries and receipt; never traverse user data.
        residual = [{'path': str(p.relative_to(payload)), 'size': p.stat().st_size}
                    for p in payload.rglob('*') if p.is_file()]
        result['payload_cleanup'] = dict(directory=str(payload), residual_files_before_cleanup=residual,
                                         disk_free_before=shutil.disk_usage(work).free)
        shutil.rmtree(payload)
        result['payload_cleanup'].update(removed=not payload.exists(), disk_free_after=shutil.disk_usage(work).free)
        result['remaining_work_files'] = [{'path': str(p.relative_to(work)), 'size': p.stat().st_size}
                                          for p in work.rglob('*') if p.is_file()]
        save()
    return good


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory', type=Path)
    parser.add_argument('--source-commit', required=True)
    parser.add_argument('--receipt', type=Path, required=True)
    parser.add_argument('--audit', action='store_true')
    raise SystemExit(0 if evaluate(parser.parse_args()) else 1)
