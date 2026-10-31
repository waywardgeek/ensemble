"""Narrow persistence and capability controls, using the real gateway tests."""
import json,pathlib,shutil,subprocess,tempfile
HERE=pathlib.Path(__file__).resolve().parent;MAIN=HERE.parents[1]/'main'
cases=[
 ('send-capability','internal/gateway/admin.go','if !s.capable(from, "send-message") {','if false {','missing send capability accepted'),
 ('declaration-capability','internal/gateway/admin.go','if !s.capable(id, "declare-contactable") {','if false {','missing declaration capability accepted'),
 ('schema-required-array','internal/gateway/admin.go','required = []string{}','required = nil','MCP required must be an array'),
 ('queue-persistence','internal/gateway/server.go','if err := writeJSON(s.queuePath(id), q); err != nil {','if err := writeJSON(s.queuePath(id), &queue{Next:q.Next,Ack:q.Next}); err != nil {','delivery timeout'),
 ('contactability-persistence','internal/gateway/server.go','for name, g := range s.state.Groups {','s.state.Contactable=map[string][]string{}\n for name, g := range s.state.Groups {','sender groups are not accepted by recipient'),
 ('membership-persistence','internal/gateway/server.go','for name, g := range s.state.Groups {','s.state.Memberships=map[string][]string{}\n for name, g := range s.state.Groups {','sender groups are not accepted by recipient'),
 ('runtime-group-persistence','internal/gateway/server.go','for name, g := range s.state.Groups {','s.state.Groups=map[string]Group{}\n for name, g := range s.state.Groups {','sender groups are not accepted by recipient'),
 ('group-delete','internal/gateway/admin.go','delete(s.state.Groups, a.Name)','return "group deleted",nil\n delete(s.state.Groups, a.Name)','deleted runtime group retained routing'),
]
results=[]
with tempfile.TemporaryDirectory(prefix='gateway26-extra-mutations-') as temp:
 root=pathlib.Path(temp)/'main';shutil.copytree(MAIN,root,ignore=shutil.ignore_patterns('node_modules','.git','cr'))
 def run():
  p=subprocess.run(['go','test','-count=1','-timeout=12s','./internal/gateway','-run','^TestChildrenAndPersistedPolicy$'],cwd=root,text=True,capture_output=True,timeout=20)
  return p.returncode,p.stdout+p.stderr
 code,out=run();results.append({'baseline':True,'exit':code,'output':out})
 if code:raise SystemExit('baseline failed')
 for name,path,old,new,expected in cases:
  target=root/path;original=target.read_text();assert original.count(old)==1,(name,original.count(old))
  target.write_text(original.replace(old,new,1))
  try:
   code,out=run();results.append({'name':name,'old':old,'new':new,'exit':code,'expected':expected,'intended_failure':code!=0 and expected in out,'output':out})
  finally:target.write_text(original)
 code,out=run();results.append({'restored':True,'exit':code,'output':out})
(HERE/'mutation-extra.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps([{k:v for k,v in r.items() if k not in ['output','old','new']} for r in results],indent=2))
