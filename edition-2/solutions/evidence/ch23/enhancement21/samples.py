"""Exact published CLI setup; only signed auth and model HTTP are external fakes."""
import hashlib,http.server,json,os,pathlib,re,shutil,subprocess,tempfile,threading,time
root=pathlib.Path(os.environ['CH21_ROOT']);out=root/'solutions/edition-2/evidence/ch23/enhancement21';chapter=root/'book/edition-2/chapter-21.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha=='20c231d92254d650f4adca70c01d89b2164c73b3112a7bf692480e5971141b5d'
blocks=re.findall(r'^```([^\n]*)\n(.*?)^```',chapter.read_text(),re.M|re.S)
assert len(blocks)==12
receipts=[];requests=[];judges=[];errors=[];script=[];demo=None;child=None

def call(name,args):return {'type':'function_call','id':'item_'+str(len(script)),'call_id':'call_'+str(len(script)),'name':name,'arguments':json.dumps(args)}
def add(name,**args):
 if name=='run_command':args['ai_callback_delay']=5
 script.append(call(name,args))
def final(text):script.append({'type':'message','role':'assistant','content':[{'type':'output_text','text':text}]})
add('write_file',path='notes.txt',content='amber apple\n');add('read_file',path='notes.txt');add('edit_file',path='notes.txt',old_text='apple',new_text='plum');add('search_files',path='.',pattern='plum');add('list_directory',path='.')
add('run_command',command='/usr/bin/python3 -c \'assert open("notes.txt").read() == "amber plum\\n"; print("PYTHON_OK")\'')
add('run_command',command='git init -q && git add notes.txt && git commit -qm sandbox-example && git log -1 --format=%s');final('SAMPLE_TURN_ONE_COMPLETE')
add('read_file',path='../outside.txt');add('read_file',path='escape');add('run_command',command="sh -c 'sh -c \"cat ../outside.txt\"'");add('run_command',command='printf BAD > ../outside-write.txt');add('run_command',command='if [ -n "$SANDBOX_FAKE_TEST_TOKEN" ]; then printf TOKEN_PRESENT; else printf TOKEN_ABSENT; fi');final('SAMPLE_TURN_TWO_COMPLETE')
add('write_file',path='safe.txt',content='safe mode file works\n');add('read_file',path='safe.txt');add('read_file',path='../notes.txt');add('run_command',command='printf BAD > guessed.txt');final('SAMPLE_SAFE_COMPLETE')
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  try:
   q=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
   assert self.path=='/v1/responses' and self.headers.get('Authorization','').startswith('Bearer ')
   assert q['model']=='gpt-6-astra'
   if not q.get('tools'):
    judges.append(q);item={'type':'message','role':'assistant','content':[{'type':'output_text','text':'[]'}]}
   else:
    assert q['reasoning']=={'effort':'high','summary':'detailed'}
    requests.append(q);item=script[len(requests)-1]
   self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
   response={'status':'completed','model':q['model'],'usage':{'input_tokens':100,'output_tokens':10,'input_tokens_details':{'cached_tokens':0}},'output':[item]}
   event={'type':'response.completed','response':response}
   self.wfile.write(('event: response.completed\ndata: '+json.dumps(event)+'\n\n').encode());self.wfile.flush()
  except Exception as e:errors.append(repr(e));raise
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
profile=pathlib.Path(tempfile.mkdtemp(prefix='ch21-sample-profile-'))
env={k:v for k,v in os.environ.items() if k in ['HOME','PATH','TMPDIR','USER','LANG','SHELL']}
env.update(EN_HOME=str(profile),OPENAI_OIDC_ISSUER=os.environ['CH21_ISSUER'],LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}')
def run(command,cwd,env):
 r=subprocess.run(['bash','-c',command],cwd=cwd,env=env,capture_output=True,text=True,timeout=30);receipts.append({'command':command,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr});assert r.returncode==0,receipts[-1];return r.stdout

def launch(command,prompts,markers):
 global child
 stdout=profile/'stdout.txt';stderr=profile/'stderr.txt'
 with stdout.open('w') as so,stderr.open('w') as se:
  child=subprocess.Popen(['bash','-c','exec '+command],cwd=demo/'work',env=env,stdin=subprocess.PIPE,stdout=so,stderr=se,text=True)
  for prompt,marker in zip(prompts,markers):
   assert child.poll() is None
   child.stdin.write(prompt);child.stdin.flush();deadline=time.monotonic()+20
   while marker not in stdout.read_text() and child.poll() is None and time.monotonic()<deadline:time.sleep(.02)
   assert marker in stdout.read_text(),(stdout.read_text(),stderr.read_text())
  child.stdin.close();code=child.wait(timeout=10)
 receipts.append({'command':command,'exit':code,'stdout':stdout.read_text(),'stderr':stderr.read_text(),'completed_before_eof':True});assert code==0
 child=None
