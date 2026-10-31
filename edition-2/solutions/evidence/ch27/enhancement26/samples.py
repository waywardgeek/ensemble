"""Exact published commands, three real processes, one scripted external provider."""
import hashlib,http.server,json,os,pathlib,re,shutil,signal,socket,subprocess,tempfile,threading,time
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch27/enhancement26';core=root/'solutions/edition-2/ch26'
chapter=root/'book/edition-2/chapter-26.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha=='55f7c95a2391c85e45efd6e8e2651d270e96fe1220c3ccd3c24b4731c09e7ab7'
blocks=re.findall(r'^```([^\n]*)\n(.*?)^```',chapter.read_text(),re.M|re.S);assert len(blocks)==14
requests=[];receipts=[];children=[];errors=[];work=None;started=time.monotonic()
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  try:
   q=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
   assert self.path=='/responses' and self.headers.get('Authorization')=='Bearer SAMPLE_FAKE_TOKEN'
   assert self.headers.get('ChatGPT-Account-ID')=='sample-account'
   assert q['model']=='gpt-6-astra' and q['reasoning']['effort']=='high'
   requests.append(q);n=len(requests);assert n<=4
   item={'type':'message','role':'assistant','content':[{'type':'output_text','text':f'LOCAL_GATEWAY_ANSWER_{n}'}]}
   response={'status':'completed','model':q['model'],'usage':{'input_tokens':100+n,'output_tokens':5,'input_tokens_details':{'cached_tokens':0}},'output':[item]}
   self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
   self.wfile.write(('event: response.completed\ndata: '+json.dumps({'type':'response.completed','response':response})+'\n\n').encode());self.wfile.flush()
  except Exception as e:errors.append(repr(e));raise
provider=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=provider.serve_forever,daemon=True).start()
env={k:v for k,v in os.environ.items() if k in ['HOME','PATH','TMPDIR','USER','LANG','SHELL']}
def run(command,cwd=root):
 r=subprocess.run(['bash','-c',command],cwd=cwd,env=env,capture_output=True,text=True,timeout=40)
 receipts.append({'command':command,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr});assert r.returncode==0,receipts[-1];return r.stdout

def launch(name,command):
 so=work/(name+'.out');se=work/(name+'.err')
 with so.open('w') as a,se.open('w') as b:p=subprocess.Popen(['bash','-c',command],env=env,cwd=work,stdout=a,stderr=b,text=True)
 children.append((name,p,so,se));return p

def ready(p,path,word):
 deadline=time.monotonic()+5
 while time.monotonic()<deadline:
  assert p.poll() is None,('child exited before ready',p.returncode,path.read_text())
  if word in path.read_text():return
  time.sleep(.01)
 raise AssertionError(('readiness timeout',path.read_text()))
