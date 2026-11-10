"""Prepare chapter-local child homes without copying conversation state.

The driver must hold its run lock and shut down the previous child before calling
prepare_home. This module cannot prove quiescence from filesystem observations.
It neither creates Agents nor changes their memory policy or sandbox grants.
"""
import json
import os
from pathlib import Path
import shutil
import tempfile

from run_state import StateError


MARKER = "fiction-role.json"
BANDS = ("memory", "bucket-0", "bucket-1")
IMPORT_MARKER = "imported-memory.txt"
IMPORT_VALUE = b"imported-memory-v1\n"


def _regular(path):
    if path.is_symlink() or not path.is_file():
        raise StateError("Role memory must contain ordinary files, not links")
    return path


def _copy_text(source, destination):
    # Match the memory corpus's strict UTF-8 contract; do not repair invalid data.
    data = _regular(source).read_bytes()
    data.decode("utf-8", errors="strict")
    with destination.open("xb") as output:
        os.chmod(destination, 0o600)
        output.write(data)
        output.flush()
        os.fsync(output.fileno())


def prepare_home(destination, role, baseline_root, previous_home=None):
    """Publish a new role home, returning its path.

    First chapter: seed SOUL.md and empty MEMORY.md by default. An explicit
    imported-memory.txt marker opts into stable files and active memory bands. Later chapters: transfer that role's stable files and active corpus
    files after its old child has been shut down. Never copy save.json, settings,
    submission records, Jobs, or retired memory. The native memory reader will
    validate corpus identities when the child loads them.

    Destination parent must already exist within the caller's authorized sandbox.
    A single run driver owns child-home creation. Refusal preserves an existing
    destination; an interrupted unpublished staging directory is not a child home.
    This is not a defense against another process racing filesystem mutations,
    nor a directory-fsync power-loss guarantee.
    """
    if role not in ("author", "editor"):
        raise StateError("Unknown writing role")
    destination = Path(destination)
    if destination.exists() or destination.is_symlink():
        raise StateError("Child home already exists; reconcile instead of resetting it")
    if not destination.parent.is_dir():
        raise StateError("Child-home parent must already exist")

    source = Path(baseline_root) / role
    transfer_corpus = previous_home is not None
    if previous_home is not None:
        source = Path(previous_home)
        marker = json.loads(_regular(source / MARKER).read_text(encoding="utf-8"))
        if marker != {"version": 1, "role": role}:
            raise StateError("Previous home belongs to a different role or format")
        if (source / ".memory-stage.json").exists():
            raise StateError("Previous memory transaction needs recovery before transfer")
    else:
        marker = source / IMPORT_MARKER
        if marker.exists() or marker.is_symlink():
            if _regular(marker).read_bytes() != IMPORT_VALUE:
                raise StateError("Invalid imported-memory baseline marker")
            transfer_corpus = True
        elif _regular(source / "MEMORY.md").read_bytes() != b"":
            raise StateError("Initial role MEMORY.md must be empty")
    if (source / ".memory-stage.json").exists():
        raise StateError("Memory transaction needs recovery before transfer")
    if source.is_symlink() or not source.is_dir():
        raise StateError("Role source must be an ordinary directory")

    staging = Path(tempfile.mkdtemp(prefix=".fiction-home-", dir=destination.parent))
    try:
        for name in ("SOUL.md", "MEMORY.md"):
            _copy_text(source / name, staging / name)
        if transfer_corpus:
            for band in BANDS:
                directory = source / band
                if directory.is_symlink():
                    raise StateError("Role memory directory must not be a link")
                if not directory.exists():
                    continue
                if not directory.is_dir():
                    raise StateError("Role memory band must be a directory")
                target = staging / band
                target.mkdir(mode=0o700)
                for entry in sorted(directory.iterdir()):
                    if entry.name.endswith(".md"):
                        _copy_text(entry, target / entry.name)
        with (staging / MARKER).open("x", encoding="utf-8") as output:
            os.chmod(staging / MARKER, 0o600)
            json.dump({"version": 1, "role": role}, output, sort_keys=True)
            output.write("\n")
            output.flush()
            os.fsync(output.fileno())
        # The run lock excludes cooperating drivers. Recheck before publication;
        # never intentionally replace any already-created child, even an empty one.
        if destination.exists() or destination.is_symlink():
            raise StateError("Child home appeared during preparation; reconcile it")
        staging.rename(destination)
        return destination
    finally:
        if staging.exists():
            shutil.rmtree(staging)
