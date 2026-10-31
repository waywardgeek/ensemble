"""Chapter-local one-defect checks in a disposable copy; never edit the grader."""
import json, pathlib, shutil, subprocess, tempfile
ROOT=pathlib.Path(__file__).resolve().parents[4]
SOURCE=ROOT/'solutions/edition-2/main'
OUT=pathlib.Path(__file__).resolve().parent
cases=[
 ('query-stopwords','internal/recall/index.go','!strings.Contains(stops, " "+word+" ")','(strings.Contains(stops, " "+word+" ") || true)','./internal/recall','^TestBM25$','query tokens'),
 ('document-stopwords','internal/recall/index.go','range tokens(c.content)','range queryTokens(c.content)','./internal/recall','^TestBM25$','document stop words removed'),
 ('small-idf','internal/recall/index.go','score += idf * f','idf = math.Max(0, math.Log((float64(len(x.chunks)-x.df[word])+0.5)/(float64(x.df[word])+0.5)))\n\t\t\tscore += idf * f','./internal/recall','^TestBM25$','search:'),
 ('score-floor','internal/recall/index.go','score >= floor','score >= floor || score > 0','./internal/recall','^TestBM25$','score floor ignored'),
 ('markdown-depth','internal/recall/index.go','sections(text, "## ", "")','sections(text, "##### ", "")','./internal/recall','^TestChunking$','sections:'),
 ('final-quotas','internal/recall/recall.go','chosen = quotas(hits, s.MaxSnippets)','chosen = hits[:min(len(hits),s.MaxSnippets)]','./ensemble','^TestRecallTurnReplayCheckpoint$','mechanical quota'),
 ('spare-quota','internal/recall/recall.go','len(out) < limit && !selected[i]','len(out) < 0 && !selected[i]','./internal/recall','^TestFinalQuotas$','unused memory slots lost'),
 ('mixed-parser','internal/recall/judge.go','ints = nil // Unmarshal','// ints = nil // Unmarshal','./internal/recall','^TestJudgeParsing$','[0,"3",null'),
 ('byte-cap','internal/recall/recall.go','if len(body) > room {','if len(body) > room && false {','./internal/recall','^TestAttachmentCap$','cap not reached'),
 ('judge-verdict','internal/recall/recall.go','judged = err == nil','judged = false','./ensemble','^TestRecallJudge/selected$','selection:'),
 ('empty-verdict','internal/recall/recall.go','judged = err == nil','judged = err == nil && len(chosen) > 0','./ensemble','^TestRecallJudge/empty$','selection:'),
 ('judge-fallback','internal/recall/recall.go','if !judged {','if !judged && err == nil {','./ensemble','^TestRecallJudge/failure$','selection:'),
 ('short-query','internal/recall/recall.go','utf8.RuneCountInString(query) < s.MinQueryLength','utf8.RuneCountInString(query) < 0','./ensemble','^TestRecallDedupRefreshIsolation$','short query recalled'),
 ('retired-corpus','internal/recall/recall.go','source.name == "memory" && strings.HasSuffix(path, ".md.retired")','source.name == "never" && strings.HasSuffix(path, ".md.retired")','./ensemble','^TestRecallTurnReplayCheckpoint$','mechanical quota'),
 ('index-refresh','internal/recall/recall.go','r.indexes == nil || digest != r.digest','r.indexes == nil','./ensemble','^TestRecallDedupRefreshIsolation$','newest excluded'),
 ('context-dedup','internal/recall/recall.go','!strings.Contains(present, unquote(h.content))','!strings.Contains(present, unquote(h.content)) || true','./ensemble','^TestRecallDedupRefreshIsolation$','already-present content recalled'),
 ('checkpoint-retires','internal/history/reduce.go','old.Kind != common.KindRecall','old.Kind != 255','./ensemble','^TestRecallTurnReplayCheckpoint$','checkpoint retained recall'),
 ('begin-hook','internal/agent/actor.go','r.recall()','r.send()','./ensemble','^TestRecallTurnReplayCheckpoint$','recall entries:'),
 ('recorded-bytes','internal/history/reduce.go','Parts: e.Recall.Parts','Parts: []common.Part{{Type:"text",Text:"lost"}}','./ensemble','^TestRecallTurnReplayCheckpoint$','mechanical quota'),
 ('judge-candidates','internal/recall/recall.go','min(len(hits), s.Candidates)','len(hits)','./ensemble','^TestRecallJudgeContext$','judge prompt limits wrong'),
 ('judge-recent','internal/recall/judge.go','recent = recent[len(recent)-s.ContextMsgs:]','recent = recent[:]','./ensemble','^TestRecallJudgeContext$','judge prompt limits wrong'),
 ('judge-deadline','internal/recall/judge.go','15*time.Second','30*time.Second','./ensemble','^TestRecallJudgeTimeout$','judge timeout/fallback'),
]
scratch=pathlib.Path(tempfile.mkdtemp(prefix='ch17-mutants-'))
results=[]
try:
 shutil.copytree(SOURCE,scratch/'source',ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 tree=scratch/'source'
 baseline=subprocess.run(['go','test','./internal/recall','./ensemble','-run','^Test(BM25|Chunking|FinalQuotas|JudgeParsing|AttachmentCap|Recall.*)$','-count=1','-timeout','45s'],cwd=tree,capture_output=True,text=True)
 (OUT/'mutation-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,path,old,new,package,test,expected in cases:
  file=tree/path;original=file.read_text()
  if original.count(old)!=1:raise RuntimeError(f'{name}: occurrence count {original.count(old)}')
  file.write_text(original.replace(old,new))
  result=subprocess.run(['go','test',package,'-run',test,'-count=1','-timeout','25s'],cwd=tree,capture_output=True,text=True)
  output=result.stdout+result.stderr
  (OUT/('mutation-'+name+'.txt')).write_text(output)
  killed=result.returncode!=0 and expected in output and '[build failed]' not in output
  results.append({'claim':name,'file':path,'old':old,'new':new,'test':test,'exit':result.returncode,'intended_failure':killed})
  file.write_text(original)
  (OUT/'mutations.json').write_text(json.dumps(results,indent=2))
  print(name,'killed' if killed else 'SURVIVED_OR_INVALID',flush=True)
finally:shutil.rmtree(scratch)
