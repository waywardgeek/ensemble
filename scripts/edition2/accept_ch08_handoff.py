#!/usr/bin/env python3
"""Hold the actual initial snapshot producer while real socket controls progress."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command
HERE=Path(__file__).resolve().parent


def evaluate(source):
    with tempfile.TemporaryDirectory(prefix='ch08-handoff-') as tmp:
        root=Path(tmp)/'source';shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence','node_modules','__pycache__'))
        hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in root.rglob('*') if p.is_file()}
        target=root/'gui/connector.go';text=target.read_text();anchor='\tfor _, e := range s.Events {'
        assert text.count(anchor)==1
        text=text.replace(anchor,'\tif hook, ok := c.parent.(interface{ Ch08SnapshotBoundary(context.Context) }); ok { hook.Ch08SnapshotBoundary(c.ctx) }\n'+anchor)
        target.write_text(text)
        helper=(HERE/'ch07-review-connector_test.go.txt').read_text()
        anchor='NewServer(e, a.ID(), "http://127.0.0.1:8088", nil)'
        assert helper.count(anchor)==1
        helper=helper.replace(anchor,'NewServer(e, a.ID(), "http://127.0.0.1:8088", nil, ServerOptions{PreferencesPath: filepath.Join(t.TempDir(), "preferences.json")})')
        (root/'gui/zz_ch08_socket_helpers_test.go').write_text(helper)
        (root/'gui/zz_ch08_handoff_test.go').write_bytes((HERE/'ch08-handoff_test.go').read_bytes())
        fmt=command(['gofmt','-w',target,root/'gui/zz_ch08_handoff_test.go'],root)
        positive=command(['go','test','-race','-json','-count=1','-timeout=30s','-run','^TestCh08Handoff','.'],root/'gui',90)
        rows=[]
        for name,old,new,reason in [
            ('remove-deferral','if !wait && c.snapshotPending {','if false && !wait && c.snapshotPending {','controls did not progress while snapshot held'),
            ('exclude-deferred-count','len(c.queue)+len(c.deferred) < ensemble.WatchItems','len(c.queue) < ensemble.WatchItems','full deferred queue deadlocked snapshot producer'),
            ('block-full-handoff','|| deferredBlocked {','|| false && deferredBlocked {','full deferred queue deadlocked snapshot producer')]:
            if positive['exit']:break
            assert text.count(old)==1,(name,text.count(old));target.write_text(text.replace(old,new))
            result=command(['go','test','-race','-json','-count=1','-timeout=30s','-run','^TestCh08Handoff','.'],root/'gui',90)
            target.write_text(text)
            events=[json.loads(s) for s in result['stdout'].splitlines() if s.startswith('{')]
            failed={e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e}
            leaves={x for x in failed if not any(y.startswith(x+'/') for y in failed)}
            expected={'remove-deferral':{'TestCh08HandoffControls',*[f'TestCh08HandoffCombinedCount/{n}' for n in (255,256,257)]},
                      'exclude-deferred-count':{'TestCh08HandoffCombinedCount/256','TestCh08HandoffCombinedCount/257'},
                      'block-full-handoff':{'TestCh08HandoffCombinedCount/256'}}[name]
            rows.append(dict(id=name,passed=result['exit']!=0 and reason in result['stdout'] and leaves==expected and 'DATA RACE' not in result['stdout'],intended_reason=reason,expected_leaves=sorted(expected),failed_leaves=sorted(leaves),command=result))
        return dict(scope=__doc__,source_files=hashes,positive=positive,format=fmt,mutants=rows,passed=positive['exit']==0 and len(rows)==3 and all(x['passed'] for x in rows),checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in [Path(__file__),HERE/'ch08-handoff_test.go',HERE/'ch07-review-connector_test.go.txt']})

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);a=p.parse_args();r=evaluate(a.source.resolve(strict=True));print(json.dumps(r,indent=2));raise SystemExit(0 if r['passed'] else 1)
