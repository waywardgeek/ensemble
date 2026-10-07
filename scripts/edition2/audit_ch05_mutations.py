#!/usr/bin/env python3
"""Deletion controls for the Edition 2 actor checker; never edit input source.

Each mutation must compile, fail its intended category, and lose that category's
points. A positive control runs the same selected categories first. Anchors are
adapter-specific; an unmatched anchor is an invalid audit, never a pass.
"""
import argparse
import json
import pathlib
import re
import shutil
import tempfile
import sys
import accept_ch05 as checker

# These delete independent properties, not public spelling. Fixture signatures
# remain unchanged, so build failures cannot masquerade as behavioral detection.
MUTATIONS = [
 ('completion-before-durability','responsiveness','ensemble.go','err = a.log.Append(owned)','if owned.Type != "turn_ended" { err = a.log.Append(owned) }'),
 ('stale-operation-accepted','responsiveness','internal/llm/actor.go','case "model":\n\t\tif a.active == nil || m.Operation != a.operation {','case "model":\n\t\tif a.active == nil {'),
 ('refusal-detail-lost','loud-refusal','internal/llm/render.go','r.fail(fmt.Sprintf("unsupported reference: model %q has no %s %s mapping for %s", r.model, r.target.Surface, kind, media))','_ = kind; _ = media; r.fail(fmt.Sprint("unsupported reference"))'),
 ('report-preparation-consumes','responsiveness','internal/jobs/report.go','snapshot := j.snapshot','j.cursor = j.snapshot.Bytes\n\tsnapshot := j.snapshot'),
 ('completion-shared-answer','responsiveness','internal/llm/actor.go','Text: TextAnswer(a.parent.Engine(), r.parts)','Text: "WRONG-SHARED-ANSWER"'),
 ('completion-consumed','responsiveness','internal/llm/actor.go','c, err := Clone(r.parent.Agent().Engine(), r.completion)','c, err := Clone(r.parent.Agent().Engine(), r.completion)\n        r.completion.Text = "CONSUMED"'),
 ('queued-cancel-ignored','responsiveness','internal/llm/actor.go','func (r *request) Cancel() error          { return r.parent.Cancel(r.id) }','func (r *request) Cancel() error          { return nil }'),
 ('hint-lost-on-wire','replay','internal/llm/render.go','entries = append(entries, c.Hints...)','// mutation: omit pending hints from render'),
 ('replay-current-config','replay','ensemble.go','return llm.Render(a.engine, state, config)','config.System = a.Config().System\n            return llm.Render(a.engine, state, config)'),
 ('observer-final-omitted','observers','internal/llm/actor.go','Kind: "part_final"','Kind: "deleted_part_final"'),
 ('observer-overflow-hidden','observers','ensemble.go','s.reason = "overflow"','s.reason = ""'),
 ('observer-copy-removed','observers','ensemble.go','case s.queue <- owned:','case s.queue <- func() Observation { _ = owned; return o }():'),
 ('collection-first-only','completion-collection','collection.go','out = append(out, v)','out = append(out, v)\n                c.returned[i] = true\n                return out, false, nil'),
 ('collection-repeat','completion-collection','collection.go','c.returned[i] = true','// mutation: completion not marked returned'),
 ('round-bound-removed','parity','internal/llm/actor.go','r.rounds == 16','r.rounds == 17'),
 ('global-session-state','architecture','independent_mutable.go','','package ensemble\nvar auditSessionCounter uint64\n'),
 ('new-spoke-sideways','architecture','internal/independent_spoke/sideways.go','','package independent_spoke\nimport _ "example.com/ensemble/internal/llm"\n'),
 ('actor-wrong-parent','architecture','internal/llm/actor.go','func (a *Actor) Agent() common.ActorAgent { return a.parent }','func (a *Actor) Agent() common.ActorAgent { return nil }'),
]


def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('source',type=pathlib.Path);p.add_argument('--only',action='append');a=p.parse_args();source=a.source.resolve(strict=True)
 selected=[m for m in MUTATIONS if not a.only or m[0] in a.only]
 categories=sorted(set(m[1] for m in selected));positive=checker.evaluate(source,categories,True,False)
 with tempfile.TemporaryDirectory(prefix='ch05-renamed-positive-') as tmp:
  renamed=pathlib.Path(tmp)/'source';shutil.copytree(source,renamed,ignore=shutil.ignore_patterns('.git','evidence'))
  path=renamed/'internal/llm/actor.go';original=path.read_text()
  # All bound uses in this file move together. The complete original file is
  # the unique replacement anchor; no external API or property changes.
  updated=re.sub(r'\bparent\b','origin',original)
  if updated==original:raise ValueError('rename positive control did not apply')
  assert path.read_text().count(original)==1
  path.write_text(updated)
  renamed_positive=checker.evaluate(renamed,categories,True,False)
 records=[]
 for name,category,relative,before,after in selected:
  print(name,file=sys.stderr,flush=True)
  with tempfile.TemporaryDirectory(prefix='ch05-mutation-') as tmp:
   root=pathlib.Path(tmp)/'source';shutil.copytree(source,root,ignore=shutil.ignore_patterns('.git','evidence'))
   path=root/relative
   if before:
    text=path.read_text()
    if text.count(before)!=1:records.append(dict(id=name,invalid=True,reason='anchor count '+str(text.count(before))));continue
    path.write_text(text.replace(before,after))
   else:
    if path.exists():raise ValueError('negative fixture already exists')
    path.parent.mkdir(parents=True,exist_ok=True);path.write_text(after)
   checker.install(root)
   built=checker.command(['go','test','-run','^$','./...'],root)
   if built['exit']:records.append(dict(id=name,invalid=True,build=built));continue
   result=checker.evaluate(root,[category],True,False)
   records.append(dict(id=name,category=category,invalid=False,detected=not result['passed'],score_loss=result['maximum']-result['score'],result=result))
 passed=positive['passed'] and renamed_positive['passed'] and all(not r['invalid'] and r['detected'] for r in records) and len(records)==len(selected)
 print(json.dumps(dict(passed=passed,positive=positive,renamed_positive=renamed_positive,mutations=records),indent=2));return 0 if passed else 1

if __name__=='__main__':raise SystemExit(main())
