"""Three finite human CLI MCP turns; keys stay only in the model client's env."""
import json, os, pathlib, subprocess, tempfile, time
out=pathlib.Path(__file__).resolve().parent
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
providers=[('anthropic','directClaudeAPIKey','claude-haiku-5-5'),('openai','directOpenAIAPIKey','gpt-4o-mini'),('gemini','directGeminiAPIKey','gemini-3.8-flash')]
server='''import json,sys,pathlib
counts={}
for line in sys.stdin:
 m=json.loads(line); method=m.get('method')
 if 'id' not in m: continue
 if method=='initialize': result={'protocolVersion':'2024-11-05','capabilities':{'tools':{}}}
 elif method=='tools/list': result={'tools':[{'name':n,'description':d,'inputSchema':{'type':'object','properties':{}},'ephemeral':mode} for n,d,mode in [('read_marker','Read the exact marker file text',''),('environment_snapshot','Current live snapshot','round'),('session_brief','Session configuration','turn')]]}
 elif method=='tools/call':
  n=m['params']['name']; counts[n]=counts.get(n,0)+1
  pathlib.Path('peer-counts.json').write_text(json.dumps(counts))
  text=pathlib.Path('marker.txt').read_text() if n=='read_marker' else (f'Snapshot number {counts[n]}' if n=='environment_snapshot' else 'Use read_marker to inspect the local marker.')
  result={'content':[{'type':'text','text':text}]}
 else: continue
 print(json.dumps({'jsonrpc':'2.0','id':m['id'],'result':result}),flush=True)
'''
for vendor,field,model in providers:
 if os.getenv("CH12_PROVIDER") and os.getenv("CH12_PROVIDER") != vendor: continue
 with tempfile.TemporaryDirectory(prefix='ensemble-ch12-live-') as scratch:
  d=pathlib.Path(scratch); marker='CH12_'+vendor.upper()+'_MCP_OK'
  (d/'marker.txt').write_text(marker); (d/'peer.py').write_text(server)
  (d/'skills/base').mkdir(parents=True)
  (d/'skills/base/SKILL.md').write_text('---\nname: base\ntype: primary\ntools: read_marker environment_snapshot session_brief\nmcp_servers:\n  - name: local-observer\n    transport: stdio\n    command: python3\n    args: ["-u", "peer.py"]\n---\nYou are a concise assistant. Read the marker through read_marker when requested. Automatic snapshot data describes the current round.')
  env={k:v for k,v in os.environ.items() if k not in ['LLM_API_KEY','ANTHROPIC_API_KEY','OPENAI_API_KEY','GEMINI_API_KEY','LLM_BASE_URL','EN_PRIMARY_SKILL','EN_SKILLS_DIR']}
  env.update(LLM_VENDOR=vendor,LLM_MODEL=model,LLM_API_KEY=settings[field],EN_SKILLS_DIR=str(d/'skills'),EN_PRIMARY_SKILL='base')
  prompt='Use read_marker exactly once. Report the exact marker and current automatic snapshot number in one sentence. Do no other work.\n'
  started=time.monotonic(); timed_out=False
  child=subprocess.Popen(['/tmp/ensemble-ch12','--save',str(d/'save.json'),'chat'],cwd=scratch,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
  try: stdout,stderr=child.communicate(prompt,timeout=60)
  except subprocess.TimeoutExpired:
   timed_out=True;child.terminate();stdout,stderr=child.communicate(timeout=5)
  def safe(text):
   for _,f,_ in providers:
    if settings.get(f):text=text.replace(settings[f],'[REDACTED]')
   return text
  requests=[]; frames=[]
  if (d/'api.log').exists():
   for line in safe((d/'api.log').read_text()).splitlines():
    e=json.loads(line)
    if e.get('event')=='request':requests.append(json.loads(e['data']))
    else: frames.append(e)
  counts=json.loads((d/'peer-counts.json').read_text()) if (d/'peer-counts.json').exists() else {}
  data={'vendor':vendor,'model':model,'model_discovery':'Reuse same-day Chapter10 model receipt; this successful inference is live availability proof','prompt':prompt,'exit':child.returncode,'timeout':timed_out,'elapsed_seconds':round(time.monotonic()-started,2),'stdout':safe(stdout),'stderr':safe(stderr),'peer_counts':counts,'requests':requests,'frames':frames,'marker_reported':marker in stdout}
  (out/(vendor+'-live.json')).write_text(json.dumps(data,indent=2)+'\n')
  print(vendor,'exit',child.returncode,'timeout',timed_out,'marker_reported',data['marker_reported'],'peer_counts',counts,flush=True)
