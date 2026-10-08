#!/usr/bin/env python3
"""Strict skill call arguments and one-shot limit consumption, actual CLI/wire.

Duplicate JSON member bytes are emitted on each adapter's real argument surface.
No model endpoint, credentials or implementation imports are used.
"""
import argparse
import http.server
import json
from pathlib import Path
import subprocess
import tempfile
import threading

from accept_ch09 import ROOT, MODELS, call, definition, environment, response, sha, dialogue


INVALID=[('missing','{}',''),('blank','{"name":""}',''),('null','{"name":null}',''),
         ('number','{"name":1}',''),('boolean','{"name":false}',''),('array','{"name":[]}',''),
         ('invalid-id','{"name":"EDIT"}',''),('unknown','{"name":"edit","unknown":true}','edit'),
         ('duplicate','{"name":"edit","name":"edit"}','edit'),
         ('limits','{"name":"edit","max_output_bytes":1}','edit')]
ATTEMPTS=[(name,'load_skill',raw,dict(error='invalid_skill_arguments',name=valid,revision=0)) for name,raw,valid in INVALID]+[
    ('loaded','load_skill','{"name":"edit"}',dict(status='loaded',name='edit',revision=1,changed=True)),
    ('unchanged','load_skill','{"name":"edit"}',dict(status='unchanged',name='edit',revision=1,changed=False)),
    ('unavailable','load_skill','{"name":"hidden"}',dict(error='skill_unavailable',name='hidden',revision=1)),
    ('unloaded','unload_skill','{"name":"edit"}',dict(status='unloaded',name='edit',revision=2,changed=True))]
MANUAL='Long manual survives one-byte consumed report budget.\n'*8
SENTINEL='ordinary-read-must-not-inherit-one-byte-report-budget-'*3


def wire(vendor):
    calls=[]
    for identity,tool,raw,_ in ATTEMPTS:
        calls.extend([call('setter-'+identity,'tool_limits',max_output_bytes=1),
                      call(identity,tool,__ch09_raw_arguments__=identity),
                      call('read-'+identity,'read_file',path='probe.txt')])
    body=json.dumps(response(vendor,calls))
    for identity,_,raw,_ in ATTEMPTS:
        marker=json.dumps(dict(__ch09_raw_arguments__=identity))
        old=json.dumps(marker) if vendor=='openai' else marker
        new=json.dumps(raw) if vendor=='openai' else raw
        assert body.count(old)==1,(vendor,identity)
        body=body.replace(old,new,1)
    return body.encode()


class Fixture(http.server.ThreadingHTTPServer):
    def __init__(self,vendor):
        super().__init__(('127.0.0.1',0),Handler);self.vendor=vendor;self.requests=[];self.responses=[]


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self,*_):pass
    def do_POST(self):
        owner=self.server;owner.requests.append(json.loads(self.rfile.read(int(self.headers['Content-Length']))))
        if len(owner.requests)>2:self.send_error(429,'fixture request bound');return
        body=wire(owner.vendor) if len(owner.requests)==1 else json.dumps(response(owner.vendor,[])).encode()
        owner.responses.append(body.decode());self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)


