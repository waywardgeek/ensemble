#!/usr/bin/env python3
"""Physical512MiB checkpoint/origin and256MiB canonical-state controls; no1GiB log proof."""
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
FIXTURES = ('ch10-large-state-files_test.go', 'ch10-remaining-aggregate_test.go')
CASES = {
    'canonical': '^TestCh10LargeCanonicalState$',
    'checkpoint': '^TestCh10LargePhysicalStateFiles$/^checkpoint.json$',
    'origin': '^TestCh10LargePhysicalStateFiles$/^origin.json$',
}
# Payload is at most one512MiB file, or origin+checkpoint each about256MiB.
# Reserve an additional256MiB for ordinary concurrent filesystem activity.
MIN_FREE = 768 << 20


def evaluate(args):
    source = args.source_directory.resolve(strict=True)
    commit, source_files = bind(source, args.source_commit)
    paths = [Path(__file__).resolve(), HERE / 'accept_ch10_lifecycle.py', *(HERE / p for p in FIXTURES)]
    checkers = {p.name: digest(p) for p in paths}
    work = Path(tempfile.mkdtemp(prefix='ch10-large-state-'))
    result = dict(source_commit=commit, source_files=source_files, source_directory=str(source),
                  checker_files=checkers, work_directory=str(work), checks=[], complete=False,
                  passed=False, full_chapter_acceptance=False, scope=__doc__)

    def save():
        save_receipt(args.receipt, result)

    def run(command, large=False):
        row = dict(command=command, started_utc=datetime.now(timezone.utc).isoformat(),
                   disk_free_before=shutil.disk_usage(work).free, complete=False)
        result['checks'].append(row)
        if large and row['disk_free_before'] < MIN_FREE:
            row.update(resource_refusal=True, required_free_bytes=MIN_FREE, passed=False)
            save()
            return False
        save()
        env = dict(os.environ, GOMEMLIMIT='1GiB', GOGC='25', TMPDIR=str(work))
        row['environment_overrides'] = {k:env[k] for k in ('GOMEMLIMIT','GOGC','TMPDIR')}
        try:
            p = subprocess.run(command, cwd=source, env=env, text=True, capture_output=True, timeout=600)
            row.update(exit=p.returncode, stdout=p.stdout, stderr=p.stderr, complete=True,
                       passed=p.returncode == 0 and (command[0] != 'gofmt' or not p.stdout))
        except subprocess.TimeoutExpired as exc:
            def text(x):
                return x.decode(errors='replace') if isinstance(x,bytes) else x
            row.update(timeout=True, passed=False, stdout=text(exc.stdout), stderr=text(exc.stderr))
        finally:
            row.update(finished_utc=datetime.now(timezone.utc).isoformat(),
                       disk_free_after=shutil.disk_usage(work).free)
            save()
        return row['passed']

    save()
    replacements = {}
    for fixture in FIXTURES:
        destination = source / fixture.replace('-','_')
        if destination.exists():
            raise ValueError('overlay collision')
        local = work / fixture
        local.write_bytes((HERE / fixture).read_bytes())
        replacements[str(destination)] = str(local)
    overlay = work / 'positive.json'
    overlay.write_text(json.dumps({'Replace':replacements}))
    binary = work / 'large-state.test'
    good = all(run(command) for command in [
        ['gofmt','-l',*replacements.values()],
        ['go','vet','-overlay',str(overlay),'./...'],
        ['go','test','./...','-count=1'],
        ['go','test','-c','-overlay',str(overlay),'-o',str(binary),'.'],
    ])
    if good:
        result['binary_sha256'] = digest(binary)
        save()
        for case in (CASES if args.case == 'all' else [args.case]):
            if not run(['/usr/bin/time','-l',str(binary),'-test.v','-test.timeout=8m','-test.run='+CASES[case]],large=True):
                good = False
                break
    if bind(source,commit)[1] != source_files or checkers != {p.name:digest(p) for p in paths}:
        raise ValueError('source/checker changed during run')
    result.update(complete=True,passed=good)
    save()
    return good


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory',type=Path)
    parser.add_argument('--source-commit',required=True)
    parser.add_argument('--receipt',type=Path,required=True)
    parser.add_argument('--case',choices=['all',*CASES],default='all')
    args = parser.parse_args()
    raise SystemExit(0 if evaluate(args) else 1)
