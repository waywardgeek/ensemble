"""Frozen review-round inputs; working drafts remain separate and mutable.

Caller holds RunStore's lock and has joined the writing children. Publication is
an atomic directory rename, not a transaction with child work or directory fsync.
"""
import hashlib
import json
import os
from pathlib import Path
import shutil
import tempfile

from run_state import StateError, _object, _reject_constant


def _digest(data):
    return hashlib.sha256(data).hexdigest()


def _read(path, limit):
    if path.is_symlink() or not path.is_file():
        raise StateError("Review input must be an ordinary file")
    with path.open("rb") as source:
        data = source.read(limit + 1)
    if len(data) > limit:
        raise StateError("Review input exceeds the configured byte limit")
    if not data.decode("utf-8").strip():
        raise StateError("Review input is empty")
    return data


def _write(path, data):
    with path.open("xb") as output:
        os.chmod(path, 0o600)
        output.write(data)
        output.flush()
        os.fsync(output.fileno())


def _freeze(store, round_number, chapters, byte_limit, contexts=()):
    """Return run-relative frozen chapter and manuscript paths.

    Chapter order and original paths are retained in the manifest. On recovery,
    verify the published snapshot without rereading mutable working chapters.
    The same bound applies to each input and the assembled UTF-8 manuscript;
    it is not a model token-capacity estimate. Refuse changed parameters.
    """
    store.snapshot()  # Require an open, healthy, exclusively held store.
    if type(round_number) is not int or not 0 <= round_number <= 3:
        raise StateError("Invalid snapshot position")
    if type(byte_limit) is not int or byte_limit < 1:
        raise StateError("Review byte limit must be positive")
    if not chapters or len(set(chapters)) != len(chapters):
        raise StateError("Expected distinct ordered chapter paths")
    if len(set(contexts)) != len(contexts) or set(contexts) & set(chapters):
        raise StateError("Snapshot context paths must be distinct")
    for name in list(chapters) + list(contexts):
        store._relative(name)
    destination = store.directory / ("review-inputs-%d" % round_number
                                     if round_number else "final-book")
    names = ["chapter-%04d.md" % (index + 1) for index in range(len(chapters))]
    context_names = ["context-%04d.md" % (index + 1) for index in range(len(contexts))]
    files = names + ["manuscript.md"] + context_names
    expected = {"version": 1, "round": round_number,
                "chapters": chapters, "byte_limit": byte_limit, "contexts": list(contexts)}
    if destination.exists() or destination.is_symlink():
        if destination.is_symlink() or not destination.is_dir():
            raise StateError("Review snapshot must be an ordinary directory")
        manifest = json.loads(_read(destination / "manifest.json", 1048576),
                              object_pairs_hook=_object,
                              parse_constant=_reject_constant)
        if (type(manifest) is not dict or set(manifest) != set(expected) | {"hashes"}
                or any(manifest[key] != value for key, value in expected.items())
                or type(manifest["hashes"]) is not dict
                or set(manifest["hashes"]) != set(files)):
            raise StateError("Review snapshot parameters changed or are malformed")
        if {path.name for path in destination.iterdir()} != set(files) | {"manifest.json"}:
            raise StateError("Unexpected review snapshot files")
        for name in files:
            if _digest(_read(destination / name, byte_limit)) != manifest["hashes"][name]:
                raise StateError("Frozen review input changed")
    else:
        staging = Path(tempfile.mkdtemp(prefix=".review-inputs-", dir=store.directory))
        try:
            parts = []
            hashes = {}
            total = 0
            for index, name in enumerate(chapters):
                source = (store.directory / name).resolve(strict=True)
                if store.directory not in source.parents:
                    raise StateError("Chapter escapes run directory")
                # Reject even an in-root final symlink rather than following it.
                if (store.directory / name).is_symlink():
                    raise StateError("Chapter must not be a link")
                data = _read(source, byte_limit)
                part = ("# Chapter %d\n\n" % (index + 1)).encode() + data + b"\n\n"
                total += len(part)
                if total > byte_limit:
                    raise StateError("Assembled manuscript exceeds the configured byte limit")
                parts.append(part)
                _write(staging / names[index], data)
                hashes[names[index]] = _digest(data)
            for original, name in zip(contexts, context_names):
                path = store.directory / original
                source = path.resolve(strict=True)
                if path.is_symlink() or store.directory not in source.parents:
                    raise StateError("Snapshot context must be an ordinary run-local file")
                data = _read(source, byte_limit)
                _write(staging / name, data)
                hashes[name] = _digest(data)
            manuscript = b"".join(parts)
            _write(staging / "manuscript.md", manuscript)
            hashes["manuscript.md"] = _digest(manuscript)
            manifest = dict(expected, hashes=hashes)
            _write(staging / "manifest.json",
                   (json.dumps(manifest, sort_keys=True, indent=2) + "\n").encode())
            if destination.exists() or destination.is_symlink():
                raise StateError("Review snapshot appeared during preparation")
            staging.rename(destination)
        finally:
            if staging.exists():
                shutil.rmtree(staging)
    prefix = destination.name + "/"
    return {"chapters": [prefix + name for name in names],
            "manuscript": prefix + "manuscript.md",
            "contexts": [prefix + name for name in context_names]}


def freeze_round(store, round_number, chapters, byte_limit, contexts=()):
    """Publish/verify one of the three pre-revision review snapshots."""
    if type(round_number) is not int or not 1 <= round_number <= 3:
        raise StateError("Review round must be one through three")
    return _freeze(store, round_number, chapters, byte_limit, contexts)


def freeze_final(store, chapters, byte_limit, contexts=()):
    """Publish the post-revision deliverable after all producing children join."""
    return _freeze(store, 0, chapters, byte_limit, contexts)
