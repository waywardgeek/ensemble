"""Exact chapter setup and real browser controls, with scripted external HTTP peers."""
import hashlib,http.server,json,os,pathlib,re,shutil,subprocess,tempfile,threading,time,urllib.request
root=pathlib.Path(os.environ['CH23_ROOT']);out=root/'solutions/edition-2/evidence/ch26/enhancement23'
chapter=root/'book/edition-2/chapter-23.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha=='6ce1110e7011d47171694deef26b64f88e8a0297b9cbc9dd6fedd6ee2973b06f'
blocks=re.findall(r'^```([^\n]*)\n(.*?)^```',chapter.read_text(),re.M|re.S);assert len(blocks)==6
receipts=[];requests=[];judges=[];errors=[];work=None;child=None
source='def slugify(text):\n    return "-".join(text.lower().split())\n'
tests='import unittest\nfrom slugify import slugify\nclass SlugifyTests(unittest.TestCase):\n    def test_empty(self): self.assertEqual(slugify(""), "")\n    def test_case(self): self.assertEqual(slugify("MiXeD"), "mixed")\n    def test_whitespace(self): self.assertEqual(slugify("  One \\t Two   "), "one-two")\n'
def call(cid,name,**args):return {'type':'function_call','id':'item_'+cid,'call_id':cid,'name':name,'arguments':json.dumps(args)}
def text(s):return {'type':'message','role':'assistant','content':[{'type':'output_text','text':s}]}
def reply(q):
 inputs=q.get('input',[]);done={v.get('call_id') for v in inputs if v.get('type')=='function_call_output'}
 # The root's task quotes the helper brief in call arguments; use actual user text.
 users=' '.join(str(v.get('content','')) for v in inputs if v.get('role')=='user')
 if 'CH23_FLEET_COMPLETE' not in users and 'CH23_HELPER_TASK' in users:
  sequence=[call('write-code','write_file',path=str(work/'slugify.py'),content=source),call('write-tests','write_file',path=str(work/'test_slugify.py'),content=tests),call('helper-test','run_command',command=f'while [ ! -f "{work}/sample-release" ]; do sleep .02; done; cd "{work}" && python3 -m unittest -v test_slugify.py',ai_callback_delay=10),call('helper-submit','submit',result='CH23_HELPER_COMPLETE; three unittest cases passed')]
 else:
  sequence=[call('spawn-helper','spawn_sub_agent',agent_id='helper',task='CH23_HELPER_TASK. Write and test slugify.py with the specified three cases, then submit.'),call('helper-status','get_sub_agent_status',agent_id='helper'),call('helper-change','wait_for_agent_change',agent_ids=['helper'],timeout_seconds=5,ai_callback_delay=10),call('join-helper','join_agents',agent_ids=['helper'],ai_callback_delay=10),call('submission','get_submission',agent_id='helper'),call('parent-test','run_command',command=f'cd "{work}" && python3 -m unittest -v test_slugify.py',ai_callback_delay=10),call('write-result','write_file',path=str(work/'fleet-result.txt'),content='Three unchanged unittest cases passed in child and independent parent runs.\n'),call('console-submit','submit',result='CH23_FLEET_COMPLETE; helper and independent parent tests passed')]
 for item in sequence:
  if item['call_id'] not in done:return item
 return text('CH23_HELPER_COMPLETE' if sequence[0]['call_id']=='write-code' else 'CH23_FLEET_COMPLETE')
class Provider(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  try:
   q=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
   assert self.path=='/v1/responses' and self.headers.get('Authorization','').startswith('Bearer ')
   assert q['model']=='gpt-6-astra'
   if not q.get('tools'):
    judges.append(q);item=text('[]')
   else:
    assert q['reasoning']=={'effort':'high','summary':'detailed'};requests.append(q);item=reply(q)
   self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
   response={'status':'completed','model':q['model'],'usage':{'input_tokens':100,'output_tokens':10,'input_tokens_details':{'cached_tokens':0}},'output':[item]}
   self.wfile.write(('event: response.completed\ndata: '+json.dumps({'type':'response.completed','response':response})+'\n\n').encode());self.wfile.flush()
  except Exception as e:errors.append(repr(e));raise
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Provider);threading.Thread(target=server.serve_forever,daemon=True).start()
profile=pathlib.Path(tempfile.mkdtemp(prefix='ch23-sample-profile-'))
env={k:v for k,v in os.environ.items() if k in ['HOME','PATH','TMPDIR','USER','LANG','SHELL']}
env.update(EN_HOME=str(profile),OPENAI_OIDC_ISSUER=os.environ['CH23_ISSUER'],LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}')
def run(command,cwd):
 r=subprocess.run(['bash','-c',command],cwd=cwd,env=env,capture_output=True,text=True,timeout=40);receipts.append({'command':command,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr});assert r.returncode==0,receipts[-1];return r.stdout

