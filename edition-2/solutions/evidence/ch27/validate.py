"""Bounded outside-module verification; it never reads provider credentials."""
import json, os, pathlib, socket, subprocess, tempfile
HERE = pathlib.Path(__file__).resolve().parent
MAIN = HERE.parents[1] / 'main'
BIN = pathlib.Path('/tmp/ensemble-gateway26-binaries/gateway')
def run(args, cwd):
    p = subprocess.run(args, cwd=cwd, text=True, capture_output=True, timeout=60, env={**os.environ, 'GOPROXY':'off'})
    return {'command': args, 'exit': p.returncode, 'stdout': p.stdout, 'stderr': p.stderr}
with tempfile.TemporaryDirectory(prefix='outside-gateway26-') as scratch:
    root = pathlib.Path(scratch)
    (root/'go.mod').write_text(f'module outside.example/gatewayconsumer\n\ngo 1.25.0\n\nrequire ensemble v0.0.0\nreplace ensemble => {MAIN}\n')
    (root/'main.go').write_bytes((HERE/'outside-consumer.go').read_bytes())
    data = root/'data'; data.mkdir()
    with socket.socket() as sock:
        sock.bind(('127.0.0.1',0)); address=f'127.0.0.1:{sock.getsockname()[1]}'
    records=[run(['go','mod','tidy'],root)]
    records.append(run(['go','build','-o',str(root/'consumer'),'.'],root))
    records.append(run([str(root/'consumer'),str(BIN),str(data),address],root))
    deps=run(['go','list','-deps','ensemble/gateway'],root)
    headless=run(['go','list','-deps','ensemble/ensemble'],MAIN)
    records.append({'client_no_internal':not any(x.startswith('ensemble/internal/') for x in deps['stdout'].splitlines()),'client_positive_websocket':'github.com/gorilla/websocket' in deps['stdout'].splitlines(),'core_no_websocket':headless['exit']==0 and 'ensemble/ensemble' in headless['stdout'].splitlines() and 'github.com/gorilla/websocket' not in headless['stdout'].splitlines(),'deps_exit':[deps['exit'],headless['exit']]})
    (HERE/'outside-result.json').write_text(json.dumps(records,indent=2)+'\n')
    print(json.dumps(records[-1])); print(records[-2]['stdout']); print('command exits:', [r['exit'] for r in records if 'exit' in r])
