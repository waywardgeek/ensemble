#!/usr/bin/env python3
"""Independent Chapter 8 headless policy consumer in a disposable source copy.

Public API only: memory/persistent ownership, watch cut, lifecycle/path claims,
two actual Agent request limits, tool effects, recorded policy and replay.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


def evaluate(source):
    fixture = Path(__file__).with_name('ch08-policy-public_test.go')
    with tempfile.TemporaryDirectory(prefix='ch08-public-policy-') as temporary:
        target = Path(temporary)/'source'
        shutil.copytree(source,target,ignore=shutil.ignore_patterns('.git'))
        source_files={str(p.relative_to(target)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(target.rglob('*')) if p.is_file() and (p.suffix=='.go' or p.name in ('go.mod','go.sum'))}
        # Keep evidence and optional-module sources in the copied delivery.
        # No source is changed apart from adding the external consumer test.
        copied = target/'zz_ch08_independent_public_test.go'
        assert not copied.exists(), 'fixture would overwrite a delivered file'
        copied.write_bytes(fixture.read_bytes())
        rows=[]
        for command in (['gofmt','-l',str(copied)],['go','vet','./...'],
                        ['go','test','-race','-count=1','-timeout=45s','-run','^TestCh08Public','-v','.']):
            result=subprocess.run(command,cwd=target,text=True,capture_output=True,timeout=180)
            passed=result.returncode==0 and (command[0]!='gofmt' or not result.stdout)
            rows.append(dict(command=command,passed=passed,exit=result.returncode,stdout=result.stdout,stderr=result.stderr))
            if not passed: break
        return dict(scope=__doc__,passed=len(rows)==3 and all(r['passed'] for r in rows),checks=rows,
                    source_files=source_files,
                    checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in (Path(__file__),fixture)})


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_directory',type=Path)
    arguments=parser.parse_args()
    result=evaluate(arguments.source_directory.resolve(strict=True))
    print(json.dumps(result,indent=2))
    raise SystemExit(0 if result['passed'] else 1)
