#!/usr/bin/env python3
"""One-render print build with measured PDF contents references and page labels.

Chrome does not implement target-counter(). Reserve a TOC number column in CSS,
render once, resolve its links, and stamp references without changing layout.
Front matter is unnumbered; the Preface begins Arabic page 1 on a recto page.
"""
import re, subprocess, sys
from collections import Counter
from pathlib import Path

import pymupdf

CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
BOOK = Path(__file__).resolve().parent
GEORGIA = Path("/System/Library/Fonts/Supplemental/Georgia.ttf")


def render(html: Path, pdf: Path) -> None:
    pdf.unlink(missing_ok=True)
    subprocess.run(
        [CHROME, "--headless=new", "--no-sandbox", f"--print-to-pdf={pdf}",
         "--no-pdf-header-footer", f"file://{html}"],
        capture_output=True, timeout=600, check=True,
    )
    if not pdf.exists():
        sys.exit(f"FATAL: Chrome produced no PDF at {pdf}")


def page_texts(pdf: Path) -> list[str]:
    out = subprocess.run(["pdftotext", str(pdf), "-"],
                         capture_output=True, text=True, timeout=300, check=True).stdout
    return out.split("\f")


def check_type_size(pdf: Path, expected_body: float = 10.5, expected_code: float = 10) -> None:
    """Check printed type, since correct CSS does not prevent Chrome shrinking it.

    Count characters rather than spans so short headings and page numbers
    cannot outweigh body prose. Code must retain its intended size everywhere.
    Small rounding differences in Chrome's PDF coordinates are harmless.
    """
    body = Counter()
    code_sizes = []
    with pymupdf.open(pdf) as document:
        for page in document:
            for block in page.get_text("dict")["blocks"]:
                for line in block.get("lines", []):
                    for span in line["spans"]:
                        if not span["text"].strip():
                            continue
                        size = round(span["size"], 2)
                        if span["font"] == "Georgia":
                            body[size] += len(span["text"])
                        if span["font"].startswith("Menlo"):
                            code_sizes.append(size)
    if not body or not code_sizes:
        sys.exit("FATAL: could not measure the expected body and code fonts")
    body_size = body.most_common(1)[0][0]
    code_size = min(code_sizes)
    if abs(body_size - expected_body) > 0.1 or code_size < expected_code - 0.1:
        sys.exit(f"FATAL: printed type is {body_size}pt body / {code_size}pt code; "
                 f"expected {expected_body}pt / {expected_code}pt. Check overflowing code and tables.")
    print(f"  printed type: {body_size}pt body, minimum {code_size}pt code")


def first_line(page: str) -> str:
    """Opening text of a page, normalized.

    A heading wraps across several printed lines, and pandoc wraps the TOC
    title at a different point, so compare normalized prefixes rather than
    single lines.
    """
    lines = [l.strip() for l in page.splitlines() if l.strip()]
    return norm(" ".join(lines[:4]))


def norm(s: str) -> str:
    return re.sub(r"\s+", " ", s).strip()


def toc_entries(html: str) -> list[tuple[str, str]]:
    """Ordered (href_id, title) pairs from pandoc's #TOC nav.

    The anchor pattern must tolerate '<a\\nhref=' because pandoc wraps long
    tags. Assuming a single space here silently dropped the three longest
    chapters, and because the completeness check counted only what this
    function returned, it agreed that nothing was missing.
    """
    nav = re.search(r'<nav id="TOC".*?</nav>', html, re.S)
    if not nav:
        sys.exit("FATAL: no <nav id=\"TOC\"> found. Did pandoc run with --toc?")
    block = nav.group(0)
    found = re.findall(r'<a\s[^>]*href="#([^"]+)"[^>]*>(.*?)</a>', block, re.S)
    want = len(re.findall(r"<li>", block))
    if len(found) != want:
        sys.exit(f"FATAL: {want} TOC items but only {len(found)} anchors parsed")
    return found


def strip_tags(s: str) -> str:
    return norm(re.sub(r"<[^>]+>", "", s))


def build_map(pdf: Path, titles: list[str]) -> dict[str, int]:
    """Walk pages in order, matching each expected title to the page it opens.

    Chapters are ordered and each starts a page (h1 page-break-before: always),
    so an ordered walk avoids matching a title where it appears inside the TOC.
    """
    pages = page_texts(pdf)
    found, idx = {}, 0
    for pageno, text in enumerate(pages, start=1):
        if idx >= len(titles):
            break
        want = titles[idx]
        if first_line(text).startswith(want):
            found[want] = pageno
            idx += 1
    return found


