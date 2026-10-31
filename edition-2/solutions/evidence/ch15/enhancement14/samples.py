import pathlib,re,tempfile,shutil,subprocess,json,os,threading,http.server,time,hashlib,websocket
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch15/enhancement14';chapter=root/'book/edition-2/chapter-14.md';text=chapter.read_text();blocks=re.findall(r'```([^\n]*)\n(.*?)```',text,re.S)
sha=hashlib.sha256(chapter.read_bytes()).hexdigest();assert sha=='55e3a4a4900f8cfbd10b65aa58ce000435d7f404129937b1c02394c8095cf816'
requests=[];replies=[]
def call(id,name,args):return {'type':'tool_use','id':id,'name':name,'input':args}
def say(text):return {'type':'text','text':text}
replies.extend([[call('sample-load','load_skill',{'name':'manual'})],[call('sample-read-1','read_file',{'path':'records/record-1.txt'})],[say('CEDAR-17 verified.')]])
for i,label in enumerate(['MAPLE','WILLOW','BIRCH','OAK','ELM'],2):replies.extend([[call('sample-read-'+str(i),'read_file',{'path':f'records/record-{i}.txt'})],[say(f'{label}-{i*17} verified.')]])
replies.extend([[call('sample-handoff','micro_handoff',{'text':'CHECKPOINT_SAMPLE: six records inspected, CEDAR-17 through ELM-102; no observed tool failures; next report these codes.'})],[call('sample-unload','unload_skill',{'name':'manual'})],[say('CEDAR-17 and ELM-102. The manual instructions remain.')]])
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  request=json.loads(self.rfile.read(int(self.headers['Content-Length'])));requests.append(request);parts=replies[len(requests)-1]
  body=json.dumps({'id':'response-'+str(len(requests)),'type':'message','role':'assistant','model':request['model'],'content':parts,'stop_reason':'tool_use' if any(p['type']=='tool_use' for p in parts) else 'end_turn','usage':{'input_tokens':100,'output_tokens':20}}).encode()
  self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
