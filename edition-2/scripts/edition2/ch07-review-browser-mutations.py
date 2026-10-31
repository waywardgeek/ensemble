#!/usr/bin/env python3
"""Comparative-review browser lifecycle/speech controls, with actual Chrome positives."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command

HERE = Path(__file__).resolve().parent
MUTANTS = {
    'root-close-admits-peer': ('ch07-review-speech-ownership.cjs', [('application.js', 'this.service.close();\n    for (const page of this.pages) page.close();', 'for (const page of this.pages) page.close();\n    this.service.close();')], {'whole-application-close'}),
    'retained-dom-listeners': ('ch07-review-browser-lifetime.cjs', [('page.js', 'for (const remove of this.listeners) remove();', '')], {'lifetime'}),
    'cancel-peer-native-speech': ('ch07-review-speech-ownership.cjs', [('speech-service.js', 'if (this.active?.page !== page) return;', '')], {'idle-page-reset', 'queued-page-cancel', 'fifo-ready-heads'}),
    'parallel-native-admission': ('ch07-review-speech-ownership.cjs', [('speech-service.js', 'this.closed || this.active || !this.pending.length', 'this.closed || !this.pending.length')], {'queued-page-cancel', 'active-page-cancel', 'fifo-ready-heads', 'closed-owner-and-stale-callbacks', 'whole-application-close'}),
    'reverse-ready-fifo': ('ch07-review-speech-ownership.cjs', [('speech-service.js', 'this.pending.shift()', 'this.pending.pop()')], {'fifo-ready-heads'}),
    'revive-closed-queue': ('ch07-review-speech-ownership.cjs', [('speech.js', 'if (this.closed || !text.trim()) return;', 'if (!text.trim()) return;')], {'closed-owner-and-stale-callbacks'}),
}


def run(root, script):
    result = command(['node', HERE/script, root], HERE.parents[1], timeout=90)
    try:
        observed = json.loads(result['stdout'])
    except ValueError:
        return result, None, set()
    failed = {r['id'] for r in observed['checks'] if not r['passed']} if 'checks' in observed else ({'lifetime'} if not observed['passed'] else set())
    return result, observed, failed


def evaluate(source):
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch07-review-browser-mutations-') as directory:
        root = Path(directory)/'source'
        shutil.copytree(source/'gui/web/gui', root/'gui/web/gui')
        assets = root/'gui/web/gui'
        hashes = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(assets.iterdir()) if p.is_file()}
        positives = {}
        for script in sorted({m[0] for m in MUTANTS.values()}):
            result, observed, failed = run(root, script)
            positives[script] = result
            if result['exit'] or not observed or not observed['passed'] or failed:
                return dict(passed=False, error='positive control failed', positives=positives, source_files=hashes)
        for name, (script, patches, expected) in MUTANTS.items():
            originals = {}
            try:
                for filename, old, new in patches:
                    p = assets/filename
                    text = p.read_text()
                    if text.count(old) != 1:
                        raise ValueError('mutation anchor absent or ambiguous: '+name)
                    originals[p] = text
                    p.write_text(text.replace(old, new))
                result, observed, failed = run(root, script)
            finally:
                for p, text in originals.items():
                    p.write_text(text)
            rows.append(dict(id=name, passed=result['exit'] != 0 and observed is not None and failed == expected, patches=patches, expected_failures=sorted(expected), actual_failures=sorted(failed), negative=result))
    return dict(scope=__doc__, source=str(source), source_files=hashes, positives=positives,
                checker_files={p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__), HERE/'ch07-review-browser-lifetime.cjs', HERE/'ch07-review-speech-ownership.cjs']}, passed=all(r['passed'] for r in rows), checks=rows)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    args = parser.parse_args()
    result = evaluate(args.source.resolve(strict=True))
    print(json.dumps(result, indent=2))
    raise SystemExit(0 if result['passed'] else 1)
