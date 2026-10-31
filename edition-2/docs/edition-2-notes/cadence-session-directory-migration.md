# Cadence application-session migration

Status: **applied by Bill and live-verified on 2026-10-14**. The procedure below is historical/recovery documentation; do not rerun the initial migration. Bill authorized placing the active application session inside Cadence's data directory. This is Cadence's separate migration work, not Codex's compaction/schema/book refactor. Coordinate deployment changes; do not touch Codex's application-source edits, stage its work, or rebuild from an in-progress tree.

## Layout and intent

Current active session: `/Users/bill/ensemble-edition2-codex/session/`.
Proposed active session: `/Users/bill/projects/cadence-memories/session/`.
Root actor data stays `/Users/bill/projects/cadence-memories`; the existing installed binary stays in the deployment directory.

The application places API/debug/GUI logs and the named-peer registry beside its settings file. Therefore the move includes the active settings, snapshot/journal/backup, logs, diagnostics, reasoning probes, and any `ensemble.json` plus `agents/` peer tree. It does not duplicate old `cr/` or earlier migration-backup directories. All originals remain at the old location, including historical artifact paths and logs that the next startup may truncate in the new active location.

Application `session/` is not the memory band named Session. The existing memory implementation uses `memory/` for Session, `bucket-0/` for 8x, and `bucket-1/` for 64x. No memory-band layout changes are part of this move.

In the copied configuration, only `save_path` and `event_log` change. The launcher changes only its explicit configuration path. Data directory, workspace, skills, credentials, model/provider, budgets, switches, root ID, and compiled code remain unchanged. Credentials remain external. Required retained filesystem paths must already be absolute, so relocating the config cannot silently reinterpret them.

`/session/` and `/.session-migration-*/` are ignored by the memory repository. Raw state/logs are not versioned memories. Do not push either repository automatically.

## Apply while stopped

1. Let the current response finish, stop the GUI normally with Ctrl-C, and wait for its save and the terminal prompt. Investigate save errors first.
2. Coordinate with Codex: no parallel installation, launcher/settings edit, or startup during this operation. The process probes are checks, not a cross-process startup lock. Keep the app stopped for the entire procedure.
3. Run the following in a separate terminal, not an agent job hosted by the GUI being stopped:

```sh
cd /Users/bill/ensemble-edition2-codex &&
python3 migrate-cadence-session.py --apply &&
./start
```

Do not bypass a refusal or run later commands after failure. The script does not start/stop the app, rebuild code, or push. It validates the staged configuration with the installed binary's `--help` path and validates the replacement launcher's shell syntax before publishing.

It captures a fixed source fingerprint, creates private backups, stages and verifies the active files without following symlinks outside the copy, rechecks process/source/control state, publishes the new directory, and switches the launcher last. Existing destination sessions are never merged or overwritten automatically. Source changes, missing peer data, unignored directories, invalid configuration, or an active GUI cause refusal. Named-peer bytes are copied and checked, but actual peer restoration still needs separate runtime verification when exercised.

## Recovery boundaries

Before the final launcher replacement, the old launcher and old session remain authoritative. If publication succeeded but launcher replacement failed, the new directory is merely an unused copy; do not blindly rerun or delete it. Inspect the failure and both states. Backups/staging remain private and are retained for inspection.

If startup at the new location fails before any new work is accepted, stop the process and use the saved launcher/control backup to return to the old session. If the new session has already accepted work, **do not blindly restore the old launcher**: its snapshot is older. Keep both snapshot/journal sets and make an explicit recovery plan that preserves the newer history. Never delete source data or restore an old snapshot simply to get past a guard.

## Post-restart verification

Cadence should verify actual process arguments and the advancing snapshot/journal at the new location, recovery of the pre-move checkpoint, retained settings, ignored/private runtime state, and no new compaction caused by the move. New API/debug/GUI logs should be beside the new settings. The root `io/` and `children/` locations should remain unchanged from the first migration. Historical named-peer/child restoration should not be claimed unless exercised.

All six memory bands are deliberately disabled for this procedure. Do not enable identity loading or tune budgets as a hidden side effect. SOUL/MEMORY loading remains a separate decision.

## Preparation evidence

The exact operator script below passed syntax checking and thirteen temporary-filesystem cases:

- success, including copied named-peer/nested state, unchanged curated identity and root IO, unchanged original source controls, and exactly two changed config values;
- active GUI, existing destination, missing ignore rules, unexpected source entry, relative retained path, and missing registered peer;
- invalid configuration rejected by the actual installed binary;
- concurrent configuration, launcher, or source changes;
- GUI restart during staging;
- injected failure of the final launcher replacement, leaving the old launcher intact.

Real temporary files, hashing, copies, Git-ignore checks, shell syntax validation, and the installed binary's `--help` validation were used. Only process-state probes and explicit failure/race injections were simulated. No GUI or inference request was started for those validations; fake credential references were used. These are procedure tests, not live migration evidence.

