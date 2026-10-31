#!/usr/bin/env python3
"""Run unchanged Chapter 5 checks at explicit plain delivery in a disposable copy.

Chapter 6 changes the default, not the older plain semantics. The only fixture
adaptation sets DisableStreaming on Config literals; assertions are retained.
"""
import argparse
import json
import os
import pathlib
import re

import accept_ch05

original_install = accept_ch05.install
original_command = accept_ch05.command


def install(root):
    original_install(root)
    stale = root / 'internal/llm/independent_ch05_stale_test.go'
    # Extend the old deliberately noncooperative transport double to the new
    # Engine-owned operation interface. Keep the original late-result assertions.
    with stale.open('a') as out:
        out.write("""
func (e *auditLateEngine) NewOperation(id, requestID string, config common.Config) common.ModelOperation {
    return &operation{parent:e, id:id, requestID:requestID, config:config, changed:make(chan struct{})}
}
func (e *auditLateEngine) ExchangeOperation(ctx context.Context, op common.ModelOperation, body []byte, config common.Config)(common.ParsedResponse,error){
    return e.ExchangeConfig(ctx,body,config)
}
""")
    for p in root.rglob('independent_ch05_*_test.go'):
        source = p.read_text()
        # Word boundary excludes RequestConfig and other suffixed type names.
        source, count = re.subn(r'\bConfig\{', 'Config{DisableStreaming: true, ', source)
        if count:
            p.write_text(source)
            result = accept_ch05.command(['gofmt', '-w', p], root)
            if result['exit']:
                raise RuntimeError(result)


def command(args, cwd, timeout=180):
    args = list(args)
    if '--workflow' in args and any(str(x).endswith('accept_ch05_clients.py') for x in args):
        args = [pathlib.Path(__file__).with_name('accept_ch06_workflow.py') if str(x).endswith('accept_ch05_clients.py') else x for x in args]
    return original_command(args, cwd, timeout)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('source', type=pathlib.Path)
    a = p.parse_args()
    accept_ch05.install = install
    accept_ch05.command = command
    os.environ['EN_DISABLE_STREAMING'] = '1'
    go_bin = pathlib.Path.home() / 'go/bin'
    if (go_bin / 'dlv').is_file():
        os.environ['PATH'] = str(go_bin) + os.pathsep + os.environ.get('PATH', '')
    result = accept_ch05.evaluate(a.source.resolve(strict=True))
    result['chapter6_adapter'] = __doc__
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