def margins(index: int, page_width: float = 504) -> tuple[float, float]:
    """Physical page parity, never the reset printed counter, owns the gutter."""
    return (63, page_width - 36) if index % 2 == 0 else (36, page_width - 63)


def finish_pagination(pdf: Path, entries: list[tuple[str, str]],
                      mapping: dict[str, int],
                      page_size: tuple[float, float] = (504, 720)) -> tuple[int, list]:
    """Stamp the reserved TOC column and footers; do not reflow any content."""
    page_width, page_height = page_size
    first = mapping.get("Preface")
    if first is None or entries[0][0] != "preface":
        sys.exit("FATAL: Preface must be the first numbered contents entry")
    if first % 2 != 1:
        sys.exit("FATAL: Preface must begin on a physical right-hand page")
    if not GEORGIA.is_file():
        sys.exit(f"FATAL: missing pagination font {GEORGIA}")
    font = pymupdf.Font(fontfile=str(GEORGIA))
    stamped = []
    pending = pdf.with_suffix(".numbered.pdf")
    with pymupdf.open(pdf) as doc:
        if any(abs(page.rect.width - page_width) > 0.01 or
               abs(page.rect.height - page_height) > 0.01 for page in doc):
            sys.exit(f"FATAL: pagination geometry requires "
                     f"{page_width / 72:g} × {page_height / 72:g} inch pages")
        for href, raw in entries:
            title = strip_tags(raw)
            target = mapping[title] - 1
            # Only front-matter links are TOC links. Chrome may produce several
            # rectangles for a wrapped anchor, so stamp once at its last line.
            hits = []
            for i in range(first - 1):
                for link in doc[i].get_links():
                    if link.get("nameddest") == href:
                        resolved, _, _ = doc.resolve_link("#" + href)
                        if resolved != target:
                            sys.exit(f"FATAL: contents link #{href} disagrees with heading")
                        hits.append((i, link["from"]))
            if not hits:
                sys.exit(f"FATAL: no PDF contents link for #{href}")
            if len({i for i, _ in hits}) != 1:
                sys.exit(f"FATAL: contents entry split across pages: {title}")
            i = hits[0][0]
            page = doc[i]
            rects = [rect for _, rect in hits]
            spans = [span for block in page.get_text("dict")["blocks"]
                     for line in block.get("lines", []) for span in line["spans"]
                     if span["text"].strip() and any(
                         (pymupdf.Rect(span["bbox"]) & rect).get_area() > 1
                         for rect in rects)]
            if not spans:
                sys.exit(f"FATAL: cannot locate TOC text baseline for {title}")
            baseline = max(span["origin"][1] for span in spans)
            number = str(mapping[title] - first + 1)
            right = margins(i, page_width)[1]
            width = font.text_length(number, fontsize=10.5)
            if max(span["bbox"][2] for span in spans) > right - width - 8:
                sys.exit(f"FATAL: insufficient reserved TOC number space for {title}")
            page.insert_font(fontname="PrintGeorgia", fontfile=str(GEORGIA))
            page.insert_text((right - width, baseline), number,
                             fontname="PrintGeorgia", fontsize=10.5, color=(0, 0, 0))
            stamped.append((i, pymupdf.Rect(right - 43, baseline - 10.5,
                                          right + 1, baseline + 3), number, target))
        for i in range(first - 1, len(doc)):
            page = doc[i]
            number = str(i - first + 2)
            left, right = margins(i, page_width)
            width = font.text_length(number, fontsize=10)
            page.insert_font(fontname="PrintGeorgia", fontfile=str(GEORGIA))
            page.insert_text(((left + right - width) / 2, page_height - 24), number,
                             fontname="PrintGeorgia", fontsize=10, color=(0, 0, 0))
        doc.set_page_labels([
            {"startpage": 0, "prefix": "", "style": ""},
            {"startpage": first - 1, "prefix": "", "style": "D", "firstpagenum": 1},
        ])
        doc.save(pending, deflate=True)
    pending.replace(pdf)
    return first, stamped


