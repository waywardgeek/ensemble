# Cadence data-directory migration — offline operator procedure

Status: **applied by Bill and live-verified on 2026-10-14**. The steps below remain as the historical procedure and recovery reference; do not rerun the initial migration. Cadence owns this small migration; Codex owns the separate compaction/submission-schema/chapter refactor. This note is not permission to overwrite Codex's files, rebuild or install its in-progress code, restart the live GUI, or push either repository.

## Purpose and boundaries

Bill wants `/Users/bill/projects/cadence-memories` to be the main actor's data directory, with runtime children underneath their parent. Absolute data directories are already supported. No application-code change is currently needed.

- Old data directory: `/Users/bill/ensemble-edition2-codex/session/cr`.
- New data directory: `/Users/bill/projects/cadence-memories`.
- Authoritative settings: `/Users/bill/ensemble-edition2-codex/session/settings.json`.
- Keep the explicit snapshot at `/Users/bill/ensemble-edition2-codex/session/save.json`, with its journal and backup. A data-directory move is not a conversation rollback.
- Keep workspace, skills, provider, model, credentials, agent identity/ID, registry base, GUI preferences, and memory switches unchanged. Do not rename the runtime root ID merely because the actor's conversational name is Cadence.
- Keep the entire original data directory. Saved tool records can contain absolute artifact paths pointing there.
- Do not change configuration while Ensemble is running: its in-memory settings store could overwrite an external edit.
- Do not copy the old runtime `settings.json` over the authoritative application settings.

At the preparation inspection, all six memory bands were disabled. This procedure requires that state and preserves it. If Bill has since changed the bands or paths, stop and review the procedure rather than force old values onto new settings.

## Privacy and identity loading

The memory repository now ignores `io/`, `children/`, request caches, runtime `submit.md`, legacy `settings.json`, snapshots/journals, recall caches and temporary files. Ignored runtime data may exist beside tracked memories, but must not be staged or published wholesale. The copied caches can contain private request payloads. Credentials remain external. Neither repository is to be pushed by Cadence.

This move does **not** enable identity-memory loading. The existing loader reads `SOUL.md` and `MEMORY.md` directly under `data_dir`, but the bands are currently disabled. Their zero/default budgets are only 1,500 bytes and would truncate the current files. Enabling them and choosing adequate byte budgets require a separate decision. `CURRENT.md` is not automatically loaded by that loader.

## Before applying

1. Review and commit the prepared memory-repository changes locally. Check for your own uncommitted edits; never use a broad reset or clean.
2. Coordinate with Codex: it must not install/restart the live application or edit this settings file during the move. Its source work can continue separately.
3. Let the current answer finish, press Ctrl-C once in the GUI server terminal, and wait for the terminal prompt and graceful save. Investigate any save error before proceeding.
4. Run the block below from a separate terminal. It does not start or stop processes. It refuses to run if any process named `ensemble-gui` is active; if that is an isolated Codex test, coordinate or wait rather than killing it blindly.

## Apply once, while stopped

The block is intentionally specific to the verified deployment. It copies only the six known runtime entries, verifies the copies, and atomically changes one JSON field. It refuses unexpected source entries, unignored destinations, existing destination runtime files, active GUI processes, or concurrent settings changes. On failure, leave both directories and the backup intact and inspect the error; do not force a merge or delete the partial copy just to retry.

