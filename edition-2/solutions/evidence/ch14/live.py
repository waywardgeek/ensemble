# Bounded speech-only task through the shipped virtual-user client and GUI.
import json, os, pathlib, shutil, subprocess, tempfile, time, urllib.request
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
providers={'anthropic':('directClaudeAPIKey','claude-haiku-5-5'),'openai':('directOpenAIAPIKey','gpt-4o-mini'),'gemini':('directGeminiAPIKey','gemini-3.8-flash')}
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
for vendor,(field,model) in providers.items():
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch14-live-'))
 key=settings[field]
 def safe(s): return s.replace(key,'[REDACTED]')
 env={'PATH':os.environ['PATH'],'HOME':os.environ['HOME'],'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_API_KEY':key}
 primary=scratch/'skills/target';primary.mkdir(parents=True)
 (primary/'SKILL.md').write_text('---\nname: target\ntype: primary\n---\nAnswer briefly in two sentences. No tools are needed.\n')
 targetenv=dict(env,EN_SKILLS_DIR=str(scratch/'skills'),EN_PRIMARY_SKILL='target')
 target=subprocess.Popen(['/tmp/ensemble-ch14-gui','--port','19415','--gui-dir',str(base/'main/gui/web/gui')],cwd=scratch,env=targetenv,stdin=subprocess.PIPE,stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,text=True)
 result={'vendor':vendor,'model':model}
 try:
  for _ in range(100):
   try:
    urllib.request.urlopen('http://127.0.0.1:19415',timeout=1).close();break
   except Exception:time.sleep(.03)
  driverdir=scratch/'driver';driverdir.mkdir()
  task='Ask the target to explain queueChunk, max_tool_rounds and HTTPServer in two short sentences using inline code markup, and to choose one fruit to mention. Then report exactly the spoken transcript, including its chosen fruit. Use only the five available tools.'
  result['task']=task
  run=subprocess.run(['/tmp/ensemble-ch14-virtual','--agent-url','http://127.0.0.1:19415','--peer',str(base/'main/gui/virtual-user.mjs'),'--browser','/Applications/Google Chrome.app/Contents/MacOS/Google Chrome','--task',task],cwd=driverdir,env=env,capture_output=True,text=True,timeout=100)
  result.update(exit=run.returncode,report=safe(run.stdout),stderr=safe(run.stderr))
  log=driverdir/'virtual-user.jsonl'
  if log.exists():result['driver_events']=[json.loads(safe(line)) for line in log.read_text().splitlines() if line.strip()]
  result['target_events']=[json.loads(safe(line)) for line in (scratch/'event.log').read_text().splitlines() if line.strip()]
  result['target_api']=[json.loads(safe(line)) for line in (scratch/'api.log').read_text().splitlines() if line.strip()]
 except Exception as error:result['error']=safe(str(error))
 finally:
  target.terminate()
  try:target.wait(timeout=8)
  except subprocess.TimeoutExpired:target.kill();target.wait()
  result['target_stderr']=safe(target.stderr.read())
  (base/f'evidence/ch14/{vendor}-live.json').write_text(json.dumps(result,indent=2))
  shutil.rmtree(scratch)
 print(json.dumps({'vendor':vendor,'exit':result.get('exit'),'error':result.get('error'),'report':result.get('report')}),flush=True)
