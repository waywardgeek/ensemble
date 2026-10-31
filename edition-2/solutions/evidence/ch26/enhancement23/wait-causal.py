import pathlib,tempfile,subprocess,json,shutil,os,time,hashlib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch26/enhancement23';scratch=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch23-wait-'))
runs=[]
try:
 tree=scratch/'source';shutil.copytree(root/'solutions/edition-2/ch23',tree)
 test=scratch/'check';subprocess.run(['go','test','-c','-race','-o',str(test),'./internal/gradeedition2'],check=True)
 for stage in ['defect','restored']:
  path=tree/'internal/fleet/management.go'
  if stage=='defect':
   original=path.read_text();subprocess.run(['git','apply','--unsafe-paths','-p1',str(out/'reviewer-wait-immediate.patch')],cwd=tree,check=True)
  else:path.write_text(original)
  binary=scratch/'gui';subprocess.run(['go','build','-o',str(binary),'./cmd'],cwd=tree/'gui',check=True)
  start=time.monotonic();r=subprocess.run([str(test),'-test.run','^(TestManagement23|TestManagementReplayText23)$','-test.v'],env=dict(os.environ,EDITION2_MANAGEMENT_BIN=str(binary)),capture_output=True,text=True,timeout=25);duration=time.monotonic()-start
  (out/('wait-causal-'+stage+'.txt')).write_text(r.stdout+r.stderr)
  checks=next(json.loads(s) for s in r.stdout.splitlines() if s.startswith('[{"id"'));failed={c['id']:c.get('details',[]) for c in checks if not c['passed']}
  runs.append(dict(stage=stage,exit=r.returncode,wall_seconds=duration,failed=failed))
  if stage=='defect':assert 'no progress' in str(failed.get('wait-change-wakes')),failed
  else:assert r.returncode==0,failed
 (out/'wait-causal-check.json').write_text(json.dumps(dict(runs=runs,patch='reviewer-wait-immediate.patch',source={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (root/'internal/gradeedition2').glob('ch23*.go')},absence_interval_ms=200,scope='Actual shell readiness before wait-producing response; bounded no-progress observation, then real release/completion. Not an exact handler-entry timing proof.'),indent=2)+'\n')
finally:shutil.rmtree(scratch)
