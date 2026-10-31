"""Three built binaries and the actual Ensemble actor; no credential values are read here."""
import hashlib, json, os, pathlib, selectors, signal, socket, subprocess, tempfile, time
HERE=pathlib.Path(__file__).resolve().parent
BIN=pathlib.Path('/tmp/ensemble-gateway26-binaries')
started=time.monotonic()
records={'source':'ab6d0516','main_tree':'0a9d7b1104eab3cd973d7ad77c9bce04e98d09b6','selected_model':'gpt-6-astra','reasoning_effort':'high','billing_mode':'codex','processes':{},'success':False}
children=[]
def launch(name,args,env=None):
    p=subprocess.Popen(args,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
    children.append((name,p));records['processes'][name]={'pid':p.pid,'binary_sha256':hashlib.sha256(pathlib.Path(args[0]).read_bytes()).hexdigest()}
    return p
def ready(p,word,seconds=15):
    sel=selectors.DefaultSelector();sel.register(p.stdout,selectors.EVENT_READ)
    until=time.monotonic()+seconds
    while time.monotonic()<until:
        if p.poll() is not None:raise RuntimeError(f'child exited before readiness: {p.returncode}')
        if sel.select(.2):
            line=p.stdout.readline()
            if word in line:return line.strip()
    raise RuntimeError('child readiness timeout')
with tempfile.TemporaryDirectory(prefix='gateway26-live-') as scratch:
    root=pathlib.Path(scratch)
    try:
        with socket.socket() as sock:
            sock.bind(('127.0.0.1',0));address=f'127.0.0.1:{sock.getsockname()[1]}'
        url='ws://'+address+'/v1/ws';records['address']=address
        public={}
        for name,binary in [('ensemble','subject'),('interviewer','interviewer')]:
            public[name]=subprocess.check_output([str(BIN/binary),'-key',str(root/(name+'.key')),'-print-key'],text=True).strip()
        cfg={'gateway':{'listen':address,'data_dir':str(root/'gateway')},'groups':{'agents':{'trust':'trusted','capabilities':['send-message','declare-contactable']}},'agents':{name:{'public_key':key,'groups':['agents']} for name,key in public.items()}}
        config=root/'gateway.json';config.write_text(json.dumps(cfg))
        gwpublic=subprocess.check_output([str(BIN/'gateway'),'-config',str(config),'-print-key'],text=True).strip()
        gw=launch('gateway',[str(BIN/'gateway'),'-config',str(config)])
        ready(gw,'x25519:')
        # Prove this newly launched child owns the chosen port; an unrelated existing
        # user's Ensemble is never contacted or stopped.
        until=time.monotonic()+5
        while True:
            if gw.poll() is not None:raise RuntimeError('gateway exited during listen')
            try:
                with socket.create_connection(('127.0.0.1',int(address.split(':')[1])),timeout=.2):break
            except OSError:
                if time.monotonic()>until:raise
                time.sleep(.05)
        env=dict(os.environ)
        for key in ['LLM_BASE_URL','OPENAI_BASE_URL','LLM_RESOLVED_MODEL','OPENAI_API_KEY','LLM_API_KEY']:
            env.pop(key,None)
        env.update(LLM_VENDOR='openai',LLM_MODEL='gpt-6-astra',OPENAI_REASONING_EFFORT='high',OPENAI_BILLING_MODE='codex',EN_RECALL_DISABLED='1')
        subject=launch('subject',[str(BIN/'subject'),'-gateway',url,'-key',str(root/'ensemble.key'),'-gateway-key',gwpublic,'-data',str(root/'subject')],env)
        records['subject_ready']=ready(subject,'Gateway Agent ensemble ready')
        interview=launch('interviewer',[str(BIN/'interviewer'),'-gateway',url,'-key',str(root/'interviewer.key'),'-gateway-key',gwpublic,'-report',str(root/'report.md'),'-timeout','90s'])
        while interview.poll() is None:
            if time.monotonic()-started>180:raise RuntimeError('bounded whole interview timeout')
            time.sleep(.5)
        if interview.returncode!=0:raise RuntimeError('interviewer failed')
        (HERE/'live-interview.md').write_bytes((root/'report.md').read_bytes())
        records['success']=True
    except Exception as error:
        records['failure']=str(error)
    finally:
        for name,p in reversed(children):
            if p.poll() is None:p.send_signal(signal.SIGTERM)
            try:out,err=p.communicate(timeout=15)
            except subprocess.TimeoutExpired:
                p.kill();out,err=p.communicate();records['processes'][name]['forced_kill']=True
            records['processes'][name].update(exit=p.returncode,stdout=out,stderr=err)
        save=root/'subject'/'save.json'
        if save.exists():
            saved=json.loads(save.read_text());events=saved.get('log',[])
            records['saved_model']=saved['config']['model']
            records['event_counts']={kind:sum(e.get('type')==kind for e in events) for kind in ['request_sent','response_ended','message_received']}
            # The save is credential-free; keep only the bounded four-turn facts.
            (HERE/'live-events.json').write_text(json.dumps({'config':saved['config'],'log':events},indent=2)+'\n')
        records['elapsed_seconds']=round(time.monotonic()-started,3)
        (HERE/'live-result.json').write_text(json.dumps(records,indent=2)+'\n')
        print(json.dumps(records,indent=2))
