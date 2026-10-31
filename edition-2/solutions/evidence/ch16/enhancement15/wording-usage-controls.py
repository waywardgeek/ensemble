import subprocess,pathlib,tempfile,shutil,json,difflib,time,os
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch16/enhancement15';source=root/'solutions/edition-2/ch15';out=out/'wording-usage-final';out.mkdir(exist_ok=True)
# The overlay exposes the existing grader to a scratch binary; it changes no runtime.
probe='''package gradeedition2
import("testing";"os";"encoding/json";"github.com/waywardgeek/ensemble/internal/grade")
func TestMemoryProbe(t *testing.T){bin,clean,e:=grade.Build(os.Getenv("MEMORY_PROBE_GUI"));if e!=nil{t.Fatal(e)};defer clean();r:=memoryChecks(bin);checks:=grade.Ch16Checks(grade.Ch16Result{Fatal:r.Fatal,Errs:r.Errs});checks=checks[:len(checks)-1];report:=grade.NewTitledReport("Focused memory",checks,"");b,_:=json.Marshal(report);if e=os.WriteFile(os.Getenv("MEMORY_PROBE_RESULT"),b,0644);e!=nil{t.Fatal(e)};if !report.Passed{t.Fail()}}
'''
(out/'probe.go.txt').write_text(probe)
cases=[
('omit-cache-occupancy','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input','forced-handoff'),
('output-occupancy','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input + u.CacheRead + u.CacheWrite + u.Output','forced-handoff'),
]

results=json.loads((out/'mutations.json').read_text()) if (out/'mutations.json').exists() else []
with tempfile.TemporaryDirectory(prefix='ch15-mutants-') as tmp:
 tmp=pathlib.Path(tmp);tree=tmp/'source';shutil.copytree(source,tree)
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/gradeedition2/ch15_test.go'):str(out/'probe.go.txt')}}))
 for label,file,old,new,intended in cases:
  if any(x['case']==label for x in results):continue
  if label=='ignore-human-disk-edit':
   prefix='func populate(c *common.Context, p *common.BandPopulatedData) {\n\tc.DisabledBands[p.Band] = false\n\t';old=prefix+old;new=prefix+new
  path=tree/file;original=path.read_text();assert original.count(old)==1,(label,original.count(old));mutated=original.replace(old,new);
  if label in ['abandoned-fold-retried','startup-recompresses-memory']:mutated=mutated.replace('\"fmt\"','\"fmt\"\n \"context\"').replace('Compress(nil,','Compress(context.Background(),')
  path.write_text(mutated)
  (out/(label+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=file,tofile=file)))
  env=dict(os.environ,MEMORY_PROBE_GUI=str(tree/'gui'),MEMORY_PROBE_RESULT=str(out/(label+'.json')))
  command=['go','test','-overlay',str(overlay),'./internal/gradeedition2','-run','^TestMemoryProbe$','-count=1','-timeout','70s']
  start=time.monotonic();r=subprocess.run(command,capture_output=True,text=True,env=env,timeout=85)
  (out/(label+'.txt')).write_text(r.stdout+r.stderr)
  try:report=json.loads((out/(label+'.json')).read_text())
  except:report={}
  failures=[c for c in report.get('checks',[]) if not c['passed']]
  result={'case':label,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'intended_check':intended,'intended_failure':any(c['id']==intended for c in failures),'score':report.get('score'),'failures':failures}
  results.append(result);(out/'mutations.json').write_text(json.dumps(results,indent=2)+'\n');print(label,r.returncode,report.get('score'),result['intended_failure'],flush=True)
  path.write_text(original)
