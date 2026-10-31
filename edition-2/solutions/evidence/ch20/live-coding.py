"""One bounded real Ensemble coding turn; retain only sanitized task evidence."""
import difflib
import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import tempfile
import time

repo = pathlib.Path.cwd()
evidence = repo / "solutions/edition-2/evidence/ch20"
binary = pathlib.Path("/Users/bill/ensemble-ch19-signin-01v0rmp3/ensemble-ch20")
auth = pathlib.Path("/Users/bill/ensemble-ch19-signin-01v0rmp3/auth")
private = []


def remember_private():
    # Only the explicitly authorized profile; never print its identity or tokens.
    value = json.loads((auth / "auth.json").read_text())
    for key in ("host_id", "active", "pending"):
        if value.get(key):
            private.append(value[key].encode())
    for profile in value["profiles"].values():
        for key in ("access_token", "refresh_token", "id_token", "email", "subject", "client_id"):
            if profile.get(key):
                private.append(profile[key].encode())


def save(name, raw):
    for secret in set(private):
        raw = raw.replace(secret, b"[REDACTED]")
    assert not any(secret in raw for secret in private)
    (evidence / name).write_bytes(raw)


def complete(path):
    if not path.exists():
        return False
    for line in path.read_text().splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        if event.get("type") == "response_ended":
            parts = event.get("response", {}).get("parts", [])
            if any(p.get("type") == "text" for p in parts) and not any(p.get("type") == "tool_call" for p in parts):
                return True
    return False


build = json.loads((evidence / "build.json").read_text())
assert hashlib.sha256(binary.read_bytes()).hexdigest() == build["binary_sha256"]
remember_private()
work = pathlib.Path(tempfile.mkdtemp(prefix="ensemble-ch20-coding-"))
for path in (evidence / "task-before").iterdir():
    shutil.copy2(path, work / path.name)
# A narrow, actual source repair. The test file is an immutable task contract.
prompt = ("Repair aggregate.py in this scratch directory. Read aggregate.py and test_aggregate.py using read_file, "
          "then run python3 -m unittest -v with run_command before editing to observe the failures. "
          "Make the smallest clear fix: totals(path) must return exact decimal currency totals grouped by merchant, "
          "support normal CSV quoting including embedded commas, escaped quotes and newlines, skip blank rows, "
          "and retain the CLI's sorted tab-separated two-decimal output. Use only the Python standard library. "
          "Do not change test_aggregate.py or read files outside this workspace. Edit the actual source using a file tool, "
          "run the unchanged tests again, and give a concise report of the bug, fix and real results. "
          "Do not add unrelated features or just describe a patch. Briefly report progress before editing.")
(evidence / "prompt.txt").write_text(prompt + "\n")
baseline = subprocess.run(["python3", "-m", "unittest", "-v"], cwd=work, text=True, capture_output=True)
save("task-baseline.txt", (baseline.stdout + baseline.stderr).encode())
assert baseline.returncode != 0
# Discard provider, skill, workspace and log overrides inherited from other work.
env = {k: v for k, v in os.environ.items() if not k.endswith("API_KEY") and not k.startswith(("LLM_", "OPENAI_", "EN_", "CH02_", "CH07_", "ENSEMBLE_"))}
env.update(LLM_VENDOR="openai", LLM_MODEL="gpt-6-astra", LLM_BASE_URL="https://api.openai.com",
           OPENAI_BILLING_MODE="chatgpt", OPENAI_OIDC_ISSUER="https://auth.openai.com",
           OPENAI_REASONING_EFFORT="high", OPENAI_REASONING_SUMMARY="detailed", EN_HOME=str(auth),
           EN_RECALL_DISABLED="1", EN_PRIMARY_SKILL="", EN_SKILLS_DIR="", CH02_LOG=str(work / "events.jsonl"))
start = time.monotonic()
closed = timed_out = False
with (work / "stdout.txt").open("wb") as stdout, (work / "stderr.txt").open("wb") as stderr:
    child = subprocess.Popen([str(binary), "chat"], cwd=work, env=env, stdin=subprocess.PIPE, stdout=stdout, stderr=stderr)
    child.stdin.write((prompt + "\n").encode())
    child.stdin.flush()
    while child.poll() is None and time.monotonic() - start < 240:
        if not closed and (complete(work / "events.jsonl") or "request failed:" in (work / "stderr.txt").read_text()):
            child.stdin.close()
            closed = True
        time.sleep(.1)
    if child.poll() is None:
        child.kill()
        timed_out = True
    exit_code = child.wait()
remember_private()
for name in ("stdout.txt", "stderr.txt", "api.log", "events.jsonl", "save.json"):
    if (work / name).exists():
        save("live-" + name, (work / name).read_bytes())
(evidence / "task-after").mkdir(exist_ok=True)
for name in ("aggregate.py", "test_aggregate.py"):
    save("task-after/" + name, (work / name).read_bytes())
final = subprocess.run(["python3", "-m", "unittest", "-v"], cwd=work, text=True, capture_output=True)
save("task-final.txt", (final.stdout + final.stderr).encode())
before = (evidence / "task-before/aggregate.py").read_text()
after = (work / "aggregate.py").read_text()
save("task.patch", "".join(difflib.unified_diff(before.splitlines(True), after.splitlines(True), fromfile="before/aggregate.py", tofile="after/aggregate.py")).encode())
result = dict(source_commit=build["source_commit"], main_tree=build["main_tree"],
              binary=str(binary), binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(), workspace=str(work),
              model="gpt-6-astra", reasoning_effort="high", billing="explicit chatgpt", api_key_configured=False,
              endpoint_from_config_and_source="https://api.openai.com/v1/responses",
              endpoint_independently_captured=False, baseline_exit_code=baseline.returncode,
              final_exit_code=final.returncode, cli_exit_code=exit_code, timed_out=timed_out,
              elapsed_seconds=round(time.monotonic() - start, 2), stdin_closed_on_answer_or_error=closed,
              tests_unchanged=(evidence / "task-before/test_aggregate.py").read_bytes() == (work / "test_aggregate.py").read_bytes(),
              source_changed=before != after, dedicated_profile_retained=True, logout=False)
(evidence / "live-result.json").write_text(json.dumps(result, indent=2) + "\n")
print(json.dumps(result, indent=2))
