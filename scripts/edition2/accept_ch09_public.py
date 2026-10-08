#!/usr/bin/env python3
"""Public Skills consumer checks with a tiny disposable external Go module.

The supplied source directory is used read-only through a module replacement;
source hashes must remain unchanged. No duplicate delivery tree is allocated.
Complete delivered-module discovery and historical behavior are separate gates.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()


def source_hashes(source):
    return {str(p.relative_to(source)):digest(p) for p in sorted(source.rglob('*')) if p.is_file() and '.git' not in p.parts and (p.suffix=='.go' or p.name in ('go.mod','go.sum'))}


def evaluate(source):
    fixture=Path(__file__).with_name('ch09-skills-public_test.go');identities=source_hashes(source)
    with tempfile.TemporaryDirectory(prefix='ch09-public-') as directory:
        work=Path(directory);(work/'go.mod').write_text('module example.com/ch09-independent-consumer\n\ngo 1.25\n\nrequire example.com/ensemble v0.0.0\nreplace example.com/ensemble => '+json.dumps(str(source))+'\n')
        shutil.copyfile(fixture,work/'skills_test.go');rows=[]
        for command in (['go','mod','tidy'],['gofmt','-l','skills_test.go'],['go','vet','./...'],['go','test','-race','-count=1','-timeout=45s','-run','^TestCh09Public','-v','./...']):
            run=subprocess.run(command,cwd=work,text=True,capture_output=True,timeout=180)
            passed=run.returncode==0 and (command[0]!='gofmt' or not run.stdout)
            rows.append(dict(command=command,exit=run.returncode,stdout=run.stdout,stderr=run.stderr,passed=passed))
            if not passed:break
    assert identities==source_hashes(source),'source changed during public checks'
    return dict(passed=len(rows)==4 and all(r['passed'] for r in rows),checks=rows,source_directory=str(source),source_files=identities,checker_files={p.name:digest(p) for p in [Path(__file__),fixture]},scope=__doc__)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source_directory',type=Path);p.add_argument('--receipt',type=Path);args=p.parse_args();result=evaluate(args.source_directory.resolve(strict=True))
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result,indent=2));raise SystemExit(int(not result['passed']))
