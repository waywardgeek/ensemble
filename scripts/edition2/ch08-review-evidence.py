#!/usr/bin/env python3
"""Independent frozen Chapter 8 live audit. Local replay only; never HTTP."""
import argparse
import copy
import hashlib
import importlib.util
import json
from pathlib import Path
import re
import struct
import subprocess
import sys
import tempfile

REPO = Path(__file__).resolve().parents[2]
HERE = REPO / 'solutions/edition-2/main/evidence/ch08'
PREFIX = str(HERE.relative_to(REPO)) + '/'
FREEZE = '7f517d8'
TEACHING_APPEND = None
BINDINGS = ('binding-bd5c05a.json', 'binding-cd9de3e.json', 'binding-a06d4f3.json', 'initial-binding.json')


def read(p): return json.loads(p.read_text())
def rows(p): return [json.loads(s) for s in p.read_text().splitlines()]
def digest(b): return hashlib.sha256(b).hexdigest()
def sha(p): return digest(p.read_bytes())
def git(*args): return subprocess.check_output(['git', *args], cwd=REPO)
def canonical(v): return json.dumps(v, sort_keys=True, separators=(',', ':'))


def index(revision):
    result = {}
    for line in git('ls-tree', '-r', revision, '--', PREFIX).decode().splitlines():
        meta, path = line.split('\t'); result[path[len(PREFIX):]] = meta.split()[2]
    return result


def bound_raw(revision, name, blob, raw):
    actual = hashlib.sha1(b'blob ' + str(len(raw)).encode() + b'\0' + raw).hexdigest()
    if actual != blob and name == 'student-review.md' and TEACHING_APPEND:
        # Only an explicitly identified append may advance the teaching record.
        # Every raw receipt, binding and verifier remains exact at FREEZE.
        bound_append = git('show', TEACHING_APPEND + ':' + PREFIX + name)
        original = git('show', revision + ':' + PREFIX + name)
        assert raw == bound_append and raw.startswith(original), 'unbound or rewritten teaching append'
        return original
    assert actual == blob, 'frozen original changed: ' + name
    return raw


def frozen_files(revision, names=None):
    result = {}
    for name, blob in index(revision).items():
        if names is not None and name.split('/')[0] not in names: continue
        result[name] = digest(bound_raw(revision, name, blob, (HERE/name).read_bytes()))
    return result


def load_helpers(binding, directory):
    directory.mkdir()
    for name in binding['support']:
        (directory / name).write_bytes(git('show', binding['source_revision'] + ':' + PREFIX + name))
    spec = importlib.util.spec_from_file_location('evidence', directory / 'evidence.py')
    helper = importlib.util.module_from_spec(spec); spec.loader.exec_module(helper)
    helper.ROOT = REPO  # Explicit repository root; historical support lives in a disposable copy.
    sys.modules['evidence'] = helper
    spec = importlib.util.spec_from_file_location('reviewed_verifier', directory / 'verify-receipts.py')
    verifier = importlib.util.module_from_spec(spec); spec.loader.exec_module(verifier)
    return helper, verifier


def launch_check(binding, launch):
    assert launch['source_revision'] == binding['source_revision'], 'launch source mismatch'
    assert launch['executables'] == {k:v['sha256'] for k,v in binding['executables'].items()}, 'launch executable mismatch'
    assert launch['support'] == binding['support'], 'launch support mismatch'
    assert launch['browser_tools'] == binding['browser_tools'], 'launch browser tools mismatch'
    assert launch['launched_executable'] in binding['executables'], 'unknown launch executable'


