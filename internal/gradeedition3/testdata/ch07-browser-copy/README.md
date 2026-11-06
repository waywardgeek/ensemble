# Chapter7 browser-copy supplement components

**Development boundary:** this directory currently contains binding validation
and the buffered browser component. It is not yet a standalone supplemental
grader. The owned-provider/GUI process runner, portable fixture generation, and
ported streaming/delayed-completion cases are still pending. No chapter points or
student acceptance follow from these components alone. The existing Chapter7
80-point wire grader is unchanged.

The browser uses public DOM controls and observes the actual WebSocket traffic.
It substitutes only browser clipboard/speech boundaries. It neither reads private
application JavaScript state nor injects display frames. A caller must supply an
owned isolated GUI and fake provider; do not point it at a working application.
Node and a trusted existing Playwright installation are explicit dependencies.
The component launches the existing Chrome channel and never installs a browser.

## Data-only DOM bindings

`bindings.json` is the reference binding, not required application markup. A GUI
may use different IDs, containers and feedback wording and supply its own JSON.
Bindings must describe the real human-facing controls and truthful states; they
are not permission to expose fake controls or weaken behavior. Independent source
review and human usability remain separate acceptance gates.

All fields are required; unknown fields, code callbacks and skip flags are refused.
The file must be ordinary UTF-8 JSON at most16 KiB. It contains version1 plus these
strings (each nonblank and at most512 UTF-16 code units):

| Field | Scope and meaning |
| --- | --- |
| `connected`, `idle` | Page selectors matching the visible connected/idle states, not an always-present placeholder |
| `composer`, `send` | Page selectors for the editable prompt and its submit control |
| `cards` | Page selector yielding individual displayed artifacts |
| `snippet` | Relative to a card: individual code-snippet containers |
| `code` | Relative to a snippet: complete code content, not a clipped preview |
| `copy` | Relative to a card/snippet: the actual snippet-copy buttons |
| `feedback` | Relative to a snippet: its readable copy-status element |
| `manual` | Relative to a snippet: the labelled read-only manual-copy textarea |
| `expand`, `expanded` | Relative selectors for expansion activation and the expanded state |
| `copied`, `copying` | Distinct, single-line visible success and pending feedback text |

Selectors use Playwright's public locator syntax. Selector syntax and meaning are
not certified by JSON validation alone. A binding cannot alter expected payloads,
control provider gates, access application variables or omit a test. Some fields
are reserved for the streaming/delayed components and are not exercised by the
buffered case yet.

The loader reads through an owned descriptor with a bounded buffer, checks a
regular file and strict UTF-8, and returns a frozen copy. It follows explicitly
selected symlinks to regular files; this is configuration reading, not Sandbox
confinement or an immutable-file guarantee.

## Current component interface

`buffered.cjs` accepts an owned loopback GUI URL, grader-owned fixture JSON,
absolute trusted Playwright module path, and optional binding file path. It owns
and closes its browser. The module-path requirement is a trusted-caller
precondition, not an independently enforced absolute-path check. Its caller still
owns the GUI/provider and must verify
canonical History, provider request counts and joined host termination. The
parent calibration uses real HTTP/GUI connections for those separate checks.

The fixture specifies `prompt`, `marker`, `human`, original `markdown`, and four
independently specified logical `payloads`. It is test data, not a candidate
setting. The full runner will generate it rather than let a candidate redefine
what counts as correct copying.

The JSON receipt contract is `chapter7-copy-browser-component-v1`. Exit0 means
this buffered component passed;1 means its configured behavior checks failed;
2 denotes setup, cleanup or interruption failure. No shared CLI scope is wired
for this unfinished supplement. Do not treat component exit0 as a full grade.

## Component tests

With an existing Node executable, no package install is needed for these tests:

```sh
/absolute/node --test bindings.test.cjs lifetime.test.cjs
```

Binding tests cover required fields, rejected code/skip additions, value limits,
strict UTF-8, regular-file/size refusal and immutable values. The lifetime test
uses a controlled Playwright boundary: SIGTERM reaches browser cleanup, a held
close prevents result publication, and a no-join variant is detected. That test
is not real-browser or candidate-lifecycle certification.

Actual-page calibration separately exercises the reference and renamed-layout
positive pages, missing-copy controls, and an1800-character-prefix copy variant.
Those parent-only pages and completed renderer code are excluded from student
prerequisites. See the reconstruction matrix for exact receipts and limitations.
