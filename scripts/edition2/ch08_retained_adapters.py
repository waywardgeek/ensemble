#!/usr/bin/env python3
"""Adapt prior fixtures to added settings domains without changing their assertions.

Only a disposable checker bundle is edited. Original historical graders remain
byte-for-byte intact. The receipt records each exact fixture/anchor adaptation.
"""
import hashlib
from pathlib import Path
import shutil

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[1]
PREFERENCES = "{type:'preferences_snapshot',revision:0,preferences:{theme:'dark',font_size:16,sidebar_width:260,actions_width:380,autoplay:false,speech_rate:1}}"
POLICY = "{revision:0,persistent:false,max_model_requests:0,effective_max_model_requests:16}"


def prepare(destination, source=None):
    shutil.copytree(HERE, destination, ignore=shutil.ignore_patterns('__pycache__'))
    changes = []
    def edit(name, old, new, count=1):
        path=destination/name; text=path.read_text()
        assert text.count(old)==count, f'{name}: expected {count} adapter anchors for {old!r}, got {text.count(old)}'
        path.write_text(text.replace(old,new))
        changes.append(dict(file=name,old=old,new=new,count=count))

    # Existing wire assertions still see the contiguous Agent group, after a
    # separately validated new preferences snapshot, in the required order.
    edit('accept_ch07.py', "    assert len(records) == 2, 'empty snapshot contains unexpected events or live records'",
         "    assert len(records) == 3, 'preferences plus empty Agent snapshot must contain three records'\n"
         "    prefs, records = records[0], records[1:]\n"
         "    assert prefs == dict(type='preferences_snapshot', revision=0, preferences=dict(theme='dark',font_size=16,sidebar_width=260,actions_width=380,autoplay=False,speech_rate=1)), 'initial preferences missing or unsafe'\n"
         "    assert len(records) == 2, 'empty snapshot contains unexpected events or live records'")
    edit('ch07-review-connector_test.go.txt', '\tif got := probeJSON(t, peer); got["type"] != "snapshot_begin" {',
         '\tif got := probeJSON(t, peer); got["type"] != "preferences_snapshot" { t.Fatalf("initial preferences: %v", got) }\n\tif got := probeJSON(t, peer); got["type"] != "snapshot_begin" {')

    # A white-box Actor fixture formerly inherited unused methods from a nil
    # interface. Its snapshot now legitimately reads its own policy owner.
    edit('ch07_internal_test.go.txt', 'func (a *c7Agent) ID() string',
         'type c8FixturePolicy struct { common.PolicyService }\nfunc (p *c8FixturePolicy) Snapshot() common.PolicySnapshot { return common.PolicySnapshot{EffectiveMaxModelRequests:16} }\nfunc (a *c7Agent) Policy() common.PolicyService { return &c8FixturePolicy{} }\nfunc (a *c7Agent) ID() string')

    # New fields arrive from the controlled server, not a production bypass.
    edit('accept_ch07_browser.cjs', 'const frame=m=>s.onmessage({data:JSON.stringify(m)});',
         "const frame=m=>{if(m.type==='snapshot_begin'){s.onmessage({data:JSON.stringify("+PREFERENCES+")});m.state={...m.state,execution_policy:"+POLICY+",active_max_model_requests:null};}s.onmessage({data:JSON.stringify(m)});};")
    for name in ['accept_ch07_browser.cjs','ch07-review-browser-lifetime.cjs','ch07-review-speech-ownership.cjs']:
        edit(name,'frame(m){this.onmessage?.({data:JSON.stringify(m)});}',
             "frame(m){if(m.type==='snapshot_begin'){this.onmessage?.({data:JSON.stringify("+PREFERENCES+")});m.state={...m.state,execution_policy:"+POLICY+",active_max_model_requests:null};}this.onmessage?.({data:JSON.stringify(m)});}")

    edit('ch05_stale_test.go.txt', 'func (a *auditLateAgent) ID() string',
         'type auditLatePolicy struct { common.PolicyService }\nfunc (p *auditLatePolicy) Snapshot() common.PolicySnapshot { return common.PolicySnapshot{EffectiveMaxModelRequests:16} }\nfunc (a *auditLateAgent) Policy() common.PolicyService { return &auditLatePolicy{} }\nfunc (a *auditLateAgent) ID() string')
    edit('accept_ch05.py', "HERE/'packagecheck/main.go'", "HERE/'ch08packagecheck/main.go'",2)
    edit('ch07-review-prior-cli.py', 'p.relative_to(REPO)', 'p.relative_to(HERE)')
    edit('ch07-review-killed-transport.py', "log=HERE.parents[1]/'solutions/edition-2/main/evidence/ch07/browser-anthropic-r1/session.log'",
         "log=Path("+repr(str(REPO/'solutions/edition-2/main/evidence/ch07/browser-anthropic-r1/session.log'))+")")
    # Settings adds a MediaQueryList listener. Track it too, without passing a
    # non-Node to Node.contains; retained disposal assertions become stronger.
    edit('ch07-review-browser-lifetime.cjs','this===root||root.contains(this)',
         'this===root||(this instanceof Node&&root.contains(this))||this instanceof MediaQueryList')

    # Queue accounting now includes deferred initial-handoff control records.
    edit('ch07-review-connector-mutations.py', 'len(c.queue) < ensemble.WatchItems', 'len(c.queue)+len(c.deferred) < ensemble.WatchItems')
    edit('ch07-review-connector-mutations.py', 'len(c.queue) <= ensemble.WatchItems', 'len(c.queue)+len(c.deferred) <= ensemble.WatchItems')
    # Only delete the completion fence, preserving the newly separate no-start
    # timeout guard; this retains the original stale-completion experiment.
    revised = source is not None and '  finish(request, error) {' in (source/'gui/web/gui/speech-service.js').read_text()
    before = '  finish(request, error) {\n    if (this.active !== request) return;' if revised else 'const finish = error => {\n      if (this.active !== request) return;'
    after = '  finish(request, error) {' if revised else 'const finish = error => {'
    edit('audit_ch07_browser.py', "anchor='if (this.active !== request) return;'", 'anchor='+repr(before))
    edit('audit_ch07_browser.py', "service_text.replace(anchor,'')", 'service_text.replace(anchor,'+repr(after)+')')
    if revised:
        # Retained document-owner tests use an immediate controlled lease, just
        # as they use controlled synthesis. Cross-tab scheduling is separately
        # tested with Chrome's real LockManager by accept_ch08_speech_tabs.cjs.
        for name in ['accept_ch07_browser.cjs','ch07-review-browser-lifetime.cjs','ch07-review-speech-ownership.cjs']:
            path=destination/name;text=path.read_text()
            marker="await page.goto(origin" # each fixture establishes one page
            # Install before any page evaluation/import, while preserving all
            # existing fixture assertions and native overlap/cancel counters.
            position=text.index('const result=await page.evaluate') if name!='accept_ch07_browser.cjs' else text.index('async function run(id,action)')
            if name=='accept_ch07_browser.cjs':
                old="await page.goto(origin+'/review.html');await action(page);"
                new="await page.goto(origin+'/review.html');await page.evaluate(()=>Object.defineProperty(navigator,'locks',{configurable:true,value:{request(_name,_options,callback){try{return Promise.resolve(callback());}catch(error){return Promise.reject(error);}}}}));await action(page);"
            else:
                old='const result=await page.evaluate'
                new="await page.evaluate(()=>Object.defineProperty(navigator,'locks',{configurable:true,value:{request(_name,_options,callback){try{return Promise.resolve(callback());}catch(error){return Promise.reject(error);}}}}));\n const result=await page.evaluate"
            edit(name,old,new)
    edit('ch07-review-browser-mutations.py', "'if (this.active?.page !== page) return;', ''", "'if (this.active?.page !== page) return;', 'if (!this.active) return;'")
    # Scripts that inspect immutable Git history still use the real repository.
    for path in sorted(destination.glob('*.py')):
        text=path.read_text()
        for old in ['REPO = HERE.parents[1]','REPO=HERE.parents[1]']:
            if old in text:
                new="REPO = __import__('pathlib').Path("+repr(str(REPO))+")"
                path.write_text(text.replace(old,new));text=path.read_text()
                changes.append(dict(file=path.name,old=old,new=new,count=1))
    return dict(changes=changes, original_files={str(p.relative_to(destination)):hashlib.sha256((HERE/p.relative_to(destination)).read_bytes()).hexdigest() for p in destination.rglob('*') if p.is_file()},
                adapted_files={str(p.relative_to(destination)):hashlib.sha256(p.read_bytes()).hexdigest() for p in destination.rglob('*') if p.is_file()})
