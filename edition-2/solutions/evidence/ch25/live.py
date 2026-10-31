"""One bounded human CLI fleet task using the authorized dedicated subscription profile."""
import hashlib,json,os,pathlib,shutil,subprocess,tempfile,time
out=pathlib.Path(__file__).resolve().parent
source=out.parents[1]/'main'
work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch25-live-'))
binary=work/'ensemble'
subprocess.run(['go','build','-o',str(binary),'.'],cwd=source,check=True)
source_tree=subprocess.check_output(['git','rev-parse','HEAD:solutions/edition-2/main'],cwd=source,text=True).strip()
prompt='''Complete this bounded fleet coding exercise in the current scratch workspace. Spawn exactly three children with IDs add, multiply, subtract, inheriting the current model and reasoning effort. Give each child one separate task: write add.py containing add(a,b), multiply.py containing multiply(a,b), or subtract.py containing subtract(a,b), then run a real python3 assertion for that function after a four-second shell sleep, then submit its result. Each child must use write_file and run_command. The add child must FIRST ask you with send_message_to_parent (ai_callback_delay 60) whether negative integers are required; answer yes with respond_to_parent_message. All three support integers including negatives. Each child must submit explicitly and spawn no children. You must call join_agents with ai_callback_delay 60 and answer any child question, then join again. When all three submissions arrive, run one combined python3 assertion importing all three modules and checking add(-2,5)==3, multiply(-2,5)==-10 and subtract(-2,5)==-7, printing COMBINED_TESTS_PASS. Finally submit a concise result beginning FLEET_DONE and include any later steering marker you received. Do not use web tools or create additional children. Keep the task small; do not add tests or files beyond these three Python files.'''
env={k:os.environ[k] for k in ('PATH','HOME','TMPDIR') if k in os.environ}
env.update(EN_HOME='/Users/bill/ensemble-ch19-signin-01v0rmp3/auth',OPENAI_BILLING_MODE='chatgpt',OPENAI_OIDC_ISSUER='https://auth.openai.com',LLM_VENDOR='openai',LLM_BASE_URL='https://api.openai.com',LLM_MODEL='gpt-6-astra',OPENAI_REASONING_EFFORT='high',EN_RECALL_DISABLED='1')
# No API key is read, copied or inherited; headers and credential stores are not captured.
stdout=(out/'live-stdout.txt').open('w');stderr=(out/'live-stderr.txt').open('w')
start=time.monotonic()
p=subprocess.Popen([str(binary),'chat'],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
p.stdin.write(prompt+'\n');p.stdin.flush()
steering=None;submission=False;events=[];status='deadline'
def facts():
 path=work/'save.json.journal'
 if not path.exists():return []
 found=[]
 for line in path.read_text().splitlines():
  try:found.append(json.loads(line))
  except json.JSONDecodeError:pass
 return found
try:
 while time.monotonic()-start<150:
  if p.poll() is not None:status='process_exit';break
  events=facts()
  returned={e.get('tool',{}).get('call_id'):e for e in events if e.get('type')=='tool_returned'}
  joins=[e for e in events if e.get('type')=='tool_called' and e.get('tool',{}).get('name')=='join_agents' and e['tool'].get('call_id') not in returned]
  if steering is None and joins:
   time.sleep(.12)
   after=facts()
   call=joins[-1]['tool']['call_id']
   if not any(e.get('type')=='tool_returned' and e.get('tool',{}).get('call_id')==call for e in after):
    steering={'seconds':round(time.monotonic()-start,3),'join_call_id':call,'join_event_time':joins[-1].get('time'),'text':'STEER_COBALT: preserve the negative-integer checks and include STEER_COBALT in your final submission.'}
    p.stdin.write(steering['text']+'\n');p.stdin.flush()
  for e in events:
   tool=e.get('tool',{})
   if e.get('type')=='tool_called' and tool.get('name')=='submit' and tool.get('call_id') in returned and not returned[tool['call_id']].get('tool',{}).get('is_error'):
    submission=True;status='submitted';break
  if submission:break
  time.sleep(.025)
 p.stdin.close()
 try:p.wait(timeout=10)
 except subprocess.TimeoutExpired:p.terminate();p.wait(timeout=5);status+=';terminated_after_grace'
finally:
 if p.poll() is None:p.kill();p.wait()
 stdout.close();stderr.close()
# Preserve artifacts before parsing claims; no auth profile enters this directory.
for name in ('save.json','ch02.jsonl','api.log','add.py','multiply.py','subtract.py'):
 if (work/name).exists():shutil.copyfile(work/name,out/('live-api.txt' if name == 'api.log' else 'live-'+name))
children=work/'cr'/'children'
if children.exists():
 for child in children.iterdir():
  if (child/'save.json').exists():shutil.copyfile(child/'save.json',out/('live-child-'+child.name+'-save.json'))
wire=[]
if (work/'api.log').exists():
 for line in (work/'api.log').read_text().splitlines():
  try:
   row=json.loads(line);data=json.loads(row.get('data','{}'))
   if row.get('event')=='request':wire.append({'model':data.get('model'),'reasoning':data.get('reasoning')})
  except (json.JSONDecodeError,TypeError):pass
check=subprocess.run(['python3','-c','from add import add; from multiply import multiply; from subtract import subtract; assert add(-2,5)==3; assert multiply(-2,5)==-10; assert subtract(-2,5)==-7; print("COMBINED_TESTS_PASS")'],cwd=work,text=True,capture_output=True)
summary={'source_tree':source_tree,'source_commit':'891d513bf142013d6888760cfae4fcaab1edd680','binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'scratch':str(work),'status':status,'exit':p.returncode,'seconds':round(time.monotonic()-start,3),'submitted':submission,'steering':steering,'wire_requests':wire,'independent_test_exit':check.returncode,'independent_test_stdout':check.stdout,'independent_test_stderr':check.stderr}
(out/'live-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({k:summary[k] for k in ('status','exit','seconds','submitted','steering','independent_test_exit')},indent=2),flush=True)
print('wire_request_records',len(wire),flush=True)
