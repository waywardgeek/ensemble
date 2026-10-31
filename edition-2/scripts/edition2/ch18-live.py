"""Bounded real-browser Chapter18 demonstration; never expose credentials."""
import json,os,pathlib,subprocess,tempfile,time,socket,urllib.request,urllib.error,shutil,signal
root=pathlib.Path(__file__).resolve().parents[2]
output=root/'solutions/edition-2/evidence/ch18'
binary=pathlib.Path(tempfile.gettempdir())/'ensemble-ch18-gui'
subprocess.run(['go','build','-o',str(binary),'./cmd'],cwd=root/'solutions/edition-2/main/gui',check=True)
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
models=[('anthropic','claude-haiku-5-5','directClaudeAPIKey','https://api.anthropic.com/v1/models'),('openai','gpt-5.6-sol','directOpenAIAPIKey','https://api.openai.com/v1/models'),('gemini','gemini-3.8-flash','directGeminiAPIKey','https://generativelanguage.googleapis.com/v1beta/models')]
for vendor,model,field,endpoint in models:
 key=settings[field]
 headers={'x-api-key':key,'anthropic-version':'2023-06-01'} if vendor=='anthropic' else {'Authorization':'Bearer '+key} if vendor=='openai' else {'x-goog-api-key':key}
 try:
  with urllib.request.urlopen(urllib.request.Request(endpoint,headers=headers),timeout=30) as response: inventory=json.load(response)
  entries=inventory.get('data',inventory.get('models',[]))
  present=any(item.get('id',item.get('name','')).removeprefix('models/')==model for item in entries)
  (output/f'{vendor}-discovery.json').write_text(json.dumps({'model':model,'listed':present}))
  print(vendor,'model listed',present,flush=True)
  if not present: continue
 except Exception as error:
  (output/f'{vendor}-discovery.json').write_text(json.dumps({'model':model,'error_type':type(error).__name__,'status':getattr(error,'code',None)}))
  print(vendor,'discovery failed',type(error).__name__,flush=True);continue
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch18-'+vendor+'-'))
 skill=scratch/'skills/base';skill.mkdir(parents=True)
 manual='---\nname: base\ntype: primary\ntools: read_file\n---\nYou are the archive assistant. Answer questions from this reference briefly; no tools are needed unless the user explicitly asks. The project code is COPPER-FERRY-18.\n'
 manual+='\n'.join(f'Archive record {i}: the color is copper; preserve the original source and record the reason for each change. This reference is factual data for the archive project.' for i in range(500))
 (skill/'SKILL.md').write_text(manual)
 (scratch/'cr').mkdir();(scratch/'cr/settings.json').write_text(json.dumps({'tts_enabled':False}))
 listener=socket.socket();listener.bind(('127.0.0.1',0));port=listener.getsockname()[1];listener.close()
 env=os.environ.copy()
 for name in list(env):
  if 'API_KEY' in name:env.pop(name)
 env.update(LLM_API_KEY=key,LLM_VENDOR=vendor,LLM_MODEL=model,EN_SKILLS_DIR=str(scratch/'skills'),EN_PRIMARY_SKILL='base')
 with (scratch/'stdout').open('w') as stdout,(scratch/'stderr').open('w') as stderr:
  child=subprocess.Popen([str(binary),'--port',str(port),'--gui-dir',str(root/'solutions/edition-2/main/gui/web/gui')],cwd=scratch,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr)
  try:
   for _ in range(100):
    try:
     urllib.request.urlopen(f'http://127.0.0.1:{port}',timeout=.5).close();break
    except Exception:time.sleep(.1)
   run=subprocess.run(['node',str(root/'scripts/edition2/ch18-browser.mjs'),f'http://127.0.0.1:{port}',str(output/f'{vendor}-browser.json')],cwd=root,timeout=355,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
   print(vendor,'browser exit',run.returncode,flush=True)
  finally:
   child.send_signal(signal.SIGINT)
   try:child.wait(timeout=15)
   except subprocess.TimeoutExpired:child.kill();child.wait()
 # Preserve raw sanitized receipts before any parsing or scratch cleanup.
 for name in ['api.log','debug.log','gui.log','event.log','save.json','stdout','stderr','cr/cache-current.txt','cr/cache-prior.txt']:
  path=scratch/name
  if path.exists():(output/(vendor+'-'+name.replace('/','-'))).write_text(path.read_text().replace(key,'[REDACTED]'))
 shutil.rmtree(scratch)
