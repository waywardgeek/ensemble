# Queued code fixes

Small, well-specified defects found by **using** the solutions rather than
grading them. All are sub-agent tasks. None are author work.

Blocked until `ch2-openai-null` lands, because it holds
`solutions/ch0{2,3}/main.go`, `solutions/ch0{2,3}/openai.go` and
`internal/fakevendor/fake.go`.

Build/test in this repo: `go build ./...`, `go test ./...`,
`go run ./cmd/grade -ch 2 solutions/ch02` (directory is positional).

---

## 1. Default log name should derive from the executable

Both solutions hardcode `envOr("CH02_LOG", "ch02.log")`, so a binary built from
`solutions/ch03` writes `ch02.log`. Use `filepath.Base(os.Args[0]) + ".log"` as
the default instead.

Free: all four grader harnesses set `CH02_LOG` explicitly
(`internal/grade/ch02_harness.go` ×3, `ch03_harness.go` ×1), so the default is
reached only by a human running the binary by hand.

Do **not** rename the `CH02_LOG` variable itself — that breaks students who
passed ch2 reading it. See `solutions-repo-design.md`.

## 2. `--help` prints nothing useful

`./ch03 --help` returns `unknown command "--help"` and no usage text. It should
print the commands table. Same for `-h` and no-argument invocation where
appropriate.

## 3. The no-subcommand error does not name the cause

Running the binary with no subcommand makes it read a JSON-lines log from stdin.
A human who types `Hi.` gets:

    {"error":"bad input: invalid character 'H' looking for beginning of value"}

Technically accurate, operationally useless. It should say that this mode expects
a JSON log on stdin and suggest `chat` for interactive use. This is the book's
own loud-failure argument applied to the book's own code.

Bill hit 2 and 3 within sixty seconds of first real use.

## 4. `live.sh` lets the demo write into the course repo

`scripts/live.sh` does `cd "$(dirname "$0")/.."`, so the agent runs with the
repo root as its working directory. Chapter 3's agent has `write_file`. The
scripted ch3 demo therefore **created `PROJECT_CODENAME.txt` in the repo root**
containing the codename it invented. Confirmed: the file existed, containing
`Phoenix`, and was deleted by hand.

The demo should run in a scratch directory (temp dir, or a gitignored
`scratch/`), not in the repository that holds the book.

Note for the author, not the coder: this is a small live demonstration of the
read-only versus mutating boundary that the security chapter argues for. A
read-only agent could not have done it. Worth citing there.

## 5. `ch02.log` is tracked and churns

It is run output under version control, so it appears as a modification after
any manual run and adds noise to every `git status`. Gitignore it (and remove it
from the index), or have runs write to a scratch path by default.

Coordinate with item 1: if the default log name becomes the binary name, the
ignore rule should cover `*.log` rather than one filename.
