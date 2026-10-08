#!/usr/bin/env python3
"""Targeted Chapter 7 implementation deletion controls, on disposable source only.

Every mutation first requires its exact probe's passing control, then its intended
failing leaf set and message. This audit is incremental, not full chapter acceptance.
"""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import tempfile
from accept_ch05 import command
from accept_ch07_public import install as install_public
from accept_ch07_components import install as install_components

HERE = Path(__file__).resolve().parent
MUTANTS = {
 'idle-watch-retention': ('internal/llm/watch.go', 'a.watches[len(a.watches)-1] = nil', '// deleted: release unused recipient slot', '.', '^TestC7PublicIdleWatchRelease$', True,
                         {'TestC7PublicIdleWatchRelease'}, ['idle actor retained closed watch']),
 'watch-count-bound': ('internal/llm/watch.go', 'len(w.items) >= common.WatchItems || size > common.WatchBytes-w.bytes', 'size > common.WatchBytes-w.bytes', '.', '^TestC7PublicWatchOverflowAndClose$', True,
                       {'TestC7PublicWatchOverflowAndClose'}, ['257th observation did not close']),
 'watch-byte-bound': ('internal/llm/watch.go', 'len(w.items) >= common.WatchItems || size > common.WatchBytes-w.bytes', 'len(w.items) >= common.WatchItems', './internal/llm', '^TestC7WatchEncodedByteCapacity$', False,
                      {'TestC7WatchEncodedByteCapacity/exact-drain-over','TestC7WatchEncodedByteCapacity/single-over'}, ['one encoded byte over aggregate capacity accepted','single over-limit payload accepted']),
 'projection-invalid-cut': ('internal/llm/watch.go', 'if a.projectionInvalid {', 'if false && a.projectionInvalid {', './internal/llm', '^TestC7ProjectionOverflowInvalidatesWatch$', False,
                            {'TestC7ProjectionOverflowInvalidatesWatch'}, ['truncated active projection claimed complete catch-up']),
 'pause-publication': ('internal/llm/watch.go', 'a.publish(common.Observation{Kind: "pause_changed", AgentID: a.parent.ID(), Paused: after.Paused, TypingClients: after.TypingClients, SpeakingClients: after.SpeakingClients})', '// deleted: publish applied pause', './internal/llm', '^TestC7PausePublicationBeforeAck$', True,
                       {'TestC7PausePublicationBeforeAck'}, ['pause never reached publication']),
 'pause-admission': ('internal/llm/actor.go', 'if a.pauseState().Paused {', 'if false && a.pauseState().Paused {', '.', '^TestC7PublicPauseAdmission$', True,
                     {'TestC7PublicPauseAdmission/release','TestC7PublicPauseAdmission/interrupt'}, ['paused call admitted before cause clear']),
 'selected-closed-recipient': ('internal/llm/watch.go', 'func (w *watch) push(record common.WatchRecord, size int) {\n\tw.mu.Lock()\n\tdefer w.mu.Unlock()\n\tif w.reason != "" {\n\t\treturn\n\t}', 'func (w *watch) push(record common.WatchRecord, size int) {\n\tw.mu.Lock()\n\tdefer w.mu.Unlock()', './internal/llm', '^TestC7SelectedRecipientClose$', True,
                               {'TestC7SelectedRecipientClose'}, ['late selected sender resurrected closed queue']),
 'opaque-display-leak': ('gui/projection.go', 'if p.Type == "opaque" {', 'if false && p.Type == "opaque" {', 'gui', '^TestC7GUIProjection$', True,
                         {'TestC7GUIProjection'}, ['typed display projection changed meaning']),
 'false-result-flag': ('gui/projection.go', 'out["tool"].(map[string]any)["is_error"] = e.Tool.IsError', '// deleted: preserve false result flag', 'gui', '^TestC7GUIProjection$', True,
                       {'TestC7GUIProjection/tool-result-explicit-false'}, ['typed display projection changed meaning']),
}


def probe(root, package, pattern, race):
    cwd = root/'gui' if package == 'gui' else root
    args = ['go','test','-json','-count=1','-timeout=90s','-run',pattern]
    if race:
        args.append('-race')
    args.append('.' if package == 'gui' else package)
    result = command(args,cwd,timeout=120)
    records = [json.loads(line) for line in result['stdout'].splitlines() if line.startswith('{')]
    failed = {r['Test'] for r in records if r.get('Action') == 'fail' and 'Test' in r}
    leaf = {test for test in failed if not any(other.startswith(test+'/') for other in failed)}
    ran = {r['Test'] for r in records if r.get('Action') == 'run' and 'Test' in r}
    return result, leaf, ran


def evaluate(source, only=None):
    rows=[]
    with tempfile.TemporaryDirectory(prefix='ch07-deletion-audit-') as directory:
        root=Path(directory)/'source'
        shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence','node_modules','__pycache__'))
        hashes={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
        install_public(root);install_components(root)
        for name in only or MUTANTS:
            file,old,new,package,pattern,race,expected,reasons=MUTANTS[name]
            target=root/file;original=target.read_text()
            if original.count(old)!=1:
                rows.append(dict(id=name,passed=False,error='mutation anchor absent or ambiguous'));continue
            positive,leaves,ran=probe(root,package,pattern,race)
            if positive['exit'] or leaves or not ran:
                rows.append(dict(id=name,passed=False,error='positive control failed',positive=positive));continue
            target.write_text(original.replace(old,new))
            try:
                negative,leaves,ran=probe(root,package,pattern,race)
            finally:
                target.write_text(original)
            passed=(negative['exit']!=0 and leaves==expected and all(reason in negative['stdout'] for reason in reasons)
                    and 'DATA RACE' not in negative['stdout']+negative['stderr'])
            rows.append(dict(id=name,passed=passed,file=file,old=old,new=new,positive=positive,negative=negative,
                             expected_failures=sorted(expected),actual_failures=sorted(leaves),required_diagnostics=reasons))
    return dict(scope=__doc__,source=str(source),source_files=hashes,
                checker_files={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in
                               [Path(__file__),HERE/'ch07_public_test.go.txt',HERE/'ch07_internal_test.go.txt',HERE/'ch07_gui_test.go.txt',HERE/'ch06_public_test.go.txt',HERE/'ch06_internal_test.go.txt']},
                passed=all(r['passed'] for r in rows),checks=rows)


def main():
    p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=Path);p.add_argument('--only',action='append',choices=MUTANTS);a=p.parse_args()
    result=evaluate(a.source.resolve(strict=True),a.only);print(json.dumps(result,indent=2));return 0 if result['passed'] else 1


if __name__=='__main__':raise SystemExit(main())
