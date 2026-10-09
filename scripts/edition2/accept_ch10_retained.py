#!/usr/bin/env python3
"""Run the unchanged preceding behavioral gate with durable stage receipts.

This is retained behavior only, not complete Chapter 10 acceptance. A private
fixture that no longer matches the new owner interface may require a separately
reviewed adapter. Every original failure remains in its first receipt.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import traceback

import accept_ch09_gate as gate
from ch10_retained_adapters import prepare


def save(path, value):
    temporary = path.with_name(path.name + '.pending')
    with temporary.open('w') as out:
        json.dump(value, out, indent=2)
        out.write('\n')
        out.flush()
        os.fsync(out.fileno())
    os.replace(temporary, path)


class Progress:
    def __init__(self, receipt, maintain_cache=False, reserve_mib=768):
        self.receipt = receipt
        self.maintain_cache = maintain_cache
        self.reserve = reserve_mib * 1024 * 1024
        self.state = dict(passed=False, complete_run=False, full_chapter_acceptance=False,
                          scope=__doc__, checks=[], maintenance=[], active_check=None)

    def __call__(self, event):
        phase = event['phase']
        if phase == 'source':
            self.state.update(source_revision=event['source_revision'],
                              source_files=event['source_files'])
        elif phase == 'started':
            self.state['active_check'] = event
            save(self.receipt, self.state)
            self.ensure_space()
        elif phase == 'completed':
            self.state['checks'].append(event['check'])
            self.state['active_check'] = None
        save(self.receipt, self.state)

    def ensure_space(self):
        available = shutil.disk_usage(gate.REPO).free
        if available >= self.reserve:
            return
        if not self.maintain_cache:
            raise RuntimeError('Insufficient disk reserve; no cache changed. Coordinate build stages.')
        # The coordinator must serialize builds before selecting this option.
        # Do not clear the shared regenerable cache if another build is active.
        processes = subprocess.check_output(['ps', '-Ao', 'comm='], text=True)
        if any(Path(line.strip()).name in ('go', 'compile', 'link') for line in processes.splitlines()):
            raise RuntimeError('Another Go build is active; cache maintenance deferred.')
        p = subprocess.run(['go', 'clean', '-cache'], cwd=gate.REPO, capture_output=True,
                           text=True, timeout=120)
        remaining = shutil.disk_usage(gate.REPO).free
        self.state['maintenance'].append(dict(command=['go', 'clean', '-cache'],
                                             exit=p.returncode, stdout=p.stdout,
                                             stderr=p.stderr, before_bytes=available,
                                             after_bytes=remaining))
        save(self.receipt, self.state)
        if p.returncode or remaining < self.reserve:
            raise RuntimeError('Disk reserve remains insufficient after build-cache maintenance.')


def evaluate(revision, receipt, only=None, maintain_cache=False, reserve_mib=768):
    # Bind immutable source before touching an existing receipt.
    revision = subprocess.check_output(['git', 'rev-parse', revision + '^{commit}'],
                                       cwd=gate.REPO, text=True).strip()
    observer = Progress(receipt, maintain_cache, reserve_mib)
    files = [Path(__file__).resolve(), Path(gate.__file__).resolve(),
             Path(__file__).with_name('ch10_retained_adapters.py')]
    before = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in files}
    observer.state.update(source_revision=revision, orchestration_files=before,
                          selected_checks=only, full_retained_run=False)
    save(receipt, observer.state)
    try:
        result = gate.evaluate(revision, only, progress=observer,
                               prepare_adapters=prepare,
                               prepared_checkers=('ch09-review-record-bounds.py',
                                                  'ch09-review-boundaries.py',
                                                  'accept_ch09_management.py'))
        if before != {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in files}:
            raise RuntimeError('Orchestration changed during retained checks')
        observer.state.update(complete_run=True, passed=result['passed'],
                              full_retained_run=result['complete_run'],
                              preceding_gate=result, active_check=None)
    except Exception as error:
        observer.state.update(passed=False, complete_run=False, error=str(error),
                              traceback=traceback.format_exc())
    save(receipt, observer.state)
    return observer.state


if __name__ == '__main__':
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('source_commit')
    p.add_argument('--receipt', required=True, type=Path)
    p.add_argument('--only', action='append')
    p.add_argument('--maintain-build-cache', action='store_true',
                   help='Only during a coordinator-serialized run; remove regenerable Go build cache between commands if low on space')
    p.add_argument('--reserve-mib', type=int, default=768)
    a = p.parse_args()
    if a.reserve_mib < 1:
        p.error('--reserve-mib must be positive')
    result = evaluate(a.source_commit, a.receipt, a.only, a.maintain_build_cache, a.reserve_mib)
    print(json.dumps(dict(passed=result['passed'], complete_run=result['complete_run'],
                         completed_checks=len(result['checks']), receipt=str(a.receipt),
                         active_check=result.get('active_check'), error=result.get('error')),
                     indent=2))
    raise SystemExit(0 if result['passed'] else 1)
