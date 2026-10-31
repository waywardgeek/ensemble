#!/usr/bin/env python3
"""Small argument-codec controls over an already accepted public checkpoint.

Uses the documented Raw-string grammar. Runtime must establish both parent
positives; emitted mutations alone prove no acceptance/refusal.
"""
import base64
import copy
import json
import sys

from accept_ch10 import canonical, digest, encoded_json, envelope

DUPLICATE = '{"name": "edit", "name": "edit"}'
ALTERNATE = '{ "name": "edit", "name": "edit" }'


def encode(value, duplicate=None, path=()):
    if isinstance(value, dict):
        fields = list(value.items())
        if duplicate and path == duplicate[0]:
            key = duplicate[1]
            fields.insert(0, (key, value[key]))
        return '{' + ','.join(json.dumps(k) + ':' + encode(v, duplicate, path + (k,))
                              for k, v in fields) + '}'
    if isinstance(value, list):
        return '[' + ','.join(encode(v, duplicate, path + (i,)) for i, v in enumerate(value)) + ']'
    return encoded_json(value)


def cases(raw):
    parent = envelope(raw)
    part = parent['state']['context']['Calls']['argument-duplicate']['Part']
    accepted = part['args']
    if (not isinstance(accepted, str) or part['name'] != 'load_skill'
            or json.loads(accepted, object_pairs_hook=list) != [('name', 'edit'), ('name', 'edit')]
            or part.get('arguments_text', DUPLICATE) != DUPLICATE):
        raise ValueError('published genuine duplicate-argument parent absent')
    # 57d4aac documents prepared Part.args plus a separate original OpenAI
    # arguments_text string. Mutate accepted Raw witnesses, not wire spelling.
    out = []

    def emit(name, value, positive=False, repair=True, duplicate=None):
        if repair:
            value['state_sha256'] = digest(canonical(value['state']).encode())
        data = encode(value, duplicate)
        if repair and not duplicate:
            envelope(data.encode())  # Hash-valid negatives reach semantic checks.
        out.append(dict(name=name, bytes=base64.b64encode(data.encode()).decode(),
                        positive=positive, state_hash_repaired=repair,
                        expected_code=None if positive else 'session_corrupt'))

    emit('valid-export-reencoded', copy.deepcopy(parent), positive=True)
    alternate = copy.deepcopy(parent)
    count = 0

    def replace(value, argument=ALTERNATE, wire_text=False):
        nonlocal count
        if isinstance(value, dict):
            for key, item in value.items():
                if item == accepted or (wire_text and key == 'arguments_text' and item == DUPLICATE):
                    value[key] = argument
                    count += 1
                else:
                    replace(item, argument, wire_text)
        elif isinstance(value, list):
            for item in value:
                replace(item, argument, wire_text)

    replace(alternate['state'])
    if count < 3:
        raise ValueError('complete duplicate correspondence witnesses absent')
    # Same invalid call/outcome, consistently different accepted text. A public
    # snapshot import must accept its repaired hash before stale-hash credit.
    emit('valid-alternate-duplicate-text', copy.deepcopy(alternate), positive=True)
    emit('changed-wrapper-text-stale-hash', alternate, repair=False)
    for name, argument in [('single-member-substitute', '{"name":"edit"}'),
                           ('different-duplicate-spelling', ALTERNATE),
                           ('broken-argument-json', '{'),
                           ('non-object-argument', '[]'),
                           ('invalid-scalar-argument', '{"name":"\\ud800"}'),
                           ('excessive-argument-depth', '{"name":' + '[' * 129 + '0' + ']' * 129 + '}')]:
        value = copy.deepcopy(parent)
        if name in ('single-member-substitute', 'different-duplicate-spelling'):
            value['state']['context']['Calls']['argument-duplicate']['Part']['args'] = argument
        else:
            # Keep every correspondence witness equal so syntax/scalar/depth
            # refusal cannot be credited to an unrelated single-copy mismatch.
            replace(value['state'], argument, wire_text=True)
        emit(name, value)
    value = copy.deepcopy(parent)
    emit('duplicate-outer-as-of', value, duplicate=((), 'as_of'))
    value = copy.deepcopy(parent)
    path = ('state', 'context', 'Calls', 'argument-duplicate', 'Part')
    emit('duplicate-call-args-member', value, duplicate=(path, 'args'))
    value = copy.deepcopy(parent)
    event = value['state']['window']['events'][0]['event']
    if 'seq' not in event:
        raise ValueError('genuine event structural parent absent')
    emit('duplicate-event-sequence', value, duplicate=(('state', 'window', 'events', 0, 'event'), 'seq'))
    return out


if __name__ == '__main__':
    print(json.dumps(cases(sys.stdin.buffer.read())))
