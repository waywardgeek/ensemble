"""Focused publication regressions; no KDP access or production artifact writes."""
import base64
import importlib.util
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
import pymupdf
from zipfile import ZipFile

spec = importlib.util.spec_from_file_location('publication', Path(__file__).with_name('build-publication.py'))
build = importlib.util.module_from_spec(spec)
spec.loader.exec_module(build)


class AssemblyTests(unittest.TestCase):
    def test_dedication_wording_and_size(self):
        with tempfile.TemporaryDirectory() as tmp:
            pdf = Path(tmp) / 'dedication.pdf'
            for size, name in [(12, 'Caitlin Cox'), (11.5, 'Caitlin Cox'), (12, 'Someone Else')]:
                with pymupdf.open() as doc:
                    page = doc.new_page(width=432, height=648)
                    page.insert_text((63, 200), f'To {name},', fontsize=size)
                    page.insert_text((63, 225), 'the sweetest, the cutest, and the smartest daughter.', fontsize=size)
                    doc.save(pdf)
                if size == 12 and name == 'Caitlin Cox':
                    build.check_dedication(pdf)
                else:
                    with self.assertRaises(ValueError):
                        build.check_dedication(pdf)

    def test_complete_manuscript_order(self):
        body, titles = build.assemble()
        self.assertEqual(len(titles), 30)
        self.assertEqual(titles[0], 'Preface')
        self.assertTrue(titles[-1].startswith('Epilogue:'))
        for n, title in enumerate(titles[1:-1]):
            self.assertTrue(title.startswith(f'Chapter {n}:'))
        self.assertNotIn('Editorial records and completion status', body)

    def test_missing_and_duplicate_chapters_fail(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory = Path(tmp)
            for name in build.PARTS:
                (directory / name).write_text(f'# {name}\n\nProse.\n')
            (directory / 'chapter-12.md').unlink()
            with self.assertRaises(FileNotFoundError):
                build.assemble(directory)
            (directory / 'chapter-12.md').write_text('# First\n\n# Second\n')
            with self.assertRaises(ValueError):
                build.assemble(directory)


@unittest.skipUnless(shutil.which('pandoc'), 'requires pandoc')
class PandocTests(unittest.TestCase):
    def test_shared_links_default_and_companion_base(self):
        text = '[source](../docs/test.md#section) [local](#heading) [web](https://example.com/)'
        command = ['pandoc', '-f', 'markdown', '-t', 'html', '--lua-filter', str(build.ROOT / 'scripts/publication-links.lua')]
        original = subprocess.run(command, input=text, text=True, capture_output=True, check=True).stdout
        companion = subprocess.run(command + ['-M', 'publication-directory=book/companion'],
                                   input=text, text=True, capture_output=True, check=True).stdout
        self.assertIn('https://github.com/waywardgeek/ensemble/blob/main/docs/test.md#section', original)
        self.assertIn('https://github.com/waywardgeek/ensemble/blob/main/book/docs/test.md#section', companion)
        self.assertIn('href="#heading"', companion)
        self.assertIn('href="https://example.com/"', companion)

    def test_epub_cover_navigation_and_broken_link_rejection(self):
        titles = ['Preface', "Chapter 0: The World's Tools", 'Epilogue: End']
        body = '\n\n'.join('# ' + title + '\n\nProse.' for title in titles)
        with tempfile.TemporaryDirectory() as tmp:
            work = Path(tmp)
            cover = work / 'test-cover.png'
            # A tiny valid packaging fixture, not proposed publication artwork.
            cover.write_bytes(base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAIAAAADCAIAAAA2iEnWAAAAFElEQVR4nGP8//8/AwMDEwMYQCkAORsDA3CXx5UAAAAASUVORK5CYII='))
            output = work / 'test.epub'
            for image in (None, cover):
                with self.subTest(cover=image is not None):
                    build.build_epub(body, output, work, titles, image)
                    with self.assertRaises(ValueError):
                        build.check_epub(output, list(reversed(titles)), image is not None)
            # Inserting nav into the spine must fail, even with a valid TOC.
            bad = work / 'duplicate-nav.epub'
            with ZipFile(output) as src, ZipFile(bad, 'w') as dest:
                for item in src.infolist():
                    data = src.read(item.filename)
                    if item.filename == 'EPUB/content.opf':
                        data = data.replace(b'</spine>', b'<itemref idref="nav" /></spine>')
                    dest.writestr(item, data)
            with self.assertRaisesRegex(ValueError, 'duplicate contents'):
                build.check_epub(bad, titles, True)
            with self.assertRaisesRegex(ValueError, 'Broken EPUB anchor'):
                build.build_epub(body + '\n\n[broken](#no-such-heading)\n', output, work, titles, None)


if __name__ == '__main__':
    unittest.main()