Operator path: `/Users/bill/ensemble-edition2-codex/migrate-cadence-session.py`.
SHA-256 at preparation: `4c2334b829a61ec8dccafca77dbb8bdfaf46708ed837d98e1aae5572702c4167`.

## Exact tested operator script

```python
#!/usr/bin/env python3
"""One-off, offline session migration. Never starts/stops Ensemble or pushes Git."""
from pathlib import Path
import hashlib
import json
import os
import shutil
import stat
import subprocess
import sys
import tempfile

DEPLOY = Path("/Users/bill/ensemble-edition2-codex")
HOME = Path("/Users/bill/projects/cadence-memories")
OLD = DEPLOY / "session"
NEW = HOME / "session"
CONFIG = OLD / "settings.json"
START = DEPLOY / "start"
BINARY = DEPLOY / "bin/ensemble-gui"
BANDS = {"conversation", "soul", "memory", "64x", "8x", "session"}
ACTIVE = {"settings.json", "save.json", "save.json.journal", "save.json.bak",
          "event.log", "api.log", "debug.log", "gui.log", "ensemble.json",
          "agents", "diagnostics", "reasoning-probes"}
ARCHIVE_PREFIXES = ("cadence-data-dir-backup-", "cadence-session-dir-backup-")


def stopped():
    result = subprocess.run(["pgrep", "-x", "ensemble-gui"], capture_output=True, text=True)
    if result.returncode == 0:
        raise SystemExit("Stop Ensemble and wait for its save before migrating.")
    if result.returncode != 1:
        raise SystemExit("Cannot establish that Ensemble is stopped; refusing migration.")


def ignored(path):
    result = subprocess.run(["git", "-C", str(HOME), "check-ignore", "-q", "--", str(path) + "/"],
                            capture_output=True, text=True)
    if result.returncode != 0:
        raise SystemExit("The session/staging directory must be ignored by Git before copying.")


def fingerprint(path):
    # Compare bytes and tree structure without following symlinks outside the copy.
    if path.is_symlink():
        return ("link", os.readlink(path))
    if path.is_dir():
        return ("directory", tuple((p.name, fingerprint(p)) for p in sorted(path.iterdir())))
    if path.is_file():
        digest = hashlib.sha256()
        with path.open("rb") as stream:
            for block in iter(lambda: stream.read(1024 * 1024), b""):
                digest.update(block)
        return ("file", digest.hexdigest())
    raise SystemExit("Unsupported special file in session state; inspect before migration.")


def atomic_bytes(path, content, mode):
    fd, temporary = tempfile.mkstemp(prefix=".session-control-", dir=path.parent)
    try:
        os.fchmod(fd, mode)
        with os.fdopen(fd, "wb") as stream:
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


if sys.argv[1:] != ["--apply"]:
    raise SystemExit("Offline only: stop Ensemble, then run this script with --apply.")
stopped()
if any(p.is_symlink() for p in (OLD, HOME, CONFIG, START)):
    raise SystemExit("A principal path is a symlink; review the layout first.")
if NEW.exists() or NEW.is_symlink():
    raise SystemExit("Destination session already exists; no automatic merge or overwrite.")
if not OLD.is_dir() or not HOME.is_dir() or not (HOME / ".git").exists():
    raise SystemExit("Expected session directory or memory repository is missing.")
original_config = CONFIG.read_bytes()
original_start = START.read_bytes()
settings = json.loads(original_config)
if (settings.get("data_dir") != str(HOME)
        or settings.get("save_path") != str(OLD / "save.json")
        or settings.get("event_log") != str(OLD / "event.log")):
    raise SystemExit("Configured paths differ from this procedure; stop and review.")
if not all(settings.get("gui", {}).get("memory", {}).get(b, {}).get("disabled") is True for b in BANDS):
    raise SystemExit("Memory switches changed; they have not been modified. Review before migrating.")
# Moving the config must not silently change the meaning of relative path values.
for key in ("workspace", "skills_dir", "auth_directory"):
    if not isinstance(settings.get(key), str) or not Path(settings[key]).is_absolute():
        raise SystemExit("A required path is relative or missing; normalize it explicitly first.")
for value in (settings.get("sandbox_root", ""), settings.get("provider", {}).get("credential_file", "")):
    if value and (not isinstance(value, str) or not Path(value).is_absolute()):
        raise SystemExit("A retained path is relative; review it before relocating the config.")
old_argument = str(CONFIG).encode()
if original_start.count(old_argument) != 1:
    raise SystemExit("Unexpected launcher; refusing a guessed rewrite.")
if not BINARY.is_file() or not os.access(BINARY, os.X_OK) or not (OLD / "save.json").is_file():
    raise SystemExit("Installed binary or saved conversation is missing.")
entries = []
for p in sorted(OLD.iterdir()):
    if p.name == "cr" or p.name.startswith(ARCHIVE_PREFIXES):
        continue
    if p.name not in ACTIVE or p.is_symlink():
        raise SystemExit("Unexpected session entry or top-level symlink; inspect before copying.")
    entries.append(p)
if (OLD / "ensemble.json").exists():
    registry = json.loads((OLD / "ensemble.json").read_text())
    if not isinstance(registry, list):
        raise SystemExit("Malformed named-peer registry; source left untouched.")
    for peer in registry:
        name = peer.get("name") if isinstance(peer, dict) else None
        if (not isinstance(name, str) or not name.strip() or name in (".", "..")
                or "/" in name or "\\" in name or not (OLD / "agents" / name).is_dir()):
            raise SystemExit("A named peer lacks a valid data directory; inspect before moving the registry.")
ignored(NEW)
ignored(HOME / ".session-migration-probe")
source_state = {p.name: fingerprint(p) for p in entries}
HOME.chmod(0o700)
backup = Path(tempfile.mkdtemp(prefix="cadence-session-dir-backup-", dir=OLD))
backup.chmod(0o700)
for p in (START, CONFIG, OLD / "save.json", OLD / "save.json.journal", OLD / "save.json.bak"):
    if p.exists():
        shutil.copy2(p, backup / p.name)
print(f"Private backup: {backup}", flush=True)
stage = Path(tempfile.mkdtemp(prefix=".session-migration-", dir=HOME))
stage.chmod(0o700)
print(f"Private staging directory: {stage}", flush=True)
for p in entries:
    target = stage / p.name
    if p.is_dir():
        shutil.copytree(p, target, symlinks=True)
    else:
        shutil.copy2(p, target)
    if fingerprint(target) != source_state[p.name]:
        raise SystemExit("Session copy failed verification; launcher unchanged. Keep the staging directory.")
updated = dict(settings)
updated["save_path"] = str(NEW / "save.json")
updated["event_log"] = str(NEW / "event.log")
atomic_bytes(stage / "settings.json", (json.dumps(updated, indent=2, ensure_ascii=False) + "\n").encode(), 0o600)
validation = subprocess.run([str(BINARY), "--config", str(stage / "settings.json"), "--help"],
                            capture_output=True, text=True)
if validation.returncode != 0:
    raise SystemExit("Installed-binary configuration validation failed; launcher unchanged. Inspect staging privately.")
replacement = original_start.replace(old_argument, str(NEW / "settings.json").encode())
syntax = subprocess.run(["/bin/sh", "-n"], input=replacement, capture_output=True)
if syntax.returncode != 0:
    raise SystemExit("Replacement launcher failed syntax validation; launcher unchanged.")
stopped()
if CONFIG.read_bytes() != original_config or START.read_bytes() != original_start:
    raise SystemExit("Source controls changed while staging; refusing to overwrite them.")
if any(fingerprint(p) != source_state[p.name] for p in entries):
    raise SystemExit("Source session changed while staging; launcher unchanged.")
if NEW.exists() or NEW.is_symlink():
    raise SystemExit("Destination appeared while staging; refusing to replace it.")
os.rename(stage, NEW)
# This is the final switch. If it fails, the old launcher and session remain usable.
atomic_bytes(START, replacement, stat.S_IMODE(START.stat().st_mode))
print("Session published and launcher switched. Old session retained; no restart or push performed.")
print(f"Active configuration on next launch: {NEW / 'settings.json'}")
print(f"Backup retained at: {backup}")
```

