"""Edition configuration regressions; rendering checks live in the reused builder."""
import importlib.util
from pathlib import Path
import tempfile
import unittest

spec = importlib.util.spec_from_file_location("csharp_publication", Path(__file__).with_name("build-publication.py"))
build = importlib.util.module_from_spec(spec)
spec.loader.exec_module(build)


class EditionTests(unittest.TestCase):
    def test_manuscript_and_front_matter(self):
        body, titles = build.assemble()
        self.assertEqual(len(titles), 28)
        self.assertEqual(titles[0], "Preface")
        self.assertTrue(titles[-1].startswith("Epilogue:"))
        for n, title in enumerate(titles[1:-1]):
            self.assertRegex(title, rf"^Chapter {n}(?::| —)")
        self.assertNotIn("chapter-26.md", build.publication.PARTS)
        self.assertEqual((build.BOOK / "preface.md").read_bytes(),
                         (build.BOOK.parents[1] / "edition-2/book/companion/preface.md").read_bytes())
        front = (build.BOOK / "front-matter-print.md").read_text()
        self.assertIn("C# Edition", front)
        self.assertIn("Book 6", front)
        self.assertIn("To Caitlin Cox,", front)

    def test_missing_chapter_fails(self):
        with tempfile.TemporaryDirectory() as temp:
            with self.assertRaises(FileNotFoundError):
                build.assemble(Path(temp))

    def test_edition_link_base(self):
        with tempfile.TemporaryDirectory() as temp:
            source, output = Path(temp) / "source.md", Path(temp) / "output.html"
            source.write_text("[codebook](../chapter-00.md) [local](#heading)")
            build.pandoc(source, output, "--to=html")
            html = output.read_text()
            self.assertIn("https://github.com/waywardgeek/ensemble/blob/main/book/chapter-00.md", html)
            self.assertIn('href="#heading"', html)


if __name__ == "__main__":
    unittest.main()
