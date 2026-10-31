#!/usr/bin/env python3
"""Build the human companion without changing the original/Edition 3 artifacts."""
import argparse
import importlib.util
import re
import subprocess
import tempfile
import xml.etree.ElementTree as ET
from pathlib import Path
from posixpath import dirname, normpath, join
from urllib.parse import urlsplit, unquote
from zipfile import ZipFile

BOOK = Path(__file__).resolve().parent
ROOT = BOOK.parents[1]
TITLE = "The Art of Building AI Coding Agents: A Human Companion"
STEM = "The Art of Building AI Coding Agents - A Human Companion"
PARTS = ["preface.md", *(f"chapter-{n:02}.md" for n in range(28)), "epilogue.md"]


def run(*args, **kwargs):
    return subprocess.run(args, check=True, text=True, **kwargs)


def assemble(directory=BOOK):
    """An explicit order prevents missing chapters being hidden by a glob."""
    texts, titles = [], []
    for name in PARTS:
        text = (directory / name).read_text()
        headings = re.findall(r"^# (.+)$", text, re.M)
        if len(headings) != 1 or not text.startswith("# "):
            raise ValueError(f"{name}: expected one opening level-1 heading")
        texts.append(text.strip())
        titles.extend(headings)
    return "\n\n".join(texts) + "\n", titles


def pandoc(source, output, *options):
    run("pandoc", str(source), "--from=markdown-smart", "-o", str(output),
        "--lua-filter", str(ROOT / "scripts/publication-links.lua"),
        "--metadata=publication-directory:book/companion",
        f"--metadata=title:{TITLE}", "--metadata=author:Bill Cox",
        f"--resource-path={BOOK}", *options)


def build_markdown(body, output, work):
    # Reuse the duplicate-aware GitHub contents builder.
    toc = run("python3", str(ROOT / "scripts/book-toc.py"),
              input="<!-- toc -->\n\n" + body, capture_output=True).stdout
    source = work / "with-contents.md"
    source.write_text(toc)
    pandoc(source, output, "--to=gfm", "--wrap=none")


def check_epub(path, titles, has_cover):
    """Check the saved package, not just a successful pandoc exit."""
    ns = {"o": "http://www.idpf.org/2007/opf", "h": "http://www.w3.org/1999/xhtml",
          "n": "http://www.daisy.org/z3986/2005/ncx/"}
    with ZipFile(path) as epub:
        opf = ET.fromstring(epub.read("EPUB/content.opf"))
        items = opf.findall("o:manifest/o:item", ns)
        nav_item = next(i for i in items if "nav" in i.get("properties", "").split())
        spine = [i.get("idref") for i in opf.findall("o:spine/o:itemref", ns)]
        if nav_item.get("id") in spine:
            raise ValueError("EPUB navigation is in the reading spine (duplicate contents)")
        nav = ET.fromstring(epub.read("EPUB/" + nav_item.get("href")))
        toc = next(n for n in nav.findall(".//h:nav", ns)
                   if n.get("{http://www.idpf.org/2007/ops}type") == "toc")
        labels = [" ".join("".join(a.itertext()).split()) for a in toc.findall(".//h:a", ns)]
        if labels != titles:
            raise ValueError("EPUB navigation differs from manuscript reading order")
        ncx = ET.fromstring(epub.read("EPUB/toc.ncx"))
        if len(ncx.findall(".//n:navPoint", ns)) != len(titles):
            raise ValueError("NCX navigation does not contain exactly the manuscript units")
        covers = [i for i in items if "cover-image" in i.get("properties", "").split()]
        if len(covers) != int(has_cover):
            raise ValueError("EPUB cover manifest does not match the supplied cover")
        documents = {name: ET.fromstring(epub.read(name)) for name in epub.namelist()
                     if name.endswith(".xhtml")}
        for name, doc in documents.items():
            for a in doc.findall(".//h:a", ns):
                target = urlsplit(a.get("href", ""))
                if target.scheme or target.netloc:
                    continue
                dest = normpath(join(dirname(name), unquote(target.path))) if target.path else name
                if dest not in epub.namelist():
                    raise ValueError(f"Broken EPUB link: {name}: {a.get('href')}")
                if target.fragment and dest in documents:
                    ids = {el.get("id") for el in documents[dest].iter()}
                    if unquote(target.fragment) not in ids:
                        raise ValueError(f"Broken EPUB anchor: {a.get('href')}")
    print(f"EPUB verified: {len(titles)} chapter entries; navigation outside spine; links resolve")


