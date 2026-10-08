"""Credential-safe launcher using an explicit immutable revised binding."""
import argparse
import datetime
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import threading

from evidence import HERE


def module(name, filename):
    spec = importlib.util.spec_from_file_location(name, HERE / filename)
    result = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(result)
    return result


def launch(args):
    adapter = module('addendum', 'verify-addendum.py')
    binding = json.loads(args.binding.read_text())
    assert binding.get('addendum_support'), 'revised launcher requires its own bound support identities'
    paths = {n: d['path'] for n, d in binding['executables'].items()}
    if args.cli:
        paths['cli'] = str(args.cli.resolve())
    if args.workflow:
        paths['workflow'] = str(args.workflow.resolve())
    adapter.validate_binding(binding, paths)
    original = module('original_launcher', 'terminal-run.py')
    run = adapter.owned_path(args.name)
    run.mkdir()
    workspace = run / 'workspace'
    workspace.mkdir()
    (run / 'requests').mkdir()
    (run / 'responses').mkdir()
    key = original.credential(args.vendor)
    relay = original.Relay(args.vendor, key, run)
    thread = threading.Thread(target=relay.serve_forever)
    thread.start()
    env = {k: v for k, v in os.environ.items() if not k.startswith(('LLM_', 'ANTHROPIC_', 'OPENAI_', 'GEMINI_'))}
    env.update(LLM_VENDOR=args.vendor, LLM_MODEL=args.model, LLM_RESOLVED_MODEL=args.model.removeprefix('models/'),
               LLM_API_KEY=key, LLM_BASE_URL=f'http://127.0.0.1:{relay.server_port}', CH02_LOG=str(run / 'session.log'), ENSEMBLE_RUN_DIRECTORY=str(run))
    executable = 'cli' if args.mode == 'chat' else 'workflow'
    command = [paths[executable], args.mode]
    receipt = {'start': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'actor': 'Codex student coder, not Bill',
               'vendor': args.vendor, 'requested_model': args.model, 'command': command, 'workspace': str(workspace),
               'source_revision': binding['source_revision'], 'executables': {n: d['sha256'] for n, d in binding['executables'].items()},
               'support': binding['support'], 'addendum_support': binding['addendum_support'], 'mode': args.mode,
               'launched_executable': executable, 'executable_paths': paths,
               'transport': 'actual execution-tool PTY and macOS script terminal; raw body relay'}
    (run / 'launch.json').write_text(json.dumps(receipt, indent=2) + '\n')
    try:
        code = subprocess.call([paths['recorder'], '-q', str(run / 'terminal.txt'), *command], cwd=workspace, env=env)
    finally:
        relay.shutdown()
        relay.server_close()
        thread.join()
    receipt.update(exit_code=code, end=datetime.datetime.now(datetime.timezone.utc).isoformat(), requests=relay.number)
    (run / 'launch.json').write_text(json.dumps(receipt, indent=2) + '\n')
    raise SystemExit(code)


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('binding', type=Path)
    parser.add_argument('vendor', choices=['anthropic', 'openai', 'gemini'])
    parser.add_argument('model')
    parser.add_argument('mode', choices=['chat', 'workflow', 'collection'])
    parser.add_argument('name')
    parser.add_argument('--cli', type=Path)
    parser.add_argument('--workflow', type=Path)
    launch(parser.parse_args())
