"""Canned oracle controls only; no implementation/reference acceptance claim."""
import copy
import json
import unittest

from accept_ch09 import MODELS, sha
from accept_ch09_graph import STATES, evaluate


def fixture(vendor):
    def material(name,identity,body):return {'name':name,'activation':identity,'body':body,'sha256':sha(body.encode())}
    names=[['load_skill','read_file','unload_skill']]
    events=[{'type':'skills_initialized','skills':{'activated':[material('base',1,'Primary remains fixed.\n')]}}]
    tools='- list_directory\n- load_skill\n- read_file\n- unload_skill\n- write_file'
    offers=['broken','cycle','review','search','unavailable','variable']
    available=[{'name':x,'description':x.title()+' scratch files'} for x in offers]
    offer_text='\n'.join('- '+x['name']+': '+x['description'] for x in available)
    activations={1:[material('read',2,'Read dependency.\n'+tools+'\n'+offer_text+'\n'),material('edit',3,'Edit material.\n'+tools+'\n'+offer_text+'\n')],
                 2:[material('review',4,'Review material.\n')],3:[material('search',5,'Search material.\n')],5:[material('edit',6,'Reload material.\n')]}
    for index,(roots,active,retired) in enumerate(STATES,1):
        grants={'load_skill','read_file','unload_skill'}
        if 'read' in active:grants.add('list_directory')
        if 'edit' in active:grants.add('write_file')
        if 'search' in active:grants.add('search_files')
        names.append(sorted(grants))
        state={'revision':index,'primary':'base','roots':roots,'active':[{'name':n,'activation':a} for n,a in sorted(active.items())],
               'retired':[{'activation':x} for x in retired],'available':available if index==1 else []}
        events.append({'type':'skills_changed','skills':{'state':state,'activated':activations.get(index,[])}})
    names+= [names[-1],names[-1]]
    for name,code in [('broken','skill_dependency'),('cycle','skill_dependency'),('unavailable','skill_tool_unavailable'),('variable','skill_variable'),('hidden','skill_unavailable'),('primary','skill_unavailable'),('dependency','skill_unavailable')]:
        ack={'error':code,'name':{'primary':'base','dependency':'read'}.get(name,name),'revision':8}
        events.append({'type':'tool_returned','tool':{'call_id':name,'is_error':True,'parts':[{'text':json.dumps(ack)}]}})
    for name,revision in [('already-search',7),('noop',8)]:
        ack={'status':'unchanged','name':'search','revision':revision,'changed':False}
        events.append({'type':'tool_returned','tool':{'call_id':name,'parts':[{'text':json.dumps(ack)}]}})
    events.append({'type':'tool_returned','tool':{'call_id':'write-refused','is_error':True}})
    text='\n'.join(f'[skill {name} activation {identity}]' for name,identity in [('read',2),('edit',3),('review',4),('search',5),('edit',6)])
    requests=[]
    for i,grants in enumerate(names):
        content=text if i==10 else ''
        if vendor=='anthropic':request={'tools':[{'name':n} for n in grants],'messages':[{'role':'user','content':[{'text':content}]}]}
        elif vendor=='openai':request={'tools':[{'function':{'name':n}} for n in grants],'messages':[{'role':'user','content':content}]}
        else:request={'tools':[{'functionDeclarations':[{'name':n} for n in grants]}],'contents':[{'role':'user','parts':[{'text':content}]}]}
        requests.append(request)
    return {'events':events,'requests':requests,'files':{},'vendor':vendor}


class GraphOracle(unittest.TestCase):
    def test_positives(self):
        for vendor in MODELS:
            result=evaluate(**fixture(vendor));self.assertTrue(all(result.values()),(vendor,result))

    def test_intended_mutations(self):
        for vendor in MODELS:
            for mutation,intended in [('file','revocation-real-effect'),('hash','fresh-material-identities'),('root','exact-closure-roots-retirement'),('manual','retired-manual-once'),('offer','dependency-offers')]:
                source=fixture(vendor);bad=copy.deepcopy(source)
                if mutation=='file':bad['files']['forbidden.txt']='wrong'
                if mutation=='hash':bad['events'][0]['skills']['activated'][0]['sha256']='0'*64
                if mutation=='root':bad['events'][4]['skills']['state']['roots']=[]
                if mutation=='offer':bad['events'][1]['skills']['state']['available']=[]
                if mutation=='manual':
                    bad['requests'][-1]=json.loads(json.dumps(bad['requests'][-1]).replace('[skill edit activation 3]','removed'))
                failures=sorted(k for k,v in evaluate(**bad).items() if not v)
                self.assertEqual(failures,[intended],(vendor,mutation,failures))


if __name__=='__main__':unittest.main()
