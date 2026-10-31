#!/usr/bin/env python3
"""Ch10 generated-reader allocation, Skills/storage refusal distinction and real-origin descriptor faults."""
import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

from accept_ch10_lifecycle import bind,digest,save_receipt

HERE=Path(__file__).resolve().parent
FIXTURES={
 'ch10_admission_allocation_test.go':'ch10-admission-allocation_test.go',
 'ch10_admission_skills_test.go':'ch10-admission-skills_test.go',
 'ch10_origin_faults_test.go':'ch10-origin-faults_test.go',
 'ch10_remaining_aggregate_test.go':'ch10-remaining-aggregate_test.go',
 'ch10_remaining_stream_test.go':'ch10-remaining-stream_test.go',
 'internal/persistence/ch10_origin_io_seam.go':'ch10-origin-io-seam.go',
}
MUTATIONS=[
 ('unbounded-record-preparation','internal/eventlog/prepared.go','data, err := encodeRecord(l.parent, event, common.SkillRecordLimit-1)','data, err := json.Marshal(event)','BoundedWriteAllocation','oversized write preparation accepted'),
 ('unbounded-record-read','internal/eventlog/log.go',"piece, err := input.ReadSlice('\\n')","piece, err := input.ReadBytes('\\n')",'BoundedAllocation','reader requested beyond admission fence'),
 ('skill-preflight-terminal','internal/eventlog/prepared.go','ok && problem.Code == "session_limit"','false && ok && problem.Code == "session_limit"','SkillsVersusStorage','expected controlled whole-record preflight refusal'),
 ('storage-overflow-nonterminal','ensemble.go','if persist && errors.As(err, &storage) && storage.Code == "session_limit" {','if false && persist && errors.As(err, &storage) && storage.Code == "session_limit" {','SkillsVersusStorage','terminal storage overflow retained skill admission'),
 ('origin-short-unchecked','internal/persistence/store.go','if err == nil && n != len(raw) {','if false && err == nil && n != len(raw) {','OriginFaults$/^origin-short$','origin fault exposed live Agent'),
 ('origin-write-unchecked','internal/persistence/store.go','if err != nil {\n\t\treturn s.failure("session_io", "cannot write immutable origin")','if false {\n\t\treturn s.failure("session_io", "cannot write immutable origin")','OriginFaults$/^origin-write$','origin fault exposed live Agent'),
 ('origin-sync-unchecked','internal/persistence/store.go','err = f.Sync()','err = nil','OriginFaults$/^origin-sync$','origin fault exposed live Agent'),
 ('origin-close-unchecked','internal/persistence/store.go','err = closeErr','err = nil; _ = closeErr','OriginFaults$/^origin-close$','origin fault exposed live Agent'),
]

def evaluate(args):
 source=args.source_directory.resolve(strict=True)
 commit,source_files=bind(source,args.source_commit)
 paths=[Path(__file__).resolve(),HERE/'accept_ch10_lifecycle.py',*(HERE/p for p in FIXTURES.values())]
 checkers={p.name:digest(p) for p in paths}
 result=dict(source_commit=commit,source_directory=str(source),source_files=source_files,checker_files=checkers,
             checks=[],complete=False,passed=False,full_chapter_acceptance=False,scope=__doc__)
 def save():save_receipt(args.receipt,result)
 save()
 with tempfile.TemporaryDirectory(prefix='ch10-admission-origin-') as d:
  work=Path(d)
  def run(command,expected=None):
   row=dict(command=command,expected_mutation_assertion=expected,started_utc=datetime.now(timezone.utc).isoformat(),disk_free_before=shutil.disk_usage(work).free,complete=False)
   result['checks'].append(row);save()
   try:
    p=subprocess.run(command,cwd=source,env=dict(os.environ,GOMEMLIMIT='1GiB',GOGC='50'),text=True,capture_output=True,timeout=240)
    row.update(exit=p.returncode,stdout=p.stdout,stderr=p.stderr,complete=True)
    row['passed']=(p.returncode!=0 and expected in p.stdout and '--- FAIL: TestCh10Admission' in p.stdout and 'panic: test timed out' not in p.stdout) if expected else (p.returncode==0 and (command[0]!='gofmt' or not p.stdout))
   except subprocess.TimeoutExpired as exc:
    def text(x):return x.decode(errors='replace') if isinstance(x,bytes) else x
    row.update(timeout=True,passed=False,stdout=text(exc.stdout),stderr=text(exc.stderr))
   finally:
    row.update(finished_utc=datetime.now(timezone.utc).isoformat(),disk_free_after=shutil.disk_usage(work).free);save()
   return row['passed']
  base={}
  for destination,fixture in FIXTURES.items():
   if (source/destination).exists():raise ValueError('overlay collision')
   local=work/fixture;local.write_bytes((HERE/fixture).read_bytes());base[str(source/destination)]=str(local)
  store=source/'internal/persistence/store.go'
  original=store.read_text()
  old='f, err := os.OpenFile(filepath.Join(s.path, "origin.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)'
  new='f, err := ch10ReviewOriginOpen(s, filepath.Join(s.path, "origin.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)'
  if original.count(old)!=1:raise ValueError('origin descriptor adapter anchor changed')
  adapted=original.replace(old,new,1)
  adapter=work/'store-adapter.go';adapter.write_text(adapted);base[str(store)]=str(adapter)
  result['adapter']=dict(path='internal/persistence/store.go',old=old,new=new,original_sha256=digest(store),adapted_sha256=digest(adapter));save()
  overlay=work/'positive.json';overlay.write_text(json.dumps({'Replace':base}))
  good=all(run(c) for c in [
   ['gofmt','-l',*base.values()],['go','vet','-overlay',str(overlay),'./...'],
   ['go','test','-overlay',str(overlay),'-count=1','-timeout=180s','-run','^TestCh10Admission','-v','.'],
   ['go','test','./...','-count=1'],
  ])
  if good and args.audit:
   for name,relative,old,new,group,expected in MUTATIONS:
    text=adapted if relative=='internal/persistence/store.go' else (source/relative).read_text()
    if relative=='internal/persistence/store.go':
     start=text.index('func (s *Store) WriteOrigin(');end=text.index('func (s *Store) Begin(',start)
     before,fragment,after=text[:start],text[start:end],text[end:]
    else:before,fragment,after='',text,''
    if fragment.count(old)!=1:raise ValueError('nonunique mutation anchor: '+name)
    local=work/(name+'.go');local.write_text(before+fragment.replace(old,new,1)+after)
    replacements=dict(base);replacements[str(source/relative)]=str(local)
    mutation=work/(name+'.json');mutation.write_text(json.dumps({'Replace':replacements}))
    result.setdefault('mutations',[]).append(dict(name=name,path=relative,old=old,new=new,changed_sha256=digest(local)));save()
    good=run(['go','test','-overlay',str(mutation),'-count=1','-timeout=180s','-run','^TestCh10Admission'+group,'-v','.'],expected) and good
 if bind(source,commit)[1]!=source_files or checkers!={p.name:digest(p) for p in paths}:raise ValueError('source/checker changed during checks')
 result.update(complete=True,passed=good);save();return good

if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__)
 p.add_argument('source_directory',type=Path);p.add_argument('--source-commit',required=True)
 p.add_argument('--receipt',type=Path,required=True);p.add_argument('--audit',action='store_true')
 raise SystemExit(0 if evaluate(p.parse_args()) else 1)
