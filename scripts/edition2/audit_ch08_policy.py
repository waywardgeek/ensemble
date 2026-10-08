#!/usr/bin/env python3
"""Deletion controls for independently protected Chapter 8 policy/disk behavior.

Every mutant starts from the same passing disposable source copy. Require the
exact intended leaf-test failure set; compile errors or generic failures do not
count. No historical grader or student source is edited.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile
from accept_ch08_persistence import instrument, change


def execute(root, module, pattern):
    command=['go','test','-race','-json','-count=1','-timeout=45s','-run',pattern,'.']
    result=subprocess.run(command,cwd=root/module,text=True,capture_output=True,timeout=120)
    events=[]
    for line in result.stdout.splitlines():
        try: events.append(json.loads(line))
        except json.JSONDecodeError: pass
    failed={e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')}
    leaves=sorted(name for name in failed if not any(other.startswith(name+'/') for other in failed))
    return dict(command=command,module=module,exit=result.returncode,failed_leaves=leaves,
                output=''.join(e.get('Output','') for e in events),stderr=result.stderr)


def evaluate(source):
    here=Path(__file__).parent
    with tempfile.TemporaryDirectory(prefix='ch08-policy-deletions-') as temporary:
        base=Path(temporary)/'positive'
        shutil.copytree(source,base,ignore=shutil.ignore_patterns('.git'))
        hashes={str(p.relative_to(base)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(base.rglob('*')) if p.is_file() and (p.suffix=='.go' or p.name in ('go.mod','go.sum'))}
        instrument(base)
        (base/'zz_ch08_public_test.go').write_bytes((here/'ch08-policy-public_test.go').read_bytes())
        controls=[execute(base,'','^TestCh08Public'),execute(base,'gui','^TestCh08Persistence')]
        if any(row['exit'] for row in controls):
            return dict(passed=False,scope=__doc__,positive_controls=controls,mutants=[],source_files=hashes)
        specs=[]
        def add(name,file,old,new,module,pattern,leaves,reason):specs.append((name,file,old,new,module,pattern,leaves,reason))
        add('delete-policy-consumer','internal/llm/actor.go',
            'if r.rounds >= r.policy.EffectiveMaxModelRequests {',
            'if r.rounds >= common.DefaultMaxModelRequests {','','^TestCh08PublicTwoAgentEffectsAndReplay$',
            ['TestCh08PublicTwoAgentEffectsAndReplay'],'policy did not control real HTTP')
        for domain,file in [('policy','internal/policy/policy.go'),('preferences','gui/internal/preferences/preferences.go')]:
            for name,old,new,stage in [
                ('short-write','if err == nil && n != len(data) {','if false && err == nil && n != len(data) {','short-write'),
                ('sync','err = ch08Sync(s, f)','_ = ch08Sync(s, f)','sync'),
                ('close','err = closeErr','_ = closeErr','close')]:
                leaf=f'TestCh08PersistenceStages/{domain}/{stage}/fault'
                add(f'{domain}-ignore-{name}',file,old,new,'gui',f'^TestCh08PersistenceStages$/{domain}/{stage}',[leaf],'wanted settings_persist_failed')
            stages=['create','write','short-write','sync','close','replace']
            leaves=[f'TestCh08PersistenceStages/{domain}/{stage}/fault' for stage in stages]
            if domain=='policy':
                old='if err == nil {\n\t\ts.value = next\n\t}'
                new='if true {\n\t\ts.value = next\n\t}'
                reason='pre-replace failure changed applied state or original bytes'
            else:
                old='if err != nil {\n\t\ts.parent.Ensemble().Logf("preferences persistence failed before replacement")'
                new='if false {\n\t\ts.parent.Ensemble().Logf("preferences persistence failed before replacement")'
                reason='wanted settings_persist_failed'
            add(domain+'-apply-failed-candidate',file,old,new,'gui','^TestCh08PersistenceStages$/'+domain,leaves,reason)
            add(domain+'-remove-busy',file,'if s.busy {','if false && s.busy {','gui',
                '^TestCh08PersistenceHeldControlsAndRetry$/'+domain,
                ['TestCh08PersistenceHeldControlsAndRetry/'+domain],'settings operation did not settle')
        # Earlier preparation had two Wait paths; the current idempotent Close
        # waits on closeDone in its second path. Remove the actual writer joins
        # in this scoped lifetime mutant, preserving ordinary successful writes.
        closefile='gui/internal/preferences/preferences.go'
        text=(base/closefile).read_text()
        prefix,body=text.split('func (s *Service) Close() {',1)
        assert body.count('s.writer.Wait()') in (1,2)
        add('preferences-remove-close-join',closefile,'func (s *Service) Close() {'+body,
            'func (s *Service) Close() {'+body.replace('s.writer.Wait()','/* deleted owned writer join */'),
            'gui','^TestCh08PersistenceCloseJoinsActualOutcome$/preferences',
            ['TestCh08PersistenceCloseJoinsActualOutcome/preferences/'+mode for mode in ['replace-success','replace-failure','committed']],
            'close returned before owned writer joined')
        add('preferences-remove-watch-overflow',closefile,'if len(w.values) >= ensemble.WatchItems {',
            'if false && len(w.values) >= ensemble.WatchItems {','gui',
            '^TestCh08PersistencePreferenceOverflowAndIdleClose$',
            ['TestCh08PersistencePreferenceOverflowAndIdleClose'],'overflowing preference watch silently remained current')
        add('preferences-delete-no-change',closefile,'if candidate == s.value.Preferences {',
            'if false && candidate == s.value.Preferences {','gui','^TestCh08PersistencePreferenceHandoff$',
            ['TestCh08PersistencePreferenceHandoff'],'no-change preference update advanced revision')
        rows=[]
        for name,file,old,new,module,pattern,leaves,reason in specs:
            mutant=Path(temporary)/name
            shutil.copytree(base,mutant)
            try:
                target=mutant/file;target.write_text(change(target.read_text(),old,new))
                result=execute(mutant,module,pattern)
                passed=result['exit']!=0 and result['failed_leaves']==sorted(leaves) and reason in result['output'] and 'build failed' not in result['output']
                rows.append(dict(id=name,passed=passed,expected_leaves=sorted(leaves),intended_reason=reason,result=result))
            finally:
                # Retain the receipt, not fourteen duplicate audio/evidence
                # trees. The next mutant still starts from the same positive.
                shutil.rmtree(mutant)
        return dict(passed=all(r['passed'] for r in rows),scope=__doc__,positive_controls=controls,mutants=rows,source_files=hashes,
                    checker_files={name:hashlib.sha256((here/name).read_bytes()).hexdigest() for name in ['audit_ch08_policy.py','accept_ch08_persistence.py','ch08-disk-control.go.tmpl','ch08-persistence_test.go','ch08-policy-public_test.go']})


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('source_directory',type=Path)
    args=parser.parse_args();result=evaluate(args.source_directory.resolve(strict=True))
    print(json.dumps(result,indent=2));raise SystemExit(0 if result['passed'] else 1)
