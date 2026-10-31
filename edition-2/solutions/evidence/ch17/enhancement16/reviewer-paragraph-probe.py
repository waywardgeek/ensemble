from pathlib import Path
import subprocess,tempfile,shutil,json,os,hashlib,time
root=Path.cwd(); e=root/'solutions/edition-2/evidence/ch17/enhancement16'
results=[]
def run(cmd,cwd,env=None):
 start=time.monotonic();p=subprocess.run(cmd,cwd=cwd,env=env,capture_output=True,text=True,timeout=60)
 return dict(command=cmd,exit=p.returncode,seconds=round(time.monotonic()-start,3),output=p.stdout+p.stderr)
with tempfile.TemporaryDirectory(prefix='reviewer-ch16-') as td:
 temp=Path(td); source=temp/'source';shutil.copytree(root/'solutions/edition-2/ch16',source)
 helper=temp/'probe.go';helper.write_bytes((e/'focus-probe.go.txt').read_bytes())
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/gradeedition2/ch16_test.go'):str(helper)}}))
 harness=temp/'review-test';built=run(['go','test','-overlay',str(overlay),'-c','-o',str(harness),'./internal/gradeedition2'],root);assert built['exit']==0,built;results.append(dict(stage='fresh final grader build',**built))
 target=source/'internal/recall/index.go';original=target.read_text();old='strings.Split(part.content, "\\n\\n")';assert original.count(old)==1
 gui=temp/'gui'
 for name,body,want in [('baseline',original,0),('paragraph split removed',original.replace(old,'[]string{part.content}'),1),('restored',original,0)]:
  target.write_text(body);built=run(['go','build','-o',str(gui),'./cmd'],source/'gui');assert built['exit']==0,built
  env=dict(os.environ,RECALL_BIN=str(gui),RECALL_GROUP='chunk',RECALL_DATA='cr')
  result=run([str(harness),'-test.run','^TestChapter16Focus$','-test.v'],root,env);assert result['exit']==want,result
  if want:assert 'chunking-splits' in result['output'] and 'wrongparagraph' in result['output']
  results.append(dict(stage=name,source_sha256=hashlib.sha256(body.encode()).hexdigest(),binary_sha256=hashlib.sha256(gui.read_bytes()).hexdigest(),**result))
record={'coder_commit':'0df30bee','reviewed_source_sha256':json.loads((e/'final-source.json').read_text())['source_sha256'],'source_tree':subprocess.check_output(['git','rev-parse','HEAD:solutions/edition-2/ch16'],text=True).strip(),'method':'Fresh Go overlay test build against final grader; isolated frozen runtime copy, only paragraph iteration removed; real GUI/Agent/HTTP fake path; baseline and restored controls. No full grade/timeout/provider run.','results':results}
(e/'reviewer-paragraph-probe.json').write_text(json.dumps(record,indent=2)+'\n')
print([(r['stage'],r['exit'],r['seconds']) for r in results])
