"""Sum a CSV of merchant,amount purchases for a small spending report."""
import sys
from pathlib import Path


def totals(path):
    """Return each merchant's exact total in currency units."""
    result = {}
    # The first line is the merchant,amount header.
    for line in Path(path).read_text().splitlines()[1:]:
        merchant, amount = line.split(",")
        result[merchant] = result.get(merchant, 0.0) + float(amount)
    return result


if __name__ == "__main__":
    for merchant, total in sorted(totals(sys.argv[1]).items()):
        print(f"{merchant}\t{total:.2f}")
