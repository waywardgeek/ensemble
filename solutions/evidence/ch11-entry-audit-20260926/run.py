#!/usr/bin/env python3
"""One bounded audit; no retries, dependency downloads, or acceptance decisions."""
import json, os, pathlib, shutil, signal, subprocess, time
ROOT=pathlib.Path('/Users/bill/projects/ensemble')
E=ROOT/'solutions/evidence/edition-3/ch11-entry-audit-20260926'
GO='/opt/homebrew/bin/go'
ENV={'PATH':'/opt/homebrew/bin:/Users/bill/go/bin:/usr/bin:/bin:/usr/sbin:/sbin','HOME':str(E/'home'),'TMPDIR':str(E/'tmp'),'GOCACHE':'/Users/bill/Library/Caches/go-build','GOPATH':'/Users/bill/go','GOMODCACHE':'/Users/bill/go/pkg/mod','GOTOOLCHAIN':'local','GOWORK':'off','GOENV':'off','GOPROXY':'off','GOSUMDB':'off','GOFLAGS':'-mod=readonly -p=2','GOMAXPROCS':'2','EN_SKILLS_DIR':'','EN_PRIMARY_SKILL':'','LLM_API_KEY':'fake-entry-audit-key','LLM_VENDOR':'anthropic','LLM_MODEL':'claude-sonnet-5','LLM_BASE_URL':'http://127.0.0.1:1/unused','FAKEVENDOR_DUMP':str(E/'fakevendor-raw.log'),'LANG':'C','LC_ALL':'C'}
(E/'environment.json').write_text(json.dumps(ENV,indent=2)+'\n')
records=[]
def save(): (E/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
def run(name,args,limit):
 rec={'name':name,'argv':args,'cwd':str(ROOT),'time_limit_seconds':limit,'started':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),'free_before':shutil.disk_usage(ROOT).free}
 records.append(rec);save()
 if rec['free_before']<1073741824: rec['not_run']='1GiB floor';save();return False
 with (E/(name+'.stdout')).open('wb') as out,(E/(name+'.stderr')).open('wb') as err:
  p=subprocess.Popen(args,cwd=ROOT,env=ENV,stdout=out,stderr=err,start_new_session=True)
  rec['pid']=p.pid;rec['pgid']=p.pid;save();start=time.monotonic();minimum=rec['free_before']
  while p.poll() is None:
   free=shutil.disk_usage(ROOT).free;minimum=min(minimum,free)
   if free<1073741824 or time.monotonic()-start>limit:
    rec['stop_reason']='1GiB floor' if free<1073741824 else 'bounded time limit'
    os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=3)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL);p.wait()
    break
   time.sleep(0.25)
  rec['exit_code']=p.wait();rec['elapsed_seconds']=round(time.monotonic()-start,3);rec['minimum_free']=minimum
  # Only this invocation's group is eligible for cleanup.
  try:
   os.killpg(p.pid,0);rec['remaining_group_cleanup']=True;os.killpg(p.pid,signal.SIGTERM);time.sleep(.25)
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
  except ProcessLookupError:rec['remaining_group_cleanup']=False
 rec['free_after']=shutil.disk_usage(ROOT).free;save()
 print(name,rec,flush=True)
 return rec['exit_code']==0 and 'stop_reason' not in rec
if not run('build-driver',[GO,'build','-o',str(E/'bin/audit'),'./solutions/evidence/edition-3/ch11-entry-audit-20260926/driver'],90):raise SystemExit(1)
if not run('build-edition3',[GO,'build','-o',str(E/'bin/grade-edition3'),'./cmd/grade-edition3'],90):raise SystemExit(1)
# This unsupported dispatch must exit before grade.Build. Paths are explicit but are not launched.
core=str(ROOT/'solutions/ch11')
run('edition3-ch11-validation',[str(E/'bin/grade-edition3'),'-ch','11','-json','-core',core,'-consumer',core+'/ch05','-workflow',core+'/ch06','-stream-exercise',core+'/ch07','-stream-model','claude-sonnet-5','-gui',core+'/gui','-gui-model','claude-sonnet-5',core],10)
# Exactly one complete historical invocation, unmodified check implementation.
run('historical-ch11',[str(E/'bin/audit'),'baseline',str(ROOT),str(E)],300)
if shutil.disk_usage(ROOT).free>=1073741824:
 run('tunnel-input-pair',[str(E/'bin/audit'),'pair',str(ROOT),str(E)],55)
else:
 records.append({'name':'tunnel-input-pair','not_run':'1GiB floor'});save()
(E/'runner-finished.json').write_text(json.dumps({'finished':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),'free_bytes':shutil.disk_usage(ROOT).free},indent=2)+'\n')
