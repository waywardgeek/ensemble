#!/usr/bin/env python3
"""Initial Chapter 10 session CLI/outer-format checks; local HTTP only.

Private state codec, public import, fault injection, GUI and full limits remain
separate. Corruption/lock checks require a genuine created/inspected session.
"""
import argparse
import copy
from decimal import Decimal, localcontext
import fcntl
import hashlib
import http.server
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import tempfile
import threading
import time

from accept_ch09 import ROOT, MODELS, environment, response

CANARY='CH10_CONFIG_KEY_PRIVATE_71c882'
MAX=18446744073709551615
OUTER={'version','session_id','identity','as_of','high_watermarks','state_version','state','state_sha256'}


class Number(str):pass


def parse(raw):
    def pairs(items):
        result={}
        for name,value in items:
            if name in result:raise ValueError('duplicate member')
            result[name]=value
        return result
    def invalid(_):raise ValueError('non-JSON number')
    return json.loads(raw.decode('utf-8') if isinstance(raw,bytes) else raw,object_pairs_hook=pairs,parse_int=Number,parse_float=Number,parse_constant=invalid)


def number(value):
    match=re.fullmatch(r'(-?)(0|[1-9][0-9]*)(?:\.([0-9]+))?(?:[eE]([+-]?[0-9]+))?',str(value))
    if not match:raise ValueError('number token')
    sign,whole,fraction,exponent=match.groups();fraction=fraction or '';digits=(whole+fraction).lstrip('0')
    if not digits:return '0'
    # Arithmetic concerns exponent digits, never an expanded power-of-ten value.
    trimmed=digits.rstrip('0');delta=len(digits)-len(trimmed)-len(fraction)
    with localcontext() as context:
        context.prec=max(len(exponent or '0'),len(str(delta)))+2
        exp=format(Decimal(exponent or '0')+Decimal(delta),'f')
    return sign+trimmed+('e'+exp if exp not in ('0','-0') else '')


def canonical(value):
    if isinstance(value,Number):return number(value)
    if value is None:return 'null'
    if value is True:return 'true'
    if value is False:return 'false'
    if isinstance(value,str):return json.dumps(value,ensure_ascii=False,separators=(',',':'))
    if isinstance(value,int):return number(Number(str(value)))
    if isinstance(value,list):return '['+','.join(canonical(x) for x in value)+']'
    if isinstance(value,dict):return '{'+','.join(canonical(k)+':'+canonical(value[k]) for k in sorted(value,key=lambda k:k.encode('utf-8')))+'}'
    raise ValueError('non-JSON value')


def digest(raw):return hashlib.sha256(raw).hexdigest()

def uint(value,positive=False):
    if isinstance(value,bool) or not isinstance(value,(Number,int)) or not re.fullmatch(r'0|[1-9][0-9]*',str(value)):raise ValueError('integer type')
    if len(str(value))>20:raise ValueError('integer range')
    value=int(value)
    if value< int(positive) or value>MAX:raise ValueError('integer range')
    return value


def envelope(raw):
    value=parse(raw)
    if not isinstance(value,dict) or set(value)!=OUTER:raise ValueError('outer fields')
    if uint(value['version'])!=1 or uint(value['state_version'])!=1:raise ValueError('version')
    if type(value['session_id']) is not str or not re.fullmatch('[0-9a-f]{32}',value['session_id']):raise ValueError('session_id')
    anchor=uint(value['as_of'],True);water=value['high_watermarks']
    if not isinstance(water,dict) or set(water)!={'event','request','activation','job'}:raise ValueError('watermark fields')
    if any(uint(v)>MAX for v in water.values()) or uint(water['event'])!=anchor:raise ValueError('event anchor')
    identity=value['identity']
    if not isinstance(identity,dict) or set(identity)!={'mode','system','skills','handlers'}:raise ValueError('identity fields')
    if identity['mode']=='plain':
        if type(identity['system']) is not str or not identity['system'] or identity['skills'] is not None:raise ValueError('plain identity')
    elif identity['mode']=='skills':
        skills=identity['skills']
        if identity['system'] is not None or not isinstance(skills,dict) or set(skills)!={'primary','catalog_sha256','bindings_sha256'}:raise ValueError('skill identity')
        if not re.fullmatch('[a-z][a-z0-9-]{0,63}',skills['primary']) or not all(re.fullmatch('[0-9a-f]{64}',skills[k]) for k in ('catalog_sha256','bindings_sha256')):raise ValueError('skill identity')
    else:raise ValueError('identity mode')
    handlers=identity['handlers']
    if not isinstance(handlers,list) or len(handlers)>1024:raise ValueError('handlers')
    if any(not isinstance(h,dict) or set(h)!={'name','description','schema'} for h in handlers):raise ValueError('handler fields')
    names=[h['name'] for h in handlers]
    if not all(isinstance(n,str) and re.fullmatch('[a-z][a-z0-9_]{0,63}',n) for n in names) or names!=sorted(set(names)):raise ValueError('handler names')
    if any(type(h['description']) is not str or not isinstance(h['schema'],dict) for h in handlers):raise ValueError('handler definition')
    state=value['state'];checksum=value['state_sha256']
    if state is None:
        if checksum is not None:raise ValueError('null hash')
    elif not isinstance(state,dict) or type(checksum) is not str or not re.fullmatch('[0-9a-f]{64}',checksum) or digest(canonical(state).encode('utf-8'))!=checksum:raise ValueError('state hash')
    return value


