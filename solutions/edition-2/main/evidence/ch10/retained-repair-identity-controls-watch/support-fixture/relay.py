"""Bounded relay and original-body recorder. No discovery or credential reader.

The caller supplies an in-memory header map after the separate release. Local
controls pass dummy values and loopback origins. No redirect, retry or pagination.
"""
import fcntl
import http.client
import http.server
import json
import os
from pathlib import Path
import threading
import time
from urllib.parse import urlsplit

HERE = Path(__file__).resolve().parent
SCHEDULE = json.loads((HERE / 'schedule.json').read_text())


class Budget:
    def __init__(self, path, vendor, clock=time.time):
        self.path, self.vendor, self.clock = Path(path), vendor, clock
        if vendor not in SCHEDULE['providers']: raise ValueError('vendor')

    def reserve(self, step, kind='generation'):
        if kind not in ('generation', 'discovery'): raise ValueError('attempt kind')
        if kind == 'generation' and step not in SCHEDULE['steps']: raise ValueError('step')
        with self.path.open('a+') as f:
            fcntl.flock(f, fcntl.LOCK_EX)
            f.seek(0); rows = [json.loads(line) for line in f if line.strip()]
            if any(x['vendor'] != self.vendor for x in rows): raise ValueError('budget owner mismatch')
            now = self.clock()
            if kind == 'discovery':
                if any(x['kind'] == kind for x in rows): raise ValueError('discovery ceiling')
                row = 'discovery'
            else:
                spec = SCHEDULE['steps'][step]; row = spec['row']
                generations = [x for x in rows if x['kind'] == kind]
                same = [x for x in generations if x['row'] == row]
                if len(generations) >= 22 or len(same) >= SCHEDULE['rows'][row]: raise ValueError('row/provider ceiling')
                if sum(x['step'] == step for x in same) >= spec['cap']: raise ValueError('step ceiling')
                if same and now - same[0]['at'] >= SCHEDULE['row_seconds']: raise ValueError('row deadline')
            item = dict(number=len(rows)+1, vendor=self.vendor, step=step, row=row, kind=kind, at=now)
            # Admission is durable BEFORE transport, including failures/cancellation.
            f.write(json.dumps(item)+'\n'); f.flush(); os.fsync(f.fileno())
            return item


class Relay(http.server.ThreadingHTTPServer):
    daemon_threads = False
    def __init__(self, origin, headers, budget, output, steps, local_only=True, timeout=120):
        url = urlsplit(origin)
        if url.username or url.password or url.path not in ('', '/') or url.query or url.fragment: raise ValueError('origin must be authority only')
        if local_only and (url.scheme != 'http' or url.hostname not in ('127.0.0.1', '::1')): raise ValueError('local relay requires literal loopback')
        if not local_only and url.scheme != 'https': raise ValueError('released upstream requires HTTPS')
        if not set(steps) <= set(SCHEDULE['steps']): raise ValueError('step routes')
        self.origin, self.headers, self.budget = url, headers, budget
        self.output, self.steps, self.timeout = Path(output), set(steps), min(timeout,120)
        self.record_lock = threading.Lock()
        super().__init__(('127.0.0.1', 0), Handler)

    def record(self, item):
        with self.record_lock, (self.output/'transport.jsonl').open('a') as f:
            f.write(json.dumps(item)+'\n'); f.flush()


class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *_): pass
    def do_POST(self):
        owner = self.server
        path = urlsplit(self.path)
        pieces = path.path.split('/', 2)
        if len(pieces) != 3 or pieces[1] not in owner.steps or path.fragment or path.netloc:
            self.send_error(400, 'unbound route'); return
        step = pieces[1]
        # Only adapter generation routes, with no credential query or redirect.
        route = '/' + pieces[2]
        valid = route in ('/v1/messages','/v1/chat/completions','/v1/responses') or (route.startswith('/v1beta/models/') and route.endswith((':generateContent', ':streamGenerateContent')))
        if not valid or path.query not in ('', 'alt=sse'):
            self.send_error(400, 'unbound generation path'); return
        try:
            length = int(self.headers.get('content-length','-1'))
            if not 0 <= length <= 2*1024*1024: raise ValueError('request size')
            self.connection.settimeout(5)
            body = self.rfile.read(length)
            if len(body) != length: raise ValueError('incomplete request')
            payload = json.loads(body)
            bound = payload.get('max_tokens',payload.get('max_output_tokens',payload.get('max_completion_tokens',payload.get('generationConfig',{}).get('maxOutputTokens'))))
            if not isinstance(bound,int) or isinstance(bound,bool) or not 0 < bound <= 4096: raise ValueError('output cap')
            attempt = owner.budget.reserve(step)
        except (ValueError, OSError, json.JSONDecodeError):
            self.send_error(429, 'bounded relay refusal'); return
        number = attempt['number']; start = time.monotonic()
        (owner.output/f'{number:03}-request.json').write_bytes(body)
        owner.record({**attempt, 'phase':'admitted', 'source':'loopback' if owner.origin.scheme == 'http' else 'provider'})
        conn = None; status = None; received = bytearray(); outcome = 'failed'; sent = False
        try:
            kind = http.client.HTTPConnection if owner.origin.scheme == 'http' else http.client.HTTPSConnection
            conn = kind(owner.origin.hostname, owner.origin.port, timeout=owner.timeout)
            conn.request('POST',route+('?' + path.query if path.query else ''), body, {**owner.headers,'Content-Type':'application/json'})
            transport_socket = conn.sock
            response = conn.getresponse(); status = response.status
            # Redirects are recorded as failures, never followed.
            self.send_response(status); self.send_header('Content-Type',response.getheader('Content-Type','application/json')); self.send_header('Connection','close'); self.end_headers(); sent = True
            while True:
                left = owner.timeout-(time.monotonic()-start)
                if left <= 0: raise TimeoutError()
                transport_socket.settimeout(left)
                chunk = response.read1(4096)
                if not chunk: break
                if len(received)+len(chunk) > 8*1024*1024: raise ValueError('response cap')
                received.extend(chunk)
                self.wfile.write(chunk); self.wfile.flush()
            outcome = 'completed' if 200 <= status < 300 else 'http_failure'
        except (OSError, ValueError, http.client.HTTPException) as error:
            outcome = type(error).__name__
            if not sent:
                try: self.send_error(502, 'bounded relay transport failure')
                except OSError: pass
        finally:
            if conn: conn.close()
            self.close_connection = True
            retained=bytes(received)
            for name,value in owner.headers.items():
                if name.lower() in ('authorization','x-api-key','x-goog-api-key'):
                    secret=value.removeprefix('Bearer ').encode()
                    if secret:retained=retained.replace(secret,b'[redacted credential]')
            with (owner.output/f'{number:03}-response.body').open('xb') as original:original.write(retained)
            owner.record({'number':number,'phase':'finished','outcome':outcome,'status':status,'bytes':len(received),'credential_redacted':retained!=bytes(received),'elapsed':time.monotonic()-start})
