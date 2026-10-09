"""Preserve preparation commands before/after; no Go commands in this phase."""
import datetime,json,os,pathlib,subprocess,sys
root=pathlib.Path(__file__).resolve().parent
label,cwd,*argv=sys.argv[1:]
assert argv and argv[0] not in ('go','gofmt'), 'compiler belongs to independent grader'
def record(data):
 with (root/'provider-prep-commands.jsonl').open('a') as f:
  f.write(json.dumps(dict(time=datetime.datetime.now(datetime.timezone.utc).isoformat(),label=label,**data))+'\n');f.flush();os.fsync(f.fileno())
out=root/('provider-prep-'+label+'.out')
with out.open('x') as f:
 record(dict(phase='start',argv=argv,cwd=cwd))
 env={k:os.environ[k] for k in ('PATH','TMPDIR','LANG') if k in os.environ};env['PYTHONDONTWRITEBYTECODE']='1'
 result=subprocess.run(argv,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT);f.flush();os.fsync(f.fileno())
record(dict(phase='end',exit=result.returncode,output=out.name))
print(label,'exit',result.returncode)
sys.exit(result.returncode)
