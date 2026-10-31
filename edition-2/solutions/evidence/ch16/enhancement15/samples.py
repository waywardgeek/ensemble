import pathlib,re,tempfile,shutil,subprocess,json,os,threading,http.server,time,hashlib,websocket
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch16/enhancement15';chapter=root/'book/edition-2/chapter-15.md';text=chapter.read_text();blocks=re.findall(r'```([^\n]*)\n(.*?)```',text,re.S)
sha=hashlib.sha256(chapter.read_bytes()).hexdigest();assert sha=='bd2acd5cef67062ee2d9700e5349bc62fdea60ea7d4236ffc2094f00c24a78da',sha
requests=[];main=0;aux=0
first='I reviewed Cedar Lantern, CL-721. Morgan owns archive review October 18. I rejected automatic deletion because support needs originals; accepted summaries retain a source identifier. Attachment placement remains unresolved. '
second='I reviewed Harbor Finch, HF-934. Priya owns the rehearsal October 21. I rejected automatic migration retry because the operation may already have committed; inspect the ledger. Ledger placement remains unresolved. '
memories=[first+'I preserve the source until completed review. '*10,second+'I verify actual process lifetime after failure. '*10,first+second]
def call(id,name,args):return {'type':'tool_use','id':id,'name':name,'input':args}
def say(t):return {'type':'text','text':t}
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  global main,aux
  q=json.loads(self.rfile.read(int(self.headers['Content-Length'])));requests.append(q)
  if [t['name'] for t in q.get('tools',[])]==['submit']:
   assert aux<3;parts=[call('memory-'+str(aux),'submit',{'memory':memories[aux]})];aux+=1
  else:
   parts=[[call('checkpoint-one','micro_handoff',{'text':first})],[say('First case recorded.')],[call('checkpoint-two','micro_handoff',{'text':second})],[say('Second case recorded.')],[say(first+second)]][main];main+=1
  b=json.dumps({'id':'response-'+str(len(requests)),'type':'message','role':'assistant','model':q['model'],'content':parts,'stop_reason':'tool_use' if any(p['type']=='tool_use' for p in parts) else 'end_turn','usage':{'input_tokens':100,'output_tokens':20}}).encode();self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(b)));self.end_headers();self.wfile.write(b)
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
try:
 with tempfile.TemporaryDirectory(prefix='ch15-samples-') as tmp:
  work=pathlib.Path(tmp)/'source';shutil.copytree(root/'solutions/edition-2/ch15',work)
  snippet=work/'snippet';snippet.mkdir();(snippet/'types.go').write_text('package snippet\n'+next(b for lang,b in blocks if lang=='go'))
  r=subprocess.run(['go','test','./snippet'],cwd=work,capture_output=True,text=True);(out/'samples-go.txt').write_text(r.stdout+r.stderr);assert r.returncode==0
  spin=text.split('## §15.12')[1];sb=re.findall(r'```([^\n]*)\n(.*?)```',spin,re.S);shell=[b for lang,b in sb if lang=='sh'];r=subprocess.run(['bash','-c',shell[0]],cwd=work,capture_output=True,text=True);assert r.returncode==0,r.stderr
  (work/'skills/base/SKILL.md').write_text(next(b for lang,b in sb if lang=='yaml'));(work/'cr/settings.json').write_text(next(b for lang,b in sb if lang=='json'))
  env=dict(os.environ,LLM_API_KEY='fake-key',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',EN_DISABLE_STREAMING='1')
  launch='exec env '+shell[1];stdout=open(work/'stdout.txt','w');stderr=open(work/'stderr.txt','w');p=subprocess.Popen(['bash','-c',launch],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
  try:
   ws=None;deadline=time.monotonic()+5
   while time.monotonic()<deadline:
    try:ws=websocket.create_connection('ws://127.0.0.1:8088/ws',timeout=8);break
    except OSError:time.sleep(.02)
   assert ws is not None;ws.send(json.dumps({'type':'subscribe'}))
   while True:
    m=json.loads(ws.recv())
    if m.get('type')=='current_settings':settings=m['settings'];break
   assert settings['memory']['conversation']['budget']==600 and settings['memory']['session']['budget']==500
   prompts=[b.strip() for lang,b in sb if lang=='text'][:3]
   for prompt in prompts:
    p.stdin.write(prompt+'\n');p.stdin.flush()
    while True:
     m=json.loads(ws.recv())
     if m.get('type')=='turn_ended':assert not m.get('error'),m;break
   p.stdin.close();p.wait(timeout=5);assert p.returncode==0;ws.close()
  finally:
   if p.poll() is None:p.kill();p.wait()
   stdout.close();stderr.close()
  assert main==5 and aux==3,(main,aux)
  assert all(q['model']=='claude-sonnet-5' for q in requests)
  last='\n'.join(b.get('text','') for m in requests[-1]['messages'] for b in m['content'] if b['type']=='text')
  assert first in last and second in last
  save=json.loads((work/'save.json').read_text());counts={k:sum(e['type']==k for e in save['log']) for k in ['compactor_launched','band_populated','band_depopulated','redacted','micro_handoff']}
  (out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n')
  corpus={str(x.relative_to(work)):x.read_text() for x in (work/'cr').rglob('*.md')};(out/'samples-memories.json').write_text(json.dumps(corpus,indent=2)+'\n')
  (out/'samples.json').write_text(json.dumps(dict(manuscript_sha256=sha,source='solutions/edition-2/ch15 isolated exact copy',provider='External local Anthropic HTTP fake; no paid inference or semantic quality proof',go='Exact vocabulary fence compiled',build=shell[0],launch=shell[1],model='claude-sonnet-5',context_window=0,forcing_tested=False,settings=settings,main_requests=main,compressor_requests=aux,turns=3,exit=p.returncode,event_counts=counts,actual_returned_memories_in_final_request=True),indent=2)+'\n')
finally:server.shutdown()
print('Exact samples PASS',sha,main,aux)
