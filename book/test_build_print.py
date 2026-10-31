"""Print regression checks: python3 -m unittest discover -s book -p 'test_build_print.py'."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import pymupdf

spec = importlib.util.spec_from_file_location("build_print", Path(__file__).with_name("build-print.py"))
build = importlib.util.module_from_spec(spec)
spec.loader.exec_module(build)


class PrintBuildTests(unittest.TestCase):
    def test_clean_html_and_wrapped_toc_anchor(self):
        html = '''<nav id="TOC"><ul>
<li><a href="#contents">Contents</a></li>
<li><a\nhref="#preface">Preface</a></li>
</ul></nav><div class="front-matter">Title</div>
<h1 id="contents">Contents</h1><p>Duplicate markdown contents</p>
<h1 id="preface">Preface</h1><p>Actual prose</p>'''
        clean = build.clean_html(html)
        self.assertEqual(build.toc_entries(clean), [("preface", "Preface")])
        self.assertNotIn('id="contents"', clean)
        self.assertNotIn('Duplicate markdown contents', clean)
        self.assertNotIn('page-break-after: always;', clean)
        self.assertNotIn('class="recto-spacer"', clean)
        self.assertIn('Actual prose', clean)

    def test_measured_recto_spacer_preserves_already_correct_layout(self):
        html = '<h1 id="preface">Preface</h1><p>Unchanged prose</p>'
        self.assertEqual(build.recto_preface_html(html, 5), html)
        repaired = build.recto_preface_html(html, 4)
        self.assertEqual(repaired.count('class="recto-spacer"'), 1)
        self.assertTrue(repaired.endswith(html))
        with self.assertRaises(SystemExit):
            build.recto_preface_html(repaired, 4)

    def test_missing_preface_is_rejected(self):
        with self.assertRaises(SystemExit):
            build.clean_html('<nav id="TOC"></nav>')

    def test_toc_parser_requires_every_item(self):
        with self.assertRaises(SystemExit):
            build.toc_entries('<nav id="TOC"><ul><li>Missing anchor</li></ul></nav>')

    def test_heading_map_skips_contents(self):
        pages = ['Book title\nPreface\nChapter 0', 'Title page', 'Dedication', '',
                 'Preface\nOpening words', 'Chapter 0\nOpening words']
        with patch.object(build, 'page_texts', return_value=pages):
            self.assertEqual(build.build_map(Path('unused'), ['Preface', 'Chapter 0']),
                             {'Preface': 5, 'Chapter 0': 6})

    def test_margins_follow_physical_page_parity(self):
        self.assertEqual(build.margins(0), (63, 468))
        self.assertEqual(build.margins(1), (36, 441))
        self.assertEqual(build.margins(4), (63, 468))

    @unittest.skipUnless(build.GEORGIA.is_file(), 'requires macOS Georgia font')
    def test_pdf_stamping_labels_black_text_and_wrapped_link(self):
        with tempfile.TemporaryDirectory() as tmp:
            pdf = Path(tmp) / 'fixture.pdf'
            doc = pymupdf.open()
            for _ in range(5):
                doc.new_page(width=504, height=720)
            page = doc[0]
            page.insert_text((63, 100), 'Preface', fontsize=10.5)
            page.insert_text((63, 115), 'A wrapped title', fontsize=10.5)
            doc[4].insert_text((63, 100), 'Preface', fontsize=18)
            target = doc[4].xref
            doc.xref_set_key(doc.pdf_catalog(), 'Names',
                             f'<< /Dests << /Names [(preface) [{target} 0 R /XYZ 0 666 0]] >> >>')
            for rect in [pymupdf.Rect(63, 88, 115, 103), pymupdf.Rect(63, 103, 150, 118)]:
                page.insert_link({'kind': pymupdf.LINK_GOTO, 'page': 4, 'from': rect})
            page = doc.reload_page(page)
            for link in page.get_links():
                doc.xref_set_key(link['xref'], 'A', '<< /S /GoTo /D (preface) >>')
            doc.save(pdf)
            doc.close()
            first, stamps = build.finish_pagination(pdf, [('preface', 'Preface')], {'Preface': 5})
            self.assertEqual(first, 5)
            self.assertEqual(len(stamps), 1)
            build.check_pagination(pdf, first, stamps)
            with pymupdf.open(pdf) as result:
                self.assertEqual(result.resolve_link('#preface')[0], 4)
                self.assertEqual(result[0].get_label(), '')
                self.assertEqual(result[4].get_label(), '1')
                # A gray quote must fail the final printed-color check.
                result[4].insert_text((63, 140), 'Gray quote', color=(.4, .4, .4))
                result.saveIncr()
            with self.assertRaises(SystemExit):
                build.check_pagination(pdf, first, stamps)

    def test_verso_preface_is_rejected(self):
        with self.assertRaises(SystemExit):
            build.finish_pagination(Path('unused'), [('preface', 'Preface')], {'Preface': 4})


if __name__ == '__main__':
    unittest.main()
