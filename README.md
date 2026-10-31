# Ensemble

Ensemble is the coding agent built from *The Art of Building AI Coding Agents*,
Bill Cox's executable teaching book. **Edition 3 is the current edition:** the
application is C# 14 / .NET 10, developed inside Ensemble with Cadence.

## Start here

- [Read the book](book/README.md), [chapter guide](book/chapters.md), or
  [assembled manuscript](book/the-art-of-building-ai-coding-agents.md).
- [Current application source](agent/) and [human companion](book/companion/).
- [Preserved chapter solutions](solutions/README.md).
- [Promotion and verification](docs/edition-3-promotion.md),
  [implementation evidence](docs/edition-3-csharp-progress.md).
- Frozen history: [Edition 1](edition-1/ARCHIVE.md) and
  [Edition 2](edition-2/ARCHIVE.md), each with its book, source and graders.

## Build and run

Use an existing .NET 10 SDK. Nothing here installs an SDK, reads credentials for
a build, starts a server automatically, or replaces an installed application.

```sh
make                         # builds the C# GUI and writes ./ensemble launcher
make cli                     # builds the C# terminal client
# Override SDK location if needed: make DOTNET=/absolute/path/to/dotnet
./ensemble --help            # explicit launch; inspect options before use
# Terminal client after make cli:
dotnet agent/Cli/bin/Debug/net10.0/ch04.dll --help
```

Choose your own data directory, workspace, funding configuration and unused port
for a real launch. Do not replace a running host just to refresh source. The GUI
embeds its assets at build time. See [the crossover chapter](book/chapter-26.md)
and the source README for application behavior and qualified test evidence.

`make test` runs the C# regression executable and root Go tooling tests;
`make test-go` tests only the latter. Go remains at the root for reusable graders,
local fake providers and publication tooling—not as the active runtime.

## Grade an implementation

Use each chapter's documented executable command and scope. C# must not be sent
to a Go-source consumer grader. For example, Chapter 5 documents the explicit
`grade-edition3 -ch 5 -executable-only` route and its 65 scoped points.
`make grade CH=5 DIR=/path/to/CLI-wrapper GRADEFLAGS='-executable-only ...'`
requires an explicit target and forwards to that Edition 3 command.

Executable grader numbers and preserved `chNN` snapshots retain their original
meanings; [the chapter mapping](book/README.md#sequential-manuscript-numbers-and-historical-endpoints)
explains later manuscript renumbering. Historical Edition 2 grading can be run
from `edition-2/`; Edition 1 remains in `edition-1/`. Neither archive is rewritten
for the current layout.

## Build the book

`make book` assembles the current manuscript; `make ebook` builds its EPUB;
`make print` builds its print and reading PDFs. These use existing local publishing
tools, never install them. See [book build instructions](book/README.md#build-locally)
for prerequisites and isolated output. Generated publications are distinct from
source edits; a source promotion alone does not publish a new book or upload it.
