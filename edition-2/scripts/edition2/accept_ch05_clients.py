#!/usr/bin/env python3
"""Barrier-controlled protocol and real PTY checks using local provider fixtures.

These checks exercise actual CLI input/output. They do not count as real-model
live evidence. They use the published protocol fields and semantic human labels;
request IDs are discovered, never assumed.
"""
import argparse
import http.server
import json
import os
import pathlib
import pty
import queue
import re
import select
import signal
import subprocess
import tempfile
import termios
import threading
import time

from accept_ch03 import response, call, MODEL, VENDORS


class Server:
    def __init__(self):
        self.calls=queue.Queue();self.stopped=threading.Event()
        outer=self
        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self,*args):pass
            def do_POST(self):
                body=self.rfile.read(int(self.headers['Content-Length']))
                release=queue.Queue();outer.calls.put((body,release))
                while not outer.stopped.is_set():
                    try:value=release.get(timeout=.1);break
                    except queue.Empty:continue
                else:return
                try:
                    data=json.dumps(value).encode();self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data)
                except (BrokenPipeError,ConnectionResetError):pass
        self.server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Handler)
        self.thread=threading.Thread(target=self.server.serve_forever,daemon=True);self.thread.start()
    def get(self):return self.calls.get(timeout=5)
    def close(self):self.stopped.set();self.server.shutdown();self.server.server_close();self.thread.join()


class Client:
    def __init__(self,binary,root,server,vendor,human=False):
        env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','CH02_','COURSE_','ANTHROPIC_','OPENAI_','GEMINI_'))}
        env.update(LLM_VENDOR=vendor,LLM_MODEL=MODEL,LLM_RESOLVED_MODEL=MODEL,LLM_API_KEY='LOCAL-ONLY',LLM_BASE_URL='http://127.0.0.1:'+str(server.server.server_port),CH02_LOG=str(root/'events.jsonl'))
        self.human=human;self.output='';self.condition=threading.Condition();self.lines=[];self.err=tempfile.TemporaryFile()
        if human:
            self.master,slave=pty.openpty();attrs=termios.tcgetattr(slave);attrs[3]&=~termios.ECHO;termios.tcsetattr(slave,termios.TCSANOW,attrs)
            self.process=subprocess.Popen([str(binary),'chat'],cwd=root,env=env,stdin=slave,stdout=slave,stderr=self.err,start_new_session=True);os.close(slave)
        else:
            self.process=subprocess.Popen([str(binary),'protocol'],cwd=root,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=self.err,text=True,start_new_session=True)
            def read():
                for line in self.process.stdout:
                    with self.condition:self.lines.append(json.loads(line));self.condition.notify_all()
            self.reader=threading.Thread(target=read,daemon=True);self.reader.start()
    def send(self,value):
        if self.human:os.write(self.master,value.encode())
        else:self.process.stdin.write(json.dumps(value)+'\n');self.process.stdin.flush()
    def record(self,predicate):
        deadline=time.monotonic()+5
        with self.condition:
            while True:
                for value in self.lines:
                    if predicate(value):return value
                remaining=deadline-time.monotonic()
                if remaining<=0:raise AssertionError('protocol barrier timed out')
                self.condition.wait(remaining)
    def text(self,pattern):
        deadline=time.monotonic()+5
        while True:
            match=re.search(pattern,self.output,re.I|re.S)
            if match:return match
            if time.monotonic()>=deadline:raise AssertionError('PTY boundary missing: '+pattern+'; output='+self.output)
            if select.select([self.master],[],[],.05)[0]:
                try:self.output+=os.read(self.master,65536).decode()
                except OSError:raise AssertionError('PTY exited before '+pattern)
    def close(self):
        if self.process.poll() is None:os.killpg(self.process.pid,signal.SIGKILL)
        self.process.wait(timeout=5)
        if self.human:os.close(self.master)
        else:
            if not self.process.stdin.closed:self.process.stdin.close()
            self.process.stdout.close();self.reader.join(timeout=1)
        self.err.close()


