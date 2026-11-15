#!/usr/bin/env python3
"""Render only Chapter 0; verify a self-contained, readable printed map page.

Requires existing pandoc, Python PyMuPDF, Node, Chrome and Playwright. This never
replaces release PDFs/EPUBs or installs tooling. Outputs remain in ignored .build.
"""
import argparse
from collections import Counter
import html
import json
from pathlib import Path
import re
import subprocess
import xml.etree.ElementTree as ET
import zipfile

import pymupdf
import generate

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("playwright", help="Absolute path to an existing Playwright module")
parser.add_argument("--node", default="node")
args = parser.parse_args()
tool = Path(__file__).resolve().parent
repo = tool.parents[1]
book = repo / "book"
output = tool / ".build"
output.mkdir(exist_ok=True)
relationships = json.loads((tool / "relationships.json").read_text())["relationships"]
expected = {e[key] for e in relationships for key in ("owner", "child")}
chapter = (book / "chapter-00.md").read_text()
guide = chapter.split("<!-- common-ownership-classes:start -->", 1)[1].split("<!-- common-ownership-classes:end -->", 1)[0]
listed = re.findall(r"^- \*\*([^*]+)\*\*: .+$", guide, re.MULTILINE)
assert set(listed) == expected and len(listed) == len(expected), "Every map type needs exactly one class description"

# Raw HTML shared with print must also be valid XHTML in EPUB. In particular,
# a browser-accepted <br> is not sufficient: use the XML-closed <br /> form.
epub = output / "chapter-00-proof.epub"
subprocess.run(["pandoc", str(book / "chapter-00.md"), "--from=markdown", "--to=epub3",
                "--resource-path=" + str(book), "-M", "title=Chapter 0 diagram proof",
                "-o", str(epub)], check=True)
with zipfile.ZipFile(epub) as archive:
    assert archive.testzip() is None
    images = [name for name in archive.namelist() if name.endswith(".svg")]
    assert len(images) == 3
    actual_images = {archive.read(name) for name in images}
    assert actual_images == {(book / f"figures/common-{key}.svg").read_bytes()
                             for key in ("ownership", "context", "contents")}
    for name in archive.namelist():
        if name.endswith((".xhtml", ".opf", ".svg")):
            ET.fromstring(archive.read(name))

namespace = {"s": "http://www.w3.org/2000/svg"}
figure = ET.parse(book / "figures/common-ownership.svg")
node_names = {node.find("s:title", namespace).text for node in figure.findall('.//s:g[@class="node"]', namespace)}
assert node_names == expected
assert len(figure.findall('.//s:g[@class="edge"]', namespace)) == len(relationships)

# A base URI resolves the real figure while keeping generated proof files out
# of book/. Use the actual print header so page-break and type checks are real.
base = output / "proof-base.html"
base.write_text('<base href="' + html.escape(book.as_uri() + "/", quote=True) + '">\n')
page = output / "chapter-00-proof.html"
pdf = output / "chapter-00-proof.pdf"
subprocess.run(["pandoc", str(book / "chapter-00.md"), "--from=markdown", "--to=html", "--standalone",
                "-M", "title=Chapter 0 print proof", "-H", str(book / "print-header.html"),
                "-H", str(base), "-o", str(page)], check=True)
subprocess.run([args.node, str(repo / "scripts/print-book.cjs"), args.playwright,
                "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", str(page), str(pdf)], check=True)
