#!/usr/bin/env python3
"""Install independent probes into an explicitly disposable source copy.

Adapt only the typed service API changed by the reviewed responsibility split.
The behavioral assertions remain the same. Repository source is refused.
"""
import argparse
import pathlib
import subprocess


def adapt_job_probe(text, source):
    declarations = (source / "internal/common/types.go").read_text()
    if "Report(Job, JobReport)" not in declarations:
        return text
    anchor = 'func (r auditRegistry) Execute(p common.Part) common.ToolEvent { return *r.a.handler(p, nil) }'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 'func (r auditRegistry) Execute(p common.Part) common.ToolEvent { panic("unused ordinary adapter") }')
    assert text.count('*common.ToolEvent') == 5
    text = text.replace('*common.ToolEvent', '*common.ExecutionResult')
    anchor = '&common.ToolEvent{Parts: []common.Part{{Type: "text", Text: &text}}}'
    assert text.count(anchor) == 3
    text = text.replace(anchor, '&common.ExecutionResult{Text: text}')
    anchor = 's.Create(common.Part{Name: "run_command"})'
    assert text.count(anchor) == 1
    text = text.replace(anchor, 's.Create()')
    assert text.count('s.Report(j,') == 6
    text = text.replace('s.Report(j,', 'auditReport(s, j,')
    text += '''
// Adapter only: the reviewed revision passes typed report requests to Jobs.
func auditReport(s *Service, j common.Job, call common.Part, limits common.Limits, note string) error {
 return s.Report(j, common.JobReport{CallID: call.CallID, Limits: limits, Note: note,
  Original: call.Name != "wait_for_job", MatchStart: -1})
}
func (auditRegistry) ResolveLimits(common.Part) (common.Limits, string, error) { panic("unused wire adapter") }
func (auditRegistry) Supervise(common.Part, common.Limits, string) error { panic("unused wire adapter") }
'''
    return text


def install(source):
    repo = pathlib.Path(__file__).resolve().parents[2]
    if source == repo or repo in source.parents:
        raise ValueError("probes require a disposable source outside the repository")
    fixtures = [("ch04_durability_test.go.txt", "independent_durability_test.go"),
                ("ch04_jobs_fault_test.go.txt", "internal/jobs/independent_fault_test.go"),
                ("ch04_public_validation_test.go.txt", "independent_public_validation_test.go")]
    if "type ParsedResponse struct" in (source / "internal/common/types.go").read_text():
        fixtures.append(("ch04_revision_boundary_test.go.txt", "independent_revision_boundary_test.go"))
    for fixture, relative in fixtures:
        text = pathlib.Path(__file__).with_name(fixture).read_text()
        if fixture == "ch04_jobs_fault_test.go.txt":
            text = adapt_job_probe(text, source)
        target = source / relative
        target.write_text(text)
        subprocess.run(["gofmt", "-w", str(target)], check=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=pathlib.Path)
    install(parser.parse_args().source.resolve(strict=True))
