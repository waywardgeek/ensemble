#!/usr/bin/env python3
"""Local command recorder; never reads credentials or launches model providers."""
import hashlib,json,pathlib,subprocess,sys,time
root=pathlib.Path(__file__).resolve().parents[2]
label=sys.argv[1]; command=sys.argv[2:]
out=root/'evidence'/'ch11'/'local';out.mkdir(exist_ok=True)
path=out/(label+'.json')
if path.exists(): raise SystemExit('receipt exists')
def inputs():
 result={}
 for p in sorted(root.rglob('*')):
  if 'evidence' in p.relative_to(root).parts or not p.is_file():continue
  if p.suffix in ('.go','.js','.html','.css') or p.name in ('go.mod','go.sum'):
   result['solutions/edition-2/main/'+str(p.relative_to(root))]=hashlib.sha256(p.read_bytes()).hexdigest()
 return result
before=inputs();start=time.time()
p=subprocess.run(command,cwd=root,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
record={'command':command,'cwd':str(root),'source_sha256':before,'after_source_sha256':inputs(),'exit':p.returncode,'output':p.stdout,'elapsed_seconds':time.time()-start}
path.write_text(json.dumps(record,indent=2)+'\n');print(p.stdout,end='');print('receipt:',path,'exit:',p.returncode)
sys.exit(p.returncode)
