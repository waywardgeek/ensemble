#!/usr/bin/env python3
"""Strict settings command shapes, no-mutation recovery and no model HTTP."""
import argparse
import contextlib
import hashlib
import http.server
import json
from pathlib import Path
import tempfile
import threading
from accept_ch07 import Socket
from accept_ch08 import Backend, Program, defaults, subscribe, update


def cases(domain):
    patch='{"theme":"light"}' if domain=='preferences' else '{"max_model_requests":2}'
    rows=[]
    for name,raw in [('null','null'),('boolean','true'),('string','"0"'),('negative','-1'),('fraction','0.5'),('array','[]'),('object','{}')]:
        rows.append(('base-'+name,'"base_revision":'+raw+',"patch":'+patch))
    rows.append(('missing-base','"patch":'+patch))
    rows.append(('missing-patch','"base_revision":0'))
    for name,raw in [('null','null'),('boolean','true'),('number','0'),('string','"x"'),('array','[]'),('empty','{}')]:
        rows.append(('patch-'+name,'"base_revision":0,"patch":'+raw))
    rows.extend([('extra-path','"base_revision":0,"patch":'+patch+',"path":"PRIVATE-PATH-SENTINEL"'),
                 ('duplicate-base','"base_revision":0,"base_revision":0,"patch":'+patch),
                 ('duplicate-patch','"base_revision":0,"patch":'+patch+',"patch":'+patch)])
    return [(name,'{"type":"'+domain+'_update","id":"'+domain+'-'+name+'",'+members+'}') for name,members in rows]


def evaluate(binary):
    backend=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend);backend.requests=[]
    worker=threading.Thread(target=backend.serve_forever,daemon=True);worker.start();rows=[]
    try:
        with tempfile.TemporaryDirectory(prefix='ch08-command-shapes-') as temporary:
            root=Path(temporary);endpoint=f'http://127.0.0.1:{backend.server_port}'
            with contextlib.closing(Program(binary,root,endpoint,1)) as program:
                with contextlib.closing(Socket(program.listen())) as client:
                    subscribe(client,defaults(),0,0,0)
                    rows.append(dict(id='required-positive-settings-subscription',passed=True))
                    for domain in ('preferences','policy'):
                        for name,raw in cases(domain):
                            identifier=domain+'-'+name
                            try:
                                client.ws.send(raw)
                                reply=client.until(lambda r:r.get('id')==identifier)
                                assert reply.get('type')=='error' and reply.get('code')=='invalid_command','correctable command shape returned wrong refusal: '+str(reply.get('code'))
                                assert isinstance(reply.get('message'),str) and reply['message'],'missing safe reason'
                                assert 'PRIVATE-PATH-SENTINEL' not in json.dumps(reply),'unsafe invalid-command diagnostic'
                                rows.append(dict(id=identifier,passed=True))
                            except Exception as error:
                                rows.append(dict(id=identifier,passed=False,details=str(error)))
                    # A successful no-op with the original revision distinguishes
                    # safe refusal from a partial change hidden by an error reply.
                    update(client,'preferences','unchanged-preferences',0,dict(autoplay=False),0,defaults())
                    update(client,'policy','unchanged-policy',0,dict(max_model_requests=0),0,0)
                    assert not (root/'preferences.json').exists() and not (root/'policy.json').exists(),'refused/no-change command wrote a settings file'
                    update(client,'preferences','recovery',0,dict(autoplay=True),1,defaults()|dict(autoplay=True))
                    update(client,'policy','recovery-policy',0,dict(max_model_requests=2),1,2)
                    rows.append(dict(id='original-revisions-and-valid-recovery',passed=True))
    except Exception as error:
        rows.append(dict(id='startup-or-recovery',passed=False,details=str(error)))
    finally:
        backend.shutdown();backend.server_close();worker.join(timeout=2)
    rows.append(dict(id='no-model-http',passed=not backend.requests,requests=backend.requests))
    return dict(scope=__doc__,passed=len(rows)==39 and all(r['passed'] for r in rows),checks=rows,
                binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                checker_files={name:hashlib.sha256(Path(__file__).with_name(name).read_bytes()).hexdigest() for name in ['accept_ch08_commands.py','accept_ch08.py','accept_ch07.py']})


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('binary',type=Path)
    args=parser.parse_args();result=evaluate(args.binary.resolve(strict=True));print(json.dumps(result,indent=2));raise SystemExit(0 if result['passed'] else 1)
