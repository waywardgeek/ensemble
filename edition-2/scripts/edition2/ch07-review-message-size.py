#!/usr/bin/env python3
"""Incoming WebSocket size controls beyond one byte over the limit; no model calls."""
import argparse,contextlib,hashlib,json
from pathlib import Path
import websocket
import accept_ch07 as wire
original_cases=wire.cases

def cases(url):
    for name,action in original_cases(url):
        if name in ('exact-command-size','malformed-oversize'):
            yield name,action
    for size,fragmented in ((131072,False),(1048576,False),(131072,True)):
        def check(size=size,fragmented=fragmented):
            base=json.dumps(dict(type='subscribe',id=''),separators=(',',':')).encode()
            body=json.dumps(dict(type='subscribe',id='z'*(size-len(base))),separators=(',',':')).encode()
            assert len(body)==size
            with contextlib.closing(wire.Socket(url)) as client:
                wire.subscribe(client)
                try:
                    if fragmented:
                        client.ws.send_frame(websocket.ABNF.create_frame(body[:40000],websocket.ABNF.OPCODE_TEXT,fin=0))
                        client.ws.send_frame(websocket.ABNF.create_frame(body[40000:80000],websocket.ABNF.OPCODE_CONT,fin=0))
                        client.ws.send_frame(websocket.ABNF.create_frame(body[80000:],websocket.ABNF.OPCODE_CONT,fin=1))
                    else:client.ws.send(body,opcode=websocket.ABNF.OPCODE_TEXT)
                except (BrokenPipeError,websocket.WebSocketConnectionClosedException):pass
                for _ in range(10):
                    frame=client.ws.recv_frame()
                    if frame.opcode==websocket.ABNF.OPCODE_CLOSE:
                        assert len(frame.data)>2 and frame.data[2:].decode('utf-8').strip(), 'oversized message close lacks explanatory reason'
                        return
                raise AssertionError('oversized message did not close')
        yield ('fragmented' if fragmented else 'single-frame')+f'-{size}-byte-close',check

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('binary',type=Path);a=p.parse_args();wire.cases=cases;r=wire.evaluate(a.binary.resolve());r['scope']=__doc__;r['review_checker_sha256']=hashlib.sha256(Path(__file__).read_bytes()).hexdigest();print(json.dumps(r,indent=2));raise SystemExit(0 if r['passed'] else 1)
