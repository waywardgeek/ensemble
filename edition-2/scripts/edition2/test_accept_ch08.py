"""Assertion controls for the partial Chapter 8 checker; no server success claim."""
import copy
import unittest

from accept_ch08 import check_policy, check_preferences, defaults, update


class RecordedSocket:
    def __init__(self, incoming):
        self.incoming = list(incoming)
        self.records = []
        self.sent = []

    def send(self, value):
        self.sent.append(value)

    def until(self, predicate):
        while self.incoming:
            record = self.incoming.pop(0)
            self.records.append(record)
            if predicate(record):
                return record
        raise AssertionError("expected correlated record absent")


class Assertions(unittest.TestCase):
    def test_false_is_required_and_typed(self):
        value = dict(revision=0, preferences=defaults())
        check_preferences(value, 0, defaults())
        for mutation in ("absent", "numeric"):
            broken = copy.deepcopy(value)
            if mutation == "absent":
                broken["preferences"].pop("autoplay")
            else:
                broken["preferences"]["autoplay"] = 0
            with self.assertRaises(AssertionError):
                check_preferences(broken, 0, defaults())

    def test_zero_policy_and_wrong_effective_value(self):
        value = dict(revision=2, persistent=True, max_model_requests=0, effective_max_model_requests=16)
        check_policy(value, 2, 0)
        with self.assertRaisesRegex(AssertionError, "effective"):
            check_policy(value | dict(effective_max_model_requests=0), 2, 0)
        with self.assertRaisesRegex(AssertionError, "unsafe/incomplete"):
            check_policy(value | dict(path="private"), 2, 0)

    def test_policy_frame_may_follow_ack(self):
        change = dict(type="observation", revision=43, observation=dict(kind="policy_changed", execution_policy=dict(revision=1, persistent=True, max_model_requests=2, effective_max_model_requests=2)))
        ack = dict(type="policy_ack", id="p", revision=1, watch_revision=43)
        for records in ([change, ack], [ack, change]):
            client = RecordedSocket(records)
            update(client, "policy", "p", 0, dict(max_model_requests=2), 1, 2)
            self.assertEqual(client.sent[0]["base_revision"], 0)
        with self.assertRaisesRegex(AssertionError, "cut differs"):
            update(RecordedSocket([ack | dict(watch_revision=44), change]), "policy", "p", 0, dict(max_model_requests=2), 1, 2)

    def test_preference_change_precedes_ack(self):
        value = defaults() | dict(autoplay=True)
        change = dict(type="preferences_changed", revision=1, preferences=value)
        ack = dict(type="preferences_ack", id="p", revision=1)
        update(RecordedSocket([change, ack]), "preferences", "p", 0, dict(autoplay=True), 1, value)
        with self.assertRaisesRegex(AssertionError, "must precede"):
            update(RecordedSocket([ack, change]), "preferences", "p", 0, dict(autoplay=True), 1, value)

    def test_no_change_cannot_broadcast(self):
        ack = dict(type="preferences_ack", id="same", revision=2)
        update(RecordedSocket([ack]), "preferences", "same", 2, dict(autoplay=False), 2, defaults())
        change = dict(type="preferences_changed", revision=2, preferences=defaults())
        with self.assertRaisesRegex(AssertionError, "no-change"):
            update(RecordedSocket([change, ack]), "preferences", "same", 2, dict(autoplay=False), 2, defaults())

    def test_boolean_revision_is_not_integer(self):
        value = defaults() | dict(autoplay=True)
        change = dict(type="preferences_changed", revision=1, preferences=value)
        ack = dict(type="preferences_ack", id="p", revision=True)
        with self.assertRaisesRegex(AssertionError, "acknowledgement revision"):
            update(RecordedSocket([change, ack]), "preferences", "p", 0, dict(autoplay=True), 1, value)


if __name__ == "__main__":
    unittest.main()