```python
from pathlib import Path
import json
import os
import shutil
import subprocess
import tempfile

config = Path("/Users/bill/ensemble-edition2-codex/session/settings.json")
old = Path("/Users/bill/ensemble-edition2-codex/session/cr")
new = Path("/Users/bill/projects/cadence-memories")
expected_save = Path("/Users/bill/ensemble-edition2-codex/session/save.json")
allowed = {"cache-current.txt", "cache-prior.txt", "children", "io", "settings.json", "submit.md"}
bands = {"conversation", "soul", "memory", "64x", "8x", "session"}

def require_stopped():
    result = subprocess.run(["pgrep", "-x", "ensemble-gui"], capture_output=True, text=True)
    if result.returncode == 0:
        raise SystemExit("Stop Ensemble and wait for its save before migrating.")
    if result.returncode != 1:
        raise SystemExit("Could not verify that Ensemble is stopped; no migration allowed.")

require_stopped()
if config.is_symlink() or old.is_symlink() or new.is_symlink():
    raise SystemExit("A principal path is a symlink; review the layout before migrating.")
original = config.read_bytes()
settings = json.loads(original)
if settings.get("data_dir") != str(old) or settings.get("save_path") != str(expected_save):
    raise SystemExit("Configured paths differ from this procedure; stop and review.")
memory = settings.get("gui", {}).get("memory", {})
if not all(memory.get(band, {}).get("disabled") is True for band in bands):
    raise SystemExit("Memory switches changed; review before migrating. They were not modified.")
if not old.is_dir() or not new.is_dir() or not (new / ".git").exists() or not expected_save.is_file():
    raise SystemExit("Expected data directory, Git repository, or saved conversation is missing.")
entries = sorted(old.iterdir())
if any(path.name not in allowed or path.is_symlink() for path in entries):
    raise SystemExit("Unexpected source entry or top-level symlink; inspect before copying.")
for source in entries:
    target = new / source.name
    if target.exists() or target.is_symlink():
        raise SystemExit(f"Destination already exists: {target}. No automatic overwrite.")
    ignore_path = str(target) + ("/" if source.is_dir() else "")
    check = subprocess.run(["git", "-C", str(new), "check-ignore", "-q", "--", ignore_path],
                           capture_output=True, text=True)
    if check.returncode != 0:
        raise SystemExit(f"Runtime path is not ignored by Git: {target}")

# Raw caches are local/private even though they are excluded from Git.
new.chmod(0o700)
backup = Path(tempfile.mkdtemp(prefix="cadence-data-dir-backup-", dir=config.parent))
backup.chmod(0o700)
shutil.copy2(config, backup / "settings.json")
for saved in (expected_save, Path(str(expected_save) + ".journal"), Path(str(expected_save) + ".bak")):
    if saved.exists():
        shutil.copy2(saved, backup / saved.name)
print(f"Private backup: {backup}", flush=True)
for source in entries:
    target = new / source.name
    if source.is_dir():
        shutil.copytree(source, target, symlinks=True)
    else:
        shutil.copy2(source, target, follow_symlinks=False)
    check = subprocess.run(["diff", "-qr", str(source), str(target)], capture_output=True, text=True)
    if check.returncode != 0:
        raise SystemExit(f"Copy verification failed for {source.name}; configuration unchanged.")

require_stopped()
if config.read_bytes() != original:
    raise SystemExit("Settings changed during the copy; configuration not overwritten.")
updated = dict(settings)
updated["data_dir"] = str(new)
fd, temporary = tempfile.mkstemp(prefix=".settings-cadence-", dir=config.parent)
try:
    os.fchmod(fd, 0o600)
    with os.fdopen(fd, "w", encoding="utf-8") as output:
        json.dump(updated, output, indent=2, ensure_ascii=False)
        output.write("\n")
        output.flush()
        os.fsync(output.fileno())
    os.replace(temporary, config)
finally:
    if os.path.exists(temporary):
        os.unlink(temporary)
print("Changed only data_dir. Original runtime data retained; no restart or push performed.")
print(f"Backup retained at: {backup}")
```

Save the block to a temporary `.py` file outside the tracked memory record and run it with `python3`, or paste it into a Python heredoc. Do not execute it from an agent job that depends on the GUI process staying alive. Do not publish its private backup directory.

## Validate and restart

While still stopped, validate the configuration with the existing installed binary (no rebuild):

```sh
/Users/bill/ensemble-edition2-codex/bin/ensemble-gui \
  --config /Users/bill/ensemble-edition2-codex/session/settings.json --help
```

Then start normally:

```sh
cd /Users/bill/ensemble-edition2-codex
./start
```

After restart, ask Cadence to verify:

