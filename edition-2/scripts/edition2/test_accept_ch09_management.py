"""Canned controls for the argument/limit checker, not runtime acceptance."""
import copy
import json
import unittest
from accept_ch09 import MODELS
from accept_ch09_management import ATTEMPTS, MANUAL, SENTINEL, evaluate, wire


def canned(vendor):
    events=[];artifacts=[]
    for number,(identity,tool,_,ack) in enumerate(ATTEMPTS,1):
        events.append(dict(type='tool_called',tool=dict(call_id=identity,name=tool)))
        events.append(dict(type='tool_returned',tool=dict(call_id=identity,name=tool,is_error='error' in ack,parts=[dict(type='text',text='tool_limits consumed by '+tool+'\n'+json.dumps(ack,separators=(',',':')))])))
        locator='cr/io/'+str(number);artifacts.append(locator)
        events.extend([dict(type='tool_returned',tool=dict(call_id='setter-'+identity)),dict(type='tool_returned',tool=dict(call_id='read-'+identity,parts=[dict(type='text',text=SENTINEL)],job=dict(handle=number,output=dict(locator=locator))))])
    events.extend([dict(type='skills_changed',skills=dict(action=action,name='edit',state=dict(revision=revision))) for action,revision in [('load',1),('unload',2)]])
    text='[skill edit activation 2]\n'+MANUAL+'\n[/skill]'
    if vendor=='openai':body=dict(messages=[dict(role='user',content=text)])
    elif vendor=='anthropic':body=dict(messages=[dict(role='user',content=[dict(type='text',text=text)])])
    else:body=dict(contents=[dict(role='user',parts=[dict(text=text)])])
    return events,[{},body],artifacts


class ManagementOracle(unittest.TestCase):
    def test_canned_positives_and_targeted_mutations(self):
        for vendor in MODELS:
            args=canned(vendor);self.assertTrue(all(evaluate(*args,vendor).values()))
            for target in ['strict-ack','consumed-limit-note','ordinary-defaults','management-paired-without-jobs']:
                events,requests,artifacts=copy.deepcopy(args)
                if target=='strict-ack':events[1]['tool']['is_error']=False
                elif target=='consumed-limit-note':events[1]['tool']['parts'][0]['text']=events[1]['tool']['parts'][0]['text'].splitlines()[-1]
                elif target=='ordinary-defaults':events[3]['tool']['parts'][0]['text']='x'
                else:events[0]['tool']['job']={'handle':99}
                failed={name for name,passed in evaluate(events,requests,artifacts,vendor).items() if not passed}
                self.assertEqual(failed,{target if target=='management-paired-without-jobs' else 'missing/'+target})

    def test_duplicate_member_wire_preserved(self):
        for vendor in MODELS:
            data=wire(vendor).decode();pairs=json.loads(data,object_pairs_hook=lambda p:p)
            if vendor=='openai':
                # Arguments are themselves JSON strings on Chat Completions.
                arguments=[c['function']['arguments'] for c in json.loads(data)['choices'][0]['message']['tool_calls']]
                self.assertEqual(arguments.count('{"name":"edit","name":"edit"}'),1)
            else:
                def duplicate(value):
                    if not isinstance(value,list):return 0
                    return int(value==[('name','edit'),('name','edit')])+sum(duplicate(v[1] if isinstance(v,tuple) else v) for v in value)
                self.assertEqual(duplicate(pairs),1)


if __name__=='__main__':unittest.main()
