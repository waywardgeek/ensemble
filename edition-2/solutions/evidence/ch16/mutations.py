"""One-use Chapter16 sensitivity checks in a disposable source copy."""
import json,pathlib,shutil,subprocess,tempfile,sys
ROOT=pathlib.Path(__file__).resolve().parents[4]
SRC=ROOT/'solutions/edition-2/main'; OUT=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path(tempfile.mkdtemp(prefix='ch16-mutants-'))
shutil.copytree(SRC,scratch,dirs_exist_ok=True,ignore=shutil.ignore_patterns('node_modules','cr','*.log'))
cases=[
('checkpoint-hook','internal/agent/dispatch.go','if t.checkpoint {','if false && t.checkpoint {','TestMemoryCompression','compressions='),
('prior-not-lifetime','internal/agent/memory.go','u := c.LastUsage','u := c.Usage','TestMemoryForcing','tools not restored'),
('include-cache-read','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input + u.CacheWrite','TestMemoryForcing','90% warning'),
('exclude-output','internal/agent/memory.go','input := u.Input + u.CacheRead + u.CacheWrite','input := u.Input + u.CacheRead + u.CacheWrite + u.Output','TestMemoryForcing','forcing premature'),
('restore-tools','internal/agent/dispatch.go','r.owner.forceHandoff.Store(false)','r.owner.forceHandoff.Store(true)','TestMemoryForcing','tools not restored'),
('source-to-compressor','internal/memory/memory.go','task.Text = input.String()','task.Text = "lost source"','TestMemoryCompression','compressor lost source'),
('recorded-cut','internal/history/reduce.go','redact(c, pending.Cut)','// deliberately omitted cut','TestMemoryCompression','did not replace raw'),
('persist-memory','internal/memory/memory.go','return os.Rename(p+".tmp", p)','return os.Remove(p+".tmp")','TestMemoryCompression','memory not persisted'),
('fresh-disk','internal/agent/memory.go','fresh := r.owner.history.LastSeq() == 0','fresh := false','TestMemoryDiskAndToggles','noncanonical startup'),
('reenable-disk','internal/agent/memory.go','fresh || wasDisabled','fresh','TestMemoryDiskAndToggles','unchanged re-enabled entries differ'),
('canonical-order','internal/history/reduce.go','return fileLess(a.File, b.File)','return fileLess(b.File, a.File)','TestMemoryDiskAndToggles','noncanonical startup'),
('disabled-neighbor','internal/memory/memory.go','if m.Settings(b - 1).Disabled {','if false && m.Settings(b - 1).Disabled {','TestMemoryWaterfall','disabled destination'),
('waterfall','internal/agent/memory.go','r.checkpoint(false)','r.send()','TestMemoryWaterfall','waterfall incomplete'),
('oldest-first','internal/memory/memory.go','files = append(files, e)','files = append([]common.Entry{e}, files...)','TestMemoryWaterfall','did not select oldest'),
('archive-retired','internal/agent/memory.go','if err := r.owner.memory.Archive(t); err != nil {','if err := error(nil); err != nil {','TestMemoryWaterfall','graduated files remained'),
('abandoned-cut','internal/history/reduce.go','c.Compression.Cut = e.Redact','redact(c,e.Redact)','TestMemoryAbandonment','uncommitted cut lost source'),
('empty-submit','internal/engine/compress.go','strings.TrimSpace(args.Memory) == ""','strings.TrimSpace(args.Memory) == "NEVER"','TestMemoryFailureKeepsSource/empty-submit','failed compression lost source'),
('memory-data','internal/engine/openai.go','entry.Kind == common.KindSkill','entry.Kind == common.KindSkill || entry.Kind == common.KindSession','TestMemoryCompression/openai','memory rendered as system'),
]
cases += [
('restart-reenable','internal/agent/memory.go','wasDisabled = r.owner.history.Context().DisabledBands[b]','wasDisabled = false','TestMemoryRestartReenable','restart re-enable missed'),
('interrupt-keeps-source','internal/agent/memory.go','t := r.active\n\tr.workers++','if task.Source==0 { r.append(common.Event{Type:"redacted",Redact:&common.RedactData{From:task.From,To:task.To,Level:"redact_summary",Reason:"premature cut"}}) }\n t := r.active\n r.workers++','TestMemoryInterruptedCompressor','interrupted compressor committed'),
('settings-effective','gui/settings.go','s.agent.SetMemorySettings(next.Memory)\n\treturn next','// omitted memory application\n return next','TestSettingsWireAndDisk','memory patch not effective'),
('settings-load','gui/settings.go','if err := s.agent.SetMemorySettings(next.Memory); err != nil {','if err := error(nil); err != nil {','TestSettingsWireAndDisk','memory restart lost patch'),
]
suffix=('-r3' if sys.argv[1:]==['settings-effective'] else '-r2' if 'settings-effective' in sys.argv else '-r1') if len(sys.argv)>1 else ''
if len(sys.argv)>1:cases=[c for c in cases if c[0] in sys.argv[1:]]
if sys.argv[1:]==['whole-segments']:
 suffix='-segments'
 cases=[('whole-segments','internal/memory/memory.go','if total-removed <= s.Budget {','if false && total-removed <= s.Budget {','TestMemoryWholeSegments','cut was not the oldest whole segment')]
if sys.argv[1:]==['startup-idle']:
 suffix='-startup-idle'
 cases=[
 ('enabled-restart','internal/agent/memory.go','} else {','} else if r.owner.history.LastSeq() == 0 || wasDisabled {','TestMemoryRestartEditedEnabled','enabled restart missed human disk correction'),
 ('idle-disable','internal/agent/actor.go','m.reply <- r.syncMemory()','m.reply <- error(nil)','TestMemoryDiskAndToggles','disabled session remains before next request'),
 ]
results=[]
try:
 baseline=subprocess.run(['go','test','./ensemble','-run','TestMemory','-count=1'],cwd=scratch,text=True,capture_output=True,timeout=30)
 (OUT/('mutation-baseline'+suffix+'.txt')).write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,file,old,new,test,expected in cases:
  path=scratch/file;source=path.read_text();assert old in source,name
  path.write_text(source.replace(old,new,1))
  try:
   run=subprocess.run(['go','test','./...' if file.startswith('gui/') else './ensemble','-run',test,'-count=1'],cwd=scratch/('gui' if file.startswith('gui/') else '.'),text=True,capture_output=True,timeout=30)
   output=run.stdout+run.stderr
   result={'claim':name,'file':file,'old':old,'new':new,'test':test,'exit':run.returncode,'intended_failure':run.returncode!=0 and expected in output and '[build failed]' not in output}
  except subprocess.TimeoutExpired as exc:
   output=str(exc);result={'claim':name,'timeout':True,'intended_failure':False}
  finally:path.write_text(source)
  (OUT/('mutation-'+name+suffix+'.txt')).write_text(output)
  results.append(result);print(json.dumps(result),flush=True)
 (OUT/('mutations'+suffix+'.json')).write_text(json.dumps(results,indent=2))
finally:shutil.rmtree(scratch)
