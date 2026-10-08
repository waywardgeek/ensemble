#!/usr/bin/env python3
"""Immutable Chapter 8 deterministic acceptance, retaining prior behavioral checks.

Live providers/audio, historical comparison and final manuscript acceptance are
separate gates. --only is a preparation run, never complete chapter acceptance.
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
from ch08_retained_adapters import prepare

HERE=Path(__file__).resolve().parent
REPO=HERE.parents[1]
PREFIX='solutions/edition-2/main/'


def evaluate(revision, only=None):
    revision=subprocess.check_output(['git','rev-parse',revision+'^{commit}'],cwd=REPO,text=True).strip()
    archive=subprocess.check_output(['git','archive',revision,PREFIX],cwd=REPO)
    rows=[];hashes={}
    with tempfile.TemporaryDirectory(prefix='ch08-full-gate-') as directory:
        temporary=Path(directory).resolve();root=temporary/'source';bundle=temporary/'checkers'
        with tarfile.open(fileobj=io.BytesIO(archive)) as source:
            for member in source:
                if member.isdir():continue
                if not member.isfile():raise ValueError('unsupported delivered archive entry: '+member.name)
                relative=Path(member.name).relative_to(PREFIX)
                if '..' in relative.parts or relative.is_absolute():raise ValueError('unsafe archive path')
                data=source.extractfile(member).read();dest=root/relative
                dest.parent.mkdir(parents=True,exist_ok=True);dest.write_bytes(data)
                hashes[str(relative)]=hashlib.sha256(data).hexdigest()
        assert hashes, 'empty source set'
        adapters=prepare(bundle,root)
        cli=temporary/'cli';gui=temporary/'gui'
        def run(name,args,cwd=root,timeout=1800,required=False):
            if only and name not in only and not required:return None
            result=command(args,cwd,timeout=timeout)
            rows.append(dict(id=name,passed=result['exit']==0,command=result))
            return result['exit']==0
        cli_ok=run('build-headless-cli',['go','build','-o',cli,'./cmd'],required=True)
        gui_ok=run('build-optional-gui',['go','build','-o',gui,'./cmd/ensemble-gui'],root/'gui',required=True)
        run('complete-delivered-package-discovery',['python3',HERE/'accept_delivered_tree.py',revision],REPO)
        run('module-aware-package-checker-controls',['go','test','./scripts/edition2/ch08packagecheck','-count=1'],REPO)
        run('historical-package-checker-controls',['go','test','./scripts/edition2/packagecheck','-count=1'],REPO)
        for name,script in [('settings-wire','accept_ch08.py'),('strict-files-and-patches','accept_ch08_validation.py'),('command-shapes','accept_ch08_commands.py'),('actual-policy-effects','accept_ch08_policy_effect.py')]:
            if gui_ok:run(name,['python3',HERE/script,gui])
        for name,script in [('public-policy','accept_ch08_public.py'),('disk-concurrency','accept_ch08_persistence.py'),('policy-disk-deletions','audit_ch08_policy.py'),('settings-browser-deletions','audit_ch08_browser.py'),('snapshot-handoff-deletions','accept_ch08_handoff.py')]:
            run(name,['python3',HERE/script,root])
        run('settings-browser',['node',HERE/'accept_ch08_browser.cjs',root])
        run('cross-tab-native-lease-deletions',['python3',HERE/'audit_ch08_speech_tabs.py',root])
        if gui_ok:run('retained-gui-wire',['python3',bundle/'accept_ch07.py',gui])
        for name,script in [('retained-public-watch-pause','accept_ch07_public.py'),('retained-queue-projection','accept_ch07_components.py'),('retained-watch-deletions','audit_ch07_mutations.py'),('retained-browser-deletions','audit_ch07_browser.py'),('retained-projection-scaling','ch07-review-scaling.py'),('retained-connector-deletions','ch07-review-connector-mutations.py'),('retained-ch06-contract','accept_ch06.py'),('retained-ch05-assertions','accept_ch07_prior.py'),('retained-ch06-deletions','audit_ch06_mutations.py')]:
            run(name,['python3',bundle/script,root])
        run('retained-browser',['node',bundle/'accept_ch07_browser.cjs',root])
        if cli_ok:
            run('retained-cli-early-barriers',['python3',bundle/'accept_ch06_clients.py',cli])
            run('retained-wire-61',['python3',bundle/'accept_ch06_wire.py',cli])
        run('retained-cli-overflow-deletions',['python3',bundle/'ch07-review-prior-cli.py',revision],REPO)
        run('retained-ch07-comparative-repairs',['python3',bundle/'ch07-review-revisions.py',revision],REPO)
        for module in sorted(root.rglob('go.mod')):
            label=str(module.parent.relative_to(root))
            run('module/'+label+'/vet',['go','vet','./...'],module.parent)
            run('module/'+label+'/tests',['go','test','./...','-count=1','-timeout=120s'],module.parent)
        optional=root/'gui';detached=temporary/'detached-gui';optional.rename(detached)
        try:run('headless-without-optional-gui',['go','test','./...','-count=1','-timeout=120s'])
        finally:detached.rename(optional)
        if only:
            absent=set(only)-{row['id'] for row in rows}
            assert not absent, 'unknown/unrun requested checks: '+str(sorted(absent))
    checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in HERE.iterdir() if p.is_file() and ('ch08' in p.name or p.name=='accept_delivered_tree.py')}
    return dict(scope=__doc__,source_revision=revision,source_files=hashes,checker_files=checker_files,retained_adapters=adapters,
                complete_run=not bool(only),selected_checks=only,passed=all(row['passed'] for row in rows),checks=rows)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('source_commit');parser.add_argument('--only',action='append')
    args=parser.parse_args();result=evaluate(args.source_commit,args.only);print(json.dumps(result,indent=2));raise SystemExit(0 if result['passed'] else 1)