def ordinary(value):
    if isinstance(value,Number):return int(value) if re.fullmatch(r'-?(0|[1-9][0-9]*)',value) else value
    if isinstance(value,list):return [ordinary(x) for x in value]
    if isinstance(value,dict):return {k:ordinary(v) for k,v in value.items()}
    return value


def inspect_created(store):
    checkpoint=(store/'checkpoint.json').read_bytes();value=envelope(checkpoint)
    if value['state'] is None:raise ValueError('writer omitted semantic snapshot')
    raw=(store/'events.log').read_bytes();lines=raw.splitlines()
    if not raw.endswith(b'\n') or lines[0]!=b'{"log_version":1}':raise ValueError('log framing')
    events=[parse(x) for x in lines[1:]]
    if not events or [uint(e['seq'],True) for e in events]!=list(range(1,len(events)+1)):raise ValueError('full-origin sequence')
    initial=events[0]
    if initial['type']!='session_initialized' or set(initial['session'])!={'session_id','identity'} or canonical(initial['session'])!=canonical({'session_id':value['session_id'],'identity':value['identity']}):raise ValueError('session initializer')
    if uint(value['as_of'])>uint(events[-1]['seq']):raise ValueError('checkpoint beyond log')
    if sum(e['type']=='session_initialized' for e in events)!=1 or (store/'origin.json').exists():raise ValueError('full-origin layout')
    return value,events


class Server(http.server.ThreadingHTTPServer):
    def __init__(self,vendor):
        super().__init__(('127.0.0.1',0),Handler);self.vendor=vendor;self.requests=[]
class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self,*_):pass
    def do_POST(self):
        self.server.requests.append(self.rfile.read(int(self.headers['Content-Length'])).decode())
        if len(self.server.requests)>1:self.send_error(429,'local fixture bound');return
        raw=json.dumps(response(self.server.vendor,[])).encode();self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(raw)));self.end_headers();self.wfile.write(raw)


def hashes(root):
    values={}
    if not root.exists():return values
    for p in sorted(root.rglob('*')):
        if p.is_symlink():values[str(p.relative_to(root))]='symlink:'+os.readlink(p)
        elif p.is_file():values[str(p.relative_to(root))]=digest(p.read_bytes())
    return values


class Runner:
    def __init__(self,binary,order):self.binary=binary;self.order=order
    def args(self,mode,store=None):
        selectors=['--session-dir',str(store)] if store is not None else []
        return [str(self.binary)]+(selectors+[mode] if self.order=='flags-first' else [mode]+selectors)
    def invoke(self,work,vendor,store=None,prompt=None,mode='protocol',extra=None,args=None):
        server=Server(vendor);thread=threading.Thread(target=lambda:server.serve_forever(poll_interval=.01));thread.start()
        try:
            env=environment(work,vendor,f'http://127.0.0.1:{server.server_port}');env.pop('CH02_LOG',None);env['LLM_API_KEY']=CANARY
            if extra:env.update(extra)
            command=args or self.args(mode,store);text=json.dumps({'user':prompt})+'\n' if prompt is not None else ''
            if mode=='chat':text='/session\n/checkpoint\n/quit\n'
            start=time.monotonic();r=subprocess.run(command,input=text,text=True,capture_output=True,cwd=work,env=env,timeout=12)
            return dict(argv=command,exit=r.returncode,stdout=r.stdout,stderr=r.stderr,requests=server.requests,elapsed=time.monotonic()-start,endpoint=env['LLM_BASE_URL'])
        finally:server.shutdown();server.server_close();thread.join()