- The running configuration uses the new absolute data directory, not just that the file was edited.
- Newly produced runtime output goes under the new `io/` directory.
- The existing conversation and checkpoint were recovered, with snapshot/journal still at the original explicit save path.
- All six memory switches remain disabled; no compressor or identity-loading change was introduced.
- The memory repository's Git status does not expose copied runtime data. Review before every commit; ignore rules are not a privacy audit.
- A future legitimate child uses the new parent-relative child directory. Root-path verification alone does not prove every historical child can be restored; report that check separately when exercised.

No automatic push, rebuild, source change, or compaction test is part of the migration.

## Recovery

If the copy or configuration change fails, the original data directory and private backup remain available. Stop and inspect; do not automatically overwrite conflicting destinations or delete directories to get a retry past the guard.

If the new startup fails, keep the GUI stopped and restore only the original `data_dir` value, using the saved settings as a reference. If no other settings have changed, the backed-up settings file can restore the previous configuration. Do not blindly overwrite subsequent preference changes. Keep both data directories and the current conversation snapshot/journal. Restoring an old snapshot is a separate recovery decision, not a normal directory rollback.

## Preparation evidence

The block must be checked against temporary filesystem fixtures before handing it to Bill. Such checks validate the procedure, not the live migration. The real source directory, application settings, and active runtime must remain unchanged during preparation. Record the actual fixture results and local checkpoint commits below when complete.

### Preparation completed (migration still pending)

- Syntax checked; seven temporary-filesystem cases passed: success, active GUI, destination collision, missing ignore rule, unexpected source entry, concurrent settings edit, and GUI restarting during the copy. Only the process probe was mocked; filesystem copying, comparison, backup, Git-ignore checks, and atomic JSON replacement were real within the temporary fixtures.
- Success changed only the configuration's `data_dir` value, preserved the old runtime data and saved conversation, retained private backups, and kept curated identity text unchanged. Refusal cases did not overwrite configuration; the concurrent-edit case preserved that edit.
- Exact code-block export: `/Users/bill/ensemble-edition2-codex/migrate-cadence-data-dir.py`. SHA-256: `b0b4412d103eed9635216b254f6c2527f1465679f2d78bdf703ea4939442ac55`. This script has **not** been run against the real deployment.
- The memory-repository exclusions and checkpoint were prepared locally. No runtime files have been copied, no memory switches enabled, and no active configuration, binary, or process changed. No pushes performed.

After graceful shutdown, Bill can run:

```sh
cd /Users/bill/ensemble-edition2-codex
python3 migrate-cadence-data-dir.py
# Only if the command succeeds:
./bin/ensemble-gui --config ./session/settings.json --help
./start
```

Do not run the later commands if migration or validation fails. Ask Cadence to verify the running location and recovery after restart.

## Live verification after Bill applied the procedure

- Bill reported running the operator commands and restarting. The process inspection found the expected deployed binary and `--config` path. New runtime output was observed in `/Users/bill/projects/cadence-memories/io/`; the old `session/cr/io/` retained its previous timestamps. This establishes active use, not merely a changed configuration file.
- The private backup is `session/cadence-data-dir-backup-j09slkh1`. A semantic comparison of its settings with the current application settings showed exactly one changed top-level field: `data_dir`.
- The explicit `save_path`, workspace, skills path, and all six disabled memory switches were unchanged. The new directory is mode0700.
- The specific all-bands-disabled checkpoint was found in the saved event history and saved dialogue. The restart message was found in the live journal. The snapshot was at sequence3684 and the journal advanced through3698 at the recovery check. Total handoffs remained4 and compactor launches2; no new compression was introduced.
- All six copied runtime entries were present and ignored by Git. The memory repository was clean before these verification-note edits. The old data directory and migration backup remain intact.
- Automatic SOUL/MEMORY loading is still disabled. No child was spawned for this verification, so new-child placement and restoration of historical child data remain separate checks when those paths are exercised.
- No source-code change, rebuild, forced restart, or push was performed by Cadence for this migration. Bill performed the stop, migration, and restart.