try:
 run(blocks[6][1],root/'solutions/edition-2/ch21',env)
 browser=profile/'browser.py';browser.write_text('import sys,urllib.request\nwith urllib.request.urlopen(sys.argv[1],timeout=5) as r:r.read()\n');env['EN_BROWSER_CMD']='python3 '+str(browser)
 run('/tmp/ensemble-ch21 auth login',profile,env)
 setup,command=blocks[7][1].rsplit('/tmp/ensemble-ch21 chat',1)
 raw=run(setup+'\npython3 -c \'import os,json;print(json.dumps(dict(os.environ)))\'\n',profile,env)
 selected=json.loads(raw);env=selected;demo=pathlib.Path(env['EN_SKILLS_DIR']).parent;env['ensemble_ch21_demo']=str(demo)
 launch('/tmp/ensemble-ch21 chat'+command,[blocks[8][1],blocks[9][1]],['SAMPLE_TURN_ONE_COMPLETE','SAMPLE_TURN_TWO_COMPLETE'])
 assert (demo/'work/notes.txt').read_text()=='amber plum\n'
 assert (demo/'outside.txt').read_text()=='OUTSIDE_FIXTURE_ONLY\n' and not (demo/'outside-write.txt').exists()
 launch(blocks[10][1],[blocks[11][1]],['SAMPLE_SAFE_COMPLETE'])
 assert (demo/'work/sub/safe.txt').read_text()=='safe mode file works\n' and not (demo/'work/sub/guessed.txt').exists()
 actual={v['call_id']:v['output'] for q in requests for v in q['input'] if v.get('type')=='function_call_output'}
 for i in [0,1,2,3,4,5,6,8,9,10,11,12,14,15,16,17]:assert script[i]['call_id'] in actual,i
 assert actual[script[1]['call_id']]=='amber apple\n'
 assert 'plum' in actual[script[3]['call_id']] and 'notes.txt' in actual[script[4]['call_id']]
 assert 'PYTHON_OK' in actual[script[5]['call_id']] and 'exit_code: 0' in actual[script[5]['call_id']]
 assert 'sandbox-example' in actual[script[6]['call_id']] and 'exit_code: 0' in actual[script[6]['call_id']]
 for i in [8,9,16]:assert 'sandbox refused' in actual[script[i]['call_id']],actual[script[i]['call_id']]
 for i in [10,11]:assert re.search(r'exit_code: [1-9]',actual[script[i]['call_id']]),actual[script[i]['call_id']]
 assert 'TOKEN_ABSENT' in actual[script[12]['call_id']] and 'TOKEN_PRESENT' not in actual[script[12]['call_id']]
 assert actual[script[15]['call_id']]=='safe mode file works\n' and 'unknown tool' in actual[script[17]['call_id']]
 assert len(requests)==len(script) and not errors
 # Source snippets are actual owning code, allowing only whitespace presentation.
 bindings=[]
 for i,path in [(2,'internal/sandbox/files.go'),(3,'ensemble/spec.go'),(4,'ensemble/spec.go'),(5,'internal/sandbox/files.go')]:
  source=(root/'solutions/edition-2/ch21'/path).read_text();assert re.sub(r'\s+','',blocks[i][1]) in re.sub(r'\s+','',source);bindings.append({'fence':i,'source':path})
 (out/'samples.json').write_text(json.dumps({'result':'PASS','manuscript_sha256':sha,'fences':[{'index':i,'language':k,'sha256':hashlib.sha256(v.encode()).hexdigest()} for i,(k,v) in enumerate(blocks)],'requests':len(requests),'judges':len(judges),'go_bindings':bindings,'limits':'Scripted external decisions prove CLI/tool mechanics; historical paid evidence remains separate. No new model quality or account eligibility claim.'},indent=2)+'\n')
finally:
 if child is not None and child.poll() is None:child.kill();child.wait()
 (out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n');(out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n');(out/'samples-judges.json').write_text(json.dumps(judges,indent=2)+'\n')
 server.shutdown();server.server_close();shutil.rmtree(profile)
 if demo:shutil.rmtree(demo)
