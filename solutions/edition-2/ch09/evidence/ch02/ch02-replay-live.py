import json,pathlib,subprocess,os
root=pathlib.Path('/tmp/ensemble-ed2-ch02-live')
results=[]
for folder in sorted(root.iterdir()):
 receipt=folder/'receipt.json'
 if not receipt.is_file():continue
 record=json.loads(receipt.read_text())
 if not (folder/'session.log').exists() or record['exit_code']!=0:continue
 rows=(folder/'session.log').read_text().splitlines()
 env=os.environ.copy()
 for key in ['LLM_API_KEY','ANTHROPIC_API_KEY','OPENAI_API_KEY','GEMINI_API_KEY']:env.pop(key,None)
 env.update(LLM_VENDOR=record['vendor'],LLM_MODEL=record['model_requested'])
 appearances=[]
 for index,row in enumerate(rows):
  event=json.loads(row)
  if event.get('type')!='request_sent':continue
  prefix=folder/('prefix-before-'+str(event['seq'])+'.log');prefix.write_text('\n'.join(rows[:index])+'\n')
  run=subprocess.run(['/tmp/ensemble-ed2-ch02','render',str(prefix)],env=env,capture_output=True,text=True,timeout=20)
  assert run.returncode==0,run.stderr
  appearances.append('LILAC-614' in run.stdout)
  (folder/('request-before-'+str(event['seq'])+'.json')).write_text(run.stdout)
 assert appearances==[False,True,False],appearances
 result={'vendor':record['vendor'],'model':record['model_requested'],'directory':str(folder),'ephemeral_in_replayed_requests':appearances,'offline_only':True}
 (folder/'ephemeral-replay.json').write_text(json.dumps(result,indent=2)+'\n');results.append(result)
print(json.dumps(results,indent=2))
