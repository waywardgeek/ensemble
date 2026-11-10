"""Durable assignment ledger for the fiction workflow.

This module owns no Agents and dispatches no tools. The driver holds RunStore
while coordinating ordinary Ensemble child operations. Persisting intent cannot
make a remote dispatch atomic: a pending assignment must be reconciled, never
silently retried. Only accepted, matching submissions advance the ledger.
"""
import copy
import fcntl
import hashlib
import json
import os
from pathlib import Path
import tempfile
import uuid


class StateError(ValueError):
    """Expected refusal with a safe, fixed diagnostic for the supervisor.

    Messages must not interpolate paths, file contents, credentials or child
    payloads. The public workflow may return this message; unexpected exceptions
    remain opaque. Keep diagnostics bounded and about the violated contract.
    """


def _reject_constant(value):
    raise StateError("Non-finite numbers are not valid workflow JSON")


def _object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise StateError("Duplicate state field")
        result[key] = value
    return result


def _keys(value, names):
    if type(value) is not dict or set(value) != set(names):
        raise StateError("Unexpected state fields")


def _text(value):
    if type(value) is not str or not value.strip():
        raise StateError("Expected nonempty identity or text")


class RunStore:
    """Exclusive, context-managed owner of one run's ledger.

    The directory must already exist inside the caller's sandbox. The lock is
    cooperative, Unix-only and deliberately never unlinked: deleting a lock file
    can let two processes lock different inodes at the same pathname. Closing the
    store releases the lock, including on errors. No owned background work exists.
    State is private to this owner; callers receive detached JSON snapshots.
    """

    MAX_BYTES = 1024 * 1024

    def __init__(self, directory):
        self.directory = Path(directory).resolve(strict=True)
        self._lock = None
        self._state = None
        self._failed = False

    def __enter__(self):
        if self._lock is not None:
            raise StateError("Run store is already open")
        lock = (self.directory / "workflow.lock").open("a+b")
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            self._lock = lock
            path = self.directory / "state.json"
            try:
                with path.open("rb") as source:
                    data = source.read(self.MAX_BYTES + 1)
            except FileNotFoundError:
                self._save({"version": 1, "run_id": uuid.uuid4().hex,
                            "pending": None, "accepted": []})
            else:
                if len(data) > self.MAX_BYTES:
                    raise StateError("Workflow state exceeds its byte limit")
                state = json.loads(data.decode("utf-8"), object_pairs_hook=_object,
                                   parse_constant=_reject_constant)
                self._validate(state)
                self._state = state
            return self
        except BaseException:
            self._lock = None
            lock.close()
            raise

    def __exit__(self, kind, error, traceback):
        self._state = None
        lock = self._lock
        self._lock = None
        if lock is not None:
            lock.close()

    def snapshot(self):
        self._require_open()
        return copy.deepcopy(self._state)

    def _require_open(self):
        if self._lock is None or self._failed:
            raise StateError("Run store is closed or requires reopening after a save failure")

    @staticmethod
    def _assignment(value):
        _keys(value, ("id", "child_id", "phase", "round", "chapter", "turn", "task"))
        for name in ("id", "child_id", "task"):
            _text(value[name])
        if value["phase"] not in ("write", "detail_review", "global_review", "revise"):
            raise StateError("Unknown fiction phase")
        for name in ("round", "chapter", "turn"):
            if type(value[name]) is not int or value[name] < 0:
                raise StateError("Invalid assignment position")

    @classmethod
    def _validate(cls, state):
        _keys(state, ("version", "run_id", "pending", "accepted"))
        if type(state["version"]) is not int or state["version"] != 1:
            raise StateError("Unsupported workflow state version")
        _text(state["run_id"])
        if type(state["accepted"]) is not list:
            raise StateError("Invalid accepted results")
        identities = set()
        for receipt in state["accepted"]:
            _keys(receipt, ("assignment", "submission", "artifacts"))
            assignment = receipt["assignment"]
            cls._assignment(assignment)
            if assignment["id"] in identities:
                raise StateError("Duplicate assignment identity")
            identities.add(assignment["id"])
            if type(receipt["submission"]) is not dict:
                raise StateError("Invalid accepted submission")
            if receipt["submission"].get("assignment_id") != assignment["id"]:
                raise StateError("Submission identity differs from assignment")
            if type(receipt["artifacts"]) is not dict or not receipt["artifacts"]:
                raise StateError("Missing accepted artifacts")
            for name, digest in receipt["artifacts"].items():
                cls._relative(name)
                if (type(digest) is not str or len(digest) != 64
                        or any(c not in "0123456789abcdef" for c in digest)):
                    raise StateError("Invalid artifact digest")
        if state["pending"] is not None:
            cls._assignment(state["pending"])
            if state["pending"]["id"] in identities:
                raise StateError("Accepted assignment is still pending")

    @staticmethod
    def _relative(name):
        if type(name) is not str or not name or "\\" in name:
            raise StateError("Invalid artifact path")
        path = Path(name)
        if path.is_absolute() or any(part in ("", ".", "..") for part in name.split("/")):
            raise StateError("Artifact must be a relative run path")
        return path

    def _save(self, candidate):
        self._require_open()
        self._validate(candidate)
        data = (json.dumps(candidate, ensure_ascii=False, allow_nan=False,
                           sort_keys=True, indent=2) + "\n").encode("utf-8")
        if len(data) > self.MAX_BYTES:
            raise StateError("Workflow state exceeds its byte limit")
        temporary = None
        try:
            # Flush complete bytes before replacing. This is atomic publication,
            # not a transaction with child files, nor a power-loss guarantee.
            with tempfile.NamedTemporaryFile(dir=self.directory, prefix=".state-",
                                             delete=False) as output:
                temporary = Path(output.name)
                output.write(data)
                output.flush()
                os.fsync(output.fileno())
            os.replace(temporary, self.directory / "state.json")
            self._state = candidate
        except BaseException:
            # A write failure must not allow further work based on stale memory.
            self._failed = True
            raise
        finally:
            if temporary is not None:
                temporary.unlink(missing_ok=True)

    def prepare(self, child_id, phase, round_number, chapter, turn, task):
        """Durably record intent. Caller dispatches only after this returns.

        A preexisting pending assignment always refuses, even if it looks like
        the same request. The driver must reconcile it with Ensemble first.
        """
        self._require_open()
        if self._state["pending"] is not None:
            raise StateError("Reconcile pending assignment before preparing work")
        assignment = {"id": uuid.uuid4().hex, "child_id": child_id,
                      "phase": phase, "round": round_number, "chapter": chapter,
                      "turn": turn, "task": task}
        self._assignment(assignment)
        candidate = self.snapshot()
        candidate["pending"] = assignment
        self._save(candidate)
        return copy.deepcopy(assignment)

    def accept(self, submission, artifacts):
        """Retain a matching result and hashes of required run-local files.

        The driver supplies required paths, not the child. It must separately
        validate phase-specific content and approval. A receipt records observed
        bytes, not a claim that later edits cannot change them. Call only after
        the producing child has stopped writing these artifacts.
        """
        self._require_open()
        pending = self._state["pending"]
        if pending is None:
            raise StateError("No pending assignment")
        if type(submission) is not dict or submission.get("assignment_id") != pending["id"]:
            raise StateError("Submission does not match pending assignment")
        if not artifacts or len(set(artifacts)) != len(artifacts):
            raise StateError("Expected distinct required artifacts")
        hashes = {}
        for name in artifacts:
            relative = self._relative(name)
            path = (self.directory / relative).resolve(strict=True)
            if self.directory not in path.parents:
                raise StateError("Artifact escapes run directory")
            digest = hashlib.sha256()
            with path.open("rb") as source:
                while True:
                    block = source.read(65536)
                    if not block:
                        break
                    digest.update(block)
            hashes[name] = digest.hexdigest()
        candidate = self.snapshot()
        candidate["accepted"].append({"assignment": pending,
                                      "submission": copy.deepcopy(submission),
                                      "artifacts": hashes})
        candidate["pending"] = None
        self._save(candidate)
