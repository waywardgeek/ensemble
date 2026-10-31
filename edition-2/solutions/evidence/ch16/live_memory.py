"""Bounded actual GUI-terminal exercise; credentials stay in the intended child."""
import json, os, pathlib, shutil, socket, subprocess, sys, tempfile, time
ROOT=pathlib.Path(__file__).resolve().parents[4]
OUT=pathlib.Path(__file__).resolve().parent
provider=sys.argv[1]
models={'anthropic':'claude-haiku-5-5','openai':'gpt-4o-mini','gemini':'gemini-3.8-flash'}
fields={'anthropic':'directClaudeAPIKey','openai':'directOpenAIAPIKey','gemini':'directGeminiAPIKey'}
secret=json.loads((pathlib.Path.home()/'.cr/settings.json').read_text())[fields[provider]]
work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch16-'+provider+'-'))
(work/'cr').mkdir()
(work/'cr/settings.json').write_text(json.dumps({'tts_enabled':False,'max_tool_rounds':3,'memory':{'conversation':{'budget':600},'session':{'budget':500,'high_watermark':1000,'low_watermark':500}}}))
with socket.socket() as sock:sock.bind(('127.0.0.1',0));port=sock.getsockname()[1]
env={k:v for k,v in os.environ.items() if not any(s in k.upper() for s in ['KEY','TOKEN','LLM_','ANTHROPIC_','OPENAI_','GEMINI_','EN_PRIMARY','EN_SKILLS'])}
env.update(LLM_VENDOR=provider,LLM_MODEL=models[provider],LLM_API_KEY=secret)
stdout=open(work/'stdout.txt','w');stderr=open(work/'stderr.txt','w')
p=subprocess.Popen(['/tmp/ensemble-ch16-gui','--port',str(port),'--gui-dir',str(ROOT/'solutions/edition-2/main/gui/web/gui')],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr,text=True)
base='This is a bounded memory exercise in an empty scratch workspace. Do not use filesystem or command tools. Analyze the supplied case, explain the decision briefly, then call micro_handoff exactly once with the important facts and reasons. After the tool returns, finish this turn without another tool call. '
case1='Case one: Our project is Cedar Lantern, release code CL-721. We rejected automatic deletion of imported notes because the support team needs the originals when customers report mistakes. The replacement design keeps source notes in an archive and stores the accepted summary separately. Morgan owns the archive review, scheduled for October 18. Acceptance requires preserving a source identifier with each summary. The unresolved question is whether attachments belong in the same archive or a separate one; do not claim that has been decided. The default retention is thirty days. A failed summarization must leave the original note available. The initial approach coupled cleanup to request arrival, so an unrelated request could delete unreviewed work. We moved cleanup to an explicit completed-review boundary instead. Explain why the latter boundary matters, and preserve these facts in your checkpoint.'
case2='Case two: The deployment rehearsal uses Harbor Finch, diagnostic tag HF-934. Priya owns the rehearsal scheduled for October 21. We rejected retrying a database migration automatically after a timeout: the remote operation may already have committed. The chosen procedure reads the migration ledger before deciding whether any retry is safe. The operator must retain the exact migration identifier and measured error, not a guess about success. We have not settled whether the ledger should live alongside application data or in a separate control database. The rollback package must remain available for seven days after deployment. Last week a log line said the worker stopped, but the child process was still alive and held the port open. The rehearsal now checks the child lifetime directly. Explain how this evidence changes the decision and checkpoint the completed analysis, preserving the unresolved question as unresolved.'
prompts=[base+case1,base+case2,'Without using tools, recall both project names, diagnostic codes, owners, dates, rejected approaches and reasons. Explicitly identify the two unresolved questions. Do not invent a resolved answer.']
receipts=[]
try:
 for prompt in prompts:
  before=(work/'event.log').stat().st_mtime_ns if (work/'event.log').exists() else 0
  p.stdin.write(prompt+'\n');p.stdin.flush()
  deadline=time.monotonic()+85
  complete=False
  while time.monotonic()<deadline and p.poll() is None:
   if (work/'event.log').exists() and (work/'event.log').stat().st_mtime_ns != before:
    try:
     lines=[json.loads(s) for s in (work/'event.log').read_text().splitlines()]
     # r.save runs only after this human turn finishes, so a complete JSONL suffix is a receipt.
     if len(lines)>1:
      complete=True;break
    except (json.JSONDecodeError,OSError):pass
   time.sleep(.2)
  receipts.append({'prompt':prompt,'completed':complete,'process_status':p.poll()})
  if not complete:break
 p.stdin.close()
 try:p.wait(timeout=15)
 except subprocess.TimeoutExpired:p.kill();p.wait()
finally:
 if p.poll() is None:p.kill();p.wait()
 stdout.close();stderr.close()
 # Save sanitized raw receipts before interpreting results or cleaning scratch.
 dest=OUT/('live-'+provider+('-'+sys.argv[2] if len(sys.argv)>2 else ''));dest.mkdir(exist_ok=True)
 for name in ['stdout.txt','stderr.txt','api.log','debug.log','event.log','save.json']:
  if (work/name).exists():
   data=(work/name).read_text().replace(secret,'[REDACTED]');(dest/name).write_text(data)
 (dest/'run.json').write_text(json.dumps({'provider':provider,'model':models[provider],'exit':p.returncode,'turns':receipts},indent=2))
 for folder in ['memory','bucket-0','bucket-1']:
  for file in (work/'cr'/folder).glob('*.md*'):
   target=dest/'corpus'/folder/file.name;target.parent.mkdir(parents=True,exist_ok=True);target.write_text(file.read_text().replace(secret,'[REDACTED]'))
 for file in ['SOUL.md','MEMORY.md']:
  if (work/'cr'/file).exists():
   target=dest/'corpus'/file;target.parent.mkdir(parents=True,exist_ok=True);target.write_text((work/'cr'/file).read_text().replace(secret,'[REDACTED]'))
 shutil.rmtree(work)
 print(json.dumps({'provider':provider,'exit':p.returncode,'completed_turns':sum(x['completed'] for x in receipts),'receipt':str(dest)}))
