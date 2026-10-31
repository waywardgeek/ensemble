"""Bounded human CLI persistence exercise; provider secrets stay in child env."""
import json, os, pathlib, subprocess, tempfile, time
out = pathlib.Path(__file__).resolve().parent
settings = json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
providers = [('anthropic','directClaudeAPIKey','claude-haiku-5-5'),('openai','directOpenAIAPIKey','gpt-4o-mini'),('gemini','directGeminiAPIKey','gemini-3.8-flash')]
def safe(text):
    for _,field,_ in providers:
        key=settings.get(field)
        if key: text=text.replace(key,'[REDACTED]')
    return text
for vendor,field,model in providers:
    with tempfile.TemporaryDirectory(prefix='ensemble-ch11-live-') as scratch:
        directory=pathlib.Path(scratch)
        env={k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_','EN_','CH02_','CH07_'))}
        env.update(LLM_VENDOR=vendor,LLM_MODEL=model,LLM_API_KEY=settings[field])
        runs=[]
        def run(label,prompt):
            started=time.monotonic()
            child=subprocess.Popen(['/tmp/ensemble-ch11','chat'],cwd=scratch,env=env,stdin=subprocess.PIPE,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
            timed_out=False
            try: stdout,stderr=child.communicate(prompt+'\n',timeout=50)
            except subprocess.TimeoutExpired:
                timed_out=True; child.terminate(); stdout,stderr=child.communicate(timeout=5)
            requests=[]
            trace=directory/'api.log'
            if trace.exists():
                for line in safe(trace.read_text()).splitlines():
                    record=json.loads(line)
                    if record.get('event')=='request':
                        body=json.loads(record['data'])
                        requests.append({k:v for k,v in body.items() if k in ('model','messages','contents')})
            runs.append({'label':label,'prompt':prompt,'exit':child.returncode,'timed_out':timed_out,'elapsed_seconds':round(time.monotonic()-started,2),'stdout':safe(stdout),'stderr':safe(stderr),'requests':requests})
            return child.returncode==0 and not timed_out
        word='cobalt-juniper'
        if not run('first process','Remember the secret word '+word+'. Reply briefly; do not use tools.'): continue
        save=(directory/'save.json').read_bytes()
        run('second process','What secret word did I ask you to remember? Reply only with that word, without tools.')
        if vendor=='anthropic':
            for variant in ['snapshot-only','log-only']:
                data=json.loads(save)
                if variant=='snapshot-only': data['log']=[]
                else: data['context']=None
                (directory/'save.json').write_text(json.dumps(data))
                run(variant,'What secret word did I ask you to remember? Reply only with that word, without tools.')
            bad=save[:len(save)//2]
            (directory/'save.json').write_bytes(bad)
            run('corrupt refusal','What was the word?')
            runs[-1]['file_unchanged']=(directory/'save.json').read_bytes()==bad
        data={'provider':vendor,'model':model,'discovery_source':'../ch10/'+vendor+'-models.json (same-day discovered model reused)','runs':runs}
        (out/(vendor+'-live.json')).write_text(json.dumps(data,indent=2)+'\n')
        print(vendor,[(r['label'],r['exit'],word in r['stdout'],r.get('file_unchanged')) for r in runs],flush=True)
