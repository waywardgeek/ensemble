"""Phase6 command receipts; no credential handling outside reviewed adapter."""
import datetime,json,os,pathlib,subprocess,sys
root=pathlib.Path(__file__).resolve().parent/'live-20261008'
label,cwd,*argv=sys.argv[1:]
def record(data):
 with (root/'commands.jsonl').open('a') as f:
  f.write(json.dumps(dict(time=datetime.datetime.now(datetime.timezone.utc).isoformat(),label=label,**data))+'\n');f.flush();os.fsync(f.fileno())
out=root/(label+'.out')
with out.open('x') as f:
 record(dict(phase='start',argv=argv,cwd=cwd))
 env={k:os.environ[k] for k in ('PATH','TERM','TMPDIR','LANG') if k in os.environ};env['PYTHONDONTWRITEBYTECODE']='1'
 result=subprocess.run(argv,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT);f.flush();os.fsync(f.fileno())
record(dict(phase='end',exit=result.returncode,output=out.name))
print(label,'exit',result.returncode)
sys.exit(result.returncode)
