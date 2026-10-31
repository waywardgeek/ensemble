import pathlib,re,tempfile,subprocess,json,os,threading,http.server,hashlib,shutil,urllib.request
root=pathlib.Path(os.environ['CH18_ROOT']);out=root/'solutions/edition-2/evidence/ch19/enhancement18';chapter=root/'book/edition-2/chapter-18.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest();assert sha=='e057e264f91a20d4c8da1d2852f64e244283afcc3e455e8e4f67113e50d6a730'
blocks=re.findall(r'```([^\n]*)\n(.*?)```',chapter.read_text(),re.S);requests=[];auth=[];expected=''
answer='Worker 1: A 0-4, C 4-9, F 9-12, H 12-13. Worker 2: B 0-3, E 3-7, D 7-9, G 9-11. The chain A-C-F-H requires 13, so this schedule is optimal.'
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_GET(self):
  req=urllib.request.Request(os.environ['CH18_MODELS']+self.path,headers={'Authorization':self.headers.get('Authorization','')})
  with urllib.request.urlopen(req) as r:b=r.read()
  self.send_response(200);self.send_header('Content-Type','application/json');self.end_headers();self.wfile.write(b)
 def do_POST(self):
  q=json.loads(self.rfile.read(int(self.headers['Content-Length'])));requests.append(q);auth.append(self.headers.get('Authorization',''));i=len(requests)
  assert self.path=='/v1/responses' and i<=2
  if i==1:items=[dict(type='function_call',call_id='sample_schedule',name='read_file',namespace='ensemble',arguments=json.dumps(dict(path='schedule.txt')))]
  else:
   results=[x for x in q['input'] if x.get('type')=='function_call_output' and x.get('call_id')=='sample_schedule'];assert len(results)==1 and expected in results[0]['output']
   items=[dict(type='message',role='assistant',content=[dict(type='output_text',text=answer)])]
  self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
  for j,item in enumerate(items):
   e=dict(type='response.output_item.done',output_index=j,item=item);self.wfile.write(('event: '+e['type']+'\ndata: '+json.dumps(e)+'\n\n').encode())
  e=dict(type='response.completed',response=dict(status='completed',model=q['model'],output=[],usage=dict(input_tokens=300,output_tokens=30,input_tokens_details=dict(cached_tokens=0))))
  self.wfile.write(('event: response.completed\ndata: '+json.dumps(e)+'\n\n').encode());self.wfile.flush()
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start();work=None;receipts=[]
def run(command,cwd,env=None,input=None):
 r=subprocess.run(['bash','-c',command],cwd=cwd,env=env,input=input,capture_output=True,text=True,timeout=25);receipts.append(dict(command=command,exit=r.returncode,stdout=r.stdout,stderr=r.stderr));assert r.returncode==0,(command,r.stderr);return r.stdout
try:
 with tempfile.TemporaryDirectory(prefix='ch18-sample-types-') as types:
  p=pathlib.Path(types);(p/'go.mod').write_text('module sample\ngo 1.25\n');(p/'types.go').write_text('package sample\nimport("context";"time")\n'+blocks[0][1]);run('go test ./...',p)
 run(blocks[3][1],root/'solutions/edition-2/ch18')
 work=pathlib.Path(run(blocks[4][1]+'\npwd',root).strip());expected=(work/'schedule.txt').read_text()
 browser=work/'browser.py';browser.write_text('import sys,urllib.request\nwith urllib.request.urlopen(sys.argv[1],timeout=5) as r:r.read()\n')
 env=dict(os.environ,EN_HOME=str(work/'profile'),EN_PRIMARY_SKILL='base',OPENAI_OIDC_ISSUER=os.environ['CH18_ISSUER'],LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',LLM_API_KEY='fake-sample-api',EN_BROWSER_CMD='python3 '+str(browser))
 run(blocks[5][1],work,env)
 result=run(blocks[6][1],work,env,blocks[7][1]);assert answer in result
 assert len(requests)==2 and all(q['model']=='gpt-6.1-sol' and q['reasoning']==dict(effort='high',summary='detailed') and q['store'] is False and q['stream'] is True for q in requests)
 assert (work/'save.json').is_file()
 run(blocks[8][1],work,env)
 status=run('/tmp/ensemble-ch18 auth status',work,env);assert 'Signed in: false' in status and 'Billing: chatgpt' in status
 assert all(x.startswith('Bearer ') and x!='Bearer fake-sample-api' for x in auth)
 for r in receipts:
  for token in auth:r['stdout']=r['stdout'].replace(token,'[FAKE AUTHORIZATION]');r['stderr']=r['stderr'].replace(token,'[FAKE AUTHORIZATION]')
 (out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n');(out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n');(out/'samples.json').write_text(json.dumps(dict(manuscript_sha256=sha,fences=[dict(index=i,language=l,sha256=hashlib.sha256(b.encode()).hexdigest()) for i,(l,b) in enumerate(blocks)],result='PASS',requests=2,model='gpt-6.1-sol',actual_schedule_result=True,save=True,logout=True,limits='Local external issuer/provider peers only; no real sign-in, paid request, summary or cache grant. Peer supplies known answer; mechanics, not scheduling intelligence. Grader command fences exercised separately by accepted-green run. Conceptual vocabulary compiles as standalone types.'),indent=2)+'\n');print('PASS exact samples',sha)
finally:
 server.shutdown();server.server_close()
 if work:shutil.rmtree(work)
