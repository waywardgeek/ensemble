#!/usr/bin/env python3
"""Retain every Chapter 5 workflow assertion with a streaming fake-wire adapter."""
import argparse
import http.server
import json
import pathlib
import queue
import threading

import accept_ch05_clients as prior
from accept_ch06_clients import frame


def stream(value):
    if 'content' in value:
        out = frame(dict(type='message_start', message=dict(model=value['model'], usage=value['usage'])))
        for index, part in enumerate(value['content']):
            out += frame(dict(type='content_block_start', index=index, content_block=part))
            out += frame(dict(type='content_block_stop', index=index))
        out += frame(dict(type='message_delta', delta=dict(stop_reason=value['stop_reason']), usage=value['usage']))
        return out + frame(dict(type='message_stop'))
    if 'choices' in value:
        choice = value['choices'][0]
        delta = dict(choice['message'])
        if 'tool_calls' in delta:
            delta['tool_calls'] = [dict(index=i, **call) for i, call in enumerate(delta['tool_calls'])]
        return (frame(dict(model=value['model'], choices=[dict(index=0, delta=delta, finish_reason=choice['finish_reason'])]))
                + frame(dict(model=value['model'], choices=[], usage=value['usage'])) + b'data: [DONE]\n\n')
    return frame(value)


class Server:
    def __init__(self):
        self.calls = queue.Queue()
        self.stopped = threading.Event()
        outer = self

        class Handler(http.server.BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def do_POST(self):
                body = self.rfile.read(int(self.headers['Content-Length']))
                release = queue.Queue()
                outer.calls.put((body, release))
                while not outer.stopped.is_set():
                    try:
                        value = release.get(timeout=.1)
                        break
                    except queue.Empty:
                        continue
                else:
                    return
                data = stream(value)
                try:
                    self.send_response(200)
                    self.send_header('Content-Type', 'text/event-stream')
                    self.end_headers()
                    self.wfile.write(data)
                    self.wfile.flush()
                except (BrokenPipeError, ConnectionResetError):
                    pass

        self.server = http.server.ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        self.server.daemon_threads = True
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()

    def get(self):
        return self.calls.get(timeout=5)

    def close(self):
        self.stopped.set()
        self.server.shutdown()
        self.server.server_close()
        self.thread.join()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('binary', type=pathlib.Path)
    p.add_argument('--workflow', action='store_true')
    a = p.parse_args()
    prior.Server = Server
    rows = []
    for vendor in prior.VENDORS:
        try:
            prior.workflow(a.binary.resolve(strict=True), vendor)
            errors = []
        except Exception as error:
            errors = [str(error)]
        rows.append(dict(id=vendor + '/workflow', passed=not errors, details=errors))
    result = dict(passed=all(x['passed'] for x in rows), checks=rows)
    print(json.dumps(result, indent=2))
    return 0 if result['passed'] else 1


if __name__ == '__main__':
    raise SystemExit(main())