def faults(checkpoint,log):
    value=envelope(checkpoint);out={}
    def outer(name,edit):v=copy.deepcopy(value);edit(v);out[name]=(canonical(v).encode(),log)
    outer('unsupported-version',lambda v:v.__setitem__('version',2))
    outer('unsupported-state-version',lambda v:v.__setitem__('state_version',2))
    outer('unknown-outer-field',lambda v:v.__setitem__('unexpected',True))
    outer('missing-required-field',lambda v:v.pop('state_version'))
    outer('invalid-session-id',lambda v:v.__setitem__('session_id','Z'*32))
    outer('numeric-session-id',lambda v:v.__setitem__('session_id',Number('1'*32)))
    outer('numeric-system',lambda v:v['identity'].__setitem__('system',Number('12')))
    outer('numeric-hash',lambda v:v.__setitem__('state_sha256',Number('1'*64)))
    outer('zero-anchor',lambda v:v.__setitem__('as_of',0))
    outer('overflow-anchor',lambda v:v.__setitem__('as_of',MAX+1))
    outer('negative-job-watermark',lambda v:v['high_watermarks'].__setitem__('job',-1))
    outer('anchor-event-disagreement',lambda v:v['high_watermarks'].__setitem__('event',uint(v['as_of'])+1))
    outer('state-hash-mismatch',lambda v:v.__setitem__('state_sha256','0'*64 if v['state_sha256']!='0'*64 else '1'*64))
    outer('null-state-with-hash',lambda v:v.__setitem__('state',None))
    outer('unknown-identity-field',lambda v:v['identity'].__setitem__('unexpected',True))
    outer('anchor-beyond-log',lambda v:(v.__setitem__('as_of',MAX),v['high_watermarks'].__setitem__('event',MAX)))
    out['duplicate-outer-member']=(b'{"version":1,'+checkpoint.lstrip()[1:],log)
    out['invalid-checkpoint-utf8']=(checkpoint+b'\xff',log)
    out['trailing-json']=(checkpoint+b'{}',log)
    out['partial-log-record']=(checkpoint,log+b'{"seq":')
    lines=log.splitlines();lines[1]=b'{"seq":1,'+lines[1][1:];out['duplicate-event-member']=(checkpoint,b'\n'.join(lines)+b'\n')
    return out


def protocol_records(result):
    try:values=[json.loads(x) for x in result['stdout'].splitlines()]
    except ValueError:return False
    return result['exit']==0 and bool(values) and all(isinstance(x,dict) for x in values)


