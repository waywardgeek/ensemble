"""Explicit API-key adapter. Import/local tests never read settings or use network.

Only the released live entrypoints call load_key. No SDK, ambient auth, redirect,
retry, model fallback, or configurable external origin. See provider-support.md.
"""
import http.client
import json
from pathlib import Path
import re
import socket
import threading
import time
from urllib.parse import quote, urlsplit

FIELDS = {'anthropic': 'directClaudeAPIKey', 'openai': 'directOpenAIAPIKey', 'gemini': 'directGeminiAPIKey'}
ORIGINS = {'anthropic': 'https://api.anthropic.com', 'openai': 'https://api.openai.com', 'gemini': 'https://generativelanguage.googleapis.com'}
LIST_PATHS = {'anthropic': '/v1/models?limit=1000', 'openai': '/v1/models', 'gemini': '/v1beta/models?pageSize=1000'}
GEMINI_TARGET = 'models/gemini-3.8-flash'


def connection(kind, host, port, timeout):
    """Bound DNS wait and connect without leaving a late request worker behind.

    A timed-out daemon may finish DNS resolution, but it owns no socket or HTTP
    operation. Select one address and one connection attempt; no address retry.
    HTTPSConnection still verifies the fixed original hostname with normal TLS.
    """
    conn = kind(host, port, timeout=timeout)
    def connect(address, limit, source_address=None):
        start = time.monotonic(); ready = threading.Event(); result = []
        def resolve():
            try: result.append(socket.getaddrinfo(address[0], address[1], type=socket.SOCK_STREAM))
            except OSError: result.append(None)
            finally: ready.set()
        worker = threading.Thread(target=resolve, daemon=True); worker.start()
        if not ready.wait(limit): raise TimeoutError('bounded resolution timeout')
        if not result[0]: raise OSError('bounded resolution failure')
        left = limit-(time.monotonic()-start)
        if left <= 0: raise TimeoutError('bounded resolution timeout')
        family, socktype, proto, _, endpoint = result[0][0]
        sock = socket.socket(family,socktype,proto)
        try:
            sock.settimeout(left)
            if source_address: sock.bind(source_address)
            sock.connect(endpoint)
            # TLS handshake inherits the remaining total budget, not a fresh
            # timeout after a slow TCP connect. HTTP send has the same bound.
            left = limit-(time.monotonic()-start)
            if left <= 0: raise TimeoutError('bounded connection timeout')
            sock.settimeout(left)
            return sock
        except OSError:
            sock.close(); raise
    conn._create_connection = connect
    return conn


def arm_deadline(connection, seconds):
    """Interrupt header/body drips too; socket idle timeouts alone are insufficient."""
    held = [None]
    def expire():
        sock = held[0] or connection.sock
        if sock:
            try: sock.shutdown(socket.SHUT_RDWR)
            except OSError: pass
            sock.close()
    timer = threading.Timer(seconds, expire); timer.daemon = True; timer.start()
    return timer, held


def load_key(vendor, *, settings=None):
    """Select only this vendor's top-level field; no env/OAuth fallback or dump."""
    if vendor not in FIELDS: raise ValueError('unsupported vendor')
    try:
        path = Path(settings) if settings is not None else Path.home()/'.cr/settings.json'
        with path.open('rb') as f:
            raw = f.read(1024*1024+1)
        if len(raw) > 1024*1024: raise ValueError()
        data = json.loads(raw)
        key = data[FIELDS[vendor]]
        if not isinstance(key, str) or not re.fullmatch(r'[A-Za-z0-9._-]{8,512}', key): raise ValueError()
        return key
    except (OSError, ValueError, KeyError, TypeError):
        raise ValueError('required API-key field unavailable or invalid') from None


def headers(vendor, key):
    if vendor == 'anthropic': return {'x-api-key': key, 'anthropic-version': '2023-06-01'}
    if vendor == 'openai': return {'Authorization': 'Bearer '+key}
    if vendor == 'gemini': return {'x-goog-api-key': key}
    raise ValueError('unsupported vendor')


class Redactor:
    """Streaming exact-key/URL/JSON-escape redaction, before either output sink.

    Hold a possible token prefix across chunk boundaries. Exceptions and arbitrary
    headers are never emitted at all. Normal response bytes stay unchanged.
    """
    def __init__(self, values):
        variants = set()
        for value in values:
            if not value: continue
            variants.update((value.encode(), quote(value, safe='').encode()))
            variants.add(''.join('%%%02X' % b for b in value.encode()).encode())
            variants.add(''.join('%%%02x' % b for b in value.encode()).encode())
            variants.add(''.join('\\u%04x' % ord(c) for c in value).encode())
        self.tokens = sorted(variants, key=len, reverse=True)
        self.pending = b''; self.changed = False

    def feed(self, data, final=False):
        self.pending += data; out = bytearray()
        while self.pending:
            found = [(self.pending.find(t), t) for t in self.tokens if t in self.pending]
            if found:
                offset, token = min(found, key=lambda pair: (pair[0], -len(pair[1])))
                out.extend(self.pending[:offset]); out.extend(b'[redacted credential]')
                self.pending = self.pending[offset+len(token):]; self.changed = True
                continue
            hold = 0
            if not final:
                for token in self.tokens:
                    for size in range(min(len(token)-1, len(self.pending)), hold, -1):
                        if self.pending.endswith(token[:size]): hold = size; break
            end = len(self.pending)-hold
            out.extend(self.pending[:end]); self.pending = self.pending[end:]
            break
        return bytes(out)



def selected_model(vendor, model):
    pattern = r'models/[A-Za-z0-9._-]+' if vendor == 'gemini' else r'[A-Za-z0-9._-]+'
    if not isinstance(model, str) or not re.fullmatch(pattern, model): raise ValueError('invalid selected model')
    if vendor == 'gemini' and model != GEMINI_TARGET: raise ValueError('required Gemini target unavailable')
    return model


