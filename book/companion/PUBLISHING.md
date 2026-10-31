# Publishing the C# human companion

**The Art of Building AI Coding Agents: A Human Companion (C# Edition)**,
by Bill Cox. Series: **The Singularity as it Happened — Book 6**.

## Build locally

From the repository root:

```sh
python3 book/companion/build-publication.py
```

The thin edition wrapper reuses `book/companion/build-publication.py`, the shared
`book/build-print.py`, CSS, and contents/link helpers. It assembles exactly the
unchanged approved preface, Chapters **00–25**, and epilogue, in that order.
There is **no companion Chapter 26**. Source chapters and the old companion's
assembled manuscript are not rewritten. All outputs stay in this directory's
`build/` subdirectory.

Requirements match the existing companion workflow: macOS, Google Chrome in
`/Applications`, Georgia/Menlo fonts, Pandoc, Poppler (`pdftotext`, `pdfinfo`),
and Python with PyMuPDF. Chrome runs headlessly with a fresh temporary profile;
no interactive browser session is reused.

Use `--format markdown`, `--format epub`, or `--format print` for individual
formats; `print` builds both PDFs. `--cover /path/to/image.png` overrides the
approved local `cover.png` for the EPUB and reading PDF only.

## Generated files

All paths below are relative to `book/companion/`:

| Purpose | File | Current pages |
| --- | --- | --- |
| Assembled Markdown | `build/companion.md` | Reflowable |
| Kindle manuscript | `build/The Art of Building AI Coding Agents - A Human Companion - CSharp Edition.epub` | Reflowable |
| KDP print interior, **no cover** | `build/The Art of Building AI Coding Agents - A Human Companion - CSharp Edition - Print Interior.pdf` | **180** |
| Website reading PDF, cover + interior | `build/The Art of Building AI Coding Agents - A Human Companion - CSharp Edition.pdf` | **181** |

The print interior is **6 × 9 inches**, no bleed, black text, 11.5-point Georgia
body and 11.5-point Menlo code. Inside/outer margins are 0.875/0.5 inches;
top/bottom margins are 0.75 inches. The dedication to Caitlin Cox is unchanged
and 12 point, on physical page 3. Preface starts at physical page 7, printed
page 1 (physical page 8 in the reading PDF). Front matter is unnumbered.

`cover.png` is the unchanged existing human-companion website artwork, reused
with user authorization from
`/Users/bill/projects/coderhapsody/web/website/public/the-art-of-building-ai-coding-agents-a-human-companion.png`.
It is embedded in the EPUB and prepended to the reading PDF. The artwork is not
new C#-specific artwork; the interior title page and EPUB title identify the
C# edition. This front-cover image is **not a paperback wrap**.

## Checks and publication boundary

```sh
python3 -m unittest discover -s book/companion -p 'test_*.py'
python3 -m unittest discover -s book/companion -p 'test_*.py'
```

The seven existing and three edition-specific tests pass. The build verifies
28 ordered EPUB navigation entries, navigation outside the reading spine,
cover packaging and internal links; printed type sizes and dedication;
28 PDF contents destinations/references, black text, and page labels. The
reading-PDF check reopens the saved artifact and verifies all 49 copied links,
coordinates, cover-offset destinations, and labels. Saved PDFs were also
checked for 432 × 648-point geometry on every page, embedded Georgia/Menlo
fonts, and no text outside page bounds. External source links resolve to public
repository URLs using `book/companion` as their source base; this is
structural link checking, not a network availability check.

Before KDP publication, inspect the actual files in Kindle Previewer and KDP's
Print Previewer and order a physical proof. Generate the separate paperback
wrap using KDP's template, the final processed page count, and selected paper.
No ISBN or spine width has been invented. Local checks are not KDP acceptance.

This workflow only builds local files. It does not upload, deploy, publish,
commit, or push them. The website PDF has not been installed on the website.
