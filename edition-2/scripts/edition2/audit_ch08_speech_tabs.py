#!/usr/bin/env python3
"""Delete cross-tab native lease protections after a passing actual-lock control."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
HERE=Path(__file__).resolve().parent

def run(root):
    result=subprocess.run(['node',str(HERE/'accept_ch08_speech_tabs.cjs'),str(root)],capture_output=True,text=True,timeout=90)
    try:receipt=json.loads(result.stdout)
    except ValueError:receipt={}
    return dict(exit=result.returncode,receipt=receipt,stderr=result.stderr)

def evaluate(source):
    with tempfile.TemporaryDirectory(prefix='ch08-speech-tab-audit-') as tmp:
        root=Path(tmp)/'source';shutil.copytree(source/'gui/web/gui',root/'gui/web/gui')
        target=root/'gui/web/gui/speech-service.js';text=target.read_text()
        positive=run(root);rows=[]
        if positive['exit']==0 and positive['receipt'].get('passed'):
            for name,old,new,expected,reason in [
                ('bypass-exclusive-native-lease',"this.parent.locks.request('ensemble-native-speech', {signal: request.wait.signal}, () => {",'Promise.resolve().then(() => {',{'same-context-held-lease-waiter-cancel-owner-close'},'waiter used native service'),
                ('cancel-without-native-ownership','request.wait.abort();','request.wait.abort(); if (!request.release) this.parent.synthesis.cancel();',{'same-context-held-lease-waiter-cancel-owner-close'},'canceled waiter invoked native cancel'),
                ('retain-closed-owner-lease','      request.release();','      // deleted: release lease after native cleanup',{'same-context-held-lease-waiter-cancel-owner-close'},'Timeout')]:
                assert text.count(old)==1,(name,text.count(old));target.write_text(text.replace(old,new))
                result=run(root);target.write_text(text)
                failed={r['id'] for r in result['receipt'].get('checks',[]) if not r['passed']}
                messages='\n'.join(r.get('error','') for r in result['receipt'].get('checks',[]))
                rows.append(dict(id=name,passed=result['exit']!=0 and failed==expected and reason in messages,expected_failures=sorted(expected),actual_failures=sorted(failed),intended_reason=reason,result=result))
        return dict(scope=__doc__,passed=positive['exit']==0 and len(rows)==3 and all(r['passed'] for r in rows),positive=positive,mutants=rows,checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__),HERE/'accept_ch08_speech_tabs.cjs',HERE/'ch08-browser-fixture.js']})

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);a=p.parse_args();r=evaluate(a.source.resolve(strict=True));print(json.dumps(r,indent=2));raise SystemExit(0 if r['passed'] else 1)
