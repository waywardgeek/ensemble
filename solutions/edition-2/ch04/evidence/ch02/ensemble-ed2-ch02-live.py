import datetime,json,os,pathlib,subprocess,sys,urllib.request,urllib.error,tempfile
out=pathlib.Path('/tmp/ensemble-ed2-ch02-live');out.mkdir(exist_ok=True)
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
keys={v:settings[k] for v,k in [('anthropic','directClaudeAPIKey'),('openai','directOpenAIAPIKey'),('gemini','directGeminiAPIKey')]}
mode=sys.argv[1]
if mode=='discover':
 records={}
 for vendor,url in [('anthropic','https://api.anthropic.com/v1/models?limit=100'),('openai','https://api.openai.com/v1/models'),('gemini','https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000')]:
  headers={'anthropic':{'x-api-key':keys[vendor],'anthropic-version':'2023-06-01'},'openai':{'Authorization':'Bearer '+keys[vendor]},'gemini':{'x-goog-api-key':keys[vendor]}}[vendor]
  try:
   with urllib.request.urlopen(urllib.request.Request(url,headers=headers),timeout=30) as response: data=json.load(response)
  except urllib.error.HTTPError as e:print(vendor,'discovery HTTP',e.code);continue
  except Exception as e:print(vendor,'discovery failed',type(e).__name__);continue
  models=data.get('data',data.get('models',[]))
  records[vendor]={'date':datetime.datetime.now(datetime.timezone.utc).isoformat(),'models':models,'more':data.get('has_more',data.get('nextPageToken',False))}
  print(vendor,json.dumps([m.get('id',m.get('name')) for m in models]))
 (out/'discovery.json').write_text(json.dumps(records,indent=2)+'\n')
else:
 vendor,model=sys.argv[2:4]
 discovery=json.loads((out/'discovery.json').read_text());ids=[m.get('id',m.get('name')) for m in discovery[vendor]['models']]
 assert model in ids, 'Model not discovered'
 env=os.environ.copy();env.update(LLM_VENDOR=vendor,LLM_MODEL=model,LLM_API_KEY=keys[vendor])
 env.pop('LLM_BASE_URL',None);env.pop('LLM_RESOLVED_MODEL',None)
 folder=pathlib.Path(tempfile.mkdtemp(prefix=vendor+'-',dir=out));env['CH02_LOG']=str(folder/'session.log')
 if mode=='cli':
  inputs=[{'user':'Invent a short two-word code name. Reply with only the name.'},{'ephemeral':'One-request diagnostic marker: LILAC-614. It is context for this request only; do not repeat it in your answer.'},{'user':'What exact code name did you invent? Reply with only that same name.'},{'user':'Reverse the code name from your first answer character by character. Reply only with the reversed text.'}]
  command=['/tmp/ensemble-ed2-ch02'];stdin='\n'.join(json.dumps(i) for i in inputs)+'\n'
 elif mode=='consumer':
  second=sys.argv[4];assert second in ids and second!=model,'Second model must be independently discovered'
  inputs=[];command=['/tmp/ensemble-ed2-ch02-consumer'];stdin='';env['DEMO_DIR']=str(folder);env['DEMO_SECOND_MODEL']=second
 else:raise SystemExit('unknown mode')
 run=subprocess.run(command,input=stdin,text=True,capture_output=True,env=env,cwd=folder,timeout=300)
 record={'date':datetime.datetime.now(datetime.timezone.utc).isoformat(),'vendor':vendor,'model_requested':model,'command':command,'inputs':inputs,'stdout':run.stdout,'stderr':run.stderr,'exit_code':run.returncode,'directory':str(folder)}
 if mode=='cli' and run.returncode==0:
  offline=env.copy();offline.pop('LLM_API_KEY',None)
  for k in ['ANTHROPIC_API_KEY','OPENAI_API_KEY','GEMINI_API_KEY']:offline.pop(k,None)
  for name,args in [('dump',['dump']),('render1',['render',env['CH02_LOG']]),('render2',['render',env['CH02_LOG']])]:
   p=subprocess.run(command+args,text=True,capture_output=True,env=offline,cwd=folder,timeout=30);record[name]={'exit_code':p.returncode,'stdout':p.stdout,'stderr':p.stderr}
   (folder/(name+'.json')).write_text(p.stdout)
  record['render_equal']=record['render1']['stdout']==record['render2']['stdout']
  events=[json.loads(line) for line in pathlib.Path(env['CH02_LOG']).read_text().splitlines()][1:]
  record['responses']=[e['response'] for e in events if e['type']=='response_ended']
  record['consumed_ephemera']=[e['request']['ephemera'] for e in events if e['type']=='request_sent']
 encoded=json.dumps(record,indent=2)
 for key in keys.values():assert key not in encoded, 'credential appeared in evidence'
 (folder/'receipt.json').write_text(encoded+'\n');print(json.dumps({'receipt':str(folder/'receipt.json'),'exit_code':run.returncode,'stdout':run.stdout,'stderr':run.stderr,'render_equal':record.get('render_equal'),'consumed_ephemera':record.get('consumed_ephemera')}))
 sys.exit(run.returncode)
