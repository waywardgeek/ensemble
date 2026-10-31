"""Rebuild only print artifacts and inspect their real Chrome print styling."""
from pathlib import Path
import hashlib, importlib.util, json, os, re, subprocess, tempfile, time, urllib.request
import websocket
from PIL import Image
root = Path.cwd()
out = root / 'solutions/evidence/publication-print-light'
book = root / 'book'
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
protected = list(book.glob('*.md')) + [book / 'The Art of Building AI Coding Agents.epub', book / 'epub.css']
protected = [p for p in protected if p.name != 'README.md']
before = {str(p.relative_to(root)): sha(p) for p in protected}
commands = []
def run(name, args):
    start = time.monotonic()
    p = subprocess.run(args, capture_output=True, text=True, timeout=180)
    (out / (name + '.txt')).write_text(p.stdout + p.stderr)
    commands.append({'command': args, 'exit': p.returncode, 'seconds': round(time.monotonic()-start, 3)})
    assert p.returncode == 0, p.stderr
run('print', ['python3', 'book/build-print.py'])
with tempfile.TemporaryDirectory(prefix='ensemble-light-cover-') as tmp:
    cover = Path(tmp) / 'cover.pdf'
    Image.open(book / 'cover.png').convert('RGB').resize((1800,2700), Image.Resampling.LANCZOS).save(cover, 'PDF', resolution=300)
    run('cover', ['pdfunite', str(cover), str(book / 'The Art of Building AI Coding Agents - Print Interior.pdf'), str(book / 'The Art of Building AI Coding Agents.pdf')])

chrome = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
with tempfile.TemporaryDirectory(prefix='ensemble-light-browser-') as tmp:
    proc = subprocess.Popen([chrome, '--headless=new', '--no-sandbox', '--remote-debugging-port=0', '--remote-allow-origins=*', '--user-data-dir='+tmp, 'about:blank'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        portfile = Path(tmp) / 'DevToolsActivePort'
        deadline = time.monotonic()+15
        while not portfile.exists():
            assert proc.poll() is None, 'browser child exited'
            assert time.monotonic()<deadline, 'browser readiness timeout'
            time.sleep(.05)
        port = int(portfile.read_text().splitlines()[0])
        targets = json.load(urllib.request.urlopen(f'http://127.0.0.1:{port}/json'))
        ws = websocket.create_connection(next(t['webSocketDebuggerUrl'] for t in targets if t['type']=='page'), timeout=20)
        seq = 0
        def call(method, params={}):
            global seq
            seq += 1
            ws.send(json.dumps({'id':seq,'method':method,'params':params}))
            while True:
                reply = json.loads(ws.recv())
                if reply.get('id') == seq:
                    assert 'error' not in reply, reply
                    return reply.get('result',{})
        def evaluate(js):
            result=call('Runtime.evaluate', {'expression':js,'returnByValue':True})
            assert 'exceptionDetails' not in result, result
            return result['result'].get('value')
        call('Emulation.setEmulatedMedia', {'media':'print'})
        call('Page.navigate', {'url':(book/'interior-print.html').as_uri()})
        deadline=time.monotonic()+15
        while not evaluate('document.readyState === "complete" && !!document.querySelector("pre .kw")'):
            assert time.monotonic()<deadline
            time.sleep(.05)
        js='''JSON.stringify(Array.from(document.querySelectorAll('pre, code, div.sourceCode, pre *, code *')).map(e=>{let s=getComputedStyle(e), r=e.getBoundingClientRect();return {tag:e.tagName,cls:e.className,color:s.color,background:s.backgroundColor,weight:s.fontWeight,style:s.fontStyle,width:r.width,height:r.height}}))'''
        styled=json.loads(evaluate(js))
        assert styled and all(e['color']=='rgb(0, 0, 0)' and e['background']=='rgb(255, 255, 255)' for e in styled)
        samples=evaluate('''Object.fromEntries(['st','co','kw','op'].map(c=>{let e=document.querySelector('pre .'+c),s=getComputedStyle(e);return [c,{text:e.textContent,color:s.color,background:s.backgroundColor,weight:s.fontWeight,style:s.fontStyle}]}))''')
        assert evaluate('matchMedia("print").matches') is True
        removed=evaluate('''(()=>{for(let sheet of document.styleSheets){for(let i=sheet.cssRules.length-1;i>=0;i--){let r=sheet.cssRules[i];if(r.selectorText==='pre, code, div.sourceCode, pre *, code *'){sheet.deleteRule(i);return true}}}return false})()''')
        assert removed
        original=json.loads(evaluate(js))
        assert len(original)==len(styled)
        for a,b in zip(original,styled):
            assert all(a[k]==b[k] for k in ['tag','cls','weight','style','width','height']), (a,b)
        assert any(e['color']!='rgb(0, 0, 0)' for e in original), 'original syntax colors not observed'
        browser={'print_media':True,'elements_checked':len(styled),'samples':samples,'unchanged_emphasis_and_geometry':True,'original_colored_elements':sum(e['color']!='rgb(0, 0, 0)' for e in original)}
        ws.close()
    finally:
        proc.terminate()
        proc.wait(timeout=10)

spec=importlib.util.spec_from_file_location('printbuild', book/'build-print.py'); module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
html=(book/'interior-print.html').read_text()
entries=module.toc_entries(html)
mapping=module.build_map(book/'The Art of Building AI Coding Agents - Print Interior.pdf',[module.strip_tags(t) for _,t in entries])
prior=json.loads((root/'solutions/evidence/publication-preface-r2/full-verification.json').read_text())
assert mapping==prior['print_heading_pages']
assert len(mapping)==30
for href,raw in entries:
    assert re.search(r'href="#'+re.escape(href)+r'"[^>]*>.*?</a>\s*<span class="tocpg">'+str(mapping[module.strip_tags(raw)])+r'</span>',html,re.S)
for name,count in [('The Art of Building AI Coding Agents - Print Interior.pdf',371),('The Art of Building AI Coding Agents.pdf',372)]:
    info=subprocess.check_output(['pdfinfo',str(book/name)],text=True)
    assert re.search(r'Pages:\s+'+str(count)+r'\b',info)
    assert '432 x 648 pts' in info
assert before=={str(p.relative_to(root)):sha(p) for p in protected},'protected manuscript/EPUB changed'
artifacts=['print-header.html','README.md','interior-print.html','The Art of Building AI Coding Agents - Print Interior.pdf','The Art of Building AI Coding Agents.pdf']
receipt={'baseline':'11f5562a','result':'PASS','commands':commands,'browser':browser,'print_heading_pages':mapping,'interior_pages':371,'cover_pages':372,'protected_unchanged':before,'outputs':{str((book/n).relative_to(root)):sha(book/n) for n in artifacts}}
(out/'check.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({'result':'PASS','browser':browser,'interior_pages':371,'cover_pages':372},indent=2))