def check_pagination(pdf: Path, first: int, stamped: list,
                     page_size: tuple[float, float] = (504, 720)) -> None:
    page_width, page_height = page_size
    with pymupdf.open(pdf) as doc:
        for i, page in enumerate(doc):
            footer = page.get_text(clip=pymupdf.Rect(0, page_height - 45, page_width, page_height)).strip()
            expected = "" if i < first - 1 else str(i - first + 2)
            if footer != expected or page.get_label() != expected:
                sys.exit(f"FATAL: footer/page label mismatch on physical page {i + 1}")
            for block in page.get_text("dict")["blocks"]:
                for line in block.get("lines", []):
                    for span in line["spans"]:
                        if span["text"].strip() and span["color"] != 0:
                            sys.exit(f"FATAL: non-black print text on page {i + 1}: "
                                     f"{span['text'][:60]}")
        for i, rect, expected, target in stamped:
            actual = doc[i].get_text(clip=rect).strip()
            if actual != expected or doc[target].get_label() != expected:
                sys.exit(f"FATAL: TOC number/destination mismatch on physical page {i + 1}")
    print(f"  verified: {len(stamped)} TOC references, black text, unnumbered front matter")
    print(f"  Preface: physical page {first}, printed page 1")


def clean_html(html: str) -> str:
    """Remove the manuscript's duplicate/self-referential Contents.

    CSS breaks after the generated nav itself. An empty break-after div used to
    create a spurious blank page when the nav already filled its own page.
    """
    # 2. Remove "Contents" entry from pandoc TOC (self-referential)
    html = re.sub(
        r'<li><a\s[^>]*href="#contents"[^>]*>Contents</a>.*?</li>\s*',
        '',
        html,
        count=1,
        flags=re.S,
    )

    # 3. Strip the markdown "# Contents" section from the body.
    #    It starts with <h1 id="contents"> and ends just before the next <h1>.
    html = re.sub(
        r'<h1 id="contents">Contents</h1>.*?(?=<h1[ >])',
        '',
        html,
        count=1,
        flags=re.S,
    )

    # Chrome currently treats break-before:right as an ordinary page break.
    # The title/TOC/dedication occupy three pages; explicitly leave the fourth
    # blank. finish_pagination verifies physical recto parity if layout changes.
    anchor = '<h1 id="preface">'
    if html.count(anchor) != 1:
        sys.exit("FATAL: expected exactly one Preface heading")
    html = html.replace(anchor, '<div class="recto-spacer" aria-hidden="true"></div>\n' + anchor)
    return html


