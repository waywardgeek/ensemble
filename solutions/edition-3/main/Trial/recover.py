#!/usr/bin/env python3
"""One-shot recovery controller. Standard library only; no installation or credentials.

Run independently of either GUI (temporary launchd recipe in README). The plan
is explicit human authority, not discovery of a private launcher. Touch switch
only AFTER observing status.json:armed. Touch restore for early restoration, or
cancel while armed to leave the original untouched. Never reuse a state folder.
"""
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import time

os.umask(0o077)


def identity(pid):
    """Read process start/executable identity, never environment or arguments."""
    result = subprocess.run(['/bin/ps', '-p', str(pid), '-o', 'lstart=', '-o', 'comm='],
                            capture_output=True, text=True, check=False)
    return result.stdout.strip() if result.returncode == 0 else ''


def listeners(port):
    result = subprocess.run(['/usr/sbin/lsof', '-nP', '-t', f'-iTCP:{port}', '-sTCP:LISTEN'],
                            capture_output=True, text=True, check=False)
    if result.returncode not in (0, 1):
        raise RuntimeError('Listener inspection failed; refusing effects')
    return {int(value) for value in result.stdout.split()}


def in_group(pid, group):
    try:
        return os.getpgid(pid) == group
    except ProcessLookupError:
        return False


def run(plan_file):
    plan = json.loads(Path(plan_file).read_text())
    state = Path(plan['state']).resolve()
    # Exclusive creation prevents stale switch/restore signals arming another run.
    state.mkdir(mode=0o700, parents=True, exist_ok=False)
    port = int(plan['port'])
    original = int(plan['original_pid'])
    seconds = float(plan.get('seconds', 1800))
    startup = float(plan.get('startup_seconds', 20))
    shutdown = float(plan.get('shutdown_seconds', 10))
    if not 1 <= port <= 65535 or min(seconds, startup, shutdown) <= 0:
        raise ValueError('Invalid port or positive durations')
    for name in ('original_argv', 'candidate_argv'):
        if not isinstance(plan[name], list) or not plan[name] or not all(isinstance(v, str) for v in plan[name]):
            raise ValueError('Launch commands must be explicit argument arrays')
    cwd = str(Path(plan['cwd']).resolve(strict=True))

    owned = {}

    def status(value, detail=''):
        temporary = state / 'status.tmp'
        temporary.write_text(json.dumps({'state': value, 'detail': detail, 'controller_pid': os.getpid(), 'owned': owned}))
        temporary.replace(state / 'status.json')

    def signal_present(name):
        return (state / name).exists()

    def launch(name, argv):
        # New session: neither the GUI process group nor a finishing controller
        # owns the restored server's lifetime. No inherited pipe waits for exit.
        with (state / (name + '.log')).open('ab', buffering=0) as log:
            process = subprocess.Popen(argv, cwd=cwd, stdin=subprocess.DEVNULL,
                                       stdout=log, stderr=log, start_new_session=True)
            owned[name] = process.pid
            return process

    def restore():
        if listeners(port):
            status('refused', 'Unexpected listener; restoration did not start')
            return
        try:
            restored = launch('restored', plan['original_argv'])
        except (OSError, ValueError):
            status('restoration-uncertain', 'Original launch refused; inspect the approved command and private logs')
            return
        until = time.monotonic() + startup
        while time.monotonic() < until:
            owners = listeners(port)
            if owners and all(in_group(pid, restored.pid) for pid in owners):
                status('restored', 'Owned listener observed; remote access and inference are not verified')
                return
            if owners or restored.poll() is not None:
                break
            time.sleep(.1)
        # Do not announce recovery on a timer or process creation alone.
        status('restoration-uncertain', 'No verified restored listener; inspect private restored.log')

    original_identity = identity(original)
    if not original_identity or listeners(port) != {original}:
        status('refused', 'Original PID is not the sole expected listener')
        return
    status('armed', 'Original untouched; explicit switch or cancel signal required')
    while not signal_present('switch'):
        if signal_present('cancel') or signal_present('restore'):
            status('cancelled', 'Original untouched')
            return
        if identity(original) != original_identity or listeners(port) != {original}:
            status('refused', 'Original identity or listener changed before switch')
            return
        time.sleep(.1)
    # Revalidate immediately before the only signal ever sent to the original.
    if signal_present('cancel') or identity(original) != original_identity or listeners(port) != {original}:
        status('refused', 'Cancelled or original identity/listener changed')
        return
    deadline = time.monotonic() + seconds
    os.kill(original, signal.SIGTERM)
    status('stopping-original', 'Graceful shutdown only; controller stays alive if shutdown is slow')
    while identity(original) == original_identity:
        # We cannot abandon this wait after requesting termination: the old
        # server may exit late, leaving the port empty after the controller dies.
        time.sleep(.1)
    if listeners(port):
        status('refused', 'Unexpected owner after original shutdown; nothing was killed')
        return
    if signal_present('restore') or time.monotonic() >= deadline:
        restore()
        return
    try:
        candidate = launch('candidate', plan['candidate_argv'])
    except (OSError, ValueError):
        status('candidate-start-failed', 'Restoring original after candidate launch refusal')
        restore()
        return
    status('starting-candidate')
    ready_until = time.monotonic() + startup
    while time.monotonic() < min(ready_until, deadline):
        owners = listeners(port)
        if owners and all(in_group(pid, candidate.pid) for pid in owners):
            status('candidate', 'Rollback timer active')
            break
        if owners or candidate.poll() is not None or signal_present('restore'):
            break
        time.sleep(.1)
    while candidate.poll() is None and time.monotonic() < deadline and not signal_present('restore'):
        owners = listeners(port)
        if owners and any(not in_group(pid, candidate.pid) for pid in owners):
            break
        if time.monotonic() >= ready_until and not owners:
            break
        time.sleep(.1)
    owners = listeners(port)
    if any(not in_group(pid, candidate.pid) for pid in owners):
        status('refused', 'Unexpected listener; never terminate its owner')
        return
    status('stopping-candidate')
    # This group was created by us. The original and arbitrary listeners are
    # never force-killed. Reaping is bounded even if this candidate ignores TERM.
    try:
        os.killpg(candidate.pid, signal.SIGTERM)
    except ProcessLookupError:
        pass
    try:
        candidate.wait(timeout=shutdown)
    except subprocess.TimeoutExpired:
        os.killpg(candidate.pid, signal.SIGKILL)
        candidate.wait(timeout=shutdown)
    if listeners(port):
        # A child in our group can outlive an exited leader. Kill only that
        # owned group, then still recheck ownership before starting restoration.
        if all(in_group(pid, candidate.pid) for pid in listeners(port)):
            try:
                os.killpg(candidate.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            end = time.monotonic() + shutdown
            while listeners(port) and time.monotonic() < end:
                time.sleep(.1)
    restore()


if __name__ == '__main__':
    try:
        run(sys.argv[1])
    except Exception as error:
        # Plans and exceptions may contain private paths/command arguments.
        print('Recovery controller failed: ' + type(error).__name__ + '; restoration is not confirmed', file=sys.stderr)
        sys.exit(1)