with pymupdf.open(pdf) as document:
    assert all(abs(p.rect.width - 504) < .1 and abs(p.rect.height - 720) < .1 for p in document)
    map_pages = [p for p in document if "Ensemble’s ownership map" in p.get_text() or "Ensemble's ownership map" in p.get_text()]
    assert len(map_pages) == 1, "The map must appear on one dedicated page"
    map_page = map_pages[0]
    text = map_page.get_text()
    assert "The names on the map" not in text, "Class guide must follow on its own page"
    assert "Parallel Agent-to-question" in text, "Legend must remain on the diagram page"
    spans = [span for block in map_page.get_text("dict")["blocks"] for line in block.get("lines", []) for span in line["spans"] if span["text"].strip()]
    labels = [span for span in spans if "Helvetica" in span["font"] or "Arial" in span["font"]]
    assert labels, "Diagram labels must remain measurable vector text"
    minimum = min(span["size"] for span in labels)
    assert minimum >= 9.9, f"Diagram labels shrank below 10pt: {minimum}"
    compact = re.sub(r"\s+", "", text)
    for name in expected:
        displayed = "WorkProductTransfer" if name == "Agent.WorkProductTransfer" else name
        assert displayed in compact, f"Missing printed label: {name}"
    for span in spans:
        x0, y0, x1, y1 = span["bbox"]
        assert x0 >= 35 and x1 <= 469 and y0 >= 53 and y1 <= 673, f"Text outside print area: {span['text']}"
    map_page.get_pixmap(matrix=pymupdf.Matrix(2, 2)).save(output / "ownership-print-proof.png")
    print(f"PASS: EPUB embeds exact SVG with valid XML; {len(expected)} class bullets match the map; {len(relationships)} edges; dedicated 7x10 page {map_page.number + 1}; minimum label {minimum:.2f}pt; labels and legend remain inside margins. Proof: {pdf}")

# Supplemental sheets derive their expected node and edge sets from the source
# inventory, not from whatever happened to render successfully.
inventory = json.loads((repo / "docs/common-diagrams/inventory.json").read_text())
reviewed = json.loads((tool / "relationships.json").read_text())
types, edges = generate.build_graph(inventory, reviewed)
views = json.loads((tool / "context-views.json").read_text())
for key, view in views.items():
    names = set(view["types"])
    guide = chapter.split(f"<!-- common-{key}-classes:start -->", 1)[1].split(f"<!-- common-{key}-classes:end -->", 1)[0]
    bullets = re.findall(r"^- \*\*([^*]+)\*\*: .+$", guide, re.MULTILINE)
    assert set(bullets) == names and len(bullets) == len(names)
    figure = ET.parse(book / f"figures/common-{key}.svg")
    nodes = {n.find("s:title", namespace).text for n in figure.findall('.//s:g[@class="node"]', namespace)}
    assert nodes == names
    expected_edges = [(e["owner"], e["child"]) for e in edges if e["owner"] in names and e["child"] in names]
    if view["inheritance"]:
        expected_edges.extend((name, base) for name in names for base in types[name]["bases"] if base in names)
    actual_edges = [tuple(n.find("s:title", namespace).text.split("->")) for n in figure.findall('.//s:g[@class="edge"]', namespace)]
    assert Counter(actual_edges) == Counter(expected_edges), key
    with pymupdf.open(pdf) as document:
        pages = [p for p in document if view["title"] in p.get_text()]
        assert len(pages) == 1, f"Dedicated {key} page"
        page = pages[0]
        text = page.get_text()
        assert "The names on" not in text
        legend = "Arrows describe stored relationships" if key == "context" else "Metadata boxes represent recorded facts"
        assert legend in text
        compact = re.sub(r"\s+", "", text)
        for name in names:
            assert name in compact, f"Missing {key} printed label: {name}"
        spans = [s for b in page.get_text("dict")["blocks"] for l in b.get("lines", []) for s in l["spans"] if s["text"].strip()]
        labels = [s for s in spans if "Helvetica" in s["font"] or "Arial" in s["font"]]
        minimum = min(s["size"] for s in labels)
        assert minimum >= 9.9, (key, minimum)
        for span in spans:
            x0, y0, x1, y1 = span["bbox"]
            assert x0 >= 35 and x1 <= 469 and y0 >= 53 and y1 <= 673, (key, span["text"], span["bbox"])
        page.get_pixmap(matrix=pymupdf.Matrix(2, 2)).save(output / f"{key}-print-proof.png")
        print(f"PASS: {key}: {len(names)} types and matching bullets; {len(expected_edges)} arrows; page {page.number + 1}; minimum {minimum:.2f}pt; printed legend and margins intact.")
