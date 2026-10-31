#!/usr/bin/env python3
"""Require intended failures for settings consumers and browser lifetime deletions."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import tempfile


def execute(source,case=None):
    command=['node',str(Path(__file__).with_name('accept_ch08_browser.cjs')),str(source)]
    if case:command.append(case)
    result=subprocess.run(command,capture_output=True,text=True,timeout=120)
    receipt=json.loads(result.stdout)
    return dict(command=command,exit=result.returncode,receipt=receipt,stderr=result.stderr)


def evaluate(source):
    with tempfile.TemporaryDirectory(prefix='ch08-browser-deletions-') as temporary:
        base=Path(temporary)/'positive'
        shutil.copytree(source/'gui/web/gui',base/'gui/web/gui')
        positive=execute(base)
        if positive['exit'] or not positive['receipt']['passed']:
            return dict(passed=False,positive=positive,mutants=[])
        speech='speech-enqueue-revision-rate-off-buffer-and-local-cancel'
        specs=[
            ('font-consumer','preferences.js',"this.root.style.setProperty('--font-size', p.font_size + 'px');",'void p.font_size;',
             'remote-style-preserves-cards-focus-and-system-theme','remote preferences lack a visible consumer'),
            ('keyboard-step','preferences.js',"? 10 : -10","? 1 : -1",'semantic-keyboard-pointer-and-unsaved-value','Timeout'),
            ('pointer-commit-boundary','preferences.js',"drag.value + 'px');","drag.value + 'px'); this.change({[field]:drag.value}, element);",'semantic-keyboard-pointer-and-unsaved-value','pointer motion persisted'),
            ('discard-off-buffer','speech.js','if (!this.enabled) for (const cursor of this.cursors.values()) cursor.sent = cursor.text.length;','if (false) for (const cursor of this.cursors.values()) cursor.sent = cursor.text.length;',speech,'off buffer replayed'),
            ('captured-rate','speech.js','}, item.rate);','}, this.rate);',speech,'queued rate changed'),
            ('captured-revision','speech.js','revision: item.revision, rate: item.rate','revision: this.revision, rate: item.rate',speech,'preference revision not captured'),
            ('shared-disable-cancels-existing','speech.js','if (!this.enabled) for (const cursor','if (!this.enabled) this.cancel();\n    if (!this.enabled) for (const cursor',speech,'shared disable canceled'),
            ('abandoned-settings-socket','connector.js','socket.onmessage = event => {\n      if (socket !== this.socket) return;','socket.onmessage = event => {',
             'old-socket-preferences-fenced-before-generation','abandoned socket overwrote preferences before new generation'),
            ('early-settings-ready','connector.js','this.owner.preferences?.(this.preferences); return;','this.owner.preferences?.(this.preferences); this.settingsReady = true; return;',
             'settings-wait-for-complete-initial-domains','settings ready before Agent snapshot end'),
            ('part-routing','page.js',"actions ? 'chat' : 'all'","'all'",'mixed-parts-safe-expansion-and-restyle-identity','final answer duplicated provisional card'),
            ('remount-transient-controls','preferences.js',"for (const control of this.root.querySelectorAll('[data-preference], [data-policy-save], [data-divider]')) control.disabled = false;",'/* deleted new owner control reset */',
             'pending-settings-disposal-and-same-dom-remount','remounted settings retained old owner disabled control'),
            ('closed-settings-continuation','preferences.js','apply(snapshot) {\n    if (this.closed || this.page.closed) return;','apply(snapshot) {',
             'pending-settings-disposal-and-same-dom-remount','late old settings continuation overwrote replacement draft'),
        ]
        for name,file,old,new,case,reason in specs:
            assert (base/'gui/web/gui'/file).read_text().count(old)==1,'browser mutant anchor not unique: '+name
        rows=[]
        for name,file,old,new,case,reason in specs:
            mutated=Path(temporary)/name
            shutil.copytree(base,mutated)
            target=mutated/'gui/web/gui'/file;text=target.read_text();assert text.count(old)==1,'browser mutant anchor not unique: '+name
            target.write_text(text.replace(old,new))
            result=execute(mutated,case);checks=result['receipt']['checks']
            passed=result['exit']!=0 and len(checks)==1 and checks[0]['id']==case and not checks[0]['passed'] and reason in checks[0].get('error','')
            rows.append(dict(id=name,passed=passed,intended_case=case,intended_reason=reason,result=result))
            shutil.rmtree(mutated)
        here=Path(__file__).parent
        return dict(passed=all(r['passed'] for r in rows),positive=positive,mutants=rows,
                    checker_files={name:hashlib.sha256((here/name).read_bytes()).hexdigest() for name in ['audit_ch08_browser.py','accept_ch08_browser.cjs','ch08-browser-fixture.js']})


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('source_directory',type=Path)
    args=parser.parse_args();result=evaluate(args.source_directory.resolve(strict=True));print(json.dumps(result,indent=2));raise SystemExit(0 if result['passed'] else 1)
