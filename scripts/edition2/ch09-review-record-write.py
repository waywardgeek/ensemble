#!/usr/bin/env python3
"""Source-bound skill write atomicity and encoding-bound mutations; no paid calls.

Reuses reviewed student valid-catalog/actor scenarios, adds exact writer boundary
and original-byte import controls. Overlays change only disposable compiler input.
"""
import argparse
import hashlib
import json
import re
from pathlib import Path
import subprocess
import tempfile

HERE=Path(__file__).resolve().parent

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def replace(text,old,new):
    assert text.count(old)==1,old
    return text.replace(old,new)
def audit(root):
    source={str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*.go')) if 'evidence' not in p.parts}
    fixtures={name:(HERE/name).read_text() for name in ['ch09-write-record_test.go.txt','ch09-import-encoding_test.go.txt']}
    rows=[]
    with tempfile.TemporaryDirectory(prefix='ch09-record-write-') as temporary:
        temporary=Path(temporary)
        def run(label,pattern,changes=None,expected=None):
            overlay={}
            for index,(name,text) in enumerate((changes or {}).items()):
                p=temporary/(label+str(index)+'.go');p.write_text(text);overlay[str(root/name)]=str(p)
            for name,text in fixtures.items():
                p=temporary/name;p.write_text(text);overlay[str(root/('zz_independent_'+name.removesuffix('.txt')))]=str(p)
            p=temporary/(label+'.json');p.write_text(json.dumps({'Replace':overlay}))
            args=['go','test','-overlay',str(p),'.','-run',pattern,'-count=1','-timeout=120s','-v']
            proc=subprocess.run(args,cwd=root,text=True,capture_output=True,timeout=150)
            output=proc.stdout+proc.stderr
            passed=proc.returncode==0 and '=== RUN ' in output if expected is None else proc.returncode!=0 and expected in output and '[build failed]' not in output
            rows.append(dict(id=label,passed=passed,expected_failure=expected,exit=proc.returncode,stdout=proc.stdout,stderr=proc.stderr,overlay_sha256={name:hashlib.sha256(text.encode()).hexdigest() for name,text in (changes or {}).items()}))
            return passed
        exact=run('exact-write-positive','^TestCh09IndependentWriteRecordExactBoundary$')
        atomic=run('typed-model-initial-positive','^TestSkill(CompleteRecordRefusalIsAtomicAndRecoverable|OversizeInitializationRefuses)$')
        run('original-encoding-import-positive','^TestCh09IndependentImportCountsOriginalEncoding$')
        path='internal/eventlog/log.go';original=(root/path).read_text()
        if exact:
            run('missing-framing-byte','^TestCh09IndependentWriteRecordExactBoundary$',{path:replace(original,'if len(data)+1 > common.SkillRecordLimit {','if len(data) > common.SkillRecordLimit {')},'one-over encoded writer limit not refused')
            run('reject-exact-boundary','^TestCh09IndependentWriteRecordExactBoundary$',{path:replace(original,'if len(data)+1 > common.SkillRecordLimit {','if len(data)+1 >= common.SkillRecordLimit {')},'exact encoded writer limit refused')
        if atomic:
            path='ensemble.go';original=(root/path).read_text()
            # Delete only the early controlled refusal. The independent log guard
            # still stops bytes, distinguishing accidental terminal faulting.
            matches=re.findall(r'(?m)^([\t]+)if err := eventlog.CheckSkillRecord\(a, event\); err != nil \{\n\1\ta.mu.Unlock\(\)\n\1\treturn err\n\1}\n',original)
            assert len(matches)==1,'unique controlled precheck anchor changed'
            indent=matches[0];old=indent+'if err := eventlog.CheckSkillRecord(a, event); err != nil {\n'+indent+'\ta.mu.Unlock()\n'+indent+'\treturn err\n'+indent+'}\n'
            run('remove-controlled-precheck','^TestSkillCompleteRecordRefusalIsAtomicAndRecoverable$',{path:replace(original,old,'')},'oversize changed state, log or storage health')
    unchanged=source=={str(p.relative_to(root)):sha(p) for p in sorted(root.rglob('*.go')) if 'evidence' not in p.parts}
    return dict(passed=unchanged and len(rows)==6 and all(r['passed'] for r in rows),source_unchanged=unchanged,checks=rows,source=str(root),source_files=source,checker_files={str(p):sha(p) for p in [Path(__file__)]+[HERE/name for name in fixtures]},scope=__doc__,limits='Exact physical writer boundary is separate from full candidate mutation; reviewed student actor scenarios prove typed/model/constructor refusal and continued healthy storage. No full gate/live/historical claim.')

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);p.add_argument('--receipt',type=Path);a=p.parse_args();r=audit(a.source.resolve(strict=True))
    if a.receipt:a.receipt.write_text(json.dumps(r,indent=2)+'\n')
    print(json.dumps(r,indent=2));raise SystemExit(not r['passed'])
