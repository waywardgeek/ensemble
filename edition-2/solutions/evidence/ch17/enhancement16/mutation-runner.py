import pathlib,shutil,tempfile,subprocess,os,json,time,difflib,hashlib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch17/enhancement16';source=root/'solutions/edition-2/ch16'
cases=[]
def add(name,file,old,new,group):cases.append((name,file,old,new,group))
R='internal/recall/recall.go';I='internal/recall/index.go';J='internal/recall/judge.go'
add('missing-handoffs',R,'{"handoffs", filepath.Join(root, "handoffs")},','{"handoffs", filepath.Join(root, "absent-handoffs")},','index')
add('missing-skills',R,'if skills := r.parent.Skills(); skills != nil {','if skills := r.parent.Skills(); skills != nil && false {','index')
add('skip-newest',R,'body, err := os.ReadFile(path)','if strings.HasPrefix(entry.Name(), "zz-") { return nil }; body, err := os.ReadFile(path)','index')
add('stale-index',R,'r.indexes == nil || digest != r.digest','r.indexes == nil','index')
add('raw-frequency',I,'score += idf * f * 2.2 / (f + 1.2*(0.25+0.75*float64(c.length)/x.average))','_ = idf; score += f','rank')
add('no-idf',I,'score += idf * f * 2.2','_ = idf; score += f * 2.2','rank')
add('no-saturation',I,'score += idf * f * 2.2 / (f + 1.2*(0.25+0.75*float64(c.length)/x.average))','score += idf * f / (0.25+0.75*float64(c.length)/x.average)','rank')
add('no-length-normalization',I,'0.25+0.75*float64(c.length)/x.average','1.0','rank')
add('no-query-stopwords',I,'!strings.Contains(stops, " "+word+" ") && !seen[word]','(len(stops)>0) && !seen[word]','rank')
add('strip-document-stopwords',I,'for _, word := range tokens(c.content) {','for _, word := range queryTokens(c.content) {','rank')
add('no-paragraph-split',I,'strings.Split(part.content, "\\n\\n")','[]string{part.content}','chunk')
add('no-child-breadcrumb',I,'header = parent + " > " + header','header = header','chunk')
add('candidate-cap-before-quota',R,'chosen = quotas(hits, s.MaxSnippets)','chosen = quotas(hits[:min(len(hits),s.Candidates)], s.MaxSnippets)','quota')
add('no-final-memory-quota',R,'map[string]int{"memory": (limit + 1) / 2}','map[string]int{"memory": 0}','quota')
add('no-spare-slots',R,'if len(out) < limit && !selected[i] {','if false && len(out) < limit && !selected[i] {','quota')
add('quota-overrides-judge',R,'if !judged {','if !judged || judged {','quota')
add('empty-judge-falls-back',R,'judged = err == nil','judged = err == nil && len(chosen)>0','judge')
add('mixed-index-partial-destination',J,'ints = nil // Unmarshal may partially populate a destination before rejecting a mixed array.','// defect: retain partially unmarshaled integers','judge')
add('no-judge-duplicate-filter',J,'&& !seen[i] && len(out)','&& len(out)','judge')
add('no-failure-fallback',R,'chosen = quotas(hits, s.MaxSnippets)','chosen = nil','fallback')
add('no-short-query-bypass',R,'s.Disabled || utf8.RuneCountInString(query) < s.MinQueryLength','s.Disabled || utf8.RuneCountInString(query) < 0','fallback')
add('no-capped-identity',R,'!seen[snippetHash(h)] && !strings.Contains(present, unquote(h.content))','!strings.Contains(present, unquote(h.content))','cap')
add('no-content-dedup',R,'!seen[snippetHash(h)] && !strings.Contains(present, unquote(h.content))','len(present)>=0','lifetime')
add('wrong-recall-kind','internal/history/reduce.go','Kind: common.KindRecall, Actor: common.Human','Kind: common.KindDialogue, Actor: common.Human','lifetime')
add('retain-recall-at-checkpoint','internal/history/reduce.go','if old.Kind != common.KindRecall {','if true {','lifetime')
add('unbounded-attachment',R,'text, hashes := format(chosen, s.MaxRecallBytes)','text, hashes := format(chosen, 999999)','cap')
add('recall-imports-spoke',R,'"ensemble/internal/common"','"ensemble/internal/common"\n_ "ensemble/internal/jobs"','topology')
add('reverse-spoke-edge','internal/engine/judge.go','"ensemble/internal/common"','"ensemble/internal/common"\n_ "ensemble/internal/recall"','topology')
results=[]
with tempfile.TemporaryDirectory(prefix='ch16-mutants-') as tmp:
 work=pathlib.Path(tmp)/'source';shutil.copytree(source,work);binary=pathlib.Path(tmp)/'gui'
 for name,file,old,new,group in cases:
  p=work/file;before=p.read_text();assert old in before,(name,file);after=before.replace(old,new);p.write_text(after)
  (out/(name+'.patch')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile=file,tofile=file)))
  build=subprocess.run(['go','build','-o',str(binary),'./cmd'],cwd=work/'gui',capture_output=True,text=True)
  if build.returncode:
   result={'name':name,'group':group,'build_exit':build.returncode,'output':build.stdout+build.stderr}
  else:
   env=dict(os.environ,RECALL_BIN=str(binary),RECALL_GROUP=group,RECALL_CORE=str(work),RECALL_GUI=str(work/'gui'))
   start=time.monotonic();r=subprocess.run(['/tmp/ch16-focus','-test.run','^TestChapter16Focus$','-test.v'],env=env,cwd=root,capture_output=True,text=True,timeout=90)
   result={'name':name,'group':group,'build_exit':0,'exit':r.returncode,'seconds':round(time.monotonic()-start,3),'output':r.stdout+r.stderr}
  (out/(name+'.txt')).write_text(result['output']);results.append(result);(out/'mutations.json').write_text(json.dumps(results,indent=2)+'\n');print(name,result.get('exit',result['build_exit']),flush=True);p.write_text(before)
print('scratch removed')
