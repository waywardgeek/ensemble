"""Real-filesystem checks for fresh conversations with role-local memory."""
import json
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "skills" / "fiction-workflow"))
from role_homes import MARKER, IMPORT_MARKER, IMPORT_VALUE, prepare_home
from run_state import StateError


class RoleHomesTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.baselines = self.root / "baseline"
        for role in ("author", "editor"):
            source = self.baselines / role
            source.mkdir(parents=True)
            (source / "SOUL.md").write_text("Identity: " + role, encoding="utf-8")
            (source / "MEMORY.md").write_bytes(b"")

    def test_separate_empty_baselines(self):
        for role in ("author", "editor"):
            home = prepare_home(self.root / role, role, self.baselines)
            self.assertEqual((home / "SOUL.md").read_text(), "Identity: " + role)
            self.assertEqual((home / "MEMORY.md").read_bytes(), b"")
            self.assertEqual(set(p.name for p in home.iterdir()),
                             {"SOUL.md", "MEMORY.md", MARKER})
            self.assertEqual((self.baselines / role / "MEMORY.md").read_bytes(), b"")

    def test_transfer_memory_not_conversation(self):
        previous = prepare_home(self.root / "chapter1", "author", self.baselines)
        (previous / "MEMORY.md").write_text("Learned voice lesson", encoding="utf-8")
        for band in ("memory", "bucket-0", "bucket-1"):
            directory = previous / band
            directory.mkdir()
            (directory / "active.md").write_text(band, encoding="utf-8")
            (directory / "old.md.retired").write_text("retired", encoding="utf-8")
        for name in ("save.json", "settings.json", "submission.json"):
            (previous / name).write_text("must not transfer", encoding="utf-8")
        home = prepare_home(self.root / "chapter2", "author", self.baselines, previous)
        self.assertEqual((home / "MEMORY.md").read_text(), "Learned voice lesson")
        for band in ("memory", "bucket-0", "bucket-1"):
            self.assertEqual((home / band / "active.md").read_text(), band)
            self.assertFalse((home / band / "old.md.retired").exists())
        self.assertFalse((home / "save.json").exists())
        self.assertFalse((home / "settings.json").exists())
        self.assertFalse((home / "submission.json").exists())
        self.assertTrue((previous / "save.json").exists())

    def test_existing_home_preserved(self):
        home = prepare_home(self.root / "child", "author", self.baselines)
        (home / "MEMORY.md").write_text("existing learning", encoding="utf-8")
        before = {p.name: p.read_bytes() for p in home.iterdir()}
        with self.assertRaises(StateError):
            prepare_home(home, "author", self.baselines)
        self.assertEqual(before, {p.name: p.read_bytes() for p in home.iterdir()})

    def test_cross_role_and_unfinished_transaction_refused(self):
        old = prepare_home(self.root / "old", "author", self.baselines)
        with self.assertRaises(StateError):
            prepare_home(self.root / "wrong", "editor", self.baselines, old)
        (old / ".memory-stage.json").write_text("{}", encoding="utf-8")
        with self.assertRaises(StateError):
            prepare_home(self.root / "new", "author", self.baselines, old)
        self.assertFalse((self.root / "new").exists())

    def test_invalid_text_and_link_leave_no_published_home(self):
        for bad in ("utf8", "link"):
            with self.subTest(bad=bad):
                soul = self.baselines / "author" / "SOUL.md"
                soul.unlink()
                if bad == "utf8":
                    soul.write_bytes(b"\xff")
                else:
                    soul.symlink_to(self.baselines / "editor" / "SOUL.md")
                with self.assertRaises((StateError, UnicodeError)):
                    prepare_home(self.root / "new", "author", self.baselines)
                self.assertFalse((self.root / "new").exists())
                self.assertEqual(list(self.root.glob(".fiction-home-*")), [])

    def test_imported_baselines_are_independent_memory_only_copies(self):
        for role in ("author", "editor"):
            source = self.baselines / role
            (source / IMPORT_MARKER).write_bytes(IMPORT_VALUE)
            (source / "MEMORY.md").write_text("Attributed donor memory")
            for band in ("memory", "bucket-0", "bucket-1"):
                (source / band).mkdir()
                (source / band / "active.md").write_text("Donor lesson")
                (source / band / "old.md.retired").write_text("Retired")
            (source / "settings.json").write_text("not memory")
            home = prepare_home(self.root / role, role, self.baselines)
            self.assertEqual((home / "MEMORY.md").read_text(), "Attributed donor memory")
            for band in ("memory", "bucket-0", "bucket-1"):
                self.assertEqual((home / band / "active.md").read_text(), "Donor lesson")
                self.assertFalse((home / band / "old.md.retired").exists())
            self.assertFalse((home / "settings.json").exists())
            self.assertFalse((home / IMPORT_MARKER).exists())
        (self.root / "author" / "MEMORY.md").write_text("Author learns")
        self.assertEqual((self.root / "editor" / "MEMORY.md").read_text(), "Attributed donor memory")
        self.assertEqual((self.baselines / "author" / "MEMORY.md").read_text(), "Attributed donor memory")

    def test_import_marker_and_unfinished_import_refuse(self):
        source = self.baselines / "author"
        marker = source / IMPORT_MARKER
        for value in (b"", b"imported-memory-v2\n"):
            marker.write_bytes(value)
            with self.assertRaises(StateError):
                prepare_home(self.root / "new", "author", self.baselines)
            self.assertFalse((self.root / "new").exists())
        marker.write_bytes(IMPORT_VALUE)
        (source / ".memory-stage.json").write_text("{}")
        with self.assertRaises(StateError):
            prepare_home(self.root / "new", "author", self.baselines)
        self.assertFalse((self.root / "new").exists())

    def test_initial_nonempty_memory_refused(self):
        (self.baselines / "author" / "MEMORY.md").write_text("imported", encoding="utf-8")
        with self.assertRaises(StateError):
            prepare_home(self.root / "new", "author", self.baselines)
        self.assertFalse((self.root / "new").exists())


if __name__ == "__main__":
    unittest.main()
