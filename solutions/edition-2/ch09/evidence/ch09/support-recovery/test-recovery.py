"""Local transport/cap controls; never accesses credentials or real origins."""
import importlib.util,json,pathlib,tempfile,threading,urllib.error,urllib.request
from unittest.mock import patch
from evidence import caps
spec=importlib.util.spec_from_file_location('launcher',pathlib.Path(__file__).with_name('terminal-run.py')); launcher=importlib.util.module_from_spec(spec);spec.loader.exec_module(launcher)
results=[]
assert caps('openai')==dict(N=16,F=5,G=10,P=5)
assert caps('anthropic')==caps('gemini')==dict(N=16,F=6,G=10,P=4)
assert sum(caps('openai').values())==36
for vendor,scenario,used,allowed in [('openai','P',dict(N=11,F=5,G=8,P=1),4),('openai','P',dict(N=16,F=5,G=10,P=5),0),('openai','F',dict(N=0,F=5,G=0,P=0),0),('gemini','P',dict(N=0,F=0,G=0,P=4),0)]:
 with tempfile.TemporaryDirectory() as raw:
  d=pathlib.Path(raw);(d/'requests').mkdir();(d/'responses').mkdir();budget=d/'budget.json';budget.write_text(json.dumps(used))
  relay=launcher.Relay(vendor,'fixture-secret',d,scenario,budget,'http://127.0.0.1:1');worker=threading.Thread(target=relay.serve_forever);worker.start()
  real=urllib.request.urlopen;forwarded=[]
  def open_url(request,*args,**kwargs):
   if request.full_url.startswith('http://127.0.0.1:1/'):
    forwarded.append(request);raise urllib.error.URLError(TimeoutError('fixture-secret must not appear'))
   return real(request,*args,**kwargs)
  try:
   with patch.object(launcher.urllib.request,'urlopen',open_url):
    for i in range(allowed+1):
     request=urllib.request.Request(f'http://127.0.0.1:{relay.server_port}/fixture',data=b'{}')
     try:real(request,timeout=3);raise AssertionError('unexpected success')
     except urllib.error.HTTPError as e:assert e.code==(502 if i<allowed else 429)
   assert len(forwarded)==allowed and relay.number==allowed
   final=json.loads(budget.read_text());assert final[scenario]==used[scenario]+allowed
   diagnostics=d/'transport-errors.jsonl'
   if allowed:
    body=diagnostics.read_text();assert 'fixture-secret' not in body and 'http://' not in body
    rows=[json.loads(x) for x in body.splitlines()];assert len(rows)==allowed
    assert all(x['stage']=='open_upstream' and x['exception_type']=='URLError' and x['reason_type']=='TimeoutError' for x in rows)
   else:assert not diagnostics.exists()
   results.append(dict(vendor=vendor,scenario=scenario,forwarded=allowed,next_refused=True,final=final))
  finally:relay.shutdown();relay.server_close();worker.join()
print(json.dumps({'status':'pass','credential_reads':0,'real_http_requests':0,'cases':results},indent=2))
