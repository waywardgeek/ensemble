"""Check the real GatewayHost's offline migration/recovery command modes."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dotnet", required=True, type=Path)
    parser.add_argument("--host", required=True, type=Path)
    args = parser.parse_args()
    dotnet, host = args.dotnet.resolve(strict=True), args.host.resolve(strict=True)
    with tempfile.TemporaryDirectory(prefix="gateway-host-modes-") as temporary:
        root = Path(temporary)
        data = root / "state"
        config = root / "config.json"
        # Deliberately not a listen endpoint: offline modes must not bind it.
        config.write_text(json.dumps({"gateway": {"listen": "not-an-endpoint", "data_dir": str(data)},
                                      "groups": {}, "agents": {}}))
        environment = {"PATH": os.environ.get("PATH", ""), "HOME": temporary, "TMPDIR": temporary,
                       "DOTNET_CLI_HOME": temporary, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": "1",
                       "DOTNET_CLI_TELEMETRY_OPTOUT": "1", "DOTNET_GENERATE_ASPNET_CERTIFICATE": "false"}

        def run(*flags, succeeds):
            result = subprocess.run([str(dotnet), str(host), "--config", str(config), *flags],
                                    cwd=root, env=environment, capture_output=True, timeout=15)
            if (result.returncode == 0) != succeeds:
                raise AssertionError("Unexpected host exit: " + result.stderr.decode(errors="replace"))

        run("--print-key", "--recover-backup", succeeds=False)
        assert not data.exists(), "Conflicting flags created identity/state"
        run("--migrate-legacy", "--recover-backup", succeeds=False)
        assert not data.exists(), "Conflicting recovery flags created state"
        run("--print-key", succeeds=True)
        assert (data / "identity.key").is_file()
        assert not (data / "gateway-state.json").exists(), "Key output initialized router state"
        queues = data / "queues"
        queues.mkdir()
        legacy = data / "state.json"
        original = b'{"groups":{},"children":{},"memberships":{},"contactable":{}}'
        legacy.write_bytes(original)
        run(succeeds=False)
        assert not (data / "gateway-state.json").exists(), "Ordinary host silently migrated"
        run("--migrate-legacy", succeeds=True)
        primary = data / "gateway-state.json"
        assert json.loads(primary.read_text())["version"] == 1
        assert legacy.read_bytes() == original, "Host migration modified original"
        committed = primary.read_bytes()
        run("--migrate-legacy", succeeds=False)
        assert primary.read_bytes() == committed, "Repeated migration changed primary"
        primary.write_bytes(b'{broken')
        run(succeeds=False)
        assert primary.read_bytes() == b'{broken', "Host silently repaired corruption"
        run("--recover-backup", succeeds=True)
        assert primary.read_bytes() == committed, "Host recovery changed validated backup content"
        assert any(path.read_bytes() == b'{broken' for path in data.glob('gateway-state.recovery-*.json'))
        (data / "gateway-state.backup.json").write_bytes(b'{invalid')
        run("--recover-backup", succeeds=False)
        assert primary.read_bytes() == committed, "Bad backup overwrote valid primary"
        print("PASS real host: flags before effects, key-only, explicit migration, corrupt refusal, recovery, originals")


if __name__ == "__main__":
    main()
