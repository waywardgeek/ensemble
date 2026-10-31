import unittest

from slugify import slugify


class SlugifyTests(unittest.TestCase):
    def test_mixed_case(self):
        self.assertEqual(slugify("HeLLo WoRLD"), "hello-world")

    def test_repeated_whitespace(self):
        self.assertEqual(slugify("  Hello   \tworld\n\nagain  "), "hello-world-again")

    def test_empty_input(self):
        self.assertEqual(slugify(""), "")


if __name__ == "__main__":
    unittest.main()
