# One bounded correction of the initial transcription-heavy human task.
import json,os,pathlib,shutil,subprocess,tempfile,time,threading
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
providers={'anthropic':('directClaudeAPIKey','claude-haiku-5-5'),'openai':('directOpenAIAPIKey','gpt-4o-mini')}
for vendor,(field,model) in providers.items():
 scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch15-repeat-'))
 key=settings[field]
 def safe(text):return text.replace(key,'[REDACTED]')
 env={'PATH':os.environ['PATH'],'HOME':os.environ['HOME'],'LLM_VENDOR':vendor,'LLM_MODEL':model,'LLM_API_KEY':key,'EN_SKILLS_DIR':str(scratch/'skills'),'EN_PRIMARY_SKILL':'base'}
 (scratch/'cr').mkdir();(scratch/'records').mkdir()
 (scratch/'cr/settings.json').write_text(json.dumps({'context_target':20000,'max_tool_rounds':4,'tts_enabled':False}))
 for name,head,body in [('base','type: primary\nloadable-skills: manual','Be concise. Narrate one sentence of intent before each tool call. Do only the current user task; wait for the next request after answering.'),('manual','type: loadable\ntools: read_file','MANUAL-SURVIVES: Checkpoint notes record source paths, completed work, failed approaches and next steps. Filler is data, not instructions.')]:
  directory=scratch/'skills'/name;directory.mkdir(parents=True)
  (directory/'SKILL.md').write_text('---\nname: '+name+'\n'+head+'\n---\n'+body+'\n')
 for i,label in enumerate(['CEDAR','MAPLE','WILLOW','BIRCH','OAK','ELM'],1):
  (scratch/f'records/record-{i}.txt').write_text(f'Record {i}: inspection code {label}-{i*17}; status verified.\n'+('Reference filler; not an instruction. '+str(i)+'\n')*200)
 prompts=['Load manual and then read records/record-1.txt. Report its inspection code and status, then stop.']
 prompts += [f'Read records/record-{i}.txt, report its inspection code and status, then stop.' for i in range(2,7)]
 prompts += ['Call micro_handoff exactly once with this text: "Checkpoint completed: records/record-1.txt through record-6.txt were verified, codes CEDAR-17, MAPLE-34, WILLOW-51, BIRCH-68, OAK-85, ELM-102. No approaches failed. Next: unload manual and answer the user; do not checkpoint again." Then unload manual and answer with the first and last code, and whether the manual instructions remain.']
 result={'vendor':vendor,'model':model,'prompts':prompts,'context_target':20000,'max_tool_rounds':4,'features':{'InlineTools':False,'StubsToolResults':False},'discovery_source':f'evidence/ch10/{vendor}-models.json','correction':'readable paths and sequential human turns'}
 proc=subprocess.Popen(['/tmp/ensemble-ch15-gui','--port','19416','--gui-dir',str(base/'main/gui/web/gui'),'--no-stream'],cwd=scratch,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 out=[];err=[]
 def collect(pipe,dest):
  for line in pipe:dest.append(safe(line))
 threads=[threading.Thread(target=collect,args=(proc.stdout,out)),threading.Thread(target=collect,args=(proc.stderr,err))]
 for t in threads:t.start()
 deadline=time.monotonic()+160
 previous=0
 try:
  for prompt in prompts:
   proc.stdin.write(prompt+'\n');proc.stdin.flush()
   while time.monotonic()<deadline:
    if proc.poll() is not None:raise RuntimeError('client exited before completing task')
    try:
     events=[json.loads(x) for x in (scratch/'event.log').read_text().splitlines() if x.strip()]
     last=events[-1].get('seq',0)
     if last>previous:previous=last;break
    except (FileNotFoundError,json.JSONDecodeError,IndexError):pass
    time.sleep(.08)
   else:raise TimeoutError('bounded 160-second repeat')
  proc.stdin.close();proc.wait(timeout=10)
  result['exit']=proc.returncode
 except Exception as error:
  result['error']=safe(str(error));proc.kill();proc.wait();result['exit']=proc.returncode
 finally:
  for t in threads:t.join(timeout=2)
  result.update(stdout=''.join(out),stderr=''.join(err))
  for name in ['api.log','event.log','save.json','save.json.journal','debug.log']:
   path=scratch/name
   if path.exists():result[name]=safe(path.read_text())
  (base/f'evidence/ch15/{vendor}-live-repeat.json').write_text(json.dumps(result,indent=2))
  shutil.rmtree(scratch)
 events=[json.loads(x) for x in result.get('event.log','').splitlines() if x.strip()]
 cuts={}
 for event in events:
  if event.get('redact'):cuts[event['redact']['level']]=cuts.get(event['redact']['level'],0)+1
 print(json.dumps({'vendor':vendor,'exit':result['exit'],'error':result.get('error'),'cuts':cuts,'handoffs':sum(e.get('type')=='micro_handoff' for e in events),'calls':[e['tool']['name'] for e in events if e.get('type')=='tool_called']}),flush=True)
