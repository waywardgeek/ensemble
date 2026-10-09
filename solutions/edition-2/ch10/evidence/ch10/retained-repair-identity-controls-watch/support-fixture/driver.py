"""Manual PTY/public launch preparation. Local mode only in this phase.

No credential reader/discovery entrypoint. A future live release must supply a
reviewed in-memory upstream adapter; this command refuses non-loopback origins.
"""
import argparse
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import threading
from urllib.parse import urlsplit
from identity import HERE, preflight, digest, canonical, file_hash, check_launch
from relay import Budget, Relay, SCHEDULE


def launch_record(binding, vendor, model, role, command, env, workspace, run):
    value = dict(binding_sha256=digest(canonical(binding)), source_revision=binding['source_revision'],
                 vendor=vendor, model=model, role=role, command=command,
                 command_sha256=digest(canonical(command)), environment=env,
                 workspace=str(workspace), provider_root=str(Path(workspace).parent), run=str(run), mode='local', actor='student coder; not Bill')
    value['identity_sha256'] = digest(canonical(value))
    check_launch(binding, value)
    return value


def seal_gui(binding, run):
    launch = json.loads((run/'launch.json').read_text()); check_launch(binding, launch)
    raw = (run/'terminal.txt').read_text(errors='replace')
    matches = set(re.findall(r'http://127\.0\.0\.1:[0-9]+', raw))
    if len(matches) != 1: raise ValueError('one actual printed GUI origin required')
    launch['gui_url'] = matches.pop()
    launch['terminal_at_attach_sha256'] = file_hash(run/'terminal.txt')
    launch.pop('identity_sha256'); launch['identity_sha256'] = digest(canonical(launch))
    # Separate immutable attachment record; original launch remains preserved.
    with (run/'browser-launch.json').open('x') as f: json.dump(launch,f,indent=2)


def launch(args, binding):
    vendor=args.vendor; root=args.root.resolve(); run=args.run.resolve()
    if args.step not in ('A1','A2','B','C','D1','D-prepare','D-seed','D2','offline'): raise ValueError('step')
    root.mkdir(exist_ok=True); workspace=root/'workspace'; workspace.mkdir(exist_ok=True)
    run.mkdir(); routes = ['C1','C2','C3'] if args.step=='C' else ([args.step] if args.step in SCHEDULE['steps'] else ['D2'])
    relay=Relay(args.origin,{},Budget(root/'attempts.jsonl',vendor),run,routes,local_only=True)
    thread=threading.Thread(target=relay.serve_forever); thread.start()
    base=f'http://127.0.0.1:{relay.server_port}'
    role = SCHEDULE['steps'].get(args.step,{}).get('role','consumer')
    env={'LLM_VENDOR':vendor,'LLM_MODEL':args.model,'LLM_RESOLVED_MODEL':args.model.removeprefix('models/'),'LLM_BASE_URL':base+'/'+args.step,'EN_DISABLE_STREAMING':'0'}
    if args.step.startswith('D'):
        env.update(LLM_SKILLS_DIR=str(args.catalog.resolve()),LLM_PRIMARY_SKILL='base')
    command=[binding['binaries'][role]['path']]
    if role=='cli':command+=['--session-dir',str(root/('D' if args.step=='D1' else 'A')),'chat']
    elif role=='gui':command+=['--session-dir',str(root/'A'),'--terminal','--port','0','--policy',str(root/'current-policy.json'),'--preferences',str(root/'preferences.json')]
    else:command+=['--root',str(root),'--output',str(run/'public'),'--relay',base,'--step',args.step,'--catalog',str(args.catalog.resolve())]
    value=launch_record(binding,vendor,args.model,role,command,env,workspace,run)
    (run/'launch.json').write_text(json.dumps(value,indent=2)+'\n')
    # Keep external credentials out of children; the local relay placeholder is
    # deliberately not a provider credential. Never read the settings file.
    child={k:os.environ[k] for k in ('PATH','TERM','LANG','LC_ALL','TMPDIR') if k in os.environ}
    child.update(env,LLM_API_KEY='local-relay-placeholder')
    process=None; timer=None
    try:
        if role in ('cli','gui'):
            print('Manual PTY: enter the scheduled prompt once, observe, then local commands; no automatic input.',flush=True)
            print(SCHEDULE['steps'][args.step]['prompt'].replace('-P','-'+vendor),flush=True)
            command=[binding['binaries']['recorder']['path'],'-q','-F',str(run/'terminal.txt'),*command]
            process=subprocess.Popen(command,cwd=workspace,env=child,start_new_session=True)
        else:
            stream=(run/'consumer-original.txt').open('xb')
            process=subprocess.Popen(command,cwd=workspace,env=child,stdout=stream,stderr=subprocess.STDOUT,start_new_session=True)
        timer=threading.Timer(600,lambda:os.killpg(process.pid,signal.SIGTERM) if process.poll() is None else None);timer.start()
        code=process.wait(timeout=620)
        if role=='consumer':stream.close()
        (run/'exit.json').write_text(json.dumps({'exit':code,'mode':'local','live':False})+'\n')
    finally:
        if timer:timer.cancel()
        if process and process.poll() is None:os.killpg(process.pid,signal.SIGTERM)
        relay.shutdown();relay.server_close();thread.join()
        if (root/'attempts.jsonl').exists():
            (run/'attempts-final.jsonl').write_bytes((root/'attempts.jsonl').read_bytes())
    return code


if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--binding',type=Path,required=True);p.add_argument('--repo',type=Path,required=True)
    p.add_argument('--seal-gui',type=Path);p.add_argument('--vendor',choices=SCHEDULE['providers']);p.add_argument('--model');p.add_argument('--step');p.add_argument('--root',type=Path);p.add_argument('--run',type=Path);p.add_argument('--origin');p.add_argument('--catalog',type=Path,default=HERE/'catalog')
    a=p.parse_args();b=json.loads(a.binding.read_text());preflight(b,a.repo)
    if a.seal_gui:seal_gui(b,a.seal_gui)
    else:
        if not all((a.vendor,a.model,a.step,a.root,a.run,a.origin)):p.error('launch arguments required')
        # Validate upstream BEFORE directory or receipt creation.
        u=urlsplit(a.origin)
        if u.scheme!='http' or u.hostname!='127.0.0.1' or u.username or u.password or u.path not in ('','/') or u.query:p.error('only literal loopback fixture upstream is released')
        raise SystemExit(launch(a,b))
