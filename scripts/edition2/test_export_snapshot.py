"""Verify snapshot provenance and preservation using an independent small Git repo."""

import hashlib
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

from export_snapshot import export, SOURCE_PREFIX


class ExportTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)
        self.repo = self.base / "repo"
        self.repo.mkdir()
        self.git("init", "-q")
        self.source = self.repo / SOURCE_PREFIX
        self.source.mkdir(parents=True)
        (self.source / "go.mod").write_text("module example.com/fixture\n")
        (self.source / ".gitignore").write_text("*.log\n")
        (self.source / "receipt.log").write_bytes(b"raw\r\n\x00receipt\n")
        (self.source / "run").write_text("#!/bin/sh\nexit 0\n")
        (self.source / "run").chmod(0o755)
        (self.source / "link").symlink_to("run")
        self.git("add", "-f", SOURCE_PREFIX)
        self.git("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid",
                 "commit", "-qm", "fixture")
        self.output = self.base / "export"
        self.manifest = self.base / "manifest.json"

    def git(self, *args):
        return subprocess.check_output(["git", "-C", str(self.repo), *args])

    def test_committed_bytes_modes_and_ignored_receipts_survive(self):
        (self.source / "go.mod").write_text("uncommitted replacement\n")
        (self.source / "secret.tmp").write_text("not part of source\n")
        result = export(self.repo, "HEAD", self.output, self.manifest)
        self.assertEqual((self.output / "go.mod").read_text(), "module example.com/fixture\n")
        self.assertEqual((self.output / "receipt.log").read_bytes(), b"raw\r\n\x00receipt\n")
        self.assertEqual((self.output / "run").stat().st_mode & 0o777, 0o755)
        self.assertTrue((self.output / "link").is_symlink())
        self.assertEqual((self.output / "link").readlink(), Path("run"))
        self.assertFalse((self.output / "secret.tmp").exists())
        self.assertEqual(json.loads(self.manifest.read_text()), result)
        for item in result["files"]:
            original = self.git("show", result["source_commit"] + ":" + SOURCE_PREFIX + "/" + item["path"])
            self.assertEqual(hashlib.sha256(original).hexdigest(), item["sha256"])

    def test_existing_export_is_never_overwritten(self):
        self.output.mkdir()
        marker = self.output / "keep"
        marker.write_text("existing chapter")
        with self.assertRaisesRegex(ValueError, "already exists"):
            export(self.repo, "HEAD", self.output, self.manifest)
        self.assertEqual(marker.read_text(), "existing chapter")
        self.assertFalse(self.manifest.exists())

    def test_manifest_cannot_change_the_exported_source_tree(self):
        with self.assertRaisesRegex(ValueError, "outside"):
            export(self.repo, "HEAD", self.output, self.output / "manifest.json")
        self.assertFalse(self.output.exists())


if __name__ == "__main__":
    unittest.main()
