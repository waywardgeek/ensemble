"""Bounded renderer HTTP contract checks against the no-inference GUI fixture.
Usage: python3 renderer-http-check.py DOTNET GUITESTS_DLL
"""
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request


def main():
    with tempfile.TemporaryDirectory(prefix="ensemble-renderer-http-") as folder:
        home = Path(folder)
        with (home / "host.log").open("wb") as log:
            process = subprocess.Popen([sys.argv[1], sys.argv[2], "--artifact-renderer-fixture", folder], stdout=log, stderr=subprocess.STDOUT)
            try:
                deadline = time.monotonic() + 30
                while not (home / "url").exists():
                    if process.poll() is not None or time.monotonic() > deadline:
                        raise RuntimeError("Fixture failed readiness: " + (home / "host.log").read_text())
                    time.sleep(.05)
                base = (home / "url").read_text().strip()
                origin = base.rstrip("/")

                def request(path, body=None, expected=200, origin_value=origin, header=True):
                    headers = {}
                    if origin_value is not None:
                        headers["Origin"] = origin_value
                    if header:
                        headers["X-Ensemble-Renderer"] = "1"
                    if body is not None:
                        headers["Content-Type"] = "application/json"
                    data = body.encode() if isinstance(body, str) else None if body is None else json.dumps(body).encode()
                    req = urllib.request.Request(base + "renderers/" + path, data=data, headers=headers)
                    try:
                        response = urllib.request.urlopen(req, timeout=10)
                    except urllib.error.HTTPError as error:
                        response = error
                    with response:
                        raw = response.read()
                        assert response.status == expected, (path, response.status, expected, raw[:200])
                        return raw

                def selection():
                    return json.loads(request("selection"))["selection"]

                assert selection() == {"accepted": None, "previous": None}
                package = {"entry": "main.js", "files": {"main.js": "import './helper.js'; export default () => document.createElement('pre');", "helper.js": "export const marker = 'one';"}}
                version = json.loads(request("stage", package))["version"]
                assert selection()["accepted"] is None, "Staging activated a candidate"
                assert request(f"modules/{version}/helper.js").decode() == package["files"]["helper.js"]
                request(f"modules/{version}/absent.js", expected=404)
                request("accept", {"version": version, "expected": None}, expected=403, origin_value=None)
                request("accept", {"version": version, "expected": None}, expected=403, origin_value="https://untrusted.invalid")
                request("accept", {"version": version, "expected": None}, expected=403, header=False)
                assert selection()["accepted"] is None
                request("stage", '{"entry":"main.js","entry":"other.js","files":{}}', expected=400)
                request("stage", {"entry": "../outside.js", "files": {"../outside.js": ""}}, expected=400)
                request("stage", {"entry": "main.js", "files": {"main.js": "x" * (1024 * 1024)}}, expected=413)
                request("accept", {"version": version, "expected": None})
                assert selection()["accepted"] == version
                package["files"]["helper.js"] = "export const marker = 'two';"
                second = json.loads(request("stage", package))["version"]
                assert second != version and selection()["accepted"] == version
                request("accept", {"version": second, "expected": None}, expected=400)
                request("accept", {"version": second, "expected": version})
                assert selection() == {"accepted": second, "previous": version}
                request("rollback", {"expected": second})
                assert selection() == {"accepted": version, "previous": second}
                snapshot = home / "renderers" / "versions" / (second + ".json")
                snapshot.write_bytes(snapshot.read_bytes() + b" ")
                request(f"modules/{second}/main.js", expected=400)
                request("accept", {"version": second, "expected": version}, expected=400)
                assert selection()["accepted"] == version
                print("PASS renderer HTTP: inert stage, exact modules, origin/header guards, bounded/strict parsing, stale refusal, acceptance, rollback, tamper refusal")
            finally:
                (home / "stop").touch()
                try:
                    code = process.wait(timeout=20)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()
                    raise RuntimeError("Fixture required forced cleanup")
                assert code == 0, ("Fixture failed", code, (home / "host.log").read_text())
                print("PASS fixture host joined")


if __name__ == "__main__":
    main()
