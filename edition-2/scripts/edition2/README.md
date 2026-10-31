# Edition 2 tooling

Current application source is `agent/`; frozen accepted solutions are
`solutions/chNN/`. Run the revised graders from the repository root:

```sh
go run ./cmd/grade-edition2 -ch 1
go run ./cmd/grade-edition2 -ch 26
make grade-dir CH=26 DIR=./agent
```

Chapter numbers follow the current book. Chapters 19 and 24 are ungraded;
Chapter 25 is the crossover placeholder. Original graders run from `edition-1/`
and retain first-edition chapter numbers and results.

## Exporting a newly accepted source

The exporter copies committed bytes and modes, never dirty working files. It
refuses an existing export or manifest and does not grant chapter acceptance:

```sh
python3 scripts/edition2/export_snapshot.py REV solutions/chNN --manifest /tmp/chNN-manifest.json
```

The default source prefix is `agent`. To reproduce a pre-migration source identity,
select its actual historical path explicitly:

```sh
python3 scripts/edition2/export_snapshot.py OLD_REV /tmp/old-source --manifest /tmp/old-manifest.json --source-prefix solutions/edition-2/main
python3 -m unittest discover -s scripts/edition2 -p test_export_snapshot.py
```

Other scripts in this directory are retained historical acceptance, mutation and
investigation tools, including the retired first attempt. Their source anchors,
module layouts and historical paths are not current setup instructions. Use their
original revision and recorded receipts when inspecting history; do not run a
historical mutation script against the active application merely to refresh a
result. The maintained current grader is `cmd/grade-edition2` with
`internal/gradeedition2`, not those older acceptance frameworks.
