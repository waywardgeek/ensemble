#!/usr/bin/env python3
"""Independent public-API Gateway persistence contract grader; no chapter renumbering."""
import argparse
import json
import os
import platform
from pathlib import Path
import shutil
import signal
import subprocess
import tempfile


def run(command, root, environment, seconds):
    # A file avoids pipe deadlock; the group includes the consumer's test host.
    with tempfile.TemporaryFile() as output:
        process = subprocess.Popen(command, cwd=root, env=environment, stdout=output,
                                   stderr=subprocess.STDOUT, start_new_session=True)
        try:
            code = process.wait(timeout=seconds)
        except BaseException:
            os.killpg(process.pid, signal.SIGKILL)
            process.wait()
            raise
        finally:
            # Also remove descendants if an unsuccessful consumer exited early.
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
        output.seek(0)
        text = output.read(262145)
        if len(text) > 262144:
            raise RuntimeError("Grader output exceeded 256 KiB")
        return code, text.decode("utf-8", errors="replace")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dotnet", required=True, type=Path)
    parser.add_argument("--candidate-directory", required=True, type=Path)
    args = parser.parse_args()
    sdk = args.dotnet.resolve(strict=True)
    candidate = args.candidate_directory.resolve(strict=True)
    for name in ("Ensemble.Gateway.dll", "Ensemble.Noise.dll", "runtimes"):
        if not (candidate / name).exists():
            raise ValueError("Candidate missing " + name)
    with tempfile.TemporaryDirectory(prefix="gateway-persistence-grade-") as directory:
        root = Path(directory)
        environment = {"PATH": os.environ.get("PATH", ""), "HOME": directory,
                       "TMPDIR": directory, "DOTNET_CLI_HOME": directory,
                       "DOTNET_SKIP_FIRST_TIME_EXPERIENCE": "1",
                       "DOTNET_CLI_TELEMETRY_OPTOUT": "1",
                       "DOTNET_GENERATE_ASPNET_CERTIFICATE": "false"}
        source = Path(__file__).resolve().parent / "gateway-persistence-grader"
        for name in ("Program.cs", "GatewayPersistence.csproj"):
            shutil.copyfile(source / name, root / name)
        nuget = root / "NuGet.Config"
        nuget.write_text("<configuration><packageSources><clear /></packageSources></configuration>")
        artifacts = root / "artifacts"
        code, output = run([str(sdk), "build", str(root / "GatewayPersistence.csproj"),
                            "-p:CandidateDirectory=" + str(candidate), "--configfile", str(nuget),
                            "--artifacts-path", str(artifacts), "--tl:off", "-v:q"], root, environment, 120)
        if code:
            print(json.dumps({"scope": "gateway-complete-state-v1", "passed": False,
                              "stage": "consumer-build", "details": output}))
            return 1
        binary = artifacts / "bin/GatewayPersistence/debug/GatewayPersistence.dll"
        systems = {"Darwin": ("osx", "libsodium.dylib"), "Linux": ("linux", "libsodium.so")}
        machines = {"arm64": "arm64", "aarch64": "arm64", "x86_64": "x64", "AMD64": "x64"}
        if platform.system() not in systems or platform.machine() not in machines:
            raise ValueError("Grader supports macOS/Linux ARM64/x64 only")
        system, library = systems[platform.system()]
        rid = system + "-" + machines[platform.machine()]
        native = candidate / "runtimes" / rid / "native" / library
        if not native.is_file():
            raise ValueError("Candidate lacks native library for " + rid)
        shutil.copyfile(native, binary.parent / library)
        code, output = run([str(sdk), str(binary), str(root / "state")], root, environment, 120)
        print(json.dumps({"scope": "gateway-complete-state-v1", "passed": code == 0,
                          "stage": "public-contract", "details": output}))
        return 0 if code == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