try:
 # Execute the actual build/key setup. Extra output does not alter the commands.
 setup=run(blocks[7][1]);work=pathlib.Path(setup.splitlines()[-1]);env['CH26_RUN']=str(work)
 public=setup.splitlines()[:2];assert all(k.startswith('x25519:') for k in public)
 with socket.socket() as sock:sock.bind(('127.0.0.1',0));port=sock.getsockname()[1]
 cfg=json.loads(blocks[8][1]);cfg['gateway']['listen']=f'127.0.0.1:{port}';cfg['gateway']['data_dir']=str(work/'router-data')
 cfg['agents']['ensemble']['public_key']=public[0];cfg['agents']['interviewer']['public_key']=public[1]
 (work/'gateway.json').write_text(json.dumps(cfg))
 env['GATEWAY_KEY']=run(blocks[9][1]).strip()
 router=launch('router',blocks[10][1]);ready(router,work/'router.out','x25519:')
 deadline=time.monotonic()+5
 while time.monotonic()<deadline:
  assert router.poll() is None,'router bind failed'
  owned=subprocess.run(['lsof','-nP','-a','-p',str(router.pid),'-iTCP:'+str(port),'-sTCP:LISTEN'],capture_output=True,text=True)
  if owned.returncode==0:break
  time.sleep(.02)
 else:raise AssertionError('new router did not own selected listener')
 receipts.append({'readiness_pid':router.pid,'lsof':owned.stdout})
 auth=work/'fake-codex-auth.json';auth.write_text(json.dumps({'auth_mode':'chatgpt','tokens':{'access_token':'SAMPLE_FAKE_TOKEN','account_id':'sample-account'}}));auth.chmod(0o600)
 env.update(EN_CODEX_AUTH_FILE=str(auth),LLM_BASE_URL=f'http://127.0.0.1:{provider.server_port}',EN_RECALL_DISABLED='1')
 subject=launch('subject',blocks[11][1].replace(':8093/',f':{port}/'));ready(subject,work/'subject.out','Gateway Agent ensemble ready')
 interviewer=launch('interviewer',blocks[12][1].replace(':8093/',f':{port}/'));assert interviewer.wait(timeout=15)==0
 report=(work/'interview.md').read_text();assert len(requests)==4 and not errors,errors
 for n,q in enumerate(requests,1):
  assert f'LOCAL_GATEWAY_ANSWER_{n}' in report
  assert '[GATEWAY source=agents user=interviewer trust=trusted]' in json.dumps(q)
 for name,p,so,se in reversed(children):
  if p.poll() is None:p.send_signal(signal.SIGINT)
  assert p.wait(timeout=5)==0,(name,se.read_text())
 saved=json.loads((work/'subject-data/save.json').read_text());events=saved['log']
 assert sum(e['type']=='request_sent' for e in events)==4 and sum(e['type']=='response_ended' for e in events)==4
 (out/'samples-interview.md').write_text(report)
 (out/'samples-events.json').write_text(json.dumps({'config':saved['config'],'log':events},indent=2)+'\n')
 # The vocabulary excerpts compile in an outside module. The Connect fragment
 # stays a construction excerpt; actual Connect runs above and in the grader.
 api=work/'api';api.mkdir();(api/'go.mod').write_text('module outside.example/samples26\n\ngo 1.25.0\nrequire ensemble v0.0.0\nreplace ensemble => '+str(core)+'\n')
 code='package main\nimport("context";"time";"reflect";gateway "ensemble/gateway")\ntype Envelope = gateway.Envelope\ntype Keypair = gateway.Keypair\n'+blocks[3][1].replace('env Envelope','env gateway.Envelope').replace('welcome WelcomePayload','welcome gateway.WelcomePayload')+'\nvar _ Handler = gateway.NopHandler{}\n'+blocks[5][1]+'\nfunc construction(ctx context.Context,key Keypair,gwPub []byte,handler gateway.Handler){'+blocks[4][1]+';_ = client;_ = err}\nfunc main(){\n'
 for typ in ['WelcomePayload','ClientConfig','OutboundMessage','ChildRegistration']:
  code+='a'+typ+':=reflect.TypeOf('+typ+'{});b'+typ+':=reflect.TypeOf(gateway.'+typ+'{});for i:=0;i<a'+typ+'.NumField();i++{f:=a'+typ+'.Field(i);g,ok:=b'+typ+'.FieldByName(f.Name);if !ok||f.Type!=g.Type||f.Tag!=g.Tag{panic("public field mismatch '+typ+' "+f.Name)}}\n'
 code+='}\n';(api/'main.go').write_text(code);env.update(GOWORK='off',GOPROXY='off',GOFLAGS='-mod=mod');run('go run .',api)
 # The decrypted JSON example is checked against actual published field tags.
 frames=[json.loads(line) for line in blocks[6][1].splitlines()];assert frames[1]['welcome']['queue_depth']==0
 (out/'samples-api.go.txt').write_text(code)
 result={'result':'PASS','manuscript_sha256':sha,'fences':[{'index':i,'language':k,'sha256':hashlib.sha256(v.encode()).hexdigest()} for i,(k,v) in enumerate(blocks)],'main_requests':len(requests),'seconds':round(time.monotonic()-started,3),'port_substitution':port,'limits':'Scripted external HTTP replies verify actual three-command/actor/routing mechanics, not a fresh autonomous or paid Astra interview. Synthetic Codex credential and loopback endpoint only; no personal profile accessed. Existing 11.739-second real interview remains separate. Grading fences reuse actual initial recorded invocation, not another full grade. Policy fragment 2 has an explicitly illustrative invalid key; full configuration 8 receives the documented public-key/directory substitutions.'}
 (out/'samples.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
finally:
 for name,p,so,se in reversed(children):
  if p.poll() is None:p.send_signal(signal.SIGINT)
  try:p.wait(timeout=5)
  except subprocess.TimeoutExpired:p.kill();p.wait()
  receipts.append({'child':name,'pid':p.pid,'exit':p.returncode,'stdout':so.read_text(),'stderr':se.read_text()})
 (out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n');(out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n')
 provider.shutdown();provider.server_close()
 if work:shutil.rmtree(work)
