#!/usr/bin/env python3
"""Contract-aligned Edition 2 Chapter 5 checker (inherited score categories).

Copies the source and installs independent probes in the disposable copy only.
The public API spelling in ch05_contract_test.go.txt is an explicit adapter;
identifiers and historical lineage never earn points. Legacy graders are unchanged.
Real-provider PTY/workflow receipts and independent semantic ownership review remain
separate mandatory acceptance gates, not claims made by this local checker.
"""
import argparse
import hashlib
import json
import pathlib
import shutil
import subprocess
import tempfile

HERE = pathlib.Path(__file__).resolve().parent
REPO = HERE.parents[1]
CATEGORIES = [('parity',10,'Parity'),('responsiveness',25,'Responsiveness'),
              ('replay',15,'Replay'),('observers',15,'Observers'),
              ('completion-collection',15,'Collection'),('loud-refusal',10,'Refusal'),
              ('architecture',10,'Architecture')]


def command(args,cwd,timeout=180):
    try:
        p=subprocess.run([str(x) for x in args],cwd=cwd,capture_output=True,text=True,timeout=timeout)
        return dict(args=[str(x) for x in args],exit=p.returncode,stdout=p.stdout,stderr=p.stderr)
    except subprocess.TimeoutExpired as err:
        return dict(args=[str(x) for x in args],exit=124,stdout=str(err.stdout or ''),stderr='checker command timed out')
    except OSError as err:
        return dict(args=[str(x) for x in args],exit=127,stdout='',stderr=str(err))


def install(source):
    if REPO == source or REPO in source.parents:
        raise ValueError('probes require a disposable copy outside repository')
    for fixture in ('contract', 'ownership', 'stale', 'report'):
        target=(source/'internal/llm' if fixture=='stale' else source/'internal/jobs' if fixture=='report' else source)/('independent_ch05_'+fixture+'_test.go')
        target.write_bytes((HERE/('ch05_'+fixture+'_test.go.txt')).read_bytes())
        result=command(['gofmt','-w',target],source)
        if result['exit']:raise RuntimeError(result)


def evaluate(source, only=None, race=True, prior=True):
    checks=[]
    with tempfile.TemporaryDirectory(prefix='ensemble-ch05-independent-') as tmp:
        root=pathlib.Path(tmp)/'source'
        shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence'))
        source_files={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file() and (p.suffix=='.go' or p.name in ('go.mod','go.sum'))}
        install(root)
        for name,points,prefix in CATEGORIES:
            if only and name not in only:continue
            commands=[]
            if name=='architecture':
                # Discover nested modules too: every executable is checked.
                for module in sorted(root.rglob('go.mod')):
                    commands.append(command(['go','run',HERE/'packagecheck/main.go',module.parent],REPO))
                    commands.append(command(['go','test','-run','^$','./...'],module.parent))
                commands.append(command(['go','run',HERE/'ch05sourcecheck/main.go',root],REPO))
                commands.append(command(['go','list','-deps','./...'],root))
                workflow=pathlib.Path(tmp)/'workflow'
                built=command(['go','build','-o',workflow,'.'],root/'examples/workflow');commands.append(built)
                if built['exit']==0:
                    commands.append(command(['python3',HERE/'accept_ch05_clients.py',workflow,'--workflow'],REPO))
                commands.append(command(['go','test','-run','^TestIndependentArchitecture','-count=1','-timeout=30s','.'],root))
            else:
                cmd=['go','test','-json','-run','^TestIndependent'+prefix,'-count=1','-timeout=60s']
                if race:cmd.append('-race')
                cmd.append('./...')
                result=command(cmd,root)
                # No-tests-found is never passing coverage.
                try:ran=any(json.loads(line).get('Action')=='run' and '/' not in json.loads(line).get('Test','') for line in result['stdout'].splitlines())
                except (ValueError,TypeError):ran=False
                if not ran:result['exit']=result['exit'] or 1;result['stderr']+='\nNo independent tests executed.'
                commands.append(result)
                if name=='responsiveness':
                    binary=pathlib.Path(tmp)/'cli-controls'
                    built=command(['go','build','-o',binary,'./cmd'],root);commands.append(built)
                    if built['exit']==0:
                        commands.append(command(['python3',HERE/'accept_ch05_clients.py',binary],REPO))
                if name=='parity' and prior:
                    binary=pathlib.Path(tmp)/'cli'
                    built=command(['go','build','-o',binary,'./cmd'],root);commands.append(built)
                    if built['exit']==0:
                        commands.append(command(['python3',HERE/'accept_ch05_prior.py',binary],REPO,600))
            passed=all(c['exit']==0 for c in commands)
            checks.append(dict(id=name,points=points,earned=points if passed else 0,passed=passed,commands=commands))
    return dict(scope=__doc__,full_local_scope=not only and race and prior,source=str(source),checker_files={str(p.relative_to(HERE)):hashlib.sha256(p.read_bytes()).hexdigest() for p in [pathlib.Path(__file__),HERE/'ch05_contract_test.go.txt',HERE/'ch05_ownership_test.go.txt',HERE/'ch05_stale_test.go.txt',HERE/'ch05_report_test.go.txt',HERE/'accept_ch05_clients.py',HERE/'accept_ch05_prior.py',HERE/'ch05sourcecheck/main.go',HERE/'packagecheck/main.go']},probe_sha256=hashlib.sha256((HERE/'ch05_contract_test.go.txt').read_bytes()).hexdigest(),
                source_files=source_files,score=sum(c['earned'] for c in checks),maximum=sum(c['points'] for c in checks),passed=all(c['passed'] for c in checks),checks=checks)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=pathlib.Path);p.add_argument('--only',action='append',choices=[c[0] for c in CATEGORIES]);p.add_argument('--no-race',action='store_true');p.add_argument('--skip-prior',action='store_true',help='targeted iteration only; not full acceptance')
    args=p.parse_args();result=evaluate(args.source.resolve(strict=True),args.only,not args.no_race,not args.skip_prior)
    print(json.dumps(result,indent=2));return 0 if result['passed'] else 1

if __name__=='__main__':raise SystemExit(main())
