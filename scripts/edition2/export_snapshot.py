#!/usr/bin/env python3
"""Export a committed Edition 2 source tree and a separate hash manifest.

This copies committed blobs, including tracked evidence ignored by outer Git.
It does not validate a chapter, stage files, overwrite an export, or create tags.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import subprocess
import tempfile


SOURCE_PREFIX = "solutions/edition-2/main"


def git(repo, *args):
    return subprocess.check_output(["git", "-C", str(repo), *args])


def export(repo, revision, destination, manifest_path):
    destination = Path(destination).absolute()
    manifest_path = Path(manifest_path).absolute()
    if destination.exists() or destination.is_symlink():
        raise ValueError("destination already exists; preserve the earlier export")
    if manifest_path.exists() or manifest_path.is_symlink():
        raise ValueError("manifest already exists; choose a new revision")
    if manifest_path == destination or destination in manifest_path.parents:
        raise ValueError("manifest must live outside the exact source export")
    commit = git(repo, "rev-parse", "--verify", revision + "^{commit}").decode().strip()
    tree = git(repo, "rev-parse", commit + ":" + SOURCE_PREFIX).decode().strip()
    entries = []
    for row in git(repo, "ls-tree", "-r", "-z", tree).split(b"\0"):
        if not row:
            continue
        metadata, raw_path = row.split(b"\t", 1)
        mode, kind, oid = metadata.decode().split()
        path = raw_path.decode()
        parts = PurePosixPath(path)
        if parts.is_absolute() or ".." in parts.parts or ".git" in parts.parts:
            raise ValueError("invalid source path: " + path)
        if kind != "blob" or mode not in ("100644", "100755", "120000"):
            raise ValueError("source must contain ordinary files or symlinks: " + path)
        entries.append((path, mode, oid))
    if not any(path == "go.mod" for path, _, _ in entries):
        raise ValueError("source tree has no root Go module")
    destination.parent.mkdir(parents=True, exist_ok=True)
    manifest_path.parent.mkdir(parents=True, exist_ok=True)
    records = []
    with tempfile.TemporaryDirectory(prefix="edition2-export-", dir=destination.parent) as tmp:
        output = Path(tmp) / "source"
        output.mkdir()
        for path, mode, oid in entries:
            data = git(repo, "cat-file", "blob", oid)
            target = output / path
            target.parent.mkdir(parents=True, exist_ok=True)
            if mode == "120000":
                target.symlink_to(os.fsdecode(data))
                saved = os.fsencode(os.readlink(target))
            else:
                target.write_bytes(data)
                target.chmod(0o755 if mode == "100755" else 0o644)
                saved = target.read_bytes()
            if saved != data:
                raise ValueError("export differs from committed blob: " + path)
            records.append({"path": path, "mode": mode, "git_blob": oid,
                            "sha256": hashlib.sha256(data).hexdigest()})
        manifest = {"source_commit": commit, "source_prefix": SOURCE_PREFIX,
                    "source_tree": tree, "files": records,
                    "validation": "Source export only; chapter acceptance is recorded separately."}
        # Exclusive creation prevents silently replacing another revision's manifest.
        with manifest_path.open("x") as stream:
            json.dump(manifest, stream, indent=2)
            stream.write("\n")
        if destination.exists() or destination.is_symlink():
            raise ValueError("destination appeared during export; no source was replaced")
        output.rename(destination)
    return manifest


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("revision", help="committed, validated main source revision")
    parser.add_argument("destination", help="new, absent export directory")
    parser.add_argument("--manifest", required=True, help="new manifest path outside the export")
    parser.add_argument("--repo", default=".", help="outer Ensemble repository")
    args = parser.parse_args()
    result = export(args.repo, args.revision, args.destination, args.manifest)
    print(f"Exported {len(result['files'])} files from {result['source_commit']}")


if __name__ == "__main__":
    main()