def protocol(binary,vendor):
    with tempfile.TemporaryDirectory(prefix='ch05-protocol-') as tmp:
        root=pathlib.Path(tmp);server=Server();c=Client(binary,root,server,vendor)
        try:
            for invalid in ({'kind':'hint','text':'PRIVATE-IDLE'}, {'kind':'unknown'}, {'kind':'prompt','text':''}, {'kind':'interrupt','extra':'PRIVATE-EXTRA'}):
                before=sum('error' in x for x in c.lines);c.send(invalid)
                deadline=time.monotonic()+5
                with c.condition:
                    while sum('error' in x for x in c.lines)==before:
                        left=deadline-time.monotonic();assert left>0,'invalid control failed to acknowledge';c.condition.wait(left)
                assert c.lines[-1]['error']['code']=='invalid_control'
            assert server.calls.empty(),'invalid control contacted provider'
            c.send({'kind':'interrupt'});idle=c.record(lambda x:x.get('ack')=='interrupt');assert idle['request_id']=='' and idle['interrupted'] is False
            c.send({'kind':'prompt','text':'FIRST-PROMPT'});first=c.record(lambda x:x.get('accepted')=='prompt')['request_id'];r1,release1=server.get()
            c.send({'kind':'prompt','text':'SECOND-PROMPT'});second=c.record(lambda x:x.get('accepted')=='prompt' and x['request_id']!=first)['request_id']
            c.send({'kind':'hint','text':'PENDING-HINT'});ack=c.record(lambda x:x.get('ack')=='hint');assert ack['request_id']==first and ack['seq']>0 and ack['sent'] is False
            assert b'PENDING-HINT' not in r1
            c.send({'kind':'interrupt'});ack=c.record(lambda x:x.get('ack')=='interrupt' and x.get('interrupted'));assert ack['request_id']==first
            interrupted=c.record(lambda x:x.get('completion',{}).get('request_id')==first)['completion'];assert interrupted['outcome']=='interrupted' and interrupted['error']['code']=='interrupted'
            r2,release2=server.get();assert b'SECOND-PROMPT' in r2 and b'PENDING-HINT' in r2
            release1.put(response(vendor,[],'STALE'));release2.put(response(vendor,[],'SECOND-ANSWER'))
            done=c.record(lambda x:x.get('completion',{}).get('request_id')==second)['completion'];assert done['outcome']=='success' and done['text']=='SECOND-ANSWER' and done['pending_hints']==0
            c.send({'user':'LEGACY-THIRD'});_,release3=server.get();release3.put(response(vendor,[],'LEGACY-ANSWER'));assert c.record(lambda x:'assistant' in x)=={'assistant':'LEGACY-ANSWER'}
            c.process.stdin.close();assert c.process.wait(timeout=5)==0;c.reader.join(timeout=1)
            assert len([x for x in c.lines if 'completion' in x])==2 and len([x for x in c.lines if 'accepted' in x])==2
            assert 'usage' in c.lines[-1] and not any('observation' in x for x in c.lines)
            for rid in (first,second):assert next(i for i,x in enumerate(c.lines) if x.get('request_id')==rid and 'accepted' in x)<next(i for i,x in enumerate(c.lines) if x.get('completion',{}).get('request_id')==rid)
            assert 'PRIVATE-' not in json.dumps(c.lines),'control error leaked input'
        finally:c.close();server.close()


def human(binary,vendor):
    with tempfile.TemporaryDirectory(prefix='ch05-pty-') as tmp:
        root=pathlib.Path(tmp);server=Server();c=Client(binary,root,server,vendor,True)
        try:
            c.text(r'You>');c.send('/help\n');c.text(r'/hint.*?/interrupt')
            c.send('FIRST-HUMAN\n');first=c.text(r'Accepted\s+(\S+)\.').group(1);r1,release1=server.get()
            c.send('SECOND-HUMAN\n');second=c.text(r'Accepted\s+'+re.escape(first)+r'\..*?Accepted\s+(\S+)\.').group(1);assert first!=second
            c.send('/hint HUMAN-GUIDANCE\n');c.text(r'Hint received.*?'+re.escape(first)+r'.*?sent=false')
            c.send('/ephemeral BUSY-REFUSAL\n');c.text(r'Command refused: Agent busy')
            c.send('/interrupt\n');c.text(r'Interrupt:.*?'+re.escape(first)+r'.*?true');c.text(r'Request\s+'+re.escape(first)+r'\s+\(interrupted')
            r2,release2=server.get();assert b'HUMAN-GUIDANCE' in r2 and b'SECOND-HUMAN' in r2 and b'BUSY-REFUSAL' not in r2
            release1.put(response(vendor,[],'STALE'));release2.put(response(vendor,[],'READABLE-ANSWER\nSECOND-LINE'))
            c.text(r'Request\s+'+re.escape(second)+r'\s+\(success.*?Assistant:\s+READABLE-ANSWER\s+SECOND-LINE')
            c.send('/quit\n');assert c.process.wait(timeout=5)==0
        finally:c.close();server.close()


