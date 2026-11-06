# Chapter7 browser-copy supplement components

**Development boundary:** this directory contains runnable buffered and
streaming-copy supplements, binding validation and component tests. Delayed-completion ports
and other Chapter7 GUI gates remain pending. This is not complete Chapter7
acceptance and awards no chapter points. The existing Chapter7 80-point wire
grader is unchanged.

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
apply only to particular cases. The streaming case exercises expansion and
idle-state bindings; the delayed-completion case remains pending.

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
setting. `run_buffered.py` generates fresh inputs and expectations rather than
letting a candidate redefine what counts as correct copying.

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
uses a controlled Playwright boundary for both browser components: SIGTERM reaches
browser cleanup, a held close prevents result publication, and no-join variants
are detected. That test is not real-browser or candidate-lifecycle certification.

Actual-page calibration separately exercises the reference and renamed-layout
positive pages, missing-copy controls, and an1800-character-prefix copy variant.
Those parent-only pages and completed renderer code are excluded from student
prerequisites. See the reconstruction matrix for exact receipts and limitations.


## Owned buffered runner

Requires POSIX Python3 (standard library), existing Node, Playwright and the
Chrome channel. Nothing is installed or downloaded. Supply an absolute executable
launcher that uses `exec` to run the already-built GUI with all received arguments.
Use explicit runtime paths inside it; the runner supplies only `/usr/bin:/bin`
as PATH. The GUI must support the chapter's `--port`, `--workspace`, `--vendor`,
`--model`, `--base-url`, `--event-log`, `--gui-dir` and `--no-stream` flags.
Do not use a launcher that backgrounds, daemonizes or connects to a live service.

```sh
PYTHONDONTWRITEBYTECODE=1 /absolute/python3 run_buffered.py \
  --launcher /absolute/candidate-gui-launcher \
  --assets /absolute/candidate-wwwroot \
  --node /absolute/node \
  --playwright /absolute/existing/playwright \
  --bindings /absolute/candidate-bindings.json \
  --work /absolute/new-work-directory
```

The work directory must not exist; its parent must exist. Evidence remains there:
random fixture, browser/host stdout and stderr, canonical log, provider requests
and `result.json`. The runner starts an owned loopback buffered provider and GUI,
then checks actual browser results, exact original human/assistant canonical
payloads, one provider exchange, no tool effects and successful host shutdown.
All named browser checks and explicit terminal fields are required. Setup refuses
relative or missing dependency paths before launching. Bindings remain optional;
the checked-in reference is the default.

Children receive fresh HOME, temporary and CLI directories, an allowlisted
environment and a synthetic key. Host environment credentials/proxy settings are
not forwarded. This is not an OS network sandbox: only run an authorized candidate
that honors its explicit loopback endpoint. Browser off-origin requests are
blocked and fail the check. No native clipboard read or audio acceptance is made.

Startup has a15-second deadline, the browser45 seconds; cleanup allows5 seconds
for the browser and10 for the host before killing their owned POSIX process groups.
Provider sockets have5-second inactivity timeouts. Direct processes and the server
thread are joined. This does not promise containment of deliberately detached
processes or indefinitely trickling hostile clients. Forced cleanup is reported
and cannot yield a passing grade. Signal interruption exits2 after cleanup;
argument/pre-launch errors also exit2. Failures detected in a launched candidate
or browser exit1; buffered success exits0. Cleanup errors are retained separately
from the primary failure. No shared CLI scope or
full-chapter grade is implied.

Runner tests exercise independent fixture values, explicit dependencies, every
required receipt field/check, cleared child environment and interruption while
an owned startup fixture ignores SIGTERM. That fixture must be killed and reaped
before the runner publishes its result:

```sh
PYTHONDONTWRITEBYTECODE=1 /absolute/python3 runner_test.py -v
```


## Streaming retention case

Use the same runner and explicit dependencies with `--case streaming` and a fresh
work directory. The filename `run_buffered.py` retains the original entry point;
omitting `--case` still selects buffered behavior. The streaming mode omits
`--no-stream` when launching the GUI and requires an actual streaming provider
request. It generates fresh markers and four appended text phases. Three ordered
grader-owned HTTP gates hold fence closure, later prose and final prose until
the real browser has checked the preceding state. These gates are fixture control,
not candidate API requirements or injected application display frames.

Nine named checks verify no premature copy control, exact closed-fence copying,
the same focused button, expansion and feedback, and the same readonly manual
field with its full selection. The last selection check waits for the exact typed
final, completed turn and visible idle state. A final delta alone is insufficient.
The runner independently checks the unchanged canonical payload and one provider
exchange. Gate releases must be exactly `[1, 2, 3]`; teardown unblocks the provider
without adding successful browser-release receipts. Provider handlers join during
server close, including after failed browser assertions.

The browser component is `streaming.cjs`, with arguments: owned GUI URL, fixture
JSON, owned provider URL, trusted Playwright module and optional bindings path.
Its receipt contract is `chapter7-stream-browser-component-v1`; the runner emits
`chapter7-streaming-runner-v1`. Expected payloads and gates remain grader-owned.
Parent-only positive/negative calibration pages are not student dependencies.
Delayed success/denial replacement checks still await their public port; streaming
retention alone does not certify that distinct ownership boundary.