def evaluate(events,requests,artifacts,vendor):
    returned={e['tool']['call_id']:e['tool'] for e in events if e.get('type')=='tool_returned'}
    checks={}
    for identity,tool,_,want in ATTEMPTS:
        record=returned.get(identity,{});text=''.join(p.get('text','') for p in record.get('parts',[]));lines=text.splitlines()
        try:ack=json.loads(lines[-1])
        except (ValueError,IndexError):ack=None
        checks[identity+'/strict-ack']=ack==want and bool(record.get('is_error',False))==('error' in want) and lines[-1:]==[json.dumps(want,separators=(',',':'))]
        checks[identity+'/consumed-limit-note']=len(lines)>=2 and 'tool_limits' in lines[0] and tool in lines[0]
        ordinary=returned.get('read-'+identity,{})
        text=''.join(p.get('text','') for p in ordinary.get('parts',[]))
        checks[identity+'/ordinary-defaults']=not ordinary.get('is_error',False) and SENTINEL in text and 'tool_limits' not in text
    transitions=[e['skills'] for e in events if e.get('type')=='skills_changed']
    checks['only-two-changed-transitions']=len(transitions)==2 and [(x['action'],x['name'],x['state']['revision']) for x in transitions]==[('load','edit',1),('unload','edit',2)]
    records=[e['tool'] for e in events if e.get('type') in ('tool_called','tool_returned') and e['tool']['call_id'] in {x[0] for x in ATTEMPTS}]
    jobs=[returned.get('read-'+identity,{}).get('job',{}) for identity,_,_,_ in ATTEMPTS]
    allowed=[job.get('output',{}).get('locator') for job in jobs]
    checks['management-paired-without-jobs']=len(records)==2*len(ATTEMPTS) and all(not t.get('job') for t in records) and sorted(artifacts)==sorted(allowed) and [job.get('handle') for job in jobs]==list(range(1,len(ATTEMPTS)+1))
    envelope='[skill edit activation 2]\n'+MANUAL+'\n[/skill]'
    checks['consumed-budget-keeps-manual']=len(requests)==2 and sum(text.count(envelope) for kind,text in dialogue(vendor,requests[-1]) if kind=='text')==1
    checks['bounded-complete-transport']=len(requests)==2 and all('setter-'+identity in returned for identity,_,_,_ in ATTEMPTS)
    return checks


def capture(binary,vendor):
    server=Fixture(vendor);thread=threading.Thread(target=server.serve_forever);thread.start()
    try:
        with tempfile.TemporaryDirectory(prefix='ch09-management-') as directory:
            work=Path(directory);(work/'probe.txt').write_text(SENTINEL)
            catalog={'base':definition('base','primary','Management fixture.','read_file tool_limits','edit'),
                     'edit':definition('edit','loadable',MANUAL,'write_file'),
                     'hidden':definition('hidden','loadable','Hidden.')}
            for name,data in catalog.items():
                path=work/'catalog'/name/'SKILL.md';path.parent.mkdir(parents=True);path.write_bytes(data)
            env=environment(work,vendor,f'http://127.0.0.1:{server.server_port}');env.update(LLM_SKILLS_DIR='catalog',LLM_PRIMARY_SKILL='base')
            process=subprocess.run([str(binary),'protocol'],input='{"user":"Exercise exact local management arguments."}\n',text=True,capture_output=True,env=env,cwd=work,timeout=30)
            path=work/'session.jsonl';events=[json.loads(x) for x in path.read_text().splitlines()][1:] if path.exists() else []
            artifacts=[str(p.relative_to(work)) for p in (work/'cr'/'io').glob('*')]
            return dict(exit=process.returncode,stdout=process.stdout,stderr=process.stderr,requests=server.requests,response_wire=server.responses,events=events,artifacts=artifacts,checks=evaluate(events,server.requests,artifacts,vendor))
    finally:server.shutdown();server.server_close();thread.join()


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);p.add_argument('--receipt',type=Path);args=p.parse_args();binary=args.binary.resolve()
    identities={str(path):sha(path.read_bytes()) for path in [Path(__file__),ROOT/'scripts/edition2/accept_ch09.py',ROOT/'book/edition-2/chapter-09.md',ROOT/'book/edition-2/chapter-04.md',binary]}
    runs={vendor:capture(binary,vendor) for vendor in MODELS};checks=[dict(id=vendor+'/'+name,passed=passed) for vendor,r in runs.items() for name,passed in r['checks'].items()]
    assert identities=={path:sha(Path(path).read_bytes()) for path in identities},'input changed'
    result=dict(passed=sum(x['passed'] for x in checks),total=len(checks),checks=checks,runs=runs,input_sha256=identities,limits='Model-issued strict arguments and next-attempt limit consumption; public controls, pause/interruption and persistence failures are separate checks.')
    if args.receipt:args.receipt.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:v for k,v in result.items() if k!='runs'},indent=2));return int(result['passed']!=result['total'])


if __name__=='__main__':raise SystemExit(main())
