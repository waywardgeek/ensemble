import pathlib,subprocess,tempfile,shutil,json,time,difflib,os
repo=pathlib.Path.cwd();e=repo/'solutions/edition-2/evidence/ch23/enhancement21';work=pathlib.Path(tempfile.mkdtemp(prefix='ch21-mutations-'));core=work/'core';shutil.copytree(repo/'solutions/edition-2/ch21',core)
env={k:v for k,v in os.environ.items() if k in ['HOME','PATH','TMPDIR','USER','LANG','SHELL']};env['GOWORK']='off'
subprocess.run(['go','build','-o',str(work/'grader'),'./cmd/grade-edition2'],check=True,env=env)
old=json.loads((repo/'solutions/edition-2/evidence/ch23/mutations.json').read_text())
names=['file-read-absolute','host-path-error','kernel-read','kernel-write','network','child-derive','child-safe','child-alarm','safe-mode']
cases=[(v['name'],[(v['file'],v['old'],v['new'])]) for v in old if v['name'] in names]
cases += [('environment',[('internal/sandbox/command.go','out = append(out, "TERM=dumb")','out = append(out, "TERM=dumb", "SERVICE_TOKEN="+os.Getenv("SERVICE_TOKEN"))')]),('read-empty-success',[('internal/tools/tools.go','data, err := owner.Agent().Sandbox().ReadFile(args.Path)','data, err := owner.Agent().Sandbox().ReadFile(args.Path)\n if err == nil { data = nil }')]),('exit-after-success',[('main.go','func main() {','func main() {\n defer os.Exit(17)')])]
results=[]
try:
 for name,edits in [('baseline',[])]+cases+[('restored',[])]:
  originals={};patch=''
  for file,a,b in edits:
   p=core/file;s=p.read_text();assert s.count(a)==1,(name,file,s.count(a));new=s.replace(a,b);originals[p]=s;p.write_text(new);patch+=''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
  if edits:(e/(name+'.patch')).write_text(patch)
  try:
   t=time.monotonic();r=subprocess.run([str(work/'grader'),'-ch','21','-json',str(core)],capture_output=True,text=True,env=env,timeout=45)
   (e/(name+'-mutation.json')).write_text(r.stdout);(e/(name+'-mutation-stderr.txt')).write_text(r.stderr)
   report=json.loads(r.stdout);item={'name':name,'exit':r.returncode,'seconds':round(time.monotonic()-t,3),'score':report['score'],'failures':{c['id']:c.get('details',[]) for c in report['checks'] if not c['passed']}};results.append(item);(e/'mutations.json').write_text(json.dumps(results,indent=2)+'\n');print(name,item['score'],list(item['failures']),flush=True)
  finally:
   for p,s in originals.items():p.write_text(s)
finally:shutil.rmtree(work)
