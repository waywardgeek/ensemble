#!/usr/bin/env python3
"""Published 8f24360 CLI System input, request capture and offline-base checks."""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile
import threading
from accept_ch09 import ROOT, MODELS, Fixture, definition, environment, sha, system_text

PRIMARY='Environment primary.\n'
OVERRIDE='Explicit operator override.'
CASES=[(False,'absent'),(False,'empty'),(False,'override'),(False,'fallback'),
       (True,'absent'),(True,'empty'),(True,'override'),(True,'equal'),(True,'fallback')]


def run(binary,vendor,skills,setting):
    server=Fixture(vendor,[[]]);thread=threading.Thread(target=server.serve_forever);thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix='ch09-config-') as directory:
            work=Path(directory);env=environment(work,vendor,f'http://127.0.0.1:{server.server_port}')
            if skills:
                p=work/'catalog/base/SKILL.md';p.parent.mkdir(parents=True);p.write_bytes(definition('base','primary',PRIMARY,'read_file'))
                env.update(LLM_SKILLS_DIR='catalog',LLM_PRIMARY_SKILL='base')
            if setting in ('empty','override','equal'):env['LLM_SYSTEM']={'empty':'','override':OVERRIDE,'equal':PRIMARY}[setting]
            if setting=='fallback':env[vendor.upper()+'_SYSTEM']='Not a supported system fallback.'
            p=subprocess.run([str(binary),'protocol'],input='{"user":"Return one local answer."}\n',text=True,capture_output=True,cwd=work,env=env,timeout=15)
            log=work/'session.jsonl';raw=log.read_text() if log.exists() else '';events=[json.loads(x) for x in raw.splitlines()][1:]
            replay={}
            if skills and setting=='absent' and p.returncode==0:
                for field in (work/'catalog').rglob('SKILL.md'):field.unlink()
                for mode,value in [('empty',''),('conflict',OVERRIDE)]:
                    offline=environment(work,vendor,'http://127.0.0.1:1');offline['LLM_SYSTEM']=value
                    r=subprocess.run([str(binary),'render',str(log)],text=True,capture_output=True,cwd=work,env=offline,timeout=10)
                    replay[mode]=dict(exit=r.returncode,stdout=r.stdout,stderr=r.stderr)
            return dict(exit=p.returncode,stdout=p.stdout,stderr=p.stderr,requests=server.requests,events=events,log_text=raw,replay=replay)
    finally:server.shutdown();server.server_close();thread.join()


def evaluate(vendor,skills,setting,run,default):
    conflict=skills and setting in ('override','equal')
    if conflict:return run['exit']!=0 and bool(run['stderr'].strip()) and not run['requests'] and not any(x.get('type')=='skills_initialized' for x in run['events'])
    requests=run['requests'];captures=[x['request']['configuration'] for x in run['events'] if x.get('type')=='request_sent']
    expected=PRIMARY if skills else OVERRIDE if setting=='override' else default
    if run['exit']!=0 or len(requests)!=1 or len(captures)!=1 or not expected:return False
    if system_text(vendor,requests[0])!=expected or 'system' not in captures[0] or captures[0]['system']!=('' if skills else expected):return False
    if skills:
        if run['log_text'].count(json.dumps(PRIMARY))!=1:return False
        if setting=='absent':
            replay=run['replay'];positive=replay.get('empty',{});negative=replay.get('conflict',{})
            try:body=json.loads(positive.get('stdout',''))
            except ValueError:return False
            if positive.get('exit')!=0 or system_text(vendor,body)!=PRIMARY or negative.get('exit')==0 or negative.get('stdout','').strip() or not negative.get('stderr','').strip():return False
    return True


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);p.add_argument('--receipt',type=Path);args=p.parse_args();binary=args.binary.resolve()
    identities={str(path):sha(path.read_bytes()) for path in [Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-09.md',binary]}
    runs={};checks=[]
    for vendor in MODELS:
        values={(skills,setting):run(binary,vendor,skills,setting) for skills,setting in CASES}
        baseline=values[(False,'absent')];default=system_text(vendor,baseline['requests'][0]) if baseline['requests'] else ''
        for (skills,setting),result in values.items():
            name=f'{vendor}/skills-{skills}/{setting}';runs[name]=result;checks.append(dict(id=name,passed=evaluate(vendor,skills,setting,result,default)))
    assert identities=={path:sha(Path(path).read_bytes()) for path in identities},'input identity changed'
    result=dict(passed=sum(x['passed'] for x in checks),total=len(checks),checks=checks,runs=runs,input_sha256=identities,scope=__doc__)
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='runs'},indent=2));return int(result['passed']!=result['total'])


if __name__=='__main__':raise SystemExit(main())
