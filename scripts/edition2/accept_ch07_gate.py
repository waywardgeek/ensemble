#!/usr/bin/env python3
"""Chapter 7 deterministic gate from immutable committed main source.

Live models/browser/audio, receipt acceptance, independent historical comparison
and final prose review are distinct gates. No paid calls are made here.
"""
import argparse
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile
from accept_ch05 import command

HERE=Path(__file__).resolve().parent
REPO=HERE.parents[1]
PREFIX='solutions/edition-2/main/'


def evaluate(revision):
    revision=subprocess.check_output(['git','rev-parse',revision+'^{commit}'],cwd=REPO,text=True).strip()
    archive=subprocess.check_output(['git','archive',revision,PREFIX],cwd=REPO)
    rows=[];hashes={}
    with tempfile.TemporaryDirectory(prefix='ch07-complete-gate-') as directory:
        root=Path(directory)/'source'
        with tarfile.open(fileobj=io.BytesIO(archive)) as source:
            for member in source:
                if not member.isfile():continue
                relative=Path(member.name).relative_to(PREFIX)
                if '..' in relative.parts:raise ValueError('unsafe archive path')
                if any(part in ('evidence','.git','node_modules','__pycache__') for part in relative.parts):continue
                data=source.extractfile(member).read();dest=root/relative
                dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(data)
                hashes[str(relative)]=hashlib.sha256(data).hexdigest()
        def run(name,args,cwd=root,timeout=1200):
            result=command(args,cwd,timeout=timeout);rows.append(dict(id=name,passed=result['exit']==0,command=result));return result['exit']==0
        cli=Path(directory)/'cli';gui=Path(directory)/'gui'
        cli_ok=run('build-headless-cli',['go','build','-o',cli,'./cmd'])
        gui_ok=run('build-optional-gui',['go','build','-o',gui,'./cmd/ensemble-gui'],root/'gui')
        if gui_ok:run('gui-wire-contract',['python3',HERE/'accept_ch07.py',gui])
        run('public-watch-pause',['python3',HERE/'accept_ch07_public.py',root])
        run('queue-projection-components',['python3',HERE/'accept_ch07_components.py',root])
        run('real-browser-controlled-components',['node',HERE/'accept_ch07_browser.cjs',root])
        run('implementation-deletions',['python3',HERE/'audit_ch07_mutations.py',root])
        run('browser-deletions',['python3',HERE/'audit_ch07_browser.py',root])
        run('incremental-projection-cost',['python3',HERE/'ch07-review-scaling.py',root])
        # Coordinator's independent actual-Connector probe is required; absence
        # remains a visible incomplete gate rather than silently skipping it.
        run('actual-connector-lifetime',['python3',HERE/'ch07-review-connector.py',root])
        if cli_ok:
            run('retained-cli-early-barriers',['python3',HERE/'accept_ch06_clients.py',cli])
            run('retained-wire-61',['python3',HERE/'accept_ch06_wire.py',cli])
        run('retained-ch06-contract',['python3',HERE/'accept_ch06.py',root])
        run('retained-ch05-assertions',['python3',HERE/'accept_ch06_prior.py',root])
        run('retained-ch06-deletions',['python3',HERE/'audit_ch06_mutations.py',root])
        run('retained-cli-overflow-deletions',['python3',HERE/'ch07-review-prior-cli.py',revision],REPO)
        for module in sorted(root.rglob('go.mod')):
            label=str(module.parent.relative_to(root));run(label+'/vet',['go','vet','./...'],module.parent)
            run(label+'/tests',['go','test','./...','-count=1','-timeout=120s'],module.parent)
        optional=root/'gui';detached=Path(directory)/'detached-gui';optional.rename(detached)
        try:run('core-with-optional-gui-absent',['go','test','./...','-count=1','-timeout=120s'])
        finally:detached.rename(optional)
    checkers={str(p.relative_to(HERE)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(HERE.iterdir()) if p.is_file() and ('ch07' in p.name or 'ch06' in p.name or 'ch05' in p.name)}
    return dict(scope=__doc__,source_revision=revision,source_files=hashes,checker_files=checkers,
                passed=all(row['passed'] for row in rows),checks=rows)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source_commit');a=p.parse_args();receipt=evaluate(a.source_commit)
    print(json.dumps(receipt,indent=2));raise SystemExit(0 if receipt['passed'] else 1)
