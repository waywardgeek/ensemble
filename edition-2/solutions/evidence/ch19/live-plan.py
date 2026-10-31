"""One bounded human CLI plan session using only Bill's dedicated new profile."""
import hashlib,json,os,pathlib,subprocess,sys,tempfile,time

def completed_answer(events):
 # A public summary is optional. After a completed answer, EOF must not wait
 # for a summary that the provider chose not to emit. Tool rounds continue.
 for line in events.splitlines():
  try:event=json.loads(line)
  except ValueError:continue
  if event.get('type')!='response_ended':continue
  parts=event.get('response',{}).get('parts',[])
  if any(p.get('type')=='text' for p in parts) and not any(p.get('type')=='tool_call' for p in parts):return True
 return False

repo=pathlib.Path.cwd(); evidence=repo/'solutions/edition-2/evidence/ch19'
handoff=json.loads((evidence/(sys.argv[1] if len(sys.argv)>1 else 'signin-handoff.json')).read_text())
prefix=sys.argv[2] if len(sys.argv)>2 else 'live-plan'
binary=pathlib.Path(handoff['binary']); auth=pathlib.Path(handoff['fresh_en_home'])
assert hashlib.sha256(binary.read_bytes()).hexdigest()==handoff['binary_sha256']
# The app logs no auth exchange/headers and redacts active/replaced bearers before
# callbacks. This capture additionally excludes this dedicated account's values.
# No credential or identity field is printed or passed to another process.
private=[]
def remember_private():
 value=json.loads((auth/'auth.json').read_text())
 for field in ['host_id','active','pending']:
  if value.get(field):private.append(value[field].encode())
 for profile in value['profiles'].values():
  for field in ['access_token','refresh_token','id_token','email','subject','client_id']:
   if profile.get(field):private.append(profile[field].encode())
remember_private()
with tempfile.TemporaryDirectory(prefix='ensemble-ch19-plan-run-') as temporary:
 work=pathlib.Path(temporary)
 (work/'puzzle.txt').write_text('Two identical workers, nonpreemptive jobs. Durations and prerequisites:\nA: 4, none\nB: 3, none\nC: 5, A\nD: 2, A\nE: 4, B\nF: 3, C and D\nG: 2, D and E\nH: 1, F and G\nFind a minimum-makespan valid schedule. Justify optimality, not merely validity.\n')
 env={k:v for k,v in os.environ.items() if not k.endswith('API_KEY')}
 env.update(LLM_VENDOR='openai',LLM_MODEL='gpt-6.1-sol',LLM_BASE_URL='https://api.openai.com',OPENAI_BILLING_MODE='chatgpt',OPENAI_OIDC_ISSUER='https://auth.openai.com',OPENAI_REASONING_EFFORT='high',OPENAI_REASONING_SUMMARY='detailed',EN_HOME=str(auth),EN_RECALL_DISABLED='1',EN_PRIMARY_SKILL='',EN_SKILLS_DIR='',CH02_LOG=str(work/'events.jsonl'))
 start=time.monotonic(); hint=False; closed=False; timed_out=False; answer_closed=False
 with open(work/'stdout.txt','wb') as stdout,open(work/'stderr.txt','wb') as stderr:
  child=subprocess.Popen([str(binary),'chat'],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr)
  child.stdin.write(b'Read puzzle.txt with read_file, solve its scheduling problem, and give a concise schedule and proof. Do not modify any files.\n');child.stdin.flush()
  while child.poll() is None and time.monotonic()-start<150:
   trace=work/'api.log'
   if not closed and trace.exists() and 'response.reasoning_summary_text.delta' in trace.read_text():
    child.stdin.write(b'Keep both workers explicit in the answer, and verify every dependency before claiming optimality.\n');child.stdin.flush();child.stdin.close();hint=True;closed=True
   if not closed and 'request failed:' in (work/'stderr.txt').read_text():child.stdin.close();closed=True
   events=work/'events.jsonl'
   if not closed and events.exists() and completed_answer(events.read_text()):
    child.stdin.close();closed=True;answer_closed=True
   time.sleep(.1)
  if child.poll() is None:child.kill();timed_out=True
  code=child.wait()
 remember_private()
 for name in ['stdout.txt','stderr.txt','api.log','events.jsonl','save.json']:
  path=work/name
  if path.exists():
   raw=path.read_bytes()
   for value in set(private):raw=raw.replace(value,b'[REDACTED]')
   assert not any(value in raw for value in private)
   (evidence/(prefix+'-'+name)).write_bytes(raw)
 result={'source_commit':handoff['source_commit'],'main_tree':handoff['main_tree'],'binary_sha256':handoff['binary_sha256'],'model':'gpt-6.1-sol','billing':'explicit chatgpt','api_key_configured':False,'reasoning_effort':'high','reasoning_summary':'detailed','hint_sent_after_summary':hint,'closed_on_completed_answer':answer_closed,'exit_code':code,'elapsed_seconds':round(time.monotonic()-start,2),'timed_out':timed_out,'dedicated_profile_retained':True,'logout':False}
 (evidence/(prefix+'-result.json')).write_text(json.dumps(result,indent=2)+'\n')
 print(json.dumps(result,indent=2))
