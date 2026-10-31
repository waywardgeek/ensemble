"""Exact published clients, with scripted decisions only at the remote HTTP boundary."""
import hashlib,http.server,json,os,pathlib,re,shutil,signal,socket,subprocess,tempfile,threading,time,urllib.request
import websocket
root=pathlib.Path(__file__).resolve().parents[5];out=pathlib.Path(__file__).parent
chapter=root/'book/edition-2/chapter-22.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha=='d2f514380905245e14720fbf3d9d9a1a6555e813077487da3385ab5505bfdf3e',sha
blocks=re.findall(r'^```([^\n]*)\n(.*?)^```',chapter.read_text(),re.M|re.S)
assert len(blocks)==16
receipts=[];requests=[];judges=[];errors=[];children=[];workdirs=[];gui_mode=True;join_requested=threading.Event()

def call(id,name,args):return dict(type='function_call',id='item-'+id,call_id=id,name=name,arguments=json.dumps(args),status='completed')
def text(s):return dict(type='message',role='assistant',content=[dict(type='output_text',text=s)])
def previous(q):return {x['call_id']:x['output'] for x in q.get('input',[]) if x.get('type')=='function_call_output'}
def route(q):
 if gui_mode:return text('GUI_SAMPLE_DONE')
 history='\n'.join(x.get('text','') for m in q['input'] if m.get('type')=='message' and m.get('role')=='user' for x in m.get('content',[]))
 done=previous(q)
 name=next((n for n in ['add','multiply','subtract'] if 'CH22_CHILD_'+n in history),None)
 if name:
  if name=='add' and 'ask-add' not in done:return call('ask-add','send_message_to_parent',{'content':'Must negative integers work?','ai_callback_delay':10})
  source='def '+name+'(a, b):\n    return a '+{'add':'+','multiply':'*','subtract':'-'}[name]+' b\n'
  if 'write-'+name not in done:return call('write-'+name,'write_file',{'path':name+'.py','content':source})
  if 'test-'+name not in done:
   expected={'add':-1,'multiply':-6,'subtract':-5}[name]
   return call('test-'+name,'run_command',{'command':"sleep 4 && python3 -c 'from "+name+' import '+name+'; assert '+name+'(-3, 2) == '+str(expected)+"; print(\"CHILD_TEST_"+name+"\")'",'ai_callback_delay':10})
  return call('submit-'+name,'submit',{'result':name+' tested negative integers'})
 for n in ['add','multiply','subtract']:
  if 'spawn-'+n not in done:return call('spawn-'+n,'spawn_sub_agent',{'agent_id':n,'task':'CH22_CHILD_'+n+'; implement and test your named function. Ask whether negatives matter if add.'})
 if 'question' not in done:return call('question','join_agents',{'agent_ids':['add','multiply','subtract'],'ai_callback_delay':10})
 if 'answer' not in done:return call('answer','respond_to_parent_message',{'agent_id':'add','response':'Yes, negative integers are required.'})
 if 'limits' not in done:return call('limits','tool_limits',{'ai_callback_delay':60})
 if 'join-all' not in done:
  join_requested.set();return call('join-all','join_agents',{'agent_ids':['add','multiply','subtract']})
 if 'running' in done['join-all'].lower() and 'wait-all' not in done:
  h=re.search(r'(?i)\b(?:handle|job)[" :=]+([0-9]+)',done['join-all']);assert h,done['join-all']
  return call('wait-all','wait_for_job',{'handle':int(h[1]),'ai_callback_delay':10})
 if 'combined' not in done:return call('combined','run_command',{'command':"python3 -c 'from add import add; from multiply import multiply; from subtract import subtract; assert add(-3,2)==-1; assert multiply(-3,2)==-6; assert subtract(-3,2)==-5; print(\"COMBINED_PASS\")'",'ai_callback_delay':10})
 return call('root-submit','submit',{'result':'FLEET_DONE STEER_COBALT; negative integers tested.'})
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  try:
   q=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
   assert self.path=='/v1/responses' and self.headers.get('Authorization')=='Bearer sample-only-key'
   assert q['model'] in ['gpt-6-astra','gpt-6.1-sol']
   if not gui_mode and not q.get('tools'):judges.append(q);item=text('[]')
   else:
    assert q['reasoning']=={'effort':'high','summary':'detailed'}
    requests.append(q);item=route(q)
   self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
   response=dict(status='completed',model=q['model'],usage=dict(input_tokens=100,output_tokens=10,input_tokens_details=dict(cached_tokens=0)),output=[item])
   self.wfile.write(('event: response.completed\ndata: '+json.dumps(dict(type='response.completed',response=response))+'\n\n').encode());self.wfile.flush()
  except Exception as e:errors.append(repr(e));raise
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
profile=pathlib.Path(tempfile.mkdtemp(prefix='ch22-profile-'));workdirs.append(profile)
env={k:v for k,v in os.environ.items() if k in ['HOME','PATH','TMPDIR','USER','LANG','SHELL']}
env.update(EN_HOME=str(profile),EN_BILLING_MODE='api_key',LLM_API_KEY='sample-only-key',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',GOWORK='off',GOPROXY='off')
def run(cmd,cwd):
 r=subprocess.run(['bash','-c',cmd],cwd=cwd,env=env,capture_output=True,text=True,timeout=40);receipts.append(dict(command=cmd,exit=r.returncode,stdout=r.stdout,stderr=r.stderr));assert r.returncode==0,receipts[-1];return r.stdout

def launch(cmd,cwd):
 so=profile/(str(len(children))+'.out');se=profile/(str(len(children))+'.err')
 with so.open('w') as f,se.open('w') as g:p=subprocess.Popen(['bash','-c','exec env '+cmd],cwd=cwd,env=env,stdin=subprocess.PIPE,stdout=f,stderr=g,text=True)
 children.append((p,so,se,cmd));return p,so,se

def wait(p,so,se,marker,seconds=20):
 until=time.monotonic()+seconds
 while marker not in so.read_text() and p.poll() is None and time.monotonic()<until:time.sleep(.02)
 assert p.poll() is None and marker in so.read_text(),(p.poll(),so.read_text(),se.read_text())

def stop(p,so,se,interrupt=True):
 if interrupt:p.send_signal(signal.SIGINT)
 else:p.stdin.close()
 code=p.wait(timeout=10);receipts.append(dict(command=next(x[3] for x in children if x[0] is p),exit=code,stdout=so.read_text(),stderr=se.read_text()));assert code==0

def get(url):
 try:
  with urllib.request.urlopen(url,timeout=3) as r:return r.status,r.read().decode()
 except urllib.error.HTTPError as e:return e.code,e.read().decode()

def reply_mcp(ws,m):
 rpc=m.get('payload',{})
 if rpc.get('method') in ['initialize','tools/list'] and 'id' in rpc:
  result={'tools':[]} if rpc['method']=='tools/list' else {'protocolVersion':'2024-11-05','capabilities':{'tools':{}},'serverInfo':{'name':'sample-browser','version':'1'}}
  ws.send(json.dumps(dict(type='jsonrpc',source=m.get('source'),payload=dict(jsonrpc='2.0',id=rpc['id'],result=result))))

try:
 setup=run(blocks[7][1]+'\nprintf "REUSE_DIR=%s\\n" "$reuse_dir"\n',root)
 reuse=pathlib.Path(re.search(r'REUSE_DIR=(.+)',setup)[1]);workdirs.append(reuse)
 (reuse/'main.go').write_text(blocks[8][1]);run(blocks[9][1].rsplit('/tmp/ensemble-ch22-reuse --components',1)[0],reuse)
 p,so,se=launch('/tmp/ensemble-ch22-reuse --components',reuse);wait(p,so,se,'HTTP ');url=re.search(r'HTTP (\S+)',so.read_text())[1]
 assert get(url+'/index.html')[0]==404 and 'ArtifactScroll' in get(url+'/artifact-scroll.js')[1]
 browser=profile/'component.mjs';browser.write_text('import { chromium } from '+json.dumps(str(root/'solutions/edition-2/main/gui/node_modules/playwright/index.mjs'))+'; const b=await chromium.launch({headless:true,executablePath:"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}); try {const p=await b.newPage();await p.goto(process.argv[2]+"/demo");await p.waitForSelector(".artifact");const s=await p.locator("body").innerText();if(!s.includes("An Artifact without the application shell."))throw Error(s);console.log(s)} finally {await b.close()}')
 run('node '+str(browser)+' '+url,reuse);stop(p,so,se)
 p,so,se=launch(blocks[10][1],reuse);wait(p,so,se,'HTTP ');url=re.search(r'HTTP (\S+)',so.read_text())[1];address=re.search(r'MCP (\S+)',so.read_text())[1]
 tcp=socket.create_connection((address.split(':')[0],int(address.split(':')[1])),3);tcp.settimeout(5);stream=tcp.makefile('rwb',buffering=0);stream.write(blocks[11][1].encode());a=json.loads(stream.readline());assert a['id']=='outside-1' and a.get('error')
 ws=websocket.create_connection(url.replace('http://','ws://')+'/ws',timeout=5);ws.send(json.dumps({'type':'subscribe'}));replay=False
 while not replay:
  m=json.loads(ws.recv());reply_mcp(ws,m);replay=m.get('type')=='user_message' and 'PUBLIC_REPLAY' in json.dumps(m.get('part',{}))
 for i,model in enumerate(['gpt-6-astra','gpt-6.1-sol']):
  if i:
   ws.send(json.dumps(dict(type='update_settings',settings={'model':model})))
   while True:
    m=json.loads(ws.recv());reply_mcp(ws,m)
    if m.get('type')=='settings_changed' and m.get('settings',{}).get('model')==model:break
  ws.send(json.dumps(dict(type='prompt',text='Sample turn '+str(i))))
  visible=False
  while True:
   m=json.loads(ws.recv());reply_mcp(ws,m);visible=visible or (m.get('type')=='part_final' and 'GUI_SAMPLE_DONE' in json.dumps(m.get('part',{})))
   if m.get('type')=='turn_ended':assert not m.get('error') and visible;break
  assert requests[-1]['model']==model
 stream.write(blocks[11][1].encode())
 while True:
  m=json.loads(ws.recv());rpc=m.get('payload',{})
  if rpc.get('method')=='tools/list':
   ws.send(json.dumps(dict(type='jsonrpc',source=m.get('source'),payload=dict(jsonrpc='2.0',id=rpc['id'],result={'tools':[],'sample':'CORRELATED_BROWSER'}))));break
 a=json.loads(stream.readline());assert a['id']=='outside-1' and a['result']['sample']=='CORRELATED_BROWSER'
 tcp.close();stream.close();stop(p,so,se);ws.close()
 override=profile/'assets';override.mkdir();(override/'index.html').write_text('SAMPLE_DISK_OVERRIDE')
 p,so,se=launch(blocks[10][1].strip()+' --gui-dir '+str(override),reuse);wait(p,so,se,'HTTP ');url=re.search(r'HTTP (\S+)',so.read_text())[1];assert get(url+'/')[1]=='SAMPLE_DISK_OVERRIDE';stop(p,so,se)
 run(blocks[12][1],reuse);assert not re.search('ensemble-gui-stub|gorilla/websocket',(reuse/'headless-deps.txt').read_text());assert 'gorilla/websocket' in (reuse/'gui-deps.txt').read_text()
 setup,launch_text=blocks[13][1].split('LLM_VENDOR=openai',1);env['ensemble_source']=str(root/'solutions/edition-2/ch22')
 raw=run(setup+'\nprintf "FLEET_DIR=%s\\n" "$fleet_dir"',reuse);fleet=pathlib.Path(re.search(r'FLEET_DIR=(.+)',raw)[1]);workdirs.append(fleet);env['fleet_dir']=str(fleet)
 gui_mode=False;p,so,se=launch('LLM_VENDOR=openai'+launch_text,fleet)
 # Human multiline prose is one terminal input; folding line wraps changes no words.
 p.stdin.write(' '.join(blocks[14][1].splitlines())+'\n');p.stdin.flush()
 assert join_requested.wait(15),'join request never reached provider boundary'
 # The real journal supplies the dispatch barrier, rather than an arbitrary delay.
 until=time.monotonic()+5;dispatched=False
 while time.monotonic()<until:
  for path in fleet.rglob('*.journal'):
   if 'children' in path.parts:continue
   for line in path.read_text().splitlines():
    try:event=json.loads(line)
    except ValueError:continue
    if event.get('type')=='tool_called' and 'join-all' in json.dumps(event):dispatched=True
  if dispatched:break
  time.sleep(.01)
 assert dispatched,'real join dispatch not observed in journal'
 p.stdin.write(' '.join(blocks[15][1].splitlines())+'\n');p.stdin.flush();wait(p,so,se,'FLEET_DONE',20);stop(p,so,se,False)
 run("python3 -c 'from add import add; from multiply import multiply; from subtract import subtract; assert add(-3,2)==-1; assert multiply(-3,2)==-6; assert subtract(-3,2)==-5; print(\"INDEPENDENT_PASS\")'",fleet)
 for n in ['add','multiply','subtract']:assert (fleet/'cr/children'/n/'save.json').is_file()
 actual={k:v for q in requests for k,v in previous(q).items()}
 for n in ['add','multiply','subtract']:assert 'CHILD_TEST_'+n in actual['test-'+n] and 'exit_code: 0' in actual['test-'+n]
 assert 'COMBINED_PASS' in actual['combined'] and 'Yes, negative integers' in actual['ask-add']
 assert not errors,errors
 (out/'samples.json').write_text(json.dumps(dict(result='PASS',manuscript_sha256=sha,fences=[dict(index=i,language=k,sha256=hashlib.sha256(v.encode()).hexdigest()) for i,(k,v) in enumerate(blocks)],requests=len(requests),judges=len(judges),limits='Faithful scripted remote decisions prove actual client/tool mechanics and component rendering, not new autonomous coding. Only fresh printed listener addresses used; selected fake API-key environment, no real profile or provider.'),indent=2)+'\n')
finally:
 for p,so,se,cmd in children:
  if p.poll() is None:p.kill();p.wait()
 (out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n');(out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n');(out/'samples-judges.json').write_text(json.dumps(judges,indent=2)+'\n');(out/'samples-errors.json').write_text(json.dumps(errors,indent=2)+'\n')
 server.shutdown();server.server_close()
 for d in workdirs:shutil.rmtree(d)
