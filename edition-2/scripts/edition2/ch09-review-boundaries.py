#!/usr/bin/env python3
"""Independent audit of reviewed student boundary tests plus targeted deletions.

Uses small Go overlays, never edits production/tests. Student-authored scenarios
remain attributed; supplemental assertions and mutations are reviewer-owned.
Owner-state exhaustion is not a fabricated full log reaching uint64 exhaustion.
"""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile

HERE=Path(__file__).resolve().parent

def digest(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def files(root):
    return {str(p.relative_to(root)):digest(p) for p in sorted(root.rglob('*')) if p.is_file() and '.git' not in p.parts and (p.suffix in ('.go','.mjs','.py') or p.name in ('go.mod','go.sum'))}
def replace(text,old,new):
    assert text.count(old)==1,repr(old)
    return text.replace(old,new)
def audit(root):
    before=files(root);rows=[]
    with tempfile.TemporaryDirectory(prefix='ch09-boundary-overlays-') as tmp:
        tmp=Path(tmp)
        def run(label,pattern,package='.',changes=None,expected=None,cwd=root):
            overlay={}
            for index,(name,text) in enumerate((changes or {}).items()):
                p=tmp/f'{label}-{index}.go';p.write_text(text);overlay[str(root/name)]=str(p)
            args=['go','test','-race']
            if overlay:
                p=tmp/(label+'.json');p.write_text(json.dumps({'Replace':overlay}));args+=['-overlay',str(p)]
            args += [package,'-count=1','-timeout=45s','-run',pattern,'-v']
            proc=subprocess.run(args,cwd=cwd,text=True,capture_output=True,timeout=90)
            output=proc.stdout+proc.stderr
            passed=proc.returncode==0 if expected is None else proc.returncode!=0 and expected in output and '[build failed]' not in output and 'DATA RACE' not in output
            rows.append(dict(id=label,passed=passed,expected_failure=expected,command=args,exit=proc.returncode,stdout=proc.stdout,stderr=proc.stderr,overlay_sha256={name:hashlib.sha256(text.encode()).hexdigest() for name,text in (changes or {}).items()}))
            assert passed,label+'\n'+output
        # Retain the student's existing scenarios, adding diagnostic-safe nil
        # handling so a rollback deletion fails an assertion, not a panic.
        fault='skills_persistence_test.go';faulttext=(root/fault).read_text()
        faulttext=replace(faulttext,'} else if after.State.Revision != 1','} else if after.State == nil || after.State.Revision != 1')
        # The public pause scenario must also prove the pending model call gets
        # its matched error pair, while the earlier typed commit remains intact.
        pause='skills_integration_test.go';pausetext=(root/pause).read_text()
        anchor='\tif len(a.Snapshot().Jobs) != 0 {\n\t\tt.Fatal("management allocated a job")\n\t}'
        pausetext=replace(pausetext,anchor,anchor+'''
    called, returned := 0, 0
    for _, event := range a.Events() {
        if event.Tool == nil || event.Tool.CallID != "load" { continue }
        if event.Type == "tool_called" { called++ }
        if event.Type == "tool_returned" {
            returned++
            if !event.Tool.IsError { t.Fatal("interrupted management result not error") }
        }
    }
    if called != 1 || returned != 1 { t.Fatalf("interrupted management missing matched pair: %d/%d", called, returned) }
''')
        # Reuse the supplied complete-state comparisons; positive counter control
        # fits its entire two-record group exactly at MaxUint64.
        counter='^TestCounterExhaustionAndOwnedCopies$'
        run('counter-positive',counter,'./internal/skills')
        run('commit-positive','^TestSkill(DurableCommitFailureBoundaries|InitialWriteFailureDoesNotApply)$',changes={fault:faulttext})
        run('pause-pair-positive','^TestSkillPauseHoldsManagementButTypedControlsProceed$',changes={pause:pausetext})
        run('window-positive','^TestSkillWatchCurrentStateOutlivesMaterialWindow$')
        run('precision-positive','^TestSkillProjectionKeepsUint64MaterialIdentities$',cwd=root/'gui')
        skills='internal/skills/skills.go';original=(root/skills).read_text()
        run('revision-guard-deletion',counter,'./internal/skills',{skills:replace(original,'if revision == math.MaxUint64 {','if false && revision == math.MaxUint64 {')},'wanted skill_too_large, got <nil>')
        run('whole-group-guard-deletion',counter,'./internal/skills',{skills:replace(original,'if uint64(newCount) > math.MaxUint64-lastID {','if false && uint64(newCount) > math.MaxUint64-lastID {')},'wanted skill_too_large, got <nil>')
        ensemble='ensemble.go';original=(root/ensemble).read_text()
        apply='\tif skillCandidate != nil {\n\t\ta.skills.Apply(skillCandidate, owned.Seq)\n\t}\n'
        early=replace(original,apply,'');early=replace(early,'\tif err == nil && persist {\n\t\terr = a.log.Append(owned)',apply+'\tif err == nil && persist {\n\t\terr = a.log.Append(owned)')
        run('apply-before-append','^TestSkillDurableCommitFailureBoundaries$/skills_changed$',changes={fault:faulttext,ensemble:early},expected='failed transition changed authority')
        rollback=replace(original,'\t\terr = a.log.Append(owned)','\t\terr = a.log.Append(owned)\n\t\tif err != nil && owned.Type == "tool_returned" { a.skills = nil }')
        run('rollback-after-result-failure','^TestSkillDurableCommitFailureBoundaries$/tool_returned$',changes={fault:faulttext,ensemble:rollback},expected='durable transition was rolled back after result failure')
        projection='gui/projection.go';original=(root/projection).read_text()
        run('float-projection-deletion','^TestSkillProjectionKeepsUint64MaterialIdentities$',changes={projection:replace(original,'\tdecoder.UseNumber()\n','')},cwd=root/'gui',expected='lost "revision":18446744073709551615')
        actor='internal/llm/actor.go';original=(root/actor).read_text()
        anchor='\t\tif a.record(common.Event{Type: "tool_returned", Tool: &common.ToolEvent{CallID: p.CallID, IsError: true, Parts: []common.Part{Text(note + "turn interrupted before execution")}}}) != nil {'
        run('interrupted-management-error-deletion','^TestSkillPauseHoldsManagementButTypedControlsProceed$',changes={pause:pausetext,actor:replace(original,anchor,anchor.replace('IsError: true','IsError: false'))},expected='interrupted management result not error')
        # Audit, execute and bind rather than duplicate three legitimate existing
        # deletion scenarios. They check concrete forbidden bytes, lost dedicated
        # manual and partial diamond publication, with a positive before each.
        command=['python3',str(root/'evidence/ch09/deletion-controls.py'),'--source',str(root)]
        proc=subprocess.run(command,text=True,capture_output=True,timeout=180)
        value=json.loads(proc.stdout) if proc.returncode==0 else None
        passed=bool(value and value.get('passed') and len(value['checks'])==6 and all(c['passed'] for c in value['checks']))
        rows.append(dict(id='reviewed-student-three-deletions',passed=passed,command=command,exit=proc.returncode,stdout=proc.stdout,stderr=proc.stderr))
        assert passed,proc.stdout+proc.stderr
    assert before==files(root),'source changed during boundary audit'
    return dict(passed=all(r['passed'] for r in rows),checks=rows,source=str(root),source_files=before,checker_sha256=digest(Path(__file__)),scope=__doc__,limits='No immutable freeze, full legacy gate, live provider acceptance or historical comparison. Student scenarios are reused after review, not relabeled reviewer-authored.')

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);p.add_argument('--receipt',type=Path);a=p.parse_args();r=audit(a.source.resolve(strict=True))
    if a.receipt:a.receipt.write_text(json.dumps(r,indent=2)+'\n')
    print(json.dumps(r,indent=2))
