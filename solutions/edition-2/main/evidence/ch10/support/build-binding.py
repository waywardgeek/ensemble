"""Run ONLY when coordinator releases the serialized Go support build stage.

Build from matching committed own source. Record actual commands/dependencies;
never manufacture a binding for planned binaries. Existing outputs are preserved.
"""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
from identity import (HERE, PREFIX, SUPPORT_PREFIX, BUILD_ROLES, source_paths,
                      historical, digest, file_hash, dependency_paths, preflight, check_sources)


def build(args):
    repo=args.repo.resolve();revision=args.revision;out=args.output.resolve()
    sources={p:digest(historical(repo,revision,p)) for p in source_paths(repo,revision)}
    if not sources:raise ValueError('empty source set')
    check_sources({'source_revision':revision,'sources':sources},repo)
    out.mkdir()  # never overwrite an earlier build or evidence directory
    env={k:v for k,v in os.environ.items() if k not in ('GOFLAGS','GOWORK')}
    env.update(GOWORK='off')
    binaries={'python':sys.executable,'node':args.node,'chrome':args.chrome,'recorder':'/usr/bin/script'}
    builds={};modules={p:h for p,h in sources.items() if Path(p).name in ('go.mod','go.sum')}
    ledger=out/'build-commands.jsonl'
    def command(argv,cwd):
        result=subprocess.run(argv,cwd=cwd,env=env,text=True,capture_output=True)
        with ledger.open('a') as f: f.write(json.dumps({'argv':argv,'cwd':str(cwd),'exit':result.returncode,'stdout':result.stdout,'stderr':result.stderr})+'\n');f.flush();os.fsync(f.fileno())
        result.check_returncode();return result.stdout
    version=command(['go','version'],repo/PREFIX)
    for role,(directory,target) in BUILD_ROLES.items():
        cwd=repo/PREFIX/directory;binary=out/role
        command(['go','vet','./...'],cwd);command(['go','test','./...','-count=1','-timeout=180s'],cwd)
        graph=command(['go','list','-m','-json','all'],cwd)
        command(['go','build','-trimpath','-o',str(binary),target],cwd)
        info=command(['go','version','-m',str(binary)],cwd)
        binaries[role]=str(binary)
        builds[role]={'source_revision':revision,'binary_sha256':file_hash(binary),'module':PREFIX+directory,'target':target,'exit':0,'modules':modules,'go_version':version,'module_graph':graph,'binary_build_info':info}
    binding={'source_revision':revision,'sources':sources,'support':{p:h for p,h in sources.items() if p.startswith(SUPPORT_PREFIX)},'modules':modules,'builds':builds,
             'binaries':{k:{'path':str(Path(v).resolve()),'sha256':file_hash(v)} for k,v in binaries.items()},
             'browser_dependencies':{'root':str(args.browser_root.resolve()),'files':{p:file_hash(args.browser_root/p) for p in dependency_paths(args.browser_root)}}}
    binding['catalog']={p:h for p,h in binding['support'].items() if '/catalog/' in p or p.endswith('/bindings.json')}
    binding['schedule_sha256']=binding['support'][SUPPORT_PREFIX+'schedule.json']
    preflight(binding,repo)
    (out/'binding.json').write_text(json.dumps(binding,indent=2)+'\n')


if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--repo',required=True,type=Path);p.add_argument('--revision',required=True);p.add_argument('--output',required=True,type=Path);p.add_argument('--node',required=True);p.add_argument('--chrome',required=True);p.add_argument('--browser-root',required=True,type=Path)
    build(p.parse_args())
