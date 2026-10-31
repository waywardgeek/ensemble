# The Art of Building AI Coding Agents

**C# Edition 3:** the locally completed book and implementation are in the
[Edition 3 guide](book/edition-3/README.md). The instructions below continue to
describe the existing Go application; they do not build or install C# Edition 3.

*Second edition. An agentic codebook: build the agent, test the teaching, improve the book.*

**[Read the book](book/the-art-of-building-ai-coding-agents.md)** ·
[Chapter guide](book/chapters.md) · [Solutions](solutions/README.md) ·
[First edition archive](edition-1/ARCHIVE.md)

Ensemble is the coding agent built by this book's exercises. It has tools and
supervised jobs, streaming CLI and browser interfaces, persistent conversations,
skills, memory, a fleet of agents and an authenticated gateway. The code is here,
along with the instructions, graders and reviews that shaped it.

The second edition was built with Codex under Bill Cox's direction, using a
student implementation followed by comparative review and revisions to both
teaching and graders. It catches up through the first edition's final exercise,
Gateway, now Chapter 26. Original Chapters 5, 22 and 24 were absorbed into related
chapters while preserving their capabilities.

**Crossover has begun.** Ensemble has performed real coding tasks with GPT-6
Astra at high reasoning. Making it the everyday tool for the next edition is
ongoing work. Chapter 25 is deliberately a placeholder for that experience;
this publication does not claim Ensemble has already produced Edition 3.
Chapter 24 is an optional, ungraded Codex-endpoint workaround that readers may
skip when their public OAuth route caches satisfactorily.

## Repository layout

| Path | Purpose |
| --- | --- |
| [`book/`](book/) | Current second-edition manuscript and publication files |
| [`agent/`](agent/) | Current application development tree, with its own Go module |
| [`agent/gui/`](agent/gui/) | Optional browser application, in a separate Go module |
| [`solutions/chNN/`](solutions/README.md) | Accepted cumulative chapter snapshots; use matching book numbers |
| [`solutions/evidence/`](solutions/evidence/) | Retained runs and checks, using original exercise numbers |
| [`cmd/grade-edition2/`](cmd/grade-edition2/) | Current chapter grader command |
| [`docs/edition-2-notes/`](docs/edition-2-notes/) | Student, author and independent review records |
| [`edition-1/`](edition-1/) | Historical first-edition book, application, solutions and original graders |

Develop in `agent/`; keep accepted snapshots intact. Chapter 25 has no accepted
solution directory yet. The retired standalone refactor snapshot is the only
chapter under `solutions/history/`. Historical manifests and receipts retain
the paths and source revisions at which they were measured; the
[relocation record](docs/edition-2-publication.md) explains the move.

## Build and run

### Prerequisites

