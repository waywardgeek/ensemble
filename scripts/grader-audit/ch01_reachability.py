#!/usr/bin/env python3
"""Round 2: reachability probes.

Round 1 asked "is each promise graded?". This asks the two vacuity questions:
can each rule inside a check actually FIRE, and does any assertion pass
vacuously? A rule no mutant can trip is decoration.
"""
import json
import sys

from ch01_promises import build_mutant, score, sub1, chain

MSGS = r'Messages:(\s+)conv,'

MUTANTS = [
    ("sys-dropped",
     "L107 'system: a string outside the messages array'",
     sub1(r'[ \t]*System:\s+systemPrompt,\n', '')),

    ("stream-true",
     "L336 wire contract: no streaming (stream: true is rejected)",
     chain(sub1(r'(type request struct \{\n)', r'\1\tStream    bool            `json:"stream"`\n'),
           sub1(r'(MaxTokens: 1024,\n)', r'\1\t\tStream:    true,\n'))),

    ("model-empty",
     "model must be non-empty",
     sub1(r'Model:(\s+)c\.Model,', r'Model:\g<1>"",')),

    ("content-empty",
     "no message may have empty content",
     sub1(MSGS, r'Messages:\g<1>blankAll(conv),')),

    ("first-not-user",
     "the first message must be a user turn",
     sub1(MSGS, r'Messages:\g<1>append(Conversation{{Role: "assistant", Content: "preamble"}}, conv...),')),

    ("last-not-user",
     "the last message must be a user turn",
     sub1(MSGS, r'Messages:\g<1>append(append(Conversation{}, conv...), Message{Role: "assistant", Content: "dangling"}),')),

    ("history-window-2",
     "the whole history is resent, not a sliding window",
     sub1(MSGS, r'Messages:\g<1>lastN(conv, 2),')),

    ("reply-mangled",
     "answers come from the API response",
     sub1(r'(return text\.String\(\), nil)', r'return strings.ToUpper(text.String()), nil')),

    ("double-call",
     "exactly one API call per round",
     sub1(r'([ \t]*)([A-Za-z_]+, err := c\.Send\(\*conv\)\n)',
          r'\1if _, e2 := c.Send(*conv); e2 != nil {\n\1\treturn "", e2\n\1}\n\1\2')),
]

HELPERS = '''

// --- audit helpers (injected by the ch1 grader audit) ---

func blankAll(conv Conversation) Conversation {
	out := append(Conversation{}, conv...)
	for i := range out {
		out[i].Content = ""
	}
	return out
}

func lastN(conv Conversation, n int) Conversation {
	if len(conv) <= n {
		return conv
	}
	return append(Conversation{}, conv[len(conv)-n:]...)
}
'''


def with_helpers(f):
    def g(src):
        return f(src) + HELPERS
    return g


def main():
    rows = []
    for name, promise, mutate in MUTANTS:
        m = with_helpers(mutate) if name in ("content-empty", "history-window-2") else mutate
        try:
            d = build_mutant(name, m)
        except AssertionError as e:
            print(f"!! {name}: {e}", file=sys.stderr)
            rows.append({"name": name, "promise": promise, "score": "ANCHOR-FAIL",
                         "failing": [], "err": str(e)})
            continue
        s, failing, err = score(d)
        rows.append({"name": name, "promise": promise, "score": s,
                     "failing": failing, "err": err})
        print(f"{name:20} score={s!s:>4}  failing={','.join(failing) or '-'}  {err}", flush=True)
    with open("/tmp/ch01audit/results2.json", "w") as f:
        json.dump(rows, f, indent=2)


if __name__ == "__main__":
    main()
