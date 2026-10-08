#!/usr/bin/env python3
"""Retain the existing thinking deletion with an emit-qualified source anchor."""
import hashlib
import json
from pathlib import Path
import sys

checker = Path(sys.argv[1]).resolve(strict=True)
source = Path(sys.argv[2]).resolve(strict=True)
sys.path.insert(0, str(checker))
import audit_ch06_mutations as audit

selected = [row for row in audit.MUTATIONS if row[0] == 'suppress-thinking-display']
assert len(selected) == 1
name, relative, before, after, pattern, package, refusal = selected[0]
assert before == 'if text == "" {' and after == 'if text == "" || channel == "thinking" {'
owner = 'func (s *streamParser) emit(ctx context.Context, id int, channel, text string) error {\n\t'
audit.MUTATIONS = [(name, relative, owner + before, owner + after, pattern, package, refusal)]
result = audit.evaluate(source, [name])
result['adapter_sha256'] = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
result['scope'] = __doc__
print(json.dumps(result, indent=2))
raise SystemExit(not result['passed'])