def workflow(binary,vendor):
    with tempfile.TemporaryDirectory(prefix='ch05-workflow-') as tmp:
        root=pathlib.Path(tmp);server=Server()
        env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ENSEMBLE_','ANTHROPIC_','OPENAI_','GEMINI_'))}
        env.update(LLM_VENDOR=vendor,LLM_MODEL=MODEL,LLM_RESOLVED_MODEL=MODEL,LLM_API_KEY='LOCAL-ONLY',LLM_BASE_URL='http://127.0.0.1:'+str(server.server.server_port),ENSEMBLE_RUN_DIRECTORY=str(root))
        process=subprocess.Popen([str(binary),'workflow'],cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
        try:
            requests=[]
            steps=[
                [call('author-write','write_file',path='draft.txt',content='ORIGINAL-DRAFT')],
                [],
                [call('editor-read','read_file',path='draft.txt'),call('editor-edit','edit_file',path='draft.txt',old_text='ORIGINAL-DRAFT',new_text='EDITED-DRAFT')],
                [],
                [call('reviewer-read','read_file',path='draft.txt')],
                [],
            ]
            finals=['AUTHOR-HANDOFF','EDITOR-HANDOFF','ACTUAL-REVIEWER-RESULT']
            for i,calls in enumerate(steps):
                body,release=server.get();requests.append(json.loads(body))
                if i==2:assert b'AUTHOR-HANDOFF' in body,'editor did not receive completed author output'
                if i==4:assert b'EDITOR-HANDOFF' in body,'reviewer did not receive completed editor output'
                release.put(response(vendor,calls,'WORKING' if calls else finals[i//2]))
            stdout,stderr=process.communicate(timeout=8)
            assert process.returncode==0,'workflow failed: '+stderr
            assert 'ACTUAL-REVIEWER-RESULT' in stdout,'reviewer result was not the model completion'
            assert (root/'workspace/draft.txt').read_text()=='EDITED-DRAFT','actual file workflow did not run'
            def names(body):
                if vendor=='gemini':return {x['name'] for group in body['tools'] for x in group.get('functionDeclarations',[])}
                return {x['function']['name'] if vendor=='openai' else x['name'] for x in body['tools']}
            assert names(requests[0])=={'read_file','write_file'}
            assert names(requests[2])=={'read_file','write_file','edit_file'}
            assert names(requests[4])=={'read_file'}
            logs=[root/(role+'.log') for role in ('author','editor','reviewer')]
            for path in logs:
                events=[json.loads(line) for line in path.read_text().splitlines()][1:]
                assert sum(e['type']=='turn_started' for e in events)==1 and sum(e['type']=='turn_ended' for e in events)==1
        finally:
            if process.poll() is None:os.killpg(process.pid,signal.SIGKILL)
            process.wait();process.stdout.close();process.stderr.close();server.close()


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=pathlib.Path);p.add_argument('--vendor',action='append',choices=VENDORS);p.add_argument('--workflow',action='store_true');a=p.parse_args();checks=[]
    for vendor in a.vendor or VENDORS:
        for name,run in ([('workflow',workflow)] if a.workflow else [('protocol',protocol),('human-pty',human)]):
            try:run(a.binary.resolve(strict=True),vendor);errors=[]
            except Exception as exc:errors=[type(exc).__name__+': '+str(exc)]
            checks.append(dict(id=vendor+'/'+name,passed=not errors,details=errors))
    result=dict(passed=all(c['passed'] for c in checks),checks=checks);print(json.dumps(result,indent=2));return 0 if result['passed'] else 1

if __name__=='__main__':raise SystemExit(main())
