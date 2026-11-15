"""Exercise the ebook builder's image guard with a real vector figure."""
import struct
import zlib
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
import zipfile


class EbookTests(unittest.TestCase):
    def test_vector_figure_is_counted_and_preserved(self):
        book = Path(__file__).resolve().parent
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'front-matter.md').write_text('Fixture book\n')
            (root / 'the-art-of-building-ai-coding-agents.md').write_text(
                '# Contents\n\nTemporary contents\n\n# Chapter 0: Fixture\n\n'
                '![A vector figure](diagram.svg)\n')
            svg = b'<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40"><rect width="40" height="40" fill="blue"/></svg>'
            (root / 'diagram.svg').write_bytes(svg)
            # A complete tiny PNG keeps this an offline packaging test.
            def chunk(kind, data):
                checksum = zlib.crc32(kind + data)
                return struct.pack('!I', len(data)) + kind + data + struct.pack('!I', checksum)
            header = struct.pack('!IIBBBBB', 1, 1, 8, 2, 0, 0, 0)
            cover = (b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', header)
                     + chunk(b'IDAT', zlib.compress(b'\x00\xff\xff\xff'))
                     + chunk(b'IEND', b''))
            (root / 'cover.png').write_bytes(cover)
            (root / 'epub.css').write_bytes((book / 'epub.css').read_bytes())
            environment = os.environ.copy()
            environment.update(BOOK_DIR=str(root), BOOK_OUTPUT_DIR=str(root))
            result = subprocess.run(
                [str(book.parent / 'scripts/build-ebook.sh')], env=environment,
                capture_output=True, text=True)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            with zipfile.ZipFile(root / 'The Art of Building AI Coding Agents.epub') as archive:
                vectors = [name for name in archive.namelist() if name.endswith('.svg')]
                self.assertEqual(len(vectors), 1)
                self.assertEqual(archive.read(vectors[0]), svg)
                self.assertIsNone(archive.testzip())
