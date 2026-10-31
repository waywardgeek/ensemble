# One plausible defect at a time in a disposable copy of the submitted core.
import pathlib,shutil,subprocess,tempfile,json
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch15-mutants-'))/'main'
shutil.copytree(base/'main',scratch,ignore=shutil.ignore_patterns('node_modules','package-lock.json'))
cases=[
 ('keep_previous','internal/agent/context.go','if call.Name == "keep_tool_results" {','if call.Name == "ignored_keep" {','TestContextKeepPreviousBatch','keep did not preserve preceding batch'),
 ('model_stubbing_gate','internal/agent/context.go','if !r.owner.engine.Features().StubsToolResults {','if false {','TestContextKeepPreviousBatch','stubbing ignored model'),
 ('unseen_results','internal/agent/context.go','len(batch.calls) != 0 || batch.last > c.SentThrough','len(batch.calls) != 0','TestContextLadderRecordedAndReplay','cut before first presentation'),
 ('result_band','internal/agent/context.go','func (r *runtime) ladder(level string, budget int) error {','func (r *runtime) ladder(level string, budget int) error {\n if level == "redact_result" { return nil }','TestContextLadderRecordedAndReplay','ladder did not exercise both bands'),
 ('calls_band','internal/agent/context.go','func (r *runtime) ladder(level string, budget int) error {','func (r *runtime) ladder(level string, budget int) error {\n if level == "redact_tool" { return nil }','TestContextLadderRecordedAndReplay','ladder did not exercise both bands'),
 ('batch_holding','internal/history/reduce.go','if !toolsComplete(c.Dialogue) {','if false {','TestContextSurvivorsAndFrozenPrefix','tool result has no matching call'),
 ('handoff_clear','internal/history/reduce.go','if entry.Kind == common.KindHandoff {','if false {','TestContextSurvivorsAndFrozenPrefix','survivor missing or handoff duplicated'),
 ('skill_survivor','internal/history/reduce.go','case "skill_loaded":','case "skill_loaded":\n return','TestContextSurvivorsAndFrozenPrefix','survivor missing or handoff duplicated'),
 ('inline_prefix','internal/engine/engine.go','if e.Features().InlineTools {\n\t\treturn *c.StartupTools','if false {\n\t\treturn *c.StartupTools','TestContextSurvivorsAndFrozenPrefix','prefix changed'),
 ('tool_deltas','internal/agent/context.go','if len(delta.Added)+len(delta.Removed) > 0 {','if false {','TestContextSurvivorsAndFrozenPrefix','declaration delta missing'),
 ('journal_append','internal/history/journal.go','if h.journal == "" {','if true {','TestContextTotalReducerAndJournal','journal tail lost'),
 ('backup','internal/history/journal.go','if err = os.WriteFile(path+".bak", old, 0600); err != nil {','if err = os.WriteFile(path+".wrong", old, 0600); err != nil {','TestContextTotalReducerAndJournal','previous snapshot backup lost'),
 ('retention','internal/history/save.go','if cfg.LogRetention > 0 && len(log) > cfg.LogRetention {','if false {','TestContextTotalReducerAndJournal','retention crossed anchor'),
 ('known_event_atomicity','internal/history/history.go','if err := validate(e); err != nil {\n\t\treturn err\n\t}','if err := validate(e); err != nil {\n\t\treturn nil\n\t}','TestContextTotalReducerAndJournal','bad event not skipped atomically'),
 ('unknown_event_refusal','internal/history/history.go','func knownEvent(kind string) bool {','func knownEvent(kind string) bool {\n if kind == "future_event" { return true }','TestContextTotalReducerAndJournal','unknown event accepted'),
 ('unload_admission','internal/skills/skills.go','delete(r.tools, tool)','delete(r.tools, "nonexistent")','TestContextSurvivorsAndFrozenPrefix','declaration delta missing'),
 ('stub_address','internal/history/reduce.go','p.Ref = common.Ref{}','p.Parts[0].Ref = p.Ref','TestJobFileCapCursorAndRedaction','stub retained potentially reused job address'),
]
cases = [
 ('remove_calls_and_results','internal/history/reduce.go','if p.Type == "tool_call" || p.Type == "tool_result" {','if false {','TestContextLadderRecordedAndReplay','old tool traffic was not removed'),
]
results=[]
try:
 baseline=subprocess.run(['go','test','./ensemble','-run','TestContext|TestJobFileCapCursorAndRedaction','-count=1'],cwd=scratch,text=True,capture_output=True,timeout=45)
 results.append({'case':'baseline','exit':baseline.returncode,'output':baseline.stdout+baseline.stderr})
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,path,old,new,test,expected in cases:
  file=scratch/path;original=file.read_text()
  count=original.count(old)
  if count!=1:
   results.append({'case':name,'setup_error':f'expected one edit site, found {count}'});continue
  file.write_text(original.replace(old,new,1))
  run=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1'],cwd=scratch,text=True,capture_output=True,timeout=45)
  output=run.stdout+run.stderr
  results.append({'case':name,'test':test,'exit':run.returncode,'expected_failure':expected,'intended_kill':run.returncode!=0 and expected in output,'output':output})
  file.write_text(original)
  (base/'evidence/ch15/mutation-removal.json').write_text(json.dumps(results,indent=2))
finally:
 (base/'evidence/ch15/mutation-removal.json').write_text(json.dumps(results,indent=2))
 shutil.rmtree(scratch.parent)
print(json.dumps([{k:v for k,v in r.items() if k not in ['output']} for r in results],indent=2))
