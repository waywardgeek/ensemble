#!/usr/bin/env python3
"""Controlled disk stages on actual settings owners in a disposable source copy.

The contract-derived checks hold real worker boundaries and inject pre-replace
failures. The adapter substitutes only OS call sites, preserving actor/service
admission, candidate application, subscriptions, close and socket loops.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


def change(text, old, new):
    assert text.count(old)==1, 'disk instrumentation anchor is not unique: '+old
    return text.replace(old,new)


def instrument(root):
    here=Path(__file__).parent
    for relative,package in [('internal/policy/policy.go','policy'),('gui/internal/preferences/preferences.go','preferences')]:
        file=root/relative
        text=file.read_text()
        for old,new in [('type Service struct {','type Service struct {\n ch08Control *Ch08DiskControl'),
                        ('os.CreateTemp(filepath.Dir(s.path),','ch08Create(s, filepath.Dir(s.path),'),
                        ('f.Write(data)','ch08Write(s, f, data)'),
                        ('f.Sync()','ch08Sync(s, f)'),
                        ('closeErr := f.Close()','closeErr := ch08Close(s, f)'),
                        ('return os.Rename(name, s.path)','return ch08Rename(s, name, s.path)')]:
            text=change(text,old,new)
        file.write_text(text)
        helper=(here/'ch08-disk-control.go.tmpl').read_text().replace('package PACKAGE','package '+package)
        if package=='preferences':
            helper+='\nfunc(s *Service) Ch08WatchCount()int{s.mu.Lock();defer s.mu.Unlock();return len(s.watches)}\n'
        (file.parent/'zz_ch08_disk_control.go').write_text(helper)
    (root/'zz_ch08_disk_control.go').write_text('''package ensemble
import "example.com/ensemble/internal/policy"
type Ch08DiskControl = policy.Ch08DiskControl
func(a *Agent) Ch08ArmPolicy(stage string, fail bool)*Ch08DiskControl { return a.policy.Ch08Arm(stage,fail) }
''')
    (root/'gui/zz_ch08_persistence_test.go').write_bytes((here/'ch08-persistence_test.go').read_bytes())


def evaluate(source):
    here=Path(__file__).parent
    with tempfile.TemporaryDirectory(prefix='ch08-persistence-') as temporary:
        root=Path(temporary)/'source'
        shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git'))
        hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file() and (p.suffix=='.go' or p.name in ('go.mod','go.sum'))}
        instrument(root)
        changed=[root/'internal/policy/policy.go',root/'gui/internal/preferences/preferences.go',*root.rglob('zz_ch08*.go')]
        formatted=subprocess.run(['gofmt','-w',*map(str,changed)],capture_output=True,text=True)
        assert formatted.returncode==0,formatted.stderr
        rows=[]
        for command in [['go','vet','./...'],['go','test','-race','-count=1','-timeout=60s','-run','^TestCh08Persistence','-v','.']]:
            result=subprocess.run(command,cwd=root/'gui',text=True,capture_output=True,timeout=180)
            rows.append(dict(command=command,passed=result.returncode==0,exit=result.returncode,stdout=result.stdout,stderr=result.stderr))
            if result.returncode: break
        return dict(scope=__doc__,passed=len(rows)==2 and all(r['passed'] for r in rows),checks=rows,source_files=hashes,
                    checker_files={name:hashlib.sha256((here/name).read_bytes()).hexdigest() for name in ['accept_ch08_persistence.py','ch08-disk-control.go.tmpl','ch08-persistence_test.go']})


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory',type=Path)
    args=parser.parse_args()
    result=evaluate(args.source_directory.resolve(strict=True))
    print(json.dumps(result,indent=2))
    raise SystemExit(0 if result['passed'] else 1)
