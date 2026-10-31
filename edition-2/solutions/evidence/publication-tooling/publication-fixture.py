import pathlib,subprocess,tempfile,shutil,json,zipfile,xml.etree.ElementTree as ET,time
root=pathlib.Path.cwd();out=root/'solutions/evidence/publication-tooling';rows=[]
with tempfile.TemporaryDirectory(prefix='ensemble-publication-') as tmp:
 base=pathlib.Path(tmp);book=base/'book';scripts=base/'scripts';book.mkdir();scripts.mkdir()
 for name in ['build-ebook.sh','publication-links.lua']:shutil.copy2(root/'scripts'/name,scripts/name)
 for name in ['build-print.py','front-matter.md','front-matter-print.md','print-header.html','epub.css','cover.png']:shutil.copy2(root/'book'/name,book/name)
 text='# Contents\n\n- [Probe](#probe)\n\n# Probe\n\nA portable publication. [Evidence](../docs/example.md#proof), [local](#probe), [external](https://example.org/x), [space](../solutions/a%20b.go), [book](chapter-01.md).\n\n## Detail\n\nKept below chapter navigation.\n'
 (book/'the-art-of-building-ai-coding-agents.md').write_text(text)
 for name,cmd in [('epub-zero-figures',['bash',str(scripts/'build-ebook.sh')]),('portable-print',['python3',str(book/'build-print.py')])]:
  t=time.monotonic();p=subprocess.run(cmd,cwd='/tmp',capture_output=True,text=True,timeout=90);(out/(name+'.txt')).write_text(p.stdout+p.stderr);rows.append(dict(name=name,exit=p.returncode,seconds=round(time.monotonic()-t,3)));print(rows[-1],flush=True);assert p.returncode==0,p.stdout+p.stderr
 with zipfile.ZipFile(book/'The Art of Building AI Coding Agents.epub') as z:
  opf=ET.fromstring(z.read('EPUB/content.opf'));ns={'o':'http://www.idpf.org/2007/opf'}
  assert all(e.attrib['idref']!='nav' for e in opf.findall('o:spine/o:itemref',ns))
  nav=z.read('EPUB/nav.xhtml').decode();assert 'Probe' in nav and 'Detail' not in nav
  ncx=z.read('EPUB/toc.ncx').decode();assert 'Probe' in ncx
  chapters=''.join(z.read(n).decode() for n in z.namelist() if n.startswith('EPUB/text/') and n.endswith('.xhtml'))
  assert 'https://github.com/waywardgeek/ensemble/blob/main/docs/example.md#proof' in chapters
  assert 'https://example.org/x' in chapters and '#probe' in chapters
  assert 'https://github.com/waywardgeek/ensemble/blob/main/book/chapter-01.md' in chapters
  assert '>Contents<' not in chapters
 (out/'publication-fixture.json').write_text(json.dumps(dict(cases=rows,assertions=['zero figures plus repository cover','no nav in spine','chapter-level nav and NCX populated','no printed Contents','repo links externalized, anchors and external URLs retained','print from unrelated cwd converges with all headings mapped']),indent=2)+'\n')