def run_checks(runner):
    checks=[];runs={}
    def record(name,passed,evidence=None,blocked=None):
        checks.append(dict(id=name,passed=bool(passed),**({'blocked_by':blocked} if blocked else {})))
        if evidence is not None:runs[name]=evidence
    for vendor in MODELS:
      with tempfile.TemporaryDirectory(prefix='ch10-session-') as directory:
        work=Path(directory);legacy=work/'legacy';legacy.mkdir();old=runner.invoke(legacy,vendor,prompt='Legacy local control.',extra={'CH02_LOG':str(legacy/'standalone.log')})
        record(vendor+'/legacy-protocol-positive',protocol_records(old) and len(old['requests'])==1 and not (legacy/'.ensemble').exists(),old)
        store=work/'session';first=runner.invoke(work,vendor,store,'REMEMBER_ALPHA');valid=False;reason=''
        try:created,events=inspect_created(store);valid=protocol_records(first) and len(first['requests'])==1
        except (ValueError,KeyError,TypeError,FileNotFoundError,UnicodeError) as error:reason=str(error)
        record(vendor+'/genuine-session-create',valid,dict(run=first,validation_error=reason,files=hashes(store)))
        if not valid:
            for name in ['resume-dialogue-once','identity-sequence-request-index','safe-configuration-files','offline-inspect','null-state-rebuild','plain-system-presence','writer-lock','default-human-and-selectors']+[f'corrupt/{name}' for name in faults_fixture_names()]:record(vendor+'/'+name,False,blocked='genuine-session-create')
            continue
        initial_log=(store/'events.log').read_bytes();session_id=created['session_id'];second=runner.invoke(work,vendor,store,'REMEMBER_BETA');updated,later=inspect_created(store)
        record(vendor+'/resume-dialogue-once',protocol_records(second) and len(second['requests'])==1 and second['requests'][0].count('REMEMBER_ALPHA')==1 and second['requests'][0].count('REMEMBER_BETA')==1,second)
        turns=[e['turn'] for e in later if e['type']=='turn_started'];indices=[uint(t['request_index'],True) for t in turns];ids=[t['request_id'] for t in turns]
        record(vendor+'/identity-sequence-request-index',updated['session_id']==session_id and (store/'events.log').read_bytes().startswith(initial_log) and len(turns)==2 and indices==sorted(set(indices)) and len(set(ids))==2 and uint(updated['high_watermarks']['request'])>=max(indices),dict(indices=indices,ids=ids))
        raw=b''.join(p.read_bytes() for p in store.iterdir() if p.is_file());record(vendor+'/safe-configuration-files',all(value.encode() not in raw for value in [CANARY,first['endpoint'],second['endpoint']]))
        before=hashes(store);inspection=runner.invoke(work,vendor,args=[str(runner.binary),'session','inspect',str(store)])
        inspected=inspection['exit']==0 and not inspection['requests'] and str(session_id) in inspection['stdout'] and hashes(store)==before;record(vendor+'/offline-inspect',inspected,inspection)
        seed=work/'seed';shutil.copytree(store,seed)
        rebuilt=work/'rebuilt';normal=work/'normal';shutil.copytree(seed,rebuilt);shutil.copytree(seed,normal);v=envelope((rebuilt/'checkpoint.json').read_bytes());v['state']=None;v['state_sha256']=None;(rebuilt/'checkpoint.json').write_text(canonical(v))
        a=runner.invoke(work,vendor,normal,'COMPARE_NEXT');b=runner.invoke(work,vendor,rebuilt,'COMPARE_NEXT');record(vendor+'/null-state-rebuild',a['exit']==b['exit']==0 and len(a['requests'])==len(b['requests'])==1 and a['requests'][0]==b['requests'][0],dict(snapshot=a,rebuild=b))
        presence=[]
        for label,value,ok in [('equal',updated['identity']['system'],True),('empty','',False),('different','Different base',False)]:
            target=work/('presence-'+label);shutil.copytree(seed,target);before=hashes(target);r=runner.invoke(work,vendor,target,extra={'LLM_SYSTEM':value});presence.append(dict(label=label,run=r,passed=(r['exit']==0 if ok else r['exit']!=0 and 'session_incompatible' in r['stderr'] and hashes(target)==before) and not r['requests']))
        record(vendor+'/plain-system-presence',all(x['passed'] for x in presence),presence)
        for name,(checkpoint,log) in faults((seed/'checkpoint.json').read_bytes(),(seed/'events.log').read_bytes()).items():
            target=work/'mutated';shutil.copytree(seed,target);(target/'checkpoint.json').write_bytes(checkpoint);(target/'events.log').write_bytes(log);before=hashes(target)
            inspect=runner.invoke(work,vendor,args=[str(runner.binary),'session','inspect',str(target)]);live=runner.invoke(work,vendor,target)
            passed=inspected and all(r['exit']!=0 and 'session_corrupt' in r['stderr'] and not r['requests'] for r in [inspect,live]) and before==hashes(target)
            record(vendor+'/corrupt/'+name,passed,dict(inspect=inspect,live=live,unchanged=before==hashes(target)),None if inspected else 'offline-inspect');shutil.rmtree(target)
        record(vendor+'/writer-lock',*lock_check(runner,work,vendor,seed))
        record(vendor+'/default-human-and-selectors',*selection_check(runner,work,vendor))
    return checks,runs


def faults_fixture_names():return list(faults(canonical(sample()).encode(),b'{"log_version":1}\n{"seq":1,"type":"fixture"}\n'))


def lock_check(runner,work,vendor,seed):
    target=work/'locked';shutil.copytree(seed,target);env=environment(work,vendor,'http://127.0.0.1:1');env.pop('CH02_LOG',None);env['LLM_API_KEY']=CANARY
    process=subprocess.Popen(runner.args('protocol',target),stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,cwd=work,env=env,text=True)
    rows=[]
    try:
        deadline=time.monotonic()+3;held=False
        with (target/'owner.lock').open('rb') as lock:
            while time.monotonic()<deadline and process.poll() is None:
                try:fcntl.flock(lock,fcntl.LOCK_EX|fcntl.LOCK_NB);fcntl.flock(lock,fcntl.LOCK_UN)
                except BlockingIOError:held=True;break
                time.sleep(.01)
        if not held:return False,dict(error='live holder did not acquire lock')
        inode=(target/'owner.lock').stat().st_ino;before=hashes(target)
        for args in [None,[str(runner.binary),'session','inspect',str(target)]]:
            r=runner.invoke(work,vendor,target,args=args);rows.append(r)
        refused=all(r['exit']!=0 and 'session_in_use' in r['stderr'] and r['elapsed']<2 and not r['requests'] for r in rows) and hashes(target)==before
        process.send_signal(signal.SIGTERM);stdout,stderr=process.communicate(timeout=5);envelope((target/'checkpoint.json').read_bytes())
        reopened=runner.invoke(work,vendor,target);same_inode=(target/'owner.lock').stat().st_ino==inode
        return refused and reopened['exit']==0 and not reopened['requests'] and same_inode,dict(contenders=rows,holder_exit=process.returncode,holder_stdout=stdout,holder_stderr=stderr,reopened=reopened,same_lock_inode=same_inode)
    finally:
        if process.poll() is None:process.kill();process.communicate(timeout=3)


