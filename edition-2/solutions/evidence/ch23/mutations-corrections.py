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
 ('file-write-absolute-repaired','internal/sandbox/files.go','func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (*os.File, error) {','func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (*os.File, error) {\n if filepath.IsAbs(path) { return os.OpenFile(path,flags,mode) }','TestSandboxFiles'),
 ('grandchild-wrapper','internal/sandbox/command.go','if root != "" {\n\t\tif runtime.GOOS', 'if root != "" && !strings.HasPrefix(script,"sh -c") {\n\t\tif runtime.GOOS','TestSandboxCommands'),
 ('nested-cache-regression','internal/sandbox/command.go','cachePattern := "^" + regexp.QuoteMeta(filepath.Join(temp, "xcrun_db")) + "[^/]*$"','cachePattern := "^" + regexp.QuoteMeta(temp) + "/.*/xcrun_db[^/]*$"','TestSandboxCommands'),
 ('cwd-path-leak','internal/sandbox/command.go','cwd, err = s.Resolve(cwd)','if cwd != "" && !filepath.IsAbs(cwd) { cwd=filepath.Join(s.parent.Config().Workspace,cwd) }; cwd, err = s.Resolve(cwd)','TestSandboxCommands'),
 ('snapshot-backup','internal/history/journal.go','h.parent.Sandbox().WriteFile(path+".bak", old, 0600)','os.WriteFile(path+".bak", old, 0600)','TestSandboxStateFiles'),
 ('job-output','internal/jobs/jobs.go','j.parent.Sandbox().OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)','os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)','TestSandboxStateFiles'),
 ('media-read','internal/engine/media.go','e.parent.Sandbox().ReadFile(p.Ref.Locator)','os.ReadFile(p.Ref.Locator)','TestSandboxMedia'),
]

results=[]
try:
 baseline=subprocess.run(['go','test','./ensemble','-run','TestSandbox','-count=1'],cwd=work,capture_output=True,text=True)
 (evidence/'mutation-final-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,file,old,new,test in cases:
  path=work/file;original=path.read_text();count=original.count(old)
  if count!=1:raise RuntimeError((name,'replacement count',count))
  mutated=original.replace(old,new)
  if name=='media-read':mutated=mutated.replace('"fmt"','"fmt"\n "os"')
  path.write_text(mutated)
  try:r=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1','-v'],cwd=work,capture_output=True,text=True,timeout=30)
  finally:path.write_text(original)
  result={'name':name,'file':file,'old':old,'new':new,'test':test,'exit':r.returncode,'output':r.stdout+r.stderr}
  results.append(result)
  (evidence/'mutations-corrections.json').write_text(json.dumps(results,indent=2)+'\n')
  print(name,r.returncode,flush=True)
finally:shutil.rmtree(work)
