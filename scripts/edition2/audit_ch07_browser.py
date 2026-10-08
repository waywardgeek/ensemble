#!/usr/bin/env python3
"""Real-Chrome distinguishing deletion controls for reusable browser components.

Uses only local static assets, controlled socket frames and controlled speech.
This does not prove provider behavior, actual speech synthesis or audible output.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command

HERE=Path(__file__).resolve().parent
MUTANTS={
 'unsafe-tool-html':('artifacts.js','pre.textContent = shown;','pre.innerHTML = shown;',{'hostile-content-expansion-keyboard-focus-scroll'}),
 'stale-speech-settlement':('speech.js','if (settled || generation !== this.generation || this.current !== item) return;','if (settled) return;',{'speech-cancel-stale-callback-and-final-dedup'}),
 'speak-replayed-prefix':('speech.js','{text, sent: text.length, tool: false}','{text, sent: 0, tool: false}',{'speech-cancel-stale-callback-and-final-dedup'}),
 'hide-acceptance-uncertainty':('page.js',"if (submission.status === 'rejected') throw submission.reason;","if (submission.status === 'rejected') throw (pause.status === 'rejected' ? pause.reason : submission.reason);",{'page-input-submit-and-lost-acceptance'}),
 'merge-incomplete-snapshot':('connector.js','this.staging = {...m, events: [], partials: []}; return;','this.staging = {...m, events: [], partials: []}; this.owner.snapshot(this.staging); return;',{'connector-atomic-reset-and-uncertain-prompt'}),
 'local-part-identity-only':('artifacts.js',"[p.agent_id, p.request_id, p.operation_id, p.part_id].join('/')","[p.part_id].join('/')",{'cards-identity-partial-final-window','speech-cancel-stale-callback-and-final-dedup'}),
 'duplicate-provisional-final':('artifacts.js','const card = this.cards.get(provisional);','const card = null;',{'cards-identity-partial-final-window'}),
}


def run(root):
    result=command(['node',HERE/'accept_ch07_browser.cjs',root],HERE.parents[1],timeout=120)
    try:receipt=json.loads(result['stdout'])
    except ValueError:return result,None
    return result,receipt


def evaluate(source,only=None):
    with tempfile.TemporaryDirectory(prefix='ch07-browser-deletions-') as directory:
        root=Path(directory)/'source';shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence','node_modules','__pycache__'))
        assets=root/'gui/web/gui';hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(assets.rglob('*')) if p.is_file()}
        positive,receipt=run(root);rows=[]
        if positive['exit'] or not receipt or not receipt.get('passed'):
            return dict(passed=False,error='browser positive control failed',positive=positive,source_files=hashes)
        for name in only or MUTANTS:
            file,old,new,expected=MUTANTS[name];target=assets/file;text=target.read_text()
            if text.count(old)!=1:rows.append(dict(id=name,passed=False,error='mutation anchor absent or ambiguous'));continue
            target.write_text(text.replace(old,new))
            try:negative,observed=run(root)
            finally:target.write_text(text)
            failed={r['id'] for r in observed['checks'] if not r['passed']} if observed else set()
            passed=negative['exit']!=0 and observed is not None and failed==expected
            rows.append(dict(id=name,passed=passed,old=old,new=new,file=file,expected_failures=sorted(expected),actual_failures=sorted(failed),negative=negative))
        return dict(scope=__doc__,source=str(source),source_files=hashes,positive=positive,
                    checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__),HERE/'accept_ch07_browser.cjs']},
                    passed=all(r['passed'] for r in rows),checks=rows)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);p.add_argument('--only',action='append',choices=MUTANTS);a=p.parse_args();r=evaluate(a.source.resolve(strict=True),a.only);print(json.dumps(r,indent=2));raise SystemExit(0 if r['passed'] else 1)
