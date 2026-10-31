"""Focused real-GUI defect controls; mutations live only in disposable copies."""
import pathlib,subprocess,tempfile,shutil,os,json,time,difflib,hashlib
root=pathlib.Path(__file__).resolve().parents[5];out=pathlib.Path(__file__).parent
source=root/'solutions/edition-2/ch23';work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-management23-'))
cases=[
 ('missing-thinking','internal/agent/activity.go','activity := common.Activity{Phase: phase','if phase == common.ActivityThinking { phase = common.ActivityIdle }; activity := common.Activity{Phase: phase','activity-observed'),
 ('status-without-activity','internal/tools/management.go','return fleetJSON(a.Fleet().Status())','return fleetJSON(map[string]string{"agent_id": a.Spec().ID, "state":"submitted"})','activity-observed'),
 ('foreign-lookup','internal/tools/management.go','a, err := owner.Agent().Fleet().Find(args.ID)','var a common.Agent; var err error; if p:=owner.Agent().ParentAgent(); p!=nil { a,err=p.Fleet().Find(args.ID) } else { a,err=owner.Agent().Fleet().Find(args.ID) }','ownership-enforced'),
 ('distinct-missing-disclosure','internal/fleet/management.go','return nil, fmt.Errorf("unknown agent %q", id)','if id=="absent23" { return nil, fmt.Errorf("no such agent %q",id) }; return nil, fmt.Errorf("unknown agent %q", id)','ownership-enforced'),
 ('wrong-watched-identity','internal/tools/management.go','return owner.Agent().Fleet().WaitChange(ctx, args.IDs, args.Pattern, time.Duration(args.Stuck*float64(time.Second)))','_, err := owner.Agent().Fleet().WaitChange(ctx, args.IDs, args.Pattern, time.Duration(args.Stuck*float64(time.Second))); return `{"agent_id":"wrong-child","state":"submitted"}`,err','wait-change-wakes'),
 ('ignore-created-model','ensemble/registry.go','cfg.Model = entry.Model','cfg.Model = e.defaults.Model','ensemble-persists'),
 ('ignore-created-skill','ensemble/registry.go','spec.ID, spec.PrimarySkill = entry.Name, entry.Skill','spec.ID, spec.PrimarySkill = entry.Name, e.defaultSpec.PrimarySkill','ensemble-persists'),
 ('skip-peer-history-load','ensemble/registry.go','if err = a.Load(); err == nil {','if err = nil; err == nil {','ensemble-persists'),
 ('retain-deleted-registry','ensemble/registry.go','if err := e.saveRegistry(entries); err != nil {','if err := e.saveRegistry(e.registry); err != nil {','ensemble-persists'),
 ('omit-nested-tree','ensemble/registry.go','n.Children = append(n.Children, treeNode(child, false, false))','_ = child','gui-tree-message'),
 ('forget-completed-turn','internal/agent/dispatch.go','err := r.append(common.Event{Type: "turn_ended"})','var err error','gui-tree-message'),
 ('always-idle-replay','gui/gui.go','state := s.agent.History().Context().Turn','state := "idle"','gui-tree-message'),
]
def command(cmd,cwd,env=None,timeout=35):
 start=time.monotonic();r=subprocess.run(cmd,cwd=cwd,env=env,capture_output=True,text=True,timeout=timeout);return r,time.monotonic()-start
results=[]
try:
 test=work/'check';r,_=command(['go','test','-c','-race','-o',str(test),'./internal/gradeedition2'],root);assert r.returncode==0,r.stderr
 tree=work/'source';shutil.copytree(source,tree)
 def probe(label):
  binary=work/'gui';r,_=command(['go','build','-o',str(binary),'./cmd'],tree/'gui');assert r.returncode==0,r.stderr
  env=dict(os.environ,EDITION2_MANAGEMENT_BIN=str(binary));r,wall=command([str(test),'-test.run','^TestManagement23$','-test.v'],root,env,30)
  (out/(label+'.txt')).write_text(r.stdout+r.stderr)
  rows=next((json.loads(line) for line in r.stdout.splitlines() if line.startswith('[{"id"')),[])
  return dict(exit=r.returncode,wall_seconds=wall,failed={c['id']:c.get('details',[]) for c in rows if not c['passed']})
 baseline=probe('sensitivity-baseline');assert baseline['exit']==0,baseline
 for name,path,before,after,intended in cases:
  target=tree/path;original=target.read_text();assert before in original,(name,before)
  changed=original.replace(before,after,1)
  if name=='ignore-created-model':changed=changed.replace('spec.Model = entry.Model','spec.Model = e.defaults.Model',1)
  target.write_text(changed)
  (out/('mutation-'+name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile=path,tofile=path)))
  try:result=probe('mutation-'+name)
  finally:target.write_text(original)
  result.update(name=name,intended=intended,caught=intended in result['failed']);results.append(result)
  (out/'sensitivity.json').write_text(json.dumps(dict(baseline=baseline,cases=results,grader_sources={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/'internal/gradeedition2').glob('ch23*.go')}),indent=2)+'\n')
  print(name,result,flush=True)
 restored=probe('sensitivity-restored');assert restored['exit']==0,restored
 (out/'sensitivity-restored.json').write_text(json.dumps(restored,indent=2)+'\n')
finally:shutil.rmtree(work)
