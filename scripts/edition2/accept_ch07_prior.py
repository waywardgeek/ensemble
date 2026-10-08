#!/usr/bin/env python3
"""Retain Chapter 5/6 assertions with declaration-based immutable-value checking.

The historical checker is unchanged. The Chapter 7 source checker additionally
recognizes literal error sentinels and embedded filesystems only when production
code contains no write or address escape. All behavioral assertions and scores
remain those of the Chapter 6 plain-delivery adapter.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import accept_ch05
import accept_ch06_prior

HERE = Path(__file__).resolve().parent


def command(args, cwd, timeout=180):
    args = [HERE / 'ch07sourcecheck/main.go' if str(arg) == str(HERE / 'ch05sourcecheck/main.go') else arg for arg in args]
    return accept_ch06_prior.command(args, cwd, timeout)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source', type=Path)
    args = parser.parse_args()
    accept_ch05.install = accept_ch06_prior.install
    accept_ch05.command = command
    os.environ['EN_DISABLE_STREAMING'] = '1'
    go_bin = Path.home() / 'go/bin'
    if (go_bin / 'dlv').is_file():
        os.environ['PATH'] = str(go_bin) + os.pathsep + os.environ.get('PATH', '')
    result = accept_ch05.evaluate(args.source.resolve(strict=True))
    result['chapter6_adapter'] = accept_ch06_prior.__doc__
    result['chapter7_adapter'] = __doc__
    result['effective_adapter_files'] = {str(path.relative_to(HERE)): hashlib.sha256(path.read_bytes()).hexdigest() for path in [Path(__file__), HERE/'accept_ch06_prior.py', HERE/'ch07sourcecheck/main.go', HERE/'ch07sourcecheck/main_test.go']}
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
