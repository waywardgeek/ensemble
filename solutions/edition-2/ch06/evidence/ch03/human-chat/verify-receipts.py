"""Reconstruct requests offline from immutable, source-bound terminal receipts."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

HERE = Path(__file__).resolve().parent
INITIAL_SOURCE = 'a347ce31511c4b124e486bb41ef98c07bd17cec5:solutions/edition-2/main'
VENDORS = ('anthropic', 'openai', 'gemini')


def preflight(args):
    # Validate all identities before executing the binary or creating output.
    # Read historical source from its commit, never from today's working tree.
    if args.source != INITIAL_SOURCE:
        raise ValueError('source must identify the immutable initial commit and path prefix')
    revision, prefix = args.source.split(':', 1)

    def historical(path):
        return subprocess.check_output(
            ['git', '-C', str(args.repository), 'show', f'{revision}:{prefix}/{path}'],
            stderr=subprocess.DEVNULL)

    evidence_prefix = 'evidence/ch03/human-chat/'
    binding_bytes = historical(evidence_prefix + 'initial-binding.json')
    if (args.evidence_dir / 'initial-binding.json').read_bytes() != binding_bytes:
        raise ValueError('initial binding differs from the immutable receipt')
    binding = json.loads(binding_bytes)
    binary = args.binary.resolve(strict=True)
    if hashlib.sha256(binary.read_bytes()).hexdigest() != binding['binary_sha256']:
        raise ValueError('executable hash differs from the recorded binary')
    paths = subprocess.check_output(
        ['git', '-C', str(args.repository), 'ls-tree', '-r', '--name-only', revision, '--', prefix],
        text=True).splitlines()
    sources = {path[len(prefix)+1:]: hashlib.sha256(historical(path[len(prefix)+1:])).hexdigest()
               for path in paths if path.endswith('.go')}
    if sources != binding['source_sha256']:
        raise ValueError('historical source differs from the recorded source hashes')
    for vendor in VENDORS:
        for kind in ('main', 'eof'):
            name = f'{vendor}-{kind}'
            run = args.evidence_dir / name
            # Launch metadata, raw terminal bytes and log records are immutable.
            for filename in ('launch.json', 'terminal.txt', 'session.log'):
                if (run / filename).read_bytes() != historical(evidence_prefix + name + '/' + filename):
                    raise ValueError(f'{name}/{filename} differs from the immutable receipt')
            launch = json.loads((run / 'launch.json').read_text())
            if launch['source_sha256'] != sources or launch['binary_sha256'] != binding['binary_sha256']:
                raise ValueError(f'{name} launch binding differs from the initial binding')
    if args.output_dir.exists():
        raise ValueError('output directory must be new; original receipts are never overwritten')
    if args.output_dir.is_relative_to(args.evidence_dir):
        raise ValueError('offline output must be outside the original evidence directory')
    return binary, binding


def objects(value):
    if isinstance(value, dict):
        yield value
        for child in value.values():
            yield from objects(child)
    elif isinstance(value, list):
        for child in value:
            yield from objects(child)


def check(run, binary, output):
    launch = json.loads((run / 'launch.json').read_text())
    records = [json.loads(line) for line in (run / 'session.log').read_text().splitlines()]
    events = records[1:]
    calls = {e['tool']['call_id']: e for e in events if e['type'] == 'tool_called'}
    results = [e for e in events if e['type'] == 'tool_returned']
    assert set(calls) == {e['tool']['call_id'] for e in results}
    assert {e['tool']['name'] for e in calls.values()} == {'list_directory','read_file','search_files','write_file','edit_file','run_command'}
    target = next(e for e in results if e['seq'] == 5)
    target_id = target['tool']['call_id']
    terminal = (run / 'terminal.txt').read_text()
    assert f'5 tool_returned call_id={target_id}' in terminal
    assert '/redact 5 5 demonstration' in terminal
    assert launch['exit_code'] == 0
    assert (run / 'workspace/created.txt').read_text() == 'replacement\ntail\n'
    assert (run / 'workspace/anchors.txt').read_text() == 'first repaired\nsecond anchor\n'
    before = {str(p.relative_to(run)): hashlib.sha256(p.read_bytes()).hexdigest() for p in [run/'session.log', *sorted((run/'workspace').iterdir())] if p.is_file()}
    env = {k:v for k,v in os.environ.items() if not k.startswith(('LLM_','ANTHROPIC_','OPENAI_','GEMINI_'))}
    env.update(LLM_VENDOR=launch['vendor'], LLM_MODEL=launch['selected_model'], LLM_RESOLVED_MODEL=launch['selected_model'].removeprefix('models/'), CH02_LOG=str(run/'session.log'))
    redaction_seq = next(e['seq'] for e in events if e['type']=='redacted')
    request_checks=[]
    ephemeral=[]
    for event in events:
        if event['type'] != 'request_sent':
            continue
        seq=event['seq']
        prefix=output/f'prefix-before-{seq}.log'
        prefix.write_text('\n'.join(json.dumps(e,separators=(',',':')) for e in records if 'seq' not in e or e['seq']<seq)+'\n')
        a=subprocess.run([str(binary),'render',str(prefix)],env=env,capture_output=True,check=True)
        b=subprocess.run([str(binary),'render',str(prefix)],env=env,capture_output=True,check=True)
        assert a.stdout==b.stdout
        (output/f'request-before-{seq}.json').write_bytes(a.stdout)
        body=json.loads(a.stdout)
        ephemeral.append({'request_seq':seq,'consumed':event['request']['ephemera'],'marker_in_projection':b'LILAC-614' in a.stdout})
        if seq>redaction_seq:
            target_objects=[o for o in objects(body) if (o.get('type')=='tool_result' and o.get('tool_use_id')==target_id) or (o.get('role')=='tool' and o.get('tool_call_id')==target_id) or ('response' in o and o.get('id')==target_id)]
            assert len(target_objects)==1, (seq,target_objects)
            assert '[redacted]' in json.dumps(target_objects[0]), (seq,target_objects)
            assert 'notes.md' not in json.dumps(target_objects[0])
            request_checks.append(seq)
    assert request_checks
    assert sum(x['marker_in_projection'] for x in ephemeral)==1
    assert sum(bool(x['consumed']) for x in ephemeral)==1
    dump=subprocess.run([str(binary),'dump'],env=env,capture_output=True,check=True)
    (output/'offline-dump.jsonl').write_bytes(dump.stdout)
    assert [json.loads(x) for x in dump.stdout.splitlines()]==records
    after={p:hashlib.sha256((run/p).read_bytes()).hexdigest() for p in before}
    assert before==after
    usage={k:sum(e['response']['usage'][k] for e in events if e['type']=='response_ended') for k in ['input','cache_write','cache_read','output']}
    final=f"Final usage: input={usage['input']}, cache write={usage['cache_write']}, cache read={usage['cache_read']}, output={usage['output']}"
    assert final in terminal
    errors=[{'seq':e['seq'],'name':calls[e['tool']['call_id']]['tool']['name'],'args':calls[e['tool']['call_id']]['tool']['args'],'parts':e['tool']['parts']} for e in results if e['tool'].get('is_error')]
    assert {e['name'] for e in errors}>={'write_file','edit_file','read_file','search_files'}
    return {'vendor':launch['vendor'],'model':launch['selected_model'],'returned_models':sorted({e['response']['from']['model'] for e in events if e['type']=='response_ended'}),'human_turns':sum(e['type']=='message_received' and e['message']['actor']=='human' for e in events),'requests':len(ephemeral),'tool_calls':len(calls),'usage':usage,'redaction_target':5,'redacted_requests_verified':request_checks,'ephemeral':ephemeral,'errors':errors,'files_unchanged_by_replay':before,'exit_code':launch['exit_code'],'terminal_sha256':hashlib.sha256((run/'terminal.txt').read_bytes()).hexdigest()}

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--source', required=True, help='immutable outer commit:path-prefix')
    parser.add_argument('--repository', type=Path, required=True, help='outer Git repository')
    parser.add_argument('--evidence-dir', type=Path, default=HERE)
    parser.add_argument('--output-dir', type=Path, required=True, help='new directory for offline reconstructions')
    args = parser.parse_args()
    args.evidence_dir = args.evidence_dir.resolve()
    args.output_dir = args.output_dir.resolve()
    try:
        binary, binding = preflight(args)
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        parser.exit(1, f'Binding verification failed: {error}\n')
    args.output_dir.mkdir(parents=True)
    receipts = []
    for vendor in VENDORS:
        output = args.output_dir / (vendor + '-main')
        output.mkdir()
        receipts.append(check(args.evidence_dir / output.name, binary, output))
    (args.output_dir/'receipts.json').write_text(json.dumps(receipts,indent=2)+'\n')
    (args.output_dir/'reconstruction.json').write_text(json.dumps({
        'kind': 'offline reconstruction; not intercepted live HTTP',
        'source': args.source, 'binary_sha256': binding['binary_sha256'],
        'source_sha256': binding['source_sha256'],
    }, indent=2)+'\n')
    print(json.dumps([{k:v for k,v in x.items() if k in ['vendor','model','usage','human_turns','requests','tool_calls','exit_code']} for x in receipts],indent=2))


if __name__ == '__main__':
    main()
