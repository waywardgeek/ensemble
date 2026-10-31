import pathlib,subprocess,zipfile,json,re,hashlib,xml.etree.ElementTree as ET,importlib.util
root=pathlib.Path.cwd();book=root/'book';out=root/'solutions/evidence/publication-tooling'
spec=importlib.util.spec_from_file_location('printbuild',book/'build-print.py');m=importlib.util.module_from_spec(spec);spec.loader.exec_module(m)
html=(book/'interior-print.html').read_text();entries=m.toc_entries(html);titles=[m.strip_tags(t) for _,t in entries];assert len(titles)==30
pdf=book/'The Art of Building AI Coding Agents - Print Interior.pdf';mapping=m.build_map(pdf,titles);assert len(mapping)==30
for href,title in entries:
 assert re.search(r'href="#'+re.escape(href)+r'"[^>]*>.*?</a>\s*<span class="tocpg">'+str(mapping[m.strip_tags(title)])+'</span>',html,re.S),title
sourcefiles=[book/'preface.md']+sorted(book.glob('chapter-[0-9][0-9].md'))+[book/'epilogue.md',book/'note-from-bill.md'];assembled=(book/'the-art-of-building-ai-coding-agents.md').read_text()
assert len(sourcefiles)==30
assert all(p.read_text().rstrip() in assembled for p in sourcefiles)
ns={'h':'http://www.w3.org/1999/xhtml','o':'http://www.idpf.org/2007/opf','n':'http://www.daisy.org/z3986/2005/ncx/','d':'http://purl.org/dc/elements/1.1/'}
with zipfile.ZipFile(book/'The Art of Building AI Coding Agents.epub') as z:
 opf=ET.fromstring(z.read('EPUB/content.opf'));assert all(e.attrib['idref']!='nav' for e in opf.findall('o:spine/o:itemref',ns))
 nav=ET.fromstring(z.read('EPUB/nav.xhtml'));toc=next(e for e in nav.findall('.//h:nav',ns) if e.attrib.get('id')=='toc')
 navtitles=[''.join(e.itertext()) for e in toc.findall('.//h:a',ns)]
 assert len(navtitles)==31 and navtitles[0]=="The Art of Building AI Coding Agents",(len(navtitles),navtitles)
 ncx=ET.fromstring(z.read('EPUB/toc.ncx'));assert len(ncx.findall('.//n:navPoint',ns))==32
 headings=[];repo_links=0;localfilelinks=[];directory_links=0
 for n in z.namelist():
  if not n.startswith('EPUB/text/') or not n.endswith('.xhtml'):continue
  doc=ET.fromstring(z.read(n));headings += [''.join(e.itertext()) for e in doc.findall('.//h:h1',ns)]
  for e in doc.findall('.//h:a',ns):
   href=e.attrib.get('href','')
   if 'github.com/waywardgeek/ensemble/' in href:repo_links+=1
   if href=='https://github.com/waywardgeek/ensemble/tree/main/docs/edition-2-notes':directory_links+=1
   if href and not href.startswith(('#','http:','https:','mailto:')) and not re.match(r'(?:\.\./)?(?:text/)?[^/]+\.xhtml(?:#.*)?$',href):localfilelinks.append(href)
 assert not localfilelinks,localfilelinks
 assert directory_links>=1
 assert headings==["The Art of Building AI Coding Agents"]+navtitles,(headings,navtitles)
 assert 'Contents' not in headings and len(headings)==32
 assert len([n for n in z.namelist() if re.search(r'\.(png|jpe?g|gif)$',n,re.I)])==1
 counts={'chapters_00_through_26':27,'content_h1':len(headings),'epub_toc':len(navtitles),'epub_ncx':32,'images':1,'repository_links':repo_links}
for name,want in [('The Art of Building AI Coding Agents - Print Interior.pdf',370),('The Art of Building AI Coding Agents.pdf',371)]:
 info=subprocess.check_output(['pdfinfo','-f','1','-l',str(want),str(book/name)],text=True);(out/(('interior' if want==370 else 'cover')+'-pdfinfo.txt')).write_text(info)
 assert re.search(r'^Pages:\s+'+str(want)+'$',info,re.M)
 sizes=re.findall(r'^Page\s+\d+ size:\s+([^\n]+)',info,re.M);assert len(sizes)==want and all(x.startswith('432 x 648 pts') for x in sizes),(len(sizes),sizes[:3])
artifacts=['the-art-of-building-ai-coding-agents.md','interior-print.html','The Art of Building AI Coding Agents.epub','The Art of Building AI Coding Agents - Print Interior.pdf','The Art of Building AI Coding Agents.pdf']
(out/'full-verification.json').write_text(json.dumps(dict(result='PASS',counts=counts,print_heading_pages=mapping,interior_pages=370,cover_prefixed_pages=371,page_size='432 x 648 pt for every page',artifacts={p:dict(bytes=(book/p).stat().st_size,sha256=hashlib.sha256((book/p).read_bytes()).hexdigest()) for p in artifacts}),indent=2)+'\n')
print(json.dumps(dict(counts=counts,pages=mapping),indent=2))
