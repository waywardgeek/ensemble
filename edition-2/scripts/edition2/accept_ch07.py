#!/usr/bin/env python3
"""Partial Chapter 7 black-box local transport gate, derived before student code.

Usage: python3 scripts/edition2/accept_ch07.py /absolute/path/to/ensemble-gui
Requires Python websocket-client. No model request is expected or authorized.
This initial scope is not browser, watch-race, tool-admission or speech acceptance.
"""
import argparse
import contextlib
import hashlib
import http.server
import json
import os
from pathlib import Path
import queue
import re
import signal
import subprocess
import tempfile
import threading
import urllib.parse
import urllib.request

import websocket


class Socket:
    def __init__(self, url, **options):
        self.url = url
        self.ws = websocket.create_connection(url.replace('http://', 'ws://')+'/ws', timeout=4,
                                              origin=url, http_proxy_host=None, **options)
        self.records = []

    def send(self, value):
        self.ws.send(json.dumps(value, separators=(',', ':')))

    def next(self):
        value = self.ws.recv()
        assert isinstance(value, str) and value, 'connection closed or emitted binary application data'
        value = json.loads(value)
        assert isinstance(value, dict), 'application message is not an object'
        self.records.append(value)
        return value

    def until(self, predicate):
        for _ in range(1000):
            value = self.next()
            if predicate(value):
                return value
        raise AssertionError('expected correlated record absent')

    def close(self):
        self.ws.close(timeout=1)


def check_empty_snapshot(records, command_id):
    assert len(records) == 2, 'empty snapshot contains unexpected events or live records'
    begin, end = records
    assert begin.get('type') == 'snapshot_begin' and begin.get('id') == command_id, 'snapshot start/correlation missing'
    generation = begin.get('generation')
    assert isinstance(generation, str) and generation, 'generation missing'
    assert isinstance(begin.get('agent_id'), str) and begin['agent_id'], 'Agent identity missing'
    assert end == dict(type='snapshot_end', generation=generation, watermark=begin['watermark']), 'snapshot end differs from cut'
    assert type(begin['watermark']) is int and begin['watermark'] >= 0, 'invalid revision watermark'
    assert begin['first_seq'] is None and begin['last_seq'] is None and begin['omitted'] == 0, 'empty history range is not null'
    assert type(begin['log_seq']) is int and begin['log_seq'] >= 0
    state = begin['state']
    assert state['lifecycle'] == 'idle' and state['active_request_id'] is None and state['active_operation'] is None
    assert state['queued_request_ids'] == [] and state['usage'] == []
    assert state['paused'] is False and state['typing_clients'] == state['speaking_clients'] == 0
    assert state['model'] == 'fixture-ch07-independent'
    assert not ({'api_key','headers','config','authorization','base_url'} & set(state)), 'unsafe state configuration fields'
    assert 'LOCAL-FIXTURE-NOT-A-SECRET' not in json.dumps(records), 'credential leaked into snapshot'
    return begin


def subscribe(client, command_id='subscribe'):
    client.send(dict(type='subscribe', id=command_id))
    records = []
    while True:
        value = client.next()
        records.append(value)
        if value.get('type') == 'snapshot_end':
            return check_empty_snapshot(records, command_id)


def check_pause_ack(records, command_id, typing, speaking):
    ack = next((x for x in records if x.get('type') == 'ack' and x.get('id') == command_id), None)
    assert ack is not None, 'pause acknowledgement missing'
    assert (ack['paused'], ack['typing_clients'], ack['speaking_clients']) == (bool(typing or speaking), typing, speaking), 'independent pause causes lost'
    observed = next((x for x in records if x.get('type') == 'observation' and x.get('revision') == ack['revision']), None)
    assert observed is not None, 'applied pause observation/revision missing'
    payload = observed['observation']
    assert payload['kind'] == 'pause_changed'
    assert (payload['paused'],payload['typing_clients'],payload['speaking_clients']) == (bool(typing or speaking),typing,speaking)
    return ack


def pause(client, command_id, typing, speaking, expected_typing, expected_speaking):
    start = len(client.records)
    client.send(dict(type='pause', id=command_id, typing=typing, speaking=speaking))
    ack = client.until(lambda x:x.get('type') == 'ack' and x.get('id') == command_id)
    if not any(x.get('type') == 'observation' and x.get('revision') == ack['revision']
               for x in client.records[start:]):
        client.until(lambda x:x.get('type') == 'observation' and x.get('revision') == ack['revision'])
    return check_pause_ack(client.records[start:], command_id, expected_typing, expected_speaking)


