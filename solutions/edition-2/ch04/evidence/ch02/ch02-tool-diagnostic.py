import json,pathlib,urllib.request,urllib.error
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
keys={v:settings[k] for v,k in [('openai','directOpenAIAPIKey'),('gemini','directGeminiAPIKey')]}
schema={'type':'object','properties':{'path':{'type':'string'}},'required':['path'],'additionalProperties':False}
prompt='Call inspect exactly once with path "demo-config". Do not invent its result. Please call the tool now.'
description='Read the controlled demonstration record named by path; the example supplies the result and executes nothing.'
system="Answer helpfully and concisely. Retain the conversation's details."
for vendor in ['openai','gemini']:
 if vendor=='openai':
  body={'model':'gpt-6-luna','max_completion_tokens':2048,'messages':[{'role':'system','content':system},{'role':'user','content':prompt}],'tools':[{'type':'function','function':{'name':'inspect','description':description,'parameters':schema}}]};url='https://api.openai.com/v1/chat/completions';headers={'Authorization':'Bearer '+keys[vendor]}
 else:
  body={'systemInstruction':{'parts':[{'text':system}]},'contents':[{'role':'user','parts':[{'text':prompt}]}],'generationConfig':{'maxOutputTokens':2048},'tools':[{'functionDeclarations':[{'name':'inspect','description':description,'parameters':schema}]}]};url='https://generativelanguage.googleapis.com/v1beta/models/gemini-3.8-flash:generateContent';headers={'x-goog-api-key':keys[vendor]}
 headers['Content-Type']='application/json'
 try:
  with urllib.request.urlopen(urllib.request.Request(url,json.dumps(body).encode(),headers),timeout=60) as response: record={'vendor':vendor,'status':response.status,'unexpected_success':True}
 except urllib.error.HTTPError as e:
  data=json.load(e);err=data.get('error',{});record={'vendor':vendor,'status':e.code,'message':err.get('message'),'type':err.get('type'),'param':err.get('param')}
 for secret in keys.values():assert secret not in json.dumps(record)
 print(json.dumps(record))
 pathlib.Path('/tmp/ensemble-ed2-ch02-live/'+vendor+'-tool-diagnostic.json').write_text(json.dumps(record,indent=2)+'\n')
