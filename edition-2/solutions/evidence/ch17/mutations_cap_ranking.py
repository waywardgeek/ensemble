"""Focused sensitivity checks for independent pre-submission review findings."""
import pathlib,shutil,tempfile,subprocess,json
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
formula='score += idf * f * 2.2 / (f + 1.2*(0.25+0.75*float64(c.length)/x.average))'
cases=[
 ('capped-identity','internal/recall/recall.go','!seen[snippetHash(h)] && ','','./ensemble','^TestRecallCappedDedup$','capped snippet recalled twice'),
 ('capped-replay','internal/history/reduce.go','RecallHashes: e.Recall.Hashes','RecallHashes: nil','./ensemble','^TestRecallCappedDedup$','capped attachment or identity missing'),
 ('raw-frequency','internal/recall/index.go',formula,'score += f + idf*0','./internal/recall','^TestBM25CompetingDocuments$','ranking chose'),
 ('rarity-weight','internal/recall/index.go',formula,'score += (idf*0+1) * f * 2.2 / (f + 1.2*(0.25+0.75*float64(c.length)/x.average))','./internal/recall','^TestBM25CompetingDocuments$/rarity$','rarity ranking chose'),
 ('tf-saturation','internal/recall/index.go',formula,'score += idf * f * 2.2 / (1.2*(0.25+0.75*float64(c.length)/x.average))','./internal/recall','^TestBM25CompetingDocuments$/saturation$','saturation ranking chose'),
 ('length-normalization','internal/recall/index.go',formula,'score += idf * f * 2.2 / (f + 1.2)','./internal/recall','^TestBM25CompetingDocuments$/length$','length ranking chose'),
]
temp=pathlib.Path(tempfile.mkdtemp(prefix='ch17-cap-ranking-'))
results=[]
try:
 shutil.copytree(ROOT/'solutions/edition-2/main',temp/'source',ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 tree=temp/'source'
 baseline=subprocess.run(['go','test','./ensemble','./internal/recall','-run','^Test(RecallCappedDedup|BM25CompetingDocuments)$','-count=1'],cwd=tree,capture_output=True,text=True)
 (OUT/'mutation-cap-ranking-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,path,old,new,package,test,expected in cases:
  file=tree/path;original=file.read_text()
  if original.count(old)!=1:raise RuntimeError(name+' occurrence mismatch')
  file.write_text(original.replace(old,new))
  run=subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout','10s'],cwd=tree,capture_output=True,text=True)
  output=run.stdout+run.stderr;(OUT/('mutation-'+name+'.txt')).write_text(output)
  killed=run.returncode!=0 and expected in output and '[build failed]' not in output
  results.append({'claim':name,'file':path,'old':old,'new':new,'test':test,'exit':run.returncode,'intended_failure':killed})
  print(name,killed,flush=True);file.write_text(original)
 (OUT/'mutations-cap-ranking.json').write_text(json.dumps(results,indent=2))
finally:shutil.rmtree(temp)