class Provider:
    def __init__(self, vendor, key, model=None, *, fixture_origin=None):
        if vendor not in ORIGINS: raise ValueError('unsupported vendor')
        if not isinstance(key, str) or not re.fullmatch(r'[A-Za-z0-9._-]{8,512}', key): raise ValueError('invalid API key')
        self.vendor, self.key = vendor, key
        self.model = selected_model(vendor, model) if model is not None else None
        self.origin = ORIGINS[vendor]
        self.local = fixture_origin is not None
        if self.local:
            u = urlsplit(fixture_origin)
            if u.scheme != 'http' or u.hostname != '127.0.0.1' or u.username or u.password or u.path not in ('', '/') or u.query or u.fragment:
                raise ValueError('adapter fixture requires literal loopback')
            self.origin = fixture_origin
        self.headers = headers(vendor, key)

    def generation(self, route, query, payload):
        if self.model is None: raise ValueError('model selection required')
        if self.vendor == 'anthropic': valid = route == '/v1/messages' and not query and payload.get('model') == self.model
        elif self.vendor == 'openai': valid = route == '/v1/chat/completions' and not query and payload.get('model') == self.model
        else:
            stem = '/v1beta/'+self.model
            valid = (route == stem+':generateContent' and not query) or (route == stem+':streamGenerateContent' and query == 'alt=sse')
        if not valid: raise ValueError('unbound provider generation route/model')

    def discover(self, budget, output, binding_sha256, timeout=120):
        """One GET, even for failure/pagination. Only after identity preflight."""
        output = Path(output)
        if output.exists() or output.with_suffix('.body').exists(): raise ValueError('discovery output already exists')
        item = budget.reserve('models', 'discovery')
        # Exclusive receipt before transport; a crash cannot invite an unseen retry.
        receipt = dict(vendor=self.vendor, origin=self.origin, path=LIST_PATHS[self.vendor],
                       mode='local' if self.local else 'live', binding_sha256=binding_sha256,
                       attempt=item, budget_path=str(budget.path.resolve()), status=None, outcome='admitted', models=[], has_more=False)
        with output.open('x') as f: json.dump(receipt, f, indent=2)
        conn = None; timer = None; start = time.monotonic(); raw = bytearray(); redactor = Redactor([self.key])
        try:
            u = urlsplit(self.origin); kind = http.client.HTTPConnection if self.local else http.client.HTTPSConnection
            conn = connection(kind, u.hostname, u.port, min(timeout,120))
            timer, held = arm_deadline(conn, min(timeout,120))
            conn.request('GET', LIST_PATHS[self.vendor], headers=self.headers)
            sock = conn.sock; held[0] = sock; response = conn.getresponse(); receipt['status'] = response.status
            while True:
                left = min(timeout,120)-(time.monotonic()-start)
                if left <= 0: raise TimeoutError()
                if response.isclosed(): break
                sock.settimeout(left); chunk = response.read1(4096)
                if not chunk: break
                if len(raw)+len(chunk) > 2*1024*1024: raise ValueError('discovery response bound')
                raw.extend(chunk)
            safe = redactor.feed(bytes(raw), final=True)
            if redactor.changed: raise ValueError('credential echo')
            if response.status != 200:
                receipt['outcome'] = 'http_failure'
            else:
                page = json.loads(safe)
                entries = page.get('models' if self.vendor == 'gemini' else 'data')
                if not isinstance(entries, list): raise ValueError('discovery shape')
                names = []
                for entry in entries:
                    if self.vendor == 'gemini' and 'generateContent' not in entry.get('supportedGenerationMethods', []): continue
                    name = entry.get('name' if self.vendor == 'gemini' else 'id')
                    # Discovery may contain other Gemini models: selection enforces target later.
                    pattern = r'models/[A-Za-z0-9._-]+' if self.vendor == 'gemini' else r'[A-Za-z0-9._-]+'
                    if not isinstance(name,str) or not re.fullmatch(pattern,name): raise ValueError('discovery model shape')
                    names.append(name)
                receipt.update(outcome='completed', models=names, has_more=bool(page.get('has_more') or page.get('nextPageToken')))
        except (OSError, ValueError, TypeError, AttributeError, http.client.HTTPException):
            receipt['outcome'] = 'bounded_transport_or_response_failure'
        finally:
            if timer: timer.cancel(); timer.join()
            if conn: conn.close()
            safe = Redactor([self.key]).feed(bytes(raw), final=True)
            with output.with_suffix('.body').open('xb') as f: f.write(safe)
            receipt['elapsed'] = time.monotonic()-start
            output.write_text(json.dumps(receipt,indent=2)+'\n')
        return receipt


def require_discovery(receipt, vendor, model, binding_sha256, budget_path=None):
    selected_model(vendor, model)
    if (receipt.get('mode') != 'live' or receipt.get('vendor') != vendor or
        receipt.get('origin') != ORIGINS[vendor] or receipt.get('path') != LIST_PATHS[vendor] or
        receipt.get('binding_sha256') != binding_sha256 or receipt.get('outcome') != 'completed' or
        receipt.get('status') != 200 or model not in receipt.get('models', [])):
        raise ValueError('model absent from bound live discovery')
    if budget_path is not None:
        rows = [json.loads(line) for line in Path(budget_path).read_text().splitlines()]
        discovery = [r for r in rows if r.get('kind') == 'discovery']
        if receipt.get('budget_path') != str(Path(budget_path).resolve()) or discovery != [receipt.get('attempt')]:
            raise ValueError('discovery budget association mismatch')
