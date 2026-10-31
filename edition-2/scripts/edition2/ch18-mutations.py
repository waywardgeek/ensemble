"""Focused defects in disposable student-source copies; preserve actual outcomes."""
import pathlib,shutil,tempfile,subprocess,json,sys
root=pathlib.Path(__file__).resolve().parents[2]
source=root/'solutions/edition-2/main'
out=root/'solutions/edition-2/evidence/ch18/mutations.json'
mutants=[
('main-observation','internal/engine/engine.go','if e.lens != nil {','if false && e.lens != nil {','TestCacheLensMainBoundary','./ensemble'),
('strip-markers','internal/cachelens/lens.go','func strip(body []byte) []byte {','func strip(body []byte) []byte {\n return body','TestCacheLensSerializedMarkers','./ensemble'),
('append-detection','internal/cachelens/lens.go','len(b) <= len(a)','len(b) >= 0','TestCacheLensMainBoundary','./ensemble'),
('gemini-array-exemption','internal/cachelens/lens.go','!conversation(now)','(!conversation(now) || key == "contents")','TestCacheLensMainBoundary/gemini','./ensemble'),
('system-marker','internal/engine/anthropic.go','"cache_control": map[string]string{"type": "ephemeral"}','"ignored_cache_control": map[string]string{"type": "ephemeral"}','TestCacheMarkers/claude','./ensemble'),
('tool-marker','internal/engine/anthropic.go','tools[len(tools)-1]["cache_control"]','tools[len(tools)-1]["ignored_cache_control"]','TestCacheMarkers/claude','./ensemble'),
('compaction-marker','internal/engine/cache.go','if e.Kind == common.KindHandoff {','if false && e.Kind == common.KindHandoff {','TestCacheMarkers','./ensemble'),
('retired-handoff-fallback','internal/engine/cache.go','if e.Kind >= common.KindSoul && e.Kind <= common.KindSession {','if false && e.Kind >= common.KindSoul && e.Kind <= common.KindSession {','TestCacheMarkers','./ensemble'),
('ephemeral-marker','internal/engine/anthropic.go','if index < len(c.Dialogue) {','if index <= len(c.Dialogue) {','TestCacheMarkers/claude','./ensemble'),
('openai-opt-in','internal/engine/openai.go','if landed {','if false && landed {','TestCacheMarkers/gpt-5.6','./ensemble'),
('openai-marker','internal/engine/cache.go','blocks[len(blocks)-1]["prompt_cache_breakpoint"]','blocks[len(blocks)-1]["ignored_breakpoint"]','TestCacheMarkers/gpt-5.6','./ensemble'),
('tool-reasoning','internal/engine/openai.go','effort = "none"','effort = "medium"','TestCacheMarkers/gpt-5.6','./ensemble'),
('last-vs-session','internal/engine/engine.go','e.last = event.Response.Usage','e.last = e.usage','TestCacheMeterSessionAndTiers','./ensemble'),
('large-prompt-tier','internal/engine/usage.go','if key.large {','if false && key.large {','TestCacheMeterSessionAndTiers','./ensemble'),
('effective-request-model','internal/engine/usage.go','key := usageKey{cfg.Vendor, cfg.Model,','key := usageKey{cfg.Vendor, e.parent.Config().Model,','TestCacheMeterAuxiliaryIdentity','./ensemble'),
('auxiliary-accounting','internal/engine/engine.go','e.account(cfg, event.Response.Usage)','if cfg.Recall.JudgeModel == "" { e.account(cfg, event.Response.Usage) }','TestCacheMeterAuxiliaryIdentity','./ensemble'),
('hit-rate','internal/engine/usage.go','out.HitRate = float64(out.Total.CacheRead) / float64(input)','out.HitRate = 0','TestCacheMeterSessionAndTiers','./ensemble'),
('anthropic-cache-write','internal/engine/anthropic.go','CacheWrite: wire.Usage.Write','CacheWrite: 0','TestCacheMeterSessionAndTiers/claude','./ensemble'),
('openai-disjoint-input','internal/engine/openai.go','*u.Input - u.Details.Read - u.Details.Write','*u.Input','TestCacheMeterSessionAndTiers/gpt','./ensemble'),
('gemini-thinking','internal/engine/gemini.go','Output: u.Output + u.Thoughts','Output: u.Output','TestCacheMeterSessionAndTiers/gemini','./ensemble'),
('gui-meter-snapshot','gui/gui.go','"input": u.Input','"input": u.Total.Input','TestUsageWireSession','./...'),
]
if len(sys.argv)>1:
 out=out.with_name('mutations-additional.json')
 mutants=[
 ('auxiliary-accounting-corrected','internal/engine/engine.go','e.account(cfg, event.Response.Usage)','if cfg.Model == e.parent.Config().Model {e.account(cfg, event.Response.Usage)}','TestCacheMeterAuxiliaryIdentity','./ensemble'),
 ('serialized-byte-count','internal/cachelens/lens.go','prefix += n','prefix += 2*n','TestCacheLensByteCounts','./ensemble'),
 ('empty-explicit-opt-in','internal/engine/openai.go','if landed {','if landed || true {','TestCacheEmptyOptIn','./ensemble'),
 ('known-model-repricing','internal/engine/usage.go','key := usageKey{cfg.Vendor, cfg.Model,','key := usageKey{cfg.Vendor, e.parent.Config().Model,','TestCacheMeterKnownJudge','./ensemble'),
 ('session-restoration','internal/engine/usage.go','for key, u := range e.byModel {','out.Total=e.parent.History().Context().Usage\n for key, u := range e.byModel {','TestCacheMeterSessionAndTiers','./ensemble'),
 ('gui-visible-input','gui/web/gui/app.js','message.input + message.cache_write + message.cache_read','message.input','browser','node'),
 ('gui-visible-unknown-price','gui/web/gui/app.js','message.priced ?','true ?','browser','node'),
 ]
