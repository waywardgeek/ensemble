#!/usr/bin/env python3
"""Two-pass print build: render, learn real page numbers, inject them into the TOC.

Chrome renders @page margin boxes (so footer page numbers work) but does NOT
support target-counter(), so a TOC cannot learn its own page references in one
pass. This renders once, reads where each chapter actually landed, injects those
numbers, renders again, then re-reads to prove pagination did not shift.
"""
import re, subprocess, sys, shutil
from pathlib import Path

CHROME = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
BOOK = Path.home() / "projects/ensemble/book"


def render(html: Path, pdf: Path) -> None:
    pdf.unlink(missing_ok=True)
    subprocess.run(
        [CHROME, "--headless=new", "--no-sandbox", f"--print-to-pdf={pdf}",
         "--no-pdf-header-footer", f"file://{html}"],
        capture_output=True, timeout=600,
    )
    if not pdf.exists():
        sys.exit(f"FATAL: Chrome produced no PDF at {pdf}")


def page_texts(pdf: Path) -> list[str]:
    out = subprocess.run(["pdftotext", str(pdf), "-"],
                         capture_output=True, text=True, timeout=300).stdout
    return out.split("\f")


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


def inject(html: str, entries: list[tuple[str, str]], mapping: dict[str, int]) -> str:
    nav = re.search(r'<nav id="TOC".*?</nav>', html, re.S)
    block = nav.group(0)
    new = block
    done = 0
    for href, raw in entries:
        title = strip_tags(raw)
        if title not in mapping:
            continue
        # Pandoc wraps long tags, so the anchor can read '<a\nhref="#id"'.
        # Matching a literal '<a href=' silently skipped every wrapped entry
        # and still produced a plausible looking TOC.
        pat = re.compile(r'<a\s[^>]*href="#' + re.escape(href) + r'"[^>]*>.*?</a>', re.S)
        m = pat.search(new)
        if not m:
            sys.exit(f"FATAL: no anchor found for #{href}")
        end = m.end()
        tail = new[end:end + 80]
        prev = re.match(r'\s*<span class="tocpg">\d+</span>', tail)
        if prev:  # drop the number a previous pass wrote
            end += prev.end()
        new = new[:end] + f'<span class="tocpg">{mapping[title]}</span>' + new[end:]
        done += 1
    if done != len(entries):
        sys.exit(f"FATAL: injected {done} numbers for {len(entries)} TOC entries")
    return html.replace(block, new)


def main() -> None:
    src = BOOK / "interior-print.html"
    pdf = BOOK / "The Art of Building AI Coding Agents - Print Interior.pdf"
    html = src.read_text()

    entries = toc_entries(html)
    titles = [strip_tags(t) for _, t in entries]
    print(f"TOC entries: {len(titles)}")

    print("pass 1: render")
    render(src, pdf)
    m1 = build_map(pdf, titles)
    print(f"  matched {len(m1)}/{len(titles)} headings")
    missing = [t for t in titles if t not in m1]
    if missing:
        sys.exit(f"FATAL: {len(missing)} headings never matched a page: {missing[:5]}")

    print("pass 2: inject + render")
    src.write_text(inject(html, entries, m1))
    render(src, pdf)

    m2 = build_map(pdf, titles)
    shifted = {t: (m1.get(t), m2.get(t)) for t in titles if m1.get(t) != m2.get(t)}
    if shifted:
        print(f"  pagination SHIFTED for {len(shifted)}; re-injecting with pass-2 map")
        src.write_text(inject(src.read_text(), entries, m2))
        render(src, pdf)
        m3 = build_map(pdf, titles)
        still = {t for t in titles if m2.get(t) != m3.get(t)}
        print("  CONVERGED" if not still else f"  STILL SHIFTING: {len(still)}")
    else:
        print("  stable: pagination identical across passes")

    n = subprocess.run(["pdfinfo", str(pdf)], capture_output=True, text=True).stdout
    print("\n" + "\n".join(l for l in n.splitlines() if l.startswith(("Pages", "Page size"))))


if __name__ == "__main__":
    main()
