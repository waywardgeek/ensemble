"""Bounded human CLI recall exercise with an isolated representative corpus."""
import json, os, pathlib, shutil, subprocess, sys, tempfile, time
OUT=pathlib.Path(__file__).resolve().parent
provider=sys.argv[1]
models={'anthropic':'claude-haiku-5-5','openai':'gpt-4o-mini','gemini':'gemini-3.8-flash'}
fields={'anthropic':'directClaudeAPIKey','openai':'directOpenAIAPIKey','gemini':'directGeminiAPIKey'}
secret=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())[fields[provider]]
work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch17-'+provider+'-'))
for name in ['memory','docs']:(work/'cr'/name).mkdir(parents=True)
fact='## Quartz Zephyr migration decision\nProject Quartz Zephyr, release QZ-731: Morgan owns the migration rehearsal on October 23. We rejected automatic retry after timeout because the remote transaction may already have committed. Read the migration ledger before deciding whether retry is safe. Ledger placement remains unresolved: application database versus separate control database. Retain rollback artifacts for nine days.\n'
(work/'cr/memory/2026-10-01-1.md.retired').write_text(fact)
(work/'cr/docs/quartz.md').write_text('## Quartz Zephyr migration name\nQuartz Zephyr migration is also the name of our cafeteria recipe contest. Quartz Zephyr migration lunch recipes use apples and cinnamon, not database transactions.\n')
for i in range(25):
 (work/'cr/memory'/f'padding-{i}.md').write_text(f'## Garden log {i}\nPlant carrots by the apple trees and water in spring.\n')
 (work/'cr/docs'/f'padding-{i}.md').write_text(f'## Workshop {i}\nThe cabinet holds brushes and blue paint.\n')
env={k:v for k,v in os.environ.items() if not any(s in k.upper() for s in ['KEY','TOKEN','LLM_','ANTHROPIC_','OPENAI_','GEMINI_','EN_PRIMARY','EN_SKILLS','EN_RECALL'])}
env.update(LLM_VENDOR=provider,LLM_MODEL=models[provider],LLM_API_KEY=secret)
stdout=open(work/'stdout.txt','w');stderr=open(work/'stderr.txt','w')
p=subprocess.Popen(['/tmp/ensemble-ch17-cli','chat'],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
prompts=['From our archived Quartz Zephyr migration discussion, tell me the release code, owner and rehearsal date, rejected approach and why, the unresolved decision, and rollback retention. Use only the context already supplied; do not call tools. If a fact is absent, say so.','Who owns that rehearsal, and why was automatic retry rejected?']
receipts=[]
try:
 for prompt in prompts:
  before=(work/'ch02.jsonl').stat().st_mtime_ns if (work/'ch02.jsonl').exists() else 0
  p.stdin.write(prompt+'\n');p.stdin.flush()
  deadline=time.monotonic()+75;complete=False
  while time.monotonic()<deadline and p.poll() is None:
   if (work/'ch02.jsonl').exists() and (work/'ch02.jsonl').stat().st_mtime_ns!=before:
    try:
     rows=[json.loads(x) for x in (work/'ch02.jsonl').read_text().splitlines()]
     if len(rows)>1:complete=True;break
    except (ValueError,OSError):pass
   time.sleep(.2)
  receipts.append({'prompt':prompt,'completed':complete,'process_status':p.poll()})
  if not complete:break
 p.stdin.close()
 try:p.wait(timeout=15)
 except subprocess.TimeoutExpired:p.kill();p.wait()
finally:
 if p.poll() is None:p.kill();p.wait()
 stdout.close();stderr.close()
 dest=OUT/('live-'+provider);dest.mkdir(exist_ok=True)
 # Persist sanitized raw evidence before parsing or cleanup.
 for name in ['stdout.txt','stderr.txt','api.log','ch02.jsonl','save.json']:
  if (work/name).exists():(dest/name).write_text((work/name).read_text().replace(secret,'[REDACTED]'))
 (dest/'run.json').write_text(json.dumps({'provider':provider,'model':models[provider],'exit':p.returncode,'turns':receipts,'corpus_fact':fact},indent=2))
 shutil.rmtree(work)
 print(json.dumps({'provider':provider,'exit':p.returncode,'completed_turns':sum(x['completed'] for x in receipts),'receipt':str(dest)}))