def selection_check(runner,work,vendor):
    human=work/'human';human.mkdir();r=runner.invoke(human,vendor,mode='chat');default=human/'.ensemble/session';rows=[r]
    try:v,_=inspect_created(default);okay=r['exit']==0 and not r['requests'] and v['session_id'] in r['stdout']
    except (ValueError,KeyError,FileNotFoundError):okay=False
    for label,store,extra in [('both',work/'conflict',{'CH02_LOG':str(work/'legacy-explicit.log')}),('blank','',{})]:
        r=runner.invoke(work,vendor,store,extra=extra);rows.append(r);okay &= r['exit']!=0 and not r['requests']
    okay &= not (work/'conflict').exists() and not (work/'legacy-explicit.log').exists()
    return okay,rows


def sample():
    # Only an outer-format oracle positive. This is not a semantic codec or a
    # fabricated runtime snapshot accepted by a student implementation.
    state={'fixture':'outer-oracle only','number':Number('9007199254740993')}
    return dict(version=1,session_id='a'*32,identity=dict(mode='plain',system='Identity.',skills=None,handlers=[]),as_of=1,high_watermarks=dict(event=1,request=0,activation=0,job=0),state_version=1,state=state,state_sha256=digest(canonical(state).encode()))


def self_test():
    examples={'1':'1','1.0':'1','10e-1':'1','1.000e+3':'1e3','12.30':'123e-1','9007199254740993':'9007199254740993','18446744073709551615.0':'18446744073709551615','-0.00e99':'0','1e1000000':'1e1000000'}
    for before,after in examples.items():assert number(Number(before))==after
    assert number(Number('10e'+('9'*5000)))=='1e1'+('0'*5000)  # No binary64 or host integer digit ceiling.
    assert canonical(parse('{"é":"<>&/\\u0001","a":1e0}'))=='{"a":1,"é":"<>&/\\u0001"}'
    valid=canonical(sample()).encode();envelope(valid);rejected=[]
    for name,(bad,_) in faults(valid,b'{"log_version":1}\n{"seq":1,"type":"fixture"}\n').items():
        if name in ('anchor-beyond-log','partial-log-record','duplicate-event-member'):continue
        try:envelope(bad)
        except (ValueError,TypeError,KeyError,UnicodeError):rejected.append(name)
        else:raise AssertionError('outer mutation accepted: '+name)
    return dict(scope='Outer/canonical checker controls only, no runtime positive',canonical_examples=len(examples)+2,outer_positive=1,intended_refusals=rejected)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',nargs='?',type=Path);p.add_argument('--self-test',action='store_true');p.add_argument('--receipt',type=Path);p.add_argument('--argument-order',choices=['flags-first','mode-first'],default='flags-first');args=p.parse_args()
    inputs=[Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-10.md']
    if args.self_test:result=self_test();success=True
    else:
        if not args.binary:p.error('binary or --self-test required')
        binary=args.binary.resolve(strict=True);inputs.append(binary);identities={str(path):digest(path.read_bytes()) for path in inputs};checks,runs=run_checks(Runner(binary,args.argument_order));assert identities=={str(path):digest(path.read_bytes()) for path in inputs},'input changed during run'
        success=all(x['passed'] for x in checks);result=dict(passed=sum(x['passed'] for x in checks),total=len(checks),checks=checks,runs=runs,argument_order=args.argument_order,limits='Initial CLI/plain-session subset. No private codec completeness, public import, Skills compatibility, nonempty-tail differential, fault injection, full size limits, GUI or live acceptance.')
    result['input_sha256']={str(path):digest(path.read_bytes()) for path in inputs}
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='runs'},indent=2));return int(not success)


if __name__=='__main__':raise SystemExit(main())
