import subprocess,pathlib,tempfile,shutil,json,difflib,time,os
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch16/enhancement15';source=root/'solutions/edition-2/ch15'
# The overlay exposes the existing grader to a scratch binary; it changes no runtime.
probe='''package gradeedition2
import("testing";"os";"encoding/json";"github.com/waywardgeek/ensemble/internal/grade")
func TestMemoryProbe(t *testing.T){bin,clean,e:=grade.Build(os.Getenv("MEMORY_PROBE_GUI"));if e!=nil{t.Fatal(e)};defer clean();r:=memoryChecks(bin);checks:=grade.Ch16Checks(grade.Ch16Result{Fatal:r.Fatal,Errs:r.Errs});checks=checks[:len(checks)-1];report:=grade.NewTitledReport("Focused memory",checks,"");b,_:=json.Marshal(report);if e=os.WriteFile(os.Getenv("MEMORY_PROBE_RESULT"),b,0644);e!=nil{t.Fatal(e)};if !report.Passed{t.Fail()}}
'''
(out/'probe.go.txt').write_text(probe)
cases=[
('no-compression','internal/memory/memory.go','if total <= s.Budget {','if true {','micro-handoff-compresses'),
('first-source-truncated','internal/engine/compress.go','ratio, source)','ratio, source[len(source)/2:])','micro-handoff-compresses'),
('returned-memory-truncated','internal/agent/memory.go','Text: m.text, Source: "compression"','Text: "I forgot the returned memory.", Source: "compression"','planted-fact-survives'),
('raw-compressed-work-retained','internal/history/reduce.go','redact(c, pending.Cut)','// redact(c, pending.Cut)','micro-handoff-compresses'),
('missing-first-person','internal/engine/compress.go','writing in first person as the agent','writing as a detached observer','memory-is-data'),
('newest-first-graduation','internal/memory/memory.go','for _, e := range files {','for i := len(files)-1; i >= 0; i-- { e := files[i]','graduation-fires-oldest-first'),
('graduation-keeps-old-source','internal/history/reduce.go','depopulate(c, pending.Retire)','// depopulate(c, pending.Retire)','graduation-fires-oldest-first'),
('no-disabled-depopulation','internal/agent/memory.go','if disabled {','if false {','disable-enable-idempotent'),
('ignore-human-disk-edit','internal/history/reduce.go','kind := common.KindSoul + common.EntryKind(p.Band-1)\n\tvar out []common.Entry','kind := common.KindSoul + common.EntryKind(p.Band-1)\n if p.Source == "disk" { for _,e := range c.Dialogue { if e.Kind == kind && e.File == p.File { return } } }\n\tvar out []common.Entry','disable-enable-idempotent'),
('no-memory-canonical-sort','internal/history/reduce.go','sort.SliceStable(out, func(i, j int) bool {','sort.SliceStable(out, func(i, j int) bool { return false;','disable-enable-idempotent'),
('duplicate-memory-identities','internal/history/reduce.go','if e.Kind != kind || (p.Band > common.BandMemory && e.File != p.File) {','if true {','disable-enable-idempotent'),
('ignore-disabled-neighbor','internal/memory/memory.go','if m.Settings(b - 1).Disabled {','if false {','disabled-neighbor-refused'),
('lifetime-occupancy','internal/agent/memory.go','u := c.LastUsage','u := c.Usage','forced-handoff'),
('omit-cache-occupancy','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input','forced-handoff'),
('double-cache-occupancy','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input + 2*u.CacheRead + 2*u.CacheWrite','forced-handoff'),
('output-occupancy','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input + u.CacheRead + u.CacheWrite + u.Output','forced-handoff'),
('sticky-handoff-restriction','internal/agent/dispatch.go','r.owner.forceHandoff.Store(false)','// r.owner.forceHandoff.Store(false)','forced-handoff'),
('no-near-limit-warning','internal/agent/memory.go','Context is at least 90% full.','Context is growing.','forced-handoff'),
('failed-fold-retires-source','internal/agent/memory.go','if m.err != nil {\n\t\tr.compressionWarning(m.err)','if m.err != nil {\n if t.Source != 0 { r.owner.memory.Archive(t); r.append(common.Event{Type:"band_depopulated",Depopulated:&common.BandDepopulatedData{Band:t.Source,Thru:&t.Thru,Reason:"disabled"}}) };\n\t\tr.compressionWarning(m.err)','abandon-on-restart'),
('pending-fold-retires-source','internal/agent/memory.go','launch := &common.CompactorLaunchedData','if task.Source != 0 { r.owner.memory.Archive(task); r.append(common.Event{Type:"band_depopulated",Depopulated:&common.BandDepopulatedData{Band:task.Source,Thru:&task.Thru,Reason:"disabled"}}) };\n launch := &common.CompactorLaunchedData','abandon-on-restart'),
('skip-startup-disk','internal/agent/memory.go','files, err := r.owner.memory.Scan(b)','files, err := r.owner.memory.Scan(b); if !r.memoryReady { files = nil }','fresh-start-populates'),
('nonzero-after-complete','gui/cmd/main.go','func main() {','func main() { defer os.Exit(17)','micro-handoff-compresses'),
('auxiliary-usage-as-main','internal/agent/memory.go','t := m.task','t := m.task; r.append(common.Event{Type:"response_ended",Response:&common.ResponseData{Usage:common.Usage{Input:12000}}})','forced-handoff'),
('abandoned-fold-retried','internal/agent/memory.go','r.owner.Ensemble().Logf("memory: abandoned compressor; source retained until a later checkpoint")','r.owner.Ensemble().Logf("memory: abandoned compressor; source retained until a later checkpoint"); if task,_ := r.owner.memory.Graduation(); task != nil { r.owner.engine.Compress(nil,task.Text,task.Ratio) }','abandon-on-restart'),

('disabled-session-still-compresses','internal/memory/memory.go','s.Disabled || m.Settings(common.BandSession).Disabled','s.Disabled','disable-enable-idempotent'),
('memory-added-to-system','internal/engine/anthropic.go','maxTokens, e.SystemPrompt(), messages','maxTokens, e.SystemPrompt()+func()string{ s:="";for _,entry:=range c.Dialogue{if entry.Kind>=common.KindSoul{for _,p:=range entry.Parts{s+=p.Text}}};return s}(), messages','memory-is-data'),
('startup-recompresses-memory','internal/agent/memory.go','r.memoryReady = true','if !r.memoryReady { for _,e := range r.owner.history.Context().Dialogue { if e.Kind >= common.KindSoul && len(e.Parts)>0 { r.owner.engine.Compress(nil,e.Parts[0].Text,2); break } } }; r.memoryReady = true','replay-needs-no-llm'),

('engine-total-occupancy','internal/agent/memory.go','u := c.LastUsage','u := r.owner.engine.Usage()','forced-handoff'),

('memory-in-inline-system','internal/engine/anthropic.go','role := "user"','role := "user"; if entry.Kind >= common.KindSoul { role = "system" }','memory-is-data'),

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