def launch(mode):
 global child
 so=profile/(mode+'.stdout');se=profile/(mode+'.stderr')
 with so.open('w') as a,se.open('w') as b:child=subprocess.Popen(['bash','-c',blocks[2][1]],cwd=root,env=env,stdin=subprocess.PIPE,stdout=a,stderr=b,text=True)
 deadline=time.monotonic()+5;url=None
 while time.monotonic()<deadline:
  assert child.poll() is None,('early child exit',se.read_text())
  match=re.search(r'Open (http://127.0.0.1:\d+/)',so.read_text())
  if match:
   url=match[1]
   try:
    with urllib.request.urlopen(url,timeout=.2) as response:assert b'agent-tree' in response.read()
    assert child.poll() is None
    # lsof binds readiness to this shell-exec'd actual GUI process, not a stranger.
    port=url.split(':')[-1].strip('/')
    observed=subprocess.run(['lsof','-nP','-a','-p',str(child.pid),'-iTCP:'+port,'-sTCP:LISTEN'],capture_output=True,text=True)
    assert observed.returncode==0,observed.stdout+observed.stderr
    break
   except (OSError,TimeoutError):pass
  time.sleep(.02)
 else:raise AssertionError(('child readiness',so.read_text(),se.read_text()))
 before=len(requests)+len(judges)
 r=subprocess.run(['node',str(out/'samples-browser.mjs'),url,str(work),str(out),mode],env=env,capture_output=True,text=True,timeout=35)
 receipts.append({'browser_mode':mode,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr,'child_pid':child.pid,'url':url,'readiness_owner':observed.stdout});assert r.returncode==0,receipts[-1]
 if mode=='restart':assert len(requests)+len(judges)==before
 child.stdin.close();code=child.wait(timeout=5);receipts.append({'launch_fence':2,'mode':mode,'exit':code,'stdout':so.read_text(),'stderr':se.read_text()});assert code==0
 child=None
try:
 raw=run(blocks[1][1]+'\nprintf "CH23_WORK=%s\\n" "$CH23_WORK"\n',root)
 work=pathlib.Path(re.search(r'^CH23_WORK=(.+)$',raw,re.M)[1]);env['CH23_WORK']=str(work)
 (work/'prompt.txt').write_text(blocks[3][1])
 browser=profile/'browser.py';browser.write_text('import sys,urllib.request\ntry:\n urllib.request.urlopen(sys.argv[1],timeout=5).read()\nexcept OSError:\n pass # terminal token exchange remains independently required\n');env['EN_BROWSER_CMD']='python3 '+str(browser)
 run('/tmp/ensemble-ch23-gui auth login',profile)
 launch('first');launch('restart')
 assert (work/'slugify.py').read_text()==source and (work/'test_slugify.py').read_text()==tests
 assert 'Three unchanged' in (work/'fleet-result.txt').read_text()
 actual={v['call_id']:v['output'] for q in requests for v in q['input'] if v.get('type')=='function_call_output'}
 for cid in ['helper-test','parent-test']:assert 'Ran 3 tests' in actual[cid] and 'OK' in actual[cid] and 'exit_code: 0' in actual[cid],actual[cid]
 for cid in ['join-helper','submission']:assert 'CH23_HELPER_COMPLETE' in actual[cid],actual[cid]
 assert 'helper' in actual['helper-change'] and 'activity' in actual['helper-status']
 sent=json.loads((out/'samples-browser-first.json').read_text())['sent']
 for line in blocks[4][1].splitlines():assert json.loads(line) in sent,line
 assert not errors
 (out/'samples.json').write_text(json.dumps({'result':'PASS','manuscript_sha256':sha,'fences':[{'index':i,'language':k,'sha256':hashlib.sha256(v.encode()).hexdigest()} for i,(k,v) in enumerate(blocks)],'main_requests':len(requests),'judge_requests':len(judges),'grader_fences':[0,5],'grader_receipt':'accepted-completion-corrected.json','limits':'Scripted external decisions establish actual GUI/fleet/tool mechanics, not new autonomous coding or account catalog eligibility. Existing actual Astra/high task evidence remains historical. Grader command fences reuse the recorded same grading invocation with -json output formatting, not a fresh full run.'},indent=2)+'\n')
finally:
 if child is not None and child.poll() is None:child.kill();child.wait()
 (out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n');(out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n');(out/'samples-judges.json').write_text(json.dumps(judges,indent=2)+'\n')
 server.shutdown();server.server_close();shutil.rmtree(profile)
 if work:shutil.rmtree(work)