def cases(url):
    def snapshot():
        with contextlib.closing(Socket(url)) as client:
            subscribe(client)
    yield 'empty-owned-snapshot-wire', snapshot

    def two_clients():
        with contextlib.closing(Socket(url)) as a, contextlib.closing(Socket(url)) as b:
            first,second=subscribe(a),subscribe(b)
            assert first['generation'] != second['generation'], 'connections reused generation'
            one=pause(a,'a-typing',True,False,1,0)
            two=pause(b,'b-speaking',False,True,1,1)
            assert two['revision']>one['revision']
            three=pause(b,'b-clear',False,False,1,0)
            assert three['revision']>two['revision']
            a.close()
            value=b.until(lambda x:x.get('type')=='observation' and x.get('observation',{}).get('kind')=='pause_changed' and x['observation'].get('paused') is False)
            assert value['revision']>three['revision'] and value['observation']['typing_clients']==value['observation']['speaking_clients']==0, 'disconnect retained pause cause'
    yield 'two-client-pause-and-disconnect', two_clients

    def correlated_refusal():
        with contextlib.closing(Socket(url)) as client:
            subscribe(client)
            client.send(dict(type='hint', id='idle-hint', text='no active request'))
            error=client.until(lambda x:x.get('type')=='error')
            assert error.get('id')=='idle-hint' and error.get('code') and error.get('message'), 'well-formed refusal lost correlation'
            pause(client,'still-usable',True,False,1,0)
            pause(client,'clear-after-refusal',False,False,0,0)
    yield 'agent-refusal-keeps-connection', correlated_refusal

    for label,origin,host in [('missing',None,None),('null','null',None),('foreign','http://invalid.example',None),('foreign-host',url,'invalid.example')]:
        def rejected(origin=origin,host=host):
            opts=dict(timeout=3,http_proxy_host=None)
            if origin is None:opts['suppress_origin']=True
            else:opts['origin']=origin
            if host:opts['host']=host
            try:
                client=websocket.create_connection(url.replace('http://','ws://')+'/ws',**opts)
            except websocket.WebSocketBadStatusException as exc:
                assert 400<=exc.status_code<500, 'upgrade refusal not a local client error'
                return
            client.close()
            raise AssertionError('foreign/missing local origin or Host upgraded')
        yield 'origin-'+label,rejected

    exact_id='x'*(65536-len(json.dumps(dict(type='subscribe',id=''),separators=(',', ':')).encode()))
    def exact_command():
        with contextlib.closing(Socket(url)) as client:
            subscribe(client,exact_id)
    yield 'exact-command-size',exact_command
    oversized=json.dumps(dict(type='subscribe',id=exact_id+'x'),separators=(',', ':')).encode()
    assert len(oversized)==65537
    for label,command in [
        ('unknown-field',dict(type='pause',id='bad',typing=True,speaking=False,extra=1)),
        ('bad-known-type',dict(type='pause',id='bad',typing='yes',speaking=False)),
        ('missing-required-field',dict(type='pause',id='bad',typing=True)),
        ('unknown-command',dict(type='not-a-command',id='bad')),
    ]:
        def correctable(command=command):
            with contextlib.closing(Socket(url)) as client:
                subscribe(client)
                client.send(command)
                error=client.until(lambda x:x.get('type')=='error')
                assert error.get('id')=='bad' and error.get('code')=='invalid_command', 'correctable command lost refusal correlation'
                pause(client,'recovery',True,False,1,0)
                pause(client,'clear-recovery',False,False,0,0)
        yield 'correctable-'+label,correctable

    def pre_subscription():
        with contextlib.closing(Socket(url)) as client:
            client.send(dict(type='prompt',id='before',text='must not contact model'))
            error=client.until(lambda x:x.get('type')=='error')
            assert error.get('id')=='before', 'pre-subscription refusal lost correlation'
            subscribe(client)
            client.send(dict(type='pause',id='subscribe',typing=True,speaking=False))
            error=client.until(lambda x:x.get('type')=='error')
            assert error.get('id')=='subscribe', 'accepted command ID reuse was not refused'
            pause(client,'fresh-id',True,False,1,0)
            pause(client,'clear-id',False,False,0,0)
    yield 'subscription-and-command-id-boundaries',pre_subscription

    def command_capacity():
        with contextlib.closing(Socket(url)) as client:
            subscribe(client)
            for index in range(4095):
                identity='capacity-'+str(index)
                client.send(dict(type='pause',id=identity,typing=False,speaking=False))
                reply=client.until(lambda x:x.get('id')==identity)
                assert reply.get('type')=='ack', 'valid command refused before 4096-command capacity'
            client.send(dict(type='pause',id='one-command-over',typing=False,speaking=False))
            for _ in range(10):
                frame=client.ws.recv_frame()
                kind,data=frame.opcode,frame.data
                if kind==websocket.ABNF.OPCODE_CLOSE:
                    assert len(data)>2 and data[2:].decode('utf-8').strip(), 'command capacity close lacks reason'
                    return
            raise AssertionError('4097th command admitted instead of closing')
    yield 'exact-command-count-and-overflow',command_capacity

    for name,payload,opcode in [('binary',b'{}',2),('invalid-json',b'{',1),
                               ('invalid-utf8',b'{"type":"subscribe","id":"\xff"}',1),
                               ('non-object',b'[]',1),('missing-id',b'{"type":"subscribe"}',1),
                               ('empty-id',b'{"type":"subscribe","id":""}',1),
                               ('nonstring-id',b'{"type":"subscribe","id":12}',1),
                               ('oversize',oversized,1)]:
        def malformed(payload=payload,opcode=opcode):
            with contextlib.closing(Socket(url)) as client:
                subscribe(client)
                try:
                    client.ws.send(payload,opcode=opcode)
                except (BrokenPipeError, websocket.WebSocketConnectionClosedException):
                    # A peer may reject the advertised oversized length before
                    # this sender finishes its body. Still require its close
                    # frame and explanatory reason, not just a broken send.
                    if len(payload)<=65536:
                        raise
                for _ in range(10):
                    frame=client.ws.recv_frame()
                    kind,data=frame.opcode,frame.data
                    if kind==websocket.ABNF.OPCODE_CLOSE:
                        assert len(data)>2 and data[2:].decode('utf-8').strip(), 'malformed close lacks explanatory reason'
                        return
                raise AssertionError('malformed command did not close transport')
        yield 'malformed-'+name,malformed