with tempfile.TemporaryDirectory(prefix='ch14-published-') as tmp:
 work=pathlib.Path(tmp)/'source';shutil.copytree(root/'solutions/edition-2/ch14',work)
 # Compile the exact current Go excerpts against their actual surrounding types.
 go=[b for lang,b in blocks if lang=='go'];vocabulary=go[0]
 sample=work/'snippet';sample.mkdir();(sample/'vocabulary.go').write_text('package snippet\n'+vocabulary)
 for name,path in [('survivor','internal/history/reduce.go'),('replay','internal/history/history.go')]:
  fence=next(b for b in go if re.search(r'func .*\b'+name+r'\(',b));p=work/path;s=p.read_text();pattern=r'func (?:\(h \*history\) )?'+name+r'\([^\n]*\n.*?^}'
  s,n=re.subn(pattern,lambda m:fence[fence.index('func '):].strip(),s,count=1,flags=re.S|re.M);assert n==1;p.write_text(s)
 checked=subprocess.run(['go','test','./snippet','./internal/history','./ensemble','-run','TestContext','-count=1'],cwd=work,capture_output=True,text=True);assert checked.returncode==0,checked.stdout+checked.stderr
 out.joinpath('samples-go.txt').write_text(checked.stdout+checked.stderr)
 spin=text.split('## §14.16')[1];sb=re.findall(r'```([^\n]*)\n(.*?)```',spin,re.S)
 shell=[b for lang,b in sb if lang=='sh'];r=subprocess.run(['bash','-c',shell[0]],cwd=work,capture_output=True,text=True);assert r.returncode==0,r.stderr
 yamls=[b for lang,b in sb if lang=='yaml'];assert len(yamls)==2
 for name,b in zip(['base','manual'],yamls):(work/'skills'/name/'SKILL.md').write_text(b)
 code=next(b for lang,b in sb if lang=='python');r=subprocess.run(['python3','-c',code],cwd=work,capture_output=True,text=True);assert r.returncode==0,r.stderr
 env=dict(os.environ,LLM_VENDOR='anthropic',LLM_MODEL='claude-sonnet-5-course',LLM_API_KEY='fake-key',LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',EN_PRIMARY_SKILL='base',EN_SKILLS_DIR='',CH02_LOG=str(work/'events.jsonl'))
 # Execute the exact published launcher; shell exec ensures EOF measures the GUI.
 launch=shell[1].replace('EN_PRIMARY_SKILL=base ','exec env EN_PRIMARY_SKILL=base ',1)
 stderr=open(work/'stderr.txt','w');stdout=open(work/'stdout.txt','w');p=subprocess.Popen(['bash','-c',launch],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
 try:
  ws=None;deadline=time.monotonic()+5
  while time.monotonic()<deadline:
   try:ws=websocket.create_connection('ws://127.0.0.1:8088/ws',timeout=5);break
   except OSError:time.sleep(.02)
  assert ws is not None
  ws.send(json.dumps({'type':'subscribe'}))
  while True:
   m=json.loads(ws.recv())
   if m.get('type')=='current_settings':settings=m['settings'];break
  assert settings['context_target']==20000 and settings['max_tool_rounds']==4 and settings['tts_enabled']==False,settings
  prompts=[b.strip() for lang,b in sb if lang=='text'];prompts=[prompts[0]]+[f'Read records/record-{i}.txt. Report its inspection code and status, then stop.' for i in range(2,7)]+[prompts[1]]
  for prompt in prompts:
   p.stdin.write(prompt+'\n');p.stdin.flush()
   while True:
    m=json.loads(ws.recv())
    if m.get('type')=='turn_ended':assert not m.get('error'),m;break
  p.stdin.close();p.wait(timeout=5);assert p.returncode==0
  ws.close()
 finally:
  if p.poll() is None:p.kill();p.wait()
  stdout.close();stderr.close()
 assert len(requests)==16,len(requests)
 out.joinpath('samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n')
 results={}
 for q in requests:
  for m in q['messages']:
   for b in (m['content'] if isinstance(m['content'],list) else []):
    if b.get('type')=='tool_result':results.setdefault(b['tool_use_id'],b)
 for i,label in enumerate(['CEDAR','MAPLE','WILLOW','BIRCH','OAK','ELM'],1):
  b=results['sample-read-'+str(i)];assert not b.get('is_error') and f'{label}-{i*17}' in json.dumps(b)
 last=json.dumps(requests[-1]['messages']);assert last.count('MANUAL-SURVIVES')==1 and last.count('CHECKPOINT_SAMPLE')==1
 save=json.loads((work/'save.json').read_text());kinds={e['type'] for e in save['log']};assert {'redacted','micro_handoff','skill_loaded'}<=kinds
 out.joinpath('samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n')
 out.joinpath('samples.json').write_text(json.dumps({'manuscript_sha256':sha,'source':'solutions/edition-2/ch14 (isolated copy)','go_excerpts':'exact current vocabulary compiled; exact survivor/replay substituted into actual owning package; Context tests pass','historical_excerpts':'not presented as current APIs; unchanged historical evidence','build_command':shell[0],'launch_command':shell[1],'provider':'small external Anthropic HTTP peer; no paid run or model quality claim','model_fixture':'declared no-stub course model; both new flags false','settings':{k:settings[k] for k in ['context_target','max_tool_rounds','tts_enabled']},'requests':len(requests),'successful_correlated_reads':6,'turns':7,'exit':p.returncode,'saved_events':{k:sum(e['type']==k for e in save['log']) for k in ['redacted','micro_handoff','skill_loaded']},'final_manual_and_handoff_once':True},indent=2)+'\n')
server.shutdown()
print('samples PASS',sha,len(requests))
