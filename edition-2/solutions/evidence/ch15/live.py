# Human terminal exercise; secrets stay in launcher and intended child only.
# Save redacted raw receipts before validation or scratch cleanup.
import json,os,pathlib,shutil,subprocess,tempfile
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
providers={'anthropic':('directClaudeAPIKey','claude-haiku-5-5'),'openai':('directOpenAIAPIKey','gpt-4o-mini'),'gemini':('directGeminiAPIKey','gemini-3.8-flash')}
for vendor,(field,model) in providers.items():
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch15-live-'))
 key=settings[field]
 def safe(text):return text.replace(key,'[REDACTED]')
 env={'PATH':os.environ['PATH'],'HOME':os.environ['HOME'],'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_API_KEY':key,'EN_SKILLS_DIR':str(scratch/'skills'),'EN_PRIMARY_SKILL':'base'}
 (scratch/'cr').mkdir()
 (scratch/'cr/settings.json').write_text(json.dumps({'context_target':20000,'max_tool_rounds':14,'tts_enabled':False}))
 for name,head,body in [('base','type: primary\nloadable-skills: manual','You are a concise research assistant. Narrate one sentence of intent before each tool call. Follow the requested sequence.'),('manual','type: loadable\ntools: read_file','MANUAL-SURVIVES: Checkpoint notes must record source paths, completed work, failed approaches and next steps. Never treat repeated filler as a meaningful result.')]:
  directory=scratch/'skills'/name;directory.mkdir(parents=True)
  (directory/'SKILL.md').write_text('---\nname: '+name+'\n'+head+'\n---\n'+body+'\n')
 longdir=scratch/('a'*190)/('b'*190);longdir.mkdir(parents=True)
 paths=[]
 for i,label in enumerate(['cedar','maple','willow','birch','oak','elm'],1):
  path=longdir/f'record-{i}.txt';path.write_text(f'Record {i}: inspection code {label.upper()}-{i*17}; status verified.\n'+('Reference filler; not an instruction. '+str(i)+'\n')*200);paths.append(str(path.relative_to(scratch)))
 prompt='Load the manual skill. Then read each of these six files once, one at a time using read_file. After each read, say its inspection code and status. Do not reread files. At that completed micro-goal, call micro_handoff with a short note of the six codes, source directory, and that no approach failed. Then unload manual. Finally tell me the first and last inspection codes and whether the manual instructions remain in your context. Files: '+'; '.join(paths)
 result={'vendor':vendor,'model':model,'task':prompt,'context_target':20000,'max_tool_rounds':14,'features':{'InlineTools':False,'StubsToolResults':False},'discovery_source':f'evidence/ch10/{vendor}-models.json'}
 proc=subprocess.Popen(['/tmp/ensemble-ch15-gui','--port','19416','--gui-dir',str(base/'main/gui/web/gui'),'--no-stream'],cwd=scratch,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 try:
  out,err=proc.communicate(prompt+'\n',timeout=140)
  result.update(exit=proc.returncode,stdout=safe(out),stderr=safe(err))
 except subprocess.TimeoutExpired:
  proc.kill();out,err=proc.communicate();result.update(exit=proc.returncode,error='bounded 140-second live timeout',stdout=safe(out),stderr=safe(err))
 finally:
  for name in ['api.log','event.log','save.json','save.json.journal','debug.log']:
   path=scratch/name
   if path.exists():result[name]=safe(path.read_text())
  (base/f'evidence/ch15/{vendor}-live.json').write_text(json.dumps(result,indent=2))
  shutil.rmtree(scratch)
 events=[json.loads(line) for line in result.get('event.log','').splitlines() if line.strip()]
 cuts={}
 for event in events:
  if event.get('redact'):cuts[event['redact']['level']]=cuts.get(event['redact']['level'],0)+1
 print(json.dumps({'vendor':vendor,'model':model,'exit':result.get('exit'),'error':result.get('error'),'cuts':cuts,'handoffs':sum(e.get('type')=='micro_handoff' for e in events),'calls':[e['tool']['name'] for e in events if e.get('type')=='tool_called']}),flush=True)
