"""Verify all immutable identities before replay, redaction or derived writes."""
import argparse
import json
from pathlib import Path
import subprocess
from evidence import HERE, preflight, check_launch, file_digest, digest, scenario_catalog, scenario_inputs

def execute(command, run):
    return subprocess.check_output(command,cwd=run)

def verify(binding, runs, paths, output, redactions=None):
    preflight(binding, paths)
    assert runs, 'empty run set; no evidence modified'
    launches=[]
    for run in runs:
        launch=json.loads((run/'launch.json').read_text())
        check_launch(binding, launch)
        assert launch['exit_code']==0, 'incomplete launch; no evidence modified'
        assert launch['requests']<=launch['http_cap'], 'request cap exceeded; no evidence modified'
        assert launch['requests']>0, 'empty request set is not a paid demonstration; no evidence modified'
        captures=sorted((run/'requests').glob('*.json'))
        responses=sorted((run/'responses').glob('*.body'))
        expected_requests=[f'requests/{n:03}.json' for n in range(1,launch['requests']+1)]
        expected_responses=[f'responses/{n:03}.body' for n in range(1,launch['requests']+1)]
        assert [str(p.relative_to(run)) for p in captures]==expected_requests, 'capture count mismatch; no evidence modified'
        assert [str(p.relative_to(run)) for p in responses]==expected_responses, 'response count mismatch; no evidence modified'
        expected_logs=['workspace/alpha/events.jsonl','workspace/beta/events.jsonl'] if launch['scenario']=='P' else ['session.log']
        assert launch['logs']==expected_logs, 'missing log identities; no evidence modified'
        assert launch.get('workspace')==str(run/'workspace'), 'workspace identity mismatch; no evidence modified'
        catalog={p:digest(b) for p,b in scenario_catalog(binding,launch['scenario']).items()}
        assert launch.get('catalog_files')==catalog, 'launch catalog file mismatch; no evidence modified'
        inputs=scenario_inputs(launch['scenario'])
        assert launch.get('scratch_inputs')=={'workspace/'+p:digest(b) for p,b in inputs.items()}, 'scratch input mismatch; no evidence modified'
        retained_inputs={'inputs/'+p:digest(b) for p,b in inputs.items()}
        assert launch.get('scratch_outputs'), 'missing scratch output identities; no evidence modified'
        assert all(p.startswith('workspace/') for p in launch['scratch_outputs']), 'invalid scratch output identity; no evidence modified'
        outputs={str(p.relative_to(run)) for p in (run/'workspace').rglob('*') if p.is_file()}-set(expected_logs)
        assert outputs==set(launch['scratch_outputs']), 'incomplete scratch output identities; no evidence modified'
        required=set(expected_requests+expected_responses+expected_logs+['terminal.txt'])|set(catalog)|set(retained_inputs)|set(launch['scratch_outputs'])
        if launch['scenario']=='G':
            required.update(['gui.jsonl','browser-binding.json'])
            browser=json.loads((run/'browser-binding.json').read_text())
            assert browser['source_revision']==binding['source_revision'] and browser['executables']==launch['executables'] and browser['support']==launch['support'], 'browser launch identity mismatch; no evidence modified'
            assert 'browser-original.jsonl' in browser['originals'] and any(p.endswith('.png') for p in browser['originals']), 'incomplete browser receipts; no evidence modified'
            required.update(browser['originals'])
            assert all(launch['originals'].get(p)==sha for p,sha in browser['originals'].items()), 'browser receipt identity mismatch; no evidence modified'
        assert required<=set(launch['originals']), 'incomplete original receipt set; no evidence modified'
        for path,sha in {**catalog,**retained_inputs,**launch['scratch_outputs']}.items():
            assert launch['originals'][path]==sha, 'launch file identity mismatch; no evidence modified'
        for path, sha in launch['originals'].items():
            candidate=(run/path).resolve()
            assert candidate.is_relative_to(run.resolve()), 'escaping receipt path; no evidence modified'
            assert file_digest(candidate)==sha, 'original receipt mismatch; no evidence modified'
        logs=[]
        for path in expected_logs:
            log=run/path
            events=[json.loads(line) for line in log.read_text().splitlines() if line.strip()][1:]
            assert events, 'empty event log; no evidence modified'
            logs.append((log,events))
        assert sum(e['type']=='request_sent' for _,events in logs for e in events)==launch['requests'], 'log request count mismatch; no evidence modified'
        launches.append((run,launch,captures,logs))
    known_logs={str(log.resolve()):events for _,_,_,logs in launches for log,events in logs}
    for log,seq in (redactions or {}).items():
        assert log in known_logs and any(e['seq']==seq and e['type']=='tool_returned' for e in known_logs[log]), 'unbound redaction target; no evidence modified'
    results=[];derived=[]
    for run,launch,captures,logs in launches:
        renders=[]
        for log,events in logs:
            for event in events:
                if event['type']!='request_sent':continue
                body=execute([paths.get('cli',binding['executables']['cli']['path']),'replay',str(log),str(event['seq'])],run)
                # CLI adds one final output LF. Original HTTP bytes remain exact.
                body=body.removesuffix(b'\n')
                renders.append(body)
                derived.append((f'{run.name}-{log.parent.name}-{log.name}-{event["seq"]}.json',body+b'\n'))
            if redactions and str(log.resolve()) in redactions:
                body=execute([paths.get('redaction',binding['executables']['redaction']['path']),'--log',str(log),'--result',str(redactions[str(log.resolve())])],run)
                result=json.loads(body)
                assert result['original_sha256']==file_digest(log) and result['new_model_requests']==0, 'invalid redaction receipt'
                derived.append((f'{run.name}-{log.parent.name}-{log.name}-redaction.json',body))
        # Multi-Agent request order may differ. Compare exact byte multisets,
        # never parsed/re-encoded equivalents mislabeled byte-exact captures.
        assert sorted(p.read_bytes() for p in captures)==sorted(renders), 'request reconstruction mismatch: '+run.name
        results.append(dict(run=run.name,requests=len(captures),reconstructed=len(renders),comparison='exact HTTP body bytes; CLI output LF removed'))
    output.mkdir()
    for name,body in derived:(output/name).write_bytes(body)
    (output/'verification.json').write_text(json.dumps(results,indent=2)+'\n')
    return results

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('runs',nargs='+',type=Path);p.add_argument('--binding',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--cli');p.add_argument('--redaction');p.add_argument('--redact-log',type=Path);p.add_argument('--result',type=int);args=p.parse_args()
    binding=json.loads(args.binding.read_text());paths={k:getattr(args,k) for k in ('cli','redaction') if getattr(args,k)}
    redactions={str(args.redact_log.resolve()):args.result} if args.redact_log else None
    print(json.dumps(verify(binding,[r.resolve() for r in args.runs],paths,args.output,redactions),indent=2))