def build_epub(body, output, work, titles, cover):
    source = work / "ebook.md"
    source.write_text((BOOK / "front-matter.md").read_text() + "\n\n" + body)
    css = work / "epub.css"
    default = run("pandoc", "--print-default-data-file=epub.css", capture_output=True).stdout
    css.write_text(default + (ROOT / "book/epub.css").read_text())
    options = ["--to=epub", "--toc-depth=1", f"--css={css}",
               "--epub-title-page=false"]
    if cover:
        options.append(f"--epub-cover-image={cover}")
    # No printed Contents section and no --toc: Kindle uses nav.xhtml/NCX.
    pandoc(source, output, *options)
    check_epub(output, titles, cover is not None)


def check_dedication(pdf):
    import pymupdf
    with pymupdf.open(pdf) as document:
        pages = [page for page in document if "To Caitlin Cox," in page.get_text()]
        if len(pages) != 1:
            raise ValueError("Expected exactly one dedication page for Caitlin Cox")
        page = pages[0]
        expected = "To Caitlin Cox, the sweetest, the cutest, and the smartest daughter."
        if " ".join(page.get_text().split()) != expected:
            raise ValueError("Dedication wording differs from the requested text")
        spans = [span for block in page.get_text("dict")["blocks"]
                 for line in block.get("lines", []) for span in line["spans"]
                 if span["text"].strip()]
        if not spans or any(abs(span["size"] - 12) > 0.1 for span in spans):
            raise ValueError("Dedication is not printed at 12 points")
        print(f"Dedication verified: physical page {page.number + 1}, 12-point type")


def build_print(source, output, titles):
    # Reuse the maintained renderer/stamper, not Edition 3's older copied script.
    spec = importlib.util.spec_from_file_location("print_build", ROOT / "book/build-print.py")
    printing = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(printing)
    html_path = output.with_suffix(".html")
    pandoc(BOOK / "front-matter-print.md", html_path, str(source),
           "--to=html", "--standalone", "--toc", "--toc-depth=1",
           "--variable=title:", f"--metadata=pagetitle:{TITLE}",
           "-H", str(ROOT / "book/print-header.html"),
           "-H", str(BOOK / "print-header.html"))
    html = html_path.read_text()
    # Put the title/copyright pages before contents. The shared cleaner inserts
    # the explicit pre-Preface recto blank needed for physical gutter parity.
    nav = re.search(r'<nav id="TOC".*?</nav>', html, re.S)
    if not nav:
        raise ValueError("Missing print contents")
    html = html[:nav.start()] + html[nav.end():]
    contents = re.sub(r'(<nav[^>]*>)', r'\1<h2>Contents</h2>', nav.group(), count=1)
    html = html.replace('<h1 id="preface">', contents + '\n<h1 id="preface">', 1)
    html = printing.clean_html(html)
    html_path.write_text(html)
    entries = printing.toc_entries(html)
    if [printing.strip_tags(t) for _, t in entries] != titles:
        raise ValueError("Print contents differ from manuscript reading order")
    printing.render(html_path, output)
    printing.check_type_size(output, expected_body=11.5, expected_code=11.5)
    check_dedication(output)
    mapping = printing.build_map(output, titles)
    if set(mapping) != set(titles):
        raise ValueError(f"Missing print headings: {set(titles) - set(mapping)}")
    first, stamped = printing.finish_pagination(output, entries, mapping, page_size=(432, 648))
    if printing.build_map(output, titles) != mapping:
        raise ValueError("Page stamping moved chapter headings")
    printing.check_pagination(output, first, stamped, page_size=(432, 648))
    run("pdfinfo", str(output))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--format", choices=["all", "markdown", "epub", "print"], default="all")
    parser.add_argument("--cover", type=Path, help="Optional companion ebook cover (not a print wrap)")
    args = parser.parse_args()
    cover = args.cover.resolve() if args.cover else None
    if cover and not cover.is_file():
        parser.error(f"Cover does not exist: {cover}")
    body, titles = assemble()
    output = BOOK / "build"
    output.mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="publication-", dir=output) as temp:
        work = Path(temp)
        source = work / "manuscript.md"
        source.write_text(body)
        if args.format in ("all", "markdown"):
            build_markdown(body, output / "companion.md", work)
        if args.format in ("all", "epub"):
            build_epub(body, output / f"{STEM}.epub", work, titles, cover)
        if args.format in ("all", "print"):
            build_print(source, output / f"{STEM} - Print Interior.pdf", titles)
    print(f"Built {args.format}: {output}")
    if args.format in ("all", "epub") and not cover:
        print("No ebook cover supplied; choose a companion cover before KDP publication.")


if __name__ == "__main__":
    main()
