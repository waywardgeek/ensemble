#!/usr/bin/env python3
"""Adapt retained fixture owners and the published Chapter 10 duplicate-name rule.

Only the disposable bundle changes. Its reader fixture becomes an external
composition test so it can construct the real new codec through its actual
parent. Production spokes and all historical fixture files remain untouched.
"""
import hashlib
from pathlib import Path
import subprocess

from ch09_retained_adapters import prepare as previous


FINAL_PART_HELPER = """// Compare exact typed facts while ignoring only an unused empty collection.
func ch10EqualFinalPart(a, b ens.Part) bool {
\tif a.Type != "tool_result" && len(a.Parts) == 0 {
\t\ta.Parts = nil
\t}
\tif b.Type != "tool_result" && len(b.Parts) == 0 {
\t\tb.Parts = nil
\t}
\treturn reflect.DeepEqual(a, b)
}

"""


def prepare(destination, source=None):
    result = previous(destination, source)

    def edit(name, old, new, count=1):
        path = destination / name
        text = path.read_text()
        assert text.count(old) == count, f'{name}: changed fixture anchor {old!r}'
        path.write_text(text.replace(old, new))
        result['changes'].append(dict(file=name, old=old, new=new, count=count))
        result['adapted_files'][name] = hashlib.sha256(path.read_bytes()).hexdigest()

    # dd1111e §10.3 explicitly gives repeated name members no unique name.
    # Keep the prior checker immutable and change only this expected field in
    # its Chapter 10 copy. Its evaluate() assertions and all other cases stay.
    edit('accept_ch09_management.py',
         "('duplicate','{\"name\":\"edit\",\"name\":\"edit\"}','edit')",
         "('duplicate','{\"name\":\"edit\",\"name\":\"edit\"}','')")
    # The adapted direct script must import the original helpers: accept_ch09
    # derives ROOT from __file__. Its prepared copy would point outside the repo.
    edit('accept_ch09_management.py', 'from accept_ch09 import ',
         'import sys\nsys.path.insert(0, ' + repr(str(Path(__file__).resolve().parent)) +
         ')\nfrom accept_ch09 import ')

    # Ch6 compares owned neutral Parts, not nilness of an omitted unused
    # collection. Ch10's snapshot clone makes that empty collection explicit.
    # Normalize only copied empty Parts on variants where it is unused; every
    # other field (including exact raw bytes) retains DeepEqual comparison.
    edit('ch06_public_test.go.txt',
         'reflect.DeepEqual(event.Response.Parts[facts[2].PartIndex], *facts[2].Part)',
         'ch10EqualFinalPart(event.Response.Parts[facts[2].PartIndex], *facts[2].Part)')
    edit('ch06_public_test.go.txt',
         'reflect.DeepEqual(*x.Part, response.Parts[i])',
         'ch10EqualFinalPart(*x.Part, response.Parts[i])')
    edit('ch06_public_test.go.txt', 'const c6Model =', FINAL_PART_HELPER + 'const c6Model =')

    reader = 'ch09-raw-record_test.go.txt'
    edit(reader, 'package eventlog\n', 'package eventlog_test\n')
    edit(reader, '"example.com/ensemble/internal/common"',
         '"example.com/ensemble/internal/common"\n'
         '\t"example.com/ensemble/internal/eventlog"\n'
         '\t"example.com/ensemble/internal/persistence"')
    edit(reader, 'func (a *rawReviewOwner) Ensemble() common.Ensemble { return a.root }',
         'func (a *rawReviewOwner) Ensemble() common.Ensemble { return a.root }\n'
         'func (a *rawReviewOwner) Codec() common.SessionCodec { return persistence.NewCodec(a) }')
    edit(reader, ':= Read(', ':= eventlog.Read(', count=2)

    # This existing controlled owner has neither a store nor storage faults.
    # Explicit answers replace newly used methods promoted from a nil interface.
    edit('ch05_stale_test.go.txt', 'func (a *auditLateAgent) ID() string',
         'func (a *auditLateAgent) AppendFailure() error { return nil }\n'
         'func (a *auditLateAgent) SessionState() *common.SessionState { return nil }\n'
         'func (a *auditLateAgent) ID() string')

    # Chapter 10 prepares session records before the shared persistence branch.
    # Keep the original standalone positives/assertions, but move the deletion
    # to the actual append boundary, covering either representation of a write.
    old_boundary = '\tif err == nil && persist {\n\t\terr = a.log.Append(owned)'
    new_boundary = '\tif err == nil && persist {\n\t\tif encoded != nil {'
    edit('ch09-review-boundaries.py', repr(old_boundary), repr(new_boundary), count=2)
    old_append = '\t\terr = a.log.Append(owned)'
    old_rollback = old_append + '\n\t\tif err != nil && owned.Type == "tool_returned" { a.skills = nil }'
    append_end = '\t\t}\n\t\tif err != nil {\n\t\t\ta.faulted = true'
    rollback_end = ('\t\t}\n\t\tif err != nil && owned.Type == "tool_returned" { a.skills = nil }'
                    '\n\t\tif err != nil {\n\t\t\ta.faulted = true')
    edit('ch09-review-boundaries.py', repr(old_append) + ',' + repr(old_rollback),
         repr(append_end) + ',' + repr(rollback_end))

    # An empty partial-argument delta added a second text guard. Qualify the
    # original mutation by its emit owner instead of weakening uniqueness or
    # accidentally mutating argument assembly. The target assertion is intact.
    emit = 'func (s *streamParser) emit(ctx context.Context, id int, channel, text string) error {\n\t'
    edit('audit_ch06_mutations.py',
         repr('if text == "" {') + ', ' + repr('if text == "" || channel == "thinking" {'),
         repr(emit + 'if text == "" {') + ', ' + repr(emit + 'if text == "" || channel == "thinking" {'))
    for name in (reader, 'ch05_stale_test.go.txt', 'ch06_public_test.go.txt'):
        path = destination / name
        before = hashlib.sha256(path.read_bytes()).hexdigest()
        subprocess.run(['gofmt', '-w', path], check=True)
        after = hashlib.sha256(path.read_bytes()).hexdigest()
        result['changes'].append(dict(file=name, formatting='gofmt', before=before, after=after))
        result['adapted_files'][name] = after
    result['chapter10_adapter_sha256'] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    return result
