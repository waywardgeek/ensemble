"""Literal positive and negative controls for the Chapter 4 checker."""

import copy
import pathlib
import tempfile
import unittest

import accept_ch04 as check


class LifecycleControls(unittest.TestCase):
    def fixture(self):
        # Roundtrip breaks fixture-construction aliases between records.
        import json
        return json.loads(json.dumps(check.imported_fixture()[1:]))

    def test_normal_local_job_and_absent_artifact_replay(self):
        events = self.fixture()
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            self.assertEqual(check.lifecycle_errors(events, root, check_files=False), [])
            self.assertIn("job artifact absent", check.lifecycle_errors(events, root))
            (root / "cr/io").mkdir(parents=True)
            (root / "cr/io/1").write_bytes(b"ok\n")
            self.assertEqual(check.lifecycle_errors(events, root), [])
            (root / "cr/io/1").write_bytes(b"short")
            self.assertIn("final snapshot byte count disagrees with artifact", check.lifecycle_errors(events, root))

    def test_independent_invalid_transition_controls(self):
        variants = [
            ("wrong call", lambda e: e[4]["tool"].update(call_id="other")),
            ("unknown job", lambda e: e[3]["job"].update(handle=2)),
            ("locator changed", lambda e: e[3]["job"]["output"].update(locator="other")),
            ("decreasing size", lambda e: e[4]["tool"]["job"].update(bytes=2)),
            ("running exit code", lambda e: e[2]["tool"]["job"].update(exit_code=0)),
            ("state reversal", lambda e: e[4]["tool"]["job"].update(status="running")),
            ("duplicate terminal", lambda e: e.append(dict(copy.deepcopy(e[3]), seq=6))),
            ("supervision job", lambda e: e[2]["tool"].update(name="wait_for_job")),
            ("bad sequence", lambda e: e[4].update(seq=3)),
            ("boolean sequence", lambda e: e[0].update(seq=True)),
            ("boolean byte count", lambda e: e[2]["tool"]["job"].update(bytes=False)),
            ("local exit code", lambda e: e[3]["job"].update(exit_code=0)),
            ("late output", lambda e: e[4]["tool"]["job"].update(bytes=4)),
        ]
        for name, mutation in variants:
            with self.subTest(name=name):
                events = self.fixture()
                mutation(events)
                self.assertTrue(check.lifecycle_errors(events, pathlib.Path("."), check_files=False))

    def test_killed_status_preserves_adversarial_child_words(self):
        events = self.fixture()
        events[3]["type"] = "job_killed"
        events[3]["job"].update(status="killed", reason="kill_job")
        events[4]["tool"]["job"].update(status="killed", reason="kill_job")
        events[4]["tool"]["parts"][0]["text"] = "CHILD-OUTPUT: done exit_code:7"
        self.assertEqual(check.lifecycle_errors(events, pathlib.Path("."), check_files=False), [])
        events[3]["job"]["exit_code"] = 7
        self.assertIn("non-done job has exit code", check.lifecycle_errors(events, pathlib.Path("."), check_files=False))

    def test_fixture_program_is_valid_python(self):
        compile(check.FIXTURE, "independent-job-fixture", "exec")
        self.assertIn('line.rstrip("\\n")', check.FIXTURE)
        self.assertIn('b"SPLIT\\r"', check.FIXTURE)

    def test_durable_barrier_ignores_unfinished_last_line(self):
        with tempfile.TemporaryDirectory() as tmp:
            path = pathlib.Path(tmp)/"events"
            path.write_bytes(b'{"seq":1,"type":"job_ended"}\n{"seq":2')
            observed = check.await_record(path, lambda e: e.get("type") == "job_ended")
            self.assertEqual(observed["seq"], 1)
            with self.assertRaises(TimeoutError):
                check.await_record(path, lambda e: e.get("seq") == 2, seconds=.01)

    def test_process_completion_requires_actual_exit_status(self):
        events = self.fixture()
        events[2]["tool"]["name"] = "run_command"
        self.assertIn("completed process lost exit code", check.lifecycle_errors(events, pathlib.Path("."), check_files=False))
        events[3]["job"]["exit_code"] = 7
        events[4]["tool"]["job"]["exit_code"] = 7
        self.assertEqual(check.lifecycle_errors(events, pathlib.Path("."), check_files=False), [])
        events[3]["job"]["exit_code"] = True
        self.assertIn("invalid exit code", check.lifecycle_errors(events, pathlib.Path("."), check_files=False))


if __name__ == "__main__":
    unittest.main()