if len(sys.argv)>1 and sys.argv[1]=='browser':
 out=out.with_name('mutations-browser-corrected.json')
 mutants=[m for m in mutants if m[-1]=='node']
if len(sys.argv)>1 and sys.argv[1]=='exclusion':
 out=out.with_name('mutations-exclusion.json')
 mutants=[
 ('unsupported-openai-marker','internal/engine/openai.go','if e.Features().ExplicitCache {','if true {','TestCacheMarkers/gpt-4o-mini','./ensemble'),
 ('judge-lens-exclusion','internal/engine/judge.go','e.send(ctx, body, common.StreamCallbacks{}, cfg)','e.Send(ctx, body, common.StreamCallbacks{})','TestCacheLensMainBoundary','./ensemble'),
 ('compressor-lens-exclusion','internal/engine/compress.go','e.send(ctx, body, common.StreamCallbacks{}, e.parent.Config())','e.Send(ctx, body, common.StreamCallbacks{})','TestCacheLensMainBoundary','./ensemble'),
 ]
if len(sys.argv)>1 and sys.argv[1]=='boundaries':
 out=out.with_name('mutations-boundaries.json')
 mutants=[
 ('rolling-marker-advance','internal/engine/anthropic.go','stableMessage = len(messages) - 1','stableMessage = 0','TestCacheMarkers/claude','./ensemble'),
 ('stable-rewrite-alarm','internal/cachelens/lens.go','if l.previous != nil && state != "identical" && !conversation(now) {','if false && l.previous != nil && state != "identical" && !conversation(now) {','TestCacheLensMainBoundary','./ensemble'),
 ]
results=[]
with tempfile.TemporaryDirectory(prefix='ensemble-ch18-mutations-') as tmp:
 tree=pathlib.Path(tmp)/'main';shutil.copytree(source,tree,ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
 (tree/'gui/node_modules').symlink_to(source/'gui/node_modules',target_is_directory=True)
 for name,path,old,new,test,package in mutants:
  file=tree/path;original=file.read_text()
  if old not in original:results.append({'name':name,'status':'setup-failure','reason':'anchor absent'});continue
  file.write_text(original.replace(old,new,1))
  cwd=tree/'gui' if path.startswith('gui/') else tree
  command=['node','--test','cache.test.mjs'] if package=='node' else ['go','test',package,'-run',test,'-count=1']
  run=subprocess.run(command,cwd=cwd,stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True,timeout=90)
  results.append({'name':name,'exit':run.returncode,'output':run.stdout})
  file.write_text(original)
  out.write_text(json.dumps(results,indent=2))
  print(name,'exit',run.returncode,flush=True)
