"""Validate the exact web/status recipe with real clients and external peers."""
import hashlib,http.server,json,os,pathlib,re,shutil,subprocess,tempfile,threading,time
root=pathlib.Path(os.environ['CH20_ROOT']);out=root/'solutions/edition-2/evidence/ch21/enhancement20'
chapter=root/'book/edition-2/chapter-20.md';sha=hashlib.sha256(chapter.read_bytes()).hexdigest()
assert sha=='b975d7aaa77d767ca7e6c7f3976367048cab89b15a6491d307b9182ee89cb1f9',sha
blocks=re.findall(r'```([^\n]*)\n(.*?)```',chapter.read_text(),re.S)
assert len(blocks)==17
requests=[];judges=[];mcp=[];receipts=[];errors=[]
nonce='LOCAL_WEB_BODY_'+os.urandom(10).hex()
page='Go course fixture. go1.27.2.linux-amd64.tar.gz\n'+nonce

def call(id,name,args):return dict(type='function_call',id='item-'+id,call_id=id,name=name,arguments=json.dumps(args),status='completed')
def text(s):return dict(type='message',id='msg-'+s,role='assistant',status='completed',content=[dict(type='output_text',text=s)])
script=[call('load-web','load_skill',{'name':'web-search'}),call('search-web','firecrawl_search',{'query':'latest stable Go release site:go.dev'}),call('scrape-web','firecrawl_scrape',{'url':'https://go.dev/dl/'}),text('SAMPLE_WEB_DONE'),call('status-first','agent_status',{}),call('unload-web','unload_skill',{'name':'web-search'}),text('SAMPLE_STATUS_ONE_DONE'),call('status-second','agent_status',{}),text('SAMPLE_STATUS_TWO_DONE')]
class Boundary(http.server.BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def do_POST(self):
  try:
   q=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
   if self.path=='/mcp':
    mcp.append(q)
    assert self.headers.get('Authorization')=='Bearer synthetic-firecrawl-key'
    if q['method']!='initialize':assert self.headers.get('Mcp-Session-Id')=='sample-session'
    if 'id' not in q:self.send_response(202);self.end_headers();return
    method=q['method']
    if method=='initialize':result=dict(protocolVersion='2024-11-05',capabilities={'tools':{}},serverInfo={'name':'sample','version':'1'})
    elif method=='tools/list':result={'tools':[dict(name=n,description=n,inputSchema=dict(type='object',properties={k:{'type':'string'}},required=[k])) for n,k in [('firecrawl_search','query'),('firecrawl_scrape','url')]]}
    else:
     assert method=='tools/call'
     params=q['params']
     if params['name']=='firecrawl_search':assert params['arguments']['query']=='latest stable Go release site:go.dev';value='Official result https://go.dev/dl/'
     else:assert params['name']=='firecrawl_scrape' and params['arguments']['url']=='https://go.dev/dl/';value=page
     result=dict(content=[dict(type='text',text=value)],isError=False)
    self.send_response(200);self.send_header('Content-Type','application/json')
    if method=='initialize':self.send_header('Mcp-Session-Id','sample-session')
    self.end_headers();self.wfile.write(json.dumps(dict(jsonrpc='2.0',id=q['id'],result=result)).encode());return
   assert self.path=='/v1/responses'
   assert self.headers.get('Authorization','').startswith('Bearer ')
   if not q.get('tools'):
    judges.append(q);item=text('[]');usage=dict(input_tokens=0,output_tokens=0,input_tokens_details=dict(cached_tokens=0))
   else:
    requests.append(q);i=len(requests)-1;assert i<len(script)
    assert q['model']==('gpt-6-astra' if i<7 else 'gpt-6.1-sol')
    assert q['reasoning']==dict(effort='high',summary='detailed')
    item=script[i];usage=dict(input_tokens=100,output_tokens=11,input_tokens_details=dict(cached_tokens=20))
   self.send_response(200);self.send_header('Content-Type','text/event-stream');self.end_headers()
   response=dict(id='course-'+str(len(requests)),status='completed',model=q['model'],output=[item],usage=usage)
   self.wfile.write(('event: response.completed\ndata: '+json.dumps(dict(type='response.completed',response=response))+'\n\n').encode());self.wfile.flush()
  except Exception as e:errors.append(repr(e));raise
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Boundary)
threading.Thread(target=server.serve_forever,daemon=True).start();work=None

def run(command,cwd,env):
 r=subprocess.run(['bash','-c',command],cwd=cwd,env=env,text=True,capture_output=True,timeout=30)
 receipts.append(dict(command=command,exit=r.returncode,stdout=r.stdout,stderr=r.stderr));assert r.returncode==0,receipts[-1];return r.stdout
