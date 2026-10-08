"""Read only needed credentials into memory; save model identities, never headers."""
import datetime,json,pathlib,urllib.request,urllib.error
settings=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())
routes=[('anthropic','https://api.anthropic.com/v1/models','directClaudeAPIKey','x-api-key'),('openai','https://api.openai.com/v1/models','directOpenAIAPIKey','Authorization'),('gemini','https://generativelanguage.googleapis.com/v1beta/models','directGeminiAPIKey','x-goog-api-key')]
out={'utc':datetime.datetime.now(datetime.timezone.utc).isoformat(),'providers':{}}
for vendor,url,key,header in routes:
 headers={header:('Bearer ' if vendor=='openai' else '')+settings[key]}
 if vendor=='anthropic':headers['anthropic-version']='2023-06-01'
 try:
  with urllib.request.urlopen(urllib.request.Request(url,headers=headers),timeout=30) as response: data=json.load(response)
  models=data.get('data',data.get('models',[]))
  out['providers'][vendor]=[{'id':m.get('id',m.get('name')),'methods':m.get('supportedGenerationMethods')} for m in models]
  print(vendor,json.dumps(out['providers'][vendor]))
 except urllib.error.HTTPError as e:out['providers'][vendor]={'status':e.code};print(vendor,'HTTP',e.code)
pathlib.Path(__file__).with_name('discovery.json').write_text(json.dumps(out,indent=2)+'\n')