## Live verification

An initial restart still used the old config and had no migration backup/staging directory; Cadence did not claim that was a successful move. Bill subsequently reported running the Python script. Verification then found:

- Launcher and running process both use `/Users/bill/projects/cadence-memories/session/settings.json`; process start was 15:27:33 local time on 2026-10-14.
- Comparing old and new configs showed only `save_path` and `event_log` changed. Root data directory, workspace, skills, credentials/provider settings, budgets, and all six disabled memory switches were preserved.
- The new session is mode0700 and ignored by Git. Journal and API/debug/GUI logs are advancing there; old files retain their pre-switch timestamps.
- Snapshot sequence3779 includes the specific pre-migration handoff in both saved event history and saved dialogue. The exact post-restart human message was confirmed in a message event at3785, not merely matched in diagnostic-command text. The journal advanced through3792 at that follow-up. Totals remain four handoffs and two compactor launches.
- Old session and private backup `/Users/bill/ensemble-edition2-codex/session/cadence-session-dir-backup-_zv8r920` are retained. No private staging directory remains.
- No application-source changes, rebuild, automatic restart, or push were performed by Cadence. Bill performed stop/apply/restart. Named-peer/child restoration has not been exercised and is not claimed.

The active actor/session layout is now nested as intended. Automatic identity-memory activation remains a separate task.
