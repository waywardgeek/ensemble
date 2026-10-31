"""Bounded Chapter 19 human CLI exercise; loads only the authorized OpenAI key."""
import json, os, pathlib, shutil, subprocess, tempfile, time

repo = pathlib.Path.cwd()
source = repo / 'solutions/edition-2/main'
evidence = repo / 'solutions/edition-2/evidence/ch19'
key = json.loads(pathlib.Path('~/.cr/settings.json').expanduser().read_text())['directOpenAIAPIKey']
with tempfile.TemporaryDirectory(prefix='ensemble-ch19-live-') as temporary:
    work = pathlib.Path(temporary)
    binary = work / 'ensemble'
    subprocess.run(['go', 'build', '-o', str(binary), '.'], cwd=source, check=True)
    (work / 'puzzle.txt').write_text('Two identical workers, nonpreemptive jobs. Durations and prerequisites:\nA: 4, none\nB: 3, none\nC: 5, A\nD: 2, A\nE: 4, B\nF: 3, C and D\nG: 2, D and E\nH: 1, F and G\nFind a minimum-makespan valid schedule. Justify optimality, not merely validity.\n')
    env = {name:value for name,value in os.environ.items() if not name.endswith('API_KEY')}
    env.update(LLM_VENDOR='openai', LLM_MODEL='gpt-6.1-sol', LLM_API_KEY=key,
               LLM_BASE_URL='https://api.openai.com', OPENAI_BILLING_MODE='api',
               OPENAI_REASONING_EFFORT='high', OPENAI_REASONING_SUMMARY='detailed',
               EN_HOME=str(work/'private-auth'), EN_RECALL_DISABLED='1', EN_PRIMARY_SKILL='', EN_SKILLS_DIR='',
               CH02_LOG=str(work/'events.jsonl'))
    start=time.monotonic()
    hint=False
    with open(work/'stdout.txt','wb') as stdout, open(work/'stderr.txt','wb') as stderr:
        child=subprocess.Popen([str(binary),'chat'],cwd=work,env=env,stdin=subprocess.PIPE,stdout=stdout,stderr=stderr)
        child.stdin.write(b'Read puzzle.txt with read_file, solve its scheduling problem, and give a concise schedule and proof. Do not modify any files.\n'); child.stdin.flush()
        while child.poll() is None and time.monotonic()-start<150:
            trace=work/'api.log'
            if not hint and trace.exists() and 'response.reasoning_summary_text.delta' in trace.read_text():
                child.stdin.write(b'Keep both workers explicit in the answer, and verify every dependency before claiming optimality.\n');child.stdin.flush();child.stdin.close();hint=True
            time.sleep(.1)
        if child.poll() is None:child.kill()
        code=child.wait()
    # Preserve sanitized raw artifacts before deriving claims or deleting scratch.
    for name in ['stdout.txt','stderr.txt','api.log','events.jsonl','save.json']:
        path=work/name
        if path.exists():(evidence/('live-openai-'+name)).write_bytes(path.read_bytes().replace(key.encode(),b'[REDACTED]'))
    result={'source_commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),
            'model':'gpt-6.1-sol','billing':'explicit api','reasoning_effort':'high','reasoning_summary':'detailed',
            'hint_sent_after_summary':hint,'exit_code':code,'elapsed_seconds':round(time.monotonic()-start,2),
            'selected_model_discovery':True,'production_oauth_login':False}
    (evidence/'live-openai-result.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps(result))
