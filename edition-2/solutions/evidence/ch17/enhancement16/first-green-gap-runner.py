import pathlib,shutil,tempfile,subprocess,os,json,time,difflib,hashlib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch17/enhancement16';source=root/'solutions/edition-2/ch16'
cases=[]
def add(name,file,old,new,group):cases.append((name,file,old,new,group))
R='internal/recall/recall.go';I='internal/recall/index.go';J='internal/recall/judge.go'
add('first-green-no-paragraph-split',I,'strings.Split(part.content, "\\n\\n")','[]string{part.content}','chunk')
add('first-green-no-spare-slots',R,'if len(out) < limit && !selected[i] {','if false && len(out) < limit && !selected[i] {','quota')
results=[]
with tempfile.TemporaryDirectory(prefix='ch16-mutants-') as tmp:
 work=pathlib.Path(tmp)/'source';shutil.copytree(source,work);binary=pathlib.Path(tmp)/'gui'
 for name,file,old,new,group in cases:
  p=work/file;before=p.read_text();assert old in before,(name,file);after=before.replace(old,new);p.write_text(after)
  (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile=file,tofile=file)))
  build=subprocess.run(['go','build','-o',str(binary),'./cmd'],cwd=work/'gui',capture_output=True,text=True)
  if build.returncode:
   result={'name':name,'group':group,'build_exit':build.returncode,'output':build.stdout+build.stderr}
  else:
   env=dict(os.environ,RECALL_BIN=str(binary),RECALL_GROUP=group,RECALL_CORE=str(work),RECALL_GUI=str(work/'gui'))
   start=time.monotonic();r=subprocess.run(['/tmp/ch16-old-focus','-test.run','^TestChapter16Focus$','-test.v'],env=env,cwd=root,capture_output=True,text=True,timeout=90)
   result={'name':name,'group':group,'build_exit':0,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'output':r.stdout+r.stderr}
  (out/(name+'.txt')).write_text(result['output']);results.append(result);(out/'first-green-gap-controls.json').write_text(json.dumps(results,indent=2)+'\n');print(name,result.get('exit',result['build_exit']),flush=True);p.write_text(before)
print('scratch removed')
