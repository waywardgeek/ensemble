#!/usr/bin/env python3
"""Finite affected-check follow-up; retained gate and current large receipts stay separate."""
import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil
import subprocess

from accept_ch10_lifecycle import bind, digest, save_receipt

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, files = bind(source, args.source_commit)
    build = json.loads(args.build_binding.read_text())
    if build['source_revision'] != commit or not files:
        raise ValueError('final executable source association mismatch')
    for name, expected in files.items():
        if build['sources'].get('solutions/edition-2/main/' + name) != expected:
            raise ValueError('incomplete final executable source map: ' + name)
    cli = Path(build['binaries']['cli']['path'])
    executable = build['binaries']['cli']['sha256']
    if digest(cli) != executable or build['builds']['cli']['exit'] != 0 or build['builds']['cli']['binary_sha256'] != executable:
        raise ValueError('final CLI build or hash mismatch')
    checkers = {p.name: digest(p) for p in HERE.iterdir() if p.is_file() and 'ch10' in p.name}
    result = dict(source_commit=commit, source_directory=str(source), source_files=files,
                  checker_files=checkers, binary=dict(path=str(cli), sha256=executable),
                  build_binding=dict(path=str(args.build_binding), sha256=digest(args.build_binding)),
                  checks=[], complete=False, passed=False, full_chapter_acceptance=False, scope=__doc__)
    def save():
        save_receipt(args.receipt, result)
    def subreceipt(name):
        return str(args.receipt.parent / ('ch10-final-' + name + '-57d4aac.json'))
    common = [str(source), '--source-commit', commit]
    commands = [
        ('storage-bounds', ['python3', str(HERE/'accept_ch10_storage_bounds.py'), *common, '--receipt', subreceipt('storage-bounds')], True),
        ('aggregate-and-public', ['python3', str(HERE/'accept_ch10_remaining.py'), *common, '--run', '^TestCh10Remaining(AggregateEntries|AggregateParts|SeenRequests|PublicConstructionFacts|StoreLeaves)$', '--receipt', subreceipt('aggregate-public')], True),
        ('independent-job-tail', ['python3', str(HERE/'accept_ch10_lifecycle.py'), *common, '--receipt', subreceipt('job-tail')], True),
        ('physical-records', ['python3', str(HERE/'accept_ch10_record_bounds.py'), str(cli), '--source-directory', str(source), '--receipt', subreceipt('record-bounds')], True),
        ('published-checker-oracle', ['python3', str(HERE/'accept_ch10.py'), '--self-test'], True),
        ('orchestration-controls', ['python3', '-m', 'unittest', 'discover', '-s', str(HERE), '-p', 'test_ch10_retained_receipts.py', '-v'], True),
        ('integration-identity-controls', ['python3', '-m', 'unittest', 'discover', '-s', str(HERE), '-p', 'test_ch10_final_integration.py', '-v'], True),
        ('inherited-diagnostic', ['make', 'grade-dir', 'CH=11', 'DIR='+str(source)], False),
    ]
    save()
    for name, command, required in commands:
        row = dict(id=name, command=command, cwd=str(ROOT), required_for_new_contract=required,
                   started_utc=datetime.now(timezone.utc).isoformat(), disk_free_before=shutil.disk_usage(ROOT).free, complete=False)
        result['checks'].append(row)
        save()
        try:
            p = subprocess.run(command, cwd=ROOT, env=dict(os.environ, GOMEMLIMIT='1GiB', GOGC='50'),
                               capture_output=True, text=True, timeout=900)
            row.update(exit=p.returncode, stdout=p.stdout, stderr=p.stderr, passed=p.returncode==0, complete=True)
        except subprocess.TimeoutExpired as exc:
            def text(value):
                return value.decode(errors='replace') if isinstance(value,bytes) else value
            row.update(timeout=True, passed=False, stdout=text(exc.stdout), stderr=text(exc.stderr))
        finally:
            row.update(finished_utc=datetime.now(timezone.utc).isoformat(), disk_free_after=shutil.disk_usage(ROOT).free)
            save()
    if bind(source,commit)[1] != files or digest(cli) != executable or digest(args.build_binding) != result['build_binding']['sha256']:
        raise ValueError('runtime/binary association changed during follow-up')
    if checkers != {p.name:digest(p) for p in HERE.iterdir() if p.is_file() and 'ch10' in p.name}:
        raise ValueError('checker inventory changed during follow-up')
    result.update(complete=True, passed=all(x['passed'] for x in result['checks'] if x['required_for_new_contract']))
    save()
    return result['passed']


if __name__ == '__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('source_directory', type=Path)
    p.add_argument('--source-commit', required=True)
    p.add_argument('--build-binding', required=True, type=Path)
    p.add_argument('--receipt', required=True, type=Path)
    raise SystemExit(0 if evaluate(p.parse_args()) else 1)
