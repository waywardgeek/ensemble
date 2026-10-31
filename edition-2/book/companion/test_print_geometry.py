"""Companion trim-size regression for the shared print stamper."""
import importlib.util
from pathlib import Path
import tempfile
import unittest
from unittest.mock import Mock, patch
import pymupdf

spec = importlib.util.spec_from_file_location('print_build', Path(__file__).resolve().parents[1] / 'build-print.py')
build = importlib.util.module_from_spec(spec)
spec.loader.exec_module(build)


class PrintGeometryTests(unittest.TestCase):
    def test_requested_type_size_and_original_default(self):
        page = Mock()
        spans = [{"font": "Georgia", "size": 11.5, "text": "Body prose"},
                 {"font": "Menlo-Regular", "size": 11.5, "text": "code"}]
        page.get_text.return_value = {"blocks": [{"lines": [{"spans": spans}]}]}
        with patch.object(build.pymupdf, "open") as opened:
            opened.return_value.__enter__.return_value = [page]
            build.check_type_size(Path("fixture"), expected_body=11.5, expected_code=11.5)
            with self.assertRaises(SystemExit):
                build.check_type_size(Path("fixture"))
            spans[0]["size"] = 10.5
            with self.assertRaises(SystemExit):
                build.check_type_size(Path("fixture"), expected_body=11.5, expected_code=11.5)
            spans[1]["size"] = 10
            build.check_type_size(Path("fixture"))

    @unittest.skipUnless(build.GEORGIA.is_file(), 'requires macOS Georgia font')
    def test_six_by_nine_stamping_and_size_guard(self):
        self.assertEqual(build.margins(0, 432), (63, 396))
        self.assertEqual(build.margins(1, 432), (36, 369))
        with tempfile.TemporaryDirectory() as tmp:
            pdf = Path(tmp) / 'small.pdf'
            with pymupdf.open() as doc:
                for _ in range(5):
                    doc.new_page(width=432, height=648)
                doc[0].insert_text((63, 100), 'Preface', fontsize=10.5)
                doc[4].insert_text((63, 100), 'Preface', fontsize=18)
                target = doc[4].xref
                doc.xref_set_key(doc.pdf_catalog(), 'Names',
                    f'<< /Dests << /Names [(preface) [{target} 0 R /XYZ 0 594 0]] >> >>')
                page = doc[0]
                page.insert_link({'kind': pymupdf.LINK_GOTO, 'page': 4,
                                  'from': pymupdf.Rect(63, 88, 115, 103)})
                page = doc.reload_page(page)
                doc.xref_set_key(page.get_links()[0]['xref'], 'A', '<< /S /GoTo /D (preface) >>')
                doc.save(pdf)
            entries, mapping = [('preface', 'Preface')], {'Preface': 5}
            with self.assertRaises(SystemExit):
                build.finish_pagination(pdf, entries, mapping)  # original default is still 7 × 10
            first, stamps = build.finish_pagination(pdf, entries, mapping, (432, 648))
            build.check_pagination(pdf, first, stamps, (432, 648))


if __name__ == '__main__':
    unittest.main()
