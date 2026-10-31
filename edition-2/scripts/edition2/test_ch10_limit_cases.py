#!/usr/bin/env python3
"""Generator-only controls: skeletal oracle data is not a valid runtime snapshot."""
import base64
import importlib.util
from pathlib import Path
import unittest
from accept_ch10 import canonical,digest,encoded_json,envelope,parse,sample
spec=importlib.util.spec_from_file_location('limit_cases',Path(__file__).with_name('ch10-limit-cases.py'))
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)


class LimitCases(unittest.TestCase):
    def test_hash_repaired_ordering_mutations_have_genuine_parent_prerequisites(self):
        value=sample();value['as_of']=value['high_watermarks']['event']=110
        value['state']={'context':{'Calls':{'setter':{'Returned':True,'CalledAt':5,'ReturnedAt':7},'next':{'Returned':True,'CalledAt':13,'ReturnedAt':14}},'LimitFacts':[{'seq':6,'kind':'set','call_id':'setter','name':'','overrides':{'max_output_bytes':80,'ai_callback_pattern':''}},{'seq':12,'kind':'consumed','call_id':'next','name':'not_installed','overrides':{'max_output_bytes':80,'ai_callback_pattern':''}}],'Guidance':[{'seq':2,'consumed_at':3}],'RequestSeqs':[3,10]},'limits':None,'window':{'events':[{'event':{'seq':100}},{'event':{'seq':110}}]}}
        value['state_sha256']=digest(canonical(value['state']).encode());raw=encoded_json(value).encode();envelope(raw)
        rows=module.cases(raw)['cases'];self.assertEqual(len(rows),11)
        self.assertEqual(len({x['name'] for x in rows}),11)
        for row in rows:
            with self.subTest(case=row['name']):
                changed=envelope(base64.b64decode(row['bytes']));self.assertEqual(changed['as_of'],'110');self.assertNotEqual(changed['state_sha256'],value['state_sha256']);self.assertEqual(row['expected_code'],'session_corrupt')
        self.assertEqual(parse(raw)['state']['context']['Calls']['setter']['CalledAt'],'5')
        value['state']['window']['events'][0]['event']['seq']=7
        value['state_sha256']=digest(canonical(value['state']).encode())
        with self.assertRaisesRegex(AssertionError,'old calls remain in window'):
            module.cases(encoded_json(value).encode())


if __name__=='__main__':unittest.main()