def native_audio(run, browser):
    if not browser: return []
    launch = next(x for x in browser if x['kind'] == 'launch')
    output = []
    for entry in [x for x in browser if x['kind'] == 'audio']:
        wav = run / Path(entry['path']).name
        assert entry['exit'] == 0 and entry['browserPID'] == launch['browser_pid']
        assert f"pid={launch['browser_pid']} " in entry['output']
        raw = subprocess.check_output(['ffmpeg','-v','error','-i',str(wav),'-f','f32le','-ac','1','-ar','48000','-'])
        values = struct.unpack('<' + 'f' * (len(raw)//4), raw)
        prefix = values[:int(1.5*48000)]; speech = values[int(2.5*48000):]
        rms = lambda xs: (sum(x*x for x in xs)/len(xs))**.5
        assert len(values) > 5*48000 and rms(prefix) < .00001 and rms(speech) > .001
        output.append({'path':str(wav.relative_to(HERE)), 'sha256':sha(wav), 'browser_pid':launch['browser_pid'],
                       'seconds':len(values)/48000, 'silent_prefix_rms':rms(prefix), 'speech_rms':rms(speech),
                       'scope':'PID-bound native audio amplitude; no transcription or listening claim'})
    return output


def feature_audit():
    results=[]
    for vendor in ('anthropic','openai','gemini'):
        cli=HERE/(vendor+'-cli'); events=rows(cli/'session.log')[1:]
        starts=[e for e in events if e['type']=='turn_started']
        assert [e['turn']['policy']['effective_max_model_requests'] for e in starts]==[16,1]
        assert [e['turn']['outcome'] for e in events if e['type']=='turn_ended']==['success','round_limit']
        assert sum(e['type']=='request_sent' for e in events[starts[1]['seq']-1:])==1
        assert any(e['type']=='tool_returned' for e in events[starts[1]['seq']-1:])
        text=(cli/'terminal.txt').read_text();assert '/usage' in text and '/history' in text and 'round_limit' in text
        headless=rows(HERE/(vendor+'-headless')/'terminal.txt')
        assert sorted(x['requests'] for x in headless if x['kind']=='request_count')==[1,2]
        assert sorted(x['policy']['max_model_requests'] for x in headless if x['kind']=='restart')==[1,2]
        run=HERE/('anthropic-browser' if vendor=='anthropic' else vendor+'-browser-native')
        events=rows(run/'session.log')[1:]; starts=[e for e in events if e['type']=='turn_started']
        assert [e['turn']['policy']['effective_max_model_requests'] for e in starts[:2]]==[1,2]
        browser=rows(run/'browser-original.jsonl')
        changes=[json.loads(x['payload']) for x in browser if x['kind']=='browser_received']
        change=next(x for x in changes if x.get('observation',{}).get('kind')=='policy_changed' and x['observation']['execution_policy']['max_model_requests']==2)
        stamp=next(x['at'] for x in browser if x['kind']=='browser_received' and json.loads(x['payload'])==change)
        # A live update while HTTP is pending must not change the active limit.
        ending=next(e for e in events if e['type']=='turn_ended')
        from datetime import datetime
        assert datetime.fromisoformat(starts[0]['time']) < datetime.fromisoformat(stamp) < datetime.fromisoformat(ending['time'])
        native=HERE/(vendor+'-browser-native'); events=rows(native/'session.log')[1:]
        assert any(e['type']=='turn_started' and e['turn']['policy']['max_model_requests']==0 and e['turn']['policy']['effective_max_model_requests']==16 for e in events)
        hints=[e for e in events if e['type']=='hint_received']; assert len(hints)==1
        consuming=[e for e in events if e['type']=='request_sent' and hints[0]['seq'] in e['request']['hints']];assert len(consuming)==1
        assert any(hints[0]['hint']['text'] in p.read_text() for p in (native/'requests').glob('*.json'))
        writes=[e['tool']['args']['content'] for e in events if e['type']=='tool_called' and e['tool']['name']=='write_file']
        assert any('port=9090' in x for x in writes)
        assert (native/'workspace/scratch-report.txt').read_text()==writes[-1]
        assert ('port=9090' in writes[-1]) == (vendor!='anthropic')
        browser=rows(native/'browser-original.jsonl'); actions=[x['action']['action'] for x in browser if x['kind']=='action']
        assert actions.count('reload')>=2 and {'divider','drag-divider','viewport','hint','audio'}<=set(actions)
        assert not [x for x in browser if x['kind'] in ('action_failed','pageerror')]
        interruption={'anthropic':'anthropic-interrupt-final','openai':'openai-embedding','gemini':'gemini-interrupt-correction'}[vendor]
        ir=HERE/interruption; logs=list(ir.glob('session.log'))+list((ir/'workspace').glob('*.jsonl'))
        assert any(e.get('type')=='turn_ended' and e['turn']['outcome']=='interrupted' for log in logs for e in rows(log)[1:])
        results.append({'vendor':vendor,'cli_default_then_limit_one':True,'headless_two_agent_restart':True,
                        'active_one_next_two':True,'zero_default_continuation':True,'hint_consumed_once':True,
                        'scratch_final_retains_correction':vendor!='anthropic','active_interruption_run':interruption})
    for vendor in ('openai','gemini'):
        run=HERE/(vendor+'-native-overlap-replay'); b=rows(run/'browser-original.jsonl')
        receipts={x['sequence']:x['text'] for x in b if x['kind']=='page_receipt'}
        for n,speaking in [(13,2),(15,1),(17,0)]:
            assert f'typing: 1, speaking: {speaking}' in receipts[n]
            assert 'autoplay off, rate 1.6' in receipts[n]
        starts=[x['detail'] for x in b if x['kind']=='speech' and x['detail']['type']=='start']
        assert len([x for x in starts if x['revision']==2 and x['rate']==1.2])==2
        assert any(x['revision']==4 and x['rate']==1.6 for x in starts)
    previous=None
    for name in ('settings-restart-positive','settings-restart-false-zero','settings-restart-confirm'):
        run=HERE/name;b=rows(run/'browser-original.jsonl')
        received=[json.loads(x['payload']) for x in b if x['kind']=='browser_received']
        ps=next(x for x in received if x['type']=='preferences_snapshot')
        policy=next(x['state']['execution_policy'] for x in received if x['type']=='snapshot_begin')
        if previous:
            assert ps['preferences']==previous[0]['preferences'] and ps['revision']==previous[0]['revision']
            assert policy['max_model_requests']==previous[1]['policy']['max_model_requests'] and policy['revision']==previous[1]['revision']
        previous=(read(run/'persisted-gui-preferences.json'),read(run/'persisted-agent-policy.json'))
        assert not list((run/'requests').glob('*.json'))
    assert previous[0]['preferences']['autoplay'] is False and previous[1]['policy']['max_model_requests']==0
    return {'providers':results,'endpoint_disabled_native_scope':'manual real retained cards; queued rate survives shared disable; cancel leaves peer and typing',
            'positive_then_false_zero_fresh_process_restoration':True}


def audit():
    # Bind ALL inputs, not just successful rows, before any derived replay is written.
    originals = frozen_files(FREEZE)
    runs = sorted(p.parent for p in HERE.glob('*/launch.json'))
    assert len(runs) == 22
    initial = frozen_files('5f9684b', {p.name for p in runs})
    bindings = [read(HERE / p) for p in BINDINGS]
    controls = []; phases = []; audio = []; summaries = []; chronological = 0
    if TEACHING_APPEND:
        file_index=index(FREEZE); name='student-review.md'; current=(HERE/name).read_bytes()
        assert bound_raw(FREEZE,name,file_index[name],current)==git('show',FREEZE+':'+PREFIX+name)
        controls.append({'id':'explicit-append-only-teaching-positive','passed':True})
        for label, target, data, reason in [
            ('unbound-teaching-edit',name,current+b'\nunbound change','unbound or rewritten teaching append'),
            ('replaced-teaching-prefix',name,b'rewritten'+current,'unbound or rewritten teaching append'),
            ('raw-receipt-still-exact','anthropic-cli/terminal.txt',(HERE/'anthropic-cli/terminal.txt').read_bytes()+b'\nchanged','frozen original changed')]:
            try:
                bound_raw(FREEZE,target,file_index[target],data)
                raise RuntimeError('mutation accepted: '+label)
            except AssertionError as error:
                assert reason in str(error)
                controls.append({'id':label,'passed':True,'refusal':str(error)})
    totals = {v:{'prompts':0,'http':0} for v in ('anthropic','openai','gemini')}
    with tempfile.TemporaryDirectory(prefix='ch08-independent-live-') as temporary:
        tmp = Path(temporary); loaded = {}
        for n, binding in enumerate(bindings):
            helper, verifier = load_helpers(binding, tmp / ('support-' + str(n)))
            helper.preflight(binding, {})
            loaded[binding['source_revision']] = (binding, helper, verifier)
        for run in runs:
            launch = read(run/'launch.json'); binding, _, _ = loaded[launch['source_revision']]
            launch_check(binding, launch)
            browser_path = run/'browser-original.jsonl'
            if browser_path.exists():
                browser_launch = next(x for x in rows(browser_path) if x['kind']=='launch')
                assert browser_launch['source_revision'] == launch['source_revision']
                assert browser_launch['executables'] == launch['executables']
        controls.append({'id':'all-four-bindings-and-22-launches-before-derived-write','passed':True})
        for revision, (binding, helper, verifier) in loaded.items():
            selected = [p for p in runs if read(p/'launch.json')['source_revision']==revision and read(p/'launch.json')['mode']!='retained-event-replay']
            if selected:
                result = verifier.verify(binding, selected, {}, tmp/('positive-'+revision))
                phases.append({'source':revision,'runs':len(result),'requests':sum(x['requests'] for x in result)})
        # Positive copied valid-path fixture is established before changing one identity.
        binding, helper, verifier = loaded['a06d4f3']
        selected = [HERE/'anthropic-interrupt-final', HERE/'settings-restart-confirm']
        copies = []
        import shutil
        for source in selected:
            dest = tmp/('copy-'+source.name); shutil.copytree(source,dest); copies.append(dest)
        verifier.verify(binding,copies,{},tmp/'positive-copy')
        controls.append({'id':'valid-copied-path-positive','passed':True})
        late = copies[-1]/'launch.json'; saved = late.read_bytes()
        mutations = [('source-hash','historical source mismatch'),('source-set','incomplete historical source set'),
                     ('empty-source','empty source set'),('executable-set','incomplete executable identities'),
                     *[(x+'-hash','interpreter mismatch' if x=='interpreter' else 'executable mismatch') for x in binding['executables']],
                     ('support-hash','historical support mismatch'),('support-set','incomplete support identities'),
                     ('browser-hash','browser dependency mismatch'),('browser-set','incomplete browser dependency set'),
                     ('last-source','launch source mismatch'),('last-binary','launch executable mismatch'),
                     ('last-support','launch support mismatch'),('last-browser','launch browser tools mismatch'),('last-role','unknown launch executable')]
        for name, reason in mutations:
            b = copy.deepcopy(binding); launch = json.loads(saved)
            if name=='source-hash': b['sources'][next(iter(b['sources']))]='0'*64
            elif name=='source-set': b['sources'].pop(next(iter(b['sources'])))
            elif name=='empty-source': b['sources']={}
            elif name=='executable-set': b['executables'].pop('gui')
            elif name.endswith('-hash') and name[:-5] in b['executables']: b['executables'][name[:-5]]['sha256']='0'*64
            elif name=='support-hash': b['support']['evidence.py']='0'*64
            elif name=='support-set': b['support'].pop('evidence.py')
            elif name=='browser-hash': b['browser_tools'][next(iter(b['browser_tools']))]='0'*64
            elif name=='browser-set': b['browser_tools'].pop(next(iter(b['browser_tools'])))
            elif name=='last-source': launch['source_revision']='incorrect'
            elif name=='last-binary': launch['executables']['gui']='0'*64
            elif name=='last-support': launch['support']['evidence.py']='0'*64
            elif name=='last-browser': launch['browser_tools'].pop(next(iter(launch['browser_tools'])))
            elif name=='last-role': launch['launched_executable']='unknown'
            else: raise AssertionError(name)
            late.write_text(json.dumps(launch)); target=tmp/name
            try:
                verifier.verify(b,copies,{},target)
                raise RuntimeError('mutation accepted: '+name)
            except AssertionError as error:
                assert reason in str(error), (name,str(error))
                assert not target.exists(), 'derived write before refusal: '+name
                controls.append({'id':name,'passed':True,'refusal':str(error)})
            finally: late.write_bytes(saved)
        # Put the changed request last: earlier valid reconstruction still may not write.
        body=copies[0]/'requests/001.json'; saved_body=body.read_bytes()
        value=json.loads(saved_body);value['model']='review-intended-body-mismatch';body.write_text(json.dumps(value))
        try:
            verifier.verify(binding,list(reversed(copies)),{},tmp/'late-body')
            raise RuntimeError('changed body accepted')
        except AssertionError as error:
            assert 'request reconstruction mismatch' in str(error) and not (tmp/'late-body').exists()
            controls.append({'id':'last-request-body-before-write','passed':True,'refusal':str(error)})
        finally: body.write_bytes(saved_body)
        for run in runs:
            launch=read(run/'launch.json'); binding,helper,_=loaded[launch['source_revision']]
            browser=rows(run/'browser-original.jsonl') if (run/'browser-original.jsonl').exists() else []
            errors=[x for x in browser if x['kind'] in ('pageerror','action_failed')]
            audio.extend(native_audio(run,browser))
            if launch['mode']=='retained-event-replay':
                retained=launch['retained']; original=git('show',retained['source_revision']+':'+retained['path'])
                assert digest(original)==retained['sha256']
                before=[json.loads(x) for x in original.splitlines()]; after=rows(run/'workspace/readmitted.jsonl')
                assert len(before)==len(after) and before[0]==after[0]
                for old,new in zip(before[1:],after[1:]):
                    assert old['time'] and new['time']
                    assert {k:v for k,v in old.items() if k!='time'}=={k:v for k,v in new.items() if k!='time'}
                assert launch['provider_endpoint']=='http://127.0.0.1:1'
                assert not list((run/'requests').glob('*'))
                summaries.append({'run':run.name,'source':launch['source_revision'],'kind':'retained-event/native replay; zero new HTTP',
                                  'events':len(after)-1,'only_admission_time_may_change':True,'browser_errors':len(errors)})
                continue
            requests=[read(p) for p in sorted((run/'requests').glob('*.json'))]
            assert len(requests)==launch['requests']==len(list((run/'responses').glob('*.body')))
            assert launch['exit_code']==(1 if run.name.endswith('-cli') else 0)
            logs=sorted(list(run.glob('*.log'))+list((run/'workspace').glob('*.jsonl')))
            active_logs=[]; prompts=0; usage={}; outcomes=[]
            for log in logs:
                events=rows(log)[1:]; assert [e['seq'] for e in events]==list(range(1,len(events)+1))
                if any(e['type']=='request_sent' for e in events): active_logs.append(log)
                prompts+=sum(e['type']=='turn_started' for e in events)
                outcomes.extend(e['turn']['outcome'] for e in events if e['type']=='turn_ended')
                for e in events:
                    if e['type']=='response_ended':
                        for k,v in e.get('response',{}).get('usage',{}).items():usage[k]=usage.get(k,0)+v
            if len(active_logs)==1:
                log=active_logs[0]
                rendered=[json.loads(subprocess.check_output([binding['executables']['cli']['path'],'replay',str(log),str(e['seq'])],cwd=run)) for e in rows(log)[1:] if e['type']=='request_sent']
                assert list(map(canonical,requests))==list(map(canonical,rendered)), 'single Agent chronology changed: '+run.name
                chronological+=1
            totals[launch['vendor']]['http']+=len(requests);totals[launch['vendor']]['prompts']+=prompts
            summaries.append({'run':run.name,'source':launch['source_revision'],'model':launch['requested_model'],'http':len(requests),
                              'prompts':prompts,'outcomes':outcomes,'response_usage':usage,'browser_errors':len(errors)})
    assert totals=={'anthropic':{'prompts':11,'http':18},'openai':{'prompts':9,'http':14},'gemini':{'prompts':10,'http':15}}, totals
    assert len(audio)==5 and sum(x['requests'] for x in phases)==47
    # Read only the three authorized values, keep them in memory, never print them.
    credentials=read(Path.home()/'.cr/settings.json')
    keys=[credentials.get(k,'').encode() for k in ('directClaudeAPIKey','directOpenAIAPIKey','directGeminiAPIKey')]
    assert all(keys), 'needed credential absent; cannot complete exact exclusion scan'
    scanned=[p for p in HERE.rglob('*') if p.is_file()]
    matches=[str(p.relative_to(HERE)) for p in scanned if any(k in p.read_bytes() for k in keys)]
    assert not matches, 'credential found in evidence (value withheld): '+str(matches)
    assert originals==frozen_files(FREEZE), 'originals changed during audit'
    return {'accepted':True,'evidence_revision':git('rev-parse',FREEZE).decode().strip(),'script_sha256':sha(Path(__file__)),
            'original_files':len(originals),'initial_retained_files':len(initial),'original_sha256':originals,
            'separate_teaching_append':({'revision':git('rev-parse',TEACHING_APPEND).decode().strip(),
                                         'path':'student-review.md','sha256':sha(HERE/'student-review.md')} if TEACHING_APPEND else None),
            'bindings':{p:sha(HERE/p) for p in BINDINGS},'phases':phases,'totals':totals,'chronological_single_agent_runs':chronological,
            'controls':controls,'audio':audio,'runs':summaries,'feature_matrix':feature_audit(),
            'exact_credential_scan':{'files':len(scanned),'matches':0},
            'limitations':['Audio amplitude and PID are verified; no transcription/human listening claimed.',
                           'Retained replay rewrites top-level admission time only and makes no new provider request.',
                           'Initial native interference, tool refusals and late interruption remain preserved.']}


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--teaching-append',help='Exact commit of an append-only student-review.md update; no other evidence may differ')
    TEACHING_APPEND=parser.parse_args().teaching_append
    print(json.dumps(audit(),indent=2))
