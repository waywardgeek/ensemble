"""Focused source defects in disposable copies; the real Gateway/client still run."""
import difflib,hashlib,json,os,pathlib,shutil,subprocess,tempfile,time
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch27/enhancement26';original=root/'solutions/edition-2/ch26'
cases=[
 ('queue-count-zero','internal/gateway/server.go','QueueDepth: len(q.Messages)','QueueDepth: 0','welcome'),
 ('body-partial','internal/gateway/admin.go',' + message.Content',' + strings.ReplaceAll(message.Content, "Original complete body 26", "body 26")','authenticated envelope'),
 ('claimed-sender','internal/gateway/admin.go','UserID: from,','UserID: "operator",','authenticated envelope'),
 ('receipt-before-handler','gateway/client.go','\t\t\tc.owner.OnDeliver(*frame.Envelope)\n','', 'welcome'),
 ('queue-not-saved','internal/gateway/server.go','if err := writeJSON(s.queuePath(id), q); err != nil {','if err := func() error { _ = id; _ = q; return nil }(); err != nil {','welcome'),
 ('private-runtime-import','gateway/noise.go','import (','import (\n _ "ensemble/internal/common"','public Gateway imports private runtime'),
 ('successful-shutdown-exits17','cmd/gateway/main.go','\n}\nfunc run() error {','\n if flag.Lookup("print-key").Value.String() != "true" { os.Exit(17) }\n}\nfunc run() error {','Gateway shutdown'),
]
records=[]
with tempfile.TemporaryDirectory(prefix='gateway26-defects-') as temp:
 copy=pathlib.Path(temp)/'submission';shutil.copytree(original,copy)
 def run(name,expected=None):
  env=dict(os.environ,GATEWAY_SUBMISSION=str(copy));t=time.monotonic();r=subprocess.run(['go','test','./internal/gradeedition2','-run','^TestGateway26$','-count=1','-v'],env=env,capture_output=True,text=True,timeout=45)
  raw=r.stdout+r.stderr;(out/(name+'.txt')).write_text(raw)
  row={'name':name,'exit':r.returncode,'seconds':round(time.monotonic()-t,3),'expected':expected,'intended_failure':r.returncode!=0 and expected in raw if expected else None}
  records.append(row);print(row,flush=True)
  if expected is None:assert r.returncode==0,raw
  return row
 run('baseline-final')
 for name,path,old,new,expected in cases:
  target=copy/path;s=target.read_text();assert s.count(old)==1,(name,s.count(old));changed=s.replace(old,new)
  if name=='body-partial':changed=changed.replace('"slices"','"slices"\n"strings"')
  if name=='receipt-before-handler':changed=changed.replace('c.ack.Store(frame.Envelope.Seq)','c.ack.Store(frame.Envelope.Seq)\n c.owner.OnDeliver(*frame.Envelope)')
  target.write_text(changed);subprocess.run(['gofmt','-w',str(target)],check=True)
  (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(s.splitlines(True),target.read_text().splitlines(True),fromfile=path,tofile=path)))
  run(name,expected);target.write_text(s)
 run('restored-final')
source={p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in ['cmd/grade-edition2/main.go','internal/gradeedition2/ch26.go','internal/gradeedition2/ch26_probe.go.txt','internal/gradeedition2/ch26_test.go']}
(out/'sensitivity.json').write_text(json.dumps({'sources':source,'cases':records,'limits':'Seven new targeted changes, alongside separately retained accepted source/mutation evidence. Coupled assertions do not establish independent security proofs.'},indent=2)+'\n')
