"""Bounded CLI demonstrations. Credentials stay in memory and child environment."""
import concurrent.futures,datetime,hashlib,json,os,pathlib,subprocess,tempfile
base=pathlib.Path(__file__).resolve().parent
binary=base/'bin/ch03-initial'
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
prompts=[
'This is an authorized tool demonstration in a scratch workspace. Use all requested tools, then answer briefly. Remember the code name AMBER-CH03. First list_directory at .; read_file notes.md lines 3 through 4; search_files for beta|delta in notes.md with context_lines 1. Then write_file created.txt with exactly "first anchor\\nsecond anchor\\n" (actual newlines), edit_file its unique "first anchor" to "first edit", and run_command "cat created.txt; printf STDOUT-MARKER; printf STDERR-MARKER >&2; exit 7". Report the exit status and code name.',
'Exercise the overwrite guard deliberately. First call write_file on existing created.txt with content "replacement\\n" (actual newline) and leave overwrite omitted/false and append false; it MUST refuse. After seeing that error, repeat with overwrite true, then append "tail\\n" using append true, then read_file created.txt. Do not skip the initial refused write. Report the final exact two lines.',
'Exercise exact-anchor recovery deliberately. First write_file anchors.txt containing exactly "first anchor\\nsecond anchor\\n" (actual newlines). Then call edit_file with old_text "anchor" and new_text "wrong" to observe the two-match refusal. Also attempt old_text "missing" to observe zero matches and read_file absent-file.txt to observe missing-file error. Recover by edit_file old_text "first anchor" new_text "first repaired". Finally read_file anchors.txt and run_command "cat anchors.txt". Do not skip the deliberate failures; these are requested tests. Report what failed and the final content.',
'Without tools, what exact code name did I ask you to remember in the first turn, and what exit status did the first shell command return? Reply only with the name and status.'
]
routes=[('anthropic','claude-sonnet-5-5','claude-sonnet-5-5','directClaudeAPIKey'),('openai','gpt-4.1-mini-2025-04-14','gpt-4.1-mini-2025-04-14','directOpenAIAPIKey'),('gemini','models/gemini-3.8-flash','gemini-3.8-flash','directGeminiAPIKey')]
def run(route):
 vendor,model,resolved,key=route
 dest=base/('live-'+vendor);dest.mkdir(exist_ok=False)
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch03-'+vendor+'-'))
 notes='alpha line one\nbeta line two\ngamma line three\ndelta line four\nepsilon line five\n'
 (scratch/'notes.md').write_text(notes)
 inp=''.join(json.dumps({'user':p})+'\n' for p in prompts);(dest/'stdin.jsonl').write_text(inp)
 env={k:v for k,v in os.environ.items() if not (k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_')) or k=='CH02_LOG')}
 safe={'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_RESOLVED_MODEL':resolved,'CH02_LOG':str(dest/'session.log')}
 env.update(safe);env['LLM_API_KEY']=settings[key]
 receipt={'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'cwd':str(scratch),'argv':[str(binary)],'environment':dict(safe,LLM_API_KEY='<inherited in child environment; redacted>'),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'initial_files':{'notes.md':notes}}
 try:
  result=subprocess.run([str(binary)],cwd=scratch,env=env,input=inp,text=True,capture_output=True,timeout=480)
  receipt['exit_code']=result.returncode;(dest/'stdout.jsonl').write_text(result.stdout);(dest/'stderr.txt').write_text(result.stderr)
 except subprocess.TimeoutExpired as e:
  receipt['timeout_seconds']=480
  for name,v in [('stdout.jsonl',e.stdout),('stderr.txt',e.stderr)]: (dest/name).write_text(v.decode() if isinstance(v,bytes) else v or '')
 receipt['finished_utc']=datetime.datetime.now(datetime.timezone.utc).isoformat()
 receipt['final_files']={p.name:p.read_text() for p in scratch.iterdir() if p.is_file()}
 if (dest/'session.log').exists():
  events=[json.loads(line) for line in (dest/'session.log').read_text().splitlines()][1:]
  receipt['calls']=[e['tool'] for e in events if e['type']=='tool_called']
  receipt['results']=[{'call_id':e['tool']['call_id'],'is_error':e['tool'].get('is_error',False)} for e in events if e['type']=='tool_returned']
  receipt['responses']=[{'from':e['response']['from'],'requested':e['response'].get('requested'),'usage':e['response']['usage']} for e in events if e['type']=='response_ended']
 (dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n')
 print(vendor,'exit',receipt.get('exit_code','timeout'),'calls',len(receipt.get('calls',[])),'outputs',(dest/'stdout.jsonl').read_text(),flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(run,routes))
