"""Check evidence-verifier identity controls without altering original receipts."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

HERE = Path(__file__).resolve().parent
SOURCE = 'a347ce31511c4b124e486bb41ef98c07bd17cec5:solutions/edition-2/main'


def hashes(directory):
    return {str(p.relative_to(directory)): hashlib.sha256(p.read_bytes()).hexdigest()
            for p in directory.rglob('*') if p.is_file()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--repository', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    args = parser.parse_args()
    reports = []
    original = hashes(HERE)
    with tempfile.TemporaryDirectory(prefix='ensemble-ch03-verifier-') as temp:
        work = Path(temp)
        evidence = work/'receipts'
        shutil.copytree(HERE, evidence)
        binary = work/'recorded-binary'
        shutil.copy2(args.binary, binary)
        wrong_binary = work/'wrong-binary'
        marker = work/'must-not-execute'
        wrong_binary.write_text(f'#!/bin/sh\ntouch "{marker}"\n')
        wrong_binary.chmod(0o700)

        def invoke(name, executable=binary, source=SOURCE):
            output = work/name
            command = [sys.executable, str(HERE/'verify-receipts.py'), '--binary', str(executable),
                       '--source', source, '--repository', str(args.repository.resolve()),
                       '--evidence-dir', str(evidence), '--output-dir', str(output)]
            before = hashes(evidence)
            result = subprocess.run(command, cwd=work, capture_output=True, text=True, timeout=60)
            unchanged = hashes(evidence) == before
            reports.append({'case':name,'exit':result.returncode,'stderr':result.stderr,
                            'evidence_byte_identical':unchanged,'output_exists':output.exists()})
            assert unchanged, name
            return result, output

        result, output = invoke('wrong-executable', executable=wrong_binary)
        assert result.returncode != 0 and 'executable hash' in result.stderr
        assert not marker.exists() and not output.exists()
        result, output = invoke('wrong-source', source='0'*40+':solutions/edition-2/main')
        assert result.returncode != 0 and 'immutable initial commit' in result.stderr
        assert not output.exists()
        launch = evidence/'gemini-main/launch.json'
        launch_bytes = launch.read_bytes()
        altered = json.loads(launch_bytes)
        altered['binary_sha256'] = '0'*64
        launch.write_text(json.dumps(altered))
        result, output = invoke('changed-launch')
        assert result.returncode != 0 and 'immutable receipt' in result.stderr
        assert not output.exists()
        launch.write_bytes(launch_bytes)
        result, output = invoke('original-binary-and-revision')
        assert result.returncode == 0, result.stderr
        for artifact in output.rglob('*'):
            if artifact.is_file() and artifact.name != 'reconstruction.json':
                assert artifact.read_bytes() == (evidence/artifact.relative_to(output)).read_bytes(), artifact
        reconstruction = json.loads((output/'reconstruction.json').read_text())
        assert reconstruction['source'] == SOURCE
        assert reconstruction['kind'] == 'offline reconstruction; not intercepted live HTTP'
        reports[-1]['all_derived_bytes_equal_initial'] = True
    assert hashes(HERE) == original
    report = {'source':SOURCE,'verifier_sha256':hashlib.sha256((HERE/'verify-receipts.py').read_bytes()).hexdigest(),
              'controls_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
              'binary_sha256':hashlib.sha256(args.binary.read_bytes()).hexdigest(),
              'original_evidence_unchanged':True,'cases':reports}
    args.report.write_text(json.dumps(report,indent=2)+'\n')
    print(json.dumps(report,indent=2))


if __name__ == '__main__':
    main()
