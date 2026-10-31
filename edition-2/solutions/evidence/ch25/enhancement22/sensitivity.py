"""Finite targeted source defects; each runs the real submitted clients in a disposable copy."""
import difflib,json,os,pathlib,shutil,subprocess,tempfile,time
root=pathlib.Path(__file__).resolve().parents[5];out=pathlib.Path(__file__).parent;source=root/'solutions/edition-2/ch22'
cases=[
 ('child-save-shared','internal/fleet/fleet.go','spec.SavePath = filepath.Join(spec.DataDir, "save.json")','spec.SavePath = filepath.Join(m.parent.Config().DataDir, "save.json")','Fleet','distinct valid parent/child'),
 ('fanin-truncated','internal/fleet/fleet.go','return encoded(results), nil','return encoded(results[:1]), nil','Fleet','one correlated join did not contain both'),
 ('live-child-hidden','internal/fleet/fleet.go','out = append(out, s)','if s.State != "running" { out = append(out, s) }','Fleet','status omitted the held live child'),
 ('reply-sender-lost','internal/fleet/fleet.go','From: from,','From: from[:0],','Fleet','ask lost middle'),
 ('four-idle-wakes','internal/fleet/fleet.go','if m.wakes >= 3 {','if m.wakes >= 4 {','Fleet','want its initial response plus exactly three wakes'),
 ('late-output-lost','internal/jobs/jobs.go','return header + text, j.err','return header + "lost output", j.err','Fleet','lost completed output'),
 ('message-no-wake','internal/jobs/jobs.go','case <-limits.Wake:','case <-(<-chan struct{})(nil):','Fleet','message-wake session'),
 ('embedded-directory-listing','gui/gui.go','assets := http.FS(Assets())','assets := http.FileSystem(http.Dir("."))','Reuse','marker "<script" present=false'),
 ('override-ignored','gui/gui.go','if directory != "" {','if false {','Reuse','EXPLICIT_DISK_22'),
 ('public-final-lost','gui/gui.go','case "response_ended":','case "no-response_ended":','Reuse','actual Agent reply did not reach'),
 ('relay-caller-id-lost','gui/mcp.go','response["id"] = original','response["id"] = request["id"]','Reuse','lost caller/result correlation'),
 ('headless-websocket-edge','internal/mcp/mcp.go','import (','import (\n _ "github.com/gorilla/websocket"','Reuse','WebSocket=true wanted=false'),
]
results=[]
def invoke(cmd,cwd,env,path):
 start=time.monotonic();p=subprocess.run(cmd,cwd=cwd,env=env,capture_output=True,text=True,timeout=70);text=p.stdout+p.stderr;path.write_text(text);return p.returncode,text,round(time.monotonic()-start,3)
with tempfile.TemporaryDirectory(prefix='ch22-defects-') as tmp:
 work=pathlib.Path(tmp)/'core';shutil.copytree(source,work,ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 env=dict(os.environ,EDITION2_FLEET_BIN=str(pathlib.Path(tmp)/'ensemble'),EDITION2_REUSE_CORE=str(work))
 def build():
  p=subprocess.run(['go','build','-o',env['EDITION2_FLEET_BIN'],'.'],cwd=work,capture_output=True,text=True);assert p.returncode==0,p.stderr
 build()
 for kind in ['Fleet','Reuse']:
  code,text,seconds=invoke(['go','test','./internal/gradeedition2','-run','^Test'+kind+'22$','-count=1','-timeout','60s','-v'],root,env,out/('sensitivity-'+kind+'-baseline.txt'));assert code==0,text
 for name,path,before,after,kind,expected in cases:
  file=work/path;original=file.read_text();assert original.count(before)==1,(name,original.count(before));changed=original.replace(before,after);file.write_text(changed)
  (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile=path,tofile=path)))
  mods={}
  if name=='headless-websocket-edge':
   for f in ['go.mod','go.sum']:mods[f]=(work/f).read_text()
   (work/'go.mod').write_text(mods['go.mod']+'\nrequire github.com/gorilla/websocket v1.5.3\n');(work/'go.sum').write_text(mods['go.sum']+(work/'gui/go.sum').read_text())
  if kind=='Fleet':build()
  cmd=['go','test','./internal/gradeedition2','-run','^Test'+kind+'22$','-count=1','-timeout','60s','-v']
  code,text,seconds=invoke(cmd,root,env,out/('mutation-'+name+'.txt'))
  row=dict(name=name,file=path,command=cmd,exit=code,seconds=seconds,intended=expected,detected=code!=0 and expected in text and '[build failed]' not in text,failures=[l for l in text.splitlines() if 'ch22_test.go:' in l]);results.append(row);(out/'sensitivity.json').write_text(json.dumps(results,indent=2)+'\n');print(name,row['detected'],flush=True)
  file.write_text(original)
  for f,s in mods.items():(work/f).write_text(s)
 build()
 for kind in ['Fleet','Reuse']:
  code,text,seconds=invoke(['go','test','./internal/gradeedition2','-run','^Test'+kind+'22$','-count=1','-timeout','60s','-v'],root,env,out/('sensitivity-'+kind+'-restored.txt'));assert code==0,text
