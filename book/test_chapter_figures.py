"""Keep descriptive image alternatives without duplicate printed captions."""
from html.parser import HTMLParser
from pathlib import Path
import subprocess
import unittest


class FigureMarkup(HTMLParser):
    def __init__(self):
        super().__init__()
        self.images = []
        self.captions = 0

    def handle_starttag(self, tag, attrs):
        if tag == 'img':
            self.images.append(dict(attrs))
        if tag == 'figcaption':
            self.captions += 1


class ChapterFigureTests(unittest.TestCase):
    def test_explicit_captions_preserve_image_alternatives(self):
        chapter = Path(__file__).with_name('chapter-29.md')
        html = subprocess.check_output(
            ['pandoc', '--from=markdown', '--to=html', str(chapter)], text=True)
        parsed = FigureMarkup()
        parsed.feed(html)
        self.assertEqual(len(parsed.images), 6)
        self.assertEqual(parsed.captions, 0)
        for image in parsed.images:
            self.assertGreater(len(image.get('alt', '')), 30)
            self.assertIn('display: block', image.get('style', ''))
            self.assertTrue((chapter.parent / image['src']).is_file())


if __name__ == '__main__':
    unittest.main()