- Go **1.25 or newer** (see the runtime modules' `go.mod` files).
- GNU Make and a Unix-like shell (macOS or Linux).
- Network access on the first build to download Go dependencies.
- For chat, a supported provider account, credentials and a model available to
  that account.

### Build

From `projects/ensemble` (the directory containing this README), run:

```sh
make
```

This creates **`./ensemble`**, the browser application with terminal input.
The browser assets are embedded: no Node, npm or separate frontend build is
needed. Run `make` again after updating the source; Go caches unchanged work.
`make build` is equivalent. The executable is built for your current platform.

### Configure and run

For API-key use, set `LLM_VENDOR` (`anthropic`, `openai` or `gemini`),
`LLM_MODEL` and `LLM_API_KEY` in your environment, then run:

```sh
./ensemble --workspace "$PWD" --skills-dir "$PWD/agent/skills"
```

Choose a model supported by the implementation and available to your account.
Supply the API key securely rather than putting it in a command or shell history.
Open **http://localhost:8088** (or the address printed at startup). You can chat
in the browser or the terminal; Ctrl-C stops the application. `./ensemble --help`
lists launch options, including `--port` for a different browser port.
Conversation state and logs are written in the launch directory by default;
`--workspace` selects where the agent's tools work.

For persistent configuration, copy
[`agent/gui/settings.example.json`](agent/gui/settings.example.json) to a private
session directory as `settings.json`. Edit the provider/model, `workspace` and
`skills_dir` paths, and set `provider.credential_file` to a private plain-text API
key file (restrict its permissions with `chmod 600`). Do not put secrets in the
settings JSON. Relative paths resolve against the settings file's directory;
the example paths must be adjusted for your chosen session directory. Launch with:

```sh
./ensemble --config /absolute/path/to/session/settings.json
```

In this mode the settings file, not provider environment variables, controls the
application. See the [GUI README](agent/gui/README.md) for the settings schema,
authentication choices and browser controls.

### Optional terminal-only client and cleanup

```sh
make cli
./bin/ensemble chat --skills-dir "$PWD/agent/skills"
```

This separate client uses the same provider environment variables. Run it from
its intended workspace, using absolute binary and skills paths if necessary.
The application [README](agent/README.md) documents its authentication, streaming
and persistence options.

`make clean` removes `ensemble` and `bin/ensemble`, but leaves settings,
credentials, conversations and logs alone. Existing grading and book targets
remain available below.

For the optional Chapter 24 route, an existing Codex sign-in supplies the
read-only credential file. This example keeps conversation state outside the
checkout while tools work in the repository:

```sh
ensemble_repo="$PWD"
mkdir -p "$HOME/ensemble-session"
(
  cd "$HOME/ensemble-session"
  unset LLM_BASE_URL OPENAI_BASE_URL LLM_RESOLVED_MODEL
  LLM_VENDOR=openai LLM_MODEL=gpt-6-astra \
  OPENAI_REASONING_EFFORT=high OPENAI_BILLING_MODE=codex \
    "$ensemble_repo/ensemble" \
    --workspace "$ensemble_repo" \
    --skills-dir "$ensemble_repo/agent/skills" --save save.json
)
```

Use the browser address printed by the application. The route reads Codex's
existing credentials without copying or renewing them. See
[Chapter 24](book/chapter-24.md) for the workaround's evidence and limitations.
A successful answer alone does not prove prompt-cache reuse.

## Grade an exercise

From the repository root:

```sh
make grade                              # current Chapter 1 snapshot
make grade-dir CH=26 DIR=./solutions/ch26
go run ./cmd/grade-edition2 -ch 8         # matching snapshot by default
go run ./cmd/grade-edition2 -ch 3 /path/to/your/submission
```

Revised graders cover Chapters 1–16, 18–23 and 26. Chapter 17 is explanatory,
Chapter 24 is optional and ungraded, and Chapter 25 is reserved. Checks run the
actual application against local fake services without paid inference. The
chapter states the public contract and any separate review or live-use work.

The October 14 chapter reorder is in the manuscript and grader commands.
Until reconstructed snapshots replace the old-order exports, grade an explicit
submission directory; the [chapter map](docs/edition-2-chapter-map.md) records
that transition.

Original grader failures remain failures in the historical record, including
independently accepted exceptions. Revised scores are separate results. Read
[the workflow](docs/agentic-codebook-workflow.md) and
[enhancement records](docs/edition-2-enhancement.md) for the distinctions.
To run an original grader, enter the archive first:

```sh
(cd edition-1 && go run ./cmd/grade -ch 1)
```

## Build the book

```sh
make book
make ebook
python3 book/build-print.py
```

The assembled Markdown is committed so it can be read directly in the repository.
The print PDFs use **10.5 pt default body text**, with 10 pt code and 11 pt
tables, on 7 × 10 inch pages.

The [publication instructions](book/README.md) describe EPUB navigation, print
page verification and the cover-prefixed PDF. The first-edition publication
files remain in `edition-1/book/`.

The book and code are [Apache-2.0 licensed](LICENSE). Fork them, change what your
agent needs, and preserve the reasons for the changes. A fresh implementation
is a useful test of how much the instructions actually teach.
