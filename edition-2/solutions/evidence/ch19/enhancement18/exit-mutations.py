import pathlib,subprocess,tempfile,shutil,json,time,difflib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch19/enhancement18';src=root/'solutions/edition-2/ch18'
def case(name,file,old,new,check,positive=False):return dict(name=name,file=file,old=old,new=new,check=check,positive=positive)
cases=[
case('panic-on-invalid-nonce','internal/auth/identity.go','var claims identityClaims','var claims identityClaims\n json.Unmarshal(raw, &claims)\n if nonce != "" && claims.Nonce != nonce { panic("invalid nonce") }','oauth-flow'),
case('panic-after-refusal','gui/gui.go','s.broadcast(map[string]any{"type": "turn_ended", "text": o.Text, "error": o.Error})','s.broadcast(map[string]any{"type": "turn_ended", "text": o.Text, "error": o.Error})\n if o.Error != "" { panic("failed turn") }','billing-mode'),
]
subprocess.run(['go','build','-o','/tmp/edition2-grade18','./cmd/grade-edition2'],check=True)
with tempfile.TemporaryDirectory(prefix='ch18-mutants-') as tmp:
 work=pathlib.Path(tmp)/'source';shutil.copytree(src,work);rows=[]
 for c in cases:
  f=work/c['file'];old=f.read_text();assert c['old'] in old,c['name'];new=old.replace(c['old'],c['new']);f.write_text(new)
  patch=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+c['file'],tofile='b/'+c['file']));(out/(c['name']+'.patch')).write_text(patch)
  t=time.monotonic();r=subprocess.run(['/tmp/edition2-grade18','-ch','18','-json',str(work)],capture_output=True,text=True,timeout=65);secs=round(time.monotonic()-t,3)
  (out/(c['name']+'.json')).write_text(r.stdout);(out/(c['name']+'.stderr')).write_text(r.stderr)
  try:
   report=json.loads(r.stdout);failed={x['id']:x.get('details',[]) for x in report['checks'] if not x['passed']};score=report['score'];caught=c['check'] in failed
  except Exception:failed={};score=None;caught=False
  row={k:c[k] for k in ['name','file','check','positive']};row.update(exit=r.returncode,seconds=secs,score=score,failed=failed,intended=(r.returncode==0 if c['positive'] else caught));rows.append(row)
  (out/'exit-mutations.json').write_text(json.dumps(rows,indent=2)+'\n');print(c['name'],score,row['intended'],flush=True);f.write_text(old)
 r=subprocess.run(['/tmp/edition2-grade18','-ch','18','-json',str(work)],capture_output=True,text=True);(out/'exit-restored.json').write_text(r.stdout);(out/'exit-restored.stderr').write_text(r.stderr);print('restored',r.returncode,flush=True)
