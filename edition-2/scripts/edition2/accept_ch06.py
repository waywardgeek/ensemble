#!/usr/bin/env python3
"""Independent Chapter 6 acceptance runner. Live use and comparative review are separate gates.

Public/internal API spellings are adapters. Checks and mutants do not award
points for names or ancestry. Historical graders remain unchanged.
"""
import argparse
import hashlib
import json
import pathlib
import shutil
import tempfile

from accept_ch05 import command

HERE = pathlib.Path(__file__).resolve().parent
REPO = HERE.parents[1]
GROUPS = {
    'public-identity': ('^TestC6PublicEarlyIdentity$', '.'),
    'opaque-replay': ('^TestC6UnknownTextReplay$', '.'),
    'block-lifecycle': ('^TestC6MessagesLifecycle$', './internal/llm'),
    'typed-content': ('^TestC6TypedCallsAndOpaque$', '.'),
    'public-cancel': ('^TestC6PublicCancelProposedEffect$', '.'),
    'public-overflow': ('^TestC6PublicObserverOverflow$', '.'),
    'job-sequence': ('^TestC6JobSequenceInterleave$', '.'),
    'replay': ('^TestC6RecordedDeliveryReplay$', '.'),
    'framing-splits': ('^TestC6ExactFraming$', './internal/llm'),
    'capacity': ('^TestC6(CapacityCancellationAndUTF8|CompletionDrainWait)$', './internal/llm'),
    'timeout-capacity': ('^TestC6HTTPTimeoutIncludesCapacityWait$', './internal/llm'),
    'required-terminal': ('^TestC6RequiredTerminal$', './internal/llm'),
    'opaque-bound': ('^TestC6OpaqueAssemblyBound$', './internal/llm'),
    'assembly-bound': ('^TestC6AssembledBound$', './internal/llm'),
}


def install(root):
    if root == REPO or REPO in root.parents:
        raise ValueError('independent probes require disposable source outside repository')
    for name, parent in [('public', root), ('internal', root / 'internal/llm')]:
        (parent / ('independent_ch06_' + name + '_test.go')).write_bytes(
            (HERE / ('ch06_' + name + '_test.go.txt')).read_bytes())


def evaluate(source, groups=None, race=True):
    with tempfile.TemporaryDirectory(prefix='ensemble-ch06-independent-') as directory:
        root = pathlib.Path(directory) / 'source'
        shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence'))
        hashes = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(root.rglob('*')) if p.is_file() and (p.suffix == '.go' or p.name in ('go.mod', 'go.sum'))}
        install(root)
        checks = []
        for name in groups or GROUPS:
            pattern, package = GROUPS[name]
            cmd = ['go', 'test', '-json', '-count=1', '-timeout=90s', '-run', pattern]
            if race and name not in ('assembly-bound', 'opaque-bound'):
                cmd.append('-race')
            result = command(cmd + [package], root, timeout=120)
            try:
                ran = any(json.loads(line).get('Action') == 'run' for line in result['stdout'].splitlines())
            except (ValueError, TypeError):
                ran = False
            if not ran:
                result['exit'] = result['exit'] or 1
                result['stderr'] += '\nNo independent test executed.'
            checks.append(dict(id=name, passed=result['exit'] == 0, command=result))
        return dict(scope='incremental independent groups; complete gate and deletion audit pending',
                    source=str(source), source_files=hashes, checker_files={x.name: hashlib.sha256(x.read_bytes()).hexdigest() for x in [pathlib.Path(__file__), HERE/'ch06_public_test.go.txt', HERE/'ch06_internal_test.go.txt']}, passed=all(c['passed'] for c in checks), checks=checks)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('source', type=pathlib.Path)
    p.add_argument('--only', choices=GROUPS, action='append')
    p.add_argument('--no-race', action='store_true')
    a = p.parse_args()
    result = evaluate(a.source.resolve(strict=True), a.only, not a.no_race)
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
