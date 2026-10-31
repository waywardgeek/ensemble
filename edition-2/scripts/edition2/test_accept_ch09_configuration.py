"""Canned verifier controls only; actual runtime positives remain independent."""
import copy
import json
import unittest
from accept_ch09 import MODELS
from accept_ch09_configuration import PRIMARY, OVERRIDE, CASES, evaluate


def canned(vendor,skills,setting):
    if skills and setting in ('override','equal'):return dict(exit=1,stderr='explicit System conflict',stdout='',requests=[],events=[],log_text='',replay={})
    system=PRIMARY if skills else OVERRIDE if setting=='override' else 'Inherited default'
    if vendor=='anthropic':body={'system':system}
    elif vendor=='openai':body={'messages':[{'role':'system','content':system}]}
    else:body={'systemInstruction':{'parts':[{'text':system}]}}
    return dict(exit=0,requests=[body],events=[{'type':'request_sent','request':{'configuration':{'system':'' if skills else system}}}],log_text=json.dumps(PRIMARY) if skills else '',replay={'empty':{'exit':0,'stdout':json.dumps(body),'stderr':''},'conflict':{'exit':1,'stdout':'','stderr':'explicit System conflict'}})


class ConfigurationOracle(unittest.TestCase):
    def test_canned_positive_matrix_and_capture_mutations(self):
        for vendor in MODELS:
            for skills,setting in CASES:self.assertTrue(evaluate(vendor,skills,setting,canned(vendor,skills,setting),'Inherited default'))
            for mutation in ('omitted','repeated','duplicate-body','ignored-offline-conflict'):
                value=copy.deepcopy(canned(vendor,True,'absent'))
                if mutation=='omitted':value['events'][0]['request']['configuration'].pop('system')
                elif mutation=='repeated':value['events'][0]['request']['configuration']['system']=PRIMARY
                elif mutation=='duplicate-body':value['log_text']+=json.dumps(PRIMARY)
                else:value['replay']['conflict']['exit']=0
                self.assertFalse(evaluate(vendor,True,'absent',value,'Inherited default'),(vendor,mutation))


if __name__=='__main__':unittest.main()
