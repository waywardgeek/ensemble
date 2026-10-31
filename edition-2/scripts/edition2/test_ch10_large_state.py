#!/usr/bin/env python3
"""Identity gate controls from a complete valid large-state receipt; never invokes Go."""
import copy
import unittest

import accept_ch10_large_mutations as audit

class PositiveIdentity(unittest.TestCase):
    def setUp(self):
        paths=['accept_ch10_large_state.py','accept_ch10_lifecycle.py',*audit.FIXTURES]
        self.positive=dict(complete=True,passed=True,source_commit='bound',source_files={'source.go':'hash'},
                          checker_files={name:audit.digest(audit.HERE/name) for name in paths},
                          checks=[dict(passed=True,command=['binary','-test.run='+p]) for p in audit.CASES.values()])
        audit.validate_positive(self.positive,{'source.go':'hash'},'bound')

    def test_identity_and_coverage_refusals_from_valid_parent(self):
        edits=[lambda p:p.update(complete=False),lambda p:p.update(passed=False),
               lambda p:p.update(source_commit='other'),lambda p:p.update(source_files={}),
               lambda p:p.update(checker_files={}),
               lambda p:p['checker_files'].update({'unexpected.py':'unbound'}),
               lambda p:p['checker_files'].update({'accept_ch10_large_state.py':'wronghash'}),
               lambda p:p.update(checks=p['checks'][:-1])]
        for edit in edits:
            p=copy.deepcopy(self.positive);edit(p)
            with self.subTest(candidate=p):
                with self.assertRaises(ValueError):
                    audit.validate_positive(p,{'source.go':'hash'},'bound')

if __name__=='__main__': unittest.main()
