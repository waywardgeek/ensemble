import concurrent.futures,datetime,hashlib,json,os,pathlib,subprocess,tempfile
base=pathlib.Path(__file__).resolve().parent
binary=base/'bin/ch03-revised'
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
routes=[('anthropic','claude-sonnet-5-5','claude-sonnet-5-5','directClaudeAPIKey'),('openai','gpt-4.1-mini-2025-04-14','gpt-4.1-mini-2025-04-14','directOpenAIAPIKey'),('gemini','models/gemini-3.8-flash','gemini-3.8-flash','directGeminiAPIKey')]
prompts=["Execute these exact authorized tests in the scratch workspace: call read_file on empty.txt with no range; call read_file on empty.txt with end_line:0 and no start_line (both must succeed with empty content); call read_file on empty.txt with start_line:1 to observe the explicit-line failure; run_command with command 'exit 7' to observe the silent shell's actual status. After doing all four calls, invent a short two-word code name and reply ONLY with that name. Do not skip the intentional failing read.","What exact two-word code name did you invent in your previous answer? Reply only with precisely that name."]
sources={str(p.relative_to(base.parent.parent)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((base.parent.parent).rglob('*.go')) if '.git' not in p.parts}
def run(route):
 vendor,model,resolved,key=route;dest=base/('revision-'+vendor);dest.mkdir()
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch03-revision-'+vendor+'-'));(scratch/'empty.txt').write_text('')
 safe={'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_RESOLVED_MODEL':resolved,'CH02_LOG':str(dest/'session.log')}
 env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))};env.update(safe);env['LLM_API_KEY']=settings[key]
 inp=''.join(json.dumps({'user':p})+'\n' for p in prompts);(dest/'stdin.jsonl').write_text(inp)
 receipt={'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'cwd':str(scratch),'argv':[str(binary)],'environment':dict(safe,LLM_API_KEY='<child environment; redacted>'),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'sources_sha256':sources}
 p=subprocess.run([str(binary)],cwd=scratch,env=env,input=inp,capture_output=True,text=True,timeout=180)
 receipt['exit_code']=p.returncode;receipt['finished_utc']=datetime.datetime.now(datetime.timezone.utc).isoformat();(dest/'stdout.jsonl').write_text(p.stdout);(dest/'stderr.txt').write_text(p.stderr)
 answers=[json.loads(l)['assistant'] for l in p.stdout.splitlines() if 'assistant' in json.loads(l)];receipt['recall_equal']=len(answers)==2 and answers[0]==answers[1];receipt['empty_file_bytes']=(scratch/'empty.txt').stat().st_size
 (dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n');print(vendor,p.returncode,receipt['recall_equal'],p.stdout,flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(run,routes))
