"""Restored-defect checks for two pre-submission production corrections."""
import pathlib,shutil,tempfile,subprocess,json
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
cases=[
 ('relative-skill-restored','internal/recall/recall.go','if source.name == "skills" {\n\t\t\t\tbase = source.path\n\t\t\t}','if source.name == "never" {\n\t\t\t\tbase = source.path\n\t\t\t}','^TestRecallSources$','Rel:'),
 ('judge-route-restored','internal/engine/engine.go','LookupModel(cfg.Model)','LookupModel(e.parent.Config().Model)','^TestRecallJudgeRoute$','judge inherited parent model header'),
 ('owner-corpus-r2','internal/recall/recall.go','root := r.parent.Config().DataDir','root := "."','^TestRecallSources$','missing source docs/record.md'),
]
temp=pathlib.Path(tempfile.mkdtemp(prefix='ch17-regressions-'))
results=[]
try:
 shutil.copytree(ROOT/'solutions/edition-2/main',temp/'source',ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 tree=temp/'source'
 baseline=subprocess.run(['go','test','./ensemble','-run','^TestRecall(Sources|JudgeRoute)$','-count=1'],cwd=tree,capture_output=True,text=True)
 (OUT/'mutation-regressions-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,path,old,new,test,expected in cases:
  file=tree/path;original=file.read_text()
  if original.count(old)!=1:raise RuntimeError(name+' occurrence mismatch')
  file.write_text(original.replace(old,new))
  run=subprocess.run(['go','test','./ensemble','-run',test,'-count=1','-timeout','10s'],cwd=tree,capture_output=True,text=True)
  output=run.stdout+run.stderr;(OUT/('mutation-'+name+'.txt')).write_text(output)
  killed=run.returncode!=0 and expected in output and '[build failed]' not in output
  results.append({'claim':name,'file':path,'test':test,'exit':run.returncode,'intended_failure':killed})
  print(name,killed,flush=True);file.write_text(original)
 (OUT/'mutations-regressions.json').write_text(json.dumps(results,indent=2))
finally:shutil.rmtree(temp)
