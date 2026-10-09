"""Local quality command receipts; no credential handling or provider access."""
import datetime
import json
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parent / 'quality-20261009'
root.mkdir(exist_ok=True)
label, cwd, *argv = sys.argv[1:]
def record(value):
    with (root/'commands.jsonl').open('a') as f:
        f.write(json.dumps(dict(time=datetime.datetime.now(datetime.timezone.utc).isoformat(),
                                label=label, **value))+'\n')
        f.flush(); os.fsync(f.fileno())
out=root/(label+'.out')
with out.open('x') as f:
    record(dict(phase='start',argv=argv,cwd=cwd))
    env={k:os.environ[k] for k in ('PATH','TERM','TMPDIR','LANG','HOME','GOCACHE','GOMODCACHE','GOPATH') if k in os.environ}
    env['PYTHONDONTWRITEBYTECODE']='1'
    result=subprocess.run(argv,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT)
    f.flush();os.fsync(f.fileno())
record(dict(phase='end',exit=result.returncode,output=out.name))
print(label,'exit',result.returncode)
sys.exit(result.returncode)
