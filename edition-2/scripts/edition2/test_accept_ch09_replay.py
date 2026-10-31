"""Checker controls; canned output is not a validated runtime reference."""
import copy
import json
import unittest

from accept_ch09 import MODELS, sha
from accept_ch09_replay import fixture, forgeries, rendered_ok


class ReplayOracle(unittest.TestCase):
    def output(self,vendor):
        texts=['Before editing.','[skill edit activation 2]\nFirst frozen manual.\n\n[/skill]',
               '[skill edit activation 3]\nFirst frozen manual.\n\n[/skill]','After editing.']
        if vendor=='openai':
            body={'messages':[{'role':'system','content':'Replay identity.'}]+[{'role':'user','content':text} for text in texts]}
        elif vendor=='anthropic':
            body={'system':'Replay identity.','messages':[{'role':'user','content':[{'type':'text','text':text} for text in texts]}]}
        else:
            body={'systemInstruction':{'parts':[{'text':'Replay identity.'}]},'contents':[{'role':'user','parts':[{'text':text} for text in texts]}]}
        return dict(exit=0,stdout=json.dumps(body))

    def test_rendering_oracle_controls(self):
        for vendor in MODELS:
            good=self.output(vendor)
            with self.subTest(vendor=vendor):self.assertTrue(rendered_ok(good,vendor))
            for old,new in [('First frozen manual.\\n\\n[/skill]','First frozen manual.\\n[/skill]'),
                            ('activation 3','activation 2'),('Replay identity.','Different system'),
                            ('After editing.','Before editing.')]:
                mutant=copy.deepcopy(good);self.assertIn(old,mutant['stdout']);mutant['stdout']=mutant['stdout'].replace(old,new)
                with self.subTest(vendor=vendor,mutation=old):self.assertFalse(rendered_ok(mutant,vendor))

    def test_fixture_identity_and_distinct_forgery_controls(self):
        original=fixture();variants=forgeries()
        # A valid parent must finish the first human turn before accepting another.
        kinds=[event['type'] for event in original]
        self.assertEqual(kinds[:3],['skills_initialized','message_received','response_ended'])
        self.assertEqual(kinds[-1],'message_received')
        self.assertEqual(len(variants),34)
        serialized=[json.dumps(value,sort_keys=True) for value in variants.values()]
        self.assertEqual(len(set(serialized)),len(serialized))
        self.assertNotIn(json.dumps(original,sort_keys=True),serialized)
        for events in [original]+list(variants.values()):
            self.assertEqual([e['seq'] for e in events],list(range(1,len(events)+1)))
        for e in original:
            for record in e.get('skills',{}).get('activated',[]):
                self.assertEqual(record['sha256'],sha(record['body'].encode()))


if __name__=='__main__':unittest.main()
