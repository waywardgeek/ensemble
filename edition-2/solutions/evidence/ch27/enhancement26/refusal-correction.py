import pathlib,subprocess,os,tempfile,shutil,time,json,hashlib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch27/enhancement26';rows=[]
paths=['cmd/grade-edition2/main.go','internal/gradeedition2/ch26.go','internal/gradeedition2/ch26_probe.go.txt','internal/gradeedition2/ch26_test.go']
def run(name,source,field='QueueDepth'):
 t=time.monotonic();r=subprocess.run(['go','test','./internal/gradeedition2','-run','^TestGateway26$','-count=1','-v'],env=dict(os.environ,GATEWAY_SUBMISSION=str(source),GATEWAY_WELCOME=field),capture_output=True,text=True,timeout=45);raw=r.stdout+r.stderr;(out/(name+'.txt')).write_text(raw);rows.append(dict(name=name,exit=r.returncode,seconds=round(time.monotonic()-t,3)));print(rows[-1],flush=True);return r,raw
with tempfile.TemporaryDirectory(prefix='gateway26-refusal-') as temp:
 copy=pathlib.Path(temp)/'submission';shutil.copytree(root/'solutions/edition-2/ch26',copy)
 assert run('refusal-baseline',copy)[0].returncode==0
 target=copy/'internal/gateway/admin.go';original=target.read_bytes()
 subprocess.run(['patch','-p1','-i',str(out/'reviewer-register-before-refusal.patch')],cwd=copy,check=True)
 r,raw=run('refusal-defect',copy);assert r.returncode!=0 and 'refused wider registration nevertheless admitted its identity' in raw
 target.write_bytes(original);assert run('refusal-restored',copy)[0].returncode==0
 assert run('refusal-legacy',root/'solutions/ch27','QueuedCount')[0].returncode==0
(out/'refusal-correction.json').write_text(json.dumps(dict(sources={p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in paths},patch='reviewer-register-before-refusal.patch',cases=rows,cleanup='disposable copy removed; frozen source unchanged'),indent=2)+'\n')
