#!/usr/bin/env python3
"""Exact policy projection positive and one numeric-conversion deletion.

Usage: python3 audit_ch08_revision_projection.py SOURCE_DIRECTORY GUI_BINARY
The caller binds the source/binary pair. The immutable chapter gate does so.
"""
import contextlib
import hashlib
import http.server
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import threading
from accept_ch07 import Socket
from accept_ch08 import Backend, Program, defaults, subscribe

source=Path(sys.argv[1]).resolve()
binary=Path(sys.argv[2]).resolve()
MAX=(1<<64)-1
backend=http.server.ThreadingHTTPServer(('127.0.0.1',0),Backend)
backend.requests=[]
thread=threading.Thread(target=backend.serve_forever,daemon=True);thread.start()

def wire(executable):
    with tempfile.TemporaryDirectory(prefix='ch08-projection-wire-') as temporary:
        directory=Path(temporary)
        for domain,value in [('preferences',defaults()),('policy',{'max_model_requests':0})]:
            (directory/(domain+'.json')).write_text(json.dumps(dict(version=1,revision=MAX-1,**{domain:value})))
        with contextlib.closing(Program(executable,directory,f'http://127.0.0.1:{backend.server_port}',1)) as program:
            with contextlib.closing(Socket(program.listen())) as socket:
                subscribe(socket,defaults(),MAX-1,0,MAX-1)
                socket.send(dict(type='policy_update',id='update',base_revision=MAX-1,patch={'max_model_requests':2}))
                observation=socket.until(lambda m:m.get('observation',{}).get('kind')=='policy_changed')
                actual=observation['observation']['execution_policy']['revision']
                if actual!=MAX:
                    return dict(passed=False,reason='policy_changed exact revision lost',expected=MAX,actual=actual,records=socket.records)
                ack=socket.until(lambda m:m.get('id')=='update')
                assert ack['type']=='policy_ack' and ack['revision']==MAX,'exact acknowledgement missing'
                return dict(passed=True,records=socket.records)
try:
    baseline=wire(binary)
    assert baseline['passed'],'valid projection positive failed'
    with tempfile.TemporaryDirectory(prefix='ch08-projection-deletion-') as temporary:
        root=Path(temporary)/'source'
        shutil.copytree(source,root,ignore=shutil.ignore_patterns('evidence','.git'))
        hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in root.rglob('*') if p.is_file()}
        target=root/'gui/projection.go';data=target.read_text();anchor='\tdecoder.UseNumber()\n'
        assert data.count(anchor)==1,'projection numeric-conversion anchor changed'
        target.write_text(data.replace(anchor,''))
        mutant=Path(temporary)/'gui'
        build=subprocess.run(['go','build','-o',str(mutant),'./cmd/ensemble-gui'],cwd=root/'gui',capture_output=True,text=True,timeout=120)
        assert build.returncode==0,build.stderr
        negative=wire(mutant)
        passed=negative.get('reason')=='policy_changed exact revision lost' and negative.get('actual')==18446744073709552000 and not backend.requests
        print(json.dumps(dict(scope=__doc__,passed=passed,source_files=hashes,positive_binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),baseline=baseline,mutation=dict(removed='one decoder.UseNumber call',binary_sha256=hashlib.sha256(mutant.read_bytes()).hexdigest(),result=negative),model_requests=backend.requests),indent=2))
        sys.exit(0 if passed else 1)
finally:
    backend.shutdown();backend.server_close();thread.join()
