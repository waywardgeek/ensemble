"""Focused, one-at-a-time defects in isolated student source; no grader edits."""
from pathlib import Path
import tempfile,shutil,subprocess,json,time,sys
root=Path(__file__).resolve().parents[2]/'main'
out=Path(__file__).resolve().parent
cases=[
 ('factory_ownership','ensemble/fleet.go','return e.factory(parent, AgentSpec(spec))','return e.NewAgent(parent.Config(), AgentSpec(spec))','TestFleetRecursive','recursive fleet:'),
 ('child_save_isolation','internal/fleet/fleet.go','spec.SavePath = filepath.Join(spec.DataDir, "save.json")','spec.SavePath = filepath.Join(m.parent.Config().DataDir, "save.json")','TestFleetRecursive','child save'),
 ('child_restore','internal/fleet/fleet.go','if err = a.Load(); err == nil {','if err = nil; err == nil {','TestFleetLifecycle','same-ID child lost previous dialogue'),
 ('live_status','internal/fleet/fleet.go','out = append(out, s)','if false { out = append(out, s) }','TestFleetLifecycle','live child invisible'),
 ('shutdown_removal','internal/fleet/fleet.go','if m.children[id] == c {','if false && m.children[id] == c {','TestFleetLifecycle','shutdown child retained'),
 ('shutdown_cascade','internal/fleet/fleet.go','for _, id := range ids {\n\t\terrs = append','for _, id := range ids[:0] {\n\t\terrs = append','TestFleetRecursive','cascade retained live children'),
 ('join_fanin','internal/fleet/fleet.go','done = done && s.State == "submitted"','done = true','TestFleetRecursive','root join lost'),
 ('join_question_yield','internal/fleet/fleet.go','if len(m.pending) > 0 {','if false && len(m.pending) > 0 {','TestFleetRecursive','join handler never returned'),
 ('reply_sender','internal/fleet/fleet.go','From: from,','From: from[:0],','TestFleetRecursive','addressed reply=false'),
 ('submit_active_children','internal/fleet/fleet.go','if ids := m.active(); len(ids) > 0 {','if ids := m.active(); len(ids) < 0 {','TestFleetLifecycle','submit accepted active child'),
 ('wake_bound','internal/fleet/fleet.go','if m.wakes >= 3 {','if m.wakes >= 4 {','TestFleetIdleBound','three-wake bound'),
 ('wake_reset','internal/fleet/fleet.go','m.wakes = 0','m.wakes += 0','TestFleetLifecycle','wake 1 after submit'),
 ('failed_turn_retry','internal/agent/actor.go','if err == nil && r.owner.Ensemble().FleetEnabled() {','if r.owner.Ensemble().FleetEnabled() {','TestFleetIdleBound','failed turn retried'),
 ('message_wake','internal/jobs/jobs.go','case <-limits.Wake:','case <-(<-chan struct{})(nil):','TestFleetMessage','message waited for command or timer'),
 ('late_tool_result','internal/jobs/jobs.go','return header + text, j.err','return header + "lost output", j.err','TestFleetMessage','background result lost'),
 ('join_question_direct','internal/fleet/fleet.go','if len(m.pending) > 0 {','if false && len(m.pending) > 0 {','TestFleetJoinQuestionYield','join did not yield pending question before child submission'),
 ('question_collision','internal/fleet/fleet.go','msg.ID = fmt.Sprintf("%s-%d", m.parent.Spec().ID, m.sequence)','msg.ID = fmt.Sprintf("%s-%d", msg.From, 1)','TestFleetSameNameQuestionCorrelation','pending question correlation incomplete'),
 ('question_recipient','internal/fleet/fleet.go','recipientMatches := messageID != "" || (agentID == "") == humanQuestion','recipientMatches := true || humanQuestion','TestFleetSameNameQuestionCorrelation','single child question was ambiguous'),
 ('escalated_status','internal/fleet/fleet.go','if m.failure != "" {','if false && m.failure != "" {','TestFleetEscalatedChildStatus','exhausted child still claims running'),
 ('human_sink','main.go','app.SetUserSink("user", terminalUser{})','// removed human sink','TestFleetCLI','human sink was unreachable'),
]
results=json.loads((out/'mutations.json').read_text()) if len(sys.argv)>1 else []
for name,file,before,after,test,expected in cases:
 if len(sys.argv)>1 and name not in sys.argv[1:]:continue
 with tempfile.TemporaryDirectory(prefix='ensemble-ch25-mutant-') as tmp:
  scratch=Path(tmp)/'main'
  shutil.copytree(root,scratch,ignore=shutil.ignore_patterns('gui','ch05','ch06','ch07','node_modules'))
  target=scratch/file
  content=target.read_text()
  if content.count(before)!=1:raise RuntimeError((name,'anchor',content.count(before)))
  target.write_text(content.replace(before,after))
  package='.' if test=='TestFleetCLI' else './ensemble'
  start=time.monotonic()
  try:
   p=subprocess.run(['go','test','-count=1','-timeout','20s',package,'-run',test],cwd=scratch,text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,timeout=25)
   output=p.stdout;code=p.returncode
  except subprocess.TimeoutExpired as e:output=(e.stdout or b'').decode() if isinstance(e.stdout,bytes) else e.stdout or '';code='launcher_timeout'
  (out/('mutation-'+name+'.txt')).write_text(output)
  outcome='intended_failure' if code==1 and expected in output and '[build failed]' not in output and 'panic: test timed out' not in output else 'survived' if code==0 else 'inspect'
  row=dict(name=name,file=file,mutation=after,test=test,exit=code,outcome=outcome,expected=expected,seconds=round(time.monotonic()-start,3))
  results.append(row);(out/'mutations.json').write_text(json.dumps(results,indent=2)+'\n')
  print(name,outcome,flush=True)
