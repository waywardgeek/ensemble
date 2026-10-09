"""Deliberate local/live launcher; live requires bound discovery and fixed adapter.

Live entrypoints are prepared for a separate coordinator release. Local mode
never loads settings; children receive only a dummy loopback credential.
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
from provider import Provider, load_key, require_discovery


def launch_record(binding, vendor, model, role, command, env, workspace, run, mode="local", discovery_sha256=None):
    value = dict(binding_sha256=digest(canonical(binding)), source_revision=binding['source_revision'],
                 vendor=vendor, model=model, role=role, command=command,
                 command_sha256=digest(canonical(command)), environment=env,
                 workspace=str(workspace), provider_root=str(Path(workspace).parent), run=str(run), mode=mode, actor='student coder; not Bill')
    if 'support_revision' in binding: value['support_revision'] = binding['support_revision']
    if discovery_sha256: value['discovery_sha256'] = discovery_sha256
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
    if args.step.startswith('D'):
        expected={p.split('/catalog/',1)[1]:h for p,h in binding['catalog'].items() if '/catalog/' in p}
        actual={str(p.relative_to(args.catalog)):file_hash(p) for p in args.catalog.rglob('*') if p.is_file()}
        if not expected or actual != expected or any(p.is_symlink() for p in args.catalog.rglob('*')): raise ValueError('selected catalog identity mismatch')
    mode = getattr(args,'mode','local'); adapter = None; discovery_hash = None
    if mode == 'live':
        if args.step not in SCHEDULE['steps'] and args.step != 'C': raise ValueError('local support step cannot use live adapter')
        if args.origin or not args.discovery: raise ValueError('live uses fixed origin and bound discovery')
        receipt = json.loads(args.discovery.read_text())
        require_discovery(receipt,vendor,args.model,digest(canonical(binding)),root/'attempts.jsonl')
        discovery_hash = file_hash(args.discovery)
        adapter = Provider(vendor,load_key(vendor),args.model)
    else:
        u=urlsplit(args.origin or '')
        if u.scheme!='http' or u.hostname!='127.0.0.1' or u.username or u.password or u.path not in ('','/') or u.query or u.fragment: raise ValueError('local mode requires literal loopback fixture')
    root.mkdir(exist_ok=True); workspace=root/'workspace'; workspace.mkdir(exist_ok=True)
    run.mkdir(); routes = ['C1','C2','C3'] if args.step=='C' else ([args.step] if args.step in SCHEDULE['steps'] else ['D2'])
    relay=Relay(adapter.origin if adapter else args.origin,adapter.headers if adapter else {},Budget(root/'attempts.jsonl',vendor),run,routes,local_only=adapter is None,provider=adapter)
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
    value=launch_record(binding,vendor,args.model,role,command,env,workspace,run,mode,discovery_hash)
    (run/'launch.json').write_text(json.dumps(value,indent=2)+'\n')
    # Provider keys belong only to the relay parent, never children/launch receipts.
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
        (run/'exit.json').write_text(json.dumps({'exit':code,'mode':mode,'live':mode=='live'})+'\n')
    finally:
        if timer:timer.cancel()
        if process and process.poll() is None:os.killpg(process.pid,signal.SIGTERM)
        relay.shutdown();relay.server_close();thread.join()
        if (root/'attempts.jsonl').exists():
            (run/'attempts-final.jsonl').write_bytes((root/'attempts.jsonl').read_bytes())
    return code


if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('--binding',type=Path,required=True);p.add_argument('--repo',type=Path,required=True)
    p.add_argument('--seal-gui',type=Path);p.add_argument('--vendor',choices=SCHEDULE['providers']);p.add_argument('--model');p.add_argument('--step');p.add_argument('--root',type=Path);p.add_argument('--run',type=Path);p.add_argument('--origin');p.add_argument('--mode',choices=('local','live'),default='local');p.add_argument('--discovery',type=Path);p.add_argument('--catalog',type=Path,default=HERE/'catalog')
    a=p.parse_args();b=json.loads(a.binding.read_text());preflight(b,a.repo)
    if a.seal_gui:seal_gui(b,a.seal_gui)
    else:
        if not all((a.vendor,a.model,a.step,a.root,a.run)):p.error('launch arguments required')
        try: raise SystemExit(launch(a,b))
        except (ValueError, OSError):
            # Never print arbitrary exception text, settings, headers or URLs.
            raise SystemExit('bounded launch refused; inspect nonsecret receipts') from None
