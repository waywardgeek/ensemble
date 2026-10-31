"""Bounded human CLI skills exercise; credentials stay in memory and child env."""
import json, os, pathlib, subprocess, tempfile, time

out = pathlib.Path(__file__).resolve().parent
settings = json.loads((pathlib.Path.home() / '.cr/settings.json').read_text())
providers = [('anthropic', 'directClaudeAPIKey', 'claude-haiku-5-5'), ('openai', 'directOpenAIAPIKey', 'gpt-4o-mini'), ('gemini', 'directGeminiAPIKey', 'gemini-3.8-flash')]
for vendor, field, model in providers:
    with tempfile.TemporaryDirectory(prefix='ensemble-ch10-live-') as scratch:
        directory = pathlib.Path(scratch)
        skills = {
            'base': '---\nname: base\ntype: primary\ntools: read_file\nloadable-skills: code-tools\n---\nYou are a concise coding assistant. Available tools: $TOOLS. Skills you can load: $SKILLS.',
            'code-tools': '---\nname: code-tools\ntype: loadable\ntools: write_file\ndepends: helper\n---\nWrite note.txt with content $CUSTOM_VAR using write_file, then use read_file to verify it. Next available skills: $SKILLS.',
            'helper': '---\nname: helper\ntype: dependency\ntools: list_directory\n---\nWork only in the current scratch directory.'
        }
        for name, text in skills.items():
            d = directory / 'skills' / name
            d.mkdir(parents=True)
            (d / 'SKILL.md').write_text(text)
        marker = 'CH10_' + vendor.upper() + '_OK'
        env = {k: v for k, v in os.environ.items() if k not in ['LLM_API_KEY', 'ANTHROPIC_API_KEY', 'OPENAI_API_KEY', 'GEMINI_API_KEY']}
        env.update(LLM_VENDOR=vendor, LLM_MODEL=model, LLM_API_KEY=settings[field], EN_SKILLS_DIR=str(directory/'skills'), EN_PRIMARY_SKILL='base', EN_CUSTOM_VAR=marker)
        prompt = 'Load the code-tools skill, follow its instructions once, then call unload_skill for code-tools and answer with the verified file content. No other work.\n'
        started = time.monotonic()
        child = subprocess.Popen(['/tmp/ensemble-ch10', 'chat'], cwd=scratch, env=env, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        timed_out = False
        try:
            stdout, stderr = child.communicate(prompt, timeout=60)
        except subprocess.TimeoutExpired:
            timed_out = True
            child.terminate()
            stdout, stderr = child.communicate(timeout=5)
        # Redact before any retained artifact or output. Application logging never
        # records headers, but this also guards unexpected provider echoes.
        def safe(text):
            for secret_field in ['directClaudeAPIKey', 'directOpenAIAPIKey', 'directGeminiAPIKey']:
                key = settings.get(secret_field)
                if key: text = text.replace(key, '[REDACTED]')
            return text
        events = []
        log = directory / 'ch02.jsonl'
        if log.exists():
            events = [json.loads(line) for line in safe(log.read_text()).splitlines() if line]
        calls = [e['tool'] for e in events if e.get('type') == 'tool_called']
        results = [e['tool'] for e in events if e.get('type') == 'tool_returned']
        api = directory / 'api.log'
        request_bodies = []
        if api.exists():
            for line in safe(api.read_text()).splitlines():
                entry = json.loads(line)
                if entry.get('event') == 'request':
                    request_bodies.append(json.loads(entry['data']))
        data = {'provider': vendor, 'model': model, 'elapsed_seconds': round(time.monotonic()-started, 2), 'exit': child.returncode, 'timed_out': timed_out, 'prompt': prompt, 'stdout': safe(stdout), 'stderr': safe(stderr), 'calls': calls, 'results': results, 'file_text': safe((directory/'note.txt').read_text()) if (directory/'note.txt').exists() else None, 'request_bodies': request_bodies}
        (out/(vendor+'-live.json')).write_text(json.dumps(data, indent=2)+'\n')
        print(vendor, 'exit', child.returncode, 'timeout', timed_out, 'calls', [c['name'] for c in calls], 'file_match', data['file_text'] == marker, flush=True)
