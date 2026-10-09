import concurrent.futures,datetime,hashlib,json,os,pathlib,subprocess
base=pathlib.Path(__file__).resolve().parent
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
binary=base/'bin/tools-consumer'
routes=[('anthropic','claude-sonnet-5-5','claude-sonnet-5-5','directClaudeAPIKey'),('openai','gpt-4.1-mini-2025-04-14','gpt-4.1-mini-2025-04-14','directOpenAIAPIKey'),('gemini','models/gemini-3.8-flash','gemini-3.8-flash','directGeminiAPIKey')]
def run(route):
 vendor,model,resolved,key=route;dest=base/('consumer-'+vendor);dest.mkdir()
 env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
 safe={'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_RESOLVED_MODEL':resolved,'DEMO_DIR':str(dest)}
 env.update(safe);env['LLM_API_KEY']=settings[key]
 receipt={'started_utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'argv':[str(binary)],'environment':dict(safe,LLM_API_KEY='<child environment; redacted>'),'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'consumer_source_sha256':hashlib.sha256((base.parent.parent/'examples/tools-consumer/main.go').read_bytes()).hexdigest()}
 try:
  p=subprocess.run([str(binary)],env=env,capture_output=True,text=True,timeout=180)
  receipt['exit_code']=p.returncode;(dest/'stdout.jsonl').write_text(p.stdout);(dest/'stderr.txt').write_text(p.stderr)
 except subprocess.TimeoutExpired:receipt['timeout_seconds']=180
 receipt['finished_utc']=datetime.datetime.now(datetime.timezone.utc).isoformat();(dest/'receipt.json').write_text(json.dumps(receipt,indent=2)+'\n');print(vendor,receipt.get('exit_code','timeout'),flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(run,routes))
