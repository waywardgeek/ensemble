"""Reconcile retained crossover facts without repeating the paid experiment."""
import collections
import hashlib
import json
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[5]
evidence = root / 'solutions/edition-2/evidence/ch20'
out = evidence / 'enhancement19'
def read(name):
    return json.loads((evidence / name).read_text())
def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

original = read('live-result.json')
coordinator = read('coordinator-crossover-check.json')
reviewer = read('reviewer-check.json')
for name, digest in coordinator['sha256'].items():
    assert sha(evidence / name) == digest, name
wire = [json.loads(line) for line in (evidence / 'live-api.log').read_text().splitlines()]
requests = [json.loads(x['data']) for x in wire if x['event'] == 'request']
completed = [json.loads(x['data'])['response'] for x in wire if x['event'] == 'response.completed']
assert len(requests) == len(completed) == 5
for q, r in zip(requests, completed):
    assert q['model'] == r['model'] == 'gpt-6-astra'
    assert q['reasoning'] == {'effort': 'high', 'summary': 'detailed'}
    assert r['reasoning']['effort'] == 'high' and r['status'] == 'completed'
    assert q['stream'] is True and q['store'] is False
    assert 'prompt_cache_options' not in q and 'prompt_cache_breakpoint' not in json.dumps(q)
    assert r['usage']['input_tokens_details'] == {'cache_write_tokens': 0, 'cached_tokens': 0}
    assert r['usage']['output_tokens_details']['reasoning_tokens'] == 0
counts = collections.Counter(x['event'] for x in wire)
assert counts['response.reasoning_summary_text.delta'] == 0
assert counts['response.output_text.delta'] == 161
log = [json.loads(line) for line in (evidence / 'live-events.jsonl').read_text().splitlines()]
calls = [e['tool'] for e in log if e.get('type') == 'tool_called']
results = {e['tool']['call_id']: e['tool'] for e in log if e.get('type') == 'tool_returned'}
assert [x['name'] for x in calls] == ['read_file', 'read_file', 'run_command', 'edit_file', 'run_command']
assert len(results) == 5 and set(results) == {x['call_id'] for x in calls}
texts = [''.join(p.get('text', '') for p in results[c['call_id']]['parts']) for c in calls]
assert texts[0] == (evidence / 'task-before/aggregate.py').read_text()
assert texts[1] == (evidence / 'task-before/test_aggregate.py').read_text()
assert 'FAILED (failures=3, errors=3)' in texts[2] and 'exit_code: 1' in texts[2]
assert 'Ran 6 tests' in texts[4] and '\nOK\n' in texts[4] and 'exit_code: 0' in texts[4]
edit = calls[3]['args']
before = (evidence / 'task-before/aggregate.py').read_text()
after = (evidence / 'task-after/aggregate.py').read_text()
assert before.count(edit['old_text']) == 1
assert before.replace(edit['old_text'], edit['new_text'], 1) == after
assert (evidence / 'task-before/test_aggregate.py').read_bytes() == (evidence / 'task-after/test_aggregate.py').read_bytes()
assert reviewer['task_before']['exit'] == 1 and reviewer['task_after']['exit'] == 0
assert original['cli_exit_code'] == 0 and original['timed_out'] is False
save = read('live-save.json')
assert save['log'] == log[1:]
assert read('mutations.json') == [{'name': 'baseline', 'exit_code': 0}, {'name': 'astra-chat-route', 'exit_code': 1}, {'name': 'drop-reasoning-effort', 'exit_code': 1}]
assert 'model transport:' in (evidence / 'astra-chat-route.txt').read_text()
assert 'configured model/reasoning lost:' in (evidence / 'drop-reasoning-effort.txt').read_text()
# Bind the archived runtime, not today's concurrently edited main tree.
tree = subprocess.check_output(['git', 'rev-parse', 'HEAD:solutions/edition-2/ch19'], cwd=root, text=True).strip()
assert tree == 'bd2998e68f976dbaf437b642b131d350e086242f'
assert not subprocess.check_output(['git', 'diff', 'HEAD', '--', 'solutions/edition-2/ch19'], cwd=root)
report = dict(result='PASS', operation='read-only reconciliation of retained evidence; no model or test rerun',
    accepted_tree=tree, live_source=original['source_commit'], live_tree=original['main_tree'],
    requests=5, completed=5, calls=[x['name'] for x in calls], command_exits=[1,0],
    test_sha256=sha(evidence / 'task-before/test_aggregate.py'), exact_edit_reconstructs_file=True,
    full_save_log_matches=True, cli_exit=0, timed_out=False, recorded_seconds=original['elapsed_seconds'],
    public_summary_deltas=0, text_deltas=161, cache_read_and_write_grants=[0]*5,
    reused_mutations=read('mutations.json'),
    original_chapter_sha256=sha(root / 'book/chapter-20.md'),
    raw_sha256=coordinator['sha256'],
    limits=['Endpoint derives from source/configuration, not independently captured HTTP URL/headers.',
            'No new live run; existing reviewer before/after test reruns retain original provenance.',
            'One bounded CLI task; no general reliability, interactive steering, GUI or Edition 3 authorship claim.',
            'Zero reported reasoning tokens does not measure private reasoning.'])
(out / 'historical-audit.json').write_text(json.dumps(report, indent=2)+'\n')
print('PASS: historical five-request coding task, exact edit, unchanged tests, saved log and exit receipts')