try:
 env={k:v for k,v in os.environ.items() if not k.endswith('API_KEY') and not k.startswith(('LLM_','OPENAI_','EN_','ENSEMBLE_','CH02_'))}
 work=pathlib.Path(run(blocks[10][1]+'\npwd',root/'solutions/edition-2/ch20',env).strip())
 (work/'skills/ensemble/SKILL.md').write_text(blocks[11][1])
 web=blocks[3][1];endpoint=re.search(r'^    url: (.+)$',web,re.M).group(1)
 local=f'http://127.0.0.1:{server.server_port}/mcp'
 (work/'skills/web-search/SKILL.md').write_text(web.replace(endpoint,local))
 env.update(EN_HOME=str(work/'profile'),OPENAI_OIDC_ISSUER=os.environ['CH20_ISSUER'],LLM_BASE_URL=f'http://127.0.0.1:{server.server_port}',CH02_LOG=str(work/'events.jsonl'),FIRECRAWL_API_KEY='synthetic-firecrawl-key')
 browser=work/'browser.py';browser.write_text('import sys,urllib.request\ntry:\n urllib.request.urlopen(sys.argv[1],timeout=5).read()\nexcept Exception:\n pass\n')
 env['EN_BROWSER_CMD']='python3 '+str(browser)
 run('/tmp/ensemble-ch20 auth login',work,env)
 with (work/'stdout.txt').open('w') as stdout,(work/'stderr.txt').open('w') as stderr:
  child=subprocess.Popen(['bash','-c',blocks[12][1]],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
  def send(block,marker,channel="stdout.txt"):
   assert child.poll() is None,'child exited before input'
   child.stdin.write(block);child.stdin.flush();start=time.monotonic()
   while child.poll() is None and time.monotonic()-start<20:
    if marker in (work/channel).read_text():return
    time.sleep(.01)
   raise AssertionError(('completion absent',marker,(work/'stderr.txt').read_text()))
  send(blocks[13][1],'SAMPLE_WEB_DONE')
  send(blocks[14][1],'SAMPLE_STATUS_ONE_DONE')
  send(blocks[15][1],'Model: gpt-6.1-sol','stderr.txt')
  send(blocks[16][1],'SAMPLE_STATUS_TWO_DONE')
  child.stdin.close();code=child.wait(timeout=5)
 receipts.append(dict(command=blocks[12][1],exit=code,stdout=(work/'stdout.txt').read_text(),stderr=(work/'stderr.txt').read_text()))
 assert code==0 and len(requests)==9 and not errors,receipts[-1]
 results={item['call_id']:item['output'] for q in requests for item in q['input'] if item.get('type')=='function_call_output'}
 assert results['scrape-web']==page and 'https://go.dev/dl/' in results['search-web']
 assert 'model: gpt-6-astra' in results['status-first'] and 'session total: 400 in, 55 out, 100 cache read, 0 cache write' in results['status-first'],results['status-first']
 assert 'model: gpt-6.1-sol' in results['status-second'] and 'session total: 640 in, 88 out, 160 cache read, 0 cache write' in results['status-second'],results['status-second']
 assert 'unknown' in results['status-first'] and 'unknown' in results['status-second']
 for q in requests[6:]:assert not any(t.get('name') in ('firecrawl_search','firecrawl_scrape') for t in q['tools'])
 assert (work/'save.json').is_file()
 (out/'samples-requests.json').write_text(json.dumps(requests,indent=2)+'\n')
 (out/'samples-judge-requests.json').write_text(json.dumps(judges,indent=2)+'\n')
 (out/'samples-mcp.json').write_text(json.dumps(mcp,indent=2)+'\n')
 (out/'samples-run.json').write_text(json.dumps(receipts,indent=2)+'\n')
 (out/'samples.json').write_text(json.dumps(dict(result='PASS',manuscript_sha256=sha,fences=[dict(index=i,language=k,sha256=hashlib.sha256(v.encode()).hexdigest()) for i,(k,v) in enumerate(blocks)],main_requests=9,judge_requests=len(judges),cli_exit=code,save=True,substitutions=['Only web skill endpoint replaced with local HTTP peer','Existing external fake signed issuer and isolated auth profile','Local Responses base URL and synthetic Firecrawl key','CH02_LOG captures actual events'],limits='Scripted provider decisions validate CLI/tool/model-switch mechanics, not fresh autonomous research, real service catalog, paid billing/cache grants or real account eligibility. Immediate scrape exercises the prompt conditional without a running-job branch; historical real wait proof retained separately.'),indent=2)+'\n')
 print('PASS exact sample',sha,'9 main requests, two status results, save and exit0')
finally:
 if 'child' in globals() and child.poll() is None:
  child.stdin.close();child.wait(timeout=5)
 server.shutdown();server.server_close()
 if work:shutil.rmtree(work)