def run_pandoc(src: Path) -> None:
    """Regenerate interior-print.html from the assembled manuscript.

    This used to be a manual step documented in book/README.md, and skipping
    it failed SILENTLY: build-print.py would re-render a stale HTML left over
    from an earlier chapter, producing a freshly-timestamped PDF with an old
    TOC and the newest chapter missing entirely. Nothing in the two-pass build
    noticed, because every heading the stale TOC listed did match a page.
    Running pandoc here makes that class of staleness impossible.
    """
    manuscript = BOOK / "the-art-of-building-ai-coding-agents.md"
    front = BOOK / "front-matter-print.md"
    header = BOOK / "print-header.html"
    for f in (front, manuscript, header):
        if not f.exists():
            sys.exit(f"FATAL: missing build input {f}")
    cmd = [
        "pandoc", front.name, manuscript.name,
        "-o", src.name,
        "--from=markdown", "--to=html", "--standalone",
        "--toc", "--toc-depth=1",
        "-M", "title=The Art of Building AI Coding Agents",
        "-M", "author=Bill Cox",
        "-H", header.name,
        "--lua-filter", str(BOOK.parent / "scripts" / "publication-links.lua"),
        "--lua-filter", str(BOOK.parent / "scripts" / "print-sidebars.lua"),
    ]
    print("pandoc: regenerate interior-print.html")
    r = subprocess.run(cmd, cwd=BOOK, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit(f"FATAL: pandoc failed:\n{r.stderr}")


def build_reading_pdf(interior_path: Path) -> None:
    """Prepend the cover without losing Chrome's named contents destinations.

    Concatenation can retain link annotations while discarding their name tree.
    Resolve those names in the original PDF, then offset explicit page targets.
    resolve_link also converts PDF coordinates to PyMuPDF page coordinates.
    """
    output = BOOK / "The Art of Building AI Coding Agents.pdf"
    with pymupdf.open(interior_path) as interior, pymupdf.open() as reading:
        rect = interior[0].rect
        cover = reading.new_page(width=rect.width, height=rect.height)
        cover.insert_image(cover.rect, filename=str(BOOK / "cover.png"),
                           keep_proportion=True)
        reading.insert_pdf(interior, links=False)
        count = 0
        for index, page in enumerate(interior):
            for original in page.get_links():
                link = {k: v for k, v in original.items()
                        if k not in ("xref", "id", "nameddest")}
                if "nameddest" in original:
                    target, x, y = interior.resolve_link("#" + original["nameddest"])
                    if target < 0:
                        sys.exit(f"FATAL: unresolved contents link {original['nameddest']}")
                    link.update(kind=pymupdf.LINK_GOTO, page=target + 1,
                                to=pymupdf.Point(x, y))
                elif link["kind"] == pymupdf.LINK_GOTO:
                    link["page"] += 1
                reading[index + 1].insert_link(link)
                count += 1
        reading.set_page_labels(
            [{"startpage": 0, "prefix": "", "style": ""}] +
            [dict(rule, startpage=rule["startpage"] + 1)
             for rule in interior.get_page_labels()]
        )
        reading.save(output, deflate=True)
    check_reading_pdf(interior_path, output)
    print(f"  reading PDF: cover plus interior, {count} links preserved")


def check_reading_pdf(interior_pdf: Path, output: Path) -> None:
    """Reopen the saved artifact to verify labels and cover-offset navigation."""
    with pymupdf.open(interior_pdf) as interior, pymupdf.open(output) as reading:
        if len(reading) != len(interior) + 1 or reading[0].get_label() != "":
            sys.exit("FATAL: reading PDF must contain an unnumbered cover plus interior")
        for i, page in enumerate(interior):
            copied = reading[i + 1]
            if copied.get_label() != page.get_label():
                sys.exit(f"FATAL: reading label changed on interior page {i + 1}")
            sources, targets = page.get_links(), copied.get_links()
            if len(sources) != len(targets):
                sys.exit(f"FATAL: reading links lost on interior page {i + 1}")
            for source, target in zip(sources, targets):
                expected = dict(source)
                if source.get("nameddest"):
                    index, x, y = interior.resolve_link("#" + source["nameddest"])
                    expected.update(kind=pymupdf.LINK_GOTO, page=index + 1,
                                    to=pymupdf.Point(x, y))
                elif source.get("kind") == pymupdf.LINK_GOTO:
                    expected["page"] += 1
                for key in ("kind", "page", "uri", "file"):
                    if target.get(key) != expected.get(key):
                        sys.exit(f"FATAL: reading link {key} changed on interior page {i + 1}")
                for key in ("from", "to"):
                    if key in expected and (key not in target or any(
                        abs(a - b) > 0.05 for a, b in zip(expected[key], target[key])
                    )):
                        sys.exit(f"FATAL: reading link position changed on interior page {i + 1}")
    print("  verified saved reading PDF links, coordinates and labels")


def main() -> None:
    src = BOOK / "interior-print.html"
    pdf = BOOK / "The Art of Building AI Coding Agents - Print Interior.pdf"

    run_pandoc(src)
    html = src.read_text()

    print("cleaning HTML: strip duplicate Contents")
    html = clean_html(html)
    src.write_text(html)
    entries = toc_entries(html)
    titles = [strip_tags(t) for _, t in entries]
    print(f"TOC entries: {len(titles)}")

    print("render: Chrome (single layout pass)")
    render(src, pdf)
    check_type_size(pdf)
    mapping = build_map(pdf, titles)
    missing = [t for t in titles if t not in mapping]
    if missing:
        sys.exit(f"FATAL: headings never matched a page: {missing[:5]}")
    first, stamped = finish_pagination(pdf, entries, mapping)
    if build_map(pdf, titles) != mapping:
        sys.exit("FATAL: page stamping changed heading locations")
    check_pagination(pdf, first, stamped)
    check_type_size(pdf)
    n = subprocess.run(["pdfinfo", str(pdf)], capture_output=True, text=True).stdout
    print("\n" + "\n".join(l for l in n.splitlines() if l.startswith(("Pages", "Page size"))))
    build_reading_pdf(pdf)


if __name__ == "__main__":
    main()
