"""Focused defects in a disposable copy; never edit the live source or grader."""
import json
import pathlib
import shutil
import subprocess
import tempfile

repo = pathlib.Path.cwd()
evidence = repo/'solutions/edition-2/evidence/ch22'
scratch = pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch22-mutants-'))/'main'
shutil.copytree(repo/'solutions/edition-2/main', scratch, ignore=shutil.ignore_patterns('node_modules','.git'))
# One changed behavior per row. Assertions exercise the real runtime and its
# external fake provider, not replacement Engine/Agent implementations.
cases = [
 ('cli-control', 'chat.go', 'a.SetModel(strings.TrimSpace(strings.TrimPrefix(text, "/model ")))', 'a.SetModel(a.Config().Model)', 'root', 'TestChatModelSelection', 'CLI model control did not reach provider'),
 ('status-declaration', 'internal/tools/tools.go', 'Name: "agent_status"', 'Name: "missing_status"', 'ensemble', 'TestAgentStatusToolLoop', 'status absent from declarations'),
 ('status-last', 'internal/tools/status.go', 'meter.Input, meter.Output, meter.CacheRead, meter.CacheWrite,', 'meter.Total.Input, meter.Total.Output, meter.Total.CacheRead, meter.Total.CacheWrite,', 'ensemble', 'TestAgentStatusModelAccounting', 'status missing'),
 ('status-unknown', 'internal/tools/status.go', 'if meter.Priced {', 'if true {', 'ensemble', 'TestAgentStatusModelAccounting', 'unknown reported as measured dollars'),
 ('status-dispatch', 'internal/tools/status.go', 'return fmt.Sprintf("model: %s', 'return fmt.Sprintf("wrong model: %s', 'ensemble', 'TestAgentStatusToolLoop', 'actual status tool result'),
 ('historical-prices', 'internal/engine/usage.go', 'LookupModel(key.model)', 'LookupModel(e.Agent().Config().Model)', 'ensemble', 'TestAgentStatusModelAccounting', 'selection changed historical counts or cost'),
 ('cache-denominator', 'internal/engine/usage.go', 'out.Total.Input + out.Total.CacheWrite + out.Total.CacheRead', 'out.Total.Input + out.Total.CacheRead', 'ensemble', 'TestAgentStatusModelAccounting', 'status missing'),
 ('model-noop', 'internal/agent/model.go', 'r.owner.config.Model = selection.model', '// Selection is accidentally ignored.', 'ensemble', 'TestAgentStatusModelAccounting', 'selection retained stale identity or changed provider'),
 ('stale-identity', 'internal/agent/model.go', 'r.owner.config.ResolvedModel = ""', '// Stale resolved identity retained.', 'ensemble', 'TestAgentStatusModelAccounting', 'selection retained stale identity or changed provider'),
 ('busy-selection', 'internal/agent/model.go', 'if r.active != nil {', 'if false {', 'ensemble', 'TestModelSelectionBusyAndInvalid', 'busy selection altered'),
 ('empty-selection', 'internal/agent/model.go', 'if model == "" {', 'if false {', 'ensemble', 'TestModelSelectionBusyAndInvalid', 'empty model accepted'),
 ('gui-owner', 'gui/settings.go', 's.agent.SetModel(model)', 's.agent.SetModel(s.agent.Config().Model)', 'gui', 'TestModelSelectionWireAndRequest', 'selection not broadcast'),
 ('gui-persist-model', 'gui/settings.go', 'Model *string `json:"model,omitempty"`', 'Model *string `json:"unused,omitempty"`', 'gui', 'TestModelSelectionWireAndRequest', 'session model persisted'),
 ('gui-mixed-patch', 'gui/settings.go', 'if len(fields) != 1 {', 'if false {', 'gui', 'TestModelSelectionWireAndRequest', 'waiting for wire event'),
]
results=[]
try:
 for name,rel,old,new,module,test,expected in cases:
  path=scratch/rel; original=path.read_text(); assert original.count(old)==1,(name,'replacement count')
  path.write_text(original.replace(old,new))
  cwd=scratch/'gui' if module=='gui' else scratch
  command=['go','test','-count=1','-timeout=15s','-run','^'+test+'$','.' if module in ('gui','root') else './ensemble']
  try:
   result=subprocess.run(command,cwd=cwd,capture_output=True,text=True,timeout=25)
   output=result.stdout+result.stderr
   killed=result.returncode != 0 and expected in output and '[build failed]' not in output
   (evidence/(name+'.txt')).write_text(output)
   results.append(dict(name=name,file=rel,old=old,new=new,command=command,exit=result.returncode,killed=killed,expected_failure=expected))
   print(name,'killed' if killed else 'SURVIVED OR UNEXPECTED',flush=True)
  finally:path.write_text(original)
 # Restoration gets the same focused tests, including the GUI path.
 for module,test in [('', 'TestAgentStatus|TestModelSelection'),('gui','TestModelSelectionWireAndRequest')]:
  result=subprocess.run(['go','test','-race','-count=1','-timeout=25s','-run',test,'./ensemble' if not module else '.'],cwd=scratch/module,capture_output=True,text=True,timeout=40)
  (evidence/('restored-'+(module or 'core')+'.txt')).write_text(result.stdout+result.stderr)
  assert result.returncode==0
finally:
 (evidence/'mutations.json').write_text(json.dumps(results,indent=2)+'\n')
 shutil.rmtree(scratch.parent)
