#!/usr/bin/env python3
"""Targeted implementation deletions with unchanged passing positive fixtures.

Only disposable copies are mutated. A compiler/setup failure cannot count as
an intended behavioral refusal. Scope grows with independently protected checks.
"""
import argparse
import hashlib
import json
import pathlib
import shutil
import tempfile

from accept_ch06 import HERE, install
from accept_ch05 import command

MUTATIONS = [
    ('drop-refusal-replay', 'internal/llm/render.go', 'm["refusal"] = refusal', '_ = refusal',
     '^TestC6TypedCallsAndOpaque$/openai', '.', 'recognized refusal missing from continuation'),
    ('drop-signed-parts', 'internal/llm/parse.go', 'part.Opaque = append(json.RawMessage(nil), signature...)', 'part.Opaque = rawValue("deleted-signature")',
     '^TestC6TypedCallsAndOpaque$/gemini', '.', 'signed part retention changed'),
    ('suppress-thinking-display', 'internal/llm/stream.go', 'if text == "" {', 'if text == "" || channel == "thinking" {',
     '^TestC6TypedCallsAndOpaque$/anthropic', '.', 'thinking fragments missing or entered wrong part'),
    ('disable-operation-deadline', 'internal/llm/operation.go', 'if o.timeout > 0 {', 'if false {',
     '^TestC6HTTPTimeoutIncludesCapacityWait$', './internal/llm', 'configured HTTP timeout did not release'),
    ('sever-operation-parent', 'internal/llm/operation.go', 'return o.parent }', 'return nil }',
     '^TestC6ExactFraming$', './internal/llm', 'operation diagnostic owner chain broken'),
    ('hide-fragments', 'internal/llm/actor.go',
     'a.observeModel(common.Observation{Kind: "part_delta", PartID: f.PartID, Channel: f.Channel, Text: f.Text})',
     '_ = f', '^TestC6PublicEarlyIdentity$/openai', '.', 'public observation barrier timed out'),
    ('wrong-final-identity', 'internal/llm/actor.go', 'PartID: partID, Part: &copy', 'PartID: partID + 100, Part: &copy',
     '^TestC6PublicEarlyIdentity$/openai', '.', 'part mapping differs'),
    ('frame-overflow', 'internal/llm/sse.go', 'frame > pendingLimit', 'frame > pendingLimit + 1',
     '^TestC6ExactFraming$', './internal/llm', 'one-byte physical overflow accepted'),
    ('dispatch-unfinished-frame', 'internal/llm/sse.go', 'return sseEvent{}, s.fail("unfinished SSE frame")',
     'return sseEvent{name, strings.Join(data, "\\n")}, nil', '^TestC6ExactFraming$', './internal/llm', 'invalid/unfinished SSE accepted'),
    ('unbounded-pending-growth', 'internal/llm/operation.go', 'const pendingLimit = 1 << 20', 'const pendingLimit = 2 << 20',
     '^TestC6CapacityCancellationAndUTF8$/cancel', './internal/llm', 'pending store exceeded one MiB'),
    ('unfair-drain', 'internal/llm/operation.go', 'const drainLimit = 64 << 10', 'const drainLimit = 128 << 10',
     '^TestC6CapacityCancellationAndUTF8$', './internal/llm', 'actor drain exceeded 64 KiB fairness budget'),
    ('completion-before-drain', 'internal/llm/operation.go', 'empty := o.bytes == 0', 'empty := true',
     '^TestC6CompletionDrainWait$', './internal/llm', 'completed response overtook pending fragments'),
    ('ignore-recorded-delivery', 'ensemble.go', 'DisableStreaming: request.Delivery != "stream"', 'DisableStreaming: a.Config().DisableStreaming',
     '^TestC6RecordedDeliveryReplay$/openaifalse', '.', 'reconstruction ignored recorded delivery'),
    ('erase-overflow-reason', 'ensemble.go', 's.reason = "overflow"', 's.reason = "lost"',
     '^TestC6PublicObserverOverflow$', '.', 'blocked observer did not report overflow'),
    ('assembled-overflow', 'internal/llm/operation.go', 'const responseLimit = 16 << 20', 'const responseLimit = 17 << 20',
     '^TestC6AssembledBound$/openai', './internal/llm', 'assembled limit overflow accepted'),
    ('accept-incomplete-response', 'internal/llm/stream.go', 'if !s.finished {', 'if false {',
     '^TestC6RequiredTerminal$/openai', './internal/llm', 'missing API terminal accepted'),
]


def test(root, pattern, package):
    return command(['go', 'test', '-json', '-run', pattern, '-count=1', '-timeout=30s', package], root, timeout=60)


def output(result):
    return ''.join(json.loads(line).get('Output', '') for line in result['stdout'].splitlines())


def evaluate(source, selected=None):
    rows = []
    with tempfile.TemporaryDirectory(prefix='ch06-deletion-audit-') as tmp:
        root = pathlib.Path(tmp) / 'source'
        shutil.copytree(source, root, ignore=shutil.ignore_patterns('.git', 'evidence'))
        hashes = {str(p.relative_to(root)): hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in sorted(root.rglob('*')) if p.is_file() and (p.suffix == '.go' or p.name in ('go.mod', 'go.sum'))}
        install(root)
        for name, relative, before, after, pattern, package, refusal in MUTATIONS:
            if selected and name not in selected:
                continue
            positive = test(root, pattern, package)
            if positive['exit']:
                rows.append(dict(id=name, passed=False, reason='positive fixture failed; mutant not credited', positive=positive))
                continue
            path = root / relative
            original = path.read_text()
            if original.count(before) != 1:
                rows.append(dict(id=name, passed=False, reason='mutation adapter anchor not unique', positive=positive))
                continue
            try:
                path.write_text(original.replace(before, after, 1))
                negative = test(root, pattern, package)
                text = output(negative)
                passed = negative['exit'] == 1 and refusal in text and '[build failed]' not in text
                rows.append(dict(id=name, passed=passed, intended_refusal=refusal, positive=positive, negative=negative))
            finally:
                path.write_text(original)
    return dict(scope='targeted implementation deletions, no chapter score', source_files=hashes,
                checker_files={x.name: hashlib.sha256(x.read_bytes()).hexdigest() for x in [pathlib.Path(__file__), HERE/'accept_ch06.py', HERE/'ch06_public_test.go.txt', HERE/'ch06_internal_test.go.txt']}, passed=bool(rows) and all(row['passed'] for row in rows), mutations=rows)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('source', type=pathlib.Path)
    p.add_argument('--only', choices=[x[0] for x in MUTATIONS], action='append')
    a = p.parse_args()
    result = evaluate(a.source.resolve(strict=True), a.only)
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
