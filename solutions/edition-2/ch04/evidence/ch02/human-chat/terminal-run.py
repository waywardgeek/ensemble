"""Launch a coder-driven terminal session; credentials exist only in child env.

Discovery prints provider model IDs only. This runner never supplies conversation
input: the coder observes the prompt and drives each turn through the live PTY.
"""
import datetime
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.request

HERE = Path(__file__).resolve().parent
MODULE = HERE.parents[2]
KEY_FIELDS = {"anthropic": "directClaudeAPIKey", "openai": "directOpenAIAPIKey", "gemini": "directGeminiAPIKey"}


def credential(vendor):
    settings = json.loads((Path.home() / ".cr/settings.json").read_text())
    return settings[KEY_FIELDS[vendor]]


def discovery():
    result = {"at": datetime.datetime.now(datetime.timezone.utc).isoformat(), "providers": {}}
    for vendor, url in [("anthropic", "https://api.anthropic.com/v1/models?limit=100"), ("openai", "https://api.openai.com/v1/models"), ("gemini", "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000")]:
        key = credential(vendor)
        headers = {"anthropic": {"x-api-key": key, "anthropic-version": "2023-06-01"}, "openai": {"Authorization": "Bearer " + key}, "gemini": {"x-goog-api-key": key}}[vendor]
        with urllib.request.urlopen(urllib.request.Request(url, headers=headers), timeout=30) as response:
            page = json.load(response)
        models = page.get("data", page.get("models", []))
        ids = [m.get("id", m.get("name")) for m in models if vendor != "gemini" or "generateContent" in m.get("supportedGenerationMethods", [])]
        result["providers"][vendor] = {"endpoint": url, "models": ids, "more_pages": bool(page.get("has_more") or page.get("nextPageToken"))}
    (HERE / "discovery.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))


def launch(vendor, model, name, mode):
    run = HERE / name
    run.mkdir()
    binary = Path("/tmp/ensemble-ed2-ch02-human-chat")
    env = os.environ.copy()
    for key in list(env):
        if key.startswith(("LLM_", "ANTHROPIC_", "OPENAI_", "GEMINI_")):
            del env[key]
    env.update(LLM_VENDOR=vendor, LLM_MODEL=model, LLM_API_KEY=credential(vendor), CH02_LOG=str(run / "session.log"))
    args = [str(binary)] + ([] if mode == "default" else [mode])
    sources = {str(p.relative_to(MODULE)): hashlib.sha256(p.read_bytes()).hexdigest() for p in MODULE.rglob("*.go") if not any(x in p.parts for x in ["tmp", ".git"])}
    receipt = {"start": datetime.datetime.now(datetime.timezone.utc).isoformat(), "actor": "Codex student coder, not Bill", "vendor": vendor, "selected_model": model, "mode": mode, "command": args, "transport": "actual tool PTY and macOS script terminal", "source_sha256": sources, "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest()}
    (run / "launch.json").write_text(json.dumps(receipt, indent=2) + "\n")
    code = subprocess.call(["/usr/bin/script", "-q", str(run / "terminal.txt"), *args], env=env)
    receipt["exit_code"] = code
    receipt["end"] = datetime.datetime.now(datetime.timezone.utc).isoformat()
    (run / "launch.json").write_text(json.dumps(receipt, indent=2) + "\n")
    raise SystemExit(code)


if sys.argv[1:] == ["discover"]:
    discovery()
else:
    launch(*sys.argv[1:])
