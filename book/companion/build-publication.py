#!/usr/bin/env python3
"""Publish the C# human companion using the maintained companion workflow."""
import argparse
import importlib.util
import tempfile
from pathlib import Path

BOOK = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location(
    "companion_publication", BOOK.parents[1] / "scripts/companion-publication.py")
publication = importlib.util.module_from_spec(spec)
spec.loader.exec_module(publication)
publication.BOOK = BOOK
publication.TITLE = "The Art of Building AI Coding Agents: A Human Companion (C# Edition)"
publication.STEM = "The Art of Building AI Coding Agents - A Human Companion - CSharp Edition"
publication.PARTS = ["preface.md", *(f"chapter-{n:02}.md" for n in range(26)), "epilogue.md"]

# The original function's default directory was bound when it was imported.
_assemble = publication.assemble

def assemble(directory=BOOK):
    return _assemble(directory)


def pandoc(source, output, *options):
    publication.run(
        "pandoc", str(source), "--from=markdown-smart", "-o", str(output),
        "--lua-filter", str(publication.ROOT / "scripts/publication-links.lua"),
        "--metadata=publication-directory:book/companion",
        f"--metadata=title:{publication.TITLE}", "--metadata=author:Bill Cox",
        f"--resource-path={BOOK}", *options)


publication.assemble = assemble
publication.pandoc = pandoc

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--format", choices=["all", "markdown", "epub", "print"], default="all")
    parser.add_argument("--cover", type=Path, default=BOOK / "cover.png",
                        help="Approved companion cover for EPUB and reading PDF, never the print interior")
    args = parser.parse_args()
    cover = args.cover.resolve()
    if not cover.is_file():
        parser.error(f"Cover does not exist: {cover}")
    body, titles = assemble()
    output = BOOK / "build"
    output.mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="publication-", dir=output) as temp:
        work = Path(temp)
        source = work / "manuscript.md"
        source.write_text(body)
        if args.format in ("all", "markdown"):
            publication.build_markdown(body, output / "companion.md", work)
        if args.format in ("all", "epub"):
            publication.build_epub(body, output / f"{publication.STEM}.epub", work, titles, cover)
        if args.format in ("all", "print"):
            interior = output / f"{publication.STEM} - Print Interior.pdf"
            publication.build_print(source, interior, titles)
            spec = importlib.util.spec_from_file_location(
                "reading_print", publication.ROOT / "book/build-print.py")
            printing = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(printing)
            # The shared cover compositor uses fixed names under BOOK. Isolate
            # those temporary names, then move only the finished reading PDF.
            printing.BOOK = work
            (work / "cover.png").symlink_to(cover)
            printing.build_reading_pdf(interior)
            (work / "The Art of Building AI Coding Agents.pdf").replace(
                output / f"{publication.STEM}.pdf")
    print(f"Built {args.format}: {output}")


if __name__ == "__main__":
    main()
