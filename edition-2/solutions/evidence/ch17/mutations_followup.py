"""Resolve the short-query survivor and check the newly strengthened assertions."""
import json, pathlib, shutil, subprocess, tempfile
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
cases=[
 ('short-query-r2','internal/recall/recall.go','utf8.RuneCountInString(query) < s.MinQueryLength','utf8.RuneCountInString(query) < 0','^TestRecallDedupRefreshIsolation$','short query recalled'),
 ('hint-during-recall','internal/history/reduce.go','e.Type == "hint_received" || ','','^TestRecallInterrupt$','hint became retained dialogue'),
 ('same-size-edit','internal/recall/recall.go','file, len(body), body','file, len(body), ""','^TestRecallDedupRefreshIsolation$','edit not indexed'),
 ('secondary-sources','internal/recall/recall.go','"memory", "docs", "handoffs", "learnings", "skills"','"memory", "docs"','^TestRecallSources$','missing source'),
 ('owner-corpus','internal/recall/recall.go','root := r.parent.Config().DataDir','root := "."','^TestRecallSources$','missing source'),
 ('judge-limit','internal/recall/judge.go','len(out) < s.MaxSnippets','len(out) < len(candidates)','^TestRecallJudge$/selected$','selection:'),
 ('judge-model','internal/engine/judge.go','cfg.Model = cfg.Recall.JudgeModel','cfg.Model = e.parent.Config().Model','^TestRecallJudge$/selected$','judge is not isolated tool-free route'),
 ('offline-replay','internal/history/history.go','func (h *history) replay(e common.Event) {','func (h *history) replay(e common.Event) {\n if e.Type == "recall_attached" { return }','^TestRecallTurnReplayCheckpoint$','replay changed bytes'),
 ('recall-disable','ensemble/ensemble.go','if !cfg.Recall.Disabled {','if true {','^TestRecallDedupRefreshIsolation$','disabled recall installed'),
]
scratch=pathlib.Path(tempfile.mkdtemp(prefix='ch17-followup-'))
results=[]
try:
 shutil.copytree(ROOT/'solutions/edition-2/main',scratch/'source',ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 tree=scratch/'source'
 base=subprocess.run(['go','test','./ensemble','-run','^TestRecall(DedupRefreshIsolation|Interrupt|Sources|Judge|TurnReplayCheckpoint)$','-count=1'],cwd=tree,capture_output=True,text=True)
 (OUT/'mutation-followup-baseline.txt').write_text(base.stdout+base.stderr)
 if base.returncode:raise RuntimeError('baseline failed')
 for name,path,old,new,test,expected in cases:
  file=tree/path;original=file.read_text()
  if original.count(old)!=1:raise RuntimeError(f'{name}: occurrence count {original.count(old)}')
  file.write_text(original.replace(old,new))
  result=subprocess.run(['go','test','./ensemble','-run',test,'-count=1','-timeout','10s'],cwd=tree,capture_output=True,text=True)
  output=result.stdout+result.stderr
  (OUT/('mutation-'+name+'.txt')).write_text(output)
  killed=result.returncode!=0 and expected in output and '[build failed]' not in output
  results.append({'claim':name,'file':path,'old':old,'new':new,'test':test,'exit':result.returncode,'intended_failure':killed})
  file.write_text(original)
  (OUT/'mutations-followup.json').write_text(json.dumps(results,indent=2))
  print(name,'killed' if killed else 'SURVIVED_OR_INVALID',flush=True)
finally:shutil.rmtree(scratch)
