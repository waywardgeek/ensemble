"""Persist each local command's start, output and completion; no secret inputs."""
import datetime, json, os, pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parent
label, cwd, *argv = sys.argv[1:]
ledger = root / 'retained-repair-commands.jsonl'
def record(row):
    with ledger.open('a') as f:
        f.write(json.dumps(dict(time=datetime.datetime.now(datetime.timezone.utc).isoformat(), label=label, **row))+'\n'); f.flush(); os.fsync(f.fileno())
out = root / ('retained-repair-'+label+'.out')
with out.open('x') as f:
    record(dict(phase='start', cwd=cwd, argv=argv, free_bytes=os.statvfs(root).f_bavail*os.statvfs(root).f_frsize))
    result=subprocess.run(argv,cwd=cwd,stdout=f,stderr=subprocess.STDOUT)
    f.flush();os.fsync(f.fileno())
record(dict(phase='end',exit=result.returncode,output=str(out.relative_to(root)),free_bytes=os.statvfs(root).f_bavail*os.statvfs(root).f_frsize))
print(label, 'exit', result.returncode, out)
sys.exit(result.returncode)
