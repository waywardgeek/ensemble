#!/usr/bin/env python3
"""Published Chapter 9 raw record bound: physical reader and actual offline CLI.

Read-side subset only. Write-side atomicity and escaped-material positives are
separate. No real provider; temporary files are generated one at a time.
"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import tempfile
from accept_ch09 import MODELS, environment, system_text

CAP=67108864
HERE=Path(__file__).resolve().parent

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def audit(root,binary=None,small=False):
    fixture=HERE/'ch09-raw-record_test.go.txt'
    sources={str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*.go')) if 'evidence' not in p.parts}
    checker={str(p):sha(p) for p in [Path(__file__),fixture,HERE/'accept_ch09.py']}
    if binary:checker[str(binary)]=sha(binary)
    rows=[]
    with tempfile.TemporaryDirectory(prefix='ch09-raw-bound-') as tmp:
        tmp=Path(tmp);overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/eventlog/zz_ch09_independent_raw_test.go'):str(fixture)}}))
        pattern='^TestCh09RawSkillReaderBounds$'+('/small-parent$' if small else '')
        command=['go','test','-overlay',str(overlay),'./internal/eventlog','-run',pattern,'-count=1','-timeout=120s','-v']
        proc=subprocess.run(command,cwd=root,text=True,capture_output=True,timeout=150)
        ran='=== RUN   TestCh09RawSkillReaderBounds/' in proc.stdout
        rows.append(dict(id='physical-reader',passed=proc.returncode==0 and ran,command=command,exit=proc.returncode,stdout=proc.stdout,stderr=proc.stderr))
        if binary:
            # Reuse the exact small physical-reader record for a genuine public
            # semantic positive before giving its larger variants any credit.
            record=re.search(r'record := `([^`]+)`',fixture.read_text()).group(1).encode()
            assert json.loads(record)['skills']['activated'][0]['sha256']==hashlib.sha256(b'Identity.').hexdigest()
            log=tmp/'input.jsonl';env=environment(tmp,'anthropic','http://127.0.0.1:1')
            cases=[('small-parent',len(record)+1,True,True)]
            if not small:cases += [('exact-LF',CAP,True,True),('exact-no-LF',CAP,False,True),('over-LF',CAP+1,True,False),('over-no-LF',CAP+1,False,False)]
            valid=False
            for name,size,lf,want in cases:
                with log.open('wb') as out:
                    out.write(b'{"log_version":1}\n'+record);remaining=size-len(record)-int(lf);block=b' '*65536
                    while remaining:amount=min(remaining,len(block));out.write(block[:amount]);remaining-=amount
                    if lf:out.write(b'\n')
                before=sha(log)
                command=[str(binary),'render',str(log)];proc=subprocess.run(command,cwd=tmp,env=env,text=True,capture_output=True,timeout=45)
                try:body=json.loads(proc.stdout);accepted=proc.returncode==0 and system_text('anthropic',body)=='Identity.'
                except ValueError:accepted=False
                if name=='small-parent':valid=accepted
                refused=proc.returncode!=0 and not proc.stdout.strip() and 'line 2' in proc.stderr
                rows.append(dict(id='public/'+name,passed=valid and (accepted if want else refused) and sha(log)==before,positive_parent=valid,raw_record_bytes=size,framing_lf=lf,exit=proc.returncode,stdout=proc.stdout[:3000],stderr=proc.stderr[:3000],unchanged=sha(log)==before))
    unchanged=sources=={str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*.go')) if 'evidence' not in p.parts}
    assert unchanged,'source changed during record checks'
    assert checker=={p:sha(Path(p)) for p in checker},'checker or binary changed'
    return dict(passed=all(r['passed'] for r in rows),checks=rows,source=str(root),source_files=sources,checker_files=checker,scope=__doc__,small_only=small,limits='Physical generator refuses reads beyond cap plus a fixed64KiB lookahead allowance; no unbounded allocation fixture. This is not write-side admission acceptance or full memory profiling.')

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);p.add_argument('--cli',type=Path);p.add_argument('--small-only',action='store_true');p.add_argument('--receipt',type=Path);a=p.parse_args();r=audit(a.source.resolve(strict=True),a.cli.resolve(strict=True) if a.cli else None,a.small_only)
    if a.receipt:a.receipt.write_text(json.dumps(r,indent=2)+'\n')
    print(json.dumps(r,indent=2));raise SystemExit(not r['passed'])
