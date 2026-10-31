import pathlib,re,tempfile,shutil,subprocess,json,os,threading,http.server,time,hashlib,queue
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch17/enhancement16';chapter=root/'book/edition-2/chapter-16.md';text=chapter.read_text();blocks=re.findall(r'```([^\n]*)\n(.*?)```',text,re.S)
sha=hashlib.sha256(chapter.read_bytes()).hexdigest();requests=[];main=0;judge=0
answer='QZ-731: Morgan owns the rehearsal on October 23. Automatic retry was rejected because the remote transaction may already have committed. Ledger placement remains unresolved between the application database and a separate control database. Retain rollback artifacts for nine days.'
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  global main,judge
  q=json.loads(self.rfile.read(int(self.headers['Content-Length'])));requests.append(q)
  if not q.get('tools'):
   judge+=1
   s='\n'.join(m['content'] if isinstance(m['content'],str) else '\n'.join(p.get('text','') for p in m['content']) for m in q['messages'])
   matches=list(re.finditer(r'\[(\d+)\] [^\n]*\n(.*?)(?=\n\[\d+\]|\Z)',s,re.S));chosen=[int(m.group(1)) for m in matches if 'QZ-731' in m.group(2)];assert len(chosen)==1,(chosen,s);reply=json.dumps(chosen)
  else:main+=1;reply=answer+f' Sample turn {main} complete.'
  b=json.dumps({'id':'sample-'+str(len(requests)),'type':'message','role':'assistant','model':q['model'],'content':[{'type':'text','text':reply}],'stop_reason':'end_turn','usage':{'input_tokens':100,'output_tokens':20}}).encode();self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(b)));self.end_headers();self.wfile.write(b)
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
try:
 with tempfile.TemporaryDirectory(prefix='ch16-samples-') as tmp:
  source=pathlib.Path(tmp)/'source';shutil.copytree(root/'solutions/edition-2/ch16',source)
  snippet=source/'snippet';snippet.mkdir();(snippet/'types.go').write_text('package snippet\n'+blocks[0][1]);r=subprocess.run(['go','test','./snippet'],cwd=source,capture_output=True,text=True);(out/'samples-go.txt').write_text(r.stdout+r.stderr);assert r.returncode==0,r.stderr
  r=subprocess.run(['bash','-c',blocks[3][1]+'\npwd'],cwd=source,capture_output=True,text=True);assert r.returncode==0,r.stderr;work=pathlib.Path(r.stdout.strip());assert work.is_dir()
  try:
   (work/'skills/base/SKILL.md').write_text(blocks[4][1]);r=subprocess.run(['bash','-c',blocks[5][1]],cwd=work,capture_output=True,text=True);assert r.returncode==0,r.stderr
   env=dict(os.environ,LLM_API_KEY='fake-key',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',EN_DISABLE_STREAMING='1')
   stderr=open(work/'stderr.txt','w');p=subprocess.Popen(['bash','-c','exec env '+blocks[6][1]],cwd=work,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=stderr,text=True,bufsize=1)
   q=queue.Queue();lines=[]
   def read():
    for line in p.stdout:lines.append(line);q.put(line)
   threading.Thread(target=read,daemon=True).start()
   try:
    for i in range(2):
     p.stdin.write(blocks[7+i][1].strip()+'\n');p.stdin.flush();deadline=time.monotonic()+10
     while True:
      line=q.get(timeout=max(.01,deadline-time.monotonic()))
      if f'Sample turn {i+1} complete.' in line:break
    p.stdin.close();p.wait(timeout=5);assert p.returncode==0
   finally:
    if p.poll() is None:p.kill();p.wait()
    stderr.close()
   assert main==2 and judge==1,(main,judge)
   assert all(x['model']=='claude-sonnet-5' for x in requests)
   for x in [x for x in requests if x.get('tools')]:
    s='\n'.join(m['content'] if isinstance(m['content'],str) else '\n'.join(p.get('text','') for p in m['content']) for m in x['messages']);assert s.count('QZ-731')>=1 and 'cafeteria' not in s
   events=[json.loads(x) for x in (work/'ch02.jsonl').read_text().splitlines() if x.strip()];recalls=[e for e in events if e.get('type')=='recall_attached'];assert len(recalls)==1
   (out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n');(out/'samples-events.json').write_text(json.dumps(recalls,indent=2)+'\n');(out/'samples-stdout.txt').write_text(''.join(lines))
   (out/'samples.json').write_text(json.dumps(dict(manuscript_sha256=sha,fences=[dict(language=l,sha256=hashlib.sha256(b.encode()).hexdigest()) for l,b in blocks],source='Exact isolated solutions/edition-2/ch16',build=blocks[3][1],corpus=blocks[5][1],launch=blocks[6][1],go='Exact vocabulary fence compiled',main_requests=main,judge_requests=judge,exit=p.returncode,recorded_recall=1,model='claude-sonnet-5',limit='External local HTTP fake verifies plumbing and exact commands; scripted answers do not establish semantic judge or model quality.'),indent=2)+'\n')
  finally:shutil.rmtree(work)
finally:server.shutdown()
print('Exact samples PASS',sha,main,judge)
