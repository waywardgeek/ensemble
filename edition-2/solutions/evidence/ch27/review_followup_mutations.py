"""Reviewer-prompted welcome and parent-logger checks against the real server."""
import json,pathlib,shutil,subprocess,tempfile
HERE=pathlib.Path(__file__).resolve().parent;MAIN=HERE.parents[1]/'main'
cases=[('welcome-count','QueueDepth: len(q.Messages)','QueueDepth: 0','TestDeliveryAckResume','wrong welcome queue count'),('parent-diagnostics','c.Parent().Logf("peer %s: decrypt: %v", c.id, err)','// diagnostic omitted','TestConnectionDiagnosticReachesRoot','connection diagnostic did not reach root logger')]
results=[]
with tempfile.TemporaryDirectory(prefix='gateway26-review-mutations-') as temp:
 root=pathlib.Path(temp)/'main';shutil.copytree(MAIN,root,ignore=shutil.ignore_patterns('node_modules','cr'))
 path=root/'internal/gateway/server.go';original=path.read_text()
 for name,old,new,test,expected in cases:
  assert original.count(old)==1,name
  for phase in ['baseline','mutant','restored']:
   path.write_text(original.replace(old,new,1) if phase=='mutant' else original)
   p=subprocess.run(['go','test','-race','-count=1','-timeout=15s','./internal/gateway','-run','^'+test+'$'],cwd=root,capture_output=True,text=True,timeout=25)
   results.append({'name':name,'phase':phase,'exit':p.returncode,'expected':expected,'output':p.stdout+p.stderr})
(HERE/'review-followup-mutations.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))
