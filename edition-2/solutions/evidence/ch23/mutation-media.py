import pathlib,tempfile,subprocess,shutil,json
repo=pathlib.Path('/Users/bill/projects/ensemble')
evidence=repo/'solutions/edition-2/evidence/ch23'
work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch23-mutants-'))
subprocess.run(['git','archive','fcb98a9459b328314f0a7ff06efbb5e7c85fe07e:solutions/edition-2/main'],cwd=repo,stdout=open(work/'source.tar','wb'),check=True)
subprocess.run(['tar','-xf',str(work/'source.tar'),'-C',str(work)],check=True)
(work/'source.tar').unlink()
# Overlay only current owned source changes on the first-green archive.
for source in (repo/'solutions/edition-2/main').rglob('*.go'):
 if 'node_modules' in source.parts: continue
 relative=source.relative_to(repo/'solutions/edition-2/main')
 target=work/relative
 if target.exists(): target.write_bytes(source.read_bytes())
cases=[
 ('media-read-repaired','internal/engine/media.go','e.parent.Sandbox().ReadFile(p.Ref.Locator)','os.ReadFile(p.Ref.Locator)','TestSandboxMedia'),
]

results=[]
try:
 baseline=subprocess.run(['go','test','./ensemble','-run','TestSandbox','-count=1'],cwd=work,capture_output=True,text=True)
 (evidence/'mutation-media-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,file,old,new,test in cases:
  path=work/file;original=path.read_text();count=original.count(old)
  if count!=1:raise RuntimeError((name,'replacement count',count))
  mutated=original.replace(old,new)
  if name=='media-read-repaired':mutated=mutated.replace('"fmt"','"fmt"\n "os"')
  path.write_text(mutated)
  try:r=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1','-v'],cwd=work,capture_output=True,text=True,timeout=30)
  finally:path.write_text(original)
  result={'name':name,'file':file,'old':old,'new':new,'test':test,'exit':r.returncode,'output':r.stdout+r.stderr}
  results.append(result)
  (evidence/'mutation-media-correction.json').write_text(json.dumps(results,indent=2)+'\n')
  print(name,r.returncode,flush=True)
finally:shutil.rmtree(work)
