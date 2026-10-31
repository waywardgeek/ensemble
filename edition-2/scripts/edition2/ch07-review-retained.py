#!/usr/bin/env python3
"""Reconcile the Chapter 7 immutable-value checker correction without rerunning unrelated groups."""
import argparse
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile
from accept_ch05 import command

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
PREFIX = 'solutions/edition-2/main/'


def extract(revision, root):
    revision = subprocess.check_output(['git', 'rev-parse', revision+'^{commit}'], cwd=REPO, text=True).strip()
    blob = subprocess.check_output(['git', 'archive', revision, PREFIX], cwd=REPO)
    hashes = {}
    with tarfile.open(fileobj=io.BytesIO(blob)) as archive:
        for member in archive:
            if not member.isfile():
                continue
            relative = Path(member.name).relative_to(PREFIX)
            if '..' in relative.parts:
                raise ValueError('unsafe archive path')
            if any(p in ('evidence', '.git', 'node_modules', '__pycache__') for p in relative.parts):
                continue
            data = archive.extractfile(member).read()
            target = root/relative
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
            hashes[str(relative)] = hashlib.sha256(data).hexdigest()
    return revision, hashes


def evaluate(prior_path):
    prior = json.loads(prior_path.read_text())
    assert not prior['passed']
    failed = [row['id'] for row in prior['checks'] if not row['passed']]
    assert failed == ['retained-ch05-assertions'], failed
    assert len(prior['checks']) == 33, 'unexpected original gate group count'
    checks = []
    def run(name, args, expected=0):
        result = command(args, REPO, timeout=1200)
        checks.append(dict(id=name, passed=result['exit']==expected, command=result))
        return result
    with tempfile.TemporaryDirectory(prefix='ch07-retained-review-') as temporary:
        current = Path(temporary)/'current'
        revision, hashes = extract(prior['source_revision'], current)
        assert hashes == prior['source_files'], 'source differs from original gate'
        baseline = run('historical-source-checker-baseline', ['go', 'run', HERE/'ch05sourcecheck/main.go', current], 1)
        findings = json.loads(baseline['stdout'])
        assert {(f['file'], f['why']) for f in findings} == {('cli/chat.go', 'package-owned runtime mutable value: ErrQuit'), ('gui/server.go', 'package-owned mutable zero value: assets')}, findings
        run('new-source-checker-controls', ['go', 'test', './scripts/edition2/ch07sourcecheck', '-count=1', '-v'])
        run('new-source-checker-vet', ['go', 'vet', './scripts/edition2/ch07sourcecheck'])
        formatted = run('new-source-checker-format', ['gofmt', '-l', HERE/'ch07sourcecheck'])
        checks[-1]['passed'] &= not formatted['stdout'].strip()
        repaired = run('retained-ch05-assertions', ['python3', HERE/'accept_ch07_prior.py', current])
        repaired_result = json.loads(repaired['stdout'])
        assert repaired_result['score'] == 100 and repaired_result['passed'], repaired_result
        legacy = Path(temporary)/'legacy'
        legacy_revision, legacy_hashes = extract('75bd14d5d2424778ebf45cb9025e9dbf314f956a', legacy)
        run('legacy-source-original-checker', ['go', 'run', HERE/'ch05sourcecheck/main.go', legacy])
        run('legacy-source-new-checker', ['go', 'run', HERE/'ch07sourcecheck/main.go', legacy])
    checker_paths = [HERE/'ch07-review-retained.py', HERE/'accept_ch07_prior.py', HERE/'ch07sourcecheck/main.go', HERE/'ch07sourcecheck/main_test.go', HERE/'accept_ch06_prior.py', HERE/'accept_ch05.py', HERE/'ch05sourcecheck/main.go']
    return dict(scope=__doc__, source_revision=revision, source_files=hashes,
                prior_receipt=str(prior_path), prior_receipt_sha256=hashlib.sha256(prior_path.read_bytes()).hexdigest(),
                original_passed_groups=32, original_failed_groups=1, corrected_group='retained-ch05-assertions',
                accepted_groups=33, live_and_comparative_acceptance=False,
                legacy_source_revision=legacy_revision, legacy_source_files=legacy_hashes,
                checker_files={str(p.relative_to(HERE)):hashlib.sha256(p.read_bytes()).hexdigest() for p in checker_paths},
                passed=all(c['passed'] for c in checks), checks=checks)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('original_gate_receipt', type=Path)
    args = parser.parse_args()
    result = evaluate(args.original_gate_receipt)
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
