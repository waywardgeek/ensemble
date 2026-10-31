"""Behavioral checks for exact totals and ordinary CSV quoting; no test doubles."""
import subprocess
import sys
import tempfile
import unittest
from decimal import Decimal
from pathlib import Path

from aggregate import totals


class TotalsTests(unittest.TestCase):
    """Keep input data real so parsing, arithmetic and CLI checks compose."""

    def check_csv(self, text, expected):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "purchases.csv"
            path.write_text(text)
            actual = totals(path)
            self.assertEqual(actual, expected)
            self.assertTrue(all(isinstance(value, Decimal) for value in actual.values()))

    def test_groups_merchants(self):
        self.check_csv("merchant,amount\nCafe,2.25\nBooks,4.50\nCafe,1.75\n",
                       {"Cafe": Decimal("4.00"), "Books": Decimal("4.50")})

    def test_decimal_exactness_and_refund(self):
        self.check_csv("merchant,amount\nCafe,0.10\nCafe,0.20\nCafe,-0.05\n",
                       {"Cafe": Decimal("0.25")})

    def test_quoted_comma_and_quote(self):
        self.check_csv('merchant,amount\n"Books, Inc.",3.20\n"Pat""s Cafe",1.10\n',
                       {"Books, Inc.": Decimal("3.20"), 'Pat"s Cafe': Decimal("1.10")})

    def test_multiline_quoted_merchant(self):
        self.check_csv('merchant,amount\n"Market\nAnnex",2.50\n',
                       {"Market\nAnnex": Decimal("2.50")})

    def test_blank_rows_and_empty_report(self):
        self.check_csv("merchant,amount\n\nCafe,1.00\n\n", {"Cafe": Decimal("1.00")})
        self.check_csv("merchant,amount\n", {})

    def test_cli_sorted_fixed_precision(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "purchases.csv"
            path.write_text('merchant,amount\n"Zed, Shop",2.5\nAlpha,0.10\nAlpha,0.20\n')
            run = subprocess.run([sys.executable, "aggregate.py", str(path)],
                                 text=True, capture_output=True)
            self.assertEqual(run.returncode, 0, run.stderr)
            self.assertEqual(run.stdout, "Alpha\t0.30\nZed, Shop\t2.50\n")


if __name__ == "__main__":
    unittest.main()
