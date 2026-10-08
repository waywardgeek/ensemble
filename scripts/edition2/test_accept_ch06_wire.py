"""Controls for the independent receipt assertions, not a provider/parser model."""
import copy
import unittest

from accept_ch06_wire import check_result


def accepted():
    part = dict(type="text", text="Hello.")
    response = dict(parts=[part], usage=dict(input=10, cache_write=0, cache_read=0, output=2))
    return dict(exit_code=0, history=[dict(type="response_ended", seq=8, response=response)], records=[
        dict(observation=dict(kind="model_begin", delivery="stream")),
        dict(observation=dict(kind="part_final", part_index=0, response_seq=8, part=copy.deepcopy(part))),
        dict(observation=dict(kind="model_end", accepted=True)),
        dict(completion=dict(outcome="success")),
    ])


class WireReceiptControls(unittest.TestCase):
    def test_positive(self):
        self.assertEqual(check_result(accepted())["parts"], [dict(type="text", text="Hello.")])

    def test_single_defects_reach_their_intended_refusal(self):
        def final_text(result):
            result["records"][1]["observation"]["part"]["text"] = "different"

        def sequence(result):
            result["records"][1]["observation"]["response_seq"] = 7

        def usage(result):
            result["history"][0]["response"]["usage"]["input"] = 20

        def duplicate(result):
            result["records"].append(dict(completion=dict(outcome="success")))

        for change, reason in [(final_text, "finals differ from persisted typed parts"),
                               (sequence, "final mapping differs from actual sequence"),
                               (usage, "normalized usage differs"),
                               (duplicate, "request did not settle exactly once")]:
            with self.subTest(change=change.__name__):
                fixture = accepted()
                check_result(fixture)
                change(fixture)
                with self.assertRaisesRegex(AssertionError, "^" + reason + "$"):
                    check_result(fixture)

    def test_failed_operation_cannot_retain_a_tool_effect_record(self):
        fixture = dict(exit_code=1, history=[], records=[
            dict(observation=dict(kind="model_begin", delivery="stream")),
            dict(observation=dict(kind="model_end", accepted=False)),
            dict(completion=dict(outcome="error")),
        ])
        self.assertIsNone(check_result(fixture, False))
        fixture["history"].append(dict(type="tool_called"))
        with self.assertRaisesRegex(AssertionError, "^failed operation dispatched a tool$"):
            check_result(fixture, False)


if __name__ == "__main__":
    unittest.main()
