#!/usr/bin/env python3
"""Complete deterministic Chapter 6 gate from immutable committed main source.

Live demonstrations, receipt verification, architecture/code comparison and
manuscript review remain distinct acceptance gates. No paid calls are made.
"""
import argparse
import hashlib
import io
import json
from pathlib import Path
import subprocess
import tarfile
import tempfile

from accept_ch05 import command
from accept_ch06 import HERE, REPO, install

PREFIX = 'solutions/edition-2/main/'


def evaluate(revision):
    revision = subprocess.check_output(['git', 'rev-parse', revision+'^{commit}'], cwd=REPO, text=True).strip()
    archive = subprocess.check_output(['git', 'archive', revision, PREFIX], cwd=REPO)
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch06-complete-gate-') as directory:
        root = Path(directory)/'source'
        hashes = {}
        with tarfile.open(fileobj=io.BytesIO(archive)) as source:
            for item in source:
                if not item.isfile():
                    continue
                relative = Path(item.name).relative_to(PREFIX)
                if not (relative.suffix == '.go' or relative.name in ('go.mod', 'go.sum') or 'testdata' in relative.parts):
                    continue
                if '..' in relative.parts:
                    raise ValueError('unsafe source archive')
                data = source.extractfile(item).read()
                path = root/relative
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(data)
                hashes[str(relative)] = hashlib.sha256(data).hexdigest()
        def run(name, args, cwd=root):
            result = command(args, cwd, timeout=1200)
            rows.append(dict(id=name, passed=result['exit'] == 0, command=result))
            return result['exit'] == 0
        binary = Path(directory)/'cli'
        built = run('cli-build', ['go', 'build', '-o', binary, './cmd'])
        if built:
            run('cli-early-barriers', ['python3', HERE/'accept_ch06_clients.py', binary])
            run('wire-fixtures', ['python3', HERE/'accept_ch06_wire.py', binary])
        run('independent-contract', ['python3', HERE/'accept_ch06.py', root])
        run('retained-prior-assertions', ['python3', HERE/'accept_ch06_prior.py', root])
        run('implementation-deletions', ['python3', HERE/'audit_ch06_mutations.py', root])
        run('cli-overflow-controls', ['python3', HERE/'ch06-review-cli-recovery.py', revision], REPO)
        install(root)
        run('assembly-scaling-observation', ['go','test','-run','^$','-bench','BenchmarkC6AssemblyScaling','-benchtime=1x','-count=1','./internal/llm'])
        for module in sorted(root.rglob('go.mod')):
            name = str(module.parent.relative_to(root))
            run(name+'/vet', ['go','vet','./...'], module.parent)
            run(name+'/tests', ['go','test','./...','-count=1','-timeout=120s'], module.parent)
        run('fixture-formatting', ['gofmt','-l',root/'independent_ch06_public_test.go',root/'internal/llm/independent_ch06_internal_test.go'])
        if rows[-1]['command']['stdout'].strip():
            rows[-1]['passed'] = False
    return dict(scope=__doc__, source_revision=revision, source_files=hashes,
                checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(HERE.glob('*ch06*')) if p.is_file()},
                passed=all(row['passed'] for row in rows), checks=rows)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source_commit')
    args=parser.parse_args()
    result=evaluate(args.source_commit)
    print(json.dumps(result,indent=2))
    return 0 if result['passed'] else 1


if __name__=='__main__':
    raise SystemExit(main())