class Backend(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        self.server.requests.append(self.path)
        self.send_error(500,'unexpected local fixture request')
    def log_message(self,*args):pass


def evaluate(binary):
    rows=[]
    backend=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend)
    backend.requests=[]
    thread=threading.Thread(target=backend.serve_forever,daemon=True);thread.start()
    with tempfile.TemporaryDirectory(prefix='ch07-wire-') as directory:
        env=os.environ.copy()
        env.update(LLM_VENDOR='openai',LLM_MODEL='fixture-ch07-independent',LLM_RESOLVED_MODEL='fixture-ch07-independent',LLM_API_KEY='LOCAL-FIXTURE-NOT-A-SECRET',LLM_BASE_URL=f'http://127.0.0.1:{backend.server_port}',CH02_LOG=str(Path(directory)/'events.log'))
        lines=queue.Queue()
        process=subprocess.Popen([str(binary),'--port','0'],cwd=directory,env=env,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True,start_new_session=True)
        def read():
            for line in process.stdout:lines.put(line)
            lines.put(None)
        reader=threading.Thread(target=read,daemon=True);reader.start()
        try:
            url=None
            for _ in range(100):
                line=lines.get(timeout=5)
                assert line is not None,'GUI exited before printing listening URL'
                match=re.search(r'http://127\.0\.0\.1:(\d+)',line)
                if match:
                    url=match.group(0);break
            assert url,'actual local listening URL absent'
            for name,action in cases(url):
                try:action();errors=[]
                except Exception as error:errors=[type(error).__name__+': '+str(error)]
                rows.append(dict(id=name,passed=not errors,details=errors))
            rows.append(dict(id='no-model-http',passed=not backend.requests,details=backend.requests))
        except Exception as error:
            rows.append(dict(id='server-start',passed=False,details=[type(error).__name__+': '+str(error)]))
        finally:
            if process.poll() is None:os.killpg(process.pid,signal.SIGTERM)
            try:process.wait(timeout=5)
            except subprocess.TimeoutExpired:os.killpg(process.pid,signal.SIGKILL);process.wait()
            process.stdout.close();reader.join(timeout=2)
            backend.shutdown();backend.server_close();thread.join(timeout=2)
    return dict(scope=__doc__,binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),checker_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),passed=bool(rows) and all(x['passed'] for x in rows),checks=rows)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);args=p.parse_args()
    result=evaluate(args.binary.resolve(strict=True));print(json.dumps(result,indent=2));return 0 if result['passed'] else 1


if __name__=='__main__':raise SystemExit(main())
