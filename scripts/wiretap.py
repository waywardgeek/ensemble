#!/usr/bin/env python3
"""A logging pass-through proxy for vendor HTTP APIs.

Point a solution at it with, e.g.:

    OPENAI_BASE_URL=http://127.0.0.1:8899 scripts/live.sh 2 openai

and every request body and response body is appended, verbatim, to the log
named by $WIRETAP_LOG (default /tmp/wiretap.log). Nothing is rewritten: the
bytes the solution sends are the bytes the vendor sees, and the bytes the
vendor returns are the bytes the solution parses. That is the point — this
exists to settle arguments about what is actually on the wire.

    WIRETAP_PORT      listen port            (default 8899)
    WIRETAP_UPSTREAM  where to forward to    (default https://api.openai.com)
    WIRETAP_LOG       append transcript here (default /tmp/wiretap.log)
"""

import http.server
import json
import os
import socketserver
import sys
import time
import urllib.error
import urllib.request

PORT = int(os.environ.get("WIRETAP_PORT", "8899"))
UPSTREAM = os.environ.get("WIRETAP_UPSTREAM", "https://api.openai.com").rstrip("/")
LOGPATH = os.environ.get("WIRETAP_LOG", "/tmp/wiretap.log")

# Headers we must not forward verbatim: hop-by-hop, or ones urllib will set.
SKIP_REQ = {"host", "content-length", "accept-encoding", "connection"}
SKIP_RESP = {"content-length", "content-encoding", "transfer-encoding", "connection"}


def log(kind, payload):
    with open(LOGPATH, "a") as f:
        f.write("\n===== %s %s =====\n" % (time.strftime("%H:%M:%S"), kind))
        f.write(payload)
        f.write("\n")
        f.flush()


class Handler(http.server.BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def log_message(self, fmt, *args):  # silence per-request stderr noise
        pass

    def do_POST(self):
        self._proxy("POST")

    def do_GET(self):
        self._proxy("GET")

    def _proxy(self, method):
        n = int(self.headers.get("Content-Length") or 0)
        body = self.rfile.read(n) if n else None

        log("REQUEST %s %s" % (method, self.path),
            body.decode("utf-8", "replace") if body else "(no body)")

        req = urllib.request.Request(UPSTREAM + self.path, data=body, method=method)
        for k, v in self.headers.items():
            if k.lower() not in SKIP_REQ:
                req.add_header(k, v)
        req.add_header("Accept-Encoding", "identity")

        try:
            with urllib.request.urlopen(req, timeout=300) as r:
                status, hdrs, rbody = r.status, r.headers, r.read()
        except urllib.error.HTTPError as e:          # 4xx/5xx: the interesting case
            status, hdrs, rbody = e.code, e.headers, e.read()
        except Exception as e:                        # network failure
            log("PROXY-ERROR", repr(e))
            self.send_response(502)
            self.send_header("Content-Length", "0")
            self.end_headers()
            return

        log("RESPONSE %d" % status, rbody.decode("utf-8", "replace"))

        self.send_response(status)
        for k, v in hdrs.items():
            if k.lower() not in SKIP_RESP:
                self.send_header(k, v)
        self.send_header("Content-Length", str(len(rbody)))
        self.end_headers()
        self.wfile.write(rbody)


class Server(socketserver.ThreadingMixIn, http.server.HTTPServer):
    daemon_threads = True
    allow_reuse_address = True


if __name__ == "__main__":
    print("wiretap: 127.0.0.1:%d -> %s, logging to %s" % (PORT, UPSTREAM, LOGPATH),
          file=sys.stderr, flush=True)
    Server(("127.0.0.1", PORT), Handler).serve_forever()
