# Next trial: freeze the prepared configuration before launch

Status: preserved preparation recipe, since executed for [trial03](../evidence/trial-03.md).
The preparation-only statements below describe that earlier boundary, not current
installation status. The host was subsequently installed and checked, and the
primary installed with approval before trial03.
Bill additionally requires the incoming-message idle-wake implementation and its
book updates before launch; this is now authorized work, not merely a queue item.
Use a fresh child ID. Trial02 is frozen and closed; never reuse it for another
implementation attempt. This tests the combined prepared configuration, not a
single-variable or causal benchmark. Bill authorized preparation; host installation
and restart remain his boundary before using source changes in a live trial.

## Before admission

1. Verify the intended host build serves the new read diagnostics using disposable
   in-root files: source clipping with a safe continuation, explicit EOF, missing
   path says not found. Do not infer installation from a successful source build.
2. Install the reviewed `skills/flint-student/SKILL.md` into the configured catalogue
   only at the approved installation boundary, preserving the previous candidate.
   Compare SHA-256 bytes with this book. Do not change any existing Agent's active
   skill. The tool surface remains unchanged; no delegation or web capability.
3. Use exact offering `openrouter/chat/flint-v1_0-free`, `confined:true`, primary
   `flint-student`, working context200000. Import only the selected Go toolchain
   root and .NET SDK root read-only. No paid fallback or real credentials.
4. Prepare a fresh project and saved memory settings before task admission using
   the recipe below. Memory budgets are UTF-8 bytes, not token capacities:
   Conversation100000, Session40000, archive8/archive64/MEMORY10000 each, SOUL5000;
   all enabled. Do not retune these mid-trial without recording an intervention.
5. Record the new ID, exact route, effective root/imports, runtime commit/build,
   working limit/capacity, memory policy, public input/primary/environment hashes,
   task text and the original spawn Job. Verify the saved launch policy after
   admission. Record diagnostic smoke evidence separately from student behavior.

## Prepare public inputs

Run from the Ensemble repository root with Python3 available. Set absolute paths
for the NEW child home, .NET SDK directory and Go executable. The child home is
its assigned parent-owned directory, not a separately chosen writable exception.
No student or provider is launched by these commands. They refuse existing project
or settings files rather than overwrite a previous trial. Choose a fresh ID if a
partial setup exists; preserve it for diagnosis rather than delete unrelated files.

```sh
export CHILD_HOME=/absolute/parent/children/NEW-ID
export SDK_ROOT=/absolute/dotnet-sdk
export GO_EXE=/absolute/go/bin/go
python3 - <<'PY'
import hashlib, json, os
from pathlib import Path
repo = Path.cwd()
home = Path(os.environ['CHILD_HOME'])
sdk = Path(os.environ['SDK_ROOT'])
go = Path(os.environ['GO_EXE'])
for path in (home, sdk, go):
    assert path.is_absolute() and "'" not in str(path) and "\n" not in str(path)
assert (sdk / 'dotnet').is_file() and go.is_file()
project = home / 'project'
assert not project.exists() and not (home / 'settings.json').exists()
project.mkdir(parents=True)
for name in ('chapter-00.md', 'chapter-01.md'):
    (project / name).write_bytes((repo / 'book' / name).read_bytes())
recipe = (repo / 'flint-book/setup/env.sh').read_text()
recipe = recipe.replace('@PROJECT_DIR@', str(project)).replace('@GO_EXE@', str(go)).replace('@DOTNET_ROOT@', str(sdk))
(project / 'env.sh').write_text(recipe)
(project / 'NuGet.Config').write_text('<configuration><packageSources><clear /></packageSources></configuration>\n')
budgets = dict(conversation=100000, session=40000, archive8=10000, archive64=10000, memory=10000, soul=5000)
with (home / 'settings.json').open('x') as stream:
    json.dump(dict(version=1, memory={name: dict(budget=size, disabled=False) for name, size in budgets.items()}), stream)
print(json.dumps({p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in project.iterdir()}, indent=2))
PY
```

Use the existing chapter task and acceptance requirements, supplying the absolute
project path. The worked Graph/Node example is active instruction in the primary,
not a second application subsystem to add. No reference implementation or private
grader source enters the workspace. Preserve the public chapters' requirements;
this preparation changes teaching/setup, not the passing standard.

## Observe, grade, close

Keep an observation on the original child (or join) while waiting for its question.
The candidate actor now records a child notice once and queues an idle model turn;
a queued model turn absorbs further notices, and active work receives a hint.
Pause defers inference until resume; shutdown cancels queued work. Root questions
for a human do not self-wake. Focused fake-provider checks demonstrate idle, busy,
paused/coalesced and shutdown behavior without timers or elevated message authority.
Before this trial, verify the installed host contains that change; source tests are
not installation or live acceptance. Keep the original observation handle and
respect a real user pause.

When the student freezes and requests grading, do not answer with an acknowledgement
and create a second result-delivery protocol. Keep the question open, hash the
source/output files, and run the unchanged trusted confined Chapter1 grader.
Allow its two62-second probes to finish. Answer the pending question with the
actual score, passed flag and exit status, plus concrete failures if any. The
student's question Job completes with that reply; it is not the parent grader Job.
Preserve both handles. No source repairs, unrestricted execution or real keys.

Require separate architecture review of direct constructor/add/remove/destroy
contracts and Common placement, not a score alone. Interview once, collect typed
submission and join shutdown. Compare observed friction with trial02; disclose
assistance, pauses and provider failures. Do not rerun a passing candidate or
advance to Chapter2 without Bill's decision. No live-provider exercise is authorized.


## Preparation evidence

Before scheduler work, the candidate runtime passed400 native tests and the full
GUI suite; the GUI host built with zero warnings/errors. New continuation and
missing-path assertions failed against the old implementation and passed after
correction. Exact-schema workflow checks required synchronizing the active package's
read-file description. The GUI suite exposed and then corrected its stale expectation
that public child deletion removes the launch record; it now requires Retired and
still verifies no automatic restart. These test repairs did not weaken runtime
safety or change historical graders. The example/recipe check and real skill loader
passed; deliberately broken registration and detachment examples failed as intended.
No runtime installation, host restart, primary catalogue overwrite or student launch
was performed during preparation. Scheduler source work follows this preparation;
installed-host verification remains a pre-launch boundary.

Idle-message source verification:401 native tests passed after the workflow fake
was updated to serve the newly awakened parent's request separately from its two
child assignments. The9process/workflow checks passed, as did the full GUI suite
and clean GUI host build. A final test-only late-question refusal assertion passed
with all5focused message-wake checks after the full run. The initial idle-wake test
failed against the old actor because no provider request arrived. No live provider,
installation, restart, or student launch was part of this verification.
