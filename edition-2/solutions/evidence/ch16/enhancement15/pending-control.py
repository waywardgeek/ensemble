import pathlib,tempfile,subprocess,json,os,shutil
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch16/enhancement15'
with tempfile.TemporaryDirectory(prefix='ch15-pending-control-') as tmp:
 tmp=pathlib.Path(tmp);tree=tmp/'source';shutil.copytree(root/'solutions/edition-2/ch15',tree)
 code=(root/'internal/gradeedition2/ch15.go').read_text();a=code.index('\tbase := newDir("compress")');b=code.index('\n\treturn r',a);code=code[:a]+'\tmemoryAbandon(run,newDir,&r)'+code[b:];code=code.replace('[]bool{false, true}','[]bool{true}');(tmp/'checks.go').write_text(code)
 probe='''package gradeedition2
import("testing";"os";"encoding/json";"github.com/waywardgeek/ensemble/internal/grade")
func TestPendingProbe(t *testing.T){bin,clean,e:=grade.Build(os.Getenv("MEMORY_PROBE_GUI"));if e!=nil{t.Fatal(e)};defer clean();r:=memoryChecks(bin);result:=map[string]string{"fatal":r.Fatal,"abandon-on-restart":r.Errs["abandon-on-restart"]};b,_:=json.Marshal(result);os.WriteFile(os.Getenv("MEMORY_PROBE_RESULT"),b,0644);if r.Fatal!=""||result["abandon-on-restart"]!=""{t.Fail()}}
''';(tmp/'probe.go').write_text(probe)
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/gradeedition2/ch15_test.go'):str(tmp/'probe.go'),str(root/'internal/gradeedition2/ch15.go'):str(tmp/'checks.go')}}))
 (out/'pending-only-selection.txt').write_text('Overlay selects only memoryAbandon and changes its existing []bool{false,true} loop to []bool{true}; actual held HTTP response, SIGKILL, restart and later successful checkpoint are unchanged. No other check is credited.\n')
 target=tree/'internal/agent/memory.go';original=target.read_text();old='launch := &common.CompactorLaunchedData';new='if task.Source != 0 { r.owner.memory.Archive(task); r.append(common.Event{Type:"band_depopulated",Depopulated:&common.BandDepopulatedData{Band:task.Source,Thru:&task.Thru,Reason:"disabled"}}) };\n launch := &common.CompactorLaunchedData'
 for label,mutation in [('pending-restored',False),('pending-premature-retirement',True)]:
  target.write_text(original.replace(old,new) if mutation else original)
  env=dict(os.environ,MEMORY_PROBE_GUI=str(tree/'gui'),MEMORY_PROBE_RESULT=str(out/(label+'.json')))
  r=subprocess.run(['go','test','-overlay',str(overlay),'./internal/gradeedition2','-run','^TestPendingProbe$','-count=1','-timeout','30s'],capture_output=True,text=True,env=env,timeout=40);(out/(label+'.txt')).write_text(r.stdout+r.stderr);print(label,r.returncode,(out/(label+'.json')).read_text());assert (r.returncode!=0)==mutation
