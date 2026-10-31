"""Sum a CSV of merchant,amount purchases for a small spending report."""
import csv
import sys
from decimal import Decimal
from pathlib import Path


def totals(path):
    """Return each merchant's exact total in currency units."""
    result = {}
    with Path(path).open(newline="") as source:
        rows = csv.reader(source)
        next(rows, None)  # Skip the merchant,amount header.
        for row in rows:
            if not row:
                continue
            merchant, amount = row
            result[merchant] = result.get(merchant, Decimal(0)) + Decimal(amount)
    return result


if __name__ == "__main__":
    for merchant, total in sorted(totals(sys.argv[1]).items()):
        print(f"{merchant}\t{total:.2f}")
