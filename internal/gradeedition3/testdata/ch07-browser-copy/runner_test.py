"""Runner contracts, separate from candidate GUI/browser acceptance."""
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
import unittest

from fixtures import buffered_fixture
from run_buffered import existing, valid_browser, BROWSER_CHECKS

HERE = Path(__file__).resolve().parent


class RunnerTests(unittest.TestCase):
    def test_fixture_expected_payloads_are_independent_values(self):
        first, second = buffered_fixture(), buffered_fixture()
        self.assertNotEqual(first['marker'], second['marker'])
        self.assertIn('\r\n', first['markdown'])
        self.assertNotIn('\r', ''.join(first['payloads']))
        self.assertEqual(first['payloads'][1], '  quoted\n\n')
        self.assertGreater(len(first['payloads'][2]), 1800)
        self.assertEqual(first['payloads'][2].count('\n'), 200)
        self.assertEqual(first['payloads'][3], '')
        self.assertTrue(first['markdown'].endswith('unfinished ' + first['marker'][7:]))

    def test_receipt_requires_every_named_check_and_terminal_fields(self):
        report = dict(contract='chapter7-copy-browser-component-v1', passed=True,
                      failure=None, checks=list(BROWSER_CHECKS), promptFrames=1,
                      errors=[], blocked=[])
        self.assertTrue(valid_browser(report))
        for key in report:
            changed = dict(report)
            del changed[key]
            self.assertFalse(valid_browser(changed), key)
        for index in range(len(BROWSER_CHECKS)):
            changed = dict(report, checks=BROWSER_CHECKS[:index] + BROWSER_CHECKS[index + 1:])
            self.assertFalse(valid_browser(changed), str(index))

    def test_paths_are_explicit_existing_dependencies(self):
        for value in ['relative', '/no-such-ch07-component']:
            with self.assertRaises(Exception):
                existing(value)
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            self.assertEqual(existing(str(directory), directory=True), directory.resolve())
            with self.assertRaises(Exception):
                existing(str(directory))

    def test_interrupt_during_startup_joins_forced_host_and_clears_environment(self):
        with tempfile.TemporaryDirectory(prefix='ch7-runner-test-') as temporary:
            root = Path(temporary).resolve()
            assets = root / 'assets'
            assets.mkdir()
            (assets / 'index.html').write_text('fixture')
            launcher = root / 'launcher'
            launcher.write_text('#!' + sys.executable + '\n' + '''
import json, os, pathlib, signal, time
signal.signal(signal.SIGTERM, signal.SIG_IGN)
pathlib.Path('host-observed.json').write_text(json.dumps(dict(pid=os.getpid(), env=dict(os.environ))))
while True: time.sleep(.1)
''')
            launcher.chmod(0o700)
            work = root / 'work'
            env = dict(os.environ, CH7_TEST_PRIVATE_SENTINEL='must-not-enter-child',
                       LLM_API_KEY='must-not-enter-child', PYTHONDONTWRITEBYTECODE='1')
            # Node/Playwright are never reached in this startup failure test.
            command = [sys.executable, str(HERE / 'run_buffered.py'),
                       '--launcher', str(launcher), '--node', sys.executable,
                       '--playwright', str(root), '--assets', str(assets), '--work', str(work)]
            process = subprocess.Popen(command, env=env, stdout=subprocess.PIPE,
                                       stderr=subprocess.PIPE, start_new_session=True)
            host_pid = None
            try:
                deadline = time.monotonic() + 5
                observed = work / 'host-observed.json'
                while not observed.exists() and process.poll() is None and time.monotonic() < deadline:
                    time.sleep(.02)
                self.assertTrue(observed.exists(), 'owned host reaches startup')
                value = json.loads(observed.read_text())
                host_pid = value['pid']
                self.assertNotIn('CH7_TEST_PRIVATE_SENTINEL', value['env'])
                self.assertEqual(value['env']['LLM_API_KEY'], 'synthetic-unused')
                self.assertEqual(value['env']['HOME'], str(work / 'home'))
                process.send_signal(signal.SIGTERM)
                output, errors = process.communicate(timeout=15)
                self.assertEqual(process.returncode, 2, errors.decode())
                result = json.loads(output)
                self.assertFalse(result['passed'])
                self.assertTrue(result['interrupted'])
                self.assertTrue(result['cleanup']['host']['forced'])
                self.assertTrue(result['cleanup']['host']['joined'])
                self.assertTrue(result['cleanup']['server_joined'])
                self.assertFalse(result['cleanup']['browser']['started'])
                with self.assertRaises(ProcessLookupError):
                    os.kill(host_pid, 0)
            finally:
                if process.poll() is None:
                    os.killpg(process.pid, signal.SIGKILL)
                process.communicate()
                if host_pid is not None:
                    try:
                        os.killpg(host_pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass


if __name__ == '__main__':
    unittest.main()
