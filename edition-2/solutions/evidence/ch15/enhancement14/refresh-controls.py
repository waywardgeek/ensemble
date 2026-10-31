import subprocess,pathlib,tempfile,shutil,json,difflib,time
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch15/enhancement14';source=root/'solutions/edition-2/ch14'
cases=[
 ('duplicate-wire-ephemera','internal/engine/engine.go','out = append(out, common.Entry{Actor: common.System, Parts: c.Ephemera})','out = append(out, common.Entry{Actor: common.System, Parts: c.Ephemera}, common.Entry{Actor: common.System, Parts: c.Ephemera})','frozen-prefix'),
 ('stale-snapshot-marker','internal/mcp/client.go','return joined, nil','if name == "gui_snapshot" { joined = strings.ReplaceAll(joined,"SNAPSHOT_ONLY_2","SNAPSHOT_ONLY_1") }; return joined, nil','frozen-prefix'),
]
results=json.loads(out.joinpath('mutations.json').read_text())
with tempfile.TemporaryDirectory(prefix='ch14-mutants-') as tmp:
 tree=pathlib.Path(tmp)/'source';shutil.copytree(source,tree)
 for label,file,old,new,intended in cases:
  path=tree/file;original=path.read_text();assert original.count(old)==1,(label,original.count(old));mutated=original.replace(old,new)
  path.write_text(mutated);out.joinpath(label+'.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=file,tofile=file)))
  command=['/tmp/grade-edition2-context','-ch','14','-json','-gui',str(tree/'gui'),str(tree)]
  start=time.monotonic();r=subprocess.run(command,capture_output=True,text=True,timeout=115)
  out.joinpath(label+'.json').write_text(r.stdout or '{}\n');out.joinpath(label+'.stderr').write_text(r.stderr)
  try:report=json.loads(r.stdout)
  except:report={}
  failures=[c for c in report.get('checks',[]) if not c['passed']]
  result={'case':label,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'intended_check':intended,'intended_failure':any(c['id']==intended for c in failures) if intended else None,'score':report.get('score'),'failures':failures}
  results.append(result);out.joinpath('mutations.json').write_text(json.dumps(results,indent=2)+'\n');print(label,r.returncode,report.get('score'),result['intended_failure'],flush=True)
  path.write_text(original)
 # Restore exact runtime source, then rerun the affected suite.
 r=subprocess.run(['/tmp/grade-edition2-context','-ch','14','-json','-gui',str(tree/'gui'),str(tree)],capture_output=True,text=True,timeout=115)
 out.joinpath('refresh-restored-full.json').write_text(r.stdout);out.joinpath('refresh-restored-full.stderr').write_text(r.stderr);assert r.returncode==0
