#!/usr/bin/env python3
"""Prepare mutations of an already accepted student checkpoint.

The caller must establish genuine public-export/inspection positives first.
This generator repairs the envelope hash; generation alone proves no refusal.
Uses the student's published 44056d1 structural grammar, not inferred Go tags.
"""
import base64
import copy
import json
import sys

from accept_ch10 import Number, canonical, digest, envelope


def wire(value, duplicate=None, path=()):
    """Keep structural integers' original tokens; canonicalizing the envelope
    would turn 10 into 1e1 and accidentally test its integer parser first."""
    if isinstance(value, Number):
        return str(value)
    if value is None:
        return 'null'
    if value is True:
        return 'true'
    if value is False:
        return 'false'
    if isinstance(value, str):
        return json.dumps(value, ensure_ascii=False)
    if isinstance(value, int):
        return str(value)
    if isinstance(value, list):
        return '[' + ','.join(wire(v, duplicate, path + (i,)) for i, v in enumerate(value)) + ']'
    if isinstance(value, dict):
        fields = [(k, v) for k, v in value.items()]
        if duplicate == path:
            fields.insert(0, fields[0])
        return '{' + ','.join(json.dumps(k) + ':' + wire(v, duplicate, path + (k,)) for k, v in fields) + '}'
    raise ValueError('unsupported fixture value')


def at(value, path):
    for key in path:
        value = value[key]
    return value


def cases(raw):
    original = envelope(raw)
    if original['state'] is None:
        raise ValueError('requires a genuine non-null export')
    state = original['state']
    # These names are explicitly documented. This is an adapter prerequisite,
    # not a claim that all other codec semantics have been independently proved.
    if set(state) != {'session', 'context', 'usage', 'skills', 'limits', 'window'}:
        raise ValueError('published state grammar differs; review adapter before mutation')
    paths = [('state',), ('state', 'session'), ('state', 'session', 'high_watermarks'),
             ('state', 'context'), ('state', 'window')]
    if state['usage']:
        paths += [('state', 'usage', 0), ('state', 'usage', 0, 'from'), ('state', 'usage', 0, 'usage')]
    context = state['context']
    if context['Responses']:
        paths += [('state', 'context', 'Responses', 0)]
    if context['Turns']:
        paths += [('state', 'context', 'Turns', next(iter(context['Turns'])))]
    if state['window']['events']:
        paths += [('state', 'window', 'events', 0), ('state', 'window', 'events', 0, 'event')]
    out, positives = [], []

    def positive(name, value, escaped=False):
        value['state_sha256'] = digest(canonical(value['state']).encode('utf-8'))
        data = wire(value)
        if escaped:
            # Original decoded strings, including Raw wrappers, remain identical.
            data = data.replace('😀', '\\ud83d\\ude00').replace('�', '\\ufffd')
        envelope(data.encode('utf-8'))
        positives.append({'name': name, 'bytes': base64.b64encode(data.encode('utf-8')).decode('ascii')})

    positive('valid-pair-replacement-and-literal-backslash', copy.deepcopy(original), True)

    def equivalent_number(value):
        if isinstance(value, dict):
            for key, item in value.items():
                if isinstance(item, Number) and '.' not in item and 'e' not in item.lower():
                    value[key] = Number(str(item) + '.0')
                    return True
                if equivalent_number(item):
                    return True
        if isinstance(value, list):
            return any(equivalent_number(item) for item in value)
        return False

    for path in [('identity',), ('state', 'session', 'identity')]:
        changed = copy.deepcopy(original)
        handlers = at(changed, path)['handlers']
        if not any(equivalent_number(item['schema']) for item in handlers):
            raise ValueError('schema-number positive absent; adapt fixture before claiming canonical identity coverage')
        positive('/'.join(path) + '/equivalent-schema-number', changed)

    def emit(name, value, duplicate=None):
        value['state_sha256'] = digest(canonical(value['state']).encode('utf-8'))
        data = wire(value, duplicate).encode('utf-8')
        if duplicate is None:
            envelope(data)  # Never credit a semantic mutation with a bad hash.
        out.append({'name': name, 'bytes': base64.b64encode(data).decode('ascii'),
                    'expected_code': 'session_corrupt', 'state_hash_repaired': True})

    for path in paths:
        parent = at(original, path)
        if not isinstance(parent, dict) or not parent:
            raise ValueError('structural positive parent absent')
        label = '/'.join(map(str, path))
        key = next(iter(parent))
        changed = copy.deepcopy(original)
        at(changed, path)['__independent_unknown__'] = True
        emit(label + '/unknown-field', changed)
        changed = copy.deepcopy(original)
        del at(changed, path)[key]
        emit(label + '/missing-' + key, changed)
        emit(label + '/duplicate-' + key, copy.deepcopy(original), path)

    def change(name, path, value):
        changed = copy.deepcopy(original)
        at(changed, path[:-1])[path[-1]] = value
        emit(name, changed)

    change('metadata-session-mismatch', ('state', 'session', 'id'), 'f' * 32 if original['session_id'] != 'f' * 32 else 'e' * 32)
    change('context-sequence-mismatch', ('state', 'context', 'LastSeq'), 0)
    change('context-skill-mode-mismatch', ('state', 'context', 'SkillMode'), not context['SkillMode'])
    change('window-event-count-underflow', ('state', 'window', 'event_count'), 0)
    if state['window']['events']:
        change('window-renderable-count-underflow', ('state', 'window', 'renderable_count'), 0)
    if state['usage']:
        value = int(state['usage'][0]['usage']['input']) + 1
        change('usage-vs-accepted-response', ('state', 'usage', 0, 'usage', 'input'), value)
    if context['Responses']:
        change('response-coordinate-zero', ('state', 'context', 'Responses', 0, 'seq'), 0)
    if context['Turns']:
        key = next(iter(context['Turns']))
        change('turn-ordinal-zero', ('state', 'context', 'Turns', key, 'index'), 0)

    # The valid parent contains a genuine U+FFFD. A replacement-decoding bug
    # would leave the same semantic metadata and hash, so it cannot be hidden
    # by an unrelated mismatch. No malformed-value canonical hash is invented.
    raw_parent = wire(original)
    if '�' not in raw_parent:
        raise ValueError('Unicode replacement-character positive absent')
    for escape in ('\\ud800', '\\udc00'):
        invalid = raw_parent.replace('�', escape, 1).encode('utf-8')
        out.append({'name': 'unicode-lone-' + escape, 'bytes': base64.b64encode(invalid).decode('ascii'),
                    'expected_code': 'session_corrupt', 'state_hash_repaired': False,
                    'note': 'Same hash as genuine replacement-character parent; must refuse before replacement decoding'})

    # Mutate all copies of the same replay-bearing usage record consistently.
    # The outer strings stay valid Unicode; their embedded JSON must be checked.
    def raw_surrogate(value):
        count = 0
        if isinstance(value, dict):
            for key, item in value.items():
                if key == 'raw_usage' and isinstance(item, str) and 'raw_marker' in item and '�' in item:
                    value[key] = item.replace('�', '\\ud800')
                    count += 1
                else:
                    count += raw_surrogate(item)
        elif isinstance(value, list):
            count += sum(raw_surrogate(item) for item in value)
        return count
    changed = copy.deepcopy(original)
    if not raw_surrogate(changed['state']):
        raise ValueError('genuine Raw usage positive absent')
    emit('invalid-surrogate-inside-Raw-wrapper', changed)
    return {'positives': positives, 'cases': out, 'scope': __doc__}


if __name__ == '__main__':
    raw = sys.stdin.buffer.read()
    print(json.dumps(cases(raw)))
